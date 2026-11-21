package agent

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Аудит незабэкапленных данных (интервью, п.3 «автоматизированный аудит»):
// агент обходит области (audit_scopes), находит подходящие под маски файлы и
// проверяет, покрывает ли их какой-нибудь включённый fs/smb-источник.
// Не покрытое — находки в audit_findings.
//
// Аудит — фоновая операция (descriptor Async): обход больших деревьев не
// укладывается в синхронный HTTP-вызов из цепочки (у component-ноды жёсткий
// таймаут 30 с). Результат появляется сразу в audit_findings, сводка —
// в статусе джоба (GET /api/jobs/{id}) и в его финальном callback.

// AuditScope — область аудита из модуля audit_scopes.
type AuditScope struct {
	ID          uint64
	Name        string
	Type        SourceType
	Path        string
	Host        string
	Share       string
	SMBPath     string
	CredID      uint64
	Masks       []string
	MinSizeMB   int64
	ContentScan bool // читать содержимое непокрытых файлов и искать ПДн (fs)
	Enabled     bool
}

func auditScopeFrom(rec composeRecord) AuditScope {
	m := recMap(rec)
	var masks []string
	for _, line := range strings.Split(m["masks"], "\n") {
		for _, pat := range strings.Split(line, ",") {
			if pat = strings.TrimSpace(pat); pat != "" {
				masks = append(masks, pat)
			}
		}
	}
	return AuditScope{
		ID:          rec.ID,
		Name:        m["name"],
		Type:        SourceType(firstNonEmpty(m["type"], "fs")),
		Path:        m["path"],
		Host:        m["host"],
		Share:       m["share"],
		SMBPath:     m["smb_path"],
		CredID:      ParseUint(m["credential"]),
		Masks:       masks,
		MinSizeMB:   int64(ParseUint(m["min_size_mb"])),
		ContentScan: ParseBool(m["content_scan"]),
		Enabled:     ParseBool(m["enabled"]),
	}
}

// guessClass — предполагаемый класс данных по имени файла.
func guessClass(name string) string {
	n := strings.ToLower(name)
	personal := []string{"персональн", "медицин", "снилс", "паспорт", "инн", "здоровь", "диагноз", "карта пациент", "personal", "medical", "patient"}
	for _, marker := range personal {
		if strings.Contains(n, marker) {
			return "personal_special"
		}
	}
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(n)), ".")
	switch ext {
	case "doc", "docx", "xls", "xlsx", "pdf", "odt", "ods":
		return "confidential"
	case "1cd", "mdp", "bak", "sql", "dump", "pgdump", "mdb", "dbf":
		return "internal"
	}
	return "internal"
}

func matchMasks(name string, masks []string) bool {
	if len(masks) == 0 {
		return true
	}
	base := filepath.Base(name)
	for _, pat := range masks {
		if ok, _ := filepath.Match(pat, name); ok {
			return true
		}
		if !strings.Contains(pat, string(filepath.Separator)) && !strings.Contains(pat, "/") {
			if ok, _ := filepath.Match(pat, base); ok {
				return true
			}
		}
	}
	return false
}

// coversReport проверяет, входит ли путь в зону ответственности хотя бы
// одного включённого fs/smb-источника. UNC- и локальные пути нормализуем
// к «чистым» относительным сравнениям; строгость намеренно консервативная:
// лучше лишняя находка, чем пропущенные данные.
func coversPath(filePath string, sources []Source) bool {
	norm := normalizeAuditPath(filePath)
	for i := range sources {
		s := &sources[i]
		if !s.Enabled || (s.Type != SourceFS && s.Type != SourceSMB) {
			continue
		}
		root := normalizeAuditPath(firstNonEmpty(s.Path, pathJoinUNC(s.Host, s.Share, s.SMBPath)))
		if root == "" {
			continue
		}
		if norm == root || strings.HasPrefix(norm, root+"/") {
			return true
		}
	}
	return false
}

func normalizeAuditPath(p string) string {
	p = strings.ReplaceAll(strings.TrimSpace(p), "\\", "/")
	p = strings.TrimSuffix(p, "/")
	p = strings.TrimPrefix(p, "//")
	p = strings.TrimPrefix(p, "./")
	return strings.ToLower(p)
}

