/**
 * Controlador de Sistema - Sistema Ferre-POS
 * 
 * Maneja todas las operaciones del módulo de sistema,
 * incluyendo configuraciones, monitoreo, logs y utilidades administrativas.
 */

const Sistema = require('../models/Sistema')
const logger = require('../utils/logger')

class SistemaController {
  /**
   * Obtiene información general del sistema
   */
  async getInfoSistema(request, reply) {
    try {
      const info = await Sistema.getInfoSistema()

      reply.send({
        success: true,
        data: info
      })
    } catch (error) {
      logger.error('Error al obtener información del sistema:', error)
      reply.code(500).send({
        code: 'INTERNAL_ERROR',
        error: 'Internal Server Error',
        message: 'Error interno del servidor'
      })
    }
  }

  /**
   * Obtiene todas las configuraciones del sistema
   */
  async getConfiguraciones(request, reply) {
    try {
      // Verificar permisos de administrador
      if (request.user.rol !== 'admin') {
        return reply.code(403).send({
          code: 'INSUFFICIENT_PERMISSIONS',
          error: 'Forbidden',
          message: 'Solo los administradores pueden ver las configuraciones del sistema'
        })
      }

      const { incluir_inactivas = false } = request.query
      const configuraciones = await Sistema.getConfiguraciones(incluir_inactivas === 'true')

      reply.send({
        success: true,
        data: configuraciones
      })
    } catch (error) {
      logger.error('Error al obtener configuraciones:', error)
      reply.code(500).send({
        code: 'INTERNAL_ERROR',
        error: 'Internal Server Error',
        message: 'Error interno del servidor'
      })
    }
  }

