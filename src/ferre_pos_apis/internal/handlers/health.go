package handlers

import (
	"context"
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	
	"ferre_pos_apis/internal/database"
	"ferre_pos_apis/internal/logger"
	"ferre_pos_apis/internal/models"
)

// HealthCheck handler para verificar el estado básico de la API
func HealthCheck(db *database.Database, log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		status := "healthy"
		statusCode := http.StatusOK
		
		// Verificar conexión a base de datos
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		if err := db.HealthCheck(ctx); err != nil {
			status = "unhealthy"
			statusCode = http.StatusServiceUnavailable
			log.WithError(err).Error("Health check failed - database error")
		}

		response := gin.H{
			"status":    status,
			"timestamp": time.Now().UTC(),
			"service":   "ferre-pos-api",
			"version":   "1.0.0",
		}

		if status == "unhealthy" {
			response["error"] = "Database connection failed"
		}

		c.JSON(statusCode, response)
	}
}

// ReadinessCheck handler para verificar si la API está lista para recibir tráfico
func ReadinessCheck(db *database.Database, log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		checks := make(map[string]interface{})
		allHealthy := true

		// Check de base de datos
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		dbStart := time.Now()
		if err := db.HealthCheck(ctx); err != nil {
			checks["database"] = map[string]interface{}{
				"status":   "unhealthy",
				"error":    err.Error(),
				"duration": time.Since(dbStart).Milliseconds(),
			}
			allHealthy = false
		} else {
			checks["database"] = map[string]interface{}{
				"status":   "healthy",
				"duration": time.Since(dbStart).Milliseconds(),
			}
		}

		// Check de memoria
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		
		memoryUsageMB := float64(m.Alloc) / 1024 / 1024
		memoryStatus := "healthy"
		if memoryUsageMB > 1000 { // Más de 1GB
			memoryStatus = "warning"
		}
		if memoryUsageMB > 2000 { // Más de 2GB
			memoryStatus = "unhealthy"
			allHealthy = false
		}

		checks["memory"] = map[string]interface{}{
			"status":      memoryStatus,
			"usage_mb":    memoryUsageMB,
			"sys_mb":      float64(m.Sys) / 1024 / 1024,
			"gc_cycles":   m.NumGC,
		}

		// Check de goroutines
		numGoroutines := runtime.NumGoroutine()
		goroutineStatus := "healthy"
		if numGoroutines > 1000 {
			goroutineStatus = "warning"
		}
		if numGoroutines > 5000 {
			goroutineStatus = "unhealthy"
			allHealthy = false
		}

		checks["goroutines"] = map[string]interface{}{
			"status": goroutineStatus,
			"count":  numGoroutines,
		}

		// Check de conexiones de base de datos
		dbStats := db.Stats()
		connectionStatus := "healthy"
		if dbStats.OpenConnections >= dbStats.MaxOpenConnections-5 {
			connectionStatus = "warning"
		}
		if dbStats.OpenConnections >= dbStats.MaxOpenConnections {
			connectionStatus = "unhealthy"
			allHealthy = false
		}

		checks["database_connections"] = map[string]interface{}{
			"status":           connectionStatus,
			"open":             dbStats.OpenConnections,
			"max_open":         dbStats.MaxOpenConnections,
			"idle":             dbStats.Idle,
			"in_use":           dbStats.InUse,
			"wait_count":       dbStats.WaitCount,
			"wait_duration_ms": dbStats.WaitDuration.Milliseconds(),
		}

		// Determinar status general
		overallStatus := "ready"
		statusCode := http.StatusOK
		
		if !allHealthy {
			overallStatus = "not_ready"
			statusCode = http.StatusServiceUnavailable
		}

		response := gin.H{
			"status":    overallStatus,
			"timestamp": time.Now().UTC(),
			"service":   "ferre-pos-api",
			"version":   "1.0.0",
			"checks":    checks,
		}

		c.JSON(statusCode, response)
	}
}

// LivenessCheck handler para verificar si la aplicación está viva
func LivenessCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "alive",
			"timestamp": time.Now().UTC(),
			"service":   "ferre-pos-api",
			"version":   "1.0.0",
			"uptime":    time.Since(startTime).Seconds(),
		})
	}
}

// MetricsHealth handler para verificar el estado de las métricas
func MetricsHealth() gin.HandlerFunc {
	return func(c *gin.Context) {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		metrics := gin.H{
			"memory": gin.H{
				"alloc_mb":      float64(m.Alloc) / 1024 / 1024,
				"total_alloc_mb": float64(m.TotalAlloc) / 1024 / 1024,
				"sys_mb":        float64(m.Sys) / 1024 / 1024,
				"num_gc":        m.NumGC,
				"gc_pause_ns":   m.PauseNs[(m.NumGC+255)%256],
			},
			"runtime": gin.H{
				"goroutines":   runtime.NumGoroutine(),
				"go_version":   runtime.Version(),
				"go_os":        runtime.GOOS,
				"go_arch":      runtime.GOARCH,
				"num_cpu":      runtime.NumCPU(),
			},
			"timestamp": time.Now().UTC(),
		}

		c.JSON(http.StatusOK, metrics)
	}
}

