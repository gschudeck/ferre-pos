// Package logger proporciona funcionalidades avanzadas de logging
// con notación húngara y configuración flexible
package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// enumLogLevel define los niveles de logging
type enumLogLevel string

const (
	EnumLogLevelDebug enumLogLevel = "debug"
	EnumLogLevelInfo  enumLogLevel = "info"
	EnumLogLevelWarn  enumLogLevel = "warn"
	EnumLogLevelError enumLogLevel = "error"
	EnumLogLevelFatal enumLogLevel = "fatal"
)

// enumLogFormat define los formatos de logging
type enumLogFormat string

const (
	EnumLogFormatJSON    enumLogFormat = "json"
	EnumLogFormatConsole enumLogFormat = "console"
)

// enumLogOutput define las salidas de logging
type enumLogOutput string

const (
	EnumLogOutputStdout enumLogOutput = "stdout"
	EnumLogOutputStderr enumLogOutput = "stderr"
	EnumLogOutputFile   enumLogOutput = "file"
	EnumLogOutputBoth   enumLogOutput = "both"
)

// Config contiene la configuración del logger
type Config struct {
	Level      string `yaml:"level" json:"level"`
	Format     string `yaml:"format" json:"format"`
	Output     string `yaml:"output" json:"output"`
	FilePath   string `yaml:"file_path" json:"file_path"`
	MaxSize    int    `yaml:"max_size" json:"max_size"`       // MB
	MaxBackups int    `yaml:"max_backups" json:"max_backups"` // archivos
	MaxAge     int    `yaml:"max_age" json:"max_age"`         // días
	Compress   bool   `yaml:"compress" json:"compress"`
}

// structLogger encapsula el logger de Zap con configuración
type structLogger struct {
	ptrZapLogger *zap.Logger
	ptrConfig    *Config
	strComponent string
}

// InterfaceLogger define la interfaz del logger
type InterfaceLogger interface {
	Debug(strMsg string, arrFields ...zap.Field)
	Info(strMsg string, arrFields ...zap.Field)
	Warn(strMsg string, arrFields ...zap.Field)
	Error(strMsg string, arrFields ...zap.Field)
	Fatal(strMsg string, arrFields ...zap.Field)

	Debugf(strFormat string, arrArgs ...interface{})
	Infof(strFormat string, arrArgs ...interface{})
	Warnf(strFormat string, arrArgs ...interface{})
	Errorf(strFormat string, arrArgs ...interface{})
	Fatalf(strFormat string, arrArgs ...interface{})

	With(arrFields ...zap.Field) InterfaceLogger
	WithComponent(strComponent string) InterfaceLogger
	WithRequestID(strRequestID string) InterfaceLogger
	WithUserID(strUserID string) InterfaceLogger

	Sync() error
	GetZapLogger() *zap.Logger
}

// NewLogger crea un nuevo logger con la configuración especificada
func NewLogger(ptrConfig *Config) (*zap.Logger, error) {
	// Validar configuración
	if err := validateConfig(ptrConfig); err != nil {
		return nil, fmt.Errorf("configuración de logger inválida: %w", err)
	}

	// Configurar nivel de logging
	enumLevel := parseLogLevel(ptrConfig.Level)
	atomicLevel := zap.NewAtomicLevelAt(enumLevel)

	// Configurar encoder
	ptrEncoderConfig := getEncoderConfig(ptrConfig.Format)
	var ptrEncoder zapcore.Encoder

	switch enumLogFormat(ptrConfig.Format) {
	case EnumLogFormatJSON:
		ptrEncoder = zapcore.NewJSONEncoder(ptrEncoderConfig)
	case EnumLogFormatConsole:
		ptrEncoder = zapcore.NewConsoleEncoder(ptrEncoderConfig)
	default:
		ptrEncoder = zapcore.NewJSONEncoder(ptrEncoderConfig)
	}

	// Configurar writer
	ptrWriter, err := getWriter(ptrConfig)
	if err != nil {
		return nil, fmt.Errorf("error configurando writer: %w", err)
	}

	// Crear core
	ptrCore := zapcore.NewCore(ptrEncoder, ptrWriter, atomicLevel)

	// Configurar opciones
	arrOptions := []zap.Option{
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
		zap.AddCallerSkip(1),
	}

	// Crear logger
	ptrLogger := zap.New(ptrCore, arrOptions...)

	return ptrLogger, nil
}

