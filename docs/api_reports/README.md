# API Reports - Documentación Técnica

**Sistema FERRE-POS - Servidor Central**  
**Versión**: 1.0.0  
**Puerto**: 8083  
**Autor**: Manus AI  
**Fecha**: Enero 2025

---

## Tabla de Contenidos

1. [Introducción](#introducción)
2. [Autenticación y Permisos](#autenticación-y-permisos)
3. [Reportes de Ventas](#reportes-de-ventas)
4. [Reportes de Inventario](#reportes-de-inventario)
5. [Reportes Financieros](#reportes-financieros)
6. [Analytics y Dashboards](#analytics-y-dashboards)
7. [Exportación de Datos](#exportación-de-datos)
8. [Reportes Personalizados](#reportes-personalizados)
9. [Programación de Reportes](#programación-de-reportes)
10. [Monitoreo y Métricas](#monitoreo-y-métricas)
11. [Códigos de Error](#códigos-de-error)
12. [Guías de Integración](#guías-de-integración)
13. [Referencias](#referencias)

---

## Introducción

La API Reports es el componente analítico central del sistema FERRE-POS, diseñado para proporcionar insights profundos sobre el rendimiento del negocio a través de reportes comprehensivos, dashboards interactivos, y análisis de datos avanzados. Esta API está optimizada para manejar grandes volúmenes de datos transaccionales y generar reportes en tiempo real o programados según las necesidades del negocio.

El sistema de reportes está construido sobre una arquitectura de data warehouse que consolida información de todas las fuentes del sistema FERRE-POS, incluyendo ventas, inventario, usuarios, y operaciones. Los datos se procesan y agregan utilizando técnicas de ETL (Extract, Transform, Load) optimizadas para garantizar que los reportes reflejen información actualizada y precisa.

La API soporta múltiples tipos de visualizaciones incluyendo gráficos de barras, líneas, torta, mapas de calor, y tablas dinámicas. Los reportes pueden exportarse en diversos formatos como PDF, Excel, CSV, y JSON para facilitar la integración con otros sistemas o el análisis offline.

### Características Principales

El sistema de reportes incluye un motor de consultas optimizado que puede procesar millones de registros de transacciones para generar reportes en segundos. Utiliza técnicas de indexación avanzadas, cacheo inteligente, y agregaciones pre-calculadas para mantener tiempos de respuesta óptimos incluso con datasets muy grandes.

Los reportes soportan filtrado dinámico por múltiples dimensiones incluyendo fechas, sucursales, productos, categorías, usuarios, y métodos de pago. Los usuarios pueden crear vistas personalizadas que se guardan automáticamente y pueden compartirse con otros usuarios del sistema.

La API incluye capacidades de análisis predictivo básico que pueden identificar tendencias, patrones estacionales, y anomalías en los datos de ventas. Estos insights ayudan a los gerentes a tomar decisiones informadas sobre inventario, precios, y estrategias de marketing.

### Arquitectura de Datos

El sistema utiliza una arquitectura de data warehouse estrella optimizada para consultas analíticas, con tablas de hechos que contienen las métricas numéricas y tablas de dimensiones que proporcionan el contexto para el análisis. Esta estructura permite consultas complejas que pueden agregar datos a través de múltiples dimensiones de manera eficiente.

Los datos se actualizan en tiempo real para métricas críticas como ventas del día y stock actual, mientras que los reportes históricos utilizan snapshots diarios que se procesan durante horas de baja actividad para no impactar el rendimiento operativo.

El sistema incluye un motor de alertas que puede monitorear métricas clave y enviar notificaciones automáticas cuando se detectan condiciones específicas, como caídas significativas en ventas, stock bajo, o patrones de comportamiento inusuales.


## Autenticación y Permisos

### Sistema de Permisos Granular

La API Reports implementa un sistema de permisos granular que controla el acceso a diferentes tipos de reportes y niveles de información según el rol del usuario y su posición en la organización. Los permisos están diseñados para proteger información sensible mientras permiten que cada usuario acceda a los datos necesarios para su trabajo.

Los administradores tienen acceso completo a todos los reportes incluyendo información financiera detallada y métricas de rendimiento de usuarios. Los supervisores pueden acceder a reportes de su sucursal y comparaciones con otras sucursales, pero no a información financiera consolidada. Los vendedores y cajeros pueden ver reportes básicos de sus propias ventas y rendimiento.

El sistema incluye permisos específicos para diferentes tipos de datos: ventas por monto, márgenes de ganancia, costos de productos, información de clientes, y métricas de rendimiento de empleados. Esto permite una configuración flexible que se adapta a las políticas de privacidad y seguridad de cada organización.

### Permisos Específicos de Reports

- **reports.sales.view**: Ver reportes básicos de ventas
- **reports.sales.detailed**: Ver reportes detallados de ventas con márgenes
- **reports.financial.view**: Ver reportes financieros y de rentabilidad
- **reports.inventory.view**: Ver reportes de inventario y stock
- **reports.users.performance**: Ver reportes de rendimiento de usuarios
- **reports.analytics.advanced**: Acceder a analytics avanzados y predictivos
- **reports.export**: Exportar reportes en diferentes formatos
- **reports.schedule**: Programar reportes automáticos
- **reports.custom.create**: Crear reportes personalizados
- **reports.all_branches**: Ver reportes de todas las sucursales

## Reportes de Ventas

### Análisis Comprehensivo de Ventas

Los reportes de ventas proporcionan una vista completa del rendimiento comercial con múltiples niveles de detalle y agregación. El sistema puede generar reportes desde transacciones individuales hasta resúmenes ejecutivos que cubren períodos extensos y múltiples sucursales.

Los reportes incluyen análisis de tendencias que identifican patrones estacionales, días de mayor actividad, y productos con mejor rendimiento. El sistema puede comparar períodos similares del año anterior para identificar crecimiento o declive en diferentes segmentos del negocio.

Las métricas incluyen no solo volúmenes de venta sino también análisis de márgenes, rentabilidad por producto, efectividad de promociones, y rendimiento por canal de venta. Esto proporciona una vista holística que ayuda a optimizar estrategias comerciales.

#### GET /api/v1/reports/sales/summary

Obtiene un resumen ejecutivo de ventas para un período específico.

**Permisos Requeridos:** reports.sales.view

**Headers:**
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
Content-Type: application/json
```

**Query Parameters:**
- `fecha_inicio` (date, requerido): Fecha de inicio del período (YYYY-MM-DD)
- `fecha_fin` (date, requerido): Fecha de fin del período (YYYY-MM-DD)
- `sucursal_id` (uuid, opcional): Filtrar por sucursal específica
- `comparar_periodo_anterior` (bool, opcional): Incluir comparación con período anterior
- `incluir_devoluciones` (bool, opcional): Incluir devoluciones en el análisis
- `agrupar_por` (string, opcional): Agrupación (dia, semana, mes) - default: dia

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "periodo": {
      "fecha_inicio": "2025-01-01",
      "fecha_fin": "2025-01-08",
      "dias_incluidos": 8,
      "sucursales_incluidas": 2
    },
    "metricas_principales": {
      "total_ventas": 116025.00,
      "cantidad_transacciones": 4,
      "ticket_promedio": 29006.25,
      "productos_vendidos": 36,
      "clientes_unicos": 3,
      "margen_bruto": 48750.00,
      "margen_porcentaje": 42.0
    },
    "comparacion_periodo_anterior": {
      "total_ventas": {
        "actual": 116025.00,
        "anterior": 98500.00,
        "variacion_absoluta": 17525.00,
        "variacion_porcentual": 17.8
      },
      "cantidad_transacciones": {
        "actual": 4,
        "anterior": 5,
        "variacion_absoluta": -1,
        "variacion_porcentual": -20.0
      },
      "ticket_promedio": {
        "actual": 29006.25,
        "anterior": 19700.00,
        "variacion_absoluta": 9306.25,
        "variacion_porcentual": 47.3
      }
    },
    "tendencias": {
      "ventas_por_dia": [
        {
          "fecha": "2025-01-01",
          "ventas": 0.00,
          "transacciones": 0
        },
        {
          "fecha": "2025-01-06",
          "ventas": 29750.00,
          "transacciones": 1
        },
        {
          "fecha": "2025-01-07",
          "ventas": 86275.00,
          "transacciones": 3
        }
      ],
      "mejor_dia": {
        "fecha": "2025-01-07",
        "ventas": 86275.00,
        "transacciones": 3
      },
      "dia_promedio": {
        "ventas": 14503.13,
        "transacciones": 0.5
      }
    },
    "distribucion_por_sucursal": [
      {
        "sucursal_id": "test-sucursal-1",
        "sucursal_nombre": "Test Sucursal Principal",
        "ventas": 116025.00,
        "transacciones": 4,
        "participacion_porcentual": 100.0
      }
    ],
    "top_productos": [
      {
        "producto_id": "test-product-4",
        "codigo": "TEST004",
        "nombre": "Destornillador Test Plano",
        "cantidad_vendida": 15,
        "monto_total": 52500.00,
        "participacion_porcentual": 45.3
      }
    ],
    "metodos_pago": [
      {
        "metodo": "efectivo",
        "monto": 58275.00,
        "transacciones": 2,
        "participacion_porcentual": 50.2
      },
      {
        "metodo": "tarjeta_credito",
        "monto": 57750.00,
        "transacciones": 2,
        "participacion_porcentual": 49.8
      }
    ]
  },
  "request_id": "reports_req_001",
  "timestamp": "2025-01-08T16:00:00Z"
}
```

#### GET /api/v1/reports/sales/detailed

Obtiene un reporte detallado de ventas con información granular.

**Permisos Requeridos:** reports.sales.detailed

**Query Parameters:**
- `fecha_inicio` (date, requerido): Fecha de inicio del período
- `fecha_fin` (date, requerido): Fecha de fin del período
- `sucursal_id` (uuid, opcional): Filtrar por sucursal
- `vendedor_id` (uuid, opcional): Filtrar por vendedor
- `categoria_id` (uuid, opcional): Filtrar por categoría de productos
- `incluir_costos` (bool, opcional): Incluir información de costos y márgenes
- `page` (int, opcional): Número de página para paginación
- `per_page` (int, opcional): Registros por página (max: 1000)

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "ventas": [
      {
        "id": "test-venta-1",
        "numero_documento": "TEST-V-001",
        "fecha_venta": "2025-01-06T10:30:00Z",
        "sucursal": {
          "id": "test-sucursal-1",
          "nombre": "Test Sucursal Principal"
        },
        "vendedor": {
          "id": "test-user-2",
          "nombre": "Vendedor Test"
        },
        "cajero": {
          "id": "test-user-3",
          "nombre": "Cajero Test"
        },
        "cliente": {
          "tipo": "generico",
          "nombre": "Cliente Genérico"
        },
        "totales": {
          "subtotal": 25000.00,
          "descuentos": 0.00,
          "impuestos": 4750.00,
          "total": 29750.00,
          "costo_total": 15000.00,
          "margen_bruto": 14750.00,
          "margen_porcentaje": 49.6
        },
        "items": [
          {
            "producto_id": "test-product-1",
            "codigo": "TEST001",
            "nombre": "Martillo Test 500g",
            "cantidad": 1,
            "precio_unitario": 15000.00,
            "costo_unitario": 8000.00,
            "subtotal": 15000.00,
            "margen_unitario": 7000.00,
            "margen_porcentaje": 46.7
          },
          {
            "producto_id": "test-product-2",
            "codigo": "TEST002",
            "nombre": "Destornillador Test Phillips",
            "cantidad": 2,
            "precio_unitario": 3500.00,
            "costo_unitario": 2000.00,
            "subtotal": 7000.00,
            "margen_unitario": 3000.00,
            "margen_porcentaje": 42.9
          }
        ],
        "metodos_pago": [
          {
            "metodo": "efectivo",
            "monto": 29750.00
          }
        ],
        "metricas": {
          "tiempo_atencion_minutos": 8.5,
          "items_por_transaccion": 3,
          "descuento_aplicado": 0.0
        }
      }
    ],
    "resumen": {
      "total_ventas": 29750.00,
      "total_costo": 15000.00,
      "margen_total": 14750.00,
      "margen_promedio": 49.6,
      "transacciones": 1,
      "items_vendidos": 3,
      "ticket_promedio": 29750.00
    }
  },
  "meta": {
    "page": 1,
    "per_page": 100,
    "total": 1,
    "total_pages": 1
  },
  "request_id": "reports_req_002",
  "timestamp": "2025-01-08T16:05:00Z"
}
```

#### GET /api/v1/reports/sales/trends

Analiza tendencias de ventas con proyecciones y patrones estacionales.

**Permisos Requeridos:** reports.analytics.advanced

**Query Parameters:**
- `periodo` (string, requerido): Período de análisis (mes, trimestre, año)
- `sucursal_id` (uuid, opcional): Filtrar por sucursal
- `incluir_proyeccion` (bool, opcional): Incluir proyecciones futuras
- `analisis_estacional` (bool, opcional): Incluir análisis estacional

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "periodo_analizado": {
      "tipo": "mes",
      "fecha_inicio": "2024-01-01",
      "fecha_fin": "2025-01-08",
      "meses_incluidos": 12
    },
    "tendencia_general": {
      "direccion": "creciente",
      "tasa_crecimiento_mensual": 8.5,
      "confianza": 0.85,
      "r_cuadrado": 0.78
    },
    "ventas_mensuales": [
      {
        "mes": "2024-01",
        "ventas": 450000.00,
        "transacciones": 180,
        "ticket_promedio": 2500.00
      },
      {
        "mes": "2024-12",
        "ventas": 680000.00,
        "transacciones": 245,
        "ticket_promedio": 2775.51
      },
      {
        "mes": "2025-01",
        "ventas": 116025.00,
        "transacciones": 4,
        "ticket_promedio": 29006.25,
        "parcial": true,
        "dias_transcurridos": 8
      }
    ],
    "patrones_estacionales": {
      "meses_altos": ["noviembre", "diciembre", "marzo"],
      "meses_bajos": ["enero", "febrero", "agosto"],
      "variacion_estacional": 35.2,
      "pico_maximo": {
        "mes": "diciembre",
        "factor": 1.45
      },
      "valle_minimo": {
        "mes": "febrero",
        "factor": 0.72
      }
    },
    "proyecciones": {
      "proximo_mes": {
        "ventas_estimadas": 720000.00,
        "rango_confianza": {
          "minimo": 650000.00,
          "maximo": 790000.00
        },
        "confianza": 0.80
      },
      "trimestre": {
        "ventas_estimadas": 2100000.00,
        "crecimiento_esperado": 12.5
      }
    },
    "anomalias_detectadas": [
      {
        "fecha": "2025-01-07",
        "tipo": "pico_ventas",
        "valor_observado": 86275.00,
        "valor_esperado": 25000.00,
        "desviacion": 245.1,
        "posibles_causas": ["promocion_especial", "evento_local"]
      }
    ],
    "recomendaciones": [
      {
        "tipo": "inventario",
        "mensaje": "Incrementar stock para el próximo mes basado en tendencia creciente",
        "prioridad": "alta"
      },
      {
        "tipo": "promociones",
        "mensaje": "Planificar promociones para febrero para contrarrestar estacionalidad baja",
        "prioridad": "media"
      }
    ]
  },
  "request_id": "reports_req_003",
  "timestamp": "2025-01-08T16:10:00Z"
}
```

#### GET /api/v1/reports/sales/performance

Analiza el rendimiento de ventas por diferentes dimensiones.

**Permisos Requeridos:** reports.sales.detailed

**Query Parameters:**
- `fecha_inicio` (date, requerido): Fecha de inicio
- `fecha_fin` (date, requerido): Fecha de fin
- `dimension` (string, requerido): Dimensión de análisis (vendedor, producto, categoria, sucursal, hora)
- `top_n` (int, opcional): Número de elementos top a mostrar (default: 10)
- `incluir_metricas_avanzadas` (bool, opcional): Incluir métricas avanzadas

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "dimension": "vendedor",
    "periodo": {
      "fecha_inicio": "2025-01-01",
      "fecha_fin": "2025-01-08"
    },
    "ranking": [
      {
        "vendedor_id": "test-user-2",
        "vendedor_nombre": "Vendedor Test",
        "metricas": {
          "ventas_total": 116025.00,
          "transacciones": 4,
          "ticket_promedio": 29006.25,
          "productos_vendidos": 36,
          "margen_generado": 48750.00,
          "comision_ganada": 2320.50
        },
        "metricas_avanzadas": {
          "conversion_rate": 85.7,
          "items_por_transaccion": 9.0,
          "tiempo_promedio_venta_minutos": 12.5,
          "satisfaccion_cliente": 4.8,
          "devoluciones_porcentaje": 2.1
        },
        "comparacion_objetivo": {
          "objetivo_mensual": 500000.00,
          "progreso_porcentual": 23.2,
          "proyeccion_mes": 435000.00,
          "probabilidad_cumplimiento": 0.75
        },
        "tendencia": {
          "direccion": "creciente",
          "variacion_semanal": 15.8
        },
        "posicion": 1
      }
    ],
    "estadisticas_generales": {
      "vendedores_activos": 1,
      "promedio_ventas": 116025.00,
      "desviacion_estandar": 0.0,
      "coeficiente_variacion": 0.0,
      "mejor_vendedor": {
        "id": "test-user-2",
        "nombre": "Vendedor Test",
        "ventas": 116025.00
      },
      "distribucion_ventas": {
        "percentil_25": 116025.00,
        "percentil_50": 116025.00,
        "percentil_75": 116025.00,
        "percentil_90": 116025.00
      }
    },
    "insights": [
      {
        "tipo": "oportunidad",
        "mensaje": "Vendedor Test está superando objetivos, considerar aumentar metas",
        "impacto": "alto"
      }
    ]
  },
  "request_id": "reports_req_004",
  "timestamp": "2025-01-08T16:15:00Z"
}
```

## Reportes de Inventario

### Análisis Comprehensivo de Stock

Los reportes de inventario proporcionan visibilidad completa sobre el estado del stock, movimientos, rotación, y optimización del inventario. El sistema puede identificar productos de lenta rotación, stock excesivo, y oportunidades de optimización de inventario.

Los reportes incluyen análisis de valorización de inventario, costos de mantenimiento, y proyecciones de necesidades futuras basadas en patrones históricos de consumo. Esto ayuda a optimizar los niveles de stock y reducir costos operativos.

#### GET /api/v1/reports/inventory/status

Obtiene el estado actual del inventario con métricas clave.

**Permisos Requeridos:** reports.inventory.view

**Query Parameters:**
- `sucursal_id` (uuid, opcional): Filtrar por sucursal específica
- `categoria_id` (uuid, opcional): Filtrar por categoría
- `incluir_valorizacion` (bool, opcional): Incluir valorización del inventario
- `stock_critico` (bool, opcional): Solo productos con stock crítico

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "resumen_general": {
      "total_productos": 8,
      "productos_con_stock": 8,
      "productos_sin_stock": 0,
      "productos_stock_critico": 1,
      "valor_total_inventario": 875000.00,
      "valor_costo_inventario": 520000.00
    },
    "distribucion_por_categoria": [
      {
        "categoria_id": "test-category-1",
        "categoria_nombre": "Test Herramientas",
        "productos_count": 2,
        "stock_total_unidades": 75,
        "valor_inventario": 375000.00,
        "rotacion_promedio_dias": 45
      }
    ],
    "productos_stock_critico": [
      {
        "producto_id": "test-product-5",
        "codigo": "TEST005",
        "nombre": "Llave Inglesa Test 10\"",
        "stock_actual": 2,
        "stock_minimo": 5,
        "stock_maximo": 20,
        "dias_stock_restante": 8,
        "venta_promedio_diaria": 0.25,
        "reorden_sugerido": 18
      }
    ],
    "productos_exceso_stock": [
      {
        "producto_id": "test-product-2",
        "codigo": "TEST002",
        "nombre": "Destornillador Test Phillips",
        "stock_actual": 50,
        "stock_optimo": 25,
        "exceso_unidades": 25,
        "valor_exceso": 50000.00,
        "meses_cobertura": 8.5
      }
    ],
    "movimientos_recientes": {
      "entradas_ultima_semana": 0,
      "salidas_ultima_semana": 36,
      "ajustes_ultima_semana": 0,
      "valor_movimientos": 54000.00
    }
  },
  "request_id": "reports_req_005",
  "timestamp": "2025-01-08T16:20:00Z"
}
```


## Reportes Financieros

### Análisis Financiero Integral

Los reportes financieros proporcionan una vista comprehensiva de la salud financiera del negocio, incluyendo análisis de rentabilidad, flujo de caja, márgenes por producto y categoría, y proyecciones financieras. Estos reportes están diseñados para satisfacer las necesidades tanto de la gestión operativa como de la dirección ejecutiva.

El sistema calcula automáticamente métricas financieras clave como ROI (Return on Investment), EBITDA, márgenes brutos y netos, y ratios de eficiencia operativa. Los reportes pueden agregarse por diferentes períodos y dimensiones para proporcionar insights tanto tácticos como estratégicos.

Los análisis incluyen comparaciones con períodos anteriores, benchmarking interno entre sucursales, y alertas automáticas cuando las métricas se desvían significativamente de los objetivos establecidos.

#### GET /api/v1/reports/financial/profitability

Analiza la rentabilidad del negocio por diferentes dimensiones.

**Permisos Requeridos:** reports.financial.view

**Query Parameters:**
- `fecha_inicio` (date, requerido): Fecha de inicio del análisis
- `fecha_fin` (date, requerido): Fecha de fin del análisis
- `dimension` (string, opcional): Dimensión de análisis (producto, categoria, sucursal, vendedor)
- `incluir_costos_operativos` (bool, opcional): Incluir costos operativos en el análisis
- `moneda` (string, opcional): Moneda para el reporte (CLP, USD)

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "periodo": {
      "fecha_inicio": "2025-01-01",
      "fecha_fin": "2025-01-08",
      "dias_operativos": 8
    },
    "resumen_financiero": {
      "ingresos_brutos": 116025.00,
      "costo_ventas": 67275.00,
      "margen_bruto": 48750.00,
      "margen_bruto_porcentaje": 42.0,
      "gastos_operativos": 25000.00,
      "utilidad_operativa": 23750.00,
      "margen_operativo_porcentaje": 20.5,
      "roi_periodo": 8.5
    },
    "analisis_por_producto": [
      {
        "producto_id": "test-product-1",
        "codigo": "TEST001",
        "nombre": "Martillo Test 500g",
        "ingresos": 15000.00,
        "costo_ventas": 8000.00,
        "margen_bruto": 7000.00,
        "margen_porcentaje": 46.7,
        "unidades_vendidas": 1,
        "contribucion_total": 6.0,
        "roi_producto": 87.5
      },
      {
        "producto_id": "test-product-4",
        "codigo": "TEST004",
        "nombre": "Destornillador Test Plano",
        "ingresos": 52500.00,
        "costo_ventas": 30000.00,
        "margen_bruto": 22500.00,
        "margen_porcentaje": 42.9,
        "unidades_vendidas": 15,
        "contribucion_total": 19.4,
        "roi_producto": 75.0
      }
    ],
    "tendencias_rentabilidad": {
      "margen_bruto_tendencia": "estable",
      "variacion_mensual": 2.3,
      "productos_mejorando": 3,
      "productos_deteriorando": 1
    },
    "metricas_eficiencia": {
      "rotacion_inventario": 4.2,
      "dias_inventario": 87,
      "rotacion_cuentas_cobrar": 12.5,
      "ciclo_conversion_efectivo": 45
    },
    "alertas_financieras": [
      {
        "tipo": "margen_bajo",
        "producto_id": "test-product-3",
        "mensaje": "Margen del producto por debajo del objetivo (25% vs 30%)",
        "impacto_estimado": -2500.00,
        "recomendacion": "Revisar costos o ajustar precio"
      }
    ],
    "proyecciones": {
      "ingresos_mes_completo": 435000.00,
      "margen_bruto_proyectado": 182700.00,
      "confianza_proyeccion": 0.78
    }
  },
  "request_id": "reports_req_006",
  "timestamp": "2025-01-08T16:25:00Z"
}
```

#### GET /api/v1/reports/financial/cashflow

Analiza el flujo de caja y liquidez del negocio.

**Permisos Requeridos:** reports.financial.view

**Query Parameters:**
- `fecha_inicio` (date, requerido): Fecha de inicio
- `fecha_fin` (date, requerido): Fecha de fin
- `incluir_proyeccion` (bool, opcional): Incluir proyección de flujo futuro
- `agrupar_por` (string, opcional): Agrupación temporal (dia, semana, mes)

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "resumen_periodo": {
      "ingresos_efectivo": 78275.00,
      "ingresos_tarjetas": 37750.00,
      "total_ingresos": 116025.00,
      "egresos_operativos": 45000.00,
      "flujo_neto": 71025.00,
      "liquidez_inicial": 150000.00,
      "liquidez_final": 221025.00
    },
    "flujo_diario": [
      {
        "fecha": "2025-01-06",
        "ingresos": 29750.00,
        "egresos": 5000.00,
        "flujo_neto": 24750.00,
        "saldo_acumulado": 174750.00
      },
      {
        "fecha": "2025-01-07",
        "ingresos": 86275.00,
        "egresos": 15000.00,
        "flujo_neto": 71275.00,
        "saldo_acumulado": 246025.00
      }
    ],
    "distribucion_ingresos": {
      "efectivo": {
        "monto": 78275.00,
        "porcentaje": 67.5,
        "transacciones": 2
      },
      "tarjeta_credito": {
        "monto": 18500.00,
        "porcentaje": 15.9,
        "transacciones": 1
      },
      "tarjeta_debito": {
        "monto": 19250.00,
        "porcentaje": 16.6,
        "transacciones": 1
      }
    },
    "metricas_liquidez": {
      "dias_efectivo_disponible": 45,
      "ratio_liquidez_corriente": 2.8,
      "velocidad_cobranza_dias": 2.5,
      "ciclo_efectivo_dias": 38
    },
    "proyeccion_flujo": {
      "proximos_7_dias": {
        "ingresos_estimados": 175000.00,
        "egresos_estimados": 65000.00,
        "flujo_neto_estimado": 110000.00
      },
      "proximo_mes": {
        "ingresos_estimados": 720000.00,
        "egresos_estimados": 280000.00,
        "flujo_neto_estimado": 440000.00
      }
    },
    "alertas_liquidez": [
      {
        "tipo": "oportunidad",
        "mensaje": "Exceso de liquidez detectado, considerar inversiones a corto plazo",
        "monto_exceso": 50000.00,
        "rendimiento_potencial": 2500.00
      }
    ]
  },
  "request_id": "reports_req_007",
  "timestamp": "2025-01-08T16:30:00Z"
}
```

## Analytics y Dashboards

### Inteligencia de Negocio Avanzada

Los analytics y dashboards proporcionan una vista ejecutiva del negocio con métricas clave, KPIs, y visualizaciones interactivas que permiten tomar decisiones informadas en tiempo real. El sistema utiliza técnicas de machine learning para identificar patrones, anomalías, y oportunidades de optimización.

Los dashboards son completamente configurables y pueden adaptarse a las necesidades específicas de diferentes roles y niveles organizacionales. Incluyen capacidades de drill-down que permiten explorar los datos desde vistas de alto nivel hasta detalles granulares.

#### GET /api/v1/reports/dashboard/executive

Obtiene el dashboard ejecutivo con métricas clave del negocio.

**Permisos Requeridos:** reports.analytics.advanced

**Query Parameters:**
- `periodo` (string, opcional): Período del dashboard (hoy, semana, mes, trimestre)
- `sucursal_id` (uuid, opcional): Filtrar por sucursal específica
- `comparar_periodo_anterior` (bool, opcional): Incluir comparación con período anterior

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "periodo": {
      "tipo": "mes",
      "fecha_inicio": "2025-01-01",
      "fecha_fin": "2025-01-08",
      "progreso_periodo": 25.8
    },
    "kpis_principales": {
      "ventas_totales": {
        "valor": 116025.00,
        "objetivo": 500000.00,
        "progreso_porcentual": 23.2,
        "tendencia": "positiva",
        "variacion_periodo_anterior": 17.8
      },
      "margen_bruto": {
        "valor": 42.0,
        "objetivo": 45.0,
        "progreso_porcentual": 93.3,
        "tendencia": "estable",
        "variacion_periodo_anterior": 1.2
      },
      "transacciones": {
        "valor": 4,
        "objetivo": 200,
        "progreso_porcentual": 2.0,
        "tendencia": "negativa",
        "variacion_periodo_anterior": -20.0
      },
      "ticket_promedio": {
        "valor": 29006.25,
        "objetivo": 25000.00,
        "progreso_porcentual": 116.0,
        "tendencia": "muy_positiva",
        "variacion_periodo_anterior": 47.3
      }
    },
    "metricas_operativas": {
      "productos_activos": 8,
      "stock_total_valor": 875000.00,
      "rotacion_inventario": 4.2,
      "productos_stock_critico": 1,
      "usuarios_activos": 4,
      "terminales_online": 1
    },
    "top_performers": {
      "productos": [
        {
          "id": "test-product-4",
          "nombre": "Destornillador Test Plano",
          "ventas": 52500.00,
          "unidades": 15
        }
      ],
      "vendedores": [
        {
          "id": "test-user-2",
          "nombre": "Vendedor Test",
          "ventas": 116025.00,
          "transacciones": 4
        }
      ],
      "categorias": [
        {
          "id": "test-category-1",
          "nombre": "Test Herramientas",
          "ventas": 116025.00,
          "participacion": 100.0
        }
      ]
    },
    "alertas_criticas": [
      {
        "tipo": "stock_critico",
        "mensaje": "1 producto con stock crítico requiere reposición",
        "prioridad": "alta",
        "accion_requerida": "generar_orden_compra"
      }
    ],
    "tendencias_graficos": {
      "ventas_diarias": [
        {
          "fecha": "2025-01-06",
          "ventas": 29750.00,
          "transacciones": 1
        },
        {
          "fecha": "2025-01-07",
          "ventas": 86275.00,
          "transacciones": 3
        }
      ],
      "distribucion_ventas_hora": [
        {
          "hora": 10,
          "ventas": 29750.00,
          "transacciones": 1
        },
        {
          "hora": 14,
          "ventas": 86275.00,
          "transacciones": 3
        }
      ]
    },
    "insights_ia": [
      {
        "tipo": "oportunidad",
        "titulo": "Incremento en ticket promedio",
        "descripcion": "El ticket promedio ha aumentado 47% vs período anterior, indicando mejora en estrategia de venta cruzada",
        "confianza": 0.89,
        "impacto_estimado": "alto"
      },
      {
        "tipo": "alerta",
        "titulo": "Reducción en número de transacciones",
        "descripcion": "20% menos transacciones vs período anterior, revisar estrategias de atracción de clientes",
        "confianza": 0.92,
        "impacto_estimado": "medio"
      }
    ]
  },
  "request_id": "reports_req_008",
  "timestamp": "2025-01-08T16:35:00Z"
}
```

#### GET /api/v1/reports/analytics/customer

Analiza el comportamiento y segmentación de clientes.

**Permisos Requeridos:** reports.analytics.advanced

**Query Parameters:**
- `fecha_inicio` (date, requerido): Fecha de inicio del análisis
- `fecha_fin` (date, requerido): Fecha de fin del análisis
- `incluir_segmentacion` (bool, opcional): Incluir análisis de segmentación
- `incluir_predicciones` (bool, opcional): Incluir predicciones de comportamiento

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "resumen_clientes": {
      "clientes_unicos": 3,
      "clientes_nuevos": 1,
      "clientes_recurrentes": 2,
      "tasa_retencion": 66.7,
      "valor_promedio_cliente": 38675.00,
      "frecuencia_compra_promedio": 1.3
    },
    "segmentacion_clientes": {
      "alto_valor": {
        "cantidad": 1,
        "valor_promedio": 57750.00,
        "frecuencia_promedio": 2.0,
        "caracteristicas": ["compras_grandes", "productos_premium"]
      },
      "medio_valor": {
        "cantidad": 1,
        "valor_promedio": 29750.00,
        "frecuencia_promedio": 1.0,
        "caracteristicas": ["compras_regulares", "precio_consciente"]
      },
      "bajo_valor": {
        "cantidad": 1,
        "valor_promedio": 28525.00,
        "frecuencia_promedio": 1.0,
        "caracteristicas": ["compras_ocasionales", "sensible_precio"]
      }
    },
    "patrones_comportamiento": {
      "horarios_preferidos": [
        {
          "hora": 10,
          "porcentaje_transacciones": 25.0
        },
        {
          "hora": 14,
          "porcentaje_transacciones": 75.0
        }
      ],
      "dias_preferidos": [
        {
          "dia": "lunes",
          "porcentaje_transacciones": 25.0
        },
        {
          "dia": "martes",
          "porcentaje_transacciones": 75.0
        }
      ],
      "productos_preferidos": [
        {
          "categoria": "herramientas",
          "porcentaje_compras": 100.0
        }
      ]
    },
    "analisis_cohorte": {
      "mes_1": {
        "clientes_adquiridos": 3,
        "retencion_mes_2": 0.0,
        "valor_promedio_mes_1": 38675.00
      }
    },
    "predicciones": {
      "clientes_riesgo_abandono": [
        {
          "cliente_id": "cliente_001",
          "probabilidad_abandono": 0.65,
          "ultima_compra": "2025-01-06",
          "valor_historico": 29750.00,
          "recomendacion": "contactar_promocion"
        }
      ],
      "oportunidades_upselling": [
        {
          "cliente_id": "cliente_002",
          "productos_recomendados": ["test-product-1", "test-product-5"],
          "valor_potencial": 25000.00,
          "probabilidad_compra": 0.78
        }
      ]
    }
  },
  "request_id": "reports_req_009",
  "timestamp": "2025-01-08T16:40:00Z"
}
```

## Exportación de Datos

### Formatos y Opciones de Exportación

El sistema de exportación permite generar reportes en múltiples formatos optimizados para diferentes casos de uso: PDF para presentaciones ejecutivas, Excel para análisis detallado, CSV para integración con otros sistemas, y JSON para desarrollo de aplicaciones.

Cada formato incluye opciones de personalización como logos corporativos, formatos de fecha y moneda específicos por región, y plantillas predefinidas que mantienen consistencia visual en todos los reportes.

#### POST /api/v1/reports/export

Exporta un reporte en el formato especificado.

**Permisos Requeridos:** reports.export

**Request Body:**
```json
{
  "tipo_reporte": "sales_summary",
  "parametros": {
    "fecha_inicio": "2025-01-01",
    "fecha_fin": "2025-01-08",
    "sucursal_id": "test-sucursal-1",
    "incluir_graficos": true
  },
  "formato_exportacion": "pdf",
  "opciones": {
    "incluir_logo": true,
    "orientacion": "landscape",
    "tamaño_pagina": "A4",
    "incluir_fecha_generacion": true,
    "marca_agua": "CONFIDENCIAL",
    "idioma": "es"
  },
  "configuracion_entrega": {
    "metodo": "download",
    "email_destinatario": "gerente@ferrepos.com",
    "nombre_archivo": "reporte_ventas_enero_2025"
  }
}
```

**Response (202 Accepted):**
```json
{
  "success": true,
  "data": {
    "export_id": "export_001",
    "estado": "procesando",
    "tipo_reporte": "sales_summary",
    "formato": "pdf",
    "tiempo_estimado_segundos": 30,
    "progreso": {
      "porcentaje": 0,
      "fase_actual": "preparando_datos"
    },
    "urls": {
      "estado": "/api/v1/reports/export/export_001/status",
      "cancelar": "/api/v1/reports/export/export_001/cancel",
      "descargar": "/api/v1/reports/export/export_001/download"
    },
    "expires_at": "2025-01-15T16:45:00Z",
    "created_at": "2025-01-08T16:45:00Z"
  },
  "request_id": "reports_req_010",
  "timestamp": "2025-01-08T16:45:00Z"
}
```

#### GET /api/v1/reports/export/{export_id}/status

Verifica el estado de una exportación en proceso.

**Permisos Requeridos:** reports.export

**Path Parameters:**
- `export_id` (string, requerido): ID de la exportación

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "export_id": "export_001",
    "estado": "completado",
    "progreso": {
      "porcentaje": 100,
      "fase_actual": "finalizado"
    },
    "archivo_resultado": {
      "url": "/api/v1/reports/export/export_001/download",
      "nombre": "reporte_ventas_enero_2025.pdf",
      "tamaño_bytes": 1048576,
      "checksum": "sha256:abcdef123456...",
      "expires_at": "2025-01-15T16:45:00Z"
    },
    "estadisticas": {
      "registros_procesados": 4,
      "tiempo_procesamiento_segundos": 25,
      "paginas_generadas": 8
    },
    "completed_at": "2025-01-08T16:46:25Z"
  },
  "request_id": "reports_req_011",
  "timestamp": "2025-01-08T16:50:00Z"
}
```

## Programación de Reportes

### Automatización de Reportes Recurrentes

El sistema permite programar reportes automáticos que se generan y distribuyen según horarios predefinidos. Esto es especialmente útil para reportes ejecutivos diarios, resúmenes semanales, y análisis mensuales que requieren distribución regular a stakeholders específicos.

#### POST /api/v1/reports/schedule

Programa un reporte para ejecución automática.

**Permisos Requeridos:** reports.schedule

**Request Body:**
```json
{
  "nombre": "Reporte Diario de Ventas",
  "descripcion": "Resumen diario de ventas para gerencia",
  "tipo_reporte": "sales_summary",
  "parametros_reporte": {
    "periodo": "yesterday",
    "incluir_comparacion": true,
    "sucursal_id": "all"
  },
  "programacion": {
    "tipo": "cron",
    "expresion": "0 8 * * 1-6",
    "zona_horaria": "America/Santiago"
  },
  "formato_exportacion": "pdf",
  "distribucion": {
    "metodo": "email",
    "destinatarios": [
      "gerente@ferrepos.com",
      "supervisor@ferrepos.com"
    ],
    "asunto": "Reporte Diario de Ventas - {fecha}",
    "mensaje": "Adjunto encontrará el reporte diario de ventas correspondiente al {fecha}."
  },
  "activo": true,
  "fecha_inicio": "2025-01-09",
  "fecha_fin": "2025-12-31"
}
```

**Response (201 Created):**
```json
{
  "success": true,
  "data": {
    "schedule_id": "schedule_001",
    "nombre": "Reporte Diario de Ventas",
    "estado": "activo",
    "proxima_ejecucion": "2025-01-09T08:00:00Z",
    "ejecuciones_programadas": 365,
    "created_at": "2025-01-08T16:55:00Z"
  },
  "request_id": "reports_req_012",
  "timestamp": "2025-01-08T16:55:00Z"
}
```

