// Package middleware proporciona middleware de validación avanzado
// con notación húngara y validaciones personalizadas
package middleware

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"ferre-pos-servidor-central/pkg/errors"
)

// structValidationConfig configuración del middleware de validación
type structValidationConfig struct {
	BoolStrictMode         bool                           `json:"strict_mode"`
	BoolValidateHeaders    bool                           `json:"validate_headers"`
	BoolValidateQuery      bool                           `json:"validate_query"`
	BoolValidateBody       bool                           `json:"validate_body"`
	IntMaxBodySize         int64                          `json:"max_body_size"`
	ArrRequiredHeaders     []string                       `json:"required_headers"`
	ArrAllowedContentTypes []string                       `json:"allowed_content_types"`
	BoolSanitizeInput      bool                           `json:"sanitize_input"`
	MapCustomValidators    map[string]funcCustomValidator `json:"-"`
}

// funcCustomValidator función para validaciones personalizadas
type funcCustomValidator func(interface{}) error

// structValidationContext contexto de validación
type structValidationContext struct {
	PtrValidator  validator.interfaceValidator
	PtrLogger     logger.interfaceLogger
	StrRequestID  string
	MapErrors     map[string][]string
	MutexErrors   sync.RWMutex
	BoolHasErrors bool
}

// structSanitizer sanitizador de datos
type structSanitizer struct {
	MapRules   map[string]funcSanitizeRule
	MutexRules sync.RWMutex
}

// funcSanitizeRule función para reglas de sanitización
type funcSanitizeRule func(interface{}) interface{}

// ValidationMiddleware middleware principal de validación
func ValidationMiddleware(ptrValidator validator.interfaceValidator, ptrLogger logger.interfaceLogger, ptrConfig *structValidationConfig) gin.HandlerFunc {
	if ptrConfig == nil {
		ptrConfig = &structValidationConfig{
			BoolStrictMode:         true,
			BoolValidateHeaders:    true,
			BoolValidateQuery:      true,
			BoolValidateBody:       true,
			IntMaxBodySize:         10 * 1024 * 1024, // 10MB
			ArrAllowedContentTypes: []string{"application/json", "application/xml", "text/plain"},
			BoolSanitizeInput:      true,
			MapCustomValidators:    make(map[string]funcCustomValidator),
		}
	}

	ptrSanitizer := NewSanitizer()

	return func(ptrCtx *gin.Context) {
		// Crear contexto de validación
		ptrValidationCtx := createValidationContext(ptrCtx, ptrValidator, ptrLogger)

		// Establecer contexto en Gin
		ptrCtx.Set("validation_context", ptrValidationCtx)

		// Validar headers si está habilitado
		if ptrConfig.BoolValidateHeaders {
			validateHeaders(ptrCtx, ptrValidationCtx, ptrConfig)
		}

		// Validar query parameters si está habilitado
		if ptrConfig.BoolValidateQuery {
			validateQueryParameters(ptrCtx, ptrValidationCtx, ptrConfig)
		}

		// Validar body si está habilitado y es necesario
		if ptrConfig.BoolValidateBody && shouldValidateBody(ptrCtx) {
			validateRequestBody(ptrCtx, ptrValidationCtx, ptrConfig, ptrSanitizer)
		}

		// Verificar si hay errores de validación
		if ptrValidationCtx.BoolHasErrors {
			handleValidationErrors(ptrCtx, ptrValidationCtx)
			return
		}

		ptrCtx.Next()
	}
}

