package service

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/madnikulin50/lowcode/server/compose/types"
	"github.com/madnikulin50/lowcode/server/pkg/errors"
	"github.com/madnikulin50/lowcode/server/pkg/id"
	"github.com/madnikulin50/lowcode/server/pkg/rulesgo"
	"github.com/madnikulin50/lowcode/server/store/adapters/rdbms"
)

type RuleChainDBPersist struct{}

func NewRuleChainPersistence() rulesgo.Persistence {
	return &RuleChainDBPersist{}
}

func (p *RuleChainDBPersist) store() *rdbms.Store {
	return rdbmsStore(DefaultStore)
}

func (p *RuleChainDBPersist) LoadChains(ctx context.Context) ([]*rulesgo.Chain, error) {
	rs := p.store()
	if rs == nil {
		return nil, errors.Internal("rule chain store not ready")
	}
	set, _, err := rdbms.SearchRuleChains(ctx, rs, types.RuleChainFilter{})
	if err != nil {
		return nil, err
	}
	out := make([]*rulesgo.Chain, 0, len(set))
	for _, rc := range set {
		c, err := rc.ToEngine()
		if err != nil {
			log.Printf("[rulechain] skip %s: %v", rc.Handle, err)
			continue
		}
		out = append(out, c)
	}
	return out, nil
}

func (p *RuleChainDBPersist) SaveChain(ctx context.Context, chain *rulesgo.Chain) error {
	return p.upsert(ctx, chain)
}

func (p *RuleChainDBPersist) UpdateChain(ctx context.Context, chain *rulesgo.Chain) error {
	return p.upsert(ctx, chain)
}

func (p *RuleChainDBPersist) DeleteChain(ctx context.Context, chainID string) error {
	rs := p.store()
	if rs == nil {
		return errors.Internal("rule chain store not ready")
	}
	return rdbms.DeleteRuleChainByHandle(ctx, rs, chainID)
}

func (p *RuleChainDBPersist) upsert(ctx context.Context, chain *rulesgo.Chain) error {
	rs := p.store()
	if rs == nil {
		return errors.Internal("rule chain store not ready")
	}
	if chain == nil || strings.TrimSpace(chain.ID) == "" {
		return nil
	}
	row, err := types.RuleChainFromEngine(chain)
	if err != nil {
		return err
	}
	row.Handle = strings.TrimSpace(row.Handle)
	if len(row.Handle) > 128 {
		row.Handle = row.Handle[:128]
	}

	existing, err := rdbms.LookupRuleChainByHandle(ctx, rs, chain.ID)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	if existing != nil {
		row.ID = existing.ID
		row.CreatedAt = existing.CreatedAt
		if row.NamespaceID == 0 {
			row.NamespaceID = existing.NamespaceID
		}
		return rdbms.UpdateRuleChain(ctx, rs, row)
	}
	row.ID = id.Next()
	row.CreatedAt = time.Now().UTC()
	return rdbms.CreateRuleChain(ctx, rs, row)
}
