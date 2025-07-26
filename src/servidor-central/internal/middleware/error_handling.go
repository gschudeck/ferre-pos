// Package middleware proporciona middleware mejorado para manejo de errores
// con notación húngara y respuestas estructuradas
package middleware

import (
	"context"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ferre-pos-servidor-central/pkg/errors"
)

// structErrorHandlerConfig configuración del manejador de errores
type structErrorHandlerConfig struct {
	BoolIncludeStackTrace bool                        `json:"include_stack_trace"`
	BoolLogAllErrors      bool                        `json:"log_all_errors"`
	DurationTimeout       time.Duration               `json:"timeout"`
	StrEnvironment        string                      `json:"environment"`
	BoolDetailedErrors    bool                        `json:"detailed_errors"`
	MapCustomHandlers     map[string]funcErrorHandler `json:"-"`
}

// funcErrorHandler función para manejar errores específicos
type funcErrorHandler func(*gin.Context, error, logger.interfaceLogger) *models.structRespuestaAPI

// structErrorResponse respuesta de error estructurada
type structErrorResponse struct {
	BoolExito     bool                   `json:"exito"`
	StrMensaje    string                 `json:"mensaje"`
	StrCodigo     string                 `json:"codigo,omitempty"`
	StrTipo       string                 `json:"tipo,omitempty"`
	ObjDetalles   interface{}            `json:"detalles,omitempty"`
	StrStackTrace string                 `json:"stack_trace,omitempty"`
	StrRequestID  string                 `json:"request_id,omitempty"`
	TimeTimestamp time.Time              `json:"timestamp"`
	MapContexto   map[string]interface{} `json:"contexto,omitempty"`
}

// ErrorHandlerMiddleware middleware principal para manejo de errores
func ErrorHandlerMiddleware(ptrLogger logger.interfaceLogger, ptrMetrics metrics.interfaceManager, ptrConfig *structErrorHandlerConfig) gin.HandlerFunc {
	if ptrConfig == nil {
		ptrConfig = &structErrorHandlerConfig{
			BoolIncludeStackTrace: false,
			BoolLogAllErrors:      true,
			DurationTimeout:       30 * time.Second,
			StrEnvironment:        "production",
			BoolDetailedErrors:    false,
			MapCustomHandlers:     make(map[string]funcErrorHandler),
		}
	}

	return func(ptrCtx *gin.Context) {
		// Configurar timeout si está especificado
		if ptrConfig.DurationTimeout > 0 {
			ctxWithTimeout, funcCancel := context.WithTimeout(ptrCtx.Request.Context(), ptrConfig.DurationTimeout)
			defer funcCancel()
			ptrCtx.Request = ptrCtx.Request.WithContext(ctxWithTimeout)
		}

		// Procesar request
		ptrCtx.Next()

		// Manejar errores si los hay
		if len(ptrCtx.Errors) > 0 {
			handleErrors(ptrCtx, ptrLogger, ptrMetrics, ptrConfig)
		}
	}
}

// PanicRecoveryMiddleware middleware para recuperación de panics
func PanicRecoveryMiddleware(ptrLogger logger.interfaceLogger, ptrMetrics metrics.interfaceManager, ptrConfig *structErrorHandlerConfig) gin.HandlerFunc {
	return func(ptrCtx *gin.Context) {
		defer func() {
			if objRecover := recover(); objRecover != nil {
				handlePanic(ptrCtx, objRecover, ptrLogger, ptrMetrics, ptrConfig)
			}
		}()

		ptrCtx.Next()
	}
}

// ValidationErrorMiddleware middleware específico para errores de validación
func ValidationErrorMiddleware(ptrLogger logger.interfaceLogger) gin.HandlerFunc {
	return func(ptrCtx *gin.Context) {
		ptrCtx.Next()

		// Buscar errores de validación específicamente
		for _, ptrGinError := range ptrCtx.Errors {
			if ptrGinError.Type == gin.ErrorTypeBind {
				handleValidationError(ptrCtx, ptrGinError.Err, ptrLogger)
				return
			}
		}
	}
}

// DatabaseErrorMiddleware middleware específico para errores de base de datos
func DatabaseErrorMiddleware(ptrLogger logger.interfaceLogger, ptrMetrics metrics.interfaceManager) gin.HandlerFunc {
	return func(ptrCtx *gin.Context) {
		ptrCtx.Next()

		// Buscar errores de base de datos
		for _, ptrGinError := range ptrCtx.Errors {
			if isDatabaseError(ptrGinError.Err) {
				handleDatabaseError(ptrCtx, ptrGinError.Err, ptrLogger, ptrMetrics)
				return
			}
		}
	}
}

