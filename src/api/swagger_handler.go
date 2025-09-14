package api

import (
	"encoding/json"
	"io/ioutil"
	"log/slog"
	"net/http"
	"path/filepath"

	"github.com/your-org/ryohi-router/src/lib/swagger"
)

// CustomSwaggerHandler creates a custom swagger doc handler with DTako microservices integration
func CustomSwaggerHandler(logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Integrate DTako microservices swagger on-demand
		swaggerMerger := swagger.NewSwaggerMerger("docs", logger)
		if err := swaggerMerger.MergeOnStartup(); err != nil {
			logger.Warn("Failed to integrate DTako microservices Swagger in handler", "error", err)
		}

		// Read the merged swagger file
		swaggerPath := filepath.Join("docs", "swagger.json")
		data, err := ioutil.ReadFile(swaggerPath)
		if err != nil {
			http.Error(w, "Failed to read swagger file", http.StatusInternalServerError)
			return
		}

		// Validate JSON
		var swaggerDoc map[string]interface{}
		if err := json.Unmarshal(data, &swaggerDoc); err != nil {
			http.Error(w, "Invalid swagger JSON", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	}
}