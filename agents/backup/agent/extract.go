package agent

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"
)

// Browse и Extract — точечная работа с содержимым снапшота без полного
// восстановления: что внутри (P5) и «верни мне один файл» (P2).
// Browse — sync (список), Extract — async (может качать много).

// ArchiveEntry — элемент листинга содержимого снапшота.
type ArchiveEntry struct {
	Name    string    `json:"name"`
	Size    int64     `json:"size,omitempty"`
	ModTime time.Time `json:"modTime,omitempty"`
	IsDir   bool      `json:"isDir,omitempty"`
}

// --- Browse ----------------------------------------------------------------

// Browse отдаёт список объектов снапшота. Движки: tar-архив (потоковый
// проход), s3copy (список ключей), restic/kopia (их ls), database —
// единственный «объект»-дамп.
func (a *Agent) Browse(ctx context.Context, req JobRequest) (any, error) {
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
	src, err := cz.LoadSource(ctx, snap.SourceID)
	if err != nil {
		return nil, fmt.Errorf("source %d: %w", snap.SourceID, err)
	}
	handle := sanitizeKey(firstNonEmpty(src.Name, fmtUint(src.ID)))
	prefix := strings.Trim(strings.TrimSpace(req.DestPath), "/") // путь-фильтр

	const limit = 2000
	var entries []ArchiveEntry
	store, err := a.storeForSnapshot(ctx, cz, snap)
	if err != nil {
		return nil, err
	}
	destRec, derr := a.destinationForSnapshot(ctx, cz, snap)
	if derr != nil {
		return nil, derr
	}
	switch {
	case snap.Engine == "archive" || snap.Engine == "k8s" || strings.HasPrefix(snap.Engine, "saas-"):
		rc, err := store.GetBucket(ctx, snap.S3Bucket, snap.S3Key)
		if err != nil {
			return nil, err
		}
		defer rc.Close()
		entries, err = browseTarGz(ctx, rc, prefix, limit)
		if err != nil {
			return nil, err
		}
	case snap.Engine == "s3copy":
		keys, err := store.ListBucket(ctx, snap.S3Bucket, snap.S3Key)
		if err != nil {
			return nil, err
		}
		for _, k := range keys {
			if prefix != "" && !strings.Contains(k, prefix) {
				continue
			}
			entries = append(entries, ArchiveEntry{Name: k})
			if len(entries) >= limit {
				break
			}
		}
	case snap.Engine == "restic":
		spec, rerr := a.engineRepoFor(ctx, destRec, handle)
		if rerr != nil {
			return nil, rerr
		}
		defer spec.close()
		entries, err = browseRestic(ctx, spec, snap.ResticID)
	case snap.Engine == "kopia":
		if kerr := kopiaConnect(ctx, a.cfg, destRec, handle, a.cfg.ResticPassword); kerr != nil {
			return nil, kerr
		}
		entries, err = browseKopia(ctx, a.cfg, handle, snap.ResticID)
	default:
		if isDatabaseDump(snap.Engine) {
			entries = []ArchiveEntry{{Name: path.Base(snap.S3Key), Size: snap.SizeBytes}}
		} else {
			return nil, fmt.Errorf("browse not supported for engine %q", snap.Engine)
		}
	}
	if entries == nil {
		entries = []ArchiveEntry{}
	}
	return map[string]any{
		"snapshotID": fmtUint(snap.ID),
		"engine":     snap.Engine,
		"count":      len(entries),
		"truncated":  len(entries) >= limit,
		"entries":    entries,
	}, nil
}

// browseTarGz — потоковый листинг tar.gz с фильтром по префиксу пути.
func browseTarGz(ctx context.Context, r io.Reader, prefix string, limit int) ([]ArchiveEntry, error) {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("gunzip: %w", err)
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	var out []ArchiveEntry
	for len(out) < limit {
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
		name := path.Clean(hdr.Name)
		if prefix != "" && name != prefix && !strings.HasPrefix(name, prefix+"/") {
			continue
		}
		out = append(out, ArchiveEntry{
			Name:    name,
			Size:    hdr.Size,
			ModTime: hdr.ModTime,
			IsDir:   hdr.FileInfo().IsDir(),
		})
	}
	return out, nil
}

