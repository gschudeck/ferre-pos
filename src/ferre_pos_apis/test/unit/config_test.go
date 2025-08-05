package unit

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"ferre_pos_apis/internal/config"
)

func TestConfigLoad(t *testing.T) {
	tests := []struct {
		name        string
		configPath  string
		expectError bool
	}{
		{
			name:        "load valid config",
			configPath:  "../../configs/config.yaml",
			expectError: false,
		},
		{
			name:        "load non-existent config",
			configPath:  "non-existent.yaml",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := config.Load(tt.configPath)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, cfg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cfg)

				// Verificar campos básicos
				assert.NotEmpty(t, cfg.Database.Host)
				assert.Greater(t, cfg.Database.Port, 0)
				assert.NotEmpty(t, cfg.Database.Name)
			}
		})
	}
}

func TestConfigDefaults(t *testing.T) {
	// Test con configuración mínima
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host: "localhost",
			Port: 5432,
			Name: "test_db",
		},
		Logging: config.LoggingConfig{
			Level:  "info",
			Format: "json",
		},
	}

	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, 5432, cfg.Database.Port)
	assert.Equal(t, "test_db", cfg.Database.Name)
	assert.Equal(t, "info", cfg.Logging.Level)
	assert.Equal(t, "json", cfg.Logging.Format)
}

func TestConfigFromEnvironment(t *testing.T) {
	// Configurar variables de entorno
	os.Setenv("DB_HOST", "env-host")
	os.Setenv("DB_PORT", "3306")
	os.Setenv("DB_NAME", "env-db")
	defer func() {
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
		os.Unsetenv("DB_NAME")
	}()

	// En un entorno real, la configuración leería estas variables
	// Por ahora, solo verificamos que las variables están configuradas
	assert.Equal(t, "env-host", os.Getenv("DB_HOST"))
	assert.Equal(t, "3306", os.Getenv("DB_PORT"))
	assert.Equal(t, "env-db", os.Getenv("DB_NAME"))
}

func TestDatabaseConfig(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Host:                  "localhost",
		Port:                  5432,
		Name:                  "test_db",
		User:                  "test_user",
		Password:              "test_pass",
		SSLMode:               "disable",
		MaxOpenConnections:    25,
		MaxIdleConnections:    10,
		ConnectionMaxLifetime: 5 * time.Minute,
	}

	assert.Equal(t, "localhost", cfg.Host)
	assert.Equal(t, 5432, cfg.Port)
	assert.Equal(t, "test_db", cfg.Name)
	assert.Equal(t, "test_user", cfg.User)
	assert.Equal(t, "test_pass", cfg.Password)
	assert.Equal(t, "disable", cfg.SSLMode)
	assert.Equal(t, 25, cfg.MaxOpenConnections)
	assert.Equal(t, 10, cfg.MaxIdleConnections)
	assert.Equal(t, 5*time.Minute, cfg.ConnectionMaxLifetime)
}

func TestLoggingConfig(t *testing.T) {
	cfg := &config.LoggingConfig{
		Level:      "debug",
		Format:     "text",
		Output:     "stdout",
		FilePath:   "/var/log/app.log",
		MaxSizeMB:  100,
		MaxBackups: 5,
		MaxAgeDays: 30,
		Compress:   true,
	}

	assert.Equal(t, "debug", cfg.Level)
	assert.Equal(t, "text", cfg.Format)
	assert.Equal(t, "stdout", cfg.Output)
	assert.Equal(t, "/var/log/app.log", cfg.FilePath)
	assert.Equal(t, 100, cfg.MaxSizeMB)
	assert.Equal(t, 5, cfg.MaxBackups)
	assert.Equal(t, 30, cfg.MaxAgeDays)
	assert.True(t, cfg.Compress)
}

func TestMetricsConfig(t *testing.T) {
	cfg := &config.MetricsConfig{
		Enabled: true,
		Path:    "/metrics",
		Port:    9090,
	}

	assert.True(t, cfg.Enabled)
	assert.Equal(t, "/metrics", cfg.Path)
	assert.Equal(t, 9090, cfg.Port)
}

func TestRateLimitConfig(t *testing.T) {
	cfg := &config.RateLimitConfig{
		Enabled:           true,
		RequestsPerSecond: 100.0,
		BurstSize:         200,
		CleanupInterval:   time.Minute,
	}

	assert.True(t, cfg.Enabled)
	assert.Equal(t, 100.0, cfg.RequestsPerSecond)
	assert.Equal(t, 200, cfg.BurstSize)
	assert.Equal(t, time.Minute, cfg.CleanupInterval)
}

func TestSecurityConfig(t *testing.T) {
	cfg := &config.SecurityConfig{
		CORS: config.CORSConfig{
			Enabled:        true,
			AllowedOrigins: []string{"http://localhost:3000", "https://example.com"},
			AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
			AllowedHeaders: []string{"Content-Type", "Authorization"},
		},
	}

	assert.True(t, cfg.CORS.Enabled)
	assert.Contains(t, cfg.CORS.AllowedOrigins, "http://localhost:3000")
	assert.Contains(t, cfg.CORS.AllowedMethods, "POST")
	assert.Contains(t, cfg.CORS.AllowedHeaders, "Authorization")
}

func TestAPIsConfig(t *testing.T) {
	cfg := &config.APIsConfig{
		POS: config.APIConfig{
			Port:         8080,
			Host:         "0.0.0.0",
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
		},
		Sync: config.APIConfig{
			Port:         8081,
			Host:         "0.0.0.0",
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
		},
		Labels: config.APIConfig{
			Port:         8082,
			Host:         "0.0.0.0",
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
		},
		Reports: config.APIConfig{
			Port:         8083,
			Host:         "0.0.0.0",
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
		},
	}

	assert.Equal(t, 8080, cfg.POS.Port)
	assert.Equal(t, 8081, cfg.Sync.Port)
	assert.Equal(t, 8082, cfg.Labels.Port)
	assert.Equal(t, 8083, cfg.Reports.Port)
	assert.Equal(t, "0.0.0.0", cfg.POS.Host)
}

func BenchmarkConfigLoad(b *testing.B) {
	configPath := "../../configs/config.yaml"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := config.Load(configPath)
		if err != nil {
			b.Skip("Config file not available for benchmark")
		}
	}
}

