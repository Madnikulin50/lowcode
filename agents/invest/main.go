package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/madnikulin50/lowcode/agents/invest/agent"
	"github.com/madnikulin50/lowcode/agents/invest/api"
)

func main() {
	listen := flag.String("listen", ":8086", "HTTP listen address")
	cortezaAPI := flag.String("api", "http://localhost:3333", "Lowcode origin (not /api — APIs live at /compose on the GoLand server)")
	token := flag.String("token", "", "lowcode API token")
	namespaceID := flag.Uint64("namespace", 0, "Default namespace ID (slug invest if 0)")
	alertsEvery := flag.Duration("alerts-every", 5*time.Minute, "Run threshold alerts and EVM on this interval (0 to disable)")
	flag.Parse()

	if *token == "" {
		*token = os.Getenv("TOKEN")
	}

	cfg := agent.Config{
		ListenAddr:  *listen,
		CortezaAPI:  *cortezaAPI,
		Token:       *token,
		NamespaceID: *namespaceID,
		HTTPTimeout: 25 * time.Second,
	}
	eng := agent.NewEngine(cfg)
	if err := eng.Discover(context.Background()); err != nil {
		log.Printf("lowcode API discover: %v (will retry on first request)", err)
	} else {
		log.Printf("corteza API %s", eng.APIOrigin())
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Route("/api", func(r chi.Router) {
		api.New(eng).Mount(r)
	})

	srv := &http.Server{Addr: *listen, Handler: r}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("invest-engine listening on %s (api=%s ns=%d)", *listen, *cortezaAPI, *namespaceID)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	if *alertsEvery > 0 {
		go func() {
			log.Printf("alert scheduler every %v", *alertsEvery)
			t := time.NewTicker(*alertsEvery)
			defer t.Stop()
			eng.RunScheduled(ctx)
			for {
				select {
				case <-ctx.Done():
					return
				case <-t.C:
					eng.RunScheduled(ctx)
				}
			}
		}()
	}

	<-ctx.Done()
	shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdown)
}
