package events

import (
	"encoding/json"
	"time"

	"github.com/terminator791/Event-Driven-Architecture-GO/shared/pkg/models"
)

// PaymentProcessed represents the event published when a payment is successfully processed
type PaymentProcessed struct {
	BaseEvent
	PaymentID     string                `json:"payment_id"`
	OrderID       string                `json:"order_id"`
	UserID        string                `json:"user_id"`
	Amount        float64               `json:"amount"`
	Currency      string                `json:"currency"`
	PaymentMethod models.PaymentMethod  `json:"payment_method"`
	TransactionID string                `json:"transaction_id"`
	ProcessedAt   time.Time             `json:"processed_at"`
}

// NewPaymentProcessed creates a new PaymentProcessed event
func NewPaymentProcessed(payment *models.Payment) *PaymentProcessed {
	return &PaymentProcessed{
		BaseEvent: BaseEvent{
			Metadata: NewEventMetadata("payment-service", "1.0"),
		},
		PaymentID:     payment.ID,
		OrderID:       payment.OrderID,
		UserID:        payment.UserID,
		Amount:        payment.Amount,
		Currency:      payment.Currency,
		PaymentMethod: payment.PaymentMethod,
		TransactionID: payment.TransactionID,
		ProcessedAt:   *payment.ProcessedAt,
	}
}

// GetMetadata returns the event metadata
func (p *PaymentProcessed) GetMetadata() EventMetadata {
	return p.Metadata
}

// GetEventType returns the event type
func (p *PaymentProcessed) GetEventType() string {
	return "payment.processed"
}

// ToJSON serializes the event to JSON
func (p *PaymentProcessed) ToJSON() ([]byte, error) {
	return json.Marshal(p)
}

// FromJSON deserializes the event from JSON
func (p *PaymentProcessed) FromJSON(data []byte) error {
	return json.Unmarshal(data, p)
}

// PaymentFailed represents the event published when a payment fails
type PaymentFailed struct {
	BaseEvent
	PaymentID     string                `json:"payment_id"`
	OrderID       string                `json:"order_id"`
	UserID        string                `json:"user_id"`
	Amount        float64               `json:"amount"`
	Currency      string                `json:"currency"`
	PaymentMethod models.PaymentMethod  `json:"payment_method"`
	FailureReason string                `json:"failure_reason"`
	FailedAt      time.Time             `json:"failed_at"`
	RetryCount    int                   `json:"retry_count"`
}

// NewPaymentFailed creates a new PaymentFailed event
func NewPaymentFailed(payment *models.Payment, retryCount int) *PaymentFailed {
	return &PaymentFailed{
		BaseEvent: BaseEvent{
			Metadata: NewEventMetadata("payment-service", "1.0"),
		},
		PaymentID:     payment.ID,
		OrderID:       payment.OrderID,
		UserID:        payment.UserID,
		Amount:        payment.Amount,
		Currency:      payment.Currency,
		PaymentMethod: payment.PaymentMethod,
		FailureReason: payment.FailureReason,
		FailedAt:      time.Now().UTC(),
		RetryCount:    retryCount,
	}
}

// GetMetadata returns the event metadata
func (p *PaymentFailed) GetMetadata() EventMetadata {
	return p.Metadata
}

// GetEventType returns the event type
func (p *PaymentFailed) GetEventType() string {
	return "payment.failed"
}

// ToJSON serializes the event to JSON
func (p *PaymentFailed) ToJSON() ([]byte, error) {
	return json.Marshal(p)
}

// FromJSON deserializes the event from JSON
func (p *PaymentFailed) FromJSON(data []byte) error {
	return json.Unmarshal(data, p)
}

// PaymentRefunded represents the event published when a payment is refunded
type PaymentRefunded struct {
	BaseEvent
	PaymentID       string                `json:"payment_id"`
	OrderID         string                `json:"order_id"`
	UserID          string                `json:"user_id"`
	RefundAmount    float64               `json:"refund_amount"`
	OriginalAmount  float64               `json:"original_amount"`
	Currency        string                `json:"currency"`
	PaymentMethod   models.PaymentMethod  `json:"payment_method"`
	RefundReason    string                `json:"refund_reason"`
	RefundedAt      time.Time             `json:"refunded_at"`
	TransactionID   string                `json:"transaction_id"`
}

// NewPaymentRefunded creates a new PaymentRefunded event
func NewPaymentRefunded(payment *models.Payment, refundAmount float64, reason string) *PaymentRefunded {
	return &PaymentRefunded{
		BaseEvent: BaseEvent{
			Metadata: NewEventMetadata("payment-service", "1.0"),
		},
		PaymentID:       payment.ID,
		OrderID:         payment.OrderID,
		UserID:          payment.UserID,
		RefundAmount:    refundAmount,
		OriginalAmount:  payment.Amount,
		Currency:        payment.Currency,
		PaymentMethod:   payment.PaymentMethod,
		RefundReason:    reason,
		RefundedAt:      time.Now().UTC(),
		TransactionID:   payment.TransactionID,
	}
}

// GetMetadata returns the event metadata
func (p *PaymentRefunded) GetMetadata() EventMetadata {
	return p.Metadata
}

// GetEventType returns the event type
func (p *PaymentRefunded) GetEventType() string {
	return "payment.refunded"
}

// ToJSON serializes the event to JSON
func (p *PaymentRefunded) ToJSON() ([]byte, error) {
	return json.Marshal(p)
}

// FromJSON deserializes the event from JSON
func (p *PaymentRefunded) FromJSON(data []byte) error {
	return json.Unmarshal(data, p)
}