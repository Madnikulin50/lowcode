// Package riskengine — this file implements Evaluate, the "Evidence ->
// Evaluate -> Assessment" step of the pipeline. Explain (attribution) lives
// in attribution.go and consumes an *already produced* RiskAssessment,
// never raw Evidence directly (see attribution.go's doc comment).
package riskengine

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Band is one step of a score -> Severity lookup, mirroring rulesgo's
// risk.band executor (server/pkg/rulesgo/score.go).
type Band struct {
	Level Severity `json:"level"`
	Max   float64  `json:"max"` // inclusive upper bound
}

// DefaultBands is used whenever EvalOptions.Bands is empty.
func DefaultBands() []Band {
	return []Band{
		{Level: SeverityLow, Max: 25},
		{Level: SeverityMedium, Max: 50},
		{Level: SeverityHigh, Max: 75},
		{Level: SeverityCritical, Max: 100},
	}
}

// ApplyBand computes the residual score (inherent score reduced by control
// effectiveness) and looks up its Severity band.
func ApplyBand(inherentScore, controlEffectiveness float64, bands []Band) (residual float64, level Severity) {
	if len(bands) == 0 {
		bands = DefaultBands()
	}
	residual = round2(inherentScore * (1 - clampFloat(controlEffectiveness, 0, 1)))
	level = bands[len(bands)-1].Level
	for _, b := range bands {
		if residual <= b.Max {
			level = b.Level
			break
		}
	}
	return residual, level
}

// EvalOptions carries the cross-cutting inputs Evaluate needs beyond the
// model graph and evidence itself.
type EvalOptions struct {
	// ControlFactorHandle, when set, reads a 0..1 control-effectiveness value
	// out of Evidence (e.g. "controlEffectiveness") the same way any other
	// factor is read. Zero value (no controls modeled) if empty or absent.
	ControlFactorHandle string
	// Bands overrides the model's own calibration for this one call — an
	// escape hatch (testing, previewing a re-calibration before publishing
	// it). Leave empty to use RiskModel.Bands, which is the normal path:
	// precedence is EvalOptions.Bands > RiskModel.Bands > DefaultBands().
	Bands []Band
	Now   time.Time // defaults to time.Now() when zero, override in tests
}

// Evaluate runs model.Strategy over evidence and returns a RiskAssessment
// with InherentScore/ResidualScore/Level/Observations populated. It does
// not populate Attribution or Explanation — call Explain separately.
func Evaluate(model *RiskModel, factors map[uint64]*RiskFactorDef, evidence Evidence) (*RiskAssessment, error) {
	return EvaluateWithOptions(model, factors, evidence, EvalOptions{})
}

// EvaluateWithOptions is Evaluate with control-effectiveness/band overrides.
func EvaluateWithOptions(model *RiskModel, factors map[uint64]*RiskFactorDef, evidence Evidence, opts EvalOptions) (*RiskAssessment, error) {
	score, obs, err := scoreStrategy(model.Strategy, model.Config, model.Nodes, model.Edges, factors, evidence)
	if err != nil {
		return nil, err
	}

	control := 0.0
	if opts.ControlFactorHandle != "" {
		if raw, ok := evidence[opts.ControlFactorHandle]; ok {
			control = clampFloat(toFloat(raw), 0, 1)
		}
	}
	bands := opts.Bands
	if len(bands) == 0 {
		bands = model.Bands
	}
	residual, level := ApplyBand(score, control, bands)

	now := opts.Now
	if now.IsZero() {
		now = time.Now()
	}

	return &RiskAssessment{
		ModelID:              model.ID,
		ModelVersion:         model.Version,
		InherentScore:        round2(score),
		ControlEffectiveness: control,
		ResidualScore:        residual,
		Level:                level,
		Observations:         obs,
		AssessedAt:           now,
	}, nil
}

