package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
	"ferre-pos-servidor-central/internal/database"
	"ferre-pos-servidor-central/internal/middleware"
)

// Config contiene toda la configuración del sistema
type Config struct {
	Server     ServerConfig     `yaml:"server" json:"server"`
	Database   DatabaseConfigs  `yaml:"database" json:"database"`
	APIs       APIConfigs       `yaml:"apis" json:"apis"`
	Security   SecurityConfig   `yaml:"security" json:"security"`
	Logging    LoggingConfigs   `yaml:"logging" json:"logging"`
	Cache      CacheConfig      `yaml:"cache" json:"cache"`
	Email      EmailConfig      `yaml:"email" json:"email"`
	SMS        SMSConfig        `yaml:"sms" json:"sms"`
	Storage    StorageConfig    `yaml:"storage" json:"storage"`
	Monitoring MonitoringConfig `yaml:"monitoring" json:"monitoring"`
}

// ServerConfig configuración del servidor
type ServerConfig struct {
	Host            string        `yaml:"host" json:"host"`
	Port            int           `yaml:"port" json:"port"`
	Mode            string        `yaml:"mode" json:"mode"` // "debug", "release", "test"
	ReadTimeout     time.Duration `yaml:"read_timeout" json:"read_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout" json:"write_timeout"`
	IdleTimeout     time.Duration `yaml:"idle_timeout" json:"idle_timeout"`
	MaxHeaderBytes  int           `yaml:"max_header_bytes" json:"max_header_bytes"`
	TrustedProxies  []string      `yaml:"trusted_proxies" json:"trusted_proxies"`
	EnableProfiling bool          `yaml:"enable_profiling" json:"enable_profiling"`
	GracefulTimeout time.Duration `yaml:"graceful_timeout" json:"graceful_timeout"`
}

// DatabaseConfigs configuraciones de base de datos por API
type DatabaseConfigs struct {
	POS     database.DatabaseConfig `yaml:"pos" json:"pos"`
	Sync    database.DatabaseConfig `yaml:"sync" json:"sync"`
	Labels  database.DatabaseConfig `yaml:"labels" json:"labels"`
	Reports database.DatabaseConfig `yaml:"reports" json:"reports"`
}

// APIConfigs configuraciones específicas por API
type APIConfigs struct {
	POS     POSConfig     `yaml:"pos" json:"pos"`
	Sync    SyncConfig    `yaml:"sync" json:"sync"`
	Labels  LabelsConfig  `yaml:"labels" json:"labels"`
	Reports ReportsConfig `yaml:"reports" json:"reports"`
}

// POSConfig configuración del API POS
type POSConfig struct {
	Enabled              bool          `yaml:"enabled" json:"enabled"`
	BasePath             string        `yaml:"base_path" json:"base_path"`
	MaxConcurrentUsers   int           `yaml:"max_concurrent_users" json:"max_concurrent_users"`
	SessionTimeout       time.Duration `yaml:"session_timeout" json:"session_timeout"`
	MaxProductsPerQuery  int           `yaml:"max_products_per_query" json:"max_products_per_query"`
	CacheProductsTTL     time.Duration `yaml:"cache_products_ttl" json:"cache_products_ttl"`
	AllowOfflineMode     bool          `yaml:"allow_offline_mode" json:"allow_offline_mode"`
	OfflineDataRetention time.Duration `yaml:"offline_data_retention" json:"offline_data_retention"`
	RequireTerminalAuth  bool          `yaml:"require_terminal_auth" json:"require_terminal_auth"`
	MaxVentaItems        int           `yaml:"max_venta_items" json:"max_venta_items"`
	EnableFidelizacion   bool          `yaml:"enable_fidelizacion" json:"enable_fidelizacion"`
	RateLimiting         RateLimitConfig `yaml:"rate_limiting" json:"rate_limiting"`
}

// SyncConfig configuración del API Sync
type SyncConfig struct {
	Enabled                bool          `yaml:"enabled" json:"enabled"`
	BasePath               string        `yaml:"base_path" json:"base_path"`
	MaxConcurrentSyncs     int           `yaml:"max_concurrent_syncs" json:"max_concurrent_syncs"`
	SyncInterval           time.Duration `yaml:"sync_interval" json:"sync_interval"`
	BatchSize              int           `yaml:"batch_size" json:"batch_size"`
	MaxRetries             int           `yaml:"max_retries" json:"max_retries"`
	RetryBackoffMultiplier float64       `yaml:"retry_backoff_multiplier" json:"retry_backoff_multiplier"`
	ConflictResolutionMode string        `yaml:"conflict_resolution_mode" json:"conflict_resolution_mode"`
	EnableCompression      bool          `yaml:"enable_compression" json:"enable_compression"`
	MaxSyncDuration        time.Duration `yaml:"max_sync_duration" json:"max_sync_duration"`
	CleanupInterval        time.Duration `yaml:"cleanup_interval" json:"cleanup_interval"`
	LogRetentionDays       int           `yaml:"log_retention_days" json:"log_retention_days"`
}

