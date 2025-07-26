package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Sucursal representa una sucursal de la ferretería
type Sucursal struct {
	BaseModel
	Codigo                    string               `json:"codigo" db:"codigo" binding:"required" validate:"required,max=50"`
	Nombre                    string               `json:"nombre" db:"nombre" binding:"required" validate:"required,max=255"`
	Direccion                 *string              `json:"direccion,omitempty" db:"direccion"`
	Comuna                    *string              `json:"comuna,omitempty" db:"comuna"`
	Region                    *string              `json:"region,omitempty" db:"region"`
	Telefono                  *string              `json:"telefono,omitempty" db:"telefono"`
	Email                     *string              `json:"email,omitempty" db:"email" validate:"omitempty,email"`
	HorarioApertura           *time.Time           `json:"horario_apertura,omitempty" db:"horario_apertura"`
	HorarioCierre             *time.Time           `json:"horario_cierre,omitempty" db:"horario_cierre"`
	Timezone                  string               `json:"timezone" db:"timezone" default:"America/Santiago"`
	Habilitada                bool                 `json:"habilitada" db:"habilitada" default:"true"`
	ConfiguracionDTE          *ConfiguracionDTE    `json:"configuracion_dte,omitempty" db:"configuracion_dte"`
	ConfiguracionPagos        *ConfiguracionPagos  `json:"configuracion_pagos,omitempty" db:"configuracion_pagos"`
	MaxConexionesConcurrentes int                  `json:"max_conexiones_concurrentes" db:"max_conexiones_concurrentes" default:"50"`
	ConfiguracionCache        *ConfiguracionCache  `json:"configuracion_cache,omitempty" db:"configuracion_cache"`
	MetricasRendimiento       *MetricasRendimiento `json:"metricas_rendimiento,omitempty" db:"metricas_rendimiento"`
}

// ConfiguracionDTE contiene la configuración de documentos tributarios electrónicos
type ConfiguracionDTE struct {
	RutEmpresa      string `json:"rut_empresa"`
	RazonSocial     string `json:"razon_social"`
	Giro            string `json:"giro"`
	DireccionFiscal string `json:"direccion_fiscal"`
	ComunaFiscal    string `json:"comuna_fiscal"`
	CiudadFiscal    string `json:"ciudad_fiscal"`
	CodigoSII       string `json:"codigo_sii"`
	ResolucionSII   string `json:"resolucion_sii"`
	FechaResolucion string `json:"fecha_resolucion"`
	Ambiente        string `json:"ambiente"` // "certificacion" o "produccion"
}

// ConfiguracionPagos contiene la configuración de medios de pago
type ConfiguracionPagos struct {
	AceptaEfectivo           bool                   `json:"acepta_efectivo"`
	AceptaTarjetaDebito      bool                   `json:"acepta_tarjeta_debito"`
	AceptaTarjetaCredito     bool                   `json:"acepta_tarjeta_credito"`
	AceptaTransferencia      bool                   `json:"acepta_transferencia"`
	AceptaCheque             bool                   `json:"acepta_cheque"`
	AceptaPuntosFidelizacion bool                   `json:"acepta_puntos_fidelizacion"`
	ConfiguracionPOS         map[string]interface{} `json:"configuracion_pos,omitempty"`
	LimiteEfectivo           *float64               `json:"limite_efectivo,omitempty"`
	ComisionesTarjetas       map[string]float64     `json:"comisiones_tarjetas,omitempty"`
}

// ConfiguracionCache contiene la configuración de cache para la sucursal
type ConfiguracionCache struct {
	TTLProductos       int  `json:"ttl_productos"`    // TTL en segundos
	TTLStock           int  `json:"ttl_stock"`        // TTL en segundos
	TTLFidelizacion    int  `json:"ttl_fidelizacion"` // TTL en segundos
	CacheHabilitado    bool `json:"cache_habilitado"`
	TamañoMaximoCache  int  `json:"tamaño_maximo_cache"` // En MB
	LimpiezaAutomatica bool `json:"limpieza_automatica"`
}

