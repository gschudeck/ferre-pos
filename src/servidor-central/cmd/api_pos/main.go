package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"ferre-pos-servidor-central/internal/config"
	"ferre-pos-servidor-central/internal/controllers"
	"ferre-pos-servidor-central/internal/database"
	"ferre-pos-servidor-central/internal/middleware"
	"ferre-pos-servidor-central/internal/services"
	"ferre-pos-servidor-central/pkg/errors"
	"ferre-pos-servidor-central/pkg/logger"
	"ferre-pos-servidor-central/pkg/metrics"
	"ferre-pos-servidor-central/pkg/validator"
)

// structAppInfo contiene información de la aplicación
type structAppInfo struct {
	strName        string
	strVersion     string
	strDescription string
	strAPIPath     string
}

// structPOSServer representa el servidor API POS
type structPOSServer struct {
	ptrLogger     *zap.Logger
	ptrConfig     *config.Config
	ptrDBManager  *database.DatabaseManager
	ptrHTTPServer *http.Server
	ptrMetrics    *metrics.Manager
	ptrValidator  *validator.Validator
	structAppInfo structAppInfo
}

var (
	// Información de la aplicación con notación húngara
	gStructAppInfo = structAppInfo{
		strName:        "Ferre-POS API POS",
		strVersion:     "1.0.0",
		strDescription: "API REST para operaciones de punto de venta",
		strAPIPath:     "/api/pos",
	}
)

func main() {
	// Configurar logging inicial
	ptrInitialLogger := setupInitialLogger()
	defer ptrInitialLogger.Sync()

	ptrInitialLogger.Info("Iniciando API POS",
		zap.String("name", gStructAppInfo.strName),
		zap.String("version", gStructAppInfo.strVersion),
	)

	// Crear instancia del servidor
	ptrServer, errCreate := createPOSServer(ptrInitialLogger)
	if errCreate != nil {
		ptrInitialLogger.Fatal("Error creando servidor POS",
			zap.Error(errCreate),
		)
	}

	// Inicializar servidor
	if errInit := ptrServer.initialize(); errInit != nil {
		ptrServer.ptrLogger.Fatal("Error inicializando servidor",
			zap.Error(errInit),
		)
	}

	// Iniciar servidor
	if errStart := ptrServer.start(); errStart != nil {
		ptrServer.ptrLogger.Fatal("Error iniciando servidor",
			zap.Error(errStart),
		)
	}

	// Configurar graceful shutdown
	ptrServer.setupGracefulShutdown()

	ptrServer.ptrLogger.Info("API POS detenido correctamente")
}

// setupInitialLogger configura el logger inicial
func setupInitialLogger() *zap.Logger {
	configLogger := zap.NewProductionConfig()
	configLogger.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	configLogger.OutputPaths = []string{"stdout"}
	configLogger.ErrorOutputPaths = []string{"stderr"}

	ptrLogger, err := configLogger.Build()
	if err != nil {
		log.Fatalf("Error configurando logger inicial: %v", err)
	}

	return ptrLogger
}

// createPOSServer crea una nueva instancia del servidor POS
func createPOSServer(ptrInitialLogger *zap.Logger) (*structPOSServer, error) {
	return &structPOSServer{
		ptrLogger:     ptrInitialLogger,
		structAppInfo: gStructAppInfo,
	}, nil
}

