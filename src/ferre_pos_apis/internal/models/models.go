package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Tipos enumerados personalizados

// RolUsuario tipos de roles de usuario
type RolUsuario string

const (
	RolCajero            RolUsuario = "cajero"
	RolVendedor          RolUsuario = "vendedor"
	RolDespacho          RolUsuario = "despacho"
	RolSupervisor        RolUsuario = "supervisor"
	RolAdmin             RolUsuario = "admin"
	RolOperadorEtiquetas RolUsuario = "operador_etiquetas"
)

// EstadoDocumento estados de documentos
type EstadoDocumento string

const (
	EstadoPendiente  EstadoDocumento = "pendiente"
	EstadoProcesado  EstadoDocumento = "procesado"
	EstadoEnviado    EstadoDocumento = "enviado"
	EstadoRechazado  EstadoDocumento = "rechazado"
	EstadoAnulado    EstadoDocumento = "anulado"
)

// EstadoSincronizacion estados de sincronización
type EstadoSincronizacion string

const (
	SincPendiente   EstadoSincronizacion = "pendiente"
	SincEnProceso   EstadoSincronizacion = "en_proceso"
	SincCompletado  EstadoSincronizacion = "completado"
	SincError       EstadoSincronizacion = "error"
)

// PrioridadProceso prioridades de procesos API
type PrioridadProceso string

const (
	PrioridadMaxima PrioridadProceso = "maxima"  // api_pos
	PrioridadMedia  PrioridadProceso = "media"   // api_sync
	PrioridadBaja   PrioridadProceso = "baja"    // api_labels
	PrioridadMinima PrioridadProceso = "minima"  // api_reports
)

// EstadoTrabajoEtiqueta estados de trabajos de etiquetas
type EstadoTrabajoEtiqueta string

const (
	TrabajoEtiquetaPendiente   EstadoTrabajoEtiqueta = "pendiente"
	TrabajoEtiquetaProcesando  EstadoTrabajoEtiqueta = "procesando"
	TrabajoEtiquetaCompletado  EstadoTrabajoEtiqueta = "completado"
	TrabajoEtiquetaError       EstadoTrabajoEtiqueta = "error"
	TrabajoEtiquetaCancelado   EstadoTrabajoEtiqueta = "cancelado"
)

// JSONB tipo personalizado para campos JSONB de PostgreSQL
type JSONB map[string]interface{}

// Value implementa driver.Valuer para JSONB
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implementa sql.Scanner para JSONB
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("no se puede escanear %T en JSONB", value)
	}

	return json.Unmarshal(bytes, j)
}

// Modelos principales

// Sucursal modelo de sucursal
type Sucursal struct {
	ID                         uuid.UUID `json:"id" db:"id"`
	Codigo                     string    `json:"codigo" db:"codigo" validate:"required,max=50"`
	Nombre                     string    `json:"nombre" db:"nombre" validate:"required,max=200"`
	Direccion                  *string   `json:"direccion,omitempty" db:"direccion"`
	Comuna                     *string   `json:"comuna,omitempty" db:"comuna"`
	Region                     *string   `json:"region,omitempty" db:"region"`
	Telefono                   *string   `json:"telefono,omitempty" db:"telefono"`
	Email                      *string   `json:"email,omitempty" db:"email" validate:"omitempty,email"`
	HorarioApertura           *time.Time `json:"horario_apertura,omitempty" db:"horario_apertura"`
	HorarioCierre             *time.Time `json:"horario_cierre,omitempty" db:"horario_cierre"`
	Timezone                   string    `json:"timezone" db:"timezone"`
	Habilitada                 bool      `json:"habilitada" db:"habilitada"`
	FechaCreacion             time.Time `json:"fecha_creacion" db:"fecha_creacion"`
	FechaModificacion         time.Time `json:"fecha_modificacion" db:"fecha_modificacion"`
	ConfiguracionDTE          JSONB     `json:"configuracion_dte,omitempty" db:"configuracion_dte"`
	ConfiguracionPagos        JSONB     `json:"configuracion_pagos,omitempty" db:"configuracion_pagos"`
	MaxConexionesConcurrentes int       `json:"max_conexiones_concurrentes" db:"max_conexiones_concurrentes"`
	ConfiguracionCache        JSONB     `json:"configuracion_cache,omitempty" db:"configuracion_cache"`
	MetricasRendimiento       JSONB     `json:"metricas_rendimiento,omitempty" db:"metricas_rendimiento"`
}

