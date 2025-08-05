# Sistema FERRE-POS - APIs REST Centralizadas

## Descripción

Sistema de APIs REST centralizadas para el sistema FERRE-POS, implementado en Go con arquitectura de microservicios. El sistema está compuesto por 4 módulos API independientes con diferentes niveles de prioridad:

- **api_pos** (Prioridad Máxima): Operaciones críticas del punto de venta
- **api_sync** (Prioridad Media): Sincronización con sistemas ERP
- **api_labels** (Prioridad Baja): Generación y gestión de etiquetas
- **api_reports** (Prioridad Mínima): Generación de reportes y análisis

## Arquitectura

### Estructura del Proyecto

```
ferre_pos_apis/
├── cmd/                          # Ejecutables principales
│   ├── api_pos/                  # API POS (Puerto 8080)
│   ├── api_sync/                 # API Sync (Puerto 8081)
│   ├── api_labels/               # API Labels (Puerto 8082)
│   └── api_reports/              # API Reports (Puerto 8083)
├── internal/                     # Código interno compartido
│   ├── common/                   # Utilidades comunes
│   ├── models/                   # Modelos de datos
│   ├── handlers/                 # Manejadores HTTP
│   ├── middleware/               # Middleware personalizado
│   ├── database/                 # Conexión y operaciones DB
│   ├── config/                   # Gestión de configuración
│   ├── metrics/                  # Métricas Prometheus
│   └── logger/                   # Sistema de logging
├── pkg/                          # Paquetes públicos
│   ├── validator/                # Validación de datos
│   ├── ratelimiter/             # Rate limiting
│   └── auth/                     # Autenticación y autorización
├── configs/                      # Archivos de configuración
├── docs/                         # Documentación
└── scripts/                      # Scripts de utilidad
```

### Características Principales

#### 1. Configuración Modificable
- Archivo YAML centralizado (`configs/config.yaml`)
- Configuración específica por API
- Variables de entorno soportadas
- Recarga en caliente (desarrollo)

#### 2. Sistema de Logging Avanzado
- Logging estructurado con Logrus
- Niveles configurables (debug, info, warn, error)
- Rotación automática de logs
- Formato JSON para análisis automatizado
- Correlación de requests con Request ID

#### 3. Manejo de Errores Mejorado
- Errores tipificados por dominio
- Stack traces detallados
- Códigos de error consistentes
- Respuestas de error estandarizadas
- Logging automático de errores

#### 4. Validación de Datos
- Validación automática con tags struct
- Validaciones personalizadas
- Sanitización de entrada
- Mensajes de error localizados

#### 5. Concurrencia Mejorada
- Control de concurrencia optimista
- Pools de conexiones configurables
- Timeouts por operación
- Context cancellation
- Prevención de condiciones de carrera

#### 6. Rate Limiting
- Rate limiting por IP y usuario
- Configuración por API
- Algoritmo token bucket
- Headers informativos
- Bypass para usuarios privilegiados

#### 7. Métricas Prometheus
- Métricas de aplicación personalizadas
- Métricas de runtime Go
- Métricas de base de datos
- Histogramas de latencia
- Contadores de errores

#### 8. Estándares Go
- Estructura de proyecto estándar
- Naming conventions
- Error handling idiomático
- Interfaces bien definidas
- Testing comprehensivo

## APIs Disponibles

### API POS (Puerto 8080)
**Prioridad: Máxima**
- Gestión de ventas y transacciones
- Consulta de productos y stock
- Autenticación de usuarios
- Operaciones de caja

**Endpoints principales:**
- `POST /api/v1/ventas` - Crear venta
- `GET /api/v1/productos/{id}` - Consultar producto
- `GET /api/v1/stock/{producto_id}` - Consultar stock
- `POST /api/v1/auth/login` - Autenticación

### API Sync (Puerto 8081)
**Prioridad: Media**
- Sincronización con ERP
- Importación/exportación de datos
- Procesamiento por lotes
- Reconciliación de datos

