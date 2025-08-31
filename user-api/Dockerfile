# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the user-api binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o user-api ./user-api/cmd/main.go

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS calls
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/user-api .

# Expose port
EXPOSE 8080

# Run the binary
CMD ["./user-api"]