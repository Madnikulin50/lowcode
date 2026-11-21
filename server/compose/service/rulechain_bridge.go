package service

import (
	"context"
	"fmt"

	atypes "github.com/madnikulin50/lowcode/server/automation/types"
	"github.com/madnikulin50/lowcode/server/pkg/expr"
	"github.com/madnikulin50/lowcode/server/pkg/rulesgo"
)

// DefaultRuleEngine is the rulesgo engine, wired from compose/mcp/bridge.go
// once it's constructed. It's a package var (same pattern as DefaultStore /
// DefaultRecord elsewhere in this package) rather than a direct reference to
// compose/mcp/handlers.RuleEngine: handlers already imports compose/service,
// so the reverse import would be a cycle.
var DefaultRuleEngine *rulesgo.EngineWithPersistence

func SetRuleEngine(e *rulesgo.EngineWithPersistence) {
	DefaultRuleEngine = e
}

// RunRuleChainFunction is an automation workflow function - a BPMN Service
// Task, in effect - that runs an existing rule chain and returns its
// output. This is the bridge that lets an automation/BPMN process reuse
// every integration node already built for rule chains (Kafka, RabbitMQ,
// 1C, format conversion, AI operations, HTTP, CRUD, ...) as a single step,
// instead of reimplementing each one as a separate workflow function.
func RunRuleChainFunction() *atypes.Function {
	return &atypes.Function{
		Ref:    "compose.runRuleChain",
		Kind:   "function",
		Labels: map[string]string{"compose": "step,workflow"},
		Meta: &atypes.FunctionMeta{
			Short:       "Run rule chain",
			Description: "Runs a rule chain (Kafka/RabbitMQ/1C/AI/HTTP/CRUD integration nodes) and returns its output - lets a BPMN Service Task reuse rule chain integration nodes instead of duplicating them as workflow functions.",
		},
		Parameters: []*atypes.Param{
			{
				Name:     "chainID",
				Types:    []string{"String"},
				Required: true,
				Meta:     &atypes.ParamMeta{Label: "Rule chain ID"},
			},
			{
				Name:  "input",
				Types: []string{"Any"},
				Meta:  &atypes.ParamMeta{Label: "Input passed to the chain"},
			},
		},
		Results: []*atypes.Param{
			{Name: "success", Types: []string{"Any"}},
			{Name: "output", Types: []string{"Any"}},
			{Name: "error", Types: []string{"Any"}},
		},
		Handler: runRuleChainHandler,
	}
}

func runRuleChainHandler(ctx context.Context, in *expr.Vars) (*expr.Vars, error) {
	if DefaultRuleEngine == nil {
		return nil, fmt.Errorf("rule chain engine not available")
	}

	chainID := ""
	if in != nil && in.Has("chainID") {
		aux, err := expr.Select(in, "chainID")
		if err != nil {
			return nil, fmt.Errorf("chainID: %w", err)
		}
		chainID = fmt.Sprintf("%v", aux.Get())
	}
	if chainID == "" {
		return nil, fmt.Errorf("chainID is required")
	}

	input := map[string]interface{}{}
	if in != nil && in.Has("input") {
		aux, err := expr.Select(in, "input")
		if err != nil {
			return nil, fmt.Errorf("input: %w", err)
		}
		switch v := aux.Get().(type) {
		case map[string]interface{}:
			input = v
		case *expr.Vars:
			input = v.Dict()
		}
	}

	result, err := DefaultRuleEngine.RunWithLog(ctx, chainID, input, "bpmn-service-task")
	if err != nil {
		return nil, fmt.Errorf("rule chain %s: %w", chainID, err)
	}

	out := &expr.Vars{}
	if err := out.Set("success", result.Success); err != nil {
		return nil, err
	}
	output := result.Output
	if output == nil {
		output = map[string]interface{}{}
	}
	if err := out.Set("output", output); err != nil {
		return nil, err
	}
	if err := out.Set("error", result.Error); err != nil {
		return nil, err
	}

	return out, nil
}
