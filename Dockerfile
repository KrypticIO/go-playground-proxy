# Use the official Go image
FROM golang:1.24-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o proxy .

# Use a minimal image for the final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create a non-root user and group
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Create app directory
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/proxy .

# Set proper permissions
RUN chmod +x /app/proxy && \
    chown -R appuser:appgroup /app

# Use non-root user
USER appuser

# Expose port 8080
EXPOSE 8080

# Run the proxy
CMD ["./proxy"]