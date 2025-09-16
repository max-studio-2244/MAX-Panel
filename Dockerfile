# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o game-panel .

# Final stage
FROM alpine:latest

# Install Docker CLI and other dependencies
RUN apk --no-cache add docker-cli ca-certificates

WORKDIR /app

# Copy the binary
COPY --from=builder /app/game-panel .
COPY --from=builder /app/web ./web

# Create data directory
RUN mkdir -p /app/data

# Expose port
EXPOSE 8080

# Run the application
CMD ["./game-panel"]