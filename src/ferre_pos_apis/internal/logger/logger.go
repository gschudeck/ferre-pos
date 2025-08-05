package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
	
	"ferre_pos_apis/internal/config"
)

// Logger interfaz personalizada para logging
type Logger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
	Panic(args ...interface{})
	
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Panicf(format string, args ...interface{})
	
	WithField(key string, value interface{}) Logger
	WithFields(fields map[string]interface{}) Logger
	WithError(err error) Logger
	WithRequestID(requestID string) Logger
	WithUserID(userID string) Logger
	WithAPI(apiName string) Logger
}

// loggerImpl implementación del logger usando Logrus
type loggerImpl struct {
	entry *logrus.Entry
}

var globalLogger Logger

// Init inicializa el sistema de logging
func Init(cfg *config.LoggingConfig, apiName string) error {
	logger := logrus.New()

	// Configurar nivel de logging
	level, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		return fmt.Errorf("nivel de logging inválido '%s': %w", cfg.Level, err)
	}
	logger.SetLevel(level)

	// Configurar formato
	switch strings.ToLower(cfg.Format) {
	case "json":
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339Nano,
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "timestamp",
				logrus.FieldKeyLevel: "level",
				logrus.FieldKeyMsg:   "message",
				logrus.FieldKeyFunc:  "function",
				logrus.FieldKeyFile:  "file",
			},
		})
	case "text":
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: time.RFC3339,
		})
	default:
		return fmt.Errorf("formato de logging no soportado: %s", cfg.Format)
	}

	// Configurar salida
	var output io.Writer
	switch strings.ToLower(cfg.Output) {
	case "stdout":
		output = os.Stdout
	case "file":
		if cfg.FilePath == "" {
			return fmt.Errorf("ruta de archivo requerida para salida 'file'")
		}
		
		// Crear directorio si no existe
		if err := os.MkdirAll(cfg.FilePath, 0755); err != nil {
			return fmt.Errorf("error creando directorio de logs: %w", err)
		}
		
		// Configurar rotación de logs
		logFile := filepath.Join(cfg.FilePath, fmt.Sprintf("%s.log", apiName))
		output = &lumberjack.Logger{
			Filename:   logFile,
			MaxSize:    cfg.MaxSizeMB,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAgeDays,
			Compress:   cfg.Compress,
		}
	default:
		return fmt.Errorf("salida de logging no soportada: %s", cfg.Output)
	}

	logger.SetOutput(output)

	// Configurar hooks personalizados
	logger.AddHook(&contextHook{})
	logger.AddHook(&callerHook{})

	// Crear logger global con campos base
	globalLogger = &loggerImpl{
		entry: logger.WithFields(logrus.Fields{
			"api":     apiName,
			"service": "ferre-pos",
			"version": getVersion(),
		}),
	}

	return nil
}

// Get obtiene el logger global
func Get() Logger {
	if globalLogger == nil {
		// Logger por defecto si no se ha inicializado
		logrus.SetLevel(logrus.InfoLevel)
		logrus.SetFormatter(&logrus.JSONFormatter{})
		globalLogger = &loggerImpl{
			entry: logrus.WithFields(logrus.Fields{
				"api":     "unknown",
				"service": "ferre-pos",
			}),
		}
	}
	return globalLogger
}

// Implementación de métodos de Logger

func (l *loggerImpl) Debug(args ...interface{}) {
	l.entry.Debug(args...)
}

func (l *loggerImpl) Info(args ...interface{}) {
	l.entry.Info(args...)
}

func (l *loggerImpl) Warn(args ...interface{}) {
	l.entry.Warn(args...)
}

func (l *loggerImpl) Error(args ...interface{}) {
	l.entry.Error(args...)
}

func (l *loggerImpl) Fatal(args ...interface{}) {
	l.entry.Fatal(args...)
}

func (l *loggerImpl) Panic(args ...interface{}) {
	l.entry.Panic(args...)
}

func (l *loggerImpl) Debugf(format string, args ...interface{}) {
	l.entry.Debugf(format, args...)
}

func (l *loggerImpl) Infof(format string, args ...interface{}) {
	l.entry.Infof(format, args...)
}

func (l *loggerImpl) Warnf(format string, args ...interface{}) {
	l.entry.Warnf(format, args...)
}

func (l *loggerImpl) Errorf(format string, args ...interface{}) {
	l.entry.Errorf(format, args...)
}

func (l *loggerImpl) Fatalf(format string, args ...interface{}) {
	l.entry.Fatalf(format, args...)
}

func (l *loggerImpl) Panicf(format string, args ...interface{}) {
	l.entry.Panicf(format, args...)
}

func (l *loggerImpl) WithField(key string, value interface{}) Logger {
	return &loggerImpl{
		entry: l.entry.WithField(key, value),
	}
}

