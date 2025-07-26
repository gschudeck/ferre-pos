// Package middleware proporciona middleware de rate limiting avanzado
// con notación húngara y algoritmos sofisticados
package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ferre-pos-servidor-central/pkg/concurrency"
	"ferre-pos-servidor-central/pkg/utils"
)

// structRateLimitConfig configuración del rate limiting
type structRateLimitConfig struct {
	BoolEnabled           bool                           `json:"enabled"`
	IntDefaultLimit       int                            `json:"default_limit"`
	DurationDefaultWindow time.Duration                  `json:"default_window"`
	MapEndpointLimits     map[string]structEndpointLimit `json:"endpoint_limits"`
	MapUserTypeLimits     map[string]structUserTypeLimit `json:"user_type_limits"`
	ArrWhitelistIPs       []string                       `json:"whitelist_ips"`
	ArrBlacklistIPs       []string                       `json:"blacklist_ips"`
	BoolUseRedis          bool                           `json:"use_redis"`
	StrRedisAddr          string                         `json:"redis_addr"`
	BoolLogViolations     bool                           `json:"log_violations"`
	BoolBlockOnViolation  bool                           `json:"block_on_violation"`
	DurationBlockDuration time.Duration                  `json:"block_duration"`
	IntBurstLimit         int                            `json:"burst_limit"`
	EnumAlgorithm         enumRateLimitAlgorithm         `json:"algorithm"`
}

// structEndpointLimit límites específicos por endpoint
type structEndpointLimit struct {
	StrMethod       string        `json:"method"`
	StrPath         string        `json:"path"`
	IntLimit        int           `json:"limit"`
	DurationWindow  time.Duration `json:"window"`
	IntBurstLimit   int           `json:"burst_limit"`
	BoolRequireAuth bool          `json:"require_auth"`
}

// structUserTypeLimit límites por tipo de usuario
type structUserTypeLimit struct {
	StrUserType    string        `json:"user_type"`
	IntLimit       int           `json:"limit"`
	DurationWindow time.Duration `json:"window"`
	IntBurstLimit  int           `json:"burst_limit"`
	FltMultiplier  float64       `json:"multiplier"`
}

// enumRateLimitAlgorithm algoritmos de rate limiting
type enumRateLimitAlgorithm string

const (
	EnumRateLimitAlgorithmTokenBucket   enumRateLimitAlgorithm = "token_bucket"
	EnumRateLimitAlgorithmSlidingWindow enumRateLimitAlgorithm = "sliding_window"
	EnumRateLimitAlgorithmFixedWindow   enumRateLimitAlgorithm = "fixed_window"
	EnumRateLimitAlgorithmLeakyBucket   enumRateLimitAlgorithm = "leaky_bucket"
)

// structRateLimiter limitador de velocidad principal
type structRateLimiter struct {
	PtrConfig     *structRateLimitConfig
	PtrLogger     logger.interfaceLogger
	PtrMetrics    metrics.interfaceManager
	MapLimiters   map[string]interfaceAlgorithmLimiter
	MutexLimiters sync.RWMutex
	MapBlockedIPs concurrency.interfaceSafeMap
	MapViolations concurrency.interfaceSafeMap
	ChanCleanup   chan struct{}
	BoolRunning   bool
}

// interfaceAlgorithmLimiter interfaz para algoritmos de rate limiting
type interfaceAlgorithmLimiter interface {
	Allow(strKey string, intTokens int) bool
	GetRemaining(strKey string) int
	Reset(strKey string)
	GetStats() map[string]interface{}
}

// structTokenBucket implementación de token bucket
type structTokenBucket struct {
	IntCapacity        int
	IntTokens          int
	DurationRefillRate time.Duration
	TimeLastRefill     time.Time
	MutexBucket        sync.Mutex
}

// structSlidingWindow implementación de sliding window
type structSlidingWindow struct {
	IntLimit       int
	DurationWindow time.Duration
	ArrRequests    []time.Time
	MutexWindow    sync.Mutex
}

// structFixedWindow implementación de fixed window
type structFixedWindow struct {
	IntLimit        int
	DurationWindow  time.Duration
	IntCurrentCount int
	TimeWindowStart time.Time
	MutexWindow     sync.Mutex
}

// structLeakyBucket implementación de leaky bucket
type structLeakyBucket struct {
	IntCapacity      int
	IntCurrentLevel  int
	DurationLeakRate time.Duration
	TimeLastLeak     time.Time
	MutexBucket      sync.Mutex
}

