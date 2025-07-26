// Package models contiene todos los modelos de datos con notación húngara
package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// structBaseModel contiene campos comunes para todos los modelos
type structBaseModel struct {
	UuidID                uuid.UUID `json:"id" db:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()" validate:"required"`
	TimeFechaCreacion     time.Time `json:"fecha_creacion" db:"fecha_creacion" gorm:"default:CURRENT_TIMESTAMP"`
	TimeFechaModificacion time.Time `json:"fecha_modificacion" db:"fecha_modificacion" gorm:"default:CURRENT_TIMESTAMP"`
}

// structBaseModelWithUser extiende structBaseModel con información de usuario
type structBaseModelWithUser struct {
	structBaseModel
	PtrUuidUsuarioCreacion     *uuid.UUID `json:"usuario_creacion,omitempty" db:"usuario_creacion"`
	PtrUuidUsuarioModificacion *uuid.UUID `json:"usuario_modificacion,omitempty" db:"usuario_modificacion"`
}

// structSoftDelete agrega funcionalidad de soft delete
type structSoftDelete struct {
	BoolEliminado             bool       `json:"eliminado" db:"eliminado" gorm:"default:false"`
	PtrTimeFechaEliminacion   *time.Time `json:"fecha_eliminacion,omitempty" db:"fecha_eliminacion"`
	PtrUuidUsuarioEliminacion *uuid.UUID `json:"usuario_eliminacion,omitempty" db:"usuario_eliminacion"`
}

// structAuditoria proporciona campos de auditoría completos
type structAuditoria struct {
	structBaseModelWithUser
	structSoftDelete
	StrIPCreacion            string `json:"ip_creacion,omitempty" db:"ip_creacion" validate:"omitempty,ip"`
	StrIPModificacion        string `json:"ip_modificacion,omitempty" db:"ip_modificacion" validate:"omitempty,ip"`
	StrUserAgentCreacion     string `json:"user_agent_creacion,omitempty" db:"user_agent_creacion"`
	StrUserAgentModificacion string `json:"user_agent_modificacion,omitempty" db:"user_agent_modificacion"`
}

// Enums con notación húngara

// enumRolUsuario define los roles de usuario del sistema
type enumRolUsuario string

const (
	EnumRolUsuarioCajero            enumRolUsuario = "cajero"
	EnumRolUsuarioVendedor          enumRolUsuario = "vendedor"
	EnumRolUsuarioDespacho          enumRolUsuario = "despacho"
	EnumRolUsuarioSupervisor        enumRolUsuario = "supervisor"
	EnumRolUsuarioAdmin             enumRolUsuario = "admin"
	EnumRolUsuarioOperadorEtiquetas enumRolUsuario = "operador_etiquetas"
	EnumRolUsuarioGerente           enumRolUsuario = "gerente"
	EnumRolUsuarioContador          enumRolUsuario = "contador"
)

// enumEstadoDocumento define los estados de documentos
type enumEstadoDocumento string

const (
	EnumEstadoDocumentoPendiente enumEstadoDocumento = "pendiente"
	EnumEstadoDocumentoProcesado enumEstadoDocumento = "procesado"
	EnumEstadoDocumentoEnviado   enumEstadoDocumento = "enviado"
	EnumEstadoDocumentoRechazado enumEstadoDocumento = "rechazado"
	EnumEstadoDocumentoAnulado   enumEstadoDocumento = "anulado"
	EnumEstadoDocumentoCancelado enumEstadoDocumento = "cancelado"
	EnumEstadoDocumentoVencido   enumEstadoDocumento = "vencido"
)

// enumTipoMovimientoFidelizacion define tipos de movimientos de fidelización
type enumTipoMovimientoFidelizacion string

const (
	EnumTipoMovimientoFidelizacionAcumulacion enumTipoMovimientoFidelizacion = "acumulacion"
	EnumTipoMovimientoFidelizacionCanje       enumTipoMovimientoFidelizacion = "canje"
	EnumTipoMovimientoFidelizacionAjuste      enumTipoMovimientoFidelizacion = "ajuste"
	EnumTipoMovimientoFidelizacionVencimiento enumTipoMovimientoFidelizacion = "vencimiento"
	EnumTipoMovimientoFidelizacionBono        enumTipoMovimientoFidelizacion = "bono"
)

// enumTipoDocumento define tipos de documentos
type enumTipoDocumento string

const (
	EnumTipoDocumentoBoleta       enumTipoDocumento = "boleta"
	EnumTipoDocumentoFactura      enumTipoDocumento = "factura"
	EnumTipoDocumentoNotaCredito  enumTipoDocumento = "nota_credito"
	EnumTipoDocumentoNotaDebito   enumTipoDocumento = "nota_debito"
	EnumTipoDocumentoGuiaDespacho enumTipoDocumento = "guia_despacho"
	EnumTipoDocumentoOrdenCompra  enumTipoDocumento = "orden_compra"
	EnumTipoDocumentoCotizacion   enumTipoDocumento = "cotizacion"
)

// enumMedioPago define medios de pago
type enumMedioPago string

const (
	EnumMedioPagoEfectivo      enumMedioPago = "efectivo"
	EnumMedioPagoDebito        enumMedioPago = "debito"
	EnumMedioPagoCredito       enumMedioPago = "credito"
	EnumMedioPagoTransferencia enumMedioPago = "transferencia"
	EnumMedioPagoCheque        enumMedioPago = "cheque"
	EnumMedioPagoVale          enumMedioPago = "vale"
	EnumMedioPagoCripto        enumMedioPago = "cripto"
	EnumMedioPagoCredito30     enumMedioPago = "credito_30"
	EnumMedioPagoCredito60     enumMedioPago = "credito_60"
	EnumMedioPagoCredito90     enumMedioPago = "credito_90"
)

// enumEstadoSincronizacion define estados de sincronización
type enumEstadoSincronizacion string

const (
	EnumEstadoSincronizacionPendiente  enumEstadoSincronizacion = "pendiente"
	EnumEstadoSincronizacionProcesando enumEstadoSincronizacion = "procesando"
	EnumEstadoSincronizacionCompletado enumEstadoSincronizacion = "completado"
	EnumEstadoSincronizacionError      enumEstadoSincronizacion = "error"
	EnumEstadoSincronizacionCancelado  enumEstadoSincronizacion = "cancelado"
	EnumEstadoSincronizacionConflicto  enumEstadoSincronizacion = "conflicto"
)

