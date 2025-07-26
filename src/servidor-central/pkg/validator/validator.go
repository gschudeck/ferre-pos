// Package validator proporciona funcionalidades avanzadas de validación
// con notación húngara y validaciones personalizadas
package validator

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"

	"ferre-pos-servidor-central/pkg/errors"
)

// structValidator encapsula el validador con funcionalidades extendidas
type structValidator struct {
	ptrValidator        *validator.Validate
	ptrLogger           *zap.Logger
	mapCustomValidators map[string]validator.Func
}

// interfaceValidator define la interfaz del validador
type interfaceValidator interface {
	Validate(interface{}) error
	ValidateStruct(interface{}) error
	ValidateField(interface{}, string) error
	ValidateVar(interface{}, string) error
	RegisterValidation(string, validator.Func) error
	RegisterCustomValidations() error
	GetValidator() *validator.Validate
}

// structValidationError representa un error de validación específico
type structValidationError struct {
	strField   string `json:"field"`
	strTag     string `json:"tag"`
	strValue   string `json:"value"`
	strMessage string `json:"message"`
	strParam   string `json:"param,omitempty"`
}

// structValidationErrors representa múltiples errores de validación
type structValidationErrors struct {
	arrErrors []structValidationError `json:"errors"`
	strPrefix string                  `json:"prefix,omitempty"`
}

// Error implementa la interfaz error para ValidationErrors
func (ptrErrs *structValidationErrors) Error() string {
	var arrMessages []string
	for _, ptrErr := range ptrErrs.arrErrors {
		arrMessages = append(arrMessages, ptrErr.strMessage)
	}

	if ptrErrs.strPrefix != "" {
		return fmt.Sprintf("%s: %s", ptrErrs.strPrefix, strings.Join(arrMessages, "; "))
	}

	return strings.Join(arrMessages, "; ")
}

// GetErrors retorna los errores de validación
func (ptrErrs *structValidationErrors) GetErrors() []structValidationError {
	return ptrErrs.arrErrors
}

// GetFieldErrors retorna errores por campo
func (ptrErrs *structValidationErrors) GetFieldErrors() map[string][]string {
	mapFieldErrors := make(map[string][]string)

	for _, ptrErr := range ptrErrs.arrErrors {
		mapFieldErrors[ptrErr.strField] = append(mapFieldErrors[ptrErr.strField], ptrErr.strMessage)
	}

	return mapFieldErrors
}

// New crea un nuevo validador
func New() *structValidator {
	ptrValidator := validator.New()

	// Configurar nombres de campos usando tags JSON
	ptrValidator.RegisterTagNameFunc(func(ptrField reflect.StructField) string {
		strName := strings.SplitN(ptrField.Tag.Get("json"), ",", 2)[0]
		if strName == "-" {
			return ""
		}
		return strName
	})

	ptrCustomValidator := &structValidator{
		ptrValidator:        ptrValidator,
		mapCustomValidators: make(map[string]validator.Func),
	}

	// Registrar validaciones personalizadas
	ptrCustomValidator.RegisterCustomValidations()

	return ptrCustomValidator
}

// NewWithLogger crea un nuevo validador con logger
func NewWithLogger(ptrLogger *zap.Logger) *structValidator {
	ptrValidator := New()
	ptrValidator.ptrLogger = ptrLogger
	return ptrValidator
}

// Validate valida una estructura y retorna errores formateados
func (ptrValidator *structValidator) Validate(objData interface{}) error {
	if err := ptrValidator.ptrValidator.Struct(objData); err != nil {
		return ptrValidator.formatValidationErrors(err)
	}
	return nil
}

// ValidateStruct valida una estructura
func (ptrValidator *structValidator) ValidateStruct(objStruct interface{}) error {
	return ptrValidator.Validate(objStruct)
}