// LabelsConfig configuración del API Labels
type LabelsConfig struct {
	Enabled              bool          `yaml:"enabled" json:"enabled"`
	BasePath             string        `yaml:"base_path" json:"base_path"`
	MaxConcurrentJobs    int           `yaml:"max_concurrent_jobs" json:"max_concurrent_jobs"`
	MaxLabelsPerBatch    int           `yaml:"max_labels_per_batch" json:"max_labels_per_batch"`
	DefaultLabelFormat   string        `yaml:"default_label_format" json:"default_label_format"`
	SupportedFormats     []string      `yaml:"supported_formats" json:"supported_formats"`
	MaxLabelSize         int           `yaml:"max_label_size" json:"max_label_size"` // KB
	CacheTemplatesTTL    time.Duration `yaml:"cache_templates_ttl" json:"cache_templates_ttl"`
	StoragePath          string        `yaml:"storage_path" json:"storage_path"`
	CleanupInterval      time.Duration `yaml:"cleanup_interval" json:"cleanup_interval"`
	FileRetentionDays    int           `yaml:"file_retention_days" json:"file_retention_days"`
	EnablePreview        bool          `yaml:"enable_preview" json:"enable_preview"`
	PreviewTimeout       time.Duration `yaml:"preview_timeout" json:"preview_timeout"`
}

// ReportsConfig configuración del API Reports
type ReportsConfig struct {
	Enabled               bool          `yaml:"enabled" json:"enabled"`
	BasePath              string        `yaml:"base_path" json:"base_path"`
	MaxConcurrentReports  int           `yaml:"max_concurrent_reports" json:"max_concurrent_reports"`
	MaxReportSize         int           `yaml:"max_report_size" json:"max_report_size"` // MB
	DefaultFormat         string        `yaml:"default_format" json:"default_format"`
	SupportedFormats      []string      `yaml:"supported_formats" json:"supported_formats"`
	CacheReportsTTL       time.Duration `yaml:"cache_reports_ttl" json:"cache_reports_ttl"`
	StoragePath           string        `yaml:"storage_path" json:"storage_path"`
	CleanupInterval       time.Duration `yaml:"cleanup_interval" json:"cleanup_interval"`
	FileRetentionDays     int           `yaml:"file_retention_days" json:"file_retention_days"`
	EnableScheduledReports bool         `yaml:"enable_scheduled_reports" json:"enable_scheduled_reports"`
	MaxScheduledReports   int           `yaml:"max_scheduled_reports" json:"max_scheduled_reports"`
	ReportTimeout         time.Duration `yaml:"report_timeout" json:"report_timeout"`
	EnableDashboards      bool          `yaml:"enable_dashboards" json:"enable_dashboards"`
	DashboardRefreshRate  time.Duration `yaml:"dashboard_refresh_rate" json:"dashboard_refresh_rate"`
}

// SecurityConfig configuración de seguridad
type SecurityConfig struct {
	JWTSecret           string        `yaml:"jwt_secret" json:"jwt_secret"`
	JWTExpiration       time.Duration `yaml:"jwt_expiration" json:"jwt_expiration"`
	RefreshTokenExpiration time.Duration `yaml:"refresh_token_expiration" json:"refresh_token_expiration"`
	PasswordMinLength   int           `yaml:"password_min_length" json:"password_min_length"`
	PasswordRequireSpecial bool       `yaml:"password_require_special" json:"password_require_special"`
	MaxLoginAttempts    int           `yaml:"max_login_attempts" json:"max_login_attempts"`
	LoginLockoutDuration time.Duration `yaml:"login_lockout_duration" json:"login_lockout_duration"`
	EnableTwoFactor     bool          `yaml:"enable_two_factor" json:"enable_two_factor"`
	APIKeys             []string      `yaml:"api_keys" json:"api_keys"`
	AllowedIPs          []string      `yaml:"allowed_ips" json:"allowed_ips"`
	CORS                CORSConfigs   `yaml:"cors" json:"cors"`
	RateLimiting        RateLimitConfig `yaml:"rate_limiting" json:"rate_limiting"`
}

