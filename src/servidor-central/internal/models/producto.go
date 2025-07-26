package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Producto representa un producto en el sistema
type Producto struct {
	BaseModelWithUser
	CodigoInterno             string                    `json:"codigo_interno" db:"codigo_interno" binding:"required" validate:"required,max=100"`
	CodigoBarra               string                    `json:"codigo_barra" db:"codigo_barra" binding:"required" validate:"required,max=100"`
	Descripcion               string                    `json:"descripcion" db:"descripcion" binding:"required" validate:"required,max=500"`
	DescripcionCorta          *string                   `json:"descripcion_corta,omitempty" db:"descripcion_corta"`
	CategoriaID               *uuid.UUID                `json:"categoria_id,omitempty" db:"categoria_id"`
	Marca                     *string                   `json:"marca,omitempty" db:"marca"`
	Modelo                    *string                   `json:"modelo,omitempty" db:"modelo"`
	PrecioUnitario            float64                   `json:"precio_unitario" db:"precio_unitario" binding:"required,min=0"`
	PrecioCosto               *float64                  `json:"precio_costo,omitempty" db:"precio_costo" validate:"omitempty,min=0"`
	UnidadMedida              string                    `json:"unidad_medida" db:"unidad_medida" default:"UN"`
	Peso                      *float64                  `json:"peso,omitempty" db:"peso" validate:"omitempty,min=0"`
	Dimensiones               *DimensionesProducto      `json:"dimensiones,omitempty" db:"dimensiones"`
	EspecificacionesTecnicas  *EspecificacionesTecnicas `json:"especificaciones_tecnicas,omitempty" db:"especificaciones_tecnicas"`
	Activo                    bool                      `json:"activo" db:"activo" default:"true"`
	RequiereSerie             bool                      `json:"requiere_serie" db:"requiere_serie" default:"false"`
	PermiteFraccionamiento    bool                      `json:"permite_fraccionamiento" db:"permite_fraccionamiento" default:"false"`
	StockMinimo               int                       `json:"stock_minimo" db:"stock_minimo" default:"0"`
	StockMaximo               *int                      `json:"stock_maximo,omitempty" db:"stock_maximo"`
	ImagenPrincipalURL        *string                   `json:"imagen_principal_url,omitempty" db:"imagen_principal_url"`
	ImagenesAdicionales       *ImagenesAdicionales      `json:"imagenes_adicionales,omitempty" db:"imagenes_adicionales"`
	DescripcionBusqueda       *string                   `json:"-" db:"descripcion_busqueda"` // TSVECTOR en PostgreSQL
	PopularidadScore          float64                   `json:"popularidad_score" db:"popularidad_score" default:"0"`
	CacheCodigoBarrasGenerado *string                   `json:"cache_codigo_barras_generado,omitempty" db:"cache_codigo_barras_generado"`
	ConfiguracionEtiqueta     *ConfiguracionEtiqueta    `json:"configuracion_etiqueta,omitempty" db:"configuracion_etiqueta"`
	FechaUltimaEtiqueta       *time.Time                `json:"fecha_ultima_etiqueta,omitempty" db:"fecha_ultima_etiqueta"`
	TotalEtiquetasGeneradas   int                       `json:"total_etiquetas_generadas" db:"total_etiquetas_generadas" default:"0"`

	// Relaciones
	Categoria               *CategoriaProducto     `json:"categoria,omitempty" gorm:"foreignKey:CategoriaID"`
	CodigosBarraAdicionales []CodigoBarraAdicional `json:"codigos_barra_adicionales,omitempty" gorm:"foreignKey:ProductoID"`
	Stock                   []StockCentral         `json:"stock,omitempty" gorm:"foreignKey:ProductoID"`
}

// DimensionesProducto contiene las dimensiones físicas del producto
type DimensionesProducto struct {
	Largo   *float64 `json:"largo,omitempty"`   // en cm
	Ancho   *float64 `json:"ancho,omitempty"`   // en cm
	Alto    *float64 `json:"alto,omitempty"`    // en cm
	Volumen *float64 `json:"volumen,omitempty"` // en cm³
}

