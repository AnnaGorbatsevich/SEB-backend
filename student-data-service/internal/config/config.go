package config

import "os"

type Config struct {
	KafkaBroker string
	KafkaTopic  string
	HTTPAddress string
}

func Load() *Config {
	return &Config{
		KafkaBroker: getEnv("KAFKA_BROKER", "kafka:9092"),
		KafkaTopic:  getEnv("KAFKA_TOPIC", "telemetry-events"),
		HTTPAddress: getEnv("HTTP_ADDRESS", ":5000"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
