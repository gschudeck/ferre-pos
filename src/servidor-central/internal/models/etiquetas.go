package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// PlantillaEtiqueta representa una plantilla para generar etiquetas
type PlantillaEtiqueta struct {
	BaseModel
	Nombre                string                    `json:"nombre" db:"nombre" binding:"required" validate:"required,max=255"`
	Descripcion           *string                   `json:"descripcion,omitempty" db:"descripcion"`
	TipoEtiqueta          TipoEtiqueta              `json:"tipo_etiqueta" db:"tipo_etiqueta" binding:"required"`
	Dimensiones           DimensionesEtiqueta       `json:"dimensiones" db:"dimensiones" binding:"required"`
	ConfiguracionDiseno   ConfiguracionDiseno       `json:"configuracion_diseno" db:"configuracion_diseno" binding:"required"`
	CamposPersonalizados  []CampoPersonalizado      `json:"campos_personalizados,omitempty" db:"campos_personalizados"`
	Activa                bool                      `json:"activa" db:"activa" default:"true"`
	EsDefault             bool                      `json:"es_default" db:"es_default" default:"false"`
	SucursalesPermitidas  []uuid.UUID               `json:"sucursales_permitidas,omitempty" db:"sucursales_permitidas"`
	CategoriasPermitidas  []uuid.UUID               `json:"categorias_permitidas,omitempty" db:"categorias_permitidas"`
	TotalGeneradas        int                       `json:"total_generadas" db:"total_generadas" default:"0"`
	FechaUltimoUso        *time.Time                `json:"fecha_ultimo_uso,omitempty" db:"fecha_ultimo_uso"`
	ConfiguracionImpresion *ConfiguracionImpresion  `json:"configuracion_impresion,omitempty" db:"configuracion_impresion"`
	VersionPlantilla      int                       `json:"version_plantilla" db:"version_plantilla" default:"1"`
	CacheRenderizado      *CacheRenderizado         `json:"cache_renderizado,omitempty" db:"cache_renderizado"`

	// Relaciones
	EtiquetasGeneradas []EtiquetaGenerada `json:"etiquetas_generadas,omitempty" gorm:"foreignKey:PlantillaID"`
}

// EtiquetaGenerada representa una etiqueta generada
type EtiquetaGenerada struct {
	BaseModel
	PlantillaID           uuid.UUID                 `json:"plantilla_id" db:"plantilla_id" binding:"required"`
	ProductoID            uuid.UUID                 `json:"producto_id" db:"producto_id" binding:"required"`
	SucursalID            uuid.UUID                 `json:"sucursal_id" db:"sucursal_id" binding:"required"`
	UsuarioID             uuid.UUID                 `json:"usuario_id" db:"usuario_id" binding:"required"`
	Cantidad              int                       `json:"cantidad" db:"cantidad" binding:"required,min=1"`
	Estado                EstadoEtiqueta            `json:"estado" db:"estado" default:"generada"`
	FechaGeneracion       time.Time                 `json:"fecha_generacion" db:"fecha_generacion" default:"NOW()"`
	FechaImpresion        *time.Time                `json:"fecha_impresion,omitempty" db:"fecha_impresion"`
	TerminalImpresion     *uuid.UUID                `json:"terminal_impresion,omitempty" db:"terminal_impresion"`
	DatosEtiqueta         DatosEtiqueta             `json:"datos_etiqueta" db:"datos_etiqueta" binding:"required"`
	ArchivoGenerado       *string                   `json:"archivo_generado,omitempty" db:"archivo_generado"`
	HashIntegridad        *string                   `json:"hash_integridad,omitempty" db:"hash_integridad"`
	ConfiguracionUsada    *ConfiguracionEtiquetaUsada `json:"configuracion_usada,omitempty" db:"configuracion_usada"`
	MetadatosGeneracion   *MetadatosGeneracion      `json:"metadatos_generacion,omitempty" db:"metadatos_generacion"`
	LoteGeneracion        *uuid.UUID                `json:"lote_generacion,omitempty" db:"lote_generacion"`
	TiempoGeneracion      *int                      `json:"tiempo_generacion_ms,omitempty" db:"tiempo_generacion_ms"`

	// Relaciones
	Plantilla *PlantillaEtiqueta `json:"plantilla,omitempty" gorm:"foreignKey:PlantillaID"`
	Producto  *Producto          `json:"producto,omitempty" gorm:"foreignKey:ProductoID"`
	Sucursal  *Sucursal          `json:"sucursal,omitempty" gorm:"foreignKey:SucursalID"`
	Usuario   *Usuario           `json:"usuario,omitempty" gorm:"foreignKey:UsuarioID"`
	Terminal  *Terminal          `json:"terminal,omitempty" gorm:"foreignKey:TerminalImpresion"`
}

