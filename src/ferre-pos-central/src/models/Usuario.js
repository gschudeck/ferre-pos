/**
 * Modelo Usuario - Sistema Ferre-POS
 * 
 * Maneja todas las operaciones relacionadas con usuarios del sistema,
 * incluyendo autenticación, autorización y gestión de sesiones.
 */

const BaseModel = require('./BaseModel')
const bcrypt = require('bcryptjs')
const logger = require('../utils/logger')
const config = require('../config')

class Usuario extends BaseModel {
  constructor() {
    super('usuarios', {
      rut: { type: 'string', required: true, maxLength: 12 },
      nombre: { type: 'string', required: true, maxLength: 100 },
      apellido: { type: 'string', maxLength: 100 },
      email: { type: 'string', maxLength: 255 },
      telefono: { type: 'string', maxLength: 20 },
      rol: { type: 'string', required: true },
      sucursal_id: { type: 'string' },
      password_hash: { type: 'string', required: true },
      salt: { type: 'string', required: true }
    })
  }

  /**
   * Crea un nuevo usuario con contraseña hasheada
   */
  async create(userData, options = {}) {
    try {
      // Validar RUT único
      const existingUser = await this.findByRut(userData.rut)
      if (existingUser) {
        throw new Error('Ya existe un usuario con este RUT')
      }

      // Validar email único si se proporciona
      if (userData.email) {
        const existingEmail = await this.findByEmail(userData.email)
        if (existingEmail) {
          throw new Error('Ya existe un usuario con este email')
        }
      }

      // Generar salt y hash de la contraseña
      const salt = await bcrypt.genSalt(config.auth.bcryptRounds)
      const passwordHash = await bcrypt.hash(userData.password, salt)

      // Preparar datos del usuario
      const userToCreate = {
        ...userData,
        password_hash: passwordHash,
        salt: salt,
        activo: true,
        intentos_fallidos: 0
      }

      // Remover la contraseña en texto plano
      delete userToCreate.password

      const newUser = await super.create(userToCreate, options)

      logger.audit('Usuario creado', {
        userId: newUser.id,
        rut: newUser.rut,
        rol: newUser.rol,
        sucursalId: newUser.sucursal_id
      })

      // Remover datos sensibles antes de retornar
      return this.sanitizeUser(newUser)
    } catch (error) {
      logger.error('Error al crear usuario:', error)
      throw error
    }
  }

  /**
   * Busca un usuario por RUT
   */
  async findByRut(rut, options = {}) {
    try {
      return await this.findOne({ rut }, options)
    } catch (error) {
      logger.error('Error al buscar usuario por RUT:', error)
      throw error
    }
  }

  /**
   * Busca un usuario por email
   */
  async findByEmail(email, options = {}) {
    try {
      return await this.findOne({ email }, options)
    } catch (error) {
      logger.error('Error al buscar usuario por email:', error)
      throw error
    }
  }

  /**
   * Autentica un usuario con RUT y contraseña
   */
  async authenticate(rut, password) {
    try {
      const user = await this.findByRut(rut, { includeInactive: false })
      
      if (!user) {
        logger.security('Intento de login con RUT inexistente', { rut })
        throw new Error('Credenciales inválidas')
      }

      // Verificar si el usuario está bloqueado
      if (user.bloqueado_hasta && new Date() < new Date(user.bloqueado_hasta)) {
        const tiempoRestante = Math.ceil((new Date(user.bloqueado_hasta) - new Date()) / 60000)
        logger.security('Intento de login con usuario bloqueado', {
          userId: user.id,
          rut: user.rut,
          tiempoRestante
        })
        throw new Error(`Usuario bloqueado. Intente nuevamente en ${tiempoRestante} minutos`)
      }

      // Verificar contraseña
      const isValidPassword = await bcrypt.compare(password, user.password_hash)
      
      if (!isValidPassword) {
        await this.incrementFailedAttempts(user.id)
        logger.security('Intento de login con contraseña incorrecta', {
          userId: user.id,
          rut: user.rut,
          intentos: user.intentos_fallidos + 1
        })
        throw new Error('Credenciales inválidas')
      }

      // Reset intentos fallidos en login exitoso
      await this.resetFailedAttempts(user.id)

      // Actualizar último acceso
      await this.updateById(user.id, {
        ultimo_acceso: new Date()
      })

      logger.audit('Login exitoso', {
        userId: user.id,
        rut: user.rut,
        rol: user.rol
      })

      return this.sanitizeUser(user)
    } catch (error) {
      logger.error('Error en autenticación:', error)
      throw error
    }
  }

