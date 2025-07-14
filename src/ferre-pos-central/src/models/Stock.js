/**
 * Modelo Stock - Sistema Ferre-POS
 * 
 * Maneja todas las operaciones relacionadas con el control de inventario,
 * incluyendo movimientos, transferencias y alertas de stock.
 */

const BaseModel = require('./BaseModel')
const logger = require('../utils/logger')

class Stock extends BaseModel {
  constructor() {
    super('stock_central', {
      producto_id: { type: 'string', required: true },
      sucursal_id: { type: 'string', required: true },
      cantidad: { type: 'number', required: true },
      cantidad_reservada: { type: 'number' },
      costo_promedio: { type: 'number' }
    })
    this.primaryKey = ['producto_id', 'sucursal_id']
    this.timestamps = false
  }

  /**
   * Obtiene el stock de un producto en una sucursal específica
   */
  async getStockProducto(productoId, sucursalId) {
    try {
      const query = `
        SELECT sc.*, p.descripcion, p.codigo_interno, p.stock_minimo, p.stock_maximo,
               s.nombre as sucursal_nombre
        FROM stock_central sc
        JOIN productos p ON sc.producto_id = p.id
        JOIN sucursales s ON sc.sucursal_id = s.id
        WHERE sc.producto_id = $1 AND sc.sucursal_id = $2
      `
      
      const result = await this.query(query, [productoId, sucursalId])
      return result.rows[0] || null
    } catch (error) {
      logger.error('Error al obtener stock de producto:', error)
      throw error
    }
  }

  /**
   * Obtiene el stock consolidado de un producto en todas las sucursales
   */
  async getStockConsolidado(productoId) {
    try {
      const query = `
        SELECT sc.*, s.nombre as sucursal_nombre, s.codigo as sucursal_codigo,
               p.descripcion, p.codigo_interno, p.stock_minimo
        FROM stock_central sc
        JOIN productos p ON sc.producto_id = p.id
        JOIN sucursales s ON sc.sucursal_id = s.id
        WHERE sc.producto_id = $1 AND s.habilitada = true
        ORDER BY s.nombre
      `
      
      const result = await this.query(query, [productoId])
      
      // Calcular totales
      const stocks = result.rows
      const totalCantidad = stocks.reduce((sum, stock) => sum + stock.cantidad, 0)
      const totalDisponible = stocks.reduce((sum, stock) => sum + stock.cantidad_disponible, 0)
      const totalReservado = stocks.reduce((sum, stock) => sum + stock.cantidad_reservada, 0)

      return {
        stocks,
        totales: {
          cantidad: totalCantidad,
          disponible: totalDisponible,
          reservado: totalReservado
        }
      }
    } catch (error) {
      logger.error('Error al obtener stock consolidado:', error)
      throw error
    }
  }

  /**
   * Registra un movimiento de stock
   */
  async registrarMovimiento(movimientoData) {
    try {
      return await this.transaction(async (client) => {
        const {
          producto_id,
          sucursal_id,
          tipo_movimiento,
          cantidad,
          costo_unitario,
          documento_referencia,
          usuario_id,
          observaciones
        } = movimientoData

        // Obtener stock actual
        const stockQuery = `
          SELECT cantidad FROM stock_central
          WHERE producto_id = $1 AND sucursal_id = $2
        `
        const stockResult = await client.query(stockQuery, [producto_id, sucursal_id])
        const cantidadAnterior = stockResult.rows[0]?.cantidad || 0

        // Calcular nueva cantidad
        let nuevaCantidad = cantidadAnterior
        if (['entrada', 'transferencia_entrada', 'devolucion', 'ajuste'].includes(tipo_movimiento)) {
          nuevaCantidad += cantidad
        } else if (['salida', 'transferencia_salida', 'venta'].includes(tipo_movimiento)) {
          nuevaCantidad -= cantidad
        } else if (tipo_movimiento === 'ajuste') {
          nuevaCantidad = cantidad // Para ajustes, la cantidad es el valor final
        }

        // Actualizar o crear registro de stock
        const upsertStockQuery = `
          INSERT INTO stock_central (producto_id, sucursal_id, cantidad, costo_promedio, fecha_sync)
          VALUES ($1, $2, $3, $4, NOW())
          ON CONFLICT (producto_id, sucursal_id)
          DO UPDATE SET 
            cantidad = $3,
            costo_promedio = CASE 
              WHEN $4 IS NOT NULL THEN $4 
              ELSE stock_central.costo_promedio 
            END,
            fecha_sync = NOW()
        `
        
        await client.query(upsertStockQuery, [
          producto_id, 
          sucursal_id, 
          nuevaCantidad, 
          costo_unitario
        ])

        // Registrar movimiento
        const movimientoQuery = `
          INSERT INTO movimientos_stock (
            producto_id, sucursal_id, tipo_movimiento, cantidad,
            cantidad_anterior, cantidad_nueva, costo_unitario,
            documento_referencia, usuario_id, observaciones
          ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
          RETURNING *
        `

        const movimientoResult = await client.query(movimientoQuery, [
          producto_id,
          sucursal_id,
          tipo_movimiento,
          tipo_movimiento.includes('salida') || tipo_movimiento === 'venta' ? -Math.abs(cantidad) : Math.abs(cantidad),
          cantidadAnterior,
          nuevaCantidad,
          costo_unitario,
          documento_referencia,
          usuario_id,
          observaciones
        ])

        logger.business('Movimiento de stock registrado', {
          movimientoId: movimientoResult.rows[0].id,
          productoId: producto_id,
          sucursalId: sucursal_id,
          tipoMovimiento: tipo_movimiento,
          cantidad: cantidad,
          cantidadAnterior,
          cantidadNueva: nuevaCantidad
        })

        return movimientoResult.rows[0]
      })
    } catch (error) {
      logger.error('Error al registrar movimiento de stock:', error)
      throw error
    }
  }

