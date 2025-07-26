// Package errors proporciona funcionalidades mejoradas de manejo de errores
// con notación húngara y wrapping de errores
package errors

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"go.uber.org/zap"
)

// enumErrorType define los tipos de errores del sistema
type enumErrorType string

const (
	// Tipos de errores con notación húngara
	EnumErrorTypeValidation    enumErrorType = "validation"
	EnumErrorTypeBusiness      enumErrorType = "business"
	EnumErrorTypeDatabase      enumErrorType = "database"
	EnumErrorTypeNetwork       enumErrorType = "network"
	EnumErrorTypeAuth          enumErrorType = "authentication"
	EnumErrorTypeAuthorization enumErrorType = "authorization"
	EnumErrorTypeInternal      enumErrorType = "internal"
	EnumErrorTypeExternal      enumErrorType = "external"
	EnumErrorTypeTimeout       enumErrorType = "timeout"
	EnumErrorTypeNotFound      enumErrorType = "not_found"
	EnumErrorTypeConflict      enumErrorType = "conflict"
	EnumErrorTypeRateLimit     enumErrorType = "rate_limit"
)

// structAppError representa un error de aplicación con contexto
type structAppError struct {
	StrMessage    string                 `json:"message"`
	StrCode       string                 `json:"code"`
	EnumType      enumErrorType          `json:"type"`
	ErrOriginal   error                  `json:"-"`
	MapDetails    map[string]interface{} `json:"details,omitempty"`
	StrStackTrace string                 `json:"stack_trace,omitempty"`
	TimeOccurred  time.Time              `json:"occurred_at"`
	StrRequestID  string                 `json:"request_id,omitempty"`
	StrUserID     string                 `json:"user_id,omitempty"`
}

// interfaceAppError define la interfaz para errores de aplicación
type interfaceAppError interface {
	error
	GetCode() string
	GetType() enumErrorType
	GetDetails() map[string]interface{}
	GetOriginal() error
	GetStackTrace() string
	GetRequestID() string
	GetUserID() string
	WithDetail(strKey string, value interface{}) interfaceAppError
	WithRequestID(strRequestID string) interfaceAppError
	WithUserID(strUserID string) interfaceAppError
	ToZapFields() []zap.Field
}

// Error implementa la interfaz error
func (ptrErr *structAppError) Error() string {
	if ptrErr.ErrOriginal != nil {
		return fmt.Sprintf("%s: %v", ptrErr.StrMessage, ptrErr.ErrOriginal)
	}
	return ptrErr.StrMessage
}

// GetCode retorna el código del error
func (ptrErr *structAppError) GetCode() string {
	return ptrErr.StrCode
}

// GetType retorna el tipo del error
func (ptrErr *structAppError) GetType() enumErrorType {
	return ptrErr.EnumType
}

// GetDetails retorna los detalles del error
func (ptrErr *structAppError) GetDetails() map[string]interface{} {
	return ptrErr.MapDetails
}

// GetOriginal retorna el error original
func (ptrErr *structAppError) GetOriginal() error {
	return ptrErr.ErrOriginal
}

// GetStackTrace retorna el stack trace
func (ptrErr *structAppError) GetStackTrace() string {
	return ptrErr.StrStackTrace
}

// GetRequestID retorna el ID de la request
func (ptrErr *structAppError) GetRequestID() string {
	return ptrErr.StrRequestID
}

// GetUserID retorna el ID del usuario
func (ptrErr *structAppError) GetUserID() string {
	return ptrErr.StrUserID
}

// WithDetail agrega un detalle al error
func (ptrErr *structAppError) WithDetail(strKey string, value interface{}) interfaceAppError {
	if ptrErr.MapDetails == nil {
		ptrErr.MapDetails = make(map[string]interface{})
	}
	ptrErr.MapDetails[strKey] = value
	return ptrErr
}

// WithRequestID agrega el ID de la request
func (ptrErr *structAppError) WithRequestID(strRequestID string) interfaceAppError {
	ptrErr.StrRequestID = strRequestID
	return ptrErr
}

// WithUserID agrega el ID del usuario
func (ptrErr *structAppError) WithUserID(strUserID string) interfaceAppError {
	ptrErr.StrUserID = strUserID
	return ptrErr
}

// ToZapFields convierte el error a campos de Zap
func (ptrErr *structAppError) ToZapFields() []zap.Field {
	arrFields := []zap.Field{
		zap.String("error_code", ptrErr.StrCode),
		zap.String("error_type", string(ptrErr.EnumType)),
		zap.String("error_message", ptrErr.StrMessage),
		zap.Time("occurred_at", ptrErr.TimeOccurred),
	}

	if ptrErr.StrRequestID != "" {
		arrFields = append(arrFields, zap.String("request_id", ptrErr.StrRequestID))
	}

	if ptrErr.StrUserID != "" {
		arrFields = append(arrFields, zap.String("user_id", ptrErr.StrUserID))
	}

	if ptrErr.ErrOriginal != nil {
		arrFields = append(arrFields, zap.Error(ptrErr.ErrOriginal))
	}

	if len(ptrErr.MapDetails) > 0 {
		arrFields = append(arrFields, zap.Any("details", ptrErr.MapDetails))
	}

	return arrFields
}

