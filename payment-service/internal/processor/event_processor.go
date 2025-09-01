package processor

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/terminator791/Event-Driven-Architecture-GO/payment-service/internal/service"
	"github.com/terminator791/Event-Driven-Architecture-GO/shared/pkg/events"
	sharedKafka "github.com/terminator791/Event-Driven-Architecture-GO/shared/pkg/kafka"
	"github.com/terminator791/Event-Driven-Architecture-GO/shared/pkg/models"
)

// EventProcessor handles consuming and processing events from Kafka
type EventProcessor struct {
	consumer       *sharedKafka.Consumer
	paymentService service.PaymentService
	retryDelay     time.Duration
	maxRetries     int
}

// NewEventProcessor creates a new event processor
func NewEventProcessor(consumer *sharedKafka.Consumer, paymentService service.PaymentService) *EventProcessor {
	return &EventProcessor{
		consumer:       consumer,
		paymentService: paymentService,
		retryDelay:     2 * time.Second,
		maxRetries:     3,
	}
}

// Start begins processing events from Kafka
func (p *EventProcessor) Start(ctx context.Context) error {
	log.Println("🚀 Starting payment event processor...")

	for {
		select {
		case <-ctx.Done():
			log.Println("🛑 Payment event processor shutting down...")
			return ctx.Err()
		default:
			if err := p.processMessage(ctx); err != nil {
				log.Printf("❌ Error processing message: %v", err)
				// Continue processing other messages
			}
		}
	}
}

// processMessage processes a single message with retry logic
func (p *EventProcessor) processMessage(ctx context.Context) error {
	msg, err := p.consumer.Consume(ctx)
	if err != nil {
		return fmt.Errorf("failed to read message: %w", err)
	}

	log.Printf("📨 Received message from topic %s, partition %d, offset %d",
		msg.Topic, msg.Partition, msg.Offset)

	return p.processWithRetry(msg)
}

// processWithRetry processes a message with retry logic
func (p *EventProcessor) processWithRetry(msg kafka.Message) error {
	var lastErr error

	for attempt := 0; attempt <= p.maxRetries; attempt++ {
		if attempt > 0 {
			log.Printf("🔄 Retry attempt %d/%d", attempt, p.maxRetries)
			time.Sleep(p.retryDelay * time.Duration(attempt)) // Exponential backoff
		}

		err := p.processEvent(msg.Value)
		if err == nil {
			log.Printf("✅ Successfully processed message after %d attempts", attempt+1)
			return nil
		}

		lastErr = err
		log.Printf("⚠️  Attempt %d failed: %v", attempt+1, err)
	}

	log.Printf("❌ Failed to process message after %d attempts: %v", p.maxRetries+1, lastErr)
	return lastErr
}

// processEvent processes different types of events
func (p *EventProcessor) processEvent(data []byte) error {
	// First, try to deserialize as an event envelope to determine the event type
	var envelope events.EventEnvelope
	if err := envelope.FromJSON(data); err != nil {
		// Fallback: try to process as legacy OrderCreated event
		return p.processLegacyOrderCreatedEvent(data)
	}

	// Process based on event type
	switch envelope.EventType {
	case "order.created":
		return p.processOrderCreatedEvent(envelope.Data)
	case "order.status_updated":
		return p.processOrderStatusUpdatedEvent(envelope.Data)
	case "order.cancelled":
		return p.processOrderCancelledEvent(envelope.Data)
	default:
		log.Printf("⚠️  Unknown event type: %s, skipping", envelope.EventType)
		return nil // Don't treat unknown events as errors
	}
}

