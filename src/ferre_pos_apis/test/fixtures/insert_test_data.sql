-- =====================================================
-- SCRIPT DE INSERCIÓN DE DATOS DE PRUEBA
-- Sistema FERRE-POS - Servidor Central
-- =====================================================

-- Limpiar datos existentes de test (opcional)
-- DELETE FROM detalle_ventas WHERE venta_id IN (SELECT id FROM ventas WHERE numero_documento LIKE 'TEST%');
-- DELETE FROM ventas WHERE numero_documento LIKE 'TEST%';
-- DELETE FROM movimientos_stock WHERE producto_id IN (SELECT id FROM productos WHERE codigo LIKE 'TEST%');
-- DELETE FROM stock_central WHERE producto_id IN (SELECT id FROM productos WHERE codigo LIKE 'TEST%');
-- DELETE FROM productos WHERE codigo LIKE 'TEST%';
-- DELETE FROM categorias_productos WHERE nombre LIKE 'Test%';
-- DELETE FROM usuarios WHERE username LIKE 'test%';
-- DELETE FROM terminales WHERE codigo LIKE 'TEST%';
-- DELETE FROM sucursales WHERE nombre LIKE 'Test%';

-- =====================================================
-- 1. SUCURSALES DE PRUEBA
-- =====================================================

INSERT INTO sucursales (
    id, nombre, direccion, telefono, email, activa, 
    created_at, updated_at
) VALUES 
(
    'test-sucursal-1'::uuid,
    'Test Sucursal Principal',
    'Av. Test 123, Santiago',
    '+56912345678',
    'sucursal.test@ferrepos.com',
    true,
    NOW(),
    NOW()
),
(
    'test-sucursal-2'::uuid,
    'Test Sucursal Secundaria',
    'Calle Test 456, Valparaíso',
    '+56987654321',
    'sucursal2.test@ferrepos.com',
    true,
    NOW(),
    NOW()
);

-- =====================================================
-- 2. USUARIOS DE PRUEBA
-- =====================================================

INSERT INTO usuarios (
    id, username, email, password_hash, nombre_completo, 
    rol, activo, sucursal_id, ultimo_acceso,
    created_at, updated_at
) VALUES 
(
    'test-user-1'::uuid,
    'admin_test',
    'admin.test@ferrepos.com',
    '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', -- password: password123
    'Administrador Test',
    'admin',
    true,
    'test-sucursal-1'::uuid,
    NOW(),
    NOW(),
    NOW()
),
(
    'test-user-2'::uuid,
    'vendedor_test',
    'vendedor.test@ferrepos.com',
    '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', -- password: password123
    'Vendedor Test',
    'vendedor',
    true,
    'test-sucursal-1'::uuid,
    NOW(),
    NOW(),
    NOW()
),
(
    'test-user-3'::uuid,
    'cajero_test',
    'cajero.test@ferrepos.com',
    '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', -- password: password123
    'Cajero Test',
    'cajero',
    true,
    'test-sucursal-1'::uuid,
    NOW(),
    NOW(),
    NOW()
),
(
    'test-user-4'::uuid,
    'supervisor_test',
    'supervisor.test@ferrepos.com',
    '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', -- password: password123
    'Supervisor Test',
    'supervisor',
    true,
    'test-sucursal-2'::uuid,
    NOW(),
    NOW(),
    NOW()
);

-- =====================================================
-- 3. TERMINALES DE PRUEBA
-- =====================================================

INSERT INTO terminales (
    id, codigo, nombre, sucursal_id, ip_address, 
    mac_address, activo, ultimo_heartbeat,
    created_at, updated_at
) VALUES 
(
    'test-terminal-1'::uuid,
    'TERM001',
    'Terminal Test Principal',
    'test-sucursal-1'::uuid,
    '192.168.1.100',
    '00:11:22:33:44:55',
    true,
    NOW(),
    NOW(),
    NOW()
),
(
    'test-terminal-2'::uuid,
    'TERM002',
    'Terminal Test Secundaria',
    'test-sucursal-1'::uuid,
    '192.168.1.101',
    '00:11:22:33:44:56',
    true,
    NOW(),
    NOW(),
    NOW()
),
(
    'test-terminal-3'::uuid,
    'TERM003',
    'Terminal Test Sucursal 2',
    'test-sucursal-2'::uuid,
    '192.168.2.100',
    '00:11:22:33:44:57',
    true,
    NOW(),
    NOW(),
    NOW()
);

