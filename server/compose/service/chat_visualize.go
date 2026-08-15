package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/madnikulin50/lowcode/server/compose/types"
	"github.com/madnikulin50/lowcode/server/pkg/chat"
	"github.com/madnikulin50/lowcode/server/pkg/filter"
)

const chatChartMaxCategories = 24

type chatChartSpec struct {
	Type   string            `json:"type"`
	Title  string            `json:"title,omitempty"`
	YName  string            `json:"yName,omitempty"`
	Labels []string          `json:"labels"`
	Series []chatChartSeries `json:"series"`
}

type chatChartSeries struct {
	Name string    `json:"name"`
	Data []float64 `json:"data"`
}

func chatVisualizeTools() []chat.ToolDef {
	return []chat.ToolDef{
		{
			Name: "sales_dynamics",
			Description: "Plot monthly sales / revenue dynamics (динамика продаж, выручка по месяцам) as a line chart. " +
				"Aggregates receipt_margin_slice (slice_kind=month) or receipt_positions. Returns a ```chart block. " +
				"Call this instead of dumping receipt rows. Never use create_chart for this.",
			Params: []chat.ParamDef{
				{Name: "type", Type: "string", Required: false, Description: "Chart type: line (default) or bar"},
				{Name: "title", Type: "string", Required: false, Description: "Chart title"},
				{Name: "metric", Type: "string", Required: false, Description: "Metric field: revenue (default), gross_profit, cogs, position_sum"},
			},
			Handler: salesDynamics,
		},
		{
			Name: "visualize_report",
			Description: "Aggregate module records into a compact chart (never dumps raw rows). " +
				"Use for график / диаграмма / динамика / сравни / chart / trend when the user is NOT naming a Chart block already on the page. " +
				"If the user names a page chart by heading/title, call show_page_chart instead. " +
				"metrics e.g. SUM(revenue) or SUM(position_sum); dimensions e.g. slice_name or DATE_FORMAT(dt, '%Y-%m-01'). Caps at 24 categories.",
			Params: []chat.ParamDef{
				{Name: "module", Type: "string", Required: true, Description: "Module handle or ID (e.g. Receipt_margin_slice, receipt_positions)"},
				{Name: "metrics", Type: "string", Required: false, Description: "Aggregate metrics, e.g. SUM(revenue) AS revenue. Empty = COUNT."},
				{Name: "dimensions", Type: "string", Required: false, Description: "Group-by field or DATE_FORMAT(dt, '%Y-%m-01')"},
				{Name: "filter", Type: "string", Required: false, Description: "Optional record filter, e.g. slice_kind = 'month'"},
				{Name: "type", Type: "string", Required: false, Description: "line, bar, pie, or doughnut (default bar; use line for trends)"},
				{Name: "title", Type: "string", Required: false, Description: "Chart title"},
			},
			Handler: visualizeReport,
		},
		{
			Name: "show_page_chart",
			Description: "Show an existing Chart page block by its heading/title (e.g. «По категориям», «Выручка по брендам»). " +
				"Use this whenever the user names a chart that is already on the current page. " +
				"Do NOT use sales_dynamics or visualize_report for named page charts. Returns a ```compose-chart block.",
			Params: []chat.ParamDef{
				{Name: "title", Type: "string", Required: false, Description: "Block heading / visible title (e.g. По категориям)"},
				{Name: "blockID", Type: "string", Required: false, Description: "Page block ID"},
				{Name: "chartID", Type: "string", Required: false, Description: "Chart resource ID"},
			},
			Handler: showPageChart,
		},
	}
}

func wantsSalesDynamics(prompt string) bool {
	p := strings.ToLower(prompt)
	hasSales := containsAny(p, "продаж", "выруч", "sales", "revenue")
	hasTrend := containsAny(p, "динамик", "тренд", "trend", "график", "диаграмм", "chart", "по месяц", "помесяч")
	return hasSales && hasTrend
}

