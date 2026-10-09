package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/madnikulin50/lowcode/server/compose/types"
	"github.com/madnikulin50/lowcode/server/pkg/auth"
	"github.com/madnikulin50/lowcode/server/pkg/filter"
	"github.com/madnikulin50/lowcode/server/pkg/locale"
	"github.com/madnikulin50/lowcode/server/pkg/rag"
	"go.uber.org/zap"
	"golang.org/x/text/language"
)

type PagesRAGProgress struct {
	Running       bool   `json:"running"`
	TotalPages    int    `json:"totalPages"`
	IndexedPages  int    `json:"indexedPages"`
	CurrentPage   string `json:"currentPage"`
	TotalBlocks   int    `json:"totalBlocks"`
	IndexedBlocks int    `json:"indexedBlocks"`
	Complete      bool   `json:"complete"`
}

type pagesFinder interface {
	Find(ctx context.Context, filter types.PageFilter) (types.PageSet, types.PageFilter, error)
}

type PagesRAGService struct {
	store      *rag.Store
	embedder   *rag.Embedder
	log        *zap.Logger
	pages      pagesFinder
	namespaces NamespaceService
	record     RecordService
	locale     ResourceTranslationsManagerService
	mu         sync.Mutex
	progress   PagesRAGProgress
}

func NewPagesRAGService(store *rag.Store, embedder *rag.Embedder, log *zap.Logger, pages pagesFinder, namespaces NamespaceService, rec RecordService, locale ResourceTranslationsManagerService) *PagesRAGService {
	return &PagesRAGService{store: store, embedder: embedder, log: log, pages: pages, namespaces: namespaces, record: rec, locale: locale}
}

// parseModuleID extracts module ID from interface{} which could be string or float64.
func parseModuleID(v interface{}) (uint64, bool) {
	switch t := v.(type) {
	case string:
		if t == "" || t == "0" {
			return 0, false
		}
		id, err := strconv.ParseUint(t, 10, 64)
		return id, err == nil
	case float64:
		id := uint64(t)
		return id, id != 0
	}
	return 0, false
}

// blockTiedToRecord reports whether a block's options reference the current
// record context (${recordID} variable). Such blocks only make sense on
// record pages and are excluded from RAG indexing.
func blockTiedToRecord(opts map[string]interface{}) bool {
	for _, v := range opts {
		switch t := v.(type) {
		case string:
			if strings.Contains(t, "${recordID}") {
				return true
			}
		case map[string]interface{}:
			if blockTiedToRecord(t) {
				return true
			}
		case []interface{}:
			for _, item := range t {
				if m, ok := item.(map[string]interface{}); ok && blockTiedToRecord(m) {
					return true
				}
			}
		}
	}
	return false
}

// translatedBlockText collects the page and block title (plus block
// description and content body) in every supported language except the
// default one, so RAG chunks are searchable in all available translations.
func (s *PagesRAGService) translatedBlockText(p *types.Page, block types.PageBlock, blockIndex int) string {
	if s.locale == nil || s.locale.Locale() == nil {
		return ""
	}
	lsvc := s.locale.Locale()

	var defTag language.Tag
	if def := lsvc.Default(); def != nil {
		defTag = def.Tag
	}

	blockID := locale.ContentID(block.BlockID, blockIndex)
	rpl := strings.NewReplacer("{{blockID}}", strconv.FormatUint(blockID, 10))
	titleKey := rpl.Replace(types.LocaleKeyPagePageBlockBlockIDTitle.Path)
	descKey := rpl.Replace(types.LocaleKeyPagePageBlockBlockIDDescription.Path)
	bodyKey := rpl.Replace(types.LocaleKeyPagePageBlockBlockIDContentBody.Path)

	baseTitle, baseDescription := block.Title, block.Description
	var baseBody string
	if block.Kind == "Content" {
		baseBody, _ = block.Options["body"].(string)
	}

	var lines []string
	for _, tag := range lsvc.Tags() {
		if tag == defTag {
			continue
		}
		tt := lsvc.ResourceTranslations(tag, p.ResourceTranslation())
		if len(tt) == 0 {
			continue
		}

		pageTitle, title, description := p.Title, baseTitle, baseDescription
		body := baseBody

		if aux := tt.FindByKey(types.LocaleKeyPageTitle.Path); aux != nil {
			pageTitle = aux.Msg
		}
		if aux := tt.FindByKey(titleKey); aux != nil {
			title = aux.Msg
		}
		if aux := tt.FindByKey(descKey); aux != nil {
			description = aux.Msg
		}
		if aux := tt.FindByKey(bodyKey); aux != nil {
			body = aux.Msg
		}

		if pageTitle == p.Title && title == baseTitle && description == baseDescription && body == "" {
			continue
		}

		line := fmt.Sprintf("[%s] %s / %s: %s", tag, pageTitle, block.Kind, title)
		if description != "" && description != baseDescription {
			line += " (" + description + ")"
		}
		lines = append(lines, line)
		if body != "" {
			lines = append(lines, fmt.Sprintf("[%s] content: %s", tag, body))
		}
	}
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n")
}

