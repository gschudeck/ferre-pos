Cliente SMTP en Go con OAuth2

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/Build-Passing-brightgreen.svg)]()

Cliente SMTP empresarial robusto y seguro con soporte completo para OAuth2, configuración externa y gestión avanzada de conexiones. Diseñado para sistemas de producción que requieren envío masivo y confiable de correos electrónicos.

## 🚀 Características

### ✅ Autenticación Avanzada

- **OAuth2 completo** para Gmail y Office 365
- **Autenticación básica** para servidores SMTP genéricos
- **Gestión automática de tokens** con renovación transparente
- **Almacenamiento seguro** de credenciales

### ✅ Configuración Externa

- **Archivos YAML/JSON** con validación automática
- **Variables de entorno** para credenciales sensibles
- **Configuración por entorno** (dev, staging, prod)
- **Validación robusta** con mensajes de error claros

### ✅ Gestión de Recursos

- **Pool de conexiones** con límites configurables
- **Timeouts inteligentes** para evitar bloqueos
- **Rate limiting** configurable por proveedor
- **Cleanup automático** de recursos

### ✅ Robustez Empresarial

- **Logging estructurado** con RequestID únicos
- **Manejo de errores** tipificado y detallado
- **Reintentos automáticos** con backoff exponencial
- **Validación completa** de mensajes y configuración

### ✅ Soporte Completo de Correo

- **Adjuntos múltiples** con encoding automático
- **HTML y texto plano** en el mismo mensaje
- **Headers personalizados** con sanitización
- **RFC822 compliant** para máxima compatibilidad

## 📦 Instalación

```bash
go get github.com/yourorg/smtp-client
```

## 🛠 Configuración Rápida

### 1. Clonar archivos de configuración

```bash
# Copiar archivo de configuración base
cp configs/config.yaml.example configs/config.yaml

# Copiar variables de entorno
cp configs/.env.example .env
```

### 2. Configurar variables de entorno

```bash
# Editar archivo .env con tus credenciales
nano .env
```

```bash
# Gmail OAuth2
GMAIL_CLIENT_ID=tu-client-id.apps.googleusercontent.com
GMAIL_CLIENT_SECRET=tu-client-secret

# Office 365 OAuth2  
OFFICE365_CLIENT_ID=tu-office365-client-id
OFFICE365_CLIENT_SECRET=tu-office365-client-secret
OFFICE365_TENANT_ID=tu-tenant-id

# Seguridad
TOKEN_ENCRYPTION_KEY=tu-clave-de-32-caracteres-aqui
```

### 3. Uso básico

```go
package main

import (
    "context"
    "log"

    "github.com/yourorg/smtp-client/config"
    "github.com/yourorg/smtp-client/pkg/auth"
    "github.com/yourorg/smtp-client/pkg/client"
    "github.com/yourorg/smtp-client/pkg/logger"
    "golang.org/x/oauth2"
)

func main() {
    // Cargar configuración
    cfg, credentials, err := config.LoadDefault()
    if err != nil {
        log.Fatal("Error cargando configuración:", err)
    }

    // Validar configuración
    validator := config.NewConfigValidator()
    if err := validator.ValidateConfig(cfg); err != nil {
        log.Fatal("Error validando configuración:", err)
    }

    // Crear logger
    logger := logger.NewLogrusLogger(&cfg.Logging)

    // Configurar autenticación Gmail
    authConfig := &auth.AuthConfig{
        Provider:     auth.ProviderGmail,
        ClientID:     credentials.GmailClientID,
        ClientSecret: credentials.GmailClientSecret,
        RedirectURL:  cfg.Providers.Gmail.OAuth2.RedirectURL,
        Scopes:       cfg.Providers.Gmail.OAuth2.Scopes,
        Timeout:      cfg.Providers.Gmail.Timeout,
    }

    // Token (en producción, cargar desde almacenamiento seguro)
    token := &oauth2.Token{
        AccessToken:  "ya29.a0AfH6SMC...",
        RefreshToken: "1//04...",
        TokenType:    "Bearer",
    }

    // Crear token manager
    tokenManager, err := auth.NewOAuth2TokenManager(authConfig, nil, logger)
    if err != nil {
        log.Fatal("Error creando token manager:", err)
    }

    // Configurar cliente SMTP
    smtpConfig := &client.SMTPConfig{
        Host:     cfg.Providers.Gmail.SMTPHost,
        Port:     cfg.Providers.Gmail.SMTPPort,
        UseTLS:   cfg.Providers.Gmail.UseTLS,
        Provider: auth.ProviderGmail,
        Timeout:  cfg.Providers.Gmail.Timeout,
    }

    poolConfig := &client.PoolConfig{
        MaxConnections: cfg.ConnectionPool.MaxConnections,
        MaxIdle:        cfg.ConnectionPool.MaxIdleTime,
        Timeout:        cfg.ConnectionPool.ConnectionTimeout,
    }

    // Crear cliente
    smtpClient := client.NewSMTPClient(smtpConfig, poolConfig, tokenManager, logger)
    defer smtpClient.Close()

    // Crear mensaje
    message := &client.EmailMessage{
        From:    "remitente@gmail.com",
        To:      []string{"destinatario@example.com"},
        Subject: "¡Hola desde Go!",
        Body:    "Este es un mensaje de prueba enviado con el cliente SMTP en Go.",
    }

    // Enviar mensaje
    ctx := context.Background()
    if err := smtpClient.SendMessage(ctx, message); err != nil {
        log.Fatal("Error enviando mensaje:", err)
    }

    log.Println("✅ Mensaje enviado exitosamente")
}
```

