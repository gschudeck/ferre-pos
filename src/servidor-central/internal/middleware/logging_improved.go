// Package middleware proporciona middleware mejorado para logging y manejo de errores
// con notación húngara y funcionalidades avanzadas
package middleware

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"ferre-pos-servidor-central/pkg/errors"
)

// structLoggingConfig configuración del middleware de logging
type structLoggingConfig struct {
	BoolSkipPaths       []string      `json:"skip_paths"`
	BoolLogRequestBody  bool          `json:"log_request_body"`
	BoolLogResponseBody bool          `json:"log_response_body"`
	IntMaxBodySize      int64         `json:"max_body_size"`
	DurationSlowRequest time.Duration `json:"slow_request_threshold"`
	ArrSensitiveFields  []string      `json:"sensitive_fields"`
	BoolLogHeaders      bool          `json:"log_headers"`
	ArrSkipHeaders      []string      `json:"skip_headers"`
}

// structRequestContext contexto de request con información adicional
type structRequestContext struct {
	StrRequestID    string
	StrUserID       string
	StrSucursalID   string
	StrIP           string
	StrUserAgent    string
	TimeStart       time.Time
	PtrLogger       logger.interfaceLogger
	PtrMetrics      metrics.interfaceManager
	MapCustomFields map[string]interface{}
	MutexFields     sync.RWMutex
}

// structResponseWriter wrapper para capturar respuesta
type structResponseWriter struct {
	gin.ResponseWriter
	PtrBody       *bytes.Buffer
	IntStatusCode int
	IntSize       int64
	BoolWritten   bool
}

// Write implementa io.Writer
func (ptrWriter *structResponseWriter) Write(arrData []byte) (int, error) {
	if !ptrWriter.BoolWritten {
		ptrWriter.BoolWritten = true
	}

	intN, err := ptrWriter.ResponseWriter.Write(arrData)
	ptrWriter.IntSize += int64(intN)

	if ptrWriter.PtrBody != nil {
		ptrWriter.PtrBody.Write(arrData)
	}

	return intN, err
}

// WriteHeader implementa ResponseWriter
func (ptrWriter *structResponseWriter) WriteHeader(intStatusCode int) {
	ptrWriter.IntStatusCode = intStatusCode
	ptrWriter.ResponseWriter.WriteHeader(intStatusCode)
}

// WriteString implementa gin.ResponseWriter
func (ptrWriter *structResponseWriter) WriteString(strData string) (int, error) {
	return ptrWriter.Write([]byte(strData))
}

// LoggingMiddleware crea middleware de logging mejorado
func LoggingMiddleware(ptrLogger logger.interfaceLogger, ptrMetrics metrics.interfaceManager, ptrConfig *structLoggingConfig) gin.HandlerFunc {
	if ptrConfig == nil {
		ptrConfig = &structLoggingConfig{
			BoolLogRequestBody:  false,
			BoolLogResponseBody: false,
			IntMaxBodySize:      1024 * 1024, // 1MB
			DurationSlowRequest: 5 * time.Second,
			BoolLogHeaders:      true,
			ArrSensitiveFields:  []string{"password", "token", "authorization", "cookie"},
			ArrSkipHeaders:      []string{"authorization", "cookie", "x-api-key"},
		}
	}

	return func(ptrCtx *gin.Context) {
		// Verificar si debe saltarse esta ruta
		if shouldSkipPath(ptrCtx.Request.URL.Path, ptrConfig.BoolSkipPaths) {
			ptrCtx.Next()
			return
		}

		// Crear contexto de request
		ptrRequestCtx := createRequestContext(ptrCtx, ptrLogger, ptrMetrics)

		// Establecer contexto en Gin
		ptrCtx.Set("request_context", ptrRequestCtx)
		ptrCtx.Set("request_id", ptrRequestCtx.StrRequestID)
		ptrCtx.Set("logger", ptrRequestCtx.PtrLogger)

		// Crear response writer
		ptrResponseWriter := &structResponseWriter{
			ResponseWriter: ptrCtx.Writer,
			IntStatusCode:  http.StatusOK,
		}

		if ptrConfig.BoolLogResponseBody {
			ptrResponseWriter.PtrBody = &bytes.Buffer{}
		}

		ptrCtx.Writer = ptrResponseWriter

		// Log de inicio de request
		logRequestStart(ptrRequestCtx, ptrCtx, ptrConfig)

		// Incrementar métricas
		if ptrMetrics != nil {
			ptrMetrics.IncActiveRequests()
		}

		// Procesar request
		ptrCtx.Next()

		// Decrementar métricas
		if ptrMetrics != nil {
			ptrMetrics.DecActiveRequests()
		}

		// Log de finalización de request
		logRequestEnd(ptrRequestCtx, ptrCtx, ptrResponseWriter, ptrConfig)
	}
}