// browseRestic парсит построчный вывод `restic ls <id>`.
func browseRestic(ctx context.Context, spec *ResticRepo, snapshotID string) ([]ArchiveEntry, error) {
	cmd := resticCmd(ctx, spec, "ls", snapshotID)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("restic ls: %w", err)
	}
	return parseLsOutput(out), nil
}

// browseKopia парсит построчный вывод `kopia ls <id>`.
func browseKopia(ctx context.Context, cfg Config, handle, snapshotID string) ([]ArchiveEntry, error) {
	cfgPath := kopiaConfigPath(handle)
	cmd := kopiaCLI(ctx, cfg, cfgPath, "ls", snapshotID)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("kopia ls: %w", err)
	}
	return parseLsOutput(out), nil
}

func parseLsOutput(out []byte) []ArchiveEntry {
	var entries []ArchiveEntry
	sc := bufio.NewScanner(strings.NewReader(string(out)))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "/") && strings.Count(line, "/") == 1 {
			// корень репозитория не показываем
			if strings.Count(strings.Trim(line, "/"), "/") == 0 {
				continue
			}
		}
		entries = append(entries, ArchiveEntry{Name: line})
	}
	return entries
}

// --- Extract ---------------------------------------------------------------

// StartExtract достаёт один файл (или поддерево) из снапшота.
// req.DestPath — путь внутри архива; req.ExtractTo — куда писать (пусто =
// destType download: вернуть содержимое текстом, лимит 1 МБ).
func (a *Agent) StartExtract(ctx context.Context, req JobRequest) (*JobStatus, error) {
	cz := a.Corteza(req)
	if cz == nil {
		return nil, fmt.Errorf("corteza is not configured")
	}
	snapID := ParseUint(req.SnapshotID)
	if snapID == 0 {
		return nil, fmt.Errorf("snapshotID is required")
	}
	if strings.TrimSpace(req.DestPath) == "" {
		return nil, fmt.Errorf("path (inside snapshot) is required")
	}
	snap, err := cz.LoadSnapshot(ctx, snapID)
	if err != nil {
		return nil, fmt.Errorf("snapshot %d: %w", snapID, err)
	}
	st := &JobStatus{
		ID:          newJobID(),
		Kind:        string(KindExtract),
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
	go a.runExtract(context.Background(), st, snap, req, cz)
	return st, nil
}

func (a *Agent) runExtract(ctx context.Context, st *JobStatus, snap *Snapshot, req JobRequest, cz *Corteza) {
	a.sem <- struct{}{}
	defer func() { <-a.sem }()
	src, _ := cz.LoadSource(ctx, snap.SourceID)
	handle := sanitizeKey(fmtUint(snap.SourceID))
	if src != nil {
		handle = sanitizeKey(firstNonEmpty(src.Name, fmtUint(src.ID)))
	}
	entryPath := strings.Trim(strings.TrimSpace(req.DestPath), "/")
	download := strings.EqualFold(strings.TrimSpace(req.ExtractTo), "")

	finish := func(status JobStatusName, errMsg string) {
		now := time.Now()
		st.Status = status
		st.Error = errMsg
		st.FinishedAt = &now
		st.Progress = 100
		a.notifyCallback(st, mapStatusEvent(status))
	}

	var (
		files  int
		n      int64
		inline string
		err    error
	)
	store, serr := a.storeForSnapshot(ctx, cz, snap)
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
		rc, gerr := store.GetBucket(ctx, snap.S3Bucket, snap.S3Key)
		if gerr != nil {
			finish(StatusFailed, gerr.Error())
			return
		}
		defer rc.Close()
		if download {
			inline, n, err = extractInlineTarGz(ctx, rc, entryPath)
			st.Message = fmt.Sprintf("extracted %s (%d bytes)", entryPath, n)
		} else {
			files, n, err = extractFromTarGz(ctx, rc, entryPath, req.ExtractTo)
			st.Message = fmt.Sprintf("extracted %d file(s), %d bytes → %s", files, n, req.ExtractTo)
		}
		st.Files = files
		st.BytesWritten = n
	case snap.Engine == "s3copy":
		// Путь внутри = суффикс ключа префикса.
		key := strings.TrimSuffix(snap.S3Key, "/") + "/" + entryPath
		rc, gerr := store.GetBucket(ctx, snap.S3Bucket, key)
		if gerr != nil {
			finish(StatusFailed, gerr.Error())
			return
		}
		defer rc.Close()
		if download {
			data, rerr := io.ReadAll(io.LimitReader(rc, 1<<20))
			inline, n, err = string(data), int64(len(data)), rerr
		} else {
			files, n, err = writeExtracted(rc, req.ExtractTo, path.Base(entryPath))
		}
		st.Files = files
		st.BytesWritten = n
	case snap.Engine == "restic":
		if download {
			err = fmt.Errorf("download not supported for restic engine; extract to a path")
			break
		}
		spec, rerr := a.engineRepoFor(ctx, destRec, handle)
		if rerr != nil {
			err = rerr
			break
		}
		defer spec.close()
		err = resticExtract(ctx, spec, snap.ResticID, entryPath, req.ExtractTo)
	case snap.Engine == "kopia":
		if download {
			err = fmt.Errorf("download not supported for kopia engine; extract to a path")
			break
		}
		err = kopiaSnapshotRestore(ctx, a.cfg, destRec, handle, snap.ResticID, req.ExtractTo, entryPath, a.cfg.ResticPassword)
	case isDatabaseDump(snap.Engine):
		err = fmt.Errorf("extract not applicable to database dump; restore the whole dump")
	default:
		err = fmt.Errorf("extract not supported for engine %q", snap.Engine)
	}
	if err != nil {
		finish(StatusFailed, err.Error())
		return
	}
	if download {
		st.Message = truncateStr(inline, 512)
	}
	finish(StatusCompleted, "")
	st.ExtractInline = inline
}