## 📋 Ejemplos Avanzados

### Mensaje con HTML y Adjuntos

```go
message := &client.EmailMessage{
    From:    "remitente@empresa.com",
    To:      []string{"cliente@example.com"},
    Cc:      []string{"supervisor@empresa.com"},
    Subject: "Reporte Mensual - Marzo 2024",
    HTMLBody: `
        <h1>Reporte Mensual</h1>
        <p>Estimado cliente,</p>
        <p>Adjunto encontrará el <strong>reporte mensual</strong> correspondiente a Marzo 2024.</p>
        <ul>
            <li>Ventas: $50,000</li>
            <li>Crecimiento: +15%</li>
            <li>Nuevos clientes: 127</li>
        </ul>
        <p>Saludos cordiales,<br>Equipo de Ventas</p>
    `,
    Attachments: []client.FileAttachment{
        {
            Filename:    "reporte-marzo-2024.pdf",
            ContentType: "application/pdf",
            Data:        pdfData, // []byte con el contenido del PDF
        },
        {
            Filename:    "grafico-ventas.png",
            ContentType: "image/png",
            Data:        imageData, // []byte con el contenido de la imagen
        },
    },
    Headers: map[string]string{
        "X-Priority":    "1",
        "X-Mailer":     "SMTP-Client-Go/1.0",
        "Reply-To":     "soporte@empresa.com",
    },
}
```

### Configuración para Office 365

```go
// Configurar para Office 365
authConfig := &auth.AuthConfig{
    Provider:     auth.ProviderOffice365,
    ClientID:     credentials.Office365ClientID,
    ClientSecret: credentials.Office365ClientSecret,
    TenantID:     credentials.Office365TenantID,
    RedirectURL:  cfg.Providers.Office365.OAuth2.RedirectURL,
    Scopes:       cfg.Providers.Office365.OAuth2.Scopes,
    Timeout:      cfg.Providers.Office365.Timeout,
}

smtpConfig := &client.SMTPConfig{
    Host:     cfg.Providers.Office365.SMTPHost, // smtp.office365.com
    Port:     cfg.Providers.Office365.SMTPPort, // 587
    UseTLS:   cfg.Providers.Office365.UseTLS,   // true
    Provider: auth.ProviderOffice365,
    Timeout:  cfg.Providers.Office365.Timeout,
}
```

### Servidor SMTP Genérico (sin OAuth2)

```go
smtpConfig := &client.SMTPConfig{
    Host:     "smtp.miservidor.com",
    Port:     587,
    UseTLS:   true,
    Provider: "generic",
    Username: "usuario@miservidor.com",
    Password: "mi-password-seguro",
    Timeout:  30 * time.Second,
}

// Crear cliente sin OAuth2
smtpClient := client.NewGenericSMTPClient(smtpConfig, poolConfig, logger)
```

### Envío Masivo con Control de Rate

