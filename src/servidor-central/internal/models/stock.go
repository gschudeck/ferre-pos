package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// StockCentral representa el stock centralizado por producto y sucursal
type StockCentral struct {
	ProductoID           uuid.UUID              `json:"producto_id" db:"producto_id" binding:"required"`
	SucursalID           uuid.UUID              `json:"sucursal_id" db:"sucursal_id" binding:"required"`
	Cantidad             int                    `json:"cantidad" db:"cantidad" default:"0"`
	CantidadReservada    int                    `json:"cantidad_reservada" db:"cantidad_reservada" default:"0"`
	CantidadDisponible   int                    `json:"cantidad_disponible" db:"cantidad_disponible"` // Campo calculado
	CostoPromedio        *float64               `json:"costo_promedio,omitempty" db:"costo_promedio"`
	FechaUltimaEntrada   *time.Time             `json:"fecha_ultima_entrada,omitempty" db:"fecha_ultima_entrada"`
	FechaUltimaSalida    *time.Time             `json:"fecha_ultima_salida,omitempty" db:"fecha_ultima_salida"`
	FechaSync            time.Time              `json:"fecha_sync" db:"fecha_sync" default:"NOW()"`
	VersionOptimisticLock int                   `json:"version_optimistic_lock" db:"version_optimistic_lock" default:"1"`
	CacheValidez         time.Time              `json:"cache_validez" db:"cache_validez" default:"NOW()"`
	AlertasConfiguradas  *AlertasStock          `json:"alertas_configuradas,omitempty" db:"alertas_configuradas"`

	// Relaciones
	Producto *Producto `json:"producto,omitempty" gorm:"foreignKey:ProductoID"`
	Sucursal *Sucursal `json:"sucursal,omitempty" gorm:"foreignKey:SucursalID"`
}

// MovimientoStock representa un movimiento de stock
type MovimientoStock struct {
	BaseModel
	ProductoID          uuid.UUID         `json:"producto_id" db:"producto_id" binding:"required"`
	SucursalID          uuid.UUID         `json:"sucursal_id" db:"sucursal_id" binding:"required"`
	TipoMovimiento      TipoMovimiento    `json:"tipo_movimiento" db:"tipo_movimiento" binding:"required"`
	Cantidad            int               `json:"cantidad" db:"cantidad" binding:"required"`
	CantidadAnterior    *int              `json:"cantidad_anterior,omitempty" db:"cantidad_anterior"`
	CantidadNueva       *int              `json:"cantidad_nueva,omitempty" db:"cantidad_nueva"`
	CostoUnitario       *float64          `json:"costo_unitario,omitempty" db:"costo_unitario"`
	DocumentoReferencia *string           `json:"documento_referencia,omitempty" db:"documento_referencia"`
	UsuarioID           *uuid.UUID        `json:"usuario_id,omitempty" db:"usuario_id"`
	Fecha               time.Time         `json:"fecha" db:"fecha" default:"NOW()"`
	Observaciones       *string           `json:"observaciones,omitempty" db:"observaciones"`
	DatosAdicionales    *DatosAdicionales `json:"datos_adicionales,omitempty" db:"datos_adicionales"`
	ProcesoOrigen       *PrioridadProceso `json:"proceso_origen,omitempty" db:"proceso_origen"`
	BatchID             *uuid.UUID        `json:"batch_id,omitempty" db:"batch_id"`

	// Relaciones
	Producto *Producto `json:"producto,omitempty" gorm:"foreignKey:ProductoID"`
	Sucursal *Sucursal `json:"sucursal,omitempty" gorm:"foreignKey:SucursalID"`
	Usuario  *Usuario  `json:"usuario,omitempty" gorm:"foreignKey:UsuarioID"`
}

// TipoMovimiento define los tipos de movimiento de stock
type TipoMovimiento string

const (
	TipoEntrada             TipoMovimiento = "entrada"
	TipoSalida              TipoMovimiento = "salida"
	TipoAjuste              TipoMovimiento = "ajuste"
	TipoTransferenciaEntrada TipoMovimiento = "transferencia_entrada"
	TipoTransferenciaSalida  TipoMovimiento = "transferencia_salida"
	TipoVenta               TipoMovimiento = "venta"
	TipoDevolucion          TipoMovimiento = "devolucion"
)