// BusinessLogicErrorMiddleware middleware para errores de lógica de negocio
func BusinessLogicErrorMiddleware(ptrLogger logger.interfaceLogger, ptrMetrics metrics.interfaceManager) gin.HandlerFunc {
	return func(ptrCtx *gin.Context) {
		ptrCtx.Next()

		// Buscar errores de lógica de negocio
		for _, ptrGinError := range ptrCtx.Errors {
			if isBusinessLogicError(ptrGinError.Err) {
				handleBusinessLogicError(ptrCtx, ptrGinError.Err, ptrLogger, ptrMetrics)
				return
			}
		}
	}
}

// handleErrors maneja todos los errores acumulados
func handleErrors(ptrCtx *gin.Context, ptrLogger logger.interfaceLogger, ptrMetrics metrics.interfaceManager, ptrConfig *structErrorHandlerConfig) {
	if ptrCtx.Writer.Written() {
		// Ya se escribió una respuesta
		return
	}

	// Obtener el primer error (más importante)
	ptrMainError := ptrCtx.Errors[0].Err
	strRequestID := ptrCtx.GetString("request_id")

	// Determinar tipo de error y manejar apropiadamente
	ptrResponse := determineErrorResponse(ptrMainError, ptrCtx, ptrLogger, ptrConfig)
	ptrResponse.StrRequestID = strRequestID
	ptrResponse.TimeTimestamp = time.Now()

	// Log del error
	logError(ptrMainError, ptrCtx, ptrLogger, ptrConfig)

	// Registrar métricas
	recordErrorMetrics(ptrMainError, ptrCtx, ptrMetrics)

	// Enviar respuesta
	intStatusCode := determineStatusCode(ptrMainError)
	ptrCtx.JSON(intStatusCode, ptrResponse)
	ptrCtx.Abort()
}

// handlePanic maneja panics del sistema
func handlePanic(ptrCtx *gin.Context, objRecover interface{}, ptrLogger logger.interfaceLogger, ptrMetrics metrics.interfaceManager, ptrConfig *structErrorHandlerConfig) {
	strRequestID := ptrCtx.GetString("request_id")
	arrStackTrace := debug.Stack()

	// Crear error de panic
	ptrPanicError := errors.New(fmt.Sprintf("Panic recuperado: %v", objRecover)).
		WithDetail("panic_value", objRecover).
		WithDetail("stack_trace", string(arrStackTrace)).
		WithRequestID(strRequestID)

	// Log crítico del panic
	ptrLogger.Error("PANIC RECUPERADO",
		zap.String("request_id", strRequestID),
		zap.String("method", ptrCtx.Request.Method),
		zap.String("path", ptrCtx.Request.URL.Path),
		zap.Any("panic", objRecover),
		zap.String("stack_trace", string(arrStackTrace)),
	)

	// Registrar métricas de panic
	if ptrMetrics != nil {
		ptrMetrics.RecordBusinessError("system", "panic")
	}

	// Crear respuesta de error
	ptrResponse := &structErrorResponse{
		BoolExito:     false,
		StrMensaje:    "Error interno del servidor",
		StrCodigo:     "INTERNAL_PANIC",
		StrTipo:       "internal",
		StrRequestID:  strRequestID,
		TimeTimestamp: time.Now(),
	}

	// Incluir stack trace en desarrollo
	if ptrConfig.StrEnvironment == "development" || ptrConfig.BoolIncludeStackTrace {
		ptrResponse.StrStackTrace = string(arrStackTrace)
		ptrResponse.ObjDetalles = map[string]interface{}{
			"panic_value": objRecover,
		}
	}

	ptrCtx.JSON(http.StatusInternalServerError, ptrResponse)
	ptrCtx.Abort()
}

// handleValidationError maneja errores de validación
func handleValidationError(ptrCtx *gin.Context, err error, ptrLogger logger.interfaceLogger) {
	strRequestID := ptrCtx.GetString("request_id")

	ptrLogger.Warn("Error de validación",
		zap.String("request_id", strRequestID),
		zap.Error(err),
	)

	ptrResponse := &structErrorResponse{
		BoolExito:     false,
		StrMensaje:    "Datos de entrada inválidos",
		StrCodigo:     "VALIDATION_ERROR",
		StrTipo:       "validation",
		StrRequestID:  strRequestID,
		TimeTimestamp: time.Now(),
		ObjDetalles:   extractValidationDetails(err),
	}

	ptrCtx.JSON(http.StatusBadRequest, ptrResponse)
	ptrCtx.Abort()
}