// EspecificacionesTecnicas contiene especificaciones técnicas del producto
type EspecificacionesTecnicas struct {
	Material         *string                `json:"material,omitempty"`
	Color            *string                `json:"color,omitempty"`
	Voltaje          *string                `json:"voltaje,omitempty"`
	Potencia         *string                `json:"potencia,omitempty"`
	Capacidad        *string                `json:"capacidad,omitempty"`
	Resistencia      *string                `json:"resistencia,omitempty"`
	Temperatura      *string                `json:"temperatura,omitempty"`
	Presion          *string                `json:"presion,omitempty"`
	Certificaciones  []string               `json:"certificaciones,omitempty"`
	Especificaciones map[string]interface{} `json:"especificaciones,omitempty"`
}

// ImagenesAdicionales contiene URLs de imágenes adicionales del producto
type ImagenesAdicionales struct {
	URLs          []string `json:"urls"`
	Thumbnails    []string `json:"thumbnails,omitempty"`
	Descripciones []string `json:"descripciones,omitempty"`
}

// ConfiguracionEtiqueta contiene configuración específica para etiquetas del producto
type ConfiguracionEtiqueta struct {
	PlantillaID                *uuid.UUID             `json:"plantilla_id,omitempty"`
	MostrarMarca               bool                   `json:"mostrar_marca"`
	MostrarModelo              bool                   `json:"mostrar_modelo"`
	MostrarDimensiones         bool                   `json:"mostrar_dimensiones"`
	MostrarPeso                bool                   `json:"mostrar_peso"`
	MostrarEspecificaciones    bool                   `json:"mostrar_especificaciones"`
	TipoCodigoBarra            string                 `json:"tipo_codigo_barra"` // CODE39, CODE128, EAN13, etc.
	TamañoEtiqueta             string                 `json:"tamaño_etiqueta"`   // pequeña, mediana, grande
	ConfiguracionPersonalizada map[string]interface{} `json:"configuracion_personalizada,omitempty"`
}

// CategoriaProducto representa una categoría de productos
type CategoriaProducto struct {
	BaseModel
	Codigo                 string                           `json:"codigo" db:"codigo" binding:"required" validate:"required,max=50"`
	Nombre                 string                           `json:"nombre" db:"nombre" binding:"required" validate:"required,max=255"`
	Descripcion            *string                          `json:"descripcion,omitempty" db:"descripcion"`
	CategoriaPadreID       *uuid.UUID                       `json:"categoria_padre_id,omitempty" db:"categoria_padre_id"`
	Nivel                  int                              `json:"nivel" db:"nivel" default:"1"`
	Activa                 bool                             `json:"activa" db:"activa" default:"true"`
	OrdenVisualizacion     *int                             `json:"orden_visualizacion,omitempty" db:"orden_visualizacion"`
	ImagenURL              *string                          `json:"imagen_url,omitempty" db:"imagen_url"`
	PathCompleto           *string                          `json:"path_completo,omitempty" db:"path_completo"`
	TotalProductos         int                              `json:"total_productos" db:"total_productos" default:"0"`
	ConfiguracionEtiquetas *ConfiguracionEtiquetasCategoria `json:"configuracion_etiquetas,omitempty" db:"configuracion_etiquetas"`

	// Relaciones
	CategoriaPadre *CategoriaProducto  `json:"categoria_padre,omitempty" gorm:"foreignKey:CategoriaPadreID"`
	Subcategorias  []CategoriaProducto `json:"subcategorias,omitempty" gorm:"foreignKey:CategoriaPadreID"`
	Productos      []Producto          `json:"productos,omitempty" gorm:"foreignKey:CategoriaID"`
}

// ConfiguracionEtiquetasCategoria contiene configuración de etiquetas específica por categoría
type ConfiguracionEtiquetasCategoria struct {
	PlantillaDefault     string                 `json:"plantilla_default"`
	MostrarMarca         bool                   `json:"mostrar_marca"`
	MostrarModelo        bool                   `json:"mostrar_modelo"`
	MostrarDimensiones   bool                   `json:"mostrar_dimensiones"`
	MostrarPeso          bool                   `json:"mostrar_peso"`
	CamposPersonalizados []string               `json:"campos_personalizados,omitempty"`
	ConfiguracionExtra   map[string]interface{} `json:"configuracion_extra,omitempty"`
}