func pathJoinUNC(host, share, sub string) string {
	p := strings.Trim(host, "/") + "/" + strings.Trim(share, "/")
	if sub = strings.Trim(sub, "/"); sub != "" {
		p += "/" + sub
	}
	return p
}

// maxAuditFindings — страховка от раздутия audit_findings на широких
// областях: за один прогон пишем не больше находок, остальное считаем
// усечённым и отмечаем в сводке. Env BACKUP_AUDIT_MAX_FINDINGS перекрывает.
func maxAuditFindings() int64 {
	if n := ParseInt(os.Getenv("BACKUP_AUDIT_MAX_FINDINGS")); n > 0 {
		return int64(n)
	}
	return 10000
}

const auditFindingsWorkers = 8

// findingSink — параллельная запись находок в audit_findings. Прямая
// серийная запись — узкое место аудита: на каждый файл отдельный HTTP-запрос.
type findingSink struct {
	cz      *Corteza
	ch      chan map[string]string
	wg      sync.WaitGroup
	capN    int64
	created int64
	trunc   atomic.Bool
}

func newFindingSink(cz *Corteza) *findingSink {
	s := &findingSink{cz: cz, ch: make(chan map[string]string, 256), capN: maxAuditFindings()}
	s.wg.Add(auditFindingsWorkers)
	for i := 0; i < auditFindingsWorkers; i++ {
		go s.worker()
	}
	return s
}

func (s *findingSink) worker() {
	defer s.wg.Done()
	for vals := range s.ch {
		if _, err := s.cz.CreateValues(context.Background(), "audit_findings", vals); err != nil {
			log.Printf("audit: create finding %s: %v", vals["path"], err)
		}
	}
}

// push после исчерпания cap молча отбрасывает находки и помечает сводку
// truncated — прогон продолжает считать статистику, но не пишет. Возвращает
// false, если находка отброшена: счётчики сводки тогда её не учитывают.
func (s *findingSink) push(vals map[string]string) bool {
	if s.trunc.Load() {
		return false
	}
	if atomic.AddInt64(&s.created, 1) > s.capN {
		s.trunc.Store(true)
		return false
	}
	s.ch <- vals
	return true
}

func (s *findingSink) close() {
	close(s.ch)
	s.wg.Wait()
}

// StartAudit регистрирует фоновый прогон по всем включённым областям.
// CallbackURL намеренно не сохраняем: инжеста джобов у аудита нет — результат
// сразу пишется в audit_findings, статус смотрят по GET /api/jobs/{id}.
func (a *Agent) StartAudit(ctx context.Context, req JobRequest) (*JobStatus, error) {
	cz := a.Corteza(req)
	if cz == nil || !cz.HasToken() {
		return nil, fmt.Errorf("corteza is not configured (no token)")
	}
	id := newJobID()
	st := &JobStatus{
		ID:          id,
		Kind:        "audit",
		Status:      StatusRunning,
		StartedAt:   time.Now(),
		Token:       firstNonEmpty(req.Token, a.cfg.Token),
		NamespaceID: ParseUint(req.NamespaceID),
	}
	if st.NamespaceID == 0 {
		st.NamespaceID = a.cfg.NamespaceID
	}
	a.mu.Lock()
	a.jobs[id] = st
	a.mu.Unlock()
	go a.runAudit(context.Background(), st, req, cz)
	return st, nil
}

func (a *Agent) runAudit(ctx context.Context, st *JobStatus, req JobRequest, cz *Corteza) {
	a.sem <- struct{}{}
	defer func() { <-a.sem }()
	finish := func(status JobStatusName, errMsg string) {
		now := time.Now()
		st.Status = status
		st.Error = errMsg
		st.FinishedAt = &now
		st.Progress = 100
		a.notifyCallback(st, mapStatusEvent(status))
	}
	summary, err := a.auditAll(ctx, cz, req.ScopeID, func(scope AuditScope, detail map[string]any, progress float64) {
		st.Files += detail["scanned"].(int)
		st.Progress = progress
		st.Message = fmt.Sprintf("область %s: сканировано %d, покрыто %d, находок %d",
			scope.Name, detail["scanned"], detail["covered"], detail["findings"])
	})
	st.Summary = summary
	if err != nil {
		finish(StatusFailed, err.Error())
		return
	}
	byCrit := summary["byCriticality"].(map[string]int)
	st.Message = fmt.Sprintf("аудит: областей %d, файлов %d, покрыто %d, находок %d (критичных %d, высоких %d)",
		summary["scopes"], summary["filesScanned"], summary["covered"], summary["findings"],
		byCrit[CritCritical], byCrit[CritHigh])
	finish(StatusCompleted, "")
}

