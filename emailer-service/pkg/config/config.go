package config

import (
	"os"
)

// Config holds application configuration
type Config struct {
	Kafka KafkaConfig
}

// KafkaConfig holds kafka configuration
type KafkaConfig struct {
	Brokers []string
	Topic   string
	GroupID string
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		Kafka: KafkaConfig{
			Brokers: []string{getEnv("KAFKA_BROKERS", "localhost:9092")},
			Topic:   getEnv("KAFKA_TOPIC", "users_topic"),
			GroupID: getEnv("KAFKA_GROUP_ID", "emailer-service"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}