package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	kafka "github.com/segmentio/kafka-go"
)

type TelemetryEvent struct {
	Event      string          `json:"event"`
	Timestamp  string          `json:"timestamp"`
	Data       json.RawMessage `json:"data"`
	ReceivedAt time.Time       `json:"received_at"`
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

var (
	store   = make(map[string]CursorPosition)
	storeMu sync.RWMutex
)

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

		storeMu.Lock()
		store[cursor.StudentID] = CursorPosition{
			StudentID: cursor.StudentID,
			X:         cursor.X,
			Y:         cursor.Y,
			Timestamp: ts,
		}
		storeMu.Unlock()

		log.Printf("[STORE] updated student=%s x=%.1f y=%.1f", cursor.StudentID, cursor.X, cursor.Y)
	}
}

func getAllCursorsHandler(w http.ResponseWriter, r *http.Request) {
	storeMu.RLock()
	result := make([]CursorPosition, 0, len(store))
	for _, pos := range store {
		result = append(result, pos)
	}
	storeMu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func main() {
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