// Audit — синхронный прогон (ручной вызов, тесты). Через цепочки ходят
// через StartAudit — фоновый джоб без HTTP-таймаута.
func (a *Agent) Audit(ctx context.Context, req JobRequest) (map[string]any, error) {
	cz := a.Corteza(req)
	if cz == nil || !cz.HasToken() {
		return nil, fmt.Errorf("corteza is not configured (no token)")
	}
	return a.auditAll(ctx, cz, req.ScopeID, nil)
}

// auditAll — обход областей; onScope зовётся после каждой области с её
// детализацией и общим прогрессом (0–100). scopeID непустой — только эта
// область, причём независимо от флага enabled (явный запуск с карточки).
func (a *Agent) auditAll(ctx context.Context, cz *Corteza, scopeID string, onScope func(AuditScope, map[string]any, float64)) (map[string]any, error) {
	scopeRecs, err := cz.ListRecords(ctx, "audit_scopes", "")
	if err != nil {
		return nil, err
	}
	srcRecs, err := cz.ListRecords(ctx, "sources", "")
	if err != nil {
		return nil, err
	}
	var sources []Source
	for i := range srcRecs {
		s := sourceFrom(&srcRecs[i])
		if s.Enabled {
			sources = append(sources, *s)
		}
	}

	onlyID := ParseUint(scopeID)
	total := 0
	for i := range scopeRecs {
		scope := auditScopeFrom(scopeRecs[i])
		if onlyID > 0 {
			if scope.ID == onlyID {
				total++
			}
			continue
		}
		if scope.Enabled {
			total++
		}
	}

	sink := newFindingSink(cz)
	now := time.Now()
	summary := map[string]any{
		"scopes": 0, "filesScanned": 0, "covered": 0, "findings": 0,
		"byCriticality":        map[string]int{CritLow: 0, CritMedium: 0, CritHigh: 0, CritCritical: 0},
		"coveredByCriticality": map[string]int{CritLow: 0, CritMedium: 0, CritHigh: 0, CritCritical: 0},
		"piiScanned":           0,
		"scopesDetail":         []map[string]any{},
	}
	done := 0
	for i := range scopeRecs {
		scope := auditScopeFrom(scopeRecs[i])
		if onlyID > 0 {
			if scope.ID != onlyID {
				continue
			}
		} else if !scope.Enabled {
			continue
		}
		detail, err := a.auditScope(ctx, cz, scope, sources, sink, now)
		if err != nil {
			detail = map[string]any{"scopeID": fmtUint(scope.ID), "name": scope.Name, "error": err.Error()}
		}
		summary["scopes"] = summary["scopes"].(int) + 1
		summary["filesScanned"] = summary["filesScanned"].(int) + detail["scanned"].(int)
		summary["covered"] = summary["covered"].(int) + detail["covered"].(int)
		summary["findings"] = summary["findings"].(int) + detail["findings"].(int)
		summary["piiScanned"] = summary["piiScanned"].(int) + detail["piiScanned"].(int)
		mergeCrit := func(key string) {
			dst := summary[key].(map[string]int)
			for k, v := range detail[key].(map[string]int) {
				dst[k] += v
			}
		}
		mergeCrit("byCriticality")
		mergeCrit("coveredByCriticality")
		summary["scopesDetail"] = append(summary["scopesDetail"].([]map[string]any), detail)
		if onScope != nil {
			done++
			progress := 100 * float64(done) / float64(max(1, total))
			onScope(scope, detail, progress)
		}
	}
	sink.close()
	if sink.trunc.Load() {
		summary["truncated"] = true
		summary["findingsCap"] = sink.capN
	}
	metricAuditFindings.Set(float64(summary["findings"].(int)))
	return summary, nil
}

