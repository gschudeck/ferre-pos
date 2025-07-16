# Módulo de Sistema - Sistema Ferre-POS

## 📋 Descripción General

El módulo de Sistema proporciona funcionalidades administrativas centrales para la gestión, configuración y monitoreo del sistema Ferre-POS. Incluye gestión de configuraciones globales, monitoreo de rendimiento, administración de logs, backup/restauración y utilidades de mantenimiento.

## 🎯 Funcionalidades Principales

### Gestión de Configuraciones

1. **Configuraciones Globales**
   - Parámetros del sistema organizados por categorías
   - Tipos de datos validados (string, number, integer, boolean, json)
   - Configuraciones de solo lectura para valores críticos
   - Auditoría completa de cambios

2. **Categorías de Configuración**
   - **Empresa**: Datos de la empresa (nombre, RUT, dirección)
   - **Impuestos**: Configuración de IVA y otros impuestos
   - **General**: Configuraciones generales del sistema
   - **Sistema**: Parámetros técnicos del sistema
   - **Seguridad**: Configuraciones de seguridad

### Monitoreo y Métricas

1. **Health Check**
   - Estado de la aplicación
   - Conectividad de base de datos
   - Uso de memoria y CPU
   - Espacio en disco

2. **Métricas de Rendimiento**
   - Uso de memoria detallado
   - Estadísticas de CPU
   - Conexiones de base de datos
   - Tiempo de actividad (uptime)

3. **Estadísticas del Sistema**
   - Contadores de entidades principales
   - Tamaño de base de datos
   - Actividad diaria

### Administración de Logs

1. **Consulta de Logs**
   - Filtrado por nivel, fecha, usuario, módulo
   - Paginación y búsqueda
   - Exportación de logs

2. **Limpieza de Logs**
   - Eliminación automática de logs antiguos
   - Configuración de períodos de retención
   - Optimización de espacio

### Backup y Restauración

1. **Backup de Configuraciones**
   - Exportación completa de configuraciones
   - Archivos JSON con metadatos
   - Versionado automático

2. **Restauración**
   - Importación selectiva de configuraciones
   - Validación de integridad
   - Reporte de errores y conflictos

### Mantenimiento del Sistema

1. **Tareas Automatizadas**
   - Limpieza de logs
   - Optimización de base de datos
   - Backup automático
   - Verificación de integridad

2. **Reinicio Controlado**
   - Reinicio graceful del sistema
   - Registro de motivos
   - Notificación a usuarios

## 🔧 API Endpoints

### Información del Sistema

```http
GET /api/sistema/info
Authorization: Bearer <token>
```

**Respuesta:**
```json
{
  "success": true,
  "data": {
    "version": "1.0.0",
    "nombre": "Ferre-POS API",
    "descripcion": "Sistema de Punto de Venta para Ferreterías",
    "ambiente": "production",
    "fecha_inicio": "2024-01-15T10:00:00.000Z",
    "uptime": 86400,
    "memoria": {
      "rss": 45678592,
      "heapTotal": 32456789,
      "heapUsed": 28123456,
      "external": 1234567
    },
    "plataforma": "linux",
    "version_node": "v18.17.0",
    "base_datos": {
      "total_usuarios": 25,
      "total_productos": 1500,
      "total_ventas_hoy": 45,
      "total_sucursales": 3,
      "total_notas_activas": 12,
      "tamaño_bd": "125 MB"
    },
    "configuraciones": {
      "empresa_nombre": "Ferretería Central",
      "empresa_rut": "76123456-7",
      "iva_porcentaje": 19,
      "moneda_codigo": "CLP",
      "backup_automatico": true,
      "logs_retention_days": 30
    }
  }
}
```

### Health Check

```http
GET /api/sistema/health
Authorization: Bearer <token>
```

**Respuesta:**
```json
{
  "status": "ok",
  "timestamp": "2024-01-15T15:30:00.000Z",
  "version": "1.0.0",
  "environment": "production",
  "uptime": 86400,
  "memory": {
    "rss": 45678592,
    "heapTotal": 32456789,
    "heapUsed": 28123456,
    "external": 1234567
  },
  "database": {
    "status": "healthy",
    "message": "Conexión exitosa"
  },
  "disk": {
    "status": "healthy",
    "message": "Espacio disponible"
  }
}
```

### Gestión de Configuraciones

#### Obtener Todas las Configuraciones

```http
GET /api/sistema/configuraciones?incluir_inactivas=false
Authorization: Bearer <token>
```

