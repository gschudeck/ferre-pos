//go:build e2e
// +build e2e

package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"ferre_pos_apis/internal/models"
)

// E2ETestSuite suite de tests end-to-end
type E2ETestSuite struct {
	suite.Suite
	processes map[string]*os.Process
	baseURLs  map[string]string
	tokens    map[string]string
}

// SetupSuite configuración inicial de la suite E2E
func (suite *E2ETestSuite) SetupSuite() {
	suite.processes = make(map[string]*os.Process)
	suite.baseURLs = make(map[string]string)
	suite.tokens = make(map[string]string)
	
	// Configurar URLs base
	suite.baseURLs["pos"] = getEnvOrDefault("E2E_POS_URL", "http://localhost:8080")
	suite.baseURLs["sync"] = getEnvOrDefault("E2E_SYNC_URL", "http://localhost:8081")
	suite.baseURLs["labels"] = getEnvOrDefault("E2E_LABELS_URL", "http://localhost:8082")
	suite.baseURLs["reports"] = getEnvOrDefault("E2E_REPORTS_URL", "http://localhost:8083")
	
	// Verificar si los servicios están ejecutándose o iniciarlos
	suite.ensureServicesRunning()
	
	// Esperar a que los servicios estén listos
	suite.waitForServices()
	
	// Configurar autenticación
	suite.setupAuthentication()
}

// TearDownSuite limpieza final de la suite E2E
func (suite *E2ETestSuite) TearDownSuite() {
	// Detener procesos si fueron iniciados por la suite
	for name, process := range suite.processes {
		if process != nil {
			suite.T().Logf("Stopping %s service (PID: %d)", name, process.Pid)
			process.Signal(syscall.SIGTERM)
			
			// Esperar a que termine gracefully
			done := make(chan error, 1)
			go func() {
				_, err := process.Wait()
				done <- err
			}()
			
			select {
			case <-done:
				suite.T().Logf("Service %s stopped gracefully", name)
			case <-time.After(10 * time.Second):
				suite.T().Logf("Force killing service %s", name)
				process.Kill()
			}
		}
	}
}

// ensureServicesRunning verifica que los servicios estén ejecutándose
func (suite *E2ETestSuite) ensureServicesRunning() {
	services := []string{"pos", "sync", "labels", "reports"}
	
	for _, service := range services {
		if !suite.isServiceRunning(service) {
			suite.T().Logf("Starting %s service", service)
			suite.startService(service)
		} else {
			suite.T().Logf("Service %s is already running", service)
		}
	}
}

// isServiceRunning verifica si un servicio está ejecutándose
func (suite *E2ETestSuite) isServiceRunning(service string) bool {
	url := suite.baseURLs[service] + "/health"
	
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	
	return resp.StatusCode == http.StatusOK
}

// startService inicia un servicio
func (suite *E2ETestSuite) startService(service string) {
	binaryPath := fmt.Sprintf("../../bin/api_%s", service)
	
	// Verificar que el binario existe
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		suite.T().Fatalf("Binary not found: %s. Run 'make build' first.", binaryPath)
	}
	
	// Configurar variables de entorno para el servicio
	cmd := exec.Command(binaryPath)
	cmd.Env = append(os.Environ(),
		"GIN_MODE=test",
		"LOG_LEVEL=error",
		fmt.Sprintf("CONFIG_PATH=../../configs/config.yaml"),
	)
	
	// Iniciar el proceso
	err := cmd.Start()
	require.NoError(suite.T(), err, "Failed to start %s service", service)
	
	suite.processes[service] = cmd.Process
	suite.T().Logf("Started %s service with PID: %d", service, cmd.Process.Pid)
}

// waitForServices espera a que todos los servicios estén listos
func (suite *E2ETestSuite) waitForServices() {
	services := []string{"pos", "sync", "labels", "reports"}
	timeout := 30 * time.Second
	interval := 1 * time.Second
	
	for _, service := range services {
		suite.T().Logf("Waiting for %s service to be ready...", service)
		
		start := time.Now()
		for time.Since(start) < timeout {
			if suite.isServiceRunning(service) {
				suite.T().Logf("Service %s is ready", service)
				break
			}
			time.Sleep(interval)
		}
		
		if time.Since(start) >= timeout {
			suite.T().Fatalf("Service %s did not become ready within %v", service, timeout)
		}
	}
}

