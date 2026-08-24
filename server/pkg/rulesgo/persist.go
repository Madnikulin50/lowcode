package rulesgo

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

type Persistence interface {
	LoadChains(ctx context.Context) ([]*Chain, error)
	SaveChain(ctx context.Context, chain *Chain) error
	DeleteChain(ctx context.Context, chainID string) error
	UpdateChain(ctx context.Context, chain *Chain) error
}

type MemoryPersistence struct {
	chains map[string]*Chain
}

func NewMemoryPersistence() *MemoryPersistence {
	return &MemoryPersistence{chains: make(map[string]*Chain)}
}

func (mp *MemoryPersistence) LoadChains(ctx context.Context) ([]*Chain, error) {
	result := make([]*Chain, 0, len(mp.chains))
	for _, c := range mp.chains {
		result = append(result, c)
	}
	return result, nil
}

func (mp *MemoryPersistence) SaveChain(ctx context.Context, chain *Chain) error {
	mp.chains[chain.ID] = chain
	return nil
}

func (mp *MemoryPersistence) DeleteChain(ctx context.Context, chainID string) error {
	delete(mp.chains, chainID)
	return nil
}

func (mp *MemoryPersistence) UpdateChain(ctx context.Context, chain *Chain) error {
	mp.chains[chain.ID] = chain
	return nil
}

type ExecRecord struct {
	ChainID     string                 `json:"chainID"`
	Input       map[string]interface{} `json:"input"`
	Result      *ChainResult           `json:"result"`
	StartedAt   time.Time              `json:"startedAt"`
	FinishedAt  time.Time              `json:"finishedAt"`
	Duration    string                 `json:"duration"`
	TriggerType string                 `json:"triggerType"` // manual, webhook, scheduled, ai
}

type ExecutionLog struct {
	records []ExecRecord
	maxSize int
}

func NewExecutionLog(maxSize int) *ExecutionLog {
	if maxSize <= 0 {
		maxSize = 1000
	}
	return &ExecutionLog{
		records: make([]ExecRecord, 0, maxSize),
		maxSize: maxSize,
	}
}

func (el *ExecutionLog) Add(record ExecRecord) {
	el.records = append(el.records, record)
	if len(el.records) > el.maxSize {
		el.records = el.records[len(el.records)-el.maxSize:]
	}
}

func (el *ExecutionLog) Recent(limit int) []ExecRecord {
	if limit <= 0 || limit > len(el.records) {
		limit = len(el.records)
	}
	start := len(el.records) - limit
	if start < 0 {
		start = 0
	}
	return el.records[start:]
}

type EngineWithPersistence struct {
	*Engine
	persist Persistence
	log     *ExecutionLog
}

func NewEngineWithPersistence(registry *Registry, persist Persistence) *EngineWithPersistence {
	eng := NewEngine(registry)
	eng.SetPersistence(persist)
	return &EngineWithPersistence{
		Engine:  eng,
		persist: persist,
		log:     NewExecutionLog(1000),
	}
}

func (e *EngineWithPersistence) LoadFromStore(ctx context.Context) error {
	chains, err := e.persist.LoadChains(ctx)
	if err != nil {
		return fmt.Errorf("load chains: %w", err)
	}
	for _, c := range chains {
		e.put(c)
	}
	log.Printf("[rulesgo] loaded %d chains from store", len(chains))
	return nil
}

func (e *EngineWithPersistence) CreateChain(ctx context.Context, chain *Chain) error {
	if err := e.persist.SaveChain(ctx, chain); err != nil {
		return err
	}
	e.RegisterChain(chain)
	return nil
}

func (e *EngineWithPersistence) DeleteChain(ctx context.Context, chainID string) error {
	if err := e.persist.DeleteChain(ctx, chainID); err != nil {
		return err
	}
	delete(e.chains, chainID)
	return nil
}

func (e *EngineWithPersistence) RunWithLog(ctx context.Context, chainID string, input map[string]interface{}, triggerType string) (*ChainResult, error) {
	start := time.Now()
	result, err := e.Engine.Run(ctx, chainID, input)
	finished := time.Now()

	record := ExecRecord{
		ChainID:     chainID,
		Input:       input,
		Result:      result,
		StartedAt:   start,
		FinishedAt:  finished,
		Duration:    finished.Sub(start).String(),
		TriggerType: triggerType,
	}
	e.log.Add(record)

	return result, err
}

func (e *EngineWithPersistence) ExecutionLogs(limit int) []ExecRecord {
	return e.log.Recent(limit)
}

func (e *EngineWithPersistence) ExportChain(chainID string) ([]byte, error) {
	chain := e.Chain(chainID)
	if chain == nil {
		return nil, fmt.Errorf("chain not found: %s", chainID)
	}
	return json.MarshalIndent(chain, "", "  ")
}

func (e *EngineWithPersistence) ImportChain(data []byte) (*Chain, error) {
	var chain Chain
	if err := json.Unmarshal(data, &chain); err != nil {
		return nil, fmt.Errorf("invalid chain JSON: %w", err)
	}
	if chain.ID == "" {
		return nil, fmt.Errorf("chain must have an ID")
	}
	e.RegisterChain(&chain)
	return &chain, nil
}
