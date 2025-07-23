-- Índices compuestos más específicos para consultas frecuentes
CREATE INDEX idx_productos_busqueda_completa ON productos 
(activo, codigo_barra, codigo_interno, precio_unitario) 
WHERE activo = true;

-- Índice para búsquedas de texto con ranking
CREATE INDEX idx_productos_texto_ranking ON productos 
USING GIN (descripcion_busqueda, (popularidad_score)) 
WHERE activo = true;

-- Índice para stock con cantidad exacta
CREATE INDEX idx_stock_cantidad_exacta ON stock_central 
(producto_id, sucursal_id, cantidad_disponible) 
WHERE cantidad_disponible > 0;

-- Particionar ventas por día en lugar de mes para mejor rendimiento
CREATE TABLE ventas_2025_07_19 PARTITION OF ventas
    FOR VALUES FROM ('2025-07-19') TO ('2025-07-20');

-- Usar particionado hash para detalle_ventas
ALTER TABLE detalle_ventas DETACH PARTITION detalle_ventas_p1;
CREATE TABLE detalle_ventas (
    -- mantener estructura actual
) PARTITION BY HASH (venta_id);

-- Crear 8 particiones hash para mejor distribución
CREATE TABLE detalle_ventas_h0 PARTITION OF detalle_ventas FOR VALUES WITH (modulus 8, remainder 0);
-- ...continuar hasta h7

-- Tabla de cache de sesiones en memoria
CREATE UNLOGGED TABLE cache_sesiones_activas (
    token_hash TEXT PRIMARY KEY,
    usuario_id UUID NOT NULL,
    permisos JSONB NOT NULL,
    fecha_expiracion TIMESTAMP NOT NULL
);
CREATE INDEX idx_cache_sesiones_usuario ON cache_sesiones_activas(usuario_id);

-- Tabla de cache de productos frecuentes
CREATE UNLOGGED TABLE cache_productos_frecuentes (
    codigo_busqueda TEXT PRIMARY KEY,
    datos_producto JSONB NOT NULL,
    stock_sucursales JSONB NOT NULL,
    fecha_cache TIMESTAMP DEFAULT NOW()
);

-- Agregar campos calculados a productos para evitar JOINs
ALTER TABLE productos ADD COLUMN stock_total_sistema INTEGER DEFAULT 0;
ALTER TABLE productos ADD COLUMN categoria_nombre TEXT;
ALTER TABLE productos ADD COLUMN ultima_venta TIMESTAMP;

-- Trigger para mantener sincronizados
CREATE OR REPLACE FUNCTION actualizar_cache_producto()
RETURNS TRIGGER AS $$
BEGIN
    -- Actualizar stock total
    UPDATE productos SET 
        stock_total_sistema = (
            SELECT COALESCE(SUM(cantidad_disponible), 0) 
            FROM stock_central 
            WHERE producto_id = NEW.producto_id
        )
    WHERE id = NEW.producto_id;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Función ultra-optimizada para validación de stock
CREATE OR REPLACE FUNCTION stock_disponible_rapido(
    p_codigo_producto TEXT,
    p_sucursal_id UUID
) RETURNS INTEGER AS $$
DECLARE
    v_stock INTEGER;
BEGIN
    -- Usar cache primero
    SELECT stock_disponible INTO v_stock
    FROM cache_stock_rapido 
    WHERE codigo_producto = p_codigo_producto 
      AND sucursal_id = p_sucursal_id
      AND fecha_cache > NOW() - INTERVAL '30 seconds';
    
    IF FOUND THEN
        RETURN v_stock;
    END IF;
    
    -- Consulta directa con índice optimizado
    SELECT sc.cantidad_disponible INTO v_stock
    FROM stock_central sc
    JOIN productos p ON sc.producto_id = p.id
    WHERE (p.codigo_barra = p_codigo_producto OR p.codigo_interno = p_codigo_producto)
      AND sc.sucursal_id = p_sucursal_id
      AND p.activo = true;
    
    -- Actualizar cache
    INSERT INTO cache_stock_rapido (codigo_producto, sucursal_id, stock_disponible)
    VALUES (p_codigo_producto, p_sucursal_id, COALESCE(v_stock, 0))
    ON CONFLICT (codigo_producto, sucursal_id) DO UPDATE SET
        stock_disponible = EXCLUDED.stock_disponible,
        fecha_cache = NOW();
    
    RETURN COALESCE(v_stock, 0);
END;
$$ LANGUAGE plpgsql;

-- Configuraciones específicas para velocidad extrema
ALTER SYSTEM SET shared_buffers = '1GB';  -- Aumentar para servidor central
ALTER SYSTEM SET effective_cache_size = '4GB';
ALTER SYSTEM SET work_mem = '16MB';  -- Para operaciones complejas
ALTER SYSTEM SET maintenance_work_mem = '256MB';
ALTER SYSTEM SET wal_compression = on;
ALTER SYSTEM SET checkpoint_completion_target = 0.95;
ALTER SYSTEM SET default_statistics_target = 500;  -- Mejores estadísticas

-- Para tablas críticas de POS
ALTER TABLE productos SET (parallel_workers = 4);
ALTER TABLE stock_central SET (parallel_workers = 4);
ALTER TABLE ventas SET (parallel_workers = 2);

-- Sistema de cache por niveles
CREATE TABLE cache_nivel1_productos (
    codigo TEXT PRIMARY KEY,
    datos JSONB NOT NULL,
    hits INTEGER DEFAULT 0,
    fecha_cache TIMESTAMP DEFAULT NOW()
);

-- Cache de agregaciones para reportes
CREATE MATERIALIZED VIEW cache_ventas_hora AS
SELECT 
    DATE_TRUNC('hour', fecha) as hora,
    sucursal_id,
    COUNT(*) as total_ventas,
    SUM(total) as monto_total,
    string_agg(DISTINCT cajero_id::text, ',') as cajeros_activos
FROM ventas
WHERE fecha >= CURRENT_DATE - INTERVAL '24 hours'
  AND estado = 'finalizada'
GROUP BY DATE_TRUNC('hour', fecha), sucursal_id;

-- Refrescar cada 5 minutos

-- Pool de conexiones diferenciado por prioridad
-- En aplicación Go, configurar:
-- api_pos: max_conns=40, acquire_timeout=100ms
-- api_sync: max_conns=10, acquire_timeout=500ms  
-- api_labels: max_conns=5, acquire_timeout=1s
-- api_report: max_conns=3, acquire_timeout=5s

-- Prepared statements para operaciones frecuentes
PREPARE buscar_producto_rapido(TEXT, UUID) AS
SELECT p.id, p.codigo_barra, p.descripcion_corta, p.precio_unitario,
       sc.cantidad_disponible
FROM productos p
JOIN stock_central sc ON p.id = sc.producto_id
WHERE (p.codigo_barra = $1 OR p.codigo_interno = $1)
  AND sc.sucursal_id = $2
  AND p.activo = true;
  
-- Vista para monitorear consultas lentas por proceso API
CREATE VIEW monitor_consultas_por_api AS
SELECT 
    application_name,
    query,
    calls,
    mean_exec_time,
    max_exec_time,
    rows
FROM pg_stat_statements pss
JOIN pg_stat_activity psa ON pss.userid = psa.usesysid
WHERE application_name LIKE 'ferre_pos_%'
ORDER BY mean_exec_time DESC;