// enumPrioridad define niveles de prioridad
type enumPrioridad string

const (
	EnumPrioridadBaja    enumPrioridad = "baja"
	EnumPrioridadNormal  enumPrioridad = "normal"
	EnumPrioridadAlta    enumPrioridad = "alta"
	EnumPrioridadCritica enumPrioridad = "critica"
	EnumPrioridadUrgente enumPrioridad = "urgente"
)

// structPaginacion contiene parámetros de paginación
type structPaginacion struct {
	IntPagina       int `json:"pagina" validate:"min=1" form:"pagina"`
	IntTamano       int `json:"tamano" validate:"min=1,max=1000" form:"tamano"`
	IntOffset       int `json:"offset" validate:"min=0"`
	IntTotal        int `json:"total" validate:"min=0"`
	IntTotalPaginas int `json:"total_paginas" validate:"min=0"`
}

// structFiltros contiene filtros comunes para consultas
type structFiltros struct {
	StrBusqueda       string      `json:"busqueda,omitempty" form:"busqueda"`
	PtrTimeFechaDesde *time.Time  `json:"fecha_desde,omitempty" form:"fecha_desde"`
	PtrTimeFechaHasta *time.Time  `json:"fecha_hasta,omitempty" form:"fecha_hasta"`
	BoolActivo        *bool       `json:"activo,omitempty" form:"activo"`
	BoolEliminado     *bool       `json:"eliminado,omitempty" form:"eliminado"`
	ArrUuidSucursales []uuid.UUID `json:"sucursales,omitempty" form:"sucursales"`
	ArrStrCamposOrden []string    `json:"campos_orden,omitempty" form:"campos_orden"`
	StrDireccionOrden string      `json:"direccion_orden,omitempty" form:"direccion_orden" validate:"omitempty,oneof=asc desc"`
}

// structRespuestaAPI estructura estándar para respuestas de API
type structRespuestaAPI struct {
	BoolExito     bool              `json:"exito"`
	StrMensaje    string            `json:"mensaje,omitempty"`
	ObjDatos      interface{}       `json:"datos,omitempty"`
	PtrPaginacion *structPaginacion `json:"paginacion,omitempty"`
	ArrErrores    []structErrorAPI  `json:"errores,omitempty"`
	StrRequestID  string            `json:"request_id,omitempty"`
	TimeTimestamp time.Time         `json:"timestamp"`
}

// structErrorAPI estructura para errores en respuestas de API
type structErrorAPI struct {
	StrCodigo   string      `json:"codigo"`
	StrMensaje  string      `json:"mensaje"`
	StrCampo    string      `json:"campo,omitempty"`
	ObjDetalles interface{} `json:"detalles,omitempty"`
}

// structRespuestaPaginada estructura para respuestas paginadas
type structRespuestaPaginada struct {
	ArrDatos      []interface{}     `json:"datos"`
	PtrPaginacion *structPaginacion `json:"paginacion"`
	StrRequestID  string            `json:"request_id,omitempty"`
	TimeTimestamp time.Time         `json:"timestamp"`
}

// structMetadatos contiene metadatos adicionales
type structMetadatos struct {
	MapDatos    map[string]interface{} `json:"datos,omitempty" db:"datos" gorm:"type:jsonb"`
	StrVersion  string                 `json:"version,omitempty" db:"version"`
	StrChecksum string                 `json:"checksum,omitempty" db:"checksum"`
	IntTamano   int64                  `json:"tamano,omitempty" db:"tamano"`
	StrMimeType string                 `json:"mime_type,omitempty" db:"mime_type"`
	StrEncoding string                 `json:"encoding,omitempty" db:"encoding"`
}

// structConfiguracion estructura base para configuraciones
type structConfiguracion struct {
	structBaseModelWithUser
	StrClave       string                 `json:"clave" db:"clave" validate:"required,min=1,max=255"`
	ObjValor       interface{}            `json:"valor" db:"valor" gorm:"type:jsonb"`
	StrDescripcion string                 `json:"descripcion,omitempty" db:"descripcion"`
	StrCategoria   string                 `json:"categoria,omitempty" db:"categoria"`
	BoolPublica    bool                   `json:"publica" db:"publica" gorm:"default:false"`
	BoolEditable   bool                   `json:"editable" db:"editable" gorm:"default:true"`
	MapValidacion  map[string]interface{} `json:"validacion,omitempty" db:"validacion" gorm:"type:jsonb"`
}

// structArchivo estructura para manejo de archivos
type structArchivo struct {
	structBaseModelWithUser
	StrNombre          string                 `json:"nombre" db:"nombre" validate:"required,min=1,max=255"`
	StrRuta            string                 `json:"ruta" db:"ruta" validate:"required"`
	StrRutaPublica     string                 `json:"ruta_publica,omitempty" db:"ruta_publica"`
	IntTamano          int64                  `json:"tamano" db:"tamano" validate:"min=0"`
	StrMimeType        string                 `json:"mime_type" db:"mime_type"`
	StrChecksum        string                 `json:"checksum,omitempty" db:"checksum"`
	MapMetadatos       map[string]interface{} `json:"metadatos,omitempty" db:"metadatos" gorm:"type:jsonb"`
	BoolPublico        bool                   `json:"publico" db:"publico" gorm:"default:false"`
	PtrTimeVencimiento *time.Time             `json:"vencimiento,omitempty" db:"vencimiento"`
}

// structNotificacion estructura para notificaciones
type structNotificacion struct {
	structBaseModelWithUser
	UuidDestinatario  uuid.UUID              `json:"destinatario" db:"destinatario" validate:"required"`
	StrTitulo         string                 `json:"titulo" db:"titulo" validate:"required,min=1,max=255"`
	StrMensaje        string                 `json:"mensaje" db:"mensaje" validate:"required,min=1"`
	EnumTipo          enumTipoNotificacion   `json:"tipo" db:"tipo" validate:"required"`
	EnumPrioridad     enumPrioridad          `json:"prioridad" db:"prioridad" gorm:"default:'normal'"`
	BoolLeida         bool                   `json:"leida" db:"leida" gorm:"default:false"`
	PtrTimeFechaLeida *time.Time             `json:"fecha_leida,omitempty" db:"fecha_leida"`
	MapDatos          map[string]interface{} `json:"datos,omitempty" db:"datos" gorm:"type:jsonb"`
	StrURL            string                 `json:"url,omitempty" db:"url"`
	StrIcono          string                 `json:"icono,omitempty" db:"icono"`
}

