package agent

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/madnikulin50/lowcode/agents/sdk"
)

type Agent struct {
	cfg   Config
	cz    *Corteza
	store *ObjectStore
	mu    sync.RWMutex
	jobs  map[string]*JobStatus
	sem   chan struct{}
	cb    *sdk.Callback
}

func New(cfg Config, store *ObjectStore, cz *Corteza) *Agent {
	n := cfg.Concurrency
	if n <= 0 {
		n = 2
	}
	return &Agent{
		cfg:   cfg,
		cz:    cz,
		store: store,
		jobs:  map[string]*JobStatus{},
		sem:   make(chan struct{}, n),
		cb:    sdk.NewCallback(),
	}
}

func (a *Agent) Corteza(req JobRequest) *Corteza {
	cz := a.cz
	if cz == nil {
		return nil
	}
	if req.Token != "" {
		cz = cz.WithToken(req.Token)
	}
	if ns := ParseUint(req.NamespaceID); ns > 0 {
		cz = cz.WithNamespace(ns)
	}
	return cz
}

func (a *Agent) GetStatus(id string) *JobStatus {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.jobs[id]
}

func (a *Agent) ListJobsStatus() []*JobStatus {
	a.mu.RLock()
	defer a.mu.RUnlock()
	out := make([]*JobStatus, 0, len(a.jobs))
	for _, j := range a.jobs {
		cp := *j
		out = append(out, &cp)
	}
	return out
}

func (a *Agent) Health(ctx context.Context) map[string]interface{} {
	minioOK := a.store != nil && a.store.Ping(ctx) == nil
	return map[string]interface{}{
		"status":      "ok",
		"minio":       minioOK,
		"restic":      resticAvailable(),
		"hostname":    hostname(),
		"jobsRunning": a.runningCount(),
	}
}

func (a *Agent) runningCount() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	n := 0
	for _, j := range a.jobs {
		if j.Status == StatusRunning {
			n++
		}
	}
	return n
}

// ReconcileStaleJobs marks "jobs"/"restores" records a *previous* process
// left in "running" as failed. Job state only ever lives in this process's
// in-memory a.jobs map — a fresh process starts with that map empty, so any
// record still "running" at this point cannot belong to it: the process
// that owned it crashed, was redeployed, or was killed before it reached
// finish()/persistJob(), and nothing was ever going to flip that status
// again.
//
// It takes an explicit *Corteza (rather than always using a.cz) because it
// runs two ways: once at startup using the agent's own static token (when
// configured — see main.go), and on demand via StartReconcile/Call using
// whatever token the caller's request carries — deployments that skip a
// static agent token (cron polling off, "Compose buttons send token in the
// job POST") never get the startup pass, so the on-demand path is the only
// one that ever runs for them.
func (a *Agent) ReconcileStaleJobs(ctx context.Context, cz *Corteza) (int, error) {
	if cz == nil || !cz.HasToken() {
		return 0, fmt.Errorf("corteza is not configured (no token)")
	}
	const staleMsg = "backup agent restarted while this job was running; marked as failed"
	now := time.Now().UTC().Format(time.RFC3339)
	n := 0
	for _, handle := range []string{"jobs", "restores"} {
		set, err := cz.ListRecords(ctx, handle, "status = 'running'")
		if err != nil {
			return n, fmt.Errorf("%s: %w", handle, err)
		}
		for i := range set {
			id := set[i].ID
			if err := cz.UpdateValues(ctx, handle, id, map[string]string{
				"status":      string(StatusFailed),
				"error":       staleMsg,
				"finished_at": now,
			}); err != nil {
				log.Printf("reconcile %s %d: %v", handle, id, err)
				continue
			}
			log.Printf("reconcile: %s %d was stuck running, marked failed", handle, id)
			n++
		}
	}
	return n, nil
}

// StartScheduler keeps the agent's heartbeat going (presence + capabilities
// in the "agents" module). It used to also decide which backup policies
// were due and start them directly — that meant this agent process was the
// only thing that would ever write the resulting "jobs" row back to
// "completed"/"failed", so an agent restart mid-backup left it stuck
// "running" forever. Policy scheduling now lives server-side: the
// "backup-run-due" rule chain (see agents/backup/compose/chains.mjs /
// compose/mcp/backup_chains.go) polls DuePolicies, creates the Compose row
// itself, and tracks status via the same poll/ingest mechanism used for
// interactively-started backups. Trigger that chain on a schedule (cron,
// systemd timer, Corteza Automation) instead of relying on this loop.
func (a *Agent) StartScheduler(ctx context.Context) {
	a.Heartbeat(ctx)
	hb := time.NewTicker(2 * time.Minute)
	defer hb.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-hb.C:
			a.Heartbeat(ctx)
		}
	}
}

