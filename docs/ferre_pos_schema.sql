-- =====================================================
-- ESQUEMA SQL OPTIMIZADO SERVIDOR CENTRAL - SISTEMA FERRE-POS
-- =====================================================
-- Versión: 2.0 - Arquitectura Centralizada Go
-- Fecha: Julio 2025
-- Autor: Manus AI
-- Descripción: Esquema optimizado para servidor central único con 4 procesos API diferenciados
--              - api_pos (máxima prioridad): Operaciones críticas POS
--              - api_sync (prioridad media): Sincronización ERP
--              - api_labels (prioridad baja): Módulo de etiquetas
--              - api_report (prioridad mínima): Reportes y análisis
-- =====================================================

-- Configuración inicial optimizada para PostgreSQL
SET timezone = 'America/Santiago';
SET default_tablespace = '';
SET default_table_access_method = heap;

-- Configuraciones específicas para alta concurrencia
SET shared_preload_libraries = 'pg_stat_statements';
SET track_activities = on;
SET track_counts = on;
SET track_io_timing = on;
SET track_functions = 'all';

-- Extensiones necesarias para arquitectura centralizada
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";
CREATE EXTENSION IF NOT EXISTS "btree_gin";
CREATE EXTENSION IF NOT EXISTS "btree_gist";
CREATE EXTENSION IF NOT EXISTS "pg_stat_statements";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- =====================================================
-- TIPOS DE DATOS PERSONALIZADOS OPTIMIZADOS
-- =====================================================

-- Tipo para roles de usuario (actualizado para incluir operador de etiquetas)
CREATE TYPE rol_usuario AS ENUM (
    'cajero', 
    'vendedor', 
    'despacho', 
    'supervisor', 
    'admin',
    'operador_etiquetas'
);

-- Tipo para estados de documentos
CREATE TYPE estado_documento AS ENUM (
    'pendiente', 
    'procesado', 
    'enviado', 
    'rechazado', 
    'anulado'
);

-- Tipo para tipos de movimiento de fidelización
CREATE TYPE tipo_movimiento_fidelizacion AS ENUM (
    'acumulacion', 
    'canje', 
    'ajuste', 
    'expiracion'
);

-- Tipo para estados de sincronización
CREATE TYPE estado_sincronizacion AS ENUM (
    'pendiente', 
    'en_proceso', 
    'completado', 
    'error'
);

-- Tipo para prioridad de procesos API
CREATE TYPE prioridad_proceso AS ENUM (
    'maxima',     -- api_pos
    'media',      -- api_sync  
    'baja',       -- api_labels
    'minima'      -- api_report
);

-- Tipo para estado de trabajos de etiquetas
CREATE TYPE estado_trabajo_etiqueta AS ENUM (
    'pendiente',
    'procesando',
    'completado',
    'error',
    'cancelado'
);

-- =====================================================
-- TABLAS PRINCIPALES OPTIMIZADAS
-- =====================================================



-- Tabla: sucursales (optimizada para acceso frecuente por api_pos)
CREATE TABLE sucursales (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    codigo TEXT UNIQUE NOT NULL,
    nombre TEXT NOT NULL,
    direccion TEXT,
    comuna TEXT,
    region TEXT,
    telefono TEXT,
    email TEXT,
    horario_apertura TIME,
    horario_cierre TIME,
    timezone TEXT DEFAULT 'America/Santiago',
    habilitada BOOLEAN DEFAULT true,
    fecha_creacion TIMESTAMP DEFAULT NOW(),
    fecha_modificacion TIMESTAMP DEFAULT NOW(),
    configuracion_dte JSONB,
    configuracion_pagos JSONB,
    -- Campos adicionales para optimización centralizada
    max_conexiones_concurrentes INTEGER DEFAULT 50,
    configuracion_cache JSONB,
    metricas_rendimiento JSONB,
    CONSTRAINT chk_horarios CHECK (horario_apertura < horario_cierre),
    CONSTRAINT chk_max_conexiones CHECK (max_conexiones_concurrentes > 0)
);

-- Tabla: usuarios (optimizada con cache de permisos para api_pos)
CREATE TABLE usuarios (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rut TEXT UNIQUE NOT NULL,
    nombre TEXT NOT NULL,
    apellido TEXT,
    email TEXT UNIQUE,
    telefono TEXT,
    rol rol_usuario NOT NULL,
    sucursal_id UUID REFERENCES sucursales(id),
    password_hash TEXT NOT NULL,
    salt TEXT NOT NULL,
    activo BOOLEAN DEFAULT true,
    ultimo_acceso TIMESTAMP,
    intentos_fallidos INTEGER DEFAULT 0,
    bloqueado_hasta TIMESTAMP,
    fecha_creacion TIMESTAMP DEFAULT NOW(),
    fecha_modificacion TIMESTAMP DEFAULT NOW(),
    configuracion_personal JSONB,
    permisos_especiales JSONB,
    -- Campos optimizados para autenticación rápida
    cache_permisos JSONB, -- Cache de permisos para api_pos
    hash_sesion_activa TEXT, -- Hash de sesión activa para validación rápida
    ultimo_terminal_id UUID, -- Último terminal utilizado
    preferencias_ui JSONB, -- Preferencias de interfaz para Node.js + Tauri
    CONSTRAINT chk_rut_formato CHECK (rut ~ '^[0-9]{7,8}-[0-9Kk]$'),
    CONSTRAINT chk_email_formato CHECK (email ~ '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$')
);

-- Tabla: categorias_productos (optimizada para búsquedas jerárquicas)
CREATE TABLE categorias_productos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    codigo TEXT UNIQUE NOT NULL,
    nombre TEXT NOT NULL,
    descripcion TEXT,
    categoria_padre_id UUID REFERENCES categorias_productos(id),
    nivel INTEGER NOT NULL DEFAULT 1,
    activa BOOLEAN DEFAULT true,
    orden_visualizacion INTEGER,
    imagen_url TEXT,
    fecha_creacion TIMESTAMP DEFAULT NOW(),
    fecha_modificacion TIMESTAMP DEFAULT NOW(),
    -- Campos optimizados para navegación rápida
    path_completo TEXT, -- Ruta completa para búsquedas rápidas
    total_productos INTEGER DEFAULT 0, -- Cache de total de productos
    configuracion_etiquetas JSONB, -- Configuración específica para módulo de etiquetas
    CONSTRAINT chk_nivel_positivo CHECK (nivel > 0)
);

-- Tabla: productos (altamente optimizada para api_pos y api_labels)
CREATE TABLE productos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    codigo_interno TEXT UNIQUE NOT NULL,
    codigo_barra TEXT UNIQUE NOT NULL,
    descripcion TEXT NOT NULL,
    descripcion_corta TEXT,
    categoria_id UUID REFERENCES categorias_productos(id),
    marca TEXT,
    modelo TEXT,
    precio_unitario NUMERIC(12,2) NOT NULL,
    precio_costo NUMERIC(12,2),
    unidad_medida TEXT NOT NULL DEFAULT 'UN',
    peso NUMERIC(8,3),
    dimensiones JSONB,
    especificaciones_tecnicas JSONB,
    activo BOOLEAN DEFAULT true,
    requiere_serie BOOLEAN DEFAULT false,
    permite_fraccionamiento BOOLEAN DEFAULT false,
    stock_minimo INTEGER DEFAULT 0,
    stock_maximo INTEGER,
    imagen_principal_url TEXT,
    imagenes_adicionales JSONB,
    fecha_creacion TIMESTAMP DEFAULT NOW(),
    fecha_modificacion TIMESTAMP DEFAULT NOW(),
    usuario_creacion UUID REFERENCES usuarios(id),
    usuario_modificacion UUID REFERENCES usuarios(id),
    -- Campos optimizados para búsquedas y etiquetas
    descripcion_busqueda TSVECTOR, -- Vector de búsqueda pre-calculado
    popularidad_score NUMERIC(5,2) DEFAULT 0, -- Score de popularidad para ordenamiento
    cache_codigo_barras_generado TEXT, -- Cache del código de barras generado
    configuracion_etiqueta JSONB, -- Configuración específica para etiquetas
    fecha_ultima_etiqueta TIMESTAMP, -- Última vez que se generó etiqueta
    total_etiquetas_generadas INTEGER DEFAULT 0, -- Contador de etiquetas generadas
    CONSTRAINT chk_precio_positivo CHECK (precio_unitario >= 0),
    CONSTRAINT chk_precio_costo_positivo CHECK (precio_costo IS NULL OR precio_costo >= 0),
    CONSTRAINT chk_stock_minimo CHECK (stock_minimo >= 0),
    CONSTRAINT chk_stock_maximo CHECK (stock_maximo IS NULL OR stock_maximo >= stock_minimo),
    CONSTRAINT chk_popularidad_score CHECK (popularidad_score >= 0 AND popularidad_score <= 100)
);

-- Tabla: codigos_barra_adicionales (optimizada para búsquedas rápidas)
CREATE TABLE codigos_barra_adicionales (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    producto_id UUID REFERENCES productos(id) ON DELETE CASCADE,
    codigo_barra TEXT NOT NULL,
    descripcion TEXT,
    activo BOOLEAN DEFAULT true,
    fecha_creacion TIMESTAMP DEFAULT NOW(),
    -- Campos optimizados
    tipo_codigo TEXT DEFAULT 'EAN13', -- Tipo de código de barras
    validado BOOLEAN DEFAULT false, -- Si el código ha sido validado
    UNIQUE(codigo_barra)
);

-- Tabla: stock_central (altamente optimizada para api_pos)
CREATE TABLE stock_central (
    producto_id UUID REFERENCES productos(id),
    sucursal_id UUID REFERENCES sucursales(id),
    cantidad INTEGER DEFAULT 0,
    cantidad_reservada INTEGER DEFAULT 0,
    cantidad_disponible INTEGER GENERATED ALWAYS AS (cantidad - cantidad_reservada) STORED,
    costo_promedio NUMERIC(12,2),
    fecha_ultima_entrada TIMESTAMP,
    fecha_ultima_salida TIMESTAMP,
    fecha_sync TIMESTAMP DEFAULT NOW(),
    -- Campos optimizados para rendimiento
    version_optimistic_lock INTEGER DEFAULT 1, -- Control de concurrencia optimista
    cache_validez TIMESTAMP DEFAULT NOW(), -- Validez del cache
    alertas_configuradas JSONB, -- Configuración de alertas de stock
    PRIMARY KEY (producto_id, sucursal_id),
    CONSTRAINT chk_cantidad_positiva CHECK (cantidad >= 0),
    CONSTRAINT chk_reservada_positiva CHECK (cantidad_reservada >= 0),
    CONSTRAINT chk_reservada_menor_cantidad CHECK (cantidad_reservada <= cantidad)
);

-- Tabla: movimientos_stock (particionada por fecha para rendimiento)
CREATE TABLE movimientos_stock (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    producto_id UUID REFERENCES productos(id),
    sucursal_id UUID REFERENCES sucursales(id),
    tipo_movimiento TEXT NOT NULL,
    cantidad INTEGER NOT NULL,
    cantidad_anterior INTEGER,
    cantidad_nueva INTEGER,
    costo_unitario NUMERIC(12,2),
    documento_referencia TEXT,
    usuario_id UUID REFERENCES usuarios(id),
    fecha TIMESTAMP DEFAULT NOW(),
    observaciones TEXT,
    datos_adicionales JSONB,
    -- Campos optimizados
    proceso_origen prioridad_proceso, -- Qué proceso API generó el movimiento
    batch_id UUID, -- ID de lote para movimientos masivos
    CONSTRAINT chk_tipo_movimiento CHECK (tipo_movimiento IN (
        'entrada', 'salida', 'ajuste', 'transferencia_entrada', 
        'transferencia_salida', 'venta', 'devolucion'
    ))
) PARTITION BY RANGE (fecha);

-- Crear particiones para movimientos_stock (últimos 12 meses + futuro)
CREATE TABLE movimientos_stock_2024_q4 PARTITION OF movimientos_stock
    FOR VALUES FROM ('2024-10-01') TO ('2025-01-01');
CREATE TABLE movimientos_stock_2025_q1 PARTITION OF movimientos_stock
    FOR VALUES FROM ('2025-01-01') TO ('2025-04-01');
CREATE TABLE movimientos_stock_2025_q2 PARTITION OF movimientos_stock
    FOR VALUES FROM ('2025-04-01') TO ('2025-07-01');
CREATE TABLE movimientos_stock_2025_q3 PARTITION OF movimientos_stock
    FOR VALUES FROM ('2025-07-01') TO ('2025-10-01');
CREATE TABLE movimientos_stock_2025_q4 PARTITION OF movimientos_stock
    FOR VALUES FROM ('2025-10-01') TO ('2026-01-01');
CREATE TABLE movimientos_stock_futuro PARTITION OF movimientos_stock
    FOR VALUES FROM ('2026-01-01') TO (MAXVALUE);

-- Tabla: terminales (optimizada para conexiones frecuentes)
CREATE TABLE terminales (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    codigo TEXT UNIQUE NOT NULL,
    nombre_terminal TEXT NOT NULL,
    tipo_terminal TEXT NOT NULL,
    sucursal_id UUID REFERENCES sucursales(id),
    direccion_ip INET,
    direccion_mac MACADDR,
    activo BOOLEAN DEFAULT true,
    ultima_conexion TIMESTAMP,
    version_software TEXT,
    configuracion JSONB,
    fecha_instalacion TIMESTAMP DEFAULT NOW(),
    fecha_modificacion TIMESTAMP DEFAULT NOW(),
    -- Campos optimizados para monitoreo
    heartbeat_interval INTEGER DEFAULT 30, -- Intervalo de heartbeat en segundos
    estado_conexion TEXT DEFAULT 'desconectado',
    metricas_rendimiento JSONB, -- Métricas de rendimiento del terminal
    configuracion_cache JSONB, -- Configuración de cache local
    CONSTRAINT chk_tipo_terminal CHECK (tipo_terminal IN (
        'caja', 'tienda', 'despacho', 'autoatencion', 'etiquetas'
    )),
    CONSTRAINT chk_estado_conexion CHECK (estado_conexion IN (
        'conectado', 'desconectado', 'error', 'mantenimiento'
    ))
);

-- Tabla: ventas (altamente optimizada para api_pos)
CREATE TABLE ventas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    numero_venta BIGSERIAL UNIQUE,
    sucursal_id UUID REFERENCES sucursales(id),
    terminal_id UUID REFERENCES terminales(id),
    cajero_id UUID REFERENCES usuarios(id),
    vendedor_id UUID REFERENCES usuarios(id),
    cliente_rut TEXT,
    cliente_nombre TEXT,
    nota_venta_id UUID,
    tipo_documento TEXT NOT NULL,
    subtotal NUMERIC(12,2) NOT NULL,
    descuento_total NUMERIC(12,2) DEFAULT 0,
    impuesto_total NUMERIC(12,2) DEFAULT 0,
    total NUMERIC(12,2) NOT NULL,
    estado TEXT DEFAULT 'finalizada',
    fecha TIMESTAMP DEFAULT NOW(),
    fecha_anulacion TIMESTAMP,
    motivo_anulacion TEXT,
    usuario_anulacion UUID REFERENCES usuarios(id),
    dte_id UUID,
    dte_emitido BOOLEAN DEFAULT false,
    sincronizada BOOLEAN DEFAULT false,
    fecha_sincronizacion TIMESTAMP,
    datos_adicionales JSONB,
    -- Campos optimizados para rendimiento
    hash_integridad TEXT, -- Hash para verificación de integridad
    tiempo_procesamiento_ms INTEGER, -- Tiempo de procesamiento en milisegundos
    proceso_origen prioridad_proceso DEFAULT 'maxima', -- Siempre api_pos
    cache_totales JSONB, -- Cache de cálculos de totales
    CONSTRAINT chk_tipo_documento CHECK (tipo_documento IN (
        'boleta', 'factura', 'guia', 'nota_venta'
    )),
    CONSTRAINT chk_estado_venta CHECK (estado IN (
        'pendiente', 'finalizada', 'anulada'
    )),
    CONSTRAINT chk_totales_positivos CHECK (
        subtotal >= 0 AND descuento_total >= 0 AND 
        impuesto_total >= 0 AND total >= 0
    ),
    CONSTRAINT chk_total_coherente CHECK (
        total = subtotal - descuento_total + impuesto_total
    )
) PARTITION BY RANGE (fecha);

-- Crear particiones para ventas (por mes para mejor rendimiento)
CREATE TABLE ventas_2025_01 PARTITION OF ventas
    FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');
CREATE TABLE ventas_2025_02 PARTITION OF ventas
    FOR VALUES FROM ('2025-02-01') TO ('2025-03-01');
CREATE TABLE ventas_2025_03 PARTITION OF ventas
    FOR VALUES FROM ('2025-03-01') TO ('2025-04-01');
CREATE TABLE ventas_2025_04 PARTITION OF ventas
    FOR VALUES FROM ('2025-04-01') TO ('2025-05-01');
CREATE TABLE ventas_2025_05 PARTITION OF ventas
    FOR VALUES FROM ('2025-05-01') TO ('2025-06-01');
CREATE TABLE ventas_2025_06 PARTITION OF ventas
    FOR VALUES FROM ('2025-06-01') TO ('2025-07-01');
CREATE TABLE ventas_2025_07 PARTITION OF ventas
    FOR VALUES FROM ('2025-07-01') TO ('2025-08-01');
CREATE TABLE ventas_2025_08 PARTITION OF ventas
    FOR VALUES FROM ('2025-08-01') TO ('2025-09-01');
CREATE TABLE ventas_2025_09 PARTITION OF ventas
    FOR VALUES FROM ('2025-09-01') TO ('2025-10-01');
CREATE TABLE ventas_2025_10 PARTITION OF ventas
    FOR VALUES FROM ('2025-10-01') TO ('2025-11-01');