// enumTipoNotificacion define tipos de notificaciones
type enumTipoNotificacion string

const (
	EnumTipoNotificacionInfo    enumTipoNotificacion = "info"
	EnumTipoNotificacionWarning enumTipoNotificacion = "warning"
	EnumTipoNotificacionError   enumTipoNotificacion = "error"
	EnumTipoNotificacionSuccess enumTipoNotificacion = "success"
	EnumTipoNotificacionVenta   enumTipoNotificacion = "venta"
	EnumTipoNotificacionStock   enumTipoNotificacion = "stock"
	EnumTipoNotificacionSistema enumTipoNotificacion = "sistema"
)

// structLog estructura para logs del sistema
type structLog struct {
	structBaseModel
	UuidUsuario   *uuid.UUID             `json:"usuario,omitempty" db:"usuario"`
	UuidSucursal  *uuid.UUID             `json:"sucursal,omitempty" db:"sucursal"`
	StrNivel      string                 `json:"nivel" db:"nivel" validate:"required,oneof=debug info warn error fatal"`
	StrMensaje    string                 `json:"mensaje" db:"mensaje" validate:"required"`
	StrComponente string                 `json:"componente,omitempty" db:"componente"`
	StrAccion     string                 `json:"accion,omitempty" db:"accion"`
	StrRecurso    string                 `json:"recurso,omitempty" db:"recurso"`
	UuidRecursoID *uuid.UUID             `json:"recurso_id,omitempty" db:"recurso_id"`
	StrIP         string                 `json:"ip,omitempty" db:"ip" validate:"omitempty,ip"`
	StrUserAgent  string                 `json:"user_agent,omitempty" db:"user_agent"`
	StrRequestID  string                 `json:"request_id,omitempty" db:"request_id"`
	MapContexto   map[string]interface{} `json:"contexto,omitempty" db:"contexto" gorm:"type:jsonb"`
	StrStackTrace string                 `json:"stack_trace,omitempty" db:"stack_trace"`
}

// Métodos helper para structBaseModel

// GetID retorna el ID del modelo
func (ptrModel *structBaseModel) GetID() uuid.UUID {
	return ptrModel.UuidID
}

// SetID establece el ID del modelo
func (ptrModel *structBaseModel) SetID(uuidID uuid.UUID) {
	ptrModel.UuidID = uuidID
}

// GetFechaCreacion retorna la fecha de creación
func (ptrModel *structBaseModel) GetFechaCreacion() time.Time {
	return ptrModel.TimeFechaCreacion
}

// GetFechaModificacion retorna la fecha de modificación
func (ptrModel *structBaseModel) GetFechaModificacion() time.Time {
	return ptrModel.TimeFechaModificacion
}

// ActualizarFechaModificacion actualiza la fecha de modificación
func (ptrModel *structBaseModel) ActualizarFechaModificacion() {
	ptrModel.TimeFechaModificacion = time.Now()
}

// Métodos helper para structPaginacion

// CalcularOffset calcula el offset basado en página y tamaño
func (ptrPaginacion *structPaginacion) CalcularOffset() {
	if ptrPaginacion.IntPagina <= 0 {
		ptrPaginacion.IntPagina = 1
	}
	if ptrPaginacion.IntTamano <= 0 {
		ptrPaginacion.IntTamano = 20
	}
	ptrPaginacion.IntOffset = (ptrPaginacion.IntPagina - 1) * ptrPaginacion.IntTamano
}

// CalcularTotalPaginas calcula el total de páginas
func (ptrPaginacion *structPaginacion) CalcularTotalPaginas() {
	if ptrPaginacion.IntTamano > 0 {
		ptrPaginacion.IntTotalPaginas = (ptrPaginacion.IntTotal + ptrPaginacion.IntTamano - 1) / ptrPaginacion.IntTamano
	}
}

// TieneSiguientePagina verifica si hay siguiente página
func (ptrPaginacion *structPaginacion) TieneSiguientePagina() bool {
	return ptrPaginacion.IntPagina < ptrPaginacion.IntTotalPaginas
}

// TienePaginaAnterior verifica si hay página anterior
func (ptrPaginacion *structPaginacion) TienePaginaAnterior() bool {
	return ptrPaginacion.IntPagina > 1
}

// Métodos helper para structRespuestaAPI

// NewRespuestaExito crea una respuesta exitosa
func NewRespuestaExito(objDatos interface{}, strMensaje string) *structRespuestaAPI {
	return &structRespuestaAPI{
		BoolExito:     true,
		StrMensaje:    strMensaje,
		ObjDatos:      objDatos,
		TimeTimestamp: time.Now(),
	}
}

// NewRespuestaError crea una respuesta de error
func NewRespuestaError(strMensaje string, arrErrores []structErrorAPI) *structRespuestaAPI {
	return &structRespuestaAPI{
		BoolExito:     false,
		StrMensaje:    strMensaje,
		ArrErrores:    arrErrores,
		TimeTimestamp: time.Now(),
	}
}

// NewRespuestaPaginada crea una respuesta paginada
func NewRespuestaPaginada(arrDatos []interface{}, ptrPaginacion *structPaginacion) *structRespuestaPaginada {
	return &structRespuestaPaginada{
		ArrDatos:      arrDatos,
		PtrPaginacion: ptrPaginacion,
		TimeTimestamp: time.Now(),
	}
}

// AgregarError agrega un error a la respuesta
func (ptrRespuesta *structRespuestaAPI) AgregarError(strCodigo, strMensaje, strCampo string) {
	if ptrRespuesta.ArrErrores == nil {
		ptrRespuesta.ArrErrores = make([]structErrorAPI, 0)
	}

	ptrRespuesta.ArrErrores = append(ptrRespuesta.ArrErrores, structErrorAPI{
		StrCodigo:  strCodigo,
		StrMensaje: strMensaje,
		StrCampo:   strCampo,
	})

	ptrRespuesta.BoolExito = false
}

// SetRequestID establece el ID de la request
func (ptrRespuesta *structRespuestaAPI) SetRequestID(strRequestID string) {
	ptrRespuesta.StrRequestID = strRequestID
}

// Funciones de utilidad

