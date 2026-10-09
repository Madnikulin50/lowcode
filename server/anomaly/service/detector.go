package service

import (
	"math"
	"sort"

	"github.com/madnikulin50/lowcode/server/anomaly/types"
	"github.com/spf13/cast"
)

type (
	// Verdict is a Detector's finding for a single value.
	Verdict struct {
		Anomalous bool
		Score     float64
		Severity  string
	}

	// Detector decides whether a record's raw field value is anomalous, and
	// folds it into a JSON-opaque running state. Each implementation owns
	// its own state shape (mean/variance, a value window, a frequency map -
	// see the individual detectors below); the scanner only tracks how many
	// observations (count) have been folded in, generically, for the
	// warm-up gates.
	Detector interface {
		// Detect reads state as it stood BEFORE this value is folded in, so
		// checking a value never compares it against a baseline that
		// already includes it.
		Detect(raw string, state types.BaselineState, count uint64, rule *types.Rule) Verdict

		// UpdateBaseline folds raw into state and returns the new state to
		// persist. Detectors with no state (RangeDetector) return state
		// unchanged.
		UpdateBaseline(raw string, state types.BaselineState, count uint64) types.BaselineState
	}
)

// detectors is the scanner's detector registry, keyed by types.Rule.Detector.
var detectors = map[string]Detector{
	types.DetectorZScore:       ZScoreDetector{},
	types.DetectorEWMA:         EWMADetector{},
	types.DetectorCUSUM:        CUSUMDetector{},
	types.DetectorRange:        RangeDetector{},
	types.DetectorMAD:          MADDetector{},
	types.DetectorRareCategory: RareCategoryDetector{},
	types.DetectorNewCategory:  NewCategoryDetector{},
}

// minBaselineCount gates the numeric detectors until a baseline has seen
// enough values for its statistics to be meaningful - otherwise the first
// handful of records in a freshly-watched field would all read as
// "anomalous" simply because the baseline hasn't settled yet.
const minBaselineCount = 30

// minCategoricalCount is the equivalent gate for the categorical detectors:
// smaller, because "this value is rare" or "never seen before" is
// meaningful with a lot less history than a numeric mean/stddev needs.
const minCategoricalCount = 10

// parseNumeric parses a record's raw field value for the numeric detectors;
// non-numeric values (wrong field, blank, free text) are skipped rather
// than treated as an error.
func parseNumeric(raw string) (float64, bool) {
	v, err := cast.ToFloat64E(raw)
	if err != nil {
		return 0, false
	}
	return v, true
}

func stateFloat(state types.BaselineState, key string) float64 {
	v, _ := cast.ToFloat64E(state[key])
	return v
}

func setFloat(state types.BaselineState, key string, value float64) types.BaselineState {
	if state == nil {
		state = types.BaselineState{}
	}
	state[key] = value
	return state
}

// zVerdict is the shared "how far past threshold, in standard deviations"
// scoring used by every detector that reduces to a z-like statistic
// (ZScoreDetector, EWMADetector, MADDetector's modified z-score).
func zVerdict(z, threshold float64) Verdict {
	if z < threshold {
		return Verdict{Score: z}
	}

	severity := types.SeverityLow
	switch {
	case z >= threshold*2:
		severity = types.SeverityHigh
	case z >= threshold*1.5:
		severity = types.SeverityMedium
	}

	return Verdict{Anomalous: true, Score: z, Severity: severity}
}

// welfordUpdate folds value into a Welford online mean/M2 accumulator
// (state keys "mean"/"m2"). Shared by ZScoreDetector and CUSUMDetector,
// which needs a stable, never-decaying reference mean/stddev to measure
// drift against - an EWMA reference would chase the drift instead of
// accumulating deviation from it.
func welfordUpdate(state types.BaselineState, count uint64, value float64) types.BaselineState {
	mean := stateFloat(state, "mean")
	m2 := stateFloat(state, "m2")

	n := float64(count + 1)
	delta := value - mean
	mean += delta / n
	delta2 := value - mean
	m2 += delta * delta2

	state = setFloat(state, "mean", mean)
	state = setFloat(state, "m2", m2)
	return state
}