// structRateLimitInfo información de rate limiting
type structRateLimitInfo struct {
	StrKey         string        `json:"key"`
	IntLimit       int           `json:"limit"`
	IntRemaining   int           `json:"remaining"`
	TimeResetTime  time.Time     `json:"reset_time"`
	DurationWindow time.Duration `json:"window"`
	BoolAllowed    bool          `json:"allowed"`
	StrReason      string        `json:"reason,omitempty"`
}

// structViolationInfo información de violación
type structViolationInfo struct {
	StrIP              string    `json:"ip"`
	StrUserID          string    `json:"user_id,omitempty"`
	StrEndpoint        string    `json:"endpoint"`
	IntCount           int       `json:"count"`
	TimeFirstViolation time.Time `json:"first_violation"`
	TimeLastViolation  time.Time `json:"last_violation"`
	BoolBlocked        bool      `json:"blocked"`
}

// RateLimitMiddleware middleware principal de rate limiting
func RateLimitMiddleware(ptrConfig *structRateLimitConfig, ptrLogger logger.interfaceLogger, ptrMetrics metrics.interfaceManager) gin.HandlerFunc {
	if ptrConfig == nil {
		ptrConfig = getDefaultRateLimitConfig()
	}

	ptrRateLimiter := NewRateLimiter(ptrConfig, ptrLogger, ptrMetrics)

	return func(ptrCtx *gin.Context) {
		if !ptrConfig.BoolEnabled {
			ptrCtx.Next()
			return
		}

		// Verificar IP en blacklist
		strClientIP := extractClientIP(ptrCtx)
		if isBlacklisted(strClientIP, ptrConfig.ArrBlacklistIPs) {
			handleRateLimitViolation(ptrCtx, "IP blacklisted", ptrLogger, ptrMetrics)
			return
		}

		// Verificar IP en whitelist
		if isWhitelisted(strClientIP, ptrConfig.ArrWhitelistIPs) {
			ptrCtx.Next()
			return
		}

		// Verificar si IP está bloqueada
		if ptrRateLimiter.IsBlocked(strClientIP) {
			handleRateLimitViolation(ptrCtx, "IP temporarily blocked", ptrLogger, ptrMetrics)
			return
		}

		// Aplicar rate limiting
		ptrRateLimitInfo := ptrRateLimiter.CheckLimit(ptrCtx)

		// Establecer headers de rate limiting
		setRateLimitHeaders(ptrCtx, ptrRateLimitInfo)

		if !ptrRateLimitInfo.BoolAllowed {
			ptrRateLimiter.RecordViolation(ptrCtx, ptrRateLimitInfo)
			handleRateLimitViolation(ptrCtx, ptrRateLimitInfo.StrReason, ptrLogger, ptrMetrics)
			return
		}

		ptrCtx.Next()
	}
}

// IPWhitelistMiddleware middleware para whitelist de IPs
func IPWhitelistMiddleware(arrWhitelistIPs []string, ptrLogger logger.interfaceLogger) gin.HandlerFunc {
	return func(ptrCtx *gin.Context) {
		strClientIP := extractClientIP(ptrCtx)

		if !isWhitelisted(strClientIP, arrWhitelistIPs) {
			ptrLogger.Warn("IP no autorizada",
				zap.String("ip", strClientIP),
				zap.String("path", ptrCtx.Request.URL.Path),
			)

			ptrCtx.JSON(http.StatusForbidden, gin.H{
				"error":   "Forbidden",
				"message": "IP no autorizada",
				"ip":      strClientIP,
			})
			ptrCtx.Abort()
			return
		}

		ptrCtx.Next()
	}
}

// UserBasedRateLimitMiddleware middleware de rate limiting basado en usuario
func UserBasedRateLimitMiddleware(ptrConfig *structRateLimitConfig, ptrLogger logger.interfaceLogger) gin.HandlerFunc {
	return func(ptrCtx *gin.Context) {
		strUserID := ptrCtx.GetString("user_id")
		strUserType := ptrCtx.GetString("user_type")

		if strUserID == "" {
			ptrCtx.Next()
			return
		}

		// Obtener límites para el tipo de usuario
		ptrUserLimit, boolExists := ptrConfig.MapUserTypeLimits[strUserType]
		if !boolExists {
			ptrCtx.Next()
			return
		}

		// Crear clave única para el usuario
		strKey := fmt.Sprintf("user:%s:%s", strUserType, strUserID)

		// Verificar límite (implementación simplificada)
		// En producción, usar Redis o sistema distribuido
		boolAllowed := checkUserRateLimit(strKey, ptrUserLimit)

		if !boolAllowed {
			ptrLogger.Warn("Rate limit excedido para usuario",
				zap.String("user_id", strUserID),
				zap.String("user_type", strUserType),
				zap.String("path", ptrCtx.Request.URL.Path),
			)

			ptrCtx.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"message":     "Demasiadas requests para este usuario",
				"user_id":     strUserID,
				"retry_after": ptrUserLimit.DurationWindow.Seconds(),
			})
			ptrCtx.Abort()
			return
		}

		ptrCtx.Next()
	}
}