// GenerarUUID genera un nuevo UUID
func GenerarUUID() uuid.UUID {
	return uuid.New()
}

// ParsearUUID parsea un string a UUID
func ParsearUUID(strUUID string) (uuid.UUID, error) {
	return uuid.Parse(strUUID)
}

// ValidarUUID valida si un string es un UUID válido
func ValidarUUID(strUUID string) bool {
	_, err := uuid.Parse(strUUID)
	return err == nil
}

// FormatearFecha formatea una fecha a string
func FormatearFecha(timeFecha time.Time, strFormato string) string {
	if strFormato == "" {
		strFormato = "2006-01-02 15:04:05"
	}
	return timeFecha.Format(strFormato)
}

// ParsearFecha parsea un string a fecha
func ParsearFecha(strFecha, strFormato string) (time.Time, error) {
	if strFormato == "" {
		strFormato = "2006-01-02 15:04:05"
	}
	return time.Parse(strFormato, strFecha)
}

// ConvertirAJSON convierte un objeto a JSON
func ConvertirAJSON(objDatos interface{}) (string, error) {
	arrBytes, err := json.Marshal(objDatos)
	if err != nil {
		return "", err
	}
	return string(arrBytes), nil
}

// ParsearDesdeJSON parsea JSON a un objeto
func ParsearDesdeJSON(strJSON string, objDestino interface{}) error {
	return json.Unmarshal([]byte(strJSON), objDestino)
}

// LogFields convierte un modelo a campos de log
func (ptrModel *structBaseModel) LogFields() []zap.Field {
	return []zap.Field{
		zap.String("id", ptrModel.UuidID.String()),
		zap.Time("fecha_creacion", ptrModel.TimeFechaCreacion),
		zap.Time("fecha_modificacion", ptrModel.TimeFechaModificacion),
	}
}

// LogFields para modelo con usuario
func (ptrModel *structBaseModelWithUser) LogFields() []zap.Field {
	arrFields := ptrModel.structBaseModel.LogFields()

	if ptrModel.PtrUuidUsuarioCreacion != nil {
		arrFields = append(arrFields, zap.String("usuario_creacion", ptrModel.PtrUuidUsuarioCreacion.String()))
	}

	if ptrModel.PtrUuidUsuarioModificacion != nil {
		arrFields = append(arrFields, zap.String("usuario_modificacion", ptrModel.PtrUuidUsuarioModificacion.String()))
	}

	return arrFields
}

// Validaciones de enums

// IsValidRolUsuario valida si un rol de usuario es válido
func IsValidRolUsuario(enumRol enumRolUsuario) bool {
	switch enumRol {
	case EnumRolUsuarioCajero, EnumRolUsuarioVendedor, EnumRolUsuarioDespacho,
		EnumRolUsuarioSupervisor, EnumRolUsuarioAdmin, EnumRolUsuarioOperadorEtiquetas,
		EnumRolUsuarioGerente, EnumRolUsuarioContador:
		return true
	default:
		return false
	}
}

// IsValidEstadoDocumento valida si un estado de documento es válido
func IsValidEstadoDocumento(enumEstado enumEstadoDocumento) bool {
	switch enumEstado {
	case EnumEstadoDocumentoPendiente, EnumEstadoDocumentoProcesado, EnumEstadoDocumentoEnviado,
		EnumEstadoDocumentoRechazado, EnumEstadoDocumentoAnulado, EnumEstadoDocumentoCancelado,
		EnumEstadoDocumentoVencido:
		return true
	default:
		return false
	}
}

// IsValidMedioPago valida si un medio de pago es válido
func IsValidMedioPago(enumMedio enumMedioPago) bool {
	switch enumMedio {
	case EnumMedioPagoEfectivo, EnumMedioPagoDebito, EnumMedioPagoCredito,
		EnumMedioPagoTransferencia, EnumMedioPagoCheque, EnumMedioPagoVale,
		EnumMedioPagoCripto, EnumMedioPagoCredito30, EnumMedioPagoCredito60,
		EnumMedioPagoCredito90:
		return true
	default:
		return false
	}
}

// Estructuras de respuesta API

// structAuthResponse respuesta de autenticación
type structAuthResponse struct {
	StrAccessToken  string    `json:"access_token"`
	StrRefreshToken string    `json:"refresh_token"`
	StrTokenType    string    `json:"token_type"`
	IntExpiresIn    int       `json:"expires_in"`
	TimeExpiresAt   time.Time `json:"expires_at"`
	StructUsuario   interface{} `json:"usuario"`
}

// structPaginationResponse respuesta paginada genérica
type structPaginationResponse struct {
	ArrData      interface{} `json:"data"`
	IntTotal     int64       `json:"total"`
	IntPage      int         `json:"page"`
	IntPageSize  int         `json:"page_size"`
	IntTotalPages int        `json:"total_pages"`
	BoolHasNext  bool        `json:"has_next"`
	BoolHasPrev  bool        `json:"has_prev"`
}

// Aliases para compatibilidad
type AuthResponse = structAuthResponse
type PaginationResponse = structPaginationResponse


// DTOs específicos para servicios

// structProductoPOSResponseDTO respuesta de producto para POS
type structProductoPOSResponseDTO struct {
	UuidID              uuid.UUID `json:"id"`
	StrNombre           string    `json:"nombre"`
	StrCodigoBarras     string    `json:"codigo_barras"`
	FltPrecio           float64   `json:"precio"`
	IntStock            int       `json:"stock"`
	BoolActivo          bool      `json:"activo"`
	StrCategoria        string    `json:"categoria"`
	StrDescripcion      string    `json:"descripcion,omitempty"`
	FltPrecioOferta     *float64  `json:"precio_oferta,omitempty"`
	BoolTieneOferta     bool      `json:"tiene_oferta"`
}

// structProductoEtiquetaDTO DTO para etiquetas de productos
type structProductoEtiquetaDTO struct {
	UuidID              uuid.UUID `json:"id"`
	StrNombre           string    `json:"nombre"`
	StrCodigoBarras     string    `json:"codigo_barras"`
	FltPrecio           float64   `json:"precio"`
	StrCategoria        string    `json:"categoria"`
	StrMarca            string    `json:"marca,omitempty"`
	StrUnidadMedida     string    `json:"unidad_medida"`
	MapDatosAdicionales map[string]interface{} `json:"datos_adicionales,omitempty"`
}

