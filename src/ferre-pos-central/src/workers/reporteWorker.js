/**
 * Worker para Reportes - Sistema Ferre-POS
 * 
 * Procesa reportes complejos y estadísticas en un hilo separado
 * para no bloquear el hilo principal de la aplicación.
 */

const { parentPort, workerData } = require('worker_threads')
const database = require('../config/database')
const logger = require('../utils/logger')

/**
 * Procesador principal del worker
 */
async function procesarReporte(data) {
  const { tipo, parametros, jobId } = data
  
  try {
    logger.info(`Worker iniciando reporte ${tipo}`, { jobId, parametros })
    
    let resultado
    
    switch (tipo) {
      case 'ventas_periodo':
        resultado = await generarReporteVentas(parametros)
        break
        
      case 'productos_mas_vendidos':
        resultado = await generarReporteProductosVendidos(parametros)
        break
        
      case 'inventario_valorizado':
        resultado = await generarReporteInventario(parametros)
        break
        
      case 'estadisticas_usuarios':
        resultado = await generarEstadisticasUsuarios(parametros)
        break
        
      case 'analisis_rentabilidad':
        resultado = await generarAnalisisRentabilidad(parametros)
        break
        
      default:
        throw new Error(`Tipo de reporte no soportado: ${tipo}`)
    }
    
    logger.info(`Worker completó reporte ${tipo}`, { 
      jobId, 
      registros: resultado.data?.length || 0,
      duracion: Date.now() - parametros.startTime
    })
    
    // Enviar resultado al hilo principal
    parentPort.postMessage({
      jobId,
      success: true,
      data: resultado
    })
    
  } catch (error) {
    logger.error(`Error en worker de reporte ${tipo}:`, error)
    
    parentPort.postMessage({
      jobId,
      success: false,
      error: error.message
    })
  }
}

/**
 * Genera reporte de ventas por período
 */
async function generarReporteVentas(parametros) {
  const { fechaInicio, fechaFin, sucursalId, incluirDetalles = false } = parametros
  
  // Consulta base de ventas
  const ventasQuery = `
    SELECT 
      v.id,
      v.numero_venta,
      v.fecha,
      v.total,
      v.descuento_total,
      v.impuesto_total,
      u.nombre as vendedor,
      s.nombre as sucursal,
      c.nombre as cliente
    FROM ventas v
    LEFT JOIN usuarios u ON v.usuario_id = u.id
    LEFT JOIN sucursales s ON v.sucursal_id = s.id
    LEFT JOIN clientes c ON v.cliente_id = c.id
    WHERE v.fecha BETWEEN $1 AND $2
    ${sucursalId ? 'AND v.sucursal_id = $3' : ''}
    ORDER BY v.fecha DESC
  `
  
  const params = [fechaInicio, fechaFin]
  if (sucursalId) params.push(sucursalId)
  
  const ventasResult = await database.query(ventasQuery, params)
  
  // Estadísticas agregadas
  const estadisticasQuery = `
    SELECT 
      COUNT(*) as total_ventas,
      SUM(total) as total_facturado,
      AVG(total) as promedio_venta,
      SUM(descuento_total) as total_descuentos,
      SUM(impuesto_total) as total_impuestos,
      COUNT(DISTINCT cliente_id) as clientes_unicos,
      COUNT(DISTINCT usuario_id) as vendedores_activos
    FROM ventas
    WHERE fecha BETWEEN $1 AND $2
    ${sucursalId ? 'AND sucursal_id = $3' : ''}
  `
  
  const estadisticasResult = await database.query(estadisticasQuery, params)
  
  // Ventas por día
  const ventasPorDiaQuery = `
    SELECT 
      DATE(fecha) as dia,
      COUNT(*) as cantidad_ventas,
      SUM(total) as total_dia
    FROM ventas
    WHERE fecha BETWEEN $1 AND $2
    ${sucursalId ? 'AND sucursal_id = $3' : ''}
    GROUP BY DATE(fecha)
    ORDER BY dia
  `
  
  const ventasPorDiaResult = await database.query(ventasPorDiaQuery, params)
  
  let detallesVentas = []
  if (incluirDetalles) {
    // Obtener detalles de productos vendidos
    const detallesQuery = `
      SELECT 
        dv.venta_id,
        dv.producto_id,
        p.descripcion as producto,
        dv.cantidad,
        dv.precio_unitario,
        dv.subtotal
      FROM detalle_ventas dv
      JOIN productos p ON dv.producto_id = p.id
      JOIN ventas v ON dv.venta_id = v.id
      WHERE v.fecha BETWEEN $1 AND $2
      ${sucursalId ? 'AND v.sucursal_id = $3' : ''}
      ORDER BY dv.venta_id, p.descripcion
    `
    
    const detallesResult = await database.query(detallesQuery, params)
    detallesVentas = detallesResult.rows
  }
  
  return {
    periodo: { fechaInicio, fechaFin },
    estadisticas: estadisticasResult.rows[0],
    ventas: ventasResult.rows,
    ventasPorDia: ventasPorDiaResult.rows,
    detalles: detallesVentas,
    generadoEn: new Date().toISOString()
  }
}