CREATE TABLE ventas_2025_11 PARTITION OF ventas
    FOR VALUES FROM ('2025-11-01') TO ('2025-12-01');
CREATE TABLE ventas_2025_12 PARTITION OF ventas
    FOR VALUES FROM ('2025-12-01') TO ('2026-01-01');
CREATE TABLE ventas_futuro PARTITION OF ventas
    FOR VALUES FROM ('2026-01-01') TO (MAXVALUE);

-- Tabla: detalle_ventas (optimizada para consultas frecuentes)
CREATE TABLE detalle_ventas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    venta_id UUID NOT NULL, -- Referencia sin FK para mejor rendimiento en particiones
    producto_id UUID REFERENCES productos(id),
    cantidad NUMERIC(10,3) NOT NULL,
    precio_unitario NUMERIC(12,2) NOT NULL,
    descuento_unitario NUMERIC(12,2) DEFAULT 0,
    precio_final NUMERIC(12,2) NOT NULL,
    total_item NUMERIC(12,2) NOT NULL,
    numero_serie TEXT,
    lote TEXT,
    fecha_vencimiento DATE,
    datos_adicionales JSONB,
    -- Campos optimizados
    margen_unitario NUMERIC(12,2), -- Margen calculado
    categoria_producto_id UUID, -- Desnormalizado para reportes rápidos
    CONSTRAINT chk_cantidad_positiva CHECK (cantidad > 0),
    CONSTRAINT chk_precios_positivos CHECK (
        precio_unitario >= 0 AND descuento_unitario >= 0 AND 
        precio_final >= 0 AND total_item >= 0
    ),
    CONSTRAINT chk_precio_final_coherente CHECK (
        precio_final = precio_unitario - descuento_unitario
    ),
    CONSTRAINT chk_total_item_coherente CHECK (
        total_item = precio_final * cantidad
    )
) PARTITION BY RANGE (venta_id);

-- Crear particiones para detalle_ventas basadas en hash del venta_id
CREATE TABLE detalle_ventas_p1 PARTITION OF detalle_ventas
    FOR VALUES FROM (MINVALUE) TO ('50000000-0000-0000-0000-000000000000');
CREATE TABLE detalle_ventas_p2 PARTITION OF detalle_ventas
    FOR VALUES FROM ('50000000-0000-0000-0000-000000000000') TO ('a0000000-0000-0000-0000-000000000000');
CREATE TABLE detalle_ventas_p3 PARTITION OF detalle_ventas
    FOR VALUES FROM ('a0000000-0000-0000-0000-000000000000') TO (MAXVALUE);

-- Tabla: medios_pago_venta (optimizada para conciliación)
CREATE TABLE medios_pago_venta (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    venta_id UUID NOT NULL, -- Referencia sin FK para particiones
    medio_pago TEXT NOT NULL,
    monto NUMERIC(12,2) NOT NULL,
    referencia_transaccion TEXT,
    codigo_autorizacion TEXT,
    datos_transaccion JSONB,
    fecha_procesamiento TIMESTAMP DEFAULT NOW(),
    -- Campos optimizados para conciliación
    estado_conciliacion TEXT DEFAULT 'pendiente',
    fecha_conciliacion TIMESTAMP,
    lote_conciliacion TEXT,
    comision NUMERIC(12,2), -- Comisión del medio de pago
    CONSTRAINT chk_medio_pago CHECK (medio_pago IN (
        'efectivo', 'tarjeta_debito', 'tarjeta_credito', 
        'transferencia', 'cheque', 'puntos_fidelizacion', 'otro'
    )),
    CONSTRAINT chk_monto_positivo CHECK (monto > 0),
    CONSTRAINT chk_estado_conciliacion CHECK (estado_conciliacion IN (
        'pendiente', 'conciliado', 'diferencia', 'error'
    ))
);

-- =====================================================
-- TABLAS ESPECÍFICAS PARA MÓDULO DE ETIQUETAS (api_labels)
-- =====================================================

-- Tabla: etiquetas_plantillas
-- Descripción: Plantillas de diseño para etiquetas
CREATE TABLE etiquetas_plantillas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    codigo TEXT UNIQUE NOT NULL,
    nombre TEXT NOT NULL,
    descripcion TEXT,
    categoria_producto_id UUID REFERENCES categorias_productos(id),
    tipo_etiqueta TEXT NOT NULL DEFAULT 'precio',
    ancho_mm NUMERIC(5,2) NOT NULL,
    alto_mm NUMERIC(5,2) NOT NULL,
    orientacion TEXT DEFAULT 'horizontal',
    configuracion_diseno JSONB NOT NULL,
    configuracion_codigo_barras JSONB,
    activa BOOLEAN DEFAULT true,
    predeterminada BOOLEAN DEFAULT false,
    fecha_creacion TIMESTAMP DEFAULT NOW(),
    fecha_modificacion TIMESTAMP DEFAULT NOW(),
    usuario_creacion UUID REFERENCES usuarios(id),
    usuario_modificacion UUID REFERENCES usuarios(id),
    -- Campos optimizados
    total_usos INTEGER DEFAULT 0, -- Contador de usos
    tiempo_renderizado_promedio_ms INTEGER, -- Tiempo promedio de renderizado
    cache_preview TEXT, -- Cache de vista previa en base64
    CONSTRAINT chk_tipo_etiqueta CHECK (tipo_etiqueta IN (
        'precio', 'inventario', 'promocion', 'codigo_barras', 'personalizada'
    )),
    CONSTRAINT chk_orientacion CHECK (orientacion IN ('horizontal', 'vertical')),
    CONSTRAINT chk_dimensiones_positivas CHECK (ancho_mm > 0 AND alto_mm > 0)
);

-- Tabla: etiquetas_trabajos_impresion
-- Descripción: Registro de trabajos de impresión de etiquetas
CREATE TABLE etiquetas_trabajos_impresion (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    numero_trabajo BIGSERIAL UNIQUE,
    usuario_id UUID REFERENCES usuarios(id),
    sucursal_id UUID REFERENCES sucursales(id),
    terminal_id UUID REFERENCES terminales(id),
    plantilla_id UUID REFERENCES etiquetas_plantillas(id),
    tipo_trabajo TEXT NOT NULL,
    estado estado_trabajo_etiqueta DEFAULT 'pendiente',
    total_etiquetas INTEGER NOT NULL,
    etiquetas_procesadas INTEGER DEFAULT 0,
    etiquetas_exitosas INTEGER DEFAULT 0,
    etiquetas_error INTEGER DEFAULT 0,
    fecha_solicitud TIMESTAMP DEFAULT NOW(),
    fecha_inicio_procesamiento TIMESTAMP,
    fecha_fin_procesamiento TIMESTAMP,
    tiempo_procesamiento_ms INTEGER,
    configuracion_impresora JSONB,
    parametros_trabajo JSONB,
    errores_detalle JSONB,
    archivo_generado_path TEXT,
    hash_archivo TEXT,
    -- Campos optimizados para monitoreo
    prioridad INTEGER DEFAULT 5, -- 1=alta, 5=normal, 10=baja
    proceso_api prioridad_proceso DEFAULT 'baja', -- Siempre api_labels
    recursos_utilizados JSONB, -- CPU, memoria utilizados
    CONSTRAINT chk_tipo_trabajo CHECK (tipo_trabajo IN (
        'individual', 'masivo', 'categoria', 'stock_bajo', 'personalizado'
    )),
    CONSTRAINT chk_total_etiquetas_positivo CHECK (total_etiquetas > 0),
    CONSTRAINT chk_contadores_coherentes CHECK (
        etiquetas_procesadas >= 0 AND
        etiquetas_exitosas >= 0 AND
        etiquetas_error >= 0 AND
        etiquetas_procesadas = etiquetas_exitosas + etiquetas_error
    ),
    CONSTRAINT chk_prioridad_valida CHECK (prioridad BETWEEN 1 AND 10)
) PARTITION BY RANGE (fecha_solicitud);

-- Crear particiones para trabajos de etiquetas (por mes)
CREATE TABLE etiquetas_trabajos_2025_07 PARTITION OF etiquetas_trabajos_impresion
    FOR VALUES FROM ('2025-07-01') TO ('2025-08-01');
CREATE TABLE etiquetas_trabajos_2025_08 PARTITION OF etiquetas_trabajos_impresion
    FOR VALUES FROM ('2025-08-01') TO ('2025-09-01');
CREATE TABLE etiquetas_trabajos_2025_09 PARTITION OF etiquetas_trabajos_impresion
    FOR VALUES FROM ('2025-09-01') TO ('2025-10-01');
CREATE TABLE etiquetas_trabajos_2025_10 PARTITION OF etiquetas_trabajos_impresion
    FOR VALUES FROM ('2025-10-01') TO ('2025-11-01');
CREATE TABLE etiquetas_trabajos_2025_11 PARTITION OF etiquetas_trabajos_impresion
    FOR VALUES FROM ('2025-11-01') TO ('2025-12-01');
CREATE TABLE etiquetas_trabajos_2025_12 PARTITION OF etiquetas_trabajos_impresion
    FOR VALUES FROM ('2025-12-01') TO ('2026-01-01');
CREATE TABLE etiquetas_trabajos_futuro PARTITION OF etiquetas_trabajos_impresion
    FOR VALUES FROM ('2026-01-01') TO (MAXVALUE);

-- Tabla: etiquetas_detalle_trabajo
-- Descripción: Detalle de productos en cada trabajo de impresión
CREATE TABLE etiquetas_detalle_trabajo (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trabajo_id UUID NOT NULL, -- Referencia sin FK para particiones
    producto_id UUID REFERENCES productos(id),
    cantidad_solicitada INTEGER NOT NULL,
    cantidad_procesada INTEGER DEFAULT 0,
    estado_item TEXT DEFAULT 'pendiente',
    parametros_especificos JSONB,
    codigo_barras_generado TEXT,
    tiempo_procesamiento_ms INTEGER,
    error_detalle TEXT,
    archivo_etiqueta_path TEXT,
    -- Campos optimizados
    orden_procesamiento INTEGER, -- Orden en el que se procesa
    cache_datos_producto JSONB, -- Cache de datos del producto al momento del trabajo
    CONSTRAINT chk_cantidad_solicitada_positiva CHECK (cantidad_solicitada > 0),
    CONSTRAINT chk_cantidad_procesada_valida CHECK (
        cantidad_procesada >= 0 AND cantidad_procesada <= cantidad_solicitada
    ),
    CONSTRAINT chk_estado_item CHECK (estado_item IN (
        'pendiente', 'procesando', 'completado', 'error', 'omitido'
    ))
);

-- Tabla: etiquetas_configuraciones_impresora
-- Descripción: Configuraciones específicas de impresoras de etiquetas
CREATE TABLE etiquetas_configuraciones_impresora (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    codigo TEXT UNIQUE NOT NULL,
    nombre TEXT NOT NULL,
    marca TEXT NOT NULL,
    modelo TEXT NOT NULL,
    tipo_conexion TEXT NOT NULL,
    parametros_conexion JSONB,
    configuracion_driver JSONB,
    resoluciones_soportadas JSONB,
    tamaños_papel_soportados JSONB,
    activa BOOLEAN DEFAULT true,
    sucursal_id UUID REFERENCES sucursales(id),
    terminal_id UUID REFERENCES terminales(id),
    fecha_instalacion TIMESTAMP DEFAULT NOW(),
    fecha_modificacion TIMESTAMP DEFAULT NOW(),
    -- Campos optimizados para monitoreo
    estado_impresora TEXT DEFAULT 'disponible',
    ultimo_trabajo_id UUID,
    fecha_ultimo_trabajo TIMESTAMP,
    total_trabajos_procesados INTEGER DEFAULT 0,
    total_etiquetas_impresas INTEGER DEFAULT 0,
    tiempo_promedio_etiqueta_ms INTEGER,
    errores_recientes JSONB,
    mantenimiento_programado TIMESTAMP,
    CONSTRAINT chk_tipo_conexion CHECK (tipo_conexion IN (
        'usb', 'red', 'bluetooth', 'serie'
    )),
    CONSTRAINT chk_estado_impresora CHECK (estado_impresora IN (
        'disponible', 'ocupada', 'error', 'mantenimiento', 'desconectada'
    ))
);

-- Tabla: etiquetas_cache_codigos_barras
-- Descripción: Cache de códigos de barras generados para optimización
CREATE TABLE etiquetas_cache_codigos_barras (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    codigo_producto TEXT NOT NULL,
    tipo_codigo TEXT NOT NULL DEFAULT 'CODE39',
    codigo_barras_generado TEXT NOT NULL,
    imagen_base64 TEXT,
    parametros_generacion JSONB,
    fecha_generacion TIMESTAMP DEFAULT NOW(),
    fecha_ultimo_uso TIMESTAMP DEFAULT NOW(),
    total_usos INTEGER DEFAULT 1,
    valido_hasta TIMESTAMP DEFAULT (NOW() + INTERVAL '30 days'),
    -- Campos optimizados
    tamaño_imagen_bytes INTEGER,
    tiempo_generacion_ms INTEGER,
    hash_parametros TEXT, -- Hash de parámetros para búsqueda rápida
    CONSTRAINT chk_tipo_codigo CHECK (tipo_codigo IN (
        'CODE39', 'CODE128', 'EAN13', 'EAN8', 'UPC'
    )),
    UNIQUE(codigo_producto, tipo_codigo, hash_parametros)
);

-- Tabla: etiquetas_metricas_rendimiento
-- Descripción: Métricas de rendimiento del módulo de etiquetas
CREATE TABLE etiquetas_metricas_rendimiento (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fecha_hora TIMESTAMP DEFAULT NOW(),
    sucursal_id UUID REFERENCES sucursales(id),
    terminal_id UUID REFERENCES terminales(id),
    tipo_metrica TEXT NOT NULL,
    valor_numerico NUMERIC(12,4),
    valor_texto TEXT,
    metadatos JSONB,
    -- Campos para agregación
    periodo_agregacion TEXT, -- 'minuto', 'hora', 'dia'
    fecha_periodo DATE,
    hora_periodo INTEGER,
    CONSTRAINT chk_tipo_metrica CHECK (tipo_metrica IN (
        'tiempo_busqueda_producto', 'tiempo_generacion_codigo_barras',
        'tiempo_renderizado_etiqueta', 'tiempo_impresion',
        'trabajos_por_hora', 'etiquetas_por_minuto',
        'errores_por_hora', 'uso_cpu', 'uso_memoria'
    )),
    CONSTRAINT chk_periodo_agregacion CHECK (periodo_agregacion IN (
        'minuto', 'hora', 'dia'
    ))
) PARTITION BY RANGE (fecha_hora);

-- Crear particiones para métricas (por día para análisis detallado)
CREATE TABLE etiquetas_metricas_2025_07 PARTITION OF etiquetas_metricas_rendimiento
    FOR VALUES FROM ('2025-07-01') TO ('2025-08-01');
CREATE TABLE etiquetas_metricas_2025_08 PARTITION OF etiquetas_metricas_rendimiento
    FOR VALUES FROM ('2025-08-01') TO ('2025-09-01');
CREATE TABLE etiquetas_metricas_2025_09 PARTITION OF etiquetas_metricas_rendimiento
    FOR VALUES FROM ('2025-09-01') TO ('2025-10-01');
CREATE TABLE etiquetas_metricas_2025_10 PARTITION OF etiquetas_metricas_rendimiento
    FOR VALUES FROM ('2025-10-01') TO ('2025-11-01');
CREATE TABLE etiquetas_metricas_2025_11 PARTITION OF etiquetas_metricas_rendimiento
    FOR VALUES FROM ('2025-11-01') TO ('2025-12-01');
CREATE TABLE etiquetas_metricas_2025_12 PARTITION OF etiquetas_metricas_rendimiento
    FOR VALUES FROM ('2025-12-01') TO ('2026-01-01');
CREATE TABLE etiquetas_metricas_futuro PARTITION OF etiquetas_metricas_rendimiento
    FOR VALUES FROM ('2026-01-01') TO (MAXVALUE);


-- =====================================================
-- TABLAS RESTANTES OPTIMIZADAS
-- =====================================================

-- Tabla: notas_venta (optimizada para flujo POS Tienda -> Caja)
CREATE TABLE notas_venta (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    numero_nota BIGSERIAL UNIQUE,
    sucursal_id UUID REFERENCES sucursales(id),
    terminal_id UUID REFERENCES terminales(id),
    vendedor_id UUID REFERENCES usuarios(id),
    cliente_rut TEXT,
    cliente_nombre TEXT,
    subtotal NUMERIC(12,2) NOT NULL,
    descuento_total NUMERIC(12,2) DEFAULT 0,
    total NUMERIC(12,2) NOT NULL,
    estado TEXT DEFAULT 'pendiente',
    fecha TIMESTAMP DEFAULT NOW(),
    fecha_vencimiento TIMESTAMP,
    venta_id UUID, -- Referencia a venta cuando se procesa
    fecha_pago TIMESTAMP,
    sincronizada BOOLEAN DEFAULT false,
    observaciones TEXT,
    -- Campos optimizados para flujo de trabajo
    qr_code TEXT, -- Código QR para identificación rápida
    hash_validacion TEXT, -- Hash para validación de integridad
    tiempo_vigencia_horas INTEGER DEFAULT 24,
    prioridad_atencion INTEGER DEFAULT 5,
    CONSTRAINT chk_estado_nota CHECK (estado IN (
        'pendiente', 'pagada', 'cancelada', 'vencida'
    )),
    CONSTRAINT chk_totales_nota_positivos CHECK (
        subtotal >= 0 AND descuento_total >= 0 AND total >= 0
    ),
    CONSTRAINT chk_prioridad_atencion CHECK (prioridad_atencion BETWEEN 1 AND 10)
);

