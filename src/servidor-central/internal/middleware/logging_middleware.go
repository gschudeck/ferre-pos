package middleware

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"ferre-pos-servidor-central/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// LoggingConfig contiene la configuración del logging
type LoggingConfig struct {
	Level           string   `yaml:"level" json:"level"`
	Format          string   `yaml:"format" json:"format"` // "json" o "text"
	Output          string   `yaml:"output" json:"output"` // "stdout", "file", "both"
	FilePath        string   `yaml:"file_path" json:"file_path"`
	MaxSize         int      `yaml:"max_size" json:"max_size"`       // MB
	MaxBackups      int      `yaml:"max_backups" json:"max_backups"` // archivos
	MaxAge          int      `yaml:"max_age" json:"max_age"`         // días
	Compress        bool     `yaml:"compress" json:"compress"`
	LogRequests     bool     `yaml:"log_requests" json:"log_requests"`
	LogResponses    bool     `yaml:"log_responses" json:"log_responses"`
	LogHeaders      bool     `yaml:"log_headers" json:"log_headers"`
	LogBody         bool     `yaml:"log_body" json:"log_body"`
	SkipPaths       []string `yaml:"skip_paths" json:"skip_paths"`
	SensitiveFields []string `yaml:"sensitive_fields" json:"sensitive_fields"`
}

// DefaultLoggingConfig retorna configuración por defecto
func DefaultLoggingConfig() LoggingConfig {
	return LoggingConfig{
		Level:        "info",
		Format:       "json",
		Output:       "stdout",
		FilePath:     "/var/log/ferre-pos/app.log",
		MaxSize:      100,
		MaxBackups:   5,
		MaxAge:       30,
		Compress:     true,
		LogRequests:  true,
		LogResponses: false,
		LogHeaders:   false,
		LogBody:      false,
		SkipPaths: []string{
			"/health",
			"/metrics",
			"/ping",
		},
		SensitiveFields: []string{
			"password",
			"token",
			"authorization",
			"api_key",
			"secret",
		},
	}
}

// responseWriter wrapper para capturar la respuesta
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// LoggingMiddleware middleware para logging de requests y responses
func LoggingMiddleware(config LoggingConfig, logService services.LogService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Verificar si debe omitir esta ruta
		if shouldSkipLogging(c.Request.URL.Path, config.SkipPaths) {
			c.Next()
			return
		}

		start := time.Now()

		// Capturar request body si está habilitado
		var requestBody []byte
		if config.LogBody && c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// Wrapper para capturar response
		var responseBody *bytes.Buffer
		if config.LogResponses {
			responseBody = &bytes.Buffer{}
			writer := &responseWriter{
				ResponseWriter: c.Writer,
				body:           responseBody,
			}
			c.Writer = writer
		}

		// Procesar request
		c.Next()

		// Calcular duración
		duration := time.Since(start)

		// Obtener información del usuario si está disponible
		var userID *uuid.UUID
		if id, exists := c.Get("user_id"); exists {
			if uid, ok := id.(uuid.UUID); ok {
				userID = &uid
			}
		}

		// Crear log entry
		logEntry := createLogEntry(c, duration, requestBody, responseBody, config, userID)

		// Enviar a servicio de logging
		if logService != nil {
			go func() {
				if err := logService.LogAccess(*userID, c.Request.URL.Path, c.Request.Method); err != nil {
					log.Printf("Error logging access: %v", err)
				}
			}()
		}

		// Log a stdout/file
		logToOutput(logEntry, config)
	}
}

// createLogEntry crea una entrada de log
func createLogEntry(c *gin.Context, duration time.Duration, requestBody []byte, responseBody *bytes.Buffer, config LoggingConfig, userID *uuid.UUID) map[string]interface{} {
	entry := map[string]interface{}{
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
		"method":      c.Request.Method,
		"path":        c.Request.URL.Path,
		"query":       c.Request.URL.RawQuery,
		"status":      c.Writer.Status(),
		"duration_ms": duration.Milliseconds(),
		"size":        c.Writer.Size(),
		"ip":          c.ClientIP(),
		"user_agent":  c.Request.UserAgent(),
	}

	// Agregar request ID si existe
	if requestID, exists := c.Get("request_id"); exists {
		entry["request_id"] = requestID
	}

	// Agregar user ID si existe
	if userID != nil {
		entry["user_id"] = userID.String()
	}

	// Agregar headers si está habilitado
	if config.LogHeaders {
		headers := make(map[string]string)
		for key, values := range c.Request.Header {
			if len(values) > 0 && !isSensitiveField(key, config.SensitiveFields) {
				headers[key] = values[0]
			}
		}
		entry["headers"] = headers
	}

	// Agregar request body si está habilitado
	if config.LogBody && len(requestBody) > 0 {
		entry["request_body"] = sanitizeBody(string(requestBody), config.SensitiveFields)
	}

	// Agregar response body si está habilitado
	if config.LogResponses && responseBody != nil {
		entry["response_body"] = sanitizeBody(responseBody.String(), config.SensitiveFields)
	}

	return entry
}

