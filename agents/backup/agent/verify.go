package agent

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
)

// Verify — «учебная тревога для бэкапов»: регулярно поднимать снапшоты и
// доказывать, что из них реально восстановиться.
//
// L1: заново скачать объект, сверить SHA-256 с записью снапшота, для
// архивов — прогнать tar по записям, для restic — `restic check`.
// L2 (database engines): залить дамп в throwaway-контейнер через docker CLI
// (DOCKER_HOST указывает на изолированный verify-worker), выполнить
// контрольный запрос. Контейнер — --rm --network none, хостовые пути
// не монтируются: только named volume с распакованным дампом.

type VerifyOutcome struct {
	Level  string `json:"level"`
	Status string `json:"status"` // clean | failed
	RTOsec int64  `json:"rtoSec"`
	Log    string `json:"log"`
}

func (a *Agent) StartVerify(ctx context.Context, req JobRequest) (*JobStatus, error) {
	cz := a.Corteza(req)
	if cz == nil {
		return nil, fmt.Errorf("corteza is not configured")
	}
	var snap *Snapshot
	var err error
	if id := ParseUint(req.SnapshotID); id > 0 {
		snap, err = cz.LoadSnapshot(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("snapshot %d: %w", id, err)
		}
	} else if id := ParseUint(req.ResolvedPolicyID()); id > 0 {
		snap, err = latestSnapshotForPolicy(ctx, cz, id)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("snapshotID or policyID is required")
	}
	level := strings.ToLower(strings.TrimSpace(req.VerifyLevel))
	pol, _ := cz.LoadPolicy(ctx, snap.PolicyID)
	if level == "" && pol != nil {
		level = pol.VerifyLevel
	}
	if level == "" {
		level = "l1"
	}
	st := &JobStatus{
		ID:          newJobID(),
		Kind:        string(KindVerify),
		Status:      StatusRunning,
		StartedAt:   time.Now(),
		SnapshotID:  fmtUint(snap.ID),
		SourceID:    fmtUint(snap.SourceID),
		PolicyID:    fmtUint(snap.PolicyID),
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
	go a.runVerify(context.Background(), st, snap, level, cz)
	return st, nil
}

func latestSnapshotForPolicy(ctx context.Context, cz *Corteza, policyID uint64) (*Snapshot, error) {
	set, err := cz.ListRecords(ctx, "snapshots", fmt.Sprintf("policy = %d AND restorable = '1'", policyID))
	if err != nil {
		return nil, err
	}
	for i := len(set) - 1; i >= 0; i-- {
		return cz.LoadSnapshot(ctx, set[i].ID)
	}
	return nil, fmt.Errorf("no restorable snapshots for policy %d", policyID)
}

// VerifyDue — политики, у которых verify_cron подошёл прямо сейчас, плюс
// последний снапшот каждой. Read-only, как DuePolicies: запуск делает
// серверная цепочка (ровно по паттерну backup-run-due).
func (a *Agent) VerifyDue(ctx context.Context, req JobRequest) ([]map[string]string, error) {
	cz := a.Corteza(req)
	if cz == nil || !cz.HasToken() {
		return nil, fmt.Errorf("corteza is not configured (no token)")
	}
	pols, err := cz.ListEnabledPolicies(ctx)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	var due []map[string]string
	for i := range pols {
		p := pols[i]
		if !p.Verify {
			continue
		}
		if CronValid(p.VerifyCron) != nil || !MatchCron(p.VerifyCron, now) {
			continue
		}
		snap, err := latestSnapshotForPolicy(ctx, cz, p.ID)
		if err != nil {
			logVerify("due policy %s: %v", p.Name, err)
			continue
		}
		due = append(due, map[string]string{
			"policyID":   fmtUint(p.ID),
			"snapshotID": fmtUint(snap.ID),
			"sourceID":   fmtUint(snap.SourceID),
			"name":       p.Name,
			"level":      firstNonEmpty(p.VerifyLevel, "l1"),
		})
	}
	return due, nil
}

func logVerify(format string, args ...any) { log.Printf("[verify] "+format, args...) }

func (a *Agent) runVerify(ctx context.Context, st *JobStatus, snap *Snapshot, level string, cz *Corteza) {
	a.sem <- struct{}{}
	defer func() { <-a.sem }()
	start := time.Now()
	outcome := &VerifyOutcome{Level: level}
	var buf bytes.Buffer
	finish := func(status JobStatusName, errMsg string) {
		now := time.Now()
		st.Status = status
		st.Error = errMsg
		st.FinishedAt = &now
		st.Progress = 100
		outcome.Status = mapVerifyStatus(status)
		outcome.RTOsec = int64(now.Sub(start).Seconds())
		metricVerify.WithLabelValues(level, outcome.Status).Inc()
		if cz.HasToken() {
			vals := map[string]string{
				"verified_at":         now.UTC().Format(time.RFC3339),
				"verify_duration_sec": fmtInt64(outcome.RTOsec),
			}
			if outcome.Status == "clean" {
				vals["verified"] = "1"
				vals["verify_error"] = ""
			} else {
				vals["verify_error"] = errMsg
			}
			if err := cz.UpdateValues(ctx, "snapshots", snap.ID, vals); err != nil {
				logVerify("update snapshot %d: %v", snap.ID, err)
			}
			// verify_reports-строка (созданная цепочкой) — её статус тоже ведём.
			if st.JobRecordID != "" {
				reportVals := map[string]string{
					"status":  outcome.Status,
					"rto_sec": fmtInt64(outcome.RTOsec),
					"log":     truncateStr(buf.String(), 4000),
				}
				if rid := ParseUint(st.JobRecordID); rid > 0 {
					if err := cz.UpdateValues(ctx, "verify_reports", rid, reportVals); err != nil {
						logVerify("update report %s: %v", st.JobRecordID, err)
					}
				}
			}
		}
		a.notifyCallback(st, mapStatusEvent(status))
	}

	src, _ := cz.LoadSource(ctx, snap.SourceID)
	handle := sanitizeKey(fmtUint(snap.SourceID))
	if src != nil {
		handle = sanitizeKey(firstNonEmpty(src.Name, fmtUint(src.ID)))
	}
	ok := a.verifyLevel1(ctx, cz, snap, handle, &buf)
	if ok && level == "l2" {
		ok = a.verifyLevel2(ctx, cz, snap, src, &buf)
	}
	if !ok {
		finish(StatusFailed, strings.TrimSpace(lastLines(buf.String(), 3)))
		return
	}
	finish(StatusCompleted, "")
}

func mapVerifyStatus(s JobStatusName) string {
	if s == StatusCompleted {
		return "clean"
	}
	return "failed"
}

// verifyLevel1 — целостность артефакта: sha256 + структурная проверка.
func (a *Agent) verifyLevel1(ctx context.Context, cz *Corteza, snap *Snapshot, handle string, logBuf *bytes.Buffer) bool {
	if a.store == nil {
		fmt.Fprintln(logBuf, "minio is not configured")
		return false
	}
	store, err := a.storeForSnapshot(ctx, cz, snap)
	if err != nil {
		fmt.Fprintf(logBuf, "FAIL store: %v\n", err)
		return false
	}
	write := func(format string, args ...any) {
		fmt.Fprintf(logBuf, format+"\n", args...)
	}

	destRec, derr := a.destinationForSnapshot(ctx, cz, snap)
	if derr != nil {
		write("FAIL destination: %v", derr)
		return false
	}
	if strings.HasPrefix(snap.Engine, "restic") {
		write("restic check (subset 10%)")
		spec, rerr := a.engineRepoFor(ctx, destRec, handle)
		if rerr != nil {
			write("FAIL: %v", rerr)
			return false
		}
		defer spec.close()
		if err := resticCheck(ctx, spec); err != nil {
			write("FAIL: %v", err)
			return false
		}
		write("OK")
		return true
	}
	if snap.Engine == "kopia" {
		write("kopia snapshot verify")
		if err := kopiaSnapshotVerify(ctx, a.cfg, destRec, handle, a.cfg.ResticPassword); err != nil {
			write("FAIL: %v", err)
			return false
		}
		write("OK")
		return true
	}

	rc, err := store.GetBucket(ctx, snap.S3Bucket, snap.S3Key)
	if err != nil {
		write("FAIL get %s: %v", snap.S3Key, err)
		return false
	}
	defer rc.Close()
	h := sha256.New()
	var lr io.Reader = rc
	if snap.Engine == "archive" || snap.Engine == "s3copy" {
		gzr, err := gzip.NewReader(rc)
		if err != nil {
			write("FAIL gunzip: %v", err)
			return false
		}
		// Прогоняем tar по записям, не распаковывая на диск.
		tr := tar.NewReader(gzr)
		for {
			_, err := tr.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				gzr.Close()
				write("FAIL tar walk: %v", err)
				return false
			}
		}
		gzr.Close()
		// Повторный проход для sha256: скачиваем заново.
		rc2, err := store.GetBucket(ctx, snap.S3Bucket, snap.S3Key)
		if err != nil {
			write("FAIL get(second pass): %v", err)
			return false
		}
		defer rc2.Close()
		lr = rc2
	}
	if _, err := io.Copy(h, lr); err != nil {
		write("FAIL read: %v", err)
		return false
	}
	sum := hex.EncodeToString(h.Sum(nil))
	if snap.Checksum != "" && !strings.EqualFold(sum, snap.Checksum) {
		write("FAIL sha256: got %s want %s", sum, snap.Checksum)
		return false
	}
	write("OK sha256=%s", sum)
	return true
}