/**
 * Genera reporte de productos más vendidos
 */
async function generarReporteProductosVendidos(parametros) {
  const { fechaInicio, fechaFin, sucursalId, limite = 50 } = parametros
  
  const query = `
    SELECT 
      p.id,
      p.codigo_interno,
      p.descripcion,
      p.marca,
      p.precio_unitario,
      SUM(dv.cantidad) as cantidad_vendida,
      SUM(dv.subtotal) as total_vendido,
      COUNT(DISTINCT dv.venta_id) as numero_ventas,
      AVG(dv.precio_unitario) as precio_promedio
    FROM detalle_ventas dv
    JOIN productos p ON dv.producto_id = p.id
    JOIN ventas v ON dv.venta_id = v.id
    WHERE v.fecha BETWEEN $1 AND $2
    ${sucursalId ? 'AND v.sucursal_id = $3' : ''}
    GROUP BY p.id, p.codigo_interno, p.descripcion, p.marca, p.precio_unitario
    ORDER BY cantidad_vendida DESC
    LIMIT $${sucursalId ? '4' : '3'}
  `
  
  const params = [fechaInicio, fechaFin]
  if (sucursalId) params.push(sucursalId)
  params.push(limite)
  
  const result = await database.query(query, params)
  
  // Estadísticas adicionales
  const resumenQuery = `
    SELECT 
      COUNT(DISTINCT p.id) as productos_vendidos,
      SUM(dv.cantidad) as total_unidades,
      SUM(dv.subtotal) as total_facturado
    FROM detalle_ventas dv
    JOIN productos p ON dv.producto_id = p.id
    JOIN ventas v ON dv.venta_id = v.id
    WHERE v.fecha BETWEEN $1 AND $2
    ${sucursalId ? 'AND v.sucursal_id = $3' : ''}
  `
  
  const resumenResult = await database.query(resumenQuery, params.slice(0, -1))
  
  return {
    periodo: { fechaInicio, fechaFin },
    resumen: resumenResult.rows[0],
    productos: result.rows,
    generadoEn: new Date().toISOString()
  }
}

/**
 * Genera reporte de inventario valorizado
 */
async function generarReporteInventario(parametros) {
  const { sucursalId, incluirInactivos = false } = parametros
  
  const query = `
    SELECT 
      p.id,
      p.codigo_interno,
      p.descripcion,
      p.marca,
      p.precio_unitario,
      p.precio_costo,
      s.cantidad_disponible,
      s.cantidad_reservada,
      s.stock_minimo,
      s.stock_maximo,
      (s.cantidad_disponible * p.precio_costo) as valor_costo,
      (s.cantidad_disponible * p.precio_unitario) as valor_venta,
      CASE 
        WHEN s.cantidad_disponible <= s.stock_minimo THEN 'BAJO'
        WHEN s.cantidad_disponible >= s.stock_maximo THEN 'ALTO'
        ELSE 'NORMAL'
      END as estado_stock
    FROM productos p
    JOIN stock s ON p.id = s.producto_id
    WHERE s.sucursal_id = $1
    ${!incluirInactivos ? 'AND p.activo = true' : ''}
    ORDER BY p.descripcion
  `
  
  const result = await database.query(query, [sucursalId])
  
  // Calcular totales
  const totales = result.rows.reduce((acc, item) => {
    acc.totalProductos++
    acc.totalUnidades += item.cantidad_disponible
    acc.valorTotalCosto += parseFloat(item.valor_costo || 0)
    acc.valorTotalVenta += parseFloat(item.valor_venta || 0)
    
    if (item.estado_stock === 'BAJO') acc.productosStockBajo++
    if (item.estado_stock === 'ALTO') acc.productosStockAlto++
    
    return acc
  }, {
    totalProductos: 0,
    totalUnidades: 0,
    valorTotalCosto: 0,
    valorTotalVenta: 0,
    productosStockBajo: 0,
    productosStockAlto: 0
  })
  
  return {
    sucursalId,
    totales,
    productos: result.rows,
    generadoEn: new Date().toISOString()
  }
}

/**
 * Genera estadísticas de usuarios
 */
