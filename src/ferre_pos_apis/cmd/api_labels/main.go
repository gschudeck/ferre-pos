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
	apiName    = "labels"
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

	log.Info("Iniciando API LABELS - Sistema FERRE-POS")

	// Verificar si la API está habilitada
	apiConfig, err := config.GetAPIConfig(apiName)
	if err != nil {
		log.WithError(err).Fatal("Error obteniendo configuración de API")
	}

	if !apiConfig.Enabled {
		log.Info("API LABELS deshabilitada por configuración")
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

	// Inicializar rate limiter
	rateLimiter := ratelimiter.NewIPRateLimiter(&apiConfig.RateLimit, log)

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
		log.WithField("address", server.Addr).Info("Servidor API LABELS iniciado")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.WithError(err).Fatal("Error iniciando servidor")
		}
	}()

	// Configurar graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Cerrando servidor API LABELS...")

	// Timeout para shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.WithError(err).Error("Error durante shutdown del servidor")
	}

	log.Info("Servidor API LABELS cerrado exitosamente")
}

// setupRoutes configura todas las rutas de la API Labels
func setupRoutes(
	router *gin.Engine,
	db *database.Database,
	log logger.Logger,
	validator validator.Validator,
	metrics *metrics.Metrics,
	cfg *config.APIConfig,
) {
	// Inicializar handlers específicos de labels
	labelsHandler := handlers.NewLabelsHandler(db, log, validator, metrics)
	templatesHandler := handlers.NewTemplatesHandler(db, log, validator, metrics)
	printHandler := handlers.NewPrintHandler(db, log, validator, metrics)
	barcodeHandler := handlers.NewBarcodeHandler(db, log, validator, metrics)

	// Rutas de salud y métricas
	router.GET("/health", handlers.HealthCheck(db, log))
	router.GET("/ready", handlers.ReadinessCheck(db, log))
	
	if metrics != nil {
		router.GET("/metrics", gin.WrapH(metrics.Handler()))
	}

	// Grupo de rutas API v1
	v1 := router.Group(fmt.Sprintf("/api/%s", apiVersion))
	{
		// Rutas de autenticación (reutilizando del sistema principal)
		auth := v1.Group("/auth")
		{
			authHandler := handlers.NewAuthHandler(db, log, validator, cfg)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.POST("/logout", middleware.Auth(cfg), authHandler.Logout)
		}

		// Rutas protegidas
		protected := v1.Group("")
		protected.Use(middleware.Auth(cfg))
		{
			// Rutas de generación de etiquetas
			labels := protected.Group("/labels")
			{
				// Generación de etiquetas individuales
				labels.POST("/generate", labelsHandler.GenerateLabel)
				labels.POST("/batch", labelsHandler.GenerateBatch)
				labels.POST("/preview", labelsHandler.PreviewLabel)
				
				// Gestión de etiquetas generadas
				labels.GET("", labelsHandler.ListLabels)
				labels.GET("/:id", labelsHandler.GetLabel)
				labels.DELETE("/:id", labelsHandler.DeleteLabel)
				labels.POST("/:id/reprint", labelsHandler.ReprintLabel)
				
				// Descarga de etiquetas
				labels.GET("/:id/download", labelsHandler.DownloadLabel)
				labels.POST("/download-batch", labelsHandler.DownloadBatch)
				
				// Estadísticas de etiquetas
				labels.GET("/stats", labelsHandler.GetLabelStats)
				labels.GET("/history", labelsHandler.GetLabelHistory)
			}

			// Rutas de plantillas de etiquetas
			templates := protected.Group("/templates")
			{
				templates.GET("", templatesHandler.ListTemplates)
				templates.GET("/:id", templatesHandler.GetTemplate)
				templates.POST("", middleware.RequireRole("admin", "supervisor"), templatesHandler.CreateTemplate)
				templates.PUT("/:id", middleware.RequireRole("admin", "supervisor"), templatesHandler.UpdateTemplate)
				templates.DELETE("/:id", middleware.RequireRole("admin"), templatesHandler.DeleteTemplate)
				
				// Operaciones de plantillas
				templates.POST("/:id/duplicate", templatesHandler.DuplicateTemplate)
				templates.POST("/:id/validate", templatesHandler.ValidateTemplate)
				templates.GET("/:id/preview", templatesHandler.PreviewTemplate)
				
				// Plantillas predefinidas
				templates.GET("/predefined", templatesHandler.GetPredefinedTemplates)
				templates.POST("/import", templatesHandler.ImportTemplate)
				templates.GET("/:id/export", templatesHandler.ExportTemplate)
			}

			// Rutas de impresión
			printing := protected.Group("/printing")
			{
				// Gestión de impresoras
				printing.GET("/printers", printHandler.ListPrinters)
				printing.GET("/printers/:id", printHandler.GetPrinter)
				printing.POST("/printers", middleware.RequireRole("admin", "supervisor"), printHandler.AddPrinter)
				printing.PUT("/printers/:id", middleware.RequireRole("admin", "supervisor"), printHandler.UpdatePrinter)
				printing.DELETE("/printers/:id", middleware.RequireRole("admin"), printHandler.DeletePrinter)
				
				// Estado de impresoras
				printing.GET("/printers/:id/status", printHandler.GetPrinterStatus)
				printing.POST("/printers/:id/test", printHandler.TestPrinter)
				
				// Cola de impresión
				printing.GET("/queue", printHandler.GetPrintQueue)
				printing.POST("/queue", printHandler.AddToPrintQueue)
				printing.DELETE("/queue/:id", printHandler.RemoveFromQueue)
				printing.POST("/queue/:id/retry", printHandler.RetryPrintJob)
				
				// Trabajos de impresión
				printing.GET("/jobs", printHandler.ListPrintJobs)
				printing.GET("/jobs/:id", printHandler.GetPrintJob)
				printing.POST("/jobs/:id/cancel", printHandler.CancelPrintJob)
			}

			// Rutas de códigos de barras
			barcodes := protected.Group("/barcodes")
			{
				// Generación de códigos de barras
				barcodes.POST("/generate", barcodeHandler.GenerateBarcode)
				barcodes.POST("/validate", barcodeHandler.ValidateBarcode)
				barcodes.GET("/formats", barcodeHandler.GetSupportedFormats)
				
				// Gestión de códigos adicionales
				barcodes.GET("/producto/:producto_id", barcodeHandler.GetProductBarcodes)
				barcodes.POST("/producto/:producto_id", barcodeHandler.AddProductBarcode)
				barcodes.DELETE("/producto/:producto_id/:codigo", barcodeHandler.RemoveProductBarcode)
				
				// Búsqueda y verificación
				barcodes.GET("/search", barcodeHandler.SearchByBarcode)
				barcodes.POST("/bulk-validate", barcodeHandler.BulkValidateBarcodes)
			}

			// Rutas de configuración de etiquetas
			config := protected.Group("/config")
			{
				config.GET("/label-settings", labelsHandler.GetLabelSettings)
				config.PUT("/label-settings", middleware.RequireRole("admin", "supervisor"), labelsHandler.UpdateLabelSettings)
				config.GET("/print-settings", printHandler.GetPrintSettings)
				config.PUT("/print-settings", middleware.RequireRole("admin", "supervisor"), printHandler.UpdatePrintSettings)
				
				// Configuración por sucursal
				config.GET("/sucursal/:sucursal_id", labelsHandler.GetSucursalConfig)
				config.PUT("/sucursal/:sucursal_id", middleware.RequireRole("admin", "supervisor"), labelsHandler.UpdateSucursalConfig)
			}

			// Rutas de reportes específicos de etiquetas
			reports := protected.Group("/reports")
			{
				reports.GET("/usage", labelsHandler.GetUsageReport)
				reports.GET("/errors", labelsHandler.GetErrorReport)
				reports.GET("/printer-stats", printHandler.GetPrinterStatsReport)
				reports.GET("/cost-analysis", labelsHandler.GetCostAnalysisReport)
			}
		}

		// Rutas públicas (para integración con sistemas externos)
		public := v1.Group("/public")
		{
			public.GET("/formats", labelsHandler.GetSupportedFormats)
			public.GET("/templates/public", templatesHandler.GetPublicTemplates)
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

