package service

import (
	"fmt"
	"strings"
	"testing"

	"github.com/madnikulin50/lowcode/server/compose/types"
)

func TestWantsSalesDynamics(t *testing.T) {
	if !wantsSalesDynamics("покажи динамику продаж") {
		t.Fatal("expected sales-dynamics intent")
	}
	if !wantsSalesDynamics("Show sales trend chart") {
		t.Fatal("expected english sales trend")
	}
	if wantsSalesDynamics("покажи магазины") {
		t.Fatal("stores list is not a sales chart")
	}
}

func TestReportEntriesToChartCapsAndSorts(t *testing.T) {
	entries := make([]map[string]any, 0, 30)
	for i := 1; i <= 30; i++ {
		entries = append(entries, map[string]any{
			"dimension_0": fmt.Sprintf("2024-%02d", i),
			"revenue":     float64(i * 10),
			"count":       1,
		})
	}
	spec, err := reportEntriesToChart(entries, visualizeArgs{
		metrics:   "SUM(revenue) AS revenue",
		chartType: "line",
		title:     "Динамика продаж",
		yName:     "Выручка",
	})
	if err != nil {
		t.Fatal(err)
	}
	if spec.Type != "line" || spec.Title != "Динамика продаж" {
		t.Fatalf("unexpected spec meta: %+v", spec)
	}
	if len(spec.Labels) != chatChartMaxCategories {
		t.Fatalf("want %d labels, got %d", chatChartMaxCategories, len(spec.Labels))
	}
	if spec.Labels[0] != "2024-07" || spec.Labels[len(spec.Labels)-1] != "2024-30" {
		t.Fatalf("expected last 24 months, got %s .. %s", spec.Labels[0], spec.Labels[len(spec.Labels)-1])
	}
	if len(spec.Series) != 1 || spec.Series[0].Name != "Выручка" {
		t.Fatalf("unexpected series: %+v", spec.Series)
	}
	fence := encodeChartFence(spec)
	if extractChartFence(fence) == "" {
		t.Fatalf("fence not detected: %s", fence)
	}
	if !strings.Contains(fence, `"type":"line"`) {
		t.Fatalf("missing type in fence: %s", fence)
	}
}

func TestExpandChatKeywordsAddsReceipt(t *testing.T) {
	kws := promptKeywords("покажи динамику продаж")
	joined := strings.Join(kws, " ")
	if !strings.Contains(joined, "receipt") || !strings.Contains(joined, "sales") {
		t.Fatalf("expected receipt/sales synonyms, got %v", kws)
	}
}

func TestMergeChartAndComment(t *testing.T) {
	fence := "```chart\n{\"type\":\"line\",\"labels\":[\"a\"],\"series\":[{\"name\":\"x\",\"data\":[1]}]}\n```"
	got := mergeChartAndComment(fence, "Комментарий без графика")
	if !strings.HasPrefix(got, "```chart") || !strings.Contains(got, "Комментарий") {
		t.Fatalf("unexpected merge: %s", got)
	}
	got = mergeChartAndComment(fence, fence)
	if got != extractChartFence(fence) {
		t.Fatalf("duplicate fence should collapse, got %s", got)
	}
}

func TestExtractChartFenceComposeChart(t *testing.T) {
	body := `{"chartID":"496731940685479937","namespaceID":"495727984893558785","title":"По категориям"}`
	fence := "```compose-chart\n" + body + "\n```"
	got := extractChartFence(fence)
	if !strings.Contains(got, "compose-chart") || !strings.Contains(got, "496731940685479937") {
		t.Fatalf("compose-chart fence not detected: %s", got)
	}
	if strings.HasPrefix(got, "```chart\n") && !strings.Contains(got, "compose-chart") {
		t.Fatalf("compose-chart must not be rewritten as ```chart: %s", got)
	}
	regular := "```chart\n{\"type\":\"line\",\"labels\":[\"a\"],\"series\":[{\"name\":\"x\",\"data\":[1]}]}\n```"
	if extractChartFence(regular) == "" {
		t.Fatal("regular chart fence should still extract")
	}

	jsonFence := "```json\n" + body + "\n```"
	got = extractChartFence(jsonFence)
	if !strings.Contains(got, "compose-chart") || !strings.Contains(got, "496731940685479937") {
		t.Fatalf("json fence with chartID should become compose-chart: %s", got)
	}

	spaced := "``` compose-chart\n" + body + "\n```"
	got = extractChartFence(spaced)
	if !strings.Contains(got, "compose-chart") {
		t.Fatalf("spaced language tag should still extract: %s", got)
	}

	comment := "График:\n" + fence + "\nготово"
	got = extractChartFence(comment)
	if !strings.Contains(got, "496731940685479937") {
		t.Fatalf("fence inside commentary should extract: %s", got)
	}
}

