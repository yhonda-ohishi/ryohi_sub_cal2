package contract

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// HealthResponse represents the expected health check response structure
type HealthResponse struct {
	Status    string                       `json:"status"`
	Timestamp string                       `json:"timestamp"`
	Services  map[string]ServiceHealthInfo `json:"services,omitempty"`
}

type ServiceHealthInfo struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

func TestHealthEndpoint_Contract(t *testing.T) {
	// This test MUST fail initially (TDD - RED phase)
	// It tests the contract defined in openapi.yaml for GET /health
	
	tests := []struct {
		name           string
		expectedStatus int
		validateBody   func(t *testing.T, body []byte)
	}{
		{
			name:           "returns 200 when service is healthy",
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body []byte) {
				var response HealthResponse
				err := json.Unmarshal(body, &response)
				require.NoError(t, err, "response should be valid JSON")
				
				// Validate required fields
				assert.Equal(t, "healthy", response.Status, "status should be 'healthy'")
				assert.NotEmpty(t, response.Timestamp, "timestamp should not be empty")
				
				// Validate timestamp format (RFC3339)
				_, err = time.Parse(time.RFC3339, response.Timestamp)
				assert.NoError(t, err, "timestamp should be in RFC3339 format")
				
				// If services are present, validate their structure
				if response.Services != nil {
					for serviceName, serviceInfo := range response.Services {
						assert.NotEmpty(t, serviceName, "service name should not be empty")
						assert.Contains(t, []string{"healthy", "unhealthy", "unknown"}, 
							serviceInfo.Status, "service status should be valid")
					}
				}
			},
		},
		{
			name:           "returns 503 when service is unhealthy",
			expectedStatus: http.StatusServiceUnavailable,
			validateBody: func(t *testing.T, body []byte) {
				var response HealthResponse
				err := json.Unmarshal(body, &response)
				require.NoError(t, err, "response should be valid JSON")
				
				assert.Equal(t, "unhealthy", response.Status, "status should be 'unhealthy'")
				assert.NotEmpty(t, response.Timestamp, "timestamp should not be empty")
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest(http.MethodGet, "/health", nil)
			w := httptest.NewRecorder()
			
			// Get router (this will fail initially as router is not implemented)
			router := setupTestRouter()
			
			// Serve the request
			router.ServeHTTP(w, req)
			
			// Validate response
			assert.Equal(t, tt.expectedStatus, w.Code, "unexpected status code")
			
			// Validate response body
			if tt.validateBody != nil {
				tt.validateBody(t, w.Body.Bytes())
			}
			
			// Validate headers
			contentType := w.Header().Get("Content-Type")
			assert.Equal(t, "application/json", contentType, "Content-Type should be application/json")
		})
	}
}

func TestHealthEndpoint_NoAuthentication(t *testing.T) {
	// Health endpoint should not require authentication
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	
	router := setupTestRouter()
	router.ServeHTTP(w, req)
	
	// Should not return 401 Unauthorized
	assert.NotEqual(t, http.StatusUnauthorized, w.Code, 
		"health endpoint should not require authentication")
}

func TestHealthEndpoint_Methods(t *testing.T) {
	// Test that only GET method is allowed
	methods := []string{
		http.MethodPost,
		http.MethodPut,
		http.MethodDelete,
		http.MethodPatch,
	}
	
	router := setupTestRouter()
	
	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/health", nil)
			w := httptest.NewRecorder()
			
			router.ServeHTTP(w, req)
			
			assert.Equal(t, http.StatusMethodNotAllowed, w.Code,
				"only GET method should be allowed for /health")
		})
	}
}