FROM golang:1.25.1 as builder

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /redis-proxy ./cmd/proxy

# Create final image
FROM alpine:3.14

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Copy binary from builder
COPY --from=builder /redis-proxy /usr/local/bin/

# Create config directory
RUN mkdir -p /etc/redis-proxy

# Expose the proxy port
EXPOSE 6380

# Run the proxy
ENTRYPOINT ["/usr/local/bin/redis-proxy", "--config", "/etc/redis-proxy/config.yaml"]