// structStockResponseDTO respuesta de stock
type structStockResponseDTO struct {
	UuidProductoID      uuid.UUID `json:"producto_id"`
	UuidSucursalID      uuid.UUID `json:"sucursal_id"`
	IntCantidad         int       `json:"cantidad"`
	IntCantidadReservada int      `json:"cantidad_reservada"`
	IntCantidadDisponible int     `json:"cantidad_disponible"`
	IntStockMinimo      int       `json:"stock_minimo"`
	IntStockMaximo      int       `json:"stock_maximo"`
	StrEstadoStock      string    `json:"estado_stock"`
	TimeFechaActualizacion time.Time `json:"fecha_actualizacion"`
}

// structReservaStockDTO DTO para reserva de stock
type structReservaStockDTO struct {
	UuidProductoID      uuid.UUID `json:"producto_id"`
	UuidSucursalID      uuid.UUID `json:"sucursal_id"`
	IntCantidad         int       `json:"cantidad"`
	StrMotivo           string    `json:"motivo"`
	StrDocumentoReferencia *string `json:"documento_referencia,omitempty"`
	IntTiempoExpiracion *int      `json:"tiempo_expiracion_minutos,omitempty"`
}

// structReservaStockResponseDTO respuesta de reserva de stock
type structReservaStockResponseDTO struct {
	UuidReservaID       uuid.UUID `json:"reserva_id"`
	UuidProductoID      uuid.UUID `json:"producto_id"`
	UuidSucursalID      uuid.UUID `json:"sucursal_id"`
	IntCantidad         int       `json:"cantidad"`
	StrEstado           string    `json:"estado"`
	TimeFechaCreacion   time.Time `json:"fecha_creacion"`
	TimeFechaExpiracion *time.Time `json:"fecha_expiracion,omitempty"`
}

// structVentaCreateDTO DTO para crear venta
type structVentaCreateDTO struct {
	UuidSucursalID      uuid.UUID `json:"sucursal_id"`
	UuidClienteID       *uuid.UUID `json:"cliente_id,omitempty"`
	ArrDetalles         []structVentaDetalleDTO `json:"detalles"`
	ArrMediosPago       []structMedioPagoDTO `json:"medios_pago"`
	FltDescuento        *float64  `json:"descuento,omitempty"`
	StrObservaciones    *string   `json:"observaciones,omitempty"`
}

// structVentaDetalleDTO DTO para detalle de venta
type structVentaDetalleDTO struct {
	UuidProductoID      uuid.UUID `json:"producto_id"`
	IntCantidad         int       `json:"cantidad"`
	FltPrecioUnitario   float64   `json:"precio_unitario"`
	FltDescuento        *float64  `json:"descuento,omitempty"`
}

// structMedioPagoDTO DTO para medio de pago
type structMedioPagoDTO struct {
	StrTipo             string    `json:"tipo"`
	FltMonto            float64   `json:"monto"`
	StrReferencia       *string   `json:"referencia,omitempty"`
	MapDatosAdicionales map[string]interface{} `json:"datos_adicionales,omitempty"`
}

// structVentaResponseDTO respuesta de venta
type structVentaResponseDTO struct {
	UuidID              uuid.UUID `json:"id"`
	StrNumeroVenta      string    `json:"numero_venta"`
	UuidSucursalID      uuid.UUID `json:"sucursal_id"`
	UuidClienteID       *uuid.UUID `json:"cliente_id,omitempty"`
	FltSubtotal         float64   `json:"subtotal"`
	FltDescuento        float64   `json:"descuento"`
	FltImpuestos        float64   `json:"impuestos"`
	FltTotal            float64   `json:"total"`
	StrEstado           string    `json:"estado"`
	TimeFechaVenta      time.Time `json:"fecha_venta"`
	ArrDetalles         []structVentaDetalleResponseDTO `json:"detalles"`
	ArrMediosPago       []structMedioPagoResponseDTO `json:"medios_pago"`
}

// structVentaDetalleResponseDTO respuesta de detalle de venta
type structVentaDetalleResponseDTO struct {
	UuidID              uuid.UUID `json:"id"`
	UuidProductoID      uuid.UUID `json:"producto_id"`
	StrProductoNombre   string    `json:"producto_nombre"`
	IntCantidad         int       `json:"cantidad"`
	FltPrecioUnitario   float64   `json:"precio_unitario"`
	FltDescuento        float64   `json:"descuento"`
	FltSubtotal         float64   `json:"subtotal"`
}

// structMedioPagoResponseDTO respuesta de medio de pago
type structMedioPagoResponseDTO struct {
	UuidID              uuid.UUID `json:"id"`
	StrTipo             string    `json:"tipo"`
	FltMonto            float64   `json:"monto"`
	StrReferencia       *string   `json:"referencia,omitempty"`
	StrEstado           string    `json:"estado"`
}

// Aliases para compatibilidad con código existente
type ProductoPOSResponseDTO = structProductoPOSResponseDTO
type ProductoEtiquetaDTO = structProductoEtiquetaDTO
type StockResponseDTO = structStockResponseDTO
type ReservaStockDTO = structReservaStockDTO
type ReservaStockResponseDTO = structReservaStockResponseDTO
type VentaCreateDTO = structVentaCreateDTO
type VentaDetalleDTO = structVentaDetalleDTO
type MedioPagoDTO = structMedioPagoDTO
type VentaResponseDTO = structVentaResponseDTO
type VentaDetalleResponseDTO = structVentaDetalleResponseDTO
type MedioPagoResponseDTO = structMedioPagoResponseDTO


// DTOs adicionales para completar interfaces de servicios

// structActualizarStockDTO DTO para actualizar stock
type structActualizarStockDTO struct {
	UuidProductoID      uuid.UUID `json:"producto_id" binding:"required"`
	UuidSucursalID      uuid.UUID `json:"sucursal_id" binding:"required"`
	IntCantidad         int       `json:"cantidad" binding:"required"`
	StrTipoMovimiento   string    `json:"tipo_movimiento" binding:"required"`
	StrMotivo           string    `json:"motivo" binding:"required"`
	StrDocumentoReferencia *string `json:"documento_referencia,omitempty"`
	StrObservaciones    *string   `json:"observaciones,omitempty"`
}

