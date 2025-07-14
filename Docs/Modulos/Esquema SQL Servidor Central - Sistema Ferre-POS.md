# Esquema SQL Servidor Central - Sistema Ferre-POS

## Descripción General

Este esquema SQL implementa la base de datos completa para el servidor central del sistema Ferre-POS, diseñado específicamente para ferreterías urbanas con operación multisucursal. El esquema está optimizado para PostgreSQL y soporta todas las funcionalidades descritas en la documentación técnica unificada.

## Características Principales

### ✅ Arquitectura Completa
- **28 tablas principales** con relaciones bien definidas
- **Tipos de datos personalizados** para mayor consistencia
- **Integridad referencial** completa con constraints
- **Campos calculados** para optimización de consultas

### ✅ Funcionalidades Implementadas
- **Gestión multisucursal** con configuraciones específicas
- **Control de inventario** distribuido con sincronización
- **Sistema de fidelización** completo con reglas configurables
- **Documentos tributarios electrónicos** (DTE) con proveedores
- **Control de despacho** y trazabilidad
- **Auditoría y seguridad** integral
- **Notas de crédito** con autorización de supervisores
- **Reimpresión controlada** de documentos

### ✅ Optimización y Rendimiento
- **70+ índices especializados** para consultas frecuentes
- **Índices de texto completo** para búsquedas aproximadas
- **Vistas materializadas** para reportes complejos
- **Particionado** recomendado para tablas de alto volumen
- **Triggers automáticos** para lógica de negocio crítica

### ✅ Seguridad y Cumplimiento
- **Roles granulares** de base de datos
- **Auditoría completa** de operaciones sensibles
- **Logs de seguridad** detallados
- **Cumplimiento normativo** SII y Ley 21.719
- **Gestión de sesiones** con expiración automática

## Estructura de Tablas

### Tablas Principales
| Tabla | Descripción | Registros Estimados |
|-------|-------------|-------------------|
| `sucursales` | Información de sucursales | 3-50 |
| `usuarios` | Usuarios del sistema | 50-500 |
| `productos` | Catálogo de productos | 5,000-50,000 |
| `stock_central` | Inventario consolidado | 150,000-2,500,000 |
| `ventas` | Transacciones de venta | 1,000-10,000/día |
| `detalle_ventas` | Productos vendidos | 3,000-30,000/día |
| `fidelizacion_clientes` | Clientes registrados | 1,000-100,000 |
| `documentos_dte` | Documentos tributarios | 1,000-10,000/día |

### Tablas de Soporte
| Tabla | Descripción | Propósito |
|-------|-------------|-----------|
| `movimientos_stock` | Historial de inventario | Trazabilidad |
| `movimientos_fidelizacion` | Historial de puntos | Auditoría |
| `logs_sincronizacion` | Sincronización sucursales | Monitoreo |
| `logs_seguridad` | Eventos de seguridad | Auditoría |
| `configuracion_sistema` | Parámetros globales | Configuración |

## Funciones Especializadas

### Gestión de Stock
- `validar_stock_disponible()` - Validación antes de ventas
- `descontar_stock_venta()` - Descuento automático
- `restaurar_stock_anulacion()` - Restauración por anulaciones

### Fidelización
- `calcular_puntos_fidelizacion()` - Cálculo según reglas
- `acumular_puntos_fidelizacion()` - Acumulación automática
- `procesar_canje_puntos()` - Procesamiento de canjes

### Documentos Tributarios
- `generar_folio_dte()` - Generación de folios consecutivos

### Mantenimiento
- `mantenimiento_diario()` - Rutinas automáticas
- `limpiar_sesiones_expiradas()` - Limpieza de sesiones
- `actualizar_niveles_fidelizacion()` - Actualización de niveles

## Vistas Especializadas

### Operacionales
- `vista_stock_consolidado` - Stock con estado crítico/normal
- `vista_ventas_resumen_diario` - Resumen de ventas por día
- `vista_productos_mas_vendidos` - Top productos por período

### Analíticas
- `vista_fidelizacion_resumen_cliente` - Resumen completo de clientes
- `vista_documentos_dte_pendientes` - DTEs pendientes de procesamiento

## Instalación y Configuración

