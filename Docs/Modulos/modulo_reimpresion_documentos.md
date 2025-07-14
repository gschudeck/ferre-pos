# Módulo de Reimpresión de Documentos - Ferre-POS

## [1. Objetivo]

Permitir la reimpresión de boletas, facturas, guías de despacho y notas de crédito desde cualquier POS o el servidor central, respetando los permisos de usuario y manteniendo registro de auditoría.

## [2. Alcance]

- Disponible en POS caja, POS tienda y servidor central.
- Se permite reimpresión si el documento fue previamente emitido electrónicamente.
- Control de accesos: sólo usuarios con rol `supervisor` pueden reimprimir documentos tributarios (boletas, facturas, NC).

## [3. Estructura de Base de Datos]

### Tabla: `reimpresiones_documentos`

```sql
CREATE TABLE reimpresiones_documentos (
  id SERIAL PRIMARY KEY,
  documento_id UUID REFERENCES documentos_dte(id),
  usuario_id UUID REFERENCES usuarios(id),
  sucursal_id UUID REFERENCES sucursales(id),
  motivo TEXT,
  fecha TIMESTAMP DEFAULT NOW(),
  ip_origen TEXT,
  dispositivo TEXT,
  reimpresiones_previas INTEGER DEFAULT 0
);

CREATE INDEX idx_reimpresion_fecha ON reimpresiones_documentos(fecha);
CREATE INDEX idx_reimpresion_documento ON reimpresiones_documentos(documento_id);
```

## [4. API REST]

### Reimprimir documento (POST)

```http
POST /api/reimpresion/documento
```

**Body:**

```json
{
  "documento_id": "uuid",
  "usuario_id": "uuid",
  "motivo": "Cliente lo perdió",
  "ip_origen": "192.168.1.20",
  "dispositivo": "POS-Caja01"
}
```

**Respuesta:** archivo PDF o señal al driver de impresión local.

## [5. Flujo de Usuario]

1. Usuario solicita documento indicando folio o ID.
2. Sistema valida que sea reimprimible.
3. Se verifica que el usuario tenga permiso.
4. Se registra la reimpresión en la tabla de auditoría.
5. Se genera la reimpresión (PDF o impresión directa).

## [6. Seguridad]

- Reimpresión sólo disponible para documentos finalizados y emitidos.
- Validación estricta de rol y registro completo en log.
- Restricción de número de reimpresiones por día configurable (opcional).

## [7. Modo Offline]

- Permite reimpresión si el documento está en caché local.
- Si no está disponible, el POS debe esperar sincronización.

## [8. UI/UX Sugerido]

- Pantalla para buscar por folio, RUT o fecha.
- Botón "Reimprimir" con confirmación de motivo.
- Log de reimpresiones visibles por supervisor.

## [9. Beneficios]

- Mejora la atención al cliente.
- Garantiza trazabilidad de reimpresiones.
- Cumplimiento normativo y prevención de fraudes.

