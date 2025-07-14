/**
 * Rutas de Productos - Sistema Ferre-POS
 * 
 * Define todos los endpoints relacionados con el catálogo de productos,
 * incluyendo CRUD, búsquedas y gestión de códigos de barras.
 */

const ProductoController = require('../controllers/ProductoController')

async function productosRoutes(fastify, options) {
  // Esquemas de validación para Swagger
  const productoSchema = {
    type: 'object',
    properties: {
      id: { type: 'string' },
      codigo_interno: { type: 'string' },
      codigo_barra: { type: 'string' },
      descripcion: { type: 'string' },
      descripcion_corta: { type: 'string' },
      categoria_id: { type: 'string' },
      marca: { type: 'string' },
      modelo: { type: 'string' },
      precio_unitario: { type: 'number' },
      precio_costo: { type: 'number' },
      unidad_medida: { type: 'string' },
      peso: { type: 'number' },
      stock_minimo: { type: 'number' },
      stock_maximo: { type: 'number' },
      activo: { type: 'boolean' },
      fecha_creacion: { type: 'string', format: 'date-time' },
      fecha_modificacion: { type: 'string', format: 'date-time' }
    }
  }

  const createProductoSchema = {
    description: 'Crea un nuevo producto',
    tags: ['Productos'],
    security: [{ Bearer: [] }],
    body: {
      type: 'object',
      required: ['codigo_interno', 'codigo_barra', 'descripcion', 'precio_unitario', 'unidad_medida'],
      properties: {
        codigo_interno: {
          type: 'string',
          description: 'Código interno del producto',
          maxLength: 50
        },
        codigo_barra: {
          type: 'string',
          description: 'Código de barras del producto',
          maxLength: 50
        },
        descripcion: {
          type: 'string',
          description: 'Descripción del producto',
          maxLength: 500
        },
        descripcion_corta: {
          type: 'string',
          description: 'Descripción corta del producto',
          maxLength: 100
        },
        categoria_id: {
          type: 'string',
          description: 'ID de la categoría'
        },
        marca: {
          type: 'string',
          description: 'Marca del producto',
          maxLength: 100
        },
        modelo: {
          type: 'string',
          description: 'Modelo del producto',
          maxLength: 100
        },
        precio_unitario: {
          type: 'number',
          description: 'Precio unitario del producto',
          minimum: 0.01
        },
        precio_costo: {
          type: 'number',
          description: 'Precio de costo del producto',
          minimum: 0
        },
        unidad_medida: {
          type: 'string',
          description: 'Unidad de medida',
          maxLength: 10
        },
        peso: {
          type: 'number',
          description: 'Peso del producto en gramos',
          minimum: 0
        },
        stock_minimo: {
          type: 'number',
          description: 'Stock mínimo del producto',
          minimum: 0
        },
        stock_maximo: {
          type: 'number',
          description: 'Stock máximo del producto',
          minimum: 0
        }
      }
    },
    response: {
      201: {
        type: 'object',
        properties: {
          success: { type: 'boolean' },
          message: { type: 'string' },
          data: productoSchema
        }
      }
    }
  }

  const getProductosSchema = {
    description: 'Obtiene lista de productos con filtros y paginación',
    tags: ['Productos'],
    security: [{ Bearer: [] }],
    querystring: {
      type: 'object',
      properties: {
        page: {
          type: 'integer',
          description: 'Número de página',
          minimum: 1,
          default: 1
        },
        limit: {
          type: 'integer',
          description: 'Cantidad de productos por página',
          minimum: 1,
          maximum: 100,
          default: 20
        },
        q: {
          type: 'string',
          description: 'Término de búsqueda general'
        },
        categoria: {
          type: 'string',
          description: 'Filtrar por categoría'
        },
        marca: {
          type: 'string',
          description: 'Filtrar por marca'
        },
        precioMin: {
          type: 'number',
          description: 'Precio mínimo'
        },
        precioMax: {
          type: 'number',
          description: 'Precio máximo'
        },
        conStock: {
          type: 'boolean',
          description: 'Solo productos con stock disponible'
        },
        orderBy: {
          type: 'string',
          description: 'Campo para ordenar',
          default: 'descripcion'
        },
        orderDirection: {
          type: 'string',
          enum: ['ASC', 'DESC'],
          description: 'Dirección del ordenamiento',
          default: 'ASC'
        }
      }
    },
    response: {
      200: {
        type: 'object',
        properties: {
          success: { type: 'boolean' },
          data: {
            type: 'array',
            items: productoSchema
          },
          pagination: {
            type: 'object',
            properties: {
              page: { type: 'integer' },
              limit: { type: 'integer' },
              total: { type: 'integer' }
            }
          }
        }
      }
    }
  }

  // Todas las rutas requieren autenticación
  fastify.addHook('preHandler', fastify.authenticate)

  /**
   * GET /api/productos
   * Obtiene lista de productos con filtros y paginación
   */
  fastify.get('/', {
    schema: getProductosSchema
  }, ProductoController.getProductos)

  /**
   * GET /api/productos/stats
   * Obtiene estadísticas del catálogo de productos
   */
  fastify.get('/stats', {
    schema: {
      description: 'Obtiene estadísticas del catálogo de productos',
      tags: ['Productos'],
      security: [{ Bearer: [] }],
      response: {
        200: {
          type: 'object',
          properties: {
            success: { type: 'boolean' },
            data: { type: 'object' }
          }
        }
      }
    }
  }, ProductoController.getProductStats)

  /**
   * GET /api/productos/stock-bajo
   * Obtiene productos con stock bajo
   */
  fastify.get('/stock-bajo', {
    schema: {
      description: 'Obtiene productos con stock bajo',
      tags: ['Productos'],
      security: [{ Bearer: [] }],
      querystring: {
        type: 'object',
        properties: {
          sucursalId: {
            type: 'string',
            description: 'ID de sucursal (solo para admin)'
          }
        }
      },
      response: {
        200: {
          type: 'object',
          properties: {
            success: { type: 'boolean' },
            data: {
              type: 'array',
              items: productoSchema
            }
          }
        }
      }
    }
  }, ProductoController.getProductosStockBajo)

  /**
   * GET /api/productos/mas-vendidos
   * Obtiene productos más vendidos
   */
  fastify.get('/mas-vendidos', {
    schema: {
      description: 'Obtiene productos más vendidos',
      tags: ['Productos'],
      security: [{ Bearer: [] }],
      querystring: {
        type: 'object',
        properties: {
          fechaInicio: {
            type: 'string',
            format: 'date',
            description: 'Fecha de inicio del período'
          },
          fechaFin: {
            type: 'string',
            format: 'date',
            description: 'Fecha de fin del período'
          },
          limit: {
            type: 'integer',
            description: 'Cantidad de productos a retornar',
            minimum: 1,
            maximum: 100,
            default: 20
          },
          sucursalId: {
            type: 'string',
            description: 'ID de sucursal (solo para admin)'
          }
        }
      },
      response: {
        200: {
          type: 'object',
          properties: {
            success: { type: 'boolean' },
            data: {
              type: 'array',
              items: productoSchema
            }
          }
        }
      }
    }
  }, ProductoController.getProductosMasVendidos)

  /**
   * POST /api/productos
   * Crea un nuevo producto
   */
  fastify.post('/', {
    schema: createProductoSchema
  }, ProductoController.createProducto)

  /**
   * POST /api/productos/validate-stock
   * Valida stock para venta
   */
  fastify.post('/validate-stock', {
    schema: {
      description: 'Valida stock disponible para una lista de productos',
      tags: ['Productos'],
      security: [{ Bearer: [] }],
      body: {
        type: 'object',
        required: ['productos'],
        properties: {
          productos: {
            type: 'array',
            items: {
              type: 'object',
              required: ['producto_id', 'cantidad'],
              properties: {
                producto_id: { type: 'string' },
                cantidad: { type: 'number', minimum: 1 }
              }
            }
          }
        }
      },
      response: {
        200: {
          type: 'object',
          properties: {
            success: { type: 'boolean' },
            data: {
              type: 'object',
              properties: {
                valido: { type: 'boolean' },
                validaciones: {
                  type: 'array',
                  items: {
                    type: 'object',
                    properties: {
                      producto_id: { type: 'string' },
                      cantidad: { type: 'number' },
                      valido: { type: 'boolean' },
                      error: { type: 'string' }
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
  }, ProductoController.validateStock)

  /**
   * GET /api/productos/codigo/:codigo
   * Busca un producto por código (interno o de barras)
   */
  fastify.get('/codigo/:codigo', {
    schema: {
      description: 'Busca un producto por código interno o de barras',
      tags: ['Productos'],
      security: [{ Bearer: [] }],
      params: {
        type: 'object',
        required: ['codigo'],
        properties: {
          codigo: {
            type: 'string',
            description: 'Código interno o código de barras del producto'
          }
        }
      },
      response: {
        200: {
          type: 'object',
          properties: {
            success: { type: 'boolean' },
            data: productoSchema
          }
        },
        404: {
          type: 'object',
          properties: {
            code: { type: 'string' },
            error: { type: 'string' },
            message: { type: 'string' }
          }
        }
      }
    }
  }, ProductoController.getProductoPorCodigo)

  /**
   * GET /api/productos/:id
   * Obtiene un producto por ID
   */
  fastify.get('/:id', {
    schema: {
      description: 'Obtiene un producto por ID',
      tags: ['Productos'],
      security: [{ Bearer: [] }],
      params: {
        type: 'object',
        required: ['id'],
        properties: {
          id: {
            type: 'string',
            description: 'ID del producto'
          }
        }
      },
      response: {
        200: {
          type: 'object',
          properties: {
            success: { type: 'boolean' },
            data: productoSchema
          }
        },
        404: {
          type: 'object',
          properties: {
            code: { type: 'string' },
            error: { type: 'string' },
            message: { type: 'string' }
          }
        }
      }
    }
  }, ProductoController.getProducto)

  /**
   * GET /api/productos/:id/relacionados
   * Obtiene productos relacionados
   */
  fastify.get('/:id/relacionados', {
    schema: {
      description: 'Obtiene productos relacionados o similares',
      tags: ['Productos'],
      security: [{ Bearer: [] }],
      params: {
        type: 'object',
        required: ['id'],
        properties: {
          id: {
            type: 'string',
            description: 'ID del producto'
          }
        }
      },
      querystring: {
        type: 'object',
        properties: {
          limit: {
            type: 'integer',
            description: 'Cantidad de productos relacionados',
            minimum: 1,
            maximum: 20,
            default: 5
          }
        }
      },
      response: {
        200: {
          type: 'object',
          properties: {
            success: { type: 'boolean' },
            data: {
              type: 'array',
              items: productoSchema
            }
          }
        }
      }
    }
  }, ProductoController.getProductosRelacionados)

  /**
   * PUT /api/productos/:id
   * Actualiza un producto existente
   */
  fastify.put('/:id', {
    schema: {
      description: 'Actualiza un producto existente',
      tags: ['Productos'],
      security: [{ Bearer: [] }],
      params: {
        type: 'object',
        required: ['id'],
        properties: {
          id: {
            type: 'string',
            description: 'ID del producto'
          }
        }
      },
      body: {
        type: 'object',
        properties: {
          descripcion: { type: 'string', maxLength: 500 },
          descripcion_corta: { type: 'string', maxLength: 100 },
          categoria_id: { type: 'string' },
          marca: { type: 'string', maxLength: 100 },
          modelo: { type: 'string', maxLength: 100 },
          precio_unitario: { type: 'number', minimum: 0.01 },
          precio_costo: { type: 'number', minimum: 0 },
          unidad_medida: { type: 'string', maxLength: 10 },
          peso: { type: 'number', minimum: 0 },
          stock_minimo: { type: 'number', minimum: 0 },
          stock_maximo: { type: 'number', minimum: 0 }
        }
      },
      response: {
        200: {
          type: 'object',
          properties: {
            success: { type: 'boolean' },
            message: { type: 'string' },
            data: productoSchema
          }
        }
      }
    }
  }, ProductoController.updateProducto)

  /**
   * PATCH /api/productos/:id/precio
   * Actualiza el precio de un producto
   */
  fastify.patch('/:id/precio', {
    schema: {
      description: 'Actualiza el precio de un producto',
      tags: ['Productos'],
      security: [{ Bearer: [] }],
      params: {
        type: 'object',
        required: ['id'],
        properties: {
          id: {
            type: 'string',
            description: 'ID del producto'
          }
        }
      },
      body: {
        type: 'object',
        required: ['precio'],
        properties: {
          precio: {
            type: 'number',
            description: 'Nuevo precio del producto',
            minimum: 0.01
          }
        }
      },
      response: {
        200: {
          type: 'object',
          properties: {
            success: { type: 'boolean' },
            message: { type: 'string' }
          }
        }
      }
    }
  }, ProductoController.updatePrecio)

  /**
   * POST /api/productos/:id/codigos-barra
   * Agrega un código de barras adicional
   */
  fastify.post('/:id/codigos-barra', {
    schema: {
      description: 'Agrega un código de barras adicional al producto',
      tags: ['Productos'],
      security: [{ Bearer: [] }],
      params: {
        type: 'object',
        required: ['id'],
        properties: {
          id: {
            type: 'string',
            description: 'ID del producto'
          }
        }
      },
      body: {
        type: 'object',
        required: ['codigoBarra'],
        properties: {
          codigoBarra: {
            type: 'string',
            description: 'Código de barras adicional'
          },
          descripcion: {
            type: 'string',
            description: 'Descripción del código de barras'
          }
        }
      },
      response: {
        201: {
          type: 'object',
          properties: {
            success: { type: 'boolean' },
            message: { type: 'string' },
            data: { type: 'object' }
          }
        }
      }
    }
  }, ProductoController.addCodigoBarra)

  /**
   * DELETE /api/productos/:id
   * Elimina un producto (eliminación lógica)
   */
  fastify.delete('/:id', {
    schema: {
      description: 'Elimina un producto (eliminación lógica)',
      tags: ['Productos'],
      security: [{ Bearer: [] }],
      params: {
        type: 'object',
        required: ['id'],
        properties: {
          id: {
            type: 'string',
            description: 'ID del producto'
          }
        }
      },
      response: {
        200: {
          type: 'object',
          properties: {
            success: { type: 'boolean' },
            message: { type: 'string' }
          }
        }
      }
    }
  }, ProductoController.deleteProducto)
}

module.exports = productosRoutes

