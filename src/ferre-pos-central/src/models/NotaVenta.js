/**
 * Modelo NotaVenta - Sistema Ferre-POS
 * 
 * Maneja todas las operaciones relacionadas con las notas de venta,
 * incluyendo cotizaciones, reservas y conversión a ventas.
 */

const BaseModel = require('./BaseModel')
const logger = require('../utils/logger')
const database = require('../config/database')

class NotaVenta extends BaseModel {
  constructor() {
    super('notas_venta', {
      sucursal_id: { type: 'string', required: true },
      vendedor_id: { type: 'string', required: true },
      cliente_rut: { type: 'string' },
      cliente_nombre: { type: 'string' },
      cliente_telefono: { type: 'string' },
      cliente_email: { type: 'string' },
      tipo_nota: { type: 'string', required: true },
      subtotal: { type: 'number', required: true },
      descuento_total: { type: 'number' },
      impuesto_total: { type: 'number' },
      total: { type: 'number', required: true },
      estado: { type: 'string', required: true },
      observaciones: { type: 'string' }
    })
  }

  /**
   * Crea una nueva nota de venta con todos sus detalles
   */
  async createNotaCompleta(notaData, options = {}) {
    try {
      return await this.transaction(async (client) => {
        const {
          nota,
          detalles
        } = notaData

        // Validar que los totales cuadren
        this.validateTotales(nota, detalles)

        // Validar stock disponible para todos los productos
        await this.validateStockDisponible(detalles, nota.sucursal_id, client)

        // Crear la nota de venta principal
        const notaCreada = await this.createNotaRecord(nota, client)

        // Crear detalles de nota de venta
        const detallesCreados = await this.createDetallesNota(
          notaCreada.id, 
          detalles, 
          client
        )

        // Reservar stock si es necesario
        if (nota.tipo_nota === 'reserva') {
          await this.reservarStockNota(detalles, nota.sucursal_id, notaCreada.id, client)
        }

        logger.business('Nota de venta creada exitosamente', {
          notaId: notaCreada.id,
          numeroNota: notaCreada.numero_nota,
          tipoNota: notaCreada.tipo_nota,
          total: notaCreada.total,
          sucursalId: nota.sucursal_id,
          vendedorId: nota.vendedor_id,
          clienteRut: nota.cliente_rut
        })

        return {
          nota: notaCreada,
          detalles: detallesCreados
        }
      })
    } catch (error) {
      logger.error('Error al crear nota de venta completa:', error)
      throw error
    }
  }

  /**
   * Crea el registro principal de la nota de venta
   */
  async createNotaRecord(notaData, client) {
    try {
      const query = `
        INSERT INTO notas_venta (
          sucursal_id, vendedor_id, cliente_rut, cliente_nombre,
          cliente_telefono, cliente_email, tipo_nota, subtotal,
          descuento_total, impuesto_total, total, estado,
          observaciones, fecha_vencimiento
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
        RETURNING *
      `

      // Calcular fecha de vencimiento según tipo de nota
      let fechaVencimiento = null
      if (notaData.tipo_nota === 'cotizacion') {
        fechaVencimiento = new Date(Date.now() + 30 * 24 * 60 * 60 * 1000) // 30 días
      } else if (notaData.tipo_nota === 'reserva') {
        fechaVencimiento = new Date(Date.now() + 7 * 24 * 60 * 60 * 1000) // 7 días
      }

      const values = [
        notaData.sucursal_id,
        notaData.vendedor_id,
        notaData.cliente_rut,
        notaData.cliente_nombre,
        notaData.cliente_telefono,
        notaData.cliente_email,
        notaData.tipo_nota,
        notaData.subtotal,
        notaData.descuento_total || 0,
        notaData.impuesto_total || 0,
        notaData.total,
        'activa',
        notaData.observaciones,
        fechaVencimiento
      ]

      const result = await client.query(query, values)
      return result.rows[0]
    } catch (error) {
      logger.error('Error al crear registro de nota de venta:', error)
      throw error
    }
  }

