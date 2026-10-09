package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/madnikulin50/lowcode/server/compose/types"
	"github.com/madnikulin50/lowcode/server/pkg/chat"
)

const chatPageChartMaxPages = 40

type pageChartRef struct {
	BlockID     uint64
	BlockTitle  string
	ChartID     uint64
	ChartName   string
	NamespaceID uint64
	PageID      uint64
}

func showPageChart(ctx context.Context, params map[string]string) string {
	if params == nil {
		params = map[string]string{}
	}
	ns := nsID(ctx, params)
	if ns == 0 {
		return "Missing namespaceID"
	}
	title := strings.TrimSpace(params["title"])
	blockID := parseUint64(firstNonEmpty(params["blockID"], params["blockId"]))
	chartID := parseUint64(firstNonEmpty(params["chartID"], params["chartId"]))
	if title == "" && blockID == 0 && chartID == 0 {
		return "Укажите title (заголовок блока на странице), blockID или chartID."
	}

	names := chartNamesByID(ctx, ns)
	pageID := pageIDFromCtx(ctx, params)

	if pageID > 0 {
		if page, err := DefaultPage.FindByID(ctx, ns, pageID); err == nil && page != nil {
			refs := collectPageCharts(page, names, layoutVisibleBlockIDs(ctx, page))
			if hit := matchPageChart(refs, title, blockID, chartID); hit != nil {
				return encodeComposeChartFence(*hit)
			}
		}
	}

	pages, _, err := DefaultPage.Find(ctx, types.PageFilter{NamespaceID: ns})
	if err != nil {
		if title != "" {
			return fmt.Sprintf("График «%s» на этой странице не найден.", title)
		}
		return "График не найден среди блоков Chart на текущей странице."
	}
	if len(pages) > chatPageChartMaxPages {
		pages = pages[:chatPageChartMaxPages]
	}
	var rest []pageChartRef
	for _, p := range pages {
		if p == nil || p.ID == pageID {
			continue
		}
		rest = append(rest, collectPageCharts(p, names, layoutVisibleBlockIDs(ctx, p))...)
	}
	if hit := matchPageChart(rest, title, blockID, chartID); hit != nil {
		return encodeComposeChartFence(*hit)
	}
	if title != "" {
		return fmt.Sprintf("График «%s» не найден среди блоков Chart на страницах пространства.", title)
	}
	return "График не найден. Укажите точный заголовок блока."
}

func showPageChartFastPath(ctx context.Context, ask *ChatPromptArguments) string {
	if ask == nil {
		return ""
	}
	title := extractShowChartTitle(ask.Prompt)
	if title == "" || ask.Namespace == 0 || ask.Page == 0 {
		return ""
	}
	page, err := DefaultPage.FindByID(ctx, ask.Namespace, ask.Page)
	if err != nil || page == nil {
		return ""
	}
	names := chartNamesByID(ctx, ask.Namespace)
	refs := collectPageCharts(page, names, layoutVisibleBlockIDs(ctx, page))
	hit := matchPageChart(refs, title, 0, 0)
	if hit == nil {
		return ""
	}
	return encodeComposeChartFence(*hit)
}

func pageChartsSystemHint(ctx context.Context, ask *ChatPromptArguments) string {
	if ask == nil || ask.Page == 0 || ask.Namespace == 0 || DefaultPage == nil {
		return ""
	}
	if ctx == nil {
		ctx = context.Background()
	}
	page, err := DefaultPage.FindByID(ctx, ask.Namespace, ask.Page)
	if err != nil || page == nil {
		return ""
	}
	names := chartNamesByID(ctx, ask.Namespace)
	refs := collectPageCharts(page, names, layoutVisibleBlockIDs(ctx, page))
	if len(refs) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("Charts on this page (show with show_page_chart by title):\n")
	for _, ref := range refs {
		fmt.Fprintf(&b, "- title=%q chartID=%d", ref.BlockTitle, ref.ChartID)
		if ref.ChartName != "" {
			fmt.Fprintf(&b, " chartName=%q", ref.ChartName)
		}
		b.WriteByte('\n')
	}
	b.WriteString("If the user names one of these titles, you MUST call show_page_chart, not sales_dynamics or visualize_report.")
	return b.String()
}

