package rulesgo

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

type Engine struct {
	registry *Registry
	chains   map[string]*Chain
	persist  Persistence
}

func NewEngine(registry *Registry) *Engine {
	return &Engine{
		registry: registry,
		chains:   make(map[string]*Chain),
	}
}

func (e *Engine) SetPersistence(p Persistence) {
	e.persist = p
}

func (e *Engine) put(chain *Chain) {
	if chain == nil || chain.ID == "" {
		return
	}
	e.chains[chain.ID] = chain
}

func (e *Engine) RegisterChain(chain *Chain) {
	e.put(chain)
	if e.persist != nil {
		if err := e.persist.SaveChain(context.Background(), chain); err != nil {
			log.Printf("[rulesgo] persist %s: %v", chain.ID, err)
		}
	}
}

func (e *Engine) PersistChain(chain *Chain) error {
	e.put(chain)
	if e.persist == nil {
		return nil
	}
	return e.persist.SaveChain(context.Background(), chain)
}

func (e *Engine) Chains() []*Chain {
	result := make([]*Chain, 0, len(e.chains))
	for _, c := range e.chains {
		result = append(result, c)
	}
	return result
}

func (e *Engine) Chain(id string) *Chain {
	return e.chains[id]
}

func (e *Engine) Run(ctx context.Context, chainID string, input map[string]interface{}) (*ChainResult, error) {
	chain := e.chains[chainID]
	if chain == nil {
		return nil, fmt.Errorf("chain not found: %s", chainID)
	}

	ec := &ExecutionContext{
		Variables: make(map[string]interface{}),
		Results:   make(map[string]interface{}),
		Input:     input,
	}

	result := &ChainResult{
		ChainID: chainID,
		Output:  make(map[string]interface{}),
	}

	nodeMap := make(map[string]*ChainNode)
	for i := range chain.Nodes {
		nodeMap[chain.Nodes[i].ID] = &chain.Nodes[i]
	}

	edgeMap := make(map[string][]ChainEdge)
	for _, edge := range chain.Edges {
		edgeMap[edge.From] = append(edgeMap[edge.From], edge)
	}

	visited := make(map[string]bool)
	queue := []string{chain.EntryNode}

	for len(queue) > 0 {
		currentID := queue[0]
		queue = queue[1:]

		if visited[currentID] {
			continue
		}
		visited[currentID] = true

		node := nodeMap[currentID]
		if node == nil {
			result.Error = fmt.Sprintf("node not found: %s", currentID)
			result.Success = false
			break
		}

		if node.Type == "foreach" {
			bodyIDs := edgesTo(edgeMap[currentID])
			if len(bodyIDs) == 0 {
				bodyIDs = nextNodeIDs(edgeMap[currentID], ec)
			}
			start := time.Now()
			nr, err := e.runForeach(ctx, node, nodeMap, edgeMap, bodyIDs, ec)
			elapsed := time.Since(start)
			result.Nodes = append(result.Nodes, nr)
			if err != nil {
				result.Error = fmt.Sprintf("node %s (foreach) failed: %v", node.ID, err)
				result.Success = false
				log.Printf("[rulesgo] node %s (foreach) FAILED in %v: %v", node.ID, elapsed, err)
				break
			}
			log.Printf("[rulesgo] node %s (foreach) OK in %v items=%v", node.ID, elapsed, nr.Output["count"])
			markDescendants(currentID, edgeMap, visited)
			continue
		}

		start := time.Now()
		output, err := e.registry.Execute(ctx, node.Type, *node, ec)
		elapsed := time.Since(start)

		nextIDs := nextNodeIDs(edgeMap[currentID], ec)
		nodeResult := NodeResult{
			NodeID: node.ID,
			Type:   node.Type,
			Output: output,
			Next:   nextIDs,
		}

		if err != nil {
			nodeResult.Error = err.Error()
			result.Nodes = append(result.Nodes, nodeResult)
			result.Error = fmt.Sprintf("node %s (%s) failed: %v", node.ID, node.Type, err)
			result.Success = false
			log.Printf("[rulesgo] node %s (%s) FAILED in %v: %v", node.ID, node.Type, elapsed, err)
			break
		}

		if output != nil {
			ec.SetResult(node.ID, output)
		}

		log.Printf("[rulesgo] node %s (%s) OK in %v", node.ID, node.Type, elapsed)
		result.Nodes = append(result.Nodes, nodeResult)

		for _, nid := range nextIDs {
			if !visited[nid] {
				queue = append(queue, nid)
			}
		}
	}

	if result.Error == "" {
		result.Success = true
	}
	result.Output = ec.Variables

	return result, nil
}

func nextNodeIDs(edges []ChainEdge, ec *ExecutionContext) []string {
	nextIDs := make([]string, 0)
	for _, edge := range edges {
		if edge.Condition == "" || ec.GetString(edge.Condition) != "" {
			nextIDs = append(nextIDs, edge.To)
		}
	}
	if len(nextIDs) == 0 {
		return edgesTo(edges)
	}
	return nextIDs
}

func edgesTo(edges []ChainEdge) []string {
	result := make([]string, 0, len(edges))
	for _, e := range edges {
		if e.Condition == "" {
			result = append(result, e.To)
		}
	}
	return result
}

func markDescendants(from string, edgeMap map[string][]ChainEdge, visited map[string]bool) {
	for _, to := range edgesTo(edgeMap[from]) {
		if visited[to] {
			continue
		}
		visited[to] = true
		markDescendants(to, edgeMap, visited)
	}
}

func (e *Engine) ExportChain(chainID string) ([]byte, error) {
	chain := e.Chain(chainID)
	if chain == nil {
		return nil, fmt.Errorf("chain not found: %s", chainID)
	}
	data, err := json.MarshalIndent(chain, "", "  ")
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (e *Engine) ImportChain(data []byte) (*Chain, error) {
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

func (e *Engine) DeleteChain(chainID string) {
	delete(e.chains, chainID)
	if e.persist != nil {
		if err := e.persist.DeleteChain(context.Background(), chainID); err != nil {
			log.Printf("[rulesgo] delete persist %s: %v", chainID, err)
		}
	}
}