func (a *Agent) Heartbeat(ctx context.Context) {
	if a.cz == nil || !a.cz.HasToken() {
		return
	}
	caps := strings.Join(sdk.CapabilitiesOf(Components()), ",")
	if resticAvailable() {
		caps += ",restic"
	}
	if err := a.cz.UpsertAgent(ctx, "backup-agent", a.cfg.PublicURL, hostname(), caps); err != nil {
		log.Printf("heartbeat: %v", err)
	}
}

// DuePolicies reports enabled policies whose cron schedule matches right
// now. Read-only by design: it starts nothing and writes nothing — the
// caller (the "backup-run-due" rule chain) is the one that creates the
// Compose "jobs" row and starts the backup, so that row is always owned by
// something that can reliably flip it to "completed"/"failed" (see
// StartScheduler's doc comment).
//
// A policy already backing up (an existing "jobs" row for it with
// status="running") is excluded, so calling this repeatedly — e.g. every
// minute from an external cron — never starts the same policy twice while
// its previous run is still in flight.
func (a *Agent) DuePolicies(ctx context.Context, req JobRequest) ([]DuePolicy, error) {
	cz := a.Corteza(req)
	if cz == nil {
		return nil, fmt.Errorf("corteza is not configured")
	}
	if !cz.HasToken() {
		return nil, fmt.Errorf("no API token")
	}
	pols, err := cz.ListEnabledPolicies(ctx)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	var due []DuePolicy
	for i := range pols {
		p := pols[i]
		if CronValid(p.Cron) != nil || !MatchCron(p.Cron, now) {
			continue
		}
		// Nothing updates policies.last_run any more (see recordSnapshot's
		// doc comment) — a currently-running job for this policy is the
		// real, live signal that it's not due again, so that's the only
		// guard left against double-firing.
		running, err := cz.ListRecords(ctx, "jobs", fmt.Sprintf("policy = %d AND status = 'running'", p.ID))
		if err != nil {
			log.Printf("due policy %s: check running: %v", p.Name, err)
			continue
		}
		if len(running) > 0 {
			continue
		}
		due = append(due, DuePolicy{ID: fmtUint(p.ID), SourceID: fmtUint(p.SourceID), Name: p.Name})
	}
	return due, nil
}

func (a *Agent) StartBackup(ctx context.Context, req JobRequest) (*JobStatus, error) {
	if a.store == nil {
		return nil, fmt.Errorf("minio is not configured")
	}
	cz := a.Corteza(req)
	if cz == nil {
		return nil, fmt.Errorf("corteza is not configured")
	}
	srcID := ParseUint(req.ResolvedSourceID())
	polID := ParseUint(req.ResolvedPolicyID())
	var pol *Policy
	var err error
	if polID > 0 {
		pol, err = cz.LoadPolicy(ctx, polID)
		if err != nil {
			return nil, fmt.Errorf("policy %d: %w", polID, err)
		}
		if srcID == 0 {
			srcID = pol.SourceID
		}
	} else if srcID > 0 {
		pol, err = cz.FindPolicyForSource(ctx, srcID)
		if err != nil {
			pol = &Policy{SourceID: srcID, Cron: "0 2 * * *", RetentionDays: 14}
		}
	} else {
		return nil, fmt.Errorf("sourceID or policyID is required")
	}
	src, err := cz.LoadSource(ctx, srcID)
	if err != nil {
		return nil, fmt.Errorf("source %d: %w", srcID, err)
	}
	cred, err := cz.LoadCredential(ctx, src.CredID)
	if err != nil {
		cred = &Credential{}
	}

	id := uuid.New().String()
	st := &JobStatus{
		ID:          id,
		Kind:        string(KindFull),
		Status:      StatusRunning,
		StartedAt:   time.Now(),
		SourceID:    fmtUint(src.ID),
		PolicyID:    fmtUint(pol.ID),
		JobRecordID: req.JobRecordID,
		CallbackURL: req.CallbackURL,
		Token:       firstNonEmpty(req.Token, a.cfg.Token),
		NamespaceID: ParseUint(req.NamespaceID),
	}
	if st.NamespaceID == 0 {
		st.NamespaceID = a.cfg.NamespaceID
	}
	if pol.Incremental {
		st.Kind = string(KindIncremental)
	}

	// st.JobRecordID (the Compose "jobs" row) is expected to come from the
	// caller — the "backup-run-source"/"backup-run-policy"/"backup-run-due"
	// chains all create it *before* calling this endpoint, and the poll/
	// ingest mechanism they set up afterwards is what tracks status from
	// here. This agent no longer creates or updates that row itself (see
	// StartScheduler's doc comment): a job started without a JobRecordID
	// simply isn't linked to Compose — still trackable via GetStatus/
	// ListJobsStatus while this process is alive, nothing more.

	a.mu.Lock()
	a.jobs[id] = st
	a.mu.Unlock()

	go a.runBackup(context.Background(), st, src, pol, cred, cz)
	return st, nil
}