-- =====================================================
-- 4. CATEGORÍAS DE PRODUCTOS DE PRUEBA
-- =====================================================

INSERT INTO categorias_productos (
    id, nombre, descripcion, activa,
    created_at, updated_at
) VALUES 
(
    'test-category-1'::uuid,
    'Test Herramientas',
    'Categoría de prueba para herramientas',
    true,
    NOW(),
    NOW()
),
(
    'test-category-2'::uuid,
    'Test Materiales',
    'Categoría de prueba para materiales de construcción',
    true,
    NOW(),
    NOW()
),
(
    'test-category-3'::uuid,
    'Test Ferretería',
    'Categoría de prueba para artículos de ferretería',
    true,
    NOW(),
    NOW()
),
(
    'test-category-4'::uuid,
    'Test Electricidad',
    'Categoría de prueba para artículos eléctricos',
    true,
    NOW(),
    NOW()
);

-- =====================================================
-- 5. PRODUCTOS DE PRUEBA
-- =====================================================

INSERT INTO productos (
    id, codigo, codigo_barras, nombre, descripcion,
    categoria_id, precio, costo, stock_minimo, activo,
    created_at, updated_at
) VALUES 
(
    'test-product-1'::uuid,
    'TEST001',
    '1234567890123',
    'Martillo Test 500g',
    'Martillo de prueba con mango de madera',
    'test-category-1'::uuid,
    15000.00,
    8000.00,
    5,
    true,
    NOW(),
    NOW()
),
(
    'test-product-2'::uuid,
    'TEST002',
    '1234567890124',
    'Destornillador Test Phillips',
    'Destornillador Phillips de prueba',
    'test-category-1'::uuid,
    3500.00,
    2000.00,
    10,
    true,
    NOW(),
    NOW()
),
(
    'test-product-3'::uuid,
    'TEST003',
    '1234567890125',
    'Tornillo Test 6x40mm',
    'Tornillo de prueba para madera',
    'test-category-3'::uuid,
    150.00,
    80.00,
    100,
    true,
    NOW(),
    NOW()
),
(
    'test-product-4'::uuid,
    'TEST004',
    '1234567890126',
    'Cable Test 2.5mm',
    'Cable eléctrico de prueba',
    'test-category-4'::uuid,
    2500.00,
    1500.00,
    20,
    true,
    NOW(),
    NOW()
),
(
    'test-product-5'::uuid,
    'TEST005',
    '1234567890127',
    'Cemento Test 25kg',
    'Saco de cemento de prueba',
    'test-category-2'::uuid,
    8500.00,
    6000.00,
    3,
    true,
    NOW(),
    NOW()
),
(
    'test-product-6'::uuid,
    'TEST006',
    '1234567890128',
    'Taladro Test 600W',
    'Taladro eléctrico de prueba',
    'test-category-1'::uuid,
    45000.00,
    30000.00,
    2,
    true,
    NOW(),
    NOW()
),
(
    'test-product-7'::uuid,
    'TEST007',
    '1234567890129',
    'Pintura Test Blanca 1L',
    'Pintura látex blanca de prueba',
    'test-category-2'::uuid,
    12000.00,
    7500.00,
    8,
    true,
    NOW(),
    NOW()
),
(
    'test-product-8'::uuid,
    'TEST008',
    '1234567890130',
    'Interruptor Test Simple',
    'Interruptor eléctrico simple de prueba',
    'test-category-4'::uuid,
    2800.00,
    1800.00,
    15,
    true,
    NOW(),
    NOW()
);

-- =====================================================
-- 6. STOCK CENTRAL DE PRUEBA
-- =====================================================