**Respuesta:**
```json
{
  "success": true,
  "data": {
    "empresa": {
      "empresa_nombre": {
        "valor": "Ferretería Central",
        "descripcion": "Nombre de la empresa",
        "tipo_dato": "string",
        "activa": true,
        "solo_lectura": false,
        "fecha_creacion": "2024-01-01T00:00:00.000Z",
        "fecha_modificacion": "2024-01-15T10:00:00.000Z"
      },
      "empresa_rut": {
        "valor": "76123456-7",
        "descripcion": "RUT de la empresa",
        "tipo_dato": "string",
        "activa": true,
        "solo_lectura": false
      }
    },
    "impuestos": {
      "iva_porcentaje": {
        "valor": 19,
        "descripcion": "Porcentaje de IVA",
        "tipo_dato": "number",
        "activa": true,
        "solo_lectura": false
      }
    }
  }
}
```

#### Obtener Configuración Específica

```http
GET /api/sistema/configuraciones/empresa_nombre
Authorization: Bearer <token>
```

#### Crear Nueva Configuración

```http
POST /api/sistema/configuraciones
Authorization: Bearer <token>
Content-Type: application/json

{
  "clave": "nueva_configuracion",
  "valor": "valor_inicial",
  "descripcion": "Descripción de la nueva configuración",
  "tipo_dato": "string",
  "categoria": "general",
  "solo_lectura": false
}
```

#### Actualizar Configuración

```http
PUT /api/sistema/configuraciones/empresa_nombre
Authorization: Bearer <token>
Content-Type: application/json

{
  "valor": "Nuevo Nombre de Empresa"
}
```

#### Eliminar Configuración

```http
DELETE /api/sistema/configuraciones/configuracion_temporal
Authorization: Bearer <token>
```

### Backup y Restauración

#### Crear Backup

```http
POST /api/sistema/backup
Authorization: Bearer <token>
```

**Respuesta:**
```json
{
  "success": true,
  "message": "Backup creado exitosamente",
  "data": {
    "archivo": "config_backup_2024-01-15_1705329600000.json",
    "ruta": "/app/backups/config_backup_2024-01-15_1705329600000.json",
    "tamaño": 15678,
    "fecha": "2024-01-15T15:30:00.000Z"
  }
}
```

#### Restaurar desde Backup

```http
POST /api/sistema/restaurar
Authorization: Bearer <token>
Content-Type: application/json

{
  "fecha": "2024-01-15T15:30:00.000Z",
  "version": "1.0.0",
  "configuraciones": {
    "empresa": {
      "empresa_nombre": {
        "valor": "Ferretería Restaurada",
        "descripcion": "Nombre de la empresa",
        "tipo_dato": "string",
        "activa": true,
        "solo_lectura": false
      }
    }
  }
}
```

### Administración de Logs

#### Obtener Logs

```http
GET /api/sistema/logs?nivel=error&fecha_inicio=2024-01-01&limit=50
Authorization: Bearer <token>
```

**Parámetros de consulta:**
- `nivel`: error, warn, info, debug
- `fecha_inicio`: Fecha de inicio (YYYY-MM-DD)
- `fecha_fin`: Fecha de fin (YYYY-MM-DD)
- `usuario`: ID del usuario
- `modulo`: Módulo del sistema
- `page`: Número de página (default: 1)
- `limit`: Logs por página (default: 100, max: 1000)

#### Limpiar Logs Antiguos

```http
POST /api/sistema/logs/limpiar
Authorization: Bearer <token>
Content-Type: application/json

{
  "dias_retencion": 30
}
```

### Métricas y Estadísticas

#### Obtener Estadísticas

```http
GET /api/sistema/estadisticas
Authorization: Bearer <token>
```

#### Obtener Métricas de Rendimiento

```http
GET /api/sistema/metricas
Authorization: Bearer <token>
```

**Respuesta:**
```json
{
  "success": true,
  "data": {
    "memoria": {
      "rss": 45678592,
      "heapTotal": 32456789,
      "heapUsed": 28123456,
      "external": 1234567
    },
    "cpu": {
      "user": 123456,
      "system": 78901
    },
    "uptime": 86400,
    "timestamp": "2024-01-15T15:30:00.000Z",
    "base_datos": {
      "conexiones_activas": 5,
      "conexiones_totales": 10
    }
  }
}
```

### Mantenimiento del Sistema

#### Ejecutar Mantenimiento

