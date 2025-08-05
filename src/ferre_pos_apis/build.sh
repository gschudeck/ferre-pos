#!/bin/bash

# Script de compilación para APIs FERRE-POS
# Genera los 4 ejecutables: api_pos, api_sync, api_labels, api_reports

set -e

echo "🔨 Compilando APIs FERRE-POS..."

# Crear directorio bin si no existe
mkdir -p bin

# Limpiar binarios anteriores
rm -f bin/api_*

echo "📦 Descargando dependencias..."
go mod tidy

echo "🏗️  Compilando API POS..."
go build -ldflags="-s -w" -o bin/api_pos ./cmd/api_pos

echo "🔄 Compilando API Sync..."
go build -ldflags="-s -w" -o bin/api_sync ./cmd/api_sync

echo "🏷️  Compilando API Labels..."
go build -ldflags="-s -w" -o bin/api_labels ./cmd/api_labels

echo "📊 Compilando API Reports..."
go build -ldflags="-s -w" -o bin/api_reports ./cmd/api_reports

echo "✅ Compilación completada exitosamente!"
echo ""
echo "Ejecutables generados:"
ls -lh bin/

echo ""
echo "Para ejecutar los APIs:"
echo "  ./bin/api_pos     - API de Punto de Venta"
echo "  ./bin/api_sync    - API de Sincronización"
echo "  ./bin/api_labels  - API de Etiquetas"
echo "  ./bin/api_reports - API de Reportes"

