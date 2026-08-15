package rulesgo

import "encoding/json"

// DemoStockHealthChain scores a store×SKU fact bag: low DOC / zero stock → risk band.
// Facts are expected precomputed (stock_reorder_fact): days_of_cover, zeroStock, order_value/stock_sum.
func DemoStockHealthChain() *Chain {
	return &Chain{
		ID:          "stock_health_sku",
		Name:        "Stock Health SKU",
		Description: "Оценка здоровья остатка SKU×магазин: DOC (invert), zeroStock, сумма → residual band; critical → уведомление. Политика/спрос считаются вне rulesgo (stock_policy + receipt_positions).",
		EntryNode:   "weighted",
		Nodes: []ChainNode{
			{
				ID:    "weighted",
				Type:  "score.weighted",
				Label: "Stock risk weighted",
				Config: json.RawMessage(`{
					"factors":[
						{"field":"days_of_cover","weight":0.55,"max":90,"invert":true},
						{"field":"zeroStock","weight":0.35,"max":1},
						{"field":"order_value","weight":0.10,"max":100000}
					]
				}`),
			},
			{
				ID:    "band",
				Type:  "risk.band",
				Label: "Health band",
				Config: json.RawMessage(`{
					"scoreField":"score",
					"criticalLevels":["critical"],
					"bands":[
						{"name":"overstock","max":20},
						{"name":"ok","max":45},
						{"name":"understock","max":70},
						{"name":"critical","max":100}
					]
				}`),
			},
			{
				ID:    "alert",
				Type:  "mail",
				Label: "Critical stock alert",
				Config: json.RawMessage(`{
					"to":"inventory@example.com",
					"subject":"Critical stock: {{product_name}} @ store {{store_id}}",
					"body":"<h2>{{product_name}}</h2><p>Store {{store_id}} · DOC {{days_of_cover}} · qty {{stock_quantity}}</p><p>Score {{score}} → {{level}}</p>",
					"contentType":"html"
				}`),
			},
		},
		Edges: []ChainEdge{
			{From: "weighted", To: "band"},
			{From: "band", To: "alert", Condition: "is_critical", Label: "critical"},
		},
	}
}

// DemoAutoReorderChain suggests / creates a purchase_order_line when reorder_qty > 0.
// Seed SQL already materializes submitted POs per store×supplier; this chain is the live path.
// CRUD may return status=crud_service_not_configured when no CRUDService is wired.
func DemoAutoReorderChain() *Chain {
	return &Chain{
		ID:          "auto_reorder_sku",
		Name:        "Auto Reorder SKU",
		Description: "Автозаказ SKU×магазин: condition reorder_qty>0 → priority score → band → crud create purchase_order_line (status implied by submitted parent PO). Demand=receipt_positions; policy=stock_policy; order grain=per store_id.",
		EntryNode:   "need",
		Nodes: []ChainNode{
			{
				ID:     "need",
				Type:   "condition",
				Label:  "Need reorder?",
				Config: json.RawMessage(`{"field":"reorder_qty","operator":"gt","value":"0"}`),
			},
			{
				ID:    "priority",
				Type:  "score.weighted",
				Label: "Reorder priority",
				Config: json.RawMessage(`{
					"factors":[
						{"field":"days_of_cover","weight":0.60,"max":90,"invert":true},
						{"field":"order_value","weight":0.40,"max":100000}
					]
				}`),
			},
			{
				ID:    "band",
				Type:  "risk.band",
				Label: "Priority band",
				Config: json.RawMessage(`{
					"scoreField":"score",
					"criticalLevels":["critical"],
					"bands":[
						{"name":"low","max":30},
						{"name":"medium","max":55},
						{"name":"high","max":75},
						{"name":"critical","max":100}
					]
				}`),
			},
			{
				ID:    "create_line",
				Type:  "crud",
				Label: "Create PO line",
				Config: json.RawMessage(`{
					"operation":"create",
					"moduleID":510020000000100005,
					"namespaceID":495727984893558785,
					"fields":{
						"product_id":"{{product_id}}",
						"ean":"{{ean}}",
						"product_name":"{{product_name}}",
						"store_id":"{{store_id}}",
						"qty_ordered":"{{reorder_qty}}",
						"qty_suggested":"{{reorder_qty}}",
						"unit_cost":"{{unit_cost}}",
						"line_sum":"{{order_value}}",
						"reorder_point":"{{reorder_point}}",
						"days_of_cover":"{{days_of_cover}}",
						"health_level":"{{health_level}}",
						"rule_score":"{{score}}"
					}
				}`),
			},
		},
		Edges: []ChainEdge{
			{From: "need", To: "priority", Condition: "need_result", Label: "reorder_qty>0"},
			{From: "priority", To: "band"},
			{From: "band", To: "create_line"},
		},
	}
}
