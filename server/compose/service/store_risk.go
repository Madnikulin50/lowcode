package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
)

// StoreRiskInput is the live ops bag fed into demo_store_risk.
type StoreRiskInput struct {
	StoreID               string  `json:"store_id" db:"store_id"`
	Name                  string  `json:"name" db:"name"`
	City                  string  `json:"city" db:"city"`
	Region                string  `json:"region" db:"region"`
	ShrinkPct             float64 `json:"shrinkPct" db:"shrink_pct"`
	Incidents90d          float64 `json:"incidents90d" db:"incidents_90d"`
	DaysSinceAudit        float64 `json:"daysSinceAudit" db:"days_since_audit"`
	RevenueImpact         float64 `json:"revenueImpact" db:"revenue_impact"`
	ControlEffectiveness  float64 `json:"controlEffectiveness" db:"control_effectiveness"`
	Likelihood            float64 `json:"likelihood" db:"likelihood"`
	Impact                float64 `json:"impact" db:"impact"`
	OpenIncidentCount     float64 `json:"openIncidentCount" db:"open_incident_count"`
	CriticalIncidentCount float64 `json:"criticalIncidentCount" db:"critical_incident_count"`
	StockCriticalCount    float64 `json:"stockCriticalCount" db:"stock_critical_count"`
	StockUnderstockCount  float64 `json:"stockUnderstockCount" db:"stock_understock_count"`
	SKUCount              float64 `json:"skuCount" db:"sku_count"`
	DaysOfCover           float64 `json:"daysOfCover" db:"days_of_cover"`
	ReorderQty            float64 `json:"reorderQty" db:"reorder_qty"`
	OrderValue            float64 `json:"orderValue" db:"order_value"`
}

