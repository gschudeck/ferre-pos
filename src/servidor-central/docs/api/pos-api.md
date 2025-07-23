# API POS - Documentación Completa

## Introducción

El API POS (Point of Sale) es el núcleo del sistema de punto de venta, diseñado para manejar todas las operaciones críticas en tiempo real. Este API está optimizado para alta concurrencia y baja latencia, soportando múltiples terminales simultáneas y operaciones de venta continuas.

## Configuración Base

### URL Base
```
http://localhost:8080/api/pos
```

### Autenticación
Todas las rutas (excepto las públicas) requieren autenticación JWT:

```http
Authorization: Bearer <jwt_token>
```

### Headers Requeridos
```http
Content-Type: application/json
X-Terminal-ID: <terminal_id>  # Para operaciones de venta
X-Sucursal-ID: <sucursal_id>  # Para operaciones específicas de sucursal
```

## Autenticación y Autorización

### POST /auth/login
Autentica un usuario en el sistema.

#### Request
```json
{
  "email": "cajero@ferreteria.com",
  "password": "password123",
  "terminal_id": "TERM-001",
  "sucursal_id": "uuid-sucursal"
}
```

#### Response (200 OK)
```json
{
  "success": true,
  "data": {
    "user": {
      "id": "uuid-usuario",
      "email": "cajero@ferreteria.com",
      "nombre": "Juan Pérez",
      "rol": "cajero",
      "sucursal_id": "uuid-sucursal",
      "permisos": ["ventas.crear", "productos.leer", "clientes.leer"]
    },
    "tokens": {
      "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
      "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
      "expires_in": 86400
    },
    "terminal": {
      "id": "TERM-001",
      "nombre": "Terminal Principal",
      "estado": "activo"
    }
  }
}
```

#### Errores Comunes
```json
// 401 Unauthorized - Credenciales inválidas
{
  "error": "Credenciales inválidas",
  "code": "AUTH_INVALID_CREDENTIALS",
  "message": "Email o contraseña incorrectos"
}

// 423 Locked - Usuario bloqueado
{
  "error": "Usuario bloqueado",
  "code": "AUTH_USER_LOCKED",
  "message": "Usuario bloqueado por múltiples intentos fallidos"
}
```

### POST /auth/refresh
Renueva el token de acceso usando el refresh token.

#### Request
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

#### Response (200 OK)
```json
{
  "success": true,
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 86400
  }
}
```

## Gestión de Productos

### GET /productos
Obtiene lista de productos con paginación y filtros.

#### Query Parameters
- `page` (int): Número de página (default: 1)
- `limit` (int): Elementos por página (default: 50, max: 1000)
- `search` (string): Búsqueda por nombre o código
- `categoria_id` (uuid): Filtrar por categoría
- `activo` (bool): Filtrar por estado activo
- `con_stock` (bool): Solo productos con stock disponible
- `sucursal_id` (uuid): Filtrar por sucursal específica

#### Request
```http
GET /api/pos/productos?page=1&limit=20&search=martillo&con_stock=true
Authorization: Bearer <token>
```

#### Response (200 OK)
```json
{
  "success": true,
  "data": {
    "productos": [
      {
        "id": "uuid-producto",
        "codigo": "MART-001",
        "nombre": "Martillo de Carpintero 16oz",
        "descripcion": "Martillo profesional con mango de madera",
        "categoria": {
          "id": "uuid-categoria",
          "nombre": "Herramientas",
          "codigo": "HERR"
        },
        "precio_venta": 25000,
        "precio_costo": 15000,
        "iva_porcentaje": 19,
        "activo": true,
        "imagen_url": "https://storage.ferreteria.com/productos/mart-001.jpg",
        "codigos_barra": [
          {
            "codigo": "7891234567890",
            "tipo": "EAN13",
            "principal": true
          }
        ],
        "especificaciones": {
          "peso": "450g",
          "material": "Acero forjado",
          "garantia": "1 año"
        },
        "stock": {
          "cantidad_disponible": 15,
          "cantidad_reservada": 2,
          "stock_minimo": 5,
          "stock_maximo": 50
        },
        "etiquetas": {
          "plantilla_id": "uuid-plantilla",
          "configuracion": {
            "mostrar_precio": true,
            "mostrar_codigo": true,
            "formato": "pequena"
          }
        },
        "created_at": "2024-01-15T10:30:00Z",
        "updated_at": "2024-01-20T14:45:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 156,
      "total_pages": 8,
      "has_next": true,
      "has_prev": false
    }
  }
}
```