func (a *Agent) auditScope(ctx context.Context, cz *Corteza, scope AuditScope, sources []Source, sink *findingSink, now time.Time) (map[string]any, error) {
	detail := map[string]any{
		"scopeID": fmtUint(scope.ID), "name": scope.Name,
		"scanned": 0, "covered": 0, "findings": 0, "piiScanned": 0,
		"byCriticality":        map[string]int{CritLow: 0, CritMedium: 0, CritHigh: 0, CritCritical: 0},
		"coveredByCriticality": map[string]int{CritLow: 0, CritMedium: 0, CritHigh: 0, CritCritical: 0},
	}

	// Закрываем прошлые открытые находки области.
	old, err := cz.ListRecords(ctx, "audit_findings", fmt.Sprintf("scope = %d AND status = 'open'", scope.ID))
	if err == nil {
		for _, rec := range old {
			_ = cz.UpdateValues(ctx, "audit_findings", rec.ID, map[string]string{"status": "resolved"})
		}
	}

	collect := func(rel string, size int64, mod time.Time, diskPath string) error {
		detail["scanned"] = detail["scanned"].(int) + 1
		if scope.MinSizeMB > 0 && size < scope.MinSizeMB*(1<<20) {
			return nil
		}
		if !matchMasks(rel, scope.Masks) {
			return nil
		}
		a := assessFile(rel, size, mod, now)
		if coversPath(rel, sources) || coversPath(pathJoinUNC(scope.Host, scope.Share, pathJoin2(scope.SMBPath, rel)), sources) {
			detail["covered"] = detail["covered"].(int) + 1
			detail["coveredByCriticality"].(map[string]int)[a.Criticality]++
			return nil
		}
		full := rel
		if scope.Type == SourceSMB {
			full = pathJoinUNC(scope.Host, scope.Share, pathJoin2(scope.SMBPath, rel))
		} else if scope.ContentScan && diskPath != "" {
			// Каталог 1С/кадры и прочие имена могут не раскрыть ПДн —
			// читаем непокрытый файл и ищем структурные паттерны.
			if hits, ok := scanFilePIITimed(diskPath); ok {
				detail["piiScanned"] = detail["piiScanned"].(int) + 1
				a.escalatePII(hits)
			}
		}
		if sink.push(map[string]string{
			"scope":         fmtUint(scope.ID),
			"path":          full,
			"size_bytes":    fmtInt64(size),
			"mod_time":      mod.UTC().Format(time.RFC3339),
			"guessed_class": a.Class,
			"criticality":   a.Criticality,
			"reasons":       strings.Join(a.Reasons, "; "),
			"reason":        "no enabled fs/smb source covers this path",
			"status":        "open",
			"created_at":    time.Now().UTC().Format(time.RFC3339),
		}) {
			detail["findings"] = detail["findings"].(int) + 1
			detail["byCriticality"].(map[string]int)[a.Criticality]++
		}
		return nil
	}

	var walkErr error
	if scope.Type == SourceSMB {
		host, share, root, err := ParseUNC(scope.Path, scope.Host, scope.Share, scope.SMBPath)
		if err != nil {
			return detail, err
		}
		cred, _ := cz.LoadCredential(ctx, scope.CredID)
		smb, err := DialSMB(host, cred.Username, ResolveSecret(cred.Handle), cred.Extra, share, root)
		if err != nil {
			return detail, err
		}
		defer smb.Close()
		walkErr = smb.Walk(ctx, func(e FileEntry, _ openFn) error {
			return collect(e.RelPath, e.Size, e.ModTime, "")
		})
	} else {
		root := scope.Path
		walkErr = filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			if err := ctx.Err(); err != nil {
				return err
			}
			info, err := d.Info()
			if err != nil {
				return nil
			}
			rel, err := filepath.Rel(root, p)
			if err != nil {
				return nil
			}
			return collect(filepath.ToSlash(rel), info.Size(), info.ModTime(), p)
		})
	}
	if walkErr != nil {
		return detail, walkErr
	}
	return detail, nil
}

func pathJoin2(base, rel string) string {
	base = strings.Trim(base, "/")
	rel = strings.Trim(rel, "/")
	if base == "" {
		return rel
	}
	if rel == "" {
		return base
	}
	return base + "/" + rel
}
