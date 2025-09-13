package models

import (
	"fmt"
	"regexp"
	"time"
)

// RouteConfig represents a routing configuration
type RouteConfig struct {
	ID         string           `json:"id" yaml:"id"`
	Path       string           `json:"path" yaml:"path"`
	Method     []string         `json:"method" yaml:"method"`
	Backend    string           `json:"backend" yaml:"backend"`
	Timeout    time.Duration    `json:"timeout" yaml:"timeout" swaggertype:"integer" example:"30000000000"`
	RateLimit  *RateLimitConfig `json:"rate_limit,omitempty" yaml:"rate_limit,omitempty"`
	Auth       *AuthConfig      `json:"auth,omitempty" yaml:"auth,omitempty"`
	Middleware []string         `json:"middleware,omitempty" yaml:"middleware,omitempty"`
	Priority   int              `json:"priority" yaml:"priority"`
	Enabled    bool             `json:"enabled" yaml:"enabled"`
	CreatedAt  time.Time        `json:"created_at" yaml:"created_at"`
	UpdatedAt  time.Time        `json:"updated_at" yaml:"updated_at"`
}

// Validate validates the route configuration
func (r *RouteConfig) Validate() error {
	if r.ID == "" {
		return fmt.Errorf("route ID is required")
	}
	
	if r.Path == "" {
		return fmt.Errorf("route path is required")
	}
	
	if !isValidPath(r.Path) {
		return fmt.Errorf("invalid route path: %s", r.Path)
	}
	
	if len(r.Method) == 0 {
		return fmt.Errorf("at least one HTTP method is required")
	}
	
	for _, method := range r.Method {
		if !isValidHTTPMethod(method) {
			return fmt.Errorf("invalid HTTP method: %s", method)
		}
	}
	
	if r.Backend == "" {
		return fmt.Errorf("backend service ID is required")
	}
	
	if r.Timeout == 0 {
		r.Timeout = 30 * time.Second // Default timeout
	} else if r.Timeout > 5*time.Minute {
		return fmt.Errorf("timeout cannot exceed 5 minutes")
	}
	
	if r.Priority < 0 || r.Priority > 1000 {
		return fmt.Errorf("priority must be between 0 and 1000")
	}
	
	if r.RateLimit != nil {
		if err := r.RateLimit.Validate(); err != nil {
			return fmt.Errorf("invalid rate limit config: %w", err)
		}
	}
	
	if r.Auth != nil {
		if err := r.Auth.Validate(); err != nil {
			return fmt.Errorf("invalid auth config: %w", err)
		}
	}
	
	return nil
}

// Match checks if the given path and method match this route
func (r *RouteConfig) Match(path, method string) bool {
	if !r.Enabled {
		return false
	}
	
	// Check method
	methodMatch := false
	for _, m := range r.Method {
		if m == method || m == "*" {
			methodMatch = true
			break
		}
	}
	if !methodMatch {
		return false
	}
	
	// Check path
	return matchPath(r.Path, path)
}

// matchPath checks if a path pattern matches a given path
func matchPath(pattern, path string) bool {
	// Convert wildcard pattern to regex
	// /api/* -> /api/.*
	// /api/*/users -> /api/.*/users
	regexPattern := regexp.QuoteMeta(pattern)
	regexPattern = "^" + regexPattern + "$"
	regexPattern = regexp.MustCompile(`\\\*`).ReplaceAllString(regexPattern, ".*")
	
	matched, _ := regexp.MatchString(regexPattern, path)
	return matched
}

// isValidPath checks if the path is valid
func isValidPath(path string) bool {
	if path == "" || path[0] != '/' {
		return false
	}
	// Basic validation - can be extended
	return true
}

// isValidHTTPMethod checks if the method is a valid HTTP method
func isValidHTTPMethod(method string) bool {
	validMethods := []string{
		"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS", "CONNECT", "TRACE", "*",
	}
	for _, m := range validMethods {
		if m == method {
			return true
		}
	}
	return false
}

// RouteCollection represents a collection of routes
type RouteCollection struct {
	Routes []*RouteConfig `json:"routes" yaml:"routes"`
}

// FindRoute finds the best matching route for a given path and method
func (rc *RouteCollection) FindRoute(path, method string) *RouteConfig {
	var bestMatch *RouteConfig
	bestPriority := -1
	
	for _, route := range rc.Routes {
		if route.Match(path, method) {
			if route.Priority > bestPriority {
				bestMatch = route
				bestPriority = route.Priority
			}
		}
	}
	
	return bestMatch
}