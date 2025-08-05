# API POS - Documentación Técnica

**Sistema FERRE-POS - Servidor Central**  
**Versión**: 1.0.0  
**Puerto**: 8080  
**Autor**: Manus AI  
**Fecha**: Enero 2025

---

## Tabla de Contenidos

1. [Introducción](#introducción)
2. [Autenticación y Autorización](#autenticación-y-autorización)
3. [Endpoints de Usuarios](#endpoints-de-usuarios)
4. [Endpoints de Productos](#endpoints-de-productos)
5. [Endpoints de Categorías](#endpoints-de-categorías)
6. [Endpoints de Ventas](#endpoints-de-ventas)
7. [Endpoints de Stock](#endpoints-de-stock)
8. [Endpoints de Sucursales](#endpoints-de-sucursales)
9. [Health Checks y Monitoreo](#health-checks-y-monitoreo)
10. [Códigos de Error](#códigos-de-error)
11. [Ejemplos de Integración](#ejemplos-de-integración)
12. [Referencias](#referencias)

---

## Introducción

La API POS (Point of Sale) es el núcleo del sistema FERRE-POS, diseñada para gestionar todas las operaciones relacionadas con el punto de venta, incluyendo la administración de usuarios, productos, categorías, ventas y control de stock. Esta API REST está construida en Go utilizando el framework Gin y está optimizada para alto rendimiento y concurrencia.

La API POS maneja las operaciones críticas del negocio, desde la autenticación de usuarios hasta el procesamiento de ventas en tiempo real. Está diseñada siguiendo los principios RESTful y proporciona respuestas consistentes en formato JSON. Todas las operaciones están protegidas por un sistema de autenticación JWT robusto y un sistema de autorización basado en roles que garantiza que solo los usuarios autorizados puedan acceder a funcionalidades específicas.

El sistema implementa rate limiting para prevenir abuso de la API, logging estructurado para auditoría y troubleshooting, y métricas de Prometheus para monitoreo en tiempo real. La API está diseñada para ser escalable y puede manejar múltiples terminales POS concurrentemente sin degradación del rendimiento.

### Características Principales

La API POS incluye un sistema completo de gestión de inventario que permite el seguimiento en tiempo real del stock disponible, reservado y en tránsito. El sistema de ventas soporta múltiples métodos de pago, descuentos, y generación automática de documentos fiscales. La gestión de usuarios incluye diferentes roles con permisos granulares, desde cajeros con acceso limitado hasta administradores con control total del sistema.

El sistema de productos permite la gestión completa del catálogo, incluyendo códigos de barras, precios dinámicos, categorización jerárquica y control de stock por sucursal. La API también proporciona funcionalidades avanzadas como búsqueda de productos, filtrado por categorías, y reportes de ventas en tiempo real.

### Arquitectura y Tecnologías

La API está construida utilizando Go 1.21+ con el framework Gin para el manejo de rutas HTTP. Utiliza PostgreSQL como base de datos principal con conexiones pooled para optimizar el rendimiento. El sistema de autenticación está basado en JWT (JSON Web Tokens) con refresh tokens para mantener sesiones seguras de larga duración.

Para el manejo de concurrencia, la API utiliza goroutines y channels de Go, implementando patrones de worker pools para operaciones intensivas. El sistema de logging utiliza Logrus con rotación automática de archivos, y las métricas se exponen en formato Prometheus para integración con sistemas de monitoreo modernos.




## Autenticación y Autorización

### Sistema de Autenticación JWT

La API POS utiliza un sistema de autenticación basado en JSON Web Tokens (JWT) que proporciona seguridad robusta y escalabilidad para entornos distribuidos. El sistema implementa un patrón de doble token con access tokens de corta duración y refresh tokens de larga duración para balancear seguridad y usabilidad.

Cuando un usuario se autentica exitosamente, el sistema genera dos tokens: un access token válido por 24 horas que se utiliza para autorizar requests a la API, y un refresh token válido por 7 días que permite renovar el access token sin requerir nuevas credenciales. Este enfoque minimiza la exposición de credenciales mientras mantiene una experiencia de usuario fluida.

Los tokens JWT incluyen claims personalizados que especifican el rol del usuario, la sucursal asignada, y permisos específicos. Esto permite que la API tome decisiones de autorización sin consultar la base de datos en cada request, mejorando significativamente el rendimiento. Los tokens están firmados utilizando el algoritmo HS256 con una clave secreta robusta que se configura a través de variables de entorno.

### Roles y Permisos

El sistema implementa un modelo de autorización basado en roles (RBAC) con cinco roles principales: admin, supervisor, vendedor, cajero, y operador_etiquetas. Cada rol tiene permisos específicos que determinan qué endpoints y operaciones puede acceder el usuario.

Los administradores tienen acceso completo a todas las funcionalidades, incluyendo la gestión de usuarios, configuración del sistema, y acceso a reportes financieros detallados. Los supervisores pueden gestionar operaciones diarias, aprobar devoluciones, y acceder a reportes de su sucursal. Los vendedores pueden crear ventas, consultar productos, y gestionar clientes, pero no pueden modificar precios o acceder a información financiera sensible.

Los cajeros tienen permisos limitados enfocados en el procesamiento de ventas y cobros, mientras que los operadores de etiquetas se especializan en la gestión del sistema de etiquetado y códigos de barras. Esta segmentación de permisos garantiza que cada usuario solo pueda acceder a las funcionalidades necesarias para su rol específico.

### Endpoints de Autenticación

#### POST /api/v1/auth/login

Autentica un usuario y devuelve tokens JWT para acceso a la API.

**Request Body:**
```json
{
  "username": "vendedor_test",
  "password": "password123",
  "terminal_id": "test-terminal-1"
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "token_type": "Bearer",
    "expires_in": 86400,
    "user": {
      "id": "test-user-2",
      "username": "vendedor_test",
      "email": "vendedor.test@ferrepos.com",
      "nombre_completo": "Vendedor Test",
      "rol": "vendedor",
      "sucursal_id": "test-sucursal-1",
      "activo": true
    }
  },
  "request_id": "req_123456789",
  "timestamp": "2025-01-08T10:30:00Z"
}
```

**Errores Comunes:**
- `401 UNAUTHORIZED`: Credenciales inválidas
- `403 FORBIDDEN`: Usuario inactivo o terminal no autorizada
- `429 TOO_MANY_REQUESTS`: Demasiados intentos de login

#### POST /api/v1/auth/refresh

Renueva un access token utilizando un refresh token válido.

**Request Body:**
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "token_type": "Bearer",
    "expires_in": 86400
  },
  "request_id": "req_123456790",
  "timestamp": "2025-01-08T10:35:00Z"
}
```

#### POST /api/v1/auth/logout

Invalida los tokens del usuario actual y cierra la sesión.

**Headers:**
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "message": "Sesión cerrada exitosamente"
  },
  "request_id": "req_123456791",
  "timestamp": "2025-01-08T10:40:00Z"
}
```

### Middleware de Autorización

Todos los endpoints protegidos utilizan middleware de autorización que valida el token JWT y verifica los permisos del usuario. El middleware extrae el token del header Authorization, valida su firma y expiración, y carga la información del usuario en el contexto del request.

Para endpoints que requieren permisos específicos, el middleware verifica que el rol del usuario tenga los permisos necesarios. Si el usuario no tiene los permisos requeridos, el middleware devuelve un error 403 Forbidden con detalles específicos sobre los permisos faltantes.

El sistema también implementa rate limiting por usuario para prevenir abuso de la API. Los límites se configuran por rol, con administradores teniendo límites más altos que usuarios regulares. El rate limiting utiliza un algoritmo token bucket que permite ráfagas de requests mientras mantiene un promedio sostenible.


## Endpoints de Usuarios

### Gestión de Usuarios

La gestión de usuarios en la API POS permite a los administradores y supervisores crear, modificar, consultar y desactivar cuentas de usuario. El sistema mantiene un registro completo de la actividad de usuarios, incluyendo último acceso, intentos de login fallidos, y cambios de configuración.

Los usuarios están asociados a una sucursal específica y tienen un rol que determina sus permisos dentro del sistema. El sistema soporta la gestión de múltiples sucursales con usuarios que pueden tener diferentes niveles de acceso según su ubicación y responsabilidades.

#### GET /api/v1/users

Obtiene una lista paginada de usuarios con filtros opcionales.

**Permisos Requeridos:** admin, supervisor

**Query Parameters:**
- `page` (int, opcional): Número de página (default: 1)
- `per_page` (int, opcional): Registros por página (default: 20, max: 100)
- `search` (string, opcional): Búsqueda por nombre o username
- `rol` (string, opcional): Filtrar por rol específico
- `sucursal_id` (uuid, opcional): Filtrar por sucursal
- `activo` (bool, opcional): Filtrar por estado activo/inactivo

**Headers:**
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
Content-Type: application/json
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": [
    {
      "id": "test-user-1",
      "username": "admin_test",
      "email": "admin.test@ferrepos.com",
      "nombre_completo": "Administrador Test",
      "rol": "admin",
      "sucursal_id": "test-sucursal-1",
      "sucursal_nombre": "Test Sucursal Principal",
      "activo": true,
      "ultimo_acceso": "2025-01-08T09:15:00Z",
      "created_at": "2025-01-01T00:00:00Z",
      "updated_at": "2025-01-08T09:15:00Z"
    },
    {
      "id": "test-user-2",
      "username": "vendedor_test",
      "email": "vendedor.test@ferrepos.com",
      "nombre_completo": "Vendedor Test",
      "rol": "vendedor",
      "sucursal_id": "test-sucursal-1",
      "sucursal_nombre": "Test Sucursal Principal",
      "activo": true,
      "ultimo_acceso": "2025-01-08T10:30:00Z",
      "created_at": "2025-01-01T00:00:00Z",
      "updated_at": "2025-01-08T10:30:00Z"
    }
  ],
  "meta": {
    "page": 1,
    "per_page": 20,
    "total": 4,
    "total_pages": 1
  },
  "request_id": "req_123456792",
  "timestamp": "2025-01-08T10:45:00Z"
}
```

#### GET /api/v1/users/{id}

Obtiene los detalles de un usuario específico.

**Permisos Requeridos:** admin, supervisor, o el propio usuario

**Path Parameters:**
- `id` (uuid, requerido): ID del usuario

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "id": "test-user-2",
    "username": "vendedor_test",
    "email": "vendedor.test@ferrepos.com",
    "nombre_completo": "Vendedor Test",
    "rol": "vendedor",
    "sucursal_id": "test-sucursal-1",
    "sucursal_nombre": "Test Sucursal Principal",
    "activo": true,
    "ultimo_acceso": "2025-01-08T10:30:00Z",
    "intentos_fallidos": 0,
    "configuracion": {
      "idioma": "es",
      "tema": "light",
      "notificaciones": true
    },
    "permisos": [
      "ventas.crear",
      "ventas.consultar",
      "productos.consultar",
      "clientes.gestionar"
    ],
    "created_at": "2025-01-01T00:00:00Z",
    "updated_at": "2025-01-08T10:30:00Z"
  },
  "request_id": "req_123456793",
  "timestamp": "2025-01-08T10:50:00Z"
}
```

#### POST /api/v1/users

Crea un nuevo usuario en el sistema.

**Permisos Requeridos:** admin

**Request Body:**
```json
{
  "username": "nuevo_vendedor",
  "email": "nuevo.vendedor@ferrepos.com",
  "password": "password123",
  "nombre_completo": "Nuevo Vendedor",
  "rol": "vendedor",
  "sucursal_id": "test-sucursal-1",
  "activo": true,
  "configuracion": {
    "idioma": "es",
    "tema": "light",
    "notificaciones": true
  }
}
```

**Validaciones:**
- `username`: Requerido, único, 3-50 caracteres, solo alfanuméricos y guiones bajos
- `email`: Requerido, único, formato de email válido
- `password`: Requerido, mínimo 8 caracteres, debe incluir mayúsculas, minúsculas y números
- `nombre_completo`: Requerido, 2-100 caracteres
- `rol`: Requerido, debe ser uno de: admin, supervisor, vendedor, cajero, operador_etiquetas
- `sucursal_id`: Requerido, debe existir en el sistema

**Response (201 Created):**
```json
{
  "success": true,
  "data": {
    "id": "new-user-uuid",
    "username": "nuevo_vendedor",
    "email": "nuevo.vendedor@ferrepos.com",
    "nombre_completo": "Nuevo Vendedor",
    "rol": "vendedor",
    "sucursal_id": "test-sucursal-1",
    "activo": true,
    "created_at": "2025-01-08T11:00:00Z",
    "updated_at": "2025-01-08T11:00:00Z"
  },
  "request_id": "req_123456794",
  "timestamp": "2025-01-08T11:00:00Z"
}
```

#### PUT /api/v1/users/{id}

Actualiza la información de un usuario existente.

**Permisos Requeridos:** admin, o el propio usuario (con limitaciones)

**Path Parameters:**
- `id` (uuid, requerido): ID del usuario

**Request Body:**
```json
{
  "email": "vendedor.actualizado@ferrepos.com",
  "nombre_completo": "Vendedor Actualizado",
  "activo": true,
  "configuracion": {
    "idioma": "es",
    "tema": "dark",
    "notificaciones": false
  }
}
```

**Notas:**
- Los usuarios no-admin solo pueden actualizar su email, nombre_completo y configuración
- Solo los admin pueden cambiar rol, sucursal_id, y estado activo
- El username no se puede cambiar después de la creación

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "id": "test-user-2",
    "username": "vendedor_test",
    "email": "vendedor.actualizado@ferrepos.com",
    "nombre_completo": "Vendedor Actualizado",
    "rol": "vendedor",
    "sucursal_id": "test-sucursal-1",
    "activo": true,
    "updated_at": "2025-01-08T11:05:00Z"
  },
  "request_id": "req_123456795",
  "timestamp": "2025-01-08T11:05:00Z"
}
```

#### POST /api/v1/users/{id}/change-password

Cambia la contraseña de un usuario.

**Permisos Requeridos:** admin, o el propio usuario

**Path Parameters:**
- `id` (uuid, requerido): ID del usuario

**Request Body:**
```json
{
  "current_password": "password123",
  "new_password": "newpassword456",
  "confirm_password": "newpassword456"
}
```

**Notas:**
- Los usuarios deben proporcionar su contraseña actual
- Los admin pueden cambiar contraseñas sin proporcionar la contraseña actual
- La nueva contraseña debe cumplir con las políticas de seguridad

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "message": "Contraseña actualizada exitosamente",
    "password_changed_at": "2025-01-08T11:10:00Z"
  },
  "request_id": "req_123456796",
  "timestamp": "2025-01-08T11:10:00Z"
}
```

#### DELETE /api/v1/users/{id}

Desactiva un usuario (soft delete).

**Permisos Requeridos:** admin

**Path Parameters:**
- `id` (uuid, requerido): ID del usuario

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "message": "Usuario desactivado exitosamente",
    "deactivated_at": "2025-01-08T11:15:00Z"
  },
  "request_id": "req_123456797",
  "timestamp": "2025-01-08T11:15:00Z"
}
```

### Gestión de Sesiones

#### GET /api/v1/users/me/sessions

Obtiene las sesiones activas del usuario actual.

**Permisos Requeridos:** Usuario autenticado

**Response (200 OK):**
```json
{
  "success": true,
  "data": [
    {
      "session_id": "sess_123456",
      "terminal_id": "test-terminal-1",
      "terminal_nombre": "Terminal Test Principal",
      "ip_address": "192.168.1.100",
      "user_agent": "FERRE-POS Terminal v1.0",
      "created_at": "2025-01-08T10:30:00Z",
      "last_activity": "2025-01-08T11:15:00Z",
      "is_current": true
    }
  ],
  "request_id": "req_123456798",
  "timestamp": "2025-01-08T11:20:00Z"
}
```

#### DELETE /api/v1/users/me/sessions/{session_id}

Cierra una sesión específica del usuario.

**Permisos Requeridos:** Usuario autenticado

**Path Parameters:**
- `session_id` (string, requerido): ID de la sesión

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "message": "Sesión cerrada exitosamente"
  },
  "request_id": "req_123456799",
  "timestamp": "2025-01-08T11:25:00Z"
}
```


## Endpoints de Productos

### Gestión de Productos

El sistema de gestión de productos es el corazón del inventario de FERRE-POS, permitiendo el control completo del catálogo de productos, precios, stock, y metadatos asociados. Cada producto tiene un código único, código de barras, información descriptiva, categorización, y configuración de precios y stock.

El sistema soporta productos con múltiples variantes, códigos de barras alternativos, precios por sucursal, y gestión de stock en tiempo real. Los productos pueden tener imágenes asociadas, especificaciones técnicas, y información de proveedores para facilitar la gestión del inventario.

#### GET /api/v1/products

Obtiene una lista paginada de productos con filtros y búsqueda avanzada.

**Permisos Requeridos:** vendedor, cajero, supervisor, admin

**Query Parameters:**
- `page` (int, opcional): Número de página (default: 1)
- `per_page` (int, opcional): Registros por página (default: 20, max: 100)
- `search` (string, opcional): Búsqueda por código, nombre, o código de barras
- `categoria_id` (uuid, opcional): Filtrar por categoría específica
- `sucursal_id` (uuid, requerido): Sucursal para consultar stock
- `con_stock` (bool, opcional): Solo productos con stock disponible
- `activos` (bool, opcional): Solo productos activos (default: true)
- `precio_min` (float, opcional): Precio mínimo
- `precio_max` (float, opcional): Precio máximo
- `order_by` (string, opcional): Campo para ordenar (codigo, nombre, precio, stock)
- `order_dir` (string, opcional): Dirección del orden (asc, desc)

**Headers:**
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
Content-Type: application/json
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": [
    {
      "id": "test-product-1",
      "codigo": "TEST001",
      "codigo_barras": "1234567890123",
      "nombre": "Martillo Test 500g",
      "descripcion": "Martillo de prueba con mango de madera",
      "categoria_id": "test-category-1",
      "categoria_nombre": "Test Herramientas",
      "precio": 15000.00,
      "costo": 8000.00,
      "margen": 87.5,
      "stock_minimo": 5,
      "activo": true,
      "stock": {
        "actual": 25,
        "reservado": 0,
        "disponible": 25,
        "en_transito": 0
      },
      "imagenes": [
        {
          "url": "/images/products/test-product-1-main.jpg",
          "tipo": "principal",
          "orden": 1
        }
      ],
      "especificaciones": {
        "peso": "500g",
        "material": "Acero forjado",
        "mango": "Madera"
      },
      "created_at": "2025-01-01T00:00:00Z",
      "updated_at": "2025-01-08T09:00:00Z"
    },
    {
      "id": "test-product-2",
      "codigo": "TEST002",
      "codigo_barras": "1234567890124",
      "nombre": "Destornillador Test Phillips",
      "descripcion": "Destornillador Phillips de prueba",
      "categoria_id": "test-category-1",
      "categoria_nombre": "Test Herramientas",
      "precio": 3500.00,
      "costo": 2000.00,
      "margen": 75.0,
      "stock_minimo": 10,
      "activo": true,
      "stock": {
        "actual": 50,
        "reservado": 2,
        "disponible": 48,
        "en_transito": 0
      },
      "especificaciones": {
        "tipo": "Phillips",
        "tamaño": "#2",
        "longitud": "150mm"
      },
      "created_at": "2025-01-01T00:00:00Z",
      "updated_at": "2025-01-08T09:00:00Z"
    }
  ],
  "meta": {
    "page": 1,
    "per_page": 20,
    "total": 8,
    "total_pages": 1,
    "filters_applied": {
      "sucursal_id": "test-sucursal-1",
      "activos": true
    }
  },
  "request_id": "req_123456800",
  "timestamp": "2025-01-08T11:30:00Z"
}
```

#### GET /api/v1/products/{id}

Obtiene los detalles completos de un producto específico.

**Permisos Requeridos:** vendedor, cajero, supervisor, admin

**Path Parameters:**
- `id` (uuid, requerido): ID del producto

**Query Parameters:**
- `sucursal_id` (uuid, requerido): Sucursal para consultar stock

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "id": "test-product-1",
    "codigo": "TEST001",
    "codigo_barras": "1234567890123",
    "codigos_alternativos": [
      "ALT001",
      "PROV123"
    ],
    "nombre": "Martillo Test 500g",
    "descripcion": "Martillo de prueba con mango de madera resistente, ideal para trabajos de carpintería y construcción general",
    "categoria_id": "test-category-1",
    "categoria_nombre": "Test Herramientas",
    "categoria_path": "Herramientas > Manuales > Martillos",
    "precio": 15000.00,
    "precio_anterior": 14500.00,
    "costo": 8000.00,
    "margen": 87.5,
    "stock_minimo": 5,
    "stock_maximo": 50,
    "activo": true,
    "destacado": false,
    "stock": {
      "actual": 25,
      "reservado": 0,
      "disponible": 25,
      "en_transito": 5,
      "ubicacion": "A-1-3",
      "ultimo_movimiento": "2025-01-07T15:30:00Z"
    },
    "precios_sucursales": [
      {
        "sucursal_id": "test-sucursal-1",
        "sucursal_nombre": "Test Sucursal Principal",
        "precio": 15000.00,
        "precio_promocional": null,
        "fecha_promocion_inicio": null,
        "fecha_promocion_fin": null
      },
      {
        "sucursal_id": "test-sucursal-2",
        "sucursal_nombre": "Test Sucursal Secundaria",
        "precio": 15500.00,
        "precio_promocional": 14000.00,
        "fecha_promocion_inicio": "2025-01-08T00:00:00Z",
        "fecha_promocion_fin": "2025-01-15T23:59:59Z"
      }
    ],
    "imagenes": [
      {
        "id": "img_001",
        "url": "/images/products/test-product-1-main.jpg",
        "tipo": "principal",
        "orden": 1,
        "alt_text": "Martillo Test 500g vista principal"
      },
      {
        "id": "img_002",
        "url": "/images/products/test-product-1-detail.jpg",
        "tipo": "detalle",
        "orden": 2,
        "alt_text": "Detalle del mango de madera"
      }
    ],
    "especificaciones": {
      "peso": "500g",
      "material": "Acero forjado",
      "mango": "Madera de fresno",
      "longitud_total": "320mm",
      "longitud_mango": "280mm",
      "garantia": "12 meses",
      "origen": "Nacional"
    },
    "proveedor": {
      "id": "prov_001",
      "nombre": "Herramientas del Sur",
      "codigo_producto": "HDS-MART-500",
      "tiempo_entrega_dias": 3
    },
    "ventas_estadisticas": {
      "unidades_vendidas_mes": 15,
      "unidades_vendidas_año": 180,
      "ultima_venta": "2025-01-08T10:30:00Z",
      "rotacion_dias": 12
    },
    "created_at": "2025-01-01T00:00:00Z",
    "updated_at": "2025-01-08T09:00:00Z"
  },
  "request_id": "req_123456801",
  "timestamp": "2025-01-08T11:35:00Z"
}
```

#### POST /api/v1/products/search

Búsqueda avanzada de productos con múltiples criterios.

**Permisos Requeridos:** vendedor, cajero, supervisor, admin

**Request Body:**
```json
{
  "query": "martillo",
  "sucursal_id": "test-sucursal-1",
  "filtros": {
    "categorias": ["test-category-1"],
    "precio_min": 10000,
    "precio_max": 50000,
    "con_stock": true,
    "activos": true,
    "destacados": false
  },
  "ordenamiento": {
    "campo": "nombre",
    "direccion": "asc"
  },
  "paginacion": {
    "page": 1,
    "per_page": 10
  }
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "productos": [
      {
        "id": "test-product-1",
        "codigo": "TEST001",
        "codigo_barras": "1234567890123",
        "nombre": "Martillo Test 500g",
        "precio": 15000.00,
        "stock_disponible": 25,
        "categoria_nombre": "Test Herramientas",
        "imagen_principal": "/images/products/test-product-1-main.jpg",
        "relevancia": 0.95
      }
    ],
    "sugerencias": [
      "martillo carpintero",
      "martillo demolición",
      "martillo goma"
    ],
    "filtros_aplicados": {
      "query": "martillo",
      "categorias": 1,
      "rango_precio": "10000-50000",
      "con_stock": true
    }
  },
  "meta": {
    "total_encontrados": 1,
    "tiempo_busqueda_ms": 45,
    "page": 1,
    "per_page": 10
  },
  "request_id": "req_123456802",
  "timestamp": "2025-01-08T11:40:00Z"
}
```

#### POST /api/v1/products

Crea un nuevo producto en el sistema.

**Permisos Requeridos:** admin, supervisor

**Request Body:**
```json
{
  "codigo": "TEST009",
  "codigo_barras": "1234567890131",
  "codigos_alternativos": ["ALT009"],
  "nombre": "Taladro Inalámbrico 18V",
  "descripcion": "Taladro inalámbrico profesional con batería de litio",
  "categoria_id": "test-category-1",
  "precio": 85000.00,
  "costo": 55000.00,
  "stock_minimo": 3,
  "stock_maximo": 15,
  "activo": true,
  "destacado": true,
  "especificaciones": {
    "voltaje": "18V",
    "tipo_bateria": "Litio",
    "torque_max": "65 Nm",
    "velocidades": "2",
    "mandril": "13mm"
  },
  "stock_inicial": [
    {
      "sucursal_id": "test-sucursal-1",
      "cantidad": 10,
      "ubicacion": "A-2-1"
    }
  ]
}
```

**Validaciones:**
- `codigo`: Requerido, único, 3-20 caracteres alfanuméricos
- `codigo_barras`: Requerido, único, formato válido (EAN-13, UPC, etc.)
- `nombre`: Requerido, 3-200 caracteres
- `categoria_id`: Requerido, debe existir en el sistema
- `precio`: Requerido, mayor a 0
- `costo`: Opcional, debe ser menor al precio si se proporciona

**Response (201 Created):**
```json
{
  "success": true,
  "data": {
    "id": "new-product-uuid",
    "codigo": "TEST009",
    "codigo_barras": "1234567890131",
    "nombre": "Taladro Inalámbrico 18V",
    "categoria_id": "test-category-1",
    "precio": 85000.00,
    "costo": 55000.00,
    "margen": 54.5,
    "activo": true,
    "created_at": "2025-01-08T11:45:00Z"
  },
  "request_id": "req_123456803",
  "timestamp": "2025-01-08T11:45:00Z"
}
```

#### PUT /api/v1/products/{id}

Actualiza un producto existente.

**Permisos Requeridos:** admin, supervisor

**Path Parameters:**
- `id` (uuid, requerido): ID del producto

**Request Body:**
```json
{
  "nombre": "Martillo Test 500g - Mejorado",
  "descripcion": "Martillo de prueba con mango de madera reforzado",
  "precio": 16000.00,
  "stock_minimo": 8,
  "especificaciones": {
    "peso": "500g",
    "material": "Acero forjado premium",
    "mango": "Madera de fresno reforzado",
    "garantia": "24 meses"
  }
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "id": "test-product-1",
    "codigo": "TEST001",
    "nombre": "Martillo Test 500g - Mejorado",
    "precio": 16000.00,
    "precio_anterior": 15000.00,
    "updated_at": "2025-01-08T11:50:00Z"
  },
  "request_id": "req_123456804",
  "timestamp": "2025-01-08T11:50:00Z"
}
```

#### GET /api/v1/products/barcode/{codigo_barras}

Busca un producto por su código de barras.

**Permisos Requeridos:** vendedor, cajero, supervisor, admin

**Path Parameters:**
- `codigo_barras` (string, requerido): Código de barras del producto

**Query Parameters:**
- `sucursal_id` (uuid, requerido): Sucursal para consultar stock

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "id": "test-product-1",
    "codigo": "TEST001",
    "codigo_barras": "1234567890123",
    "nombre": "Martillo Test 500g",
    "precio": 15000.00,
    "stock_disponible": 25,
    "categoria_nombre": "Test Herramientas",
    "activo": true,
    "imagen_principal": "/images/products/test-product-1-main.jpg"
  },
  "request_id": "req_123456805",
  "timestamp": "2025-01-08T11:55:00Z"
}
```

#### POST /api/v1/products/{id}/images

Sube una imagen para un producto.

**Permisos Requeridos:** admin, supervisor

**Path Parameters:**
- `id` (uuid, requerido): ID del producto

**Request:** Multipart form data
- `image` (file, requerido): Archivo de imagen (JPG, PNG, max 5MB)
- `tipo` (string, requerido): Tipo de imagen (principal, detalle, galeria)
- `orden` (int, opcional): Orden de visualización
- `alt_text` (string, opcional): Texto alternativo

**Response (201 Created):**
```json
{
  "success": true,
  "data": {
    "id": "img_003",
    "url": "/images/products/test-product-1-new.jpg",
    "tipo": "detalle",
    "orden": 3,
    "alt_text": "Nueva vista del producto",
    "uploaded_at": "2025-01-08T12:00:00Z"
  },
  "request_id": "req_123456806",
  "timestamp": "2025-01-08T12:00:00Z"
}
```


## Endpoints de Ventas

### Gestión de Ventas

El sistema de ventas de FERRE-POS permite el procesamiento completo de transacciones, desde la creación de ventas hasta la generación de documentos fiscales. El sistema soporta múltiples métodos de pago, descuentos, devoluciones, y diferentes tipos de documentos como boletas, facturas, y notas de venta.

Cada venta se registra con información completa del cliente, vendedor, cajero, terminal utilizada, y detalles de todos los productos vendidos. El sistema mantiene trazabilidad completa de todas las transacciones para auditoría y reportes.

#### POST /api/v1/sales

Crea una nueva venta en el sistema.

**Permisos Requeridos:** vendedor, cajero, supervisor, admin

**Request Body:**
```json
{
  "sucursal_id": "test-sucursal-1",
  "terminal_id": "test-terminal-1",
  "cajero_id": "test-user-3",
  "vendedor_id": "test-user-2",
  "cliente": {
    "rut": "12345678-9",
    "nombre": "Juan Pérez",
    "email": "juan.perez@email.com",
    "telefono": "+56912345678",
    "direccion": "Av. Principal 123"
  },
  "tipo_documento": "boleta",
  "items": [
    {
      "producto_id": "test-product-1",
      "cantidad": 2,
      "precio_unitario": 15000.00,
      "descuento_unitario": 0.00,
      "numero_serie": null,
      "lote": null
    },
    {
      "producto_id": "test-product-3",
      "cantidad": 10,
      "precio_unitario": 150.00,
      "descuento_unitario": 10.00,
      "numero_serie": null,
      "lote": "LOTE2025001"
    }
  ],
  "medios_pago": [
    {
      "medio_pago": "efectivo",
      "monto": 20000.00
    },
    {
      "medio_pago": "tarjeta_debito",
      "monto": 11400.00,
      "referencia_transaccion": "TXN123456789",
      "codigo_autorizacion": "AUTH001"
    }
  ],
  "descuento_global": 0.00,
  "observaciones": "Venta con descuento por volumen en tornillos",
  "datos_adicionales": {
    "promocion_aplicada": "DESC_VOLUMEN_10",
    "vendedor_comision": 2.5
  }
}
```

**Validaciones:**
- Stock suficiente para todos los productos
- Precios válidos y actualizados
- Métodos de pago válidos
- Total de medios de pago igual al total de la venta
- Terminal activa y autorizada

**Response (201 Created):**
```json
{
  "success": true,
  "data": {
    "id": "new-sale-uuid",
    "numero_documento": "B-001-00000123",
    "tipo_documento": "boleta",
    "fecha_venta": "2025-01-08T12:15:00Z",
    "sucursal": {
      "id": "test-sucursal-1",
      "nombre": "Test Sucursal Principal"
    },
    "terminal": {
      "id": "test-terminal-1",
      "codigo": "TERM001"
    },
    "cajero": {
      "id": "test-user-3",
      "nombre_completo": "Cajero Test"
    },
    "vendedor": {
      "id": "test-user-2",
      "nombre_completo": "Vendedor Test"
    },
    "cliente": {
      "rut": "12345678-9",
      "nombre": "Juan Pérez"
    },
    "items": [
      {
        "producto_id": "test-product-1",
        "codigo": "TEST001",
        "nombre": "Martillo Test 500g",
        "cantidad": 2,
        "precio_unitario": 15000.00,
        "descuento_unitario": 0.00,
        "subtotal_linea": 30000.00,
        "numero_linea": 1
      },
      {
        "producto_id": "test-product-3",
        "codigo": "TEST003",
        "nombre": "Tornillo Test 6x40mm",
        "cantidad": 10,
        "precio_unitario": 150.00,
        "descuento_unitario": 10.00,
        "subtotal_linea": 1400.00,
        "numero_linea": 2
      }
    ],
    "totales": {
      "subtotal": 31400.00,
      "descuento_total": 100.00,
      "subtotal_neto": 31300.00,
      "impuesto_total": 5947.00,
      "total": 31400.00
    },
    "medios_pago": [
      {
        "medio_pago": "efectivo",
        "monto": 20000.00
      },
      {
        "medio_pago": "tarjeta_debito",
        "monto": 11400.00,
        "referencia_transaccion": "TXN123456789"
      }
    ],
    "estado": "procesado",
    "documentos_generados": [
      {
        "tipo": "boleta_electronica",
        "numero": "B-001-00000123",
        "url": "/documents/boletas/B-001-00000123.pdf"
      }
    ],
    "created_at": "2025-01-08T12:15:00Z"
  },
  "request_id": "req_123456807",
  "timestamp": "2025-01-08T12:15:00Z"
}
```

#### GET /api/v1/sales

Obtiene una lista paginada de ventas con filtros.

**Permisos Requeridos:** vendedor, cajero, supervisor, admin

**Query Parameters:**
- `page` (int, opcional): Número de página (default: 1)
- `per_page` (int, opcional): Registros por página (default: 20, max: 100)
- `sucursal_id` (uuid, opcional): Filtrar por sucursal
- `terminal_id` (uuid, opcional): Filtrar por terminal
- `cajero_id` (uuid, opcional): Filtrar por cajero
- `vendedor_id` (uuid, opcional): Filtrar por vendedor
- `fecha_desde` (date, opcional): Fecha inicio (YYYY-MM-DD)
- `fecha_hasta` (date, opcional): Fecha fin (YYYY-MM-DD)
- `tipo_documento` (string, opcional): Tipo de documento
- `estado` (string, opcional): Estado de la venta
- `cliente_rut` (string, opcional): RUT del cliente
- `numero_documento` (string, opcional): Número de documento

**Response (200 OK):**
```json
{
  "success": true,
  "data": [
    {
      "id": "test-venta-1",
      "numero_documento": "TEST-V-001",
      "tipo_documento": "boleta",
      "fecha_venta": "2025-01-06T10:30:00Z",
      "cliente_nombre": "Cliente Test 1",
      "cajero_nombre": "Cajero Test",
      "vendedor_nombre": "Vendedor Test",
      "total": 29750.00,
      "estado": "procesado",
      "metodo_pago": "efectivo",
      "items_count": 3
    }
  ],
  "meta": {
    "page": 1,
    "per_page": 20,
    "total": 4,
    "total_pages": 1,
    "totales_periodo": {
      "cantidad_ventas": 4,
      "monto_total": 116025.00,
      "monto_promedio": 29006.25
    }
  },
  "request_id": "req_123456808",
  "timestamp": "2025-01-08T12:20:00Z"
}
```

#### GET /api/v1/sales/{id}

Obtiene los detalles completos de una venta específica.

**Permisos Requeridos:** vendedor, cajero, supervisor, admin

**Path Parameters:**
- `id` (uuid, requerido): ID de la venta

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "id": "test-venta-1",
    "numero_documento": "TEST-V-001",
    "tipo_documento": "boleta",
    "fecha_venta": "2025-01-06T10:30:00Z",
    "sucursal": {
      "id": "test-sucursal-1",
      "nombre": "Test Sucursal Principal",
      "direccion": "Av. Test 123, Santiago"
    },
    "terminal": {
      "id": "test-terminal-1",
      "codigo": "TERM001",
      "nombre": "Terminal Test Principal"
    },
    "cajero": {
      "id": "test-user-3",
      "username": "cajero_test",
      "nombre_completo": "Cajero Test"
    },
    "vendedor": {
      "id": "test-user-2",
      "username": "vendedor_test",
      "nombre_completo": "Vendedor Test"
    },
    "cliente": {
      "rut": null,
      "nombre": "Cliente Genérico",
      "email": null,
      "telefono": null
    },
    "items": [
      {
        "id": "detail-1",
        "producto_id": "test-product-1",
        "codigo": "TEST001",
        "codigo_barras": "1234567890123",
        "nombre": "Martillo Test 500g",
        "cantidad": 1,
        "precio_unitario": 15000.00,
        "descuento_unitario": 0.00,
        "subtotal_linea": 15000.00,
        "numero_linea": 1,
        "numero_serie": null,
        "lote": null
      },
      {
        "id": "detail-2",
        "producto_id": "test-product-2",
        "codigo": "TEST002",
        "codigo_barras": "1234567890124",
        "nombre": "Destornillador Test Phillips",
        "cantidad": 2,
        "precio_unitario": 3500.00,
        "descuento_unitario": 0.00,
        "subtotal_linea": 7000.00,
        "numero_linea": 2,
        "numero_serie": null,
        "lote": null
      }
    ],
    "totales": {
      "subtotal": 25000.00,
      "descuento_total": 0.00,
      "subtotal_neto": 25000.00,
      "impuesto_total": 4750.00,
      "total": 29750.00
    },
    "medios_pago": [
      {
        "medio_pago": "efectivo",
        "monto": 29750.00,
        "referencia_transaccion": null,
        "codigo_autorizacion": null
      }
    ],
    "estado": "procesado",
    "observaciones": "Venta de prueba 1",
    "documentos": [
      {
        "tipo": "boleta",
        "numero": "TEST-V-001",
        "url": "/documents/boletas/TEST-V-001.pdf",
        "generado_at": "2025-01-06T10:30:00Z"
      }
    ],
    "auditoria": {
      "created_at": "2025-01-06T10:30:00Z",
      "updated_at": "2025-01-06T10:30:00Z",
      "created_by": "test-user-3",
      "ip_address": "192.168.1.100",
      "user_agent": "FERRE-POS Terminal v1.0"
    }
  },
  "request_id": "req_123456809",
  "timestamp": "2025-01-08T12:25:00Z"
}
```

