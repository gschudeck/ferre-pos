package unit

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"ferre_pos_apis/internal/config"
	"ferre_pos_apis/internal/handlers"
	"ferre_pos_apis/internal/logger"
	"ferre_pos_apis/internal/models"
	"ferre_pos_apis/test/mocks"
	testutils "ferre_pos_apis/test/utils"
)

func setupPOSTestServer(t *testing.T) *testutils.TestServer {
	ts := testutils.SetupTestServer(t)
	
	// Configurar rutas específicas del API POS
	setupPOSRoutes(ts)
	
	return ts
}

func setupPOSRoutes(ts *testutils.TestServer) {
	// Mock dependencies
	mockDB := mocks.NewMockDatabase()
	mockValidator := &mocks.MockValidator{}
	mockMetrics := &mocks.MockMetrics{}
	
	// Handlers
	authHandler := handlers.NewAuthHandler(mockDB, logger.Get(), mockValidator, ts.Config)
	productosHandler := handlers.NewProductosHandler(mockDB, logger.Get(), mockValidator, mockMetrics)
	ventasHandler := handlers.NewVentasHandler(mockDB, logger.Get(), mockValidator, mockMetrics)
	usuariosHandler := handlers.NewUsuariosHandler(mockDB, logger.Get(), mockValidator, mockMetrics)
	
	// Rutas de autenticación
	auth := ts.Router.Group("/api/v1/auth")
	{
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.RefreshToken)
	}
	
	// Rutas de productos
	productos := ts.Router.Group("/api/v1/productos")
	{
		productos.GET("", productosHandler.List)
		productos.GET("/:id", productosHandler.GetByID)
		productos.POST("", productosHandler.Create)
		productos.PUT("/:id", productosHandler.Update)
		productos.DELETE("/:id", productosHandler.Delete)
		productos.GET("/search", productosHandler.Search)
		productos.GET("/barcode/:codigo", productosHandler.GetByBarcode)
	}
	
	// Rutas de ventas
	ventas := ts.Router.Group("/api/v1/ventas")
	{
		ventas.GET("", ventasHandler.List)
		ventas.GET("/:id", ventasHandler.GetByID)
		ventas.POST("", ventasHandler.Create)
		ventas.PUT("/:id", ventasHandler.Update)
		ventas.POST("/:id/anular", ventasHandler.Anular)
	}
	
	// Rutas de usuarios
	usuarios := ts.Router.Group("/api/v1/usuarios")
	{
		usuarios.GET("", usuariosHandler.List)
		usuarios.GET("/:id", usuariosHandler.GetByID)
		usuarios.POST("", usuariosHandler.Create)
		usuarios.PUT("/:id", usuariosHandler.Update)
	}
}

