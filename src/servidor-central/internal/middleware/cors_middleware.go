package middleware

import (
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// CORSConfig contiene la configuración de CORS
type CORSConfig struct {
	AllowOrigins     []string `yaml:"allow_origins" json:"allow_origins"`
	AllowMethods     []string `yaml:"allow_methods" json:"allow_methods"`
	AllowHeaders     []string `yaml:"allow_headers" json:"allow_headers"`
	ExposeHeaders    []string `yaml:"expose_headers" json:"expose_headers"`
	AllowCredentials bool     `yaml:"allow_credentials" json:"allow_credentials"`
	MaxAge           int      `yaml:"max_age" json:"max_age"`
}

// DefaultCORSConfig retorna una configuración CORS por defecto
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Requested-With",
			"X-API-Key",
			"X-Terminal-ID",
			"X-Sucursal-ID",
			"X-Request-ID",
		},
		ExposeHeaders: []string{
			"Content-Length",
			"Content-Type",
			"X-Request-ID",
			"X-Response-Time",
		},
		AllowCredentials: true,
		MaxAge:           86400, // 24 horas
	}
}

// CORS middleware que maneja Cross-Origin Resource Sharing
func CORS(config CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		// Verificar si el origen está permitido
		if isOriginAllowed(origin, config.AllowOrigins) {
			c.Header("Access-Control-Allow-Origin", origin)
		} else if len(config.AllowOrigins) == 1 && config.AllowOrigins[0] == "*" {
			c.Header("Access-Control-Allow-Origin", "*")
		}

		// Configurar headers CORS
		if len(config.AllowMethods) > 0 {
			c.Header("Access-Control-Allow-Methods", strings.Join(config.AllowMethods, ", "))
		}

		if len(config.AllowHeaders) > 0 {
			c.Header("Access-Control-Allow-Headers", strings.Join(config.AllowHeaders, ", "))
		}

		if len(config.ExposeHeaders) > 0 {
			c.Header("Access-Control-Expose-Headers", strings.Join(config.ExposeHeaders, ", "))
		}

		if config.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		if config.MaxAge > 0 {
			c.Header("Access-Control-Max-Age", string(rune(config.MaxAge)))
		}

		// Manejar preflight requests
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// isOriginAllowed verifica si un origen está permitido
func isOriginAllowed(origin string, allowedOrigins []string) bool {
	for _, allowed := range allowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
		
		// Verificar wildcards
		if strings.Contains(allowed, "*") {
			if matchWildcard(allowed, origin) {
				return true
			}
		}
	}
	return false
}

// matchWildcard verifica si un patrón con wildcard coincide con un origen
func matchWildcard(pattern, origin string) bool {
	// Implementación simple de wildcard matching
	if pattern == "*" {
		return true
	}
	
	if strings.HasPrefix(pattern, "*.") {
		domain := pattern[2:]
		return strings.HasSuffix(origin, domain)
	}
	
	return pattern == origin
}

// CORSForAPI retorna configuración CORS específica para cada API
func CORSForAPI(apiName string) CORSConfig {
	baseConfig := DefaultCORSConfig()
	
	switch apiName {
	case "pos":
		// API POS requiere configuración más estricta
		baseConfig.AllowOrigins = []string{
			"http://localhost:3000",
			"https://pos.ferreteria.com",
			"https://*.ferreteria.com",
		}
		baseConfig.AllowCredentials = true
		
	case "sync":
		// API Sync permite orígenes internos
		baseConfig.AllowOrigins = []string{
			"http://localhost:*",
			"https://sync.ferreteria.com",
			"https://admin.ferreteria.com",
		}
		
	case "labels":
		// API Labels permite acceso desde herramientas de diseño
		baseConfig.AllowOrigins = []string{
			"http://localhost:*",
			"https://labels.ferreteria.com",
			"https://design.ferreteria.com",
		}
		
	case "reports":
		// API Reports permite acceso desde dashboards
		baseConfig.AllowOrigins = []string{
			"http://localhost:*",
			"https://reports.ferreteria.com",
			"https://dashboard.ferreteria.com",
			"https://analytics.ferreteria.com",
		}
	}
	
	return baseConfig
}