```go
func enviarReportesMasivos(clientes []Cliente, smtpClient *client.SMTPClient) error {
    ctx := context.Background()

    // Limitar a 10 correos simultáneos
    semaphore := make(chan struct{}, 10)

    for _, cliente := range clientes {
        semaphore <- struct{}{} // Acquire

        go func(c Cliente) {
            defer func() { <-semaphore }() // Release

            message := &client.EmailMessage{
                From:    "reportes@empresa.com",
                To:      []string{c.Email},
                Subject: fmt.Sprintf("Reporte personalizado para %s", c.Nombre),
                HTMLBody: generarReporteHTML(c),
            }

            if err := smtpClient.SendMessage(ctx, message); err != nil {
                log.Printf("Error enviando a %s: %v", c.Email, err)
            } else {
                log.Printf("✅ Reporte enviado a %s", c.Email)
            }
        }(cliente)

        // Rate limiting manual
        time.Sleep(100 * time.Millisecond)
    }

    return nil
}
```

## ⚙️ Configuración Avanzada

### Configuración por Entornos

```bash
# Desarrollo
SMTP_CLIENT_ENV=development go run main.go

# Staging  
SMTP_CLIENT_ENV=staging go run main.go

# Producción
SMTP_CLIENT_ENV=production go run main.go
```

### Archivo de Configuración Personalizado

```go
// Cargar configuración específica
cfg, creds, err := config.LoadFromPath("./configs/config.prod.yaml")

// Cargar con archivo .env específico
cfg, creds, err := config.LoadWithEnv("./configs/.env.production")
```

### Configuración YAML Completa

```yaml
app:
  name: "smtp-client"
  version: "1.0.0"
  environment: "production"

logging:
  level: "info"
  format: "json"
  output: "file"
  file_path: "/var/log/smtp-client.log"

providers:
  gmail:
    smtp_host: "smtp.gmail.com"
    smtp_port: 587
    use_tls: true
    timeout: "30s"
    oauth2:
      scopes:
        - "https://www.googleapis.com/auth/gmail.send"
      redirect_url: "http://localhost:8080/callback"

connection_pool:
  max_connections: 50
  max_idle_time: "10m"
  connection_timeout: "30s"
  idle_timeout: "10m"
  cleanup_interval: "1m"

security:
  max_token_age: "1h"
  require_tls: true
  allowed_domains:
    - "empresa.com"
    - "clienteimportante.com"

rate_limiting:
  enabled: true
  requests_per_minute: 300
  burst: 50

retry:
  max_attempts: 3
  initial_backoff: "1s"
  max_backoff: "30s"
  multiplier: 2.0
```

## 🔒 Seguridad

### Mejores Prácticas Implementadas

- **🔐 Encriptación de tokens** en almacenamiento
- **🚫 Sin credenciales en código** - solo variables de entorno
- **🔒 TLS obligatorio** en producción
- **✅ Validación estricta** de entrada de datos
- **🛡️ Sanitización de headers** para prevenir inyecciones
- **⏱️ Timeouts configurables** para evitar ataques DoS
- **📝 Logging sin credenciales** sensibles

### Configuración de Permisos OAuth2

#### Gmail

