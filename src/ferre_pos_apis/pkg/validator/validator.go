package validator

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// Validator interfaz principal del validador
type Validator interface {
	ValidateStruct(s interface{}) error
	ValidateField(field interface{}, tag string) error
	RegisterValidation(tag string, fn validator.Func) error
	GetValidationErrors(err error) []ValidationError
}

// ValidationError error de validación estructurado
type ValidationError struct {
	Field   string      `json:"field"`
	Tag     string      `json:"tag"`
	Value   interface{} `json:"value,omitempty"`
	Message string      `json:"message"`
	Param   string      `json:"param,omitempty"`
}

// validatorImpl implementación del validador
type validatorImpl struct {
	validate *validator.Validate
}

// New crea una nueva instancia del validador
func New() Validator {
	validate := validator.New()
	
	v := &validatorImpl{
		validate: validate,
	}

	// Registrar validaciones personalizadas
	v.registerCustomValidations()
	
	// Configurar nombres de campos usando tags JSON
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return v
}

// ValidateStruct valida una estructura completa
func (v *validatorImpl) ValidateStruct(s interface{}) error {
	return v.validate.Struct(s)
}

// ValidateField valida un campo individual
func (v *validatorImpl) ValidateField(field interface{}, tag string) error {
	return v.validate.Var(field, tag)
}

// RegisterValidation registra una validación personalizada
func (v *validatorImpl) RegisterValidation(tag string, fn validator.Func) error {
	return v.validate.RegisterValidation(tag, fn)
}

// GetValidationErrors convierte errores de validación a formato estructurado
func (v *validatorImpl) GetValidationErrors(err error) []ValidationError {
	var validationErrors []ValidationError

	if err == nil {
		return validationErrors
	}

	if validatorErrors, ok := err.(validator.ValidationErrors); ok {
		for _, err := range validatorErrors {
			validationError := ValidationError{
				Field:   err.Field(),
				Tag:     err.Tag(),
				Value:   err.Value(),
				Param:   err.Param(),
				Message: getErrorMessage(err),
			}
			validationErrors = append(validationErrors, validationError)
		}
	}

	return validationErrors
}

// registerCustomValidations registra todas las validaciones personalizadas
func (v *validatorImpl) registerCustomValidations() {
	// Validación de RUT chileno
	v.validate.RegisterValidation("rut", validateRUT)
	
	// Validación de código de barras
	v.validate.RegisterValidation("barcode", validateBarcode)
	
	// Validación de UUID
	v.validate.RegisterValidation("uuid", validateUUID)
	
	// Validación de precio (debe ser positivo)
	v.validate.RegisterValidation("price", validatePrice)
	
	// Validación de cantidad (debe ser positiva)
	v.validate.RegisterValidation("quantity", validateQuantity)
	
	// Validación de código interno de producto
	v.validate.RegisterValidation("product_code", validateProductCode)
	
	// Validación de tipo de documento
	v.validate.RegisterValidation("document_type", validateDocumentType)
	
	// Validación de medio de pago
	v.validate.RegisterValidation("payment_method", validatePaymentMethod)
	
	// Validación de rol de usuario
	v.validate.RegisterValidation("user_role", validateUserRole)
	
	// Validación de estado
	v.validate.RegisterValidation("status", validateStatus)
	
	// Validación de teléfono chileno
	v.validate.RegisterValidation("phone_cl", validatePhoneCL)
	
	// Validación de código postal chileno
	v.validate.RegisterValidation("postal_code_cl", validatePostalCodeCL)
	
	// Validación de formato de etiqueta
	v.validate.RegisterValidation("label_format", validateLabelFormat)
	
	// Validación de tipo de reporte
	v.validate.RegisterValidation("report_type", validateReportType)
}

// validateRUT valida formato y dígito verificador de RUT chileno
func validateRUT(fl validator.FieldLevel) bool {
	rut := fl.Field().String()
	if rut == "" {
		return true // Permitir vacío si no es required
	}

	// Formato básico: 12345678-9 o 12345678-K
	rutRegex := regexp.MustCompile(`^(\d{7,8})-([0-9Kk])$`)
	matches := rutRegex.FindStringSubmatch(rut)
	if len(matches) != 3 {
		return false
	}

	number, err := strconv.Atoi(matches[1])
	if err != nil {
		return false
	}

	expectedDV := calculateRUTDV(number)
	providedDV := strings.ToUpper(matches[2])

	return expectedDV == providedDV
}

// calculateRUTDV calcula el dígito verificador del RUT
func calculateRUTDV(rut int) string {
	var sum int
	multiplier := 2

	for rut > 0 {
		sum += (rut % 10) * multiplier
		rut /= 10
		multiplier++
		if multiplier > 7 {
			multiplier = 2
		}
	}

	remainder := sum % 11
	dv := 11 - remainder

	switch dv {
	case 10:
		return "K"
	case 11:
		return "0"
	default:
		return strconv.Itoa(dv)
	}
}

