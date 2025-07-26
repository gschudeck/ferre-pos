// Package metrics proporciona funcionalidades avanzadas de métricas
// con notación húngara y soporte para Prometheus
package metrics

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// Config contiene la configuración de métricas
type Config struct {
	Enabled     bool   `yaml:"enabled" json:"enabled"`
	Namespace   string `yaml:"namespace" json:"namespace"`
	Subsystem   string `yaml:"subsystem" json:"subsystem"`
	MetricsPath string `yaml:"metrics_path" json:"metrics_path"`
}

// structManager gestiona todas las métricas del sistema
type structManager struct {
	ptrConfig    *Config
	ptrRegistry  *prometheus.Registry
	ptrLogger    *zap.Logger
	mutexMetrics sync.RWMutex

	// Métricas HTTP
	ptrHTTPRequestsTotal   *prometheus.CounterVec
	ptrHTTPRequestDuration *prometheus.HistogramVec
	ptrHTTPRequestSize     *prometheus.HistogramVec
	ptrHTTPResponseSize    *prometheus.HistogramVec
	ptrHTTPActiveRequests  prometheus.Gauge

	// Métricas de base de datos
	ptrDBConnectionsActive prometheus.Gauge
	ptrDBConnectionsIdle   prometheus.Gauge
	ptrDBConnectionsMax    prometheus.Gauge
	ptrDBQueryDuration     *prometheus.HistogramVec
	ptrDBQueriesTotal      *prometheus.CounterVec

	// Métricas de negocio
	ptrBusinessOperationsTotal   *prometheus.CounterVec
	ptrBusinessOperationDuration *prometheus.HistogramVec
	ptrBusinessErrors            *prometheus.CounterVec

	// Métricas de sistema
	ptrSystemMemoryUsage prometheus.Gauge
	ptrSystemCPUUsage    prometheus.Gauge
	ptrSystemGoroutines  prometheus.Gauge
	ptrSystemGCDuration  prometheus.Histogram

	// Métricas de cache
	ptrCacheHits      *prometheus.CounterVec
	ptrCacheMisses    *prometheus.CounterVec
	ptrCacheSize      prometheus.Gauge
	ptrCacheEvictions *prometheus.CounterVec

	// Métricas de autenticación
	ptrAuthAttempts    *prometheus.CounterVec
	ptrAuthFailures    *prometheus.CounterVec
	ptrActiveSessions  prometheus.Gauge
	ptrSessionDuration prometheus.Histogram
}

// interfaceManager define la interfaz del gestor de métricas
type interfaceManager interface {
	// Métricas HTTP
	RecordHTTPRequest(strMethod, strPath, strStatus string, fltDuration float64, intRequestSize, intResponseSize int64)
	IncActiveRequests()
	DecActiveRequests()

	// Métricas de base de datos
	RecordDBQuery(strOperation string, fltDuration float64, boolSuccess bool)
	SetDBConnections(intActive, intIdle, intMax int)

	// Métricas de negocio
	RecordBusinessOperation(strOperation, strResult string, fltDuration float64)
	RecordBusinessError(strOperation, strErrorType string)

	// Métricas de sistema
	SetSystemMetrics(fltMemory, fltCPU float64, intGoroutines int)
	RecordGCDuration(fltDuration float64)

	// Métricas de cache
	RecordCacheHit(strCacheType string)
	RecordCacheMiss(strCacheType string)
	SetCacheSize(fltSize float64)
	RecordCacheEviction(strCacheType, strReason string)

	// Métricas de autenticación
	RecordAuthAttempt(strMethod, strResult string)
	RecordAuthFailure(strMethod, strReason string)
	SetActiveSessions(intCount int)
	RecordSessionDuration(fltDuration float64)

	// Gestión
	Handler() gin.HandlerFunc
	Close() error
	GetRegistry() *prometheus.Registry
}

