-- =====================================================
-- ESQUEMA SQL MÓDULO DE ETIQUETAS - SISTEMA FERRE-POS
-- =====================================================
-- Versión: 1.0
-- Fecha: Julio 2025
-- Autor: Manus AI
-- Descripción: Esquema específico para el módulo de etiquetas
-- =====================================================

-- Extensiones necesarias para el módulo de etiquetas
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- =====================================================
-- TIPOS DE DATOS PERSONALIZADOS
-- =====================================================

-- Tipo para estados de trabajos de impresión
CREATE TYPE estado_trabajo_impresion AS ENUM (
    'pendiente', 
    'en_proceso', 
    'completado', 
    'error', 
    'cancelado'
);

-- Tipo para tipos de impresora
CREATE TYPE tipo_impresora AS ENUM (
    'termica_directa', 
    'transferencia_termica', 
    'laser', 
    'inyeccion_tinta'
);

-- Tipo para formatos de etiqueta
CREATE TYPE formato_etiqueta AS ENUM (
    'pequena',     -- 30x20mm - productos pequeños
    'mediana',     -- 50x30mm - herramientas
    'grande',      -- 70x50mm - materiales construcción
    'extra_grande' -- 100x70mm - productos especiales
);

-- =====================================================
-- TABLAS DEL MÓDULO DE ETIQUETAS
-- =====================================================

-- Tabla: etiquetas_plantillas
-- Descripción: Plantillas de diseño para diferentes tipos de etiquetas
CREATE TABLE etiquetas_plantillas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    nombre TEXT NOT NULL,
    descripcion TEXT,
    formato formato_etiqueta NOT NULL,
    ancho_mm NUMERIC(5,2) NOT NULL,
    alto_mm NUMERIC(5,2) NOT NULL,
    categoria_producto_id UUID REFERENCES categorias_productos(id),
    definicion_json JSONB NOT NULL,
    activa BOOLEAN DEFAULT true,
    predeterminada BOOLEAN DEFAULT false,
    sucursal_id UUID REFERENCES sucursales(id),
    usuario_creacion UUID REFERENCES usuarios(id),
    fecha_creacion TIMESTAMP DEFAULT NOW(),
    fecha_modificacion TIMESTAMP DEFAULT NOW(),
    veces_utilizada INTEGER DEFAULT 0,
    CONSTRAINT chk_dimensiones_positivas CHECK (ancho_mm > 0 AND alto_mm > 0),
    CONSTRAINT chk_dimensiones_realistas CHECK (
        ancho_mm BETWEEN 10 AND 200 AND alto_mm BETWEEN 10 AND 200
    )
);

-- Tabla: etiquetas_configuraciones_impresora
-- Descripción: Configuraciones específicas por impresora y sucursal
CREATE TABLE etiquetas_configuraciones_impresora (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    nombre TEXT NOT NULL,
    sucursal_id UUID REFERENCES sucursales(id),
    tipo tipo_impresora NOT NULL,
    marca TEXT,
    modelo TEXT,
    driver_nombre TEXT,
    resolucion_dpi INTEGER DEFAULT 203,
    velocidad_impresion INTEGER DEFAULT 4,
    ancho_papel_mm NUMERIC(5,2),
    configuracion_driver JSONB,
    activa BOOLEAN DEFAULT true,
    predeterminada BOOLEAN DEFAULT false,
    fecha_creacion TIMESTAMP DEFAULT NOW(),
    fecha_modificacion TIMESTAMP DEFAULT NOW(),
    CONSTRAINT chk_resolucion_valida CHECK (resolucion_dpi IN (203, 300, 600)),
    CONSTRAINT chk_velocidad_valida CHECK (velocidad_impresion BETWEEN 1 AND 10),
    UNIQUE(sucursal_id, nombre)
);

