# Módulo de Notas de Crédito - Ferre-POS

## [1. Objetivo]

Permitir emitir **Notas de Crédito Electrónicas** asociadas a documentos tributarios previos (boletas o facturas), respetando normativa chilena y con **autorización obligatoria de un supervisor** antes de su emisión.

## [2. Arquitectura]

- Disponible en el servidor local de sucursal.
- Validación de identidad y rol del usuario (cajero o supervisor).
- Integración con proveedor DTE para envío electrónico.
- Registro sincronizado con servidor central.

## [3. Estructura de Base de Datos]

### Tabla: `notas_credito`

```sql
CREATE TABLE notas_credito (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  documento_origen_id UUID NOT NULL REFERENCES documentos_dte(id),
  supervisor_id UUID NOT NULL REFERENCES usuarios(id),
  cajero_id UUID NOT NULL REFERENCES usuarios(id),
  motivo TEXT NOT NULL,
  total NUMERIC(10,2) NOT NULL,
  estado TEXT CHECK (estado IN ('pendiente', 'autorizada', 'enviada', 'rechazada')) DEFAULT 'pendiente',
  fecha TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_notas_credito_documento ON notas_credito(documento_origen_id);
CREATE INDEX idx_notas_credito_estado ON notas_credito(estado);
```

## [4. Endpoints API REST]

### Solicitar nota de crédito (POST)

```http
POST /api/nota-credito/solicitar
```

**Body:**

```json
{
  "documento_origen_id": "uuid",
  "cajero_id": "uuid",
  "motivo": "Producto dañado",
  "total": 5000
}
```

### Autorizar nota de crédito (POST)

```http
POST /api/nota-credito/autorizar
```

**Body:**

```json
{
  "nota_credito_id": "uuid",
  "supervisor_id": "uuid"
}
```

### Emitir nota de crédito (POST)

```http
POST /api/nota-credito/emitir
```

**Body:**

```json
{
  "nota_credito_id": "uuid"
}
```

## [5. Flujo de Usuario]

1. Cajero solicita nota de crédito desde POS con documento de referencia.
2. Supervisor valida, aprueba y firma electrónicamente.
3. El sistema genera XML y lo envía al proveedor DTE.
4. Se registra nota como enviada y se asocia al documento original.

## [6. Seguridad]

- Emisión solo permitida tras validación de supervisor.
- Registro de identidad de cajero y supervisor.
- Firma con token o contraseña de supervisor.

## [7. Modo Offline]

- Solicitudes pueden generarse offline.
- Emisión se realiza solo cuando hay conexión y confirmación del supervisor.

## [8. UI/UX Sugerido]

- POS con botón “Solicitar Nota de Crédito” solo si hay documento válido.
- Ventana emergente para ingresar motivo.
- Interfaz de supervisor con listado de solicitudes pendientes y botón “Autorizar y Enviar”.

## [9. Beneficios]

- Control de devoluciones y correcciones contables.
- Trazabilidad y cumplimiento tributario.
- Prevención de fraudes mediante doble validación.

