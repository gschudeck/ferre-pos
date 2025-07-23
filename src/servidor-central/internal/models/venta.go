package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Venta representa una venta en el sistema
type Venta struct {
	BaseModel
	NumeroVenta           int64                 `json:"numero_venta" db:"numero_venta"`
	SucursalID            uuid.UUID             `json:"sucursal_id" db:"sucursal_id" binding:"required"`
	TerminalID            *uuid.UUID            `json:"terminal_id,omitempty" db:"terminal_id"`
	CajeroID              uuid.UUID             `json:"cajero_id" db:"cajero_id" binding:"required"`
	VendedorID            *uuid.UUID            `json:"vendedor_id,omitempty" db:"vendedor_id"`
	ClienteRut            *string               `json:"cliente_rut,omitempty" db:"cliente_rut"`
	ClienteNombre         *string               `json:"cliente_nombre,omitempty" db:"cliente_nombre"`
	NotaVentaID           *uuid.UUID            `json:"nota_venta_id,omitempty" db:"nota_venta_id"`
	TipoDocumento         TipoDocumento         `json:"tipo_documento" db:"tipo_documento" binding:"required"`
	Subtotal              float64               `json:"subtotal" db:"subtotal" binding:"required,min=0"`
	DescuentoTotal        float64               `json:"descuento_total" db:"descuento_total" default:"0"`
	ImpuestoTotal         float64               `json:"impuesto_total" db:"impuesto_total" default:"0"`
	Total                 float64               `json:"total" db:"total" binding:"required,min=0"`
	Estado                EstadoVenta           `json:"estado" db:"estado" default:"finalizada"`
	Fecha                 time.Time             `json:"fecha" db:"fecha" default:"NOW()"`
	FechaAnulacion        *time.Time            `json:"fecha_anulacion,omitempty" db:"fecha_anulacion"`
	MotivoAnulacion       *string               `json:"motivo_anulacion,omitempty" db:"motivo_anulacion"`
	UsuarioAnulacion      *uuid.UUID            `json:"usuario_anulacion,omitempty" db:"usuario_anulacion"`
	DteID                 *uuid.UUID            `json:"dte_id,omitempty" db:"dte_id"`
	DteEmitido            bool                  `json:"dte_emitido" db:"dte_emitido" default:"false"`
	Sincronizada          bool                  `json:"sincronizada" db:"sincronizada" default:"false"`
	FechaSincronizacion   *time.Time            `json:"fecha_sincronizacion,omitempty" db:"fecha_sincronizacion"`
	DatosAdicionales      *DatosAdicionalesVenta `json:"datos_adicionales,omitempty" db:"datos_adicionales"`
	HashIntegridad        *string               `json:"hash_integridad,omitempty" db:"hash_integridad"`
	TiempoProcesamiento   *int                  `json:"tiempo_procesamiento_ms,omitempty" db:"tiempo_procesamiento_ms"`
	ProcesoOrigen         PrioridadProceso      `json:"proceso_origen" db:"proceso_origen" default:"maxima"`
	CacheTotales          *CacheTotales         `json:"cache_totales,omitempty" db:"cache_totales"`

	// Relaciones
	Sucursal         *Sucursal         `json:"sucursal,omitempty" gorm:"foreignKey:SucursalID"`
	Terminal         *Terminal         `json:"terminal,omitempty" gorm:"foreignKey:TerminalID"`
	Cajero           *Usuario          `json:"cajero,omitempty" gorm:"foreignKey:CajeroID"`
	Vendedor         *Usuario          `json:"vendedor,omitempty" gorm:"foreignKey:VendedorID"`
	NotaVenta        *NotaVenta        `json:"nota_venta,omitempty" gorm:"foreignKey:NotaVentaID"`
	DetalleVentas    []DetalleVenta    `json:"detalle_ventas,omitempty" gorm:"foreignKey:VentaID"`
	MediosPago       []MedioPagoVenta  `json:"medios_pago,omitempty" gorm:"foreignKey:VentaID"`
	UsuarioAnula     *Usuario          `json:"usuario_anula,omitempty" gorm:"foreignKey:UsuarioAnulacion"`
}