// logToOutput envía el log a la salida configurada
func logToOutput(entry map[string]interface{}, config LoggingConfig) {
	var output string

	if config.Format == "json" {
		// Formato JSON
		output = formatJSON(entry)
	} else {
		// Formato texto
		output = formatText(entry)
	}

	// Enviar a stdout
	if config.Output == "stdout" || config.Output == "both" {
		fmt.Println(output)
	}

	// Enviar a archivo
	if config.Output == "file" || config.Output == "both" {
		writeToFile(output, config.FilePath)
	}
}

// formatJSON formatea la entrada como JSON
func formatJSON(entry map[string]interface{}) string {
	// Implementación simple de JSON formatting
	// En producción, usar json.Marshal
	return fmt.Sprintf(`{"timestamp":"%v","method":"%v","path":"%v","status":%v,"duration_ms":%v}`,
		entry["timestamp"], entry["method"], entry["path"], entry["status"], entry["duration_ms"])
}

// formatText formatea la entrada como texto
func formatText(entry map[string]interface{}) string {
	return fmt.Sprintf("[%v] %v %v - %v (%vms)",
		entry["timestamp"], entry["method"], entry["path"], entry["status"], entry["duration_ms"])
}

// writeToFile escribe el log a un archivo
func writeToFile(content, filePath string) {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("Error opening log file: %v", err)
		return
	}
	defer file.Close()

	if _, err := file.WriteString(content + "\n"); err != nil {
		log.Printf("Error writing to log file: %v", err)
	}
}

// shouldSkipLogging verifica si debe omitir el logging para una ruta
func shouldSkipLogging(path string, skipPaths []string) bool {
	for _, skipPath := range skipPaths {
		if path == skipPath {
			return true
		}
	}
	return false
}

// isSensitiveField verifica si un campo es sensible
func isSensitiveField(field string, sensitiveFields []string) bool {
	fieldLower := strings.ToLower(field)
	for _, sensitive := range sensitiveFields {
		if strings.Contains(fieldLower, strings.ToLower(sensitive)) {
			return true
		}
	}
	return false
}

// sanitizeBody sanitiza el body removiendo campos sensibles
func sanitizeBody(body string, sensitiveFields []string) string {
	// Implementación simple - en producción usar regex o JSON parsing
	sanitized := body
	for _, field := range sensitiveFields {
		// Reemplazar valores de campos sensibles
		sanitized = strings.ReplaceAll(sanitized, fmt.Sprintf(`"%s":"`, field), fmt.Sprintf(`"%s":"***"`, field))
	}
	return sanitized
}

// MetricsMiddleware middleware para recolectar métricas
func MetricsMiddleware(metricsService services.MetricsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Procesar request
		c.Next()

		// Calcular métricas
		duration := time.Since(start)

		// Tags para las métricas
		tags := map[string]string{
			"method": c.Request.Method,
			"path":   c.Request.URL.Path,
			"status": strconv.Itoa(c.Writer.Status()),
		}

		// Agregar información de API
		if apiName := getAPIName(c.Request.URL.Path); apiName != "" {
			tags["api"] = apiName
		}

		// Enviar métricas
		if metricsService != nil {
			go func() {
				// Contador de requests
				metricsService.IncrementCounter("http_requests_total", tags)

				// Duración de requests
				metricsService.RecordDuration("http_request_duration", duration, tags)

				// Tamaño de response
				if c.Writer.Size() > 0 {
					metricsService.RecordMetric("http_response_size_bytes", float64(c.Writer.Size()), tags)
				}
			}()
		}
	}
}

