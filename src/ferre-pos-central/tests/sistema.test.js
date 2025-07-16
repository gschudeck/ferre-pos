/**
 * Tests para Sistema - Sistema Ferre-POS
 * 
 * Pruebas unitarias e integración para el módulo de sistema
 */

const { test, beforeAll, afterAll, beforeEach } = require('tap')
const { createServer } = require('../src/server')
const database = require('../src/config/database')

let server
let authToken
let adminToken

beforeAll(async () => {
  // Configurar entorno de test
  process.env.NODE_ENV = 'test'
  process.env.DB_NAME = 'ferre_pos_test'
  
  // Crear servidor de test
  server = await createServer()
  
  // Limpiar y preparar base de datos de test
  await setupTestDatabase()
  
  // Obtener tokens de autenticación
  authToken = await getAuthToken('vendedor')
  adminToken = await getAuthToken('admin')
})

afterAll(async () => {
  await server.close()
  await database.disconnect()
})

beforeEach(async () => {
  // Limpiar configuraciones de test antes de cada prueba
  await cleanupTestConfigurations()
})

/**
 * Configuración inicial de base de datos de test
 */
async function setupTestDatabase() {
  // Crear tabla de configuraciones si no existe
  await database.query(`
    CREATE TABLE IF NOT EXISTS configuracion_sistema (
      id SERIAL PRIMARY KEY,
      clave VARCHAR(100) UNIQUE NOT NULL,
      valor TEXT NOT NULL,
      descripcion TEXT,
      tipo_dato VARCHAR(20) NOT NULL,
      categoria VARCHAR(50) NOT NULL,
      activa BOOLEAN DEFAULT true,
      solo_lectura BOOLEAN DEFAULT false,
      fecha_creacion TIMESTAMP DEFAULT NOW(),
      fecha_modificacion TIMESTAMP DEFAULT NOW(),
      usuario_creacion UUID
    )
  `)

  // Crear tabla de auditoría si no existe
  await database.query(`
    CREATE TABLE IF NOT EXISTS auditoria_configuracion (
      id SERIAL PRIMARY KEY,
      clave_configuracion VARCHAR(100) NOT NULL,
      valor_anterior TEXT,
      valor_nuevo TEXT,
      usuario_id UUID,
      fecha_cambio TIMESTAMP DEFAULT NOW()
    )
  `)

  // Insertar configuraciones de test
  await database.query(`
    INSERT INTO configuracion_sistema (clave, valor, descripcion, tipo_dato, categoria, solo_lectura)
    VALUES 
      ('empresa_nombre', 'Ferretería Test', 'Nombre de la empresa', 'string', 'empresa', false),
      ('empresa_rut', '76123456-7', 'RUT de la empresa', 'string', 'empresa', false),
      ('iva_porcentaje', '19', 'Porcentaje de IVA', 'number', 'impuestos', false),
      ('moneda_codigo', 'CLP', 'Código de moneda', 'string', 'general', false),
      ('sistema_version', '1.0.0', 'Versión del sistema', 'string', 'sistema', true),
      ('backup_automatico', 'true', 'Backup automático habilitado', 'boolean', 'sistema', false)
    ON CONFLICT (clave) DO NOTHING
  `)

  // Crear usuario admin de test
  await database.query(`
    INSERT INTO usuarios (rut, nombre, email, rol, password_hash, salt, activo)
    VALUES ('11111111-1', 'Admin Test', 'admin@test.com', 'admin', 'hash', 'salt', true)
    ON CONFLICT (rut) DO NOTHING
  `)

  // Crear usuario vendedor de test
  await database.query(`
    INSERT INTO usuarios (rut, nombre, email, rol, password_hash, salt, activo)
    VALUES ('22222222-2', 'Vendedor Test', 'vendedor@test.com', 'vendedor', 'hash', 'salt', true)
    ON CONFLICT (rut) DO NOTHING
  `)
}

/**
 * Obtener token de autenticación para tests
 */
async function getAuthToken(rol) {
  const rut = rol === 'admin' ? '11111111-1' : '22222222-2'
  
  const response = await server.inject({
    method: 'POST',
    url: '/api/auth/login',
    payload: {
      rut,
      password: 'test123'
    }
  })
  
  return JSON.parse(response.payload).data.token
}