INSERT INTO stock_central (
    id, producto_id, sucursal_id, stock_actual,
    stock_reservado, stock_disponible,
    created_at, updated_at
) VALUES 
(
    gen_random_uuid(),
    'test-product-1'::uuid,
    'test-sucursal-1'::uuid,
    25,
    0,
    25,
    NOW(),
    NOW()
),
(
    gen_random_uuid(),
    'test-product-2'::uuid,
    'test-sucursal-1'::uuid,
    50,
    2,
    48,
    NOW(),
    NOW()
),
(
    gen_random_uuid(),
    'test-product-3'::uuid,
    'test-sucursal-1'::uuid,
    500,
    10,
    490,
    NOW(),
    NOW()
),
(
    gen_random_uuid(),
    'test-product-4'::uuid,
    'test-sucursal-1'::uuid,
    100,
    5,
    95,
    NOW(),
    NOW()
),
(
    gen_random_uuid(),
    'test-product-5'::uuid,
    'test-sucursal-1'::uuid,
    15,
    0,
    15,
    NOW(),
    NOW()
),
(
    gen_random_uuid(),
    'test-product-6'::uuid,
    'test-sucursal-1'::uuid,
    8,
    1,
    7,
    NOW(),
    NOW()
),
(
    gen_random_uuid(),
    'test-product-7'::uuid,
    'test-sucursal-1'::uuid,
    30,
    0,
    30,
    NOW(),
    NOW()
),
(
    gen_random_uuid(),
    'test-product-8'::uuid,
    'test-sucursal-1'::uuid,
    40,
    3,
    37,
    NOW(),
    NOW()
),
-- Stock para sucursal 2
(
    gen_random_uuid(),
    'test-product-1'::uuid,
    'test-sucursal-2'::uuid,
    15,
    0,
    15,
    NOW(),
    NOW()
),
(
    gen_random_uuid(),
    'test-product-2'::uuid,
    'test-sucursal-2'::uuid,
    30,
    1,
    29,
    NOW(),
    NOW()
),
(
    gen_random_uuid(),
    'test-product-3'::uuid,
    'test-sucursal-2'::uuid,
    200,
    5,
    195,
    NOW(),
    NOW()
),
(
    gen_random_uuid(),
    'test-product-4'::uuid,
    'test-sucursal-2'::uuid,
    50,
    2,
    48,
    NOW(),
    NOW()
);

-- =====================================================
-- 7. VENTAS DE PRUEBA
-- =====================================================

INSERT INTO ventas (
    id, numero_documento, tipo_documento, sucursal_id,
    terminal_id, cajero_id, vendedor_id, fecha_venta,
    subtotal, descuento_total, impuesto_total, total,
    estado, metodo_pago, observaciones,
    created_at, updated_at
) VALUES 
(
    'test-venta-1'::uuid,
    'TEST-V-001',
    'boleta',
    'test-sucursal-1'::uuid,
    'test-terminal-1'::uuid,
    'test-user-3'::uuid,
    'test-user-2'::uuid,
    NOW() - INTERVAL '2 days',
    25000.00,
    0.00,
    4750.00,
    29750.00,
    'procesado',
    'efectivo',
    'Venta de prueba 1',
    NOW() - INTERVAL '2 days',
    NOW() - INTERVAL '2 days'
),
(
    'test-venta-2'::uuid,
    'TEST-V-002',
    'boleta',
    'test-sucursal-1'::uuid,
    'test-terminal-1'::uuid,
    'test-user-3'::uuid,
    'test-user-2'::uuid,
    NOW() - INTERVAL '1 day',
    18500.00,
    1000.00,
    3325.00,
    20825.00,
    'procesado',
    'tarjeta_debito',
    'Venta de prueba 2 con descuento',
    NOW() - INTERVAL '1 day',
    NOW() - INTERVAL '1 day'
),
(
    'test-venta-3'::uuid,
    'TEST-V-003',
    'factura',
    'test-sucursal-1'::uuid,
    'test-terminal-2'::uuid,
    'test-user-3'::uuid,
    'test-user-2'::uuid,
    NOW(),
    45000.00,
    2000.00,
    8170.00,
    51170.00,
    'procesado',
    'tarjeta_credito',
    'Venta de prueba 3 - Factura',
    NOW(),
    NOW()
),
(
    'test-venta-4'::uuid,
    'TEST-V-004',
    'boleta',
    'test-sucursal-2'::uuid,
    'test-terminal-3'::uuid,
    'test-user-4'::uuid,
    'test-user-4'::uuid,
    NOW() - INTERVAL '3 hours',
    12000.00,
    0.00,
    2280.00,
    14280.00,
    'procesado',
    'efectivo',
    'Venta de prueba sucursal 2',
    NOW() - INTERVAL '3 hours',
    NOW() - INTERVAL '3 hours'
);

