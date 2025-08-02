# Use the official Go image as a build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Download dependencies first to leverage caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .
COPY .env .env

# Build the application binary
RUN go build -o authorization-service ./cmd

# Final runtime image
FROM alpine:3.18

WORKDIR /app

# Copy executable and configuration
COPY --from=builder /app/authorization-service ./authorization-service
COPY configs ./configs

# Default port for the service
ENV PORT=8080
EXPOSE 8080

# Run the service
CMD ["./authorization-service"]
