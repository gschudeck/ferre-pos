package unit

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ferre_pos_apis/internal/config"
	"ferre_pos_apis/internal/logger"
	"ferre_pos_apis/pkg/ratelimiter"
)

func TestNewIPRateLimiter(t *testing.T) {
	// Configurar logger para tests
	err := logger.Init(&config.LoggingConfig{
		Level:  "error",
		Format: "json",
		Output: "stdout",
	}, "test")
	require.NoError(t, err)

	rateLimitConfig := &config.RateLimitConfig{
		Enabled:           true,
		RequestsPerSecond: 10.0,
		BurstSize:         20,
		CleanupInterval:   time.Minute,
	}

	limiter := ratelimiter.NewIPRateLimiter(rateLimitConfig, logger.Get())
	assert.NotNil(t, limiter)
}

func TestRateLimiterAllow(t *testing.T) {
	rateLimitConfig := &config.RateLimitConfig{
		Enabled:           true,
		RequestsPerSecond: 2.0, // 2 requests per second
		BurstSize:         5,    // burst of 5
		CleanupInterval:   time.Minute,
	}

	limiter := ratelimiter.NewIPRateLimiter(rateLimitConfig, logger.Get())
	
	tests := []struct {
		name     string
		key      string
		requests int
		expected []bool
	}{
		{
			name:     "allow initial burst",
			key:      "192.168.1.1",
			requests: 5,
			expected: []bool{true, true, true, true, true},
		},
		{
			name:     "deny after burst",
			key:      "192.168.1.2",
			requests: 7,
			expected: []bool{true, true, true, true, true, false, false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i, expectedResult := range tt.expected {
				result := limiter.Allow(tt.key)
				assert.Equal(t, expectedResult, result, "Request %d failed", i+1)
			}
		})
	}
}

func TestRateLimiterAllowN(t *testing.T) {
	rateLimitConfig := &config.RateLimitConfig{
		Enabled:           true,
		RequestsPerSecond: 10.0,
		BurstSize:         20,
		CleanupInterval:   time.Minute,
	}

	limiter := ratelimiter.NewIPRateLimiter(rateLimitConfig, logger.Get())
	
	tests := []struct {
		name     string
		key      string
		n        int
		expected bool
	}{
		{
			name:     "allow small batch",
			key:      "192.168.1.3",
			n:        5,
			expected: true,
		},
		{
			name:     "allow exact burst size",
			key:      "192.168.1.4",
			n:        20,
			expected: true,
		},
		{
			name:     "deny over burst size",
			key:      "192.168.1.5",
			n:        25,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := limiter.AllowN(tt.key, tt.n)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRateLimiterGinMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rateLimitConfig := &config.RateLimitConfig{
		Enabled:           true,
		RequestsPerSecond: 1.0, // Very restrictive for testing
		BurstSize:         2,
		CleanupInterval:   time.Minute,
	}

	limiter := ratelimiter.NewIPRateLimiter(rateLimitConfig, logger.Get())
	
	router := gin.New()
	router.Use(limiter.GinMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	tests := []struct {
		name           string
		requests       int
		expectedStatus []int
	}{
		{
			name:           "allow initial requests",
			requests:       2,
			expectedStatus: []int{http.StatusOK, http.StatusOK},
		},
		{
			name:           "rate limit after burst",
			requests:       4,
			expectedStatus: []int{http.StatusOK, http.StatusOK, http.StatusTooManyRequests, http.StatusTooManyRequests},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset rate limiter for each test
			limiter = ratelimiter.NewIPRateLimiter(rateLimitConfig, logger.Get())
			router = gin.New()
			router.Use(limiter.GinMiddleware())
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			for i, expectedStatus := range tt.expectedStatus {
				req, _ := http.NewRequest("GET", "/test", nil)
				req.RemoteAddr = "192.168.1.100:12345" // Same IP for all requests
				
				recorder := httptest.NewRecorder()
				router.ServeHTTP(recorder, req)
				
				assert.Equal(t, expectedStatus, recorder.Code, "Request %d failed", i+1)
			}
		})
	}
}

func TestRateLimiterGetLimitInfo(t *testing.T) {
	rateLimitConfig := &config.RateLimitConfig{
		Enabled:           true,
		RequestsPerSecond: 10.0,
		BurstSize:         20,
		CleanupInterval:   time.Minute,
	}

	limiter := ratelimiter.NewIPRateLimiter(rateLimitConfig, logger.Get())
	
	key := "192.168.1.6"
	
	// Hacer algunas requests
	limiter.Allow(key)
	limiter.Allow(key)
	limiter.Allow(key)
	
	info := limiter.GetLimitInfo(key)
	assert.NotNil(t, info)
	assert.Equal(t, 10.0, info.RequestsPerSecond)
	assert.Equal(t, 20, info.BurstSize)
	assert.True(t, info.RemainingTokens >= 0)
	assert.True(t, info.RemainingTokens <= 20)
}

func TestRateLimiterSetCustomLimit(t *testing.T) {
	rateLimitConfig := &config.RateLimitConfig{
		Enabled:           true,
		RequestsPerSecond: 10.0,
		BurstSize:         20,
		CleanupInterval:   time.Minute,
	}

	limiter := ratelimiter.NewIPRateLimiter(rateLimitConfig, logger.Get())
	
	key := "192.168.1.7"
	
	// Establecer límite personalizado
	limiter.SetCustomLimit(key, 5.0, 10)
	
	info := limiter.GetLimitInfo(key)
	assert.Equal(t, 5.0, info.RequestsPerSecond)
	assert.Equal(t, 10, info.BurstSize)
}

func TestRateLimiterRemoveCustomLimit(t *testing.T) {
	rateLimitConfig := &config.RateLimitConfig{
		Enabled:           true,
		RequestsPerSecond: 10.0,
		BurstSize:         20,
		CleanupInterval:   time.Minute,
	}

	limiter := ratelimiter.NewIPRateLimiter(rateLimitConfig, logger.Get())
	
	key := "192.168.1.8"
	
	// Establecer límite personalizado
	limiter.SetCustomLimit(key, 5.0, 10)
	
	info := limiter.GetLimitInfo(key)
	assert.Equal(t, 5.0, info.RequestsPerSecond)
	
	// Remover límite personalizado
	limiter.RemoveCustomLimit(key)
	
	info = limiter.GetLimitInfo(key)
	assert.Equal(t, 10.0, info.RequestsPerSecond) // Vuelve al default
}

func TestRateLimiterDisabled(t *testing.T) {
	rateLimitConfig := &config.RateLimitConfig{
		Enabled:           false, // Disabled
		RequestsPerSecond: 1.0,
		BurstSize:         1,
		CleanupInterval:   time.Minute,
	}

	limiter := ratelimiter.NewIPRateLimiter(rateLimitConfig, logger.Get())
	
	// Cuando está deshabilitado, siempre debe permitir
	for i := 0; i < 100; i++ {
		result := limiter.Allow("192.168.1.9")
		assert.True(t, result, "Request %d should be allowed when rate limiter is disabled", i+1)
	}
}

func TestRateLimiterConcurrency(t *testing.T) {
	rateLimitConfig := &config.RateLimitConfig{
		Enabled:           true,
		RequestsPerSecond: 100.0,
		BurstSize:         200,
		CleanupInterval:   time.Minute,
	}

	limiter := ratelimiter.NewIPRateLimiter(rateLimitConfig, logger.Get())
	
	// Test concurrencia con múltiples goroutines
	const numGoroutines = 10
	const requestsPerGoroutine = 10
	
	results := make(chan bool, numGoroutines*requestsPerGoroutine)
	
	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			key := "192.168.1.10"
			for j := 0; j < requestsPerGoroutine; j++ {
				result := limiter.Allow(key)
				results <- result
			}
		}(i)
	}
	
	// Recoger resultados
	allowedCount := 0
	for i := 0; i < numGoroutines*requestsPerGoroutine; i++ {
		if <-results {
			allowedCount++
		}
	}
	
	// Debe permitir al menos algunas requests (hasta el burst size)
	assert.Greater(t, allowedCount, 0)
	assert.LessOrEqual(t, allowedCount, 200) // No más que el burst size
}

