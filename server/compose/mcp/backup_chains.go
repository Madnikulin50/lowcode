package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/madnikulin50/lowcode/server/compose/service"
	"github.com/madnikulin50/lowcode/server/pkg/rulesgo"
	"github.com/madnikulin50/lowcode/server/store"
)

func registerBackupChains(engine *rulesgo.EngineWithPersistence) {
	if engine == nil || engine.Engine == nil {
		return
	}
	ctx := context.Background()
	if service.DefaultStore == nil {
		log.Printf("[bridge] skip Backup chains: store not ready")
		return
	}
	ns, err := store.LookupComposeNamespaceBySlug(ctx, service.DefaultStore, "backup")
	if err != nil || ns == nil {
		log.Printf("[bridge] skip Backup chains: namespace backup not found: %v", err)
		return
	}
	mod := func(handle string) uint64 {
		m, err := store.LookupComposeModuleByNamespaceIDHandle(ctx, service.DefaultStore, ns.ID, handle)
		if err != nil || m == nil {
			log.Printf("[bridge] Backup module %s: %v", handle, err)
			return 0
		}
		return m.ID
	}
	jobs := mod("jobs")
	restores := mod("restores")
	agentURL := strings.TrimRight(os.Getenv("BACKUP_AGENT_URL"), "/")
	if agentURL == "" {
		agentURL = "http://localhost:8087/api"
	}
	for _, c := range backupRuleChains(ns.ID, jobs, restores, agentURL) {
		engine.RegisterChain(c)
		log.Printf("[bridge] registered Backup chain: %s", c.ID)
	}
}

func backupComponentConfig(agentURL string, fields map[string]string) json.RawMessage {
	m := map[string]string{}
	for k, v := range fields {
		if strings.TrimSpace(v) != "" {
			m[k] = v
		}
	}
	if agentURL != "" {
		m["url"] = agentURL
	}
	b, _ := json.Marshal(m)
	return b
}

