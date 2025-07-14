/**
 * Modelo Base - Sistema Ferre-POS
 * 
 * Clase base que proporciona funcionalidades comunes para todos los modelos
 * incluyendo CRUD básico, validaciones y utilidades.
 */

const database = require('../config/database')
const logger = require('../utils/logger')
const { v4: uuidv4 } = require('uuid')

class BaseModel {
  constructor(tableName, schema = {}) {
    this.tableName = tableName
    this.schema = schema
    this.primaryKey = 'id'
    this.timestamps = true
    this.softDelete = true
  }

  /**
   * Busca un registro por ID
   */
  async findById(id, options = {}) {
    try {
      const { includeInactive = false } = options
      
      let query = `SELECT * FROM ${this.tableName} WHERE ${this.primaryKey} = $1`
      
      if (this.softDelete && !includeInactive) {
        query += ' AND activo = true'
      }
      
      const result = await database.query(query, [id])
      return result.rows[0] || null
    } catch (error) {
      logger.error(`Error al buscar ${this.tableName} por ID:`, error)
      throw error
    }
  }

  /**
   * Busca todos los registros con filtros opcionales
   */
  async findAll(options = {}) {
    try {
      const {
        where = {},
        orderBy = this.primaryKey,
        orderDirection = 'ASC',
        limit,
        offset = 0,
        includeInactive = false
      } = options

      let query = `SELECT * FROM ${this.tableName}`
      const params = []
      const conditions = []

      // Agregar condiciones WHERE
      if (this.softDelete && !includeInactive) {
        conditions.push('activo = true')
      }

      Object.entries(where).forEach(([key, value], index) => {
        if (value !== undefined && value !== null) {
          conditions.push(`${key} = $${params.length + 1}`)
          params.push(value)
        }
      })

      if (conditions.length > 0) {
        query += ` WHERE ${conditions.join(' AND ')}`
      }

      // Agregar ORDER BY
      query += ` ORDER BY ${orderBy} ${orderDirection}`

      // Agregar LIMIT y OFFSET
      if (limit) {
        query += ` LIMIT $${params.length + 1}`
        params.push(limit)
      }

      if (offset > 0) {
        query += ` OFFSET $${params.length + 1}`
        params.push(offset)
      }

      const result = await database.query(query, params)
      return result.rows
    } catch (error) {
      logger.error(`Error al buscar registros en ${this.tableName}:`, error)
      throw error
    }
  }

  /**
   * Busca registros con paginación
   */
  async findPaginated(options = {}) {
    try {
      const {
        where = {},
        page = 1,
        limit = 20,
        orderBy = this.primaryKey,
        orderDirection = 'ASC',
        includeInactive = false
      } = options

      const offset = (page - 1) * limit
      
      // Construir query base para contar
      let baseQuery = `FROM ${this.tableName}`
      const params = []
      const conditions = []

      if (this.softDelete && !includeInactive) {
        conditions.push('activo = true')
      }

      Object.entries(where).forEach(([key, value]) => {
        if (value !== undefined && value !== null) {
          conditions.push(`${key} = $${params.length + 1}`)
          params.push(value)
        }
      })

      if (conditions.length > 0) {
        baseQuery += ` WHERE ${conditions.join(' AND ')}`
      }

      // Contar total de registros
      const countQuery = `SELECT COUNT(*) as total ${baseQuery}`
      const countResult = await database.query(countQuery, params)
      const total = parseInt(countResult.rows[0].total)

      // Obtener registros paginados
      const dataQuery = `SELECT * ${baseQuery} ORDER BY ${orderBy} ${orderDirection} LIMIT $${params.length + 1} OFFSET $${params.length + 2}`
      const dataResult = await database.query(dataQuery, [...params, limit, offset])

      return {
        data: dataResult.rows,
        pagination: {
          page,
          limit,
          total,
          totalPages: Math.ceil(total / limit),
          hasNext: page < Math.ceil(total / limit),
          hasPrev: page > 1
        }
      }
    } catch (error) {
      logger.error(`Error en paginación de ${this.tableName}:`, error)
      throw error
    }
  }