// New crea un nuevo error de aplicación
func New(strMessage string) interfaceAppError {
	return &structAppError{
		StrMessage:    strMessage,
		StrCode:       "INTERNAL_ERROR",
		EnumType:      EnumErrorTypeInternal,
		TimeOccurred:  time.Now(),
		StrStackTrace: getStackTrace(),
	}
}

// Newf crea un nuevo error de aplicación con formato
func Newf(strFormat string, arrArgs ...interface{}) interfaceAppError {
	return New(fmt.Sprintf(strFormat, arrArgs...))
}

// Wrap envuelve un error existente
func Wrap(err error, strMessage string) interfaceAppError {
	if err == nil {
		return nil
	}

	// Si ya es un AppError, mantener la información original
	if ptrAppErr, boolOk := err.(interfaceAppError); boolOk {
		return &structAppError{
			StrMessage:    strMessage,
			StrCode:       ptrAppErr.GetCode(),
			EnumType:      ptrAppErr.GetType(),
			ErrOriginal:   ptrAppErr.GetOriginal(),
			MapDetails:    ptrAppErr.GetDetails(),
			StrStackTrace: ptrAppErr.GetStackTrace(),
			TimeOccurred:  time.Now(),
			StrRequestID:  ptrAppErr.GetRequestID(),
			StrUserID:     ptrAppErr.GetUserID(),
		}
	}

	return &structAppError{
		StrMessage:    strMessage,
		StrCode:       "INTERNAL_ERROR",
		EnumType:      EnumErrorTypeInternal,
		ErrOriginal:   err,
		TimeOccurred:  time.Now(),
		StrStackTrace: getStackTrace(),
	}
}

// Wrapf envuelve un error existente con formato
func Wrapf(err error, strFormat string, arrArgs ...interface{}) interfaceAppError {
	return Wrap(err, fmt.Sprintf(strFormat, arrArgs...))
}

// NewValidation crea un error de validación
func NewValidation(strMessage string, strCode string) interfaceAppError {
	return &structAppError{
		StrMessage:    strMessage,
		StrCode:       strCode,
		EnumType:      EnumErrorTypeValidation,
		TimeOccurred:  time.Now(),
		StrStackTrace: getStackTrace(),
	}
}

// NewBusiness crea un error de negocio
func NewBusiness(strMessage string, strCode string) interfaceAppError {
	return &structAppError{
		StrMessage:    strMessage,
		StrCode:       strCode,
		EnumType:      EnumErrorTypeBusiness,
		TimeOccurred:  time.Now(),
		StrStackTrace: getStackTrace(),
	}
}

// NewDatabase crea un error de base de datos
func NewDatabase(strMessage string, strCode string) interfaceAppError {
	return &structAppError{
		StrMessage:    strMessage,
		StrCode:       strCode,
		EnumType:      EnumErrorTypeDatabase,
		TimeOccurred:  time.Now(),
		StrStackTrace: getStackTrace(),
	}
}

// NewAuth crea un error de autenticación
func NewAuth(strMessage string, strCode string) interfaceAppError {
	return &structAppError{
		StrMessage:    strMessage,
		StrCode:       strCode,
		EnumType:      EnumErrorTypeAuth,
		TimeOccurred:  time.Now(),
		StrStackTrace: getStackTrace(),
	}
}

// NewAuthorization crea un error de autorización
func NewAuthorization(strMessage string, strCode string) interfaceAppError {
	return &structAppError{
		StrMessage:    strMessage,
		StrCode:       strCode,
		EnumType:      EnumErrorTypeAuthorization,
		TimeOccurred:  time.Now(),
		StrStackTrace: getStackTrace(),
	}
}

// NewNotFound crea un error de recurso no encontrado
func NewNotFound(strMessage string, strCode string) interfaceAppError {
	return &structAppError{
		StrMessage:    strMessage,
		StrCode:       strCode,
		EnumType:      EnumErrorTypeNotFound,
		TimeOccurred:  time.Now(),
		StrStackTrace: getStackTrace(),
	}
}

// NewConflict crea un error de conflicto
func NewConflict(strMessage string, strCode string) interfaceAppError {
	return &structAppError{
		StrMessage:    strMessage,
		StrCode:       strCode,
		EnumType:      EnumErrorTypeConflict,
		TimeOccurred:  time.Now(),
		StrStackTrace: getStackTrace(),
	}
}

// NewTimeout crea un error de timeout
func NewTimeout(strMessage string, strCode string) interfaceAppError {
	return &structAppError{
		StrMessage:    strMessage,
		StrCode:       strCode,
		EnumType:      EnumErrorTypeTimeout,
		TimeOccurred:  time.Now(),
		StrStackTrace: getStackTrace(),
	}
}

