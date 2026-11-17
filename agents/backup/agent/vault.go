package agent

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
)

// StartVault копирует существующий снапшот в WORM-сейф: содержимое
// сканируется (clamd либо эвристики), при чистом/неизвестном вердикте —
// копия в vault-бакет с retention, при infected — копии нет вообще.
// StartScan делает только сканирование без копии (кнопка «перепроверить»).
// CleanPoint — sync-запрос последней чистой точки восстановления.

func (a *Agent) StartVault(ctx context.Context, req JobRequest) (*JobStatus, error) {
	return a.startVaultLike(ctx, req, KindVault, true)
}

func (a *Agent) StartScan(ctx context.Context, req JobRequest) (*JobStatus, error) {
	return a.startVaultLike(ctx, req, KindScan, false)
}

func (a *Agent) startVaultLike(ctx context.Context, req JobRequest, kind JobKind, copyToVault bool) (*JobStatus, error) {
	cz := a.Corteza(req)
	if cz == nil {
		return nil, fmt.Errorf("corteza is not configured")
	}
	snapID := ParseUint(req.SnapshotID)
	if snapID == 0 {
		return nil, fmt.Errorf("snapshotID is required")
	}
	snap, err := cz.LoadSnapshot(ctx, snapID)
	if err != nil {
		return nil, fmt.Errorf("snapshot %d: %w", snapID, err)
	}
	id := newJobID()
	st := &JobStatus{
		ID:          id,
		Kind:        string(kind),
		Status:      StatusRunning,
		StartedAt:   time.Now(),
		SnapshotID:  fmtUint(snap.ID),
		SourceID:    fmtUint(snap.SourceID),
		JobRecordID: req.JobRecordID,
		CallbackURL: req.CallbackURL,
		Token:       firstNonEmpty(req.Token, a.cfg.Token),
		NamespaceID: ParseUint(req.NamespaceID),
	}
	if st.NamespaceID == 0 {
		st.NamespaceID = a.cfg.NamespaceID
	}
	a.mu.Lock()
	a.jobs[id] = st
	a.mu.Unlock()
	go a.runVault(context.Background(), st, snap, copyToVault, cz)
	return st, nil
}

