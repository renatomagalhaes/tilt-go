.PHONY: tilt-up tilt-down clean build test install-tilt help check-tilt

# Check if Tilt is installed
check-tilt:
	@if ! command -v tilt &> /dev/null; then \
		echo "Tilt is not installed. Installing now..."; \
		curl -fsSL https://raw.githubusercontent.com/tilt-dev/tilt/master/scripts/install.sh | bash; \
		echo "Tilt installed successfully!"; \
	else \
		echo "Tilt is already installed."; \
	fi

# Install Tilt.dev
install-tilt:
	@echo "Installing Tilt.dev..."
	curl -fsSL https://raw.githubusercontent.com/tilt-dev/tilt/master/scripts/install.sh | bash
	@echo "Tilt.dev installed successfully!"

# Start Tilt development environment
tilt-up: check-tilt
	@echo "Starting Tilt development environment..."
	@if command -v tilt &> /dev/null; then \
		tilt up; \
	else \
		echo "Error: Tilt command not found. Please run 'make install-tilt' first."; \
		exit 1; \
	fi

# Stop Tilt development environment
tilt-down: check-tilt
	@echo "Stopping Tilt development environment..."
	@if command -v tilt &> /dev/null; then \
		tilt down; \
	else \
		echo "Error: Tilt command not found. Please run 'make install-tilt' first."; \
		exit 1; \
	fi

# Clean build artifacts and temporary files
clean:
	@echo "Cleaning build artifacts..."
	rm -f api/api-server
	rm -f worker/worker-server
	rm -rf api/vendor
	rm -rf worker/vendor
	@echo "Clean completed!"

# Build both services
build:
	@echo "Building services..."
	cd api && go mod tidy && go build -o api-server
	cd worker && go mod tidy && go build -o worker-server
	@echo "Build completed!"

# Run tests
test:
	@echo "Running tests..."
	cd api && go test -v ./...
	cd worker && go test -v ./...
	@echo "Tests completed!"

# Help command
help:
	@echo "Available commands:"
	@echo "  make install-tilt - Install Tilt.dev"
	@echo "  make tilt-up     - Start Tilt development environment"
	@echo "  make tilt-down   - Stop Tilt development environment"
	@echo "  make clean       - Clean build artifacts and temporary files"
	@echo "  make build       - Build both services"
	@echo "  make test        - Run tests"
	@echo "  make help        - Show this help message" 