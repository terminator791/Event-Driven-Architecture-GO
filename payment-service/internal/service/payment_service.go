package service

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/terminator791/Event-Driven-Architecture-GO/payment-service/internal/repository"
	"github.com/terminator791/Event-Driven-Architecture-GO/shared/pkg/events"
	"github.com/terminator791/Event-Driven-Architecture-GO/shared/pkg/models"
)

// ProducerInterface defines the interface for publishing messages to Kafka
type ProducerInterface interface {
	Publish(ctx context.Context, key, value []byte) error
	Close() error
}

// PaymentGateway defines the interface for external payment processing
type PaymentGateway interface {
	ProcessPayment(ctx context.Context, payment *models.Payment) (*PaymentResult, error)
	RefundPayment(ctx context.Context, transactionID string, amount float64) (*RefundResult, error)
}

// PaymentResult represents the result of a payment processing attempt
type PaymentResult struct {
	Success       bool
	TransactionID string
	FailureReason string
	Gateway       string
}

// RefundResult represents the result of a refund attempt
type RefundResult struct {
	Success       bool
	RefundID      string
	FailureReason string
}

// PaymentService defines the interface for payment business logic
type PaymentService interface {
	ProcessPayment(ctx context.Context, req models.ProcessPaymentRequest) (*models.ProcessPaymentResponse, error)
	GetPayment(ctx context.Context, id string) (*models.Payment, error)
	GetPaymentByOrderID(ctx context.Context, orderID string) (*models.Payment, error)
	RefundPayment(ctx context.Context, paymentID string, amount float64, reason string) error
	RetryPendingPayments(ctx context.Context) error
}

// paymentService implements PaymentService
type paymentService struct {
	paymentRepo repository.PaymentRepository
	gateway     PaymentGateway
	producer    ProducerInterface
	maxRetries  int
}

// NewPaymentService creates a new payment service
func NewPaymentService(paymentRepo repository.PaymentRepository, gateway PaymentGateway, producer ProducerInterface) PaymentService {
	return &paymentService{
		paymentRepo: paymentRepo,
		gateway:     gateway,
		producer:    producer,
		maxRetries:  3,
	}
}