// NewStructuredLogger crea un logger estructurado con interfaz personalizada
func NewStructuredLogger(ptrConfig *Config, strComponent string) (InterfaceLogger, error) {
	ptrZapLogger, err := NewLogger(ptrConfig)
	if err != nil {
		return nil, err
	}

	return &structLogger{
		ptrZapLogger: ptrZapLogger,
		ptrConfig:    ptrConfig,
		strComponent: strComponent,
	}, nil
}

// Debug registra un mensaje de debug
func (ptrLogger *structLogger) Debug(strMsg string, arrFields ...zap.Field) {
	arrFields = ptrLogger.addComponentField(arrFields)
	ptrLogger.ptrZapLogger.Debug(strMsg, arrFields...)
}

// Info registra un mensaje de información
func (ptrLogger *structLogger) Info(strMsg string, arrFields ...zap.Field) {
	arrFields = ptrLogger.addComponentField(arrFields)
	ptrLogger.ptrZapLogger.Info(strMsg, arrFields...)
}

// Warn registra un mensaje de advertencia
func (ptrLogger *structLogger) Warn(strMsg string, arrFields ...zap.Field) {
	arrFields = ptrLogger.addComponentField(arrFields)
	ptrLogger.ptrZapLogger.Warn(strMsg, arrFields...)
}

// Error registra un mensaje de error
func (ptrLogger *structLogger) Error(strMsg string, arrFields ...zap.Field) {
	arrFields = ptrLogger.addComponentField(arrFields)
	ptrLogger.ptrZapLogger.Error(strMsg, arrFields...)
}

// Fatal registra un mensaje fatal y termina la aplicación
func (ptrLogger *structLogger) Fatal(strMsg string, arrFields ...zap.Field) {
	arrFields = ptrLogger.addComponentField(arrFields)
	ptrLogger.ptrZapLogger.Fatal(strMsg, arrFields...)
}

// Debugf registra un mensaje de debug con formato
func (ptrLogger *structLogger) Debugf(strFormat string, arrArgs ...interface{}) {
	ptrLogger.Debug(fmt.Sprintf(strFormat, arrArgs...))
}

// Infof registra un mensaje de información con formato
func (ptrLogger *structLogger) Infof(strFormat string, arrArgs ...interface{}) {
	ptrLogger.Info(fmt.Sprintf(strFormat, arrArgs...))
}

// Warnf registra un mensaje de advertencia con formato
func (ptrLogger *structLogger) Warnf(strFormat string, arrArgs ...interface{}) {
	ptrLogger.Warn(fmt.Sprintf(strFormat, arrArgs...))
}

// Errorf registra un mensaje de error con formato
func (ptrLogger *structLogger) Errorf(strFormat string, arrArgs ...interface{}) {
	ptrLogger.Error(fmt.Sprintf(strFormat, arrArgs...))
}

// Fatalf registra un mensaje fatal con formato y termina la aplicación
func (ptrLogger *structLogger) Fatalf(strFormat string, arrArgs ...interface{}) {
	ptrLogger.Fatal(fmt.Sprintf(strFormat, arrArgs...))
}

// With agrega campos al logger
func (ptrLogger *structLogger) With(arrFields ...zap.Field) InterfaceLogger {
	return &structLogger{
		ptrZapLogger: ptrLogger.ptrZapLogger.With(arrFields...),
		ptrConfig:    ptrLogger.ptrConfig,
		strComponent: ptrLogger.strComponent,
	}
}