// Usuario modelo de usuario
type Usuario struct {
	ID                    uuid.UUID  `json:"id" db:"id"`
	RUT                   string     `json:"rut" db:"rut" validate:"required,rut"`
	Nombre                string     `json:"nombre" db:"nombre" validate:"required,max=100"`
	Apellido              *string    `json:"apellido,omitempty" db:"apellido"`
	Email                 *string    `json:"email,omitempty" db:"email" validate:"omitempty,email"`
	Telefono              *string    `json:"telefono,omitempty" db:"telefono"`
	Rol                   RolUsuario `json:"rol" db:"rol" validate:"required"`
	SucursalID            *uuid.UUID `json:"sucursal_id,omitempty" db:"sucursal_id"`
	PasswordHash          string     `json:"-" db:"password_hash"`
	Salt                  string     `json:"-" db:"salt"`
	Activo                bool       `json:"activo" db:"activo"`
	UltimoAcceso          *time.Time `json:"ultimo_acceso,omitempty" db:"ultimo_acceso"`
	IntentosFallidos      int        `json:"intentos_fallidos" db:"intentos_fallidos"`
	BloqueadoHasta        *time.Time `json:"bloqueado_hasta,omitempty" db:"bloqueado_hasta"`
	FechaCreacion         time.Time  `json:"fecha_creacion" db:"fecha_creacion"`
	FechaModificacion     time.Time  `json:"fecha_modificacion" db:"fecha_modificacion"`
	ConfiguracionPersonal JSONB      `json:"configuracion_personal,omitempty" db:"configuracion_personal"`
	PermisosEspeciales    JSONB      `json:"permisos_especiales,omitempty" db:"permisos_especiales"`
	CachePermisos         JSONB      `json:"cache_permisos,omitempty" db:"cache_permisos"`
	HashSesionActiva      *string    `json:"-" db:"hash_sesion_activa"`
	UltimoTerminalID      *uuid.UUID `json:"ultimo_terminal_id,omitempty" db:"ultimo_terminal_id"`
	PreferenciasUI        JSONB      `json:"preferencias_ui,omitempty" db:"preferencias_ui"`
}

// CategoriaProducto modelo de categoría de producto
type CategoriaProducto struct {
	ID                      uuid.UUID  `json:"id" db:"id"`
	Codigo                  string     `json:"codigo" db:"codigo" validate:"required,max=50"`
	Nombre                  string     `json:"nombre" db:"nombre" validate:"required,max=200"`
	Descripcion             *string    `json:"descripcion,omitempty" db:"descripcion"`
	CategoriaPadreID        *uuid.UUID `json:"categoria_padre_id,omitempty" db:"categoria_padre_id"`
	Nivel                   int        `json:"nivel" db:"nivel"`
	Activa                  bool       `json:"activa" db:"activa"`
	OrdenVisualizacion      *int       `json:"orden_visualizacion,omitempty" db:"orden_visualizacion"`
	ImagenURL               *string    `json:"imagen_url,omitempty" db:"imagen_url"`
	FechaCreacion           time.Time  `json:"fecha_creacion" db:"fecha_creacion"`
	FechaModificacion       time.Time  `json:"fecha_modificacion" db:"fecha_modificacion"`
	PathCompleto            *string    `json:"path_completo,omitempty" db:"path_completo"`
	TotalProductos          int        `json:"total_productos" db:"total_productos"`
	ConfiguracionEtiquetas  JSONB      `json:"configuracion_etiquetas,omitempty" db:"configuracion_etiquetas"`
}

