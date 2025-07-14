# Módulo de Notas de Venta - Sistema Ferre-POS

## 📋 Descripción General

El módulo de Notas de Venta permite gestionar cotizaciones y reservas de productos antes de convertirlas en ventas reales. Es una funcionalidad clave para el proceso comercial de ferreterías, permitiendo a los vendedores preparar presupuestos y reservar productos para los clientes.

## 🎯 Funcionalidades Principales

### Tipos de Notas de Venta

1. **Cotizaciones**
   - Presupuestos para clientes
   - No afectan el stock disponible
   - Válidas por 30 días por defecto
   - Pueden convertirse en ventas reales

2. **Reservas**
   - Apartan productos del stock disponible
   - Válidas por 7 días por defecto
   - Reservan stock automáticamente
   - Liberan stock al vencer o anularse

### Estados de Notas

- **Activa**: Nota vigente y operativa
- **Convertida**: Transformada en venta real
- **Anulada**: Cancelada manualmente
- **Vencida**: Expirada por tiempo

## 🔧 API Endpoints

### Crear Nota de Venta

```http
POST /api/notas-venta
Authorization: Bearer <token>
Content-Type: application/json

{
  "nota": {
    "sucursal_id": "uuid-sucursal",
    "cliente_rut": "12345678-9",
    "cliente_nombre": "Juan Pérez",
    "cliente_telefono": "+56912345678",
    "cliente_email": "juan@email.com",
    "tipo_nota": "cotizacion",
    "subtotal": 10000,
    "descuento_total": 500,
    "impuesto_total": 1805,
    "total": 11305,
    "observaciones": "Cotización para proyecto de construcción"
  },
  "detalles": [
    {
      "producto_id": "uuid-producto",
      "cantidad": 2,
      "precio_unitario": 5000,
      "descuento_unitario": 250,
      "precio_final": 4750,
      "total_item": 9500,
      "observaciones": "Descuento por volumen"
    }
  ]
}
```

**Respuesta:**
```json
{
  "success": true,
  "message": "Nota de venta creada exitosamente",
  "data": {
    "nota": {
      "id": "uuid-nota",
      "numero_nota": 1001,
      "sucursal_id": "uuid-sucursal",
      "vendedor_id": "uuid-vendedor",
      "cliente_rut": "12345678-9",
      "cliente_nombre": "Juan Pérez",
      "tipo_nota": "cotizacion",
      "total": 11305,
      "estado": "activa",
      "fecha": "2024-01-15T10:30:00.000Z",
      "fecha_vencimiento": "2024-02-14T10:30:00.000Z"
    },
    "detalles": [...]
  }
}
```

### Obtener Lista de Notas

```http
GET /api/notas-venta?page=1&limit=20&tipoNota=cotizacion&estado=activa
Authorization: Bearer <token>
```

**Parámetros de consulta:**
- `page`: Número de página (default: 1)
- `limit`: Elementos por página (default: 20, max: 100)
- `sucursalId`: Filtrar por sucursal (solo admin)
- `vendedorId`: Filtrar por vendedor
- `fechaInicio`: Fecha de inicio (YYYY-MM-DD)
- `fechaFin`: Fecha de fin (YYYY-MM-DD)
- `tipoNota`: cotizacion | reserva
- `estado`: activa | convertida | anulada | vencida
- `clienteRut`: RUT del cliente

### Buscar Notas Avanzado

```http
GET /api/notas-venta/search?q=juan&montoMin=1000&montoMax=50000
Authorization: Bearer <token>
```

**Parámetros adicionales:**
- `q`: Búsqueda general (número, cliente, RUT)
- `numeroNota`: Número específico de nota
- `clienteNombre`: Nombre del cliente
- `montoMin`: Monto mínimo
- `montoMax`: Monto máximo

### Obtener Nota Específica

```http
GET /api/notas-venta/{id}
Authorization: Bearer <token>
```

**Respuesta:**
```json
{
  "success": true,
  "data": {
    "nota": {...},
    "detalles": [...],
    "reservas": [...] // Solo para reservas
  }
}
```

### Convertir a Venta Real

