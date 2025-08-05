//go:build integration
// +build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"ferre_pos_apis/internal/config"
	"ferre_pos_apis/internal/database"
	"ferre_pos_apis/internal/handlers"
	"ferre_pos_apis/internal/logger"
	"ferre_pos_apis/internal/metrics"
	"ferre_pos_apis/internal/middleware"
	"ferre_pos_apis/internal/models"
	"ferre_pos_apis/pkg/ratelimiter"
	"ferre_pos_apis/pkg/validator"
	testutils "ferre_pos_apis/test/utils"
)

// IntegrationTestSuite suite de tests de integración
type IntegrationTestSuite struct {
	suite.Suite
	db       *database.Database
	config   *config.Config
	servers  map[string]*httptest.Server
	routers  map[string]*gin.Engine
	tokens   map[string]string
}

// SetupSuite configuración inicial de la suite
func (suite *IntegrationTestSuite) SetupSuite() {
	// Configurar modo de test
	gin.SetMode(gin.TestMode)
	
	// Cargar configuración de test
	cfg, err := config.Load("../../configs/config.yaml")
	require.NoError(suite.T(), err)
	suite.config = cfg
	
	// Configurar logger
	err = logger.Init(&cfg.Logging, "integration_test")
	require.NoError(suite.T(), err)
	
	// Configurar base de datos de test
	suite.setupTestDatabase()
	
	// Configurar servidores de test
	suite.setupTestServers()
	
	// Configurar datos de test
	suite.setupTestData()
	
	// Configurar tokens de autenticación
	suite.setupAuthTokens()
}

// TearDownSuite limpieza final de la suite
func (suite *IntegrationTestSuite) TearDownSuite() {
	// Cerrar servidores
	for _, server := range suite.servers {
		server.Close()
	}
	
	// Limpiar base de datos
	suite.cleanupTestData()
	
	// Cerrar conexión de base de datos
	if suite.db != nil {
		suite.db.Close()
	}
}

// setupTestDatabase configura la base de datos de test
func (suite *IntegrationTestSuite) setupTestDatabase() {
	// Usar base de datos de test desde variables de entorno o configuración por defecto
	testDBConfig := &config.DatabaseConfig{
		Host:            getEnvOrDefault("TEST_DB_HOST", "localhost"),
		Port:            5432,
		User:            getEnvOrDefault("TEST_DB_USER", "test"),
		Password:        getEnvOrDefault("TEST_DB_PASSWORD", "test"),
		Name:            getEnvOrDefault("TEST_DB_NAME", "ferre_pos_test"),
		SSLMode:         "disable",
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
	}
	
	db, err := database.Init(testDBConfig, logger.Get())
	if err != nil {
		suite.T().Skipf("Skipping integration tests: database not available: %v", err)
		return
	}
	
	suite.db = db
}

// setupTestServers configura los servidores de test para cada API
func (suite *IntegrationTestSuite) setupTestServers() {
	suite.servers = make(map[string]*httptest.Server)
	suite.routers = make(map[string]*gin.Engine)
	
	// Configurar métricas y dependencias comunes
	metricsInstance := metrics.NewMetrics("test")
	validatorInstance := validator.New()
	rateLimiterConfig := &config.RateLimitConfig{
		Enabled:           false, // Deshabilitado para tests
		RequestsPerSecond: 1000.0,
		BurstSize:         2000,
		CleanupInterval:   time.Minute,
	}
	rateLimiterInstance := ratelimiter.NewIPRateLimiter(rateLimiterConfig, logger.Get())
	
	// API POS
	suite.setupPOSServer(metricsInstance, validatorInstance, rateLimiterInstance)
	
	// API Sync
	suite.setupSyncServer(metricsInstance, validatorInstance, rateLimiterInstance)
	
	// API Labels
	suite.setupLabelsServer(metricsInstance, validatorInstance, rateLimiterInstance)
	
	// API Reports
	suite.setupReportsServer(metricsInstance, validatorInstance, rateLimiterInstance)
}

