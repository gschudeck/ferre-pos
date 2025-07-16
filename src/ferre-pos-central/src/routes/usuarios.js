/**
 * Rutas de Usuarios - Sistema Ferre-POS
 * 
 * Define todos los endpoints relacionados con la gestión de usuarios,
 * perfiles, autenticación y administración de cuentas.
 */

const UsuarioController = require('../controllers/UsuarioController')
const { validateBody, validateQuery, validateParams } = require('../middleware/validation')

async function usuarioRoutes(fastify, options) {
  // Esquemas de validación
  const usuarioSchema = {
    type: 'object',
    properties: {
      id: { type: 'string', format: 'uuid' },
      rut: { type: 'string' },
      nombre: { type: 'string' },
      email: { type: 'string', format: 'email' },
      telefono: { type: 'string' },
      rol: { type: 'string', enum: ['admin', 'gerente', 'vendedor', 'cajero'] },
      sucursal_id: { type: 'string', format: 'uuid' },
      activo: { type: 'boolean' },
      ultimo_acceso: { type: 'string', format: 'date-time' },
      debe_cambiar_password: { type: 'boolean' },
      fecha_creacion: { type: 'string', format: 'date-time' }
    }
  }

  const createUsuarioSchema = {
    description: 'Crea un nuevo usuario en el sistema',
    tags: ['Usuarios'],
    security: [{ Bearer: [] }],
    body: {
      type: 'object',
      required: ['rut', 'nombre', 'email', 'rol', 'password'],
      properties: {
        rut: {
          type: 'string',
          pattern: '^[0-9]{7,8}-[0-9Kk]$',
          description: 'RUT del usuario (formato: 12345678-9)'
        },
        nombre: {
          type: 'string',
          minLength: 2,
          maxLength: 100,
          description: 'Nombre completo del usuario'
        },
        email: {
          type: 'string',
          format: 'email',
          maxLength: 100,
          description: 'Email del usuario'
        },
        telefono: {
          type: 'string',
          maxLength: 20,
          description: 'Teléfono del usuario'
        },
        rol: {
          type: 'string',
          enum: ['admin', 'gerente', 'vendedor', 'cajero'],
          description: 'Rol del usuario en el sistema'
        },
        sucursal_id: {
          type: 'string',
          format: 'uuid',
          description: 'ID de la sucursal asignada'
        },
        password: {
          type: 'string',
          minLength: 8,
          description: 'Contraseña inicial del usuario'
        },
        debe_cambiar_password: {
          type: 'boolean',
          default: true,
          description: 'Si debe cambiar contraseña en el primer login'
        }
      }
    },
    response: {
      201: {
        type: 'object',
        properties: {
          success: { type: 'boolean' },
          message: { type: 'string' },
          data: usuarioSchema
        }
      }
    }
  }

  const updateUsuarioSchema = {
    description: 'Actualiza un usuario existente',
    tags: ['Usuarios'],
    security: [{ Bearer: [] }],
    params: {
      type: 'object',
      required: ['id'],
      properties: {
        id: {
          type: 'string',
          format: 'uuid',
          description: 'ID del usuario a actualizar'
        }
      }
    },
    body: {
      type: 'object',
      properties: {
        nombre: {
          type: 'string',
          minLength: 2,
          maxLength: 100,
          description: 'Nombre completo del usuario'
        },
        email: {
          type: 'string',
          format: 'email',
          maxLength: 100,
          description: 'Email del usuario'
        },
        telefono: {
          type: 'string',
          maxLength: 20,
          description: 'Teléfono del usuario'
        },
        rol: {
          type: 'string',
          enum: ['admin', 'gerente', 'vendedor', 'cajero'],
          description: 'Rol del usuario en el sistema'
        },
        sucursal_id: {
          type: 'string',
          format: 'uuid',
          description: 'ID de la sucursal asignada'
        },
        activo: {
          type: 'boolean',
          description: 'Estado activo del usuario'
        }
      }
    },
    response: {
      200: {
        type: 'object',
        properties: {
          success: { type: 'boolean' },
          message: { type: 'string' },
          data: usuarioSchema
        }
      }
    }
  }

  const cambiarPasswordSchema = {
    description: 'Cambia la contraseña de un usuario',
    tags: ['Usuarios'],
    security: [{ Bearer: [] }],
    params: {
      type: 'object',
      required: ['id'],
      properties: {
        id: {
          type: 'string',
          format: 'uuid',
          description: 'ID del usuario'
        }
      }
    },
    body: {
      type: 'object',
      required: ['password_nueva'],
      properties: {
        password_actual: {
          type: 'string',
          description: 'Contraseña actual (requerida si no es admin)'
        },
        password_nueva: {
          type: 'string',
          minLength: 8,
          description: 'Nueva contraseña'
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

  const recuperarPasswordSchema = {
    description: 'Inicia el proceso de recuperación de contraseña',
    tags: ['Usuarios'],
    body: {
      type: 'object',
      required: ['email'],
      properties: {
        email: {
          type: 'string',
          format: 'email',
          description: 'Email del usuario'
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

  const completarRecuperacionSchema = {
    description: 'Completa el proceso de recuperación de contraseña',
    tags: ['Usuarios'],
    body: {
      type: 'object',
      required: ['token', 'password_nueva'],
      properties: {
        token: {
          type: 'string',
          description: 'Token de recuperación'
        },
        password_nueva: {
          type: 'string',
          minLength: 8,
          description: 'Nueva contraseña'
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

  const usuariosListSchema = {
    description: 'Obtiene lista de usuarios con filtros y paginación',
    tags: ['Usuarios'],
    security: [{ Bearer: [] }],
    querystring: {
      type: 'object',
      properties: {
        sucursal_id: {
          type: 'string',
          format: 'uuid',
          description: 'Filtrar por sucursal'
        },
        rol: {
          type: 'string',
          enum: ['admin', 'gerente', 'vendedor', 'cajero'],
          description: 'Filtrar por rol'
        },
        activo: {
          type: 'boolean',
          description: 'Filtrar por estado activo'
        },
        busqueda: {
          type: 'string',
          description: 'Búsqueda en nombre, email o RUT'
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
          maximum: 100,
          default: 20,
          description: 'Usuarios por página'
        },
        order_by: {
          type: 'string',
          enum: ['nombre', 'email', 'rol', 'fecha_creacion', 'ultimo_acceso'],
          default: 'nombre',
          description: 'Campo para ordenar'
        },
        order_direction: {
          type: 'string',
          enum: ['ASC', 'DESC'],
          default: 'ASC',
          description: 'Dirección del ordenamiento'
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
            items: usuarioSchema
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

  // Todas las rutas requieren autenticación excepto recuperación de contraseña
  fastify.addHook('preHandler', async (request, reply) => {
    // Rutas públicas
    const rutasPublicas = [
      '/recuperar-password',
      '/completar-recuperacion'
    ]
    
    if (!rutasPublicas.includes(request.routerPath)) {
      await fastify.authenticate(request, reply)
    }
  })

  /**
   * GET /api/usuarios
   * Obtiene lista de usuarios con filtros y paginación
   */
  fastify.get('/', {
    schema: usuariosListSchema
  }, UsuarioController.getUsuarios)

  /**
   * GET /api/usuarios/perfil
   * Obtiene el perfil del usuario autenticado
   */
  fastify.get('/perfil', {
    schema: {
      description: 'Obtiene el perfil del usuario autenticado',
      tags: ['Usuarios'],
      security: [{ Bearer: [] }],
      response: {
        200: {
          type: 'object',
          properties: {
            success: { type: 'boolean' },
            data: usuarioSchema
          }
        }
      }
    }
  }, UsuarioController.getPerfil)

  /**
   * PUT /api/usuarios/perfil
   * Actualiza el perfil del usuario autenticado
   */
  fastify.put('/perfil', {
    schema: {
      description: 'Actualiza el perfil del usuario autenticado',
      tags: ['Usuarios'],
      security: [{ Bearer: [] }],
      body: {
        type: 'object',
        properties: {
          nombre: {
            type: 'string',
            minLength: 2,
            maxLength: 100,
            description: 'Nombre completo del usuario'
          },
          email: {
            type: 'string',
            format: 'email',
            maxLength: 100,
            description: 'Email del usuario'
          },
          telefono: {
            type: 'string',
            maxLength: 20,
            description: 'Teléfono del usuario'
          }
        }
      },
      response: {
        200: {
          type: 'object',
          properties: {
            success: { type: 'boolean' },
            message: { type: 'string' },
            data: usuarioSchema
          }
        }
      }
    }
  }, UsuarioController.updatePerfil)

  /**
   * GET /api/usuarios/estadisticas
   * Obtiene estadísticas de usuarios
   */
  fastify.get('/estadisticas', {
    schema: {
      description: 'Obtiene estadísticas de usuarios',
      tags: ['Usuarios'],
      security: [{ Bearer: [] }],
      querystring: {
        type: 'object',
        properties: {
          sucursal_id: {
            type: 'string',
            format: 'uuid',
            description: 'Filtrar por sucursal'
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
                total_usuarios: { type: 'integer' },
                usuarios_activos: { type: 'integer' },
                usuarios_inactivos: { type: 'integer' },
                administradores: { type: 'integer' },
                gerentes: { type: 'integer' },
                vendedores: { type: 'integer' },
                cajeros: { type: 'integer' },
                activos_hoy: { type: 'integer' },
                activos_semana: { type: 'integer' }
              }
            }
          }
        }
      }
    }
  }, UsuarioController.getEstadisticas)

  /**
   * POST /api/usuarios
   * Crea un nuevo usuario
   */
  fastify.post('/', {
    schema: createUsuarioSchema
  }, UsuarioController.createUsuario)

  /**
   * GET /api/usuarios/:id
   * Obtiene un usuario específico por ID
   */
  fastify.get('/:id', {
    schema: {
      description: 'Obtiene un usuario específico por ID',
      tags: ['Usuarios'],
      security: [{ Bearer: [] }],
      params: {
        type: 'object',
        required: ['id'],
        properties: {
          id: {
            type: 'string',
            format: 'uuid',
            description: 'ID del usuario'
          }
        }
      },
      response: {
        200: {
          type: 'object',
          properties: {
            success: { type: 'boolean' },
            data: {
              allOf: [
                usuarioSchema,
                {
                  type: 'object',
                  properties: {
                    sucursal_nombre: { type: 'string' },
                    creador_nombre: { type: 'string' }
                  }
                }
              ]
            }
          }
        }
      }
    }
  }, UsuarioController.getUsuario)

  /**
   * PUT /api/usuarios/:id
   * Actualiza un usuario existente
   */
  fastify.put('/:id', {
    schema: updateUsuarioSchema
  }, UsuarioController.updateUsuario)

  /**
   * PUT /api/usuarios/:id/password
   * Cambia la contraseña de un usuario
   */
  fastify.put('/:id/password', {
    schema: cambiarPasswordSchema
  }, UsuarioController.cambiarPassword)

  /**
   * POST /api/usuarios/:id/desactivar
   * Desactiva un usuario (eliminación lógica)
   */
  fastify.post('/:id/desactivar', {
    schema: {
      description: 'Desactiva un usuario (eliminación lógica)',
      tags: ['Usuarios'],
      security: [{ Bearer: [] }],
      params: {
        type: 'object',
        required: ['id'],
        properties: {
          id: {
            type: 'string',
            format: 'uuid',
            description: 'ID del usuario a desactivar'
          }
        }
      },
      body: {
        type: 'object',
        required: ['motivo'],
        properties: {
          motivo: {
            type: 'string',
            minLength: 10,
            maxLength: 500,
            description: 'Motivo de la desactivación'
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
  }, UsuarioController.desactivarUsuario)

  /**
   * POST /api/usuarios/:id/reactivar
   * Reactiva un usuario
   */
  fastify.post('/:id/reactivar', {
    schema: {
      description: 'Reactiva un usuario previamente desactivado',
      tags: ['Usuarios'],
      security: [{ Bearer: [] }],
      params: {
        type: 'object',
        required: ['id'],
        properties: {
          id: {
            type: 'string',
            format: 'uuid',
            description: 'ID del usuario a reactivar'
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
  }, UsuarioController.reactivarUsuario)

  /**
   * POST /api/usuarios/:id/desbloquear
   * Desbloquea un usuario bloqueado
   */
  fastify.post('/:id/desbloquear', {
    schema: {
      description: 'Desbloquea un usuario bloqueado por intentos fallidos',
      tags: ['Usuarios'],
      security: [{ Bearer: [] }],
      params: {
        type: 'object',
        required: ['id'],
        properties: {
          id: {
            type: 'string',
            format: 'uuid',
            description: 'ID del usuario a desbloquear'
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
  }, UsuarioController.desbloquearUsuario)

  /**
   * POST /api/usuarios/:id/forzar-cambio-password
   * Fuerza el cambio de contraseña en el próximo login
   */
  fastify.post('/:id/forzar-cambio-password', {
    schema: {
      description: 'Fuerza el cambio de contraseña en el próximo login',
      tags: ['Usuarios'],
      security: [{ Bearer: [] }],
      params: {
        type: 'object',
        required: ['id'],
        properties: {
          id: {
            type: 'string',
            format: 'uuid',
            description: 'ID del usuario'
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
  }, UsuarioController.forzarCambioPassword)

  /**
   * GET /api/usuarios/:id/historial-accesos
   * Obtiene el historial de accesos de un usuario
   */
  fastify.get('/:id/historial-accesos', {
    schema: {
      description: 'Obtiene el historial de accesos de un usuario',
      tags: ['Usuarios'],
      security: [{ Bearer: [] }],
      params: {
        type: 'object',
        required: ['id'],
        properties: {
          id: {
            type: 'string',
            format: 'uuid',
            description: 'ID del usuario'
          }
        }
      },
      querystring: {
        type: 'object',
        properties: {
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
          exitosos: {
            type: 'boolean',
            description: 'Filtrar solo accesos exitosos'
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
            maximum: 100,
            default: 50,
            description: 'Registros por página'
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
                  exitoso: { type: 'boolean' },
                  motivo: { type: 'string' },
                  ip_address: { type: 'string' }
                }
              }
            }
          }
        }
      }
    }
  }, UsuarioController.getHistorialAccesos)

  /**
   * POST /api/usuarios/recuperar-password
   * Inicia el proceso de recuperación de contraseña
   */
  fastify.post('/recuperar-password', {
    schema: recuperarPasswordSchema
  }, UsuarioController.iniciarRecuperacionPassword)

  /**
   * POST /api/usuarios/completar-recuperacion
   * Completa el proceso de recuperación de contraseña
   */
  fastify.post('/completar-recuperacion', {
    schema: completarRecuperacionSchema
  }, UsuarioController.completarRecuperacionPassword)

  /**
   * POST /api/usuarios/mantenimiento
   * Ejecuta tareas de mantenimiento de usuarios
   */
  fastify.post('/mantenimiento', {
    schema: {
      description: 'Ejecuta tareas de mantenimiento de usuarios',
      tags: ['Usuarios'],
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
                timestamp: { type: 'string' },
                tareas_ejecutadas: { type: 'array' }
              }
            }
          }
        }
      }
    }
  }, UsuarioController.ejecutarMantenimiento)
}

module.exports = usuarioRoutes