// WithComponent agrega el componente al logger
func (ptrLogger *structLogger) WithComponent(strComponent string) InterfaceLogger {
	return &structLogger{
		ptrZapLogger: ptrLogger.ptrZapLogger,
		ptrConfig:    ptrLogger.ptrConfig,
		strComponent: strComponent,
	}
}

// WithRequestID agrega el ID de request al logger
func (ptrLogger *structLogger) WithRequestID(strRequestID string) InterfaceLogger {
	return ptrLogger.With(zap.String("request_id", strRequestID))
}

// WithUserID agrega el ID de usuario al logger
func (ptrLogger *structLogger) WithUserID(strUserID string) InterfaceLogger {
	return ptrLogger.With(zap.String("user_id", strUserID))
}

// Sync sincroniza el logger
func (ptrLogger *structLogger) Sync() error {
	return ptrLogger.ptrZapLogger.Sync()
}

// GetZapLogger retorna el logger de Zap subyacente
func (ptrLogger *structLogger) GetZapLogger() *zap.Logger {
	return ptrLogger.ptrZapLogger
}

// addComponentField agrega el campo de componente si está configurado
func (ptrLogger *structLogger) addComponentField(arrFields []zap.Field) []zap.Field {
	if ptrLogger.strComponent != "" {
		arrFields = append(arrFields, zap.String("component", ptrLogger.strComponent))
	}
	return arrFields
}

// validateConfig valida la configuración del logger
func validateConfig(ptrConfig *Config) error {
	if ptrConfig == nil {
		return fmt.Errorf("configuración no puede ser nil")
	}

	// Validar nivel
	enumLevel := enumLogLevel(strings.ToLower(ptrConfig.Level))
	switch enumLevel {
	case EnumLogLevelDebug, EnumLogLevelInfo, EnumLogLevelWarn, EnumLogLevelError, EnumLogLevelFatal:
		// Válido
	default:
		return fmt.Errorf("nivel de log inválido: %s", ptrConfig.Level)
	}

	// Validar formato
	enumFormat := enumLogFormat(strings.ToLower(ptrConfig.Format))
	switch enumFormat {
	case EnumLogFormatJSON, EnumLogFormatConsole:
		// Válido
	default:
		return fmt.Errorf("formato de log inválido: %s", ptrConfig.Format)
	}

	// Validar salida
	enumOutput := enumLogOutput(strings.ToLower(ptrConfig.Output))
	switch enumOutput {
	case EnumLogOutputStdout, EnumLogOutputStderr, EnumLogOutputFile, EnumLogOutputBoth:
		// Válido
	default:
		return fmt.Errorf("salida de log inválida: %s", ptrConfig.Output)
	}

	// Validar configuración de archivo si es necesario
	if enumOutput == EnumLogOutputFile || enumOutput == EnumLogOutputBoth {
		if ptrConfig.FilePath == "" {
			return fmt.Errorf("file_path es requerido cuando output es 'file' o 'both'")
		}

		// Crear directorio si no existe
		strDir := filepath.Dir(ptrConfig.FilePath)
		if err := os.MkdirAll(strDir, 0755); err != nil {
			return fmt.Errorf("error creando directorio de logs: %w", err)
		}
	}

	return nil
}

