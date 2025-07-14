/**
 * Modelo Producto - Sistema Ferre-POS
 * 
 * Maneja todas las operaciones relacionadas con el catálogo de productos,
 * incluyendo búsquedas avanzadas, gestión de códigos de barras y categorización.
 */

const BaseModel = require('./BaseModel')
const logger = require('../utils/logger')

class Producto extends BaseModel {
  constructor() {
    super('productos', {
      codigo_interno: { type: 'string', required: true, maxLength: 50 },
      codigo_barra: { type: 'string', required: true, maxLength: 50 },
      descripcion: { type: 'string', required: true, maxLength: 500 },
      descripcion_corta: { type: 'string', maxLength: 100 },
      categoria_id: { type: 'string' },
      marca: { type: 'string', maxLength: 100 },
      modelo: { type: 'string', maxLength: 100 },
      precio_unitario: { type: 'number', required: true },
      precio_costo: { type: 'number' },
      unidad_medida: { type: 'string', required: true, maxLength: 10 },
      peso: { type: 'number' },
      stock_minimo: { type: 'number' },
      stock_maximo: { type: 'number' }
    })
  }

  /**
   * Crea un nuevo producto con validaciones específicas
   */
  async create(productData, options = {}) {
    try {
      // Validar código interno único
      const existingInternal = await this.findByCodigoInterno(productData.codigo_interno)
      if (existingInternal) {
        throw new Error('Ya existe un producto con este código interno')
      }

      // Validar código de barras único
      const existingBarcode = await this.findByCodigoBarra(productData.codigo_barra)
      if (existingBarcode) {
        throw new Error('Ya existe un producto con este código de barras')
      }

      // Validar precios
      if (productData.precio_unitario <= 0) {
        throw new Error('El precio unitario debe ser mayor a 0')
      }

      if (productData.precio_costo && productData.precio_costo < 0) {
        throw new Error('El precio de costo no puede ser negativo')
      }

      // Validar stock
      if (productData.stock_minimo < 0) {
        throw new Error('El stock mínimo no puede ser negativo')
      }

      if (productData.stock_maximo && productData.stock_maximo < productData.stock_minimo) {
        throw new Error('El stock máximo no puede ser menor al stock mínimo')
      }

      const newProduct = await super.create(productData, options)

      logger.business('Producto creado', {
        productId: newProduct.id,
        codigoInterno: newProduct.codigo_interno,
        descripcion: newProduct.descripcion,
        precio: newProduct.precio_unitario
      })

      return newProduct
    } catch (error) {
      logger.error('Error al crear producto:', error)
      throw error
    }
  }

  /**
   * Busca un producto por código interno
   */
  async findByCodigoInterno(codigoInterno, options = {}) {
    try {
      return await this.findOne({ codigo_interno: codigoInterno }, options)
    } catch (error) {
      logger.error('Error al buscar producto por código interno:', error)
      throw error
    }
  }

  /**
   * Busca un producto por código de barras
   */
  async findByCodigoBarra(codigoBarra, options = {}) {
    try {
      // Buscar en tabla principal
      let product = await this.findOne({ codigo_barra: codigoBarra }, options)
      
      // Si no se encuentra, buscar en códigos adicionales
      if (!product) {
        const query = `
          SELECT p.*
          FROM productos p
          JOIN codigos_barra_adicionales cba ON p.id = cba.producto_id
          WHERE cba.codigo_barra = $1 AND cba.activo = true AND p.activo = true
        `
        const result = await this.query(query, [codigoBarra])
        product = result.rows[0] || null
      }

      return product
    } catch (error) {
      logger.error('Error al buscar producto por código de barras:', error)
      throw error
    }
  }

  /**
   * Busca productos por categoría
   */
  async findByCategoria(categoriaId, options = {}) {
    try {
      return await this.findAll({
        where: { categoria_id: categoriaId },
        ...options
      })
    } catch (error) {
      logger.error('Error al buscar productos por categoría:', error)
      throw error
    }
  }

