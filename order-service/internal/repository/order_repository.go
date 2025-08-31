package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/terminator791/Event-Driven-Architecture-GO/shared/pkg/models"
)

// OrderRepository defines the interface for order data operations
type OrderRepository interface {
	Create(ctx context.Context, order *models.Order) error
	GetByID(ctx context.Context, id string) (*models.Order, error)
	GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.Order, error)
	UpdateStatus(ctx context.Context, id string, status models.OrderStatus, notes string) error
	UpdatePaymentStatus(ctx context.Context, id string, paymentStatus models.PaymentStatus) error
	Delete(ctx context.Context, id string) error
}

// orderRepository implements OrderRepository
type orderRepository struct {
	db *sqlx.DB
}

// NewOrderRepository creates a new order repository
func NewOrderRepository(db *sqlx.DB) OrderRepository {
	return &orderRepository{db: db}
}

// Create creates a new order
func (r *orderRepository) Create(ctx context.Context, order *models.Order) error {
	// Generate ID if not provided
	if order.ID == "" {
		order.ID = uuid.New().String()
	}

	// Convert addresses to JSON
	shippingAddrJSON, err := json.Marshal(order.ShippingAddress)
	if err != nil {
		return fmt.Errorf("failed to marshal shipping address: %w", err)
	}

	billingAddrJSON, err := json.Marshal(order.BillingAddress)
	if err != nil {
		return fmt.Errorf("failed to marshal billing address: %w", err)
	}

	query := `
		INSERT INTO orders (id, user_id, status, total_amount, currency, payment_status, 
						   payment_method, shipping_address, billing_address, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	_, err = r.db.ExecContext(ctx, query,
		order.ID, order.UserID, order.Status, order.TotalAmount, order.Currency,
		order.PaymentStatus, order.PaymentMethod, shippingAddrJSON, billingAddrJSON,
		order.Notes, order.CreatedAt, order.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}

	// Create order items
	for i := range order.Items {
		item := &order.Items[i]
		if item.ID == "" {
			item.ID = uuid.New().String()
		}
		item.OrderID = order.ID

		itemQuery := `
			INSERT INTO order_items (id, order_id, product_id, quantity, unit_price, total_price)
			VALUES ($1, $2, $3, $4, $5, $6)`

		_, err = r.db.ExecContext(ctx, itemQuery,
			item.ID, item.OrderID, item.ProductID, item.Quantity, item.UnitPrice, item.TotalPrice)

		if err != nil {
			return fmt.Errorf("failed to create order item: %w", err)
		}
	}

	return nil
}

// GetByID retrieves an order by ID
func (r *orderRepository) GetByID(ctx context.Context, id string) (*models.Order, error) {
	var order models.Order
	var shippingAddrJSON, billingAddrJSON []byte

	query := `
		SELECT id, user_id, status, total_amount, currency, payment_status, payment_method,
			   shipping_address, billing_address, notes, created_at, updated_at
		FROM orders WHERE id = $1`

	row := r.db.QueryRowContext(ctx, query, id)
	err := row.Scan(
		&order.ID, &order.UserID, &order.Status, &order.TotalAmount, &order.Currency,
		&order.PaymentStatus, &order.PaymentMethod, &shippingAddrJSON, &billingAddrJSON,
		&order.Notes, &order.CreatedAt, &order.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	// Unmarshal addresses
	if err := json.Unmarshal(shippingAddrJSON, &order.ShippingAddress); err != nil {
		return nil, fmt.Errorf("failed to unmarshal shipping address: %w", err)
	}

	if err := json.Unmarshal(billingAddrJSON, &order.BillingAddress); err != nil {
		return nil, fmt.Errorf("failed to unmarshal billing address: %w", err)
	}

	// Get order items
	itemsQuery := `
		SELECT id, order_id, product_id, quantity, unit_price, total_price
		FROM order_items WHERE order_id = $1`

	err = r.db.SelectContext(ctx, &order.Items, itemsQuery, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get order items: %w", err)
	}

	return &order, nil
}

// GetByUserID retrieves orders by user ID with pagination
func (r *orderRepository) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.Order, error) {
	query := `
		SELECT id, user_id, status, total_amount, currency, payment_status, payment_method,
			   shipping_address, billing_address, notes, created_at, updated_at
		FROM orders 
		WHERE user_id = $1 
		ORDER BY created_at DESC 
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}
	defer rows.Close()

	var orders []*models.Order
	for rows.Next() {
		var order models.Order
		var shippingAddrJSON, billingAddrJSON []byte

		err := rows.Scan(
			&order.ID, &order.UserID, &order.Status, &order.TotalAmount, &order.Currency,
			&order.PaymentStatus, &order.PaymentMethod, &shippingAddrJSON, &billingAddrJSON,
			&order.Notes, &order.CreatedAt, &order.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}

		// Unmarshal addresses
		if err := json.Unmarshal(shippingAddrJSON, &order.ShippingAddress); err != nil {
			return nil, fmt.Errorf("failed to unmarshal shipping address: %w", err)
		}

		if err := json.Unmarshal(billingAddrJSON, &order.BillingAddress); err != nil {
			return nil, fmt.Errorf("failed to unmarshal billing address: %w", err)
		}

		orders = append(orders, &order)
	}

	return orders, nil
}

// UpdateStatus updates an order's status
func (r *orderRepository) UpdateStatus(ctx context.Context, id string, status models.OrderStatus, notes string) error {
	query := `UPDATE orders SET status = $1, notes = $2, updated_at = NOW() WHERE id = $3`
	
	_, err := r.db.ExecContext(ctx, query, status, notes, id)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	return nil
}

// UpdatePaymentStatus updates an order's payment status
func (r *orderRepository) UpdatePaymentStatus(ctx context.Context, id string, paymentStatus models.PaymentStatus) error {
	query := `UPDATE orders SET payment_status = $1, updated_at = NOW() WHERE id = $2`
	
	_, err := r.db.ExecContext(ctx, query, paymentStatus, id)
	if err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	return nil
}

// Delete deletes an order (soft delete by updating status)
func (r *orderRepository) Delete(ctx context.Context, id string) error {
	query := `UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2`
	
	_, err := r.db.ExecContext(ctx, query, models.OrderStatusCancelled, id)
	if err != nil {
		return fmt.Errorf("failed to delete order: %w", err)
	}

	return nil
}