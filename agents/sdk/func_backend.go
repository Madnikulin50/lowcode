package sdk

import (
	"context"
	"encoding/json"
)

// AsyncBackend runs exec in a goroutine and tracks jobs for GET /api/jobs.
type AsyncBackend struct {
	svc  *Service
	run  *Runner
	exec func(context.Context, *Job) error
	call func(context.Context, string, StartRequest) (any, error)
}

// UseAsync installs the default backend for a long-running skill (scan, backup).
func (s *Service) UseAsync(exec func(context.Context, *Job) error) *AsyncBackend {
	b := &AsyncBackend{svc: s, run: NewRunner(), exec: exec}
	s.SetBackend(b)
	return b
}

func (b *AsyncBackend) WithCall(call func(context.Context, string, StartRequest) (any, error)) *AsyncBackend {
	b.call = call
	return b
}

func (b *AsyncBackend) StartJob(_ context.Context, req StartRequest) (*Envelope, error) {
	op := req.Operation
	if op == "" {
		op = b.svc.defaultOp(req)
	}
	j := newJob(b.svc, op, req)
	b.run.Put(j)
	go j.run(context.Background(), b.exec)
	env := j.Envelope(KindProgress)
	return &env, nil
}

func (b *AsyncBackend) GetJob(id string) *Envelope {
	j := b.run.Get(id)
	if j == nil {
		return nil
	}
	env := j.EnvelopeAuto()
	return &env
}

func (b *AsyncBackend) ListJobs() []*Envelope {
	jobs := b.run.List()
	out := make([]*Envelope, 0, len(jobs))
	for _, j := range jobs {
		env := j.EnvelopeAuto()
		out = append(out, &env)
	}
	return out
}

func (b *AsyncBackend) JobItems(id string) any {
	j := b.run.Get(id)
	if j == nil {
		return nil
	}
	j.mu.Lock()
	defer j.mu.Unlock()
	if j.Items == nil {
		return []any{}
	}
	return j.Items
}

func (b *AsyncBackend) Call(ctx context.Context, operation string, req StartRequest) (any, error) {
	if b.call == nil {
		return nil, ErrNoCall
	}
	return b.call(ctx, operation, req)
}

func (b *AsyncBackend) Health(_ context.Context) map[string]any {
	return map[string]any{"status": "ok", "service": b.svc.cfg.Handle}
}

// CallBackend runs Call on POST /api/call/{op} and, if the op is marked Sync,
// on POST /api/jobs. Used for pure functions (EVM, CPM).
type CallBackend struct {
	svc  *Service
	run  *Runner
	call func(context.Context, string, StartRequest) (any, error)
}

func (s *Service) UseSync(call func(context.Context, string, StartRequest) (any, error)) *CallBackend {
	b := &CallBackend{svc: s, run: NewRunner(), call: call}
	s.SetBackend(b)
	return b
}

func (b *CallBackend) Call(ctx context.Context, operation string, req StartRequest) (any, error) {
	return b.call(ctx, operation, req)
}

func (b *CallBackend) StartJob(ctx context.Context, req StartRequest) (*Envelope, error) {
	op := req.Operation
	if op == "" {
		op = b.svc.defaultOp(req)
	}
	out, err := b.call(ctx, op, req)
	j := newJob(b.svc, op, req)
	if err != nil {
		j.finish(err)
		b.run.Put(j)
		return nil, err
	}
	attachOutput(j, out)
	j.finish(nil)
	b.run.Put(j)
	env := j.Envelope(KindComplete)
	return &env, nil
}

func (b *CallBackend) GetJob(id string) *Envelope {
	j := b.run.Get(id)
	if j == nil {
		return nil
	}
	env := j.EnvelopeAuto()
	return &env
}

func (b *CallBackend) ListJobs() []*Envelope {
	jobs := b.run.List()
	out := make([]*Envelope, 0, len(jobs))
	for _, j := range jobs {
		env := j.EnvelopeAuto()
		out = append(out, &env)
	}
	return out
}

func (b *CallBackend) Health(_ context.Context) map[string]any {
	return map[string]any{"status": "ok", "service": b.svc.cfg.Handle}
}

func attachOutput(j *Job, out any) {
	if out == nil {
		return
	}
	raw, err := json.Marshal(out)
	if err != nil {
		return
	}
	var m map[string]any
	if json.Unmarshal(raw, &m) != nil {
		return
	}
	if items, ok := m["items"]; ok {
		j.SetItems(items)
		delete(m, "items")
	}
	j.SetResult(m)
}
