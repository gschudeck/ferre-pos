/**
 * Modelo Sistema - Sistema Ferre-POS
 * 
 * Maneja todas las operaciones relacionadas con la configuración del sistema,
 * parámetros globales, información de la empresa y utilidades administrativas.
 */

const BaseModel = require('./BaseModel')
const logger = require('../utils/logger')
const database = require('../config/database')
const fs = require('fs').promises
const path = require('path')

class Sistema extends BaseModel {
  constructor() {
    super('configuracion_sistema', {
      clave: { type: 'string', required: true },
      valor: { type: 'string', required: true },
      descripcion: { type: 'string' },
      tipo_dato: { type: 'string', required: true },
      categoria: { type: 'string', required: true },
      activa: { type: 'boolean', required: true },
      solo_lectura: { type: 'boolean', required: true }
    })
  }

  /**
   * Obtiene todas las configuraciones del sistema agrupadas por categoría
   */
  async getConfiguraciones(incluirInactivas = false) {
    try {
      let query = `
        SELECT clave, valor, descripcion, tipo_dato, categoria, activa, solo_lectura,
               fecha_creacion, fecha_modificacion
        FROM configuracion_sistema
      `
      
      const params = []
      
      if (!incluirInactivas) {
        query += ' WHERE activa = true'
      }
      
      query += ' ORDER BY categoria, clave'

      const result = await this.query(query, params)
      
      // Agrupar por categoría
      const configuraciones = {}
      for (const config of result.rows) {
        if (!configuraciones[config.categoria]) {
          configuraciones[config.categoria] = {}
        }
        
        // Convertir valor según tipo de dato
        const valorConvertido = this.convertirValor(config.valor, config.tipo_dato)
        
        configuraciones[config.categoria][config.clave] = {
          valor: valorConvertido,
          descripcion: config.descripcion,
          tipo_dato: config.tipo_dato,
          activa: config.activa,
          solo_lectura: config.solo_lectura,
          fecha_creacion: config.fecha_creacion,
          fecha_modificacion: config.fecha_modificacion
        }
      }

      return configuraciones
    } catch (error) {
      logger.error('Error al obtener configuraciones del sistema:', error)
      throw error
    }
  }

  /**
   * Obtiene una configuración específica por clave
   */
  async getConfiguracion(clave) {
    try {
      const query = `
        SELECT clave, valor, descripcion, tipo_dato, categoria, activa, solo_lectura
        FROM configuracion_sistema
        WHERE clave = $1 AND activa = true
      `
      
      const result = await this.query(query, [clave])
      
      if (!result.rows.length) {
        throw new Error(`Configuración '${clave}' no encontrada`)
      }

      const config = result.rows[0]
      return {
        ...config,
        valor: this.convertirValor(config.valor, config.tipo_dato)
      }
    } catch (error) {
      logger.error('Error al obtener configuración:', error)
      throw error
    }
  }

  /**
   * Actualiza una configuración del sistema
   */
  async updateConfiguracion(clave, nuevoValor, usuarioId) {
    try {
      return await this.transaction(async (client) => {
        // Verificar que la configuración existe y no es solo lectura
        const configQuery = `
          SELECT clave, valor, tipo_dato, solo_lectura, activa
          FROM configuracion_sistema
          WHERE clave = $1
        `
        
        const configResult = await client.query(configQuery, [clave])
        
        if (!configResult.rows.length) {
          throw new Error(`Configuración '${clave}' no encontrada`)
        }

        const config = configResult.rows[0]
        
        if (config.solo_lectura) {
          throw new Error(`Configuración '${clave}' es de solo lectura`)
        }

        if (!config.activa) {
          throw new Error(`Configuración '${clave}' está inactiva`)
        }

        // Validar el nuevo valor según el tipo de dato
        const valorValidado = this.validarValor(nuevoValor, config.tipo_dato)

        // Actualizar configuración
        const updateQuery = `
          UPDATE configuracion_sistema
          SET valor = $1, fecha_modificacion = NOW()
          WHERE clave = $2
          RETURNING *
        `
        
        const updateResult = await client.query(updateQuery, [valorValidado, clave])

        // Registrar cambio en auditoría
        await this.registrarCambioConfiguracion(
          clave, 
          config.valor, 
          valorValidado, 
          usuarioId, 
          client
        )

        logger.business('Configuración del sistema actualizada', {
          clave,
          valorAnterior: config.valor,
          valorNuevo: valorValidado,
          usuarioId
        })

        const configActualizada = updateResult.rows[0]
        return {
          ...configActualizada,
          valor: this.convertirValor(configActualizada.valor, configActualizada.tipo_dato)
        }
      })
    } catch (error) {
      logger.error('Error al actualizar configuración:', error)
      throw error
    }
  }

