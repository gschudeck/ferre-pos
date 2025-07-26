// Package metrics proporciona métricas avanzadas para Prometheus
// con notación húngara y métricas específicas de negocio
package metrics

import (
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"ferre-pos-servidor-central/pkg/logger"
)

// InterfaceManagerImproved interfaz principal del manager de métricas mejorado
type InterfaceManagerImproved interface {
	// Métricas HTTP
	RecordHTTPRequest(strMethod, strPath, strStatusCode string, fltDuration float64, int64RequestSize, int64ResponseSize int64)
	IncActiveRequests()
	DecActiveRequests()

	// Métricas de base de datos
	RecordDatabaseQuery(strOperation, strTable string, fltDuration float64, boolSuccess bool)
	IncDatabaseConnections(strPool string)
	DecDatabaseConnections(strPool string)
	RecordDatabaseConnectionDuration(strPool string, fltDuration float64)

	// Métricas de negocio
	RecordBusinessOperation(strOperation, strResult string, fltDuration float64)
	RecordBusinessError(strOperation, strErrorType string)
	IncBusinessCounter(strMetric, strLabel string)
	SetBusinessGauge(strMetric string, fltValue float64, mapLabels map[string]string)

	// Métricas de sistema
	RecordMemoryUsage(fltHeapMB, fltStackMB float64)
	RecordGoroutines(intCount int)
	RecordGCDuration(fltDuration float64)

	// Métricas de cache
	RecordCacheOperation(strOperation, strResult string)
	SetCacheSize(strCacheName string, intSize int)
	RecordCacheHitRate(strCacheName string, fltHitRate float64)

	// Métricas de autenticación
	RecordAuthAttempt(strMethod, strResult string)
	RecordSessionDuration(fltDuration float64)
	IncActiveUsers()
	DecActiveUsers()

	// Métricas de rate limiting
	RecordRateLimit(strKey string, boolAllowed bool, intRemaining int)

	// Métricas específicas de APIs
	RecordPOSOperation(strOperation, strResult string, fltDuration float64)
	RecordSyncOperation(strOperation, strResult string, fltDuration float64, intRecords int)
	RecordLabelGeneration(strType, strResult string, fltDuration float64, intCount int)
	RecordReportGeneration(strType, strResult string, fltDuration float64, intSize int64)

	// Utilidades
	GetHandler() http.Handler
	GetRegistry() *prometheus.Registry
	Reset()
}