// CORSConfigs configuraciones CORS por API
type CORSConfigs struct {
	POS     middleware.CORSConfig `yaml:"pos" json:"pos"`
	Sync    middleware.CORSConfig `yaml:"sync" json:"sync"`
	Labels  middleware.CORSConfig `yaml:"labels" json:"labels"`
	Reports middleware.CORSConfig `yaml:"reports" json:"reports"`
}

// RateLimitConfig configuración de rate limiting
type RateLimitConfig struct {
	Enabled        bool `yaml:"enabled" json:"enabled"`
	RequestsPerMinute int `yaml:"requests_per_minute" json:"requests_per_minute"`
	BurstSize      int  `yaml:"burst_size" json:"burst_size"`
	CleanupInterval time.Duration `yaml:"cleanup_interval" json:"cleanup_interval"`
}

// LoggingConfigs configuraciones de logging por API
type LoggingConfigs struct {
	Global  middleware.LoggingConfig `yaml:"global" json:"global"`
	POS     middleware.LoggingConfig `yaml:"pos" json:"pos"`
	Sync    middleware.LoggingConfig `yaml:"sync" json:"sync"`
	Labels  middleware.LoggingConfig `yaml:"labels" json:"labels"`
	Reports middleware.LoggingConfig `yaml:"reports" json:"reports"`
}

// CacheConfig configuración de cache
type CacheConfig struct {
	Type           string        `yaml:"type" json:"type"` // "memory", "redis"
	Host           string        `yaml:"host" json:"host"`
	Port           int           `yaml:"port" json:"port"`
	Password       string        `yaml:"password" json:"password"`
	Database       int           `yaml:"database" json:"database"`
	MaxConnections int           `yaml:"max_connections" json:"max_connections"`
	DefaultTTL     time.Duration `yaml:"default_ttl" json:"default_ttl"`
	CleanupInterval time.Duration `yaml:"cleanup_interval" json:"cleanup_interval"`
}

// EmailConfig configuración de email
type EmailConfig struct {
	Enabled    bool   `yaml:"enabled" json:"enabled"`
	SMTPHost   string `yaml:"smtp_host" json:"smtp_host"`
	SMTPPort   int    `yaml:"smtp_port" json:"smtp_port"`
	Username   string `yaml:"username" json:"username"`
	Password   string `yaml:"password" json:"password"`
	FromEmail  string `yaml:"from_email" json:"from_email"`
	FromName   string `yaml:"from_name" json:"from_name"`
	UseTLS     bool   `yaml:"use_tls" json:"use_tls"`
	UseSSL     bool   `yaml:"use_ssl" json:"use_ssl"`
}

// SMSConfig configuración de SMS
type SMSConfig struct {
	Enabled   bool   `yaml:"enabled" json:"enabled"`
	Provider  string `yaml:"provider" json:"provider"` // "twilio", "aws"
	APIKey    string `yaml:"api_key" json:"api_key"`
	APISecret string `yaml:"api_secret" json:"api_secret"`
	FromNumber string `yaml:"from_number" json:"from_number"`
}

// StorageConfig configuración de almacenamiento
type StorageConfig struct {
	Type        string `yaml:"type" json:"type"` // "local", "s3", "gcs"
	BasePath    string `yaml:"base_path" json:"base_path"`
	MaxFileSize int    `yaml:"max_file_size" json:"max_file_size"` // MB
	AllowedExtensions []string `yaml:"allowed_extensions" json:"allowed_extensions"`
	
	// S3 Config
	S3Bucket    string `yaml:"s3_bucket" json:"s3_bucket"`
	S3Region    string `yaml:"s3_region" json:"s3_region"`
	S3AccessKey string `yaml:"s3_access_key" json:"s3_access_key"`
	S3SecretKey string `yaml:"s3_secret_key" json:"s3_secret_key"`
}

// MonitoringConfig configuración de monitoreo
type MonitoringConfig struct {
	Enabled         bool          `yaml:"enabled" json:"enabled"`
	MetricsPath     string        `yaml:"metrics_path" json:"metrics_path"`
	HealthPath      string        `yaml:"health_path" json:"health_path"`
	EnablePprof     bool          `yaml:"enable_pprof" json:"enable_pprof"`
	CollectInterval time.Duration `yaml:"collect_interval" json:"collect_interval"`
	RetentionDays   int           `yaml:"retention_days" json:"retention_days"`
}

