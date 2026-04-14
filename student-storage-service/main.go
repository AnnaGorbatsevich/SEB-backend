package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	kafka "github.com/segmentio/kafka-go"
)

type TelemetryEvent struct {
	Event     string          `json:"event"`
	Timestamp string          `json:"timestamp"`
	Data      json.RawMessage `json:"data"`
}

type CursorData struct {
	StudentID string  `json:"student_id"`
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
}

type CursorPosition struct {
	StudentID string    `json:"student_id"`
	X         float64   `json:"x"`
	Y         float64   `json:"y"`
	Timestamp time.Time `json:"timestamp"`
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
		CREATE TABLE IF NOT EXISTS cursor_positions (
			student_id TEXT PRIMARY KEY,
			x          DOUBLE PRECISION NOT NULL,
			y          DOUBLE PRECISION NOT NULL,
			timestamp  TIMESTAMPTZ      NOT NULL
		)
	`)
	if err != nil {
		log.Fatalf("[DB] failed to create table: %v", err)
	}

	log.Println("[DB] connected and table ready")
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

		if event.Event != "cursor_position" {
			continue
		}

		var cursor CursorData
		if err := json.Unmarshal(event.Data, &cursor); err != nil {
			log.Printf("[KAFKA] failed to parse cursor data: %v", err)
			continue
		}

		ts, _ := time.Parse(time.RFC3339, event.Timestamp)

		_, err = db.ExecContext(ctx, `
			INSERT INTO cursor_positions (student_id, x, y, timestamp)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (student_id) DO UPDATE SET x = $2, y = $3, timestamp = $4
		`, cursor.StudentID, cursor.X, cursor.Y, ts)
		if err != nil {
			log.Printf("[DB] upsert error: %v", err)
			continue
		}

		log.Printf("[STORE] updated student=%s x=%.1f y=%.1f", cursor.StudentID, cursor.X, cursor.Y)
	}
}

func getAllCursorsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.QueryContext(r.Context(), `SELECT student_id, x, y, timestamp FROM cursor_positions`)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		log.Printf("[DB] query error: %v", err)
		return
	}
	defer rows.Close()

	result := make([]CursorPosition, 0)
	for rows.Next() {
		var pos CursorPosition
		if err := rows.Scan(&pos.StudentID, &pos.X, &pos.Y, &pos.Timestamp); err != nil {
			http.Error(w, "scan error", http.StatusInternalServerError)
			log.Printf("[DB] scan error: %v", err)
			return
		}
		result = append(result, pos)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func main() {
	initDB()

	ctx := context.Background()
	go consumeKafka(ctx)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /students/cursors", getAllCursorsHandler)

	addr := ":8080"
	log.Printf("student-storage-service starting on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
