package client

import (
    "context"
    "crypto/tls"
    "fmt"
    "net/smtp"
    "time"
    
    "github.com/google/uuid"
    "github.com/yourorg/smtp-client/config"
    "github.com/yourorg/smtp-client/pkg/auth"
    "github.com/yourorg/smtp-client/pkg/errors"
    "github.com/yourorg/smtp-client/pkg/logger"
)

// SMTPConfig configuración del cliente SMTP
type SMTPConfig struct {
    Host     string
    Port     int
    UseTLS   bool
    Provider auth.ProviderType
    Timeout  time.Duration
    Username string // Para autenticación básica
    Password string // Para autenticación básica
}

// SMTPClient cliente SMTP principal con todas las funcionalidades
type SMTPClient struct {
    config         *SMTPConfig
    poolConfig     *PoolConfig
    tokenManager   auth.TokenManager
    pool           *ConnectionPool
    validator      *MessageValidator
    builder        *MessageBuilder
    logger         logger.Logger
    retryManager   *RetryManager
    rateLimiter    *RateLimiter
    metrics        *MetricsCollector
}

// NewSMTPClient crea un nuevo cliente SMTP con OAuth2 y todas las funcionalidades
func NewSMTPClient(smtpConfig *SMTPConfig, poolConfig *PoolConfig, tokenManager auth.TokenManager, 
                   retryConfig *config.RetryConfig, rateLimitConfig *config.RateLimitingConfig, log logger.Logger) *SMTPClient {
    
    pool := NewConnectionPool(smtpConfig, poolConfig, log)
    retryManager := NewRetryManager(retryConfig, log)
    rateLimiter := NewRateLimiter(rateLimitConfig.RequestsPerMinute, rateLimitConfig.Burst, rateLimitConfig.Enabled)
    metrics := NewMetricsCollector(string(smtpConfig.Provider))
    
    return &SMTPClient{
        config:         smtpConfig,
        poolConfig:     poolConfig,
        tokenManager:   tokenManager,
        pool:           pool,
        validator:      NewMessageValidator(),
        builder:        NewMessageBuilder(),
        logger:         log,
        retryManager:   retryManager,
        rateLimiter:    rateLimiter,
        metrics:        metrics,
    }
}

// NewGenericSMTPClient crea un cliente SMTP para servidores genéricos (sin OAuth2)
func NewGenericSMTPClient(smtpConfig *SMTPConfig, poolConfig *PoolConfig, 
                         retryConfig *config.RetryConfig, rateLimitConfig *config.RateLimitingConfig, log logger.Logger) *SMTPClient {
    
    pool := NewConnectionPool(smtpConfig, poolConfig, log)
    retryManager := NewRetryManager(retryConfig, log)
    rateLimiter := NewRateLimiter(rateLimitConfig.RequestsPerMinute, rateLimitConfig.Burst, rateLimitConfig.Enabled)
    metrics := NewMetricsCollector("generic")
    
    return &SMTPClient{
        config:         smtpConfig,
        poolConfig:     poolConfig,
        tokenManager:   nil, // Sin OAuth2
        pool:           pool,
        validator:      NewMessageValidator(),
        builder:        NewMessageBuilder(),
        logger:         log,
        retryManager:   retryManager,
        rateLimiter:    rateLimiter,
        metrics:        metrics,
    }
}

// SendMessage envía un mensaje de correo con reintentos y rate limiting
func (c *SMTPClient) SendMessage(ctx context.Context, message *EmailMessage) error {
    requestID := uuid.New().String()
    startTime := time.Now()
    
    c.logger.Info("Iniciando envío de mensaje", map[string]interface{}{
        "request_id": requestID,
        "to":         message.To,
        "subject":    message.Subject,
        "provider":   c.config.Provider,
    })
    
    // Aplicar rate limiting
    if err := c.rateLimiter.Wait(ctx); err != nil {
        c.metrics.RecordMessageFailed("rate_limited")
        return errors.NewTimeoutError(
            "RATE_LIMITED",
            "Mensaje bloqueado por rate limiting",
            requestID,
            err,
        )
    }
    
    // Ejecutar envío con reintentos
    err := c.retryManager.ExecuteWithRetry(ctx, func() error {
        return c.sendMessageInternal(ctx, message, requestID)
    }, requestID)
    
    // Registrar métricas
    duration := time.Since(startTime)
    c.metrics.RecordSendDuration(duration)
    
    if err != nil {
        c.metrics.RecordMessageFailed(c.getErrorType(err))
        c.logger.Error("Error enviando mensaje", err)
        return err
    }
    
    c.metrics.RecordMessageSent()
    c.logger.Info("Mensaje enviado exitosamente", map[string]interface{}{
        "request_id": requestID,
        "to":         message.To,
        "subject":    message.Subject,
        "duration":   duration.String(),
    })
    
    return nil
}

