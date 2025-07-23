package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"ferre-pos-servidor-central/internal/models"
)

// BaseController contiene funcionalidades comunes para todos los controladores
type BaseController struct{}

// ResponseSuccess envía una respuesta exitosa
func (bc *BaseController) ResponseSuccess(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    data,
		Message: "Operación exitosa",
	})
}

// ResponseCreated envía una respuesta de creación exitosa
func (bc *BaseController) ResponseCreated(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Data:    data,
		Message: "Recurso creado exitosamente",
	})
}

// ResponseError envía una respuesta de error
func (bc *BaseController) ResponseError(c *gin.Context, statusCode int, message string, err error) {
	errorResponse := models.ErrorResponse{
		Code:    statusCode,
		Message: message,
	}
	
	if err != nil {
		errorResponse.Details = err.Error()
	}
	
	c.JSON(statusCode, models.APIResponse{
		Success: false,
		Error:   &errorResponse,
		Message: message,
	})
}

// ResponseValidationError envía una respuesta de error de validación
func (bc *BaseController) ResponseValidationError(c *gin.Context, errors []models.ValidationError) {
	c.JSON(http.StatusBadRequest, models.APIResponse{
		Success:          false,
		ValidationErrors: errors,
		Message:          "Error de validación",
	})
}

// ResponsePaginated envía una respuesta paginada
func (bc *BaseController) ResponsePaginated(c *gin.Context, data interface{}, pagination models.PaginationResponse) {
	c.JSON(http.StatusOK, models.PaginatedResponse{
		Success:    true,
		Data:       data,
		Pagination: pagination,
		Message:    "Operación exitosa",
	})
}

// ParseUUID parsea un UUID desde un parámetro de ruta
func (bc *BaseController) ParseUUID(c *gin.Context, param string) (uuid.UUID, error) {
	idStr := c.Param(param)
	if idStr == "" {
		return uuid.Nil, &models.ValidationError{
			Field:   param,
			Message: "ID requerido",
		}
	}
	
	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, &models.ValidationError{
			Field:   param,
			Message: "ID inválido",
		}
	}
	
	return id, nil
}

// ParsePagination parsea parámetros de paginación
func (bc *BaseController) ParsePagination(c *gin.Context) models.PaginationFilter {
	page := 1
	limit := 20
	
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	
	return models.PaginationFilter{
		Page:  page,
		Limit: limit,
	}
}

// ParseSort parsea parámetros de ordenamiento
func (bc *BaseController) ParseSort(c *gin.Context) models.SortFilter {
	sortBy := c.Query("sort_by")
	sortOrder := c.Query("sort_order")
	
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}
	
	return models.SortFilter{
		SortBy:    sortBy,
		SortOrder: sortOrder,
	}
}

// ParseDateRange parsea parámetros de rango de fechas
func (bc *BaseController) ParseDateRange(c *gin.Context) models.DateRangeFilter {
	var filter models.DateRangeFilter
	
	if fechaInicio := c.Query("fecha_inicio"); fechaInicio != "" {
		filter.FechaInicio = &fechaInicio
	}
	
	if fechaFin := c.Query("fecha_fin"); fechaFin != "" {
		filter.FechaFin = &fechaFin
	}
	
	return filter
}

// ValidateJSON valida y parsea JSON del request body
func (bc *BaseController) ValidateJSON(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindJSON(obj); err != nil {
		return err
	}
	
	// Aquí se pueden agregar validaciones adicionales usando validator
	return nil
}

// GetUserFromContext obtiene el usuario autenticado del contexto
func (bc *BaseController) GetUserFromContext(c *gin.Context) (*models.Usuario, error) {
	userInterface, exists := c.Get("user")
	if !exists {
		return nil, &models.ValidationError{
			Field:   "authorization",
			Message: "Usuario no autenticado",
		}
	}
	
	user, ok := userInterface.(*models.Usuario)
	if !ok {
		return nil, &models.ValidationError{
			Field:   "authorization",
			Message: "Contexto de usuario inválido",
		}
	}
	
	return user, nil
}

// GetSucursalFromContext obtiene la sucursal del contexto (si aplica)
func (bc *BaseController) GetSucursalFromContext(c *gin.Context) (*uuid.UUID, error) {
	sucursalInterface, exists := c.Get("sucursal_id")
	if !exists {
		return nil, nil // Sucursal opcional
	}
	
	sucursalID, ok := sucursalInterface.(uuid.UUID)
	if !ok {
		return nil, &models.ValidationError{
			Field:   "sucursal",
			Message: "Contexto de sucursal inválido",
		}
	}
	
	return &sucursalID, nil
}

// CheckPermission verifica si el usuario tiene un permiso específico
func (bc *BaseController) CheckPermission(user *models.Usuario, permission string) bool {
	if user == nil {
		return false
	}
	
	// Verificar permisos en la configuración del usuario
	if user.ConfiguracionUsuario != nil {
		for _, perm := range user.ConfiguracionUsuario.PermisosPersonalizados {
			if perm == permission {
				return true
			}
		}
	}
	
	// Verificar permisos por rol (implementación simplificada)
	switch user.Rol {
	case models.RolAdministrador:
		return true // Admin tiene todos los permisos
	case models.RolGerente:
		// Gerente tiene la mayoría de permisos
		return permission != "admin_only"
	case models.RolVendedor:
		// Vendedor tiene permisos limitados
		return permission == "ventas" || permission == "productos" || permission == "clientes"
	case models.RolCajero:
		// Cajero tiene permisos muy limitados
		return permission == "ventas" || permission == "caja"
	default:
		return false
	}
}

