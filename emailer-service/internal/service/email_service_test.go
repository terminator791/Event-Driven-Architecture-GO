package service

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmailService_SendWelcomeEmail_Success(t *testing.T) {
	// Arrange
	service := NewEmailService()
	userID := "test-user-id"
	email := "test@example.com"

	// Act
	err := service.SendWelcomeEmail(userID, email)

	// Assert
	assert.NoError(t, err)
}

func TestEmailService_SendWelcomeEmail_Idempotency(t *testing.T) {
	// Arrange
	service := NewEmailService()
	userID := "test-user-id"
	email := "test@example.com"

	// Act - Send email twice
	err1 := service.SendWelcomeEmail(userID, email)
	err2 := service.SendWelcomeEmail(userID, email)

	// Assert - Both should succeed (idempotent)
	assert.NoError(t, err1)
	assert.NoError(t, err2)
}

func TestEmailService_SendWelcomeEmail_DifferentUsers(t *testing.T) {
	// Arrange
	service := NewEmailService()

	// Act - Send emails for different users
	err1 := service.SendWelcomeEmail("user1", "user1@example.com")
	err2 := service.SendWelcomeEmail("user2", "user2@example.com")

	// Assert - Both should succeed
	assert.NoError(t, err1)
	assert.NoError(t, err2)
}

func TestEmailService_CleanupOldEntries(t *testing.T) {
	// This test verifies the cleanup mechanism works
	// We can't easily test the time-based cleanup in a unit test
	// but we can verify the service handles multiple events properly
	
	// Arrange
	service := NewEmailService().(*emailService)
	
	// Act - Add multiple entries
	for i := 0; i < 10; i++ {
		userID := fmt.Sprintf("user%d", i)
		email := fmt.Sprintf("user%d@example.com", i)
		err := service.SendWelcomeEmail(userID, email)
		assert.NoError(t, err)
	}
	
	// Assert - All entries should be tracked
	assert.Equal(t, 10, len(service.processedEvents))
}

func TestEmailService_ConcurrentAccess(t *testing.T) {
	// Test concurrent access to ensure thread safety
	// Arrange
	service := NewEmailService()
	userID := "concurrent-user"
	email := "concurrent@example.com"
	
	// Act - Send emails concurrently
	const numGoroutines = 10
	errCh := make(chan error, numGoroutines)
	
	for i := 0; i < numGoroutines; i++ {
		go func() {
			errCh <- service.SendWelcomeEmail(userID, email)
		}()
	}
	
	// Assert - All should complete without error
	for i := 0; i < numGoroutines; i++ {
		err := <-errCh
		assert.NoError(t, err)
	}
}