func (a *Agent) runVault(ctx context.Context, st *JobStatus, snap *Snapshot, copyToVault bool, cz *Corteza) {
	a.sem <- struct{}{}
	defer func() { <-a.sem }()
	handle := ""
	if src, err := cz.LoadSource(ctx, snap.SourceID); err == nil {
		handle = sanitizeKey(firstNonEmpty(src.Name, fmtUint(src.ID)))
	}
	if handle == "" {
		handle = sanitizeKey(fmtUint(snap.SourceID))
	}
	finish := func(status JobStatusName, errMsg string) {
		now := time.Now()
		st.Status = status
		st.Error = errMsg
		st.FinishedAt = &now
		st.Progress = 100
		a.notifyCallback(st, mapStatusEvent(status))
	}
	if a.store == nil {
		finish(StatusFailed, "minio is not configured")
		return
	}
	snapStore, serr := a.storeForSnapshot(ctx, cz, snap)
	if serr != nil {
		finish(StatusFailed, serr.Error())
		return
	}
	snapCfg := a.cfgForStore(snapStore)
	if snap.Engine == "kopia" {
		// Блобы kopia зашифрованы и порезаны на чанки: ClamAV нечего сканировать,
		// а per-snapshot копия в сейф для content-addressed репозитория
		// бессмысленна (данные общие между снапшотами). Честно помечаем.
		now := time.Now().UTC().Format(time.RFC3339)
		if cz.HasToken() {
			_ = cz.UpdateValues(ctx, "snapshots", snap.ID, map[string]string{
				"scan_status": string(ScanUnknown),
				"scanned_at":  now,
				"scan_log":    "kopia engine: encrypted content-defined chunks are not scannable; vault copy not applicable",
			})
		}
		st.Message = "kopia: scan/vault not applicable"
		finish(StatusCompleted, "")
		return
	}

	// --- Сканирование ---
	res, scanLog, err := a.scanSnapshot(ctx, st, snap, snapStore, handle)
	if err != nil {
		finish(StatusFailed, err.Error())
		return
	}
	metricScan.WithLabelValues(string(res.Status)).Inc()
	st.Message = "scan: " + string(res.Status)

	updates := map[string]string{
		"scan_status": string(res.Status),
		"scanned_at":  time.Now().UTC().Format(time.RFC3339),
		"scan_log":    scanLog,
	}
	if cz.HasToken() {
		if err := cz.UpdateValues(ctx, "snapshots", snap.ID, updates); err != nil {
			log.Printf("vault: update snapshot %d: %v", snap.ID, err)
		}
	}

	if res.Status == ScanInfected {
		finish(StatusFailed, "infected: "+res.Signature)
		return
	}
	if !copyToVault {
		finish(StatusCompleted, "")
		return
	}
	if a.vault == nil || !a.vault.Enabled() {
		finish(StatusFailed, "vault is not enabled (bucket missing or no object lock)")
		return
	}

	// --- Копия в сейф ---
	pol, _ := cz.LoadPolicy(ctx, snap.PolicyID)
	retDays := a.cfg.Vault.RetentionDays
	legalHold := false
	if pol != nil {
		if pol.VaultRetentionDays > 0 {
			retDays = pol.VaultRetentionDays
		}
		legalHold = pol.VaultLegalHold && res.Status == ScanClean
	}
	if retDays <= 0 {
		retDays = 30
	}
	lockedUntil := time.Now().Add(time.Duration(retDays) * 24 * time.Hour)

	vaultKey := ""
	switch {
	case strings.HasPrefix(snap.Engine, "restic"):
		if err := resticCopyVault(ctx, snapCfg, handle, snap.ResticID, a.cfg.ResticPassword); err != nil {
			finish(StatusFailed, err.Error())
			return
		}
		vaultKey = resticKeyHint(handle, snap.ResticID) + " (restic repo)"
	default:
		vaultKey = "vault/" + snap.S3Key
		var err error
		if a.vault.SameEndpoint(snapStore.Config()) {
			// Один MinIO — серверная копия без трафика через агента.
			err = a.vault.CopyRetained(ctx, snap.S3Bucket, snap.S3Key, vaultKey, lockedUntil)
		} else {
			// Снапшот на другом endpoint — качаем и кладём в сейф потоком.
			err = a.vault.CopyRetainedFrom(ctx, snapStore, snap.S3Bucket, snap.S3Key, vaultKey, lockedUntil)
		}
		if err != nil {
			finish(StatusFailed, err.Error())
			return
		}
	}
	if legalHold && !strings.HasPrefix(snap.Engine, "restic") {
		_ = a.vault.SetLegalHold(ctx, vaultKey, true) // best-effort, ругаемся в лог
	}
	if cz.HasToken() {
		if err := cz.UpdateValues(ctx, "snapshots", snap.ID, map[string]string{
			"vault_key":          vaultKey,
			"vault_locked_until": lockedUntil.UTC().Format(time.RFC3339),
		}); err != nil {
			log.Printf("vault: update snapshot %d: %v", snap.ID, err)
		}
	}
	metricVaultObjects.Inc()
	metricCleanPoint.WithLabelValues(handle).Set(float64(snap.ID))
	finish(StatusCompleted, "")
}

