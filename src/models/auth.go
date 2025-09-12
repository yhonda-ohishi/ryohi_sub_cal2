package models

import (
	"fmt"
)

// AuthConfig represents authentication configuration
type AuthConfig struct {
	Enabled  bool     `json:"enabled" yaml:"enabled"`
	Type     string   `json:"type" yaml:"type"`
	Required bool     `json:"required" yaml:"required"`
	Roles    []string `json:"roles,omitempty" yaml:"roles,omitempty"`
}

// Validate validates the authentication configuration
func (a *AuthConfig) Validate() error {
	if !a.Enabled {
		return nil
	}
	
	validTypes := []string{"none", "basic", "bearer", "api-key", "jwt", "oauth2"}
	valid := false
	for _, t := range validTypes {
		if a.Type == t {
			valid = true
			break
		}
	}
	
	if !valid {
		return fmt.Errorf("invalid auth type: %s", a.Type)
	}
	
	if a.Type == "none" && a.Required {
		return fmt.Errorf("auth type 'none' cannot be required")
	}
	
	return nil
}

// RequiresRole checks if a specific role is required
func (a *AuthConfig) RequiresRole(role string) bool {
	if !a.Enabled || len(a.Roles) == 0 {
		return false
	}
	
	for _, r := range a.Roles {
		if r == role {
			return true
		}
	}
	
	return false
}

// AuthRequest represents an authentication request
type AuthRequest struct {
	Type        string            `json:"type"`
	Credentials map[string]string `json:"credentials"`
	Token       string            `json:"token,omitempty"`
}

// AuthResponse represents an authentication response
type AuthResponse struct {
	Authenticated bool     `json:"authenticated"`
	UserID        string   `json:"user_id,omitempty"`
	Username      string   `json:"username,omitempty"`
	Roles         []string `json:"roles,omitempty"`
	ExpiresAt     int64    `json:"expires_at,omitempty"`
	Error         string   `json:"error,omitempty"`
}

// AuthContext represents the authentication context for a request
type AuthContext struct {
	Authenticated bool              `json:"authenticated"`
	UserID        string            `json:"user_id,omitempty"`
	Username      string            `json:"username,omitempty"`
	Roles         []string          `json:"roles,omitempty"`
	Method        string            `json:"method,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

// HasRole checks if the auth context has a specific role
func (ac *AuthContext) HasRole(role string) bool {
	for _, r := range ac.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// HasAnyRole checks if the auth context has any of the specified roles
func (ac *AuthContext) HasAnyRole(roles []string) bool {
	for _, role := range roles {
		if ac.HasRole(role) {
			return true
		}
	}
	return false
}

// APIKeyConfig represents API key configuration
type APIKeyConfig struct {
	Enabled    bool              `json:"enabled" yaml:"enabled"`
	HeaderName string            `json:"header_name" yaml:"header_name"`
	QueryParam string            `json:"query_param" yaml:"query_param"`
	Keys       map[string]APIKey `json:"keys" yaml:"keys"`
}

// APIKey represents an API key
type APIKey struct {
	Key         string   `json:"key" yaml:"key"`
	Name        string   `json:"name" yaml:"name"`
	Roles       []string `json:"roles,omitempty" yaml:"roles,omitempty"`
	RateLimitID string   `json:"rate_limit_id,omitempty" yaml:"rate_limit_id,omitempty"`
	Enabled     bool     `json:"enabled" yaml:"enabled"`
}

// Validate validates the API key configuration
func (a *APIKeyConfig) Validate() error {
	if !a.Enabled {
		return nil
	}
	
	if a.HeaderName == "" {
		a.HeaderName = "X-API-Key" // Default header name
	}
	
	if len(a.Keys) == 0 {
		return fmt.Errorf("at least one API key must be configured when API key auth is enabled")
	}
	
	for id, key := range a.Keys {
		if key.Key == "" {
			return fmt.Errorf("API key value is required for key ID: %s", id)
		}
		if key.Name == "" {
			return fmt.Errorf("API key name is required for key ID: %s", id)
		}
	}
	
	return nil
}

// ValidateKey validates an API key and returns the associated key info
func (a *APIKeyConfig) ValidateKey(key string) (*APIKey, error) {
	if !a.Enabled {
		return nil, fmt.Errorf("API key authentication is not enabled")
	}
	
	for _, apiKey := range a.Keys {
		if apiKey.Key == key && apiKey.Enabled {
			return &apiKey, nil
		}
	}
	
	return nil, fmt.Errorf("invalid or disabled API key")
}

// JWTConfig represents JWT configuration
type JWTConfig struct {
	Enabled       bool   `json:"enabled" yaml:"enabled"`
	Secret        string `json:"secret" yaml:"secret"`
	Issuer        string `json:"issuer" yaml:"issuer"`
	Audience      string `json:"audience" yaml:"audience"`
	Algorithm     string `json:"algorithm" yaml:"algorithm"`
	ExpiryMinutes int    `json:"expiry_minutes" yaml:"expiry_minutes"`
}

// Validate validates the JWT configuration
func (j *JWTConfig) Validate() error {
	if !j.Enabled {
		return nil
	}
	
	if j.Secret == "" {
		return fmt.Errorf("JWT secret is required")
	}
	
	if j.Algorithm == "" {
		j.Algorithm = "HS256" // Default algorithm
	}
	
	validAlgorithms := []string{"HS256", "HS384", "HS512", "RS256", "RS384", "RS512"}
	valid := false
	for _, algo := range validAlgorithms {
		if j.Algorithm == algo {
			valid = true
			break
		}
	}
	
	if !valid {
		return fmt.Errorf("invalid JWT algorithm: %s", j.Algorithm)
	}
	
	if j.ExpiryMinutes <= 0 {
		j.ExpiryMinutes = 60 // Default to 1 hour
	}
	
	return nil
}