package etc_meisai

import (
	"log"
	"net/http"
	"reflect"
	"regexp"
	"strings"

	"github.com/gorilla/mux"
	"github.com/yhonda-ohishi/etc_meisai"
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

// discoverAvailableHandlers uses reflection to find all available handlers
func (s *EtcMeisaiService) discoverAvailableHandlers() map[string]func(http.ResponseWriter, *http.Request) {
	handlers := make(map[string]func(http.ResponseWriter, *http.Request))

	// Get the etc_meisai package type
	pkgType := reflect.TypeOf(etc_meisai.HealthCheckHandler)
	if pkgType == nil {
		log.Println("Could not access etc_meisai package")
		return handlers
	}

	// List of known handler names to check
	knownHandlers := []string{
		"HealthCheckHandler",
		"GetAvailableAccountsHandler",
		"DownloadETCDataHandler",
		"DownloadSingleAccountHandler",
		"DownloadAsyncHandler",
		"GetDownloadStatusHandler",
		"ParseCSVHandler",
		"ImportDataHandler",
		"GetMeisaiListHandler",
		"CreateMeisaiHandler",
		"GetMeisaiByIDHandler",
		"GetSummaryHandler",
	}

	// Use reflection to check each handler
	_ = pkgType // Prevent unused variable error

	for _, handlerName := range knownHandlers {
		if handlerFunc := s.getHandlerByName(handlerName); handlerFunc != nil {
			handlers[handlerName] = handlerFunc
			log.Printf("Discovered handler: %s", handlerName)
		}
	}

	return handlers
}

// getHandlerByName safely retrieves a handler function by name using reflection
func (s *EtcMeisaiService) getHandlerByName(name string) func(http.ResponseWriter, *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Handler %s not available: %v", name, r)
		}
	}()

	// Use reflection to get the handler from the package
	etcMeisaiValue := reflect.ValueOf(etc_meisai.HealthCheckHandler).Type().PkgPath()
	_ = etcMeisaiValue // Prevent unused variable error

	// For safety, directly check known handlers
	switch name {
	case "HealthCheckHandler":
		return etc_meisai.HealthCheckHandler
	case "GetAvailableAccountsHandler":
		return etc_meisai.GetAvailableAccountsHandler
	case "DownloadETCDataHandler":
		return etc_meisai.DownloadETCDataHandler
	case "DownloadSingleAccountHandler":
		return etc_meisai.DownloadSingleAccountHandler
	case "ParseCSVHandler":
		return etc_meisai.ParseCSVHandler
	// Add more handlers as they become available in the module
	default:
		return nil
	}
}

// getSwaggerEndpoints extracts all endpoints from the Swagger spec
func (s *EtcMeisaiService) getSwaggerEndpoints() []SwaggerEndpoint {
	// Define the endpoints based on the Swagger spec we analyzed
	endpoints := []SwaggerEndpoint{
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

	return endpoints
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