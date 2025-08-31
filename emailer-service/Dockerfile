# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the emailer-service binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o emailer-service ./emailer-service/cmd/main.go

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS calls
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/emailer-service .

# Run the binary
CMD ["./emailer-service"]