// welfordStats returns the current mean and sample variance from a Welford
// state (see welfordUpdate), for count prior observations.
func welfordStats(state types.BaselineState, count uint64) (mean, variance float64) {
	mean = stateFloat(state, "mean")
	if count < 2 {
		return mean, 0
	}
	return mean, stateFloat(state, "m2") / float64(count-1)
}

// --- Z-score -----------------------------------------------------------

// ZScoreDetector flags a value as anomalous when its distance from the
// baseline mean, in standard deviations, meets the rule's threshold. The
// mean/variance never forget - a single outlier permanently pulls the
// baseline towards it. EWMADetector and MADDetector exist specifically to
// not have this property.
type ZScoreDetector struct{}

func (ZScoreDetector) Detect(raw string, state types.BaselineState, count uint64, rule *types.Rule) Verdict {
	value, ok := parseNumeric(raw)
	if !ok || count < minBaselineCount {
		return Verdict{}
	}

	mean, variance := welfordStats(state, count)
	stddev := math.Sqrt(variance)
	if stddev == 0 {
		// A perfectly constant field: any deviation is notable, but
		// without a stddev to divide by we can't score it.
		return Verdict{}
	}

	return zVerdict(math.Abs(value-mean)/stddev, rule.Threshold)
}

func (ZScoreDetector) UpdateBaseline(raw string, state types.BaselineState, count uint64) types.BaselineState {
	value, ok := parseNumeric(raw)
	if !ok {
		return state
	}
	return welfordUpdate(state, count, value)
}

// --- EWMA ----------------------------------------------------------------

// ewmaAlpha is the decay applied on every update - fixed for now rather
// than exposed per-rule, to keep the rule-settings UI simple. Higher means
// faster forgetting of old values.
const ewmaAlpha = 0.1

// EWMADetector is ZScoreDetector's shape (distance from a running mean, in
// standard deviations) but with a mean/variance that decays: a past outlier
// stops distorting the baseline once enough later, normal values have
// pushed it back out, instead of permanently pulling the mean towards it.
type EWMADetector struct{}

func (EWMADetector) Detect(raw string, state types.BaselineState, count uint64, rule *types.Rule) Verdict {
	value, ok := parseNumeric(raw)
	if !ok || count < minBaselineCount {
		return Verdict{}
	}

	stddev := math.Sqrt(stateFloat(state, "var"))
	if stddev == 0 {
		return Verdict{}
	}

	return zVerdict(math.Abs(value-stateFloat(state, "mean"))/stddev, rule.Threshold)
}

func (EWMADetector) UpdateBaseline(raw string, state types.BaselineState, count uint64) types.BaselineState {
	value, ok := parseNumeric(raw)
	if !ok {
		return state
	}

	if count == 0 {
		state = setFloat(state, "mean", value)
		return setFloat(state, "var", 0)
	}

	mean := stateFloat(state, "mean")
	variance := stateFloat(state, "var")

	delta := value - mean
	mean += ewmaAlpha * delta
	variance = (1 - ewmaAlpha) * (variance + ewmaAlpha*delta*delta)

	state = setFloat(state, "mean", mean)
	return setFloat(state, "var", variance)
}

// --- CUSUM -----------------------------------------------------------------

// cusumSlack is the "allowance" subtracted before accumulating deviation -
// the standard default is about half a standard deviation. Fixed for now,
// same rationale as ewmaAlpha.
const cusumSlack = 0.5

// CUSUMDetector accumulates standardized deviation from a stable
// (Welford, non-decaying) reference mean and flags sustained drift in one
// direction - catching a value that creeps away from normal in small
// steps, each too small on its own to trip ZScoreDetector/EWMADetector.
type CUSUMDetector struct{}

func (CUSUMDetector) Detect(raw string, state types.BaselineState, count uint64, rule *types.Rule) Verdict {
	value, ok := parseNumeric(raw)
	if !ok || count < minBaselineCount {
		return Verdict{}
	}

	cumPos, cumNeg, ok := cusumAccumulate(state, count, value)
	if !ok {
		return Verdict{}
	}

	score := math.Max(cumPos, -cumNeg)
	if score < rule.Threshold {
		return Verdict{Score: score}
	}

	severity := types.SeverityLow
	switch {
	case score >= rule.Threshold*2:
		severity = types.SeverityHigh
	case score >= rule.Threshold*1.5:
		severity = types.SeverityMedium
	}
	return Verdict{Anomalous: true, Score: score, Severity: severity}
}