func containsAny(s string, needles ...string) bool {
	for _, n := range needles {
		if strings.Contains(s, n) {
			return true
		}
	}
	return false
}

func salesDynamics(ctx context.Context, params map[string]string) string {
	if params == nil {
		params = map[string]string{}
	}
	ns := nsID(ctx, params)
	if ns == 0 {
		return "Missing namespaceID"
	}
	chartType := normalizeChartType(params["type"], "line")
	title := strings.TrimSpace(params["title"])
	if title == "" {
		title = "Динамика продаж"
	}
	metric := strings.TrimSpace(params["metric"])
	if metric == "" {
		metric = "revenue"
	}

	if mod := findModuleFlexible(ctx, ns, "Receipt_margin_slice", "receipt_margin_slice"); mod != nil {
		field := pickNumericField(mod, metric, "revenue", "gross_profit", "position_sum")
		out := runVisualize(ctx, ns, mod, visualizeArgs{
			metrics:    fmt.Sprintf("SUM(%s) AS %s", field, field),
			dimensions: "slice_name",
			filter:     "slice_kind = 'month'",
			chartType:  chartType,
			title:      title,
			yName:      fieldLabel(mod, field, "Выручка"),
		})
		if extractChartFence(out) != "" {
			return out
		}
		if fallback := salesDynamicsFromSliceRecords(ctx, ns, mod, field, chartType, title); extractChartFence(fallback) != "" {
			return fallback
		}
	}

	if mod := findModuleFlexible(ctx, ns, "Receipt_positions", "receipt_positions"); mod != nil {
		field := pickNumericField(mod, metric, "position_sum", "revenue", "quantity")
		dt := pickDateField(mod, "dt")
		if dt != "" && field != "" {
			out := runVisualize(ctx, ns, mod, visualizeArgs{
				metrics:    fmt.Sprintf("SUM(%s) AS %s", field, field),
				dimensions: fmt.Sprintf("DATE_FORMAT(%s, '%%Y-%%m-01')", dt),
				chartType:  chartType,
				title:      title,
				yName:      fieldLabel(mod, field, "Выручка"),
			})
			if extractChartFence(out) != "" {
				return out
			}
		}
	}

	return "Не найдены модули продаж (receipt_margin_slice / receipt_positions) для построения графика."
}

func visualizeReport(ctx context.Context, params map[string]string) string {
	if params == nil {
		params = map[string]string{}
	}
	ns := nsID(ctx, params)
	if ns == 0 {
		return "Missing namespaceID"
	}
	ident := strings.TrimSpace(firstNonEmpty(params["module"], params["handle"], params["moduleID"]))
	if ident == "" {
		return "Укажите module (handle или ID). Для динамики продаж вызовите sales_dynamics."
	}
	mod := findModuleFlexible(ctx, ns, ident)
	if mod == nil {
		return fmt.Sprintf("Модуль '%s' не найден.", ident)
	}

	chartType := normalizeChartType(params["type"], "")
	title := strings.TrimSpace(params["title"])
	if title == "" {
		title = mod.Name
	}
	metrics := strings.TrimSpace(params["metrics"])
	dimensions := strings.TrimSpace(params["dimensions"])
	if metrics == "" || dimensions == "" {
		metrics, dimensions = defaultReportExprs(mod, metrics, dimensions)
	}
	if chartType == "" {
		if strings.Contains(strings.ToLower(dimensions), "date_format") || looksLikeTimeDimension(mod, dimensions) {
			chartType = "line"
		} else {
			chartType = "bar"
		}
	}

	return runVisualize(ctx, ns, mod, visualizeArgs{
		metrics:    metrics,
		dimensions: dimensions,
		filter:     strings.TrimSpace(params["filter"]),
		chartType:  chartType,
		title:      title,
		yName:      strings.TrimSpace(params["yName"]),
	})
}

type visualizeArgs struct {
	metrics    string
	dimensions string
	filter     string
	chartType  string
	title      string
	yName      string
}

