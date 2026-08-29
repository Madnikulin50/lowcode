package main

import (
	"context"
	"embed"
	"flag"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/madnikulin50/lowcode/agents/cmdb/agent"
	"github.com/madnikulin50/lowcode/agents/cmdb/api"
	"github.com/madnikulin50/lowcode/agents/sdk"
)

//go:embed web/dist/*
var webFS embed.FS

func main() {
	listen := flag.String("listen", ":8085", "HTTP listen address")
	cortezaAPI := flag.String("api", "http://localhost:3333/api", "Lowcode API base URL")
	token := flag.String("token", "", "Lowcode API token")
	namespaceID := flag.Uint64("namespace", 0, "Default namespace ID")
	llmURL := flag.String("llm-url", "http://localhost:11434", "Ollama base URL")
	llmModel := flag.String("llm-model", "deepseek-v2", "Ollama model for classification")
	mcpAddr := flag.String("mcp", "", "MCP server address (e.g. :9091, or 'stdio')")
	staticDir := flag.String("static", "", "Static files directory (overrides embedded)")
	dbType := flag.String("db", "embedded", "Storage backend: 'embedded' (SQLite) or 'lowcode'")
	dbPath := flag.String("db-path", "cmdb.db", "Path to embedded database file")
	scanInterval := flag.Duration("scan-interval", 20*time.Minute, "Periodic scan interval (e.g. 20m, 1h). 0 = disabled")
	statusInterval := flag.Duration("status-interval", 5*time.Minute, "Device status check interval (e.g. 5m). 0 = disabled")
	autoCIDRs := flag.String("auto-cidrs", "", "Comma-separated CIDRs to auto-scan periodically")
	flag.Parse()

	if *token == "" {
		*token = os.Getenv("TOKEN")
	}

	var autoCIDRsList []string
	if *autoCIDRs != "" {
		for _, c := range strings.Split(*autoCIDRs, ",") {
			c = strings.TrimSpace(c)
			if c != "" {
				autoCIDRsList = append(autoCIDRsList, c)
			}
		}
	}

	cfg := agent.Config{
		CortezaAPI:     *cortezaAPI,
		Token:          *token,
		NamespaceID:    *namespaceID,
		LLMBaseURL:     *llmURL,
		LLMModel:       *llmModel,
		Concurrency:    20,
		PingTimeout:    400 * time.Millisecond,
		ScanInterval:   *scanInterval,
		StatusInterval: *statusInterval,
		AutoCIDRs:      autoCIDRsList,
	}

	var store agent.Storage
	switch *dbType {
	case "lowcode", "corteza":
		if *token == "" {
			log.Fatal("--token is required with --db=lowcode")
		}
		store = agent.NewCortezaStore(*cortezaAPI, *token, *namespaceID)
		log.Printf("Storage: Lowcode API at %s", *cortezaAPI)
	default:
		s, err := agent.NewEmbeddedStore(*dbPath)
		if err != nil {
			log.Fatalf("Cannot open embedded db: %v", err)
		}
		store = s
		log.Printf("Storage: embedded SQLite at %s", *dbPath)
	}

	ag := agent.New(cfg, store)

	if *scanInterval > 0 {
		go ag.StartPeriodicScan(context.Background())
		log.Printf("Periodic scan every %v", *scanInterval)
		if len(autoCIDRsList) > 0 {
			log.Printf("Auto-scan CIDRs: %v", autoCIDRsList)
		}
	}
	if *statusInterval > 0 {
		go ag.StartStatusChecker(context.Background())
		log.Printf("Status check every %v", *statusInterval)
	}

	svc := sdk.New(sdk.Config{
		Handle:      "cmdb",
		Name:        "CMDB Discovery Agent",
		Listen:      *listen,
		PublicURL:   "http://localhost" + *listen,
		Slug:        "cmdb",
		CortezaAPI:  *cortezaAPI,
		Token:       *token,
		NamespaceID: *namespaceID,
	})
	svc.SetBackend(ag)
	svc.Register(agent.Components()...)
	svc.MountAPI(api.New(ag).Mount)
	svc.MountRoot(func(r chi.Router) {
		if *staticDir != "" {
			r.Handle("/*", http.StripPrefix("/", http.FileServer(http.Dir(*staticDir))))
			return
		}
		sub, err := fs.Sub(webFS, "web/dist")
		if err != nil {
			log.Printf("No embedded frontend found (build it with 'make web'): %v", err)
			return
		}
		r.Handle("/*", http.FileServer(http.FS(sub)))
	})

	if *mcpAddr != "" {
		go func() {
			log.Printf("MCP server starting on %s", *mcpAddr)
			if err := agent.StartMCPServer(context.Background(), ag, *mcpAddr); err != nil {
				log.Fatalf("MCP server error: %v", err)
			}
		}()
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log.Printf("CMDB Discovery Agent listening on %s", *listen)
	log.Printf("LLM: %s (%s)", *llmURL, *llmModel)
	if err := svc.Listen(ctx); err != nil {
		log.Fatalf("HTTP server error: %v", err)
	}
}