func (CUSUMDetector) UpdateBaseline(raw string, state types.BaselineState, count uint64) types.BaselineState {
	value, ok := parseNumeric(raw)
	if !ok {
		return state
	}

	if cumPos, cumNeg, ok := cusumAccumulate(state, count, value); ok {
		state = setFloat(state, "cumPos", cumPos)
		state = setFloat(state, "cumNeg", cumNeg)
	}

	return welfordUpdate(state, count, value)
}

// cusumAccumulate computes what this detector's two one-sided cumulative
// sums become if value is folded in - shared by Detect (to score against
// the rule's threshold) and UpdateBaseline (to persist), so the two always
// agree on the same numbers for the same value.
func cusumAccumulate(state types.BaselineState, count uint64, value float64) (cumPos, cumNeg float64, ok bool) {
	mean, variance := welfordStats(state, count)
	stddev := math.Sqrt(variance)
	if stddev == 0 {
		return 0, 0, false
	}

	z := (value - mean) / stddev

	cumPos = stateFloat(state, "cumPos") + z - cusumSlack
	if cumPos < 0 {
		cumPos = 0
	}

	cumNeg = stateFloat(state, "cumNeg") + z + cusumSlack
	if cumNeg > 0 {
		cumNeg = 0
	}

	return cumPos, cumNeg, true
}

// --- Range -----------------------------------------------------------------

// RangeDetector flags a value outside a fixed [min, max] configured on the
// rule's Params - no baseline or statistics at all, for known business
// constraints (a discount must be within 0-100, an age within 0-120...).
type RangeDetector struct{}

func (RangeDetector) Detect(raw string, _ types.BaselineState, _ uint64, rule *types.Rule) Verdict {
	value, ok := parseNumeric(raw)
	if !ok {
		return Verdict{}
	}

	if min, hasMin := rangeParam(rule, "min"); hasMin && value < min {
		return Verdict{Anomalous: true, Score: min - value, Severity: types.SeverityHigh}
	}
	if max, hasMax := rangeParam(rule, "max"); hasMax && value > max {
		return Verdict{Anomalous: true, Score: value - max, Severity: types.SeverityHigh}
	}
	return Verdict{}
}

func (RangeDetector) UpdateBaseline(_ string, state types.BaselineState, _ uint64) types.BaselineState {
	return state
}

func rangeParam(rule *types.Rule, key string) (float64, bool) {
	if rule.Params == nil {
		return 0, false
	}
	v, ok := rule.Params[key]
	if !ok {
		return 0, false
	}
	f, err := cast.ToFloat64E(v)
	return f, err == nil
}

// --- MAD (median absolute deviation) ----------------------------------------

// madWindowCap bounds how many recent values MADDetector keeps, so the
// baseline row doesn't grow without limit.
const madWindowCap = 200

// MADDetector uses the median and median absolute deviation of a bounded
// recent-value window instead of mean/stddev: unlike ZScoreDetector
// (permanently skewed by one outlier) or EWMADetector (needs time to
// recover), a single outlier barely moves a median at all.
type MADDetector struct{}

func (MADDetector) Detect(raw string, state types.BaselineState, count uint64, rule *types.Rule) Verdict {
	value, ok := parseNumeric(raw)
	if !ok || count < minBaselineCount {
		return Verdict{}
	}

	window := madWindow(state)
	if len(window) == 0 {
		return Verdict{}
	}

	median := medianOf(window)
	mad := madOf(window, median)
	if mad == 0 {
		return Verdict{}
	}

	// 0.6745 makes the modified z-score comparable to a normal-distribution
	// z-score (the standard consistency constant for MAD).
	return zVerdict(0.6745*math.Abs(value-median)/mad, rule.Threshold)
}

func (MADDetector) UpdateBaseline(raw string, state types.BaselineState, _ uint64) types.BaselineState {
	value, ok := parseNumeric(raw)
	if !ok {
		return state
	}

	window := append(madWindow(state), value)
	if len(window) > madWindowCap {
		window = window[len(window)-madWindowCap:]
	}

	if state == nil {
		state = types.BaselineState{}
	}
	state["window"] = window
	return state
}

