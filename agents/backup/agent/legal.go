package agent

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

// Юридически значимые электронные архивы (интервью, п.4 и 6):
//   - backup/archive-manifest: для снапшота строит манифест {файл, sha256,
//     размер} с HMAC-подписью агента — доказательство «документ не менялся
//     с момента архивации». Манифест кладётся в MinIO, хэш — в снапшот.
//   - backup/destroy: уничтожение по истечении срока хранения. Объекты под
//     legal hold не удаляются (бакет их не отдаст) — статус blocked.

// ArchiveManifest — подписанный перечень файлов снапшота.
type ArchiveManifest struct {
	SnapshotID uint64          `json:"snapshotID"`
	S3Key      string          `json:"s3Key"`
	CreatedAt  string          `json:"createdAt"`
	Files      []ManifestEntry `json:"files"`
	Signature  string          `json:"signature"` // HMAC-SHA256(manifest json, ключ агента)
}

type ManifestEntry struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size"`
}

// archiveKey — ключ для подписи манифестов. Переопределяется env
// BACKUP_ARCHIVE_KEY; иначе — от restic-пароля (не идеально, но лучше хардкода).
func (a *Agent) archiveKey() []byte {
	if k := strings.TrimSpace(firstNonEmpty(os.Getenv("BACKUP_ARCHIVE_KEY"), a.cfg.ResticPassword)); k != "" {
		return []byte(k)
	}
	return []byte("backup-agent-archive-key")
}

// StartArchiveManifest — async op: построить и сохранить манифест.
func (a *Agent) StartArchiveManifest(ctx context.Context, req JobRequest) (*JobStatus, error) {
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
	st := &JobStatus{
		ID:          newJobID(),
		Kind:        string(KindManifest),
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
	a.jobs[st.ID] = st
	a.mu.Unlock()
	go a.runManifest(context.Background(), st, snap, cz)
	return st, nil
}

func (a *Agent) runManifest(ctx context.Context, st *JobStatus, snap *Snapshot, cz *Corteza) {
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
	var entries []ManifestEntry
	var err error
	snapStore, serr := a.storeForSnapshot(ctx, cz, snap)
	if serr != nil {
		finish(StatusFailed, serr.Error())
		return
	}
	destRec, derr := a.destinationForSnapshot(ctx, cz, snap)
	if derr != nil {
		finish(StatusFailed, derr.Error())
		return
	}
	switch {
	case snap.Engine == "archive" || snap.Engine == "k8s" || strings.HasPrefix(snap.Engine, "saas-"):
		rc, gerr := snapStore.GetBucket(ctx, snap.S3Bucket, snap.S3Key)
		if gerr != nil {
			finish(StatusFailed, gerr.Error())
			return
		}
		defer rc.Close()
		entries, err = hashTarGz(ctx, rc)
	case snap.Engine == "kopia":
		src, _ := cz.LoadSource(ctx, snap.SourceID)
		handle := sanitizeKey(fmtUint(snap.SourceID))
		if src != nil {
			handle = sanitizeKey(firstNonEmpty(src.Name, fmtUint(src.ID)))
		}
		if kerr := kopiaConnect(ctx, a.cfg, destRec, handle, a.cfg.ResticPassword); kerr != nil {
			finish(StatusFailed, kerr.Error())
			return
		}
		entries, err = kopiaManifestEntries(ctx, a.cfg, handle, snap.ResticID)
	case strings.HasPrefix(snap.Engine, "restic"):
		src, _ := cz.LoadSource(ctx, snap.SourceID)
		handle := sanitizeKey(fmtUint(snap.SourceID))
		if src != nil {
			handle = sanitizeKey(firstNonEmpty(src.Name, fmtUint(src.ID)))
		}
		spec, rerr := a.engineRepoFor(ctx, destRec, handle)
		if rerr != nil {
			finish(StatusFailed, rerr.Error())
			return
		}
		defer spec.close()
		entries, err = resticManifestEntries(ctx, spec, snap.ResticID)
	default:
		finish(StatusFailed, "manifest supported for archive engines only")
		return
	}
	if err != nil {
		finish(StatusFailed, err.Error())
		return
	}
	manifest := ArchiveManifest{
		SnapshotID: snap.ID,
		S3Key:      snap.S3Key,
		CreatedAt:  time.Now().UTC().Format(time.RFC3339),
		Files:      entries,
	}
	unsigned, _ := json.Marshal(struct {
		SnapshotID uint64          `json:"snapshotID"`
		S3Key      string          `json:"s3Key"`
		CreatedAt  string          `json:"createdAt"`
		Files      []ManifestEntry `json:"files"`
	}{manifest.SnapshotID, manifest.S3Key, manifest.CreatedAt, manifest.Files})
	mac := hmac.New(sha256.New, a.archiveKey())
	mac.Write(unsigned)
	manifest.Signature = hex.EncodeToString(mac.Sum(nil))
	manifestJSON, _ := json.MarshalIndent(manifest, "", "  ")

	key := "manifests/" + strings.TrimSuffix(snap.S3Key, "/") + ".manifest.json"
	// Манифест хранится рядом с данными — в том же назначении, что и снапшот.
	if _, err := snapStore.Put(ctx, key, strings.NewReader(string(manifestJSON)), int64(len(manifestJSON)), "application/json"); err != nil {
		finish(StatusFailed, err.Error())
		return
	}
	sum := sha256.Sum256(manifestJSON)
	if cz.HasToken() {
		_ = cz.UpdateValues(ctx, "snapshots", snap.ID, map[string]string{
			"manifest_key": key,
			"manifest_sha": hex.EncodeToString(sum[:]),
		})
	}
	st.Files = len(entries)
	st.Message = fmt.Sprintf("manifest: %d files, key=%s", len(entries), key)
	finish(StatusCompleted, "")
}

// hashTarGz считает sha256 каждой записи архива.
func hashTarGz(ctx context.Context, r io.Reader) ([]ManifestEntry, error) {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("gunzip: %w", err)
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	var out []ManifestEntry
	for {
		if err := ctx.Err(); err != nil {
			return out, err
		}
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return out, fmt.Errorf("tar: %w", err)
		}
		if hdr.FileInfo().IsDir() || hdr.Typeflag != tar.TypeReg {
			continue
		}
		h := sha256.New()
		if _, err := io.Copy(h, tr); err != nil {
			return out, err
		}
		out = append(out, ManifestEntry{Path: hdr.Name, SHA256: hex.EncodeToString(h.Sum(nil)), Size: hdr.Size})
	}
	return out, nil
}