// setupPOSServer configura el servidor del API POS
func (suite *IntegrationTestSuite) setupPOSServer(metrics metrics.Metrics, validator *validator.Validator, rateLimiter *ratelimiter.IPRateLimiter) {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.RequestID())
	router.Use(middleware.Logger(logger.Get()))
	router.Use(middleware.CORS(suite.config.Security.CORS))
	router.Use(middleware.Metrics(metrics, "pos"))
	
	// Handlers
	authHandler := handlers.NewAuthHandler(suite.db, logger.Get(), validator, suite.config)
	productosHandler := handlers.NewProductosHandler(suite.db, logger.Get(), validator, metrics)
	ventasHandler := handlers.NewVentasHandler(suite.db, logger.Get(), validator, metrics)
	usuariosHandler := handlers.NewUsuariosHandler(suite.db, logger.Get(), validator, metrics)
	
	// Rutas
	api := router.Group("/api/v1")
	{
		// Autenticación
		auth := api.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
		}
		
		// Productos
		productos := api.Group("/productos")
		productos.Use(middleware.JWTAuth(suite.config.Security.JWTSecret))
		{
			productos.GET("", productosHandler.List)
			productos.GET("/:id", productosHandler.GetByID)
			productos.POST("", productosHandler.Create)
			productos.PUT("/:id", productosHandler.Update)
			productos.DELETE("/:id", productosHandler.Delete)
			productos.GET("/search", productosHandler.Search)
			productos.GET("/barcode/:codigo", productosHandler.GetByBarcode)
		}
		
		// Ventas
		ventas := api.Group("/ventas")
		ventas.Use(middleware.JWTAuth(suite.config.Security.JWTSecret))
		{
			ventas.GET("", ventasHandler.List)
			ventas.GET("/:id", ventasHandler.GetByID)
			ventas.POST("", ventasHandler.Create)
			ventas.PUT("/:id", ventasHandler.Update)
			ventas.POST("/:id/anular", ventasHandler.Anular)
		}
		
		// Usuarios
		usuarios := api.Group("/usuarios")
		usuarios.Use(middleware.JWTAuth(suite.config.Security.JWTSecret))
		{
			usuarios.GET("", usuariosHandler.List)
			usuarios.GET("/:id", usuariosHandler.GetByID)
			usuarios.POST("", usuariosHandler.Create)
			usuarios.PUT("/:id", usuariosHandler.Update)
		}
	}
	
	suite.routers["pos"] = router
	suite.servers["pos"] = httptest.NewServer(router)
}

// setupSyncServer configura el servidor del API Sync
func (suite *IntegrationTestSuite) setupSyncServer(metrics metrics.Metrics, validator *validator.Validator, rateLimiter *ratelimiter.IPRateLimiter) {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.RequestID())
	router.Use(middleware.Logger(logger.Get()))
	router.Use(middleware.CORS(suite.config.Security.CORS))
	router.Use(middleware.Metrics(metrics, "sync"))
	
	// Handlers
	syncHandler := handlers.NewSyncHandler(suite.db, logger.Get(), validator, metrics)
	
	// Rutas
	api := router.Group("/api/v1/sync")
	{
		api.POST("/auth", syncHandler.AuthenticateTerminal)
		
		// Rutas protegidas con autenticación de terminal
		protected := api.Group("")
		protected.Use(middleware.TerminalAuth(suite.config.Security.JWTSecret))
		{
			protected.POST("/productos", syncHandler.SyncProductos)
			protected.POST("/ventas", syncHandler.SyncVentas)
			protected.GET("/cambios", syncHandler.GetCambios)
			protected.GET("/status", syncHandler.GetSyncStatus)
			protected.POST("/heartbeat", syncHandler.Heartbeat)
		}
	}
	
	suite.routers["sync"] = router
	suite.servers["sync"] = httptest.NewServer(router)
}

// setupLabelsServer configura el servidor del API Labels
func (suite *IntegrationTestSuite) setupLabelsServer(metrics metrics.Metrics, validator *validator.Validator, rateLimiter *ratelimiter.IPRateLimiter) {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.RequestID())
	router.Use(middleware.Logger(logger.Get()))
	router.Use(middleware.CORS(suite.config.Security.CORS))
	router.Use(middleware.Metrics(metrics, "labels"))
	
	// Handlers
	labelsHandler := handlers.NewLabelsHandler(suite.db, logger.Get(), validator, metrics)
	
	// Rutas
	api := router.Group("/api/v1/labels")
	{
		api.POST("/generate", labelsHandler.GenerateLabels)
		api.POST("/generate/batch", labelsHandler.GenerateBatchLabels)
		api.GET("/templates", labelsHandler.GetTemplates)
		api.POST("/templates", labelsHandler.CreateTemplate)
		api.POST("/barcode/generate", labelsHandler.GenerateBarcode)
		api.POST("/barcode/validate", labelsHandler.ValidateBarcode)
		api.GET("/barcode/formats", labelsHandler.GetBarcodeFormats)
		api.POST("/preview", labelsHandler.PreviewLabel)
		api.GET("/history", labelsHandler.GetHistory)
	}
	
	suite.routers["labels"] = router
	suite.servers["labels"] = httptest.NewServer(router)
}