```http
POST /api/notas-venta/{id}/convertir
Authorization: Bearer <token>
Content-Type: application/json

{
  "terminal_id": "uuid-terminal",
  "cajero_id": "uuid-cajero",
  "tipo_documento": "boleta",
  "mediosPago": [
    {
      "medio_pago": "efectivo",
      "monto": 10000
    },
    {
      "medio_pago": "tarjeta_debito",
      "monto": 1305,
      "referencia_transaccion": "TXN123456",
      "codigo_autorizacion": "AUTH789"
    }
  ]
}
```

### Anular Nota

```http
POST /api/notas-venta/{id}/anular
Authorization: Bearer <token>
Content-Type: application/json

{
  "motivo": "Cliente canceló el pedido por cambio de presupuesto"
}
```

### Duplicar Nota

```http
POST /api/notas-venta/{id}/duplicar
Authorization: Bearer <token>
```

### Actualizar Nota

```http
PUT /api/notas-venta/{id}
Authorization: Bearer <token>
Content-Type: application/json

{
  "cliente_telefono": "+56987654321",
  "observaciones": "Observaciones actualizadas"
}
```

### Estadísticas

```http
GET /api/notas-venta/stats?fechaInicio=2024-01-01&fechaFin=2024-01-31
Authorization: Bearer <token>
```

**Respuesta:**
```json
{
  "success": true,
  "data": {
    "total_notas": 150,
    "notas_activas": 45,
    "notas_convertidas": 85,
    "notas_anuladas": 20,
    "cotizaciones": 100,
    "reservas": 50,
    "monto_total": 2500000,
    "promedio_nota": 16666.67,
    "vendedores_activos": 8,
    "clientes_unicos": 120
  }
}
```

### Notas Próximas a Vencer

```http
GET /api/notas-venta/proximas-vencer?diasAnticipacion=3
Authorization: Bearer <token>
```

### Exportar a PDF

```http
GET /api/notas-venta/{id}/pdf
Authorization: Bearer <token>
```

### Historial de Cambios

```http
GET /api/notas-venta/{id}/historial
Authorization: Bearer <token>
```

## 🔒 Permisos y Seguridad

### Roles y Accesos

| Acción | Admin | Gerente | Vendedor | Cajero |
|--------|-------|---------|----------|--------|
| Crear nota | ✅ | ✅ | ✅ | ❌ |
| Ver notas | ✅ | ✅ | ✅* | ❌ |
| Editar nota | ✅ | ✅ | ✅* | ❌ |
| Anular nota | ✅ | ✅ | ❌ | ❌ |
| Convertir a venta | ✅ | ✅ | ❌ | ✅ |
| Ver estadísticas | ✅ | ✅ | ❌ | ❌ |

*Solo sus propias notas o de su sucursal

### Validaciones de Seguridad

1. **Autenticación JWT**: Todas las rutas requieren token válido
2. **Autorización por rol**: Permisos específicos según rol de usuario
3. **Filtrado por sucursal**: Usuarios no-admin solo ven datos de su sucursal
4. **Validación de datos**: Esquemas Joi para todas las entradas
5. **Sanitización**: Limpieza automática de datos de entrada

## 📊 Lógica de Negocio

### Gestión de Stock

#### Cotizaciones
- No afectan el stock disponible
- Solo validan que hay stock suficiente al momento de creación
- Al convertir a venta, se valida stock nuevamente

#### Reservas
- Reducen la cantidad disponible automáticamente
- Crean registros en `reservas_stock`
- Liberan stock al anular o vencer
- Al convertir a venta, transfieren la reserva a venta real

### Cálculo de Totales

```javascript
// Validación automática de totales
totalDetalles = sum(detalle.total_item for detalle in detalles)
totalCalculado = totalDetalles - descuento_total + impuesto_total

// Debe cumplir: |totalCalculado - total| <= 0.01
```

### Fechas de Vencimiento

- **Cotizaciones**: 30 días desde creación
- **Reservas**: 7 días desde creación
- **Personalizable**: Se puede modificar en configuración

### Numeración Automática

- Numeración secuencial por sucursal
- Formato: `{año}{mes}{sucursal_codigo}{numero}`
- Ejemplo: `202401PRIN001` (Enero 2024, Sucursal Principal, Nota 001)