// NewRateLimit crea un error de rate limiting
func NewRateLimit(strMessage string, strCode string) interfaceAppError {
	return &structAppError{
		StrMessage:    strMessage,
		StrCode:       strCode,
		EnumType:      EnumErrorTypeRateLimit,
		TimeOccurred:  time.Now(),
		StrStackTrace: getStackTrace(),
	}
}

// IsType verifica si un error es de un tipo específico
func IsType(err error, enumType enumErrorType) bool {
	if ptrAppErr, boolOk := err.(interfaceAppError); boolOk {
		return ptrAppErr.GetType() == enumType
	}
	return false
}

// IsCode verifica si un error tiene un código específico
func IsCode(err error, strCode string) bool {
	if ptrAppErr, boolOk := err.(interfaceAppError); boolOk {
		return ptrAppErr.GetCode() == strCode
	}
	return false
}

// GetHTTPStatus retorna el código HTTP apropiado para el tipo de error
func GetHTTPStatus(err error) int {
	if ptrAppErr, boolOk := err.(interfaceAppError); boolOk {
		switch ptrAppErr.GetType() {
		case EnumErrorTypeValidation:
			return 400 // Bad Request
		case EnumErrorTypeAuth:
			return 401 // Unauthorized
		case EnumErrorTypeAuthorization:
			return 403 // Forbidden
		case EnumErrorTypeNotFound:
			return 404 // Not Found
		case EnumErrorTypeConflict:
			return 409 // Conflict
		case EnumErrorTypeBusiness:
			return 422 // Unprocessable Entity
		case EnumErrorTypeRateLimit:
			return 429 // Too Many Requests
		case EnumErrorTypeTimeout:
			return 408 // Request Timeout
		case EnumErrorTypeDatabase, EnumErrorTypeNetwork, EnumErrorTypeExternal:
			return 503 // Service Unavailable
		default:
			return 500 // Internal Server Error
		}
	}
	return 500 // Internal Server Error
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

// structErrorCollector recolecta múltiples errores
type structErrorCollector struct {
	arrErrors []interfaceAppError
	strPrefix string
}

// interfaceErrorCollector define la interfaz para recolectar errores
type interfaceErrorCollector interface {
	Add(err error)
	AddValidation(strField string, strMessage string)
	AddBusiness(strMessage string, strCode string)
	HasErrors() bool
	GetErrors() []interfaceAppError
	GetFirst() interfaceAppError
	ToError() error
	Count() int
}

// NewErrorCollector crea un nuevo recolector de errores
func NewErrorCollector(strPrefix string) interfaceErrorCollector {
	return &structErrorCollector{
		arrErrors: make([]interfaceAppError, 0),
		strPrefix: strPrefix,
	}
}

// Add agrega un error al recolector
func (ptrCollector *structErrorCollector) Add(err error) {
	if err == nil {
		return
	}

	if ptrAppErr, boolOk := err.(interfaceAppError); boolOk {
		ptrCollector.arrErrors = append(ptrCollector.arrErrors, ptrAppErr)
	} else {
		ptrCollector.arrErrors = append(ptrCollector.arrErrors, Wrap(err, ptrCollector.strPrefix))
	}
}

// AddValidation agrega un error de validación
func (ptrCollector *structErrorCollector) AddValidation(strField string, strMessage string) {
	ptrErr := NewValidation(strMessage, "VALIDATION_ERROR").
		WithDetail("field", strField)
	ptrCollector.arrErrors = append(ptrCollector.arrErrors, ptrErr)
}

// AddBusiness agrega un error de negocio
func (ptrCollector *structErrorCollector) AddBusiness(strMessage string, strCode string) {
	ptrErr := NewBusiness(strMessage, strCode)
	ptrCollector.arrErrors = append(ptrCollector.arrErrors, ptrErr)
}

// HasErrors verifica si hay errores
func (ptrCollector *structErrorCollector) HasErrors() bool {
	return len(ptrCollector.arrErrors) > 0
}

// GetErrors retorna todos los errores
func (ptrCollector *structErrorCollector) GetErrors() []interfaceAppError {
	return ptrCollector.arrErrors
}

// GetFirst retorna el primer error
func (ptrCollector *structErrorCollector) GetFirst() interfaceAppError {
	if len(ptrCollector.arrErrors) > 0 {
		return ptrCollector.arrErrors[0]
	}
	return nil
}

// ToError convierte a un error único
func (ptrCollector *structErrorCollector) ToError() error {
	if !ptrCollector.HasErrors() {
		return nil
	}

	if len(ptrCollector.arrErrors) == 1 {
		return ptrCollector.arrErrors[0]
	}

	// Crear un error compuesto
	var arrMessages []string
	for _, ptrErr := range ptrCollector.arrErrors {
		arrMessages = append(arrMessages, ptrErr.Error())
	}

	return New(fmt.Sprintf("%s: %s", ptrCollector.strPrefix, strings.Join(arrMessages, "; ")))
}

// Count retorna el número de errores
func (ptrCollector *structErrorCollector) Count() int {
	return len(ptrCollector.arrErrors)
}
