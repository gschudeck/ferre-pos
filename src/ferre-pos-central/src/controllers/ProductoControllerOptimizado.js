/**
 * Controlador de Productos Optimizado - Sistema Ferre-POS
 * 
 * Versión optimizada con implementación de concurrencia y multithreading
 * para mejorar el rendimiento en operaciones críticas.
 */

const Producto = require('../models/Producto')
const Stock = require('../models/Stock')
const logger = require('../utils/logger')
const { WorkerPool, ConcurrencyManager, ParallelUtils, TTLCache } = require('../utils/concurrency')
const path = require('path')

class ProductoControllerOptimizado {
  constructor() {
    // Pool de workers para operaciones CPU-intensivas
    this.reporteWorkerPool = new WorkerPool(
      path.join(__dirname, '../workers/reporteWorker.js'),
      2 // 2 workers para reportes
    )
    
    // Gestor de concurrencia para operaciones de BD
    this.concurrencyManager = new ConcurrencyManager(10)
    
    // Cache para consultas frecuentes
    this.cache = new TTLCache(300000) // 5 minutos
    
    // Métricas de rendimiento
    this.metrics = {
      requestsProcessed: 0,
      averageResponseTime: 0,
      cacheHits: 0,
      cacheMisses: 0
    }
  }

  /**
   * Obtiene productos con optimizaciones de concurrencia
   */
  async getProductos(request, reply) {
    const startTime = Date.now()
    
    try {
      const {
        page = 1,
        limit = 20,
        q,
        categoria,
        marca,
        precioMin,
        precioMax,
        conStock = false,
        orderBy = 'descripcion',
        orderDirection = 'ASC'
      } = request.query

      const sucursalId = request.user.sucursal_id
      
      // Crear clave de cache
      const cacheKey = `productos:${JSON.stringify({
        page, limit, q, categoria, marca, precioMin, precioMax, 
        conStock, orderBy, orderDirection, sucursalId
      })}`

      // Intentar obtener del cache
      const cachedResult = await this.cache.get(cacheKey, async () => {
        this.metrics.cacheMisses++
        return await this.buscarProductosOptimizado({
          page: parseInt(page),
          limit: Math.min(parseInt(limit), 100),
          q, categoria, marca, precioMin, precioMax, conStock,
          orderBy, orderDirection, sucursalId
        })
      })

      if (cachedResult !== await this.cache.get(cacheKey, () => null)) {
        this.metrics.cacheHits++
      }

      // Actualizar métricas
      const responseTime = Date.now() - startTime
      this.updateMetrics(responseTime)

      reply.send({
        success: true,
        data: cachedResult.productos,
        pagination: cachedResult.pagination,
        metadata: {
          responseTime,
          cached: this.metrics.cacheHits > this.metrics.cacheMisses,
          totalQueries: cachedResult.totalQueries
        }
      })

    } catch (error) {
      logger.error('Error al obtener productos optimizado:', error)
      reply.code(500).send({
        code: 'INTERNAL_ERROR',
        error: 'Internal Server Error',
        message: 'Error interno del servidor'
      })
    }
  }

  /**
   * Búsqueda optimizada de productos con consultas paralelas
   */
  async buscarProductosOptimizado(options) {
    const {
      page, limit, q, categoria, marca, precioMin, precioMax,
      conStock, orderBy, orderDirection, sucursalId
    } = options

    const offset = (page - 1) * limit

    // Preparar consultas paralelas
    const queries = [
      {
        name: 'productos',
        query: this.buildProductosQuery(options),
        params: this.buildProductosParams(options)
      },
      {
        name: 'total',
        query: this.buildCountQuery(options),
        params: this.buildCountParams(options)
      }
    ]

    // Si se requiere stock, agregar consulta de stock
    if (conStock && sucursalId) {
      queries.push({
        name: 'stock',
        query: `
          SELECT producto_id, cantidad_disponible, cantidad_reservada
          FROM stock 
          WHERE sucursal_id = $1 AND cantidad_disponible > 0
        `,
        params: [sucursalId]
      })
    }

    // Ejecutar consultas en paralelo
    const results = await ParallelUtils.parallelQueries(
      require('../config/database'),
      queries
    )

    // Procesar resultados
    const productosResult = results.find(r => r.name === 'productos')
    const totalResult = results.find(r => r.name === 'total')
    const stockResult = results.find(r => r.name === 'stock')

    let productos = productosResult.result.rows

    // Combinar con información de stock si está disponible
    if (stockResult) {
      const stockMap = new Map(
        stockResult.result.rows.map(s => [s.producto_id, s])
      )

      productos = productos.map(producto => ({
        ...producto,
        stock: stockMap.get(producto.id) || null
      }))
    }

    const total = parseInt(totalResult.result.rows[0].total)

    return {
      productos,
      pagination: {
        page,
        limit,
        total,
        totalPages: Math.ceil(total / limit),
        hasNext: page < Math.ceil(total / limit),
        hasPrev: page > 1
      },
      totalQueries: results.length
    }
  }