// scoreStrategy dispatches on strategy and is the single place both
// Evaluate and the "blend" strategy (which re-runs it once per component
// over the same Nodes/Edges) call into.
func scoreStrategy(strategy Strategy, config json.RawMessage, nodes []RiskNode, edges []RiskEdge, factors map[uint64]*RiskFactorDef, evidence Evidence) (float64, []RiskFactorObservation, error) {
	switch strategy {
	case StrategyWeighted:
		return evalWeighted(config, nodes, edges, factors, evidence)
	case StrategyMatrix:
		return evalMatrix(config, nodes, factors, evidence)
	case StrategyBayes:
		return evalBayes(config, nodes, edges, factors, evidence)
	case StrategyBlend:
		return evalBlend(config, nodes, edges, factors, evidence)
	default:
		return 0, nil, fmt.Errorf("riskengine: unknown strategy %q", strategy)
	}
}

// ---------------------------------------------------------------------------
// weighted
// ---------------------------------------------------------------------------

func findOutputNode(nodes []RiskNode) (*RiskNode, error) {
	var out *RiskNode
	for i := range nodes {
		if nodes[i].Role == RoleOutput {
			if out != nil {
				return nil, fmt.Errorf("riskengine: model has more than one output node")
			}
			out = &nodes[i]
		}
	}
	if out == nil {
		return nil, fmt.Errorf("riskengine: model has no output node")
	}
	return out, nil
}

func nodeByID(nodes []RiskNode, id string) (*RiskNode, bool) {
	for i := range nodes {
		if nodes[i].ID == id {
			return &nodes[i], true
		}
	}
	return nil, false
}

// evidenceRaw resolves a node's raw evidence value via its factor's Handle.
func evidenceRaw(node *RiskNode, factors map[uint64]*RiskFactorDef, evidence Evidence) (*RiskFactorDef, interface{}, bool) {
	fd, ok := factors[node.FactorID]
	if !ok {
		return nil, nil, false
	}
	v, ok := evidence[fd.Handle]
	return fd, v, ok
}

// normalize maps a factor's raw evidence value to 0..1, honoring Max/Invert.
func normalize(fd *RiskFactorDef, raw interface{}) (float64, error) {
	val := toFloat(raw)
	switch fd.Scale {
	case ScaleBool:
		if val > 0 {
			val = 1
		} else {
			val = 0
		}
	case ScaleNumeric:
		if fd.Max <= 0 {
			return 0, fmt.Errorf("riskengine: factor %q: max must be > 0", fd.Handle)
		}
		val = clampFloat(val/fd.Max, 0, 1)
	default:
		return 0, fmt.Errorf("riskengine: factor %q: scale %q is not usable by the weighted strategy", fd.Handle, fd.Scale)
	}
	if fd.Invert {
		val = 1 - val
	}
	return val, nil
}

// baselineNormalized is the same 0..1 normalization applied to
// fd.Baseline instead of an observed value — used by Explain to build the
// "factor absent" reference point.
func baselineNormalized(fd *RiskFactorDef) float64 {
	if fd.Baseline == "" {
		if fd.Invert {
			return 1
		}
		return 0
	}
	v, err := strconv.ParseFloat(strings.TrimSpace(fd.Baseline), 64)
	if err != nil {
		return 0
	}
	n, err := normalize(fd, v)
	if err != nil {
		return 0
	}
	return n
}

