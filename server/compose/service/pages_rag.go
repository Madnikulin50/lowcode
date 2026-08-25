package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/madnikulin50/lowcode/server/compose/types"
	"github.com/madnikulin50/lowcode/server/pkg/auth"
	"github.com/madnikulin50/lowcode/server/pkg/filter"
	"github.com/madnikulin50/lowcode/server/pkg/rag"
	"go.uber.org/zap"
)

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
}

func NewPagesRAGService(store *rag.Store, embedder *rag.Embedder, log *zap.Logger, pages pagesFinder, namespaces NamespaceService, rec RecordService) *PagesRAGService {
	return &PagesRAGService{store: store, embedder: embedder, log: log, pages: pages, namespaces: namespaces, record: rec}
}

func (s *PagesRAGService) Crawl(ctx context.Context) error {
	s.log.Info("pages RAG: starting daily crawl")
	start := time.Now()

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

	if err := s.store.ClearPages(); err != nil {
		return fmt.Errorf("pages RAG: clear: %w", err)
	}

	var indexed int
	for _, p := range allPages {
		for _, block := range p.Blocks {
			cfgText := rag.ExtractBlockText(block)

			var dataText string
			switch block.Kind {
			case "Metric":
				dataText = s.fetchMetricData(ctx, p.NamespaceID, block.Options)
			case "RecordList":
				dataText = s.fetchRecordListData(ctx, p.NamespaceID, block.Options)
			}

			var text string
			if dataText != "" {
				text = fmt.Sprintf("Page: %s (%d)\nBlock: %s (%s)\nData:\n%s\n---\nConfig: %s", p.Title, p.ID, block.Title, block.Kind, dataText, cfgText)
			} else if cfgText != "" {
				text = fmt.Sprintf("Page: %s (%d)\nBlock: %s (%s)\n%s", p.Title, p.ID, block.Title, block.Kind, cfgText)
			} else {
				continue
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
		modIDStr, _ := mm["moduleID"].(string)
		metricField, _ := mm["metricField"].(string)
		operation, _ := mm["operation"].(string)
		filter, _ := mm["filter"].(string)
		label, _ := mm["label"].(string)
		prefix, _ := mm["prefix"].(string)
		suffix, _ := mm["suffix"].(string)

		if modIDStr == "" || modIDStr == "0" {
			continue
		}
		modID, err := strconv.ParseUint(modIDStr, 10, 64)
		if err != nil {
			continue
		}
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

func (s *PagesRAGService) fetchRecordListData(ctx context.Context, namespaceID uint64, opts map[string]interface{}) string {
	modIDStr, _ := opts["moduleID"].(string)
	if modIDStr == "" || modIDStr == "0" {
		return ""
	}
	modID, err := strconv.ParseUint(modIDStr, 10, 64)
	if err != nil {
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
		time.Sleep(1 * time.Minute)
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