// ConfigManager gestiona la configuración del sistema
type ConfigManager struct {
	config     *Config
	configPath string
	mutex      sync.RWMutex
	watchers   []func(*Config)
}

// NewConfigManager crea un nuevo gestor de configuración
func NewConfigManager(configPath string) *ConfigManager {
	return &ConfigManager{
		configPath: configPath,
		watchers:   make([]func(*Config), 0),
	}
}

// LoadConfig carga la configuración desde archivo
func (cm *ConfigManager) LoadConfig() error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// Verificar si el archivo existe
	if _, err := os.Stat(cm.configPath); os.IsNotExist(err) {
		// Crear configuración por defecto
		cm.config = DefaultConfig()
		return cm.SaveConfig()
	}

	// Leer archivo
	data, err := os.ReadFile(cm.configPath)
	if err != nil {
		return fmt.Errorf("error leyendo archivo de configuración: %w", err)
	}

	// Parsear YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("error parseando configuración: %w", err)
	}

	cm.config = &config
	
	// Notificar watchers
	cm.notifyWatchers()
	
	return nil
}

// SaveConfig guarda la configuración actual al archivo
func (cm *ConfigManager) SaveConfig() error {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	if cm.config == nil {
		return fmt.Errorf("no hay configuración para guardar")
	}

	// Crear directorio si no existe
	dir := filepath.Dir(cm.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("error creando directorio: %w", err)
	}

	// Convertir a YAML
	data, err := yaml.Marshal(cm.config)
	if err != nil {
		return fmt.Errorf("error serializando configuración: %w", err)
	}

	// Escribir archivo
	if err := os.WriteFile(cm.configPath, data, 0644); err != nil {
		return fmt.Errorf("error escribiendo archivo: %w", err)
	}

	return nil
}

// GetConfig obtiene la configuración actual
func (cm *ConfigManager) GetConfig() *Config {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	return cm.config
}

// UpdateConfig actualiza la configuración
func (cm *ConfigManager) UpdateConfig(newConfig *Config) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cm.config = newConfig
	
	// Guardar cambios
	if err := cm.SaveConfig(); err != nil {
		return err
	}
	
	// Notificar watchers
	cm.notifyWatchers()
	
	return nil
}

// AddWatcher agrega un watcher para cambios de configuración
func (cm *ConfigManager) AddWatcher(watcher func(*Config)) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	cm.watchers = append(cm.watchers, watcher)
}

// notifyWatchers notifica a todos los watchers
func (cm *ConfigManager) notifyWatchers() {
	for _, watcher := range cm.watchers {
		go watcher(cm.config)
	}
}

// ReloadConfig recarga la configuración desde archivo
func (cm *ConfigManager) ReloadConfig() error {
	return cm.LoadConfig()
}

// ValidateConfig valida la configuración actual
func (cm *ConfigManager) ValidateConfig() error {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	if cm.config == nil {
		return fmt.Errorf("configuración no cargada")
	}

	// Validaciones básicas
	if cm.config.Server.Port <= 0 || cm.config.Server.Port > 65535 {
		return fmt.Errorf("puerto del servidor inválido: %d", cm.config.Server.Port)
	}

	if cm.config.Security.JWTSecret == "" {
		return fmt.Errorf("JWT secret no puede estar vacío")
	}

	if cm.config.Security.JWTExpiration <= 0 {
		return fmt.Errorf("expiración JWT debe ser mayor a 0")
	}

	// Validar configuraciones de base de datos
	if err := cm.validateDatabaseConfigs(); err != nil {
		return fmt.Errorf("error en configuración de base de datos: %w", err)
	}

	return nil
}

// validateDatabaseConfigs valida las configuraciones de base de datos
func (cm *ConfigManager) validateDatabaseConfigs() error {
	configs := map[string]database.DatabaseConfig{
		"pos":     cm.config.Database.POS,
		"sync":    cm.config.Database.Sync,
		"labels":  cm.config.Database.Labels,
		"reports": cm.config.Database.Reports,
	}

	for name, config := range configs {
		if config.Host == "" {
			return fmt.Errorf("host de base de datos %s no puede estar vacío", name)
		}
		if config.Port <= 0 {
			return fmt.Errorf("puerto de base de datos %s inválido: %d", name, config.Port)
		}
		if config.Database == "" {
			return fmt.Errorf("nombre de base de datos %s no puede estar vacío", name)
		}
	}

	return nil
}

