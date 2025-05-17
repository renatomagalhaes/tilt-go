.PHONY: tilt-up tilt-down clean build test install-tilt help check-tilt test-shutdown-api test-shutdown-worker

# Check if Tilt is installed
check-tilt:
	@if ! which tilt >/dev/null 2>&1 || ! tilt version >/dev/null 2>&1; then \
		echo "Tilt is not installed or not working properly. Installing now..."; \
		curl -fsSL https://raw.githubusercontent.com/tilt-dev/tilt/master/scripts/install.sh | bash; \
		echo "Tilt installed successfully!"; \
	else \
		echo "Tilt is already installed and working."; \
	fi

# Install Tilt.dev
install-tilt:
	@echo "Installing Tilt.dev..."
	curl -fsSL https://raw.githubusercontent.com/tilt-dev/tilt/master/scripts/install.sh | bash
	@echo "Tilt.dev installed successfully!"

# Start Tilt development environment
tilt-up: check-tilt
	@echo "Starting Tilt development environment..."
	tilt up

# Stop Tilt development environment
tilt-down: check-tilt
	@echo "Stopping Tilt development environment..."
	tilt down

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

# Test graceful shutdown
test-shutdown-api:
	@echo "Testing API graceful shutdown..."
	@kubectl get pod -l app=api-server -o name | xargs -I {} kubectl delete {} --grace-period=30

test-shutdown-worker:
	@echo "Testing Worker graceful shutdown..."
	@kubectl get pod -l app=worker-server -o name | xargs -I {} kubectl delete {} --grace-period=30

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