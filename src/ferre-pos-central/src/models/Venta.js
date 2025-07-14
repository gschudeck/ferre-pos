/**
 * Modelo Venta - Sistema Ferre-POS
 * 
 * Maneja todas las operaciones relacionadas con las ventas,
 * incluyendo creación, anulación, reportes y integración con DTE.
 */

const BaseModel = require('./BaseModel')
const logger = require('../utils/logger')
const database = require('../config/database')

class Venta extends BaseModel {
  constructor() {
    super('ventas', {
      sucursal_id: { type: 'string', required: true },
      terminal_id: { type: 'string', required: true },
      cajero_id: { type: 'string', required: true },
      vendedor_id: { type: 'string' },
      cliente_rut: { type: 'string' },
      cliente_nombre: { type: 'string' },
      tipo_documento: { type: 'string', required: true },
      subtotal: { type: 'number', required: true },
      descuento_total: { type: 'number' },
      impuesto_total: { type: 'number' },
      total: { type: 'number', required: true }
    })
  }

  /**
   * Crea una nueva venta con todos sus detalles y medios de pago
   */
  async createVentaCompleta(ventaData, options = {}) {
    try {
      return await this.transaction(async (client) => {
        const {
          venta,
          detalles,
          mediosPago,
          aplicarFidelizacion = false
        } = ventaData

        // Validar que los totales cuadren
        this.validateTotales(venta, detalles, mediosPago)

        // Validar stock disponible para todos los productos
        await this.validateStockDisponible(detalles, venta.sucursal_id, client)

        // Crear la venta principal
        const ventaCreada = await this.createVentaRecord(venta, client)

        // Crear detalles de venta
        const detallesCreados = await this.createDetallesVenta(
          ventaCreada.id, 
          detalles, 
          client
        )

        // Crear medios de pago
        const mediosPagoCreados = await this.createMediosPago(
          ventaCreada.id, 
          mediosPago, 
          client
        )

        // Actualizar stock
        await this.updateStockVenta(detalles, venta.sucursal_id, client)

        // Procesar fidelización si aplica
        let fidelizacionResult = null
        if (aplicarFidelizacion && venta.cliente_rut) {
          fidelizacionResult = await this.procesarFidelizacion(
            ventaCreada.id,
            venta.cliente_rut,
            venta.total,
            venta.sucursal_id,
            client
          )
        }

        logger.business('Venta creada exitosamente', {
          ventaId: ventaCreada.id,
          numeroVenta: ventaCreada.numero_venta,
          total: ventaCreada.total,
          sucursalId: venta.sucursal_id,
          cajeroId: venta.cajero_id,
          clienteRut: venta.cliente_rut
        })

        return {
          venta: ventaCreada,
          detalles: detallesCreados,
          mediosPago: mediosPagoCreados,
          fidelizacion: fidelizacionResult
        }
      })
    } catch (error) {
      logger.error('Error al crear venta completa:', error)
      throw error
    }
  }

  /**
   * Crea el registro principal de la venta
   */
  async createVentaRecord(ventaData, client) {
    try {
      const query = `
        INSERT INTO ventas (
          sucursal_id, terminal_id, cajero_id, vendedor_id,
          cliente_rut, cliente_nombre, nota_venta_id, tipo_documento,
          subtotal, descuento_total, impuesto_total, total, estado
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, 'finalizada')
        RETURNING *
      `

      const values = [
        ventaData.sucursal_id,
        ventaData.terminal_id,
        ventaData.cajero_id,
        ventaData.vendedor_id,
        ventaData.cliente_rut,
        ventaData.cliente_nombre,
        ventaData.nota_venta_id,
        ventaData.tipo_documento,
        ventaData.subtotal,
        ventaData.descuento_total || 0,
        ventaData.impuesto_total || 0,
        ventaData.total
      ]

      const result = await client.query(query, values)
      return result.rows[0]
    } catch (error) {
      logger.error('Error al crear registro de venta:', error)
      throw error
    }
  }

