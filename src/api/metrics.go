package api

import (
	"fmt"
	"net/http"
	"runtime"
	"time"
)

// MetricsHandler returns a handler for Prometheus metrics
func MetricsHandler() http.HandlerFunc {
	startTime := time.Now()
	
	return func(w http.ResponseWriter, r *http.Request) {
		// Set content type for Prometheus
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		
		// Collect runtime metrics
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		
		uptime := time.Since(startTime).Seconds()
		
		// Write Prometheus metrics
		fmt.Fprintf(w, "# HELP http_requests_total Total number of HTTP requests\n")
		fmt.Fprintf(w, "# TYPE http_requests_total counter\n")
		fmt.Fprintf(w, "http_requests_total{method=\"GET\",status=\"200\"} 0\n")
		fmt.Fprintf(w, "http_requests_total{method=\"POST\",status=\"200\"} 0\n")
		
		fmt.Fprintf(w, "\n# HELP http_request_duration_seconds HTTP request latency\n")
		fmt.Fprintf(w, "# TYPE http_request_duration_seconds histogram\n")
		fmt.Fprintf(w, "http_request_duration_seconds_bucket{method=\"GET\",path=\"/\",status=\"200\",le=\"0.1\"} 0\n")
		fmt.Fprintf(w, "http_request_duration_seconds_bucket{method=\"GET\",path=\"/\",status=\"200\",le=\"0.5\"} 0\n")
		fmt.Fprintf(w, "http_request_duration_seconds_bucket{method=\"GET\",path=\"/\",status=\"200\",le=\"1\"} 0\n")
		fmt.Fprintf(w, "http_request_duration_seconds_bucket{method=\"GET\",path=\"/\",status=\"200\",le=\"+Inf\"} 0\n")
		
		fmt.Fprintf(w, "\n# HELP http_requests_in_flight Current number of HTTP requests being served\n")
		fmt.Fprintf(w, "# TYPE http_requests_in_flight gauge\n")
		fmt.Fprintf(w, "http_requests_in_flight 0\n")
		
		fmt.Fprintf(w, "\n# HELP backend_health_status Health status of backend services (1=healthy, 0=unhealthy)\n")
		fmt.Fprintf(w, "# TYPE backend_health_status gauge\n")
		fmt.Fprintf(w, "backend_health_status{backend=\"example\"} 1\n")
		
		fmt.Fprintf(w, "\n# HELP process_resident_memory_bytes Resident memory size in bytes\n")
		fmt.Fprintf(w, "# TYPE process_resident_memory_bytes gauge\n")
		fmt.Fprintf(w, "process_resident_memory_bytes %d\n", m.Sys)
		
		fmt.Fprintf(w, "\n# HELP process_cpu_seconds_total Total user and system CPU time spent in seconds\n")
		fmt.Fprintf(w, "# TYPE process_cpu_seconds_total counter\n")
		fmt.Fprintf(w, "process_cpu_seconds_total 0\n")
		
		fmt.Fprintf(w, "\n# HELP process_start_time_seconds Start time of the process since unix epoch in seconds\n")
		fmt.Fprintf(w, "# TYPE process_start_time_seconds gauge\n")
		fmt.Fprintf(w, "process_start_time_seconds %f\n", float64(startTime.Unix()))
		
		fmt.Fprintf(w, "\n# HELP router_uptime_seconds Uptime of the router in seconds\n")
		fmt.Fprintf(w, "# TYPE router_uptime_seconds gauge\n")
		fmt.Fprintf(w, "router_uptime_seconds %f\n", uptime)
	}
}