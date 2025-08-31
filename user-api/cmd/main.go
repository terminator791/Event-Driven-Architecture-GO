package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/terminator791/Event-Driven-Architecture-GO/shared/pkg/database"
	"github.com/terminator791/Event-Driven-Architecture-GO/shared/pkg/kafka"
	"github.com/terminator791/Event-Driven-Architecture-GO/user-api/internal/handler"
	"github.com/terminator791/Event-Driven-Architecture-GO/user-api/internal/repository"
	"github.com/terminator791/Event-Driven-Architecture-GO/user-api/internal/service"
	"github.com/terminator791/Event-Driven-Architecture-GO/user-api/pkg/config"
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

	// Create tables
	if err := database.CreateUserTable(db); err != nil {
		log.Fatalf("Failed to create user table: %v", err)
	}

	// Initialize Kafka producer
	producer := kafka.NewProducer(cfg.Kafka.Brokers, cfg.Kafka.Topic)
	defer producer.Close()

	// Create topic if it doesn't exist
	if err := kafka.CreateTopic(cfg.Kafka.Brokers, cfg.Kafka.Topic, 1, 1); err != nil {
		log.Printf("Warning: Failed to create Kafka topic: %v", err)
	}

	// Initialize layers
	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo, producer)
	userHandler := handler.NewUserHandler(userService)

	// Setup Gin router
	router := gin.Default()

	// Add middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Routes
	api := router.Group("/api/v1")
	{
		api.POST("/users", userHandler.CreateUser)
		api.GET("/health", userHandler.HealthCheck)
	}

	// Start server
	log.Printf("Starting user-api server on port %s", cfg.Server.Port)
	if err := router.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}