package models

import (
	"time"
)

// RequestMetrics represents metrics for a single request
type RequestMetrics struct {
	RequestID    string        `json:"request_id"`
	Method       string        `json:"method"`
	Path         string        `json:"path"`
	StatusCode   int           `json:"status_code"`
	Duration     time.Duration `json:"duration"`
	BackendID    string        `json:"backend_id,omitempty"`
	BackendURL   string        `json:"backend_url,omitempty"`
	Error        string        `json:"error,omitempty"`
	Timestamp    time.Time     `json:"timestamp"`
	UserAgent    string        `json:"user_agent,omitempty"`
	ClientIP     string        `json:"client_ip,omitempty"`
	BytesIn      int64         `json:"bytes_in"`
	BytesOut     int64         `json:"bytes_out"`
}

// SystemMetrics represents system-wide metrics
type SystemMetrics struct {
	ActiveConnections  int64     `json:"active_connections"`
	TotalRequests      int64     `json:"total_requests"`
	SuccessfulRequests int64     `json:"successful_requests"`
	FailedRequests     int64     `json:"failed_requests"`
	AverageLatency     float64   `json:"average_latency_ms"`
	P50Latency         float64   `json:"p50_latency_ms"`
	P95Latency         float64   `json:"p95_latency_ms"`
	P99Latency         float64   `json:"p99_latency_ms"`
	MemoryUsage        int64     `json:"memory_usage_bytes"`
	CPUUsage           float64   `json:"cpu_usage_percent"`
	Uptime             time.Duration `json:"uptime"`
	Timestamp          time.Time `json:"timestamp"`
}

// BackendMetrics represents metrics for a specific backend
type BackendMetrics struct {
	BackendID          string        `json:"backend_id"`
	TotalRequests      int64         `json:"total_requests"`
	SuccessfulRequests int64         `json:"successful_requests"`
	FailedRequests     int64         `json:"failed_requests"`
	AverageLatency     float64       `json:"average_latency_ms"`
	P95Latency         float64       `json:"p95_latency_ms"`
	P99Latency         float64       `json:"p99_latency_ms"`
	HealthStatus       string        `json:"health_status"`
	CircuitBreakerState string       `json:"circuit_breaker_state,omitempty"`
	ActiveConnections  int64         `json:"active_connections"`
	LastError          string        `json:"last_error,omitempty"`
	LastErrorTime      *time.Time    `json:"last_error_time,omitempty"`
	UpdatedAt          time.Time     `json:"updated_at"`
}

// RouteMetrics represents metrics for a specific route
type RouteMetrics struct {
	RouteID            string    `json:"route_id"`
	Path               string    `json:"path"`
	TotalRequests      int64     `json:"total_requests"`
	SuccessfulRequests int64     `json:"successful_requests"`
	FailedRequests     int64     `json:"failed_requests"`
	AverageLatency     float64   `json:"average_latency_ms"`
	P95Latency         float64   `json:"p95_latency_ms"`
	P99Latency         float64   `json:"p99_latency_ms"`
	RateLimitedRequests int64    `json:"rate_limited_requests"`
	UnauthorizedRequests int64   `json:"unauthorized_requests"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// MetricsCollector represents the interface for collecting metrics
type MetricsCollector interface {
	RecordRequest(metrics *RequestMetrics)
	GetSystemMetrics() *SystemMetrics
	GetBackendMetrics(backendID string) *BackendMetrics
	GetRouteMetrics(routeID string) *RouteMetrics
	Reset()
}

// PrometheusMetrics represents Prometheus metric labels and values
type PrometheusMetrics struct {
	HTTPRequestsTotal    map[string]int64    `json:"http_requests_total"`
	HTTPRequestDuration  map[string]float64  `json:"http_request_duration_seconds"`
	HTTPRequestsInFlight int64               `json:"http_requests_in_flight"`
	BackendHealthStatus  map[string]int      `json:"backend_health_status"`
	CircuitBreakerState  map[string]string   `json:"circuit_breaker_state"`
	RateLimitRemaining   map[string]int      `json:"rate_limit_remaining"`
}

// LatencyHistogram represents a histogram of latency values
type LatencyHistogram struct {
	Count  int64              `json:"count"`
	Sum    float64            `json:"sum"`
	Min    float64            `json:"min"`
	Max    float64            `json:"max"`
	Mean   float64            `json:"mean"`
	StdDev float64            `json:"stddev"`
	Percentiles map[string]float64 `json:"percentiles"`
}

// Calculate updates the histogram with a new value
func (h *LatencyHistogram) Calculate(values []float64) {
	if len(values) == 0 {
		return
	}
	
	h.Count = int64(len(values))
	h.Sum = 0
	h.Min = values[0]
	h.Max = values[0]
	
	for _, v := range values {
		h.Sum += v
		if v < h.Min {
			h.Min = v
		}
		if v > h.Max {
			h.Max = v
		}
	}
	
	h.Mean = h.Sum / float64(h.Count)
	
	// Calculate percentiles (simplified - in production use a proper algorithm)
	h.Percentiles = make(map[string]float64)
	if h.Count > 0 {
		h.Percentiles["p50"] = percentile(values, 50)
		h.Percentiles["p75"] = percentile(values, 75)
		h.Percentiles["p90"] = percentile(values, 90)
		h.Percentiles["p95"] = percentile(values, 95)
		h.Percentiles["p99"] = percentile(values, 99)
	}
}

// percentile calculates the percentile value (simplified implementation)
func percentile(values []float64, p float64) float64 {
	if len(values) == 0 {
		return 0
	}
	
	// This is a simplified implementation
	// In production, use a proper percentile algorithm
	index := int(float64(len(values)) * p / 100)
	if index >= len(values) {
		index = len(values) - 1
	}
	
	return values[index]
}

// MetricsConfig represents metrics configuration
type MetricsConfig struct {
	Enabled         bool          `json:"enabled" yaml:"enabled"`
	Path            string        `json:"path" yaml:"path"`
	Port            int           `json:"port" yaml:"port"`
	CollectInterval time.Duration `json:"collect_interval" yaml:"collect_interval"`
	RetentionPeriod time.Duration `json:"retention_period" yaml:"retention_period"`
}

// Validate validates the metrics configuration
func (m *MetricsConfig) Validate() error {
	if !m.Enabled {
		return nil
	}
	
	if m.Path == "" {
		m.Path = "/metrics" // Default path
	}
	
	if m.Port == 0 {
		m.Port = 9090 // Default Prometheus port
	}
	
	if m.CollectInterval == 0 {
		m.CollectInterval = 10 * time.Second // Default collection interval
	}
	
	if m.RetentionPeriod == 0 {
		m.RetentionPeriod = 1 * time.Hour // Default retention period
	}
	
	return nil
}