// Producto modelo de producto
type Producto struct {
	ID                           uuid.UUID  `json:"id" db:"id"`
	CodigoInterno                string     `json:"codigo_interno" db:"codigo_interno" validate:"required,max=50"`
	CodigoBarra                  string     `json:"codigo_barra" db:"codigo_barra" validate:"required,max=50"`
	Descripcion                  string     `json:"descripcion" db:"descripcion" validate:"required,max=500"`
	DescripcionCorta             *string    `json:"descripcion_corta,omitempty" db:"descripcion_corta"`
	CategoriaID                  *uuid.UUID `json:"categoria_id,omitempty" db:"categoria_id"`
	Marca                        *string    `json:"marca,omitempty" db:"marca"`
	Modelo                       *string    `json:"modelo,omitempty" db:"modelo"`
	PrecioUnitario               float64    `json:"precio_unitario" db:"precio_unitario" validate:"gte=0"`
	PrecioCosto                  *float64   `json:"precio_costo,omitempty" db:"precio_costo" validate:"omitempty,gte=0"`
	UnidadMedida                 string     `json:"unidad_medida" db:"unidad_medida"`
	Peso                         *float64   `json:"peso,omitempty" db:"peso"`
	Dimensiones                  JSONB      `json:"dimensiones,omitempty" db:"dimensiones"`
	EspecificacionesTecnicas     JSONB      `json:"especificaciones_tecnicas,omitempty" db:"especificaciones_tecnicas"`
	Activo                       bool       `json:"activo" db:"activo"`
	RequiereSerie                bool       `json:"requiere_serie" db:"requiere_serie"`
	PermiteFraccionamiento       bool       `json:"permite_fraccionamiento" db:"permite_fraccionamiento"`
	StockMinimo                  int        `json:"stock_minimo" db:"stock_minimo"`
	StockMaximo                  *int       `json:"stock_maximo,omitempty" db:"stock_maximo"`
	ImagenPrincipalURL           *string    `json:"imagen_principal_url,omitempty" db:"imagen_principal_url"`
	ImagenesAdicionales          JSONB      `json:"imagenes_adicionales,omitempty" db:"imagenes_adicionales"`
	FechaCreacion                time.Time  `json:"fecha_creacion" db:"fecha_creacion"`
	FechaModificacion            time.Time  `json:"fecha_modificacion" db:"fecha_modificacion"`
	UsuarioCreacion              *uuid.UUID `json:"usuario_creacion,omitempty" db:"usuario_creacion"`
	UsuarioModificacion          *uuid.UUID `json:"usuario_modificacion,omitempty" db:"usuario_modificacion"`
	DescripcionBusqueda          *string    `json:"-" db:"descripcion_busqueda"`
	PopularidadScore             float64    `json:"popularidad_score" db:"popularidad_score"`
	CacheCodigoBarrasGenerado    *string    `json:"cache_codigo_barras_generado,omitempty" db:"cache_codigo_barras_generado"`
	ConfiguracionEtiqueta        JSONB      `json:"configuracion_etiqueta,omitempty" db:"configuracion_etiqueta"`
	FechaUltimaEtiqueta          *time.Time `json:"fecha_ultima_etiqueta,omitempty" db:"fecha_ultima_etiqueta"`
	TotalEtiquetasGeneradas      int        `json:"total_etiquetas_generadas" db:"total_etiquetas_generadas"`
}

// StockCentral modelo de stock central
type StockCentral struct {
	ProductoID           uuid.UUID  `json:"producto_id" db:"producto_id"`
	SucursalID           uuid.UUID  `json:"sucursal_id" db:"sucursal_id"`
	Cantidad             int        `json:"cantidad" db:"cantidad"`
	CantidadReservada    int        `json:"cantidad_reservada" db:"cantidad_reservada"`
	CantidadDisponible   int        `json:"cantidad_disponible" db:"cantidad_disponible"`
	CostoPromedio        *float64   `json:"costo_promedio,omitempty" db:"costo_promedio"`
	FechaUltimaEntrada   *time.Time `json:"fecha_ultima_entrada,omitempty" db:"fecha_ultima_entrada"`
	FechaUltimaSalida    *time.Time `json:"fecha_ultima_salida,omitempty" db:"fecha_ultima_salida"`
	FechaSync            time.Time  `json:"fecha_sync" db:"fecha_sync"`
	VersionOptimisticLock int       `json:"version_optimistic_lock" db:"version_optimistic_lock"`
	CacheValidez         time.Time  `json:"cache_validez" db:"cache_validez"`
	AlertasConfiguradas  JSONB      `json:"alertas_configuradas,omitempty" db:"alertas_configuradas"`
}

