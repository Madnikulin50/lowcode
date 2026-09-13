package rulesgo

import "context"

type DefaultConfig struct {
	CRUD CRUDService
	Mail MailService
	// AICall backs both the ai and ai.operation nodes. allowMutating stands
	// in for the interactive "да" a chat user would type before a mutating
	// tool call executes - the implementation is expected to route it
	// through to the agent runtime's confirmation gate (see
	// aiagent.Agent.RunConfirmed) rather than always allowing it.
	AICall      func(ctx context.Context, agent, prompt, model string, allowMutating bool) (*AIOperationResult, error)
	ScriptExec  func(ctx context.Context, code string, ec *ExecutionContext) (map[string]interface{}, error)
	DetachStart DetachStartFunc
	ExtractExec NodeExecutor

	// KafkaSubscribeStart/RabbitMQSubscribeStart back the kafka.subscribe and
	// rabbitmq.subscribe nodes - normally BrokerSubscriber.StartKafkaSubscribe
	// / StartRabbitMQSubscribe. Left nil, those node types report
	// "not_configured" instead of failing.
	KafkaSubscribeStart    func(ctx context.Context, subKey string, cfg KafkaConfig, ingestChainID string) error
	RabbitMQSubscribeStart func(ctx context.Context, subKey string, cfg RabbitMQConfig, ingestChainID string) error

	// ResolveCorrelation backs the automation.correlate node - normally
	// automation/service.ResolveCorrelation. Left nil, the node reports
	// "not_configured" instead of failing.
	ResolveCorrelation func(ctx context.Context, key string, input map[string]interface{}) error
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

	var aiCall func(ctx context.Context, agent, prompt, model string, allowMutating bool) (*AIOperationResult, error)
	if cfg != nil {
		aiCall = cfg.AICall
	}
	r.Register("ai", &aiExecutor{call: aiCall})
	r.Register("ai.operation", &aiOperationExecutor{call: aiCall})

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

	r.Register("kafka.produce", &kafkaProduceExecutor{})
	r.Register("kafka.consume", &kafkaConsumeExecutor{})
	r.Register("rabbitmq.publish", &rabbitmqPublishExecutor{})
	r.Register("rabbitmq.consume", &rabbitmqConsumeExecutor{})
	r.Register("1c.sync", &oneCSyncExecutor{})
	r.Register("format.convert", &formatConvertExecutor{})

	var resolveCorrelation func(ctx context.Context, key string, input map[string]interface{}) error
	if cfg != nil {
		resolveCorrelation = cfg.ResolveCorrelation
	}
	r.Register("automation.correlate", &automationCorrelateExecutor{resolve: resolveCorrelation})

	var kafkaSubStart func(ctx context.Context, subKey string, cfg KafkaConfig, ingestChainID string) error
	var rabbitmqSubStart func(ctx context.Context, subKey string, cfg RabbitMQConfig, ingestChainID string) error
	if cfg != nil {
		kafkaSubStart = cfg.KafkaSubscribeStart
		rabbitmqSubStart = cfg.RabbitMQSubscribeStart
	}
	r.Register("kafka.subscribe", &kafkaSubscribeExecutor{start: kafkaSubStart})
	r.Register("rabbitmq.subscribe", &rabbitmqSubscribeExecutor{start: rabbitmqSubStart})

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
		{Type: "backup/restore", Service: "backup", Operation: "restore", Async: true, Ingest: "backup-ingest-restore"},
		{Type: "backup/prune", Service: "backup", Operation: "prune", Async: true, Ingest: "backup-ingest-job"},
		{Type: "backup/due", Service: "backup", Operation: "due", Async: false},
	}
}
