package rulesgo

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// --- score.matrix: likelihood × impact (or custom cell lookup) ---

type scoreMatrixConfig struct {
	LikelihoodField string      `json:"likelihoodField"`
	ImpactField     string      `json:"impactField"`
	Likelihood      string      `json:"likelihood,omitempty"` // literal / {{template}}
	Impact          string      `json:"impact,omitempty"`
	Scale           int         `json:"scale,omitempty"`   // clamp 1..scale, default 5
	Formula         string      `json:"formula,omitempty"` // product (default) | sum
	Matrix          [][]float64 `json:"matrix,omitempty"`  // optional custom cells [L-1][I-1]
	OutScore        string      `json:"outScore,omitempty"`
	OutX            string      `json:"outX,omitempty"`
	OutY            string      `json:"outY,omitempty"`
}

type scoreMatrixExecutor struct{}

func (n *scoreMatrixExecutor) Execute(ctx context.Context, node ChainNode, ec *ExecutionContext) (map[string]interface{}, error) {
	cfg, err := ParseNodeConfig[scoreMatrixConfig](node.Config)
	if err != nil {
		return nil, err
	}

	scale := cfg.Scale
	if scale <= 0 {
		scale = 5
	}

	lx := resolveNumeric(cfg.Likelihood, firstNonEmpty(cfg.LikelihoodField, "likelihood"), ec)
	iy := resolveNumeric(cfg.Impact, firstNonEmpty(cfg.ImpactField, "impact"), ec)
	xi := clampInt(int(math.Round(lx)), 1, scale)
	yi := clampInt(int(math.Round(iy)), 1, scale)

	var score float64
	if len(cfg.Matrix) > 0 {
		if xi-1 >= len(cfg.Matrix) || yi-1 >= len(cfg.Matrix[xi-1]) {
			return nil, fmt.Errorf("matrix cell [%d][%d] out of range", xi, yi)
		}
		score = cfg.Matrix[xi-1][yi-1]
	} else if strings.EqualFold(cfg.Formula, "sum") {
		score = float64(xi + yi)
	} else {
		score = float64(xi * yi)
	}

	outScore := firstNonEmpty(cfg.OutScore, "score")
	outX := firstNonEmpty(cfg.OutX, "likelihood")
	outY := firstNonEmpty(cfg.OutY, "impact")

	ec.Set(outScore, score)
	ec.Set(outX, float64(xi))
	ec.Set(outY, float64(yi))
	if outScore == "score" {
		ec.Set("inherentScore", score)
	}

	return map[string]interface{}{
		"score":      score,
		"likelihood": xi,
		"impact":     yi,
		"formula":    firstNonEmpty(cfg.Formula, "product"),
	}, nil
}

// --- score.weighted: Σ wᵢ · normalize(xᵢ) ---

type scoreFactor struct {
	Field  string  `json:"field"`
	Weight float64 `json:"weight"`
	Max    float64 `json:"max"`              // normalize divisor; required > 0
	Invert bool    `json:"invert,omitempty"` // use (1 - x/max)
}

type scoreWeightedConfig struct {
	Factors   []scoreFactor `json:"factors"`
	Normalize *bool         `json:"normalize,omitempty"` // default true → 0..100
	OutScore  string        `json:"outScore,omitempty"`
	ScaleMax  float64       `json:"scaleMax,omitempty"` // default 100 when normalize
}

type scoreWeightedExecutor struct{}

func (n *scoreWeightedExecutor) Execute(ctx context.Context, node ChainNode, ec *ExecutionContext) (map[string]interface{}, error) {
	cfg, err := ParseNodeConfig[scoreWeightedConfig](node.Config)
	if err != nil {
		return nil, err
	}
	if len(cfg.Factors) == 0 {
		return nil, fmt.Errorf("score.weighted requires at least one factor")
	}

	normalize := true
	if cfg.Normalize != nil {
		normalize = *cfg.Normalize
	}
	scaleMax := cfg.ScaleMax
	if scaleMax <= 0 {
		scaleMax = 100
	}

	parts := make([]map[string]interface{}, 0, len(cfg.Factors))
	var weightSum, raw float64
	for _, f := range cfg.Factors {
		if f.Field == "" {
			return nil, fmt.Errorf("factor field is required")
		}
		if f.Max <= 0 {
			return nil, fmt.Errorf("factor %s: max must be > 0", f.Field)
		}
		val := resolveNumeric("", f.Field, ec)
		norm := clampFloat(val/f.Max, 0, 1)
		if f.Invert {
			norm = 1 - norm
		}
		contrib := f.Weight * norm
		raw += contrib
		weightSum += f.Weight
		parts = append(parts, map[string]interface{}{
			"field":   f.Field,
			"value":   val,
			"weight":  f.Weight,
			"norm":    round2(norm),
			"contrib": round2(contrib),
		})
	}

	score := raw
	if normalize {
		if weightSum <= 0 {
			return nil, fmt.Errorf("total weight must be > 0")
		}
		score = (raw / weightSum) * scaleMax
	}
	score = round2(score)

	outScore := firstNonEmpty(cfg.OutScore, "score")
	ec.Set(outScore, score)
	ec.Set("inherentScore", score)

	return map[string]interface{}{
		"score":   score,
		"raw":     round2(raw),
		"factors": parts,
	}, nil
}

