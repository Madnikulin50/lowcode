// This file implements Explain — the "Assessment -> Attribution" step. It
// reads an already-produced RiskAssessment (score, evidence) and never
// re-derives it from raw domain data, so a Layer-C narrative built on top of
// AttributionResult explains a computed number instead of reasoning about
// risk from scratch.
package riskengine

import (
	"encoding/json"
	"fmt"
	"math/bits"
	"math/rand"
	"sort"
	"strconv"
	"strings"
)

const (
	MethodClosedForm = "closed-form"
	MethodExact      = "exact"
	MethodSampled    = "sampled"
	MethodBlend      = "blend"

	// defaultAttributionMaxExact is used when RiskModel.AttributionMaxExact
	// is zero — the threshold agreed for the bayes strategy: exact 2^n
	// Shapley enumeration up to this many evidence factors, permutation
	// sampling above it.
	defaultAttributionMaxExact = 15
	sampledPermutations        = 500
)

// Explain computes the SHAP-style AttributionResult for model/evidence.
// assessment is the RiskAssessment Evaluate already produced for the same
// inputs — Explain trusts its InherentScore rather than recomputing it,
// so the breakdown is guaranteed to reconcile with what was reported.
func Explain(model *RiskModel, factors map[uint64]*RiskFactorDef, evidence Evidence, assessment *RiskAssessment) (*AttributionResult, error) {
	switch model.Strategy {
	case StrategyWeighted:
		return explainWeighted(model, factors, evidence)
	case StrategyMatrix:
		return explainMatrix(model, factors, evidence)
	case StrategyBayes:
		return explainBayes(model, factors, evidence)
	case StrategyBlend:
		return explainBlend(model, factors, evidence)
	default:
		return nil, fmt.Errorf("riskengine: unknown strategy %q", model.Strategy)
	}
}

// ---------------------------------------------------------------------------
// weighted — exact, closed form (Shapley value of a linear function is each
// term directly; no enumeration needed)
// ---------------------------------------------------------------------------

func explainWeighted(model *RiskModel, factors map[uint64]*RiskFactorDef, evidence Evidence) (*AttributionResult, error) {
	var cfg WeightedConfig
	if len(model.Config) > 0 {
		if err := json.Unmarshal(model.Config, &cfg); err != nil {
			return nil, fmt.Errorf("riskengine: invalid weighted config: %w", err)
		}
	}
	normalizeOut := cfg.Normalize == nil || *cfg.Normalize
	scaleMax := cfg.ScaleMax
	if scaleMax <= 0 {
		scaleMax = 100
	}

	out, err := findOutputNode(model.Nodes)
	if err != nil {
		return nil, err
	}

	var weightSum float64
	type term struct {
		fd      *RiskFactorDef
		nodeID  string
		weight  float64
		norm    float64
		baseNrm float64
	}
	var terms []term
	for _, e := range model.Edges {
		if e.To != out.ID || e.Weight == 0 {
			continue
		}
		from, ok := nodeByID(model.Nodes, e.From)
		if !ok {
			continue
		}
		fd, rawVal, ok := evidenceRaw(from, factors, evidence)
		if !ok {
			return nil, fmt.Errorf("riskengine: missing evidence for factor referenced by node %q", from.ID)
		}
		norm, err := normalize(fd, rawVal)
		if err != nil {
			return nil, err
		}
		weightSum += e.Weight
		terms = append(terms, term{fd: fd, nodeID: from.ID, weight: e.Weight, norm: norm, baseNrm: baselineNormalized(fd)})
	}
	if weightSum <= 0 {
		return nil, fmt.Errorf("riskengine: weighted strategy: total weight must be > 0")
	}

	scaleFactor := 1.0
	if normalizeOut {
		scaleFactor = scaleMax / weightSum
	}

	var baseline float64
	contribs := make([]FactorContribution, 0, len(terms))
	for _, t := range terms {
		baseline += t.weight * t.baseNrm * scaleFactor
		contribs = append(contribs, FactorContribution{
			NodeID:   t.nodeID,
			Handle:   t.fd.Handle,
			Value:    t.weight * (t.norm - t.baseNrm) * scaleFactor,
			Baseline: round2(t.baseNrm),
		})
	}

	return &AttributionResult{Method: MethodClosedForm, Baseline: baseline, Contributions: contribs}, nil
}

// ---------------------------------------------------------------------------
// matrix — exact 2-player Shapley over {likelihood, impact}, masking the
// excluded input to its factor's Baseline (interventional, point reference)
// ---------------------------------------------------------------------------