// StartDestroy — async op: уничтожение снапшота по истечении срока хранения.
// Логирует основание; объекты под legal hold бакет не отдаст — job failed,
// статус blocked проставляет вызывающая цепочка в destruction_acts.
func (a *Agent) StartDestroy(ctx context.Context, req JobRequest) (*JobStatus, error) {
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
	st := &JobStatus{
		ID:          newJobID(),
		Kind:        string(KindDestroy),
		Status:      StatusRunning,
		StartedAt:   time.Now(),
		SnapshotID:  fmtUint(snap.ID),
		SourceID:    fmtUint(snap.SourceID),
		S3Bucket:    snap.S3Bucket,
		S3Key:       snap.S3Key,
		Engine:      snap.Engine,
		JobRecordID: req.JobRecordID,
		CallbackURL: req.CallbackURL,
		Token:       firstNonEmpty(req.Token, a.cfg.Token),
		NamespaceID: ParseUint(req.NamespaceID),
	}
	st.Message = "destroy: " + firstNonEmpty(req.Basis, "retention expired")
	a.mu.Lock()
	a.jobs[st.ID] = st
	a.mu.Unlock()
	go a.runDestroy(context.Background(), st, snap, cz)
	return st, nil
}

func (a *Agent) runDestroy(ctx context.Context, st *JobStatus, snap *Snapshot, cz *Corteza) {
	a.sem <- struct{}{}
	defer func() { <-a.sem }()
	finish := func(status JobStatusName, errMsg string) {
		now := time.Now()
		st.Status = status
		st.Error = errMsg
		st.FinishedAt = &now
		st.Progress = 100
		if cz.HasToken() {
			vals := map[string]string{"restorable": "0"}
			if status == StatusCompleted {
				vals["expires_at"] = time.Now().UTC().Format(time.RFC3339)
			}
			_ = cz.UpdateValues(ctx, "snapshots", snap.ID, vals)
			// Акт уничтожения (строка из цепочки) — фиксируем итог.
			if st.JobRecordID != "" {
				if rid := ParseUint(st.JobRecordID); rid > 0 {
					actStatus := "executed"
					if status != StatusCompleted {
						actStatus = "blocked"
					}
					_ = cz.UpdateValues(ctx, "destruction_acts", rid, map[string]string{
						"status":       actStatus,
						"destroyed_by": hostname(),
						"files_count":  strconv.Itoa(st.Files),
					})
				}
			}
		}
		a.notifyCallback(st, mapStatusEvent(status))
	}
	if strings.HasPrefix(snap.Engine, "restic") || snap.Engine == "kopia" {
		finish(StatusFailed, "destroy for restic/kopia engines is not supported; expire via engine retention")
		return
	}
	store, err := a.storeForSnapshot(ctx, cz, snap)
	if err != nil {
		finish(StatusFailed, err.Error())
		return
	}
	// Если копия в сейфе под legal hold — бакет сам не отдаст объект,
	// удаление упадёт со статусом blocked в destruction_acts.
	key := snap.S3Key
	if strings.HasSuffix(key, "/") {
		keys, lerr := store.ListBucket(ctx, snap.S3Bucket, key)
		if lerr != nil {
			finish(StatusFailed, lerr.Error())
			return
		}
		deleted := 0
		for _, k := range keys {
			if err := store.RemoveBucket(ctx, snap.S3Bucket, k); err != nil {
				finish(StatusFailed, fmt.Sprintf("remove %s: %v (legal hold?)", k, err))
				return
			}
			deleted++
		}
		st.Files = deleted
	} else if err := store.RemoveBucket(ctx, snap.S3Bucket, key); err != nil {
		finish(StatusFailed, fmt.Sprintf("remove %s: %v (legal hold?)", key, err))
		return
	} else {
		st.Files = 1
	}
	finish(StatusCompleted, "")
}