```http
POST /api/sistema/mantenimiento
Authorization: Bearer <token>
Content-Type: application/json

{
  "limpiar_logs": true,
  "dias_retencion_logs": 30,
  "optimizar_bd": true,
  "backup_configuraciones": true
}
```

**Respuesta:**
```json
{
  "success": true,
  "message": "Mantenimiento ejecutado",
  "data": {
    "timestamp": "2024-01-15T15:30:00.000Z",
    "tareas_ejecutadas": [
      {
        "tarea": "limpiar_logs",
        "resultado": {
          "logsEliminados": 1500,
          "fechaLimite": "2023-12-16T15:30:00.000Z"
        }
      },
      {
        "tarea": "backup_configuraciones",
        "resultado": {
          "archivo": "config_backup_2024-01-15_1705329600000.json",
          "tamaño": 15678
        }
      },
      {
        "tarea": "optimizar_bd",
        "resultado": {
          "mensaje": "Base de datos optimizada exitosamente"
        }
      }
    ],
    "errores": []
  }
}
```

#### Reiniciar Sistema

```http
POST /api/sistema/reiniciar
Authorization: Bearer <token>
Content-Type: application/json

{
  "motivo": "Actualización de configuraciones críticas del sistema"
}
```

## 🔒 Permisos y Seguridad

### Roles y Accesos

| Acción | Admin | Gerente | Vendedor | Cajero |
|--------|-------|---------|----------|--------|
| Ver info sistema | ✅ | ✅ | ✅ | ✅ |
| Health check | ✅ | ✅ | ✅ | ✅ |
| Ver estadísticas | ✅ | ✅ | ❌ | ❌ |
| Ver métricas | ✅ | ✅ | ❌ | ❌ |
| Ver configuraciones | ✅ | ❌ | ❌ | ❌ |
| Crear configuraciones | ✅ | ❌ | ❌ | ❌ |
| Modificar configuraciones | ✅ | ❌ | ❌ | ❌ |
| Eliminar configuraciones | ✅ | ❌ | ❌ | ❌ |
| Crear backup | ✅ | ❌ | ❌ | ❌ |
| Restaurar backup | ✅ | ❌ | ❌ | ❌ |
| Ver logs | ✅ | ✅ | ❌ | ❌ |
| Limpiar logs | ✅ | ❌ | ❌ | ❌ |
| Ejecutar mantenimiento | ✅ | ❌ | ❌ | ❌ |
| Reiniciar sistema | ✅ | ❌ | ❌ | ❌ |

### Validaciones de Seguridad

1. **Autenticación JWT**: Todas las rutas requieren token válido
2. **Autorización granular**: Permisos específicos por rol
3. **Validación de datos**: Esquemas Joi para todas las entradas
4. **Auditoría completa**: Registro de todos los cambios críticos
5. **Configuraciones protegidas**: Algunas configuraciones son de solo lectura

## 📊 Configuraciones del Sistema

### Categorías Principales

#### Empresa
- `empresa_nombre`: Nombre de la empresa
- `empresa_rut`: RUT de la empresa
- `empresa_direccion`: Dirección principal
- `empresa_telefono`: Teléfono de contacto
- `empresa_email`: Email de contacto
- `empresa_sitio_web`: Sitio web corporativo

#### Impuestos
- `iva_porcentaje`: Porcentaje de IVA (default: 19)
- `impuesto_adicional`: Impuestos adicionales
- `exento_iva`: Productos exentos de IVA

#### General
- `moneda_codigo`: Código de moneda (default: CLP)
- `moneda_simbolo`: Símbolo de moneda (default: $)
- `idioma_sistema`: Idioma del sistema (default: es)
- `zona_horaria`: Zona horaria (default: America/Santiago)

#### Sistema
- `sistema_version`: Versión del sistema (solo lectura)
- `backup_automatico`: Backup automático habilitado
- `logs_retention_days`: Días de retención de logs
- `session_timeout`: Tiempo de expiración de sesión
- `max_login_attempts`: Intentos máximos de login

#### Seguridad
- `password_min_length`: Longitud mínima de contraseña
- `password_require_special`: Requerir caracteres especiales
- `jwt_expiration`: Tiempo de expiración de JWT
- `rate_limit_requests`: Límite de requests por minuto

### Tipos de Datos Soportados

1. **string**: Texto simple
2. **number**: Números decimales
3. **integer**: Números enteros
4. **boolean**: Verdadero/Falso
5. **json**: Objetos JSON complejos

## 🔄 Flujos de Trabajo

### Flujo de Configuración

