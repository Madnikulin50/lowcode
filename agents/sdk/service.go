package sdk

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type alias struct {
	Method    string
	Path      string
	Operation string
}

type Service struct {
	cfg      Config
	comps    []Component
	backend  Backend
	compose  *Client
	callback *Callback
	aliases   []alias
	syncOps   map[string]bool
	extraAPI  func(chi.Router)
	extraRoot func(chi.Router)
}

func New(cfg Config) *Service {
	if cfg.Listen == "" {
		cfg.Listen = ":8080"
	}
	if cfg.Version == "" {
		cfg.Version = "v1"
	}
	s := &Service{
		cfg:      cfg,
		callback: NewCallback(),
		syncOps:  map[string]bool{},
	}
	if cfg.CortezaAPI != "" {
		s.compose = NewClient(cfg.CortezaAPI, cfg.Token, cfg.NamespaceID).WithSlug(cfg.Slug)
	}
	return s
}

func (s *Service) Callback() *Callback { return s.callback }
func (s *Service) Compose() *Client    { return s.compose }
func (s *Service) Config() Config      { return s.cfg }

func (s *Service) SetBackend(b Backend) *Service {
	s.backend = b
	return s
}

func (s *Service) Register(cs ...Component) *Service {
	s.comps = append(s.comps, cs...)
	return s
}

func (s *Service) Alias(method, path, operation string) *Service {
	s.aliases = append(s.aliases, alias{Method: method, Path: path, Operation: operation})
	return s
}

func (s *Service) Sync(operation string) *Service {
	if s.syncOps == nil {
		s.syncOps = map[string]bool{}
	}
	s.syncOps[operation] = true
	return s
}

func (s *Service) MountAPI(fn func(chi.Router)) *Service {
	s.extraAPI = fn
	return s
}

func (s *Service) MountRoot(fn func(chi.Router)) *Service {
	s.extraRoot = fn
	return s
}

func (s *Service) Meta() Meta {
	comps := make([]Descriptor, 0, len(s.comps))
	for _, c := range s.comps {
		d := c.Descriptor()
		if d.Service == "" {
			d.Service = s.cfg.Handle
		}
		if d.Execution == "" {
			d.Execution = ExecRemote
		}
		comps = append(comps, d)
	}
	return Meta{
		Handle:       s.cfg.Handle,
		Name:         s.cfg.Name,
		Version:      s.cfg.Version,
		PublicURL:    s.cfg.PublicURL,
		Components:   comps,
		Capabilities: CapabilitiesOf(s.comps),
	}
}

func (s *Service) Router() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(corsMiddleware)
	r.Get("/health", s.healthHTTP)
	r.Route("/api", s.mountAPI)
	if s.extraRoot != nil {
		s.extraRoot(r)
	}
	return r
}

func (s *Service) mountAPI(r chi.Router) {
	r.Get("/health", s.healthHTTP)
	r.Get("/meta", s.metaHTTP)
	r.Post("/register", s.registerHTTP)
	r.Post("/jobs", s.startJobHTTP)
	r.Get("/jobs", s.listJobsHTTP)
	r.Get("/jobs/{jobID}/items", s.itemsHTTP)
	r.Get("/jobs/{jobID}", s.getJobHTTP)
	r.Post("/call/{op}", s.callHTTP)

	aliases := append([]alias(nil), s.aliases...)
	sort.SliceStable(aliases, func(i, j int) bool {
		// Static paths first so /jobs/due is not captured by /jobs/{jobID}.
		return !pathHasParam(aliases[i].Path) && pathHasParam(aliases[j].Path)
	})
	for _, a := range aliases {
		a := a
		p := a.Path
		if p == "" {
			continue
		}
		if p[0] != '/' {
			p = "/" + p
		}
		h := func(w http.ResponseWriter, r *http.Request) {
			s.startJobHTTPOp(w, r, a.Operation)
		}
		switch a.Method {
		case http.MethodGet:
			if a.Operation == "" {
				r.Get(p, s.getJobHTTP)
			} else {
				r.Get(p, h)
			}
		default:
			r.Method(a.Method, p, http.HandlerFunc(h))
		}
	}
	if s.extraAPI != nil {
		s.extraAPI(r)
	}
}

func (s *Service) Listen(ctx context.Context) error {
	if ctx == nil {
		var stop context.CancelFunc
		ctx, stop = signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()
	}
	if s.compose != nil {
		if err := s.compose.Discover(context.Background()); err != nil {
			log.Printf("sdk compose discover: %v", err)
		}
	}
	if s.cfg.Heartbeat.Module != "" && s.cfg.Heartbeat.Interval > 0 && s.compose != nil {
		go s.heartbeatLoop(ctx)
	}
	srv := &http.Server{Addr: s.cfg.Listen, Handler: s.Router()}
	errCh := make(chan error, 1)
	go func() {
		log.Printf("%s listening on %s", firstNonEmpty(s.cfg.Handle, "agent"), s.cfg.Listen)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()
	select {
	case <-ctx.Done():
		shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdown)
		return nil
	case err := <-errCh:
		return err
	}
}

func (s *Service) heartbeatLoop(ctx context.Context) {
	tick := time.NewTicker(s.cfg.Heartbeat.Interval)
	defer tick.Stop()
	ident := s.cfg.Identity(s.Meta().Capabilities)
	_ = s.compose.Heartbeat(ctx, s.cfg.Heartbeat.Module, ident)
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			ident = s.cfg.Identity(s.Meta().Capabilities)
			if err := s.compose.Heartbeat(ctx, s.cfg.Heartbeat.Module, ident); err != nil {
				log.Printf("sdk heartbeat: %v", err)
			}
		}
	}
}

func Env(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}
