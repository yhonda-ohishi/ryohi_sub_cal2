package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/your-org/ryohi-router/src/api"
	"github.com/your-org/ryohi-router/src/lib/config"
	"github.com/your-org/ryohi-router/src/lib/middleware"
	"github.com/your-org/ryohi-router/src/services/health"
	"github.com/your-org/ryohi-router/src/services/router"
)

// Server represents the main router server
type Server struct {
	config       *config.Config
	logger       *slog.Logger
	mainServer   *http.Server
	adminServer  *http.Server
	metricsServer *http.Server
	router       *router.Router
	healthChecker *health.Checker
	wg           sync.WaitGroup
}

// New creates a new server instance
func New(cfg *config.Config, logger *slog.Logger) (*Server, error) {
	s := &Server{
		config: cfg,
		logger: logger,
	}

	// Initialize router
	routerService, err := router.New(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create router: %w", err)
	}
	s.router = routerService

	// Initialize health checker
	s.healthChecker = health.NewChecker(cfg, logger)

	// Setup main server
	mainRouter := s.setupMainRouter()
	s.mainServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Router.Port),
		Handler:      mainRouter,
		ReadTimeout:  cfg.Router.ReadTimeout,
		WriteTimeout: cfg.Router.WriteTimeout,
		IdleTimeout:  cfg.Router.IdleTimeout,
	}

	// Setup admin server if enabled
	if cfg.Admin.Enabled {
		adminRouter := s.setupAdminRouter()
		s.adminServer = &http.Server{
			Addr:         fmt.Sprintf(":%d", cfg.Admin.Port),
			Handler:      adminRouter,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
		}
	}

	// Setup metrics server if enabled
	if cfg.Metrics.Enabled {
		metricsRouter := s.setupMetricsRouter()
		s.metricsServer = &http.Server{
			Addr:    fmt.Sprintf(":%d", cfg.Metrics.Port),
			Handler: metricsRouter,
		}
	}

	return s, nil
}

// setupMainRouter sets up the main router with all routes and middleware
func (s *Server) setupMainRouter() http.Handler {
	r := mux.NewRouter()

	// Apply global middleware
	handler := middleware.Chain(
		r,
		middleware.RequestID(),
		middleware.Logger(s.logger),
		middleware.Recovery(s.logger),
		middleware.Metrics(),
	)

	// Health endpoint (no auth required)
	r.HandleFunc("/health", api.HealthHandler(s.healthChecker)).Methods("GET")

	// Setup route handlers
	for _, route := range s.config.Routes {
		if !route.Enabled {
			continue
		}

		// Create route-specific handler
		var routeHandler http.Handler = s.router.CreateHandler(&route)

		// Apply route-specific middleware
		if route.RateLimit != nil && route.RateLimit.Enabled {
			routeHandler = middleware.RateLimit(route.RateLimit)(routeHandler)
		}

		if route.Auth != nil && route.Auth.Enabled {
			routeHandler = middleware.Auth(route.Auth)(routeHandler)
		}

		// Register route
		r.PathPrefix(route.Path).Handler(routeHandler).Methods(route.Method...)
	}

	return handler
}

// setupAdminRouter sets up the admin API router
func (s *Server) setupAdminRouter() http.Handler {
	r := mux.NewRouter()

	// Apply admin middleware
	handler := middleware.Chain(
		r,
		middleware.RequestID(),
		middleware.Logger(s.logger),
		middleware.APIKeyAuth(s.config.Admin.APIKey),
	)

	// Admin API endpoints
	r.HandleFunc("/admin/routes", api.GetRoutesHandler(s.config)).Methods("GET")
	r.HandleFunc("/admin/routes", api.CreateRouteHandler(s.config)).Methods("POST")
	r.HandleFunc("/admin/routes/{id}", api.GetRouteHandler(s.config)).Methods("GET")
	r.HandleFunc("/admin/routes/{id}", api.UpdateRouteHandler(s.config)).Methods("PUT")
	r.HandleFunc("/admin/routes/{id}", api.DeleteRouteHandler(s.config)).Methods("DELETE")

	r.HandleFunc("/admin/backends", api.GetBackendsHandler(s.config)).Methods("GET")
	r.HandleFunc("/admin/backends", api.CreateBackendHandler(s.config)).Methods("POST")
	r.HandleFunc("/admin/backends/{id}/health", api.GetBackendHealthHandler(s.healthChecker)).Methods("GET")

	r.HandleFunc("/admin/reload", api.ReloadConfigHandler(s.config, s.router)).Methods("POST")

	return handler
}

// setupMetricsRouter sets up the metrics endpoint router
func (s *Server) setupMetricsRouter() http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/metrics", api.MetricsHandler()).Methods("GET")
	return r
}

// Start starts all servers
func (s *Server) Start(ctx context.Context) error {
	// Start health checker
	s.healthChecker.Start(ctx)

	// Start main server
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.logger.Info("Starting main server", "port", s.config.Router.Port)
		if err := s.mainServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("Main server error", "error", err)
		}
	}()

	// Start admin server
	if s.adminServer != nil {
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			s.logger.Info("Starting admin server", "port", s.config.Admin.Port)
			if err := s.adminServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				s.logger.Error("Admin server error", "error", err)
			}
		}()
	}

	// Start metrics server
	if s.metricsServer != nil {
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			s.logger.Info("Starting metrics server", "port", s.config.Metrics.Port)
			if err := s.metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				s.logger.Error("Metrics server error", "error", err)
			}
		}()
	}

	// Wait for context cancellation
	<-ctx.Done()
	return nil
}

// Shutdown gracefully shuts down all servers
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down servers...")

	// Stop health checker
	s.healthChecker.Stop()

	// Shutdown servers
	var shutdownErr error

	if err := s.mainServer.Shutdown(ctx); err != nil {
		s.logger.Error("Failed to shutdown main server", "error", err)
		shutdownErr = err
	}

	if s.adminServer != nil {
		if err := s.adminServer.Shutdown(ctx); err != nil {
			s.logger.Error("Failed to shutdown admin server", "error", err)
			shutdownErr = err
		}
	}

	if s.metricsServer != nil {
		if err := s.metricsServer.Shutdown(ctx); err != nil {
			s.logger.Error("Failed to shutdown metrics server", "error", err)
			shutdownErr = err
		}
	}

	// Wait for all goroutines to finish
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		s.logger.Info("All servers stopped gracefully")
	case <-ctx.Done():
		s.logger.Warn("Shutdown timeout exceeded")
		return ctx.Err()
	}

	return shutdownErr
}

// GetRouter returns the router for testing
func (s *Server) GetRouter() http.Handler {
	return s.setupMainRouter()
}

// GetAdminRouter returns the admin router for testing
func (s *Server) GetAdminRouter() http.Handler {
	return s.setupAdminRouter()
}

// GetMetricsRouter returns the metrics router for testing  
func (s *Server) GetMetricsRouter() http.Handler {
	return s.setupMetricsRouter()
}