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
func GetRoutesHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cfg.Routes)
	}
}

// CreateRouteHandler creates a new route
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
func GetBackendsHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cfg.Backends)
	}
}

// CreateBackendHandler creates a new backend
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