-- Tabla: detalle_notas_venta (optimizada)
CREATE TABLE detalle_notas_venta (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    nota_venta_id UUID REFERENCES notas_venta(id) ON DELETE CASCADE,
    producto_id UUID REFERENCES productos(id),
    cantidad NUMERIC(10,3) NOT NULL,
    precio_unitario NUMERIC(12,2) NOT NULL,
    descuento_unitario NUMERIC(12,2) DEFAULT 0,
    total_item NUMERIC(12,2) NOT NULL,
    observaciones TEXT,
    -- Campos optimizados
    disponibilidad_verificada BOOLEAN DEFAULT false,
    fecha_verificacion_stock TIMESTAMP,
    CONSTRAINT chk_cantidad_nota_positiva CHECK (cantidad > 0),
    CONSTRAINT chk_precios_nota_positivos CHECK (
        precio_unitario >= 0 AND descuento_unitario >= 0 AND total_item >= 0
    )
);

-- Tabla: proveedores_dte (optimizada para integración)
CREATE TABLE proveedores_dte (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    nombre TEXT NOT NULL,
    codigo TEXT UNIQUE NOT NULL,
    api_url TEXT NOT NULL,
    api_version TEXT,
    certificado_digital TEXT,
    activo BOOLEAN DEFAULT true,
    configuracion JSONB,
    fecha_creacion TIMESTAMP DEFAULT NOW(),
    fecha_modificacion TIMESTAMP DEFAULT NOW(),
    -- Campos optimizados para monitoreo
    tiempo_respuesta_promedio_ms INTEGER,
    disponibilidad_porcentaje NUMERIC(5,2) DEFAULT 100.0,
    ultimo_error TEXT,
    fecha_ultimo_error TIMESTAMP,
    total_documentos_procesados INTEGER DEFAULT 0,
    limite_documentos_por_hora INTEGER DEFAULT 1000
);

-- Tabla: configuracion_dte_sucursal (optimizada)
CREATE TABLE configuracion_dte_sucursal (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sucursal_id UUID REFERENCES sucursales(id),
    proveedor_dte_id UUID REFERENCES proveedores_dte(id),
    rut_empresa TEXT NOT NULL,
    razon_social TEXT NOT NULL,
    giro TEXT,
    direccion_fiscal TEXT,
    comuna_fiscal TEXT,
    ciudad_fiscal TEXT,
    codigo_sii TEXT,
    resolucion_sii TEXT,
    fecha_resolucion DATE,
    ambiente TEXT DEFAULT 'produccion',
    activa BOOLEAN DEFAULT true,
    configuracion_especifica JSONB,
    -- Campos optimizados
    cache_configuracion JSONB, -- Cache para acceso rápido
    fecha_ultimo_uso TIMESTAMP,
    CONSTRAINT chk_ambiente CHECK (ambiente IN ('certificacion', 'produccion'))
);

-- Tabla: documentos_dte (altamente optimizada para api_pos)
CREATE TABLE documentos_dte (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sucursal_id UUID REFERENCES sucursales(id),
    proveedor_dte_id UUID REFERENCES proveedores_dte(id),
    venta_id UUID, -- Referencia sin FK para particiones
    tipo_documento TEXT NOT NULL,
    folio BIGINT NOT NULL,
    rut_receptor TEXT,
    razon_social_receptor TEXT,
    fecha_emision TIMESTAMP DEFAULT NOW(),
    monto_neto NUMERIC(12,2),
    monto_iva NUMERIC(12,2),
    monto_total NUMERIC(12,2),
    estado estado_documento DEFAULT 'pendiente',
    xml_documento TEXT,
    pdf_documento BYTEA,
    track_id TEXT,
    codigo_sii TEXT,
    mensaje_sii TEXT,
    fecha_recepcion_sii TIMESTAMP,
    fecha_aceptacion_sii TIMESTAMP,
    intentos_envio INTEGER DEFAULT 0,
    ultimo_error TEXT,
    fecha_ultimo_intento TIMESTAMP,
    datos_adicionales JSONB,
    -- Campos optimizados para rendimiento
    hash_documento TEXT, -- Hash del XML para verificación
    tamaño_xml_bytes INTEGER,
    tamaño_pdf_bytes INTEGER,
    tiempo_generacion_ms INTEGER,
    tiempo_envio_ms INTEGER,
    prioridad_procesamiento INTEGER DEFAULT 5,
    CONSTRAINT chk_tipo_dte CHECK (tipo_documento IN (
        'boleta_electronica', 'factura_electronica', 'guia_despacho_electronica',
        'nota_credito_electronica', 'nota_debito_electronica'
    )),
    CONSTRAINT chk_folio_positivo CHECK (folio > 0),
    CONSTRAINT chk_montos_positivos CHECK (
        monto_neto >= 0 AND monto_iva >= 0 AND monto_total >= 0
    ),
    CONSTRAINT chk_prioridad_procesamiento CHECK (prioridad_procesamiento BETWEEN 1 AND 10),
    UNIQUE(sucursal_id, tipo_documento, folio)
) PARTITION BY RANGE (fecha_emision);

-- Crear particiones para documentos DTE (por mes)
CREATE TABLE documentos_dte_2025_07 PARTITION OF documentos_dte
    FOR VALUES FROM ('2025-07-01') TO ('2025-08-01');
CREATE TABLE documentos_dte_2025_08 PARTITION OF documentos_dte
    FOR VALUES FROM ('2025-08-01') TO ('2025-09-01');
CREATE TABLE documentos_dte_2025_09 PARTITION OF documentos_dte
    FOR VALUES FROM ('2025-09-01') TO ('2025-10-01');
CREATE TABLE documentos_dte_2025_10 PARTITION OF documentos_dte
    FOR VALUES FROM ('2025-10-01') TO ('2025-11-01');
CREATE TABLE documentos_dte_2025_11 PARTITION OF documentos_dte
    FOR VALUES FROM ('2025-11-01') TO ('2025-12-01');
CREATE TABLE documentos_dte_2025_12 PARTITION OF documentos_dte
    FOR VALUES FROM ('2025-12-01') TO ('2026-01-01');
CREATE TABLE documentos_dte_futuro PARTITION OF documentos_dte
    FOR VALUES FROM ('2026-01-01') TO (MAXVALUE);

-- Tabla: folios_dte (optimizada para generación rápida)
CREATE TABLE folios_dte (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sucursal_id UUID REFERENCES sucursales(id),
    tipo_documento TEXT NOT NULL,
    folio_desde BIGINT NOT NULL,
    folio_hasta BIGINT NOT NULL,
    folio_actual BIGINT NOT NULL,
    fecha_asignacion TIMESTAMP DEFAULT NOW(),
    fecha_vencimiento TIMESTAMP,
    activo BOOLEAN DEFAULT true,
    -- Campos optimizados para concurrencia
    version_lock INTEGER DEFAULT 1, -- Control de concurrencia optimista
    folios_reservados INTEGER DEFAULT 0, -- Folios reservados para transacciones en curso
    alerta_agotamiento_porcentaje NUMERIC(5,2) DEFAULT 90.0,
    CONSTRAINT chk_folios_coherentes CHECK (
        folio_desde <= folio_actual AND folio_actual <= folio_hasta
    ),
    CONSTRAINT chk_folios_reservados CHECK (
        folios_reservados >= 0 AND folios_reservados <= (folio_hasta - folio_actual)
    ),
    UNIQUE(sucursal_id, tipo_documento, folio_desde)
);

-- Tabla: fidelizacion_clientes (optimizada para consultas frecuentes)
CREATE TABLE fidelizacion_clientes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rut TEXT UNIQUE NOT NULL,
    nombre TEXT NOT NULL,
    apellido TEXT,
    email TEXT,
    telefono TEXT,
    fecha_nacimiento DATE,
    direccion TEXT,
    comuna TEXT,
    region TEXT,
    puntos_actuales INTEGER DEFAULT 0,
    puntos_acumulados_total INTEGER DEFAULT 0,
    nivel_fidelizacion TEXT DEFAULT 'bronce',
    fecha_ultima_compra TIMESTAMP,
    fecha_ultima_actividad TIMESTAMP,
    activo BOOLEAN DEFAULT true,
    acepta_marketing BOOLEAN DEFAULT false,
    fecha_creacion TIMESTAMP DEFAULT NOW(),
    fecha_modificacion TIMESTAMP DEFAULT NOW(),
    datos_adicionales JSONB,
    -- Campos optimizados para rendimiento
    hash_busqueda TEXT, -- Hash para búsquedas rápidas por nombre/apellido
    qr_fidelizacion TEXT, -- Código QR único para identificación rápida
    cache_estadisticas JSONB, -- Cache de estadísticas calculadas
    fecha_proximo_vencimiento_puntos TIMESTAMP,
    puntos_por_vencer INTEGER DEFAULT 0,
    CONSTRAINT chk_rut_cliente_formato CHECK (rut ~ '^[0-9]{7,8}-[0-9Kk]$'),
    CONSTRAINT chk_puntos_positivos CHECK (
        puntos_actuales >= 0 AND puntos_acumulados_total >= 0
    ),
    CONSTRAINT chk_nivel_fidelizacion CHECK (nivel_fidelizacion IN (
        'bronce', 'plata', 'oro', 'platino'
    ))
);

-- Tabla: movimientos_fidelizacion (particionada para rendimiento)
CREATE TABLE movimientos_fidelizacion (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cliente_id UUID REFERENCES fidelizacion_clientes(id),
    sucursal_id UUID REFERENCES sucursales(id),
    venta_id UUID, -- Referencia sin FK para particiones
    tipo tipo_movimiento_fidelizacion NOT NULL,
    puntos INTEGER NOT NULL,
    puntos_anteriores INTEGER,
    puntos_nuevos INTEGER,
    multiplicador NUMERIC(4,2) DEFAULT 1.0,
    detalle TEXT,
    fecha TIMESTAMP DEFAULT NOW(),
    fecha_vencimiento TIMESTAMP,
    usuario_id UUID REFERENCES usuarios(id),
    datos_adicionales JSONB,
    -- Campos optimizados
    regla_aplicada_id UUID, -- Referencia a regla que generó el movimiento
    lote_procesamiento UUID, -- Para movimientos masivos
    CONSTRAINT chk_puntos_movimiento CHECK (puntos != 0)
) PARTITION BY RANGE (fecha);

-- Crear particiones para movimientos de fidelización (por trimestre)
CREATE TABLE movimientos_fidelizacion_2025_q3 PARTITION OF movimientos_fidelizacion
    FOR VALUES FROM ('2025-07-01') TO ('2025-10-01');
CREATE TABLE movimientos_fidelizacion_2025_q4 PARTITION OF movimientos_fidelizacion
    FOR VALUES FROM ('2025-10-01') TO ('2026-01-01');
CREATE TABLE movimientos_fidelizacion_2026_q1 PARTITION OF movimientos_fidelizacion
    FOR VALUES FROM ('2026-01-01') TO ('2026-04-01');
CREATE TABLE movimientos_fidelizacion_futuro PARTITION OF movimientos_fidelizacion
    FOR VALUES FROM ('2026-04-01') TO (MAXVALUE);

-- Tabla: reglas_fidelizacion (optimizada para evaluación rápida)
CREATE TABLE reglas_fidelizacion (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    nombre TEXT NOT NULL,
    descripcion TEXT,
    tipo_regla TEXT NOT NULL,
    condiciones JSONB NOT NULL,
    acciones JSONB NOT NULL,
    activa BOOLEAN DEFAULT true,
    fecha_inicio TIMESTAMP,
    fecha_fin TIMESTAMP,
    prioridad INTEGER DEFAULT 0,
    fecha_creacion TIMESTAMP DEFAULT NOW(),
    fecha_modificacion TIMESTAMP DEFAULT NOW(),
    -- Campos optimizados para evaluación
    condiciones_compiladas JSONB, -- Condiciones pre-compiladas para evaluación rápida
    cache_evaluacion JSONB, -- Cache de evaluaciones recientes
    total_aplicaciones INTEGER DEFAULT 0,
    total_puntos_otorgados INTEGER DEFAULT 0,
    CONSTRAINT chk_tipo_regla CHECK (tipo_regla IN (
        'acumulacion', 'canje', 'promocion', 'nivel'
    ))
);

-- Tabla: notas_credito (optimizada para autorización)
CREATE TABLE notas_credito (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    numero_nota BIGSERIAL UNIQUE,
    sucursal_id UUID REFERENCES sucursales(id),
    documento_origen_id UUID REFERENCES documentos_dte(id),
    venta_origen_id UUID, -- Referencia sin FK para particiones
    supervisor_id UUID REFERENCES usuarios(id),
    cajero_id UUID REFERENCES usuarios(id),
    motivo TEXT NOT NULL,
    subtotal NUMERIC(12,2) NOT NULL,
    descuento_total NUMERIC(12,2) DEFAULT 0,
    impuesto_total NUMERIC(12,2) DEFAULT 0,
    total NUMERIC(12,2) NOT NULL,
    estado estado_documento DEFAULT 'pendiente',
    fecha_solicitud TIMESTAMP DEFAULT NOW(),
    fecha_autorizacion TIMESTAMP,
    fecha_emision TIMESTAMP,
    dte_id UUID,
    observaciones TEXT,
    datos_adicionales JSONB,
    -- Campos optimizados para flujo de autorización
    codigo_autorizacion TEXT, -- Código único de autorización
    tiempo_autorizacion_ms INTEGER,
    requiere_autorizacion_adicional BOOLEAN DEFAULT false,
    CONSTRAINT chk_totales_nc_positivos CHECK (
        subtotal >= 0 AND descuento_total >= 0 AND 
        impuesto_total >= 0 AND total >= 0
    )
);

-- Tabla: detalle_notas_credito (optimizada)
CREATE TABLE detalle_notas_credito (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    nota_credito_id UUID REFERENCES notas_credito(id) ON DELETE CASCADE,
    producto_id UUID REFERENCES productos(id),
    cantidad NUMERIC(10,3) NOT NULL,
    precio_unitario NUMERIC(12,2) NOT NULL,
    total_item NUMERIC(12,2) NOT NULL,
    motivo_item TEXT,
    -- Campos optimizados
    afecta_stock BOOLEAN DEFAULT true,
    stock_restaurado BOOLEAN DEFAULT false,
    CONSTRAINT chk_cantidad_nc_positiva CHECK (cantidad > 0),
    CONSTRAINT chk_precios_nc_positivos CHECK (
        precio_unitario >= 0 AND total_item >= 0
    )
);

-- Tabla: despachos (optimizada para control de entregas)
CREATE TABLE despachos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    numero_despacho BIGSERIAL UNIQUE,
    sucursal_id UUID REFERENCES sucursales(id),
    documento_id UUID,
    venta_id UUID,
    usuario_despacho_id UUID REFERENCES usuarios(id),
    cliente_rut TEXT,
    cliente_nombre TEXT,
    estado TEXT DEFAULT 'pendiente',
    fecha_programada TIMESTAMP,
    fecha_inicio TIMESTAMP,
    fecha_completado TIMESTAMP,
    observaciones TEXT,
    datos_adicionales JSONB,
    -- Campos optimizados para eficiencia
    prioridad_despacho INTEGER DEFAULT 5,
    tiempo_estimado_minutos INTEGER,
    codigo_seguimiento TEXT,
    ubicacion_picking TEXT,
    CONSTRAINT chk_estado_despacho CHECK (estado IN (
        'pendiente', 'en_proceso', 'completo', 'parcial', 'rechazado'
    )),
    CONSTRAINT chk_prioridad_despacho CHECK (prioridad_despacho BETWEEN 1 AND 10)
);

-- Tabla: detalle_despacho (optimizada)
CREATE TABLE detalle_despacho (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    despacho_id UUID REFERENCES despachos(id) ON DELETE CASCADE,
    producto_id UUID REFERENCES productos(id),
    cantidad_solicitada NUMERIC(10,3) NOT NULL,
    cantidad_despachada NUMERIC(10,3) DEFAULT 0,
    numero_serie TEXT,
    lote TEXT,
    observaciones TEXT,
    fecha_despacho TIMESTAMP,
    usuario_despacho_id UUID REFERENCES usuarios(id),
    -- Campos optimizados
    ubicacion_producto TEXT,
    orden_picking INTEGER,
    tiempo_picking_segundos INTEGER,
    CONSTRAINT chk_cantidades_despacho_positivas CHECK (
        cantidad_solicitada > 0 AND cantidad_despachada >= 0
    ),
    CONSTRAINT chk_cantidad_despachada_valida CHECK (
        cantidad_despachada <= cantidad_solicitada
    )
);

-- Tabla: reimpresiones_documentos (optimizada para auditoría)
CREATE TABLE reimpresiones_documentos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    documento_id UUID,
    usuario_id UUID REFERENCES usuarios(id),
    sucursal_id UUID REFERENCES sucursales(id),
    tipo_documento TEXT NOT NULL,
    motivo TEXT NOT NULL,
    fecha TIMESTAMP DEFAULT NOW(),
    ip_origen INET,
    dispositivo TEXT,
    reimpresiones_previas INTEGER DEFAULT 0,
    autorizado_por UUID REFERENCES usuarios(id),
    datos_adicionales JSONB,
    -- Campos optimizados
    hash_documento_original TEXT,
    tiempo_procesamiento_ms INTEGER,
    tamaño_archivo_bytes INTEGER
);

-- =====================================================
-- TABLAS DE MONITOREO Y LOGS OPTIMIZADAS
-- =====================================================

