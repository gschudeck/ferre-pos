/**
 * Script de Inicialización de Base de Datos - Sistema Ferre-POS
 * 
 * Crea datos iniciales necesarios para el funcionamiento del sistema,
 * incluyendo usuario administrador, sucursal principal y configuraciones básicas.
 */

const database = require('../config/database')
const Usuario = require('../models/Usuario')
const logger = require('./logger')
const bcrypt = require('bcryptjs')
const config = require('../config')

/**
 * Datos iniciales del sistema
 */
const initialData = {
  // Usuario administrador por defecto
  adminUser: {
    rut: '11111111-1',
    nombre: 'Administrador',
    apellido: 'Sistema',
    email: 'admin@ferre-pos.cl',
    rol: 'admin',
    password: 'admin123'
  },

  // Sucursal principal
  sucursalPrincipal: {
    codigo: 'PRINCIPAL',
    nombre: 'Sucursal Principal',
    direccion: 'Dirección Principal',
    telefono: '+56912345678',
    email: 'principal@ferre-pos.cl',
    habilitada: true
  },

  // Terminal principal
  terminalPrincipal: {
    nombre_terminal: 'Terminal Principal',
    ubicacion: 'Caja Principal',
    habilitado: true
  },

  // Categorías básicas
  categorias: [
    { codigo: 'HERR', nombre: 'Herramientas', descripcion: 'Herramientas manuales y eléctricas' },
    { codigo: 'FIJE', nombre: 'Fijaciones', descripcion: 'Tornillos, clavos, pernos' },
    { codigo: 'PINT', nombre: 'Pinturas', descripcion: 'Pinturas y accesorios' },
    { codigo: 'ELEC', nombre: 'Eléctrico', descripcion: 'Material eléctrico' },
    { codigo: 'PLOM', nombre: 'Plomería', descripcion: 'Accesorios de plomería' },
    { codigo: 'JARD', nombre: 'Jardín', descripcion: 'Herramientas y accesorios de jardín' }
  ],

  // Productos de ejemplo
  productos: [
    {
      codigo_interno: 'MART001',
      codigo_barra: '7801234567890',
      descripcion: 'Martillo de Carpintero 16 oz',
      descripcion_corta: 'Martillo 16oz',
      marca: 'Stanley',
      modelo: 'STHT51512',
      precio_unitario: 15990,
      precio_costo: 9990,
      unidad_medida: 'UN',
      peso: 450,
      stock_minimo: 5,
      stock_maximo: 50
    },
    {
      codigo_interno: 'TORN001',
      codigo_barra: '7801234567891',
      descripcion: 'Tornillo Autorroscante 6x1" (100 unidades)',
      descripcion_corta: 'Tornillo 6x1" x100',
      marca: 'Hilti',
      modelo: 'HUS-H 6x25',
      precio_unitario: 2990,
      precio_costo: 1890,
      unidad_medida: 'PQ',
      peso: 200,
      stock_minimo: 10,
      stock_maximo: 100
    },
    {
      codigo_interno: 'PINT001',
      codigo_barra: '7801234567892',
      descripcion: 'Pintura Látex Blanco 1 Galón',
      descripcion_corta: 'Látex Blanco 1Gal',
      marca: 'Sherwin Williams',
      modelo: 'ProClassic',
      precio_unitario: 25990,
      precio_costo: 18990,
      unidad_medida: 'GL',
      peso: 4000,
      stock_minimo: 3,
      stock_maximo: 30
    }
  ]
}

/**
 * Inicializa la base de datos con datos básicos
 */
