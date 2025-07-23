package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// PlantillaReporte representa una plantilla para generar reportes
type PlantillaReporte struct {
	BaseModel
	Nombre                string                    `json:"nombre" db:"nombre" binding:"required" validate:"required,max=255"`
	Descripcion           *string                   `json:"descripcion,omitempty" db:"descripcion"`
	TipoReporte           TipoReporte               `json:"tipo_reporte" db:"tipo_reporte" binding:"required"`
	CategoriaReporte      CategoriaReporte          `json:"categoria_reporte" db:"categoria_reporte" binding:"required"`
	ConfiguracionConsulta ConfiguracionConsulta     `json:"configuracion_consulta" db:"configuracion_consulta" binding:"required"`
	ConfiguracionFormato  ConfiguracionFormato      `json:"configuracion_formato" db:"configuracion_formato" binding:"required"`
	ParametrosRequeridos  []ParametroReporte        `json:"parametros_requeridos,omitempty" db:"parametros_requeridos"`
	ParametrosOpcionales  []ParametroReporte        `json:"parametros_opcionales,omitempty" db:"parametros_opcionales"`
	Activa                bool                      `json:"activa" db:"activa" default:"true"`
	EsPublica             bool                      `json:"es_publica" db:"es_publica" default:"false"`
	RolesPermitidos       []RolUsuario              `json:"roles_permitidos,omitempty" db:"roles_permitidos"`
	SucursalesPermitidas  []uuid.UUID               `json:"sucursales_permitidas,omitempty" db:"sucursales_permitidas"`
	TotalGeneraciones     int                       `json:"total_generaciones" db:"total_generaciones" default:"0"`
	FechaUltimoUso        *time.Time                `json:"fecha_ultimo_uso,omitempty" db:"fecha_ultimo_uso"`
	TiempoPromedioEjecucion *int                    `json:"tiempo_promedio_ejecucion_ms,omitempty" db:"tiempo_promedio_ejecucion_ms"`
	ConfiguracionCache    *ConfiguracionCacheReporte `json:"configuracion_cache,omitempty" db:"configuracion_cache"`
	VersionPlantilla      int                       `json:"version_plantilla" db:"version_plantilla" default:"1"`
	MetadatosPlantilla    *MetadatosPlantilla       `json:"metadatos_plantilla,omitempty" db:"metadatos_plantilla"`

	// Relaciones
	ReportesGenerados []ReporteGenerado `json:"reportes_generados,omitempty" gorm:"foreignKey:PlantillaID"`
}

// ReporteGenerado representa un reporte generado
type ReporteGenerado struct {
	BaseModel
	PlantillaID           uuid.UUID                 `json:"plantilla_id" db:"plantilla_id" binding:"required"`
	SucursalID            *uuid.UUID                `json:"sucursal_id,omitempty" db:"sucursal_id"`
	UsuarioID             uuid.UUID                 `json:"usuario_id" db:"usuario_id" binding:"required"`
	Nombre                string                    `json:"nombre" db:"nombre" binding:"required"`
	Descripcion           *string                   `json:"descripcion,omitempty" db:"descripcion"`
	ParametrosUsados      ParametrosReporte         `json:"parametros_usados" db:"parametros_usados" binding:"required"`
	Estado                EstadoReporte             `json:"estado" db:"estado" default:"generando"`
	FechaGeneracion       time.Time                 `json:"fecha_generacion" db:"fecha_generacion" default:"NOW()"`
	FechaCompletado       *time.Time                `json:"fecha_completado,omitempty" db:"fecha_completado"`
	TiempoEjecucion       *int                      `json:"tiempo_ejecucion_ms,omitempty" db:"tiempo_ejecucion_ms"`
	TotalRegistros        *int                      `json:"total_registros,omitempty" db:"total_registros"`
	TamañoArchivo         *int                      `json:"tamaño_archivo_bytes,omitempty" db:"tamaño_archivo_bytes"`
	FormatoSalida         FormatoReporte            `json:"formato_salida" db:"formato_salida" binding:"required"`
	ArchivoGenerado       *string                   `json:"archivo_generado,omitempty" db:"archivo_generado"`
	URLDescarga           *string                   `json:"url_descarga,omitempty" db:"url_descarga"`
	FechaExpiracion       *time.Time                `json:"fecha_expiracion,omitempty" db:"fecha_expiracion"`
	ErrorDetalle          *string                   `json:"error_detalle,omitempty" db:"error_detalle"`
	CodigoError           *string                   `json:"codigo_error,omitempty" db:"codigo_error"`
	MetadatosGeneracion   *MetadatosGeneracionReporte `json:"metadatos_generacion,omitempty" db:"metadatos_generacion"`
	ConfiguracionUsada    *ConfiguracionReporteUsada `json:"configuracion_usada,omitempty" db:"configuracion_usada"`
	HashIntegridad        *string                   `json:"hash_integridad,omitempty" db:"hash_integridad"`
	EsPublico             bool                      `json:"es_publico" db:"es_publico" default:"false"`
	CompartidoCon         []uuid.UUID               `json:"compartido_con,omitempty" db:"compartido_con"`
	TotalDescargas        int                       `json:"total_descargas" db:"total_descargas" default:"0"`

	// Relaciones
	Plantilla *PlantillaReporte `json:"plantilla,omitempty" gorm:"foreignKey:PlantillaID"`
	Sucursal  *Sucursal         `json:"sucursal,omitempty" gorm:"foreignKey:SucursalID"`
	Usuario   *Usuario          `json:"usuario,omitempty" gorm:"foreignKey:UsuarioID"`
}

// ReporteProgramado representa un reporte programado para ejecución automática
type ReporteProgramado struct {
	BaseModel
	PlantillaID           uuid.UUID                 `json:"plantilla_id" db:"plantilla_id" binding:"required"`
	SucursalID            *uuid.UUID                `json:"sucursal_id,omitempty" db:"sucursal_id"`
	UsuarioID             uuid.UUID                 `json:"usuario_id" db:"usuario_id" binding:"required"`
	Nombre                string                    `json:"nombre" db:"nombre" binding:"required"`
	Descripcion           *string                   `json:"descripcion,omitempty" db:"descripcion"`
	ParametrosProgramados ParametrosReporte         `json:"parametros_programados" db:"parametros_programados" binding:"required"`
	ConfiguracionCron     ConfiguracionCron         `json:"configuracion_cron" db:"configuracion_cron" binding:"required"`
	Activo                bool                      `json:"activo" db:"activo" default:"true"`
	ProximaEjecucion      *time.Time                `json:"proxima_ejecucion,omitempty" db:"proxima_ejecucion"`
	UltimaEjecucion       *time.Time                `json:"ultima_ejecucion,omitempty" db:"ultima_ejecucion"`
	TotalEjecuciones      int                       `json:"total_ejecuciones" db:"total_ejecuciones" default:"0"`
	EjecucionesExitosas   int                       `json:"ejecuciones_exitosas" db:"ejecuciones_exitosas" default:"0"`
	EjecucionesError      int                       `json:"ejecuciones_error" db:"ejecuciones_error" default:"0"`
	UltimoError           *string                   `json:"ultimo_error,omitempty" db:"ultimo_error"`
	FechaUltimoError      *time.Time                `json:"fecha_ultimo_error,omitempty" db:"fecha_ultimo_error"`
	ConfiguracionNotificacion *ConfiguracionNotificacionReporte `json:"configuracion_notificacion,omitempty" db:"configuracion_notificacion"`
	ConfiguracionRetencion *ConfiguracionRetencion   `json:"configuracion_retencion,omitempty" db:"configuracion_retencion"`
	MetadatosProgramacion *MetadatosProgramacion    `json:"metadatos_programacion,omitempty" db:"metadatos_programacion"`

	// Relaciones
	Plantilla         *PlantillaReporte `json:"plantilla,omitempty" gorm:"foreignKey:PlantillaID"`
	Sucursal          *Sucursal         `json:"sucursal,omitempty" gorm:"foreignKey:SucursalID"`
	Usuario           *Usuario          `json:"usuario,omitempty" gorm:"foreignKey:UsuarioID"`
	ReportesGenerados []ReporteGenerado `json:"reportes_generados,omitempty" gorm:"foreignKey:ReporteProgramadoID"`
}

