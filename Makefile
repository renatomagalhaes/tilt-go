.PHONY: tilt-up tilt-down clean build test install-tilt help check-tilt test-shutdown-api test-shutdown-worker logs-services restart-services

# Check if Tilt is installed
check-tilt:
	@if ! command -v tilt >/dev/null 2>&1; then \
		echo "Tilt não está instalado. Instalando..."; \
		curl -fsSL https://raw.githubusercontent.com/tilt-dev/tilt/master/scripts/install.sh | bash; \
	else \
		echo "Tilt já está instalado."; \
	fi

# Install Tilt
install-tilt: check-tilt

# Start development environment
tilt-up: install-tilt
	@echo "Iniciando ambiente de desenvolvimento..."
	@tilt up

# Stop development environment
tilt-down:
	@echo "Parando ambiente de desenvolvimento..."
	@tilt down

# Build the application
build:
	@echo "Construindo a aplicação..."
	@cd api && go build -o api-server
	@cd worker && go build -o worker-server

# Run tests
test:
	@echo "Executando testes..."
	@cd api && go test ./...
	@cd worker && go test ./...
	@echo "Tests completed!"

# Test graceful shutdown
test-shutdown-api:
	@echo "Testing API graceful shutdown..."
	@kubectl get pod -l app=api-server -o name | xargs -I {} kubectl delete {} --grace-period=30

test-shutdown-worker:
	@echo "Testing Worker graceful shutdown..."
	@kubectl get pod -l app=worker-server -o name | xargs -I {} kubectl delete {} --grace-period=30

# Development services
logs-services:
	@echo "Mostrando logs dos serviços..."
	@docker-compose -f docker-compose.dev.yml logs -f

restart-services:
	@echo "Reiniciando serviços..."
	@docker-compose -f docker-compose.dev.yml restart

# Clean up
clean:
	@echo "Limpando ambiente..."
	@docker-compose -f docker-compose.dev.yml down -v
	@tilt down

# Help command
help:
	@echo "Comandos disponíveis:"
	@echo "  make tilt-up          - Inicia o ambiente de desenvolvimento"
	@echo "  make tilt-down        - Para o ambiente de desenvolvimento"
	@echo "  make build           - Compila a aplicação"
	@echo "  make test            - Executa os testes"
	@echo "  make clean           - Limpa o ambiente"
	@echo "  make logs-services   - Mostra logs dos serviços"
	@echo "  make restart-services - Reinicia os serviços"
	@echo "  make help            - Mostra esta mensagem" 