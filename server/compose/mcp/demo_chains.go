package mcp

import (
	"log"

	"github.com/madnikulin50/lowcode/server/pkg/rulesgo"
)

func registerDemoChains(engine *rulesgo.EngineWithPersistence) {
	chains := []*rulesgo.Chain{
		demoEmailOnRecordCreate(),
		demoLeadScoring(),
		demoDataCleanup(),
		rulesgo.DemoStoreRiskChain(),
		rulesgo.DemoStockHealthChain(),
		rulesgo.DemoAutoReorderChain(),
	}

	for _, c := range chains {
		if engine.Chain(c.ID) == nil {
			engine.RegisterChain(c)
			log.Printf("[bridge] registered demo chain: %s", c.Name)
		}
	}
}

func demoEmailOnRecordCreate() *rulesgo.Chain {
	return &rulesgo.Chain{
		ID:          "demo_welcome_email",
		Name:        "Welcome Email on New Record",
		Description: "Sends a welcome email when a new record is created. Use with a 'create' trigger.",
		EntryNode:   "check_subject",
		Nodes: []rulesgo.ChainNode{
			{
				ID:     "check_subject",
				Type:   "condition",
				Label:  "Has email?",
				Config: jsonRaw(`{"field":"email","operator":"notEmpty"}`),
			},
			{
				ID:    "send_welcome",
				Type:  "mail",
				Label: "Send Welcome",
				Config: jsonRaw(`{
					"to":"{{email}}",
					"subject":"Welcome to Corteza!",
					"body":"<h1>Hi {{name}}</h1><p>Welcome aboard! Your record has been created.</p>",
					"contentType":"html"
				}`),
			},
		},
		Edges: []rulesgo.ChainEdge{
			{From: "check_subject", To: "send_welcome", Condition: "check_subject_result"},
		},
	}
}

func demoLeadScoring() *rulesgo.Chain {
	return &rulesgo.Chain{
		ID:          "demo_lead_scoring",
		Name:        "Lead Scoring",
		Description: "Scores leads based on field values and qualifies hot leads.",
		EntryNode:   "score_lead",
		Nodes: []rulesgo.ChainNode{
			{
				ID:     "score_lead",
				Type:   "condition",
				Label:  "Budget > 10000?",
				Config: jsonRaw(`{"field":"budget","operator":"gt","value":"10000"}`),
			},
			{
				ID:    "qualify_hot",
				Type:  "mail",
				Label: "Notify Sales",
				Config: jsonRaw(`{
					"to":"sales@example.com",
					"subject":"Hot Lead: {{name}}",
					"body":"<h2>New Hot Lead</h2><p>Company: {{company}}</p><p>Budget: {{budget}}</p><p>Contact: {{email}}</p>",
					"contentType":"html"
				}`),
			},
			{
				ID:    "ai_suggest",
				Type:  "ai",
				Label: "AI Follow-up",
				Config: jsonRaw(`{
					"agent":"assistant",
					"prompt":"Suggest a personalized follow-up message for lead {{name}} from {{company}} with budget {{budget}}. Keep it professional and warm."
				}`),
			},
		},
		Edges: []rulesgo.ChainEdge{
			{From: "score_lead", To: "qualify_hot", Condition: "score_lead_result"},
			{From: "qualify_hot", To: "ai_suggest"},
		},
	}
}

func demoDataCleanup() *rulesgo.Chain {
	return &rulesgo.Chain{
		ID:          "demo_data_cleanup",
		Name:        "Data Cleanup Script",
		Description: "Runs a JavaScript snippet to clean and normalize record data.",
		EntryNode:   "cleanup",
		Nodes: []rulesgo.ChainNode{
			{
				ID:    "cleanup",
				Type:  "script",
				Label: "Normalize Data",
				Config: jsonRaw(`{
					"code":"runtime.log.info('Processing record: ' + context.name); var name = (context.name || '').trim().replace(/\\s+/g, ' '); var email = (context.email || '').toLowerCase().trim(); runtime.log.info('Normalized: ' + name + ' / ' + email); '{\"name\":\"' + name + '\",\"email\":\"' + email + '\"}';"
				}`),
			},
		},
		Edges: []rulesgo.ChainEdge{},
	}
}

func jsonRaw(s string) []byte {
	return []byte(s)
}