// setupAuthentication configura la autenticación para los tests
func (suite *E2ETestSuite) setupAuthentication() {
	// Autenticación en API POS
	loginData := map[string]interface{}{
		"username": "admin_test",
		"password": "password123",
	}
	
	resp := suite.makeRequest("pos", "POST", "/api/v1/auth/login", loginData, "")
	if resp.StatusCode == http.StatusOK {
		var result models.APIResponse
		json.NewDecoder(resp.Body).Decode(&result)
		if result.Success {
			if data, ok := result.Data.(map[string]interface{}); ok {
				if token, ok := data["token"].(string); ok {
					suite.tokens["pos"] = token
					suite.T().Logf("Obtained POS authentication token")
				}
			}
		}
	}
	resp.Body.Close()
	
	// Autenticación de terminal en API Sync
	terminalData := map[string]interface{}{
		"terminal_id": "TERM001",
		"api_key":     "valid-api-key",
		"sucursal_id": "test-sucursal-1",
	}
	
	resp = suite.makeRequest("sync", "POST", "/api/v1/sync/auth", terminalData, "")
	if resp.StatusCode == http.StatusOK {
		var result models.APIResponse
		json.NewDecoder(resp.Body).Decode(&result)
		if result.Success {
			if data, ok := result.Data.(map[string]interface{}); ok {
				if token, ok := data["token"].(string); ok {
					suite.tokens["sync"] = token
					suite.T().Logf("Obtained Sync authentication token")
				}
			}
		}
	}
	resp.Body.Close()
}

