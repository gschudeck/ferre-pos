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

func setupReportsTestServer(t *testing.T) *testutils.TestServer {
	ts := testutils.SetupTestServer(t)
	
	// Configurar rutas específicas del API Reports
	setupReportsRoutes(ts)
	
	return ts
}

func setupReportsRoutes(ts *testutils.TestServer) {
	// Mock dependencies
	mockDB := mocks.NewMockDatabase()
	mockValidator := &mocks.MockValidator{}
	mockMetrics := &mocks.MockMetrics{}
	
	// Handlers
	reportsHandler := handlers.NewReportsHandler(mockDB, logger.Get(), mockValidator, mockMetrics)
	
	// Rutas de reportes
	reports := ts.Router.Group("/api/v1/reports")
	{
		// Reportes de ventas
		reports.GET("/ventas", reportsHandler.ReporteVentas)
		reports.GET("/ventas/diarias", reportsHandler.VentasDiarias)
		reports.GET("/ventas/mensuales", reportsHandler.VentasMensuales)
		reports.GET("/ventas/por-producto", reportsHandler.VentasPorProducto)
		reports.GET("/ventas/por-usuario", reportsHandler.VentasPorUsuario)
		reports.GET("/ventas/por-sucursal", reportsHandler.VentasPorSucursal)
		
		// Reportes de inventario
		reports.GET("/inventario", reportsHandler.ReporteInventario)
		reports.GET("/inventario/stock-bajo", reportsHandler.StockBajo)
		reports.GET("/inventario/movimientos", reportsHandler.MovimientosStock)
		reports.GET("/inventario/valoracion", reportsHandler.ValoracionInventario)
		
		// Reportes financieros
		reports.GET("/financiero/resumen", reportsHandler.ResumenFinanciero)
		reports.GET("/financiero/flujo-caja", reportsHandler.FlujoCaja)
		reports.GET("/financiero/rentabilidad", reportsHandler.Rentabilidad)
		
		// Analytics y dashboards
		reports.GET("/dashboard/resumen", reportsHandler.DashboardResumen)
		reports.GET("/analytics/productos-top", reportsHandler.ProductosTop)
		reports.GET("/analytics/tendencias", reportsHandler.Tendencias)
		reports.GET("/analytics/clientes", reportsHandler.AnalyticsClientes)
		
		// Exportación de reportes
		reports.POST("/export", reportsHandler.ExportarReporte)
		reports.GET("/export/:id/status", reportsHandler.EstadoExportacion)
		reports.GET("/export/:id/download", reportsHandler.DescargarReporte)
		
		// Reportes personalizados
		reports.GET("/custom", reportsHandler.ReportesPersonalizados)
		reports.POST("/custom", reportsHandler.CrearReportePersonalizado)
		reports.PUT("/custom/:id", reportsHandler.ActualizarReportePersonalizado)
		reports.DELETE("/custom/:id", reportsHandler.EliminarReportePersonalizado)
		reports.POST("/custom/:id/execute", reportsHandler.EjecutarReportePersonalizado)
		
		// Configuración de reportes
		reports.GET("/config", reportsHandler.GetConfiguracion)
		reports.PUT("/config", reportsHandler.UpdateConfiguracion)
	}
}

func TestReportsVentas(t *testing.T) {
	ts := setupReportsTestServer(t)
	defer ts.TeardownTestDatabase(t)
	
	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "reporte ventas básico",
			queryParams:    "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "reporte ventas con rango de fechas",
			queryParams:    "?start_date=2024-01-01&end_date=2024-12-31",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "reporte ventas por sucursal",
			queryParams:    "?sucursal_id=test-sucursal-1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "reporte ventas con agrupación",
			queryParams:    "?group_by=day&start_date=2024-01-01&end_date=2024-01-31",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "reporte ventas con formato específico",
			queryParams:    "?format=pdf&start_date=2024-01-01&end_date=2024-01-31",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "reporte ventas con fecha inválida",
			queryParams:    "?start_date=invalid-date",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_DATE_FORMAT",
		},
		{
			name:           "reporte ventas con rango de fechas inválido",
			queryParams:    "?start_date=2024-12-31&end_date=2024-01-01",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_DATE_RANGE",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := ts.MakeRequest("GET", "/api/v1/reports/ventas"+tt.queryParams, nil, nil)
			
			assert.Equal(t, tt.expectedStatus, recorder.Code)
			
			var response models.APIResponse
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			require.NoError(t, err)
			
			if tt.expectedStatus == http.StatusOK {
				assert.True(t, response.Success)
				assert.NotNil(t, response.Data)
				
				// Verificar estructura del reporte
				data, ok := response.Data.(map[string]interface{})
				require.True(t, ok)
				assert.Contains(t, data, "summary")
				assert.Contains(t, data, "data")
				assert.Contains(t, data, "period")
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

func TestReportsVentasDiarias(t *testing.T) {
	ts := setupReportsTestServer(t)
	defer ts.TeardownTestDatabase(t)
	
	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
	}{
		{
			name:           "ventas diarias del mes actual",
			queryParams:    "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "ventas diarias de un mes específico",
			queryParams:    "?year=2024&month=1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "ventas diarias con comparación",
			queryParams:    "?year=2024&month=1&compare=true",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "ventas diarias por sucursal",
			queryParams:    "?sucursal_id=test-sucursal-1&year=2024&month=1",
			expectedStatus: http.StatusOK,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := ts.MakeRequest("GET", "/api/v1/reports/ventas/diarias"+tt.queryParams, nil, nil)
			
			assert.Equal(t, tt.expectedStatus, recorder.Code)
			
			var response models.APIResponse
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			require.NoError(t, err)
			
			if tt.expectedStatus == http.StatusOK {
				assert.True(t, response.Success)
				assert.NotNil(t, response.Data)
				
				// Verificar estructura específica de ventas diarias
				data, ok := response.Data.(map[string]interface{})
				require.True(t, ok)
				assert.Contains(t, data, "daily_sales")
				assert.Contains(t, data, "total_period")
				assert.Contains(t, data, "average_daily")
			}
		})
	}
}