// NewManager crea un nuevo gestor de métricas
func NewManager(ptrConfig *Config) (*structManager, error) {
	if ptrConfig == nil {
		return nil, fmt.Errorf("configuración de métricas no puede ser nil")
	}

	if !ptrConfig.Enabled {
		return &structManager{ptrConfig: ptrConfig}, nil
	}

	// Crear registry personalizado
	ptrRegistry := prometheus.NewRegistry()

	ptrManager := &structManager{
		ptrConfig:   ptrConfig,
		ptrRegistry: ptrRegistry,
	}

	// Inicializar métricas
	if err := ptrManager.initializeMetrics(); err != nil {
		return nil, fmt.Errorf("error inicializando métricas: %w", err)
	}

	// Registrar métricas
	if err := ptrManager.registerMetrics(); err != nil {
		return nil, fmt.Errorf("error registrando métricas: %w", err)
	}

	return ptrManager, nil
}

// initializeMetrics inicializa todas las métricas
func (ptrManager *structManager) initializeMetrics() error {
	strNamespace := ptrManager.ptrConfig.Namespace
	strSubsystem := ptrManager.ptrConfig.Subsystem

	// Métricas HTTP
	ptrManager.ptrHTTPRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: strNamespace,
			Subsystem: strSubsystem,
			Name:      "http_requests_total",
			Help:      "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	ptrManager.ptrHTTPRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: strNamespace,
			Subsystem: strSubsystem,
			Name:      "http_request_duration_seconds",
			Help:      "HTTP request duration in seconds",
			Buckets:   []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"method", "path", "status"},
	)

	ptrManager.ptrHTTPRequestSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: strNamespace,
			Subsystem: strSubsystem,
			Name:      "http_request_size_bytes",
			Help:      "HTTP request size in bytes",
			Buckets:   prometheus.ExponentialBuckets(100, 10, 8),
		},
		[]string{"method", "path"},
	)

	ptrManager.ptrHTTPResponseSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: strNamespace,
			Subsystem: strSubsystem,
			Name:      "http_response_size_bytes",
			Help:      "HTTP response size in bytes",
			Buckets:   prometheus.ExponentialBuckets(100, 10, 8),
		},
		[]string{"method", "path", "status"},
	)

	ptrManager.ptrHTTPActiveRequests = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: strNamespace,
			Subsystem: strSubsystem,
			Name:      "http_active_requests",
			Help:      "Number of active HTTP requests",
		},
	)

	// Métricas de base de datos
	ptrManager.ptrDBConnectionsActive = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: strNamespace,
			Subsystem: strSubsystem,
			Name:      "db_connections_active",
			Help:      "Number of active database connections",
		},
	)

	ptrManager.ptrDBConnectionsIdle = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: strNamespace,
			Subsystem: strSubsystem,
			Name:      "db_connections_idle",
			Help:      "Number of idle database connections",
		},
	)

	ptrManager.ptrDBConnectionsMax = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: strNamespace,
			Subsystem: strSubsystem,
			Name:      "db_connections_max",
			Help:      "Maximum number of database connections",
		},
	)

	ptrManager.ptrDBQueryDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: strNamespace,
			Subsystem: strSubsystem,
			Name:      "db_query_duration_seconds",
			Help:      "Database query duration in seconds",
			Buckets:   []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
		},
		[]string{"operation", "success"},
	)

	ptrManager.ptrDBQueriesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: strNamespace,
			Subsystem: strSubsystem,
			Name:      "db_queries_total",
			Help:      "Total number of database queries",
		},
		[]string{"operation", "success"},
	)

	// Métricas de negocio
	ptrManager.ptrBusinessOperationsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: strNamespace,
			Subsystem: strSubsystem,
			Name:      "business_operations_total",
			Help:      "Total number of business operations",
		},
		[]string{"operation", "result"},
	)

	ptrManager.ptrBusinessOperationDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: strNamespace,
			Subsystem: strSubsystem,
			Name:      "business_operation_duration_seconds",
			Help:      "Business operation duration in seconds",
			Buckets:   []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30},
		},
		[]string{"operation", "result"},
	)

	ptrManager.ptrBusinessErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: strNamespace,
			Subsystem: strSubsystem,
			Name:      "business_errors_total",
			Help:      "Total number of business errors",
		},
		[]string{"operation", "error_type"},
	)

	// Métricas de sistema
	ptrManager.ptrSystemMemoryUsage = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: strNamespace,
			Subsystem: strSubsystem,
			Name:      "system_memory_usage_bytes",
			Help:      "System memory usage in bytes",
		},
	)

	ptrManager.ptrSystemCPUUsage = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: strNamespace,
			Subsystem: strSubsystem,
			Name:      "system_cpu_usage_percent",
			Help:      "System CPU usage percentage",
		},
	)

	ptrManager.ptrSystemGoroutines = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: strNamespace,
			Subsystem: strSubsystem,
			Name:      "system_goroutines",
			Help:      "Number of goroutines",
		},
	)

	ptrManager.ptrSystemGCDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: strNamespace,
			Subsystem: strSubsystem,
			Name:      "system_gc_duration_seconds",
			Help:      "Garbage collection duration in seconds",
			Buckets:   []float64{0.0001, 0.0005, 0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1},
		},
	)

	// Métricas de cache
	ptrManager.ptrCacheHits = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: strNamespace,
			Subsystem: strSubsystem,
			Name:      "cache_hits_total",
			Help:      "Total number of cache hits",
		},
		[]string{"cache_type"},
	)

	ptrManager.ptrCacheMisses = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: strNamespace,
			Subsystem: strSubsystem,
			Name:      "cache_misses_total",
			Help:      "Total number of cache misses",
		},
		[]string{"cache_type"},
	)

	ptrManager.ptrCacheSize = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: strNamespace,
			Subsystem: strSubsystem,
			Name:      "cache_size_bytes",
			Help:      "Cache size in bytes",
		},
	)

	ptrManager.ptrCacheEvictions = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: strNamespace,
			Subsystem: strSubsystem,
			Name:      "cache_evictions_total",
			Help:      "Total number of cache evictions",
		},
		[]string{"cache_type", "reason"},
	)

	// Métricas de autenticación
	ptrManager.ptrAuthAttempts = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: strNamespace,
			Subsystem: strSubsystem,
			Name:      "auth_attempts_total",
			Help:      "Total number of authentication attempts",
		},
		[]string{"method", "result"},
	)

	ptrManager.ptrAuthFailures = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: strNamespace,
			Subsystem: strSubsystem,
			Name:      "auth_failures_total",
			Help:      "Total number of authentication failures",
		},
		[]string{"method", "reason"},
	)

	ptrManager.ptrActiveSessions = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: strNamespace,
			Subsystem: strSubsystem,
			Name:      "active_sessions",
			Help:      "Number of active user sessions",
		},
	)

	ptrManager.ptrSessionDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: strNamespace,
			Subsystem: strSubsystem,
			Name:      "session_duration_seconds",
			Help:      "User session duration in seconds",
			Buckets:   []float64{60, 300, 900, 1800, 3600, 7200, 14400, 28800, 86400},
		},
	)

	return nil
}

