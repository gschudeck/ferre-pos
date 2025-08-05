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
	apiName    = "reports"
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

	log.Info("Iniciando API REPORTS - Sistema FERRE-POS")

	// Verificar si la API está habilitada
	apiConfig, err := config.GetAPIConfig(apiName)
	if err != nil {
		log.WithError(err).Fatal("Error obteniendo configuración de API")
	}

	if !apiConfig.Enabled {
		log.Info("API REPORTS deshabilitada por configuración")
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

	// Inicializar rate limiter con límites más conservadores para reportes
	reportsRateLimit := apiConfig.RateLimit
	reportsRateLimit.RequestsPerSecond = reportsRateLimit.RequestsPerSecond * 0.5 // Menor límite para reportes
	reportsRateLimit.BurstSize = reportsRateLimit.BurstSize / 2
	
	rateLimiter := ratelimiter.NewIPRateLimiter(&reportsRateLimit, log)

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
		ReadTimeout:    apiConfig.ReadTimeout * 2,    // Más tiempo para reportes
		WriteTimeout:   apiConfig.WriteTimeout * 3,   // Más tiempo para generar reportes
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
		log.WithField("address", server.Addr).Info("Servidor API REPORTS iniciado")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.WithError(err).Fatal("Error iniciando servidor")
		}
	}()

	// Configurar graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Cerrando servidor API REPORTS...")

	// Timeout para shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.WithError(err).Error("Error durante shutdown del servidor")
	}

	log.Info("Servidor API REPORTS cerrado exitosamente")
}

