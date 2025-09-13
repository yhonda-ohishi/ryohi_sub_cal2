package router

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"

	"github.com/your-org/ryohi-router/src/lib/config"
	"github.com/your-org/ryohi-router/src/models"
	"github.com/your-org/ryohi-router/src/services/loadbalancer"
)

// Router handles request routing to backend services
type Router struct {
	config       *config.Config
	logger       *slog.Logger
	backends     map[string]*Backend
	routes       *models.RouteCollection
	mutex        sync.RWMutex
}

// Backend represents a backend service with load balancer and proxy
type Backend struct {
	Service      *models.BackendService
	LoadBalancer loadbalancer.LoadBalancer
	Proxies      map[string]*httputil.ReverseProxy
}

// New creates a new router
func New(cfg *config.Config, logger *slog.Logger) (*Router, error) {
	r := &Router{
		config:   cfg,
		logger:   logger,
		backends: make(map[string]*Backend),
		routes: &models.RouteCollection{
			Routes: make([]*models.RouteConfig, 0),
		},
	}

	// Initialize backends
	for i := range cfg.Backends {
		backend := &cfg.Backends[i]
		if err := r.initializeBackend(backend); err != nil {
			return nil, fmt.Errorf("failed to initialize backend %s: %w", backend.ID, err)
		}
	}

	// Initialize routes
	for i := range cfg.Routes {
		route := cfg.Routes[i]
		r.routes.Routes = append(r.routes.Routes, &route)
	}

	return r, nil
}

// initializeBackend initializes a backend with its load balancer and proxies
func (r *Router) initializeBackend(service *models.BackendService) error {
	// Create load balancer
	lb, err := loadbalancer.New(&service.LoadBalancer, service.Endpoints)
	if err != nil {
		return fmt.Errorf("failed to create load balancer: %w", err)
	}

	// Create proxies for each endpoint
	proxies := make(map[string]*httputil.ReverseProxy)
	for _, endpoint := range service.Endpoints {
		targetURL, err := url.Parse(endpoint.URL)
		if err != nil {
			return fmt.Errorf("invalid endpoint URL %s: %w", endpoint.URL, err)
		}

		proxy := httputil.NewSingleHostReverseProxy(targetURL)
		
		// Customize proxy behavior
		proxy.ErrorHandler = func(w http.ResponseWriter, req *http.Request, err error) {
			r.logger.Error("Proxy error", 
				"backend", service.ID,
				"url", endpoint.URL,
				"error", err,
			)
			w.WriteHeader(http.StatusBadGateway)
			w.Write([]byte("Bad Gateway"))
		}

		proxies[endpoint.URL] = proxy
	}

	r.backends[service.ID] = &Backend{
		Service:      service,
		LoadBalancer: lb,
		Proxies:      proxies,
	}

	return nil
}

// CreateHandler creates an HTTP handler for a route
func (r *Router) CreateHandler(route *models.RouteConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		// Get backend
		backend, err := r.getBackend(route.Backend)
		if err != nil {
			r.logger.Error("Backend not found", "backend", route.Backend, "error", err)
			http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
			return
		}

		// Select endpoint using load balancer
		endpoint := backend.LoadBalancer.Next()
		if endpoint == nil {
			r.logger.Error("No healthy endpoints", "backend", route.Backend)
			http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
			return
		}

		// Get proxy for endpoint
		proxy, exists := backend.Proxies[endpoint.URL]
		if !exists {
			r.logger.Error("Proxy not found for endpoint", "url", endpoint.URL)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Forward request
		r.logger.Debug("Forwarding request",
			"path", req.URL.Path,
			"method", req.Method,
			"backend", route.Backend,
			"endpoint", endpoint.URL,
		)

		proxy.ServeHTTP(w, req)
	}
}

// getBackend retrieves a backend by ID
func (r *Router) getBackend(id string) (*Backend, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	backend, exists := r.backends[id]
	if !exists {
		return nil, fmt.Errorf("backend not found: %s", id)
	}

	if !backend.Service.Enabled {
		return nil, fmt.Errorf("backend is disabled: %s", id)
	}

	return backend, nil
}

// Reload reloads the router configuration
func (r *Router) Reload(cfg *config.Config) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Clear existing configuration
	r.backends = make(map[string]*Backend)
	r.routes = &models.RouteCollection{
		Routes: make([]*models.RouteConfig, 0),
	}

	// Reinitialize backends
	for i := range cfg.Backends {
		backend := &cfg.Backends[i]
		if err := r.initializeBackend(backend); err != nil {
			return fmt.Errorf("failed to initialize backend %s: %w", backend.ID, err)
		}
	}

	// Reinitialize routes
	for i := range cfg.Routes {
		route := cfg.Routes[i]
		r.routes.Routes = append(r.routes.Routes, &route)
	}

	r.config = cfg
	r.logger.Info("Router configuration reloaded")

	return nil
}