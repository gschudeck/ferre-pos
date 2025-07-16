/**
 * Controlador de Usuarios - Sistema Ferre-POS
 * 
 * Maneja todas las operaciones del módulo de usuarios,
 * incluyendo CRUD, autenticación, perfiles y administración.
 */

const Usuario = require('../models/Usuario')
const logger = require('../utils/logger')

class UsuarioController {
  /**
   * Obtiene lista de usuarios con filtros y paginación
   */
  async getUsuarios(request, reply) {
    try {
      const {
        sucursal_id,
        rol,
        activo,
        busqueda,
        page = 1,
        limit = 20,
        order_by = 'nombre',
        order_direction = 'ASC'
      } = request.query

      // Verificar permisos
      if (!['admin', 'gerente'].includes(request.user.rol)) {
        return reply.code(403).send({
          code: 'INSUFFICIENT_PERMISSIONS',
          error: 'Forbidden',
          message: 'No tiene permisos para ver la lista de usuarios'
        })
      }

      // Los gerentes solo pueden ver usuarios de su sucursal
      let filtroSucursal = sucursal_id
      if (request.user.rol === 'gerente' && !filtroSucursal) {
        filtroSucursal = request.user.sucursal_id
      }

      const options = {
        sucursalId: filtroSucursal,
        rol,
        activo: activo !== undefined ? activo === 'true' : null,
        busqueda,
        page: parseInt(page),
        limit: Math.min(parseInt(limit), 100),
        orderBy: order_by,
        orderDirection: order_direction
      }

      const resultado = await Usuario.getUsuarios(options)

      reply.send({
        success: true,
        data: resultado.data,
        pagination: resultado.pagination
      })
    } catch (error) {
      logger.error('Error al obtener usuarios:', error)
      reply.code(500).send({
        code: 'INTERNAL_ERROR',
        error: 'Internal Server Error',
        message: 'Error interno del servidor'
      })
    }
  }

  /**
   * Obtiene un usuario específico por ID
   */
  async getUsuario(request, reply) {
    try {
      const { id } = request.params

      // Verificar permisos
      if (!['admin', 'gerente'].includes(request.user.rol) && request.user.id !== id) {
        return reply.code(403).send({
          code: 'INSUFFICIENT_PERMISSIONS',
          error: 'Forbidden',
          message: 'No tiene permisos para ver este usuario'
        })
      }

      const usuario = await Usuario.getUsuarioCompleto(id)

      // Los gerentes solo pueden ver usuarios de su sucursal
      if (request.user.rol === 'gerente' && 
          usuario.sucursal_id !== request.user.sucursal_id) {
        return reply.code(403).send({
          code: 'INSUFFICIENT_PERMISSIONS',
          error: 'Forbidden',
          message: 'No tiene permisos para ver este usuario'
        })
      }

      reply.send({
        success: true,
        data: usuario
      })
    } catch (error) {
      logger.error('Error al obtener usuario:', error)
      
      if (error.message.includes('no encontrado')) {
        return reply.code(404).send({
          code: 'USER_NOT_FOUND',
          error: 'Not Found',
          message: 'Usuario no encontrado'
        })
      }

      reply.code(500).send({
        code: 'INTERNAL_ERROR',
        error: 'Internal Server Error',
        message: 'Error interno del servidor'
      })
    }
  }

  /**
   * Crea un nuevo usuario
   */
  async createUsuario(request, reply) {
    try {
      // Solo administradores pueden crear usuarios
      if (request.user.rol !== 'admin') {
        return reply.code(403).send({
          code: 'INSUFFICIENT_PERMISSIONS',
          error: 'Forbidden',
          message: 'Solo los administradores pueden crear usuarios'
        })
      }

      const usuario = await Usuario.createUsuario(request.body, request.user.id)

      logger.business('Usuario creado exitosamente', {
        usuarioId: usuario.id,
        rut: usuario.rut,
        nombre: usuario.nombre,
        rol: usuario.rol,
        creadorId: request.user.id
      })

      reply.code(201).send({
        success: true,
        message: 'Usuario creado exitosamente',
        data: usuario
      })
    } catch (error) {
      logger.error('Error al crear usuario:', error)
      
      let statusCode = 500
      let errorCode = 'INTERNAL_ERROR'
      
      if (error.message.includes('Ya existe un usuario')) {
        statusCode = 409
        errorCode = 'USER_ALREADY_EXISTS'
      } else if (error.message.includes('Rol') && error.message.includes('no es válido')) {
        statusCode = 400
        errorCode = 'INVALID_ROLE'
      } else if (error.message.includes('Sucursal no encontrada')) {
        statusCode = 400
        errorCode = 'INVALID_SUCURSAL'
      } else if (error.message.includes('contraseña')) {
        statusCode = 400
        errorCode = 'INVALID_PASSWORD'
      }

      reply.code(statusCode).send({
        code: errorCode,
        error: statusCode === 500 ? 'Internal Server Error' : 'Bad Request',
        message: statusCode === 500 ? 'Error interno del servidor' : error.message
      })
    }
  }