// NewRateLimiter crea un nuevo rate limiter
func NewRateLimiter(ptrConfig *structRateLimitConfig, ptrLogger logger.interfaceLogger, ptrMetrics metrics.interfaceManager) *structRateLimiter {
	ptrRateLimiter := &structRateLimiter{
		PtrConfig:     ptrConfig,
		PtrLogger:     ptrLogger,
		PtrMetrics:    ptrMetrics,
		MapLimiters:   make(map[string]interfaceAlgorithmLimiter),
		MapBlockedIPs: concurrency.NewSafeMap(),
		MapViolations: concurrency.NewSafeMap(),
		ChanCleanup:   make(chan struct{}),
		BoolRunning:   true,
	}

	// Iniciar limpieza periódica
	go ptrRateLimiter.cleanup()

	return ptrRateLimiter
}

// CheckLimit verifica el límite de rate para una request
func (ptrRateLimiter *structRateLimiter) CheckLimit(ptrCtx *gin.Context) *structRateLimitInfo {
	strClientIP := extractClientIP(ptrCtx)
	strMethod := ptrCtx.Request.Method
	strPath := ptrCtx.Request.URL.Path
	strUserID := ptrCtx.GetString("user_id")

	// Determinar límites aplicables
	intLimit, durationWindow, intBurstLimit := ptrRateLimiter.determineLimits(strMethod, strPath, strUserID)

	// Crear clave única
	strKey := ptrRateLimiter.createKey(strClientIP, strMethod, strPath, strUserID)

	// Obtener o crear limiter para esta clave
	ptrLimiter := ptrRateLimiter.getLimiter(strKey, intLimit, durationWindow)

	// Verificar límite
	boolAllowed := ptrLimiter.Allow(strKey, 1)
	intRemaining := ptrLimiter.GetRemaining(strKey)

	ptrRateLimitInfo := &structRateLimitInfo{
		StrKey:         strKey,
		IntLimit:       intLimit,
		IntRemaining:   intRemaining,
		TimeResetTime:  time.Now().Add(durationWindow),
		DurationWindow: durationWindow,
		BoolAllowed:    boolAllowed,
	}

	if !boolAllowed {
		ptrRateLimitInfo.StrReason = "Rate limit exceeded"
	}

	// Registrar métricas
	if ptrRateLimiter.PtrMetrics != nil {
		ptrRateLimiter.PtrMetrics.RecordRateLimit(strKey, boolAllowed, intRemaining)
	}

	return ptrRateLimitInfo
}

// IsBlocked verifica si una IP está bloqueada
func (ptrRateLimiter *structRateLimiter) IsBlocked(strIP string) bool {
	objBlockInfo, boolExists := ptrRateLimiter.MapBlockedIPs.Get(strIP)
	if !boolExists {
		return false
	}

	if mapBlockInfo, boolOk := objBlockInfo.(map[string]interface{}); boolOk {
		if objBlockedUntil, boolExists := mapBlockInfo["blocked_until"]; boolExists {
			if timeBlockedUntil, boolOk := objBlockedUntil.(time.Time); boolOk {
				return time.Now().Before(timeBlockedUntil)
			}
		}
	}

	return false
}

