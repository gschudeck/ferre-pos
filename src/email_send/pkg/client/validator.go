package client

import (
    "fmt"
    "net/mail"
    "strings"
    
    "github.com/yourorg/smtp-client/pkg/errors"
)

// MessageValidator valida mensajes de correo
type MessageValidator struct{}

// NewMessageValidator crea un nuevo validador de mensajes
func NewMessageValidator() *MessageValidator {
    return &MessageValidator{}
}

// ValidateMessage valida un mensaje completo
func (v *MessageValidator) ValidateMessage(msg *EmailMessage, requestID string) error {
    if msg == nil {
        return errors.NewValidationError(
            "MESSAGE_NIL",
            "El mensaje no puede ser nil",
            requestID,
            nil,
        )
    }
    
    if err := v.validateSender(msg.From, requestID); err != nil {
        return err
    }
    
    if err := v.validateRecipients(msg.To, requestID); err != nil {
        return err
    }
    
    if err := v.validateOptionalRecipients(msg.Cc, "Cc", requestID); err != nil {
        return err
    }
    
    if err := v.validateOptionalRecipients(msg.Bcc, "Bcc", requestID); err != nil {
        return err
    }
    
    if err := v.validateSubject(msg.Subject, requestID); err != nil {
        return err
    }
    
    if err := v.validateContent(msg, requestID); err != nil {
        return err
    }
    
    return nil
}

func (v *MessageValidator) validateSender(from, requestID string) error {
    if strings.TrimSpace(from) == "" {
        return errors.NewValidationError(
            "MISSING_SENDER",
            "El campo From es requerido",
            requestID,
            nil,
        )
    }
    
    if _, err := mail.ParseAddress(from); err != nil {
        return errors.NewValidationError(
            "INVALID_SENDER_FORMAT",
            "Formato de email inválido en From",
            requestID,
            err,
        )
    }
    
    return nil
}

func (v *MessageValidator) validateRecipients(recipients []string, requestID string) error {
    if len(recipients) == 0 {
        return errors.NewValidationError(
            "MISSING_RECIPIENTS",
            "Al menos un destinatario en To es requerido",
            requestID,
            nil,
        )
    }
    
    return v.validateEmailList(recipients, "To", requestID)
}

func (v *MessageValidator) validateOptionalRecipients(recipients []string, fieldName, requestID string) error {
    if len(recipients) == 0 {
        return nil
    }
    
    return v.validateEmailList(recipients, fieldName, requestID)
}

func (v *MessageValidator) validateEmailList(emails []string, fieldName, requestID string) error {
    for i, email := range emails {
        if strings.TrimSpace(email) == "" {
            return errors.NewValidationError(
                "EMPTY_EMAIL",
                fmt.Sprintf("Email vacío en posición %d del campo %s", i, fieldName),
                requestID,
                nil,
            )
        }
        
        if _, err := mail.ParseAddress(email); err != nil {
            return errors.NewValidationError(
                "INVALID_EMAIL_FORMAT",
                fmt.Sprintf("Formato de email inválido '%s' en campo %s", email, fieldName),
                requestID,
                err,
            )
        }
    }
    
    return nil
}

func (v *MessageValidator) validateSubject(subject, requestID string) error {
    if strings.TrimSpace(subject) == "" {
        return errors.NewValidationError(
            "MISSING_SUBJECT",
            "El campo Subject es requerido",
            requestID,
            nil,
        )
    }
    
    // Verificar caracteres de control que podrían causar problemas
    for _, char := range subject {
        if char == '\r' || char == '\n' {
            return errors.NewValidationError(
                "INVALID_SUBJECT_CHARS",
                "El Subject no puede contener caracteres de control (\\r, \\n)",
                requestID,
                nil,
            )
        }
    }
    
    return nil
}

func (v *MessageValidator) validateContent(msg *EmailMessage, requestID string) error {
    if strings.TrimSpace(msg.Body) == "" && strings.TrimSpace(msg.HTMLBody) == "" {
        return errors.NewValidationError(
            "MISSING_CONTENT",
            "El mensaje debe tener contenido en Body o HTMLBody",
            requestID,
            nil,
        )
    }
    
    // Validar adjuntos si existen
    for i, attachment := range msg.Attachments {
        if err := v.validateAttachment(attachment, i, requestID); err != nil {
            return err
        }
    }
    
    return nil
}

func (v *MessageValidator) validateAttachment(attachment FileAttachment, index int, requestID string) error {
    if strings.TrimSpace(attachment.Filename) == "" {
        return errors.NewValidationError(
            "INVALID_ATTACHMENT_FILENAME",
            fmt.Sprintf("Filename vacío en adjunto %d", index),
            requestID,
            nil,
        )
    }
    
    if strings.TrimSpace(attachment.ContentType) == "" {
        return errors.NewValidationError(
            "INVALID_ATTACHMENT_CONTENT_TYPE",
            fmt.Sprintf("ContentType vacío en adjunto %d", index),
            requestID,
            nil,
        )
    }
    
    if len(attachment.Data) == 0 {
        return errors.NewValidationError(
            "INVALID_ATTACHMENT_DATA",
            fmt.Sprintf("Data vacía en adjunto %d", index),
            requestID,
            nil,
        )
    }
    
    return nil
}