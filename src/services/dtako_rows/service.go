package dtako_rows

import (
	"encoding/json"
	"net/http"
	"github.com/gorilla/mux"
	"time"
)

// Service represents the dtako_rows service
type Service struct {
	enabled bool
}

// NewService creates a new dtako_rows service
func NewService(enabled bool) *Service {
	return &Service{enabled: enabled}
}

// RegisterRoutes registers all routes for this module
func (s *Service) RegisterRoutes(router *mux.Router) {
	if !s.enabled {
		return
	}
	router.HandleFunc("/rows", s.listRows).Methods("GET")
	router.HandleFunc("/rows/{id}", s.getRow).Methods("GET")
	router.HandleFunc("/rows/import", s.importRows).Methods("POST")
}

// ModuleName returns the module name for path prefix
func (s *Service) ModuleName() string {
	return "dtako_rows"
}

// SwaggerURL returns the Swagger documentation URL
func (s *Service) SwaggerURL() string {
	// 将来的にSwagger統合時に使用
	return ""
}

// IsEnabled returns whether the module is enabled
func (s *Service) IsEnabled() bool {
	return s.enabled
}

// Handler methods
func (s *Service) listRows(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// クエリパラメータ取得
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	// サンプルレスポンス
	rows := []DtakoRow{
		{
			ID:         "row-001",
			UnkoNo:     "2025010101",
			Date:       time.Now(),
			VehicleNo:  "vehicle-001",
			DriverCode: "driver-123",
			RouteCode:  "route-A",
			Distance:   123.45,
			FuelAmount: 45.67,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
	}

	response := map[string]interface{}{
		"rows": rows,
		"from": from,
		"to":   to,
		"count": len(rows),
	}

	json.NewEncoder(w).Encode(response)
}

func (s *Service) getRow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(DtakoRow{
		ID:         id,
		UnkoNo:     "2025010101",
		Date:       time.Now(),
		VehicleNo:  "vehicle-001",
		DriverCode: "driver-123",
		RouteCode:  "route-A",
		Distance:   123.45,
		FuelAmount: 45.67,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	})
}

func (s *Service) importRows(w http.ResponseWriter, r *http.Request) {
	var req ImportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Code:    400,
			Message: "Invalid request body",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ImportResult{
		Success:      true,
		ImportedRows: 150,
		Message:      "Imported 150 rows successfully",
		ImportedAt:   time.Now(),
	})
}