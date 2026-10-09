package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/madnikulin50/lowcode/agents/backup/agent"
	"github.com/madnikulin50/lowcode/agents/sdk"
)

func main() {
	log.SetOutput(os.Stdout)
	listen := flag.String("listen", ":8087", "HTTP listen address")
	publicURL := flag.String("public-url", "http://localhost:8087", "URL Compose uses to reach this agent")
	cortezaAPI := flag.String("api", "http://localhost:3333", "Lowcode origin (not /api — APIs live at /compose on the GoLand server)")
	token := flag.String("token", "", "Lowcode API token")
	namespaceID := flag.Uint64("namespace", 0, "Default namespace ID (slug backup if 0)")
	poll := flag.Duration("poll", time.Minute, "Policy poll interval. 0 = disabled")
	minioEndpoint := flag.String("minio-endpoint", env("MINIO_ENDPOINT", "127.0.0.1:9000"), "MinIO endpoint host:port")
	minioAccess := flag.String("minio-access", env("MINIO_ACCESS_KEY", env("MINIO_ROOT_USER", "minioadmin")), "MinIO access key")
	minioSecret := flag.String("minio-secret", env("MINIO_SECRET_KEY", env("MINIO_ROOT_PASSWORD", "minioadmin")), "MinIO secret key")
	minioBucket := flag.String("minio-bucket", env("MINIO_BUCKET", "backups"), "Destination bucket")
	minioSecure := flag.Bool("minio-secure", env("MINIO_SECURE", "") == "true", "HTTPS to MinIO")
	minioRegion := flag.String("minio-region", env("MINIO_REGION", ""), "MinIO region")
	resticPassword := flag.String("restic-password", env("RESTIC_PASSWORD", env("BACKUP_RESTIC_PASSWORD", "")), "Restic repo password")
	flag.Parse()

	if *token == "" {
		*token = sdk.SelfToken(*cortezaAPI)
	}

	cfg := agent.Config{
		ListenAddr:     *listen,
		PublicURL:      strings.TrimRight(*publicURL, "/"),
		CortezaAPI:     *cortezaAPI,
		Token:          *token,
		NamespaceID:    *namespaceID,
		PollInterval:   *poll,
		ResticPassword: *resticPassword,
		Concurrency:    2,
		Minio: agent.MinioConfig{
			Endpoint: *minioEndpoint,
			Access:   *minioAccess,
			Secret:   *minioSecret,
			Bucket:   *minioBucket,
			Secure:   *minioSecure,
			Region:   *minioRegion,
		},
	}

	store, err := agent.NewObjectStore(cfg.Minio)
	if err != nil {
		log.Fatalf("minio: %v", err)
	}
	cz := agent.NewCorteza(cfg.CortezaAPI, cfg.Token, cfg.NamespaceID)
	if err := cz.Discover(context.Background()); err != nil {
		log.Printf("corteza API discover: %v (will retry on first request)", err)
	} else {
		log.Printf("corteza API %s", cz.BaseURL())
	}
	if err := store.Ping(context.Background()); err != nil {
		log.Printf("minio not reachable at %s (%v); agent still listens, backup jobs need MinIO", cfg.Minio.Endpoint, err)
	} else if err := store.EnsureBucket(context.Background()); err != nil {
		log.Printf("minio bucket %s: %v", cfg.Minio.Bucket, err)
	} else {
		log.Printf("minio ok %s bucket=%s", cfg.Minio.Endpoint, cfg.Minio.Bucket)
	}
	ag := agent.New(cfg, store, cz)

	svc := sdk.New(sdk.Config{
		Handle:      "backup",
		Name:        "Backup Agent",
		Listen:      *listen,
		PublicURL:   cfg.PublicURL,
		Slug:        "backup",
		CortezaAPI:  cfg.CortezaAPI,
		Token:       cfg.Token,
		NamespaceID: cfg.NamespaceID,
		Heartbeat:   sdk.HeartbeatConfig{Module: "agents"},
	})
	svc.SetBackend(ag)
	svc.Register(agent.Components()...)
	svc.Alias(http.MethodPost, "/restore", "restore")
	svc.Alias(http.MethodPost, "/prune", "prune")
	svc.Alias(http.MethodPost, "/jobs/due", "due")
	svc.Alias(http.MethodPost, "/jobs/reconcile", "reconcile")
	svc.Sync("due")
	svc.Sync("reconcile")
	svc.MountRoot(func(r chi.Router) {
		r.Handle("/metrics", promhttp.Handler())
	})

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if cfg.Token != "" {
		// This process starts with no in-memory job state, so any "jobs"/
		// "restores" record still marked "running" belongs to whatever
		// process owned it before this restart/deploy — it's never going
		// to finish. Runs once, before we start accepting new job requests.
		// Deployments without a static token (cron polling off, tokens
		// supplied per-request instead) skip this — POST /jobs/reconcile
		// with a bearer token does the same cleanup on demand.
		reconcileCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		if n, err := ag.ReconcileStaleJobs(reconcileCtx, cz); err != nil {
			log.Printf("reconcile on startup: %v", err)
		} else if n > 0 {
			log.Printf("reconcile on startup: %d stale job(s) marked failed", n)
		}
		cancel()
	}

	// StartScheduler is heartbeat-only now (see its doc comment) — policy
	// scheduling moved to the server-side "backup-run-due" rule chain, so
	// there's no PollInterval knob to gate this on any more.
	if cfg.Token != "" {
		go ag.StartScheduler(ctx)
		log.Printf("heartbeat enabled")
	} else {
		log.Printf("heartbeat off (no --token); this agent won't show as online in Compose")
	}

	log.Printf("backup-agent listening on %s (minio=%s bucket=%s api=%s)", *listen, cfg.Minio.Endpoint, cfg.Minio.Bucket, cfg.CortezaAPI)
	if err := svc.Listen(ctx); err != nil {
		log.Fatalf("HTTP: %v", err)
	}
}

func env(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}