**Endpoints principales:**
- `POST /api/v1/sync/productos` - Sincronizar productos
- `POST /api/v1/sync/ventas` - Sincronizar ventas
- `GET /api/v1/sync/status` - Estado de sincronización
- `POST /api/v1/sync/batch` - Procesamiento por lotes

### API Labels (Puerto 8082)
**Prioridad: Baja**
- Generación de etiquetas
- Gestión de plantillas
- Configuración de impresoras
- Trabajos de impresión

**Endpoints principales:**
- `POST /api/v1/etiquetas/generar` - Generar etiquetas
- `GET /api/v1/plantillas` - Listar plantillas
- `POST /api/v1/trabajos` - Crear trabajo de impresión
- `GET /api/v1/trabajos/{id}/status` - Estado del trabajo

### API Reports (Puerto 8083)
**Prioridad: Mínima**
- Generación de reportes
- Análisis de datos
- Exportación de información
- Dashboard de métricas

**Endpoints principales:**
- `POST /api/v1/reportes/ventas` - Reporte de ventas
- `POST /api/v1/reportes/stock` - Reporte de stock
- `GET /api/v1/reportes/{id}/download` - Descargar reporte
- `GET /api/v1/dashboard/metricas` - Métricas del dashboard

## Instalación y Configuración

### Prerrequisitos
- Go 1.21 o superior
- PostgreSQL 14 o superior
- Redis (opcional, para cache)

### Instalación

1. Clonar el repositorio:
```bash
git clone <repository-url>
cd ferre_pos_apis
```

2. Instalar dependencias:
```bash
go mod download
```

3. Configurar base de datos:
```bash
# Ejecutar migraciones
go run scripts/migrate.go up
```

4. Configurar archivo de configuración:
```bash
cp configs/config.yaml.example configs/config.yaml
# Editar configs/config.yaml según el entorno
```

### Ejecución

#### Desarrollo (todos los servicios)
```bash
# Terminal 1 - API POS
go run cmd/api_pos/main.go

# Terminal 2 - API Sync
go run cmd/api_sync/main.go

# Terminal 3 - API Labels
go run cmd/api_labels/main.go

# Terminal 4 - API Reports
go run cmd/api_reports/main.go
```

#### Producción (compilado)
```bash
# Compilar todos los servicios
make build

# Ejecutar servicios
./bin/api_pos &
./bin/api_sync &
./bin/api_labels &
./bin/api_reports &
```

### Docker (Opcional)
```bash
# Construir imágenes
docker-compose build

# Ejecutar servicios
docker-compose up -d
```

## Monitoreo y Observabilidad

### Métricas Prometheus
- Endpoint: `http://localhost:9090/metrics` (cada API)
- Dashboards Grafana incluidos en `/docs/grafana/`

### Health Checks
- API POS: `http://localhost:8080/health`
- API Sync: `http://localhost:8081/health`
- API Labels: `http://localhost:8082/health`
- API Reports: `http://localhost:8083/health`

### Logs
- Ubicación: `/var/log/ferre_pos/` (configurable)
- Formato: JSON estructurado
- Rotación automática

## Testing

```bash
# Ejecutar todos los tests
go test ./...

# Tests con coverage
go test -cover ./...

# Tests de integración
go test -tags=integration ./...

# Benchmarks
go test -bench=. ./...
```

## Contribución

1. Fork del proyecto
2. Crear rama feature (`git checkout -b feature/nueva-funcionalidad`)
3. Commit cambios (`git commit -am 'Agregar nueva funcionalidad'`)
4. Push a la rama (`git push origin feature/nueva-funcionalidad`)
5. Crear Pull Request

## Licencia

Este proyecto está bajo la Licencia MIT. Ver archivo `LICENSE` para más detalles.

## Soporte

Para soporte técnico o consultas:
- Email: soporte@ferre-pos.com
- Documentación: `/docs/`
- Issues: GitHub Issues

