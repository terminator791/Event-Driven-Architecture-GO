.PHONY: build test run clean docker-build docker-up docker-down help

# Default target
help:
	@echo "Available targets:"
	@echo "  build           - Build all services"
	@echo "  test            - Run all tests"
	@echo "  test-unit       - Run unit tests only"
	@echo "  test-integration- Run integration tests"
	@echo "  run-user-api    - Run user-api locally"
	@echo "  run-emailer     - Run emailer-service locally"
	@echo "  clean           - Clean build artifacts"
	@echo "  docker-build    - Build Docker images"
	@echo "  docker-up       - Start all services with Docker Compose"
	@echo "  docker-down     - Stop all services"
	@echo "  docker-logs     - Show logs from all services"
	@echo "  create-user     - Create a test user (requires services to be running)"

# Go build settings
GO_VERSION := 1.25
GOFLAGS := -mod=readonly

# Build all services
build:
	@echo "🔨 Building user-api..."
	@cd user-api && go build -o bin/user-api ./cmd/main.go
	@echo "🔨 Building emailer-service..."
	@cd emailer-service && go build -o bin/emailer-service ./cmd/main.go
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

# Run emailer-service locally (requires Kafka)
run-emailer:
	@echo "🚀 Starting emailer-service..."
	@cd emailer-service && go run ./cmd/main.go

# Clean build artifacts
clean:
	@echo "🧹 Cleaning..."
	@rm -rf user-api/bin
	@rm -rf emailer-service/bin
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

health-check:
	@echo "🏥 Checking service health..."
	@curl -s http://localhost:8080/api/v1/health | jq '.' || echo "Error: User API not responding"

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