1. **Creación**: Admin crea nueva configuración
2. **Validación**: Sistema valida tipo de dato y formato
3. **Almacenamiento**: Configuración se guarda en BD
4. **Auditoría**: Cambio se registra en auditoría
5. **Aplicación**: Configuración está disponible inmediatamente

### Flujo de Backup

1. **Solicitud**: Admin solicita backup
2. **Extracción**: Sistema extrae todas las configuraciones
3. **Empaquetado**: Datos se empaquetan en JSON
4. **Almacenamiento**: Archivo se guarda en directorio de backups
5. **Notificación**: Admin recibe confirmación con detalles

### Flujo de Mantenimiento

1. **Programación**: Admin programa tareas de mantenimiento
2. **Ejecución**: Sistema ejecuta tareas en orden
3. **Monitoreo**: Progreso se registra en logs
4. **Reporte**: Resultados se entregan al admin
5. **Limpieza**: Recursos temporales se liberan

## 📈 Monitoreo y Alertas

### Métricas Clave

1. **Rendimiento**
   - Uso de memoria (RSS, Heap)
   - Uso de CPU
   - Tiempo de respuesta de APIs
   - Conexiones de base de datos

2. **Disponibilidad**
   - Uptime del sistema
   - Estado de servicios críticos
   - Conectividad de base de datos
   - Espacio en disco

3. **Actividad**
   - Requests por minuto
   - Usuarios activos
   - Transacciones por hora
   - Errores por minuto

### Alertas Automáticas

1. **Críticas**
   - Caída del sistema
   - Error de base de datos
   - Espacio en disco bajo
   - Memoria insuficiente

2. **Advertencias**
   - Alto uso de CPU
   - Muchas conexiones de BD
   - Logs creciendo rápidamente
   - Backup fallido

## 🛠️ Configuración

### Variables de Entorno

```env
# Configuración de logs
LOG_LEVEL=info
LOG_RETENTION_DAYS=30
LOG_MAX_SIZE=100MB

# Configuración de backup
BACKUP_ENABLED=true
BACKUP_SCHEDULE=0 2 * * *
BACKUP_RETENTION_DAYS=90

# Configuración de mantenimiento
MAINTENANCE_WINDOW=02:00-04:00
AUTO_OPTIMIZE_DB=true
AUTO_CLEANUP_LOGS=true
```

### Configuración de Base de Datos

El módulo de sistema utiliza las siguientes tablas:

- `configuracion_sistema`: Configuraciones principales
- `auditoria_configuracion`: Auditoría de cambios
- `system_logs`: Logs del sistema (opcional)

## 🧪 Testing

### Ejecutar Tests

```bash
# Tests específicos del módulo
npm test tests/sistema.test.js

# Tests de integración
npm run test:integration

# Coverage del módulo
npm run test:coverage -- tests/sistema.test.js
```

### Casos de Prueba Incluidos

1. **Información del sistema** y health checks
2. **CRUD completo** de configuraciones
3. **Validación de tipos** de datos
4. **Permisos por rol** de usuario
5. **Backup y restauración** completos
6. **Administración de logs** y limpieza
7. **Mantenimiento del sistema** y tareas
8. **Métricas y estadísticas** del sistema

## 🔧 Mantenimiento

### Tareas Programadas Recomendadas

1. **Backup de configuraciones**: Diario a las 02:00
2. **Limpieza de logs**: Semanal los domingos
3. **Optimización de BD**: Mensual el primer domingo
4. **Verificación de salud**: Cada 5 minutos

### Scripts de Utilidad

```bash
# Backup manual de configuraciones
node scripts/backup-config.js

# Limpieza manual de logs
node scripts/cleanup-logs.js --days 30

# Verificación de salud del sistema
node scripts/health-check.js

# Optimización de base de datos
node scripts/optimize-db.js
```

### Monitoreo Continuo

- **Logs de aplicación**: Monitoreo en tiempo real
- **Métricas de sistema**: Recolección cada minuto
- **Alertas automáticas**: Notificaciones por email/Slack
- **Dashboard de salud**: Interfaz web para monitoreo

## 📞 Soporte

Para soporte específico del módulo de sistema:

- **Documentación técnica**: `/docs/api` en el servidor
- **Logs del sistema**: `logs/ferre-pos-api.log`
- **Health check**: `GET /api/sistema/health`
- **Métricas en vivo**: `GET /api/sistema/metricas`
- **Contacto**: soporte@ferre-pos.cl

---

**Módulo desarrollado para administración completa del sistema Ferre-POS** ⚙️🔧