// DashboardPersonalizado representa un dashboard personalizado
type DashboardPersonalizado struct {
	BaseModel
	UsuarioID             uuid.UUID                 `json:"usuario_id" db:"usuario_id" binding:"required"`
	SucursalID            *uuid.UUID                `json:"sucursal_id,omitempty" db:"sucursal_id"`
	Nombre                string                    `json:"nombre" db:"nombre" binding:"required"`
	Descripcion           *string                   `json:"descripcion,omitempty" db:"descripcion"`
	ConfiguracionLayout   ConfiguracionLayoutDashboard `json:"configuracion_layout" db:"configuracion_layout" binding:"required"`
	Widgets               []WidgetDashboard         `json:"widgets" db:"widgets" binding:"required"`
	EsPublico             bool                      `json:"es_publico" db:"es_publico" default:"false"`
	CompartidoCon         []uuid.UUID               `json:"compartido_con,omitempty" db:"compartido_con"`
	ConfiguracionRefresh  ConfiguracionRefreshDashboard `json:"configuracion_refresh" db:"configuracion_refresh"`
	TotalVisualizaciones  int                       `json:"total_visualizaciones" db:"total_visualizaciones" default:"0"`
	FechaUltimaVisualizacion *time.Time             `json:"fecha_ultima_visualizacion,omitempty" db:"fecha_ultima_visualizacion"`
	MetadatosDashboard    *MetadatosDashboard       `json:"metadatos_dashboard,omitempty" db:"metadatos_dashboard"`
	VersionDashboard      int                       `json:"version_dashboard" db:"version_dashboard" default:"1"`

	// Relaciones
	Usuario  *Usuario  `json:"usuario,omitempty" gorm:"foreignKey:UsuarioID"`
	Sucursal *Sucursal `json:"sucursal,omitempty" gorm:"foreignKey:SucursalID"`
}

// Enums para reportes

type TipoReporte string

const (
	ReporteVentas        TipoReporte = "ventas"
	ReporteStock         TipoReporte = "stock"
	ReporteProductos     TipoReporte = "productos"
	ReporteClientes      TipoReporte = "clientes"
	ReporteFidelizacion  TipoReporte = "fidelizacion"
	ReporteFinanciero    TipoReporte = "financiero"
	ReporteOperacional   TipoReporte = "operacional"
	ReportePersonalizado TipoReporte = "personalizado"
)

type CategoriaReporte string

const (
	CategoriaComercial    CategoriaReporte = "comercial"
	CategoriaInventario   CategoriaReporte = "inventario"
	CategoriaFinanzas     CategoriaReporte = "finanzas"
	CategoriaOperaciones  CategoriaReporte = "operaciones"
	CategoriaMarketing    CategoriaReporte = "marketing"
	CategoriaRRHH         CategoriaReporte = "rrhh"
	CategoriaAuditoria    CategoriaReporte = "auditoria"
)

type EstadoReporte string

const (
	ReporteGenerando  EstadoReporte = "generando"
	ReporteCompletado EstadoReporte = "completado"
	ReporteError      EstadoReporte = "error"
	ReporteCancelado  EstadoReporte = "cancelado"
	ReporteExpirado   EstadoReporte = "expirado"
)

type FormatoReporte string

const (
	FormatoPDF   FormatoReporte = "pdf"
	FormatoExcel FormatoReporte = "excel"
	FormatoCSV   FormatoReporte = "csv"
	FormatoJSON  FormatoReporte = "json"
	FormatoHTML  FormatoReporte = "html"
)

type TipoParametro string

const (
	ParametroTexto        TipoParametro = "texto"
	ParametroNumero       TipoParametro = "numero"
	ParametroFecha        TipoParametro = "fecha"
	ParametroRangoFecha   TipoParametro = "rango_fecha"
	ParametroBooleano     TipoParametro = "booleano"
	ParametroSeleccion    TipoParametro = "seleccion"
	ParametroMultiseleccion TipoParametro = "multiseleccion"
)

type TipoWidget string

const (
	WidgetGrafico     TipoWidget = "grafico"
	WidgetTabla       TipoWidget = "tabla"
	WidgetMetrica     TipoWidget = "metrica"
	WidgetIndicador   TipoWidget = "indicador"
	WidgetMapa        TipoWidget = "mapa"
	WidgetTexto       TipoWidget = "texto"
	WidgetPersonalizado TipoWidget = "personalizado"
)

// Estructuras JSON

type ConfiguracionConsulta struct {
	ConsultaSQL             string                 `json:"consulta_sql"`
	TablasInvolucradas      []string               `json:"tablas_involucradas"`
	CamposSeleccionados     []string               `json:"campos_seleccionados"`
	FiltrosDefault          map[string]interface{} `json:"filtros_default,omitempty"`
	OrdenamientoDefault     []string               `json:"ordenamiento_default,omitempty"`
	LimiteRegistros         *int                   `json:"limite_registros,omitempty"`
	RequiereOptimizacion    bool                   `json:"requiere_optimizacion"`
	TimeoutSegundos         int                    `json:"timeout_segundos"`
	UsarCache               bool                   `json:"usar_cache"`
	TTLCacheMinutos         int                    `json:"ttl_cache_minutos"`
	ConsultasAdicionales    map[string]string      `json:"consultas_adicionales,omitempty"`
}

type ConfiguracionFormato struct {
	TituloReporte           string                 `json:"titulo_reporte"`
	SubtituloReporte        *string                `json:"subtitulo_reporte,omitempty"`
	MostrarLogo             bool                   `json:"mostrar_logo"`
	MostrarFechaGeneracion  bool                   `json:"mostrar_fecha_generacion"`
	MostrarParametros       bool                   `json:"mostrar_parametros"`
	MostrarTotalRegistros   bool                   `json:"mostrar_total_registros"`
	ConfiguracionColumnas   []ConfiguracionColumna `json:"configuracion_columnas"`
	ConfiguracionEstilos    ConfiguracionEstilos   `json:"configuracion_estilos"`
	ConfiguracionGraficos   *ConfiguracionGraficos `json:"configuracion_graficos,omitempty"`
	PiePagina               *string                `json:"pie_pagina,omitempty"`
	ConfiguracionPersonalizada map[string]interface{} `json:"configuracion_personalizada,omitempty"`
}