// setupReportsServer configura el servidor del API Reports
func (suite *IntegrationTestSuite) setupReportsServer(metrics metrics.Metrics, validator *validator.Validator, rateLimiter *ratelimiter.IPRateLimiter) {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.RequestID())
	router.Use(middleware.Logger(logger.Get()))
	router.Use(middleware.CORS(suite.config.Security.CORS))
	router.Use(middleware.Metrics(metrics, "reports"))
	
	// Handlers
	reportsHandler := handlers.NewReportsHandler(suite.db, logger.Get(), validator, metrics)
	
	// Rutas
	api := router.Group("/api/v1/reports")
	{
		api.GET("/ventas", reportsHandler.ReporteVentas)
		api.GET("/ventas/diarias", reportsHandler.VentasDiarias)
		api.GET("/inventario", reportsHandler.ReporteInventario)
		api.GET("/inventario/stock-bajo", reportsHandler.StockBajo)
		api.GET("/dashboard/resumen", reportsHandler.DashboardResumen)
		api.GET("/analytics/productos-top", reportsHandler.ProductosTop)
		api.POST("/export", reportsHandler.ExportarReporte)
		api.GET("/export/:id/status", reportsHandler.EstadoExportacion)
		api.POST("/custom", reportsHandler.CrearReportePersonalizado)
		api.POST("/custom/:id/execute", reportsHandler.EjecutarReportePersonalizado)
	}
	
	suite.routers["reports"] = router
	suite.servers["reports"] = httptest.NewServer(router)
}

// setupTestData configura datos de prueba en la base de datos
func (suite *IntegrationTestSuite) setupTestData() {
	if suite.db == nil {
		return
	}
	
	ctx := context.Background()
	
	// Crear datos de test usando las fixtures
	var testData map[string]interface{}
	testutils.ParseJSONFixture(suite.T(), "test_data.json", &testData)
	
	// Insertar usuarios de test
	if usuarios, ok := testData["usuarios"].([]interface{}); ok {
		for _, usuario := range usuarios {
			if userMap, ok := usuario.(map[string]interface{}); ok {
				suite.insertTestUser(ctx, userMap)
			}
		}
	}
	
	// Insertar productos de test
	if productos, ok := testData["productos"].([]interface{}); ok {
		for _, producto := range productos {
			if prodMap, ok := producto.(map[string]interface{}); ok {
				suite.insertTestProduct(ctx, prodMap)
			}
		}
	}
	
	// Insertar categorías de test
	if categorias, ok := testData["categorias"].([]interface{}); ok {
		for _, categoria := range categorias {
			if catMap, ok := categoria.(map[string]interface{}); ok {
				suite.insertTestCategory(ctx, catMap)
			}
		}
	}
	
	// Insertar sucursales de test
	if sucursales, ok := testData["sucursales"].([]interface{}); ok {
		for _, sucursal := range sucursales {
			if sucMap, ok := sucursal.(map[string]interface{}); ok {
				suite.insertTestSucursal(ctx, sucMap)
			}
		}
	}
}

// setupAuthTokens configura tokens de autenticación para los tests
func (suite *IntegrationTestSuite) setupAuthTokens() {
	suite.tokens = make(map[string]string)
	
	// Token de admin
	adminToken, err := testutils.GenerateTestJWT("test-user-1", "admin", suite.config)
	if err == nil {
		suite.tokens["admin"] = adminToken
	}
	
	// Token de vendedor
	vendedorToken, err := testutils.GenerateTestJWT("test-user-2", "vendedor", suite.config)
	if err == nil {
		suite.tokens["vendedor"] = vendedorToken
	}
	
	// Token de terminal
	terminalToken, err := testutils.GenerateTestJWT("test-terminal-1", "terminal", suite.config)
	if err == nil {
		suite.tokens["terminal"] = terminalToken
	}
}