  /**
   * Procesamiento masivo de productos con workers
   */
  async procesarProductosMasivo(request, reply) {
    try {
      const { operacion, productos, parametros } = request.body

      // Verificar permisos
      if (!['admin', 'gerente'].includes(request.user.rol)) {
        return reply.code(403).send({
          code: 'INSUFFICIENT_PERMISSIONS',
          error: 'Forbidden',
          message: 'No tiene permisos para procesamiento masivo'
        })
      }

      logger.info('Iniciando procesamiento masivo de productos', {
        operacion,
        cantidad: productos.length,
        usuarioId: request.user.id
      })

      // Procesar en chunks usando workers
      const resultados = await ParallelUtils.processInChunks(
        productos,
        async (chunk, chunkIndex) => {
          return await this.concurrencyManager.execute(async () => {
            switch (operacion) {
              case 'actualizar_precios':
                return await this.actualizarPreciosChunk(chunk, parametros)
              case 'importar_productos':
                return await this.importarProductosChunk(chunk, parametros)
              case 'actualizar_stock':
                return await this.actualizarStockChunk(chunk, parametros)
              default:
                throw new Error(`Operación no soportada: ${operacion}`)
            }
          })
        },
        50, // Chunks de 50 productos
        3   // Máximo 3 chunks concurrentes
      )

      // Consolidar resultados
      const resumenResultados = resultados.reduce((acc, chunk) => {
        acc.procesados += chunk.procesados
        acc.exitosos += chunk.exitosos
        acc.errores += chunk.errores
        acc.detalles.push(...chunk.detalles)
        return acc
      }, {
        procesados: 0,
        exitosos: 0,
        errores: 0,
        detalles: []
      })

      logger.business('Procesamiento masivo completado', {
        operacion,
        ...resumenResultados,
        usuarioId: request.user.id
      })

      reply.send({
        success: true,
        message: 'Procesamiento masivo completado',
        data: resumenResultados
      })

    } catch (error) {
      logger.error('Error en procesamiento masivo:', error)
      reply.code(500).send({
        code: 'INTERNAL_ERROR',
        error: 'Internal Server Error',
        message: 'Error en procesamiento masivo'
      })
    }
  }

  /**
   * Genera reporte de productos usando worker
   */
  async generarReporte(request, reply) {
    try {
      const { tipo, parametros } = request.body

      // Verificar permisos
      if (!['admin', 'gerente'].includes(request.user.rol)) {
        return reply.code(403).send({
          code: 'INSUFFICIENT_PERMISSIONS',
          error: 'Forbidden',
          message: 'No tiene permisos para generar reportes'
        })
      }

      logger.info('Iniciando generación de reporte', {
        tipo,
        parametros,
        usuarioId: request.user.id
      })

      // Ejecutar reporte en worker
      const resultado = await this.reporteWorkerPool.execute({
        tipo,
        parametros: {
          ...parametros,
          startTime: Date.now()
        }
      }, 120000) // Timeout de 2 minutos

      logger.business('Reporte generado exitosamente', {
        tipo,
        registros: resultado.data?.length || 0,
        usuarioId: request.user.id
      })

      reply.send({
        success: true,
        message: 'Reporte generado exitosamente',
        data: resultado
      })

    } catch (error) {
      logger.error('Error al generar reporte:', error)
      reply.code(500).send({
        code: 'INTERNAL_ERROR',
        error: 'Internal Server Error',
        message: 'Error al generar reporte'
      })
    }
  }