func collectPageCharts(page *types.Page, names map[uint64]string, visible map[uint64]bool) []pageChartRef {
	if page == nil {
		return nil
	}
	tabbed := tabbedChartBlockIDs(page, visible)
	out := make([]pageChartRef, 0)
	for _, blk := range page.Blocks {
		if !strings.EqualFold(blk.Kind, "Chart") {
			continue
		}
		chartID := parseUint64(optionString(blk.Options, "chartID"))
		if chartID == 0 {
			continue
		}
		if len(visible) > 0 && !visible[blk.BlockID] && !tabbed[blk.BlockID] {
			continue
		}
		ref := pageChartRef{
			BlockID:     blk.BlockID,
			BlockTitle:  strings.TrimSpace(blk.Title),
			ChartID:     chartID,
			ChartName:   names[chartID],
			NamespaceID: page.NamespaceID,
			PageID:      page.ID,
		}
		out = append(out, ref)
	}
	return out
}

func matchPageChart(refs []pageChartRef, title string, blockID, chartID uint64) *pageChartRef {
	if blockID > 0 {
		for i := range refs {
			if refs[i].BlockID == blockID && refs[i].ChartID > 0 {
				return &refs[i]
			}
		}
	}
	if chartID > 0 {
		for i := range refs {
			if refs[i].ChartID == chartID {
				return &refs[i]
			}
		}
	}
	q := strings.ToLower(strings.TrimSpace(title))
	if q == "" {
		return nil
	}
	var best *pageChartRef
	bestScore := 0
	for i := range refs {
		if refs[i].ChartID == 0 {
			continue
		}
		score := titleMatchScore(q, refs[i].BlockTitle, refs[i].ChartName)
		if score > bestScore {
			bestScore = score
			best = &refs[i]
		}
	}
	if bestScore == 0 {
		return nil
	}
	return best
}

func titleMatchScore(q, blockTitle, chartName string) int {
	bt := strings.ToLower(strings.TrimSpace(blockTitle))
	cn := strings.ToLower(strings.TrimSpace(chartName))
	switch {
	case bt != "" && bt == q:
		return 400
	case cn != "" && cn == q:
		return 300
	case bt != "" && strings.Contains(bt, q):
		return 200 + closeness(bt, q)
	case bt != "" && strings.Contains(q, bt):
		return 180 + closeness(q, bt)
	case cn != "" && strings.Contains(cn, q):
		return 100 + closeness(cn, q)
	case cn != "" && strings.Contains(q, cn):
		return 80 + closeness(q, cn)
	}
	return 0
}

func closeness(a, b string) int {
	d := len(a) - len(b)
	if d < 0 {
		d = -d
	}
	if d > 50 {
		d = 50
	}
	return 50 - d
}

func extractShowChartTitle(prompt string) string {
	p := strings.TrimSpace(prompt)
	if p == "" {
		return ""
	}
	lower := strings.ToLower(p)
	prefixes := []string{
		"покажите график ", "покажи график ",
		"покажите диаграмму ", "покажи диаграмму ",
		"покажите диаграмма ", "покажи диаграмма ",
		"show the chart ", "show chart ",
		"display the chart ", "display chart ",
		"покажите ", "покажи ",
		"выведите ", "выведи ",
		"откройте ", "открой ",
		"show ", "display ",
	}
	rest := ""
	for _, pre := range prefixes {
		if strings.HasPrefix(lower, pre) {
			rest = strings.TrimSpace(p[len(pre):])
			break
		}
	}
	if rest == "" {
		return ""
	}
	rest = strings.Trim(rest, " \t\"'«»“”.,!?:;")
	for _, extra := range []string{"график ", "графика ", "диаграмму ", "диаграмма ", "chart "} {
		if strings.HasPrefix(strings.ToLower(rest), extra) {
			rest = strings.TrimSpace(rest[len(extra):])
			rest = strings.Trim(rest, " \t\"'«»“”.,!?:;")
		}
	}
	if rest == "" || isGenericChartWord(rest) {
		return ""
	}
	return rest
}

func isGenericChartWord(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "график", "графики", "диаграмма", "диаграмму", "диаграммы", "chart", "charts":
		return true
	}
	return false
}

