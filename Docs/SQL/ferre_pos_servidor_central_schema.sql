-- =====================================================
-- ESQUEMA SQL SERVIDOR CENTRAL - SISTEMA FERRE-POS
-- =====================================================
-- Versión: 1.0
-- Fecha: Julio 2025
-- Autor: Manus AI
-- Descripción: Esquema completo para el servidor central del sistema Ferre-POS
--              basado en la documentación técnica unificada
-- =====================================================

-- Configuración inicial de PostgreSQL
SET timezone = 'America/Santiago';
SET default_tablespace = '';
SET default_table_access_method = heap;

-- Extensiones necesarias
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";
CREATE EXTENSION IF NOT EXISTS "btree_gin";

-- =====================================================
-- TIPOS DE DATOS PERSONALIZADOS
-- =====================================================

-- Tipo para roles de usuario
CREATE TYPE rol_usuario AS ENUM (
    'cajero', 
    'vendedor', 
    'despacho', 
    'supervisor', 
    'admin'
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

-- =====================================================
-- TABLAS PRINCIPALES
-- =====================================================



-- Tabla: sucursales
-- Descripción: Almacena información de todas las sucursales de la cadena
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
    CONSTRAINT chk_horarios CHECK (horario_apertura < horario_cierre)
);

-- Tabla: usuarios
-- Descripción: Gestión de usuarios del sistema con roles y permisos
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
    CONSTRAINT chk_rut_formato CHECK (rut ~ '^[0-9]{7,8}-[0-9Kk]$'),
    CONSTRAINT chk_email_formato CHECK (email ~ '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$')
);

-- Tabla: categorias_productos
-- Descripción: Categorización jerárquica de productos
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
    CONSTRAINT chk_nivel_positivo CHECK (nivel > 0)
);

-- Tabla: productos
-- Descripción: Catálogo maestro de productos
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
    CONSTRAINT chk_precio_positivo CHECK (precio_unitario >= 0),
    CONSTRAINT chk_precio_costo_positivo CHECK (precio_costo IS NULL OR precio_costo >= 0),
    CONSTRAINT chk_stock_minimo CHECK (stock_minimo >= 0),
    CONSTRAINT chk_stock_maximo CHECK (stock_maximo IS NULL OR stock_maximo >= stock_minimo)
);

-- Tabla: codigos_barra_adicionales
-- Descripción: Códigos de barra alternativos para productos
CREATE TABLE codigos_barra_adicionales (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    producto_id UUID REFERENCES productos(id) ON DELETE CASCADE,
    codigo_barra TEXT NOT NULL,
    descripcion TEXT,
    activo BOOLEAN DEFAULT true,
    fecha_creacion TIMESTAMP DEFAULT NOW(),
    UNIQUE(codigo_barra)
);

-- Tabla: stock_central
-- Descripción: Consolidación de stock de todas las sucursales
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
    PRIMARY KEY (producto_id, sucursal_id),
    CONSTRAINT chk_cantidad_positiva CHECK (cantidad >= 0),
    CONSTRAINT chk_reservada_positiva CHECK (cantidad_reservada >= 0),
    CONSTRAINT chk_reservada_menor_cantidad CHECK (cantidad_reservada <= cantidad)
);

-- Tabla: movimientos_stock
-- Descripción: Historial de todos los movimientos de inventario
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
    CONSTRAINT chk_tipo_movimiento CHECK (tipo_movimiento IN (
        'entrada', 'salida', 'ajuste', 'transferencia_entrada', 
        'transferencia_salida', 'venta', 'devolucion'
    ))
);

-- Tabla: terminales
-- Descripción: Registro de terminales POS y sus configuraciones
CREATE TABLE terminales (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    codigo TEXT UNIQUE NOT NULL,
    nombre_terminal TEXT NOT NULL,
    tipo_terminal TEXT NOT NULL,
    sucursal_id UUID REFERENCES sucursales(id),
    direccion_ip TEXT,
    direccion_mac TEXT,
    activo BOOLEAN DEFAULT true,
    ultima_conexion TIMESTAMP,
    version_software TEXT,
    configuracion JSONB,
    fecha_instalacion TIMESTAMP DEFAULT NOW(),
    fecha_modificacion TIMESTAMP DEFAULT NOW(),
    CONSTRAINT chk_tipo_terminal CHECK (tipo_terminal IN (
        'caja', 'tienda', 'despacho', 'autoatencion'
    ))
);

-- Tabla: ventas
-- Descripción: Registro maestro de todas las ventas
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
);

-- Tabla: detalle_ventas
-- Descripción: Detalle de productos vendidos en cada venta
CREATE TABLE detalle_ventas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    venta_id UUID REFERENCES ventas(id) ON DELETE CASCADE,
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
);

-- Tabla: medios_pago_venta
-- Descripción: Medios de pago utilizados en cada venta
CREATE TABLE medios_pago_venta (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    venta_id UUID REFERENCES ventas(id) ON DELETE CASCADE,
    medio_pago TEXT NOT NULL,
    monto NUMERIC(12,2) NOT NULL,
    referencia_transaccion TEXT,
    codigo_autorizacion TEXT,
    datos_transaccion JSONB,
    fecha_procesamiento TIMESTAMP DEFAULT NOW(),
    CONSTRAINT chk_medio_pago CHECK (medio_pago IN (
        'efectivo', 'tarjeta_debito', 'tarjeta_credito', 
        'transferencia', 'cheque', 'puntos_fidelizacion', 'otro'
    )),
    CONSTRAINT chk_monto_positivo CHECK (monto > 0)
);

-- Tabla: notas_venta
-- Descripción: Notas de venta generadas en POS Tienda
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
    venta_id UUID REFERENCES ventas(id),
    fecha_pago TIMESTAMP,
    sincronizada BOOLEAN DEFAULT false,
    observaciones TEXT,
    CONSTRAINT chk_estado_nota CHECK (estado IN (
        'pendiente', 'pagada', 'cancelada', 'vencida'
    )),
    CONSTRAINT chk_totales_nota_positivos CHECK (
        subtotal >= 0 AND descuento_total >= 0 AND total >= 0
    )
);

-- Tabla: detalle_notas_venta
-- Descripción: Detalle de productos en notas de venta
CREATE TABLE detalle_notas_venta (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    nota_venta_id UUID REFERENCES notas_venta(id) ON DELETE CASCADE,
    producto_id UUID REFERENCES productos(id),
    cantidad NUMERIC(10,3) NOT NULL,
    precio_unitario NUMERIC(12,2) NOT NULL,
    descuento_unitario NUMERIC(12,2) DEFAULT 0,
    total_item NUMERIC(12,2) NOT NULL,
    observaciones TEXT,
    CONSTRAINT chk_cantidad_nota_positiva CHECK (cantidad > 0),
    CONSTRAINT chk_precios_nota_positivos CHECK (
        precio_unitario >= 0 AND descuento_unitario >= 0 AND total_item >= 0
    )
);


-- Tabla: proveedores_dte
-- Descripción: Proveedores autorizados de documentos tributarios electrónicos
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
    fecha_modificacion TIMESTAMP DEFAULT NOW()
);

-- Tabla: configuracion_dte_sucursal
-- Descripción: Configuración de DTE por sucursal
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
    CONSTRAINT chk_ambiente CHECK (ambiente IN ('certificacion', 'produccion'))
);

-- Tabla: documentos_dte
-- Descripción: Registro de documentos tributarios electrónicos emitidos
CREATE TABLE documentos_dte (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sucursal_id UUID REFERENCES sucursales(id),
    proveedor_dte_id UUID REFERENCES proveedores_dte(id),
    venta_id UUID REFERENCES ventas(id),
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
    CONSTRAINT chk_tipo_dte CHECK (tipo_documento IN (
        'boleta_electronica', 'factura_electronica', 'guia_despacho_electronica',
        'nota_credito_electronica', 'nota_debito_electronica'
    )),
    CONSTRAINT chk_folio_positivo CHECK (folio > 0),
    CONSTRAINT chk_montos_positivos CHECK (
        monto_neto >= 0 AND monto_iva >= 0 AND monto_total >= 0
    ),
    UNIQUE(sucursal_id, tipo_documento, folio)
);