func (l *loggerImpl) WithFields(fields map[string]interface{}) Logger {
	return &loggerImpl{
		entry: l.entry.WithFields(fields),
	}
}

func (l *loggerImpl) WithError(err error) Logger {
	return &loggerImpl{
		entry: l.entry.WithError(err),
	}
}

func (l *loggerImpl) WithRequestID(requestID string) Logger {
	return l.WithField("request_id", requestID)
}

func (l *loggerImpl) WithUserID(userID string) Logger {
	return l.WithField("user_id", userID)
}

func (l *loggerImpl) WithAPI(apiName string) Logger {
	return l.WithField("api", apiName)
}

// contextHook hook para agregar información de contexto
type contextHook struct{}

func (h *contextHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *contextHook) Fire(entry *logrus.Entry) error {
	// Agregar timestamp en formato Unix para facilitar análisis
	entry.Data["timestamp_unix"] = entry.Time.Unix()
	
	// Agregar información del proceso
	entry.Data["pid"] = os.Getpid()
	
	// Agregar información de memoria si es nivel debug
	if entry.Level <= logrus.DebugLevel {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		entry.Data["memory_alloc"] = m.Alloc
		entry.Data["memory_sys"] = m.Sys
		entry.Data["num_gc"] = m.NumGC
	}
	
	return nil
}

// callerHook hook para agregar información del caller
type callerHook struct{}

func (h *callerHook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.ErrorLevel,
		logrus.FatalLevel,
		logrus.PanicLevel,
	}
}

func (h *callerHook) Fire(entry *logrus.Entry) error {
	// Obtener información del caller
	if pc, file, line, ok := runtime.Caller(8); ok {
		funcName := runtime.FuncForPC(pc).Name()
		
		// Simplificar el nombre de la función
		if lastSlash := strings.LastIndex(funcName, "/"); lastSlash >= 0 {
			funcName = funcName[lastSlash+1:]
		}
		
		// Simplificar el path del archivo
		if lastSlash := strings.LastIndex(file, "/"); lastSlash >= 0 {
			file = file[lastSlash+1:]
		}
		
		entry.Data["caller"] = fmt.Sprintf("%s:%d", file, line)
		entry.Data["function"] = funcName
	}
	
	return nil
}

// getVersion obtiene la versión de la aplicación
func getVersion() string {
	// En un entorno real, esto podría venir de variables de build
	return "1.0.0"
}

// LogRequest middleware para logging de requests HTTP
func LogRequest(logger Logger, method, path, userAgent, clientIP string, statusCode int, duration time.Duration, requestID string) {
	fields := map[string]interface{}{
		"method":      method,
		"path":        path,
		"status_code": statusCode,
		"duration_ms": duration.Milliseconds(),
		"user_agent":  userAgent,
		"client_ip":   clientIP,
		"request_id":  requestID,
	}

	logLevel := getLogLevelForStatusCode(statusCode)
	
	switch logLevel {
	case logrus.ErrorLevel:
		logger.WithFields(fields).Error("HTTP request completed with error")
	case logrus.WarnLevel:
		logger.WithFields(fields).Warn("HTTP request completed with warning")
	default:
		logger.WithFields(fields).Info("HTTP request completed")
	}
}

// getLogLevelForStatusCode determina el nivel de log basado en el código de estado HTTP
func getLogLevelForStatusCode(statusCode int) logrus.Level {
	switch {
	case statusCode >= 500:
		return logrus.ErrorLevel
	case statusCode >= 400:
		return logrus.WarnLevel
	default:
		return logrus.InfoLevel
	}
}

// LogError helper para logging de errores con stack trace
func LogError(logger Logger, err error, message string, fields map[string]interface{}) {
	if fields == nil {
		fields = make(map[string]interface{})
	}
	
	// Agregar stack trace si es posible
	if stackTracer, ok := err.(interface{ StackTrace() []uintptr }); ok {
		fields["stack_trace"] = formatStackTrace(stackTracer.StackTrace())
	}
	
	logger.WithFields(fields).WithError(err).Error(message)
}

// formatStackTrace formatea un stack trace para logging
func formatStackTrace(stackTrace []uintptr) []string {
	var frames []string
	for _, pc := range stackTrace {
		fn := runtime.FuncForPC(pc)
		if fn != nil {
			file, line := fn.FileLine(pc)
			frames = append(frames, fmt.Sprintf("%s:%d %s", file, line, fn.Name()))
		}
	}
	return frames
}

// NewRequestLogger crea un logger específico para una request
func NewRequestLogger(baseLogger Logger, requestID, userID, method, path string) Logger {
	return baseLogger.
		WithRequestID(requestID).
		WithUserID(userID).
		WithField("method", method).
		WithField("path", path)
}

