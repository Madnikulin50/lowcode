package riskengine

import "fmt"

// bnNode is the resolved, evaluation-ready view of one RiskNode: its own
// state codes, its parents in CPT column order, and either a CPT (non-root)
// or a prior distribution over its own states (root — used only when the
// node's value is not part of the given evidence, i.e. during attribution
// marginalization; see Explain).
type bnNode struct {
	id      string
	states  []string
	parents []string // node IDs, in the order they must be looked up in cpt.ParentStates
	cpt     *CPT
	prior   []float64 // len(states); uniform by default
}

// bayesNet is a small discrete Bayesian network built once from a
// RiskModel + its factor library, then queried many times (Evaluate, and
// once per coalition during Explain's Shapley enumeration).
type bayesNet struct {
	nodes  map[string]*bnNode
	order  []string          // topological order, root-first
	handle map[string]string // node ID -> RiskFactorDef.Handle, for evidence lookup
}

// buildBayesNet resolves a RiskModel's graph into a bayesNet. factors must
// contain every RiskFactorDef referenced by model.Nodes. priors optionally
// overrides the default uniform prior for root nodes (keyed by RiskNode.ID);
// only used when Explain marginalizes that node.
func buildBayesNet(model *RiskModel, factors map[uint64]*RiskFactorDef, priors map[string][]float64) (*bayesNet, error) {
	net := &bayesNet{
		nodes:  make(map[string]*bnNode, len(model.Nodes)),
		handle: make(map[string]string, len(model.Nodes)),
	}

	for _, n := range model.Nodes {
		fd, ok := factors[n.FactorID]
		if !ok {
			return nil, fmt.Errorf("riskengine: node %q references unknown factor %d", n.ID, n.FactorID)
		}
		if fd.Scale != ScaleStates {
			return nil, fmt.Errorf("riskengine: node %q's factor %q is not a states-scale factor", n.ID, fd.Handle)
		}
		states := make([]string, len(fd.States))
		for i, s := range fd.States {
			states[i] = s.Code
		}
		net.nodes[n.ID] = &bnNode{id: n.ID, states: states, cpt: n.CPT}
		net.handle[n.ID] = fd.Handle
	}

	// Parents come from edges, in declaration order — this is also the
	// order a node's own CPT.ParentStates must follow.
	for _, e := range model.Edges {
		child, ok := net.nodes[e.To]
		if !ok {
			return nil, fmt.Errorf("riskengine: edge references unknown node %q", e.To)
		}
		if _, ok := net.nodes[e.From]; !ok {
			return nil, fmt.Errorf("riskengine: edge references unknown node %q", e.From)
		}
		child.parents = append(child.parents, e.From)
	}

	for id, n := range net.nodes {
		if n.cpt != nil {
			if len(n.parents) != len(n.cpt.ParentStates) {
				return nil, fmt.Errorf("riskengine: node %q CPT has %d parent dimensions, graph has %d incoming edges", id, len(n.cpt.ParentStates), len(n.parents))
			}
			continue
		}
		// Root node: uniform prior unless overridden.
		if p, ok := priors[id]; ok {
			if len(p) != len(n.states) {
				return nil, fmt.Errorf("riskengine: prior for node %q has %d entries, factor has %d states", id, len(p), len(n.states))
			}
			n.prior = p
			continue
		}
		n.prior = uniform(len(n.states))
	}

	order, err := topoSort(net)
	if err != nil {
		return nil, err
	}
	net.order = order
	return net, nil
}

func uniform(n int) []float64 {
	p := make([]float64, n)
	if n == 0 {
		return p
	}
	v := 1.0 / float64(n)
	for i := range p {
		p[i] = v
	}
	return p
}

func topoSort(net *bayesNet) ([]string, error) {
	indeg := make(map[string]int, len(net.nodes))
	children := make(map[string][]string, len(net.nodes))
	for id, n := range net.nodes {
		indeg[id] += len(n.parents)
		for _, p := range n.parents {
			children[p] = append(children[p], id)
		}
	}

	var queue, order []string
	for id, d := range indeg {
		if d == 0 {
			queue = append(queue, id)
		}
	}
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		order = append(order, id)
		for _, c := range children[id] {
			indeg[c]--
			if indeg[c] == 0 {
				queue = append(queue, c)
			}
		}
	}
	if len(order) != len(net.nodes) {
		return nil, fmt.Errorf("riskengine: bayes graph has a cycle")
	}
	return order, nil
}

