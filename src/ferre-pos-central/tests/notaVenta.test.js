/**
 * Tests para Notas de Venta - Sistema Ferre-POS
 * 
 * Pruebas unitarias e integración para el módulo de notas de venta
 */

const { test, beforeAll, afterAll, beforeEach } = require('tap')
const { createServer } = require('../src/server')
const database = require('../src/config/database')

let server
let authToken
let testSucursalId
let testProductoId
let testVendedorId

beforeAll(async () => {
  // Configurar entorno de test
  process.env.NODE_ENV = 'test'
  process.env.DB_NAME = 'ferre_pos_test'
  
  // Crear servidor de test
  server = await createServer()
  
  // Limpiar y preparar base de datos de test
  await setupTestDatabase()
  
  // Obtener token de autenticación
  authToken = await getAuthToken()
})

afterAll(async () => {
  await server.close()
  await database.disconnect()
})

beforeEach(async () => {
  // Limpiar datos de test antes de cada prueba
  await cleanupTestData()
})

/**
 * Configuración inicial de base de datos de test
 */
async function setupTestDatabase() {
  // Crear sucursal de test
  const sucursalResult = await database.query(`
    INSERT INTO sucursales (codigo, nombre, direccion, habilitada)
    VALUES ('TEST', 'Sucursal Test', 'Dirección Test', true)
    RETURNING id
  `)
  testSucursalId = sucursalResult.rows[0].id

  // Crear vendedor de test
  const vendedorResult = await database.query(`
    INSERT INTO usuarios (rut, nombre, email, rol, sucursal_id, password_hash, salt, activo)
    VALUES ('12345678-9', 'Vendedor Test', 'vendedor@test.com', 'vendedor', $1, 'hash', 'salt', true)
    RETURNING id
  `, [testSucursalId])
  testVendedorId = vendedorResult.rows[0].id

  // Crear categoría de test
  const categoriaResult = await database.query(`
    INSERT INTO categorias_productos (codigo, nombre, activa)
    VALUES ('TEST', 'Categoría Test', true)
    RETURNING id
  `)
  const testCategoriaId = categoriaResult.rows[0].id

  // Crear producto de test
  const productoResult = await database.query(`
    INSERT INTO productos (codigo_interno, codigo_barra, descripcion, categoria_id, 
                          precio_unitario, precio_costo, unidad_medida, activo, usuario_creacion)
    VALUES ('TEST001', '1234567890', 'Producto Test', $1, 1000, 500, 'UN', true, $2)
    RETURNING id
  `, [testCategoriaId, testVendedorId])
  testProductoId = productoResult.rows[0].id

  // Crear stock inicial
  await database.query(`
    INSERT INTO stock_central (producto_id, sucursal_id, cantidad, costo_promedio)
    VALUES ($1, $2, 100, 500)
  `, [testProductoId, testSucursalId])
}

/**
 * Obtener token de autenticación para tests
 */
async function getAuthToken() {
  const response = await server.inject({
    method: 'POST',
    url: '/api/auth/login',
    payload: {
      rut: '12345678-9',
      password: 'test123'
    }
  })
  
  return JSON.parse(response.payload).data.token
}

/**
 * Limpiar datos de test
 */
async function cleanupTestData() {
  await database.query('DELETE FROM detalle_notas_venta WHERE 1=1')
  await database.query('DELETE FROM notas_venta WHERE 1=1')
  await database.query('DELETE FROM reservas_stock WHERE 1=1')
}

/**
 * Tests de creación de notas de venta
 */
test('Crear nota de venta - cotización', async (t) => {
  const notaData = {
    nota: {
      sucursal_id: testSucursalId,
      cliente_rut: '11111111-1',
      cliente_nombre: 'Cliente Test',
      cliente_email: 'cliente@test.com',
      tipo_nota: 'cotizacion',
      subtotal: 1000,
      descuento_total: 0,
      impuesto_total: 190,
      total: 1190,
      observaciones: 'Cotización de prueba'
    },
    detalles: [
      {
        producto_id: testProductoId,
        cantidad: 1,
        precio_unitario: 1000,
        descuento_unitario: 0,
        precio_final: 1000,
        total_item: 1000
      }
    ]
  }

  const response = await server.inject({
    method: 'POST',
    url: '/api/notas-venta',
    headers: {
      authorization: `Bearer ${authToken}`
    },
    payload: notaData
  })

  t.equal(response.statusCode, 201)
  
  const result = JSON.parse(response.payload)
  t.ok(result.success)
  t.equal(result.data.nota.tipo_nota, 'cotizacion')
  t.equal(result.data.nota.total, 1190)
  t.equal(result.data.detalles.length, 1)
})