// structMetricsManager implementación del manager de métricas
type structMetricsManager struct {
	PtrRegistry  *prometheus.Registry
	PtrLogger    logger.InterfaceLogger
	MutexMetrics sync.RWMutex

	// Métricas HTTP
	PtrHTTPRequestsTotal   *prometheus.CounterVec
	PtrHTTPRequestDuration *prometheus.HistogramVec
	PtrHTTPRequestSize     *prometheus.HistogramVec
	PtrHTTPResponseSize    *prometheus.HistogramVec
	PtrHTTPActiveRequests  prometheus.Gauge

	// Métricas de base de datos
	PtrDBQueriesTotal       *prometheus.CounterVec
	PtrDBQueryDuration      *prometheus.HistogramVec
	PtrDBConnectionsActive  *prometheus.GaugeVec
	PtrDBConnectionDuration *prometheus.HistogramVec
	PtrDBTransactionsTotal  *prometheus.CounterVec

	// Métricas de negocio
	PtrBusinessOperationsTotal   *prometheus.CounterVec
	PtrBusinessOperationDuration *prometheus.HistogramVec
	PtrBusinessErrorsTotal       *prometheus.CounterVec
	PtrBusinessCounters          *prometheus.CounterVec
	PtrBusinessGauges            *prometheus.GaugeVec

	// Métricas de sistema
	PtrSystemMemoryUsage *prometheus.GaugeVec
	PtrSystemGoroutines  prometheus.Gauge
	PtrSystemGCDuration  prometheus.Histogram
	PtrSystemCPUUsage    prometheus.Gauge
	PtrSystemDiskUsage   *prometheus.GaugeVec

	// Métricas de cache
	PtrCacheOperationsTotal *prometheus.CounterVec
	PtrCacheSize            *prometheus.GaugeVec
	PtrCacheHitRate         *prometheus.GaugeVec
	PtrCacheEvictions       *prometheus.CounterVec

	// Métricas de autenticación
	PtrAuthAttemptsTotal   *prometheus.CounterVec
	PtrAuthSessionDuration prometheus.Histogram
	PtrAuthActiveUsers     prometheus.Gauge
	PtrAuthTokensIssued    *prometheus.CounterVec

	// Métricas de rate limiting
	PtrRateLimitTotal      *prometheus.CounterVec
	PtrRateLimitRemaining  *prometheus.GaugeVec
	PtrRateLimitViolations *prometheus.CounterVec

	// Métricas específicas de APIs
	PtrPOSOperationsTotal      *prometheus.CounterVec
	PtrPOSOperationDuration    *prometheus.HistogramVec
	PtrPOSVentasTotal          *prometheus.CounterVec
	PtrPOSProductosConsultados *prometheus.CounterVec

	PtrSyncOperationsTotal   *prometheus.CounterVec
	PtrSyncOperationDuration *prometheus.HistogramVec
	PtrSyncRecordsProcessed  *prometheus.CounterVec
	PtrSyncConflictsTotal    *prometheus.CounterVec

	PtrLabelsGeneratedTotal     *prometheus.CounterVec
	PtrLabelsGenerationDuration *prometheus.HistogramVec
	PtrLabelsTemplatesUsed      *prometheus.CounterVec

	PtrReportsGeneratedTotal     *prometheus.CounterVec
	PtrReportsGenerationDuration *prometheus.HistogramVec
	PtrReportsSize               *prometheus.HistogramVec

	// Métricas de infraestructura
	PtrInfraHealthChecks   *prometheus.CounterVec
	PtrInfraServiceUptime  *prometheus.GaugeVec
	PtrInfraNetworkLatency *prometheus.HistogramVec
}

// NewManagerImproved crea un nuevo manager de métricas mejorado
func NewManagerImproved(ptrLogger logger.InterfaceLogger) InterfaceManagerImproved {
	ptrRegistry := prometheus.NewRegistry()

	ptrManager := &structMetricsManager{
		PtrRegistry: ptrRegistry,
		PtrLogger:   ptrLogger,
	}

	ptrManager.initializeMetrics()
	ptrManager.registerMetrics()

	return ptrManager
}