func explainMatrix(model *RiskModel, factors map[uint64]*RiskFactorDef, evidence Evidence) (*AttributionResult, error) {
	var cfg MatrixConfig
	if err := json.Unmarshal(model.Config, &cfg); err != nil {
		return nil, fmt.Errorf("riskengine: invalid matrix config: %w", err)
	}
	scale := cfg.ScaleSize
	if scale <= 0 {
		scale = 5
	}

	lNode, ok := nodeByID(model.Nodes, cfg.LikelihoodNode)
	if !ok {
		return nil, fmt.Errorf("riskengine: matrix config: unknown likelihoodNode %q", cfg.LikelihoodNode)
	}
	iNode, ok := nodeByID(model.Nodes, cfg.ImpactNode)
	if !ok {
		return nil, fmt.Errorf("riskengine: matrix config: unknown impactNode %q", cfg.ImpactNode)
	}
	lFD, lRaw, ok := evidenceRaw(lNode, factors, evidence)
	if !ok {
		return nil, fmt.Errorf("riskengine: missing evidence for likelihood node %q", lNode.ID)
	}
	iFD, iRaw, ok := evidenceRaw(iNode, factors, evidence)
	if !ok {
		return nil, fmt.Errorf("riskengine: missing evidence for impact node %q", iNode.ID)
	}

	li := clampInt(int(toFloat(lRaw)+0.5), 1, scale)
	ii := clampInt(int(toFloat(iRaw)+0.5), 1, scale)
	lBase := matrixBaselineRaw(lFD, scale)
	iBase := matrixBaselineRaw(iFD, scale)

	value := func(mask int) float64 {
		l, i := lBase, iBase
		if mask&1 != 0 {
			l = float64(li)
		}
		if mask&2 != 0 {
			i = float64(ii)
		}
		s, _ := matrixCellScore(clampInt(int(l+0.5), 1, scale), clampInt(int(i+0.5), 1, scale), cfg)
		return s
	}

	phi := exactShapley(2, value)
	return &AttributionResult{
		Method:   MethodExact,
		Baseline: value(0),
		Contributions: []FactorContribution{
			{NodeID: lNode.ID, Handle: lFD.Handle, Value: phi[0], Baseline: round2(lBase)},
			{NodeID: iNode.ID, Handle: iFD.Handle, Value: phi[1], Baseline: round2(iBase)},
		},
	}, nil
}

func matrixBaselineRaw(fd *RiskFactorDef, scale int) float64 {
	if fd.Baseline != "" {
		if v, err := strconv.ParseFloat(strings.TrimSpace(fd.Baseline), 64); err == nil {
			return v
		}
	}
	return float64(scale+1) / 2 // scale midpoint, e.g. 3 on a 1..5 scale
}

// ---------------------------------------------------------------------------
// bayes — exact Shapley via subset marginalization through the CPTs
// (up to AttributionMaxExact evidence factors), permutation sampling above it
// ---------------------------------------------------------------------------

func explainBayes(model *RiskModel, factors map[uint64]*RiskFactorDef, evidence Evidence) (*AttributionResult, error) {
	var cfg BayesConfig
	if err := json.Unmarshal(model.Config, &cfg); err != nil {
		return nil, fmt.Errorf("riskengine: invalid bayes config: %w", err)
	}

	net, err := buildBayesNet(model, factors, nil)
	if err != nil {
		return nil, err
	}
	if _, ok := net.nodes[cfg.OutputNode]; !ok {
		return nil, fmt.Errorf("riskengine: bayes config: unknown outputNode %q", cfg.OutputNode)
	}
	full, err := bayesEvidenceAssignment(net, factors, evidence)
	if err != nil {
		return nil, err
	}

	// Stable ordering: topological order restricted to evidence nodes.
	var evidenceNodes []string
	for _, id := range net.order {
		if _, ok := full[id]; ok {
			evidenceNodes = append(evidenceNodes, id)
		}
	}
	n := len(evidenceNodes)

	expectedScore := func(assignment map[string]string) (float64, error) {
		post, err := net.posterior(cfg.OutputNode, assignment)
		if err != nil {
			return 0, err
		}
		var s float64
		for state, p := range post {
			s += p * cfg.ScoreByState[state]
		}
		return s, nil
	}

	var evalErr error
	value := func(mask int) float64 {
		assignment := make(map[string]string, n)
		for i, id := range evidenceNodes {
			if mask&(1<<uint(i)) != 0 {
				assignment[id] = full[id]
			}
		}
		s, err := expectedScore(assignment)
		if err != nil {
			evalErr = err
		}
		return s
	}

	maxExact := model.AttributionMaxExact
	if maxExact <= 0 {
		maxExact = defaultAttributionMaxExact
	}

	var phi []float64
	method := MethodExact
	sampleCount := 0
	if n <= maxExact {
		phi = exactShapley(n, value)
	} else {
		method = MethodSampled
		sampleCount = sampledPermutations
		phi = sampledShapley(n, value, sampleCount)
	}
	if evalErr != nil {
		return nil, evalErr
	}

	baseline := value(0)
	contribs := make([]FactorContribution, n)
	for i, id := range evidenceNodes {
		contribs[i] = FactorContribution{
			NodeID: id,
			Handle: net.handle[id],
			Value:  phi[i],
			// No single point reference exists for a marginalized bayes
			// factor — it's averaged over its prior, not pinned to one value.
			Baseline: 0,
		}
	}

	return &AttributionResult{Method: method, Baseline: baseline, Contributions: contribs, SampleCount: sampleCount}, nil
}