// RecoveryMiddleware middleware de recovery mejorado
func RecoveryMiddleware(ptrLogger logger.interfaceLogger, ptrMetrics metrics.interfaceManager) gin.HandlerFunc {
	return func(ptrCtx *gin.Context) {
		defer func() {
			if objRecover := recover(); objRecover != nil {
				handlePanic(ptrCtx, objRecover, ptrLogger, ptrMetrics)
			}
		}()

		ptrCtx.Next()
	}
}

// ErrorHandlingMiddleware middleware para manejo centralizado de errores
func ErrorHandlingMiddleware(ptrLogger logger.interfaceLogger) gin.HandlerFunc {
	return func(ptrCtx *gin.Context) {
		ptrCtx.Next()

		// Procesar errores acumulados
		if len(ptrCtx.Errors) > 0 {
			handleGinErrors(ptrCtx, ptrLogger)
		}
	}
}

// RequestIDMiddleware middleware para generar Request ID
func RequestIDMiddleware() gin.HandlerFunc {
	return func(ptrCtx *gin.Context) {
		strRequestID := ptrCtx.GetHeader("X-Request-ID")
		if strRequestID == "" {
			strRequestID = generateRequestID()
		}

		ptrCtx.Set("request_id", strRequestID)
		ptrCtx.Header("X-Request-ID", strRequestID)
		ptrCtx.Next()
	}
}

// TimeoutMiddleware middleware para timeout de requests
func TimeoutMiddleware(durationTimeout time.Duration) gin.HandlerFunc {
	return func(ptrCtx *gin.Context) {
		ctxWithTimeout, funcCancel := context.WithTimeout(ptrCtx.Request.Context(), durationTimeout)
		defer funcCancel()

		ptrCtx.Request = ptrCtx.Request.WithContext(ctxWithTimeout)

		chanDone := make(chan struct{})
		go func() {
			defer close(chanDone)
			ptrCtx.Next()
		}()

		select {
		case <-chanDone:
			// Request completada normalmente
		case <-ctxWithTimeout.Done():
			// Timeout alcanzado
			if ctxWithTimeout.Err() == context.DeadlineExceeded {
				ptrCtx.JSON(http.StatusRequestTimeout, gin.H{
					"error":      "Request timeout",
					"message":    "La request tardó demasiado en procesarse",
					"request_id": ptrCtx.GetString("request_id"),
					"timeout":    durationTimeout.String(),
				})
				ptrCtx.Abort()
			}
		}
	}
}

// Funciones helper

// createRequestContext crea el contexto de request
func createRequestContext(ptrCtx *gin.Context, ptrLogger logger.interfaceLogger, ptrMetrics metrics.interfaceManager) *structRequestContext {
	strRequestID := ptrCtx.GetString("request_id")
	if strRequestID == "" {
		strRequestID = generateRequestID()
	}

	strUserID := extractUserID(ptrCtx)
	strSucursalID := extractSucursalID(ptrCtx)
	strIP := extractClientIP(ptrCtx)
	strUserAgent := ptrCtx.GetHeader("User-Agent")

	ptrRequestLogger := ptrLogger.WithRequestID(strRequestID)
	if strUserID != "" {
		ptrRequestLogger = ptrRequestLogger.WithUserID(strUserID)
	}

	return &structRequestContext{
		StrRequestID:    strRequestID,
		StrUserID:       strUserID,
		StrSucursalID:   strSucursalID,
		StrIP:           strIP,
		StrUserAgent:    strUserAgent,
		TimeStart:       time.Now(),
		PtrLogger:       ptrRequestLogger,
		PtrMetrics:      ptrMetrics,
		MapCustomFields: make(map[string]interface{}),
	}
}

