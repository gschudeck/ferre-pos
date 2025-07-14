/**
 * Configuración de Base de Datos - Sistema Ferre-POS
 * 
 * Maneja la conexión a PostgreSQL con pool de conexiones,
 * transacciones y utilidades para consultas.
 */

const { Pool } = require('pg')
const config = require('./index')
const logger = require('../utils/logger')

class Database {
  constructor() {
    this.pool = null
    this.isConnected = false
  }

  /**
   * Inicializa la conexión a la base de datos
   */
  async connect() {
    try {
      const dbConfig = config.getDatabaseConfig()
      
      this.pool = new Pool({
        host: dbConfig.host,
        port: dbConfig.port,
        database: dbConfig.database,
        user: dbConfig.user,
        password: dbConfig.password,
        ssl: dbConfig.ssl ? { rejectUnauthorized: false } : false,
        min: dbConfig.min,
        max: dbConfig.max,
        idleTimeoutMillis: dbConfig.idleTimeoutMillis,
        connectionTimeoutMillis: dbConfig.connectionTimeoutMillis,
        statement_timeout: config.database.query.timeout,
        query_timeout: config.database.query.timeout
      })

      // Configurar eventos del pool
      this.pool.on('connect', (client) => {
        logger.debug('Nueva conexión establecida a PostgreSQL')
      })

      this.pool.on('error', (err, client) => {
        logger.error('Error en conexión de PostgreSQL:', err)
      })

      this.pool.on('remove', (client) => {
        logger.debug('Conexión removida del pool de PostgreSQL')
      })

      // Probar la conexión
      const client = await this.pool.connect()
      const result = await client.query('SELECT NOW() as current_time, version() as version')
      client.release()

      this.isConnected = true
      logger.info('Conexión a PostgreSQL establecida exitosamente', {
        host: dbConfig.host,
        port: dbConfig.port,
        database: dbConfig.database,
        version: result.rows[0].version.split(' ')[1],
        poolSize: `${dbConfig.min}-${dbConfig.max}`
      })

      return this.pool
    } catch (error) {
      logger.error('Error al conectar con PostgreSQL:', error)
      throw error
    }
  }

  /**
   * Cierra todas las conexiones del pool
   */
  async disconnect() {
    if (this.pool) {
      await this.pool.end()
      this.isConnected = false
      logger.info('Conexiones a PostgreSQL cerradas')
    }
  }

  /**
   * Obtiene una conexión del pool
   */
  async getClient() {
    if (!this.pool) {
      throw new Error('Base de datos no inicializada')
    }
    return await this.pool.connect()
  }

  /**
   * Ejecuta una consulta simple
   */
  async query(text, params = []) {
    const start = Date.now()
    try {
      const result = await this.pool.query(text, params)
      const duration = Date.now() - start
      
      logger.debug('Consulta ejecutada', {
        query: text.substring(0, 100) + (text.length > 100 ? '...' : ''),
        duration: `${duration}ms`,
        rows: result.rowCount
      })
      
      return result
    } catch (error) {
      const duration = Date.now() - start
      logger.error('Error en consulta SQL', {
        query: text.substring(0, 100) + (text.length > 100 ? '...' : ''),
        params: params,
        duration: `${duration}ms`,
        error: error.message
      })
      throw error
    }
  }

  /**
   * Ejecuta múltiples consultas en una transacción
   */
  async transaction(callback) {
    const client = await this.getClient()
    
    try {
      await client.query('BEGIN')
      logger.debug('Transacción iniciada')
      
      const result = await callback(client)
      
      await client.query('COMMIT')
      logger.debug('Transacción confirmada')
      
      return result
    } catch (error) {
      await client.query('ROLLBACK')
      logger.error('Transacción revertida:', error)
      throw error
    } finally {
      client.release()
    }
  }

  /**
   * Verifica el estado de la conexión
   */
  async healthCheck() {
    try {
      const result = await this.query('SELECT 1 as health_check')
      return {
        status: 'healthy',
        connected: this.isConnected,
        timestamp: new Date().toISOString(),
        poolSize: this.pool ? this.pool.totalCount : 0,
        idleConnections: this.pool ? this.pool.idleCount : 0,
        waitingClients: this.pool ? this.pool.waitingCount : 0
      }
    } catch (error) {
      return {
        status: 'unhealthy',
        connected: false,
        error: error.message,
        timestamp: new Date().toISOString()
      }
    }
  }

