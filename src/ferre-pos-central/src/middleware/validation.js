/**
 * Middleware de Validación - Sistema Ferre-POS
 * 
 * Proporciona validación de datos usando Joi para asegurar
 * que todos los datos de entrada sean válidos y seguros.
 */

const Joi = require('joi')
const logger = require('../utils/logger')

/**
 * Esquemas de validación comunes
 */
const commonSchemas = {
  // Validación de UUID
  uuid: Joi.string().uuid().required(),
  
  // Validación de RUT chileno
  rut: Joi.string().pattern(/^\d{7,8}-[\dkK]$/).required()
    .messages({
      'string.pattern.base': 'RUT debe tener formato válido (ej: 12345678-9)'
    }),
  
  // Validación de email
  email: Joi.string().email().max(255),
  
  // Validación de teléfono
  telefono: Joi.string().pattern(/^\+?[\d\s\-\(\)]{8,15}$/),
  
  // Validación de fecha
  fecha: Joi.date().iso(),
  
  // Validación de paginación
  pagination: {
    page: Joi.number().integer().min(1).default(1),
    limit: Joi.number().integer().min(1).max(100).default(20)
  },
  
  // Validación de ordenamiento
  ordering: {
    orderBy: Joi.string().default('id'),
    orderDirection: Joi.string().valid('ASC', 'DESC').default('ASC')
  }
}

/**
 * Esquemas de validación para autenticación
 */
const authSchemas = {
  login: Joi.object({
    rut: commonSchemas.rut,
    password: Joi.string().min(6).max(100).required()
  }),
  
  changePassword: Joi.object({
    currentPassword: Joi.string().required(),
    newPassword: Joi.string().min(6).max(100).required(),
    confirmPassword: Joi.string().valid(Joi.ref('newPassword')).required()
      .messages({
        'any.only': 'Las contraseñas no coinciden'
      })
  }),
  
  resetPassword: Joi.object({
    userId: commonSchemas.uuid,
    newPassword: Joi.string().min(6).max(100).required()
  })
}

/**
 * Esquemas de validación para usuarios
 */
const usuarioSchemas = {
  create: Joi.object({
    rut: commonSchemas.rut,
    nombre: Joi.string().min(2).max(100).required(),
    apellido: Joi.string().max(100),
    email: commonSchemas.email,
    telefono: commonSchemas.telefono,
    rol: Joi.string().valid('admin', 'gerente', 'cajero', 'vendedor').required(),
    sucursal_id: commonSchemas.uuid.when('rol', {
      is: Joi.valid('cajero', 'vendedor'),
      then: Joi.required(),
      otherwise: Joi.optional()
    }),
    password: Joi.string().min(6).max(100).required()
  }),
  
  update: Joi.object({
    nombre: Joi.string().min(2).max(100),
    apellido: Joi.string().max(100),
    email: commonSchemas.email,
    telefono: commonSchemas.telefono,
    rol: Joi.string().valid('admin', 'gerente', 'cajero', 'vendedor'),
    sucursal_id: commonSchemas.uuid,
    activo: Joi.boolean()
  }).min(1),
  
  search: Joi.object({
    rut: Joi.string(),
    nombre: Joi.string(),
    email: Joi.string(),
    rol: Joi.string().valid('admin', 'gerente', 'cajero', 'vendedor'),
    sucursalId: commonSchemas.uuid,
    activo: Joi.boolean(),
    ...commonSchemas.pagination,
    ...commonSchemas.ordering
  })
}

/**
 * Esquemas de validación para productos
 */