// ValidateField valida un campo específico
func (ptrValidator *structValidator) ValidateField(objValue interface{}, strTag string) error {
	if err := ptrValidator.ptrValidator.Var(objValue, strTag); err != nil {
		return ptrValidator.formatValidationErrors(err)
	}
	return nil
}

// ValidateVar valida una variable
func (ptrValidator *structValidator) ValidateVar(objValue interface{}, strTag string) error {
	return ptrValidator.ValidateField(objValue, strTag)
}

// RegisterValidation registra una validación personalizada
func (ptrValidator *structValidator) RegisterValidation(strTag string, funcValidator validator.Func) error {
	ptrValidator.mapCustomValidators[strTag] = funcValidator
	return ptrValidator.ptrValidator.RegisterValidation(strTag, funcValidator)
}

// GetValidator retorna el validador subyacente
func (ptrValidator *structValidator) GetValidator() *validator.Validate {
	return ptrValidator.ptrValidator
}

// RegisterCustomValidations registra todas las validaciones personalizadas
func (ptrValidator *structValidator) RegisterCustomValidations() error {
	// Validación de RUT chileno
	ptrValidator.RegisterValidation("rut", validateRUT)

	// Validación de código de barras
	ptrValidator.RegisterValidation("barcode", validateBarcode)

	// Validación de SKU
	ptrValidator.RegisterValidation("sku", validateSKU)

	// Validación de precio
	ptrValidator.RegisterValidation("price", validatePrice)

	// Validación de stock
	ptrValidator.RegisterValidation("stock", validateStock)

	// Validación de porcentaje
	ptrValidator.RegisterValidation("percentage", validatePercentage)

	// Validación de teléfono chileno
	ptrValidator.RegisterValidation("phone_cl", validatePhoneCL)

	// Validación de fecha futura
	ptrValidator.RegisterValidation("future_date", validateFutureDate)

	// Validación de fecha pasada
	ptrValidator.RegisterValidation("past_date", validatePastDate)

	// Validación de moneda
	ptrValidator.RegisterValidation("currency", validateCurrency)

	// Validación de coordenadas
	ptrValidator.RegisterValidation("latitude", validateLatitude)
	ptrValidator.RegisterValidation("longitude", validateLongitude)

	// Validación de JSON válido
	ptrValidator.RegisterValidation("valid_json", validateJSON)

	// Validación de UUID v4
	ptrValidator.RegisterValidation("uuid4", validateUUIDv4)

	// Validación de color hexadecimal
	ptrValidator.RegisterValidation("hex_color", validateHexColor)

	// Validación de URL de imagen
	ptrValidator.RegisterValidation("image_url", validateImageURL)

	return nil
}

// formatValidationErrors formatea los errores de validación
func (ptrValidator *structValidator) formatValidationErrors(err error) error {
	var arrValidationErrors []structValidationError

	if ptrValidationErrors, boolOk := err.(validator.ValidationErrors); boolOk {
		for _, ptrFieldError := range ptrValidationErrors {
			ptrValidationError := structValidationError{
				strField: ptrFieldError.Field(),
				strTag:   ptrFieldError.Tag(),
				strValue: fmt.Sprintf("%v", ptrFieldError.Value()),
				strParam: ptrFieldError.Param(),
			}

			// Generar mensaje personalizado
			ptrValidationError.strMessage = ptrValidator.getCustomMessage(ptrFieldError)

			arrValidationErrors = append(arrValidationErrors, ptrValidationError)
		}
	}

	ptrFormattedErrors := &structValidationErrors{
		arrErrors: arrValidationErrors,
		strPrefix: "Errores de validación",
	}

	// Log del error si hay logger configurado
	if ptrValidator.ptrLogger != nil {
		ptrValidator.ptrLogger.Warn("Errores de validación detectados",
			zap.Any("errors", arrValidationErrors),
			zap.Error(err),
		)
	}

	return errors.NewValidation(ptrFormattedErrors.Error(), "VALIDATION_ERROR").
		WithDetail("validation_errors", arrValidationErrors)
}

