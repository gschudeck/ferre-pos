package auth

import (
    "context"
    "time"
    
    "golang.org/x/oauth2"
)

// ProviderType representa un tipo de proveedor de autenticación
type ProviderType string

const (
    ProviderGmail     ProviderType = "gmail"
    ProviderOffice365 ProviderType = "office365"
    ProviderGeneric   ProviderType = "generic"
)

// AuthConfig contiene la configuración de autenticación
type AuthConfig struct {
    Provider     ProviderType
    ClientID     string
    ClientSecret string
    RedirectURL  string
    Scopes       []string
    TenantID     string // Solo para Office 365
    Timeout      time.Duration
}

// TokenManager maneja la obtención y actualización de tokens
type TokenManager interface {
    GetToken(ctx context.Context, requestID string) (*oauth2.Token, error)
    RefreshToken(ctx context.Context, token *oauth2.Token, requestID string) (*oauth2.Token, error)
    GetAuthURL(state string) string
    ExchangeCode(ctx context.Context, code, requestID string) (*oauth2.Token, error)
}

// TokenStore define la interfaz para almacenar tokens
type TokenStore interface {
    StoreToken(ctx context.Context, userID string, token *oauth2.Token) error
    GetToken(ctx context.Context, userID string) (*oauth2.Token, error)
    DeleteToken(ctx context.Context, userID string) error
}
