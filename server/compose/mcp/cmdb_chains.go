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

func registerCMDBChains(engine *rulesgo.EngineWithPersistence) {
	if engine == nil || engine.Engine == nil {
		return
	}
	ctx := context.Background()
	if service.DefaultStore == nil {
		log.Printf("[bridge] skip CMDB chains: store not ready")
		return
	}
	ns, err := store.LookupComposeNamespaceBySlug(ctx, service.DefaultStore, "cmdb")
	if err != nil || ns == nil {
		log.Printf("[bridge] skip CMDB chains: namespace cmdb not found: %v", err)
		return
	}
	mod := func(handle string) uint64 {
		m, err := store.LookupComposeModuleByNamespaceIDHandle(ctx, service.DefaultStore, ns.ID, handle)
		if err != nil || m == nil {
			log.Printf("[bridge] CMDB module %s: %v", handle, err)
			return 0
		}
		return m.ID
	}
	devices := mod("devices")
	services := mod("services")
	vulns := mod("vulnerabilities")
	scans := mod("scans")
	agentURL := strings.TrimRight(os.Getenv("CMDB_AGENT_URL"), "/")
	if agentURL == "" {
		agentURL = "http://localhost:8085/api"
	}

	for _, c := range cmdbRuleChains(ns.ID, devices, services, vulns, scans, agentURL) {
		engine.RegisterChain(c)
		log.Printf("[bridge] registered CMDB chain: %s", c.ID)
	}
}

func ensureChainAvailable(ctx context.Context, engine *rulesgo.EngineWithPersistence, chainID string) {
	if engine == nil || engine.Engine == nil || chainID == "" {
		return
	}
	if engine.Chain(chainID) != nil {
		return
	}
	if err := engine.LoadFromStore(ctx); err != nil {
		log.Printf("[bridge] reload rule chains: %v", err)
	}
	if engine.Chain(chainID) != nil {
		return
	}
	registerCMDBChains(engine)
}