// ProcessPayment processes a payment with retry logic and state management
func (s *paymentService) ProcessPayment(ctx context.Context, req models.ProcessPaymentRequest) (*models.ProcessPaymentResponse, error) {
	// Validate request
	if err := s.validateProcessPaymentRequest(req); err != nil {
		return nil, err
	}

	// Check if payment already exists for this order
	existingPayment, err := s.paymentRepo.GetByOrderID(ctx, req.OrderID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing payment: %w", err)
	}

	if existingPayment != nil {
		switch existingPayment.Status {
		case models.PaymentStatusCompleted:
			return &models.ProcessPaymentResponse{
				ID:            existingPayment.ID,
				OrderID:       existingPayment.OrderID,
				Status:        existingPayment.Status,
				TransactionID: existingPayment.TransactionID,
				ProcessedAt:   existingPayment.ProcessedAt,
			}, nil
		case models.PaymentStatusPending:
			return nil, fmt.Errorf("payment is already being processed for order %s", req.OrderID)
		case models.PaymentStatusFailed:
			// Allow retry if within retry limit
			if existingPayment.ID != "" {
				return s.retryPayment(ctx, existingPayment)
			}
		}
	}

	// Create new payment
	now := time.Now().UTC()
	payment := &models.Payment{
		ID:            uuid.New().String(),
		OrderID:       req.OrderID,
		UserID:        "", // Will be set when we integrate with order service
		Amount:        req.Amount,
		Currency:      req.Currency,
		PaymentMethod: req.PaymentMethod,
		Status:        models.PaymentStatusPending,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	// Save payment to database
	if err := s.paymentRepo.Create(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	// Process payment through gateway
	result, err := s.gateway.ProcessPayment(ctx, payment)
	if err != nil {
		// Mark payment as failed
		s.paymentRepo.UpdateStatus(ctx, payment.ID, models.PaymentStatusFailed, "", err.Error())
		
		// Publish payment failed event
		s.publishPaymentFailedEvent(ctx, payment, 0)
		
		return nil, fmt.Errorf("payment processing failed: %w", err)
	}

	if result.Success {
		// Update payment as completed
		err = s.paymentRepo.UpdateStatus(ctx, payment.ID, models.PaymentStatusCompleted, result.TransactionID, "")
		if err != nil {
			return nil, fmt.Errorf("failed to update payment status: %w", err)
		}

		payment.Status = models.PaymentStatusCompleted
		payment.TransactionID = result.TransactionID
		processedAt := time.Now().UTC()
		payment.ProcessedAt = &processedAt

		// Publish payment processed event
		if err := s.publishPaymentProcessedEvent(ctx, payment); err != nil {
			fmt.Printf("Warning: failed to publish payment processed event: %v\n", err)
		}

		return &models.ProcessPaymentResponse{
			ID:            payment.ID,
			OrderID:       payment.OrderID,
			Status:        payment.Status,
			TransactionID: payment.TransactionID,
			ProcessedAt:   payment.ProcessedAt,
		}, nil
	} else {
		// Mark payment as failed
		err = s.paymentRepo.UpdateStatus(ctx, payment.ID, models.PaymentStatusFailed, "", result.FailureReason)
		if err != nil {
			return nil, fmt.Errorf("failed to update payment status: %w", err)
		}

		payment.Status = models.PaymentStatusFailed
		payment.FailureReason = result.FailureReason

		// Publish payment failed event
		if err := s.publishPaymentFailedEvent(ctx, payment, 0); err != nil {
			fmt.Printf("Warning: failed to publish payment failed event: %v\n", err)
		}

		return nil, fmt.Errorf("payment failed: %s", result.FailureReason)
	}
}

// GetPayment retrieves a payment by ID
func (s *paymentService) GetPayment(ctx context.Context, id string) (*models.Payment, error) {
	payment, err := s.paymentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}

	if payment == nil {
		return nil, fmt.Errorf("payment not found")
	}

	return payment, nil
}

// GetPaymentByOrderID retrieves a payment by order ID
func (s *paymentService) GetPaymentByOrderID(ctx context.Context, orderID string) (*models.Payment, error) {
	payment, err := s.paymentRepo.GetByOrderID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment by order ID: %w", err)
	}

	if payment == nil {
		return nil, fmt.Errorf("payment not found for order")
	}

	return payment, nil
}

