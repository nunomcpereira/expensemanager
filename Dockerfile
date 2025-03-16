FROM golang:1.24-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/server

# Start a new stage from scratch
FROM alpine:latest

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/main .
COPY --from=builder /app/cmd/server/templates ./cmd/server/templates
COPY --from=builder /app/cmd/server/static ./cmd/server/static
COPY --from=builder /app/internal/i18n/locales ./internal/i18n/locales

# Expose port
EXPOSE 8080

# Command to run
CMD ["./main"] 