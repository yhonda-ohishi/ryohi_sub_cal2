package health

import (
	"context"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/your-org/ryohi-router/src/lib/config"
	"github.com/your-org/ryohi-router/src/models"
)

// Checker performs health checks on backend services
type Checker struct {
	config    *config.Config
	logger    *slog.Logger
	statuses  map[string]*models.HealthStatus
	mutex     sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc
	client    *http.Client
}

// NewChecker creates a new health checker
func NewChecker(cfg *config.Config, logger *slog.Logger) *Checker {
	return &Checker{
		config:   cfg,
		logger:   logger,
		statuses: make(map[string]*models.HealthStatus),
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// Start starts the health checker
func (c *Checker) Start(ctx context.Context) {
	c.ctx, c.cancel = context.WithCancel(ctx)
	
	c.logger.Info("Starting health checker")
	
	// Initialize health status for each backend
	for _, backend := range c.config.Backends {
		if !backend.Enabled {
			continue
		}
		
		c.logger.Info("Initializing health check for backend", 
			"backend", backend.ID, 
			"healthcheck_enabled", backend.HealthCheck.Enabled,
			"interval", backend.HealthCheck.Interval,
			"path", backend.HealthCheck.Path)
		
		c.statuses[backend.ID] = &models.HealthStatus{
			ServiceID: backend.ID,
			Status:    "unknown",
			LastCheck: time.Now(),
		}
		
		// Start health check goroutine for this backend
		if backend.HealthCheck.Enabled {
			c.logger.Info("Starting health check goroutine", "backend", backend.ID, "interval", backend.HealthCheck.Interval)
			go c.checkBackendHealth(backend)
		}
	}
}

// Stop stops the health checker
func (c *Checker) Stop() {
	if c.cancel != nil {
		c.cancel()
	}
}

// GetStatus returns the health status for a specific service
func (c *Checker) GetStatus(serviceID string) *models.HealthStatus {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	status, exists := c.statuses[serviceID]
	if !exists {
		return &models.HealthStatus{
			ServiceID: serviceID,
			Status:    "unknown",
			LastCheck: time.Now(),
		}
	}
	
	return status
}

// GetAllStatuses returns all health statuses
func (c *Checker) GetAllStatuses() map[string]*models.HealthStatus {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	// Create a copy of the statuses map
	result := make(map[string]*models.HealthStatus)
	for k, v := range c.statuses {
		result[k] = v
	}
	
	return result
}

// checkBackendHealth performs health checks for a backend
func (c *Checker) checkBackendHealth(backend models.BackendService) {
	ticker := time.NewTicker(backend.HealthCheck.Interval)
	defer ticker.Stop()
	
	// Perform initial check
	c.performHealthCheck(&backend)
	
	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			c.performHealthCheck(&backend)
		}
	}
}

// performHealthCheck performs a single health check
func (c *Checker) performHealthCheck(backend *models.BackendService) {
	c.logger.Debug("Performing health check", "backend", backend.ID)
	
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	status, exists := c.statuses[backend.ID]
	if !exists {
		status = &models.HealthStatus{
			ServiceID: backend.ID,
			Status:    "unknown",
		}
		c.statuses[backend.ID] = status
	}
	
	// Check each endpoint
	atLeastOneHealthy := false
	var lastError string
	
	for _, endpoint := range backend.Endpoints {
		healthy, responseTime, err := c.checkEndpoint(endpoint.URL, backend.HealthCheck)
		
		c.logger.Debug("Endpoint health check result", 
			"backend", backend.ID, 
			"endpoint", endpoint.URL, 
			"healthy", healthy, 
			"responseTime", responseTime,
			"error", err)
		
		endpointHealth := &models.EndpointHealth{
			URL:          endpoint.URL,
			Healthy:      healthy,
			LastCheck:    time.Now(),
			ResponseTime: responseTime,
		}
		
		if err != nil {
			endpointHealth.Error = err.Error()
			lastError = err.Error()
		}
		
		if healthy {
			atLeastOneHealthy = true
		}
		
		// Update endpoint status
		status.UpdateEndpoint(endpoint.URL, endpointHealth)
	}
	
	// Update overall status - backend is healthy if at least one endpoint is healthy
	if atLeastOneHealthy {
		status.Update(true, 0, "At least one endpoint healthy")
		c.logger.Debug("Backend status updated to healthy", "backend", backend.ID)
	} else {
		status.Update(false, 0, lastError)
		c.logger.Debug("Backend status updated to unhealthy", "backend", backend.ID, "error", lastError)
	}
}

// checkEndpoint checks a single endpoint
func (c *Checker) checkEndpoint(url string, config models.HealthCheckConfig) (bool, time.Duration, error) {
	healthURL := url + config.Path
	
	start := time.Now()
	req, err := http.NewRequest("GET", healthURL, nil)
	if err != nil {
		return false, 0, err
	}
	
	ctx, cancel := context.WithTimeout(c.ctx, config.Timeout)
	defer cancel()
	req = req.WithContext(ctx)
	
	resp, err := c.client.Do(req)
	duration := time.Since(start)
	
	if err != nil {
		return false, duration, err
	}
	defer resp.Body.Close()
	
	// Check if status code is expected
	expectedStatus := config.ExpectedStatus
	if len(expectedStatus) == 0 {
		expectedStatus = []int{200} // Default to 200 if not specified
	}
	
	isExpected := false
	for _, expected := range expectedStatus {
		if resp.StatusCode == expected {
			isExpected = true
			break
		}
	}
	
	if !isExpected {
		c.logger.Debug("Status code not expected", "url", healthURL, "statusCode", resp.StatusCode, "expected", expectedStatus)
		return false, duration, nil
	}
	
	return true, duration, nil
}