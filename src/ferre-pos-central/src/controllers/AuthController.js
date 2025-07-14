/**
 * Controlador de Autenticación - Sistema Ferre-POS
 * 
 * Maneja todas las operaciones de autenticación y autorización,
 * incluyendo login, logout, renovación de tokens y gestión de sesiones.
 */

const Usuario = require('../models/Usuario')
const logger = require('../utils/logger')
const config = require('../config')

class AuthController {
  /**
   * Inicia sesión de usuario
   */
  async login(request, reply) {
    try {
      const { rut, password } = request.body

      // Validar datos requeridos
      if (!rut || !password) {
        return reply.code(400).send({
          code: 'MISSING_CREDENTIALS',
          error: 'Bad Request',
          message: 'RUT y contraseña son requeridos'
        })
      }

      // Autenticar usuario
      const user = await Usuario.authenticate(rut, password)

      // Generar token JWT
      const token = await reply.jwtSign(
        { 
          id: user.id,
          rut: user.rut,
          rol: user.rol,
          sucursal_id: user.sucursal_id
        },
        { 
          expiresIn: config.auth.jwtExpiresIn 
        }
      )

      // Obtener perfil completo del usuario
      const profile = await Usuario.getProfile(user.id)

      logger.audit('Login exitoso', {
        userId: user.id,
        rut: user.rut,
        rol: user.rol,
        ip: request.ip,
        userAgent: request.headers['user-agent']
      })

      reply.send({
        success: true,
        message: 'Login exitoso',
        data: {
          token,
          user: profile,
          expiresIn: config.auth.jwtExpiresIn
        }
      })
    } catch (error) {
      logger.error('Error en login:', error)
      
      // Determinar código de error apropiado
      let statusCode = 500
      let errorCode = 'INTERNAL_ERROR'
      
      if (error.message.includes('Credenciales inválidas')) {
        statusCode = 401
        errorCode = 'INVALID_CREDENTIALS'
      } else if (error.message.includes('bloqueado')) {
        statusCode = 423
        errorCode = 'ACCOUNT_LOCKED'
      }

      reply.code(statusCode).send({
        code: errorCode,
        error: statusCode === 500 ? 'Internal Server Error' : 'Authentication Failed',
        message: statusCode === 500 ? 'Error interno del servidor' : error.message
      })
    }
  }

  /**
   * Cierra sesión de usuario
   */
  async logout(request, reply) {
    try {
      const userId = request.user.id

      logger.audit('Logout exitoso', {
        userId,
        ip: request.ip
      })

      reply.send({
        success: true,
        message: 'Sesión cerrada exitosamente'
      })
    } catch (error) {
      logger.error('Error en logout:', error)
      reply.code(500).send({
        code: 'INTERNAL_ERROR',
        error: 'Internal Server Error',
        message: 'Error interno del servidor'
      })
    }
  }

  /**
   * Renueva el token JWT
   */
  async refreshToken(request, reply) {
    try {
      const user = request.user

      // Verificar que el usuario sigue activo
      const currentUser = await Usuario.findById(user.id)
      if (!currentUser || !currentUser.activo) {
        return reply.code(401).send({
          code: 'USER_INACTIVE',
          error: 'Unauthorized',
          message: 'Usuario inactivo o no encontrado'
        })
      }

      // Generar nuevo token
      const newToken = await reply.jwtSign(
        { 
          id: user.id,
          rut: user.rut,
          rol: user.rol,
          sucursal_id: user.sucursal_id
        },
        { 
          expiresIn: config.auth.jwtExpiresIn 
        }
      )

      logger.audit('Token renovado', {
        userId: user.id,
        ip: request.ip
      })

      reply.send({
        success: true,
        message: 'Token renovado exitosamente',
        data: {
          token: newToken,
          expiresIn: config.auth.jwtExpiresIn
        }
      })
    } catch (error) {
      logger.error('Error al renovar token:', error)
      reply.code(500).send({
        code: 'INTERNAL_ERROR',
        error: 'Internal Server Error',
        message: 'Error interno del servidor'
      })
    }
  }

  /**
   * Obtiene el perfil del usuario autenticado
   */
  async getProfile(request, reply) {
    try {
      const userId = request.user.id
      const profile = await Usuario.getProfile(userId)

      reply.send({
        success: true,
        data: profile
      })
    } catch (error) {
      logger.error('Error al obtener perfil:', error)
      reply.code(500).send({
        code: 'INTERNAL_ERROR',
        error: 'Internal Server Error',
        message: 'Error interno del servidor'
      })
    }
  }