-- Tabla: folios_dte
-- Descripción: Control de folios por sucursal y tipo de documento
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
    CONSTRAINT chk_folios_coherentes CHECK (
        folio_desde <= folio_actual AND folio_actual <= folio_hasta
    ),
    UNIQUE(sucursal_id, tipo_documento, folio_desde)
);

-- Tabla: fidelizacion_clientes
-- Descripción: Clientes registrados en el programa de fidelización
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
    CONSTRAINT chk_rut_cliente_formato CHECK (rut ~ '^[0-9]{7,8}-[0-9Kk]$'),
    CONSTRAINT chk_puntos_positivos CHECK (
        puntos_actuales >= 0 AND puntos_acumulados_total >= 0
    ),
    CONSTRAINT chk_nivel_fidelizacion CHECK (nivel_fidelizacion IN (
        'bronce', 'plata', 'oro', 'platino'
    ))
);

-- Tabla: movimientos_fidelizacion
-- Descripción: Historial de movimientos de puntos de fidelización
CREATE TABLE movimientos_fidelizacion (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cliente_id UUID REFERENCES fidelizacion_clientes(id),
    sucursal_id UUID REFERENCES sucursales(id),
    venta_id UUID REFERENCES ventas(id),
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
    CONSTRAINT chk_puntos_movimiento CHECK (puntos != 0)
);

-- Tabla: reglas_fidelizacion
-- Descripción: Reglas de acumulación y canje de puntos
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
    CONSTRAINT chk_tipo_regla CHECK (tipo_regla IN (
        'acumulacion', 'canje', 'promocion', 'nivel'
    ))
);

-- Tabla: notas_credito
-- Descripción: Notas de crédito emitidas
CREATE TABLE notas_credito (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    numero_nota BIGSERIAL UNIQUE,
    sucursal_id UUID REFERENCES sucursales(id),
    documento_origen_id UUID REFERENCES documentos_dte(id),
    venta_origen_id UUID REFERENCES ventas(id),
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
    dte_id UUID REFERENCES documentos_dte(id),
    observaciones TEXT,
    datos_adicionales JSONB,
    CONSTRAINT chk_totales_nc_positivos CHECK (
        subtotal >= 0 AND descuento_total >= 0 AND 
        impuesto_total >= 0 AND total >= 0
    )
);

-- Tabla: detalle_notas_credito
-- Descripción: Detalle de productos en notas de crédito
CREATE TABLE detalle_notas_credito (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    nota_credito_id UUID REFERENCES notas_credito(id) ON DELETE CASCADE,
    producto_id UUID REFERENCES productos(id),
    cantidad NUMERIC(10,3) NOT NULL,
    precio_unitario NUMERIC(12,2) NOT NULL,
    total_item NUMERIC(12,2) NOT NULL,
    motivo_item TEXT,
    CONSTRAINT chk_cantidad_nc_positiva CHECK (cantidad > 0),
    CONSTRAINT chk_precios_nc_positivos CHECK (
        precio_unitario >= 0 AND total_item >= 0
    )
);

-- Tabla: despachos
-- Descripción: Control de despachos de mercadería
CREATE TABLE despachos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    numero_despacho BIGSERIAL UNIQUE,
    sucursal_id UUID REFERENCES sucursales(id),
    documento_id UUID REFERENCES documentos_dte(id),
    venta_id UUID REFERENCES ventas(id),
    usuario_despacho_id UUID REFERENCES usuarios(id),
    cliente_rut TEXT,
    cliente_nombre TEXT,
    estado TEXT DEFAULT 'pendiente',
    fecha_programada TIMESTAMP,
    fecha_inicio TIMESTAMP,
    fecha_completado TIMESTAMP,
    observaciones TEXT,
    datos_adicionales JSONB,
    CONSTRAINT chk_estado_despacho CHECK (estado IN (
        'pendiente', 'en_proceso', 'completo', 'parcial', 'rechazado'
    ))
);

-- Tabla: detalle_despacho
-- Descripción: Detalle de productos despachados
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
    CONSTRAINT chk_cantidades_despacho_positivas CHECK (
        cantidad_solicitada > 0 AND cantidad_despachada >= 0
    ),
    CONSTRAINT chk_cantidad_despachada_valida CHECK (
        cantidad_despachada <= cantidad_solicitada
    )
);

-- Tabla: reimpresiones_documentos
-- Descripción: Registro de reimpresiones de documentos
CREATE TABLE reimpresiones_documentos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    documento_id UUID REFERENCES documentos_dte(id),
    usuario_id UUID REFERENCES usuarios(id),
    sucursal_id UUID REFERENCES sucursales(id),
    tipo_documento TEXT NOT NULL,
    motivo TEXT NOT NULL,
    fecha TIMESTAMP DEFAULT NOW(),
    ip_origen INET,
    dispositivo TEXT,
    reimpresiones_previas INTEGER DEFAULT 0,
    autorizado_por UUID REFERENCES usuarios(id),
    datos_adicionales JSONB
);

-- Tabla: logs_sincronizacion
-- Descripción: Registro de sincronizaciones entre sucursales y servidor central
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
    CONSTRAINT chk_tipo_sync CHECK (tipo_sincronizacion IN (
        'ventas', 'stock', 'productos', 'clientes', 'configuracion', 'ping'
    )),
    CONSTRAINT chk_registros_positivos CHECK (
        registros_enviados >= 0 AND registros_procesados >= 0 AND registros_error >= 0
    )
);

-- Tabla: logs_seguridad
-- Descripción: Registro de eventos de seguridad del sistema
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
    CONSTRAINT chk_nivel_severidad CHECK (nivel_severidad IN (
        'info', 'warning', 'error', 'critical'
    ))
);

-- Tabla: configuracion_sistema
-- Descripción: Configuraciones globales del sistema
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
    CONSTRAINT chk_tipo_dato CHECK (tipo_dato IN (
        'string', 'integer', 'decimal', 'boolean', 'json', 'date', 'timestamp'
    ))
);

-- Tabla: sesiones_usuario
-- Descripción: Control de sesiones activas de usuarios
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
    datos_sesion JSONB
);


-- =====================================================
-- ÍNDICES PARA OPTIMIZACIÓN DE CONSULTAS
-- =====================================================

-- Índices para tabla sucursales
CREATE INDEX idx_sucursales_codigo ON sucursales(codigo);
CREATE INDEX idx_sucursales_habilitada ON sucursales(habilitada);
CREATE INDEX idx_sucursales_region ON sucursales(region);

-- Índices para tabla usuarios
CREATE INDEX idx_usuarios_rut ON usuarios(rut);
CREATE INDEX idx_usuarios_email ON usuarios(email);
CREATE INDEX idx_usuarios_rol ON usuarios(rol);
CREATE INDEX idx_usuarios_sucursal ON usuarios(sucursal_id);
CREATE INDEX idx_usuarios_activo ON usuarios(activo);
CREATE INDEX idx_usuarios_ultimo_acceso ON usuarios(ultimo_acceso);

-- Índices para tabla productos
CREATE INDEX idx_productos_codigo_interno ON productos(codigo_interno);
CREATE INDEX idx_productos_codigo_barra ON productos(codigo_barra);
CREATE INDEX idx_productos_categoria ON productos(categoria_id);
CREATE INDEX idx_productos_activo ON productos(activo);
CREATE INDEX idx_productos_precio ON productos(precio_unitario);
CREATE INDEX idx_productos_marca ON productos(marca);
-- Índice GIN para búsqueda de texto completo en descripción
CREATE INDEX idx_productos_descripcion_trgm ON productos USING GIN (descripcion gin_trgm_ops);
CREATE INDEX idx_productos_descripcion_corta_trgm ON productos USING GIN (descripcion_corta gin_trgm_ops);

-- Índices para tabla categorias_productos
CREATE INDEX idx_categorias_codigo ON categorias_productos(codigo);
CREATE INDEX idx_categorias_padre ON categorias_productos(categoria_padre_id);
CREATE INDEX idx_categorias_nivel ON categorias_productos(nivel);
CREATE INDEX idx_categorias_activa ON categorias_productos(activa);

