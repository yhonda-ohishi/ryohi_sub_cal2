package integration

import (
	"encoding/json"
	"net/http"

	"github.com/yhonda-ohishi/dtako_mod/models"
)

// MockDtakoHandlers provides mock implementations for testing
type MockDtakoHandlers struct{}

// ListRows handles GET /dtako/rows
func (h *MockDtakoHandlers) ListRows(w http.ResponseWriter, r *http.Request) {
	// Return empty list for testing
	rows := []models.DtakoRow{}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rows)
}

// GetRowByID handles GET /dtako/rows/{id}
func (h *MockDtakoHandlers) GetRowByID(w http.ResponseWriter, r *http.Request) {
	row := models.DtakoRow{
		ID: "test-id-123",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(row)
}

// ImportRows handles POST /dtako/rows/import
func (h *MockDtakoHandlers) ImportRows(w http.ResponseWriter, r *http.Request) {
	result := models.ImportResult{
		Success:      true,
		ImportedRows: 0,
		Message:      "Test import successful",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// ListEvents handles GET /dtako/events
func (h *MockDtakoHandlers) ListEvents(w http.ResponseWriter, r *http.Request) {
	events := []models.DtakoEvent{}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

// ImportEvents handles POST /dtako/events/import
func (h *MockDtakoHandlers) ImportEvents(w http.ResponseWriter, r *http.Request) {
	result := models.ImportResult{
		Success:      true,
		ImportedRows: 0,
		Message:      "Test import successful",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// ListFerry handles GET /dtako/ferry
func (h *MockDtakoHandlers) ListFerry(w http.ResponseWriter, r *http.Request) {
	ferry := []models.DtakoFerry{}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ferry)
}

// ImportFerry handles POST /dtako/ferry/import
func (h *MockDtakoHandlers) ImportFerry(w http.ResponseWriter, r *http.Request) {
	result := models.ImportResult{
		Success:      true,
		ImportedRows: 0,
		Message:      "Test import successful",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}