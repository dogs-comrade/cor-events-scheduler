# Build stage
FROM golang:latest AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git build-base

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main cmd/app/main.go

# Final stage
FROM alpine:latest

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/main .
COPY --from=builder /app/config ./config

# Create non-root user
RUN adduser -D appuser
USER appuser

# Expose port
EXPOSE 8282

# Command to run
CMD ["./main"]