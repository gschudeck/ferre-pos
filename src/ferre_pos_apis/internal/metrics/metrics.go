package metrics

import (
	"database/sql"
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	
	"ferre_pos_apis/internal/config"
	"ferre_pos_apis/internal/logger"
)

// Metrics estructura principal de métricas
type Metrics struct {
	registry *prometheus.Registry
	logger   logger.Logger
	config   *config.MetricsConfig
	
	// Métricas HTTP
	httpRequestsTotal     *prometheus.CounterVec
	httpRequestDuration   *prometheus.HistogramVec
	httpRequestsInFlight  *prometheus.GaugeVec
	
	// Métricas de Base de Datos
	dbConnectionsOpen     *prometheus.GaugeVec
	dbConnectionsIdle     *prometheus.GaugeVec
	dbConnectionsInUse    *prometheus.GaugeVec
	dbQueryDuration       *prometheus.HistogramVec
	dbQueryTotal          *prometheus.CounterVec
	dbTransactionDuration *prometheus.HistogramVec
	
	// Métricas de Aplicación
	apiRequestsTotal      *prometheus.CounterVec
	apiErrorsTotal        *prometheus.CounterVec
	apiResponseTime       *prometheus.HistogramVec
	
	// Métricas de Negocio
	ventasTotal           *prometheus.CounterVec
	ventasAmount          *prometheus.CounterVec
	productosConsultados  *prometheus.CounterVec
	etiquetasGeneradas    *prometheus.CounterVec
	reportesGenerados     *prometheus.CounterVec
	
	// Métricas de Sistema
	systemInfo            *prometheus.GaugeVec
	buildInfo             *prometheus.GaugeVec
}

var globalMetrics *Metrics

// Init inicializa el sistema de métricas
func Init(cfg *config.MetricsConfig, apiName string, log logger.Logger) (*Metrics, error) {
	if !cfg.Enabled {
		log.Info("Métricas deshabilitadas por configuración")
		return nil, nil
	}

	registry := prometheus.NewRegistry()
	
	m := &Metrics{
		registry: registry,
		logger:   log,
		config:   cfg,
	}

	// Inicializar métricas HTTP
	m.initHTTPMetrics(apiName)
	
	// Inicializar métricas de base de datos
	m.initDatabaseMetrics(apiName)
	
	// Inicializar métricas de aplicación
	m.initApplicationMetrics(apiName)
	
	// Inicializar métricas de negocio
	m.initBusinessMetrics(apiName)
	
	// Inicializar métricas de sistema
	m.initSystemMetrics(apiName)

	// Registrar todas las métricas
	if err := m.registerMetrics(); err != nil {
		return nil, fmt.Errorf("error registrando métricas: %w", err)
	}

	// Inicializar métricas de runtime si está habilitado
	if cfg.CollectRuntimeMetrics {
		registry.MustRegister(prometheus.NewGoCollector())
		registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	}

	globalMetrics = m
	log.WithField("api", apiName).Info("Sistema de métricas inicializado")
	
	return m, nil
}

// Get obtiene la instancia global de métricas
func Get() *Metrics {
	return globalMetrics
}

// initHTTPMetrics inicializa métricas HTTP
func (m *Metrics) initHTTPMetrics(apiName string) {
	m.httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: m.config.Namespace,
			Subsystem: "http",
			Name:      "requests_total",
			Help:      "Total number of HTTP requests",
		},
		[]string{"api", "method", "endpoint", "status_code"},
	)

	m.httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: m.config.Namespace,
			Subsystem: "http",
			Name:      "request_duration_seconds",
			Help:      "HTTP request duration in seconds",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"api", "method", "endpoint"},
	)

	m.httpRequestsInFlight = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: m.config.Namespace,
			Subsystem: "http",
			Name:      "requests_in_flight",
			Help:      "Number of HTTP requests currently being processed",
		},
		[]string{"api"},
	)
}