#### POST /api/v1/sales/{id}/refund

Procesa una devolución total o parcial de una venta.

**Permisos Requeridos:** supervisor, admin

**Path Parameters:**
- `id` (uuid, requerido): ID de la venta original

**Request Body:**
```json
{
  "tipo_devolucion": "parcial",
  "items": [
    {
      "detalle_venta_id": "detail-1",
      "cantidad": 1,
      "motivo": "Producto defectuoso"
    }
  ],
  "motivo_general": "Cliente reporta defecto en el producto",
  "metodo_devolucion": "efectivo",
  "observaciones": "Devolución autorizada por supervisor"
}
```

**Response (201 Created):**
```json
{
  "success": true,
  "data": {
    "id": "refund-uuid",
    "venta_original_id": "test-venta-1",
    "numero_documento": "DEV-001-00000001",
    "tipo_devolucion": "parcial",
    "fecha_devolucion": "2025-01-08T12:30:00Z",
    "items_devueltos": [
      {
        "producto_id": "test-product-1",
        "codigo": "TEST001",
        "nombre": "Martillo Test 500g",
        "cantidad": 1,
        "monto_devuelto": 15000.00,
        "motivo": "Producto defectuoso"
      }
    ],
    "monto_total_devuelto": 15000.00,
    "metodo_devolucion": "efectivo",
    "autorizado_por": "test-user-4",
    "estado": "procesado",
    "created_at": "2025-01-08T12:30:00Z"
  },
  "request_id": "req_123456810",
  "timestamp": "2025-01-08T12:30:00Z"
}
```

