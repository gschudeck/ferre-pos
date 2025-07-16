/**
 * Tests para Usuarios - Sistema Ferre-POS
 * 
 * Pruebas unitarias e integración para el módulo de usuarios
 */

const { test, beforeAll, afterAll, beforeEach } = require('tap')
const { createServer } = require('../src/server')
const database = require('../src/config/database')

let server
let adminToken
let gerenteToken
let vendedorToken
let testUserId

beforeAll(async () => {
  // Configurar entorno de test
  process.env.NODE_ENV = 'test'
  process.env.DB_NAME = 'ferre_pos_test'
  
  // Crear servidor de test
  server = await createServer()
  
  // Limpiar y preparar base de datos de test
  await setupTestDatabase()
  
  // Obtener tokens de autenticación
  adminToken = await getAuthToken('admin')
  gerenteToken = await getAuthToken('gerente')
  vendedorToken = await getAuthToken('vendedor')
})

afterAll(async () => {
  await server.close()
  await database.disconnect()
})

beforeEach(async () => {
  // Limpiar usuarios de test antes de cada prueba
  await cleanupTestUsers()
})

/**
 * Configuración inicial de base de datos de test
 */
async function setupTestDatabase() {
  // Crear tabla de usuarios si no existe
  await database.query(`
    CREATE TABLE IF NOT EXISTS usuarios (
      id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
      rut VARCHAR(12) UNIQUE NOT NULL,
      nombre VARCHAR(100) NOT NULL,
      email VARCHAR(100) UNIQUE NOT NULL,
      telefono VARCHAR(20),
      rol VARCHAR(20) NOT NULL,
      sucursal_id UUID,
      password_hash TEXT NOT NULL,
      salt TEXT NOT NULL,
      activo BOOLEAN DEFAULT true,
      ultimo_acceso TIMESTAMP,
      intentos_fallidos INTEGER DEFAULT 0,
      bloqueado_hasta TIMESTAMP,
      debe_cambiar_password BOOLEAN DEFAULT true,
      token_recuperacion TEXT,
      token_recuperacion_expira TIMESTAMP,
      fecha_creacion TIMESTAMP DEFAULT NOW(),
      fecha_modificacion TIMESTAMP DEFAULT NOW(),
      usuario_creacion UUID,
      usuario_modificacion UUID
    )
  `)

  // Crear tabla de intentos de acceso
  await database.query(`
    CREATE TABLE IF NOT EXISTS intentos_acceso (
      id SERIAL PRIMARY KEY,
      usuario_rut VARCHAR(12) NOT NULL,
      exitoso BOOLEAN NOT NULL,
      motivo VARCHAR(100),
      ip_address INET,
      fecha TIMESTAMP DEFAULT NOW()
    )
  `)

  // Crear tabla de auditoría de usuarios
  await database.query(`
    CREATE TABLE IF NOT EXISTS auditoria_usuarios (
      id SERIAL PRIMARY KEY,
      usuario_id UUID NOT NULL,
      accion VARCHAR(50) NOT NULL,
      detalles JSONB,
      realizado_por_id UUID,
      fecha TIMESTAMP DEFAULT NOW()
    )
  `)

  // Crear tabla de sucursales para tests
  await database.query(`
    CREATE TABLE IF NOT EXISTS sucursales (
      id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
      nombre VARCHAR(100) NOT NULL,
      habilitada BOOLEAN DEFAULT true
    )
  `)

  // Insertar sucursal de test
  await database.query(`
    INSERT INTO sucursales (id, nombre, habilitada)
    VALUES ('550e8400-e29b-41d4-a716-446655440000', 'Sucursal Test', true)
    ON CONFLICT (id) DO NOTHING
  `)

  // Crear usuarios de test con contraseñas hasheadas
  const bcrypt = require('bcrypt')
  const salt = await bcrypt.genSalt(12)
  const hash = await bcrypt.hash('Test123!', salt)

  await database.query(`
    INSERT INTO usuarios (id, rut, nombre, email, rol, password_hash, salt, activo, debe_cambiar_password, sucursal_id)
    VALUES 
      ('11111111-1111-1111-1111-111111111111', '11111111-1', 'Admin Test', 'admin@test.com', 'admin', $1, $2, true, false, null),
      ('22222222-2222-2222-2222-222222222222', '22222222-2', 'Gerente Test', 'gerente@test.com', 'gerente', $1, $2, true, false, '550e8400-e29b-41d4-a716-446655440000'),
      ('33333333-3333-3333-3333-333333333333', '33333333-3', 'Vendedor Test', 'vendedor@test.com', 'vendedor', $1, $2, true, false, '550e8400-e29b-41d4-a716-446655440000')
    ON CONFLICT (rut) DO NOTHING
  `, [hash, salt])
}