/**
 * Limpiar configuraciones de test
 */
async function cleanupTestConfigurations() {
  await database.query(`
    DELETE FROM configuracion_sistema 
    WHERE clave LIKE 'test_%'
  `)
}

/**
 * Tests de información del sistema
 */
test('Obtener información del sistema', async (t) => {
  const response = await server.inject({
    method: 'GET',
    url: '/api/sistema/info',
    headers: {
      authorization: `Bearer ${authToken}`
    }
  })

  t.equal(response.statusCode, 200)
  
  const result = JSON.parse(response.payload)
  t.ok(result.success)
  t.ok(result.data.version)
  t.ok(result.data.nombre)
  t.ok(result.data.ambiente)
  t.ok(result.data.base_datos)
  t.ok(result.data.configuraciones)
})

test('Health check del sistema', async (t) => {
  const response = await server.inject({
    method: 'GET',
    url: '/api/sistema/health',
    headers: {
      authorization: `Bearer ${authToken}`
    }
  })

  t.equal(response.statusCode, 200)
  
  const result = JSON.parse(response.payload)
  t.equal(result.status, 'ok')
  t.ok(result.timestamp)
  t.ok(result.version)
  t.ok(result.database)
})

test('Obtener estadísticas del sistema', async (t) => {
  const response = await server.inject({
    method: 'GET',
    url: '/api/sistema/estadisticas',
    headers: {
      authorization: `Bearer ${adminToken}`
    }
  })

  t.equal(response.statusCode, 200)
  
  const result = JSON.parse(response.payload)
  t.ok(result.success)
  t.ok(typeof result.data.total_usuarios === 'number')
  t.ok(typeof result.data.total_productos === 'number')
})

test('Error de permisos para estadísticas - vendedor', async (t) => {
  const response = await server.inject({
    method: 'GET',
    url: '/api/sistema/estadisticas',
    headers: {
      authorization: `Bearer ${authToken}`
    }
  })

  t.equal(response.statusCode, 403)
  
  const result = JSON.parse(response.payload)
  t.equal(result.code, 'INSUFFICIENT_PERMISSIONS')
})

/**
 * Tests de configuraciones del sistema
 */
test('Obtener todas las configuraciones - admin', async (t) => {
  const response = await server.inject({
    method: 'GET',
    url: '/api/sistema/configuraciones',
    headers: {
      authorization: `Bearer ${adminToken}`
    }
  })

  t.equal(response.statusCode, 200)
  
  const result = JSON.parse(response.payload)
  t.ok(result.success)
  t.ok(result.data.empresa)
  t.ok(result.data.impuestos)
  t.ok(result.data.general)
})

test('Error de permisos para configuraciones - vendedor', async (t) => {
  const response = await server.inject({
    method: 'GET',
    url: '/api/sistema/configuraciones',
    headers: {
      authorization: `Bearer ${authToken}`
    }
  })

  t.equal(response.statusCode, 403)
  
  const result = JSON.parse(response.payload)
  t.equal(result.code, 'INSUFFICIENT_PERMISSIONS')
})

test('Obtener configuración específica', async (t) => {
  const response = await server.inject({
    method: 'GET',
    url: '/api/sistema/configuraciones/empresa_nombre',
    headers: {
      authorization: `Bearer ${adminToken}`
    }
  })

  t.equal(response.statusCode, 200)
  
  const result = JSON.parse(response.payload)
  t.ok(result.success)
  t.equal(result.data.clave, 'empresa_nombre')
  t.equal(result.data.valor, 'Ferretería Test')
})

test('Error al obtener configuración inexistente', async (t) => {
  const response = await server.inject({
    method: 'GET',
    url: '/api/sistema/configuraciones/config_inexistente',
    headers: {
      authorization: `Bearer ${adminToken}`
    }
  })

  t.equal(response.statusCode, 404)
  
  const result = JSON.parse(response.payload)
  t.equal(result.code, 'CONFIG_NOT_FOUND')
})