// LoteEtiquetas representa un lote de etiquetas generadas
type LoteEtiquetas struct {
	BaseModel
	Nombre                string                    `json:"nombre" db:"nombre" binding:"required"`
	Descripcion           *string                   `json:"descripcion,omitempty" db:"descripcion"`
	SucursalID            uuid.UUID                 `json:"sucursal_id" db:"sucursal_id" binding:"required"`
	UsuarioID             uuid.UUID                 `json:"usuario_id" db:"usuario_id" binding:"required"`
	TotalEtiquetas        int                       `json:"total_etiquetas" db:"total_etiquetas" default:"0"`
	EtiquetasImpresas     int                       `json:"etiquetas_impresas" db:"etiquetas_impresas" default:"0"`
	Estado                EstadoLote                `json:"estado" db:"estado" default:"pendiente"`
	FechaGeneracion       time.Time                 `json:"fecha_generacion" db:"fecha_generacion" default:"NOW()"`
	FechaCompletado       *time.Time                `json:"fecha_completado,omitempty" db:"fecha_completado"`
	ConfiguracionLote     *ConfiguracionLote        `json:"configuracion_lote,omitempty" db:"configuracion_lote"`
	ArchivoConsolidado    *string                   `json:"archivo_consolidado,omitempty" db:"archivo_consolidado"`
	MetadatosLote         *MetadatosLote            `json:"metadatos_lote,omitempty" db:"metadatos_lote"`

	// Relaciones
	Sucursal           *Sucursal          `json:"sucursal,omitempty" gorm:"foreignKey:SucursalID"`
	Usuario            *Usuario           `json:"usuario,omitempty" gorm:"foreignKey:UsuarioID"`
	EtiquetasGeneradas []EtiquetaGenerada `json:"etiquetas_generadas,omitempty" gorm:"foreignKey:LoteGeneracion"`
}

// Enums para etiquetas

type TipoEtiqueta string

const (
	EtiquetaProducto    TipoEtiqueta = "producto"
	EtiquetaPrecio      TipoEtiqueta = "precio"
	EtiquetaPromocion   TipoEtiqueta = "promocion"
	EtiquetaInventario  TipoEtiqueta = "inventario"
	EtiquetaPersonalizada TipoEtiqueta = "personalizada"
)

type EstadoEtiqueta string

const (
	EtiquetaGenerada  EstadoEtiqueta = "generada"
	EtiquetaImpresa   EstadoEtiqueta = "impresa"
	EtiquetaCancelada EstadoEtiqueta = "cancelada"
	EtiquetaError     EstadoEtiqueta = "error"
)

type EstadoLote string

const (
	LotePendiente   EstadoLote = "pendiente"
	LoteGenerando   EstadoLote = "generando"
	LoteCompletado  EstadoLote = "completado"
	LoteCancelado   EstadoLote = "cancelado"
	LoteError       EstadoLote = "error"
)

// Estructuras JSON

type DimensionesEtiqueta struct {
	Ancho         float64 `json:"ancho"`          // en mm
	Alto          float64 `json:"alto"`           // en mm
	MargenSuperior float64 `json:"margen_superior"` // en mm
	MargenInferior float64 `json:"margen_inferior"` // en mm
	MargenIzquierdo float64 `json:"margen_izquierdo"` // en mm
	MargenDerecho  float64 `json:"margen_derecho"`  // en mm
	Orientacion    string  `json:"orientacion"`     // "vertical", "horizontal"
	DPI            int     `json:"dpi"`             // resolución
}

type ConfiguracionDiseno struct {
	ColorFondo           string                 `json:"color_fondo"`
	ColorTexto           string                 `json:"color_texto"`
	ColorBorde           string                 `json:"color_borde"`
	FuentePrincipal      string                 `json:"fuente_principal"`
	FuenteSecundaria     string                 `json:"fuente_secundaria"`
	TamañoFuenteTitulo   int                    `json:"tamaño_fuente_titulo"`
	TamañoFuenteTexto    int                    `json:"tamaño_fuente_texto"`
	TamañoFuentePrecio   int                    `json:"tamaño_fuente_precio"`
	MostrarLogo          bool                   `json:"mostrar_logo"`
	PosicionLogo         string                 `json:"posicion_logo"` // "superior", "inferior", "izquierda", "derecha"
	MostrarCodigoBarra   bool                   `json:"mostrar_codigo_barra"`
	TipoCodigoBarra      string                 `json:"tipo_codigo_barra"` // "CODE39", "CODE128", "EAN13", etc.
	PosicionCodigoBarra  string                 `json:"posicion_codigo_barra"`
	TamañoCodigoBarra    string                 `json:"tamaño_codigo_barra"` // "pequeño", "mediano", "grande"
	ElementosPersonalizados map[string]interface{} `json:"elementos_personalizados,omitempty"`
}

