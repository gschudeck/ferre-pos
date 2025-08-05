package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"ferre_pos_apis/internal/config"
	"ferre_pos_apis/internal/logger"
)

// RequestID middleware que agrega un ID único a cada request
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		
		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)
		c.Next()
	}
}

// Logger middleware de logging personalizado
func Logger(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		method := c.Request.Method
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()
		requestID, _ := c.Get("request_id")

		// Procesar request
		c.Next()

		// Calcular latencia
		latency := time.Since(start)
		statusCode := c.Writer.Status()

		// Construir path completo
		if raw != "" {
			path = path + "?" + raw
		}

		// Log de la request
		logger.LogRequest(
			log,
			method,
			path,
			userAgent,
			clientIP,
			statusCode,
			latency,
			requestID.(string),
		)
	}
}

// CORS middleware de CORS personalizado
func CORS(cfg *config.CORSConfig) gin.HandlerFunc {
	if !cfg.Enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	corsConfig := cors.Config{
		AllowOrigins:     cfg.AllowedOrigins,
		AllowMethods:     cfg.AllowedMethods,
		AllowHeaders:     cfg.AllowedHeaders,
		ExposeHeaders:    cfg.ExposedHeaders,
		AllowCredentials: cfg.AllowCredentials,
		MaxAge:           time.Duration(cfg.MaxAge) * time.Second,
	}

	// Si se permite cualquier origen
	if len(cfg.AllowedOrigins) == 1 && cfg.AllowedOrigins[0] == "*" {
		corsConfig.AllowAllOrigins = true
	}

	return cors.New(corsConfig)
}

// Auth middleware de autenticación JWT
func Auth(cfg *config.APIConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "MISSING_TOKEN",
					"message": "Token de autorización requerido",
				},
			})
			c.Abort()
			return
		}

		// Extraer token del header
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "INVALID_TOKEN_FORMAT",
					"message": "Formato de token inválido. Use: Bearer <token>",
				},
			})
			c.Abort()
			return
		}

		tokenString := tokenParts[1]

		// Validar token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Verificar método de firma
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(cfg.Auth.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "INVALID_TOKEN",
					"message": "Token inválido o expirado",
				},
			})
			c.Abort()
			return
		}

		// Extraer claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "INVALID_CLAIMS",
					"message": "Claims del token inválidos",
				},
			})
			c.Abort()
			return
		}

		// Verificar expiración
		if exp, ok := claims["exp"].(float64); ok {
			if time.Now().Unix() > int64(exp) {
				c.JSON(http.StatusUnauthorized, gin.H{
					"success": false,
					"error": gin.H{
						"code":    "TOKEN_EXPIRED",
						"message": "Token expirado",
					},
				})
				c.Abort()
				return
			}
		}

		// Extraer información del usuario
		userID, _ := claims["user_id"].(string)
		userRUT, _ := claims["rut"].(string)
		userRole, _ := claims["role"].(string)
		sucursalID, _ := claims["sucursal_id"].(string)
		terminalID, _ := claims["terminal_id"].(string)

		// Agregar información al contexto
		c.Set("user_id", userID)
		c.Set("user_rut", userRUT)
		c.Set("user_role", userRole)
		c.Set("sucursal_id", sucursalID)
		c.Set("terminal_id", terminalID)
		c.Set("jwt_claims", claims)

		c.Next()
	}
}

// RequireRole middleware que requiere roles específicos
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "ROLE_NOT_FOUND",
					"message": "Rol de usuario no encontrado",
				},
			})
			c.Abort()
			return
		}

		userRoleStr, ok := userRole.(string)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "INVALID_ROLE",
					"message": "Rol de usuario inválido",
				},
			})
			c.Abort()
			return
		}

		// Verificar si el rol está permitido
		roleAllowed := false
		for _, allowedRole := range allowedRoles {
			if userRoleStr == allowedRole {
				roleAllowed = true
				break
			}
		}

		if !roleAllowed {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "INSUFFICIENT_PERMISSIONS",
					"message": "Permisos insuficientes para esta operación",
					"details": gin.H{
						"required_roles": allowedRoles,
						"user_role":      userRoleStr,
					},
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireSucursal middleware que requiere que el usuario pertenezca a una sucursal específica
func RequireSucursal(sucursalID string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userSucursalID, exists := c.Get("sucursal_id")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "SUCURSAL_NOT_FOUND",
					"message": "Sucursal de usuario no encontrada",
				},
			})
			c.Abort()
			return
		}

		userSucursalIDStr, ok := userSucursalID.(string)
		if !ok || userSucursalIDStr != sucursalID {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "SUCURSAL_MISMATCH",
					"message": "Usuario no pertenece a la sucursal requerida",
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// ValidateContentType middleware que valida el Content-Type
func ValidateContentType(allowedTypes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "GET" || c.Request.Method == "DELETE" {
			c.Next()
			return
		}

		contentType := c.GetHeader("Content-Type")
		if contentType == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "MISSING_CONTENT_TYPE",
					"message": "Content-Type header requerido",
				},
			})
			c.Abort()
			return
		}

		// Verificar si el Content-Type está permitido
		typeAllowed := false
		for _, allowedType := range allowedTypes {
			if strings.Contains(contentType, allowedType) {
				typeAllowed = true
				break
			}
		}

		if !typeAllowed {
			c.JSON(http.StatusUnsupportedMediaType, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "UNSUPPORTED_MEDIA_TYPE",
					"message": "Content-Type no soportado",
					"details": gin.H{
						"allowed_types":  allowedTypes,
						"received_type":  contentType,
					},
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// Timeout middleware que establece timeout para requests
func Timeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Crear contexto con timeout
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		// Reemplazar contexto de la request
		c.Request = c.Request.WithContext(ctx)

		// Crear canal para manejar completion
		done := make(chan struct{})
		go func() {
			defer close(done)
			c.Next()
		}()

		select {
		case <-done:
			// Request completada normalmente
			return
		case <-ctx.Done():
			// Timeout alcanzado
			c.JSON(http.StatusRequestTimeout, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "REQUEST_TIMEOUT",
					"message": "Request timeout alcanzado",
					"details": gin.H{
						"timeout_seconds": timeout.Seconds(),
					},
				},
			})
			c.Abort()
			return
		}
	}
}

