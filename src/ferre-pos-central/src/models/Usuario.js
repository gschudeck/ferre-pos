/**
 * Modelo Usuario - Sistema Ferre-POS
 * 
 * Maneja todas las operaciones relacionadas con usuarios,
 * autenticación, autorización, perfiles y auditoría de accesos.
 */

const BaseModel = require('./BaseModel')
const logger = require('../utils/logger')
const bcrypt = require('bcrypt')
const crypto = require('crypto')

class Usuario extends BaseModel {
  constructor() {
    super('usuarios', {
      rut: { type: 'string', required: true },
      nombre: { type: 'string', required: true },
      email: { type: 'string', required: true },
      telefono: { type: 'string' },
      rol: { type: 'string', required: true },
      sucursal_id: { type: 'uuid' },
      password_hash: { type: 'string', required: true },
      salt: { type: 'string', required: true },
      activo: { type: 'boolean', required: true },
      ultimo_acceso: { type: 'timestamp' },
      intentos_fallidos: { type: 'integer' },
      bloqueado_hasta: { type: 'timestamp' },
      debe_cambiar_password: { type: 'boolean' },
      token_recuperacion: { type: 'string' },
      token_recuperacion_expira: { type: 'timestamp' }
    })
  }

  /**
   * Crea un nuevo usuario con validaciones completas
   */
  async createUsuario(userData, creadorId) {
    try {
      return await this.transaction(async (client) => {
        const {
          rut,
          nombre,
          email,
          telefono,
          rol,
          sucursal_id,
          password,
          debe_cambiar_password = true
        } = userData

        // Validar que el RUT no existe
        const existeRut = await client.query(
          'SELECT id FROM usuarios WHERE rut = $1',
          [rut]
        )
        
        if (existeRut.rows.length > 0) {
          throw new Error(`Ya existe un usuario con RUT ${rut}`)
        }

        // Validar que el email no existe
        const existeEmail = await client.query(
          'SELECT id FROM usuarios WHERE email = $1',
          [email]
        )
        
        if (existeEmail.rows.length > 0) {
          throw new Error(`Ya existe un usuario con email ${email}`)
        }

        // Validar rol
        const rolesValidos = ['admin', 'gerente', 'vendedor', 'cajero']
        if (!rolesValidos.includes(rol)) {
          throw new Error(`Rol '${rol}' no es válido`)
        }

        // Validar sucursal si se especifica
        if (sucursal_id) {
          const sucursal = await client.query(
            'SELECT id FROM sucursales WHERE id = $1 AND habilitada = true',
            [sucursal_id]
          )
          
          if (!sucursal.rows.length) {
            throw new Error('Sucursal no encontrada o no habilitada')
          }
        }

        // Generar hash de contraseña
        const { hash, salt } = await this.hashPassword(password)

        // Crear usuario
        const query = `
          INSERT INTO usuarios (
            rut, nombre, email, telefono, rol, sucursal_id,
            password_hash, salt, activo, debe_cambiar_password,
            intentos_fallidos, usuario_creacion
          ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, true, $9, 0, $10)
          RETURNING id, rut, nombre, email, telefono, rol, sucursal_id, 
                   activo, debe_cambiar_password, fecha_creacion
        `

        const result = await client.query(query, [
          rut, nombre, email, telefono, rol, sucursal_id,
          hash, salt, debe_cambiar_password, creadorId
        ])

        const usuario = result.rows[0]

        // Registrar en auditoría
        await this.registrarAuditoria(
          usuario.id,
          'USUARIO_CREADO',
          { rut, nombre, rol, sucursal_id },
          creadorId,
          client
        )

        logger.business('Usuario creado', {
          usuarioId: usuario.id,
          rut: usuario.rut,
          nombre: usuario.nombre,
          rol: usuario.rol,
          creadorId
        })

        return usuario
      })
    } catch (error) {
      logger.error('Error al crear usuario:', error)
      throw error
    }
  }