func TestPOSAuthLogin(t *testing.T) {
	ts := setupPOSTestServer(t)
	defer ts.TeardownTestDatabase(t)
	
	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful login",
			requestBody: map[string]interface{}{
				"username": "admin_test",
				"password": "password123",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid credentials",
			requestBody: map[string]interface{}{
				"username": "invalid",
				"password": "wrong",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "INVALID_CREDENTIALS",
		},
		{
			name: "missing username",
			requestBody: map[string]interface{}{
				"password": "password123",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
		{
			name: "missing password",
			requestBody: map[string]interface{}{
				"username": "admin_test",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
		{
			name:           "empty request body",
			requestBody:    map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := ts.MakeRequest("POST", "/api/v1/auth/login", tt.requestBody, nil)
			
			assert.Equal(t, tt.expectedStatus, recorder.Code)
			
			var response models.APIResponse
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			require.NoError(t, err)
			
			if tt.expectedStatus == http.StatusOK {
				assert.True(t, response.Success)
				assert.NotNil(t, response.Data)
			} else {
				assert.False(t, response.Success)
				if tt.expectedError != "" {
					errorData, ok := response.Error.(map[string]interface{})
					require.True(t, ok)
					assert.Equal(t, tt.expectedError, errorData["code"])
				}
			}
		})
	}
}

func TestPOSProductosList(t *testing.T) {
	ts := setupPOSTestServer(t)
	defer ts.TeardownTestDatabase(t)
	
	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
	}{
		{
			name:           "list all products",
			queryParams:    "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "list with pagination",
			queryParams:    "?limit=10&offset=0",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "list with invalid limit",
			queryParams:    "?limit=-1",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "list with invalid offset",
			queryParams:    "?offset=-1",
			expectedStatus: http.StatusBadRequest,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := ts.MakeRequest("GET", "/api/v1/productos"+tt.queryParams, nil, nil)
			
			assert.Equal(t, tt.expectedStatus, recorder.Code)
			
			var response models.APIResponse
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			require.NoError(t, err)
			
			if tt.expectedStatus == http.StatusOK {
				assert.True(t, response.Success)
				assert.NotNil(t, response.Data)
			} else {
				assert.False(t, response.Success)
			}
		})
	}
}

func TestPOSProductosGetByID(t *testing.T) {
	ts := setupPOSTestServer(t)
	defer ts.TeardownTestDatabase(t)
	
	tests := []struct {
		name           string
		productID      string
		expectedStatus int
	}{
		{
			name:           "get existing product",
			productID:      "test-product-1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "get non-existing product",
			productID:      "non-existing",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "get with invalid ID format",
			productID:      "",
			expectedStatus: http.StatusBadRequest,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/v1/productos/" + tt.productID
			recorder := ts.MakeRequest("GET", url, nil, nil)
			
			assert.Equal(t, tt.expectedStatus, recorder.Code)
			
			var response models.APIResponse
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			require.NoError(t, err)
			
			if tt.expectedStatus == http.StatusOK {
				assert.True(t, response.Success)
				assert.NotNil(t, response.Data)
			} else {
				assert.False(t, response.Success)
			}
		})
	}
}

