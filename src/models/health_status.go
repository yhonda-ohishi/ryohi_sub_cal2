package models

// ServiceHealthStatus represents the health status of a service
type ServiceHealthStatus struct {
	ServiceID        string                      `json:"service_id"`
	Status           string                      `json:"status"`
	Message          string                      `json:"message,omitempty"`
	EndpointStatuses map[string]EndpointHealthStatus `json:"endpoint_statuses,omitempty"`
	LastChecked      string                      `json:"last_checked"`
}

// EndpointHealthStatus represents the health status of an endpoint
type EndpointHealthStatus struct {
	URL         string `json:"url"`
	Status      string `json:"status"`
	ResponseTime int64  `json:"response_time_ms"`
	Message     string `json:"message,omitempty"`
}