  /**
   * Busca productos por marca
   */
  async findByMarca(marca, options = {}) {
    try {
      return await this.findAll({
        where: { marca },
        ...options
      })
    } catch (error) {
      logger.error('Error al buscar productos por marca:', error)
      throw error
    }
  }

  /**
   * Búsqueda avanzada de productos con múltiples filtros
   */
  async search(filters = {}, options = {}) {
    try {
      const {
        q, // Término de búsqueda general
        codigo,
        descripcion,
        marca,
        categoria,
        precioMin,
        precioMax,
        conStock = false,
        sucursalId
      } = filters

      let query = `
        SELECT DISTINCT p.*, 
               c.nombre as categoria_nombre,
               ${conStock && sucursalId ? 'sc.cantidad as stock_actual,' : ''}
               ${conStock && sucursalId ? 'sc.cantidad_disponible as stock_disponible,' : ''}
               CASE 
                 WHEN p.codigo_interno ILIKE $1 THEN 1
                 WHEN p.codigo_barra ILIKE $1 THEN 2
                 WHEN p.descripcion ILIKE $1 THEN 3
                 WHEN p.marca ILIKE $1 THEN 4
                 ELSE 5
               END as relevancia
        FROM productos p
        LEFT JOIN categorias_productos c ON p.categoria_id = c.id
        ${conStock && sucursalId ? 'LEFT JOIN stock_central sc ON p.id = sc.producto_id AND sc.sucursal_id = $' + (sucursalId ? '2' : '1') : ''}
        WHERE p.activo = true
      `

      const params = []
      let paramIndex = 1

      // Búsqueda general
      if (q) {
        query += ` AND (
          p.codigo_interno ILIKE $${paramIndex} OR
          p.codigo_barra ILIKE $${paramIndex} OR
          p.descripcion ILIKE $${paramIndex} OR
          p.descripcion_corta ILIKE $${paramIndex} OR
          p.marca ILIKE $${paramIndex} OR
          p.modelo ILIKE $${paramIndex}
        )`
        params.push(`%${q}%`)
        paramIndex++
      }

      // Filtro por sucursal para stock
      if (conStock && sucursalId) {
        params.push(sucursalId)
        paramIndex++
      }

      // Filtros específicos
      if (codigo) {
        query += ` AND (p.codigo_interno ILIKE $${paramIndex} OR p.codigo_barra ILIKE $${paramIndex})`
        params.push(`%${codigo}%`)
        paramIndex++
      }

      if (descripcion) {
        query += ` AND p.descripcion ILIKE $${paramIndex}`
        params.push(`%${descripcion}%`)
        paramIndex++
      }

      if (marca) {
        query += ` AND p.marca ILIKE $${paramIndex}`
        params.push(`%${marca}%`)
        paramIndex++
      }

      if (categoria) {
        query += ` AND p.categoria_id = $${paramIndex}`
        params.push(categoria)
        paramIndex++
      }

      if (precioMin !== undefined) {
        query += ` AND p.precio_unitario >= $${paramIndex}`
        params.push(precioMin)
        paramIndex++
      }

      if (precioMax !== undefined) {
        query += ` AND p.precio_unitario <= $${paramIndex}`
        params.push(precioMax)
        paramIndex++
      }

      if (conStock && sucursalId) {
        query += ` AND sc.cantidad_disponible > 0`
      }

      // Ordenamiento
      const { orderBy = 'relevancia, p.descripcion', orderDirection = 'ASC' } = options
      query += ` ORDER BY ${orderBy} ${orderDirection}`

      // Paginación
      const { limit, offset } = options
      if (limit) {
        query += ` LIMIT $${paramIndex}`
        params.push(limit)
        paramIndex++
      }

      if (offset) {
        query += ` OFFSET $${paramIndex}`
        params.push(offset)
        paramIndex++
      }

      const result = await this.query(query, params)
      return result.rows
    } catch (error) {
      logger.error('Error en búsqueda de productos:', error)
      throw error
    }
  }

