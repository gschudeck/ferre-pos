package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	
	"ferre_pos_apis/internal/config"
	"ferre_pos_apis/internal/database"
	"ferre_pos_apis/internal/logger"
	"ferre_pos_apis/internal/metrics"
	"ferre_pos_apis/internal/handlers"
	"ferre_pos_apis/internal/middleware"
	"ferre_pos_apis/pkg/ratelimiter"
	"ferre_pos_apis/pkg/validator"
)

const (
	apiName    = "sync"
	apiVersion = "v1"
)

func main() {
	// Cargar configuración
	cfg, err := config.Load("")
	if err != nil {
		fmt.Printf("Error cargando configuración: %v\n", err)
		os.Exit(1)
	}

	// Inicializar logger
	if err := logger.Init(&cfg.Logging, apiName); err != nil {
		fmt.Printf("Error inicializando logger: %v\n", err)
		os.Exit(1)
	}
	log := logger.Get()

	log.Info("Iniciando API SYNC - Sistema FERRE-POS")

	// Verificar si la API está habilitada
	apiConfig, err := config.GetAPIConfig(apiName)
	if err != nil {
		log.WithError(err).Fatal("Error obteniendo configuración de API")
	}

	if !apiConfig.Enabled {
		log.Info("API SYNC deshabilitada por configuración")
		return
	}

	// Inicializar base de datos
	db, err := database.Init(&cfg.Database, log)
	if err != nil {
		log.WithError(err).Fatal("Error inicializando base de datos")
	}
	defer db.Close()

	// Inicializar métricas
	metricsInstance, err := metrics.Init(&cfg.Metrics, apiName, log)
	if err != nil {
		log.WithError(err).Fatal("Error inicializando métricas")
	}

	// Inicializar validador
	validatorInstance := validator.New()

	// Inicializar rate limiter con límites más altos para sync
	syncRateLimit := apiConfig.RateLimit
	syncRateLimit.RequestsPerSecond = syncRateLimit.RequestsPerSecond * 2 // Doble límite para sync
	syncRateLimit.BurstSize = syncRateLimit.BurstSize * 2
	
	rateLimiter := ratelimiter.NewIPRateLimiter(&syncRateLimit, log)

	// Configurar Gin
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Middleware global
	router.Use(gin.Recovery())
	router.Use(middleware.RequestID())
	router.Use(middleware.Logger(log))
	router.Use(middleware.CORS(&cfg.Security.CORS))
	router.Use(gzip.Gzip(gzip.DefaultCompression))
	
	// Middleware de métricas
	if metricsInstance != nil {
		router.Use(metricsInstance.GinMiddleware(apiName))
	}

	// Middleware de rate limiting
	router.Use(rateLimiter.GinMiddleware())

	// Configurar rutas
	setupRoutes(router, db, log, validatorInstance, metricsInstance, apiConfig)

	// Configurar servidor HTTP
	server := &http.Server{
		Addr:           fmt.Sprintf("%s:%d", apiConfig.Host, apiConfig.Port),
		Handler:        router,
		ReadTimeout:    apiConfig.ReadTimeout,
		WriteTimeout:   apiConfig.WriteTimeout,
		IdleTimeout:    apiConfig.IdleTimeout,
		MaxHeaderBytes: apiConfig.MaxHeaderBytes,
	}

	// Iniciar collector de métricas
	if metricsInstance != nil {
		metricsInstance.StartMetricsCollector(apiName, db)
		metricsInstance.SetSystemInfo(apiName, "1.0.0")
		metricsInstance.SetBuildInfo(apiName, "1.0.0", "dev", time.Now().Format("2006-01-02"))
	}

	// Iniciar servidor en goroutine
	go func() {
		log.WithField("address", server.Addr).Info("Servidor API SYNC iniciado")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.WithError(err).Fatal("Error iniciando servidor")
		}
	}()

	// Configurar graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Cerrando servidor API SYNC...")

	// Timeout para shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.WithError(err).Error("Error durante shutdown del servidor")
	}

	log.Info("Servidor API SYNC cerrado exitosamente")
}

