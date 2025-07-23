package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// FidelizacionCliente representa un cliente en el sistema de fidelización
type FidelizacionCliente struct {
	BaseModel
	Rut                           string                        `json:"rut" db:"rut" binding:"required" validate:"required,rut"`
	Nombre                        string                        `json:"nombre" db:"nombre" binding:"required" validate:"required,max=255"`
	Apellido                      *string                       `json:"apellido,omitempty" db:"apellido"`
	Email                         *string                       `json:"email,omitempty" db:"email" validate:"omitempty,email"`
	Telefono                      *string                       `json:"telefono,omitempty" db:"telefono"`
	FechaNacimiento               *time.Time                    `json:"fecha_nacimiento,omitempty" db:"fecha_nacimiento"`
	Direccion                     *string                       `json:"direccion,omitempty" db:"direccion"`
	Comuna                        *string                       `json:"comuna,omitempty" db:"comuna"`
	Region                        *string                       `json:"region,omitempty" db:"region"`
	PuntosActuales                int                           `json:"puntos_actuales" db:"puntos_actuales" default:"0"`
	PuntosAcumuladosTotal         int                           `json:"puntos_acumulados_total" db:"puntos_acumulados_total" default:"0"`
	NivelFidelizacion             NivelFidelizacion             `json:"nivel_fidelizacion" db:"nivel_fidelizacion" default:"bronce"`
	FechaUltimaCompra             *time.Time                    `json:"fecha_ultima_compra,omitempty" db:"fecha_ultima_compra"`
	FechaUltimaActividad          *time.Time                    `json:"fecha_ultima_actividad,omitempty" db:"fecha_ultima_actividad"`
	Activo                        bool                          `json:"activo" db:"activo" default:"true"`
	AceptaMarketing               bool                          `json:"acepta_marketing" db:"acepta_marketing" default:"false"`
	DatosAdicionales              *DatosAdicionalesCliente      `json:"datos_adicionales,omitempty" db:"datos_adicionales"`
	HashBusqueda                  *string                       `json:"-" db:"hash_busqueda"`
	QRFidelizacion                *string                       `json:"qr_fidelizacion,omitempty" db:"qr_fidelizacion"`
	CacheEstadisticas             *CacheEstadisticasCliente     `json:"cache_estadisticas,omitempty" db:"cache_estadisticas"`
	FechaProximoVencimiento       *time.Time                    `json:"fecha_proximo_vencimiento_puntos,omitempty" db:"fecha_proximo_vencimiento_puntos"`
	PuntosPorVencer               int                           `json:"puntos_por_vencer" db:"puntos_por_vencer" default:"0"`

	// Relaciones
	MovimientosFidelizacion []MovimientoFidelizacion `json:"movimientos_fidelizacion,omitempty" gorm:"foreignKey:ClienteID"`
}

// MovimientoFidelizacion representa un movimiento de puntos de fidelización
type MovimientoFidelizacion struct {
	BaseModel
	ClienteID         uuid.UUID                     `json:"cliente_id" db:"cliente_id" binding:"required"`
	SucursalID        uuid.UUID                     `json:"sucursal_id" db:"sucursal_id" binding:"required"`
	VentaID           *uuid.UUID                    `json:"venta_id,omitempty" db:"venta_id"`
	Tipo              TipoMovimientoFidelizacion    `json:"tipo" db:"tipo" binding:"required"`
	Puntos            int                           `json:"puntos" db:"puntos" binding:"required"`
	PuntosAnteriores  *int                          `json:"puntos_anteriores,omitempty" db:"puntos_anteriores"`
	PuntosNuevos      *int                          `json:"puntos_nuevos,omitempty" db:"puntos_nuevos"`
	Multiplicador     float64                       `json:"multiplicador" db:"multiplicador" default:"1.0"`
	Detalle           *string                       `json:"detalle,omitempty" db:"detalle"`
	Fecha             time.Time                     `json:"fecha" db:"fecha" default:"NOW()"`
	FechaVencimiento  *time.Time                    `json:"fecha_vencimiento,omitempty" db:"fecha_vencimiento"`
	UsuarioID         *uuid.UUID                    `json:"usuario_id,omitempty" db:"usuario_id"`
	DatosAdicionales  *DatosAdicionalesMovimiento   `json:"datos_adicionales,omitempty" db:"datos_adicionales"`
	ReglaAplicadaID   *uuid.UUID                    `json:"regla_aplicada_id,omitempty" db:"regla_aplicada_id"`
	LoteProcesamiento *uuid.UUID                    `json:"lote_procesamiento,omitempty" db:"lote_procesamiento"`

	// Relaciones
	Cliente   *FidelizacionCliente `json:"cliente,omitempty" gorm:"foreignKey:ClienteID"`
	Sucursal  *Sucursal            `json:"sucursal,omitempty" gorm:"foreignKey:SucursalID"`
	Venta     *Venta               `json:"venta,omitempty" gorm:"foreignKey:VentaID"`
	Usuario   *Usuario             `json:"usuario,omitempty" gorm:"foreignKey:UsuarioID"`
	Regla     *ReglaFidelizacion   `json:"regla,omitempty" gorm:"foreignKey:ReglaAplicadaID"`
}