// SendBulkMessages envía múltiples mensajes de forma eficiente
func (c *SMTPClient) SendBulkMessages(ctx context.Context, messages []*EmailMessage, concurrency int) error {
    if concurrency <= 0 {
        concurrency = 5 // Valor por defecto
    }
    
    semaphore := make(chan struct{}, concurrency)
    errChan := make(chan error, len(messages))
    
    c.logger.Info("Iniciando envío masivo", map[string]interface{}{
        "total_messages": len(messages),
        "concurrency":    concurrency,
    })
    
    for i, message := range messages {
        select {
        case semaphore <- struct{}{}: // Acquire
            go func(msg *EmailMessage, index int) {
                defer func() { <-semaphore }() // Release
                
                if err := c.SendMessage(ctx, msg); err != nil {
                    c.logger.Error("Error en envío masivo", err)
                    errChan <- fmt.Errorf("mensaje %d falló: %w", index, err)
                } else {
                    errChan <- nil
                }
            }(message, i)
        case <-ctx.Done():
            return ctx.Err()
        }
    }
    
    // Esperar que todos los mensajes se procesen
    var errors []error
    for i := 0; i < len(messages); i++ {
        select {
        case err := <-errChan:
            if err != nil {
                errors = append(errors, err)
            }
        case <-ctx.Done():
            return ctx.Err()
        }
    }
    
    if len(errors) > 0 {
        return fmt.Errorf("fallos en envío masivo: %d de %d mensajes fallaron", len(errors), len(messages))
    }
    
    c.logger.Info("Envío masivo completado exitosamente", map[string]interface{}{
        "total_messages": len(messages),
    })
    
    return nil
}

func (c *SMTPClient) sendMessageInternal(ctx context.Context, message *EmailMessage, requestID string) error {
    // Validar mensaje
    if err := c.validator.ValidateMessage(message, requestID); err != nil {
        return err
    }
    
    // Construir mensaje RFC822
    messageData, err := c.builder.BuildRFC822Message(message, requestID)
    if err != nil {
        return err
    }
    
    // Obtener conexión del pool
    conn, err := c.pool.GetConnection(ctx, requestID)
    if err != nil {
        return err
    }
    defer c.pool.ReturnConnection(conn, requestID)
    
    // Autenticar
    if err := c.authenticateConnection(ctx, conn, message.From, requestID); err != nil {
        return err
    }
    
    // Enviar mensaje
    return c.sendMessageData(conn, message, messageData, requestID)
}

func (c *SMTPClient) authenticateConnection(ctx context.Context, conn *smtp.Client, from, requestID string) error {
    if c.tokenManager != nil {
        // Autenticación OAuth2
        token, err := c.tokenManager.GetToken(ctx, requestID)
        if err != nil {
            c.metrics.RecordAuthFailure()
            return errors.NewAuthenticationError(
                "TOKEN_RETRIEVAL_FAILED",
                "Error obteniendo token de acceso",
                requestID,
                err,
            )
        }
        
        oauthAuth := &oauth2SMTPAuth{
            username: from,
            token:    token.AccessToken,
        }
        
        if err := conn.Auth(oauthAuth); err != nil {
            c.metrics.RecordAuthFailure()
            return errors.NewAuthenticationError(
                "OAUTH2_AUTH_FAILED",
                "Error en autenticación OAuth2",
                requestID,
                err,
            )
        }
    } else if c.config.Username != "" && c.config.Password != "" {
        // Autenticación básica
        auth := smtp.PlainAuth("", c.config.Username, c.config.Password, c.config.Host)
        if err := conn.Auth(auth); err != nil {
            c.metrics.RecordAuthFailure()
            return errors.NewAuthenticationError(
                "BASIC_AUTH_FAILED",
                "Error en autenticación básica",
                requestID,
                err,
            )
        }
    }
    
    return nil
}

