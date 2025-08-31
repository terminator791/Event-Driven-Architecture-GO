package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/terminator791/Event-Driven-Architecture-GO/order-service/internal/repository"
	"github.com/terminator791/Event-Driven-Architecture-GO/shared/pkg/events"
	"github.com/terminator791/Event-Driven-Architecture-GO/shared/pkg/models"
)

// ProducerInterface defines the interface for publishing messages to Kafka
type ProducerInterface interface {
	Publish(ctx context.Context, key, value []byte) error
	Close() error
}

// ProductService defines the interface for product operations
type ProductService interface {
	GetByID(ctx context.Context, id string) (*models.Product, error)
	GetMultiple(ctx context.Context, ids []string) ([]*models.Product, error)
}

// OrderService defines the interface for order business logic
type OrderService interface {
	CreateOrder(ctx context.Context, req models.CreateOrderRequest) (*models.CreateOrderResponse, error)
	GetOrder(ctx context.Context, id string) (*models.Order, error)
	GetUserOrders(ctx context.Context, userID string, limit, offset int) ([]*models.Order, error)
	UpdateOrderStatus(ctx context.Context, id string, req models.UpdateOrderStatusRequest) error
	CancelOrder(ctx context.Context, id, userID, reason string) error
}

// orderService implements OrderService
type orderService struct {
	orderRepo      repository.OrderRepository
	productService ProductService
	producer       ProducerInterface
}

// NewOrderService creates a new order service
func NewOrderService(orderRepo repository.OrderRepository, productService ProductService, producer ProducerInterface) OrderService {
	return &orderService{
		orderRepo:      orderRepo,
		productService: productService,
		producer:       producer,
	}
}