// insertTestUser inserta un usuario de test
func (suite *IntegrationTestSuite) insertTestUser(ctx context.Context, userData map[string]interface{}) {
	query := `
		INSERT INTO usuarios (id, username, email, password_hash, rol, activo, sucursal_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id) DO NOTHING
	`
	
	_, err := suite.db.ExecContext(ctx, query,
		userData["id"],
		userData["username"],
		userData["email"],
		userData["password_hash"],
		userData["rol"],
		userData["activo"],
		userData["sucursal_id"],
		userData["created_at"],
		userData["updated_at"],
	)
	
	if err != nil {
		suite.T().Logf("Warning: failed to insert test user: %v", err)
	}
}

// insertTestProduct inserta un producto de test
func (suite *IntegrationTestSuite) insertTestProduct(ctx context.Context, productData map[string]interface{}) {
	query := `
		INSERT INTO productos (id, codigo, codigo_barras, nombre, descripcion, categoria_id, precio, costo, stock, stock_minimo, activo, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (id) DO NOTHING
	`
	
	_, err := suite.db.ExecContext(ctx, query,
		productData["id"],
		productData["codigo"],
		productData["codigo_barras"],
		productData["nombre"],
		productData["descripcion"],
		productData["categoria_id"],
		productData["precio"],
		productData["costo"],
		productData["stock"],
		productData["stock_minimo"],
		productData["activo"],
		productData["created_at"],
		productData["updated_at"],
	)
	
	if err != nil {
		suite.T().Logf("Warning: failed to insert test product: %v", err)
	}
}

// insertTestCategory inserta una categoría de test
func (suite *IntegrationTestSuite) insertTestCategory(ctx context.Context, categoryData map[string]interface{}) {
	query := `
		INSERT INTO categorias (id, nombre, descripcion, activa, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO NOTHING
	`
	
	_, err := suite.db.ExecContext(ctx, query,
		categoryData["id"],
		categoryData["nombre"],
		categoryData["descripcion"],
		categoryData["activa"],
		categoryData["created_at"],
		categoryData["updated_at"],
	)
	
	if err != nil {
		suite.T().Logf("Warning: failed to insert test category: %v", err)
	}
}

// insertTestSucursal inserta una sucursal de test
func (suite *IntegrationTestSuite) insertTestSucursal(ctx context.Context, sucursalData map[string]interface{}) {
	query := `
		INSERT INTO sucursales (id, nombre, direccion, telefono, email, activa, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO NOTHING
	`
	
	_, err := suite.db.ExecContext(ctx, query,
		sucursalData["id"],
		sucursalData["nombre"],
		sucursalData["direccion"],
		sucursalData["telefono"],
		sucursalData["email"],
		sucursalData["activa"],
		sucursalData["created_at"],
		sucursalData["updated_at"],
	)
	
	if err != nil {
		suite.T().Logf("Warning: failed to insert test sucursal: %v", err)
	}
}

// cleanupTestData limpia los datos de test
func (suite *IntegrationTestSuite) cleanupTestData() {
	if suite.db == nil {
		return
	}
	
	ctx := context.Background()
	tables := []string{
		"detalle_ventas",
		"ventas",
		"productos",
		"categorias",
		"usuarios",
		"sucursales",
		"terminales",
	}
	
	testutils.CleanupTestData(suite.T(), suite.db, tables...)
}

// makeRequest realiza una petición HTTP a un API específico
func (suite *IntegrationTestSuite) makeRequest(apiName, method, path string, body interface{}, token string) *http.Response {
	server, exists := suite.servers[apiName]
	require.True(suite.T(), exists, "API server not found: %s", apiName)
	
	var bodyReader *bytes.Buffer
	if body != nil {
		bodyBytes, _ := json.Marshal(body)
		bodyReader = bytes.NewBuffer(bodyBytes)
	} else {
		bodyReader = bytes.NewBuffer(nil)
	}
	
	url := server.URL + path
	req, err := http.NewRequest(method, url, bodyReader)
	require.NoError(suite.T(), err)
	
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	require.NoError(suite.T(), err)
	
	return resp
}

