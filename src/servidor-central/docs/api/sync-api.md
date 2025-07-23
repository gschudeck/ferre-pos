# API Sync - Documentación de Sincronización

## Introducción

El API Sync es responsable de la sincronización bidireccional de datos entre el servidor central y las sucursales. Está diseñado para manejar grandes volúmenes de datos, resolver conflictos automáticamente y mantener la consistencia de datos en tiempo real.

## Configuración Base

### URL Base
```
http://localhost:8080/api/sync
```

### Autenticación
Requiere autenticación JWT con permisos de sincronización:

```http
Authorization: Bearer <jwt_token>
X-API-Key: <api_key>  # Para sistemas externos
```

## Gestión de Sincronización

### POST /iniciar
Inicia un proceso de sincronización.

#### Request
```json
{
  "sucursal_id": "uuid-sucursal",
  "entidades": ["productos", "stock", "precios", "ventas"],
  "direccion": "bidireccional", // bidireccional, servidor_a_cliente, cliente_a_servidor
  "modo": "incremental", // incremental, completo
  "configuracion": {
    "batch_size": 100,
    "max_duration": "30m",
    "enable_compression": true,
    "conflict_resolution": "manual"
  },
  "filtros": {
    "fecha_desde": "2024-01-01T00:00:00Z",
    "fecha_hasta": "2024-01-20T23:59:59Z",
    "categorias": ["uuid-cat-1", "uuid-cat-2"]
  }
}
```

#### Response (202 Accepted)
```json
{
  "success": true,
  "data": {
    "sincronizacion": {
      "id": "uuid-sync",
      "sucursal_id": "uuid-sucursal",
      "estado": "iniciada",
      "tipo": "incremental",
      "direccion": "bidireccional",
      "entidades": ["productos", "stock", "precios", "ventas"],
      "progreso": {
        "porcentaje": 0,
        "entidad_actual": null,
        "registros_procesados": 0,
        "registros_totales": 0
      },
      "configuracion": {
        "batch_size": 100,
        "max_duration": "30m",
        "enable_compression": true,
        "conflict_resolution": "manual"
      },
      "estadisticas": {
        "inicio": "2024-01-20T15:30:00Z",
        "estimacion_fin": null,
        "duracion_estimada": null
      },
      "created_at": "2024-01-20T15:30:00Z"
    }
  }
}
```

### GET /estado/:id
Obtiene el estado actual de una sincronización.

#### Response (200 OK)
```json
{
  "success": true,
  "data": {
    "sincronizacion": {
      "id": "uuid-sync",
      "estado": "en_progreso",
      "progreso": {
        "porcentaje": 65,
        "entidad_actual": "stock",
        "registros_procesados": 6500,
        "registros_totales": 10000,
        "entidades_completadas": ["productos", "precios"],
        "entidades_pendientes": ["stock", "ventas"]
      },
      "estadisticas": {
        "inicio": "2024-01-20T15:30:00Z",
        "duracion_actual": "15m30s",
        "estimacion_fin": "2024-01-20T16:00:00Z",
        "velocidad_promedio": "120 registros/min"
      },
      "resultados_por_entidad": {
        "productos": {
          "estado": "completada",
          "registros_procesados": 2500,
          "registros_sincronizados": 2450,
          "registros_con_conflicto": 50,
          "errores": 0,
          "duracion": "8m15s"
        },
        "precios": {
          "estado": "completada",
          "registros_procesados": 2000,
          "registros_sincronizados": 2000,
          "registros_con_conflicto": 0,
          "errores": 0,
          "duracion": "5m20s"
        },
        "stock": {
          "estado": "en_progreso",
          "registros_procesados": 2000,
          "registros_sincronizados": 1980,
          "registros_con_conflicto": 15,
          "errores": 5
        }
      },
      "conflictos": [
        {
          "id": "uuid-conflicto",
          "entidad": "productos",
          "registro_id": "uuid-producto",
          "tipo": "modificacion_concurrente",
          "estado": "pendiente"
        }
      ],
      "errores": [
        {
          "entidad": "stock",
          "registro_id": "uuid-stock",
          "error": "Producto no encontrado",
          "codigo": "PRODUCT_NOT_FOUND",
          "timestamp": "2024-01-20T15:45:00Z"
        }
      ]
    }
  }
}
```

### POST /detener/:id
Detiene una sincronización en progreso.

#### Request
```json
{
  "motivo": "Mantenimiento programado",
  "forzar": false, // Si true, detiene inmediatamente
  "completar_lote_actual": true
}
```

#### Response (200 OK)
```json
{
  "success": true,
  "data": {
    "sincronizacion": {
      "id": "uuid-sync",
      "estado": "detenida",
      "motivo_detencion": "Mantenimiento programado",
      "detenida_en": "2024-01-20T15:50:00Z",
      "progreso_final": {
        "porcentaje": 75,
        "registros_procesados": 7500,
        "registros_totales": 10000
      },
      "puede_reanudar": true
    }
  }
}
```