func TestPOSProductosCreate(t *testing.T) {
	ts := setupPOSTestServer(t)
	defer ts.TeardownTestDatabase(t)
	
	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "create valid product",
			requestBody: map[string]interface{}{
				"codigo":        "TEST001",
				"codigo_barras": "1234567890123",
				"nombre":        "Producto Test",
				"descripcion":   "Descripción del producto test",
				"categoria_id":  "test-category-1",
				"precio":        100.00,
				"costo":         60.00,
				"stock":         50,
				"stock_minimo":  10,
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "create product with missing required fields",
			requestBody: map[string]interface{}{
				"nombre": "Producto Test",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
		{
			name: "create product with invalid price",
			requestBody: map[string]interface{}{
				"codigo":        "TEST002",
				"nombre":        "Producto Test",
				"precio":        -10.00,
				"stock":         50,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
		{
			name: "create product with duplicate code",
			requestBody: map[string]interface{}{
				"codigo":        "EXISTING_CODE",
				"nombre":        "Producto Test",
				"precio":        100.00,
				"stock":         50,
			},
			expectedStatus: http.StatusConflict,
			expectedError:  "DUPLICATE_CODE",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := ts.MakeRequest("POST", "/api/v1/productos", tt.requestBody, nil)
			
			assert.Equal(t, tt.expectedStatus, recorder.Code)
			
			var response models.APIResponse
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			require.NoError(t, err)
			
			if tt.expectedStatus == http.StatusCreated {
				assert.True(t, response.Success)
				assert.NotNil(t, response.Data)
			} else {
				assert.False(t, response.Success)
				if tt.expectedError != "" {
					errorData, ok := response.Error.(map[string]interface{})
					require.True(t, ok)
					assert.Equal(t, tt.expectedError, errorData["code"])
				}
			}
		})
	}
}

func TestPOSVentasCreate(t *testing.T) {
	ts := setupPOSTestServer(t)
	defer ts.TeardownTestDatabase(t)
	
	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "create valid sale",
			requestBody: map[string]interface{}{
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
					{
						"producto_id":     "test-product-2",
						"cantidad":        1,
						"precio_unitario": 200.00,
					},
				},
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "create sale with missing details",
			requestBody: map[string]interface{}{
				"tipo_documento": "boleta",
				"usuario_id":     "test-user-1",
				"metodo_pago":    "efectivo",
				"detalles":       []map[string]interface{}{},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
		{
			name: "create sale with invalid product",
			requestBody: map[string]interface{}{
				"tipo_documento": "boleta",
				"usuario_id":     "test-user-1",
				"metodo_pago":    "efectivo",
				"detalles": []map[string]interface{}{
					{
						"producto_id":     "non-existing-product",
						"cantidad":        1,
						"precio_unitario": 100.00,
					},
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "PRODUCT_NOT_FOUND",
		},
		{
			name: "create sale with insufficient stock",
			requestBody: map[string]interface{}{
				"tipo_documento": "boleta",
				"usuario_id":     "test-user-1",
				"metodo_pago":    "efectivo",
				"detalles": []map[string]interface{}{
					{
						"producto_id":     "test-product-1",
						"cantidad":        1000, // Más que el stock disponible
						"precio_unitario": 100.00,
					},
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INSUFFICIENT_STOCK",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := ts.MakeRequest("POST", "/api/v1/ventas", tt.requestBody, nil)
			
			assert.Equal(t, tt.expectedStatus, recorder.Code)
			
			var response models.APIResponse
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			require.NoError(t, err)
			
			if tt.expectedStatus == http.StatusCreated {
				assert.True(t, response.Success)
				assert.NotNil(t, response.Data)
			} else {
				assert.False(t, response.Success)
				if tt.expectedError != "" {
					errorData, ok := response.Error.(map[string]interface{})
					require.True(t, ok)
					assert.Equal(t, tt.expectedError, errorData["code"])
				}
			}
		})
	}
}

func TestPOSProductosSearch(t *testing.T) {
	ts := setupPOSTestServer(t)
	defer ts.TeardownTestDatabase(t)
	
	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
	}{
		{
			name:           "search by name",
			queryParams:    "?q=Producto",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "search by code",
			queryParams:    "?q=TEST001",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "search with pagination",
			queryParams:    "?q=Test&limit=5&offset=0",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "search with empty query",
			queryParams:    "?q=",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "search without query parameter",
			queryParams:    "",
			expectedStatus: http.StatusBadRequest,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := ts.MakeRequest("GET", "/api/v1/productos/search"+tt.queryParams, nil, nil)
			
			assert.Equal(t, tt.expectedStatus, recorder.Code)
			
			var response models.APIResponse
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			require.NoError(t, err)
			
			if tt.expectedStatus == http.StatusOK {
				assert.True(t, response.Success)
				assert.NotNil(t, response.Data)
			} else {
				assert.False(t, response.Success)
			}
		})
	}
}

func TestPOSProductosGetByBarcode(t *testing.T) {
	ts := setupPOSTestServer(t)
	defer ts.TeardownTestDatabase(t)
	
	tests := []struct {
		name           string
		barcode        string
		expectedStatus int
	}{
		{
			name:           "get by valid barcode",
			barcode:        "1234567890123",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "get by non-existing barcode",
			barcode:        "9999999999999",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "get by invalid barcode format",
			barcode:        "invalid-barcode",
			expectedStatus: http.StatusBadRequest,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/v1/productos/barcode/" + tt.barcode
			recorder := ts.MakeRequest("GET", url, nil, nil)
			
			assert.Equal(t, tt.expectedStatus, recorder.Code)
			
			var response models.APIResponse
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			require.NoError(t, err)
			
			if tt.expectedStatus == http.StatusOK {
				assert.True(t, response.Success)
				assert.NotNil(t, response.Data)
			} else {
				assert.False(t, response.Success)
			}
		})
	}
}

func TestPOSVentasAnular(t *testing.T) {
	ts := setupPOSTestServer(t)
	defer ts.TeardownTestDatabase(t)
	
	tests := []struct {
		name           string
		ventaID        string
		requestBody    interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name:    "anular valid sale",
			ventaID: "test-venta-1",
			requestBody: map[string]interface{}{
				"motivo": "Error en la venta",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "anular non-existing sale",
			ventaID:        "non-existing",
			requestBody:    map[string]interface{}{"motivo": "Test"},
			expectedStatus: http.StatusNotFound,
			expectedError:  "SALE_NOT_FOUND",
		},
		{
			name:           "anular without reason",
			ventaID:        "test-venta-1",
			requestBody:    map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
		{
			name:    "anular already cancelled sale",
			ventaID: "test-venta-cancelled",
			requestBody: map[string]interface{}{
				"motivo": "Test",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "SALE_ALREADY_CANCELLED",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/v1/ventas/" + tt.ventaID + "/anular"
			recorder := ts.MakeRequest("POST", url, tt.requestBody, nil)
			
			assert.Equal(t, tt.expectedStatus, recorder.Code)
			
			var response models.APIResponse
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			require.NoError(t, err)
			
			if tt.expectedStatus == http.StatusOK {
				assert.True(t, response.Success)
				assert.NotNil(t, response.Data)
			} else {
				assert.False(t, response.Success)
				if tt.expectedError != "" {
					errorData, ok := response.Error.(map[string]interface{})
					require.True(t, ok)
					assert.Equal(t, tt.expectedError, errorData["code"])
				}
			}
		})
	}
}

func TestPOSUsuariosList(t *testing.T) {
	ts := setupPOSTestServer(t)
	defer ts.TeardownTestDatabase(t)
	
	// Simular token de admin
	adminToken := "admin-token-123"
	
	tests := []struct {
		name           string
		token          string
		queryParams    string
		expectedStatus int
	}{
		{
			name:           "list users as admin",
			token:          adminToken,
			queryParams:    "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "list users with pagination",
			token:          adminToken,
			queryParams:    "?limit=10&offset=0",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "list users without token",
			token:          "",
			queryParams:    "",
			expectedStatus: http.StatusUnauthorized,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var recorder *httptest.ResponseRecorder
			if tt.token != "" {
				recorder = ts.MakeAuthenticatedRequest("GET", "/api/v1/usuarios"+tt.queryParams, nil, tt.token)
			} else {
				recorder = ts.MakeRequest("GET", "/api/v1/usuarios"+tt.queryParams, nil, nil)
			}
			
			assert.Equal(t, tt.expectedStatus, recorder.Code)
			
			var response models.APIResponse
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			require.NoError(t, err)
			
			if tt.expectedStatus == http.StatusOK {
				assert.True(t, response.Success)
				assert.NotNil(t, response.Data)
			} else {
				assert.False(t, response.Success)
			}
		})
	}
}

func BenchmarkPOSProductosList(b *testing.B) {
	gin.SetMode(gin.TestMode)
	ts := &testutils.TestServer{
		Router: gin.New(),
	}
	setupPOSRoutes(ts)
	
	req, _ := http.NewRequest("GET", "/api/v1/productos", nil)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		recorder := httptest.NewRecorder()
		ts.Router.ServeHTTP(recorder, req)
	}
}

func BenchmarkPOSVentasCreate(b *testing.B) {
	gin.SetMode(gin.TestMode)
	ts := &testutils.TestServer{
		Router: gin.New(),
	}
	setupPOSRoutes(ts)
	
	requestBody := map[string]interface{}{
		"tipo_documento": "boleta",
		"usuario_id":     "test-user-1",
		"metodo_pago":    "efectivo",
		"detalles": []map[string]interface{}{
			{
				"producto_id":     "test-product-1",
				"cantidad":        1,
				"precio_unitario": 100.00,
			},
		},
	}
	
	bodyBytes, _ := json.Marshal(requestBody)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("POST", "/api/v1/ventas", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		
		recorder := httptest.NewRecorder()
		ts.Router.ServeHTTP(recorder, req)
	}
}

