package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/terminator791/Event-Driven-Architecture-GO/shared/pkg/models"
	"github.com/terminator791/Event-Driven-Architecture-GO/user-api/internal/service"
)

// UserHandler handles HTTP requests for user operations
type UserHandler struct {
	userService service.UserService
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// CreateUser handles POST /users requests
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req models.CreateUserRequest
	
	// Bind and validate JSON request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Call service to create user
	response, err := h.userService.CreateUser(c.Request.Context(), req)
	if err != nil {
		// Check if it's a business logic error (user already exists)
		if contains(err.Error(), "already exists") {
			c.JSON(http.StatusConflict, gin.H{
				"error": "User already exists",
				"details": err.Error(),
			})
			return
		}

		// Check if it's a validation error
		if contains(err.Error(), "validation failed") {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Validation failed",
				"details": err.Error(),
			})
			return
		}

		// Internal server error
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create user",
			"details": err.Error(),
		})
		return
	}

	// Return success response
	c.JSON(http.StatusCreated, response)
}

// HealthCheck handles GET /health requests
func (h *UserHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"service": "user-api",
	})
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(s) > len(substr) && 
		(func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())))
}