-- Tabla: etiquetas_trabajos_impresion
-- Descripción: Registro de todos los trabajos de impresión realizados
CREATE TABLE etiquetas_trabajos_impresion (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    numero_trabajo BIGSERIAL UNIQUE,
    usuario_id UUID REFERENCES usuarios(id),
    sucursal_id UUID REFERENCES sucursales(id),
    plantilla_id UUID REFERENCES etiquetas_plantillas(id),
    impresora_id UUID REFERENCES etiquetas_configuraciones_impresora(id),
    tipo_trabajo TEXT NOT NULL,
    productos_incluidos JSONB NOT NULL,
    cantidad_total INTEGER NOT NULL,
    cantidad_exitosa INTEGER DEFAULT 0,
    cantidad_error INTEGER DEFAULT 0,
    estado estado_trabajo_impresion DEFAULT 'pendiente',
    fecha_inicio TIMESTAMP DEFAULT NOW(),
    fecha_fin TIMESTAMP,
    tiempo_procesamiento_ms INTEGER,
    mensaje_error TEXT,
    configuracion_utilizada JSONB,
    datos_adicionales JSONB,
    CONSTRAINT chk_tipo_trabajo CHECK (tipo_trabajo IN (
        'individual', 'masivo', 'reimpresion', 'promocional'
    )),
    CONSTRAINT chk_cantidades_positivas CHECK (
        cantidad_total > 0 AND cantidad_exitosa >= 0 AND cantidad_error >= 0
    ),
    CONSTRAINT chk_cantidades_coherentes CHECK (
        cantidad_exitosa + cantidad_error <= cantidad_total
    )
);

-- Tabla: etiquetas_preferencias_usuario
-- Descripción: Preferencias personalizadas por usuario
CREATE TABLE etiquetas_preferencias_usuario (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    usuario_id UUID REFERENCES usuarios(id),
    plantilla_favorita_id UUID REFERENCES etiquetas_plantillas(id),
    impresora_predeterminada_id UUID REFERENCES etiquetas_configuraciones_impresora(id),
    formato_preferido formato_etiqueta,
    configuracion_interfaz JSONB,
    atajos_personalizados JSONB,
    fecha_creacion TIMESTAMP DEFAULT NOW(),
    fecha_modificacion TIMESTAMP DEFAULT NOW(),
    UNIQUE(usuario_id)
);

-- Tabla: etiquetas_logs_operaciones
-- Descripción: Log detallado de todas las operaciones del módulo
CREATE TABLE etiquetas_logs_operaciones (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    usuario_id UUID REFERENCES usuarios(id),
    sucursal_id UUID REFERENCES sucursales(id),
    operacion TEXT NOT NULL,
    entidad_tipo TEXT,
    entidad_id UUID,
    parametros_entrada JSONB,
    resultado JSONB,
    duracion_ms INTEGER,
    ip_origen INET,
    user_agent TEXT,
    fecha TIMESTAMP DEFAULT NOW(),
    CONSTRAINT chk_operacion_valida CHECK (operacion IN (
        'buscar_productos', 'generar_codigo_barras', 'crear_plantilla',
        'modificar_plantilla', 'eliminar_plantilla', 'vista_previa',
        'imprimir_etiquetas', 'configurar_impresora'
    ))
);

-- Tabla: etiquetas_codigos_barras_cache
-- Descripción: Cache de códigos de barras generados para mejorar rendimiento
CREATE TABLE etiquetas_codigos_barras_cache (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    codigo_producto TEXT NOT NULL,
    formato_codigo TEXT DEFAULT 'code39',
    configuracion_hash TEXT NOT NULL,
    imagen_svg TEXT,
    imagen_png BYTEA,
    ancho_px INTEGER,
    alto_px INTEGER,
    fecha_generacion TIMESTAMP DEFAULT NOW(),
    fecha_ultimo_uso TIMESTAMP DEFAULT NOW(),
    veces_utilizado INTEGER DEFAULT 1,
    UNIQUE(codigo_producto, configuracion_hash)
);

-- =====================================================
-- ÍNDICES ESPECIALIZADOS
-- =====================================================

-- Índices para tabla etiquetas_plantillas
CREATE INDEX idx_etiquetas_plantillas_formato ON etiquetas_plantillas(formato);
CREATE INDEX idx_etiquetas_plantillas_categoria ON etiquetas_plantillas(categoria_producto_id);
CREATE INDEX idx_etiquetas_plantillas_activa ON etiquetas_plantillas(activa);
CREATE INDEX idx_etiquetas_plantillas_sucursal ON etiquetas_plantillas(sucursal_id);
CREATE INDEX idx_etiquetas_plantillas_predeterminada ON etiquetas_plantillas(predeterminada) WHERE predeterminada = true;
CREATE INDEX idx_etiquetas_plantillas_uso ON etiquetas_plantillas(veces_utilizada DESC);