// ReglaFidelizacion representa una regla del sistema de fidelización
type ReglaFidelizacion struct {
	BaseModel
	Nombre                  string                    `json:"nombre" db:"nombre" binding:"required" validate:"required,max=255"`
	Descripcion             *string                   `json:"descripcion,omitempty" db:"descripcion"`
	TipoRegla               TipoReglaFidelizacion     `json:"tipo_regla" db:"tipo_regla" binding:"required"`
	Condiciones             CondicionesRegla          `json:"condiciones" db:"condiciones" binding:"required"`
	Acciones                AccionesRegla             `json:"acciones" db:"acciones" binding:"required"`
	Activa                  bool                      `json:"activa" db:"activa" default:"true"`
	FechaInicio             *time.Time                `json:"fecha_inicio,omitempty" db:"fecha_inicio"`
	FechaFin                *time.Time                `json:"fecha_fin,omitempty" db:"fecha_fin"`
	Prioridad               int                       `json:"prioridad" db:"prioridad" default:"0"`
	CondicionesCompiladas   *CondicionesCompiladas    `json:"condiciones_compiladas,omitempty" db:"condiciones_compiladas"`
	CacheEvaluacion         *CacheEvaluacionRegla     `json:"cache_evaluacion,omitempty" db:"cache_evaluacion"`
	TotalAplicaciones       int                       `json:"total_aplicaciones" db:"total_aplicaciones" default:"0"`
	TotalPuntosOtorgados    int                       `json:"total_puntos_otorgados" db:"total_puntos_otorgados" default:"0"`

	// Relaciones
	MovimientosGenerados []MovimientoFidelizacion `json:"movimientos_generados,omitempty" gorm:"foreignKey:ReglaAplicadaID"`
}

// Enums para fidelización

type NivelFidelizacion string

const (
	NivelBronce  NivelFidelizacion = "bronce"
	NivelPlata   NivelFidelizacion = "plata"
	NivelOro     NivelFidelizacion = "oro"
	NivelPlatino NivelFidelizacion = "platino"
)

type TipoReglaFidelizacion string

const (
	ReglaAcumulacion TipoReglaFidelizacion = "acumulacion"
	ReglaCanje       TipoReglaFidelizacion = "canje"
	ReglaPromocion   TipoReglaFidelizacion = "promocion"
	ReglaNivel       TipoReglaFidelizacion = "nivel"
)

// Estructuras JSON

type DatosAdicionalesCliente struct {
	Profesion               *string                `json:"profesion,omitempty"`
	Empresa                 *string                `json:"empresa,omitempty"`
	PreferenciasContacto    []string               `json:"preferencias_contacto,omitempty"`
	InteresesProductos      []string               `json:"intereses_productos,omitempty"`
	FrecuenciaCompra        *string                `json:"frecuencia_compra,omitempty"`
	MontoPromedioCompra     *float64               `json:"monto_promedio_compra,omitempty"`
	CanalPreferido          *string                `json:"canal_preferido,omitempty"`
	Observaciones           *string                `json:"observaciones,omitempty"`
	DatosPersonalizados     map[string]interface{} `json:"datos_personalizados,omitempty"`
}

