package middleware

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/your-org/ryohi-router/src/models"
)

// Chain applies multiple middleware to a handler
func Chain(h http.Handler, middleware ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middleware) - 1; i >= 0; i-- {
		h = middleware[i](h)
	}
	return h
}

// RequestID adds a request ID to the context
func RequestID() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = uuid.New().String()
			}
			
			w.Header().Set("X-Request-ID", requestID)
			next.ServeHTTP(w, r)
		})
	}
}

// Logger logs HTTP requests
func Logger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			
			// Wrap response writer to capture status code
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			
			next.ServeHTTP(wrapped, r)
			
			duration := time.Since(start)
			
			logger.Info("HTTP Request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", wrapped.statusCode,
				"duration", duration.String(),
				"remote_addr", r.RemoteAddr,
			)
		})
	}
}

// Recovery recovers from panics
func Recovery(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.Error("Panic recovered",
						"error", err,
						"path", r.URL.Path,
						"method", r.Method,
					)
					
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()
			
			next.ServeHTTP(w, r)
		})
	}
}

// Metrics collects request metrics
func Metrics() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// TODO: Implement metrics collection
			next.ServeHTTP(w, r)
		})
	}
}

// RateLimit implements rate limiting
func RateLimit(config *models.RateLimitConfig) func(http.Handler) http.Handler {
	limiter := models.NewRateLimiter(config)
	
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract key based on key type
			var key string
			switch config.KeyType {
			case "IP":
				key = getClientIP(r)
			case "API_KEY":
				key = r.Header.Get("X-API-Key")
			default:
				key = "global"
			}
			
			if !limiter.Allow(key) {
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

// Auth implements authentication
func Auth(config *models.AuthConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !config.Enabled {
				next.ServeHTTP(w, r)
				return
			}
			
			// Simple auth check (to be expanded)
			var authenticated bool
			
			switch config.Type {
			case "bearer":
				token := r.Header.Get("Authorization")
				authenticated = strings.HasPrefix(token, "Bearer ")
			case "api-key":
				apiKey := r.Header.Get("X-API-Key")
				authenticated = apiKey != ""
			case "none":
				authenticated = true
			default:
				authenticated = false
			}
			
			if !authenticated && config.Required {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

// APIKeyAuth implements API key authentication for admin endpoints
func APIKeyAuth(validKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKey := r.Header.Get("X-API-Key")
			
			if apiKey != validKey {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// getClientIP extracts the client IP address from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	
	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}
	
	// Fall back to RemoteAddr
	addr := r.RemoteAddr
	if idx := strings.LastIndex(addr, ":"); idx != -1 {
		return addr[:idx]
	}
	
	return addr
}