// DetalleVenta representa el detalle de una venta
type DetalleVenta struct {
	BaseModel
	VentaID            uuid.UUID `json:"venta_id" db:"venta_id" binding:"required"`
	ProductoID         uuid.UUID `json:"producto_id" db:"producto_id" binding:"required"`
	Cantidad           float64   `json:"cantidad" db:"cantidad" binding:"required,min=0"`
	PrecioUnitario     float64   `json:"precio_unitario" db:"precio_unitario" binding:"required,min=0"`
	DescuentoUnitario  float64   `json:"descuento_unitario" db:"descuento_unitario" default:"0"`
	PrecioFinal        float64   `json:"precio_final" db:"precio_final" binding:"required,min=0"`
	TotalItem          float64   `json:"total_item" db:"total_item" binding:"required,min=0"`
	NumeroSerie        *string   `json:"numero_serie,omitempty" db:"numero_serie"`
	Lote               *string   `json:"lote,omitempty" db:"lote"`
	FechaVencimiento   *time.Time `json:"fecha_vencimiento,omitempty" db:"fecha_vencimiento"`
	DatosAdicionales   *DatosAdicionalesDetalle `json:"datos_adicionales,omitempty" db:"datos_adicionales"`
	MargenUnitario     *float64  `json:"margen_unitario,omitempty" db:"margen_unitario"`
	CategoriaProductoID *uuid.UUID `json:"categoria_producto_id,omitempty" db:"categoria_producto_id"`

	// Relaciones
	Venta    *Venta    `json:"venta,omitempty" gorm:"foreignKey:VentaID"`
	Producto *Producto `json:"producto,omitempty" gorm:"foreignKey:ProductoID"`
}

// MedioPagoVenta representa los medios de pago utilizados en una venta
type MedioPagoVenta struct {
	BaseModel
	VentaID                uuid.UUID           `json:"venta_id" db:"venta_id" binding:"required"`
	MedioPago              MedioPago           `json:"medio_pago" db:"medio_pago" binding:"required"`
	Monto                  float64             `json:"monto" db:"monto" binding:"required,min=0"`
	ReferenciaTransaccion  *string             `json:"referencia_transaccion,omitempty" db:"referencia_transaccion"`
	CodigoAutorizacion     *string             `json:"codigo_autorizacion,omitempty" db:"codigo_autorizacion"`
	DatosTransaccion       *DatosTransaccion   `json:"datos_transaccion,omitempty" db:"datos_transaccion"`
	FechaProcesamiento     time.Time           `json:"fecha_procesamiento" db:"fecha_procesamiento" default:"NOW()"`
	EstadoConciliacion     EstadoConciliacion  `json:"estado_conciliacion" db:"estado_conciliacion" default:"pendiente"`
	FechaConciliacion      *time.Time          `json:"fecha_conciliacion,omitempty" db:"fecha_conciliacion"`
	LoteConciliacion       *string             `json:"lote_conciliacion,omitempty" db:"lote_conciliacion"`
	Comision               *float64            `json:"comision,omitempty" db:"comision"`

	// Relaciones
	Venta *Venta `json:"venta,omitempty" gorm:"foreignKey:VentaID"`
}

