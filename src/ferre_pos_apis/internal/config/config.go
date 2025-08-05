package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config estructura principal de configuración
type Config struct {
	Database   DatabaseConfig   `mapstructure:"database"`
	Logging    LoggingConfig    `mapstructure:"logging"`
	Metrics    MetricsConfig    `mapstructure:"metrics"`
	RateLimit  RateLimitConfig  `mapstructure:"rate_limiting"`
	APIs       APIsConfig       `mapstructure:"apis"`
	Security   SecurityConfig   `mapstructure:"security"`
	Monitoring MonitoringConfig `mapstructure:"monitoring"`
	Development DevelopmentConfig `mapstructure:"development"`
}

// DatabaseConfig configuración de base de datos
type DatabaseConfig struct {
	Host                   string        `mapstructure:"host"`
	Port                   int           `mapstructure:"port"`
	Name                   string        `mapstructure:"name"`
	User                   string        `mapstructure:"user"`
	Password               string        `mapstructure:"password"`
	SSLMode                string        `mapstructure:"sslmode"`
	MaxOpenConnections     int           `mapstructure:"max_open_connections"`
	MaxIdleConnections     int           `mapstructure:"max_idle_connections"`
	ConnectionMaxLifetime  time.Duration `mapstructure:"connection_max_lifetime"`
	ConnectionMaxIdleTime  time.Duration `mapstructure:"connection_max_idle_time"`
}

// LoggingConfig configuración de logging
type LoggingConfig struct {
	Level        string `mapstructure:"level"`
	Format       string `mapstructure:"format"`
	Output       string `mapstructure:"output"`
	FilePath     string `mapstructure:"file_path"`
	MaxSizeMB    int    `mapstructure:"max_size_mb"`
	MaxBackups   int    `mapstructure:"max_backups"`
	MaxAgeDays   int    `mapstructure:"max_age_days"`
	Compress     bool   `mapstructure:"compress"`
}

// MetricsConfig configuración de métricas Prometheus
type MetricsConfig struct {
	Enabled                  bool   `mapstructure:"enabled"`
	Port                     int    `mapstructure:"port"`
	Path                     string `mapstructure:"path"`
	Namespace                string `mapstructure:"namespace"`
	Subsystem                string `mapstructure:"subsystem"`
	CollectRuntimeMetrics    bool   `mapstructure:"collect_runtime_metrics"`
	CollectDatabaseMetrics   bool   `mapstructure:"collect_database_metrics"`
}

// RateLimitConfig configuración de rate limiting
type RateLimitConfig struct {
	Enabled            bool          `mapstructure:"enabled"`
	RequestsPerSecond  float64       `mapstructure:"requests_per_second"`
	BurstSize          int           `mapstructure:"burst_size"`
	CleanupInterval    time.Duration `mapstructure:"cleanup_interval"`
}

// APIsConfig configuración de todas las APIs
type APIsConfig struct {
	POS     APIConfig `mapstructure:"pos"`
	Sync    APIConfig `mapstructure:"sync"`
	Labels  APIConfig `mapstructure:"labels"`
	Reports APIConfig `mapstructure:"reports"`
}

// APIConfig configuración específica de cada API
type APIConfig struct {
	Enabled          bool                   `mapstructure:"enabled"`
	Port             int                    `mapstructure:"port"`
	Host             string                 `mapstructure:"host"`
	ReadTimeout      time.Duration          `mapstructure:"read_timeout"`
	WriteTimeout     time.Duration          `mapstructure:"write_timeout"`
	IdleTimeout      time.Duration          `mapstructure:"idle_timeout"`
	MaxHeaderBytes   int                    `mapstructure:"max_header_bytes"`
	RateLimit        RateLimitConfig        `mapstructure:"rate_limiting"`
	Cache            CacheConfig            `mapstructure:"cache"`
	Auth             AuthConfig             `mapstructure:"auth"`
	BatchProcessing  BatchProcessingConfig  `mapstructure:"batch_processing"`
	Retry            RetryConfig            `mapstructure:"retry"`
	LabelGeneration  LabelGenerationConfig  `mapstructure:"label_generation"`
	ReportGeneration ReportGenerationConfig `mapstructure:"report_generation"`
}