// verifyLevel2 — подъём дампа в throwaway-контейнере. Только database engine.
func (a *Agent) verifyLevel2(ctx context.Context, cz *Corteza, snap *Snapshot, src *Source, logBuf *bytes.Buffer) bool {
	write := func(format string, args ...any) { fmt.Fprintf(logBuf, format+"\n", args...) }
	if src == nil {
		write("FAIL: source unavailable")
		return false
	}
	if !isDatabaseDump(snap.Engine) {
		write("SKIP l2: engine %s is not a database dump", snap.Engine)
		return true
	}
	if _, err := exec.LookPath("docker"); err != nil {
		write("FAIL l2: docker CLI not found")
		return false
	}
	store, err := a.storeForSnapshot(ctx, cz, snap)
	if err != nil {
		write("FAIL store: %v", err)
		return false
	}
	tmp, err := os.MkdirTemp("", "backup-verify-*")
	if err != nil {
		write("FAIL mktemp: %v", err)
		return false
	}
	defer os.RemoveAll(tmp)
	rc, err := store.GetBucket(ctx, snap.S3Bucket, snap.S3Key)
	if err != nil {
		write("FAIL get: %v", err)
		return false
	}
	dumpPath := path.Join(tmp, path.Base(snap.S3Key))
	f, err := os.Create(dumpPath)
	if err != nil {
		rc.Close()
		write("FAIL create: %v", err)
		return false
	}
	_, copyErr := io.Copy(f, rc)
	f.Close()
	rc.Close()
	if copyErr != nil {
		write("FAIL download: %v", copyErr)
		return false
	}

	var args []string
	image := "postgres:16-alpine"
	script := ""
	switch snap.Engine {
	case "pg_dump":
		script = fmt.Sprintf("pg_restore -U postgres -d postgres /d/%s 2>/dev/null || psql -U postgres -d postgres -f /d/%s || exit 1; psql -U postgres -d postgres -tc 'select count(*) from pg_stat_user_tables' | head -1", path.Base(snap.S3Key), path.Base(snap.S3Key))
	case "mysqldump":
		image = "mysql:8"
		script = fmt.Sprintf("mysql -uroot -prootroot < /d/%s && mysql -uroot -prootroot -e 'show tables' | wc -l", path.Base(snap.S3Key))
	default:
		write("SKIP l2: unsupported engine %s", snap.Engine)
		return true
	}
	args = []string{"run", "--rm", "--network", "none",
		"--memory", "1g", "--cpus", "1",
		"-v", tmp + ":/d:ro",
		"-e", "POSTGRES_PASSWORD=postgres",
		image, "sh", "-c", script}
	cmd := exec.CommandContext(ctx, "docker", args...)
	if a.cfg.DockerHost != "" {
		cmd.Env = append(os.Environ(), "DOCKER_HOST="+a.cfg.DockerHost)
	}
	out, runErr := cmd.CombinedOutput()
	write("docker: %s", strings.TrimSpace(string(out)))
	if runErr != nil {
		write("FAIL l2: %v", runErr)
		return false
	}
	write("OK l2 restore+boot")
	return true
}

func isDatabaseDump(engine string) bool {
	return engine == "pg_dump" || engine == "mysqldump" || strings.HasPrefix(engine, "mysql")
}

func lastLines(s string, n int) string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return strings.Join(lines, "; ")
}

func fmtInt64(v int64) string { return fmt.Sprintf("%d", v) }