async function initializeDatabase() {
  try {
    logger.info('Iniciando inicialización de base de datos...')

    // Conectar a la base de datos
    await database.connect()

    // Verificar si ya existe el usuario administrador
    const existingAdmin = await Usuario.findByRut(initialData.adminUser.rut)
    if (existingAdmin) {
      logger.info('Base de datos ya inicializada (usuario admin existe)')
      return
    }

    await database.transaction(async (client) => {
      // 1. Crear sucursal principal
      logger.info('Creando sucursal principal...')
      const sucursalQuery = `
        INSERT INTO sucursales (codigo, nombre, direccion, telefono, email, habilitada)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id
      `
      const sucursalResult = await client.query(sucursalQuery, [
        initialData.sucursalPrincipal.codigo,
        initialData.sucursalPrincipal.nombre,
        initialData.sucursalPrincipal.direccion,
        initialData.sucursalPrincipal.telefono,
        initialData.sucursalPrincipal.email,
        initialData.sucursalPrincipal.habilitada
      ])
      const sucursalId = sucursalResult.rows[0].id

      // 2. Crear terminal principal
      logger.info('Creando terminal principal...')
      const terminalQuery = `
        INSERT INTO terminales (sucursal_id, nombre_terminal, ubicacion, habilitado)
        VALUES ($1, $2, $3, $4)
        RETURNING id
      `
      const terminalResult = await client.query(terminalQuery, [
        sucursalId,
        initialData.terminalPrincipal.nombre_terminal,
        initialData.terminalPrincipal.ubicacion,
        initialData.terminalPrincipal.habilitado
      ])
      const terminalId = terminalResult.rows[0].id

      // 3. Crear usuario administrador
      logger.info('Creando usuario administrador...')
      const salt = await bcrypt.genSalt(config.auth.bcryptRounds)
      const passwordHash = await bcrypt.hash(initialData.adminUser.password, salt)

      const userQuery = `
        INSERT INTO usuarios (rut, nombre, apellido, email, rol, sucursal_id, password_hash, salt, activo)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
        RETURNING id
      `
      const userResult = await client.query(userQuery, [
        initialData.adminUser.rut,
        initialData.adminUser.nombre,
        initialData.adminUser.apellido,
        initialData.adminUser.email,
        initialData.adminUser.rol,
        sucursalId,
        passwordHash,
        salt,
        true
      ])
      const userId = userResult.rows[0].id

      // 4. Crear categorías
      logger.info('Creando categorías básicas...')
      for (const categoria of initialData.categorias) {
        const categoriaQuery = `
          INSERT INTO categorias_productos (codigo, nombre, descripcion, activa)
          VALUES ($1, $2, $3, $4)
          RETURNING id
        `
        await client.query(categoriaQuery, [
          categoria.codigo,
          categoria.nombre,
          categoria.descripcion,
          true
        ])
      }

      // 5. Obtener ID de primera categoría para productos
      const categoriaResult = await client.query(
        'SELECT id FROM categorias_productos WHERE codigo = $1',
        ['HERR']
      )
      const categoriaId = categoriaResult.rows[0].id

      // 6. Crear productos de ejemplo
      logger.info('Creando productos de ejemplo...')
      for (const producto of initialData.productos) {
        const productoQuery = `
          INSERT INTO productos (
            codigo_interno, codigo_barra, descripcion, descripcion_corta,
            categoria_id, marca, modelo, precio_unitario, precio_costo,
            unidad_medida, peso, stock_minimo, stock_maximo, activo,
            usuario_creacion
          ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
          RETURNING id
        `
        const productoResult = await client.query(productoQuery, [
          producto.codigo_interno,
          producto.codigo_barra,
          producto.descripcion,
          producto.descripcion_corta,
          categoriaId,
          producto.marca,
          producto.modelo,
          producto.precio_unitario,
          producto.precio_costo,
          producto.unidad_medida,
          producto.peso,
          producto.stock_minimo,
          producto.stock_maximo,
          true,
          userId
        ])
        const productoId = productoResult.rows[0].id

        // Crear stock inicial para el producto
        const stockQuery = `
          INSERT INTO stock_central (producto_id, sucursal_id, cantidad, costo_promedio)
          VALUES ($1, $2, $3, $4)
        `
        await client.query(stockQuery, [
          productoId,
          sucursalId,
          producto.stock_maximo,
          producto.precio_costo
        ])
      }

      // 7. Crear configuraciones del sistema
      logger.info('Creando configuraciones del sistema...')
      const configuraciones = [
        { clave: 'empresa_nombre', valor: 'Ferretería Demo', descripcion: 'Nombre de la empresa' },
        { clave: 'empresa_rut', valor: '76123456-7', descripcion: 'RUT de la empresa' },
        { clave: 'empresa_direccion', valor: 'Av. Principal 123', descripcion: 'Dirección de la empresa' },
        { clave: 'empresa_telefono', valor: '+56912345678', descripcion: 'Teléfono de la empresa' },
        { clave: 'empresa_email', valor: 'contacto@ferreteria.cl', descripcion: 'Email de la empresa' },
        { clave: 'iva_porcentaje', valor: '19', descripcion: 'Porcentaje de IVA' },
        { clave: 'moneda_codigo', valor: 'CLP', descripcion: 'Código de moneda' },
        { clave: 'moneda_simbolo', valor: '$', descripcion: 'Símbolo de moneda' }
      ]

      for (const config of configuraciones) {
        const configQuery = `
          INSERT INTO configuracion_sistema (clave, valor, descripcion, activa)
          VALUES ($1, $2, $3, $4)
        `
        await client.query(configQuery, [
          config.clave,
          config.valor,
          config.descripcion,
          true
        ])
      }

      logger.info('Base de datos inicializada exitosamente')
      logger.info(`Usuario administrador creado:`)
      logger.info(`  RUT: ${initialData.adminUser.rut}`)
      logger.info(`  Password: ${initialData.adminUser.password}`)
      logger.info(`  Sucursal: ${initialData.sucursalPrincipal.nombre}`)
    })

  } catch (error) {
    logger.error('Error al inicializar base de datos:', error)
    throw error
  }
}

