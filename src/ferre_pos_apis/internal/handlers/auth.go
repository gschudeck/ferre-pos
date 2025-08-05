package handlers

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"ferre_pos_apis/internal/config"
	"ferre_pos_apis/internal/database"
	"ferre_pos_apis/internal/logger"
	"ferre_pos_apis/internal/models"
	"ferre_pos_apis/pkg/validator"
)

// AuthHandler handler de autenticación
type AuthHandler struct {
	db        *database.Database
	logger    logger.Logger
	validator validator.Validator
	config    *config.APIConfig
}

// NewAuthHandler crea un nuevo handler de autenticación
func NewAuthHandler(db *database.Database, log logger.Logger, val validator.Validator, cfg *config.APIConfig) *AuthHandler {
	return &AuthHandler{
		db:        db,
		logger:    log,
		validator: val,
		config:    cfg,
	}
}

// Login maneja el login de usuarios
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_JSON",
				Message: "JSON inválido en el cuerpo de la petición",
			},
			RequestID: getRequestID(c),
			Timestamp: time.Now(),
		})
		return
	}

	// Validar datos de entrada
	if err := h.validator.ValidateStruct(&req); err != nil {
		validationErrors := h.validator.GetValidationErrors(err)
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "VALIDATION_ERROR",
				Message: "Error de validación en los datos enviados",
				Details: models.JSONB{"validation_errors": validationErrors},
			},
			RequestID: getRequestID(c),
			Timestamp: time.Now(),
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Buscar usuario por RUT
	usuario, err := h.getUserByRUT(ctx, req.RUT)
	if err != nil {
		if err == sql.ErrNoRows {
			h.logger.WithField("rut", req.RUT).Warn("Intento de login con RUT inexistente")
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "INVALID_CREDENTIALS",
					Message: "Credenciales inválidas",
				},
				RequestID: getRequestID(c),
				Timestamp: time.Now(),
			})
			return
		}

		h.logger.WithError(err).Error("Error consultando usuario")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "Error interno del servidor",
			},
			RequestID: getRequestID(c),
			Timestamp: time.Now(),
		})
		return
	}

	// Verificar si el usuario está activo
	if !usuario.Activo {
		h.logger.WithField("user_id", usuario.ID).Warn("Intento de login con usuario inactivo")
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "USER_INACTIVE",
				Message: "Usuario inactivo",
			},
			RequestID: getRequestID(c),
			Timestamp: time.Now(),
		})
		return
	}

	// Verificar si el usuario está bloqueado
	if usuario.BloqueadoHasta != nil && usuario.BloqueadoHasta.After(time.Now()) {
		h.logger.WithField("user_id", usuario.ID).Warn("Intento de login con usuario bloqueado")
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "USER_BLOCKED",
				Message: "Usuario bloqueado temporalmente",
				Details: models.JSONB{"blocked_until": usuario.BloqueadoHasta},
			},
			RequestID: getRequestID(c),
			Timestamp: time.Now(),
		})
		return
	}

	// Verificar password
	if !h.verifyPassword(req.Password, usuario.PasswordHash, usuario.Salt) {
		// Incrementar intentos fallidos
		if err := h.incrementFailedAttempts(ctx, usuario.ID); err != nil {
			h.logger.WithError(err).Error("Error incrementando intentos fallidos")
		}

		h.logger.WithField("user_id", usuario.ID).Warn("Intento de login con password incorrecta")
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_CREDENTIALS",
				Message: "Credenciales inválidas",
			},
			RequestID: getRequestID(c),
			Timestamp: time.Now(),
		})
		return
	}

	// Resetear intentos fallidos y actualizar último acceso
	if err := h.resetFailedAttempts(ctx, usuario.ID); err != nil {
		h.logger.WithError(err).Error("Error reseteando intentos fallidos")
	}

	// Generar tokens
	token, refreshToken, expiresAt, err := h.generateTokens(usuario)
	if err != nil {
		h.logger.WithError(err).Error("Error generando tokens")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "TOKEN_GENERATION_ERROR",
				Message: "Error generando token de acceso",
			},
			RequestID: getRequestID(c),
			Timestamp: time.Now(),
		})
		return
	}

	// Actualizar hash de sesión activa
	sessionHash := h.generateSessionHash(token)
	if err := h.updateActiveSession(ctx, usuario.ID, sessionHash); err != nil {
		h.logger.WithError(err).Error("Error actualizando sesión activa")
	}

	// Preparar respuesta
	response := models.LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		User:         *usuario,
	}

	// Limpiar datos sensibles del usuario en la respuesta
	response.User.PasswordHash = ""
	response.User.Salt = ""
	response.User.HashSesionActiva = nil

	h.logger.WithField("user_id", usuario.ID).Info("Login exitoso")

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      response,
		RequestID: getRequestID(c),
		Timestamp: time.Now(),
	})
}

