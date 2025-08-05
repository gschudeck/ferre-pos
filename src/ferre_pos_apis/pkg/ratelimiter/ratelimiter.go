package ratelimiter

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
	"golang.org/x/time/rate"
	
	"ferre_pos_apis/internal/config"
	"ferre_pos_apis/internal/logger"
)

// RateLimiter interfaz principal del rate limiter
type RateLimiter interface {
	Allow(key string) bool
	AllowN(key string, n int) bool
	GinMiddleware() gin.HandlerFunc
	GetLimitInfo(key string) LimitInfo
	SetCustomLimit(key string, requestsPerSecond float64, burstSize int)
	RemoveCustomLimit(key string)
	GetStats() Stats
}

// LimitInfo información sobre límites
type LimitInfo struct {
	RequestsPerSecond float64   `json:"requests_per_second"`
	BurstSize         int       `json:"burst_size"`
	Remaining         int       `json:"remaining"`
	ResetTime         time.Time `json:"reset_time"`
	RetryAfter        int       `json:"retry_after_seconds"`
}

// Stats estadísticas del rate limiter
type Stats struct {
	TotalRequests     int64 `json:"total_requests"`
	AllowedRequests   int64 `json:"allowed_requests"`
	BlockedRequests   int64 `json:"blocked_requests"`
	ActiveLimiters    int   `json:"active_limiters"`
	CustomLimiters    int   `json:"custom_limiters"`
}

// rateLimiterImpl implementación del rate limiter
type rateLimiterImpl struct {
	config         *config.RateLimitConfig
	logger         logger.Logger
	limiters       *cache.Cache
	customLimits   map[string]*config.RateLimitConfig
	customMutex    sync.RWMutex
	stats          *Stats
	statsMutex     sync.RWMutex
	keyGenerator   KeyGenerator
}

// KeyGenerator interfaz para generar claves de rate limiting
type KeyGenerator interface {
	GenerateKey(c *gin.Context) string
}

// IPKeyGenerator generador de claves basado en IP
type IPKeyGenerator struct{}

// GenerateKey genera clave basada en IP del cliente
func (g *IPKeyGenerator) GenerateKey(c *gin.Context) string {
	clientIP := getClientIP(c)
	return fmt.Sprintf("ip:%s", clientIP)
}

// UserKeyGenerator generador de claves basado en usuario
type UserKeyGenerator struct{}

// GenerateKey genera clave basada en usuario autenticado
func (g *UserKeyGenerator) GenerateKey(c *gin.Context) string {
	userID, exists := c.Get("user_id")
	if !exists {
		// Fallback a IP si no hay usuario autenticado
		return (&IPKeyGenerator{}).GenerateKey(c)
	}
	return fmt.Sprintf("user:%v", userID)
}

// CompositeKeyGenerator generador de claves compuesto
type CompositeKeyGenerator struct {
	generators []KeyGenerator
}

// NewCompositeKeyGenerator crea un generador compuesto
func NewCompositeKeyGenerator(generators ...KeyGenerator) *CompositeKeyGenerator {
	return &CompositeKeyGenerator{generators: generators}
}

// GenerateKey genera clave compuesta
func (g *CompositeKeyGenerator) GenerateKey(c *gin.Context) string {
	var keys []string
	for _, generator := range g.generators {
		keys = append(keys, generator.GenerateKey(c))
	}
	return fmt.Sprintf("composite:%v", keys)
}

// New crea una nueva instancia de rate limiter
func New(cfg *config.RateLimitConfig, log logger.Logger) RateLimiter {
	if !cfg.Enabled {
		return &noOpRateLimiter{}
	}

	return &rateLimiterImpl{
		config:       cfg,
		logger:       log,
		limiters:     cache.New(cfg.CleanupInterval, cfg.CleanupInterval*2),
		customLimits: make(map[string]*config.RateLimitConfig),
		stats: &Stats{
			TotalRequests:   0,
			AllowedRequests: 0,
			BlockedRequests: 0,
		},
		keyGenerator: &IPKeyGenerator{}, // Por defecto usar IP
	}
}