/**
 * Obtener token de autenticación para tests
 */
async function getAuthToken(rol) {
  const rutMap = {
    admin: '11111111-1',
    gerente: '22222222-2',
    vendedor: '33333333-3'
  }
  
  const response = await server.inject({
    method: 'POST',
    url: '/api/auth/login',
    payload: {
      rut: rutMap[rol],
      password: 'Test123!'
    }
  })
  
  return JSON.parse(response.payload).data.token
}

/**
 * Limpiar usuarios de test
 */
async function cleanupTestUsers() {
  await database.query(`
    DELETE FROM usuarios 
    WHERE rut LIKE 'test%' OR email LIKE '%test-temp%'
  `)
}

/**
 * Tests de gestión de usuarios
 */
test('Obtener lista de usuarios - admin', async (t) => {
  const response = await server.inject({
    method: 'GET',
    url: '/api/usuarios',
    headers: {
      authorization: `Bearer ${adminToken}`
    }
  })

  t.equal(response.statusCode, 200)
  
  const result = JSON.parse(response.payload)
  t.ok(result.success)
  t.ok(Array.isArray(result.data))
  t.ok(result.pagination)
  t.ok(result.data.length >= 3) // Al menos los 3 usuarios de test
})

test('Obtener lista de usuarios con filtros', async (t) => {
  const response = await server.inject({
    method: 'GET',
    url: '/api/usuarios?rol=admin&activo=true&limit=5',
    headers: {
      authorization: `Bearer ${adminToken}`
    }
  })

  t.equal(response.statusCode, 200)
  
  const result = JSON.parse(response.payload)
  t.ok(result.success)
  t.ok(result.data.every(user => user.rol === 'admin'))
  t.ok(result.data.every(user => user.activo === true))
})

test('Error de permisos para lista de usuarios - vendedor', async (t) => {
  const response = await server.inject({
    method: 'GET',
    url: '/api/usuarios',
    headers: {
      authorization: `Bearer ${vendedorToken}`
    }
  })

  t.equal(response.statusCode, 403)
  
  const result = JSON.parse(response.payload)
  t.equal(result.code, 'INSUFFICIENT_PERMISSIONS')
})

test('Crear nuevo usuario - admin', async (t) => {
  const userData = {
    rut: 'test12345-6',
    nombre: 'Usuario Test',
    email: 'usuario-test-temp@test.com',
    telefono: '+56912345678',
    rol: 'vendedor',
    sucursal_id: '550e8400-e29b-41d4-a716-446655440000',
    password: 'TestPass123!',
    debe_cambiar_password: true
  }

  const response = await server.inject({
    method: 'POST',
    url: '/api/usuarios',
    headers: {
      authorization: `Bearer ${adminToken}`
    },
    payload: userData
  })

  t.equal(response.statusCode, 201)
  
  const result = JSON.parse(response.payload)
  t.ok(result.success)
  t.equal(result.data.rut, userData.rut)
  t.equal(result.data.nombre, userData.nombre)
  t.equal(result.data.email, userData.email)
  t.equal(result.data.rol, userData.rol)
  
  // Guardar ID para otros tests
  testUserId = result.data.id
})

test('Error al crear usuario con RUT duplicado', async (t) => {
  const userData = {
    rut: '11111111-1', // RUT que ya existe
    nombre: 'Usuario Duplicado',
    email: 'duplicado-test-temp@test.com',
    rol: 'vendedor',
    password: 'TestPass123!'
  }

  const response = await server.inject({
    method: 'POST',
    url: '/api/usuarios',
    headers: {
      authorization: `Bearer ${adminToken}`
    },
    payload: userData
  })

  t.equal(response.statusCode, 409)
  
  const result = JSON.parse(response.payload)
  t.equal(result.code, 'USER_ALREADY_EXISTS')
})

