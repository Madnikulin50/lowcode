// automation.correlate bridges a rule chain back into a suspended
// automation/BPMN process: it's the mirror image of the compose.runRuleChain
// Service Task bridge (compose/service/rulechain_bridge.go) - instead of a
// BPMN Service Task calling into a rule chain, this node lets a rule chain
// (typically an ingest chain fed by kafka.subscribe/rabbitmq.subscribe)
// resume whatever BPMN process is waiting on a correlation key, when the
// message that arrives is the external event a "message intermediate catch
// event" was compiled to wait for (see
// automation/service/bpmn_compile.go's compileIntermediateCatchEvent and
// automation/service/correlation.go's ResolveCorrelation).
package rulesgo

import (
	"context"
	"fmt"
)

type AutomationCorrelateConfig struct {
	// Key: the correlation key to match against a pending prompt's
	// correlationKey argument - typically a template referencing the
	// ingest envelope, e.g. "{{value}}" or "{{key}}".
	Key string `json:"key"`

	// Input: named value -> template, forwarded as the resumed process's
	// input (e.g. the parsed message body). All ingest envelope fields are
	// forwarded by default if Input is empty.
	Input map[string]string `json:"input,omitempty"`
}

type automationCorrelateExecutor struct {
	resolve func(ctx context.Context, key string, input map[string]interface{}) error
}

func (n *automationCorrelateExecutor) Execute(ctx context.Context, node ChainNode, ec *ExecutionContext) (map[string]interface{}, error) {
	cfg, err := ParseNodeConfig[AutomationCorrelateConfig](node.Config)
	if err != nil {
		return nil, err
	}
	key := resolveTemplateValue(cfg.Key, ec)
	if key == "" {
		return nil, fmt.Errorf("automation.correlate: key is required")
	}

	var input map[string]interface{}
	if len(cfg.Input) > 0 {
		input = make(map[string]interface{}, len(cfg.Input))
		for k, v := range cfg.Input {
			input[k] = resolveTemplateValue(v, ec)
		}
	} else if ec != nil {
		input = ec.Input
	}
	if input == nil {
		input = map[string]interface{}{}
	}

	if n.resolve == nil {
		return map[string]interface{}{"status": "not_configured", "key": key}, nil
	}

	if err := n.resolve(ctx, key, input); err != nil {
		return nil, fmt.Errorf("automation.correlate: %w", err)
	}

	return map[string]interface{}{"success": true, "key": key}, nil
}