// TestCompleteWorkflow test de flujo completo de trabajo
func (suite *IntegrationTestSuite) TestCompleteWorkflow() {
	// 1. Autenticación en API POS
	loginResp := suite.makeRequest("pos", "POST", "/api/v1/auth/login", map[string]interface{}{
		"username": "admin_test",
		"password": "password123",
	}, "")
	
	assert.Equal(suite.T(), http.StatusOK, loginResp.StatusCode)
	
	var loginResult models.APIResponse
	err := json.NewDecoder(loginResp.Body).Decode(&loginResult)
	require.NoError(suite.T(), err)
	loginResp.Body.Close()
	
	assert.True(suite.T(), loginResult.Success)
	
	// Extraer token
	loginData, ok := loginResult.Data.(map[string]interface{})
	require.True(suite.T(), ok)
	token, ok := loginData["token"].(string)
	require.True(suite.T(), ok)
	require.NotEmpty(suite.T(), token)
	
	// 2. Crear producto en API POS
	productResp := suite.makeRequest("pos", "POST", "/api/v1/productos", map[string]interface{}{
		"codigo":        "INTEGRATION001",
		"codigo_barras": "1234567890999",
		"nombre":        "Producto Integración",
		"descripcion":   "Producto para test de integración",
		"categoria_id":  "test-category-1",
		"precio":        150.00,
		"costo":         90.00,
		"stock":         100,
		"stock_minimo":  10,
	}, token)
	
	assert.Equal(suite.T(), http.StatusCreated, productResp.StatusCode)
	productResp.Body.Close()
	
	// 3. Crear venta en API POS
	ventaResp := suite.makeRequest("pos", "POST", "/api/v1/ventas", map[string]interface{}{
		"tipo_documento": "boleta",
		"usuario_id":     "test-user-1",
		"sucursal_id":    "test-sucursal-1",
		"metodo_pago":    "efectivo",
		"detalles": []map[string]interface{}{
			{
				"producto_id":     "test-product-1",
				"cantidad":        2,
				"precio_unitario": 100.00,
			},
		},
	}, token)
	
	assert.Equal(suite.T(), http.StatusCreated, ventaResp.StatusCode)
	ventaResp.Body.Close()
	
	// 4. Generar etiqueta en API Labels
	labelResp := suite.makeRequest("labels", "POST", "/api/v1/labels/generate", map[string]interface{}{
		"template_id": "template-basic",
		"format":      "pdf",
		"productos": []map[string]interface{}{
			{
				"id":            "test-product-1",
				"codigo":        "TEST001",
				"nombre":        "Producto Test",
				"precio":        100.00,
				"codigo_barras": "1234567890123",
			},
		},
	}, "")
	
	assert.Equal(suite.T(), http.StatusOK, labelResp.StatusCode)
	labelResp.Body.Close()
	
	// 5. Generar reporte en API Reports
	reportResp := suite.makeRequest("reports", "GET", "/api/v1/reports/ventas?start_date=2024-01-01&end_date=2024-12-31", nil, "")
	
	assert.Equal(suite.T(), http.StatusOK, reportResp.StatusCode)
	reportResp.Body.Close()
	
	// 6. Sincronización en API Sync (autenticación de terminal)
	terminalAuthResp := suite.makeRequest("sync", "POST", "/api/v1/sync/auth", map[string]interface{}{
		"terminal_id": "TERM001",
		"api_key":     "valid-api-key",
		"sucursal_id": "test-sucursal-1",
	}, "")
	
	assert.Equal(suite.T(), http.StatusOK, terminalAuthResp.StatusCode)
	terminalAuthResp.Body.Close()
}

// TestCrossAPIDataConsistency test de consistencia de datos entre APIs
func (suite *IntegrationTestSuite) TestCrossAPIDataConsistency() {
	token := suite.tokens["admin"]
	if token == "" {
		suite.T().Skip("Admin token not available")
	}
	
	// 1. Crear producto en API POS
	productData := map[string]interface{}{
		"codigo":        "CONSISTENCY001",
		"codigo_barras": "1234567890888",
		"nombre":        "Producto Consistencia",
		"precio":        200.00,
		"stock":         50,
	}
	
	productResp := suite.makeRequest("pos", "POST", "/api/v1/productos", productData, token)
	assert.Equal(suite.T(), http.StatusCreated, productResp.StatusCode)
	productResp.Body.Close()
	
	// 2. Verificar que el producto aparece en reportes
	time.Sleep(1 * time.Second) // Esperar a que se procese
	
	reportResp := suite.makeRequest("reports", "GET", "/api/v1/reports/inventario", nil, "")
	assert.Equal(suite.T(), http.StatusOK, reportResp.StatusCode)
	
	var reportResult models.APIResponse
	err := json.NewDecoder(reportResp.Body).Decode(&reportResult)
	require.NoError(suite.T(), err)
	reportResp.Body.Close()
	
	assert.True(suite.T(), reportResult.Success)
	
	// 3. Verificar que se puede generar etiqueta del producto
	labelResp := suite.makeRequest("labels", "POST", "/api/v1/labels/generate", map[string]interface{}{
		"template_id": "template-basic",
		"format":      "pdf",
		"productos": []map[string]interface{}{
			{
				"codigo":        "CONSISTENCY001",
				"nombre":        "Producto Consistencia",
				"precio":        200.00,
				"codigo_barras": "1234567890888",
			},
		},
	}, "")
	
	assert.Equal(suite.T(), http.StatusOK, labelResp.StatusCode)
	labelResp.Body.Close()
}

