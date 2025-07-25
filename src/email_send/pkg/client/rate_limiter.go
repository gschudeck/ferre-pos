package client

import (
    "context"
    "sync"
    "time"
    
    "golang.org/x/time/rate"
)

// RateLimiter controla la tasa de envío de mensajes
type RateLimiter struct {
    limiter *rate.Limiter
    enabled bool
    mu      sync.RWMutex
}

// NewRateLimiter crea un nuevo limitador de tasa
func NewRateLimiter(requestsPerMinute int, burst int, enabled bool) *RateLimiter {
    // Convertir requests por minuto a requests por segundo
    rps := rate.Limit(float64(requestsPerMinute) / 60.0)
    
    return &RateLimiter{
        limiter: rate.NewLimiter(rps, burst),
        enabled: enabled,
    }
}

// Wait espera hasta que se permita la siguiente operación
func (rl *RateLimiter) Wait(ctx context.Context) error {
    rl.mu.RLock()
    enabled := rl.enabled
    rl.mu.RUnlock()
    
    if !enabled {
        return nil
    }
    
    return rl.limiter.Wait(ctx)
}

// Allow verifica si una operación está permitida sin esperar
func (rl *RateLimiter) Allow() bool {
    rl.mu.RLock()
    enabled := rl.enabled
    rl.mu.RUnlock()
    
    if !enabled {
        return true
    }
    
    return rl.limiter.Allow()
}

// SetEnabled habilita o deshabilita el rate limiting
func (rl *RateLimiter) SetEnabled(enabled bool) {
    rl.mu.Lock()
    rl.enabled = enabled
    rl.mu.Unlock()
}

// UpdateLimits actualiza los límites del rate limiter
func (rl *RateLimiter) UpdateLimits(requestsPerMinute int, burst int) {
    rl.mu.Lock()
    defer rl.mu.Unlock()
    
    rps := rate.Limit(float64(requestsPerMinute) / 60.0)
    rl.limiter.SetLimit(rps)
    rl.limiter.SetBurst(burst)
}