// MovimientoStock modelo de movimiento de stock
type MovimientoStock struct {
	ID                   uuid.UUID         `json:"id" db:"id"`
	ProductoID           uuid.UUID         `json:"producto_id" db:"producto_id"`
	SucursalID           uuid.UUID         `json:"sucursal_id" db:"sucursal_id"`
	TipoMovimiento       string            `json:"tipo_movimiento" db:"tipo_movimiento"`
	Cantidad             int               `json:"cantidad" db:"cantidad"`
	CantidadAnterior     *int              `json:"cantidad_anterior,omitempty" db:"cantidad_anterior"`
	CantidadNueva        *int              `json:"cantidad_nueva,omitempty" db:"cantidad_nueva"`
	CostoUnitario        *float64          `json:"costo_unitario,omitempty" db:"costo_unitario"`
	DocumentoReferencia  *string           `json:"documento_referencia,omitempty" db:"documento_referencia"`
	UsuarioID            *uuid.UUID        `json:"usuario_id,omitempty" db:"usuario_id"`
	Fecha                time.Time         `json:"fecha" db:"fecha"`
	Observaciones        *string           `json:"observaciones,omitempty" db:"observaciones"`
	DatosAdicionales     JSONB             `json:"datos_adicionales,omitempty" db:"datos_adicionales"`
	ProcesoOrigen        *PrioridadProceso `json:"proceso_origen,omitempty" db:"proceso_origen"`
	BatchID              *uuid.UUID        `json:"batch_id,omitempty" db:"batch_id"`
}

// Terminal modelo de terminal
type Terminal struct {
	ID                     uuid.UUID `json:"id" db:"id"`
	Codigo                 string    `json:"codigo" db:"codigo" validate:"required,max=50"`
	NombreTerminal         string    `json:"nombre_terminal" db:"nombre_terminal" validate:"required,max=200"`
	TipoTerminal           string    `json:"tipo_terminal" db:"tipo_terminal" validate:"required"`
	SucursalID             uuid.UUID `json:"sucursal_id" db:"sucursal_id"`
	DireccionIP            *string   `json:"direccion_ip,omitempty" db:"direccion_ip"`
	DireccionMAC           *string   `json:"direccion_mac,omitempty" db:"direccion_mac"`
	Activo                 bool      `json:"activo" db:"activo"`
	UltimaConexion         *time.Time `json:"ultima_conexion,omitempty" db:"ultima_conexion"`
	VersionSoftware        *string   `json:"version_software,omitempty" db:"version_software"`
	Configuracion          JSONB     `json:"configuracion,omitempty" db:"configuracion"`
	FechaInstalacion       time.Time `json:"fecha_instalacion" db:"fecha_instalacion"`
	FechaModificacion      time.Time `json:"fecha_modificacion" db:"fecha_modificacion"`
	HeartbeatInterval      int       `json:"heartbeat_interval" db:"heartbeat_interval"`
	EstadoConexion         string    `json:"estado_conexion" db:"estado_conexion"`
	MetricasRendimiento    JSONB     `json:"metricas_rendimiento,omitempty" db:"metricas_rendimiento"`
	ConfiguracionCache     JSONB     `json:"configuracion_cache,omitempty" db:"configuracion_cache"`
}