## Health Checks y Monitoreo

### Endpoints de Salud

#### GET /health

Verifica el estado general de la API.

**Permisos Requeridos:** Ninguno (público)

**Response (200 OK):**
```json
{
  "status": "healthy",
  "timestamp": "2025-01-08T12:35:00Z",
  "version": "1.0.0",
  "uptime": "72h15m30s",
  "checks": {
    "database": "healthy",
    "redis": "healthy",
    "external_services": "healthy"
  }
}
```

#### GET /health/detailed

Verifica el estado detallado de todos los componentes.

**Permisos Requeridos:** admin

**Response (200 OK):**
```json
{
  "status": "healthy",
  "timestamp": "2025-01-08T12:35:00Z",
  "version": "1.0.0",
  "uptime": "72h15m30s",
  "system": {
    "memory_usage": "45.2%",
    "cpu_usage": "12.8%",
    "disk_usage": "67.3%",
    "goroutines": 156
  },
  "database": {
    "status": "healthy",
    "connections_active": 8,
    "connections_idle": 12,
    "response_time_ms": 2.3
  },
  "external_services": {
    "sii_service": {
      "status": "healthy",
      "last_check": "2025-01-08T12:34:00Z",
      "response_time_ms": 145
    }
  }
}
```

#### GET /metrics

