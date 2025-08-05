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
	"github.com/stretchr/testify/require"

	"ferre_pos_apis/internal/config"
	"ferre_pos_apis/internal/handlers"
	"ferre_pos_apis/internal/logger"
	"ferre_pos_apis/internal/models"
	"ferre_pos_apis/test/mocks"
	testutils "ferre_pos_apis/test/utils"
)

func setupSyncTestServer(t *testing.T) *testutils.TestServer {
	ts := testutils.SetupTestServer(t)
	
	// Configurar rutas específicas del API Sync
	setupSyncRoutes(ts)
	
	return ts
}

func setupSyncRoutes(ts *testutils.TestServer) {
	// Mock dependencies
	mockDB := mocks.NewMockDatabase()
	mockValidator := &mocks.MockValidator{}
	mockMetrics := &mocks.MockMetrics{}
	
	// Handlers
	syncHandler := handlers.NewSyncHandler(mockDB, logger.Get(), mockValidator, mockMetrics)
	
	// Rutas de sincronización
	sync := ts.Router.Group("/api/v1/sync")
	{
		// Autenticación de terminal
		sync.POST("/auth", syncHandler.AuthenticateTerminal)
		
		// Sincronización de datos
		sync.POST("/productos", syncHandler.SyncProductos)
		sync.POST("/ventas", syncHandler.SyncVentas)
		sync.POST("/usuarios", syncHandler.SyncUsuarios)
		sync.POST("/configuracion", syncHandler.SyncConfiguracion)
		
		// Obtener cambios desde el servidor
		sync.GET("/cambios", syncHandler.GetCambios)
		sync.GET("/productos/cambios", syncHandler.GetProductosCambios)
		sync.GET("/usuarios/cambios", syncHandler.GetUsuariosCambios)
		
		// Estado de sincronización
		sync.GET("/status", syncHandler.GetSyncStatus)
		sync.POST("/heartbeat", syncHandler.Heartbeat)
		
		// Resolución de conflictos
		sync.POST("/conflictos/resolver", syncHandler.ResolverConflictos)
		sync.GET("/conflictos", syncHandler.GetConflictos)
	}
}

