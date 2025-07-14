/**
 * Controlador de Notas de Venta - Sistema Ferre-POS
 * 
 * Maneja todas las operaciones de las notas de venta,
 * incluyendo cotizaciones, reservas y conversión a ventas.
 */

const NotaVenta = require('../models/NotaVenta')
const logger = require('../utils/logger')

class NotaVentaController {
  /**
   * Crea una nueva nota de venta
   */
  async createNota(request, reply) {
    try {
      const notaData = {
        nota: {
          ...request.body.nota,
          vendedor_id: request.user.id,
          sucursal_id: request.user.sucursal_id || request.body.nota.sucursal_id
        },
        detalles: request.body.detalles
      }

      // Verificar permisos
      if (!['admin', 'gerente', 'vendedor'].includes(request.user.rol)) {
        return reply.code(403).send({
          code: 'INSUFFICIENT_PERMISSIONS',
          error: 'Forbidden',
          message: 'No tiene permisos para crear notas de venta'
        })
      }

      const resultado = await NotaVenta.createNotaCompleta(notaData)

      logger.business('Nota de venta creada', {
        notaId: resultado.nota.id,
        numeroNota: resultado.nota.numero_nota,
        tipoNota: resultado.nota.tipo_nota,
        total: resultado.nota.total,
        usuarioId: request.user.id
      })

      reply.code(201).send({
        success: true,
        message: 'Nota de venta creada exitosamente',
        data: resultado
      })
    } catch (error) {
      logger.error('Error al crear nota de venta:', error)
      
      let statusCode = 500
      let errorCode = 'INTERNAL_ERROR'
      
      if (error.message.includes('Stock insuficiente')) {
        statusCode = 400
        errorCode = 'INSUFFICIENT_STOCK'
      } else if (error.message.includes('total no coincide')) {
        statusCode = 400
        errorCode = 'INVALID_TOTALS'
      } else if (error.message.includes('no encontrado')) {
        statusCode = 404
        errorCode = 'PRODUCT_NOT_FOUND'
      }

      reply.code(statusCode).send({
        code: errorCode,
        error: statusCode === 500 ? 'Internal Server Error' : 'Bad Request',
        message: statusCode === 500 ? 'Error interno del servidor' : error.message
      })
    }
  }