// kopiaManifestEntries — `kopia ls -l <id>`: "perm size date path" в строке.
// SHA для чанковых движков не считаем: манифест фиксирует состав и размер.
func kopiaManifestEntries(ctx context.Context, cfg Config, handle, snapshotID string) ([]ManifestEntry, error) {
	cfgPath := kopiaConfigPath(handle)
	cmd := kopiaCLI(ctx, cfg, cfgPath, "ls", "-l", snapshotID)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("kopia ls -l: %w", err)
	}
	return parseKopiaLsLong(string(out)), nil
}

// parseKopiaLsLong парсит `kopia ls -l`: "perm size date time path".
func parseKopiaLsLong(out string) []ManifestEntry {
	var entries []ManifestEntry
	for _, line := range strings.Split(out, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		size, _ := strconv.ParseInt(fields[1], 10, 64)
		entries = append(entries, ManifestEntry{Path: fields[len(fields)-1], Size: size})
	}
	return entries
}

// resticManifestEntries — `restic ls <id>`: пути построчно, размер не отдаёт.
func resticManifestEntries(ctx context.Context, spec *ResticRepo, snapshotID string) ([]ManifestEntry, error) {
	cmd := resticCmd(ctx, spec, "ls", snapshotID)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("restic ls: %w", err)
	}
	return parseResticLs(string(out)), nil
}

func parseResticLs(out string) []ManifestEntry {
	var entries []ManifestEntry
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// корень снапшота (первый уровень) не включаем
		if strings.Count(strings.Trim(line, "/"), "/") == 0 {
			continue
		}
		entries = append(entries, ManifestEntry{Path: line})
	}
	return entries
}