  /**
   * Autentica un usuario con RUT y contraseña
   */
  async autenticarUsuario(rut, password, ipAddress = null) {
    try {
      return await this.transaction(async (client) => {
        // Buscar usuario
        const query = `
          SELECT id, rut, nombre, email, rol, sucursal_id, password_hash, salt,
                 activo, ultimo_acceso, intentos_fallidos, bloqueado_hasta,
                 debe_cambiar_password
          FROM usuarios
          WHERE rut = $1
        `
        
        const result = await client.query(query, [rut])
        
        if (!result.rows.length) {
          await this.registrarIntentoAcceso(rut, false, 'USUARIO_NO_ENCONTRADO', ipAddress)
          throw new Error('Credenciales inválidas')
        }

        const usuario = result.rows[0]

        // Verificar si el usuario está activo
        if (!usuario.activo) {
          await this.registrarIntentoAcceso(rut, false, 'USUARIO_INACTIVO', ipAddress)
          throw new Error('Usuario inactivo')
        }

        // Verificar si el usuario está bloqueado
        if (usuario.bloqueado_hasta && new Date() < usuario.bloqueado_hasta) {
          await this.registrarIntentoAcceso(rut, false, 'USUARIO_BLOQUEADO', ipAddress)
          const tiempoRestante = Math.ceil((usuario.bloqueado_hasta - new Date()) / 60000)
          throw new Error(`Usuario bloqueado. Intente nuevamente en ${tiempoRestante} minutos`)
        }

        // Verificar contraseña
        const passwordValida = await this.verificarPassword(password, usuario.password_hash, usuario.salt)
        
        if (!passwordValida) {
          // Incrementar intentos fallidos
          const nuevosIntentos = usuario.intentos_fallidos + 1
          let bloqueadoHasta = null
          
          // Bloquear después de 5 intentos fallidos
          if (nuevosIntentos >= 5) {
            bloqueadoHasta = new Date(Date.now() + 30 * 60 * 1000) // 30 minutos
          }

          await client.query(`
            UPDATE usuarios 
            SET intentos_fallidos = $1, bloqueado_hasta = $2
            WHERE id = $3
          `, [nuevosIntentos, bloqueadoHasta, usuario.id])

          await this.registrarIntentoAcceso(rut, false, 'PASSWORD_INCORRECTA', ipAddress)
          throw new Error('Credenciales inválidas')
        }

        // Autenticación exitosa - resetear intentos fallidos y actualizar último acceso
        await client.query(`
          UPDATE usuarios 
          SET intentos_fallidos = 0, bloqueado_hasta = NULL, ultimo_acceso = NOW()
          WHERE id = $1
        `, [usuario.id])

        await this.registrarIntentoAcceso(rut, true, 'LOGIN_EXITOSO', ipAddress)

        // Registrar en auditoría
        await this.registrarAuditoria(
          usuario.id,
          'LOGIN_EXITOSO',
          { ip_address: ipAddress },
          usuario.id,
          client
        )

        logger.business('Login exitoso', {
          usuarioId: usuario.id,
          rut: usuario.rut,
          nombre: usuario.nombre,
          rol: usuario.rol,
          ipAddress
        })

        // Retornar datos del usuario sin información sensible
        return {
          id: usuario.id,
          rut: usuario.rut,
          nombre: usuario.nombre,
          email: usuario.email,
          rol: usuario.rol,
          sucursal_id: usuario.sucursal_id,
          debe_cambiar_password: usuario.debe_cambiar_password,
          ultimo_acceso: usuario.ultimo_acceso
        }
      })
    } catch (error) {
      logger.error('Error en autenticación:', error)
      throw error
    }
  }

  /**
   * Obtiene un usuario por ID con información completa
   */
  async getUsuarioCompleto(id) {
    try {
      const query = `
        SELECT u.id, u.rut, u.nombre, u.email, u.telefono, u.rol, u.sucursal_id,
               u.activo, u.ultimo_acceso, u.debe_cambiar_password, u.fecha_creacion,
               u.fecha_modificacion, s.nombre as sucursal_nombre,
               uc.nombre as creador_nombre
        FROM usuarios u
        LEFT JOIN sucursales s ON u.sucursal_id = s.id
        LEFT JOIN usuarios uc ON u.usuario_creacion = uc.id
        WHERE u.id = $1
      `
      
      const result = await this.query(query, [id])
      
      if (!result.rows.length) {
        throw new Error('Usuario no encontrado')
      }

      return result.rows[0]
    } catch (error) {
      logger.error('Error al obtener usuario completo:', error)
      throw error
    }
  }