// initializeMetrics inicializa todas las métricas
func (ptrManager *structMetricsManager) initializeMetrics() {
	// Métricas HTTP
	ptrManager.PtrHTTPRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ferre_pos_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status_code", "api"},
	)

	ptrManager.PtrHTTPRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ferre_pos_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"method", "path", "status_code", "api"},
	)

	ptrManager.PtrHTTPRequestSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ferre_pos_http_request_size_bytes",
			Help:    "HTTP request size in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 10, 8),
		},
		[]string{"method", "path", "api"},
	)

	ptrManager.PtrHTTPResponseSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ferre_pos_http_response_size_bytes",
			Help:    "HTTP response size in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 10, 8),
		},
		[]string{"method", "path", "status_code", "api"},
	)

	ptrManager.PtrHTTPActiveRequests = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "ferre_pos_http_active_requests",
			Help: "Number of active HTTP requests",
		},
	)

	// Métricas de base de datos
	ptrManager.PtrDBQueriesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ferre_pos_database_queries_total",
			Help: "Total number of database queries",
		},
		[]string{"operation", "table", "result", "pool"},
	)

	ptrManager.PtrDBQueryDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ferre_pos_database_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2, 5},
		},
		[]string{"operation", "table", "pool"},
	)

	ptrManager.PtrDBConnectionsActive = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ferre_pos_database_connections_active",
			Help: "Number of active database connections",
		},
		[]string{"pool", "database"},
	)

	ptrManager.PtrDBConnectionDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ferre_pos_database_connection_duration_seconds",
			Help:    "Database connection duration in seconds",
			Buckets: []float64{0.001, 0.01, 0.1, 1, 10, 60, 300},
		},
		[]string{"pool", "database"},
	)

	ptrManager.PtrDBTransactionsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ferre_pos_database_transactions_total",
			Help: "Total number of database transactions",
		},
		[]string{"result", "pool"},
	)

	// Métricas de negocio
	ptrManager.PtrBusinessOperationsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ferre_pos_business_operations_total",
			Help: "Total number of business operations",
		},
		[]string{"operation", "result", "api"},
	)

	ptrManager.PtrBusinessOperationDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ferre_pos_business_operation_duration_seconds",
			Help:    "Business operation duration in seconds",
			Buckets: []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30},
		},
		[]string{"operation", "api"},
	)

	ptrManager.PtrBusinessErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ferre_pos_business_errors_total",
			Help: "Total number of business errors",
		},
		[]string{"operation", "error_type", "api"},
	)

	ptrManager.PtrBusinessCounters = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ferre_pos_business_counters_total",
			Help: "Custom business counters",
		},
		[]string{"metric", "label", "api"},
	)

	ptrManager.PtrBusinessGauges = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ferre_pos_business_gauges",
			Help: "Custom business gauges",
		},
		[]string{"metric", "api", "sucursal", "categoria"},
	)

	// Métricas de sistema
	ptrManager.PtrSystemMemoryUsage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ferre_pos_system_memory_usage_mb",
			Help: "System memory usage in MB",
		},
		[]string{"type"},
	)

	ptrManager.PtrSystemGoroutines = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "ferre_pos_system_goroutines",
			Help: "Number of goroutines",
		},
	)

	ptrManager.PtrSystemGCDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "ferre_pos_system_gc_duration_seconds",
			Help:    "Garbage collection duration in seconds",
			Buckets: []float64{0.0001, 0.0005, 0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1},
		},
	)

	ptrManager.PtrSystemCPUUsage = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "ferre_pos_system_cpu_usage_percent",
			Help: "System CPU usage percentage",
		},
	)

	ptrManager.PtrSystemDiskUsage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ferre_pos_system_disk_usage_percent",
			Help: "System disk usage percentage",
		},
		[]string{"mount_point"},
	)

	// Métricas de cache
	ptrManager.PtrCacheOperationsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ferre_pos_cache_operations_total",
			Help: "Total number of cache operations",
		},
		[]string{"operation", "result", "cache_name"},
	)

	ptrManager.PtrCacheSize = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ferre_pos_cache_size_items",
			Help: "Number of items in cache",
		},
		[]string{"cache_name"},
	)

	ptrManager.PtrCacheHitRate = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ferre_pos_cache_hit_rate_percent",
			Help: "Cache hit rate percentage",
		},
		[]string{"cache_name"},
	)

	ptrManager.PtrCacheEvictions = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ferre_pos_cache_evictions_total",
			Help: "Total number of cache evictions",
		},
		[]string{"cache_name", "reason"},
	)

	// Métricas de autenticación
	ptrManager.PtrAuthAttemptsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ferre_pos_auth_attempts_total",
			Help: "Total number of authentication attempts",
		},
		[]string{"method", "result"},
	)

	ptrManager.PtrAuthSessionDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "ferre_pos_auth_session_duration_seconds",
			Help:    "Authentication session duration in seconds",
			Buckets: []float64{60, 300, 900, 1800, 3600, 7200, 14400, 28800, 86400},
		},
	)

	ptrManager.PtrAuthActiveUsers = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "ferre_pos_auth_active_users",
			Help: "Number of active authenticated users",
		},
	)

	ptrManager.PtrAuthTokensIssued = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ferre_pos_auth_tokens_issued_total",
			Help: "Total number of authentication tokens issued",
		},
		[]string{"token_type"},
	)

	// Métricas de rate limiting
	ptrManager.PtrRateLimitTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ferre_pos_rate_limit_requests_total",
			Help: "Total number of rate limit checks",
		},
		[]string{"key", "result"},
	)

	ptrManager.PtrRateLimitRemaining = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ferre_pos_rate_limit_remaining",
			Help: "Remaining rate limit quota",
		},
		[]string{"key"},
	)

	ptrManager.PtrRateLimitViolations = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ferre_pos_rate_limit_violations_total",
			Help: "Total number of rate limit violations",
		},
		[]string{"key", "reason"},
	)

	// Métricas específicas de API POS
	ptrManager.PtrPOSOperationsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ferre_pos_pos_operations_total",
			Help: "Total number of POS operations",
		},
		[]string{"operation", "result", "sucursal"},
	)

	ptrManager.PtrPOSOperationDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ferre_pos_pos_operation_duration_seconds",
			Help:    "POS operation duration in seconds",
			Buckets: []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1, 2, 5},
		},
		[]string{"operation", "sucursal"},
	)

	ptrManager.PtrPOSVentasTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ferre_pos_pos_ventas_total",
			Help: "Total number of sales processed",
		},
		[]string{"sucursal", "medio_pago"},
	)

	ptrManager.PtrPOSProductosConsultados = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ferre_pos_pos_productos_consultados_total",
			Help: "Total number of product lookups",
		},
		[]string{"sucursal", "tipo_busqueda"},
	)

	// Métricas específicas de API Sync
	ptrManager.PtrSyncOperationsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ferre_pos_sync_operations_total",
			Help: "Total number of sync operations",
		},
		[]string{"operation", "result", "sucursal"},
	)

	ptrManager.PtrSyncOperationDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ferre_pos_sync_operation_duration_seconds",
			Help:    "Sync operation duration in seconds",
			Buckets: []float64{0.1, 0.5, 1, 2, 5, 10, 30, 60, 300},
		},
		[]string{"operation", "sucursal"},
	)

	ptrManager.PtrSyncRecordsProcessed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ferre_pos_sync_records_processed_total",
			Help: "Total number of records processed in sync",
		},
		[]string{"entity_type", "operation", "sucursal"},
	)

	ptrManager.PtrSyncConflictsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ferre_pos_sync_conflicts_total",
			Help: "Total number of sync conflicts",
		},
		[]string{"entity_type", "resolution", "sucursal"},
	)

	// Métricas específicas de API Labels
	ptrManager.PtrLabelsGeneratedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ferre_pos_labels_generated_total",
			Help: "Total number of labels generated",
		},
		[]string{"type", "result", "sucursal"},
	)

	ptrManager.PtrLabelsGenerationDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ferre_pos_labels_generation_duration_seconds",
			Help:    "Label generation duration in seconds",
			Buckets: []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1, 2, 5, 10},
		},
		[]string{"type", "sucursal"},
	)

	ptrManager.PtrLabelsTemplatesUsed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ferre_pos_labels_templates_used_total",
			Help: "Total number of label templates used",
		},
		[]string{"template_id", "sucursal"},
	)

	// Métricas específicas de API Reports
	ptrManager.PtrReportsGeneratedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ferre_pos_reports_generated_total",
			Help: "Total number of reports generated",
		},
		[]string{"type", "result", "sucursal"},
	)

	ptrManager.PtrReportsGenerationDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ferre_pos_reports_generation_duration_seconds",
			Help:    "Report generation duration in seconds",
			Buckets: []float64{0.1, 0.5, 1, 2, 5, 10, 30, 60, 300, 600},
		},
		[]string{"type", "sucursal"},
	)

	ptrManager.PtrReportsSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ferre_pos_reports_size_bytes",
			Help:    "Report size in bytes",
			Buckets: prometheus.ExponentialBuckets(1024, 10, 8),
		},
		[]string{"type", "format", "sucursal"},
	)

	// Métricas de infraestructura
	ptrManager.PtrInfraHealthChecks = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ferre_pos_infra_health_checks_total",
			Help: "Total number of health checks",
		},
		[]string{"service", "result"},
	)

	ptrManager.PtrInfraServiceUptime = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ferre_pos_infra_service_uptime_seconds",
			Help: "Service uptime in seconds",
		},
		[]string{"service", "version"},
	)

	ptrManager.PtrInfraNetworkLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ferre_pos_infra_network_latency_seconds",
			Help:    "Network latency in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
		},
		[]string{"target", "protocol"},
	)
}