test('Error al crear usuario con email duplicado', async (t) => {
  const userData = {
    rut: 'test99999-9',
    nombre: 'Usuario Email Duplicado',
    email: 'admin@test.com', // Email que ya existe
    rol: 'vendedor',
    password: 'TestPass123!'
  }

  const response = await server.inject({
    method: 'POST',
    url: '/api/usuarios',
    headers: {
      authorization: `Bearer ${adminToken}`
    },
    payload: userData
  })

  t.equal(response.statusCode, 409)
  
  const result = JSON.parse(response.payload)
  t.equal(result.code, 'USER_ALREADY_EXISTS')
})

test('Error de permisos para crear usuario - vendedor', async (t) => {
  const userData = {
    rut: 'test88888-8',
    nombre: 'Usuario Sin Permisos',
    email: 'sinpermisos-test-temp@test.com',
    rol: 'vendedor',
    password: 'TestPass123!'
  }

  const response = await server.inject({
    method: 'POST',
    url: '/api/usuarios',
    headers: {
      authorization: `Bearer ${vendedorToken}`
    },
    payload: userData
  })

  t.equal(response.statusCode, 403)
  
  const result = JSON.parse(response.payload)
  t.equal(result.code, 'INSUFFICIENT_PERMISSIONS')
})

test('Obtener usuario específico por ID', async (t) => {
  // Primero crear un usuario
  const createResponse = await server.inject({
    method: 'POST',
    url: '/api/usuarios',
    headers: {
      authorization: `Bearer ${adminToken}`
    },
    payload: {
      rut: 'test77777-7',
      nombre: 'Usuario Específico',
      email: 'especifico-test-temp@test.com',
      rol: 'cajero',
      password: 'TestPass123!'
    }
  })

  const createdUser = JSON.parse(createResponse.payload).data

  // Obtener el usuario
  const response = await server.inject({
    method: 'GET',
    url: `/api/usuarios/${createdUser.id}`,
    headers: {
      authorization: `Bearer ${adminToken}`
    }
  })

  t.equal(response.statusCode, 200)
  
  const result = JSON.parse(response.payload)
  t.ok(result.success)
  t.equal(result.data.id, createdUser.id)
  t.equal(result.data.rut, 'test77777-7')
  t.equal(result.data.nombre, 'Usuario Específico')
})

test('Error al obtener usuario inexistente', async (t) => {
  const response = await server.inject({
    method: 'GET',
    url: '/api/usuarios/99999999-9999-9999-9999-999999999999',
    headers: {
      authorization: `Bearer ${adminToken}`
    }
  })

  t.equal(response.statusCode, 404)
  
  const result = JSON.parse(response.payload)
  t.equal(result.code, 'USER_NOT_FOUND')
})

/**
 * Tests de actualización de usuarios
 */
test('Actualizar usuario existente', async (t) => {
  // Crear usuario para actualizar
  const createResponse = await server.inject({
    method: 'POST',
    url: '/api/usuarios',
    headers: {
      authorization: `Bearer ${adminToken}`
    },
    payload: {
      rut: 'test66666-6',
      nombre: 'Usuario Original',
      email: 'original-test-temp@test.com',
      rol: 'vendedor',
      password: 'TestPass123!'
    }
  })

  const userId = JSON.parse(createResponse.payload).data.id

  // Actualizar usuario
  const updateData = {
    nombre: 'Usuario Actualizado',
    telefono: '+56987654321'
  }

  const response = await server.inject({
    method: 'PUT',
    url: `/api/usuarios/${userId}`,
    headers: {
      authorization: `Bearer ${adminToken}`
    },
    payload: updateData
  })

  t.equal(response.statusCode, 200)
  
  const result = JSON.parse(response.payload)
  t.ok(result.success)
  t.equal(result.data.nombre, updateData.nombre)
  t.equal(result.data.telefono, updateData.telefono)
})

/**
 * Tests de perfil de usuario
 */
test('Obtener perfil propio', async (t) => {
  const response = await server.inject({
    method: 'GET',
    url: '/api/usuarios/perfil',
    headers: {
      authorization: `Bearer ${vendedorToken}`
    }
  })

  t.equal(response.statusCode, 200)
  
  const result = JSON.parse(response.payload)
  t.ok(result.success)
  t.equal(result.data.rut, '33333333-3')
  t.equal(result.data.nombre, 'Vendedor Test')
  t.equal(result.data.rol, 'vendedor')
})