type CacheEstadisticasCliente struct {
	TotalCompras            int       `json:"total_compras"`
	MontoTotalCompras       float64   `json:"monto_total_compras"`
	PromedioCompra          float64   `json:"promedio_compra"`
	DiasConCompras          int       `json:"dias_con_compras"`
	FechaUltimaVenta        *time.Time `json:"fecha_ultima_venta,omitempty"`
	DiasSinComprar          int       `json:"dias_sin_comprar"`
	EstadoActividad         string    `json:"estado_actividad"` // ACTIVO, INACTIVO, PERDIDO
	FechaCalculado          time.Time `json:"fecha_calculado"`
	ValidoHasta             time.Time `json:"valido_hasta"`
}

type DatosAdicionalesMovimiento struct {
	MontoCompra             *float64               `json:"monto_compra,omitempty"`
	ProductosComprados      []string               `json:"productos_comprados,omitempty"`
	CanalVenta              *string                `json:"canal_venta,omitempty"`
	PromocionAplicada       *string                `json:"promocion_aplicada,omitempty"`
	ObservacionesMovimiento *string                `json:"observaciones_movimiento,omitempty"`
	MetadatosPersonalizados map[string]interface{} `json:"metadatos_personalizados,omitempty"`
}

type CondicionesRegla struct {
	MontoMinimo             *float64               `json:"monto_minimo,omitempty"`
	MontoMaximo             *float64               `json:"monto_maximo,omitempty"`
	ProductosIncluidos      []string               `json:"productos_incluidos,omitempty"`
	ProductosExcluidos      []string               `json:"productos_excluidos,omitempty"`
	CategoriasIncluidas     []string               `json:"categorias_incluidas,omitempty"`
	CategoriasExcluidas     []string               `json:"categorias_excluidas,omitempty"`
	SucursalesIncluidas     []string               `json:"sucursales_incluidas,omitempty"`
	SucursalesExcluidas     []string               `json:"sucursales_excluidas,omitempty"`
	NivelCliente            *NivelFidelizacion     `json:"nivel_cliente,omitempty"`
	DiasVigencia            *int                   `json:"dias_vigencia,omitempty"`
	HoraInicio              *string                `json:"hora_inicio,omitempty"`
	HoraFin                 *string                `json:"hora_fin,omitempty"`
	DiasSemana              []int                  `json:"dias_semana,omitempty"` // 0=Domingo, 1=Lunes, etc.
	CondicionesPersonalizadas map[string]interface{} `json:"condiciones_personalizadas,omitempty"`
}

type AccionesRegla struct {
	PuntosPorPeso           *float64               `json:"puntos_por_peso,omitempty"`
	PuntosFijos             *int                   `json:"puntos_fijos,omitempty"`
	Multiplicador           *float64               `json:"multiplicador,omitempty"`
	BonusAdicional          *int                   `json:"bonus_adicional,omitempty"`
	PorcentajeDescuento     *float64               `json:"porcentaje_descuento,omitempty"`
	MontoDescuento          *float64               `json:"monto_descuento,omitempty"`
	ProductoGratis          *string                `json:"producto_gratis,omitempty"`
	CambioNivel             *NivelFidelizacion     `json:"cambio_nivel,omitempty"`
	DiasVencimientoPuntos   *int                   `json:"dias_vencimiento_puntos,omitempty"`
	AccionesPersonalizadas  map[string]interface{} `json:"acciones_personalizadas,omitempty"`
}

type CondicionesCompiladas struct {
	CacheKey            string                 `json:"cache_key"`
	EvaluacionRapida    bool                   `json:"evaluacion_rapida"`
	CondicionesSQL      *string                `json:"condiciones_sql,omitempty"`
	ParametrosSQL       map[string]interface{} `json:"parametros_sql,omitempty"`
	FechaCompilacion    time.Time              `json:"fecha_compilacion"`
}