func (s *PagesRAGService) Crawl(ctx context.Context) error {
	s.log.Info("pages RAG: starting daily crawl")
	start := time.Now()

	s.mu.Lock()
	s.progress = PagesRAGProgress{Running: true}
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		s.progress.Running = false
		s.progress.Complete = true
		s.mu.Unlock()
	}()

	// Fetch all namespaces
	nsSet, _, err := s.namespaces.Find(ctx, types.NamespaceFilter{})
	if err != nil {
		return fmt.Errorf("pages RAG: find namespaces: %w", err)
	}

	// Collect all pages across all namespaces
	var allPages types.PageSet
	for _, ns := range nsSet {
		filter := types.PageFilter{NamespaceID: ns.ID}
		filter.Limit = 10000
		pages, _, err := s.pages.Find(ctx, filter)
		if err != nil {
			s.log.Warn("pages RAG: find pages for namespace", zap.Uint64("nsID", ns.ID), zap.Error(err))
			continue
		}
		allPages = append(allPages, pages...)
	}

	s.mu.Lock()
	s.progress.TotalPages = len(allPages)
	s.mu.Unlock()

	if err := s.store.ClearPages(); err != nil {
		return fmt.Errorf("pages RAG: clear: %w", err)
	}

	var indexed int
	for pi, p := range allPages {
		s.mu.Lock()
		s.progress.CurrentPage = p.Title
		s.progress.IndexedPages = pi
		s.mu.Unlock()

		for i, block := range p.Blocks {
			if blockTiedToRecord(block.Options) {
				s.log.Debug("pages RAG: skip record-bound block", zap.String("page", p.Title), zap.String("kind", block.Kind), zap.String("title", block.Title))
				continue
			}

			cfgText := rag.ExtractBlockText(block)

			var dataText string
			switch block.Kind {
			case "Metric":
				dataText = s.fetchMetricData(ctx, p.NamespaceID, block.Options)
			case "RecordList":
				dataText = s.fetchRecordListData(ctx, p.NamespaceID, p.ModuleID, block.Options)
			default:
				continue
			}

			var text string
			if dataText != "" {
				text = fmt.Sprintf("Page: %s (%d)\nBlock: %s (%s)\nData:\n%s\n---\nConfig: %s", p.Title, p.ID, block.Title, block.Kind, dataText, cfgText)
			} else if cfgText != "" {
				text = fmt.Sprintf("Page: %s (%d)\nBlock: %s (%s)\n%s", p.Title, p.ID, block.Title, block.Kind, cfgText)
			} else {
				continue
			}

			if translated := s.translatedBlockText(p, block, i); translated != "" {
				text += "\n\nTranslations:\n" + translated
			}

			chunks := rag.ChunkText(text, 512, 64)
			for i, chunk := range chunks {
				emb, err := s.embedder.Embed(chunk)
				if err != nil {
					s.log.Warn("pages RAG: embed failed", zap.String("page", p.Title), zap.Int("chunk", i), zap.Error(err))
					continue
				}
				pc := rag.PageChunk{
					ID:          fmt.Sprintf("%d:%d:%d", p.ID, block.BlockID, time.Now().UnixNano()),
					PageID:      p.ID,
					NamespaceID: p.NamespaceID,
					Title:       p.Title,
					BlockKind:   block.Kind,
					BlockTitle:  block.Title,
					Text:        chunk,
					ChunkIndex:  i,
				}
				if err := s.store.PutPageChunk(pc, emb); err != nil {
					s.log.Warn("pages RAG: store failed", zap.Error(err))
					continue
				}
			}
			indexed++
		}
	}

	if err := s.store.SetPagesCrawlTime(time.Now().Unix()); err != nil {
		s.log.Warn("pages RAG: save crawl time", zap.Error(err))
	}

	s.mu.Lock()
	s.progress.IndexedPages = len(allPages)
	s.progress.IndexedBlocks = indexed
	s.progress.CurrentPage = ""
	s.mu.Unlock()

	s.log.Info("pages RAG: crawl done", zap.Int("blocks", indexed), zap.Int("namespaces", len(nsSet)), zap.Duration("took", time.Since(start)))
	return nil
}

func (s *PagesRAGService) fetchMetricData(ctx context.Context, namespaceID uint64, opts map[string]interface{}) string {
	metricsRaw, ok := opts["metrics"]
	if !ok {
		return ""
	}
	arr, ok := metricsRaw.([]interface{})
	if !ok {
		return ""
	}
	var lines []string
	for _, m := range arr {
		mm, ok := m.(map[string]interface{})
		if !ok {
			continue
		}
		modID, ok := parseModuleID(mm["moduleID"])
		if !ok || modID == 0 {
			continue
		}
		metricField, _ := mm["metricField"].(string)
		operation, _ := mm["operation"].(string)
		filter, _ := mm["filter"].(string)
		label, _ := mm["label"].(string)
		prefix, _ := mm["prefix"].(string)
		suffix, _ := mm["suffix"].(string)

		if label == "" {
			label = metricField
		}

		var metricExpr string
		if metricField == "number_expression" {
			expr, _ := mm["expression"].(string)
			metricExpr = fmt.Sprintf("(%s) AS rp", expr)
		} else if operation != "" && metricField != "" {
			metricExpr = fmt.Sprintf("%s(%s) AS rp", strings.ToUpper(operation), metricField)
		} else {
			metricExpr = "COUNT(*) AS rp"
		}

		val := s.computeMetricValue(ctx, namespaceID, modID, metricExpr, filter)
		if val == "" {
			continue
		}
		lines = append(lines, fmt.Sprintf("%s: %s%s%s", label, prefix, val, suffix))
	}
	return strings.Join(lines, "\n")
}