// GetAPIConfig obtiene configuración específica de una API
func (cm *ConfigManager) GetAPIConfig(apiName string) interface{} {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	switch apiName {
	case "pos":
		return cm.config.APIs.POS
	case "sync":
		return cm.config.APIs.Sync
	case "labels":
		return cm.config.APIs.Labels
	case "reports":
		return cm.config.APIs.Reports
	default:
		return nil
	}
}

// UpdateAPIConfig actualiza configuración específica de una API
func (cm *ConfigManager) UpdateAPIConfig(apiName string, config interface{}) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	switch apiName {
	case "pos":
		if posConfig, ok := config.(POSConfig); ok {
			cm.config.APIs.POS = posConfig
		} else {
			return fmt.Errorf("tipo de configuración inválido para API POS")
		}
	case "sync":
		if syncConfig, ok := config.(SyncConfig); ok {
			cm.config.APIs.Sync = syncConfig
		} else {
			return fmt.Errorf("tipo de configuración inválido para API Sync")
		}
	case "labels":
		if labelsConfig, ok := config.(LabelsConfig); ok {
			cm.config.APIs.Labels = labelsConfig
		} else {
			return fmt.Errorf("tipo de configuración inválido para API Labels")
		}
	case "reports":
		if reportsConfig, ok := config.(ReportsConfig); ok {
			cm.config.APIs.Reports = reportsConfig
		} else {
			return fmt.Errorf("tipo de configuración inválido para API Reports")
		}
	default:
		return fmt.Errorf("API desconocida: %s", apiName)
	}

	// Guardar cambios
	if err := cm.SaveConfig(); err != nil {
		return err
	}

	// Notificar watchers
	cm.notifyWatchers()

	return nil
}

// DefaultConfig retorna una configuración por defecto
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host:            "0.0.0.0",
			Port:            8080,
			Mode:            "release",
			ReadTimeout:     30 * time.Second,
			WriteTimeout:    30 * time.Second,
			IdleTimeout:     60 * time.Second,
			MaxHeaderBytes:  1 << 20, // 1MB
			TrustedProxies:  []string{"127.0.0.1"},
			EnableProfiling: false,
			GracefulTimeout: 30 * time.Second,
		},
		Database: DatabaseConfigs{
			POS:     database.DatabaseConfigForAPI("pos", database.DefaultDatabaseConfig()),
			Sync:    database.DatabaseConfigForAPI("sync", database.DefaultDatabaseConfig()),
			Labels:  database.DatabaseConfigForAPI("labels", database.DefaultDatabaseConfig()),
			Reports: database.DatabaseConfigForAPI("reports", database.DefaultDatabaseConfig()),
		},
		APIs: APIConfigs{
			POS:     DefaultPOSConfig(),
			Sync:    DefaultSyncConfig(),
			Labels:  DefaultLabelsConfig(),
			Reports: DefaultReportsConfig(),
		},
		Security: SecurityConfig{
			JWTSecret:              "your-super-secret-jwt-key-change-in-production",
			JWTExpiration:          24 * time.Hour,
			RefreshTokenExpiration: 7 * 24 * time.Hour,
			PasswordMinLength:      8,
			PasswordRequireSpecial: true,
			MaxLoginAttempts:       5,
			LoginLockoutDuration:   15 * time.Minute,
			EnableTwoFactor:        false,
			APIKeys:                []string{},
			AllowedIPs:             []string{"0.0.0.0"},
			CORS: CORSConfigs{
				POS:     middleware.CORSForAPI("pos"),
				Sync:    middleware.CORSForAPI("sync"),
				Labels:  middleware.CORSForAPI("labels"),
				Reports: middleware.CORSForAPI("reports"),
			},
			RateLimiting: RateLimitConfig{
				Enabled:           true,
				RequestsPerMinute: 100,
				BurstSize:         10,
				CleanupInterval:   5 * time.Minute,
			},
		},
		Logging: LoggingConfigs{
			Global:  middleware.DefaultLoggingConfig(),
			POS:     middleware.DefaultLoggingConfig(),
			Sync:    middleware.DefaultLoggingConfig(),
			Labels:  middleware.DefaultLoggingConfig(),
			Reports: middleware.DefaultLoggingConfig(),
		},
		Cache: CacheConfig{
			Type:            "memory",
			Host:            "localhost",
			Port:            6379,
			Password:        "",
			Database:        0,
			MaxConnections:  10,
			DefaultTTL:      1 * time.Hour,
			CleanupInterval: 10 * time.Minute,
		},
		Email: EmailConfig{
			Enabled:   false,
			SMTPHost:  "smtp.gmail.com",
			SMTPPort:  587,
			Username:  "",
			Password:  "",
			FromEmail: "noreply@ferreteria.com",
			FromName:  "Ferre-POS",
			UseTLS:    true,
			UseSSL:    false,
		},
		SMS: SMSConfig{
			Enabled:    false,
			Provider:   "twilio",
			APIKey:     "",
			APISecret:  "",
			FromNumber: "",
		},
		Storage: StorageConfig{
			Type:        "local",
			BasePath:    "/var/lib/ferre-pos/storage",
			MaxFileSize: 10, // 10MB
			AllowedExtensions: []string{".pdf", ".png", ".jpg", ".jpeg", ".xlsx", ".csv"},
			S3Bucket:    "",
			S3Region:    "us-east-1",
			S3AccessKey: "",
			S3SecretKey: "",
		},
		Monitoring: MonitoringConfig{
			Enabled:         true,
			MetricsPath:     "/metrics",
			HealthPath:      "/health",
			EnablePprof:     false,
			CollectInterval: 30 * time.Second,
			RetentionDays:   30,
		},
	}
}

