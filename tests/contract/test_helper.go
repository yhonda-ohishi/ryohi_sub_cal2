package contract

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/your-org/ryohi-router/src/lib/config"
	"github.com/your-org/ryohi-router/src/models"
	"github.com/your-org/ryohi-router/src/server"
)

// setupTestRouter creates a test router for contract tests
func setupTestRouter() http.Handler {
	// Create test configuration
	cfg := createTestConfig()
	
	// Create logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError, // Only show errors in tests
	}))
	
	// Create server
	srv, err := server.New(cfg, logger)
	if err != nil {
		panic(err)
	}
	
	// Return the router handler
	return srv.GetRouter()
}

// createTestConfig creates a test configuration
func createTestConfig() *config.Config {
	return &config.Config{
		Version: "1.0",
		Router: config.RouterConfig{
			Port:         8080,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  120 * time.Second,
		},
		Admin: config.AdminConfig{
			Enabled: true,
			APIKey:  "valid-api-key",
			Port:    8081,
		},
		Logging: config.LoggingConfig{
			Level:  "error",
			Format: "json",
			Output: "stdout",
		},
		Metrics: config.MetricsConfig{
			Enabled: true,
			Path:    "/metrics",
			Port:    9090,
		},
		Backends: []models.BackendService{
			{
				ID:   "test-backend",
				Name: "Test Backend",
				Endpoints: []models.EndpointConfig{
					{
						URL:     "http://localhost:3000",
						Weight:  100,
						Healthy: true,
					},
				},
				LoadBalancer: models.LoadBalancerConfig{
					Algorithm: "round-robin",
				},
				HealthCheck: models.HealthCheckConfig{
					Enabled:  true,
					Path:     "/health",
					Interval: 30 * time.Second,
					Timeout:  5 * time.Second,
				},
				CircuitBreaker: models.CircuitBreakerConfig{
					Enabled:      true,
					MaxRequests:  3,
					FailureRatio: 0.6,
					Timeout:      30 * time.Second,
				},
				Enabled:   true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
		Routes: []models.RouteConfig{
			{
				ID:        "test-route",
				Path:      "/api/v1/*",
				Method:    []string{"GET", "POST", "PUT", "DELETE"},
				Backend:   "test-backend",
				Timeout:   30 * time.Second,
				Priority:  100,
				Enabled:   true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
	}
}