// RefundPayment processes a refund for a payment
func (s *paymentService) RefundPayment(ctx context.Context, paymentID string, amount float64, reason string) error {
	// Get payment
	payment, err := s.paymentRepo.GetByID(ctx, paymentID)
	if err != nil {
		return fmt.Errorf("failed to get payment: %w", err)
	}

	if payment == nil {
		return fmt.Errorf("payment not found")
	}

	if payment.Status != models.PaymentStatusCompleted {
		return fmt.Errorf("can only refund completed payments")
	}

	if amount > payment.Amount {
		return fmt.Errorf("refund amount cannot exceed original payment amount")
	}

	// Process refund through gateway
	result, err := s.gateway.RefundPayment(ctx, payment.TransactionID, amount)
	if err != nil {
		return fmt.Errorf("refund processing failed: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("refund failed: %s", result.FailureReason)
	}

	// Update payment status
	err = s.paymentRepo.UpdateStatus(ctx, paymentID, models.PaymentStatusRefunded, payment.TransactionID, "")
	if err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	// Publish payment refunded event
	if err := s.publishPaymentRefundedEvent(ctx, payment, amount, reason); err != nil {
		fmt.Printf("Warning: failed to publish payment refunded event: %v\n", err)
	}

	return nil
}

// RetryPendingPayments retries payments that are in pending or failed state
func (s *paymentService) RetryPendingPayments(ctx context.Context) error {
	pendingPayments, err := s.paymentRepo.GetPendingPayments(ctx, 100)
	if err != nil {
		return fmt.Errorf("failed to get pending payments: %w", err)
	}

	for _, payment := range pendingPayments {
		// Skip if exceeded max retries
		if payment.Status == models.PaymentStatusFailed && s.shouldRetryPayment(payment) {
			_, err := s.retryPayment(ctx, payment)
			if err != nil {
				fmt.Printf("Failed to retry payment %s: %v\n", payment.ID, err)
			}
		}
	}

	return nil
}

// retryPayment retries a failed payment
func (s *paymentService) retryPayment(ctx context.Context, payment *models.Payment) (*models.ProcessPaymentResponse, error) {
	// Increment retry count
	if err := s.paymentRepo.IncrementRetryCount(ctx, payment.ID); err != nil {
		return nil, fmt.Errorf("failed to increment retry count: %w", err)
	}

	// Update status to pending for retry
	if err := s.paymentRepo.UpdateStatus(ctx, payment.ID, models.PaymentStatusPending, "", ""); err != nil {
		return nil, fmt.Errorf("failed to update payment status for retry: %w", err)
	}

	// Process payment through gateway
	result, err := s.gateway.ProcessPayment(ctx, payment)
	if err != nil {
		// Mark payment as failed again
		s.paymentRepo.UpdateStatus(ctx, payment.ID, models.PaymentStatusFailed, "", err.Error())
		
		// Publish payment failed event with retry count
		s.publishPaymentFailedEvent(ctx, payment, payment.Status == models.PaymentStatusFailed)
		
		return nil, fmt.Errorf("payment retry failed: %w", err)
	}

	if result.Success {
		// Update payment as completed
		err = s.paymentRepo.UpdateStatus(ctx, payment.ID, models.PaymentStatusCompleted, result.TransactionID, "")
		if err != nil {
			return nil, fmt.Errorf("failed to update payment status: %w", err)
		}

		payment.Status = models.PaymentStatusCompleted
		payment.TransactionID = result.TransactionID
		processedAt := time.Now().UTC()
		payment.ProcessedAt = &processedAt

		// Publish payment processed event
		if err := s.publishPaymentProcessedEvent(ctx, payment); err != nil {
			fmt.Printf("Warning: failed to publish payment processed event: %v\n", err)
		}

		return &models.ProcessPaymentResponse{
			ID:            payment.ID,
			OrderID:       payment.OrderID,
			Status:        payment.Status,
			TransactionID: payment.TransactionID,
			ProcessedAt:   payment.ProcessedAt,
		}, nil
	} else {
		// Mark payment as failed
		err = s.paymentRepo.UpdateStatus(ctx, payment.ID, models.PaymentStatusFailed, "", result.FailureReason)
		if err != nil {
			return nil, fmt.Errorf("failed to update payment status: %w", err)
		}

		// Publish payment failed event
		if err := s.publishPaymentFailedEvent(ctx, payment, payment.Status == models.PaymentStatusFailed); err != nil {
			fmt.Printf("Warning: failed to publish payment failed event: %v\n", err)
		}

		return nil, fmt.Errorf("payment retry failed: %s", result.FailureReason)
	}
}

// shouldRetryPayment determines if a payment should be retried
func (s *paymentService) shouldRetryPayment(payment *models.Payment) bool {
	// Don't retry if max retries exceeded
	maxRetries := s.maxRetries
	if payment.RetryCount >= maxRetries {
		return false
	}

	// Don't retry if too old (e.g., older than 24 hours)
	if time.Since(payment.CreatedAt) > 24*time.Hour {
		return false
	}

	return true
}

// validateProcessPaymentRequest validates the process payment request
func (s *paymentService) validateProcessPaymentRequest(req models.ProcessPaymentRequest) error {
	if req.OrderID == "" {
		return fmt.Errorf("order ID is required")
	}

	if req.Amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}

	if req.Currency == "" {
		return fmt.Errorf("currency is required")
	}

	if req.PaymentMethod == "" {
		return fmt.Errorf("payment method is required")
	}

	return nil
}