func stateIndex(states []string, code string) int {
	for i, s := range states {
		if s == code {
			return i
		}
	}
	return -1
}

// cptRowIndex computes the mixed-radix row into cpt.Rows for the given
// parent-value assignment, using n.parents for column order and
// cpt.ParentStates for each column's radix. The last parent varies fastest.
func cptRowIndex(n *bnNode, assignment map[string]string) (int, error) {
	row := 0
	for i, parentID := range n.parents {
		val, ok := assignment[parentID]
		if !ok {
			return 0, fmt.Errorf("riskengine: node %q missing assignment for parent %q", n.id, parentID)
		}
		idx := stateIndex(n.cpt.ParentStates[i], val)
		if idx < 0 {
			return 0, fmt.Errorf("riskengine: node %q parent %q has unknown state %q", n.id, parentID, val)
		}
		row = row*len(n.cpt.ParentStates[i]) + idx
	}
	return row, nil
}

// prob returns P(node = value | assignment of node's parents), reading a CPT
// row for non-root nodes or the prior for root nodes.
func (net *bayesNet) prob(nodeID, value string, assignment map[string]string) (float64, error) {
	n := net.nodes[nodeID]
	if n.cpt == nil {
		idx := stateIndex(n.states, value)
		if idx < 0 {
			return 0, fmt.Errorf("riskengine: node %q has unknown state %q", nodeID, value)
		}
		return n.prior[idx], nil
	}
	row, err := cptRowIndex(n, assignment)
	if err != nil {
		return 0, err
	}
	if row >= len(n.cpt.Rows) {
		return 0, fmt.Errorf("riskengine: node %q CPT missing row %d", nodeID, row)
	}
	col := stateIndex(n.cpt.NodeStates, value)
	if col < 0 {
		return 0, fmt.Errorf("riskengine: node %q has unknown state %q", nodeID, value)
	}
	return n.cpt.Rows[row][col], nil
}

// posterior computes the (possibly unnormalized then normalized) posterior
// distribution over queryNode's states given a partial evidence assignment,
// via exact enumeration (AIMA's enumeration-ask): every node not fixed in
// evidence is summed out weighted by its own CPT/prior — which is exactly
// probabilistic marginalization, not mean-imputation.
func (net *bayesNet) posterior(queryNode string, evidence map[string]string) (map[string]float64, error) {
	q := net.nodes[queryNode]
	out := make(map[string]float64, len(q.states))
	var sum float64
	for _, v := range q.states {
		e := cloneAssignment(evidence)
		e[queryNode] = v
		p, err := net.enumerateAll(net.order, e)
		if err != nil {
			return nil, err
		}
		out[v] = p
		sum += p
	}
	if sum <= 0 {
		return nil, fmt.Errorf("riskengine: evidence has zero probability")
	}
	for v := range out {
		out[v] /= sum
	}
	return out, nil
}

func (net *bayesNet) enumerateAll(vars []string, e map[string]string) (float64, error) {
	if len(vars) == 0 {
		return 1, nil
	}
	y, rest := vars[0], vars[1:]
	n := net.nodes[y]

	if v, ok := e[y]; ok {
		p, err := net.prob(y, v, e)
		if err != nil {
			return 0, err
		}
		restP, err := net.enumerateAll(rest, e)
		if err != nil {
			return 0, err
		}
		return p * restP, nil
	}

	var total float64
	for _, v := range n.states {
		p, err := net.prob(y, v, e)
		if err != nil {
			return 0, err
		}
		if p == 0 {
			continue
		}
		e2 := cloneAssignment(e)
		e2[y] = v
		restP, err := net.enumerateAll(rest, e2)
		if err != nil {
			return 0, err
		}
		total += p * restP
	}
	return total, nil
}

func cloneAssignment(e map[string]string) map[string]string {
	c := make(map[string]string, len(e)+1)
	for k, v := range e {
		c[k] = v
	}
	return c
}
