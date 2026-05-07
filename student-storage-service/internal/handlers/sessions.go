package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"
	"studentstorage/internal/db"
)

type SessionEvent struct {
	Email     string    `json:"email"`
	EventType string    `json:"event_type"`
	EventTime time.Time `json:"event_time"`
}

func GetSessionEventsHandler(dbService *db.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID := r.URL.Query().Get("session_id")
		if sessionID == "" {
			http.Error(w, "session_id parameter is required", http.StatusBadRequest)
			return
		}

		query := `
			WITH all_events AS (
          SELECT email, ts, 'cursor_position' AS event_type FROM cursor_positions WHERE session_id = $1
          UNION ALL
          SELECT email, ts, 'key_press'       AS event_type FROM key_presses     WHERE session_id = $1
          UNION ALL
          SELECT email, ts, 'log'             AS event_type FROM logs            WHERE session_id = $1
      ),
      ranked_events AS (
          SELECT
              email,
              event_type,
              ts,
              ROW_NUMBER() OVER (PARTITION BY email ORDER BY ts DESC) as rn
          FROM all_events
      )
      SELECT email, event_type, ts AS event_time
      FROM ranked_events
      WHERE rn = 1
      ORDER BY event_time DESC;
		`

		rows, err := dbService.DB.QueryContext(r.Context(), query, sessionID)
		if err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			log.Printf("[DB] query error: %v", err)
			return
		}
		defer rows.Close()

		var results []SessionEvent
		for rows.Next() {
			var event SessionEvent
			if err := rows.Scan(&event.Email, &event.EventType, &event.EventTime); err != nil {
				if err == sql.ErrNoRows {
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode([]SessionEvent{})
					return
				}
				http.Error(w, "scan error", http.StatusInternalServerError)
				log.Printf("[DB] scan error: %v", err)
				return
			}
			results = append(results, event)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(results)
	}
}