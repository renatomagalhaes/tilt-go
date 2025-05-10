.PHONY: tilt-up tilt-down clean build test install-tilt help check-tilt all run load-test metrics-server-up metrics-server-down

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
	rm -f worker/worker
	rm -rf api/vendor
	rm -rf worker/vendor
	@echo "Clean completed!"

# Build both services
build:
	@echo "Building services..."
	cd api && go mod tidy && go build -o api-server
	cd worker && go mod tidy && go build -o worker
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

all: build

run:
	@echo "Running API..."
	cd api && go run main.go
	@echo "Running Worker..."
	cd worker && go run main.go

observability-up:
	@echo "Starting observability stack..."
	kubectl apply -f k8s/observability/prometheus.yaml
	kubectl apply -f k8s/observability/grafana.yaml
	kubectl apply -f k8s/observability/elasticsearch.yaml
	kubectl apply -f k8s/observability/kibana.yaml
	kubectl apply -f k8s/observability/jaeger.yaml
	@echo "Waiting for services to be ready..."
	kubectl wait --for=condition=available --timeout=300s deployment/prometheus
	kubectl wait --for=condition=available --timeout=300s deployment/grafana
	kubectl wait --for=condition=available --timeout=300s statefulset/elasticsearch
	kubectl wait --for=condition=available --timeout=300s deployment/kibana
	kubectl wait --for=condition=available --timeout=300s deployment/jaeger
	@echo "Observability stack is ready!"
	@echo "Prometheus: http://localhost:9090"
	@echo "Grafana: http://localhost:3000 (admin/admin)"
	@echo "Kibana: http://localhost:5601"
	@echo "Jaeger: http://localhost:16686"

observability-down:
	@echo "Stopping observability stack..."
	kubectl delete -f k8s/observability/jaeger.yaml
	kubectl delete -f k8s/observability/kibana.yaml
	kubectl delete -f k8s/observability/elasticsearch.yaml
	kubectl delete -f k8s/observability/grafana.yaml
	kubectl delete -f k8s/observability/prometheus.yaml

# Check if hey is installed
check-hey:
	@if ! command -v hey &> /dev/null; then \
		echo "hey is not installed. Installing now..."; \
		go install github.com/rakyll/hey@latest; \
		echo "hey installed successfully!"; \
	else \
		echo "hey is already installed."; \
	fi

load-test: check-hey
	@echo "Iniciando teste de carga..."
	@bash scripts/load-test.sh

metrics-server-up:
	@echo "Installing Metrics Server..."
	kubectl apply -f k8s/observability/metrics-server-auth.yaml
	kubectl apply -f k8s/observability/metrics-server.yaml
	@echo "Waiting for Metrics Server to be ready..."
	kubectl wait --for=condition=available --timeout=300s deployment/metrics-server -n kube-system
	@echo "Metrics Server is ready!"

metrics-server-down:
	@echo "Removing Metrics Server..."
	kubectl delete -f k8s/observability/metrics-server.yaml
	kubectl delete -f k8s/observability/metrics-server-auth.yaml 