// StructValidationMiddleware middleware para validar estructuras específicas
func StructValidationMiddleware(objStructType interface{}) gin.HandlerFunc {
	return func(ptrCtx *gin.Context) {
		ptrValidationCtx := getValidationContext(ptrCtx)
		if ptrValidationCtx == nil {
			ptrCtx.Next()
			return
		}

		// Crear instancia del tipo de estructura
		ptrStructValue := reflect.New(reflect.TypeOf(objStructType)).Interface()

		// Bind JSON al struct
		if err := ptrCtx.ShouldBindJSON(ptrStructValue); err != nil {
			addValidationError(ptrValidationCtx, "body", "Error al parsear JSON: "+err.Error())
			handleValidationErrors(ptrCtx, ptrValidationCtx)
			return
		}

		// Validar estructura
		if err := ptrValidationCtx.PtrValidator.ValidateStruct(ptrStructValue); err != nil {
			if ptrAppError, boolOk := err.(errors.interfaceAppError); boolOk {
				if mapDetails := ptrAppError.GetDetails(); mapDetails != nil {
					if arrValidationErrors, boolOk := mapDetails["validation_errors"].([]interface{}); boolOk {
						for _, objError := range arrValidationErrors {
							if mapError, boolOk := objError.(map[string]interface{}); boolOk {
								strField := fmt.Sprintf("%v", mapError["field"])
								strMessage := fmt.Sprintf("%v", mapError["message"])
								addValidationError(ptrValidationCtx, strField, strMessage)
							}
						}
					}
				}
			} else {
				addValidationError(ptrValidationCtx, "struct", err.Error())
			}

			handleValidationErrors(ptrCtx, ptrValidationCtx)
			return
		}

		// Establecer estructura validada en el contexto
		ptrCtx.Set("validated_struct", ptrStructValue)
		ptrCtx.Next()
	}
}

// RequiredFieldsMiddleware middleware para validar campos requeridos
func RequiredFieldsMiddleware(arrRequiredFields []string) gin.HandlerFunc {
	return func(ptrCtx *gin.Context) {
		ptrValidationCtx := getValidationContext(ptrCtx)
		if ptrValidationCtx == nil {
			ptrCtx.Next()
			return
		}

		var mapData map[string]interface{}
		if err := ptrCtx.ShouldBindJSON(&mapData); err != nil {
			addValidationError(ptrValidationCtx, "body", "Error al parsear JSON")
			handleValidationErrors(ptrCtx, ptrValidationCtx)
			return
		}

		// Validar campos requeridos
		for _, strField := range arrRequiredFields {
			if objValue, boolExists := mapData[strField]; !boolExists || isEmptyValue(objValue) {
				addValidationError(ptrValidationCtx, strField, fmt.Sprintf("El campo '%s' es requerido", strField))
			}
		}

		if ptrValidationCtx.BoolHasErrors {
			handleValidationErrors(ptrCtx, ptrValidationCtx)
			return
		}

		ptrCtx.Set("validated_data", mapData)
		ptrCtx.Next()
	}
}

// UUIDValidationMiddleware middleware para validar UUIDs en parámetros
func UUIDValidationMiddleware(arrUUIDParams []string) gin.HandlerFunc {
	return func(ptrCtx *gin.Context) {
		ptrValidationCtx := getValidationContext(ptrCtx)
		if ptrValidationCtx == nil {
			ptrCtx.Next()
			return
		}

		for _, strParam := range arrUUIDParams {
			strValue := ptrCtx.Param(strParam)
			if strValue != "" {
				if _, err := uuid.Parse(strValue); err != nil {
					addValidationError(ptrValidationCtx, strParam, fmt.Sprintf("El parámetro '%s' debe ser un UUID válido", strParam))
				}
			}
		}

		if ptrValidationCtx.BoolHasErrors {
			handleValidationErrors(ptrCtx, ptrValidationCtx)
			return
		}

		ptrCtx.Next()
	}
}

// PaginationValidationMiddleware middleware para validar parámetros de paginación
func PaginationValidationMiddleware() gin.HandlerFunc {
	return func(ptrCtx *gin.Context) {
		ptrValidationCtx := getValidationContext(ptrCtx)
		if ptrValidationCtx == nil {
			ptrCtx.Next()
			return
		}

		// Validar página
		strPagina := ptrCtx.DefaultQuery("pagina", "1")
		intPagina, err := strconv.Atoi(strPagina)
		if err != nil || intPagina < 1 {
			addValidationError(ptrValidationCtx, "pagina", "La página debe ser un número entero mayor a 0")
		}

		// Validar tamaño
		strTamano := ptrCtx.DefaultQuery("tamano", "20")
		intTamano, err := strconv.Atoi(strTamano)
		if err != nil || intTamano < 1 || intTamano > 1000 {
			addValidationError(ptrValidationCtx, "tamano", "El tamaño debe ser un número entre 1 y 1000")
		}

		// Validar orden
		strOrden := ptrCtx.Query("orden")
		if strOrden != "" && strOrden != "asc" && strOrden != "desc" {
			addValidationError(ptrValidationCtx, "orden", "El orden debe ser 'asc' o 'desc'")
		}

		if ptrValidationCtx.BoolHasErrors {
			handleValidationErrors(ptrCtx, ptrValidationCtx)
			return
		}

		// Establecer paginación validada
		ptrPaginacion := &models.structPaginacion{
			IntPagina: intPagina,
			IntTamano: intTamano,
		}
		ptrPaginacion.CalcularOffset()

		ptrCtx.Set("pagination", ptrPaginacion)
		ptrCtx.Next()
	}
}

