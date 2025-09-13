package models

import (
	"fmt"
	"time"
)

// HealthCheckConfig represents health check configuration
type HealthCheckConfig struct {
	Enabled            bool          `json:"enabled" yaml:"enabled" mapstructure:"enabled"`
	Path               string        `json:"path" yaml:"path" mapstructure:"path"`
	Interval           time.Duration `json:"interval" yaml:"interval" mapstructure:"interval" swaggertype:"integer"`
	Timeout            time.Duration `json:"timeout" yaml:"timeout" mapstructure:"timeout" swaggertype:"integer"`
	HealthyThreshold   int           `json:"healthy_threshold" yaml:"healthy_threshold" mapstructure:"healthy_threshold"`
	UnhealthyThreshold int           `json:"unhealthy_threshold" yaml:"unhealthy_threshold" mapstructure:"unhealthy_threshold"`
	ExpectedStatus     []int         `json:"expected_status" yaml:"expected_status" mapstructure:"expected_status"`
}

// Validate validates the health check configuration
func (h *HealthCheckConfig) Validate() error {
	if !h.Enabled {
		return nil
	}
	
	if h.Path == "" {
		h.Path = "/health" // Default path
	}
	
	if h.Interval == 0 {
		h.Interval = 30 * time.Second // Default interval
	} else if h.Interval < 1*time.Second {
		return fmt.Errorf("health check interval must be at least 1 second")
	}
	
	if h.Timeout == 0 {
		h.Timeout = 5 * time.Second // Default timeout
	} else if h.Timeout >= h.Interval {
		return fmt.Errorf("health check timeout must be less than interval")
	}
	
	if h.HealthyThreshold <= 0 {
		h.HealthyThreshold = 2 // Default healthy threshold
	}
	
	if h.UnhealthyThreshold <= 0 {
		h.UnhealthyThreshold = 3 // Default unhealthy threshold
	}
	
	if len(h.ExpectedStatus) == 0 {
		h.ExpectedStatus = []int{200} // Default expected status
	}
	
	for _, status := range h.ExpectedStatus {
		if status < 100 || status > 599 {
			return fmt.Errorf("invalid expected status code: %d", status)
		}
	}
	
	return nil
}

// IsExpectedStatus checks if the given status code is expected
func (h *HealthCheckConfig) IsExpectedStatus(statusCode int) bool {
	for _, expected := range h.ExpectedStatus {
		if statusCode == expected {
			return true
		}
	}
	return false
}

// HealthStatus represents the health status of a service or endpoint
type HealthStatus struct {
	ServiceID        string                 `json:"service_id"`
	Status           string                 `json:"status"` // healthy, unhealthy, unknown
	LastCheck        time.Time              `json:"last_check"`
	ConsecutiveOK    int                    `json:"consecutive_ok"`
	ConsecutiveFail  int                    `json:"consecutive_fail"`
	ResponseTime     time.Duration          `json:"response_time"`
	Message          string                 `json:"message,omitempty"`
	EndpointStatuses map[string]*EndpointHealth `json:"endpoint_statuses,omitempty"`
}

// EndpointHealth represents the health status of a single endpoint
type EndpointHealth struct {
	URL           string        `json:"url"`
	Healthy       bool          `json:"healthy"`
	LastCheck     time.Time     `json:"last_check"`
	ResponseTime  time.Duration `json:"response_time"`
	StatusCode    int           `json:"status_code"`
	ConsecutiveOK int           `json:"consecutive_ok"`
	ConsecutiveFail int         `json:"consecutive_fail"`
	Error         string        `json:"error,omitempty"`
}

// Update updates the health status based on a check result
func (h *HealthStatus) Update(success bool, responseTime time.Duration, message string) {
	h.LastCheck = time.Now()
	h.ResponseTime = responseTime
	h.Message = message
	
	if success {
		h.ConsecutiveOK++
		h.ConsecutiveFail = 0
	} else {
		h.ConsecutiveFail++
		h.ConsecutiveOK = 0
	}
	
	// Update overall status
	if h.ConsecutiveOK > 0 {
		h.Status = "healthy"
	} else if h.ConsecutiveFail > 0 {
		h.Status = "unhealthy"
	} else {
		h.Status = "unknown"
	}
}

// IsHealthy returns true if the service is considered healthy
func (h *HealthStatus) IsHealthy() bool {
	return h.Status == "healthy"
}

// UpdateEndpoint updates the health status of a specific endpoint
func (h *HealthStatus) UpdateEndpoint(url string, health *EndpointHealth) {
	if h.EndpointStatuses == nil {
		h.EndpointStatuses = make(map[string]*EndpointHealth)
	}
	h.EndpointStatuses[url] = health
	
	// Update overall status based on endpoints
	h.updateOverallStatus()
}

// updateOverallStatus updates the overall service status based on endpoint statuses
func (h *HealthStatus) updateOverallStatus() {
	if len(h.EndpointStatuses) == 0 {
		h.Status = "unknown"
		return
	}
	
	healthyCount := 0
	for _, endpoint := range h.EndpointStatuses {
		if endpoint.Healthy {
			healthyCount++
		}
	}
	
	// Service is healthy if at least one endpoint is healthy
	if healthyCount > 0 {
		h.Status = "healthy"
	} else {
		h.Status = "unhealthy"
	}
}

// HealthResponse represents the API health response
type HealthResponse struct {
	Status    string                       `json:"status"`
	Timestamp string                       `json:"timestamp"`
	Services  map[string]ServiceHealthInfo `json:"services,omitempty"`
}

// ServiceHealthInfo represents health information for a service
type ServiceHealthInfo struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}