-- Índices para tabla etiquetas_configuraciones_impresora
CREATE INDEX idx_etiquetas_config_impresora_sucursal ON etiquetas_configuraciones_impresora(sucursal_id);
CREATE INDEX idx_etiquetas_config_impresora_tipo ON etiquetas_configuraciones_impresora(tipo);
CREATE INDEX idx_etiquetas_config_impresora_activa ON etiquetas_configuraciones_impresora(activa);
CREATE INDEX idx_etiquetas_config_impresora_predeterminada ON etiquetas_configuraciones_impresora(predeterminada) WHERE predeterminada = true;

-- Índices para tabla etiquetas_trabajos_impresion
CREATE INDEX idx_etiquetas_trabajos_usuario ON etiquetas_trabajos_impresion(usuario_id);
CREATE INDEX idx_etiquetas_trabajos_sucursal ON etiquetas_trabajos_impresion(sucursal_id);
CREATE INDEX idx_etiquetas_trabajos_fecha ON etiquetas_trabajos_impresion(fecha_inicio);
CREATE INDEX idx_etiquetas_trabajos_estado ON etiquetas_trabajos_impresion(estado);
CREATE INDEX idx_etiquetas_trabajos_tipo ON etiquetas_trabajos_impresion(tipo_trabajo);
CREATE INDEX idx_etiquetas_trabajos_plantilla ON etiquetas_trabajos_impresion(plantilla_id);

-- Índices para tabla etiquetas_logs_operaciones
CREATE INDEX idx_etiquetas_logs_usuario ON etiquetas_logs_operaciones(usuario_id);
CREATE INDEX idx_etiquetas_logs_fecha ON etiquetas_logs_operaciones(fecha);
CREATE INDEX idx_etiquetas_logs_operacion ON etiquetas_logs_operaciones(operacion);
CREATE INDEX idx_etiquetas_logs_entidad ON etiquetas_logs_operaciones(entidad_tipo, entidad_id);

-- Índices para tabla etiquetas_codigos_barras_cache
CREATE INDEX idx_etiquetas_cache_codigo ON etiquetas_codigos_barras_cache(codigo_producto);
CREATE INDEX idx_etiquetas_cache_ultimo_uso ON etiquetas_codigos_barras_cache(fecha_ultimo_uso);
CREATE INDEX idx_etiquetas_cache_generacion ON etiquetas_codigos_barras_cache(fecha_generacion);

-- =====================================================
-- FUNCIONES ESPECIALIZADAS
-- =====================================================

-- Función: obtener_plantilla_por_producto
-- Descripción: Obtiene la plantilla más apropiada para un producto específico
CREATE OR REPLACE FUNCTION obtener_plantilla_por_producto(
    p_producto_id UUID,
    p_sucursal_id UUID DEFAULT NULL
) RETURNS UUID AS $$
DECLARE
    v_plantilla_id UUID;
    v_categoria_id UUID;
BEGIN
    -- Obtener categoría del producto
    SELECT categoria_id INTO v_categoria_id
    FROM productos
    WHERE id = p_producto_id;
    
    -- Buscar plantilla específica para la categoría y sucursal
    SELECT id INTO v_plantilla_id
    FROM etiquetas_plantillas
    WHERE categoria_producto_id = v_categoria_id
      AND (sucursal_id = p_sucursal_id OR sucursal_id IS NULL)
      AND activa = true
    ORDER BY 
        CASE WHEN sucursal_id = p_sucursal_id THEN 1 ELSE 2 END,
        predeterminada DESC,
        veces_utilizada DESC
    LIMIT 1;
    
    -- Si no hay plantilla específica, buscar plantilla general
    IF v_plantilla_id IS NULL THEN
        SELECT id INTO v_plantilla_id
        FROM etiquetas_plantillas
        WHERE categoria_producto_id IS NULL
          AND (sucursal_id = p_sucursal_id OR sucursal_id IS NULL)
          AND activa = true
        ORDER BY 
            CASE WHEN sucursal_id = p_sucursal_id THEN 1 ELSE 2 END,
            predeterminada DESC,
            veces_utilizada DESC
        LIMIT 1;
    END IF;
    
    RETURN v_plantilla_id;
END;
$$ LANGUAGE plpgsql;

-- Función: generar_codigo_barras_code39
-- Descripción: Valida y prepara código para generación Code 39
CREATE OR REPLACE FUNCTION generar_codigo_barras_code39(
    p_codigo TEXT
) RETURNS JSONB AS $$
DECLARE
    v_codigo_limpio TEXT;
    v_resultado JSONB;
