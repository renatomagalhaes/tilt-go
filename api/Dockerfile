# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod file
COPY go.mod ./

# Download dependencies and generate go.sum
RUN go mod download && go mod tidy

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Final stage
FROM golang:1.21-alpine

WORKDIR /app

# Copy the binary and go.mod from builder
COPY --from=builder /app/main .
COPY --from=builder /app/go.mod .
COPY --from=builder /app/go.sum .

# Expose port
EXPOSE 8080

# Run the application
CMD ["./main"] 