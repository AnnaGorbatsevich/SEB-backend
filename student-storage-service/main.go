package main

import (
	"context"
	"log"
	"net/http"

	"studentstorage/internal/config"
	"studentstorage/internal/db"
	"studentstorage/internal/handlers"
	"studentstorage/internal/kafka"
)

func main() {
	cfg := config.Load()

	ctx := context.Background()

	dbService, err := db.New(ctx, cfg.DatabaseURL, cfg.MaxRetryAttempts, cfg.RetryInterval)
	if err != nil {
		log.Fatalf("[DB] failed to connect: %v", err)
	}
	defer dbService.Close()

	log.Println("[DB] connected and ready")

	go kafka.Consume(ctx, cfg.KafkaBroker, cfg.KafkaTopic, cfg.KafkaGroupID, dbService)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /cursors", handlers.GetAllCursorsHandler(dbService))
	mux.HandleFunc("GET /keypresses", handlers.GetAllKeyPressesHandler(dbService))
	mux.HandleFunc("GET /logs", handlers.GetAllLogsHandler(dbService))
	mux.HandleFunc("GET /session-events", handlers.GetSessionEventsHandler(dbService))
	mux.HandleFunc("GET /diagnostics", handlers.GetDiagnosticsHandler(dbService))
	mux.HandleFunc("GET /health", handlers.HealthHandler(dbService))

	log.Printf("student-storage-service starting on http://localhost%s", cfg.HTTPAddress)
	if err := http.ListenAndServe(cfg.HTTPAddress, mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
