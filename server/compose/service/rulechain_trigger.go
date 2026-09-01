package service

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"

	"github.com/madnikulin50/lowcode/server/compose/types"
	"github.com/madnikulin50/lowcode/server/pkg/auth"
	"github.com/madnikulin50/lowcode/server/pkg/eventbus"
	"github.com/madnikulin50/lowcode/server/pkg/rulesgo"
)

var startRecordTriggersOnce sync.Once

type composeRecordEvent interface {
	eventbus.Event
	Record() *types.Record
	OldRecord() *types.Record
	Module() *types.Module
	Namespace() *types.Namespace
}

// StartRuleChainRecordTriggers subscribes once to compose:record afterCreate/afterUpdate.
func StartRuleChainRecordTriggers(engine *rulesgo.Engine) {
	if engine == nil {
		return
	}
	startRecordTriggersOnce.Do(func() {
		eventbus.Service().Register(func(ctx context.Context, ev eventbus.Event) error {
			runMatchingRecordChains(ctx, engine, ev)
			return nil
		},
			eventbus.For("compose:record"),
			eventbus.On("afterCreate", "afterUpdate"),
			eventbus.Weight(100),
		)
	})
}

func runMatchingRecordChains(ctx context.Context, engine *rulesgo.Engine, ev eventbus.Event) {
	re, ok := ev.(composeRecordEvent)
	if !ok || engine == nil {
		return
	}
	rec := re.Record()
	mod := re.Module()
	ns := re.Namespace()
	if rec == nil || mod == nil {
		return
	}
	eventType := ev.EventType()
	moduleHandle := mod.Handle
	nsID := rec.NamespaceID
	if ns != nil && ns.ID != 0 {
		nsID = ns.ID
	}

	ident := auth.GetIdentityFromContext(ctx)
	bag := recordTriggerBag(rec, mod, nsID)
	old := re.OldRecord()

	for _, chain := range engine.Chains() {
		if chain.NamespaceID != 0 && chain.NamespaceID != nsID {
			continue
		}
		for _, trig := range rulesgo.ParseChainTriggers(chain) {
			if !trig.MatchesEvent("compose:record", eventType, moduleHandle) {
				continue
			}
			field := trig.FileField
			if field == "" {
				field = "file"
			}
			newIDs := fileFieldKey(rec, field)
			if newIDs == "" {
				continue
			}
			if eventType == "afterUpdate" && old != nil && fileFieldKey(old, field) == newIDs {
				continue
			}
			input := copyStringMap(bag)
			input["trigger"] = eventType
			run := func() {
				defer func() {
					if p := recover(); p != nil {
						log.Printf("[rulechain] trigger %s panic: %v", chain.ID, p)
					}
				}()
				bg := context.Background()
				if ident != nil && ident.Valid() {
					bg = auth.SetIdentityToContext(bg, ident)
				}
				if _, err := engine.Run(bg, chain.ID, input); err != nil {
					log.Printf("[rulechain] trigger %s: %v", chain.ID, err)
				}
			}
			if trig.Async {
				go run()
			} else {
				run()
			}
		}
	}
}

func recordTriggerBag(rec *types.Record, mod *types.Module, nsID uint64) map[string]interface{} {
	out := map[string]interface{}{}
	if rec != nil {
		out["recordID"] = fmt.Sprintf("%d", rec.ID)
		out["documentID"] = fmt.Sprintf("%d", rec.ID)
		grouped := map[string][]string{}
		for _, v := range rec.Values {
			if v == nil || v.DeletedAt != nil {
				continue
			}
			grouped[v.Name] = append(grouped[v.Name], v.Value)
		}
		for name, vals := range grouped {
			if len(vals) == 1 {
				out[name] = vals[0]
			} else {
				out[name] = strings.Join(vals, "\n")
			}
		}
	}
	if mod != nil {
		out["moduleID"] = fmt.Sprintf("%d", mod.ID)
		out["moduleHandle"] = mod.Handle
	}
	if nsID > 0 {
		out["namespaceID"] = fmt.Sprintf("%d", nsID)
	}
	return out
}

func fileFieldKey(rec *types.Record, field string) string {
	if rec == nil || field == "" {
		return ""
	}
	var raw []string
	for _, v := range rec.Values.FilterByName(field) {
		if v == nil || v.DeletedAt != nil {
			continue
		}
		raw = append(raw, v.Value)
	}
	ids := rulesgo.AttachmentIDsFromValue(strings.Join(raw, "\n"))
	if len(ids) == 0 {
		return ""
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	parts := make([]string, len(ids))
	for i, id := range ids {
		parts[i] = fmt.Sprintf("%d", id)
	}
	return strings.Join(parts, ",")
}

func copyStringMap(in map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
