# Servidor Central Ferre-POS

## Descripción General

El **Servidor Central Ferre-POS** es una solución completa de backend desarrollada en Go que proporciona APIs REST especializadas para la gestión integral de un sistema de punto de venta para ferreterías. Este sistema está diseñado para manejar operaciones críticas como ventas, inventario, sincronización de datos, generación de etiquetas y reportes analíticos.

### Características Principales

- **Arquitectura Modular**: Cuatro APIs especializadas (POS, Sync, Labels, Reports)
- **Configuración Modificable**: Sistema de configuración en tiempo real sin reiniciar
- **Base de Datos Optimizada**: Conexiones diferenciadas por API con pools optimizados
- **Seguridad Robusta**: Autenticación JWT, autorización por roles, CORS configurable
- **Monitoreo Integrado**: Métricas, logging estructurado y health checks
- **Sincronización Avanzada**: Resolución de conflictos y sincronización bidireccional
- **Generación de Contenido**: Etiquetas personalizables y reportes programables

## Arquitectura del Sistema

### APIs Disponibles

#### 1. API POS (`/api/pos`)
Maneja todas las operaciones críticas del punto de venta:
- Gestión de productos y stock
- Procesamiento de ventas
- Administración de clientes
- Sistema de fidelización
- Autenticación de usuarios y terminales

#### 2. API Sync (`/api/sync`)
Sincronización de datos entre servidor central y sucursales:
- Sincronización bidireccional
- Resolución automática de conflictos
- Logs detallados de operaciones
- Configuración granular por entidad

#### 3. API Labels (`/api/labels`)
Generación y gestión de etiquetas:
- Plantillas personalizables
- Generación por lotes
- Múltiples formatos (PDF, PNG, JPG)
- Sistema de preview

#### 4. API Reports (`/api/reports`)
Sistema de reportes y análisis:
- Reportes programados
- Dashboards interactivos
- Múltiples formatos de exportación
- Plantillas SQL personalizables

## Instalación y Configuración

### Requisitos del Sistema

- **Go**: 1.19 o superior
- **PostgreSQL**: 12 o superior
- **Sistema Operativo**: Linux, macOS, Windows
- **Memoria RAM**: Mínimo 2GB, recomendado 4GB
- **Espacio en Disco**: Mínimo 1GB para logs y archivos temporales

### Instalación

1. **Clonar el repositorio**:
```bash
git clone https://github.com/tu-organizacion/ferre-pos-servidor-central.git
cd ferre-pos-servidor-central
```

2. **Instalar dependencias**:
```bash
go mod download
```

3. **Configurar base de datos**:
```bash
# Crear base de datos
createdb ferre_pos_central

# Ejecutar migraciones
psql -d ferre_pos_central -f schema/ferre_pos_servidor_central_schema_optimizado.sql
```

4. **Configurar archivo de configuración**:
```bash
cp configs/config.yaml.example configs/config.yaml
# Editar configs/config.yaml con tus configuraciones
```

5. **Compilar y ejecutar**:
```bash
go build -o bin/ferre-pos-server cmd/server/main.go
./bin/ferre-pos-server
```

### Variables de Entorno

| Variable | Descripción | Valor por Defecto |
|----------|-------------|-------------------|
| `FERRE_POS_CONFIG` | Ruta del archivo de configuración | `configs/config.yaml` |
| `FERRE_POS_ENV` | Entorno de ejecución | `production` |
| `FERRE_POS_LOG_LEVEL` | Nivel de logging | `info` |

## Configuración

### Archivo Principal (`configs/config.yaml`)

El sistema utiliza un archivo de configuración YAML principal que contiene todas las configuraciones del servidor. Este archivo puede ser modificado en tiempo real y los cambios se aplicarán automáticamente.

#### Secciones Principales:

- **server**: Configuración del servidor HTTP
- **database**: Configuraciones de base de datos por API
- **apis**: Configuraciones específicas de cada API
- **security**: Configuración de seguridad y autenticación
- **logging**: Configuración de logging por API
- **cache**: Configuración de cache
- **storage**: Configuración de almacenamiento
- **monitoring**: Configuración de monitoreo

### Configuraciones por API

Cada API tiene su propio archivo de configuración que puede ser modificado independientemente:

- `configs/pos/pos-config.yaml`: Configuración del API POS
- `configs/sync/sync-config.yaml`: Configuración del API Sync
- `configs/labels/labels-config.yaml`: Configuración del API Labels
- `configs/reports/reports-config.yaml`: Configuración del API Reports

### Recarga en Caliente

El sistema incluye un mecanismo de recarga en caliente que permite modificar configuraciones sin reiniciar el servidor:

- **Monitoreo automático** de cambios en archivos de configuración
- **Validación previa** antes de aplicar cambios
- **Rollback automático** en caso de configuración inválida
- **Notificaciones** a todos los componentes afectados

## Uso de las APIs

### Autenticación

Todas las APIs (excepto endpoints públicos) requieren autenticación JWT:

```bash
# Login
curl -X POST http://localhost:8080/api/pos/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "usuario@ferreteria.com", "password": "password"}'

# Usar token en requests
curl -X GET http://localhost:8080/api/pos/productos \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Ejemplos de Uso

#### API POS - Crear Venta
```bash
curl -X POST http://localhost:8080/api/pos/ventas \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "cliente_id": "uuid-del-cliente",
    "items": [
      {
        "producto_id": "uuid-del-producto",
        "cantidad": 2,
        "precio_unitario": 15000
      }
    ],
    "medio_pago": "efectivo"
  }'
