package client

import (
    "encoding/base64"
    "fmt"
    "mime"
    "strings"
    "time"
    
    "github.com/yourorg/smtp-client/pkg/errors"
)

// FileAttachment representa un archivo adjunto
type FileAttachment struct {
    Filename    string
    ContentType string
    Data        []byte
}

// EmailMessage representa un mensaje de correo electrónico
type EmailMessage struct {
    From        string
    To          []string
    Cc          []string
    Bcc         []string
    Subject     string
    Body        string
    HTMLBody    string
    Attachments []FileAttachment
    Headers     map[string]string
}

// MessageBuilder construye mensajes de correo
type MessageBuilder struct {
    validator *MessageValidator
}

// NewMessageBuilder crea un nuevo constructor de mensajes
func NewMessageBuilder() *MessageBuilder {
    return &MessageBuilder{
        validator: NewMessageValidator(),
    }
}

// BuildRFC822Message convierte un mensaje a formato RFC822
func (mb *MessageBuilder) BuildRFC822Message(msg *EmailMessage, requestID string) (string, error) {
    if err := mb.validator.ValidateMessage(msg, requestID); err != nil {
        return "", err
    }
    
    var builder strings.Builder
    
    // Headers básicos
    mb.writeBasicHeaders(&builder, msg)
    
    // Headers personalizados
    mb.writeCustomHeaders(&builder, msg.Headers)
    
    // Contenido del mensaje
    if err := mb.writeMessageContent(&builder, msg, requestID); err != nil {
        return "", err
    }
    
    return builder.String(), nil
}

func (mb *MessageBuilder) writeBasicHeaders(builder *strings.Builder, msg *EmailMessage) {
    builder.WriteString(fmt.Sprintf("From: %s\r\n", msg.From))
    builder.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(msg.To, ", ")))
    
    if len(msg.Cc) > 0 {
        builder.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(msg.Cc, ", ")))
    }
    
    builder.WriteString(fmt.Sprintf("Subject: %s\r\n", mime.QEncoding.Encode("utf-8", msg.Subject)))
    builder.WriteString(fmt.Sprintf("Date: %s\r\n", time.Now().Format(time.RFC1123Z)))
    builder.WriteString("MIME-Version: 1.0\r\n")
}

func (mb *MessageBuilder) writeCustomHeaders(builder *strings.Builder, headers map[string]string) {
    for key, value := range headers {
        // Sanitizar headers para prevenir inyección
        sanitizedValue := mb.sanitizeHeaderValue(value)
        builder.WriteString(fmt.Sprintf("%s: %s\r\n", key, sanitizedValue))
    }
}

func (mb *MessageBuilder) sanitizeHeaderValue(value string) string {
    // Remover caracteres de control y CRLF
    sanitized := strings.ReplaceAll(value, "\r", "")
    sanitized = strings.ReplaceAll(sanitized, "\n", "")
    return sanitized
}

func (mb *MessageBuilder) writeMessageContent(builder *strings.Builder, msg *EmailMessage, requestID string) error {
    if len(msg.Attachments) == 0 {
        return mb.writeSimpleContent(builder, msg)
    }
    
    return mb.writeMultipartContent(builder, msg, requestID)
}

func (mb *MessageBuilder) writeSimpleContent(builder *strings.Builder, msg *EmailMessage) error {
    if msg.HTMLBody != "" {
        builder.WriteString("Content-Type: text/html; charset=utf-8\r\n")
        builder.WriteString("\r\n")
        builder.WriteString(msg.HTMLBody)
    } else {
        builder.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
        builder.WriteString("\r\n")
        builder.WriteString(msg.Body)
    }
    
    return nil
}

func (mb *MessageBuilder) writeMultipartContent(builder *strings.Builder, msg *EmailMessage, requestID string) error {
    boundary := fmt.Sprintf("boundary_%d", time.Now().UnixNano())
    
    builder.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\r\n", boundary))
    builder.WriteString("\r\n")
    
    // Escribir cuerpo del mensaje
    mb.writeMessagePart(builder, boundary, msg)
    
    // Escribir adjuntos
    for _, attachment := range msg.Attachments {
        if err := mb.writeAttachmentPart(builder, boundary, attachment, requestID); err != nil {
            return err
        }
    }
    
    builder.WriteString(fmt.Sprintf("--%s--\r\n", boundary))
    
    return nil
}

func (mb *MessageBuilder) writeMessagePart(builder *strings.Builder, boundary string, msg *EmailMessage) {
    builder.WriteString(fmt.Sprintf("--%s\r\n", boundary))
    
    if msg.HTMLBody != "" {
        builder.WriteString("Content-Type: text/html; charset=utf-8\r\n")
        builder.WriteString("\r\n")
        builder.WriteString(msg.HTMLBody)
    } else {
        builder.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
        builder.WriteString("\r\n")
        builder.WriteString(msg.Body)
    }
    
    builder.WriteString("\r\n")
}

func (mb *MessageBuilder) writeAttachmentPart(builder *strings.Builder, boundary string, attachment FileAttachment, requestID string) error {
    builder.WriteString(fmt.Sprintf("--%s\r\n", boundary))
    builder.WriteString(fmt.Sprintf("Content-Type: %s\r\n", attachment.ContentType))
    builder.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n", attachment.Filename))
    builder.WriteString("Content-Transfer-Encoding: base64\r\n")
    builder.WriteString("\r\n")
    
    // Codificar datos en base64
    encoded := base64.StdEncoding.EncodeToString(attachment.Data)
    
    // Dividir en líneas de 76 caracteres según RFC
    for i := 0; i < len(encoded); i += 76 {
        end := i + 76
        if end > len(encoded) {
            end = len(encoded)
        }
        builder.WriteString(encoded[i:end])
        builder.WriteString("\r\n")
    }
    
    return nil
}