// NewWithKeyGenerator crea rate limiter con generador de claves personalizado
func NewWithKeyGenerator(cfg *config.RateLimitConfig, log logger.Logger, keyGen KeyGenerator) RateLimiter {
	if !cfg.Enabled {
		return &noOpRateLimiter{}
	}

	rl := &rateLimiterImpl{
		config:       cfg,
		logger:       log,
		limiters:     cache.New(cfg.CleanupInterval, cfg.CleanupInterval*2),
		customLimits: make(map[string]*config.RateLimitConfig),
		stats: &Stats{
			TotalRequests:   0,
			AllowedRequests: 0,
			BlockedRequests: 0,
		},
		keyGenerator: keyGen,
	}

	return rl
}

// Allow verifica si una request está permitida
func (rl *rateLimiterImpl) Allow(key string) bool {
	return rl.AllowN(key, 1)
}

// AllowN verifica si N requests están permitidas
func (rl *rateLimiterImpl) AllowN(key string, n int) bool {
	rl.updateStats(func(s *Stats) {
		s.TotalRequests += int64(n)
	})

	limiter := rl.getLimiter(key)
	allowed := limiter.AllowN(time.Now(), n)

	if allowed {
		rl.updateStats(func(s *Stats) {
			s.AllowedRequests += int64(n)
		})
	} else {
		rl.updateStats(func(s *Stats) {
			s.BlockedRequests += int64(n)
		})
		
		rl.logger.WithField("key", key).WithField("requests", n).Debug("Rate limit exceeded")
	}

	return allowed
}

// GinMiddleware retorna middleware de Gin para rate limiting
func (rl *rateLimiterImpl) GinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := rl.keyGenerator.GenerateKey(c)
		
		if !rl.Allow(key) {
			limitInfo := rl.GetLimitInfo(key)
			
			// Agregar headers informativos
			c.Header("X-RateLimit-Limit", fmt.Sprintf("%.0f", limitInfo.RequestsPerSecond))
			c.Header("X-RateLimit-Remaining", strconv.Itoa(limitInfo.Remaining))
			c.Header("X-RateLimit-Reset", strconv.FormatInt(limitInfo.ResetTime.Unix(), 10))
			c.Header("Retry-After", strconv.Itoa(limitInfo.RetryAfter))
			
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": gin.H{
					"code":    "RATE_LIMIT_EXCEEDED",
					"message": "Rate limit exceeded",
					"details": gin.H{
						"limit":       limitInfo.RequestsPerSecond,
						"remaining":   limitInfo.Remaining,
						"reset_time":  limitInfo.ResetTime,
						"retry_after": limitInfo.RetryAfter,
					},
				},
				"success": false,
			})
			c.Abort()
			return
		}

		// Agregar headers informativos para requests exitosas
		limitInfo := rl.GetLimitInfo(key)
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%.0f", limitInfo.RequestsPerSecond))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(limitInfo.Remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(limitInfo.ResetTime.Unix(), 10))

		c.Next()
	}
}

// GetLimitInfo obtiene información sobre límites para una clave
func (rl *rateLimiterImpl) GetLimitInfo(key string) LimitInfo {
	limiter := rl.getLimiter(key)
	cfg := rl.getConfigForKey(key)
	
	// Calcular tokens restantes aproximados
	reservation := limiter.Reserve()
	remaining := cfg.BurstSize
	if !reservation.OK() {
		remaining = 0
	} else {
		delay := reservation.Delay()
		if delay > 0 {
			remaining = 0
		}
	}
	reservation.Cancel()

	resetTime := time.Now().Add(time.Second)
	retryAfter := 1

	return LimitInfo{
		RequestsPerSecond: cfg.RequestsPerSecond,
		BurstSize:         cfg.BurstSize,
		Remaining:         remaining,
		ResetTime:         resetTime,
		RetryAfter:        retryAfter,
	}
}

