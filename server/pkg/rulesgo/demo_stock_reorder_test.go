package rulesgo

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

type stockReorderSample struct {
	Name         string  `json:"name"`
	StoreID      float64 `json:"store_id"`
	ProductID    float64 `json:"product_id"`
	EAN          string  `json:"ean"`
	ProductName  string  `json:"product_name"`
	SupplierID   float64 `json:"supplier_id"`
	StockQty     float64 `json:"stock_quantity"`
	StockSum     float64 `json:"stock_sum"`
	AvgDaily     float64 `json:"avg_daily_qty"`
	DemandStd    float64 `json:"demand_std_daily"`
	DaysOfCover  float64 `json:"days_of_cover"`
	ZeroStock    float64 `json:"zeroStock"`
	LeadTime     float64 `json:"lead_time_days"`
	ReviewPeriod float64 `json:"review_period_days"`
	TargetDOC    float64 `json:"target_doc_days"`
	MOQ          float64 `json:"moq"`
	PackSize     float64 `json:"pack_size"`
	ServiceZ     float64 `json:"service_z"`
	UnitCost     float64 `json:"unit_cost"`
	SafetyStock  float64 `json:"safety_stock"`
	ReorderPoint float64 `json:"reorder_point"`
	TargetQty    float64 `json:"target_qty"`
	ReorderQty   float64 `json:"reorder_qty"`
	OrderValue   float64 `json:"order_value"`
	HealthLevel  string  `json:"health_level"`
	ExpectLevel  string  `json:"expect_level"`
}

func (s stockReorderSample) Input() map[string]interface{} {
	return map[string]interface{}{
		"store_id":           s.StoreID,
		"product_id":         s.ProductID,
		"ean":                s.EAN,
		"product_name":       s.ProductName,
		"supplier_id":        s.SupplierID,
		"stock_quantity":     s.StockQty,
		"stock_sum":          s.StockSum,
		"avg_daily_qty":      s.AvgDaily,
		"demand_std_daily":   s.DemandStd,
		"days_of_cover":      s.DaysOfCover,
		"zeroStock":          s.ZeroStock,
		"lead_time_days":     s.LeadTime,
		"review_period_days": s.ReviewPeriod,
		"target_doc_days":    s.TargetDOC,
		"moq":                s.MOQ,
		"pack_size":          s.PackSize,
		"service_z":          s.ServiceZ,
		"unit_cost":          s.UnitCost,
		"safety_stock":       s.SafetyStock,
		"reorder_point":      s.ReorderPoint,
		"target_qty":         s.TargetQty,
		"reorder_qty":        s.ReorderQty,
		"order_value":        s.OrderValue,
		"health_level":       s.HealthLevel,
	}
}

func loadStockReorderSamples(t *testing.T) []stockReorderSample {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	path := filepath.Join(filepath.Dir(file), "testdata", "stock_reorder_samples.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read samples: %v", err)
	}
	var samples []stockReorderSample
	if err := json.Unmarshal(raw, &samples); err != nil {
		t.Fatalf("parse samples: %v", err)
	}
	if len(samples) == 0 {
		t.Fatal("no samples")
	}
	return samples
}

func TestDemoStockHealthChain(t *testing.T) {
	reg := DefaultRegistry(nil)
	eng := NewEngine(reg)
	chain := DemoStockHealthChain()
	eng.RegisterChain(chain)

	for _, sample := range loadStockReorderSamples(t) {
		res, err := eng.Run(context.Background(), chain.ID, sample.Input())
		if err != nil {
			t.Fatalf("%s: %v", sample.Name, err)
		}
		if !res.Success {
			t.Fatalf("%s failed: %s", sample.Name, res.Error)
		}
		level, _ := res.Output["level"].(string)
		t.Logf("%-22s DOC=%.1f zero=%.0f score=%v level=%s (expect %s)",
			sample.Name, sample.DaysOfCover, sample.ZeroStock, res.Output["score"], level, sample.ExpectLevel)
		if level == "" {
			t.Fatalf("%s: empty level", sample.Name)
		}
		if sample.ExpectLevel != "" && level != sample.ExpectLevel {
			t.Fatalf("%s: level want %s got %s (score=%v residual=%v)",
				sample.Name, sample.ExpectLevel, level, res.Output["score"], res.Output["residualScore"])
		}
	}
}

func TestDemoAutoReorderChainNeedGate(t *testing.T) {
	reg := DefaultRegistry(nil)
	eng := NewEngine(reg)
	chain := DemoAutoReorderChain()
	eng.RegisterChain(chain)

	samples := loadStockReorderSamples(t)
	var withNeed, withoutNeed *stockReorderSample
	for i := range samples {
		s := &samples[i]
		if s.ReorderQty > 0 && withNeed == nil {
			withNeed = s
		}
		if s.ReorderQty == 0 && withoutNeed == nil {
			withoutNeed = s
		}
	}
	if withNeed == nil || withoutNeed == nil {
		t.Fatal("need both reorder and zero-reorder samples")
	}

	res, err := eng.Run(context.Background(), chain.ID, withNeed.Input())
	if err != nil {
		t.Fatal(err)
	}
	if !res.Success {
		t.Fatalf("need>0 failed: %s", res.Error)
	}
	if res.Output["score"] == nil {
		t.Fatalf("expected priority score when reorder_qty>0, got %#v", res.Output)
	}
	// CRUD service is nil in unit tests → status from create node
	if status, _ := res.Output["status"].(string); status != "" && status != "crud_service_not_configured" {
		t.Fatalf("unexpected crud status %q", status)
	}

	res2, err := eng.Run(context.Background(), chain.ID, withoutNeed.Input())
	if err != nil {
		t.Fatal(err)
	}
	if !res2.Success {
		t.Fatalf("need=0 failed: %s", res2.Error)
	}
	if pass, _ := res2.Output["need_result"].(string); pass == "true" {
		t.Fatalf("expected empty/false need_result for reorder_qty=0, got %q", pass)
	}
	visited := make([]string, 0, len(res2.Nodes))
	for _, n := range res2.Nodes {
		visited = append(visited, n.NodeID)
	}
	for _, n := range res2.Nodes {
		if n.NodeID == "priority" || n.NodeID == "create_line" {
			t.Fatalf("reorder_qty=0 must not follow need→priority edge, visited %v", visited)
		}
	}
}
