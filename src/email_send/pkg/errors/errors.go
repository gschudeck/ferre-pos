package errors

import (
    "fmt"
)

// ErrorType representa el tipo de error
type ErrorType string

const (
    ErrorTypeValidation     ErrorType = "validation"
    ErrorTypeAuthentication ErrorType = "authentication"
    ErrorTypeConnection     ErrorType = "connection"
    ErrorTypeTimeout        ErrorType = "timeout"
    ErrorTypeProvider       ErrorType = "provider"
    ErrorTypeMessage        ErrorType = "message"
)

// SMTPError representa un error específico del cliente SMTP
type SMTPError struct {
    Type      ErrorType
    Code      string
    Message   string
    RequestID string
    Err       error
}

func (e *SMTPError) Error() string {
    if e.Err != nil {
        return fmt.Sprintf("[%s:%s] %s (RequestID: %s): %v", e.Type, e.Code, e.Message, e.RequestID, e.Err)
    }
    return fmt.Sprintf("[%s:%s] %s (RequestID: %s)", e.Type, e.Code, e.Message, e.RequestID)
}

func (e *SMTPError) Unwrap() error {
    return e.Err
}

// NewValidationError crea un error de validación
func NewValidationError(code, message, requestID string, err error) *SMTPError {
    return &SMTPError{
        Type:      ErrorTypeValidation,
        Code:      code,
        Message:   message,
        RequestID: requestID,
        Err:       err,
    }
}

// NewAuthenticationError crea un error de autenticación
func NewAuthenticationError(code, message, requestID string, err error) *SMTPError {
    return &SMTPError{
        Type:      ErrorTypeAuthentication,
        Code:      code,
        Message:   message,
        RequestID: requestID,
        Err:       err,
    }
}

// NewConnectionError crea un error de conexión
func NewConnectionError(code, message, requestID string, err error) *SMTPError {
    return &SMTPError{
        Type:      ErrorTypeConnection,
        Code:      code,
        Message:   message,
        RequestID: requestID,
        Err:       err,
    }
}

// NewTimeoutError crea un error de timeout
func NewTimeoutError(code, message, requestID string, err error) *SMTPError {
    return &SMTPError{
        Type:      ErrorTypeTimeout,
        Code:      code,
        Message:   message,
        RequestID: requestID,
        Err:       err,
    }
}