// RecordViolation registra una violación de rate limiting
func (ptrRateLimiter *structRateLimiter) RecordViolation(ptrCtx *gin.Context, ptrRateLimitInfo *structRateLimitInfo) {
	strClientIP := extractClientIP(ptrCtx)
	strUserID := ptrCtx.GetString("user_id")
	strEndpoint := fmt.Sprintf("%s %s", ptrCtx.Request.Method, ptrCtx.Request.URL.Path)

	// Obtener información de violación existente
	objViolationInfo, boolExists := ptrRateLimiter.MapViolations.Get(strClientIP)
	var ptrViolationInfo *structViolationInfo

	if boolExists {
		if mapInfo, boolOk := objViolationInfo.(map[string]interface{}); boolOk {
			ptrViolationInfo = mapToViolationInfo(mapInfo)
		}
	}

	if ptrViolationInfo == nil {
		ptrViolationInfo = &structViolationInfo{
			StrIP:              strClientIP,
			StrUserID:          strUserID,
			StrEndpoint:        strEndpoint,
			IntCount:           0,
			TimeFirstViolation: time.Now(),
			BoolBlocked:        false,
		}
	}

	ptrViolationInfo.IntCount++
	ptrViolationInfo.TimeLastViolation = time.Now()

	// Verificar si debe bloquearse
	if ptrRateLimiter.PtrConfig.BoolBlockOnViolation && ptrViolationInfo.IntCount >= 5 {
		ptrRateLimiter.blockIP(strClientIP, ptrRateLimiter.PtrConfig.DurationBlockDuration)
		ptrViolationInfo.BoolBlocked = true
	}

	// Guardar información de violación
	ptrRateLimiter.MapViolations.Set(strClientIP, violationInfoToMap(ptrViolationInfo))

	// Log de violación
	if ptrRateLimiter.PtrConfig.BoolLogViolations {
		ptrRateLimiter.PtrLogger.Warn("Rate limit violation",
			zap.String("ip", strClientIP),
			zap.String("user_id", strUserID),
			zap.String("endpoint", strEndpoint),
			zap.Int("violation_count", ptrViolationInfo.IntCount),
			zap.Bool("blocked", ptrViolationInfo.BoolBlocked),
		)
	}
}

// determineLimits determina los límites aplicables
func (ptrRateLimiter *structRateLimiter) determineLimits(strMethod, strPath, strUserID string) (int, time.Duration, int) {
	// Verificar límites específicos por endpoint
	strEndpointKey := fmt.Sprintf("%s %s", strMethod, strPath)
	if ptrEndpointLimit, boolExists := ptrRateLimiter.PtrConfig.MapEndpointLimits[strEndpointKey]; boolExists {
		return ptrEndpointLimit.IntLimit, ptrEndpointLimit.DurationWindow, ptrEndpointLimit.IntBurstLimit
	}

	// Usar límites por defecto
	return ptrRateLimiter.PtrConfig.IntDefaultLimit,
		ptrRateLimiter.PtrConfig.DurationDefaultWindow,
		ptrRateLimiter.PtrConfig.IntBurstLimit
}

// createKey crea una clave única para rate limiting
func (ptrRateLimiter *structRateLimiter) createKey(strIP, strMethod, strPath, strUserID string) string {
	if strUserID != "" {
		return fmt.Sprintf("user:%s:%s:%s", strUserID, strMethod, strPath)
	}
	return fmt.Sprintf("ip:%s:%s:%s", strIP, strMethod, strPath)
}

// getLimiter obtiene o crea un limiter para una clave
func (ptrRateLimiter *structRateLimiter) getLimiter(strKey string, intLimit int, durationWindow time.Duration) interfaceAlgorithmLimiter {
	ptrRateLimiter.MutexLimiters.RLock()
	if ptrLimiter, boolExists := ptrRateLimiter.MapLimiters[strKey]; boolExists {
		ptrRateLimiter.MutexLimiters.RUnlock()
		return ptrLimiter
	}
	ptrRateLimiter.MutexLimiters.RUnlock()

	ptrRateLimiter.MutexLimiters.Lock()
	defer ptrRateLimiter.MutexLimiters.Unlock()

	// Double-check locking
	if ptrLimiter, boolExists := ptrRateLimiter.MapLimiters[strKey]; boolExists {
		return ptrLimiter
	}

	// Crear nuevo limiter según algoritmo configurado
	var ptrLimiter interfaceAlgorithmLimiter
	switch ptrRateLimiter.PtrConfig.EnumAlgorithm {
	case EnumRateLimitAlgorithmTokenBucket:
		ptrLimiter = NewTokenBucket(intLimit, durationWindow)
	case EnumRateLimitAlgorithmSlidingWindow:
		ptrLimiter = NewSlidingWindow(intLimit, durationWindow)
	case EnumRateLimitAlgorithmFixedWindow:
		ptrLimiter = NewFixedWindow(intLimit, durationWindow)
	case EnumRateLimitAlgorithmLeakyBucket:
		ptrLimiter = NewLeakyBucket(intLimit, durationWindow)
	default:
		ptrLimiter = NewTokenBucket(intLimit, durationWindow)
	}

	ptrRateLimiter.MapLimiters[strKey] = ptrLimiter
	return ptrLimiter
}

