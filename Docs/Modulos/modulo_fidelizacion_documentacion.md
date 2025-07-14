# Documentación Técnica - Módulo Fidelización de Clientes - Ferre-POS

## [1. Descripción General]
El módulo de fidelización permite a los clientes acumular puntos por cada compra y luego canjearlos por descuentos o beneficios en futuras compras. Es multisucursal y permite trazabilidad completa de los movimientos.

## [2. Componentes Principales]

### Tablas

#### `fidelizacion_clientes`
- `id`: UUID (clave primaria)
- `rut`: RUT único del cliente
- `nombre`: Nombre del cliente
- `puntos`: Puntos actuales acumulados
- `fecha_ultima_compra`: Última vez que sumó puntos
- `fecha_creacion`: Fecha de registro

#### `movimientos_fidelizacion`
- `id`: UUID
- `cliente_id`: Relación con cliente
- `sucursal_id`: Sucursal donde ocurrió el movimiento
- `tipo`: acumulación o canje
- `puntos`: cantidad de puntos involucrados
- `detalle`: descripción del motivo
- `fecha`: timestamp del movimiento

### Triggers

#### `validar_canje_puntos`
- Impide que se canjeen más puntos de los disponibles.

#### `actualizar_saldo_fidelizacion`
- Aumenta o disminuye puntos automáticamente según el tipo de movimiento.
- Actualiza fecha de última compra.

## [3. API REST]

### `GET /api/fidelizacion/cliente/{rut}`
- Devuelve puntos actuales y nombre del cliente.

### `POST /api/fidelizacion/acumular`
- Agrega puntos al cliente por una venta específica.

### `POST /api/fidelizacion/canjear`
- Descuenta puntos y registra un movimiento de canje.

## [4. Reglas de Negocio]
- Solo se permite canjear si hay puntos suficientes.
- Cada canje o acumulación queda registrado con detalle.
- Mismo cliente puede operar en distintas sucursales.

## [5. Seguridad]
- Solo cajeros y roles con permisos pueden acumular o canjear.
- El sistema valida la existencia del RUT antes de operar.

## [6. Operación Offline]
- Se registran movimientos localmente.
- Al reconectar, se sincronizan con el servidor central.

## [7. UI/UX]
- Visualización del saldo al ingresar RUT.
- Botón “Canjear puntos” visible si hay saldo.
- Impresión o visualización en pantalla del historial al cliente.

## [8. Informes y Trazabilidad]
- Reportes por cliente, sucursal, tipo de operación.
- Control de acumulación/canje por período.

## [9. Ejemplo de Consulta Historial Cliente]
```sql
SELECT m.fecha, m.tipo, m.puntos, m.detalle, s.nombre AS sucursal
FROM movimientos_fidelizacion m
JOIN sucursales s ON s.id = m.sucursal_id
JOIN fidelizacion_clientes f ON f.id = m.cliente_id
WHERE f.rut = '11111111-1'
ORDER BY m.fecha DESC
LIMIT 20;
```

## [10. Beneficios]
- Fomenta repetición de compra.
- Permite campañas de marketing personalizadas.
- Brinda datos sobre comportamiento de clientes.
- Incrementa satisfacción y lealtad del cliente.