  /**
   * Obtiene estadísticas del pool de conexiones
   */
  getPoolStats() {
    if (!this.pool) {
      return null
    }

    return {
      totalCount: this.pool.totalCount,
      idleCount: this.pool.idleCount,
      waitingCount: this.pool.waitingCount,
      maxConnections: this.pool.options.max,
      minConnections: this.pool.options.min
    }
  }

  /**
   * Ejecuta una consulta preparada con parámetros nombrados
   */
  async queryNamed(text, namedParams = {}) {
    // Convertir parámetros nombrados a posicionales
    let paramIndex = 1
    const params = []
    const paramMap = {}

    const queryText = text.replace(/:(\w+)/g, (match, paramName) => {
      if (namedParams.hasOwnProperty(paramName)) {
        if (!paramMap[paramName]) {
          paramMap[paramName] = paramIndex++
          params.push(namedParams[paramName])
        }
        return `$${paramMap[paramName]}`
      }
      throw new Error(`Parámetro no encontrado: ${paramName}`)
    })

    return await this.query(queryText, params)
  }

  /**
   * Ejecuta una consulta con paginación
   */
  async queryPaginated(baseQuery, params = [], options = {}) {
    const {
      page = 1,
      limit = config.pagination.defaultLimit,
      orderBy = 'id',
      orderDirection = 'ASC'
    } = options

    const offset = (page - 1) * limit
    const limitValue = Math.min(limit, config.pagination.maxLimit)

    // Consulta para contar total de registros
    const countQuery = `SELECT COUNT(*) as total FROM (${baseQuery}) as count_query`
    const countResult = await this.query(countQuery, params)
    const total = parseInt(countResult.rows[0].total)

    // Consulta con paginación
    const paginatedQuery = `
      ${baseQuery}
      ORDER BY ${orderBy} ${orderDirection}
      LIMIT $${params.length + 1} OFFSET $${params.length + 2}
    `
    
    const result = await this.query(paginatedQuery, [...params, limitValue, offset])

    return {
      data: result.rows,
      pagination: {
        page,
        limit: limitValue,
        total,
        totalPages: Math.ceil(total / limitValue),
        hasNext: page < Math.ceil(total / limitValue),
        hasPrev: page > 1
      }
    }
  }

  /**
   * Ejecuta una función almacenada
   */
  async callFunction(functionName, params = []) {
    const placeholders = params.map((_, index) => `$${index + 1}`).join(', ')
    const query = `SELECT * FROM ${functionName}(${placeholders})`
    return await this.query(query, params)
  }

  /**
   * Inserta un registro y retorna el ID generado
   */
  async insertAndGetId(table, data, idColumn = 'id') {
    const columns = Object.keys(data)
    const values = Object.values(data)
    const placeholders = values.map((_, index) => `$${index + 1}`).join(', ')
    
    const query = `
      INSERT INTO ${table} (${columns.join(', ')})
      VALUES (${placeholders})
      RETURNING ${idColumn}
    `
    
    const result = await this.query(query, values)
    return result.rows[0][idColumn]
  }

  /**
   * Actualiza un registro por ID
   */
  async updateById(table, id, data, idColumn = 'id') {
    const columns = Object.keys(data)
    const values = Object.values(data)
    const setClause = columns.map((col, index) => `${col} = $${index + 2}`).join(', ')
    
    const query = `
      UPDATE ${table}
      SET ${setClause}, fecha_modificacion = NOW()
      WHERE ${idColumn} = $1
      RETURNING *
    `
    
    const result = await this.query(query, [id, ...values])
    return result.rows[0]
  }

  /**
   * Elimina un registro por ID (eliminación lógica si existe campo 'activo')
   */
  async deleteById(table, id, idColumn = 'id', soft = true) {
    if (soft) {
      // Intentar eliminación lógica primero
      try {
        const query = `
          UPDATE ${table}
          SET activo = false, fecha_modificacion = NOW()
          WHERE ${idColumn} = $1
          RETURNING *
        `
        const result = await this.query(query, [id])
        if (result.rows.length > 0) {
          return result.rows[0]
        }
      } catch (error) {
        // Si no existe campo 'activo', continuar con eliminación física
      }
    }

    // Eliminación física
    const query = `DELETE FROM ${table} WHERE ${idColumn} = $1 RETURNING *`
    const result = await this.query(query, [id])
    return result.rows[0]
  }
}

// Instancia singleton de la base de datos
const database = new Database()

module.exports = database