Expone métricas en formato Prometheus.

**Permisos Requeridos:** Ninguno (público, pero restringido por IP)

**Response (200 OK):**
```
# HELP api_requests_total Total number of API requests
# TYPE api_requests_total counter
api_requests_total{method="GET",endpoint="/api/v1/products",status="200"} 1234
api_requests_total{method="POST",endpoint="/api/v1/sales",status="201"} 567

# HELP api_request_duration_seconds API request duration in seconds
# TYPE api_request_duration_seconds histogram
api_request_duration_seconds_bucket{method="GET",endpoint="/api/v1/products",le="0.1"} 890
api_request_duration_seconds_bucket{method="GET",endpoint="/api/v1/products",le="0.5"} 1200
api_request_duration_seconds_bucket{method="GET",endpoint="/api/v1/products",le="1.0"} 1230
api_request_duration_seconds_bucket{method="GET",endpoint="/api/v1/products",le="+Inf"} 1234

# HELP database_connections_active Number of active database connections
# TYPE database_connections_active gauge
database_connections_active 8

# HELP sales_total Total number of sales processed
# TYPE sales_total counter
sales_total{sucursal="test-sucursal-1"} 1567
sales_total{sucursal="test-sucursal-2"} 892
```

## Códigos de Error

### Códigos de Estado HTTP