### Requisitos Previos
- PostgreSQL 14 o superior
- Extensiones: `uuid-ossp`, `pg_trgm`, `btree_gin`
- Memoria RAM: mínimo 4GB, recomendado 8GB+
- Almacenamiento: SSD recomendado para mejor rendimiento

### Pasos de Instalación

1. **Crear base de datos**
```sql
CREATE DATABASE ferre_pos_central;
```

2. **Ejecutar esquema**
```bash
psql -d ferre_pos_central -f ferre_pos_servidor_central_schema.sql
```

3. **Configurar postgresql.conf** (ver recomendaciones en el script)

4. **Crear usuarios de aplicación**
```sql
CREATE USER ferre_pos_app_user WITH PASSWORD 'password_seguro';
GRANT ferre_pos_app TO ferre_pos_app_user;
```

### Configuración Inicial

1. **Cargar sucursales**
```sql
INSERT INTO sucursales (codigo, nombre, direccion, comuna, region) 
VALUES ('SUC001', 'Sucursal Centro', 'Av. Principal 123', 'Santiago', 'Metropolitana');
```

2. **Crear usuario administrador**
```sql
INSERT INTO usuarios (rut, nombre, rol, sucursal_id, password_hash, salt)
VALUES ('12345678-9', 'Admin Sistema', 'admin', 
        (SELECT id FROM sucursales WHERE codigo = 'SUC001'),
        'hash_password', 'salt_value');
```

3. **Configurar proveedores DTE**
```sql
INSERT INTO proveedores_dte (nombre, codigo, api_url, activo)
VALUES ('Proveedor DTE', 'PROV001', 'https://api.proveedor.cl', true);
```

## Mantenimiento

### Rutinas Diarias
```sql
-- Ejecutar mantenimiento automático
SELECT mantenimiento_diario();

-- Actualizar niveles de fidelización
SELECT actualizar_niveles_fidelizacion();

-- Consolidar stock
SELECT consolidar_stock_diario();
```

### Monitoreo
```sql
-- Stock crítico
SELECT * FROM reporte_stock_critico();

-- DTEs pendientes
SELECT * FROM vista_documentos_dte_pendientes;

-- Ventas del día
SELECT * FROM vista_ventas_resumen_diario WHERE fecha = CURRENT_DATE;
```

### Respaldos
```bash
# Respaldo completo
pg_dump -h localhost -U postgres -d ferre_pos_central > backup_$(date +%Y%m%d).sql

# Respaldo solo datos
pg_dump -h localhost -U postgres -d ferre_pos_central --data-only > data_backup_$(date +%Y%m%d).sql
```

## Consideraciones de Rendimiento

### Índices Críticos
- Productos: búsqueda por código de barras y descripción
- Ventas: consultas por fecha, sucursal y cajero
- Stock: consultas por producto y sucursal
- Fidelización: búsqueda por RUT y nombre

### Particionado Recomendado
Para implementaciones con alto volumen:
- `ventas` - Particionado por mes
- `movimientos_stock` - Particionado por mes
- `movimientos_fidelizacion` - Particionado por trimestre
- `logs_sincronizacion` - Particionado por semana

### Optimizaciones
- Configurar `shared_buffers` al 25% de RAM
- Ajustar `work_mem` según consultas complejas
- Habilitar `pg_stat_statements` para monitoreo
- Configurar autovacuum apropiadamente

## Seguridad

### Roles de Base de Datos
- `ferre_pos_admin` - Acceso completo
- `ferre_pos_app` - Acceso de aplicación
- `ferre_pos_readonly` - Solo lectura

### Auditoría
- Todos los cambios críticos se registran automáticamente
- Logs de seguridad para eventos sensibles
- Trazabilidad completa de operaciones

### Cumplimiento
- Protección de datos personales (Ley 21.719)
- Requisitos SII para documentos tributarios
- Auditoría de accesos y modificaciones

## Soporte y Contacto

Para consultas técnicas sobre el esquema:
- Revisar logs de PostgreSQL
- Consultar vistas de monitoreo
- Ejecutar funciones de diagnóstico

**Versión:** 1.0  
**Fecha:** Julio 2025  
**Autor:** Manus AI

