package unit

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

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

func setupLabelsTestServer(t *testing.T) *testutils.TestServer {
	ts := testutils.SetupTestServer(t)
	
	// Configurar rutas específicas del API Labels
	setupLabelsRoutes(ts)
	
	return ts
}

func setupLabelsRoutes(ts *testutils.TestServer) {
	// Mock dependencies
	mockDB := mocks.NewMockDatabase()
	mockValidator := &mocks.MockValidator{}
	mockMetrics := &mocks.MockMetrics{}
	
	// Handlers
	labelsHandler := handlers.NewLabelsHandler(mockDB, logger.Get(), mockValidator, mockMetrics)
	
	// Rutas de etiquetas
	labels := ts.Router.Group("/api/v1/labels")
	{
		// Generación de etiquetas
		labels.POST("/generate", labelsHandler.GenerateLabels)
		labels.POST("/generate/batch", labelsHandler.GenerateBatchLabels)
		
		// Plantillas de etiquetas
		labels.GET("/templates", labelsHandler.GetTemplates)
		labels.GET("/templates/:id", labelsHandler.GetTemplate)
		labels.POST("/templates", labelsHandler.CreateTemplate)
		labels.PUT("/templates/:id", labelsHandler.UpdateTemplate)
		labels.DELETE("/templates/:id", labelsHandler.DeleteTemplate)
		
		// Códigos de barras
		labels.POST("/barcode/generate", labelsHandler.GenerateBarcode)
		labels.POST("/barcode/validate", labelsHandler.ValidateBarcode)
		labels.GET("/barcode/formats", labelsHandler.GetBarcodeFormats)
		
		// Previsualización
		labels.POST("/preview", labelsHandler.PreviewLabel)
		labels.POST("/preview/batch", labelsHandler.PreviewBatchLabels)
		
		// Historial
		labels.GET("/history", labelsHandler.GetHistory)
		labels.GET("/history/:id", labelsHandler.GetHistoryItem)
		
		// Configuración
		labels.GET("/config", labelsHandler.GetConfig)
		labels.PUT("/config", labelsHandler.UpdateConfig)
	}
}