-- =====================================================
-- 8. DETALLE DE VENTAS DE PRUEBA
-- =====================================================

INSERT INTO detalle_ventas (
    id, venta_id, producto_id, cantidad, precio_unitario,
    descuento_unitario, subtotal_linea, numero_linea,
    created_at, updated_at
) VALUES 
-- Detalle Venta 1
(
    gen_random_uuid(),
    'test-venta-1'::uuid,
    'test-product-1'::uuid,
    1,
    15000.00,
    0.00,
    15000.00,
    1,
    NOW() - INTERVAL '2 days',
    NOW() - INTERVAL '2 days'
),
(
    gen_random_uuid(),
    'test-venta-1'::uuid,
    'test-product-2'::uuid,
    2,
    3500.00,
    0.00,
    7000.00,
    2,
    NOW() - INTERVAL '2 days',
    NOW() - INTERVAL '2 days'
),
(
    gen_random_uuid(),
    'test-venta-1'::uuid,
    'test-product-3'::uuid,
    20,
    150.00,
    0.00,
    3000.00,
    3,
    NOW() - INTERVAL '2 days',
    NOW() - INTERVAL '2 days'
),
-- Detalle Venta 2
(
    gen_random_uuid(),
    'test-venta-2'::uuid,
    'test-product-4'::uuid,
    5,
    2500.00,
    0.00,
    12500.00,
    1,
    NOW() - INTERVAL '1 day',
    NOW() - INTERVAL '1 day'
),
(
    gen_random_uuid(),
    'test-venta-2'::uuid,
    'test-product-8'::uuid,
    2,
    2800.00,
    500.00,
    5100.00,
    2,
    NOW() - INTERVAL '1 day',
    NOW() - INTERVAL '1 day'
),
(
    gen_random_uuid(),
    'test-venta-2'::uuid,
    'test-product-3'::uuid,
    6,
    150.00,
    0.00,
    900.00,
    3,
    NOW() - INTERVAL '1 day',
    NOW() - INTERVAL '1 day'
),
-- Detalle Venta 3
(
    gen_random_uuid(),
    'test-venta-3'::uuid,
    'test-product-6'::uuid,
    1,
    45000.00,
    2000.00,
    43000.00,
    1,
    NOW(),
    NOW()
),
-- Detalle Venta 4
(
    gen_random_uuid(),
    'test-venta-4'::uuid,
    'test-product-7'::uuid,
    1,
    12000.00,
    0.00,
    12000.00,
    1,
    NOW() - INTERVAL '3 hours',
    NOW() - INTERVAL '3 hours'
);

-- =====================================================
-- 9. MOVIMIENTOS DE STOCK DE PRUEBA
-- =====================================================

