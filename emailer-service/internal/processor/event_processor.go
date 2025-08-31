package processor

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/terminator791/Event-Driven-Architecture-GO/emailer-service/internal/service"
	"github.com/terminator791/Event-Driven-Architecture-GO/shared/pkg/events"
	sharedKafka "github.com/terminator791/Event-Driven-Architecture-GO/shared/pkg/kafka"
)

// EventProcessor handles consuming and processing events from Kafka
type EventProcessor struct {
	consumer     *sharedKafka.Consumer
	emailService service.EmailService
	retryDelay   time.Duration
	maxRetries   int
}

// NewEventProcessor creates a new event processor
func NewEventProcessor(consumer *sharedKafka.Consumer, emailService service.EmailService) *EventProcessor {
	return &EventProcessor{
		consumer:     consumer,
		emailService: emailService,
		retryDelay:   2 * time.Second,
		maxRetries:   3,
	}
}

// Start begins processing events from Kafka
func (p *EventProcessor) Start(ctx context.Context) error {
	log.Println("🚀 Starting event processor...")

	for {
		select {
		case <-ctx.Done():
			log.Println("🛑 Event processor shutting down...")
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
		return fmt.Errorf("failed to consume message: %w", err)
	}

	log.Printf("📨 Received message: offset=%d, partition=%d", msg.Offset, msg.Partition)

	// Process with retry logic
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

		err := p.processUserCreatedEvent(msg.Value)
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

// processUserCreatedEvent processes a UserCreated event
func (p *EventProcessor) processUserCreatedEvent(data []byte) error {
	var event events.UserCreated
	if err := event.FromJSON(data); err != nil {
		return fmt.Errorf("failed to deserialize event: %w", err)
	}

	log.Printf("👤 Processing UserCreated event for user: %s (Event ID: %s)", event.Email, event.EventID)

	// Send welcome email
	if err := p.emailService.SendWelcomeEmail(event.ID, event.Email); err != nil {
		return fmt.Errorf("failed to send welcome email: %w", err)
	}

	return nil
}