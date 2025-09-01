package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/terminator791/Event-Driven-Architecture-GO/order-service/internal/handler"
	"github.com/terminator791/Event-Driven-Architecture-GO/order-service/internal/repository"
	"github.com/terminator791/Event-Driven-Architecture-GO/order-service/internal/service"
	"github.com/terminator791/Event-Driven-Architecture-GO/order-service/pkg/config"
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

	// Create order tables
	if err := database.CreateOrderTables(db); err != nil {
		log.Fatalf("Failed to create order tables: %v", err)
	}

	// Initialize Kafka producer
	producer := kafka.NewProducer(cfg.Kafka.Brokers, cfg.Kafka.Topic)
	defer producer.Close()

	// Create topic if it doesn't exist
	if err := kafka.CreateTopic(cfg.Kafka.Brokers, cfg.Kafka.Topic, 1, 1); err != nil {
		log.Printf("Warning: Failed to create Kafka topic: %v", err)
		// Continue anyway - topic might already exist
	}

	// Initialize layers
	orderRepo := repository.NewOrderRepository(db)
	productService := service.NewMockProductService()
	orderService := service.NewOrderService(orderRepo, productService, producer)
	orderHandler := handler.NewOrderHandler(orderService)

	// Setup Gin router
	router := gin.Default()

	// Add middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Routes
	api := router.Group("/api/v1")
	{
		// Order management
		api.POST("/orders", orderHandler.CreateOrder)
		api.GET("/orders/:id", orderHandler.GetOrder)
		api.PUT("/orders/:id/status", orderHandler.UpdateOrderStatus)
		
		// User orders
		api.GET("/users/:userId/orders", orderHandler.GetUserOrders)
		api.POST("/users/:userId/orders/:id/cancel", orderHandler.CancelOrder)
		
		// Health check
		api.GET("/health", orderHandler.HealthCheck)
	}

	// Start server
	log.Printf("Starting order-service server on port %s", cfg.Server.Port)
	if err := router.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}