type CampoPersonalizado struct {
	Nombre       string                 `json:"nombre"`
	Etiqueta     string                 `json:"etiqueta"`
	Tipo         string                 `json:"tipo"` // "texto", "numero", "fecha", "booleano"
	Obligatorio  bool                   `json:"obligatorio"`
	ValorDefault *string                `json:"valor_default,omitempty"`
	Validacion   *ValidacionCampo       `json:"validacion,omitempty"`
	Posicion     PosicionCampo          `json:"posicion"`
	Estilo       EstiloCampo            `json:"estilo"`
}

type ValidacionCampo struct {
	LongitudMinima *int    `json:"longitud_minima,omitempty"`
	LongitudMaxima *int    `json:"longitud_maxima,omitempty"`
	PatronRegex    *string `json:"patron_regex,omitempty"`
	ValoresPermitidos []string `json:"valores_permitidos,omitempty"`
}

type PosicionCampo struct {
	X      float64 `json:"x"`      // posición X en mm
	Y      float64 `json:"y"`      // posición Y en mm
	Ancho  float64 `json:"ancho"`  // ancho en mm
	Alto   float64 `json:"alto"`   // alto en mm
	ZIndex int     `json:"z_index"` // orden de superposición
}

type EstiloCampo struct {
	Color           string  `json:"color"`
	TamañoFuente    int     `json:"tamaño_fuente"`
	Fuente          string  `json:"fuente"`
	Negrita         bool    `json:"negrita"`
	Cursiva         bool    `json:"cursiva"`
	Subrayado       bool    `json:"subrayado"`
	Alineacion      string  `json:"alineacion"` // "izquierda", "centro", "derecha"
	ColorFondo      *string `json:"color_fondo,omitempty"`
	Borde           *string `json:"borde,omitempty"`
}

type ConfiguracionImpresion struct {
	ImpresoraDefault     string                 `json:"impresora_default"`
	CalidadImpresion     string                 `json:"calidad_impresion"` // "borrador", "normal", "alta"
	TipoMaterial         string                 `json:"tipo_material"`     // "papel", "adhesivo", "plastico"
	VelocidadImpresion   string                 `json:"velocidad_impresion"` // "lenta", "normal", "rapida"
	ConfiguracionAvanzada map[string]interface{} `json:"configuracion_avanzada,omitempty"`
}

type CacheRenderizado struct {
	TemplateCompilado    string    `json:"template_compilado"`
	HashConfiguracion    string    `json:"hash_configuracion"`
	FechaCompilacion     time.Time `json:"fecha_compilacion"`
	ValidoHasta          time.Time `json:"valido_hasta"`
	TamañoCache          int       `json:"tamaño_cache"` // en bytes
}

type DatosEtiqueta struct {
	CodigoProducto       string                 `json:"codigo_producto"`
	DescripcionProducto  string                 `json:"descripcion_producto"`
	PrecioUnitario       float64                `json:"precio_unitario"`
	CodigoBarra          string                 `json:"codigo_barra"`
	Marca                *string                `json:"marca,omitempty"`
	Modelo               *string                `json:"modelo,omitempty"`
	UnidadMedida         string                 `json:"unidad_medida"`
	FechaGeneracion      time.Time              `json:"fecha_generacion"`
	SucursalNombre       string                 `json:"sucursal_nombre"`
	CamposPersonalizados map[string]interface{} `json:"campos_personalizados,omitempty"`
	DatosAdicionales     map[string]interface{} `json:"datos_adicionales,omitempty"`
}

type ConfiguracionEtiquetaUsada struct {
	PlantillaVersion     int                    `json:"plantilla_version"`
	ConfiguracionDiseno  ConfiguracionDiseno    `json:"configuracion_diseno"`
	DimensionesUsadas    DimensionesEtiqueta    `json:"dimensiones_usadas"`
	CamposUsados         []CampoPersonalizado   `json:"campos_usados,omitempty"`
	ConfiguracionImpresion *ConfiguracionImpresion `json:"configuracion_impresion,omitempty"`
}

type MetadatosGeneracion struct {
	VersionSistema       string                 `json:"version_sistema"`
	MotorRenderizado     string                 `json:"motor_renderizado"`
	TiempoRenderizado    int                    `json:"tiempo_renderizado_ms"`
	TamañoArchivo        int                    `json:"tamaño_archivo_bytes"`
	FormatoSalida        string                 `json:"formato_salida"` // "PDF", "PNG", "SVG"
	ResolucionGenerada   string                 `json:"resolucion_generada"`
	ErroresGeneracion    []string               `json:"errores_generacion,omitempty"`
	AdvertenciasGeneracion []string             `json:"advertencias_generacion,omitempty"`
	DatosDebug           map[string]interface{} `json:"datos_debug,omitempty"`
}

