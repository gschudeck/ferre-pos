package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"ferre-pos-servidor-central/internal/config"
	"ferre-pos-servidor-central/internal/controllers"
	"ferre-pos-servidor-central/internal/database"
	"ferre-pos-servidor-central/internal/middleware"
)

func main() {
	// Configurar logging básico
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Iniciando Servidor Central Ferre-POS...")

	// Cargar configuración
	configPath := getConfigPath()
	configManager := config.NewConfigManager(configPath)
	
	if err := configManager.LoadConfig(); err != nil {
		log.Fatalf("Error cargando configuración: %v", err)
	}

	if err := configManager.ValidateConfig(); err != nil {
		log.Fatalf("Configuración inválida: %v", err)
	}

	cfg := configManager.GetConfig()
	log.Printf("Configuración cargada desde: %s", configPath)

	// Configurar modo de Gin
	gin.SetMode(cfg.Server.Mode)

	// Inicializar sistema de recarga en caliente
	hotReloadManager, err := config.NewHotReloadManager(configManager)
	if err != nil {
		log.Fatalf("Error inicializando recarga en caliente: %v", err)
	}

	if err := hotReloadManager.Start(); err != nil {
		log.Fatalf("Error iniciando recarga en caliente: %v", err)
	}
	defer hotReloadManager.Stop()

	// Inicializar base de datos
	dbManager, err := database.NewDatabaseManager(cfg.Database)
	if err != nil {
		log.Fatalf("Error inicializando base de datos: %v", err)
	}
	defer dbManager.Close()

	// Verificar conexiones de base de datos
	if err := dbManager.HealthCheck(); err != nil {
		log.Fatalf("Error en health check de base de datos: %v", err)
	}
	log.Println("Conexiones de base de datos establecidas correctamente")

	// Crear router principal
	router := gin.New()

	// Configurar middleware global
	setupGlobalMiddleware(router, cfg)

	// Configurar rutas por API
	setupAPIRoutes(router, cfg, dbManager)

	// Crear servidor HTTP
	server := &http.Server{
		Addr:           fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:        router,
		ReadTimeout:    cfg.Server.ReadTimeout,
		WriteTimeout:   cfg.Server.WriteTimeout,
		IdleTimeout:    cfg.Server.IdleTimeout,
		MaxHeaderBytes: cfg.Server.MaxHeaderBytes,
	}

	// Configurar trusted proxies
	if len(cfg.Server.TrustedProxies) > 0 {
		router.SetTrustedProxies(cfg.Server.TrustedProxies)
	}

	// Iniciar servidor en goroutine
	go func() {
		log.Printf("Servidor iniciando en %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error iniciando servidor: %v", err)
		}
	}()

	// Configurar graceful shutdown
	setupGracefulShutdown(server, cfg.Server.GracefulTimeout)

	log.Println("Servidor Central Ferre-POS detenido correctamente")
}

// getConfigPath obtiene la ruta del archivo de configuración
func getConfigPath() string {
	if configPath := os.Getenv("FERRE_POS_CONFIG"); configPath != "" {
		return configPath
	}
	return "configs/config.yaml"
}

// setupGlobalMiddleware configura middleware global
func setupGlobalMiddleware(router *gin.Engine, cfg *config.Config) {
	// Recovery middleware
	router.Use(gin.Recovery())

	// Request ID middleware
	router.Use(middleware.RequestID())

	// Response time middleware
	router.Use(middleware.ResponseTimeMiddleware())

	// Security headers middleware
	router.Use(middleware.SecurityHeaders())

	// Health check middleware
	router.Use(middleware.HealthCheckMiddleware())

	// Rate limiting global si está habilitado
	if cfg.Security.RateLimiting.Enabled {
		router.Use(middleware.RateLimitMiddleware(cfg.Security.RateLimiting.RequestsPerMinute))
	}

	// Logging middleware global
	// En una implementación real, se pasaría el servicio de logging
	router.Use(middleware.LoggingMiddleware(cfg.Logging.Global, nil))

	// Metrics middleware si está habilitado
	if cfg.Monitoring.Enabled {
		// En una implementación real, se pasaría el servicio de métricas
		router.Use(middleware.MetricsMiddleware(nil))
	}

	log.Println("Middleware global configurado")
}

// setupAPIRoutes configura las rutas por API
func setupAPIRoutes(router *gin.Engine, cfg *config.Config, dbManager *database.DatabaseManager) {
	// Crear middleware de autenticación
	authMiddleware := middleware.NewAuthMiddleware(nil, cfg.Security.JWTSecret)

	// Configurar API POS
	if cfg.APIs.POS.Enabled {
		setupPOSAPI(router, cfg, dbManager, authMiddleware)
	}

	// Configurar API Sync
	if cfg.APIs.Sync.Enabled {
		setupSyncAPI(router, cfg, dbManager, authMiddleware)
	}

	// Configurar API Labels
	if cfg.APIs.Labels.Enabled {
		setupLabelsAPI(router, cfg, dbManager, authMiddleware)
	}

	// Configurar API Reports
	if cfg.APIs.Reports.Enabled {
		setupReportsAPI(router, cfg, dbManager, authMiddleware)
	}

	// Ruta de información del servidor
	router.GET("/info", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"name":    "Ferre-POS Servidor Central",
			"version": "1.0.0",
			"apis": gin.H{
				"pos":     cfg.APIs.POS.Enabled,
				"sync":    cfg.APIs.Sync.Enabled,
				"labels":  cfg.APIs.Labels.Enabled,
				"reports": cfg.APIs.Reports.Enabled,
			},
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})

	log.Println("Rutas de API configuradas")
}