INSERT INTO movimientos_stock (
    id, producto_id, sucursal_id, tipo_movimiento,
    cantidad, stock_anterior, stock_nuevo, motivo,
    documento_referencia, usuario_id,
    created_at, updated_at
) VALUES 
(
    gen_random_uuid(),
    'test-product-1'::uuid,
    'test-sucursal-1'::uuid,
    'entrada',
    30,
    0,
    30,
    'Stock inicial de prueba',
    'INIT-001',
    'test-user-1'::uuid,
    NOW() - INTERVAL '7 days',
    NOW() - INTERVAL '7 days'
),
(
    gen_random_uuid(),
    'test-product-1'::uuid,
    'test-sucursal-1'::uuid,
    'salida',
    5,
    30,
    25,
    'Venta',
    'TEST-V-001',
    'test-user-3'::uuid,
    NOW() - INTERVAL '2 days',
    NOW() - INTERVAL '2 days'
),
(
    gen_random_uuid(),
    'test-product-2'::uuid,
    'test-sucursal-1'::uuid,
    'entrada',
    60,
    0,
    60,
    'Stock inicial de prueba',
    'INIT-002',
    'test-user-1'::uuid,
    NOW() - INTERVAL '7 days',
    NOW() - INTERVAL '7 days'
),
(
    gen_random_uuid(),
    'test-product-2'::uuid,
    'test-sucursal-1'::uuid,
    'salida',
    10,
    60,
    50,
    'Ventas múltiples',
    'VENTAS-VARIAS',
    'test-user-3'::uuid,
    NOW() - INTERVAL '1 day',
    NOW() - INTERVAL '1 day'
),
(
    gen_random_uuid(),
    'test-product-6'::uuid,
    'test-sucursal-1'::uuid,
    'entrada',
    10,
    0,
    10,
    'Stock inicial de prueba',
    'INIT-006',
    'test-user-1'::uuid,
    NOW() - INTERVAL '5 days',
    NOW() - INTERVAL '5 days'
),
(
    gen_random_uuid(),
    'test-product-6'::uuid,
    'test-sucursal-1'::uuid,
    'salida',
    2,
    10,
    8,
    'Venta',
    'TEST-V-003',
    'test-user-3'::uuid,
    NOW(),
    NOW()
);

-- =====================================================
-- 10. PLANTILLAS DE ETIQUETAS DE PRUEBA
-- =====================================================

INSERT INTO etiquetas_plantillas (
    id, nombre, descripcion, ancho_mm, alto_mm,
    elementos_json, activa,
    created_at, updated_at
) VALUES 
(
    'template-basic'::uuid,
    'Plantilla Básica Test',
    'Plantilla básica para tests con código y precio',
    50.0,
    30.0,
    '[
        {
            "tipo": "text",
            "campo": "nombre",
            "x": 2,
            "y": 2,
            "fuente": "Arial",
            "tamaño": 8,
            "negrita": true
        },
        {
            "tipo": "text",
            "campo": "codigo",
            "x": 2,
            "y": 8,
            "fuente": "Arial",
            "tamaño": 6
        },
        {
            "tipo": "text",
            "campo": "precio",
            "x": 2,
            "y": 14,
            "fuente": "Arial",
            "tamaño": 10,
            "negrita": true,
            "formato": "$#,##0"
        },
        {
            "tipo": "barcode",
            "campo": "codigo_barras",
            "x": 2,
            "y": 20,
            "ancho": 46,
            "alto": 8,
            "formato": "CODE128"
        }
    ]'::jsonb,
    true,
    NOW(),
    NOW()
),
(
    'template-advanced'::uuid,
    'Plantilla Avanzada Test',
    'Plantilla avanzada para tests con más elementos',
    70.0,
    40.0,
    '[
        {
            "tipo": "text",
            "campo": "nombre",
            "x": 2,
            "y": 2,
            "fuente": "Arial",
            "tamaño": 10,
            "negrita": true
        },
        {
            "tipo": "text",
            "campo": "descripcion",
            "x": 2,
            "y": 10,
            "fuente": "Arial",
            "tamaño": 6
        },
        {
            "tipo": "text",
            "campo": "codigo",
            "x": 2,
            "y": 16,
            "fuente": "Arial",
            "tamaño": 8
        },
        {
            "tipo": "text",
            "campo": "precio",
            "x": 40,
            "y": 16,
            "fuente": "Arial",
            "tamaño": 12,
            "negrita": true,
            "formato": "$#,##0"
        },
        {
            "tipo": "barcode",
            "campo": "codigo_barras",
            "x": 2,
            "y": 25,
            "ancho": 66,
            "alto": 12,
            "formato": "CODE128"
        }
    ]'::jsonb,
    true,
    NOW(),
    NOW()
);

