package rulesgo

import (
	"context"
	"fmt"
)

type foreachConfig struct {
	Items    string `json:"items"`
	ItemVar  string `json:"itemVar,omitempty"`
	MaxItems int    `json:"maxItems,omitempty"`
	FailFast bool   `json:"failFast,omitempty"`
}

type foreachExecutor struct{}

func (n *foreachExecutor) Execute(ctx context.Context, node ChainNode, ec *ExecutionContext) (map[string]interface{}, error) {
	cfg, err := ParseNodeConfig[foreachConfig](node.Config)
	if err != nil {
		return nil, err
	}
	itemsKey := cfg.Items
	if itemsKey == "" {
		itemsKey = "items"
	}
	items := CollectItems(ec.Get(itemsKey))
	itemVar := cfg.ItemVar
	if itemVar == "" {
		itemVar = "item"
	}
	return map[string]interface{}{
		"itemsKey": itemsKey,
		"itemVar":  itemVar,
		"count":    len(items),
		"maxItems": cfg.MaxItems,
		"failFast": cfg.FailFast,
	}, nil
}

func (e *Engine) runForeach(ctx context.Context, node *ChainNode, nodeMap map[string]*ChainNode, bodyIDs []string, ec *ExecutionContext) (NodeResult, error) {
	cfg, err := ParseNodeConfig[foreachConfig](node.Config)
	if err != nil {
		return NodeResult{NodeID: node.ID, Type: node.Type}, err
	}
	if cfg.Items == "" {
		cfg.Items = "items"
	}
	if cfg.ItemVar == "" {
		cfg.ItemVar = "item"
	}
	items := CollectItems(ec.Get(cfg.Items))
	ok, fail := 0, 0
	nr := NodeResult{NodeID: node.ID, Type: node.Type, Next: bodyIDs}

	for i, item := range items {
		if cfg.MaxItems > 0 && i >= cfg.MaxItems {
			break
		}
		FlattenItem(ec, cfg.ItemVar, item)
		itemFail := false
		for _, bid := range bodyIDs {
			body := nodeMap[bid]
			if body == nil {
				return nr, fmt.Errorf("foreach body node not found: %s", bid)
			}
			out, err := e.registry.Execute(ctx, body.Type, *body, ec)
			bodyRes := NodeResult{NodeID: fmt.Sprintf("%s#%d", bid, i), Type: body.Type, Output: out}
			if err != nil {
				bodyRes.Error = err.Error()
				nr.Output = map[string]interface{}{"ok": ok, "failed": fail + 1, "count": len(items)}
				if cfg.FailFast {
					return nr, fmt.Errorf("foreach item %d node %s: %w", i, bid, err)
				}
				itemFail = true
				continue
			}
			if out != nil {
				ec.SetResult(body.ID, out)
			}
			_ = bodyRes
		}
		if itemFail {
			fail++
		} else {
			ok++
		}
	}
	nr.Output = map[string]interface{}{"ok": ok, "failed": fail, "count": len(items)}
	ec.Set("foreachCount", len(items))
	ec.Set("foreachOk", ok)
	ec.Set("foreachFailed", fail)
	return nr, nil
}