// logRequestStart registra el inicio de la request
func logRequestStart(ptrRequestCtx *structRequestContext, ptrCtx *gin.Context, ptrConfig *structLoggingConfig) {
	arrFields := []zap.Field{
		zap.String("method", ptrCtx.Request.Method),
		zap.String("path", ptrCtx.Request.URL.Path),
		zap.String("query", ptrCtx.Request.URL.RawQuery),
		zap.String("ip", ptrRequestCtx.StrIP),
		zap.String("user_agent", ptrRequestCtx.StrUserAgent),
		zap.Time("start_time", ptrRequestCtx.TimeStart),
	}

	if ptrRequestCtx.StrUserID != "" {
		arrFields = append(arrFields, zap.String("user_id", ptrRequestCtx.StrUserID))
	}

	if ptrRequestCtx.StrSucursalID != "" {
		arrFields = append(arrFields, zap.String("sucursal_id", ptrRequestCtx.StrSucursalID))
	}

	// Log headers si está habilitado
	if ptrConfig.BoolLogHeaders {
		mapHeaders := sanitizeHeaders(ptrCtx.Request.Header, ptrConfig.ArrSkipHeaders)
		if len(mapHeaders) > 0 {
			arrFields = append(arrFields, zap.Any("headers", mapHeaders))
		}
	}

	// Log request body si está habilitado
	if ptrConfig.BoolLogRequestBody && shouldLogBody(ptrCtx.Request) {
		strBody := readRequestBody(ptrCtx, ptrConfig.IntMaxBodySize)
		if strBody != "" {
			strSanitizedBody := sanitizeBody(strBody, ptrConfig.ArrSensitiveFields)
			arrFields = append(arrFields, zap.String("request_body", strSanitizedBody))
		}
	}

	ptrRequestCtx.PtrLogger.Info("Request iniciada", arrFields...)
}

// logRequestEnd registra el final de la request
func logRequestEnd(ptrRequestCtx *structRequestContext, ptrCtx *gin.Context, ptrResponseWriter *structResponseWriter, ptrConfig *structLoggingConfig) {
	timeDuration := time.Since(ptrRequestCtx.TimeStart)
	intStatusCode := ptrResponseWriter.IntStatusCode
	intResponseSize := ptrResponseWriter.IntSize

	arrFields := []zap.Field{
		zap.String("method", ptrCtx.Request.Method),
		zap.String("path", ptrCtx.Request.URL.Path),
		zap.Int("status_code", intStatusCode),
		zap.Int64("response_size", intResponseSize),
		zap.Duration("duration", timeDuration),
		zap.Int64("duration_ms", timeDuration.Milliseconds()),
	}

	// Agregar campos personalizados
	ptrRequestCtx.MutexFields.RLock()
	for strKey, objValue := range ptrRequestCtx.MapCustomFields {
		arrFields = append(arrFields, zap.Any(strKey, objValue))
	}
	ptrRequestCtx.MutexFields.RUnlock()

	// Log response body si está habilitado
	if ptrConfig.BoolLogResponseBody && ptrResponseWriter.PtrBody != nil {
		strBody := ptrResponseWriter.PtrBody.String()
		if strBody != "" {
			strSanitizedBody := sanitizeBody(strBody, ptrConfig.ArrSensitiveFields)
			arrFields = append(arrFields, zap.String("response_body", strSanitizedBody))
		}
	}

	// Registrar métricas
	if ptrRequestCtx.PtrMetrics != nil {
		ptrRequestCtx.PtrMetrics.RecordHTTPRequest(
			ptrCtx.Request.Method,
			ptrCtx.Request.URL.Path,
			strconv.Itoa(intStatusCode),
			timeDuration.Seconds(),
			ptrCtx.Request.ContentLength,
			intResponseSize,
		)
	}

	// Determinar nivel de log
	strLogLevel := determineLogLevel(intStatusCode, timeDuration, ptrConfig.DurationSlowRequest)

	switch strLogLevel {
	case "error":
		ptrRequestCtx.PtrLogger.Error("Request completada con error", arrFields...)
	case "warn":
		ptrRequestCtx.PtrLogger.Warn("Request lenta detectada", arrFields...)
	default:
		ptrRequestCtx.PtrLogger.Info("Request completada", arrFields...)
	}
}

