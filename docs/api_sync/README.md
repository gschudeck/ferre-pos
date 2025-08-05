# API Sync - Documentación Técnica

**Sistema FERRE-POS - Servidor Central**  
**Versión**: 1.0.0  
**Puerto**: 8081  
**Autor**: Manus AI  
**Fecha**: Enero 2025

---

## Tabla de Contenidos

1. [Introducción](#introducción)
2. [Autenticación de Terminales](#autenticación-de-terminales)
3. [Sincronización de Datos](#sincronización-de-datos)
4. [Gestión de Conflictos](#gestión-de-conflictos)
5. [Heartbeat y Estado](#heartbeat-y-estado)
6. [Endpoints de Sincronización](#endpoints-de-sincronización)
7. [Resolución de Conflictos](#resolución-de-conflictos)
8. [Monitoreo y Métricas](#monitoreo-y-métricas)
9. [Códigos de Error](#códigos-de-error)
10. [Guías de Implementación](#guías-de-implementación)
11. [Referencias](#referencias)

---

## Introducción

La API Sync es el componente crítico del sistema FERRE-POS responsable de mantener la sincronización de datos entre el servidor central y las terminales distribuidas en las sucursales. Esta API está diseñada para manejar entornos con conectividad intermitente, garantizando que las operaciones puedan continuar offline y sincronizarse cuando la conectividad se restablezca.

El sistema de sincronización implementa un modelo de replicación bidireccional con resolución automática de conflictos, permitiendo que múltiples terminales operen independientemente mientras mantienen la consistencia eventual de los datos. La API utiliza técnicas avanzadas de versionado, timestamps vectoriales, y algoritmos de merge para garantizar la integridad de los datos durante el proceso de sincronización.

La arquitectura de sincronización está optimizada para minimizar el ancho de banda utilizado, transmitiendo solo los cambios incrementales (deltas) en lugar de datasets completos. Esto permite que el sistema funcione eficientemente incluso en conexiones de baja velocidad o con limitaciones de datos.

### Características Principales

La API Sync implementa un sistema de sincronización inteligente que detecta automáticamente qué datos han cambiado desde la última sincronización, utilizando timestamps y checksums para optimizar la transferencia de datos. El sistema soporta sincronización parcial, permitiendo que las terminales soliciten solo los datos específicos que necesitan, como productos de ciertas categorías o transacciones de un período específico.

El manejo de conflictos se basa en reglas de negocio configurables que pueden priorizar diferentes fuentes de datos según el contexto. Por ejemplo, los cambios de precios siempre se resuelven a favor del servidor central, mientras que las ventas locales tienen prioridad sobre las transacciones remotas para evitar duplicaciones.

La API incluye un sistema robusto de recuperación ante fallos que puede reanudar sincronizaciones interrumpidas desde el punto donde se detuvieron, evitando la necesidad de retransmitir datos ya procesados. Esto es especialmente importante en entornos con conectividad inestable donde las interrupciones son frecuentes.

### Arquitectura de Sincronización

El sistema utiliza un modelo de sincronización basado en eventos donde cada cambio en los datos genera un evento que se almacena en una cola de sincronización. Estos eventos incluyen metadatos completos sobre el cambio, incluyendo el usuario que lo realizó, la terminal de origen, y el timestamp preciso de la modificación.

La sincronización opera en múltiples niveles: sincronización de esquema para cambios estructurales, sincronización de datos maestros para productos y configuraciones, y sincronización transaccional para ventas y movimientos de stock. Cada nivel tiene sus propias reglas de prioridad y resolución de conflictos.

Para garantizar la consistencia, la API implementa un sistema de bloqueos distribuidos que previene modificaciones concurrentes de los mismos registros durante el proceso de sincronización. Estos bloqueos son de corta duración y se liberan automáticamente si una terminal se desconecta inesperadamente.


## Autenticación de Terminales

### Sistema de Autenticación Específico

La API Sync utiliza un sistema de autenticación especializado diseñado específicamente para terminales POS, diferente del sistema de autenticación de usuarios de la API POS. Este sistema está optimizado para dispositivos que operan de forma autónoma y requieren acceso continuo a los servicios de sincronización.

Cada terminal tiene credenciales únicas que incluyen un identificador de terminal, una clave secreta compartida, y certificados digitales para comunicación segura. El sistema implementa rotación automática de claves para mantener la seguridad a largo plazo, y soporta revocación inmediata de credenciales en caso de compromiso de seguridad.

La autenticación de terminales incluye validación de la dirección MAC, verificación de la ubicación geográfica aproximada, y análisis de patrones de comportamiento para detectar posibles intentos de suplantación. El sistema mantiene un registro detallado de todos los intentos de autenticación para auditoría y detección de anomalías.

#### POST /api/v1/sync/auth/terminal

Autentica una terminal y obtiene tokens de acceso para sincronización.

**Request Body:**
```json
{
  "terminal_id": "test-terminal-1",
  "terminal_secret": "secret_key_for_terminal_1",
  "mac_address": "00:11:22:33:44:55",
  "ip_address": "192.168.1.100",
  "software_version": "FERRE-POS Terminal v1.0.5",
  "hardware_info": {
    "model": "HP EliteDesk 800",
    "serial": "ABC123456",
    "cpu": "Intel i5-8500",
    "memory_gb": 8,
    "storage_gb": 256
  },
  "location": {
    "sucursal_id": "test-sucursal-1",
    "zona": "Caja Principal",
    "coordenadas": {
      "lat": -33.4489,
      "lng": -70.6693
    }
  }
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
    "expires_in": 3600,
    "terminal_info": {
      "id": "test-terminal-1",
      "codigo": "TERM001",
      "nombre": "Terminal Test Principal",
      "sucursal_id": "test-sucursal-1",
      "sucursal_nombre": "Test Sucursal Principal",
      "estado": "activo",
      "ultima_sincronizacion": "2025-01-08T11:30:00Z",
      "configuracion": {
        "sync_interval_minutes": 15,
        "offline_mode_enabled": true,
        "max_offline_transactions": 1000,
        "auto_sync_enabled": true
      }
    },
    "sync_endpoints": {
      "heartbeat": "/api/v1/sync/heartbeat",
      "pull_changes": "/api/v1/sync/pull",
      "push_changes": "/api/v1/sync/push",
      "resolve_conflicts": "/api/v1/sync/conflicts"
    },
    "permissions": [
      "sync.pull.productos",
      "sync.pull.precios",
      "sync.push.ventas",
      "sync.push.movimientos_stock"
    ]
  },
  "request_id": "sync_req_001",
  "timestamp": "2025-01-08T13:00:00Z"
}
```

#### POST /api/v1/sync/auth/refresh

Renueva el token de acceso de una terminal.

**Request Body:**
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "terminal_id": "test-terminal-1"
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "token_type": "Bearer",
    "expires_in": 3600,
    "issued_at": "2025-01-08T13:05:00Z"
  },
  "request_id": "sync_req_002",
  "timestamp": "2025-01-08T13:05:00Z"
}
```

### Gestión de Certificados

#### GET /api/v1/sync/auth/certificate

Obtiene el certificado digital actualizado para la terminal.

**Headers:**
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
X-Terminal-ID: test-terminal-1
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "certificate": "-----BEGIN CERTIFICATE-----\nMIIDXTCCAkWgAwIBAgIJAKoK...",
    "private_key": "-----BEGIN PRIVATE KEY-----\nMIIEvgIBADANBgkqhkiG9w0B...",
    "ca_certificate": "-----BEGIN CERTIFICATE-----\nMIIDXTCCAkWgAwIBAgIJAKoK...",
    "expires_at": "2026-01-08T13:00:00Z",
    "serial_number": "ABC123456789",
    "fingerprint": "SHA256:1234567890abcdef..."
  },
  "request_id": "sync_req_003",
  "timestamp": "2025-01-08T13:10:00Z"
}
```

## Sincronización de Datos

### Modelo de Sincronización

La API Sync implementa un modelo de sincronización bidireccional que permite tanto la descarga de datos desde el servidor central (pull) como la carga de datos desde las terminales (push). El sistema utiliza timestamps vectoriales y checksums para detectar cambios y optimizar la transferencia de datos.

El proceso de sincronización se divide en varias fases: detección de cambios, preparación de deltas, transferencia de datos, aplicación de cambios, y resolución de conflictos. Cada fase incluye verificaciones de integridad y puntos de recuperación para garantizar la consistencia de los datos.

La sincronización soporta múltiples estrategias según el tipo de datos: sincronización completa para datos maestros críticos, sincronización incremental para transacciones, y sincronización bajo demanda para datos de gran volumen como imágenes de productos.

#### POST /api/v1/sync/pull

Descarga cambios desde el servidor central hacia la terminal.

**Headers:**
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
X-Terminal-ID: test-terminal-1
Content-Type: application/json
```

**Request Body:**
```json
{
  "last_sync_timestamp": "2025-01-08T12:00:00Z",
  "sync_types": [
    "productos",
    "precios",
    "categorias",
    "usuarios",
    "configuracion"
  ],
  "filters": {
    "sucursal_id": "test-sucursal-1",
    "only_active": true,
    "categories": ["test-category-1", "test-category-2"]
  },
  "options": {
    "include_images": false,
    "max_records": 1000,
    "compression": true
  }
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "sync_id": "sync_pull_001",
    "sync_timestamp": "2025-01-08T13:15:00Z",
    "changes": {
      "productos": {
        "total_changes": 5,
        "created": [
          {
            "id": "new-product-uuid",
            "codigo": "TEST009",
            "codigo_barras": "1234567890131",
            "nombre": "Taladro Inalámbrico 18V",
            "categoria_id": "test-category-1",
            "precio": 85000.00,
            "activo": true,
            "version": 1,
            "created_at": "2025-01-08T11:45:00Z",
            "updated_at": "2025-01-08T11:45:00Z"
          }
        ],
        "updated": [
          {
            "id": "test-product-1",
            "changes": {
              "precio": {
                "old_value": 15000.00,
                "new_value": 16000.00
              },
              "nombre": {
                "old_value": "Martillo Test 500g",
                "new_value": "Martillo Test 500g - Mejorado"
              }
            },
            "version": 3,
            "updated_at": "2025-01-08T11:50:00Z",
            "updated_by": "test-user-1"
          }
        ],
        "deleted": [
          {
            "id": "deleted-product-uuid",
            "deleted_at": "2025-01-08T10:00:00Z",
            "deleted_by": "test-user-1"
          }
        ]
      },
      "precios": {
        "total_changes": 3,
        "updated": [
          {
            "producto_id": "test-product-2",
            "sucursal_id": "test-sucursal-1",
            "precio_anterior": 3500.00,
            "precio_nuevo": 3800.00,
            "fecha_cambio": "2025-01-08T12:30:00Z",
            "motivo": "Ajuste por inflación"
          }
        ]
      },
      "configuracion": {
        "total_changes": 1,
        "updated": [
          {
            "clave": "impuesto_iva",
            "valor_anterior": "19",
            "valor_nuevo": "19.5",
            "fecha_cambio": "2025-01-08T09:00:00Z"
          }
        ]
      }
    },
    "metadata": {
      "total_records": 9,
      "compressed_size_kb": 45.2,
      "uncompressed_size_kb": 128.7,
      "checksum": "sha256:abcdef123456...",
      "next_sync_recommended": "2025-01-08T13:30:00Z"
    }
  },
  "request_id": "sync_req_004",
  "timestamp": "2025-01-08T13:15:00Z"
}
```

#### POST /api/v1/sync/push

Carga cambios desde la terminal hacia el servidor central.

**Request Body:**
```json
{
  "terminal_id": "test-terminal-1",
  "sync_timestamp": "2025-01-08T13:20:00Z",
  "changes": {
    "ventas": {
      "created": [
        {
          "id": "local-sale-001",
          "numero_documento": "LOCAL-001",
          "tipo_documento": "boleta",
          "fecha_venta": "2025-01-08T13:15:00Z",
          "sucursal_id": "test-sucursal-1",
          "terminal_id": "test-terminal-1",
          "cajero_id": "test-user-3",
          "vendedor_id": "test-user-2",
          "total": 18500.00,
          "estado": "procesado",
          "items": [
            {
              "producto_id": "test-product-2",
              "cantidad": 3,
              "precio_unitario": 3800.00,
              "subtotal_linea": 11400.00
            },
            {
              "producto_id": "test-product-3",
              "cantidad": 20,
              "precio_unitario": 150.00,
              "subtotal_linea": 3000.00
            }
          ],
          "medios_pago": [
            {
              "medio_pago": "tarjeta_credito",
              "monto": 18500.00,
              "referencia_transaccion": "TXN987654321"
            }
          ],
          "created_offline": true,
          "local_timestamp": "2025-01-08T13:15:00Z"
        }
      ]
    },
    "movimientos_stock": {
      "created": [
        {
          "id": "local-movement-001",
          "producto_id": "test-product-2",
          "sucursal_id": "test-sucursal-1",
          "tipo_movimiento": "salida",
          "cantidad": 3,
          "stock_anterior": 50,
          "stock_nuevo": 47,
          "motivo": "Venta",
          "documento_referencia": "LOCAL-001",
          "usuario_id": "test-user-3",
          "created_at": "2025-01-08T13:15:00Z"
        }
      ]
    }
  },
  "metadata": {
    "offline_duration_minutes": 45,
    "total_transactions": 1,
    "checksum": "sha256:fedcba654321..."
  }
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "sync_id": "sync_push_001",
    "processed_timestamp": "2025-01-08T13:20:00Z",
    "results": {
      "ventas": {
        "processed": 1,
        "accepted": 1,
        "rejected": 0,
        "conflicts": 0,
        "details": [
          {
            "local_id": "local-sale-001",
            "server_id": "server-sale-uuid",
            "status": "accepted",
            "numero_documento_final": "B-001-00000124"
          }
        ]
      },
      "movimientos_stock": {
        "processed": 1,
        "accepted": 1,
        "rejected": 0,
        "conflicts": 0
      }
    },
    "conflicts": [],
    "warnings": [
      {
        "type": "price_change",
        "message": "El precio del producto TEST002 cambió durante la venta offline",
        "details": {
          "producto_id": "test-product-2",
          "precio_venta": 3800.00,
          "precio_actual": 3850.00
        }
      }
    ],
    "next_pull_recommended": true,
    "server_timestamp": "2025-01-08T13:20:00Z"
  },
  "request_id": "sync_req_005",
  "timestamp": "2025-01-08T13:20:00Z"
}
```

### Sincronización Incremental

#### GET /api/v1/sync/changes/{entity_type}

Obtiene cambios incrementales para un tipo específico de entidad.

**Path Parameters:**
- `entity_type` (string, requerido): Tipo de entidad (productos, precios, stock, etc.)

**Query Parameters:**
- `since` (timestamp, requerido): Timestamp desde el cual obtener cambios
- `limit` (int, opcional): Máximo número de registros (default: 100, max: 1000)
- `include_deleted` (bool, opcional): Incluir registros eliminados (default: true)
- `sucursal_id` (uuid, opcional): Filtrar por sucursal específica

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "entity_type": "productos",
    "changes": [
      {
        "id": "test-product-1",
        "operation": "update",
        "timestamp": "2025-01-08T11:50:00Z",
        "version": 3,
        "data": {
          "nombre": "Martillo Test 500g - Mejorado",
          "precio": 16000.00
        },
        "changed_fields": ["nombre", "precio"],
        "changed_by": "test-user-1"
      }
    ],
    "metadata": {
      "total_changes": 1,
      "has_more": false,
      "next_cursor": null,
      "last_change_timestamp": "2025-01-08T11:50:00Z"
    }
  },
  "request_id": "sync_req_006",
  "timestamp": "2025-01-08T13:25:00Z"
}
```


## Gestión de Conflictos

### Detección y Resolución Automática

El sistema de gestión de conflictos de la API Sync está diseñado para manejar situaciones donde los mismos datos han sido modificados tanto en el servidor central como en las terminales durante períodos de desconexión. El sistema utiliza algoritmos sofisticados de detección de conflictos basados en timestamps vectoriales, checksums de contenido, y análisis semántico de los cambios.

Los conflictos se clasifican en diferentes categorías según su naturaleza y criticidad: conflictos de datos maestros (productos, precios), conflictos transaccionales (ventas duplicadas), y conflictos de configuración (parámetros del sistema). Cada categoría tiene estrategias de resolución específicas que pueden ser automáticas o requerir intervención manual.

El sistema implementa reglas de resolución configurables que pueden adaptarse a las necesidades específicas del negocio. Por ejemplo, los cambios de precios siempre se resuelven a favor del servidor central para mantener consistencia, mientras que las ventas locales tienen prioridad para evitar pérdida de transacciones.

#### GET /api/v1/sync/conflicts

Obtiene la lista de conflictos pendientes de resolución.

**Headers:**
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
X-Terminal-ID: test-terminal-1
```

**Query Parameters:**
- `status` (string, opcional): Estado del conflicto (pending, resolved, ignored)
- `entity_type` (string, opcional): Tipo de entidad en conflicto
- `priority` (string, opcional): Prioridad del conflicto (high, medium, low)
- `created_since` (timestamp, opcional): Conflictos creados desde fecha específica

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "conflicts": [
      {
        "id": "conflict_001",
        "entity_type": "producto",
        "entity_id": "test-product-1",
        "conflict_type": "concurrent_modification",
        "priority": "medium",
        "status": "pending",
        "created_at": "2025-01-08T13:10:00Z",
        "description": "Producto modificado simultáneamente en servidor y terminal",
        "local_version": {
          "version": 2,
          "timestamp": "2025-01-08T12:30:00Z",
          "changes": {
            "precio": 15500.00,
            "descripcion": "Martillo mejorado con mango ergonómico"
          },
          "changed_by": "test-user-2",
          "terminal_id": "test-terminal-1"
        },
        "server_version": {
          "version": 3,
          "timestamp": "2025-01-08T11:50:00Z",
          "changes": {
            "precio": 16000.00,
            "nombre": "Martillo Test 500g - Mejorado"
          },
          "changed_by": "test-user-1",
          "source": "central_server"
        },
        "suggested_resolution": {
          "strategy": "merge_non_conflicting",
          "merged_data": {
            "precio": 16000.00,
            "nombre": "Martillo Test 500g - Mejorado",
            "descripcion": "Martillo mejorado con mango ergonómico"
          },
          "conflicts_remaining": [],
          "confidence": 0.95
        },
        "resolution_options": [
          {
            "id": "accept_server",
            "description": "Aceptar cambios del servidor",
            "impact": "Se perderán los cambios locales de descripción"
          },
          {
            "id": "accept_local",
            "description": "Aceptar cambios locales",
            "impact": "Se perderán los cambios del servidor en nombre y precio"
          },
          {
            "id": "merge_manual",
            "description": "Fusionar manualmente",
            "impact": "Requiere revisión manual de cada campo"
          }
        ]
      }
    ],
    "summary": {
      "total_conflicts": 1,
      "by_priority": {
        "high": 0,
        "medium": 1,
        "low": 0
      },
      "by_type": {
        "concurrent_modification": 1,
        "duplicate_transaction": 0,
        "data_integrity": 0
      }
    }
  },
  "request_id": "sync_req_007",
  "timestamp": "2025-01-08T13:30:00Z"
}
```

#### POST /api/v1/sync/conflicts/{conflict_id}/resolve

Resuelve un conflicto específico aplicando una estrategia de resolución.

**Path Parameters:**
- `conflict_id` (string, requerido): ID del conflicto a resolver

**Request Body:**
```json
{
  "resolution_strategy": "merge_manual",
  "resolved_data": {
    "precio": 16000.00,
    "nombre": "Martillo Test 500g - Mejorado",
    "descripcion": "Martillo mejorado con mango ergonómico"
  },
  "resolution_notes": "Fusionado manualmente: precio del servidor, descripción de terminal",
  "resolved_by": "test-user-4"
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "conflict_id": "conflict_001",
    "resolution_status": "resolved",
    "resolution_timestamp": "2025-01-08T13:35:00Z",
    "applied_changes": {
      "entity_id": "test-product-1",
      "new_version": 4,
      "changes_applied": {
        "precio": 16000.00,
        "nombre": "Martillo Test 500g - Mejorado",
        "descripcion": "Martillo mejorado con mango ergonómico"
      }
    },
    "propagation": {
      "to_terminals": ["test-terminal-1", "test-terminal-2"],
      "estimated_sync_time": "2025-01-08T13:40:00Z"
    }
  },
  "request_id": "sync_req_008",
  "timestamp": "2025-01-08T13:35:00Z"
}
```

### Resolución Automática

#### POST /api/v1/sync/conflicts/auto-resolve

Ejecuta resolución automática de conflictos según reglas predefinidas.

**Request Body:**
```json
{
  "conflict_ids": ["conflict_001", "conflict_002"],
  "auto_resolve_rules": {
    "price_conflicts": "prefer_server",
    "description_conflicts": "prefer_local",
    "stock_conflicts": "sum_values",
    "transaction_conflicts": "keep_both"
  },
  "max_confidence_threshold": 0.8
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "auto_resolved": [
      {
        "conflict_id": "conflict_002",
        "strategy_applied": "prefer_server",
        "confidence": 0.95,
        "resolution_time_ms": 45
      }
    ],
    "manual_review_required": [
      {
        "conflict_id": "conflict_001",
        "reason": "Confidence below threshold",
        "confidence": 0.75,
        "recommended_action": "manual_review"
      }
    ],
    "summary": {
      "total_processed": 2,
      "auto_resolved": 1,
      "manual_required": 1,
      "processing_time_ms": 123
    }
  },
  "request_id": "sync_req_009",
  "timestamp": "2025-01-08T13:40:00Z"
}
```

## Heartbeat y Estado

### Monitoreo de Conectividad

El sistema de heartbeat mantiene un monitoreo continuo del estado de conectividad y salud de las terminales. Cada terminal debe enviar señales de vida periódicas que incluyen información sobre su estado operativo, métricas de rendimiento, y cualquier problema detectado localmente.

El heartbeat no solo confirma que la terminal está online, sino que también proporciona información valiosa sobre la calidad de la conexión, latencia de red, y estado de los servicios locales. Esta información se utiliza para optimizar los intervalos de sincronización y detectar problemas antes de que afecten las operaciones.

#### POST /api/v1/sync/heartbeat

Envía señal de vida desde la terminal al servidor central.

**Headers:**
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
X-Terminal-ID: test-terminal-1
```

**Request Body:**
```json
{
  "terminal_id": "test-terminal-1",
  "timestamp": "2025-01-08T13:45:00Z",
  "status": "online",
  "health": {
    "cpu_usage": 15.2,
    "memory_usage": 68.5,
    "disk_usage": 45.8,
    "network_latency_ms": 23,
    "last_sync_success": "2025-01-08T13:20:00Z",
    "pending_transactions": 0,
    "offline_duration_minutes": 0
  },
  "services": {
    "pos_application": "running",
    "local_database": "running",
    "printer_service": "running",
    "barcode_scanner": "connected",
    "cash_drawer": "connected"
  },
  "metrics": {
    "transactions_today": 15,
    "last_transaction": "2025-01-08T13:15:00Z",
    "average_transaction_time_seconds": 45.2,
    "errors_last_hour": 0
  },
  "alerts": [
    {
      "level": "warning",
      "message": "Impresora con papel bajo",
      "timestamp": "2025-01-08T13:30:00Z",
      "code": "PRINTER_LOW_PAPER"
    }
  ]
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "heartbeat_received": "2025-01-08T13:45:00Z",
    "terminal_status": "healthy",
    "next_heartbeat_expected": "2025-01-08T13:50:00Z",
    "server_instructions": {
      "sync_required": false,
      "config_update_available": false,
      "maintenance_window": null
    },
    "server_time": "2025-01-08T13:45:00Z",
    "time_drift_seconds": 0.5,
    "network_quality": {
      "latency_ms": 23,
      "bandwidth_mbps": 45.2,
      "quality_score": 0.95
    },
    "alerts_acknowledged": ["PRINTER_LOW_PAPER"],
    "recommendations": [
      {
        "type": "maintenance",
        "message": "Reemplazar papel de impresora",
        "priority": "low",
        "due_date": "2025-01-09T09:00:00Z"
      }
    ]
  },
  "request_id": "sync_req_010",
  "timestamp": "2025-01-08T13:45:00Z"
}
```

#### GET /api/v1/sync/terminal/status

Obtiene el estado detallado de la terminal desde el servidor.

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "terminal_id": "test-terminal-1",
    "status": "online",
    "last_heartbeat": "2025-01-08T13:45:00Z",
    "uptime": "72h30m15s",
    "connection_quality": "excellent",
    "sync_status": {
      "last_pull": "2025-01-08T13:15:00Z",
      "last_push": "2025-01-08T13:20:00Z",
      "pending_changes": 0,
      "conflicts_pending": 1,
      "next_scheduled_sync": "2025-01-08T14:00:00Z"
    },
    "performance_metrics": {
      "average_response_time_ms": 145,
      "success_rate_24h": 99.8,
      "transactions_processed_today": 15,
      "data_transferred_mb": 12.5
    },
    "health_indicators": {
      "overall_score": 0.95,
      "cpu_health": "good",
      "memory_health": "good",
      "disk_health": "good",
      "network_health": "excellent"
    },
    "configuration": {
      "sync_interval_minutes": 15,
      "offline_mode_enabled": true,
      "auto_sync_enabled": true,
      "max_offline_transactions": 1000
    }
  },
  "request_id": "sync_req_011",
  "timestamp": "2025-01-08T13:50:00Z"
}
```

### Gestión de Estado Offline

#### POST /api/v1/sync/offline/enter

Notifica al servidor que la terminal está entrando en modo offline.

**Request Body:**
```json
{
  "terminal_id": "test-terminal-1",
  "reason": "network_disconnection",
  "estimated_duration_minutes": 30,
  "offline_capabilities": {
    "can_process_sales": true,
    "can_accept_returns": false,
    "can_modify_prices": false,
    "max_transaction_amount": 100000.00
  },
  "last_sync_data": {
    "products_count": 1250,
    "last_product_sync": "2025-01-08T13:15:00Z",
    "prices_last_update": "2025-01-08T12:00:00Z"
  }
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "offline_session_id": "offline_session_001",
    "offline_started": "2025-01-08T14:00:00Z",
    "offline_permissions": {
      "max_transaction_amount": 100000.00,
      "can_process_sales": true,
      "can_accept_returns": false,
      "requires_supervisor_approval": ["price_override", "large_discount"]
    },
    "sync_instructions": {
      "queue_all_transactions": true,
      "compress_data": true,
      "priority_sync_on_reconnect": ["sales", "stock_movements"]
    },
    "estimated_reconnect": "2025-01-08T14:30:00Z"
  },
  "request_id": "sync_req_012",
  "timestamp": "2025-01-08T14:00:00Z"
}
```

#### POST /api/v1/sync/offline/exit

Notifica al servidor que la terminal está saliendo del modo offline.

**Request Body:**
```json
{
  "terminal_id": "test-terminal-1",
  "offline_session_id": "offline_session_001",
  "offline_duration_minutes": 25,
  "transactions_queued": 3,
  "data_to_sync": {
    "sales": 3,
    "stock_movements": 8,
    "configuration_changes": 0
  },
  "offline_summary": {
    "total_sales_amount": 45000.00,
    "transactions_processed": 3,
    "errors_encountered": 0,
    "warnings_generated": 1
  }
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "reconnection_confirmed": "2025-01-08T14:25:00Z",
    "sync_priority": "high",
    "immediate_sync_required": true,
    "estimated_sync_time_minutes": 5,
    "server_changes_pending": {
      "products": 2,
      "prices": 1,
      "configuration": 0
    },
    "sync_schedule": {
      "immediate": ["sales", "stock_movements"],
      "next_cycle": ["products", "prices"],
      "background": ["logs", "metrics"]
    }
  },
  "request_id": "sync_req_013",
  "timestamp": "2025-01-08T14:25:00Z"
}
```

## Monitoreo y Métricas

### Métricas de Sincronización

#### GET /api/v1/sync/metrics

Obtiene métricas detalladas del proceso de sincronización.

**Query Parameters:**
- `period` (string, opcional): Período de métricas (hour, day, week, month)
- `terminal_id` (uuid, opcional): Métricas de terminal específica
- `metric_types` (array, opcional): Tipos de métricas específicas

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "period": "day",
    "start_time": "2025-01-08T00:00:00Z",
    "end_time": "2025-01-08T23:59:59Z",
    "sync_metrics": {
      "total_sync_operations": 96,
      "successful_syncs": 94,
      "failed_syncs": 2,
      "success_rate": 97.9,
      "average_sync_duration_seconds": 12.5,
      "data_transferred_mb": 145.8,
      "conflicts_detected": 3,
      "conflicts_resolved": 2
    },
    "terminal_metrics": [
      {
        "terminal_id": "test-terminal-1",
        "sync_count": 24,
        "success_rate": 100.0,
        "average_duration_seconds": 8.2,
        "data_transferred_mb": 35.2,
        "last_sync": "2025-01-08T13:45:00Z",
        "status": "healthy"
      }
    ],
    "performance_trends": {
      "sync_duration_trend": "stable",
      "data_volume_trend": "increasing",
      "error_rate_trend": "decreasing",
      "network_quality_trend": "improving"
    }
  },
  "request_id": "sync_req_014",
  "timestamp": "2025-01-08T15:00:00Z"
}
```

