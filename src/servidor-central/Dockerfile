# Dockerfile para Ferre-POS Servidor Central
# Multi-stage build optimizado con estándares Go y notación húngara

# Stage 1: Builder con notación húngara en comentarios
FROM golang:1.21-alpine AS builder

# Variables de build con notación húngara
ARG STR_VERSION=dev
ARG STR_BUILD_TIME
ARG STR_GIT_COMMIT

# Metadatos con notación húngara
LABEL maintainer="Ferre-POS Team"
LABEL version="${STR_VERSION}"
LABEL description="Servidor Central Ferre-POS con notación húngara"

# Instalar dependencias del sistema con notación húngara
RUN apk add --no-cache \
    git \
    ca-certificates \
    tzdata \
    make

# Configurar directorio de trabajo con notación húngara
WORKDIR /app

# Copiar archivos de dependencias con notación húngara
COPY go.mod go.sum ./

# Descargar dependencias con notación húngara
RUN go mod download && go mod verify

# Copiar código fuente con notación húngara
COPY . .

# Construir ejecutables con notación húngara y optimizaciones
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.strVersion=${STR_VERSION} -X main.strBuildTime=${STR_BUILD_TIME} -X main.strGitCommit=${STR_GIT_COMMIT}" \
    -a -installsuffix cgo \
    -o bin/api_pos ./cmd/api_pos/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.strVersion=${STR_VERSION} -X main.strBuildTime=${STR_BUILD_TIME} -X main.strGitCommit=${STR_GIT_COMMIT}" \
    -a -installsuffix cgo \
    -o bin/api_sync ./cmd/api_sync/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.strVersion=${STR_VERSION} -X main.strBuildTime=${STR_BUILD_TIME} -X main.strGitCommit=${STR_GIT_COMMIT}" \
    -a -installsuffix cgo \
    -o bin/api_labels ./cmd/api_labels/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.strVersion=${STR_VERSION} -X main.strBuildTime=${STR_BUILD_TIME} -X main.strGitCommit=${STR_GIT_COMMIT}" \
    -a -installsuffix cgo \
    -o bin/api_reports ./cmd/api_reports/

# Stage 2: Runtime con notación húngara en comentarios
FROM alpine:3.18 AS runtime

# Instalar dependencias de runtime con notación húngara
RUN apk --no-cache add \
    ca-certificates \
    tzdata \
    curl \
    && update-ca-certificates

# Crear usuario no-root con notación húngara
RUN addgroup -g 1001 -S ferrepos && \
    adduser -u 1001 -S ferrepos -G ferrepos

# Crear directorios con notación húngara
RUN mkdir -p /app/bin /app/configs /app/logs /app/data && \
    chown -R ferrepos:ferrepos /app

# Configurar zona horaria con notación húngara
ENV TZ=America/Santiago
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# Configurar directorio de trabajo con notación húngara
WORKDIR /app

# Copiar ejecutables desde builder con notación húngara
COPY --from=builder --chown=ferrepos:ferrepos /app/bin/ ./bin/
COPY --from=builder --chown=ferrepos:ferrepos /app/configs/ ./configs/