test('Crear nueva configuración', async (t) => {
  const configData = {
    clave: 'test_nueva_config',
    valor: 'valor_test',
    descripcion: 'Configuración de prueba',
    tipo_dato: 'string',
    categoria: 'test',
    solo_lectura: false
  }

  const response = await server.inject({
    method: 'POST',
    url: '/api/sistema/configuraciones',
    headers: {
      authorization: `Bearer ${adminToken}`
    },
    payload: configData
  })

  t.equal(response.statusCode, 201)
  
  const result = JSON.parse(response.payload)
  t.ok(result.success)
  t.equal(result.data.clave, 'test_nueva_config')
  t.equal(result.data.valor, 'valor_test')
})

test('Error al crear configuración con clave duplicada', async (t) => {
  const configData = {
    clave: 'empresa_nombre', // Clave que ya existe
    valor: 'nuevo_valor',
    tipo_dato: 'string',
    categoria: 'test'
  }

  const response = await server.inject({
    method: 'POST',
    url: '/api/sistema/configuraciones',
    headers: {
      authorization: `Bearer ${adminToken}`
    },
    payload: configData
  })

  t.equal(response.statusCode, 409)
  
  const result = JSON.parse(response.payload)
  t.equal(result.code, 'CONFIG_ALREADY_EXISTS')
})

test('Actualizar configuración existente', async (t) => {
  const response = await server.inject({
    method: 'PUT',
    url: '/api/sistema/configuraciones/empresa_nombre',
    headers: {
      authorization: `Bearer ${adminToken}`
    },
    payload: {
      valor: 'Ferretería Actualizada'
    }
  })

  t.equal(response.statusCode, 200)
  
  const result = JSON.parse(response.payload)
  t.ok(result.success)
  t.equal(result.data.valor, 'Ferretería Actualizada')
})

test('Error al actualizar configuración de solo lectura', async (t) => {
  const response = await server.inject({
    method: 'PUT',
    url: '/api/sistema/configuraciones/sistema_version',
    headers: {
      authorization: `Bearer ${adminToken}`
    },
    payload: {
      valor: '2.0.0'
    }
  })

  t.equal(response.statusCode, 400)
  
  const result = JSON.parse(response.payload)
  t.equal(result.code, 'CONFIG_READ_ONLY')
})

test('Eliminar configuración', async (t) => {
  // Crear configuración para eliminar
  await server.inject({
    method: 'POST',
    url: '/api/sistema/configuraciones',
    headers: {
      authorization: `Bearer ${adminToken}`
    },
    payload: {
      clave: 'test_eliminar',
      valor: 'valor',
      tipo_dato: 'string',
      categoria: 'test'
    }
  })

  const response = await server.inject({
    method: 'DELETE',
    url: '/api/sistema/configuraciones/test_eliminar',
    headers: {
      authorization: `Bearer ${adminToken}`
    }
  })

  t.equal(response.statusCode, 200)
  
  const result = JSON.parse(response.payload)
  t.ok(result.success)
})

/**
 * Tests de backup y restauración
 */
test('Crear backup de configuraciones', async (t) => {
  const response = await server.inject({
    method: 'POST',
    url: '/api/sistema/backup',
    headers: {
      authorization: `Bearer ${adminToken}`
    }
  })

  t.equal(response.statusCode, 200)
  
  const result = JSON.parse(response.payload)
  t.ok(result.success)
  t.ok(result.data.archivo)
  t.ok(result.data.fecha)
})

test('Restaurar configuraciones desde backup', async (t) => {
  const backupData = {
    fecha: new Date().toISOString(),
    version: '1.0.0',
    configuraciones: {
      test: {
        test_restaurar: {
          valor: 'valor_restaurado',
          descripcion: 'Config restaurada',
          tipo_dato: 'string',
          activa: true,
          solo_lectura: false
        }
      }
    }
  }

  const response = await server.inject({
    method: 'POST',
    url: '/api/sistema/restaurar',
    headers: {
      authorization: `Bearer ${adminToken}`
    },
    payload: backupData
  })

  t.equal(response.statusCode, 200)
  
  const result = JSON.parse(response.payload)
  t.ok(result.success)
  t.ok(result.data.configuracionesRestauradas >= 0)
})

/**
 * Tests de logs del sistema
 */
