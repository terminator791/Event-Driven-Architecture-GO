package events

import (
	"encoding/json"
	"time"

	"github.com/terminator791/Event-Driven-Architecture-GO/shared/pkg/models"
)

// OrderCreated represents the event published when an order is created
type OrderCreated struct {
	BaseEvent
	OrderID         string                `json:"order_id"`
	UserID          string                `json:"user_id"`
	TotalAmount     float64               `json:"total_amount"`
	Currency        string                `json:"currency"`
	PaymentMethod   models.PaymentMethod  `json:"payment_method"`
	Status          models.OrderStatus    `json:"status"`
	Items           []OrderItemData       `json:"items"`
	ShippingAddress models.Address        `json:"shipping_address"`
	BillingAddress  models.Address        `json:"billing_address"`
	CreatedAt       time.Time             `json:"created_at"`
}

// OrderItemData represents order item data in events
type OrderItemData struct {
	ProductID  string  `json:"product_id"`
	Quantity   int     `json:"quantity"`
	UnitPrice  float64 `json:"unit_price"`
	TotalPrice float64 `json:"total_price"`
}

// NewOrderCreated creates a new OrderCreated event
func NewOrderCreated(order *models.Order) *OrderCreated {
	items := make([]OrderItemData, len(order.Items))
	for i, item := range order.Items {
		items[i] = OrderItemData{
			ProductID:  item.ProductID,
			Quantity:   item.Quantity,
			UnitPrice:  item.UnitPrice,
			TotalPrice: item.TotalPrice,
		}
	}

	return &OrderCreated{
		BaseEvent: BaseEvent{
			Metadata: NewEventMetadata("order-service", "1.0"),
		},
		OrderID:         order.ID,
		UserID:          order.UserID,
		TotalAmount:     order.TotalAmount,
		Currency:        order.Currency,
		PaymentMethod:   order.PaymentMethod,
		Status:          order.Status,
		Items:           items,
		ShippingAddress: order.ShippingAddress,
		BillingAddress:  order.BillingAddress,
		CreatedAt:       order.CreatedAt,
	}
}

// GetMetadata returns the event metadata
func (o *OrderCreated) GetMetadata() EventMetadata {
	return o.Metadata
}

// GetEventType returns the event type
func (o *OrderCreated) GetEventType() string {
	return "order.created"
}

// ToJSON serializes the event to JSON
func (o *OrderCreated) ToJSON() ([]byte, error) {
	return json.Marshal(o)
}

// FromJSON deserializes the event from JSON
func (o *OrderCreated) FromJSON(data []byte) error {
	return json.Unmarshal(data, o)
}

// OrderStatusUpdated represents the event published when an order status is updated
type OrderStatusUpdated struct {
	BaseEvent
	OrderID       string             `json:"order_id"`
	UserID        string             `json:"user_id"`
	PreviousStatus models.OrderStatus `json:"previous_status"`
	NewStatus     models.OrderStatus `json:"new_status"`
	Reason        string             `json:"reason,omitempty"`
	UpdatedAt     time.Time          `json:"updated_at"`
}

// NewOrderStatusUpdated creates a new OrderStatusUpdated event
func NewOrderStatusUpdated(orderID, userID string, previousStatus, newStatus models.OrderStatus, reason string) *OrderStatusUpdated {
	return &OrderStatusUpdated{
		BaseEvent: BaseEvent{
			Metadata: NewEventMetadata("order-service", "1.0"),
		},
		OrderID:       orderID,
		UserID:        userID,
		PreviousStatus: previousStatus,
		NewStatus:     newStatus,
		Reason:        reason,
		UpdatedAt:     time.Now().UTC(),
	}
}

// GetMetadata returns the event metadata
func (o *OrderStatusUpdated) GetMetadata() EventMetadata {
	return o.Metadata
}

// GetEventType returns the event type
func (o *OrderStatusUpdated) GetEventType() string {
	return "order.status_updated"
}

// ToJSON serializes the event to JSON
func (o *OrderStatusUpdated) ToJSON() ([]byte, error) {
	return json.Marshal(o)
}

// FromJSON deserializes the event from JSON
func (o *OrderStatusUpdated) FromJSON(data []byte) error {
	return json.Unmarshal(data, o)
}

// OrderCancelled represents the event published when an order is cancelled
type OrderCancelled struct {
	BaseEvent
	OrderID     string    `json:"order_id"`
	UserID      string    `json:"user_id"`
	Reason      string    `json:"reason"`
	CancelledAt time.Time `json:"cancelled_at"`
}

// NewOrderCancelled creates a new OrderCancelled event
func NewOrderCancelled(orderID, userID, reason string) *OrderCancelled {
	return &OrderCancelled{
		BaseEvent: BaseEvent{
			Metadata: NewEventMetadata("order-service", "1.0"),
		},
		OrderID:     orderID,
		UserID:      userID,
		Reason:      reason,
		CancelledAt: time.Now().UTC(),
	}
}

// GetMetadata returns the event metadata
func (o *OrderCancelled) GetMetadata() EventMetadata {
	return o.Metadata
}

// GetEventType returns the event type
func (o *OrderCancelled) GetEventType() string {
	return "order.cancelled"
}

// ToJSON serializes the event to JSON
func (o *OrderCancelled) ToJSON() ([]byte, error) {
	return json.Marshal(o)
}

// FromJSON deserializes the event from JSON
func (o *OrderCancelled) FromJSON(data []byte) error {
	return json.Unmarshal(data, o)
}