// getCustomMessage genera mensajes personalizados para errores de validación
func (ptrValidator *structValidator) getCustomMessage(ptrFieldError validator.FieldError) string {
	strField := ptrFieldError.Field()
	strTag := ptrFieldError.Tag()
	strParam := ptrFieldError.Param()

	switch strTag {
	case "required":
		return fmt.Sprintf("El campo '%s' es requerido", strField)
	case "email":
		return fmt.Sprintf("El campo '%s' debe ser un email válido", strField)
	case "min":
		return fmt.Sprintf("El campo '%s' debe tener al menos %s caracteres", strField, strParam)
	case "max":
		return fmt.Sprintf("El campo '%s' debe tener máximo %s caracteres", strField, strParam)
	case "len":
		return fmt.Sprintf("El campo '%s' debe tener exactamente %s caracteres", strField, strParam)
	case "numeric":
		return fmt.Sprintf("El campo '%s' debe ser numérico", strField)
	case "alpha":
		return fmt.Sprintf("El campo '%s' debe contener solo letras", strField)
	case "alphanum":
		return fmt.Sprintf("El campo '%s' debe contener solo letras y números", strField)
	case "url":
		return fmt.Sprintf("El campo '%s' debe ser una URL válida", strField)
	case "uuid":
		return fmt.Sprintf("El campo '%s' debe ser un UUID válido", strField)
	case "uuid4":
		return fmt.Sprintf("El campo '%s' debe ser un UUID v4 válido", strField)
	case "rut":
		return fmt.Sprintf("El campo '%s' debe ser un RUT chileno válido", strField)
	case "barcode":
		return fmt.Sprintf("El campo '%s' debe ser un código de barras válido", strField)
	case "sku":
		return fmt.Sprintf("El campo '%s' debe ser un SKU válido", strField)
	case "price":
		return fmt.Sprintf("El campo '%s' debe ser un precio válido", strField)
	case "stock":
		return fmt.Sprintf("El campo '%s' debe ser una cantidad de stock válida", strField)
	case "percentage":
		return fmt.Sprintf("El campo '%s' debe ser un porcentaje válido (0-100)", strField)
	case "phone_cl":
		return fmt.Sprintf("El campo '%s' debe ser un teléfono chileno válido", strField)
	case "future_date":
		return fmt.Sprintf("El campo '%s' debe ser una fecha futura", strField)
	case "past_date":
		return fmt.Sprintf("El campo '%s' debe ser una fecha pasada", strField)
	case "currency":
		return fmt.Sprintf("El campo '%s' debe ser un código de moneda válido", strField)
	case "latitude":
		return fmt.Sprintf("El campo '%s' debe ser una latitud válida (-90 a 90)", strField)
	case "longitude":
		return fmt.Sprintf("El campo '%s' debe ser una longitud válida (-180 a 180)", strField)
	case "valid_json":
		return fmt.Sprintf("El campo '%s' debe contener JSON válido", strField)
	case "hex_color":
		return fmt.Sprintf("El campo '%s' debe ser un color hexadecimal válido", strField)
	case "image_url":
		return fmt.Sprintf("El campo '%s' debe ser una URL de imagen válida", strField)
	case "gte":
		return fmt.Sprintf("El campo '%s' debe ser mayor o igual a %s", strField, strParam)
	case "lte":
		return fmt.Sprintf("El campo '%s' debe ser menor o igual a %s", strField, strParam)
	case "gt":
		return fmt.Sprintf("El campo '%s' debe ser mayor a %s", strField, strParam)
	case "lt":
		return fmt.Sprintf("El campo '%s' debe ser menor a %s", strField, strParam)
	case "oneof":
		return fmt.Sprintf("El campo '%s' debe ser uno de: %s", strField, strParam)
	default:
		return fmt.Sprintf("El campo '%s' no cumple con la validación '%s'", strField, strTag)
	}
}

// Validaciones personalizadas