test('Actualizar perfil propio', async (t) => {
  const updateData = {
    nombre: 'Vendedor Actualizado',
    telefono: '+56911111111'
  }

  const response = await server.inject({
    method: 'PUT',
    url: '/api/usuarios/perfil',
    headers: {
      authorization: `Bearer ${vendedorToken}`
    },
    payload: updateData
  })

  t.equal(response.statusCode, 200)
  
  const result = JSON.parse(response.payload)
  t.ok(result.success)
  t.equal(result.data.nombre, updateData.nombre)
  t.equal(result.data.telefono, updateData.telefono)
})

/**
 * Tests de cambio de contraseña
 */
test('Cambiar contraseña propia', async (t) => {
  const response = await server.inject({
    method: 'PUT',
    url: '/api/usuarios/33333333-3333-3333-3333-333333333333/password',
    headers: {
      authorization: `Bearer ${vendedorToken}`
    },
    payload: {
      password_actual: 'Test123!',
      password_nueva: 'NewPass123!'
    }
  })

  t.equal(response.statusCode, 200)
  
  const result = JSON.parse(response.payload)
  t.ok(result.success)
})

test('Error al cambiar contraseña con contraseña actual incorrecta', async (t) => {
  const response = await server.inject({
    method: 'PUT',
    url: '/api/usuarios/33333333-3333-3333-3333-333333333333/password',
    headers: {
      authorization: `Bearer ${vendedorToken}`
    },
    payload: {
      password_actual: 'PasswordIncorrecta',
      password_nueva: 'NewPass123!'
    }
  })

  t.equal(response.statusCode, 400)
  
  const result = JSON.parse(response.payload)
  t.equal(result.code, 'INVALID_CURRENT_PASSWORD')
})

test('Admin cambia contraseña de otro usuario', async (t) => {
  const response = await server.inject({
    method: 'PUT',
    url: '/api/usuarios/33333333-3333-3333-3333-333333333333/password',
    headers: {
      authorization: `Bearer ${adminToken}`
    },
    payload: {
      password_nueva: 'AdminNewPass123!'
    }
  })

  t.equal(response.statusCode, 200)
  
  const result = JSON.parse(response.payload)
  t.ok(result.success)
})

/**
 * Tests de recuperación de contraseña
 */
test('Iniciar recuperación de contraseña', async (t) => {
  const response = await server.inject({
    method: 'POST',
    url: '/api/usuarios/recuperar-password',
    payload: {
      email: 'vendedor@test.com'
    }
  })

  t.equal(response.statusCode, 200)
  
  const result = JSON.parse(response.payload)
  t.ok(result.success)
  t.ok(result.message.includes('Instrucciones enviadas'))
})

test('Recuperación de contraseña con email inexistente', async (t) => {
  const response = await server.inject({
    method: 'POST',
    url: '/api/usuarios/recuperar-password',
    payload: {
      email: 'inexistente@test.com'
    }
  })

  t.equal(response.statusCode, 200)
  
  const result = JSON.parse(response.payload)
  t.ok(result.success)
  // Por seguridad, siempre retorna éxito
})

/**
 * Tests de desactivación y reactivación
 */
test('Desactivar usuario', async (t) => {
  // Crear usuario para desactivar
  const createResponse = await server.inject({
    method: 'POST',
    url: '/api/usuarios',
    headers: {
      authorization: `Bearer ${adminToken}`
    },
    payload: {
      rut: 'test55555-5',
      nombre: 'Usuario Para Desactivar',
      email: 'desactivar-test-temp@test.com',
      rol: 'cajero',
      password: 'TestPass123!'
    }
  })

  const userId = JSON.parse(createResponse.payload).data.id

  const response = await server.inject({
    method: 'POST',
    url: `/api/usuarios/${userId}/desactivar`,
    headers: {
      authorization: `Bearer ${adminToken}`
    },
    payload: {
      motivo: 'Usuario ya no trabaja en la empresa'
    }
  })

  t.equal(response.statusCode, 200)
  
  const result = JSON.parse(response.payload)
  t.ok(result.success)
})

