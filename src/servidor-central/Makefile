# Makefile para Ferre-POS Servidor Central
# Cumple con estándares Go y usa notación húngara en comentarios

# Variables con notación húngara
STR_APP_NAME := ferre-pos-servidor-central
STR_VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
STR_BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
STR_GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Directorios con notación húngara
DIR_BUILD := ./build
DIR_BIN := ./bin
DIR_DOCS := ./docs
DIR_CONFIGS := ./configs

# Ejecutables con notación húngara
ARR_EXECUTABLES := api_pos api_sync api_labels api_reports

# Flags de compilación con notación húngara
STR_LDFLAGS := -ldflags "-X main.strVersion=$(STR_VERSION) -X main.strBuildTime=$(STR_BUILD_TIME) -X main.strGitCommit=$(STR_GIT_COMMIT) -w -s"

# Colores para output con notación húngara
STR_COLOR_RESET := \033[0m
STR_COLOR_GREEN := \033[32m
STR_COLOR_YELLOW := \033[33m
STR_COLOR_BLUE := \033[34m
STR_COLOR_RED := \033[31m

.PHONY: help build build-all clean test test-coverage lint fmt vet security deps tidy run-pos run-sync run-labels run-reports docker-build docker-run install uninstall docs swagger

# Target por defecto
all: clean deps tidy fmt vet lint test build-all

# Ayuda con notación húngara en comentarios
help: ## Muestra esta ayuda con notación húngara
	@echo "$(STR_COLOR_BLUE)Ferre-POS Servidor Central - Makefile$(STR_COLOR_RESET)"
	@echo "$(STR_COLOR_YELLOW)Versión: $(STR_VERSION)$(STR_COLOR_RESET)"
	@echo ""
	@echo "$(STR_COLOR_GREEN)Targets disponibles:$(STR_COLOR_RESET)"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(STR_COLOR_BLUE)%-20s$(STR_COLOR_RESET) %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Construcción con notación húngara
build: ## Construye todos los ejecutables con notación húngara
	@echo "$(STR_COLOR_GREEN)Construyendo ejecutables con notación húngara...$(STR_COLOR_RESET)"
	@mkdir -p $(DIR_BIN)
	@for strExecutable in $(ARR_EXECUTABLES); do \
		echo "$(STR_COLOR_YELLOW)Construyendo $$strExecutable...$(STR_COLOR_RESET)"; \
		go build $(STR_LDFLAGS) -o $(DIR_BIN)/$$strExecutable ./cmd/$$strExecutable/; \
	done
	@echo "$(STR_COLOR_GREEN)✅ Construcción completada$(STR_COLOR_RESET)"

build-all: build ## Alias para build con notación húngara

# Construcción específica por API con notación húngara
build-pos: ## Construye solo API POS con notación húngara
	@echo "$(STR_COLOR_GREEN)Construyendo API POS con notación húngara...$(STR_COLOR_RESET)"
	@mkdir -p $(DIR_BIN)
	go build $(STR_LDFLAGS) -o $(DIR_BIN)/api_pos ./cmd/api_pos/

build-sync: ## Construye solo API Sync con notación húngara
	@echo "$(STR_COLOR_GREEN)Construyendo API Sync con notación húngara...$(STR_COLOR_RESET)"
	@mkdir -p $(DIR_BIN)
	go build $(STR_LDFLAGS) -o $(DIR_BIN)/api_sync ./cmd/api_sync/

build-labels: ## Construye solo API Labels con notación húngara
	@echo "$(STR_COLOR_GREEN)Construyendo API Labels con notación húngara...$(STR_COLOR_RESET)"
	@mkdir -p $(DIR_BIN)
	go build $(STR_LDFLAGS) -o $(DIR_BIN)/api_labels ./cmd/api_labels/

build-reports: ## Construye solo API Reports con notación húngara
	@echo "$(STR_COLOR_GREEN)Construyendo API Reports con notación húngara...$(STR_COLOR_RESET)"
	@mkdir -p $(DIR_BIN)
	go build $(STR_LDFLAGS) -o $(DIR_BIN)/api_reports ./cmd/api_reports/

# Limpieza con notación húngara
clean: ## Limpia archivos de construcción con notación húngara
	@echo "$(STR_COLOR_YELLOW)Limpiando archivos con notación húngara...$(STR_COLOR_RESET)"
	@rm -rf $(DIR_BIN)
	@rm -rf $(DIR_BUILD)
	@rm -f coverage.out
	@rm -f coverage.html
	@echo "$(STR_COLOR_GREEN)✅ Limpieza completada$(STR_COLOR_RESET)"

# Testing con notación húngara
test: ## Ejecuta tests con notación húngara
	@echo "$(STR_COLOR_GREEN)Ejecutando tests con notación húngara...$(STR_COLOR_RESET)"
	go test -v -race ./...

test-coverage: ## Ejecuta tests con coverage con notación húngara
	@echo "$(STR_COLOR_GREEN)Ejecutando tests con coverage y notación húngara...$(STR_COLOR_RESET)"
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "$(STR_COLOR_BLUE)Coverage report: coverage.html$(STR_COLOR_RESET)"

test-benchmark: ## Ejecuta benchmarks con notación húngara
	@echo "$(STR_COLOR_GREEN)Ejecutando benchmarks con notación húngara...$(STR_COLOR_RESET)"
	go test -bench=. -benchmem ./...

# Linting y formateo con notación húngara
fmt: ## Formatea código con notación húngara
	@echo "$(STR_COLOR_GREEN)Formateando código con notación húngara...$(STR_COLOR_RESET)"
	go fmt ./...

vet: ## Ejecuta go vet con notación húngara
	@echo "$(STR_COLOR_GREEN)Ejecutando go vet con notación húngara...$(STR_COLOR_RESET)"
	go vet ./...

lint: ## Ejecuta golangci-lint con notación húngara
	@echo "$(STR_COLOR_GREEN)Ejecutando linting con notación húngara...$(STR_COLOR_RESET)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "$(STR_COLOR_YELLOW)golangci-lint no instalado, instalando...$(STR_COLOR_RESET)"; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
		golangci-lint run; \
	fi

# Seguridad con notación húngara
security: ## Ejecuta análisis de seguridad con notación húngara
	@echo "$(STR_COLOR_GREEN)Ejecutando análisis de seguridad con notación húngara...$(STR_COLOR_RESET)"
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "$(STR_COLOR_YELLOW)gosec no instalado, instalando...$(STR_COLOR_RESET)"; \
		go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest; \
		gosec ./...; \
	fi

# Dependencias con notación húngara
deps: ## Descarga dependencias con notación húngara
	@echo "$(STR_COLOR_GREEN)Descargando dependencias con notación húngara...$(STR_COLOR_RESET)"
	go mod download

tidy: ## Limpia dependencias con notación húngara
	@echo "$(STR_COLOR_GREEN)Limpiando dependencias con notación húngara...$(STR_COLOR_RESET)"
	go mod tidy

vendor: ## Crea directorio vendor con notación húngara
	@echo "$(STR_COLOR_GREEN)Creando vendor con notación húngara...$(STR_COLOR_RESET)"
	go mod vendor

# Ejecución con notación húngara
run-pos: build-pos ## Ejecuta API POS con notación húngara
	@echo "$(STR_COLOR_GREEN)Ejecutando API POS con notación húngara...$(STR_COLOR_RESET)"
	./$(DIR_BIN)/api_pos

run-sync: build-sync ## Ejecuta API Sync con notación húngara
	@echo "$(STR_COLOR_GREEN)Ejecutando API Sync con notación húngara...$(STR_COLOR_RESET)"
	./$(DIR_BIN)/api_sync

run-labels: build-labels ## Ejecuta API Labels con notación húngara
	@echo "$(STR_COLOR_GREEN)Ejecutando API Labels con notación húngara...$(STR_COLOR_RESET)"
	./$(DIR_BIN)/api_labels

run-reports: build-reports ## Ejecuta API Reports con notación húngara
	@echo "$(STR_COLOR_GREEN)Ejecutando API Reports con notación húngara...$(STR_COLOR_RESET)"
	./$(DIR_BIN)/api_reports

# Docker con notación húngara
docker-build: ## Construye imagen Docker con notación húngara
	@echo "$(STR_COLOR_GREEN)Construyendo imagen Docker con notación húngara...$(STR_COLOR_RESET)"
	docker build -t $(STR_APP_NAME):$(STR_VERSION) .
	docker build -t $(STR_APP_NAME):latest .

docker-run: ## Ejecuta contenedor Docker con notación húngara
	@echo "$(STR_COLOR_GREEN)Ejecutando contenedor Docker con notación húngara...$(STR_COLOR_RESET)"
	docker run -p 8080:8080 -p 8081:8081 -p 8082:8082 -p 8083:8083 $(STR_APP_NAME):latest

docker-compose-up: ## Ejecuta docker-compose con notación húngara
	@echo "$(STR_COLOR_GREEN)Ejecutando docker-compose con notación húngara...$(STR_COLOR_RESET)"
	docker-compose up -d

docker-compose-down: ## Detiene docker-compose con notación húngara
	@echo "$(STR_COLOR_GREEN)Deteniendo docker-compose con notación húngara...$(STR_COLOR_RESET)"
	docker-compose down

# Instalación con notación húngara
install: build ## Instala ejecutables con notación húngara
	@echo "$(STR_COLOR_GREEN)Instalando ejecutables con notación húngara...$(STR_COLOR_RESET)"
	@for strExecutable in $(ARR_EXECUTABLES); do \
		echo "$(STR_COLOR_YELLOW)Instalando $$strExecutable...$(STR_COLOR_RESET)"; \
		go install $(STR_LDFLAGS) ./cmd/$$strExecutable/; \
	done
	@echo "$(STR_COLOR_GREEN)✅ Instalación completada$(STR_COLOR_RESET)"

uninstall: ## Desinstala ejecutables con notación húngara
	@echo "$(STR_COLOR_GREEN)Desinstalando ejecutables con notación húngara...$(STR_COLOR_RESET)"
	@for strExecutable in $(ARR_EXECUTABLES); do \
		echo "$(STR_COLOR_YELLOW)Desinstalando $$strExecutable...$(STR_COLOR_RESET)"; \
		rm -f $(GOPATH)/bin/$$strExecutable; \
	done
	@echo "$(STR_COLOR_GREEN)✅ Desinstalación completada$(STR_COLOR_RESET)"

# Documentación con notación húngara
docs: ## Genera documentación con notación húngara
	@echo "$(STR_COLOR_GREEN)Generando documentación con notación húngara...$(STR_COLOR_RESET)"
	@mkdir -p $(DIR_DOCS)
	godoc -http=:6060 &
	@echo "$(STR_COLOR_BLUE)Documentación disponible en: http://localhost:6060$(STR_COLOR_RESET)"

swagger: ## Genera documentación Swagger con notación húngara
	@echo "$(STR_COLOR_GREEN)Generando Swagger con notación húngara...$(STR_COLOR_RESET)"
	@if command -v swag >/dev/null 2>&1; then \
		swag init -g cmd/api_pos/main.go -o docs/swagger/pos; \
		swag init -g cmd/api_sync/main.go -o docs/swagger/sync; \
		swag init -g cmd/api_labels/main.go -o docs/swagger/labels; \
		swag init -g cmd/api_reports/main.go -o docs/swagger/reports; \
	else \
		echo "$(STR_COLOR_YELLOW)swag no instalado, instalando...$(STR_COLOR_RESET)"; \
		go install github.com/swaggo/swag/cmd/swag@latest; \
		swag init -g cmd/api_pos/main.go -o docs/swagger/pos; \
		swag init -g cmd/api_sync/main.go -o docs/swagger/sync; \
		swag init -g cmd/api_labels/main.go -o docs/swagger/labels; \
		swag init -g cmd/api_reports/main.go -o docs/swagger/reports; \
	fi

# Herramientas de desarrollo con notación húngara
tools: ## Instala herramientas de desarrollo con notación húngara
	@echo "$(STR_COLOR_GREEN)Instalando herramientas con notación húngara...$(STR_COLOR_RESET)"
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	go install golang.org/x/tools/cmd/godoc@latest
	go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Migración de base de datos con notación húngara
migrate-up: ## Ejecuta migraciones up con notación húngara
	@echo "$(STR_COLOR_GREEN)Ejecutando migraciones up con notación húngara...$(STR_COLOR_RESET)"
	migrate -path migrations -database "postgres://user:password@localhost/ferre_pos?sslmode=disable" up

migrate-down: ## Ejecuta migraciones down con notación húngara
	@echo "$(STR_COLOR_GREEN)Ejecutando migraciones down con notación húngara...$(STR_COLOR_RESET)"
	migrate -path migrations -database "postgres://user:password@localhost/ferre_pos?sslmode=disable" down

migrate-create: ## Crea nueva migración con notación húngara
	@echo "$(STR_COLOR_GREEN)Creando migración con notación húngara...$(STR_COLOR_RESET)"
	@read -p "Nombre de la migración: " strMigrationName; \
	migrate create -ext sql -dir migrations $$strMigrationName

# Métricas y monitoreo con notación húngara
metrics: ## Muestra métricas con notación húngara
	@echo "$(STR_COLOR_GREEN)Métricas disponibles con notación húngara:$(STR_COLOR_RESET)"
	@echo "$(STR_COLOR_BLUE)- POS API: http://localhost:8080/metrics$(STR_COLOR_RESET)"
	@echo "$(STR_COLOR_BLUE)- Sync API: http://localhost:8081/metrics$(STR_COLOR_RESET)"
	@echo "$(STR_COLOR_BLUE)- Labels API: http://localhost:8082/metrics$(STR_COLOR_RESET)"
	@echo "$(STR_COLOR_BLUE)- Reports API: http://localhost:8083/metrics$(STR_COLOR_RESET)"

health: ## Verifica salud de APIs con notación húngara
	@echo "$(STR_COLOR_GREEN)Verificando salud de APIs con notación húngara...$(STR_COLOR_RESET)"
	@curl -s http://localhost:8080/health || echo "$(STR_COLOR_RED)POS API no disponible$(STR_COLOR_RESET)"
	@curl -s http://localhost:8081/health || echo "$(STR_COLOR_RED)Sync API no disponible$(STR_COLOR_RESET)"
	@curl -s http://localhost:8082/health || echo "$(STR_COLOR_RED)Labels API no disponible$(STR_COLOR_RESET)"
	@curl -s http://localhost:8083/health || echo "$(STR_COLOR_RED)Reports API no disponible$(STR_COLOR_RESET)"

# Release con notación húngara
release: clean deps tidy fmt vet lint test security build-all ## Prepara release con notación húngara
	@echo "$(STR_COLOR_GREEN)Preparando release $(STR_VERSION) con notación húngara...$(STR_COLOR_RESET)"
	@mkdir -p $(DIR_BUILD)
	@for strExecutable in $(ARR_EXECUTABLES); do \
		echo "$(STR_COLOR_YELLOW)Empaquetando $$strExecutable...$(STR_COLOR_RESET)"; \
		tar -czf $(DIR_BUILD)/$$strExecutable-$(STR_VERSION)-linux-amd64.tar.gz -C $(DIR_BIN) $$strExecutable; \
	done
	@echo "$(STR_COLOR_GREEN)✅ Release $(STR_VERSION) preparado en $(DIR_BUILD)$(STR_COLOR_RESET)"

# Información del proyecto con notación húngara
info: ## Muestra información del proyecto con notación húngara
	@echo "$(STR_COLOR_BLUE)Ferre-POS Servidor Central$(STR_COLOR_RESET)"
	@echo "$(STR_COLOR_YELLOW)Versión:$(STR_COLOR_RESET) $(STR_VERSION)"
	@echo "$(STR_COLOR_YELLOW)Commit:$(STR_COLOR_RESET) $(STR_GIT_COMMIT)"
	@echo "$(STR_COLOR_YELLOW)Build Time:$(STR_COLOR_RESET) $(STR_BUILD_TIME)"
	@echo "$(STR_COLOR_YELLOW)Go Version:$(STR_COLOR_RESET) $(shell go version)"
	@echo "$(STR_COLOR_YELLOW)Ejecutables:$(STR_COLOR_RESET) $(ARR_EXECUTABLES)"