// registerMetrics registra todas las métricas en el registry
func (ptrManager *structMetricsManager) registerMetrics() {
	// HTTP metrics
	ptrManager.PtrRegistry.MustRegister(ptrManager.PtrHTTPRequestsTotal)
	ptrManager.PtrRegistry.MustRegister(ptrManager.PtrHTTPRequestDuration)
	ptrManager.PtrRegistry.MustRegister(ptrManager.PtrHTTPRequestSize)
	ptrManager.PtrRegistry.MustRegister(ptrManager.PtrHTTPResponseSize)
	ptrManager.PtrRegistry.MustRegister(ptrManager.PtrHTTPActiveRequests)

	// Database metrics
	ptrManager.PtrRegistry.MustRegister(ptrManager.PtrDBQueriesTotal)
	ptrManager.PtrRegistry.MustRegister(ptrManager.PtrDBQueryDuration)
	ptrManager.PtrRegistry.MustRegister(ptrManager.PtrDBConnectionsActive)
	ptrManager.PtrRegistry.MustRegister(ptrManager.PtrDBConnectionDuration)
	ptrManager.PtrRegistry.MustRegister(ptrManager.PtrDBTransactionsTotal)

	// Business metrics
	ptrManager.PtrRegistry.MustRegister(ptrManager.PtrBusinessOperationsTotal)
	ptrManager.PtrRegistry.MustRegister(ptrManager.PtrBusinessOperationDuration)
	ptrManager.PtrRegistry.MustRegister(ptrManager.PtrBusinessErrorsTotal)
	ptrManager.PtrRegistry.MustRegister(ptrManager.PtrBusinessCounters)
	ptrManager.PtrRegistry.MustRegister(ptrManager.PtrBusinessGauges)

	// System metrics
	ptrManager.PtrRegistry.MustRegister(ptrManager.PtrSystemMemoryUsage)
	ptrManager.PtrRegistry.MustRegister(ptrManager.PtrSystemGoroutines)
	ptrManager.PtrRegistry.MustRegister(ptrManager.PtrSystemGCDuration)
	ptrManager.PtrRegistry.MustRegister(ptrManager.PtrSystemCPUUsage)
	ptrManager.PtrRegistry.MustRegister(ptrManager.PtrSystemDiskUsage)

	// Cache metrics
	ptrManager.PtrRegistry.MustRegister(ptrManager.PtrCacheOperationsTotal)
	ptrManager.PtrRegistry.MustRegister(ptrManager.PtrCacheSize)
	ptrManager.PtrRegistry.MustRegister(ptrManager.PtrCacheHitRate)
	ptrManager.PtrRegistry.MustRegister(ptrManager.PtrCacheEvictions)

	// Auth metrics
	ptrManager.PtrRegistry.MustRegister(ptrManager.PtrAuthAttemptsTotal)
	ptrManager.PtrRegistry.MustRegister(ptrManager.PtrAuthSessionDuration)
	ptrManager.PtrRegistry.MustRegister(ptrManager.PtrAuthActiveUsers)
	ptrManager.PtrRegistry.MustRegister(ptrManager.PtrAuthTokensIssued)

	// Rate limit metrics
	ptrManager.PtrRegistry.MustRegister(ptrManager.PtrRateLimitTotal)
	ptrManager.PtrRegistry.MustRegister(ptrManager.PtrRateLimitRemaining)
	ptrManager.PtrRegistry.MustRegister(ptrManager.PtrRateLimitViolations)

	// POS API metrics
	ptrManager.PtrRegistry.MustRegister(ptrManager.PtrPOSOperationsTotal)
	ptrManager.PtrRegistry.MustRegister(ptrManager.PtrPOSOperationDuration)
	ptrManager.PtrRegistry.MustRegister(ptrManager.PtrPOSVentasTotal)
	ptrManager.PtrRegistry.MustRegister(ptrManager.PtrPOSProductosConsultados)

	// Sync API metrics
	ptrManager.PtrRegistry.MustRegister(ptrManager.PtrSyncOperationsTotal)
	ptrManager.PtrRegistry.MustRegister(ptrManager.PtrSyncOperationDuration)
	ptrManager.PtrRegistry.MustRegister(ptrManager.PtrSyncRecordsProcessed)
	ptrManager.PtrRegistry.MustRegister(ptrManager.PtrSyncConflictsTotal)

	// Labels API metrics
	ptrManager.PtrRegistry.MustRegister(ptrManager.PtrLabelsGeneratedTotal)
	ptrManager.PtrRegistry.MustRegister(ptrManager.PtrLabelsGenerationDuration)
	ptrManager.PtrRegistry.MustRegister(ptrManager.PtrLabelsTemplatesUsed)

	// Reports API metrics
	ptrManager.PtrRegistry.MustRegister(ptrManager.PtrReportsGeneratedTotal)
	ptrManager.PtrRegistry.MustRegister(ptrManager.PtrReportsGenerationDuration)
	ptrManager.PtrRegistry.MustRegister(ptrManager.PtrReportsSize)

	// Infrastructure metrics
	ptrManager.PtrRegistry.MustRegister(ptrManager.PtrInfraHealthChecks)
	ptrManager.PtrRegistry.MustRegister(ptrManager.PtrInfraServiceUptime)
	ptrManager.PtrRegistry.MustRegister(ptrManager.PtrInfraNetworkLatency)
}