type ConfiguracionColumna struct {
	NombreCampo             string                 `json:"nombre_campo"`
	TituloColumna           string                 `json:"titulo_columna"`
	TipoDato                string                 `json:"tipo_dato"`
	Ancho                   *int                   `json:"ancho,omitempty"`
	Alineacion              string                 `json:"alineacion"` // "izquierda", "centro", "derecha"
	FormatoNumero           *string                `json:"formato_numero,omitempty"`
	FormatoFecha            *string                `json:"formato_fecha,omitempty"`
	MostrarTotales          bool                   `json:"mostrar_totales"`
	FuncionAgregacion       *string                `json:"funcion_agregacion,omitempty"` // "sum", "avg", "count", etc.
	Visible                 bool                   `json:"visible"`
	OrdenVisualizacion      int                    `json:"orden_visualizacion"`
	ConfiguracionCondicional *ConfiguracionCondicional `json:"configuracion_condicional,omitempty"`
}

type ConfiguracionCondicional struct {
	Condiciones             []CondicionFormato     `json:"condiciones"`
}

type CondicionFormato struct {
	Campo                   string                 `json:"campo"`
	Operador                string                 `json:"operador"` // "=", ">", "<", ">=", "<=", "!=", "contains"
	Valor                   interface{}            `json:"valor"`
	EstiloAplicar           EstiloCondicional      `json:"estilo_aplicar"`
}

type EstiloCondicional struct {
	ColorTexto              *string                `json:"color_texto,omitempty"`
	ColorFondo              *string                `json:"color_fondo,omitempty"`
	Negrita                 bool                   `json:"negrita"`
	Cursiva                 bool                   `json:"cursiva"`
	Subrayado               bool                   `json:"subrayado"`
}

type ConfiguracionEstilos struct {
	FuentePrincipal         string                 `json:"fuente_principal"`
	TamañoFuenteTitulo      int                    `json:"tamaño_fuente_titulo"`
	TamañoFuenteSubtitulo   int                    `json:"tamaño_fuente_subtitulo"`
	TamañoFuenteTexto       int                    `json:"tamaño_fuente_texto"`
	ColorTitulo             string                 `json:"color_titulo"`
	ColorTexto              string                 `json:"color_texto"`
	ColorFondo              string                 `json:"color_fondo"`
	ColorBordes             string                 `json:"color_bordes"`
	EstiloTabla             string                 `json:"estilo_tabla"` // "simple", "rayado", "bordeado"
	MargenesPersonalizados  *MargenesReporte       `json:"margenes_personalizados,omitempty"`
}

type MargenesReporte struct {
	Superior                float64                `json:"superior"`
	Inferior                float64                `json:"inferior"`
	Izquierdo               float64                `json:"izquierdo"`
	Derecho                 float64                `json:"derecho"`
}

type ConfiguracionGraficos struct {
	TipoGrafico             string                 `json:"tipo_grafico"` // "barras", "lineas", "pie", "area"
	CampoX                  string                 `json:"campo_x"`
	CampoY                  string                 `json:"campo_y"`
	CampoSerie              *string                `json:"campo_serie,omitempty"`
	TituloGrafico           string                 `json:"titulo_grafico"`
	TituloEjeX              string                 `json:"titulo_eje_x"`
	TituloEjeY              string                 `json:"titulo_eje_y"`
	MostrarLeyenda          bool                   `json:"mostrar_leyenda"`
	MostrarValores          bool                   `json:"mostrar_valores"`
	ColoresPersonalizados   []string               `json:"colores_personalizados,omitempty"`
	ConfiguracionAvanzada   map[string]interface{} `json:"configuracion_avanzada,omitempty"`
}

type ParametroReporte struct {
	Nombre                  string                 `json:"nombre"`
	Etiqueta                string                 `json:"etiqueta"`
	Descripcion             *string                `json:"descripcion,omitempty"`
	TipoParametro           TipoParametro          `json:"tipo_parametro"`
	ValorDefault            *interface{}           `json:"valor_default,omitempty"`
	Obligatorio             bool                   `json:"obligatorio"`
	OpcionesSeleccion       []OpcionSeleccion      `json:"opciones_seleccion,omitempty"`
	ValidacionParametro     *ValidacionParametro   `json:"validacion_parametro,omitempty"`
	DependeDe               *string                `json:"depende_de,omitempty"`
	ConfiguracionEspecial   map[string]interface{} `json:"configuracion_especial,omitempty"`
}

type OpcionSeleccion struct {
	Valor                   interface{}            `json:"valor"`
	Etiqueta                string                 `json:"etiqueta"`
	Descripcion             *string                `json:"descripcion,omitempty"`
	Activo                  bool                   `json:"activo"`
}

type ValidacionParametro struct {
	ValorMinimo             *interface{}           `json:"valor_minimo,omitempty"`
	ValorMaximo             *interface{}           `json:"valor_maximo,omitempty"`
	LongitudMinima          *int                   `json:"longitud_minima,omitempty"`
	LongitudMaxima          *int                   `json:"longitud_maxima,omitempty"`
	PatronRegex             *string                `json:"patron_regex,omitempty"`
	MensajeError            *string                `json:"mensaje_error,omitempty"`
}

type ParametrosReporte struct {
	Parametros              map[string]interface{} `json:"parametros"`
	FechaGeneracion         time.Time              `json:"fecha_generacion"`
	UsuarioGeneracion       uuid.UUID              `json:"usuario_generacion"`
	SucursalGeneracion      *uuid.UUID             `json:"sucursal_generacion,omitempty"`
	MetadatosParametros     map[string]interface{} `json:"metadatos_parametros,omitempty"`
}

type ConfiguracionCacheReporte struct {
	HabilitarCache          bool                   `json:"habilitar_cache"`
	TTLMinutos              int                    `json:"ttl_minutos"`
	CacheCompartido         bool                   `json:"cache_compartido"`
	ClavesInvalidacion      []string               `json:"claves_invalidacion"`
	TamañoMaximoMB          int                    `json:"tamaño_maximo_mb"`
	CompressionCache        bool                   `json:"compression_cache"`
}

type MetadatosPlantilla struct {
	Autor                   string                 `json:"autor"`
	FechaCreacion           time.Time              `json:"fecha_creacion"`
	UltimaModificacion      time.Time              `json:"ultima_modificacion"`
	VersionMinimaSistema    string                 `json:"version_minima_sistema"`
	Dependencias            []string               `json:"dependencias,omitempty"`
	Etiquetas               []string               `json:"etiquetas,omitempty"`
	Documentacion           *string                `json:"documentacion,omitempty"`
	EjemplosUso             []string               `json:"ejemplos_uso,omitempty"`
	MetadatosPersonalizados map[string]interface{} `json:"metadatos_personalizados,omitempty"`
}

type MetadatosGeneracionReporte struct {
	VersionPlantilla        int                    `json:"version_plantilla"`
	VersionSistema          string                 `json:"version_sistema"`
	ServidorGeneracion      string                 `json:"servidor_generacion"`
	TiempoConsulta          int                    `json:"tiempo_consulta_ms"`
	TiempoFormato           int                    `json:"tiempo_formato_ms"`
	MemoriaUtilizada        int                    `json:"memoria_utilizada_mb"`
	CacheUtilizado          bool                   `json:"cache_utilizado"`
	OptimizacionesAplicadas []string               `json:"optimizaciones_aplicadas,omitempty"`
	AdvertenciasGeneracion  []string               `json:"advertencias_generacion,omitempty"`
	EstadisticasConsulta    map[string]interface{} `json:"estadisticas_consulta,omitempty"`
}

type ConfiguracionReporteUsada struct {
	ConfiguracionConsulta   ConfiguracionConsulta  `json:"configuracion_consulta"`
	ConfiguracionFormato    ConfiguracionFormato   `json:"configuracion_formato"`
	ParametrosUsados        ParametrosReporte      `json:"parametros_usados"`
	ConfiguracionCache      *ConfiguracionCacheReporte `json:"configuracion_cache,omitempty"`
	FechaConfiguracion      time.Time              `json:"fecha_configuracion"`
}

