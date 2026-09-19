package rulesgo

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

// recordingPersistence wraps MemoryPersistence and also implements
// RunLogPersistence, so we can assert that RunWithLog actually calls it.
type recordingPersistence struct {
	*MemoryPersistence
	saved chan ExecRecord
}

func newRecordingPersistence() *recordingPersistence {
	return &recordingPersistence{
		MemoryPersistence: NewMemoryPersistence(),
		saved:             make(chan ExecRecord, 4),
	}
}

func (r *recordingPersistence) SaveRun(ctx context.Context, rec ExecRecord) error {
	r.saved <- rec
	return nil
}

func TestEngineWithPersistence_RunWithLog_PersistsRun(t *testing.T) {
	persist := newRecordingPersistence()
	engine := NewEngineWithPersistence(DefaultRegistry(&DefaultConfig{}), persist)

	chain := &Chain{
		ID:        "test_run_with_log",
		Name:      "Test",
		EntryNode: "n1",
		Nodes: []ChainNode{
			{ID: "n1", Type: "condition", Config: json.RawMessage(`{"field":"x","operator":"eq","value":"1"}`)},
		},
	}
	engine.RegisterChain(chain)

	result, err := engine.RunWithLog(context.Background(), chain.ID, map[string]interface{}{"x": 1}, "unit-test")
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
	if !result.Success {
		t.Fatalf("expected success, got: %s", result.Error)
	}

	// in-memory log (L1) should have it immediately
	logs := engine.ExecutionLogs(0)
	if len(logs) != 1 {
		t.Fatalf("expected 1 in-memory log entry, got %d", len(logs))
	}
	if logs[0].TriggerType != "unit-test" {
		t.Fatalf("unexpected triggerType: %s", logs[0].TriggerType)
	}

	// durable persistence (RunLogPersistence.SaveRun) happens in a detached
	// goroutine, so wait for it instead of asserting synchronously.
	select {
	case rec := <-persist.saved:
		if rec.ChainID != chain.ID {
			t.Fatalf("unexpected chainID persisted: %s", rec.ChainID)
		}
		if rec.TriggerType != "unit-test" {
			t.Fatalf("unexpected triggerType persisted: %s", rec.TriggerType)
		}
		if rec.Result == nil || !rec.Result.Success {
			t.Fatalf("expected persisted result to be successful")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("SaveRun was not called within timeout")
	}
}

func TestEngineWithPersistence_RunWithLog_MemoryOnlyPersistenceIsFine(t *testing.T) {
	// MemoryPersistence does NOT implement RunLogPersistence - RunWithLog
	// must not panic or block, it should just skip durable persistence.
	engine := NewEngineWithPersistence(DefaultRegistry(&DefaultConfig{}), NewMemoryPersistence())

	chain := &Chain{
		ID:        "test_run_with_log_memory",
		Name:      "Test",
		EntryNode: "n1",
		Nodes: []ChainNode{
			{ID: "n1", Type: "condition", Config: json.RawMessage(`{"field":"x","operator":"eq","value":"1"}`)},
		},
	}
	engine.RegisterChain(chain)

	if _, err := engine.RunWithLog(context.Background(), chain.ID, map[string]interface{}{"x": 1}, "unit-test"); err != nil {
		t.Fatalf("run failed: %v", err)
	}
}