BEGIN
    -- Limpiar y validar código
    v_codigo_limpio := UPPER(TRIM(p_codigo));
    
    -- Validar caracteres permitidos en Code 39
    IF v_codigo_limpio !~ '^[0-9A-Z\-\.\$\/\+\%\s]*$' THEN
        RETURN jsonb_build_object(
            'valido', false,
            'error', 'Código contiene caracteres no válidos para Code 39'
        );
    END IF;
    
    -- Validar longitud
    IF LENGTH(v_codigo_limpio) > 43 THEN
        RETURN jsonb_build_object(
            'valido', false,
            'error', 'Código excede longitud máxima para Code 39'
        );
    END IF;
    
    -- Retornar código válido
    RETURN jsonb_build_object(
        'valido', true,
        'codigo_limpio', v_codigo_limpio,
        'longitud', LENGTH(v_codigo_limpio),
        'checksum_requerido', LENGTH(v_codigo_limpio) > 20
    );
END;
$$ LANGUAGE plpgsql;

-- Función: registrar_trabajo_impresion
-- Descripción: Registra un nuevo trabajo de impresión
CREATE OR REPLACE FUNCTION registrar_trabajo_impresion(
    p_usuario_id UUID,
    p_sucursal_id UUID,
    p_plantilla_id UUID,
    p_impresora_id UUID,
    p_tipo_trabajo TEXT,
    p_productos JSONB,
    p_cantidad_total INTEGER
) RETURNS UUID AS $$
DECLARE
    v_trabajo_id UUID;
BEGIN
    INSERT INTO etiquetas_trabajos_impresion (
        usuario_id, sucursal_id, plantilla_id, impresora_id,
        tipo_trabajo, productos_incluidos, cantidad_total
    ) VALUES (
        p_usuario_id, p_sucursal_id, p_plantilla_id, p_impresora_id,
        p_tipo_trabajo, p_productos, p_cantidad_total
    ) RETURNING id INTO v_trabajo_id;
    
    -- Incrementar contador de uso de plantilla
    UPDATE etiquetas_plantillas
    SET veces_utilizada = veces_utilizada + 1
    WHERE id = p_plantilla_id;
    
    RETURN v_trabajo_id;
END;
$$ LANGUAGE plpgsql;

-- Función: actualizar_estado_trabajo
-- Descripción: Actualiza el estado de un trabajo de impresión
CREATE OR REPLACE FUNCTION actualizar_estado_trabajo(
    p_trabajo_id UUID,
    p_estado estado_trabajo_impresion,
    p_cantidad_exitosa INTEGER DEFAULT NULL,
    p_cantidad_error INTEGER DEFAULT NULL,
    p_mensaje_error TEXT DEFAULT NULL
) RETURNS BOOLEAN AS $$
BEGIN
    UPDATE etiquetas_trabajos_impresion
    SET estado = p_estado,
        cantidad_exitosa = COALESCE(p_cantidad_exitosa, cantidad_exitosa),
        cantidad_error = COALESCE(p_cantidad_error, cantidad_error),
        mensaje_error = p_mensaje_error,
        fecha_fin = CASE WHEN p_estado IN ('completado', 'error', 'cancelado') 
                         THEN NOW() ELSE fecha_fin END,
        tiempo_procesamiento_ms = CASE WHEN p_estado IN ('completado', 'error', 'cancelado')
                                      THEN EXTRACT(EPOCH FROM (NOW() - fecha_inicio)) * 1000
                                      ELSE tiempo_procesamiento_ms END
    WHERE id = p_trabajo_id;
    
    RETURN FOUND;
END;
$$ LANGUAGE plpgsql;

-- Función: limpiar_cache_codigos_barras
-- Descripción: Limpia códigos de barras antiguos del cache
CREATE OR REPLACE FUNCTION limpiar_cache_codigos_barras(
    p_dias_antiguedad INTEGER DEFAULT 30
) RETURNS INTEGER AS $$
DECLARE
    v_eliminados INTEGER;
BEGIN
    DELETE FROM etiquetas_codigos_barras_cache
    WHERE fecha_ultimo_uso < NOW() - INTERVAL '1 day' * p_dias_antiguedad;
    
    GET DIAGNOSTICS v_eliminados = ROW_COUNT;
    
    RETURN v_eliminados;
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- TRIGGERS
-- =====================================================

-- Trigger: Actualizar fecha_modificacion en plantillas
CREATE OR REPLACE FUNCTION actualizar_fecha_modificacion_etiquetas()
RETURNS TRIGGER AS $$
BEGIN
    NEW.fecha_modificacion = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_etiquetas_plantillas_fecha_mod
    BEFORE UPDATE ON etiquetas_plantillas
    FOR EACH ROW EXECUTE FUNCTION actualizar_fecha_modificacion_etiquetas();