  /**
   * Crea los detalles de la venta
   */
  async createDetallesVenta(ventaId, detalles, client) {
    try {
      const detallesCreados = []

      for (const detalle of detalles) {
        const query = `
          INSERT INTO detalle_ventas (
            venta_id, producto_id, cantidad, precio_unitario,
            descuento_unitario, precio_final, total_item,
            numero_serie, lote, fecha_vencimiento
          ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
          RETURNING *
        `

        const values = [
          ventaId,
          detalle.producto_id,
          detalle.cantidad,
          detalle.precio_unitario,
          detalle.descuento_unitario || 0,
          detalle.precio_final,
          detalle.total_item,
          detalle.numero_serie,
          detalle.lote,
          detalle.fecha_vencimiento
        ]

        const result = await client.query(query, values)
        detallesCreados.push(result.rows[0])
      }

      return detallesCreados
    } catch (error) {
      logger.error('Error al crear detalles de venta:', error)
      throw error
    }
  }

  /**
   * Crea los medios de pago de la venta
   */
  async createMediosPago(ventaId, mediosPago, client) {
    try {
      const mediosPagoCreados = []

      for (const medio of mediosPago) {
        const query = `
          INSERT INTO medios_pago_venta (
            venta_id, medio_pago, monto, referencia_transaccion,
            codigo_autorizacion, datos_transaccion
          ) VALUES ($1, $2, $3, $4, $5, $6)
          RETURNING *
        `

        const values = [
          ventaId,
          medio.medio_pago,
          medio.monto,
          medio.referencia_transaccion,
          medio.codigo_autorizacion,
          medio.datos_transaccion ? JSON.stringify(medio.datos_transaccion) : null
        ]

        const result = await client.query(query, values)
        mediosPagoCreados.push(result.rows[0])
      }

      return mediosPagoCreados
    } catch (error) {
      logger.error('Error al crear medios de pago:', error)
      throw error
    }
  }

  /**
   * Valida que los totales de la venta sean coherentes
   */
  validateTotales(venta, detalles, mediosPago) {
    // Validar total de detalles
    const totalDetalles = detalles.reduce((sum, detalle) => sum + detalle.total_item, 0)
    const subtotalCalculado = totalDetalles
    const totalCalculado = subtotalCalculado - (venta.descuento_total || 0) + (venta.impuesto_total || 0)

    if (Math.abs(totalCalculado - venta.total) > 0.01) {
      throw new Error('El total de la venta no coincide con la suma de los detalles')
    }

    // Validar total de medios de pago
    const totalMediosPago = mediosPago.reduce((sum, medio) => sum + medio.monto, 0)
    
    if (Math.abs(totalMediosPago - venta.total) > 0.01) {
      throw new Error('El total de medios de pago no coincide con el total de la venta')
    }
  }

  /**
   * Valida que hay stock disponible para todos los productos
   */
  async validateStockDisponible(detalles, sucursalId, client) {
    for (const detalle of detalles) {
      const query = `
        SELECT sc.cantidad_disponible, p.descripcion
        FROM stock_central sc
        JOIN productos p ON sc.producto_id = p.id
        WHERE sc.producto_id = $1 AND sc.sucursal_id = $2
      `

      const result = await client.query(query, [detalle.producto_id, sucursalId])
      
      if (!result.rows.length) {
        throw new Error(`Producto ${detalle.producto_id} no encontrado en stock`)
      }

      const { cantidad_disponible, descripcion } = result.rows[0]
      
      if (cantidad_disponible < detalle.cantidad) {
        throw new Error(`Stock insuficiente para ${descripcion}. Disponible: ${cantidad_disponible}, Solicitado: ${detalle.cantidad}`)
      }
    }
  }