func (s *PagesRAGService) computeMetricValue(ctx context.Context, namespaceID, moduleID uint64, metricsExpr, filter string) string {
	if s.record == nil {
		return ""
	}
	raw, err := s.record.Report(ctx, namespaceID, moduleID, metricsExpr, "deletedAt", filter)
	if err != nil {
		return ""
	}
	entries, ok := raw.([]recordReportEntry)
	if !ok || len(entries) == 0 {
		return "0"
	}
	var total float64
	for _, e := range entries {
		if rpVal, ok := e["rp"]; ok && rpVal != nil {
			switch v := rpVal.(type) {
			case float64:
				total += v
			case int64:
				total += float64(v)
			case int:
				total += float64(v)
			}
		}
	}
	if total == 0 && len(entries) > 0 {
		if cnt, ok := entries[0]["count"]; ok && cnt != nil {
			switch v := cnt.(type) {
			case float64:
				total = v
			case int64:
				total = float64(v)
			case int:
				total = float64(v)
			}
		}
	}
	if total == float64(int64(total)) {
		return fmt.Sprintf("%d", int64(total))
	}
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", total), "0"), ".")
}

func (s *PagesRAGService) fetchRecordListData(ctx context.Context, namespaceID, pageModuleID uint64, opts map[string]interface{}) string {
	modID, ok := parseModuleID(opts["moduleID"])
	if !ok || modID == 0 {
		modID = pageModuleID
	}
	if modID == 0 {
		return ""
	}
	prefilter, _ := opts["prefilter"].(string)
	if s.record == nil {
		return ""
	}
	rf := types.RecordFilter{
		ModuleID:    modID,
		NamespaceID: namespaceID,
		Query:       prefilter,
		Deleted:     0,
	}
	rf.Limit = 5
	rf.Sort = filter.SortExprSet{&filter.SortExpr{Column: "createdAt", Descending: true}}

	set, _, err := s.record.Find(ctx, rf)
	if err != nil || len(set) == 0 {
		return fmt.Sprintf("Module %d records (no data available)", modID)
	}

	var lines []string
	lines = append(lines, fmt.Sprintf("Module %d: %d records shown (latest)", modID, len(set)))
	for i, rec := range set {
		var vals []string
		for _, rv := range rec.Values {
			vals = append(vals, fmt.Sprintf("%s: %v", rv.Name, rv.Value))
		}
		if len(vals) > 5 {
			vals = vals[:5]
			vals = append(vals, "...")
		}
		lines = append(lines, fmt.Sprintf("  Record %d: %s", i+1, strings.Join(vals, ", ")))
	}
	return strings.Join(lines, "\n")
}

func (s *PagesRAGService) Search(ctx context.Context, query string, topK int) ([]rag.SearchResult, error) {
	emb, err := s.embedder.Embed(query)
	if err != nil {
		return nil, fmt.Errorf("pages RAG: embed: %w", err)
	}
	return s.store.SearchPages(emb, topK)
}

func (s *PagesRAGService) ListPages(ctx context.Context) ([]rag.PageChunk, error) {
	return s.store.ListPageChunks()
}

func (s *PagesRAGService) Reindex(ctx context.Context) error {
	return s.Crawl(ctx)
}

func (s *PagesRAGService) Progress() PagesRAGProgress {
	s.mu.Lock()
	defer s.mu.Unlock()
	p := s.progress
	return p
}

func (s *PagesRAGService) BuildContext(ctx context.Context, query string, topK int) string {
	results, err := s.Search(ctx, query, topK)
	if err != nil || len(results) == 0 {
		return ""
	}
	var out string
	for i, r := range results {
		out += fmt.Sprintf("\n[Page %d]: %s\n", i+1, r.Chunk.Text)
	}
	return out
}

func (s *PagesRAGService) StartDailyCrawl(ctx context.Context) {
	serveCtx := auth.SetIdentityToContext(ctx, auth.ServiceUser())
	go func() {
		timer := time.NewTimer(1 * time.Minute)
		defer timer.Stop()
		select {
		case <-timer.C:
		case <-ctx.Done():
			return
		}
		s.log.Info("pages RAG: initial crawl starting")
		if err := s.Crawl(serveCtx); err != nil {
			s.log.Error("pages RAG: initial crawl failed", zap.Error(err))
		}
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := s.Crawl(serveCtx); err != nil {
					s.log.Error("pages RAG: daily crawl failed", zap.Error(err))
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}