// getAPIName extrae el nombre de la API de la ruta
func getAPIName(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) >= 3 && parts[1] == "api" {
		return parts[2]
	}
	return ""
}

// ErrorLoggingMiddleware middleware para logging de errores
func ErrorLoggingMiddleware(logService services.LogService) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Verificar si hay errores
		if len(c.Errors) > 0 {
			var userID *uuid.UUID
			if id, exists := c.Get("user_id"); exists {
				if uid, ok := id.(uuid.UUID); ok {
					userID = &uid
				}
			}

			// Log cada error
			for _, err := range c.Errors {
				if logService != nil {
					go func(error string) {
						context := map[string]interface{}{
							"method":     c.Request.Method,
							"path":       c.Request.URL.Path,
							"status":     c.Writer.Status(),
							"ip":         c.ClientIP(),
							"user_agent": c.Request.UserAgent(),
						}

						if requestID, exists := c.Get("request_id"); exists {
							context["request_id"] = requestID
						}

						logService.LogError(userID, error, context)
					}(err.Error())
				}
			}
		}
	}
}

// RateLimitMiddleware middleware básico de rate limiting
func RateLimitMiddleware(requestsPerMinute int) gin.HandlerFunc {
	// Mapa simple para tracking (en producción usar Redis)
	requestCounts := make(map[string][]time.Time)

	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		now := time.Now()

		// Limpiar requests antiguos
		if requests, exists := requestCounts[clientIP]; exists {
			var validRequests []time.Time
			for _, reqTime := range requests {
				if now.Sub(reqTime) < time.Minute {
					validRequests = append(validRequests, reqTime)
				}
			}
			requestCounts[clientIP] = validRequests
		}

		// Verificar límite
		if len(requestCounts[clientIP]) >= requestsPerMinute {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Rate limit exceeded",
				"code":    "RATE_LIMIT_EXCEEDED",
				"message": fmt.Sprintf("Máximo %d requests por minuto", requestsPerMinute),
			})
			c.Abort()
			return
		}

		// Agregar request actual
		requestCounts[clientIP] = append(requestCounts[clientIP], now)

		c.Next()
	}
}

// ResponseTimeMiddleware middleware que agrega header de tiempo de respuesta
func ResponseTimeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)
		c.Header("X-Response-Time", fmt.Sprintf("%.2fms", float64(duration.Nanoseconds())/1e6))
	}
}

// CompressionMiddleware middleware básico de compresión
func CompressionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Verificar si el cliente acepta compresión
		acceptEncoding := c.GetHeader("Accept-Encoding")
		if !strings.Contains(acceptEncoding, "gzip") {
			c.Next()
			return
		}

		// Verificar Content-Type
		contentType := c.Writer.Header().Get("Content-Type")
		if !shouldCompress(contentType) {
			c.Next()
			return
		}

		// En una implementación real, usar middleware de compresión de Gin
		c.Header("Content-Encoding", "gzip")
		c.Next()
	}
}

// shouldCompress verifica si un Content-Type debe ser comprimido
func shouldCompress(contentType string) bool {
	compressibleTypes := []string{
		"application/json",
		"application/xml",
		"text/html",
		"text/plain",
		"text/css",
		"text/javascript",
		"application/javascript",
	}

	for _, compressible := range compressibleTypes {
		if strings.Contains(contentType, compressible) {
			return true
		}
	}
	return false
}

// HealthCheckMiddleware middleware para health checks
func HealthCheckMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/health" {
			c.JSON(http.StatusOK, gin.H{
				"status":    "healthy",
				"timestamp": time.Now().UTC().Format(time.RFC3339),
				"version":   "1.0.0",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// TimeoutMiddleware middleware para timeout de requests
func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Crear contexto con timeout
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		// Reemplazar contexto del request
		c.Request = c.Request.WithContext(ctx)

		// Canal para señalar completion
		done := make(chan bool, 1)

		go func() {
			c.Next()
			done <- true
		}()

		select {
		case <-done:
			// Request completado a tiempo
			return
		case <-ctx.Done():
			// Timeout
			c.JSON(http.StatusRequestTimeout, gin.H{
				"error":   "Request timeout",
				"code":    "REQUEST_TIMEOUT",
				"message": "El request tardó demasiado en procesarse",
			})
			c.Abort()
			return
		}
	}
}