// parseLogLevel convierte string a zapcore.Level
func parseLogLevel(strLevel string) zapcore.Level {
	switch enumLogLevel(strings.ToLower(strLevel)) {
	case EnumLogLevelDebug:
		return zapcore.DebugLevel
	case EnumLogLevelInfo:
		return zapcore.InfoLevel
	case EnumLogLevelWarn:
		return zapcore.WarnLevel
	case EnumLogLevelError:
		return zapcore.ErrorLevel
	case EnumLogLevelFatal:
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// getEncoderConfig retorna la configuración del encoder
func getEncoderConfig(strFormat string) zapcore.EncoderConfig {
	ptrConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	if enumLogFormat(strings.ToLower(strFormat)) == EnumLogFormatConsole {
		ptrConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		ptrConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")
	}

	return ptrConfig
}

// getWriter configura el writer según la configuración
func getWriter(ptrConfig *Config) (zapcore.WriteSyncer, error) {
	enumOutput := enumLogOutput(strings.ToLower(ptrConfig.Output))

	switch enumOutput {
	case EnumLogOutputStdout:
		return zapcore.AddSync(os.Stdout), nil

	case EnumLogOutputStderr:
		return zapcore.AddSync(os.Stderr), nil

	case EnumLogOutputFile:
		return getFileWriter(ptrConfig), nil

	case EnumLogOutputBoth:
		ptrFileWriter := getFileWriter(ptrConfig)
		ptrStdoutWriter := zapcore.AddSync(os.Stdout)
		return zapcore.NewMultiWriteSyncer(ptrFileWriter, ptrStdoutWriter), nil

	default:
		return zapcore.AddSync(os.Stdout), nil
	}
}

// getFileWriter configura el writer de archivo con rotación
func getFileWriter(ptrConfig *Config) zapcore.WriteSyncer {
	ptrLumberjack := &lumberjack.Logger{
		Filename:   ptrConfig.FilePath,
		MaxSize:    ptrConfig.MaxSize,
		MaxBackups: ptrConfig.MaxBackups,
		MaxAge:     ptrConfig.MaxAge,
		Compress:   ptrConfig.Compress,
	}

	return zapcore.AddSync(ptrLumberjack)
}

// structRequestLogger es un logger específico para requests HTTP
type structRequestLogger struct {
	InterfaceLogger
	strRequestID string
	strUserID    string
	strMethod    string
	strPath      string
	timeStart    time.Time
}

// NewRequestLogger crea un logger específico para requests
func NewRequestLogger(ptrBaseLogger InterfaceLogger, strRequestID, strUserID, strMethod, strPath string) *structRequestLogger {
	return &structRequestLogger{
		InterfaceLogger: ptrBaseLogger.WithRequestID(strRequestID).WithUserID(strUserID),
		strRequestID:    strRequestID,
		strUserID:       strUserID,
		strMethod:       strMethod,
		strPath:         strPath,
		timeStart:       time.Now(),
	}
}

// LogStart registra el inicio de la request
func (ptrReqLogger *structRequestLogger) LogStart() {
	ptrReqLogger.Info("Request iniciada",
		zap.String("method", ptrReqLogger.strMethod),
		zap.String("path", ptrReqLogger.strPath),
		zap.Time("start_time", ptrReqLogger.timeStart),
	)
}

// LogEnd registra el final de la request
func (ptrReqLogger *structRequestLogger) LogEnd(intStatusCode int, intResponseSize int64) {
	timeDuration := time.Since(ptrReqLogger.timeStart)

	ptrReqLogger.Info("Request completada",
		zap.String("method", ptrReqLogger.strMethod),
		zap.String("path", ptrReqLogger.strPath),
		zap.Int("status_code", intStatusCode),
		zap.Int64("response_size", intResponseSize),
		zap.Duration("duration", timeDuration),
		zap.Int64("duration_ms", timeDuration.Milliseconds()),
	)
}

// LogError registra un error en la request
func (ptrReqLogger *structRequestLogger) LogError(err error, intStatusCode int) {
	timeDuration := time.Since(ptrReqLogger.timeStart)

	ptrReqLogger.Error("Request falló",
		zap.String("method", ptrReqLogger.strMethod),
		zap.String("path", ptrReqLogger.strPath),
		zap.Int("status_code", intStatusCode),
		zap.Error(err),
		zap.Duration("duration", timeDuration),
		zap.Int64("duration_ms", timeDuration.Milliseconds()),
	)
}