// RequirePermission middleware para verificar permisos
func (bc *BaseController) RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := bc.GetUserFromContext(c)
		if err != nil {
			bc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
			c.Abort()
			return
		}
		
		if !bc.CheckPermission(user, permission) {
			bc.ResponseError(c, http.StatusForbidden, "Permisos insuficientes", nil)
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// RequireRole middleware para verificar rol específico
func (bc *BaseController) RequireRole(roles ...models.RolUsuario) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := bc.GetUserFromContext(c)
		if err != nil {
			bc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
			c.Abort()
			return
		}
		
		hasRole := false
		for _, role := range roles {
			if user.Rol == role {
				hasRole = true
				break
			}
		}
		
		if !hasRole {
			bc.ResponseError(c, http.StatusForbidden, "Rol insuficiente", nil)
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// LogActivity registra actividad del usuario
func (bc *BaseController) LogActivity(c *gin.Context, action string, details interface{}) {
	user, err := bc.GetUserFromContext(c)
	if err != nil {
		return // No logear si no hay usuario
	}
	
	// Aquí se implementaría el logging de actividades
	// Por ahora solo un log simple
	c.Header("X-User-Activity", action)
	c.Header("X-User-ID", user.ID.String())
}

// HandlePanic maneja panics en los controladores
func (bc *BaseController) HandlePanic() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		bc.ResponseError(c, http.StatusInternalServerError, "Error interno del servidor", nil)
		c.Abort()
	})
}

// SetCacheHeaders establece headers de cache
func (bc *BaseController) SetCacheHeaders(c *gin.Context, maxAge int) {
	if maxAge > 0 {
		c.Header("Cache-Control", "public, max-age="+strconv.Itoa(maxAge))
	} else {
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")
	}
}

// SetCORSHeaders establece headers CORS
func (bc *BaseController) SetCORSHeaders(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Requested-With")
	c.Header("Access-Control-Expose-Headers", "Content-Length, Content-Type")
	c.Header("Access-Control-Max-Age", "86400")
}

// HealthCheck endpoint de salud
func (bc *BaseController) HealthCheck(c *gin.Context) {
	bc.ResponseSuccess(c, gin.H{
		"status":    "healthy",
		"timestamp": models.GetCurrentTime(),
		"version":   "1.0.0",
	})
}

// NotFound maneja rutas no encontradas
func (bc *BaseController) NotFound(c *gin.Context) {
	bc.ResponseError(c, http.StatusNotFound, "Endpoint no encontrado", nil)
}

// MethodNotAllowed maneja métodos no permitidos
func (bc *BaseController) MethodNotAllowed(c *gin.Context) {
	bc.ResponseError(c, http.StatusMethodNotAllowed, "Método no permitido", nil)
}

// ValidateContentType valida el content-type del request
func (bc *BaseController) ValidateContentType(expectedType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
			contentType := c.GetHeader("Content-Type")
			if contentType != expectedType {
				bc.ResponseError(c, http.StatusUnsupportedMediaType, "Content-Type no soportado", nil)
				c.Abort()
				return
			}
		}
		c.Next()
	}
}

// RateLimitByUser implementa rate limiting por usuario
func (bc *BaseController) RateLimitByUser(requestsPerMinute int) gin.HandlerFunc {
	// Implementación simplificada - en producción usar Redis
	return func(c *gin.Context) {
		user, err := bc.GetUserFromContext(c)
		if err != nil {
			c.Next()
			return
		}
		
		// Aquí se implementaría el rate limiting real
		// Por ahora solo agregar header informativo
		c.Header("X-RateLimit-Limit", strconv.Itoa(requestsPerMinute))
		c.Header("X-RateLimit-User", user.ID.String())
		c.Next()
	}
}

// ValidateAPIKey valida API key para acceso externo
func (bc *BaseController) ValidateAPIKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			apiKey = c.Query("api_key")
		}
		
		if apiKey == "" {
			bc.ResponseError(c, http.StatusUnauthorized, "API Key requerida", nil)
			c.Abort()
			return
		}
		
		// Aquí se validaría la API key contra la base de datos
		// Por ahora implementación simplificada
		if apiKey != "demo-api-key" {
			bc.ResponseError(c, http.StatusUnauthorized, "API Key inválida", nil)
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// AddSecurityHeaders agrega headers de seguridad
func (bc *BaseController) AddSecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	}
}

// RequestLogger middleware para logging de requests
func (bc *BaseController) RequestLogger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format("02/Jan/2006:15:04:05 -0700"),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})
}

// ValidateJSONSchema valida JSON contra un esquema
func (bc *BaseController) ValidateJSONSchema(schema interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Implementación de validación de esquema JSON
		// Por ahora solo continuar
		c.Next()
	}
}

// CompressResponse middleware para compresión de respuestas
func (bc *BaseController) CompressResponse() gin.HandlerFunc {
	// En implementación real usar gzip middleware
	return func(c *gin.Context) {
		c.Header("Content-Encoding", "identity")
		c.Next()
	}
}

// Timeout middleware para timeout de requests
func (bc *BaseController) Timeout(duration time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Implementación de timeout
		c.Next()
	}
}

// Metrics middleware para métricas
func (bc *BaseController) Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)
		
		// Agregar métricas como headers
		c.Header("X-Response-Time", duration.String())
		c.Header("X-Request-ID", uuid.New().String())
	}
}