  /**
   * Actualiza el stock después de una venta
   */
  async updateStockVenta(detalles, sucursalId, client) {
    for (const detalle of detalles) {
      // Actualizar stock central
      const updateStockQuery = `
        UPDATE stock_central
        SET cantidad = cantidad - $1,
            fecha_ultima_salida = NOW(),
            fecha_sync = NOW()
        WHERE producto_id = $2 AND sucursal_id = $3
      `

      await client.query(updateStockQuery, [
        detalle.cantidad,
        detalle.producto_id,
        sucursalId
      ])

      // Registrar movimiento de stock
      const movimientoQuery = `
        INSERT INTO movimientos_stock (
          producto_id, sucursal_id, tipo_movimiento, cantidad,
          documento_referencia, observaciones
        ) VALUES ($1, $2, 'venta', $3, $4, 'Venta automática')
      `

      await client.query(movimientoQuery, [
        detalle.producto_id,
        sucursalId,
        -detalle.cantidad,
        `VENTA-${Date.now()}`
      ])
    }
  }

  /**
   * Procesa la fidelización de la venta
   */
  async procesarFidelizacion(ventaId, clienteRut, total, sucursalId, client) {
    try {
      // Verificar si el cliente existe en fidelización
      const clienteQuery = `
        SELECT id, puntos_actuales, nivel_fidelizacion
        FROM fidelizacion_clientes
        WHERE rut = $1 AND activo = true
      `
      const clienteResult = await client.query(clienteQuery, [clienteRut])
      
      if (!clienteResult.rows.length) {
        return null // Cliente no está en programa de fidelización
      }

      const cliente = clienteResult.rows[0]
      
      // Calcular puntos a acumular (1 punto por cada $100)
      const puntosAcumular = Math.floor(total / 100)
      
      if (puntosAcumular > 0) {
        // Actualizar puntos del cliente
        const updateClienteQuery = `
          UPDATE fidelizacion_clientes
          SET puntos_actuales = puntos_actuales + $1,
              puntos_acumulados_total = puntos_acumulados_total + $1,
              fecha_ultima_compra = NOW(),
              fecha_ultima_actividad = NOW()
          WHERE id = $2
        `
        await client.query(updateClienteQuery, [puntosAcumular, cliente.id])

        // Registrar movimiento de fidelización
        const movimientoQuery = `
          INSERT INTO movimientos_fidelizacion (
            cliente_id, sucursal_id, venta_id, tipo, puntos,
            puntos_anteriores, puntos_nuevos, detalle
          ) VALUES ($1, $2, $3, 'acumulacion', $4, $5, $6, $7)
        `
        
        await client.query(movimientoQuery, [
          cliente.id,
          sucursalId,
          ventaId,
          puntosAcumular,
          cliente.puntos_actuales,
          cliente.puntos_actuales + puntosAcumular,
          `Acumulación por venta de $${total}`
        ])

        return {
          puntosAcumulados: puntosAcumular,
          puntosNuevos: cliente.puntos_actuales + puntosAcumular
        }
      }

      return null
    } catch (error) {
      logger.error('Error al procesar fidelización:', error)
      throw error
    }
  }