// CodigoBarraAdicional representa códigos de barra adicionales para un producto
type CodigoBarraAdicional struct {
	BaseModel
	ProductoID  uuid.UUID `json:"producto_id" db:"producto_id" binding:"required"`
	CodigoBarra string    `json:"codigo_barra" db:"codigo_barra" binding:"required"`
	Descripcion *string   `json:"descripcion,omitempty" db:"descripcion"`
	Activo      bool      `json:"activo" db:"activo" default:"true"`
	TipoCodigo  string    `json:"tipo_codigo" db:"tipo_codigo" default:"EAN13"`
	Validado    bool      `json:"validado" db:"validado" default:"false"`

	// Relaciones
	Producto *Producto `json:"producto,omitempty" gorm:"foreignKey:ProductoID"`
}

// Implementar driver.Valuer para tipos JSON personalizados
func (d DimensionesProducto) Value() (driver.Value, error) {
	return json.Marshal(d)
}

func (d *DimensionesProducto) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, d)
}

func (e EspecificacionesTecnicas) Value() (driver.Value, error) {
	return json.Marshal(e)
}

func (e *EspecificacionesTecnicas) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, e)
}

func (i ImagenesAdicionales) Value() (driver.Value, error) {
	return json.Marshal(i)
}

func (i *ImagenesAdicionales) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, i)
}

func (c ConfiguracionEtiqueta) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c *ConfiguracionEtiqueta) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, c)
}

func (c ConfiguracionEtiquetasCategoria) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c *ConfiguracionEtiquetasCategoria) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, c)
}

// DTOs para Producto

type ProductoCreateDTO struct {
	CodigoInterno            string                    `json:"codigo_interno" binding:"required" validate:"required,max=100"`
	CodigoBarra              string                    `json:"codigo_barra" binding:"required" validate:"required,max=100"`
	Descripcion              string                    `json:"descripcion" binding:"required" validate:"required,max=500"`
	DescripcionCorta         *string                   `json:"descripcion_corta,omitempty"`
	CategoriaID              *uuid.UUID                `json:"categoria_id,omitempty"`
	Marca                    *string                   `json:"marca,omitempty"`
	Modelo                   *string                   `json:"modelo,omitempty"`
	PrecioUnitario           float64                   `json:"precio_unitario" binding:"required,min=0"`
	PrecioCosto              *float64                  `json:"precio_costo,omitempty" validate:"omitempty,min=0"`
	UnidadMedida             string                    `json:"unidad_medida"`
	Peso                     *float64                  `json:"peso,omitempty" validate:"omitempty,min=0"`
	Dimensiones              *DimensionesProducto      `json:"dimensiones,omitempty"`
	EspecificacionesTecnicas *EspecificacionesTecnicas `json:"especificaciones_tecnicas,omitempty"`
	RequiereSerie            bool                      `json:"requiere_serie"`
	PermiteFraccionamiento   bool                      `json:"permite_fraccionamiento"`
	StockMinimo              int                       `json:"stock_minimo"`
	StockMaximo              *int                      `json:"stock_maximo,omitempty"`
	ImagenPrincipalURL       *string                   `json:"imagen_principal_url,omitempty"`
	ImagenesAdicionales      *ImagenesAdicionales      `json:"imagenes_adicionales,omitempty"`
	ConfiguracionEtiqueta    *ConfiguracionEtiqueta    `json:"configuracion_etiqueta,omitempty"`
	CodigosBarraAdicionales  []string                  `json:"codigos_barra_adicionales,omitempty"`
}

type ProductoUpdateDTO struct {
	Descripcion              *string                   `json:"descripcion,omitempty" validate:"omitempty,max=500"`
	DescripcionCorta         *string                   `json:"descripcion_corta,omitempty"`
	CategoriaID              *uuid.UUID                `json:"categoria_id,omitempty"`
	Marca                    *string                   `json:"marca,omitempty"`
	Modelo                   *string                   `json:"modelo,omitempty"`
	PrecioUnitario           *float64                  `json:"precio_unitario,omitempty" validate:"omitempty,min=0"`
	PrecioCosto              *float64                  `json:"precio_costo,omitempty" validate:"omitempty,min=0"`
	UnidadMedida             *string                   `json:"unidad_medida,omitempty"`
	Peso                     *float64                  `json:"peso,omitempty" validate:"omitempty,min=0"`
	Dimensiones              *DimensionesProducto      `json:"dimensiones,omitempty"`
	EspecificacionesTecnicas *EspecificacionesTecnicas `json:"especificaciones_tecnicas,omitempty"`
	Activo                   *bool                     `json:"activo,omitempty"`
	RequiereSerie            *bool                     `json:"requiere_serie,omitempty"`
	PermiteFraccionamiento   *bool                     `json:"permite_fraccionamiento,omitempty"`
	StockMinimo              *int                      `json:"stock_minimo,omitempty"`
	StockMaximo              *int                      `json:"stock_maximo,omitempty"`
	ImagenPrincipalURL       *string                   `json:"imagen_principal_url,omitempty"`
	ImagenesAdicionales      *ImagenesAdicionales      `json:"imagenes_adicionales,omitempty"`
	ConfiguracionEtiqueta    *ConfiguracionEtiqueta    `json:"configuracion_etiqueta,omitempty"`
}