// BusinessRulesValidationMiddleware middleware para validaciones de reglas de negocio
func BusinessRulesValidationMiddleware(mapRules map[string]funcCustomValidator) gin.HandlerFunc {
	return func(ptrCtx *gin.Context) {
		ptrValidationCtx := getValidationContext(ptrCtx)
		if ptrValidationCtx == nil {
			ptrCtx.Next()
			return
		}

		var mapData map[string]interface{}
		if err := ptrCtx.ShouldBindJSON(&mapData); err != nil {
			addValidationError(ptrValidationCtx, "body", "Error al parsear JSON")
			handleValidationErrors(ptrCtx, ptrValidationCtx)
			return
		}

		// Aplicar reglas de negocio
		for strRuleName, funcRule := range mapRules {
			if err := funcRule(mapData); err != nil {
				addValidationError(ptrValidationCtx, strRuleName, err.Error())
			}
		}

		if ptrValidationCtx.BoolHasErrors {
			handleValidationErrors(ptrCtx, ptrValidationCtx)
			return
		}

		ptrCtx.Next()
	}
}

// Funciones helper

// createValidationContext crea el contexto de validación
func createValidationContext(ptrCtx *gin.Context, ptrValidator validator.interfaceValidator, ptrLogger logger.interfaceLogger) *structValidationContext {
	strRequestID := ptrCtx.GetString("request_id")

	return &structValidationContext{
		PtrValidator:  ptrValidator,
		PtrLogger:     ptrLogger.WithRequestID(strRequestID),
		StrRequestID:  strRequestID,
		MapErrors:     make(map[string][]string),
		BoolHasErrors: false,
	}
}

// getValidationContext obtiene el contexto de validación
func getValidationContext(ptrCtx *gin.Context) *structValidationContext {
	if objCtx, boolExists := ptrCtx.Get("validation_context"); boolExists {
		if ptrValidationCtx, boolOk := objCtx.(*structValidationContext); boolOk {
			return ptrValidationCtx
		}
	}
	return nil
}

// addValidationError agrega un error de validación
func addValidationError(ptrValidationCtx *structValidationContext, strField, strMessage string) {
	ptrValidationCtx.MutexErrors.Lock()
	defer ptrValidationCtx.MutexErrors.Unlock()

	ptrValidationCtx.MapErrors[strField] = append(ptrValidationCtx.MapErrors[strField], strMessage)
	ptrValidationCtx.BoolHasErrors = true
}

// validateHeaders valida headers de la request
func validateHeaders(ptrCtx *gin.Context, ptrValidationCtx *structValidationContext, ptrConfig *structValidationConfig) {
	// Validar headers requeridos
	for _, strHeader := range ptrConfig.ArrRequiredHeaders {
		strValue := ptrCtx.GetHeader(strHeader)
		if strValue == "" {
			addValidationError(ptrValidationCtx, "headers."+strHeader, fmt.Sprintf("Header '%s' es requerido", strHeader))
		}
	}

	// Validar Content-Type si hay body
	if ptrCtx.Request.ContentLength > 0 {
		strContentType := ptrCtx.GetHeader("Content-Type")
		if strContentType != "" && !isAllowedContentType(strContentType, ptrConfig.ArrAllowedContentTypes) {
			addValidationError(ptrValidationCtx, "headers.content-type", "Content-Type no permitido")
		}
	}

	// Validar Authorization si está presente
	strAuth := ptrCtx.GetHeader("Authorization")
	if strAuth != "" && !isValidAuthorizationHeader(strAuth) {
		addValidationError(ptrValidationCtx, "headers.authorization", "Header Authorization inválido")
	}
}

