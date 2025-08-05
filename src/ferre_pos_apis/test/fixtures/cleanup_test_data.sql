-- =====================================================
-- SCRIPT DE LIMPIEZA DE DATOS DE PRUEBA
-- Sistema FERRE-POS - Servidor Central
-- =====================================================

-- IMPORTANTE: Este script elimina TODOS los datos de prueba
-- Ejecutar solo en entornos de desarrollo/testing

-- =====================================================
-- LIMPIEZA EN ORDEN CORRECTO (RESPETANDO FOREIGN KEYS)
-- =====================================================

-- 1. Eliminar trabajos de impresión de etiquetas
DELETE FROM etiquetas_trabajos_impresion 
WHERE plantilla_id IN (
    SELECT id FROM etiquetas_plantillas 
    WHERE nombre LIKE '%Test%'
);

-- 2. Eliminar detalles de ventas
DELETE FROM detalle_ventas 
WHERE venta_id IN (
    SELECT id FROM ventas 
    WHERE numero_documento LIKE 'TEST%'
);

-- 3. Eliminar ventas
DELETE FROM ventas 
WHERE numero_documento LIKE 'TEST%';

-- 4. Eliminar movimientos de stock
DELETE FROM movimientos_stock 
WHERE producto_id IN (
    SELECT id FROM productos 
    WHERE codigo LIKE 'TEST%'
);

-- 5. Eliminar stock central
DELETE FROM stock_central 
WHERE producto_id IN (
    SELECT id FROM productos 
    WHERE codigo LIKE 'TEST%'
);

-- 6. Eliminar productos
DELETE FROM productos 
WHERE codigo LIKE 'TEST%';

-- 7. Eliminar categorías de productos
DELETE FROM categorias_productos 
WHERE nombre LIKE 'Test%';

-- 8. Eliminar plantillas de etiquetas
DELETE FROM etiquetas_plantillas 
WHERE nombre LIKE '%Test%';

-- 9. Eliminar terminales
DELETE FROM terminales 
WHERE codigo LIKE 'TERM%';

-- 10. Eliminar usuarios
DELETE FROM usuarios 
WHERE username LIKE '%test%';

-- 11. Eliminar sucursales
DELETE FROM sucursales 
WHERE nombre LIKE 'Test%';

-- =====================================================
-- VERIFICACIÓN DE LIMPIEZA
-- =====================================================

-- Mostrar conteo después de la limpieza
SELECT 'Sucursales' as tabla, COUNT(*) as registros_restantes 
FROM sucursales WHERE nombre LIKE 'Test%'
UNION ALL
SELECT 'Usuarios', COUNT(*) 
FROM usuarios WHERE username LIKE '%test%'
UNION ALL
SELECT 'Terminales', COUNT(*) 
FROM terminales WHERE codigo LIKE 'TERM%'
UNION ALL
SELECT 'Categorías', COUNT(*) 
FROM categorias_productos WHERE nombre LIKE 'Test%'
UNION ALL
SELECT 'Productos', COUNT(*) 
FROM productos WHERE codigo LIKE 'TEST%'
UNION ALL
SELECT 'Stock Central', COUNT(*) 
FROM stock_central WHERE producto_id IN (
    SELECT id FROM productos WHERE codigo LIKE 'TEST%'
)
UNION ALL
SELECT 'Ventas', COUNT(*) 
FROM ventas WHERE numero_documento LIKE 'TEST%'
UNION ALL
SELECT 'Detalle Ventas', COUNT(*) 
FROM detalle_ventas WHERE venta_id IN (
    SELECT id FROM ventas WHERE numero_documento LIKE 'TEST%'
)
UNION ALL
SELECT 'Movimientos Stock', COUNT(*) 
FROM movimientos_stock WHERE producto_id IN (
    SELECT id FROM productos WHERE codigo LIKE 'TEST%'
)
UNION ALL
SELECT 'Plantillas Etiquetas', COUNT(*) 
FROM etiquetas_plantillas WHERE nombre LIKE '%Test%';

-- =====================================================
-- RESET DE SECUENCIAS (SI ES NECESARIO)
-- =====================================================

-- Si se usan secuencias para IDs numéricos, resetear aquí
-- ALTER SEQUENCE nombre_secuencia RESTART WITH 1;

-- =====================================================
-- VACUUM Y ANALYZE (OPCIONAL)
-- =====================================================

-- Limpiar espacio y actualizar estadísticas después de eliminaciones masivas
-- VACUUM ANALYZE;

PRINT 'Limpieza de datos de prueba completada.';
PRINT 'Verificar el conteo de registros restantes arriba.';
PRINT 'Todos los conteos deberían ser 0 para una limpieza exitosa.';