func runVisualize(ctx context.Context, ns uint64, mod *types.Module, args visualizeArgs) string {
	if DefaultRecord == nil {
		return "Record service is not available"
	}
	raw, err := DefaultRecord.Report(ctx, ns, mod.ID, args.metrics, args.dimensions, args.filter)
	if err != nil {
		return fmt.Sprintf("Не удалось построить отчёт по модулю '%s': %v", mod.Name, err)
	}
	entries := asReportEntries(raw)
	if len(entries) == 0 {
		return fmt.Sprintf("Нет данных для графика в модуле '%s'.", mod.Name)
	}
	spec, err := reportEntriesToChart(entries, args)
	if err != nil {
		return err.Error()
	}
	return encodeChartFence(spec)
}

func salesDynamicsFromSliceRecords(ctx context.Context, ns uint64, mod *types.Module, metric, chartType, title string) string {
	if DefaultRecord == nil {
		return ""
	}
	flt := types.RecordFilter{
		ModuleID:    mod.ID,
		NamespaceID: ns,
		Query:       "slice_kind = 'month'",
	}
	flt.Limit = uint(chatChartMaxCategories * 2)
	flt.Sort = filter.SortExprSet{{Column: "slice_name"}}
	set, _, err := DefaultRecord.Find(ctx, flt)
	if err != nil || len(set) == 0 {
		return ""
	}
	type row struct {
		label string
		value float64
	}
	rows := make([]row, 0, len(set))
	for _, r := range set {
		label := recordValue(r, "slice_name")
		if label == "" {
			continue
		}
		val, ok := parseFloat(recordValue(r, metric))
		if !ok {
			continue
		}
		rows = append(rows, row{label: label, value: round2(val)})
	}
	if len(rows) == 0 {
		return ""
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].label < rows[j].label })
	if len(rows) > chatChartMaxCategories {
		rows = rows[len(rows)-chatChartMaxCategories:]
	}
	labels := make([]string, len(rows))
	data := make([]float64, len(rows))
	for i, r := range rows {
		labels[i] = r.label
		data[i] = r.value
	}
	yName := fieldLabel(mod, metric, "Выручка")
	return encodeChartFence(&chatChartSpec{
		Type:   chartType,
		Title:  title,
		YName:  yName,
		Labels: labels,
		Series: []chatChartSeries{{Name: yName, Data: data}},
	})
}

func reportEntriesToChart(entries []map[string]any, args visualizeArgs) (*chatChartSpec, error) {
	if len(entries) == 0 {
		return nil, fmt.Errorf("Нет данных для графика")
	}

	metricKeys := collectMetricKeys(entries, args.metrics)
	if len(metricKeys) == 0 {
		return nil, fmt.Errorf("В отчёте нет числовых рядов для графика")
	}

	type point struct {
		label  string
		values []float64
	}
	points := make([]point, 0, len(entries))
	for _, e := range entries {
		label := stringifyChartVal(firstPresent(e, "dimension_0", "dimension_1", "slice_name"))
		if label == "" {
			continue
		}
		vals := make([]float64, len(metricKeys))
		for i, k := range metricKeys {
			vals[i] = round2(toFloat(e[k]))
		}
		points = append(points, point{label: label, values: vals})
	}
	if len(points) == 0 {
		return nil, fmt.Errorf("Не удалось прочитать категории отчёта")
	}

	sort.SliceStable(points, func(i, j int) bool { return points[i].label < points[j].label })
	if len(points) > chatChartMaxCategories {
		points = points[len(points)-chatChartMaxCategories:]
	}

	labels := make([]string, len(points))
	series := make([]chatChartSeries, len(metricKeys))
	for i, k := range metricKeys {
		name := k
		if i == 0 && args.yName != "" {
			name = args.yName
		}
		series[i] = chatChartSeries{Name: name, Data: make([]float64, len(points))}
	}
	for i, p := range points {
		labels[i] = p.label
		for s := range metricKeys {
			series[s].Data[i] = p.values[s]
		}
	}

	return &chatChartSpec{
		Type:   args.chartType,
		Title:  args.title,
		YName:  args.yName,
		Labels: labels,
		Series: series,
	}, nil
}

