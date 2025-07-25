package config

import (
    "fmt"
    "os"
    "path/filepath"
    "strings"
    
    "github.com/joho/godotenv"
    "github.com/spf13/viper"
)

// ConfigLoader maneja la carga de configuración
type ConfigLoader struct {
    configPath string
    envPath    string
}

// NewConfigLoader crea un nuevo cargador de configuración
func NewConfigLoader(configPath, envPath string) *ConfigLoader {
    return &ConfigLoader{
        configPath: configPath,
        envPath:    envPath,
    }
}

// Load carga la configuración completa
func (cl *ConfigLoader) Load() (*Config, *EnvironmentCredentials, error) {
    // Cargar variables de entorno desde archivo .env
    if err := cl.loadEnvFile(); err != nil {
        return nil, nil, fmt.Errorf("error cargando archivo .env: %w", err)
    }
    
    // Configurar viper
    if err := cl.configureViper(); err != nil {
        return nil, nil, fmt.Errorf("error configurando viper: %w", err)
    }
    
    // Cargar configuración principal
    config, err := cl.loadMainConfig()
    if err != nil {
        return nil, nil, fmt.Errorf("error cargando configuración principal: %w", err)
    }
    
    // Cargar credenciales de entorno
    credentials, err := cl.loadEnvironmentCredentials()
    if err != nil {
        return nil, nil, fmt.Errorf("error cargando credenciales de entorno: %w", err)
    }
    
    // Aplicar sobreescrituras de entorno
    cl.applyEnvironmentOverrides(config, credentials)
    
    return config, credentials, nil
}

func (cl *ConfigLoader) loadEnvFile() error {
    if cl.envPath == "" {
        // Intentar cargar .env desde el directorio actual
        if _, err := os.Stat(".env"); err == nil {
            cl.envPath = ".env"
        } else {
            // No hay archivo .env, continuar sin error
            return nil
        }
    }
    
    if err := godotenv.Load(cl.envPath); err != nil {
        return fmt.Errorf("error cargando archivo .env desde %s: %w", cl.envPath, err)
    }
    
    return nil
}

func (cl *ConfigLoader) configureViper() error {
    viper.AutomaticEnv()
    viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
    viper.SetEnvPrefix("SMTP_CLIENT")
    
    // Configurar paths de búsqueda
    if cl.configPath != "" {
        dir := filepath.Dir(cl.configPath)
        filename := filepath.Base(cl.configPath)
        ext := filepath.Ext(filename)
        name := strings.TrimSuffix(filename, ext)
        
        viper.AddConfigPath(dir)
        viper.SetConfigName(name)
        viper.SetConfigType(strings.TrimPrefix(ext, "."))
    } else {
        // Configuración por defecto
        viper.AddConfigPath("./configs")
        viper.AddConfigPath("./config")
        viper.AddConfigPath(".")
        viper.SetConfigName("config")
        viper.SetConfigType("yaml")
    }
    
    return nil
}

func (cl *ConfigLoader) loadMainConfig() (*Config, error) {
    // Leer configuración base
    if err := viper.ReadInConfig(); err != nil {
        return nil, fmt.Errorf("error leyendo configuración: %w", err)
    }
    
    // Cargar configuración específica del entorno si existe
    env := viper.GetString("app.environment")
    if env == "" {
        env = os.Getenv("SMTP_CLIENT_ENV")
    }
    
    if env != "" {
        cl.loadEnvironmentSpecificConfig(env)
    }
    
    // Deserializar configuración
    var config Config
    if err := viper.Unmarshal(&config); err != nil {
        return nil, fmt.Errorf("error deserializando configuración: %w", err)
    }
    
    return &config, nil
}

func (cl *ConfigLoader) loadEnvironmentSpecificConfig(env string) {
    envConfigPath := fmt.Sprintf("config.%s", env)
    
    // Intentar cargar configuración específica del entorno
    viper.SetConfigName(envConfigPath)
    if err := viper.MergeInConfig(); err != nil {
        // Si no existe configuración específica del entorno, continuar
        return
    }
}

func (cl *ConfigLoader) loadEnvironmentCredentials() (*EnvironmentCredentials, error) {
    credentials := &EnvironmentCredentials{
        GmailClientID:         os.Getenv("GMAIL_CLIENT_ID"),
        GmailClientSecret:     os.Getenv("GMAIL_CLIENT_SECRET"),
        Office365ClientID:     os.Getenv("OFFICE365_CLIENT_ID"),
        Office365ClientSecret: os.Getenv("OFFICE365_CLIENT_SECRET"),
        Office365TenantID:     os.Getenv("OFFICE365_TENANT_ID"),
        TokenEncryptionKey:    os.Getenv("TOKEN_ENCRYPTION_KEY"),
        JWTSecret:            os.Getenv("JWT_SECRET"),
        DatabaseURL:          os.Getenv("DATABASE_URL"),
        RedisURL:             os.Getenv("REDIS_URL"),
        AdminEmail:           os.Getenv("ADMIN_EMAIL"),
        ErrorNotificationEnabled: os.Getenv("ERROR_NOTIFICATION_ENABLED") == "true",
    }
    
    return credentials, nil
}

func (cl *ConfigLoader) applyEnvironmentOverrides(config *Config, credentials *EnvironmentCredentials) {
    // Aplicar sobreescrituras desde variables de entorno
    if credentials.TokenEncryptionKey != "" {
        config.Security.TokenEncryptionKey = credentials.TokenEncryptionKey
    }
    
    // Aplicar otras sobreescrituras según sea necesario
    if env := os.Getenv("SMTP_CLIENT_ENV"); env != "" {
        config.App.Environment = env
    }
}

// LoadFromPath carga configuración desde un path específico
func LoadFromPath(configPath string) (*Config, *EnvironmentCredentials, error) {
    loader := NewConfigLoader(configPath, "")
    return loader.Load()
}

// LoadDefault carga configuración usando paths por defecto
func LoadDefault() (*Config, *EnvironmentCredentials, error) {
    loader := NewConfigLoader("", "")
    return loader.Load()
}

// LoadWithEnv carga configuración con archivo .env específico
func LoadWithEnv(envPath string) (*Config, *EnvironmentCredentials, error) {
    loader := NewConfigLoader("", envPath)
    return loader.Load()
}