// setupRoutes configura todas las rutas de la API Sync
func setupRoutes(
	router *gin.Engine,
	db *database.Database,
	log logger.Logger,
	validator validator.Validator,
	metrics *metrics.Metrics,
	cfg *config.APIConfig,
) {
	// Inicializar handlers específicos de sync
	syncHandler := handlers.NewSyncHandler(db, log, validator, metrics)
	terminalHandler := handlers.NewTerminalHandler(db, log, validator, metrics)
	dataHandler := handlers.NewDataHandler(db, log, validator, metrics)
	conflictHandler := handlers.NewConflictHandler(db, log, validator, metrics)

	// Rutas de salud y métricas
	router.GET("/health", handlers.HealthCheck(db, log))
	router.GET("/ready", handlers.ReadinessCheck(db, log))
	
	if metrics != nil {
		router.GET("/metrics", gin.WrapH(metrics.Handler()))
	}

	// Grupo de rutas API v1
	v1 := router.Group(fmt.Sprintf("/api/%s", apiVersion))
	{
		// Rutas de autenticación de terminales
		auth := v1.Group("/auth")
		{
			auth.POST("/terminal", terminalHandler.AuthenticateTerminal)
			auth.POST("/refresh", terminalHandler.RefreshTerminalToken)
		}

		// Rutas protegidas (requieren autenticación de terminal)
		protected := v1.Group("")
		protected.Use(middleware.TerminalAuth(cfg))
		{
			// Rutas de sincronización principal
			sync := protected.Group("/sync")
			{
				// Sincronización bidireccional
				sync.POST("/full", syncHandler.FullSync)
				sync.POST("/incremental", syncHandler.IncrementalSync)
				sync.POST("/delta", syncHandler.DeltaSync)
				
				// Sincronización por entidad
				sync.POST("/productos", syncHandler.SyncProductos)
				sync.POST("/stock", syncHandler.SyncStock)
				sync.POST("/ventas", syncHandler.SyncVentas)
				sync.POST("/usuarios", syncHandler.SyncUsuarios)
				sync.POST("/configuracion", syncHandler.SyncConfiguracion)
				
				// Estado de sincronización
				sync.GET("/status", syncHandler.GetSyncStatus)
				sync.POST("/status", syncHandler.UpdateSyncStatus)
				sync.GET("/pending", syncHandler.GetPendingChanges)
			}

			// Rutas de gestión de terminales
			terminals := protected.Group("/terminals")
			{
				terminals.POST("/heartbeat", terminalHandler.Heartbeat)
				terminals.GET("/config", terminalHandler.GetConfiguration)
				terminals.POST("/config", terminalHandler.UpdateConfiguration)
				terminals.GET("/status", terminalHandler.GetStatus)
				terminals.POST("/log", terminalHandler.ReceiveLog)
			}

			// Rutas de gestión de datos
			data := protected.Group("/data")
			{
				// Upload de datos desde terminales
				data.POST("/upload", dataHandler.UploadData)
				data.POST("/batch", dataHandler.BatchUpload)
				
				// Download de datos hacia terminales
				data.GET("/download", dataHandler.DownloadData)
				data.GET("/changes", dataHandler.GetChanges)
				
				// Validación de integridad
				data.POST("/validate", dataHandler.ValidateData)
				data.GET("/checksum", dataHandler.GetChecksum)
			}

			// Rutas de resolución de conflictos
			conflicts := protected.Group("/conflicts")
			{
				conflicts.GET("", conflictHandler.ListConflicts)
				conflicts.GET("/:id", conflictHandler.GetConflict)
				conflicts.POST("/:id/resolve", conflictHandler.ResolveConflict)
				conflicts.POST("/auto-resolve", conflictHandler.AutoResolveConflicts)
			}

			// Rutas de monitoreo y estadísticas
			monitoring := protected.Group("/monitoring")
			{
				monitoring.GET("/stats", syncHandler.GetSyncStats)
				monitoring.GET("/performance", syncHandler.GetPerformanceMetrics)
				monitoring.GET("/errors", syncHandler.GetSyncErrors)
				monitoring.POST("/reset", syncHandler.ResetSyncState)
			}
		}

		// Rutas públicas (sin autenticación)
		public := v1.Group("/public")
		{
			public.GET("/server-info", syncHandler.GetServerInfo)
			public.GET("/sync-policies", syncHandler.GetSyncPolicies)
		}
	}

	// Ruta catch-all para 404
	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "NOT_FOUND",
				"message": "Endpoint no encontrado",
			},
		})
	})
}