  /**
   * Obtiene lista de usuarios con filtros y paginación
   */
  async getUsuarios(options = {}) {
    try {
      const {
        sucursalId = null,
        rol = null,
        activo = null,
        busqueda = null,
        page = 1,
        limit = 20,
        orderBy = 'nombre',
        orderDirection = 'ASC'
      } = options

      let query = `
        SELECT u.id, u.rut, u.nombre, u.email, u.telefono, u.rol, u.sucursal_id,
               u.activo, u.ultimo_acceso, u.fecha_creacion,
               s.nombre as sucursal_nombre
        FROM usuarios u
        LEFT JOIN sucursales s ON u.sucursal_id = s.id
        WHERE 1=1
      `
      
      const params = []
      let paramIndex = 1

      if (sucursalId) {
        query += ` AND u.sucursal_id = $${paramIndex}`
        params.push(sucursalId)
        paramIndex++
      }

      if (rol) {
        query += ` AND u.rol = $${paramIndex}`
        params.push(rol)
        paramIndex++
      }

      if (activo !== null) {
        query += ` AND u.activo = $${paramIndex}`
        params.push(activo)
        paramIndex++
      }

      if (busqueda) {
        query += ` AND (
          u.nombre ILIKE $${paramIndex} OR 
          u.email ILIKE $${paramIndex} OR 
          u.rut ILIKE $${paramIndex}
        )`
        params.push(`%${busqueda}%`)
        paramIndex++
      }

      // Contar total de registros
      const countQuery = query.replace(/SELECT.*FROM/, 'SELECT COUNT(*) FROM')
      const countResult = await this.query(countQuery, params)
      const total = parseInt(countResult.rows[0].count)

      // Agregar ordenamiento y paginación
      const validOrderBy = ['nombre', 'email', 'rol', 'fecha_creacion', 'ultimo_acceso']
      const orderByField = validOrderBy.includes(orderBy) ? orderBy : 'nombre'
      const direction = orderDirection.toUpperCase() === 'DESC' ? 'DESC' : 'ASC'
      
      query += ` ORDER BY u.${orderByField} ${direction}`
      query += ` LIMIT $${paramIndex} OFFSET $${paramIndex + 1}`
      params.push(limit, (page - 1) * limit)

      const result = await this.query(query, params)

      return {
        data: result.rows,
        pagination: {
          page,
          limit,
          total,
          totalPages: Math.ceil(total / limit)
        }
      }
    } catch (error) {
      logger.error('Error al obtener usuarios:', error)
      throw error
    }
  }

