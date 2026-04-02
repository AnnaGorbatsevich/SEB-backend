package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type TelemetryEvent struct {
	Event     string          `json:"event"`
	Timestamp string          `json:"timestamp"`
	Data      json.RawMessage `json:"data"`
	ReceivedAt time.Time      `json:"received_at"`
}

var (
	events []TelemetryEvent
)

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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func main() {
	http.HandleFunc("/telemetry", telemetryHandler)

	addr := ":5000"
	log.Printf("Server starting on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