func TestRateLimiterCleanup(t *testing.T) {
	rateLimitConfig := &config.RateLimitConfig{
		Enabled:           true,
		RequestsPerSecond: 10.0,
		BurstSize:         20,
		CleanupInterval:   100 * time.Millisecond, // Cleanup muy frecuente para testing
	}

	limiter := ratelimiter.NewIPRateLimiter(rateLimitConfig, logger.Get())
	
	// Hacer requests con diferentes IPs
	limiter.Allow("192.168.1.11")
	limiter.Allow("192.168.1.12")
	limiter.Allow("192.168.1.13")
	
	// Esperar a que se ejecute el cleanup
	time.Sleep(200 * time.Millisecond)
	
	// El cleanup debería haber ejecutado (no podemos verificar directamente,
	// pero al menos verificamos que el sistema sigue funcionando)
	result := limiter.Allow("192.168.1.14")
	assert.True(t, result)
}

func BenchmarkRateLimiterAllow(b *testing.B) {
	rateLimitConfig := &config.RateLimitConfig{
		Enabled:           true,
		RequestsPerSecond: 1000.0,
		BurstSize:         2000,
		CleanupInterval:   time.Minute,
	}

	limiter := ratelimiter.NewIPRateLimiter(rateLimitConfig, logger.Get())
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		limiter.Allow("192.168.1.100")
	}
}

func BenchmarkRateLimiterGinMiddleware(b *testing.B) {
	gin.SetMode(gin.TestMode)
	
	rateLimitConfig := &config.RateLimitConfig{
		Enabled:           true,
		RequestsPerSecond: 1000.0,
		BurstSize:         2000,
		CleanupInterval:   time.Minute,
	}

	limiter := ratelimiter.NewIPRateLimiter(rateLimitConfig, logger.Get())
	
	router := gin.New()
	router.Use(limiter.GinMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	
	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.100:12345"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
	}
}