// AlertasStock contiene configuración de alertas de stock
type AlertasStock struct {
	AlertaStockMinimo    bool    `json:"alerta_stock_minimo"`
	AlertaStockCritico   bool    `json:"alerta_stock_critico"`
	UmbralCritico        int     `json:"umbral_critico"`
	NotificarEmail       bool    `json:"notificar_email"`
	NotificarSistema     bool    `json:"notificar_sistema"`
	EmailsNotificacion   []string `json:"emails_notificacion,omitempty"`
	FrecuenciaNotificacion string `json:"frecuencia_notificacion"` // "inmediata", "diaria", "semanal"
}

// DatosAdicionales contiene información adicional del movimiento
type DatosAdicionales struct {
	Lote              *string                `json:"lote,omitempty"`
	NumeroSerie       *string                `json:"numero_serie,omitempty"`
	FechaVencimiento  *time.Time             `json:"fecha_vencimiento,omitempty"`
	Proveedor         *string                `json:"proveedor,omitempty"`
	NumeroFactura     *string                `json:"numero_factura,omitempty"`
	Ubicacion         *string                `json:"ubicacion,omitempty"`
	Motivo            *string                `json:"motivo,omitempty"`
	UsuarioAutoriza   *uuid.UUID             `json:"usuario_autoriza,omitempty"`
	DatosPersonalizados map[string]interface{} `json:"datos_personalizados,omitempty"`
}

// Implementar driver.Valuer para tipos JSON personalizados
func (a AlertasStock) Value() (driver.Value, error) {
	return json.Marshal(a)
}

func (a *AlertasStock) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, a)
}

func (d DatosAdicionales) Value() (driver.Value, error) {
	return json.Marshal(d)
}

func (d *DatosAdicionales) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, d)
}

// DTOs para StockCentral

type StockCentralDTO struct {
	ProductoID         uuid.UUID     `json:"producto_id"`
	SucursalID         uuid.UUID     `json:"sucursal_id"`
	Cantidad           int           `json:"cantidad"`
	CantidadReservada  int           `json:"cantidad_reservada"`
	CantidadDisponible int           `json:"cantidad_disponible"`
	CostoPromedio      *float64      `json:"costo_promedio,omitempty"`
	FechaUltimaEntrada *time.Time    `json:"fecha_ultima_entrada,omitempty"`
	FechaUltimaSalida  *time.Time    `json:"fecha_ultima_salida,omitempty"`
	FechaSync          time.Time     `json:"fecha_sync"`
	EstadoStock        EstadoStock   `json:"estado_stock"`
	ProductoNombre     *string       `json:"producto_nombre,omitempty"`
	SucursalNombre     *string       `json:"sucursal_nombre,omitempty"`
}

type StockUpdateDTO struct {
	Cantidad          *int                  `json:"cantidad,omitempty" validate:"omitempty,min=0"`
	CantidadReservada *int                  `json:"cantidad_reservada,omitempty" validate:"omitempty,min=0"`
	CostoPromedio     *float64              `json:"costo_promedio,omitempty" validate:"omitempty,min=0"`
	AlertasConfiguradas *AlertasStock       `json:"alertas_configuradas,omitempty"`
}

type StockAjusteDTO struct {
	ProductoID      uuid.UUID `json:"producto_id" binding:"required"`
	SucursalID      uuid.UUID `json:"sucursal_id" binding:"required"`
	CantidadNueva   int       `json:"cantidad_nueva" binding:"required,min=0"`
	Motivo          string    `json:"motivo" binding:"required"`
	Observaciones   *string   `json:"observaciones,omitempty"`
	CostoUnitario   *float64  `json:"costo_unitario,omitempty" validate:"omitempty,min=0"`
}

type StockTransferenciaDTO struct {
	ProductoID        uuid.UUID `json:"producto_id" binding:"required"`
	SucursalOrigenID  uuid.UUID `json:"sucursal_origen_id" binding:"required"`
	SucursalDestinoID uuid.UUID `json:"sucursal_destino_id" binding:"required"`
	Cantidad          int       `json:"cantidad" binding:"required,min=1"`
	Motivo            string    `json:"motivo" binding:"required"`
	Observaciones     *string   `json:"observaciones,omitempty"`
}

// DTOs para MovimientoStock

type MovimientoStockCreateDTO struct {
	ProductoID          uuid.UUID         `json:"producto_id" binding:"required"`
	SucursalID          uuid.UUID         `json:"sucursal_id" binding:"required"`
	TipoMovimiento      TipoMovimiento    `json:"tipo_movimiento" binding:"required"`
	Cantidad            int               `json:"cantidad" binding:"required"`
	CostoUnitario       *float64          `json:"costo_unitario,omitempty" validate:"omitempty,min=0"`
	DocumentoReferencia *string           `json:"documento_referencia,omitempty"`
	Observaciones       *string           `json:"observaciones,omitempty"`
	DatosAdicionales    *DatosAdicionales `json:"datos_adicionales,omitempty"`
}