// Implementación de métodos de la interfaz

// RecordHTTPRequest registra una request HTTP
func (ptrManager *structMetricsManager) RecordHTTPRequest(strMethod, strPath, strStatusCode string, fltDuration float64, int64RequestSize, int64ResponseSize int64) {
	strAPI := extractAPIFromPath(strPath)

	ptrManager.PtrHTTPRequestsTotal.WithLabelValues(strMethod, strPath, strStatusCode, strAPI).Inc()
	ptrManager.PtrHTTPRequestDuration.WithLabelValues(strMethod, strPath, strStatusCode, strAPI).Observe(fltDuration)

	if int64RequestSize > 0 {
		ptrManager.PtrHTTPRequestSize.WithLabelValues(strMethod, strPath, strAPI).Observe(float64(int64RequestSize))
	}

	if int64ResponseSize > 0 {
		ptrManager.PtrHTTPResponseSize.WithLabelValues(strMethod, strPath, strStatusCode, strAPI).Observe(float64(int64ResponseSize))
	}
}

// IncActiveRequests incrementa requests activas
func (ptrManager *structMetricsManager) IncActiveRequests() {
	ptrManager.PtrHTTPActiveRequests.Inc()
}

// DecActiveRequests decrementa requests activas
func (ptrManager *structMetricsManager) DecActiveRequests() {
	ptrManager.PtrHTTPActiveRequests.Dec()
}

