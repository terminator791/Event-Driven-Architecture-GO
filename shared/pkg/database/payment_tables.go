package database

import (
	"github.com/jmoiron/sqlx"
)

// CreatePaymentTables creates the payment-related tables
func CreatePaymentTables(db *sqlx.DB) error {
	// Create payments table
	paymentsTable := `
	CREATE TABLE IF NOT EXISTS payments (
		id VARCHAR(255) PRIMARY KEY,
		order_id VARCHAR(255) NOT NULL UNIQUE,
		user_id VARCHAR(255),
		amount DECIMAL(10,2) NOT NULL,
		currency VARCHAR(3) NOT NULL,
		payment_method VARCHAR(50) NOT NULL,
		status VARCHAR(50) NOT NULL,
		transaction_id VARCHAR(255),
		failure_reason TEXT,
		retry_count INTEGER DEFAULT 0,
		processed_at TIMESTAMP WITH TIME ZONE,
		created_at TIMESTAMP WITH TIME ZONE NOT NULL,
		updated_at TIMESTAMP WITH TIME ZONE NOT NULL
	);`

	if _, err := db.Exec(paymentsTable); err != nil {
		return err
	}

	// Create indexes
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_payments_order_id ON payments(order_id);",
		"CREATE INDEX IF NOT EXISTS idx_payments_user_id ON payments(user_id);",
		"CREATE INDEX IF NOT EXISTS idx_payments_status ON payments(status);",
		"CREATE INDEX IF NOT EXISTS idx_payments_created_at ON payments(created_at);",
		"CREATE INDEX IF NOT EXISTS idx_payments_transaction_id ON payments(transaction_id);",
	}

	for _, index := range indexes {
		if _, err := db.Exec(index); err != nil {
			return err
		}
	}

	return nil
}