func (a *Agent) StartRestore(ctx context.Context, req JobRequest) (*JobStatus, error) {
	cz := a.Corteza(req)
	if cz == nil || a.store == nil {
		return nil, fmt.Errorf("store/corteza is not configured")
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
	cred, _ := cz.LoadCredential(ctx, src.CredID)
	id := uuid.New().String()
	st := &JobStatus{
		ID:          id,
		Kind:        string(KindRestore),
		Status:      StatusRunning,
		StartedAt:   time.Now(),
		SourceID:    fmtUint(src.ID),
		SnapshotID:  fmtUint(snap.ID),
		JobRecordID: req.RestoreID,
		CallbackURL: req.CallbackURL,
		Token:       firstNonEmpty(req.Token, a.cfg.Token),
		NamespaceID: ParseUint(req.NamespaceID),
		S3Bucket:    snap.S3Bucket,
		S3Key:       snap.S3Key,
		Engine:      snap.Engine,
	}
	// See StartBackup: st.JobRecordID comes from the caller (the
	// "backup-restore" chain's "record" step), not created here.
	a.mu.Lock()
	a.jobs[id] = st
	a.mu.Unlock()
	go a.runRestore(context.Background(), st, src, snap, cred, req)
	return st, nil
}

func (a *Agent) StartPrune(ctx context.Context, req JobRequest) (*JobStatus, error) {
	cz := a.Corteza(req)
	if cz == nil || a.store == nil {
		return nil, fmt.Errorf("store/corteza is not configured")
	}
	id := uuid.New().String()
	st := &JobStatus{
		ID:          id,
		Kind:        string(KindPrune),
		Status:      StatusRunning,
		StartedAt:   time.Now(),
		PolicyID:    req.PolicyID,
		SourceID:    req.SourceID,
		CallbackURL: req.CallbackURL,
		Token:       firstNonEmpty(req.Token, a.cfg.Token),
		NamespaceID: ParseUint(req.NamespaceID),
	}
	a.mu.Lock()
	a.jobs[id] = st
	a.mu.Unlock()
	go a.runPrune(context.Background(), st, req, cz)
	return st, nil
}

func (a *Agent) runBackup(ctx context.Context, st *JobStatus, src *Source, pol *Policy, cred *Credential, cz *Corteza) {
	a.sem <- struct{}{}
	defer func() { <-a.sem }()
	label := string(src.Type)
	if label == "" {
		label = "unknown"
	}
	handle := sanitizeKey(firstNonEmpty(src.Name, fmtUint(src.ID)))
	metricInProgress.WithLabelValues(handle).Set(1)
	defer metricInProgress.WithLabelValues(handle).Set(0)
	start := time.Now()

	finish := func(status JobStatusName, errMsg string, res *BackupResult) {
		now := time.Now()
		st.Status = status
		st.Error = errMsg
		st.FinishedAt = &now
		st.Progress = 100
		if res != nil {
			st.BytesRead = res.BytesRead
			st.BytesWritten = res.BytesWritten
			st.Files = res.Files
			st.Checksum = res.Checksum
			st.Engine = res.Engine
			st.S3Bucket = res.Bucket
			st.S3Key = res.Key
			st.ResticID = res.ResticID
			st.Kind = string(res.Kind)
		}
		kind := "complete"
		if status == StatusFailed {
			kind = "failed"
			metricJobs.WithLabelValues(label, "failed").Inc()
		} else {
			metricJobs.WithLabelValues(label, "completed").Inc()
			metricLastSuccess.WithLabelValues(handle).Set(float64(now.Unix()))
			if res != nil {
				metricBytesRead.WithLabelValues(label).Add(float64(res.BytesRead))
				metricBytesWritten.WithLabelValues(label).Add(float64(res.BytesWritten))
			}
		}
		metricDuration.WithLabelValues(label).Observe(time.Since(start).Seconds())
		a.recordSnapshot(ctx, cz, st, pol, res)
		a.notifyCallback(st, kind)
	}

	if err := a.store.EnsureBucket(ctx); err != nil {
		finish(StatusFailed, err.Error(), nil)
		return
	}

	secret := ResolveSecret(cred.Handle)
	username := cred.Username
	extra := ""
	if cred != nil {
		extra = cred.Extra
	}
	res, err := a.backupSource(ctx, st, src, pol, username, secret, extra, handle)
	if err != nil {
		finish(StatusFailed, err.Error(), res)
		return
	}
	finish(StatusCompleted, "", res)
}

func (a *Agent) backupSource(ctx context.Context, st *JobStatus, src *Source, pol *Policy, username, secret, extra, handle string) (*BackupResult, error) {
	useRestic := pol.Incremental && resticAvailable() && src.Type == SourceFS
	if useRestic {
		st.Message = "restic backup"
		id, err := resticBackupPath(ctx, a.cfg, handle, src.Path, a.cfg.ResticPassword)
		if err != nil {
			return nil, err
		}
		res := &BackupResult{
			Bucket:   a.store.Bucket(),
			Key:      resticKeyHint(handle, id),
			Engine:   "restic",
			Kind:     KindIncremental,
			ResticID: id,
		}
		return res, nil
	}

	jobTag := firstNonEmpty(st.JobRecordID, st.ID[:8])
	switch src.Type {
	case SourceFS:
		walker, err := NewLocalFS(src.Path)
		if err != nil {
			return nil, err
		}
		return a.packToMinio(ctx, st, walker, handle, jobTag, "archive.tar.gz")
	case SourceSMB:
		host, share, root, err := ParseUNC(src.Path, src.Host, src.Share, src.SMBPath)
		if err != nil {
			return nil, err
		}
		smb, err := DialSMB(host, username, secret, extra, share, root)
		if err != nil {
			return nil, err
		}
		defer smb.Close()
		return a.packToMinio(ctx, st, smb, handle, jobTag, "archive.tar.gz")
	case SourceDatabase:
		return a.dumpToMinio(ctx, st, src, username, secret, handle, jobTag)
	case SourceS3:
		prefix := ObjectKey(handle, jobTag, "s3")
		on := func(_ string, n int64) {
			st.BytesWritten += n
			st.Files++
			st.Progress = 50
			a.notifyCallback(st, "progress")
		}
		res, err := CopyS3(ctx, *src, username, secret, a.store, prefix, on)
		return &res, err
	default:
		return nil, fmt.Errorf("unknown source type %q", src.Type)
	}
}

func (a *Agent) packToMinio(ctx context.Context, st *JobStatus, walker FileWalker, handle, jobTag, filename string) (*BackupResult, error) {
	key := ObjectKey(handle, jobTag, filename)
	pr, pw := io.Pipe()
	ch := newCountingHasher(pw)
	var packErr error
	var files int
	var bytesRead int64
	go func() {
		files, bytesRead, packErr = PackTarGz(ctx, ch, walker, func(_ FileEntry, n int64) {
			st.Files++
			st.BytesRead += n
			if st.Progress < 90 {
				st.Progress += 0.5
			}
			a.notifyCallback(st, "progress")
		})
		if packErr != nil {
			pw.CloseWithError(packErr)
			return
		}
		pw.Close()
	}()
	written, err := a.store.Put(ctx, key, pr, -1, "application/gzip")
	if err != nil {
		return nil, err
	}
	if packErr != nil {
		return nil, packErr
	}
	return &BackupResult{
		Bucket:       a.store.Bucket(),
		Key:          key,
		BytesRead:    bytesRead,
		BytesWritten: written,
		Files:        files,
		Checksum:     ch.SumHex(),
		Engine:       "archive",
		Kind:         KindFull,
	}, nil
}

func (a *Agent) dumpToMinio(ctx context.Context, st *JobStatus, src *Source, username, secret, handle, jobTag string) (*BackupResult, error) {
	ext := "dump"
	if strings.HasPrefix(strings.ToLower(src.DBEngine), "mysql") {
		ext = "sql"
	} else {
		ext = "pgdump"
	}
	key := ObjectKey(handle, jobTag, "dump."+ext)
	pr, pw := io.Pipe()
	ch := newCountingHasher(pw)
	var dumpErr error
	var engine string
	go func() {
		engine, dumpErr = DumpDatabase(ctx, *src, username, secret, ch)
		if dumpErr != nil {
			pw.CloseWithError(dumpErr)
			return
		}
		pw.Close()
	}()
	st.Message = "dumping database"
	written, err := a.store.Put(ctx, key, pr, -1, "application/octet-stream")
	if err != nil {
		return nil, err
	}
	if dumpErr != nil {
		return nil, dumpErr
	}
	return &BackupResult{
		Bucket:       a.store.Bucket(),
		Key:          key,
		BytesRead:    ch.Bytes(),
		BytesWritten: written,
		Files:        1,
		Checksum:     ch.SumHex(),
		Engine:       engine,
		Kind:         KindFull,
	}, nil
}

func (a *Agent) runRestore(ctx context.Context, st *JobStatus, src *Source, snap *Snapshot, cred *Credential, req JobRequest) {
	a.sem <- struct{}{}
	defer func() { <-a.sem }()
	secret := ""
	username := ""
	if cred != nil {
		secret = ResolveSecret(cred.Handle)
		username = cred.Username
	}
	destType := firstNonEmpty(req.DestType, "path")
	dest := req.DestPath
	var err error
	switch destType {
	case "download":
		st.Message = snap.S3Key
	case "original":
		err = a.restoreOriginal(ctx, src, snap, username, secret)
	default:
		if dest == "" {
			dest = path.Join(os.TempDir(), "backup-restore", fmtUint(snap.ID))
		}
		err = a.restoreToPath(ctx, src, snap, dest, username, secret)
	}
	now := time.Now()
	st.FinishedAt = &now
	st.Progress = 100
	if err != nil {
		st.Status = StatusFailed
		st.Error = err.Error()
		metricRestores.WithLabelValues("failed").Inc()
		a.notifyCallback(st, "failed")
	} else {
		st.Status = StatusCompleted
		metricRestores.WithLabelValues("completed").Inc()
		a.notifyCallback(st, "complete")
	}
	// st.Status/Progress are already updated above — GetStatus/ListJobsStatus
	// reflect it immediately. The "backup-ingest-restore" chain's poller
	// picks it up from there and writes the "restores" row; this agent
	// doesn't write Compose directly (see StartScheduler's doc comment).
}

func (a *Agent) restoreOriginal(ctx context.Context, src *Source, snap *Snapshot, username, secret string) error {
	switch src.Type {
	case SourceFS:
		return a.restoreToPath(ctx, src, snap, src.Path, username, secret)
	case SourceDatabase:
		rc, err := a.store.Get(ctx, snap.S3Key)
		if err != nil {
			return err
		}
		defer rc.Close()
		return RestoreDatabase(ctx, *src, username, secret, rc, snap.Engine)
	case SourceS3:
		_, err := RestoreS3Copy(ctx, a.store, strings.TrimSuffix(snap.S3Key, "/"), src.Host, src.S3Bucket, username, secret, src.S3Prefix, src.S3Secure, src.S3Region)
		return err
	case SourceSMB:
		tmp := path.Join(os.TempDir(), "backup-restore-smb", fmtUint(snap.ID))
		return a.restoreToPath(ctx, src, snap, tmp, username, secret)
	default:
		return fmt.Errorf("restore original not supported for %s", src.Type)
	}
}

func (a *Agent) restoreToPath(ctx context.Context, src *Source, snap *Snapshot, dest, username, secret string) error {
	if snap.Engine == "restic" {
		handle := sanitizeKey(firstNonEmpty(src.Name, fmtUint(src.ID)))
		return resticRestore(ctx, a.cfg, handle, snap.ResticID, dest, a.cfg.ResticPassword)
	}
	if snap.Engine == "s3copy" {
		_, err := RestoreS3Copy(ctx, a.store, strings.TrimSuffix(snap.S3Key, "/"), src.Host, src.S3Bucket, username, secret, dest, src.S3Secure, src.S3Region)
		return err
	}
	rc, err := a.store.Get(ctx, snap.S3Key)
	if err != nil {
		return err
	}
	defer rc.Close()
	if snap.Engine == "pg_dump" || snap.Engine == "mysqldump" {
		if err := os.MkdirAll(dest, 0o755); err != nil {
			return err
		}
		f, err := os.Create(path.Join(dest, path.Base(snap.S3Key)))
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(f, rc)
		return err
	}
	_, err = UnpackTarGz(ctx, rc, dest)
	return err
}

func (a *Agent) runPrune(ctx context.Context, st *JobStatus, req JobRequest, cz *Corteza) {
	a.sem <- struct{}{}
	defer func() { <-a.sem }()
	snaps, err := cz.ListRecords(ctx, "snapshots", "")
	if err != nil {
		st.Status = StatusFailed
		st.Error = err.Error()
		a.notifyCallback(st, "failed")
		return
	}
	now := time.Now()
	deleted := 0
	retention := req.RetentionDays
	srcFilter := ParseUint(req.SourceID)
	polFilter := ParseUint(req.PolicyID)
	if polFilter > 0 {
		if p, err := cz.LoadPolicy(ctx, polFilter); err == nil && retention == 0 {
			retention = p.RetentionDays
		}
	}
	for i := range snaps {
		m := recMap(snaps[i])
		if srcFilter > 0 && ParseUint(m["source"]) != srcFilter {
			continue
		}
		if polFilter > 0 && ParseUint(m["policy"]) != polFilter {
			continue
		}
		exp, _ := time.Parse(time.RFC3339, m["expires_at"])
		expired := !exp.IsZero() && exp.Before(now)
		if !expired && retention > 0 {
			created := snaps[i].UpdatedAt
			if t, err := time.Parse(time.RFC3339, created); err == nil && now.Sub(t) > time.Duration(retention)*24*time.Hour {
				expired = true
			}
		}
		if !expired {
			continue
		}
		key := m["s3_key"]
		if key != "" && !strings.HasPrefix(m["engine"], "restic") {
			if strings.HasSuffix(key, "/") {
				keys, _ := a.store.List(ctx, key)
				for _, k := range keys {
					if a.store.Remove(ctx, k) == nil {
						deleted++
					}
				}
			} else if a.store.Remove(ctx, key) == nil {
				deleted++
			}
		}
		if m["engine"] == "restic" {
			src, err := cz.LoadSource(ctx, ParseUint(m["source"]))
			if err == nil {
				handle := sanitizeKey(firstNonEmpty(src.Name, m["source"]))
				_ = resticForget(ctx, a.cfg, handle, retention, a.cfg.ResticPassword)
			}
		}
		_ = cz.UpdateValues(ctx, "snapshots", snaps[i].ID, map[string]string{"restorable": "0"})
	}
	metricPruneDeleted.Add(float64(deleted))
	fin := time.Now()
	st.Status = StatusCompleted
	st.FinishedAt = &fin
	st.Progress = 100
	st.Files = deleted
	st.Message = fmt.Sprintf("pruned %d objects", deleted)
	a.notifyCallback(st, "complete")
}

// recordSnapshot creates the "snapshots" row for a completed backup — the
// actual artifact metadata (where the data landed, its checksum/size),
// which only this agent knows and nobody else can reconstruct from polling.
// This is data the backup produced, not job-status bookkeeping, so it's
// kept here unlike the "jobs"/"restores" status writes this agent used to
// also do (see StartScheduler's doc comment for why those moved to the
// server-side poll/ingest chains instead).
func (a *Agent) recordSnapshot(ctx context.Context, cz *Corteza, st *JobStatus, pol *Policy, res *BackupResult) {
	if cz == nil || res == nil || st.Status != StatusCompleted {
		return
	}
	exp := ""
	if pol != nil && pol.RetentionDays > 0 {
		exp = time.Now().Add(time.Duration(pol.RetentionDays) * 24 * time.Hour).UTC().Format(time.RFC3339)
	}
	sid, err := cz.CreateValues(ctx, "snapshots", map[string]string{
		"job":         st.JobRecordID,
		"source":      st.SourceID,
		"policy":      st.PolicyID,
		"s3_bucket":   res.Bucket,
		"s3_key":      res.Key,
		"size_bytes":  fmtInt(res.BytesWritten),
		"checksum":    res.Checksum,
		"files_count": strconv.Itoa(res.Files),
		"kind":        string(res.Kind),
		"engine":      res.Engine,
		"restic_id":   res.ResticID,
		"expires_at":  exp,
		"restorable":  "1",
		"verified":    "0",
	})
	if err != nil {
		log.Printf("record snapshot for job %s: %v", st.ID, err)
		return
	}
	st.SnapshotID = fmtUint(sid)
}