  /**
   * Incrementa los intentos fallidos de login
   */
  async incrementFailedAttempts(userId) {
    try {
      const user = await this.findById(userId)
      const newAttempts = (user.intentos_fallidos || 0) + 1
      
      const updateData = {
        intentos_fallidos: newAttempts
      }

      // Bloquear usuario si excede el máximo de intentos
      if (newAttempts >= config.auth.maxLoginAttempts) {
        updateData.bloqueado_hasta = new Date(Date.now() + config.auth.lockoutDuration)
        
        logger.security('Usuario bloqueado por exceso de intentos fallidos', {
          userId,
          intentos: newAttempts,
          bloqueadoHasta: updateData.bloqueado_hasta
        })
      }

      await this.updateById(userId, updateData)
    } catch (error) {
      logger.error('Error al incrementar intentos fallidos:', error)
      throw error
    }
  }

  /**
   * Resetea los intentos fallidos de login
   */
  async resetFailedAttempts(userId) {
    try {
      await this.updateById(userId, {
        intentos_fallidos: 0,
        bloqueado_hasta: null
      })
    } catch (error) {
      logger.error('Error al resetear intentos fallidos:', error)
      throw error
    }
  }

  /**
   * Cambia la contraseña de un usuario
   */
  async changePassword(userId, currentPassword, newPassword) {
    try {
      const user = await this.findById(userId)
      
      if (!user) {
        throw new Error('Usuario no encontrado')
      }

      // Verificar contraseña actual
      const isValidPassword = await bcrypt.compare(currentPassword, user.password_hash)
      if (!isValidPassword) {
        logger.security('Intento de cambio de contraseña con contraseña actual incorrecta', {
          userId
        })
        throw new Error('Contraseña actual incorrecta')
      }

      // Generar nuevo hash
      const salt = await bcrypt.genSalt(config.auth.bcryptRounds)
      const passwordHash = await bcrypt.hash(newPassword, salt)

      await this.updateById(userId, {
        password_hash: passwordHash,
        salt: salt
      })

      logger.audit('Contraseña cambiada', { userId })

      return true
    } catch (error) {
      logger.error('Error al cambiar contraseña:', error)
      throw error
    }
  }

  /**
   * Resetea la contraseña de un usuario (solo admin)
   */
  async resetPassword(userId, newPassword, adminUserId) {
    try {
      const salt = await bcrypt.genSalt(config.auth.bcryptRounds)
      const passwordHash = await bcrypt.hash(newPassword, salt)

      await this.updateById(userId, {
        password_hash: passwordHash,
        salt: salt,
        intentos_fallidos: 0,
        bloqueado_hasta: null
      })

      logger.audit('Contraseña reseteada por administrador', {
        userId,
        adminUserId
      })

      return true
    } catch (error) {
      logger.error('Error al resetear contraseña:', error)
      throw error
    }
  }

  /**
   * Obtiene usuarios por sucursal
   */
  async findBySucursal(sucursalId, options = {}) {
    try {
      return await this.findAll({
        where: { sucursal_id: sucursalId },
        ...options
      })
    } catch (error) {
      logger.error('Error al buscar usuarios por sucursal:', error)
      throw error
    }
  }

  /**
   * Obtiene usuarios por rol
   */
  async findByRol(rol, options = {}) {
    try {
      return await this.findAll({
        where: { rol },
        ...options
      })
    } catch (error) {
      logger.error('Error al buscar usuarios por rol:', error)
      throw error
    }
  }

  /**
   * Verifica si un usuario tiene un rol específico
   */
  async hasRole(userId, role) {
    try {
      const user = await this.findById(userId)
      return user && user.rol === role
    } catch (error) {
      logger.error('Error al verificar rol de usuario:', error)
      throw error
    }
  }

  /**
   * Verifica si un usuario pertenece a una sucursal específica
   */
  async belongsToSucursal(userId, sucursalId) {
    try {
      const user = await this.findById(userId)
      return user && (user.sucursal_id === sucursalId || user.rol === 'admin')
    } catch (error) {
      logger.error('Error al verificar sucursal de usuario:', error)
      throw error
    }
  }

