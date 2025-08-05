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
	apiName    = "pos"
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

	log.Info("Iniciando API POS - Sistema FERRE-POS")

	// Verificar si la API está habilitada
	apiConfig, err := config.GetAPIConfig(apiName)
	if err != nil {
		log.WithError(err).Fatal("Error obteniendo configuración de API")
	}

	if !apiConfig.Enabled {
		log.Info("API POS deshabilitada por configuración")
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
		log.WithField("address", server.Addr).Info("Servidor API POS iniciado")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.WithError(err).Fatal("Error iniciando servidor")
		}
	}()

	// Configurar graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Cerrando servidor API POS...")

	// Timeout para shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.WithError(err).Error("Error durante shutdown del servidor")
	}

	log.Info("Servidor API POS cerrado exitosamente")
}

// setupRoutes configura todas las rutas de la API
func setupRoutes(
	router *gin.Engine,
	db *database.Database,
	log logger.Logger,
	validator validator.Validator,
	metrics *metrics.Metrics,
	cfg *config.APIConfig,
) {
	// Inicializar handlers
	authHandler := handlers.NewAuthHandler(db, log, validator, cfg)
	productosHandler := handlers.NewProductosHandler(db, log, validator, metrics)
	ventasHandler := handlers.NewVentasHandler(db, log, validator, metrics)
	stockHandler := handlers.NewStockHandler(db, log, validator, metrics)
	usuariosHandler := handlers.NewUsuariosHandler(db, log, validator)

	// Rutas de salud y métricas
	router.GET("/health", handlers.HealthCheck(db, log))
	router.GET("/ready", handlers.ReadinessCheck(db, log))
	
	if metrics != nil {
		router.GET("/metrics", gin.WrapH(metrics.Handler()))
	}

	// Grupo de rutas API v1
	v1 := router.Group(fmt.Sprintf("/api/%s", apiVersion))
	{
		// Rutas de autenticación (sin middleware de auth)
		auth := v1.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.POST("/logout", middleware.Auth(cfg), authHandler.Logout)
		}

		// Rutas protegidas
		protected := v1.Group("")
		protected.Use(middleware.Auth(cfg))
		{
			// Rutas de productos
			productos := protected.Group("/productos")
			{
				productos.GET("", productosHandler.List)
				productos.GET("/:id", productosHandler.GetByID)
				productos.GET("/buscar", productosHandler.Search)
				productos.GET("/codigo-barra/:codigo", productosHandler.GetByBarcode)
				productos.POST("", middleware.RequireRole("admin", "supervisor"), productosHandler.Create)
				productos.PUT("/:id", middleware.RequireRole("admin", "supervisor"), productosHandler.Update)
				productos.DELETE("/:id", middleware.RequireRole("admin"), productosHandler.Delete)
			}

			// Rutas de stock
			stock := protected.Group("/stock")
			{
				stock.GET("/producto/:producto_id", stockHandler.GetByProducto)
				stock.GET("/sucursal/:sucursal_id", stockHandler.GetBySucursal)
				stock.POST("/reservar", stockHandler.ReservarStock)
				stock.POST("/liberar", stockHandler.LiberarStock)
				stock.GET("/alertas", stockHandler.GetAlertas)
			}

			// Rutas de ventas
			ventas := protected.Group("/ventas")
			{
				ventas.POST("", ventasHandler.Create)
				ventas.GET("/:id", ventasHandler.GetByID)
				ventas.PUT("/:id/anular", middleware.RequireRole("supervisor", "admin"), ventasHandler.Anular)
				ventas.GET("", ventasHandler.List)
				ventas.GET("/numero/:numero", ventasHandler.GetByNumero)
				ventas.POST("/:id/dte", ventasHandler.GenerarDTE)
			}

			// Rutas de usuarios
			usuarios := protected.Group("/usuarios")
			{
				usuarios.GET("/perfil", usuariosHandler.GetPerfil)
				usuarios.PUT("/perfil", usuariosHandler.UpdatePerfil)
				usuarios.POST("/cambiar-password", usuariosHandler.CambiarPassword)
				usuarios.GET("", middleware.RequireRole("supervisor", "admin"), usuariosHandler.List)
				usuarios.POST("", middleware.RequireRole("admin"), usuariosHandler.Create)
				usuarios.PUT("/:id", middleware.RequireRole("admin"), usuariosHandler.Update)
			}

		// Rutas de terminales
		terminales := protected.Group("/terminales")
		{
			terminales.GET("", ventasHandler.ListTerminales)
			terminales.GET("/:id", ventasHandler.GetTerminal)
			terminales.POST("", middleware.RequireRole("admin", "supervisor"), ventasHandler.CreateTerminal)
			terminales.PUT("/:id", middleware.RequireRole("admin", "supervisor"), ventasHandler.UpdateTerminal)
		}

		// Rutas de sucursales
		sucursales := protected.Group("/sucursales")
		{
			sucursales.GET("", ventasHandler.ListSucursales)
			sucursales.GET("/:id", ventasHandler.GetSucursal)
			sucursales.PUT("/:id", middleware.RequireRole("admin", "supervisor"), ventasHandler.UpdateSucursal)
		}

		// Rutas de reportes básicos
		reportes := protected.Group("/reportes")
		{
			reportes.GET("/ventas-diarias", ventasHandler.GetVentasDiarias)
			reportes.GET("/productos-vendidos", ventasHandler.GetProductosVendidos)
			reportes.GET("/stock-bajo", productosHandler.GetStockBajo)
		}
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