const productoSchemas = {
  create: Joi.object({
    codigo_interno: Joi.string().max(50).required(),
    codigo_barra: Joi.string().max(50).required(),
    descripcion: Joi.string().min(3).max(500).required(),
    descripcion_corta: Joi.string().max(100),
    categoria_id: commonSchemas.uuid,
    marca: Joi.string().max(100),
    modelo: Joi.string().max(100),
    precio_unitario: Joi.number().positive().precision(2).required(),
    precio_costo: Joi.number().min(0).precision(2),
    unidad_medida: Joi.string().max(10).required(),
    peso: Joi.number().min(0).precision(3),
    stock_minimo: Joi.number().integer().min(0).default(0),
    stock_maximo: Joi.number().integer().min(0)
  }).custom((value, helpers) => {
    if (value.stock_maximo && value.stock_maximo < value.stock_minimo) {
      return helpers.error('custom.stockMaximoMenor')
    }
    return value
  }).messages({
    'custom.stockMaximoMenor': 'El stock máximo no puede ser menor al stock mínimo'
  }),
  
  update: Joi.object({
    descripcion: Joi.string().min(3).max(500),
    descripcion_corta: Joi.string().max(100),
    categoria_id: commonSchemas.uuid,
    marca: Joi.string().max(100),
    modelo: Joi.string().max(100),
    precio_unitario: Joi.number().positive().precision(2),
    precio_costo: Joi.number().min(0).precision(2),
    unidad_medida: Joi.string().max(10),
    peso: Joi.number().min(0).precision(3),
    stock_minimo: Joi.number().integer().min(0),
    stock_maximo: Joi.number().integer().min(0)
  }).min(1),
  
  search: Joi.object({
    q: Joi.string().max(100),
    codigo: Joi.string().max(50),
    descripcion: Joi.string().max(100),
    marca: Joi.string().max(100),
    categoria: commonSchemas.uuid,
    precioMin: Joi.number().min(0),
    precioMax: Joi.number().min(0),
    conStock: Joi.boolean(),
    sucursalId: commonSchemas.uuid,
    ...commonSchemas.pagination,
    ...commonSchemas.ordering
  }).custom((value, helpers) => {
    if (value.precioMin && value.precioMax && value.precioMin > value.precioMax) {
      return helpers.error('custom.precioMinMayor')
    }
    return value
  }).messages({
    'custom.precioMinMayor': 'El precio mínimo no puede ser mayor al precio máximo'
  }),
  
  updatePrecio: Joi.object({
    precio: Joi.number().positive().precision(2).required()
  }),
  
  addCodigoBarra: Joi.object({
    codigoBarra: Joi.string().max(50).required(),
    descripcion: Joi.string().max(100)
  }),
  
  validateStock: Joi.object({
    productos: Joi.array().items(
      Joi.object({
        producto_id: commonSchemas.uuid,
        cantidad: Joi.number().positive().required()
      })
    ).min(1).required()
  })
}

/**
 * Esquemas de validación para ventas
 */
const ventaSchemas = {
  create: Joi.object({
    venta: Joi.object({
      sucursal_id: commonSchemas.uuid,
      terminal_id: commonSchemas.uuid,
      cajero_id: commonSchemas.uuid,
      vendedor_id: commonSchemas.uuid,
      cliente_rut: commonSchemas.rut.optional(),
      cliente_nombre: Joi.string().max(200),
      nota_venta_id: commonSchemas.uuid,
      tipo_documento: Joi.string().valid('boleta', 'factura').required(),
      subtotal: Joi.number().min(0).precision(2).required(),
      descuento_total: Joi.number().min(0).precision(2).default(0),
      impuesto_total: Joi.number().min(0).precision(2).default(0),
      total: Joi.number().positive().precision(2).required()
    }).required(),
    
    detalles: Joi.array().items(
      Joi.object({
        producto_id: commonSchemas.uuid,
        cantidad: Joi.number().positive().required(),
        precio_unitario: Joi.number().positive().precision(2).required(),
        descuento_unitario: Joi.number().min(0).precision(2).default(0),
        precio_final: Joi.number().positive().precision(2).required(),
        total_item: Joi.number().positive().precision(2).required(),
        numero_serie: Joi.string().max(100),
        lote: Joi.string().max(50),
        fecha_vencimiento: commonSchemas.fecha
      })
    ).min(1).required(),
    
    mediosPago: Joi.array().items(
      Joi.object({
        medio_pago: Joi.string().valid('efectivo', 'tarjeta_debito', 'tarjeta_credito', 'transferencia', 'cheque').required(),
        monto: Joi.number().positive().precision(2).required(),
        referencia_transaccion: Joi.string().max(100),
        codigo_autorizacion: Joi.string().max(50),
        datos_transaccion: Joi.object()
      })
    ).min(1).required(),
    
    aplicarFidelizacion: Joi.boolean().default(false)
  }),
  
  anular: Joi.object({
    motivo: Joi.string().min(10).max(500).required()
  }),
  
  search: Joi.object({
    sucursalId: commonSchemas.uuid,
    fechaInicio: commonSchemas.fecha,
    fechaFin: commonSchemas.fecha,
    cajeroId: commonSchemas.uuid,
    estado: Joi.string().valid('finalizada', 'anulada'),
    clienteRut: commonSchemas.rut.optional(),
    ...commonSchemas.pagination,
    ...commonSchemas.ordering
  })
}

/**
 * Esquemas de validación para stock
 */
