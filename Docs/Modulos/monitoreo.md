# Módulo de Monitoreo - Ferre-POS

## [1. Objetivo]

Supervisar en tiempo real la actividad, sincronización y estado de todos los componentes del sistema Ferre-POS: POS, servidor central, stock y conexiones con proveedor DTE y ERP.

## [2. Arquitectura General]

```
POS / Tienda / Tótem
     ⇅ (ping + logs)        → Servidor Central
                             ⇅ (export Prometheus)
Grafana / Prometheus / Metabase
     ← consulta métricas y alertas
```

## [3. Registro de Logs de Sincronización]

### Tabla: logs\_sincronizacion

```sql
CREATE TABLE logs_sincronizacion (
  id SERIAL PRIMARY KEY,
  terminal_id UUID REFERENCES terminales(id),
  sucursal_id UUID REFERENCES sucursales(id),
  tipo TEXT CHECK (tipo IN ('venta', 'stock', 'ping', 'error')),
  resultado TEXT CHECK (resultado IN ('ok', 'fallo', 'parcial')),
  detalles TEXT,
  intentos INTEGER DEFAULT 0,
  fecha TIMESTAMP DEFAULT NOW(),
  ip_origen TEXT
);
```

## [4. Ping de Estado del POS]

### Endpoint: `/api/ping`

**Request:**

```json
{
  "terminal_id": "uuid-terminal",
  "sucursal_id": "uuid-sucursal",
  "version": "1.2.7",
  "timestamp_ultima_venta": "2025-07-10T13:22:00",
  "impresora_estado": "ok",
  "dte_estado": "conectado",
  "modo": "online"
}
```

**Response:**

```json
{
  "status": "ok",
  "reintentar_sync": false
}
```

## [5. Métricas para Prometheus]

Ejemplo de exporter en Node.js:

```js
const ventasNoSync = new client.Gauge({ name: 'ferrepos_ventas_no_sync_total', help: 'Ventas sin sincronizar', labelNames: ['terminal'] });
const reintentosSync = new client.Gauge({ name: 'ferrepos_reintentos_sync', help: 'Reintentos de sincronización', labelNames: ['terminal'] });
```

## [6. Panel POS Activos - SQL]

```sql
SELECT
  t.nombre_terminal,
  s.nombre AS sucursal,
  MAX(l.fecha) AS ultima_actividad,
  CASE
    WHEN MAX(l.fecha) > NOW() - INTERVAL '5 minutes' THEN 'activo'
    ELSE 'inactivo'
  END AS estado
FROM logs_sincronizacion l
JOIN terminales t ON l.terminal_id = t.id
JOIN sucursales s ON t.sucursal_id = s.id
WHERE l.tipo = 'ping'
GROUP BY t.nombre_terminal, s.nombre
ORDER BY ultima_actividad DESC;
```

## [7. Stock Crítico - SQL para Metabase]

```sql
SELECT
  p.codigo,
  p.descripcion,
  s.nombre AS sucursal,
  stk.cantidad
FROM stock stk
JOIN productos p ON stk.producto_id = p.id
JOIN sucursales s ON stk.sucursal_id = s.id
WHERE stk.cantidad < 5
ORDER BY cantidad ASC;
```

## [8. Alerta POS Inactivo - Prometheus Rule]

```yaml
- alert: POSInactivo
  expr: time() - ferrepos_ping_timestamp > 600
  for: 5m
  labels:
    severity: critical
  annotations:
    summary: "POS sin actividad detectado"
    description: "El POS {{ $labels.terminal }} no ha enviado ping en los últimos 10 minutos."
```

## [9. Herramientas Sugeridas]

- **Prometheus**: recolección de métricas
- **Grafana**: visualización
- **Metabase**: dashboards SQL ejecutivos
- **pgAgent**: tareas programadas
- **Alertmanager**: notificaciones

## [10. Recomendaciones]

- Configurar Prometheus para scrapear `/metrics` de exportadores en cada servidor.
- Registrar timestamp de última venta en la tabla `terminales` para detectar inactividad prolongada.
- Incluir integridad en pings y logs para evitar falsos positivos.

