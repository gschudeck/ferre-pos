/**
 * Controlador de Productos - Sistema Ferre-POS
 * 
 * Maneja todas las operaciones del catálogo de productos,
 * incluyendo CRUD, búsquedas avanzadas y gestión de códigos de barras.
 */

const Producto = require('../models/Producto')
const Stock = require('../models/Stock')
const logger = require('../utils/logger')

class ProductoController {
  /**
   * Obtiene todos los productos con paginación y filtros
   */
  async getProductos(request, reply) {
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

      const filters = {
        q,
        categoria,
        marca,
        precioMin: precioMin ? parseFloat(precioMin) : undefined,
        precioMax: precioMax ? parseFloat(precioMax) : undefined,
        conStock: conStock === 'true',
        sucursalId: conStock === 'true' ? sucursalId : undefined
      }

      const options = {
        page: parseInt(page),
        limit: Math.min(parseInt(limit), 100),
        orderBy,
        orderDirection: orderDirection.toUpperCase()
      }

      const productos = await Producto.search(filters, options)

      reply.send({
        success: true,
        data: productos,
        pagination: {
          page: options.page,
          limit: options.limit,
          total: productos.length
        }
      })
    } catch (error) {
      logger.error('Error al obtener productos:', error)
      reply.code(500).send({
        code: 'INTERNAL_ERROR',
        error: 'Internal Server Error',
        message: 'Error interno del servidor'
      })
    }
  }

  /**
   * Obtiene un producto por ID
   */
  async getProducto(request, reply) {
    try {
      const { id } = request.params
      const sucursalId = request.user.sucursal_id

      const producto = await Producto.findById(id)
      
      if (!producto) {
        return reply.code(404).send({
          code: 'PRODUCT_NOT_FOUND',
          error: 'Not Found',
          message: 'Producto no encontrado'
        })
      }

      // Obtener stock si el usuario tiene sucursal asignada
      let stockInfo = null
      if (sucursalId) {
        stockInfo = await Stock.getStockProducto(id, sucursalId)
      }

      // Obtener códigos de barras adicionales
      const codigosBarras = await Producto.getCodigosBarras(id)

      reply.send({
        success: true,
        data: {
          ...producto,
          stock: stockInfo,
          codigosBarras
        }
      })
    } catch (error) {
      logger.error('Error al obtener producto:', error)
      reply.code(500).send({
        code: 'INTERNAL_ERROR',
        error: 'Internal Server Error',
        message: 'Error interno del servidor'
      })
    }
  }

  /**
   * Busca un producto por código (interno o de barras)
   */
  async getProductoPorCodigo(request, reply) {
    try {
      const { codigo } = request.params
      const sucursalId = request.user.sucursal_id

      // Buscar por código interno primero
      let producto = await Producto.findByCodigoInterno(codigo)
      
      // Si no se encuentra, buscar por código de barras
      if (!producto) {
        producto = await Producto.findByCodigoBarra(codigo)
      }

      if (!producto) {
        return reply.code(404).send({
          code: 'PRODUCT_NOT_FOUND',
          error: 'Not Found',
          message: 'Producto no encontrado con el código proporcionado'
        })
      }

      // Obtener stock si el usuario tiene sucursal asignada
      let stockInfo = null
      if (sucursalId) {
        stockInfo = await Stock.getStockProducto(producto.id, sucursalId)
      }

      reply.send({
        success: true,
        data: {
          ...producto,
          stock: stockInfo
        }
      })
    } catch (error) {
      logger.error('Error al buscar producto por código:', error)
      reply.code(500).send({
        code: 'INTERNAL_ERROR',
        error: 'Internal Server Error',
        message: 'Error interno del servidor'
      })
    }
  }

  /**
   * Crea un nuevo producto
   */
  async createProducto(request, reply) {
    try {
      // Verificar permisos
      if (!['admin', 'gerente'].includes(request.user.rol)) {
        return reply.code(403).send({
          code: 'INSUFFICIENT_PERMISSIONS',
          error: 'Forbidden',
          message: 'No tiene permisos para crear productos'
        })
      }

      const productData = {
        ...request.body,
        usuario_creacion: request.user.id
      }

      const nuevoProducto = await Producto.create(productData)

      logger.business('Producto creado', {
        productId: nuevoProducto.id,
        codigoInterno: nuevoProducto.codigo_interno,
        descripcion: nuevoProducto.descripcion,
        usuarioId: request.user.id
      })

      reply.code(201).send({
        success: true,
        message: 'Producto creado exitosamente',
        data: nuevoProducto
      })
    } catch (error) {
      logger.error('Error al crear producto:', error)
      
      let statusCode = 500
      let errorCode = 'INTERNAL_ERROR'
      
      if (error.message.includes('Ya existe')) {
        statusCode = 409
        errorCode = 'DUPLICATE_PRODUCT'
      } else if (error.message.includes('debe ser mayor')) {
        statusCode = 400
        errorCode = 'INVALID_PRICE'
      }

      reply.code(statusCode).send({
        code: errorCode,
        error: statusCode === 500 ? 'Internal Server Error' : 'Bad Request',
        message: statusCode === 500 ? 'Error interno del servidor' : error.message
      })
    }
  }

  /**
   * Actualiza un producto existente
   */
  async updateProducto(request, reply) {
    try {
      const { id } = request.params

      // Verificar permisos
      if (!['admin', 'gerente'].includes(request.user.rol)) {
        return reply.code(403).send({
          code: 'INSUFFICIENT_PERMISSIONS',
          error: 'Forbidden',
          message: 'No tiene permisos para actualizar productos'
        })
      }

      // Verificar que el producto existe
      const productoExistente = await Producto.findById(id)
      if (!productoExistente) {
        return reply.code(404).send({
          code: 'PRODUCT_NOT_FOUND',
          error: 'Not Found',
          message: 'Producto no encontrado'
        })
      }

      const updateData = {
        ...request.body,
        usuario_modificacion: request.user.id
      }

      const productoActualizado = await Producto.updateById(id, updateData)

      logger.business('Producto actualizado', {
        productId: id,
        codigoInterno: productoActualizado.codigo_interno,
        usuarioId: request.user.id
      })

      reply.send({
        success: true,
        message: 'Producto actualizado exitosamente',
        data: productoActualizado
      })
    } catch (error) {
      logger.error('Error al actualizar producto:', error)
      reply.code(500).send({
        code: 'INTERNAL_ERROR',
        error: 'Internal Server Error',
        message: 'Error interno del servidor'
      })
    }
  }

  /**
   * Elimina un producto (eliminación lógica)
   */
  async deleteProducto(request, reply) {
    try {
      const { id } = request.params

      // Verificar permisos de administrador
      if (request.user.rol !== 'admin') {
        return reply.code(403).send({
          code: 'INSUFFICIENT_PERMISSIONS',
          error: 'Forbidden',
          message: 'Solo los administradores pueden eliminar productos'
        })
      }

      // Verificar que el producto existe
      const producto = await Producto.findById(id)
      if (!producto) {
        return reply.code(404).send({
          code: 'PRODUCT_NOT_FOUND',
          error: 'Not Found',
          message: 'Producto no encontrado'
        })
      }

      await Producto.deleteById(id)

      logger.business('Producto eliminado', {
        productId: id,
        codigoInterno: producto.codigo_interno,
        usuarioId: request.user.id
      })

      reply.send({
        success: true,
        message: 'Producto eliminado exitosamente'
      })
    } catch (error) {
      logger.error('Error al eliminar producto:', error)
      reply.code(500).send({
        code: 'INTERNAL_ERROR',
        error: 'Internal Server Error',
        message: 'Error interno del servidor'
      })
    }
  }

  /**
   * Actualiza el precio de un producto
   */
  async updatePrecio(request, reply) {
    try {
      const { id } = request.params
      const { precio } = request.body

      // Verificar permisos
      if (!['admin', 'gerente'].includes(request.user.rol)) {
        return reply.code(403).send({
          code: 'INSUFFICIENT_PERMISSIONS',
          error: 'Forbidden',
          message: 'No tiene permisos para actualizar precios'
        })
      }

      if (!precio || precio <= 0) {
        return reply.code(400).send({
          code: 'INVALID_PRICE',
          error: 'Bad Request',
          message: 'El precio debe ser mayor a 0'
        })
      }

      await Producto.updatePrecio(id, precio, request.user.id)

      reply.send({
        success: true,
        message: 'Precio actualizado exitosamente'
      })
    } catch (error) {
      logger.error('Error al actualizar precio:', error)
      reply.code(500).send({
        code: 'INTERNAL_ERROR',
        error: 'Internal Server Error',
        message: 'Error interno del servidor'
      })
    }
  }

  /**
   * Agrega un código de barras adicional a un producto
   */
  async addCodigoBarra(request, reply) {
    try {
      const { id } = request.params
      const { codigoBarra, descripcion } = request.body

      // Verificar permisos
      if (!['admin', 'gerente'].includes(request.user.rol)) {
        return reply.code(403).send({
          code: 'INSUFFICIENT_PERMISSIONS',
          error: 'Forbidden',
          message: 'No tiene permisos para agregar códigos de barras'
        })
      }

      if (!codigoBarra) {
        return reply.code(400).send({
          code: 'MISSING_BARCODE',
          error: 'Bad Request',
          message: 'Código de barras es requerido'
        })
      }

      const codigoAgregado = await Producto.addCodigoBarraAdicional(id, codigoBarra, descripcion)

      reply.code(201).send({
        success: true,
        message: 'Código de barras agregado exitosamente',
        data: codigoAgregado
      })
    } catch (error) {
      logger.error('Error al agregar código de barras:', error)
      
      let statusCode = 500
      let errorCode = 'INTERNAL_ERROR'
      
      if (error.message.includes('ya está en uso')) {
        statusCode = 409
        errorCode = 'DUPLICATE_BARCODE'
      }

      reply.code(statusCode).send({
        code: errorCode,
        error: statusCode === 500 ? 'Internal Server Error' : 'Conflict',
        message: statusCode === 500 ? 'Error interno del servidor' : error.message
      })
    }
  }

  /**
   * Obtiene productos con stock bajo
   */
  async getProductosStockBajo(request, reply) {
    try {
      const sucursalId = request.user.rol === 'admin' ? request.query.sucursalId : request.user.sucursal_id

      const productos = await Producto.getProductosStockBajo(sucursalId)

      reply.send({
        success: true,
        data: productos
      })
    } catch (error) {
      logger.error('Error al obtener productos con stock bajo:', error)
      reply.code(500).send({
        code: 'INTERNAL_ERROR',
        error: 'Internal Server Error',
        message: 'Error interno del servidor'
      })
    }
  }

  /**
   * Obtiene productos más vendidos
   */
  async getProductosMasVendidos(request, reply) {
    try {
      const {
        fechaInicio,
        fechaFin,
        limit = 20
      } = request.query

      const sucursalId = request.user.rol === 'admin' ? request.query.sucursalId : request.user.sucursal_id

      const options = {
        sucursalId,
        fechaInicio: fechaInicio ? new Date(fechaInicio) : undefined,
        fechaFin: fechaFin ? new Date(fechaFin) : undefined,
        limit: parseInt(limit)
      }

      const productos = await Producto.getProductosMasVendidos(options)

      reply.send({
        success: true,
        data: productos
      })
    } catch (error) {
      logger.error('Error al obtener productos más vendidos:', error)
      reply.code(500).send({
        code: 'INTERNAL_ERROR',
        error: 'Internal Server Error',
        message: 'Error interno del servidor'
      })
    }
  }

  /**
   * Obtiene productos relacionados
   */
  async getProductosRelacionados(request, reply) {
    try {
      const { id } = request.params
      const { limit = 5 } = request.query

      const productos = await Producto.getProductosRelacionados(id, parseInt(limit))

      reply.send({
        success: true,
        data: productos
      })
    } catch (error) {
      logger.error('Error al obtener productos relacionados:', error)
      reply.code(500).send({
        code: 'INTERNAL_ERROR',
        error: 'Internal Server Error',
        message: 'Error interno del servidor'
      })
    }
  }

  /**
   * Obtiene estadísticas del catálogo de productos
   */
  async getProductStats(request, reply) {
    try {
      // Verificar permisos para estadísticas completas
      if (!['admin', 'gerente'].includes(request.user.rol)) {
        return reply.code(403).send({
          code: 'INSUFFICIENT_PERMISSIONS',
          error: 'Forbidden',
          message: 'No tiene permisos para ver estadísticas completas'
        })
      }

      const stats = await Producto.getProductStats()

      reply.send({
        success: true,
        data: stats
      })
    } catch (error) {
      logger.error('Error al obtener estadísticas de productos:', error)
      reply.code(500).send({
        code: 'INTERNAL_ERROR',
        error: 'Internal Server Error',
        message: 'Error interno del servidor'
      })
    }
  }

  /**
   * Valida stock para venta
   */
  async validateStock(request, reply) {
    try {
      const { productos } = request.body
      const sucursalId = request.user.sucursal_id

      if (!productos || !Array.isArray(productos)) {
        return reply.code(400).send({
          code: 'INVALID_PRODUCTS',
          error: 'Bad Request',
          message: 'Lista de productos es requerida'
        })
      }

      const validaciones = []

      for (const item of productos) {
        try {
          await Producto.validateStockForSale(item.producto_id, item.cantidad, sucursalId)
          validaciones.push({
            producto_id: item.producto_id,
            cantidad: item.cantidad,
            valido: true
          })
        } catch (error) {
          validaciones.push({
            producto_id: item.producto_id,
            cantidad: item.cantidad,
            valido: false,
            error: error.message
          })
        }
      }

      const todasValidas = validaciones.every(v => v.valido)

      reply.send({
        success: true,
        data: {
          valido: todasValidas,
          validaciones
        }
      })
    } catch (error) {
      logger.error('Error al validar stock:', error)
      reply.code(500).send({
        code: 'INTERNAL_ERROR',
        error: 'Internal Server Error',
        message: 'Error interno del servidor'
      })
    }
  }
}

module.exports = new ProductoController()

