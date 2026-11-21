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
	cfg     Config
	cz      *Corteza
	store   Storage
	vault   *VaultStore
	mu      sync.RWMutex
	jobs    map[string]*JobStatus
	sem     chan struct{}
	cb      *sdk.Callback
	stores  map[string]Storage
	storesM sync.Mutex
}

func New(cfg Config, store Storage, cz *Corteza) *Agent {
	return NewWithVault(cfg, store, nil, cz)
}

func NewWithVault(cfg Config, store Storage, vault *VaultStore, cz *Corteza) *Agent {
	n := cfg.Concurrency
	if n <= 0 {
		n = 2
	}
	return &Agent{
		cfg:    cfg,
		cz:     cz,
		store:  store,
		vault:  vault,
		jobs:   map[string]*JobStatus{},
		sem:    make(chan struct{}, n),
		cb:     sdk.NewCallback(),
		stores: map[string]Storage{},
	}
}

// keyFor собирает ключ объекта с учётом префикса: dest_prefix политики
// побеждает, иначе — префикс назначения по умолчанию.
func keyFor(pol *Policy, dest *Destination, handle, jobTag, filename string) string {
	key := ObjectKey(handle, jobTag, filename)
	prefix := ""
	if pol != nil {
		prefix = strings.TrimSpace(pol.DestPrefix)
	}
	if prefix == "" && dest != nil {
		prefix = strings.TrimSpace(dest.Prefix)
	}
	if prefix != "" {
		key = path.Join(strings.Trim(prefix, "/"), key)
	}
	return key
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
	if kopiaAvailable() {
		caps += ",kopia"
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

// enforceDataClass — правила раздельного хранения (152-ФЗ, интервью п.3):
// ПДн спецкатегорий: для fs — только шифрованный kopia; для остальных типов
// (БД, SMB, SaaS, k8s) — обязательна WORM-копия в сейфе. Конфиденциальные —
// сейф обязателен для всех типов.
func enforceDataClass(pol *Policy, src *Source) error {
	switch src.DataClass {
	case "personal_special":
		if src.Type == SourceFS {
			if strings.ToLower(strings.TrimSpace(pol.Engine)) != "kopia" {
				return fmt.Errorf("data_class=personal_special (fs) requires engine=kopia (client-side encryption); set it on policy %q", pol.Name)
			}
		} else if !pol.Vault {
			return fmt.Errorf("data_class=personal_special requires vault=true (WORM copy); set it on policy %q", pol.Name)
		}
	case "confidential":
		if !pol.Vault {
			return fmt.Errorf("data_class=confidential requires vault=true (WORM copy); set it on policy %q", pol.Name)
		}
	}
	if pol.Archive && !pol.VaultLegalHold {
		log.Printf("policy %q: archive mode without vault legal hold — long-term legal validity is at risk", pol.Name)
	}
	return nil
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
	if err := enforceDataClass(pol, src); err != nil {
		return nil, err
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

	go a.runBackup(context.Background(), st, src, pol, cred, cz, req)
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
	go a.runRestore(context.Background(), st, src, snap, cred, req, cz)
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

func (a *Agent) runBackup(ctx context.Context, st *JobStatus, src *Source, pol *Policy, cred *Credential, cz *Corteza, jr JobRequest) {
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

	var dest *Destination
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
			// SaaS: фиксируем курсор инкремента после успешной выгрузки.
			if src.Type == SourceSaaS && cz.HasToken() {
				if err := cz.UpdateSourceLastSync(ctx, src.ID, now); err != nil {
					log.Printf("last_sync source %d: %v", src.ID, err)
				}
			}
		}
		metricDuration.WithLabelValues(label).Observe(time.Since(start).Seconds())
		a.recordSnapshot(ctx, cz, st, src, pol, dest, res)
		a.notifyCallback(st, kind)
	}

	dest, err := a.resolveDestination(ctx, cz, jr, pol)
	if err != nil {
		finish(StatusFailed, err.Error(), nil)
		return
	}
	store, err := a.StoreForDestination(ctx, dest)
	if err != nil {
		finish(StatusFailed, err.Error(), nil)
		return
	}
	st.S3Endpoint = store.Location()

	secret := ResolveSecret(cred.Handle)
	username := cred.Username
	extra := ""
	if cred != nil {
		extra = cred.Extra
	}
	res, err := a.backupSource(ctx, st, store, src, pol, dest, cred, username, secret, extra, handle)
	if err != nil {
		finish(StatusFailed, err.Error(), res)
		return
	}
	finish(StatusCompleted, "", res)
}

func (a *Agent) backupSource(ctx context.Context, st *JobStatus, store Storage, src *Source, pol *Policy, dest *Destination, cred *Credential, username, secret, extra, handle string) (*BackupResult, error) {
	// Движки restic/kopia строят репозиторий по назначению (engineRepoFor /
	// kopiaConnectArgs) — тип назначения решает, какая схема repo будет.
	engine := strings.ToLower(strings.TrimSpace(pol.Engine))
	if engine == "" {
		// auto: прежнее поведение — restic-инкремент для fs при флаге incremental.
		engine = "restic"
	}
	excludes := parseExclude(pol.Exclude)
	switch {
	case engine == "kopia":
		if !kopiaAvailable() {
			return nil, fmt.Errorf("kopia engine requested but kopia binary is not installed")
		}
		if src.Type != SourceFS {
			return nil, fmt.Errorf("kopia engine supports fs sources only (got %s)", src.Type)
		}
		st.Message = "kopia snapshot create"
		id, err := kopiaSnapshotCreate(ctx, a.cfg, dest, handle, src.Path, excludes, a.cfg.ResticPassword)
		if err != nil {
			return nil, err
		}
		return &BackupResult{
			Bucket:   store.Bucket(),
			Key:      kopiaKeyHint(handle, id),
			Engine:   "kopia",
			Kind:     KindIncremental,
			ResticID: id,
		}, nil
	case engine == "restic" && pol.Incremental && resticAvailable() && src.Type == SourceFS:
		st.Message = "restic backup"
		spec, err := a.engineRepoFor(ctx, dest, handle)
		if err != nil {
			return nil, err
		}
		defer spec.close()
		id, err := resticBackupPath(ctx, spec, src.Path, excludes, a.cfg.ResticPassword)
		if err != nil {
			return nil, err
		}
		res := &BackupResult{
			Bucket:   store.Bucket(),
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
		return a.packToMinio(ctx, st, store, newExcludeWalker(walker, pol.Exclude), keyFor(pol, dest, handle, jobTag, "archive.tar.gz"))
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
		return a.packToMinio(ctx, st, store, newExcludeWalker(smb, pol.Exclude), keyFor(pol, dest, handle, jobTag, "archive.tar.gz"))
	case SourceDatabase:
		return a.dumpToMinio(ctx, st, store, src, username, secret, keyFor(pol, dest, handle, jobTag, "dump."+dumpExt(src)))
	case SourceS3:
		prefix := keyFor(pol, dest, handle, jobTag, "s3")
		on := func(_ string, n int64) {
			st.BytesWritten += n
			st.Files++
			st.Progress = 50
			a.notifyCallback(st, "progress")
		}
		res, err := CopyS3(ctx, *src, username, secret, store, prefix, on)
		return &res, err
	case SourceSaaS:
		return a.saasToMinio(ctx, st, store, src, cred, keyFor(pol, dest, handle, jobTag, "saas.tar.gz"))
	case SourceK8s:
		return a.k8sToMinio(ctx, st, store, src, cred, keyFor(pol, dest, handle, jobTag, "k8s.tar.gz"))
	default:
		return nil, fmt.Errorf("unknown source type %q", src.Type)
	}
}

func dumpExt(src *Source) string {
	if strings.HasPrefix(strings.ToLower(src.DBEngine), "mysql") {
		return "sql"
	}
	return "pgdump"
}

func (a *Agent) packToMinio(ctx context.Context, st *JobStatus, store Storage, walker FileWalker, key string) (*BackupResult, error) {
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
	written, err := store.Put(ctx, key, pr, -1, "application/gzip")
	if err != nil {
		return nil, err
	}
	if packErr != nil {
		return nil, packErr
	}
	return &BackupResult{
		Bucket:       store.Bucket(),
		Key:          key,
		BytesRead:    bytesRead,
		BytesWritten: written,
		Files:        files,
		Checksum:     ch.SumHex(),
		Engine:       "archive",
		Kind:         KindFull,
	}, nil
}

func (a *Agent) dumpToMinio(ctx context.Context, st *JobStatus, store Storage, src *Source, username, secret, key string) (*BackupResult, error) {
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
	written, err := store.Put(ctx, key, pr, -1, "application/octet-stream")
	if err != nil {
		return nil, err
	}
	if dumpErr != nil {
		return nil, dumpErr
	}
	return &BackupResult{
		Bucket:       store.Bucket(),
		Key:          key,
		BytesRead:    ch.Bytes(),
		BytesWritten: written,
		Files:        1,
		Checksum:     ch.SumHex(),
		Engine:       engine,
		Kind:         KindFull,
	}, nil
}

func (a *Agent) runRestore(ctx context.Context, st *JobStatus, src *Source, snap *Snapshot, cred *Credential, req JobRequest, cz *Corteza) {
	a.sem <- struct{}{}
	defer func() { <-a.sem }()
	secret := ""
	username := ""
	if cred != nil {
		secret = ResolveSecret(cred.Handle)
		username = cred.Username
	}
	store, err := a.storeForSnapshot(ctx, cz, snap)
	if err != nil {
		st.Status = StatusFailed
		st.Error = err.Error()
		a.notifyCallback(st, "failed")
		return
	}
	destRec, derr := a.destinationForSnapshot(ctx, cz, snap)
	if derr != nil {
		st.Status = StatusFailed
		st.Error = derr.Error()
		a.notifyCallback(st, "failed")
		return
	}
	destType := firstNonEmpty(req.DestType, "path")
	dest := req.DestPath
	switch destType {
	case "download":
		st.Message = snap.S3Key
	case "cluster":
		st.Message = "kubectl apply to cluster"
	case "vault":
		if dest == "" {
			dest = path.Join(os.TempDir(), "backup-restore-vault", fmtUint(snap.ID)+"-"+st.ID[:8])
		}
		st.Message = "restore from vault → " + dest
	case "original":
		st.Message = "restore to original location"
	default:
		if dest == "" {
			// Каталог уникален на прогон: повторное восстановление не натыкается
			// на файлы/права прошлого прогона.
			dest = path.Join(os.TempDir(), "backup-restore", fmtUint(snap.ID)+"-"+st.ID[:8])
		}
		st.Message = "restore to " + dest
	}
	switch destType {
	case "download":
		// только ключ, ничего не пишем
	case "cluster":
		// K8s: применить манифесты снапшота к кластеру (kubeconfig источника).
		err = a.restoreToCluster(ctx, src, snap, cred, store)
	case "vault":
		// Восстановление из WORM-сейфа — гарантированно чистая копия.
		err = a.restoreFromVault(ctx, src, snap, dest, username, secret, store, destRec)
	case "original":
		err = a.restoreOriginal(ctx, src, snap, username, secret, store, destRec)
	default:
		err = a.restoreToPath(ctx, src, snap, dest, username, secret, store, destRec)
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

// restoreFromVault восстанавливает из WORM-копии: для архивных движков —
// потоковое чтение объекта из vault-бакета, для restic — restore из
// vault-репозитория. kopia-снапшоты в сейф не копировались — честный отказ.
// store — стор рабочего снапшота (s3copy-кейс читает из него).
func (a *Agent) restoreFromVault(ctx context.Context, src *Source, snap *Snapshot, dest, username, secret string, store Storage, destRec *Destination) error {
	if a.vault == nil || !a.vault.Enabled() {
		return fmt.Errorf("vault is not enabled")
	}
	cfg := a.cfg
	cfg.Minio = store.Config()
	handle := sanitizeKey(firstNonEmpty(src.Name, fmtUint(src.ID)))
	if dest == "" {
		dest = path.Join(os.TempDir(), "backup-restore-vault", fmtUint(snap.ID))
	}
	switch {
	case snap.Engine == "kopia":
		return fmt.Errorf("kopia snapshots are not copied to vault (encrypted repo)")
	case strings.HasPrefix(snap.Engine, "restic"):
		return resticRestoreFromVault(ctx, cfg, handle, snap.ResticID, dest, a.cfg.ResticPassword)
	case snap.Engine == "s3copy":
		if snap.VaultKey == "" {
			return fmt.Errorf("snapshot has no vault copy")
		}
		// Источник — копия в vault-бакете (endpoint агента), а не рабочий снапшот.
		_, err := RestoreS3Copy(ctx, a.store, a.vault.Bucket(), strings.TrimSuffix(snap.VaultKey, "/"), src.Host, src.S3Bucket, username, secret, dest, src.S3Secure, src.S3Region)
		return err
	default:
		if snap.VaultKey == "" {
			return fmt.Errorf("snapshot has no vault copy")
		}
		rc, err := a.vault.Get(ctx, snap.VaultKey)
		if err != nil {
			return err
		}
		defer rc.Close()
		_, err = UnpackTarGz(ctx, rc, dest)
		return err
	}
}

func (a *Agent) restoreOriginal(ctx context.Context, src *Source, snap *Snapshot, username, secret string, store Storage, destRec *Destination) error {
	switch src.Type {
	case SourceFS:
		return a.restoreToPath(ctx, src, snap, src.Path, username, secret, store, destRec)
	case SourceDatabase:
		rc, err := store.GetBucket(ctx, snap.S3Bucket, snap.S3Key)
		if err != nil {
			return err
		}
		defer rc.Close()
		return RestoreDatabase(ctx, *src, username, secret, rc, snap.Engine)
	case SourceS3:
		_, err := RestoreS3Copy(ctx, store, snap.S3Bucket, strings.TrimSuffix(snap.S3Key, "/"), src.Host, src.S3Bucket, username, secret, src.S3Prefix, src.S3Secure, src.S3Region)
		return err
	case SourceSMB:
		tmp := path.Join(os.TempDir(), "backup-restore-smb", fmtUint(snap.ID))
		return a.restoreToPath(ctx, src, snap, tmp, username, secret, store, destRec)
	default:
		return fmt.Errorf("restore original not supported for %s", src.Type)
	}
}

func (a *Agent) restoreToPath(ctx context.Context, src *Source, snap *Snapshot, dest, username, secret string, store Storage, destRec *Destination) error {
	if snap.Engine == "kopia" {
		handle := sanitizeKey(firstNonEmpty(src.Name, fmtUint(src.ID)))
		return kopiaSnapshotRestore(ctx, a.cfg, destRec, handle, snap.ResticID, dest, "", a.cfg.ResticPassword)
	}
	if snap.Engine == "restic" {
		handle := sanitizeKey(firstNonEmpty(src.Name, fmtUint(src.ID)))
		spec, err := a.engineRepoFor(ctx, destRec, handle)
		if err != nil {
			return err
		}
		defer spec.close()
		return resticRestore(ctx, spec, snap.ResticID, dest, a.cfg.ResticPassword)
	}
	if snap.Engine == "s3copy" {
		_, err := RestoreS3Copy(ctx, store, snap.S3Bucket, strings.TrimSuffix(snap.S3Key, "/"), src.Host, src.S3Bucket, username, secret, dest, src.S3Secure, src.S3Region)
		return err
	}
	rc, err := store.GetBucket(ctx, snap.S3Bucket, snap.S3Key)
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
		snap := &Snapshot{ID: snaps[i].ID, DestID: ParseUint(m["destination"]), DestType: m["dest_type"], S3Endpoint: m["s3_endpoint"], S3Bucket: m["s3_bucket"]}
		store, serr := a.storeForSnapshot(ctx, cz, snap)
		if serr != nil {
			log.Printf("prune snapshot %d: %v", snaps[i].ID, serr)
			continue
		}
		key := m["s3_key"]
		if key != "" && !strings.HasPrefix(m["engine"], "restic") && m["engine"] != "kopia" {
			bucket := firstNonEmpty(m["s3_bucket"], store.Bucket())
			if strings.HasSuffix(key, "/") {
				keys, _ := store.ListBucket(ctx, bucket, key)
				for _, k := range keys {
					if store.RemoveBucket(ctx, bucket, k) == nil {
						deleted++
					}
				}
			} else if store.RemoveBucket(ctx, bucket, key) == nil {
				deleted++
			}
		}
		if m["engine"] == "restic" || m["engine"] == "kopia" {
			src, err := cz.LoadSource(ctx, ParseUint(m["source"]))
			if err == nil {
				handle := sanitizeKey(firstNonEmpty(src.Name, m["source"]))
				destRec, derr := a.destinationForSnapshot(ctx, cz, snap)
				if derr != nil {
					log.Printf("prune snapshot %d: %v", snaps[i].ID, derr)
				} else if m["engine"] == "restic" {
					if spec, rerr := a.engineRepoFor(ctx, destRec, handle); rerr == nil {
						_ = resticForget(ctx, spec, retention, a.cfg.ResticPassword)
						spec.close()
					} else {
						log.Printf("prune snapshot %d: %v", snaps[i].ID, rerr)
					}
				} else {
					_ = kopiaExpire(ctx, a.cfg, destRec, handle, retention, a.cfg.ResticPassword)
				}
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
func (a *Agent) recordSnapshot(ctx context.Context, cz *Corteza, st *JobStatus, src *Source, pol *Policy, dest *Destination, res *BackupResult) {
	if cz == nil || res == nil || st.Status != StatusCompleted {
		return
	}
	exp := ""
	if pol != nil && pol.RetentionDays > 0 {
		exp = time.Now().Add(time.Duration(pol.RetentionDays) * 24 * time.Hour).UTC().Format(time.RFC3339)
	}
	vals := map[string]string{
		"job":         st.JobRecordID,
		"source":      st.SourceID,
		"policy":      st.PolicyID,
		"s3_endpoint": st.S3Endpoint,
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
		"data_class":  src.DataClass,
	}
	if dest.IsRecord() {
		vals["destination"] = fmtUint(dest.ID)
	}
	if kind := StoreKind(dest); kind != "" && kind != "minio" {
		vals["dest_type"] = kind
	}
	sid, err := cz.CreateValues(ctx, "snapshots", vals)
	if err != nil {
		log.Printf("record snapshot for job %s: %v", st.ID, err)
		return
	}
	st.SnapshotID = fmtUint(sid)
}