func evalWeighted(config json.RawMessage, nodes []RiskNode, edges []RiskEdge, factors map[uint64]*RiskFactorDef, evidence Evidence) (float64, []RiskFactorObservation, error) {
	var cfg WeightedConfig
	if len(config) > 0 {
		if err := json.Unmarshal(config, &cfg); err != nil {
			return 0, nil, fmt.Errorf("riskengine: invalid weighted config: %w", err)
		}
	}
	normalizeOut := cfg.Normalize == nil || *cfg.Normalize
	scaleMax := cfg.ScaleMax
	if scaleMax <= 0 {
		scaleMax = 100
	}

	out, err := findOutputNode(nodes)
	if err != nil {
		return 0, nil, err
	}

	var raw, weightSum float64
	var obs []RiskFactorObservation
	for _, e := range edges {
		if e.To != out.ID || e.Weight == 0 {
			continue
		}
		from, ok := nodeByID(nodes, e.From)
		if !ok {
			return 0, nil, fmt.Errorf("riskengine: edge references unknown node %q", e.From)
		}
		fd, rawVal, ok := evidenceRaw(from, factors, evidence)
		if !ok {
			return 0, nil, fmt.Errorf("riskengine: missing evidence for factor referenced by node %q", from.ID)
		}
		norm, err := normalize(fd, rawVal)
		if err != nil {
			return 0, nil, err
		}
		raw += e.Weight * norm
		weightSum += e.Weight
		obs = append(obs, RiskFactorObservation{
			FactorHandle: fd.Handle,
			NodeID:       from.ID,
			Raw:          rawVal,
			Normalized:   round2(norm),
		})
	}
	if weightSum <= 0 {
		return 0, nil, fmt.Errorf("riskengine: weighted strategy: total weight must be > 0")
	}

	score := raw
	if normalizeOut {
		score = (raw / weightSum) * scaleMax
	}
	return score, obs, nil
}

// ---------------------------------------------------------------------------
// matrix
// ---------------------------------------------------------------------------

func evalMatrix(config json.RawMessage, nodes []RiskNode, factors map[uint64]*RiskFactorDef, evidence Evidence) (float64, []RiskFactorObservation, error) {
	var cfg MatrixConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return 0, nil, fmt.Errorf("riskengine: invalid matrix config: %w", err)
	}
	scale := cfg.ScaleSize
	if scale <= 0 {
		scale = 5
	}

	lNode, ok := nodeByID(nodes, cfg.LikelihoodNode)
	if !ok {
		return 0, nil, fmt.Errorf("riskengine: matrix config: unknown likelihoodNode %q", cfg.LikelihoodNode)
	}
	iNode, ok := nodeByID(nodes, cfg.ImpactNode)
	if !ok {
		return 0, nil, fmt.Errorf("riskengine: matrix config: unknown impactNode %q", cfg.ImpactNode)
	}
	lFD, lRaw, ok := evidenceRaw(lNode, factors, evidence)
	if !ok {
		return 0, nil, fmt.Errorf("riskengine: missing evidence for likelihood node %q", lNode.ID)
	}
	iFD, iRaw, ok := evidenceRaw(iNode, factors, evidence)
	if !ok {
		return 0, nil, fmt.Errorf("riskengine: missing evidence for impact node %q", iNode.ID)
	}

	li := clampInt(int(toFloat(lRaw)+0.5), 1, scale)
	ii := clampInt(int(toFloat(iRaw)+0.5), 1, scale)

	score, err := matrixCellScore(li, ii, cfg)
	if err != nil {
		return 0, nil, err
	}

	obs := []RiskFactorObservation{
		{FactorHandle: lFD.Handle, NodeID: lNode.ID, Raw: lRaw, Normalized: float64(li)},
		{FactorHandle: iFD.Handle, NodeID: iNode.ID, Raw: iRaw, Normalized: float64(ii)},
	}
	return score, obs, nil
}

// matrixCellScore is the likelihood x impact -> score formula, shared by
// evalMatrix and attribution.go's masked-baseline value function so both
// always agree on how a cell is computed.
func matrixCellScore(li, ii int, cfg MatrixConfig) (float64, error) {
	switch {
	case len(cfg.Matrix) > 0:
		if li-1 >= len(cfg.Matrix) || ii-1 >= len(cfg.Matrix[li-1]) {
			return 0, fmt.Errorf("riskengine: matrix cell [%d][%d] out of range", li, ii)
		}
		return cfg.Matrix[li-1][ii-1], nil
	case strings.EqualFold(cfg.Formula, "sum"):
		return float64(li + ii), nil
	default:
		return float64(li * ii), nil
	}
}