// Venta modelo de venta
type Venta struct {
	ID                      uuid.UUID         `json:"id" db:"id"`
	NumeroVenta             int64             `json:"numero_venta" db:"numero_venta"`
	SucursalID              uuid.UUID         `json:"sucursal_id" db:"sucursal_id"`
	TerminalID              uuid.UUID         `json:"terminal_id" db:"terminal_id"`
	CajeroID                uuid.UUID         `json:"cajero_id" db:"cajero_id"`
	VendedorID              *uuid.UUID        `json:"vendedor_id,omitempty" db:"vendedor_id"`
	ClienteRUT              *string           `json:"cliente_rut,omitempty" db:"cliente_rut"`
	ClienteNombre           *string           `json:"cliente_nombre,omitempty" db:"cliente_nombre"`
	NotaVentaID             *uuid.UUID        `json:"nota_venta_id,omitempty" db:"nota_venta_id"`
	TipoDocumento           string            `json:"tipo_documento" db:"tipo_documento"`
	Subtotal                float64           `json:"subtotal" db:"subtotal"`
	DescuentoTotal          float64           `json:"descuento_total" db:"descuento_total"`
	ImpuestoTotal           float64           `json:"impuesto_total" db:"impuesto_total"`
	Total                   float64           `json:"total" db:"total"`
	Estado                  string            `json:"estado" db:"estado"`
	Fecha                   time.Time         `json:"fecha" db:"fecha"`
	FechaAnulacion          *time.Time        `json:"fecha_anulacion,omitempty" db:"fecha_anulacion"`
	MotivoAnulacion         *string           `json:"motivo_anulacion,omitempty" db:"motivo_anulacion"`
	UsuarioAnulacion        *uuid.UUID        `json:"usuario_anulacion,omitempty" db:"usuario_anulacion"`
	DTEID                   *uuid.UUID        `json:"dte_id,omitempty" db:"dte_id"`
	DTEEmitido              bool              `json:"dte_emitido" db:"dte_emitido"`
	Sincronizada            bool              `json:"sincronizada" db:"sincronizada"`
	FechaSincronizacion     *time.Time        `json:"fecha_sincronizacion,omitempty" db:"fecha_sincronizacion"`
	DatosAdicionales        JSONB             `json:"datos_adicionales,omitempty" db:"datos_adicionales"`
	HashIntegridad          *string           `json:"hash_integridad,omitempty" db:"hash_integridad"`
	TiempoProcesamiento     *int              `json:"tiempo_procesamiento_ms,omitempty" db:"tiempo_procesamiento_ms"`
	ProcesoOrigen           PrioridadProceso  `json:"proceso_origen" db:"proceso_origen"`
	CacheTotales            JSONB             `json:"cache_totales,omitempty" db:"cache_totales"`
}

// DetalleVenta modelo de detalle de venta
type DetalleVenta struct {
	ID                   uuid.UUID  `json:"id" db:"id"`
	VentaID              uuid.UUID  `json:"venta_id" db:"venta_id"`
	ProductoID           uuid.UUID  `json:"producto_id" db:"producto_id"`
	Cantidad             float64    `json:"cantidad" db:"cantidad"`
	PrecioUnitario       float64    `json:"precio_unitario" db:"precio_unitario"`
	DescuentoUnitario    float64    `json:"descuento_unitario" db:"descuento_unitario"`
	PrecioFinal          float64    `json:"precio_final" db:"precio_final"`
	TotalItem            float64    `json:"total_item" db:"total_item"`
	NumeroSerie          *string    `json:"numero_serie,omitempty" db:"numero_serie"`
	Lote                 *string    `json:"lote,omitempty" db:"lote"`
	FechaVencimiento     *time.Time `json:"fecha_vencimiento,omitempty" db:"fecha_vencimiento"`
	DatosAdicionales     JSONB      `json:"datos_adicionales,omitempty" db:"datos_adicionales"`
	MargenUnitario       *float64   `json:"margen_unitario,omitempty" db:"margen_unitario"`
	CategoriaProductoID  *uuid.UUID `json:"categoria_producto_id,omitempty" db:"categoria_producto_id"`
}

// Modelos específicos para etiquetas

// EtiquetaPlantilla modelo de plantilla de etiqueta
type EtiquetaPlantilla struct {
	ID                           uuid.UUID  `json:"id" db:"id"`
	Codigo                       string     `json:"codigo" db:"codigo" validate:"required,max=50"`
	Nombre                       string     `json:"nombre" db:"nombre" validate:"required,max=200"`
	Descripcion                  *string    `json:"descripcion,omitempty" db:"descripcion"`
	CategoriaProductoID          *uuid.UUID `json:"categoria_producto_id,omitempty" db:"categoria_producto_id"`
	TipoEtiqueta                 string     `json:"tipo_etiqueta" db:"tipo_etiqueta"`
	AnchoMM                      float64    `json:"ancho_mm" db:"ancho_mm"`
	AltoMM                       float64    `json:"alto_mm" db:"alto_mm"`
	Orientacion                  string     `json:"orientacion" db:"orientacion"`
	ConfiguracionDiseno          JSONB      `json:"configuracion_diseno" db:"configuracion_diseno"`
	ConfiguracionCodigoBarra     JSONB      `json:"configuracion_codigo_barras,omitempty" db:"configuracion_codigo_barras"`
	Activa                       bool       `json:"activa" db:"activa"`
	Predeterminada               bool       `json:"predeterminada" db:"predeterminada"`
	FechaCreacion                time.Time  `json:"fecha_creacion" db:"fecha_creacion"`
	FechaModificacion            time.Time  `json:"fecha_modificacion" db:"fecha_modificacion"`
	UsuarioCreacion              *uuid.UUID `json:"usuario_creacion,omitempty" db:"usuario_creacion"`
	UsuarioModificacion          *uuid.UUID `json:"usuario_modificacion,omitempty" db:"usuario_modificacion"`
	TotalUsos                    int        `json:"total_usos" db:"total_usos"`
	TiempoRenderizadoPromedioMS  *int       `json:"tiempo_renderizado_promedio_ms,omitempty" db:"tiempo_renderizado_promedio_ms"`
	CachePreview                 *string    `json:"cache_preview,omitempty" db:"cache_preview"`
}

