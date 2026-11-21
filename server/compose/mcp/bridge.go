package mcp

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/madnikulin50/lowcode/server/automation/automation"
	automationService "github.com/madnikulin50/lowcode/server/automation/service"
	"github.com/madnikulin50/lowcode/server/compose/mcp/handlers"
	"github.com/madnikulin50/lowcode/server/compose/service"
	"github.com/madnikulin50/lowcode/server/compose/types"
	"github.com/madnikulin50/lowcode/server/pkg/aiagent"
	"github.com/madnikulin50/lowcode/server/pkg/auth"
	"github.com/madnikulin50/lowcode/server/pkg/chat"
	"github.com/madnikulin50/lowcode/server/pkg/expr"
	"github.com/madnikulin50/lowcode/server/pkg/gonec"
	"github.com/madnikulin50/lowcode/server/pkg/jsruntime"
	"github.com/madnikulin50/lowcode/server/pkg/riskstore"
	"github.com/madnikulin50/lowcode/server/pkg/rulesgo"
)

func initBridge() {
	gonecEngine := gonec.New(os.TempDir())
	handlers.SetGonecEngine(gonecEngine)

	jsRt := jsruntime.New(jsruntime.Services{
		RecordCreate: func(ctx context.Context, nsID, modID uint64, values map[string]interface{}) (string, string, error) {
			r := &types.Record{NamespaceID: nsID, ModuleID: modID, Values: make([]*types.RecordValue, 0)}
			for name, val := range values {
				r.Values = append(r.Values, &types.RecordValue{Name: name, Value: fmt.Sprintf("%v", val)})
			}
			created, _, err := service.DefaultRecord.Create(ctx, r)
			if err != nil {
				return "", "", err
			}
			return fmt.Sprintf("%d", created.ID), created.CreatedAt.String(), nil
		},
		RecordUpdate: func(ctx context.Context, nsID, modID uint64, recordID string, values map[string]interface{}) (string, error) {
			var rid uint64
			fmt.Sscanf(recordID, "%d", &rid)
			existing, _, err := service.DefaultRecord.FindByID(ctx, nsID, modID, rid)
			if err != nil {
				return "", err
			}
			for name, val := range values {
				found := false
				for _, rv := range existing.Values {
					if rv.Name == name {
						rv.Value = fmt.Sprintf("%v", val)
						found = true
						break
					}
				}
				if !found {
					existing.Values = append(existing.Values, &types.RecordValue{Name: name, Value: fmt.Sprintf("%v", val)})
				}
			}
			updated, _, err := service.DefaultRecord.Update(ctx, existing)
			if err != nil {
				return "", err
			}
			return updated.UpdatedAt.String(), nil
		},
		RecordDelete: func(ctx context.Context, nsID, modID uint64, recordID string) error {
			var rid uint64
			fmt.Sscanf(recordID, "%d", &rid)
			return service.DefaultRecord.DeleteByID(ctx, nsID, modID, rid)
		},
		RecordSearch: func(ctx context.Context, nsID, modID uint64, query string, limit int) ([]map[string]interface{}, error) {
			ff := types.RecordFilter{NamespaceID: nsID, ModuleID: modID, Query: query}
			if limit > 0 {
				ff.Limit = uint(limit)
			}
			set, _, err := service.DefaultRecord.Find(ctx, ff)
			if err != nil {
				return nil, err
			}
			result := make([]map[string]interface{}, 0, len(set))
			for _, r := range set {
				record := map[string]interface{}{"recordID": fmt.Sprintf("%d", r.ID)}
				for _, v := range r.Values {
					record[v.Name] = trimQuotes(v.Value)
				}
				result = append(result, record)
			}
			return result, nil
		},
		MailSend: func(ctx context.Context, to []string, subject, body string, cc []string, contentType string) error {
			n := &types.EmailNotification{To: to, Cc: cc, Subject: subject}
			if contentType == "html" {
				n.ContentHTML = body
			} else {
				n.ContentPlain = body
			}
			return service.DefaultNotification.SendEmail(ctx, n)
		},
	})
	handlers.SetJSRuntime(jsRt)

	service.RegisterComposeToolKits(aiagent.DefaultCatalog())
	aiagent.DefaultCatalog().StartRemoteDiscovery()

	// Agent registry
	chatClient, err := chat.NewClient(chat.ModelForRole(chat.RoleMCPAgent))
	if err != nil {
		log.Printf("[bridge] agent client: %v", err)
	} else {
		reg := aiagent.NewRegistry(chatClient)
		reg.RegisterDefault(nil)
		handlers.SetAgentRegistry(reg)
		handlers.SetAgentClient(chatClient)
	}

	// Rulesgo engine: chains persist in compose_rule_chain (PostgreSQL)
	persist := service.NewRuleChainPersistence()
	poller := rulesgo.NewAgentPoller()
	brokerSub := rulesgo.NewBrokerSubscriber()
	rulesCfg := &rulesgo.DefaultConfig{
		CRUD:                   composeCRUD{},
		DetachStart:            poller.StartFromDetach,
		ExtractExec:            service.NewDocumentExtractExecutor(),
		KafkaSubscribeStart:    brokerSub.StartKafkaSubscribe,
		RabbitMQSubscribeStart: brokerSub.StartRabbitMQSubscribe,
		AICall:                 ruleChainAICall,
		ResolveCorrelation: func(ctx context.Context, key string, input map[string]interface{}) error {
			vars, err := expr.NewVars(input)
			if err != nil {
				return fmt.Errorf("build resume input: %w", err)
			}
			return automationService.ResolveCorrelation(ctx, key, vars)
		},
		WorkflowExec: automationService.ExecWorkflowByRef,
		ScriptExec: func(ctx context.Context, code string, ec *rulesgo.ExecutionContext) (map[string]interface{}, error) {
			input := make(map[string]interface{})
			for k, v := range ec.Variables {
				input[k] = v
			}
			for k, v := range ec.Input {
				input[k] = v
			}
			result := jsRt.Run(ctx, code, input)
			return map[string]interface{}{
				"output": result.Output,
				"logs":   result.Logs,
				"error":  result.Error,
			}, nil
		},
	}
	// Knowledge-base search for the workflow aiRagSearch step
	automation.SetRAGSearch(func(ctx context.Context, ns, query string, topK int) ([]automation.RAGHit, error) {
		if service.DefaultRAG == nil {
			return nil, fmt.Errorf("rag service not available")
		}
		res, err := service.DefaultRAG.Search(ctx, ns, query, topK)
		if err != nil {
			return nil, err
		}
		hits := make([]automation.RAGHit, 0, len(res))
		for _, r := range res {
			hits = append(hits, automation.RAGHit{Text: r.Chunk.Text, Score: r.Score})
		}
		return hits, nil
	})

	engine := rulesgo.NewEngineWithPersistence(rulesgo.DefaultRegistry(rulesCfg), persist)
	poller.SetEngine(engine)
	brokerSub.SetEngine(engine)
	service.SetRuleEngine(engine)
	rulesgo.SetDefaultPoller(poller)
	rulesgo.CapturePollIdentity = func(ctx context.Context) (uint64, []uint64) {
		ident := auth.GetIdentityFromContext(ctx)
		if ident == nil || !ident.Valid() {
			return 0, nil
		}
		return ident.Identity(), ident.Roles()
	}
	rulesgo.RestorePollIdentity = func(ctx context.Context, userID uint64, roles []uint64) context.Context {
		if userID == 0 {
			return ctx
		}
		return auth.SetIdentityToContext(ctx, auth.Authenticated(userID, roles...))
	}
	if err := engine.LoadFromStore(context.Background()); err != nil {
		log.Printf("[bridge] load rule chains from DB: %v", err)
	} else {
		log.Printf("[bridge] loaded %d rule chains from PostgreSQL", len(engine.Chains()))
	}
	handlers.SetRuleEngine(engine)
	service.StartRuleChainRecordTriggers(engine)
	riskstore.EnsurePersistence()
	service.StartRiskRecordTriggers()
	handlers.SetOnChainMissing(func(ctx context.Context, chainID string) {
		ensureChainAvailable(ctx, engine, chainID)
	})

	registerDemoChains(engine)
	registerCMDBChains(engine)
	registerBackupChains(engine)
	rulesgo.EnsureChain = func(ctx context.Context, chainID string) {
		ensureChainAvailable(ctx, engine, chainID)
	}

	log.Println("[bridge] all services wired")
}

func recordSetToMap(set types.RecordSet) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(set))
	for _, r := range set {
		record := map[string]interface{}{"recordID": fmt.Sprintf("%d", r.ID)}
		for _, v := range r.Values {
			record[v.Name] = trimQuotes(v.Value)
		}
		result = append(result, record)
	}
	return result
}

func trimQuotes(s string) string {
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}