type ConfiguracionCron struct {
	ExpresionCron           string                 `json:"expresion_cron"`
	ZonaHoraria             string                 `json:"zona_horaria"`
	FechaInicio             *time.Time             `json:"fecha_inicio,omitempty"`
	FechaFin                *time.Time             `json:"fecha_fin,omitempty"`
	MaxEjecuciones          *int                   `json:"max_ejecuciones,omitempty"`
	ReintentarEnError       bool                   `json:"reintentar_en_error"`
	MaxReintentos           int                    `json:"max_reintentos"`
	IntervaloReintentoMin   int                    `json:"intervalo_reintento_min"`
}

type ConfiguracionNotificacionReporte struct {
	NotificarCompletado     bool                   `json:"notificar_completado"`
	NotificarError          bool                   `json:"notificar_error"`
	EmailsNotificacion      []string               `json:"emails_notificacion"`
	IncluirArchivoEmail     bool                   `json:"incluir_archivo_email"`
	PlantillaEmailCompletado *string               `json:"plantilla_email_completado,omitempty"`
	PlantillaEmailError     *string                `json:"plantilla_email_error,omitempty"`
	WebhooksNotificacion    []string               `json:"webhooks_notificacion,omitempty"`
	ConfiguracionPersonalizada map[string]interface{} `json:"configuracion_personalizada,omitempty"`
}

type ConfiguracionRetencion struct {
	RetenerArchivos         bool                   `json:"retener_archivos"`
	DiasRetencion           int                    `json:"dias_retencion"`
	MaxArchivosRetenidos    int                    `json:"max_archivos_retenidos"`
	ComprimirArchivos       bool                   `json:"comprimir_archivos"`
	AlmacenamientoExterno   *string                `json:"almacenamiento_externo,omitempty"`
	ConfiguracionAlmacenamiento map[string]interface{} `json:"configuracion_almacenamiento,omitempty"`
}

type MetadatosProgramacion struct {
	TotalEjecucionesProgramadas int                `json:"total_ejecuciones_programadas"`
	PromedioTiempoEjecucion     float64            `json:"promedio_tiempo_ejecucion_ms"`
	TasaExito                   float64            `json:"tasa_exito_porcentaje"`
	UltimaEjecucionExitosa      *time.Time         `json:"ultima_ejecucion_exitosa,omitempty"`
	PatronErrores               []string           `json:"patron_errores,omitempty"`
	EstadisticasRendimiento     map[string]interface{} `json:"estadisticas_rendimiento,omitempty"`
}

type ConfiguracionLayoutDashboard struct {
	TipoLayout              string                 `json:"tipo_layout"` // "grid", "flex", "custom"
	Columnas                int                    `json:"columnas"`
	EspaciadoWidgets        int                    `json:"espaciado_widgets"`
	AlturaFilaDefault       int                    `json:"altura_fila_default"`
	ResponsiveBreakpoints   map[string]int         `json:"responsive_breakpoints"`
	ConfiguracionTema       ConfiguracionTemaDashboard `json:"configuracion_tema"`
	ConfiguracionPersonalizada map[string]interface{} `json:"configuracion_personalizada,omitempty"`
}

type ConfiguracionTemaDashboard struct {
	ColorPrimario           string                 `json:"color_primario"`
	ColorSecundario         string                 `json:"color_secundario"`
	ColorFondo              string                 `json:"color_fondo"`
	ColorTexto              string                 `json:"color_texto"`
	FuentePrincipal         string                 `json:"fuente_principal"`
	TamañoFuenteBase        int                    `json:"tamaño_fuente_base"`
	BorderRadius            int                    `json:"border_radius"`
	Sombras                 bool                   `json:"sombras"`
}

type WidgetDashboard struct {
	ID                      string                 `json:"id"`
	Nombre                  string                 `json:"nombre"`
	TipoWidget              TipoWidget             `json:"tipo_widget"`
	PlantillaReporteID      *uuid.UUID             `json:"plantilla_reporte_id,omitempty"`
	ConfiguracionWidget     ConfiguracionWidget    `json:"configuracion_widget"`
	PosicionLayout          PosicionLayoutWidget   `json:"posicion_layout"`
	ConfiguracionRefresh    ConfiguracionRefreshWidget `json:"configuracion_refresh"`
	Visible                 bool                   `json:"visible"`
	ConfiguracionPersonalizada map[string]interface{} `json:"configuracion_personalizada,omitempty"`
}

type ConfiguracionWidget struct {
	TituloWidget            string                 `json:"titulo_widget"`
	MostrarTitulo           bool                   `json:"mostrar_titulo"`
	ConfiguracionVisualizacion ConfiguracionVisualizacion `json:"configuracion_visualizacion"`
	FiltrosWidget           map[string]interface{} `json:"filtros_widget,omitempty"`
	ParametrosWidget        map[string]interface{} `json:"parametros_widget,omitempty"`
	ConfiguracionInteraccion ConfiguracionInteraccion `json:"configuracion_interaccion"`
}

type ConfiguracionVisualizacion struct {
	TipoVisualizacion       string                 `json:"tipo_visualizacion"`
	ConfiguracionGrafico    *ConfiguracionGraficos `json:"configuracion_grafico,omitempty"`
	ConfiguracionTabla      *ConfiguracionTablaWidget `json:"configuracion_tabla,omitempty"`
	ConfiguracionMetrica    *ConfiguracionMetricaWidget `json:"configuracion_metrica,omitempty"`
	ColoresPersonalizados   []string               `json:"colores_personalizados,omitempty"`
	ConfiguracionAvanzada   map[string]interface{} `json:"configuracion_avanzada,omitempty"`
}

type ConfiguracionTablaWidget struct {
	MostrarEncabezados      bool                   `json:"mostrar_encabezados"`
	FilasPorPagina          int                    `json:"filas_por_pagina"`
	HabilitarPaginacion     bool                   `json:"habilitar_paginacion"`
	HabilitarOrdenamiento   bool                   `json:"habilitar_ordenamiento"`
	HabilitarFiltros        bool                   `json:"habilitar_filtros"`
	EstiloTabla             string                 `json:"estilo_tabla"`
}

type ConfiguracionMetricaWidget struct {
	FormatoNumero           string                 `json:"formato_numero"`
	MostrarTendencia        bool                   `json:"mostrar_tendencia"`
	MostrarComparacion      bool                   `json:"mostrar_comparacion"`
	PeriodoComparacion      string                 `json:"periodo_comparacion"`
	ColorMetrica            string                 `json:"color_metrica"`
	IconoMetrica            *string                `json:"icono_metrica,omitempty"`
}

type ConfiguracionInteraccion struct {
	HabilitarDrillDown      bool                   `json:"habilitar_drill_down"`
	HabilitarTooltips       bool                   `json:"habilitar_tooltips"`
	HabilitarZoom           bool                   `json:"habilitar_zoom"`
	AccionesPersonalizadas  []AccionPersonalizada  `json:"acciones_personalizadas,omitempty"`
}

type AccionPersonalizada struct {
	Nombre                  string                 `json:"nombre"`
	Tipo                    string                 `json:"tipo"` // "link", "modal", "export", "custom"
	Configuracion           map[string]interface{} `json:"configuracion"`
}