  /**
   * Obtiene estadísticas de usuarios
   */
  async getUserStats() {
    try {
      const baseStats = await this.getStats()
      
      // Estadísticas por rol
      const roleStatsQuery = `
        SELECT rol, COUNT(*) as count
        FROM usuarios
        WHERE activo = true
        GROUP BY rol
        ORDER BY count DESC
      `
      const roleStatsResult = await this.query(roleStatsQuery)
      
      // Estadísticas por sucursal
      const sucursalStatsQuery = `
        SELECT s.nombre as sucursal, COUNT(u.id) as count
        FROM usuarios u
        JOIN sucursales s ON u.sucursal_id = s.id
        WHERE u.activo = true AND s.habilitada = true
        GROUP BY s.id, s.nombre
        ORDER BY count DESC
      `
      const sucursalStatsResult = await this.query(sucursalStatsQuery)

      // Usuarios activos en las últimas 24 horas
      const activeUsersQuery = `
        SELECT COUNT(*) as active_users
        FROM usuarios
        WHERE ultimo_acceso >= NOW() - INTERVAL '24 hours'
        AND activo = true
      `
      const activeUsersResult = await this.query(activeUsersQuery)

      return {
        ...baseStats,
        byRole: roleStatsResult.rows,
        bySucursal: sucursalStatsResult.rows,
        activeIn24h: parseInt(activeUsersResult.rows[0].active_users)
      }
    } catch (error) {
      logger.error('Error al obtener estadísticas de usuarios:', error)
      throw error
    }
  }

  /**
   * Busca usuarios con filtros avanzados
   */
  async search(filters = {}, options = {}) {
    try {
      const {
        rut,
        nombre,
        email,
        rol,
        sucursalId,
        activo = true
      } = filters

      let query = `
        SELECT u.*, s.nombre as sucursal_nombre
        FROM usuarios u
        LEFT JOIN sucursales s ON u.sucursal_id = s.id
        WHERE 1=1
      `
      const params = []

      if (activo !== undefined) {
        query += ` AND u.activo = $${params.length + 1}`
        params.push(activo)
      }

      if (rut) {
        query += ` AND u.rut ILIKE $${params.length + 1}`
        params.push(`%${rut}%`)
      }

      if (nombre) {
        query += ` AND (u.nombre ILIKE $${params.length + 1} OR u.apellido ILIKE $${params.length + 1})`
        params.push(`%${nombre}%`)
      }

      if (email) {
        query += ` AND u.email ILIKE $${params.length + 1}`
        params.push(`%${email}%`)
      }

      if (rol) {
        query += ` AND u.rol = $${params.length + 1}`
        params.push(rol)
      }

      if (sucursalId) {
        query += ` AND u.sucursal_id = $${params.length + 1}`
        params.push(sucursalId)
      }

      const { orderBy = 'nombre', orderDirection = 'ASC' } = options
      query += ` ORDER BY u.${orderBy} ${orderDirection}`

      const result = await this.query(query, params)
      return result.rows.map(user => this.sanitizeUser(user))
    } catch (error) {
      logger.error('Error en búsqueda de usuarios:', error)
      throw error
    }
  }

  /**
   * Elimina datos sensibles del objeto usuario
   */
  sanitizeUser(user) {
    if (!user) return null
    
    const sanitized = { ...user }
    delete sanitized.password_hash
    delete sanitized.salt
    delete sanitized.intentos_fallidos
    delete sanitized.bloqueado_hasta
    
    return sanitized
  }

  /**
   * Obtiene el perfil completo de un usuario
   */
  async getProfile(userId) {
    try {
      const query = `
        SELECT 
          u.*,
          s.nombre as sucursal_nombre,
          s.direccion as sucursal_direccion
        FROM usuarios u
        LEFT JOIN sucursales s ON u.sucursal_id = s.id
        WHERE u.id = $1 AND u.activo = true
      `
      
      const result = await this.query(query, [userId])
      
      if (!result.rows.length) {
        throw new Error('Usuario no encontrado')
      }

      return this.sanitizeUser(result.rows[0])
    } catch (error) {
      logger.error('Error al obtener perfil de usuario:', error)
      throw error
    }
  }
}

module.exports = new Usuario()

