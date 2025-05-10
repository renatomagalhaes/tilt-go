# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod file
COPY go.mod ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o api-server

# Final stage
FROM golang:1.21-alpine

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/api-server .

# Expose port
EXPOSE 8080

# Run the application
CMD ["./api-server"] 