package auth

import (
    "context"
    "fmt"
    "net/url"
    "strings"
    "time"
    
    "github.com/yourorg/smtp-client/pkg/errors"
)

// AuthValidator valida configuraciones de autenticación
type AuthValidator struct{}

// NewAuthValidator crea un nuevo validador de autenticación
func NewAuthValidator() *AuthValidator {
    return &AuthValidator{}
}

// ValidateConfig valida la configuración de autenticación
func (v *AuthValidator) ValidateConfig(cfg *AuthConfig, requestID string) error {
    if cfg == nil {
        return errors.NewValidationError(
            "AUTH_CONFIG_NIL",
            "La configuración de autenticación no puede ser nil",
            requestID,
            nil,
        )
    }
    
    if err := v.validateProvider(cfg.Provider, requestID); err != nil {
        return err
    }
    
    if err := v.validateClientCredentials(cfg, requestID); err != nil {
        return err
    }
    
    if err := v.validateRedirectURL(cfg.RedirectURL, requestID); err != nil {
        return err
    }
    
    if err := v.validateScopes(cfg.Scopes, requestID); err != nil {
        return err
    }
    
    if cfg.Provider == ProviderOffice365 {
        if err := v.validateTenantID(cfg.TenantID, requestID); err != nil {
            return err
        }
    }
    
    return nil
}

func (v *AuthValidator) validateProvider(provider ProviderType, requestID string) error {
    switch provider {
    case ProviderGmail, ProviderOffice365, ProviderGeneric:
        return nil
    default:
        return errors.NewValidationError(
            "INVALID_PROVIDER",
            fmt.Sprintf("Proveedor no soportado: %s", provider),
            requestID,
            nil,
        )
    }
}

func (v *AuthValidator) validateClientCredentials(cfg *AuthConfig, requestID string) error {
    if strings.TrimSpace(cfg.ClientID) == "" {
        return errors.NewValidationError(
            "MISSING_CLIENT_ID",
            "ClientID es requerido",
            requestID,
            nil,
        )
    }
    
    if strings.TrimSpace(cfg.ClientSecret) == "" {
        return errors.NewValidationError(
            "MISSING_CLIENT_SECRET",
            "ClientSecret es requerido",
            requestID,
            nil,
        )
    }
    
    return nil
}

func (v *AuthValidator) validateRedirectURL(redirectURL, requestID string) error {
    if strings.TrimSpace(redirectURL) == "" {
        return errors.NewValidationError(
            "MISSING_REDIRECT_URL",
            "RedirectURL es requerido",
            requestID,
            nil,
        )
    }
    
    if _, err := url.Parse(redirectURL); err != nil {
        return errors.NewValidationError(
            "INVALID_REDIRECT_URL",
            "RedirectURL tiene formato inválido",
            requestID,
            err,
        )
    }
    
    return nil
}

func (v *AuthValidator) validateScopes(scopes []string, requestID string) error {
    if len(scopes) == 0 {
        return errors.NewValidationError(
            "MISSING_SCOPES",
            "Al menos un scope es requerido",
            requestID,
            nil,
        )
    }
    
    for _, scope := range scopes {
        if strings.TrimSpace(scope) == "" {
            return errors.NewValidationError(
                "INVALID_SCOPE",
                "Los scopes no pueden estar vacíos",
                requestID,
                nil,
            )
        }
    }
    
    return nil
}

func (v *AuthValidator) validateTenantID(tenantID, requestID string) error {
    if strings.TrimSpace(tenantID) == "" {
        return errors.NewValidationError(
            "MISSING_TENANT_ID",
            "TenantID es requerido para Office 365",
            requestID,
            nil,
        )
    }
    
    return nil
}

// ValidateToken valida un token OAuth2
func (v *AuthValidator) ValidateToken(token *oauth2.Token, requestID string) error {
    if token == nil {
        return errors.NewValidationError(
            "TOKEN_NIL",
            "Token no puede ser nil",
            requestID,
            nil,
        )
    }
    
    if strings.TrimSpace(token.AccessToken) == "" {
        return errors.NewValidationError(
            "MISSING_ACCESS_TOKEN",
            "AccessToken es requerido",
            requestID,
            nil,
        )
    }
    
    if token.Expiry.Before(time.Now().Add(time.Minute)) {
        return errors.NewValidationError(
            "TOKEN_EXPIRED",
            "Token está expirado o expirará pronto",
            requestID,
            nil,
        )
    }
    
    return nil
}
