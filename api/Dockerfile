# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go mod file
COPY go.mod ./go.mod
COPY go.sum ./go.sum

# Download dependencies
RUN go mod download && go mod tidy

# Copy source code
COPY . .

# Build the application
RUN go build -o api-server

# Final stage
FROM alpine:latest

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/api-server .

# Expose port
EXPOSE 8080

# Run the application
ENTRYPOINT ["/app/api-server"] 