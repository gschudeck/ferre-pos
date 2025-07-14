
-- Habilitar extensiones necesarias
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE EXTENSION IF NOT EXISTS fuzzystrmatch;

-- Tabla productos
CREATE TABLE productos (
  id UUID PRIMARY KEY,
  codigo TEXT NOT NULL,
  descripcion TEXT NOT NULL,
  proveedor_id UUID REFERENCES proveedores(id),
  precio NUMERIC(10, 2),
  activo BOOLEAN DEFAULT TRUE
);

CREATE INDEX idx_productos_codigo ON productos (codigo);
CREATE INDEX idx_productos_proveedor ON productos (proveedor_id);
CREATE INDEX idx_productos_descripcion_trgm ON productos USING gin (descripcion gin_trgm_ops);

-- Tabla stock
CREATE TABLE stock (
  producto_id UUID REFERENCES productos(id),
  sucursal_id UUID REFERENCES sucursales(id),
  cantidad INTEGER,
  PRIMARY KEY (producto_id, sucursal_id)
);

CREATE INDEX idx_stock_critico ON stock (cantidad) WHERE cantidad < 5;

-- Tabla ventas
CREATE TABLE ventas (
  id UUID PRIMARY KEY,
  sucursal_id UUID REFERENCES sucursales(id),
  terminal_id UUID,
  cliente_rut TEXT,
  total NUMERIC(10,2),
  fecha TIMESTAMP DEFAULT NOW(),
  sincronizado BOOLEAN DEFAULT FALSE
);

CREATE INDEX idx_ventas_fecha ON ventas (fecha);
CREATE INDEX idx_ventas_sync ON ventas (sincronizado);

-- Tabla detalle_ventas
CREATE TABLE detalle_ventas (
  id SERIAL PRIMARY KEY,
  venta_id UUID REFERENCES ventas(id),
  producto_id UUID REFERENCES productos(id),
  cantidad INTEGER,
  precio_unitario NUMERIC(10,2),
  total_item NUMERIC(10,2)
);

-- Tabla sucursales
CREATE TABLE sucursales (
  id UUID PRIMARY KEY,
  nombre TEXT NOT NULL,
  direccion TEXT
);

-- Tabla proveedores
CREATE TABLE proveedores (
  id UUID PRIMARY KEY,
  nombre TEXT
);

-- Tabla terminales
CREATE TABLE terminales (
  id UUID PRIMARY KEY,
  sucursal_id UUID REFERENCES sucursales(id),
  nombre_terminal TEXT,
  timestamp_ultima_venta TIMESTAMP
);

-- Tabla logs_sincronizacion
CREATE TABLE logs_sincronizacion (
  id SERIAL PRIMARY KEY,
  terminal_id UUID REFERENCES terminales(id),
  sucursal_id UUID REFERENCES sucursales(id),
  tipo TEXT CHECK (tipo IN ('venta', 'stock', 'ping', 'error')),
  resultado TEXT CHECK (resultado IN ('ok', 'fallo', 'parcial')),
  detalles TEXT,
  intentos INTEGER DEFAULT 0,
  fecha TIMESTAMP DEFAULT NOW(),
  ip_origen TEXT
);

CREATE INDEX idx_logs_sincronizacion_fecha ON logs_sincronizacion (fecha);
CREATE INDEX idx_logs_sincronizacion_terminal_tipo ON logs_sincronizacion (terminal_id, tipo);


-- Trigger para actualizar stock automáticamente cuando se realiza una venta
CREATE OR REPLACE FUNCTION actualizar_stock_despues_venta()
RETURNS TRIGGER AS $$
BEGIN
  UPDATE stock
  SET cantidad = cantidad - NEW.cantidad
  WHERE producto_id = NEW.producto_id AND sucursal_id = (SELECT sucursal_id FROM ventas WHERE id = NEW.venta_id);
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_actualizar_stock
AFTER INSERT ON detalle_ventas
FOR EACH ROW
EXECUTE FUNCTION actualizar_stock_despues_venta();

-- Trigger para marcar ventas como sincronizadas al recibir confirmación
CREATE OR REPLACE FUNCTION marcar_venta_sincronizada()
RETURNS TRIGGER AS $$
BEGIN
  UPDATE ventas SET sincronizado = TRUE WHERE id = NEW.venta_id;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- (Ejemplo de uso: se podría asociar este trigger a una tabla de confirmaciones si se desea)

-- Tabla de fidelización
CREATE TABLE puntos_fidelizacion (
  id SERIAL PRIMARY KEY,
  cliente_rut TEXT NOT NULL,
  puntos_acumulados INTEGER DEFAULT 0,
  fecha TIMESTAMP DEFAULT NOW()
);

-- Trigger para otorgar puntos después de una venta
CREATE OR REPLACE FUNCTION otorgar_puntos_fidelizacion()
RETURNS TRIGGER AS $$
DECLARE
  puntos INTEGER;
BEGIN
  -- Supongamos 1 punto por cada $1.000 CLP
  puntos := FLOOR(NEW.total / 1000);
  IF puntos > 0 THEN
    INSERT INTO puntos_fidelizacion (cliente_rut, puntos_acumulados, fecha)
    VALUES (NEW.cliente_rut, puntos, NOW());
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_otorgar_puntos
AFTER INSERT ON ventas
FOR EACH ROW
WHEN (NEW.cliente_rut IS NOT NULL)
EXECUTE FUNCTION otorgar_puntos_fidelizacion();


-- Tabla de notas de crédito
CREATE TABLE notas_credito (
  id UUID PRIMARY KEY,
  venta_id UUID REFERENCES ventas(id),
  motivo TEXT,
  aprobada_por UUID,
  fecha TIMESTAMP DEFAULT NOW()
);

-- Trigger: al insertar nota de crédito, reponer stock
CREATE OR REPLACE FUNCTION reponer_stock_nota_credito()
RETURNS TRIGGER AS $$
BEGIN
  UPDATE stock
  SET cantidad = cantidad + (
    SELECT dv.cantidad
    FROM detalle_ventas dv
    WHERE dv.producto_id = stock.producto_id
      AND dv.venta_id = NEW.venta_id
  )
  WHERE EXISTS (
    SELECT 1 FROM detalle_ventas dv
    WHERE dv.producto_id = stock.producto_id
      AND dv.venta_id = NEW.venta_id
      AND stock.sucursal_id = (SELECT sucursal_id FROM ventas WHERE id = NEW.venta_id)
  );
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_reponer_stock
AFTER INSERT ON notas_credito
FOR EACH ROW
EXECUTE FUNCTION reponer_stock_nota_credito();

-- Trigger: descuenta puntos fidelización si la venta tenía RUT
CREATE OR REPLACE FUNCTION revertir_puntos_fidelizacion()
RETURNS TRIGGER AS $$
DECLARE
  total_puntos INTEGER;
BEGIN
  SELECT FLOOR(v.total / 1000) INTO total_puntos FROM ventas v WHERE v.id = NEW.venta_id;
  IF total_puntos > 0 THEN
    INSERT INTO puntos_fidelizacion (cliente_rut, puntos_acumulados, fecha)
    SELECT v.cliente_rut, -total_puntos, NOW() FROM ventas v WHERE v.id = NEW.venta_id;
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_revertir_puntos
AFTER INSERT ON notas_credito
FOR EACH ROW
WHEN (EXISTS (SELECT 1 FROM ventas WHERE id = NEW.venta_id AND cliente_rut IS NOT NULL))
EXECUTE FUNCTION revertir_puntos_fidelizacion();