// DetailedHealthCheck handler para un health check detallado
func DetailedHealthCheck(db *database.Database, log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		checks := make(map[string]interface{})
		allHealthy := true

		// Database health check con detalles
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		dbCheckStart := time.Now()
		dbError := db.HealthCheck(ctx)
		dbCheckDuration := time.Since(dbCheckStart)

		if dbError != nil {
			checks["database"] = map[string]interface{}{
				"status":      "unhealthy",
				"error":       dbError.Error(),
				"duration_ms": dbCheckDuration.Milliseconds(),
			}
			allHealthy = false
		} else {
			dbStats := db.Stats()
			checks["database"] = map[string]interface{}{
				"status":                "healthy",
				"duration_ms":           dbCheckDuration.Milliseconds(),
				"open_connections":      dbStats.OpenConnections,
				"max_open_connections":  dbStats.MaxOpenConnections,
				"idle_connections":      dbStats.Idle,
				"in_use_connections":    dbStats.InUse,
				"wait_count":            dbStats.WaitCount,
				"wait_duration_ms":      dbStats.WaitDuration.Milliseconds(),
				"max_idle_closed":       dbStats.MaxIdleClosed,
				"max_lifetime_closed":   dbStats.MaxLifetimeClosed,
			}
		}

		// Memory check detallado
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		memoryUsageMB := float64(m.Alloc) / 1024 / 1024
		memoryStatus := "healthy"
		if memoryUsageMB > 1000 {
			memoryStatus = "warning"
		}
		if memoryUsageMB > 2000 {
			memoryStatus = "unhealthy"
			allHealthy = false
		}

		checks["memory"] = map[string]interface{}{
			"status":          memoryStatus,
			"alloc_mb":        memoryUsageMB,
			"total_alloc_mb":  float64(m.TotalAlloc) / 1024 / 1024,
			"sys_mb":          float64(m.Sys) / 1024 / 1024,
			"heap_alloc_mb":   float64(m.HeapAlloc) / 1024 / 1024,
			"heap_sys_mb":     float64(m.HeapSys) / 1024 / 1024,
			"heap_idle_mb":    float64(m.HeapIdle) / 1024 / 1024,
			"heap_inuse_mb":   float64(m.HeapInuse) / 1024 / 1024,
			"heap_released_mb": float64(m.HeapReleased) / 1024 / 1024,
			"heap_objects":    m.HeapObjects,
			"stack_inuse_mb":  float64(m.StackInuse) / 1024 / 1024,
			"stack_sys_mb":    float64(m.StackSys) / 1024 / 1024,
			"num_gc":          m.NumGC,
			"gc_pause_ns":     m.PauseNs[(m.NumGC+255)%256],
		}

		// Runtime check
		numGoroutines := runtime.NumGoroutine()
		runtimeStatus := "healthy"
		if numGoroutines > 1000 {
			runtimeStatus = "warning"
		}
		if numGoroutines > 5000 {
			runtimeStatus = "unhealthy"
			allHealthy = false
		}

		checks["runtime"] = map[string]interface{}{
			"status":       runtimeStatus,
			"goroutines":   numGoroutines,
			"go_version":   runtime.Version(),
			"go_os":        runtime.GOOS,
			"go_arch":      runtime.GOARCH,
			"num_cpu":      runtime.NumCPU(),
			"num_cgo_call": runtime.NumCgoCall(),
		}

		// Disk space check (si es posible)
		// Esto requeriría implementación específica del sistema operativo

		// Overall status
		overallStatus := "healthy"
		statusCode := http.StatusOK
		
		if !allHealthy {
			overallStatus = "unhealthy"
			statusCode = http.StatusServiceUnavailable
		}

		totalDuration := time.Since(startTime)

		response := models.APIResponse{
			Success: allHealthy,
			Data: gin.H{
				"status":          overallStatus,
				"timestamp":       time.Now().UTC(),
				"service":         "ferre-pos-api",
				"version":         "1.0.0",
				"uptime_seconds":  time.Since(startTime).Seconds(),
				"check_duration_ms": totalDuration.Milliseconds(),
				"checks":          checks,
			},
			RequestID: getRequestID(c),
			Timestamp: time.Now(),
		}

		if !allHealthy {
			response.Error = &models.APIError{
				Code:    "HEALTH_CHECK_FAILED",
				Message: "Uno o más checks de salud fallaron",
			}
		}

		c.JSON(statusCode, response)
	}
}

// startTime para calcular uptime
var startTime = time.Now()

// init inicializa el tiempo de inicio
func init() {
	startTime = time.Now()
}

