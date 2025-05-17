# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go mod file
COPY go.mod go.sum ./ 

# Copy source code
COPY . . 

# Download dependencies and generate vendor directory
RUN go mod tidy
RUN go mod vendor

# Build the application using the vendor directory
RUN go build -mod=vendor -o worker-server

# Final stage
FROM alpine:latest

WORKDIR /app

# Install tzdata package
RUN apk add --no-cache tzdata

# Copy the binary from builder
COPY --from=builder /app/worker-server .

# Expose port
EXPOSE 8081

# Run the application
ENTRYPOINT ["/app/worker-server"] 