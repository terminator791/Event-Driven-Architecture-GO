package service

import (
	"context"
	"fmt"

	"github.com/terminator791/Event-Driven-Architecture-GO/shared/pkg/models"
)

// mockProductService is a simple mock implementation for demo purposes
type mockProductService struct {
	products map[string]*models.Product
}

// NewMockProductService creates a new mock product service
func NewMockProductService() ProductService {
	products := map[string]*models.Product{
		"prod-1": {
			ID:          "prod-1",
			Name:        "Laptop",
			Description: "High-performance laptop",
			Price:       1299.99,
			SKU:         "LAP-001",
			Category:    "Electronics",
			InStock:     true,
			StockQty:    10,
		},
		"prod-2": {
			ID:          "prod-2",
			Name:        "Wireless Mouse",
			Description: "Ergonomic wireless mouse",
			Price:       29.99,
			SKU:         "MOU-001",
			Category:    "Electronics",
			InStock:     true,
			StockQty:    50,
		},
		"prod-3": {
			ID:          "prod-3",
			Name:        "Keyboard",
			Description: "Mechanical keyboard",
			Price:       89.99,
			SKU:         "KEY-001",
			Category:    "Electronics",
			InStock:     true,
			StockQty:    25,
		},
		"prod-4": {
			ID:          "prod-4",
			Name:        "Out of Stock Item",
			Description: "This item is out of stock",
			Price:       99.99,
			SKU:         "OOS-001",
			Category:    "Electronics",
			InStock:     false,
			StockQty:    0,
		},
	}

	return &mockProductService{products: products}
}

// GetByID retrieves a product by ID
func (s *mockProductService) GetByID(ctx context.Context, id string) (*models.Product, error) {
	product, exists := s.products[id]
	if !exists {
		return nil, fmt.Errorf("product not found")
	}

	// Return a copy to avoid mutation
	productCopy := *product
	return &productCopy, nil
}

// GetMultiple retrieves multiple products by IDs
func (s *mockProductService) GetMultiple(ctx context.Context, ids []string) ([]*models.Product, error) {
	var products []*models.Product

	for _, id := range ids {
		product, exists := s.products[id]
		if !exists {
			return nil, fmt.Errorf("product %s not found", id)
		}

		// Return a copy to avoid mutation
		productCopy := *product
		products = append(products, &productCopy)
	}

	return products, nil
}