// NotaVenta representa una nota de venta (pre-venta)
type NotaVenta struct {
	BaseModel
	NumeroNota       int64                `json:"numero_nota" db:"numero_nota"`
	SucursalID       uuid.UUID            `json:"sucursal_id" db:"sucursal_id" binding:"required"`
	TerminalID       *uuid.UUID           `json:"terminal_id,omitempty" db:"terminal_id"`
	VendedorID       uuid.UUID            `json:"vendedor_id" db:"vendedor_id" binding:"required"`
	ClienteRut       *string              `json:"cliente_rut,omitempty" db:"cliente_rut"`
	ClienteNombre    *string              `json:"cliente_nombre,omitempty" db:"cliente_nombre"`
	Subtotal         float64              `json:"subtotal" db:"subtotal" binding:"required,min=0"`
	DescuentoTotal   float64              `json:"descuento_total" db:"descuento_total" default:"0"`
	Total            float64              `json:"total" db:"total" binding:"required,min=0"`
	Estado           EstadoNotaVenta      `json:"estado" db:"estado" default:"pendiente"`
	Fecha            time.Time            `json:"fecha" db:"fecha" default:"NOW()"`
	FechaVencimiento *time.Time           `json:"fecha_vencimiento,omitempty" db:"fecha_vencimiento"`
	VentaID          *uuid.UUID           `json:"venta_id,omitempty" db:"venta_id"`
	FechaPago        *time.Time           `json:"fecha_pago,omitempty" db:"fecha_pago"`
	Sincronizada     bool                 `json:"sincronizada" db:"sincronizada" default:"false"`
	Observaciones    *string              `json:"observaciones,omitempty" db:"observaciones"`
	QRCode           *string              `json:"qr_code,omitempty" db:"qr_code"`
	HashValidacion   *string              `json:"hash_validacion,omitempty" db:"hash_validacion"`
	TiempoVigencia   int                  `json:"tiempo_vigencia_horas" db:"tiempo_vigencia_horas" default:"24"`
	PrioridadAtencion int                 `json:"prioridad_atencion" db:"prioridad_atencion" default:"5"`

	// Relaciones
	Sucursal      *Sucursal           `json:"sucursal,omitempty" gorm:"foreignKey:SucursalID"`
	Terminal      *Terminal           `json:"terminal,omitempty" gorm:"foreignKey:TerminalID"`
	Vendedor      *Usuario            `json:"vendedor,omitempty" gorm:"foreignKey:VendedorID"`
	Venta         *Venta              `json:"venta,omitempty" gorm:"foreignKey:VentaID"`
	DetalleNotas  []DetalleNotaVenta  `json:"detalle_notas,omitempty" gorm:"foreignKey:NotaVentaID"`
}

// DetalleNotaVenta representa el detalle de una nota de venta
type DetalleNotaVenta struct {
	BaseModel
	NotaVentaID              uuid.UUID `json:"nota_venta_id" db:"nota_venta_id" binding:"required"`
	ProductoID               uuid.UUID `json:"producto_id" db:"producto_id" binding:"required"`
	Cantidad                 float64   `json:"cantidad" db:"cantidad" binding:"required,min=0"`
	PrecioUnitario           float64   `json:"precio_unitario" db:"precio_unitario" binding:"required,min=0"`
	DescuentoUnitario        float64   `json:"descuento_unitario" db:"descuento_unitario" default:"0"`
	TotalItem                float64   `json:"total_item" db:"total_item" binding:"required,min=0"`
	Observaciones            *string   `json:"observaciones,omitempty" db:"observaciones"`
	DisponibilidadVerificada bool      `json:"disponibilidad_verificada" db:"disponibilidad_verificada" default:"false"`
	FechaVerificacionStock   *time.Time `json:"fecha_verificacion_stock,omitempty" db:"fecha_verificacion_stock"`

	// Relaciones
	NotaVenta *NotaVenta `json:"nota_venta,omitempty" gorm:"foreignKey:NotaVentaID"`
	Producto  *Producto  `json:"producto,omitempty" gorm:"foreignKey:ProductoID"`
}

// Terminal representa un terminal de punto de venta
type Terminal struct {
	BaseModel
	Codigo                string                `json:"codigo" db:"codigo" binding:"required" validate:"required,max=50"`
	NombreTerminal        string                `json:"nombre_terminal" db:"nombre_terminal" binding:"required"`
	TipoTerminal          TipoTerminal          `json:"tipo_terminal" db:"tipo_terminal" binding:"required"`
	SucursalID            uuid.UUID             `json:"sucursal_id" db:"sucursal_id" binding:"required"`
	DireccionIP           *string               `json:"direccion_ip,omitempty" db:"direccion_ip"`
	DireccionMAC          *string               `json:"direccion_mac,omitempty" db:"direccion_mac"`
	Activo                bool                  `json:"activo" db:"activo" default:"true"`
	UltimaConexion        *time.Time            `json:"ultima_conexion,omitempty" db:"ultima_conexion"`
	VersionSoftware       *string               `json:"version_software,omitempty" db:"version_software"`
	Configuracion         *ConfiguracionTerminal `json:"configuracion,omitempty" db:"configuracion"`
	FechaInstalacion      time.Time             `json:"fecha_instalacion" db:"fecha_instalacion" default:"NOW()"`
	HeartbeatInterval     int                   `json:"heartbeat_interval" db:"heartbeat_interval" default:"30"`
	EstadoConexion        EstadoConexion        `json:"estado_conexion" db:"estado_conexion" default:"desconectado"`
	MetricasRendimiento   *MetricasTerminal     `json:"metricas_rendimiento,omitempty" db:"metricas_rendimiento"`
	ConfiguracionCache    *ConfiguracionCacheTerminal `json:"configuracion_cache,omitempty" db:"configuracion_cache"`

	// Relaciones
	Sucursal *Sucursal `json:"sucursal,omitempty" gorm:"foreignKey:SucursalID"`
}

