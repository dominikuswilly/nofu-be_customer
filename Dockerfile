# ============================
# 1️⃣ Build Stage
# ============================
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install git for Go modules
RUN apk add --no-cache git

# Copy modules first
COPY go.mod go.sum ./
RUN go mod download

# Copy entire project
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o main cmd/api/main.go


# ============================
# 2️⃣ Run Stage
# ============================
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/main .

# Copy .env if needed
COPY .env .env

EXPOSE 8080

CMD ["./main"]