CREATE TRIGGER trg_etiquetas_config_impresora_fecha_mod
    BEFORE UPDATE ON etiquetas_configuraciones_impresora
    FOR EACH ROW EXECUTE FUNCTION actualizar_fecha_modificacion_etiquetas();

CREATE TRIGGER trg_etiquetas_preferencias_fecha_mod
    BEFORE UPDATE ON etiquetas_preferencias_usuario
    FOR EACH ROW EXECUTE FUNCTION actualizar_fecha_modificacion_etiquetas();

-- Trigger: Validar plantilla predeterminada única por categoría
CREATE OR REPLACE FUNCTION validar_plantilla_predeterminada()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.predeterminada = true THEN
        -- Desactivar otras plantillas predeterminadas para la misma categoría
        UPDATE etiquetas_plantillas
        SET predeterminada = false
        WHERE categoria_producto_id = NEW.categoria_producto_id
          AND sucursal_id = NEW.sucursal_id
          AND id != NEW.id
          AND predeterminada = true;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_validar_plantilla_predeterminada
    AFTER INSERT OR UPDATE ON etiquetas_plantillas
    FOR EACH ROW EXECUTE FUNCTION validar_plantilla_predeterminada();

-- =====================================================
-- VISTAS ESPECIALIZADAS
-- =====================================================

-- Vista: vista_plantillas_disponibles
-- Descripción: Plantillas disponibles con información de categoría
CREATE VIEW vista_plantillas_disponibles AS
SELECT 
    ep.id,
    ep.nombre,
    ep.descripcion,
    ep.formato,
    ep.ancho_mm,
    ep.alto_mm,
    ep.predeterminada,
    ep.veces_utilizada,
    cp.nombre AS categoria_nombre,
    cp.codigo AS categoria_codigo,
    s.nombre AS sucursal_nombre,
    u.nombre AS usuario_creacion_nombre,
    ep.fecha_creacion,
    ep.fecha_modificacion
FROM etiquetas_plantillas ep
LEFT JOIN categorias_productos cp ON ep.categoria_producto_id = cp.id
LEFT JOIN sucursales s ON ep.sucursal_id = s.id
LEFT JOIN usuarios u ON ep.usuario_creacion = u.id
WHERE ep.activa = true
ORDER BY ep.predeterminada DESC, ep.veces_utilizada DESC;

-- Vista: vista_trabajos_impresion_resumen
-- Descripción: Resumen de trabajos de impresión con estadísticas
CREATE VIEW vista_trabajos_impresion_resumen AS
SELECT 
    eti.id,
    eti.numero_trabajo,
    u.nombre AS usuario_nombre,
    s.nombre AS sucursal_nombre,
    ep.nombre AS plantilla_nombre,
    eip.nombre AS impresora_nombre,
    eti.tipo_trabajo,
    eti.cantidad_total,
    eti.cantidad_exitosa,
    eti.cantidad_error,
    eti.estado,
    eti.fecha_inicio,
    eti.fecha_fin,
    eti.tiempo_procesamiento_ms,
    CASE 
        WHEN eti.cantidad_total > 0 THEN 
            ROUND((eti.cantidad_exitosa::NUMERIC / eti.cantidad_total) * 100, 2)
        ELSE 0 
    END AS porcentaje_exito
FROM etiquetas_trabajos_impresion eti
JOIN usuarios u ON eti.usuario_id = u.id
JOIN sucursales s ON eti.sucursal_id = s.id
LEFT JOIN etiquetas_plantillas ep ON eti.plantilla_id = ep.id
LEFT JOIN etiquetas_configuraciones_impresora eip ON eti.impresora_id = eip.id
ORDER BY eti.fecha_inicio DESC;

-- =====================================================
-- DATOS INICIALES
-- =====================================================