  /**
   * Obtiene productos con stock bajo
   */
  async getProductosStockBajo(sucursalId = null) {
    try {
      let query = `
        SELECT p.*, sc.cantidad, sc.cantidad_disponible, s.nombre as sucursal_nombre
        FROM productos p
        JOIN stock_central sc ON p.id = sc.producto_id
        JOIN sucursales s ON sc.sucursal_id = s.id
        WHERE p.activo = true 
        AND s.habilitada = true
        AND sc.cantidad_disponible <= p.stock_minimo
      `
      
      const params = []
      
      if (sucursalId) {
        query += ` AND sc.sucursal_id = $1`
        params.push(sucursalId)
      }

      query += ` ORDER BY sc.cantidad_disponible ASC, p.descripcion`

      const result = await this.query(query, params)
      return result.rows
    } catch (error) {
      logger.error('Error al obtener productos con stock bajo:', error)
      throw error
    }
  }

  /**
   * Obtiene productos más vendidos
   */
  async getProductosMasVendidos(options = {}) {
    try {
      const {
        sucursalId,
        fechaInicio = new Date(Date.now() - 30 * 24 * 60 * 60 * 1000), // 30 días atrás
        fechaFin = new Date(),
        limit = 20
      } = options

      let query = `
        SELECT 
          p.*,
          SUM(dv.cantidad) as cantidad_vendida,
          COUNT(DISTINCT dv.venta_id) as numero_ventas,
          SUM(dv.total_item) as monto_total_vendido,
          AVG(dv.precio_final) as precio_promedio_venta
        FROM productos p
        JOIN detalle_ventas dv ON p.id = dv.producto_id
        JOIN ventas v ON dv.venta_id = v.id
        WHERE p.activo = true 
        AND v.estado = 'finalizada'
        AND v.fecha >= $1 
        AND v.fecha <= $2
      `

      const params = [fechaInicio, fechaFin]

      if (sucursalId) {
        query += ` AND v.sucursal_id = $3`
        params.push(sucursalId)
      }

      query += `
        GROUP BY p.id
        ORDER BY cantidad_vendida DESC
        LIMIT $${params.length + 1}
      `
      params.push(limit)

      const result = await this.query(query, params)
      return result.rows
    } catch (error) {
      logger.error('Error al obtener productos más vendidos:', error)
      throw error
    }
  }

  /**
   * Actualiza el precio de un producto
   */
  async updatePrecio(productId, nuevoPrecio, usuarioId) {
    try {
      const producto = await this.findById(productId)
      if (!producto) {
        throw new Error('Producto no encontrado')
      }

      const precioAnterior = producto.precio_unitario

      await this.updateById(productId, {
        precio_unitario: nuevoPrecio,
        usuario_modificacion: usuarioId
      })

      logger.business('Precio de producto actualizado', {
        productId,
        codigoInterno: producto.codigo_interno,
        precioAnterior,
        precioNuevo: nuevoPrecio,
        usuarioId
      })

      return true
    } catch (error) {
      logger.error('Error al actualizar precio de producto:', error)
      throw error
    }
  }

  /**
   * Agrega un código de barras adicional a un producto
   */
  async addCodigoBarraAdicional(productId, codigoBarra, descripcion = null) {
    try {
      // Verificar que el código no existe
      const existingCode = await this.findByCodigoBarra(codigoBarra)
      if (existingCode) {
        throw new Error('Este código de barras ya está en uso')
      }

      const query = `
        INSERT INTO codigos_barra_adicionales (producto_id, codigo_barra, descripcion)
        VALUES ($1, $2, $3)
        RETURNING *
      `

      const result = await this.query(query, [productId, codigoBarra, descripcion])

      logger.business('Código de barras adicional agregado', {
        productId,
        codigoBarra,
        descripcion
      })

      return result.rows[0]
    } catch (error) {
      logger.error('Error al agregar código de barras adicional:', error)
      throw error
    }
  }

  /**
   * Obtiene todos los códigos de barras de un producto
   */
  async getCodigosBarras(productId) {
    try {
      const query = `
        SELECT codigo_barra as codigo, 'principal' as tipo, null as descripcion
        FROM productos 
        WHERE id = $1 AND activo = true
        UNION ALL
        SELECT codigo_barra as codigo, 'adicional' as tipo, descripcion
        FROM codigos_barra_adicionales 
        WHERE producto_id = $1 AND activo = true
        ORDER BY tipo, codigo
      `

      const result = await this.query(query, [productId])
      return result.rows
    } catch (error) {
      logger.error('Error al obtener códigos de barras:', error)
      throw error
    }
  }