func TestMatchPageChartTitle(t *testing.T) {
	refs := []pageChartRef{
		{BlockID: 1, BlockTitle: "По категориям", ChartID: 100, ChartName: "By Category", NamespaceID: 9},
		{BlockID: 2, BlockTitle: "Выручка по брендам", ChartID: 200, ChartName: "Revenue by brand", NamespaceID: 9},
		{BlockID: 3, BlockTitle: "Empty", ChartID: 0, NamespaceID: 9},
	}
	hit := matchPageChart(refs, "по категориям", 0, 0)
	if hit == nil || hit.ChartID != 100 {
		t.Fatalf("expected По категориям, got %+v", hit)
	}
	hit = matchPageChart(refs, "КАТЕГОРИЯМ", 0, 0)
	if hit == nil || hit.ChartID != 100 {
		t.Fatalf("contains match should find По категориям, got %+v", hit)
	}
	hit = matchPageChart(refs, "By Category", 0, 0)
	if hit == nil || hit.ChartID != 100 {
		t.Fatalf("chart.name match, got %+v", hit)
	}
	hit = matchPageChart(refs, "", 0, 200)
	if hit == nil || hit.ChartID != 200 {
		t.Fatalf("exact chartID, got %+v", hit)
	}
	if matchPageChart(refs, "нет такого", 0, 0) != nil {
		t.Fatal("unknown title should miss")
	}
	if matchPageChart(refs, "динамику продаж", 0, 0) != nil {
		t.Fatal("sales-dynamics phrase must not match unrelated page charts")
	}
}

func TestCollectPageChartsSkipsEmptyChartID(t *testing.T) {
	page := &types.Page{
		ID:          1,
		NamespaceID: 9,
		Blocks: types.PageBlocks{
			{BlockID: 1, Kind: "Chart", Title: "По категориям", Options: map[string]any{"chartID": "100"}},
			{BlockID: 2, Kind: "Chart", Title: "Broken", Options: map[string]any{"chartID": ""}},
			{BlockID: 3, Kind: "Chart", Title: "Also broken", Options: map[string]any{}},
			{BlockID: 4, Kind: "Metric", Title: "KPI", Options: map[string]any{"chartID": "999"}},
		},
	}
	refs := collectPageCharts(page, map[uint64]string{100: "By Category"}, nil)
	if len(refs) != 1 || refs[0].ChartID != 100 || refs[0].BlockTitle != "По категориям" {
		t.Fatalf("expected one Chart with chartID, got %+v", refs)
	}
}

func TestCollectPageChartsIncludesTabbed(t *testing.T) {
	page := &types.Page{
		ID:          1,
		NamespaceID: 9,
		Blocks: types.PageBlocks{
			{BlockID: 10, Kind: "Tabs", Title: "Tabs", Options: map[string]any{
				"tabs": []any{map[string]any{"blockID": "20", "title": "t"}},
			}},
			{BlockID: 20, Kind: "Chart", Title: "Hidden tab chart", Options: map[string]any{"chartID": "300"}},
			{BlockID: 30, Kind: "Chart", Title: "Not in layout", Options: map[string]any{"chartID": "400"}},
		},
	}
	visible := map[uint64]bool{10: true}
	refs := collectPageCharts(page, map[uint64]string{300: "TabChart", 400: "Other"}, visible)
	if len(refs) != 1 || refs[0].ChartID != 300 {
		t.Fatalf("expected tabbed chart only, got %+v", refs)
	}
}

func TestExtractShowChartTitle(t *testing.T) {
	got := extractShowChartTitle("покажи график По категориям")
	if !strings.EqualFold(got, "По категориям") {
		t.Fatalf("got %q", got)
	}
	got = extractShowChartTitle("покажи «Выручка по брендам»")
	if !strings.EqualFold(got, "Выручка по брендам") {
		t.Fatalf("got %q", got)
	}
	if extractShowChartTitle("покажи график") != "" {
		t.Fatal("generic 'покажи график' should not extract a title")
	}
	if extractShowChartTitle("покажи динамику продаж") != "динамику продаж" {
		t.Fatalf("sales-dynamics phrase should still extract remainder for page-title matching")
	}
}
