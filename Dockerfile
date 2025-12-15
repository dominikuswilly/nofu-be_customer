# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with stripped binary
RUN go build -o main -ldflags="-s -w" cmd/api/main.go

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create a non-root user
RUN adduser -D -s /bin/sh appuser

WORKDIR /home/appuser

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Give execute permission
RUN chmod +x main

# Switch to non-root user
USER appuser

# Health check (add this if you have a /health endpoint)
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost:8080/health || exit 1

# Expose port 8080
EXPOSE 8080

# Command to run the application
CMD ["./main"]
