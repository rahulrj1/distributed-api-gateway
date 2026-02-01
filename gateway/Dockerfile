# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod ./

# Download dependencies (none for now, but good practice)
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o gateway main.go

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Add ca-certificates for HTTPS calls (future use)
RUN apk --no-cache add ca-certificates

# Copy binary from builder
COPY --from=builder /app/gateway .

# Expose port (default gateway port per LLD)
EXPOSE 5000

# Set environment variables (per LLD specification)
ENV SERVER_HOST=0.0.0.0
ENV SERVER_PORT=5000

# Run the gateway
CMD ["./gateway"]