type CacheEvaluacionRegla struct {
	EvaluacionesRecientes   map[string]bool        `json:"evaluaciones_recientes"`
	FechaUltimaEvaluacion   time.Time              `json:"fecha_ultima_evaluacion"`
	TotalEvaluaciones       int                    `json:"total_evaluaciones"`
	TotalAplicacionesCache  int                    `json:"total_aplicaciones_cache"`
	ValidoHasta             time.Time              `json:"valido_hasta"`
}

// Implementar driver.Valuer para tipos JSON personalizados
func (d DatosAdicionalesCliente) Value() (driver.Value, error) {
	return json.Marshal(d)
}

func (d *DatosAdicionalesCliente) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, d)
}

func (c CacheEstadisticasCliente) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c *CacheEstadisticasCliente) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, c)
}

func (d DatosAdicionalesMovimiento) Value() (driver.Value, error) {
	return json.Marshal(d)
}

func (d *DatosAdicionalesMovimiento) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, d)
}

func (c CondicionesRegla) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c *CondicionesRegla) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, c)
}

func (a AccionesRegla) Value() (driver.Value, error) {
	return json.Marshal(a)
}

func (a *AccionesRegla) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, a)
}

func (c CondicionesCompiladas) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c *CondicionesCompiladas) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, c)
}

func (c CacheEvaluacionRegla) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c *CacheEvaluacionRegla) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, c)
}

// DTOs para FidelizacionCliente

type FidelizacionClienteCreateDTO struct {
	Rut                 string                   `json:"rut" binding:"required" validate:"required,rut"`
	Nombre              string                   `json:"nombre" binding:"required" validate:"required,max=255"`
	Apellido            *string                  `json:"apellido,omitempty"`
	Email               *string                  `json:"email,omitempty" validate:"omitempty,email"`
	Telefono            *string                  `json:"telefono,omitempty"`
	FechaNacimiento     *time.Time               `json:"fecha_nacimiento,omitempty"`
	Direccion           *string                  `json:"direccion,omitempty"`
	Comuna              *string                  `json:"comuna,omitempty"`
	Region              *string                  `json:"region,omitempty"`
	AceptaMarketing     bool                     `json:"acepta_marketing"`
	DatosAdicionales    *DatosAdicionalesCliente `json:"datos_adicionales,omitempty"`
}

type FidelizacionClienteUpdateDTO struct {
	Nombre              *string                  `json:"nombre,omitempty" validate:"omitempty,max=255"`
	Apellido            *string                  `json:"apellido,omitempty"`
	Email               *string                  `json:"email,omitempty" validate:"omitempty,email"`
	Telefono            *string                  `json:"telefono,omitempty"`
	FechaNacimiento     *time.Time               `json:"fecha_nacimiento,omitempty"`
	Direccion           *string                  `json:"direccion,omitempty"`
	Comuna              *string                  `json:"comuna,omitempty"`
	Region              *string                  `json:"region,omitempty"`
	Activo              *bool                    `json:"activo,omitempty"`
	AceptaMarketing     *bool                    `json:"acepta_marketing,omitempty"`
	DatosAdicionales    *DatosAdicionalesCliente `json:"datos_adicionales,omitempty"`
}

type FidelizacionClienteResponseDTO struct {
	ID                        uuid.UUID                 `json:"id"`
	Rut                       string                    `json:"rut"`
	Nombre                    string                    `json:"nombre"`
	Apellido                  *string                   `json:"apellido,omitempty"`
	Email                     *string                   `json:"email,omitempty"`
	Telefono                  *string                   `json:"telefono,omitempty"`
	FechaNacimiento           *time.Time                `json:"fecha_nacimiento,omitempty"`
	Direccion                 *string                   `json:"direccion,omitempty"`
	Comuna                    *string                   `json:"comuna,omitempty"`
	Region                    *string                   `json:"region,omitempty"`
	PuntosActuales            int                       `json:"puntos_actuales"`
	PuntosAcumuladosTotal     int                       `json:"puntos_acumulados_total"`
	NivelFidelizacion         NivelFidelizacion         `json:"nivel_fidelizacion"`
	FechaUltimaCompra         *time.Time                `json:"fecha_ultima_compra,omitempty"`
	FechaUltimaActividad      *time.Time                `json:"fecha_ultima_actividad,omitempty"`
	Activo                    bool                      `json:"activo"`
	AceptaMarketing           bool                      `json:"acepta_marketing"`
	QRFidelizacion            *string                   `json:"qr_fidelizacion,omitempty"`
	CacheEstadisticas         *CacheEstadisticasCliente `json:"cache_estadisticas,omitempty"`
	FechaProximoVencimiento   *time.Time                `json:"fecha_proximo_vencimiento_puntos,omitempty"`
	PuntosPorVencer           int                       `json:"puntos_por_vencer"`
	FechaCreacion             time.Time                 `json:"fecha_creacion"`
	FechaModificacion         time.Time                 `json:"fecha_modificacion"`
}