// blockIP bloquea una IP temporalmente
func (ptrRateLimiter *structRateLimiter) blockIP(strIP string, durationBlock time.Duration) {
	mapBlockInfo := map[string]interface{}{
		"blocked_at":    time.Now(),
		"blocked_until": time.Now().Add(durationBlock),
		"reason":        "Rate limit violations",
	}

	ptrRateLimiter.MapBlockedIPs.Set(strIP, mapBlockInfo)

	ptrRateLimiter.PtrLogger.Warn("IP bloqueada temporalmente",
		zap.String("ip", strIP),
		zap.Duration("duration", durationBlock),
	)
}

// cleanup limpia datos expirados
func (ptrRateLimiter *structRateLimiter) cleanup() {
	ptrTicker := time.NewTicker(5 * time.Minute)
	defer ptrTicker.Stop()

	for {
		select {
		case <-ptrTicker.C:
			ptrRateLimiter.cleanupExpiredData()
		case <-ptrRateLimiter.ChanCleanup:
			return
		}
	}
}

// cleanupExpiredData limpia datos expirados
func (ptrRateLimiter *structRateLimiter) cleanupExpiredData() {
	timeNow := time.Now()

	// Limpiar IPs bloqueadas expiradas
	arrBlockedKeys := ptrRateLimiter.MapBlockedIPs.Keys()
	for _, strKey := range arrBlockedKeys {
		if objBlockInfo, boolExists := ptrRateLimiter.MapBlockedIPs.Get(strKey); boolExists {
			if mapBlockInfo, boolOk := objBlockInfo.(map[string]interface{}); boolOk {
				if objBlockedUntil, boolExists := mapBlockInfo["blocked_until"]; boolExists {
					if timeBlockedUntil, boolOk := objBlockedUntil.(time.Time); boolOk {
						if timeNow.After(timeBlockedUntil) {
							ptrRateLimiter.MapBlockedIPs.Delete(strKey)
						}
					}
				}
			}
		}
	}

	// Limpiar violaciones antiguas
	arrViolationKeys := ptrRateLimiter.MapViolations.Keys()
	for _, strKey := range arrViolationKeys {
		if objViolationInfo, boolExists := ptrRateLimiter.MapViolations.Get(strKey); boolExists {
			if mapInfo, boolOk := objViolationInfo.(map[string]interface{}); boolOk {
				if objLastViolation, boolExists := mapInfo["last_violation"]; boolExists {
					if timeLastViolation, boolOk := objLastViolation.(time.Time); boolOk {
						if timeNow.Sub(timeLastViolation) > 24*time.Hour {
							ptrRateLimiter.MapViolations.Delete(strKey)
						}
					}
				}
			}
		}
	}
}

// Implementaciones de algoritmos

// NewTokenBucket crea un nuevo token bucket
func NewTokenBucket(intCapacity int, durationRefillRate time.Duration) interfaceAlgorithmLimiter {
	return &structTokenBucket{
		IntCapacity:        intCapacity,
		IntTokens:          intCapacity,
		DurationRefillRate: durationRefillRate,
		TimeLastRefill:     time.Now(),
	}
}

// Allow verifica si se permite la operación
func (ptrBucket *structTokenBucket) Allow(strKey string, intTokens int) bool {
	ptrBucket.MutexBucket.Lock()
	defer ptrBucket.MutexBucket.Unlock()

	ptrBucket.refill()

	if ptrBucket.IntTokens >= intTokens {
		ptrBucket.IntTokens -= intTokens
		return true
	}

	return false
}

// GetRemaining obtiene tokens restantes
func (ptrBucket *structTokenBucket) GetRemaining(strKey string) int {
	ptrBucket.MutexBucket.Lock()
	defer ptrBucket.MutexBucket.Unlock()

	ptrBucket.refill()
	return ptrBucket.IntTokens
}

// Reset resetea el bucket
func (ptrBucket *structTokenBucket) Reset(strKey string) {
	ptrBucket.MutexBucket.Lock()
	defer ptrBucket.MutexBucket.Unlock()

	ptrBucket.IntTokens = ptrBucket.IntCapacity
	ptrBucket.TimeLastRefill = time.Now()
}

// GetStats obtiene estadísticas
func (ptrBucket *structTokenBucket) GetStats() map[string]interface{} {
	ptrBucket.MutexBucket.Lock()
	defer ptrBucket.MutexBucket.Unlock()

	return map[string]interface{}{
		"capacity":    ptrBucket.IntCapacity,
		"tokens":      ptrBucket.IntTokens,
		"refill_rate": ptrBucket.DurationRefillRate.String(),
		"last_refill": ptrBucket.TimeLastRefill,
	}
}

