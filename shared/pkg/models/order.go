package models

import (
	"time"
)

// OrderStatus represents the status of an order
type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "pending"
	OrderStatusConfirmed  OrderStatus = "confirmed"
	OrderStatusProcessing OrderStatus = "processing"
	OrderStatusShipped    OrderStatus = "shipped"
	OrderStatusDelivered  OrderStatus = "delivered"
	OrderStatusCancelled  OrderStatus = "cancelled"
	OrderStatusRefunded   OrderStatus = "refunded"
)

// PaymentStatus represents the status of a payment
type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusCompleted PaymentStatus = "completed"
	PaymentStatusFailed    PaymentStatus = "failed"
	PaymentStatusRefunded  PaymentStatus = "refunded"
)

// PaymentMethod represents a payment method
type PaymentMethod string

const (
	PaymentMethodCreditCard PaymentMethod = "credit_card"
	PaymentMethodDebitCard  PaymentMethod = "debit_card"
	PaymentMethodPayPal     PaymentMethod = "paypal"
	PaymentMethodBankTransfer PaymentMethod = "bank_transfer"
)

// Product represents a product in the system
type Product struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Price       float64   `json:"price" db:"price"`
	SKU         string    `json:"sku" db:"sku"`
	Category    string    `json:"category" db:"category"`
	InStock     bool      `json:"in_stock" db:"in_stock"`
	StockQty    int       `json:"stock_quantity" db:"stock_quantity"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// OrderItem represents an item in an order
type OrderItem struct {
	ID        string  `json:"id" db:"id"`
	OrderID   string  `json:"order_id" db:"order_id"`
	ProductID string  `json:"product_id" db:"product_id"`
	Quantity  int     `json:"quantity" db:"quantity"`
	UnitPrice float64 `json:"unit_price" db:"unit_price"`
	TotalPrice float64 `json:"total_price" db:"total_price"`
	Product   *Product `json:"product,omitempty"`
}

// Order represents an order in the system
type Order struct {
	ID            string        `json:"id" db:"id"`
	UserID        string        `json:"user_id" db:"user_id"`
	Status        OrderStatus   `json:"status" db:"status"`
	TotalAmount   float64       `json:"total_amount" db:"total_amount"`
	Currency      string        `json:"currency" db:"currency"`
	PaymentStatus PaymentStatus `json:"payment_status" db:"payment_status"`
	PaymentMethod PaymentMethod `json:"payment_method" db:"payment_method"`
	ShippingAddress Address     `json:"shipping_address"`
	BillingAddress  Address     `json:"billing_address"`
	Items         []OrderItem   `json:"items,omitempty"`
	CreatedAt     time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at" db:"updated_at"`
	Notes         string        `json:"notes,omitempty" db:"notes"`
}

// Address represents a shipping or billing address
type Address struct {
	Street     string `json:"street"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
}

// CreateOrderRequest represents the request to create a new order
type CreateOrderRequest struct {
	UserID          string        `json:"user_id" binding:"required"`
	Items           []OrderItemRequest `json:"items" binding:"required,min=1"`
	Currency        string        `json:"currency" binding:"required"`
	PaymentMethod   PaymentMethod `json:"payment_method" binding:"required"`
	ShippingAddress Address       `json:"shipping_address" binding:"required"`
	BillingAddress  Address       `json:"billing_address" binding:"required"`
	Notes           string        `json:"notes,omitempty"`
}

// OrderItemRequest represents an item in a create order request
type OrderItemRequest struct {
	ProductID string `json:"product_id" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required,min=1"`
}

// CreateOrderResponse represents the response after creating an order
type CreateOrderResponse struct {
	ID            string        `json:"id"`
	UserID        string        `json:"user_id"`
	Status        OrderStatus   `json:"status"`
	TotalAmount   float64       `json:"total_amount"`
	Currency      string        `json:"currency"`
	PaymentStatus PaymentStatus `json:"payment_status"`
	CreatedAt     time.Time     `json:"created_at"`
}

// UpdateOrderStatusRequest represents the request to update order status
type UpdateOrderStatusRequest struct {
	Status OrderStatus `json:"status" binding:"required"`
	Notes  string      `json:"notes,omitempty"`
}

// Payment represents a payment in the system
type Payment struct {
	ID              string        `json:"id" db:"id"`
	OrderID         string        `json:"order_id" db:"order_id"`
	UserID          string        `json:"user_id" db:"user_id"`
	Amount          float64       `json:"amount" db:"amount"`
	Currency        string        `json:"currency" db:"currency"`
	PaymentMethod   PaymentMethod `json:"payment_method" db:"payment_method"`
	Status          PaymentStatus `json:"status" db:"status"`
	TransactionID   string        `json:"transaction_id,omitempty" db:"transaction_id"`
	FailureReason   string        `json:"failure_reason,omitempty" db:"failure_reason"`
	ProcessedAt     *time.Time    `json:"processed_at,omitempty" db:"processed_at"`
	CreatedAt       time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at" db:"updated_at"`
}

// ProcessPaymentRequest represents the request to process a payment
type ProcessPaymentRequest struct {
	OrderID       string        `json:"order_id" binding:"required"`
	PaymentMethod PaymentMethod `json:"payment_method" binding:"required"`
	Amount        float64       `json:"amount" binding:"required,min=0.01"`
	Currency      string        `json:"currency" binding:"required"`
}

// ProcessPaymentResponse represents the response after processing a payment
type ProcessPaymentResponse struct {
	ID            string        `json:"id"`
	OrderID       string        `json:"order_id"`
	Status        PaymentStatus `json:"status"`
	TransactionID string        `json:"transaction_id,omitempty"`
	ProcessedAt   *time.Time    `json:"processed_at,omitempty"`
}