  /**
   * Búsqueda inteligente con múltiples criterios en paralelo
   */
  async busquedaInteligente(request, reply) {
    try {
      const { termino, filtros = {}, opciones = {} } = request.query
      const sucursalId = request.user.sucursal_id

      if (!termino || termino.length < 2) {
        return reply.code(400).send({
          code: 'INVALID_SEARCH_TERM',
          error: 'Bad Request',
          message: 'El término de búsqueda debe tener al menos 2 caracteres'
        })
      }

      // Ejecutar múltiples tipos de búsqueda en paralelo
      const busquedas = [
        // Búsqueda por código
        this.concurrencyManager.execute(() => 
          Producto.findByCodigoInterno(termino)
        ),
        // Búsqueda por código de barras
        this.concurrencyManager.execute(() => 
          Producto.findByCodigoBarra(termino)
        ),
        // Búsqueda por descripción
        this.concurrencyManager.execute(() => 
          Producto.searchByDescripcion(termino, filtros)
        ),
        // Búsqueda por marca
        this.concurrencyManager.execute(() => 
          Producto.searchByMarca(termino, filtros)
        )
      ]

      const resultados = await Promise.allSettled(busquedas)

      // Procesar resultados exitosos
      const productosEncontrados = new Map()
      
      resultados.forEach((resultado, index) => {
        if (resultado.status === 'fulfilled' && resultado.value) {
          const productos = Array.isArray(resultado.value) 
            ? resultado.value 
            : [resultado.value]
          
          productos.forEach(producto => {
            if (!productosEncontrados.has(producto.id)) {
              productosEncontrados.set(producto.id, {
                ...producto,
                tipoCoincidencia: this.getTipoCoincidencia(index),
                relevancia: this.calcularRelevancia(producto, termino, index)
              })
            }
          })
        }
      })

      // Ordenar por relevancia
      const productosOrdenados = Array.from(productosEncontrados.values())
        .sort((a, b) => b.relevancia - a.relevancia)

      // Obtener stock en paralelo si es necesario
      if (sucursalId && productosOrdenados.length > 0) {
        const stockPromises = productosOrdenados.map(producto =>
          this.concurrencyManager.execute(() =>
            Stock.getStockProducto(producto.id, sucursalId)
          )
        )

        const stockResults = await Promise.allSettled(stockPromises)
        
        stockResults.forEach((stockResult, index) => {
          if (stockResult.status === 'fulfilled') {
            productosOrdenados[index].stock = stockResult.value
          }
        })
      }

      reply.send({
        success: true,
        data: {
          termino,
          resultados: productosOrdenados,
          estadisticas: {
            totalEncontrados: productosOrdenados.length,
            tiempoRespuesta: Date.now() - request.startTime
          }
        }
      })

    } catch (error) {
      logger.error('Error en búsqueda inteligente:', error)
      reply.code(500).send({
        code: 'INTERNAL_ERROR',
        error: 'Internal Server Error',
        message: 'Error en búsqueda inteligente'
      })
    }
  }

  /**
   * Validación masiva de productos con concurrencia
   */
  async validarProductosMasivo(productos) {
    const validaciones = productos.map(producto => async () => {
      const errores = []

      // Validar código interno único
      if (await Producto.existeCodigoInterno(producto.codigo_interno, producto.id)) {
        errores.push('Código interno ya existe')
      }

      // Validar código de barras único
      if (producto.codigo_barra && 
          await Producto.existeCodigoBarra(producto.codigo_barra, producto.id)) {
        errores.push('Código de barras ya existe')
      }

      // Validar precios
      if (producto.precio_unitario <= 0) {
        errores.push('Precio unitario debe ser mayor a 0')
      }

      if (producto.precio_costo && producto.precio_costo < 0) {
        errores.push('Precio de costo no puede ser negativo')
      }

      return {
        producto: producto.codigo_interno,
        valido: errores.length === 0,
        errores
      }
    })

    return await ParallelUtils.parallelValidations(validaciones)
  }

  // Métodos auxiliares para construcción de queries
  buildProductosQuery(options) {
    const { q, categoria, marca, precioMin, precioMax, conStock, orderBy, orderDirection } = options
    
    let query = `
      SELECT p.*, c.nombre as categoria_nombre
      FROM productos p
      LEFT JOIN categorias c ON p.categoria_id = c.id
      WHERE p.activo = true
    `
    
    if (q) {
      query += ` AND (
        p.descripcion ILIKE $1 OR 
        p.codigo_interno ILIKE $1 OR 
        p.codigo_barra ILIKE $1 OR
        p.marca ILIKE $1
      )`
    }
    
    // Agregar más filtros según sea necesario
    // ... resto de la lógica de construcción de query
    
    query += ` ORDER BY ${orderBy} ${orderDirection} LIMIT $${this.getParamCount(options)} OFFSET $${this.getParamCount(options) + 1}`
    
    return query
  }