type ProductoResponseDTO struct {
	ID                       uuid.UUID                 `json:"id"`
	CodigoInterno            string                    `json:"codigo_interno"`
	CodigoBarra              string                    `json:"codigo_barra"`
	Descripcion              string                    `json:"descripcion"`
	DescripcionCorta         *string                   `json:"descripcion_corta,omitempty"`
	CategoriaID              *uuid.UUID                `json:"categoria_id,omitempty"`
	Marca                    *string                   `json:"marca,omitempty"`
	Modelo                   *string                   `json:"modelo,omitempty"`
	PrecioUnitario           float64                   `json:"precio_unitario"`
	PrecioCosto              *float64                  `json:"precio_costo,omitempty"`
	UnidadMedida             string                    `json:"unidad_medida"`
	Peso                     *float64                  `json:"peso,omitempty"`
	Dimensiones              *DimensionesProducto      `json:"dimensiones,omitempty"`
	EspecificacionesTecnicas *EspecificacionesTecnicas `json:"especificaciones_tecnicas,omitempty"`
	Activo                   bool                      `json:"activo"`
	RequiereSerie            bool                      `json:"requiere_serie"`
	PermiteFraccionamiento   bool                      `json:"permite_fraccionamiento"`
	StockMinimo              int                       `json:"stock_minimo"`
	StockMaximo              *int                      `json:"stock_maximo,omitempty"`
	ImagenPrincipalURL       *string                   `json:"imagen_principal_url,omitempty"`
	ImagenesAdicionales      *ImagenesAdicionales      `json:"imagenes_adicionales,omitempty"`
	PopularidadScore         float64                   `json:"popularidad_score"`
	ConfiguracionEtiqueta    *ConfiguracionEtiqueta    `json:"configuracion_etiqueta,omitempty"`
	FechaUltimaEtiqueta      *time.Time                `json:"fecha_ultima_etiqueta,omitempty"`
	TotalEtiquetasGeneradas  int                       `json:"total_etiquetas_generadas"`
	FechaCreacion            time.Time                 `json:"fecha_creacion"`
	FechaModificacion        time.Time                 `json:"fecha_modificacion"`
	Categoria                *CategoriaProductoListDTO `json:"categoria,omitempty"`
	CodigosBarraAdicionales  []CodigoBarraAdicionalDTO `json:"codigos_barra_adicionales,omitempty"`
	Stock                    []StockCentralDTO         `json:"stock,omitempty"`
}

type ProductoListDTO struct {
	ID                uuid.UUID `json:"id"`
	CodigoInterno     string    `json:"codigo_interno"`
	CodigoBarra       string    `json:"codigo_barra"`
	Descripcion       string    `json:"descripcion"`
	DescripcionCorta  *string   `json:"descripcion_corta,omitempty"`
	Marca             *string   `json:"marca,omitempty"`
	Modelo            *string   `json:"modelo,omitempty"`
	PrecioUnitario    float64   `json:"precio_unitario"`
	UnidadMedida      string    `json:"unidad_medida"`
	Activo            bool      `json:"activo"`
	PopularidadScore  float64   `json:"popularidad_score"`
	FechaCreacion     time.Time `json:"fecha_creacion"`
	FechaModificacion time.Time `json:"fecha_modificacion"`
	CategoriaNombre   *string   `json:"categoria_nombre,omitempty"`
}

type ProductoBusquedaDTO struct {
	ID               uuid.UUID `json:"id"`
	CodigoInterno    string    `json:"codigo_interno"`
	CodigoBarra      string    `json:"codigo_barra"`
	Descripcion      string    `json:"descripcion"`
	DescripcionCorta *string   `json:"descripcion_corta,omitempty"`
	Marca            *string   `json:"marca,omitempty"`
	PrecioUnitario   float64   `json:"precio_unitario"`
	UnidadMedida     string    `json:"unidad_medida"`
	StockDisponible  int       `json:"stock_disponible"`
	Score            float64   `json:"score"`
}