### GET /historial
Lista el historial de sincronizaciones.

#### Query Parameters
- `sucursal_id` (uuid): Filtrar por sucursal
- `estado` (string): Filtrar por estado
- `fecha_inicio` (date): Fecha inicio
- `fecha_fin` (date): Fecha fin
- `page` (int): Página
- `limit` (int): Límite por página

#### Response (200 OK)
```json
{
  "success": true,
  "data": {
    "sincronizaciones": [
      {
        "id": "uuid-sync",
        "sucursal": {
          "id": "uuid-sucursal",
          "nombre": "Sucursal Centro",
          "codigo": "SUC-001"
        },
        "estado": "completada",
        "tipo": "incremental",
        "direccion": "bidireccional",
        "entidades": ["productos", "stock", "precios"],
        "inicio": "2024-01-20T14:00:00Z",
        "fin": "2024-01-20T14:25:00Z",
        "duracion": "25m",
        "registros_procesados": 5000,
        "registros_sincronizados": 4950,
        "conflictos": 30,
        "errores": 20,
        "usuario": {
          "id": "uuid-usuario",
          "nombre": "Sistema Automático"
        }
      }
    ],
    "estadisticas": {
      "total_sincronizaciones": 156,
      "sincronizaciones_exitosas": 142,
      "sincronizaciones_con_errores": 14,
      "promedio_duracion": "18m30s",
      "ultima_sincronizacion": "2024-01-20T15:30:00Z"
    },
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 156,
      "total_pages": 8
    }
  }
}
```

## Gestión de Conflictos

### GET /conflictos
Lista conflictos de sincronización.

#### Query Parameters
- `estado` (string): pendiente, resuelto, ignorado
- `entidad` (string): productos, stock, precios, ventas
- `sucursal_id` (uuid): ID de sucursal
- `fecha_inicio` (date): Fecha inicio
- `fecha_fin` (date): Fecha fin

#### Response (200 OK)
```json
{
  "success": true,
  "data": {
    "conflictos": [
      {
        "id": "uuid-conflicto",
        "sincronizacion_id": "uuid-sync",
        "entidad": "productos",
        "registro_id": "uuid-producto",
        "tipo": "modificacion_concurrente",
        "estado": "pendiente",
        "prioridad": "alta",
        "descripcion": "Producto modificado simultáneamente en servidor y cliente",
        "datos_servidor": {
          "nombre": "Martillo de Carpintero 16oz",
          "precio": 25000,
          "modificado_en": "2024-01-20T15:20:00Z",
          "modificado_por": "uuid-usuario-servidor"
        },
        "datos_cliente": {
          "nombre": "Martillo Carpintero 16oz",
          "precio": 24000,
          "modificado_en": "2024-01-20T15:18:00Z",
          "modificado_por": "uuid-usuario-cliente"
        },
        "diferencias": [
          {
            "campo": "nombre",
            "valor_servidor": "Martillo de Carpintero 16oz",
            "valor_cliente": "Martillo Carpintero 16oz"
          },
          {
            "campo": "precio",
            "valor_servidor": 25000,
            "valor_cliente": 24000
          }
        ],
        "sugerencias_resolucion": [
          {
            "estrategia": "servidor_gana",
            "descripcion": "Mantener valores del servidor",
            "confianza": 0.7
          },
          {
            "estrategia": "cliente_gana",
            "descripcion": "Mantener valores del cliente",
            "confianza": 0.3
          },
          {
            "estrategia": "merge",
            "descripcion": "Combinar cambios",
            "confianza": 0.8,
            "valores_propuestos": {
              "nombre": "Martillo de Carpintero 16oz",
              "precio": 25000
            }
          }
        ],
        "impacto": {
          "nivel": "medio",
          "entidades_afectadas": ["stock", "ventas"],
          "registros_dependientes": 15
        },
        "created_at": "2024-01-20T15:30:00Z"
      }
    ],
    "resumen": {
      "total_conflictos": 45,
      "pendientes": 30,
      "resueltos": 15,
      "por_entidad": {
        "productos": 20,
        "stock": 15,
        "precios": 8,
        "ventas": 2
      },
      "por_prioridad": {
        "alta": 5,
        "media": 25,
        "baja": 15
      }
    }
  }
}
```

### POST /conflictos/:id/resolver
Resuelve un conflicto específico.

#### Request
```json
{
  "estrategia": "merge", // servidor_gana, cliente_gana, merge, manual
  "valores_finales": {
    "nombre": "Martillo de Carpintero 16oz",
    "precio": 25000
  },
  "observaciones": "Se mantiene el nombre completo del servidor y el precio actualizado",
  "aplicar_a_similares": true, // Aplicar la misma estrategia a conflictos similares
  "notificar_sucursal": true
}
```