// EtiquetaTrabajoImpresion modelo de trabajo de impresión de etiquetas
type EtiquetaTrabajoImpresion struct {
	ID                         uuid.UUID              `json:"id" db:"id"`
	NumeroTrabajo              int64                  `json:"numero_trabajo" db:"numero_trabajo"`
	UsuarioID                  uuid.UUID              `json:"usuario_id" db:"usuario_id"`
	SucursalID                 uuid.UUID              `json:"sucursal_id" db:"sucursal_id"`
	TerminalID                 uuid.UUID              `json:"terminal_id" db:"terminal_id"`
	PlantillaID                uuid.UUID              `json:"plantilla_id" db:"plantilla_id"`
	TipoTrabajo                string                 `json:"tipo_trabajo" db:"tipo_trabajo"`
	Estado                     EstadoTrabajoEtiqueta  `json:"estado" db:"estado"`
	TotalEtiquetas             int                    `json:"total_etiquetas" db:"total_etiquetas"`
	EtiquetasProcesadas        int                    `json:"etiquetas_procesadas" db:"etiquetas_procesadas"`
	EtiquetasExitosas          int                    `json:"etiquetas_exitosas" db:"etiquetas_exitosas"`
	EtiquetasError             int                    `json:"etiquetas_error" db:"etiquetas_error"`
	FechaSolicitud             time.Time              `json:"fecha_solicitud" db:"fecha_solicitud"`
	FechaInicioProcesamiento   *time.Time             `json:"fecha_inicio_procesamiento,omitempty" db:"fecha_inicio_procesamiento"`
	FechaFinProcesamiento      *time.Time             `json:"fecha_fin_procesamiento,omitempty" db:"fecha_fin_procesamiento"`
	TiempoProcesamiento        *int                   `json:"tiempo_procesamiento_ms,omitempty" db:"tiempo_procesamiento_ms"`
	ConfiguracionImpresora     JSONB                  `json:"configuracion_impresora,omitempty" db:"configuracion_impresora"`
	ParametrosTrabajo          JSONB                  `json:"parametros_trabajo,omitempty" db:"parametros_trabajo"`
	ErroresDetalle             JSONB                  `json:"errores_detalle,omitempty" db:"errores_detalle"`
	ArchivoGeneradoPath        *string                `json:"archivo_generado_path,omitempty" db:"archivo_generado_path"`
	HashArchivo                *string                `json:"hash_archivo,omitempty" db:"hash_archivo"`
	Prioridad                  int                    `json:"prioridad" db:"prioridad"`
	ProcesoAPI                 PrioridadProceso       `json:"proceso_api" db:"proceso_api"`
	RecursosUtilizados         JSONB                  `json:"recursos_utilizados,omitempty" db:"recursos_utilizados"`
}

// Respuestas de API