// extractFromTarGz потоково вытаскивает файл/поддерево в destDir.
func extractFromTarGz(ctx context.Context, r io.Reader, entryPath, destDir string) (files int, n int64, err error) {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return 0, 0, fmt.Errorf("gunzip: %w", err)
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	matched := false
	for {
		if err := ctx.Err(); err != nil {
			return files, n, err
		}
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return files, n, fmt.Errorf("tar: %w", err)
		}
		name := path.Clean(hdr.Name)
		if name != entryPath && !strings.HasPrefix(name, entryPath+"/") {
			continue
		}
		matched = true
		if hdr.FileInfo().IsDir() {
			continue
		}
		dst, err := safeJoin(destDir, strings.TrimPrefix(name, entryPath+"/"))
		if name == entryPath {
			dst = path.Join(destDir, path.Base(name))
		}
		if err != nil {
			return files, n, err
		}
		if err := os.MkdirAll(path.Dir(dst), 0o755); err != nil {
			return files, n, err
		}
		f, err := os.Create(dst)
		if err != nil {
			return files, n, err
		}
		w, err := io.Copy(f, tr)
		f.Close()
		if err != nil {
			return files, n, err
		}
		files++
		n += w
	}
	if !matched {
		return files, n, fmt.Errorf("entry %q not found in snapshot", entryPath)
	}
	return files, n, nil
}

// extractInlineTarGz возвращает содержимое одного файла строкой (лимит 1 МБ).
func extractInlineTarGz(ctx context.Context, r io.Reader, entryPath string) (string, int64, error) {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return "", 0, fmt.Errorf("gunzip: %w", err)
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		if err := ctx.Err(); err != nil {
			return "", 0, err
		}
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", 0, fmt.Errorf("tar: %w", err)
		}
		if path.Clean(hdr.Name) != entryPath {
			continue
		}
		data, err := io.ReadAll(io.LimitReader(tr, 1<<20))
		return string(data), int64(len(data)), err
	}
	return "", 0, fmt.Errorf("entry %q not found in snapshot", entryPath)
}

// writeExtracted пишет поток в destDir/name.
func writeExtracted(r io.Reader, destDir, name string) (int, int64, error) {
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return 0, 0, err
	}
	dst := path.Join(destDir, name)
	f, err := os.Create(dst)
	if err != nil {
		return 0, 0, err
	}
	defer f.Close()
	n, err := io.Copy(f, r)
	return 1, n, err
}