// validateBarcode valida formato de código de barras
func validateBarcode(fl validator.FieldLevel) bool {
	barcode := fl.Field().String()
	if barcode == "" {
		return true
	}

	// Validar longitud y que solo contenga dígitos
	if len(barcode) < 8 || len(barcode) > 14 {
		return false
	}

	for _, char := range barcode {
		if !unicode.IsDigit(char) {
			return false
		}
	}

	return true
}

// validateUUID valida formato UUID
func validateUUID(fl validator.FieldLevel) bool {
	uuidStr := fl.Field().String()
	if uuidStr == "" {
		return true
	}

	_, err := uuid.Parse(uuidStr)
	return err == nil
}

// validatePrice valida que el precio sea positivo
func validatePrice(fl validator.FieldLevel) bool {
	price := fl.Field().Float()
	return price >= 0
}

// validateQuantity valida que la cantidad sea positiva
func validateQuantity(fl validator.FieldLevel) bool {
	quantity := fl.Field().Float()
	return quantity > 0
}

// validateProductCode valida código interno de producto
func validateProductCode(fl validator.FieldLevel) bool {
	code := fl.Field().String()
	if code == "" {
		return true
	}

	// Código debe tener entre 3 y 50 caracteres, solo alfanuméricos y guiones
	if len(code) < 3 || len(code) > 50 {
		return false
	}

	codeRegex := regexp.MustCompile(`^[a-zA-Z0-9\-_]+$`)
	return codeRegex.MatchString(code)
}

// validateDocumentType valida tipo de documento
func validateDocumentType(fl validator.FieldLevel) bool {
	docType := fl.Field().String()
	validTypes := map[string]bool{
		"boleta":     true,
		"factura":    true,
		"guia":       true,
		"nota_venta": true,
	}
	return validTypes[docType]
}

// validatePaymentMethod valida medio de pago
func validatePaymentMethod(fl validator.FieldLevel) bool {
	method := fl.Field().String()
	validMethods := map[string]bool{
		"efectivo":             true,
		"tarjeta_debito":       true,
		"tarjeta_credito":      true,
		"transferencia":        true,
		"cheque":               true,
		"puntos_fidelizacion":  true,
		"otro":                 true,
	}
	return validMethods[method]
}

// validateUserRole valida rol de usuario
func validateUserRole(fl validator.FieldLevel) bool {
	role := fl.Field().String()
	validRoles := map[string]bool{
		"cajero":             true,
		"vendedor":           true,
		"despacho":           true,
		"supervisor":         true,
		"admin":              true,
		"operador_etiquetas": true,
	}
	return validRoles[role]
}

// validateStatus valida estados generales
func validateStatus(fl validator.FieldLevel) bool {
	status := fl.Field().String()
	validStatuses := map[string]bool{
		"pendiente":  true,
		"procesado":  true,
		"completado": true,
		"error":      true,
		"cancelado":  true,
		"activo":     true,
		"inactivo":   true,
	}
	return validStatuses[status]
}

// validatePhoneCL valida formato de teléfono chileno
func validatePhoneCL(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	if phone == "" {
		return true
	}

	// Formatos válidos: +56912345678, 56912345678, 912345678, 221234567
	phoneRegex := regexp.MustCompile(`^(\+?56)?([2-9]\d{8}|[2-9]\d{7})$`)
	return phoneRegex.MatchString(phone)
}

// validatePostalCodeCL valida código postal chileno
func validatePostalCodeCL(fl validator.FieldLevel) bool {
	postalCode := fl.Field().String()
	if postalCode == "" {
		return true
	}

	// Código postal chileno: 7 dígitos
	postalRegex := regexp.MustCompile(`^\d{7}$`)
	return postalRegex.MatchString(postalCode)
}

// validateLabelFormat valida formato de etiqueta
func validateLabelFormat(fl validator.FieldLevel) bool {
	format := fl.Field().String()
	validFormats := map[string]bool{
		"pdf": true,
		"png": true,
		"zpl": true,
		"jpg": true,
	}
	return validFormats[format]
}

// validateReportType valida tipo de reporte
func validateReportType(fl validator.FieldLevel) bool {
	reportType := fl.Field().String()
	validTypes := map[string]bool{
		"ventas":      true,
		"stock":       true,
		"productos":   true,
		"usuarios":    true,
		"movimientos": true,
		"financiero":  true,
	}
	return validTypes[reportType]
}