-- Tabla: logs_sincronizacion (optimizada para api_sync)
CREATE TABLE logs_sincronizacion (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sucursal_id UUID REFERENCES sucursales(id),
    terminal_id UUID REFERENCES terminales(id),
    tipo_sincronizacion TEXT NOT NULL,
    estado estado_sincronizacion DEFAULT 'pendiente',
    registros_enviados INTEGER DEFAULT 0,
    registros_procesados INTEGER DEFAULT 0,
    registros_error INTEGER DEFAULT 0,
    fecha_inicio TIMESTAMP DEFAULT NOW(),
    fecha_fin TIMESTAMP,
    duracion_segundos INTEGER,
    detalles_error TEXT,
    datos_sincronizados JSONB,
    -- Campos optimizados para api_sync
    proceso_api prioridad_proceso DEFAULT 'media',
    lote_sincronizacion UUID,
    prioridad_sync INTEGER DEFAULT 5,
    recursos_utilizados JSONB,
    CONSTRAINT chk_tipo_sync CHECK (tipo_sincronizacion IN (
        'ventas', 'stock', 'productos', 'clientes', 'configuracion', 'ping'
    )),
    CONSTRAINT chk_registros_positivos CHECK (
        registros_enviados >= 0 AND registros_procesados >= 0 AND registros_error >= 0
    ),
    CONSTRAINT chk_prioridad_sync CHECK (prioridad_sync BETWEEN 1 AND 10)
) PARTITION BY RANGE (fecha_inicio);

-- Crear particiones para logs de sincronización (por semana)
CREATE TABLE logs_sync_2025_w29 PARTITION OF logs_sincronizacion
    FOR VALUES FROM ('2025-07-14') TO ('2025-07-21');
CREATE TABLE logs_sync_2025_w30 PARTITION OF logs_sincronizacion
    FOR VALUES FROM ('2025-07-21') TO ('2025-07-28');
CREATE TABLE logs_sync_2025_w31 PARTITION OF logs_sincronizacion
    FOR VALUES FROM ('2025-07-28') TO ('2025-08-04');
CREATE TABLE logs_sync_2025_w32 PARTITION OF logs_sincronizacion
    FOR VALUES FROM ('2025-08-04') TO ('2025-08-11');
CREATE TABLE logs_sync_futuro PARTITION OF logs_sincronizacion
    FOR VALUES FROM ('2025-08-11') TO (MAXVALUE);

-- Tabla: logs_seguridad (optimizada para auditoría)
CREATE TABLE logs_seguridad (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    usuario_id UUID REFERENCES usuarios(id),
    sucursal_id UUID REFERENCES sucursales(id),
    terminal_id UUID REFERENCES terminales(id),
    evento TEXT NOT NULL,
    nivel_severidad TEXT NOT NULL,
    descripcion TEXT,
    ip_origen INET,
    user_agent TEXT,
    datos_evento JSONB,
    fecha TIMESTAMP DEFAULT NOW(),
    -- Campos optimizados para análisis
    categoria_evento TEXT,
    hash_evento TEXT, -- Para detección de eventos duplicados
    correlacion_id UUID, -- Para correlacionar eventos relacionados
    CONSTRAINT chk_nivel_severidad CHECK (nivel_severidad IN (
        'info', 'warning', 'error', 'critical'
    )),
    CONSTRAINT chk_categoria_evento CHECK (categoria_evento IN (
        'autenticacion', 'autorizacion', 'acceso_datos', 'configuracion',
        'transaccion', 'sistema', 'red', 'aplicacion'
    ))
) PARTITION BY RANGE (fecha);

-- Crear particiones para logs de seguridad (por día)
CREATE TABLE logs_seguridad_2025_07_15 PARTITION OF logs_seguridad
    FOR VALUES FROM ('2025-07-15') TO ('2025-07-16');
CREATE TABLE logs_seguridad_2025_07_16 PARTITION OF logs_seguridad
    FOR VALUES FROM ('2025-07-16') TO ('2025-07-17');
CREATE TABLE logs_seguridad_2025_07_17 PARTITION OF logs_seguridad
    FOR VALUES FROM ('2025-07-17') TO ('2025-07-18');
CREATE TABLE logs_seguridad_2025_07_18 PARTITION OF logs_seguridad
    FOR VALUES FROM ('2025-07-18') TO ('2025-07-19');
CREATE TABLE logs_seguridad_2025_07_19 PARTITION OF logs_seguridad
    FOR VALUES FROM ('2025-07-19') TO ('2025-07-20');
CREATE TABLE logs_seguridad_futuro PARTITION OF logs_seguridad
    FOR VALUES FROM ('2025-07-20') TO (MAXVALUE);

-- Tabla: configuracion_sistema (optimizada para acceso frecuente)
CREATE TABLE configuracion_sistema (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    clave TEXT UNIQUE NOT NULL,
    valor TEXT,
    tipo_dato TEXT DEFAULT 'string',
    descripcion TEXT,
    categoria TEXT,
    modificable BOOLEAN DEFAULT true,
    fecha_creacion TIMESTAMP DEFAULT NOW(),
    fecha_modificacion TIMESTAMP DEFAULT NOW(),
    usuario_modificacion UUID REFERENCES usuarios(id),
    -- Campos optimizados para cache
    cache_ttl_segundos INTEGER DEFAULT 300, -- TTL del cache en segundos
    requiere_reinicio BOOLEAN DEFAULT false, -- Si requiere reinicio para aplicar
    proceso_afectado prioridad_proceso, -- Qué proceso se ve afectado
    valor_anterior TEXT, -- Valor anterior para rollback
    CONSTRAINT chk_tipo_dato CHECK (tipo_dato IN (
        'string', 'integer', 'decimal', 'boolean', 'json', 'date', 'timestamp'
    ))
);

-- Tabla: sesiones_usuario (optimizada para validación rápida)
CREATE TABLE sesiones_usuario (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    usuario_id UUID REFERENCES usuarios(id),
    token_hash TEXT NOT NULL,
    sucursal_id UUID REFERENCES sucursales(id),
    terminal_id UUID REFERENCES terminales(id),
    ip_origen INET,
    user_agent TEXT,
    fecha_inicio TIMESTAMP DEFAULT NOW(),
    fecha_ultimo_acceso TIMESTAMP DEFAULT NOW(),
    fecha_expiracion TIMESTAMP NOT NULL,
    activa BOOLEAN DEFAULT true,
    datos_sesion JSONB,
    -- Campos optimizados para rendimiento
    refresh_token_hash TEXT,
    permisos_cache JSONB, -- Cache de permisos para evitar consultas
    configuracion_cache JSONB, -- Cache de configuración del usuario
    metricas_sesion JSONB -- Métricas de uso de la sesión
);

-- =====================================================
-- TABLAS ESPECÍFICAS PARA REPORTES (api_report)
-- =====================================================

-- Tabla: reportes_programados
-- Descripción: Reportes programados para ejecución automática por api_report
CREATE TABLE reportes_programados (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    nombre TEXT NOT NULL,
    descripcion TEXT,
    tipo_reporte TEXT NOT NULL,
    parametros JSONB NOT NULL,
    cron_expresion TEXT NOT NULL,
    activo BOOLEAN DEFAULT true,
    sucursal_id UUID REFERENCES sucursales(id),
    usuario_creacion UUID REFERENCES usuarios(id),
    destinatarios JSONB, -- Lista de destinatarios del reporte
    formato_salida TEXT DEFAULT 'pdf',
    fecha_creacion TIMESTAMP DEFAULT NOW(),
    fecha_modificacion TIMESTAMP DEFAULT NOW(),
    -- Campos optimizados para ejecución
    proxima_ejecucion TIMESTAMP,
    ultima_ejecucion TIMESTAMP,
    total_ejecuciones INTEGER DEFAULT 0,
    tiempo_promedio_ejecucion_ms INTEGER,
    proceso_api prioridad_proceso DEFAULT 'minima',
    CONSTRAINT chk_tipo_reporte CHECK (tipo_reporte IN (
        'ventas_diarias', 'stock_critico', 'fidelizacion_resumen',
        'productos_mas_vendidos', 'rendimiento_cajeros', 'conciliacion_pagos'
    )),
    CONSTRAINT chk_formato_salida CHECK (formato_salida IN (
        'pdf', 'excel', 'csv', 'json'
    ))
);

-- Tabla: cache_reportes
-- Descripción: Cache de reportes generados para optimizar consultas repetitivas
CREATE TABLE cache_reportes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    hash_consulta TEXT UNIQUE NOT NULL,
    tipo_reporte TEXT NOT NULL,
    parametros JSONB NOT NULL,
    resultado JSONB,
    fecha_generacion TIMESTAMP DEFAULT NOW(),
    fecha_expiracion TIMESTAMP NOT NULL,
    tamaño_resultado_bytes INTEGER,
    tiempo_generacion_ms INTEGER,
    total_accesos INTEGER DEFAULT 0,
    fecha_ultimo_acceso TIMESTAMP DEFAULT NOW(),
    -- Campos para invalidación inteligente
    tablas_dependientes TEXT[], -- Tablas de las que depende el reporte
    version_datos INTEGER DEFAULT 1, -- Versión de los datos para invalidación
    CONSTRAINT chk_tamaño_resultado CHECK (tamaño_resultado_bytes >= 0)
);

-- =====================================================
-- ÍNDICES OPTIMIZADOS PARA ARQUITECTURA CENTRALIZADA
-- =====================================================


-- =====================================================
-- ÍNDICES OPTIMIZADOS PARA PROCESOS API DIFERENCIADOS
-- =====================================================

-- ÍNDICES CRÍTICOS PARA API_POS (MÁXIMA PRIORIDAD)
-- Estos índices deben estar siempre en memoria y optimizados para acceso instantáneo

-- Índices para autenticación y autorización ultra-rápida
CREATE UNIQUE INDEX idx_usuarios_rut_activo ON usuarios(rut) WHERE activo = true;
CREATE INDEX idx_usuarios_hash_sesion ON usuarios(hash_sesion_activa) WHERE hash_sesion_activa IS NOT NULL;
CREATE INDEX idx_sesiones_token_activa ON sesiones_usuario(token_hash) WHERE activa = true;
CREATE INDEX idx_sesiones_usuario_activa ON sesiones_usuario(usuario_id, activa) WHERE activa = true;

-- Índices para búsqueda ultra-rápida de productos en POS
CREATE UNIQUE INDEX idx_productos_codigo_barra_activo ON productos(codigo_barra) WHERE activo = true;
CREATE UNIQUE INDEX idx_productos_codigo_interno_activo ON productos(codigo_interno) WHERE activo = true;
CREATE INDEX idx_productos_busqueda_gin ON productos USING GIN (descripcion_busqueda);
CREATE INDEX idx_productos_popularidad ON productos(popularidad_score DESC, activo) WHERE activo = true;

-- Índices para códigos de barras adicionales
CREATE INDEX idx_codigos_adicionales_codigo_activo ON codigos_barra_adicionales(codigo_barra) WHERE activo = true;
CREATE INDEX idx_codigos_adicionales_producto_activo ON codigos_barra_adicionales(producto_id) WHERE activo = true;

-- Índices críticos para stock en tiempo real
CREATE INDEX idx_stock_central_disponible_critico ON stock_central(sucursal_id, cantidad_disponible) WHERE cantidad_disponible > 0;
CREATE INDEX idx_stock_central_producto_sucursal ON stock_central(producto_id, sucursal_id);
CREATE INDEX idx_stock_central_cache_validez ON stock_central(cache_validez) WHERE cache_validez > NOW() - INTERVAL '5 minutes';

-- Índices para ventas en tiempo real
CREATE INDEX idx_ventas_numero_unico ON ventas(numero_venta);
CREATE INDEX idx_ventas_sucursal_fecha_estado ON ventas(sucursal_id, fecha, estado) WHERE estado = 'finalizada';
CREATE INDEX idx_ventas_terminal_fecha ON ventas(terminal_id, fecha) WHERE fecha >= CURRENT_DATE;
CREATE INDEX idx_ventas_cajero_hoy ON ventas(cajero_id, fecha) WHERE fecha >= CURRENT_DATE;

-- Índices para detalle de ventas (crítico para cálculos)
CREATE INDEX idx_detalle_ventas_venta_producto ON detalle_ventas(venta_id, producto_id);
CREATE INDEX idx_detalle_ventas_producto_fecha ON detalle_ventas(producto_id, venta_id);

-- Índices para medios de pago (crítico para cierre de caja)
CREATE INDEX idx_medios_pago_venta_medio ON medios_pago_venta(venta_id, medio_pago);
CREATE INDEX idx_medios_pago_fecha_medio ON medios_pago_venta(fecha_procesamiento, medio_pago) WHERE fecha_procesamiento >= CURRENT_DATE;

-- Índices para notas de venta (flujo POS Tienda -> Caja)
CREATE INDEX idx_notas_venta_qr ON notas_venta(qr_code) WHERE qr_code IS NOT NULL;
CREATE INDEX idx_notas_venta_estado_sucursal ON notas_venta(estado, sucursal_id) WHERE estado = 'pendiente';
CREATE INDEX idx_notas_venta_prioridad ON notas_venta(prioridad_atencion, fecha) WHERE estado = 'pendiente';

-- Índices para fidelización (consultas frecuentes en POS)
CREATE UNIQUE INDEX idx_fidelizacion_rut_activo ON fidelizacion_clientes(rut) WHERE activo = true;
CREATE INDEX idx_fidelizacion_qr ON fidelizacion_clientes(qr_fidelizacion) WHERE qr_fidelizacion IS NOT NULL;
CREATE INDEX idx_fidelizacion_hash_busqueda ON fidelizacion_clientes(hash_busqueda) WHERE activo = true;

-- ÍNDICES PARA API_SYNC (PRIORIDAD MEDIA)
-- Optimizados para operaciones de sincronización masiva

-- Índices para sincronización de ventas
CREATE INDEX idx_ventas_sincronizada_fecha ON ventas(sincronizada, fecha) WHERE sincronizada = false;
CREATE INDEX idx_ventas_fecha_sync ON ventas(fecha_sincronizacion) WHERE fecha_sincronizacion IS NOT NULL;

-- Índices para sincronización de stock
CREATE INDEX idx_stock_central_fecha_sync ON stock_central(fecha_sync);
CREATE INDEX idx_movimientos_stock_fecha_sync ON movimientos_stock(fecha) WHERE fecha >= CURRENT_DATE - INTERVAL '7 days';
CREATE INDEX idx_movimientos_stock_batch ON movimientos_stock(batch_id) WHERE batch_id IS NOT NULL;

-- Índices para logs de sincronización
CREATE INDEX idx_logs_sync_estado_fecha ON logs_sincronizacion(estado, fecha_inicio) WHERE estado IN ('pendiente', 'en_proceso');
CREATE INDEX idx_logs_sync_lote ON logs_sincronizacion(lote_sincronizacion) WHERE lote_sincronizacion IS NOT NULL;
CREATE INDEX idx_logs_sync_tipo_sucursal ON logs_sincronizacion(tipo_sincronizacion, sucursal_id);

-- ÍNDICES PARA API_LABELS (PRIORIDAD BAJA)
-- Optimizados para operaciones de etiquetas sin afectar rendimiento crítico

-- Índices para plantillas de etiquetas
CREATE INDEX idx_etiquetas_plantillas_categoria ON etiquetas_plantillas(categoria_producto_id) WHERE activa = true;
CREATE INDEX idx_etiquetas_plantillas_tipo ON etiquetas_plantillas(tipo_etiqueta, activa) WHERE activa = true;
CREATE INDEX idx_etiquetas_plantillas_predeterminada ON etiquetas_plantillas(predeterminada) WHERE predeterminada = true;

-- Índices para trabajos de impresión
CREATE INDEX idx_etiquetas_trabajos_estado ON etiquetas_trabajos_impresion(estado) WHERE estado IN ('pendiente', 'procesando');
CREATE INDEX idx_etiquetas_trabajos_prioridad ON etiquetas_trabajos_impresion(prioridad, fecha_solicitud) WHERE estado = 'pendiente';
CREATE INDEX idx_etiquetas_trabajos_usuario ON etiquetas_trabajos_impresion(usuario_id, fecha_solicitud);
CREATE INDEX idx_etiquetas_trabajos_sucursal ON etiquetas_trabajos_impresion(sucursal_id, fecha_solicitud);

-- Índices para detalle de trabajos
CREATE INDEX idx_etiquetas_detalle_trabajo ON etiquetas_detalle_trabajo(trabajo_id);
CREATE INDEX idx_etiquetas_detalle_producto ON etiquetas_detalle_trabajo(producto_id);
CREATE INDEX idx_etiquetas_detalle_estado ON etiquetas_detalle_trabajo(estado_item) WHERE estado_item IN ('pendiente', 'procesando');

-- Índices para configuraciones de impresora
CREATE INDEX idx_etiquetas_impresoras_activa ON etiquetas_configuraciones_impresora(activa, sucursal_id) WHERE activa = true;
CREATE INDEX idx_etiquetas_impresoras_estado ON etiquetas_configuraciones_impresora(estado_impresora) WHERE estado_impresora = 'disponible';

-- Índices para cache de códigos de barras
CREATE INDEX idx_etiquetas_cache_codigo_tipo ON etiquetas_cache_codigos_barras(codigo_producto, tipo_codigo);
CREATE INDEX idx_etiquetas_cache_hash ON etiquetas_cache_codigos_barras(hash_parametros);
CREATE INDEX idx_etiquetas_cache_valido ON etiquetas_cache_codigos_barras(valido_hasta) WHERE valido_hasta > NOW();

-- Índices para métricas de rendimiento
CREATE INDEX idx_etiquetas_metricas_tipo_fecha ON etiquetas_metricas_rendimiento(tipo_metrica, fecha_hora);
CREATE INDEX idx_etiquetas_metricas_sucursal_periodo ON etiquetas_metricas_rendimiento(sucursal_id, periodo_agregacion, fecha_periodo);

-- ÍNDICES PARA API_REPORT (PRIORIDAD MÍNIMA)
-- Optimizados para consultas analíticas sin impactar operaciones críticas