  /**
   * Crea los detalles de la nota de venta
   */
  async createDetallesNota(notaId, detalles, client) {
    try {
      const detallesCreados = []

      for (const detalle of detalles) {
        const query = `
          INSERT INTO detalle_notas_venta (
            nota_venta_id, producto_id, cantidad, precio_unitario,
            descuento_unitario, precio_final, total_item,
            observaciones
          ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
          RETURNING *
        `

        const values = [
          notaId,
          detalle.producto_id,
          detalle.cantidad,
          detalle.precio_unitario,
          detalle.descuento_unitario || 0,
          detalle.precio_final,
          detalle.total_item,
          detalle.observaciones
        ]

        const result = await client.query(query, values)
        detallesCreados.push(result.rows[0])
      }

      return detallesCreados
    } catch (error) {
      logger.error('Error al crear detalles de nota de venta:', error)
      throw error
    }
  }

  /**
   * Valida que los totales de la nota sean coherentes
   */
  validateTotales(nota, detalles) {
    // Validar total de detalles
    const totalDetalles = detalles.reduce((sum, detalle) => sum + detalle.total_item, 0)
    const subtotalCalculado = totalDetalles
    const totalCalculado = subtotalCalculado - (nota.descuento_total || 0) + (nota.impuesto_total || 0)

    if (Math.abs(totalCalculado - nota.total) > 0.01) {
      throw new Error('El total de la nota no coincide con la suma de los detalles')
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
   * Reserva stock para una nota de venta tipo reserva
   */
  async reservarStockNota(detalles, sucursalId, notaId, client) {
    for (const detalle of detalles) {
      // Actualizar cantidad reservada en stock
      const updateStockQuery = `
        UPDATE stock_central
        SET cantidad_reservada = cantidad_reservada + $1,
            fecha_sync = NOW()
        WHERE producto_id = $2 AND sucursal_id = $3
      `

      await client.query(updateStockQuery, [
        detalle.cantidad,
        detalle.producto_id,
        sucursalId
      ])

      // Registrar reserva de stock
      const reservaQuery = `
        INSERT INTO reservas_stock (
          producto_id, sucursal_id, nota_venta_id, cantidad,
          fecha_vencimiento, estado
        ) VALUES ($1, $2, $3, $4, $5, 'activa')
      `

      const fechaVencimiento = new Date(Date.now() + 7 * 24 * 60 * 60 * 1000) // 7 días

      await client.query(reservaQuery, [
        detalle.producto_id,
        sucursalId,
        notaId,
        detalle.cantidad,
        fechaVencimiento
      ])
    }
  }

  /**
   * Convierte una nota de venta en venta real
   */
  async convertirAVenta(notaId, datosVenta, usuarioId) {
    try {
      return await this.transaction(async (client) => {
        // Verificar que la nota existe y está activa
        const notaQuery = `
          SELECT nv.*, s.nombre as sucursal_nombre
          FROM notas_venta nv
          JOIN sucursales s ON nv.sucursal_id = s.id
          WHERE nv.id = $1 AND nv.estado = 'activa'
        `
        const notaResult = await client.query(notaQuery, [notaId])
        
        if (!notaResult.rows.length) {
          throw new Error('Nota de venta no encontrada o no está activa')
        }

        const nota = notaResult.rows[0]

        // Obtener detalles de la nota
        const detallesQuery = `
          SELECT dnv.*, p.descripcion, p.codigo_interno
          FROM detalle_notas_venta dnv
          JOIN productos p ON dnv.producto_id = p.id
          WHERE dnv.nota_venta_id = $1
        `
        const detallesResult = await client.query(detallesQuery, [notaId])

        // Validar stock nuevamente (por si cambió desde la creación de la nota)
        await this.validateStockDisponible(detallesResult.rows, nota.sucursal_id, client)

        // Crear la venta
        const ventaQuery = `
          INSERT INTO ventas (
            sucursal_id, terminal_id, cajero_id, vendedor_id,
            cliente_rut, cliente_nombre, nota_venta_id, tipo_documento,
            subtotal, descuento_total, impuesto_total, total, estado
          ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, 'finalizada')
          RETURNING *
        `

        const ventaValues = [
          nota.sucursal_id,
          datosVenta.terminal_id,
          datosVenta.cajero_id || usuarioId,
          nota.vendedor_id,
          nota.cliente_rut,
          nota.cliente_nombre,
          notaId,
          datosVenta.tipo_documento,
          nota.subtotal,
          nota.descuento_total,
          nota.impuesto_total,
          nota.total
        ]

        const ventaResult = await client.query(ventaQuery, ventaValues)
        const venta = ventaResult.rows[0]

        // Crear detalles de venta
        for (const detalle of detallesResult.rows) {
          const detalleVentaQuery = `
            INSERT INTO detalle_ventas (
              venta_id, producto_id, cantidad, precio_unitario,
              descuento_unitario, precio_final, total_item
            ) VALUES ($1, $2, $3, $4, $5, $6, $7)
          `

          await client.query(detalleVentaQuery, [
            venta.id,
            detalle.producto_id,
            detalle.cantidad,
            detalle.precio_unitario,
            detalle.descuento_unitario,
            detalle.precio_final,
            detalle.total_item
          ])

          // Actualizar stock
          const updateStockQuery = `
            UPDATE stock_central
            SET cantidad = cantidad - $1,
                cantidad_reservada = CASE 
                  WHEN $4 = 'reserva' THEN cantidad_reservada - $1
                  ELSE cantidad_reservada
                END,
                fecha_ultima_salida = NOW(),
                fecha_sync = NOW()
            WHERE producto_id = $2 AND sucursal_id = $3
          `

          await client.query(updateStockQuery, [
            detalle.cantidad,
            detalle.producto_id,
            nota.sucursal_id,
            nota.tipo_nota
          ])

          // Registrar movimiento de stock
          const movimientoQuery = `
            INSERT INTO movimientos_stock (
              producto_id, sucursal_id, tipo_movimiento, cantidad,
              documento_referencia, observaciones, usuario_id
            ) VALUES ($1, $2, 'venta', $3, $4, $5, $6)
          `

          await client.query(movimientoQuery, [
            detalle.producto_id,
            nota.sucursal_id,
            -detalle.cantidad,
            `VENTA-${venta.numero_venta}`,
            `Venta desde nota ${nota.numero_nota}`,
            usuarioId
          ])
        }

        // Crear medios de pago
        if (datosVenta.mediosPago) {
          for (const medio of datosVenta.mediosPago) {
            const medioPagoQuery = `
              INSERT INTO medios_pago_venta (
                venta_id, medio_pago, monto, referencia_transaccion,
                codigo_autorizacion, datos_transaccion
              ) VALUES ($1, $2, $3, $4, $5, $6)
            `

            await client.query(medioPagoQuery, [
              venta.id,
              medio.medio_pago,
              medio.monto,
              medio.referencia_transaccion,
              medio.codigo_autorizacion,
              medio.datos_transaccion ? JSON.stringify(medio.datos_transaccion) : null
            ])
          }
        }

        // Liberar reservas de stock si las había
        if (nota.tipo_nota === 'reserva') {
          await client.query(
            'UPDATE reservas_stock SET estado = $1 WHERE nota_venta_id = $2',
            ['liberada', notaId]
          )
        }

        // Marcar nota como convertida
        await client.query(
          'UPDATE notas_venta SET estado = $1, fecha_conversion = NOW(), venta_id = $2 WHERE id = $3',
          ['convertida', venta.id, notaId]
        )

        logger.business('Nota de venta convertida a venta', {
          notaId,
          ventaId: venta.id,
          numeroNota: nota.numero_nota,
          numeroVenta: venta.numero_venta,
          total: venta.total,
          usuarioId
        })

        return {
          venta,
          notaOriginal: nota
        }
      })
    } catch (error) {
      logger.error('Error al convertir nota de venta:', error)
      throw error
    }
  }

  /**
   * Anula una nota de venta
   */
  async anularNota(notaId, motivo, usuarioId) {
    try {
      return await this.transaction(async (client) => {
        // Verificar que la nota existe y está activa
        const notaQuery = `
          SELECT * FROM notas_venta WHERE id = $1 AND estado = 'activa'
        `
        const notaResult = await client.query(notaQuery, [notaId])
        
        if (!notaResult.rows.length) {
          throw new Error('Nota de venta no encontrada o no está activa')
        }

        const nota = notaResult.rows[0]

        // Liberar reservas de stock si las había
        if (nota.tipo_nota === 'reserva') {
          // Obtener detalles para liberar stock
          const detallesQuery = `
            SELECT * FROM detalle_notas_venta WHERE nota_venta_id = $1
          `
          const detallesResult = await client.query(detallesQuery, [notaId])

          for (const detalle of detallesResult.rows) {
            // Liberar stock reservado
            const updateStockQuery = `
              UPDATE stock_central
              SET cantidad_reservada = GREATEST(cantidad_reservada - $1, 0),
                  fecha_sync = NOW()
              WHERE producto_id = $2 AND sucursal_id = $3
            `

            await client.query(updateStockQuery, [
              detalle.cantidad,
              detalle.producto_id,
              nota.sucursal_id
            ])
          }

          // Marcar reservas como liberadas
          await client.query(
            'UPDATE reservas_stock SET estado = $1 WHERE nota_venta_id = $2',
            ['liberada', notaId]
          )
        }

        // Anular la nota
        const anularQuery = `
          UPDATE notas_venta
          SET estado = 'anulada',
              fecha_anulacion = NOW(),
              motivo_anulacion = $2,
              usuario_anulacion = $3
          WHERE id = $1
          RETURNING *
        `

        const result = await client.query(anularQuery, [notaId, motivo, usuarioId])

        logger.business('Nota de venta anulada', {
          notaId,
          numeroNota: nota.numero_nota,
          tipoNota: nota.tipo_nota,
          motivo,
          usuarioId
        })

        return result.rows[0]
      })
    } catch (error) {
      logger.error('Error al anular nota de venta:', error)
      throw error
    }
  }

  /**
   * Obtiene el detalle completo de una nota de venta
   */
  async getNotaCompleta(notaId) {
    try {
      // Obtener nota principal
      const notaQuery = `
        SELECT nv.*, s.nombre as sucursal_nombre,
               u.nombre as vendedor_nombre,
               v.numero_venta as venta_numero
        FROM notas_venta nv
        JOIN sucursales s ON nv.sucursal_id = s.id
        LEFT JOIN usuarios u ON nv.vendedor_id = u.id
        LEFT JOIN ventas v ON nv.venta_id = v.id
        WHERE nv.id = $1
      `
      const notaResult = await this.query(notaQuery, [notaId])
      
      if (!notaResult.rows.length) {
        throw new Error('Nota de venta no encontrada')
      }

      const nota = notaResult.rows[0]

      // Obtener detalles
      const detallesQuery = `
        SELECT dnv.*, p.descripcion, p.codigo_interno, p.unidad_medida
        FROM detalle_notas_venta dnv
        JOIN productos p ON dnv.producto_id = p.id
        WHERE dnv.nota_venta_id = $1
        ORDER BY dnv.id
      `
      const detallesResult = await this.query(detallesQuery, [notaId])

      // Obtener reservas de stock si las hay
      let reservas = []
      if (nota.tipo_nota === 'reserva') {
        const reservasQuery = `
          SELECT rs.*, p.descripcion as producto_descripcion
          FROM reservas_stock rs
          JOIN productos p ON rs.producto_id = p.id
          WHERE rs.nota_venta_id = $1
          ORDER BY rs.id
        `
        const reservasResult = await this.query(reservasQuery, [notaId])
        reservas = reservasResult.rows
      }

      return {
        nota,
        detalles: detallesResult.rows,
        reservas
      }
    } catch (error) {
      logger.error('Error al obtener nota completa:', error)
      throw error
    }
  }

  /**
   * Obtiene notas de venta por período y filtros
   */
  async getNotasPorPeriodo(options = {}) {
    try {
      const {
        sucursalId,
        vendedorId,
        fechaInicio,
        fechaFin,
        tipoNota,
        estado = 'activa',
        clienteRut,
        page = 1,
        limit = 20
      } = options

      let query = `
        SELECT nv.*, s.nombre as sucursal_nombre, u.nombre as vendedor_nombre
        FROM notas_venta nv
        JOIN sucursales s ON nv.sucursal_id = s.id
        LEFT JOIN usuarios u ON nv.vendedor_id = u.id
        WHERE 1=1
      `
      const params = []

      if (sucursalId) {
        query += ` AND nv.sucursal_id = $${params.length + 1}`
        params.push(sucursalId)
      }

      if (vendedorId) {
        query += ` AND nv.vendedor_id = $${params.length + 1}`
        params.push(vendedorId)
      }

      if (fechaInicio) {
        query += ` AND nv.fecha >= $${params.length + 1}`
        params.push(fechaInicio)
      }

      if (fechaFin) {
        query += ` AND nv.fecha <= $${params.length + 1}`
        params.push(fechaFin)
      }

      if (tipoNota) {
        query += ` AND nv.tipo_nota = $${params.length + 1}`
        params.push(tipoNota)
      }

      if (estado) {
        query += ` AND nv.estado = $${params.length + 1}`
        params.push(estado)
      }

      if (clienteRut) {
        query += ` AND nv.cliente_rut = $${params.length + 1}`
        params.push(clienteRut)
      }

      query += ` ORDER BY nv.fecha DESC`

      return await this.queryPaginated(query, params, { page, limit })
    } catch (error) {
      logger.error('Error al obtener notas por período:', error)
      throw error
    }
  }

  /**
   * Obtiene estadísticas de notas de venta
   */
  async getNotasStats(options = {}) {
    try {
      const {
        sucursalId,
        fechaInicio = new Date(Date.now() - 30 * 24 * 60 * 60 * 1000),
        fechaFin = new Date()
      } = options

      let whereClause = `WHERE nv.fecha >= $1 AND nv.fecha <= $2`
      const params = [fechaInicio, fechaFin]

      if (sucursalId) {
        whereClause += ` AND nv.sucursal_id = $3`
        params.push(sucursalId)
      }

      const statsQuery = `
        SELECT 
          COUNT(*) as total_notas,
          COUNT(CASE WHEN nv.estado = 'activa' THEN 1 END) as notas_activas,
          COUNT(CASE WHEN nv.estado = 'convertida' THEN 1 END) as notas_convertidas,
          COUNT(CASE WHEN nv.estado = 'anulada' THEN 1 END) as notas_anuladas,
          COUNT(CASE WHEN nv.tipo_nota = 'cotizacion' THEN 1 END) as cotizaciones,
          COUNT(CASE WHEN nv.tipo_nota = 'reserva' THEN 1 END) as reservas,
          SUM(nv.total) as monto_total,
          AVG(nv.total) as promedio_nota,
          COUNT(DISTINCT nv.vendedor_id) as vendedores_activos,
          COUNT(DISTINCT nv.cliente_rut) as clientes_unicos
        FROM notas_venta nv
        ${whereClause}
      `

      const result = await this.query(statsQuery, params)
      return result.rows[0]
    } catch (error) {
      logger.error('Error al obtener estadísticas de notas:', error)
      throw error
    }
  }

  /**
   * Obtiene notas próximas a vencer
   */
  async getNotasProximasVencer(diasAnticipacion = 3) {
    try {
      const fechaLimite = new Date(Date.now() + diasAnticipacion * 24 * 60 * 60 * 1000)

      const query = `
        SELECT nv.*, s.nombre as sucursal_nombre, u.nombre as vendedor_nombre,
               DATE_PART('day', nv.fecha_vencimiento - NOW()) as dias_restantes
        FROM notas_venta nv
        JOIN sucursales s ON nv.sucursal_id = s.id
        LEFT JOIN usuarios u ON nv.vendedor_id = u.id
        WHERE nv.estado = 'activa'
        AND nv.fecha_vencimiento IS NOT NULL
        AND nv.fecha_vencimiento <= $1
        ORDER BY nv.fecha_vencimiento ASC
      `

      const result = await this.query(query, [fechaLimite])
      return result.rows
    } catch (error) {
      logger.error('Error al obtener notas próximas a vencer:', error)
      throw error
    }
  }

  /**
   * Busca notas de venta con filtros avanzados
   */
  async searchNotas(filters = {}, options = {}) {
    try {
      const {
        q, // Término de búsqueda general
        numeroNota,
        clienteRut,
        clienteNombre,
        vendedorId,
        sucursalId,
        tipoNota,
        estado,
        fechaInicio,
        fechaFin,
        montoMin,
        montoMax
      } = filters

      let query = `
        SELECT nv.*, s.nombre as sucursal_nombre, u.nombre as vendedor_nombre,
               CASE 
                 WHEN nv.numero_nota::text ILIKE $1 THEN 1
                 WHEN nv.cliente_nombre ILIKE $1 THEN 2
                 WHEN nv.cliente_rut ILIKE $1 THEN 3
                 ELSE 4
               END as relevancia
        FROM notas_venta nv
        JOIN sucursales s ON nv.sucursal_id = s.id
        LEFT JOIN usuarios u ON nv.vendedor_id = u.id
        WHERE 1=1
      `

      const params = []
      let paramIndex = 1

      // Búsqueda general
      if (q) {
        query += ` AND (
          nv.numero_nota::text ILIKE $${paramIndex} OR
          nv.cliente_nombre ILIKE $${paramIndex} OR
          nv.cliente_rut ILIKE $${paramIndex} OR
          nv.observaciones ILIKE $${paramIndex}
        )`
        params.push(`%${q}%`)
        paramIndex++
      } else {
        params.push('')
        paramIndex++
      }

      // Filtros específicos
      if (numeroNota) {
        query += ` AND nv.numero_nota = $${paramIndex}`
        params.push(numeroNota)
        paramIndex++
      }

      if (clienteRut) {
        query += ` AND nv.cliente_rut = $${paramIndex}`
        params.push(clienteRut)
        paramIndex++
      }

      if (clienteNombre) {
        query += ` AND nv.cliente_nombre ILIKE $${paramIndex}`
        params.push(`%${clienteNombre}%`)
        paramIndex++
      }

      if (vendedorId) {
        query += ` AND nv.vendedor_id = $${paramIndex}`
        params.push(vendedorId)
        paramIndex++
      }

      if (sucursalId) {
        query += ` AND nv.sucursal_id = $${paramIndex}`
        params.push(sucursalId)
        paramIndex++
      }

      if (tipoNota) {
        query += ` AND nv.tipo_nota = $${paramIndex}`
        params.push(tipoNota)
        paramIndex++
      }

      if (estado) {
        query += ` AND nv.estado = $${paramIndex}`
        params.push(estado)
        paramIndex++
      }

      if (fechaInicio) {
        query += ` AND nv.fecha >= $${paramIndex}`
        params.push(fechaInicio)
        paramIndex++
      }

      if (fechaFin) {
        query += ` AND nv.fecha <= $${paramIndex}`
        params.push(fechaFin)
        paramIndex++
      }

      if (montoMin !== undefined) {
        query += ` AND nv.total >= $${paramIndex}`
        params.push(montoMin)
        paramIndex++
      }

      if (montoMax !== undefined) {
        query += ` AND nv.total <= $${paramIndex}`
        params.push(montoMax)
        paramIndex++
      }

      // Ordenamiento
      const { orderBy = 'relevancia, nv.fecha', orderDirection = 'DESC' } = options
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
      logger.error('Error en búsqueda de notas:', error)
      throw error
    }
  }
}

module.exports = new NotaVenta()

