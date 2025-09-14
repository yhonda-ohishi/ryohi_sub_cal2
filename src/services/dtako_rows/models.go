package dtako_rows

import "time"

// DtakoRow represents vehicle operation data
type DtakoRow struct {
	ID         string    `json:"id" example:"row-123"`
	UnkoNo     string    `json:"unko_no" example:"2025010101"` // 運行NO
	Date       time.Time `json:"date" example:"2025-01-13T00:00:00Z"`
	VehicleNo  string    `json:"vehicle_no" example:"vehicle-001"`
	DriverCode string    `json:"driver_code" example:"driver-123"`
	RouteCode  string    `json:"route_code" example:"route-A"`
	Distance   float64   `json:"distance" example:"123.45"`
	FuelAmount float64   `json:"fuel_amount" example:"45.67"`
	CreatedAt  time.Time `json:"created_at" example:"2025-01-13T15:04:05Z"`
	UpdatedAt  time.Time `json:"updated_at" example:"2025-01-13T15:04:05Z"`
}

// ImportRequest represents an import request
type ImportRequest struct {
	FromDate string `json:"from_date" example:"2025-01-01"`
	ToDate   string `json:"to_date" example:"2025-01-31"`
}

// ImportResult represents the result of an import operation
type ImportResult struct {
	Success      bool      `json:"success" example:"true"`
	ImportedRows int       `json:"imported_rows" example:"150"`
	Message      string    `json:"message" example:"Imported 150 rows successfully"`
	ImportedAt   time.Time `json:"imported_at" example:"2025-01-13T15:04:05Z"`
	Errors       []string  `json:"errors,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Code    int    `json:"code" example:"400"`
	Message string `json:"message" example:"Invalid request parameters"`
}