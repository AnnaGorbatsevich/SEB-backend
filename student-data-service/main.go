package main

import (
	"log"
	"net/http"

	"serviceforreceivingstudentdata/internal/config"
	"serviceforreceivingstudentdata/internal/handlers"
	"serviceforreceivingstudentdata/internal/kafka"
)

func main() {
	cfg := config.Load()

	producer := kafka.NewProducer(cfg.KafkaBroker, cfg.KafkaTopic)
	defer producer.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("POST /telemetry", handlers.TelemetryHandler(producer))

	log.Printf("student-data-service starting on http://localhost%s", cfg.HTTPAddress)
	if err := http.ListenAndServe(cfg.HTTPAddress, mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
