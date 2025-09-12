package models

import (
	"fmt"
	"net/url"
	"time"
)

// BackendService represents a backend service configuration
type BackendService struct {
	ID             string                `json:"id" yaml:"id"`
	Name           string                `json:"name" yaml:"name"`
	Endpoints      []EndpointConfig      `json:"endpoints" yaml:"endpoints"`
	LoadBalancer   LoadBalancerConfig    `json:"load_balancer" yaml:"load_balancer"`
	HealthCheck    HealthCheckConfig     `json:"health_check" yaml:"health_check"`
	CircuitBreaker CircuitBreakerConfig  `json:"circuit_breaker" yaml:"circuit_breaker"`
	RetryPolicy    RetryPolicyConfig     `json:"retry_policy" yaml:"retry_policy"`
	Enabled        bool                  `json:"enabled" yaml:"enabled"`
	CreatedAt      time.Time             `json:"created_at" yaml:"created_at"`
	UpdatedAt      time.Time             `json:"updated_at" yaml:"updated_at"`
}

// EndpointConfig represents a single endpoint in a backend service
type EndpointConfig struct {
	URL      string            `json:"url" yaml:"url"`
	Weight   int               `json:"weight" yaml:"weight"`
	Healthy  bool              `json:"healthy" yaml:"healthy"`
	Metadata map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// LoadBalancerConfig represents load balancer configuration
type LoadBalancerConfig struct {
	Algorithm     string `json:"algorithm" yaml:"algorithm"`
	StickySession bool   `json:"sticky_session" yaml:"sticky_session"`
}

// RetryPolicyConfig represents retry policy configuration
type RetryPolicyConfig struct {
	Enabled         bool          `json:"enabled" yaml:"enabled"`
	MaxAttempts     int           `json:"max_attempts" yaml:"max_attempts"`
	Backoff         string        `json:"backoff" yaml:"backoff"`
	InitialInterval time.Duration `json:"initial_interval" yaml:"initial_interval"`
	MaxInterval     time.Duration `json:"max_interval" yaml:"max_interval"`
}

// Validate validates the backend service configuration
func (b *BackendService) Validate() error {
	if b.ID == "" {
		return fmt.Errorf("backend ID is required")
	}
	
	if b.Name == "" {
		return fmt.Errorf("backend name is required")
	}
	
	if len(b.Name) > 255 {
		return fmt.Errorf("backend name cannot exceed 255 characters")
	}
	
	if len(b.Endpoints) == 0 {
		return fmt.Errorf("at least one endpoint is required")
	}
	
	for i, endpoint := range b.Endpoints {
		if err := endpoint.Validate(); err != nil {
			return fmt.Errorf("invalid endpoint %d: %w", i, err)
		}
	}
	
	if err := b.LoadBalancer.Validate(); err != nil {
		return fmt.Errorf("invalid load balancer config: %w", err)
	}
	
	if err := b.HealthCheck.Validate(); err != nil {
		return fmt.Errorf("invalid health check config: %w", err)
	}
	
	if err := b.CircuitBreaker.Validate(); err != nil {
		return fmt.Errorf("invalid circuit breaker config: %w", err)
	}
	
	if err := b.RetryPolicy.Validate(); err != nil {
		return fmt.Errorf("invalid retry policy config: %w", err)
	}
	
	return nil
}

// GetHealthyEndpoints returns only healthy endpoints
func (b *BackendService) GetHealthyEndpoints() []EndpointConfig {
	var healthy []EndpointConfig
	for _, endpoint := range b.Endpoints {
		if endpoint.Healthy {
			healthy = append(healthy, endpoint)
		}
	}
	return healthy
}

// Validate validates the endpoint configuration
func (e *EndpointConfig) Validate() error {
	if e.URL == "" {
		return fmt.Errorf("endpoint URL is required")
	}
	
	parsedURL, err := url.Parse(e.URL)
	if err != nil {
		return fmt.Errorf("invalid endpoint URL: %w", err)
	}
	
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("endpoint URL must use http or https scheme")
	}
	
	if e.Weight < 1 || e.Weight > 100 {
		return fmt.Errorf("endpoint weight must be between 1 and 100")
	}
	
	return nil
}

// Validate validates the load balancer configuration
func (l *LoadBalancerConfig) Validate() error {
	validAlgorithms := []string{"round-robin", "weighted", "least-conn", "ip-hash", "random"}
	valid := false
	for _, algo := range validAlgorithms {
		if l.Algorithm == algo {
			valid = true
			break
		}
	}
	
	if !valid {
		if l.Algorithm == "" {
			l.Algorithm = "round-robin" // Default algorithm
		} else {
			return fmt.Errorf("invalid load balancer algorithm: %s", l.Algorithm)
		}
	}
	
	return nil
}

// Validate validates the retry policy configuration
func (r *RetryPolicyConfig) Validate() error {
	if !r.Enabled {
		return nil
	}
	
	if r.MaxAttempts < 1 || r.MaxAttempts > 10 {
		return fmt.Errorf("max retry attempts must be between 1 and 10")
	}
	
	if r.Backoff != "constant" && r.Backoff != "exponential" && r.Backoff != "linear" {
		if r.Backoff == "" {
			r.Backoff = "exponential" // Default backoff
		} else {
			return fmt.Errorf("invalid backoff strategy: %s", r.Backoff)
		}
	}
	
	if r.InitialInterval == 0 {
		r.InitialInterval = 100 * time.Millisecond
	}
	
	if r.MaxInterval == 0 {
		r.MaxInterval = 10 * time.Second
	}
	
	if r.InitialInterval > r.MaxInterval {
		return fmt.Errorf("initial interval cannot be greater than max interval")
	}
	
	return nil
}

// BackendRegistry manages backend services
type BackendRegistry struct {
	Backends map[string]*BackendService `json:"backends" yaml:"backends"`
}

// GetBackend retrieves a backend by ID
func (br *BackendRegistry) GetBackend(id string) (*BackendService, error) {
	backend, exists := br.Backends[id]
	if !exists {
		return nil, fmt.Errorf("backend not found: %s", id)
	}
	if !backend.Enabled {
		return nil, fmt.Errorf("backend is disabled: %s", id)
	}
	return backend, nil
}

// RegisterBackend registers a new backend service
func (br *BackendRegistry) RegisterBackend(backend *BackendService) error {
	if err := backend.Validate(); err != nil {
		return fmt.Errorf("invalid backend configuration: %w", err)
	}
	
	if br.Backends == nil {
		br.Backends = make(map[string]*BackendService)
	}
	
	br.Backends[backend.ID] = backend
	return nil
}