// SetCustomLimit establece límite personalizado para una clave
func (rl *rateLimiterImpl) SetCustomLimit(key string, requestsPerSecond float64, burstSize int) {
	rl.customMutex.Lock()
	defer rl.customMutex.Unlock()

	rl.customLimits[key] = &config.RateLimitConfig{
		Enabled:           true,
		RequestsPerSecond: requestsPerSecond,
		BurstSize:         burstSize,
	}

	// Remover limiter existente para forzar recreación con nuevos límites
	rl.limiters.Delete(key)
	
	rl.logger.WithField("key", key).
		WithField("rps", requestsPerSecond).
		WithField("burst", burstSize).
		Info("Custom rate limit set")
}

// RemoveCustomLimit remueve límite personalizado
func (rl *rateLimiterImpl) RemoveCustomLimit(key string) {
	rl.customMutex.Lock()
	defer rl.customMutex.Unlock()

	delete(rl.customLimits, key)
	rl.limiters.Delete(key)
	
	rl.logger.WithField("key", key).Info("Custom rate limit removed")
}

// GetStats obtiene estadísticas del rate limiter
func (rl *rateLimiterImpl) GetStats() Stats {
	rl.statsMutex.RLock()
	defer rl.statsMutex.RUnlock()

	stats := *rl.stats
	stats.ActiveLimiters = rl.limiters.ItemCount()
	
	rl.customMutex.RLock()
	stats.CustomLimiters = len(rl.customLimits)
	rl.customMutex.RUnlock()

	return stats
}

// getLimiter obtiene o crea un limiter para una clave
func (rl *rateLimiterImpl) getLimiter(key string) *rate.Limiter {
	if limiter, found := rl.limiters.Get(key); found {
		return limiter.(*rate.Limiter)
	}

	cfg := rl.getConfigForKey(key)
	limiter := rate.NewLimiter(rate.Limit(cfg.RequestsPerSecond), cfg.BurstSize)
	
	rl.limiters.Set(key, limiter, cache.DefaultExpiration)
	return limiter
}

// getConfigForKey obtiene configuración para una clave específica
func (rl *rateLimiterImpl) getConfigForKey(key string) *config.RateLimitConfig {
	rl.customMutex.RLock()
	defer rl.customMutex.RUnlock()

	if customCfg, exists := rl.customLimits[key]; exists {
		return customCfg
	}

	return rl.config
}

// updateStats actualiza estadísticas de forma thread-safe
func (rl *rateLimiterImpl) updateStats(fn func(*Stats)) {
	rl.statsMutex.Lock()
	defer rl.statsMutex.Unlock()
	fn(rl.stats)
}

// getClientIP obtiene la IP real del cliente
func getClientIP(c *gin.Context) string {
	// Verificar headers de proxy
	if ip := c.GetHeader("X-Forwarded-For"); ip != "" {
		// X-Forwarded-For puede contener múltiples IPs separadas por coma
		if idx := len(ip); idx > 0 {
			if commaIdx := 0; commaIdx < idx {
				for i, char := range ip {
					if char == ',' {
						commaIdx = i
						break
					}
				}
				if commaIdx > 0 {
					ip = ip[:commaIdx]
				}
			}
		}
		if net.ParseIP(ip) != nil {
			return ip
		}
	}

	if ip := c.GetHeader("X-Real-IP"); ip != "" {
		if net.ParseIP(ip) != nil {
			return ip
		}
	}

	if ip := c.GetHeader("X-Client-IP"); ip != "" {
		if net.ParseIP(ip) != nil {
			return ip
		}
	}

	// Fallback a RemoteAddr
	ip, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		return c.Request.RemoteAddr
	}

	return ip
}

// noOpRateLimiter implementación que no hace nada (cuando está deshabilitado)
type noOpRateLimiter struct{}