// structVentaListDTO DTO para listar ventas
type structVentaListDTO struct {
	UuidID              uuid.UUID `json:"id"`
	StrNumeroVenta      string    `json:"numero_venta"`
	UuidSucursalID      uuid.UUID `json:"sucursal_id"`
	StrSucursalNombre   string    `json:"sucursal_nombre"`
	FltTotal            float64   `json:"total"`
	StrEstado           string    `json:"estado"`
	TimeFechaVenta      time.Time `json:"fecha_venta"`
	StrClienteNombre    *string   `json:"cliente_nombre,omitempty"`
	StrVendedorNombre   *string   `json:"vendedor_nombre,omitempty"`
}

// structAnulacionVentaDTO DTO para anular venta
type structAnulacionVentaDTO struct {
	UuidVentaID         uuid.UUID `json:"venta_id" binding:"required"`
	StrMotivo           string    `json:"motivo" binding:"required"`
	StrObservaciones    *string   `json:"observaciones,omitempty"`
	BoolReintegrarStock bool      `json:"reintegrar_stock"`
}

// structEstadisticasVentaDTO DTO para estadísticas de venta
type structEstadisticasVentaDTO struct {
	FltTotalVentas      float64   `json:"total_ventas"`
	IntCantidadVentas   int       `json:"cantidad_ventas"`
	FltPromedioVenta    float64   `json:"promedio_venta"`
	FltVentaMaxima      float64   `json:"venta_maxima"`
	FltVentaMinima      float64   `json:"venta_minima"`
	TimePeriodoInicio   time.Time `json:"periodo_inicio"`
	TimePeriodoFin      time.Time `json:"periodo_fin"`
	ArrVentasPorHora    []float64 `json:"ventas_por_hora"`
	ArrVentasPorDia     []float64 `json:"ventas_por_dia"`
}

// structSincronizacionDTO DTO para sincronización
type structSincronizacionDTO struct {
	UuidSucursalID      uuid.UUID `json:"sucursal_id" binding:"required"`
	StrTipoEntidad      string    `json:"tipo_entidad" binding:"required"`
	ArrEntidades        []interface{} `json:"entidades" binding:"required"`
	BoolForzarActualizacion bool  `json:"forzar_actualizacion"`
	StrVersionCliente   *string   `json:"version_cliente,omitempty"`
}

// structSincronizacionResultDTO DTO resultado de sincronización
type structSincronizacionResultDTO struct {
	UuidSincronizacionID uuid.UUID `json:"sincronizacion_id"`
	StrEstado           string    `json:"estado"`
	IntEntidadesProcesadas int    `json:"entidades_procesadas"`
	IntEntidadesExitosas int      `json:"entidades_exitosas"`
	IntEntidadesError   int       `json:"entidades_error"`
	ArrErrores          []string  `json:"errores,omitempty"`
	TimeFechaProceso    time.Time `json:"fecha_proceso"`
	IntTiempoProceso    int       `json:"tiempo_proceso_ms"`
}

// structClienteResponseDTO DTO respuesta de cliente
type structClienteResponseDTO struct {
	UuidID              uuid.UUID `json:"id"`
	StrRut              string    `json:"rut"`
	StrNombre           string    `json:"nombre"`
	StrEmail            *string   `json:"email,omitempty"`
	StrTelefono         *string   `json:"telefono,omitempty"`
	StrDireccion        *string   `json:"direccion,omitempty"`
	BoolActivo          bool      `json:"activo"`
	IntPuntosFidelidad  int       `json:"puntos_fidelidad"`
	TimeFechaRegistro   time.Time `json:"fecha_registro"`
	TimeFechaUltimaCompra *time.Time `json:"fecha_ultima_compra,omitempty"`
}

// structClienteListDTO DTO para listar clientes
type structClienteListDTO struct {
	UuidID              uuid.UUID `json:"id"`
	StrRut              string    `json:"rut"`
	StrNombre           string    `json:"nombre"`
	StrEmail            *string   `json:"email,omitempty"`
	StrTelefono         *string   `json:"telefono,omitempty"`
	BoolActivo          bool      `json:"activo"`
	IntPuntosFidelidad  int       `json:"puntos_fidelidad"`
	FltTotalCompras     float64   `json:"total_compras"`
	IntCantidadCompras  int       `json:"cantidad_compras"`
}

// structClienteCreateDTO DTO para crear cliente
type structClienteCreateDTO struct {
	StrRut              string    `json:"rut" binding:"required"`
	StrNombre           string    `json:"nombre" binding:"required"`
	StrEmail            *string   `json:"email,omitempty" binding:"omitempty,email"`
	StrTelefono         *string   `json:"telefono,omitempty"`
	StrDireccion        *string   `json:"direccion,omitempty"`
	BoolAceptaMarketing bool      `json:"acepta_marketing"`
	MapDatosAdicionales map[string]interface{} `json:"datos_adicionales,omitempty"`
}

// structClienteUpdateDTO DTO para actualizar cliente
type structClienteUpdateDTO struct {
	StrNombre           *string   `json:"nombre,omitempty"`
	StrEmail            *string   `json:"email,omitempty" binding:"omitempty,email"`
	StrTelefono         *string   `json:"telefono,omitempty"`
	StrDireccion        *string   `json:"direccion,omitempty"`
	BoolActivo          *bool     `json:"activo,omitempty"`
	BoolAceptaMarketing *bool     `json:"acepta_marketing,omitempty"`
	MapDatosAdicionales map[string]interface{} `json:"datos_adicionales,omitempty"`
}

// structEtiquetaCreateDTO DTO para crear etiqueta
type structEtiquetaCreateDTO struct {
	UuidPlantillaID     uuid.UUID `json:"plantilla_id" binding:"required"`
	ArrProductosID      []uuid.UUID `json:"productos_id" binding:"required,min=1"`
	UuidSucursalID      uuid.UUID `json:"sucursal_id" binding:"required"`
	IntCantidadCopias   int       `json:"cantidad_copias" binding:"min=1,max=1000"`
	StrFormatoSalida    string    `json:"formato_salida" binding:"required,oneof=pdf png jpg"`
	MapParametros       map[string]interface{} `json:"parametros,omitempty"`
	BoolGenerarPreview  bool      `json:"generar_preview"`
}

