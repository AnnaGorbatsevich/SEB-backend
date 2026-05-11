package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"serviceforreceivingstudentdata/internal/kafka"
)

func TelemetryHandler(producer *kafka.Producer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var event kafka.TelemetryEvent
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
			return
		}

		if event.Event == "" {
			http.Error(w, "Missing event field", http.StatusBadRequest)
			return
		}
		if event.Timestamp == "" {
			http.Error(w, "Missing timestamp field", http.StatusBadRequest)
			return
		}
		if event.SessionID == "" {
			http.Error(w, "Missing sessionId field", http.StatusBadRequest)
			return
		}
		if event.CustomUserID == "" {
			http.Error(w, "Missing customUserId field", http.StatusBadRequest)
			return
		}

		if err := validateData(event); err != nil {
			http.Error(w, fmt.Sprintf("Invalid data: %v", err), http.StatusBadRequest)
			return
		}

		event.ReceivedAt = time.Now()

		log.Printf("[TELEMETRY] event=%s sessionId=%s customUserId=%s", event.Event, event.SessionID, event.CustomUserID)

		if err := producer.Publish(context.Background(), event); err != nil {
			log.Printf("[KAFKA] failed to publish event: %v", err)
			http.Error(w, "Failed to publish event", http.StatusInternalServerError)
			return
		}

		log.Printf("[KAFKA] event published: %s", event.Event)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}
}

func validateData(event kafka.TelemetryEvent) error {
	switch event.Event {
	case "cursor_position":
		var d kafka.CursorData
		if err := json.Unmarshal(event.Data, &d); err != nil {
			return fmt.Errorf("invalid cursor_position data: %w", err)
		}
		if d.Ts == "" {
			return fmt.Errorf("cursor_position: missing ts")
		}
	case "key_press":
		var d kafka.KeyPressData
		if err := json.Unmarshal(event.Data, &d); err != nil {
			return fmt.Errorf("invalid key_press data: %w", err)
		}
		if d.KeyName == "" {
			return fmt.Errorf("key_press: missing keyName")
		}
		if d.Ts == "" {
			return fmt.Errorf("key_press: missing ts")
		}
	case "log":
		var d kafka.LogData
		if err := json.Unmarshal(event.Data, &d); err != nil {
			return fmt.Errorf("invalid log data: %w", err)
		}
		if d.Level == "" {
			return fmt.Errorf("log: missing level")
		}
		if d.Message == "" {
			return fmt.Errorf("log: missing message")
		}
		if d.Ts == "" {
			return fmt.Errorf("log: missing ts")
		}
	case "diagnostic":
		var d kafka.DiagnosticData
		if err := json.Unmarshal(event.Data, &d); err != nil {
			return fmt.Errorf("invalid diagnostic data: %w", err)
		}
		if d.Code == "" {
			return fmt.Errorf("diagnostic: missing code")
		}
		if d.Status == "" {
			return fmt.Errorf("diagnostic: missing status")
		}
	default:
		return fmt.Errorf("unknown event type: %s", event.Event)
	}
	return nil
}