1. Ir a [Google Cloud Console](https://console.cloud.google.com)
2. Crear proyecto o seleccionar existente
3. Habilitar Gmail API
4. Crear credenciales OAuth2
5. Configurar pantalla de consentimiento
6. Scopes requeridos: `https://www.googleapis.com/auth/gmail.send`

#### Office 365

1. Ir a [Azure Portal](https://portal.azure.com)
2. Registrar aplicación en Azure AD
3. Configurar permisos de API
4. Permisos requeridos: `SMTP.Send`
5. Configurar URLs de redirección

### Variables de Entorno Críticas

```bash
# ⚠️ MANTENER ESTAS VARIABLES SEGURAS
TOKEN_ENCRYPTION_KEY=clave-aleatoria-de-32-caracteres-minimo
GMAIL_CLIENT_SECRET=tu-client-secret-de-gmail
OFFICE365_CLIENT_SECRET=tu-client-secret-de-office365

# 💡 Usar diferentes claves por entorno
TOKEN_ENCRYPTION_KEY_DEV=clave-desarrollo
TOKEN_ENCRYPTION_KEY_PROD=clave-produccion-diferente
```

## 📊 Monitoreo y Observabilidad

### Logging Estructurado

Todos los logs incluyen RequestID únicos para trazabilidad:

```json
{
  "level": "info",
  "msg": "Mensaje enviado exitosamente",
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "to": ["destinatario@example.com"],
  "subject": "Test Message",
  "provider": "gmail",
  "timestamp": "2024-03-15T10:30:00Z"
}
```

### Métricas Disponibles

El cliente expone métricas Prometheus en `/metrics`:

- `smtp_messages_sent_total` - Total de mensajes enviados
- `smtp_messages_failed_total` - Total de mensajes fallidos
- `smtp_send_duration_seconds` - Duración de envío de mensajes
- `smtp_auth_failures_total` - Fallos de autenticación
- `smtp_token_refreshes_total` - Renovaciones de token
- `smtp_connection_pool_active` - Conexiones activas en pool
- `smtp_connection_pool_idle` - Conexiones inactivas en pool

### Health Checks

```go
// Endpoint de salud
func healthCheck(smtpClient *client.SMTPClient) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
        defer cancel()

        if err := smtpClient.HealthCheck(ctx); err != nil {
            w.WriteHeader(http.StatusServiceUnavailable)
            json.NewEncoder(w).Encode(map[string]string{
                "status": "unhealthy",
                "error":  err.Error(),
            })
            return
        }

        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]string{
            "status": "healthy",
        })
    }
}
```

## 🔧 Solución de Problemas

### Errores Comunes

#### Error de Autenticación OAuth2

```bash
Error: [authentication:OAUTH2_AUTH_FAILED] Error en autenticación OAuth2
```

**Solución:**

1. Verificar que las credenciales OAuth2 sean correctas
2. Confirmar que los scopes estén configurados en el proveedor
3. Verificar que el token no haya expirado
4. Revisar la configuración del redirect URL

#### Error de Conexión SMTP

```bash
Error: [connection:CONNECTION_FAILED] Error conectando a smtp.gmail.com:587
```

**Solución:**

1. Verificar conectividad de red
2. Confirmar que el puerto 587 esté abierto
3. Verificar configuración de firewall
4. Revisar configuración TLS

#### Error de Validación de Configuración

```bash
Error: errores de validación: Campo 'SMTPPort' debe ser mayor a 0
```

**Solución:**

1. Revisar archivo de configuración YAML/JSON
2. Verificar que todos los campos requeridos estén presentes
3. Validar tipos de datos y rangos de valores
4. Verificar sintaxis YAML/JSON

### Logs de Debug

Habilitar logs detallados para troubleshooting:

```yaml
logging:
  level: "debug"  # Cambiar de "info" a "debug"
```

### Validación de Configuración

Ejecutar validación manual de configuración:

```go
func main() {
    cfg, creds, err := config.LoadDefault()
    if err != nil {
        log.Fatal("Error cargando configuración:", err)
    }

    validator := config.NewConfigValidator()

    if err := validator.ValidateConfig(cfg); err != nil {
        log.Fatal("❌ Configuración inválida:", err)
    }

    if err := validator.ValidateCredentials(creds, cfg); err != nil {
        log.Fatal("❌ Credenciales inválidas:", err)
    }

    log.Println("✅ Configuración válida")
}
```

## 🤝 Contribuir

1. Fork el proyecto
2. Crear feature branch (`git checkout -b feature/nueva-funcionalidad`)
3. Commit cambios (`git commit -am 'Agregar nueva funcionalidad'`)
4. Push al branch (`git push origin feature/nueva-funcionalidad`)
5. Crear Pull Request

### Guías de Desarrollo

- Usar Go 1.21+
- Seguir convenciones de Go (gofmt, golint)
- Agregar tests para nueva funcionalidad
- Actualizar documentación
- Mantener cobertura de tests > 80%

## 📄 Licencia

Este proyecto está bajo la licencia MIT. Ver [LICENSE](LICENSE) para detalles.

## 🆘 Soporte

- **Documentación**: [Wiki del proyecto](https://github.com/yourorg/smtp-client/wiki)
- **Issues**: [GitHub Issues](https://github.com/yourorg/smtp-client/issues)
- **Discusiones**: [GitHub Discussions](https://github.com/yourorg/smtp-client/discussions)
- **Email**: soporte@yourorg.com
