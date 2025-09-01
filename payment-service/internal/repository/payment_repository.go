package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/terminator791/Event-Driven-Architecture-GO/shared/pkg/models"
)

// PaymentRepository defines the interface for payment data operations
type PaymentRepository interface {
	Create(ctx context.Context, payment *models.Payment) error
	GetByID(ctx context.Context, id string) (*models.Payment, error)
	GetByOrderID(ctx context.Context, orderID string) (*models.Payment, error)
	UpdateStatus(ctx context.Context, id string, status models.PaymentStatus, transactionID, failureReason string) error
	GetPendingPayments(ctx context.Context, limit int) ([]*models.Payment, error)
	IncrementRetryCount(ctx context.Context, id string) error
}

// paymentRepository implements PaymentRepository
type paymentRepository struct {
	db *sqlx.DB
}

// NewPaymentRepository creates a new payment repository
func NewPaymentRepository(db *sqlx.DB) PaymentRepository {
	return &paymentRepository{db: db}
}

// Create creates a new payment
func (r *paymentRepository) Create(ctx context.Context, payment *models.Payment) error {
	// Generate ID if not provided
	if payment.ID == "" {
		payment.ID = uuid.New().String()
	}

	query := `
		INSERT INTO payments (id, order_id, user_id, amount, currency, payment_method, 
						     status, transaction_id, failure_reason, retry_count, processed_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`

	_, err := r.db.ExecContext(ctx, query,
		payment.ID, payment.OrderID, payment.UserID, payment.Amount, payment.Currency,
		payment.PaymentMethod, payment.Status, payment.TransactionID, payment.FailureReason,
		payment.RetryCount, payment.ProcessedAt, payment.CreatedAt, payment.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create payment: %w", err)
	}

	return nil
}

// GetByID retrieves a payment by ID
func (r *paymentRepository) GetByID(ctx context.Context, id string) (*models.Payment, error) {
	var payment models.Payment

	query := `
		SELECT id, order_id, user_id, amount, currency, payment_method, status,
			   transaction_id, failure_reason, retry_count, processed_at, created_at, updated_at
		FROM payments WHERE id = $1`

	err := r.db.GetContext(ctx, &payment, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}

	return &payment, nil
}

// GetByOrderID retrieves a payment by order ID
func (r *paymentRepository) GetByOrderID(ctx context.Context, orderID string) (*models.Payment, error) {
	var payment models.Payment

	query := `
		SELECT id, order_id, user_id, amount, currency, payment_method, status,
			   transaction_id, failure_reason, retry_count, processed_at, created_at, updated_at
		FROM payments WHERE order_id = $1`

	err := r.db.GetContext(ctx, &payment, query, orderID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get payment by order ID: %w", err)
	}

	return &payment, nil
}

// UpdateStatus updates a payment's status
func (r *paymentRepository) UpdateStatus(ctx context.Context, id string, status models.PaymentStatus, transactionID, failureReason string) error {
	query := `
		UPDATE payments 
		SET status = $1, transaction_id = $2, failure_reason = $3, 
		    processed_at = CASE WHEN $1 = 'completed' THEN NOW() ELSE processed_at END,
		    updated_at = NOW() 
		WHERE id = $4`

	_, err := r.db.ExecContext(ctx, query, status, transactionID, failureReason, id)
	if err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	return nil
}

// GetPendingPayments retrieves pending payments for retry processing
func (r *paymentRepository) GetPendingPayments(ctx context.Context, limit int) ([]*models.Payment, error) {
	query := `
		SELECT id, order_id, user_id, amount, currency, payment_method, status,
			   transaction_id, failure_reason, retry_count, processed_at, created_at, updated_at
		FROM payments 
		WHERE status = 'pending' OR status = 'failed'
		ORDER BY created_at ASC
		LIMIT $1`

	var payments []*models.Payment
	err := r.db.SelectContext(ctx, &payments, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending payments: %w", err)
	}

	return payments, nil
}

// IncrementRetryCount increments the retry count for a payment
func (r *paymentRepository) IncrementRetryCount(ctx context.Context, id string) error {
	query := `UPDATE payments SET retry_count = retry_count + 1, updated_at = NOW() WHERE id = $1`
	
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to increment retry count: %w", err)
	}

	return nil
}