func cmdbRuleChains(nsID, devices, services, vulns, scans uint64, agentURL string) []*rulesgo.Chain {
	return []*rulesgo.Chain{
		{
			ID:          "cmdb-trigger-scan",
			NamespaceID: nsID,
			Name:        "CMDB: trigger network scan",
			Description: "POST cidr to the CMDB agent and create a scans row.",
			EntryNode:   "record_scan",
			Nodes: []rulesgo.ChainNode{
				{
					ID:    "record_scan",
					Type:  "crud",
					Label: "Create scan record",
					Config: jsonRaw(fmt.Sprintf(`{
						"operation":"create",
						"namespaceID":"%d",
						"moduleID":"%d",
						"moduleHandle":"scans",
						"fields":{
							"target":"{{cidr}}",
							"network":"{{recordID}}",
							"status":"running",
							"progress":"0",
							"found":"0"
						}
					}`, nsID, scans)),
				},
				{
					ID:    "http_scan",
					Type:  "http",
					Label: "Start agent scan",
					Config: jsonRaw(fmt.Sprintf(`{
						"url":%q,
						"method":"POST",
						"body":"{\"cidr\":\"{{cidr}}\",\"namespaceID\":\"{{namespaceID}}\",\"token\":\"{{authToken}}\",\"scanRecordID\":\"{{createdRecordID}}\",\"callbackUrl\":\"{{callbackUrl}}\"}",
						"timeout":30
					}`, agentURL+"/scan")),
				},
				{
					ID:    "detach_poll",
					Type:  "detach",
					Label: "Poll agent if no webhook",
					Config: jsonRaw(`{
						"kind":"poll",
						"ingestChainID":"cmdb-ingest-scan",
						"statusUrl":"{{agentUrl}}/scans/{{scanID}}",
						"itemsUrl":"{{agentUrl}}/devices",
						"interval":2,
						"timeout":900,
						"until":"done,error,completed,failed"
					}`),
				},
			},
			Edges: []rulesgo.ChainEdge{
				{From: "record_scan", To: "http_scan"},
				{From: "http_scan", To: "detach_poll"},
			},
		},
		cmdbIngestChain(nsID, devices, services, vulns, scans),
		{
			ID:          "cmdb-high-vuln-alert",
			NamespaceID: nsID,
			Name:        "CMDB: high / critical vulnerability alert",
			Description: "Mail when the finding severity is HIGH or CRITICAL.",
			EntryNode:   "fork",
			Nodes: []rulesgo.ChainNode{
				{ID: "fork", Type: "fork", Label: "HIGH or CRITICAL", Config: jsonRaw(`{"branches":2}`)},
				{ID: "is_high", Type: "condition", Label: "Severity HIGH", Config: jsonRaw(`{"field":"severity","operator":"eq","value":"HIGH"}`)},
				{ID: "is_crit", Type: "condition", Label: "Severity CRITICAL", Config: jsonRaw(`{"field":"severity","operator":"eq","value":"CRITICAL"}`)},
				{ID: "mail", Type: "mail", Label: "Notify", Config: jsonRaw(`{"to":"{{to}}","subject":"[CMDB] {{severity}} {{name}}","body":"<p>Severity: <b>{{severity}}</b></p><p>{{name}}</p><p>Device: {{device}}</p><p>CVE: {{cve}}</p>","contentType":"html"}`)},
			},
			Edges: []rulesgo.ChainEdge{
				{From: "fork", To: "is_high"},
				{From: "fork", To: "is_crit"},
				{From: "is_high", To: "mail", Condition: "is_high_result"},
				{From: "is_crit", To: "mail", Condition: "is_crit_result"},
			},
		},
		{
			ID:          "cmdb-insecure-service",
			NamespaceID: nsID,
			Name:        "CMDB: flag insecure service",
			Description: "Create an open HIGH finding when the service name looks like Telnet or FTP.",
			EntryNode:   "fork",
			Nodes: []rulesgo.ChainNode{
				{ID: "fork", Type: "fork", Label: "Telnet or FTP", Config: jsonRaw(`{"branches":2}`)},
				{ID: "is_telnet", Type: "condition", Label: "Telnet", Config: jsonRaw(`{"field":"service","operator":"contains","value":"telnet"}`)},
				{ID: "is_ftp", Type: "condition", Label: "FTP", Config: jsonRaw(`{"field":"service","operator":"contains","value":"ftp"}`)},
				{
					ID:    "create_finding",
					Type:  "crud",
					Label: "Create finding",
					Config: jsonRaw(fmt.Sprintf(`{
						"operation":"create","namespaceID":"%d","moduleID":"%d","moduleHandle":"vulnerabilities",
						"fields":{"device":"{{device}}","name":"Insecure service: {{service}}","severity":"HIGH","description":"{{service}} on port {{port}} transmits data in cleartext.","remediation":"Disable the service and use an encrypted alternative (SSH/SFTP).","status":"open"}
					}`, nsID, vulns)),
				},
			},
			Edges: []rulesgo.ChainEdge{
				{From: "fork", To: "is_telnet"},
				{From: "fork", To: "is_ftp"},
				{From: "is_telnet", To: "create_finding", Condition: "is_telnet_result"},
				{From: "is_ftp", To: "create_finding", Condition: "is_ftp_result"},
			},
		},
		{
			ID:          "cmdb-flag-insecure-ports",
			NamespaceID: nsID,
			Name:        "CMDB: flag insecure ports on a device",
			Description: "Create a HIGH finding if open_ports/services mention Telnet (23) or FTP (21).",
			EntryNode:   "fork",
			Nodes: []rulesgo.ChainNode{
				{ID: "fork", Type: "fork", Label: "Port 21/23 or telnet/ftp", Config: jsonRaw(`{"branches":4}`)},
				{ID: "port_23", Type: "condition", Label: "TCP 23", Config: jsonRaw(`{"field":"open_ports","operator":"contains","value":"\"port\":23,"}`)},
				{ID: "port_21", Type: "condition", Label: "TCP 21", Config: jsonRaw(`{"field":"open_ports","operator":"contains","value":"\"port\":21,"}`)},
				{ID: "svc_telnet", Type: "condition", Label: "Telnet in services", Config: jsonRaw(`{"field":"services","operator":"contains","value":"telnet"}`)},
				{ID: "svc_ftp", Type: "condition", Label: "FTP in services", Config: jsonRaw(`{"field":"services","operator":"contains","value":"ftp"}`)},
				{
					ID:    "create_finding",
					Type:  "crud",
					Label: "Create finding",
					Config: jsonRaw(fmt.Sprintf(`{
						"operation":"create","namespaceID":"%d","moduleID":"%d","moduleHandle":"vulnerabilities",
						"fields":{"device":"{{recordID}}","name":"Insecure cleartext service on {{hostname}}","severity":"HIGH","description":"Host {{ip_address}} exposes Telnet and/or FTP (ports 21/23).","remediation":"Disable the service and use an encrypted alternative (SSH/SFTP).","status":"open"}
					}`, nsID, vulns)),
				},
			},
			Edges: []rulesgo.ChainEdge{
				{From: "fork", To: "port_23"},
				{From: "fork", To: "port_21"},
				{From: "fork", To: "svc_telnet"},
				{From: "fork", To: "svc_ftp"},
				{From: "port_23", To: "create_finding", Condition: "port_23_result"},
				{From: "port_21", To: "create_finding", Condition: "port_21_result"},
				{From: "svc_telnet", To: "create_finding", Condition: "svc_telnet_result"},
				{From: "svc_ftp", To: "create_finding", Condition: "svc_ftp_result"},
			},
		},
		{
			ID:          "cmdb-stale-devices",
			NamespaceID: nsID,
			Name:        "CMDB: list stale online devices",
			Description: "Search devices. Pass query (Corteza QL) in block context.",
			EntryNode:   "search",
			Nodes: []rulesgo.ChainNode{
				{
					ID:     "search",
					Type:   "crud",
					Label:  "Search devices",
					Config: jsonRaw(fmt.Sprintf(`{"operation":"search","namespaceID":"%d","moduleID":"%d","moduleHandle":"devices","query":"{{query}}","limit":100}`, nsID, devices)),
				},
			},
		},
		{
			ID:          "cmdb-ack-finding",
			NamespaceID: nsID,
			Name:        "CMDB: acknowledge finding",
			Description: "Set the current vulnerability record status to acknowledged.",
			EntryNode:   "upd",
			Nodes: []rulesgo.ChainNode{
				{
					ID:     "upd",
					Type:   "crud",
					Label:  "Acknowledge",
					Config: jsonRaw(fmt.Sprintf(`{"operation":"update","namespaceID":"%d","moduleID":"%d","moduleHandle":"vulnerabilities","recordID":"{{recordID}}","fields":{"status":"acknowledged"}}`, nsID, vulns)),
				},
			},
		},
		{
			ID:          "cmdb-close-finding",
			NamespaceID: nsID,
			Name:        "CMDB: mark finding fixed",
			Description: "Set the current vulnerability record status to fixed.",
			EntryNode:   "upd",
			Nodes: []rulesgo.ChainNode{
				{
					ID:     "upd",
					Type:   "crud",
					Label:  "Mark fixed",
					Config: jsonRaw(fmt.Sprintf(`{"operation":"update","namespaceID":"%d","moduleID":"%d","moduleHandle":"vulnerabilities","recordID":"{{recordID}}","fields":{"status":"fixed"}}`, nsID, vulns)),
				},
			},
		},
	}
}