// DTOs para CategoriaProducto

type CategoriaProductoCreateDTO struct {
	Codigo                 string                           `json:"codigo" binding:"required" validate:"required,max=50"`
	Nombre                 string                           `json:"nombre" binding:"required" validate:"required,max=255"`
	Descripcion            *string                          `json:"descripcion,omitempty"`
	CategoriaPadreID       *uuid.UUID                       `json:"categoria_padre_id,omitempty"`
	OrdenVisualizacion     *int                             `json:"orden_visualizacion,omitempty"`
	ImagenURL              *string                          `json:"imagen_url,omitempty"`
	ConfiguracionEtiquetas *ConfiguracionEtiquetasCategoria `json:"configuracion_etiquetas,omitempty"`
}

type CategoriaProductoUpdateDTO struct {
	Nombre                 *string                          `json:"nombre,omitempty" validate:"omitempty,max=255"`
	Descripcion            *string                          `json:"descripcion,omitempty"`
	CategoriaPadreID       *uuid.UUID                       `json:"categoria_padre_id,omitempty"`
	Activa                 *bool                            `json:"activa,omitempty"`
	OrdenVisualizacion     *int                             `json:"orden_visualizacion,omitempty"`
	ImagenURL              *string                          `json:"imagen_url,omitempty"`
	ConfiguracionEtiquetas *ConfiguracionEtiquetasCategoria `json:"configuracion_etiquetas,omitempty"`
}

type CategoriaProductoResponseDTO struct {
	ID                     uuid.UUID                        `json:"id"`
	Codigo                 string                           `json:"codigo"`
	Nombre                 string                           `json:"nombre"`
	Descripcion            *string                          `json:"descripcion,omitempty"`
	CategoriaPadreID       *uuid.UUID                       `json:"categoria_padre_id,omitempty"`
	Nivel                  int                              `json:"nivel"`
	Activa                 bool                             `json:"activa"`
	OrdenVisualizacion     *int                             `json:"orden_visualizacion,omitempty"`
	ImagenURL              *string                          `json:"imagen_url,omitempty"`
	PathCompleto           *string                          `json:"path_completo,omitempty"`
	TotalProductos         int                              `json:"total_productos"`
	ConfiguracionEtiquetas *ConfiguracionEtiquetasCategoria `json:"configuracion_etiquetas,omitempty"`
	FechaCreacion          time.Time                        `json:"fecha_creacion"`
	FechaModificacion      time.Time                        `json:"fecha_modificacion"`
	CategoriaPadre         *CategoriaProductoListDTO        `json:"categoria_padre,omitempty"`
	Subcategorias          []CategoriaProductoListDTO       `json:"subcategorias,omitempty"`
}

type CategoriaProductoListDTO struct {
	ID                 uuid.UUID `json:"id"`
	Codigo             string    `json:"codigo"`
	Nombre             string    `json:"nombre"`
	Nivel              int       `json:"nivel"`
	Activa             bool      `json:"activa"`
	OrdenVisualizacion *int      `json:"orden_visualizacion,omitempty"`
	TotalProductos     int       `json:"total_productos"`
	FechaCreacion      time.Time `json:"fecha_creacion"`
	FechaModificacion  time.Time `json:"fecha_modificacion"`
}

// DTOs para CodigoBarraAdicional

type CodigoBarraAdicionalDTO struct {
	ID          uuid.UUID `json:"id"`
	CodigoBarra string    `json:"codigo_barra"`
	Descripcion *string   `json:"descripcion,omitempty"`
	Activo      bool      `json:"activo"`
	TipoCodigo  string    `json:"tipo_codigo"`
	Validado    bool      `json:"validado"`
}

// Filtros para búsqueda