  /**
   * Busca un registro por criterios específicos
   */
  async findOne(where = {}, options = {}) {
    try {
      const { includeInactive = false } = options
      
      const conditions = []
      const params = []

      if (this.softDelete && !includeInactive) {
        conditions.push('activo = true')
      }

      Object.entries(where).forEach(([key, value]) => {
        if (value !== undefined && value !== null) {
          conditions.push(`${key} = $${params.length + 1}`)
          params.push(value)
        }
      })

      if (conditions.length === 0) {
        throw new Error('Se requiere al menos una condición WHERE')
      }

      const query = `SELECT * FROM ${this.tableName} WHERE ${conditions.join(' AND ')} LIMIT 1`
      const result = await database.query(query, params)
      
      return result.rows[0] || null
    } catch (error) {
      logger.error(`Error al buscar registro en ${this.tableName}:`, error)
      throw error
    }
  }

  /**
   * Crea un nuevo registro
   */
  async create(data, options = {}) {
    try {
      const { returning = '*' } = options
      
      // Agregar ID si no existe
      if (!data[this.primaryKey]) {
        data[this.primaryKey] = uuidv4()
      }

      // Agregar timestamps si están habilitados
      if (this.timestamps) {
        data.fecha_creacion = new Date()
        data.fecha_modificacion = new Date()
      }

      // Validar datos si hay esquema definido
      if (Object.keys(this.schema).length > 0) {
        this.validateData(data)
      }

      const columns = Object.keys(data)
      const values = Object.values(data)
      const placeholders = values.map((_, index) => `$${index + 1}`).join(', ')

      const query = `
        INSERT INTO ${this.tableName} (${columns.join(', ')})
        VALUES (${placeholders})
        RETURNING ${returning}
      `

      const result = await database.query(query, values)
      
      logger.audit(`Registro creado en ${this.tableName}`, {
        id: result.rows[0][this.primaryKey],
        table: this.tableName
      })

      return result.rows[0]
    } catch (error) {
      logger.error(`Error al crear registro en ${this.tableName}:`, error)
      throw error
    }
  }

  /**
   * Actualiza un registro por ID
   */
  async updateById(id, data, options = {}) {
    try {
      const { returning = '*' } = options
      
      // Agregar timestamp de modificación
      if (this.timestamps) {
        data.fecha_modificacion = new Date()
      }

      // Validar datos si hay esquema definido
      if (Object.keys(this.schema).length > 0) {
        this.validateData(data, true)
      }

      const columns = Object.keys(data)
      const values = Object.values(data)
      const setClause = columns.map((col, index) => `${col} = $${index + 2}`).join(', ')

      const query = `
        UPDATE ${this.tableName}
        SET ${setClause}
        WHERE ${this.primaryKey} = $1
        RETURNING ${returning}
      `

      const result = await database.query(query, [id, ...values])
      
      if (result.rows.length === 0) {
        throw new Error(`Registro con ID ${id} no encontrado`)
      }

      logger.audit(`Registro actualizado en ${this.tableName}`, {
        id,
        table: this.tableName,
        changes: Object.keys(data)
      })

      return result.rows[0]
    } catch (error) {
      logger.error(`Error al actualizar registro en ${this.tableName}:`, error)
      throw error
    }
  }

  /**
   * Elimina un registro por ID (lógica o física)
   */
  async deleteById(id, options = {}) {
    try {
      const { soft = this.softDelete, returning = '*' } = options
      
      let query
      let params = [id]

      if (soft) {
        // Eliminación lógica
        query = `
          UPDATE ${this.tableName}
          SET activo = false, fecha_modificacion = NOW()
          WHERE ${this.primaryKey} = $1
          RETURNING ${returning}
        `
      } else {
        // Eliminación física
        query = `
          DELETE FROM ${this.tableName}
          WHERE ${this.primaryKey} = $1
          RETURNING ${returning}
        `
      }

      const result = await database.query(query, params)
      
      if (result.rows.length === 0) {
        throw new Error(`Registro con ID ${id} no encontrado`)
      }

      logger.audit(`Registro eliminado en ${this.tableName}`, {
        id,
        table: this.tableName,
        soft
      })

      return result.rows[0]
    } catch (error) {
      logger.error(`Error al eliminar registro en ${this.tableName}:`, error)
      throw error
    }
  }

