package database

import (
	"github.com/jmoiron/sqlx"
)

// CreateOrderTables creates the order-related tables
func CreateOrderTables(db *sqlx.DB) error {
	// Create orders table
	ordersTable := `
	CREATE TABLE IF NOT EXISTS orders (
		id VARCHAR(255) PRIMARY KEY,
		user_id VARCHAR(255) NOT NULL,
		status VARCHAR(50) NOT NULL,
		total_amount DECIMAL(10,2) NOT NULL,
		currency VARCHAR(3) NOT NULL,
		payment_status VARCHAR(50) NOT NULL,
		payment_method VARCHAR(50) NOT NULL,
		shipping_address JSONB NOT NULL,
		billing_address JSONB NOT NULL,
		notes TEXT,
		created_at TIMESTAMP WITH TIME ZONE NOT NULL,
		updated_at TIMESTAMP WITH TIME ZONE NOT NULL
	);`

	if _, err := db.Exec(ordersTable); err != nil {
		return err
	}

	// Create order_items table
	orderItemsTable := `
	CREATE TABLE IF NOT EXISTS order_items (
		id VARCHAR(255) PRIMARY KEY,
		order_id VARCHAR(255) NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
		product_id VARCHAR(255) NOT NULL,
		quantity INTEGER NOT NULL,
		unit_price DECIMAL(10,2) NOT NULL,
		total_price DECIMAL(10,2) NOT NULL
	);`

	if _, err := db.Exec(orderItemsTable); err != nil {
		return err
	}

	// Create indexes
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);",
		"CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);",
		"CREATE INDEX IF NOT EXISTS idx_orders_created_at ON orders(created_at);",
		"CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items(order_id);",
		"CREATE INDEX IF NOT EXISTS idx_order_items_product_id ON order_items(product_id);",
	}

	for _, index := range indexes {
		if _, err := db.Exec(index); err != nil {
			return err
		}
	}

	return nil
}