type ProductoFilter struct {
	PaginationFilter
	SortFilter
	CodigoInterno   *string    `json:"codigo_interno,omitempty" form:"codigo_interno"`
	CodigoBarra     *string    `json:"codigo_barra,omitempty" form:"codigo_barra"`
	Descripcion     *string    `json:"descripcion,omitempty" form:"descripcion"`
	Marca           *string    `json:"marca,omitempty" form:"marca"`
	CategoriaID     *uuid.UUID `json:"categoria_id,omitempty" form:"categoria_id"`
	Activo          *bool      `json:"activo,omitempty" form:"activo"`
	PrecioMin       *float64   `json:"precio_min,omitempty" form:"precio_min"`
	PrecioMax       *float64   `json:"precio_max,omitempty" form:"precio_max"`
	StockBajo       *bool      `json:"stock_bajo,omitempty" form:"stock_bajo"`
	RequiereSerie   *bool      `json:"requiere_serie,omitempty" form:"requiere_serie"`
	TerminoBusqueda *string    `json:"termino_busqueda,omitempty" form:"q"`
}

type CategoriaProductoFilter struct {
	PaginationFilter
	SortFilter
	Codigo           *string    `json:"codigo,omitempty" form:"codigo"`
	Nombre           *string    `json:"nombre,omitempty" form:"nombre"`
	CategoriaPadreID *uuid.UUID `json:"categoria_padre_id,omitempty" form:"categoria_padre_id"`
	Nivel            *int       `json:"nivel,omitempty" form:"nivel"`
	Activa           *bool      `json:"activa,omitempty" form:"activa"`
}

// Métodos helper

func (p *Producto) ToResponseDTO() ProductoResponseDTO {
	dto := ProductoResponseDTO{
		ID:                       p.ID,
		CodigoInterno:            p.CodigoInterno,
		CodigoBarra:              p.CodigoBarra,
		Descripcion:              p.Descripcion,
		DescripcionCorta:         p.DescripcionCorta,
		CategoriaID:              p.CategoriaID,
		Marca:                    p.Marca,
		Modelo:                   p.Modelo,
		PrecioUnitario:           p.PrecioUnitario,
		PrecioCosto:              p.PrecioCosto,
		UnidadMedida:             p.UnidadMedida,
		Peso:                     p.Peso,
		Dimensiones:              p.Dimensiones,
		EspecificacionesTecnicas: p.EspecificacionesTecnicas,
		Activo:                   p.Activo,
		RequiereSerie:            p.RequiereSerie,
		PermiteFraccionamiento:   p.PermiteFraccionamiento,
		StockMinimo:              p.StockMinimo,
		StockMaximo:              p.StockMaximo,
		ImagenPrincipalURL:       p.ImagenPrincipalURL,
		ImagenesAdicionales:      p.ImagenesAdicionales,
		PopularidadScore:         p.PopularidadScore,
		ConfiguracionEtiqueta:    p.ConfiguracionEtiqueta,
		FechaUltimaEtiqueta:      p.FechaUltimaEtiqueta,
		TotalEtiquetasGeneradas:  p.TotalEtiquetasGeneradas,
		FechaCreacion:            p.FechaCreacion,
		FechaModificacion:        p.FechaModificacion,
	}

	if p.Categoria != nil {
		categoriaDTO := p.Categoria.ToListDTO()
		dto.Categoria = &categoriaDTO
	}

	// Convertir códigos de barra adicionales
	for _, codigo := range p.CodigosBarraAdicionales {
		dto.CodigosBarraAdicionales = append(dto.CodigosBarraAdicionales, CodigoBarraAdicionalDTO{
			ID:          codigo.ID,
			CodigoBarra: codigo.CodigoBarra,
			Descripcion: codigo.Descripcion,
			Activo:      codigo.Activo,
			TipoCodigo:  codigo.TipoCodigo,
			Validado:    codigo.Validado,
		})
	}

	return dto
}

func (p *Producto) ToListDTO() ProductoListDTO {
	dto := ProductoListDTO{
		ID:                p.ID,
		CodigoInterno:     p.CodigoInterno,
		CodigoBarra:       p.CodigoBarra,
		Descripcion:       p.Descripcion,
		DescripcionCorta:  p.DescripcionCorta,
		Marca:             p.Marca,
		Modelo:            p.Modelo,
		PrecioUnitario:    p.PrecioUnitario,
		UnidadMedida:      p.UnidadMedida,
		Activo:            p.Activo,
		PopularidadScore:  p.PopularidadScore,
		FechaCreacion:     p.FechaCreacion,
		FechaModificacion: p.FechaModificacion,
	}

	if p.Categoria != nil {
		dto.CategoriaNombre = &p.Categoria.Nombre
	}

	return dto
}