// initDatabaseMetrics inicializa métricas de base de datos
func (m *Metrics) initDatabaseMetrics(apiName string) {
	m.dbConnectionsOpen = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: m.config.Namespace,
			Subsystem: "database",
			Name:      "connections_open",
			Help:      "Number of open database connections",
		},
		[]string{"api"},
	)

	m.dbConnectionsIdle = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: m.config.Namespace,
			Subsystem: "database",
			Name:      "connections_idle",
			Help:      "Number of idle database connections",
		},
		[]string{"api"},
	)

	m.dbConnectionsInUse = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: m.config.Namespace,
			Subsystem: "database",
			Name:      "connections_in_use",
			Help:      "Number of database connections in use",
		},
		[]string{"api"},
	)

	m.dbQueryDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: m.config.Namespace,
			Subsystem: "database",
			Name:      "query_duration_seconds",
			Help:      "Database query duration in seconds",
			Buckets:   []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"api", "operation", "table"},
	)

	m.dbQueryTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: m.config.Namespace,
			Subsystem: "database",
			Name:      "queries_total",
			Help:      "Total number of database queries",
		},
		[]string{"api", "operation", "table", "status"},
	)

	m.dbTransactionDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: m.config.Namespace,
			Subsystem: "database",
			Name:      "transaction_duration_seconds",
			Help:      "Database transaction duration in seconds",
			Buckets:   []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30},
		},
		[]string{"api", "status"},
	)
}

// initApplicationMetrics inicializa métricas de aplicación
func (m *Metrics) initApplicationMetrics(apiName string) {
	m.apiRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: m.config.Namespace,
			Subsystem: "api",
			Name:      "requests_total",
			Help:      "Total number of API requests",
		},
		[]string{"api", "endpoint", "method", "status"},
	)

	m.apiErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: m.config.Namespace,
			Subsystem: "api",
			Name:      "errors_total",
			Help:      "Total number of API errors",
		},
		[]string{"api", "endpoint", "error_type", "error_code"},
	)

	m.apiResponseTime = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: m.config.Namespace,
			Subsystem: "api",
			Name:      "response_time_seconds",
			Help:      "API response time in seconds",
			Buckets:   []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"api", "endpoint", "method"},
	)
}

// initBusinessMetrics inicializa métricas de negocio
func (m *Metrics) initBusinessMetrics(apiName string) {
	m.ventasTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: m.config.Namespace,
			Subsystem: "business",
			Name:      "ventas_total",
			Help:      "Total number of sales",
		},
		[]string{"api", "sucursal", "tipo_documento", "estado"},
	)

	m.ventasAmount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: m.config.Namespace,
			Subsystem: "business",
			Name:      "ventas_amount_total",
			Help:      "Total sales amount",
		},
		[]string{"api", "sucursal", "tipo_documento"},
	)

	m.productosConsultados = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: m.config.Namespace,
			Subsystem: "business",
			Name:      "productos_consultados_total",
			Help:      "Total number of product queries",
		},
		[]string{"api", "sucursal", "tipo_consulta"},
	)

	m.etiquetasGeneradas = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: m.config.Namespace,
			Subsystem: "business",
			Name:      "etiquetas_generadas_total",
			Help:      "Total number of labels generated",
		},
		[]string{"api", "sucursal", "tipo_etiqueta", "formato"},
	)

	m.reportesGenerados = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: m.config.Namespace,
			Subsystem: "business",
			Name:      "reportes_generados_total",
			Help:      "Total number of reports generated",
		},
		[]string{"api", "sucursal", "tipo_reporte", "formato"},
	)
}

// initSystemMetrics inicializa métricas de sistema
func (m *Metrics) initSystemMetrics(apiName string) {
	m.systemInfo = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: m.config.Namespace,
			Subsystem: "system",
			Name:      "info",
			Help:      "System information",
		},
		[]string{"api", "version", "go_version", "os", "arch"},
	)

	m.buildInfo = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: m.config.Namespace,
			Subsystem: "build",
			Name:      "info",
			Help:      "Build information",
		},
		[]string{"api", "version", "commit", "build_date"},
	)
}