type ConfiguracionLote struct {
	FormatoSalida        string                 `json:"formato_salida"` // "PDF", "ZIP"
	AgruparPorProducto   bool                   `json:"agrupar_por_producto"`
	AgruparPorPlantilla  bool                   `json:"agrupar_por_plantilla"`
	IncluirIndice        bool                   `json:"incluir_indice"`
	ConfiguracionPDF     *ConfiguracionPDF      `json:"configuracion_pdf,omitempty"`
	ConfiguracionZIP     *ConfiguracionZIP      `json:"configuracion_zip,omitempty"`
	MetadatosIncluir     []string               `json:"metadatos_incluir,omitempty"`
}

type ConfiguracionPDF struct {
	EtiquetasPorPagina   int    `json:"etiquetas_por_pagina"`
	OrientacionPagina    string `json:"orientacion_pagina"` // "vertical", "horizontal"
	TamañoPagina         string `json:"tamaño_pagina"`      // "A4", "Letter", "Legal"
	MargenPagina         float64 `json:"margen_pagina"`     // en mm
	EspaciadoEtiquetas   float64 `json:"espaciado_etiquetas"` // en mm
}

type ConfiguracionZIP struct {
	NivelCompresion      int    `json:"nivel_compresion"` // 0-9
	IncluirMetadatos     bool   `json:"incluir_metadatos"`
	FormatoNombreArchivo string `json:"formato_nombre_archivo"`
	CrearCarpetas        bool   `json:"crear_carpetas"`
}

type MetadatosLote struct {
	TotalProductosUnicos int                    `json:"total_productos_unicos"`
	TotalPlantillasUsadas int                   `json:"total_plantillas_usadas"`
	TiempoGeneracionTotal int                   `json:"tiempo_generacion_total_ms"`
	TamañoArchivoTotal    int                   `json:"tamaño_archivo_total_bytes"`
	EstadisticasGeneracion map[string]int       `json:"estadisticas_generacion"`
	ErroresLote           []string               `json:"errores_lote,omitempty"`
	AdvertenciasLote      []string               `json:"advertencias_lote,omitempty"`
}

// Implementar driver.Valuer para tipos JSON personalizados
func (d DimensionesEtiqueta) Value() (driver.Value, error) {
	return json.Marshal(d)
}

func (d *DimensionesEtiqueta) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, d)
}

func (c ConfiguracionDiseno) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c *ConfiguracionDiseno) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, c)
}

func (c []CampoPersonalizado) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c *[]CampoPersonalizado) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, c)
}

func (c ConfiguracionImpresion) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c *ConfiguracionImpresion) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, c)
}

func (c CacheRenderizado) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c *CacheRenderizado) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, c)
}

func (d DatosEtiqueta) Value() (driver.Value, error) {
	return json.Marshal(d)
}

func (d *DatosEtiqueta) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, d)
}

func (c ConfiguracionEtiquetaUsada) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c *ConfiguracionEtiquetaUsada) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, c)
}

func (m MetadatosGeneracion) Value() (driver.Value, error) {
	return json.Marshal(m)
}

func (m *MetadatosGeneracion) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, m)
}

func (c ConfiguracionLote) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c *ConfiguracionLote) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, c)
}

func (m MetadatosLote) Value() (driver.Value, error) {
	return json.Marshal(m)
}

func (m *MetadatosLote) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, m)
}

// DTOs para PlantillaEtiqueta

type PlantillaEtiquetaCreateDTO struct {
	Nombre                string                  `json:"nombre" binding:"required" validate:"required,max=255"`
	Descripcion           *string                 `json:"descripcion,omitempty"`
	TipoEtiqueta          TipoEtiqueta            `json:"tipo_etiqueta" binding:"required"`
	Dimensiones           DimensionesEtiqueta     `json:"dimensiones" binding:"required"`
	ConfiguracionDiseno   ConfiguracionDiseno     `json:"configuracion_diseno" binding:"required"`
	CamposPersonalizados  []CampoPersonalizado    `json:"campos_personalizados,omitempty"`
	EsDefault             bool                    `json:"es_default"`
	SucursalesPermitidas  []uuid.UUID             `json:"sucursales_permitidas,omitempty"`
	CategoriasPermitidas  []uuid.UUID             `json:"categorias_permitidas,omitempty"`
	ConfiguracionImpresion *ConfiguracionImpresion `json:"configuracion_impresion,omitempty"`
}

type PlantillaEtiquetaUpdateDTO struct {
	Nombre                *string                 `json:"nombre,omitempty" validate:"omitempty,max=255"`
	Descripcion           *string                 `json:"descripcion,omitempty"`
	Dimensiones           *DimensionesEtiqueta    `json:"dimensiones,omitempty"`
	ConfiguracionDiseno   *ConfiguracionDiseno    `json:"configuracion_diseno,omitempty"`
	CamposPersonalizados  *[]CampoPersonalizado   `json:"campos_personalizados,omitempty"`
	Activa                *bool                   `json:"activa,omitempty"`
	EsDefault             *bool                   `json:"es_default,omitempty"`
	SucursalesPermitidas  *[]uuid.UUID            `json:"sucursales_permitidas,omitempty"`
	CategoriasPermitidas  *[]uuid.UUID            `json:"categorias_permitidas,omitempty"`
	ConfiguracionImpresion *ConfiguracionImpresion `json:"configuracion_impresion,omitempty"`
}