// ---------------------------------------------------------------------------
// bayes
// ---------------------------------------------------------------------------

// bayesEvidenceAssignment resolves the observed state code for every root
// (evidence) node in net from Evidence, keyed by each node's factor Handle.
func bayesEvidenceAssignment(net *bayesNet, factors map[uint64]*RiskFactorDef, evidence Evidence) (map[string]string, error) {
	assignment := make(map[string]string, len(net.nodes))
	for id, n := range net.nodes {
		if n.cpt != nil {
			continue // not a root — its value is inferred, not observed
		}
		handle := net.handle[id]
		v, ok := evidence[handle]
		if !ok {
			return nil, fmt.Errorf("riskengine: missing evidence for factor %q (node %q)", handle, id)
		}
		code, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("riskengine: evidence for states-factor %q must be a state code string, got %T", handle, v)
		}
		assignment[id] = code
	}
	return assignment, nil
}

func evalBayes(config json.RawMessage, nodes []RiskNode, edges []RiskEdge, factors map[uint64]*RiskFactorDef, evidence Evidence) (float64, []RiskFactorObservation, error) {
	var cfg BayesConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return 0, nil, fmt.Errorf("riskengine: invalid bayes config: %w", err)
	}

	model := &RiskModel{Nodes: nodes, Edges: edges}
	net, err := buildBayesNet(model, factors, nil)
	if err != nil {
		return 0, nil, err
	}
	if _, ok := net.nodes[cfg.OutputNode]; !ok {
		return 0, nil, fmt.Errorf("riskengine: bayes config: unknown outputNode %q", cfg.OutputNode)
	}

	assignment, err := bayesEvidenceAssignment(net, factors, evidence)
	if err != nil {
		return 0, nil, err
	}

	post, err := net.posterior(cfg.OutputNode, assignment)
	if err != nil {
		return 0, nil, err
	}

	var score float64
	for state, p := range post {
		score += p * cfg.ScoreByState[state]
	}

	obs := make([]RiskFactorObservation, 0, len(assignment))
	for nodeID, code := range assignment {
		obs = append(obs, RiskFactorObservation{
			FactorHandle: net.handle[nodeID],
			NodeID:       nodeID,
			Raw:          code,
			StateCode:    code,
		})
	}
	return score, obs, nil
}

// ---------------------------------------------------------------------------
// blend
// ---------------------------------------------------------------------------

func evalBlend(config json.RawMessage, nodes []RiskNode, edges []RiskEdge, factors map[uint64]*RiskFactorDef, evidence Evidence) (float64, []RiskFactorObservation, error) {
	var cfg BlendConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return 0, nil, fmt.Errorf("riskengine: invalid blend config: %w", err)
	}
	if len(cfg.Components) == 0 {
		return 0, nil, fmt.Errorf("riskengine: blend strategy requires at least one component")
	}

	var weightSum, score float64
	var obs []RiskFactorObservation
	seen := make(map[string]bool)
	for _, c := range cfg.Components {
		if c.Strategy == StrategyBlend {
			return 0, nil, fmt.Errorf("riskengine: blend components cannot nest another blend")
		}
		compScore, compObs, err := scoreStrategy(c.Strategy, c.Config, nodes, edges, factors, evidence)
		if err != nil {
			return 0, nil, fmt.Errorf("riskengine: blend component %q: %w", c.Strategy, err)
		}
		score += c.Weight * compScore
		weightSum += c.Weight
		for _, o := range compObs {
			key := o.NodeID + "|" + o.FactorHandle
			if seen[key] {
				continue
			}
			seen[key] = true
			obs = append(obs, o)
		}
	}
	if weightSum <= 0 {
		return 0, nil, fmt.Errorf("riskengine: blend strategy: total component weight must be > 0")
	}
	return score / weightSum, obs, nil
}
