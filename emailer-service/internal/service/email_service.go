package service

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// EmailService defines the interface for email operations
type EmailService interface {
	SendWelcomeEmail(userID, email string) error
}

// emailService implements EmailService
type emailService struct {
	processedEvents map[string]time.Time // For idempotency
	mutex           sync.RWMutex
}

// NewEmailService creates a new email service
func NewEmailService() EmailService {
	return &emailService{
		processedEvents: make(map[string]time.Time),
	}
}

// SendWelcomeEmail simulates sending a welcome email
func (s *emailService) SendWelcomeEmail(userID, email string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Check if we've already processed this event (idempotency)
	eventKey := fmt.Sprintf("welcome_%s", userID)
	if processedAt, exists := s.processedEvents[eventKey]; exists {
		log.Printf("Event already processed at %v for user %s, skipping", processedAt, email)
		return nil
	}

	// Simulate sending email
	log.Printf("📧 Welcome email sent to: %s (User ID: %s)", email, userID)
	
	// Additional simulation of email processing
	time.Sleep(100 * time.Millisecond) // Simulate network latency
	
	// Mark as processed
	s.processedEvents[eventKey] = time.Now()
	
	// Clean up old entries (prevent memory leak)
	s.cleanupOldEntries()
	
	return nil
}

// cleanupOldEntries removes entries older than 1 hour to prevent memory leaks
func (s *emailService) cleanupOldEntries() {
	cutoff := time.Now().Add(-1 * time.Hour)
	for key, processedAt := range s.processedEvents {
		if processedAt.Before(cutoff) {
			delete(s.processedEvents, key)
		}
	}
}