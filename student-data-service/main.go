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

	event.ReceivedAt = time.Now()

	log.Printf("[TELEMETRY] event=%s timestamp=%s data=%s", event.Event, event.Timestamp, event.Data)

	payload, err := json.Marshal(event)
	if err != nil {
		log.Printf("[KAFKA] failed to marshal event: %v", err)
	} else {
		if err := kafkaWriter.WriteMessages(context.Background(), kafka.Message{
			Key:   []byte(event.Event),
			Value: payload,
		}); err != nil {
			log.Printf("[KAFKA] failed to publish event: %v", err)
		} else {
			log.Printf("[KAFKA] event published: %s", event.Event)
		}
	}

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