func (c *CategoriaProducto) ToResponseDTO() CategoriaProductoResponseDTO {
	dto := CategoriaProductoResponseDTO{
		ID:                     c.ID,
		Codigo:                 c.Codigo,
		Nombre:                 c.Nombre,
		Descripcion:            c.Descripcion,
		CategoriaPadreID:       c.CategoriaPadreID,
		Nivel:                  c.Nivel,
		Activa:                 c.Activa,
		OrdenVisualizacion:     c.OrdenVisualizacion,
		ImagenURL:              c.ImagenURL,
		PathCompleto:           c.PathCompleto,
		TotalProductos:         c.TotalProductos,
		ConfiguracionEtiquetas: c.ConfiguracionEtiquetas,
		FechaCreacion:          c.FechaCreacion,
		FechaModificacion:      c.FechaModificacion,
	}

	if c.CategoriaPadre != nil {
		padreDTO := c.CategoriaPadre.ToListDTO()
		dto.CategoriaPadre = &padreDTO
	}

	for _, sub := range c.Subcategorias {
		dto.Subcategorias = append(dto.Subcategorias, sub.ToListDTO())
	}

	return dto
}

func (c *CategoriaProducto) ToListDTO() CategoriaProductoListDTO {
	return CategoriaProductoListDTO{
		ID:                 c.ID,
		Codigo:             c.Codigo,
		Nombre:             c.Nombre,
		Nivel:              c.Nivel,
		Activa:             c.Activa,
		OrdenVisualizacion: c.OrdenVisualizacion,
		TotalProductos:     c.TotalProductos,
		FechaCreacion:      c.FechaCreacion,
		FechaModificacion:  c.FechaModificacion,
	}
}

func (dto *ProductoCreateDTO) ToModel(usuarioID uuid.UUID) *Producto {
	producto := &Producto{
		CodigoInterno:            dto.CodigoInterno,
		CodigoBarra:              dto.CodigoBarra,
		Descripcion:              dto.Descripcion,
		DescripcionCorta:         dto.DescripcionCorta,
		CategoriaID:              dto.CategoriaID,
		Marca:                    dto.Marca,
		Modelo:                   dto.Modelo,
		PrecioUnitario:           dto.PrecioUnitario,
		PrecioCosto:              dto.PrecioCosto,
		UnidadMedida:             dto.UnidadMedida,
		Peso:                     dto.Peso,
		Dimensiones:              dto.Dimensiones,
		EspecificacionesTecnicas: dto.EspecificacionesTecnicas,
		Activo:                   true,
		RequiereSerie:            dto.RequiereSerie,
		PermiteFraccionamiento:   dto.PermiteFraccionamiento,
		StockMinimo:              dto.StockMinimo,
		StockMaximo:              dto.StockMaximo,
		ImagenPrincipalURL:       dto.ImagenPrincipalURL,
		ImagenesAdicionales:      dto.ImagenesAdicionales,
		ConfiguracionEtiqueta:    dto.ConfiguracionEtiqueta,
		PopularidadScore:         0,
		TotalEtiquetasGeneradas:  0,
	}

	producto.BaseModelWithUser.UsuarioCreacion = &usuarioID

	if producto.UnidadMedida == "" {
		producto.UnidadMedida = "UN"
	}

	return producto
}

// Validaciones personalizadas

func (p *Producto) Validate() error {
	if p.PrecioUnitario < 0 {
		return fmt.Errorf("precio unitario debe ser mayor o igual a 0")
	}

	if p.PrecioCosto != nil && *p.PrecioCosto < 0 {
		return fmt.Errorf("precio costo debe ser mayor o igual a 0")
	}

	if p.StockMinimo < 0 {
		return fmt.Errorf("stock mínimo debe ser mayor o igual a 0")
	}

	if p.StockMaximo != nil && *p.StockMaximo < p.StockMinimo {
		return fmt.Errorf("stock máximo debe ser mayor o igual al stock mínimo")
	}

	if p.PopularidadScore < 0 || p.PopularidadScore > 100 {
		return fmt.Errorf("score de popularidad debe estar entre 0 y 100")
	}

	return nil
}

func (c *CategoriaProducto) Validate() error {
	if c.Nivel <= 0 {
		return fmt.Errorf("nivel debe ser mayor a 0")
	}

	return nil
}
