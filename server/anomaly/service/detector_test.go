package service

import (
	"fmt"
	"math"
	"testing"

	"github.com/madnikulin50/lowcode/server/anomaly/types"
	"github.com/stretchr/testify/require"
)

// --- test helpers ------------------------------------------------------

func feed(d Detector, values []float64) (types.BaselineState, uint64) {
	state := types.BaselineState{}
	var count uint64
	for _, v := range values {
		state = d.UpdateBaseline(fmt.Sprintf("%v", v), state, count)
		count++
	}
	return state, count
}

func detectValue(d Detector, value float64, state types.BaselineState, count uint64, rule *types.Rule) Verdict {
	return d.Detect(fmt.Sprintf("%v", value), state, count, rule)
}

func alternate(a, b float64, n int) []float64 {
	out := make([]float64, n)
	for i := range out {
		if i%2 == 0 {
			out[i] = a
		} else {
			out[i] = b
		}
	}
	return out
}

func repeat(v float64, n int) []float64 {
	out := make([]float64, n)
	for i := range out {
		out[i] = v
	}
	return out
}

func repeatStr(v string, n int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = v
	}
	return out
}

// --- Z-score -------------------------------------------------------------

func TestZScoreDetectorGatesOnMinBaselineCount(t *testing.T) {
	req := require.New(t)
	rule := &types.Rule{Threshold: 3}

	state, count := feed(ZScoreDetector{}, repeat(10, minBaselineCount-1))
	v := detectValue(ZScoreDetector{}, 1000, state, count, rule)
	req.False(v.Anomalous, "must not flag anomalies before the baseline has minBaselineCount samples")
}

func TestZScoreDetectorFlagsOutlier(t *testing.T) {
	req := require.New(t)
	rule := &types.Rule{Threshold: 3}

	state, count := feed(ZScoreDetector{}, alternate(9.9, 10.1, 100))

	req.False(detectValue(ZScoreDetector{}, 10.0, state, count, rule).Anomalous)

	outlier := detectValue(ZScoreDetector{}, 1000, state, count, rule)
	req.True(outlier.Anomalous)
	req.Equal(types.SeverityHigh, outlier.Severity)
}

func TestZScoreDetectorZeroStdDev(t *testing.T) {
	req := require.New(t)
	rule := &types.Rule{Threshold: 3}

	state, count := feed(ZScoreDetector{}, repeat(42, minBaselineCount))
	v := detectValue(ZScoreDetector{}, 9999, state, count, rule)
	req.False(v.Anomalous, "a constant field has zero stddev - nothing is scoreable, so nothing should be flagged")
}

// --- EWMA: the whole point is recovering after a single outlier ------------

func TestEWMARecoversVarianceFasterThanZScore(t *testing.T) {
	req := require.New(t)

	normal := alternate(9.9, 10.1, minBaselineCount+5)
	recovery := alternate(9.9, 10.1, 60)

	ewmaState, ewmaCount := feed(EWMADetector{}, normal)
	ewmaState = EWMADetector{}.UpdateBaseline("1000", ewmaState, ewmaCount)
	ewmaCount++
	for _, v := range recovery {
		ewmaState = EWMADetector{}.UpdateBaseline(fmt.Sprintf("%v", v), ewmaState, ewmaCount)
		ewmaCount++
	}

	zState, zCount := feed(ZScoreDetector{}, normal)
	zState = ZScoreDetector{}.UpdateBaseline("1000", zState, zCount)
	zCount++
	for _, v := range recovery {
		zState = ZScoreDetector{}.UpdateBaseline(fmt.Sprintf("%v", v), zState, zCount)
		zCount++
	}

	ewmaStdDev := math.Sqrt(stateFloat(ewmaState, "var"))
	_, zVariance := welfordStats(zState, zCount)
	zStdDev := math.Sqrt(zVariance)

	req.Less(ewmaStdDev, zStdDev,
		"EWMA's decayed variance should recover from a single outlier much faster than Welford's cumulative variance, which never forgets it")
}

// --- CUSUM: catches sustained small drift, not just point outliers --------