// ---------------------------------------------------------------------------
// blend — additive: combine each component's own AttributionResult with the
// blend's weights, never recomputed from scratch
// ---------------------------------------------------------------------------

func explainBlend(model *RiskModel, factors map[uint64]*RiskFactorDef, evidence Evidence) (*AttributionResult, error) {
	var cfg BlendConfig
	if err := json.Unmarshal(model.Config, &cfg); err != nil {
		return nil, fmt.Errorf("riskengine: invalid blend config: %w", err)
	}
	if len(cfg.Components) == 0 {
		return nil, fmt.Errorf("riskengine: blend strategy requires at least one component")
	}

	var weightSum, baseline float64
	byNode := map[string]*FactorContribution{}
	var order []string
	for _, c := range cfg.Components {
		sub := &RiskModel{
			Nodes:               model.Nodes,
			Edges:               model.Edges,
			Strategy:            c.Strategy,
			Config:              c.Config,
			AttributionMaxExact: model.AttributionMaxExact,
		}
		res, err := Explain(sub, factors, evidence, nil)
		if err != nil {
			return nil, fmt.Errorf("riskengine: blend component %q: %w", c.Strategy, err)
		}
		weightSum += c.Weight
		baseline += c.Weight * res.Baseline
		for _, fc := range res.Contributions {
			if existing, ok := byNode[fc.NodeID]; ok {
				existing.Value += c.Weight * fc.Value
			} else {
				copyFC := FactorContribution{NodeID: fc.NodeID, Handle: fc.Handle, Value: c.Weight * fc.Value, Baseline: fc.Baseline}
				byNode[fc.NodeID] = &copyFC
				order = append(order, fc.NodeID)
			}
		}
	}
	if weightSum <= 0 {
		return nil, fmt.Errorf("riskengine: blend strategy: total component weight must be > 0")
	}

	contribs := make([]FactorContribution, 0, len(order))
	for _, id := range order {
		fc := *byNode[id]
		fc.Value = fc.Value / weightSum
		contribs = append(contribs, fc)
	}
	sort.Slice(contribs, func(i, j int) bool { return contribs[i].NodeID < contribs[j].NodeID })

	return &AttributionResult{Method: MethodBlend, Baseline: baseline / weightSum, Contributions: contribs}, nil
}

// ---------------------------------------------------------------------------
// generic exact / sampled Shapley value, shared by matrix and bayes
// ---------------------------------------------------------------------------

// exactShapley computes Shapley values for an n-player game via full 2^n
// coalition enumeration. v is called once per coalition bitmask (bit i set
// means player i is "in"); results are cached internally so each mask is
// evaluated exactly once regardless of n.
func exactShapley(n int, v func(mask int) float64) []float64 {
	if n == 0 {
		return nil
	}
	total := 1 << uint(n)
	vals := make([]float64, total)
	for m := 0; m < total; m++ {
		vals[m] = v(m)
	}
	fact := factorials(n)
	phi := make([]float64, n)
	for m := 0; m < total; m++ {
		s := bits.OnesCount(uint(m))
		for i := 0; i < n; i++ {
			bit := 1 << uint(i)
			if m&bit != 0 {
				continue
			}
			weight := fact[s] * fact[n-s-1] / fact[n]
			phi[i] += weight * (vals[m|bit] - vals[m])
		}
	}
	return phi
}

func factorials(n int) []float64 {
	f := make([]float64, n+1)
	f[0] = 1
	for i := 1; i <= n; i++ {
		f[i] = f[i-1] * float64(i)
	}
	return f
}

// sampledShapley approximates Shapley values via permutation sampling
// (standard Monte Carlo Shapley) for n too large for exact 2^n enumeration.
// The seed is fixed so the same evidence always yields the same explanation.
func sampledShapley(n int, v func(mask int) float64, samples int) []float64 {
	phi := make([]float64, n)
	rng := rand.New(rand.NewSource(42))
	perm := make([]int, n)
	for i := range perm {
		perm[i] = i
	}
	for s := 0; s < samples; s++ {
		rng.Shuffle(n, func(i, j int) { perm[i], perm[j] = perm[j], perm[i] })
		mask := 0
		prev := v(mask)
		for _, idx := range perm {
			next := mask | (1 << uint(idx))
			cur := v(next)
			phi[idx] += cur - prev
			mask, prev = next, cur
		}
	}
	for i := range phi {
		phi[i] /= float64(samples)
	}
	return phi
}
