package server

import (
	"net/http"
	"strconv"
	"time"

	"github.com/99designs/basicauth-go"
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// statusRecorder captures the HTTP status code written by the wrapped handler.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// metricsMiddleware is the request middleware that provides prometheus metrics,
// replacing the retired 766b/chi-prometheus (which depended on go-chi/chi v3).
func metricsMiddleware(name string) func(http.Handler) http.Handler {
	requests := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: name,
			Name:      "http_requests_total",
			Help:      "Total number of HTTP requests.",
		},
		[]string{"code", "method", "path"},
	)
	duration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: name,
			Name:      "http_request_duration_milliseconds",
			Help:      "HTTP request duration in milliseconds.",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)
	prometheus.MustRegister(requests, duration)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(rec, r)

			routePath := r.URL.Path
			if rctx := chi.RouteContext(r.Context()); rctx != nil && rctx.RoutePattern() != "" {
				routePath = rctx.RoutePattern()
			}

			requests.WithLabelValues(strconv.Itoa(rec.status), r.Method, routePath).Inc()
			duration.WithLabelValues(r.Method, routePath).Observe(float64(time.Since(start).Milliseconds()))
		})
	}
}

func metricsMount(r chi.Router, username, password string) {
	r.Route("/metrics", func(r chi.Router) {
		r.Use(basicauth.New("Metrics", map[string][]string{
			username: {password},
		}))
		r.Handle("/", promhttp.Handler())
	})
}
