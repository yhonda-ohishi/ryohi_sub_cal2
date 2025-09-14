package services

import "github.com/gorilla/mux"

// ModuleService defines the interface for pluggable modules
type ModuleService interface {
	// RegisterRoutes registers all routes for this module
	RegisterRoutes(router *mux.Router)

	// ModuleName returns the module name for path prefix
	ModuleName() string

	// SwaggerURL returns the URL to the module's Swagger documentation
	SwaggerURL() string

	// IsEnabled returns whether the module is enabled
	IsEnabled() bool
}

// ModuleRegistry manages all registered modules
type ModuleRegistry struct {
	modules []ModuleService
}

// NewModuleRegistry creates a new module registry
func NewModuleRegistry() *ModuleRegistry {
	return &ModuleRegistry{
		modules: make([]ModuleService, 0),
	}
}

// Register adds a module to the registry
func (r *ModuleRegistry) Register(module ModuleService) {
	if module.IsEnabled() {
		r.modules = append(r.modules, module)
	}
}

// GetModules returns all registered modules
func (r *ModuleRegistry) GetModules() []ModuleService {
	return r.modules
}

// RegisterAllRoutes registers routes for all modules
func (r *ModuleRegistry) RegisterAllRoutes(router *mux.Router) {
	for _, module := range r.modules {
		if !module.IsEnabled() {
			continue
		}

		// Create subrouter with module prefix
		subrouter := router.PathPrefix("/" + module.ModuleName()).Subrouter()
		module.RegisterRoutes(subrouter)
	}
}

// GetSwaggerURLs returns all Swagger URLs for integration
func (r *ModuleRegistry) GetSwaggerURLs() map[string]string {
	urls := make(map[string]string)
	for _, module := range r.modules {
		if module.IsEnabled() && module.SwaggerURL() != "" {
			urls[module.ModuleName()] = module.SwaggerURL()
		}
	}
	return urls
}