-- Índices para tabla codigos_barra_adicionales
CREATE INDEX idx_codigos_adicionales_producto ON codigos_barra_adicionales(producto_id);
CREATE INDEX idx_codigos_adicionales_codigo ON codigos_barra_adicionales(codigo_barra);

-- Índices para tabla stock_central
CREATE INDEX idx_stock_central_sucursal ON stock_central(sucursal_id);
CREATE INDEX idx_stock_central_producto ON stock_central(producto_id);
CREATE INDEX idx_stock_central_cantidad ON stock_central(cantidad);
CREATE INDEX idx_stock_central_disponible ON stock_central(cantidad_disponible);
CREATE INDEX idx_stock_central_sync ON stock_central(fecha_sync);

-- Índices para tabla movimientos_stock
CREATE INDEX idx_movimientos_stock_producto ON movimientos_stock(producto_id);
CREATE INDEX idx_movimientos_stock_sucursal ON movimientos_stock(sucursal_id);
CREATE INDEX idx_movimientos_stock_tipo ON movimientos_stock(tipo_movimiento);
CREATE INDEX idx_movimientos_stock_fecha ON movimientos_stock(fecha);
CREATE INDEX idx_movimientos_stock_usuario ON movimientos_stock(usuario_id);
CREATE INDEX idx_movimientos_stock_documento ON movimientos_stock(documento_referencia);

-- Índices para tabla terminales
CREATE INDEX idx_terminales_codigo ON terminales(codigo);
CREATE INDEX idx_terminales_sucursal ON terminales(sucursal_id);
CREATE INDEX idx_terminales_tipo ON terminales(tipo_terminal);
CREATE INDEX idx_terminales_activo ON terminales(activo);
CREATE INDEX idx_terminales_ultima_conexion ON terminales(ultima_conexion);

-- Índices para tabla ventas
CREATE INDEX idx_ventas_numero ON ventas(numero_venta);
CREATE INDEX idx_ventas_sucursal ON ventas(sucursal_id);
CREATE INDEX idx_ventas_terminal ON ventas(terminal_id);
CREATE INDEX idx_ventas_cajero ON ventas(cajero_id);
CREATE INDEX idx_ventas_vendedor ON ventas(vendedor_id);
CREATE INDEX idx_ventas_cliente ON ventas(cliente_rut);
CREATE INDEX idx_ventas_fecha ON ventas(fecha);
CREATE INDEX idx_ventas_tipo_documento ON ventas(tipo_documento);
CREATE INDEX idx_ventas_estado ON ventas(estado);
CREATE INDEX idx_ventas_total ON ventas(total);
CREATE INDEX idx_ventas_dte_emitido ON ventas(dte_emitido);
CREATE INDEX idx_ventas_sincronizada ON ventas(sincronizada);
-- Índice compuesto para reportes por período y sucursal
CREATE INDEX idx_ventas_fecha_sucursal ON ventas(fecha, sucursal_id);
CREATE INDEX idx_ventas_fecha_cajero ON ventas(fecha, cajero_id);

-- Índices para tabla detalle_ventas
CREATE INDEX idx_detalle_ventas_venta ON detalle_ventas(venta_id);
CREATE INDEX idx_detalle_ventas_producto ON detalle_ventas(producto_id);
CREATE INDEX idx_detalle_ventas_numero_serie ON detalle_ventas(numero_serie);

-- Índices para tabla medios_pago_venta
CREATE INDEX idx_medios_pago_venta ON medios_pago_venta(venta_id);
CREATE INDEX idx_medios_pago_tipo ON medios_pago_venta(medio_pago);
CREATE INDEX idx_medios_pago_referencia ON medios_pago_venta(referencia_transaccion);

-- Índices para tabla notas_venta
CREATE INDEX idx_notas_venta_numero ON notas_venta(numero_nota);
CREATE INDEX idx_notas_venta_sucursal ON notas_venta(sucursal_id);
CREATE INDEX idx_notas_venta_vendedor ON notas_venta(vendedor_id);
CREATE INDEX idx_notas_venta_cliente ON notas_venta(cliente_rut);
CREATE INDEX idx_notas_venta_estado ON notas_venta(estado);
CREATE INDEX idx_notas_venta_fecha ON notas_venta(fecha);
CREATE INDEX idx_notas_venta_sincronizada ON notas_venta(sincronizada);

-- Índices para tabla detalle_notas_venta
CREATE INDEX idx_detalle_notas_venta_nota ON detalle_notas_venta(nota_venta_id);
CREATE INDEX idx_detalle_notas_venta_producto ON detalle_notas_venta(producto_id);

-- Índices para tabla documentos_dte
CREATE INDEX idx_documentos_dte_sucursal ON documentos_dte(sucursal_id);
CREATE INDEX idx_documentos_dte_venta ON documentos_dte(venta_id);
CREATE INDEX idx_documentos_dte_tipo ON documentos_dte(tipo_documento);
CREATE INDEX idx_documentos_dte_folio ON documentos_dte(folio);
CREATE INDEX idx_documentos_dte_estado ON documentos_dte(estado);
CREATE INDEX idx_documentos_dte_fecha ON documentos_dte(fecha_emision);
CREATE INDEX idx_documentos_dte_receptor ON documentos_dte(rut_receptor);
CREATE INDEX idx_documentos_dte_track ON documentos_dte(track_id);
-- Índice único compuesto para evitar duplicados
CREATE UNIQUE INDEX idx_documentos_dte_unico ON documentos_dte(sucursal_id, tipo_documento, folio);

-- Índices para tabla folios_dte
CREATE INDEX idx_folios_dte_sucursal ON folios_dte(sucursal_id);
CREATE INDEX idx_folios_dte_tipo ON folios_dte(tipo_documento);
CREATE INDEX idx_folios_dte_activo ON folios_dte(activo);

-- Índices para tabla fidelizacion_clientes
CREATE INDEX idx_fidelizacion_rut ON fidelizacion_clientes(rut);
CREATE INDEX idx_fidelizacion_email ON fidelizacion_clientes(email);
CREATE INDEX idx_fidelizacion_telefono ON fidelizacion_clientes(telefono);
CREATE INDEX idx_fidelizacion_nivel ON fidelizacion_clientes(nivel_fidelizacion);
CREATE INDEX idx_fidelizacion_activo ON fidelizacion_clientes(activo);
CREATE INDEX idx_fidelizacion_ultima_compra ON fidelizacion_clientes(fecha_ultima_compra);
-- Índice GIN para búsqueda de texto en nombre
CREATE INDEX idx_fidelizacion_nombre_trgm ON fidelizacion_clientes USING GIN ((nombre || ' ' || apellido) gin_trgm_ops);

-- Índices para tabla movimientos_fidelizacion
CREATE INDEX idx_movimientos_fid_cliente ON movimientos_fidelizacion(cliente_id);
CREATE INDEX idx_movimientos_fid_sucursal ON movimientos_fidelizacion(sucursal_id);
CREATE INDEX idx_movimientos_fid_venta ON movimientos_fidelizacion(venta_id);
CREATE INDEX idx_movimientos_fid_tipo ON movimientos_fidelizacion(tipo);
CREATE INDEX idx_movimientos_fid_fecha ON movimientos_fidelizacion(fecha);
CREATE INDEX idx_movimientos_fid_vencimiento ON movimientos_fidelizacion(fecha_vencimiento);

-- Índices para tabla reglas_fidelizacion
CREATE INDEX idx_reglas_fid_tipo ON reglas_fidelizacion(tipo_regla);
CREATE INDEX idx_reglas_fid_activa ON reglas_fidelizacion(activa);
CREATE INDEX idx_reglas_fid_fechas ON reglas_fidelizacion(fecha_inicio, fecha_fin);
CREATE INDEX idx_reglas_fid_prioridad ON reglas_fidelizacion(prioridad);

-- Índices para tabla notas_credito
CREATE INDEX idx_notas_credito_numero ON notas_credito(numero_nota);
CREATE INDEX idx_notas_credito_sucursal ON notas_credito(sucursal_id);
CREATE INDEX idx_notas_credito_documento_origen ON notas_credito(documento_origen_id);
CREATE INDEX idx_notas_credito_venta_origen ON notas_credito(venta_origen_id);
CREATE INDEX idx_notas_credito_supervisor ON notas_credito(supervisor_id);
CREATE INDEX idx_notas_credito_estado ON notas_credito(estado);
CREATE INDEX idx_notas_credito_fecha ON notas_credito(fecha_solicitud);

