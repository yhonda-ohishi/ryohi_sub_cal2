package contract

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RouteConfig represents the route configuration structure
type RouteConfig struct {
	ID         string           `json:"id"`
	Path       string           `json:"path"`
	Method     []string         `json:"method"`
	Backend    string           `json:"backend"`
	Timeout    int64            `json:"timeout,omitempty"`
	RateLimit  *RateLimitConfig `json:"rate_limit,omitempty"`
	Auth       *AuthConfig      `json:"auth,omitempty"`
	Middleware []string         `json:"middleware,omitempty"`
	Priority   int              `json:"priority,omitempty"`
	Enabled    bool             `json:"enabled"`
}

type RateLimitConfig struct {
	Enabled   bool     `json:"enabled"`
	Rate      int      `json:"rate"`
	Period    string   `json:"period"`
	BurstSize int      `json:"burst_size,omitempty"`
	KeyType   string   `json:"key_type,omitempty"`
	WhiteList []string `json:"white_list,omitempty"`
}

type AuthConfig struct {
	Enabled  bool     `json:"enabled"`
	Type     string   `json:"type"`
	Required bool     `json:"required"`
	Roles    []string `json:"roles,omitempty"`
}

func TestAdminRoutesEndpoint_GetAll(t *testing.T) {
	// Test GET /admin/routes
	tests := []struct {
		name           string
		apiKey         string
		expectedStatus int
		validateBody   func(t *testing.T, body []byte)
	}{
		{
			name:           "returns 200 with valid API key",
			apiKey:         "valid-api-key",
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body []byte) {
				var routes []RouteConfig
				err := json.Unmarshal(body, &routes)
				require.NoError(t, err, "response should be valid JSON array")
				
				// Validate route structure if routes exist
				for _, route := range routes {
					assert.NotEmpty(t, route.ID, "route ID should not be empty")
					assert.NotEmpty(t, route.Path, "route path should not be empty")
					assert.NotEmpty(t, route.Method, "route methods should not be empty")
					assert.NotEmpty(t, route.Backend, "route backend should not be empty")
				}
			},
		},
		{
			name:           "returns 401 without API key",
			apiKey:         "",
			expectedStatus: http.StatusUnauthorized,
			validateBody:   nil,
		},
		{
			name:           "returns 401 with invalid API key",
			apiKey:         "invalid-key",
			expectedStatus: http.StatusUnauthorized,
			validateBody:   nil,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/admin/routes", nil)
			if tt.apiKey != "" {
				req.Header.Set("X-API-Key", tt.apiKey)
			}
			w := httptest.NewRecorder()
			
			router := setupTestAdminRouter()
			router.ServeHTTP(w, req)
			
			assert.Equal(t, tt.expectedStatus, w.Code, "unexpected status code")
			
			if tt.validateBody != nil {
				tt.validateBody(t, w.Body.Bytes())
			}
		})
	}
}

func TestAdminRoutesEndpoint_Create(t *testing.T) {
	// Test POST /admin/routes
	newRoute := RouteConfig{
		ID:      "test-route",
		Path:    "/test/*",
		Method:  []string{"GET", "POST"},
		Backend: "test-backend",
		Timeout: 30000000000, // 30 seconds in nanoseconds
		Enabled: true,
		Priority: 100,
	}
	
	tests := []struct {
		name           string
		apiKey         string
		payload        interface{}
		expectedStatus int
		validateBody   func(t *testing.T, body []byte)
	}{
		{
			name:           "creates route with valid data",
			apiKey:         "valid-api-key",
			payload:        newRoute,
			expectedStatus: http.StatusCreated,
			validateBody: func(t *testing.T, body []byte) {
				var route RouteConfig
				err := json.Unmarshal(body, &route)
				require.NoError(t, err, "response should be valid JSON")
				
				assert.Equal(t, newRoute.ID, route.ID, "route ID should match")
				assert.Equal(t, newRoute.Path, route.Path, "route path should match")
				assert.Equal(t, newRoute.Backend, route.Backend, "route backend should match")
			},
		},
		{
			name:           "returns 400 with invalid data",
			apiKey:         "valid-api-key",
			payload:        map[string]string{"invalid": "data"},
			expectedStatus: http.StatusBadRequest,
			validateBody:   nil,
		},
		{
			name:           "returns 401 without API key",
			apiKey:         "",
			payload:        newRoute,
			expectedStatus: http.StatusUnauthorized,
			validateBody:   nil,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest(http.MethodPost, "/admin/routes", 
				bytes.NewReader(payload))
			req.Header.Set("Content-Type", "application/json")
			if tt.apiKey != "" {
				req.Header.Set("X-API-Key", tt.apiKey)
			}
			w := httptest.NewRecorder()
			
			router := setupTestAdminRouter()
			router.ServeHTTP(w, req)
			
			assert.Equal(t, tt.expectedStatus, w.Code, "unexpected status code")
			
			if tt.validateBody != nil {
				tt.validateBody(t, w.Body.Bytes())
			}
		})
	}
}

func TestAdminRoutesEndpoint_GetByID(t *testing.T) {
	// Test GET /admin/routes/{routeId}
	tests := []struct {
		name           string
		routeID        string
		apiKey         string
		expectedStatus int
	}{
		{
			name:           "returns 200 for existing route",
			routeID:        "test-route",
			apiKey:         "valid-api-key",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "returns 404 for non-existing route",
			routeID:        "non-existing",
			apiKey:         "valid-api-key",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "returns 401 without API key",
			routeID:        "existing-route",
			apiKey:         "",
			expectedStatus: http.StatusUnauthorized,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, 
				"/admin/routes/"+tt.routeID, nil)
			if tt.apiKey != "" {
				req.Header.Set("X-API-Key", tt.apiKey)
			}
			w := httptest.NewRecorder()
			
			router := setupTestAdminRouter()
			router.ServeHTTP(w, req)
			
			assert.Equal(t, tt.expectedStatus, w.Code, "unexpected status code")
		})
	}
}

func TestAdminRoutesEndpoint_Update(t *testing.T) {
	// Test PUT /admin/routes/{routeId}
	updateRoute := RouteConfig{
		ID:      "test-route",
		Path:    "/test/v2/*",
		Method:  []string{"GET", "POST", "PUT"},
		Backend: "test-backend-v2",
		Timeout: 60000000000, // 60 seconds in nanoseconds
		Enabled: true,
	}
	
	payload, _ := json.Marshal(updateRoute)
	req := httptest.NewRequest(http.MethodPut, "/admin/routes/test-route", 
		bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "valid-api-key")
	w := httptest.NewRecorder()
	
	router := setupTestAdminRouter()
	router.ServeHTTP(w, req)
	
	// Initially will fail as router is not implemented
	assert.Equal(t, http.StatusOK, w.Code, "should return 200 for successful update")
}

func TestAdminRoutesEndpoint_Delete(t *testing.T) {
	// Test DELETE /admin/routes/{routeId}
	req := httptest.NewRequest(http.MethodDelete, "/admin/routes/test-route", nil)
	req.Header.Set("X-API-Key", "valid-api-key")
	w := httptest.NewRecorder()
	
	router := setupTestAdminRouter()
	router.ServeHTTP(w, req)
	
	// Initially will fail as router is not implemented
	assert.Equal(t, http.StatusNoContent, w.Code, 
		"should return 204 for successful deletion")
}