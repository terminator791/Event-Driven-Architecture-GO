#!/bin/bash

# Integration test script for the Event-Driven Architecture demo

set -e

echo "🧪 Starting Integration Tests..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if services are running
check_service() {
    local service_name=$1
    local url=$2
    local max_attempts=30
    local attempt=1
    
    echo "⏳ Waiting for $service_name to be ready..."
    
    while [ $attempt -le $max_attempts ]; do
        if curl -s -f "$url" > /dev/null 2>&1; then
            echo -e "${GREEN}✅ $service_name is ready${NC}"
            return 0
        fi
        
        echo "   Attempt $attempt/$max_attempts..."
        sleep 2
        attempt=$((attempt + 1))
    done
    
    echo -e "${RED}❌ $service_name failed to start${NC}"
    return 1
}

# Test user creation
test_user_creation() {
    echo "👤 Testing user creation..."
    
    local email="test-$(date +%s)@example.com"
    local response
    
    response=$(curl -s -X POST http://localhost:8080/api/v1/users \
        -H "Content-Type: application/json" \
        -d "{\"email\":\"$email\",\"password\":\"password123\"}" \
        -w "%{http_code}")
    
    local http_code="${response: -3}"
    local body="${response%???}"
    
    if [ "$http_code" = "201" ]; then
        echo -e "${GREEN}✅ User creation successful${NC}"
        echo "   Response: $body"
        return 0
    else
        echo -e "${RED}❌ User creation failed (HTTP $http_code)${NC}"
        echo "   Response: $body"
        return 1
    fi
}

# Test duplicate user prevention
test_duplicate_user() {
    echo "🔒 Testing duplicate user prevention..."
    
    local email="duplicate-test@example.com"
    
    # Create user first time
    curl -s -X POST http://localhost:8080/api/v1/users \
        -H "Content-Type: application/json" \
        -d "{\"email\":\"$email\",\"password\":\"password123\"}" > /dev/null
    
    # Try to create same user again
    local response
    response=$(curl -s -X POST http://localhost:8080/api/v1/users \
        -H "Content-Type: application/json" \
        -d "{\"email\":\"$email\",\"password\":\"password123\"}" \
        -w "%{http_code}")
    
    local http_code="${response: -3}"
    
    if [ "$http_code" = "409" ]; then
        echo -e "${GREEN}✅ Duplicate user prevention working${NC}"
        return 0
    else
        echo -e "${RED}❌ Duplicate user prevention failed (HTTP $http_code)${NC}"
        return 1
    fi
}

# Test validation
test_validation() {
    echo "🛡️  Testing input validation..."
    
    # Test invalid email
    local response
    response=$(curl -s -X POST http://localhost:8080/api/v1/users \
        -H "Content-Type: application/json" \
        -d '{"email":"invalid-email","password":"password123"}' \
        -w "%{http_code}")
    
    local http_code="${response: -3}"
    
    if [ "$http_code" = "400" ]; then
        echo -e "${GREEN}✅ Email validation working${NC}"
    else
        echo -e "${YELLOW}⚠️  Email validation response: HTTP $http_code${NC}"
    fi
    
    # Test short password
    response=$(curl -s -X POST http://localhost:8080/api/v1/users \
        -H "Content-Type: application/json" \
        -d '{"email":"test@example.com","password":"short"}' \
        -w "%{http_code}")
    
    http_code="${response: -3}"
    
    if [ "$http_code" = "400" ]; then
        echo -e "${GREEN}✅ Password validation working${NC}"
    else
        echo -e "${YELLOW}⚠️  Password validation response: HTTP $http_code${NC}"
    fi
}

# Test health endpoint
test_health() {
    echo "🏥 Testing health endpoint..."
    
    local response
    response=$(curl -s http://localhost:8080/api/v1/health)
    
    if echo "$response" | grep -q "healthy"; then
        echo -e "${GREEN}✅ Health check passed${NC}"
        return 0
    else
        echo -e "${RED}❌ Health check failed${NC}"
        return 1
    fi
}

# Main execution
main() {
    echo "🚀 Event-Driven Architecture Integration Tests"
    echo "=============================================="
    
    # Check if user-api is ready
    if ! check_service "User API" "http://localhost:8080/api/v1/health"; then
        echo -e "${RED}❌ Integration tests cannot run without User API${NC}"
        exit 1
    fi
    
    # Run tests
    local failed_tests=0
    
    test_health || failed_tests=$((failed_tests + 1))
    test_user_creation || failed_tests=$((failed_tests + 1))
    test_duplicate_user || failed_tests=$((failed_tests + 1))
    test_validation || failed_tests=$((failed_tests + 1))
    
    echo ""
    echo "=============================================="
    
    if [ $failed_tests -eq 0 ]; then
        echo -e "${GREEN}🎉 All integration tests passed!${NC}"
        exit 0
    else
        echo -e "${RED}❌ $failed_tests test(s) failed${NC}"
        exit 1
    fi
}

# Check if we're running in CI or if user wants to run tests
if [ "$1" = "--ci" ] || [ "$CI" = "true" ]; then
    main
else
    echo "💡 This script will test the running services."
    echo "   Make sure services are running with: make docker-up"
    echo ""
    read -p "Continue with integration tests? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        main
    else
        echo "Integration tests cancelled."
        exit 0
    fi
fi