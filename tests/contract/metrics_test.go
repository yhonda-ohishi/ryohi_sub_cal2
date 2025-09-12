package contract

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetricsEndpoint_Contract(t *testing.T) {
	// This test MUST fail initially (TDD - RED phase)
	// It tests the contract defined in openapi.yaml for GET /metrics
	
	tests := []struct {
		name           string
		expectedStatus int
		validateBody   func(t *testing.T, body string)
	}{
		{
			name:           "returns 200 with Prometheus metrics",
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body string) {
				// Validate Prometheus format
				assert.Contains(t, body, "# HELP", "should contain HELP comments")
				assert.Contains(t, body, "# TYPE", "should contain TYPE comments")
				
				// Check for standard metrics
				assert.Contains(t, body, "http_requests_total", 
					"should contain http_requests_total metric")
				assert.Contains(t, body, "http_request_duration_seconds", 
					"should contain http_request_duration_seconds metric")
				assert.Contains(t, body, "http_requests_in_flight", 
					"should contain http_requests_in_flight metric")
				
				// Check for custom metrics
				assert.Contains(t, body, "backend_health_status", 
					"should contain backend_health_status metric")
				
				// Validate metric format (basic check)
				lines := strings.Split(body, "\n")
				for _, line := range lines {
					if strings.HasPrefix(line, "#") || line == "" {
						continue // Skip comments and empty lines
					}
					// Basic validation: metric lines should contain metric name and value
					parts := strings.Fields(line)
					assert.GreaterOrEqual(t, len(parts), 2, 
						"metric line should have at least name and value: %s", line)
				}
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
			w := httptest.NewRecorder()
			
			// Get router
			router := setupTestMetricsRouter()
			
			// Serve the request
			router.ServeHTTP(w, req)
			
			// Validate response
			assert.Equal(t, tt.expectedStatus, w.Code, "unexpected status code")
			
			// Validate response body
			if tt.validateBody != nil {
				tt.validateBody(t, w.Body.String())
			}
			
			// Validate headers
			contentType := w.Header().Get("Content-Type")
			assert.Contains(t, contentType, "text/plain", 
				"Content-Type should be text/plain for Prometheus metrics")
		})
	}
}

func TestMetricsEndpoint_NoAuthentication(t *testing.T) {
	// Metrics endpoint should not require authentication (for Prometheus scraping)
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()
	
	router := setupTestMetricsRouter()
	router.ServeHTTP(w, req)
	
	// Should not return 401 Unauthorized
	assert.NotEqual(t, http.StatusUnauthorized, w.Code, 
		"metrics endpoint should not require authentication for Prometheus scraping")
}

func TestMetricsEndpoint_Methods(t *testing.T) {
	// Test that only GET method is allowed
	methods := []string{
		http.MethodPost,
		http.MethodPut,
		http.MethodDelete,
		http.MethodPatch,
	}
	
	router := setupTestMetricsRouter()
	
	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/metrics", nil)
			w := httptest.NewRecorder()
			
			router.ServeHTTP(w, req)
			
			assert.Equal(t, http.StatusMethodNotAllowed, w.Code,
				"only GET method should be allowed for /metrics")
		})
	}
}

func TestMetricsEndpoint_MetricLabels(t *testing.T) {
	// Test that metrics include appropriate labels
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()
	
	router := setupTestMetricsRouter()
	router.ServeHTTP(w, req)
	
	body := w.Body.String()
	
	// Check for labeled metrics
	assert.Contains(t, body, `method="`, "metrics should include method label")
	assert.Contains(t, body, `status="`, "metrics should include status label")
	assert.Contains(t, body, `path="`, "metrics should include path label")
}