-- Índices para tabla detalle_notas_credito
CREATE INDEX idx_detalle_nc_nota ON detalle_notas_credito(nota_credito_id);
CREATE INDEX idx_detalle_nc_producto ON detalle_notas_credito(producto_id);

-- Índices para tabla despachos
CREATE INDEX idx_despachos_numero ON despachos(numero_despacho);
CREATE INDEX idx_despachos_sucursal ON despachos(sucursal_id);
CREATE INDEX idx_despachos_documento ON despachos(documento_id);
CREATE INDEX idx_despachos_venta ON despachos(venta_id);
CREATE INDEX idx_despachos_usuario ON despachos(usuario_despacho_id);
CREATE INDEX idx_despachos_cliente ON despachos(cliente_rut);
CREATE INDEX idx_despachos_estado ON despachos(estado);
CREATE INDEX idx_despachos_fecha_programada ON despachos(fecha_programada);

-- Índices para tabla detalle_despacho
CREATE INDEX idx_detalle_despacho_despacho ON detalle_despacho(despacho_id);
CREATE INDEX idx_detalle_despacho_producto ON detalle_despacho(producto_id);
CREATE INDEX idx_detalle_despacho_usuario ON detalle_despacho(usuario_despacho_id);

-- Índices para tabla reimpresiones_documentos
CREATE INDEX idx_reimpresiones_documento ON reimpresiones_documentos(documento_id);
CREATE INDEX idx_reimpresiones_usuario ON reimpresiones_documentos(usuario_id);
CREATE INDEX idx_reimpresiones_sucursal ON reimpresiones_documentos(sucursal_id);
CREATE INDEX idx_reimpresiones_fecha ON reimpresiones_documentos(fecha);
CREATE INDEX idx_reimpresiones_tipo ON reimpresiones_documentos(tipo_documento);

-- Índices para tabla logs_sincronizacion
CREATE INDEX idx_logs_sync_sucursal ON logs_sincronizacion(sucursal_id);
CREATE INDEX idx_logs_sync_terminal ON logs_sincronizacion(terminal_id);
CREATE INDEX idx_logs_sync_tipo ON logs_sincronizacion(tipo_sincronizacion);
CREATE INDEX idx_logs_sync_estado ON logs_sincronizacion(estado);
CREATE INDEX idx_logs_sync_fecha ON logs_sincronizacion(fecha_inicio);

-- Índices para tabla logs_seguridad
CREATE INDEX idx_logs_seguridad_usuario ON logs_seguridad(usuario_id);
CREATE INDEX idx_logs_seguridad_sucursal ON logs_seguridad(sucursal_id);
CREATE INDEX idx_logs_seguridad_evento ON logs_seguridad(evento);
CREATE INDEX idx_logs_seguridad_nivel ON logs_seguridad(nivel_severidad);
CREATE INDEX idx_logs_seguridad_fecha ON logs_seguridad(fecha);
CREATE INDEX idx_logs_seguridad_ip ON logs_seguridad(ip_origen);

-- Índices para tabla configuracion_sistema
CREATE INDEX idx_config_categoria ON configuracion_sistema(categoria);
CREATE INDEX idx_config_modificable ON configuracion_sistema(modificable);

-- Índices para tabla sesiones_usuario
CREATE INDEX idx_sesiones_usuario ON sesiones_usuario(usuario_id);
CREATE INDEX idx_sesiones_token ON sesiones_usuario(token_hash);
CREATE INDEX idx_sesiones_activa ON sesiones_usuario(activa);
CREATE INDEX idx_sesiones_expiracion ON sesiones_usuario(fecha_expiracion);
CREATE INDEX idx_sesiones_ultimo_acceso ON sesiones_usuario(fecha_ultimo_acceso);


-- =====================================================
-- FUNCIONES ESPECIALIZADAS
-- =====================================================

-- Función: actualizar_fecha_modificacion
-- Descripción: Actualiza automáticamente el campo fecha_modificacion
CREATE OR REPLACE FUNCTION actualizar_fecha_modificacion()
RETURNS TRIGGER AS $$
BEGIN
    NEW.fecha_modificacion = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Función: validar_stock_disponible
-- Descripción: Valida que hay stock suficiente antes de una venta
CREATE OR REPLACE FUNCTION validar_stock_disponible(
    p_producto_id UUID,
    p_sucursal_id UUID,
    p_cantidad NUMERIC
) RETURNS BOOLEAN AS $$
DECLARE
    v_stock_disponible INTEGER;
BEGIN
    SELECT cantidad_disponible INTO v_stock_disponible
    FROM stock_central
    WHERE producto_id = p_producto_id AND sucursal_id = p_sucursal_id;
    
    IF v_stock_disponible IS NULL THEN
        v_stock_disponible := 0;
    END IF;
    
    RETURN v_stock_disponible >= p_cantidad;
END;
$$ LANGUAGE plpgsql;

-- Función: descontar_stock_venta
-- Descripción: Descuenta stock automáticamente al registrar una venta
CREATE OR REPLACE FUNCTION descontar_stock_venta()
RETURNS TRIGGER AS $$
DECLARE
    v_sucursal_id UUID;
    v_stock_actual INTEGER;
BEGIN
    -- Obtener sucursal de la venta
    SELECT sucursal_id INTO v_sucursal_id
    FROM ventas
    WHERE id = NEW.venta_id;
    
    -- Verificar stock disponible
    IF NOT validar_stock_disponible(NEW.producto_id, v_sucursal_id, NEW.cantidad) THEN
        RAISE EXCEPTION 'Stock insuficiente para producto % en sucursal %', 
            NEW.producto_id, v_sucursal_id;
    END IF;
    
    -- Actualizar stock
    UPDATE stock_central
    SET cantidad = cantidad - NEW.cantidad,
        fecha_sync = NOW()
    WHERE producto_id = NEW.producto_id AND sucursal_id = v_sucursal_id;
    
    -- Si no existe registro de stock, crearlo con cantidad negativa
    IF NOT FOUND THEN
        INSERT INTO stock_central (producto_id, sucursal_id, cantidad, fecha_sync)
        VALUES (NEW.producto_id, v_sucursal_id, -NEW.cantidad, NOW());
    END IF;
    
    -- Registrar movimiento de stock
    INSERT INTO movimientos_stock (
        producto_id, sucursal_id, tipo_movimiento, cantidad,
        cantidad_anterior, cantidad_nueva, documento_referencia,
        usuario_id, observaciones
    )
    SELECT 
        NEW.producto_id, v_sucursal_id, 'venta', -NEW.cantidad,
        sc.cantidad + NEW.cantidad, sc.cantidad,
        'VENTA-' || v.numero_venta,
        v.cajero_id, 'Descuento automático por venta'
    FROM stock_central sc, ventas v
    WHERE sc.producto_id = NEW.producto_id 
      AND sc.sucursal_id = v_sucursal_id
      AND v.id = NEW.venta_id;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Función: restaurar_stock_anulacion
-- Descripción: Restaura stock cuando se anula una venta
CREATE OR REPLACE FUNCTION restaurar_stock_anulacion()
RETURNS TRIGGER AS $$
DECLARE
    detalle RECORD;
BEGIN
    -- Solo procesar si el estado cambió a 'anulada'
    IF OLD.estado != 'anulada' AND NEW.estado = 'anulada' THEN
        -- Restaurar stock para todos los productos de la venta
        FOR detalle IN 
            SELECT producto_id, cantidad 
            FROM detalle_ventas 
            WHERE venta_id = NEW.id
        LOOP
            UPDATE stock_central
            SET cantidad = cantidad + detalle.cantidad,
                fecha_sync = NOW()
            WHERE producto_id = detalle.producto_id 
              AND sucursal_id = NEW.sucursal_id;
            
            -- Registrar movimiento de stock
            INSERT INTO movimientos_stock (
                producto_id, sucursal_id, tipo_movimiento, cantidad,
                documento_referencia, usuario_id, observaciones
            ) VALUES (
                detalle.producto_id, NEW.sucursal_id, 'devolucion', detalle.cantidad,
                'ANULACION-' || NEW.numero_venta, NEW.usuario_anulacion,
                'Restauración automática por anulación de venta'
            );
        END LOOP;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Función: calcular_puntos_fidelizacion