// CacheConfig configuración de cache
type CacheConfig struct {
	Enabled    bool          `mapstructure:"enabled"`
	TTL        time.Duration `mapstructure:"ttl"`
	MaxEntries int           `mapstructure:"max_entries"`
}

// AuthConfig configuración de autenticación
type AuthConfig struct {
	JWTSecret           string        `mapstructure:"jwt_secret"`
	TokenExpiry         time.Duration `mapstructure:"token_expiry"`
	RefreshTokenExpiry  time.Duration `mapstructure:"refresh_token_expiry"`
}

// BatchProcessingConfig configuración de procesamiento por lotes
type BatchProcessingConfig struct {
	MaxBatchSize int           `mapstructure:"max_batch_size"`
	Timeout      time.Duration `mapstructure:"timeout"`
}

// RetryConfig configuración de reintentos
type RetryConfig struct {
	MaxAttempts  int           `mapstructure:"max_attempts"`
	InitialDelay time.Duration `mapstructure:"initial_delay"`
	MaxDelay     time.Duration `mapstructure:"max_delay"`
}

// LabelGenerationConfig configuración de generación de etiquetas
type LabelGenerationConfig struct {
	MaxConcurrentJobs int      `mapstructure:"max_concurrent_jobs"`
	JobTimeout        time.Duration `mapstructure:"job_timeout"`
	TempDirectory     string   `mapstructure:"temp_directory"`
	OutputFormats     []string `mapstructure:"output_formats"`
}

// ReportGenerationConfig configuración de generación de reportes
type ReportGenerationConfig struct {
	MaxConcurrentReports int      `mapstructure:"max_concurrent_reports"`
	ReportTimeout        time.Duration `mapstructure:"report_timeout"`
	TempDirectory        string   `mapstructure:"temp_directory"`
	OutputFormats        []string `mapstructure:"output_formats"`
	MaxRowsPerReport     int      `mapstructure:"max_rows_per_report"`
}

// SecurityConfig configuración de seguridad
type SecurityConfig struct {
	CORS       CORSConfig       `mapstructure:"cors"`
	Encryption EncryptionConfig `mapstructure:"encryption"`
	Session    SessionConfig    `mapstructure:"session"`
}