// registerMetrics registra todas las métricas en el registry
func (ptrManager *structManager) registerMetrics() error {
	arrMetrics := []prometheus.Collector{
		// HTTP
		ptrManager.ptrHTTPRequestsTotal,
		ptrManager.ptrHTTPRequestDuration,
		ptrManager.ptrHTTPRequestSize,
		ptrManager.ptrHTTPResponseSize,
		ptrManager.ptrHTTPActiveRequests,

		// Database
		ptrManager.ptrDBConnectionsActive,
		ptrManager.ptrDBConnectionsIdle,
		ptrManager.ptrDBConnectionsMax,
		ptrManager.ptrDBQueryDuration,
		ptrManager.ptrDBQueriesTotal,

		// Business
		ptrManager.ptrBusinessOperationsTotal,
		ptrManager.ptrBusinessOperationDuration,
		ptrManager.ptrBusinessErrors,

		// System
		ptrManager.ptrSystemMemoryUsage,
		ptrManager.ptrSystemCPUUsage,
		ptrManager.ptrSystemGoroutines,
		ptrManager.ptrSystemGCDuration,

		// Cache
		ptrManager.ptrCacheHits,
		ptrManager.ptrCacheMisses,
		ptrManager.ptrCacheSize,
		ptrManager.ptrCacheEvictions,

		// Auth
		ptrManager.ptrAuthAttempts,
		ptrManager.ptrAuthFailures,
		ptrManager.ptrActiveSessions,
		ptrManager.ptrSessionDuration,
	}

	for _, ptrMetric := range arrMetrics {
		if err := ptrManager.ptrRegistry.Register(ptrMetric); err != nil {
			return fmt.Errorf("error registrando métrica: %w", err)
		}
	}

	return nil
}