async function generarEstadisticasUsuarios(parametros) {
  const { fechaInicio, fechaFin, sucursalId } = parametros
  
  // Estadísticas de accesos
  const accesosQuery = `
    SELECT 
      u.id,
      u.nombre,
      u.rol,
      COUNT(ia.id) as total_accesos,
      COUNT(CASE WHEN ia.exitoso THEN 1 END) as accesos_exitosos,
      COUNT(CASE WHEN NOT ia.exitoso THEN 1 END) as accesos_fallidos,
      MAX(ia.fecha) as ultimo_acceso
    FROM usuarios u
    LEFT JOIN intentos_acceso ia ON u.rut = ia.usuario_rut 
      AND ia.fecha BETWEEN $1 AND $2
    WHERE u.activo = true
    ${sucursalId ? 'AND u.sucursal_id = $3' : ''}
    GROUP BY u.id, u.nombre, u.rol
    ORDER BY total_accesos DESC
  `
  
  const params = [fechaInicio, fechaFin]
  if (sucursalId) params.push(sucursalId)
  
  const accesosResult = await database.query(accesosQuery, params)
  
  // Estadísticas de ventas por usuario
  const ventasQuery = `
    SELECT 
      u.id,
      u.nombre,
      COUNT(v.id) as total_ventas,
      SUM(v.total) as total_facturado,
      AVG(v.total) as promedio_venta
    FROM usuarios u
    LEFT JOIN ventas v ON u.id = v.usuario_id 
      AND v.fecha BETWEEN $1 AND $2
    WHERE u.activo = true
    ${sucursalId ? 'AND u.sucursal_id = $3' : ''}
    GROUP BY u.id, u.nombre
    ORDER BY total_ventas DESC
  `
  
  const ventasResult = await database.query(ventasQuery, params)
  
  // Combinar resultados
  const estadisticas = accesosResult.rows.map(acceso => {
    const venta = ventasResult.rows.find(v => v.id === acceso.id) || {}
    return {
      ...acceso,
      total_ventas: venta.total_ventas || 0,
      total_facturado: venta.total_facturado || 0,
      promedio_venta: venta.promedio_venta || 0
    }
  })
  
  return {
    periodo: { fechaInicio, fechaFin },
    estadisticas,
    generadoEn: new Date().toISOString()
  }
}

/**
 * Genera análisis de rentabilidad
 */
async function generarAnalisisRentabilidad(parametros) {
  const { fechaInicio, fechaFin, sucursalId } = parametros
  
  const query = `
    SELECT 
      p.id,
      p.codigo_interno,
      p.descripcion,
      p.marca,
      p.precio_costo,
      p.precio_unitario,
      SUM(dv.cantidad) as cantidad_vendida,
      SUM(dv.cantidad * p.precio_costo) as costo_total,
      SUM(dv.subtotal) as venta_total,
      SUM(dv.subtotal - (dv.cantidad * p.precio_costo)) as utilidad_bruta,
      CASE 
        WHEN SUM(dv.subtotal) > 0 
        THEN ((SUM(dv.subtotal - (dv.cantidad * p.precio_costo)) / SUM(dv.subtotal)) * 100)
        ELSE 0 
      END as margen_porcentaje
    FROM detalle_ventas dv
    JOIN productos p ON dv.producto_id = p.id
    JOIN ventas v ON dv.venta_id = v.id
    WHERE v.fecha BETWEEN $1 AND $2
    ${sucursalId ? 'AND v.sucursal_id = $3' : ''}
    GROUP BY p.id, p.codigo_interno, p.descripcion, p.marca, p.precio_costo, p.precio_unitario
    HAVING SUM(dv.cantidad) > 0
    ORDER BY utilidad_bruta DESC
  `
  
  const params = [fechaInicio, fechaFin]
  if (sucursalId) params.push(sucursalId)
  
  const result = await database.query(query, params)
  
  // Calcular totales
  const totales = result.rows.reduce((acc, item) => {
    acc.costoTotal += parseFloat(item.costo_total || 0)
    acc.ventaTotal += parseFloat(item.venta_total || 0)
    acc.utilidadBruta += parseFloat(item.utilidad_bruta || 0)
    return acc
  }, {
    costoTotal: 0,
    ventaTotal: 0,
    utilidadBruta: 0
  })
  
  totales.margenPromedio = totales.ventaTotal > 0 
    ? (totales.utilidadBruta / totales.ventaTotal) * 100 
    : 0
  
  return {
    periodo: { fechaInicio, fechaFin },
    totales,
    productos: result.rows,
    generadoEn: new Date().toISOString()
  }
}

// Escuchar mensajes del hilo principal
if (parentPort) {
  parentPort.on('message', procesarReporte)
} else {
  // Ejecutar directamente si se llama como script independiente
  if (workerData) {
    procesarReporte(workerData)
  }
}

