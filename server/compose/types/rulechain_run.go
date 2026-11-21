package types

import (
	"encoding/json"
	"time"

	"github.com/madnikulin50/lowcode/server/pkg/rulesgo"
)

type (
	// RuleChainRun is one persisted execution of a rule chain (rulesgo.ExecRecord),
	// kept for tracing/debugging and for the executions dashboard.
	RuleChainRun struct {
		ID          uint64          `json:"runID,string"`
		NamespaceID uint64          `json:"namespaceID,string,omitempty"`
		ChainID     string          `json:"chainID"`
		TriggerType string          `json:"triggerType,omitempty"`
		Success     bool            `json:"success"`
		Error       string          `json:"error,omitempty"`
		Input       json.RawMessage `json:"input,omitempty"`
		Output      json.RawMessage `json:"output,omitempty"`
		Nodes       json.RawMessage `json:"nodes,omitempty"`
		StartedAt   time.Time       `json:"startedAt,omitempty"`
		FinishedAt  *time.Time      `json:"finishedAt,omitempty"`
		DurationMs  int64           `json:"durationMs"`
		CreatedAt   time.Time       `json:"createdAt,omitempty"`
	}

	RuleChainRunFilter struct {
		ChainID     string `json:"chainID"`
		NamespaceID uint64 `json:"namespaceID,string"`
		TriggerType string `json:"triggerType"`

		// nil = don't filter by success
		Success *bool `json:"success"`

		Limit  int `json:"limit"`
		Offset int `json:"offset"`
	}

	RuleChainRunSet []*RuleChainRun
)

// RuleChainRunFromExecRecord converts a rulesgo run-log record into a storable row.
// namespaceID is resolved by the caller (usually looked up from the owning chain)
// since rulesgo.ExecRecord itself carries no namespace context.
func RuleChainRunFromExecRecord(namespaceID uint64, rec rulesgo.ExecRecord) *RuleChainRun {
	run := &RuleChainRun{
		NamespaceID: namespaceID,
		ChainID:     rec.ChainID,
		TriggerType: rec.TriggerType,
		StartedAt:   rec.StartedAt,
	}

	if !rec.FinishedAt.IsZero() {
		finishedAt := rec.FinishedAt
		run.FinishedAt = &finishedAt
		run.DurationMs = rec.FinishedAt.Sub(rec.StartedAt).Milliseconds()
	}

	if input, err := json.Marshal(rec.Input); err == nil {
		run.Input = input
	}

	if rec.Result != nil {
		run.Success = rec.Result.Success
		run.Error = rec.Result.Error
		if output, err := json.Marshal(rec.Result.Output); err == nil {
			run.Output = output
		}
		if nodes, err := json.Marshal(rec.Result.Nodes); err == nil {
			run.Nodes = nodes
		}
	}

	return run
}