# Hacer ejecutables con notación húngara
RUN chmod +x ./bin/*

# Cambiar a usuario no-root con notación húngara
USER ferrepos

# Exponer puertos con notación húngara
EXPOSE 8080 8081 8082 8083

# Variables de entorno con notación húngara
ENV STR_ENV=production
ENV STR_LOG_LEVEL=info
ENV STR_CONFIG_PATH=/app/configs

# Health check con notación húngara
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Comando por defecto con notación húngara (ejecuta API POS)
CMD ["./bin/api_pos"]

# Stage 3: Development con notación húngara en comentarios
FROM golang:1.21-alpine AS development

# Instalar herramientas de desarrollo con notación húngara
RUN apk add --no-cache \
    git \
    make \
    curl \
    bash \
    vim \
    ca-certificates

# Instalar herramientas Go con notación húngara
RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest && \
    go install github.com/swaggo/swag/cmd/swag@latest && \
    go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest && \
    go install golang.org/x/tools/cmd/godoc@latest

# Configurar directorio de trabajo con notación húngara
WORKDIR /app

# Copiar archivos de configuración con notación húngara
COPY go.mod go.sum Makefile .golangci.yml ./

# Descargar dependencias con notación húngara
RUN go mod download

# Comando por defecto para desarrollo con notación húngara
CMD ["make", "help"]

# Stage 4: Multi-service con notación húngara en comentarios
FROM runtime AS multiservice

# Script de inicio para múltiples servicios con notación húngara
COPY <<EOF /app/start-services.sh
#!/bin/sh
# Script de inicio para múltiples APIs con notación húngara

# Función de logging con notación húngara
log_with_hungarian() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] \$1"
}

# Función para iniciar servicio con notación húngara
start_service_hungarian() {
    local str_service_name=\$1
    local int_port=\$2
    local str_executable=\$3
    
    log_with_hungarian "Iniciando \$str_service_name en puerto \$int_port..."
    
    # Configurar puerto específico con notación húngara
    export PORT=\$int_port
    export SERVICE_NAME=\$str_service_name
    
    # Ejecutar servicio en background con notación húngara
    ./bin/\$str_executable &
    local int_pid=\$!
    
    log_with_hungarian "\$str_service_name iniciado con PID \$int_pid"
    echo \$int_pid > /tmp/\$str_service_name.pid
}

# Función de cleanup con notación húngara
cleanup_hungarian() {
    log_with_hungarian "Deteniendo servicios con notación húngara..."
    
    for str_service in api_pos api_sync api_labels api_reports; do
        if [ -f /tmp/\$str_service.pid ]; then
            local int_pid=\$(cat /tmp/\$str_service.pid)
            log_with_hungarian "Deteniendo \$str_service (PID: \$int_pid)..."
            kill \$int_pid 2>/dev/null || true
            rm -f /tmp/\$str_service.pid
        fi
    done
    
    log_with_hungarian "Todos los servicios detenidos"
    exit 0
}

# Configurar trap para cleanup con notación húngara
trap cleanup_hungarian SIGTERM SIGINT

# Iniciar todos los servicios con notación húngara
start_service_hungarian "api_pos" 8080 "api_pos"
start_service_hungarian "api_sync" 8081 "api_sync"
start_service_hungarian "api_labels" 8082 "api_labels"
start_service_hungarian "api_reports" 8083 "api_reports"

log_with_hungarian "Todos los servicios iniciados. Esperando..."

# Esperar indefinidamente con notación húngara
while true; do
    # Verificar que todos los servicios estén corriendo con notación húngara
    for str_service in api_pos api_sync api_labels api_reports; do
        if [ -f /tmp/\$str_service.pid ]; then
            local int_pid=\$(cat /tmp/\$str_service.pid)
            if ! kill -0 \$int_pid 2>/dev/null; then
                log_with_hungarian "Servicio \$str_service (PID: \$int_pid) ha fallado. Reiniciando..."
                case \$str_service in
                    api_pos) start_service_hungarian "api_pos" 8080 "api_pos" ;;
                    api_sync) start_service_hungarian "api_sync" 8081 "api_sync" ;;
                    api_labels) start_service_hungarian "api_labels" 8082 "api_labels" ;;
                    api_reports) start_service_hungarian "api_reports" 8083 "api_reports" ;;
                esac
            fi
        fi
    done
    
    sleep 10
done
EOF

# Hacer ejecutable el script con notación húngara
RUN chmod +x /app/start-services.sh

# Health check para múltiples servicios con notación húngara
HEALTHCHECK --interval=30s --timeout=10s --start-period=15s --retries=3 \
    CMD curl -f http://localhost:8080/health && \
        curl -f http://localhost:8081/health && \
        curl -f http://localhost:8082/health && \
        curl -f http://localhost:8083/health || exit 1

# Comando para múltiples servicios con notación húngara
CMD ["/app/start-services.sh"]