// RecordHTTPRequest registra una request HTTP
func (ptrManager *structManager) RecordHTTPRequest(strMethod, strPath, strStatus string, fltDuration float64, intRequestSize, intResponseSize int64) {
	if !ptrManager.ptrConfig.Enabled {
		return
	}

	ptrManager.mutexMetrics.Lock()
	defer ptrManager.mutexMetrics.Unlock()

	ptrManager.ptrHTTPRequestsTotal.WithLabelValues(strMethod, strPath, strStatus).Inc()
	ptrManager.ptrHTTPRequestDuration.WithLabelValues(strMethod, strPath, strStatus).Observe(fltDuration)
	ptrManager.ptrHTTPRequestSize.WithLabelValues(strMethod, strPath).Observe(float64(intRequestSize))
	ptrManager.ptrHTTPResponseSize.WithLabelValues(strMethod, strPath, strStatus).Observe(float64(intResponseSize))
}

// IncActiveRequests incrementa el contador de requests activas
func (ptrManager *structManager) IncActiveRequests() {
	if !ptrManager.ptrConfig.Enabled {
		return
	}
	ptrManager.ptrHTTPActiveRequests.Inc()
}

// DecActiveRequests decrementa el contador de requests activas
func (ptrManager *structManager) DecActiveRequests() {
	if !ptrManager.ptrConfig.Enabled {
		return
	}
	ptrManager.ptrHTTPActiveRequests.Dec()
}

// RecordDBQuery registra una query de base de datos
func (ptrManager *structManager) RecordDBQuery(strOperation string, fltDuration float64, boolSuccess bool) {
	if !ptrManager.ptrConfig.Enabled {
		return
	}

	strSuccess := strconv.FormatBool(boolSuccess)
	ptrManager.ptrDBQueriesTotal.WithLabelValues(strOperation, strSuccess).Inc()
	ptrManager.ptrDBQueryDuration.WithLabelValues(strOperation, strSuccess).Observe(fltDuration)
}

// SetDBConnections establece las métricas de conexiones de BD
func (ptrManager *structManager) SetDBConnections(intActive, intIdle, intMax int) {
	if !ptrManager.ptrConfig.Enabled {
		return
	}

	ptrManager.ptrDBConnectionsActive.Set(float64(intActive))
	ptrManager.ptrDBConnectionsIdle.Set(float64(intIdle))
	ptrManager.ptrDBConnectionsMax.Set(float64(intMax))
}

// RecordBusinessOperation registra una operación de negocio
func (ptrManager *structManager) RecordBusinessOperation(strOperation, strResult string, fltDuration float64) {
	if !ptrManager.ptrConfig.Enabled {
		return
	}

	ptrManager.ptrBusinessOperationsTotal.WithLabelValues(strOperation, strResult).Inc()
	ptrManager.ptrBusinessOperationDuration.WithLabelValues(strOperation, strResult).Observe(fltDuration)
}

// RecordBusinessError registra un error de negocio
func (ptrManager *structManager) RecordBusinessError(strOperation, strErrorType string) {
	if !ptrManager.ptrConfig.Enabled {
		return
	}

	ptrManager.ptrBusinessErrors.WithLabelValues(strOperation, strErrorType).Inc()
}