### GET /productos/:id
Obtiene un producto específico por ID.

#### Response (200 OK)
```json
{
  "success": true,
  "data": {
    "producto": {
      "id": "uuid-producto",
      "codigo": "MART-001",
      "nombre": "Martillo de Carpintero 16oz",
      // ... resto de campos como en la lista
      "historial_precios": [
        {
          "precio": 25000,
          "fecha_inicio": "2024-01-01T00:00:00Z",
          "fecha_fin": null,
          "usuario_id": "uuid-usuario"
        }
      ],
      "movimientos_stock": [
        {
          "tipo": "entrada",
          "cantidad": 20,
          "fecha": "2024-01-15T10:00:00Z",
          "referencia": "COMPRA-001",
          "usuario_id": "uuid-usuario"
        }
      ]
    }
  }
}
```

### GET /productos/buscar
Búsqueda avanzada de productos.

#### Query Parameters
- `q` (string): Término de búsqueda
- `tipo` (string): Tipo de búsqueda (nombre, codigo, barcode, descripcion)
- `categoria_ids` (array): IDs de categorías
- `precio_min` (float): Precio mínimo
- `precio_max` (float): Precio máximo
- `con_stock` (bool): Solo con stock
- `limit` (int): Límite de resultados

#### Request
```http
GET /api/pos/productos/buscar?q=martillo&tipo=nombre&con_stock=true&limit=10
```

#### Response (200 OK)
```json
{
  "success": true,
  "data": {
    "productos": [
      // Array de productos que coinciden con la búsqueda
    ],
    "total_encontrados": 5,
    "tiempo_busqueda_ms": 45
  }
}
```

### GET /productos/barcode/:codigo
Obtiene un producto por código de barras.

#### Request
```http
GET /api/pos/productos/barcode/7891234567890
```

#### Response (200 OK)
```json
{
  "success": true,
  "data": {
    "producto": {
      // Datos completos del producto
    },
    "codigo_barra": {
      "codigo": "7891234567890",
      "tipo": "EAN13",
      "principal": true
    }
  }
}
```

## Gestión de Stock

### GET /stock
Obtiene información de stock con filtros.

#### Query Parameters
- `producto_id` (uuid): ID del producto específico
- `sucursal_id` (uuid): ID de la sucursal
- `stock_bajo` (bool): Solo productos con stock bajo
- `sin_stock` (bool): Solo productos sin stock
- `page` (int): Página
- `limit` (int): Límite por página

#### Response (200 OK)
```json
{
  "success": true,
  "data": {
    "stock_items": [
      {
        "producto_id": "uuid-producto",
        "producto": {
          "codigo": "MART-001",
          "nombre": "Martillo de Carpintero 16oz"
        },
        "sucursal_id": "uuid-sucursal",
        "cantidad_disponible": 15,
        "cantidad_reservada": 2,
        "cantidad_total": 17,
        "stock_minimo": 5,
        "stock_maximo": 50,
        "costo_promedio": 15000,
        "valor_total": 255000,
        "estado": "normal", // normal, bajo, agotado
        "ultima_entrada": "2024-01-15T10:00:00Z",
        "ultima_salida": "2024-01-20T15:30:00Z"
      }
    ],
    "resumen": {
      "total_productos": 156,
      "productos_con_stock": 142,
      "productos_stock_bajo": 8,
      "productos_agotados": 6,
      "valor_total_inventario": 15750000
    }
  }
}
```

### GET /stock/:producto_id
Obtiene stock específico de un producto.

#### Response (200 OK)
```json
{
  "success": true,
  "data": {
    "stock": {
      "producto_id": "uuid-producto",
      "sucursal_id": "uuid-sucursal",
      "cantidad_disponible": 15,
      "cantidad_reservada": 2,
      "reservas_activas": [
        {
          "id": "uuid-reserva",
          "cantidad": 2,
          "usuario_id": "uuid-usuario",
          "terminal_id": "TERM-001",
          "expira_en": "2024-01-20T16:00:00Z",
          "motivo": "venta_en_proceso"
        }
      ],
      "movimientos_recientes": [
        {
          "tipo": "salida",
          "cantidad": 1,
          "fecha": "2024-01-20T15:30:00Z",
          "referencia": "VENTA-001",
          "usuario_id": "uuid-usuario"
        }
      ]
    }
  }
}
```