  /**
   * Crea una nueva configuración del sistema
   */
  async createConfiguracion(configData, usuarioId) {
    try {
      const {
        clave,
        valor,
        descripcion,
        tipo_dato,
        categoria,
        solo_lectura = false
      } = configData

      // Validar que la clave no existe
      const existeQuery = `
        SELECT clave FROM configuracion_sistema WHERE clave = $1
      `
      const existeResult = await this.query(existeQuery, [clave])
      
      if (existeResult.rows.length > 0) {
        throw new Error(`Ya existe una configuración con la clave '${clave}'`)
      }

      // Validar el valor según el tipo de dato
      const valorValidado = this.validarValor(valor, tipo_dato)

      const query = `
        INSERT INTO configuracion_sistema (
          clave, valor, descripcion, tipo_dato, categoria, 
          activa, solo_lectura, usuario_creacion
        ) VALUES ($1, $2, $3, $4, $5, true, $6, $7)
        RETURNING *
      `

      const result = await this.query(query, [
        clave,
        valorValidado,
        descripcion,
        tipo_dato,
        categoria,
        solo_lectura,
        usuarioId
      ])

      logger.business('Nueva configuración del sistema creada', {
        clave,
        valor: valorValidado,
        categoria,
        usuarioId
      })

      const configCreada = result.rows[0]
      return {
        ...configCreada,
        valor: this.convertirValor(configCreada.valor, configCreada.tipo_dato)
      }
    } catch (error) {
      logger.error('Error al crear configuración:', error)
      throw error
    }
  }

  /**
   * Elimina una configuración del sistema (eliminación lógica)
   */
  async deleteConfiguracion(clave, usuarioId) {
    try {
      return await this.transaction(async (client) => {
        // Verificar que la configuración existe y no es solo lectura
        const configQuery = `
          SELECT clave, solo_lectura FROM configuracion_sistema WHERE clave = $1
        `
        
        const configResult = await client.query(configQuery, [clave])
        
        if (!configResult.rows.length) {
          throw new Error(`Configuración '${clave}' no encontrada`)
        }

        const config = configResult.rows[0]
        
        if (config.solo_lectura) {
          throw new Error(`Configuración '${clave}' es de solo lectura y no puede eliminarse`)
        }

        // Desactivar configuración
        const updateQuery = `
          UPDATE configuracion_sistema
          SET activa = false, fecha_modificacion = NOW()
          WHERE clave = $1
          RETURNING *
        `
        
        const result = await client.query(updateQuery, [clave])

        // Registrar eliminación en auditoría
        await this.registrarCambioConfiguracion(
          clave, 
          'ACTIVA', 
          'ELIMINADA', 
          usuarioId, 
          client
        )

        logger.business('Configuración del sistema eliminada', {
          clave,
          usuarioId
        })

        return result.rows[0]
      })
    } catch (error) {
      logger.error('Error al eliminar configuración:', error)
      throw error
    }
  }

  /**
   * Obtiene información del sistema y estadísticas
   */
  async getInfoSistema() {
    try {
      const info = {
        version: '1.0.0',
        nombre: 'Ferre-POS API',
        descripcion: 'Sistema de Punto de Venta para Ferreterías',
        ambiente: process.env.NODE_ENV || 'development',
        fecha_inicio: new Date().toISOString(),
        uptime: process.uptime(),
        memoria: process.memoryUsage(),
        plataforma: process.platform,
        version_node: process.version
      }

      // Obtener estadísticas de base de datos
      const dbStats = await this.getEstadisticasBaseDatos()
      info.base_datos = dbStats

      // Obtener configuraciones críticas
      const configsCriticas = await this.getConfiguracionesCriticas()
      info.configuraciones = configsCriticas

      return info
    } catch (error) {
      logger.error('Error al obtener información del sistema:', error)
      throw error
    }
  }