// SetSystemMetrics establece las métricas del sistema
func (ptrManager *structManager) SetSystemMetrics(fltMemory, fltCPU float64, intGoroutines int) {
	if !ptrManager.ptrConfig.Enabled {
		return
	}

	ptrManager.ptrSystemMemoryUsage.Set(fltMemory)
	ptrManager.ptrSystemCPUUsage.Set(fltCPU)
	ptrManager.ptrSystemGoroutines.Set(float64(intGoroutines))
}

// RecordGCDuration registra la duración del garbage collection
func (ptrManager *structManager) RecordGCDuration(fltDuration float64) {
	if !ptrManager.ptrConfig.Enabled {
		return
	}

	ptrManager.ptrSystemGCDuration.Observe(fltDuration)
}

// RecordCacheHit registra un cache hit
func (ptrManager *structManager) RecordCacheHit(strCacheType string) {
	if !ptrManager.ptrConfig.Enabled {
		return
	}

	ptrManager.ptrCacheHits.WithLabelValues(strCacheType).Inc()
}

// RecordCacheMiss registra un cache miss
func (ptrManager *structManager) RecordCacheMiss(strCacheType string) {
	if !ptrManager.ptrConfig.Enabled {
		return
	}

	ptrManager.ptrCacheMisses.WithLabelValues(strCacheType).Inc()
}

// SetCacheSize establece el tamaño del cache
func (ptrManager *structManager) SetCacheSize(fltSize float64) {
	if !ptrManager.ptrConfig.Enabled {
		return
	}

	ptrManager.ptrCacheSize.Set(fltSize)
}

// RecordCacheEviction registra una evicción de cache
func (ptrManager *structManager) RecordCacheEviction(strCacheType, strReason string) {
	if !ptrManager.ptrConfig.Enabled {
		return
	}

	ptrManager.ptrCacheEvictions.WithLabelValues(strCacheType, strReason).Inc()
}

// RecordAuthAttempt registra un intento de autenticación
func (ptrManager *structManager) RecordAuthAttempt(strMethod, strResult string) {
	if !ptrManager.ptrConfig.Enabled {
		return
	}

	ptrManager.ptrAuthAttempts.WithLabelValues(strMethod, strResult).Inc()
}

// RecordAuthFailure registra un fallo de autenticación
func (ptrManager *structManager) RecordAuthFailure(strMethod, strReason string) {
	if !ptrManager.ptrConfig.Enabled {
		return
	}

	ptrManager.ptrAuthFailures.WithLabelValues(strMethod, strReason).Inc()
}

// SetActiveSessions establece el número de sesiones activas
func (ptrManager *structManager) SetActiveSessions(intCount int) {
	if !ptrManager.ptrConfig.Enabled {
		return
	}

	ptrManager.ptrActiveSessions.Set(float64(intCount))
}

// RecordSessionDuration registra la duración de una sesión
func (ptrManager *structManager) RecordSessionDuration(fltDuration float64) {
	if !ptrManager.ptrConfig.Enabled {
		return
	}

	ptrManager.ptrSessionDuration.Observe(fltDuration)
}

// Handler retorna el handler de métricas para Gin
func (ptrManager *structManager) Handler() gin.HandlerFunc {
	if !ptrManager.ptrConfig.Enabled {
		return func(ptrCtx *gin.Context) {
			ptrCtx.String(http.StatusNotFound, "Metrics disabled")
		}
	}

	ptrHandler := promhttp.HandlerFor(ptrManager.ptrRegistry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
		Timeout:           30 * time.Second,
	})

	return gin.WrapH(ptrHandler)
}

// Close cierra el gestor de métricas
func (ptrManager *structManager) Close() error {
	// No hay recursos específicos que cerrar en Prometheus
	return nil
}

// GetRegistry retorna el registry de Prometheus
func (ptrManager *structManager) GetRegistry() *prometheus.Registry {
	return ptrManager.ptrRegistry
}