// Enums para ventas

type TipoDocumento string

const (
	TipoDocumentoBoleta   TipoDocumento = "boleta"
	TipoDocumentoFactura  TipoDocumento = "factura"
	TipoDocumentoGuia     TipoDocumento = "guia"
	TipoDocumentoNotaVenta TipoDocumento = "nota_venta"
)

type EstadoVenta string

const (
	EstadoVentaPendiente  EstadoVenta = "pendiente"
	EstadoVentaFinalizada EstadoVenta = "finalizada"
	EstadoVentaAnulada    EstadoVenta = "anulada"
)

type EstadoNotaVenta string

const (
	EstadoNotaPendiente EstadoNotaVenta = "pendiente"
	EstadoNotaPagada    EstadoNotaVenta = "pagada"
	EstadoNotaCancelada EstadoNotaVenta = "cancelada"
	EstadoNotaVencida   EstadoNotaVenta = "vencida"
)

type MedioPago string

const (
	MedioPagoEfectivo            MedioPago = "efectivo"
	MedioPagoTarjetaDebito       MedioPago = "tarjeta_debito"
	MedioPagoTarjetaCredito      MedioPago = "tarjeta_credito"
	MedioPagoTransferencia       MedioPago = "transferencia"
	MedioPagoCheque              MedioPago = "cheque"
	MedioPagoPuntosFidelizacion  MedioPago = "puntos_fidelizacion"
	MedioPagoOtro                MedioPago = "otro"
)

type EstadoConciliacion string

const (
	ConciliacionPendiente  EstadoConciliacion = "pendiente"
	ConciliacionConciliado EstadoConciliacion = "conciliado"
	ConciliacionDiferencia EstadoConciliacion = "diferencia"
	ConciliacionError      EstadoConciliacion = "error"
)

type TipoTerminal string

const (
	TerminalCaja         TipoTerminal = "caja"
	TerminalTienda       TipoTerminal = "tienda"
	TerminalDespacho     TipoTerminal = "despacho"
	TerminalAutoatencion TipoTerminal = "autoatencion"
	TerminalEtiquetas    TipoTerminal = "etiquetas"
)

type EstadoConexion string

const (
	ConexionConectado     EstadoConexion = "conectado"
	ConexionDesconectado  EstadoConexion = "desconectado"
	ConexionError         EstadoConexion = "error"
	ConexionMantenimiento EstadoConexion = "mantenimiento"
)

// Estructuras JSON

type DatosAdicionalesVenta struct {
	Observaciones       *string                `json:"observaciones,omitempty"`
	ReferenciaExterna   *string                `json:"referencia_externa,omitempty"`
	CanalVenta          *string                `json:"canal_venta,omitempty"`
	PromocionesAplicadas []string              `json:"promociones_aplicadas,omitempty"`
	MetadatosPersonalizados map[string]interface{} `json:"metadatos_personalizados,omitempty"`
}

type DatosAdicionalesDetalle struct {
	Garantia            *string                `json:"garantia,omitempty"`
	Instalacion         *bool                  `json:"instalacion,omitempty"`
	Promocion           *string                `json:"promocion,omitempty"`
	MetadatosItem       map[string]interface{} `json:"metadatos_item,omitempty"`
}

type DatosTransaccion struct {
	NumeroTarjeta       *string                `json:"numero_tarjeta,omitempty"`
	TipoTarjeta         *string                `json:"tipo_tarjeta,omitempty"`
	BancoEmisor         *string                `json:"banco_emisor,omitempty"`
	NumeroLote          *string                `json:"numero_lote,omitempty"`
	NumeroVoucher       *string                `json:"numero_voucher,omitempty"`
	CodigoRespuesta     *string                `json:"codigo_respuesta,omitempty"`
	MensajeRespuesta    *string                `json:"mensaje_respuesta,omitempty"`
	DatosAdicionales    map[string]interface{} `json:"datos_adicionales,omitempty"`
}

