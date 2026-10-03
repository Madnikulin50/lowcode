package service

import (
	"context"
	"time"

	"github.com/getsentry/sentry-go"
	"go.uber.org/zap"

	"github.com/madnikulin50/lowcode/server/anomaly/types"
	composeService "github.com/madnikulin50/lowcode/server/compose/service"
	composeTypes "github.com/madnikulin50/lowcode/server/compose/types"
	"github.com/madnikulin50/lowcode/server/pkg/auth"
	"github.com/madnikulin50/lowcode/server/pkg/errors"
	"github.com/madnikulin50/lowcode/server/pkg/eventbus"
	"github.com/madnikulin50/lowcode/server/pkg/id"
	"github.com/madnikulin50/lowcode/server/store"
)

type (
	// scanner is the periodic, platform-internal batch job that walks every
	// Compose record covered by an enabled Rule, scores it with the Rule's
	// configured Detector (see detector.go's registry) and records the
	// result. It runs on a time.Ticker (see Watch) rather than the
	// scheduler+eventbus pipeline: that pipeline only dispatches to
	// user-authored automation workflows (see automation/service/trigger.go),
	// there is no built-in-Go-code hook into it.
	scanner struct {
		log    *zap.Logger
		store  store.Storer
		record composeService.RecordService
	}
)

func Scanner(log *zap.Logger, s store.Storer, record composeService.RecordService) *scanner {
	return &scanner{
		log:    log.Named("anomaly-scanner"),
		store:  s,
		record: record,
	}
}

// Watch starts the ticker-driven scan loop, mirroring
// system/service/reminder.go's Watch(ctx) - the established pattern in this
// codebase for internal periodic work that isn't a user-configurable
// automation trigger.
func (svc *scanner) Watch(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)

	go func() {
		defer sentry.Recover()
		defer ticker.Stop()
		defer svc.log.Info("stopped")

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// sentry.Recover() at the top of this goroutine only fires
				// once - after that the whole loop (and every future tick)
				// is gone. Recover per-tick instead, so a bug in one pass
				// costs one scan, not the scanner for the rest of the
				// process's life.
				svc.tick(ctx)
			}
		}
	}()
}

func (svc *scanner) tick(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			svc.log.Error("anomaly scan pass panicked, will retry next tick", zap.Any("panic", r))
		}
	}()
	svc.scan(ctx)
}

// scan runs one full pass over every enabled Rule, grouped by module so each
// module's records are only walked once per pass regardless of how many
// fields on it are being watched.
func (svc *scanner) scan(ctx context.Context) {
	ctx = auth.SetIdentityToContext(ctx, auth.ServiceUser())

	rules, _, err := store.SearchAnomalyRules(ctx, svc.store, types.RuleFilter{Enabled: true})
	if err != nil {
		svc.log.Error("failed to load anomaly rules", zap.Error(err))
		return
	}

	type moduleKey struct{ namespaceID, moduleID uint64 }
	byModule := map[moduleKey]types.RuleSet{}
	for _, r := range rules {
		key := moduleKey{r.NamespaceID, r.ModuleID}
		byModule[key] = append(byModule[key], r)
	}

	for key, moduleRules := range byModule {
		svc.scanModule(ctx, key.namespaceID, key.moduleID, moduleRules)
	}
}

func (svc *scanner) scanModule(ctx context.Context, namespaceID, moduleID uint64, rules types.RuleSet) {
	byField := make(map[string]*types.Rule, len(rules))
	for _, r := range rules {
		byField[r.Field] = r
	}

	// @todo scan only records changed since each rule's LastScannedAt
	// (rule.LastScannedAt) instead of the module's whole history every
	// pass - deferred until the Compose record query-expression grammar
	// for an "updatedAt since" filter is confirmed.
	rf := composeTypes.RecordFilter{NamespaceID: namespaceID, ModuleID: moduleID}

	err := svc.record.Iterator(ctx, rf, func(ctx context.Context, ev eventbus.Event) error {
		withRecord, ok := ev.(interface{ Record() *composeTypes.Record })
		if !ok || withRecord.Record() == nil {
			return nil
		}

		svc.scanRecord(ctx, withRecord.Record(), byField)
		return nil
	}, "")

	if err != nil {
		svc.log.Error("failed to scan module",
			zap.Uint64("namespaceID", namespaceID),
			zap.Uint64("moduleID", moduleID),
			zap.Error(err))
		return
	}

	// Record that this pass happened, regardless of whether it found
	// anything - the Anomaly Center surfaces this so admins can tell a
	// quiet field from one nothing has scanned yet.
	now := time.Now()
	for _, rule := range rules {
		rule.LastScannedAt = &now
		if err := store.UpdateAnomalyRule(ctx, svc.store, rule); err != nil {
			svc.log.Error("failed to update rule scan timestamp", zap.Uint64("ruleID", rule.ID), zap.Error(err))
		}
	}
}