-- Índices para reportes programados
CREATE INDEX idx_reportes_programados_proxima ON reportes_programados(proxima_ejecucion) WHERE activo = true;
CREATE INDEX idx_reportes_programados_tipo ON reportes_programados(tipo_reporte, activo) WHERE activo = true;

-- Índices para cache de reportes
CREATE INDEX idx_cache_reportes_hash ON cache_reportes(hash_consulta);
CREATE INDEX idx_cache_reportes_expiracion ON cache_reportes(fecha_expiracion) WHERE fecha_expiracion > NOW();
CREATE INDEX idx_cache_reportes_tipo_acceso ON cache_reportes(tipo_reporte, fecha_ultimo_acceso);

-- Índices para análisis de ventas (reportes)
CREATE INDEX idx_ventas_reporte_fecha_sucursal ON ventas(DATE(fecha), sucursal_id) WHERE estado = 'finalizada';
CREATE INDEX idx_ventas_reporte_cajero_fecha ON ventas(cajero_id, DATE(fecha)) WHERE estado = 'finalizada';
CREATE INDEX idx_ventas_reporte_total_fecha ON ventas(total, fecha) WHERE estado = 'finalizada' AND fecha >= CURRENT_DATE - INTERVAL '90 days';

-- Índices para análisis de productos (reportes)
CREATE INDEX idx_detalle_ventas_reporte_producto ON detalle_ventas(producto_id, venta_id);
CREATE INDEX idx_productos_reporte_categoria ON productos(categoria_id, activo) WHERE activo = true;

-- ÍNDICES GENERALES OPTIMIZADOS

-- Índices para sucursales
CREATE INDEX idx_sucursales_codigo_habilitada ON sucursales(codigo) WHERE habilitada = true;
CREATE INDEX idx_sucursales_region_habilitada ON sucursales(region, habilitada) WHERE habilitada = true;

-- Índices para terminales
CREATE INDEX idx_terminales_sucursal_activo ON terminales(sucursal_id, activo) WHERE activo = true;
CREATE INDEX idx_terminales_estado_conexion ON terminales(estado_conexion, ultima_conexion);
CREATE INDEX idx_terminales_heartbeat ON terminales(ultima_conexion) WHERE activo = true;

-- Índices para categorías de productos
CREATE INDEX idx_categorias_padre_activa ON categorias_productos(categoria_padre_id, activa) WHERE activa = true;
CREATE INDEX idx_categorias_nivel_orden ON categorias_productos(nivel, orden_visualizacion) WHERE activa = true;

-- Índices para documentos DTE
CREATE INDEX idx_documentos_dte_estado_fecha ON documentos_dte(estado, fecha_emision) WHERE estado IN ('pendiente', 'error');
CREATE INDEX idx_documentos_dte_sucursal_tipo ON documentos_dte(sucursal_id, tipo_documento);
CREATE INDEX idx_documentos_dte_track ON documentos_dte(track_id) WHERE track_id IS NOT NULL;
CREATE INDEX idx_documentos_dte_prioridad ON documentos_dte(prioridad_procesamiento, fecha_emision) WHERE estado = 'pendiente';

-- Índices para folios DTE
CREATE INDEX idx_folios_dte_sucursal_tipo_activo ON folios_dte(sucursal_id, tipo_documento, activo) WHERE activo = true;
CREATE INDEX idx_folios_dte_agotamiento ON folios_dte(sucursal_id, tipo_documento, folio_actual, folio_hasta) WHERE activo = true;

-- Índices para despachos
CREATE INDEX idx_despachos_estado_fecha ON despachos(estado, fecha_programada) WHERE estado IN ('pendiente', 'en_proceso');
CREATE INDEX idx_despachos_prioridad ON despachos(prioridad_despacho, fecha_programada) WHERE estado = 'pendiente';
CREATE INDEX idx_despachos_usuario_fecha ON despachos(usuario_despacho_id, fecha_programada);

-- Índices para logs de seguridad
CREATE INDEX idx_logs_seguridad_evento_fecha ON logs_seguridad(evento, fecha);
CREATE INDEX idx_logs_seguridad_usuario_fecha ON logs_seguridad(usuario_id, fecha) WHERE usuario_id IS NOT NULL;
CREATE INDEX idx_logs_seguridad_nivel_fecha ON logs_seguridad(nivel_severidad, fecha) WHERE nivel_severidad IN ('error', 'critical');
CREATE INDEX idx_logs_seguridad_correlacion ON logs_seguridad(correlacion_id) WHERE correlacion_id IS NOT NULL;

-- Índices para configuración del sistema
CREATE INDEX idx_config_categoria_modificable ON configuracion_sistema(categoria, modificable);
CREATE INDEX idx_config_proceso_afectado ON configuracion_sistema(proceso_afectado) WHERE proceso_afectado IS NOT NULL;

-- =====================================================
-- ÍNDICES ESPECIALIZADOS PARA OPTIMIZACIÓN EXTREMA
-- =====================================================

-- Índices parciales para operaciones críticas de POS
CREATE INDEX idx_productos_pos_critico ON productos(id, codigo_barra, descripcion_corta, precio_unitario) 
    WHERE activo = true AND codigo_barra IS NOT NULL;

CREATE INDEX idx_stock_pos_critico ON stock_central(producto_id, sucursal_id, cantidad_disponible) 
    WHERE cantidad_disponible > 0;

-- Índices de cobertura para consultas frecuentes
CREATE INDEX idx_ventas_cobertura_pos ON ventas(id, numero_venta, sucursal_id, terminal_id, cajero_id, total, estado, fecha)
    WHERE estado = 'finalizada';

CREATE INDEX idx_usuarios_cobertura_auth ON usuarios(id, rut, rol, sucursal_id, activo, cache_permisos)
    WHERE activo = true;

-- Índices para búsqueda de texto optimizada
CREATE INDEX idx_productos_descripcion_trgm ON productos USING GIN (descripcion gin_trgm_ops) WHERE activo = true;
CREATE INDEX idx_productos_marca_trgm ON productos USING GIN (marca gin_trgm_ops) WHERE activo = true AND marca IS NOT NULL;
CREATE INDEX idx_fidelizacion_nombre_trgm ON fidelizacion_clientes USING GIN ((nombre || ' ' || COALESCE(apellido, '')) gin_trgm_ops) WHERE activo = true;

-- Índices para agregaciones rápidas (reportes)
CREATE INDEX idx_ventas_agregacion_diaria ON ventas(DATE(fecha), sucursal_id, total) WHERE estado = 'finalizada';
CREATE INDEX idx_movimientos_stock_agregacion ON movimientos_stock(DATE(fecha), tipo_movimiento, sucursal_id, cantidad);

-- =====================================================
-- FUNCIONES OPTIMIZADAS PARA ARQUITECTURA CENTRALIZADA
-- =====================================================

-- Función optimizada para validación de stock con control de concurrencia
CREATE OR REPLACE FUNCTION validar_stock_disponible_optimizado(
    p_producto_id UUID,
    p_sucursal_id UUID,
    p_cantidad NUMERIC
) RETURNS BOOLEAN AS $$
DECLARE
    v_stock_disponible INTEGER;
    v_version_lock INTEGER;
BEGIN
    -- Consulta optimizada con lock para evitar condiciones de carrera
    SELECT cantidad_disponible, version_optimistic_lock 
    INTO v_stock_disponible, v_version_lock
    FROM stock_central
    WHERE producto_id = p_producto_id AND sucursal_id = p_sucursal_id
    FOR UPDATE NOWAIT;
    
    IF NOT FOUND THEN
        RETURN false;
    END IF;
    
    RETURN v_stock_disponible >= p_cantidad;
EXCEPTION
    WHEN lock_not_available THEN
        -- Si no se puede obtener el lock, asumir que no hay stock
        RETURN false;
END;
$$ LANGUAGE plpgsql;

-- Función optimizada para descuento de stock con control de concurrencia optimista
CREATE OR REPLACE FUNCTION descontar_stock_optimizado(
    p_producto_id UUID,
    p_sucursal_id UUID,
    p_cantidad NUMERIC,
    p_venta_id UUID,
    p_usuario_id UUID
) RETURNS BOOLEAN AS $$
DECLARE
    v_stock_actual INTEGER;
    v_version_actual INTEGER;
    v_filas_afectadas INTEGER;
BEGIN
    -- Obtener stock y versión actual
    SELECT cantidad_disponible, version_optimistic_lock
    INTO v_stock_actual, v_version_actual
    FROM stock_central
    WHERE producto_id = p_producto_id AND sucursal_id = p_sucursal_id;
    
    IF NOT FOUND OR v_stock_actual < p_cantidad THEN
        RETURN false;
    END IF;
    
    -- Actualización con control de concurrencia optimista
    UPDATE stock_central
    SET cantidad = cantidad - p_cantidad,
        version_optimistic_lock = version_optimistic_lock + 1,
        fecha_sync = NOW(),
        fecha_ultima_salida = NOW()
    WHERE producto_id = p_producto_id 
      AND sucursal_id = p_sucursal_id
      AND version_optimistic_lock = v_version_actual;
    
    GET DIAGNOSTICS v_filas_afectadas = ROW_COUNT;
    
    IF v_filas_afectadas = 0 THEN
        -- Conflicto de concurrencia, reintentar
        RETURN false;
    END IF;
    
    -- Registrar movimiento de stock de forma asíncrona
    INSERT INTO movimientos_stock (
        producto_id, sucursal_id, tipo_movimiento, cantidad,
        cantidad_anterior, cantidad_nueva, documento_referencia,
        usuario_id, proceso_origen, observaciones
    ) VALUES (
        p_producto_id, p_sucursal_id, 'venta', -p_cantidad,
        v_stock_actual, v_stock_actual - p_cantidad,
        'VENTA-' || (SELECT numero_venta FROM ventas WHERE id = p_venta_id),
        p_usuario_id, 'maxima', 'Descuento automático por venta'
    );
    
    RETURN true;
END;
$$ LANGUAGE plpgsql;

-- Función optimizada para búsqueda de productos en POS
CREATE OR REPLACE FUNCTION buscar_producto_pos(
    p_termino_busqueda TEXT,
    p_sucursal_id UUID,
    p_limite INTEGER DEFAULT 10
) RETURNS TABLE (
    producto_id UUID,
    codigo_interno TEXT,
    codigo_barra TEXT,
    descripcion TEXT,
    precio_unitario NUMERIC,
    stock_disponible INTEGER,
    score REAL
) AS $$
BEGIN
    RETURN QUERY
    WITH busqueda_productos AS (
        -- Búsqueda exacta por código de barras (máxima prioridad)
        SELECT p.id, p.codigo_interno, p.codigo_barra, p.descripcion, 
               p.precio_unitario, COALESCE(sc.cantidad_disponible, 0) as stock,
               1.0::REAL as score
        FROM productos p
        LEFT JOIN stock_central sc ON p.id = sc.producto_id AND sc.sucursal_id = p_sucursal_id
        WHERE p.activo = true 
          AND (p.codigo_barra = p_termino_busqueda OR p.codigo_interno = p_termino_busqueda)
        
        UNION ALL
        
        -- Búsqueda por códigos adicionales
        SELECT p.id, p.codigo_interno, p.codigo_barra, p.descripcion,
               p.precio_unitario, COALESCE(sc.cantidad_disponible, 0) as stock,
               0.9::REAL as score
        FROM productos p
        JOIN codigos_barra_adicionales cba ON p.id = cba.producto_id
        LEFT JOIN stock_central sc ON p.id = sc.producto_id AND sc.sucursal_id = p_sucursal_id
        WHERE p.activo = true AND cba.activo = true
          AND cba.codigo_barra = p_termino_busqueda
        
        UNION ALL
        
        -- Búsqueda por texto en descripción
        SELECT p.id, p.codigo_interno, p.codigo_barra, p.descripcion,
               p.precio_unitario, COALESCE(sc.cantidad_disponible, 0) as stock,
               ts_rank(p.descripcion_busqueda, plainto_tsquery('spanish', p_termino_busqueda))::REAL as score
        FROM productos p
        LEFT JOIN stock_central sc ON p.id = sc.producto_id AND sc.sucursal_id = p_sucursal_id
        WHERE p.activo = true
          AND p.descripcion_busqueda @@ plainto_tsquery('spanish', p_termino_busqueda)
    )
    SELECT DISTINCT bp.id, bp.codigo_interno, bp.codigo_barra, bp.descripcion,
           bp.precio_unitario, bp.stock, MAX(bp.score) as max_score
    FROM busqueda_productos bp
    GROUP BY bp.id, bp.codigo_interno, bp.codigo_barra, bp.descripcion, bp.precio_unitario, bp.stock
    ORDER BY max_score DESC, bp.stock DESC
    LIMIT p_limite;
END;
$$ LANGUAGE plpgsql;

-- Función optimizada para autenticación rápida
CREATE OR REPLACE FUNCTION autenticar_usuario_rapido(
    p_rut TEXT,
    p_password_hash TEXT
) RETURNS TABLE (
    usuario_id UUID,
    nombre_completo TEXT,
    rol rol_usuario,
    sucursal_id UUID,
    permisos JSONB,
    token_sesion TEXT
) AS $$
DECLARE
    v_usuario_id UUID;
    v_salt TEXT;
    v_password_hash_stored TEXT;
    v_intentos_fallidos INTEGER;
    v_bloqueado_hasta TIMESTAMP;
    v_token_sesion TEXT;
BEGIN
    -- Consulta optimizada con índice específico
    SELECT u.id, u.salt, u.password_hash, u.intentos_fallidos, u.bloqueado_hasta
    INTO v_usuario_id, v_salt, v_password_hash_stored, v_intentos_fallidos, v_bloqueado_hasta
    FROM usuarios u
    WHERE u.rut = p_rut AND u.activo = true;
    
    IF NOT FOUND THEN
        RETURN;
    END IF;
    
    -- Verificar si está bloqueado
    IF v_bloqueado_hasta IS NOT NULL AND v_bloqueado_hasta > NOW() THEN
        RETURN;
    END IF;
    
    -- Verificar contraseña
    IF v_password_hash_stored != crypt(p_password_hash, v_salt) THEN
        -- Incrementar intentos fallidos
        UPDATE usuarios 
        SET intentos_fallidos = intentos_fallidos + 1,
            bloqueado_hasta = CASE 
                WHEN intentos_fallidos + 1 >= 5 THEN NOW() + INTERVAL '30 minutes'
                ELSE NULL
            END
        WHERE id = v_usuario_id;
        RETURN;
    END IF;
    
    -- Generar token de sesión
    v_token_sesion := encode(gen_random_bytes(32), 'hex');
    
    -- Actualizar último acceso y limpiar intentos fallidos
    UPDATE usuarios 
    SET ultimo_acceso = NOW(),
        intentos_fallidos = 0,
        bloqueado_hasta = NULL,
        hash_sesion_activa = v_token_sesion
    WHERE id = v_usuario_id;
    
    -- Retornar información del usuario
    RETURN QUERY
    SELECT u.id, u.nombre || ' ' || COALESCE(u.apellido, ''), u.rol, u.sucursal_id,
           u.cache_permisos, v_token_sesion
    FROM usuarios u
    WHERE u.id = v_usuario_id;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Función para generar código de barras optimizada para api_labels
CREATE OR REPLACE FUNCTION generar_codigo_barras_optimizado(
    p_codigo_producto TEXT,
    p_tipo_codigo TEXT DEFAULT 'CODE39',
    p_parametros JSONB DEFAULT '{}'::JSONB
) RETURNS TABLE (
    codigo_barras TEXT,
    imagen_base64 TEXT,
    cache_hit BOOLEAN
) AS $$
DECLARE
    v_hash_parametros TEXT;
    v_codigo_barras TEXT;
    v_imagen_base64 TEXT;
    v_cache_hit BOOLEAN := false;
BEGIN
    -- Generar hash de parámetros para búsqueda en cache
    v_hash_parametros := encode(digest(p_codigo_producto || p_tipo_codigo || p_parametros::TEXT, 'sha256'), 'hex');
    
    -- Buscar en cache
    SELECT codigo_barras_generado, imagen_base64, true
    INTO v_codigo_barras, v_imagen_base64, v_cache_hit
    FROM etiquetas_cache_codigos_barras
    WHERE hash_parametros = v_hash_parametros
      AND valido_hasta > NOW();
    
    IF FOUND THEN
        -- Actualizar estadísticas de uso
        UPDATE etiquetas_cache_codigos_barras
        SET total_usos = total_usos + 1,
            fecha_ultimo_uso = NOW()
        WHERE hash_parametros = v_hash_parametros;
        
        RETURN QUERY SELECT v_codigo_barras, v_imagen_base64, v_cache_hit;
        RETURN;
    END IF;
    
    -- Generar nuevo código de barras (simulado - en implementación real usar librería específica)
    v_codigo_barras := '*' || UPPER(p_codigo_producto) || '*'; -- Formato CODE39 básico
    v_imagen_base64 := 'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChwGA60e6kgAAAABJRU5ErkJggg=='; -- Imagen placeholder
    
    -- Guardar en cache
    INSERT INTO etiquetas_cache_codigos_barras (
        codigo_producto, tipo_codigo, codigo_barras_generado, imagen_base64,
        parametros_generacion, hash_parametros, tiempo_generacion_ms
    ) VALUES (
        p_codigo_producto, p_tipo_codigo, v_codigo_barras, v_imagen_base64,
        p_parametros, v_hash_parametros, 50 -- Tiempo simulado
    ) ON CONFLICT (codigo_producto, tipo_codigo, hash_parametros) DO UPDATE SET
        fecha_ultimo_uso = NOW(),
        total_usos = etiquetas_cache_codigos_barras.total_usos + 1;
    
    RETURN QUERY SELECT v_codigo_barras, v_imagen_base64, false;
END;
$$ LANGUAGE plpgsql;

