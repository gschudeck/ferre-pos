# API Labels - Documentación Técnica

**Sistema FERRE-POS - Servidor Central**  
**Versión**: 1.0.0  
**Puerto**: 8082  
**Autor**: Manus AI  
**Fecha**: Enero 2025

---

## Tabla de Contenidos

1. [Introducción](#introducción)
2. [Autenticación y Permisos](#autenticación-y-permisos)
3. [Gestión de Plantillas](#gestión-de-plantillas)
4. [Generación de Etiquetas](#generación-de-etiquetas)
5. [Códigos de Barras](#códigos-de-barras)
6. [Trabajos de Impresión](#trabajos-de-impresión)
7. [Previsualización](#previsualización)
8. [Formatos de Salida](#formatos-de-salida)
9. [Monitoreo y Métricas](#monitoreo-y-métricas)
10. [Códigos de Error](#códigos-de-error)
11. [Guías de Integración](#guías-de-integración)
12. [Referencias](#referencias)

---

## Introducción

La API Labels es el componente especializado del sistema FERRE-POS dedicado a la generación, gestión y impresión de etiquetas de productos y códigos de barras. Esta API está diseñada para manejar las necesidades complejas de etiquetado en entornos de retail, desde etiquetas simples de precio hasta etiquetas promocionales complejas con múltiples elementos gráficos y de texto.

El sistema de etiquetas soporta múltiples formatos de salida incluyendo PDF para impresión en impresoras láser, PNG para visualización digital, y ZPL (Zebra Programming Language) para impresoras térmicas especializadas. La API está optimizada para generar grandes volúmenes de etiquetas de manera eficiente, utilizando técnicas de procesamiento en lotes y generación asíncrona para mantener tiempos de respuesta óptimos.

La arquitectura del sistema permite la personalización completa de las etiquetas a través de plantillas configurables que definen la posición, formato y contenido de cada elemento. Las plantillas soportan elementos dinámicos que se populan automáticamente con datos de productos, precios promocionales, códigos de barras, y otra información relevante del inventario.

### Características Principales

La API Labels incluye un motor de plantillas avanzado que permite la creación de diseños complejos con elementos posicionados con precisión milimétrica. El sistema soporta múltiples tipos de elementos incluyendo texto con formateo completo, imágenes, códigos de barras en diversos formatos, líneas, rectángulos, y elementos gráficos personalizados.

El sistema de códigos de barras es particularmente robusto, soportando los formatos más comunes en retail incluyendo EAN-13, UPC-A, Code 128, Code 39, y QR codes. Cada código de barras se genera con verificación automática de dígitos de control y optimización para diferentes tipos de impresoras y resoluciones.

La gestión de trabajos de impresión permite el seguimiento completo del ciclo de vida de cada trabajo, desde la solicitud inicial hasta la confirmación de impresión. El sistema mantiene un historial detallado de todos los trabajos para auditoría y reimpresión cuando sea necesario.

### Arquitectura y Tecnologías

La API está construida utilizando Go con librerías especializadas para generación de gráficos y códigos de barras. El motor de renderizado utiliza bibliotecas optimizadas para generar salidas de alta calidad en diferentes formatos, garantizando que las etiquetas se vean consistentes independientemente del método de salida.

Para el procesamiento de grandes volúmenes, la API implementa un sistema de colas que permite el procesamiento asíncrono de trabajos de etiquetado. Esto es especialmente importante para operaciones como la generación masiva de etiquetas para inventarios completos o campañas promocionales.

El sistema de almacenamiento está optimizado para manejar tanto las plantillas de diseño como las etiquetas generadas, con políticas de retención configurables que balancean el espacio de almacenamiento con la necesidad de mantener historiales para reimpresión.


## Autenticación y Permisos

### Sistema de Autenticación Especializado

La API Labels utiliza el mismo sistema de autenticación JWT que las otras APIs del sistema FERRE-POS, pero implementa un conjunto específico de permisos relacionados con la gestión de etiquetas y códigos de barras. Los permisos están diseñados para permitir diferentes niveles de acceso según el rol del usuario y las necesidades operativas.

Los operadores de etiquetas tienen permisos completos para generar etiquetas y gestionar trabajos de impresión, pero acceso limitado a la modificación de plantillas. Los supervisores y administradores pueden crear y modificar plantillas, mientras que los vendedores y cajeros pueden generar etiquetas usando plantillas existentes pero no pueden modificar configuraciones.

El sistema incluye permisos granulares para diferentes tipos de operaciones: generación de etiquetas individuales, generación masiva, gestión de plantillas, configuración de impresoras, y acceso a historiales de trabajos. Esto permite una gestión flexible de accesos según las responsabilidades específicas de cada usuario.

### Permisos Específicos de Labels

- **labels.generate**: Generar etiquetas usando plantillas existentes
- **labels.generate.bulk**: Generar etiquetas en lotes grandes
- **labels.templates.view**: Ver plantillas de etiquetas
- **labels.templates.create**: Crear nuevas plantillas
- **labels.templates.edit**: Modificar plantillas existentes
- **labels.templates.delete**: Eliminar plantillas
- **labels.jobs.view**: Ver trabajos de impresión
- **labels.jobs.manage**: Gestionar trabajos de impresión
- **labels.printers.configure**: Configurar impresoras
- **labels.history.view**: Ver historial de etiquetas generadas

## Gestión de Plantillas

### Sistema de Plantillas Avanzado

Las plantillas de etiquetas en FERRE-POS son estructuras JSON complejas que definen todos los aspectos visuales y de contenido de las etiquetas. Cada plantilla especifica las dimensiones físicas de la etiqueta, la posición exacta de cada elemento, y las reglas de formateo para el contenido dinámico.

El sistema de plantillas soporta herencia y composición, permitiendo que las plantillas compartan elementos comunes mientras mantienen características específicas. Esto facilita la gestión de múltiples tipos de etiquetas que comparten elementos como logos corporativos o información de contacto.

Las plantillas incluyen validación automática que verifica que todos los elementos encajen dentro de las dimensiones especificadas y que no haya superposiciones no deseadas. El sistema también proporciona advertencias cuando los elementos están muy cerca de los bordes o cuando el texto podría ser demasiado pequeño para ser legible.

#### GET /api/v1/templates

Obtiene una lista de plantillas de etiquetas disponibles.

**Permisos Requeridos:** labels.templates.view

**Headers:**
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
Content-Type: application/json
```

**Query Parameters:**
- `page` (int, opcional): Número de página (default: 1)
- `per_page` (int, opcional): Registros por página (default: 20, max: 100)
- `search` (string, opcional): Búsqueda por nombre o descripción
- `activas` (bool, opcional): Solo plantillas activas (default: true)
- `categoria` (string, opcional): Filtrar por categoría de plantilla
- `tamaño` (string, opcional): Filtrar por tamaño de etiqueta

**Response (200 OK):**
```json
{
  "success": true,
  "data": [
    {
      "id": "template-basic",
      "nombre": "Plantilla Básica Test",
      "descripcion": "Plantilla básica para tests con código y precio",
      "categoria": "basica",
      "dimensiones": {
        "ancho_mm": 50.0,
        "alto_mm": 30.0,
        "dpi": 300
      },
      "elementos_count": 4,
      "activa": true,
      "preview_url": "/api/v1/templates/template-basic/preview",
      "uso_estadisticas": {
        "etiquetas_generadas_mes": 1250,
        "ultimo_uso": "2025-01-08T14:30:00Z"
      },
      "created_at": "2025-01-01T00:00:00Z",
      "updated_at": "2025-01-05T10:00:00Z"
    },
    {
      "id": "template-advanced",
      "nombre": "Plantilla Avanzada Test",
      "descripcion": "Plantilla avanzada para tests con más elementos",
      "categoria": "avanzada",
      "dimensiones": {
        "ancho_mm": 70.0,
        "alto_mm": 40.0,
        "dpi": 300
      },
      "elementos_count": 5,
      "activa": true,
      "preview_url": "/api/v1/templates/template-advanced/preview",
      "uso_estadisticas": {
        "etiquetas_generadas_mes": 850,
        "ultimo_uso": "2025-01-08T13:15:00Z"
      },
      "created_at": "2025-01-01T00:00:00Z",
      "updated_at": "2025-01-03T15:30:00Z"
    }
  ],
  "meta": {
    "page": 1,
    "per_page": 20,
    "total": 2,
    "total_pages": 1,
    "categorias_disponibles": ["basica", "avanzada", "promocional", "inventario"],
    "tamaños_comunes": ["50x30mm", "70x40mm", "100x60mm"]
  },
  "request_id": "labels_req_001",
  "timestamp": "2025-01-08T15:00:00Z"
}
```

#### GET /api/v1/templates/{id}

Obtiene los detalles completos de una plantilla específica.

**Permisos Requeridos:** labels.templates.view

**Path Parameters:**
- `id` (string, requerido): ID de la plantilla

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "id": "template-basic",
    "nombre": "Plantilla Básica Test",
    "descripcion": "Plantilla básica para tests con código y precio",
    "categoria": "basica",
    "dimensiones": {
      "ancho_mm": 50.0,
      "alto_mm": 30.0,
      "dpi": 300,
      "orientacion": "horizontal"
    },
    "elementos": [
      {
        "id": "elemento_nombre",
        "tipo": "text",
        "campo_datos": "nombre",
        "posicion": {
          "x_mm": 2.0,
          "y_mm": 2.0,
          "ancho_mm": 46.0,
          "alto_mm": 6.0
        },
        "formato": {
          "fuente": "Arial",
          "tamaño_pt": 8,
          "negrita": true,
          "cursiva": false,
          "color": "#000000",
          "alineacion": "left",
          "ajuste_texto": "truncate"
        },
        "validaciones": {
          "max_caracteres": 50,
          "requerido": true
        }
      },
      {
        "id": "elemento_codigo",
        "tipo": "text",
        "campo_datos": "codigo",
        "posicion": {
          "x_mm": 2.0,
          "y_mm": 8.0,
          "ancho_mm": 20.0,
          "alto_mm": 4.0
        },
        "formato": {
          "fuente": "Arial",
          "tamaño_pt": 6,
          "negrita": false,
          "color": "#666666",
          "alineacion": "left"
        }
      },
      {
        "id": "elemento_precio",
        "tipo": "text",
        "campo_datos": "precio",
        "posicion": {
          "x_mm": 2.0,
          "y_mm": 14.0,
          "ancho_mm": 25.0,
          "alto_mm": 6.0
        },
        "formato": {
          "fuente": "Arial",
          "tamaño_pt": 10,
          "negrita": true,
          "color": "#000000",
          "alineacion": "left",
          "formato_numero": "$#,##0"
        }
      },
      {
        "id": "elemento_barcode",
        "tipo": "barcode",
        "campo_datos": "codigo_barras",
        "posicion": {
          "x_mm": 2.0,
          "y_mm": 20.0,
          "ancho_mm": 46.0,
          "alto_mm": 8.0
        },
        "formato": {
          "tipo_codigo": "CODE128",
          "mostrar_texto": true,
          "tamaño_texto_pt": 6,
          "altura_barras_mm": 6.0,
          "margen_mm": 1.0
        },
        "validaciones": {
          "verificar_digito": true,
          "longitud_minima": 8
        }
      }
    ],
    "configuracion": {
      "margen_seguridad_mm": 1.0,
      "resolucion_impresion_dpi": 300,
      "formato_salida_default": "pdf",
      "calidad_imagen": "high"
    },
    "metadatos": {
      "version": "1.2",
      "autor": "test-user-1",
      "etiquetas_generadas_total": 15420,
      "ultima_modificacion": "2025-01-05T10:00:00Z",
      "tags": ["basica", "precio", "codigo_barras"]
    },
    "activa": true,
    "created_at": "2025-01-01T00:00:00Z",
    "updated_at": "2025-01-05T10:00:00Z"
  },
  "request_id": "labels_req_002",
  "timestamp": "2025-01-08T15:05:00Z"
}
```

#### POST /api/v1/templates

Crea una nueva plantilla de etiquetas.

**Permisos Requeridos:** labels.templates.create

**Request Body:**
```json
{
  "nombre": "Plantilla Promocional",
  "descripcion": "Plantilla para etiquetas promocionales con descuentos",
  "categoria": "promocional",
  "dimensiones": {
    "ancho_mm": 60.0,
    "alto_mm": 40.0,
    "dpi": 300,
    "orientacion": "horizontal"
  },
  "elementos": [
    {
      "id": "elemento_nombre",
      "tipo": "text",
      "campo_datos": "nombre",
      "posicion": {
        "x_mm": 2.0,
        "y_mm": 2.0,
        "ancho_mm": 56.0,
        "alto_mm": 8.0
      },
      "formato": {
        "fuente": "Arial",
        "tamaño_pt": 10,
        "negrita": true,
        "color": "#000000",
        "alineacion": "center"
      }
    },
    {
      "id": "elemento_precio_anterior",
      "tipo": "text",
      "campo_datos": "precio_anterior",
      "posicion": {
        "x_mm": 5.0,
        "y_mm": 12.0,
        "ancho_mm": 20.0,
        "alto_mm": 6.0
      },
      "formato": {
        "fuente": "Arial",
        "tamaño_pt": 8,
        "color": "#999999",
        "alineacion": "left",
        "tachado": true,
        "formato_numero": "$#,##0"
      }
    },
    {
      "id": "elemento_precio_oferta",
      "tipo": "text",
      "campo_datos": "precio_oferta",
      "posicion": {
        "x_mm": 30.0,
        "y_mm": 12.0,
        "ancho_mm": 25.0,
        "alto_mm": 8.0
      },
      "formato": {
        "fuente": "Arial",
        "tamaño_pt": 12,
        "negrita": true,
        "color": "#FF0000",
        "alineacion": "right",
        "formato_numero": "$#,##0"
      }
    },
    {
      "id": "elemento_descuento",
      "tipo": "text",
      "campo_datos": "porcentaje_descuento",
      "posicion": {
        "x_mm": 45.0,
        "y_mm": 22.0,
        "ancho_mm": 12.0,
        "alto_mm": 6.0
      },
      "formato": {
        "fuente": "Arial",
        "tamaño_pt": 10,
        "negrita": true,
        "color": "#FFFFFF",
        "fondo_color": "#FF0000",
        "alineacion": "center",
        "formato_numero": "#%"
      }
    },
    {
      "id": "elemento_barcode",
      "tipo": "barcode",
      "campo_datos": "codigo_barras",
      "posicion": {
        "x_mm": 2.0,
        "y_mm": 30.0,
        "ancho_mm": 56.0,
        "alto_mm": 8.0
      },
      "formato": {
        "tipo_codigo": "EAN13",
        "mostrar_texto": true,
        "tamaño_texto_pt": 6
      }
    }
  ],
  "configuracion": {
    "margen_seguridad_mm": 1.5,
    "resolucion_impresion_dpi": 300,
    "formato_salida_default": "pdf"
  },
  "activa": true,
  "tags": ["promocional", "descuento", "oferta"]
}
```

**Validaciones:**
- Dimensiones deben ser positivas y realistas (max 200mm x 200mm)
- Elementos no pueden superponerse
- Todos los elementos deben estar dentro de los límites de la etiqueta
- Campos de datos deben ser válidos
- Formatos de fuente y colores deben ser válidos

**Response (201 Created):**
```json
{
  "success": true,
  "data": {
    "id": "template-promocional-uuid",
    "nombre": "Plantilla Promocional",
    "descripcion": "Plantilla para etiquetas promocionales con descuentos",
    "categoria": "promocional",
    "dimensiones": {
      "ancho_mm": 60.0,
      "alto_mm": 40.0,
      "dpi": 300
    },
    "elementos_count": 5,
    "activa": true,
    "preview_url": "/api/v1/templates/template-promocional-uuid/preview",
    "created_at": "2025-01-08T15:10:00Z",
    "updated_at": "2025-01-08T15:10:00Z"
  },
  "request_id": "labels_req_003",
  "timestamp": "2025-01-08T15:10:00Z"
}
```

#### PUT /api/v1/templates/{id}

Actualiza una plantilla existente.

**Permisos Requeridos:** labels.templates.edit

**Path Parameters:**
- `id` (string, requerido): ID de la plantilla

**Request Body:**
```json
{
  "descripcion": "Plantilla básica actualizada con mejores elementos",
  "elementos": [
    {
      "id": "elemento_nombre",
      "tipo": "text",
      "campo_datos": "nombre",
      "posicion": {
        "x_mm": 2.0,
        "y_mm": 2.0,
        "ancho_mm": 46.0,
        "alto_mm": 6.0
      },
      "formato": {
        "fuente": "Arial",
        "tamaño_pt": 9,
        "negrita": true,
        "color": "#000000",
        "alineacion": "left"
      }
    }
  ]
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "id": "template-basic",
    "nombre": "Plantilla Básica Test",
    "descripcion": "Plantilla básica actualizada con mejores elementos",
    "version": "1.3",
    "updated_at": "2025-01-08T15:15:00Z",
    "cambios_aplicados": [
      "Descripción actualizada",
      "Tamaño de fuente del nombre cambiado de 8pt a 9pt"
    ]
  },
  "request_id": "labels_req_004",
  "timestamp": "2025-01-08T15:15:00Z"
}
```

#### GET /api/v1/templates/{id}/preview

Genera una previsualización de la plantilla con datos de ejemplo.

**Permisos Requeridos:** labels.templates.view

**Path Parameters:**
- `id` (string, requerido): ID de la plantilla

**Query Parameters:**
- `formato` (string, opcional): Formato de salida (png, pdf, svg) - default: png
- `dpi` (int, opcional): Resolución para imágenes (default: 300)
- `datos_ejemplo` (bool, opcional): Usar datos de ejemplo (default: true)

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "preview_url": "/api/v1/templates/template-basic/preview.png",
    "preview_base64": "iVBORw0KGgoAAAANSUhEUgAAASwAAAEsCAYAAAB5fY51...",
    "dimensiones": {
      "ancho_px": 590,
      "alto_px": 354,
      "dpi": 300
    },
    "datos_utilizados": {
      "nombre": "Producto de Ejemplo",
      "codigo": "EJ001",
      "precio": 15000.00,
      "codigo_barras": "1234567890123"
    },
    "elementos_renderizados": 4,
    "tiempo_generacion_ms": 125
  },
  "request_id": "labels_req_005",
  "timestamp": "2025-01-08T15:20:00Z"
}
```


## Generación de Etiquetas

### Sistema de Generación Masiva

La generación de etiquetas en FERRE-POS está optimizada para manejar tanto solicitudes individuales como generación masiva de miles de etiquetas. El sistema utiliza procesamiento asíncrono para trabajos grandes, permitiendo que los usuarios continúen con otras tareas mientras las etiquetas se generan en segundo plano.

El motor de generación soporta múltiples estrategias de optimización incluyendo reutilización de elementos comunes, compresión de imágenes, y generación paralela para maximizar el rendimiento. Para trabajos muy grandes, el sistema puede dividir la generación en lotes más pequeños para evitar timeouts y permitir un mejor control del progreso.

La generación incluye validación automática de todos los datos antes del procesamiento, verificando que los productos existan, que los precios estén actualizados, y que todos los campos requeridos estén presentes. Esto previene errores en las etiquetas generadas y garantiza la calidad de la salida.

#### POST /api/v1/labels/generate

Genera etiquetas para uno o múltiples productos.

**Permisos Requeridos:** labels.generate

**Request Body:**
```json
{
  "plantilla_id": "template-basic",
  "productos": [
    {
      "producto_id": "test-product-1",
      "cantidad": 10,
      "datos_personalizados": {
        "precio_promocional": 14000.00,
        "fecha_promocion": "2025-01-15"
      }
    },
    {
      "producto_id": "test-product-2",
      "cantidad": 25
    }
  ],
  "opciones": {
    "formato_salida": "pdf",
    "dpi": 300,
    "incluir_margenes_corte": true,
    "etiquetas_por_hoja": 12,
    "orientacion_hoja": "portrait"
  },
  "configuracion_trabajo": {
    "nombre_trabajo": "Etiquetas Productos Herramientas",
    "prioridad": "normal",
    "notificar_completado": true,
    "email_notificacion": "operador@ferrepos.com"
  }
}
```

**Response (202 Accepted):**
```json
{
  "success": true,
  "data": {
    "trabajo_id": "job_labels_001",
    "estado": "en_proceso",
    "total_etiquetas": 35,
    "productos_procesados": 2,
    "tiempo_estimado_minutos": 2,
    "progreso": {
      "porcentaje": 0,
      "etiquetas_generadas": 0,
      "etiquetas_pendientes": 35,
      "fase_actual": "validacion_datos"
    },
    "urls": {
      "estado": "/api/v1/labels/jobs/job_labels_001/status",
      "cancelar": "/api/v1/labels/jobs/job_labels_001/cancel",
      "descargar": "/api/v1/labels/jobs/job_labels_001/download"
    },
    "created_at": "2025-01-08T15:25:00Z"
  },
  "request_id": "labels_req_006",
  "timestamp": "2025-01-08T15:25:00Z"
}
```

#### POST /api/v1/labels/generate/single

Genera una etiqueta individual de forma síncrona.

**Permisos Requeridos:** labels.generate

**Request Body:**
```json
{
  "plantilla_id": "template-basic",
  "producto_id": "test-product-1",
  "datos_personalizados": {
    "precio_promocional": 14000.00,
    "texto_adicional": "¡OFERTA!"
  },
  "formato_salida": "png",
  "dpi": 300
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "etiqueta_id": "label_single_001",
    "formato": "png",
    "dimensiones": {
      "ancho_px": 590,
      "alto_px": 354,
      "dpi": 300
    },
    "archivo": {
      "url": "/api/v1/labels/download/label_single_001.png",
      "base64": "iVBORw0KGgoAAAANSUhEUgAAASwAAAEsCAYAAAB5fY51...",
      "tamaño_bytes": 45280,
      "checksum": "sha256:abcdef123456..."
    },
    "datos_utilizados": {
      "nombre": "Martillo Test 500g",
      "codigo": "TEST001",
      "precio": 14000.00,
      "codigo_barras": "1234567890123",
      "texto_adicional": "¡OFERTA!"
    },
    "tiempo_generacion_ms": 85,
    "expires_at": "2025-01-09T15:25:00Z"
  },
  "request_id": "labels_req_007",
  "timestamp": "2025-01-08T15:25:00Z"
}
```

#### POST /api/v1/labels/generate/bulk

Genera etiquetas masivas con configuración avanzada.

**Permisos Requeridos:** labels.generate.bulk

**Request Body:**
```json
{
  "plantilla_id": "template-basic",
  "filtros_productos": {
    "categoria_ids": ["test-category-1"],
    "sucursal_id": "test-sucursal-1",
    "activos": true,
    "con_stock": true,
    "precio_min": 1000.00,
    "precio_max": 100000.00
  },
  "configuracion_etiquetas": {
    "cantidad_por_producto": 5,
    "incluir_productos_sin_codigo_barras": false,
    "aplicar_precios_promocionales": true
  },
  "opciones_salida": {
    "formato": "pdf",
    "etiquetas_por_hoja": 12,
    "orientacion": "portrait",
    "incluir_margenes_corte": true,
    "calidad": "high"
  },
  "configuracion_trabajo": {
    "nombre": "Etiquetas Masivas Herramientas",
    "descripcion": "Generación masiva para reposición de etiquetas",
    "prioridad": "high",
    "procesar_en_lotes": true,
    "tamaño_lote": 100
  }
}
```

**Response (202 Accepted):**
```json
{
  "success": true,
  "data": {
    "trabajo_id": "job_bulk_001",
    "estado": "iniciado",
    "productos_encontrados": 8,
    "total_etiquetas_estimadas": 40,
    "lotes_planificados": 1,
    "tiempo_estimado_minutos": 5,
    "configuracion_aplicada": {
      "formato": "pdf",
      "etiquetas_por_hoja": 12,
      "cantidad_por_producto": 5
    },
    "progreso_url": "/api/v1/labels/jobs/job_bulk_001/progress",
    "created_at": "2025-01-08T15:30:00Z"
  },
  "request_id": "labels_req_008",
  "timestamp": "2025-01-08T15:30:00Z"
}
```

## Códigos de Barras

### Generación y Validación de Códigos

El sistema de códigos de barras de FERRE-POS soporta los formatos más utilizados en retail y puede generar códigos tanto como parte de etiquetas completas como elementos independientes. El sistema incluye validación automática de dígitos de control, verificación de longitud, y optimización para diferentes tipos de impresoras.

La generación de códigos de barras incluye opciones avanzadas como ajuste automático del tamaño según la resolución de salida, configuración de márgenes de seguridad, y optimización de la relación ancho/alto de las barras para maximizar la legibilidad en diferentes condiciones de escaneo.

El sistema mantiene un registro de todos los códigos de barras generados para evitar duplicaciones y proporciona herramientas de validación que pueden verificar códigos existentes contra estándares internacionales.

#### POST /api/v1/barcodes/generate

Genera códigos de barras independientes.

**Permisos Requeridos:** labels.generate

**Request Body:**
```json
{
  "codigos": [
    {
      "valor": "1234567890123",
      "tipo": "EAN13",
      "formato_salida": "png"
    },
    {
      "valor": "TEST001",
      "tipo": "CODE128",
      "formato_salida": "svg"
    }
  ],
  "opciones": {
    "ancho_mm": 40.0,
    "alto_mm": 15.0,
    "dpi": 300,
    "mostrar_texto": true,
    "tamaño_texto_pt": 8,
    "margen_mm": 2.0,
    "color_barras": "#000000",
    "color_fondo": "#FFFFFF"
  }
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "codigos_generados": [
      {
        "valor": "1234567890123",
        "tipo": "EAN13",
        "formato": "png",
        "valido": true,
        "digito_control": "3",
        "archivo": {
          "url": "/api/v1/barcodes/download/ean13_1234567890123.png",
          "base64": "iVBORw0KGgoAAAANSUhEUgAAASwAAAEsCAYAAAB5fY51...",
          "tamaño_bytes": 2048
        },
        "dimensiones": {
          "ancho_px": 472,
          "alto_px": 177,
          "dpi": 300
        }
      },
      {
        "valor": "TEST001",
        "tipo": "CODE128",
        "formato": "svg",
        "valido": true,
        "archivo": {
          "url": "/api/v1/barcodes/download/code128_TEST001.svg",
          "contenido_svg": "<svg xmlns=\"http://www.w3.org/2000/svg\"...",
          "tamaño_bytes": 1024
        },
        "dimensiones": {
          "ancho_mm": 40.0,
          "alto_mm": 15.0
        }
      }
    ],
    "resumen": {
      "total_solicitados": 2,
      "generados_exitosamente": 2,
      "errores": 0,
      "tiempo_total_ms": 156
    }
  },
  "request_id": "labels_req_009",
  "timestamp": "2025-01-08T15:35:00Z"
}
```

#### POST /api/v1/barcodes/validate

Valida códigos de barras según estándares internacionales.

**Permisos Requeridos:** labels.generate

**Request Body:**
```json
{
  "codigos": [
    {
      "valor": "1234567890123",
      "tipo": "EAN13"
    },
    {
      "valor": "INVALID123",
      "tipo": "EAN13"
    },
    {
      "valor": "TEST001",
      "tipo": "CODE128"
    }
  ]
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "validaciones": [
      {
        "valor": "1234567890123",
        "tipo": "EAN13",
        "valido": true,
        "digito_control_calculado": "3",
        "digito_control_proporcionado": "3",
        "longitud_correcta": true,
        "formato_valido": true
      },
      {
        "valor": "INVALID123",
        "tipo": "EAN13",
        "valido": false,
        "errores": [
          "Longitud incorrecta: esperado 13, recibido 10",
          "Contiene caracteres no numéricos"
        ],
        "longitud_correcta": false,
        "formato_valido": false
      },
      {
        "valor": "TEST001",
        "tipo": "CODE128",
        "valido": true,
        "longitud_correcta": true,
        "formato_valido": true,
        "caracteres_soportados": true
      }
    ],
    "resumen": {
      "total_validados": 3,
      "validos": 2,
      "invalidos": 1,
      "tasa_validez": 66.7
    }
  },
  "request_id": "labels_req_010",
  "timestamp": "2025-01-08T15:40:00Z"
}
```

#### GET /api/v1/barcodes/formats

Obtiene información sobre los formatos de códigos de barras soportados.

**Permisos Requeridos:** labels.generate

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "formatos_soportados": [
      {
        "codigo": "EAN13",
        "nombre": "European Article Number 13",
        "descripcion": "Código de barras estándar para productos de consumo",
        "longitud_fija": true,
        "longitud_caracteres": 13,
        "tipo_caracteres": "numerico",
        "incluye_digito_control": true,
        "uso_recomendado": "Productos de retail",
        "ejemplo": "1234567890123"
      },
      {
        "codigo": "EAN8",
        "nombre": "European Article Number 8",
        "descripcion": "Versión compacta de EAN13 para productos pequeños",
        "longitud_fija": true,
        "longitud_caracteres": 8,
        "tipo_caracteres": "numerico",
        "incluye_digito_control": true,
        "uso_recomendado": "Productos pequeños",
        "ejemplo": "12345678"
      },
      {
        "codigo": "CODE128",
        "nombre": "Code 128",
        "descripcion": "Código alfanumérico de alta densidad",
        "longitud_fija": false,
        "longitud_minima": 1,
        "longitud_maxima": 80,
        "tipo_caracteres": "alfanumerico",
        "incluye_digito_control": true,
        "uso_recomendado": "Códigos internos, inventario",
        "ejemplo": "TEST001"
      },
      {
        "codigo": "CODE39",
        "nombre": "Code 39",
        "descripcion": "Código alfanumérico ampliamente compatible",
        "longitud_fija": false,
        "longitud_minima": 1,
        "longitud_maxima": 43,
        "tipo_caracteres": "alfanumerico_limitado",
        "incluye_digito_control": false,
        "uso_recomendado": "Aplicaciones industriales",
        "ejemplo": "ABC123"
      },
      {
        "codigo": "QR",
        "nombre": "QR Code",
        "descripcion": "Código bidimensional de alta capacidad",
        "longitud_fija": false,
        "longitud_maxima": 4296,
        "tipo_caracteres": "cualquiera",
        "incluye_correccion_errores": true,
        "uso_recomendado": "URLs, información extendida",
        "ejemplo": "https://ferrepos.com/producto/TEST001"
      }
    ],
    "configuraciones_recomendadas": {
      "retail_general": "EAN13",
      "productos_pequeños": "EAN8",
      "inventario_interno": "CODE128",
      "informacion_extendida": "QR"
    }
  },
  "request_id": "labels_req_011",
  "timestamp": "2025-01-08T15:45:00Z"
}
```

## Trabajos de Impresión

### Gestión de Trabajos Asíncronos

El sistema de trabajos de impresión permite el seguimiento completo del ciclo de vida de cada solicitud de generación de etiquetas, desde la creación inicial hasta la finalización y descarga. Los trabajos se procesan de forma asíncrona para permitir que los usuarios continúen con otras tareas mientras se generan las etiquetas.

Cada trabajo incluye información detallada sobre el progreso, estimaciones de tiempo, y notificaciones automáticas cuando se completa. El sistema mantiene un historial completo de todos los trabajos para auditoría y permite la reimpresión de trabajos anteriores sin necesidad de regenerar las etiquetas.

#### GET /api/v1/labels/jobs

Obtiene una lista de trabajos de etiquetas.

**Permisos Requeridos:** labels.jobs.view

**Query Parameters:**
- `page` (int, opcional): Número de página (default: 1)
- `per_page` (int, opcional): Registros por página (default: 20, max: 100)
- `estado` (string, opcional): Filtrar por estado (pendiente, en_proceso, completado, error, cancelado)
- `usuario_id` (uuid, opcional): Filtrar por usuario que creó el trabajo
- `fecha_desde` (date, opcional): Trabajos desde fecha específica
- `fecha_hasta` (date, opcional): Trabajos hasta fecha específica
- `plantilla_id` (string, opcional): Filtrar por plantilla utilizada

**Response (200 OK):**
```json
{
  "success": true,
  "data": [
    {
      "id": "job_labels_001",
      "nombre": "Etiquetas Productos Herramientas",
      "estado": "completado",
      "progreso": {
        "porcentaje": 100,
        "etiquetas_generadas": 35,
        "etiquetas_totales": 35
      },
      "plantilla": {
        "id": "template-basic",
        "nombre": "Plantilla Básica Test"
      },
      "usuario": {
        "id": "test-user-2",
        "nombre": "Vendedor Test"
      },
      "estadisticas": {
        "productos_procesados": 2,
        "tiempo_total_segundos": 125,
        "tamaño_archivo_mb": 2.5
      },
      "archivo_resultado": {
        "url": "/api/v1/labels/jobs/job_labels_001/download",
        "formato": "pdf",
        "tamaño_bytes": 2621440,
        "expires_at": "2025-01-15T15:25:00Z"
      },
      "created_at": "2025-01-08T15:25:00Z",
      "completed_at": "2025-01-08T15:27:05Z"
    }
  ],
  "meta": {
    "page": 1,
    "per_page": 20,
    "total": 1,
    "total_pages": 1,
    "estadisticas_periodo": {
      "trabajos_completados": 15,
      "trabajos_en_proceso": 2,
      "trabajos_con_error": 1,
      "tiempo_promedio_segundos": 145
    }
  },
  "request_id": "labels_req_012",
  "timestamp": "2025-01-08T15:50:00Z"
}
```

#### GET /api/v1/labels/jobs/{job_id}/status

Obtiene el estado detallado de un trabajo específico.

**Permisos Requeridos:** labels.jobs.view

**Path Parameters:**
- `job_id` (string, requerido): ID del trabajo

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "id": "job_bulk_001",
    "nombre": "Etiquetas Masivas Herramientas",
    "estado": "en_proceso",
    "progreso": {
      "porcentaje": 65,
      "fase_actual": "generacion_etiquetas",
      "etiquetas_generadas": 26,
      "etiquetas_totales": 40,
      "lote_actual": 1,
      "lotes_totales": 1,
      "tiempo_transcurrido_segundos": 195,
      "tiempo_estimado_restante_segundos": 105
    },
    "detalles_procesamiento": {
      "productos_procesados": 5,
      "productos_totales": 8,
      "producto_actual": {
        "id": "test-product-6",
        "codigo": "TEST006",
        "nombre": "Taladro Test 600W",
        "etiquetas_generadas": 3,
        "etiquetas_objetivo": 5
      },
      "errores_encontrados": 0,
      "advertencias": [
        {
          "producto_id": "test-product-3",
          "mensaje": "Código de barras muy largo para el espacio asignado",
          "nivel": "warning"
        }
      ]
    },
    "configuracion": {
      "plantilla_id": "template-basic",
      "formato_salida": "pdf",
      "etiquetas_por_hoja": 12,
      "cantidad_por_producto": 5
    },
    "recursos": {
      "cpu_usage": 45.2,
      "memory_usage_mb": 256,
      "temp_files_count": 8,
      "temp_storage_mb": 15.8
    },
    "created_at": "2025-01-08T15:30:00Z",
    "started_at": "2025-01-08T15:30:15Z",
    "estimated_completion": "2025-01-08T15:33:00Z"
  },
  "request_id": "labels_req_013",
  "timestamp": "2025-01-08T15:53:00Z"
}
```

#### POST /api/v1/labels/jobs/{job_id}/cancel

Cancela un trabajo en proceso.

**Permisos Requeridos:** labels.jobs.manage

**Path Parameters:**
- `job_id` (string, requerido): ID del trabajo

**Request Body:**
```json
{
  "motivo": "Cambio en los requerimientos del usuario",
  "forzar_cancelacion": false
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "job_id": "job_bulk_001",
    "estado_anterior": "en_proceso",
    "estado_actual": "cancelado",
    "cancelado_at": "2025-01-08T15:55:00Z",
    "progreso_al_cancelar": {
      "porcentaje": 65,
      "etiquetas_generadas": 26,
      "etiquetas_totales": 40
    },
    "recursos_liberados": {
      "temp_files_eliminados": 8,
      "memoria_liberada_mb": 256,
      "espacio_disco_liberado_mb": 15.8
    },
    "archivo_parcial": {
      "disponible": true,
      "url": "/api/v1/labels/jobs/job_bulk_001/download?partial=true",
      "etiquetas_incluidas": 26,
      "expires_at": "2025-01-09T15:55:00Z"
    }
  },
  "request_id": "labels_req_014",
  "timestamp": "2025-01-08T15:55:00Z"
}
```

#### GET /api/v1/labels/jobs/{job_id}/download

Descarga el archivo resultado de un trabajo completado.

**Permisos Requeridos:** labels.jobs.view

**Path Parameters:**
- `job_id` (string, requerido): ID del trabajo

**Query Parameters:**
- `partial` (bool, opcional): Descargar resultado parcial si el trabajo fue cancelado
- `formato` (string, opcional): Formato alternativo si está disponible

**Response (200 OK):**
- Content-Type: application/pdf (o el formato correspondiente)
- Content-Disposition: attachment; filename="etiquetas_job_labels_001.pdf"
- Content-Length: 2621440

**Headers de Respuesta:**
```
X-Job-ID: job_labels_001
X-Labels-Count: 35
X-Generation-Time: 125
X-File-Checksum: sha256:abcdef123456...
```

