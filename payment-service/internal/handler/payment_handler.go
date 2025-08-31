package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/terminator791/Event-Driven-Architecture-GO/payment-service/internal/service"
	"github.com/terminator791/Event-Driven-Architecture-GO/shared/pkg/models"
)

// PaymentHandler handles HTTP requests for payments
type PaymentHandler struct {
	paymentService service.PaymentService
}

// NewPaymentHandler creates a new payment handler
func NewPaymentHandler(paymentService service.PaymentService) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
	}
}

// ProcessPayment handles payment processing requests
func (h *PaymentHandler) ProcessPayment(c *gin.Context) {
	var req models.ProcessPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.paymentService.ProcessPayment(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// GetPayment handles getting a payment by ID
func (h *PaymentHandler) GetPayment(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payment ID is required"})
		return
	}

	payment, err := h.paymentService.GetPayment(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "payment not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, payment)
}

// GetPaymentByOrderID handles getting a payment by order ID
func (h *PaymentHandler) GetPaymentByOrderID(c *gin.Context) {
	orderID := c.Param("orderId")
	if orderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "order ID is required"})
		return
	}

	payment, err := h.paymentService.GetPaymentByOrderID(c.Request.Context(), orderID)
	if err != nil {
		if err.Error() == "payment not found for order" {
			c.JSON(http.StatusNotFound, gin.H{"error": "payment not found for order"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, payment)
}

// RefundPayment handles payment refund requests
func (h *PaymentHandler) RefundPayment(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payment ID is required"})
		return
	}

	var req struct {
		Amount float64 `json:"amount" binding:"required,min=0.01"`
		Reason string  `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.paymentService.RefundPayment(c.Request.Context(), id, req.Amount, req.Reason)
	if err != nil {
		if err.Error() == "payment not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "payment refunded successfully"})
}

// RetryPendingPayments handles manual retry of pending payments
func (h *PaymentHandler) RetryPendingPayments(c *gin.Context) {
	err := h.paymentService.RetryPendingPayments(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "pending payments retry initiated"})
}

// HealthCheck handles health check requests
func (h *PaymentHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "payment-service",
	})
}