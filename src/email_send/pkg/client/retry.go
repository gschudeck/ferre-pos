package client

import (
    "context"
    "fmt"
    "math"
    "math/rand"
    "time"
    
    "github.com/yourorg/smtp-client/config"
    "github.com/yourorg/smtp-client/pkg/errors"
    "github.com/yourorg/smtp-client/pkg/logger"
)

// RetryManager maneja la lógica de reintentos
type RetryManager struct {
    config *config.RetryConfig
    logger logger.Logger
}

// NewRetryManager crea un nuevo manager de reintentos
func NewRetryManager(cfg *config.RetryConfig, log logger.Logger) *RetryManager {
    return &RetryManager{
        config: cfg,
        logger: log,
    }
}

// ExecuteWithRetry ejecuta una función con reintentos automáticos
func (rm *RetryManager) ExecuteWithRetry(ctx context.Context, operation func() error, requestID string) error {
    var lastErr error
    
    for attempt := 1; attempt <= rm.config.MaxAttempts; attempt++ {
        err := operation()
        if err == nil {
            if attempt > 1 {
                rm.logger.Info("Operación exitosa después de reintentos", map[string]interface{}{
                    "request_id": requestID,
                    "attempt":    attempt,
                })
            }
            return nil
        }
        
        lastErr = err
        
        // No reintentar en ciertos tipos de error
        if !rm.shouldRetry(err) {
            rm.logger.Debug("Error no recuperable, no se reintentará", map[string]interface{}{
                "request_id": requestID,
                "error":      err.Error(),
                "attempt":    attempt,
            })
            return err
        }
        
        // Si es el último intento, no esperar
        if attempt == rm.config.MaxAttempts {
            break
        }
        
        // Calcular tiempo de espera con backoff exponencial
        backoff := rm.calculateBackoff(attempt)
        
        rm.logger.Warn("Operación falló, reintentando", map[string]interface{}{
            "request_id": requestID,
            "attempt":    attempt,
            "max_attempts": rm.config.MaxAttempts,
            "backoff":    backoff.String(),
            "error":      err.Error(),
        })
        
        // Esperar con jitter
        select {
        case <-time.After(backoff):
            continue
        case <-ctx.Done():
            return ctx.Err()
        }
    }
    
    return fmt.Errorf("operación falló después de %d intentos: %w", rm.config.MaxAttempts, lastErr)
}

func (rm *RetryManager) shouldRetry(err error) bool {
    // Verificar si es un error específico de SMTP
    if smtpErr, ok := err.(*errors.SMTPError); ok {
        switch smtpErr.Type {
        case errors.ErrorTypeValidation:
            // Errores de validación no se reintentan
            return false
        case errors.ErrorTypeAuthentication:
            // Errores de autenticación se reintentan (token expirado)
            return true
        case errors.ErrorTypeConnection:
            // Errores de conexión se reintentan
            return true
        case errors.ErrorTypeTimeout:
            // Timeouts se reintentan
            return true
        default:
            return true
        }
    }
    
    // Por defecto, reintentar errores desconocidos
    return true
}

func (rm *RetryManager) calculateBackoff(attempt int) time.Duration {
    // Backoff exponencial con jitter
    backoff := float64(rm.config.InitialBackoff) * math.Pow(rm.config.Multiplier, float64(attempt-1))
    
    // Aplicar límite máximo
    if backoff > float64(rm.config.MaxBackoff) {
        backoff = float64(rm.config.MaxBackoff)
    }
    
    // Agregar jitter (±25%)
    jitter := backoff * 0.25 * (rand.Float64()*2 - 1)
    backoff += jitter
    
    return time.Duration(backoff)
}