  /**
   * Anula una venta
   */
  async anularVenta(ventaId, motivo, usuarioId) {
    try {
      return await this.transaction(async (client) => {
        // Verificar que la venta existe y no está anulada
        const ventaQuery = `
          SELECT * FROM ventas WHERE id = $1 AND estado != 'anulada'
        `
        const ventaResult = await client.query(ventaQuery, [ventaId])
        
        if (!ventaResult.rows.length) {
          throw new Error('Venta no encontrada o ya está anulada')
        }

        const venta = ventaResult.rows[0]

        // Obtener detalles de la venta para restaurar stock
        const detallesQuery = `
          SELECT * FROM detalle_ventas WHERE venta_id = $1
        `
        const detallesResult = await client.query(detallesQuery, [ventaId])

        // Restaurar stock
        for (const detalle of detallesResult.rows) {
          const updateStockQuery = `
            UPDATE stock_central
            SET cantidad = cantidad + $1,
                fecha_sync = NOW()
            WHERE producto_id = $2 AND sucursal_id = $3
          `
          
          await client.query(updateStockQuery, [
            detalle.cantidad,
            detalle.producto_id,
            venta.sucursal_id
          ])

          // Registrar movimiento de stock
          const movimientoQuery = `
            INSERT INTO movimientos_stock (
              producto_id, sucursal_id, tipo_movimiento, cantidad,
              documento_referencia, observaciones, usuario_id
            ) VALUES ($1, $2, 'devolucion', $3, $4, $5, $6)
          `

          await client.query(movimientoQuery, [
            detalle.producto_id,
            venta.sucursal_id,
            detalle.cantidad,
            `ANULACION-${venta.numero_venta}`,
            `Anulación de venta: ${motivo}`,
            usuarioId
          ])
        }

        // Revertir fidelización si aplica
        if (venta.cliente_rut) {
          await this.revertirFidelizacion(ventaId, venta.cliente_rut, client)
        }

        // Anular la venta
        const anularQuery = `
          UPDATE ventas
          SET estado = 'anulada',
              fecha_anulacion = NOW(),
              motivo_anulacion = $2,
              usuario_anulacion = $3
          WHERE id = $1
          RETURNING *
        `

        const result = await client.query(anularQuery, [ventaId, motivo, usuarioId])

        logger.business('Venta anulada', {
          ventaId,
          numeroVenta: venta.numero_venta,
          motivo,
          usuarioId
        })

        return result.rows[0]
      })
    } catch (error) {
      logger.error('Error al anular venta:', error)
      throw error
    }
  }

  /**
   * Revierte los puntos de fidelización de una venta anulada
   */
  async revertirFidelizacion(ventaId, clienteRut, client) {
    try {
      // Buscar movimiento de fidelización de la venta
      const movimientoQuery = `
        SELECT mf.*, fc.puntos_actuales
        FROM movimientos_fidelizacion mf
        JOIN fidelizacion_clientes fc ON mf.cliente_id = fc.id
        WHERE mf.venta_id = $1 AND mf.tipo = 'acumulacion'
      `
      const movimientoResult = await client.query(movimientoQuery, [ventaId])

      if (movimientoResult.rows.length > 0) {
        const movimiento = movimientoResult.rows[0]
        
        // Restar puntos del cliente
        const updateClienteQuery = `
          UPDATE fidelizacion_clientes
          SET puntos_actuales = puntos_actuales - $1,
              fecha_ultima_actividad = NOW()
          WHERE id = $2
        `
        await client.query(updateClienteQuery, [movimiento.puntos, movimiento.cliente_id])

        // Registrar movimiento de reversión
        const revertirQuery = `
          INSERT INTO movimientos_fidelizacion (
            cliente_id, sucursal_id, venta_id, tipo, puntos,
            puntos_anteriores, puntos_nuevos, detalle
          ) VALUES ($1, $2, $3, 'ajuste', $4, $5, $6, $7)
        `
        
        await client.query(revertirQuery, [
          movimiento.cliente_id,
          movimiento.sucursal_id,
          ventaId,
          -movimiento.puntos,
          movimiento.puntos_actuales,
          movimiento.puntos_actuales - movimiento.puntos,
          `Reversión por anulación de venta`
        ])
      }
    } catch (error) {
      logger.error('Error al revertir fidelización:', error)
      throw error
    }
  }