type FidelizacionClienteListDTO struct {
	ID                    uuid.UUID         `json:"id"`
	Rut                   string            `json:"rut"`
	Nombre                string            `json:"nombre"`
	Apellido              *string           `json:"apellido,omitempty"`
	Email                 *string           `json:"email,omitempty"`
	PuntosActuales        int               `json:"puntos_actuales"`
	NivelFidelizacion     NivelFidelizacion `json:"nivel_fidelizacion"`
	FechaUltimaCompra     *time.Time        `json:"fecha_ultima_compra,omitempty"`
	Activo                bool              `json:"activo"`
	FechaCreacion         time.Time         `json:"fecha_creacion"`
	FechaModificacion     time.Time         `json:"fecha_modificacion"`
}

// DTOs para MovimientoFidelizacion

type MovimientoFidelizacionCreateDTO struct {
	ClienteID        uuid.UUID                   `json:"cliente_id" binding:"required"`
	SucursalID       uuid.UUID                   `json:"sucursal_id" binding:"required"`
	VentaID          *uuid.UUID                  `json:"venta_id,omitempty"`
	Tipo             TipoMovimientoFidelizacion  `json:"tipo" binding:"required"`
	Puntos           int                         `json:"puntos" binding:"required"`
	Multiplicador    float64                     `json:"multiplicador" default:"1.0"`
	Detalle          *string                     `json:"detalle,omitempty"`
	FechaVencimiento *time.Time                  `json:"fecha_vencimiento,omitempty"`
	DatosAdicionales *DatosAdicionalesMovimiento `json:"datos_adicionales,omitempty"`
}

type MovimientoFidelizacionResponseDTO struct {
	ID               uuid.UUID                   `json:"id"`
	ClienteID        uuid.UUID                   `json:"cliente_id"`
	SucursalID       uuid.UUID                   `json:"sucursal_id"`
	VentaID          *uuid.UUID                  `json:"venta_id,omitempty"`
	Tipo             TipoMovimientoFidelizacion  `json:"tipo"`
	Puntos           int                         `json:"puntos"`
	PuntosAnteriores *int                        `json:"puntos_anteriores,omitempty"`
	PuntosNuevos     *int                        `json:"puntos_nuevos,omitempty"`
	Multiplicador    float64                     `json:"multiplicador"`
	Detalle          *string                     `json:"detalle,omitempty"`
	Fecha            time.Time                   `json:"fecha"`
	FechaVencimiento *time.Time                  `json:"fecha_vencimiento,omitempty"`
	UsuarioID        *uuid.UUID                  `json:"usuario_id,omitempty"`
	FechaCreacion    time.Time                   `json:"fecha_creacion"`
	ClienteNombre    *string                     `json:"cliente_nombre,omitempty"`
	SucursalNombre   *string                     `json:"sucursal_nombre,omitempty"`
	UsuarioNombre    *string                     `json:"usuario_nombre,omitempty"`
}

// DTOs para ReglaFidelizacion

type ReglaFidelizacionCreateDTO struct {
	Nombre                string                `json:"nombre" binding:"required" validate:"required,max=255"`
	Descripcion           *string               `json:"descripcion,omitempty"`
	TipoRegla             TipoReglaFidelizacion `json:"tipo_regla" binding:"required"`
	Condiciones           CondicionesRegla      `json:"condiciones" binding:"required"`
	Acciones              AccionesRegla         `json:"acciones" binding:"required"`
	FechaInicio           *time.Time            `json:"fecha_inicio,omitempty"`
	FechaFin              *time.Time            `json:"fecha_fin,omitempty"`
	Prioridad             int                   `json:"prioridad" default:"0"`
}