// RecordDatabaseQuery registra una query de base de datos
func (ptrManager *structMetricsManager) RecordDatabaseQuery(strOperation, strTable string, fltDuration float64, boolSuccess bool) {
	strResult := "success"
	if !boolSuccess {
		strResult = "error"
	}

	ptrManager.PtrDBQueriesTotal.WithLabelValues(strOperation, strTable, strResult, "default").Inc()
	ptrManager.PtrDBQueryDuration.WithLabelValues(strOperation, strTable, "default").Observe(fltDuration)
}

// IncDatabaseConnections incrementa conexiones de base de datos
func (ptrManager *structMetricsManager) IncDatabaseConnections(strPool string) {
	ptrManager.PtrDBConnectionsActive.WithLabelValues(strPool, "postgres").Inc()
}

// DecDatabaseConnections decrementa conexiones de base de datos
func (ptrManager *structMetricsManager) DecDatabaseConnections(strPool string) {
	ptrManager.PtrDBConnectionsActive.WithLabelValues(strPool, "postgres").Dec()
}

// RecordDatabaseConnectionDuration registra duración de conexión
func (ptrManager *structMetricsManager) RecordDatabaseConnectionDuration(strPool string, fltDuration float64) {
	ptrManager.PtrDBConnectionDuration.WithLabelValues(strPool, "postgres").Observe(fltDuration)
}

// RecordBusinessOperation registra una operación de negocio
func (ptrManager *structMetricsManager) RecordBusinessOperation(strOperation, strResult string, fltDuration float64) {
	strAPI := "general"
	ptrManager.PtrBusinessOperationsTotal.WithLabelValues(strOperation, strResult, strAPI).Inc()
	ptrManager.PtrBusinessOperationDuration.WithLabelValues(strOperation, strAPI).Observe(fltDuration)
}

// RecordBusinessError registra un error de negocio
func (ptrManager *structMetricsManager) RecordBusinessError(strOperation, strErrorType string) {
	strAPI := "general"
	ptrManager.PtrBusinessErrorsTotal.WithLabelValues(strOperation, strErrorType, strAPI).Inc()
}

