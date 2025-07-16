/**
 * Rutas de Sistema - Sistema Ferre-POS
 * 
 * Define todos los endpoints relacionados con la administración del sistema,
 * configuraciones, monitoreo, logs y utilidades administrativas.
 */

const SistemaController = require('../controllers/SistemaController')
const { validateBody, validateQuery } = require('../middleware/validation')

async function sistemaRoutes(fastify, options) {
  // Esquemas de validación
  const configuracionSchema = {
    type: 'object',
    properties: {
      clave: { type: 'string' },
      valor: { type: 'string' },
      descripcion: { type: 'string' },
      tipo_dato: { type: 'string', enum: ['string', 'number', 'integer', 'boolean', 'json'] },
      categoria: { type: 'string' },
      activa: { type: 'boolean' },
      solo_lectura: { type: 'boolean' }
    }
  }

  const infoSistemaSchema = {
    description: 'Obtiene información general del sistema',
    tags: ['Sistema'],
    security: [{ Bearer: [] }],
    response: {
      200: {
        type: 'object',
        properties: {
          success: { type: 'boolean' },
          data: {
            type: 'object',
            properties: {
              version: { type: 'string' },
              nombre: { type: 'string' },
              descripcion: { type: 'string' },
              ambiente: { type: 'string' },
              fecha_inicio: { type: 'string' },
              uptime: { type: 'number' },
              memoria: { type: 'object' },
              plataforma: { type: 'string' },
              version_node: { type: 'string' },
              base_datos: { type: 'object' },
              configuraciones: { type: 'object' }
            }
          }
        }
      }
    }
  }

  const configuracionesSchema = {
    description: 'Obtiene todas las configuraciones del sistema',
    tags: ['Sistema'],
    security: [{ Bearer: [] }],
    querystring: {
      type: 'object',
      properties: {
        incluir_inactivas: {
          type: 'boolean',
          description: 'Incluir configuraciones inactivas',
          default: false
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
            additionalProperties: {
              type: 'object',
              additionalProperties: configuracionSchema
            }
          }
        }
      }
    }
  }

  const createConfigSchema = {
    description: 'Crea una nueva configuración del sistema',
    tags: ['Sistema'],
    security: [{ Bearer: [] }],
    body: {
      type: 'object',
      required: ['clave', 'valor', 'tipo_dato', 'categoria'],
      properties: {
        clave: {
          type: 'string',
          description: 'Clave única de la configuración',
          pattern: '^[a-z0-9_]+$',
          maxLength: 100
        },
        valor: {
          type: 'string',
          description: 'Valor de la configuración',
          maxLength: 1000
        },
        descripcion: {
          type: 'string',
          description: 'Descripción de la configuración',
          maxLength: 500
        },
        tipo_dato: {
          type: 'string',
          enum: ['string', 'number', 'integer', 'boolean', 'json'],
          description: 'Tipo de dato del valor'
        },
        categoria: {
          type: 'string',
          description: 'Categoría de la configuración',
          maxLength: 50
        },
        solo_lectura: {
          type: 'boolean',
          description: 'Si la configuración es de solo lectura',
          default: false
        }
      }
    },
    response: {
      201: {
        type: 'object',
        properties: {
          success: { type: 'boolean' },
          message: { type: 'string' },
          data: configuracionSchema
        }
      }
    }
  }

  const updateConfigSchema = {
    description: 'Actualiza una configuración del sistema',
    tags: ['Sistema'],
    security: [{ Bearer: [] }],
    params: {
      type: 'object',
      required: ['clave'],
      properties: {
        clave: {
          type: 'string',
          description: 'Clave de la configuración a actualizar'
        }
      }
    },
    body: {
      type: 'object',
      required: ['valor'],
      properties: {
        valor: {
          type: 'string',
          description: 'Nuevo valor de la configuración',
          maxLength: 1000
        }
      }
    },
    response: {
      200: {
        type: 'object',
        properties: {
          success: { type: 'boolean' },
          message: { type: 'string' },
          data: configuracionSchema
        }
      }
    }
  }

  const logsSchema = {
    description: 'Obtiene logs del sistema',
    tags: ['Sistema'],
    security: [{ Bearer: [] }],
    querystring: {
      type: 'object',
      properties: {
        nivel: {
          type: 'string',
          enum: ['error', 'warn', 'info', 'debug'],
          description: 'Nivel de log a filtrar'
        },
        fecha_inicio: {
          type: 'string',
          format: 'date-time',
          description: 'Fecha de inicio del período'
        },
        fecha_fin: {
          type: 'string',
          format: 'date-time',
          description: 'Fecha de fin del período'
        },
        usuario: {
          type: 'string',
          description: 'ID del usuario'
        },
        modulo: {
          type: 'string',
          description: 'Módulo del sistema'
        },
        page: {
          type: 'integer',
          minimum: 1,
          default: 1,
          description: 'Número de página'
        },
        limit: {
          type: 'integer',
          minimum: 1,
          maximum: 1000,
          default: 100,
          description: 'Cantidad de logs por página'
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
                timestamp: { type: 'string' },
                level: { type: 'string' },
                message: { type: 'string' },
                meta: { type: 'object' },
                module: { type: 'string' },
                user_id: { type: 'string' }
              }
            }
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

  const mantenimientoSchema = {
    description: 'Ejecuta tareas de mantenimiento del sistema',
    tags: ['Sistema'],
    security: [{ Bearer: [] }],
    body: {
      type: 'object',
      properties: {
        limpiar_logs: {
          type: 'boolean',
          description: 'Limpiar logs antiguos',
          default: false
        },
        dias_retencion_logs: {
          type: 'integer',
          minimum: 7,
          maximum: 365,
          default: 30,
          description: 'Días de retención de logs'
        },
        optimizar_bd: {
          type: 'boolean',
          description: 'Optimizar base de datos',
          default: false
        },
        backup_configuraciones: {
          type: 'boolean',
          description: 'Crear backup de configuraciones',
          default: false
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
              timestamp: { type: 'string' },
              tareas_ejecutadas: { type: 'array' },
              errores: { type: 'array' }
            }
          }
        }
      }
    }
  }

  // Todas las rutas requieren autenticación
  fastify.addHook('preHandler', fastify.authenticate)

  /**
   * GET /api/sistema/info
   * Obtiene información general del sistema
   */
  fastify.get('/info', {
    schema: infoSistemaSchema
  }, SistemaController.getInfoSistema)

  /**
   * GET /api/sistema/health
   * Health check del sistema
   */
  fastify.get('/health', {
    schema: {
      description: 'Verifica el estado de salud del sistema',
      tags: ['Sistema'],
      response: {
        200: {
          type: 'object',
          properties: {
            status: { type: 'string' },
            timestamp: { type: 'string' },
            version: { type: 'string' },
            environment: { type: 'string' },
            uptime: { type: 'number' },
            memory: { type: 'object' },
            database: { type: 'object' },
            disk: { type: 'object' }
          }
        }
      }
    }
  }, SistemaController.getHealthCheck)

  /**
   * GET /api/sistema/estadisticas
   * Obtiene estadísticas generales del sistema
   */
  fastify.get('/estadisticas', {
    schema: {
      description: 'Obtiene estadísticas generales del sistema',
      tags: ['Sistema'],
      security: [{ Bearer: [] }],
      response: {
        200: {
          type: 'object',
          properties: {
            success: { type: 'boolean' },
            data: {
              type: 'object',
              properties: {
                total_usuarios: { type: 'integer' },
                total_productos: { type: 'integer' },
                total_ventas_hoy: { type: 'integer' },
                total_sucursales: { type: 'integer' },
                total_notas_activas: { type: 'integer' },
                tamaño_bd: { type: 'string' }
              }
            }
          }
        }
      }
    }
  }, SistemaController.getEstadisticas)

  /**
   * GET /api/sistema/metricas
   * Obtiene métricas de rendimiento del sistema
   */
  fastify.get('/metricas', {
    schema: {
      description: 'Obtiene métricas de rendimiento del sistema',
      tags: ['Sistema'],
      security: [{ Bearer: [] }],
      response: {
        200: {
          type: 'object',
          properties: {
            success: { type: 'boolean' },
            data: {
              type: 'object',
              properties: {
                memoria: { type: 'object' },
                cpu: { type: 'object' },
                uptime: { type: 'number' },
                timestamp: { type: 'string' },
                base_datos: { type: 'object' }
              }
            }
          }
        }
      }
    }
  }, SistemaController.getMetricasRendimiento)

  /**
   * GET /api/sistema/configuraciones
   * Obtiene todas las configuraciones del sistema
   */
  fastify.get('/configuraciones', {
    schema: configuracionesSchema
  }, SistemaController.getConfiguraciones)

  /**
   * GET /api/sistema/configuraciones/:clave
   * Obtiene una configuración específica
   */
  fastify.get('/configuraciones/:clave', {
    schema: {
      description: 'Obtiene una configuración específica del sistema',
      tags: ['Sistema'],
      security: [{ Bearer: [] }],
      params: {
        type: 'object',
        required: ['clave'],
        properties: {
          clave: {
            type: 'string',
            description: 'Clave de la configuración'
          }
        }
      },
      response: {
        200: {
          type: 'object',
          properties: {
            success: { type: 'boolean' },
            data: configuracionSchema
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
  }, SistemaController.getConfiguracion)

  /**
   * POST /api/sistema/configuraciones
   * Crea una nueva configuración del sistema
   */
  fastify.post('/configuraciones', {
    schema: createConfigSchema
  }, SistemaController.createConfiguracion)

  /**
   * PUT /api/sistema/configuraciones/:clave
   * Actualiza una configuración del sistema
   */
  fastify.put('/configuraciones/:clave', {
    schema: updateConfigSchema
  }, SistemaController.updateConfiguracion)

  /**
   * DELETE /api/sistema/configuraciones/:clave
   * Elimina una configuración del sistema
   */
  fastify.delete('/configuraciones/:clave', {
    schema: {
      description: 'Elimina una configuración del sistema',
      tags: ['Sistema'],
      security: [{ Bearer: [] }],
      params: {
        type: 'object',
        required: ['clave'],
        properties: {
          clave: {
            type: 'string',
            description: 'Clave de la configuración a eliminar'
          }
        }
      },
      response: {
        200: {
          type: 'object',
          properties: {
            success: { type: 'boolean' },
            message: { type: 'string' }
          }
        }
      }
    }
  }, SistemaController.deleteConfiguracion)

  /**
   * POST /api/sistema/backup
   * Crea un backup de las configuraciones del sistema
   */
  fastify.post('/backup', {
    schema: {
      description: 'Crea un backup de las configuraciones del sistema',
      tags: ['Sistema'],
      security: [{ Bearer: [] }],
      response: {
        200: {
          type: 'object',
          properties: {
            success: { type: 'boolean' },
            message: { type: 'string' },
            data: {
              type: 'object',
              properties: {
                archivo: { type: 'string' },
                ruta: { type: 'string' },
                tamaño: { type: 'number' },
                fecha: { type: 'string' }
              }
            }
          }
        }
      }
    }
  }, SistemaController.backupConfiguraciones)

  /**
   * POST /api/sistema/restaurar
   * Restaura configuraciones desde un backup
   */
  fastify.post('/restaurar', {
    schema: {
      description: 'Restaura configuraciones desde un backup',
      tags: ['Sistema'],
      security: [{ Bearer: [] }],
      body: {
        type: 'object',
        required: ['configuraciones'],
        properties: {
          fecha: { type: 'string' },
          version: { type: 'string' },
          configuraciones: { type: 'object' }
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
                configuracionesRestauradas: { type: 'integer' },
                errores: { type: 'array' }
              }
            }
          }
        }
      }
    }
  }, SistemaController.restaurarConfiguraciones)

  /**
   * GET /api/sistema/logs
   * Obtiene logs del sistema
   */
  fastify.get('/logs', {
    schema: logsSchema
  }, SistemaController.getLogs)

  /**
   * POST /api/sistema/logs/limpiar
   * Limpia logs antiguos del sistema
   */
  fastify.post('/logs/limpiar', {
    schema: {
      description: 'Limpia logs antiguos del sistema',
      tags: ['Sistema'],
      security: [{ Bearer: [] }],
      body: {
        type: 'object',
        properties: {
          dias_retencion: {
            type: 'integer',
            minimum: 7,
            maximum: 365,
            default: 30,
            description: 'Días de retención de logs'
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
                logsEliminados: { type: 'integer' },
                fechaLimite: { type: 'string' }
              }
            }
          }
        }
      }
    }
  }, SistemaController.limpiarLogs)

  /**
   * POST /api/sistema/mantenimiento
   * Ejecuta tareas de mantenimiento del sistema
   */
  fastify.post('/mantenimiento', {
    schema: mantenimientoSchema
  }, SistemaController.ejecutarMantenimiento)

  /**
   * POST /api/sistema/reiniciar
   * Reinicia el sistema de forma controlada
   */
  fastify.post('/reiniciar', {
    schema: {
      description: 'Reinicia el sistema de forma controlada',
      tags: ['Sistema'],
      security: [{ Bearer: [] }],
      body: {
        type: 'object',
        required: ['motivo'],
        properties: {
          motivo: {
            type: 'string',
            minLength: 10,
            maxLength: 500,
            description: 'Motivo del reinicio del sistema'
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
                motivo: { type: 'string' },
                tiempo_estimado: { type: 'string' }
              }
            }
          }
        }
      }
    }
  }, SistemaController.reiniciarSistema)
}

module.exports = sistemaRoutes

