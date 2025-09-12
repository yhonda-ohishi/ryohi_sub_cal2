package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/your-org/ryohi-router/src/models"
	"github.com/your-org/ryohi-router/src/services/health"
)

// HealthHandler returns an HTTP handler for health checks
func HealthHandler(checker *health.Checker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get health status from checker
		statuses := checker.GetAllStatuses()
		
		// Determine overall health
		overallHealthy := true
		services := make(map[string]models.ServiceHealthInfo)
		
		for serviceID, status := range statuses {
			info := models.ServiceHealthInfo{
				Status: status.Status,
			}
			
			if status.Status != "healthy" {
				overallHealthy = false
				if status.Message != "" {
					info.Message = status.Message
				}
			}
			
			services[serviceID] = info
		}
		
		// Create response
		response := models.HealthResponse{
			Status:    "healthy",
			Timestamp: time.Now().Format(time.RFC3339),
		}
		
		if !overallHealthy {
			response.Status = "unhealthy"
		}
		
		if len(services) > 0 {
			response.Services = services
		}
		
		// Set status code
		statusCode := http.StatusOK
		if !overallHealthy {
			statusCode = http.StatusServiceUnavailable
		}
		
		// Send response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(response)
	}
}