// RefreshToken maneja la renovación de tokens
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" validate:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_JSON",
				Message: "JSON inválido en el cuerpo de la petición",
			},
			RequestID: getRequestID(c),
			Timestamp: time.Now(),
		})
		return
	}

	// Validar refresh token
	token, err := jwt.Parse(req.RefreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(h.config.Auth.JWTSecret), nil
	})

	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REFRESH_TOKEN",
				Message: "Refresh token inválido o expirado",
			},
			RequestID: getRequestID(c),
			Timestamp: time.Now(),
		})
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_CLAIMS",
				Message: "Claims del token inválidos",
			},
			RequestID: getRequestID(c),
			Timestamp: time.Now(),
		})
		return
	}

	// Verificar que es un refresh token
	tokenType, _ := claims["type"].(string)
	if tokenType != "refresh" {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_TOKEN_TYPE",
				Message: "Token no es de tipo refresh",
			},
			RequestID: getRequestID(c),
			Timestamp: time.Now(),
		})
		return
	}

	// Obtener usuario
	userID, _ := claims["user_id"].(string)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	usuario, err := h.getUserByID(ctx, userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "USER_NOT_FOUND",
				Message: "Usuario no encontrado",
			},
			RequestID: getRequestID(c),
			Timestamp: time.Now(),
		})
		return
	}

	// Verificar que el usuario sigue activo
	if !usuario.Activo {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "USER_INACTIVE",
				Message: "Usuario inactivo",
			},
			RequestID: getRequestID(c),
			Timestamp: time.Now(),
		})
		return
	}

	// Generar nuevos tokens
	newToken, newRefreshToken, expiresAt, err := h.generateTokens(usuario)
	if err != nil {
		h.logger.WithError(err).Error("Error generando nuevos tokens")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "TOKEN_GENERATION_ERROR",
				Message: "Error generando nuevos tokens",
			},
			RequestID: getRequestID(c),
			Timestamp: time.Now(),
		})
		return
	}

	response := models.LoginResponse{
		Token:        newToken,
		RefreshToken: newRefreshToken,
		ExpiresAt:    expiresAt,
		User:         *usuario,
	}

	// Limpiar datos sensibles
	response.User.PasswordHash = ""
	response.User.Salt = ""
	response.User.HashSesionActiva = nil

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      response,
		RequestID: getRequestID(c),
		Timestamp: time.Now(),
	})
}

// Logout maneja el cierre de sesión
func (h *AuthHandler) Logout(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "USER_NOT_AUTHENTICATED",
				Message: "Usuario no autenticado",
			},
			RequestID: getRequestID(c),
			Timestamp: time.Now(),
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Limpiar sesión activa
	if err := h.clearActiveSession(ctx, userID.(string)); err != nil {
		h.logger.WithError(err).Error("Error limpiando sesión activa")
	}

	h.logger.WithField("user_id", userID).Info("Logout exitoso")

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      gin.H{"message": "Sesión cerrada exitosamente"},
		RequestID: getRequestID(c),
		Timestamp: time.Now(),
	})
}

// Métodos auxiliares

// getUserByRUT obtiene un usuario por RUT
func (h *AuthHandler) getUserByRUT(ctx context.Context, rut string) (*models.Usuario, error) {
	query := `
		SELECT id, rut, nombre, apellido, email, telefono, rol, sucursal_id, 
		       password_hash, salt, activo, ultimo_acceso, intentos_fallidos, 
		       bloqueado_hasta, fecha_creacion, fecha_modificacion,
		       configuracion_personal, permisos_especiales, cache_permisos,
		       hash_sesion_activa, ultimo_terminal_id, preferencias_ui
		FROM usuarios 
		WHERE rut = $1`

	var usuario models.Usuario
	err := h.db.QueryRowContext(ctx, query, rut).Scan(
		&usuario.ID, &usuario.RUT, &usuario.Nombre, &usuario.Apellido,
		&usuario.Email, &usuario.Telefono, &usuario.Rol, &usuario.SucursalID,
		&usuario.PasswordHash, &usuario.Salt, &usuario.Activo, &usuario.UltimoAcceso,
		&usuario.IntentosFallidos, &usuario.BloqueadoHasta, &usuario.FechaCreacion,
		&usuario.FechaModificacion, &usuario.ConfiguracionPersonal, &usuario.PermisosEspeciales,
		&usuario.CachePermisos, &usuario.HashSesionActiva, &usuario.UltimoTerminalID,
		&usuario.PreferenciasUI,
	)

	if err != nil {
		return nil, err
	}

	return &usuario, nil
}