func TestReportsInventario(t *testing.T) {
	ts := setupReportsTestServer(t)
	defer ts.TeardownTestDatabase(t)
	
	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
	}{
		{
			name:           "reporte inventario completo",
			queryParams:    "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "reporte inventario por categoría",
			queryParams:    "?categoria_id=test-category-1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "reporte inventario con stock bajo",
			queryParams:    "?stock_bajo=true",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "reporte inventario con valoración",
			queryParams:    "?incluir_valoracion=true",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "reporte inventario por sucursal",
			queryParams:    "?sucursal_id=test-sucursal-1",
			expectedStatus: http.StatusOK,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := ts.MakeRequest("GET", "/api/v1/reports/inventario"+tt.queryParams, nil, nil)
			
			assert.Equal(t, tt.expectedStatus, recorder.Code)
			
			var response models.APIResponse
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			require.NoError(t, err)
			
			if tt.expectedStatus == http.StatusOK {
				assert.True(t, response.Success)
				assert.NotNil(t, response.Data)
				
				// Verificar estructura del reporte de inventario
				data, ok := response.Data.(map[string]interface{})
				require.True(t, ok)
				assert.Contains(t, data, "productos")
				assert.Contains(t, data, "resumen")
				assert.Contains(t, data, "total_items")
			}
		})
	}
}

func TestReportsStockBajo(t *testing.T) {
	ts := setupReportsTestServer(t)
	defer ts.TeardownTestDatabase(t)
	
	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
	}{
		{
			name:           "productos con stock bajo",
			queryParams:    "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "stock bajo por categoría",
			queryParams:    "?categoria_id=test-category-1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "stock bajo con umbral personalizado",
			queryParams:    "?umbral=5",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "stock bajo por sucursal",
			queryParams:    "?sucursal_id=test-sucursal-1",
			expectedStatus: http.StatusOK,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := ts.MakeRequest("GET", "/api/v1/reports/inventario/stock-bajo"+tt.queryParams, nil, nil)
			
			assert.Equal(t, tt.expectedStatus, recorder.Code)
			
			var response models.APIResponse
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			require.NoError(t, err)
			
			if tt.expectedStatus == http.StatusOK {
				assert.True(t, response.Success)
				assert.NotNil(t, response.Data)
				
				// Verificar estructura del reporte de stock bajo
				data, ok := response.Data.(map[string]interface{})
				require.True(t, ok)
				assert.Contains(t, data, "productos_stock_bajo")
				assert.Contains(t, data, "total_productos")
				assert.Contains(t, data, "umbral_utilizado")
			}
		})
	}
}

func TestReportsDashboardResumen(t *testing.T) {
	ts := setupReportsTestServer(t)
	defer ts.TeardownTestDatabase(t)
	
	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
	}{
		{
			name:           "dashboard resumen básico",
			queryParams:    "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "dashboard resumen con período",
			queryParams:    "?period=week",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "dashboard resumen por sucursal",
			queryParams:    "?sucursal_id=test-sucursal-1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "dashboard resumen con comparación",
			queryParams:    "?compare_previous=true",
			expectedStatus: http.StatusOK,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := ts.MakeRequest("GET", "/api/v1/reports/dashboard/resumen"+tt.queryParams, nil, nil)
			
			assert.Equal(t, tt.expectedStatus, recorder.Code)
			
			var response models.APIResponse
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			require.NoError(t, err)
			
			if tt.expectedStatus == http.StatusOK {
				assert.True(t, response.Success)
				assert.NotNil(t, response.Data)
				
				// Verificar estructura del dashboard
				data, ok := response.Data.(map[string]interface{})
				require.True(t, ok)
				assert.Contains(t, data, "ventas_hoy")
				assert.Contains(t, data, "ventas_mes")
				assert.Contains(t, data, "productos_vendidos")
				assert.Contains(t, data, "stock_bajo_count")
				assert.Contains(t, data, "tendencias")
			}
		})
	}
}