  /**
   * Obtiene estadísticas del catálogo de productos
   */
  async getProductStats() {
    try {
      const baseStats = await this.getStats()

      // Estadísticas por categoría
      const categoryStatsQuery = `
        SELECT c.nombre as categoria, COUNT(p.id) as count
        FROM productos p
        LEFT JOIN categorias_productos c ON p.categoria_id = c.id
        WHERE p.activo = true
        GROUP BY c.id, c.nombre
        ORDER BY count DESC
      `
      const categoryStatsResult = await this.query(categoryStatsQuery)

      // Estadísticas por marca
      const brandStatsQuery = `
        SELECT marca, COUNT(*) as count
        FROM productos
        WHERE activo = true AND marca IS NOT NULL
        GROUP BY marca
        ORDER BY count DESC
        LIMIT 10
      `
      const brandStatsResult = await this.query(brandStatsQuery)

      // Rangos de precios
      const priceRangesQuery = `
        SELECT 
          COUNT(CASE WHEN precio_unitario < 1000 THEN 1 END) as bajo_1000,
          COUNT(CASE WHEN precio_unitario >= 1000 AND precio_unitario < 10000 THEN 1 END) as entre_1000_10000,
          COUNT(CASE WHEN precio_unitario >= 10000 AND precio_unitario < 50000 THEN 1 END) as entre_10000_50000,
          COUNT(CASE WHEN precio_unitario >= 50000 THEN 1 END) as sobre_50000,
          AVG(precio_unitario) as precio_promedio,
          MIN(precio_unitario) as precio_minimo,
          MAX(precio_unitario) as precio_maximo
        FROM productos
        WHERE activo = true
      `
      const priceRangesResult = await this.query(priceRangesQuery)

      return {
        ...baseStats,
        byCategory: categoryStatsResult.rows,
        byBrand: brandStatsResult.rows,
        priceRanges: priceRangesResult.rows[0]
      }
    } catch (error) {
      logger.error('Error al obtener estadísticas de productos:', error)
      throw error
    }
  }

  /**
   * Obtiene productos relacionados o similares
   */
  async getProductosRelacionados(productId, limit = 5) {
    try {
      const producto = await this.findById(productId)
      if (!producto) {
        throw new Error('Producto no encontrado')
      }

      const query = `
        SELECT p.*, 
               CASE 
                 WHEN p.categoria_id = $2 THEN 1
                 WHEN p.marca = $3 THEN 2
                 ELSE 3
               END as relevancia
        FROM productos p
        WHERE p.activo = true 
        AND p.id != $1
        AND (p.categoria_id = $2 OR p.marca = $3)
        ORDER BY relevancia, RANDOM()
        LIMIT $4
      `

      const result = await this.query(query, [
        productId,
        producto.categoria_id,
        producto.marca,
        limit
      ])

      return result.rows
    } catch (error) {
      logger.error('Error al obtener productos relacionados:', error)
      throw error
    }
  }

  /**
   * Valida la disponibilidad de stock para venta
   */
  async validateStockForSale(productId, cantidad, sucursalId) {
    try {
      const query = `
        SELECT sc.cantidad_disponible, p.descripcion
        FROM stock_central sc
        JOIN productos p ON sc.producto_id = p.id
        WHERE sc.producto_id = $1 AND sc.sucursal_id = $2
      `

      const result = await this.query(query, [productId, sucursalId])
      
      if (!result.rows.length) {
        throw new Error('Producto no encontrado en esta sucursal')
      }

      const { cantidad_disponible, descripcion } = result.rows[0]
      
      if (cantidad_disponible < cantidad) {
        throw new Error(`Stock insuficiente para ${descripcion}. Disponible: ${cantidad_disponible}`)
      }

      return true
    } catch (error) {
      logger.error('Error al validar stock para venta:', error)
      throw error
    }
  }
}

module.exports = new Producto()

