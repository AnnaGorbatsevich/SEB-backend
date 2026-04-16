package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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

var kafkaWriter *kafka.Writer

func initKafka() {
	topic := "telemetry-events"

	kafkaWriter = &kafka.Writer{
		Addr:         kafka.TCP("kafka:9092"),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}
	log.Printf("[KAFKA] producer initialised: brokers=%s topic=%s", "kafka:9092", topic)
}

func validateData(event TelemetryEvent) error {
	switch event.Event {
	case "cursor_position":
		var d CursorData
		if err := json.Unmarshal(event.Data, &d); err != nil {
			return fmt.Errorf("invalid cursor_position data: %w", err)
		}
		if d.Ts == "" {
			return fmt.Errorf("cursor_position: missing ts")
		}
	case "key_press":
		var d KeyPressData
		if err := json.Unmarshal(event.Data, &d); err != nil {
			return fmt.Errorf("invalid key_press data: %w", err)
		}
		if d.KeyName == "" {
			return fmt.Errorf("key_press: missing keyName")
		}
		if d.Ts == "" {
			return fmt.Errorf("key_press: missing ts")
		}
	default:
		return fmt.Errorf("unknown event type: %s", event.Event)
	}
	return nil
}

func telemetryHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		receiveTelemetry(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func receiveTelemetry(w http.ResponseWriter, r *http.Request) {
	var event TelemetryEvent

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

	if err := validateData(event); err != nil {
		http.Error(w, fmt.Sprintf("Invalid data: %v", err), http.StatusBadRequest)
		return
	}

	event.ReceivedAt = time.Now()

	log.Printf("[TELEMETRY] event=%s timestamp=%s data=%s", event.Event, event.Timestamp, event.Data)

	payload, err := json.Marshal(event)
	if err != nil {
		log.Printf("[KAFKA] failed to marshal event: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	if err := kafkaWriter.WriteMessages(context.Background(), kafka.Message{
		Key:   []byte(event.Event),
		Value: payload,
	}); err != nil {
		log.Printf("[KAFKA] failed to publish event: %v", err)
		http.Error(w, "Failed to publish event", http.StatusInternalServerError)
		return
	}

	log.Printf("[KAFKA] event published: %s", event.Event)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func main() {
	initKafka()
	defer kafkaWriter.Close()

	http.HandleFunc("/telemetry", telemetryHandler)

	addr := ":5000"
	log.Printf("Server starting on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
