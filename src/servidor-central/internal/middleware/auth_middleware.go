package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"ferre-pos-servidor-central/internal/models"
	"ferre-pos-servidor-central/internal/services"
)

// AuthMiddleware contiene la configuración del middleware de autenticación
type AuthMiddleware struct {
	authService services.AuthService
	jwtSecret   string
	skipPaths   []string
}

// NewAuthMiddleware crea una nueva instancia del middleware de autenticación
func NewAuthMiddleware(authService services.AuthService, jwtSecret string) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
		jwtSecret:   jwtSecret,
		skipPaths: []string{
			"/api/pos/auth/login",
			"/api/pos/auth/refresh",
			"/api/pos/health",
			"/api/sync/health",
			"/api/labels/health",
			"/api/reports/health",
			"/metrics",
			"/docs",
			"/swagger",
		},
	}
}

// JWTClaims define las claims del JWT
type JWTClaims struct {
	UserID     uuid.UUID `json:"user_id"`
	Email      string    `json:"email"`
	Rol        string    `json:"rol"`
	SucursalID *uuid.UUID `json:"sucursal_id,omitempty"`
	TerminalID string    `json:"terminal_id,omitempty"`
	Permisos   []string  `json:"permisos"`
	jwt.RegisteredClaims
}

// RequireAuth middleware que requiere autenticación
func (am *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Verificar si la ruta debe ser omitida
		if am.shouldSkipPath(c.Request.URL.Path) {
			c.Next()
			return
		}

		// Obtener token del header
		token := am.extractToken(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Token de autenticación requerido",
				"code":    "AUTH_TOKEN_REQUIRED",
				"message": "Debe proporcionar un token de autenticación válido",
			})
			c.Abort()
			return
		}

		// Validar token
		claims, err := am.validateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Token inválido",
				"code":    "AUTH_TOKEN_INVALID",
				"message": err.Error(),
			})
			c.Abort()
			return
		}

		// Verificar expiración
		if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Token expirado",
				"code":    "AUTH_TOKEN_EXPIRED",
				"message": "El token de autenticación ha expirado",
			})
			c.Abort()
			return
		}

		// Obtener usuario completo del servicio
		user, err := am.authService.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Usuario no válido",
				"code":    "AUTH_USER_INVALID",
				"message": err.Error(),
			})
			c.Abort()
			return
		}

		// Verificar que el usuario esté activo
		if !user.Activo {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Usuario inactivo",
				"code":    "AUTH_USER_INACTIVE",
				"message": "La cuenta de usuario está desactivada",
			})
			c.Abort()
			return
		}

		// Agregar información del usuario al contexto
		c.Set("user", user)
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_rol", claims.Rol)
		c.Set("user_permisos", claims.Permisos)
		
		if claims.SucursalID != nil {
			c.Set("sucursal_id", *claims.SucursalID)
		}
		
		if claims.TerminalID != "" {
			c.Set("terminal_id", claims.TerminalID)
		}

		c.Next()
	}
}

// RequireRole middleware que requiere un rol específico
func (am *AuthMiddleware) RequireRole(roles ...models.RolUsuario) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Usuario no autenticado",
				"code":    "AUTH_USER_NOT_FOUND",
				"message": "No se encontró información del usuario en el contexto",
			})
			c.Abort()
			return
		}

		userObj := user.(*models.Usuario)
		
		// Verificar si el usuario tiene alguno de los roles requeridos
		hasRole := false
		for _, role := range roles {
			if userObj.Rol == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Rol insuficiente",
				"code":    "AUTH_INSUFFICIENT_ROLE",
				"message": "No tiene permisos suficientes para acceder a este recurso",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequirePermission middleware que requiere un permiso específico
func (am *AuthMiddleware) RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		permisos, exists := c.Get("user_permisos")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Permisos no encontrados",
				"code":    "AUTH_PERMISSIONS_NOT_FOUND",
				"message": "No se encontraron permisos en el contexto",
			})
			c.Abort()
			return
		}

		permisosSlice := permisos.([]string)
		
		// Verificar si el usuario tiene el permiso requerido
		hasPermission := false
		for _, p := range permisosSlice {
			if p == permission || p == "admin" { // admin tiene todos los permisos
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Permiso insuficiente",
				"code":    "AUTH_INSUFFICIENT_PERMISSION",
				"message": "No tiene el permiso requerido: " + permission,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireSucursal middleware que requiere que el usuario tenga una sucursal asignada
func (am *AuthMiddleware) RequireSucursal() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, exists := c.Get("sucursal_id")
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Sucursal requerida",
				"code":    "AUTH_SUCURSAL_REQUIRED",
				"message": "Esta operación requiere que el usuario tenga una sucursal asignada",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireTerminal middleware que requiere que el usuario tenga un terminal asignado
func (am *AuthMiddleware) RequireTerminal() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, exists := c.Get("terminal_id")
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Terminal requerido",
				"code":    "AUTH_TERMINAL_REQUIRED",
				"message": "Esta operación requiere que el usuario tenga un terminal asignado",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// OptionalAuth middleware que permite autenticación opcional
func (am *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := am.extractToken(c)
		if token == "" {
			c.Next()
			return
		}

		claims, err := am.validateToken(token)
		if err != nil {
			c.Next()
			return
		}

		// Verificar expiración
		if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
			c.Next()
			return
		}

		// Obtener usuario
		user, err := am.authService.ValidateToken(token)
		if err != nil || !user.Activo {
			c.Next()
			return
		}

		// Agregar información del usuario al contexto si es válido
		c.Set("user", user)
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_rol", claims.Rol)
		c.Set("user_permisos", claims.Permisos)
		
		if claims.SucursalID != nil {
			c.Set("sucursal_id", *claims.SucursalID)
		}
		
		if claims.TerminalID != "" {
			c.Set("terminal_id", claims.TerminalID)
		}

		c.Next()
	}
}