func (c *SMTPClient) sendMessageData(conn *smtp.Client, message *EmailMessage, messageData, requestID string) error {
    // Establecer remitente
    if err := conn.Mail(message.From); err != nil {
        return errors.NewConnectionError(
            "MAIL_FROM_FAILED",
            "Error estableciendo remitente",
            requestID,
            err,
        )
    }
    
    // Establecer destinatarios
    allRecipients := append(message.To, append(message.Cc, message.Bcc...)...)
    for _, recipient := range allRecipients {
        if err := conn.Rcpt(recipient); err != nil {
            return errors.NewConnectionError(
                "RCPT_TO_FAILED",
                fmt.Sprintf("Error estableciendo destinatario %s", recipient),
                requestID,
                err,
            )
        }
    }
    
    // Enviar datos del mensaje
    writer, err := conn.Data()
    if err != nil {
        return errors.NewConnectionError(
            "DATA_START_FAILED",
            "Error iniciando envío de datos",
            requestID,
            err,
        )
    }
    defer writer.Close()
    
    if _, err := writer.Write([]byte(messageData)); err != nil {
        return errors.NewConnectionError(
            "DATA_WRITE_FAILED",
            "Error escribiendo datos del mensaje",
            requestID,
            err,
        )
    }
    
    return nil
}

// HealthCheck verifica el estado del cliente
func (c *SMTPClient) HealthCheck(ctx context.Context) error {
    if err := c.pool.HealthCheck(ctx); err != nil {
        return fmt.Errorf("health check del pool falló: %w", err)
    }
    
    // Verificar autenticación si es OAuth2
    if c.tokenManager != nil {
        requestID := "health-check"
        _, err := c.tokenManager.GetToken(ctx, requestID)
        if err != nil {
            return fmt.Errorf("health check de autenticación falló: %w", err)
        }
    }
    
    return nil
}

// GetStats retorna estadísticas del cliente
func (c *SMTPClient) GetStats() map[string]interface{} {
    return map[string]interface{}{
        "provider":           c.config.Provider,
        "host":               c.config.Host,
        "port":               c.config.Port,
        "tls_enabled":        c.config.UseTLS,
        "rate_limit_enabled": c.rateLimiter.enabled,
        "pool_max_connections": c.poolConfig.MaxConnections,
    }
}

// UpdateRateLimits actualiza los límites de rate limiting
func (c *SMTPClient) UpdateRateLimits(requestsPerMinute int, burst int) {
    c.rateLimiter.UpdateLimits(requestsPerMinute, burst)
    c.logger.Info("Rate limits actualizados", map[string]interface{}{
        "requests_per_minute": requestsPerMinute,
        "burst":              burst,
    })
}

// EnableRateLimit habilita o deshabilita el rate limiting
func (c *SMTPClient) EnableRateLimit(enabled bool) {
    c.rateLimiter.SetEnabled(enabled)
    c.logger.Info("Rate limiting actualizado", map[string]interface{}{
        "enabled": enabled,
    })
}

// Close cierra el cliente y libera recursos
func (c *SMTPClient) Close() error {
    c.logger.Info("Cerrando cliente SMTP", map[string]interface{}{
        "provider": c.config.Provider,
    })
    
    if c.pool != nil {
        return c.pool.Close()
    }
    return nil
}

func (c *SMTPClient) getErrorType(err error) string {
    if smtpErr, ok := err.(*errors.SMTPError); ok {
        return string(smtpErr.Type)
    }
    return "unknown"
}

// oauth2SMTPAuth implementa smtp.Auth para OAuth2
type oauth2SMTPAuth struct {
    username string
    token    string
}

func (a *oauth2SMTPAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
    return "XOAUTH2", []byte(fmt.Sprintf("user=%s\x01auth=Bearer %s\x01\x01", a.username, a.token)), nil
}

func (a *oauth2SMTPAuth) Next(fromServer []byte, more bool) ([]byte, error) {
    if more {
        return nil, fmt.Errorf("error de autenticación OAuth2: %s", string(fromServer))
    }
    return nil, nil
}