  /**
   * Actualiza un usuario
   */
  async updateUsuario(id, updateData, modificadorId) {
    try {
      return await this.transaction(async (client) => {
        // Verificar que el usuario existe
        const usuarioExistente = await client.query(
          'SELECT * FROM usuarios WHERE id = $1',
          [id]
        )
        
        if (!usuarioExistente.rows.length) {
          throw new Error('Usuario no encontrado')
        }

        const usuario = usuarioExistente.rows[0]
        const cambios = {}

        // Construir query de actualización dinámicamente
        const setClauses = []
        const params = []
        let paramIndex = 1

        // Campos que se pueden actualizar
        const camposPermitidos = ['nombre', 'email', 'telefono', 'rol', 'sucursal_id', 'activo']
        
        for (const campo of camposPermitidos) {
          if (updateData[campo] !== undefined && updateData[campo] !== usuario[campo]) {
            setClauses.push(`${campo} = $${paramIndex}`)
            params.push(updateData[campo])
            cambios[campo] = { anterior: usuario[campo], nuevo: updateData[campo] }
            paramIndex++
          }
        }

        if (setClauses.length === 0) {
          throw new Error('No hay cambios para actualizar')
        }

        // Validaciones específicas
        if (updateData.email && updateData.email !== usuario.email) {
          const existeEmail = await client.query(
            'SELECT id FROM usuarios WHERE email = $1 AND id != $2',
            [updateData.email, id]
          )
          
          if (existeEmail.rows.length > 0) {
            throw new Error('Ya existe un usuario con ese email')
          }
        }

        if (updateData.rol) {
          const rolesValidos = ['admin', 'gerente', 'vendedor', 'cajero']
          if (!rolesValidos.includes(updateData.rol)) {
            throw new Error(`Rol '${updateData.rol}' no es válido`)
          }
        }

        if (updateData.sucursal_id) {
          const sucursal = await client.query(
            'SELECT id FROM sucursales WHERE id = $1 AND habilitada = true',
            [updateData.sucursal_id]
          )
          
          if (!sucursal.rows.length) {
            throw new Error('Sucursal no encontrada o no habilitada')
          }
        }

        // Ejecutar actualización
        setClauses.push(`fecha_modificacion = NOW()`)
        setClauses.push(`usuario_modificacion = $${paramIndex}`)
        params.push(modificadorId)
        params.push(id)

        const query = `
          UPDATE usuarios 
          SET ${setClauses.join(', ')}
          WHERE id = $${paramIndex + 1}
          RETURNING id, rut, nombre, email, telefono, rol, sucursal_id, activo
        `

        const result = await client.query(query, params)

        // Registrar en auditoría
        await this.registrarAuditoria(
          id,
          'USUARIO_ACTUALIZADO',
          cambios,
          modificadorId,
          client
        )

        logger.business('Usuario actualizado', {
          usuarioId: id,
          cambios,
          modificadorId
        })

        return result.rows[0]
      })
    } catch (error) {
      logger.error('Error al actualizar usuario:', error)
      throw error
    }
  }

  /**
   * Cambia la contraseña de un usuario
   */
  async cambiarPassword(id, passwordActual, passwordNueva, cambiadoPorId = null) {
    try {
      return await this.transaction(async (client) => {
        // Obtener usuario
        const result = await client.query(
          'SELECT password_hash, salt, debe_cambiar_password FROM usuarios WHERE id = $1',
          [id]
        )
        
        if (!result.rows.length) {
          throw new Error('Usuario no encontrado')
        }

        const usuario = result.rows[0]

        // Si no es un cambio forzado por admin, verificar contraseña actual
        if (!cambiadoPorId || cambiadoPorId === id) {
          const passwordValida = await this.verificarPassword(
            passwordActual, 
            usuario.password_hash, 
            usuario.salt
          )
          
          if (!passwordValida) {
            throw new Error('Contraseña actual incorrecta')
          }
        }

        // Validar nueva contraseña
        this.validarPassword(passwordNueva)

        // Generar nuevo hash
        const { hash, salt } = await this.hashPassword(passwordNueva)

        // Actualizar contraseña
        await client.query(`
          UPDATE usuarios 
          SET password_hash = $1, salt = $2, debe_cambiar_password = false,
              fecha_modificacion = NOW()
          WHERE id = $3
        `, [hash, salt, id])

        // Registrar en auditoría
        await this.registrarAuditoria(
          id,
          'PASSWORD_CAMBIADA',
          { cambio_forzado: cambiadoPorId && cambiadoPorId !== id },
          cambiadoPorId || id,
          client
        )

        logger.business('Contraseña cambiada', {
          usuarioId: id,
          cambiadoPorId: cambiadoPorId || id,
          forzado: cambiadoPorId && cambiadoPorId !== id
        })

        return { success: true }
      })
    } catch (error) {
      logger.error('Error al cambiar contraseña:', error)
      throw error
    }
  }