func encodeChartFence(spec *chatChartSpec) string {
	if spec == nil || len(spec.Series) == 0 || len(spec.Labels) == 0 {
		return ""
	}
	if spec.Type == "" {
		spec.Type = "bar"
	}
	body, err := json.Marshal(spec)
	if err != nil {
		return fmt.Sprintf("Failed to encode chart: %v", err)
	}
	return "```chart\n" + string(body) + "\n```"
}

var chatFenceOpen = regexp.MustCompile("(?i)```+\\s*(compose-chart|chart|echarts|json)\\b")

func extractChartFence(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	lower := strings.ToLower(s)
	loc := chatFenceOpen.FindStringSubmatchIndex(lower)
	if loc != nil {
		lang := lower[loc[2]:loc[3]]
		rest := s[loc[1]:]
		body := rest
		if nl := strings.IndexAny(rest, "\r\n"); nl >= 0 {
			same := strings.TrimSpace(rest[:nl])
			after := rest[nl+1:]
			if strings.HasPrefix(same, "{") {
				body = same
			} else {
				body = after
			}
		}
		if closeIdx := strings.Index(body, "```"); closeIdx >= 0 {
			body = body[:closeIdx]
		}
		body = strings.TrimSpace(body)
		if lang == "compose-chart" || looksLikeComposeChartJSON(body) {
			if looksLikeComposeChartJSON(body) {
				return "```compose-chart\n" + jsonObject(body) + "\n```"
			}
		}
		if looksLikeChartJSON(body) {
			return "```chart\n" + jsonObject(body) + "\n```"
		}
	}
	if obj := jsonObject(s); looksLikeComposeChartJSON(obj) {
		return "```compose-chart\n" + obj + "\n```"
	}
	if obj := jsonObject(s); looksLikeChartJSON(obj) {
		return "```chart\n" + obj + "\n```"
	}
	return ""
}

func jsonObject(s string) string {
	s = strings.TrimSpace(s)
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start < 0 || end <= start {
		return strings.TrimSpace(s)
	}
	return strings.TrimSpace(s[start : end+1])
}

func looksLikeComposeChartJSON(s string) bool {
	s = jsonObject(s)
	if s == "" || s[0] != '{' {
		return false
	}
	var spec struct {
		ChartID any `json:"chartID"`
		Series  any `json:"series"`
	}
	if err := json.Unmarshal([]byte(s), &spec); err != nil {
		start := strings.Index(s, "{")
		end := strings.LastIndex(s, "}")
		if start < 0 || end <= start {
			return false
		}
		if err := json.Unmarshal([]byte(s[start:end+1]), &spec); err != nil {
			return false
		}
	}
	if spec.Series != nil {
		return false
	}
	id := strings.TrimSpace(fmt.Sprint(spec.ChartID))
	return spec.ChartID != nil && id != "" && id != "0" && id != "<nil>"
}

func looksLikeChartJSON(s string) bool {
	s = jsonObject(s)
	if s == "" || s[0] != '{' {
		return false
	}
	var spec chatChartSpec
	if err := json.Unmarshal([]byte(s), &spec); err != nil {
		start := strings.Index(s, "{")
		end := strings.LastIndex(s, "}")
		if start < 0 || end <= start {
			return false
		}
		if err := json.Unmarshal([]byte(s[start:end+1]), &spec); err != nil {
			return false
		}
	}
	return len(spec.Series) > 0 && (len(spec.Labels) > 0 || len(spec.Series[0].Data) > 0)
}