### POST /stock/reservar
Reserva stock para una venta en proceso.

#### Request
```json
{
  "items": [
    {
      "producto_id": "uuid-producto",
      "cantidad": 2,
      "precio_unitario": 25000
    }
  ],
  "terminal_id": "TERM-001",
  "duracion_minutos": 30,
  "motivo": "venta_en_proceso"
}
```

#### Response (200 OK)
```json
{
  "success": true,
  "data": {
    "reserva": {
      "id": "uuid-reserva",
      "items": [
        {
          "producto_id": "uuid-producto",
          "cantidad": 2,
          "cantidad_reservada": 2,
          "precio_unitario": 25000
        }
      ],
      "terminal_id": "TERM-001",
      "usuario_id": "uuid-usuario",
      "expira_en": "2024-01-20T16:00:00Z",
      "estado": "activa"
    }
  }
}
```

### POST /stock/liberar
Libera una reserva de stock.

#### Request
```json
{
  "reserva_id": "uuid-reserva",
  "motivo": "venta_cancelada"
}
```

#### Response (200 OK)
```json
{
  "success": true,
  "data": {
    "reserva": {
      "id": "uuid-reserva",
      "estado": "liberada",
      "liberada_en": "2024-01-20T15:45:00Z",
      "items_liberados": [
        {
          "producto_id": "uuid-producto",
          "cantidad": 2
        }
      ]
    }
  }
}
```

## Gestión de Ventas

### POST /ventas
Crea una nueva venta.

#### Request
```json
{
  "cliente_id": "uuid-cliente", // Opcional
  "items": [
    {
      "producto_id": "uuid-producto",
      "cantidad": 2,
      "precio_unitario": 25000,
      "descuento_porcentaje": 0,
      "descuento_valor": 0
    },
    {
      "producto_id": "uuid-producto-2",
      "cantidad": 1,
      "precio_unitario": 15000
    }
  ],
  "descuento_general": {
    "tipo": "porcentaje", // porcentaje, valor
    "valor": 5
  },
  "medios_pago": [
    {
      "tipo": "efectivo",
      "valor": 60000
    },
    {
      "tipo": "tarjeta_credito",
      "valor": 5000,
      "referencia": "AUTH-123456",
      "ultimos_digitos": "1234"
    }
  ],
  "observaciones": "Cliente frecuente",
  "requiere_factura": true,
  "datos_facturacion": {
    "nit": "123456789-1",
    "razon_social": "Empresa XYZ SAS",
    "direccion": "Calle 123 #45-67",
    "telefono": "3001234567",
    "email": "facturacion@empresa.com"
  }
}
```

#### Response (201 Created)
```json
{
  "success": true,
  "data": {
    "venta": {
      "id": "uuid-venta",
      "numero": "VENTA-000001",
      "fecha": "2024-01-20T15:30:00Z",
      "cliente": {
        "id": "uuid-cliente",
        "nombre": "Juan Pérez",
        "documento": "12345678",
        "telefono": "3001234567"
      },
      "items": [
        {
          "id": "uuid-item",
          "producto": {
            "id": "uuid-producto",
            "codigo": "MART-001",
            "nombre": "Martillo de Carpintero 16oz"
          },
          "cantidad": 2,
          "precio_unitario": 25000,
          "descuento_valor": 0,
          "subtotal": 50000,
          "iva_valor": 9500,
          "total": 59500
        }
      ],
      "subtotal": 65000,
      "descuento_general": 3250,
      "subtotal_con_descuento": 61750,
      "iva_total": 11732,
      "total": 73482,
      "medios_pago": [
        {
          "tipo": "efectivo",
          "valor": 60000
        },
        {
          "tipo": "tarjeta_credito",
          "valor": 13482,
          "referencia": "AUTH-123456"
        }
      ],
      "estado": "completada",
      "terminal": {
        "id": "TERM-001",
        "nombre": "Terminal Principal"
      },
      "usuario": {
        "id": "uuid-usuario",
        "nombre": "María González"
      },
      "sucursal": {
        "id": "uuid-sucursal",
        "nombre": "Sucursal Centro"
      },
      "factura": {
        "numero": "FACT-000001",
        "estado": "generada",
        "url_pdf": "https://storage.ferreteria.com/facturas/FACT-000001.pdf"
      },
      "puntos_otorgados": 73, // Si tiene programa de fidelización
      "created_at": "2024-01-20T15:30:00Z"
    }
  }
}
```