// handleDatabaseError maneja errores de base de datos
func handleDatabaseError(ptrCtx *gin.Context, err error, ptrLogger logger.interfaceLogger, ptrMetrics metrics.interfaceManager) {
	strRequestID := ptrCtx.GetString("request_id")

	ptrLogger.Error("Error de base de datos",
		zap.String("request_id", strRequestID),
		zap.Error(err),
	)

	if ptrMetrics != nil {
		ptrMetrics.RecordBusinessError("database", "query_error")
	}

	ptrResponse := &structErrorResponse{
		BoolExito:     false,
		StrMensaje:    "Error en el acceso a datos",
		StrCodigo:     "DATABASE_ERROR",
		StrTipo:       "database",
		StrRequestID:  strRequestID,
		TimeTimestamp: time.Now(),
	}

	ptrCtx.JSON(http.StatusServiceUnavailable, ptrResponse)
	ptrCtx.Abort()
}

// handleBusinessLogicError maneja errores de lógica de negocio
func handleBusinessLogicError(ptrCtx *gin.Context, err error, ptrLogger logger.interfaceLogger, ptrMetrics metrics.interfaceManager) {
	strRequestID := ptrCtx.GetString("request_id")

	ptrLogger.Warn("Error de lógica de negocio",
		zap.String("request_id", strRequestID),
		zap.Error(err),
	)

	if ptrMetrics != nil {
		ptrMetrics.RecordBusinessError("business_logic", "validation_failed")
	}

	ptrResponse := &structErrorResponse{
		BoolExito:     false,
		StrMensaje:    "Error en la lógica de negocio",
		StrCodigo:     "BUSINESS_ERROR",
		StrTipo:       "business",
		StrRequestID:  strRequestID,
		TimeTimestamp: time.Now(),
		ObjDetalles:   err.Error(),
	}

	ptrCtx.JSON(http.StatusUnprocessableEntity, ptrResponse)
	ptrCtx.Abort()
}

// determineErrorResponse determina la respuesta apropiada para un error
func determineErrorResponse(err error, ptrCtx *gin.Context, ptrLogger logger.interfaceLogger, ptrConfig *structErrorHandlerConfig) *structErrorResponse {
	// Verificar si es un error de aplicación
	if ptrAppError, boolOk := err.(errors.interfaceAppError); boolOk {
		return createAppErrorResponse(ptrAppError, ptrConfig)
	}

	// Verificar handlers personalizados
	strErrorType := fmt.Sprintf("%T", err)
	if funcHandler, boolExists := ptrConfig.MapCustomHandlers[strErrorType]; boolExists {
		if ptrCustomResponse := funcHandler(ptrCtx, err, ptrLogger); ptrCustomResponse != nil {
			return convertToErrorResponse(ptrCustomResponse)
		}
	}

	// Error genérico
	return &structErrorResponse{
		BoolExito:  false,
		StrMensaje: "Error interno del servidor",
		StrCodigo:  "INTERNAL_ERROR",
		StrTipo:    "internal",
		ObjDetalles: func() interface{} {
			if ptrConfig.BoolDetailedErrors {
				return err.Error()
			}
			return nil
		}(),
	}
}

// createAppErrorResponse crea respuesta para errores de aplicación
func createAppErrorResponse(ptrAppError errors.interfaceAppError, ptrConfig *structErrorHandlerConfig) *structErrorResponse {
	ptrResponse := &structErrorResponse{
		BoolExito:   false,
		StrMensaje:  ptrAppError.Error(),
		StrCodigo:   ptrAppError.GetCode(),
		StrTipo:     string(ptrAppError.GetType()),
		ObjDetalles: ptrAppError.GetDetails(),
	}

	if ptrConfig.BoolIncludeStackTrace {
		ptrResponse.StrStackTrace = ptrAppError.GetStackTrace()
	}

	return ptrResponse
}

// convertToErrorResponse convierte respuesta de API a respuesta de error
func convertToErrorResponse(ptrAPIResponse *models.structRespuestaAPI) *structErrorResponse {
	return &structErrorResponse{
		BoolExito:     ptrAPIResponse.BoolExito,
		StrMensaje:    ptrAPIResponse.StrMensaje,
		ObjDetalles:   ptrAPIResponse.ObjDatos,
		StrRequestID:  ptrAPIResponse.StrRequestID,
		TimeTimestamp: ptrAPIResponse.TimeTimestamp,
	}
}