func encodeComposeChartFence(ref pageChartRef) string {
	title := ref.BlockTitle
	if title == "" {
		title = ref.ChartName
	}
	body, err := json.Marshal(composeChartSpecJSON{
		ChartID:     strconv.FormatUint(ref.ChartID, 10),
		NamespaceID: strconv.FormatUint(ref.NamespaceID, 10),
		Title:       title,
	})
	if err != nil {
		return fmt.Sprintf("Failed to encode chart: %v", err)
	}
	return "```compose-chart\n" + string(body) + "\n```"
}

type composeChartSpecJSON struct {
	ChartID     string `json:"chartID"`
	NamespaceID string `json:"namespaceID"`
	Title       string `json:"title,omitempty"`
}

func layoutVisibleBlockIDs(ctx context.Context, page *types.Page) map[uint64]bool {
	if DefaultPageLayout == nil || page == nil {
		return nil
	}
	set, _, err := DefaultPageLayout.Find(ctx, types.PageLayoutFilter{
		NamespaceID: page.NamespaceID,
		PageID:      page.ID,
	})
	if err != nil || len(set) == 0 {
		return nil
	}
	var layout *types.PageLayout
	for _, l := range set {
		if l != nil && l.Primary {
			layout = l
			break
		}
	}
	if layout == nil {
		layout = set[0]
	}
	if layout == nil {
		return nil
	}
	out := map[uint64]bool{}
	for _, b := range layout.Blocks {
		if b.BlockID > 0 {
			out[b.BlockID] = true
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func tabbedChartBlockIDs(page *types.Page, visible map[uint64]bool) map[uint64]bool {
	out := map[uint64]bool{}
	if page == nil {
		return out
	}
	for _, blk := range page.Blocks {
		if !strings.EqualFold(blk.Kind, "Tabs") {
			continue
		}
		if len(visible) > 0 && !visible[blk.BlockID] {
			continue
		}
		raw, ok := blk.Options["tabs"]
		if !ok || raw == nil {
			continue
		}
		var tabs []any
		switch t := raw.(type) {
		case []any:
			tabs = t
		case []map[string]any:
			for _, m := range t {
				tabs = append(tabs, m)
			}
		default:
			continue
		}
		for _, item := range tabs {
			m, ok := item.(map[string]any)
			if !ok {
				continue
			}
			id := anyToUint64(m["blockID"])
			if id > 0 {
				out[id] = true
			}
		}
	}
	return out
}

func chartNamesByID(ctx context.Context, ns uint64) map[uint64]string {
	out := map[uint64]string{}
	if DefaultChart == nil || ns == 0 {
		return out
	}
	set, _, err := DefaultChart.Find(ctx, types.ChartFilter{NamespaceID: ns})
	if err != nil {
		return out
	}
	for _, ch := range set {
		if ch != nil {
			out[ch.ID] = ch.Name
		}
	}
	return out
}

func pageIDFromCtx(ctx context.Context, params map[string]string) uint64 {
	if ctx != nil {
		if v := ctx.Value(chat.EnvPageID); v != nil {
			if id, ok := v.(uint64); ok && id > 0 {
				return id
			}
		}
	}
	if params == nil {
		return 0
	}
	return parseUint64(firstNonEmpty(params["pageID"], params["pageId"]))
}

func optionString(opts map[string]any, key string) string {
	if opts == nil {
		return ""
	}
	v, ok := opts[key]
	if !ok || v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	case json.Number:
		return strings.TrimSpace(t.String())
	case float64:
		return strconv.FormatUint(uint64(t), 10)
	case uint64:
		return strconv.FormatUint(t, 10)
	case int:
		return strconv.Itoa(t)
	default:
		return strings.TrimSpace(fmt.Sprint(t))
	}
}

func anyToUint64(v any) uint64 {
	if v == nil {
		return 0
	}
	switch t := v.(type) {
	case uint64:
		return t
	case int:
		if t < 0 {
			return 0
		}
		return uint64(t)
	case int64:
		if t < 0 {
			return 0
		}
		return uint64(t)
	case float64:
		return uint64(t)
	case json.Number:
		n, err := t.Int64()
		if err != nil || n < 0 {
			return 0
		}
		return uint64(n)
	case string:
		return parseUint64(t)
	default:
		return parseUint64(fmt.Sprint(v))
	}
}
