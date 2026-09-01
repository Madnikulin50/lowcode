package rulesgo

import "context"

type DefaultConfig struct {
	CRUD        CRUDService
	Mail        MailService
	AICall      func(ctx context.Context, agent, prompt, model string) (string, error)
	ScriptExec  func(ctx context.Context, code string, ec *ExecutionContext) (map[string]interface{}, error)
	DetachStart DetachStartFunc
	ExtractExec NodeExecutor
}

func DefaultRegistry(cfg *DefaultConfig) *Registry {
	r := NewRegistry()

	crud := &crudExecutor{}
	mail := &mailExecutor{}
	if cfg != nil {
		if cfg.CRUD != nil {
			crud.svc = cfg.CRUD
		}
		if cfg.Mail != nil {
			mail.svc = cfg.Mail
		}
	}

	r.Register("crud", crud)
	r.Register("crud.upsert", &upsertExecutor{svc: crud.svc})
	r.Register("foreach", &foreachExecutor{})
	r.Register("mail", mail)
	r.Register("http", &httpExecutor{})
	r.Register("condition", &conditionExecutor{})

	if cfg != nil && cfg.AICall != nil {
		r.Register("ai", &aiExecutor{call: cfg.AICall})
	} else {
		r.Register("ai", &aiExecutor{})
	}

	r.Register("workflow", &wfExecutor{})
	r.Register("fork", &forkExecutor{})
	var detachStart DetachStartFunc
	if cfg != nil {
		detachStart = cfg.DetachStart
	}
	r.Register("detach", &detachExecutor{start: detachStart})

	r.Register("gonec", &gonecExecutor{})

	if cfg != nil && cfg.ScriptExec != nil {
		r.Register("script", &scriptExecutor{exec: cfg.ScriptExec})
	} else {
		r.Register("script", &scriptExecutor{})
	}

	r.Register("score.matrix", &scoreMatrixExecutor{})
	r.Register("score.weighted", &scoreWeightedExecutor{})
	r.Register("risk.band", &riskBandExecutor{})

	if cfg != nil && cfg.ExtractExec != nil {
		r.Register("document.extract", cfg.ExtractExec)
	} else {
		r.Register("document.extract", extractStub{})
	}

	r.Register("service.call", &componentExecutor{detach: detachStart})
	for _, spec := range RemoteCatalog() {
		spec := spec
		t := spec.Type
		if t == "" {
			t = spec.Service + "/" + spec.Operation
		}
		r.Register(t, &componentExecutor{spec: spec, detach: detachStart})
	}

	return r
}

func RemoteCatalog() []RemoteSpec {
	return []RemoteSpec{
		{Type: "cmdb/scan", Service: "cmdb", Operation: "scan", Async: true, Ingest: "cmdb-ingest-scan"},
		{Type: "backup/run", Service: "backup", Operation: "backup", Async: true, Ingest: "backup-ingest-job"},
		{Type: "backup/restore", Service: "backup", Operation: "restore", Async: true, Ingest: "backup-ingest-job"},
		{Type: "backup/prune", Service: "backup", Operation: "prune", Async: true, Ingest: "backup-ingest-job"},
		{Type: "backup/due", Service: "backup", Operation: "due", Async: false},
	}
}