  /**
   * Obtiene estadísticas de la base de datos
   */
  async getEstadisticasBaseDatos() {
    try {
      const queries = [
        {
          nombre: 'total_usuarios',
          query: 'SELECT COUNT(*) as count FROM usuarios WHERE activo = true'
        },
        {
          nombre: 'total_productos',
          query: 'SELECT COUNT(*) as count FROM productos WHERE activo = true'
        },
        {
          nombre: 'total_ventas_hoy',
          query: `SELECT COUNT(*) as count FROM ventas 
                  WHERE DATE(fecha) = CURRENT_DATE AND estado = 'finalizada'`
        },
        {
          nombre: 'total_sucursales',
          query: 'SELECT COUNT(*) as count FROM sucursales WHERE habilitada = true'
        },
        {
          nombre: 'total_notas_activas',
          query: 'SELECT COUNT(*) as count FROM notas_venta WHERE estado = \'activa\''
        }
      ]

      const estadisticas = {}
      
      for (const queryInfo of queries) {
        try {
          const result = await this.query(queryInfo.query)
          estadisticas[queryInfo.nombre] = parseInt(result.rows[0].count)
        } catch (error) {
          logger.warn(`Error al obtener estadística ${queryInfo.nombre}:`, error)
          estadisticas[queryInfo.nombre] = 0
        }
      }

      // Obtener tamaño de base de datos
      try {
        const sizeQuery = `
          SELECT pg_size_pretty(pg_database_size(current_database())) as size
        `
        const sizeResult = await this.query(sizeQuery)
        estadisticas.tamaño_bd = sizeResult.rows[0].size
      } catch (error) {
        estadisticas.tamaño_bd = 'No disponible'
      }

      return estadisticas
    } catch (error) {
      logger.error('Error al obtener estadísticas de base de datos:', error)
      return {}
    }
  }

  /**
   * Obtiene configuraciones críticas del sistema
   */
  async getConfiguracionesCriticas() {
    try {
      const clavesCriticas = [
        'empresa_nombre',
        'empresa_rut',
        'iva_porcentaje',
        'moneda_codigo',
        'backup_automatico',
        'logs_retention_days'
      ]

      const configuraciones = {}
      
      for (const clave of clavesCriticas) {
        try {
          const config = await this.getConfiguracion(clave)
          configuraciones[clave] = config.valor
        } catch (error) {
          configuraciones[clave] = null
        }
      }

      return configuraciones
    } catch (error) {
      logger.error('Error al obtener configuraciones críticas:', error)
      return {}
    }
  }

  /**
   * Realiza backup de configuraciones del sistema
   */
  async backupConfiguraciones(usuarioId) {
    try {
      const configuraciones = await this.getConfiguraciones(true)
      
      const backup = {
        fecha: new Date().toISOString(),
        version: '1.0.0',
        usuario_backup: usuarioId,
        configuraciones
      }

      const backupDir = path.join(process.cwd(), 'backups')
      await fs.mkdir(backupDir, { recursive: true })

      const filename = `config_backup_${new Date().toISOString().split('T')[0]}_${Date.now()}.json`
      const filepath = path.join(backupDir, filename)

      await fs.writeFile(filepath, JSON.stringify(backup, null, 2))

      logger.business('Backup de configuraciones creado', {
        archivo: filename,
        usuarioId,
        totalConfiguraciones: Object.keys(configuraciones).length
      })

      return {
        archivo: filename,
        ruta: filepath,
        tamaño: (await fs.stat(filepath)).size,
        fecha: backup.fecha
      }
    } catch (error) {
      logger.error('Error al crear backup de configuraciones:', error)
      throw error
    }
  }