#### Response (200 OK)
```json
{
  "success": true,
  "data": {
    "conflicto": {
      "id": "uuid-conflicto",
      "estado": "resuelto",
      "estrategia_aplicada": "merge",
      "valores_finales": {
        "nombre": "Martillo de Carpintero 16oz",
        "precio": 25000
      },
      "resuelto_en": "2024-01-20T16:00:00Z",
      "resuelto_por": "uuid-usuario",
      "observaciones": "Se mantiene el nombre completo del servidor y el precio actualizado",
      "conflictos_similares_resueltos": 3
    },
    "sincronizacion_actualizada": {
      "id": "uuid-sync",
      "conflictos_pendientes": 29,
      "puede_continuar": true
    }
  }
}
```

### GET /conflictos/:id
Obtiene detalles de un conflicto específico.

#### Response (200 OK)
```json
{
  "success": true,
  "data": {
    "conflicto": {
      // Datos completos del conflicto como en la lista
      "historial_resoluciones": [
        {
          "intento": 1,
          "estrategia": "auto_servidor",
          "resultado": "fallido",
          "motivo": "Validación de negocio falló",
          "timestamp": "2024-01-20T15:35:00Z"
        }
      ],
      "analisis_impacto": {
        "registros_bloqueados": 5,
        "operaciones_pendientes": 12,
        "tiempo_bloqueo": "25m",
        "costo_estimado": "medio"
      }
    }
  }
}
```

## Configuración de Sincronización

### GET /configuracion
Obtiene la configuración actual de sincronización.

#### Response (200 OK)
```json
{
  "success": true,
  "data": {
    "configuracion": {
      "global": {
        "max_concurrent_syncs": 5,
        "sync_interval": "15m",
        "batch_size": 100,
        "max_retries": 3,
        "retry_backoff_multiplier": 2.0,
        "conflict_resolution_mode": "manual",
        "enable_compression": true,
        "max_sync_duration": "30m"
      },
      "por_entidad": {
        "productos": {
          "enabled": true,
          "priority": 1,
          "sync_frequency": "5m",
          "batch_size": 50,
          "enable_real_time": true,
          "conflict_strategy": "server_wins"
        },
        "stock": {
          "enabled": true,
          "priority": 2,
          "sync_frequency": "2m",
          "batch_size": 100,
          "enable_real_time": true,
          "conflict_strategy": "client_wins"
        },
        "precios": {
          "enabled": true,
          "priority": 1,
          "sync_frequency": "10m",
          "batch_size": 200,
          "enable_real_time": false,
          "conflict_strategy": "server_wins"
        },
        "ventas": {
          "enabled": true,
          "priority": 3,
          "sync_frequency": "1m",
          "batch_size": 25,
          "enable_real_time": true,
          "conflict_strategy": "client_wins"
        }
      },
      "notificaciones": {
        "enable_sync_notifications": true,
        "enable_conflict_notifications": true,
        "enable_error_notifications": true,
        "recipients": [
          "sync-admin@ferreteria.com",
          "soporte@ferreteria.com"
        ]
      },
      "limpieza": {
        "log_retention_days": 90,
        "completed_sync_retention_days": 7,
        "failed_sync_retention_days": 30,
        "cleanup_interval": "24h"
      }
    }
  }
}
```

### PUT /configuracion
Actualiza la configuración de sincronización.

#### Request
```json
{
  "global": {
    "max_concurrent_syncs": 8,
    "sync_interval": "10m",
    "conflict_resolution_mode": "auto_server"
  },
  "por_entidad": {
    "productos": {
      "sync_frequency": "3m",
      "conflict_strategy": "server_wins"
    }
  },
  "notificaciones": {
    "enable_conflict_notifications": false
  }
}
```

#### Response (200 OK)
```json
{
  "success": true,
  "data": {
    "configuracion": {
      // Configuración actualizada completa
    },
    "cambios_aplicados": [
      "max_concurrent_syncs: 5 -> 8",
      "sync_interval: 15m -> 10m",
      "productos.sync_frequency: 5m -> 3m"
    ],
    "reinicio_requerido": false,
    "aplicado_en": "2024-01-20T16:15:00Z"
  }
}
```

## Monitoreo y Métricas

### GET /metricas
Obtiene métricas de sincronización.

#### Query Parameters
- `periodo` (string): 1h, 24h, 7d, 30d
- `sucursal_id` (uuid): Filtrar por sucursal
- `entidad` (string): Filtrar por entidad