test('Crear nota de venta - reserva', async (t) => {
  const notaData = {
    nota: {
      sucursal_id: testSucursalId,
      cliente_rut: '22222222-2',
      cliente_nombre: 'Cliente Reserva',
      tipo_nota: 'reserva',
      subtotal: 2000,
      descuento_total: 100,
      impuesto_total: 361,
      total: 2261
    },
    detalles: [
      {
        producto_id: testProductoId,
        cantidad: 2,
        precio_unitario: 1000,
        descuento_unitario: 50,
        precio_final: 950,
        total_item: 1900
      }
    ]
  }

  const response = await server.inject({
    method: 'POST',
    url: '/api/notas-venta',
    headers: {
      authorization: `Bearer ${authToken}`
    },
    payload: notaData
  })

  t.equal(response.statusCode, 201)
  
  const result = JSON.parse(response.payload)
  t.ok(result.success)
  t.equal(result.data.nota.tipo_nota, 'reserva')
  t.equal(result.data.nota.estado, 'activa')
  
  // Verificar que se creó la reserva de stock
  const stockResult = await database.query(`
    SELECT cantidad_reservada FROM stock_central 
    WHERE producto_id = $1 AND sucursal_id = $2
  `, [testProductoId, testSucursalId])
  
  t.equal(stockResult.rows[0].cantidad_reservada, 2)
})

test('Error al crear nota con stock insuficiente', async (t) => {
  const notaData = {
    nota: {
      sucursal_id: testSucursalId,
      tipo_nota: 'reserva',
      subtotal: 150000,
      total: 150000
    },
    detalles: [
      {
        producto_id: testProductoId,
        cantidad: 150, // Más del stock disponible (100)
        precio_unitario: 1000,
        precio_final: 1000,
        total_item: 150000
      }
    ]
  }

  const response = await server.inject({
    method: 'POST',
    url: '/api/notas-venta',
    headers: {
      authorization: `Bearer ${authToken}`
    },
    payload: notaData
  })

  t.equal(response.statusCode, 400)
  
  const result = JSON.parse(response.payload)
  t.equal(result.code, 'INSUFFICIENT_STOCK')
  t.match(result.message, /Stock insuficiente/)
})

test('Error de validación - totales no coinciden', async (t) => {
  const notaData = {
    nota: {
      sucursal_id: testSucursalId,
      tipo_nota: 'cotizacion',
      subtotal: 1000,
      total: 2000 // Total incorrecto
    },
    detalles: [
      {
        producto_id: testProductoId,
        cantidad: 1,
        precio_unitario: 1000,
        precio_final: 1000,
        total_item: 1000
      }
    ]
  }

  const response = await server.inject({
    method: 'POST',
    url: '/api/notas-venta',
    headers: {
      authorization: `Bearer ${authToken}`
    },
    payload: notaData
  })

  t.equal(response.statusCode, 400)
  
  const result = JSON.parse(response.payload)
  t.equal(result.code, 'INVALID_TOTALS')
})

/**
 * Tests de consulta de notas de venta
 */
test('Obtener lista de notas de venta', async (t) => {
  // Crear nota de test
  await createTestNota()

  const response = await server.inject({
    method: 'GET',
    url: '/api/notas-venta',
    headers: {
      authorization: `Bearer ${authToken}`
    }
  })

  t.equal(response.statusCode, 200)
  
  const result = JSON.parse(response.payload)
  t.ok(result.success)
  t.ok(Array.isArray(result.data))
  t.ok(result.pagination)
})

test('Buscar notas de venta', async (t) => {
  // Crear nota de test
  const nota = await createTestNota()

  const response = await server.inject({
    method: 'GET',
    url: `/api/notas-venta/search?clienteRut=11111111-1`,
    headers: {
      authorization: `Bearer ${authToken}`
    }
  })

  t.equal(response.statusCode, 200)
  
  const result = JSON.parse(response.payload)
  t.ok(result.success)
  t.equal(result.data.length, 1)
  t.equal(result.data[0].cliente_rut, '11111111-1')
})

