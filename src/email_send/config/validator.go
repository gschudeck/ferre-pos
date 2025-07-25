package config

import (
    "fmt"
    "strings"
    
    "github.com/go-playground/validator/v10"
)

// ConfigValidator valida la configuración
type ConfigValidator struct {
    validator *validator.Validate
}

// NewConfigValidator crea un nuevo validador de configuración
func NewConfigValidator() *ConfigValidator {
    v := validator.New()
    return &ConfigValidator{
        validator: v,
    }
}

// ValidateConfig valida toda la configuración
func (cv *ConfigValidator) ValidateConfig(config *Config) error {
    if err := cv.validator.Struct(config); err != nil {
        return cv.formatValidationError(err)
    }
    
    // Validaciones personalizadas
    if err := cv.validateCustomRules(config); err != nil {
        return err
    }
    
    return nil
}

// ValidateCredentials valida las credenciales de entorno
func (cv *ConfigValidator) ValidateCredentials(credentials *EnvironmentCredentials, config *Config) error {
    var errors []string
    
    // Validar credenciales requeridas según el entorno
    if config.IsProduction() {
        if credentials.TokenEncryptionKey == "" {
            errors = append(errors, "TOKEN_ENCRYPTION_KEY es requerido en producción")
        }
        
        if len(credentials.TokenEncryptionKey) < 32 {
            errors = append(errors, "TOKEN_ENCRYPTION_KEY debe tener al menos 32 caracteres")
        }
    }
    
    // Validar credenciales OAuth2 si se van a usar
    if credentials.GmailClientID != "" && credentials.GmailClientSecret == "" {
        errors = append(errors, "GMAIL_CLIENT_SECRET es requerido cuando se define GMAIL_CLIENT_ID")
    }
    
    if credentials.Office365ClientID != "" {
        if credentials.Office365ClientSecret == "" {
            errors = append(errors, "OFFICE365_CLIENT_SECRET es requerido cuando se define OFFICE365_CLIENT_ID")
        }
        if credentials.Office365TenantID == "" {
            errors = append(errors, "OFFICE365_TENANT_ID es requerido cuando se define OFFICE365_CLIENT_ID")
        }
    }
    
    if len(errors) > 0 {
        return fmt.Errorf("errores de validación de credenciales: %s", strings.Join(errors, "; "))
    }
    
    return nil
}

func (cv *ConfigValidator) formatValidationError(err error) error {
    var errors []string
    
    for _, e := range err.(validator.ValidationErrors) {
        switch e.Tag() {
        case "required":
            errors = append(errors, fmt.Sprintf("Campo '%s' es requerido", e.Field()))
        case "min":
            errors = append(errors, fmt.Sprintf("Campo '%s' debe ser mayor a %s", e.Field(), e.Param()))
        case "max":
            errors = append(errors, fmt.Sprintf("Campo '%s' debe ser menor a %s", e.Field(), e.Param()))
        case "oneof":
            errors = append(errors, fmt.Sprintf("Campo '%s' debe ser uno de: %s", e.Field(), e.Param()))
        case "url":
            errors = append(errors, fmt.Sprintf("Campo '%s' debe ser una URL válida", e.Field()))
        case "hostname":
            errors = append(errors, fmt.Sprintf("Campo '%s' debe ser un hostname válido", e.Field()))
        default:
            errors = append(errors, fmt.Sprintf("Campo '%s' no es válido: %s", e.Field(), e.Tag()))
        }
    }
    
    return fmt.Errorf("errores de validación: %s", strings.Join(errors, "; "))
}

func (cv *ConfigValidator) validateCustomRules(config *Config) error {
    var errors []string
    
    // Validar que MaxBackoff sea mayor que InitialBackoff
    if config.Retry.MaxBackoff <= config.Retry.InitialBackoff {
        errors = append(errors, "retry.max_backoff debe ser mayor que retry.initial_backoff")
    }
    
    // Validar que ConnectionTimeout sea menor que IdleTimeout
    if config.ConnectionPool.ConnectionTimeout >= config.ConnectionPool.IdleTimeout {
        errors = append(errors, "connection_pool.connection_timeout debe ser menor que connection_pool.idle_timeout")
    }
    
    // Validar configuración de rate limiting
    if config.RateLimiting.Enabled && config.RateLimiting.Burst > config.RateLimiting.RequestsPerMinute {
        errors = append(errors, "rate_limiting.burst no puede ser mayor que rate_limiting.requests_per_minute")
    }
    
    if len(errors) > 0 {
        return fmt.Errorf("errores de validación personalizada: %s", strings.Join(errors, "; "))
    }
    
    return nil
}
}