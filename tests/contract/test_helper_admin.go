package contract

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/your-org/ryohi-router/src/server"
)

// setupTestAdminRouter creates a test admin router for contract tests
func setupTestAdminRouter() http.Handler {
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
	
	// Return the admin router handler
	return srv.GetAdminRouter()
}

// setupTestMetricsRouter creates a test metrics router for contract tests
func setupTestMetricsRouter() http.Handler {
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
	
	// Return the metrics router handler
	return srv.GetMetricsRouter()
}