package config

import (
	"os"
	"strconv"
)

// Config holds application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Kafka    KafkaConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port string
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// KafkaConfig holds kafka configuration
type KafkaConfig struct {
	Brokers   []string
	OrdersTopic    string
	PaymentsTopic  string
	GroupID   string
}

// Load loads configuration from environment variables
func Load() *Config {
	dbPort, _ := strconv.Atoi(getEnv("DB_PORT", "5432"))

	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8082"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     dbPort,
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "paymentdb"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Kafka: KafkaConfig{
			Brokers:       []string{getEnv("KAFKA_BROKERS", "localhost:9092")},
			OrdersTopic:   getEnv("KAFKA_ORDERS_TOPIC", "orders_topic"),
			PaymentsTopic: getEnv("KAFKA_PAYMENTS_TOPIC", "payments_topic"),
			GroupID:       getEnv("KAFKA_GROUP_ID", "payment-service"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}