// getUserByID obtiene un usuario por ID
func (h *AuthHandler) getUserByID(ctx context.Context, userID string) (*models.Usuario, error) {
	query := `
		SELECT id, rut, nombre, apellido, email, telefono, rol, sucursal_id, 
		       password_hash, salt, activo, ultimo_acceso, intentos_fallidos, 
		       bloqueado_hasta, fecha_creacion, fecha_modificacion,
		       configuracion_personal, permisos_especiales, cache_permisos,
		       hash_sesion_activa, ultimo_terminal_id, preferencias_ui
		FROM usuarios 
		WHERE id = $1`

	var usuario models.Usuario
	err := h.db.QueryRowContext(ctx, query, userID).Scan(
		&usuario.ID, &usuario.RUT, &usuario.Nombre, &usuario.Apellido,
		&usuario.Email, &usuario.Telefono, &usuario.Rol, &usuario.SucursalID,
		&usuario.PasswordHash, &usuario.Salt, &usuario.Activo, &usuario.UltimoAcceso,
		&usuario.IntentosFallidos, &usuario.BloqueadoHasta, &usuario.FechaCreacion,
		&usuario.FechaModificacion, &usuario.ConfiguracionPersonal, &usuario.PermisosEspeciales,
		&usuario.CachePermisos, &usuario.HashSesionActiva, &usuario.UltimoTerminalID,
		&usuario.PreferenciasUI,
	)

	if err != nil {
		return nil, err
	}

	return &usuario, nil
}

// verifyPassword verifica la contraseña del usuario
func (h *AuthHandler) verifyPassword(password, hash, salt string) bool {
	// Combinar password con salt
	saltedPassword := password + salt
	
	// Verificar con bcrypt
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(saltedPassword))
	return err == nil
}

// generateTokens genera tokens de acceso y refresh
func (h *AuthHandler) generateTokens(usuario *models.Usuario) (string, string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(h.config.Auth.TokenExpiry)
	refreshExpiresAt := now.Add(h.config.Auth.RefreshTokenExpiry)

	// Claims para el token de acceso
	accessClaims := jwt.MapClaims{
		"user_id":     usuario.ID.String(),
		"rut":         usuario.RUT,
		"nombre":      usuario.Nombre,
		"role":        string(usuario.Rol),
		"sucursal_id": usuario.SucursalID,
		"type":        "access",
		"iat":         now.Unix(),
		"exp":         expiresAt.Unix(),
	}

	// Claims para el refresh token
	refreshClaims := jwt.MapClaims{
		"user_id": usuario.ID.String(),
		"type":    "refresh",
		"iat":     now.Unix(),
		"exp":     refreshExpiresAt.Unix(),
	}

	// Generar token de acceso
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(h.config.Auth.JWTSecret))
	if err != nil {
		return "", "", time.Time{}, err
	}

	// Generar refresh token
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(h.config.Auth.JWTSecret))
	if err != nil {
		return "", "", time.Time{}, err
	}

	return accessTokenString, refreshTokenString, expiresAt, nil
}

// generateSessionHash genera hash de sesión
func (h *AuthHandler) generateSessionHash(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// incrementFailedAttempts incrementa los intentos fallidos
func (h *AuthHandler) incrementFailedAttempts(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE usuarios 
		SET intentos_fallidos = intentos_fallidos + 1,
		    bloqueado_hasta = CASE 
		        WHEN intentos_fallidos + 1 >= 5 THEN NOW() + INTERVAL '30 minutes'
		        ELSE bloqueado_hasta
		    END
		WHERE id = $1`

	_, err := h.db.ExecContext(ctx, query, userID)
	return err
}

// resetFailedAttempts resetea los intentos fallidos
func (h *AuthHandler) resetFailedAttempts(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE usuarios 
		SET intentos_fallidos = 0,
		    bloqueado_hasta = NULL,
		    ultimo_acceso = NOW()
		WHERE id = $1`

	_, err := h.db.ExecContext(ctx, query, userID)
	return err
}

// updateActiveSession actualiza la sesión activa
func (h *AuthHandler) updateActiveSession(ctx context.Context, userID uuid.UUID, sessionHash string) error {
	query := `
		UPDATE usuarios 
		SET hash_sesion_activa = $2
		WHERE id = $1`

	_, err := h.db.ExecContext(ctx, query, userID, sessionHash)
	return err
}

// clearActiveSession limpia la sesión activa
func (h *AuthHandler) clearActiveSession(ctx context.Context, userID string) error {
	query := `
		UPDATE usuarios 
		SET hash_sesion_activa = NULL
		WHERE id = $1`

	_, err := h.db.ExecContext(ctx, query, userID)
	return err
}

// getRequestID obtiene el ID de la request del contexto
func getRequestID(c *gin.Context) string {
	if requestID, exists := c.Get("request_id"); exists {
		return requestID.(string)
	}
	return ""
}

