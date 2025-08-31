.PHONY: build test run clean docker-build docker-up docker-down help

# Default target
help:
	@echo "Available targets:"
	@echo "  build           - Build all services"
	@echo "  test            - Run all tests"
	@echo "  test-unit       - Run unit tests only"
	@echo "  test-integration- Run integration tests"
	@echo "  run-user-api    - Run user-api locally"
	@echo "  run-order-service- Run order-service locally"
	@echo "  run-payment-service- Run payment-service locally"
	@echo "  run-emailer     - Run emailer-service locally"
	@echo "  clean           - Clean build artifacts"
	@echo "  docker-build    - Build Docker images"
	@echo "  docker-up       - Start all services with Docker Compose"
	@echo "  docker-down     - Stop all services"
	@echo "  docker-logs     - Show logs from all services"
	@echo "  create-user     - Create a test user (requires services to be running)"
	@echo "  create-order    - Create a test order (requires services to be running)"
	@echo "  process-payment - Process a test payment (requires services to be running)"
	@echo "  health-check    - Check all services health"

# Go build settings
GO_VERSION := 1.25
GOFLAGS := -mod=readonly

# Build all services
build:
	@echo "🔨 Building user-api..."
	@cd user-api && go build -o bin/user-api ./cmd/main.go
	@echo "🔨 Building emailer-service..."
	@cd emailer-service && go build -o bin/emailer-service ./cmd/main.go
	@echo "🔨 Building order-service..."
	@cd order-service && go build -o bin/order-service ./cmd/main.go
	@echo "🔨 Building payment-service..."
	@cd payment-service && go build -o bin/payment-service ./cmd/main.go
	@echo "✅ Build complete"

# Run tests
test:
	@echo "🧪 Running all tests..."
	@go test ./... -v -race -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Tests complete. Coverage report: coverage.html"

# Run unit tests only
test-unit:
	@echo "🧪 Running unit tests..."
	@go test ./user-api/internal/service -v
	@go test ./emailer-service/internal/service -v
	@echo "✅ Unit tests complete"

# Run integration tests  
test-integration:
	@echo "🧪 Running integration tests..."
	@./scripts/integration-test.sh --ci
	@echo "✅ Integration tests complete"

# Run user-api locally (requires PostgreSQL and Kafka)
run-user-api:
	@echo "🚀 Starting user-api..."
	@cd user-api && go run ./cmd/main.go

# Run order-service locally (requires PostgreSQL and Kafka)
run-order-service:
	@echo "🚀 Starting order-service..."
	@cd order-service && go run ./cmd/main.go

# Run payment-service locally (requires PostgreSQL and Kafka)
run-payment-service:
	@echo "🚀 Starting payment-service..."
	@cd payment-service && go run ./cmd/main.go

# Run emailer-service locally (requires Kafka)
run-emailer:
	@echo "🚀 Starting emailer-service..."
	@cd emailer-service && go run ./cmd/main.go

# Clean build artifacts
clean:
	@echo "🧹 Cleaning..."
	@rm -rf user-api/bin
	@rm -rf emailer-service/bin
	@rm -rf order-service/bin
	@rm -rf payment-service/bin
	@rm -f coverage.out coverage.html
	@echo "✅ Clean complete"

# Docker targets
docker-build:
	@echo "🐳 Building Docker images..."
	@docker-compose build --no-cache
	@echo "✅ Docker build complete"

docker-up:
	@echo "🐳 Starting services with Docker Compose..."
	@docker-compose up -d
	@echo "✅ Services started. Use 'make docker-logs' to see logs"

docker-down:
	@echo "🐳 Stopping services..."
	@docker-compose down
	@echo "✅ Services stopped"

docker-logs:
	@echo "📋 Showing logs..."
	@docker-compose logs -f

# Development helpers
create-user:
	@echo "👤 Creating test user..."
	@curl -X POST http://localhost:8080/api/v1/users \
		-H "Content-Type: application/json" \
		-d '{"email":"test@example.com","password":"password123"}' \
		| jq '.' || echo "Error: jq not installed or request failed"

create-order:
	@echo "🛒 Creating test order..."
	@curl -X POST http://localhost:8081/api/v1/orders \
		-H "Content-Type: application/json" \
		-d '{"user_id":"test-user","items":[{"product_id":"prod-1","quantity":1}],"currency":"USD","payment_method":"credit_card","shipping_address":{"street":"123 Main St","city":"Anytown","state":"NY","postal_code":"12345","country":"USA"},"billing_address":{"street":"123 Main St","city":"Anytown","state":"NY","postal_code":"12345","country":"USA"}}' \
		| jq '.' || echo "Error: jq not installed or request failed"

process-payment:
	@echo "💳 Processing test payment..."
	@curl -X POST http://localhost:8082/api/v1/payments \
		-H "Content-Type: application/json" \
		-d '{"order_id":"test-order","payment_method":"credit_card","amount":1299.99,"currency":"USD"}' \
		| jq '.' || echo "Error: jq not installed or request failed"

health-check:
	@echo "🏥 Checking service health..."
	@echo "User API:"
	@curl -s http://localhost:8080/api/v1/health | jq '.' || echo "User API not responding"
	@echo "Order Service:"
	@curl -s http://localhost:8081/api/v1/health | jq '.' || echo "Order Service not responding"
	@echo "Payment Service:"
	@curl -s http://localhost:8082/api/v1/health | jq '.' || echo "Payment Service not responding"

# Format code
fmt:
	@echo "📝 Formatting code..."
	@go fmt ./...
	@echo "✅ Code formatted"

# Lint code (requires golangci-lint)
lint:
	@echo "🔍 Linting code..."
	@golangci-lint run ./...
	@echo "✅ Linting complete"

# Install dependencies
deps:
	@echo "📦 Installing dependencies..."
	@go mod download
	@go mod tidy
	@echo "✅ Dependencies installed"

# Setup development environment
setup: deps
	@echo "🛠️  Setting up development environment..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "✅ Development environment ready"