func TestCUSUMDetectorCatchesSustainedSmallDrift(t *testing.T) {
	req := require.New(t)
	rule := &types.Rule{Threshold: 5}
	d := CUSUMDetector{}

	// Long, stable warm-up: mean ~10, stddev ~0.1 (alternating 9.9/10.1) -
	// large enough that the next 20 additions barely move it.
	state, count := feed(d, alternate(9.9, 10.1, 300))

	// Each of these sits about 1 stddev above the (near-stable) mean -
	// individually far below any reasonable z-score threshold, but
	// sustained across many steps.
	var last Verdict
	for i := 0; i < 20; i++ {
		raw := "10.1"
		last = d.Detect(raw, state, count, rule)
		state = d.UpdateBaseline(raw, state, count)
		count++
	}

	req.True(last.Anomalous, "CUSUM should flag sustained one-sided deviation even though no single step trips a z-score threshold of 3")

	zVerdict := ZScoreDetector{}.Detect("10.1", state, count, &types.Rule{Threshold: 3})
	req.False(zVerdict.Anomalous, "sanity check: the same final value alone should not trip a z-score detector")
}

func TestCUSUMDetectorGatesOnMinBaselineCount(t *testing.T) {
	req := require.New(t)
	rule := &types.Rule{Threshold: 5}
	state, count := feed(CUSUMDetector{}, repeat(10, minBaselineCount-1))
	v := detectValue(CUSUMDetector{}, 1000, state, count, rule)
	req.False(v.Anomalous)
}

// --- Range: no baseline/statistics at all ---------------------------------

func TestRangeDetector(t *testing.T) {
	req := require.New(t)
	rule := &types.Rule{Params: types.FindingExplanation{"min": 0.0, "max": 100.0}}
	d := RangeDetector{}

	req.False(d.Detect("50", nil, 0, rule).Anomalous)
	req.True(d.Detect("-5", nil, 0, rule).Anomalous)
	req.True(d.Detect("150", nil, 0, rule).Anomalous)
}

func TestRangeDetectorOpenEnded(t *testing.T) {
	req := require.New(t)
	rule := &types.Rule{Params: types.FindingExplanation{"max": 100.0}}
	d := RangeDetector{}

	req.False(d.Detect("-1000000", nil, 0, rule).Anomalous, "no min configured - only the max side fires")
	req.True(d.Detect("150", nil, 0, rule).Anomalous)
}

// --- MAD: robust to a single outlier already inside the window -----------

func TestMADDetectorRobustToSingleOutlierInWindow(t *testing.T) {
	req := require.New(t)
	rule := &types.Rule{Threshold: 3.5}
	d := MADDetector{}

	values := alternate(9.9, 10.1, minBaselineCount+9)
	values = append(values, 1000) // one contaminating point already in the window
	state, count := feed(d, values)

	// A moderate deviation (tiny next to 1000, but well outside 9.9-10.1)
	// should still be flagged - median/MAD barely move for one outlier in
	// a window this size, unlike mean/stddev.
	v := detectValue(d, 15, state, count, rule)
	req.True(v.Anomalous)
}

func TestMADDetectorGatesOnMinBaselineCount(t *testing.T) {
	req := require.New(t)
	rule := &types.Rule{Threshold: 3.5}
	state, count := feed(MADDetector{}, alternate(9.9, 10.1, minBaselineCount-1))
	v := detectValue(MADDetector{}, 1000, state, count, rule)
	req.False(v.Anomalous)
}

// --- Categorical detectors -------------------------------------------------

func TestRareCategoryDetector(t *testing.T) {
	req := require.New(t)
	rule := &types.Rule{Threshold: 0.05}
	d := RareCategoryDetector{}

	state := types.BaselineState{}
	var count uint64
	values := append(repeatStr("common", 99), "rare")
	for _, v := range values {
		state = d.UpdateBaseline(v, state, count)
		count++
	}

	req.False(d.Detect("common", state, count, rule).Anomalous)
	req.True(d.Detect("rare", state, count, rule).Anomalous)

	unseen := d.Detect("never-seen", state, count, rule)
	req.True(unseen.Anomalous)
	req.Equal(types.SeverityHigh, unseen.Severity)
}

func TestRareCategoryDetectorGatesOnMinCount(t *testing.T) {
	req := require.New(t)
	rule := &types.Rule{Threshold: 0.05}
	d := RareCategoryDetector{}

	state := types.BaselineState{}
	var count uint64
	for i := 0; i < minCategoricalCount-1; i++ {
		state = d.UpdateBaseline("common", state, count)
		count++
	}

	req.False(d.Detect("rare", state, count, rule).Anomalous)
}

func TestNewCategoryDetector(t *testing.T) {
	req := require.New(t)
	d := NewCategoryDetector{}

	state := types.BaselineState{}
	var count uint64
	for i := 0; i < minCategoricalCount+5; i++ {
		state = d.UpdateBaseline("known", state, count)
		count++
	}

	req.False(d.Detect("known", state, count, nil).Anomalous)

	v := d.Detect("brand-new", state, count, nil)
	req.True(v.Anomalous)
	req.Equal(types.SeverityMedium, v.Severity)
}