func asReportEntries(raw any) []map[string]any {
	switch v := raw.(type) {
	case []recordReportEntry:
		out := make([]map[string]any, len(v))
		for i, e := range v {
			out[i] = map[string]any(e)
		}
		return out
	case []map[string]any:
		return v
	default:
		b, err := json.Marshal(raw)
		if err != nil {
			return nil
		}
		var out []map[string]any
		if json.Unmarshal(b, &out) != nil {
			return nil
		}
		return out
	}
}

func collectMetricKeys(entries []map[string]any, metricsExpr string) []string {
	preferred := metricIdents(metricsExpr)
	seen := map[string]bool{}
	var keys []string
	add := func(k string) {
		if k == "" || seen[k] || strings.HasPrefix(k, "dimension_") {
			return
		}
		seen[k] = true
		keys = append(keys, k)
	}
	for _, k := range preferred {
		if len(entries) > 0 {
			if _, ok := entries[0][k]; ok {
				add(k)
			}
		}
	}
	if len(keys) == 0 && len(entries) > 0 {
		for k, v := range entries[0] {
			if strings.HasPrefix(k, "dimension_") {
				continue
			}
			if k == "count" {
				continue
			}
			if _, ok := toFloatOK(v); ok {
				add(k)
			}
		}
		if len(keys) == 0 {
			if _, ok := entries[0]["count"]; ok {
				add("count")
			}
		}
		sort.Strings(keys)
	}
	return keys
}

func metricIdents(metrics string) []string {
	if strings.TrimSpace(metrics) == "" {
		return nil
	}
	var out []string
	for _, part := range strings.Split(metrics, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		ident := part
		if i := strings.Index(strings.ToUpper(part), " AS "); i >= 0 {
			ident = strings.TrimSpace(part[i+4:])
		}
		out = append(out, ident)
	}
	return out
}

func defaultReportExprs(mod *types.Module, metrics, dimensions string) (string, string) {
	if metrics == "" {
		if f := firstNumberField(mod); f != nil {
			metrics = fmt.Sprintf("SUM(%s) AS %s", f.Name, f.Name)
		}
	}
	if dimensions == "" {
		if f := firstDateField(mod); f != nil {
			dimensions = fmt.Sprintf("DATE_FORMAT(%s, '%%Y-%%m-01')", f.Name)
		} else if f := firstCategoryField(mod); f != nil {
			dimensions = f.Name
		}
	}
	return metrics, wrapTimeDimension(mod, dimensions)
}

func wrapTimeDimension(mod *types.Module, dim string) string {
	dim = strings.TrimSpace(dim)
	if dim == "" || strings.Contains(dim, "(") {
		return dim
	}
	f := mod.Fields.FindByName(dim)
	if f != nil && strings.EqualFold(f.Kind, "DateTime") {
		return fmt.Sprintf("DATE_FORMAT(%s, '%%Y-%%m-01')", f.Name)
	}
	return dim
}

func looksLikeTimeDimension(mod *types.Module, dim string) bool {
	d := strings.ToLower(dim)
	if strings.Contains(d, "date_format") || strings.Contains(d, "month") || strings.Contains(d, "year") {
		return true
	}
	f := mod.Fields.FindByName(strings.TrimSpace(dim))
	return f != nil && strings.EqualFold(f.Kind, "DateTime")
}

func findModuleFlexible(ctx context.Context, ns uint64, idents ...string) *types.Module {
	if DefaultModule == nil || ns == 0 {
		return nil
	}
	set, _, err := DefaultModule.Find(ctx, types.ModuleFilter{NamespaceID: ns})
	if err != nil {
		return nil
	}
	for _, ident := range idents {
		ident = strings.TrimSpace(ident)
		if ident == "" {
			continue
		}
		if id := parseUint64(ident); id > 0 {
			for _, m := range set {
				if m.ID == id {
					return m
				}
			}
		}
		want := strings.ToLower(strings.ReplaceAll(ident, " ", "_"))
		for _, m := range set {
			handle := strings.ToLower(strings.ReplaceAll(m.Handle, " ", "_"))
			name := strings.ToLower(strings.ReplaceAll(m.Name, " ", "_"))
			if handle == want || name == want {
				return m
			}
		}
	}
	return nil
}

