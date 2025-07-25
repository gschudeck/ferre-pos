package main

import (
    "context"
    "fmt"
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
    
    if err := validator.ValidateCredentials(credentials, cfg); err != nil {
        log.Fatal("Error validando credenciales:", err)
    }
    
    // Crear logger
    loggerInstance := logger.NewLogrusLogger(&cfg.Logging)
    
    // Configurar autenticación Gmail
    providerConfig := cfg.Providers.Gmail
    authConfig := &auth.AuthConfig{
        Provider:     auth.ProviderGmail,
        ClientID:     credentials.GmailClientID,
        ClientSecret: credentials.GmailClientSecret,
        RedirectURL:  providerConfig.OAuth2.RedirectURL,
        Scopes:       providerConfig.OAuth2.Scopes,
        Timeout:      providerConfig.Timeout,
    }
    
    // Token existente (en producción, cargar desde almacenamiento seguro)
    token := &oauth2.Token{
        AccessToken:  "ya29.a0AfH6SMC...",
        RefreshToken: "1//04...",
        TokenType:    "Bearer",
    }
    
    // Crear token manager
    tokenManager, err := auth.NewOAuth2TokenManager(authConfig, nil, loggerInstance)
    if err != nil {
        log.Fatal("Error creando token manager:", err)
    }
    
    // Configurar cliente SMTP
    smtpConfig := &client.SMTPConfig{
        Host:     providerConfig.# Cliente SMTP en Go con OAuth2 - Refactorizado