// setupPOSAPI configura el API POS
func setupPOSAPI(router *gin.Engine, cfg *config.Config, dbManager *database.DatabaseManager, authMiddleware *middleware.AuthMiddleware) {
	posGroup := router.Group(cfg.APIs.POS.BasePath)

	// CORS específico para POS
	posGroup.Use(middleware.CORS(cfg.Security.CORS.POS))

	// Rate limiting específico para POS
	if cfg.APIs.POS.RateLimiting.Enabled {
		posGroup.Use(middleware.RateLimitMiddleware(cfg.APIs.POS.RateLimiting.RequestsPerMinute))
	}

	// Logging específico para POS
	posGroup.Use(middleware.LoggingMiddleware(cfg.Logging.POS, nil))

	// Crear controlador POS
	posController := controllers.NewPOSController(
		dbManager.GetConnection("pos"),
		cfg.APIs.POS,
	)

	// Rutas públicas (sin autenticación)
	posGroup.POST("/auth/login", posController.Login)
	posGroup.POST("/auth/refresh", posController.RefreshToken)
	posGroup.GET("/health", posController.Health)

	// Rutas protegidas (con autenticación)
	protected := posGroup.Group("")
	protected.Use(authMiddleware.RequireAuth())

	// Rutas de productos
	products := protected.Group("/productos")
	products.GET("", posController.GetProductos)
	products.GET("/:id", posController.GetProducto)
	products.GET("/buscar", posController.BuscarProductos)
	products.GET("/barcode/:codigo", posController.GetProductoPorBarcode)

	// Rutas de stock
	stock := protected.Group("/stock")
	stock.GET("", posController.GetStock)
	stock.GET("/:producto_id", posController.GetStockProducto)
	stock.POST("/reservar", posController.ReservarStock)
	stock.POST("/liberar", posController.LiberarStock)

	// Rutas de ventas
	ventas := protected.Group("/ventas")
	ventas.Use(authMiddleware.RequireTerminal())
	ventas.POST("", posController.CrearVenta)
	ventas.GET("/:id", posController.GetVenta)
	ventas.POST("/:id/anular", posController.AnularVenta)
	ventas.GET("", posController.GetVentas)

	// Rutas de clientes
	clientes := protected.Group("/clientes")
	clientes.GET("", posController.GetClientes)
	clientes.GET("/:id", posController.GetCliente)
	clientes.POST("", posController.CrearCliente)
	clientes.PUT("/:id", posController.ActualizarCliente)

	log.Printf("API POS configurada en %s", cfg.APIs.POS.BasePath)
}

// setupSyncAPI configura el API Sync
func setupSyncAPI(router *gin.Engine, cfg *config.Config, dbManager *database.DatabaseManager, authMiddleware *middleware.AuthMiddleware) {
	syncGroup := router.Group(cfg.APIs.Sync.BasePath)

	// CORS específico para Sync
	syncGroup.Use(middleware.CORS(cfg.Security.CORS.Sync))

	// Logging específico para Sync
	syncGroup.Use(middleware.LoggingMiddleware(cfg.Logging.Sync, nil))

	// Crear controlador Sync
	syncController := controllers.NewSyncController(
		dbManager.GetConnection("sync"),
		cfg.APIs.Sync,
	)

	// Rutas públicas
	syncGroup.GET("/health", syncController.Health)

	// Rutas protegidas
	protected := syncGroup.Group("")
	protected.Use(authMiddleware.RequireAuth())

	// Rutas de sincronización
	protected.POST("/iniciar", syncController.IniciarSincronizacion)
	protected.GET("/estado/:id", syncController.GetEstadoSincronizacion)
	protected.POST("/detener/:id", syncController.DetenerSincronizacion)
	protected.GET("/historial", syncController.GetHistorialSincronizacion)

	// Rutas de conflictos
	conflicts := protected.Group("/conflictos")
	conflicts.GET("", syncController.GetConflictos)
	conflicts.POST("/:id/resolver", syncController.ResolverConflicto)
	conflicts.GET("/:id", syncController.GetConflicto)

	// Rutas de configuración
	config := protected.Group("/configuracion")
	config.Use(authMiddleware.RequireRole("admin"))
	config.GET("", syncController.GetConfiguracion)
	config.PUT("", syncController.ActualizarConfiguracion)

	log.Printf("API Sync configurada en %s", cfg.APIs.Sync.BasePath)
}