-- Descripción: Calcula puntos de fidelización según reglas configuradas
CREATE OR REPLACE FUNCTION calcular_puntos_fidelizacion(
    p_cliente_rut TEXT,
    p_monto_compra NUMERIC,
    p_sucursal_id UUID
) RETURNS INTEGER AS $$
DECLARE
    v_puntos INTEGER := 0;
    v_multiplicador NUMERIC := 1.0;
    v_nivel_cliente TEXT;
    regla RECORD;
BEGIN
    -- Obtener nivel del cliente
    SELECT nivel_fidelizacion INTO v_nivel_cliente
    FROM fidelizacion_clientes
    WHERE rut = p_cliente_rut AND activo = true;
    
    IF v_nivel_cliente IS NULL THEN
        RETURN 0;
    END IF;
    
    -- Aplicar reglas de acumulación activas
    FOR regla IN 
        SELECT * FROM reglas_fidelizacion
        WHERE tipo_regla = 'acumulacion' 
          AND activa = true
          AND (fecha_inicio IS NULL OR fecha_inicio <= NOW())
          AND (fecha_fin IS NULL OR fecha_fin >= NOW())
        ORDER BY prioridad DESC
    LOOP
        -- Lógica simplificada: 1 punto por cada $100 pesos
        -- En implementación real, evaluar condiciones JSONB
        v_puntos := v_puntos + FLOOR(p_monto_compra / 100);
        
        -- Aplicar multiplicador según nivel
        CASE v_nivel_cliente
            WHEN 'plata' THEN v_multiplicador := 1.2;
            WHEN 'oro' THEN v_multiplicador := 1.5;
            WHEN 'platino' THEN v_multiplicador := 2.0;
            ELSE v_multiplicador := 1.0;
        END CASE;
        
        EXIT; -- Solo aplicar la primera regla por simplicidad
    END LOOP;
    
    RETURN FLOOR(v_puntos * v_multiplicador);
END;
$$ LANGUAGE plpgsql;

-- Función: acumular_puntos_fidelizacion
-- Descripción: Acumula puntos automáticamente al completar una venta
CREATE OR REPLACE FUNCTION acumular_puntos_fidelizacion()
RETURNS TRIGGER AS $$
DECLARE
    v_puntos INTEGER;
    v_cliente_id UUID;
BEGIN
    -- Solo procesar si hay cliente y la venta está finalizada
    IF NEW.cliente_rut IS NOT NULL AND NEW.estado = 'finalizada' THEN
        -- Verificar si el cliente existe en fidelización
        SELECT id INTO v_cliente_id
        FROM fidelizacion_clientes
        WHERE rut = NEW.cliente_rut AND activo = true;
        
        IF v_cliente_id IS NOT NULL THEN
            -- Calcular puntos a acumular
            v_puntos := calcular_puntos_fidelizacion(
                NEW.cliente_rut, NEW.total, NEW.sucursal_id
            );
            
            IF v_puntos > 0 THEN
                -- Actualizar puntos del cliente
                UPDATE fidelizacion_clientes
                SET puntos_actuales = puntos_actuales + v_puntos,
                    puntos_acumulados_total = puntos_acumulados_total + v_puntos,
                    fecha_ultima_compra = NEW.fecha,
                    fecha_ultima_actividad = NOW()
                WHERE id = v_cliente_id;
                
                -- Registrar movimiento
                INSERT INTO movimientos_fidelizacion (
                    cliente_id, sucursal_id, venta_id, tipo, puntos,
                    puntos_anteriores, puntos_nuevos, detalle, usuario_id
                )
                SELECT 
                    v_cliente_id, NEW.sucursal_id, NEW.id, 'acumulacion', v_puntos,
                    fc.puntos_actuales - v_puntos, fc.puntos_actuales,
                    'Acumulación por venta #' || NEW.numero_venta, NEW.cajero_id
                FROM fidelizacion_clientes fc
                WHERE fc.id = v_cliente_id;
            END IF;
        END IF;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Función: validar_canje_puntos
-- Descripción: Valida que el cliente tenga puntos suficientes para canje
CREATE OR REPLACE FUNCTION validar_canje_puntos(
    p_cliente_rut TEXT,
    p_puntos_canje INTEGER
) RETURNS BOOLEAN AS $$
DECLARE
    v_puntos_disponibles INTEGER;
BEGIN
    SELECT puntos_actuales INTO v_puntos_disponibles
    FROM fidelizacion_clientes
    WHERE rut = p_cliente_rut AND activo = true;
    
    IF v_puntos_disponibles IS NULL THEN
        RETURN false;
    END IF;
    
    RETURN v_puntos_disponibles >= p_puntos_canje;
END;
$$ LANGUAGE plpgsql;

-- Función: procesar_canje_puntos
-- Descripción: Procesa el canje de puntos de fidelización
CREATE OR REPLACE FUNCTION procesar_canje_puntos(
    p_cliente_rut TEXT,
    p_puntos_canje INTEGER,
    p_sucursal_id UUID,
    p_usuario_id UUID,
    p_detalle TEXT
) RETURNS BOOLEAN AS $$
DECLARE
    v_cliente_id UUID;
BEGIN
    -- Validar canje
    IF NOT validar_canje_puntos(p_cliente_rut, p_puntos_canje) THEN
        RETURN false;
    END IF;
    
    -- Obtener ID del cliente
    SELECT id INTO v_cliente_id
    FROM fidelizacion_clientes
    WHERE rut = p_cliente_rut AND activo = true;
    
    -- Descontar puntos
    UPDATE fidelizacion_clientes
    SET puntos_actuales = puntos_actuales - p_puntos_canje,
        fecha_ultima_actividad = NOW()
    WHERE id = v_cliente_id;
    
    -- Registrar movimiento
    INSERT INTO movimientos_fidelizacion (
        cliente_id, sucursal_id, tipo, puntos,
        puntos_anteriores, puntos_nuevos, detalle, usuario_id
    )
    SELECT 
        v_cliente_id, p_sucursal_id, 'canje', -p_puntos_canje,
        fc.puntos_actuales + p_puntos_canje, fc.puntos_actuales,
        p_detalle, p_usuario_id
    FROM fidelizacion_clientes fc
    WHERE fc.id = v_cliente_id;
    
    RETURN true;
END;
$$ LANGUAGE plpgsql;

-- Función: generar_folio_dte
-- Descripción: Genera el siguiente folio disponible para un tipo de documento
CREATE OR REPLACE FUNCTION generar_folio_dte(
    p_sucursal_id UUID,
    p_tipo_documento TEXT
) RETURNS BIGINT AS $$
DECLARE
    v_folio BIGINT;
BEGIN
    -- Obtener y actualizar el folio actual
    UPDATE folios_dte
    SET folio_actual = folio_actual + 1
    WHERE sucursal_id = p_sucursal_id 
      AND tipo_documento = p_tipo_documento
      AND activo = true
      AND folio_actual < folio_hasta
    RETURNING folio_actual INTO v_folio;
    
    IF v_folio IS NULL THEN
        RAISE EXCEPTION 'No hay folios disponibles para documento % en sucursal %', 
            p_tipo_documento, p_sucursal_id;
    END IF;
    
    RETURN v_folio;
END;
$$ LANGUAGE plpgsql;

-- Función: limpiar_sesiones_expiradas
-- Descripción: Limpia sesiones de usuario expiradas
CREATE OR REPLACE FUNCTION limpiar_sesiones_expiradas()
RETURNS INTEGER AS $$
DECLARE
    v_eliminadas INTEGER;
BEGIN
    UPDATE sesiones_usuario
    SET activa = false
    WHERE fecha_expiracion < NOW() AND activa = true;
    
    GET DIAGNOSTICS v_eliminadas = ROW_COUNT;
    
    RETURN v_eliminadas;
END;
$$ LANGUAGE plpgsql;