// processOrderCreatedEvent processes OrderCreated events
func (p *EventProcessor) processOrderCreatedEvent(data []byte) error {
	var event events.OrderCreated
	if err := event.FromJSON(data); err != nil {
		return fmt.Errorf("failed to deserialize OrderCreated event: %w", err)
	}

	log.Printf("🛒 Processing OrderCreated event for order %s (amount: %.2f %s)",
		event.OrderID, event.TotalAmount, event.Currency)

	// Automatically initiate payment processing for new orders
	paymentRequest := models.ProcessPaymentRequest{
		OrderID:       event.OrderID,
		PaymentMethod: event.PaymentMethod,
		Amount:        event.TotalAmount,
		Currency:      event.Currency,
	}

	ctx := context.Background()
	// Set correlation ID for distributed tracing
	if event.Metadata.CorrelationID != "" {
		ctx = context.WithValue(ctx, "correlation_id", event.Metadata.CorrelationID)
	}

	_, err := p.paymentService.ProcessPayment(ctx, paymentRequest)
	if err != nil {
		log.Printf("❌ Failed to process payment for order %s: %v", event.OrderID, err)
		return err
	}

	log.Printf("✅ Successfully initiated payment for order %s", event.OrderID)
	return nil
}

// processOrderStatusUpdatedEvent processes OrderStatusUpdated events
func (p *EventProcessor) processOrderStatusUpdatedEvent(data []byte) error {
	var event events.OrderStatusUpdated
	if err := event.FromJSON(data); err != nil {
		return fmt.Errorf("failed to deserialize OrderStatusUpdated event: %w", err)
	}

	log.Printf("📝 Processing OrderStatusUpdated event for order %s (%s -> %s)",
		event.OrderID, event.PreviousStatus, event.NewStatus)

	// Handle order status changes that affect payments
	if event.NewStatus == models.OrderStatusCancelled {
		// Check if there's a completed payment that needs to be refunded
		ctx := context.Background()
		if event.Metadata.CorrelationID != "" {
			ctx = context.WithValue(ctx, "correlation_id", event.Metadata.CorrelationID)
		}

		payment, err := p.paymentService.GetPaymentByOrderID(ctx, event.OrderID)
		if err != nil {
			log.Printf("⚠️  No payment found for cancelled order %s", event.OrderID)
			return nil // This is not necessarily an error
		}

		if payment.Status == models.PaymentStatusCompleted {
			log.Printf("💰 Initiating refund for cancelled order %s", event.OrderID)
			err := p.paymentService.RefundPayment(ctx, payment.ID, payment.Amount, "Order cancelled")
			if err != nil {
				log.Printf("❌ Failed to refund payment for cancelled order %s: %v", event.OrderID, err)
				return err
			}
			log.Printf("✅ Successfully refunded payment for cancelled order %s", event.OrderID)
		}
	}

	return nil
}

// processOrderCancelledEvent processes OrderCancelled events
func (p *EventProcessor) processOrderCancelledEvent(data []byte) error {
	var event events.OrderCancelled
	if err := event.FromJSON(data); err != nil {
		return fmt.Errorf("failed to deserialize OrderCancelled event: %w", err)
	}

	log.Printf("🚫 Processing OrderCancelled event for order %s", event.OrderID)

	// Handle refund for cancelled orders (similar to order status update)
	ctx := context.Background()
	if event.Metadata.CorrelationID != "" {
		ctx = context.WithValue(ctx, "correlation_id", event.Metadata.CorrelationID)
	}

	payment, err := p.paymentService.GetPaymentByOrderID(ctx, event.OrderID)
	if err != nil {
		log.Printf("⚠️  No payment found for cancelled order %s", event.OrderID)
		return nil
	}

	if payment.Status == models.PaymentStatusCompleted {
		log.Printf("💰 Initiating refund for cancelled order %s", event.OrderID)
		err := p.paymentService.RefundPayment(ctx, payment.ID, payment.Amount, event.Reason)
		if err != nil {
			log.Printf("❌ Failed to refund payment for cancelled order %s: %v", event.OrderID, err)
			return err
		}
		log.Printf("✅ Successfully refunded payment for cancelled order %s", event.OrderID)
	}

	return nil
}

// processLegacyOrderCreatedEvent processes legacy OrderCreated events (for backward compatibility)
func (p *EventProcessor) processLegacyOrderCreatedEvent(data []byte) error {
	// Try to process as legacy UserCreated event (just log and ignore)
	var userEvent events.UserCreated
	if err := userEvent.FromJSON(data); err == nil {
		log.Printf("👤 Received UserCreated event for user %s (ignored by payment service)", userEvent.Email)
		return nil
	}

	log.Printf("⚠️  Unable to process legacy event format, skipping")
	return nil
}