func TestSyncAuthenticateTerminal(t *testing.T) {
	ts := setupSyncTestServer(t)
	defer ts.TeardownTestDatabase(t)
	
	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful terminal authentication",
			requestBody: map[string]interface{}{
				"terminal_id": "TERM001",
				"api_key":     "valid-api-key",
				"sucursal_id": "test-sucursal-1",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid api key",
			requestBody: map[string]interface{}{
				"terminal_id": "TERM001",
				"api_key":     "invalid-key",
				"sucursal_id": "test-sucursal-1",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "INVALID_API_KEY",
		},
		{
			name: "missing terminal id",
			requestBody: map[string]interface{}{
				"api_key":     "valid-api-key",
				"sucursal_id": "test-sucursal-1",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
		{
			name: "inactive terminal",
			requestBody: map[string]interface{}{
				"terminal_id": "TERM_INACTIVE",
				"api_key":     "valid-api-key",
				"sucursal_id": "test-sucursal-1",
			},
			expectedStatus: http.StatusForbidden,
			expectedError:  "TERMINAL_INACTIVE",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := ts.MakeRequest("POST", "/api/v1/sync/auth", tt.requestBody, nil)
			
			assert.Equal(t, tt.expectedStatus, recorder.Code)
			
			var response models.APIResponse
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			require.NoError(t, err)
			
			if tt.expectedStatus == http.StatusOK {
				assert.True(t, response.Success)
				assert.NotNil(t, response.Data)
				
				// Verificar que se devuelve un token
				data, ok := response.Data.(map[string]interface{})
				require.True(t, ok)
				assert.NotEmpty(t, data["token"])
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

func TestSyncProductos(t *testing.T) {
	ts := setupSyncTestServer(t)
	defer ts.TeardownTestDatabase(t)
	
	terminalToken := "valid-terminal-token"
	
	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "sync productos successfully",
			requestBody: map[string]interface{}{
				"terminal_id": "TERM001",
				"timestamp":   time.Now().Unix(),
				"productos": []map[string]interface{}{
					{
						"id":            "local-product-1",
						"codigo":        "SYNC001",
						"nombre":        "Producto Sincronizado",
						"precio":        150.00,
						"stock":         25,
						"last_updated":  time.Now().Unix(),
						"action":        "create",
					},
					{
						"id":            "local-product-2",
						"codigo":        "SYNC002",
						"nombre":        "Producto Actualizado",
						"precio":        200.00,
						"stock":         30,
						"last_updated":  time.Now().Unix(),
						"action":        "update",
					},
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "sync with invalid product data",
			requestBody: map[string]interface{}{
				"terminal_id": "TERM001",
				"timestamp":   time.Now().Unix(),
				"productos": []map[string]interface{}{
					{
						"id":     "invalid-product",
						"codigo": "", // Código vacío
						"nombre": "Producto Inválido",
						"precio": -10.00, // Precio negativo
						"action": "create",
					},
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
		{
			name: "sync without authentication",
			requestBody: map[string]interface{}{
				"terminal_id": "TERM001",
				"productos":   []map[string]interface{}{},
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "AUTHENTICATION_REQUIRED",
		},
		{
			name: "sync with conflicting data",
			requestBody: map[string]interface{}{
				"terminal_id": "TERM001",
				"timestamp":   time.Now().Add(-time.Hour).Unix(), // Timestamp antiguo
				"productos": []map[string]interface{}{
					{
						"id":            "conflict-product",
						"codigo":        "CONFLICT001",
						"nombre":        "Producto Conflictivo",
						"precio":        100.00,
						"last_updated":  time.Now().Add(-time.Hour).Unix(),
						"action":        "update",
					},
				},
			},
			expectedStatus: http.StatusConflict,
			expectedError:  "SYNC_CONFLICT",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var recorder *httptest.ResponseRecorder
			if tt.expectedStatus == http.StatusUnauthorized {
				recorder = ts.MakeRequest("POST", "/api/v1/sync/productos", tt.requestBody, nil)
			} else {
				recorder = ts.MakeAuthenticatedRequest("POST", "/api/v1/sync/productos", tt.requestBody, terminalToken)
			}
			
			assert.Equal(t, tt.expectedStatus, recorder.Code)
			
			var response models.APIResponse
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			require.NoError(t, err)
			
			if tt.expectedStatus == http.StatusOK {
				assert.True(t, response.Success)
				assert.NotNil(t, response.Data)
				
				// Verificar estadísticas de sincronización
				data, ok := response.Data.(map[string]interface{})
				require.True(t, ok)
				assert.Contains(t, data, "processed")
				assert.Contains(t, data, "conflicts")
				assert.Contains(t, data, "errors")
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

func TestSyncVentas(t *testing.T) {
	ts := setupSyncTestServer(t)
	defer ts.TeardownTestDatabase(t)
	
	terminalToken := "valid-terminal-token"
	
	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "sync ventas successfully",
			requestBody: map[string]interface{}{
				"terminal_id": "TERM001",
				"timestamp":   time.Now().Unix(),
				"ventas": []map[string]interface{}{
					{
						"id":               "local-venta-1",
						"numero_documento": "V001",
						"tipo_documento":   "boleta",
						"total":            300.00,
						"fecha_venta":      time.Now().Unix(),
						"detalles": []map[string]interface{}{
							{
								"producto_id":     "test-product-1",
								"cantidad":        2,
								"precio_unitario": 150.00,
							},
						},
						"action": "create",
					},
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "sync venta with missing details",
			requestBody: map[string]interface{}{
				"terminal_id": "TERM001",
				"timestamp":   time.Now().Unix(),
				"ventas": []map[string]interface{}{
					{
						"id":               "local-venta-2",
						"numero_documento": "V002",
						"tipo_documento":   "boleta",
						"total":            100.00,
						"detalles":         []map[string]interface{}{}, // Sin detalles
						"action":           "create",
					},
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
		{
			name: "sync venta with invalid total",
			requestBody: map[string]interface{}{
				"terminal_id": "TERM001",
				"timestamp":   time.Now().Unix(),
				"ventas": []map[string]interface{}{
					{
						"id":               "local-venta-3",
						"numero_documento": "V003",
						"tipo_documento":   "boleta",
						"total":            -100.00, // Total negativo
						"detalles": []map[string]interface{}{
							{
								"producto_id":     "test-product-1",
								"cantidad":        1,
								"precio_unitario": 100.00,
							},
						},
						"action": "create",
					},
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := ts.MakeAuthenticatedRequest("POST", "/api/v1/sync/ventas", tt.requestBody, terminalToken)
			
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

func TestGetCambios(t *testing.T) {
	ts := setupSyncTestServer(t)
	defer ts.TeardownTestDatabase(t)
	
	terminalToken := "valid-terminal-token"
	
	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
	}{
		{
			name:           "get all changes",
			queryParams:    "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "get changes since timestamp",
			queryParams:    "?since=" + string(rune(time.Now().Add(-time.Hour).Unix())),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "get changes with limit",
			queryParams:    "?limit=10",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "get changes for specific table",
			queryParams:    "?table=productos",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "get changes with invalid timestamp",
			queryParams:    "?since=invalid",
			expectedStatus: http.StatusBadRequest,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := ts.MakeAuthenticatedRequest("GET", "/api/v1/sync/cambios"+tt.queryParams, nil, terminalToken)
			
			assert.Equal(t, tt.expectedStatus, recorder.Code)
			
			var response models.APIResponse
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			require.NoError(t, err)
			
			if tt.expectedStatus == http.StatusOK {
				assert.True(t, response.Success)
				assert.NotNil(t, response.Data)
				
				// Verificar estructura de cambios
				data, ok := response.Data.(map[string]interface{})
				require.True(t, ok)
				assert.Contains(t, data, "changes")
				assert.Contains(t, data, "timestamp")
				assert.Contains(t, data, "has_more")
			} else {
				assert.False(t, response.Success)
			}
		})
	}
}

func TestGetSyncStatus(t *testing.T) {
	ts := setupSyncTestServer(t)
	defer ts.TeardownTestDatabase(t)
	
	terminalToken := "valid-terminal-token"
	
	tests := []struct {
		name           string
		expectedStatus int
	}{
		{
			name:           "get sync status",
			expectedStatus: http.StatusOK,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := ts.MakeAuthenticatedRequest("GET", "/api/v1/sync/status", nil, terminalToken)
			
			assert.Equal(t, tt.expectedStatus, recorder.Code)
			
			var response models.APIResponse
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			require.NoError(t, err)
			
			if tt.expectedStatus == http.StatusOK {
				assert.True(t, response.Success)
				assert.NotNil(t, response.Data)
				
				// Verificar estructura del status
				data, ok := response.Data.(map[string]interface{})
				require.True(t, ok)
				assert.Contains(t, data, "terminal_id")
				assert.Contains(t, data, "last_sync")
				assert.Contains(t, data, "pending_changes")
				assert.Contains(t, data, "sync_status")
			}
		})
	}
}

func TestHeartbeat(t *testing.T) {
	ts := setupSyncTestServer(t)
	defer ts.TeardownTestDatabase(t)
	
	terminalToken := "valid-terminal-token"
	
	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
	}{
		{
			name: "successful heartbeat",
			requestBody: map[string]interface{}{
				"terminal_id": "TERM001",
				"timestamp":   time.Now().Unix(),
				"status":      "online",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "heartbeat with metrics",
			requestBody: map[string]interface{}{
				"terminal_id": "TERM001",
				"timestamp":   time.Now().Unix(),
				"status":      "online",
				"metrics": map[string]interface{}{
					"cpu_usage":    45.5,
					"memory_usage": 67.2,
					"disk_usage":   23.8,
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "heartbeat with missing terminal_id",
			requestBody: map[string]interface{}{
				"timestamp": time.Now().Unix(),
				"status":    "online",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := ts.MakeAuthenticatedRequest("POST", "/api/v1/sync/heartbeat", tt.requestBody, terminalToken)
			
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

func TestResolverConflictos(t *testing.T) {
	ts := setupSyncTestServer(t)
	defer ts.TeardownTestDatabase(t)
	
	terminalToken := "valid-terminal-token"
	
	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "resolve conflicts successfully",
			requestBody: map[string]interface{}{
				"terminal_id": "TERM001",
				"conflictos": []map[string]interface{}{
					{
						"id":         "conflict-1",
						"table":      "productos",
						"record_id":  "product-123",
						"resolution": "server_wins",
					},
					{
						"id":         "conflict-2",
						"table":      "productos",
						"record_id":  "product-456",
						"resolution": "terminal_wins",
					},
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "resolve with invalid resolution",
			requestBody: map[string]interface{}{
				"terminal_id": "TERM001",
				"conflictos": []map[string]interface{}{
					{
						"id":         "conflict-3",
						"table":      "productos",
						"record_id":  "product-789",
						"resolution": "invalid_resolution",
					},
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_RESOLUTION",
		},
		{
			name: "resolve non-existing conflict",
			requestBody: map[string]interface{}{
				"terminal_id": "TERM001",
				"conflictos": []map[string]interface{}{
					{
						"id":         "non-existing-conflict",
						"table":      "productos",
						"record_id":  "product-999",
						"resolution": "server_wins",
					},
				},
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "CONFLICT_NOT_FOUND",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := ts.MakeAuthenticatedRequest("POST", "/api/v1/sync/conflictos/resolver", tt.requestBody, terminalToken)
			
			assert.Equal(t, tt.expectedStatus, recorder.Code)
			
			var response models.APIResponse
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			require.NoError(t, err)
			
			if tt.expectedStatus == http.StatusOK {
				assert.True(t, response.Success)
				assert.NotNil(t, response.Data)
				
				// Verificar estadísticas de resolución
				data, ok := response.Data.(map[string]interface{})
				require.True(t, ok)
				assert.Contains(t, data, "resolved")
				assert.Contains(t, data, "failed")
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

func TestGetConflictos(t *testing.T) {
	ts := setupSyncTestServer(t)
	defer ts.TeardownTestDatabase(t)
	
	terminalToken := "valid-terminal-token"
	
	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
	}{
		{
			name:           "get all conflicts",
			queryParams:    "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "get conflicts for specific table",
			queryParams:    "?table=productos",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "get conflicts with pagination",
			queryParams:    "?limit=10&offset=0",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "get unresolved conflicts only",
			queryParams:    "?status=unresolved",
			expectedStatus: http.StatusOK,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := ts.MakeAuthenticatedRequest("GET", "/api/v1/sync/conflictos"+tt.queryParams, nil, terminalToken)
			
			assert.Equal(t, tt.expectedStatus, recorder.Code)
			
			var response models.APIResponse
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			require.NoError(t, err)
			
			if tt.expectedStatus == http.StatusOK {
				assert.True(t, response.Success)
				assert.NotNil(t, response.Data)
				
				// Verificar estructura de conflictos
				data, ok := response.Data.(map[string]interface{})
				require.True(t, ok)
				assert.Contains(t, data, "conflicts")
				assert.Contains(t, data, "total")
			}
		})
	}
}

func BenchmarkSyncProductos(b *testing.B) {
	gin.SetMode(gin.TestMode)
	ts := &testutils.TestServer{
		Router: gin.New(),
	}
	setupSyncRoutes(ts)
	
	requestBody := map[string]interface{}{
		"terminal_id": "TERM001",
		"timestamp":   time.Now().Unix(),
		"productos": []map[string]interface{}{
			{
				"id":     "bench-product-1",
				"codigo": "BENCH001",
				"nombre": "Producto Benchmark",
				"precio": 100.00,
				"stock":  10,
				"action": "create",
			},
		},
	}
	
	bodyBytes, _ := json.Marshal(requestBody)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("POST", "/api/v1/sync/productos", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer valid-terminal-token")
		
		recorder := httptest.NewRecorder()
		ts.Router.ServeHTTP(recorder, req)
	}
}

func BenchmarkGetCambios(b *testing.B) {
	gin.SetMode(gin.TestMode)
	ts := &testutils.TestServer{
		Router: gin.New(),
	}
	setupSyncRoutes(ts)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/api/v1/sync/cambios", nil)
		req.Header.Set("Authorization", "Bearer valid-terminal-token")
		
		recorder := httptest.NewRecorder()
		ts.Router.ServeHTTP(recorder, req)
	}
}

