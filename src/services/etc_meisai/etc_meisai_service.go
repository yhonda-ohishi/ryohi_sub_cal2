package etc_meisai

import (
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/gorilla/mux"
	"github.com/yhonda-ohishi/etc_meisai"
	"gopkg.in/yaml.v3"
)

// EtcMeisaiService manages the etc_meisai module integration
type EtcMeisaiService struct {
	enabled bool
}

// NewEtcMeisaiService creates a new etc_meisai service instance
func NewEtcMeisaiService(enabled bool) *EtcMeisaiService {
	return &EtcMeisaiService{
		enabled: enabled,
	}
}

// RegisterRoutes registers all etc_meisai routes with the main router
func (s *EtcMeisaiService) RegisterRoutes(router *mux.Router) {
	if !s.enabled {
		return
	}

	log.Println("Starting automatic route discovery for ETC Meisai module...")

	// Automatically discover and register all available routes
	s.autoDiscoverAndRegisterRoutes(router)
}

// SwaggerEndpoint represents an endpoint from Swagger spec
type SwaggerEndpoint struct {
	Path    string
	Methods []string
}

// autoDiscoverAndRegisterRoutes automatically discovers and registers all available routes
func (s *EtcMeisaiService) autoDiscoverAndRegisterRoutes(router *mux.Router) {
	// Get all available endpoints from reflection
	availableHandlers := s.discoverAvailableHandlers()

	// Get all endpoints from Swagger spec
	swaggerEndpoints := s.getSwaggerEndpoints()

	registered := 0
	total := len(swaggerEndpoints)

	log.Printf("Found %d endpoints in Swagger spec, %d available handlers", total, len(availableHandlers))

	// Try to register each endpoint
	for _, endpoint := range swaggerEndpoints {
		handlerName := s.pathToHandlerName(endpoint.Path, endpoint.Methods)

		if handler, exists := availableHandlers[handlerName]; exists {
			s.registerHandler(router, endpoint.Path, endpoint.Methods, handler)
			registered++
		} else {
			log.Printf("Handler not found for %s %v (expected: %s)", endpoint.Path, endpoint.Methods, handlerName)
		}
	}

	log.Printf("Successfully registered %d/%d endpoints automatically", registered, total)
}

// discoverAvailableHandlers dynamically discovers all available handlers
func (s *EtcMeisaiService) discoverAvailableHandlers() map[string]func(http.ResponseWriter, *http.Request) {
	// Use GlobalRegistry from etc_meisai v0.0.15+ for complete automation
	if etc_meisai.GlobalRegistry != nil {
		return etc_meisai.GlobalRegistry.GetAll()
	}
	// Fallback to generated registry if GlobalRegistry is not available
	return s.DynamicHandlerRegistry()
}


// getSwaggerEndpoints dynamically extracts all endpoints from the etc_meisai Swagger spec
func (s *EtcMeisaiService) getSwaggerEndpoints() []SwaggerEndpoint {
	var endpoints []SwaggerEndpoint

	// Try to fetch Swagger from GitHub
	resp, err := http.Get("https://raw.githubusercontent.com/yhonda-ohishi/etc_meisai/master/docs/swagger.yaml")
	if err != nil {
		log.Printf("Failed to fetch Swagger from GitHub: %v", err)
		// Fall back to hardcoded endpoints if fetch fails
		return s.getFallbackEndpoints()
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read Swagger response: %v", err)
		return s.getFallbackEndpoints()
	}

	// Parse YAML
	var swagger map[string]interface{}
	if err := yaml.Unmarshal(body, &swagger); err != nil {
		log.Printf("Failed to parse Swagger YAML: %v", err)
		return s.getFallbackEndpoints()
	}

	// Extract paths
	paths, ok := swagger["paths"].(map[string]interface{})
	if !ok {
		log.Printf("No paths found in Swagger spec")
		return s.getFallbackEndpoints()
	}

	// Extract endpoints from paths
	for path, pathItem := range paths {
		if pathMap, ok := pathItem.(map[string]interface{}); ok {
			var methods []string
			for method := range pathMap {
				// Filter out non-HTTP methods
				upperMethod := strings.ToUpper(method)
				if upperMethod == "GET" || upperMethod == "POST" || upperMethod == "PUT" || upperMethod == "DELETE" || upperMethod == "PATCH" {
					methods = append(methods, upperMethod)
				}
			}
			if len(methods) > 0 {
				endpoints = append(endpoints, SwaggerEndpoint{
					Path:    path,
					Methods: methods,
				})
			}
		}
	}

	log.Printf("Dynamically discovered %d endpoints from Swagger", len(endpoints))
	return endpoints
}