// scanSnapshot выбирает способ чтения по движку снапшота и возвращает
// суммарный вердикт: худший из вердиктов по всем проверенным частям.
// store — стор снапшота (его endpoint, а не обязательно дефолт агента).
func (a *Agent) scanSnapshot(ctx context.Context, st *JobStatus, snap *Snapshot, store Storage, handle string) (*ScanResult, string, error) {
	sc := resolveScanner(a.cfg.ClamAVAddr, a.cfg.ScanMaxMB)
	rc, err := store.GetBucket(ctx, snap.S3Bucket, snap.S3Key)
	if err != nil {
		return nil, "", fmt.Errorf("get %s: %w", snap.S3Key, err)
	}
	defer rc.Close()

	var parts []func() (*ScanResult, error)
	if snap.Engine == "archive" || snap.Engine == "s3copy" {
		parts = append(parts, func() (*ScanResult, error) { return scanTarGz(ctx, sc, rc, st) })
	} else {
		name := snap.S3Key
		parts = append(parts, func() (*ScanResult, error) {
			st.Message = "scanning " + name
			return sc.Scan(ctx, rc, name)
		})
	}

	worst := &ScanResult{Status: ScanClean}
	var logLines []string
	for _, part := range parts {
		res, err := part()
		if err != nil {
			// Сканер упал — unknown, но копия в сейф всё равно имеет смысл.
			res = &ScanResult{Status: ScanUnknown, Detail: err.Error()}
			err = nil
		}
		logLines = append(logLines, res.LogLine(handle))
		if rankScan(res.Status) > rankScan(worst.Status) {
			worst = res
		}
	}
	return worst, strings.Join(logLines, "\n"), nil
}

func rankScan(s ScanStatus) int {
	switch s {
	case ScanInfected:
		return 3
	case ScanUnknown:
		return 2
	case ScanQueued:
		return 1
	default:
		return 0
	}
}

// scanTarGz разворачивает архив по записям и шлёт каждый файл в сканер.
// Останавливается на первом infected; записи больше лимита сканера пропускает.
func scanTarGz(ctx context.Context, sc Scanner, r io.Reader, st *JobStatus) (*ScanResult, error) {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return &ScanResult{Status: ScanUnknown, Detail: "gunzip: " + err.Error()}, nil
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	worst := &ScanResult{Status: ScanClean, Detail: "empty archive"}
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return &ScanResult{Status: ScanUnknown, Detail: "tar: " + err.Error()}, nil
		}
		if hdr.FileInfo().IsDir() || hdr.Typeflag != tar.TypeReg {
			continue
		}
		st.Message = "scanning " + hdr.Name
		res, err := sc.Scan(ctx, tr, hdr.Name)
		if err != nil {
			return &ScanResult{Status: ScanUnknown, Detail: err.Error()}, nil
		}
		worst.Bytes += res.Bytes
		if rankScan(res.Status) > rankScan(worst.Status) {
			worst.Status = res.Status
			worst.Signature = res.Signature
			worst.Detail = res.Detail
		}
		if worst.Status == ScanInfected {
			break
		}
	}
	return worst, nil
}

// CleanPoint — последняя чистая точка восстановления (sync-инструмент LLM):
// новейший снапшот, который можно восстановить и который не заражён.
func (a *Agent) CleanPoint(ctx context.Context, req JobRequest) (map[string]any, error) {
	cz := a.Corteza(req)
	if cz == nil || !cz.HasToken() {
		return nil, fmt.Errorf("corteza is not configured (no token)")
	}
	snaps, err := cz.ListRecords(ctx, "snapshots", "restorable = '1'")
	if err != nil {
		return nil, err
	}
	srcFilter := ParseUint(req.SourceID)
	for i := len(snaps) - 1; i >= 0; i-- {
		m := recMap(snaps[i])
		if srcFilter > 0 && ParseUint(m["source"]) != srcFilter {
			continue
		}
		if ScanStatus(m["scan_status"]) == ScanInfected {
			continue
		}
		return map[string]any{
			"snapshotID":       fmtUint(snaps[i].ID),
			"createdAt":        snaps[i].UpdatedAt,
			"sourceID":         m["source"],
			"s3Key":            m["s3_key"],
			"sizeBytes":        ParseUint(m["size_bytes"]),
			"engine":           m["engine"],
			"resticID":         m["restic_id"],
			"scanStatus":       m["scan_status"],
			"vaultKey":         m["vault_key"],
			"vaultLockedUntil": m["vault_locked_until"],
		}, nil
	}
	return map[string]any{"message": "no clean restore point found"}, nil
}

func mapStatusEvent(s JobStatusName) string {
	if s == StatusFailed {
		return "failed"
	}
	return "complete"
}

func newJobID() string { return uuid.New().String() }