// refill rellena tokens
func (ptrBucket *structTokenBucket) refill() {
	timeNow := time.Now()
	timeSinceLastRefill := timeNow.Sub(ptrBucket.TimeLastRefill)

	if timeSinceLastRefill >= ptrBucket.DurationRefillRate {
		intTokensToAdd := int(timeSinceLastRefill / ptrBucket.DurationRefillRate)
		ptrBucket.IntTokens += intTokensToAdd

		if ptrBucket.IntTokens > ptrBucket.IntCapacity {
			ptrBucket.IntTokens = ptrBucket.IntCapacity
		}

		ptrBucket.TimeLastRefill = timeNow
	}
}

// NewSlidingWindow crea una nueva sliding window
func NewSlidingWindow(intLimit int, durationWindow time.Duration) interfaceAlgorithmLimiter {
	return &structSlidingWindow{
		IntLimit:       intLimit,
		DurationWindow: durationWindow,
		ArrRequests:    make([]time.Time, 0),
	}
}

// Allow verifica si se permite la operación
func (ptrWindow *structSlidingWindow) Allow(strKey string, intTokens int) bool {
	ptrWindow.MutexWindow.Lock()
	defer ptrWindow.MutexWindow.Unlock()

	timeNow := time.Now()
	timeWindowStart := timeNow.Add(-ptrWindow.DurationWindow)

	// Filtrar requests dentro de la ventana
	arrValidRequests := make([]time.Time, 0)
	for _, timeRequest := range ptrWindow.ArrRequests {
		if timeRequest.After(timeWindowStart) {
			arrValidRequests = append(arrValidRequests, timeRequest)
		}
	}

	ptrWindow.ArrRequests = arrValidRequests

	if len(ptrWindow.ArrRequests) < ptrWindow.IntLimit {
		ptrWindow.ArrRequests = append(ptrWindow.ArrRequests, timeNow)
		return true
	}

	return false
}

// GetRemaining obtiene requests restantes
func (ptrWindow *structSlidingWindow) GetRemaining(strKey string) int {
	ptrWindow.MutexWindow.Lock()
	defer ptrWindow.MutexWindow.Unlock()

	timeNow := time.Now()
	timeWindowStart := timeNow.Add(-ptrWindow.DurationWindow)

	intValidRequests := 0
	for _, timeRequest := range ptrWindow.ArrRequests {
		if timeRequest.After(timeWindowStart) {
			intValidRequests++
		}
	}

	intRemaining := ptrWindow.IntLimit - intValidRequests
	if intRemaining < 0 {
		intRemaining = 0
	}

	return intRemaining
}

// Reset resetea la ventana
func (ptrWindow *structSlidingWindow) Reset(strKey string) {
	ptrWindow.MutexWindow.Lock()
	defer ptrWindow.MutexWindow.Unlock()

	ptrWindow.ArrRequests = make([]time.Time, 0)
}

// GetStats obtiene estadísticas
func (ptrWindow *structSlidingWindow) GetStats() map[string]interface{} {
	ptrWindow.MutexWindow.Lock()
	defer ptrWindow.MutexWindow.Unlock()

	return map[string]interface{}{
		"limit":         ptrWindow.IntLimit,
		"window":        ptrWindow.DurationWindow.String(),
		"current_count": len(ptrWindow.ArrRequests),
		"remaining":     ptrWindow.IntLimit - len(ptrWindow.ArrRequests),
	}
}

// Funciones de utilidad

// getDefaultRateLimitConfig obtiene configuración por defecto
func getDefaultRateLimitConfig() *structRateLimitConfig {
	return &structRateLimitConfig{
		BoolEnabled:           true,
		IntDefaultLimit:       100,
		DurationDefaultWindow: time.Minute,
		MapEndpointLimits:     make(map[string]structEndpointLimit),
		MapUserTypeLimits:     make(map[string]structUserTypeLimit),
		ArrWhitelistIPs:       []string{},
		ArrBlacklistIPs:       []string{},
		BoolLogViolations:     true,
		BoolBlockOnViolation:  false,
		DurationBlockDuration: 15 * time.Minute,
		IntBurstLimit:         10,
		EnumAlgorithm:         EnumRateLimitAlgorithmTokenBucket,
	}
}

// isWhitelisted verifica si una IP está en whitelist
func isWhitelisted(strIP string, arrWhitelistIPs []string) bool {
	for _, strWhitelistIP := range arrWhitelistIPs {
		if strIP == strWhitelistIP {
			return true
		}
	}
	return false
}

