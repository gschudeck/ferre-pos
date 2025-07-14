/**
 * Configuración Principal - Sistema Ferre-POS API
 * 
 * Centraliza toda la configuración del sistema basada en variables de entorno
 * con valores por defecto seguros y validaciones.
 */

require('dotenv').config()

const config = {
  // Configuración del servidor
  server: {
    env: process.env.NODE_ENV || 'development',
    port: parseInt(process.env.PORT) || 3000,
    host: process.env.HOST || '0.0.0.0',
    isDevelopment: process.env.NODE_ENV === 'development',
    isProduction: process.env.NODE_ENV === 'production',
    isTest: process.env.NODE_ENV === 'test'
  },

  // Configuración de base de datos
  database: {
    host: process.env.DB_HOST || 'localhost',
    port: parseInt(process.env.DB_PORT) || 5432,
    database: process.env.DB_NAME || 'ferre_pos',
    user: process.env.DB_USER || 'ferre_pos_app',
    password: process.env.DB_PASSWORD || '',
    ssl: process.env.DB_SSL === 'true',
    pool: {
      min: parseInt(process.env.DB_POOL_MIN) || 2,
      max: parseInt(process.env.DB_POOL_MAX) || 10,
      idleTimeoutMillis: 30000,
      connectionTimeoutMillis: 2000
    },
    // Configuración para consultas
    query: {
      timeout: 30000
    }
  },

  // Configuración de JWT y seguridad
  auth: {
    jwtSecret: process.env.JWT_SECRET || 'fallback_secret_key_change_in_production',
    jwtExpiresIn: process.env.JWT_EXPIRES_IN || '8h',
    bcryptRounds: parseInt(process.env.BCRYPT_ROUNDS) || 12,
    sessionTimeout: 8 * 60 * 60 * 1000, // 8 horas en milisegundos
    maxLoginAttempts: 5,
    lockoutDuration: 30 * 60 * 1000 // 30 minutos en milisegundos
  },

  // Configuración de rate limiting
  rateLimit: {
    max: parseInt(process.env.RATE_LIMIT_MAX) || 100,
    timeWindow: parseInt(process.env.RATE_LIMIT_WINDOW) || 60000, // 1 minuto
    skipSuccessfulRequests: false,
    skipFailedRequests: false
  },

  // Configuración de CORS
  cors: {
    origin: process.env.CORS_ORIGIN === '*' ? true : process.env.CORS_ORIGIN?.split(',') || ['http://localhost:3000'],
    credentials: process.env.CORS_CREDENTIALS === 'true',
    methods: ['GET', 'POST', 'PUT', 'PATCH', 'DELETE', 'OPTIONS'],
    allowedHeaders: ['Content-Type', 'Authorization', 'X-Requested-With']
  },

  // Configuración de logging
  logging: {
    level: process.env.LOG_LEVEL || 'info',
    file: process.env.LOG_FILE || 'logs/ferre-pos-api.log',
    maxSize: '20m',
    maxFiles: '14d',
    format: 'combined'
  },

  // Configuración de DTE (Documentos Tributarios Electrónicos)
  dte: {
    providerUrl: process.env.DTE_PROVIDER_URL || '',
    providerToken: process.env.DTE_PROVIDER_TOKEN || '',
    environment: process.env.DTE_ENVIRONMENT || 'certificacion',
    timeout: 30000,
    retries: 3,
    retryDelay: 1000
  },

  // Configuración de medios de pago
  payments: {
    transbank: {
      apiUrl: process.env.TRANSBANK_API_URL || 'https://webpay3gint.transbank.cl',
      commerceCode: process.env.TRANSBANK_COMMERCE_CODE || '',
      apiKey: process.env.TRANSBANK_API_KEY || '',
      timeout: 30000
    },
    mercadopago: {
      accessToken: process.env.MERCADOPAGO_ACCESS_TOKEN || '',
      publicKey: process.env.MERCADOPAGO_PUBLIC_KEY || '',
      timeout: 30000
    }
  },

  // Configuración de fidelización
  fidelizacion: {
    puntosPorPeso: parseInt(process.env.FIDELIZACION_PUNTOS_POR_PESO) || 1,
    minimoCanje: parseInt(process.env.FIDELIZACION_MINIMO_CANJE) || 100,
    expiracionDias: parseInt(process.env.FIDELIZACION_EXPIRACION_DIAS) || 365,
    niveles: {
      bronce: { minPuntos: 0, multiplicador: 1.0 },
      plata: { minPuntos: 5000, multiplicador: 1.2 },
      oro: { minPuntos: 20000, multiplicador: 1.5 },
      platino: { minPuntos: 50000, multiplicador: 2.0 }
    }
  },

  // Configuración de sincronización
  sync: {
    intervalMinutes: parseInt(process.env.SYNC_INTERVAL_MINUTES) || 5,
    batchSize: parseInt(process.env.SYNC_BATCH_SIZE) || 100,
    timeout: 60000,
    retries: 3
  },

  // Configuración de monitoreo
  monitoring: {
    enableMetrics: process.env.ENABLE_METRICS === 'true',
    metricsPort: parseInt(process.env.METRICS_PORT) || 9090,
    healthCheckInterval: parseInt(process.env.HEALTH_CHECK_INTERVAL) || 30000
  },

  // Configuración de archivos y uploads
  uploads: {
    maxSize: parseInt(process.env.UPLOAD_MAX_SIZE) || 10 * 1024 * 1024, // 10MB
    allowedTypes: process.env.UPLOAD_ALLOWED_TYPES?.split(',') || [
      'image/jpeg',
      'image/png',
      'application/pdf'
    ],
    destination: 'uploads/',
    tempDir: 'temp/'
  },

  // Configuración de cache
  cache: {
    ttl: parseInt(process.env.CACHE_TTL) || 300, // 5 minutos
    maxKeys: parseInt(process.env.CACHE_MAX_KEYS) || 1000,
    checkPeriod: 60 // 1 minuto
  },

  // Configuración de Swagger/OpenAPI
  swagger: {
    routePrefix: '/docs',
    exposeRoute: true,
    swagger: {
      info: {
        title: 'Ferre-POS API',
        description: 'API REST para Sistema de Punto de Venta para Ferreterías',
        version: '1.0.0',
        contact: {
          name: 'Soporte Técnico',
          email: 'soporte@ferre-pos.cl'
        },
        license: {
          name: 'MIT',
          url: 'https://opensource.org/licenses/MIT'
        }
      },
      host: 'localhost:3000',
      schemes: ['http', 'https'],
      consumes: ['application/json'],
      produces: ['application/json'],
      securityDefinitions: {
        Bearer: {
          type: 'apiKey',
          name: 'Authorization',
          in: 'header',
          description: 'JWT token. Formato: Bearer {token}'
        }
      }
    }
  },

  // Configuración de validaciones
  validation: {
    abortEarly: false,
    allowUnknown: false,
    stripUnknown: true
  },

  // Configuración de paginación
  pagination: {
    defaultLimit: 20,
    maxLimit: 100,
    defaultOffset: 0
  },

  // Configuración de stock
  stock: {
    alertaMinimo: 5,
    reservaTimeout: 30 * 60 * 1000, // 30 minutos
    movimientosRetencionDias: 90
  },

  // Configuración de reportes
  reportes: {
    retencionDias: 90,
    formatosPermitidos: ['pdf', 'excel', 'csv'],
    maxRegistros: 10000
  }
}

