package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"studentstorage/internal/db"
)

type Diagnostic struct {
	ID         int             `json:"id"`
	SessionID  string          `json:"session_id"`
	Code       string          `json:"code"`
	Status     string          `json:"status"`
	Details    json.RawMessage `json:"details"`
	ReceivedAt time.Time       `json:"received_at"`
	Email      string          `json:"email"`
}

func GetDiagnosticsHandler(dbService *db.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pagination := GetPagination(r)

		sessionID := r.URL.Query().Get("session_id")
		email := r.URL.Query().Get("email")

		if sessionID == "" {
			http.Error(w, "session_id parameter is required", http.StatusBadRequest)
			return
		}
		if email == "" {
			http.Error(w, "email parameter is required", http.StatusBadRequest)
			return
		}

		var total int64
		err := dbService.DB.QueryRowContext(r.Context(),
			"SELECT COUNT(*) FROM diagnostics WHERE session_id = $1 AND email = $2",
			sessionID, email).Scan(&total)
		if err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			log.Printf("[DB] error: %v", err)
			return
		}

		rows, err := dbService.DB.QueryContext(r.Context(),
			`SELECT id, session_id, code, status, details, received_at, email
			FROM diagnostics
			WHERE session_id = $1 AND email = $2
			ORDER BY received_at DESC
			LIMIT $3 OFFSET $4`,
			sessionID, email, pagination.Limit, pagination.Offset)
		if err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			log.Printf("[DB] query error: %v", err)
			return
		}
		defer rows.Close()

		result := make([]Diagnostic, 0)
		for rows.Next() {
			var d Diagnostic
			var detailsStr string
			if err := rows.Scan(&d.ID, &d.SessionID, &d.Code, &d.Status, &detailsStr, &d.ReceivedAt, &d.Email); err != nil {
				http.Error(w, "scan error", http.StatusInternalServerError)
				log.Printf("[DB] scan error: %v", err)
				return
			}
			d.Details = json.RawMessage(detailsStr)
			result = append(result, d)
		}

		paginatedResponse := PaginatedResponse{
			Data:    result,
			Total:   total,
			Limit:   pagination.Limit,
			Offset:  pagination.Offset,
			HasMore: pagination.Offset+pagination.Limit < int(total),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(paginatedResponse)
	}
}