// isBlacklisted verifica si una IP está en blacklist
func isBlacklisted(strIP string, arrBlacklistIPs []string) bool {
	for _, strBlacklistIP := range arrBlacklistIPs {
		if strIP == strBlacklistIP {
			return true
		}
	}
	return false
}

// setRateLimitHeaders establece headers de rate limiting
func setRateLimitHeaders(ptrCtx *gin.Context, ptrRateLimitInfo *structRateLimitInfo) {
	ptrCtx.Header("X-RateLimit-Limit", strconv.Itoa(ptrRateLimitInfo.IntLimit))
	ptrCtx.Header("X-RateLimit-Remaining", strconv.Itoa(ptrRateLimitInfo.IntRemaining))
	ptrCtx.Header("X-RateLimit-Reset", strconv.FormatInt(ptrRateLimitInfo.TimeResetTime.Unix(), 10))
	ptrCtx.Header("X-RateLimit-Window", ptrRateLimitInfo.DurationWindow.String())
}

// handleRateLimitViolation maneja violaciones de rate limiting
func handleRateLimitViolation(ptrCtx *gin.Context, strReason string, ptrLogger logger.interfaceLogger, ptrMetrics metrics.interfaceManager) {
	strClientIP := extractClientIP(ptrCtx)
	strRequestID := ptrCtx.GetString("request_id")

	ptrLogger.Warn("Rate limit violation",
		zap.String("ip", strClientIP),
		zap.String("reason", strReason),
		zap.String("path", ptrCtx.Request.URL.Path),
		zap.String("request_id", strRequestID),
	)

	if ptrMetrics != nil {
		ptrMetrics.RecordRateLimit("violation", false, 0)
	}

	ptrCtx.JSON(http.StatusTooManyRequests, gin.H{
		"error":       "Rate limit exceeded",
		"message":     "Demasiadas requests. Intente más tarde.",
		"reason":      strReason,
		"request_id":  strRequestID,
		"retry_after": 60,
	})

	ptrCtx.Abort()
}

// checkUserRateLimit verifica rate limit para usuario (implementación simplificada)
func checkUserRateLimit(strKey string, ptrUserLimit structUserTypeLimit) bool {
	// En producción, implementar con Redis o sistema distribuido
	// Esta es una implementación simplificada para demostración
	return true
}

// Funciones helper para conversión de datos

// mapToViolationInfo convierte map a ViolationInfo
func mapToViolationInfo(mapInfo map[string]interface{}) *structViolationInfo {
	ptrViolationInfo := &structViolationInfo{}

	if objIP, boolExists := mapInfo["ip"]; boolExists {
		if strIP, boolOk := objIP.(string); boolOk {
			ptrViolationInfo.StrIP = strIP
		}
	}

	if objUserID, boolExists := mapInfo["user_id"]; boolExists {
		if strUserID, boolOk := objUserID.(string); boolOk {
			ptrViolationInfo.StrUserID = strUserID
		}
	}

	if objEndpoint, boolExists := mapInfo["endpoint"]; boolExists {
		if strEndpoint, boolOk := objEndpoint.(string); boolOk {
			ptrViolationInfo.StrEndpoint = strEndpoint
		}
	}

	if objCount, boolExists := mapInfo["count"]; boolExists {
		if intCount, boolOk := objCount.(int); boolOk {
			ptrViolationInfo.IntCount = intCount
		}
	}

	if objFirstViolation, boolExists := mapInfo["first_violation"]; boolExists {
		if timeFirstViolation, boolOk := objFirstViolation.(time.Time); boolOk {
			ptrViolationInfo.TimeFirstViolation = timeFirstViolation
		}
	}

	if objLastViolation, boolExists := mapInfo["last_violation"]; boolExists {
		if timeLastViolation, boolOk := objLastViolation.(time.Time); boolOk {
			ptrViolationInfo.TimeLastViolation = timeLastViolation
		}
	}

	if objBlocked, boolExists := mapInfo["blocked"]; boolExists {
		if boolBlocked, boolOk := objBlocked.(bool); boolOk {
			ptrViolationInfo.BoolBlocked = boolBlocked
		}
	}

	return ptrViolationInfo
}

