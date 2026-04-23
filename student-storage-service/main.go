package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	kafka "github.com/segmentio/kafka-go"
)

type TelemetryEvent struct {
	Event     string          `json:"event"`
	Timestamp string          `json:"timestamp"`
	SessionID string          `json:"sessionId"`
	Data      json.RawMessage `json:"data"`
}

type CursorData struct {
	X  float64 `json:"x"`
	Y  float64 `json:"y"`
	Ts string  `json:"ts"`
}

type KeyPressData struct {
	KeyCode   int      `json:"keyCode"`
	KeyName   string   `json:"keyName"`
	Modifiers []string `json:"modifiers"`
	IsCombo   bool     `json:"isCombo"`
	Ts        string   `json:"ts"`
}

type LogData struct {
	Level   string `json:"level"`
	Message string `json:"message"`
	Ts      string `json:"ts"`
}

type CursorPosition struct {
	ID         int       `json:"id"`
	SessionID  string    `json:"session_id"`
	X          float64   `json:"x"`
	Y          float64   `json:"y"`
	Ts         time.Time `json:"ts"`
	ReceivedAt time.Time `json:"received_at"`
}

type KeyPress struct {
	ID         int       `json:"id"`
	SessionID  string    `json:"session_id"`
	KeyCode    int       `json:"key_code"`
	KeyName    string    `json:"key_name"`
	Modifiers  []string  `json:"modifiers"`
	IsCombo    bool      `json:"is_combo"`
	Ts         time.Time `json:"ts"`
	ReceivedAt time.Time `json:"received_at"`
}