  /**
   * Realiza una transferencia de stock entre sucursales
   */
  async transferirStock(transferenciaData) {
    try {
      return await this.transaction(async (client) => {
        const {
          producto_id,
          sucursal_origen_id,
          sucursal_destino_id,
          cantidad,
          usuario_id,
          observaciones
        } = transferenciaData

        // Validar stock disponible en origen
        const stockOrigenQuery = `
          SELECT cantidad_disponible FROM stock_central
          WHERE producto_id = $1 AND sucursal_id = $2
        `
        const stockOrigenResult = await client.query(stockOrigenQuery, [producto_id, sucursal_origen_id])
        
        if (!stockOrigenResult.rows.length || stockOrigenResult.rows[0].cantidad_disponible < cantidad) {
          throw new Error('Stock insuficiente en sucursal de origen')
        }

        const documentoReferencia = `TRANSFER-${Date.now()}`

        // Registrar salida en sucursal origen
        await this.registrarMovimiento({
          producto_id,
          sucursal_id: sucursal_origen_id,
          tipo_movimiento: 'transferencia_salida',
          cantidad,
          documento_referencia: documentoReferencia,
          usuario_id,
          observaciones: `Transferencia a sucursal destino. ${observaciones || ''}`
        })

        // Registrar entrada en sucursal destino
        await this.registrarMovimiento({
          producto_id,
          sucursal_id: sucursal_destino_id,
          tipo_movimiento: 'transferencia_entrada',
          cantidad,
          documento_referencia: documentoReferencia,
          usuario_id,
          observaciones: `Transferencia desde sucursal origen. ${observaciones || ''}`
        })

        logger.business('Transferencia de stock completada', {
          productoId: producto_id,
          sucursalOrigenId: sucursal_origen_id,
          sucursalDestinoId: sucursal_destino_id,
          cantidad,
          documentoReferencia
        })

        return { documentoReferencia, cantidad }
      })
    } catch (error) {
      logger.error('Error al transferir stock:', error)
      throw error
    }
  }

  /**
   * Reserva stock para una venta
   */
  async reservarStock(reservaData) {
    try {
      return await this.transaction(async (client) => {
        const { producto_id, sucursal_id, cantidad, referencia } = reservaData

        // Verificar stock disponible
        const stockQuery = `
          SELECT cantidad_disponible FROM stock_central
          WHERE producto_id = $1 AND sucursal_id = $2
        `
        const stockResult = await client.query(stockQuery, [producto_id, sucursal_id])
        
        if (!stockResult.rows.length || stockResult.rows[0].cantidad_disponible < cantidad) {
          throw new Error('Stock insuficiente para reservar')
        }

        // Actualizar cantidad reservada
        const updateQuery = `
          UPDATE stock_central
          SET cantidad_reservada = cantidad_reservada + $1,
              fecha_sync = NOW()
          WHERE producto_id = $2 AND sucursal_id = $3
        `
        
        await client.query(updateQuery, [cantidad, producto_id, sucursal_id])

        logger.business('Stock reservado', {
          productoId: producto_id,
          sucursalId: sucursal_id,
          cantidad,
          referencia
        })

        return true
      })
    } catch (error) {
      logger.error('Error al reservar stock:', error)
      throw error
    }
  }

  /**
   * Libera stock reservado
   */
  async liberarStock(reservaData) {
    try {
      const { producto_id, sucursal_id, cantidad } = reservaData

      const updateQuery = `
        UPDATE stock_central
        SET cantidad_reservada = GREATEST(cantidad_reservada - $1, 0),
            fecha_sync = NOW()
        WHERE producto_id = $2 AND sucursal_id = $3
      `
      
      await this.query(updateQuery, [cantidad, producto_id, sucursal_id])

      logger.business('Stock liberado', {
        productoId: producto_id,
        sucursalId: sucursal_id,
        cantidad
      })

      return true
    } catch (error) {
      logger.error('Error al liberar stock:', error)
      throw error
    }
  }

