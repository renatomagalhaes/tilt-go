# Build stage
FROM golang:1.22-alpine

WORKDIR /app

# Copy go mod file
COPY go.mod ./
COPY go.sum ./

# Download dependencies and generate go.sum
RUN go mod download && go mod tidy

# Copy source code
COPY . .

# Build the application
RUN go build -o worker-server .

# Final stage
FROM alpine:latest

WORKDIR /app

# Install tzdata package
RUN apk add --no-cache tzdata

# Copy the binary from builder
COPY --from=0 /app/worker-server .

# Expose port
EXPOSE 8081

# Run the application
CMD ["./worker-server"] 