type Log struct {
	ID         int       `json:"id"`
	SessionID  string    `json:"session_id"`
	Level      string    `json:"level"`
	Message    string    `json:"message"`
	Ts         time.Time `json:"ts"`
	ReceivedAt time.Time `json:"received_at"`
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

var db *sql.DB

func initDB() {
	dsn := "postgres://postgres:postgres@postgres:5432/studentdata?sslmode=disable"

	var err error
	for i := 0; i < 10; i++ {
		db, err = sql.Open("pgx", dsn)
		if err == nil {
			if err = db.Ping(); err == nil {
				break
			}
		}
		log.Printf("[DB] connection attempt %d failed: %v", i+1, err)
		time.Sleep(3 * time.Second)
	}
	if err != nil {
		log.Fatalf("[DB] failed to connect: %v", err)
	}



	_, err = db.Exec(`
		CREATE TABLE cursor_positions (
			id          SERIAL PRIMARY KEY,
			session_id  TEXT             NOT NULL,
			x           DOUBLE PRECISION NOT NULL,
			y           DOUBLE PRECISION NOT NULL,
			ts          TIMESTAMPTZ      NOT NULL,
			received_at TIMESTAMPTZ      NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		log.Fatalf("[DB] failed to create cursor_positions table: %v", err)
	}


	_, err = db.Exec(`
		CREATE TABLE key_presses (
			id          SERIAL PRIMARY KEY,
			session_id  TEXT        NOT NULL,
			key_code    INTEGER     NOT NULL,
			key_name    TEXT        NOT NULL,
			modifiers   TEXT        NOT NULL DEFAULT '[]',
			is_combo    BOOLEAN     NOT NULL,
			ts          TIMESTAMPTZ NOT NULL,
			received_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		log.Fatalf("[DB] failed to create key_presses table: %v", err)
	}


	_, err = db.Exec(`
		CREATE TABLE logs (
			id          SERIAL PRIMARY KEY,
			session_id  TEXT        NOT NULL,
			level       TEXT        NOT NULL,
			message     TEXT        NOT NULL,
			ts          TIMESTAMPTZ NOT NULL,
			received_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		log.Fatalf("[DB] failed to create logs table: %v", err)
	}

	log.Println("[DB] connected and tables ready")
}

func consumeKafka(ctx context.Context) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:           []string{"kafka:9092"},
		Topic:             "telemetry-events",
		GroupID:           "student-storage-service",
		MinBytes:          1,
		MaxBytes:          10e6,
		HeartbeatInterval: 3 * time.Second,
		SessionTimeout:    30 * time.Second,
		RebalanceTimeout:  30 * time.Second,
		StartOffset:       kafka.FirstOffset,
	})
	defer r.Close()

	log.Printf("[KAFKA] consumer started: topic=telemetry-events group=student-storage-service")

	for {
		msg, err := r.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("[KAFKA] read error: %v", err)
			continue
		}

		var event TelemetryEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("[KAFKA] failed to parse message: %v", err)
			continue
		}

		switch event.Event {
		case "cursor_position":
			storeCursor(ctx, event)
		case "key_press":
			storeKeyPress(ctx, event)
		case "log":
			storeLog(ctx, event)
		default:
			log.Printf("[KAFKA] unknown event type: %s", event.Event)
		}
	}
}

func storeCursor(ctx context.Context, event TelemetryEvent) {
	var cursor CursorData
	if err := json.Unmarshal(event.Data, &cursor); err != nil {
		log.Printf("[KAFKA] failed to parse cursor data: %v", err)
		return
	}

	ts, err := time.Parse(time.RFC3339, cursor.Ts)
	if err != nil {
		ts, _ = time.Parse(time.RFC3339, event.Timestamp)
	}

	_, err = db.ExecContext(ctx, `
		INSERT INTO cursor_positions (session_id, x, y, ts)
		VALUES ($1, $2, $3, $4)
	`, event.SessionID, cursor.X, cursor.Y, ts)
	if err != nil {
		log.Printf("[DB] insert cursor error: %v", err)
		return
	}

	log.Printf("[STORE] cursor session=%s x=%.1f y=%.1f ts=%s", event.SessionID, cursor.X, cursor.Y, ts.Format(time.RFC3339))
}

func storeKeyPress(ctx context.Context, event TelemetryEvent) {
	var kp KeyPressData
	if err := json.Unmarshal(event.Data, &kp); err != nil {
		log.Printf("[KAFKA] failed to parse key_press data: %v", err)
		return
	}

	ts, err := time.Parse(time.RFC3339, kp.Ts)
	if err != nil {
		ts, _ = time.Parse(time.RFC3339, event.Timestamp)
	}

	if kp.Modifiers == nil {
		kp.Modifiers = []string{}
	}

	_, err = db.ExecContext(ctx, `
		INSERT INTO key_presses (session_id, key_code, key_name, modifiers, is_combo, ts)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, event.SessionID, kp.KeyCode, kp.KeyName, marshalModifiers(kp.Modifiers), kp.IsCombo, ts)
	if err != nil {
		log.Printf("[DB] insert key_press error: %v", err)
		return
	}

	log.Printf("[STORE] key_press session=%s key=%s(%d) modifiers=%v isCombo=%v ts=%s",
		event.SessionID, kp.KeyName, kp.KeyCode, kp.Modifiers, kp.IsCombo, ts.Format(time.RFC3339))
}

func storeLog(ctx context.Context, event TelemetryEvent) {
	var l LogData
	if err := json.Unmarshal(event.Data, &l); err != nil {
		log.Printf("[KAFKA] failed to parse log data: %v", err)
		return
	}

	ts, err := time.Parse(time.RFC3339, l.Ts)
	if err != nil {
		ts, _ = time.Parse(time.RFC3339, event.Timestamp)
	}

	_, err = db.ExecContext(ctx, `
		INSERT INTO logs (session_id, level, message, ts)
		VALUES ($1, $2, $3, $4)
	`, event.SessionID, l.Level, l.Message, ts)
	if err != nil {
		log.Printf("[DB] insert log error: %v", err)
		return
	}

	log.Printf("[STORE] log session=%s level=%s message=%q ts=%s", event.SessionID, l.Level, l.Message, ts.Format(time.RFC3339))
}

func getAllCursorsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.QueryContext(r.Context(), `
		SELECT id, session_id, x, y, ts, received_at FROM cursor_positions ORDER BY ts DESC
	`)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		log.Printf("[DB] query error: %v", err)
		return
	}
	defer rows.Close()

	result := make([]CursorPosition, 0)
	for rows.Next() {
		var pos CursorPosition
		if err := rows.Scan(&pos.ID, &pos.SessionID, &pos.X, &pos.Y, &pos.Ts, &pos.ReceivedAt); err != nil {
			http.Error(w, "scan error", http.StatusInternalServerError)
			log.Printf("[DB] scan error: %v", err)
			return
		}
		result = append(result, pos)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func getAllKeyPressesHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.QueryContext(r.Context(), `
		SELECT id, session_id, key_code, key_name, modifiers, is_combo, ts, received_at FROM key_presses ORDER BY ts DESC
	`)
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
		if err := rows.Scan(&kp.ID, &kp.SessionID, &kp.KeyCode, &kp.KeyName, &modifiersJSON, &kp.IsCombo, &kp.Ts, &kp.ReceivedAt); err != nil {
			http.Error(w, "scan error", http.StatusInternalServerError)
			log.Printf("[DB] scan error: %v", err)
			return
		}
		kp.Modifiers = unmarshalModifiers(modifiersJSON)
		result = append(result, kp)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func getAllLogsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.QueryContext(r.Context(), `
		SELECT id, session_id, level, message, ts, received_at FROM logs ORDER BY ts DESC
	`)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		log.Printf("[DB] query error: %v", err)
		return
	}
	defer rows.Close()

	result := make([]Log, 0)
	for rows.Next() {
		var l Log
		if err := rows.Scan(&l.ID, &l.SessionID, &l.Level, &l.Message, &l.Ts, &l.ReceivedAt); err != nil {
			http.Error(w, "scan error", http.StatusInternalServerError)
			log.Printf("[DB] scan error: %v", err)
			return
		}
		result = append(result, l)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func main() {
	initDB()

	ctx := context.Background()
	go consumeKafka(ctx)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /cursors", getAllCursorsHandler)
	mux.HandleFunc("GET /keypresses", getAllKeyPressesHandler)
	mux.HandleFunc("GET /logs", getAllLogsHandler)

	addr := ":8080"
	log.Printf("student-storage-service starting on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