// registerMetrics registra todas las métricas en el registry
func (m *Metrics) registerMetrics() error {
	metrics := []prometheus.Collector{
		m.httpRequestsTotal,
		m.httpRequestDuration,
		m.httpRequestsInFlight,
		m.dbConnectionsOpen,
		m.dbConnectionsIdle,
		m.dbConnectionsInUse,
		m.dbQueryDuration,
		m.dbQueryTotal,
		m.dbTransactionDuration,
		m.apiRequestsTotal,
		m.apiErrorsTotal,
		m.apiResponseTime,
		m.ventasTotal,
		m.ventasAmount,
		m.productosConsultados,
		m.etiquetasGeneradas,
		m.reportesGenerados,
		m.systemInfo,
		m.buildInfo,
	}

	for _, metric := range metrics {
		if err := m.registry.Register(metric); err != nil {
			return err
		}
	}

	return nil
}

// Handler retorna el handler HTTP para métricas
func (m *Metrics) Handler() http.Handler {
	if m == nil {
		return http.NotFoundHandler()
	}
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
}

// Middleware de Gin para métricas HTTP
func (m *Metrics) GinMiddleware(apiName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// Incrementar requests en vuelo
		m.httpRequestsInFlight.WithLabelValues(apiName).Inc()
		defer m.httpRequestsInFlight.WithLabelValues(apiName).Dec()

		// Procesar request
		c.Next()

		// Calcular duración
		duration := time.Since(start)
		
		// Obtener información del request
		method := c.Request.Method
		endpoint := c.FullPath()
		if endpoint == "" {
			endpoint = "unknown"
		}
		statusCode := strconv.Itoa(c.Writer.Status())

		// Registrar métricas
		m.httpRequestsTotal.WithLabelValues(apiName, method, endpoint, statusCode).Inc()
		m.httpRequestDuration.WithLabelValues(apiName, method, endpoint).Observe(duration.Seconds())
		m.apiRequestsTotal.WithLabelValues(apiName, endpoint, method, getStatusCategory(c.Writer.Status())).Inc()
		m.apiResponseTime.WithLabelValues(apiName, endpoint, method).Observe(duration.Seconds())

		// Registrar errores si es necesario
		if c.Writer.Status() >= 400 {
			errorType := "client_error"
			if c.Writer.Status() >= 500 {
				errorType = "server_error"
			}
			m.apiErrorsTotal.WithLabelValues(apiName, endpoint, errorType, statusCode).Inc()
		}
	}
}

// RecordDatabaseQuery registra métricas de query de base de datos
func (m *Metrics) RecordDatabaseQuery(apiName, operation, table string, duration time.Duration, err error) {
	if m == nil {
		return
	}

	status := "success"
	if err != nil {
		status = "error"
	}

	m.dbQueryDuration.WithLabelValues(apiName, operation, table).Observe(duration.Seconds())
	m.dbQueryTotal.WithLabelValues(apiName, operation, table, status).Inc()
}

// RecordDatabaseTransaction registra métricas de transacción de base de datos
func (m *Metrics) RecordDatabaseTransaction(apiName string, duration time.Duration, err error) {
	if m == nil {
		return
	}

	status := "success"
	if err != nil {
		status = "error"
	}

	m.dbTransactionDuration.WithLabelValues(apiName, status).Observe(duration.Seconds())
}

// UpdateDatabaseConnectionMetrics actualiza métricas de conexiones de base de datos
func (m *Metrics) UpdateDatabaseConnectionMetrics(apiName string, stats sql.DBStats) {
	if m == nil {
		return
	}

	m.dbConnectionsOpen.WithLabelValues(apiName).Set(float64(stats.OpenConnections))
	m.dbConnectionsIdle.WithLabelValues(apiName).Set(float64(stats.Idle))
	m.dbConnectionsInUse.WithLabelValues(apiName).Set(float64(stats.InUse))
}

// RecordVenta registra métricas de venta
func (m *Metrics) RecordVenta(apiName, sucursal, tipoDocumento, estado string, monto float64) {
	if m == nil {
		return
	}

	m.ventasTotal.WithLabelValues(apiName, sucursal, tipoDocumento, estado).Inc()
	if estado == "finalizada" {
		m.ventasAmount.WithLabelValues(apiName, sucursal, tipoDocumento).Add(monto)
	}
}