type PlantillaEtiquetaResponseDTO struct {
	ID                    uuid.UUID               `json:"id"`
	Nombre                string                  `json:"nombre"`
	Descripcion           *string                 `json:"descripcion,omitempty"`
	TipoEtiqueta          TipoEtiqueta            `json:"tipo_etiqueta"`
	Dimensiones           DimensionesEtiqueta     `json:"dimensiones"`
	ConfiguracionDiseno   ConfiguracionDiseno     `json:"configuracion_diseno"`
	CamposPersonalizados  []CampoPersonalizado    `json:"campos_personalizados,omitempty"`
	Activa                bool                    `json:"activa"`
	EsDefault             bool                    `json:"es_default"`
	SucursalesPermitidas  []uuid.UUID             `json:"sucursales_permitidas,omitempty"`
	CategoriasPermitidas  []uuid.UUID             `json:"categorias_permitidas,omitempty"`
	TotalGeneradas        int                     `json:"total_generadas"`
	FechaUltimoUso        *time.Time              `json:"fecha_ultimo_uso,omitempty"`
	ConfiguracionImpresion *ConfiguracionImpresion `json:"configuracion_impresion,omitempty"`
	VersionPlantilla      int                     `json:"version_plantilla"`
	FechaCreacion         time.Time               `json:"fecha_creacion"`
	FechaModificacion     time.Time               `json:"fecha_modificacion"`
}

type PlantillaEtiquetaListDTO struct {
	ID               uuid.UUID    `json:"id"`
	Nombre           string       `json:"nombre"`
	TipoEtiqueta     TipoEtiqueta `json:"tipo_etiqueta"`
	Activa           bool         `json:"activa"`
	EsDefault        bool         `json:"es_default"`
	TotalGeneradas   int          `json:"total_generadas"`
	FechaUltimoUso   *time.Time   `json:"fecha_ultimo_uso,omitempty"`
	VersionPlantilla int          `json:"version_plantilla"`
	FechaCreacion    time.Time    `json:"fecha_creacion"`
	FechaModificacion time.Time   `json:"fecha_modificacion"`
}

// DTOs para EtiquetaGenerada

type EtiquetaGeneradaCreateDTO struct {
	PlantillaID           uuid.UUID                   `json:"plantilla_id" binding:"required"`
	ProductoID            uuid.UUID                   `json:"producto_id" binding:"required"`
	SucursalID            uuid.UUID                   `json:"sucursal_id" binding:"required"`
	Cantidad              int                         `json:"cantidad" binding:"required,min=1"`
	CamposPersonalizados  map[string]interface{}      `json:"campos_personalizados,omitempty"`
	ConfiguracionEspecial *ConfiguracionEtiquetaUsada `json:"configuracion_especial,omitempty"`
}

type EtiquetaGeneradaResponseDTO struct {
	ID                  uuid.UUID                   `json:"id"`
	PlantillaID         uuid.UUID                   `json:"plantilla_id"`
	ProductoID          uuid.UUID                   `json:"producto_id"`
	SucursalID          uuid.UUID                   `json:"sucursal_id"`
	UsuarioID           uuid.UUID                   `json:"usuario_id"`
	Cantidad            int                         `json:"cantidad"`
	Estado              EstadoEtiqueta              `json:"estado"`
	FechaGeneracion     time.Time                   `json:"fecha_generacion"`
	FechaImpresion      *time.Time                  `json:"fecha_impresion,omitempty"`
	TerminalImpresion   *uuid.UUID                  `json:"terminal_impresion,omitempty"`
	DatosEtiqueta       DatosEtiqueta               `json:"datos_etiqueta"`
	ArchivoGenerado     *string                     `json:"archivo_generado,omitempty"`
	ConfiguracionUsada  *ConfiguracionEtiquetaUsada `json:"configuracion_usada,omitempty"`
	MetadatosGeneracion *MetadatosGeneracion        `json:"metadatos_generacion,omitempty"`
	LoteGeneracion      *uuid.UUID                  `json:"lote_generacion,omitempty"`
	TiempoGeneracion    *int                        `json:"tiempo_generacion_ms,omitempty"`
	FechaCreacion       time.Time                   `json:"fecha_creacion"`
	PlantillaNombre     *string                     `json:"plantilla_nombre,omitempty"`
	ProductoNombre      *string                     `json:"producto_nombre,omitempty"`
	SucursalNombre      *string                     `json:"sucursal_nombre,omitempty"`
	UsuarioNombre       *string                     `json:"usuario_nombre,omitempty"`
}