type ReglaFidelizacionUpdateDTO struct {
	Nombre      *string               `json:"nombre,omitempty" validate:"omitempty,max=255"`
	Descripcion *string               `json:"descripcion,omitempty"`
	Condiciones *CondicionesRegla     `json:"condiciones,omitempty"`
	Acciones    *AccionesRegla        `json:"acciones,omitempty"`
	Activa      *bool                 `json:"activa,omitempty"`
	FechaInicio *time.Time            `json:"fecha_inicio,omitempty"`
	FechaFin    *time.Time            `json:"fecha_fin,omitempty"`
	Prioridad   *int                  `json:"prioridad,omitempty"`
}

type ReglaFidelizacionResponseDTO struct {
	ID                   uuid.UUID             `json:"id"`
	Nombre               string                `json:"nombre"`
	Descripcion          *string               `json:"descripcion,omitempty"`
	TipoRegla            TipoReglaFidelizacion `json:"tipo_regla"`
	Condiciones          CondicionesRegla      `json:"condiciones"`
	Acciones             AccionesRegla         `json:"acciones"`
	Activa               bool                  `json:"activa"`
	FechaInicio          *time.Time            `json:"fecha_inicio,omitempty"`
	FechaFin             *time.Time            `json:"fecha_fin,omitempty"`
	Prioridad            int                   `json:"prioridad"`
	TotalAplicaciones    int                   `json:"total_aplicaciones"`
	TotalPuntosOtorgados int                   `json:"total_puntos_otorgados"`
	FechaCreacion        time.Time             `json:"fecha_creacion"`
	FechaModificacion    time.Time             `json:"fecha_modificacion"`
}

// Filtros para búsqueda

type FidelizacionClienteFilter struct {
	PaginationFilter
	SortFilter
	Rut               *string            `json:"rut,omitempty" form:"rut"`
	Nombre            *string            `json:"nombre,omitempty" form:"nombre"`
	Email             *string            `json:"email,omitempty" form:"email"`
	NivelFidelizacion *NivelFidelizacion `json:"nivel_fidelizacion,omitempty" form:"nivel_fidelizacion"`
	Activo            *bool              `json:"activo,omitempty" form:"activo"`
	PuntosMin         *int               `json:"puntos_min,omitempty" form:"puntos_min"`
	PuntosMax         *int               `json:"puntos_max,omitempty" form:"puntos_max"`
	Comuna            *string            `json:"comuna,omitempty" form:"comuna"`
	Region            *string            `json:"region,omitempty" form:"region"`
	AceptaMarketing   *bool              `json:"acepta_marketing,omitempty" form:"acepta_marketing"`
}

type MovimientoFidelizacionFilter struct {
	PaginationFilter
	SortFilter
	DateRangeFilter
	ClienteID       *uuid.UUID                  `json:"cliente_id,omitempty" form:"cliente_id"`
	SucursalID      *uuid.UUID                  `json:"sucursal_id,omitempty" form:"sucursal_id"`
	VentaID         *uuid.UUID                  `json:"venta_id,omitempty" form:"venta_id"`
	Tipo            *TipoMovimientoFidelizacion `json:"tipo,omitempty" form:"tipo"`
	UsuarioID       *uuid.UUID                  `json:"usuario_id,omitempty" form:"usuario_id"`
	ReglaAplicadaID *uuid.UUID                  `json:"regla_aplicada_id,omitempty" form:"regla_aplicada_id"`
	PuntosMin       *int                        `json:"puntos_min,omitempty" form:"puntos_min"`
	PuntosMax       *int                        `json:"puntos_max,omitempty" form:"puntos_max"`
}

// Métodos helper

func (c *FidelizacionCliente) CalcularNivel() NivelFidelizacion {
	switch {
	case c.PuntosAcumuladosTotal >= 10000:
		return NivelPlatino
	case c.PuntosAcumuladosTotal >= 5000:
		return NivelOro
	case c.PuntosAcumuladosTotal >= 1000:
		return NivelPlata
	default:
		return NivelBronce
	}
}