// DefaultPOSConfig configuración por defecto para API POS
func DefaultPOSConfig() POSConfig {
	return POSConfig{
		Enabled:              true,
		BasePath:             "/api/pos",
		MaxConcurrentUsers:   100,
		SessionTimeout:       8 * time.Hour,
		MaxProductsPerQuery:  1000,
		CacheProductsTTL:     30 * time.Minute,
		AllowOfflineMode:     true,
		OfflineDataRetention: 7 * 24 * time.Hour,
		RequireTerminalAuth:  true,
		MaxVentaItems:        100,
		EnableFidelizacion:   true,
		RateLimiting: RateLimitConfig{
			Enabled:           true,
			RequestsPerMinute: 200,
			BurstSize:         20,
			CleanupInterval:   5 * time.Minute,
		},
	}
}

// DefaultSyncConfig configuración por defecto para API Sync
func DefaultSyncConfig() SyncConfig {
	return SyncConfig{
		Enabled:                true,
		BasePath:               "/api/sync",
		MaxConcurrentSyncs:     5,
		SyncInterval:           15 * time.Minute,
		BatchSize:              100,
		MaxRetries:             3,
		RetryBackoffMultiplier: 2.0,
		ConflictResolutionMode: "manual",
		EnableCompression:      true,
		MaxSyncDuration:        30 * time.Minute,
		CleanupInterval:        24 * time.Hour,
		LogRetentionDays:       90,
	}
}

// DefaultLabelsConfig configuración por defecto para API Labels
func DefaultLabelsConfig() LabelsConfig {
	return LabelsConfig{
		Enabled:           true,
		BasePath:          "/api/labels",
		MaxConcurrentJobs: 10,
		MaxLabelsPerBatch: 1000,
		DefaultLabelFormat: "pdf",
		SupportedFormats:  []string{"pdf", "png", "jpg"},
		MaxLabelSize:      5120, // 5MB
		CacheTemplatesTTL: 2 * time.Hour,
		StoragePath:       "/var/lib/ferre-pos/labels",
		CleanupInterval:   24 * time.Hour,
		FileRetentionDays: 30,
		EnablePreview:     true,
		PreviewTimeout:    30 * time.Second,
	}
}

// DefaultReportsConfig configuración por defecto para API Reports
func DefaultReportsConfig() ReportsConfig {
	return ReportsConfig{
		Enabled:               true,
		BasePath:              "/api/reports",
		MaxConcurrentReports:  5,
		MaxReportSize:         100, // 100MB
		DefaultFormat:         "pdf",
		SupportedFormats:      []string{"pdf", "excel", "csv", "json"},
		CacheReportsTTL:       1 * time.Hour,
		StoragePath:           "/var/lib/ferre-pos/reports",
		CleanupInterval:       24 * time.Hour,
		FileRetentionDays:     90,
		EnableScheduledReports: true,
		MaxScheduledReports:   50,
		ReportTimeout:         10 * time.Minute,
		EnableDashboards:      true,
		DashboardRefreshRate:  5 * time.Minute,
	}
}