### GET /ventas/:id
Obtiene una venta específica.

#### Response (200 OK)
```json
{
  "success": true,
  "data": {
    "venta": {
      // Datos completos de la venta como en POST
      "historial_estados": [
        {
          "estado": "en_proceso",
          "fecha": "2024-01-20T15:25:00Z",
          "usuario_id": "uuid-usuario"
        },
        {
          "estado": "completada",
          "fecha": "2024-01-20T15:30:00Z",
          "usuario_id": "uuid-usuario"
        }
      ],
      "movimientos_stock": [
        {
          "producto_id": "uuid-producto",
          "cantidad": -2,
          "tipo": "venta",
          "fecha": "2024-01-20T15:30:00Z"
        }
      ]
    }
  }
}
```

### POST /ventas/:id/anular
Anula una venta existente.

#### Request
```json
{
  "motivo": "Error en el precio",
  "observaciones": "Cliente solicitó anulación por precio incorrecto",
  "devolver_stock": true,
  "generar_nota_credito": true
}
```

#### Response (200 OK)
```json
{
  "success": true,
  "data": {
    "venta": {
      "id": "uuid-venta",
      "estado": "anulada",
      "anulada_en": "2024-01-20T16:00:00Z",
      "motivo_anulacion": "Error en el precio",
      "usuario_anulacion": "uuid-usuario",
      "nota_credito": {
        "numero": "NC-000001",
        "valor": 73482,
        "estado": "generada"
      },
      "stock_devuelto": [
        {
          "producto_id": "uuid-producto",
          "cantidad": 2
        }
      ]
    }
  }
}
```

### GET /ventas
Lista ventas con filtros y paginación.

#### Query Parameters
- `fecha_inicio` (date): Fecha inicio (YYYY-MM-DD)
- `fecha_fin` (date): Fecha fin (YYYY-MM-DD)
- `cliente_id` (uuid): ID del cliente
- `usuario_id` (uuid): ID del usuario/cajero
- `terminal_id` (string): ID del terminal
- `estado` (string): Estado de la venta
- `page` (int): Página
- `limit` (int): Límite por página

#### Response (200 OK)
```json
{
  "success": true,
  "data": {
    "ventas": [
      // Array de ventas
    ],
    "resumen": {
      "total_ventas": 45,
      "valor_total": 2150000,
      "promedio_venta": 47777,
      "ventas_por_estado": {
        "completadas": 42,
        "anuladas": 3
      }
    },
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 45,
      "total_pages": 3
    }
  }
}
```

## Gestión de Clientes

### GET /clientes
Lista clientes con filtros.

#### Query Parameters
- `search` (string): Búsqueda por nombre, documento o teléfono
- `activo` (bool): Filtrar por estado activo
- `con_fidelizacion` (bool): Solo clientes con programa de fidelización
- `page` (int): Página
- `limit` (int): Límite por página

#### Response (200 OK)
```json
{
  "success": true,
  "data": {
    "clientes": [
      {
        "id": "uuid-cliente",
        "tipo_documento": "cedula",
        "numero_documento": "12345678",
        "nombre": "Juan Pérez",
        "apellido": "González",
        "email": "juan.perez@email.com",
        "telefono": "3001234567",
        "direccion": "Calle 123 #45-67",
        "fecha_nacimiento": "1985-05-15",
        "activo": true,
        "fidelizacion": {
          "numero_tarjeta": "FIDE-001234",
          "puntos_disponibles": 1250,
          "puntos_acumulados": 5680,
          "nivel": "oro",
          "fecha_afiliacion": "2023-06-01T00:00:00Z"
        },
        "estadisticas": {
          "total_compras": 15,
          "valor_total_compras": 750000,
          "promedio_compra": 50000,
          "ultima_compra": "2024-01-15T14:30:00Z"
        },
        "created_at": "2023-06-01T10:00:00Z",
        "updated_at": "2024-01-20T15:30:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 156,
      "total_pages": 8
    }
  }
}
```

### POST /clientes
Crea un nuevo cliente.