// CreateOrder creates a new order with business logic validation
func (s *orderService) CreateOrder(ctx context.Context, req models.CreateOrderRequest) (*models.CreateOrderResponse, error) {
	// Validate request
	if err := s.validateCreateOrderRequest(req); err != nil {
		return nil, err
	}

	// Get product information for all items
	productIDs := make([]string, len(req.Items))
	for i, item := range req.Items {
		productIDs[i] = item.ProductID
	}

	products, err := s.productService.GetMultiple(ctx, productIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get products: %w", err)
	}

	if len(products) != len(productIDs) {
		return nil, fmt.Errorf("some products not found")
	}

	// Create product map for easy lookup
	productMap := make(map[string]*models.Product)
	for _, product := range products {
		productMap[product.ID] = product
	}

	// Calculate order details and create items
	var totalAmount float64
	orderItems := make([]models.OrderItem, len(req.Items))

	for i, reqItem := range req.Items {
		product, exists := productMap[reqItem.ProductID]
		if !exists {
			return nil, fmt.Errorf("product %s not found", reqItem.ProductID)
		}

		if !product.InStock || product.StockQty < reqItem.Quantity {
			return nil, fmt.Errorf("product %s is out of stock or insufficient quantity", product.Name)
		}

		itemTotal := product.Price * float64(reqItem.Quantity)
		totalAmount += itemTotal

		orderItems[i] = models.OrderItem{
			ID:         uuid.New().String(),
			ProductID:  reqItem.ProductID,
			Quantity:   reqItem.Quantity,
			UnitPrice:  product.Price,
			TotalPrice: itemTotal,
			Product:    product,
		}
	}

	// Create order
	now := time.Now().UTC()
	order := &models.Order{
		ID:              uuid.New().String(),
		UserID:          req.UserID,
		Status:          models.OrderStatusPending,
		TotalAmount:     totalAmount,
		Currency:        req.Currency,
		PaymentStatus:   models.PaymentStatusPending,
		PaymentMethod:   req.PaymentMethod,
		ShippingAddress: req.ShippingAddress,
		BillingAddress:  req.BillingAddress,
		Items:           orderItems,
		Notes:           req.Notes,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	// Save order to database
	if err := s.orderRepo.Create(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Publish OrderCreated event
	if err := s.publishOrderCreatedEvent(ctx, order); err != nil {
		// Log error but don't fail the order creation
		fmt.Printf("Warning: failed to publish order created event: %v\n", err)
	}

	return &models.CreateOrderResponse{
		ID:            order.ID,
		UserID:        order.UserID,
		Status:        order.Status,
		TotalAmount:   order.TotalAmount,
		Currency:      order.Currency,
		PaymentStatus: order.PaymentStatus,
		CreatedAt:     order.CreatedAt,
	}, nil
}

// GetOrder retrieves an order by ID
func (s *orderService) GetOrder(ctx context.Context, id string) (*models.Order, error) {
	order, err := s.orderRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	if order == nil {
		return nil, fmt.Errorf("order not found")
	}

	return order, nil
}

// GetUserOrders retrieves orders for a user with pagination
func (s *orderService) GetUserOrders(ctx context.Context, userID string, limit, offset int) ([]*models.Order, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	orders, err := s.orderRepo.GetByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get user orders: %w", err)
	}

	return orders, nil
}

// UpdateOrderStatus updates an order's status
func (s *orderService) UpdateOrderStatus(ctx context.Context, id string, req models.UpdateOrderStatusRequest) error {
	// Get current order
	order, err := s.orderRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	if order == nil {
		return fmt.Errorf("order not found")
	}

	previousStatus := order.Status

	// Validate status transition
	if err := s.validateStatusTransition(previousStatus, req.Status); err != nil {
		return err
	}

	// Update status
	if err := s.orderRepo.UpdateStatus(ctx, id, req.Status, req.Notes); err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	// Publish OrderStatusUpdated event
	if err := s.publishOrderStatusUpdatedEvent(ctx, order.ID, order.UserID, previousStatus, req.Status, req.Notes); err != nil {
		fmt.Printf("Warning: failed to publish order status updated event: %v\n", err)
	}

	return nil
}

// CancelOrder cancels an order
func (s *orderService) CancelOrder(ctx context.Context, id, userID, reason string) error {
	// Get current order
	order, err := s.orderRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	if order == nil {
		return fmt.Errorf("order not found")
	}

	if order.UserID != userID {
		return fmt.Errorf("order does not belong to user")
	}

	// Check if order can be cancelled
	if order.Status == models.OrderStatusCancelled || order.Status == models.OrderStatusDelivered {
		return fmt.Errorf("order cannot be cancelled in current status: %s", order.Status)
	}

	// Update status to cancelled
	if err := s.orderRepo.UpdateStatus(ctx, id, models.OrderStatusCancelled, reason); err != nil {
		return fmt.Errorf("failed to cancel order: %w", err)
	}

	// Publish OrderCancelled event
	if err := s.publishOrderCancelledEvent(ctx, order.ID, order.UserID, reason); err != nil {
		fmt.Printf("Warning: failed to publish order cancelled event: %v\n", err)
	}

	return nil
}

// validateCreateOrderRequest validates the create order request
func (s *orderService) validateCreateOrderRequest(req models.CreateOrderRequest) error {
	if req.UserID == "" {
		return fmt.Errorf("user ID is required")
	}

	if len(req.Items) == 0 {
		return fmt.Errorf("at least one item is required")
	}

	if req.Currency == "" {
		return fmt.Errorf("currency is required")
	}

	if req.PaymentMethod == "" {
		return fmt.Errorf("payment method is required")
	}

	if req.ShippingAddress.Street == "" || req.ShippingAddress.City == "" ||
		req.ShippingAddress.Country == "" {
		return fmt.Errorf("complete shipping address is required")
	}

	if req.BillingAddress.Street == "" || req.BillingAddress.City == "" ||
		req.BillingAddress.Country == "" {
		return fmt.Errorf("complete billing address is required")
	}

	for i, item := range req.Items {
		if item.ProductID == "" {
			return fmt.Errorf("product ID is required for item %d", i+1)
		}
		if item.Quantity <= 0 {
			return fmt.Errorf("quantity must be positive for item %d", i+1)
		}
	}

	return nil
}

// validateStatusTransition validates that a status transition is allowed
func (s *orderService) validateStatusTransition(from, to models.OrderStatus) error {
	validTransitions := map[models.OrderStatus][]models.OrderStatus{
		models.OrderStatusPending:    {models.OrderStatusConfirmed, models.OrderStatusCancelled},
		models.OrderStatusConfirmed:  {models.OrderStatusProcessing, models.OrderStatusCancelled},
		models.OrderStatusProcessing: {models.OrderStatusShipped, models.OrderStatusCancelled},
		models.OrderStatusShipped:    {models.OrderStatusDelivered},
		models.OrderStatusDelivered:  {models.OrderStatusRefunded},
		models.OrderStatusCancelled:  {}, // No transitions from cancelled
		models.OrderStatusRefunded:   {}, // No transitions from refunded
	}

	validNextStates, exists := validTransitions[from]
	if !exists {
		return fmt.Errorf("invalid current status: %s", from)
	}

	for _, validState := range validNextStates {
		if validState == to {
			return nil
		}
	}

	return fmt.Errorf("invalid status transition from %s to %s", from, to)
}

// publishOrderCreatedEvent publishes an order created event to Kafka
func (s *orderService) publishOrderCreatedEvent(ctx context.Context, order *models.Order) error {
	event := events.NewOrderCreated(order)

	// Set correlation ID from context if available
	if correlationID := ctx.Value("correlation_id"); correlationID != nil {
		if strCorrelationID, ok := correlationID.(string); ok {
			event.Metadata.WithCorrelationID(strCorrelationID)
		}
	}

	// Set user ID in metadata
	event.Metadata.WithUserID(order.UserID)

	eventData, err := event.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize event: %w", err)
	}

	return s.producer.Publish(ctx, []byte(order.ID), eventData)
}

// publishOrderStatusUpdatedEvent publishes an order status updated event
func (s *orderService) publishOrderStatusUpdatedEvent(ctx context.Context, orderID, userID string, previousStatus, newStatus models.OrderStatus, reason string) error {
	event := events.NewOrderStatusUpdated(orderID, userID, previousStatus, newStatus, reason)

	// Set correlation ID from context if available
	if correlationID := ctx.Value("correlation_id"); correlationID != nil {
		if strCorrelationID, ok := correlationID.(string); ok {
			event.Metadata.WithCorrelationID(strCorrelationID)
		}
	}

	// Set user ID in metadata
	event.Metadata.WithUserID(userID)

	eventData, err := event.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize event: %w", err)
	}

	return s.producer.Publish(ctx, []byte(orderID), eventData)
}

// publishOrderCancelledEvent publishes an order cancelled event
func (s *orderService) publishOrderCancelledEvent(ctx context.Context, orderID, userID, reason string) error {
	event := events.NewOrderCancelled(orderID, userID, reason)

	// Set correlation ID from context if available
	if correlationID := ctx.Value("correlation_id"); correlationID != nil {
		if strCorrelationID, ok := correlationID.(string); ok {
			event.Metadata.WithCorrelationID(strCorrelationID)
		}
	}

	// Set user ID in metadata
	event.Metadata.WithUserID(userID)

	eventData, err := event.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize event: %w", err)
	}

	return s.producer.Publish(ctx, []byte(orderID), eventData)
}