-- =====================================================
-- VERIFICACIÓN DE DATOS INSERTADOS
-- =====================================================

-- Mostrar resumen de datos insertados
SELECT 'Sucursales' as tabla, COUNT(*) as registros FROM sucursales WHERE nombre LIKE 'Test%'
UNION ALL
SELECT 'Usuarios', COUNT(*) FROM usuarios WHERE username LIKE '%test%'
UNION ALL
SELECT 'Terminales', COUNT(*) FROM terminales WHERE codigo LIKE 'TERM%'
UNION ALL
SELECT 'Categorías', COUNT(*) FROM categorias_productos WHERE nombre LIKE 'Test%'
UNION ALL
SELECT 'Productos', COUNT(*) FROM productos WHERE codigo LIKE 'TEST%'
UNION ALL
SELECT 'Stock Central', COUNT(*) FROM stock_central WHERE producto_id IN (SELECT id FROM productos WHERE codigo LIKE 'TEST%')
UNION ALL
SELECT 'Ventas', COUNT(*) FROM ventas WHERE numero_documento LIKE 'TEST%'
UNION ALL
SELECT 'Detalle Ventas', COUNT(*) FROM detalle_ventas WHERE venta_id IN (SELECT id FROM ventas WHERE numero_documento LIKE 'TEST%')
UNION ALL
SELECT 'Movimientos Stock', COUNT(*) FROM movimientos_stock WHERE producto_id IN (SELECT id FROM productos WHERE codigo LIKE 'TEST%')
UNION ALL
SELECT 'Plantillas Etiquetas', COUNT(*) FROM etiquetas_plantillas WHERE nombre LIKE '%Test%';

-- =====================================================
-- CONSULTAS DE VERIFICACIÓN ADICIONALES
-- =====================================================

-- Verificar productos con stock
SELECT 
    p.codigo,
    p.nombre,
    sc.stock_actual,
    sc.stock_disponible,
    s.nombre as sucursal
FROM productos p
JOIN stock_central sc ON p.id = sc.producto_id
JOIN sucursales s ON sc.sucursal_id = s.id
WHERE p.codigo LIKE 'TEST%'
ORDER BY p.codigo, s.nombre;

-- Verificar ventas con totales
SELECT 
    v.numero_documento,
    v.fecha_venta,
    v.total,
    s.nombre as sucursal,
    u.username as cajero
FROM ventas v
JOIN sucursales s ON v.sucursal_id = s.id
JOIN usuarios u ON v.cajero_id = u.id
WHERE v.numero_documento LIKE 'TEST%'
ORDER BY v.fecha_venta DESC;

-- =====================================================
-- NOTAS IMPORTANTES
-- =====================================================

/*
NOTAS PARA USO:

1. Este script inserta datos de prueba completos para todas las APIs
2. Los IDs están hardcodeados para facilitar los tests
3. Las contraseñas están hasheadas con bcrypt (password: "password123")
4. Los datos incluyen relaciones completas entre tablas
5. Se incluyen diferentes escenarios: stock bajo, ventas, movimientos

USUARIOS DE PRUEBA:
- admin_test / password123 (Administrador)
- vendedor_test / password123 (Vendedor)
- cajero_test / password123 (Cajero)
- supervisor_test / password123 (Supervisor)

PRODUCTOS DE PRUEBA:
- TEST001 a TEST008 con diferentes categorías y precios
- Stock configurado en ambas sucursales
- Movimientos de entrada y salida

VENTAS DE PRUEBA:
- 4 ventas con diferentes métodos de pago
- Detalles completos con múltiples productos
- Fechas distribuidas en los últimos días

Para limpiar los datos de prueba, ejecutar las sentencias DELETE
comentadas al inicio del script.
*/

