# Módulo de Control de Despacho - Ferre-POS

## [1. Objetivo]

Permitir validar en bodega que los productos vendidos o documentados (boletas, facturas, guías) coincidan con lo efectivamente entregado al cliente. Mejora trazabilidad, reduce errores logísticos y registra diferencias de entrega.

## [2. Arquitectura]

- Módulo se ejecuta en equipo local de bodega.
- Conexión al servidor local de sucursal (consulta de documentos, ventas y stock).
- Sincronización eventual con servidor central para auditoría.

## [3. Estructura de Base de Datos]

### Tabla: `despachos`

```sql
CREATE TABLE despachos (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  documento_id UUID REFERENCES documentos_dte(id),
  usuario_id UUID REFERENCES usuarios(id),
  fecha TIMESTAMP DEFAULT NOW(),
  estado TEXT CHECK (estado IN ('completo', 'parcial', 'rechazado')) DEFAULT 'completo',
  observacion TEXT
);
```

### Tabla: `detalle_despacho`

```sql
CREATE TABLE detalle_despacho (
  id SERIAL PRIMARY KEY,
  despacho_id UUID REFERENCES despachos(id),
  producto_id UUID REFERENCES productos(id),
  cantidad_vendida INTEGER,
  cantidad_entregada INTEGER
);
```

## [4. Endpoints API REST]

### Iniciar despacho (GET)

```http
GET /api/despacho/iniciar?documento_id=uuid
```

**Respuesta:**

```json
{
  "documento_id": "uuid",
  "productos": [
    { "producto_id": "uuid", "descripcion": "Martillo", "cantidad_vendida": 2 }
  ]
}
```

### Registrar despacho (POST)

```http
POST /api/despacho/confirmar
```

**Body:**

```json
{
  "documento_id": "uuid",
  "usuario_id": "uuid",
  "estado": "completo",
  "productos": [
    { "producto_id": "uuid", "cantidad_entregada": 2 }
  ]
}
```

## [5. Flujo de Usuario]

1. Cliente se presenta en bodega con boleta o guía.
2. Operario busca o escanea documento.
3. Sistema muestra productos vendidos.
4. Se validan y confirman cantidades.
5. Se registra despacho y se actualiza estado.

## [6. Reporte de Entregas Parciales]

```sql
SELECT d.fecha, s.nombre AS sucursal, u.nombre AS operador, p.descripcion AS producto,
       dd.cantidad_vendida, dd.cantidad_entregada,
       (dd.cantidad_vendida - dd.cantidad_entregada) AS diferencia
FROM despachos d
JOIN detalle_despacho dd ON d.id = dd.despacho_id
JOIN productos p ON dd.producto_id = p.id
JOIN usuarios u ON d.usuario_id = u.id
JOIN sucursales s ON u.sucursal_id = s.id
WHERE d.estado = 'parcial'
ORDER BY d.fecha DESC
LIMIT 20;
```

## [7. Seguridad]

- Solo usuarios con rol `despacho` o `supervisor` pueden registrar o consultar despachos.
- Registro completo de usuario y fecha de cada despacho.

## [8. Modo Offline]

- Opera con datos locales.
- Despachos se almacenan localmente y se sincronizan con el servidor central al reconectarse.

## [9. UI/UX Sugerido]

- Pantalla con productos y cantidad esperada.
- Validación rápida por lector o teclado.
- Botón de confirmación total o por línea.
- Estado visual por producto (✓ o ✗).

## [10. Beneficios]

- Prevención de errores de entrega.
- Registro confiable de trazabilidad.
- Control y estadísticas por usuario, producto y sucursal.
- Mejora la experiencia de cliente final.