  /**
   * Inicia el proceso de recuperación de contraseña
   */
  async iniciarRecuperacionPassword(email) {
    try {
      return await this.transaction(async (client) => {
        // Buscar usuario por email
        const result = await client.query(
          'SELECT id, rut, nombre, email, activo FROM usuarios WHERE email = $1',
          [email]
        )
        
        if (!result.rows.length) {
          // Por seguridad, no revelar si el email existe o no
          return { success: true, message: 'Si el email existe, recibirá instrucciones' }
        }

        const usuario = result.rows[0]

        if (!usuario.activo) {
          throw new Error('Usuario inactivo')
        }

        // Generar token de recuperación
        const token = crypto.randomBytes(32).toString('hex')
        const expira = new Date(Date.now() + 60 * 60 * 1000) // 1 hora

        // Guardar token
        await client.query(`
          UPDATE usuarios 
          SET token_recuperacion = $1, token_recuperacion_expira = $2
          WHERE id = $3
        `, [token, expira, usuario.id])

        // Registrar en auditoría
        await this.registrarAuditoria(
          usuario.id,
          'RECUPERACION_PASSWORD_INICIADA',
          { email },
          usuario.id,
          client
        )

        logger.business('Recuperación de contraseña iniciada', {
          usuarioId: usuario.id,
          email: usuario.email
        })

        // En un sistema real, aquí se enviaría el email
        // Por ahora retornamos el token para testing
        return {
          success: true,
          message: 'Instrucciones enviadas al email',
          token: token // Solo para testing, remover en producción
        }
      })
    } catch (error) {
      logger.error('Error al iniciar recuperación de contraseña:', error)
      throw error
    }
  }

  /**
   * Completa el proceso de recuperación de contraseña
   */
  async completarRecuperacionPassword(token, passwordNueva) {
    try {
      return await this.transaction(async (client) => {
        // Buscar usuario por token válido
        const result = await client.query(`
          SELECT id, rut, nombre, email 
          FROM usuarios 
          WHERE token_recuperacion = $1 
            AND token_recuperacion_expira > NOW()
            AND activo = true
        `, [token])
        
        if (!result.rows.length) {
          throw new Error('Token inválido o expirado')
        }

        const usuario = result.rows[0]

        // Validar nueva contraseña
        this.validarPassword(passwordNueva)

        // Generar nuevo hash
        const { hash, salt } = await this.hashPassword(passwordNueva)

        // Actualizar contraseña y limpiar token
        await client.query(`
          UPDATE usuarios 
          SET password_hash = $1, salt = $2, debe_cambiar_password = false,
              token_recuperacion = NULL, token_recuperacion_expira = NULL,
              intentos_fallidos = 0, bloqueado_hasta = NULL,
              fecha_modificacion = NOW()
          WHERE id = $3
        `, [hash, salt, usuario.id])

        // Registrar en auditoría
        await this.registrarAuditoria(
          usuario.id,
          'PASSWORD_RECUPERADA',
          {},
          usuario.id,
          client
        )

        logger.business('Contraseña recuperada exitosamente', {
          usuarioId: usuario.id,
          email: usuario.email
        })

        return { success: true, message: 'Contraseña actualizada exitosamente' }
      })
    } catch (error) {
      logger.error('Error al completar recuperación de contraseña:', error)
      throw error
    }
  }

  /**
   * Desactiva un usuario (eliminación lógica)
   */
  async desactivarUsuario(id, motivo, desactivadoPorId) {
    try {
      return await this.transaction(async (client) => {
        // Verificar que el usuario existe y está activo
        const result = await client.query(
          'SELECT id, rut, nombre, activo FROM usuarios WHERE id = $1',
          [id]
        )
        
        if (!result.rows.length) {
          throw new Error('Usuario no encontrado')
        }

        const usuario = result.rows[0]

        if (!usuario.activo) {
          throw new Error('Usuario ya está inactivo')
        }

        // No permitir auto-desactivación
        if (id === desactivadoPorId) {
          throw new Error('No puede desactivar su propia cuenta')
        }

        // Desactivar usuario
        await client.query(`
          UPDATE usuarios 
          SET activo = false, fecha_modificacion = NOW()
          WHERE id = $1
        `, [id])

        // Registrar en auditoría
        await this.registrarAuditoria(
          id,
          'USUARIO_DESACTIVADO',
          { motivo },
          desactivadoPorId,
          client
        )

        logger.business('Usuario desactivado', {
          usuarioId: id,
          rut: usuario.rut,
          nombre: usuario.nombre,
          motivo,
          desactivadoPorId
        })

        return { success: true, message: 'Usuario desactivado exitosamente' }
      })
    } catch (error) {
      logger.error('Error al desactivar usuario:', error)
      throw error
    }
  }

