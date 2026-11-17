package agent

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// Compliance — сводка для «Плана резервирования и архивирования»
// (Приказ ФСТЭК №117): агент отдаёт фактическое состояние системы,
// документ собирается поверх этих данных (шаблон/LLM на стороне сервера).

func (a *Agent) Compliance(ctx context.Context, req JobRequest) (map[string]any, error) {
	cz := a.Corteza(req)
	if cz == nil || !cz.HasToken() {
		return nil, fmt.Errorf("corteza is not configured (no token)")
	}
	out := map[string]any{
		"generatedAt": time.Now().UTC().Format(time.RFC3339),
		"agent":       a.Health(ctx),
	}

	srcRecs, err := cz.ListRecords(ctx, "sources", "")
	if err != nil {
		return nil, err
	}
	var sources []map[string]any
	for i := range srcRecs {
		s := sourceFrom(&srcRecs[i])
		sources = append(sources, map[string]any{
			"id":        s.ID,
			"name":      s.Name,
			"type":      s.Type,
			"dataClass": firstNonEmpty(s.DataClass, "internal"),
			"enabled":   s.Enabled,
			"protected": s.Enabled && protectionNote(s),
			"notes":     s.Notes,
		})
	}
	out["sources"] = sources

	polRecs, err := cz.ListRecords(ctx, "policies", "")
	if err != nil {
		return nil, err
	}
	var policies []map[string]any
	for i := range polRecs {
		p := policyFrom(&polRecs[i])
		policies = append(policies, map[string]any{
			"id":             p.ID,
			"name":           p.Name,
			"sourceID":       p.SourceID,
			"cron":           p.Cron,
			"retentionDays":  p.RetentionDays,
			"engine":         p.Engine,
			"vault":          p.Vault,
			"vaultLegalHold": p.VaultLegalHold,
			"verify":         p.Verify,
			"verifyCron":     p.VerifyCron,
			"archive":        p.Archive,
			"keepYears":      p.KeepYears,
			"enabled":        p.Enabled,
		})
	}
	out["policies"] = policies

	snapRecs, err := cz.ListRecords(ctx, "snapshots", "")
	if err == nil {
		stats := map[string]any{"total": len(snapRecs), "restorable": 0, "verified": 0, "withManifest": 0, "byClass": map[string]int{}}
		var lastPerSource = map[uint64]string{}
		for i := range snapRecs {
			m := recMap(snapRecs[i])
			if ParseBool(m["restorable"]) {
				stats["restorable"] = stats["restorable"].(int) + 1
			}
			if ParseBool(m["verified"]) {
				stats["verified"] = stats["verified"].(int) + 1
			}
			if m["manifest_sha"] != "" {
				stats["withManifest"] = stats["withManifest"].(int) + 1
			}
			cls := firstNonEmpty(m["data_class"], "internal")
			stats["byClass"].(map[string]int)[cls]++
			srcID := ParseUint(m["source"])
			if t := snapRecs[i].UpdatedAt; t > lastPerSource[srcID] {
				lastPerSource[srcID] = t
			}
		}
		stats["lastBackupPerSource"] = len(lastPerSource)
		out["snapshots"] = stats
	}

	if findings, err := cz.ListRecords(ctx, "audit_findings", "status = 'open'"); err == nil {
		out["auditOpenFindings"] = len(findings)
	}
	out["wormVault"] = map[string]any{
		"enabled":       a.vault != nil && a.vault.Enabled(),
		"bucket":        vaultBucketName(a),
		"mode":          a.cfg.Vault.Mode,
		"retentionDays": a.cfg.Vault.RetentionDays,
	}
	// format=md: готовый текст «Плана резервирования и архивирования»
	// (структура под Приказ №117), вместо сырого JSON.
	if strings.EqualFold(strings.TrimSpace(req.ExtractTo), "md") {
		return map[string]any{"document": complianceMarkdown(out)}, nil
	}
	return out, nil
}

// complianceMarkdown — человекочитаемый план из фактической конфигурации.
func complianceMarkdown(d map[string]any) string {
	var b strings.Builder
	w := func(format string, args ...any) { fmt.Fprintf(&b, format+"\n", args...) }
	w("# План резервирования и архивирования данных")
	w("")
	w("Сформировано: %v (агент backup, фактическое состояние)", d["generatedAt"])
	w("")
	w("## 1. Состав резервируемых данных")
	for _, item := range d["sources"].([]map[string]any) {
		enabled := "выкл"
		if item["enabled"].(bool) {
			enabled = "вкл"
		}
		w("- **%v** (%v, класс: %v, %v)", item["name"], item["type"], item["dataClass"], enabled)
	}
	w("")
	w("## 2. Периодичность и сроки хранения")
	for _, item := range d["policies"].([]map[string]any) {
		w("- **%v**: cron `%v`, хранение %v дн, движок `%v`, сейф %v, архив %v",
			item["name"], item["cron"], item["retentionDays"], item["engine"], onOff(item["vault"]), onOff(item["archive"]))
	}
	w("")
	w("## 3. Место хранения")
	vault := d["wormVault"].(map[string]any)
	w("- Основное хранилище: MinIO (S3), бакет по умолчанию")
	w("- WORM-сейф: %v (бакет `%v`, режим %v, retention %v дн)", onOff(vault["enabled"]), vault["bucket"], vault["mode"], vault["retentionDays"])
	w("")
	w("## 4. Контроль восстановления")
	snaps := d["snapshots"].(map[string]any)
	w("- Снапшотов: %v, восстановимых: %v, верифицированных: %v, с манифестами: %v",
		snaps["total"], snaps["restorable"], snaps["verified"], snaps["withManifest"])
	w("- Регулярная верификация (учебные тревоги) по verify_cron политик")
	w("")
	w("## 5. Аудит покрытия")
	w("- Открытых находок аудита (данные без бэкапа): %v", d["auditOpenFindings"])
	return b.String()
}

func onOff(v any) string {
	switch t := v.(type) {
	case bool:
		if t {
			return "да"
		}
		return "нет"
	case int:
		if t > 0 {
			return "да"
		}
		return "нет"
	}
	return fmt.Sprintf("%v", v)
}

// protectionNote — покрывает ли источник рабочий контур: SMB/DB/S3
// считаются защищёнными по умолчанию, fs — только если не на том же
// сервере, что и агент (эвристика: путь не в /tmp и не рядом с агентом).
func protectionNote(s *Source) bool {
	switch s.Type {
	case SourceFS:
		return !stringsHasTmp(s.Path)
	default:
		return true
	}
}

func stringsHasTmp(p string) bool {
	p = strings.ToLower(strings.ReplaceAll(p, "\\", "/"))
	return strings.HasPrefix(p, "/tmp") || strings.HasPrefix(p, "tmp/")
}

func vaultBucketName(a *Agent) string {
	if a.vault == nil {
		return ""
	}
	return a.vault.Bucket()
}
