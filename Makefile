.PHONY: init tilt-up tilt-down clean build test

# Initialize the project
init:
	@echo "Initializing Git repository..."
	git init
	git add .
	git commit -m "Initial commit"
	@echo "Repository initialized successfully!"

# Start Tilt development environment
tilt-up:
	@echo "Starting Tilt development environment..."
	tilt up

# Stop Tilt development environment
tilt-down:
	@echo "Stopping Tilt development environment..."
	tilt down

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f api/api-server
	rm -f worker/worker
	@echo "Clean completed!"

# Build both services
build:
	@echo "Building services..."
	cd api && go build -o api-server
	cd worker && go build -o worker
	@echo "Build completed!"

# Run tests
test:
	@echo "Running tests..."
	cd api && go test ./...
	cd worker && go test ./...
	@echo "Tests completed!"

# Help command
help:
	@echo "Available commands:"
	@echo "  make init      - Initialize Git repository"
	@echo "  make tilt-up   - Start Tilt development environment"
	@echo "  make tilt-down - Stop Tilt development environment"
	@echo "  make clean     - Clean build artifacts"
	@echo "  make build     - Build both services"
	@echo "  make test      - Run tests"
	@echo "  make help      - Show this help message" 