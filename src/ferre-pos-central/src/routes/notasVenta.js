/**
 * Rutas de Notas de Venta - Sistema Ferre-POS
 * 
 * Define todos los endpoints relacionados con las notas de venta,
 * incluyendo cotizaciones, reservas y conversión a ventas.
 */

const NotaVentaController = require('../controllers/NotaVentaController')
const { validateBody, validateQuery, validateUuidParam } = require('../middleware/validation')
const Joi = require('joi')

async function notasVentaRoutes(fastify, options) {
  // Esquemas de validación
  const notaVentaSchema = {
    type: 'object',
    properties: {
      id: { type: 'string' },
      numero_nota: { type: 'integer' },
      sucursal_id: { type: 'string' },
      vendedor_id: { type: 'string' },
      cliente_rut: { type: 'string' },
      cliente_nombre: { type: 'string' },
      cliente_telefono: { type: 'string' },
      cliente_email: { type: 'string' },
      tipo_nota: { type: 'string', enum: ['cotizacion', 'reserva'] },
      subtotal: { type: 'number' },
      descuento_total: { type: 'number' },
      impuesto_total: { type: 'number' },
      total: { type: 'number' },
      estado: { type: 'string', enum: ['activa', 'convertida', 'anulada', 'vencida'] },
      observaciones: { type: 'string' },
      fecha: { type: 'string', format: 'date-time' },
      fecha_vencimiento: { type: 'string', format: 'date-time' },
      fecha_conversion: { type: 'string', format: 'date-time' },
      fecha_anulacion: { type: 'string', format: 'date-time' }
    }
  }

  const detalleNotaSchema = {
    type: 'object',
    properties: {
      id: { type: 'string' },
      nota_venta_id: { type: 'string' },
      producto_id: { type: 'string' },
      cantidad: { type: 'number' },
      precio_unitario: { type: 'number' },
      descuento_unitario: { type: 'number' },
      precio_final: { type: 'number' },
      total_item: { type: 'number' },
      observaciones: { type: 'string' }
    }
  }

  const createNotaSchema = {
    description: 'Crea una nueva nota de venta',
    tags: ['Notas de Venta'],
    security: [{ Bearer: [] }],
    body: {
      type: 'object',
      required: ['nota', 'detalles'],
      properties: {
        nota: {
          type: 'object',
          required: ['tipo_nota', 'subtotal', 'total'],
          properties: {
            sucursal_id: {
              type: 'string',
              description: 'ID de la sucursal (opcional si el usuario tiene sucursal asignada)'
            },
            cliente_rut: {
              type: 'string',
              description: 'RUT del cliente',
              pattern: '^\\d{7,8}-[\\dkK]$'
            },
            cliente_nombre: {
              type: 'string',
              description: 'Nombre del cliente',
              maxLength: 200
            },
            cliente_telefono: {
              type: 'string',
              description: 'Teléfono del cliente',
              maxLength: 20
            },
            cliente_email: {
              type: 'string',
              description: 'Email del cliente',
              format: 'email'
            },
            tipo_nota: {
              type: 'string',
              enum: ['cotizacion', 'reserva'],
              description: 'Tipo de nota de venta'
            },
            subtotal: {
              type: 'number',
              description: 'Subtotal de la nota',
              minimum: 0
            },
            descuento_total: {
              type: 'number',
              description: 'Descuento total aplicado',
              minimum: 0,
              default: 0
            },
            impuesto_total: {
              type: 'number',
              description: 'Impuestos totales',
              minimum: 0,
              default: 0
            },
            total: {
              type: 'number',
              description: 'Total de la nota',
              minimum: 0.01
            },
            observaciones: {
              type: 'string',
              description: 'Observaciones adicionales',
              maxLength: 1000
            }
          }
        },
        detalles: {
          type: 'array',
          description: 'Detalles de productos en la nota',
          minItems: 1,
          items: {
            type: 'object',
            required: ['producto_id', 'cantidad', 'precio_unitario', 'precio_final', 'total_item'],
            properties: {
              producto_id: {
                type: 'string',
                description: 'ID del producto'
              },
              cantidad: {
                type: 'number',
                description: 'Cantidad del producto',
                minimum: 0.01
              },
              precio_unitario: {
                type: 'number',
                description: 'Precio unitario del producto',
                minimum: 0
              },
              descuento_unitario: {
                type: 'number',
                description: 'Descuento por unidad',
                minimum: 0,
                default: 0
              },
              precio_final: {
                type: 'number',
                description: 'Precio final por unidad',
                minimum: 0
              },
              total_item: {
                type: 'number',
                description: 'Total del item',
                minimum: 0
              },
              observaciones: {
                type: 'string',
                description: 'Observaciones del item',
                maxLength: 500
              }
            }
          }
        }
      }
    },
    response: {
      201: {
        type: 'object',
        properties: {
          success: { type: 'boolean' },
          message: { type: 'string' },
          data: {
            type: 'object',
            properties: {
              nota: notaVentaSchema,
              detalles: {
                type: 'array',
                items: detalleNotaSchema
              }
            }
          }
        }
      }
    }
  }

  const getNotasSchema = {
    description: 'Obtiene lista de notas de venta con filtros',
    tags: ['Notas de Venta'],
    security: [{ Bearer: [] }],
    querystring: {
      type: 'object',
      properties: {
        page: {
          type: 'integer',
          description: 'Número de página',
          minimum: 1,
          default: 1
        },
        limit: {
          type: 'integer',
          description: 'Cantidad de notas por página',
          minimum: 1,
          maximum: 100,
          default: 20
        },
        sucursalId: {
          type: 'string',
          description: 'Filtrar por sucursal (solo admin)'
        },
        vendedorId: {
          type: 'string',
          description: 'Filtrar por vendedor'
        },
        fechaInicio: {
          type: 'string',
          format: 'date',
          description: 'Fecha de inicio del período'
        },
        fechaFin: {
          type: 'string',
          format: 'date',
          description: 'Fecha de fin del período'
        },
        tipoNota: {
          type: 'string',
          enum: ['cotizacion', 'reserva'],
          description: 'Filtrar por tipo de nota'
        },
        estado: {
          type: 'string',
          enum: ['activa', 'convertida', 'anulada', 'vencida'],
          description: 'Filtrar por estado'
        },
        clienteRut: {
          type: 'string',
          description: 'Filtrar por RUT del cliente'
        }
      }
    },
    response: {
      200: {
        type: 'object',
        properties: {
          success: { type: 'boolean' },
          data: {
            type: 'array',
            items: notaVentaSchema
          },
          pagination: {
            type: 'object',
            properties: {
              page: { type: 'integer' },
              limit: { type: 'integer' },
              total: { type: 'integer' },
              totalPages: { type: 'integer' }
            }
          }
        }
      }
    }
  }

  const convertirAVentaSchema = {
    description: 'Convierte una nota de venta en venta real',
    tags: ['Notas de Venta'],
    security: [{ Bearer: [] }],
    params: {
      type: 'object',
      required: ['id'],
      properties: {
        id: {
          type: 'string',
          description: 'ID de la nota de venta'
        }
      }
    },
    body: {
      type: 'object',
      required: ['terminal_id', 'tipo_documento', 'mediosPago'],
      properties: {
        terminal_id: {
          type: 'string',
          description: 'ID del terminal donde se procesa la venta'
        },
        cajero_id: {
          type: 'string',
          description: 'ID del cajero (opcional, usa el usuario actual si no se especifica)'
        },
        tipo_documento: {
          type: 'string',
          enum: ['boleta', 'factura'],
          description: 'Tipo de documento a emitir'
        },
        mediosPago: {
          type: 'array',
          description: 'Medios de pago utilizados',
          minItems: 1,
          items: {
            type: 'object',
            required: ['medio_pago', 'monto'],
            properties: {
              medio_pago: {
                type: 'string',
                enum: ['efectivo', 'tarjeta_debito', 'tarjeta_credito', 'transferencia', 'cheque'],
                description: 'Tipo de medio de pago'
              },
              monto: {
                type: 'number',
                description: 'Monto pagado con este medio',
                minimum: 0.01
              },
              referencia_transaccion: {
                type: 'string',
                description: 'Referencia de la transacción'
              },
              codigo_autorizacion: {
                type: 'string',
                description: 'Código de autorización'
              },
              datos_transaccion: {
                type: 'object',
                description: 'Datos adicionales de la transacción'
              }
            }
          }
        }
      }
    },
    response: {
      200: {
        type: 'object',
        properties: {
          success: { type: 'boolean' },
          message: { type: 'string' },
          data: {
            type: 'object',
            properties: {
              venta: { type: 'object' },
              notaOriginal: notaVentaSchema
            }
          }
        }
      }
    }
  }

  // Todas las rutas requieren autenticación
  fastify.addHook('preHandler', fastify.authenticate)

  /**
   * GET /api/notas-venta
   * Obtiene lista de notas de venta con filtros
   */
  fastify.get('/', {
    schema: getNotasSchema
  }, NotaVentaController.getNotas)

  /**
   * GET /api/notas-venta/search
   * Busca notas de venta con filtros avanzados
   */
  fastify.get('/search', {
    schema: {
      description: 'Busca notas de venta con filtros avanzados',
      tags: ['Notas de Venta'],
      security: [{ Bearer: [] }],
      querystring: {
        type: 'object',
        properties: {
          q: {
            type: 'string',
            description: 'Término de búsqueda general'
          },
          numeroNota: {
            type: 'integer',
            description: 'Número específico de nota'
          },
          clienteRut: {
            type: 'string',
            description: 'RUT del cliente'
          },
          clienteNombre: {
            type: 'string',
            description: 'Nombre del cliente'
          },
          vendedorId: {
            type: 'string',
            description: 'ID del vendedor'
          },
          sucursalId: {
            type: 'string',
            description: 'ID de la sucursal'
          },
          tipoNota: {
            type: 'string',
            enum: ['cotizacion', 'reserva'],
            description: 'Tipo de nota'
          },
          estado: {
            type: 'string',
            enum: ['activa', 'convertida', 'anulada', 'vencida'],
            description: 'Estado de la nota'
          },
          fechaInicio: {
            type: 'string',
            format: 'date',
            description: 'Fecha de inicio'
          },
          fechaFin: {
            type: 'string',
            format: 'date',
            description: 'Fecha de fin'
          },
          montoMin: {
            type: 'number',
            description: 'Monto mínimo'
          },
          montoMax: {
            type: 'number',
            description: 'Monto máximo'
          },
          page: {
            type: 'integer',
            minimum: 1,
            default: 1
          },
          limit: {
            type: 'integer',
            minimum: 1,
            maximum: 100,
            default: 20
          }
        }
      },
      response: {
        200: {
          type: 'object',
          properties: {
            success: { type: 'boolean' },
            data: {
              type: 'array',
              items: notaVentaSchema
            },
            pagination: {
              type: 'object',
              properties: {
                page: { type: 'integer' },
                limit: { type: 'integer' },
                total: { type: 'integer' }
              }
            }
          }
        }
      }
    }
  }, NotaVentaController.searchNotas)

  /**
   * GET /api/notas-venta/stats
   * Obtiene estadísticas de notas de venta
   */
  fastify.get('/stats', {
    schema: {
      description: 'Obtiene estadísticas de notas de venta',
      tags: ['Notas de Venta'],
      security: [{ Bearer: [] }],
      querystring: {
        type: 'object',
        properties: {
          fechaInicio: {
            type: 'string',
            format: 'date',
            description: 'Fecha de inicio del período'
          },
          fechaFin: {
            type: 'string',
            format: 'date',
            description: 'Fecha de fin del período'
          },
          sucursalId: {
            type: 'string',
            description: 'ID de sucursal (solo para admin)'
          }
        }
      },
      response: {
        200: {
          type: 'object',
          properties: {
            success: { type: 'boolean' },
            data: {
              type: 'object',
              properties: {
                total_notas: { type: 'integer' },
                notas_activas: { type: 'integer' },
                notas_convertidas: { type: 'integer' },
                notas_anuladas: { type: 'integer' },
                cotizaciones: { type: 'integer' },
                reservas: { type: 'integer' },
                monto_total: { type: 'number' },
                promedio_nota: { type: 'number' },
                vendedores_activos: { type: 'integer' },
                clientes_unicos: { type: 'integer' }
              }
            }
          }
        }
      }
    }
  }, NotaVentaController.getNotasStats)

  /**
   * GET /api/notas-venta/proximas-vencer
   * Obtiene notas próximas a vencer
   */
  fastify.get('/proximas-vencer', {
    schema: {
      description: 'Obtiene notas próximas a vencer',
      tags: ['Notas de Venta'],
      security: [{ Bearer: [] }],
      querystring: {
        type: 'object',
        properties: {
          diasAnticipacion: {
            type: 'integer',
            description: 'Días de anticipación para alertas',
            minimum: 1,
            maximum: 30,
            default: 3
          }
        }
      },
      response: {
        200: {
          type: 'object',
          properties: {
            success: { type: 'boolean' },
            data: {
              type: 'array',
              items: {
                allOf: [
                  notaVentaSchema,
                  {
                    type: 'object',
                    properties: {
                      dias_restantes: { type: 'number' }
                    }
                  }
                ]
              }
            }
          }
        }
      }
    }
  }, NotaVentaController.getNotasProximasVencer)

  /**
   * POST /api/notas-venta
   * Crea una nueva nota de venta
   */
  fastify.post('/', {
    schema: createNotaSchema
  }, NotaVentaController.createNota)

  /**
   * GET /api/notas-venta/:id
   * Obtiene una nota de venta por ID
   */
  fastify.get('/:id', {
    preHandler: validateUuidParam,
    schema: {
      description: 'Obtiene una nota de venta por ID',
      tags: ['Notas de Venta'],
      security: [{ Bearer: [] }],
      params: {
        type: 'object',
        required: ['id'],
        properties: {
          id: {
            type: 'string',
            description: 'ID de la nota de venta'
          }
        }
      },
      response: {
        200: {
          type: 'object',
          properties: {
            success: { type: 'boolean' },
            data: {
              type: 'object',
              properties: {
                nota: notaVentaSchema,
                detalles: {
                  type: 'array',
                  items: detalleNotaSchema
                },
                reservas: {
                  type: 'array',
                  items: { type: 'object' }
                }
              }
            }
          }
        },
        404: {
          type: 'object',
          properties: {
            code: { type: 'string' },
            error: { type: 'string' },
            message: { type: 'string' }
          }
        }
      }
    }
  }, NotaVentaController.getNota)

  /**
   * PUT /api/notas-venta/:id
   * Actualiza una nota de venta
   */
  fastify.put('/:id', {
    preHandler: validateUuidParam,
    schema: {
      description: 'Actualiza una nota de venta',
      tags: ['Notas de Venta'],
      security: [{ Bearer: [] }],
      params: {
        type: 'object',
        required: ['id'],
        properties: {
          id: {
            type: 'string',
            description: 'ID de la nota de venta'
          }
        }
      },
      body: {
        type: 'object',
        properties: {
          cliente_rut: { type: 'string' },
          cliente_nombre: { type: 'string', maxLength: 200 },
          cliente_telefono: { type: 'string', maxLength: 20 },
          cliente_email: { type: 'string', format: 'email' },
          observaciones: { type: 'string', maxLength: 1000 }
        }
      },
      response: {
        200: {
          type: 'object',
          properties: {
            success: { type: 'boolean' },
            message: { type: 'string' },
            data: notaVentaSchema
          }
        }
      }
    }
  }, NotaVentaController.updateNota)

  /**
   * POST /api/notas-venta/:id/convertir
   * Convierte una nota de venta en venta real
   */
  fastify.post('/:id/convertir', {
    preHandler: validateUuidParam,
    schema: convertirAVentaSchema
  }, NotaVentaController.convertirAVenta)

  /**
   * POST /api/notas-venta/:id/anular
   * Anula una nota de venta
   */
  fastify.post('/:id/anular', {
    preHandler: validateUuidParam,
    schema: {
      description: 'Anula una nota de venta',
      tags: ['Notas de Venta'],
      security: [{ Bearer: [] }],
      params: {
        type: 'object',
        required: ['id'],
        properties: {
          id: {
            type: 'string',
            description: 'ID de la nota de venta'
          }
        }
      },
      body: {
        type: 'object',
        required: ['motivo'],
        properties: {
          motivo: {
            type: 'string',
            description: 'Motivo de la anulación',
            minLength: 10,
            maxLength: 500
          }
        }
      },
      response: {
        200: {
          type: 'object',
          properties: {
            success: { type: 'boolean' },
            message: { type: 'string' },
            data: notaVentaSchema
          }
        }
      }
    }
  }, NotaVentaController.anularNota)

  /**
   * POST /api/notas-venta/:id/duplicar
   * Duplica una nota de venta existente
   */
  fastify.post('/:id/duplicar', {
    preHandler: validateUuidParam,
    schema: {
      description: 'Duplica una nota de venta existente',
      tags: ['Notas de Venta'],
      security: [{ Bearer: [] }],
      params: {
        type: 'object',
        required: ['id'],
        properties: {
          id: {
            type: 'string',
            description: 'ID de la nota de venta a duplicar'
          }
        }
      },
      response: {
        201: {
          type: 'object',
          properties: {
            success: { type: 'boolean' },
            message: { type: 'string' },
            data: {
              type: 'object',
              properties: {
                nota: notaVentaSchema,
                detalles: {
                  type: 'array',
                  items: detalleNotaSchema
                }
              }
            }
          }
        }
      }
    }
  }, NotaVentaController.duplicarNota)

  /**
   * GET /api/notas-venta/:id/pdf
   * Exporta una nota de venta a PDF
   */
  fastify.get('/:id/pdf', {
    preHandler: validateUuidParam,
    schema: {
      description: 'Exporta una nota de venta a PDF',
      tags: ['Notas de Venta'],
      security: [{ Bearer: [] }],
      params: {
        type: 'object',
        required: ['id'],
        properties: {
          id: {
            type: 'string',
            description: 'ID de la nota de venta'
          }
        }
      },
      response: {
        200: {
          type: 'object',
          properties: {
            success: { type: 'boolean' },
            message: { type: 'string' },
            data: { type: 'object' }
          }
        }
      }
    }
  }, NotaVentaController.exportarNotaPDF)

  /**
   * GET /api/notas-venta/:id/historial
   * Obtiene el historial de una nota de venta
   */
  fastify.get('/:id/historial', {
    preHandler: validateUuidParam,
    schema: {
      description: 'Obtiene el historial de cambios de una nota de venta',
      tags: ['Notas de Venta'],
      security: [{ Bearer: [] }],
      params: {
        type: 'object',
        required: ['id'],
        properties: {
          id: {
            type: 'string',
            description: 'ID de la nota de venta'
          }
        }
      },
      response: {
        200: {
          type: 'object',
          properties: {
            success: { type: 'boolean' },
            data: {
              type: 'array',
              items: {
                type: 'object',
                properties: {
                  fecha: { type: 'string', format: 'date-time' },
                  accion: { type: 'string' },
                  usuario: { type: 'string' },
                  detalles: { type: 'string' }
                }
              }
            }
          }
        }
      }
    }
  }, NotaVentaController.getHistorialNota)
}

module.exports = notasVentaRoutes