// APIResponse respuesta estándar de API
type APIResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     *APIError   `json:"error,omitempty"`
	Meta      *APIMeta    `json:"meta,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// APIError estructura de error de API
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details JSONB  `json:"details,omitempty"`
}

// APIMeta metadatos de respuesta de API
type APIMeta struct {
	Page       int `json:"page,omitempty"`
	PerPage    int `json:"per_page,omitempty"`
	Total      int `json:"total,omitempty"`
	TotalPages int `json:"total_pages,omitempty"`
}

// Requests de API

// LoginRequest request de login
type LoginRequest struct {
	RUT      string `json:"rut" validate:"required,rut"`
	Password string `json:"password" validate:"required,min=6"`
	Terminal string `json:"terminal,omitempty"`
}

// LoginResponse response de login
type LoginResponse struct {
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	User         Usuario   `json:"user"`
}

// ProductoSearchRequest request de búsqueda de productos
type ProductoSearchRequest struct {
	Query        string     `json:"query,omitempty"`
	CategoriaID  *uuid.UUID `json:"categoria_id,omitempty"`
	SucursalID   uuid.UUID  `json:"sucursal_id" validate:"required"`
	ConStock     bool       `json:"con_stock,omitempty"`
	Activos      bool       `json:"activos,omitempty"`
	Page         int        `json:"page,omitempty"`
	PerPage      int        `json:"per_page,omitempty"`
}

// VentaRequest request de creación de venta
type VentaRequest struct {
	SucursalID      uuid.UUID            `json:"sucursal_id" validate:"required"`
	TerminalID      uuid.UUID            `json:"terminal_id" validate:"required"`
	CajeroID        uuid.UUID            `json:"cajero_id" validate:"required"`
	VendedorID      *uuid.UUID           `json:"vendedor_id,omitempty"`
	ClienteRUT      *string              `json:"cliente_rut,omitempty"`
	ClienteNombre   *string              `json:"cliente_nombre,omitempty"`
	TipoDocumento   string               `json:"tipo_documento" validate:"required,oneof=boleta factura guia nota_venta"`
	Items           []VentaItemRequest   `json:"items" validate:"required,min=1,dive"`
	MediosPago      []MedioPagoRequest   `json:"medios_pago" validate:"required,min=1,dive"`
	DatosAdicionales JSONB               `json:"datos_adicionales,omitempty"`
}

// VentaItemRequest item de venta
type VentaItemRequest struct {
	ProductoID        uuid.UUID `json:"producto_id" validate:"required"`
	Cantidad          float64   `json:"cantidad" validate:"required,gt=0"`
	PrecioUnitario    float64   `json:"precio_unitario" validate:"required,gte=0"`
	DescuentoUnitario float64   `json:"descuento_unitario,omitempty"`
	NumeroSerie       *string   `json:"numero_serie,omitempty"`
	Lote              *string   `json:"lote,omitempty"`
}

// MedioPagoRequest medio de pago
type MedioPagoRequest struct {
	MedioPago             string  `json:"medio_pago" validate:"required,oneof=efectivo tarjeta_debito tarjeta_credito transferencia cheque puntos_fidelizacion otro"`
	Monto                 float64 `json:"monto" validate:"required,gt=0"`
	ReferenciaTransaccion *string `json:"referencia_transaccion,omitempty"`
	CodigoAutorizacion    *string `json:"codigo_autorizacion,omitempty"`
}

// EtiquetaGenerarRequest request de generación de etiquetas
type EtiquetaGenerarRequest struct {
	PlantillaID       uuid.UUID   `json:"plantilla_id" validate:"required"`
	ProductosIDs      []uuid.UUID `json:"productos_ids" validate:"required,min=1"`
	Cantidad          int         `json:"cantidad" validate:"required,min=1,max=1000"`
	FormatoSalida     string      `json:"formato_salida" validate:"required,oneof=pdf png zpl"`
	ParametrosEspeciales JSONB    `json:"parametros_especiales,omitempty"`
}

// Funciones de validación personalizadas

// IsValidRUT valida formato de RUT chileno
func IsValidRUT(rut string) bool {
	// Implementación básica de validación de RUT
	// En producción debería incluir validación de dígito verificador
	return len(rut) >= 9 && len(rut) <= 12
}

// TableName métodos para especificar nombres de tabla

func (Sucursal) TableName() string                    { return "sucursales" }
func (Usuario) TableName() string                     { return "usuarios" }
func (CategoriaProducto) TableName() string           { return "categorias_productos" }
func (Producto) TableName() string                    { return "productos" }
func (StockCentral) TableName() string                { return "stock_central" }
func (MovimientoStock) TableName() string             { return "movimientos_stock" }
func (Terminal) TableName() string                    { return "terminales" }
func (Venta) TableName() string                       { return "ventas" }
func (DetalleVenta) TableName() string                { return "detalle_ventas" }
func (EtiquetaPlantilla) TableName() string           { return "etiquetas_plantillas" }
func (EtiquetaTrabajoImpresion) TableName() string    { return "etiquetas_trabajos_impresion" }

