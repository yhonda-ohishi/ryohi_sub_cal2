package dtako

import (
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/mux"
	"github.com/yhonda-ohishi/dtako_mod"
	"github.com/your-org/ryohi-router/src/lib/adapters"
)

// DtakoService manages the dtako_mod integration
type DtakoService struct {
	enabled bool
}

// NewDtakoService creates a new dtako service instance
func NewDtakoService(enabled bool) *DtakoService {
	return &DtakoService{
		enabled: enabled,
	}
}

// RegisterRoutes registers all dtako routes with the main router
func (s *DtakoService) RegisterRoutes(router *mux.Router) {
	if !s.enabled {
		return
	}
	
	// Use the adapter to mount chi routes on mux
	adapters.AdaptChiToMux(router, "/dtako", func(r chi.Router) {
		// Register dtako_mod routes
		dtako_mod.RegisterRoutes(r)
	})
}

// IsEnabled returns whether the dtako service is enabled
func (s *DtakoService) IsEnabled() bool {
	return s.enabled
}

// HealthCheck returns the health status of the dtako service
func (s *DtakoService) HealthCheck() map[string]interface{} {
	status := "disabled"
	if s.enabled {
		status = "healthy"
	}
	
	return map[string]interface{}{
		"status": status,
		"enabled": s.enabled,
	}
}