type MovimientoStockResponseDTO struct {
	ID                  uuid.UUID         `json:"id"`
	ProductoID          uuid.UUID         `json:"producto_id"`
	SucursalID          uuid.UUID         `json:"sucursal_id"`
	TipoMovimiento      TipoMovimiento    `json:"tipo_movimiento"`
	Cantidad            int               `json:"cantidad"`
	CantidadAnterior    *int              `json:"cantidad_anterior,omitempty"`
	CantidadNueva       *int              `json:"cantidad_nueva,omitempty"`
	CostoUnitario       *float64          `json:"costo_unitario,omitempty"`
	DocumentoReferencia *string           `json:"documento_referencia,omitempty"`
	UsuarioID           *uuid.UUID        `json:"usuario_id,omitempty"`
	Fecha               time.Time         `json:"fecha"`
	Observaciones       *string           `json:"observaciones,omitempty"`
	DatosAdicionales    *DatosAdicionales `json:"datos_adicionales,omitempty"`
	ProcesoOrigen       *PrioridadProceso `json:"proceso_origen,omitempty"`
	FechaCreacion       time.Time         `json:"fecha_creacion"`
	ProductoNombre      *string           `json:"producto_nombre,omitempty"`
	SucursalNombre      *string           `json:"sucursal_nombre,omitempty"`
	UsuarioNombre       *string           `json:"usuario_nombre,omitempty"`
}

type MovimientoStockListDTO struct {
	ID                  uuid.UUID      `json:"id"`
	ProductoID          uuid.UUID      `json:"producto_id"`
	SucursalID          uuid.UUID      `json:"sucursal_id"`
	TipoMovimiento      TipoMovimiento `json:"tipo_movimiento"`
	Cantidad            int            `json:"cantidad"`
	CostoUnitario       *float64       `json:"costo_unitario,omitempty"`
	DocumentoReferencia *string        `json:"documento_referencia,omitempty"`
	Fecha               time.Time      `json:"fecha"`
	ProductoNombre      *string        `json:"producto_nombre,omitempty"`
	SucursalNombre      *string        `json:"sucursal_nombre,omitempty"`
	UsuarioNombre       *string        `json:"usuario_nombre,omitempty"`
}

// EstadoStock define los estados posibles del stock
type EstadoStock string

const (
	EstadoSinStock EstadoStock = "SIN_STOCK"
	EstadoCritico  EstadoStock = "CRITICO"
	EstadoBajo     EstadoStock = "BAJO"
	EstadoNormal   EstadoStock = "NORMAL"
)

// Filtros para búsqueda

type StockFilter struct {
	PaginationFilter
	SortFilter
	ProductoID     *uuid.UUID   `json:"producto_id,omitempty" form:"producto_id"`
	SucursalID     *uuid.UUID   `json:"sucursal_id,omitempty" form:"sucursal_id"`
	EstadoStock    *EstadoStock `json:"estado_stock,omitempty" form:"estado_stock"`
	CantidadMin    *int         `json:"cantidad_min,omitempty" form:"cantidad_min"`
	CantidadMax    *int         `json:"cantidad_max,omitempty" form:"cantidad_max"`
	StockBajo      *bool        `json:"stock_bajo,omitempty" form:"stock_bajo"`
	SinStock       *bool        `json:"sin_stock,omitempty" form:"sin_stock"`
	ConMovimientos *bool        `json:"con_movimientos,omitempty" form:"con_movimientos"`
}

type MovimientoStockFilter struct {
	PaginationFilter
	SortFilter
	DateRangeFilter
	ProductoID      *uuid.UUID      `json:"producto_id,omitempty" form:"producto_id"`
	SucursalID      *uuid.UUID      `json:"sucursal_id,omitempty" form:"sucursal_id"`
	TipoMovimiento  *TipoMovimiento `json:"tipo_movimiento,omitempty" form:"tipo_movimiento"`
	UsuarioID       *uuid.UUID      `json:"usuario_id,omitempty" form:"usuario_id"`
	ProcesoOrigen   *PrioridadProceso `json:"proceso_origen,omitempty" form:"proceso_origen"`
	BatchID         *uuid.UUID      `json:"batch_id,omitempty" form:"batch_id"`
	DocumentoReferencia *string     `json:"documento_referencia,omitempty" form:"documento_referencia"`
}