#### Request
```json
{
  "tipo_documento": "cedula",
  "numero_documento": "87654321",
  "nombre": "María",
  "apellido": "Rodríguez",
  "email": "maria.rodriguez@email.com",
  "telefono": "3009876543",
  "direccion": "Carrera 45 #67-89",
  "fecha_nacimiento": "1990-08-22",
  "acepta_marketing": true,
  "inscribir_fidelizacion": true
}
```

#### Response (201 Created)
```json
{
  "success": true,
  "data": {
    "cliente": {
      "id": "uuid-cliente-nuevo",
      "tipo_documento": "cedula",
      "numero_documento": "87654321",
      "nombre": "María",
      "apellido": "Rodríguez",
      // ... resto de campos
      "fidelizacion": {
        "numero_tarjeta": "FIDE-001235",
        "puntos_disponibles": 0,
        "nivel": "bronce",
        "fecha_afiliacion": "2024-01-20T15:45:00Z"
      }
    }
  }
}
```

## Health Check y Monitoreo

### GET /health
Verifica el estado del API POS.

#### Response (200 OK)
```json
{
  "status": "healthy",
  "timestamp": "2024-01-20T15:30:00Z",
  "version": "1.0.0",
  "checks": {
    "database": {
      "status": "healthy",
      "response_time_ms": 5,
      "connections_active": 8,
      "connections_max": 50
    },
    "cache": {
      "status": "healthy",
      "hit_rate": 0.85,
      "memory_usage_mb": 128
    },
    "external_services": {
      "erp_system": {
        "status": "healthy",
        "last_sync": "2024-01-20T15:25:00Z"
      }
    }
  },
  "metrics": {
    "requests_per_minute": 45,
    "average_response_time_ms": 120,
    "active_sessions": 12,
    "active_terminals": 3
  }
}
```

## Códigos de Error

### Errores de Autenticación (401)
- `AUTH_TOKEN_REQUIRED`: Token de autenticación requerido
- `AUTH_TOKEN_INVALID`: Token inválido o expirado
- `AUTH_INVALID_CREDENTIALS`: Credenciales incorrectas
- `AUTH_USER_INACTIVE`: Usuario inactivo
- `AUTH_TERMINAL_REQUIRED`: Terminal requerido para esta operación

### Errores de Autorización (403)
- `AUTH_INSUFFICIENT_ROLE`: Rol insuficiente
- `AUTH_INSUFFICIENT_PERMISSION`: Permiso insuficiente
- `AUTH_SUCURSAL_REQUIRED`: Sucursal requerida

### Errores de Validación (400)
- `VALIDATION_REQUIRED_FIELD`: Campo requerido faltante
- `VALIDATION_INVALID_FORMAT`: Formato inválido
- `VALIDATION_INVALID_VALUE`: Valor inválido
- `VALIDATION_DUPLICATE_VALUE`: Valor duplicado

### Errores de Negocio (422)
- `BUSINESS_INSUFFICIENT_STOCK`: Stock insuficiente
- `BUSINESS_PRODUCT_INACTIVE`: Producto inactivo
- `BUSINESS_INVALID_PRICE`: Precio inválido
- `BUSINESS_PAYMENT_INSUFFICIENT`: Pago insuficiente

### Errores del Sistema (500)
- `SYSTEM_DATABASE_ERROR`: Error de base de datos
- `SYSTEM_EXTERNAL_SERVICE_ERROR`: Error en servicio externo
- `SYSTEM_INTERNAL_ERROR`: Error interno del sistema

## Rate Limiting

El API POS implementa rate limiting para proteger el sistema:

- **Límite por defecto**: 200 requests por minuto por IP
- **Burst**: 20 requests simultáneos
- **Headers de respuesta**:
  - `X-RateLimit-Limit`: Límite por minuto
  - `X-RateLimit-Remaining`: Requests restantes
  - `X-RateLimit-Reset`: Timestamp de reset

## Configuración

El API POS puede configurarse mediante el archivo `configs/pos/pos-config.yaml`:

```yaml
api:
  enabled: true
  base_path: "/api/pos"

performance:
  max_concurrent_users: 100
  session_timeout: "8h"
  max_products_per_query: 1000

ventas:
  max_venta_items: 100
  allow_negative_stock: false
  require_customer_for_credit: true

fidelizacion:
  enable_fidelizacion: true
  points_per_peso: 1
  min_purchase_for_points: 1000
```

---

**Autor**: Manus AI  
**Versión**: 1.0.0  
**Fecha**: 2024-01-XX