## 🔄 Flujos de Trabajo

### Flujo de Cotización

1. **Creación**: Vendedor crea cotización con productos y precios
2. **Envío**: Se puede exportar a PDF para enviar al cliente
3. **Seguimiento**: Monitoreo de estado y vencimiento
4. **Conversión**: Cliente acepta y se convierte en venta
5. **Cierre**: Nota queda marcada como convertida

### Flujo de Reserva

1. **Creación**: Vendedor crea reserva apartando productos
2. **Reserva de Stock**: Sistema reduce stock disponible automáticamente
3. **Notificación**: Alertas de vencimiento próximo
4. **Conversión o Liberación**: 
   - Cliente compra → Conversión a venta
   - Cliente no compra → Liberación automática de stock

### Flujo de Conversión a Venta

1. **Validación**: Verificar stock disponible actual
2. **Creación de Venta**: Generar registro de venta completo
3. **Actualización de Stock**: Reducir inventario físico
4. **Medios de Pago**: Registrar formas de pago utilizadas
5. **Documentos**: Generar boleta/factura según corresponda
6. **Cierre**: Marcar nota como convertida

## 📈 Reportes y Análisis

### Métricas Disponibles

1. **Conversión**: % de notas convertidas vs creadas
2. **Tiempo promedio**: Días entre creación y conversión
3. **Valor promedio**: Monto promedio por nota
4. **Productos más cotizados**: Top productos en notas
5. **Vendedores más efectivos**: Mejor ratio de conversión
6. **Tendencias temporales**: Evolución mensual/semanal

### Alertas Automáticas

1. **Notas próximas a vencer**: 3 días antes por defecto
2. **Stock insuficiente**: Al intentar crear reserva sin stock
3. **Notas vencidas**: Proceso automático de liberación
4. **Conversiones pendientes**: Recordatorios de seguimiento

## 🛠️ Configuración

### Variables de Entorno

```env
# Días de validez por defecto
COTIZACION_DIAS_VALIDEZ=30
RESERVA_DIAS_VALIDEZ=7

# Alertas de vencimiento
ALERTA_VENCIMIENTO_DIAS=3

# Límites de consulta
MAX_NOTAS_PER_PAGE=100
DEFAULT_NOTAS_PER_PAGE=20
```

### Configuración de Base de Datos

Las notas de venta utilizan las siguientes tablas:

- `notas_venta`: Registro principal
- `detalle_notas_venta`: Productos incluidos
- `reservas_stock`: Control de reservas (solo para tipo reserva)

## 🧪 Testing

### Ejecutar Tests

```bash
# Tests específicos del módulo
npm test tests/notaVenta.test.js

# Tests de integración
npm run test:integration

# Coverage del módulo
npm run test:coverage -- tests/notaVenta.test.js
```

### Casos de Prueba Incluidos

1. **Creación exitosa** de cotizaciones y reservas
2. **Validación de stock** insuficiente
3. **Validación de totales** incorrectos
4. **Conversión a venta** completa
5. **Anulación** con liberación de stock
6. **Búsquedas y filtros** avanzados
7. **Permisos por rol** de usuario
8. **Estadísticas** y reportes

## 🔧 Mantenimiento

### Tareas Programadas Recomendadas

1. **Limpieza de notas vencidas**: Diario a las 02:00
2. **Alertas de vencimiento**: Diario a las 09:00
3. **Liberación de reservas vencidas**: Cada hora
4. **Backup de notas**: Semanal

### Monitoreo

- **Logs de negocio**: Todas las operaciones importantes
- **Métricas de rendimiento**: Tiempo de respuesta de APIs
- **Alertas de error**: Fallos en conversiones o validaciones
- **Uso de recursos**: Consultas pesadas o lentas

## 📞 Soporte

Para soporte específico del módulo de notas de venta:

- **Documentación técnica**: `/docs/api` en el servidor
- **Logs del sistema**: `logs/ferre-pos-api.log`
- **Tests de validación**: `npm test tests/notaVenta.test.js`
- **Contacto**: soporte@ferre-pos.cl

---

**Módulo desarrollado para optimizar el proceso comercial de ferreterías** 🔧⚡