func (c *FidelizacionCliente) ActualizarNivel() bool {
	nivelAnterior := c.NivelFidelizacion
	nivelNuevo := c.CalcularNivel()
	
	if nivelAnterior != nivelNuevo {
		c.NivelFidelizacion = nivelNuevo
		return true
	}
	
	return false
}

func (c *FidelizacionCliente) PuedeUsarPuntos(puntos int) bool {
	return c.Activo && c.PuntosActuales >= puntos && puntos > 0
}

func (c *FidelizacionCliente) AgregarPuntos(puntos int) error {
	if puntos <= 0 {
		return fmt.Errorf("los puntos a agregar deben ser positivos")
	}
	
	c.PuntosActuales += puntos
	c.PuntosAcumuladosTotal += puntos
	c.ActualizarNivel()
	
	now := time.Now()
	c.FechaUltimaActividad = &now
	
	return nil
}

func (c *FidelizacionCliente) DescontarPuntos(puntos int) error {
	if !c.PuedeUsarPuntos(puntos) {
		return fmt.Errorf("no se pueden descontar %d puntos, solo tiene %d disponibles", puntos, c.PuntosActuales)
	}
	
	c.PuntosActuales -= puntos
	
	now := time.Now()
	c.FechaUltimaActividad = &now
	
	return nil
}

func (c *FidelizacionCliente) GenerarQR() string {
	// En implementación real, generar QR único
	return fmt.Sprintf("FIDELIZACION_%s_%d", c.Rut, time.Now().Unix())
}

func (c *FidelizacionCliente) ToResponseDTO() FidelizacionClienteResponseDTO {
	return FidelizacionClienteResponseDTO{
		ID:                      c.ID,
		Rut:                     c.Rut,
		Nombre:                  c.Nombre,
		Apellido:                c.Apellido,
		Email:                   c.Email,
		Telefono:                c.Telefono,
		FechaNacimiento:         c.FechaNacimiento,
		Direccion:               c.Direccion,
		Comuna:                  c.Comuna,
		Region:                  c.Region,
		PuntosActuales:          c.PuntosActuales,
		PuntosAcumuladosTotal:   c.PuntosAcumuladosTotal,
		NivelFidelizacion:       c.NivelFidelizacion,
		FechaUltimaCompra:       c.FechaUltimaCompra,
		FechaUltimaActividad:    c.FechaUltimaActividad,
		Activo:                  c.Activo,
		AceptaMarketing:         c.AceptaMarketing,
		QRFidelizacion:          c.QRFidelizacion,
		CacheEstadisticas:       c.CacheEstadisticas,
		FechaProximoVencimiento: c.FechaProximoVencimiento,
		PuntosPorVencer:         c.PuntosPorVencer,
		FechaCreacion:           c.FechaCreacion,
		FechaModificacion:       c.FechaModificacion,
	}
}

func (c *FidelizacionCliente) ToListDTO() FidelizacionClienteListDTO {
	return FidelizacionClienteListDTO{
		ID:                c.ID,
		Rut:               c.Rut,
		Nombre:            c.Nombre,
		Apellido:          c.Apellido,
		Email:             c.Email,
		PuntosActuales:    c.PuntosActuales,
		NivelFidelizacion: c.NivelFidelizacion,
		FechaUltimaCompra: c.FechaUltimaCompra,
		Activo:            c.Activo,
		FechaCreacion:     c.FechaCreacion,
		FechaModificacion: c.FechaModificacion,
	}
}