/**
 * Valida que las configuraciones críticas estén presentes
 */
function validateConfig() {
  const requiredEnvVars = [
    'DB_HOST',
    'DB_NAME',
    'DB_USER',
    'DB_PASSWORD',
    'JWT_SECRET'
  ]

  const missing = requiredEnvVars.filter(envVar => !process.env[envVar])
  
  if (missing.length > 0) {
    throw new Error(`Variables de entorno requeridas faltantes: ${missing.join(', ')}`)
  }

  // Validar JWT secret en producción
  if (config.server.isProduction && config.auth.jwtSecret === 'fallback_secret_key_change_in_production') {
    throw new Error('JWT_SECRET debe ser configurado en producción')
  }

  // Validar configuración de base de datos
  if (!config.database.password && config.server.isProduction) {
    throw new Error('DB_PASSWORD es requerido en producción')
  }
}

/**
 * Obtiene la configuración de conexión a la base de datos
 */
function getDatabaseConfig() {
  return {
    host: config.database.host,
    port: config.database.port,
    database: config.database.database,
    user: config.database.user,
    password: config.database.password,
    ssl: config.database.ssl,
    ...config.database.pool
  }
}

/**
 * Obtiene la configuración de Fastify
 */
function getFastifyConfig() {
  return {
    logger: {
      level: config.logging.level,
      file: config.logging.file
    },
    trustProxy: true,
    bodyLimit: config.uploads.maxSize,
    keepAliveTimeout: 30000,
    connectionTimeout: 30000
  }
}

// Validar configuración al cargar el módulo
if (process.env.NODE_ENV !== 'test') {
  validateConfig()
}

module.exports = {
  ...config,
  validateConfig,
  getDatabaseConfig,
  getFastifyConfig
}