func cmdbIngestChain(nsID, devices, services, vulns, scans uint64) *rulesgo.Chain {
	return &rulesgo.Chain{
		ID:          "cmdb-ingest-scan",
		NamespaceID: nsID,
		Name:        "CMDB: ingest agent job",
		Description: "Webhook/poll envelope → update scans row, upsert devices, services and vulnerabilities.",
		EntryNode:   "update_scan",
		Nodes: []rulesgo.ChainNode{
			{
				ID:    "update_scan",
				Type:  "crud",
				Label: "Update scan record",
				Config: jsonRaw(fmt.Sprintf(`{
					"operation":"update",
					"namespaceID":"%d",
					"moduleID":"%d",
					"moduleHandle":"scans",
					"recordID":"{{scanRecordID}}",
					"omitEmpty":true,
					"continueOnError":true,
					"fields":{
						"status":"{{status}}",
						"progress":"{{progress}}",
						"found":"{{found}}",
						"error":"{{error}}",
						"scanning_ip":"{{scanningIP}}",
						"target":"{{target}}",
						"started_at":"{{startedAt}}",
						"finished_at":"{{finishedAt}}"
					}
				}`, nsID, scans)),
			},
			{
				ID:     "foreach_items",
				Type:   "foreach",
				Label:  "Each device",
				Config: jsonRaw(`{"items":"items","itemVar":"item"}`),
			},
			{
				ID:    "upsert_device",
				Type:  "crud.upsert",
				Label: "Upsert device",
				Config: jsonRaw(fmt.Sprintf(`{
					"namespaceID":"%d",
					"moduleID":"%d",
					"moduleHandle":"devices",
					"matchBy":["mac_address","ip_address","hostname"],
					"omitEmpty":true,
					"resultVar":"deviceRecordID",
					"fields":{
						"ip_address":"{{item.ip}}",
						"mac_address":"{{item.mac}}",
						"hostname":"{{item.hostname}}",
						"vendor":"{{item.vendor}}",
						"device_type":"{{item.deviceType}}",
						"os":"{{item.os}}",
						"domain":"{{item.domain}}",
						"open_ports":"{{item.openPorts}}",
						"services":"{{item.services}}",
						"shares":"{{item.shares}}",
						"vulnerabilities":"{{item.vulnerabilities}}",
						"last_seen":"{{item.lastSeen}}",
						"status":"{{item.status}}"
					}
				}`, nsID, devices)),
			},
			{
				ID:     "foreach_ports",
				Type:   "foreach",
				Label:  "Each open port",
				Config: jsonRaw(`{"items":"item.openPorts","itemVar":"port"}`),
			},
			{
				ID:    "upsert_service",
				Type:  "crud.upsert",
				Label: "Upsert service",
				Config: jsonRaw(fmt.Sprintf(`{
					"namespaceID":"%d",
					"moduleID":"%d",
					"moduleHandle":"services",
					"matchBy":["device","port","proto"],
					"matchAll":true,
					"omitEmpty":true,
					"continueOnError":true,
					"fields":{
						"device":"{{deviceRecordID}}",
						"port":"{{port.port}}",
						"proto":"{{port.proto}}",
						"service":"{{port.service}}",
						"version":"{{port.version}}",
						"banner":"{{port.banner}}"
					}
				}`, nsID, services)),
			},
			{
				ID:     "foreach_vulns",
				Type:   "foreach",
				Label:  "Each vulnerability",
				Config: jsonRaw(`{"items":"item.vulnerabilities","itemVar":"vuln"}`),
			},
			{
				ID:    "upsert_vuln",
				Type:  "crud.upsert",
				Label: "Upsert vulnerability",
				Config: jsonRaw(fmt.Sprintf(`{
					"namespaceID":"%d",
					"moduleID":"%d",
					"moduleHandle":"vulnerabilities",
					"matchBy":["device","name"],
					"matchAll":true,
					"omitEmpty":true,
					"continueOnError":true,
					"fields":{
						"device":"{{deviceRecordID}}",
						"name":"{{vuln.name}}",
						"severity":"{{vuln.severity}}",
						"cve":"{{vuln.cve}}",
						"description":"{{vuln.description}}",
						"remediation":"{{vuln.remediation}}",
						"status":"open",
						"detected_at":"{{item.lastSeen}}"
					}
				}`, nsID, vulns)),
			},
		},
		Edges: []rulesgo.ChainEdge{
			{From: "update_scan", To: "foreach_items"},
			{From: "foreach_items", To: "upsert_device"},
			{From: "foreach_items", To: "foreach_ports"},
			{From: "foreach_ports", To: "upsert_service"},
			{From: "foreach_items", To: "foreach_vulns"},
			{From: "foreach_vulns", To: "upsert_vuln"},
		},
	}
}
