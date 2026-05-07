package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
	"studentstorage/internal/db"
)

type KeyPress struct {
	ID         int       `json:"id"`
	SessionID  string    `json:"session_id"`
	KeyCode    int       `json:"key_code"`
	KeyName    string    `json:"key_name"`
	Modifiers  []string  `json:"modifiers"`
	IsCombo    bool      `json:"is_combo"`
	Ts         time.Time `json:"ts"`
	ReceivedAt time.Time `json:"received_at"`
	Email      string    `json:"email"`
}

func marshalModifiers(m []string) string {
	b, _ := json.Marshal(m)
	return string(b)
}

func unmarshalModifiers(s string) []string {
	var m []string
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		return []string{}
	}
	return m
}

func GetAllKeyPressesHandler(dbService *db.Service) http.HandlerFunc {
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
			"SELECT COUNT(*) FROM key_presses WHERE session_id = $1 AND email = $2", 
			sessionID, email).Scan(&total)
		if err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			log.Printf("[DB] count error: %v", err)
			return
		}

		rows, err := dbService.DB.QueryContext(r.Context(), 
			`SELECT id, session_id, key_code, key_name, modifiers, is_combo, ts, received_at, email 
			FROM key_presses 
			WHERE session_id = $1 AND email = $2 
			ORDER BY ts DESC 
			LIMIT $3 OFFSET $4`, 
			sessionID, email, pagination.Limit, pagination.Offset)
		if err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			log.Printf("[DB] query error: %v", err)
			return
		}
		defer rows.Close()

		result := make([]KeyPress, 0)
		for rows.Next() {
			var kp KeyPress
			var modifiersJSON string
			if err := rows.Scan(&kp.ID, &kp.SessionID, &kp.KeyCode, &kp.KeyName, &modifiersJSON, &kp.IsCombo, &kp.Ts, &kp.ReceivedAt, &kp.Email); err != nil {
				http.Error(w, "scan error", http.StatusInternalServerError)
				log.Printf("[DB] scan error: %v", err)
				return
			}
			kp.Modifiers = unmarshalModifiers(modifiersJSON)
			result = append(result, kp)
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