func (m *MovimientoFidelizacion) ToResponseDTO() MovimientoFidelizacionResponseDTO {
	dto := MovimientoFidelizacionResponseDTO{
		ID:               m.ID,
		ClienteID:        m.ClienteID,
		SucursalID:       m.SucursalID,
		VentaID:          m.VentaID,
		Tipo:             m.Tipo,
		Puntos:           m.Puntos,
		PuntosAnteriores: m.PuntosAnteriores,
		PuntosNuevos:     m.PuntosNuevos,
		Multiplicador:    m.Multiplicador,
		Detalle:          m.Detalle,
		Fecha:            m.Fecha,
		FechaVencimiento: m.FechaVencimiento,
		UsuarioID:        m.UsuarioID,
		FechaCreacion:    m.FechaCreacion,
	}
	
	if m.Cliente != nil {
		nombreCompleto := m.Cliente.Nombre
		if m.Cliente.Apellido != nil {
			nombreCompleto += " " + *m.Cliente.Apellido
		}
		dto.ClienteNombre = &nombreCompleto
	}
	
	if m.Sucursal != nil {
		dto.SucursalNombre = &m.Sucursal.Nombre
	}
	
	if m.Usuario != nil {
		nombreUsuario := m.Usuario.Nombre
		if m.Usuario.Apellido != nil {
			nombreUsuario += " " + *m.Usuario.Apellido
		}
		dto.UsuarioNombre = &nombreUsuario
	}
	
	return dto
}

func (r *ReglaFidelizacion) ToResponseDTO() ReglaFidelizacionResponseDTO {
	return ReglaFidelizacionResponseDTO{
		ID:                   r.ID,
		Nombre:               r.Nombre,
		Descripcion:          r.Descripcion,
		TipoRegla:            r.TipoRegla,
		Condiciones:          r.Condiciones,
		Acciones:             r.Acciones,
		Activa:               r.Activa,
		FechaInicio:          r.FechaInicio,
		FechaFin:             r.FechaFin,
		Prioridad:            r.Prioridad,
		TotalAplicaciones:    r.TotalAplicaciones,
		TotalPuntosOtorgados: r.TotalPuntosOtorgados,
		FechaCreacion:        r.FechaCreacion,
		FechaModificacion:    r.FechaModificacion,
	}
}

func (dto *FidelizacionClienteCreateDTO) ToModel() *FidelizacionCliente {
	cliente := &FidelizacionCliente{
		Rut:                   dto.Rut,
		Nombre:                dto.Nombre,
		Apellido:              dto.Apellido,
		Email:                 dto.Email,
		Telefono:              dto.Telefono,
		FechaNacimiento:       dto.FechaNacimiento,
		Direccion:             dto.Direccion,
		Comuna:                dto.Comuna,
		Region:                dto.Region,
		PuntosActuales:        0,
		PuntosAcumuladosTotal: 0,
		NivelFidelizacion:     NivelBronce,
		Activo:                true,
		AceptaMarketing:       dto.AceptaMarketing,
		DatosAdicionales:      dto.DatosAdicionales,
	}
	
	// Generar QR de fidelización
	qr := cliente.GenerarQR()
	cliente.QRFidelizacion = &qr
	
	return cliente
}

// Validaciones personalizadas

func (c *FidelizacionCliente) Validate() error {
	if c.PuntosActuales < 0 {
		return fmt.Errorf("puntos actuales no pueden ser negativos")
	}
	
	if c.PuntosAcumuladosTotal < 0 {
		return fmt.Errorf("puntos acumulados totales no pueden ser negativos")
	}
	
	if c.PuntosPorVencer < 0 {
		return fmt.Errorf("puntos por vencer no pueden ser negativos")
	}
	
	return nil
}

func (m *MovimientoFidelizacion) Validate() error {
	if m.Puntos == 0 {
		return fmt.Errorf("puntos del movimiento no pueden ser cero")
	}
	
	if m.Multiplicador <= 0 {
		return fmt.Errorf("multiplicador debe ser mayor a cero")
	}
	
	return nil
}

func (r *ReglaFidelizacion) Validate() error {
	if r.FechaInicio != nil && r.FechaFin != nil && r.FechaInicio.After(*r.FechaFin) {
		return fmt.Errorf("fecha de inicio no puede ser posterior a fecha de fin")
	}
	
	return nil
}

// Funciones de utilidad

func CalcularPuntosPorCompra(montoCompra float64, puntosPorPeso float64, multiplicador float64) int {
	puntos := montoCompra * puntosPorPeso * multiplicador
	return int(puntos)
}

func CalcularDescuentoPorPuntos(puntos int, valorPunto float64) float64 {
	return float64(puntos) * valorPunto
}