```

#### API Labels - Generar Etiquetas
```bash
curl -X POST http://localhost:8080/api/labels/generar/lote \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "plantilla_id": "uuid-de-plantilla",
    "productos": ["uuid1", "uuid2", "uuid3"],
    "formato": "pdf"
  }'
```

#### API Reports - Generar Reporte
```bash
curl -X POST http://localhost:8080/api/reports/generar \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "plantilla_id": "uuid-de-plantilla",
    "parametros": {
      "fecha_inicio": "2024-01-01",
      "fecha_fin": "2024-01-31"
    },
    "formato": "pdf"
  }'
```

## Monitoreo y Logging

### Health Checks

El sistema proporciona endpoints de health check para cada API:

- `GET /health` - Health check global
- `GET /api/pos/health` - Health check API POS
- `GET /api/sync/health` - Health check API Sync
- `GET /api/labels/health` - Health check API Labels
- `GET /api/reports/health` - Health check API Reports

### Métricas

Las métricas están disponibles en formato Prometheus:

```bash
curl http://localhost:8080/metrics
```

### Logs

Los logs se generan en formato JSON estructurado y pueden configurarse por API:

- **Logs globales**: `/var/log/ferre-pos/app.log`
- **Logs por API**: `/var/log/ferre-pos/{api}.log`

## Desarrollo

### Estructura del Proyecto

```
ferre-pos-servidor-central/
├── cmd/
│   └── server/
│       └── main.go              # Punto de entrada principal
├── internal/
│   ├── config/                  # Sistema de configuración
│   ├── controllers/             # Controladores de API
│   ├── database/                # Gestión de base de datos
│   ├── middleware/              # Middleware HTTP
│   ├── models/                  # Modelos de datos
│   ├── repositories/            # Capa de acceso a datos
│   └── services/                # Lógica de negocio
├── configs/                     # Archivos de configuración
├── docs/                        # Documentación
├── schema/                      # Esquemas de base de datos
└── README.md
```

### Agregar Nueva Funcionalidad

1. **Crear modelo** en `internal/models/`
2. **Implementar repositorio** en `internal/repositories/`
3. **Crear servicio** en `internal/services/`
4. **Implementar controlador** en `internal/controllers/`
5. **Agregar rutas** en `cmd/server/main.go`
6. **Actualizar configuración** si es necesario

### Testing

```bash
# Ejecutar tests
go test ./...

# Tests con coverage
go test -cover ./...

# Tests de integración
go test -tags=integration ./...
```

## Deployment

### Docker

```dockerfile
FROM golang:1.19-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o ferre-pos-server cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/ferre-pos-server .
COPY --from=builder /app/configs ./configs
CMD ["./ferre-pos-server"]
```

### Docker Compose

```yaml
version: '3.8'
services:
  ferre-pos-server:
    build: .
    ports:
      - "8080:8080"
    environment:
      - FERRE_POS_CONFIG=/app/configs/config.yaml
    volumes:
      - ./configs:/app/configs
      - ./logs:/var/log/ferre-pos
    depends_on:
      - postgres

  postgres:
    image: postgres:14
    environment:
      POSTGRES_DB: ferre_pos_central
      POSTGRES_USER: ferre_pos_user
      POSTGRES_PASSWORD: secure_password
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./schema:/docker-entrypoint-initdb.d

volumes:
  postgres_data:
```

## Seguridad

### Mejores Prácticas Implementadas

- **Autenticación JWT** con expiración configurable
- **Autorización basada en roles** y permisos granulares
- **Rate limiting** configurable por API
- **Validación de entrada** en todos los endpoints
- **Headers de seguridad** automáticos
- **Logging de auditoría** para todas las operaciones

### Configuración de Seguridad

```yaml
security:
  jwt_secret: "your-super-secret-jwt-key"
  jwt_expiration: "24h"
  max_login_attempts: 5
  login_lockout_duration: "15m"
  enable_two_factor: false
  rate_limiting:
    enabled: true
    requests_per_minute: 100
```

## Troubleshooting

### Problemas Comunes

#### Error de Conexión a Base de Datos
```
Error: failed to connect to database
```
**Solución**: Verificar configuración de base de datos en `configs/config.yaml`

#### Error de Configuración
```
Error: configuración inválida
```
**Solución**: Validar sintaxis YAML y valores de configuración

#### Error de Permisos
```
Error: insufficient permissions
```
**Solución**: Verificar roles y permisos del usuario autenticado

### Logs de Debug

Para habilitar logs de debug:

```yaml
logging:
  global:
    level: "debug"
```

## Contribución

### Guías de Contribución

1. Fork el repositorio
2. Crear branch para feature (`git checkout -b feature/nueva-funcionalidad`)
3. Commit cambios (`git commit -am 'Agregar nueva funcionalidad'`)
4. Push al branch (`git push origin feature/nueva-funcionalidad`)
5. Crear Pull Request

### Estándares de Código

- Seguir convenciones de Go (gofmt, golint)
- Documentar funciones públicas
- Escribir tests para nueva funcionalidad
- Mantener cobertura de tests > 80%

## Licencia

Este proyecto está licenciado bajo la Licencia MIT. Ver archivo `LICENSE` para más detalles.

## Soporte

Para soporte técnico:
- **Email**: soporte@ferreteria.com
- **Issues**: GitHub Issues
- **Documentación**: `/docs` directory

## Changelog

### v1.0.0 (2024-01-XX)
- Implementación inicial del servidor central
- APIs POS, Sync, Labels y Reports
- Sistema de configuración en tiempo real
- Autenticación y autorización completa
- Monitoreo y logging integrado

---

**Desarrollado por**: Manus AI  
**Versión**: 1.0.0  
**Última actualización**: 2024-01-XX