type EtiquetaGeneradaListDTO struct {
	ID                uuid.UUID      `json:"id"`
	PlantillaID       uuid.UUID      `json:"plantilla_id"`
	ProductoID        uuid.UUID      `json:"producto_id"`
	SucursalID        uuid.UUID      `json:"sucursal_id"`
	Cantidad          int            `json:"cantidad"`
	Estado            EstadoEtiqueta `json:"estado"`
	FechaGeneracion   time.Time      `json:"fecha_generacion"`
	FechaImpresion    *time.Time     `json:"fecha_impresion,omitempty"`
	TiempoGeneracion  *int           `json:"tiempo_generacion_ms,omitempty"`
	PlantillaNombre   *string        `json:"plantilla_nombre,omitempty"`
	ProductoNombre    *string        `json:"producto_nombre,omitempty"`
	SucursalNombre    *string        `json:"sucursal_nombre,omitempty"`
}

// DTOs para LoteEtiquetas

type LoteEtiquetasCreateDTO struct {
	Nombre            string                    `json:"nombre" binding:"required"`
	Descripcion       *string                   `json:"descripcion,omitempty"`
	SucursalID        uuid.UUID                 `json:"sucursal_id" binding:"required"`
	ConfiguracionLote *ConfiguracionLote        `json:"configuracion_lote,omitempty"`
	Etiquetas         []EtiquetaGeneradaCreateDTO `json:"etiquetas" binding:"required,min=1"`
}

type LoteEtiquetasResponseDTO struct {
	ID                 uuid.UUID          `json:"id"`
	Nombre             string             `json:"nombre"`
	Descripcion        *string            `json:"descripcion,omitempty"`
	SucursalID         uuid.UUID          `json:"sucursal_id"`
	UsuarioID          uuid.UUID          `json:"usuario_id"`
	TotalEtiquetas     int                `json:"total_etiquetas"`
	EtiquetasImpresas  int                `json:"etiquetas_impresas"`
	Estado             EstadoLote         `json:"estado"`
	FechaGeneracion    time.Time          `json:"fecha_generacion"`
	FechaCompletado    *time.Time         `json:"fecha_completado,omitempty"`
	ConfiguracionLote  *ConfiguracionLote `json:"configuracion_lote,omitempty"`
	ArchivoConsolidado *string            `json:"archivo_consolidado,omitempty"`
	MetadatosLote      *MetadatosLote     `json:"metadatos_lote,omitempty"`
	FechaCreacion      time.Time          `json:"fecha_creacion"`
	SucursalNombre     *string            `json:"sucursal_nombre,omitempty"`
	UsuarioNombre      *string            `json:"usuario_nombre,omitempty"`
}

// Filtros para búsqueda

type PlantillaEtiquetaFilter struct {
	PaginationFilter
	SortFilter
	Nombre       *string       `json:"nombre,omitempty" form:"nombre"`
	TipoEtiqueta *TipoEtiqueta `json:"tipo_etiqueta,omitempty" form:"tipo_etiqueta"`
	Activa       *bool         `json:"activa,omitempty" form:"activa"`
	EsDefault    *bool         `json:"es_default,omitempty" form:"es_default"`
	SucursalID   *uuid.UUID    `json:"sucursal_id,omitempty" form:"sucursal_id"`
	CategoriaID  *uuid.UUID    `json:"categoria_id,omitempty" form:"categoria_id"`
}

type EtiquetaGeneradaFilter struct {
	PaginationFilter
	SortFilter
	DateRangeFilter
	PlantillaID       *uuid.UUID      `json:"plantilla_id,omitempty" form:"plantilla_id"`
	ProductoID        *uuid.UUID      `json:"producto_id,omitempty" form:"producto_id"`
	SucursalID        *uuid.UUID      `json:"sucursal_id,omitempty" form:"sucursal_id"`
	UsuarioID         *uuid.UUID      `json:"usuario_id,omitempty" form:"usuario_id"`
	Estado            *EstadoEtiqueta `json:"estado,omitempty" form:"estado"`
	LoteGeneracion    *uuid.UUID      `json:"lote_generacion,omitempty" form:"lote_generacion"`
	TerminalImpresion *uuid.UUID      `json:"terminal_impresion,omitempty" form:"terminal_impresion"`
}

type LoteEtiquetasFilter struct {
	PaginationFilter
	SortFilter
	DateRangeFilter
	Nombre     *string     `json:"nombre,omitempty" form:"nombre"`
	SucursalID *uuid.UUID  `json:"sucursal_id,omitempty" form:"sucursal_id"`
	UsuarioID  *uuid.UUID  `json:"usuario_id,omitempty" form:"usuario_id"`
	Estado     *EstadoLote `json:"estado,omitempty" form:"estado"`
}

// Métodos helper

func (p *PlantillaEtiqueta) PuedeUsarEnSucursal(sucursalID uuid.UUID) bool {
	if len(p.SucursalesPermitidas) == 0 {
		return true // Sin restricciones
	}
	
	for _, id := range p.SucursalesPermitidas {
		if id == sucursalID {
			return true
		}
	}
	
	return false
}