// violationInfoToMap convierte ViolationInfo a map
func violationInfoToMap(ptrViolationInfo *structViolationInfo) map[string]interface{} {
	return map[string]interface{}{
		"ip":              ptrViolationInfo.StrIP,
		"user_id":         ptrViolationInfo.StrUserID,
		"endpoint":        ptrViolationInfo.StrEndpoint,
		"count":           ptrViolationInfo.IntCount,
		"first_violation": ptrViolationInfo.TimeFirstViolation,
		"last_violation":  ptrViolationInfo.TimeLastViolation,
		"blocked":         ptrViolationInfo.BoolBlocked,
	}
}

// NewFixedWindow y NewLeakyBucket (implementaciones básicas)
func NewFixedWindow(intLimit int, durationWindow time.Duration) interfaceAlgorithmLimiter {
	return &structFixedWindow{
		IntLimit:        intLimit,
		DurationWindow:  durationWindow,
		TimeWindowStart: time.Now(),
	}
}

func (ptrWindow *structFixedWindow) Allow(strKey string, intTokens int) bool {
	ptrWindow.MutexWindow.Lock()
	defer ptrWindow.MutexWindow.Unlock()

	timeNow := time.Now()
	if timeNow.Sub(ptrWindow.TimeWindowStart) >= ptrWindow.DurationWindow {
		ptrWindow.IntCurrentCount = 0
		ptrWindow.TimeWindowStart = timeNow
	}

	if ptrWindow.IntCurrentCount < ptrWindow.IntLimit {
		ptrWindow.IntCurrentCount += intTokens
		return true
	}

	return false
}

func (ptrWindow *structFixedWindow) GetRemaining(strKey string) int {
	ptrWindow.MutexWindow.Lock()
	defer ptrWindow.MutexWindow.Unlock()

	return ptrWindow.IntLimit - ptrWindow.IntCurrentCount
}

func (ptrWindow *structFixedWindow) Reset(strKey string) {
	ptrWindow.MutexWindow.Lock()
	defer ptrWindow.MutexWindow.Unlock()

	ptrWindow.IntCurrentCount = 0
	ptrWindow.TimeWindowStart = time.Now()
}

func (ptrWindow *structFixedWindow) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"limit":         ptrWindow.IntLimit,
		"current_count": ptrWindow.IntCurrentCount,
		"window_start":  ptrWindow.TimeWindowStart,
	}
}

func NewLeakyBucket(intCapacity int, durationLeakRate time.Duration) interfaceAlgorithmLimiter {
	return &structLeakyBucket{
		IntCapacity:      intCapacity,
		DurationLeakRate: durationLeakRate,
		TimeLastLeak:     time.Now(),
	}
}

func (ptrBucket *structLeakyBucket) Allow(strKey string, intTokens int) bool {
	ptrBucket.MutexBucket.Lock()
	defer ptrBucket.MutexBucket.Unlock()

	ptrBucket.leak()

	if ptrBucket.IntCurrentLevel+intTokens <= ptrBucket.IntCapacity {
		ptrBucket.IntCurrentLevel += intTokens
		return true
	}

	return false
}

func (ptrBucket *structLeakyBucket) GetRemaining(strKey string) int {
	ptrBucket.MutexBucket.Lock()
	defer ptrBucket.MutexBucket.Unlock()

	ptrBucket.leak()
	return ptrBucket.IntCapacity - ptrBucket.IntCurrentLevel
}

func (ptrBucket *structLeakyBucket) Reset(strKey string) {
	ptrBucket.MutexBucket.Lock()
	defer ptrBucket.MutexBucket.Unlock()

	ptrBucket.IntCurrentLevel = 0
	ptrBucket.TimeLastLeak = time.Now()
}

func (ptrBucket *structLeakyBucket) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"capacity":      ptrBucket.IntCapacity,
		"current_level": ptrBucket.IntCurrentLevel,
		"leak_rate":     ptrBucket.DurationLeakRate.String(),
	}
}

func (ptrBucket *structLeakyBucket) leak() {
	timeNow := time.Now()
	timeSinceLastLeak := timeNow.Sub(ptrBucket.TimeLastLeak)

	if timeSinceLastLeak >= ptrBucket.DurationLeakRate {
		intLeakAmount := int(timeSinceLastLeak / ptrBucket.DurationLeakRate)
		ptrBucket.IntCurrentLevel -= intLeakAmount

		if ptrBucket.IntCurrentLevel < 0 {
			ptrBucket.IntCurrentLevel = 0
		}

		ptrBucket.TimeLastLeak = timeNow
	}
}

// extractClientIP extrae la IP del cliente usando utilidades de red
func extractClientIP(ptrCtx *gin.Context) string {
	return utils.ExtractClientIP(ptrCtx)
}