-- Función: registrar_evento_seguridad
-- Descripción: Registra eventos de seguridad del sistema
CREATE OR REPLACE FUNCTION registrar_evento_seguridad(
    p_usuario_id UUID,
    p_sucursal_id UUID,
    p_terminal_id UUID,
    p_evento TEXT,
    p_nivel_severidad TEXT,
    p_descripcion TEXT,
    p_ip_origen INET,
    p_datos_evento JSONB DEFAULT NULL
) RETURNS UUID AS $$
DECLARE
    v_log_id UUID;
BEGIN
    INSERT INTO logs_seguridad (
        usuario_id, sucursal_id, terminal_id, evento, nivel_severidad,
        descripcion, ip_origen, datos_evento
    ) VALUES (
        p_usuario_id, p_sucursal_id, p_terminal_id, p_evento, p_nivel_severidad,
        p_descripcion, p_ip_origen, p_datos_evento
    ) RETURNING id INTO v_log_id;
    
    RETURN v_log_id;
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- TRIGGERS
-- =====================================================

-- Trigger: Actualizar fecha_modificacion en tablas principales
CREATE TRIGGER trg_sucursales_fecha_mod
    BEFORE UPDATE ON sucursales
    FOR EACH ROW EXECUTE FUNCTION actualizar_fecha_modificacion();

CREATE TRIGGER trg_usuarios_fecha_mod
    BEFORE UPDATE ON usuarios
    FOR EACH ROW EXECUTE FUNCTION actualizar_fecha_modificacion();

CREATE TRIGGER trg_productos_fecha_mod
    BEFORE UPDATE ON productos
    FOR EACH ROW EXECUTE FUNCTION actualizar_fecha_modificacion();

CREATE TRIGGER trg_categorias_fecha_mod
    BEFORE UPDATE ON categorias_productos
    FOR EACH ROW EXECUTE FUNCTION actualizar_fecha_modificacion();

CREATE TRIGGER trg_fidelizacion_fecha_mod
    BEFORE UPDATE ON fidelizacion_clientes
    FOR EACH ROW EXECUTE FUNCTION actualizar_fecha_modificacion();

CREATE TRIGGER trg_config_sistema_fecha_mod
    BEFORE UPDATE ON configuracion_sistema
    FOR EACH ROW EXECUTE FUNCTION actualizar_fecha_modificacion();

-- Trigger: Descontar stock automáticamente al registrar detalle de venta
CREATE TRIGGER trg_descontar_stock_venta
    AFTER INSERT ON detalle_ventas
    FOR EACH ROW EXECUTE FUNCTION descontar_stock_venta();

-- Trigger: Restaurar stock al anular venta
CREATE TRIGGER trg_restaurar_stock_anulacion
    AFTER UPDATE ON ventas
    FOR EACH ROW EXECUTE FUNCTION restaurar_stock_anulacion();

-- Trigger: Acumular puntos de fidelización automáticamente
CREATE TRIGGER trg_acumular_puntos_fidelizacion
    AFTER INSERT ON ventas
    FOR EACH ROW EXECUTE FUNCTION acumular_puntos_fidelizacion();

CREATE TRIGGER trg_acumular_puntos_fidelizacion_update
    AFTER UPDATE ON ventas
    FOR EACH ROW EXECUTE FUNCTION acumular_puntos_fidelizacion();

-- =====================================================
-- VISTAS ESPECIALIZADAS
-- =====================================================

-- Vista: stock_consolidado
-- Descripción: Vista consolidada de stock con información de productos
CREATE VIEW vista_stock_consolidado AS
SELECT 
    sc.producto_id,
    p.codigo_interno,
    p.codigo_barra,
    p.descripcion,
    p.marca,
    p.precio_unitario,
    sc.sucursal_id,
    s.nombre AS sucursal_nombre,
    sc.cantidad,
    sc.cantidad_reservada,
    sc.cantidad_disponible,
    sc.costo_promedio,
    sc.fecha_ultima_entrada,
    sc.fecha_ultima_salida,
    sc.fecha_sync,
    CASE 
        WHEN sc.cantidad_disponible <= p.stock_minimo THEN 'CRITICO'
        WHEN sc.cantidad_disponible <= p.stock_minimo * 2 THEN 'BAJO'
        ELSE 'NORMAL'
    END AS estado_stock
FROM stock_central sc
JOIN productos p ON sc.producto_id = p.id
JOIN sucursales s ON sc.sucursal_id = s.id
WHERE p.activo = true AND s.habilitada = true;

-- Vista: ventas_resumen_diario
-- Descripción: Resumen de ventas por día y sucursal
CREATE VIEW vista_ventas_resumen_diario AS
SELECT 
    DATE(v.fecha) AS fecha,
    v.sucursal_id,
    s.nombre AS sucursal_nombre,
    COUNT(*) AS total_ventas,
    SUM(v.total) AS total_monto,
    AVG(v.total) AS promedio_venta,
    COUNT(DISTINCT v.cajero_id) AS cajeros_activos,
    COUNT(DISTINCT v.cliente_rut) AS clientes_unicos
FROM ventas v
JOIN sucursales s ON v.sucursal_id = s.id
WHERE v.estado = 'finalizada'
GROUP BY DATE(v.fecha), v.sucursal_id, s.nombre;

-- Vista: fidelizacion_resumen_cliente
-- Descripción: Resumen completo de fidelización por cliente
CREATE VIEW vista_fidelizacion_resumen_cliente AS
SELECT 
    fc.id,
    fc.rut,
    fc.nombre,
    fc.apellido,
    fc.email,
    fc.puntos_actuales,
    fc.puntos_acumulados_total,
    fc.nivel_fidelizacion,
    fc.fecha_ultima_compra,
    fc.fecha_ultima_actividad,
    COUNT(mf.id) AS total_movimientos,
    SUM(CASE WHEN mf.tipo = 'acumulacion' THEN mf.puntos ELSE 0 END) AS puntos_acumulados,
    SUM(CASE WHEN mf.tipo = 'canje' THEN ABS(mf.puntos) ELSE 0 END) AS puntos_canjeados,
    COUNT(DISTINCT v.id) AS total_compras,
    SUM(v.total) AS monto_total_compras
FROM fidelizacion_clientes fc
LEFT JOIN movimientos_fidelizacion mf ON fc.id = mf.cliente_id
LEFT JOIN ventas v ON fc.rut = v.cliente_rut AND v.estado = 'finalizada'
WHERE fc.activo = true
GROUP BY fc.id, fc.rut, fc.nombre, fc.apellido, fc.email, 
         fc.puntos_actuales, fc.puntos_acumulados_total, fc.nivel_fidelizacion,
         fc.fecha_ultima_compra, fc.fecha_ultima_actividad;

-- Vista: productos_mas_vendidos
-- Descripción: Productos más vendidos por período
CREATE VIEW vista_productos_mas_vendidos AS
SELECT 
    dv.producto_id,
    p.codigo_interno,
    p.codigo_barra,
    p.descripcion,
    p.marca,
    p.precio_unitario,
    SUM(dv.cantidad) AS cantidad_vendida,
    COUNT(DISTINCT dv.venta_id) AS numero_ventas,
    SUM(dv.total_item) AS monto_total_vendido,
    AVG(dv.precio_final) AS precio_promedio_venta
FROM detalle_ventas dv
JOIN productos p ON dv.producto_id = p.id
JOIN ventas v ON dv.venta_id = v.id
WHERE v.estado = 'finalizada'
  AND v.fecha >= CURRENT_DATE - INTERVAL '30 days'
GROUP BY dv.producto_id, p.codigo_interno, p.codigo_barra, 
         p.descripcion, p.marca, p.precio_unitario
ORDER BY cantidad_vendida DESC;

-- Vista: documentos_dte_pendientes
-- Descripción: Documentos DTE pendientes de procesamiento
CREATE VIEW vista_documentos_dte_pendientes AS
SELECT 
    d.id,
    d.sucursal_id,
    s.nombre AS sucursal_nombre,
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
    v.numero_venta
FROM documentos_dte d
JOIN sucursales s ON d.sucursal_id = s.id
LEFT JOIN ventas v ON d.venta_id = v.id
WHERE d.estado IN ('pendiente', 'error')
ORDER BY d.fecha_emision DESC;


-- =====================================================
-- DATOS INICIALES DEL SISTEMA
-- =====================================================

