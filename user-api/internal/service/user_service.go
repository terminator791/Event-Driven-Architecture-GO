package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/terminator791/Event-Driven-Architecture-GO/shared/pkg/events"
	"github.com/terminator791/Event-Driven-Architecture-GO/shared/pkg/models"
	"github.com/terminator791/Event-Driven-Architecture-GO/user-api/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

// ProducerInterface defines the interface for publishing messages to Kafka
type ProducerInterface interface {
	Publish(ctx context.Context, key, value []byte) error
	Close() error
}

// UserService defines the interface for user business logic
type UserService interface {
	CreateUser(ctx context.Context, req models.CreateUserRequest) (*models.CreateUserResponse, error)
}

// userService implements UserService
type userService struct {
	userRepo repository.UserRepository
	producer ProducerInterface
}

// NewUserService creates a new user service
func NewUserService(userRepo repository.UserRepository, producer ProducerInterface) UserService {
	return &userService{
		userRepo: userRepo,
		producer: producer,
	}
}

// CreateUser creates a new user with idempotency check
func (s *userService) CreateUser(ctx context.Context, req models.CreateUserRequest) (*models.CreateUserResponse, error) {
	// Validate input
	if err := s.validateCreateUserRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check if user already exists (idempotency)
	exists, err := s.userRepo.Exists(req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}

	if exists {
		return nil, fmt.Errorf("user with email %s already exists", req.Email)
	}

	// Hash password
	hashedPassword, err := s.hashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &models.User{
		ID:           uuid.New().String(),
		Email:        req.Email,
		PasswordHash: hashedPassword,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Save to database
	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Publish event
	if err := s.publishUserCreatedEvent(ctx, user); err != nil {
		// Log error but don't fail the operation
		// In a production system, you might want to implement retry logic
		fmt.Printf("Warning: failed to publish user created event: %v\n", err)
	}

	return &models.CreateUserResponse{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}, nil
}

// validateCreateUserRequest validates the create user request
func (s *userService) validateCreateUserRequest(req models.CreateUserRequest) error {
	if req.Email == "" {
		return fmt.Errorf("email is required")
	}

	if req.Password == "" {
		return fmt.Errorf("password is required")
	}

	if len(req.Password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	return nil
}

// hashPassword hashes the password using bcrypt
func (s *userService) hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// publishUserCreatedEvent publishes a user created event to Kafka
func (s *userService) publishUserCreatedEvent(ctx context.Context, user *models.User) error {
	event := events.UserCreated{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		EventID:   uuid.New().String(),
	}

	eventData, err := event.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize event: %w", err)
	}

	return s.producer.Publish(ctx, []byte(user.ID), eventData)
}