// ErrorHandler middleware de manejo de errores
func ErrorHandler(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Verificar si hay errores
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			
			requestID, _ := c.Get("request_id")
			userID, _ := c.Get("user_id")
			
			// Log del error
			log.WithField("request_id", requestID).
				WithField("user_id", userID).
				WithField("path", c.Request.URL.Path).
				WithField("method", c.Request.Method).
				WithError(err).
				Error("Request error")

			// Si no se ha enviado respuesta aún
			if !c.Writer.Written() {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error": gin.H{
						"code":       "INTERNAL_ERROR",
						"message":    "Error interno del servidor",
						"request_id": requestID,
					},
				})
			}
		}
	}
}

// SecurityHeaders middleware que agrega headers de seguridad
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Next()
	}
}

// TerminalAuth middleware de autenticación para terminales
func TerminalAuth(cfg *config.APIConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "MISSING_TOKEN",
					"message": "Token de autorización requerido",
				},
			})
			c.Abort()
			return
		}

		// Extraer token del header
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "INVALID_TOKEN_FORMAT",
					"message": "Formato de token inválido. Use: Bearer <token>",
				},
			})
			c.Abort()
			return
		}

		tokenString := tokenParts[1]

		// Validar token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Verificar método de firma
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(cfg.Auth.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "INVALID_TOKEN",
					"message": "Token inválido o expirado",
				},
			})
			c.Abort()
			return
		}

		// Extraer claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "INVALID_CLAIMS",
					"message": "Claims del token inválidos",
				},
			})
			c.Abort()
			return
		}

		// Verificar que es un token de terminal
		tokenType, _ := claims["type"].(string)
		if tokenType != "terminal" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "INVALID_TOKEN_TYPE",
					"message": "Token no es de tipo terminal",
				},
			})
			c.Abort()
			return
		}

		// Verificar expiración
		if exp, ok := claims["exp"].(float64); ok {
			if time.Now().Unix() > int64(exp) {
				c.JSON(http.StatusUnauthorized, gin.H{
					"success": false,
					"error": gin.H{
						"code":    "TOKEN_EXPIRED",
						"message": "Token expirado",
					},
				})
				c.Abort()
				return
			}
		}

		// Extraer información del terminal
		terminalID, _ := claims["terminal_id"].(string)
		sucursalID, _ := claims["sucursal_id"].(string)
		terminalCode, _ := claims["terminal_code"].(string)

		// Agregar información al contexto
		c.Set("terminal_id", terminalID)
		c.Set("sucursal_id", sucursalID)
		c.Set("terminal_code", terminalCode)
		c.Set("jwt_claims", claims)

		c.Next()
	}
}

// RateLimitByUser middleware de rate limiting por usuario
func RateLimitByUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Este middleware se puede usar junto con el rate limiter principal
		// para implementar límites específicos por usuario
		userID, exists := c.Get("user_id")
		if exists {
			c.Set("rate_limit_key", fmt.Sprintf("user:%v", userID))
		}
		c.Next()
	}
}

// AuditLog middleware de auditoría
func AuditLog(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Solo auditar operaciones de escritura
		if c.Request.Method != "POST" && c.Request.Method != "PUT" && c.Request.Method != "DELETE" {
			c.Next()
			return
		}

		start := time.Now()
		
		// Capturar información antes del procesamiento
		userID, _ := c.Get("user_id")
		userRUT, _ := c.Get("user_rut")
		requestID, _ := c.Get("request_id")
		
		c.Next()

		// Log de auditoría después del procesamiento
		duration := time.Since(start)
		
		auditFields := map[string]interface{}{
			"audit_type":    "api_operation",
			"user_id":       userID,
			"user_rut":      userRUT,
			"request_id":    requestID,
			"method":        c.Request.Method,
			"path":          c.Request.URL.Path,
			"status_code":   c.Writer.Status(),
			"duration_ms":   duration.Milliseconds(),
			"client_ip":     c.ClientIP(),
			"user_agent":    c.Request.UserAgent(),
		}

		if c.Writer.Status() >= 400 {
			log.WithFields(auditFields).Warn("API operation completed with error")
		} else {
			log.WithFields(auditFields).Info("API operation completed successfully")
		}
	}
}