  /**
   * Reactiva un usuario
   */
  async reactivarUsuario(id, reactivadoPorId) {
    try {
      return await this.transaction(async (client) => {
        // Verificar que el usuario existe y está inactivo
        const result = await client.query(
          'SELECT id, rut, nombre, activo FROM usuarios WHERE id = $1',
          [id]
        )
        
        if (!result.rows.length) {
          throw new Error('Usuario no encontrado')
        }

        const usuario = result.rows[0]

        if (usuario.activo) {
          throw new Error('Usuario ya está activo')
        }

        // Reactivar usuario y resetear bloqueos
        await client.query(`
          UPDATE usuarios 
          SET activo = true, intentos_fallidos = 0, bloqueado_hasta = NULL,
              debe_cambiar_password = true, fecha_modificacion = NOW()
          WHERE id = $1
        `, [id])

        // Registrar en auditoría
        await this.registrarAuditoria(
          id,
          'USUARIO_REACTIVADO',
          {},
          reactivadoPorId,
          client
        )

        logger.business('Usuario reactivado', {
          usuarioId: id,
          rut: usuario.rut,
          nombre: usuario.nombre,
          reactivadoPorId
        })

        return { success: true, message: 'Usuario reactivado exitosamente' }
      })
    } catch (error) {
      logger.error('Error al reactivar usuario:', error)
      throw error
    }
  }

  /**
   * Obtiene el historial de accesos de un usuario
   */
  async getHistorialAccesos(usuarioId, options = {}) {
    try {
      const {
        fechaInicio = null,
        fechaFin = null,
        exitosos = null,
        page = 1,
        limit = 50
      } = options

      let query = `
        SELECT fecha, exitoso, motivo, ip_address
        FROM intentos_acceso
        WHERE usuario_rut = (SELECT rut FROM usuarios WHERE id = $1)
      `
      
      const params = [usuarioId]
      let paramIndex = 2

      if (fechaInicio) {
        query += ` AND fecha >= $${paramIndex}`
        params.push(fechaInicio)
        paramIndex++
      }

      if (fechaFin) {
        query += ` AND fecha <= $${paramIndex}`
        params.push(fechaFin)
        paramIndex++
      }

      if (exitosos !== null) {
        query += ` AND exitoso = $${paramIndex}`
        params.push(exitosos)
        paramIndex++
      }

      query += ` ORDER BY fecha DESC LIMIT $${paramIndex} OFFSET $${paramIndex + 1}`
      params.push(limit, (page - 1) * limit)

      const result = await this.query(query, params)
      return result.rows
    } catch (error) {
      logger.error('Error al obtener historial de accesos:', error)
      throw error
    }
  }

  /**
   * Obtiene estadísticas de usuarios
   */
  async getEstadisticasUsuarios(sucursalId = null) {
    try {
      let query = `
        SELECT 
          COUNT(*) as total_usuarios,
          COUNT(*) FILTER (WHERE activo = true) as usuarios_activos,
          COUNT(*) FILTER (WHERE activo = false) as usuarios_inactivos,
          COUNT(*) FILTER (WHERE rol = 'admin') as administradores,
          COUNT(*) FILTER (WHERE rol = 'gerente') as gerentes,
          COUNT(*) FILTER (WHERE rol = 'vendedor') as vendedores,
          COUNT(*) FILTER (WHERE rol = 'cajero') as cajeros,
          COUNT(*) FILTER (WHERE ultimo_acceso >= NOW() - INTERVAL '24 hours') as activos_hoy,
          COUNT(*) FILTER (WHERE ultimo_acceso >= NOW() - INTERVAL '7 days') as activos_semana
        FROM usuarios
      `
      
      const params = []
      
      if (sucursalId) {
        query += ' WHERE sucursal_id = $1'
        params.push(sucursalId)
      }

      const result = await this.query(query, params)
      return result.rows[0]
    } catch (error) {
      logger.error('Error al obtener estadísticas de usuarios:', error)
      throw error
    }
  }