type CacheTotales struct {
	SubtotalCalculado   float64   `json:"subtotal_calculado"`
	DescuentoCalculado  float64   `json:"descuento_calculado"`
	ImpuestoCalculado   float64   `json:"impuesto_calculado"`
	TotalCalculado      float64   `json:"total_calculado"`
	FechaCalculo        time.Time `json:"fecha_calculo"`
	ValidoHasta         time.Time `json:"valido_hasta"`
}

type ConfiguracionTerminal struct {
	ImpresionAutomatica    bool                   `json:"impresion_automatica"`
	AbrirCajonAutomatico   bool                   `json:"abrir_cajon_automatico"`
	SonidosHabilitados     bool                   `json:"sonidos_habilitados"`
	TiempoEsperaCliente    int                    `json:"tiempo_espera_cliente"` // en segundos
	ConfiguracionImpresora map[string]interface{} `json:"configuracion_impresora,omitempty"`
	ConfiguracionPantalla  map[string]interface{} `json:"configuracion_pantalla,omitempty"`
	ConfiguracionRed       map[string]interface{} `json:"configuracion_red,omitempty"`
}

type MetricasTerminal struct {
	VentasHoy              int       `json:"ventas_hoy"`
	MontoVentasHoy         float64   `json:"monto_ventas_hoy"`
	TiempoPromedioVenta    float64   `json:"tiempo_promedio_venta"` // en milisegundos
	UltimaVenta            *time.Time `json:"ultima_venta,omitempty"`
	ErroresRecientes       int       `json:"errores_recientes"`
	UltimaActualizacion    time.Time `json:"ultima_actualizacion"`
}

type ConfiguracionCacheTerminal struct {
	CacheProductos     bool `json:"cache_productos"`
	CacheStock         bool `json:"cache_stock"`
	CacheClientes      bool `json:"cache_clientes"`
	TTLCache           int  `json:"ttl_cache"` // en segundos
	TamañoMaximoCache  int  `json:"tamaño_maximo_cache"` // en MB
}

// Implementar driver.Valuer para tipos JSON personalizados
func (d DatosAdicionalesVenta) Value() (driver.Value, error) {
	return json.Marshal(d)
}

func (d *DatosAdicionalesVenta) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, d)
}

func (d DatosAdicionalesDetalle) Value() (driver.Value, error) {
	return json.Marshal(d)
}

func (d *DatosAdicionalesDetalle) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, d)
}

func (d DatosTransaccion) Value() (driver.Value, error) {
	return json.Marshal(d)
}

func (d *DatosTransaccion) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, d)
}

func (c CacheTotales) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c *CacheTotales) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, c)
}

func (c ConfiguracionTerminal) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c *ConfiguracionTerminal) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, c)
}

func (m MetricasTerminal) Value() (driver.Value, error) {
	return json.Marshal(m)
}

func (m *MetricasTerminal) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, m)
}

func (c ConfiguracionCacheTerminal) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c *ConfiguracionCacheTerminal) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, c)
}

// DTOs para Venta

type VentaCreateDTO struct {
	SucursalID       uuid.UUID                `json:"sucursal_id" binding:"required"`
	TerminalID       *uuid.UUID               `json:"terminal_id,omitempty"`
	VendedorID       *uuid.UUID               `json:"vendedor_id,omitempty"`
	ClienteRut       *string                  `json:"cliente_rut,omitempty"`
	ClienteNombre    *string                  `json:"cliente_nombre,omitempty"`
	NotaVentaID      *uuid.UUID               `json:"nota_venta_id,omitempty"`
	TipoDocumento    TipoDocumento            `json:"tipo_documento" binding:"required"`
	DatosAdicionales *DatosAdicionalesVenta   `json:"datos_adicionales,omitempty"`
	DetalleVentas    []DetalleVentaCreateDTO  `json:"detalle_ventas" binding:"required,min=1"`
	MediosPago       []MedioPagoVentaCreateDTO `json:"medios_pago" binding:"required,min=1"`
}

