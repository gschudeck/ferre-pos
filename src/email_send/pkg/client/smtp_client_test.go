package client

import (
    "context"
    "testing"
    "time"
    
    "github.com/yourorg/smtp-client/config"
    "github.com/yourorg/smtp-client/pkg/auth"
    "github.com/yourorg/smtp-client/pkg/logger"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

// MockTokenManager mock del token manager
type MockTokenManager struct {
    mock.Mock
}

func (m *MockTokenManager) GetToken(ctx context.Context, requestID string) (*oauth2.Token, error) {
    args := m.Called(ctx, requestID)
    return args.Get(0).(*oauth2.Token), args.Error(1)
}

func (m *MockTokenManager) RefreshToken(ctx context.Context, token *oauth2.Token, requestID string) (*oauth2.Token, error) {
    args := m.Called(ctx, token, requestID)
    return args.Get(0).(*oauth2.Token), args.Error(1)
}

func (m *MockTokenManager) GetAuthURL(state string) string {
    args := m.Called(state)
    return args.String(0)
}

func (m *MockTokenManager) ExchangeCode(ctx context.Context, code, requestID string) (*oauth2.Token, error) {
    args := m.Called(ctx, code, requestID)
    return args.Get(0).(*oauth2.Token), args.Error(1)
}

func TestSMTPClient_ValidateMessage(t *testing.T) {
    // Configuración de prueba
    smtpConfig := &SMTPConfig{
        Host:     "localhost",
        Port:     587,
        UseTLS:   true,
        Provider: auth.ProviderGeneric,
        Username: "test",
        Password: "test",
    }
    
    poolConfig := &PoolConfig{
        MaxConnections: 5,
        MaxIdle:        5 * time.Minute,
        Timeout:        30 * time.Second,
    }
    
    retryConfig := &config.RetryConfig{
        MaxAttempts:    3,
        InitialBackoff: 1 * time.Second,
        MaxBackoff:     30 * time.Second,
        Multiplier:     2.0,
    }
    
    rateLimitConfig := &config.RateLimitingConfig{
        Enabled:           false,
        RequestsPerMinute: 60,
        Burst:            10,
    }
    
    mockLogger := &logger.MockLogger{}
    
    client := NewGenericSMTPClient(smtpConfig, poolConfig, retryConfig, rateLimitConfig, mockLogger)
    
    // Test casos válidos
    t.Run("ValidMessage", func(t *testing.T) {
        message := &EmailMessage{
            From:    "sender@test.com",
            To:      []string{"recipient@test.com"},
            Subject: "Test Subject",
            Body:    "Test Body",
        }
        
        err := client.validator.ValidateMessage(message, "test-request-id")
        assert.NoError(t, err)
    })
    
    // Test casos inválidos
    t.Run("InvalidMessage_NoFrom", func(t *testing.T) {
        message := &EmailMessage{
            To:      []string{"recipient@test.com"},
            Subject: "Test Subject",
            Body:    "Test Body",
        }
        
        err := client.validator.ValidateMessage(message, "test-request-id")
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "From es requerido")
    })
    
    t.Run("InvalidMessage_NoRecipients", func(t *testing.T) {
        message := &EmailMessage{
            From:    "sender@test.com",
            Subject: "Test Subject",
            Body:    "Test Body",
        }
        
        err := client.validator.ValidateMessage(message, "test-request-id")
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "al menos un destinatario")
    })
    
    t.Run("InvalidMessage_BadEmailFormat", func(t *testing.T) {
        message := &EmailMessage{
            From:    "invalid-email",
            To:      []string{"recipient@test.com"},
            Subject: "Test Subject",
            Body:    "Test Body",
        }
        
        err := client.validator.ValidateMessage(message, "test-request-id")
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "formato de email inválido")
    })
}

func TestRetryManager_ShouldRetry(t *testing.T) {
    retryConfig := &config.RetryConfig{
        MaxAttempts:    3,
        InitialBackoff: 1 * time.Second,
        MaxBackoff:     30 * time.Second,
        Multiplier:     2.0,
    }
    
    mockLogger := &logger.MockLogger{}
    retryManager := NewRetryManager(retryConfig, mockLogger)
    
    // Test casos que deben reintentarse
    t.Run("ShouldRetry_ConnectionError", func(t *testing.T) {
        err := &errors.SMTPError{Type: errors.ErrorTypeConnection}
        assert.True(t, retryManager.shouldRetry(err))
    })
    
    t.Run("ShouldRetry_TimeoutError", func(t *testing.T) {
        err := &errors.SMTPError{Type: errors.ErrorTypeTimeout}
        assert.True(t, retryManager.shouldRetry(err))
    })
    
    // Test casos que NO deben reintentarse
    t.Run("ShouldNotRetry_ValidationError", func(t *testing.T) {
        err := &errors.SMTPError{Type: errors.ErrorTypeValidation}
        assert.False(t, retryManager.shouldRetry(err))
    })
}

func TestRateLimiter(t *testing.T) {
    // Test rate limiter deshabilitado
    t.Run("Disabled", func(t *testing.T) {
        rl := NewRateLimiter(60, 10, false)
        assert.True(t, rl.Allow())
        assert.True(t, rl.Allow())
        assert.True(t, rl.Allow())
    })
    
    // Test rate limiter habilitado
    t.Run("Enabled", func(t *testing.T) {
        rl := NewRateLimiter(1, 1, true) // 1 request per minute, burst 1
        
        // Primer request debe ser permitido
        assert.True(t, rl.Allow())
        
        // Segundo request inmediato debe ser rechazado
        assert.False(t, rl.Allow())
    })
    
    // Test actualización de límites
    t.Run("UpdateLimits", func(t *testing.T) {
        rl := NewRateLimiter(1, 1, true)
        
        // Actualizar a límites más altos
        rl.UpdateLimits(60, 10)
        
        // Ahora debe permitir múltiples requests
        assert.True(t, rl.Allow())
        assert.True(t, rl.Allow())
    })
}