func (svc *scanner) scanRecord(ctx context.Context, rec *composeTypes.Record, byField map[string]*types.Rule) {
	for field, rule := range byField {
		rv := rec.Values.Get(field, 0)
		if rv == nil || rv.Value == "" {
			continue
		}
		raw := rv.Value

		detector, ok := detectors[rule.Detector]
		if !ok {
			svc.log.Error("unknown anomaly detector", zap.Uint64("ruleID", rule.ID), zap.String("detector", rule.Detector))
			continue
		}

		baseline, err := svc.loadBaseline(ctx, rule)
		if err != nil {
			svc.log.Error("failed to load baseline", zap.Uint64("ruleID", rule.ID), zap.Error(err))
			continue
		}

		verdict := detector.Detect(raw, baseline.State, baseline.Count, rule)

		baseline.State = detector.UpdateBaseline(raw, baseline.State, baseline.Count)
		baseline.Count++
		if err = svc.saveBaseline(ctx, baseline); err != nil {
			svc.log.Error("failed to save baseline", zap.Uint64("ruleID", rule.ID), zap.Error(err))
		}

		if verdict.Anomalous {
			if err = svc.upsertFinding(ctx, rule, rec, raw, verdict); err != nil {
				svc.log.Error("failed to save finding",
					zap.Uint64("ruleID", rule.ID),
					zap.Uint64("recordID", rec.ID),
					zap.Error(err))
			}
		}
	}
}

// loadBaseline fetches (or initializes, unsaved) the calling rule's
// baseline. Baselines are scoped per rule, not per module+field: two rules
// on the same field can use detectors with mutually incompatible state
// shapes (see detector.go), so sharing a row would let one corrupt the
// other's state.
func (svc *scanner) loadBaseline(ctx context.Context, rule *types.Rule) (*types.Baseline, error) {
	b, err := store.LookupAnomalyBaselineByRuleID(ctx, svc.store, rule.ID)
	if errors.IsNotFound(err) {
		return &types.Baseline{
			NamespaceID: rule.NamespaceID,
			ModuleID:    rule.ModuleID,
			RuleID:      rule.ID,
			Field:       rule.Field,
			State:       types.BaselineState{},
		}, nil
	} else if err != nil {
		return nil, err
	}

	if b.State == nil {
		b.State = types.BaselineState{}
	}
	return b, nil
}

func (svc *scanner) saveBaseline(ctx context.Context, b *types.Baseline) error {
	now := time.Now()
	b.UpdatedAt = &now

	if b.ID == 0 {
		b.ID = id.Next()
		return store.CreateAnomalyBaseline(ctx, svc.store, b)
	}
	return store.UpdateAnomalyBaseline(ctx, svc.store, b)
}

func (svc *scanner) upsertFinding(ctx context.Context, rule *types.Rule, rec *composeTypes.Record, raw string, v Verdict) error {
	now := time.Now()

	explanation := types.FindingExplanation{
		"field":     rule.Field,
		"value":     raw,
		"detector":  rule.Detector,
		"threshold": rule.Threshold,
		"score":     v.Score,
	}

	existing, err := store.LookupAnomalyFindingByRuleIDRecordID(ctx, svc.store, rule.ID, rec.ID)
	if errors.IsNotFound(err) {
		return store.CreateAnomalyFinding(ctx, svc.store, &types.Finding{
			ID:          id.Next(),
			NamespaceID: rule.NamespaceID,
			ModuleID:    rule.ModuleID,
			RecordID:    rec.ID,
			RuleID:      rule.ID,
			Score:       v.Score,
			Severity:    v.Severity,
			Status:      types.StatusNew,
			Explanation: explanation,
			CreatedAt:   now,
			UpdatedAt:   &now,
		})
	} else if err != nil {
		return err
	}

	// Re-flagging an already-resolved/dismissed finding re-opens it: the
	// value is still anomalous now, so a stale "resolved" status would
	// hide an active problem from the Anomaly Center.
	existing.Score = v.Score
	existing.Severity = v.Severity
	existing.Status = types.StatusNew
	existing.Explanation = explanation
	existing.UpdatedAt = &now
	return store.UpdateAnomalyFinding(ctx, svc.store, existing)
}