La API utiliza códigos de estado HTTP estándar para indicar el resultado de las operaciones:

- **200 OK**: Operación exitosa
- **201 Created**: Recurso creado exitosamente
- **400 Bad Request**: Error en los datos enviados
- **401 Unauthorized**: Autenticación requerida o inválida
- **403 Forbidden**: Permisos insuficientes
- **404 Not Found**: Recurso no encontrado
- **409 Conflict**: Conflicto con el estado actual del recurso
- **422 Unprocessable Entity**: Datos válidos pero lógicamente incorrectos
- **429 Too Many Requests**: Rate limit excedido
- **500 Internal Server Error**: Error interno del servidor
- **503 Service Unavailable**: Servicio temporalmente no disponible

### Estructura de Errores

Todos los errores siguen una estructura consistente:

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Los datos proporcionados no son válidos",
    "details": {
      "field": "email",
      "reason": "Formato de email inválido",
      "value": "email-invalido"
    }
  },
  "request_id": "req_123456811",
  "timestamp": "2025-01-08T12:40:00Z"
}
```

### Códigos de Error Específicos

#### Autenticación y Autorización
- **AUTH_REQUIRED**: Token de autenticación requerido
- **AUTH_INVALID**: Token de autenticación inválido o expirado
- **AUTH_INSUFFICIENT_PERMISSIONS**: Permisos insuficientes para la operación
- **AUTH_USER_INACTIVE**: Usuario inactivo o suspendido
- **AUTH_TERMINAL_UNAUTHORIZED**: Terminal no autorizada para el usuario

#### Validación de Datos
- **VALIDATION_ERROR**: Error general de validación
- **VALIDATION_REQUIRED_FIELD**: Campo requerido faltante
- **VALIDATION_INVALID_FORMAT**: Formato de datos inválido
- **VALIDATION_INVALID_VALUE**: Valor fuera del rango permitido
- **VALIDATION_DUPLICATE_VALUE**: Valor duplicado no permitido

#### Productos y Stock
- **PRODUCT_NOT_FOUND**: Producto no encontrado
- **PRODUCT_INACTIVE**: Producto inactivo
- **STOCK_INSUFFICIENT**: Stock insuficiente para la operación
- **STOCK_RESERVED**: Stock reservado no disponible
- **BARCODE_INVALID**: Código de barras inválido
- **BARCODE_DUPLICATE**: Código de barras ya existe

#### Ventas
- **SALE_INVALID_TOTAL**: Total de venta no coincide con items
- **SALE_PAYMENT_MISMATCH**: Medios de pago no coinciden con total
- **SALE_TERMINAL_OFFLINE**: Terminal fuera de línea
- **SALE_ALREADY_PROCESSED**: Venta ya procesada
- **REFUND_NOT_ALLOWED**: Devolución no permitida
- **REFUND_AMOUNT_EXCEEDED**: Monto de devolución excede el original

#### Sistema
- **RATE_LIMIT_EXCEEDED**: Límite de requests excedido
- **DATABASE_ERROR**: Error de base de datos
- **EXTERNAL_SERVICE_ERROR**: Error en servicio externo
- **MAINTENANCE_MODE**: Sistema en modo mantenimiento

### Ejemplos de Respuestas de Error

#### Error de Validación (400)
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Error de validación en los datos enviados",
    "details": {
      "errors": [
        {
          "field": "email",
          "message": "El formato del email es inválido",
          "value": "email-invalido"
        },
        {
          "field": "password",
          "message": "La contraseña debe tener al menos 8 caracteres",
          "value": "***"
        }
      ]
    }
  },
  "request_id": "req_123456812",
  "timestamp": "2025-01-08T12:45:00Z"
}
```

#### Error de Autenticación (401)
```json
{
  "success": false,
  "error": {
    "code": "AUTH_INVALID",
    "message": "Token de autenticación inválido o expirado",
    "details": {
      "reason": "Token expirado",
      "expired_at": "2025-01-08T11:30:00Z"
    }
  },
  "request_id": "req_123456813",
  "timestamp": "2025-01-08T12:45:00Z"
}
```

#### Error de Stock (422)
```json
{
  "success": false,
  "error": {
    "code": "STOCK_INSUFFICIENT",
    "message": "Stock insuficiente para completar la operación",
    "details": {
      "producto_id": "test-product-1",
      "producto_codigo": "TEST001",
      "stock_disponible": 5,
      "cantidad_solicitada": 10,
      "sucursal_id": "test-sucursal-1"
    }
  },
  "request_id": "req_123456814",
  "timestamp": "2025-01-08T12:45:00Z"
}
```

#### Error de Rate Limit (429)
```json
{
  "success": false,
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Límite de requests por minuto excedido",
    "details": {
      "limit": 100,
      "window": "1m",
      "retry_after": 45
    }
  },
  "request_id": "req_123456815",
  "timestamp": "2025-01-08T12:45:00Z"
}
```