// madWindow reads state["window"] regardless of whether it came from a
// fresh Detect/UpdateBaseline call in this process (stored as []float64 by
// UpdateBaseline itself) or was just loaded from the store (decoded from
// JSON as []interface{}).
func madWindow(state types.BaselineState) []float64 {
	if fw, ok := state["window"].([]float64); ok {
		return fw
	}

	raw, _ := state["window"].([]interface{})
	out := make([]float64, 0, len(raw))
	for _, v := range raw {
		if f, err := cast.ToFloat64E(v); err == nil {
			out = append(out, f)
		}
	}
	return out
}

func medianOf(values []float64) float64 {
	sorted := append([]float64(nil), values...)
	sort.Float64s(sorted)

	n := len(sorted)
	if n == 0 {
		return 0
	}
	if n%2 == 1 {
		return sorted[n/2]
	}
	return (sorted[n/2-1] + sorted[n/2]) / 2
}

func madOf(values []float64, median float64) float64 {
	deviations := make([]float64, len(values))
	for i, v := range values {
		deviations[i] = math.Abs(v - median)
	}
	return medianOf(deviations)
}

// --- Categorical detectors ---------------------------------------------

// categoryCountsCap bounds how many distinct values a rule's frequency map
// tracks, so a near-unique field (an ID column watched by mistake) can't
// grow the baseline row without limit. Once full, further never-seen
// values still get evaluated as "new"/"rare" but aren't added to the map.
const categoryCountsCap = 500

// RareCategoryDetector flags a value whose observed frequency (count/total)
// is below the rule's threshold (a fraction, e.g. 0.01 for "under 1% of
// records") - for categorical/enum-like fields, not numeric ones.
type RareCategoryDetector struct{}

func (RareCategoryDetector) Detect(raw string, state types.BaselineState, _ uint64, rule *types.Rule) Verdict {
	counts, total := categoryCounts(state)
	if raw == "" || total < minCategoricalCount {
		return Verdict{}
	}

	freq := float64(counts[raw]) / float64(total)
	if freq >= rule.Threshold {
		return Verdict{Score: freq}
	}

	severity := types.SeverityLow
	switch {
	case freq == 0:
		severity = types.SeverityHigh
	case freq < rule.Threshold/2:
		severity = types.SeverityMedium
	}
	return Verdict{Anomalous: true, Score: freq, Severity: severity}
}

func (RareCategoryDetector) UpdateBaseline(raw string, state types.BaselineState, _ uint64) types.BaselineState {
	return updateCategoryCounts(raw, state)
}

// NewCategoryDetector flags a value that has never been observed before.
// Unlike RareCategoryDetector this is a binary signal - there's no
// "somewhat new" - so it doesn't score by frequency.
type NewCategoryDetector struct{}

func (NewCategoryDetector) Detect(raw string, state types.BaselineState, _ uint64, _ *types.Rule) Verdict {
	counts, total := categoryCounts(state)
	if raw == "" || total < minCategoricalCount {
		return Verdict{}
	}

	if _, seen := counts[raw]; seen {
		return Verdict{}
	}
	return Verdict{Anomalous: true, Severity: types.SeverityMedium}
}

func (NewCategoryDetector) UpdateBaseline(raw string, state types.BaselineState, _ uint64) types.BaselineState {
	return updateCategoryCounts(raw, state)
}

// categoryCounts reads the shared {counts, total} shape used by both
// categorical detectors. Baselines are per-rule (see anomaly/types/baseline.go),
// so RareCategoryDetector and NewCategoryDetector each keep their own copy
// even when watching the same field.
func categoryCounts(state types.BaselineState) (map[string]int64, int64) {
	counts := map[string]int64{}
	if raw, ok := state["counts"].(map[string]interface{}); ok {
		for k, v := range raw {
			if n, err := cast.ToInt64E(v); err == nil {
				counts[k] = n
			}
		}
	}
	total, _ := cast.ToInt64E(state["total"])
	return counts, total
}

func updateCategoryCounts(raw string, state types.BaselineState) types.BaselineState {
	if raw == "" {
		return state
	}

	counts, total := categoryCounts(state)
	if _, seen := counts[raw]; seen || len(counts) < categoryCountsCap {
		counts[raw]++
	}
	total++

	if state == nil {
		state = types.BaselineState{}
	}
	countsOut := make(map[string]any, len(counts))
	for k, v := range counts {
		countsOut[k] = v
	}
	state["counts"] = countsOut
	state["total"] = total
	return state
}