-- Plantillas predeterminadas para diferentes formatos
INSERT INTO etiquetas_plantillas (nombre, descripcion, formato, ancho_mm, alto_mm, definicion_json, activa, predeterminada) VALUES
(
    'Etiqueta Pequeña Estándar',
    'Plantilla estándar para productos pequeños de ferretería',
    'pequena',
    30.0,
    20.0,
    '{
        "elementos": [
            {
                "tipo": "texto",
                "campo": "codigo_interno",
                "x": 2,
                "y": 2,
                "ancho": 26,
                "alto": 3,
                "fuente": "Arial",
                "tamano": 8,
                "negrita": true
            },
            {
                "tipo": "texto",
                "campo": "descripcion",
                "x": 2,
                "y": 6,
                "ancho": 26,
                "alto": 6,
                "fuente": "Arial",
                "tamano": 6,
                "truncar": true
            },
            {
                "tipo": "texto",
                "campo": "precio_unitario",
                "x": 2,
                "y": 13,
                "ancho": 12,
                "alto": 4,
                "fuente": "Arial",
                "tamano": 10,
                "negrita": true,
                "prefijo": "$"
            },
            {
                "tipo": "codigo_barras",
                "campo": "codigo_interno",
                "x": 15,
                "y": 13,
                "ancho": 13,
                "alto": 5,
                "formato": "code39",
                "mostrar_texto": false
            }
        ]
    }',
    true,
    true
),
(
    'Etiqueta Mediana Estándar',
    'Plantilla estándar para herramientas y productos medianos',
    'mediana',
    50.0,
    30.0,
    '{
        "elementos": [
            {
                "tipo": "texto",
                "campo": "codigo_interno",
                "x": 2,
                "y": 2,
                "ancho": 20,
                "alto": 4,
                "fuente": "Arial",
                "tamano": 10,
                "negrita": true
            },
            {
                "tipo": "texto",
                "campo": "marca",
                "x": 25,
                "y": 2,
                "ancho": 23,
                "alto": 4,
                "fuente": "Arial",
                "tamano": 8,
                "alineacion": "derecha"
            },
            {
                "tipo": "texto",
                "campo": "descripcion",
                "x": 2,
                "y": 7,
                "ancho": 46,
                "alto": 8,
                "fuente": "Arial",
                "tamano": 8,
                "truncar": true
            },
            {
                "tipo": "texto",
                "campo": "precio_unitario",
                "x": 2,
                "y": 16,
                "ancho": 20,
                "alto": 6,
                "fuente": "Arial",
                "tamano": 14,
                "negrita": true,
                "prefijo": "$"
            },
            {
                "tipo": "codigo_barras",
                "campo": "codigo_interno",
                "x": 25,
                "y": 16,
                "ancho": 23,
                "alto": 12,
                "formato": "code39",
                "mostrar_texto": true
            }
        ]
    }',
    true,
    true
);

-- Configuración de impresora térmica estándar
INSERT INTO etiquetas_configuraciones_impresora (nombre, tipo, marca, modelo, resolucion_dpi, velocidad_impresion, ancho_papel_mm, activa, predeterminada, configuracion_driver) VALUES
(
    'Impresora Térmica Estándar',
    'termica_directa',
    'Zebra',
    'GK420d',
    203,
    4,
    104.0,
    true,
    true,
    '{
        "densidad": 8,
        "velocidad": 4,
        "modo_impresion": "directo",
        "sensor_papel": "gap",
        "calibracion_automatica": true
    }'
);

-- =====================================================
-- COMENTARIOS FINALES
-- =====================================================

COMMENT ON TABLE etiquetas_plantillas IS 'Plantillas de diseño para diferentes tipos de etiquetas de productos';
COMMENT ON TABLE etiquetas_configuraciones_impresora IS 'Configuraciones específicas de impresoras por sucursal';
COMMENT ON TABLE etiquetas_trabajos_impresion IS 'Registro de todos los trabajos de impresión realizados';
COMMENT ON TABLE etiquetas_preferencias_usuario IS 'Preferencias personalizadas de usuarios del módulo';
COMMENT ON TABLE etiquetas_logs_operaciones IS 'Log detallado de operaciones del módulo para auditoría';
COMMENT ON TABLE etiquetas_codigos_barras_cache IS 'Cache de códigos de barras generados para optimizar rendimiento';

-- Finalización del script
DO $$
BEGIN
    RAISE NOTICE '=================================================';
    RAISE NOTICE 'MÓDULO DE ETIQUETAS - ESQUEMA CREADO EXITOSAMENTE';
    RAISE NOTICE '=================================================';
    RAISE NOTICE 'Tablas creadas: 6';
    RAISE NOTICE 'Funciones creadas: 5';
    RAISE NOTICE 'Vistas creadas: 2';
    RAISE NOTICE 'Plantillas iniciales: 2';
    RAISE NOTICE '=================================================';
END $$;