  /**
   * Restaura configuraciones desde un backup
   */
  async restaurarConfiguraciones(backupData, usuarioId) {
    try {
      return await this.transaction(async (client) => {
        const { configuraciones } = backupData
        let configuracionesRestauradas = 0
        let errores = []

        for (const categoria in configuraciones) {
          for (const clave in configuraciones[categoria]) {
            try {
              const config = configuraciones[categoria][clave]
              
              // Verificar si la configuración existe
              const existeQuery = `
                SELECT clave, solo_lectura FROM configuracion_sistema WHERE clave = $1
              `
              const existeResult = await client.query(existeQuery, [clave])

              if (existeResult.rows.length > 0) {
                const configExistente = existeResult.rows[0]
                
                if (!configExistente.solo_lectura) {
                  // Actualizar configuración existente
                  await client.query(`
                    UPDATE configuracion_sistema
                    SET valor = $1, fecha_modificacion = NOW()
                    WHERE clave = $2
                  `, [config.valor.toString(), clave])
                  
                  configuracionesRestauradas++
                }
              } else {
                // Crear nueva configuración
                await client.query(`
                  INSERT INTO configuracion_sistema (
                    clave, valor, descripcion, tipo_dato, categoria,
                    activa, solo_lectura, usuario_creacion
                  ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
                `, [
                  clave,
                  config.valor.toString(),
                  config.descripcion,
                  config.tipo_dato,
                  categoria,
                  config.activa,
                  config.solo_lectura,
                  usuarioId
                ])
                
                configuracionesRestauradas++
              }
            } catch (error) {
              errores.push({
                clave,
                error: error.message
              })
            }
          }
        }

        logger.business('Configuraciones restauradas desde backup', {
          configuracionesRestauradas,
          errores: errores.length,
          usuarioId
        })

        return {
          configuracionesRestauradas,
          errores
        }
      })
    } catch (error) {
      logger.error('Error al restaurar configuraciones:', error)
      throw error
    }
  }

  /**
   * Obtiene logs del sistema con filtros
   */
  async getLogs(options = {}) {
    try {
      const {
        nivel = null,
        fechaInicio = null,
        fechaFin = null,
        usuario = null,
        modulo = null,
        limit = 100,
        offset = 0
      } = options

      // Esta implementación asume que los logs se guardan en base de datos
      // Si se usan archivos, se necesitaría una implementación diferente
      let query = `
        SELECT timestamp, level, message, meta, module, user_id
        FROM system_logs
        WHERE 1=1
      `
      const params = []
      let paramIndex = 1

      if (nivel) {
        query += ` AND level = $${paramIndex}`
        params.push(nivel)
        paramIndex++
      }

      if (fechaInicio) {
        query += ` AND timestamp >= $${paramIndex}`
        params.push(fechaInicio)
        paramIndex++
      }

      if (fechaFin) {
        query += ` AND timestamp <= $${paramIndex}`
        params.push(fechaFin)
        paramIndex++
      }

      if (usuario) {
        query += ` AND user_id = $${paramIndex}`
        params.push(usuario)
        paramIndex++
      }

      if (modulo) {
        query += ` AND module = $${paramIndex}`
        params.push(modulo)
        paramIndex++
      }

      query += ` ORDER BY timestamp DESC LIMIT $${paramIndex} OFFSET $${paramIndex + 1}`
      params.push(limit, offset)

      const result = await this.query(query, params)
      return result.rows
    } catch (error) {
      logger.error('Error al obtener logs del sistema:', error)
      // Si falla la consulta a BD, intentar leer desde archivos
      return this.getLogsFromFile(options)
    }
  }

  /**
   * Obtiene logs desde archivos (fallback)
   */
  async getLogsFromFile(options = {}) {
    try {
      const { limit = 100 } = options
      const logFile = path.join(process.cwd(), 'logs', 'ferre-pos-api.log')
      
      try {
        const logContent = await fs.readFile(logFile, 'utf8')
        const lines = logContent.split('\n').filter(line => line.trim())
        
        // Tomar las últimas líneas
        const recentLines = lines.slice(-limit).reverse()
        
        return recentLines.map(line => {
          try {
            return JSON.parse(line)
          } catch {
            return { message: line, timestamp: new Date().toISOString() }
          }
        })
      } catch (fileError) {
        return []
      }
    } catch (error) {
      logger.error('Error al leer logs desde archivo:', error)
      return []
    }
  }