// initialize inicializa todos los componentes del servidor
func (ptrServer *structPOSServer) initialize() error {
	var err error

	// Cargar configuración
	strConfigPath := getConfigPath()
	ptrConfigManager := config.NewConfigManager(strConfigPath)

	if err = ptrConfigManager.LoadConfig(); err != nil {
		return errors.Wrap(err, "error cargando configuración")
	}

	if err = ptrConfigManager.ValidateConfig(); err != nil {
		return errors.Wrap(err, "configuración inválida")
	}

	ptrServer.ptrConfig = ptrConfigManager.GetConfig()

	// Configurar logger mejorado
	ptrServer.ptrLogger, err = logger.NewLogger(&logger.Config{
		Level:      ptrServer.ptrConfig.Logging.POS.Level,
		Format:     ptrServer.ptrConfig.Logging.POS.Format,
		Output:     ptrServer.ptrConfig.Logging.POS.Output,
		FilePath:   ptrServer.ptrConfig.Logging.POS.FilePath,
		MaxSize:    ptrServer.ptrConfig.Logging.POS.MaxSize,
		MaxBackups: ptrServer.ptrConfig.Logging.POS.MaxBackups,
		MaxAge:     ptrServer.ptrConfig.Logging.POS.MaxAge,
		Compress:   ptrServer.ptrConfig.Logging.POS.Compress,
	})
	if err != nil {
		return errors.Wrap(err, "error configurando logger")
	}

	ptrServer.ptrLogger.Info("Configuración cargada",
		zap.String("config_path", strConfigPath),
		zap.String("api_path", ptrServer.structAppInfo.strAPIPath),
	)

	// Inicializar métricas
	ptrServer.ptrMetrics, err = metrics.NewManager(&metrics.Config{
		Enabled:     ptrServer.ptrConfig.Monitoring.Enabled,
		Namespace:   "ferre_pos",
		Subsystem:   "api_pos",
		MetricsPath: ptrServer.ptrConfig.Monitoring.MetricsPath,
	})
	if err != nil {
		return errors.Wrap(err, "error inicializando métricas")
	}

	// Inicializar validador
	ptrServer.ptrValidator = validator.New()

	// Inicializar base de datos
	ptrServer.ptrDBManager, err = database.NewDatabaseManager(ptrServer.ptrConfig.Database)
	if err != nil {
		return errors.Wrap(err, "error inicializando base de datos")
	}

	// Verificar conexión de base de datos
	if err = ptrServer.ptrDBManager.HealthCheck(); err != nil {
		return errors.Wrap(err, "error en health check de base de datos")
	}

	ptrServer.ptrLogger.Info("Base de datos inicializada correctamente")

	return nil
}

// start inicia el servidor HTTP
func (ptrServer *structPOSServer) start() error {
	// Configurar modo de Gin
	gin.SetMode(ptrServer.ptrConfig.Server.Mode)

	// Crear router
	ptrRouter := gin.New()

	// Configurar middleware
	ptrServer.setupMiddleware(ptrRouter)

	// Configurar rutas
	ptrServer.setupRoutes(ptrRouter)

	// Crear servidor HTTP
	ptrServer.ptrHTTPServer = &http.Server{
		Addr:           fmt.Sprintf("%s:%d", ptrServer.ptrConfig.Server.Host, ptrServer.ptrConfig.APIs.POS.Port),
		Handler:        ptrRouter,
		ReadTimeout:    ptrServer.ptrConfig.Server.ReadTimeout,
		WriteTimeout:   ptrServer.ptrConfig.Server.WriteTimeout,
		IdleTimeout:    ptrServer.ptrConfig.Server.IdleTimeout,
		MaxHeaderBytes: ptrServer.ptrConfig.Server.MaxHeaderBytes,
	}

	// Configurar trusted proxies
	if len(ptrServer.ptrConfig.Server.TrustedProxies) > 0 {
		ptrRouter.SetTrustedProxies(ptrServer.ptrConfig.Server.TrustedProxies)
	}

	// Iniciar servidor en goroutine
	go func() {
		ptrServer.ptrLogger.Info("Servidor API POS iniciando",
			zap.String("address", ptrServer.ptrHTTPServer.Addr),
			zap.String("api_path", ptrServer.structAppInfo.strAPIPath),
		)

		if err := ptrServer.ptrHTTPServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			ptrServer.ptrLogger.Fatal("Error iniciando servidor HTTP",
				zap.Error(err),
			)
		}
	}()

	return nil
}