type DetalleVentaCreateDTO struct {
	ProductoID        uuid.UUID                `json:"producto_id" binding:"required"`
	Cantidad          float64                  `json:"cantidad" binding:"required,min=0"`
	PrecioUnitario    float64                  `json:"precio_unitario" binding:"required,min=0"`
	DescuentoUnitario float64                  `json:"descuento_unitario" default:"0"`
	NumeroSerie       *string                  `json:"numero_serie,omitempty"`
	Lote              *string                  `json:"lote,omitempty"`
	FechaVencimiento  *time.Time               `json:"fecha_vencimiento,omitempty"`
	DatosAdicionales  *DatosAdicionalesDetalle `json:"datos_adicionales,omitempty"`
}

type MedioPagoVentaCreateDTO struct {
	MedioPago             MedioPago         `json:"medio_pago" binding:"required"`
	Monto                 float64           `json:"monto" binding:"required,min=0"`
	ReferenciaTransaccion *string           `json:"referencia_transaccion,omitempty"`
	CodigoAutorizacion    *string           `json:"codigo_autorizacion,omitempty"`
	DatosTransaccion      *DatosTransaccion `json:"datos_transaccion,omitempty"`
}

type VentaResponseDTO struct {
	ID                  uuid.UUID                  `json:"id"`
	NumeroVenta         int64                      `json:"numero_venta"`
	SucursalID          uuid.UUID                  `json:"sucursal_id"`
	TerminalID          *uuid.UUID                 `json:"terminal_id,omitempty"`
	CajeroID            uuid.UUID                  `json:"cajero_id"`
	VendedorID          *uuid.UUID                 `json:"vendedor_id,omitempty"`
	ClienteRut          *string                    `json:"cliente_rut,omitempty"`
	ClienteNombre       *string                    `json:"cliente_nombre,omitempty"`
	NotaVentaID         *uuid.UUID                 `json:"nota_venta_id,omitempty"`
	TipoDocumento       TipoDocumento              `json:"tipo_documento"`
	Subtotal            float64                    `json:"subtotal"`
	DescuentoTotal      float64                    `json:"descuento_total"`
	ImpuestoTotal       float64                    `json:"impuesto_total"`
	Total               float64                    `json:"total"`
	Estado              EstadoVenta                `json:"estado"`
	Fecha               time.Time                  `json:"fecha"`
	FechaAnulacion      *time.Time                 `json:"fecha_anulacion,omitempty"`
	MotivoAnulacion     *string                    `json:"motivo_anulacion,omitempty"`
	DteEmitido          bool                       `json:"dte_emitido"`
	Sincronizada        bool                       `json:"sincronizada"`
	TiempoProcesamiento *int                       `json:"tiempo_procesamiento_ms,omitempty"`
	FechaCreacion       time.Time                  `json:"fecha_creacion"`
	SucursalNombre      *string                    `json:"sucursal_nombre,omitempty"`
	CajeroNombre        *string                    `json:"cajero_nombre,omitempty"`
	VendedorNombre      *string                    `json:"vendedor_nombre,omitempty"`
	DetalleVentas       []DetalleVentaResponseDTO  `json:"detalle_ventas,omitempty"`
	MediosPago          []MedioPagoVentaResponseDTO `json:"medios_pago,omitempty"`
}

type DetalleVentaResponseDTO struct {
	ID                uuid.UUID                `json:"id"`
	ProductoID        uuid.UUID                `json:"producto_id"`
	Cantidad          float64                  `json:"cantidad"`
	PrecioUnitario    float64                  `json:"precio_unitario"`
	DescuentoUnitario float64                  `json:"descuento_unitario"`
	PrecioFinal       float64                  `json:"precio_final"`
	TotalItem         float64                  `json:"total_item"`
	NumeroSerie       *string                  `json:"numero_serie,omitempty"`
	Lote              *string                  `json:"lote,omitempty"`
	FechaVencimiento  *time.Time               `json:"fecha_vencimiento,omitempty"`
	MargenUnitario    *float64                 `json:"margen_unitario,omitempty"`
	ProductoNombre    *string                  `json:"producto_nombre,omitempty"`
	ProductoCodigo    *string                  `json:"producto_codigo,omitempty"`
}

type MedioPagoVentaResponseDTO struct {
	ID                    uuid.UUID           `json:"id"`
	MedioPago             MedioPago           `json:"medio_pago"`
	Monto                 float64             `json:"monto"`
	ReferenciaTransaccion *string             `json:"referencia_transaccion,omitempty"`
	CodigoAutorizacion    *string             `json:"codigo_autorizacion,omitempty"`
	FechaProcesamiento    time.Time           `json:"fecha_procesamiento"`
	EstadoConciliacion    EstadoConciliacion  `json:"estado_conciliacion"`
	FechaConciliacion     *time.Time          `json:"fecha_conciliacion,omitempty"`
	Comision              *float64            `json:"comision,omitempty"`
}

