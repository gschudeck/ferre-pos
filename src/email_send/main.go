package main

import (
    "context"
    "flag"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"
    
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "github.com/yourorg/smtp-client/config"
    "github.com/yourorg/smtp-client/pkg/auth"
    "github.com/yourorg/smtp-client/pkg/client"
    "github.com/yourorg/smtp-client/pkg/logger"
    "github.com/yourorg/smtp-client/pkg/providers"
    "golang.org/x/oauth2"
)

var (
    configPath = flag.String("config", "", "Ruta al archivo de configuración")
    envPath    = flag.String("env", "", "Ruta al archivo .env")
)

func main() {
    flag.Parse()
    
    // Cargar configuración
    cfg, credentials, err := loadConfiguration()
    if err != nil {
        log.Fatal("Error cargando configuración:", err)
    }
    
    // Validar configuración
    if err := validateConfiguration(cfg, credentials); err != nil {
        log.Fatal("Error validando configuración:", err)
    }
    
    // Crear logger
    logger := logger.NewLogrusLogger(&cfg.Logging)
    logger.Info("Iniciando cliente SMTP", map[string]interface{}{
        "version":     cfg.App.Version,
        "environment": cfg.App.Environment,
    })
    
    // Crear factory de proveedores
    factory := providers.NewProviderFactory(cfg, credentials, logger)
    
    // Crear cliente para el proveedor deseado
    smtpClient, err := createSMTPClient(factory, cfg, logger)
    if err != nil {
        log.Fatal("Error creando cliente SMTP:", err)
    }
    defer smtpClient.Close()
    
    // Configurar servidor de métricas si está habilitado
    if cfg.Metrics.Enabled {
        go startMetricsServer(cfg.Metrics.Port, cfg.Metrics.Path, logger)
    }
    
    // Configurar health check
    http.HandleFunc("/health", healthCheckHandler(smtpClient))
    go func() {
        log.Println("Health check disponible en :8080/health")
        if err := http.ListenAndServe(":8080", nil); err != nil {
            logger.Error("Error en servidor health check", err)
        }
    }()
    
    // Ejemplo de uso
    if err := runExamples(smtpClient, logger); err != nil {
        logger.Error("Error ejecutando ejemplos", err)
    }
    
    // Esperar señal de terminación
    waitForShutdown(logger)
}

func loadConfiguration() (*config.Config, *config.EnvironmentCredentials, error) {
    if *configPath != "" {
        return config.LoadFromPath(*configPath)
    } else if *envPath != "" {
        return config.LoadWithEnv(*envPath)
    } else {
        return config.LoadDefault()
    }
}

func validateConfiguration(cfg *config.Config, credentials *config.EnvironmentCredentials) error {
    validator := config.NewConfigValidator()
    
    if err := validator.ValidateConfig(cfg); err != nil {
        return err
    }
    
    return validator.ValidateCredentials(credentials, cfg)
}

func createSMTPClient(factory *providers.ProviderFactory, cfg *config.Config, logger logger.Logger) (*client.SMTPClient, error) {
    // Determinar qué proveedor usar basado en configuración
    availableProviders := factory.GetAvailableProviders()
    
    if len(availableProviders) == 0 {
        return nil, fmt.Errorf("no hay proveedores disponibles")
    }
    
    // Por simplicidad, usar el primer proveedor disponible
    provider := availableProviders[0]
    
    switch provider {
    case "gmail":
        return factory.CreateGmailClient()
    case "office365":
        return factory.CreateOffice365Client()
    default:
        // Para servidor genérico, necesitaríamos configuración adicional
        return factory.CreateGenericClient("smtp.example.com", 587, "user", "pass")
    }
}

func startMetricsServer(port int, path string, logger logger.Logger) {
    http.Handle(path, promhttp.Handler())
    
    logger.Info("Servidor de métricas iniciado", map[string]interface{}{
        "port": port,
        "path": path,
    })
    
    if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
        logger.Error("Error en servidor de métricas", err)
    }
}

func healthCheckHandler(smtpClient *client.SMTPClient) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
        defer cancel()
        
        if err := smtpClient.HealthCheck(ctx); err != nil {
            w.WriteHeader(http.StatusServiceUnavailable)
            w.Write([]byte(fmt.Sprintf("Health check failed: %v", err)))
            return
        }
        
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("OK"))
    }
}

func runExamples(smtpClient *client.SMTPClient, logger logger.Logger) error {
    ctx := context.Background()
    
    // Ejemplo 1: Mensaje simple
    simpleMessage := &client.EmailMessage{
        From:    "remitente@example.com",
        To:      []string{"destinatario@example.com"},
        Subject: "Mensaje de prueba",
        Body:    "Este es un mensaje de prueba desde el cliente SMTP en Go.",
    }
    
    if err := smtpClient.SendMessage(ctx, simpleMessage); err != nil {
        return fmt.Errorf("error enviando mensaje simple: %w", err)
    }
    
    logger.Info("Mensaje simple enviado exitosamente", map[string]interface{}{})
    
    // Ejemplo 2: Mensaje con HTML y adjunto
    htmlMessage := &client.EmailMessage{
        From:    "remitente@example.com",
        To:      []string{"destinatario@example.com"},
        Cc:      []string{"copia@example.com"},
        Subject: "Reporte mensual",
        HTMLBody: `
            <h1>Reporte Mensual</h1>
            <p>Estimado cliente,</p>
            <p>Adjunto encontrará el reporte mensual.</p>
            <p>Saludos cordiales,<br>Equipo de Ventas</p>
        `,
        Attachments: []client.FileAttachment{
            {
                Filename:    "reporte.txt",
                ContentType: "text/plain",
                Data:        []byte("Contenido del reporte..."),
            },
        },
    }
    
    if err := smtpClient.SendMessage(ctx, htmlMessage); err != nil {
        return fmt.Errorf("error enviando mensaje HTML: %w", err)
    }
    
    logger.Info("Mensaje HTML enviado exitosamente", map[string]interface{}{})
    
    // Ejemplo 3: Envío masivo
    var bulkMessages []*client.EmailMessage
    for i := 1; i <= 5; i++ {
        message := &client.EmailMessage{
            From:    "remitente@example.com",
            To:      []string{fmt.Sprintf("destinatario%d@example.com", i)},
            Subject: fmt.Sprintf("Mensaje masivo #%d", i),
            Body:    fmt.Sprintf("Este es el mensaje número %d del envío masivo.", i),
        }
        bulkMessages = append(bulkMessages, message)
    }
    
    if err := smtpClient.SendBulkMessages(ctx, bulkMessages, 3); err != nil {
        return fmt.Errorf("error en envío masivo: %w", err)
    }
    
    logger.Info("Envío masivo completado exitosamente", map[string]interface{}{
        "total_messages": len(bulkMessages),
    })
    
    return nil
}

func waitForShutdown(logger logger.Logger) {
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    
    sig := <-sigChan
    logger.Info("Señal de terminación recibida", map[string]interface{}{
        "signal": sig.String(),
    })
}
