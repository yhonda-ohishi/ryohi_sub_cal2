package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	port := flag.Int("port", 9002, "Port to listen on")
	flag.Parse()

	mux := http.NewServeMux()

	// Health endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "healthy",
			"timestamp": time.Now().Format(time.RFC3339),
			"service": fmt.Sprintf("mock-backend-port-%d", *port),
		})
	})

	// Echo endpoint
	mux.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"method": r.Method,
			"path": r.URL.Path,
			"query": r.URL.RawQuery,
			"headers": r.Header,
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	// Test endpoint
	mux.HandleFunc("/api/test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Hello from mock backend",
			"port": *port,
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	// Default handler
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Mock backend server",
			"port": *port,
			"path": r.URL.Path,
		})
	})

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Starting mock backend server on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}