// structEtiquetaResponseDTO DTO respuesta de etiqueta
type structEtiquetaResponseDTO struct {
	UuidID              uuid.UUID `json:"id"`
	UuidPlantillaID     uuid.UUID `json:"plantilla_id"`
	StrPlantillaNombre  string    `json:"plantilla_nombre"`
	UuidProductoID      uuid.UUID `json:"producto_id"`
	StrProductoNombre   string    `json:"producto_nombre"`
	StrEstado           string    `json:"estado"`
	StrRutaArchivo      *string   `json:"ruta_archivo,omitempty"`
	StrRutaPreview      *string   `json:"ruta_preview,omitempty"`
	TimeFechaGeneracion time.Time `json:"fecha_generacion"`
	IntTiempoGeneracion *int      `json:"tiempo_generacion_ms,omitempty"`
	StrMensajeError     *string   `json:"mensaje_error,omitempty"`
}

// structReporteCreateDTO DTO para crear reporte
type structReporteCreateDTO struct {
	UuidPlantillaID     uuid.UUID `json:"plantilla_id" binding:"required"`
	UuidSucursalID      *uuid.UUID `json:"sucursal_id,omitempty"`
	MapParametros       map[string]interface{} `json:"parametros" binding:"required"`
	StrFormatoSalida    string    `json:"formato_salida" binding:"required,oneof=pdf excel csv"`
	BoolEnviarEmail     bool      `json:"enviar_email"`
	StrEmailDestino     *string   `json:"email_destino,omitempty" binding:"omitempty,email"`
	BoolProgramar       bool      `json:"programar"`
	TimeFechaProgramada *time.Time `json:"fecha_programada,omitempty"`
}

// structReporteResponseDTO DTO respuesta de reporte
type structReporteResponseDTO struct {
	UuidID              uuid.UUID `json:"id"`
	UuidPlantillaID     uuid.UUID `json:"plantilla_id"`
	StrPlantillaNombre  string    `json:"plantilla_nombre"`
	StrEstado           string    `json:"estado"`
	StrRutaArchivo      *string   `json:"ruta_archivo,omitempty"`
	FltProgreso         float64   `json:"progreso"`
	TimeFechaGeneracion time.Time `json:"fecha_generacion"`
	TimeFechaCompletado *time.Time `json:"fecha_completado,omitempty"`
	IntTiempoGeneracion *int      `json:"tiempo_generacion_ms,omitempty"`
	StrMensajeError     *string   `json:"mensaje_error,omitempty"`
	IntTamanoArchivo    *int64    `json:"tamano_archivo_bytes,omitempty"`
}

// Aliases para compatibilidad con interfaces existentes
type ActualizarStockDTO = structActualizarStockDTO
type VentaListDTO = structVentaListDTO
type AnulacionVentaDTO = structAnulacionVentaDTO
type EstadisticasVentaDTO = structEstadisticasVentaDTO
type SincronizacionDTO = structSincronizacionDTO
type SincronizacionResultDTO = structSincronizacionResultDTO
type ClienteResponseDTO = structClienteResponseDTO
type ClienteListDTO = structClienteListDTO
type ClienteCreateDTO = structClienteCreateDTO
type ClienteUpdateDTO = structClienteUpdateDTO
type EtiquetaCreateDTO = structEtiquetaCreateDTO
type EtiquetaResponseDTO = structEtiquetaResponseDTO
type ReporteCreateDTO = structReporteCreateDTO
type ReporteResponseDTO = structReporteResponseDTO


// DTOs finales para completar todas las interfaces

// structPuntosClienteResponseDTO DTO respuesta de puntos de cliente
type structPuntosClienteResponseDTO struct {
	UuidClienteID       uuid.UUID `json:"cliente_id"`
	StrClienteNombre    string    `json:"cliente_nombre"`
	IntPuntosActuales   int       `json:"puntos_actuales"`
	IntPuntosAcumulados int       `json:"puntos_acumulados_total"`
	IntPuntosCanjeados  int       `json:"puntos_canjeados_total"`
	IntPuntosExpirados  int       `json:"puntos_expirados"`
	TimeFechaUltimaActualizacion time.Time `json:"fecha_ultima_actualizacion"`
	ArrHistorialPuntos  []interface{} `json:"historial_puntos,omitempty"`
}

// structCanjePuntosDTO DTO para canje de puntos
type structCanjePuntosDTO struct {
	UuidClienteID       uuid.UUID `json:"cliente_id" binding:"required"`
	IntPuntosCanjear    int       `json:"puntos_canjear" binding:"required,min=1"`
	StrTipoCanje        string    `json:"tipo_canje" binding:"required"`
	UuidProductoID      *uuid.UUID `json:"producto_id,omitempty"`
	FltDescuentoMonto   *float64  `json:"descuento_monto,omitempty"`
	StrObservaciones    *string   `json:"observaciones,omitempty"`
}

// structCanjeResponseDTO DTO respuesta de canje
type structCanjeResponseDTO struct {
	UuidCanjeID         uuid.UUID `json:"canje_id"`
	UuidClienteID       uuid.UUID `json:"cliente_id"`
	IntPuntosCanjeados  int       `json:"puntos_canjeados"`
	IntPuntosRestantes  int       `json:"puntos_restantes"`
	FltValorCanje       float64   `json:"valor_canje"`
	StrEstado           string    `json:"estado"`
	TimeFechaCanje      time.Time `json:"fecha_canje"`
	StrCodigoReferencia string    `json:"codigo_referencia"`
}

// structAcumularPuntosDTO DTO para acumular puntos
type structAcumularPuntosDTO struct {
	UuidClienteID       uuid.UUID `json:"cliente_id" binding:"required"`
	UuidVentaID         uuid.UUID `json:"venta_id" binding:"required"`
	FltMontoCompra      float64   `json:"monto_compra" binding:"required,min=0"`
	IntPuntosBase       *int      `json:"puntos_base,omitempty"`
	FltMultiplicador    *float64  `json:"multiplicador,omitempty"`
	StrConcepto         *string   `json:"concepto,omitempty"`
}

// structAcumulacionResponseDTO DTO respuesta de acumulación
type structAcumulacionResponseDTO struct {
	UuidAcumulacionID   uuid.UUID `json:"acumulacion_id"`
	UuidClienteID       uuid.UUID `json:"cliente_id"`
	IntPuntosAcumulados int       `json:"puntos_acumulados"`
	IntPuntosTotales    int       `json:"puntos_totales"`
	FltMontoCompra      float64   `json:"monto_compra"`
	TimeFechaAcumulacion time.Time `json:"fecha_acumulacion"`
	StrReglaAplicada    *string   `json:"regla_aplicada,omitempty"`
}