  /**
   * Actualiza un usuario existente
   */
  async updateUsuario(request, reply) {
    try {
      const { id } = request.params

      // Verificar permisos
      if (!['admin', 'gerente'].includes(request.user.rol) && request.user.id !== id) {
        return reply.code(403).send({
          code: 'INSUFFICIENT_PERMISSIONS',
          error: 'Forbidden',
          message: 'No tiene permisos para modificar este usuario'
        })
      }

      // Los usuarios normales solo pueden modificar ciertos campos
      if (request.user.id === id && !['admin', 'gerente'].includes(request.user.rol)) {
        const camposPermitidos = ['nombre', 'email', 'telefono']
        const camposEnviados = Object.keys(request.body)
        const camposNoPermitidos = camposEnviados.filter(campo => !camposPermitidos.includes(campo))
        
        if (camposNoPermitidos.length > 0) {
          return reply.code(403).send({
            code: 'INSUFFICIENT_PERMISSIONS',
            error: 'Forbidden',
            message: `No puede modificar los campos: ${camposNoPermitidos.join(', ')}`
          })
        }
      }

      const usuario = await Usuario.updateUsuario(id, request.body, request.user.id)

      reply.send({
        success: true,
        message: 'Usuario actualizado exitosamente',
        data: usuario
      })
    } catch (error) {
      logger.error('Error al actualizar usuario:', error)
      
      let statusCode = 500
      let errorCode = 'INTERNAL_ERROR'
      
      if (error.message.includes('no encontrado')) {
        statusCode = 404
        errorCode = 'USER_NOT_FOUND'
      } else if (error.message.includes('Ya existe un usuario')) {
        statusCode = 409
        errorCode = 'EMAIL_ALREADY_EXISTS'
      } else if (error.message.includes('No hay cambios')) {
        statusCode = 400
        errorCode = 'NO_CHANGES'
      } else if (error.message.includes('Rol') && error.message.includes('no es válido')) {
        statusCode = 400
        errorCode = 'INVALID_ROLE'
      }

      reply.code(statusCode).send({
        code: errorCode,
        error: statusCode === 500 ? 'Internal Server Error' : 'Bad Request',
        message: statusCode === 500 ? 'Error interno del servidor' : error.message
      })
    }
  }

  /**
   * Cambia la contraseña de un usuario
   */
  async cambiarPassword(request, reply) {
    try {
      const { id } = request.params
      const { password_actual, password_nueva } = request.body

      // Verificar permisos
      if (request.user.id !== id && request.user.rol !== 'admin') {
        return reply.code(403).send({
          code: 'INSUFFICIENT_PERMISSIONS',
          error: 'Forbidden',
          message: 'No tiene permisos para cambiar la contraseña de este usuario'
        })
      }

      // Si es admin cambiando contraseña de otro usuario, no requiere contraseña actual
      const esAdminCambiandoOtro = request.user.rol === 'admin' && request.user.id !== id

      await Usuario.cambiarPassword(
        id, 
        esAdminCambiandoOtro ? null : password_actual, 
        password_nueva,
        request.user.id
      )

      reply.send({
        success: true,
        message: 'Contraseña cambiada exitosamente'
      })
    } catch (error) {
      logger.error('Error al cambiar contraseña:', error)
      
      let statusCode = 500
      let errorCode = 'INTERNAL_ERROR'
      
      if (error.message.includes('no encontrado')) {
        statusCode = 404
        errorCode = 'USER_NOT_FOUND'
      } else if (error.message.includes('Contraseña actual incorrecta')) {
        statusCode = 400
        errorCode = 'INVALID_CURRENT_PASSWORD'
      } else if (error.message.includes('contraseña debe')) {
        statusCode = 400
        errorCode = 'INVALID_PASSWORD_FORMAT'
      }

      reply.code(statusCode).send({
        code: errorCode,
        error: statusCode === 500 ? 'Internal Server Error' : 'Bad Request',
        message: statusCode === 500 ? 'Error interno del servidor' : error.message
      })
    }
  }