#### Response (200 OK)
```json
{
  "success": true,
  "data": {
    "periodo": "24h",
    "metricas": {
      "sincronizaciones": {
        "total": 96,
        "exitosas": 88,
        "con_errores": 6,
        "canceladas": 2,
        "tasa_exito": 0.917
      },
      "rendimiento": {
        "duracion_promedio": "18m30s",
        "duracion_minima": "5m15s",
        "duracion_maxima": "45m20s",
        "registros_por_minuto": 125,
        "throughput_promedio": "2.1 MB/min"
      },
      "conflictos": {
        "total": 156,
        "resueltos_automaticamente": 98,
        "resueltos_manualmente": 45,
        "pendientes": 13,
        "tasa_resolucion_automatica": 0.628
      },
      "errores": {
        "total": 45,
        "por_tipo": {
          "timeout": 15,
          "connection_error": 12,
          "validation_error": 10,
          "business_rule_error": 8
        },
        "tasa_error": 0.047
      },
      "por_entidad": {
        "productos": {
          "sincronizaciones": 24,
          "registros_procesados": 12000,
          "conflictos": 45,
          "errores": 8
        },
        "stock": {
          "sincronizaciones": 48,
          "registros_procesados": 24000,
          "conflictos": 78,
          "errores": 15
        }
      },
      "por_sucursal": [
        {
          "sucursal_id": "uuid-sucursal-1",
          "nombre": "Sucursal Centro",
          "sincronizaciones": 32,
          "tasa_exito": 0.94,
          "duracion_promedio": "15m"
        }
      ]
    },
    "tendencias": {
      "sincronizaciones_por_hora": [
        {"hora": "00:00", "total": 2},
        {"hora": "01:00", "total": 1},
        // ... datos por hora
      ],
      "conflictos_por_dia": [
        {"fecha": "2024-01-19", "total": 23},
        {"fecha": "2024-01-20", "total": 18}
      ]
    }
  }
}
```

### GET /health
Health check del API Sync.

#### Response (200 OK)
```json
{
  "status": "healthy",
  "timestamp": "2024-01-20T16:30:00Z",
  "version": "1.0.0",
  "checks": {
    "database": {
      "status": "healthy",
      "response_time_ms": 8,
      "connections_active": 5,
      "connections_max": 20
    },
    "queue": {
      "status": "healthy",
      "size": 12,
      "max_size": 10000,
      "workers_active": 3,
      "workers_max": 5
    },
    "external_services": {
      "sucursales": [
        {
          "sucursal_id": "uuid-sucursal-1",
          "status": "healthy",
          "last_sync": "2024-01-20T16:15:00Z",
          "response_time_ms": 150
        }
      ]
    }
  },
  "metrics": {
    "active_syncs": 2,
    "pending_conflicts": 13,
    "queue_size": 12,
    "average_sync_duration": "18m30s"
  }
}
```

## Webhooks y Notificaciones

### POST /webhooks/configurar
Configura webhooks para eventos de sincronización.

#### Request
```json
{
  "url": "https://mi-sistema.com/webhooks/sync",
  "eventos": [
    "sync.started",
    "sync.completed",
    "sync.failed",
    "conflict.detected",
    "conflict.resolved"
  ],
  "headers": {
    "Authorization": "Bearer webhook-token",
    "X-Source": "ferre-pos-sync"
  },
  "retry_config": {
    "max_retries": 3,
    "retry_delay": "5s",
    "timeout": "30s"
  },
  "activo": true
}
```

#### Response (201 Created)
```json
{
  "success": true,
  "data": {
    "webhook": {
      "id": "uuid-webhook",
      "url": "https://mi-sistema.com/webhooks/sync",
      "eventos": ["sync.started", "sync.completed", "sync.failed"],
      "activo": true,
      "secret": "webhook-secret-key",
      "created_at": "2024-01-20T16:45:00Z"
    }
  }
}
```

## Códigos de Error Específicos

### Errores de Sincronización (422)
- `SYNC_ALREADY_RUNNING`: Sincronización ya en progreso
- `SYNC_INVALID_ENTITY`: Entidad no válida para sincronización
- `SYNC_CONFIGURATION_ERROR`: Error en configuración de sincronización
- `SYNC_TIMEOUT`: Timeout en sincronización
- `SYNC_CONFLICT_UNRESOLVED`: Conflicto no resuelto bloquea sincronización

### Errores de Conflictos (409)
- `CONFLICT_RESOLUTION_FAILED`: Falló la resolución de conflicto
- `CONFLICT_INVALID_STRATEGY`: Estrategia de resolución inválida
- `CONFLICT_DATA_INCONSISTENT`: Datos inconsistentes en conflicto

### Errores de Conectividad (503)
- `SYNC_SERVICE_UNAVAILABLE`: Servicio de sincronización no disponible
- `SYNC_SUCURSAL_UNREACHABLE`: Sucursal no alcanzable
- `SYNC_NETWORK_ERROR`: Error de red durante sincronización

---

**Autor**: Manus AI  
**Versión**: 1.0.0  
**Fecha**: 2024-01-XX

