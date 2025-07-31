FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the encoder service
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o encoder ./internal/encoder-service

# Final stage
FROM alpine:latest

# Install FFmpeg and other runtime dependencies
RUN apk add --no-cache \
    ffmpeg \
    ca-certificates \
    && rm -rf /var/cache/apk/*

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/encoder .

# Create HLS output directory
RUN mkdir -p /tmp/hls

CMD ["./encoder"] 