  /**
   * Genera hash de contraseña con salt
   */
  async hashPassword(password) {
    try {
      const salt = await bcrypt.genSalt(12)
      const hash = await bcrypt.hash(password, salt)
      return { hash, salt }
    } catch (error) {
      logger.error('Error al generar hash de contraseña:', error)
      throw error
    }
  }

  /**
   * Verifica una contraseña contra su hash
   */
  async verificarPassword(password, hash, salt) {
    try {
      return await bcrypt.compare(password, hash)
    } catch (error) {
      logger.error('Error al verificar contraseña:', error)
      return false
    }
  }

  /**
   * Valida que una contraseña cumple los requisitos
   */
  validarPassword(password) {
    if (!password || password.length < 8) {
      throw new Error('La contraseña debe tener al menos 8 caracteres')
    }

    if (!/[A-Z]/.test(password)) {
      throw new Error('La contraseña debe contener al menos una letra mayúscula')
    }

    if (!/[a-z]/.test(password)) {
      throw new Error('La contraseña debe contener al menos una letra minúscula')
    }

    if (!/[0-9]/.test(password)) {
      throw new Error('La contraseña debe contener al menos un número')
    }

    if (!/[!@#$%^&*(),.?":{}|<>]/.test(password)) {
      throw new Error('La contraseña debe contener al menos un carácter especial')
    }

    return true
  }

  /**
   * Registra un intento de acceso
   */
  async registrarIntentoAcceso(rut, exitoso, motivo, ipAddress = null) {
    try {
      await this.query(`
        INSERT INTO intentos_acceso (usuario_rut, exitoso, motivo, ip_address, fecha)
        VALUES ($1, $2, $3, $4, NOW())
      `, [rut, exitoso, motivo, ipAddress])
    } catch (error) {
      logger.error('Error al registrar intento de acceso:', error)
      // No lanzar error para no afectar el flujo principal
    }
  }

  /**
   * Registra eventos en auditoría de usuarios
   */
  async registrarAuditoria(usuarioId, accion, detalles, realizadoPorId, client = null) {
    try {
      const queryClient = client || this

      await queryClient.query(`
        INSERT INTO auditoria_usuarios (
          usuario_id, accion, detalles, realizado_por_id, fecha
        ) VALUES ($1, $2, $3, $4, NOW())
      `, [usuarioId, accion, JSON.stringify(detalles), realizadoPorId])
    } catch (error) {
      logger.error('Error al registrar auditoría:', error)
      // No lanzar error para no afectar el flujo principal
    }
  }

  /**
   * Limpia tokens de recuperación expirados
   */
  async limpiarTokensExpirados() {
    try {
      const result = await this.query(`
        UPDATE usuarios 
        SET token_recuperacion = NULL, token_recuperacion_expira = NULL
        WHERE token_recuperacion_expira < NOW()
      `)

      logger.info('Tokens de recuperación expirados limpiados', {
        tokensLimpiados: result.rowCount
      })

      return result.rowCount
    } catch (error) {
      logger.error('Error al limpiar tokens expirados:', error)
      throw error
    }
  }

  /**
   * Desbloquea usuarios con bloqueo expirado
   */
  async desbloquearUsuariosExpirados() {
    try {
      const result = await this.query(`
        UPDATE usuarios 
        SET bloqueado_hasta = NULL, intentos_fallidos = 0
        WHERE bloqueado_hasta < NOW()
      `)

      logger.info('Usuarios desbloqueados automáticamente', {
        usuariosDesbloqueados: result.rowCount
      })

      return result.rowCount
    } catch (error) {
      logger.error('Error al desbloquear usuarios:', error)
      throw error
    }
  }
}

module.exports = new Usuario()

