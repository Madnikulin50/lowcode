package rulesgo

import "encoding/json"

// DemoStore is a sample retail site used by the store-risk pilot.
type DemoStore struct {
	Name                 string
	City                 string
	Region               string
	ShrinkPct            float64
	Incidents90d         float64
	DaysSinceAudit       float64
	RevenueImpact        float64 // 1–5 business impact
	ControlEffectiveness float64 // 0–1
	Likelihood           float64 // 1–5 for matrix view
	Impact               float64 // 1–5 for matrix view
}

func (s DemoStore) Input() map[string]interface{} {
	return map[string]interface{}{
		"name":                 s.Name,
		"city":                 s.City,
		"region":               s.Region,
		"shrinkPct":            s.ShrinkPct,
		"incidents90d":         s.Incidents90d,
		"daysSinceAudit":       s.DaysSinceAudit,
		"revenueImpact":        s.RevenueImpact,
		"controlEffectiveness": s.ControlEffectiveness,
		"likelihood":           s.Likelihood,
		"impact":               s.Impact,
	}
}

// DemoStores returns a realistic Russian retail portfolio for the risk pilot.
// Field semantics mirror typical «Магазины» ops metrics (shrink, incidents, audit lag, revenue band).
func DemoStores() []DemoStore {
	return []DemoStore{
		{Name: "Москва · Тверская", City: "Москва", Region: "ЦФО", ShrinkPct: 2.1, Incidents90d: 1, DaysSinceAudit: 45, RevenueImpact: 4, ControlEffectiveness: 0.70, Likelihood: 2, Impact: 4},
		{Name: "Москва · Авиапарк", City: "Москва", Region: "ЦФО", ShrinkPct: 5.8, Incidents90d: 8, DaysSinceAudit: 210, RevenueImpact: 5, ControlEffectiveness: 0.30, Likelihood: 4, Impact: 5},
		{Name: "СПб · Невский", City: "Санкт-Петербург", Region: "СЗФО", ShrinkPct: 3.2, Incidents90d: 3, DaysSinceAudit: 90, RevenueImpact: 4, ControlEffectiveness: 0.55, Likelihood: 3, Impact: 4},
		{Name: "Казань · Мега", City: "Казань", Region: "ПФО", ShrinkPct: 1.5, Incidents90d: 0, DaysSinceAudit: 30, RevenueImpact: 3, ControlEffectiveness: 0.80, Likelihood: 1, Impact: 3},
		{Name: "Екатеринбург · Радуга", City: "Екатеринбург", Region: "УФО", ShrinkPct: 4.4, Incidents90d: 5, DaysSinceAudit: 180, RevenueImpact: 3, ControlEffectiveness: 0.40, Likelihood: 3, Impact: 3},
		{Name: "Новосибирск · Галерея", City: "Новосибирск", Region: "СФО", ShrinkPct: 6.2, Incidents90d: 12, DaysSinceAudit: 300, RevenueImpact: 4, ControlEffectiveness: 0.10, Likelihood: 5, Impact: 4},
		{Name: "Краснодар · Красная", City: "Краснодар", Region: "ЮФО", ShrinkPct: 2.8, Incidents90d: 2, DaysSinceAudit: 60, RevenueImpact: 3, ControlEffectiveness: 0.65, Likelihood: 2, Impact: 3},
		{Name: "Самара · Космопорт", City: "Самара", Region: "ПФО", ShrinkPct: 3.9, Incidents90d: 4, DaysSinceAudit: 120, RevenueImpact: 3, ControlEffectiveness: 0.50, Likelihood: 3, Impact: 3},
		{Name: "Воронеж · Центр", City: "Воронеж", Region: "ЦФО", ShrinkPct: 7.1, Incidents90d: 9, DaysSinceAudit: 250, RevenueImpact: 2, ControlEffectiveness: 0.25, Likelihood: 5, Impact: 2},
		{Name: "Калининград · Европа", City: "Калининград", Region: "СЗФО", ShrinkPct: 1.8, Incidents90d: 1, DaysSinceAudit: 40, RevenueImpact: 2, ControlEffectiveness: 0.75, Likelihood: 2, Impact: 2},
	}
}

// DemoStoreRiskChain scores a store: weighted ops factors → residual band → critical mail.
func DemoStoreRiskChain() *Chain {
	return &Chain{
		ID:          "demo_store_risk",
		Name:        "Store Operational Risk",
		Description: "Оценка операционного риска магазина: усадка, инциденты, давность аудита, влияние выручки → residual score/level; critical → уведомление.",
		EntryNode:   "weighted",
		Nodes: []ChainNode{
			{
				ID:    "weighted",
				Type:  "score.weighted",
				Label: "Ops weighted score",
				Config: json.RawMessage(`{
					"factors":[
						{"field":"shrinkPct","weight":0.35,"max":10},
						{"field":"incidents90d","weight":0.25,"max":20},
						{"field":"daysSinceAudit","weight":0.15,"max":365},
						{"field":"revenueImpact","weight":0.25,"max":5}
					]
				}`),
			},
			{
				ID:    "matrix",
				Type:  "score.matrix",
				Label: "L×I matrix (audit trail)",
				Config: json.RawMessage(`{
					"likelihoodField":"likelihood",
					"impactField":"impact",
					"outScore":"matrixScore"
				}`),
			},
			{
				ID:    "band",
				Type:  "risk.band",
				Label: "Residual band",
				Config: json.RawMessage(`{
					"scoreField":"score",
					"controlField":"controlEffectiveness",
					"bands":[
						{"name":"low","max":20},
						{"name":"medium","max":40},
						{"name":"high","max":60},
						{"name":"critical","max":100}
					]
				}`),
			},
			{
				ID:    "alert",
				Type:  "mail",
				Label: "Critical alert",
				Config: json.RawMessage(`{
					"to":"risk@example.com",
					"subject":"Critical store risk: {{name}}",
					"body":"<h2>{{name}} ({{city}})</h2><p>Residual score: <b>{{residualScore}}</b> ({{level}})</p><p>Inherent: {{score}} · Controls: {{controlEffectiveness}}</p>",
					"contentType":"html"
				}`),
			},
		},
		Edges: []ChainEdge{
			{From: "weighted", To: "matrix"},
			{From: "matrix", To: "band"},
			{From: "band", To: "alert", Condition: "is_critical", Label: "critical"},
		},
	}
}
