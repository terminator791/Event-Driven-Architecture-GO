package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/terminator791/Event-Driven-Architecture-GO/payment-service/internal/handler"
	"github.com/terminator791/Event-Driven-Architecture-GO/payment-service/internal/processor"
	"github.com/terminator791/Event-Driven-Architecture-GO/payment-service/internal/repository"
	"github.com/terminator791/Event-Driven-Architecture-GO/payment-service/internal/service"
	"github.com/terminator791/Event-Driven-Architecture-GO/payment-service/pkg/config"
	"github.com/terminator791/Event-Driven-Architecture-GO/shared/pkg/database"
	"github.com/terminator791/Event-Driven-Architecture-GO/shared/pkg/kafka"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database
	dbConfig := database.Config{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		DBName:   cfg.Database.DBName,
		SSLMode:  cfg.Database.SSLMode,
	}

	db, err := database.NewConnection(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Create payment tables
	if err := database.CreatePaymentTables(db); err != nil {
		log.Fatalf("Failed to create payment tables: %v", err)
	}

	// Initialize Kafka producer
	producer := kafka.NewProducer(cfg.Kafka.Brokers, cfg.Kafka.PaymentsTopic)
	defer producer.Close()

	// Initialize Kafka consumer for order events
	consumer := kafka.NewConsumer(cfg.Kafka.Brokers, cfg.Kafka.OrdersTopic, cfg.Kafka.GroupID)
	defer consumer.Close()

	// Create topics if they don't exist
	if err := kafka.CreateTopic(cfg.Kafka.Brokers, cfg.Kafka.PaymentsTopic, 1, 1); err != nil {
		log.Printf("Warning: Failed to create payments Kafka topic: %v", err)
	}
	if err := kafka.CreateTopic(cfg.Kafka.Brokers, cfg.Kafka.OrdersTopic, 1, 1); err != nil {
		log.Printf("Warning: Failed to create orders Kafka topic: %v", err)
	}

	// Initialize layers
	paymentRepo := repository.NewPaymentRepository(db)
	gateway := service.NewMockPaymentGateway()
	paymentService := service.NewPaymentService(paymentRepo, gateway, producer)
	paymentHandler := handler.NewPaymentHandler(paymentService)

	// Initialize event processor
	eventProcessor := processor.NewEventProcessor(consumer, paymentService)

	// Setup context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Wait group for goroutines
	var wg sync.WaitGroup

	// Start event processor in a separate goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := eventProcessor.Start(ctx); err != nil && err != context.Canceled {
			log.Printf("Event processor error: %v", err)
		}
	}()

	// Start payment retry scheduler in a separate goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(5 * time.Minute) // Retry every 5 minutes
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				log.Println("🔄 Running payment retry scheduler...")
				if err := paymentService.RetryPendingPayments(ctx); err != nil {
					log.Printf("Payment retry error: %v", err)
				}
			}
		}
	}()

	// Setup Gin router
	router := gin.Default()

	// Add middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Routes
	api := router.Group("/api/v1")
	{
		// Payment management
		api.POST("/payments", paymentHandler.ProcessPayment)
		api.GET("/payments/:id", paymentHandler.GetPayment)
		api.GET("/orders/:orderId/payment", paymentHandler.GetPaymentByOrderID)
		api.POST("/payments/:id/refund", paymentHandler.RefundPayment)
		api.POST("/payments/retry", paymentHandler.RetryPendingPayments)
		
		// Health check
		api.GET("/health", paymentHandler.HealthCheck)
	}

	// Start HTTP server in a separate goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Printf("Starting payment-service server on port %s", cfg.Server.Port)
		if err := router.Run(":" + cfg.Server.Port); err != nil {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("🛑 Shutting down payment service...")

	// Cancel context to stop goroutines
	cancel()

	// Wait for all goroutines to finish with timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("✅ Payment service shutdown complete")
	case <-time.After(30 * time.Second):
		log.Println("⚠️  Shutdown timeout, forcing exit")
	}
}