// Métodos helper

func (s *StockCentral) CalcularEstadoStock(stockMinimo int) EstadoStock {
	if s.CantidadDisponible <= 0 {
		return EstadoSinStock
	}
	if s.CantidadDisponible <= stockMinimo {
		return EstadoCritico
	}
	if s.CantidadDisponible <= stockMinimo*2 {
		return EstadoBajo
	}
	return EstadoNormal
}

func (s *StockCentral) PuedeReservar(cantidad int) bool {
	return s.CantidadDisponible >= cantidad
}

func (s *StockCentral) ReservarStock(cantidad int) error {
	if !s.PuedeReservar(cantidad) {
		return fmt.Errorf("stock insuficiente para reservar %d unidades", cantidad)
	}
	s.CantidadReservada += cantidad
	return nil
}

func (s *StockCentral) LiberarReserva(cantidad int) error {
	if s.CantidadReservada < cantidad {
		return fmt.Errorf("no se puede liberar %d unidades, solo hay %d reservadas", cantidad, s.CantidadReservada)
	}
	s.CantidadReservada -= cantidad
	return nil
}

func (s *StockCentral) ActualizarStock(cantidad int, tipoMovimiento TipoMovimiento) error {
	cantidadAnterior := s.Cantidad
	
	switch tipoMovimiento {
	case TipoEntrada, TipoTransferenciaEntrada, TipoDevolucion:
		s.Cantidad += cantidad
	case TipoSalida, TipoTransferenciaSalida, TipoVenta:
		if s.Cantidad < cantidad {
			return fmt.Errorf("stock insuficiente para descontar %d unidades", cantidad)
		}
		s.Cantidad -= cantidad
	case TipoAjuste:
		s.Cantidad = cantidad
	default:
		return fmt.Errorf("tipo de movimiento no válido: %s", tipoMovimiento)
	}
	
	// Actualizar fechas según el tipo de movimiento
	now := time.Now()
	if tipoMovimiento == TipoEntrada || tipoMovimiento == TipoTransferenciaEntrada || tipoMovimiento == TipoDevolucion {
		s.FechaUltimaEntrada = &now
	} else if tipoMovimiento == TipoSalida || tipoMovimiento == TipoTransferenciaSalida || tipoMovimiento == TipoVenta {
		s.FechaUltimaSalida = &now
	}
	
	s.FechaSync = now
	s.VersionOptimisticLock++
	s.CacheValidez = now
	
	return nil
}

func (s *StockCentral) ToDTO() StockCentralDTO {
	dto := StockCentralDTO{
		ProductoID:         s.ProductoID,
		SucursalID:         s.SucursalID,
		Cantidad:           s.Cantidad,
		CantidadReservada:  s.CantidadReservada,
		CantidadDisponible: s.CantidadDisponible,
		CostoPromedio:      s.CostoPromedio,
		FechaUltimaEntrada: s.FechaUltimaEntrada,
		FechaUltimaSalida:  s.FechaUltimaSalida,
		FechaSync:          s.FechaSync,
	}
	
	// Calcular estado del stock (necesitaríamos el stock mínimo del producto)
	dto.EstadoStock = s.CalcularEstadoStock(0) // Por defecto, se puede mejorar
	
	if s.Producto != nil {
		dto.ProductoNombre = &s.Producto.Descripcion
		dto.EstadoStock = s.CalcularEstadoStock(s.Producto.StockMinimo)
	}
	
	if s.Sucursal != nil {
		dto.SucursalNombre = &s.Sucursal.Nombre
	}
	
	return dto
}

func (m *MovimientoStock) ToResponseDTO() MovimientoStockResponseDTO {
	dto := MovimientoStockResponseDTO{
		ID:                  m.ID,
		ProductoID:          m.ProductoID,
		SucursalID:          m.SucursalID,
		TipoMovimiento:      m.TipoMovimiento,
		Cantidad:            m.Cantidad,
		CantidadAnterior:    m.CantidadAnterior,
		CantidadNueva:       m.CantidadNueva,
		CostoUnitario:       m.CostoUnitario,
		DocumentoReferencia: m.DocumentoReferencia,
		UsuarioID:           m.UsuarioID,
		Fecha:               m.Fecha,
		Observaciones:       m.Observaciones,
		DatosAdicionales:    m.DatosAdicionales,
		ProcesoOrigen:       m.ProcesoOrigen,
		FechaCreacion:       m.FechaCreacion,
	}
	
	if m.Producto != nil {
		dto.ProductoNombre = &m.Producto.Descripcion
	}
	
	if m.Sucursal != nil {
		dto.SucursalNombre = &m.Sucursal.Nombre
	}
	
	if m.Usuario != nil {
		nombreCompleto := m.Usuario.Nombre
		if m.Usuario.Apellido != nil {
			nombreCompleto += " " + *m.Usuario.Apellido
		}
		dto.UsuarioNombre = &nombreCompleto
	}
	
	return dto
}