// validateRUT valida un RUT chileno
func validateRUT(ptrFieldLevel validator.FieldLevel) bool {
	strRUT := ptrFieldLevel.Field().String()
	if strRUT == "" {
		return false
	}

	// Remover puntos y guión
	strRUT = strings.ReplaceAll(strRUT, ".", "")
	strRUT = strings.ReplaceAll(strRUT, "-", "")

	if len(strRUT) < 2 {
		return false
	}

	// Separar número y dígito verificador
	strNumber := strRUT[:len(strRUT)-1]
	strDV := strings.ToUpper(strRUT[len(strRUT)-1:])

	// Validar que el número sea numérico
	intNumber, err := strconv.Atoi(strNumber)
	if err != nil {
		return false
	}

	// Calcular dígito verificador
	intSum := 0
	intMultiplier := 2

	for intI := len(strNumber) - 1; intI >= 0; intI-- {
		intDigit, _ := strconv.Atoi(string(strNumber[intI]))
		intSum += intDigit * intMultiplier
		intMultiplier++
		if intMultiplier > 7 {
			intMultiplier = 2
		}
	}

	intRemainder := intSum % 11
	strCalculatedDV := ""

	switch intRemainder {
	case 0:
		strCalculatedDV = "0"
	case 1:
		strCalculatedDV = "K"
	default:
		strCalculatedDV = strconv.Itoa(11 - intRemainder)
	}

	return strDV == strCalculatedDV && intNumber >= 1000000
}

// validateBarcode valida un código de barras
func validateBarcode(ptrFieldLevel validator.FieldLevel) bool {
	strBarcode := ptrFieldLevel.Field().String()
	if strBarcode == "" {
		return false
	}

	// Validar longitud (8, 12, 13, 14 dígitos son comunes)
	intLen := len(strBarcode)
	if intLen != 8 && intLen != 12 && intLen != 13 && intLen != 14 {
		return false
	}

	// Validar que solo contenga números
	for _, charRune := range strBarcode {
		if charRune < '0' || charRune > '9' {
			return false
		}
	}

	return true
}

// validateSKU valida un SKU
func validateSKU(ptrFieldLevel validator.FieldLevel) bool {
	strSKU := ptrFieldLevel.Field().String()
	if strSKU == "" {
		return false
	}

	// SKU debe tener entre 3 y 50 caracteres alfanuméricos, guiones y guiones bajos
	ptrRegex := regexp.MustCompile(`^[A-Za-z0-9_-]{3,50}$`)
	return ptrRegex.MatchString(strSKU)
}

// validatePrice valida un precio
func validatePrice(ptrFieldLevel validator.FieldLevel) bool {
	fltPrice := ptrFieldLevel.Field().Float()
	return fltPrice >= 0 && fltPrice <= 999999999.99
}

// validateStock valida una cantidad de stock
func validateStock(ptrFieldLevel validator.FieldLevel) bool {
	intStock := ptrFieldLevel.Field().Int()
	return intStock >= 0 && intStock <= 999999999
}

// validatePercentage valida un porcentaje
func validatePercentage(ptrFieldLevel validator.FieldLevel) bool {
	fltPercentage := ptrFieldLevel.Field().Float()
	return fltPercentage >= 0 && fltPercentage <= 100
}

// validatePhoneCL valida un teléfono chileno
func validatePhoneCL(ptrFieldLevel validator.FieldLevel) bool {
	strPhone := ptrFieldLevel.Field().String()
	if strPhone == "" {
		return false
	}

	// Remover espacios, guiones y paréntesis
	strPhone = strings.ReplaceAll(strPhone, " ", "")
	strPhone = strings.ReplaceAll(strPhone, "-", "")
	strPhone = strings.ReplaceAll(strPhone, "(", "")
	strPhone = strings.ReplaceAll(strPhone, ")", "")
	strPhone = strings.ReplaceAll(strPhone, "+", "")

	// Validar formato chileno: 9 dígitos o con código país
	ptrRegex := regexp.MustCompile(`^(56)?[2-9]\d{8}$`)
	return ptrRegex.MatchString(strPhone)
}

