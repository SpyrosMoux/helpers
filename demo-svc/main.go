package main

import (
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spyrosmoux/helpers/demo-svc/handlers"
	"go.uber.org/zap"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "demo_svc_http_requests_total",
			Help: "Total number of HTTP requests processed, labeled by method, route, and status_code.",
		},
		[]string{"method", "route", "status_code"},
	)
	httpRequestDurationSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "demo_svc_http_request_duration_seconds",
			Help:    "Latency of HTTP requests.",
			Buckets: prometheus.DefBuckets, // tune if needed
		},
		[]string{"method", "route", "status_code"},
	)
	logger  *zap.Logger
	slogger *zap.SugaredLogger
	version string
)

func init() {
	prometheus.MustRegister(httpRequestsTotal, httpRequestDurationSeconds)

	logger, _ := zap.NewProduction()
	defer logger.Sync()
	slogger = logger.Sugar()

	version = os.Getenv("VERSION")
	if version == "" {
		slogger.Warn("VERSION environment variable not found. Defaulting to v1.0.0")
		version = "v1.0.0"
	}
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// Middleware expects you to pass the logical route/template as `routeLabel`
// (e.g., "/users/:id" or from your router's route name) to avoid high-cardinality paths.
func Middleware(next http.Handler, routeLabel string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &statusRecorder{ResponseWriter: w, status: 200}
		start := time.Now()

		next.ServeHTTP(rec, r)

		method := r.Method
		status := strconv.Itoa(rec.status)

		httpRequestsTotal.WithLabelValues(method, routeLabel, status).Inc()
		httpRequestDurationSeconds.WithLabelValues(method, routeLabel, status).
			Observe(time.Since(start).Seconds())
	})
}

func main() {
	demoHandler := handlers.NewDemoHandler(slogger, version)

	http.Handle("/metrics", promhttp.Handler())
	http.Handle("/v1/ok", Middleware(http.HandlerFunc(demoHandler.HandleOk), "/v1/ok"))
	http.Handle("/v1/user-error", Middleware(http.HandlerFunc(demoHandler.HandleUserError), "/v1/user-error"))
	http.Handle("/v1/server-error", Middleware(http.HandlerFunc(demoHandler.HandleServerError), "/v1/server-error"))

	port := os.Getenv("HTTP_PORT")
	if port == "" {
		slogger.Warn("variable HTTP_PORT not found. Will use default port :8080")
		port = "8080"
	}

	slogger.Infow("starting server on",
		"port", port,
	)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		slogger.Fatalw("failed to start server on",
			"port", port,
			"err", err,
		)
	}
}