// RecordProductoConsulta registra métricas de consulta de producto
func (m *Metrics) RecordProductoConsulta(apiName, sucursal, tipoConsulta string) {
	if m == nil {
		return
	}

	m.productosConsultados.WithLabelValues(apiName, sucursal, tipoConsulta).Inc()
}

// RecordEtiquetaGenerada registra métricas de etiqueta generada
func (m *Metrics) RecordEtiquetaGenerada(apiName, sucursal, tipoEtiqueta, formato string, cantidad int) {
	if m == nil {
		return
	}

	m.etiquetasGeneradas.WithLabelValues(apiName, sucursal, tipoEtiqueta, formato).Add(float64(cantidad))
}

// RecordReporteGenerado registra métricas de reporte generado
func (m *Metrics) RecordReporteGenerado(apiName, sucursal, tipoReporte, formato string) {
	if m == nil {
		return
	}

	m.reportesGenerados.WithLabelValues(apiName, sucursal, tipoReporte, formato).Inc()
}

// SetSystemInfo establece información del sistema
func (m *Metrics) SetSystemInfo(apiName, version string) {
	if m == nil {
		return
	}

	m.systemInfo.WithLabelValues(
		apiName,
		version,
		runtime.Version(),
		runtime.GOOS,
		runtime.GOARCH,
	).Set(1)
}

// SetBuildInfo establece información de build
func (m *Metrics) SetBuildInfo(apiName, version, commit, buildDate string) {
	if m == nil {
		return
	}

	m.buildInfo.WithLabelValues(apiName, version, commit, buildDate).Set(1)
}

// StartMetricsCollector inicia el collector de métricas en background
func (m *Metrics) StartMetricsCollector(apiName string, db interface{}) {
	if m == nil {
		return
	}

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			// Actualizar métricas de base de datos si está disponible
			if sqlDB, ok := db.(interface{ Stats() sql.DBStats }); ok {
				m.UpdateDatabaseConnectionMetrics(apiName, sqlDB.Stats())
			}
		}
	}()
}

// getStatusCategory categoriza códigos de estado HTTP
func getStatusCategory(statusCode int) string {
	switch {
	case statusCode >= 200 && statusCode < 300:
		return "2xx"
	case statusCode >= 300 && statusCode < 400:
		return "3xx"
	case statusCode >= 400 && statusCode < 500:
		return "4xx"
	case statusCode >= 500:
		return "5xx"
	default:
		return "unknown"
	}
}


// noOpMetrics implementación que no hace nada cuando las métricas están deshabilitadas
type noOpMetrics struct{}

func (n *noOpMetrics) GinMiddleware(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) { c.Next() }
}

func (n *noOpMetrics) Handler() http.Handler {
	return http.NotFoundHandler()
}

func (n *noOpMetrics) RecordDatabaseQuery(apiName, operation, table string, duration time.Duration, err error) {}
func (n *noOpMetrics) RecordCacheOperation(apiName, operation string, hit bool, duration time.Duration) {}
func (n *noOpMetrics) RecordProductoConsulta(apiName, sucursalID, operationType string) {}
func (n *noOpMetrics) RecordVentaProcesada(apiName, sucursalID, tipoDocumento string, monto float64) {}
func (n *noOpMetrics) RecordStockMovimiento(apiName, sucursalID, tipoMovimiento string, cantidad int) {}
func (n *noOpMetrics) RecordSyncOperation(apiName, operationType string, recordsProcessed int, duration time.Duration, success bool) {}
func (n *noOpMetrics) RecordLabelGenerated(apiName, templateType, format string) {}
func (n *noOpMetrics) RecordReportGenerated(apiName, reportType, format string, duration time.Duration) {}
func (n *noOpMetrics) UpdateDatabaseConnectionMetrics(apiName string, stats sql.DBStats) {}
func (n *noOpMetrics) StartMetricsCollector(apiName string, db interface{}) {}
func (n *noOpMetrics) SetSystemInfo(apiName, version string) {}
func (n *noOpMetrics) SetBuildInfo(apiName, version, environment, buildDate string) {}

