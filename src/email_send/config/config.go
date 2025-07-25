package config

import (
    "time"
)

// AppConfig configuración principal de la aplicación
type AppConfig struct {
    Name        string `yaml:"name" json:"name" validate:"required"`
    Version     string `yaml:"version" json:"version" validate:"required"`
    Environment string `yaml:"environment" json:"environment" validate:"required,oneof=development staging production"`
}

// LoggingConfig configuración de logging
type LoggingConfig struct {
    Level       string `yaml:"level" json:"level" validate:"required,oneof=debug info warn error"`
    Format      string `yaml:"format" json:"format" validate:"required,oneof=json text"`
    Output      string `yaml:"output" json:"output" validate:"required,oneof=stdout file"`
    FilePath    string `yaml:"file_path" json:"file_path"`
    MaxSize     int    `yaml:"max_size" json:"max_size" validate:"min=1"`
    MaxBackups  int    `yaml:"max_backups" json:"max_backups" validate:"min=0"`
    MaxAge      int    `yaml:"max_age" json:"max_age" validate:"min=1"`
}

// OAuth2Config configuración OAuth2
type OAuth2Config struct {
    Scopes      []string `yaml:"scopes" json:"scopes" validate:"required,min=1"`
    RedirectURL string   `yaml:"redirect_url" json:"redirect_url" validate:"required,url"`
}

// ProviderConfig configuración de un proveedor específico
type ProviderConfig struct {
    SMTPHost string        `yaml:"smtp_host" json:"smtp_host" validate:"required,hostname"`
    SMTPPort int           `yaml:"smtp_port" json:"smtp_port" validate:"required,min=1,max=65535"`
    UseTLS   bool          `yaml:"use_tls" json:"use_tls"`
    Timeout  time.Duration `yaml:"timeout" json:"timeout" validate:"required"`
    OAuth2   OAuth2Config  `yaml:"oauth2" json:"oauth2"`
}

// ProvidersConfig configuración de todos los proveedores
type ProvidersConfig struct {
    Gmail     ProviderConfig `yaml:"gmail" json:"gmail"`
    Office365 ProviderConfig `yaml:"office365" json:"office365"`
    Generic   ProviderConfig `yaml:"generic" json:"generic"`
}

// ConnectionPoolConfig configuración del pool de conexiones
type ConnectionPoolConfig struct {
    MaxConnections    int           `yaml:"max_connections" json:"max_connections" validate:"required,min=1"`
    MaxIdleTime       time.Duration `yaml:"max_idle_time" json:"max_idle_time" validate:"required"`
    ConnectionTimeout time.Duration `yaml:"connection_timeout" json:"connection_timeout" validate:"required"`
    IdleTimeout       time.Duration `yaml:"idle_timeout" json:"idle_timeout" validate:"required"`
    CleanupInterval   time.Duration `yaml:"cleanup_interval" json:"cleanup_interval" validate:"required"`
}

// SecurityConfig configuración de seguridad
type SecurityConfig struct {
    TokenEncryptionKey string        `yaml:"token_encryption_key" json:"token_encryption_key"`
    MaxTokenAge        time.Duration `yaml:"max_token_age" json:"max_token_age" validate:"required"`
    RequireTLS         bool          `yaml:"require_tls" json:"require_tls"`
    AllowedDomains     []string      `yaml:"allowed_domains" json:"allowed_domains"`
}

// RateLimitingConfig configuración de rate limiting
type RateLimitingConfig struct {
    Enabled           bool `yaml:"enabled" json:"enabled"`
    RequestsPerMinute int  `yaml:"requests_per_minute" json:"requests_per_minute" validate:"min=1"`
    Burst             int  `yaml:"burst" json:"burst" validate:"min=1"`
}

// RetryConfig configuración de reintentos
type RetryConfig struct {
    MaxAttempts     int           `yaml:"max_attempts" json:"max_attempts" validate:"required,min=1"`
    InitialBackoff  time.Duration `yaml:"initial_backoff" json:"initial_backoff" validate:"required"`
    MaxBackoff      time.Duration `yaml:"max_backoff" json:"max_backoff" validate:"required"`
    Multiplier      float64       `yaml:"multiplier" json:"multiplier" validate:"required,min=1"`
}

// MetricsConfig configuración de métricas
type MetricsConfig struct {
    Enabled bool   `yaml:"enabled" json:"enabled"`
    Port    int    `yaml:"port" json:"port" validate:"min=1,max=65535"`
    Path    string `yaml:"path" json:"path" validate:"required"`
}

// Config estructura principal de configuración
type Config struct {
    App            AppConfig            `yaml:"app" json:"app" validate:"required"`
    Logging        LoggingConfig        `yaml:"logging" json:"logging" validate:"required"`
    Providers      ProvidersConfig      `yaml:"providers" json:"providers" validate:"required"`
    ConnectionPool ConnectionPoolConfig `yaml:"connection_pool" json:"connection_pool" validate:"required"`
    Security       SecurityConfig       `yaml:"security" json:"security" validate:"required"`
    RateLimiting   RateLimitingConfig   `yaml:"rate_limiting" json:"rate_limiting" validate:"required"`
    Retry          RetryConfig          `yaml:"retry" json:"retry" validate:"required"`
    Metrics        MetricsConfig        `yaml:"metrics" json:"metrics" validate:"required"`
}

// EnvironmentCredentials credenciales específicas del entorno
type EnvironmentCredentials struct {
    GmailClientID         string `env:"GMAIL_CLIENT_ID"`
    GmailClientSecret     string `env:"GMAIL_CLIENT_SECRET"`
    Office365ClientID     string `env:"OFFICE365_CLIENT_ID"`
    Office365ClientSecret string `env:"OFFICE365_CLIENT_SECRET"`
    Office365TenantID     string `env:"OFFICE365_TENANT_ID"`
    TokenEncryptionKey    string `env:"TOKEN_ENCRYPTION_KEY"`
    JWTSecret            string `env:"JWT_SECRET"`
    DatabaseURL          string `env:"DATABASE_URL"`
    RedisURL             string `env:"REDIS_URL"`
    AdminEmail           string `env:"ADMIN_EMAIL"`
    ErrorNotificationEnabled bool `env:"ERROR_NOTIFICATION_ENABLED"`
}

// GetProviderConfig obtiene la configuración de un proveedor específico
func (c *Config) GetProviderConfig(provider string) (*ProviderConfig, bool) {
    switch provider {
    case "gmail":
        return &c.Providers.Gmail, true
    case "office365":
        return &c.Providers.Office365, true
    case "generic":
        return &c.Providers.Generic, true
    default:
        return nil, false
    }
}

// IsProduction verifica si estamos en entorno de producción
func (c *Config) IsProduction() bool {
    return c.App.Environment == "production"
}

// IsDevelopment verifica si estamos en entorno de desarrollo
func (c *Config) IsDevelopment() bool {
    return c.App.Environment == "development"
}