// handlePanic maneja panics del sistema
func handlePanic(ptrCtx *gin.Context, objRecover interface{}, ptrLogger logger.interfaceLogger, ptrMetrics metrics.interfaceManager) {
	strRequestID := ptrCtx.GetString("request_id")
	strStackTrace := getStackTrace()

	// Log del panic
	ptrLogger.Error("Panic recuperado",
		zap.String("request_id", strRequestID),
		zap.String("method", ptrCtx.Request.Method),
		zap.String("path", ptrCtx.Request.URL.Path),
		zap.Any("panic", objRecover),
		zap.String("stack_trace", strStackTrace),
	)

	// Registrar métrica de error
	if ptrMetrics != nil {
		ptrMetrics.RecordBusinessError("panic", "system_panic")
	}

	// Respuesta de error
	ptrAppError := errors.NewInternal("Error interno del servidor", "INTERNAL_PANIC").
		WithDetail("panic", fmt.Sprintf("%v", objRecover)).
		WithRequestID(strRequestID)

	ptrCtx.JSON(http.StatusInternalServerError, gin.H{
		"error":      "Internal server error",
		"message":    "Ha ocurrido un error interno del servidor",
		"request_id": strRequestID,
		"timestamp":  time.Now().Format(time.RFC3339),
	})

	ptrCtx.Abort()
}

// handleGinErrors maneja errores acumulados en Gin
func handleGinErrors(ptrCtx *gin.Context, ptrLogger logger.interfaceLogger) {
	strRequestID := ptrCtx.GetString("request_id")

	for _, ptrGinError := range ptrCtx.Errors {
		arrFields := []zap.Field{
			zap.String("request_id", strRequestID),
			zap.String("method", ptrCtx.Request.Method),
			zap.String("path", ptrCtx.Request.URL.Path),
			zap.String("error_type", ptrGinError.Type.String()),
			zap.Error(ptrGinError.Err),
		}

		if ptrGinError.Meta != nil {
			arrFields = append(arrFields, zap.Any("meta", ptrGinError.Meta))
		}

		switch ptrGinError.Type {
		case gin.ErrorTypeBind:
			ptrLogger.Warn("Error de binding", arrFields...)
		case gin.ErrorTypeRender:
			ptrLogger.Error("Error de renderizado", arrFields...)
		case gin.ErrorTypePublic:
			ptrLogger.Info("Error público", arrFields...)
		default:
			ptrLogger.Error("Error no categorizado", arrFields...)
		}
	}
}

// Funciones de utilidad

// generateRequestID genera un ID único para la request
func generateRequestID() string {
	return uuid.New().String()
}

// extractUserID extrae el ID del usuario del contexto
func extractUserID(ptrCtx *gin.Context) string {
	if objUserID, boolExists := ptrCtx.Get("user_id"); boolExists {
		if strUserID, boolOk := objUserID.(string); boolOk {
			return strUserID
		}
	}
	return ""
}

// extractSucursalID extrae el ID de la sucursal del contexto
func extractSucursalID(ptrCtx *gin.Context) string {
	if objSucursalID, boolExists := ptrCtx.Get("sucursal_id"); boolExists {
		if strSucursalID, boolOk := objSucursalID.(string); boolOk {
			return strSucursalID
		}
	}
	return ""
}

// extractClientIP extrae la IP del cliente
func extractClientIP(ptrCtx *gin.Context) string {
	// Verificar headers de proxy
	strIP := ptrCtx.GetHeader("X-Forwarded-For")
	if strIP != "" {
		// Tomar la primera IP si hay múltiples
		arrIPs := strings.Split(strIP, ",")
		return strings.TrimSpace(arrIPs[0])
	}

	strIP = ptrCtx.GetHeader("X-Real-IP")
	if strIP != "" {
		return strIP
	}

	return ptrCtx.ClientIP()
}

// shouldSkipPath verifica si debe saltarse el logging para una ruta
func shouldSkipPath(strPath string, arrSkipPaths []string) bool {
	for _, strSkipPath := range arrSkipPaths {
		if strings.HasPrefix(strPath, strSkipPath) {
			return true
		}
	}
	return false
}