/**
 * Verifica si la base de datos está inicializada
 */
async function isDatabaseInitialized() {
  try {
    const result = await database.query(
      'SELECT COUNT(*) as count FROM usuarios WHERE rol = $1',
      ['admin']
    )
    return parseInt(result.rows[0].count) > 0
  } catch (error) {
    logger.error('Error al verificar inicialización de base de datos:', error)
    return false
  }
}

/**
 * Resetea la base de datos (solo para desarrollo)
 */
async function resetDatabase() {
  if (process.env.NODE_ENV === 'production') {
    throw new Error('No se puede resetear la base de datos en producción')
  }

  try {
    logger.warn('Reseteando base de datos...')

    await database.transaction(async (client) => {
      // Eliminar datos en orden correcto para evitar violaciones de FK
      const tables = [
        'movimientos_fidelizacion',
        'fidelizacion_clientes',
        'medios_pago_venta',
        'detalle_ventas',
        'ventas',
        'movimientos_stock',
        'stock_central',
        'codigos_barra_adicionales',
        'productos',
        'categorias_productos',
        'usuarios',
        'terminales',
        'sucursales',
        'configuracion_sistema'
      ]

      for (const table of tables) {
        await client.query(`DELETE FROM ${table}`)
        logger.debug(`Tabla ${table} limpiada`)
      }
    })

    logger.info('Base de datos reseteada exitosamente')
  } catch (error) {
    logger.error('Error al resetear base de datos:', error)
    throw error
  }
}

// Ejecutar inicialización si el script es llamado directamente
if (require.main === module) {
  const command = process.argv[2]

  switch (command) {
    case 'init':
      initializeDatabase()
        .then(() => {
          logger.info('Inicialización completada')
          process.exit(0)
        })
        .catch((error) => {
          logger.error('Error en inicialización:', error)
          process.exit(1)
        })
      break

    case 'reset':
      resetDatabase()
        .then(() => initializeDatabase())
        .then(() => {
          logger.info('Reset e inicialización completados')
          process.exit(0)
        })
        .catch((error) => {
          logger.error('Error en reset:', error)
          process.exit(1)
        })
      break

    case 'check':
      isDatabaseInitialized()
        .then((initialized) => {
          logger.info(`Base de datos inicializada: ${initialized}`)
          process.exit(0)
        })
        .catch((error) => {
          logger.error('Error al verificar:', error)
          process.exit(1)
        })
      break

    default:
      console.log('Uso: node dbInit.js [init|reset|check]')
      console.log('  init  - Inicializa la base de datos con datos básicos')
      console.log('  reset - Resetea e inicializa la base de datos')
      console.log('  check - Verifica si la base de datos está inicializada')
      process.exit(1)
  }
}

module.exports = {
  initializeDatabase,
  isDatabaseInitialized,
  resetDatabase,
  initialData
}

