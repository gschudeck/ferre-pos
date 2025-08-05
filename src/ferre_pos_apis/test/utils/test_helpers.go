package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"ferre_pos_apis/internal/config"
	"ferre_pos_apis/internal/database"
	"ferre_pos_apis/internal/models"
)

// TestServer estructura para servidor de test
type TestServer struct {
	Router   *gin.Engine
	Config   *config.Config
	Database *database.Database
}

// SetupTestServer configura un servidor de test
func SetupTestServer(t *testing.T) *TestServer {
	gin.SetMode(gin.TestMode)
	
	router := gin.New()
	router.Use(gin.Recovery())
	
	// Configuración de test
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:                  "localhost",
			Port:                  5432,
			Name:                  "test_db",
			User:                  "test",
			Password:              "test",
			SSLMode:               "disable",
			MaxOpenConnections:    10,
			MaxIdleConnections:    5,
			ConnectionMaxLifetime: 5 * time.Minute,
		},
		Security: config.SecurityConfig{
			CORS: config.CORSConfig{
				Enabled:        true,
				AllowedOrigins: []string{"*"},
			},
		},
	}
	
	// Crear router
	return &TestServer{
		Router: router,
		Config: cfg,
	}
}

// TeardownTestDatabase limpia la base de datos de test
func (ts *TestServer) TeardownTestDatabase(t *testing.T) {
	if ts.Database != nil {
		ts.Database.Close()
	}
}

// MakeRequest realiza una petición HTTP de test
func (ts *TestServer) MakeRequest(method, url string, body interface{}, headers map[string]string) *httptest.ResponseRecorder {
	var bodyReader *bytes.Buffer
	if body != nil {
		bodyBytes, _ := json.Marshal(body)
		bodyReader = bytes.NewBuffer(bodyBytes)
	} else {
		bodyReader = bytes.NewBuffer(nil)
	}
	
	req, _ := http.NewRequest(method, url, bodyReader)
	
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	
	recorder := httptest.NewRecorder()
	ts.Router.ServeHTTP(recorder, req)
	
	return recorder
}

// MakeAuthenticatedRequest realiza una petición HTTP autenticada
func (ts *TestServer) MakeAuthenticatedRequest(method, url string, body interface{}, token string) *httptest.ResponseRecorder {
	headers := map[string]string{
		"Authorization": "Bearer " + token,
	}
	return ts.MakeRequest(method, url, body, headers)
}

// ParseJSONFixture parsea un archivo JSON de fixture
func ParseJSONFixture(t *testing.T, filename string, target interface{}) {
	// En un entorno real, esto leería desde test/fixtures/
	// Por ahora, usamos datos mock
	testData := map[string]interface{}{
		"usuarios": []interface{}{
			map[string]interface{}{
				"id":            "test-user-1",
				"email":         "admin@test.com",
				"password_hash": "$2a$10$test",
				"rol":           "admin",
				"activo":        true,
				"sucursal_id":   "test-sucursal-1",
				"created_at":    time.Now(),
				"updated_at":    time.Now(),
			},
		},
		"productos": []interface{}{
			map[string]interface{}{
				"id":            "test-product-1",
				"codigo":        "TEST001",
				"codigo_barras": "1234567890123",
				"nombre":        "Producto Test",
				"precio":        100.00,
				"stock":         50,
				"activo":        true,
				"categoria_id":  "test-category-1",
				"created_at":    time.Now(),
				"updated_at":    time.Now(),
			},
		},
		"categorias": []interface{}{
			map[string]interface{}{
				"id":          "test-category-1",
				"nombre":      "Categoría Test",
				"descripcion": "Categoría para tests",
				"activa":      true,
				"created_at":  time.Now(),
				"updated_at":  time.Now(),
			},
		},
		"sucursales": []interface{}{
			map[string]interface{}{
				"id":         "test-sucursal-1",
				"nombre":     "Sucursal Test",
				"direccion":  "Dirección Test",
				"telefono":   "123456789",
				"email":      "sucursal@test.com",
				"activa":     true,
				"created_at": time.Now(),
				"updated_at": time.Now(),
			},
		},
	}
	
	// Convertir a JSON y deserializar en target
	jsonData, _ := json.Marshal(testData)
	json.Unmarshal(jsonData, target)
}

// CleanupTestData limpia datos de test
func CleanupTestData(t *testing.T, db *database.Database, tables ...string) {
	if db == nil {
		return
	}
	
	ctx := context.Background()
	for _, table := range tables {
		query := "DELETE FROM " + table + " WHERE id LIKE 'test-%'"
		db.ExecContext(ctx, query)
	}
}

// AssertErrorResponse verifica que una respuesta sea de error
func AssertErrorResponse(t *testing.T, recorder *httptest.ResponseRecorder, expectedStatus int, expectedErrorCode string) {
	assert.Equal(t, expectedStatus, recorder.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.False(t, response.Success)
	
	if response.Error != nil {
		assert.Equal(t, expectedErrorCode, response.Error.Code)
	}
}

// CreateTestUser crea usuario de test
func CreateTestUser(t *testing.T, db *database.Database) *models.Usuario {
	userID := uuid.New()
	user := &models.Usuario{
		ID:     userID,
		Email:  stringPtr("test@example.com"),
		Activo: true,
		Rol:    models.RolVendedor,
	}
	
	// Aquí normalmente insertarías en la base de datos
	// Por ahora retornamos el usuario mock
	return user
}

// CreateTestProduct crea producto de test
func CreateTestProduct(t *testing.T, db *database.Database) *models.Producto {
	productID := uuid.New()
	product := &models.Producto{
		ID:     productID,
		Activo: true,
	}
	
	return product
}

// CreateTestVenta crea venta de test
func CreateTestVenta(t *testing.T, db *database.Database) *models.Venta {
	ventaID := uuid.New()
	venta := &models.Venta{
		ID: ventaID,
	}
	
	return venta
}

// GenerateTestJWT genera un JWT de test
func GenerateTestJWT(userID, role string, config *config.Config) (string, error) {
	// En un entorno real, esto generaría un JWT válido
	// Por ahora retornamos un token mock
	return "test-jwt-token", nil
}

// stringPtr retorna un puntero a string
func stringPtr(s string) *string {
	return &s
}