// TestAPIPerformance test de rendimiento de las APIs
func (suite *IntegrationTestSuite) TestAPIPerformance() {
	token := suite.tokens["admin"]
	if token == "" {
		suite.T().Skip("Admin token not available")
	}
	
	// Test de rendimiento para listado de productos
	start := time.Now()
	productResp := suite.makeRequest("pos", "GET", "/api/v1/productos?limit=100", nil, token)
	duration := time.Since(start)
	
	assert.Equal(suite.T(), http.StatusOK, productResp.StatusCode)
	assert.Less(suite.T(), duration, 2*time.Second, "Product listing should complete within 2 seconds")
	productResp.Body.Close()
	
	// Test de rendimiento para dashboard
	start = time.Now()
	dashboardResp := suite.makeRequest("reports", "GET", "/api/v1/reports/dashboard/resumen", nil, "")
	duration = time.Since(start)
	
	assert.Equal(suite.T(), http.StatusOK, dashboardResp.StatusCode)
	assert.Less(suite.T(), duration, 3*time.Second, "Dashboard should load within 3 seconds")
	dashboardResp.Body.Close()
}

// TestConcurrentRequests test de requests concurrentes
func (suite *IntegrationTestSuite) TestConcurrentRequests() {
	token := suite.tokens["admin"]
	if token == "" {
		suite.T().Skip("Admin token not available")
	}
	
	const numRequests = 10
	results := make(chan int, numRequests)
	
	// Realizar múltiples requests concurrentes
	for i := 0; i < numRequests; i++ {
		go func(id int) {
			resp := suite.makeRequest("pos", "GET", fmt.Sprintf("/api/v1/productos?limit=10&offset=%d", id*10), nil, token)
			results <- resp.StatusCode
			resp.Body.Close()
		}(i)
	}
	
	// Verificar que todas las requests fueron exitosas
	successCount := 0
	for i := 0; i < numRequests; i++ {
		statusCode := <-results
		if statusCode == http.StatusOK {
			successCount++
		}
	}
	
	assert.Equal(suite.T(), numRequests, successCount, "All concurrent requests should succeed")
}

// TestErrorHandling test de manejo de errores
func (suite *IntegrationTestSuite) TestErrorHandling() {
	// Test de endpoint no existente
	resp := suite.makeRequest("pos", "GET", "/api/v1/nonexistent", nil, "")
	assert.Equal(suite.T(), http.StatusNotFound, resp.StatusCode)
	resp.Body.Close()
	
	// Test de request sin autenticación
	resp = suite.makeRequest("pos", "GET", "/api/v1/productos", nil, "")
	assert.Equal(suite.T(), http.StatusUnauthorized, resp.StatusCode)
	resp.Body.Close()
	
	// Test de request con token inválido
	resp = suite.makeRequest("pos", "GET", "/api/v1/productos", nil, "invalid-token")
	assert.Equal(suite.T(), http.StatusUnauthorized, resp.StatusCode)
	resp.Body.Close()
	
	// Test de request con datos inválidos
	resp = suite.makeRequest("pos", "POST", "/api/v1/productos", map[string]interface{}{
		"nombre": "", // Nombre vacío
		"precio": -10, // Precio negativo
	}, suite.tokens["admin"])
	assert.Equal(suite.T(), http.StatusBadRequest, resp.StatusCode)
	resp.Body.Close()
}

// Función auxiliar para obtener variables de entorno con valor por defecto
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// TestIntegrationSuite ejecuta la suite de tests de integración
func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