func TestLabelsGenerateLabels(t *testing.T) {
	ts := setupLabelsTestServer(t)
	defer ts.TeardownTestDatabase(t)
	
	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "generate single product label",
			requestBody: map[string]interface{}{
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
				"opciones": map[string]interface{}{
					"cantidad_por_producto": 1,
					"incluir_precio":        true,
					"incluir_codigo":        true,
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "generate multiple product labels",
			requestBody: map[string]interface{}{
				"template_id": "template-basic",
				"format":      "pdf",
				"productos": []map[string]interface{}{
					{
						"id":            "test-product-1",
						"codigo":        "TEST001",
						"nombre":        "Producto Test 1",
						"precio":        100.00,
						"codigo_barras": "1234567890123",
					},
					{
						"id":            "test-product-2",
						"codigo":        "TEST002",
						"nombre":        "Producto Test 2",
						"precio":        200.00,
						"codigo_barras": "1234567890124",
					},
				},
				"opciones": map[string]interface{}{
					"cantidad_por_producto": 2,
					"incluir_precio":        true,
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "generate with invalid template",
			requestBody: map[string]interface{}{
				"template_id": "non-existing-template",
				"format":      "pdf",
				"productos": []map[string]interface{}{
					{
						"id":     "test-product-1",
						"codigo": "TEST001",
						"nombre": "Producto Test",
						"precio": 100.00,
					},
				},
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "TEMPLATE_NOT_FOUND",
		},
		{
			name: "generate with invalid format",
			requestBody: map[string]interface{}{
				"template_id": "template-basic",
				"format":      "invalid-format",
				"productos": []map[string]interface{}{
					{
						"id":     "test-product-1",
						"codigo": "TEST001",
						"nombre": "Producto Test",
						"precio": 100.00,
					},
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_FORMAT",
		},
		{
			name: "generate without products",
			requestBody: map[string]interface{}{
				"template_id": "template-basic",
				"format":      "pdf",
				"productos":   []map[string]interface{}{},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "NO_PRODUCTS_PROVIDED",
		},
		{
			name: "generate with missing required fields",
			requestBody: map[string]interface{}{
				"format": "pdf",
				"productos": []map[string]interface{}{
					{
						"id":     "test-product-1",
						"codigo": "TEST001",
					},
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := ts.MakeRequest("POST", "/api/v1/labels/generate", tt.requestBody, nil)
			
			assert.Equal(t, tt.expectedStatus, recorder.Code)
			
			var response models.APIResponse
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			require.NoError(t, err)
			
			if tt.expectedStatus == http.StatusOK {
				assert.True(t, response.Success)
				assert.NotNil(t, response.Data)
				
				// Verificar estructura de respuesta
				data, ok := response.Data.(map[string]interface{})
				require.True(t, ok)
				assert.Contains(t, data, "file_url")
				assert.Contains(t, data, "file_size")
				assert.Contains(t, data, "labels_generated")
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

func TestLabelsGenerateBatchLabels(t *testing.T) {
	ts := setupLabelsTestServer(t)
	defer ts.TeardownTestDatabase(t)
	
	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "generate batch labels successfully",
			requestBody: map[string]interface{}{
				"template_id": "template-basic",
				"format":      "pdf",
				"filtros": map[string]interface{}{
					"categoria_id": "test-category-1",
					"stock_minimo": true,
					"activos":      true,
				},
				"opciones": map[string]interface{}{
					"cantidad_por_producto": 1,
					"incluir_precio":        true,
					"incluir_codigo":        true,
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "generate batch with product IDs",
			requestBody: map[string]interface{}{
				"template_id": "template-basic",
				"format":      "pdf",
				"producto_ids": []string{
					"test-product-1",
					"test-product-2",
					"test-product-3",
				},
				"opciones": map[string]interface{}{
					"cantidad_por_producto": 2,
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "generate batch with no matching products",
			requestBody: map[string]interface{}{
				"template_id": "template-basic",
				"format":      "pdf",
				"filtros": map[string]interface{}{
					"categoria_id": "non-existing-category",
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "NO_PRODUCTS_FOUND",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := ts.MakeRequest("POST", "/api/v1/labels/generate/batch", tt.requestBody, nil)
			
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

func TestLabelsGetTemplates(t *testing.T) {
	ts := setupLabelsTestServer(t)
	defer ts.TeardownTestDatabase(t)
	
	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
	}{
		{
			name:           "get all templates",
			queryParams:    "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "get templates with pagination",
			queryParams:    "?limit=10&offset=0",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "get active templates only",
			queryParams:    "?active=true",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "get templates by category",
			queryParams:    "?category=product",
			expectedStatus: http.StatusOK,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := ts.MakeRequest("GET", "/api/v1/labels/templates"+tt.queryParams, nil, nil)
			
			assert.Equal(t, tt.expectedStatus, recorder.Code)
			
			var response models.APIResponse
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			require.NoError(t, err)
			
			if tt.expectedStatus == http.StatusOK {
				assert.True(t, response.Success)
				assert.NotNil(t, response.Data)
			}
		})
	}
}

func TestLabelsCreateTemplate(t *testing.T) {
	ts := setupLabelsTestServer(t)
	defer ts.TeardownTestDatabase(t)
	
	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "create template successfully",
			requestBody: map[string]interface{}{
				"nombre":      "Template Test",
				"descripcion": "Template de prueba",
				"categoria":   "product",
				"dimensiones": map[string]interface{}{
					"ancho":  50.0,
					"alto":   30.0,
					"unidad": "mm",
				},
				"elementos": []map[string]interface{}{
					{
						"tipo":     "text",
						"campo":    "nombre",
						"x":        5.0,
						"y":        5.0,
						"fuente":   "Arial",
						"tamaño":   12,
					},
					{
						"tipo":     "barcode",
						"campo":    "codigo_barras",
						"x":        5.0,
						"y":        15.0,
						"formato":  "CODE128",
					},
				},
				"activo": true,
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "create template with invalid dimensions",
			requestBody: map[string]interface{}{
				"nombre":      "Template Invalid",
				"descripcion": "Template con dimensiones inválidas",
				"categoria":   "product",
				"dimensiones": map[string]interface{}{
					"ancho":  -10.0, // Ancho negativo
					"alto":   30.0,
					"unidad": "mm",
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_DIMENSIONS",
		},
		{
			name: "create template with missing required fields",
			requestBody: map[string]interface{}{
				"descripcion": "Template sin nombre",
				"categoria":   "product",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
		{
			name: "create template with duplicate name",
			requestBody: map[string]interface{}{
				"nombre":      "Existing Template",
				"descripcion": "Template duplicado",
				"categoria":   "product",
				"dimensiones": map[string]interface{}{
					"ancho":  50.0,
					"alto":   30.0,
					"unidad": "mm",
				},
			},
			expectedStatus: http.StatusConflict,
			expectedError:  "TEMPLATE_NAME_EXISTS",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := ts.MakeRequest("POST", "/api/v1/labels/templates", tt.requestBody, nil)
			
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

func TestLabelsGenerateBarcode(t *testing.T) {
	ts := setupLabelsTestServer(t)
	defer ts.TeardownTestDatabase(t)
	
	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "generate barcode successfully",
			requestBody: map[string]interface{}{
				"data":    "1234567890123",
				"format":  "CODE128",
				"width":   200,
				"height":  50,
				"output":  "png",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "generate EAN-13 barcode",
			requestBody: map[string]interface{}{
				"data":    "1234567890123",
				"format":  "EAN13",
				"width":   150,
				"height":  40,
				"output":  "svg",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "generate with invalid format",
			requestBody: map[string]interface{}{
				"data":   "1234567890123",
				"format": "INVALID_FORMAT",
				"width":  200,
				"height": 50,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_BARCODE_FORMAT",
		},
		{
			name: "generate with invalid data for format",
			requestBody: map[string]interface{}{
				"data":   "INVALID_DATA_FOR_EAN13",
				"format": "EAN13",
				"width":  200,
				"height": 50,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_BARCODE_DATA",
		},
		{
			name: "generate with missing required fields",
			requestBody: map[string]interface{}{
				"format": "CODE128",
				"width":  200,
				"height": 50,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := ts.MakeRequest("POST", "/api/v1/labels/barcode/generate", tt.requestBody, nil)
			
			assert.Equal(t, tt.expectedStatus, recorder.Code)
			
			var response models.APIResponse
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			require.NoError(t, err)
			
			if tt.expectedStatus == http.StatusOK {
				assert.True(t, response.Success)
				assert.NotNil(t, response.Data)
				
				// Verificar estructura de respuesta
				data, ok := response.Data.(map[string]interface{})
				require.True(t, ok)
				assert.Contains(t, data, "barcode_url")
				assert.Contains(t, data, "format")
				assert.Contains(t, data, "dimensions")
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

func TestLabelsValidateBarcode(t *testing.T) {
	ts := setupLabelsTestServer(t)
	defer ts.TeardownTestDatabase(t)
	
	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "validate valid EAN-13",
			requestBody: map[string]interface{}{
				"data":   "1234567890123",
				"format": "EAN13",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "validate valid CODE128",
			requestBody: map[string]interface{}{
				"data":   "HELLO123",
				"format": "CODE128",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "validate invalid EAN-13",
			requestBody: map[string]interface{}{
				"data":   "123", // Muy corto para EAN-13
				"format": "EAN13",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_BARCODE_DATA",
		},
		{
			name: "validate with unsupported format",
			requestBody: map[string]interface{}{
				"data":   "1234567890123",
				"format": "UNSUPPORTED_FORMAT",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "UNSUPPORTED_FORMAT",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := ts.MakeRequest("POST", "/api/v1/labels/barcode/validate", tt.requestBody, nil)
			
			assert.Equal(t, tt.expectedStatus, recorder.Code)
			
			var response models.APIResponse
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			require.NoError(t, err)
			
			if tt.expectedStatus == http.StatusOK {
				assert.True(t, response.Success)
				assert.NotNil(t, response.Data)
				
				// Verificar estructura de respuesta
				data, ok := response.Data.(map[string]interface{})
				require.True(t, ok)
				assert.Contains(t, data, "valid")
				assert.Contains(t, data, "format")
				assert.Contains(t, data, "checksum_valid")
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

func TestLabelsPreviewLabel(t *testing.T) {
	ts := setupLabelsTestServer(t)
	defer ts.TeardownTestDatabase(t)
	
	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
	}{
		{
			name: "preview label successfully",
			requestBody: map[string]interface{}{
				"template_id": "template-basic",
				"producto": map[string]interface{}{
					"id":            "test-product-1",
					"codigo":        "TEST001",
					"nombre":        "Producto Test",
					"precio":        100.00,
					"codigo_barras": "1234567890123",
				},
				"opciones": map[string]interface{}{
					"incluir_precio": true,
					"incluir_codigo": true,
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "preview with non-existing template",
			requestBody: map[string]interface{}{
				"template_id": "non-existing-template",
				"producto": map[string]interface{}{
					"id":     "test-product-1",
					"codigo": "TEST001",
					"nombre": "Producto Test",
					"precio": 100.00,
				},
			},
			expectedStatus: http.StatusNotFound,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := ts.MakeRequest("POST", "/api/v1/labels/preview", tt.requestBody, nil)
			
			assert.Equal(t, tt.expectedStatus, recorder.Code)
			
			var response models.APIResponse
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			require.NoError(t, err)
			
			if tt.expectedStatus == http.StatusOK {
				assert.True(t, response.Success)
				assert.NotNil(t, response.Data)
				
				// Verificar estructura de respuesta
				data, ok := response.Data.(map[string]interface{})
				require.True(t, ok)
				assert.Contains(t, data, "preview_url")
				assert.Contains(t, data, "template_info")
			} else {
				assert.False(t, response.Success)
			}
		})
	}
}

func TestLabelsGetBarcodeFormats(t *testing.T) {
	ts := setupLabelsTestServer(t)
	defer ts.TeardownTestDatabase(t)
	
	recorder := ts.MakeRequest("GET", "/api/v1/labels/barcode/formats", nil, nil)
	
	assert.Equal(t, http.StatusOK, recorder.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.True(t, response.Success)
	assert.NotNil(t, response.Data)
	
	// Verificar que se devuelven los formatos soportados
	data, ok := response.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, data, "formats")
	
	formats, ok := data["formats"].([]interface{})
	require.True(t, ok)
	assert.Greater(t, len(formats), 0)
}

func TestLabelsGetHistory(t *testing.T) {
	ts := setupLabelsTestServer(t)
	defer ts.TeardownTestDatabase(t)
	
	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
	}{
		{
			name:           "get all history",
			queryParams:    "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "get history with pagination",
			queryParams:    "?limit=10&offset=0",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "get history by user",
			queryParams:    "?user_id=test-user-1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "get history by date range",
			queryParams:    "?start_date=2024-01-01&end_date=2024-12-31",
			expectedStatus: http.StatusOK,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := ts.MakeRequest("GET", "/api/v1/labels/history"+tt.queryParams, nil, nil)
			
			assert.Equal(t, tt.expectedStatus, recorder.Code)
			
			var response models.APIResponse
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			require.NoError(t, err)
			
			if tt.expectedStatus == http.StatusOK {
				assert.True(t, response.Success)
				assert.NotNil(t, response.Data)
			}
		})
	}
}

func BenchmarkLabelsGenerateLabels(b *testing.B) {
	gin.SetMode(gin.TestMode)
	ts := &testutils.TestServer{
		Router: gin.New(),
	}
	setupLabelsRoutes(ts)
	
	requestBody := map[string]interface{}{
		"template_id": "template-basic",
		"format":      "pdf",
		"productos": []map[string]interface{}{
			{
				"id":            "bench-product-1",
				"codigo":        "BENCH001",
				"nombre":        "Producto Benchmark",
				"precio":        100.00,
				"codigo_barras": "1234567890123",
			},
		},
	}
	
	bodyBytes, _ := json.Marshal(requestBody)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("POST", "/api/v1/labels/generate", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		
		recorder := httptest.NewRecorder()
		ts.Router.ServeHTTP(recorder, req)
	}
}

func BenchmarkLabelsGenerateBarcode(b *testing.B) {
	gin.SetMode(gin.TestMode)
	ts := &testutils.TestServer{
		Router: gin.New(),
	}
	setupLabelsRoutes(ts)
	
	requestBody := map[string]interface{}{
		"data":   "1234567890123",
		"format": "CODE128",
		"width":  200,
		"height": 50,
		"output": "png",
	}
	
	bodyBytes, _ := json.Marshal(requestBody)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("POST", "/api/v1/labels/barcode/generate", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		
		recorder := httptest.NewRecorder()
		ts.Router.ServeHTTP(recorder, req)
	}
}