  /**
   * Cuenta registros con filtros opcionales
   */
  async count(where = {}, options = {}) {
    try {
      const { includeInactive = false } = options
      
      let query = `SELECT COUNT(*) as total FROM ${this.tableName}`
      const params = []
      const conditions = []

      if (this.softDelete && !includeInactive) {
        conditions.push('activo = true')
      }

      Object.entries(where).forEach(([key, value]) => {
        if (value !== undefined && value !== null) {
          conditions.push(`${key} = $${params.length + 1}`)
          params.push(value)
        }
      })

      if (conditions.length > 0) {
        query += ` WHERE ${conditions.join(' AND ')}`
      }

      const result = await database.query(query, params)
      return parseInt(result.rows[0].total)
    } catch (error) {
      logger.error(`Error al contar registros en ${this.tableName}:`, error)
      throw error
    }
  }

  /**
   * Verifica si existe un registro
   */
  async exists(where = {}, options = {}) {
    try {
      const count = await this.count(where, options)
      return count > 0
    } catch (error) {
      logger.error(`Error al verificar existencia en ${this.tableName}:`, error)
      throw error
    }
  }

  /**
   * Ejecuta una consulta SQL personalizada
   */
  async query(sql, params = []) {
    try {
      return await database.query(sql, params)
    } catch (error) {
      logger.error(`Error en consulta personalizada para ${this.tableName}:`, error)
      throw error
    }
  }

  /**
   * Ejecuta múltiples operaciones en una transacción
   */
  async transaction(callback) {
    try {
      return await database.transaction(callback)
    } catch (error) {
      logger.error(`Error en transacción para ${this.tableName}:`, error)
      throw error
    }
  }

  /**
   * Valida datos según el esquema definido
   */
  validateData(data, isUpdate = false) {
    // Implementación básica de validación
    // En una implementación real, se usaría Joi o similar
    
    for (const [field, rules] of Object.entries(this.schema)) {
      const value = data[field]
      
      // Verificar campos requeridos
      if (rules.required && !isUpdate && (value === undefined || value === null)) {
        throw new Error(`Campo requerido: ${field}`)
      }
      
      // Verificar tipos de datos
      if (value !== undefined && value !== null && rules.type) {
        const actualType = typeof value
        if (actualType !== rules.type) {
          throw new Error(`Tipo incorrecto para ${field}: esperado ${rules.type}, recibido ${actualType}`)
        }
      }
      
      // Verificar longitud máxima
      if (value && rules.maxLength && value.length > rules.maxLength) {
        throw new Error(`${field} excede la longitud máxima de ${rules.maxLength}`)
      }
    }
  }

  /**
   * Obtiene estadísticas básicas de la tabla
   */
  async getStats() {
    try {
      const totalQuery = `SELECT COUNT(*) as total FROM ${this.tableName}`
      const totalResult = await database.query(totalQuery)
      
      const stats = {
        total: parseInt(totalResult.rows[0].total)
      }

      if (this.softDelete) {
        const activeQuery = `SELECT COUNT(*) as active FROM ${this.tableName} WHERE activo = true`
        const activeResult = await database.query(activeQuery)
        stats.active = parseInt(activeResult.rows[0].active)
        stats.inactive = stats.total - stats.active
      }

      if (this.timestamps) {
        const recentQuery = `
          SELECT COUNT(*) as recent 
          FROM ${this.tableName} 
          WHERE fecha_creacion >= NOW() - INTERVAL '24 hours'
        `
        const recentResult = await database.query(recentQuery)
        stats.recent = parseInt(recentResult.rows[0].recent)
      }

      return stats
    } catch (error) {
      logger.error(`Error al obtener estadísticas de ${this.tableName}:`, error)
      throw error
    }
  }
}

module.exports = BaseModel

