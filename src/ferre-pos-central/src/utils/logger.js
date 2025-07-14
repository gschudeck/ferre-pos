/**
 * Sistema de Logging - Sistema Ferre-POS
 * 
 * Configuración centralizada de logging usando Winston
 * con diferentes niveles y formatos según el entorno.
 */

const winston = require('winston')
const path = require('path')
const fs = require('fs')
const config = require('../config')

// Crear directorio de logs si no existe
const logDir = path.dirname(config.logging.file)
if (!fs.existsSync(logDir)) {
  fs.mkdirSync(logDir, { recursive: true })
}

/**
 * Formato personalizado para logs
 */
const customFormat = winston.format.combine(
  winston.format.timestamp({
    format: 'YYYY-MM-DD HH:mm:ss'
  }),
  winston.format.errors({ stack: true }),
  winston.format.json(),
  winston.format.printf(({ timestamp, level, message, ...meta }) => {
    let logMessage = `${timestamp} [${level.toUpperCase()}]: ${message}`
    
    // Agregar metadatos si existen
    if (Object.keys(meta).length > 0) {
      logMessage += ` | ${JSON.stringify(meta)}`
    }
    
    return logMessage
  })
)

/**
 * Formato para desarrollo (más legible)
 */
const developmentFormat = winston.format.combine(
  winston.format.colorize(),
  winston.format.timestamp({
    format: 'HH:mm:ss'
  }),
  winston.format.printf(({ timestamp, level, message, ...meta }) => {
    let logMessage = `${timestamp} ${level}: ${message}`
    
    if (Object.keys(meta).length > 0) {
      logMessage += `\n${JSON.stringify(meta, null, 2)}`
    }
    
    return logMessage
  })
)

/**
 * Configuración de transports
 */
const transports = []

// Console transport (siempre activo)
transports.push(
  new winston.transports.Console({
    level: config.logging.level,
    format: config.server.isDevelopment ? developmentFormat : customFormat,
    handleExceptions: true,
    handleRejections: true
  })
)

// File transport (solo en producción o si se especifica)
if (config.server.isProduction || process.env.LOG_TO_FILE === 'true') {
  transports.push(
    new winston.transports.File({
      filename: config.logging.file,
      level: config.logging.level,
      format: customFormat,
      maxsize: 20 * 1024 * 1024, // 20MB
      maxFiles: 5,
      tailable: true,
      handleExceptions: true,
      handleRejections: true
    })
  )

  // Transport separado para errores
  transports.push(
    new winston.transports.File({
      filename: path.join(logDir, 'error.log'),
      level: 'error',
      format: customFormat,
      maxsize: 20 * 1024 * 1024,
      maxFiles: 5,
      tailable: true
    })
  )
}

/**
 * Crear instancia de logger
 */
const logger = winston.createLogger({
  level: config.logging.level,
  format: customFormat,
  transports,
  exitOnError: false,
  silent: config.server.isTest
})

/**
 * Métodos de logging extendidos
 */

/**
 * Log de eventos de seguridad
 */
logger.security = (message, meta = {}) => {
  logger.warn(`[SECURITY] ${message}`, {
    ...meta,
    category: 'security',
    timestamp: new Date().toISOString()
  })
}

/**
 * Log de eventos de auditoría
 */
logger.audit = (message, meta = {}) => {
  logger.info(`[AUDIT] ${message}`, {
    ...meta,
    category: 'audit',
    timestamp: new Date().toISOString()
  })
}

/**
 * Log de eventos de negocio
 */
logger.business = (message, meta = {}) => {
  logger.info(`[BUSINESS] ${message}`, {
    ...meta,
    category: 'business',
    timestamp: new Date().toISOString()
  })
}

/**
 * Log de rendimiento
 */
logger.performance = (message, meta = {}) => {
  logger.debug(`[PERFORMANCE] ${message}`, {
    ...meta,
    category: 'performance',
    timestamp: new Date().toISOString()
  })
}

/**
 * Log de sincronización
 */
logger.sync = (message, meta = {}) => {
  logger.info(`[SYNC] ${message}`, {
    ...meta,
    category: 'sync',
    timestamp: new Date().toISOString()
  })
}

/**
 * Log de DTE (Documentos Tributarios Electrónicos)
 */
logger.dte = (message, meta = {}) => {
  logger.info(`[DTE] ${message}`, {
    ...meta,
    category: 'dte',
    timestamp: new Date().toISOString()
  })
}

/**
 * Log de pagos
 */
logger.payment = (message, meta = {}) => {
  logger.info(`[PAYMENT] ${message}`, {
    ...meta,
    category: 'payment',
    timestamp: new Date().toISOString()
  })
}

/**
 * Log de fidelización
 */
logger.loyalty = (message, meta = {}) => {
  logger.info(`[LOYALTY] ${message}`, {
    ...meta,
    category: 'loyalty',
    timestamp: new Date().toISOString()
  })
}

/**
 * Middleware para logging de requests HTTP
 */
logger.httpMiddleware = (request, reply, done) => {
  const start = Date.now()
  
  reply.addHook('onSend', (request, reply, payload, done) => {
    const duration = Date.now() - start
    const logData = {
      method: request.method,
      url: request.url,
      statusCode: reply.statusCode,
      duration: `${duration}ms`,
      userAgent: request.headers['user-agent'],
      ip: request.ip,
      userId: request.user?.id,
      sucursalId: request.user?.sucursal_id
    }

    if (reply.statusCode >= 400) {
      logger.warn('HTTP Request Error', logData)
    } else {
      logger.info('HTTP Request', logData)
    }
    
    done()
  })
  
  done()
}

/**
 * Función para logging de errores no capturados
 */
logger.handleUncaughtExceptions = () => {
  process.on('uncaughtException', (error) => {
    logger.error('Uncaught Exception:', error)
    process.exit(1)
  })

  process.on('unhandledRejection', (reason, promise) => {
    logger.error('Unhandled Rejection at:', { promise, reason })
    process.exit(1)
  })
}

/**
 * Función para crear un logger hijo con contexto específico
 */
logger.child = (meta = {}) => {
  return {
    debug: (message, additionalMeta = {}) => logger.debug(message, { ...meta, ...additionalMeta }),
    info: (message, additionalMeta = {}) => logger.info(message, { ...meta, ...additionalMeta }),
    warn: (message, additionalMeta = {}) => logger.warn(message, { ...meta, ...additionalMeta }),
    error: (message, additionalMeta = {}) => logger.error(message, { ...meta, ...additionalMeta }),
    security: (message, additionalMeta = {}) => logger.security(message, { ...meta, ...additionalMeta }),
    audit: (message, additionalMeta = {}) => logger.audit(message, { ...meta, ...additionalMeta }),
    business: (message, additionalMeta = {}) => logger.business(message, { ...meta, ...additionalMeta })
  }
}

/**
 * Función para obtener estadísticas de logging
 */
logger.getStats = () => {
  return {
    level: logger.level,
    transports: logger.transports.length,
    logFile: config.logging.file,
    logDir: logDir
  }
}

module.exports = logger

