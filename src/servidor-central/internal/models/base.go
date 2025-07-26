package models

import (
	"time"

	"github.com/google/uuid"
)

// BaseModel contiene campos comunes para todos los modelos
type BaseModel struct {
	ID                uuid.UUID `json:"id" db:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	FechaCreacion     time.Time `json:"fecha_creacion" db:"fecha_creacion" gorm:"default:CURRENT_TIMESTAMP"`
	FechaModificacion time.Time `json:"fecha_modificacion" db:"fecha_modificacion" gorm:"default:CURRENT_TIMESTAMP"`
}

// BaseModelWithUser extiende BaseModel con información de usuario
type BaseModelWithUser struct {
	BaseModel
	UsuarioCreacion     *uuid.UUID `json:"usuario_creacion,omitempty" db:"usuario_creacion"`
	UsuarioModificacion *uuid.UUID `json:"usuario_modificacion,omitempty" db:"usuario_modificacion"`
}

// Enums definidos en la base de datos

type RolUsuario string

const (
	RolCajero            RolUsuario = "cajero"
	RolVendedor          RolUsuario = "vendedor"
	RolDespacho          RolUsuario = "despacho"
	RolSupervisor        RolUsuario = "supervisor"
	RolAdmin             RolUsuario = "admin"
	RolOperadorEtiquetas RolUsuario = "operador_etiquetas"
)

type EstadoDocumento string

const (
	EstadoPendiente EstadoDocumento = "pendiente"
	EstadoProcesado EstadoDocumento = "procesado"
	EstadoEnviado   EstadoDocumento = "enviado"
	EstadoRechazado EstadoDocumento = "rechazado"
	EstadoAnulado   EstadoDocumento = "anulado"
)

type TipoMovimientoFidelizacion string

const (
	TipoAcumulacion TipoMovimientoFidelizacion = "acumulacion"
	TipoCanje       TipoMovimientoFidelizacion = "canje"
	TipoAjuste      TipoMovimientoFidelizacion = "ajuste"
	TipoExpiracion  TipoMovimientoFidelizacion = "expiracion"
)

type EstadoSincronizacion string

const (
	SincPendiente  EstadoSincronizacion = "pendiente"
	SincEnProceso  EstadoSincronizacion = "en_proceso"
	SincCompletado EstadoSincronizacion = "completado"
	SincError      EstadoSincronizacion = "error"
)

type PrioridadProceso string

const (
	PrioridadMaxima PrioridadProceso = "maxima" // api_pos
	PrioridadMedia  PrioridadProceso = "media"  // api_sync
	PrioridadBaja   PrioridadProceso = "baja"   // api_labels
	PrioridadMinima PrioridadProceso = "minima" // api_report
)

type EstadoTrabajoEtiqueta string

const (
	EtiquetaPendiente  EstadoTrabajoEtiqueta = "pendiente"
	EtiquetaProcesando EstadoTrabajoEtiqueta = "procesando"
	EtiquetaCompletado EstadoTrabajoEtiqueta = "completado"
	EtiquetaError      EstadoTrabajoEtiqueta = "error"
	EtiquetaCancelado  EstadoTrabajoEtiqueta = "cancelado"
)

// Estructuras de respuesta comunes

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type PaginatedResponse struct {
	APIResponse
	Pagination PaginationInfo `json:"pagination"`
}

type PaginationInfo struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// Estructuras de filtros comunes

type DateRangeFilter struct {
	FechaInicio *time.Time `json:"fecha_inicio" form:"fecha_inicio"`
	FechaFin    *time.Time `json:"fecha_fin" form:"fecha_fin"`
}

type PaginationFilter struct {
	Page  int `json:"page" form:"page" binding:"min=1" default:"1"`
	Limit int `json:"limit" form:"limit" binding:"min=1,max=100" default:"20"`
}

type SortFilter struct {
	SortBy    string `json:"sort_by" form:"sort_by"`
	SortOrder string `json:"sort_order" form:"sort_order" binding:"oneof=asc desc" default:"asc"`
}

// Métodos helper para BaseModel

func (b *BaseModel) BeforeCreate() {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	now := time.Now()
	b.FechaCreacion = now
	b.FechaModificacion = now
}

func (b *BaseModel) BeforeUpdate() {
	b.FechaModificacion = time.Now()
}

// Métodos helper para respuestas

func NewSuccessResponse(message string, data interface{}) APIResponse {
	return APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
}

func NewErrorResponse(message string, err error) APIResponse {
	response := APIResponse{
		Success: false,
		Message: message,
	}
	if err != nil {
		response.Error = err.Error()
	}
	return response
}

func NewPaginatedResponse(message string, data interface{}, pagination PaginationInfo) PaginatedResponse {
	return PaginatedResponse{
		APIResponse: NewSuccessResponse(message, data),
		Pagination:  pagination,
	}
}

// Calcular información de paginación
func CalculatePagination(page, limit int, total int64) PaginationInfo {
	totalPages := int((total + int64(limit) - 1) / int64(limit))
	return PaginationInfo{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}
}
