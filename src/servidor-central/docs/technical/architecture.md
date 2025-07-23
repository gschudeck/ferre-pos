# Arquitectura Técnica del Servidor Central Ferre-POS

## Introducción

El Servidor Central Ferre-POS está diseñado siguiendo principios de arquitectura moderna, escalabilidad y mantenibilidad. Este documento describe en detalle la arquitectura técnica del sistema, sus componentes principales, patrones de diseño implementados y decisiones arquitectónicas tomadas durante el desarrollo.

## Visión General de la Arquitectura

### Arquitectura de Capas

El sistema implementa una arquitectura de capas bien definida que separa las responsabilidades y facilita el mantenimiento:

```
┌─────────────────────────────────────────────────────────────┐
│                    Capa de Presentación                     │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌────────┐ │
│  │   API POS   │ │  API Sync   │ │ API Labels  │ │API Rpts│ │
│  └─────────────┘ └─────────────┘ └─────────────┘ └────────┘ │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│                    Capa de Middleware                       │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌────────┐ │
│  │    Auth     │ │    CORS     │ │   Logging   │ │Metrics │ │
│  └─────────────┘ └─────────────┘ └─────────────┘ └────────┘ │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│                   Capa de Controladores                     │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌────────┐ │
│  │POS Controller│ │Sync Controller│ │Labels Ctrl│ │Rpts Ctrl│ │
│  └─────────────┘ └─────────────┘ └─────────────┘ └────────┘ │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│                    Capa de Servicios                        │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌────────┐ │
│  │ POS Service │ │Sync Service │ │Labels Service│ │Rpts Svc│ │
│  └─────────────┘ └─────────────┘ └─────────────┘ └────────┘ │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│                  Capa de Repositorios                       │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌────────┐ │
│  │POS Repository│ │Sync Repository│ │Labels Repo│ │Rpts Repo│ │
│  └─────────────┘ └─────────────┘ └─────────────┘ └────────┘ │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│                   Capa de Datos                             │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌────────┐ │
│  │  DB POS     │ │   DB Sync   │ │  DB Labels  │ │DB Rpts │ │
│  └─────────────┘ └─────────────┘ └─────────────┘ └────────┘ │
└─────────────────────────────────────────────────────────────┘
```

### Principios Arquitectónicos

#### 1. Separación de Responsabilidades (SoC)
Cada capa tiene una responsabilidad específica y bien definida. Los controladores manejan las peticiones HTTP, los servicios contienen la lógica de negocio, y los repositorios gestionan el acceso a datos.

#### 2. Inversión de Dependencias (DIP)
Las capas superiores dependen de abstracciones (interfaces) en lugar de implementaciones concretas. Esto facilita el testing y permite cambiar implementaciones sin afectar otras capas.

#### 3. Principio Abierto/Cerrado (OCP)
El sistema está diseñado para ser extensible sin modificar código existente. Nuevas funcionalidades se pueden agregar implementando interfaces existentes.

#### 4. Responsabilidad Única (SRP)
Cada componente tiene una única razón para cambiar. Los modelos representan datos, los servicios implementan lógica de negocio, y los controladores manejan HTTP.

## Componentes Principales

### 1. Sistema de Configuración

El sistema de configuración es uno de los componentes más críticos, diseñado para proporcionar flexibilidad y facilidad de mantenimiento.

#### ConfigManager
```go
type ConfigManager struct {
    config     *Config
    configPath string
    mutex      sync.RWMutex
    watchers   []func(*Config)
}
```

**Responsabilidades:**
- Cargar y validar configuraciones desde archivos YAML
- Gestionar actualizaciones de configuración en tiempo real
- Notificar cambios a componentes suscritos
- Mantener consistencia de configuración entre componentes

#### HotReloadManager
```go
type HotReloadManager struct {
    configManager *ConfigManager
    watcher       *fsnotify.Watcher
    watchedFiles  map[string]string
    reloadChan    chan ReloadEvent
    debounceTime  time.Duration
}
```

**Características:**
- Monitoreo automático de archivos de configuración
- Debounce inteligente para evitar recargas múltiples
- Validación previa antes de aplicar cambios
- Rollback automático en caso de configuración inválida

### 2. Sistema de Base de Datos

#### DatabaseManager
El DatabaseManager implementa un patrón de pool de conexiones diferenciado por API, optimizando el rendimiento según las características de cada API.

```go
type DatabaseManager struct {
    connections map[string]*sql.DB
    configs     DatabaseConfigs
    mutex       sync.RWMutex
}
```