  buildProductosParams(options) {
    const params = []
    
    if (options.q) {
      params.push(`%${options.q}%`)
    }
    
    // Agregar más parámetros según filtros
    // ... resto de la lógica de parámetros
    
    params.push(options.limit, (options.page - 1) * options.limit)
    
    return params
  }

  buildCountQuery(options) {
    // Similar a buildProductosQuery pero con COUNT(*)
    return this.buildProductosQuery(options).replace(
      'SELECT p.*, c.nombre as categoria_nombre',
      'SELECT COUNT(*)'
    ).replace(/ORDER BY.*$/, '').replace(/LIMIT.*$/, '')
  }

  buildCountParams(options) {
    // Similar a buildProductosParams pero sin LIMIT y OFFSET
    const params = this.buildProductosParams(options)
    return params.slice(0, -2)
  }

  getParamCount(options) {
    let count = 0
    if (options.q) count++
    // Agregar más contadores según filtros
    return count
  }

  getTipoCoincidencia(index) {
    const tipos = ['codigo_interno', 'codigo_barra', 'descripcion', 'marca']
    return tipos[index] || 'otro'
  }

  calcularRelevancia(producto, termino, tipoIndex) {
    let relevancia = 0
    
    // Puntuación base según tipo de coincidencia
    const puntuacionesTipo = [100, 90, 70, 60] // código interno, código barras, descripción, marca
    relevancia += puntuacionesTipo[tipoIndex] || 50
    
    // Bonificación por coincidencia exacta
    if (producto.descripcion?.toLowerCase() === termino.toLowerCase()) {
      relevancia += 50
    }
    
    // Bonificación por coincidencia al inicio
    if (producto.descripcion?.toLowerCase().startsWith(termino.toLowerCase())) {
      relevancia += 25
    }
    
    return relevancia
  }

  updateMetrics(responseTime) {
    this.metrics.requestsProcessed++
    this.metrics.averageResponseTime = 
      (this.metrics.averageResponseTime * (this.metrics.requestsProcessed - 1) + responseTime) / 
      this.metrics.requestsProcessed
  }

  async actualizarPreciosChunk(productos, parametros) {
    // Implementación de actualización de precios en chunk
    const resultados = { procesados: 0, exitosos: 0, errores: 0, detalles: [] }
    
    for (const producto of productos) {
      try {
        await Producto.updatePrecio(producto.id, parametros.nuevoPrecio)
        resultados.exitosos++
        resultados.detalles.push({
          id: producto.id,
          status: 'success'
        })
      } catch (error) {
        resultados.errores++
        resultados.detalles.push({
          id: producto.id,
          status: 'error',
          error: error.message
        })
      }
      resultados.procesados++
    }
    
    return resultados
  }

  async importarProductosChunk(productos, parametros) {
    // Implementación de importación de productos en chunk
    const resultados = { procesados: 0, exitosos: 0, errores: 0, detalles: [] }
    
    for (const productoData of productos) {
      try {
        const nuevoProducto = await Producto.create(productoData)
        resultados.exitosos++
        resultados.detalles.push({
          codigo: productoData.codigo_interno,
          status: 'success',
          id: nuevoProducto.id
        })
      } catch (error) {
        resultados.errores++
        resultados.detalles.push({
          codigo: productoData.codigo_interno,
          status: 'error',
          error: error.message
        })
      }
      resultados.procesados++
    }
    
    return resultados
  }

  async actualizarStockChunk(productos, parametros) {
    // Implementación de actualización de stock en chunk
    const resultados = { procesados: 0, exitosos: 0, errores: 0, detalles: [] }
    
    for (const item of productos) {
      try {
        await Stock.updateStock(item.producto_id, parametros.sucursal_id, item.cantidad)
        resultados.exitosos++
        resultados.detalles.push({
          producto_id: item.producto_id,
          status: 'success'
        })
      } catch (error) {
        resultados.errores++
        resultados.detalles.push({
          producto_id: item.producto_id,
          status: 'error',
          error: error.message
        })
      }
      resultados.procesados++
    }
    
    return resultados
  }

  /**
   * Obtiene métricas de rendimiento del controlador
   */
  getMetrics() {
    return {
      ...this.metrics,
      workerPoolStats: this.reporteWorkerPool.getStats(),
      cacheStats: this.cache.getStats()
    }
  }

  /**
   * Limpia recursos al cerrar la aplicación
   */
  async cleanup() {
    await this.reporteWorkerPool.close()
    this.cache.clear()
    logger.info('ProductoControllerOptimizado: Recursos limpiados')
  }
}

module.exports = new ProductoControllerOptimizado()