func (n *noOpRateLimiter) Allow(key string) bool                                                    { return true }
func (n *noOpRateLimiter) AllowN(key string, num int) bool                                           { return true }
func (n *noOpRateLimiter) GinMiddleware() gin.HandlerFunc                                          { return func(c *gin.Context) { c.Next() } }
func (n *noOpRateLimiter) GetLimitInfo(key string) LimitInfo                                       { return LimitInfo{} }
func (n *noOpRateLimiter) SetCustomLimit(key string, requestsPerSecond float64, burstSize int)    {}
func (n *noOpRateLimiter) RemoveCustomLimit(key string)                                            {}
func (n *noOpRateLimiter) GetStats() Stats                                                         { return Stats{} }

// Funciones de utilidad para crear rate limiters específicos

// NewIPRateLimiter crea rate limiter basado en IP
func NewIPRateLimiter(cfg *config.RateLimitConfig, log logger.Logger) RateLimiter {
	return NewWithKeyGenerator(cfg, log, &IPKeyGenerator{})
}

// NewUserRateLimiter crea rate limiter basado en usuario
func NewUserRateLimiter(cfg *config.RateLimitConfig, log logger.Logger) RateLimiter {
	return NewWithKeyGenerator(cfg, log, &UserKeyGenerator{})
}

// NewCompositeRateLimiter crea rate limiter compuesto
func NewCompositeRateLimiter(cfg *config.RateLimitConfig, log logger.Logger) RateLimiter {
	keyGen := NewCompositeKeyGenerator(&IPKeyGenerator{}, &UserKeyGenerator{})
	return NewWithKeyGenerator(cfg, log, keyGen)
}

// BypassKeyGenerator generador que permite bypass para ciertos usuarios/IPs
type BypassKeyGenerator struct {
	baseGenerator KeyGenerator
	bypassList    map[string]bool
	mutex         sync.RWMutex
}

// NewBypassKeyGenerator crea generador con lista de bypass
func NewBypassKeyGenerator(base KeyGenerator, bypassList []string) *BypassKeyGenerator {
	bypass := make(map[string]bool)
	for _, item := range bypassList {
		bypass[item] = true
	}

	return &BypassKeyGenerator{
		baseGenerator: base,
		bypassList:    bypass,
	}
}

// GenerateKey genera clave verificando bypass
func (g *BypassKeyGenerator) GenerateKey(c *gin.Context) string {
	baseKey := g.baseGenerator.GenerateKey(c)
	
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	
	// Verificar si está en lista de bypass
	clientIP := getClientIP(c)
	if g.bypassList[clientIP] {
		return fmt.Sprintf("bypass:%s", baseKey)
	}

	if userID, exists := c.Get("user_id"); exists {
		if g.bypassList[fmt.Sprintf("user:%v", userID)] {
			return fmt.Sprintf("bypass:%s", baseKey)
		}
	}

	return baseKey
}

// AddBypass agrega elemento a lista de bypass
func (g *BypassKeyGenerator) AddBypass(key string) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.bypassList[key] = true
}

// RemoveBypass remueve elemento de lista de bypass
func (g *BypassKeyGenerator) RemoveBypass(key string) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	delete(g.bypassList, key)
}

// NewBypassRateLimiter crea rate limiter con bypass
func NewBypassRateLimiter(cfg *config.RateLimitConfig, log logger.Logger, bypassList []string) RateLimiter {
	keyGen := NewBypassKeyGenerator(&IPKeyGenerator{}, bypassList)
	rl := NewWithKeyGenerator(cfg, log, keyGen)
	
	// Configurar límites muy altos para claves de bypass
	if impl, ok := rl.(*rateLimiterImpl); ok {
		// Configurar límite alto para bypass
		for _, bypass := range bypassList {
			impl.SetCustomLimit(fmt.Sprintf("bypass:ip:%s", bypass), 10000, 10000)
		}
	}
	
	return rl
}

