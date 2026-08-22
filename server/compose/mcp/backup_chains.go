package mcp

import (
	"context"
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

func backupRuleChains(nsID, jobs, restores uint64, agentURL string) []*rulesgo.Chain {
	ns := fmt.Sprintf("%d", nsID)
	return []*rulesgo.Chain{
		{
			ID:          "backup-run-source",
			NamespaceID: nsID,
			Name:        "Backup: запуск с источника",
			Description: "Create jobs row and POST sourceID to the backup agent.",
			EntryNode:   "record",
			Nodes: []rulesgo.ChainNode{
				{
					ID:     "record",
					Type:   "crud",
					Label:  "Create job",
					Config: jsonRaw(fmt.Sprintf(`{"operation":"create","namespaceID":"%s","moduleID":"%d","moduleHandle":"jobs","fields":{"source":"{{recordID}}","status":"running","progress":"0","kind":"full"}}`, ns, jobs)),
				},
				{
					ID:     "http",
					Type:   "http",
					Label:  "Start backup",
					Config: jsonRaw(fmt.Sprintf(`{"url":%q,"method":"POST","body":"{\"sourceID\":\"{{recordID}}\",\"jobID\":\"{{createdRecordID}}\",\"namespaceID\":\"{{namespaceID}}\",\"token\":\"{{authToken}}\",\"callbackUrl\":\"{{callbackUrl}}\"}","timeout":30}`, agentURL+"/jobs")),
				},
				{
					ID:     "detach_poll",
					Type:   "detach",
					Label:  "Poll",
					Config: jsonRaw(`{"kind":"poll","ingestChainID":"backup-ingest-job","statusUrl":"{{agentUrl}}/jobs/{{scanID}}","interval":2,"timeout":3600,"until":"completed,failed,done,error"}`),
				},
			},
			Edges: []rulesgo.ChainEdge{
				{From: "record", To: "http"},
				{From: "http", To: "detach_poll"},
			},
		},
		{
			ID:          "backup-run-policy",
			NamespaceID: nsID,
			Name:        "Backup: запуск по политике",
			Description: "Create jobs row and POST policyID to the backup agent.",
			EntryNode:   "record",
			Nodes: []rulesgo.ChainNode{
				{
					ID:     "record",
					Type:   "crud",
					Label:  "Create job",
					Config: jsonRaw(fmt.Sprintf(`{"operation":"create","namespaceID":"%s","moduleID":"%d","moduleHandle":"jobs","fields":{"policy":"{{recordID}}","source":"{{source}}","status":"running","progress":"0"}}`, ns, jobs)),
				},
				{
					ID:     "http",
					Type:   "http",
					Label:  "Start backup",
					Config: jsonRaw(fmt.Sprintf(`{"url":%q,"method":"POST","body":"{\"policyID\":\"{{recordID}}\",\"sourceID\":\"{{source}}\",\"jobID\":\"{{createdRecordID}}\",\"namespaceID\":\"{{namespaceID}}\",\"token\":\"{{authToken}}\",\"callbackUrl\":\"{{callbackUrl}}\"}","timeout":30}`, agentURL+"/jobs")),
				},
			},
			Edges: []rulesgo.ChainEdge{{From: "record", To: "http"}},
		},
		{
			ID:          "backup-run-due",
			NamespaceID: nsID,
			Name:        "Backup: запустить due-политики",
			Description: "POST /jobs/due.",
			EntryNode:   "http",
			Nodes: []rulesgo.ChainNode{
				{
					ID:     "http",
					Type:   "http",
					Label:  "Run due",
					Config: jsonRaw(fmt.Sprintf(`{"url":%q,"method":"POST","body":"{\"namespaceID\":\"{{namespaceID}}\",\"token\":\"{{authToken}}\"}","timeout":60}`, agentURL+"/jobs/due")),
				},
			},
		},
		{
			ID:          "backup-restore",
			NamespaceID: nsID,
			Name:        "Backup: восстановить снапшот",
			Description: "Create restores row and POST /restore.",
			EntryNode:   "record",
			Nodes: []rulesgo.ChainNode{
				{
					ID:     "record",
					Type:   "crud",
					Label:  "Create restore",
					Config: jsonRaw(fmt.Sprintf(`{"operation":"create","namespaceID":"%s","moduleID":"%d","moduleHandle":"restores","fields":{"snapshot":"{{recordID}}","dest_type":"path","dest_path":"{{destPath}}","status":"running","progress":"0"}}`, ns, restores)),
				},
				{
					ID:     "http",
					Type:   "http",
					Label:  "Restore",
					Config: jsonRaw(fmt.Sprintf(`{"url":%q,"method":"POST","body":"{\"snapshotID\":\"{{recordID}}\",\"restoreID\":\"{{createdRecordID}}\",\"destType\":\"{{destType}}\",\"destPath\":\"{{destPath}}\",\"namespaceID\":\"{{namespaceID}}\",\"token\":\"{{authToken}}\",\"callbackUrl\":\"{{callbackUrl}}\"}","timeout":30}`, agentURL+"/restore")),
				},
			},
			Edges: []rulesgo.ChainEdge{{From: "record", To: "http"}},
		},
		{
			ID:          "backup-prune",
			NamespaceID: nsID,
			Name:        "Backup: prune по retention",
			Description: "POST /prune.",
			EntryNode:   "http",
			Nodes: []rulesgo.ChainNode{
				{
					ID:     "http",
					Type:   "http",
					Label:  "Prune",
					Config: jsonRaw(fmt.Sprintf(`{"url":%q,"method":"POST","body":"{\"policyID\":\"{{recordID}}\",\"namespaceID\":\"{{namespaceID}}\",\"token\":\"{{authToken}}\"}","timeout":60}`, agentURL+"/prune")),
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
