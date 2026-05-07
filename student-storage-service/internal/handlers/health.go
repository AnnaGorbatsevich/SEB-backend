package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"studentstorage/internal/db"
)

type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Database  string    `json:"database,omitempty"`
}

func HealthHandler(dbService *db.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		dbStatus := "unhealthy"
		if err := dbService.DB.PingContext(ctx); err == nil {
			dbStatus = "healthy"
		}

		w.Header().Set("Content-Type", "application/json")

		if dbStatus == "unhealthy" {
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}

		response := HealthResponse{
			Status:    dbStatus,
			Timestamp: time.Now().UTC(),
			Database:  dbStatus,
		}

		json.NewEncoder(w).Encode(response)
	}
}
