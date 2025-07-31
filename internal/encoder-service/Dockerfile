# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the encoder service
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o encoder-service ./internal/encoder-service

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates ffmpeg postgresql-client

# Create app directory
WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/encoder-service .

# Copy migrations
COPY --from=builder /app/internal/encoder-service/migrations ./migrations

# Make migration script executable
RUN chmod +x ./migrations/run_migrations.sh

# Create output directory
RUN mkdir -p /tmp/hls

# Expose port
EXPOSE 8082

# Run migrations and start service
CMD ["./migrations/run_migrations.sh"] 