func (m *MovimientoStock) ToListDTO() MovimientoStockListDTO {
	dto := MovimientoStockListDTO{
		ID:                  m.ID,
		ProductoID:          m.ProductoID,
		SucursalID:          m.SucursalID,
		TipoMovimiento:      m.TipoMovimiento,
		Cantidad:            m.Cantidad,
		CostoUnitario:       m.CostoUnitario,
		DocumentoReferencia: m.DocumentoReferencia,
		Fecha:               m.Fecha,
	}
	
	if m.Producto != nil {
		dto.ProductoNombre = &m.Producto.Descripcion
	}
	
	if m.Sucursal != nil {
		dto.SucursalNombre = &m.Sucursal.Nombre
	}
	
	if m.Usuario != nil {
		nombreCompleto := m.Usuario.Nombre
		if m.Usuario.Apellido != nil {
			nombreCompleto += " " + *m.Usuario.Apellido
		}
		dto.UsuarioNombre = &nombreCompleto
	}
	
	return dto
}

func (dto *MovimientoStockCreateDTO) ToModel(usuarioID uuid.UUID) *MovimientoStock {
	return &MovimientoStock{
		ProductoID:          dto.ProductoID,
		SucursalID:          dto.SucursalID,
		TipoMovimiento:      dto.TipoMovimiento,
		Cantidad:            dto.Cantidad,
		CostoUnitario:       dto.CostoUnitario,
		DocumentoReferencia: dto.DocumentoReferencia,
		UsuarioID:           &usuarioID,
		Fecha:               time.Now(),
		Observaciones:       dto.Observaciones,
		DatosAdicionales:    dto.DatosAdicionales,
		ProcesoOrigen:       &PrioridadMaxima, // Por defecto, movimientos manuales son de máxima prioridad
	}
}

// Validaciones personalizadas

func (s *StockCentral) Validate() error {
	if s.Cantidad < 0 {
		return fmt.Errorf("cantidad no puede ser negativa")
	}
	
	if s.CantidadReservada < 0 {
		return fmt.Errorf("cantidad reservada no puede ser negativa")
	}
	
	if s.CantidadReservada > s.Cantidad {
		return fmt.Errorf("cantidad reservada no puede ser mayor a la cantidad total")
	}
	
	return nil
}

func (m *MovimientoStock) Validate() error {
	if m.Cantidad == 0 {
		return fmt.Errorf("cantidad del movimiento no puede ser cero")
	}
	
	if m.CostoUnitario != nil && *m.CostoUnitario < 0 {
		return fmt.Errorf("costo unitario no puede ser negativo")
	}
	
	// Validar tipos de movimiento
	validTipos := []TipoMovimiento{
		TipoEntrada, TipoSalida, TipoAjuste, TipoTransferenciaEntrada,
		TipoTransferenciaSalida, TipoVenta, TipoDevolucion,
	}
	
	valid := false
	for _, tipo := range validTipos {
		if m.TipoMovimiento == tipo {
			valid = true
			break
		}
	}
	
	if !valid {
		return fmt.Errorf("tipo de movimiento no válido: %s", m.TipoMovimiento)
	}
	
	return nil
}

// Funciones de utilidad para stock

func CalcularValorStock(cantidad int, costoPromedio float64) float64 {
	return float64(cantidad) * costoPromedio
}

func CalcularCostoPromedioNuevo(cantidadActual int, costoActual float64, cantidadNueva int, costoNuevo float64) float64 {
	if cantidadActual+cantidadNueva == 0 {
		return 0
	}
	
	valorActual := float64(cantidadActual) * costoActual
	valorNuevo := float64(cantidadNueva) * costoNuevo
	cantidadTotal := cantidadActual + cantidadNueva
	
	return (valorActual + valorNuevo) / float64(cantidadTotal)
}