// Filtros para búsqueda

type VentaFilter struct {
	PaginationFilter
	SortFilter
	DateRangeFilter
	SucursalID      *uuid.UUID      `json:"sucursal_id,omitempty" form:"sucursal_id"`
	TerminalID      *uuid.UUID      `json:"terminal_id,omitempty" form:"terminal_id"`
	CajeroID        *uuid.UUID      `json:"cajero_id,omitempty" form:"cajero_id"`
	VendedorID      *uuid.UUID      `json:"vendedor_id,omitempty" form:"vendedor_id"`
	ClienteRut      *string         `json:"cliente_rut,omitempty" form:"cliente_rut"`
	TipoDocumento   *TipoDocumento  `json:"tipo_documento,omitempty" form:"tipo_documento"`
	Estado          *EstadoVenta    `json:"estado,omitempty" form:"estado"`
	MontoMin        *float64        `json:"monto_min,omitempty" form:"monto_min"`
	MontoMax        *float64        `json:"monto_max,omitempty" form:"monto_max"`
	DteEmitido      *bool           `json:"dte_emitido,omitempty" form:"dte_emitido"`
	Sincronizada    *bool           `json:"sincronizada,omitempty" form:"sincronizada"`
	NumeroVenta     *int64          `json:"numero_venta,omitempty" form:"numero_venta"`
}

// Métodos helper

func (v *Venta) CalcularTotales() {
	v.Subtotal = 0
	v.DescuentoTotal = 0
	v.ImpuestoTotal = 0
	
	for _, detalle := range v.DetalleVentas {
		v.Subtotal += detalle.PrecioUnitario * detalle.Cantidad
		v.DescuentoTotal += detalle.DescuentoUnitario * detalle.Cantidad
	}
	
	// Calcular IVA (19% en Chile)
	baseImponible := v.Subtotal - v.DescuentoTotal
	v.ImpuestoTotal = baseImponible * 0.19
	v.Total = baseImponible + v.ImpuestoTotal
}

func (v *Venta) ValidarMediosPago() error {
	totalMediosPago := 0.0
	for _, medio := range v.MediosPago {
		totalMediosPago += medio.Monto
	}
	
	if totalMediosPago != v.Total {
		return fmt.Errorf("el total de medios de pago (%.2f) no coincide con el total de la venta (%.2f)", totalMediosPago, v.Total)
	}
	
	return nil
}

func (v *Venta) PuedeAnular() bool {
	return v.Estado == EstadoVentaFinalizada && v.FechaAnulacion == nil
}

func (v *Venta) Anular(motivo string, usuarioID uuid.UUID) error {
	if !v.PuedeAnular() {
		return fmt.Errorf("la venta no puede ser anulada")
	}
	
	now := time.Now()
	v.Estado = EstadoVentaAnulada
	v.FechaAnulacion = &now
	v.MotivoAnulacion = &motivo
	v.UsuarioAnulacion = &usuarioID
	
	return nil
}