// validateQueryParameters valida parámetros de query
func validateQueryParameters(ptrCtx *gin.Context, ptrValidationCtx *structValidationContext, ptrConfig *structValidationConfig) {
	mapQuery := ptrCtx.Request.URL.Query()

	for strKey, arrValues := range mapQuery {
		if len(arrValues) > 0 {
			strValue := arrValues[0]

			// Validaciones específicas por parámetro
			switch strKey {
			case "limit", "offset", "page", "size":
				if _, err := strconv.Atoi(strValue); err != nil {
					addValidationError(ptrValidationCtx, "query."+strKey, fmt.Sprintf("Parámetro '%s' debe ser un número", strKey))
				}
			case "sort", "order":
				if strValue != "asc" && strValue != "desc" {
					addValidationError(ptrValidationCtx, "query."+strKey, fmt.Sprintf("Parámetro '%s' debe ser 'asc' o 'desc'", strKey))
				}
			}

			// Validar longitud
			if len(strValue) > 1000 {
				addValidationError(ptrValidationCtx, "query."+strKey, fmt.Sprintf("Parámetro '%s' es demasiado largo", strKey))
			}
		}
	}
}

// validateRequestBody valida el body de la request
func validateRequestBody(ptrCtx *gin.Context, ptrValidationCtx *structValidationContext, ptrConfig *structValidationConfig, ptrSanitizer *structSanitizer) {
	if ptrCtx.Request.ContentLength > ptrConfig.IntMaxBodySize {
		addValidationError(ptrValidationCtx, "body", "Body demasiado grande")
		return
	}

	// Leer body
	arrBody, err := io.ReadAll(ptrCtx.Request.Body)
	if err != nil {
		addValidationError(ptrValidationCtx, "body", "Error al leer body")
		return
	}

	// Restaurar body para uso posterior
	ptrCtx.Request.Body = io.NopCloser(strings.NewReader(string(arrBody)))

	// Validar JSON si es el content type
	strContentType := ptrCtx.GetHeader("Content-Type")
	if strings.Contains(strContentType, "application/json") {
		var objData interface{}
		if err := json.Unmarshal(arrBody, &objData); err != nil {
			addValidationError(ptrValidationCtx, "body", "JSON inválido: "+err.Error())
			return
		}

		// Sanitizar datos si está habilitado
		if ptrConfig.BoolSanitizeInput {
			objData = ptrSanitizer.Sanitize(objData)
			ptrCtx.Set("sanitized_data", objData)
		}
	}
}

// handleValidationErrors maneja errores de validación
func handleValidationErrors(ptrCtx *gin.Context, ptrValidationCtx *structValidationContext) {
	ptrValidationCtx.MutexErrors.RLock()
	mapErrors := make(map[string][]string)
	for strKey, arrValues := range ptrValidationCtx.MapErrors {
		mapErrors[strKey] = arrValues
	}
	ptrValidationCtx.MutexErrors.RUnlock()

	// Log de errores de validación
	ptrValidationCtx.PtrLogger.Warn("Errores de validación detectados",
		zap.String("request_id", ptrValidationCtx.StrRequestID),
		zap.Any("errors", mapErrors),
	)

	// Crear respuesta de error
	ptrAppError := errors.NewValidation("Errores de validación en los datos de entrada", "VALIDATION_ERROR").
		WithDetail("field_errors", mapErrors).
		WithRequestID(ptrValidationCtx.StrRequestID)

	ptrCtx.JSON(http.StatusBadRequest, gin.H{
		"exito":      false,
		"mensaje":    "Errores de validación",
		"codigo":     "VALIDATION_ERROR",
		"errores":    mapErrors,
		"request_id": ptrValidationCtx.StrRequestID,
		"timestamp":  time.Now().Format(time.RFC3339),
	})

	ptrCtx.Error(ptrAppError)
	ptrCtx.Abort()
}

// Funciones de utilidad

// shouldValidateBody verifica si debe validar el body
func shouldValidateBody(ptrCtx *gin.Context) bool {
	return ptrCtx.Request.Method == "POST" ||
		ptrCtx.Request.Method == "PUT" ||
		ptrCtx.Request.Method == "PATCH"
}

// isAllowedContentType verifica si el content type está permitido
func isAllowedContentType(strContentType string, arrAllowed []string) bool {
	for _, strAllowed := range arrAllowed {
		if strings.Contains(strContentType, strAllowed) {
			return true
		}
	}
	return false
}

// isValidAuthorizationHeader valida el header de autorización
func isValidAuthorizationHeader(strAuth string) bool {
	arrParts := strings.Split(strAuth, " ")
	if len(arrParts) != 2 {
		return false
	}

	strScheme := strings.ToLower(arrParts[0])
	return strScheme == "bearer" || strScheme == "basic"
}