// setupMiddleware configura el middleware del servidor
func (ptrServer *structPOSServer) setupMiddleware(ptrRouter *gin.Engine) {
	// Recovery middleware con logging
	ptrRouter.Use(middleware.RecoveryWithZap(ptrServer.ptrLogger, true))

	// Request ID middleware
	ptrRouter.Use(middleware.RequestID())

	// Logging middleware
	ptrRouter.Use(middleware.LoggingWithZap(ptrServer.ptrLogger))

	// Metrics middleware
	if ptrServer.ptrConfig.Monitoring.Enabled {
		ptrRouter.Use(middleware.MetricsMiddleware(ptrServer.ptrMetrics))
	}

	// CORS middleware
	ptrRouter.Use(middleware.CORS(ptrServer.ptrConfig.Security.CORS.POS))

	// Rate limiting middleware
	if ptrServer.ptrConfig.APIs.POS.RateLimiting.Enabled {
		ptrRouter.Use(middleware.RateLimitMiddleware(&middleware.RateLimitConfig{
			RequestsPerMinute: ptrServer.ptrConfig.APIs.POS.RateLimiting.RequestsPerMinute,
			BurstSize:         ptrServer.ptrConfig.APIs.POS.RateLimiting.BurstSize,
			KeyFunc:           middleware.IPKeyFunc,
		}))
	}

	// Security headers middleware
	ptrRouter.Use(middleware.SecurityHeaders())

	// Timeout middleware
	ptrRouter.Use(middleware.TimeoutMiddleware(ptrServer.ptrConfig.APIs.POS.RequestTimeout))

	ptrServer.ptrLogger.Info("Middleware configurado correctamente")
}