// MetricasRendimiento contiene métricas de rendimiento de la sucursal
type MetricasRendimiento struct {
	VentasPromedioDiario   float64            `json:"ventas_promedio_diario"`
	TiempoPromedioVenta    float64            `json:"tiempo_promedio_venta"` // En milisegundos
	ConexionesActivas      int                `json:"conexiones_activas"`
	ConexionesMaximas      int                `json:"conexiones_maximas"`
	UsoMemoria             float64            `json:"uso_memoria"` // En MB
	UsoCPU                 float64            `json:"uso_cpu"`     // En porcentaje
	UltimaActualizacion    time.Time          `json:"ultima_actualizacion"`
	MetricasPersonalizadas map[string]float64 `json:"metricas_personalizadas,omitempty"`
}

// Implementar driver.Valuer para tipos JSON personalizados
func (c ConfiguracionDTE) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c *ConfiguracionDTE) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, c)
}

func (c ConfiguracionPagos) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c *ConfiguracionPagos) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, c)
}

func (c ConfiguracionCache) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c *ConfiguracionCache) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, c)
}

func (m MetricasRendimiento) Value() (driver.Value, error) {
	return json.Marshal(m)
}

func (m *MetricasRendimiento) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, m)
}

// DTOs para Sucursal

type SucursalCreateDTO struct {
	Codigo                    string              `json:"codigo" binding:"required" validate:"required,max=50"`
	Nombre                    string              `json:"nombre" binding:"required" validate:"required,max=255"`
	Direccion                 *string             `json:"direccion,omitempty"`
	Comuna                    *string             `json:"comuna,omitempty"`
	Region                    *string             `json:"region,omitempty"`
	Telefono                  *string             `json:"telefono,omitempty"`
	Email                     *string             `json:"email,omitempty" validate:"omitempty,email"`
	HorarioApertura           *time.Time          `json:"horario_apertura,omitempty"`
	HorarioCierre             *time.Time          `json:"horario_cierre,omitempty"`
	Timezone                  string              `json:"timezone"`
	ConfiguracionDTE          *ConfiguracionDTE   `json:"configuracion_dte,omitempty"`
	ConfiguracionPagos        *ConfiguracionPagos `json:"configuracion_pagos,omitempty"`
	MaxConexionesConcurrentes int                 `json:"max_conexiones_concurrentes"`
	ConfiguracionCache        *ConfiguracionCache `json:"configuracion_cache,omitempty"`
}

type SucursalUpdateDTO struct {
	Nombre                    *string             `json:"nombre,omitempty" validate:"omitempty,max=255"`
	Direccion                 *string             `json:"direccion,omitempty"`
	Comuna                    *string             `json:"comuna,omitempty"`
	Region                    *string             `json:"region,omitempty"`
	Telefono                  *string             `json:"telefono,omitempty"`
	Email                     *string             `json:"email,omitempty" validate:"omitempty,email"`
	HorarioApertura           *time.Time          `json:"horario_apertura,omitempty"`
	HorarioCierre             *time.Time          `json:"horario_cierre,omitempty"`
	Timezone                  *string             `json:"timezone,omitempty"`
	Habilitada                *bool               `json:"habilitada,omitempty"`
	ConfiguracionDTE          *ConfiguracionDTE   `json:"configuracion_dte,omitempty"`
	ConfiguracionPagos        *ConfiguracionPagos `json:"configuracion_pagos,omitempty"`
	MaxConexionesConcurrentes *int                `json:"max_conexiones_concurrentes,omitempty"`
	ConfiguracionCache        *ConfiguracionCache `json:"configuracion_cache,omitempty"`
}

type SucursalResponseDTO struct {
	ID                        uuid.UUID            `json:"id"`
	Codigo                    string               `json:"codigo"`
	Nombre                    string               `json:"nombre"`
	Direccion                 *string              `json:"direccion,omitempty"`
	Comuna                    *string              `json:"comuna,omitempty"`
	Region                    *string              `json:"region,omitempty"`
	Telefono                  *string              `json:"telefono,omitempty"`
	Email                     *string              `json:"email,omitempty"`
	HorarioApertura           *time.Time           `json:"horario_apertura,omitempty"`
	HorarioCierre             *time.Time           `json:"horario_cierre,omitempty"`
	Timezone                  string               `json:"timezone"`
	Habilitada                bool                 `json:"habilitada"`
	MaxConexionesConcurrentes int                  `json:"max_conexiones_concurrentes"`
	MetricasRendimiento       *MetricasRendimiento `json:"metricas_rendimiento,omitempty"`
	FechaCreacion             time.Time            `json:"fecha_creacion"`
	FechaModificacion         time.Time            `json:"fecha_modificacion"`
}