  /**
   * Obtiene productos con stock bajo
   */
  async getProductosStockBajo(sucursalId = null) {
    try {
      let query = `
        SELECT sc.*, p.descripcion, p.codigo_interno, p.stock_minimo,
               s.nombre as sucursal_nombre,
               CASE 
                 WHEN sc.cantidad_disponible = 0 THEN 'AGOTADO'
                 WHEN sc.cantidad_disponible <= p.stock_minimo THEN 'CRITICO'
                 WHEN sc.cantidad_disponible <= p.stock_minimo * 2 THEN 'BAJO'
                 ELSE 'NORMAL'
               END as estado_stock
        FROM stock_central sc
        JOIN productos p ON sc.producto_id = p.id
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
   * Obtiene el historial de movimientos de un producto
   */
  async getHistorialMovimientos(productoId, sucursalId = null, options = {}) {
    try {
      const {
        fechaInicio,
        fechaFin,
        tipoMovimiento,
        limit = 50
      } = options

      let query = `
        SELECT ms.*, p.descripcion, p.codigo_interno,
               s.nombre as sucursal_nombre,
               u.nombre as usuario_nombre
        FROM movimientos_stock ms
        JOIN productos p ON ms.producto_id = p.id
        JOIN sucursales s ON ms.sucursal_id = s.id
        LEFT JOIN usuarios u ON ms.usuario_id = u.id
        WHERE ms.producto_id = $1
      `
      
      const params = [productoId]

      if (sucursalId) {
        query += ` AND ms.sucursal_id = $${params.length + 1}`
        params.push(sucursalId)
      }

      if (fechaInicio) {
        query += ` AND ms.fecha >= $${params.length + 1}`
        params.push(fechaInicio)
      }

      if (fechaFin) {
        query += ` AND ms.fecha <= $${params.length + 1}`
        params.push(fechaFin)
      }

      if (tipoMovimiento) {
        query += ` AND ms.tipo_movimiento = $${params.length + 1}`
        params.push(tipoMovimiento)
      }

      query += ` ORDER BY ms.fecha DESC LIMIT $${params.length + 1}`
      params.push(limit)

      const result = await this.query(query, params)
      return result.rows
    } catch (error) {
      logger.error('Error al obtener historial de movimientos:', error)
      throw error
    }
  }

  /**
   * Realiza un ajuste de inventario
   */
  async ajustarInventario(ajusteData) {
    try {
      const {
        producto_id,
        sucursal_id,
        cantidad_fisica,
        usuario_id,
        observaciones
      } = ajusteData

      // Obtener cantidad actual en sistema
      const stockActual = await this.getStockProducto(producto_id, sucursal_id)
      const cantidadSistema = stockActual?.cantidad || 0

      if (cantidad_fisica === cantidadSistema) {
        throw new Error('No hay diferencia entre cantidad física y sistema')
      }

      const diferencia = cantidad_fisica - cantidadSistema
      const tipoAjuste = diferencia > 0 ? 'entrada' : 'salida'

      await this.registrarMovimiento({
        producto_id,
        sucursal_id,
        tipo_movimiento: 'ajuste',
        cantidad: Math.abs(diferencia),
        documento_referencia: `AJUSTE-${Date.now()}`,
        usuario_id,
        observaciones: `Ajuste de inventario. Cantidad física: ${cantidad_fisica}, Sistema: ${cantidadSistema}. ${observaciones || ''}`
      })

      logger.business('Ajuste de inventario realizado', {
        productoId: producto_id,
        sucursalId: sucursal_id,
        cantidadSistema,
        cantidadFisica: cantidad_fisica,
        diferencia,
        usuarioId: usuario_id
      })

      return {
        cantidadAnterior: cantidadSistema,
        cantidadNueva: cantidad_fisica,
        diferencia
      }
    } catch (error) {
      logger.error('Error al ajustar inventario:', error)
      throw error
    }
  }

  /**
   * Obtiene estadísticas de stock por sucursal
   */
  async getStockStats(sucursalId = null) {
    try {
      let query = `
        SELECT 
          COUNT(*) as total_productos,
          SUM(sc.cantidad) as cantidad_total,
          SUM(sc.cantidad_disponible) as disponible_total,
          SUM(sc.cantidad_reservada) as reservado_total,
          COUNT(CASE WHEN sc.cantidad_disponible = 0 THEN 1 END) as productos_agotados,
          COUNT(CASE WHEN sc.cantidad_disponible <= p.stock_minimo THEN 1 END) as productos_stock_bajo,
          AVG(sc.costo_promedio) as costo_promedio_general
        FROM stock_central sc
        JOIN productos p ON sc.producto_id = p.id
        JOIN sucursales s ON sc.sucursal_id = s.id
        WHERE p.activo = true AND s.habilitada = true
      `
      
      const params = []
      
      if (sucursalId) {
        query += ` AND sc.sucursal_id = $1`
        params.push(sucursalId)
      }

      const result = await this.query(query, params)
      return result.rows[0]
    } catch (error) {
      logger.error('Error al obtener estadísticas de stock:', error)
      throw error
    }
  }
}

module.exports = new Stock()