// setupRoutes configura las rutas del API POS
func (ptrServer *structPOSServer) setupRoutes(ptrRouter *gin.Engine) {
	// Crear servicios
	ptrPOSService := services.NewPOSService(
		ptrServer.ptrDBManager.GetConnection("pos"),
		ptrServer.ptrLogger,
		ptrServer.ptrValidator,
		ptrServer.ptrMetrics,
	)

	// Crear controlador
	ptrPOSController := controllers.NewPOSController(
		ptrPOSService,
		ptrServer.ptrLogger,
		ptrServer.ptrValidator,
		ptrServer.ptrConfig.APIs.POS,
	)

	// Crear middleware de autenticación
	ptrAuthMiddleware := middleware.NewAuthMiddleware(&middleware.AuthConfig{
		JWTSecret:            ptrServer.ptrConfig.Security.JWTSecret,
		JWTExpiration:        ptrServer.ptrConfig.Security.JWTExpiration,
		RefreshExpiration:    ptrServer.ptrConfig.Security.RefreshTokenExpiration,
		MaxLoginAttempts:     ptrServer.ptrConfig.Security.MaxLoginAttempts,
		LoginLockoutDuration: ptrServer.ptrConfig.Security.LoginLockoutDuration,
	}, ptrServer.ptrLogger)

	// Grupo de rutas API POS
	apiGroup := ptrRouter.Group(ptrServer.structAppInfo.strAPIPath)

	// Rutas públicas
	apiGroup.GET("/health", ptrPOSController.Health)
	apiGroup.GET("/info", ptrPOSController.Info)

	// Rutas de autenticación
	authGroup := apiGroup.Group("/auth")
	{
		authGroup.POST("/login", ptrPOSController.Login)
		authGroup.POST("/refresh", ptrPOSController.RefreshToken)
		authGroup.POST("/logout", ptrAuthMiddleware.RequireAuth(), ptrPOSController.Logout)
	}

	// Rutas protegidas
	protectedGroup := apiGroup.Group("")
	protectedGroup.Use(ptrAuthMiddleware.RequireAuth())

	// Rutas de productos
	productosGroup := protectedGroup.Group("/productos")
	{
		productosGroup.GET("", ptrPOSController.GetProductos)
		productosGroup.GET("/:id", ptrPOSController.GetProducto)
		productosGroup.GET("/buscar", ptrPOSController.BuscarProductos)
		productosGroup.GET("/barcode/:codigo", ptrPOSController.GetProductoPorBarcode)
	}

	// Rutas de stock
	stockGroup := protectedGroup.Group("/stock")
	{
		stockGroup.GET("", ptrPOSController.GetStock)
		stockGroup.GET("/:producto_id", ptrPOSController.GetStockProducto)
		stockGroup.POST("/reservar", ptrPOSController.ReservarStock)
		stockGroup.POST("/liberar", ptrPOSController.LiberarStock)
	}

	// Rutas de ventas (requieren terminal)
	ventasGroup := protectedGroup.Group("/ventas")
	ventasGroup.Use(ptrAuthMiddleware.RequireTerminal())
	{
		ventasGroup.POST("", ptrPOSController.CrearVenta)
		ventasGroup.GET("/:id", ptrPOSController.GetVenta)
		ventasGroup.POST("/:id/anular", ptrPOSController.AnularVenta)
		ventasGroup.GET("", ptrPOSController.GetVentas)
	}

	// Rutas de clientes
	clientesGroup := protectedGroup.Group("/clientes")
	{
		clientesGroup.GET("", ptrPOSController.GetClientes)
		clientesGroup.GET("/:id", ptrPOSController.GetCliente)
		clientesGroup.POST("", ptrPOSController.CrearCliente)
		clientesGroup.PUT("/:id", ptrPOSController.ActualizarCliente)
	}

	// Rutas de métricas (protegidas)
	if ptrServer.ptrConfig.Monitoring.Enabled {
		metricsGroup := ptrRouter.Group("/metrics")
		metricsGroup.Use(middleware.IPWhitelist([]string{"127.0.0.1", "::1"}))
		metricsGroup.GET("", ptrServer.ptrMetrics.Handler())
	}

	ptrServer.ptrLogger.Info("Rutas configuradas correctamente",
		zap.String("api_path", ptrServer.structAppInfo.strAPIPath),
		zap.Int("total_routes", len(ptrRouter.Routes())),
	)
}

// setupGracefulShutdown configura el shutdown graceful del servidor
func (ptrServer *structPOSServer) setupGracefulShutdown() {
	// Canal para recibir señales del sistema
	chanQuit := make(chan os.Signal, 1)
	signal.Notify(chanQuit, syscall.SIGINT, syscall.SIGTERM)

	// Esperar señal
	<-chanQuit
	ptrServer.ptrLogger.Info("Recibida señal de shutdown, deteniendo servidor...")

	// Crear contexto con timeout
	ctxShutdown, funcCancel := context.WithTimeout(context.Background(), ptrServer.ptrConfig.Server.GracefulTimeout)
	defer funcCancel()

	// Intentar shutdown graceful
	if err := ptrServer.ptrHTTPServer.Shutdown(ctxShutdown); err != nil {
		ptrServer.ptrLogger.Error("Error durante shutdown graceful",
			zap.Error(err),
		)
		ptrServer.ptrLogger.Info("Forzando cierre del servidor...")
		ptrServer.ptrHTTPServer.Close()
	}

	// Cerrar conexiones de base de datos
	if ptrServer.ptrDBManager != nil {
		ptrServer.ptrDBManager.Close()
	}

	// Cerrar métricas
	if ptrServer.ptrMetrics != nil {
		ptrServer.ptrMetrics.Close()
	}
}

// getConfigPath obtiene la ruta del archivo de configuración
func getConfigPath() string {
	if strConfigPath := os.Getenv("FERRE_POS_CONFIG"); strConfigPath != "" {
		return strConfigPath
	}
	return "configs/config.yaml"
}