  /**
   * Limpia logs antiguos del sistema
   */
  async limpiarLogs(diasRetencion = 30, usuarioId) {
    try {
      const fechaLimite = new Date()
      fechaLimite.setDate(fechaLimite.getDate() - diasRetencion)

      // Limpiar logs de base de datos
      const deleteQuery = `
        DELETE FROM system_logs
        WHERE timestamp < $1
      `
      
      const result = await this.query(deleteQuery, [fechaLimite])
      const logsEliminados = result.rowCount || 0

      logger.business('Limpieza de logs del sistema', {
        logsEliminados,
        diasRetencion,
        fechaLimite: fechaLimite.toISOString(),
        usuarioId
      })

      return {
        logsEliminados,
        fechaLimite
      }
    } catch (error) {
      logger.error('Error al limpiar logs del sistema:', error)
      throw error
    }
  }

  /**
   * Registra un cambio de configuración en auditoría
   */
  async registrarCambioConfiguracion(clave, valorAnterior, valorNuevo, usuarioId, client = null) {
    try {
      const queryClient = client || this

      const query = `
        INSERT INTO auditoria_configuracion (
          clave_configuracion, valor_anterior, valor_nuevo, usuario_id, fecha_cambio
        ) VALUES ($1, $2, $3, $4, NOW())
      `

      await queryClient.query(query, [clave, valorAnterior, valorNuevo, usuarioId])
    } catch (error) {
      logger.error('Error al registrar cambio de configuración:', error)
      // No lanzar error para no afectar la operación principal
    }
  }

  /**
   * Convierte un valor string al tipo de dato correspondiente
   */
  convertirValor(valor, tipoDato) {
    if (valor === null || valor === undefined) return null

    switch (tipoDato) {
      case 'boolean':
        return valor === 'true' || valor === true
      case 'number':
        return parseFloat(valor)
      case 'integer':
        return parseInt(valor)
      case 'json':
        try {
          return JSON.parse(valor)
        } catch {
          return valor
        }
      case 'string':
      default:
        return valor.toString()
    }
  }

  /**
   * Valida un valor según el tipo de dato
   */
  validarValor(valor, tipoDato) {
    switch (tipoDato) {
      case 'boolean':
        if (typeof valor !== 'boolean' && valor !== 'true' && valor !== 'false') {
          throw new Error(`Valor debe ser boolean para tipo ${tipoDato}`)
        }
        return valor.toString()
      
      case 'number':
        const num = parseFloat(valor)
        if (isNaN(num)) {
          throw new Error(`Valor debe ser numérico para tipo ${tipoDato}`)
        }
        return num.toString()
      
      case 'integer':
        const int = parseInt(valor)
        if (isNaN(int) || !Number.isInteger(parseFloat(valor))) {
          throw new Error(`Valor debe ser entero para tipo ${tipoDato}`)
        }
        return int.toString()
      
      case 'json':
        try {
          JSON.parse(typeof valor === 'string' ? valor : JSON.stringify(valor))
          return typeof valor === 'string' ? valor : JSON.stringify(valor)
        } catch {
          throw new Error(`Valor debe ser JSON válido para tipo ${tipoDato}`)
        }
      
      case 'string':
      default:
        return valor.toString()
    }
  }

  /**
   * Obtiene métricas de rendimiento del sistema
   */
  async getMetricasRendimiento() {
    try {
      const metricas = {
        memoria: process.memoryUsage(),
        cpu: process.cpuUsage(),
        uptime: process.uptime(),
        timestamp: new Date().toISOString()
      }

      // Métricas de base de datos
      try {
        const dbMetricas = await this.query(`
          SELECT 
            (SELECT COUNT(*) FROM pg_stat_activity WHERE state = 'active') as conexiones_activas,
            (SELECT COUNT(*) FROM pg_stat_activity) as conexiones_totales
        `)
        
        metricas.base_datos = dbMetricas.rows[0]
      } catch (error) {
        metricas.base_datos = { error: 'No disponible' }
      }

      return metricas
    } catch (error) {
      logger.error('Error al obtener métricas de rendimiento:', error)
      throw error
    }
  }
}

module.exports = new Sistema()

