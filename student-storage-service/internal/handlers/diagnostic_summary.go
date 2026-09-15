package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"studentstorage/internal/db"
)

type DiagnosticSummary struct {
	Email     string `json:"email"`
	CritCount int    `json:"crit"`
	WarnCount int    `json:"warn"`
	OKCount   int    `json:"ok"`
}

func GetDiagnosticSummaryHandler(dbService *db.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID := r.URL.Query().Get("session_id")
		if sessionID == "" {
			http.Error(w, "session_id parameter is required", http.StatusBadRequest)
			return
		}

		rows, err := dbService.DB.QueryContext(r.Context(), `
			WITH worst_per_code AS (
				SELECT
					email,
					code,
					MAX(CASE WHEN status = 'CRIT' THEN 3
					         WHEN status = 'WARN' THEN 2
					         WHEN status = 'OK'   THEN 1
					         ELSE 0 END) AS severity
				FROM diagnostics
				WHERE session_id = $1
				GROUP BY email, code
			)
			SELECT
				email,
				COUNT(*) FILTER (WHERE severity = 3) AS crit_count,
				COUNT(*) FILTER (WHERE severity = 2) AS warn_count,
				COUNT(*) FILTER (WHERE severity = 1) AS ok_count
			FROM worst_per_code
			GROUP BY email
		`, sessionID)
		if err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			log.Printf("[DB] diagnostic summary error: %v", err)
			return
		}
		defer rows.Close()

		result := make([]DiagnosticSummary, 0)
		for rows.Next() {
			var s DiagnosticSummary
			if err := rows.Scan(&s.Email, &s.CritCount, &s.WarnCount, &s.OKCount); err != nil {
				http.Error(w, "scan error", http.StatusInternalServerError)
				log.Printf("[DB] scan error: %v", err)
				return
			}
			result = append(result, s)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}
