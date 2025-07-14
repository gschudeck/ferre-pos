/**
 * Servidor Principal - Sistema Ferre-POS API
 * 
 * Configuración y arranque del servidor Fastify con todos los
 * plugins, middleware y rutas necesarias.
 */

const fastify = require('fastify')
const config = require('./config')
const database = require('./config/database')
const logger = require('./utils/logger')

/**
 * Crea y configura la instancia de Fastify
 */
async function createServer() {
  // Crear instancia de Fastify
  const server = fastify({
    logger: false, // Usamos nuestro logger personalizado
    trustProxy: true,
    bodyLimit: config.uploads.maxSize,
    keepAliveTimeout: 30000,
    connectionTimeout: 30000,
    ignoreTrailingSlash: true,
    ignoreDuplicateSlashes: true
  })

  // Registrar plugins de seguridad
  await server.register(require('@fastify/helmet'), {
    contentSecurityPolicy: {
      directives: {
        defaultSrc: ["'self'"],
        styleSrc: ["'self'", "'unsafe-inline'"],
        scriptSrc: ["'self'"],
        imgSrc: ["'self'", "data:", "https:"]
      }
    }
  })

  // Registrar CORS
  await server.register(require('@fastify/cors'), config.cors)

  // Registrar rate limiting
  await server.register(require('@fastify/rate-limit'), {
    max: config.rateLimit.max,
    timeWindow: config.rateLimit.timeWindow,
    skipSuccessfulRequests: config.rateLimit.skipSuccessfulRequests,
    skipFailedRequests: config.rateLimit.skipFailedRequests,
    keyGenerator: (request) => {
      return request.user?.id || request.ip
    },
    errorResponseBuilder: (request, context) => {
      return {
        code: 'RATE_LIMIT_EXCEEDED',
        error: 'Too Many Requests',
        message: `Rate limit exceeded, retry in ${Math.round(context.ttl / 1000)} seconds`,
        statusCode: 429,
        ttl: context.ttl
      }
    }
  })

  // Registrar JWT
  await server.register(require('@fastify/jwt'), {
    secret: config.auth.jwtSecret,
    sign: {
      expiresIn: config.auth.jwtExpiresIn
    },
    verify: {
      maxAge: config.auth.jwtExpiresIn
    }
  })

  // Registrar multipart para uploads
  await server.register(require('@fastify/multipart'), {
    limits: {
      fileSize: config.uploads.maxSize,
      files: 5
    }
  })

  // Registrar Swagger para documentación
  await server.register(require('@fastify/swagger'), config.swagger.swagger)
  await server.register(require('@fastify/swagger-ui'), {
    routePrefix: config.swagger.routePrefix,
    exposeRoute: config.swagger.exposeRoute,
    uiConfig: {
      docExpansion: 'list',
      deepLinking: false
    }
  })

  // Middleware de logging personalizado
  server.addHook('onRequest', logger.httpMiddleware)

  // Middleware de autenticación
  server.decorate('authenticate', async function(request, reply) {
    try {
      await request.jwtVerify()
      
      // Verificar que el usuario existe y está activo
      const userResult = await database.query(
        'SELECT id, rut, nombre, apellido, rol, sucursal_id, activo FROM usuarios WHERE id = $1',
        [request.user.id]
      )
      
      if (!userResult.rows.length || !userResult.rows[0].activo) {
        throw new Error('Usuario no válido o inactivo')
      }
      
      request.user = userResult.rows[0]
      
      // Actualizar último acceso
      await database.query(
        'UPDATE usuarios SET ultimo_acceso = NOW() WHERE id = $1',
        [request.user.id]
      )
      
    } catch (error) {
      logger.security('Intento de acceso no autorizado', {
        ip: request.ip,
        userAgent: request.headers['user-agent'],
        error: error.message
      })
      reply.code(401).send({
        code: 'UNAUTHORIZED',
        error: 'Unauthorized',
        message: 'Token de autenticación inválido o expirado'
      })
    }
  })

  // Middleware de autorización por roles
  server.decorate('authorize', (allowedRoles = []) => {
    return async function(request, reply) {
      if (!request.user) {
        return reply.code(401).send({
          code: 'UNAUTHORIZED',
          error: 'Unauthorized',
          message: 'Autenticación requerida'
        })
      }

      if (allowedRoles.length > 0 && !allowedRoles.includes(request.user.rol)) {
        logger.security('Acceso denegado por rol insuficiente', {
          userId: request.user.id,
          userRole: request.user.rol,
          requiredRoles: allowedRoles,
          endpoint: request.url
        })
        
        return reply.code(403).send({
          code: 'FORBIDDEN',
          error: 'Forbidden',
          message: 'No tiene permisos para acceder a este recurso'
        })
      }
    }
  })

  // Decorador para validar sucursal
  server.decorate('validateSucursal', async function(request, reply) {
    const sucursalId = request.params.sucursalId || request.body?.sucursal_id
    
    if (sucursalId && request.user.rol !== 'admin') {
      if (request.user.sucursal_id !== sucursalId) {
        return reply.code(403).send({
          code: 'FORBIDDEN',
          error: 'Forbidden',
          message: 'No tiene acceso a esta sucursal'
        })
      }
    }
  })

  // Hook para manejo de errores
  server.setErrorHandler(async (error, request, reply) => {
    logger.error('Error en request', {
      error: error.message,
      stack: error.stack,
      method: request.method,
      url: request.url,
      userId: request.user?.id,
      ip: request.ip
    })

    // Errores de validación
    if (error.validation) {
      return reply.code(400).send({
        code: 'VALIDATION_ERROR',
        error: 'Bad Request',
        message: 'Error de validación en los datos enviados',
        details: error.validation
      })
    }

    // Errores de base de datos
    if (error.code && error.code.startsWith('23')) {
      return reply.code(409).send({
        code: 'DATABASE_CONSTRAINT',
        error: 'Conflict',
        message: 'Violación de restricción en base de datos'
      })
    }

    // Error genérico
    const statusCode = error.statusCode || 500
    reply.code(statusCode).send({
      code: 'INTERNAL_ERROR',
      error: statusCode === 500 ? 'Internal Server Error' : error.name,
      message: config.server.isDevelopment ? error.message : 'Error interno del servidor'
    })
  })

  // Hook para respuestas no encontradas
  server.setNotFoundHandler(async (request, reply) => {
    reply.code(404).send({
      code: 'NOT_FOUND',
      error: 'Not Found',
      message: `Ruta ${request.method} ${request.url} no encontrada`
    })
  })

  // Registrar rutas principales
  await server.register(require('./routes/auth'), { prefix: '/api/auth' })
  await server.register(require('./routes/usuarios'), { prefix: '/api/usuarios' })
  await server.register(require('./routes/sucursales'), { prefix: '/api/sucursales' })
  await server.register(require('./routes/productos'), { prefix: '/api/productos' })
  await server.register(require('./routes/categorias'), { prefix: '/api/categorias' })
  await server.register(require('./routes/stock'), { prefix: '/api/stock' })
  await server.register(require('./routes/ventas'), { prefix: '/api/ventas' })
  await server.register(require('./routes/notas-venta'), { prefix: '/api/notas-venta' })
  await server.register(require('./routes/fidelizacion'), { prefix: '/api/fidelizacion' })
  await server.register(require('./routes/dte'), { prefix: '/api/dte' })
  await server.register(require('./routes/despachos'), { prefix: '/api/despachos' })
  await server.register(require('./routes/reportes'), { prefix: '/api/reportes' })
  await server.register(require('./routes/sistema'), { prefix: '/api/sistema' })

  // Ruta de health check
  server.get('/health', async (request, reply) => {
    const dbHealth = await database.healthCheck()
    const health = {
      status: dbHealth.status === 'healthy' ? 'ok' : 'error',
      timestamp: new Date().toISOString(),
      version: '1.0.0',
      environment: config.server.env,
      database: dbHealth,
      uptime: process.uptime(),
      memory: process.memoryUsage()
    }

    const statusCode = health.status === 'ok' ? 200 : 503
    reply.code(statusCode).send(health)
  })

  // Ruta de información del sistema
  server.get('/info', async (request, reply) => {
    reply.send({
      name: 'Ferre-POS API',
      version: '1.0.0',
      description: 'API REST para Sistema de Punto de Venta para Ferreterías',
      environment: config.server.env,
      documentation: '/docs',
      health: '/health'
    })
  })

  return server
}