test('Reactivar usuario', async (t) => {
  // Crear y desactivar usuario
  const createResponse = await server.inject({
    method: 'POST',
    url: '/api/usuarios',
    headers: {
      authorization: `Bearer ${adminToken}`
    },
    payload: {
      rut: 'test44444-4',
      nombre: 'Usuario Para Reactivar',
      email: 'reactivar-test-temp@test.com',
      rol: 'cajero',
      password: 'TestPass123!'
    }
  })

  const userId = JSON.parse(createResponse.payload).data.id

  await server.inject({
    method: 'POST',
    url: `/api/usuarios/${userId}/desactivar`,
    headers: {
      authorization: `Bearer ${adminToken}`
    },
    payload: {
      motivo: 'Test de desactivación'
    }
  })

  // Reactivar
  const response = await server.inject({
    method: 'POST',
    url: `/api/usuarios/${userId}/reactivar`,
    headers: {
      authorization: `Bearer ${adminToken}`
    }
  })

  t.equal(response.statusCode, 200)
  
  const result = JSON.parse(response.payload)
  t.ok(result.success)
})

/**
 * Tests de estadísticas
 */
test('Obtener estadísticas de usuarios - admin', async (t) => {
  const response = await server.inject({
    method: 'GET',
    url: '/api/usuarios/estadisticas',
    headers: {
      authorization: `Bearer ${adminToken}`
    }
  })

  t.equal(response.statusCode, 200)
  
  const result = JSON.parse(response.payload)
  t.ok(result.success)
  t.ok(typeof result.data.total_usuarios === 'number')
  t.ok(typeof result.data.usuarios_activos === 'number')
  t.ok(typeof result.data.administradores === 'number')
})

test('Error de permisos para estadísticas - vendedor', async (t) => {
  const response = await server.inject({
    method: 'GET',
    url: '/api/usuarios/estadisticas',
    headers: {
      authorization: `Bearer ${vendedorToken}`
    }
  })

  t.equal(response.statusCode, 403)
  
  const result = JSON.parse(response.payload)
  t.equal(result.code, 'INSUFFICIENT_PERMISSIONS')
})

/**
 * Tests de mantenimiento
 */
test('Ejecutar mantenimiento de usuarios', async (t) => {
  const response = await server.inject({
    method: 'POST',
    url: '/api/usuarios/mantenimiento',
    headers: {
      authorization: `Bearer ${adminToken}`
    }
  })

  t.equal(response.statusCode, 200)
  
  const result = JSON.parse(response.payload)
  t.ok(result.success)
  t.ok(Array.isArray(result.data.tareas_ejecutadas))
})

test('Error de permisos para mantenimiento - gerente', async (t) => {
  const response = await server.inject({
    method: 'POST',
    url: '/api/usuarios/mantenimiento',
    headers: {
      authorization: `Bearer ${gerenteToken}`
    }
  })

  t.equal(response.statusCode, 403)
  
  const result = JSON.parse(response.payload)
  t.equal(result.code, 'INSUFFICIENT_PERMISSIONS')
})

/**
 * Tests de validación
 */
test('Error de validación - contraseña débil', async (t) => {
  const userData = {
    rut: 'test33333-3',
    nombre: 'Usuario Contraseña Débil',
    email: 'debil-test-temp@test.com',
    rol: 'vendedor',
    password: '123' // Contraseña muy débil
  }

  const response = await server.inject({
    method: 'POST',
    url: '/api/usuarios',
    headers: {
      authorization: `Bearer ${adminToken}`
    },
    payload: userData
  })

  t.equal(response.statusCode, 400)
  
  const result = JSON.parse(response.payload)
  t.equal(result.code, 'VALIDATION_ERROR')
})

test('Error de validación - RUT inválido', async (t) => {
  const userData = {
    rut: 'rut-invalido',
    nombre: 'Usuario RUT Inválido',
    email: 'rutinvalido-test-temp@test.com',
    rol: 'vendedor',
    password: 'TestPass123!'
  }

  const response = await server.inject({
    method: 'POST',
    url: '/api/usuarios',
    headers: {
      authorization: `Bearer ${adminToken}`
    },
    payload: userData
  })

  t.equal(response.statusCode, 400)
  
  const result = JSON.parse(response.payload)
  t.equal(result.code, 'VALIDATION_ERROR')
})

test('Error de validación - email inválido', async (t) => {
  const userData = {
    rut: 'test22222-2',
    nombre: 'Usuario Email Inválido',
    email: 'email-invalido',
    rol: 'vendedor',
    password: 'TestPass123!'
  }

  const response = await server.inject({
    method: 'POST',
    url: '/api/usuarios',
    headers: {
      authorization: `Bearer ${adminToken}`
    },
    payload: userData
  })

  t.equal(response.statusCode, 400)
  
  const result = JSON.parse(response.payload)
  t.equal(result.code, 'VALIDATION_ERROR')
})

