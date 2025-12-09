# syntax=docker/dockerfile:1
# ============================
# 1️⃣ Build Stage (arm64)
# ============================
FROM --platform=linux/amd64 golang:1.25-alpine AS builder

WORKDIR /app

# Install git for Go modules
RUN apk add --no-cache git

# Copy modules first to leverage cache
COPY go.mod go.sum ./
RUN go env -w GOPROXY=https://proxy.golang.org,direct && go mod download

# Copy entire project
COPY . .

# Build static amd64 binary
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN go build -ldflags="-s -w" -o /app/main cmd/api/main.go

# ============================
# 2️⃣ Run Stage (amd64)
# ============================
FROM --platform=linux/amd64 alpine:latest

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/main .

# Optional: copy .env if you need it in image
COPY .env .env

EXPOSE 8080

# Use exec form so signals are delivered to your process
ENTRYPOINT ["/app/main"]