  /**
   * Inicia el proceso de recuperación de contraseña
   */
  async iniciarRecuperacionPassword(request, reply) {
    try {
      const { email } = request.body

      const resultado = await Usuario.iniciarRecuperacionPassword(email)

      reply.send({
        success: true,
        message: resultado.message,
        // Solo incluir token en desarrollo
        ...(process.env.NODE_ENV === 'development' && { token: resultado.token })
      })
    } catch (error) {
      logger.error('Error al iniciar recuperación de contraseña:', error)
      
      if (error.message.includes('Usuario inactivo')) {
        return reply.code(400).send({
          code: 'USER_INACTIVE',
          error: 'Bad Request',
          message: 'Usuario inactivo'
        })
      }

      reply.code(500).send({
        code: 'INTERNAL_ERROR',
        error: 'Internal Server Error',
        message: 'Error interno del servidor'
      })
    }
  }

  /**
   * Completa el proceso de recuperación de contraseña
   */
  async completarRecuperacionPassword(request, reply) {
    try {
      const { token, password_nueva } = request.body

      const resultado = await Usuario.completarRecuperacionPassword(token, password_nueva)

      reply.send({
        success: true,
        message: resultado.message
      })
    } catch (error) {
      logger.error('Error al completar recuperación de contraseña:', error)
      
      let statusCode = 500
      let errorCode = 'INTERNAL_ERROR'
      
      if (error.message.includes('Token inválido')) {
        statusCode = 400
        errorCode = 'INVALID_TOKEN'
      } else if (error.message.includes('contraseña debe')) {
        statusCode = 400
        errorCode = 'INVALID_PASSWORD_FORMAT'
      }

      reply.code(statusCode).send({
        code: errorCode,
        error: statusCode === 500 ? 'Internal Server Error' : 'Bad Request',
        message: statusCode === 500 ? 'Error interno del servidor' : error.message
      })
    }
  }

  /**
   * Desactiva un usuario (eliminación lógica)
   */
  async desactivarUsuario(request, reply) {
    try {
      const { id } = request.params
      const { motivo } = request.body

      // Solo administradores pueden desactivar usuarios
      if (request.user.rol !== 'admin') {
        return reply.code(403).send({
          code: 'INSUFFICIENT_PERMISSIONS',
          error: 'Forbidden',
          message: 'Solo los administradores pueden desactivar usuarios'
        })
      }

      await Usuario.desactivarUsuario(id, motivo, request.user.id)

      reply.send({
        success: true,
        message: 'Usuario desactivado exitosamente'
      })
    } catch (error) {
      logger.error('Error al desactivar usuario:', error)
      
      let statusCode = 500
      let errorCode = 'INTERNAL_ERROR'
      
      if (error.message.includes('no encontrado')) {
        statusCode = 404
        errorCode = 'USER_NOT_FOUND'
      } else if (error.message.includes('ya está inactivo')) {
        statusCode = 400
        errorCode = 'USER_ALREADY_INACTIVE'
      } else if (error.message.includes('No puede desactivar su propia cuenta')) {
        statusCode = 400
        errorCode = 'CANNOT_DEACTIVATE_SELF'
      }

      reply.code(statusCode).send({
        code: errorCode,
        error: statusCode === 500 ? 'Internal Server Error' : 'Bad Request',
        message: statusCode === 500 ? 'Error interno del servidor' : error.message
      })
    }
  }

  /**
   * Reactiva un usuario
   */
  async reactivarUsuario(request, reply) {
    try {
      const { id } = request.params

      // Solo administradores pueden reactivar usuarios
      if (request.user.rol !== 'admin') {
        return reply.code(403).send({
          code: 'INSUFFICIENT_PERMISSIONS',
          error: 'Forbidden',
          message: 'Solo los administradores pueden reactivar usuarios'
        })
      }

      await Usuario.reactivarUsuario(id, request.user.id)

      reply.send({
        success: true,
        message: 'Usuario reactivado exitosamente'
      })
    } catch (error) {
      logger.error('Error al reactivar usuario:', error)
      
      let statusCode = 500
      let errorCode = 'INTERNAL_ERROR'
      
      if (error.message.includes('no encontrado')) {
        statusCode = 404
        errorCode = 'USER_NOT_FOUND'
      } else if (error.message.includes('ya está activo')) {
        statusCode = 400
        errorCode = 'USER_ALREADY_ACTIVE'
      }

      reply.code(statusCode).send({
        code: errorCode,
        error: statusCode === 500 ? 'Internal Server Error' : 'Bad Request',
        message: statusCode === 500 ? 'Error interno del servidor' : error.message
      })
    }
  }

