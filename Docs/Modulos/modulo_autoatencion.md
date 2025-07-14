# Módulo de Autoatención (Tótem) - Ferre-POS

## [1. Objetivo]

Proveer un canal rápido, autónomo y versátil de atención al cliente en ferreterías urbanas. El módulo de autoatención permite:

- Generar notas de venta para pagar en caja.
- Realizar pagos directamente desde el tótem (si habilitado).
- Consultar precios y stock multisucursal.

## [2. Tipos de Operación]

1. **Consulta**: permite al cliente escanear o buscar productos, ver stock y precios sin registrar compra.
2. **Preventa / Nota de Venta**: permite seleccionar productos y generar una nota para pagar en caja.
3. **Venta Directa + Pago**: en tótems habilitados, permite pagar con medios electrónicos (Webpay, MercadoPago).

## [3. Identificación de Usuario]

- El cliente puede usar RUT, escanear su QR de fidelización o continuar como anónimo.
- Si se identifica, se asocian los puntos y promociones vigentes.

## [4. Interfaz UI/UX Sugerida]

- Pantalla táctil intuitiva.
- Buscador de productos con texto predictivo y escáner de código.
- Categorías visuales.
- Carrito con totalizador dinámico.
- Botón “Generar nota de venta” o “Pagar ahora”.
- Impresión de comprobante si aplica.

## [5. Base de Datos - Esquema Relacionado]

Utiliza las mismas tablas `notas_venta` y `detalle_notas_venta` que POS Tienda:

- Relaciona con cliente si está identificado.
- Incluye campo `origen = 'autoatencion'` en tabla `notas_venta` para trazabilidad.

```sql
ALTER TABLE notas_venta ADD COLUMN origen TEXT DEFAULT 'tienda';
```

## [6. API REST]

### Buscar producto

```http
GET /api/productos?query=martillo
```

### Crear nota de venta (POST)

```http
POST /api/autoatencion/nota-venta
```

### Pagar nota de venta (si aplica)

```http
POST /api/autoatencion/nota-venta/{id}/pagar
```

## [7. Medios de Pago]

- Si habilitado, permite usar:
  - Webpay (API integrada o QR).
  - MercadoPago.
  - Otros gateways configurables.
- Requiere integración segura con backend y proveedor de pago.

## [8. Modo Offline]

- Solo habilitado para generación de notas de venta.
- Las transacciones se almacenan localmente y sincronizan al reconectarse.

## [9. Seguridad]

- Modo kiosk: sesión anónima con expiración.
- Validación del estado de pago si aplica.
- Límites por ticket y tiempo de sesión.

## [10. Beneficios]

- Reduce filas y agiliza atención en horas punta.
- Mejora experiencia del cliente.
- Extiende cobertura de fidelización.
- Funciona como canal de autoservicio sin intervención del personal.

