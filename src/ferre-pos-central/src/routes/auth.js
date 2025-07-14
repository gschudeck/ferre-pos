/**
 * Rutas de Autenticación - Sistema Ferre-POS
 * 
 * Define todos los endpoints relacionados con autenticación,
 * autorización y gestión de sesiones.
 */

const AuthController = require('../controllers/AuthController')

async function authRoutes(fastify, options) {
  // Esquemas de validación para Swagger
  const loginSchema = {
    description: 'Inicia sesión de usuario',
    tags: ['Autenticación'],
    body: {
      type: 'object',
      required: ['rut', 'password'],
      properties: {
        rut: {
          type: 'string',
          description: 'RUT del usuario',
          example: '12345678-9'
        },
        password: {
          type: 'string',
          description: 'Contraseña del usuario',
          example: 'password123'
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
              token: { type: 'string' },
              user: { type: 'object' },
              expiresIn: { type: 'string' }
            }
          }
        }
      },
      401: {
        type: 'object',
        properties: {
          code: { type: 'string' },
          error: { type: 'string' },
          message: { type: 'string' }
        }
      }
    }
  }

  const changePasswordSchema = {
    description: 'Cambia la contraseña del usuario autenticado',
    tags: ['Autenticación'],
    security: [{ Bearer: [] }],
    body: {
      type: 'object',
      required: ['currentPassword', 'newPassword', 'confirmPassword'],
      properties: {
        currentPassword: {
          type: 'string',
          description: 'Contraseña actual'
        },
        newPassword: {
          type: 'string',
          description: 'Nueva contraseña',
          minLength: 6
        },
        confirmPassword: {
          type: 'string',
          description: 'Confirmación de nueva contraseña'
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

  const resetPasswordSchema = {
    description: 'Resetea la contraseña de un usuario (solo admin)',
    tags: ['Autenticación'],
    security: [{ Bearer: [] }],
    body: {
      type: 'object',
      required: ['userId', 'newPassword'],
      properties: {
        userId: {
          type: 'string',
          description: 'ID del usuario'
        },
        newPassword: {
          type: 'string',
          description: 'Nueva contraseña',
          minLength: 6
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

  // Rutas públicas (sin autenticación)
  
  /**
   * POST /api/auth/login
   * Inicia sesión de usuario
   */
  fastify.post('/login', {
    schema: loginSchema
  }, AuthController.login)

  // Rutas protegidas (requieren autenticación)
  
  /**
   * POST /api/auth/logout
   * Cierra sesión de usuario
   */
  fastify.post('/logout', {
    preHandler: fastify.authenticate,
    schema: {
      description: 'Cierra sesión de usuario',
      tags: ['Autenticación'],
      security: [{ Bearer: [] }],
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
  }, AuthController.logout)

  /**
   * POST /api/auth/refresh
   * Renueva el token JWT
   */
  fastify.post('/refresh', {
    preHandler: fastify.authenticate,
    schema: {
      description: 'Renueva el token JWT',
      tags: ['Autenticación'],
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
                token: { type: 'string' },
                expiresIn: { type: 'string' }
              }
            }
          }
        }
      }
    }
  }, AuthController.refreshToken)

  /**
   * GET /api/auth/profile
   * Obtiene el perfil del usuario autenticado
   */
  fastify.get('/profile', {
    preHandler: fastify.authenticate,
    schema: {
      description: 'Obtiene el perfil del usuario autenticado',
      tags: ['Autenticación'],
      security: [{ Bearer: [] }],
      response: {
        200: {
          type: 'object',
          properties: {
            success: { type: 'boolean' },
            data: { type: 'object' }
          }
        }
      }
    }
  }, AuthController.getProfile)

  /**
   * POST /api/auth/change-password
   * Cambia la contraseña del usuario autenticado
   */
  fastify.post('/change-password', {
    preHandler: fastify.authenticate,
    schema: changePasswordSchema
  }, AuthController.changePassword)

  /**
   * POST /api/auth/verify
   * Verifica si el token es válido
   */
  fastify.post('/verify', {
    preHandler: fastify.authenticate,
    schema: {
      description: 'Verifica si el token es válido',
      tags: ['Autenticación'],
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
                valid: { type: 'boolean' },
                user: { type: 'object' }
              }
            }
          }
        }
      }
    }
  }, AuthController.verifyToken)

  /**
   * GET /api/auth/session
   * Obtiene información de la sesión actual
   */
  fastify.get('/session', {
    preHandler: fastify.authenticate,
    schema: {
      description: 'Obtiene información de la sesión actual',
      tags: ['Autenticación'],
      security: [{ Bearer: [] }],
      response: {
        200: {
          type: 'object',
          properties: {
            success: { type: 'boolean' },
            data: { type: 'object' }
          }
        }
      }
    }
  }, AuthController.getSessionInfo)

  /**
   * POST /api/auth/reset-password
   * Resetea la contraseña de un usuario (solo admin)
   */
  fastify.post('/reset-password', {
    preHandler: [fastify.authenticate, fastify.authorize(['admin'])],
    schema: resetPasswordSchema
  }, AuthController.resetPassword)

  /**
   * GET /api/auth/stats
   * Obtiene estadísticas de autenticación (solo admin)
   */
  fastify.get('/stats', {
    preHandler: [fastify.authenticate, fastify.authorize(['admin'])],
    schema: {
      description: 'Obtiene estadísticas de autenticación (solo admin)',
      tags: ['Autenticación'],
      security: [{ Bearer: [] }],
      response: {
        200: {
          type: 'object',
          properties: {
            success: { type: 'boolean' },
            data: { type: 'object' }
          }
        }
      }
    }
  }, AuthController.getAuthStats)
}

module.exports = authRoutes