-- Función para limpiar cache de reportes (api_report)
CREATE OR REPLACE FUNCTION limpiar_cache_reportes()
RETURNS INTEGER AS $$
DECLARE
    v_eliminados INTEGER;
BEGIN
    -- Eliminar reportes expirados
    DELETE FROM cache_reportes 
    WHERE fecha_expiracion < NOW();
    
    GET DIAGNOSTICS v_eliminados = ROW_COUNT;
    
    -- Eliminar reportes menos utilizados si el cache está muy grande
    WITH reportes_menos_usados AS (
        SELECT id
        FROM cache_reportes
        ORDER BY total_accesos ASC, fecha_ultimo_acceso ASC
        OFFSET 1000 -- Mantener solo los 1000 reportes más utilizados
    )
    DELETE FROM cache_reportes
    WHERE id IN (SELECT id FROM reportes_menos_usados);
    
    GET DIAGNOSTICS v_eliminados = v_eliminados + ROW_COUNT;
    
    RETURN v_eliminados;
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- TRIGGERS OPTIMIZADOS PARA ARQUITECTURA CENTRALIZADA
-- =====================================================

-- Trigger optimizado para actualización de fecha de modificación
CREATE OR REPLACE FUNCTION actualizar_fecha_modificacion_optimizado()
RETURNS TRIGGER AS $$
BEGIN
    NEW.fecha_modificacion = NOW();
    
    -- Invalidar cache relacionado si existe
    IF TG_TABLE_NAME = 'productos' THEN
        NEW.cache_codigo_barras_generado = NULL;
    ELSIF TG_TABLE_NAME = 'usuarios' THEN
        NEW.cache_permisos = NULL;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger optimizado para descuento de stock
CREATE OR REPLACE FUNCTION trigger_descontar_stock_optimizado()
RETURNS TRIGGER AS $$
DECLARE
    v_sucursal_id UUID;
    v_exito BOOLEAN;
BEGIN
    -- Obtener sucursal de la venta
    SELECT sucursal_id INTO v_sucursal_id
    FROM ventas
    WHERE id = NEW.venta_id;
    
    -- Intentar descuento optimizado
    SELECT descontar_stock_optimizado(
        NEW.producto_id, 
        v_sucursal_id, 
        NEW.cantidad, 
        NEW.venta_id,
        (SELECT cajero_id FROM ventas WHERE id = NEW.venta_id)
    ) INTO v_exito;
    
    IF NOT v_exito THEN
        RAISE EXCEPTION 'No se pudo descontar stock para producto % en sucursal %', 
            NEW.producto_id, v_sucursal_id;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger para actualizar popularidad de productos
CREATE OR REPLACE FUNCTION actualizar_popularidad_producto()
RETURNS TRIGGER AS $$
BEGIN
    -- Incrementar popularidad basada en ventas recientes
    UPDATE productos
    SET popularidad_score = LEAST(100, popularidad_score + 0.1),
        fecha_modificacion = NOW()
    WHERE id = NEW.producto_id;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger para invalidar cache de reportes
CREATE OR REPLACE FUNCTION invalidar_cache_reportes()
RETURNS TRIGGER AS $$
DECLARE
    v_tablas_afectadas TEXT[];
BEGIN
    -- Determinar qué tablas se ven afectadas
    CASE TG_TABLE_NAME
        WHEN 'ventas' THEN v_tablas_afectadas := ARRAY['ventas', 'detalle_ventas', 'medios_pago_venta'];
        WHEN 'productos' THEN v_tablas_afectadas := ARRAY['productos', 'stock_central'];
        WHEN 'stock_central' THEN v_tablas_afectadas := ARRAY['stock_central', 'movimientos_stock'];
        ELSE v_tablas_afectadas := ARRAY[TG_TABLE_NAME];
    END CASE;
    
    -- Invalidar cache de reportes que dependen de estas tablas
    UPDATE cache_reportes
    SET fecha_expiracion = NOW()
    WHERE tablas_dependientes && v_tablas_afectadas;
    
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

-- Aplicar triggers optimizados
CREATE TRIGGER trg_productos_fecha_mod_opt
    BEFORE UPDATE ON productos
    FOR EACH ROW EXECUTE FUNCTION actualizar_fecha_modificacion_optimizado();

CREATE TRIGGER trg_usuarios_fecha_mod_opt
    BEFORE UPDATE ON usuarios
    FOR EACH ROW EXECUTE FUNCTION actualizar_fecha_modificacion_optimizado();

CREATE TRIGGER trg_descontar_stock_opt
    AFTER INSERT ON detalle_ventas
    FOR EACH ROW EXECUTE FUNCTION trigger_descontar_stock_optimizado();

CREATE TRIGGER trg_popularidad_producto
    AFTER INSERT ON detalle_ventas
    FOR EACH ROW EXECUTE FUNCTION actualizar_popularidad_producto();

CREATE TRIGGER trg_invalidar_cache_ventas
    AFTER INSERT OR UPDATE OR DELETE ON ventas
    FOR EACH ROW EXECUTE FUNCTION invalidar_cache_reportes();

CREATE TRIGGER trg_invalidar_cache_productos
    AFTER UPDATE ON productos
    FOR EACH ROW EXECUTE FUNCTION invalidar_cache_reportes();

-- =====================================================
-- VISTAS OPTIMIZADAS PARA ARQUITECTURA CENTRALIZADA
-- =====================================================


-- =====================================================
-- VISTAS OPTIMIZADAS PARA ARQUITECTURA CENTRALIZADA
-- =====================================================

-- Vista optimizada para stock consolidado (crítica para api_pos)
CREATE MATERIALIZED VIEW vista_stock_consolidado_optimizada AS
SELECT 
    sc.producto_id,
    p.codigo_interno,
    p.codigo_barra,
    p.descripcion,
    p.descripcion_corta,
    p.marca,
    p.precio_unitario,
    sc.sucursal_id,
    s.nombre AS sucursal_nombre,
    s.codigo AS sucursal_codigo,
    sc.cantidad,
    sc.cantidad_reservada,
    sc.cantidad_disponible,
    sc.costo_promedio,
    sc.fecha_ultima_entrada,
    sc.fecha_ultima_salida,
    sc.fecha_sync,
    CASE 
        WHEN sc.cantidad_disponible <= 0 THEN 'SIN_STOCK'
        WHEN sc.cantidad_disponible <= p.stock_minimo THEN 'CRITICO'
        WHEN sc.cantidad_disponible <= p.stock_minimo * 2 THEN 'BAJO'
        ELSE 'NORMAL'
    END AS estado_stock,
    p.popularidad_score,
    p.activo AS producto_activo,
    s.habilitada AS sucursal_habilitada,
    -- Campos calculados para optimización
    (sc.cantidad_disponible * p.precio_unitario) AS valor_stock_disponible,
    EXTRACT(DAYS FROM NOW() - COALESCE(sc.fecha_ultima_salida, sc.fecha_sync)) AS dias_sin_movimiento
FROM stock_central sc
JOIN productos p ON sc.producto_id = p.id
JOIN sucursales s ON sc.sucursal_id = s.id
WHERE p.activo = true AND s.habilitada = true;

-- Índices para la vista materializada
CREATE UNIQUE INDEX idx_vista_stock_producto_sucursal ON vista_stock_consolidada_optimizada(producto_id, sucursal_id);
CREATE INDEX idx_vista_stock_codigo_barra ON vista_stock_consolidada_optimizada(codigo_barra);
CREATE INDEX idx_vista_stock_estado ON vista_stock_consolidada_optimizada(estado_stock, sucursal_id);
CREATE INDEX idx_vista_stock_disponible ON vista_stock_consolidada_optimizada(cantidad_disponible) WHERE cantidad_disponible > 0;

-- Vista optimizada para resumen de ventas diarias (para api_report)
CREATE MATERIALIZED VIEW vista_ventas_resumen_diario_optimizada AS
SELECT 
    DATE(v.fecha) AS fecha,
    v.sucursal_id,
    s.nombre AS sucursal_nombre,
    s.codigo AS sucursal_codigo,
    COUNT(*) AS total_ventas,
    SUM(v.total) AS total_monto,
    AVG(v.total) AS promedio_venta,
    MIN(v.total) AS venta_minima,
    MAX(v.total) AS venta_maxima,
    COUNT(DISTINCT v.cajero_id) AS cajeros_activos,
    COUNT(DISTINCT v.cliente_rut) AS clientes_unicos,
    COUNT(DISTINCT v.terminal_id) AS terminales_activos,
    -- Métricas por medio de pago
    SUM(CASE WHEN mp.medio_pago = 'efectivo' THEN mp.monto ELSE 0 END) AS total_efectivo,
    SUM(CASE WHEN mp.medio_pago = 'tarjeta_debito' THEN mp.monto ELSE 0 END) AS total_debito,
    SUM(CASE WHEN mp.medio_pago = 'tarjeta_credito' THEN mp.monto ELSE 0 END) AS total_credito,
    -- Métricas de rendimiento
    AVG(v.tiempo_procesamiento_ms) AS tiempo_promedio_procesamiento_ms,
    COUNT(*) FILTER (WHERE v.tiempo_procesamiento_ms > 5000) AS ventas_lentas,
    -- Métricas de productos
    SUM((SELECT SUM(dv.cantidad) FROM detalle_ventas dv WHERE dv.venta_id = v.id)) AS total_items_vendidos
FROM ventas v
JOIN sucursales s ON v.sucursal_id = s.id
LEFT JOIN medios_pago_venta mp ON v.id = mp.venta_id
WHERE v.estado = 'finalizada'
GROUP BY DATE(v.fecha), v.sucursal_id, s.nombre, s.codigo;

-- Índices para resumen de ventas diarias
CREATE UNIQUE INDEX idx_vista_ventas_fecha_sucursal ON vista_ventas_resumen_diario_optimizada(fecha, sucursal_id);
CREATE INDEX idx_vista_ventas_fecha ON vista_ventas_resumen_diario_optimizada(fecha DESC);
CREATE INDEX idx_vista_ventas_monto ON vista_ventas_resumen_diario_optimizada(total_monto DESC);

-- Vista optimizada para productos más vendidos (para api_report)
CREATE MATERIALIZED VIEW vista_productos_mas_vendidos_optimizada AS
WITH ventas_recientes AS (
    SELECT v.id, v.fecha, v.sucursal_id
    FROM ventas v
    WHERE v.estado = 'finalizada' 
      AND v.fecha >= CURRENT_DATE - INTERVAL '30 days'
),
detalle_con_ventas AS (
    SELECT dv.*, vr.fecha, vr.sucursal_id
    FROM detalle_ventas dv
    JOIN ventas_recientes vr ON dv.venta_id = vr.id
)
SELECT 
    dv.producto_id,
    p.codigo_interno,
    p.codigo_barra,
    p.descripcion,
    p.descripcion_corta,
    p.marca,
    p.precio_unitario,
    c.nombre AS categoria_nombre,
    SUM(dv.cantidad) AS cantidad_vendida,
    COUNT(DISTINCT dv.venta_id) AS numero_ventas,
    COUNT(DISTINCT dv.sucursal_id) AS sucursales_vendido,
    SUM(dv.total_item) AS monto_total_vendido,
    AVG(dv.precio_final) AS precio_promedio_venta,
    AVG(dv.margen_unitario) AS margen_promedio,
    -- Métricas de tendencia
    SUM(CASE WHEN dv.fecha >= CURRENT_DATE - INTERVAL '7 days' THEN dv.cantidad ELSE 0 END) AS cantidad_ultima_semana,
    SUM(CASE WHEN dv.fecha >= CURRENT_DATE - INTERVAL '7 days' THEN dv.total_item ELSE 0 END) AS monto_ultima_semana,
    -- Ranking
    RANK() OVER (ORDER BY SUM(dv.cantidad) DESC) AS ranking_cantidad,
    RANK() OVER (ORDER BY SUM(dv.total_item) DESC) AS ranking_monto
FROM detalle_con_ventas dv
JOIN productos p ON dv.producto_id = p.id
LEFT JOIN categorias_productos c ON p.categoria_id = c.id
WHERE p.activo = true
GROUP BY dv.producto_id, p.codigo_interno, p.codigo_barra, p.descripcion, 
         p.descripcion_corta, p.marca, p.precio_unitario, c.nombre;

-- Índices para productos más vendidos
CREATE UNIQUE INDEX idx_vista_productos_vendidos_id ON vista_productos_mas_vendidos_optimizada(producto_id);
CREATE INDEX idx_vista_productos_vendidos_ranking ON vista_productos_mas_vendidos_optimizada(ranking_cantidad);
CREATE INDEX idx_vista_productos_vendidos_categoria ON vista_productos_mas_vendidos_optimizada(categoria_nombre, cantidad_vendida DESC);

-- Vista optimizada para fidelización (para api_pos y api_report)
CREATE MATERIALIZED VIEW vista_fidelizacion_resumen_optimizada AS
SELECT 
    fc.id,
    fc.rut,
    fc.nombre,
    fc.apellido,
    fc.email,
    fc.telefono,
    fc.puntos_actuales,
    fc.puntos_acumulados_total,
    fc.nivel_fidelizacion,
    fc.fecha_ultima_compra,
    fc.fecha_ultima_actividad,
    fc.activo,
    -- Estadísticas calculadas
    COUNT(mf.id) AS total_movimientos,
    SUM(CASE WHEN mf.tipo = 'acumulacion' THEN mf.puntos ELSE 0 END) AS puntos_acumulados,
    SUM(CASE WHEN mf.tipo = 'canje' THEN ABS(mf.puntos) ELSE 0 END) AS puntos_canjeados,
    COUNT(DISTINCT v.id) AS total_compras,
    COALESCE(SUM(v.total), 0) AS monto_total_compras,
    COALESCE(AVG(v.total), 0) AS promedio_compra,
    -- Métricas de actividad
    COUNT(DISTINCT DATE(v.fecha)) AS dias_con_compras,
    MAX(v.fecha) AS fecha_ultima_venta,
    EXTRACT(DAYS FROM NOW() - MAX(v.fecha)) AS dias_sin_comprar,
    -- Métricas de fidelización
    CASE 
        WHEN MAX(v.fecha) >= CURRENT_DATE - INTERVAL '30 days' THEN 'ACTIVO'
        WHEN MAX(v.fecha) >= CURRENT_DATE - INTERVAL '90 days' THEN 'INACTIVO'
        ELSE 'PERDIDO'
    END AS estado_actividad,
    fc.puntos_por_vencer,
    fc.fecha_proximo_vencimiento_puntos
FROM fidelizacion_clientes fc
LEFT JOIN movimientos_fidelizacion mf ON fc.id = mf.cliente_id
LEFT JOIN ventas v ON fc.rut = v.cliente_rut AND v.estado = 'finalizada'
WHERE fc.activo = true
GROUP BY fc.id, fc.rut, fc.nombre, fc.apellido, fc.email, fc.telefono,
         fc.puntos_actuales, fc.puntos_acumulados_total, fc.nivel_fidelizacion,
         fc.fecha_ultima_compra, fc.fecha_ultima_actividad, fc.activo,
         fc.puntos_por_vencer, fc.fecha_proximo_vencimiento_puntos;

-- Índices para fidelización
CREATE UNIQUE INDEX idx_vista_fidelizacion_id ON vista_fidelizacion_resumen_optimizada(id);
CREATE UNIQUE INDEX idx_vista_fidelizacion_rut ON vista_fidelizacion_resumen_optimizada(rut);
CREATE INDEX idx_vista_fidelizacion_nivel ON vista_fidelizacion_resumen_optimizada(nivel_fidelizacion, puntos_actuales DESC);
CREATE INDEX idx_vista_fidelizacion_actividad ON vista_fidelizacion_resumen_optimizada(estado_actividad, fecha_ultima_venta DESC);

-- Vista para documentos DTE pendientes (para api_sync)
CREATE VIEW vista_documentos_dte_pendientes_optimizada AS
SELECT 
    d.id,
    d.sucursal_id,
    s.nombre AS sucursal_nombre,
    s.codigo AS sucursal_codigo,
    d.tipo_documento,
    d.folio,
    d.rut_receptor,
    d.razon_social_receptor,
    d.monto_total,
    d.estado,
    d.fecha_emision,
    d.intentos_envio,
    d.ultimo_error,
    d.fecha_ultimo_intento,
    d.prioridad_procesamiento,
    v.numero_venta,
    -- Campos calculados
    EXTRACT(HOURS FROM NOW() - d.fecha_emision) AS horas_pendiente,
    CASE 
        WHEN d.intentos_envio = 0 THEN 'NUEVO'
        WHEN d.intentos_envio < 3 THEN 'REINTENTO'
        ELSE 'CRITICO'
    END AS estado_procesamiento,
    -- Prioridad calculada
    CASE 
        WHEN d.prioridad_procesamiento <= 2 THEN 'ALTA'
        WHEN d.prioridad_procesamiento <= 5 THEN 'MEDIA'
        ELSE 'BAJA'
    END AS prioridad_texto
FROM documentos_dte d
JOIN sucursales s ON d.sucursal_id = s.id
LEFT JOIN ventas v ON d.venta_id = v.id
WHERE d.estado IN ('pendiente', 'error')
ORDER BY d.prioridad_procesamiento ASC, d.fecha_emision ASC;

-- Vista para métricas de rendimiento de etiquetas (para api_labels)
CREATE VIEW vista_etiquetas_metricas_resumen AS
SELECT 
    DATE(em.fecha_hora) AS fecha,
    em.sucursal_id,
    s.nombre AS sucursal_nombre,
    em.tipo_metrica,
    COUNT(*) AS total_mediciones,
    AVG(em.valor_numerico) AS valor_promedio,
    MIN(em.valor_numerico) AS valor_minimo,
    MAX(em.valor_numerico) AS valor_maximo,
    PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY em.valor_numerico) AS mediana,
    PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY em.valor_numerico) AS percentil_95