func (p *PlantillaEtiqueta) PuedeUsarEnCategoria(categoriaID uuid.UUID) bool {
	if len(p.CategoriasPermitidas) == 0 {
		return true // Sin restricciones
	}
	
	for _, id := range p.CategoriasPermitidas {
		if id == categoriaID {
			return true
		}
	}
	
	return false
}

func (p *PlantillaEtiqueta) ActualizarUso() {
	p.TotalGeneradas++
	now := time.Now()
	p.FechaUltimoUso = &now
}

func (e *EtiquetaGenerada) MarcarComoImpresa(terminalID *uuid.UUID) {
	e.Estado = EtiquetaImpresa
	now := time.Now()
	e.FechaImpresion = &now
	e.TerminalImpresion = terminalID
}

func (e *EtiquetaGenerada) PuedeImprimir() bool {
	return e.Estado == EtiquetaGenerada && e.ArchivoGenerado != nil
}

func (l *LoteEtiquetas) ActualizarProgreso() {
	// En implementación real, contar etiquetas impresas
	if l.EtiquetasImpresas >= l.TotalEtiquetas && l.Estado == LoteGenerando {
		l.Estado = LoteCompletado
		now := time.Now()
		l.FechaCompletado = &now
	}
}

func (l *LoteEtiquetas) PuedeCompletar() bool {
	return l.Estado == LoteGenerando || l.Estado == LotePendiente
}

func (p *PlantillaEtiqueta) ToResponseDTO() PlantillaEtiquetaResponseDTO {
	return PlantillaEtiquetaResponseDTO{
		ID:                    p.ID,
		Nombre:                p.Nombre,
		Descripcion:           p.Descripcion,
		TipoEtiqueta:          p.TipoEtiqueta,
		Dimensiones:           p.Dimensiones,
		ConfiguracionDiseno:   p.ConfiguracionDiseno,
		CamposPersonalizados:  p.CamposPersonalizados,
		Activa:                p.Activa,
		EsDefault:             p.EsDefault,
		SucursalesPermitidas:  p.SucursalesPermitidas,
		CategoriasPermitidas:  p.CategoriasPermitidas,
		TotalGeneradas:        p.TotalGeneradas,
		FechaUltimoUso:        p.FechaUltimoUso,
		ConfiguracionImpresion: p.ConfiguracionImpresion,
		VersionPlantilla:      p.VersionPlantilla,
		FechaCreacion:         p.FechaCreacion,
		FechaModificacion:     p.FechaModificacion,
	}
}

func (p *PlantillaEtiqueta) ToListDTO() PlantillaEtiquetaListDTO {
	return PlantillaEtiquetaListDTO{
		ID:               p.ID,
		Nombre:           p.Nombre,
		TipoEtiqueta:     p.TipoEtiqueta,
		Activa:           p.Activa,
		EsDefault:        p.EsDefault,
		TotalGeneradas:   p.TotalGeneradas,
		FechaUltimoUso:   p.FechaUltimoUso,
		VersionPlantilla: p.VersionPlantilla,
		FechaCreacion:    p.FechaCreacion,
		FechaModificacion: p.FechaModificacion,
	}
}

func (e *EtiquetaGenerada) ToResponseDTO() EtiquetaGeneradaResponseDTO {
	dto := EtiquetaGeneradaResponseDTO{
		ID:                  e.ID,
		PlantillaID:         e.PlantillaID,
		ProductoID:          e.ProductoID,
		SucursalID:          e.SucursalID,
		UsuarioID:           e.UsuarioID,
		Cantidad:            e.Cantidad,
		Estado:              e.Estado,
		FechaGeneracion:     e.FechaGeneracion,
		FechaImpresion:      e.FechaImpresion,
		TerminalImpresion:   e.TerminalImpresion,
		DatosEtiqueta:       e.DatosEtiqueta,
		ArchivoGenerado:     e.ArchivoGenerado,
		ConfiguracionUsada:  e.ConfiguracionUsada,
		MetadatosGeneracion: e.MetadatosGeneracion,
		LoteGeneracion:      e.LoteGeneracion,
		TiempoGeneracion:    e.TiempoGeneracion,
		FechaCreacion:       e.FechaCreacion,
	}
	
	if e.Plantilla != nil {
		dto.PlantillaNombre = &e.Plantilla.Nombre
	}
	
	if e.Producto != nil {
		dto.ProductoNombre = &e.Producto.Descripcion
	}
	
	if e.Sucursal != nil {
		dto.SucursalNombre = &e.Sucursal.Nombre
	}
	
	if e.Usuario != nil {
		nombreUsuario := e.Usuario.Nombre
		if e.Usuario.Apellido != nil {
			nombreUsuario += " " + *e.Usuario.Apellido
		}
		dto.UsuarioNombre = &nombreUsuario
	}
	
	return dto
}