// makeRequest realiza una petición HTTP a un servicio
func (suite *E2ETestSuite) makeRequest(service, method, path string, body interface{}, token string) *http.Response {
	baseURL, exists := suite.baseURLs[service]
	require.True(suite.T(), exists, "Service URL not configured: %s", service)
	
	var bodyReader *bytes.Buffer
	if body != nil {
		bodyBytes, _ := json.Marshal(body)
		bodyReader = bytes.NewBuffer(bodyBytes)
	} else {
		bodyReader = bytes.NewBuffer(nil)
	}
	
	url := baseURL + path
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

// TestHealthChecks verifica que todos los servicios respondan a health checks
func (suite *E2ETestSuite) TestHealthChecks() {
	services := []string{"pos", "sync", "labels", "reports"}
	
	for _, service := range services {
		suite.T().Run(fmt.Sprintf("HealthCheck_%s", service), func(t *testing.T) {
			resp := suite.makeRequest(service, "GET", "/health", nil, "")
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		})
	}
}

// TestCompleteBusinessFlow test de flujo completo de negocio
func (suite *E2ETestSuite) TestCompleteBusinessFlow() {
	// 1. Autenticación
	suite.T().Log("Step 1: Authentication")
	posToken := suite.tokens["pos"]
	require.NotEmpty(suite.T(), posToken, "POS token should be available")
	
	// 2. Crear producto
	suite.T().Log("Step 2: Create product")
	productData := map[string]interface{}{
		"codigo":        "E2E001",
		"codigo_barras": "1234567890111",
		"nombre":        "Producto E2E Test",
		"descripcion":   "Producto para test end-to-end",
		"categoria_id":  "test-category-1",
		"precio":        250.00,
		"costo":         150.00,
		"stock":         75,
		"stock_minimo":  15,
	}
	
	resp := suite.makeRequest("pos", "POST", "/api/v1/productos", productData, posToken)
	assert.Equal(suite.T(), http.StatusCreated, resp.StatusCode)
	
	var createResult models.APIResponse
	err := json.NewDecoder(resp.Body).Decode(&createResult)
	require.NoError(suite.T(), err)
	resp.Body.Close()
	
	assert.True(suite.T(), createResult.Success)
	
	// Extraer ID del producto creado
	productInfo, ok := createResult.Data.(map[string]interface{})
	require.True(suite.T(), ok)
	productID, ok := productInfo["id"].(string)
	require.True(suite.T(), ok)
	
	// 3. Verificar que el producto se puede consultar
	suite.T().Log("Step 3: Verify product can be retrieved")
	resp = suite.makeRequest("pos", "GET", "/api/v1/productos/"+productID, nil, posToken)
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	resp.Body.Close()
	
	// 4. Crear venta con el producto
	suite.T().Log("Step 4: Create sale")
	ventaData := map[string]interface{}{
		"tipo_documento": "boleta",
		"usuario_id":     "test-user-1",
		"sucursal_id":    "test-sucursal-1",
		"metodo_pago":    "efectivo",
		"detalles": []map[string]interface{}{
			{
				"producto_id":     productID,
				"cantidad":        3,
				"precio_unitario": 250.00,
			},
		},
	}
	
	resp = suite.makeRequest("pos", "POST", "/api/v1/ventas", ventaData, posToken)
	assert.Equal(suite.T(), http.StatusCreated, resp.StatusCode)
	
	var ventaResult models.APIResponse
	err = json.NewDecoder(resp.Body).Decode(&ventaResult)
	require.NoError(suite.T(), err)
	resp.Body.Close()
	
	assert.True(suite.T(), ventaResult.Success)
	
	// 5. Generar etiqueta del producto
	suite.T().Log("Step 5: Generate product label")
	labelData := map[string]interface{}{
		"template_id": "template-basic",
		"format":      "pdf",
		"productos": []map[string]interface{}{
			{
				"id":            productID,
				"codigo":        "E2E001",
				"nombre":        "Producto E2E Test",
				"precio":        250.00,
				"codigo_barras": "1234567890111",
			},
		},
	}
	
	resp = suite.makeRequest("labels", "POST", "/api/v1/labels/generate", labelData, "")
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	resp.Body.Close()
	
	// 6. Generar reporte de ventas
	suite.T().Log("Step 6: Generate sales report")
	resp = suite.makeRequest("reports", "GET", "/api/v1/reports/ventas?start_date=2024-01-01&end_date=2024-12-31", nil, "")
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	resp.Body.Close()
	
	// 7. Verificar dashboard
	suite.T().Log("Step 7: Check dashboard")
	resp = suite.makeRequest("reports", "GET", "/api/v1/reports/dashboard/resumen", nil, "")
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	resp.Body.Close()
	
	suite.T().Log("Complete business flow test completed successfully")
}

// TestSyncFlow test de flujo de sincronización
func (suite *E2ETestSuite) TestSyncFlow() {
	syncToken := suite.tokens["sync"]
	if syncToken == "" {
		suite.T().Skip("Sync token not available")
	}
	
	// 1. Verificar estado de sincronización
	suite.T().Log("Step 1: Check sync status")
	resp := suite.makeRequest("sync", "GET", "/api/v1/sync/status", nil, syncToken)
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	resp.Body.Close()
	
	// 2. Enviar heartbeat
	suite.T().Log("Step 2: Send heartbeat")
	heartbeatData := map[string]interface{}{
		"terminal_id": "TERM001",
		"timestamp":   time.Now().Unix(),
		"status":      "online",
	}
	
	resp = suite.makeRequest("sync", "POST", "/api/v1/sync/heartbeat", heartbeatData, syncToken)
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	resp.Body.Close()
	
	// 3. Sincronizar productos
	suite.T().Log("Step 3: Sync products")
	syncData := map[string]interface{}{
		"terminal_id": "TERM001",
		"timestamp":   time.Now().Unix(),
		"productos": []map[string]interface{}{
			{
				"id":           "sync-product-1",
				"codigo":       "SYNC001",
				"nombre":       "Producto Sincronizado",
				"precio":       100.00,
				"stock":        20,
				"last_updated": time.Now().Unix(),
				"action":       "create",
			},
		},
	}
	
	resp = suite.makeRequest("sync", "POST", "/api/v1/sync/productos", syncData, syncToken)
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	resp.Body.Close()
	
	// 4. Obtener cambios del servidor
	suite.T().Log("Step 4: Get server changes")
	resp = suite.makeRequest("sync", "GET", "/api/v1/sync/cambios", nil, syncToken)
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	resp.Body.Close()
	
	suite.T().Log("Sync flow test completed successfully")
}

// TestReportGeneration test de generación de reportes
func (suite *E2ETestSuite) TestReportGeneration() {
	// 1. Reporte de ventas
	suite.T().Log("Step 1: Generate sales report")
	resp := suite.makeRequest("reports", "GET", "/api/v1/reports/ventas", nil, "")
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	resp.Body.Close()
	
	// 2. Reporte de inventario
	suite.T().Log("Step 2: Generate inventory report")
	resp = suite.makeRequest("reports", "GET", "/api/v1/reports/inventario", nil, "")
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	resp.Body.Close()
	
	// 3. Productos con stock bajo
	suite.T().Log("Step 3: Generate low stock report")
	resp = suite.makeRequest("reports", "GET", "/api/v1/reports/inventario/stock-bajo", nil, "")
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	resp.Body.Close()
	
	// 4. Dashboard resumen
	suite.T().Log("Step 4: Generate dashboard summary")
	resp = suite.makeRequest("reports", "GET", "/api/v1/reports/dashboard/resumen", nil, "")
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	resp.Body.Close()
	
	// 5. Productos top
	suite.T().Log("Step 5: Generate top products report")
	resp = suite.makeRequest("reports", "GET", "/api/v1/reports/analytics/productos-top", nil, "")
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	resp.Body.Close()
	
	// 6. Exportar reporte
	suite.T().Log("Step 6: Export report")
	exportData := map[string]interface{}{
		"tipo_reporte": "ventas",
		"formato":      "pdf",
		"parametros": map[string]interface{}{
			"start_date": "2024-01-01",
			"end_date":   "2024-01-31",
		},
	}
	
	resp = suite.makeRequest("reports", "POST", "/api/v1/reports/export", exportData, "")
	assert.Equal(suite.T(), http.StatusAccepted, resp.StatusCode)
	
	var exportResult models.APIResponse
	err := json.NewDecoder(resp.Body).Decode(&exportResult)
	require.NoError(suite.T(), err)
	resp.Body.Close()
	
	if exportResult.Success {
		if data, ok := exportResult.Data.(map[string]interface{}); ok {
			if exportID, ok := data["export_id"].(string); ok {
				// Verificar estado de exportación
				suite.T().Log("Step 7: Check export status")
				resp = suite.makeRequest("reports", "GET", "/api/v1/reports/export/"+exportID+"/status", nil, "")
				assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
				resp.Body.Close()
			}
		}
	}
	
	suite.T().Log("Report generation test completed successfully")
}

// TestLabelGeneration test de generación de etiquetas
func (suite *E2ETestSuite) TestLabelGeneration() {
	// 1. Obtener plantillas disponibles
	suite.T().Log("Step 1: Get available templates")
	resp := suite.makeRequest("labels", "GET", "/api/v1/labels/templates", nil, "")
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	resp.Body.Close()
	
	// 2. Obtener formatos de código de barras
	suite.T().Log("Step 2: Get barcode formats")
	resp = suite.makeRequest("labels", "GET", "/api/v1/labels/barcode/formats", nil, "")
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	resp.Body.Close()
	
	// 3. Validar código de barras
	suite.T().Log("Step 3: Validate barcode")
	validateData := map[string]interface{}{
		"data":   "1234567890123",
		"format": "EAN13",
	}
	
	resp = suite.makeRequest("labels", "POST", "/api/v1/labels/barcode/validate", validateData, "")
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	resp.Body.Close()
	
	// 4. Generar código de barras
	suite.T().Log("Step 4: Generate barcode")
	barcodeData := map[string]interface{}{
		"data":   "1234567890123",
		"format": "CODE128",
		"width":  200,
		"height": 50,
		"output": "png",
	}
	
	resp = suite.makeRequest("labels", "POST", "/api/v1/labels/barcode/generate", barcodeData, "")
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	resp.Body.Close()
	
	// 5. Previsualizar etiqueta
	suite.T().Log("Step 5: Preview label")
	previewData := map[string]interface{}{
		"template_id": "template-basic",
		"producto": map[string]interface{}{
			"id":            "preview-product",
			"codigo":        "PREV001",
			"nombre":        "Producto Preview",
			"precio":        150.00,
			"codigo_barras": "1234567890123",
		},
	}
	
	resp = suite.makeRequest("labels", "POST", "/api/v1/labels/preview", previewData, "")
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	resp.Body.Close()
	
	// 6. Generar etiquetas
	suite.T().Log("Step 6: Generate labels")
	labelData := map[string]interface{}{
		"template_id": "template-basic",
		"format":      "pdf",
		"productos": []map[string]interface{}{
			{
				"id":            "label-product-1",
				"codigo":        "LABEL001",
				"nombre":        "Producto Etiqueta",
				"precio":        100.00,
				"codigo_barras": "1234567890123",
			},
		},
	}
	
	resp = suite.makeRequest("labels", "POST", "/api/v1/labels/generate", labelData, "")
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	resp.Body.Close()
	
	suite.T().Log("Label generation test completed successfully")
}

// TestErrorScenarios test de escenarios de error
func (suite *E2ETestSuite) TestErrorScenarios() {
	// 1. Endpoint no existente
	suite.T().Log("Step 1: Test non-existent endpoint")
	resp := suite.makeRequest("pos", "GET", "/api/v1/nonexistent", nil, "")
	assert.Equal(suite.T(), http.StatusNotFound, resp.StatusCode)
	resp.Body.Close()
	
	// 2. Request sin autenticación
	suite.T().Log("Step 2: Test unauthenticated request")
	resp = suite.makeRequest("pos", "GET", "/api/v1/productos", nil, "")
	assert.Equal(suite.T(), http.StatusUnauthorized, resp.StatusCode)
	resp.Body.Close()
	
	// 3. Request con token inválido
	suite.T().Log("Step 3: Test invalid token")
	resp = suite.makeRequest("pos", "GET", "/api/v1/productos", nil, "invalid-token")
	assert.Equal(suite.T(), http.StatusUnauthorized, resp.StatusCode)
	resp.Body.Close()
	
	// 4. Request con datos inválidos
	suite.T().Log("Step 4: Test invalid data")
	invalidData := map[string]interface{}{
		"nombre": "", // Nombre vacío
		"precio": -10, // Precio negativo
	}
	
	resp = suite.makeRequest("pos", "POST", "/api/v1/productos", invalidData, suite.tokens["pos"])
	assert.Equal(suite.T(), http.StatusBadRequest, resp.StatusCode)
	resp.Body.Close()
	
	suite.T().Log("Error scenarios test completed successfully")
}

// TestPerformance test de rendimiento básico
func (suite *E2ETestSuite) TestPerformance() {
	posToken := suite.tokens["pos"]
	if posToken == "" {
		suite.T().Skip("POS token not available")
	}
	
	// Test de rendimiento para listado de productos
	suite.T().Log("Testing product listing performance")
	start := time.Now()
	resp := suite.makeRequest("pos", "GET", "/api/v1/productos?limit=100", nil, posToken)
	duration := time.Since(start)
	
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	assert.Less(suite.T(), duration, 5*time.Second, "Product listing should complete within 5 seconds")
	resp.Body.Close()
	
	suite.T().Logf("Product listing took: %v", duration)
	
	// Test de rendimiento para dashboard
	suite.T().Log("Testing dashboard performance")
	start = time.Now()
	resp = suite.makeRequest("reports", "GET", "/api/v1/reports/dashboard/resumen", nil, "")
	duration = time.Since(start)
	
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	assert.Less(suite.T(), duration, 10*time.Second, "Dashboard should load within 10 seconds")
	resp.Body.Close()
	
	suite.T().Logf("Dashboard loading took: %v", duration)
}

// Función auxiliar para obtener variables de entorno con valor por defecto
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// TestE2ESuite ejecuta la suite de tests end-to-end
func TestE2ESuite(t *testing.T) {
	suite.Run(t, new(E2ETestSuite))
}

