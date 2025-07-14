# Módulo POS Tienda (Sala de Ventas) - Ferre-POS

## [1. Objetivo]

Ofrecer una interfaz gráfica ágil y de fácil uso para vendedores en sala, que permita emitir notas de venta internas, aplicar descuentos y gestionar el flujo comercial previo al pago en caja.

## [2. Alcance]

- Usado por vendedores en sala.
- Opera con interfaz gráfica compatible con mouse, teclado y lector de código de barras.
- Enlazado con servidor central y compatible con modo offline.
- Genera notas de venta internas que deben ser pagadas en caja.
- Integrado con el módulo de fidelización para acumulación futura.

## [3. Estructura de Base de Datos]

### Índices recomendados para búsqueda de productos

```sql
-- Índice tradicional para búsqueda exacta por código de barras
CREATE INDEX idx_productos_codigo_barra ON productos(codigo_barra);

-- Habilitar extensión para búsquedas por similitud
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Índice GIN para búsqueda por descripción aproximada
CREATE INDEX idx_productos_descripcion_trgm ON productos USING GIN (descripcion gin_trgm_ops);

-- Índice opcional para ordenamiento por precio
CREATE INDEX idx_productos_precio ON productos(precio_unitario);
```

### Tabla: `notas_venta`

```sql
CREATE TABLE notas_venta (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  vendedor_id UUID REFERENCES usuarios(id),
  sucursal_id UUID REFERENCES sucursales(id),
  cliente_rut TEXT,
  total NUMERIC(10,2),
  estado TEXT CHECK (estado IN ('pendiente', 'pagada', 'cancelada')) DEFAULT 'pendiente',
  fecha TIMESTAMP DEFAULT NOW(),
  sincronizada BOOLEAN DEFAULT false
);
```

### Tabla: `detalle_notas_venta`

```sql
CREATE TABLE detalle_notas_venta (
  id SERIAL PRIMARY KEY,
  nota_venta_id UUID REFERENCES notas_venta(id),
  producto_id UUID REFERENCES productos(id),
  cantidad INTEGER,
  precio_unitario NUMERIC(10,2),
  total_item NUMERIC(10,2)
);
```

## [4. API REST]

### Crear nota de venta (POST)

```http
POST /api/pos-tienda/nota-venta
```

**Body:**

```json
{
  "vendedor_id": "uuid",
  "cliente_rut": "11111111-1",
  "productos": [
    {"producto_id": "uuid", "cantidad": 2, "precio_unitario": 3500}
  ]
}
```

### Consultar nota de venta por ID (GET)

```http
GET /api/pos-tienda/nota-venta/{id}
```

### Cambiar estado a pagada (PATCH)

```http
PATCH /api/pos-tienda/nota-venta/{id}/pagar
```

## [5. Flujo de Usuario]

1. El vendedor escanea o selecciona productos.
2. El sistema calcula totales y muestra opciones de descuento.
3. Se genera una nota de venta interna.
4. El cliente se dirige a caja a pagar.
5. La caja marca la nota como “pagada” al momento del cobro.
6. Si cliente está registrado, se envía acumulación de puntos a módulo de fidelización.

## [6. Seguridad]

- Solo usuarios con rol `vendedor` pueden crear notas.
- El cambio de estado requiere verificación del POS caja.
- Las notas pagadas no pueden modificarse.

## [7. Modo Offline]

- El POS Tienda puede operar en modo desconectado.
- Las notas se almacenan localmente con `sincronizada = false`.
- Al reconectarse, el sistema sincroniza con el servidor central o caja para registro y acumulación.

## [8. UI/UX Sugerido]

- Pantalla con buscador rápido de productos.
- Campo de búsqueda por código o nombre con sugerencias.
- Soporte para búsquedas aproximadas (Levenshtein) y fonéticas (Soundex o similitud).
- Lista dinámica de ítems agregados con totales.
- Botón “Generar Nota de Venta”.
- Impresión inmediata del comprobante para el cliente.
- Indicador visual de estado de sincronización.

## [9. Sincronización con Caja]

- La nota de venta se transmite al POS caja cuando se inicia el cobro.
- Si se encuentra `sincronizada = false`, se prioriza su envío inmediato.
- Confirmado el pago, la caja envía PATCH a `/nota-venta/{id}/pagar`.
- El POS tienda puede consultar estado con `GET /nota-venta/{id}`.
- Si se trata de cliente registrado, se dispara POST a `/api/fidelizacion/acumular`.

## [10. Beneficios]

- Agiliza la operación comercial en piso.
- Separa venta de cobro, optimizando atención.
- Mejora el control de stock al registrar intención de compra.
- Compatible con fidelización y promociones.
- Soporta operación distribuida entre caja y vendedores.