// CORSConfig configuración de CORS
type CORSConfig struct {
	Enabled          bool     `mapstructure:"enabled"`
	AllowedOrigins   []string `mapstructure:"allowed_origins"`
	AllowedMethods   []string `mapstructure:"allowed_methods"`
	AllowedHeaders   []string `mapstructure:"allowed_headers"`
	ExposedHeaders   []string `mapstructure:"exposed_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
	MaxAge           int      `mapstructure:"max_age"`
}

// EncryptionConfig configuración de encriptación
type EncryptionConfig struct {
	Algorithm           string        `mapstructure:"algorithm"`
	KeyRotationInterval time.Duration `mapstructure:"key_rotation_interval"`
}

// SessionConfig configuración de sesiones
type SessionConfig struct {
	Secure   bool   `mapstructure:"secure"`
	HTTPOnly bool   `mapstructure:"http_only"`
	SameSite string `mapstructure:"same_site"`
	MaxAge   int    `mapstructure:"max_age"`
}

// MonitoringConfig configuración de monitoreo
type MonitoringConfig struct {
	HealthCheck HealthCheckConfig `mapstructure:"health_check"`
	Profiling   ProfilingConfig   `mapstructure:"profiling"`
	Tracing     TracingConfig     `mapstructure:"tracing"`
}

// HealthCheckConfig configuración de health checks
type HealthCheckConfig struct {
	Enabled  bool          `mapstructure:"enabled"`
	Path     string        `mapstructure:"path"`
	Interval time.Duration `mapstructure:"interval"`
	Timeout  time.Duration `mapstructure:"timeout"`
}

// ProfilingConfig configuración de profiling
type ProfilingConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Path    string `mapstructure:"path"`
}

// TracingConfig configuración de tracing
type TracingConfig struct {
	Enabled         bool   `mapstructure:"enabled"`
	JaegerEndpoint  string `mapstructure:"jaeger_endpoint"`
	ServiceName     string `mapstructure:"service_name"`
}

// DevelopmentConfig configuración de desarrollo
type DevelopmentConfig struct {
	Debug                  bool `mapstructure:"debug"`
	HotReload             bool `mapstructure:"hot_reload"`
	MockExternalServices  bool `mapstructure:"mock_external_services"`
	TestDataEnabled       bool `mapstructure:"test_data_enabled"`
}

var globalConfig *Config

// Load carga la configuración desde archivo y variables de entorno
func Load(configPath string) (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	
	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		viper.AddConfigPath("./configs")
		viper.AddConfigPath("../configs")
		viper.AddConfigPath("../../configs")
		viper.AddConfigPath("/etc/ferre_pos")
	}

	// Configurar variables de entorno
	viper.SetEnvPrefix("FERRE_POS")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Leer archivo de configuración
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil, fmt.Errorf("archivo de configuración no encontrado: %w", err)
		}
		return nil, fmt.Errorf("error leyendo archivo de configuración: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error deserializando configuración: %w", err)
	}

	// Validar configuración
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("configuración inválida: %w", err)
	}

	// Procesar variables de entorno para passwords
	if dbPassword := os.Getenv("FERRE_POS_DATABASE_PASSWORD"); dbPassword != "" {
		config.Database.Password = dbPassword
	}

	globalConfig = &config
	return &config, nil
}

// Get obtiene la configuración global
func Get() *Config {
	if globalConfig == nil {
		panic("configuración no inicializada. Llamar Load() primero")
	}
	return globalConfig
}

// GetAPIConfig obtiene la configuración de una API específica
func GetAPIConfig(apiName string) (*APIConfig, error) {
	config := Get()
	
	switch strings.ToLower(apiName) {
	case "pos":
		return &config.APIs.POS, nil
	case "sync":
		return &config.APIs.Sync, nil
	case "labels":
		return &config.APIs.Labels, nil
	case "reports":
		return &config.APIs.Reports, nil
	default:
		return nil, fmt.Errorf("API no reconocida: %s", apiName)
	}
}

// validateConfig valida la configuración cargada
func validateConfig(config *Config) error {
	// Validar configuración de base de datos
	if config.Database.Host == "" {
		return fmt.Errorf("host de base de datos requerido")
	}
	if config.Database.Port <= 0 || config.Database.Port > 65535 {
		return fmt.Errorf("puerto de base de datos inválido: %d", config.Database.Port)
	}
	if config.Database.Name == "" {
		return fmt.Errorf("nombre de base de datos requerido")
	}
	if config.Database.User == "" {
		return fmt.Errorf("usuario de base de datos requerido")
	}

	// Validar puertos de APIs
	ports := make(map[int]string)
	apis := map[string]APIConfig{
		"pos":     config.APIs.POS,
		"sync":    config.APIs.Sync,
		"labels":  config.APIs.Labels,
		"reports": config.APIs.Reports,
	}

	for name, apiConfig := range apis {
		if !apiConfig.Enabled {
			continue
		}
		
		if apiConfig.Port <= 0 || apiConfig.Port > 65535 {
			return fmt.Errorf("puerto inválido para API %s: %d", name, apiConfig.Port)
		}
		
		if existingAPI, exists := ports[apiConfig.Port]; exists {
			return fmt.Errorf("puerto %d duplicado entre APIs %s y %s", apiConfig.Port, existingAPI, name)
		}
		ports[apiConfig.Port] = name

		// Validar JWT secrets
		if apiConfig.Auth.JWTSecret == "" {
			return fmt.Errorf("JWT secret requerido para API %s", name)
		}
		if len(apiConfig.Auth.JWTSecret) < 32 {
			return fmt.Errorf("JWT secret muy corto para API %s (mínimo 32 caracteres)", name)
		}
	}

	return nil
}

// GetDatabaseDSN construye el DSN de conexión a la base de datos
func (c *Config) GetDatabaseDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.Name,
		c.Database.SSLMode,
	)
}

// IsProduction determina si estamos en ambiente de producción
func (c *Config) IsProduction() bool {
	return !c.Development.Debug
}

// IsDevelopment determina si estamos en ambiente de desarrollo
func (c *Config) IsDevelopment() bool {
	return c.Development.Debug
}