test('Obtener nota de venta por ID', async (t) => {
  // Crear nota de test
  const nota = await createTestNota()

  const response = await server.inject({
    method: 'GET',
    url: `/api/notas-venta/${nota.id}`,
    headers: {
      authorization: `Bearer ${authToken}`
    }
  })

  t.equal(response.statusCode, 200)
  
  const result = JSON.parse(response.payload)
  t.ok(result.success)
  t.equal(result.data.nota.id, nota.id)
  t.ok(Array.isArray(result.data.detalles))
})

/**
 * Tests de conversión a venta
 */
test('Convertir nota de venta a venta real', async (t) => {
  // Crear terminal de test
  const terminalResult = await database.query(`
    INSERT INTO terminales (sucursal_id, nombre_terminal, ubicacion, habilitado)
    VALUES ($1, 'Terminal Test', 'Caja Test', true)
    RETURNING id
  `, [testSucursalId])
  const terminalId = terminalResult.rows[0].id

  // Crear nota de test
  const nota = await createTestNota()

  const conversionData = {
    terminal_id: terminalId,
    tipo_documento: 'boleta',
    mediosPago: [
      {
        medio_pago: 'efectivo',
        monto: 1190
      }
    ]
  }

  const response = await server.inject({
    method: 'POST',
    url: `/api/notas-venta/${nota.id}/convertir`,
    headers: {
      authorization: `Bearer ${authToken}`
    },
    payload: conversionData
  })

  t.equal(response.statusCode, 200)
  
  const result = JSON.parse(response.payload)
  t.ok(result.success)
  t.ok(result.data.venta)
  t.equal(result.data.venta.total, 1190)
  
  // Verificar que la nota cambió de estado
  const notaActualizada = await database.query(`
    SELECT estado FROM notas_venta WHERE id = $1
  `, [nota.id])
  
  t.equal(notaActualizada.rows[0].estado, 'convertida')
})

/**
 * Tests de anulación
 */
test('Anular nota de venta', async (t) => {
  // Crear nota de test
  const nota = await createTestNota()

  const response = await server.inject({
    method: 'POST',
    url: `/api/notas-venta/${nota.id}/anular`,
    headers: {
      authorization: `Bearer ${authToken}`
    },
    payload: {
      motivo: 'Anulación de prueba para testing'
    }
  })

  t.equal(response.statusCode, 200)
  
  const result = JSON.parse(response.payload)
  t.ok(result.success)
  t.equal(result.data.estado, 'anulada')
})

/**
 * Tests de estadísticas
 */
test('Obtener estadísticas de notas de venta', async (t) => {
  // Crear algunas notas de test
  await createTestNota()
  await createTestNota('reserva')

  const response = await server.inject({
    method: 'GET',
    url: '/api/notas-venta/stats',
    headers: {
      authorization: `Bearer ${authToken}`
    }
  })

  t.equal(response.statusCode, 200)
  
  const result = JSON.parse(response.payload)
  t.ok(result.success)
  t.ok(result.data.total_notas >= 2)
  t.ok(result.data.cotizaciones >= 1)
  t.ok(result.data.reservas >= 1)
})

/**
 * Función auxiliar para crear nota de test
 */
async function createTestNota(tipoNota = 'cotizacion') {
  const notaResult = await database.query(`
    INSERT INTO notas_venta (
      sucursal_id, vendedor_id, cliente_rut, cliente_nombre,
      tipo_nota, subtotal, impuesto_total, total, estado
    ) VALUES ($1, $2, '11111111-1', 'Cliente Test', $3, 1000, 190, 1190, 'activa')
    RETURNING *
  `, [testSucursalId, testVendedorId, tipoNota])

  const nota = notaResult.rows[0]

  // Crear detalle
  await database.query(`
    INSERT INTO detalle_notas_venta (
      nota_venta_id, producto_id, cantidad, precio_unitario,
      precio_final, total_item
    ) VALUES ($1, $2, 1, 1000, 1000, 1000)
  `, [nota.id, testProductoId])

  return nota
}