// GatherStoreRiskInput computes shrink / incidents / audit lag / impact from live tables.
func GatherStoreRiskInput(ctx context.Context, storeID string) (*StoreRiskInput, error) {
	db, err := incidentDB(ctx)
	if err != nil {
		return nil, err
	}
	sid := strings.TrimSpace(storeID)
	if sid == "" {
		return nil, fmt.Errorf("store_id is required")
	}

	var in StoreRiskInput
	err = sqlx.GetContext(ctx, db, &in, `
WITH sid AS (
  SELECT $1::numeric AS store_id
),
st AS (
  SELECT s.store_id, s.store_name, s.store_area, s.region_id, s.address
    FROM stores s
    JOIN sid ON s.store_id = sid.store_id OR s.id = sid.store_id
   LIMIT 1
),
inc AS (
  SELECT
    COUNT(*) FILTER (WHERE i.dt >= CURRENT_TIMESTAMP - INTERVAL '90 days')::float8 AS incidents_90d,
    COUNT(*) FILTER (WHERE i.incident_status IN ('Open', 'In Progress', 'Escalated'))::float8 AS open_incident_count,
    COUNT(*) FILTER (
      WHERE i.criticality = 'Critical'
        AND i.incident_status IN ('Open', 'In Progress', 'Escalated')
    )::float8 AS critical_incident_count
    FROM incidents i
    JOIN sid ON i.store_id = sid.store_id
),
stk AS (
  SELECT
    COUNT(*)::float8 AS sku_count,
    COUNT(*) FILTER (WHERE f.health_level = 'critical')::float8 AS stock_critical_count,
    COUNT(*) FILTER (WHERE f.health_level = 'understock')::float8 AS stock_understock_count,
    COALESCE(AVG(f.days_of_cover), 0)::float8 AS days_of_cover,
    COALESCE(SUM(f.reorder_qty), 0)::float8 AS reorder_qty,
    COALESCE(SUM(f.order_value), 0)::float8 AS order_value
    FROM stock_reorder_fact f
    JOIN sid ON f.store_id = sid.store_id
),
inv AS (
  SELECT COALESCE(
           (CURRENT_DATE - MAX(inv.date)::date),
           365
         )::float8 AS days_since_audit
    FROM inventories inv
    JOIN sid ON inv.store_id = sid.store_id
)
SELECT
  COALESCE(st.store_id, sid.store_id)::text AS store_id,
  COALESCE(st.store_name, 'store ' || sid.store_id::text) AS name,
  COALESCE(NULLIF(split_part(st.address, ',', 1), ''), '') AS city,
  COALESCE(st.region_id::text, '') AS region,
  COALESCE(LEAST(10, 10.0 * COALESCE(stk.stock_critical_count, 0) / NULLIF(stk.sku_count, 0)), 0) AS shrink_pct,
  COALESCE(inc.incidents_90d, 0) AS incidents_90d,
  COALESCE(inv.days_since_audit, 365) AS days_since_audit,
  CASE
    WHEN COALESCE(st.store_area, 0) >= 1500 THEN 5
    WHEN COALESCE(st.store_area, 0) >= 1000 THEN 4
    WHEN COALESCE(st.store_area, 0) >= 600 THEN 3
    WHEN COALESCE(st.store_area, 0) >= 300 THEN 2
    ELSE 1
  END::float8 AS revenue_impact,
  GREATEST(0.10, LEAST(0.90,
    1.0 - LEAST(1.0,
      COALESCE(inc.critical_incident_count, 0) * 0.12
      + COALESCE(inc.open_incident_count, 0) * 0.04
      + COALESCE(stk.stock_critical_count, 0) * 0.01
    )
  )) AS control_effectiveness,
  LEAST(5, GREATEST(1, 1 + ROUND(COALESCE(inc.incidents_90d, 0) / 4.0)))::float8 AS likelihood,
  CASE
    WHEN COALESCE(st.store_area, 0) >= 1500 THEN 5
    WHEN COALESCE(st.store_area, 0) >= 1000 THEN 4
    WHEN COALESCE(st.store_area, 0) >= 600 THEN 3
    WHEN COALESCE(st.store_area, 0) >= 300 THEN 2
    ELSE 1
  END::float8 AS impact,
  COALESCE(inc.open_incident_count, 0) AS open_incident_count,
  COALESCE(inc.critical_incident_count, 0) AS critical_incident_count,
  COALESCE(stk.stock_critical_count, 0) AS stock_critical_count,
  COALESCE(stk.stock_understock_count, 0) AS stock_understock_count,
  COALESCE(stk.sku_count, 0) AS sku_count,
  COALESCE(stk.days_of_cover, 0) AS days_of_cover,
  COALESCE(stk.reorder_qty, 0) AS reorder_qty,
  COALESCE(stk.order_value, 0) AS order_value
FROM sid
LEFT JOIN st ON TRUE
LEFT JOIN inc ON TRUE
LEFT JOIN stk ON TRUE
LEFT JOIN inv ON TRUE
`, sid)
	if err != nil {
		return nil, fmt.Errorf("gather store risk: %w", err)
	}
	return &in, nil
}

// StoreRiskInputMap flattens the bag for the rulesgo engine.
func StoreRiskInputMap(in *StoreRiskInput) map[string]interface{} {
	if in == nil {
		return map[string]interface{}{}
	}
	return map[string]interface{}{
		"store_id":              in.StoreID,
		"name":                  in.Name,
		"city":                  in.City,
		"region":                in.Region,
		"shrinkPct":             in.ShrinkPct,
		"incidents90d":          in.Incidents90d,
		"daysSinceAudit":        in.DaysSinceAudit,
		"revenueImpact":         in.RevenueImpact,
		"controlEffectiveness":  in.ControlEffectiveness,
		"likelihood":            in.Likelihood,
		"impact":                in.Impact,
		"openIncidentCount":     in.OpenIncidentCount,
		"criticalIncidentCount": in.CriticalIncidentCount,
		"stockCriticalCount":    in.StockCriticalCount,
		"stockUnderstockCount":  in.StockUnderstockCount,
		"skuCount":              in.SKUCount,
		"daysOfCover":           in.DaysOfCover,
		"reorderQty":            in.ReorderQty,
		"orderValue":            in.OrderValue,
	}
}

