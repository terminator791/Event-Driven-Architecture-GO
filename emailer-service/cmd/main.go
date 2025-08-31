package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/terminator791/Event-Driven-Architecture-GO/emailer-service/internal/processor"
	"github.com/terminator791/Event-Driven-Architecture-GO/emailer-service/internal/service"
	"github.com/terminator791/Event-Driven-Architecture-GO/emailer-service/pkg/config"
	"github.com/terminator791/Event-Driven-Architecture-GO/shared/pkg/kafka"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize Kafka consumer
	consumer := kafka.NewConsumer(cfg.Kafka.Brokers, cfg.Kafka.Topic, cfg.Kafka.GroupID)
	defer consumer.Close()

	// Initialize email service
	emailService := service.NewEmailService()

	// Initialize event processor
	eventProcessor := processor.NewEventProcessor(consumer, emailService)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("📩 Received shutdown signal")
		cancel()
	}()

	// Start processing events
	log.Printf("🔥 Starting emailer-service...")
	log.Printf("📡 Kafka Brokers: %v", cfg.Kafka.Brokers)
	log.Printf("📝 Topic: %s", cfg.Kafka.Topic)
	log.Printf("👥 Group ID: %s", cfg.Kafka.GroupID)

	if err := eventProcessor.Start(ctx); err != nil && err != context.Canceled {
		log.Fatalf("❌ Event processor failed: %v", err)
	}

	log.Println("👋 Emailer service shutting down gracefully")
}