const stockSchemas = {
  movimiento: Joi.object({
    producto_id: commonSchemas.uuid,
    sucursal_id: commonSchemas.uuid,
    tipo_movimiento: Joi.string().valid('entrada', 'salida', 'transferencia_entrada', 'transferencia_salida', 'ajuste', 'devolucion').required(),
    cantidad: Joi.number().positive().required(),
    costo_unitario: Joi.number().min(0).precision(2),
    documento_referencia: Joi.string().max(100),
    observaciones: Joi.string().max(500)
  }),
  
  transferencia: Joi.object({
    producto_id: commonSchemas.uuid,
    sucursal_origen_id: commonSchemas.uuid,
    sucursal_destino_id: commonSchemas.uuid,
    cantidad: Joi.number().positive().required(),
    observaciones: Joi.string().max(500)
  }).custom((value, helpers) => {
    if (value.sucursal_origen_id === value.sucursal_destino_id) {
      return helpers.error('custom.sucursalesIguales')
    }
    return value
  }).messages({
    'custom.sucursalesIguales': 'La sucursal de origen y destino no pueden ser iguales'
  }),
  
  ajuste: Joi.object({
    producto_id: commonSchemas.uuid,
    sucursal_id: commonSchemas.uuid,
    cantidad_fisica: Joi.number().integer().min(0).required(),
    observaciones: Joi.string().max(500)
  }),
  
  reserva: Joi.object({
    producto_id: commonSchemas.uuid,
    sucursal_id: commonSchemas.uuid,
    cantidad: Joi.number().positive().required(),
    referencia: Joi.string().max(100)
  })
}

/**
 * Middleware de validación genérico
 */
function validate(schema, property = 'body') {
  return async (request, reply) => {
    try {
      const dataToValidate = request[property]
      const { error, value } = schema.validate(dataToValidate, {
        abortEarly: false,
        allowUnknown: false,
        stripUnknown: true
      })

      if (error) {
        const validationErrors = error.details.map(detail => ({
          field: detail.path.join('.'),
          message: detail.message,
          value: detail.context?.value
        }))

        logger.warn('Error de validación', {
          endpoint: request.url,
          method: request.method,
          errors: validationErrors,
          userId: request.user?.id
        })

        return reply.code(400).send({
          code: 'VALIDATION_ERROR',
          error: 'Bad Request',
          message: 'Error de validación en los datos enviados',
          details: validationErrors
        })
      }

      // Reemplazar los datos originales con los validados y sanitizados
      request[property] = value
    } catch (error) {
      logger.error('Error en middleware de validación:', error)
      return reply.code(500).send({
        code: 'INTERNAL_ERROR',
        error: 'Internal Server Error',
        message: 'Error interno del servidor'
      })
    }
  }
}

/**
 * Middleware de validación para parámetros de URL
 */
function validateParams(schema) {
  return validate(schema, 'params')
}

/**
 * Middleware de validación para query parameters
 */
function validateQuery(schema) {
  return validate(schema, 'query')
}

/**
 * Middleware de validación para body
 */
function validateBody(schema) {
  return validate(schema, 'body')
}

/**
 * Validador de UUID para parámetros
 */
const validateUuidParam = validateParams(Joi.object({
  id: commonSchemas.uuid
}))

/**
 * Validador de código para parámetros
 */
const validateCodigoParam = validateParams(Joi.object({
  codigo: Joi.string().max(50).required()
}))

/**
 * Middleware de sanitización de datos
 */
function sanitizeData(request, reply, done) {
  // Sanitizar strings en body
  if (request.body && typeof request.body === 'object') {
    sanitizeObject(request.body)
  }

  // Sanitizar strings en query
  if (request.query && typeof request.query === 'object') {
    sanitizeObject(request.query)
  }

  done()
}

/**
 * Función auxiliar para sanitizar objetos
 */
function sanitizeObject(obj) {
  for (const key in obj) {
    if (typeof obj[key] === 'string') {
      // Remover espacios al inicio y final
      obj[key] = obj[key].trim()
      
      // Convertir strings vacíos a null
      if (obj[key] === '') {
        obj[key] = null
      }
    } else if (typeof obj[key] === 'object' && obj[key] !== null) {
      sanitizeObject(obj[key])
    }
  }
}

module.exports = {
  // Esquemas
  commonSchemas,
  authSchemas,
  usuarioSchemas,
  productoSchemas,
  ventaSchemas,
  stockSchemas,
  
  // Middlewares
  validate,
  validateParams,
  validateQuery,
  validateBody,
  validateUuidParam,
  validateCodigoParam,
  sanitizeData
}

