package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	DatabaseURL     string        `json:"database_url"`
	KafkaBroker     string        `json:"kafka_broker"`
	KafkaTopic      string        `json:"kafka_topic"`
	KafkaGroupID    string        `json:"kafka_group_id"`
	HTTPAddress     string        `json:"http_address"`
	MaxRetryAttempts int           `json:"max_retry_attempts"`
	RetryInterval   time.Duration `json:"retry_interval"`
}

func Load() *Config {
	return &Config{
		DatabaseURL:     getEnv("DATABASE_URL", "postgres://postgres:@localhost:5432/studentdata?sslmode=disable"),
		KafkaBroker:     getEnv("KAFKA_BROKER", "localhost:9093"),
		KafkaTopic:      getEnv("KAFKA_TOPIC", "telemetry-events"),
		KafkaGroupID:    getEnv("KAFKA_GROUP_ID", "student-storage-service"),
		HTTPAddress:     getEnv("HTTP_ADDRESS", ":8080"),
		MaxRetryAttempts: getIntEnv("MAX_RETRY_ATTEMPTS", 10),
		RetryInterval:   time.Duration(getIntEnv("RETRY_INTERVAL_SECONDS", 3)) * time.Second,
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getIntEnv(key string, fallback int) int {
	if value, exists := os.LookupEnv(key); exists {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return fallback
}