// validateFutureDate valida que una fecha sea futura
func validateFutureDate(ptrFieldLevel validator.FieldLevel) bool {
	timeValue := ptrFieldLevel.Field().Interface()

	switch ptrTime := timeValue.(type) {
	case time.Time:
		return ptrTime.After(time.Now())
	case *time.Time:
		if ptrTime == nil {
			return false
		}
		return ptrTime.After(time.Now())
	default:
		return false
	}
}

// validatePastDate valida que una fecha sea pasada
func validatePastDate(ptrFieldLevel validator.FieldLevel) bool {
	timeValue := ptrFieldLevel.Field().Interface()

	switch ptrTime := timeValue.(type) {
	case time.Time:
		return ptrTime.Before(time.Now())
	case *time.Time:
		if ptrTime == nil {
			return false
		}
		return ptrTime.Before(time.Now())
	default:
		return false
	}
}

// validateCurrency valida un código de moneda ISO 4217
func validateCurrency(ptrFieldLevel validator.FieldLevel) bool {
	strCurrency := strings.ToUpper(ptrFieldLevel.Field().String())

	// Lista de monedas comunes
	arrValidCurrencies := []string{
		"CLP", "USD", "EUR", "GBP", "JPY", "CAD", "AUD", "CHF", "CNY", "SEK", "NZD", "MXN", "SGD", "HKD", "NOK", "TRY", "ZAR", "BRL", "INR", "KRW", "RUB",
	}

	for _, strValidCurrency := range arrValidCurrencies {
		if strCurrency == strValidCurrency {
			return true
		}
	}

	return false
}

// validateLatitude valida una latitud
func validateLatitude(ptrFieldLevel validator.FieldLevel) bool {
	fltLatitude := ptrFieldLevel.Field().Float()
	return fltLatitude >= -90 && fltLatitude <= 90
}

// validateLongitude valida una longitud
func validateLongitude(ptrFieldLevel validator.FieldLevel) bool {
	fltLongitude := ptrFieldLevel.Field().Float()
	return fltLongitude >= -180 && fltLongitude <= 180
}

// validateJSON valida que una cadena sea JSON válido
func validateJSON(ptrFieldLevel validator.FieldLevel) bool {
	strJSON := ptrFieldLevel.Field().String()
	if strJSON == "" {
		return true // JSON vacío es válido
	}

	// Intentar parsear como JSON
	var objInterface interface{}
	return json.Unmarshal([]byte(strJSON), &objInterface) == nil
}

// validateUUIDv4 valida un UUID v4
func validateUUIDv4(ptrFieldLevel validator.FieldLevel) bool {
	strUUID := ptrFieldLevel.Field().String()
	if strUUID == "" {
		return false
	}

	ptrRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	return ptrRegex.MatchString(strings.ToLower(strUUID))
}

// validateHexColor valida un color hexadecimal
func validateHexColor(ptrFieldLevel validator.FieldLevel) bool {
	strColor := ptrFieldLevel.Field().String()
	if strColor == "" {
		return false
	}

	ptrRegex := regexp.MustCompile(`^#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})$`)
	return ptrRegex.MatchString(strColor)
}

// validateImageURL valida una URL de imagen
func validateImageURL(ptrFieldLevel validator.FieldLevel) bool {
	strURL := ptrFieldLevel.Field().String()
	if strURL == "" {
		return false
	}

	// Validar que sea una URL válida
	ptrRegexURL := regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
	if !ptrRegexURL.MatchString(strURL) {
		return false
	}

	// Validar que termine en extensión de imagen
	ptrRegexImage := regexp.MustCompile(`\.(jpg|jpeg|png|gif|bmp|webp|svg)(\?.*)?$`)
	return ptrRegexImage.MatchString(strings.ToLower(strURL))
}