test('Obtener logs del sistema', async (t) => {
  const response = await server.inject({
    method: 'GET',
    url: '/api/sistema/logs?limit=10',
    headers: {
      authorization: `Bearer ${adminToken}`
    }
  })

  t.equal(response.statusCode, 200)
  
  const result = JSON.parse(response.payload)
  t.ok(result.success)
  t.ok(Array.isArray(result.data))
  t.ok(result.pagination)
})

test('Limpiar logs antiguos', async (t) => {
  const response = await server.inject({
    method: 'POST',
    url: '/api/sistema/logs/limpiar',
    headers: {
      authorization: `Bearer ${adminToken}`
    },
    payload: {
      dias_retencion: 30
    }
  })

  t.equal(response.statusCode, 200)
  
  const result = JSON.parse(response.payload)
  t.ok(result.success)
  t.ok(typeof result.data.logsEliminados === 'number')
})

/**
 * Tests de mantenimiento del sistema
 */
test('Ejecutar mantenimiento del sistema', async (t) => {
  const response = await server.inject({
    method: 'POST',
    url: '/api/sistema/mantenimiento',
    headers: {
      authorization: `Bearer ${adminToken}`
    },
    payload: {
      limpiar_logs: true,
      dias_retencion_logs: 30,
      backup_configuraciones: true
    }
  })

  t.equal(response.statusCode, 200)
  
  const result = JSON.parse(response.payload)
  t.ok(result.success)
  t.ok(Array.isArray(result.data.tareas_ejecutadas))
  t.ok(Array.isArray(result.data.errores))
})

test('Obtener métricas de rendimiento', async (t) => {
  const response = await server.inject({
    method: 'GET',
    url: '/api/sistema/metricas',
    headers: {
      authorization: `Bearer ${adminToken}`
    }
  })

  t.equal(response.statusCode, 200)
  
  const result = JSON.parse(response.payload)
  t.ok(result.success)
  t.ok(result.data.memoria)
  t.ok(result.data.uptime)
  t.ok(result.data.timestamp)
})

/**
 * Tests de validación
 */
test('Error de validación - clave inválida', async (t) => {
  const configData = {
    clave: 'CLAVE-INVALIDA', // No debe tener mayúsculas ni guiones
    valor: 'valor',
    tipo_dato: 'string',
    categoria: 'test'
  }

  const response = await server.inject({
    method: 'POST',
    url: '/api/sistema/configuraciones',
    headers: {
      authorization: `Bearer ${adminToken}`
    },
    payload: configData
  })

  t.equal(response.statusCode, 400)
  
  const result = JSON.parse(response.payload)
  t.equal(result.code, 'VALIDATION_ERROR')
})

test('Error de validación - tipo de dato inválido', async (t) => {
  const configData = {
    clave: 'test_tipo_invalido',
    valor: 'valor',
    tipo_dato: 'tipo_inexistente',
    categoria: 'test'
  }

  const response = await server.inject({
    method: 'POST',
    url: '/api/sistema/configuraciones',
    headers: {
      authorization: `Bearer ${adminToken}`
    },
    payload: configData
  })

  t.equal(response.statusCode, 400)
  
  const result = JSON.parse(response.payload)
  t.equal(result.code, 'VALIDATION_ERROR')
})

/**
 * Tests de permisos
 */
test('Error de permisos - vendedor intentando crear configuración', async (t) => {
  const configData = {
    clave: 'test_sin_permisos',
    valor: 'valor',
    tipo_dato: 'string',
    categoria: 'test'
  }

  const response = await server.inject({
    method: 'POST',
    url: '/api/sistema/configuraciones',
    headers: {
      authorization: `Bearer ${authToken}`
    },
    payload: configData
  })

  t.equal(response.statusCode, 403)
  
  const result = JSON.parse(response.payload)
  t.equal(result.code, 'INSUFFICIENT_PERMISSIONS')
})

test('Error de permisos - vendedor intentando ejecutar mantenimiento', async (t) => {
  const response = await server.inject({
    method: 'POST',
    url: '/api/sistema/mantenimiento',
    headers: {
      authorization: `Bearer ${authToken}`
    },
    payload: {
      limpiar_logs: true
    }
  })

  t.equal(response.statusCode, 403)
  
  const result = JSON.parse(response.payload)
  t.equal(result.code, 'INSUFFICIENT_PERMISSIONS')
})