// logError registra el error en los logs
func logError(err error, ptrCtx *gin.Context, ptrLogger logger.interfaceLogger, ptrConfig *structErrorHandlerConfig) {
	if !ptrConfig.BoolLogAllErrors {
		return
	}

	arrFields := []zap.Field{
		zap.String("method", ptrCtx.Request.Method),
		zap.String("path", ptrCtx.Request.URL.Path),
		zap.Error(err),
	}

	if ptrAppError, boolOk := err.(errors.interfaceAppError); boolOk {
		arrFields = append(arrFields, ptrAppError.ToZapFields()...)
	}

	// Determinar nivel de log según tipo de error
	if ptrAppError, boolOk := err.(errors.interfaceAppError); boolOk {
		switch ptrAppError.GetType() {
		case errors.EnumErrorTypeValidation, errors.EnumErrorTypeBusiness:
			ptrLogger.Warn("Error de aplicación", arrFields...)
		case errors.EnumErrorTypeAuth, errors.EnumErrorTypeAuthorization:
			ptrLogger.Warn("Error de autenticación/autorización", arrFields...)
		default:
			ptrLogger.Error("Error del sistema", arrFields...)
		}
	} else {
		ptrLogger.Error("Error no categorizado", arrFields...)
	}
}

// recordErrorMetrics registra métricas de errores
func recordErrorMetrics(err error, ptrCtx *gin.Context, ptrMetrics metrics.interfaceManager) {
	if ptrMetrics == nil {
		return
	}

	strOperation := fmt.Sprintf("%s %s", ptrCtx.Request.Method, ptrCtx.Request.URL.Path)

	if ptrAppError, boolOk := err.(errors.interfaceAppError); boolOk {
		ptrMetrics.RecordBusinessError(strOperation, string(ptrAppError.GetType()))
	} else {
		ptrMetrics.RecordBusinessError(strOperation, "unknown")
	}
}

// determineStatusCode determina el código de estado HTTP
func determineStatusCode(err error) int {
	if ptrAppError, boolOk := err.(errors.interfaceAppError); boolOk {
		return errors.GetHTTPStatus(ptrAppError)
	}

	// Códigos por defecto para errores comunes
	strErrorMsg := strings.ToLower(err.Error())

	if strings.Contains(strErrorMsg, "not found") {
		return http.StatusNotFound
	}

	if strings.Contains(strErrorMsg, "unauthorized") {
		return http.StatusUnauthorized
	}

	if strings.Contains(strErrorMsg, "forbidden") {
		return http.StatusForbidden
	}

	if strings.Contains(strErrorMsg, "timeout") {
		return http.StatusRequestTimeout
	}

	return http.StatusInternalServerError
}

// Funciones de utilidad para detectar tipos de errores

// isDatabaseError verifica si es un error de base de datos
func isDatabaseError(err error) bool {
	strErrorMsg := strings.ToLower(err.Error())
	arrDBKeywords := []string{
		"sql", "database", "connection", "query", "transaction",
		"constraint", "foreign key", "unique", "duplicate",
	}

	for _, strKeyword := range arrDBKeywords {
		if strings.Contains(strErrorMsg, strKeyword) {
			return true
		}
	}

	return false
}

// isBusinessLogicError verifica si es un error de lógica de negocio
func isBusinessLogicError(err error) bool {
	if ptrAppError, boolOk := err.(errors.interfaceAppError); boolOk {
		return ptrAppError.GetType() == errors.EnumErrorTypeBusiness
	}
	return false
}

// extractValidationDetails extrae detalles de errores de validación
func extractValidationDetails(err error) interface{} {
	// Aquí se puede implementar lógica específica para extraer
	// detalles de diferentes tipos de errores de validación
	return map[string]interface{}{
		"error": err.Error(),
	}
}

// RegisterCustomErrorHandler registra un manejador personalizado de errores
func RegisterCustomErrorHandler(ptrConfig *structErrorHandlerConfig, strErrorType string, funcHandler funcErrorHandler) {
	if ptrConfig.MapCustomHandlers == nil {
		ptrConfig.MapCustomHandlers = make(map[string]funcErrorHandler)
	}
	ptrConfig.MapCustomHandlers[strErrorType] = funcHandler
}

// CreateErrorResponse crea una respuesta de error estándar
func CreateErrorResponse(strMensaje, strCodigo, strTipo string, objDetalles interface{}) *structErrorResponse {
	return &structErrorResponse{
		BoolExito:     false,
		StrMensaje:    strMensaje,
		StrCodigo:     strCodigo,
		StrTipo:       strTipo,
		ObjDetalles:   objDetalles,
		TimeTimestamp: time.Now(),
	}
}

// AbortWithError aborta la request con un error específico
func AbortWithError(ptrCtx *gin.Context, intStatusCode int, err error) {
	ptrCtx.Error(err)
	ptrCtx.Abort()
}
