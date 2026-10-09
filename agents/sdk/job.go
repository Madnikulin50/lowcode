package sdk

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Job is the in-process job record. Domain code talks to this, not HTTP.
type Job struct {
	mu          sync.Mutex
	ID          string
	Operation   string
	Status      Status
	Progress    float64
	Message     string
	Error       string
	Params      map[string]any
	RecordID    string
	NamespaceID uint64
	Token       string
	CallbackURL string
	StartedAt   time.Time
	FinishedAt  *time.Time
	Result      map[string]any
	Items       any
	svc         *Service
}

func (j *Job) Job() *Job { return j }

func (j *Job) Param(key string) string {
	j.mu.Lock()
	defer j.mu.Unlock()
	if j.Params == nil {
		return ""
	}
	return stringify(j.Params[key])
}

func (j *Job) EnvelopeAuto() Envelope {
	j.mu.Lock()
	st := j.Status
	j.mu.Unlock()
	return j.Envelope(KindFromStatus(st))
}

func (j *Job) SetProgress(p float64, msg string) {
	j.mu.Lock()
	j.Progress = p
	if msg != "" {
		j.Message = msg
	}
	j.mu.Unlock()
	if j.svc != nil {
		j.svc.callback.Notify(j, KindProgress)
	}
}

func (j *Job) SetResult(v map[string]any) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.Result = v
}

func (j *Job) SetItems(items any) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.Items = items
}

func (j *Job) Envelope(kind EventKind) Envelope {
	j.mu.Lock()
	defer j.mu.Unlock()
	svc := ""
	if j.svc != nil {
		svc = j.svc.cfg.Handle
	}
	return Envelope{
		Schema:      SchemaV1,
		Service:     svc,
		Operation:   j.Operation,
		ID:          j.ID,
		JobID:       j.ID,
		Kind:        kind,
		Status:      j.Status,
		Progress:    j.Progress,
		Message:     j.Message,
		Error:       j.Error,
		NamespaceID: formatNS(j.NamespaceID),
		RecordID:    j.RecordID,
		StartedAt:   j.StartedAt,
		FinishedAt:  j.FinishedAt,
		Result:      j.Result,
		Items:       j.Items,
	}
}

func formatNS(n uint64) string {
	if n == 0 {
		return ""
	}
	return FlexID(n).String()
}

type Runner struct {
	mu   sync.RWMutex
	jobs map[string]*Job
}

func NewRunner() *Runner {
	return &Runner{jobs: map[string]*Job{}}
}

func (r *Runner) Put(j *Job) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.jobs[j.ID] = j
}

func (r *Runner) Get(id string) *Job {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.jobs[id]
}

func (r *Runner) List() []*Job {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*Job, 0, len(r.jobs))
	for _, j := range r.jobs {
		out = append(out, j)
	}
	return out
}

func newJob(svc *Service, op string, req StartRequest) *Job {
	id := uuid.New().String()
	return &Job{
		ID:          id,
		Operation:   op,
		Status:      StatusRunning,
		Params:      req.Params,
		RecordID:    firstNonEmpty(req.RecordID, req.JobID),
		NamespaceID: req.NamespaceID.Uint64(),
		Token:       req.Token,
		CallbackURL: req.CallbackURL,
		StartedAt:   time.Now(),
		Result:      map[string]any{},
		svc:         svc,
	}
}

func (j *Job) finish(err error) {
	j.mu.Lock()
	now := time.Now()
	j.FinishedAt = &now
	j.Progress = 100
	if err != nil {
		j.Status = StatusFailed
		j.Error = err.Error()
	} else {
		j.Status = StatusCompleted
	}
	kind := KindComplete
	if j.Status == StatusFailed {
		kind = KindFailed
	}
	j.mu.Unlock()
	if j.svc != nil {
		j.svc.callback.Notify(j, kind)
	}
}

func (j *Job) run(ctx context.Context, fn func(context.Context, *Job) error) {
	err := fn(ctx, j)
	j.finish(err)
}