// extractToken extrae el token del header Authorization
func (am *AuthMiddleware) extractToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}

	// Verificar formato "Bearer <token>"
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}

	return parts[1]
}

// validateToken valida un token JWT y retorna las claims
func (am *AuthMiddleware) validateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verificar método de firma
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(am.jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrTokenInvalid
}

// shouldSkipPath verifica si una ruta debe omitir la autenticación
func (am *AuthMiddleware) shouldSkipPath(path string) bool {
	for _, skipPath := range am.skipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}

// GenerateToken genera un nuevo token JWT
func (am *AuthMiddleware) GenerateToken(user *models.Usuario, sucursalID *uuid.UUID, terminalID string) (string, error) {
	now := time.Now()
	claims := JWTClaims{
		UserID:     user.ID,
		Email:      user.Email,
		Rol:        string(user.Rol),
		SucursalID: sucursalID,
		TerminalID: terminalID,
		Permisos:   user.Permisos,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)), // 24 horas
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "ferre-pos-servidor-central",
			Subject:   user.ID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(am.jwtSecret))
}

// GenerateRefreshToken genera un token de refresh
func (am *AuthMiddleware) GenerateRefreshToken(user *models.Usuario) (string, error) {
	now := time.Now()
	claims := JWTClaims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(7 * 24 * time.Hour)), // 7 días
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "ferre-pos-servidor-central",
			Subject:   user.ID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(am.jwtSecret))
}

// RefreshToken refresca un token existente
func (am *AuthMiddleware) RefreshToken(refreshTokenString string) (string, error) {
	claims, err := am.validateToken(refreshTokenString)
	if err != nil {
		return "", err
	}

	// Obtener usuario actualizado
	user, err := am.authService.ValidateToken(refreshTokenString)
	if err != nil {
		return "", err
	}

	// Generar nuevo token
	return am.GenerateToken(user, claims.SucursalID, claims.TerminalID)
}

// RevokeToken revoca un token (implementación básica)
func (am *AuthMiddleware) RevokeToken(tokenString string) error {
	// En una implementación real, esto podría agregar el token a una blacklist
	// Por ahora, simplemente validamos que el token sea válido
	_, err := am.validateToken(tokenString)
	return err
}

// GetUserFromContext obtiene el usuario del contexto de Gin
func GetUserFromContext(c *gin.Context) (*models.Usuario, error) {
	user, exists := c.Get("user")
	if !exists {
		return nil, fmt.Errorf("usuario no encontrado en el contexto")
	}

	userObj, ok := user.(*models.Usuario)
	if !ok {
		return nil, fmt.Errorf("tipo de usuario inválido en el contexto")
	}

	return userObj, nil
}

// GetUserIDFromContext obtiene el ID del usuario del contexto
func GetUserIDFromContext(c *gin.Context) (uuid.UUID, error) {
	userID, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, fmt.Errorf("ID de usuario no encontrado en el contexto")
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		return uuid.Nil, fmt.Errorf("tipo de ID de usuario inválido en el contexto")
	}

	return userUUID, nil
}

// GetSucursalIDFromContext obtiene el ID de la sucursal del contexto
func GetSucursalIDFromContext(c *gin.Context) (*uuid.UUID, error) {
	sucursalID, exists := c.Get("sucursal_id")
	if !exists {
		return nil, nil // No es error, puede ser opcional
	}

	sucursalUUID, ok := sucursalID.(uuid.UUID)
	if !ok {
		return nil, fmt.Errorf("tipo de ID de sucursal inválido en el contexto")
	}

	return &sucursalUUID, nil
}

// GetTerminalIDFromContext obtiene el ID del terminal del contexto
func GetTerminalIDFromContext(c *gin.Context) (string, error) {
	terminalID, exists := c.Get("terminal_id")
	if !exists {
		return "", nil // No es error, puede ser opcional
	}

	terminalStr, ok := terminalID.(string)
	if !ok {
		return "", fmt.Errorf("tipo de ID de terminal inválido en el contexto")
	}

	return terminalStr, nil
}

// CheckPermission verifica si el usuario tiene un permiso específico
func CheckPermission(c *gin.Context, permission string) bool {
	permisos, exists := c.Get("user_permisos")
	if !exists {
		return false
	}

	permisosSlice, ok := permisos.([]string)
	if !ok {
		return false
	}

	for _, p := range permisosSlice {
		if p == permission || p == "admin" {
			return true
		}
	}

	return false
}

// HasRole verifica si el usuario tiene un rol específico
func HasRole(c *gin.Context, role models.RolUsuario) bool {
	userRol, exists := c.Get("user_rol")
	if !exists {
		return false
	}

	rolStr, ok := userRol.(string)
	if !ok {
		return false
	}

	return models.RolUsuario(rolStr) == role
}

// IsAdmin verifica si el usuario es administrador
func IsAdmin(c *gin.Context) bool {
	return HasRole(c, models.RolAdmin)
}

// IsGerente verifica si el usuario es gerente
func IsGerente(c *gin.Context) bool {
	return HasRole(c, models.RolGerente)
}

// IsCajero verifica si el usuario es cajero
func IsCajero(c *gin.Context) bool {
	return HasRole(c, models.RolCajero)
}