// PersistStoreRiskSlice upserts risk_slice for slice_kind=store.
func PersistStoreRiskSlice(ctx context.Context, in *StoreRiskInput, residual float64) error {
	if in == nil {
		return fmt.Errorf("store risk input is required")
	}
	db, err := incidentDB(ctx)
	if err != nil {
		return err
	}

	res, err := db.ExecContext(ctx, `
UPDATE risk_slice SET
  slice_label = $2,
  risk_score = $3,
  incident_count = $4,
  open_incident_count = $5,
  critical_incident_count = $6,
  escalated_count = 0,
  stock_critical_count = $7,
  stock_understock_count = $8,
  stock_risk_count = $7 + $8,
  store_at_risk_count = CASE WHEN $3 >= 40 THEN 1 ELSE 0 END,
  sku_count = $9,
  days_of_cover = $10,
  reorder_qty = $11,
  order_value = $12
 WHERE slice_kind = 'store' AND slice_key = $1
`, in.StoreID, in.Name, residual,
		in.Incidents90d, in.OpenIncidentCount, in.CriticalIncidentCount,
		in.StockCriticalCount, in.StockUnderstockCount,
		in.SKUCount, in.DaysOfCover, in.ReorderQty, in.OrderValue)
	if err != nil {
		return fmt.Errorf("update risk_slice: %w", err)
	}
	n, _ := res.RowsAffected()
	if n > 0 {
		return nil
	}

	_, err = db.ExecContext(ctx, `
INSERT INTO risk_slice (
  id, slice_kind, slice_key, slice_label, risk_score,
  incident_count, open_incident_count, critical_incident_count, escalated_count,
  stock_critical_count, stock_understock_count, stock_risk_count, store_at_risk_count,
  sku_count, days_of_cover, reorder_qty, order_value
)
SELECT
  COALESCE((SELECT MAX(id) FROM risk_slice), 0) + 1,
  'store', $1, $2, $3,
  $4, $5, $6, 0,
  $7, $8, $7 + $8, CASE WHEN $3 >= 40 THEN 1 ELSE 0 END,
  $9, $10, $11, $12
`, in.StoreID, in.Name, residual,
		in.Incidents90d, in.OpenIncidentCount, in.CriticalIncidentCount,
		in.StockCriticalCount, in.StockUnderstockCount,
		in.SKUCount, in.DaysOfCover, in.ReorderQty, in.OrderValue)
	if err != nil {
		return fmt.Errorf("insert risk_slice: %w", err)
	}
	return nil
}

// ContextStoreID picks store_id from a flattened trigger context.
func ContextStoreID(ctx map[string]interface{}) string {
	for _, k := range []string{"store_id", "storeID"} {
		if v, ok := ctx[k]; ok {
			return stringifyID(v)
		}
	}
	return ""
}

func stringifyID(v interface{}) string {
	if v == nil {
		return ""
	}
	var s string
	switch t := v.(type) {
	case string:
		s = strings.TrimSpace(t)
	case float64:
		if t == float64(int64(t)) {
			s = strconv.FormatInt(int64(t), 10)
		} else {
			s = strconv.FormatFloat(t, 'f', -1, 64)
		}
	case json.Number:
		s = t.String()
	case int64:
		s = strconv.FormatInt(t, 10)
	case int:
		s = strconv.Itoa(t)
	case uint64:
		s = strconv.FormatUint(t, 10)
	case []interface{}:
		if len(t) == 0 {
			return ""
		}
		return stringifyID(t[0])
	default:
		s = strings.TrimSpace(fmt.Sprint(v))
	}
	if s == "" || s == "0" || s == "<nil>" || s == "undefined" || s == "null" {
		return ""
	}
	return s
}