// --- risk.band: score → level + optional residual ---

type riskBandDef struct {
	Name string  `json:"name"`
	Max  float64 `json:"max"` // inclusive upper bound
}

type riskBandConfig struct {
	ScoreField      string        `json:"scoreField,omitempty"`
	ControlField    string        `json:"controlField,omitempty"` // 0..1 effectiveness
	Bands           []riskBandDef `json:"bands,omitempty"`
	OutLevel        string        `json:"outLevel,omitempty"`
	OutResidual     string        `json:"outResidual,omitempty"`
	CriticalLevels  []string      `json:"criticalLevels,omitempty"` // default ["critical"]
	OutCriticalFlag string        `json:"outCriticalFlag,omitempty"`
}

type riskBandExecutor struct{}

func (n *riskBandExecutor) Execute(ctx context.Context, node ChainNode, ec *ExecutionContext) (map[string]interface{}, error) {
	cfg, err := ParseNodeConfig[riskBandConfig](node.Config)
	if err != nil {
		return nil, err
	}

	scoreField := firstNonEmpty(cfg.ScoreField, "score")
	score := resolveNumeric("", scoreField, ec)

	control := 0.0
	if cfg.ControlField != "" {
		control = clampFloat(resolveNumeric("", cfg.ControlField, ec), 0, 1)
	}
	residual := round2(score * (1 - control))

	bands := cfg.Bands
	if len(bands) == 0 {
		bands = []riskBandDef{
			{Name: "low", Max: 25},
			{Name: "medium", Max: 50},
			{Name: "high", Max: 75},
			{Name: "critical", Max: 100},
		}
	}

	level := bands[len(bands)-1].Name
	for _, b := range bands {
		if residual <= b.Max {
			level = b.Name
			break
		}
	}

	criticalNames := cfg.CriticalLevels
	if len(criticalNames) == 0 {
		criticalNames = []string{"critical"}
	}
	isCritical := ""
	for _, n := range criticalNames {
		if strings.EqualFold(level, n) {
			isCritical = "true"
			break
		}
	}

	outLevel := firstNonEmpty(cfg.OutLevel, "level")
	outResidual := firstNonEmpty(cfg.OutResidual, "residualScore")
	outFlag := firstNonEmpty(cfg.OutCriticalFlag, "is_critical")

	ec.Set(outLevel, level)
	ec.Set(outResidual, residual)
	// Empty string so conditional edges (GetString != "") skip non-critical paths.
	ec.Set(outFlag, isCritical)
	ec.Set(node.ID+"_result", isCritical)

	return map[string]interface{}{
		"score":         score,
		"control":       control,
		"residualScore": residual,
		"level":         level,
		"is_critical":   isCritical == "true",
	}, nil
}

// --- helpers ---

func resolveNumeric(literal, field string, ec *ExecutionContext) float64 {
	if literal != "" {
		s := resolveTemplateValue(literal, ec)
		if f, err := strconv.ParseFloat(strings.TrimSpace(s), 64); err == nil {
			return f
		}
	}
	if field == "" {
		return 0
	}
	v := ec.Get(field)
	return toFloat(v)
}

func toFloat(v interface{}) float64 {
	switch t := v.(type) {
	case nil:
		return 0
	case float64:
		return t
	case float32:
		return float64(t)
	case int:
		return float64(t)
	case int64:
		return float64(t)
	case json.Number:
		f, _ := t.Float64()
		return f
	case string:
		f, _ := strconv.ParseFloat(strings.TrimSpace(t), 64)
		return f
	default:
		f, _ := strconv.ParseFloat(fmt.Sprintf("%v", t), 64)
		return f
	}
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func clampFloat(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}
