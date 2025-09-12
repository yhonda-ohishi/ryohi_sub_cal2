package services

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTPメトリクス
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)
	
	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)
	
	HTTPRequestsInFlight = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Current number of HTTP requests being served",
		},
	)
	
	// バックエンドメトリクス
	BackendHealthStatus = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "backend_health_status",
			Help: "Health status of backend services (1=healthy, 0=unhealthy)",
		},
		[]string{"backend", "endpoint"},
	)
	
	BackendRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "backend_requests_total",
			Help: "Total requests to backend services",
		},
		[]string{"backend", "endpoint", "status"},
	)
	
	BackendRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "backend_request_duration_seconds",
			Help:    "Backend request latency",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"backend", "endpoint"},
	)
	
	// ルーティングメトリクス
	RouteMatchDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "route_match_duration_seconds",
			Help:    "Time to match a route",
			Buckets: []float64{0.00001, 0.00005, 0.0001, 0.0005, 0.001, 0.005, 0.01},
		},
		[]string{"route"},
	)
	
	// レート制限メトリクス
	RateLimitExceeded = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rate_limit_exceeded_total",
			Help: "Total number of rate limit exceeded events",
		},
		[]string{"route", "client"},
	)
	
	// サーキットブレーカーメトリクス
	CircuitBreakerState = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "circuit_breaker_state",
			Help: "Circuit breaker state (0=closed, 1=open, 2=half-open)",
		},
		[]string{"backend"},
	)
	
	CircuitBreakerTrips = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "circuit_breaker_trips_total",
			Help: "Total number of circuit breaker trips",
		},
		[]string{"backend"},
	)
)

// MetricsCollector manages metrics collection
type MetricsCollector struct {
	registry *prometheus.Registry
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		registry: prometheus.DefaultRegisterer.(*prometheus.Registry),
	}
}

// RecordHTTPRequest records an HTTP request metric
func RecordHTTPRequest(method, path, status string, duration float64) {
	HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()
	HTTPRequestDuration.WithLabelValues(method, path, status).Observe(duration)
}

// RecordBackendRequest records a backend request metric
func RecordBackendRequest(backend, endpoint, status string, duration float64) {
	BackendRequestsTotal.WithLabelValues(backend, endpoint, status).Inc()
	BackendRequestDuration.WithLabelValues(backend, endpoint).Observe(duration)
}

// SetBackendHealth sets the health status of a backend
func SetBackendHealth(backend, endpoint string, healthy bool) {
	value := 0.0
	if healthy {
		value = 1.0
	}
	BackendHealthStatus.WithLabelValues(backend, endpoint).Set(value)
}

// SetCircuitBreakerState sets the circuit breaker state
func SetCircuitBreakerState(backend string, state int) {
	CircuitBreakerState.WithLabelValues(backend).Set(float64(state))
}

// RecordCircuitBreakerTrip records a circuit breaker trip
func RecordCircuitBreakerTrip(backend string) {
	CircuitBreakerTrips.WithLabelValues(backend).Inc()
}

// RecordRateLimitExceeded records a rate limit exceeded event
func RecordRateLimitExceeded(route, client string) {
	RateLimitExceeded.WithLabelValues(route, client).Inc()
}