  /**
   * Obtiene una nota de venta por ID
   */
  async getNota(request, reply) {
    try {
      const { id } = request.params

      const notaCompleta = await NotaVenta.getNotaCompleta(id)

      // Verificar permisos de acceso
      if (request.user.rol !== 'admin' && 
          request.user.sucursal_id !== notaCompleta.nota.sucursal_id) {
        return reply.code(403).send({
          code: 'INSUFFICIENT_PERMISSIONS',
          error: 'Forbidden',
          message: 'No tiene permisos para ver esta nota de venta'
        })
      }

      reply.send({
        success: true,
        data: notaCompleta
      })
    } catch (error) {
      logger.error('Error al obtener nota de venta:', error)
      
      if (error.message.includes('no encontrada')) {
        return reply.code(404).send({
          code: 'NOTA_NOT_FOUND',
          error: 'Not Found',
          message: 'Nota de venta no encontrada'
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
   * Obtiene lista de notas de venta con filtros
   */
  async getNotas(request, reply) {
    try {
      const {
        page = 1,
        limit = 20,
        sucursalId,
        vendedorId,
        fechaInicio,
        fechaFin,
        tipoNota,
        estado,
        clienteRut
      } = request.query

      const options = {
        sucursalId: request.user.rol === 'admin' ? sucursalId : request.user.sucursal_id,
        vendedorId: request.user.rol === 'vendedor' ? request.user.id : vendedorId,
        fechaInicio: fechaInicio ? new Date(fechaInicio) : undefined,
        fechaFin: fechaFin ? new Date(fechaFin) : undefined,
        tipoNota,
        estado,
        clienteRut,
        page: parseInt(page),
        limit: Math.min(parseInt(limit), 100)
      }

      const notas = await NotaVenta.getNotasPorPeriodo(options)

      reply.send({
        success: true,
        data: notas.data,
        pagination: notas.pagination
      })
    } catch (error) {
      logger.error('Error al obtener notas de venta:', error)
      reply.code(500).send({
        code: 'INTERNAL_ERROR',
        error: 'Internal Server Error',
        message: 'Error interno del servidor'
      })
    }
  }

  /**
   * Busca notas de venta
   */
  async searchNotas(request, reply) {
    try {
      const {
        q,
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
        montoMax,
        page = 1,
        limit = 20
      } = request.query

      const filters = {
        q,
        numeroNota,
        clienteRut,
        clienteNombre,
        vendedorId: request.user.rol === 'vendedor' ? request.user.id : vendedorId,
        sucursalId: request.user.rol === 'admin' ? sucursalId : request.user.sucursal_id,
        tipoNota,
        estado,
        fechaInicio: fechaInicio ? new Date(fechaInicio) : undefined,
        fechaFin: fechaFin ? new Date(fechaFin) : undefined,
        montoMin: montoMin ? parseFloat(montoMin) : undefined,
        montoMax: montoMax ? parseFloat(montoMax) : undefined
      }

      const options = {
        limit: Math.min(parseInt(limit), 100),
        offset: (parseInt(page) - 1) * Math.min(parseInt(limit), 100)
      }

      const notas = await NotaVenta.searchNotas(filters, options)

      reply.send({
        success: true,
        data: notas,
        pagination: {
          page: parseInt(page),
          limit: parseInt(limit),
          total: notas.length
        }
      })
    } catch (error) {
      logger.error('Error en búsqueda de notas:', error)
      reply.code(500).send({
        code: 'INTERNAL_ERROR',
        error: 'Internal Server Error',
        message: 'Error interno del servidor'
      })
    }
  }

  /**
   * Convierte una nota de venta en venta real
   */
  async convertirAVenta(request, reply) {
    try {
      const { id } = request.params
      const datosVenta = request.body

      // Verificar permisos
      if (!['admin', 'gerente', 'cajero'].includes(request.user.rol)) {
        return reply.code(403).send({
          code: 'INSUFFICIENT_PERMISSIONS',
          error: 'Forbidden',
          message: 'No tiene permisos para convertir notas a ventas'
        })
      }

      const resultado = await NotaVenta.convertirAVenta(id, datosVenta, request.user.id)

      logger.business('Nota convertida a venta', {
        notaId: id,
        ventaId: resultado.venta.id,
        numeroVenta: resultado.venta.numero_venta,
        total: resultado.venta.total,
        usuarioId: request.user.id
      })

      reply.send({
        success: true,
        message: 'Nota de venta convertida exitosamente',
        data: resultado
      })
    } catch (error) {
      logger.error('Error al convertir nota de venta:', error)
      
      let statusCode = 500
      let errorCode = 'INTERNAL_ERROR'
      
      if (error.message.includes('no encontrada') || error.message.includes('no está activa')) {
        statusCode = 404
        errorCode = 'NOTA_NOT_FOUND'
      } else if (error.message.includes('Stock insuficiente')) {
        statusCode = 400
        errorCode = 'INSUFFICIENT_STOCK'
      }

      reply.code(statusCode).send({
        code: errorCode,
        error: statusCode === 500 ? 'Internal Server Error' : 'Bad Request',
        message: statusCode === 500 ? 'Error interno del servidor' : error.message
      })
    }
  }

  /**
   * Anula una nota de venta
   */
  async anularNota(request, reply) {
    try {
      const { id } = request.params
      const { motivo } = request.body

      // Verificar permisos
      if (!['admin', 'gerente'].includes(request.user.rol)) {
        return reply.code(403).send({
          code: 'INSUFFICIENT_PERMISSIONS',
          error: 'Forbidden',
          message: 'No tiene permisos para anular notas de venta'
        })
      }

      if (!motivo || motivo.trim().length < 10) {
        return reply.code(400).send({
          code: 'INVALID_REASON',
          error: 'Bad Request',
          message: 'El motivo de anulación debe tener al menos 10 caracteres'
        })
      }

      const notaAnulada = await NotaVenta.anularNota(id, motivo, request.user.id)

      reply.send({
        success: true,
        message: 'Nota de venta anulada exitosamente',
        data: notaAnulada
      })
    } catch (error) {
      logger.error('Error al anular nota de venta:', error)
      
      if (error.message.includes('no encontrada') || error.message.includes('no está activa')) {
        return reply.code(404).send({
          code: 'NOTA_NOT_FOUND',
          error: 'Not Found',
          message: 'Nota de venta no encontrada o no está activa'
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
   * Actualiza una nota de venta
   */
  async updateNota(request, reply) {
    try {
      const { id } = request.params
      const updateData = request.body

      // Verificar permisos
      if (!['admin', 'gerente', 'vendedor'].includes(request.user.rol)) {
        return reply.code(403).send({
          code: 'INSUFFICIENT_PERMISSIONS',
          error: 'Forbidden',
          message: 'No tiene permisos para actualizar notas de venta'
        })
      }

      // Verificar que la nota existe y está activa
      const notaExistente = await NotaVenta.findById(id)
      if (!notaExistente) {
        return reply.code(404).send({
          code: 'NOTA_NOT_FOUND',
          error: 'Not Found',
          message: 'Nota de venta no encontrada'
        })
      }

      if (notaExistente.estado !== 'activa') {
        return reply.code(400).send({
          code: 'NOTA_NOT_EDITABLE',
          error: 'Bad Request',
          message: 'Solo se pueden editar notas activas'
        })
      }

      // Verificar permisos de sucursal
      if (request.user.rol !== 'admin' && 
          request.user.sucursal_id !== notaExistente.sucursal_id) {
        return reply.code(403).send({
          code: 'INSUFFICIENT_PERMISSIONS',
          error: 'Forbidden',
          message: 'No tiene permisos para editar esta nota de venta'
        })
      }

      const notaActualizada = await NotaVenta.updateById(id, updateData)

      logger.business('Nota de venta actualizada', {
        notaId: id,
        numeroNota: notaActualizada.numero_nota,
        usuarioId: request.user.id
      })

      reply.send({
        success: true,
        message: 'Nota de venta actualizada exitosamente',
        data: notaActualizada
      })
    } catch (error) {
      logger.error('Error al actualizar nota de venta:', error)
      reply.code(500).send({
        code: 'INTERNAL_ERROR',
        error: 'Internal Server Error',
        message: 'Error interno del servidor'
      })
    }
  }

  /**
   * Obtiene estadísticas de notas de venta
   */
  async getNotasStats(request, reply) {
    try {
      // Verificar permisos
      if (!['admin', 'gerente'].includes(request.user.rol)) {
        return reply.code(403).send({
          code: 'INSUFFICIENT_PERMISSIONS',
          error: 'Forbidden',
          message: 'No tiene permisos para ver estadísticas'
        })
      }

      const {
        fechaInicio,
        fechaFin,
        sucursalId
      } = request.query

      const options = {
        sucursalId: request.user.rol === 'admin' ? sucursalId : request.user.sucursal_id,
        fechaInicio: fechaInicio ? new Date(fechaInicio) : undefined,
        fechaFin: fechaFin ? new Date(fechaFin) : undefined
      }

      const stats = await NotaVenta.getNotasStats(options)

      reply.send({
        success: true,
        data: stats
      })
    } catch (error) {
      logger.error('Error al obtener estadísticas de notas:', error)
      reply.code(500).send({
        code: 'INTERNAL_ERROR',
        error: 'Internal Server Error',
        message: 'Error interno del servidor'
      })
    }
  }

  /**
   * Obtiene notas próximas a vencer
   */
  async getNotasProximasVencer(request, reply) {
    try {
      const { diasAnticipacion = 3 } = request.query

      const notas = await NotaVenta.getNotasProximasVencer(parseInt(diasAnticipacion))

      // Filtrar por sucursal si no es admin
      let notasFiltradas = notas
      if (request.user.rol !== 'admin') {
        notasFiltradas = notas.filter(nota => nota.sucursal_id === request.user.sucursal_id)
      }

      reply.send({
        success: true,
        data: notasFiltradas
      })
    } catch (error) {
      logger.error('Error al obtener notas próximas a vencer:', error)
      reply.code(500).send({
        code: 'INTERNAL_ERROR',
        error: 'Internal Server Error',
        message: 'Error interno del servidor'
      })
    }
  }

  /**
   * Duplica una nota de venta existente
   */
  async duplicarNota(request, reply) {
    try {
      const { id } = request.params

      // Verificar permisos
      if (!['admin', 'gerente', 'vendedor'].includes(request.user.rol)) {
        return reply.code(403).send({
          code: 'INSUFFICIENT_PERMISSIONS',
          error: 'Forbidden',
          message: 'No tiene permisos para duplicar notas de venta'
        })
      }

      // Obtener nota original
      const notaOriginal = await NotaVenta.getNotaCompleta(id)

      // Verificar permisos de sucursal
      if (request.user.rol !== 'admin' && 
          request.user.sucursal_id !== notaOriginal.nota.sucursal_id) {
        return reply.code(403).send({
          code: 'INSUFFICIENT_PERMISSIONS',
          error: 'Forbidden',
          message: 'No tiene permisos para duplicar esta nota de venta'
        })
      }

      // Preparar datos para nueva nota
      const nuevaNotaData = {
        nota: {
          sucursal_id: notaOriginal.nota.sucursal_id,
          vendedor_id: request.user.id,
          cliente_rut: notaOriginal.nota.cliente_rut,
          cliente_nombre: notaOriginal.nota.cliente_nombre,
          cliente_telefono: notaOriginal.nota.cliente_telefono,
          cliente_email: notaOriginal.nota.cliente_email,
          tipo_nota: notaOriginal.nota.tipo_nota,
          subtotal: notaOriginal.nota.subtotal,
          descuento_total: notaOriginal.nota.descuento_total,
          impuesto_total: notaOriginal.nota.impuesto_total,
          total: notaOriginal.nota.total,
          observaciones: `Duplicada de nota ${notaOriginal.nota.numero_nota}`
        },
        detalles: notaOriginal.detalles.map(detalle => ({
          producto_id: detalle.producto_id,
          cantidad: detalle.cantidad,
          precio_unitario: detalle.precio_unitario,
          descuento_unitario: detalle.descuento_unitario,
          precio_final: detalle.precio_final,
          total_item: detalle.total_item,
          observaciones: detalle.observaciones
        }))
      }

      const resultado = await NotaVenta.createNotaCompleta(nuevaNotaData)

      logger.business('Nota de venta duplicada', {
        notaOriginalId: id,
        notaNuevaId: resultado.nota.id,
        numeroNotaOriginal: notaOriginal.nota.numero_nota,
        numeroNotaNueva: resultado.nota.numero_nota,
        usuarioId: request.user.id
      })

      reply.code(201).send({
        success: true,
        message: 'Nota de venta duplicada exitosamente',
        data: resultado
      })
    } catch (error) {
      logger.error('Error al duplicar nota de venta:', error)
      
      if (error.message.includes('no encontrada')) {
        return reply.code(404).send({
          code: 'NOTA_NOT_FOUND',
          error: 'Not Found',
          message: 'Nota de venta no encontrada'
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
   * Exporta notas de venta a PDF
   */
  async exportarNotaPDF(request, reply) {
    try {
      const { id } = request.params

      const notaCompleta = await NotaVenta.getNotaCompleta(id)

      // Verificar permisos de acceso
      if (request.user.rol !== 'admin' && 
          request.user.sucursal_id !== notaCompleta.nota.sucursal_id) {
        return reply.code(403).send({
          code: 'INSUFFICIENT_PERMISSIONS',
          error: 'Forbidden',
          message: 'No tiene permisos para exportar esta nota de venta'
        })
      }

      // Aquí se implementaría la generación del PDF
      // Por ahora retornamos los datos para que el frontend genere el PDF
      reply.send({
        success: true,
        message: 'Datos preparados para exportación PDF',
        data: {
          nota: notaCompleta.nota,
          detalles: notaCompleta.detalles,
          empresa: {
            nombre: 'Ferretería Demo',
            rut: '76123456-7',
            direccion: 'Av. Principal 123',
            telefono: '+56912345678',
            email: 'contacto@ferreteria.cl'
          }
        }
      })
    } catch (error) {
      logger.error('Error al exportar nota a PDF:', error)
      
      if (error.message.includes('no encontrada')) {
        return reply.code(404).send({
          code: 'NOTA_NOT_FOUND',
          error: 'Not Found',
          message: 'Nota de venta no encontrada'
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
   * Obtiene el historial de una nota de venta
   */
  async getHistorialNota(request, reply) {
    try {
      const { id } = request.params

      // Verificar que la nota existe
      const nota = await NotaVenta.findById(id)
      if (!nota) {
        return reply.code(404).send({
          code: 'NOTA_NOT_FOUND',
          error: 'Not Found',
          message: 'Nota de venta no encontrada'
        })
      }

      // Verificar permisos de acceso
      if (request.user.rol !== 'admin' && 
          request.user.sucursal_id !== nota.sucursal_id) {
        return reply.code(403).send({
          code: 'INSUFFICIENT_PERMISSIONS',
          error: 'Forbidden',
          message: 'No tiene permisos para ver el historial de esta nota'
        })
      }

      // Obtener historial de cambios (esto requeriría una tabla de auditoría)
      // Por ahora retornamos información básica
      const historial = [
        {
          fecha: nota.fecha,
          accion: 'Creación',
          usuario: nota.vendedor_id,
          detalles: 'Nota de venta creada'
        }
      ]

      if (nota.fecha_conversion) {
        historial.push({
          fecha: nota.fecha_conversion,
          accion: 'Conversión',
          usuario: nota.usuario_conversion,
          detalles: `Convertida a venta ${nota.venta_id}`
        })
      }

      if (nota.fecha_anulacion) {
        historial.push({
          fecha: nota.fecha_anulacion,
          accion: 'Anulación',
          usuario: nota.usuario_anulacion,
          detalles: nota.motivo_anulacion
        })
      }

      reply.send({
        success: true,
        data: historial
      })
    } catch (error) {
      logger.error('Error al obtener historial de nota:', error)
      reply.code(500).send({
        code: 'INTERNAL_ERROR',
        error: 'Internal Server Error',
        message: 'Error interno del servidor'
      })
    }
  }
}

module.exports = new NotaVentaController()