type PosicionLayoutWidget struct {
	Fila                    int                    `json:"fila"`
	Columna                 int                    `json:"columna"`
	Ancho                   int                    `json:"ancho"` // en columnas
	Alto                    int                    `json:"alto"`  // en filas
	ZIndex                  int                    `json:"z_index"`
}

type ConfiguracionRefreshWidget struct {
	AutoRefresh             bool                   `json:"auto_refresh"`
	IntervaloRefreshSegundos int                   `json:"intervalo_refresh_segundos"`
	RefreshEnFoco           bool                   `json:"refresh_en_foco"`
	MostrarIndicadorCarga   bool                   `json:"mostrar_indicador_carga"`
}

type ConfiguracionRefreshDashboard struct {
	AutoRefresh             bool                   `json:"auto_refresh"`
	IntervaloRefreshSegundos int                   `json:"intervalo_refresh_segundos"`
	RefreshEnFoco           bool                   `json:"refresh_en_foco"`
	RefreshParcial          bool                   `json:"refresh_parcial"`
	WidgetsExcluidos        []string               `json:"widgets_excluidos,omitempty"`
}

type MetadatosDashboard struct {
	TotalWidgets            int                    `json:"total_widgets"`
	TiposWidgetsUsados      []string               `json:"tipos_widgets_usados"`
	UltimaModificacionLayout time.Time             `json:"ultima_modificacion_layout"`
	TiempoPromedioRenderizado float64              `json:"tiempo_promedio_renderizado_ms"`
	EstadisticasUso         map[string]interface{} `json:"estadisticas_uso,omitempty"`
	ConfiguracionOptimizacion map[string]interface{} `json:"configuracion_optimizacion,omitempty"`
}

// Implementar driver.Valuer para tipos JSON personalizados
func (c ConfiguracionConsulta) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c *ConfiguracionConsulta) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, c)
}

func (c ConfiguracionFormato) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c *ConfiguracionFormato) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, c)
}

func (p []ParametroReporte) Value() (driver.Value, error) {
	return json.Marshal(p)
}

func (p *[]ParametroReporte) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, p)
}

func (r []RolUsuario) Value() (driver.Value, error) {
	return json.Marshal(r)
}

func (r *[]RolUsuario) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, r)
}

func (c ConfiguracionCacheReporte) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c *ConfiguracionCacheReporte) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, c)
}

func (m MetadatosPlantilla) Value() (driver.Value, error) {
	return json.Marshal(m)
}

func (m *MetadatosPlantilla) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, m)
}

func (p ParametrosReporte) Value() (driver.Value, error) {
	return json.Marshal(p)
}

func (p *ParametrosReporte) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, p)
}

func (m MetadatosGeneracionReporte) Value() (driver.Value, error) {
	return json.Marshal(m)
}

func (m *MetadatosGeneracionReporte) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, m)
}

func (c ConfiguracionReporteUsada) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c *ConfiguracionReporteUsada) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, c)
}

func (c ConfiguracionCron) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c *ConfiguracionCron) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, c)
}

func (c ConfiguracionNotificacionReporte) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c *ConfiguracionNotificacionReporte) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, c)
}

func (c ConfiguracionRetencion) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c *ConfiguracionRetencion) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, c)
}

func (m MetadatosProgramacion) Value() (driver.Value, error) {
	return json.Marshal(m)
}

func (m *MetadatosProgramacion) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, m)
}

func (c ConfiguracionLayoutDashboard) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c *ConfiguracionLayoutDashboard) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, c)
}

func (w []WidgetDashboard) Value() (driver.Value, error) {
	return json.Marshal(w)
}

func (w *[]WidgetDashboard) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, w)
}

func (c ConfiguracionRefreshDashboard) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c *ConfiguracionRefreshDashboard) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, c)
}

func (m MetadatosDashboard) Value() (driver.Value, error) {
	return json.Marshal(m)
}

func (m *MetadatosDashboard) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, m)
}

// DTOs para PlantillaReporte

type PlantillaReporteCreateDTO struct {
	Nombre                string                    `json:"nombre" binding:"required" validate:"required,max=255"`
	Descripcion           *string                   `json:"descripcion,omitempty"`
	TipoReporte           TipoReporte               `json:"tipo_reporte" binding:"required"`
	CategoriaReporte      CategoriaReporte          `json:"categoria_reporte" binding:"required"`
	ConfiguracionConsulta ConfiguracionConsulta     `json:"configuracion_consulta" binding:"required"`
	ConfiguracionFormato  ConfiguracionFormato      `json:"configuracion_formato" binding:"required"`
	ParametrosRequeridos  []ParametroReporte        `json:"parametros_requeridos,omitempty"`
	ParametrosOpcionales  []ParametroReporte        `json:"parametros_opcionales,omitempty"`
	EsPublica             bool                      `json:"es_publica"`
	RolesPermitidos       []RolUsuario              `json:"roles_permitidos,omitempty"`
	SucursalesPermitidas  []uuid.UUID               `json:"sucursales_permitidas,omitempty"`
	ConfiguracionCache    *ConfiguracionCacheReporte `json:"configuracion_cache,omitempty"`
	MetadatosPlantilla    *MetadatosPlantilla       `json:"metadatos_plantilla,omitempty"`
}

type PlantillaReporteUpdateDTO struct {
	Nombre                *string                   `json:"nombre,omitempty" validate:"omitempty,max=255"`
	Descripcion           *string                   `json:"descripcion,omitempty"`
	ConfiguracionConsulta *ConfiguracionConsulta    `json:"configuracion_consulta,omitempty"`
	ConfiguracionFormato  *ConfiguracionFormato     `json:"configuracion_formato,omitempty"`
	ParametrosRequeridos  *[]ParametroReporte       `json:"parametros_requeridos,omitempty"`
	ParametrosOpcionales  *[]ParametroReporte       `json:"parametros_opcionales,omitempty"`
	Activa                *bool                     `json:"activa,omitempty"`
	EsPublica             *bool                     `json:"es_publica,omitempty"`
	RolesPermitidos       *[]RolUsuario             `json:"roles_permitidos,omitempty"`
	SucursalesPermitidas  *[]uuid.UUID              `json:"sucursales_permitidas,omitempty"`
	ConfiguracionCache    *ConfiguracionCacheReporte `json:"configuracion_cache,omitempty"`
	MetadatosPlantilla    *MetadatosPlantilla       `json:"metadatos_plantilla,omitempty"`
}

type PlantillaReporteResponseDTO struct {
	ID                      uuid.UUID                 `json:"id"`
	Nombre                  string                    `json:"nombre"`
	Descripcion             *string                   `json:"descripcion,omitempty"`
	TipoReporte             TipoReporte               `json:"tipo_reporte"`
	CategoriaReporte        CategoriaReporte          `json:"categoria_reporte"`
	ConfiguracionConsulta   ConfiguracionConsulta     `json:"configuracion_consulta"`
	ConfiguracionFormato    ConfiguracionFormato      `json:"configuracion_formato"`
	ParametrosRequeridos    []ParametroReporte        `json:"parametros_requeridos,omitempty"`
	ParametrosOpcionales    []ParametroReporte        `json:"parametros_opcionales,omitempty"`
	Activa                  bool                      `json:"activa"`
	EsPublica               bool                      `json:"es_publica"`
	RolesPermitidos         []RolUsuario              `json:"roles_permitidos,omitempty"`
	SucursalesPermitidas    []uuid.UUID               `json:"sucursales_permitidas,omitempty"`
	TotalGeneraciones       int                       `json:"total_generaciones"`
	FechaUltimoUso          *time.Time                `json:"fecha_ultimo_uso,omitempty"`
	TiempoPromedioEjecucion *int                      `json:"tiempo_promedio_ejecucion_ms,omitempty"`
	ConfiguracionCache      *ConfiguracionCacheReporte `json:"configuracion_cache,omitempty"`
	VersionPlantilla        int                       `json:"version_plantilla"`
	MetadatosPlantilla      *MetadatosPlantilla       `json:"metadatos_plantilla,omitempty"`
	FechaCreacion           time.Time                 `json:"fecha_creacion"`
	FechaModificacion       time.Time                 `json:"fecha_modificacion"`
}