// setupLabelsAPI configura el API Labels
func setupLabelsAPI(router *gin.Engine, cfg *config.Config, dbManager *database.DatabaseManager, authMiddleware *middleware.AuthMiddleware) {
	labelsGroup := router.Group(cfg.APIs.Labels.BasePath)

	// CORS específico para Labels
	labelsGroup.Use(middleware.CORS(cfg.Security.CORS.Labels))

	// Logging específico para Labels
	labelsGroup.Use(middleware.LoggingMiddleware(cfg.Logging.Labels, nil))

	// Crear controlador Labels
	labelsController := controllers.NewLabelsController(
		dbManager.GetConnection("labels"),
		cfg.APIs.Labels,
	)

	// Rutas públicas
	labelsGroup.GET("/health", labelsController.Health)

	// Rutas protegidas
	protected := labelsGroup.Group("")
	protected.Use(authMiddleware.RequireAuth())

	// Rutas de plantillas
	templates := protected.Group("/plantillas")
	templates.GET("", labelsController.GetPlantillas)
	templates.POST("", labelsController.CrearPlantilla)
	templates.GET("/:id", labelsController.GetPlantilla)
	templates.PUT("/:id", labelsController.ActualizarPlantilla)
	templates.DELETE("/:id", labelsController.EliminarPlantilla)

	// Rutas de generación
	generation := protected.Group("/generar")
	generation.POST("/individual", labelsController.GenerarEtiquetaIndividual)
	generation.POST("/lote", labelsController.GenerarLoteEtiquetas)
	generation.GET("/estado/:job_id", labelsController.GetEstadoGeneracion)
	generation.GET("/descargar/:job_id", labelsController.DescargarEtiquetas)

	// Rutas de preview
	preview := protected.Group("/preview")
	preview.POST("", labelsController.GenerarPreview)

	log.Printf("API Labels configurada en %s", cfg.APIs.Labels.BasePath)
}

// setupReportsAPI configura el API Reports
func setupReportsAPI(router *gin.Engine, cfg *config.Config, dbManager *database.DatabaseManager, authMiddleware *middleware.AuthMiddleware) {
	reportsGroup := router.Group(cfg.APIs.Reports.BasePath)

	// CORS específico para Reports
	reportsGroup.Use(middleware.CORS(cfg.Security.CORS.Reports))

	// Logging específico para Reports
	reportsGroup.Use(middleware.LoggingMiddleware(cfg.Logging.Reports, nil))

	// Crear controlador Reports
	reportsController := controllers.NewReportsController(
		dbManager.GetConnection("reports"),
		cfg.APIs.Reports,
	)

	// Rutas públicas
	reportsGroup.GET("/health", reportsController.Health)

	// Rutas protegidas
	protected := reportsGroup.Group("")
	protected.Use(authMiddleware.RequireAuth())

	// Rutas de plantillas
	templates := protected.Group("/plantillas")
	templates.GET("", reportsController.GetPlantillas)
	templates.POST("", reportsController.CrearPlantilla)
	templates.GET("/:id", reportsController.GetPlantilla)
	templates.PUT("/:id", reportsController.ActualizarPlantilla)
	templates.DELETE("/:id", reportsController.EliminarPlantilla)

	// Rutas de generación
	generation := protected.Group("/generar")
	generation.POST("", reportsController.GenerarReporte)
	generation.GET("/estado/:job_id", reportsController.GetEstadoGeneracion)
	generation.GET("/descargar/:job_id", reportsController.DescargarReporte)

	// Rutas de reportes programados
	scheduled := protected.Group("/programados")
	scheduled.GET("", reportsController.GetReportesProgramados)
	scheduled.POST("", reportsController.CrearReporteProgramado)
	scheduled.GET("/:id", reportsController.GetReporteProgramado)
	scheduled.PUT("/:id", reportsController.ActualizarReporteProgramado)
	scheduled.DELETE("/:id", reportsController.EliminarReporteProgramado)

	// Rutas de dashboards
	dashboards := protected.Group("/dashboards")
	dashboards.GET("", reportsController.GetDashboards)
	dashboards.POST("", reportsController.CrearDashboard)
	dashboards.GET("/:id", reportsController.GetDashboard)
	dashboards.PUT("/:id", reportsController.ActualizarDashboard)
	dashboards.DELETE("/:id", reportsController.EliminarDashboard)
	dashboards.GET("/:id/data", reportsController.GetDatosDashboard)

	log.Printf("API Reports configurada en %s", cfg.APIs.Reports.BasePath)
}

// setupGracefulShutdown configura el shutdown graceful del servidor
func setupGracefulShutdown(server *http.Server, timeout time.Duration) {
	// Canal para recibir señales del sistema
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Esperar señal
	<-quit
	log.Println("Recibida señal de shutdown, deteniendo servidor...")

	// Crear contexto con timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Intentar shutdown graceful
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Error durante shutdown graceful: %v", err)
		log.Println("Forzando cierre del servidor...")
		server.Close()
	}
}