// isEmptyValue verifica si un valor está vacío
func isEmptyValue(objValue interface{}) bool {
	if objValue == nil {
		return true
	}

	switch objTypedValue := objValue.(type) {
	case string:
		return strings.TrimSpace(objTypedValue) == ""
	case []interface{}:
		return len(objTypedValue) == 0
	case map[string]interface{}:
		return len(objTypedValue) == 0
	default:
		return false
	}
}

// Sanitizer

// NewSanitizer crea un nuevo sanitizador
func NewSanitizer() *structSanitizer {
	ptrSanitizer := &structSanitizer{
		MapRules: make(map[string]funcSanitizeRule),
	}

	// Reglas por defecto
	ptrSanitizer.AddRule("trim", func(objValue interface{}) interface{} {
		if strValue, boolOk := objValue.(string); boolOk {
			return strings.TrimSpace(strValue)
		}
		return objValue
	})

	ptrSanitizer.AddRule("lowercase", func(objValue interface{}) interface{} {
		if strValue, boolOk := objValue.(string); boolOk {
			return strings.ToLower(strValue)
		}
		return objValue
	})

	ptrSanitizer.AddRule("remove_html", func(objValue interface{}) interface{} {
		if strValue, boolOk := objValue.(string); boolOk {
			// Implementación simple de remoción de HTML
			return removeHTMLTags(strValue)
		}
		return objValue
	})

	return ptrSanitizer
}

// AddRule agrega una regla de sanitización
func (ptrSanitizer *structSanitizer) AddRule(strName string, funcRule funcSanitizeRule) {
	ptrSanitizer.MutexRules.Lock()
	defer ptrSanitizer.MutexRules.Unlock()
	ptrSanitizer.MapRules[strName] = funcRule
}

// Sanitize sanitiza datos aplicando todas las reglas
func (ptrSanitizer *structSanitizer) Sanitize(objData interface{}) interface{} {
	ptrSanitizer.MutexRules.RLock()
	defer ptrSanitizer.MutexRules.RUnlock()

	return ptrSanitizer.sanitizeValue(objData)
}

// sanitizeValue sanitiza un valor específico
func (ptrSanitizer *structSanitizer) sanitizeValue(objValue interface{}) interface{} {
	switch objTypedValue := objValue.(type) {
	case string:
		objResult := objTypedValue
		for _, funcRule := range ptrSanitizer.MapRules {
			objResult = funcRule(objResult).(string)
		}
		return objResult

	case map[string]interface{}:
		mapResult := make(map[string]interface{})
		for strKey, objVal := range objTypedValue {
			mapResult[strKey] = ptrSanitizer.sanitizeValue(objVal)
		}
		return mapResult

	case []interface{}:
		arrResult := make([]interface{}, len(objTypedValue))
		for intI, objVal := range objTypedValue {
			arrResult[intI] = ptrSanitizer.sanitizeValue(objVal)
		}
		return arrResult

	default:
		return objValue
	}
}

// removeHTMLTags remueve tags HTML básicos
func removeHTMLTags(strInput string) string {
	// Implementación simple - en producción usar una librería como bluemonday
	strResult := strInput
	strResult = strings.ReplaceAll(strResult, "<script>", "")
	strResult = strings.ReplaceAll(strResult, "</script>", "")
	strResult = strings.ReplaceAll(strResult, "<style>", "")
	strResult = strings.ReplaceAll(strResult, "</style>", "")
	return strResult
}

// GetValidatedStruct obtiene la estructura validada del contexto
func GetValidatedStruct(ptrCtx *gin.Context) interface{} {
	if objStruct, boolExists := ptrCtx.Get("validated_struct"); boolExists {
		return objStruct
	}
	return nil
}

// GetValidatedData obtiene los datos validados del contexto
func GetValidatedData(ptrCtx *gin.Context) map[string]interface{} {
	if objData, boolExists := ptrCtx.Get("validated_data"); boolExists {
		if mapData, boolOk := objData.(map[string]interface{}); boolOk {
			return mapData
		}
	}
	return nil
}

// GetSanitizedData obtiene los datos sanitizados del contexto
func GetSanitizedData(ptrCtx *gin.Context) interface{} {
	if objData, boolExists := ptrCtx.Get("sanitized_data"); boolExists {
		return objData
	}
	return nil
}