/**
 * Inicia el servidor
 */
async function startServer() {
  try {
    // Conectar a la base de datos
    await database.connect()
    
    // Crear servidor
    const server = await createServer()
    
    // Iniciar servidor
    await server.listen({
      port: config.server.port,
      host: config.server.host
    })
    
    logger.info('Servidor Ferre-POS API iniciado exitosamente', {
      port: config.server.port,
      host: config.server.host,
      environment: config.server.env,
      documentation: `http://${config.server.host}:${config.server.port}/docs`,
      health: `http://${config.server.host}:${config.server.port}/health`
    })

    // Manejar cierre graceful
    const gracefulShutdown = async (signal) => {
      logger.info(`Recibida señal ${signal}, cerrando servidor...`)
      
      try {
        await server.close()
        await database.disconnect()
        logger.info('Servidor cerrado exitosamente')
        process.exit(0)
      } catch (error) {
        logger.error('Error al cerrar servidor:', error)
        process.exit(1)
      }
    }

    process.on('SIGTERM', () => gracefulShutdown('SIGTERM'))
    process.on('SIGINT', () => gracefulShutdown('SIGINT'))
    
    return server
    
  } catch (error) {
    logger.error('Error al iniciar servidor:', error)
    process.exit(1)
  }
}

// Manejar excepciones no capturadas
logger.handleUncaughtExceptions()

// Iniciar servidor si este archivo es ejecutado directamente
if (require.main === module) {
  startServer()
}

module.exports = { createServer, startServer }

