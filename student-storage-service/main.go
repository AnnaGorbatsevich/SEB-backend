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
	// Загружаем конфигурацию
	cfg := config.Load()

	// Инициализируем контекст
	ctx := context.Background()

	// Подключаемся к базе данных
	dbService, err := db.New(ctx, cfg.DatabaseURL, cfg.MaxRetryAttempts, cfg.RetryInterval)
	if err != nil {
		log.Fatalf("[DB] failed to connect: %v", err)
	}
	defer dbService.Close()

	log.Println("[DB] connected and ready")

	// Запускаем Kafka consumer в отдельной горутине
	go kafka.Consume(ctx, cfg.KafkaBroker, cfg.KafkaTopic, cfg.KafkaGroupID, dbService)

	// Настраиваем HTTP маршрутизацию
	mux := http.NewServeMux()

	mux.HandleFunc("GET /cursors", handlers.GetAllCursorsHandler(dbService))
	mux.HandleFunc("GET /keypresses", handlers.GetAllKeyPressesHandler(dbService))
	mux.HandleFunc("GET /logs", handlers.GetAllLogsHandler(dbService))
	mux.HandleFunc("GET /session-events", handlers.GetSessionEventsHandler(dbService))
	mux.HandleFunc("GET /diagnostics", handlers.GetDiagnosticsHandler(dbService))
	mux.HandleFunc("GET /diagnostic-summary", handlers.GetDiagnosticSummaryHandler(dbService))
	mux.HandleFunc("GET /health", handlers.HealthHandler(dbService))

	cors := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}

	// Запускаем HTTP сервер
	log.Printf("student-storage-service starting on http://localhost%s", cfg.HTTPAddress)
	if err := http.ListenAndServe(cfg.HTTPAddress, cors(mux)); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
