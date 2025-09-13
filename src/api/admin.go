package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/your-org/ryohi-router/src/lib/config"
	"github.com/your-org/ryohi-router/src/models"
	"github.com/your-org/ryohi-router/src/services/health"
	"github.com/your-org/ryohi-router/src/services/router"
)

// GetRoutesHandler returns all routes
// @Summary      List all routes
// @Description  Get a list of all configured routes
// @Tags         admin
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200  {array}   models.RouteConfig
// @Router       /admin/routes [get]
func GetRoutesHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cfg.Routes)
	}
}

// CreateRouteHandler creates a new route
// @Summary      Create a new route
// @Description  Add a new route configuration
// @Tags         admin
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        route  body      models.RouteConfig  true  "Route configuration"
// @Success      201    {object}  models.RouteConfig
// @Failure      400    {string}  string  "Invalid request body"
// @Router       /admin/routes [post]
func CreateRouteHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var route models.RouteConfig
		if err := json.NewDecoder(r.Body).Decode(&route); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		
		if err := route.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		
		// Add route to config (in memory only for now)
		cfg.Routes = append(cfg.Routes, route)
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(route)
	}
}

// GetRouteHandler returns a specific route
// @Summary      Get a route by ID
// @Description  Get details of a specific route
// @Tags         admin
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id   path      string  true  "Route ID"
// @Success      200  {object}  models.RouteConfig
// @Failure      404  {string}  string  "Route not found"
// @Router       /admin/routes/{id} [get]
func GetRouteHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		routeID := vars["id"]
		
		for _, route := range cfg.Routes {
			if route.ID == routeID {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(route)
				return
			}
		}
		
		http.Error(w, "Route not found", http.StatusNotFound)
	}
}

// UpdateRouteHandler updates a route
// @Summary      Update a route
// @Description  Update an existing route configuration
// @Tags         admin
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id     path      string              true  "Route ID"
// @Param        route  body      models.RouteConfig  true  "Updated route configuration"
// @Success      200    {object}  models.RouteConfig
// @Failure      400    {string}  string  "Invalid request body"
// @Failure      404    {string}  string  "Route not found"
// @Router       /admin/routes/{id} [put]
func UpdateRouteHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		routeID := vars["id"]
		
		var updatedRoute models.RouteConfig
		if err := json.NewDecoder(r.Body).Decode(&updatedRoute); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		
		if err := updatedRoute.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		
		for i, route := range cfg.Routes {
			if route.ID == routeID {
				cfg.Routes[i] = updatedRoute
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(updatedRoute)
				return
			}
		}
		
		http.Error(w, "Route not found", http.StatusNotFound)
	}
}

// DeleteRouteHandler deletes a route
// @Summary      Delete a route
// @Description  Remove a route configuration
// @Tags         admin
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id   path      string  true  "Route ID"
// @Success      204  {string}  string  "No content"
// @Failure      404  {string}  string  "Route not found"
// @Router       /admin/routes/{id} [delete]
func DeleteRouteHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		routeID := vars["id"]
		
		for i, route := range cfg.Routes {
			if route.ID == routeID {
				// Remove route from slice
				cfg.Routes = append(cfg.Routes[:i], cfg.Routes[i+1:]...)
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}
		
		http.Error(w, "Route not found", http.StatusNotFound)
	}
}

// GetBackendsHandler returns all backends
// @Summary      List all backends
// @Description  Get a list of all configured backend services
// @Tags         admin
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200  {array}   models.BackendService
// @Router       /admin/backends [get]
func GetBackendsHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cfg.Backends)
	}
}

// CreateBackendHandler creates a new backend
// @Summary      Create a new backend
// @Description  Add a new backend service configuration
// @Tags         admin
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        backend  body      models.BackendService  true  "Backend configuration"
// @Success      201      {object}  models.BackendService
// @Failure      400      {string}  string  "Invalid request body"
// @Router       /admin/backends [post]
func CreateBackendHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var backend models.BackendService
		if err := json.NewDecoder(r.Body).Decode(&backend); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		
		if err := backend.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		
		// Add backend to config (in memory only for now)
		cfg.Backends = append(cfg.Backends, backend)
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(backend)
	}
}

// GetBackendHealthHandler returns health status for a backend
// @Summary      Get backend health status
// @Description  Get health status of a specific backend service
// @Tags         admin
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id   path      string  true  "Backend ID"
// @Success      200  {object}  models.ServiceHealthStatus
// @Failure      404  {string}  string  "Backend not found"
// @Router       /admin/backends/{id}/health [get]
func GetBackendHealthHandler(checker *health.Checker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		backendID := vars["id"]
		
		status := checker.GetStatus(backendID)
		
		response := map[string]interface{}{
			"backend_id": backendID,
			"status":     status.Status,
			"endpoints":  status.EndpointStatuses,
		}
		
		if status.Status == "unknown" {
			w.WriteHeader(http.StatusNotFound)
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// ReloadConfigHandler reloads the configuration
// @Summary      Reload configuration
// @Description  Reload the router configuration from file
// @Tags         admin
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200  {object}  map[string]string
// @Failure      500  {string}  string  "Failed to reload configuration"
// @Router       /admin/reload [post]
func ReloadConfigHandler(cfg *config.Config, router *router.Router) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// In a real implementation, this would reload from file
		// For now, just acknowledge the request
		
		if err := router.Reload(cfg); err != nil {
			http.Error(w, "Failed to reload configuration", http.StatusInternalServerError)
			return
		}
		
		response := map[string]string{
			"message":   "Configuration reloaded successfully",
			"timestamp": "2025-09-12T00:00:00Z",
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}