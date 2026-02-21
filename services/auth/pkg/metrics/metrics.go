package metrics

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

var (
	// gRPC metrics
	grpcRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_requests_total",
			Help: "Total number of gRPC requests",
		},
		[]string{"service", "method", "status"},
	)

	grpcRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "grpc_request_duration_seconds",
			Help:    "Duration of gRPC requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "method"},
	)

	// Auth-specific metrics
	registrationsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "auth_registrations_total",
			Help: "Total number of user registrations",
		},
	)

	loginsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_logins_total",
			Help: "Total number of login attempts",
		},
		[]string{"status"},
	)

	oauthLoginsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_oauth_logins_total",
			Help: "Total number of OAuth login attempts",
		},
		[]string{"provider", "status"},
	)

	activeSessionsGauge = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "auth_active_sessions",
			Help: "Number of active user sessions",
		},
	)

	tokenValidationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_token_validations_total",
			Help: "Total number of token validations",
		},
		[]string{"status"},
	)
)

type Server struct {
	port   int
	path   string
	server *http.Server
}

func NewServer(port int, path string) *Server {
	return &Server{
		port: port,
		path: path,
	}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()
	mux.Handle(s.path, promhttp.Handler())

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: mux,
	}

	return s.server.ListenAndServe()
}

func (s *Server) Shutdown() error {
	if s.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return s.server.Shutdown(ctx)
	}
	return nil
}

// UnaryServerInterceptor returns a gRPC interceptor for metrics
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()

		// Call handler
		resp, err := handler(ctx, req)

		// Record metrics
		duration := time.Since(start).Seconds()
		statusCode := "success"
		if err != nil {
			statusCode = status.Code(err).String()
		}

		grpcRequestsTotal.WithLabelValues("auth-service", info.FullMethod, statusCode).Inc()
		grpcRequestDuration.WithLabelValues("auth-service", info.FullMethod).Observe(duration)

		return resp, err
	}
}

// Helper functions for recording auth-specific metrics

func RecordRegistration() {
	registrationsTotal.Inc()
}

func RecordLogin(success bool) {
	status := "success"
	if !success {
		status = "failure"
	}
	loginsTotal.WithLabelValues(status).Inc()
}

func RecordOAuthLogin(provider string, success bool) {
	status := "success"
	if !success {
		status = "failure"
	}
	oauthLoginsTotal.WithLabelValues(provider, status).Inc()
}

func SetActiveSessions(count float64) {
	activeSessionsGauge.Set(count)
}

func RecordTokenValidation(valid bool) {
	status := "valid"
	if !valid {
		status = "invalid"
	}
	tokenValidationsTotal.WithLabelValues(status).Inc()
}