// SecurityHeaders middleware que agrega headers de seguridad
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevenir ataques XSS
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		
		// Política de seguridad de contenido básica
		c.Header("Content-Security-Policy", "default-src 'self'")
		
		// Prevenir información de servidor
		c.Header("Server", "Ferre-POS-Server")
		
		// Política de referrer
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		
		// Prevenir MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")
		
		c.Next()
	}
}

// APIKeyAuth middleware para autenticación por API Key
func APIKeyAuth(validKeys []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "API Key requerida",
				"code":    "API_KEY_REQUIRED",
				"message": "Debe proporcionar una API Key válida",
			})
			c.Abort()
			return
		}

		// Verificar si la API Key es válida
		isValid := false
		for _, validKey := range validKeys {
			if apiKey == validKey {
				isValid = true
				break
			}
		}

		if !isValid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "API Key inválida",
				"code":    "API_KEY_INVALID",
				"message": "La API Key proporcionada no es válida",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequestID middleware que agrega un ID único a cada request
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			// Generar nuevo ID si no se proporciona
			requestID = generateRequestID()
		}

		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)
		c.Next()
	}
}

// generateRequestID genera un ID único para el request
func generateRequestID() string {
	// Implementación simple usando timestamp y random
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), rand.Intn(10000))
}

// ContentType middleware que valida el Content-Type
func ContentType(allowedTypes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodGet || c.Request.Method == http.MethodDelete {
			c.Next()
			return
		}

		contentType := c.GetHeader("Content-Type")
		if contentType == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Content-Type requerido",
				"code":    "CONTENT_TYPE_REQUIRED",
				"message": "Debe especificar el Content-Type",
			})
			c.Abort()
			return
		}

		// Verificar si el Content-Type está permitido
		isAllowed := false
		for _, allowedType := range allowedTypes {
			if strings.HasPrefix(contentType, allowedType) {
				isAllowed = true
				break
			}
		}

		if !isAllowed {
			c.JSON(http.StatusUnsupportedMediaType, gin.H{
				"error":   "Content-Type no soportado",
				"code":    "CONTENT_TYPE_UNSUPPORTED",
				"message": "El Content-Type especificado no está soportado",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// JSONOnly middleware que solo permite requests JSON
func JSONOnly() gin.HandlerFunc {
	return ContentType("application/json")
}

// UserAgent middleware que valida el User-Agent
func UserAgent(requiredAgents ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userAgent := c.GetHeader("User-Agent")
		if userAgent == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "User-Agent requerido",
				"code":    "USER_AGENT_REQUIRED",
				"message": "Debe especificar un User-Agent válido",
			})
			c.Abort()
			return
		}

		// Si no se especifican agentes requeridos, permitir cualquiera
		if len(requiredAgents) == 0 {
			c.Next()
			return
		}

		// Verificar si el User-Agent está permitido
		isAllowed := false
		for _, requiredAgent := range requiredAgents {
			if strings.Contains(userAgent, requiredAgent) {
				isAllowed = true
				break
			}
		}

		if !isAllowed {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "User-Agent no permitido",
				"code":    "USER_AGENT_FORBIDDEN",
				"message": "El User-Agent especificado no está permitido",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// IPWhitelist middleware que solo permite IPs específicas
func IPWhitelist(allowedIPs []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		
		// Verificar si la IP está en la whitelist
		isAllowed := false
		for _, allowedIP := range allowedIPs {
			if allowedIP == clientIP || allowedIP == "0.0.0.0" {
				isAllowed = true
				break
			}
		}

		if !isAllowed {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "IP no permitida",
				"code":    "IP_FORBIDDEN",
				"message": "Su dirección IP no tiene permisos para acceder a este recurso",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// Maintenance middleware que muestra mensaje de mantenimiento
func Maintenance(enabled bool, message string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if enabled {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":   "Servicio en mantenimiento",
				"code":    "SERVICE_MAINTENANCE",
				"message": message,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