type PlantillaReporteListDTO struct {
	ID                      uuid.UUID        `json:"id"`
	Nombre                  string           `json:"nombre"`
	TipoReporte             TipoReporte      `json:"tipo_reporte"`
	CategoriaReporte        CategoriaReporte `json:"categoria_reporte"`
	Activa                  bool             `json:"activa"`
	EsPublica               bool             `json:"es_publica"`
	TotalGeneraciones       int              `json:"total_generaciones"`
	FechaUltimoUso          *time.Time       `json:"fecha_ultimo_uso,omitempty"`
	TiempoPromedioEjecucion *int             `json:"tiempo_promedio_ejecucion_ms,omitempty"`
	VersionPlantilla        int              `json:"version_plantilla"`
	FechaCreacion           time.Time        `json:"fecha_creacion"`
	FechaModificacion       time.Time        `json:"fecha_modificacion"`
}

// DTOs para ReporteGenerado

type ReporteGeneradoCreateDTO struct {
	PlantillaID      uuid.UUID         `json:"plantilla_id" binding:"required"`
	SucursalID       *uuid.UUID        `json:"sucursal_id,omitempty"`
	Nombre           string            `json:"nombre" binding:"required"`
	Descripcion      *string           `json:"descripcion,omitempty"`
	ParametrosUsados ParametrosReporte `json:"parametros_usados" binding:"required"`
	FormatoSalida    FormatoReporte    `json:"formato_salida" binding:"required"`
	EsPublico        bool              `json:"es_publico"`
	CompartidoCon    []uuid.UUID       `json:"compartido_con,omitempty"`
}

type ReporteGeneradoResponseDTO struct {
	ID                    uuid.UUID                   `json:"id"`
	PlantillaID           uuid.UUID                   `json:"plantilla_id"`
	SucursalID            *uuid.UUID                  `json:"sucursal_id,omitempty"`
	UsuarioID             uuid.UUID                   `json:"usuario_id"`
	Nombre                string                      `json:"nombre"`
	Descripcion           *string                     `json:"descripcion,omitempty"`
	ParametrosUsados      ParametrosReporte           `json:"parametros_usados"`
	Estado                EstadoReporte               `json:"estado"`
	FechaGeneracion       time.Time                   `json:"fecha_generacion"`
	FechaCompletado       *time.Time                  `json:"fecha_completado,omitempty"`
	TiempoEjecucion       *int                        `json:"tiempo_ejecucion_ms,omitempty"`
	TotalRegistros        *int                        `json:"total_registros,omitempty"`
	TamañoArchivo         *int                        `json:"tamaño_archivo_bytes,omitempty"`
	FormatoSalida         FormatoReporte              `json:"formato_salida"`
	ArchivoGenerado       *string                     `json:"archivo_generado,omitempty"`
	URLDescarga           *string                     `json:"url_descarga,omitempty"`
	FechaExpiracion       *time.Time                  `json:"fecha_expiracion,omitempty"`
	ErrorDetalle          *string                     `json:"error_detalle,omitempty"`
	CodigoError           *string                     `json:"codigo_error,omitempty"`
	MetadatosGeneracion   *MetadatosGeneracionReporte `json:"metadatos_generacion,omitempty"`
	ConfiguracionUsada    *ConfiguracionReporteUsada  `json:"configuracion_usada,omitempty"`
	EsPublico             bool                        `json:"es_publico"`
	CompartidoCon         []uuid.UUID                 `json:"compartido_con,omitempty"`
	TotalDescargas        int                         `json:"total_descargas"`
	FechaCreacion         time.Time                   `json:"fecha_creacion"`
	PlantillaNombre       *string                     `json:"plantilla_nombre,omitempty"`
	SucursalNombre        *string                     `json:"sucursal_nombre,omitempty"`
	UsuarioNombre         *string                     `json:"usuario_nombre,omitempty"`
}

type ReporteGeneradoListDTO struct {
	ID                uuid.UUID      `json:"id"`
	PlantillaID       uuid.UUID      `json:"plantilla_id"`
	SucursalID        *uuid.UUID     `json:"sucursal_id,omitempty"`
	Nombre            string         `json:"nombre"`
	Estado            EstadoReporte  `json:"estado"`
	FechaGeneracion   time.Time      `json:"fecha_generacion"`
	FechaCompletado   *time.Time     `json:"fecha_completado,omitempty"`
	TiempoEjecucion   *int           `json:"tiempo_ejecucion_ms,omitempty"`
	TotalRegistros    *int           `json:"total_registros,omitempty"`
	FormatoSalida     FormatoReporte `json:"formato_salida"`
	TotalDescargas    int            `json:"total_descargas"`
	PlantillaNombre   *string        `json:"plantilla_nombre,omitempty"`
	SucursalNombre    *string        `json:"sucursal_nombre,omitempty"`
}

// Filtros para búsqueda

type PlantillaReporteFilter struct {
	PaginationFilter
	SortFilter
	Nombre           *string           `json:"nombre,omitempty" form:"nombre"`
	TipoReporte      *TipoReporte      `json:"tipo_reporte,omitempty" form:"tipo_reporte"`
	CategoriaReporte *CategoriaReporte `json:"categoria_reporte,omitempty" form:"categoria_reporte"`
	Activa           *bool             `json:"activa,omitempty" form:"activa"`
	EsPublica        *bool             `json:"es_publica,omitempty" form:"es_publica"`
	SucursalID       *uuid.UUID        `json:"sucursal_id,omitempty" form:"sucursal_id"`
	RolUsuario       *RolUsuario       `json:"rol_usuario,omitempty" form:"rol_usuario"`
}

type ReporteGeneradoFilter struct {
	PaginationFilter
	SortFilter
	DateRangeFilter
	PlantillaID   *uuid.UUID     `json:"plantilla_id,omitempty" form:"plantilla_id"`
	SucursalID    *uuid.UUID     `json:"sucursal_id,omitempty" form:"sucursal_id"`
	UsuarioID     *uuid.UUID     `json:"usuario_id,omitempty" form:"usuario_id"`
	Estado        *EstadoReporte `json:"estado,omitempty" form:"estado"`
	FormatoSalida *FormatoReporte `json:"formato_salida,omitempty" form:"formato_salida"`
	EsPublico     *bool          `json:"es_publico,omitempty" form:"es_publico"`
	Nombre        *string        `json:"nombre,omitempty" form:"nombre"`
}

// Métodos helper

