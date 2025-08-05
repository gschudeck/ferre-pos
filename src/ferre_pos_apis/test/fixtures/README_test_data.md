# Datos de Prueba - Sistema FERRE-POS

## 📋 Descripción

Este directorio contiene los scripts SQL y archivos de configuración para insertar y gestionar datos de prueba en el sistema FERRE-POS.

## 📁 Archivos Incluidos

### Scripts SQL
- **`insert_test_data.sql`**: Script principal para insertar todos los datos de prueba
- **`cleanup_test_data.sql`**: Script para limpiar/eliminar todos los datos de prueba
- **`test_data.json`**: Datos de prueba en formato JSON para los tests

## 🗃️ Datos de Prueba Incluidos

### 👥 Usuarios de Prueba
| Username | Password | Rol | Email | Sucursal |
|----------|----------|-----|-------|----------|
| `admin_test` | `password123` | admin | admin.test@ferrepos.com | Test Sucursal Principal |
| `vendedor_test` | `password123` | vendedor | vendedor.test@ferrepos.com | Test Sucursal Principal |
| `cajero_test` | `password123` | cajero | cajero.test@ferrepos.com | Test Sucursal Principal |
| `supervisor_test` | `password123` | supervisor | supervisor.test@ferrepos.com | Test Sucursal Secundaria |

### 🏢 Sucursales de Prueba
- **Test Sucursal Principal**: Av. Test 123, Santiago
- **Test Sucursal Secundaria**: Calle Test 456, Valparaíso

### 💻 Terminales de Prueba
- **TERM001**: Terminal Test Principal (Sucursal 1)
- **TERM002**: Terminal Test Secundaria (Sucursal 1)
- **TERM003**: Terminal Test Sucursal 2 (Sucursal 2)

### 📦 Categorías de Productos
- **Test Herramientas**: Herramientas de construcción
- **Test Materiales**: Materiales de construcción
- **Test Ferretería**: Artículos de ferretería general
- **Test Electricidad**: Artículos eléctricos

### 🛠️ Productos de Prueba
| Código | Nombre | Categoría | Precio | Stock |
|--------|--------|-----------|--------|-------|
| TEST001 | Martillo Test 500g | Herramientas | $15.000 | 25 |
| TEST002 | Destornillador Test Phillips | Herramientas | $3.500 | 50 |
| TEST003 | Tornillo Test 6x40mm | Ferretería | $150 | 500 |
| TEST004 | Cable Test 2.5mm | Electricidad | $2.500 | 100 |
| TEST005 | Cemento Test 25kg | Materiales | $8.500 | 15 |
| TEST006 | Taladro Test 600W | Herramientas | $45.000 | 8 |
| TEST007 | Pintura Test Blanca 1L | Materiales | $12.000 | 30 |
| TEST008 | Interruptor Test Simple | Electricidad | $2.800 | 40 |

### 🧾 Ventas de Prueba
- **TEST-V-001**: Venta con martillo, destornilladores y tornillos ($29.750)
- **TEST-V-002**: Venta con cables e interruptores con descuento ($20.825)
- **TEST-V-003**: Factura con taladro ($51.170)
- **TEST-V-004**: Venta en sucursal 2 con pintura ($14.280)

### 🏷️ Plantillas de Etiquetas
- **Plantilla Básica Test**: 50x30mm con código, nombre, precio y código de barras
- **Plantilla Avanzada Test**: 70x40mm con elementos adicionales

## 🚀 Uso de los Scripts

### Insertar Datos de Prueba
```sql
-- Conectar a la base de datos
psql -h localhost -U usuario -d ferre_pos_central

-- Ejecutar script de inserción
\i test/fixtures/insert_test_data.sql
```

### Limpiar Datos de Prueba
```sql
-- Ejecutar script de limpieza
\i test/fixtures/cleanup_test_data.sql
```

### Desde Línea de Comandos
```bash
# Insertar datos
psql -h localhost -U usuario -d ferre_pos_central -f test/fixtures/insert_test_data.sql

# Limpiar datos
psql -h localhost -U usuario -d ferre_pos_central -f test/fixtures/cleanup_test_data.sql
```

## 🔧 Configuración para Tests

### Variables de Entorno
```bash
export TEST_DB_HOST=localhost
export TEST_DB_PORT=5432
export TEST_DB_USER=test_user
export TEST_DB_PASSWORD=test_password
export TEST_DB_NAME=ferre_pos_test
```

### Uso en Tests Go
```go
// Cargar fixtures en tests
func setupTestData(t *testing.T) {
    var testData map[string]interface{}
    testutils.ParseJSONFixture(t, "test_data.json", &testData)
    // Usar datos en tests...
}
```

## 📊 Escenarios de Prueba Cubiertos

### ✅ Casos de Éxito
- Usuarios con diferentes roles y permisos
- Productos con stock suficiente
- Ventas completas con múltiples items
- Movimientos de stock normales
- Plantillas de etiquetas funcionales

### ⚠️ Casos Límite
- Productos con stock bajo (Cemento: 15 unidades, stock mínimo: 3)
- Productos con stock reservado
- Ventas con descuentos
- Diferentes métodos de pago

### ❌ Casos de Error
- Productos inactivos (para tests de validación)
- Usuarios inactivos (para tests de autenticación)
- Stock insuficiente (para tests de validación de ventas)

## 🔒 Seguridad

### Contraseñas
- Todas las contraseñas de prueba están hasheadas con bcrypt
- Password por defecto: `password123`
- Hash: `$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi`

### Datos Sensibles
- Los datos de prueba NO contienen información real
- Todos los emails usan el dominio `@ferrepos.com`
- Los números de teléfono son ficticios
- Las direcciones son de prueba

## 🧪 Integración con Tests

### Tests Unitarios
Los datos están diseñados para soportar:
- Tests de autenticación con diferentes roles
- Tests de CRUD de productos
- Tests de validación de ventas
- Tests de generación de reportes

### Tests de Integración
- Flujos completos de venta
- Sincronización entre terminales
- Generación de etiquetas
- Reportes con datos reales

### Tests E2E
- Escenarios de usuario completos
- Flujos de negocio end-to-end
- Tests de rendimiento con datos

## 📝 Mantenimiento

### Actualizar Datos
1. Modificar `insert_test_data.sql`
2. Ejecutar script de limpieza
3. Ejecutar script de inserción actualizado
4. Verificar que los tests siguen funcionando

### Agregar Nuevos Datos
1. Seguir el patrón de nomenclatura (`TEST-*`, `test-*`)
2. Mantener consistencia en las relaciones
3. Actualizar documentación
4. Agregar casos de prueba correspondientes

## 🐛 Troubleshooting

### Error de Foreign Key
- Verificar que las dependencias se insertan en orden correcto
- Revisar que los IDs referenciados existen

### Error de Duplicados
- Ejecutar script de limpieza antes de insertar
- Verificar que no hay datos previos con los mismos IDs

### Tests Fallando
- Verificar que los datos están insertados correctamente
- Revisar que los IDs en los tests coinciden con los datos
- Confirmar que la base de datos de test está limpia

## 📞 Soporte

Para problemas con los datos de prueba:
1. Verificar logs de la base de datos
2. Ejecutar consultas de verificación incluidas en los scripts
3. Revisar la documentación de tests
4. Contactar al equipo de desarrollo