// publishPaymentProcessedEvent publishes a payment processed event
func (s *paymentService) publishPaymentProcessedEvent(ctx context.Context, payment *models.Payment) error {
	event := events.NewPaymentProcessed(payment)

	// Set correlation ID from context if available
	if correlationID := ctx.Value("correlation_id"); correlationID != nil {
		if strCorrelationID, ok := correlationID.(string); ok {
			event.Metadata.WithCorrelationID(strCorrelationID)
		}
	}

	// Set user ID in metadata
	event.Metadata.WithUserID(payment.UserID)

	eventData, err := event.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize event: %w", err)
	}

	return s.producer.Publish(ctx, []byte(payment.ID), eventData)
}

// publishPaymentFailedEvent publishes a payment failed event
func (s *paymentService) publishPaymentFailedEvent(ctx context.Context, payment *models.Payment, retryCount interface{}) error {
	var retryCountInt int
	if rc, ok := retryCount.(int); ok {
		retryCountInt = rc
	}

	event := events.NewPaymentFailed(payment, retryCountInt)

	// Set correlation ID from context if available
	if correlationID := ctx.Value("correlation_id"); correlationID != nil {
		if strCorrelationID, ok := correlationID.(string); ok {
			event.Metadata.WithCorrelationID(strCorrelationID)
		}
	}

	// Set user ID in metadata
	event.Metadata.WithUserID(payment.UserID)

	eventData, err := event.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize event: %w", err)
	}

	return s.producer.Publish(ctx, []byte(payment.ID), eventData)
}

// publishPaymentRefundedEvent publishes a payment refunded event
func (s *paymentService) publishPaymentRefundedEvent(ctx context.Context, payment *models.Payment, refundAmount float64, reason string) error {
	event := events.NewPaymentRefunded(payment, refundAmount, reason)

	// Set correlation ID from context if available
	if correlationID := ctx.Value("correlation_id"); correlationID != nil {
		if strCorrelationID, ok := correlationID.(string); ok {
			event.Metadata.WithCorrelationID(strCorrelationID)
		}
	}

	// Set user ID in metadata
	event.Metadata.WithUserID(payment.UserID)

	eventData, err := event.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize event: %w", err)
	}

	return s.producer.Publish(ctx, []byte(payment.ID), eventData)
}

// MockPaymentGateway is a mock implementation for testing/demo purposes
type MockPaymentGateway struct{}

// NewMockPaymentGateway creates a new mock payment gateway
func NewMockPaymentGateway() PaymentGateway {
	return &MockPaymentGateway{}
}

// ProcessPayment simulates payment processing with random success/failure
func (g *MockPaymentGateway) ProcessPayment(ctx context.Context, payment *models.Payment) (*PaymentResult, error) {
	// Simulate processing time
	time.Sleep(100 * time.Millisecond)

	// Simulate 90% success rate
	if rand.Float64() < 0.9 {
		return &PaymentResult{
			Success:       true,
			TransactionID: fmt.Sprintf("txn_%s_%d", payment.ID[:8], time.Now().Unix()),
			Gateway:       "mock-gateway",
		}, nil
	}

	// Simulate various failure reasons
	failureReasons := []string{
		"insufficient funds",
		"card declined",
		"expired card",
		"invalid card details",
		"gateway timeout",
	}

	reason := failureReasons[rand.Intn(len(failureReasons))]
	return &PaymentResult{
		Success:       false,
		FailureReason: reason,
		Gateway:       "mock-gateway",
	}, nil
}

// RefundPayment simulates refund processing
func (g *MockPaymentGateway) RefundPayment(ctx context.Context, transactionID string, amount float64) (*RefundResult, error) {
	// Simulate processing time
	time.Sleep(100 * time.Millisecond)

	// Simulate 95% success rate for refunds
	if rand.Float64() < 0.95 {
		return &RefundResult{
			Success:  true,
			RefundID: fmt.Sprintf("ref_%s_%d", transactionID[:8], time.Now().Unix()),
		}, nil
	}

	return &RefundResult{
		Success:       false,
		FailureReason: "refund processing failed",
	}, nil
}