func TestReportsExportarReporte(t *testing.T) {
	ts := setupReportsTestServer(t)
	defer ts.TeardownTestDatabase(t)
	
	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "exportar reporte de ventas en PDF",
			requestBody: map[string]interface{}{
				"tipo_reporte": "ventas",
				"formato":      "pdf",
				"parametros": map[string]interface{}{
					"start_date": "2024-01-01",
					"end_date":   "2024-01-31",
				},
			},
			expectedStatus: http.StatusAccepted,
		},
		{
			name: "exportar reporte de inventario en Excel",
			requestBody: map[string]interface{}{
				"tipo_reporte": "inventario",
				"formato":      "xlsx",
				"parametros": map[string]interface{}{
					"incluir_valoracion": true,
				},
			},
			expectedStatus: http.StatusAccepted,
		},
		{
			name: "exportar con formato no soportado",
			requestBody: map[string]interface{}{
				"tipo_reporte": "ventas",
				"formato":      "unsupported_format",
				"parametros":   map[string]interface{}{},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "UNSUPPORTED_FORMAT",
		},
		{
			name: "exportar reporte inexistente",
			requestBody: map[string]interface{}{
				"tipo_reporte": "non_existing_report",
				"formato":      "pdf",
				"parametros":   map[string]interface{}{},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_REPORT_TYPE",
		},
		{
			name: "exportar sin parámetros requeridos",
			requestBody: map[string]interface{}{
				"formato": "pdf",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := ts.MakeRequest("POST", "/api/v1/reports/export", tt.requestBody, nil)
			
			assert.Equal(t, tt.expectedStatus, recorder.Code)
			
			var response models.APIResponse
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			require.NoError(t, err)
			
			if tt.expectedStatus == http.StatusAccepted {
				assert.True(t, response.Success)
				assert.NotNil(t, response.Data)
				
				// Verificar estructura de respuesta de exportación
				data, ok := response.Data.(map[string]interface{})
				require.True(t, ok)
				assert.Contains(t, data, "export_id")
				assert.Contains(t, data, "status")
				assert.Contains(t, data, "estimated_time")
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

func TestReportsEstadoExportacion(t *testing.T) {
	ts := setupReportsTestServer(t)
	defer ts.TeardownTestDatabase(t)
	
	tests := []struct {
		name           string
		exportID       string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "estado de exportación existente",
			exportID:       "test-export-1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "estado de exportación completada",
			exportID:       "test-export-completed",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "estado de exportación en progreso",
			exportID:       "test-export-in-progress",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "exportación no encontrada",
			exportID:       "non-existing-export",
			expectedStatus: http.StatusNotFound,
			expectedError:  "EXPORT_NOT_FOUND",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/v1/reports/export/" + tt.exportID + "/status"
			recorder := ts.MakeRequest("GET", url, nil, nil)
			
			assert.Equal(t, tt.expectedStatus, recorder.Code)
			
			var response models.APIResponse
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			require.NoError(t, err)
			
			if tt.expectedStatus == http.StatusOK {
				assert.True(t, response.Success)
				assert.NotNil(t, response.Data)
				
				// Verificar estructura del estado
				data, ok := response.Data.(map[string]interface{})
				require.True(t, ok)
				assert.Contains(t, data, "export_id")
				assert.Contains(t, data, "status")
				assert.Contains(t, data, "progress")
				assert.Contains(t, data, "created_at")
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

func TestReportsCrearReportePersonalizado(t *testing.T) {
	ts := setupReportsTestServer(t)
	defer ts.TeardownTestDatabase(t)
	
	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "crear reporte personalizado válido",
			requestBody: map[string]interface{}{
				"nombre":      "Reporte Personalizado Test",
				"descripcion": "Descripción del reporte personalizado",
				"tipo":        "ventas",
				"consulta_sql": "SELECT * FROM ventas WHERE fecha_venta >= ? AND fecha_venta <= ?",
				"parametros": []map[string]interface{}{
					{
						"nombre":      "start_date",
						"tipo":        "date",
						"requerido":   true,
						"descripcion": "Fecha de inicio",
					},
					{
						"nombre":      "end_date",
						"tipo":        "date",
						"requerido":   true,
						"descripcion": "Fecha de fin",
					},
				},
				"columnas": []map[string]interface{}{
					{
						"nombre":    "fecha_venta",
						"titulo":    "Fecha",
						"tipo":      "date",
						"formato":   "DD/MM/YYYY",
					},
					{
						"nombre":    "total",
						"titulo":    "Total",
						"tipo":      "currency",
						"formato":   "$#,##0.00",
					},
				},
				"activo": true,
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "crear reporte con consulta SQL inválida",
			requestBody: map[string]interface{}{
				"nombre":       "Reporte Inválido",
				"descripcion":  "Reporte con SQL inválido",
				"tipo":         "ventas",
				"consulta_sql": "DROP TABLE ventas;", // SQL peligroso
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_SQL_QUERY",
		},
		{
			name: "crear reporte sin nombre",
			requestBody: map[string]interface{}{
				"descripcion":  "Reporte sin nombre",
				"tipo":         "ventas",
				"consulta_sql": "SELECT * FROM ventas",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
		{
			name: "crear reporte con nombre duplicado",
			requestBody: map[string]interface{}{
				"nombre":       "Reporte Existente",
				"descripcion":  "Reporte con nombre duplicado",
				"tipo":         "ventas",
				"consulta_sql": "SELECT * FROM ventas",
			},
			expectedStatus: http.StatusConflict,
			expectedError:  "REPORT_NAME_EXISTS",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := ts.MakeRequest("POST", "/api/v1/reports/custom", tt.requestBody, nil)
			
			assert.Equal(t, tt.expectedStatus, recorder.Code)
			
			var response models.APIResponse
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			require.NoError(t, err)
			
			if tt.expectedStatus == http.StatusCreated {
				assert.True(t, response.Success)
				assert.NotNil(t, response.Data)
				
				// Verificar estructura del reporte creado
				data, ok := response.Data.(map[string]interface{})
				require.True(t, ok)
				assert.Contains(t, data, "id")
				assert.Contains(t, data, "nombre")
				assert.Contains(t, data, "created_at")
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

func TestReportsProductosTop(t *testing.T) {
	ts := setupReportsTestServer(t)
	defer ts.TeardownTestDatabase(t)
	
	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
	}{
		{
			name:           "productos top por ventas",
			queryParams:    "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "top 10 productos",
			queryParams:    "?limit=10",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "productos top por período",
			queryParams:    "?start_date=2024-01-01&end_date=2024-01-31",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "productos top por categoría",
			queryParams:    "?categoria_id=test-category-1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "productos top por sucursal",
			queryParams:    "?sucursal_id=test-sucursal-1",
			expectedStatus: http.StatusOK,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := ts.MakeRequest("GET", "/api/v1/reports/analytics/productos-top"+tt.queryParams, nil, nil)
			
			assert.Equal(t, tt.expectedStatus, recorder.Code)
			
			var response models.APIResponse
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			require.NoError(t, err)
			
			if tt.expectedStatus == http.StatusOK {
				assert.True(t, response.Success)
				assert.NotNil(t, response.Data)
				
				// Verificar estructura de productos top
				data, ok := response.Data.(map[string]interface{})
				require.True(t, ok)
				assert.Contains(t, data, "productos")
				assert.Contains(t, data, "period")
				assert.Contains(t, data, "total_sales")
			}
		})
	}
}

func BenchmarkReportsVentas(b *testing.B) {
	gin.SetMode(gin.TestMode)
	ts := &testutils.TestServer{
		Router: gin.New(),
	}
	setupReportsRoutes(ts)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/api/v1/reports/ventas", nil)
		
		recorder := httptest.NewRecorder()
		ts.Router.ServeHTTP(recorder, req)
	}
}

func BenchmarkReportsDashboardResumen(b *testing.B) {
	gin.SetMode(gin.TestMode)
	ts := &testutils.TestServer{
		Router: gin.New(),
	}
	setupReportsRoutes(ts)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/api/v1/reports/dashboard/resumen", nil)
		
		recorder := httptest.NewRecorder()
		ts.Router.ServeHTTP(recorder, req)
	}
}

func BenchmarkReportsExportarReporte(b *testing.B) {
	gin.SetMode(gin.TestMode)
	ts := &testutils.TestServer{
		Router: gin.New(),
	}
	setupReportsRoutes(ts)
	
	requestBody := map[string]interface{}{
		"tipo_reporte": "ventas",
		"formato":      "pdf",
		"parametros": map[string]interface{}{
			"start_date": "2024-01-01",
			"end_date":   "2024-01-31",
		},
	}
	
	bodyBytes, _ := json.Marshal(requestBody)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("POST", "/api/v1/reports/export", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		
		recorder := httptest.NewRecorder()
		ts.Router.ServeHTTP(recorder, req)
	}
}