func pickNumericField(mod *types.Module, preferred string, fallbacks ...string) string {
	cands := append([]string{preferred}, fallbacks...)
	for _, name := range cands {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		if f := mod.Fields.FindByName(name); f != nil && strings.EqualFold(f.Kind, "Number") {
			return f.Name
		}
		for _, f := range mod.Fields {
			if strings.EqualFold(f.Name, name) && strings.EqualFold(f.Kind, "Number") {
				return f.Name
			}
		}
	}
	if f := firstNumberField(mod); f != nil {
		return f.Name
	}
	return strings.TrimSpace(preferred)
}

func pickDateField(mod *types.Module, preferred string) string {
	if f := mod.Fields.FindByName(preferred); f != nil && strings.EqualFold(f.Kind, "DateTime") {
		return f.Name
	}
	if f := firstDateField(mod); f != nil {
		return f.Name
	}
	return ""
}

func firstNumberField(mod *types.Module) *types.ModuleField {
	for _, f := range mod.Fields {
		if strings.EqualFold(f.Kind, "Number") {
			return f
		}
	}
	return nil
}

func firstDateField(mod *types.Module) *types.ModuleField {
	for _, f := range mod.Fields {
		if strings.EqualFold(f.Kind, "DateTime") {
			return f
		}
	}
	return nil
}

func firstCategoryField(mod *types.Module) *types.ModuleField {
	for _, f := range mod.Fields {
		switch strings.ToLower(f.Kind) {
		case "string", "select", "email", "url":
			return f
		}
	}
	return nil
}

func fieldLabel(mod *types.Module, name, fallback string) string {
	if f := mod.Fields.FindByName(name); f != nil && f.Label != "" {
		return f.Label
	}
	return fallback
}

func recordValue(r *types.Record, name string) string {
	if r == nil {
		return ""
	}
	for _, v := range r.Values {
		if v != nil && strings.EqualFold(v.Name, name) {
			return v.Value
		}
	}
	return ""
}

func normalizeChartType(t, def string) string {
	t = strings.ToLower(strings.TrimSpace(t))
	switch t {
	case "line", "bar", "pie", "doughnut":
		return t
	}
	if def != "" {
		return def
	}
	return "bar"
}

func firstNonEmpty(ss ...string) string {
	for _, s := range ss {
		if strings.TrimSpace(s) != "" {
			return strings.TrimSpace(s)
		}
	}
	return ""
}

func firstPresent(m map[string]any, keys ...string) any {
	for _, k := range keys {
		if v, ok := m[k]; ok && v != nil && stringifyChartVal(v) != "" {
			return v
		}
	}
	return nil
}

func stringifyChartVal(v any) string {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	case fmt.Stringer:
		return strings.TrimSpace(t.String())
	default:
		return strings.TrimSpace(fmt.Sprint(t))
	}
}

func toFloat(v any) float64 {
	n, _ := toFloatOK(v)
	return n
}

func toFloatOK(v any) (float64, bool) {
	switch t := v.(type) {
	case float64:
		return t, !math.IsNaN(t) && !math.IsInf(t, 0)
	case float32:
		return float64(t), true
	case int:
		return float64(t), true
	case int64:
		return float64(t), true
	case uint64:
		return float64(t), true
	case json.Number:
		n, err := t.Float64()
		return n, err == nil
	case string:
		return parseFloat(t)
	default:
		return parseFloat(fmt.Sprint(v))
	}
}

func parseFloat(s string) (float64, bool) {
	s = strings.TrimSpace(strings.ReplaceAll(s, ",", "."))
	if s == "" {
		return 0, false
	}
	n, err := strconv.ParseFloat(s, 64)
	if err != nil || math.IsNaN(n) || math.IsInf(n, 0) {
		return 0, false
	}
	return n, true
}

func round2(n float64) float64 {
	return math.Round(n*100) / 100
}