  /**
   * Cambia la contraseña del usuario autenticado
   */
  async changePassword(request, reply) {
    try {
      const { currentPassword, newPassword, confirmPassword } = request.body
      const userId = request.user.id

      // Validar datos requeridos
      if (!currentPassword || !newPassword || !confirmPassword) {
        return reply.code(400).send({
          code: 'MISSING_DATA',
          error: 'Bad Request',
          message: 'Todos los campos son requeridos'
        })
      }

      // Validar que las contraseñas coincidan
      if (newPassword !== confirmPassword) {
        return reply.code(400).send({
          code: 'PASSWORD_MISMATCH',
          error: 'Bad Request',
          message: 'Las contraseñas nuevas no coinciden'
        })
      }

      // Validar longitud mínima
      if (newPassword.length < 6) {
        return reply.code(400).send({
          code: 'PASSWORD_TOO_SHORT',
          error: 'Bad Request',
          message: 'La contraseña debe tener al menos 6 caracteres'
        })
      }

      // Cambiar contraseña
      await Usuario.changePassword(userId, currentPassword, newPassword)

      logger.audit('Contraseña cambiada', {
        userId,
        ip: request.ip
      })

      reply.send({
        success: true,
        message: 'Contraseña cambiada exitosamente'
      })
    } catch (error) {
      logger.error('Error al cambiar contraseña:', error)
      
      let statusCode = 500
      let errorCode = 'INTERNAL_ERROR'
      
      if (error.message.includes('Contraseña actual incorrecta')) {
        statusCode = 400
        errorCode = 'INVALID_CURRENT_PASSWORD'
      }

      reply.code(statusCode).send({
        code: errorCode,
        error: statusCode === 500 ? 'Internal Server Error' : 'Bad Request',
        message: statusCode === 500 ? 'Error interno del servidor' : error.message
      })
    }
  }

  /**
   * Verifica si el token es válido
   */
  async verifyToken(request, reply) {
    try {
      const user = request.user

      reply.send({
        success: true,
        message: 'Token válido',
        data: {
          valid: true,
          user: {
            id: user.id,
            rut: user.rut,
            nombre: user.nombre,
            rol: user.rol,
            sucursal_id: user.sucursal_id
          }
        }
      })
    } catch (error) {
      logger.error('Error al verificar token:', error)
      reply.code(500).send({
        code: 'INTERNAL_ERROR',
        error: 'Internal Server Error',
        message: 'Error interno del servidor'
      })
    }
  }

  /**
   * Obtiene información de la sesión actual
   */
  async getSessionInfo(request, reply) {
    try {
      const user = request.user

      // Obtener estadísticas de la sesión
      const sessionInfo = {
        user: {
          id: user.id,
          rut: user.rut,
          nombre: user.nombre,
          apellido: user.apellido,
          rol: user.rol,
          sucursal_id: user.sucursal_id,
          ultimo_acceso: user.ultimo_acceso
        },
        session: {
          ip: request.ip,
          userAgent: request.headers['user-agent'],
          loginTime: new Date().toISOString()
        }
      }

      reply.send({
        success: true,
        data: sessionInfo
      })
    } catch (error) {
      logger.error('Error al obtener información de sesión:', error)
      reply.code(500).send({
        code: 'INTERNAL_ERROR',
        error: 'Internal Server Error',
        message: 'Error interno del servidor'
      })
    }
  }

  /**
   * Resetea la contraseña de un usuario (solo admin)
   */
  async resetPassword(request, reply) {
    try {
      const { userId, newPassword } = request.body
      const adminUserId = request.user.id

      // Verificar que el usuario actual es admin
      if (request.user.rol !== 'admin') {
        return reply.code(403).send({
          code: 'INSUFFICIENT_PERMISSIONS',
          error: 'Forbidden',
          message: 'Solo los administradores pueden resetear contraseñas'
        })
      }

      // Validar datos requeridos
      if (!userId || !newPassword) {
        return reply.code(400).send({
          code: 'MISSING_DATA',
          error: 'Bad Request',
          message: 'ID de usuario y nueva contraseña son requeridos'
        })
      }

      // Validar longitud mínima
      if (newPassword.length < 6) {
        return reply.code(400).send({
          code: 'PASSWORD_TOO_SHORT',
          error: 'Bad Request',
          message: 'La contraseña debe tener al menos 6 caracteres'
        })
      }

      // Resetear contraseña
      await Usuario.resetPassword(userId, newPassword, adminUserId)

      logger.audit('Contraseña reseteada por administrador', {
        targetUserId: userId,
        adminUserId,
        ip: request.ip
      })

      reply.send({
        success: true,
        message: 'Contraseña reseteada exitosamente'
      })
    } catch (error) {
      logger.error('Error al resetear contraseña:', error)
      reply.code(500).send({
        code: 'INTERNAL_ERROR',
        error: 'Internal Server Error',
        message: 'Error interno del servidor'
      })
    }
  }

  /**
   * Obtiene estadísticas de autenticación (solo admin)
   */
  async getAuthStats(request, reply) {
    try {
      // Verificar permisos de administrador
      if (request.user.rol !== 'admin') {
        return reply.code(403).send({
          code: 'INSUFFICIENT_PERMISSIONS',
          error: 'Forbidden',
          message: 'Solo los administradores pueden ver estadísticas'
        })
      }

      const stats = await Usuario.getUserStats()

      reply.send({
        success: true,
        data: stats
      })
    } catch (error) {
      logger.error('Error al obtener estadísticas de autenticación:', error)
      reply.code(500).send({
        code: 'INTERNAL_ERROR',
        error: 'Internal Server Error',
        message: 'Error interno del servidor'
      })
    }
  }
}

module.exports = new AuthController()