// getFallbackEndpoints returns hardcoded endpoints as fallback
func (s *EtcMeisaiService) getFallbackEndpoints() []SwaggerEndpoint {
	log.Println("Using fallback hardcoded endpoints")
	return []SwaggerEndpoint{
		{Path: "/health", Methods: []string{"GET"}},
		{Path: "/api/etc/accounts", Methods: []string{"GET"}},
		{Path: "/api/etc/download", Methods: []string{"POST"}},
		{Path: "/api/etc/download-single", Methods: []string{"POST"}},
		{Path: "/api/etc/download-async", Methods: []string{"POST"}},
		{Path: "/api/etc/download-status/{job_id}", Methods: []string{"GET"}},
		{Path: "/api/etc/parse-csv", Methods: []string{"POST"}},
		{Path: "/api/etc/import", Methods: []string{"POST"}},
		{Path: "/api/etc/meisai", Methods: []string{"GET", "POST"}},
		{Path: "/api/etc/meisai/{id}", Methods: []string{"GET"}},
		{Path: "/api/etc/summary", Methods: []string{"GET"}},
	}
}

// pathToHandlerName converts a path and methods to expected handler name
func (s *EtcMeisaiService) pathToHandlerName(path string, methods []string) string {
	// Remove parameter placeholders for pattern matching
	_ = regexp.MustCompile(`\{[^}]+\}`).ReplaceAllString(path, "")

	// Simple mapping logic based on common patterns
	switch {
	case path == "/health":
		return "HealthCheckHandler"
	case path == "/api/etc/accounts":
		return "GetAvailableAccountsHandler"
	case path == "/api/etc/download" && contains(methods, "POST"):
		return "DownloadETCDataHandler"
	case path == "/api/etc/download-single":
		return "DownloadSingleAccountHandler"
	case path == "/api/etc/download-async":
		return "DownloadAsyncHandler"
	case strings.HasPrefix(path, "/api/etc/download-status/"):
		return "GetDownloadStatusHandler"
	case path == "/api/etc/parse-csv":
		return "ParseCSVHandler"
	case path == "/api/etc/import":
		return "ImportDataHandler"
	case path == "/api/etc/meisai" && contains(methods, "GET"):
		return "GetMeisaiListHandler"
	case path == "/api/etc/meisai" && contains(methods, "POST"):
		return "CreateMeisaiHandler"
	case strings.HasPrefix(path, "/api/etc/meisai/"):
		return "GetMeisaiByIDHandler"
	case path == "/api/etc/summary":
		return "GetSummaryHandler"
	default:
		return ""
	}
}

// registerHandler registers a handler with the router
func (s *EtcMeisaiService) registerHandler(router *mux.Router, path string, methods []string, handler func(http.ResponseWriter, *http.Request)) {
	router.HandleFunc(path, handler).Methods(methods...)
	log.Printf("✓ Auto-registered: %s %v", path, methods)
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// IsEnabled returns whether the etc_meisai service is enabled
func (s *EtcMeisaiService) IsEnabled() bool {
	return s.enabled
}

// HealthCheck returns the health status of the etc_meisai service
func (s *EtcMeisaiService) HealthCheck() map[string]interface{} {
	status := "disabled"
	if s.enabled {
		status = "healthy"
	}

	return map[string]interface{}{
		"status": status,
		"enabled": s.enabled,
	}
}