-- Configuraciones iniciales del sistema
INSERT INTO configuracion_sistema (clave, valor, tipo_dato, descripcion, categoria, modificable) VALUES
('sistema.version', '1.0.0', 'string', 'Versión actual del sistema', 'sistema', false),
('sistema.nombre', 'Ferre-POS', 'string', 'Nombre del sistema', 'sistema', false),
('sistema.timezone', 'America/Santiago', 'string', 'Zona horaria del sistema', 'sistema', true),
('fidelizacion.puntos_por_peso', '1', 'integer', 'Puntos otorgados por cada peso gastado', 'fidelizacion', true),
('fidelizacion.minimo_canje', '100', 'integer', 'Mínimo de puntos para realizar canje', 'fidelizacion', true),
('fidelizacion.expiracion_puntos_dias', '365', 'integer', 'Días para expiración de puntos', 'fidelizacion', true),
('stock.alerta_minimo', '5', 'integer', 'Cantidad mínima para alerta de stock', 'stock', true),
('dte.ambiente_default', 'produccion', 'string', 'Ambiente por defecto para DTE', 'dte', true),
('dte.reintentos_maximo', '3', 'integer', 'Máximo de reintentos para envío DTE', 'dte', true),
('seguridad.intentos_login_maximo', '5', 'integer', 'Máximo intentos fallidos de login', 'seguridad', true),
('seguridad.bloqueo_minutos', '30', 'integer', 'Minutos de bloqueo tras intentos fallidos', 'seguridad', true),
('sesiones.duracion_horas', '8', 'integer', 'Duración de sesiones en horas', 'seguridad', true),
('sync.intervalo_minutos', '5', 'integer', 'Intervalo de sincronización en minutos', 'sincronizacion', true),
('reportes.retencion_dias', '90', 'integer', 'Días de retención de logs de reportes', 'reportes', true);

-- Categorías de productos iniciales
INSERT INTO categorias_productos (codigo, nombre, descripcion, nivel, orden_visualizacion) VALUES
('HERR', 'Herramientas', 'Herramientas manuales y eléctricas', 1, 1),
('FERRE', 'Ferretería', 'Artículos de ferretería general', 1, 2),
('CONST', 'Construcción', 'Materiales de construcción', 1, 3),
('ELECT', 'Eléctrico', 'Materiales eléctricos', 1, 4),
('PLOM', 'Plomería', 'Materiales de plomería', 1, 5),
('PINT', 'Pintura', 'Pinturas y accesorios', 1, 6),
('JARD', 'Jardín', 'Herramientas y accesorios de jardín', 1, 7),
('SEG', 'Seguridad', 'Elementos de seguridad', 1, 8);

-- Subcategorías de herramientas
INSERT INTO categorias_productos (codigo, nombre, descripcion, categoria_padre_id, nivel, orden_visualizacion)
SELECT 'HERR-MAN', 'Herramientas Manuales', 'Herramientas de uso manual', id, 2, 1
FROM categorias_productos WHERE codigo = 'HERR';

INSERT INTO categorias_productos (codigo, nombre, descripcion, categoria_padre_id, nivel, orden_visualizacion)
SELECT 'HERR-ELEC', 'Herramientas Eléctricas', 'Herramientas con motor eléctrico', id, 2, 2
FROM categorias_productos WHERE codigo = 'HERR';

-- Reglas de fidelización iniciales
INSERT INTO reglas_fidelizacion (nombre, descripcion, tipo_regla, condiciones, acciones, activa, prioridad) VALUES
(
    'Acumulación Base',
    'Regla base de acumulación de puntos',
    'acumulacion',
    '{"monto_minimo": 100, "productos_excluidos": []}',
    '{"puntos_por_peso": 1, "multiplicador_base": 1.0}',
    true,
    1
),
(
    'Multiplicador Plata',
    'Multiplicador para clientes nivel plata',
    'acumulacion',
    '{"nivel_cliente": "plata"}',
    '{"multiplicador": 1.2}',
    true,
    2
),
(
    'Multiplicador Oro',
    'Multiplicador para clientes nivel oro',
    'acumulacion',
    '{"nivel_cliente": "oro"}',
    '{"multiplicador": 1.5}',
    true,
    3
),
(
    'Multiplicador Platino',
    'Multiplicador para clientes nivel platino',
    'acumulacion',
    '{"nivel_cliente": "platino"}',
    '{"multiplicador": 2.0}',
    true,
    4
);

-- =====================================================
-- PROCEDIMIENTOS DE MANTENIMIENTO
-- =====================================================

-- Procedimiento: mantenimiento_diario
-- Descripción: Tareas de mantenimiento que deben ejecutarse diariamente
CREATE OR REPLACE FUNCTION mantenimiento_diario()
RETURNS TEXT AS $$
DECLARE
    v_sesiones_eliminadas INTEGER;
    v_logs_antiguos INTEGER;
    v_resultado TEXT := '';
BEGIN
    -- Limpiar sesiones expiradas
    SELECT limpiar_sesiones_expiradas() INTO v_sesiones_eliminadas;
    v_resultado := v_resultado || 'Sesiones expiradas eliminadas: ' || v_sesiones_eliminadas || E'\n';
    
    -- Limpiar logs de sincronización antiguos (más de 30 días)
    DELETE FROM logs_sincronizacion 
    WHERE fecha_inicio < NOW() - INTERVAL '30 days';
    GET DIAGNOSTICS v_logs_antiguos = ROW_COUNT;
    v_resultado := v_resultado || 'Logs de sincronización antiguos eliminados: ' || v_logs_antiguos || E'\n';
    
    -- Actualizar estadísticas de tablas principales
    ANALYZE ventas;
    ANALYZE detalle_ventas;
    ANALYZE stock_central;
    ANALYZE movimientos_stock;
    ANALYZE fidelizacion_clientes;
    ANALYZE movimientos_fidelizacion;
    v_resultado := v_resultado || 'Estadísticas de tablas actualizadas' || E'\n';
    
    -- Registrar ejecución del mantenimiento
    INSERT INTO logs_seguridad (
        evento, nivel_severidad, descripcion, datos_evento
    ) VALUES (
        'mantenimiento_diario', 'info', 
        'Mantenimiento diario ejecutado exitosamente',
        jsonb_build_object('resultado', v_resultado)
    );
    
    RETURN v_resultado;
END;
$$ LANGUAGE plpgsql;