func (p *PlantillaReporte) PuedeUsarUsuario(usuarioRol RolUsuario, sucursalID *uuid.UUID) bool {
	// Verificar roles permitidos
	if len(p.RolesPermitidos) > 0 {
		rolPermitido := false
		for _, rol := range p.RolesPermitidos {
			if rol == usuarioRol {
				rolPermitido = true
				break
			}
		}
		if !rolPermitido {
			return false
		}
	}
	
	// Verificar sucursales permitidas
	if len(p.SucursalesPermitidas) > 0 && sucursalID != nil {
		sucursalPermitida := false
		for _, id := range p.SucursalesPermitidas {
			if id == *sucursalID {
				sucursalPermitida = true
				break
			}
		}
		if !sucursalPermitida {
			return false
		}
	}
	
	return p.Activa
}

func (p *PlantillaReporte) ActualizarUso(tiempoEjecucion int) {
	p.TotalGeneraciones++
	now := time.Now()
	p.FechaUltimoUso = &now
	
	// Calcular tiempo promedio de ejecución
	if p.TiempoPromedioEjecucion == nil {
		p.TiempoPromedioEjecucion = &tiempoEjecucion
	} else {
		// Promedio móvil simple
		nuevoPromedio := (*p.TiempoPromedioEjecucion + tiempoEjecucion) / 2
		p.TiempoPromedioEjecucion = &nuevoPromedio
	}
}

func (r *ReporteGenerado) MarcarComoCompletado(archivoGenerado string, totalRegistros int, tamañoArchivo int) {
	r.Estado = ReporteCompletado
	now := time.Now()
	r.FechaCompletado = &now
	r.ArchivoGenerado = &archivoGenerado
	r.TotalRegistros = &totalRegistros
	r.TamañoArchivo = &tamañoArchivo
	
	// Calcular tiempo de ejecución
	if r.FechaCompletado != nil {
		tiempoEjecucion := int(r.FechaCompletado.Sub(r.FechaGeneracion).Milliseconds())
		r.TiempoEjecucion = &tiempoEjecucion
	}
	
	// Establecer fecha de expiración (30 días por defecto)
	fechaExpiracion := now.AddDate(0, 0, 30)
	r.FechaExpiracion = &fechaExpiracion
}

func (r *ReporteGenerado) MarcarComoError(codigoError string, errorDetalle string) {
	r.Estado = ReporteError
	r.CodigoError = &codigoError
	r.ErrorDetalle = &errorDetalle
	
	now := time.Now()
	tiempoEjecucion := int(now.Sub(r.FechaGeneracion).Milliseconds())
	r.TiempoEjecucion = &tiempoEjecucion
}

func (r *ReporteGenerado) PuedeDescargar(usuarioID uuid.UUID) bool {
	if r.Estado != ReporteCompletado {
		return false
	}
	
	if r.FechaExpiracion != nil && r.FechaExpiracion.Before(time.Now()) {
		return false
	}
	
	if r.EsPublico {
		return true
	}
	
	if r.UsuarioID == usuarioID {
		return true
	}
	
	// Verificar si está compartido con el usuario
	for _, id := range r.CompartidoCon {
		if id == usuarioID {
			return true
		}
	}
	
	return false
}

func (r *ReporteGenerado) IncrementarDescargas() {
	r.TotalDescargas++
}

func (r *ReporteProgramado) CalcularProximaEjecucion() error {
	// En implementación real, usar librería de cron para calcular próxima ejecución
	// Por ahora, ejemplo simple
	if r.ConfiguracionCron.ExpresionCron == "0 0 9 * * *" { // Diario a las 9 AM
		proximaEjecucion := time.Now().AddDate(0, 0, 1)
		proximaEjecucion = time.Date(proximaEjecucion.Year(), proximaEjecucion.Month(), proximaEjecucion.Day(), 9, 0, 0, 0, proximaEjecucion.Location())
		r.ProximaEjecucion = &proximaEjecucion
	}
	
	return nil
}

func (r *ReporteProgramado) ActualizarEstadisticas(exitoso bool, tiempoEjecucion int) {
	r.TotalEjecuciones++
	now := time.Now()
	r.UltimaEjecucion = &now
	
	if exitoso {
		r.EjecucionesExitosas++
	} else {
		r.EjecucionesError++
	}
	
	// Actualizar metadatos de programación
	if r.MetadatosProgramacion == nil {
		r.MetadatosProgramacion = &MetadatosProgramacion{}
	}
	
	r.MetadatosProgramacion.TotalEjecucionesProgramadas = r.TotalEjecuciones
	
	// Calcular promedio de tiempo de ejecución
	if r.MetadatosProgramacion.PromedioTiempoEjecucion == 0 {
		r.MetadatosProgramacion.PromedioTiempoEjecucion = float64(tiempoEjecucion)
	} else {
		r.MetadatosProgramacion.PromedioTiempoEjecucion = (r.MetadatosProgramacion.PromedioTiempoEjecucion + float64(tiempoEjecucion)) / 2
	}
	
	// Calcular tasa de éxito
	if r.TotalEjecuciones > 0 {
		r.MetadatosProgramacion.TasaExito = float64(r.EjecucionesExitosas) / float64(r.TotalEjecuciones) * 100
	}
	
	if exitoso {
		r.MetadatosProgramacion.UltimaEjecucionExitosa = &now
	}
}

func (d *DashboardPersonalizado) ActualizarVisualizacion() {
	d.TotalVisualizaciones++
	now := time.Now()
	d.FechaUltimaVisualizacion = &now
}

func (d *DashboardPersonalizado) PuedeAcceder(usuarioID uuid.UUID) bool {
	if d.UsuarioID == usuarioID {
		return true
	}
	
	if d.EsPublico {
		return true
	}
	
	// Verificar si está compartido con el usuario
	for _, id := range d.CompartidoCon {
		if id == usuarioID {
			return true
		}
	}
	
	return false
}

func (p *PlantillaReporte) ToResponseDTO() PlantillaReporteResponseDTO {
	return PlantillaReporteResponseDTO{
		ID:                      p.ID,
		Nombre:                  p.Nombre,
		Descripcion:             p.Descripcion,
		TipoReporte:             p.TipoReporte,
		CategoriaReporte:        p.CategoriaReporte,
		ConfiguracionConsulta:   p.ConfiguracionConsulta,
		ConfiguracionFormato:    p.ConfiguracionFormato,
		ParametrosRequeridos:    p.ParametrosRequeridos,
		ParametrosOpcionales:    p.ParametrosOpcionales,
		Activa:                  p.Activa,
		EsPublica:               p.EsPublica,
		RolesPermitidos:         p.RolesPermitidos,
		SucursalesPermitidas:    p.SucursalesPermitidas,
		TotalGeneraciones:       p.TotalGeneraciones,
		FechaUltimoUso:          p.FechaUltimoUso,
		TiempoPromedioEjecucion: p.TiempoPromedioEjecucion,
		ConfiguracionCache:      p.ConfiguracionCache,
		VersionPlantilla:        p.VersionPlantilla,
		MetadatosPlantilla:      p.MetadatosPlantilla,
		FechaCreacion:           p.FechaCreacion,
		FechaModificacion:       p.FechaModificacion,
	}
}

func (p *PlantillaReporte) ToListDTO() PlantillaReporteListDTO {
	return PlantillaReporteListDTO{
		ID:                      p.ID,
		Nombre:                  p.Nombre,
		TipoReporte:             p.TipoReporte,
		CategoriaReporte:        p.CategoriaReporte,
		Activa:                  p.Activa,
		EsPublica:               p.EsPublica,
		TotalGeneraciones:       p.TotalGeneraciones,
		FechaUltimoUso:          p.FechaUltimoUso,
		TiempoPromedioEjecucion: p.TiempoPromedioEjecucion,
		VersionPlantilla:        p.VersionPlantilla,
		FechaCreacion:           p.FechaCreacion,
		FechaModificacion:       p.FechaModificacion,
	}
}