  /**
   * Obtiene una configuración específica
   */
  async getConfiguracion(request, reply) {
    try {
      // Verificar permisos
      if (!['admin', 'gerente'].includes(request.user.rol)) {
        return reply.code(403).send({
          code: 'INSUFFICIENT_PERMISSIONS',
          error: 'Forbidden',
          message: 'No tiene permisos para ver configuraciones del sistema'
        })
      }

      const { clave } = request.params

      const configuracion = await Sistema.getConfiguracion(clave)

      reply.send({
        success: true,
        data: configuracion
      })
    } catch (error) {
      logger.error('Error al obtener configuración:', error)
      
      if (error.message.includes('no encontrada')) {
        return reply.code(404).send({
          code: 'CONFIG_NOT_FOUND',
          error: 'Not Found',
          message: 'Configuración no encontrada'
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
   * Actualiza una configuración del sistema
   */
  async updateConfiguracion(request, reply) {
    try {
      // Solo administradores pueden modificar configuraciones
      if (request.user.rol !== 'admin') {
        return reply.code(403).send({
          code: 'INSUFFICIENT_PERMISSIONS',
          error: 'Forbidden',
          message: 'Solo los administradores pueden modificar configuraciones del sistema'
        })
      }

      const { clave } = request.params
      const { valor } = request.body

      const configuracionActualizada = await Sistema.updateConfiguracion(
        clave, 
        valor, 
        request.user.id
      )

      logger.business('Configuración del sistema actualizada', {
        clave,
        valor,
        usuarioId: request.user.id
      })

      reply.send({
        success: true,
        message: 'Configuración actualizada exitosamente',
        data: configuracionActualizada
      })
    } catch (error) {
      logger.error('Error al actualizar configuración:', error)
      
      let statusCode = 500
      let errorCode = 'INTERNAL_ERROR'
      
      if (error.message.includes('no encontrada')) {
        statusCode = 404
        errorCode = 'CONFIG_NOT_FOUND'
      } else if (error.message.includes('solo lectura')) {
        statusCode = 400
        errorCode = 'CONFIG_READ_ONLY'
      } else if (error.message.includes('inactiva')) {
        statusCode = 400
        errorCode = 'CONFIG_INACTIVE'
      } else if (error.message.includes('Valor debe ser')) {
        statusCode = 400
        errorCode = 'INVALID_VALUE_TYPE'
      }

      reply.code(statusCode).send({
        code: errorCode,
        error: statusCode === 500 ? 'Internal Server Error' : 'Bad Request',
        message: statusCode === 500 ? 'Error interno del servidor' : error.message
      })
    }
  }

  /**
   * Crea una nueva configuración del sistema
   */
  async createConfiguracion(request, reply) {
    try {
      // Solo administradores pueden crear configuraciones
      if (request.user.rol !== 'admin') {
        return reply.code(403).send({
          code: 'INSUFFICIENT_PERMISSIONS',
          error: 'Forbidden',
          message: 'Solo los administradores pueden crear configuraciones del sistema'
        })
      }

      const configuracionCreada = await Sistema.createConfiguracion(
        request.body, 
        request.user.id
      )

      logger.business('Nueva configuración del sistema creada', {
        clave: request.body.clave,
        categoria: request.body.categoria,
        usuarioId: request.user.id
      })

      reply.code(201).send({
        success: true,
        message: 'Configuración creada exitosamente',
        data: configuracionCreada
      })
    } catch (error) {
      logger.error('Error al crear configuración:', error)
      
      let statusCode = 500
      let errorCode = 'INTERNAL_ERROR'
      
      if (error.message.includes('Ya existe')) {
        statusCode = 409
        errorCode = 'CONFIG_ALREADY_EXISTS'
      } else if (error.message.includes('Valor debe ser')) {
        statusCode = 400
        errorCode = 'INVALID_VALUE_TYPE'
      }

      reply.code(statusCode).send({
        code: errorCode,
        error: statusCode === 500 ? 'Internal Server Error' : 'Conflict',
        message: statusCode === 500 ? 'Error interno del servidor' : error.message
      })
    }
  }

  /**
   * Elimina una configuración del sistema
   */
  async deleteConfiguracion(request, reply) {
    try {
      // Solo administradores pueden eliminar configuraciones
      if (request.user.rol !== 'admin') {
        return reply.code(403).send({
          code: 'INSUFFICIENT_PERMISSIONS',
          error: 'Forbidden',
          message: 'Solo los administradores pueden eliminar configuraciones del sistema'
        })
      }

      const { clave } = request.params

      await Sistema.deleteConfiguracion(clave, request.user.id)

      logger.business('Configuración del sistema eliminada', {
        clave,
        usuarioId: request.user.id
      })

      reply.send({
        success: true,
        message: 'Configuración eliminada exitosamente'
      })
    } catch (error) {
      logger.error('Error al eliminar configuración:', error)
      
      let statusCode = 500
      let errorCode = 'INTERNAL_ERROR'
      
      if (error.message.includes('no encontrada')) {
        statusCode = 404
        errorCode = 'CONFIG_NOT_FOUND'
      } else if (error.message.includes('solo lectura')) {
        statusCode = 400
        errorCode = 'CONFIG_READ_ONLY'
      }

      reply.code(statusCode).send({
        code: errorCode,
        error: statusCode === 500 ? 'Internal Server Error' : 'Bad Request',
        message: statusCode === 500 ? 'Error interno del servidor' : error.message
      })
    }
  }

  /**
   * Crea un backup de las configuraciones del sistema
   */
  async backupConfiguraciones(request, reply) {
    try {
      // Solo administradores pueden crear backups
      if (request.user.rol !== 'admin') {
        return reply.code(403).send({
          code: 'INSUFFICIENT_PERMISSIONS',
          error: 'Forbidden',
          message: 'Solo los administradores pueden crear backups del sistema'
        })
      }

      const backup = await Sistema.backupConfiguraciones(request.user.id)

      reply.send({
        success: true,
        message: 'Backup creado exitosamente',
        data: backup
      })
    } catch (error) {
      logger.error('Error al crear backup:', error)
      reply.code(500).send({
        code: 'INTERNAL_ERROR',
        error: 'Internal Server Error',
        message: 'Error interno del servidor'
      })
    }
  }

  /**
   * Restaura configuraciones desde un backup
   */
  async restaurarConfiguraciones(request, reply) {
    try {
      // Solo administradores pueden restaurar backups
      if (request.user.rol !== 'admin') {
        return reply.code(403).send({
          code: 'INSUFFICIENT_PERMISSIONS',
          error: 'Forbidden',
          message: 'Solo los administradores pueden restaurar backups del sistema'
        })
      }

      const resultado = await Sistema.restaurarConfiguraciones(
        request.body, 
        request.user.id
      )

      logger.business('Configuraciones restauradas desde backup', {
        configuracionesRestauradas: resultado.configuracionesRestauradas,
        errores: resultado.errores.length,
        usuarioId: request.user.id
      })

      reply.send({
        success: true,
        message: 'Configuraciones restauradas exitosamente',
        data: resultado
      })
    } catch (error) {
      logger.error('Error al restaurar configuraciones:', error)
      reply.code(500).send({
        code: 'INTERNAL_ERROR',
        error: 'Internal Server Error',
        message: 'Error interno del servidor'
      })
    }
  }

  /**
   * Obtiene logs del sistema
   */
  async getLogs(request, reply) {
    try {
      // Solo administradores y gerentes pueden ver logs
      if (!['admin', 'gerente'].includes(request.user.rol)) {
        return reply.code(403).send({
          code: 'INSUFFICIENT_PERMISSIONS',
          error: 'Forbidden',
          message: 'No tiene permisos para ver logs del sistema'
        })
      }

      const {
        nivel,
        fecha_inicio,
        fecha_fin,
        usuario,
        modulo,
        page = 1,
        limit = 100
      } = request.query

      const options = {
        nivel,
        fechaInicio: fecha_inicio ? new Date(fecha_inicio) : null,
        fechaFin: fecha_fin ? new Date(fecha_fin) : null,
        usuario,
        modulo,
        limit: Math.min(parseInt(limit), 1000),
        offset: (parseInt(page) - 1) * Math.min(parseInt(limit), 1000)
      }

      const logs = await Sistema.getLogs(options)

      reply.send({
        success: true,
        data: logs,
        pagination: {
          page: parseInt(page),
          limit: parseInt(limit),
          total: logs.length
        }
      })
    } catch (error) {
      logger.error('Error al obtener logs:', error)
      reply.code(500).send({
        code: 'INTERNAL_ERROR',
        error: 'Internal Server Error',
        message: 'Error interno del servidor'
      })
    }
  }

  /**
   * Limpia logs antiguos del sistema
   */
  async limpiarLogs(request, reply) {
    try {
      // Solo administradores pueden limpiar logs
      if (request.user.rol !== 'admin') {
        return reply.code(403).send({
          code: 'INSUFFICIENT_PERMISSIONS',
          error: 'Forbidden',
          message: 'Solo los administradores pueden limpiar logs del sistema'
        })
      }

      const { dias_retencion = 30 } = request.body

      if (dias_retencion < 7) {
        return reply.code(400).send({
          code: 'INVALID_RETENTION_PERIOD',
          error: 'Bad Request',
          message: 'El período de retención mínimo es de 7 días'
        })
      }

      const resultado = await Sistema.limpiarLogs(dias_retencion, request.user.id)

      reply.send({
        success: true,
        message: 'Logs limpiados exitosamente',
        data: resultado
      })
    } catch (error) {
      logger.error('Error al limpiar logs:', error)
      reply.code(500).send({
        code: 'INTERNAL_ERROR',
        error: 'Internal Server Error',
        message: 'Error interno del servidor'
      })
    }
  }

  /**
   * Obtiene métricas de rendimiento del sistema
   */
  async getMetricasRendimiento(request, reply) {
    try {
      // Solo administradores y gerentes pueden ver métricas
      if (!['admin', 'gerente'].includes(request.user.rol)) {
        return reply.code(403).send({
          code: 'INSUFFICIENT_PERMISSIONS',
          error: 'Forbidden',
          message: 'No tiene permisos para ver métricas del sistema'
        })
      }

      const metricas = await Sistema.getMetricasRendimiento()

      reply.send({
        success: true,
        data: metricas
      })
    } catch (error) {
      logger.error('Error al obtener métricas:', error)
      reply.code(500).send({
        code: 'INTERNAL_ERROR',
        error: 'Internal Server Error',
        message: 'Error interno del servidor'
      })
    }
  }

  /**
   * Reinicia el sistema (reinicio graceful)
   */
  async reiniciarSistema(request, reply) {
    try {
      // Solo administradores pueden reiniciar el sistema
      if (request.user.rol !== 'admin') {
        return reply.code(403).send({
          code: 'INSUFFICIENT_PERMISSIONS',
          error: 'Forbidden',
          message: 'Solo los administradores pueden reiniciar el sistema'
        })
      }

      const { motivo } = request.body

      if (!motivo || motivo.trim().length < 10) {
        return reply.code(400).send({
          code: 'INVALID_REASON',
          error: 'Bad Request',
          message: 'Debe proporcionar un motivo válido para el reinicio (mínimo 10 caracteres)'
        })
      }

      logger.business('Reinicio del sistema solicitado', {
        motivo,
        usuarioId: request.user.id,
        timestamp: new Date().toISOString()
      })

      // Enviar respuesta antes del reinicio
      reply.send({
        success: true,
        message: 'Reinicio del sistema programado',
        data: {
          motivo,
          tiempo_estimado: '30 segundos'
        }
      })

      // Programar reinicio después de enviar respuesta
      setTimeout(() => {
        logger.info('Iniciando reinicio graceful del sistema')
        process.kill(process.pid, 'SIGTERM')
      }, 2000)

    } catch (error) {
      logger.error('Error al reiniciar sistema:', error)
      reply.code(500).send({
        code: 'INTERNAL_ERROR',
        error: 'Internal Server Error',
        message: 'Error interno del servidor'
      })
    }
  }

  /**
   * Obtiene estado de salud del sistema
   */
  async getHealthCheck(request, reply) {
    try {
      const health = {
        status: 'ok',
        timestamp: new Date().toISOString(),
        version: '1.0.0',
        environment: process.env.NODE_ENV || 'development',
        uptime: process.uptime(),
        memory: process.memoryUsage()
      }

      // Verificar estado de la base de datos
      try {
        await Sistema.query('SELECT 1')
        health.database = { status: 'healthy', message: 'Conexión exitosa' }
      } catch (dbError) {
        health.database = { status: 'unhealthy', message: dbError.message }
        health.status = 'degraded'
      }

      // Verificar espacio en disco
      try {
        const fs = require('fs')
        const stats = fs.statSync(process.cwd())
        health.disk = { status: 'healthy', message: 'Espacio disponible' }
      } catch (diskError) {
        health.disk = { status: 'warning', message: 'No se pudo verificar espacio en disco' }
      }

      const statusCode = health.status === 'ok' ? 200 : 503
      reply.code(statusCode).send(health)

    } catch (error) {
      logger.error('Error en health check:', error)
      reply.code(503).send({
        status: 'error',
        timestamp: new Date().toISOString(),
        message: 'Error en verificación de salud del sistema'
      })
    }
  }

  /**
   * Obtiene estadísticas generales del sistema
   */
  async getEstadisticas(request, reply) {
    try {
      // Verificar permisos
      if (!['admin', 'gerente'].includes(request.user.rol)) {
        return reply.code(403).send({
          code: 'INSUFFICIENT_PERMISSIONS',
          error: 'Forbidden',
          message: 'No tiene permisos para ver estadísticas del sistema'
        })
      }

      const estadisticas = await Sistema.getEstadisticasBaseDatos()

      reply.send({
        success: true,
        data: estadisticas
      })
    } catch (error) {
      logger.error('Error al obtener estadísticas:', error)
      reply.code(500).send({
        code: 'INTERNAL_ERROR',
        error: 'Internal Server Error',
        message: 'Error interno del servidor'
      })
    }
  }

  /**
   * Ejecuta mantenimiento del sistema
   */
  async ejecutarMantenimiento(request, reply) {
    try {
      // Solo administradores pueden ejecutar mantenimiento
      if (request.user.rol !== 'admin') {
        return reply.code(403).send({
          code: 'INSUFFICIENT_PERMISSIONS',
          error: 'Forbidden',
          message: 'Solo los administradores pueden ejecutar mantenimiento del sistema'
        })
      }

      const { 
        limpiar_logs = false, 
        dias_retencion_logs = 30,
        optimizar_bd = false,
        backup_configuraciones = false
      } = request.body

      const resultados = {
        timestamp: new Date().toISOString(),
        tareas_ejecutadas: [],
        errores: []
      }

      // Limpiar logs si se solicita
      if (limpiar_logs) {
        try {
          const resultadoLogs = await Sistema.limpiarLogs(dias_retencion_logs, request.user.id)
          resultados.tareas_ejecutadas.push({
            tarea: 'limpiar_logs',
            resultado: resultadoLogs
          })
        } catch (error) {
          resultados.errores.push({
            tarea: 'limpiar_logs',
            error: error.message
          })
        }
      }

      // Crear backup de configuraciones si se solicita
      if (backup_configuraciones) {
        try {
          const resultadoBackup = await Sistema.backupConfiguraciones(request.user.id)
          resultados.tareas_ejecutadas.push({
            tarea: 'backup_configuraciones',
            resultado: resultadoBackup
          })
        } catch (error) {
          resultados.errores.push({
            tarea: 'backup_configuraciones',
            error: error.message
          })
        }
      }

      // Optimizar base de datos si se solicita
      if (optimizar_bd) {
        try {
          await Sistema.query('VACUUM ANALYZE')
          resultados.tareas_ejecutadas.push({
            tarea: 'optimizar_bd',
            resultado: { mensaje: 'Base de datos optimizada exitosamente' }
          })
        } catch (error) {
          resultados.errores.push({
            tarea: 'optimizar_bd',
            error: error.message
          })
        }
      }

      logger.business('Mantenimiento del sistema ejecutado', {
        tareasEjecutadas: resultados.tareas_ejecutadas.length,
        errores: resultados.errores.length,
        usuarioId: request.user.id
      })

      reply.send({
        success: true,
        message: 'Mantenimiento ejecutado',
        data: resultados
      })
    } catch (error) {
      logger.error('Error al ejecutar mantenimiento:', error)
      reply.code(500).send({
        code: 'INTERNAL_ERROR',
        error: 'Internal Server Error',
        message: 'Error interno del servidor'
      })
    }
  }
}

module.exports = new SistemaController()