func (e *EtiquetaGenerada) ToListDTO() EtiquetaGeneradaListDTO {
	dto := EtiquetaGeneradaListDTO{
		ID:               e.ID,
		PlantillaID:      e.PlantillaID,
		ProductoID:       e.ProductoID,
		SucursalID:       e.SucursalID,
		Cantidad:         e.Cantidad,
		Estado:           e.Estado,
		FechaGeneracion:  e.FechaGeneracion,
		FechaImpresion:   e.FechaImpresion,
		TiempoGeneracion: e.TiempoGeneracion,
	}
	
	if e.Plantilla != nil {
		dto.PlantillaNombre = &e.Plantilla.Nombre
	}
	
	if e.Producto != nil {
		dto.ProductoNombre = &e.Producto.Descripcion
	}
	
	if e.Sucursal != nil {
		dto.SucursalNombre = &e.Sucursal.Nombre
	}
	
	return dto
}

func (l *LoteEtiquetas) ToResponseDTO() LoteEtiquetasResponseDTO {
	dto := LoteEtiquetasResponseDTO{
		ID:                 l.ID,
		Nombre:             l.Nombre,
		Descripcion:        l.Descripcion,
		SucursalID:         l.SucursalID,
		UsuarioID:          l.UsuarioID,
		TotalEtiquetas:     l.TotalEtiquetas,
		EtiquetasImpresas:  l.EtiquetasImpresas,
		Estado:             l.Estado,
		FechaGeneracion:    l.FechaGeneracion,
		FechaCompletado:    l.FechaCompletado,
		ConfiguracionLote:  l.ConfiguracionLote,
		ArchivoConsolidado: l.ArchivoConsolidado,
		MetadatosLote:      l.MetadatosLote,
		FechaCreacion:      l.FechaCreacion,
	}
	
	if l.Sucursal != nil {
		dto.SucursalNombre = &l.Sucursal.Nombre
	}
	
	if l.Usuario != nil {
		nombreUsuario := l.Usuario.Nombre
		if l.Usuario.Apellido != nil {
			nombreUsuario += " " + *l.Usuario.Apellido
		}
		dto.UsuarioNombre = &nombreUsuario
	}
	
	return dto
}

func (dto *PlantillaEtiquetaCreateDTO) ToModel() *PlantillaEtiqueta {
	return &PlantillaEtiqueta{
		Nombre:                dto.Nombre,
		Descripcion:           dto.Descripcion,
		TipoEtiqueta:          dto.TipoEtiqueta,
		Dimensiones:           dto.Dimensiones,
		ConfiguracionDiseno:   dto.ConfiguracionDiseno,
		CamposPersonalizados:  dto.CamposPersonalizados,
		Activa:                true,
		EsDefault:             dto.EsDefault,
		SucursalesPermitidas:  dto.SucursalesPermitidas,
		CategoriasPermitidas:  dto.CategoriasPermitidas,
		ConfiguracionImpresion: dto.ConfiguracionImpresion,
		TotalGeneradas:        0,
		VersionPlantilla:      1,
	}
}

// Validaciones personalizadas

func (p *PlantillaEtiqueta) Validate() error {
	if p.Dimensiones.Ancho <= 0 || p.Dimensiones.Alto <= 0 {
		return fmt.Errorf("dimensiones de etiqueta deben ser positivas")
	}
	
	if p.Dimensiones.DPI <= 0 {
		return fmt.Errorf("DPI debe ser positivo")
	}
	
	if p.VersionPlantilla <= 0 {
		return fmt.Errorf("versión de plantilla debe ser positiva")
	}
	
	return nil
}

func (e *EtiquetaGenerada) Validate() error {
	if e.Cantidad <= 0 {
		return fmt.Errorf("cantidad debe ser mayor a 0")
	}
	
	return nil
}

func (l *LoteEtiquetas) Validate() error {
	if l.TotalEtiquetas < 0 {
		return fmt.Errorf("total de etiquetas no puede ser negativo")
	}
	
	if l.EtiquetasImpresas < 0 {
		return fmt.Errorf("etiquetas impresas no puede ser negativo")
	}
	
	if l.EtiquetasImpresas > l.TotalEtiquetas {
		return fmt.Errorf("etiquetas impresas no puede ser mayor al total")
	}
	
	return nil
}

// Funciones de utilidad

func GenerarHashIntegridad(datos DatosEtiqueta) string {
	// En implementación real, generar hash SHA256
	return fmt.Sprintf("hash_%d", time.Now().Unix())
}

func ValidarDimensionesEtiqueta(dimensiones DimensionesEtiqueta) error {
	if dimensiones.Ancho <= 0 || dimensiones.Alto <= 0 {
		return fmt.Errorf("ancho y alto deben ser positivos")
	}
	
	if dimensiones.DPI < 72 || dimensiones.DPI > 600 {
		return fmt.Errorf("DPI debe estar entre 72 y 600")
	}
	
	return nil
}

