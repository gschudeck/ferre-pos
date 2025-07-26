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

// structLabelsServer representa el servidor API Labels
type structLabelsServer struct {
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
		strName:        "Ferre-POS API Labels",
		strVersion:     "1.0.0",
		strDescription: "API REST para generación de etiquetas",
		strAPIPath:     "/api/labels",
	}
)

func main() {
	// Configurar logging inicial
	ptrInitialLogger := setupInitialLogger()
	defer ptrInitialLogger.Sync()

	ptrInitialLogger.Info("Iniciando API Labels",
		zap.String("name", gStructAppInfo.strName),
		zap.String("version", gStructAppInfo.strVersion),
	)

	// Crear instancia del servidor
	ptrServer, errCreate := createLabelsServer(ptrInitialLogger)
	if errCreate != nil {
		ptrInitialLogger.Fatal("Error creando servidor Labels",
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

	ptrServer.ptrLogger.Info("API Labels detenido correctamente")
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

// createLabelsServer crea una nueva instancia del servidor Labels
func createLabelsServer(ptrInitialLogger *zap.Logger) (*structLabelsServer, error) {
	return &structLabelsServer{
		ptrLogger:     ptrInitialLogger,
		structAppInfo: gStructAppInfo,
	}, nil
}

// initialize inicializa todos los componentes del servidor
func (ptrServer *structLabelsServer) initialize() error {
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
		Level:      ptrServer.ptrConfig.Logging.Labels.Level,
		Format:     ptrServer.ptrConfig.Logging.Labels.Format,
		Output:     ptrServer.ptrConfig.Logging.Labels.Output,
		FilePath:   ptrServer.ptrConfig.Logging.Labels.FilePath,
		MaxSize:    ptrServer.ptrConfig.Logging.Labels.MaxSize,
		MaxBackups: ptrServer.ptrConfig.Logging.Labels.MaxBackups,
		MaxAge:     ptrServer.ptrConfig.Logging.Labels.MaxAge,
		Compress:   ptrServer.ptrConfig.Logging.Labels.Compress,
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
		Subsystem:   "api_labels",
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
func (ptrServer *structLabelsServer) start() error {
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
		Addr:           fmt.Sprintf("%s:%d", ptrServer.ptrConfig.Server.Host, ptrServer.ptrConfig.APIs.Labels.Port),
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
		ptrServer.ptrLogger.Info("Servidor API Labels iniciando",
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
func (ptrServer *structLabelsServer) setupMiddleware(ptrRouter *gin.Engine) {
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
	ptrRouter.Use(middleware.CORS(ptrServer.ptrConfig.Security.CORS.Labels))

	// Rate limiting middleware
	if ptrServer.ptrConfig.APIs.Labels.RateLimiting.Enabled {
		ptrRouter.Use(middleware.RateLimitMiddleware(&middleware.RateLimitConfig{
			RequestsPerMinute: ptrServer.ptrConfig.APIs.Labels.RateLimiting.RequestsPerMinute,
			BurstSize:         ptrServer.ptrConfig.APIs.Labels.RateLimiting.BurstSize,
			KeyFunc:           middleware.IPKeyFunc,
		}))
	}

	// Security headers middleware
	ptrRouter.Use(middleware.SecurityHeaders())

	// Timeout middleware
	ptrRouter.Use(middleware.TimeoutMiddleware(ptrServer.ptrConfig.APIs.Labels.RequestTimeout))

	ptrServer.ptrLogger.Info("Middleware configurado correctamente")
}

// setupRoutes configura las rutas del API Labels
func (ptrServer *structLabelsServer) setupRoutes(ptrRouter *gin.Engine) {
	// Crear servicios
	ptrLabelsService := services.NewLabelsService(
		ptrServer.ptrDBManager.GetConnection("labels"),
		ptrServer.ptrLogger,
		ptrServer.ptrValidator,
		ptrServer.ptrMetrics,
	)

	// Crear controlador
	ptrLabelsController := controllers.NewLabelsController(
		ptrLabelsService,
		ptrServer.ptrLogger,
		ptrServer.ptrValidator,
		ptrServer.ptrConfig.APIs.Labels,
	)

	// Crear middleware de autenticación
	ptrAuthMiddleware := middleware.NewAuthMiddleware(&middleware.AuthConfig{
		JWTSecret:            ptrServer.ptrConfig.Security.JWTSecret,
		JWTExpiration:        ptrServer.ptrConfig.Security.JWTExpiration,
		RefreshExpiration:    ptrServer.ptrConfig.Security.RefreshTokenExpiration,
		MaxLoginAttempts:     ptrServer.ptrConfig.Security.MaxLoginAttempts,
		LoginLockoutDuration: ptrServer.ptrConfig.Security.LoginLockoutDuration,
	}, ptrServer.ptrLogger)

	// Grupo de rutas API Labels
	apiGroup := ptrRouter.Group(ptrServer.structAppInfo.strAPIPath)

	// Rutas públicas
	apiGroup.GET("/health", ptrLabelsController.Health)
	apiGroup.GET("/info", ptrLabelsController.Info)

	// Rutas protegidas
	protectedGroup := apiGroup.Group("")
	protectedGroup.Use(ptrAuthMiddleware.RequireAuth())

	// Rutas de plantillas
	plantillasGroup := protectedGroup.Group("/plantillas")
	{
		plantillasGroup.GET("", ptrLabelsController.GetPlantillas)
		plantillasGroup.POST("", ptrLabelsController.CrearPlantilla)
		plantillasGroup.GET("/:id", ptrLabelsController.GetPlantilla)
		plantillasGroup.PUT("/:id", ptrLabelsController.ActualizarPlantilla)
		plantillasGroup.DELETE("/:id", ptrLabelsController.EliminarPlantilla)
		plantillasGroup.POST("/:id/duplicar", ptrLabelsController.DuplicarPlantilla)
		plantillasGroup.GET("/:id/preview", ptrLabelsController.PreviewPlantilla)
	}

	// Rutas de generación
	generacionGroup := protectedGroup.Group("/generar")
	{
		generacionGroup.POST("/individual", ptrLabelsController.GenerarEtiquetaIndividual)
		generacionGroup.POST("/lote", ptrLabelsController.GenerarLoteEtiquetas)
		generacionGroup.GET("/estado/:job_id", ptrLabelsController.GetEstadoGeneracion)
		generacionGroup.GET("/descargar/:job_id", ptrLabelsController.DescargarEtiquetas)
		generacionGroup.POST("/cancelar/:job_id", ptrLabelsController.CancelarGeneracion)
	}

	// Rutas de preview
	previewGroup := protectedGroup.Group("/preview")
	{
		previewGroup.POST("", ptrLabelsController.GenerarPreview)
		previewGroup.POST("/batch", ptrLabelsController.GenerarPreviewLote)
	}

	// Rutas de historial
	historialGroup := protectedGroup.Group("/historial")
	{
		historialGroup.GET("", ptrLabelsController.GetHistorialGeneracion)
		historialGroup.GET("/:id", ptrLabelsController.GetDetalleHistorial)
		historialGroup.DELETE("/:id", ptrLabelsController.EliminarHistorial)
	}

	// Rutas de configuración (solo admin)
	configGroup := protectedGroup.Group("/configuracion")
	configGroup.Use(ptrAuthMiddleware.RequireRole("admin"))
	{
		configGroup.GET("", ptrLabelsController.GetConfiguracion)
		configGroup.PUT("", ptrLabelsController.ActualizarConfiguracion)
		configGroup.POST("/test-impresora", ptrLabelsController.TestImpresora)
	}

	// Rutas de métricas
	metricsGroup := protectedGroup.Group("/metricas")
	{
		metricsGroup.GET("", ptrLabelsController.GetMetricas)
		metricsGroup.GET("/uso", ptrLabelsController.GetMetricasUso)
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
func (ptrServer *structLabelsServer) setupGracefulShutdown() {
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