func backupRuleChains(nsID, jobs, restores uint64, agentURL string) []*rulesgo.Chain {
	ns := fmt.Sprintf("%d", nsID)
	return []*rulesgo.Chain{
		{
			ID:          "backup-run-source",
			NamespaceID: nsID,
			Name:        "Backup: запуск с источника",
			Description: "Create jobs row and call backup/run on the agent.",
			EntryNode:   "record",
			Nodes: []rulesgo.ChainNode{
				{
					ID:     "record",
					Type:   "crud",
					Label:  "Create job",
					Config: jsonRaw(fmt.Sprintf(`{"operation":"create","namespaceID":"%s","moduleID":"%d","moduleHandle":"jobs","fields":{"source":"{{recordID}}","status":"running","progress":"0","kind":"full"}}`, ns, jobs)),
				},
				{
					ID:    "run",
					Type:  "backup/run",
					Label: "Start backup",
					Config: backupComponentConfig(agentURL, map[string]string{
						"sourceID": "{{sourceID}}",
						"source":   "{{recordID}}",
						"jobID":    "{{createdRecordID}}",
					}),
				},
			},
			Edges: []rulesgo.ChainEdge{{From: "record", To: "run"}},
		},
		{
			ID:          "backup-run-policy",
			NamespaceID: nsID,
			Name:        "Backup: запуск по политике",
			Description: "Create jobs row and call backup/run with policyID.",
			EntryNode:   "record",
			Nodes: []rulesgo.ChainNode{
				{
					ID:     "record",
					Type:   "crud",
					Label:  "Create job",
					Config: jsonRaw(fmt.Sprintf(`{"operation":"create","namespaceID":"%s","moduleID":"%d","moduleHandle":"jobs","fields":{"policy":"{{recordID}}","source":"{{source}}","status":"running","progress":"0"}}`, ns, jobs)),
				},
				{
					ID:    "run",
					Type:  "backup/run",
					Label: "Start backup",
					Config: backupComponentConfig(agentURL, map[string]string{
						"policyID": "{{policyID}}",
						"sourceID": "{{source}}",
						"jobID":    "{{createdRecordID}}",
					}),
				},
			},
			Edges: []rulesgo.ChainEdge{{From: "record", To: "run"}},
		},
		{
			ID:          "backup-run-due",
			NamespaceID: nsID,
			Name:        "Backup: запустить due-политики",
			Description: "Периодически запускайте эту цепочку (cron/systemd timer/Corteza Automation) с {\"token\":\"...\"} в теле. " +
				"backup/due только читает у агента список политик, чей cron сейчас подошёл и у которых нет активного running-джоба " +
				"— сам агент ничего не пишет в Compose. Для каждой due-политики цепочка сама создаёт запись jobs (как backup-run-policy) " +
				"и запускает backup/run — дальше статус ведёт тот же поллер/ingest, что и у ручного запуска, так что упавший между " +
				"перезапусками агент-процесс не оставляет джобы висеть в running.",
			EntryNode: "due",
			Nodes: []rulesgo.ChainNode{
				{
					ID:     "due",
					Type:   "backup/due",
					Label:  "List due policies",
					Config: backupComponentConfig(agentURL, nil),
				},
				{
					ID:     "loop",
					Type:   "foreach",
					Label:  "For each due policy",
					Config: jsonRaw(`{"items":"due","itemVar":"policy"}`),
				},
				{
					ID:     "record",
					Type:   "crud",
					Label:  "Create job",
					Config: jsonRaw(fmt.Sprintf(`{"operation":"create","namespaceID":"%s","moduleID":"%d","moduleHandle":"jobs","fields":{"policy":"{{policy.id}}","source":"{{policy.sourceID}}","status":"running","progress":"0"}}`, ns, jobs)),
				},
				{
					ID:    "run",
					Type:  "backup/run",
					Label: "Start backup",
					Config: backupComponentConfig(agentURL, map[string]string{
						"policyID": "{{policy.id}}",
						"sourceID": "{{policy.sourceID}}",
						"jobID":    "{{createdRecordID}}",
					}),
				},
			},
			Edges: []rulesgo.ChainEdge{
				{From: "due", To: "loop"},
				{From: "loop", To: "record"},
				{From: "loop", To: "run"},
			},
		},
		{
			ID:          "backup-restore",
			NamespaceID: nsID,
			Name:        "Backup: восстановить снапшот",
			Description: "Create restores row and call backup/restore.",
			EntryNode:   "record",
			Nodes: []rulesgo.ChainNode{
				{
					ID:     "record",
					Type:   "crud",
					Label:  "Create restore",
					Config: jsonRaw(fmt.Sprintf(`{"operation":"create","namespaceID":"%s","moduleID":"%d","moduleHandle":"restores","fields":{"snapshot":"{{recordID}}","dest_type":"path","dest_path":"{{destPath}}","status":"running","progress":"0"}}`, ns, restores)),
				},
				{
					ID:    "run",
					Type:  "backup/restore",
					Label: "Restore",
					Config: backupComponentConfig(agentURL, map[string]string{
						"snapshotID": "{{snapshotID}}",
						"restoreID":  "{{createdRecordID}}",
						"destType":   "{{destType}}",
						"destPath":   "{{destPath}}",
					}),
				},
			},
			Edges: []rulesgo.ChainEdge{{From: "record", To: "run"}},
		},
		{
			ID:          "backup-prune",
			NamespaceID: nsID,
			Name:        "Backup: prune по retention",
			Description: "backup/prune. Из политики передаёт policyID.",
			EntryNode:   "run",
			Nodes: []rulesgo.ChainNode{
				{
					ID:    "run",
					Type:  "backup/prune",
					Label: "Prune",
					Config: backupComponentConfig(agentURL, map[string]string{
						"policyID": "{{policyID}}",
					}),
				},
			},
		},
		{
			ID:          "backup-ingest-job",
			NamespaceID: nsID,
			Name:        "Backup: ingest статуса джоба",
			Description: "Callback/poll updates the jobs row.",
			EntryNode:   "update_job",
			Nodes: []rulesgo.ChainNode{
				{
					ID:     "update_job",
					Type:   "crud",
					Label:  "Update job",
					Config: jsonRaw(fmt.Sprintf(`{"operation":"update","namespaceID":"%s","moduleID":"%d","moduleHandle":"jobs","recordID":"{{createdRecordID}}","omitEmpty":true,"continueOnError":true,"fields":{"status":"{{status}}","progress":"{{progress}}","bytes_read":"{{bytesRead}}","bytes_written":"{{bytesWritten}}","files_count":"{{files}}","error":"{{error}}","message":"{{message}}","engine":"{{engine}}"}}`, ns, jobs)),
				},
			},
		},
		{
			ID:          "backup-ingest-restore",
			NamespaceID: nsID,
			Name:        "Backup: ingest статуса restore",
			Description: "Callback/poll updates the restores row.",
			EntryNode:   "update_restore",
			Nodes: []rulesgo.ChainNode{
				{
					ID:     "update_restore",
					Type:   "crud",
					Label:  "Update restore",
					Config: jsonRaw(fmt.Sprintf(`{"operation":"update","namespaceID":"%s","moduleID":"%d","moduleHandle":"restores","recordID":"{{createdRecordID}}","omitEmpty":true,"continueOnError":true,"fields":{"status":"{{status}}","progress":"{{progress}}","error":"{{error}}","message":"{{message}}"}}`, ns, restores)),
				},
			},
		},
		{
			ID:          "backup-failed-alert",
			NamespaceID: nsID,
			Name:        "Backup: письмо при ошибке",
			Description: "Mail when a job status is failed.",
			EntryNode:   "check",
			Nodes: []rulesgo.ChainNode{
				{ID: "check", Type: "condition", Label: "Failed?", Config: jsonRaw(`{"field":"status","operator":"eq","value":"failed"}`)},
				{ID: "mail", Type: "mail", Label: "Notify", Config: jsonRaw(`{"to":"{{to}}","subject":"[Backup] failed {{source}}","body":"<p>{{error}}</p><p>{{message}}</p>","contentType":"html"}`)},
			},
			Edges: []rulesgo.ChainEdge{{From: "check", To: "mail", Condition: "check_result"}},
		},
	}
}