  /**
   * Obtiene el perfil del usuario autenticado
   */
  async getPerfil(request, reply) {
    try {
      const usuario = await Usuario.getUsuarioCompleto(request.user.id)

      reply.send({
        success: true,
        data: {
          id: usuario.id,
          rut: usuario.rut,
          nombre: usuario.nombre,
          email: usuario.email,
          telefono: usuario.telefono,
          rol: usuario.rol,
          sucursal_id: usuario.sucursal_id,
          sucursal_nombre: usuario.sucursal_nombre,
          ultimo_acceso: usuario.ultimo_acceso,
          debe_cambiar_password: usuario.debe_cambiar_password,
          fecha_creacion: usuario.fecha_creacion
        }
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
   * Actualiza el perfil del usuario autenticado
   */
  async updatePerfil(request, reply) {
    try {
      // Solo permitir actualizar ciertos campos del perfil
      const camposPermitidos = ['nombre', 'email', 'telefono']
      const datosActualizacion = {}
      
      for (const campo of camposPermitidos) {
        if (request.body[campo] !== undefined) {
          datosActualizacion[campo] = request.body[campo]
        }
      }

      if (Object.keys(datosActualizacion).length === 0) {
        return reply.code(400).send({
          code: 'NO_VALID_FIELDS',
          error: 'Bad Request',
          message: 'No se proporcionaron campos válidos para actualizar'
        })
      }

      const usuario = await Usuario.updateUsuario(
        request.user.id, 
        datosActualizacion, 
        request.user.id
      )

      reply.send({
        success: true,
        message: 'Perfil actualizado exitosamente',
        data: usuario
      })
    } catch (error) {
      logger.error('Error al actualizar perfil:', error)
      
      let statusCode = 500
      let errorCode = 'INTERNAL_ERROR'
      
      if (error.message.includes('Ya existe un usuario')) {
        statusCode = 409
        errorCode = 'EMAIL_ALREADY_EXISTS'
      }

      reply.code(statusCode).send({
        code: errorCode,
        error: statusCode === 500 ? 'Internal Server Error' : 'Conflict',
        message: statusCode === 500 ? 'Error interno del servidor' : error.message
      })
    }
  }

  /**
   * Obtiene el historial de accesos de un usuario
   */
  async getHistorialAccesos(request, reply) {
    try {
      const { id } = request.params
      const {
        fecha_inicio,
        fecha_fin,
        exitosos,
        page = 1,
        limit = 50
      } = request.query

      // Verificar permisos
      if (!['admin', 'gerente'].includes(request.user.rol) && request.user.id !== id) {
        return reply.code(403).send({
          code: 'INSUFFICIENT_PERMISSIONS',
          error: 'Forbidden',
          message: 'No tiene permisos para ver el historial de este usuario'
        })
      }

      const options = {
        fechaInicio: fecha_inicio ? new Date(fecha_inicio) : null,
        fechaFin: fecha_fin ? new Date(fecha_fin) : null,
        exitosos: exitosos !== undefined ? exitosos === 'true' : null,
        page: parseInt(page),
        limit: Math.min(parseInt(limit), 100)
      }

      const historial = await Usuario.getHistorialAccesos(id, options)

      reply.send({
        success: true,
        data: historial
      })
    } catch (error) {
      logger.error('Error al obtener historial de accesos:', error)
      reply.code(500).send({
        code: 'INTERNAL_ERROR',
        error: 'Internal Server Error',
        message: 'Error interno del servidor'
      })
    }
  }

  /**
   * Obtiene estadísticas de usuarios
   */
  async getEstadisticas(request, reply) {
    try {
      // Solo administradores y gerentes pueden ver estadísticas
      if (!['admin', 'gerente'].includes(request.user.rol)) {
        return reply.code(403).send({
          code: 'INSUFFICIENT_PERMISSIONS',
          error: 'Forbidden',
          message: 'No tiene permisos para ver estadísticas de usuarios'
        })
      }

      // Los gerentes solo ven estadísticas de su sucursal
      const sucursalId = request.user.rol === 'gerente' ? request.user.sucursal_id : request.query.sucursal_id

      const estadisticas = await Usuario.getEstadisticasUsuarios(sucursalId)

      reply.send({
        success: true,
        data: estadisticas
      })
    } catch (error) {
      logger.error('Error al obtener estadísticas de usuarios:', error)
      reply.code(500).send({
        code: 'INTERNAL_ERROR',
        error: 'Internal Server Error',
        message: 'Error interno del servidor'
      })
    }
  }

  /**
   * Desbloquea un usuario bloqueado
   */
  async desbloquearUsuario(request, reply) {
    try {
      const { id } = request.params

      // Solo administradores pueden desbloquear usuarios
      if (request.user.rol !== 'admin') {
        return reply.code(403).send({
          code: 'INSUFFICIENT_PERMISSIONS',
          error: 'Forbidden',
          message: 'Solo los administradores pueden desbloquear usuarios'
        })
      }

      // Desbloquear usuario manualmente
      await Usuario.query(`
        UPDATE usuarios 
        SET intentos_fallidos = 0, bloqueado_hasta = NULL
        WHERE id = $1
      `, [id])

      // Registrar en auditoría
      await Usuario.registrarAuditoria(
        id,
        'USUARIO_DESBLOQUEADO_MANUAL',
        {},
        request.user.id
      )

      reply.send({
        success: true,
        message: 'Usuario desbloqueado exitosamente'
      })
    } catch (error) {
      logger.error('Error al desbloquear usuario:', error)
      reply.code(500).send({
        code: 'INTERNAL_ERROR',
        error: 'Internal Server Error',
        message: 'Error interno del servidor'
      })
    }
  }

  /**
   * Fuerza el cambio de contraseña en el próximo login
   */
  async forzarCambioPassword(request, reply) {
    try {
      const { id } = request.params

      // Solo administradores pueden forzar cambio de contraseña
      if (request.user.rol !== 'admin') {
        return reply.code(403).send({
          code: 'INSUFFICIENT_PERMISSIONS',
          error: 'Forbidden',
          message: 'Solo los administradores pueden forzar cambio de contraseña'
        })
      }

      await Usuario.query(`
        UPDATE usuarios 
        SET debe_cambiar_password = true
        WHERE id = $1
      `, [id])

      // Registrar en auditoría
      await Usuario.registrarAuditoria(
        id,
        'CAMBIO_PASSWORD_FORZADO',
        {},
        request.user.id
      )

      reply.send({
        success: true,
        message: 'Cambio de contraseña forzado exitosamente'
      })
    } catch (error) {
      logger.error('Error al forzar cambio de contraseña:', error)
      reply.code(500).send({
        code: 'INTERNAL_ERROR',
        error: 'Internal Server Error',
        message: 'Error interno del servidor'
      })
    }
  }

  /**
   * Ejecuta tareas de mantenimiento de usuarios
   */
  async ejecutarMantenimiento(request, reply) {
    try {
      // Solo administradores pueden ejecutar mantenimiento
      if (request.user.rol !== 'admin') {
        return reply.code(403).send({
          code: 'INSUFFICIENT_PERMISSIONS',
          error: 'Forbidden',
          message: 'Solo los administradores pueden ejecutar mantenimiento'
        })
      }

      const resultados = {
        timestamp: new Date().toISOString(),
        tareas_ejecutadas: []
      }

      // Limpiar tokens de recuperación expirados
      const tokensLimpiados = await Usuario.limpiarTokensExpirados()
      resultados.tareas_ejecutadas.push({
        tarea: 'limpiar_tokens_expirados',
        resultado: { tokensLimpiados }
      })

      // Desbloquear usuarios con bloqueo expirado
      const usuariosDesbloqueados = await Usuario.desbloquearUsuariosExpirados()
      resultados.tareas_ejecutadas.push({
        tarea: 'desbloquear_usuarios_expirados',
        resultado: { usuariosDesbloqueados }
      })

      logger.business('Mantenimiento de usuarios ejecutado', {
        tokensLimpiados,
        usuariosDesbloqueados,
        ejecutadoPorId: request.user.id
      })

      reply.send({
        success: true,
        message: 'Mantenimiento ejecutado exitosamente',
        data: resultados
      })
    } catch (error) {
      logger.error('Error al ejecutar mantenimiento de usuarios:', error)
      reply.code(500).send({
        code: 'INTERNAL_ERROR',
        error: 'Internal Server Error',
        message: 'Error interno del servidor'
      })
    }
  }
}

module.exports = new UsuarioController()