  /**
   * Obtiene el detalle completo de una venta
   */
  async getVentaCompleta(ventaId) {
    try {
      // Obtener venta principal
      const ventaQuery = `
        SELECT v.*, s.nombre as sucursal_nombre, t.nombre_terminal,
               u1.nombre as cajero_nombre, u2.nombre as vendedor_nombre
        FROM ventas v
        JOIN sucursales s ON v.sucursal_id = s.id
        LEFT JOIN terminales t ON v.terminal_id = t.id
        LEFT JOIN usuarios u1 ON v.cajero_id = u1.id
        LEFT JOIN usuarios u2 ON v.vendedor_id = u2.id
        WHERE v.id = $1
      `
      const ventaResult = await this.query(ventaQuery, [ventaId])
      
      if (!ventaResult.rows.length) {
        throw new Error('Venta no encontrada')
      }

      const venta = ventaResult.rows[0]

      // Obtener detalles
      const detallesQuery = `
        SELECT dv.*, p.descripcion, p.codigo_interno, p.unidad_medida
        FROM detalle_ventas dv
        JOIN productos p ON dv.producto_id = p.id
        WHERE dv.venta_id = $1
        ORDER BY dv.id
      `
      const detallesResult = await this.query(detallesQuery, [ventaId])

      // Obtener medios de pago
      const mediosPagoQuery = `
        SELECT * FROM medios_pago_venta
        WHERE venta_id = $1
        ORDER BY id
      `
      const mediosPagoResult = await this.query(mediosPagoQuery, [ventaId])

      return {
        venta,
        detalles: detallesResult.rows,
        mediosPago: mediosPagoResult.rows
      }
    } catch (error) {
      logger.error('Error al obtener venta completa:', error)
      throw error
    }
  }

  /**
   * Obtiene ventas por período y sucursal
   */
  async getVentasPorPeriodo(options = {}) {
    try {
      const {
        sucursalId,
        fechaInicio,
        fechaFin,
        cajeroId,
        estado = 'finalizada',
        page = 1,
        limit = 20
      } = options

      let query = `
        SELECT v.*, s.nombre as sucursal_nombre, u.nombre as cajero_nombre
        FROM ventas v
        JOIN sucursales s ON v.sucursal_id = s.id
        LEFT JOIN usuarios u ON v.cajero_id = u.id
        WHERE 1=1
      `
      const params = []

      if (sucursalId) {
        query += ` AND v.sucursal_id = $${params.length + 1}`
        params.push(sucursalId)
      }

      if (fechaInicio) {
        query += ` AND v.fecha >= $${params.length + 1}`
        params.push(fechaInicio)
      }

      if (fechaFin) {
        query += ` AND v.fecha <= $${params.length + 1}`
        params.push(fechaFin)
      }

      if (cajeroId) {
        query += ` AND v.cajero_id = $${params.length + 1}`
        params.push(cajeroId)
      }

      if (estado) {
        query += ` AND v.estado = $${params.length + 1}`
        params.push(estado)
      }

      query += ` ORDER BY v.fecha DESC`

      return await this.queryPaginated(query, params, { page, limit })
    } catch (error) {
      logger.error('Error al obtener ventas por período:', error)
      throw error
    }
  }

  /**
   * Obtiene estadísticas de ventas
   */
  async getVentasStats(options = {}) {
    try {
      const {
        sucursalId,
        fechaInicio = new Date(Date.now() - 30 * 24 * 60 * 60 * 1000),
        fechaFin = new Date()
      } = options

      let whereClause = `WHERE v.estado = 'finalizada' AND v.fecha >= $1 AND v.fecha <= $2`
      const params = [fechaInicio, fechaFin]

      if (sucursalId) {
        whereClause += ` AND v.sucursal_id = $3`
        params.push(sucursalId)
      }

      const statsQuery = `
        SELECT 
          COUNT(*) as total_ventas,
          SUM(v.total) as monto_total,
          AVG(v.total) as promedio_venta,
          MIN(v.total) as venta_minima,
          MAX(v.total) as venta_maxima,
          COUNT(DISTINCT v.cajero_id) as cajeros_activos,
          COUNT(DISTINCT v.cliente_rut) as clientes_unicos,
          COUNT(DISTINCT DATE(v.fecha)) as dias_con_ventas
        FROM ventas v
        ${whereClause}
      `

      const result = await this.query(statsQuery, params)
      return result.rows[0]
    } catch (error) {
      logger.error('Error al obtener estadísticas de ventas:', error)
      throw error
    }
  }
}

module.exports = new Venta()