// IncBusinessCounter incrementa un contador de negocio
func (ptrManager *structMetricsManager) IncBusinessCounter(strMetric, strLabel string) {
	strAPI := "general"
	ptrManager.PtrBusinessCounters.WithLabelValues(strMetric, strLabel, strAPI).Inc()
}

// SetBusinessGauge establece un gauge de negocio
func (ptrManager *structMetricsManager) SetBusinessGauge(strMetric string, fltValue float64, mapLabels map[string]string) {
	strAPI := getStringFromMap(mapLabels, "api", "general")
	strSucursal := getStringFromMap(mapLabels, "sucursal", "")
	strCategoria := getStringFromMap(mapLabels, "categoria", "")

	ptrManager.PtrBusinessGauges.WithLabelValues(strMetric, strAPI, strSucursal, strCategoria).Set(fltValue)
}

// RecordMemoryUsage registra uso de memoria
func (ptrManager *structMetricsManager) RecordMemoryUsage(fltHeapMB, fltStackMB float64) {
	ptrManager.PtrSystemMemoryUsage.WithLabelValues("heap").Set(fltHeapMB)
	ptrManager.PtrSystemMemoryUsage.WithLabelValues("stack").Set(fltStackMB)
}

// RecordGoroutines registra número de goroutines
func (ptrManager *structMetricsManager) RecordGoroutines(intCount int) {
	ptrManager.PtrSystemGoroutines.Set(float64(intCount))
}

// RecordGCDuration registra duración de garbage collection
func (ptrManager *structMetricsManager) RecordGCDuration(fltDuration float64) {
	ptrManager.PtrSystemGCDuration.Observe(fltDuration)
}

// RecordCacheOperation registra una operación de cache
func (ptrManager *structMetricsManager) RecordCacheOperation(strOperation, strResult string) {
	ptrManager.PtrCacheOperationsTotal.WithLabelValues(strOperation, strResult, "default").Inc()
}

// SetCacheSize establece el tamaño del cache
func (ptrManager *structMetricsManager) SetCacheSize(strCacheName string, intSize int) {
	ptrManager.PtrCacheSize.WithLabelValues(strCacheName).Set(float64(intSize))
}

// RecordCacheHitRate registra la tasa de aciertos del cache
func (ptrManager *structMetricsManager) RecordCacheHitRate(strCacheName string, fltHitRate float64) {
	ptrManager.PtrCacheHitRate.WithLabelValues(strCacheName).Set(fltHitRate)
}

// RecordAuthAttempt registra un intento de autenticación
func (ptrManager *structMetricsManager) RecordAuthAttempt(strMethod, strResult string) {
	ptrManager.PtrAuthAttemptsTotal.WithLabelValues(strMethod, strResult).Inc()
}

// RecordSessionDuration registra duración de sesión
func (ptrManager *structMetricsManager) RecordSessionDuration(fltDuration float64) {
	ptrManager.PtrAuthSessionDuration.Observe(fltDuration)
}

// IncActiveUsers incrementa usuarios activos
func (ptrManager *structMetricsManager) IncActiveUsers() {
	ptrManager.PtrAuthActiveUsers.Inc()
}

// DecActiveUsers decrementa usuarios activos
func (ptrManager *structMetricsManager) DecActiveUsers() {
	ptrManager.PtrAuthActiveUsers.Dec()
}

// RecordRateLimit registra rate limiting
func (ptrManager *structMetricsManager) RecordRateLimit(strKey string, boolAllowed bool, intRemaining int) {
	strResult := "allowed"
	if !boolAllowed {
		strResult = "denied"
	}

	ptrManager.PtrRateLimitTotal.WithLabelValues(strKey, strResult).Inc()
	ptrManager.PtrRateLimitRemaining.WithLabelValues(strKey).Set(float64(intRemaining))
}

// RecordPOSOperation registra operación POS
func (ptrManager *structMetricsManager) RecordPOSOperation(strOperation, strResult string, fltDuration float64) {
	strSucursal := "default"
	ptrManager.PtrPOSOperationsTotal.WithLabelValues(strOperation, strResult, strSucursal).Inc()
	ptrManager.PtrPOSOperationDuration.WithLabelValues(strOperation, strSucursal).Observe(fltDuration)
}