FROM etiquetas_metricas_rendimiento em
JOIN sucursales s ON em.sucursal_id = s.id
WHERE em.fecha_hora >= CURRENT_DATE - INTERVAL '30 days'
GROUP BY DATE(em.fecha_hora), em.sucursal_id, s.nombre, em.tipo_metrica
ORDER BY fecha DESC, sucursal_nombre, tipo_metrica;

-- Vista para trabajos de etiquetas activos (para api_labels)
CREATE VIEW vista_etiquetas_trabajos_activos AS
SELECT 
    t.id,
    t.numero_trabajo,
    t.usuario_id,
    u.nombre || ' ' || COALESCE(u.apellido, '') AS usuario_nombre,
    t.sucursal_id,
    s.nombre AS sucursal_nombre,
    t.tipo_trabajo,
    t.estado,
    t.total_etiquetas,
    t.etiquetas_procesadas,
    t.etiquetas_exitosas,
    t.etiquetas_error,
    t.fecha_solicitud,
    t.fecha_inicio_procesamiento,
    t.prioridad,
    -- Campos calculados
    ROUND((t.etiquetas_procesadas::NUMERIC / t.total_etiquetas * 100), 2) AS porcentaje_completado,
    CASE 
        WHEN t.estado = 'pendiente' THEN EXTRACT(MINUTES FROM NOW() - t.fecha_solicitud)
        WHEN t.estado = 'procesando' THEN EXTRACT(MINUTES FROM NOW() - t.fecha_inicio_procesamiento)
        ELSE NULL
    END AS minutos_en_estado,
    CASE 
        WHEN t.estado = 'procesando' AND t.tiempo_procesamiento_ms IS NOT NULL THEN
            (t.tiempo_procesamiento_ms * (t.total_etiquetas - t.etiquetas_procesadas) / GREATEST(t.etiquetas_procesadas, 1))
        ELSE NULL
    END AS tiempo_estimado_restante_ms
FROM etiquetas_trabajos_impresion t
JOIN usuarios u ON t.usuario_id = u.id
JOIN sucursales s ON t.sucursal_id = s.id
WHERE t.estado IN ('pendiente', 'procesando')
ORDER BY t.prioridad ASC, t.fecha_solicitud ASC;

-- =====================================================
-- PROCEDIMIENTOS DE MANTENIMIENTO OPTIMIZADOS
-- =====================================================

-- Procedimiento para refrescar vistas materializadas
CREATE OR REPLACE FUNCTION refrescar_vistas_materializadas()
RETURNS TEXT AS $$
DECLARE
    v_resultado TEXT := '';
    v_inicio TIMESTAMP;
    v_duracion INTERVAL;
BEGIN
    v_inicio := NOW();
    
    -- Refrescar vista de stock (crítica para api_pos)
    REFRESH MATERIALIZED VIEW CONCURRENTLY vista_stock_consolidado_optimizada;
    v_resultado := v_resultado || 'Vista stock consolidado refrescada' || E'\n';
    
    -- Refrescar vista de ventas diarias (para api_report)
    REFRESH MATERIALIZED VIEW CONCURRENTLY vista_ventas_resumen_diario_optimizada;
    v_resultado := v_resultado || 'Vista ventas diarias refrescada' || E'\n';
    
    -- Refrescar vista de productos más vendidos (para api_report)
    REFRESH MATERIALIZED VIEW CONCURRENTLY vista_productos_mas_vendidos_optimizada;
    v_resultado := v_resultado || 'Vista productos más vendidos refrescada' || E'\n';
    
    -- Refrescar vista de fidelización (para api_pos y api_report)
    REFRESH MATERIALIZED VIEW CONCURRENTLY vista_fidelizacion_resumen_optimizada;
    v_resultado := v_resultado || 'Vista fidelización refrescada' || E'\n';
    
    v_duracion := NOW() - v_inicio;
    v_resultado := v_resultado || 'Tiempo total: ' || v_duracion || E'\n';
    
    -- Registrar en logs
    INSERT INTO logs_seguridad (
        evento, nivel_severidad, descripcion, datos_evento
    ) VALUES (
        'refrescar_vistas_materializadas', 'info',
        'Vistas materializadas refrescadas exitosamente',
        jsonb_build_object('duracion_segundos', EXTRACT(EPOCH FROM v_duracion), 'resultado', v_resultado)
    );
    
    RETURN v_resultado;
END;
$$ LANGUAGE plpgsql;

-- Procedimiento de mantenimiento específico para arquitectura centralizada
CREATE OR REPLACE FUNCTION mantenimiento_servidor_central()
RETURNS TEXT AS $$
DECLARE
    v_resultado TEXT := '';
    v_sesiones_eliminadas INTEGER;
    v_cache_limpiado INTEGER;
    v_particiones_creadas INTEGER := 0;
    v_estadisticas_actualizadas INTEGER := 0;
BEGIN
    -- Limpiar sesiones expiradas
    UPDATE sesiones_usuario SET activa = false WHERE fecha_expiracion < NOW() AND activa = true;
    GET DIAGNOSTICS v_sesiones_eliminadas = ROW_COUNT;
    v_resultado := v_resultado || 'Sesiones expiradas: ' || v_sesiones_eliminadas || E'\n';
    
    -- Limpiar cache de reportes
    SELECT limpiar_cache_reportes() INTO v_cache_limpiado;
    v_resultado := v_resultado || 'Cache reportes limpiado: ' || v_cache_limpiado || E'\n';
    
    -- Limpiar cache de códigos de barras expirados
    DELETE FROM etiquetas_cache_codigos_barras WHERE valido_hasta < NOW();
    GET DIAGNOSTICS v_cache_limpiado = ROW_COUNT;
    v_resultado := v_resultado || 'Cache códigos de barras limpiado: ' || v_cache_limpiado || E'\n';
    
    -- Actualizar estadísticas de tablas críticas para api_pos
    ANALYZE usuarios;
    ANALYZE productos;
    ANALYZE stock_central;
    ANALYZE ventas;
    v_estadisticas_actualizadas := 4;
    
    -- Actualizar estadísticas de tablas para api_sync
    ANALYZE logs_sincronizacion;
    ANALYZE documentos_dte;
    v_estadisticas_actualizadas := v_estadisticas_actualizadas + 2;
    
    -- Actualizar estadísticas de tablas para api_labels
    ANALYZE etiquetas_trabajos_impresion;
    ANALYZE etiquetas_cache_codigos_barras;
    v_estadisticas_actualizadas := v_estadisticas_actualizadas + 2;
    
    -- Actualizar estadísticas de tablas para api_report
    ANALYZE cache_reportes;
    v_estadisticas_actualizadas := v_estadisticas_actualizadas + 1;
    
    v_resultado := v_resultado || 'Estadísticas actualizadas: ' || v_estadisticas_actualizadas || ' tablas' || E'\n';
    
    -- Refrescar vistas materializadas críticas
    REFRESH MATERIALIZED VIEW CONCURRENTLY vista_stock_consolidado_optimizada;
    v_resultado := v_resultado || 'Vista stock consolidado refrescada' || E'\n';
    
    -- Crear particiones futuras si es necesario
    -- (En implementación real, aquí iría lógica para crear particiones automáticamente)
    
    -- Registrar ejecución del mantenimiento
    INSERT INTO logs_seguridad (
        evento, nivel_severidad, descripcion, datos_evento
    ) VALUES (
        'mantenimiento_servidor_central', 'info',
        'Mantenimiento del servidor central ejecutado exitosamente',
        jsonb_build_object('resultado', v_resultado)
    );
    
    RETURN v_resultado;
END;
$$ LANGUAGE plpgsql;

-- Procedimiento para monitorear rendimiento de procesos API
CREATE OR REPLACE FUNCTION monitorear_rendimiento_procesos_api()
RETURNS TABLE (
    proceso_api TEXT,
    consultas_activas INTEGER,
    consultas_lentas INTEGER,
    conexiones_activas INTEGER,
    tiempo_promedio_respuesta_ms NUMERIC
) AS $$
BEGIN
    RETURN QUERY
    WITH estadisticas_procesos AS (
        -- Simular estadísticas por proceso (en implementación real usar pg_stat_activity)
        SELECT 'api_pos' as proceso, 45 as consultas, 2 as lentas, 25 as conexiones, 150.5 as tiempo_ms
        UNION ALL
        SELECT 'api_sync' as proceso, 12 as consultas, 1 as lentas, 8 as conexiones, 850.2 as tiempo_ms
        UNION ALL
        SELECT 'api_labels' as proceso, 5 as consultas, 0 as lentas, 3 as conexiones, 2500.0 as tiempo_ms
        UNION ALL
        SELECT 'api_report' as proceso, 8 as consultas, 3 as lentas, 5 as conexiones, 5200.8 as tiempo_ms
    )
    SELECT ep.proceso, ep.consultas, ep.lentas, ep.conexiones, ep.tiempo_ms
    FROM estadisticas_procesos ep
    ORDER BY 
        CASE ep.proceso
            WHEN 'api_pos' THEN 1
            WHEN 'api_sync' THEN 2
            WHEN 'api_labels' THEN 3
            WHEN 'api_report' THEN 4
        END;
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- DATOS INICIALES OPTIMIZADOS PARA ARQUITECTURA CENTRALIZADA
-- =====================================================

-- Configuraciones del sistema optimizadas para Go y arquitectura centralizada
INSERT INTO configuracion_sistema (clave, valor, tipo_dato, descripcion, categoria, modificable, proceso_afectado, cache_ttl_segundos) VALUES
-- Configuraciones generales del sistema
('sistema.version', '2.0.0', 'string', 'Versión del sistema con arquitectura centralizada Go', 'sistema', false, NULL, 3600),
('sistema.nombre', 'Ferre-POS Centralizado', 'string', 'Nombre del sistema', 'sistema', false, NULL, 3600),
('sistema.timezone', 'America/Santiago', 'string', 'Zona horaria del sistema', 'sistema', true, NULL, 300),
('sistema.max_conexiones_concurrentes', '100', 'integer', 'Máximo de conexiones concurrentes al servidor central', 'sistema', true, NULL, 60),

-- Configuraciones específicas para api_pos (máxima prioridad)
('api_pos.pool_conexiones_max', '50', 'integer', 'Máximo conexiones de BD para api_pos', 'api_pos', true, 'maxima', 60),
('api_pos.timeout_consulta_ms', '1000', 'integer', 'Timeout máximo para consultas críticas de POS', 'api_pos', true, 'maxima', 30),
('api_pos.cache_productos_ttl', '300', 'integer', 'TTL del cache de productos en segundos', 'api_pos', true, 'maxima', 60),
('api_pos.cache_stock_ttl', '60', 'integer', 'TTL del cache de stock en segundos', 'api_pos', true, 'maxima', 30),
('api_pos.validacion_stock_estricta', 'true', 'boolean', 'Validación estricta de stock antes de venta', 'api_pos', true, 'maxima', 30),

-- Configuraciones para api_sync (prioridad media)
('api_sync.pool_conexiones_max', '15', 'integer', 'Máximo conexiones de BD para api_sync', 'api_sync', true, 'media', 300),
('api_sync.intervalo_sincronizacion_segundos', '300', 'integer', 'Intervalo entre sincronizaciones automáticas', 'api_sync', true, 'media', 60),
('api_sync.lote_maximo_registros', '1000', 'integer', 'Máximo registros por lote de sincronización', 'api_sync', true, 'media', 300),
('api_sync.reintentos_maximo', '3', 'integer', 'Máximo reintentos para sincronización fallida', 'api_sync', true, 'media', 300),

-- Configuraciones para api_labels (prioridad baja)
('api_labels.pool_conexiones_max', '10', 'integer', 'Máximo conexiones de BD para api_labels', 'api_labels', true, 'baja', 300),
('api_labels.trabajos_concurrentes_max', '3', 'integer', 'Máximo trabajos de etiquetas concurrentes', 'api_labels', true, 'baja', 60),
('api_labels.cache_codigos_barras_ttl', '1800', 'integer', 'TTL del cache de códigos de barras en segundos', 'api_labels', true, 'baja', 300),
('api_labels.timeout_generacion_ms', '10000', 'integer', 'Timeout para generación de etiquetas', 'api_labels', true, 'baja', 300),

-- Configuraciones para api_report (prioridad mínima)
('api_report.pool_conexiones_max', '5', 'integer', 'Máximo conexiones de BD para api_report', 'api_report', true, 'minima', 600),
('api_report.cache_reportes_ttl', '3600', 'integer', 'TTL del cache de reportes en segundos', 'api_report', true, 'minima', 600),
('api_report.timeout_consulta_ms', '30000', 'integer', 'Timeout para consultas de reportes', 'api_report', true, 'minima', 600),
('api_report.reportes_concurrentes_max', '2', 'integer', 'Máximo reportes concurrentes', 'api_report', true, 'minima', 300),

-- Configuraciones de fidelización
('fidelizacion.puntos_por_peso', '1', 'integer', 'Puntos otorgados por cada peso gastado', 'fidelizacion', true, 'maxima', 300),
('fidelizacion.minimo_canje', '100', 'integer', 'Mínimo de puntos para realizar canje', 'fidelizacion', true, 'maxima', 300),
('fidelizacion.expiracion_puntos_dias', '365', 'integer', 'Días para expiración de puntos', 'fidelizacion', true, 'media', 3600),

-- Configuraciones de stock
('stock.alerta_minimo', '5', 'integer', 'Cantidad mínima para alerta de stock', 'stock', true, 'maxima', 300),
('stock.validacion_concurrencia', 'true', 'boolean', 'Usar validación de concurrencia optimista', 'stock', true, 'maxima', 60),

-- Configuraciones de DTE
('dte.ambiente_default', 'produccion', 'string', 'Ambiente por defecto para DTE', 'dte', true, 'media', 3600),
('dte.reintentos_maximo', '3', 'integer', 'Máximo de reintentos para envío DTE', 'dte', true, 'media', 300),
('dte.timeout_proveedor_ms', '15000', 'integer', 'Timeout para proveedores DTE', 'dte', true, 'media', 300),

-- Configuraciones de seguridad
('seguridad.intentos_login_maximo', '5', 'integer', 'Máximo intentos fallidos de login', 'seguridad', true, 'maxima', 300),
('seguridad.bloqueo_minutos', '30', 'integer', 'Minutos de bloqueo tras intentos fallidos', 'seguridad', true, 'maxima', 300),
('seguridad.sesiones_duracion_horas', '8', 'integer', 'Duración de sesiones en horas', 'seguridad', true, 'maxima', 300),
('seguridad.token_refresh_horas', '24', 'integer', 'Duración de refresh tokens en horas', 'seguridad', true, 'maxima', 300),

-- Configuraciones de mantenimiento
('mantenimiento.vistas_materializadas_horas', '4', 'integer', 'Intervalo para refrescar vistas materializadas', 'mantenimiento', true, 'minima', 3600),
('mantenimiento.limpieza_logs_dias', '30', 'integer', 'Días de retención de logs', 'mantenimiento', true, 'minima', 3600),
('mantenimiento.limpieza_cache_horas', '6', 'integer', 'Intervalo para limpieza de cache', 'mantenimiento', true, 'minima', 3600);

-- Categorías de productos optimizadas con configuración para etiquetas
INSERT INTO categorias_productos (codigo, nombre, descripcion, nivel, orden_visualizacion, configuracion_etiquetas) VALUES
('HERR', 'Herramientas', 'Herramientas manuales y eléctricas', 1, 1, '{"plantilla_default": "herramientas", "mostrar_marca": true, "mostrar_modelo": true}'),
('FERRE', 'Ferretería', 'Artículos de ferretería general', 1, 2, '{"plantilla_default": "ferreteria", "mostrar_dimensiones": true}'),
('CONST', 'Construcción', 'Materiales de construcción', 1, 3, '{"plantilla_default": "construccion", "mostrar_peso": true}'),
('ELECT', 'Eléctrico', 'Materiales eléctricos', 1, 4, '{"plantilla_default": "electrico", "mostrar_especificaciones": true}'),
('PLOM', 'Plomería', 'Materiales de plomería', 1, 5, '{"plantilla_default": "plomeria", "mostrar_diametro": true}'),
('PINT', 'Pintura', 'Pinturas y accesorios', 1, 6, '{"plantilla_default": "pintura", "mostrar_color": true, "mostrar_rendimiento": true}'),
('JARD', 'Jardín', 'Herramientas y accesorios de jardín', 1, 7, '{"plantilla_default": "jardin", "mostrar_temporada": true}'),
('SEG', 'Seguridad', 'Elementos de seguridad', 1, 8, '{"plantilla_default": "seguridad", "mostrar_certificacion": true}');

-- Plantillas de etiquetas predeterminadas
INSERT INTO etiquetas_plantillas (codigo, nombre, descripcion, tipo_etiqueta, ancho_mm, alto_mm, orientacion, configuracion_diseno, configuracion_codigo_barras, activa, predeterminada) VALUES
('PRECIO_STD', 'Etiqueta de Precio Estándar', 'Plantilla estándar para etiquetas de precio', 'precio', 50.0, 30.0, 'horizontal', 
 '{"fuente_principal": "Arial", "tamaño_fuente_precio": 14, "tamaño_fuente_descripcion": 10, "color_fondo": "#FFFFFF", "color_texto": "#000000", "mostrar_logo": true}',
 '{"tipo": "CODE39", "altura": 8, "mostrar_texto": true, "posicion": "inferior"}', true, true),

('INVENTARIO_STD', 'Etiqueta de Inventario', 'Plantilla para etiquetas de inventario', 'inventario', 40.0, 25.0, 'horizontal',
 '{"fuente_principal": "Arial", "tamaño_fuente_codigo": 12, "tamaño_fuente_descripcion": 8, "color_fondo": "#F0F0F0", "color_texto": "#000000"}',
 '{"tipo": "CODE39", "altura": 6, "mostrar_texto": true, "posicion": "superior"}', true, false),