// structConfiguracionPOSDTO DTO configuración POS
type structConfiguracionPOSDTO struct {
	UuidSucursalID      uuid.UUID `json:"sucursal_id"`
	BoolModoOffline     bool      `json:"modo_offline"`
	IntTiempoSincronizacion int   `json:"tiempo_sincronizacion_minutos"`
	BoolImpresionAutomatica bool  `json:"impresion_automatica"`
	StrTipoImpresora    string    `json:"tipo_impresora"`
	BoolCajonAutomatico bool      `json:"cajon_automatico"`
	MapConfiguracionHardware map[string]interface{} `json:"configuracion_hardware"`
	MapConfiguracionUI  map[string]interface{} `json:"configuracion_ui"`
}

// structDuplicarPlantillaDTO DTO para duplicar plantilla
type structDuplicarPlantillaDTO struct {
	UuidPlantillaOrigenID uuid.UUID `json:"plantilla_origen_id" binding:"required"`
	StrNuevoNombre      string    `json:"nuevo_nombre" binding:"required"`
	StrDescripcion      *string   `json:"descripcion,omitempty"`
	BoolCopiarPermisos  bool      `json:"copiar_permisos"`
	UuidSucursalID      *uuid.UUID `json:"sucursal_id,omitempty"`
}

// structValidationError DTO para errores de validación
type structValidationError struct {
	StrCampo            string    `json:"campo"`
	StrValor            string    `json:"valor"`
	StrMensaje          string    `json:"mensaje"`
	StrCodigo           string    `json:"codigo"`
	MapContexto         map[string]interface{} `json:"contexto,omitempty"`
}

// structPreviewPlantillaDTO DTO para preview de plantilla
type structPreviewPlantillaDTO struct {
	UuidPlantillaID     uuid.UUID `json:"plantilla_id" binding:"required"`
	UuidProductoID      uuid.UUID `json:"producto_id" binding:"required"`
	MapParametros       map[string]interface{} `json:"parametros,omitempty"`
	StrFormatoSalida    string    `json:"formato_salida" binding:"required,oneof=png jpg pdf"`
	IntCalidad          *int      `json:"calidad,omitempty" binding:"omitempty,min=1,max=100"`
}

// structGenerarEtiquetaDTO DTO para generar etiqueta
type structGenerarEtiquetaDTO struct {
	UuidPlantillaID     uuid.UUID `json:"plantilla_id" binding:"required"`
	ArrProductosID      []uuid.UUID `json:"productos_id" binding:"required,min=1"`
	IntCantidadCopias   int       `json:"cantidad_copias" binding:"min=1,max=1000"`
	StrFormatoSalida    string    `json:"formato_salida" binding:"required,oneof=pdf png jpg"`
	BoolGenerarLote     bool      `json:"generar_lote"`
	StrNombreLote       *string   `json:"nombre_lote,omitempty"`
	MapParametrosGlobales map[string]interface{} `json:"parametros_globales,omitempty"`
}

// structLoteEtiquetasResponseDTO DTO respuesta de lote de etiquetas
type structLoteEtiquetasResponseDTO struct {
	UuidLoteID          uuid.UUID `json:"lote_id"`
	StrNombreLote       string    `json:"nombre_lote"`
	IntTotalEtiquetas   int       `json:"total_etiquetas"`
	IntEtiquetasGeneradas int     `json:"etiquetas_generadas"`
	IntEtiquetasError   int       `json:"etiquetas_error"`
	StrEstado           string    `json:"estado"`
	FltProgreso         float64   `json:"progreso"`
	StrRutaArchivoZip   *string   `json:"ruta_archivo_zip,omitempty"`
	TimeFechaCreacion   time.Time `json:"fecha_creacion"`
	TimeFechaCompletado *time.Time `json:"fecha_completado,omitempty"`
	ArrEtiquetas        []EtiquetaResponseDTO `json:"etiquetas,omitempty"`
}

// structPlantillaEtiquetaCreateDTO DTO para crear plantilla de etiqueta
type structPlantillaEtiquetaCreateDTO struct {
	StrNombre           string    `json:"nombre" binding:"required"`
	StrDescripcion      *string   `json:"descripcion,omitempty"`
	StrTipoEtiqueta     string    `json:"tipo_etiqueta" binding:"required"`
	FltAncho            float64   `json:"ancho" binding:"required,min=0"`
	FltAlto             float64   `json:"alto" binding:"required,min=0"`
	StrUnidadMedida     string    `json:"unidad_medida" binding:"required,oneof=mm cm inch"`
	MapConfiguracionDiseno map[string]interface{} `json:"configuracion_diseno" binding:"required"`
	ArrCamposPersonalizados []interface{} `json:"campos_personalizados,omitempty"`
	BoolActiva          bool      `json:"activa"`
	ArrSucursalesPermitidas []uuid.UUID `json:"sucursales_permitidas,omitempty"`
}

// structDashboardDataDTO DTO para datos de dashboard
type structDashboardDataDTO struct {
	UuidDashboardID     uuid.UUID `json:"dashboard_id"`
	StrTitulo           string    `json:"titulo"`
	TimeFechaActualizacion time.Time `json:"fecha_actualizacion"`
	ArrWidgets          []interface{} `json:"widgets"`
	MapConfiguracion    map[string]interface{} `json:"configuracion"`
	MapFiltrosAplicados map[string]interface{} `json:"filtros_aplicados,omitempty"`
}

// Aliases finales para compatibilidad
type PuntosClienteResponseDTO = structPuntosClienteResponseDTO
type CanjePuntosDTO = structCanjePuntosDTO
type CanjeResponseDTO = structCanjeResponseDTO
type AcumularPuntosDTO = structAcumularPuntosDTO
type AcumulacionResponseDTO = structAcumulacionResponseDTO
type ConfiguracionPOSDTO = structConfiguracionPOSDTO
type DuplicarPlantillaDTO = structDuplicarPlantillaDTO
type ValidationError = structValidationError
type PreviewPlantillaDTO = structPreviewPlantillaDTO
type GenerarEtiquetaDTO = structGenerarEtiquetaDTO
type LoteEtiquetasResponseDTO = structLoteEtiquetasResponseDTO
type PlantillaEtiquetaCreateDTO = structPlantillaEtiquetaCreateDTO
type DashboardDataDTO = structDashboardDataDTO