// shouldLogBody verifica si debe loggear el body de la request
func shouldLogBody(ptrRequest *http.Request) bool {
	strContentType := ptrRequest.Header.Get("Content-Type")

	// Solo loggear para tipos de contenido de texto
	return strings.Contains(strContentType, "application/json") ||
		strings.Contains(strContentType, "application/xml") ||
		strings.Contains(strContentType, "text/")
}

// readRequestBody lee el body de la request
func readRequestBody(ptrCtx *gin.Context, intMaxSize int64) string {
	if ptrCtx.Request.Body == nil {
		return ""
	}

	// Leer body limitado
	ptrLimitedReader := io.LimitReader(ptrCtx.Request.Body, intMaxSize)
	arrBody, err := io.ReadAll(ptrLimitedReader)
	if err != nil {
		return ""
	}

	// Restaurar body para uso posterior
	ptrCtx.Request.Body = io.NopCloser(bytes.NewBuffer(arrBody))

	return string(arrBody)
}

// sanitizeHeaders sanitiza headers sensibles
func sanitizeHeaders(mapHeaders http.Header, arrSkipHeaders []string) map[string]string {
	mapSanitized := make(map[string]string)

	for strKey, arrValues := range mapHeaders {
		strLowerKey := strings.ToLower(strKey)

		boolSkip := false
		for _, strSkipHeader := range arrSkipHeaders {
			if strings.ToLower(strSkipHeader) == strLowerKey {
				boolSkip = true
				break
			}
		}

		if !boolSkip && len(arrValues) > 0 {
			mapSanitized[strKey] = arrValues[0]
		}
	}

	return mapSanitized
}

// sanitizeBody sanitiza campos sensibles del body
func sanitizeBody(strBody string, arrSensitiveFields []string) string {
	for _, strField := range arrSensitiveFields {
		// Patrón simple para JSON
		strPattern := fmt.Sprintf(`"%s"\s*:\s*"[^"]*"`, strField)
		strReplacement := fmt.Sprintf(`"%s":"***"`, strField)
		strBody = strings.ReplaceAll(strBody, strPattern, strReplacement)
	}

	return strBody
}

// determineLogLevel determina el nivel de log basado en status y duración
func determineLogLevel(intStatusCode int, timeDuration time.Duration, durationSlowThreshold time.Duration) string {
	if intStatusCode >= 500 {
		return "error"
	}

	if intStatusCode >= 400 || timeDuration > durationSlowThreshold {
		return "warn"
	}

	return "info"
}

// getStackTrace obtiene el stack trace actual
func getStackTrace() string {
	const intDepth = 32
	var arrPcs [intDepth]uintptr
	intN := runtime.Callers(3, arrPcs[:])

	var arrLines []string
	for intI := 0; intI < intN; intI++ {
		ptrFrame := runtime.FuncForPC(arrPcs[intI])
		if ptrFrame != nil {
			strFile, intLine := ptrFrame.FileLine(arrPcs[intI])
			arrLines = append(arrLines, fmt.Sprintf("%s:%d %s", strFile, intLine, ptrFrame.Name()))
		}
	}

	return strings.Join(arrLines, "\n")
}

// AddCustomField agrega un campo personalizado al contexto de request
func AddCustomField(ptrCtx *gin.Context, strKey string, objValue interface{}) {
	if objRequestCtx, boolExists := ptrCtx.Get("request_context"); boolExists {
		if ptrRequestCtx, boolOk := objRequestCtx.(*structRequestContext); boolOk {
			ptrRequestCtx.MutexFields.Lock()
			ptrRequestCtx.MapCustomFields[strKey] = objValue
			ptrRequestCtx.MutexFields.Unlock()
		}
	}
}

// GetRequestLogger obtiene el logger de la request
func GetRequestLogger(ptrCtx *gin.Context) logger.interfaceLogger {
	if objRequestCtx, boolExists := ptrCtx.Get("request_context"); boolExists {
		if ptrRequestCtx, boolOk := objRequestCtx.(*structRequestContext); boolOk {
			return ptrRequestCtx.PtrLogger
		}
	}

	// Fallback al logger genérico
	if objLogger, boolExists := ptrCtx.Get("logger"); boolExists {
		if ptrLogger, boolOk := objLogger.(logger.interfaceLogger); boolOk {
			return ptrLogger
		}
	}

	return nil
}