('PROMOCION_STD', 'Etiqueta de Promoción', 'Plantilla para etiquetas promocionales', 'promocion', 60.0, 40.0, 'horizontal',
 '{"fuente_principal": "Arial Bold", "tamaño_fuente_precio": 16, "tamaño_fuente_descuento": 12, "color_fondo": "#FF0000", "color_texto": "#FFFFFF", "mostrar_porcentaje_descuento": true}',
 '{"tipo": "CODE39", "altura": 8, "mostrar_texto": false, "posicion": "inferior"}', true, false),

('CODIGO_BARRAS_SIMPLE', 'Código de Barras Simple', 'Solo código de barras con texto', 'codigo_barras', 35.0, 15.0, 'horizontal',
 '{"fuente_principal": "Arial", "tamaño_fuente_codigo": 8, "color_fondo": "#FFFFFF", "color_texto": "#000000", "mostrar_solo_codigo": true}',
 '{"tipo": "CODE39", "altura": 10, "mostrar_texto": true, "posicion": "inferior"}', true, false);

-- Reglas de fidelización optimizadas
INSERT INTO reglas_fidelizacion (nombre, descripcion, tipo_regla, condiciones, acciones, activa, prioridad, condiciones_compiladas) VALUES
(
    'Acumulación Base Optimizada',
    'Regla base de acumulación con cache optimizado',
    'acumulacion',
    '{"monto_minimo": 100, "productos_excluidos": [], "sucursales_incluidas": "todas"}',
    '{"puntos_por_peso": 1, "multiplicador_base": 1.0, "redondeo": "inferior"}',
    true,
    1,
    '{"cache_key": "acumulacion_base", "evaluacion_rapida": true}'
),
(
    'Multiplicador Plata Optimizado',
    'Multiplicador para clientes nivel plata con cache',
    'acumulacion',
    '{"nivel_cliente": "plata", "monto_minimo": 50}',
    '{"multiplicador": 1.2, "bonus_adicional": 0}',
    true,
    2,
    '{"cache_key": "mult_plata", "evaluacion_rapida": true}'
),
(
    'Multiplicador Oro Optimizado',
    'Multiplicador para clientes nivel oro con cache',
    'acumulacion',
    '{"nivel_cliente": "oro", "monto_minimo": 50}',
    '{"multiplicador": 1.5, "bonus_adicional": 10}',
    true,
    3,
    '{"cache_key": "mult_oro", "evaluacion_rapida": true}'
),
(
    'Multiplicador Platino Optimizado',
    'Multiplicador para clientes nivel platino con cache',
    'acumulacion',
    '{"nivel_cliente": "platino", "monto_minimo": 0}',
    '{"multiplicador": 2.0, "bonus_adicional": 25}',
    true,
    4,
    '{"cache_key": "mult_platino", "evaluacion_rapida": true}'
);

-- Configuraciones de impresoras de etiquetas de ejemplo
INSERT INTO etiquetas_configuraciones_impresora (codigo, nombre, marca, modelo, tipo_conexion, parametros_conexion, configuracion_driver, activa, estado_impresora) VALUES
('ZEBRA_001', 'Zebra Principal Sucursal Centro', 'Zebra', 'ZD420', 'usb', 
 '{"puerto": "/dev/usb/lp0", "velocidad": "9600"}',
 '{"dpi": 203, "velocidad_impresion": "2", "densidad": "8", "modo_transferencia": "directo"}',
 true, 'disponible'),

('BROTHER_001', 'Brother Etiquetas Sucursal Norte', 'Brother', 'QL-820NWB', 'red',
 '{"ip": "192.168.1.100", "puerto": 9100}',
 '{"dpi": 300, "velocidad_impresion": "3", "calidad": "alta", "modo_corte": "automatico"}',
 true, 'disponible');

-- =====================================================
-- CONFIGURACIONES FINALES DE POSTGRESQL OPTIMIZADAS
-- =====================================================

-- Configurar autovacuum optimizado para arquitectura centralizada
ALTER SYSTEM SET autovacuum = on;
ALTER SYSTEM SET autovacuum_max_workers = 4;
ALTER SYSTEM SET autovacuum_naptime = '30s';
ALTER SYSTEM SET autovacuum_vacuum_threshold = 50;
ALTER SYSTEM SET autovacuum_analyze_threshold = 50;
ALTER SYSTEM SET autovacuum_vacuum_scale_factor = 0.1;
ALTER SYSTEM SET autovacuum_analyze_scale_factor = 0.05;

-- Configurar parámetros específicos para tablas críticas de api_pos
ALTER TABLE usuarios SET (autovacuum_vacuum_scale_factor = 0.05);
ALTER TABLE productos SET (autovacuum_vacuum_scale_factor = 0.05);
ALTER TABLE stock_central SET (autovacuum_vacuum_scale_factor = 0.02);
ALTER TABLE ventas SET (autovacuum_vacuum_scale_factor = 0.1);
ALTER TABLE sesiones_usuario SET (autovacuum_vacuum_scale_factor = 0.02);

-- Configurar parámetros para tablas de logs (crecimiento rápido)
ALTER TABLE logs_seguridad SET (autovacuum_vacuum_scale_factor = 0.2);
ALTER TABLE logs_sincronizacion SET (autovacuum_vacuum_scale_factor = 0.2);
ALTER TABLE movimientos_stock SET (autovacuum_vacuum_scale_factor = 0.2);

-- Configurar fill factor optimizado para tablas con muchas actualizaciones
ALTER TABLE stock_central SET (fillfactor = 80);
ALTER TABLE usuarios SET (fillfactor = 85);
ALTER TABLE sesiones_usuario SET (fillfactor = 70);
ALTER TABLE configuracion_sistema SET (fillfactor = 90);

-- =====================================================
-- ROLES Y PERMISOS OPTIMIZADOS PARA PROCESOS API
-- =====================================================

-- Crear roles específicos para cada proceso API
CREATE ROLE ferre_pos_api_pos;
CREATE ROLE ferre_pos_api_sync;
CREATE ROLE ferre_pos_api_labels;
CREATE ROLE ferre_pos_api_report;
CREATE ROLE ferre_pos_readonly;
CREATE ROLE ferre_pos_admin;

-- Permisos para api_pos (acceso completo a tablas críticas)
GRANT CONNECT ON DATABASE postgres TO ferre_pos_api_pos;
GRANT USAGE ON SCHEMA public TO ferre_pos_api_pos;
GRANT SELECT, INSERT, UPDATE, DELETE ON usuarios, productos, stock_central, ventas, detalle_ventas, 
      medios_pago_venta, notas_venta, detalle_notas_venta, fidelizacion_clientes, 
      movimientos_fidelizacion, sesiones_usuario, terminales, sucursales TO ferre_pos_api_pos;
GRANT SELECT ON categorias_productos, codigos_barra_adicionales, reglas_fidelizacion, 
      configuracion_sistema TO ferre_pos_api_pos;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO ferre_pos_api_pos;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO ferre_pos_api_pos; -- Para vistas

-- Permisos para api_sync (acceso a sincronización y DTE)
GRANT CONNECT ON DATABASE postgres TO ferre_pos_api_sync;
GRANT USAGE ON SCHEMA public TO ferre_pos_api_sync;
GRANT SELECT, INSERT, UPDATE ON documentos_dte, folios_dte, logs_sincronizacion, 
      proveedores_dte, configuracion_dte_sucursal TO ferre_pos_api_sync;
GRANT SELECT ON ventas, detalle_ventas, productos, stock_central, movimientos_stock,
      sucursales, usuarios TO ferre_pos_api_sync;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO ferre_pos_api_sync;

-- Permisos para api_labels (acceso a módulo de etiquetas)
GRANT CONNECT ON DATABASE postgres TO ferre_pos_api_labels;
GRANT USAGE ON SCHEMA public TO ferre_pos_api_labels;
GRANT SELECT, INSERT, UPDATE, DELETE ON etiquetas_plantillas, etiquetas_trabajos_impresion,
      etiquetas_detalle_trabajo, etiquetas_configuraciones_impresora, etiquetas_cache_codigos_barras,
      etiquetas_metricas_rendimiento TO ferre_pos_api_labels;
GRANT SELECT ON productos, categorias_productos, stock_central, usuarios, sucursales, 
      terminales TO ferre_pos_api_labels;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO ferre_pos_api_labels;

-- Permisos para api_report (acceso de solo lectura + cache)
GRANT CONNECT ON DATABASE postgres TO ferre_pos_api_report;
GRANT USAGE ON SCHEMA public TO ferre_pos_api_report;
GRANT SELECT, INSERT, UPDATE, DELETE ON cache_reportes, reportes_programados TO ferre_pos_api_report;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO ferre_pos_api_report;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO ferre_pos_api_report;

-- Permisos para solo lectura
GRANT CONNECT ON DATABASE postgres TO ferre_pos_readonly;
GRANT USAGE ON SCHEMA public TO ferre_pos_readonly;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO ferre_pos_readonly;

-- Permisos para administración
GRANT ALL PRIVILEGES ON DATABASE postgres TO ferre_pos_admin;

-- =====================================================
-- JOBS PROGRAMADOS PARA MANTENIMIENTO AUTOMÁTICO
-- =====================================================

-- Crear extensión para jobs programados si está disponible
-- CREATE EXTENSION IF NOT EXISTS pg_cron;

-- Programar mantenimiento diario (comentado - requiere pg_cron)
-- SELECT cron.schedule('mantenimiento-diario', '0 2 * * *', 'SELECT mantenimiento_servidor_central();');

-- Programar refrescado de vistas materializadas cada 4 horas
-- SELECT cron.schedule('refrescar-vistas', '0 */4 * * *', 'SELECT refrescar_vistas_materializadas();');

-- Programar limpieza de cache cada 6 horas
-- SELECT cron.schedule('limpiar-cache', '0 */6 * * *', 'SELECT limpiar_cache_reportes();');

-- =====================================================
-- FINALIZACIÓN DEL SCRIPT OPTIMIZADO
-- =====================================================

-- Función para validar integridad del esquema
CREATE OR REPLACE FUNCTION validar_integridad_esquema()
RETURNS TABLE (
    categoria TEXT,
    elemento TEXT,
    estado TEXT,
    detalles TEXT
) AS $$
BEGIN
    -- Validar tablas críticas para api_pos
    RETURN QUERY
    SELECT 'api_pos'::TEXT, 'tabla_usuarios'::TEXT, 
           CASE WHEN EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'usuarios') 
                THEN 'OK' ELSE 'ERROR' END::TEXT,
           'Tabla de usuarios para autenticación'::TEXT;
    
    RETURN QUERY
    SELECT 'api_pos'::TEXT, 'tabla_productos'::TEXT,
           CASE WHEN EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'productos')
                THEN 'OK' ELSE 'ERROR' END::TEXT,
           'Tabla de productos para POS'::TEXT;
    
    RETURN QUERY
    SELECT 'api_pos'::TEXT, 'tabla_stock_central'::TEXT,
           CASE WHEN EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'stock_central')
                THEN 'OK' ELSE 'ERROR' END::TEXT,
           'Tabla de stock centralizado'::TEXT;
    
    -- Validar índices críticos
    RETURN QUERY
    SELECT 'indices'::TEXT, 'idx_productos_codigo_barra_activo'::TEXT,
           CASE WHEN EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_productos_codigo_barra_activo')
                THEN 'OK' ELSE 'ERROR' END::TEXT,
           'Índice crítico para búsqueda de productos'::TEXT;
    
    -- Validar vistas materializadas
    RETURN QUERY
    SELECT 'vistas'::TEXT, 'vista_stock_consolidado_optimizada'::TEXT,
           CASE WHEN EXISTS (SELECT 1 FROM pg_matviews WHERE matviewname = 'vista_stock_consolidado_optimizada')
                THEN 'OK' ELSE 'ERROR' END::TEXT,
           'Vista materializada de stock'::TEXT;
    
    -- Validar funciones críticas
    RETURN QUERY
    SELECT 'funciones'::TEXT, 'validar_stock_disponible_optimizado'::TEXT,
           CASE WHEN EXISTS (SELECT 1 FROM pg_proc WHERE proname = 'validar_stock_disponible_optimizado')
                THEN 'OK' ELSE 'ERROR' END::TEXT,
           'Función optimizada de validación de stock'::TEXT;
    
    -- Validar configuraciones
    RETURN QUERY
    SELECT 'configuracion'::TEXT, 'configuraciones_api_pos'::TEXT,
           CASE WHEN EXISTS (SELECT 1 FROM configuracion_sistema WHERE proceso_afectado = 'maxima')
                THEN 'OK' ELSE 'ERROR' END::TEXT,
           'Configuraciones específicas para api_pos'::TEXT;
END;
$$ LANGUAGE plpgsql;

-- Mensaje de finalización con validación
DO $$
DECLARE
    v_validacion RECORD;
    v_errores INTEGER := 0;
BEGIN
    RAISE NOTICE '=======================================================';
    RAISE NOTICE 'ESQUEMA FERRE-POS SERVIDOR CENTRAL OPTIMIZADO CREADO';
    RAISE NOTICE '=======================================================';
    RAISE NOTICE 'Versión: 2.0 - Arquitectura Centralizada Go';
    RAISE NOTICE 'Fecha: %', NOW();
    RAISE NOTICE 'Optimizado para 4 procesos API diferenciados:';
    RAISE NOTICE '  - api_pos (máxima prioridad): % conexiones máx', 
        (SELECT valor FROM configuracion_sistema WHERE clave = 'api_pos.pool_conexiones_max');
    RAISE NOTICE '  - api_sync (prioridad media): % conexiones máx', 
        (SELECT valor FROM configuracion_sistema WHERE clave = 'api_sync.pool_conexiones_max');
    RAISE NOTICE '  - api_labels (prioridad baja): % conexiones máx', 
        (SELECT valor FROM configuracion_sistema WHERE clave = 'api_labels.pool_conexiones_max');
    RAISE NOTICE '  - api_report (prioridad mínima): % conexiones máx', 
        (SELECT valor FROM configuracion_sistema WHERE clave = 'api_report.pool_conexiones_max');
    RAISE NOTICE '=======================================================';
    RAISE NOTICE 'ESTADÍSTICAS DEL ESQUEMA:';
    RAISE NOTICE 'Tablas creadas: %', (
        SELECT COUNT(*) FROM information_schema.tables 
        WHERE table_schema = 'public' AND table_type = 'BASE TABLE'
    );
    RAISE NOTICE 'Índices creados: %', (
        SELECT COUNT(*) FROM pg_indexes WHERE schemaname = 'public'
    );
    RAISE NOTICE 'Funciones creadas: %', (
        SELECT COUNT(*) FROM information_schema.routines 
        WHERE routine_schema = 'public' AND routine_type = 'FUNCTION'
    );
    RAISE NOTICE 'Vistas materializadas: %', (
        SELECT COUNT(*) FROM pg_matviews WHERE schemaname = 'public'
    );
    RAISE NOTICE 'Particiones creadas: %', (
        SELECT COUNT(*) FROM pg_tables WHERE schemaname = 'public' AND tablename LIKE '%_202%'
    );
    RAISE NOTICE '=======================================================';
    RAISE NOTICE 'VALIDACIÓN DE INTEGRIDAD:';
    
    FOR v_validacion IN SELECT * FROM validar_integridad_esquema() LOOP
        IF v_validacion.estado = 'ERROR' THEN
            v_errores := v_errores + 1;
            RAISE NOTICE '[ERROR] %: % - %', v_validacion.categoria, v_validacion.elemento, v_validacion.detalles;
        ELSE
            RAISE NOTICE '[OK] %: %', v_validacion.categoria, v_validacion.elemento;
        END IF;
    END LOOP;
    
    IF v_errores > 0 THEN
        RAISE NOTICE '=======================================================';
        RAISE NOTICE 'ATENCIÓN: Se encontraron % errores en la validación', v_errores;
    ELSE
        RAISE NOTICE 'Validación completada exitosamente - Sin errores';
    END IF;
    
    RAISE NOTICE '=======================================================';
    RAISE NOTICE 'PRÓXIMOS PASOS PARA IMPLEMENTACIÓN:';
    RAISE NOTICE '1. Configurar postgresql.conf según arquitectura centralizada';
    RAISE NOTICE '2. Crear usuarios de aplicación con roles específicos por proceso API';
    RAISE NOTICE '3. Configurar pools de conexiones diferenciados en aplicación Go';
    RAISE NOTICE '4. Implementar monitoreo de rendimiento por proceso API';
    RAISE NOTICE '5. Programar jobs de mantenimiento automático';
    RAISE NOTICE '6. Configurar respaldos optimizados para arquitectura centralizada';
    RAISE NOTICE '7. Cargar datos iniciales de sucursales y productos';
    RAISE NOTICE '8. Configurar plantillas de etiquetas específicas por categoría';
    RAISE NOTICE '9. Probar rendimiento con carga simulada por proceso API';
    RAISE NOTICE '10. Implementar alertas de rendimiento y disponibilidad';
    RAISE NOTICE '=======================================================';
    RAISE NOTICE 'CONFIGURACIONES RECOMENDADAS POSTGRESQL.CONF:';
    RAISE NOTICE 'shared_buffers = 512MB (para servidor central)';
    RAISE NOTICE 'effective_cache_size = 2GB';
    RAISE NOTICE 'work_mem = 8MB';
    RAISE NOTICE 'maintenance_work_mem = 128MB';
    RAISE NOTICE 'max_connections = 100';
    RAISE NOTICE 'checkpoint_completion_target = 0.9';
    RAISE NOTICE 'wal_buffers = 32MB';
    RAISE NOTICE 'random_page_cost = 1.1 (para SSD)';
    RAISE NOTICE '=======================================================';
END $$;

-- Fin del script optimizado
-- =====================================================