func (v *Venta) ToResponseDTO() VentaResponseDTO {
	dto := VentaResponseDTO{
		ID:                  v.ID,
		NumeroVenta:         v.NumeroVenta,
		SucursalID:          v.SucursalID,
		TerminalID:          v.TerminalID,
		CajeroID:            v.CajeroID,
		VendedorID:          v.VendedorID,
		ClienteRut:          v.ClienteRut,
		ClienteNombre:       v.ClienteNombre,
		NotaVentaID:         v.NotaVentaID,
		TipoDocumento:       v.TipoDocumento,
		Subtotal:            v.Subtotal,
		DescuentoTotal:      v.DescuentoTotal,
		ImpuestoTotal:       v.ImpuestoTotal,
		Total:               v.Total,
		Estado:              v.Estado,
		Fecha:               v.Fecha,
		FechaAnulacion:      v.FechaAnulacion,
		MotivoAnulacion:     v.MotivoAnulacion,
		DteEmitido:          v.DteEmitido,
		Sincronizada:        v.Sincronizada,
		TiempoProcesamiento: v.TiempoProcesamiento,
		FechaCreacion:       v.FechaCreacion,
	}
	
	if v.Sucursal != nil {
		dto.SucursalNombre = &v.Sucursal.Nombre
	}
	
	if v.Cajero != nil {
		nombreCajero := v.Cajero.Nombre
		if v.Cajero.Apellido != nil {
			nombreCajero += " " + *v.Cajero.Apellido
		}
		dto.CajeroNombre = &nombreCajero
	}
	
	if v.Vendedor != nil {
		nombreVendedor := v.Vendedor.Nombre
		if v.Vendedor.Apellido != nil {
			nombreVendedor += " " + *v.Vendedor.Apellido
		}
		dto.VendedorNombre = &nombreVendedor
	}
	
	// Convertir detalles
	for _, detalle := range v.DetalleVentas {
		dto.DetalleVentas = append(dto.DetalleVentas, detalle.ToResponseDTO())
	}
	
	// Convertir medios de pago
	for _, medio := range v.MediosPago {
		dto.MediosPago = append(dto.MediosPago, medio.ToResponseDTO())
	}
	
	return dto
}

func (d *DetalleVenta) ToResponseDTO() DetalleVentaResponseDTO {
	dto := DetalleVentaResponseDTO{
		ID:                d.ID,
		ProductoID:        d.ProductoID,
		Cantidad:          d.Cantidad,
		PrecioUnitario:    d.PrecioUnitario,
		DescuentoUnitario: d.DescuentoUnitario,
		PrecioFinal:       d.PrecioFinal,
		TotalItem:         d.TotalItem,
		NumeroSerie:       d.NumeroSerie,
		Lote:              d.Lote,
		FechaVencimiento:  d.FechaVencimiento,
		MargenUnitario:    d.MargenUnitario,
	}
	
	if d.Producto != nil {
		dto.ProductoNombre = &d.Producto.Descripcion
		dto.ProductoCodigo = &d.Producto.CodigoInterno
	}
	
	return dto
}

func (m *MedioPagoVenta) ToResponseDTO() MedioPagoVentaResponseDTO {
	return MedioPagoVentaResponseDTO{
		ID:                    m.ID,
		MedioPago:             m.MedioPago,
		Monto:                 m.Monto,
		ReferenciaTransaccion: m.ReferenciaTransaccion,
		CodigoAutorizacion:    m.CodigoAutorizacion,
		FechaProcesamiento:    m.FechaProcesamiento,
		EstadoConciliacion:    m.EstadoConciliacion,
		FechaConciliacion:     m.FechaConciliacion,
		Comision:              m.Comision,
	}
}

// Validaciones personalizadas

func (v *Venta) Validate() error {
	if v.Subtotal < 0 || v.DescuentoTotal < 0 || v.ImpuestoTotal < 0 || v.Total < 0 {
		return fmt.Errorf("los montos no pueden ser negativos")
	}
	
	// Validar coherencia de totales
	expectedTotal := v.Subtotal - v.DescuentoTotal + v.ImpuestoTotal
	if expectedTotal != v.Total {
		return fmt.Errorf("el total no es coherente con subtotal, descuento e impuesto")
	}
	
	return nil
}

func (d *DetalleVenta) Validate() error {
	if d.Cantidad <= 0 {
		return fmt.Errorf("cantidad debe ser mayor a 0")
	}
	
	if d.PrecioUnitario < 0 || d.DescuentoUnitario < 0 || d.PrecioFinal < 0 || d.TotalItem < 0 {
		return fmt.Errorf("los precios no pueden ser negativos")
	}
	
	// Validar coherencia de precios
	expectedPrecioFinal := d.PrecioUnitario - d.DescuentoUnitario
	if expectedPrecioFinal != d.PrecioFinal {
		return fmt.Errorf("precio final no es coherente")
	}
	
	expectedTotal := d.PrecioFinal * d.Cantidad
	if expectedTotal != d.TotalItem {
		return fmt.Errorf("total del item no es coherente")
	}
	
	return nil
}

func (m *MedioPagoVenta) Validate() error {
	if m.Monto <= 0 {
		return fmt.Errorf("monto del medio de pago debe ser mayor a 0")
	}
	
	return nil
}