func (r *ReporteGenerado) ToResponseDTO() ReporteGeneradoResponseDTO {
	dto := ReporteGeneradoResponseDTO{
		ID:                  r.ID,
		PlantillaID:         r.PlantillaID,
		SucursalID:          r.SucursalID,
		UsuarioID:           r.UsuarioID,
		Nombre:              r.Nombre,
		Descripcion:         r.Descripcion,
		ParametrosUsados:    r.ParametrosUsados,
		Estado:              r.Estado,
		FechaGeneracion:     r.FechaGeneracion,
		FechaCompletado:     r.FechaCompletado,
		TiempoEjecucion:     r.TiempoEjecucion,
		TotalRegistros:      r.TotalRegistros,
		TamañoArchivo:       r.TamañoArchivo,
		FormatoSalida:       r.FormatoSalida,
		ArchivoGenerado:     r.ArchivoGenerado,
		URLDescarga:         r.URLDescarga,
		FechaExpiracion:     r.FechaExpiracion,
		ErrorDetalle:        r.ErrorDetalle,
		CodigoError:         r.CodigoError,
		MetadatosGeneracion: r.MetadatosGeneracion,
		ConfiguracionUsada:  r.ConfiguracionUsada,
		EsPublico:           r.EsPublico,
		CompartidoCon:       r.CompartidoCon,
		TotalDescargas:      r.TotalDescargas,
		FechaCreacion:       r.FechaCreacion,
	}
	
	if r.Plantilla != nil {
		dto.PlantillaNombre = &r.Plantilla.Nombre
	}
	
	if r.Sucursal != nil {
		dto.SucursalNombre = &r.Sucursal.Nombre
	}
	
	if r.Usuario != nil {
		nombreUsuario := r.Usuario.Nombre
		if r.Usuario.Apellido != nil {
			nombreUsuario += " " + *r.Usuario.Apellido
		}
		dto.UsuarioNombre = &nombreUsuario
	}
	
	return dto
}

func (r *ReporteGenerado) ToListDTO() ReporteGeneradoListDTO {
	dto := ReporteGeneradoListDTO{
		ID:              r.ID,
		PlantillaID:     r.PlantillaID,
		SucursalID:      r.SucursalID,
		Nombre:          r.Nombre,
		Estado:          r.Estado,
		FechaGeneracion: r.FechaGeneracion,
		FechaCompletado: r.FechaCompletado,
		TiempoEjecucion: r.TiempoEjecucion,
		TotalRegistros:  r.TotalRegistros,
		FormatoSalida:   r.FormatoSalida,
		TotalDescargas:  r.TotalDescargas,
	}
	
	if r.Plantilla != nil {
		dto.PlantillaNombre = &r.Plantilla.Nombre
	}
	
	if r.Sucursal != nil {
		dto.SucursalNombre = &r.Sucursal.Nombre
	}
	
	return dto
}

func (dto *PlantillaReporteCreateDTO) ToModel() *PlantillaReporte {
	return &PlantillaReporte{
		Nombre:                dto.Nombre,
		Descripcion:           dto.Descripcion,
		TipoReporte:           dto.TipoReporte,
		CategoriaReporte:      dto.CategoriaReporte,
		ConfiguracionConsulta: dto.ConfiguracionConsulta,
		ConfiguracionFormato:  dto.ConfiguracionFormato,
		ParametrosRequeridos:  dto.ParametrosRequeridos,
		ParametrosOpcionales:  dto.ParametrosOpcionales,
		Activa:                true,
		EsPublica:             dto.EsPublica,
		RolesPermitidos:       dto.RolesPermitidos,
		SucursalesPermitidas:  dto.SucursalesPermitidas,
		ConfiguracionCache:    dto.ConfiguracionCache,
		VersionPlantilla:      1,
		TotalGeneraciones:     0,
		MetadatosPlantilla:    dto.MetadatosPlantilla,
	}
}

// Validaciones personalizadas

func (p *PlantillaReporte) Validate() error {
	if p.VersionPlantilla <= 0 {
		return fmt.Errorf("versión de plantilla debe ser positiva")
	}
	
	if p.TotalGeneraciones < 0 {
		return fmt.Errorf("total de generaciones no puede ser negativo")
	}
	
	if p.ConfiguracionConsulta.TimeoutSegundos <= 0 {
		return fmt.Errorf("timeout de consulta debe ser positivo")
	}
	
	return nil
}

func (r *ReporteGenerado) Validate() error {
	if r.TotalRegistros != nil && *r.TotalRegistros < 0 {
		return fmt.Errorf("total de registros no puede ser negativo")
	}
	
	if r.TamañoArchivo != nil && *r.TamañoArchivo < 0 {
		return fmt.Errorf("tamaño de archivo no puede ser negativo")
	}
	
	if r.TotalDescargas < 0 {
		return fmt.Errorf("total de descargas no puede ser negativo")
	}
	
	if r.FechaExpiracion != nil && r.FechaExpiracion.Before(r.FechaGeneracion) {
		return fmt.Errorf("fecha de expiración no puede ser anterior a fecha de generación")
	}
	
	return nil
}

func (r *ReporteProgramado) Validate() error {
	if r.TotalEjecuciones < 0 {
		return fmt.Errorf("total de ejecuciones no puede ser negativo")
	}
	
	if r.EjecucionesExitosas < 0 {
		return fmt.Errorf("ejecuciones exitosas no puede ser negativo")
	}
	
	if r.EjecucionesError < 0 {
		return fmt.Errorf("ejecuciones con error no puede ser negativo")
	}
	
	if r.EjecucionesExitosas+r.EjecucionesError > r.TotalEjecuciones {
		return fmt.Errorf("suma de ejecuciones exitosas y con error no puede ser mayor al total")
	}
	
	return nil
}

func (d *DashboardPersonalizado) Validate() error {
	if d.TotalVisualizaciones < 0 {
		return fmt.Errorf("total de visualizaciones no puede ser negativo")
	}
	
	if d.VersionDashboard <= 0 {
		return fmt.Errorf("versión de dashboard debe ser positiva")
	}
	
	if len(d.Widgets) == 0 {
		return fmt.Errorf("dashboard debe tener al menos un widget")
	}
	
	return nil
}

// Funciones de utilidad

func GenerarHashReporte(parametros ParametrosReporte, configuracion ConfiguracionConsulta) string {
	// En implementación real, generar hash SHA256
	return fmt.Sprintf("hash_reporte_%d", time.Now().Unix())
}

func ValidarExpresionCron(expresion string) error {
	// En implementación real, validar expresión cron
	if expresion == "" {
		return fmt.Errorf("expresión cron no puede estar vacía")
	}
	return nil
}

func CalcularTamañoEstimadoReporte(totalRegistros int, formatoSalida FormatoReporte) int {
	// Estimación básica en bytes
	switch formatoSalida {
	case FormatoPDF:
		return totalRegistros * 100 // ~100 bytes por registro en PDF
	case FormatoExcel:
		return totalRegistros * 80  // ~80 bytes por registro en Excel
	case FormatoCSV:
		return totalRegistros * 50  // ~50 bytes por registro en CSV
	case FormatoJSON:
		return totalRegistros * 120 // ~120 bytes por registro en JSON
	case FormatoHTML:
		return totalRegistros * 150 // ~150 bytes por registro en HTML
	default:
		return totalRegistros * 100
	}
}