// getErrorMessage obtiene mensaje de error personalizado
func getErrorMessage(err validator.FieldError) string {
	field := err.Field()
	tag := err.Tag()
	param := err.Param()

	switch tag {
	case "required":
		return fmt.Sprintf("El campo '%s' es requerido", field)
	case "email":
		return fmt.Sprintf("El campo '%s' debe ser un email válido", field)
	case "min":
		return fmt.Sprintf("El campo '%s' debe tener al menos %s caracteres", field, param)
	case "max":
		return fmt.Sprintf("El campo '%s' debe tener máximo %s caracteres", field, param)
	case "len":
		return fmt.Sprintf("El campo '%s' debe tener exactamente %s caracteres", field, param)
	case "gte":
		return fmt.Sprintf("El campo '%s' debe ser mayor o igual a %s", field, param)
	case "gt":
		return fmt.Sprintf("El campo '%s' debe ser mayor a %s", field, param)
	case "lte":
		return fmt.Sprintf("El campo '%s' debe ser menor o igual a %s", field, param)
	case "lt":
		return fmt.Sprintf("El campo '%s' debe ser menor a %s", field, param)
	case "oneof":
		return fmt.Sprintf("El campo '%s' debe ser uno de: %s", field, param)
	case "rut":
		return fmt.Sprintf("El campo '%s' debe ser un RUT válido (formato: 12345678-9)", field)
	case "barcode":
		return fmt.Sprintf("El campo '%s' debe ser un código de barras válido", field)
	case "uuid":
		return fmt.Sprintf("El campo '%s' debe ser un UUID válido", field)
	case "price":
		return fmt.Sprintf("El campo '%s' debe ser un precio válido (mayor o igual a 0)", field)
	case "quantity":
		return fmt.Sprintf("El campo '%s' debe ser una cantidad válida (mayor a 0)", field)
	case "product_code":
		return fmt.Sprintf("El campo '%s' debe ser un código de producto válido", field)
	case "document_type":
		return fmt.Sprintf("El campo '%s' debe ser un tipo de documento válido", field)
	case "payment_method":
		return fmt.Sprintf("El campo '%s' debe ser un medio de pago válido", field)
	case "user_role":
		return fmt.Sprintf("El campo '%s' debe ser un rol de usuario válido", field)
	case "phone_cl":
		return fmt.Sprintf("El campo '%s' debe ser un teléfono chileno válido", field)
	case "postal_code_cl":
		return fmt.Sprintf("El campo '%s' debe ser un código postal chileno válido", field)
	case "label_format":
		return fmt.Sprintf("El campo '%s' debe ser un formato de etiqueta válido", field)
	case "report_type":
		return fmt.Sprintf("El campo '%s' debe ser un tipo de reporte válido", field)
	default:
		return fmt.Sprintf("El campo '%s' no es válido", field)
	}
}

// SanitizeInput sanitiza entrada de usuario
func SanitizeInput(input string) string {
	// Remover caracteres de control
	sanitized := strings.Map(func(r rune) rune {
		if unicode.IsControl(r) && r != '\n' && r != '\r' && r != '\t' {
			return -1
		}
		return r
	}, input)

	// Trim espacios
	sanitized = strings.TrimSpace(sanitized)

	return sanitized
}

// SanitizeStruct sanitiza todos los campos string de una estructura
func SanitizeStruct(s interface{}) {
	v := reflect.ValueOf(s)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		if !field.CanSet() {
			continue
		}

		switch field.Kind() {
		case reflect.String:
			sanitized := SanitizeInput(field.String())
			field.SetString(sanitized)
		case reflect.Ptr:
			if !field.IsNil() && field.Elem().Kind() == reflect.String {
				sanitized := SanitizeInput(field.Elem().String())
				field.Elem().SetString(sanitized)
			}
		case reflect.Struct:
			if field.CanAddr() {
				SanitizeStruct(field.Addr().Interface())
			}
		case reflect.Slice:
			if fieldType.Type.Elem().Kind() == reflect.Struct {
				for j := 0; j < field.Len(); j++ {
					if field.Index(j).CanAddr() {
						SanitizeStruct(field.Index(j).Addr().Interface())
					}
				}
			}
		}
	}
}

// ValidateAndSanitize valida y sanitiza una estructura
func ValidateAndSanitize(v Validator, s interface{}) error {
	// Primero sanitizar
	SanitizeStruct(s)
	
	// Luego validar
	return v.ValidateStruct(s)
}

// IsValidEmail valida formato de email
func IsValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// IsValidURL valida formato de URL
func IsValidURL(url string) bool {
	urlRegex := regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
	return urlRegex.MatchString(url)
}

// IsValidJSONB valida que una cadena sea JSON válido
func IsValidJSONB(jsonStr string) bool {
	if jsonStr == "" {
		return true
	}
	
	// Intentar parsear como JSON
	var js interface{}
	return json.Unmarshal([]byte(jsonStr), &js) == nil
}

// ValidateBusinessRules valida reglas de negocio específicas
func ValidateBusinessRules(data interface{}) []ValidationError {
	var errors []ValidationError

	// Aquí se pueden agregar validaciones de reglas de negocio específicas
	// Por ejemplo, validar que el total de una venta coincida con la suma de items

	return errors
}

// CreateValidationResponse crea respuesta estándar de error de validación
func CreateValidationResponse(errors []ValidationError) map[string]interface{} {
	return map[string]interface{}{
		"success": false,
		"error": map[string]interface{}{
			"code":    "VALIDATION_ERROR",
			"message": "Error de validación en los datos enviados",
			"details": map[string]interface{}{
				"validation_errors": errors,
				"total_errors":      len(errors),
			},
		},
	}
}

