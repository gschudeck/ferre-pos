
-- Sucursales
CREATE TABLE sucursales (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  nombre TEXT NOT NULL,
  direccion TEXT,
  comuna TEXT,
  region TEXT,
  habilitada BOOLEAN DEFAULT true
);

-- Usuarios
CREATE TABLE usuarios (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  rut TEXT UNIQUE NOT NULL,
  nombre TEXT,
  rol TEXT CHECK (rol IN ('cajero', 'vendedor', 'despacho', 'admin', 'supervisor')),
  sucursal_id UUID REFERENCES sucursales(id),
  activo BOOLEAN DEFAULT true
);

-- Productos
CREATE TABLE productos (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  codigo_barra TEXT UNIQUE NOT NULL,
  descripcion TEXT NOT NULL,
  precio_unitario NUMERIC(10,2),
  unidad TEXT,
  activo BOOLEAN DEFAULT true
);

-- Stock local
CREATE TABLE stock (
  producto_id UUID REFERENCES productos(id),
  sucursal_id UUID REFERENCES sucursales(id),
  cantidad INTEGER DEFAULT 0,
  PRIMARY KEY (producto_id, sucursal_id)
);

-- Stock central (réplica)
CREATE TABLE stock_central (
  producto_id UUID,
  sucursal_id UUID,
  cantidad INTEGER,
  fecha_sync TIMESTAMP,
  PRIMARY KEY (producto_id, sucursal_id)
);

-- Log sincronización de stock
CREATE TABLE log_sync_stock (
  id SERIAL PRIMARY KEY,
  sucursal_id UUID,
  registros_enviados INTEGER,
  usuario_api TEXT,
  fecha TIMESTAMP DEFAULT NOW()
);

-- Historial de sincronización de stock
CREATE TABLE historial_stock_sync (
  id SERIAL PRIMARY KEY,
  producto_id UUID,
  sucursal_id UUID,
  cantidad_anterior INTEGER,
  cantidad_nueva INTEGER,
  fecha TIMESTAMP DEFAULT NOW()
);

-- Ventas
CREATE TABLE ventas (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  fecha TIMESTAMP DEFAULT NOW(),
  usuario_id UUID REFERENCES usuarios(id),
  sucursal_id UUID REFERENCES sucursales(id),
  tipo_documento TEXT,
  total NUMERIC(10,2),
  dte_emitido BOOLEAN DEFAULT false,
  dte_id UUID
);

-- Detalle de venta
CREATE TABLE detalle_venta (
  id SERIAL PRIMARY KEY,
  venta_id UUID REFERENCES ventas(id),
  producto_id UUID REFERENCES productos(id),
  cantidad INTEGER,
  precio_unitario NUMERIC(10,2),
  total_item NUMERIC(10,2)
);

-- Documentos DTE
CREATE TABLE documentos_dte (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tipo TEXT,
  folio INTEGER,
  estado TEXT,
  xml TEXT,
  proveedor_id UUID,
  respuesta_proveedor TEXT,
  fecha TIMESTAMP DEFAULT NOW()
);

-- Proveedores DTE
CREATE TABLE proveedores_dte (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  sucursal_id UUID REFERENCES sucursales(id),
  nombre TEXT,
  api_url TEXT,
  api_key TEXT,
  habilitado BOOLEAN DEFAULT true
);

-- Fidelización clientes
CREATE TABLE fidelizacion_clientes (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  rut TEXT UNIQUE NOT NULL,
  nombre TEXT,
  puntos INTEGER DEFAULT 0,
  fecha_ultima_compra TIMESTAMP,
  fecha_creacion TIMESTAMP DEFAULT NOW()
);

-- Movimientos fidelización
CREATE TABLE movimientos_fidelizacion (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  cliente_id UUID REFERENCES fidelizacion_clientes(id),
  sucursal_id UUID REFERENCES sucursales(id),
  tipo TEXT CHECK (tipo IN ('acumulacion', 'canje')),
  puntos INTEGER CHECK (puntos > 0),
  detalle TEXT,
  fecha TIMESTAMP DEFAULT NOW()
);

-- Notas de crédito
CREATE TABLE notas_credito (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  documento_origen_id UUID NOT NULL REFERENCES documentos_dte(id),
  supervisor_id UUID NOT NULL REFERENCES usuarios(id),
  cajero_id UUID NOT NULL REFERENCES usuarios(id),
  motivo TEXT NOT NULL,
  total NUMERIC(10,2) NOT NULL,
  estado TEXT CHECK (estado IN ('pendiente', 'autorizada', 'enviada', 'rechazada')) DEFAULT 'pendiente',
  fecha TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_notas_credito_documento ON notas_credito(documento_origen_id);
CREATE INDEX idx_notas_credito_estado ON notas_credito(estado);

-- Despachos
CREATE TABLE despachos (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  documento_id UUID REFERENCES documentos_dte(id),
  usuario_id UUID REFERENCES usuarios(id),
  fecha TIMESTAMP DEFAULT NOW(),
  estado TEXT CHECK (estado IN ('completo', 'parcial', 'rechazado')) DEFAULT 'completo',
  observacion TEXT
);

CREATE TABLE detalle_despacho (
  id SERIAL PRIMARY KEY,
  despacho_id UUID REFERENCES despachos(id),
  producto_id UUID REFERENCES productos(id),
  cantidad_vendida INTEGER,
  cantidad_entregada INTEGER
);
