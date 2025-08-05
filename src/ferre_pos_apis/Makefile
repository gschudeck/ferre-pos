# Makefile para APIs FERRE-POS

.PHONY: help build test test-unit test-integration test-e2e test-coverage clean deps lint fmt vet

# Variables
GO_VERSION := 1.21
PROJECT_NAME := ferre_pos_apis
BINARY_DIR := bin
TEST_TIMEOUT := 30m
COVERAGE_DIR := coverage

# Colores para output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[1;33m
BLUE := \033[0;34m
NC := \033[0m # No Color

help: ## Muestra esta ayuda
	@echo "$(BLUE)Comandos disponibles para $(PROJECT_NAME):$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-20s$(NC) %s\n", $$1, $$2}'

deps: ## Instala dependencias
	@echo "$(YELLOW)Instalando dependencias...$(NC)"
	go mod download
	go mod tidy
	go mod verify

build: deps ## Compila todos los ejecutables
	@echo "$(YELLOW)Compilando ejecutables...$(NC)"
	@mkdir -p $(BINARY_DIR)
	@echo "  - API POS"
	@go build -ldflags="-s -w" -o $(BINARY_DIR)/api_pos ./cmd/api_pos
	@echo "  - API Sync"
	@go build -ldflags="-s -w" -o $(BINARY_DIR)/api_sync ./cmd/api_sync
	@echo "  - API Labels"
	@go build -ldflags="-s -w" -o $(BINARY_DIR)/api_labels ./cmd/api_labels
	@echo "  - API Reports"
	@go build -ldflags="-s -w" -o $(BINARY_DIR)/api_reports ./cmd/api_reports
	@echo "$(GREEN)✓ Compilación completada$(NC)"

test: test-unit test-integration ## Ejecuta todos los tests

test-unit: ## Ejecuta tests unitarios
	@echo "$(YELLOW)Ejecutando tests unitarios...$(NC)"
	@mkdir -p $(COVERAGE_DIR)
	go test -v -race -timeout $(TEST_TIMEOUT) \
		-coverprofile=$(COVERAGE_DIR)/unit.out \
		-covermode=atomic \
		./test/unit/... \
		./internal/... \
		./pkg/...
	@echo "$(GREEN)✓ Tests unitarios completados$(NC)"

test-integration: ## Ejecuta tests de integración
	@echo "$(YELLOW)Ejecutando tests de integración...$(NC)"
	@mkdir -p $(COVERAGE_DIR)
	go test -v -race -timeout $(TEST_TIMEOUT) \
		-coverprofile=$(COVERAGE_DIR)/integration.out \
		-covermode=atomic \
		-tags=integration \
		./test/integration/...
	@echo "$(GREEN)✓ Tests de integración completados$(NC)"

test-e2e: ## Ejecuta tests end-to-end
	@echo "$(YELLOW)Ejecutando tests E2E...$(NC)"
	@mkdir -p $(COVERAGE_DIR)
	go test -v -timeout $(TEST_TIMEOUT) \
		-tags=e2e \
		./test/e2e/...
	@echo "$(GREEN)✓ Tests E2E completados$(NC)"

test-coverage: test-unit test-integration ## Genera reporte de cobertura
	@echo "$(YELLOW)Generando reporte de cobertura...$(NC)"
	@mkdir -p $(COVERAGE_DIR)
	@echo "mode: atomic" > $(COVERAGE_DIR)/coverage.out
	@tail -n +2 $(COVERAGE_DIR)/unit.out >> $(COVERAGE_DIR)/coverage.out
	@tail -n +2 $(COVERAGE_DIR)/integration.out >> $(COVERAGE_DIR)/coverage.out
	go tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	go tool cover -func=$(COVERAGE_DIR)/coverage.out | tail -1
	@echo "$(GREEN)✓ Reporte de cobertura generado en $(COVERAGE_DIR)/coverage.html$(NC)"

test-benchmark: ## Ejecuta benchmarks
	@echo "$(YELLOW)Ejecutando benchmarks...$(NC)"
	go test -bench=. -benchmem -timeout $(TEST_TIMEOUT) ./test/...
	@echo "$(GREEN)✓ Benchmarks completados$(NC)"

test-race: ## Ejecuta tests con detección de race conditions
	@echo "$(YELLOW)Ejecutando tests con detección de race conditions...$(NC)"
	go test -race -timeout $(TEST_TIMEOUT) ./...
	@echo "$(GREEN)✓ Tests de race conditions completados$(NC)"

lint: ## Ejecuta linter
	@echo "$(YELLOW)Ejecutando linter...$(NC)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "$(RED)golangci-lint no está instalado. Instalando...$(NC)"; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
		golangci-lint run; \
	fi
	@echo "$(GREEN)✓ Linting completado$(NC)"

fmt: ## Formatea el código
	@echo "$(YELLOW)Formateando código...$(NC)"
	go fmt ./...
	@echo "$(GREEN)✓ Formateo completado$(NC)"

vet: ## Ejecuta go vet
	@echo "$(YELLOW)Ejecutando go vet...$(NC)"
	go vet ./...
	@echo "$(GREEN)✓ Vet completado$(NC)"

clean: ## Limpia archivos generados
	@echo "$(YELLOW)Limpiando archivos generados...$(NC)"
	rm -rf $(BINARY_DIR)
	rm -rf $(COVERAGE_DIR)
	go clean -testcache
	go clean -cache
	@echo "$(GREEN)✓ Limpieza completada$(NC)"

docker-test: ## Ejecuta tests en contenedor Docker
	@echo "$(YELLOW)Ejecutando tests en Docker...$(NC)"
	docker run --rm -v $(PWD):/app -w /app golang:$(GO_VERSION) make test
	@echo "$(GREEN)✓ Tests en Docker completados$(NC)"

test-db-setup: ## Configura base de datos de test
	@echo "$(YELLOW)Configurando base de datos de test...$(NC)"
	@if command -v docker >/dev/null 2>&1; then \
		docker run --name postgres-test -e POSTGRES_PASSWORD=test -e POSTGRES_USER=test -e POSTGRES_DB=test_db -p 5433:5432 -d postgres:13; \
		sleep 5; \
		echo "$(GREEN)✓ Base de datos de test configurada en puerto 5433$(NC)"; \
	else \
		echo "$(RED)Docker no está disponible$(NC)"; \
	fi

test-db-teardown: ## Elimina base de datos de test
	@echo "$(YELLOW)Eliminando base de datos de test...$(NC)"
	@if command -v docker >/dev/null 2>&1; then \
		docker stop postgres-test || true; \
		docker rm postgres-test || true; \
		echo "$(GREEN)✓ Base de datos de test eliminada$(NC)"; \
	fi

ci: deps lint vet test-coverage ## Pipeline de CI completo
	@echo "$(GREEN)✓ Pipeline de CI completado$(NC)"

install-tools: ## Instala herramientas de desarrollo
	@echo "$(YELLOW)Instalando herramientas de desarrollo...$(NC)"
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/vektra/mockery/v2@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	@echo "$(GREEN)✓ Herramientas instaladas$(NC)"

generate-mocks: ## Genera mocks automáticamente
	@echo "$(YELLOW)Generando mocks...$(NC)"
	@if command -v mockery >/dev/null 2>&1; then \
		mockery --all --dir=internal --output=test/mocks --case=underscore; \
	else \
		echo "$(RED)mockery no está instalado. Ejecuta 'make install-tools' primero$(NC)"; \
	fi
	@echo "$(GREEN)✓ Mocks generados$(NC)"

docs: ## Genera documentación de API
	@echo "$(YELLOW)Generando documentación...$(NC)"
	@if command -v swag >/dev/null 2>&1; then \
		swag init -g cmd/api_pos/main.go -o docs/pos; \
		swag init -g cmd/api_sync/main.go -o docs/sync; \
		swag init -g cmd/api_labels/main.go -o docs/labels; \
		swag init -g cmd/api_reports/main.go -o docs/reports; \
	else \
		echo "$(RED)swag no está instalado. Ejecuta 'make install-tools' primero$(NC)"; \
	fi
	@echo "$(GREEN)✓ Documentación generada$(NC)"

run-pos: build ## Ejecuta API POS
	@echo "$(YELLOW)Iniciando API POS...$(NC)"
	./$(BINARY_DIR)/api_pos

run-sync: build ## Ejecuta API Sync
	@echo "$(YELLOW)Iniciando API Sync...$(NC)"
	./$(BINARY_DIR)/api_sync

run-labels: build ## Ejecuta API Labels
	@echo "$(YELLOW)Iniciando API Labels...$(NC)"
	./$(BINARY_DIR)/api_labels

run-reports: build ## Ejecuta API Reports
	@echo "$(YELLOW)Iniciando API Reports...$(NC)"
	./$(BINARY_DIR)/api_reports

run-all: build ## Ejecuta todas las APIs en paralelo
	@echo "$(YELLOW)Iniciando todas las APIs...$(NC)"
	./$(BINARY_DIR)/api_pos & \
	./$(BINARY_DIR)/api_sync & \
	./$(BINARY_DIR)/api_labels & \
	./$(BINARY_DIR)/api_reports & \
	wait

version: ## Muestra versión de Go
	@go version

# Targets por defecto
.DEFAULT_GOAL := help