**Optimizaciones por API:**
- **POS**: Pool grande (50 conexiones) para alta concurrencia
- **Sync**: Pool mediano (20 conexiones) para operaciones batch
- **Labels**: Pool pequeño (15 conexiones) para generación asíncrona
- **Reports**: Pool mediano (30 conexiones) para consultas complejas

#### BaseRepository
Implementa patrones comunes de acceso a datos:

```go
type BaseRepository struct {
    db          *sql.DB
    tableName   string
    primaryKey  string
}
```

**Funcionalidades:**
- CRUD completo con soft delete
- Paginación avanzada con filtros dinámicos
- Búsqueda de texto completo
- Operaciones en lote optimizadas
- Query builder flexible

### 3. Sistema de Middleware

#### Middleware de Autenticación
Implementa autenticación JWT con soporte para múltiples tipos de claims:

```go
type JWTClaims struct {
    UserID     uuid.UUID `json:"user_id"`
    Email      string    `json:"email"`
    Rol        string    `json:"rol"`
    SucursalID *uuid.UUID `json:"sucursal_id,omitempty"`
    TerminalID string    `json:"terminal_id,omitempty"`
    Permisos   []string  `json:"permisos"`
    jwt.RegisteredClaims
}
```

**Características:**
- Validación de tokens con múltiples niveles
- Autorización basada en roles y permisos
- Soporte para terminales específicos
- Refresh tokens automáticos

#### Middleware de CORS
Configuración CORS diferenciada por API:

```go
type CORSConfig struct {
    AllowOrigins     []string
    AllowMethods     []string
    AllowHeaders     []string
    ExposeHeaders    []string
    AllowCredentials bool
    MaxAge           int
}
```

#### Middleware de Logging
Sistema de logging estructurado con múltiples niveles:

```go
type LoggingConfig struct {
    Level          string
    Format         string
    Output         string
    LogRequests    bool
    LogResponses   bool
    SensitiveFields []string
}
```

### 4. Controladores de API

Cada API tiene su controlador especializado que maneja las peticiones HTTP específicas de su dominio.

#### POSController
Maneja operaciones críticas del punto de venta:
- Autenticación de usuarios y terminales
- Gestión de productos y stock
- Procesamiento de ventas
- Administración de clientes

#### SyncController
Gestiona la sincronización de datos:
- Iniciar/detener sincronizaciones
- Resolución de conflictos
- Monitoreo de estado
- Configuración de sincronización

#### LabelsController
Controla la generación de etiquetas:
- Gestión de plantillas
- Generación individual y por lotes
- Preview de etiquetas
- Descarga de archivos generados

#### ReportsController
Administra reportes y dashboards:
- Gestión de plantillas de reportes
- Generación de reportes programados
- Dashboards interactivos
- Exportación en múltiples formatos

## Patrones de Diseño Implementados

### 1. Repository Pattern
Abstrae el acceso a datos y proporciona una interfaz uniforme para operaciones de persistencia.

```go
type ProductoRepository interface {
    Create(producto *models.Producto) error
    GetByID(id uuid.UUID) (*models.Producto, error)
    Update(producto *models.Producto) error
    Delete(id uuid.UUID) error
    List(filters map[string]interface{}) ([]*models.Producto, error)
}
```

### 2. Service Layer Pattern
Encapsula la lógica de negocio y coordina operaciones entre múltiples repositorios.

```go
type POSService interface {
    ProcesarVenta(venta *models.VentaCreateDTO) (*models.Venta, error)
    ValidarStock(items []models.VentaItem) error
    AplicarDescuentos(venta *models.Venta) error
    GenerarRecibo(ventaID uuid.UUID) (*models.Recibo, error)
}
```

### 3. Factory Pattern
Utilizado para crear instancias de componentes con configuraciones específicas.

```go
func NewDatabaseConnection(config DatabaseConfig) (*sql.DB, error) {
    // Configuración específica según el tipo de API
}
```

### 4. Observer Pattern
Implementado en el sistema de configuración para notificar cambios.

```go
func (cm *ConfigManager) AddWatcher(watcher func(*Config)) {
    cm.watchers = append(cm.watchers, watcher)
}
```

### 5. Strategy Pattern
Utilizado en la resolución de conflictos de sincronización.

```go
type ConflictResolver interface {
    Resolve(conflict *models.ConflictoSincronizacion) (*models.ResolucionConflicto, error)
}

type ServerWinsResolver struct{}
type ClientWinsResolver struct{}
type ManualResolver struct{}
```

## Gestión de Concurrencia

### 1. Pool de Conexiones
Cada API tiene su propio pool de conexiones optimizado:

```go
type ConnectionPool struct {
    maxOpen     int
    maxIdle     int
    maxLifetime time.Duration
    idleTimeout time.Duration
}
```

### 2. Workers Concurrentes
Para operaciones asíncronas como generación de etiquetas y reportes:

```go
type WorkerPool struct {
    workers    int
    jobQueue   chan Job
    resultChan chan Result
    quit       chan bool
}
```

### 3. Mutex y RWMutex
Protección de recursos compartidos:

```go
type SafeCache struct {
    data  map[string]interface{}
    mutex sync.RWMutex
}
```

## Manejo de Errores

### 1. Errores Estructurados
```go
type APIError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details map[string]interface{} `json:"details,omitempty"`
}
```

### 2. Middleware de Recuperación
```go
func RecoveryMiddleware() gin.HandlerFunc {
    return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
        // Log del error y respuesta estructurada
    })
}
```

### 3. Validación de Entrada
```go
type Validator struct {
    validate *validator.Validate
}

func (v *Validator) ValidateStruct(s interface{}) error {
    return v.validate.Struct(s)
}
```

## Seguridad

### 1. Autenticación JWT
- Tokens con expiración configurable
- Refresh tokens para sesiones largas
- Claims personalizadas por contexto

### 2. Autorización Granular
- Roles jerárquicos (Admin > Gerente > Cajero)
- Permisos específicos por operación
- Validación a nivel de endpoint

### 3. Validación de Entrada
- Sanitización de datos
- Validación de tipos y rangos
- Protección contra inyección SQL

### 4. Rate Limiting
- Límites configurables por API
- Whitelist de IPs
- Burst control

## Monitoreo y Observabilidad

### 1. Métricas
```go
type Metrics struct {
    RequestCount    prometheus.CounterVec
    RequestDuration prometheus.HistogramVec
    ActiveUsers     prometheus.Gauge
}
```

### 2. Health Checks
```go
type HealthChecker struct {
    checks map[string]HealthCheck
}

type HealthCheck interface {
    Check() error
    Name() string
}
```

### 3. Logging Estructurado
```go
type LogEntry struct {
    Timestamp time.Time `json:"timestamp"`
    Level     string    `json:"level"`
    Message   string    `json:"message"`
    Fields    map[string]interface{} `json:"fields"`
}
```

## Escalabilidad

### 1. Horizontal Scaling
- APIs stateless
- Configuración externalizada
- Load balancing ready

### 2. Vertical Scaling
- Pools de conexiones ajustables
- Workers configurables
- Cache optimizable

### 3. Database Scaling
- Conexiones diferenciadas por API
- Read replicas support
- Connection pooling

## Testing

### 1. Unit Tests
```go
func TestProductoService_Create(t *testing.T) {
    // Mock repository
    mockRepo := &MockProductoRepository{}
    service := NewProductoService(mockRepo)
    
    // Test logic
}
```

### 2. Integration Tests
```go
func TestPOSAPI_CreateVenta(t *testing.T) {
    // Setup test database
    // Create test server
    // Execute HTTP requests
    // Verify results
}
```

### 3. Load Tests
```go
func BenchmarkPOSAPI_GetProductos(b *testing.B) {
    for i := 0; i < b.N; i++ {
        // Execute request
    }
}
```

## Deployment

### 1. Containerización
```dockerfile
FROM golang:1.19-alpine AS builder
# Build steps

FROM alpine:latest
# Runtime setup
```

### 2. Configuración de Producción
- Variables de entorno
- Secrets management
- Health checks
- Resource limits

### 3. Monitoring en Producción
- Prometheus metrics
- Grafana dashboards
- Alerting rules
- Log aggregation

## Consideraciones de Rendimiento

### 1. Database Optimization
- Índices optimizados
- Query optimization
- Connection pooling
- Prepared statements

### 2. Memory Management
- Object pooling
- Garbage collection tuning
- Memory profiling
- Leak detection

### 3. CPU Optimization
- Goroutine management
- CPU profiling
- Concurrent processing
- Algorithm optimization

## Conclusión

La arquitectura del Servidor Central Ferre-POS está diseñada para ser robusta, escalable y mantenible. Los patrones de diseño implementados, junto con las mejores prácticas de Go, proporcionan una base sólida para el crecimiento futuro del sistema.

La separación clara de responsabilidades, el sistema de configuración flexible y la arquitectura modular permiten que el sistema evolucione de manera controlada y predecible, facilitando tanto el mantenimiento como la adición de nuevas funcionalidades.

---

**Autor**: Manus AI  
**Versión**: 1.0.0  
**Fecha**: 2024-01-XX