type SucursalListDTO struct {
	ID                uuid.UUID `json:"id"`
	Codigo            string    `json:"codigo"`
	Nombre            string    `json:"nombre"`
	Comuna            *string   `json:"comuna,omitempty"`
	Region            *string   `json:"region,omitempty"`
	Habilitada        bool      `json:"habilitada"`
	FechaCreacion     time.Time `json:"fecha_creacion"`
	FechaModificacion time.Time `json:"fecha_modificacion"`
}

// Filtros para búsqueda de sucursales

type SucursalFilter struct {
	PaginationFilter
	SortFilter
	Codigo     *string `json:"codigo,omitempty" form:"codigo"`
	Nombre     *string `json:"nombre,omitempty" form:"nombre"`
	Comuna     *string `json:"comuna,omitempty" form:"comuna"`
	Region     *string `json:"region,omitempty" form:"region"`
	Habilitada *bool   `json:"habilitada,omitempty" form:"habilitada"`
}

// Métodos helper

func (s *Sucursal) ToResponseDTO() SucursalResponseDTO {
	return SucursalResponseDTO{
		ID:                        s.ID,
		Codigo:                    s.Codigo,
		Nombre:                    s.Nombre,
		Direccion:                 s.Direccion,
		Comuna:                    s.Comuna,
		Region:                    s.Region,
		Telefono:                  s.Telefono,
		Email:                     s.Email,
		HorarioApertura:           s.HorarioApertura,
		HorarioCierre:             s.HorarioCierre,
		Timezone:                  s.Timezone,
		Habilitada:                s.Habilitada,
		MaxConexionesConcurrentes: s.MaxConexionesConcurrentes,
		MetricasRendimiento:       s.MetricasRendimiento,
		FechaCreacion:             s.FechaCreacion,
		FechaModificacion:         s.FechaModificacion,
	}
}

func (s *Sucursal) ToListDTO() SucursalListDTO {
	return SucursalListDTO{
		ID:                s.ID,
		Codigo:            s.Codigo,
		Nombre:            s.Nombre,
		Comuna:            s.Comuna,
		Region:            s.Region,
		Habilitada:        s.Habilitada,
		FechaCreacion:     s.FechaCreacion,
		FechaModificacion: s.FechaModificacion,
	}
}

func (dto *SucursalCreateDTO) ToModel() *Sucursal {
	sucursal := &Sucursal{
		Codigo:                    dto.Codigo,
		Nombre:                    dto.Nombre,
		Direccion:                 dto.Direccion,
		Comuna:                    dto.Comuna,
		Region:                    dto.Region,
		Telefono:                  dto.Telefono,
		Email:                     dto.Email,
		HorarioApertura:           dto.HorarioApertura,
		HorarioCierre:             dto.HorarioCierre,
		Timezone:                  dto.Timezone,
		Habilitada:                true,
		ConfiguracionDTE:          dto.ConfiguracionDTE,
		ConfiguracionPagos:        dto.ConfiguracionPagos,
		MaxConexionesConcurrentes: dto.MaxConexionesConcurrentes,
		ConfiguracionCache:        dto.ConfiguracionCache,
	}

	if sucursal.Timezone == "" {
		sucursal.Timezone = "America/Santiago"
	}
	if sucursal.MaxConexionesConcurrentes == 0 {
		sucursal.MaxConexionesConcurrentes = 50
	}

	return sucursal
}

// Validaciones personalizadas

func (s *Sucursal) Validate() error {
	if s.HorarioApertura != nil && s.HorarioCierre != nil {
		if s.HorarioApertura.After(*s.HorarioCierre) {
			return fmt.Errorf("horario de apertura no puede ser posterior al horario de cierre")
		}
	}

	if s.MaxConexionesConcurrentes <= 0 {
		return fmt.Errorf("máximo de conexiones concurrentes debe ser mayor a 0")
	}

	return nil
}