// RecordSyncOperation registra operación de sincronización
func (ptrManager *structMetricsManager) RecordSyncOperation(strOperation, strResult string, fltDuration float64, intRecords int) {
	strSucursal := "default"
	ptrManager.PtrSyncOperationsTotal.WithLabelValues(strOperation, strResult, strSucursal).Inc()
	ptrManager.PtrSyncOperationDuration.WithLabelValues(strOperation, strSucursal).Observe(fltDuration)
	ptrManager.PtrSyncRecordsProcessed.WithLabelValues("general", strOperation, strSucursal).Add(float64(intRecords))
}

// RecordLabelGeneration registra generación de etiquetas
func (ptrManager *structMetricsManager) RecordLabelGeneration(strType, strResult string, fltDuration float64, intCount int) {
	strSucursal := "default"
	ptrManager.PtrLabelsGeneratedTotal.WithLabelValues(strType, strResult, strSucursal).Add(float64(intCount))
	ptrManager.PtrLabelsGenerationDuration.WithLabelValues(strType, strSucursal).Observe(fltDuration)
}

// RecordReportGeneration registra generación de reportes
func (ptrManager *structMetricsManager) RecordReportGeneration(strType, strResult string, fltDuration float64, intSize int64) {
	strSucursal := "default"
	ptrManager.PtrReportsGeneratedTotal.WithLabelValues(strType, strResult, strSucursal).Inc()
	ptrManager.PtrReportsGenerationDuration.WithLabelValues(strType, strSucursal).Observe(fltDuration)
	ptrManager.PtrReportsSize.WithLabelValues(strType, "pdf", strSucursal).Observe(float64(intSize))
}

// GetHandler obtiene el handler de métricas
func (ptrManager *structMetricsManager) GetHandler() http.Handler {
	return promhttp.HandlerFor(ptrManager.PtrRegistry, promhttp.HandlerOpts{})
}

// GetRegistry obtiene el registry de Prometheus
func (ptrManager *structMetricsManager) GetRegistry() *prometheus.Registry {
	return ptrManager.PtrRegistry
}

// Reset resetea todas las métricas
func (ptrManager *structMetricsManager) Reset() {
	ptrManager.MutexMetrics.Lock()
	defer ptrManager.MutexMetrics.Unlock()

	// Crear nuevo registry y re-registrar métricas
	ptrManager.PtrRegistry = prometheus.NewRegistry()
	ptrManager.initializeMetrics()
	ptrManager.registerMetrics()
}

// Funciones de utilidad

// extractAPIFromPath extrae el nombre de la API del path
func extractAPIFromPath(strPath string) string {
	if len(strPath) < 2 {
		return "unknown"
	}

	arrParts := strings.Split(strPath[1:], "/")
	if len(arrParts) > 0 {
		switch arrParts[0] {
		case "api":
			if len(arrParts) > 1 {
				return arrParts[1]
			}
		case "pos":
			return "pos"
		case "sync":
			return "sync"
		case "labels":
			return "labels"
		case "reports":
			return "reports"
		}
	}

	return "unknown"
}

// getStringFromMap obtiene un string de un map con valor por defecto
func getStringFromMap(mapData map[string]string, strKey, strDefault string) string {
	if strValue, boolExists := mapData[strKey]; boolExists {
		return strValue
	}
	return strDefault
}

// PrometheusMiddleware middleware para métricas automáticas
func PrometheusMiddleware(ptrMetrics interfaceManager) gin.HandlerFunc {
	return func(ptrCtx *gin.Context) {
		timeStart := time.Now()

		// Incrementar requests activas
		ptrMetrics.IncActiveRequests()
		defer ptrMetrics.DecActiveRequests()

		// Procesar request
		ptrCtx.Next()

		// Registrar métricas
		timeDuration := time.Since(timeStart)
		strStatusCode := strconv.Itoa(ptrCtx.Writer.Status())

		ptrMetrics.RecordHTTPRequest(
			ptrCtx.Request.Method,
			ptrCtx.Request.URL.Path,
			strStatusCode,
			timeDuration.Seconds(),
			ptrCtx.Request.ContentLength,
			int64(ptrCtx.Writer.Size()),
		)
	}
}

// Close cierra el manager de métricas
func (ptrManager *structMetricsManager) Close() error {
	// No hay recursos específicos que cerrar en Prometheus
	return nil
}