-- Procedimiento: reporte_stock_critico
-- Descripción: Genera reporte de productos con stock crítico
CREATE OR REPLACE FUNCTION reporte_stock_critico()
RETURNS TABLE (
    sucursal_nombre TEXT,
    codigo_producto TEXT,
    descripcion_producto TEXT,
    stock_actual INTEGER,
    stock_minimo INTEGER,
    dias_sin_movimiento INTEGER
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        s.nombre,
        p.codigo_interno,
        p.descripcion,
        sc.cantidad,
        p.stock_minimo,
        EXTRACT(DAY FROM NOW() - COALESCE(sc.fecha_ultima_salida, sc.fecha_sync))::INTEGER
    FROM stock_central sc
    JOIN productos p ON sc.producto_id = p.id
    JOIN sucursales s ON sc.sucursal_id = s.id
    WHERE sc.cantidad <= p.stock_minimo
      AND p.activo = true
      AND s.habilitada = true
    ORDER BY s.nombre, sc.cantidad ASC;
END;
$$ LANGUAGE plpgsql;

-- Procedimiento: actualizar_niveles_fidelizacion
-- Descripción: Actualiza niveles de fidelización según puntos acumulados
CREATE OR REPLACE FUNCTION actualizar_niveles_fidelizacion()
RETURNS INTEGER AS $$
DECLARE
    v_actualizados INTEGER := 0;
    cliente RECORD;
    v_nuevo_nivel TEXT;
BEGIN
    FOR cliente IN 
        SELECT id, puntos_acumulados_total, nivel_fidelizacion
        FROM fidelizacion_clientes
        WHERE activo = true
    LOOP
        -- Determinar nuevo nivel según puntos acumulados
        IF cliente.puntos_acumulados_total >= 50000 THEN
            v_nuevo_nivel := 'platino';
        ELSIF cliente.puntos_acumulados_total >= 20000 THEN
            v_nuevo_nivel := 'oro';
        ELSIF cliente.puntos_acumulados_total >= 5000 THEN
            v_nuevo_nivel := 'plata';
        ELSE
            v_nuevo_nivel := 'bronce';
        END IF;
        
        -- Actualizar si el nivel cambió
        IF v_nuevo_nivel != cliente.nivel_fidelizacion THEN
            UPDATE fidelizacion_clientes
            SET nivel_fidelizacion = v_nuevo_nivel,
                fecha_modificacion = NOW()
            WHERE id = cliente.id;
            
            v_actualizados := v_actualizados + 1;
            
            -- Registrar cambio de nivel
            INSERT INTO movimientos_fidelizacion (
                cliente_id, tipo, puntos, detalle
            ) VALUES (
                cliente.id, 'ajuste', 0,
                'Cambio de nivel: ' || cliente.nivel_fidelizacion || ' -> ' || v_nuevo_nivel
            );
        END IF;
    END LOOP;
    
    RETURN v_actualizados;
END;
$$ LANGUAGE plpgsql;

-- Procedimiento: consolidar_stock_diario
-- Descripción: Consolida movimientos de stock y actualiza promedios
CREATE OR REPLACE FUNCTION consolidar_stock_diario()
RETURNS TEXT AS $$
DECLARE
    v_resultado TEXT := '';
    v_productos_actualizados INTEGER := 0;
    producto RECORD;
    v_costo_promedio NUMERIC;
BEGIN
    -- Actualizar costo promedio para productos con movimientos recientes
    FOR producto IN 
        SELECT DISTINCT ms.producto_id, ms.sucursal_id
        FROM movimientos_stock ms
        WHERE ms.fecha >= CURRENT_DATE
          AND ms.tipo_movimiento IN ('entrada', 'transferencia_entrada')
          AND ms.costo_unitario IS NOT NULL
    LOOP
        -- Calcular costo promedio ponderado
        SELECT 
            SUM(ms.cantidad * ms.costo_unitario) / NULLIF(SUM(ms.cantidad), 0)
        INTO v_costo_promedio
        FROM movimientos_stock ms
        WHERE ms.producto_id = producto.producto_id
          AND ms.sucursal_id = producto.sucursal_id
          AND ms.tipo_movimiento IN ('entrada', 'transferencia_entrada')
          AND ms.costo_unitario IS NOT NULL
          AND ms.fecha >= CURRENT_DATE - INTERVAL '30 days';
        
        -- Actualizar stock central
        UPDATE stock_central
        SET costo_promedio = v_costo_promedio
        WHERE producto_id = producto.producto_id
          AND sucursal_id = producto.sucursal_id;
        
        v_productos_actualizados := v_productos_actualizados + 1;
    END LOOP;
    
    v_resultado := 'Productos con costo promedio actualizado: ' || v_productos_actualizados;
    
    RETURN v_resultado;
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- COMENTARIOS Y DOCUMENTACIÓN FINAL
-- =====================================================

-- Comentarios en tablas principales
COMMENT ON TABLE sucursales IS 'Registro de todas las sucursales de la cadena con su configuración específica';
COMMENT ON TABLE usuarios IS 'Usuarios del sistema con roles y permisos granulares';
COMMENT ON TABLE productos IS 'Catálogo maestro de productos con información completa';
COMMENT ON TABLE stock_central IS 'Consolidación de inventario de todas las sucursales';
COMMENT ON TABLE ventas IS 'Registro maestro de todas las transacciones de venta';
COMMENT ON TABLE documentos_dte IS 'Documentos tributarios electrónicos emitidos';
COMMENT ON TABLE fidelizacion_clientes IS 'Clientes registrados en programa de fidelización';
COMMENT ON TABLE movimientos_fidelizacion IS 'Historial de movimientos de puntos de fidelización';

-- Comentarios en funciones principales
COMMENT ON FUNCTION validar_stock_disponible(UUID, UUID, NUMERIC) IS 'Valida disponibilidad de stock antes de ventas';
COMMENT ON FUNCTION calcular_puntos_fidelizacion(TEXT, NUMERIC, UUID) IS 'Calcula puntos según reglas de fidelización';
COMMENT ON FUNCTION generar_folio_dte(UUID, TEXT) IS 'Genera folios consecutivos para documentos DTE';
COMMENT ON FUNCTION mantenimiento_diario() IS 'Rutina de mantenimiento diario automatizado';

-- =====================================================
-- PERMISOS Y SEGURIDAD
-- =====================================================

-- Crear roles de base de datos
CREATE ROLE ferre_pos_app;
CREATE ROLE ferre_pos_readonly;
CREATE ROLE ferre_pos_admin;

-- Permisos para aplicación
GRANT CONNECT ON DATABASE postgres TO ferre_pos_app;
GRANT USAGE ON SCHEMA public TO ferre_pos_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO ferre_pos_app;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO ferre_pos_app;
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO ferre_pos_app;

-- Permisos para consultas de solo lectura
GRANT CONNECT ON DATABASE postgres TO ferre_pos_readonly;
GRANT USAGE ON SCHEMA public TO ferre_pos_readonly;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO ferre_pos_readonly;
GRANT SELECT ON ALL SEQUENCES IN SCHEMA public TO ferre_pos_readonly;

-- Permisos para administración
GRANT ALL PRIVILEGES ON DATABASE postgres TO ferre_pos_admin;

-- Configurar permisos por defecto para objetos futuros
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO ferre_pos_app;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT USAGE, SELECT ON SEQUENCES TO ferre_pos_app;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT EXECUTE ON FUNCTIONS TO ferre_pos_app;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO ferre_pos_readonly;

-- =====================================================
-- CONFIGURACIÓN FINAL DE POSTGRESQL
-- =====================================================

-- Configuraciones recomendadas para optimización
-- Estas configuraciones deben aplicarse en postgresql.conf

/*
Configuraciones recomendadas para postgresql.conf:

# Memoria
shared_buffers = 256MB                    # 25% de RAM disponible
effective_cache_size = 1GB                # 75% de RAM disponible
work_mem = 4MB                            # Para operaciones de ordenamiento
maintenance_work_mem = 64MB               # Para operaciones de mantenimiento

# Checkpoint y WAL
checkpoint_completion_target = 0.9
wal_buffers = 16MB
checkpoint_timeout = 10min
max_wal_size = 1GB
min_wal_size = 80MB

# Logging
log_destination = 'stderr'
logging_collector = on
log_directory = 'log'
log_filename = 'postgresql-%Y-%m-%d_%H%M%S.log'
log_min_duration_statement = 1000         # Log consultas > 1 segundo
log_checkpoints = on
log_connections = on
log_disconnections = on
log_lock_waits = on

# Conexiones
max_connections = 100
shared_preload_libraries = 'pg_stat_statements'

# Autovacuum
autovacuum = on
autovacuum_max_workers = 3
autovacuum_naptime = 1min
*/

-- =====================================================
-- FINALIZACIÓN DEL SCRIPT
-- =====================================================

-- Mensaje de finalización
DO $$
BEGIN
    RAISE NOTICE '=================================================';
    RAISE NOTICE 'ESQUEMA FERRE-POS SERVIDOR CENTRAL CREADO EXITOSAMENTE';
    RAISE NOTICE '=================================================';
    RAISE NOTICE 'Versión: 1.0';
    RAISE NOTICE 'Fecha: %', NOW();
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
    RAISE NOTICE 'Vistas creadas: %', (
        SELECT COUNT(*) FROM information_schema.views WHERE table_schema = 'public'
    );
    RAISE NOTICE '=================================================';
    RAISE NOTICE 'PRÓXIMOS PASOS:';
    RAISE NOTICE '1. Configurar postgresql.conf según recomendaciones';
    RAISE NOTICE '2. Crear usuarios de aplicación con roles apropiados';
    RAISE NOTICE '3. Configurar respaldos automáticos';
    RAISE NOTICE '4. Programar ejecución de mantenimiento_diario()';
    RAISE NOTICE '5. Cargar datos iniciales de sucursales y productos';
    RAISE NOTICE '=================================================';
END $$;

-- Fin del script
-- =====================================================

