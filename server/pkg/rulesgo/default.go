package rulesgo

import "context"

type DefaultConfig struct {
	CRUD       CRUDService
	Mail       MailService
	AICall     func(ctx context.Context, agent, prompt, model string) (string, error)
	ScriptExec func(ctx context.Context, code string, ec *ExecutionContext) (map[string]interface{}, error)
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

	r.Register("gonec", &gonecExecutor{})

	if cfg != nil && cfg.ScriptExec != nil {
		r.Register("script", &scriptExecutor{exec: cfg.ScriptExec})
	} else {
		r.Register("script", &scriptExecutor{})
	}

	return r
}