// setupRoutes configura todas las rutas de la API Reports
func setupRoutes(
	router *gin.Engine,
	db *database.Database,
	log logger.Logger,
	validator validator.Validator,
	metrics *metrics.Metrics,
	cfg *config.APIConfig,
) {
	// Inicializar handlers específicos de reports
	reportsHandler := handlers.NewReportsHandler(db, log, validator, metrics)
	salesReportsHandler := handlers.NewSalesReportsHandler(db, log, validator, metrics)
	inventoryReportsHandler := handlers.NewInventoryReportsHandler(db, log, validator, metrics)
	financialReportsHandler := handlers.NewFinancialReportsHandler(db, log, validator, metrics)
	analyticsHandler := handlers.NewAnalyticsHandler(db, log, validator, metrics)
	dashboardHandler := handlers.NewDashboardHandler(db, log, validator, metrics)

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
			// Rutas de reportes generales
			reports := protected.Group("/reports")
			{
				// Gestión de reportes
				reports.GET("", reportsHandler.ListReports)
				reports.GET("/:id", reportsHandler.GetReport)
				reports.POST("", reportsHandler.CreateReport)
				reports.PUT("/:id", middleware.RequireRole("admin", "supervisor"), reportsHandler.UpdateReport)
				reports.DELETE("/:id", middleware.RequireRole("admin"), reportsHandler.DeleteReport)
				
				// Generación de reportes
				reports.POST("/:id/generate", reportsHandler.GenerateReport)
				reports.GET("/:id/status", reportsHandler.GetReportStatus)
				reports.GET("/:id/download", reportsHandler.DownloadReport)
				
				// Programación de reportes
				reports.POST("/:id/schedule", middleware.RequireRole("admin", "supervisor"), reportsHandler.ScheduleReport)
				reports.GET("/scheduled", reportsHandler.GetScheduledReports)
				reports.DELETE("/scheduled/:schedule_id", middleware.RequireRole("admin"), reportsHandler.DeleteScheduledReport)
				
				// Historial de reportes
				reports.GET("/history", reportsHandler.GetReportHistory)
				reports.GET("/:id/versions", reportsHandler.GetReportVersions)
			}

			// Rutas de reportes de ventas
			sales := protected.Group("/sales")
			{
				// Reportes básicos de ventas
				sales.GET("/summary", salesReportsHandler.GetSalesSummary)
				sales.GET("/daily", salesReportsHandler.GetDailySales)
				sales.GET("/weekly", salesReportsHandler.GetWeeklySales)
				sales.GET("/monthly", salesReportsHandler.GetMonthlySales)
				sales.GET("/yearly", salesReportsHandler.GetYearlySales)
				
				// Reportes detallados
				sales.GET("/by-product", salesReportsHandler.GetSalesByProduct)
				sales.GET("/by-category", salesReportsHandler.GetSalesByCategory)
				sales.GET("/by-user", salesReportsHandler.GetSalesByUser)
				sales.GET("/by-terminal", salesReportsHandler.GetSalesByTerminal)
				sales.GET("/by-payment-method", salesReportsHandler.GetSalesByPaymentMethod)
				
				// Análisis de ventas
				sales.GET("/trends", salesReportsHandler.GetSalesTrends)
				sales.GET("/forecasting", salesReportsHandler.GetSalesForecasting)
				sales.GET("/performance", salesReportsHandler.GetSalesPerformance)
				sales.GET("/comparison", salesReportsHandler.GetSalesComparison)
				
				// Reportes de documentos tributarios
				sales.GET("/dte", salesReportsHandler.GetDTEReports)
				sales.GET("/tax-summary", salesReportsHandler.GetTaxSummary)
			}

			// Rutas de reportes de inventario
			inventory := protected.Group("/inventory")
			{
				// Estado de inventario
				inventory.GET("/current-stock", inventoryReportsHandler.GetCurrentStock)
				inventory.GET("/low-stock", inventoryReportsHandler.GetLowStockReport)
				inventory.GET("/out-of-stock", inventoryReportsHandler.GetOutOfStockReport)
				inventory.GET("/overstock", inventoryReportsHandler.GetOverstockReport)
				
				// Movimientos de inventario
				inventory.GET("/movements", inventoryReportsHandler.GetStockMovements)
				inventory.GET("/adjustments", inventoryReportsHandler.GetStockAdjustments)
				inventory.GET("/transfers", inventoryReportsHandler.GetStockTransfers)
				
				// Análisis de inventario
				inventory.GET("/turnover", inventoryReportsHandler.GetInventoryTurnover)
				inventory.GET("/aging", inventoryReportsHandler.GetInventoryAging)
				inventory.GET("/valuation", inventoryReportsHandler.GetInventoryValuation)
				inventory.GET("/abc-analysis", inventoryReportsHandler.GetABCAnalysis)
				
				// Reportes de productos
				inventory.GET("/products/performance", inventoryReportsHandler.GetProductPerformance)
				inventory.GET("/products/profitability", inventoryReportsHandler.GetProductProfitability)
				inventory.GET("/categories/analysis", inventoryReportsHandler.GetCategoryAnalysis)
			}

			// Rutas de reportes financieros
			financial := protected.Group("/financial")
			{
				// Estados financieros básicos
				financial.GET("/profit-loss", financialReportsHandler.GetProfitLossReport)
				financial.GET("/cash-flow", financialReportsHandler.GetCashFlowReport)
				financial.GET("/balance-sheet", financialReportsHandler.GetBalanceSheetReport)
				
				// Análisis financiero
				financial.GET("/margins", financialReportsHandler.GetMarginAnalysis)
				financial.GET("/costs", financialReportsHandler.GetCostAnalysis)
				financial.GET("/profitability", financialReportsHandler.GetProfitabilityAnalysis)
				financial.GET("/roi", financialReportsHandler.GetROIAnalysis)
				
				// Reportes de pagos
				financial.GET("/payments", financialReportsHandler.GetPaymentReports)
				financial.GET("/receivables", financialReportsHandler.GetReceivablesReport)
				financial.GET("/payables", financialReportsHandler.GetPayablesReport)
				
				// Reportes tributarios
				financial.GET("/tax-reports", financialReportsHandler.GetTaxReports)
				financial.GET("/vat-summary", financialReportsHandler.GetVATSummary)
			}

			// Rutas de analytics avanzados
			analytics := protected.Group("/analytics")
			{
				// Analytics de clientes
				analytics.GET("/customers/behavior", analyticsHandler.GetCustomerBehavior)
				analytics.GET("/customers/segmentation", analyticsHandler.GetCustomerSegmentation)
				analytics.GET("/customers/lifetime-value", analyticsHandler.GetCustomerLifetimeValue)
				
				// Analytics de productos
				analytics.GET("/products/recommendations", analyticsHandler.GetProductRecommendations)
				analytics.GET("/products/cross-sell", analyticsHandler.GetCrossSellAnalysis)
				analytics.GET("/products/seasonality", analyticsHandler.GetSeasonalityAnalysis)
				
				// Analytics operacionales
				analytics.GET("/operations/efficiency", analyticsHandler.GetOperationalEfficiency)
				analytics.GET("/operations/bottlenecks", analyticsHandler.GetBottleneckAnalysis)
				analytics.GET("/operations/capacity", analyticsHandler.GetCapacityAnalysis)
				
				// Predictive analytics
				analytics.GET("/predictions/demand", analyticsHandler.GetDemandPrediction)
				analytics.GET("/predictions/churn", analyticsHandler.GetChurnPrediction)
				analytics.GET("/predictions/pricing", analyticsHandler.GetPricingOptimization)
			}

			// Rutas de dashboard
			dashboard := protected.Group("/dashboard")
			{
				// Dashboards principales
				dashboard.GET("/executive", dashboardHandler.GetExecutiveDashboard)
				dashboard.GET("/operations", dashboardHandler.GetOperationsDashboard)
				dashboard.GET("/sales", dashboardHandler.GetSalesDashboard)
				dashboard.GET("/inventory", dashboardHandler.GetInventoryDashboard)
				dashboard.GET("/financial", dashboardHandler.GetFinancialDashboard)
				
				// Widgets específicos
				dashboard.GET("/widgets/kpis", dashboardHandler.GetKPIWidgets)
				dashboard.GET("/widgets/charts", dashboardHandler.GetChartWidgets)
				dashboard.GET("/widgets/alerts", dashboardHandler.GetAlertWidgets)
				dashboard.GET("/widgets/recent-activity", dashboardHandler.GetRecentActivityWidget)
				
				// Configuración de dashboard
				dashboard.GET("/config", dashboardHandler.GetDashboardConfig)
				dashboard.PUT("/config", dashboardHandler.UpdateDashboardConfig)
				dashboard.POST("/widgets", dashboardHandler.CreateCustomWidget)
				dashboard.PUT("/widgets/:id", dashboardHandler.UpdateCustomWidget)
				dashboard.DELETE("/widgets/:id", dashboardHandler.DeleteCustomWidget)
			}

			// Rutas de exportación
			export := protected.Group("/export")
			{
				export.POST("/excel", reportsHandler.ExportToExcel)
				export.POST("/pdf", reportsHandler.ExportToPDF)
				export.POST("/csv", reportsHandler.ExportToCSV)
				export.POST("/json", reportsHandler.ExportToJSON)
				
				// Exportación programada
				export.POST("/schedule", middleware.RequireRole("admin", "supervisor"), reportsHandler.ScheduleExport)
				export.GET("/scheduled", reportsHandler.GetScheduledExports)
			}

			// Rutas de configuración de reportes
			config := protected.Group("/config")
			{
				config.GET("/templates", reportsHandler.GetReportTemplates)
				config.POST("/templates", middleware.RequireRole("admin", "supervisor"), reportsHandler.CreateReportTemplate)
				config.PUT("/templates/:id", middleware.RequireRole("admin", "supervisor"), reportsHandler.UpdateReportTemplate)
				config.DELETE("/templates/:id", middleware.RequireRole("admin"), reportsHandler.DeleteReportTemplate)
				
				// Configuración de métricas
				config.GET("/metrics", reportsHandler.GetMetricsConfig)
				config.PUT("/metrics", middleware.RequireRole("admin"), reportsHandler.UpdateMetricsConfig)
				
				// Configuración de alertas
				config.GET("/alerts", reportsHandler.GetAlertsConfig)
				config.PUT("/alerts", middleware.RequireRole("admin", "supervisor"), reportsHandler.UpdateAlertsConfig)
			}
		}

		// Rutas públicas (para integración con sistemas externos)
		public := v1.Group("/public")
		{
			public.GET("/report-types", reportsHandler.GetReportTypes)
			public.GET("/export-formats", reportsHandler.GetExportFormats)
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

