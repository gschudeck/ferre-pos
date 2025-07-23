package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"ferre-pos-servidor-central/internal/models"
	"ferre-pos-servidor-central/internal/services"
)

// LabelsController maneja las operaciones del API de Etiquetas
type LabelsController struct {
	BaseController
	etiquetaService services.EtiquetaService
	productoService services.ProductoService
}

// NewLabelsController crea una nueva instancia del controlador de etiquetas
func NewLabelsController(
	etiquetaService services.EtiquetaService,
	productoService services.ProductoService,
) *LabelsController {
	return &LabelsController{
		etiquetaService: etiquetaService,
		productoService: productoService,
	}
}

// ===== PLANTILLAS DE ETIQUETAS =====

// GetPlantillas obtiene lista de plantillas de etiquetas
func (lc *LabelsController) GetPlantillas(c *gin.Context) {
	user, err := lc.GetUserFromContext(c)
	if err != nil {
		lc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	sucursalID, _ := lc.GetSucursalFromContext(c)

	// Parsear filtros
	filter := models.PlantillaEtiquetaFilter{
		PaginationFilter: lc.ParsePagination(c),
		SortFilter:       lc.ParseSort(c),
		SucursalID:       sucursalID,
	}

	// Filtros específicos
	if nombre := c.Query("nombre"); nombre != "" {
		filter.Nombre = &nombre
	}
	if tipoEtiqueta := c.Query("tipo_etiqueta"); tipoEtiqueta != "" {
		tipo := models.TipoEtiqueta(tipoEtiqueta)
		filter.TipoEtiqueta = &tipo
	}
	if activa := c.Query("activa"); activa != "" {
		activaBool := activa == "true"
		filter.Activa = &activaBool
	}

	// Verificar permisos por sucursal
	if !lc.CheckPermission(user, "ver_todas_plantillas") && sucursalID != nil {
		filter.SucursalID = sucursalID
	}

	plantillas, pagination, err := lc.etiquetaService.GetPlantillas(filter)
	if err != nil {
		lc.ResponseError(c, http.StatusInternalServerError, "Error al obtener plantillas", err)
		return
	}

	lc.SetCacheHeaders(c, 300) // Cache por 5 minutos
	lc.ResponsePaginated(c, plantillas, pagination)
}

// GetPlantilla obtiene una plantilla específica
func (lc *LabelsController) GetPlantilla(c *gin.Context) {
	plantillaID, err := lc.ParseUUID(c, "id")
	if err != nil {
		lc.ResponseError(c, http.StatusBadRequest, "ID de plantilla inválido", err)
		return
	}

	user, err := lc.GetUserFromContext(c)
	if err != nil {
		lc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	sucursalID, _ := lc.GetSucursalFromContext(c)

	plantilla, err := lc.etiquetaService.GetPlantilla(plantillaID, user.ID, sucursalID)
	if err != nil {
		lc.ResponseError(c, http.StatusNotFound, "Plantilla no encontrada", err)
		return
	}

	lc.SetCacheHeaders(c, 600) // Cache por 10 minutos
	lc.ResponseSuccess(c, plantilla)
}

// CrearPlantilla crea una nueva plantilla de etiqueta
func (lc *LabelsController) CrearPlantilla(c *gin.Context) {
	var plantillaDTO models.PlantillaEtiquetaCreateDTO
	if err := lc.ValidateJSON(c, &plantillaDTO); err != nil {
		lc.ResponseError(c, http.StatusBadRequest, "Datos de plantilla inválidos", err)
		return
	}

	user, err := lc.GetUserFromContext(c)
	if err != nil {
		lc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	// Verificar permisos para crear plantillas
	if !lc.CheckPermission(user, "crear_plantillas") {
		lc.ResponseError(c, http.StatusForbidden, "Sin permisos para crear plantillas", nil)
		return
	}

	sucursalID, _ := lc.GetSucursalFromContext(c)

	plantilla, err := lc.etiquetaService.CrearPlantilla(plantillaDTO, user.ID, sucursalID)
	if err != nil {
		lc.ResponseError(c, http.StatusBadRequest, "Error al crear plantilla", err)
		return
	}

	lc.LogActivity(c, "crear_plantilla", plantilla)
	lc.ResponseCreated(c, plantilla)
}

// ActualizarPlantilla actualiza una plantilla existente
func (lc *LabelsController) ActualizarPlantilla(c *gin.Context) {
	plantillaID, err := lc.ParseUUID(c, "id")
	if err != nil {
		lc.ResponseError(c, http.StatusBadRequest, "ID de plantilla inválido", err)
		return
	}

	var plantillaDTO models.PlantillaEtiquetaUpdateDTO
	if err := lc.ValidateJSON(c, &plantillaDTO); err != nil {
		lc.ResponseError(c, http.StatusBadRequest, "Datos de plantilla inválidos", err)
		return
	}

	user, err := lc.GetUserFromContext(c)
	if err != nil {
		lc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	// Verificar permisos para editar plantillas
	if !lc.CheckPermission(user, "editar_plantillas") {
		lc.ResponseError(c, http.StatusForbidden, "Sin permisos para editar plantillas", nil)
		return
	}

	plantilla, err := lc.etiquetaService.ActualizarPlantilla(plantillaID, plantillaDTO, user.ID)
	if err != nil {
		lc.ResponseError(c, http.StatusBadRequest, "Error al actualizar plantilla", err)
		return
	}

	lc.LogActivity(c, "actualizar_plantilla", plantilla)
	lc.ResponseSuccess(c, plantilla)
}

// EliminarPlantilla elimina una plantilla
func (lc *LabelsController) EliminarPlantilla(c *gin.Context) {
	plantillaID, err := lc.ParseUUID(c, "id")
	if err != nil {
		lc.ResponseError(c, http.StatusBadRequest, "ID de plantilla inválido", err)
		return
	}

	user, err := lc.GetUserFromContext(c)
	if err != nil {
		lc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	// Verificar permisos para eliminar plantillas
	if !lc.CheckPermission(user, "eliminar_plantillas") {
		lc.ResponseError(c, http.StatusForbidden, "Sin permisos para eliminar plantillas", nil)
		return
	}

	if err := lc.etiquetaService.EliminarPlantilla(plantillaID, user.ID); err != nil {
		lc.ResponseError(c, http.StatusBadRequest, "Error al eliminar plantilla", err)
		return
	}

	lc.LogActivity(c, "eliminar_plantilla", gin.H{"plantilla_id": plantillaID})
	lc.ResponseSuccess(c, gin.H{"message": "Plantilla eliminada exitosamente"})
}

// DuplicarPlantilla duplica una plantilla existente
func (lc *LabelsController) DuplicarPlantilla(c *gin.Context) {
	plantillaID, err := lc.ParseUUID(c, "id")
	if err != nil {
		lc.ResponseError(c, http.StatusBadRequest, "ID de plantilla inválido", err)
		return
	}

	var duplicarDTO models.DuplicarPlantillaDTO
	if err := lc.ValidateJSON(c, &duplicarDTO); err != nil {
		lc.ResponseError(c, http.StatusBadRequest, "Datos de duplicación inválidos", err)
		return
	}

	user, err := lc.GetUserFromContext(c)
	if err != nil {
		lc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	// Verificar permisos para crear plantillas
	if !lc.CheckPermission(user, "crear_plantillas") {
		lc.ResponseError(c, http.StatusForbidden, "Sin permisos para duplicar plantillas", nil)
		return
	}

	plantilla, err := lc.etiquetaService.DuplicarPlantilla(plantillaID, duplicarDTO, user.ID)
	if err != nil {
		lc.ResponseError(c, http.StatusBadRequest, "Error al duplicar plantilla", err)
		return
	}

	lc.LogActivity(c, "duplicar_plantilla", plantilla)
	lc.ResponseCreated(c, plantilla)
}

// ===== GENERACIÓN DE ETIQUETAS =====

// GenerarEtiqueta genera una etiqueta individual
func (lc *LabelsController) GenerarEtiqueta(c *gin.Context) {
	var generarDTO models.GenerarEtiquetaDTO
	if err := lc.ValidateJSON(c, &generarDTO); err != nil {
		lc.ResponseError(c, http.StatusBadRequest, "Datos de generación inválidos", err)
		return
	}

	user, err := lc.GetUserFromContext(c)
	if err != nil {
		lc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	sucursalID, err := lc.GetSucursalFromContext(c)
	if err != nil {
		lc.ResponseError(c, http.StatusBadRequest, "Sucursal requerida", err)
		return
	}

	etiqueta, err := lc.etiquetaService.GenerarEtiqueta(generarDTO, user.ID, sucursalID)
	if err != nil {
		lc.ResponseError(c, http.StatusBadRequest, "Error al generar etiqueta", err)
		return
	}

	lc.LogActivity(c, "generar_etiqueta", etiqueta)
	lc.ResponseCreated(c, etiqueta)
}

// GenerarLoteEtiquetas genera un lote de etiquetas
func (lc *LabelsController) GenerarLoteEtiquetas(c *gin.Context) {
	var loteDTO models.GenerarLoteEtiquetasDTO
	if err := lc.ValidateJSON(c, &loteDTO); err != nil {
		lc.ResponseError(c, http.StatusBadRequest, "Datos de lote inválidos", err)
		return
	}

	user, err := lc.GetUserFromContext(c)
	if err != nil {
		lc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	sucursalID, err := lc.GetSucursalFromContext(c)
	if err != nil {
		lc.ResponseError(c, http.StatusBadRequest, "Sucursal requerida", err)
		return
	}

	lote, err := lc.etiquetaService.GenerarLoteEtiquetas(loteDTO, user.ID, sucursalID)
	if err != nil {
		lc.ResponseError(c, http.StatusBadRequest, "Error al generar lote", err)
		return
	}

	lc.LogActivity(c, "generar_lote_etiquetas", lote)
	lc.ResponseCreated(c, lote)
}

// GetEtiquetasGeneradas obtiene lista de etiquetas generadas
func (lc *LabelsController) GetEtiquetasGeneradas(c *gin.Context) {
	user, err := lc.GetUserFromContext(c)
	if err != nil {
		lc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	sucursalID, _ := lc.GetSucursalFromContext(c)

	// Parsear filtros
	filter := models.EtiquetaGeneradaFilter{
		PaginationFilter: lc.ParsePagination(c),
		SortFilter:       lc.ParseSort(c),
		DateRangeFilter:  lc.ParseDateRange(c),
		SucursalID:       sucursalID,
	}

	// Filtros específicos
	if plantillaID := c.Query("plantilla_id"); plantillaID != "" {
		if id, err := uuid.Parse(plantillaID); err == nil {
			filter.PlantillaID = &id
		}
	}
	if productoID := c.Query("producto_id"); productoID != "" {
		if id, err := uuid.Parse(productoID); err == nil {
			filter.ProductoID = &id
		}
	}
	if estado := c.Query("estado"); estado != "" {
		estadoEtiqueta := models.EstadoEtiqueta(estado)
		filter.Estado = &estadoEtiqueta
	}

	// Solo mostrar etiquetas del usuario si no es admin/gerente
	if !lc.CheckPermission(user, "ver_todas_etiquetas") {
		filter.UsuarioID = &user.ID
	}

	etiquetas, pagination, err := lc.etiquetaService.GetEtiquetasGeneradas(filter)
	if err != nil {
		lc.ResponseError(c, http.StatusInternalServerError, "Error al obtener etiquetas", err)
		return
	}

	lc.ResponsePaginated(c, etiquetas, pagination)
}

// GetEtiqueta obtiene una etiqueta específica
func (lc *LabelsController) GetEtiqueta(c *gin.Context) {
	etiquetaID, err := lc.ParseUUID(c, "id")
	if err != nil {
		lc.ResponseError(c, http.StatusBadRequest, "ID de etiqueta inválido", err)
		return
	}

	user, err := lc.GetUserFromContext(c)
	if err != nil {
		lc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	etiqueta, err := lc.etiquetaService.GetEtiqueta(etiquetaID, user.ID)
	if err != nil {
		lc.ResponseError(c, http.StatusNotFound, "Etiqueta no encontrada", err)
		return
	}

	lc.ResponseSuccess(c, etiqueta)
}

// DescargarEtiqueta descarga el archivo de una etiqueta
func (lc *LabelsController) DescargarEtiqueta(c *gin.Context) {
	etiquetaID, err := lc.ParseUUID(c, "id")
	if err != nil {
		lc.ResponseError(c, http.StatusBadRequest, "ID de etiqueta inválido", err)
		return
	}

	user, err := lc.GetUserFromContext(c)
	if err != nil {
		lc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	archivo, contentType, err := lc.etiquetaService.DescargarEtiqueta(etiquetaID, user.ID)
	if err != nil {
		lc.ResponseError(c, http.StatusNotFound, "Archivo no encontrado", err)
		return
	}

	lc.LogActivity(c, "descargar_etiqueta", gin.H{"etiqueta_id": etiquetaID})

	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", "attachment; filename=etiqueta_"+etiquetaID.String()+".pdf")
	c.Data(http.StatusOK, contentType, archivo)
}

// ===== LOTES DE ETIQUETAS =====

// GetLotesEtiquetas obtiene lista de lotes de etiquetas
func (lc *LabelsController) GetLotesEtiquetas(c *gin.Context) {
	user, err := lc.GetUserFromContext(c)
	if err != nil {
		lc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	sucursalID, _ := lc.GetSucursalFromContext(c)

	// Parsear filtros
	filter := models.LoteEtiquetasFilter{
		PaginationFilter: lc.ParsePagination(c),
		SortFilter:       lc.ParseSort(c),
		DateRangeFilter:  lc.ParseDateRange(c),
		SucursalID:       sucursalID,
	}

	// Filtros específicos
	if estado := c.Query("estado"); estado != "" {
		estadoLote := models.EstadoLote(estado)
		filter.Estado = &estadoLote
	}

	// Solo mostrar lotes del usuario si no es admin/gerente
	if !lc.CheckPermission(user, "ver_todos_lotes") {
		filter.UsuarioID = &user.ID
	}

	lotes, pagination, err := lc.etiquetaService.GetLotesEtiquetas(filter)
	if err != nil {
		lc.ResponseError(c, http.StatusInternalServerError, "Error al obtener lotes", err)
		return
	}

	lc.ResponsePaginated(c, lotes, pagination)
}

// GetLoteEtiquetas obtiene un lote específico
func (lc *LabelsController) GetLoteEtiquetas(c *gin.Context) {
	loteID, err := lc.ParseUUID(c, "id")
	if err != nil {
		lc.ResponseError(c, http.StatusBadRequest, "ID de lote inválido", err)
		return
	}

	user, err := lc.GetUserFromContext(c)
	if err != nil {
		lc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	lote, err := lc.etiquetaService.GetLoteEtiquetas(loteID, user.ID)
	if err != nil {
		lc.ResponseError(c, http.StatusNotFound, "Lote no encontrado", err)
		return
	}

	lc.ResponseSuccess(c, lote)
}

// DescargarLoteEtiquetas descarga el archivo consolidado de un lote
func (lc *LabelsController) DescargarLoteEtiquetas(c *gin.Context) {
	loteID, err := lc.ParseUUID(c, "id")
	if err != nil {
		lc.ResponseError(c, http.StatusBadRequest, "ID de lote inválido", err)
		return
	}

	user, err := lc.GetUserFromContext(c)
	if err != nil {
		lc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	formato := c.Query("formato") // "pdf" o "zip"
	if formato == "" {
		formato = "pdf"
	}

	archivo, contentType, filename, err := lc.etiquetaService.DescargarLoteEtiquetas(loteID, formato, user.ID)
	if err != nil {
		lc.ResponseError(c, http.StatusNotFound, "Archivo no encontrado", err)
		return
	}

	lc.LogActivity(c, "descargar_lote", gin.H{
		"lote_id": loteID,
		"formato": formato,
	})

	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, contentType, archivo)
}

// CancelarLote cancela un lote en proceso
func (lc *LabelsController) CancelarLote(c *gin.Context) {
	loteID, err := lc.ParseUUID(c, "id")
	if err != nil {
		lc.ResponseError(c, http.StatusBadRequest, "ID de lote inválido", err)
		return
	}

	user, err := lc.GetUserFromContext(c)
	if err != nil {
		lc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	if err := lc.etiquetaService.CancelarLote(loteID, user.ID); err != nil {
		lc.ResponseError(c, http.StatusBadRequest, "Error al cancelar lote", err)
		return
	}

	lc.LogActivity(c, "cancelar_lote", gin.H{"lote_id": loteID})
	lc.ResponseSuccess(c, gin.H{"message": "Lote cancelado exitosamente"})
}

// ===== PRODUCTOS Y ETIQUETAS =====

// GetProductosParaEtiquetas obtiene productos disponibles para generar etiquetas
func (lc *LabelsController) GetProductosParaEtiquetas(c *gin.Context) {
	sucursalID, err := lc.GetSucursalFromContext(c)
	if err != nil {
		lc.ResponseError(c, http.StatusBadRequest, "Sucursal requerida", err)
		return
	}

	// Parsear filtros
	filter := models.ProductoFilter{
		PaginationFilter: lc.ParsePagination(c),
		SortFilter:       lc.ParseSort(c),
		SucursalID:       sucursalID,
		Activo:           &[]bool{true}[0], // Solo productos activos
	}

	// Filtros específicos
	if categoria := c.Query("categoria"); categoria != "" {
		filter.Categoria = &categoria
	}
	if nombre := c.Query("nombre"); nombre != "" {
		filter.Nombre = &nombre
	}
	if codigoBarra := c.Query("codigo_barra"); codigoBarra != "" {
		filter.CodigoBarra = &codigoBarra
	}

	productos, pagination, err := lc.productoService.GetProductosParaEtiquetas(filter)
	if err != nil {
		lc.ResponseError(c, http.StatusInternalServerError, "Error al obtener productos", err)
		return
	}

	lc.SetCacheHeaders(c, 300) // Cache por 5 minutos
	lc.ResponsePaginated(c, productos, pagination)
}

// BuscarProductosParaEtiquetas busca productos para generar etiquetas
func (lc *LabelsController) BuscarProductosParaEtiquetas(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		lc.ResponseError(c, http.StatusBadRequest, "Parámetro de búsqueda requerido", nil)
		return
	}

	sucursalID, err := lc.GetSucursalFromContext(c)
	if err != nil {
		lc.ResponseError(c, http.StatusBadRequest, "Sucursal requerida", err)
		return
	}

	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 50 {
			limit = l
		}
	}

	productos, err := lc.productoService.BuscarProductosParaEtiquetas(query, sucursalID, limit)
	if err != nil {
		lc.ResponseError(c, http.StatusInternalServerError, "Error en búsqueda", err)
		return
	}

	lc.SetCacheHeaders(c, 60) // Cache por 1 minuto
	lc.ResponseSuccess(c, productos)
}

// ===== CONFIGURACIÓN Y UTILIDADES =====

// GetTiposEtiquetas obtiene los tipos de etiquetas disponibles
func (lc *LabelsController) GetTiposEtiquetas(c *gin.Context) {
	tipos := []gin.H{
		{"value": "producto", "label": "Producto", "descripcion": "Etiqueta básica de producto"},
		{"value": "precio", "label": "Precio", "descripcion": "Etiqueta de precio"},
		{"value": "promocion", "label": "Promoción", "descripcion": "Etiqueta promocional"},
		{"value": "inventario", "label": "Inventario", "descripcion": "Etiqueta para inventario"},
		{"value": "gondola", "label": "Góndola", "descripcion": "Etiqueta para góndola"},
		{"value": "oferta", "label": "Oferta", "descripcion": "Etiqueta de oferta especial"},
	}

	lc.SetCacheHeaders(c, 3600) // Cache por 1 hora
	lc.ResponseSuccess(c, tipos)
}

// GetTamañosEtiquetas obtiene los tamaños de etiquetas disponibles
func (lc *LabelsController) GetTamañosEtiquetas(c *gin.Context) {
	tamaños := []gin.H{
		{"value": "pequeña", "label": "Pequeña", "ancho": 50, "alto": 30, "unidad": "mm"},
		{"value": "mediana", "label": "Mediana", "ancho": 70, "alto": 40, "unidad": "mm"},
		{"value": "grande", "label": "Grande", "ancho": 100, "alto": 60, "unidad": "mm"},
		{"value": "gondola", "label": "Góndola", "ancho": 150, "alto": 100, "unidad": "mm"},
		{"value": "personalizada", "label": "Personalizada", "ancho": 0, "alto": 0, "unidad": "mm"},
	}

	lc.SetCacheHeaders(c, 3600) // Cache por 1 hora
	lc.ResponseSuccess(c, tamaños)
}

// ValidarPlantilla valida una plantilla de etiqueta
func (lc *LabelsController) ValidarPlantilla(c *gin.Context) {
	var plantillaDTO models.PlantillaEtiquetaCreateDTO
	if err := lc.ValidateJSON(c, &plantillaDTO); err != nil {
		lc.ResponseError(c, http.StatusBadRequest, "Datos de plantilla inválidos", err)
		return
	}

	errores, err := lc.etiquetaService.ValidarPlantilla(plantillaDTO)
	if err != nil {
		lc.ResponseError(c, http.StatusInternalServerError, "Error al validar plantilla", err)
		return
	}

	if len(errores) > 0 {
		lc.ResponseValidationError(c, errores)
		return
	}

	lc.ResponseSuccess(c, gin.H{
		"valida": true,
		"message": "Plantilla válida",
	})
}

// PreviewPlantilla genera una vista previa de la plantilla
func (lc *LabelsController) PreviewPlantilla(c *gin.Context) {
	plantillaID, err := lc.ParseUUID(c, "id")
	if err != nil {
		lc.ResponseError(c, http.StatusBadRequest, "ID de plantilla inválido", err)
		return
	}

	var previewDTO models.PreviewPlantillaDTO
	if err := lc.ValidateJSON(c, &previewDTO); err != nil {
		lc.ResponseError(c, http.StatusBadRequest, "Datos de preview inválidos", err)
		return
	}

	user, err := lc.GetUserFromContext(c)
	if err != nil {
		lc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	preview, contentType, err := lc.etiquetaService.GenerarPreviewPlantilla(plantillaID, previewDTO, user.ID)
	if err != nil {
		lc.ResponseError(c, http.StatusBadRequest, "Error al generar preview", err)
		return
	}

	c.Header("Content-Type", contentType)
	c.Data(http.StatusOK, contentType, preview)
}

// GetEstadisticasEtiquetas obtiene estadísticas de etiquetas
func (lc *LabelsController) GetEstadisticasEtiquetas(c *gin.Context) {
	user, err := lc.GetUserFromContext(c)
	if err != nil {
		lc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	sucursalID, _ := lc.GetSucursalFromContext(c)

	periodo := c.Query("periodo") // "dia", "semana", "mes", "año"
	if periodo == "" {
		periodo = "mes"
	}

	estadisticas, err := lc.etiquetaService.GetEstadisticasEtiquetas(user.ID, sucursalID, periodo)
	if err != nil {
		lc.ResponseError(c, http.StatusInternalServerError, "Error al obtener estadísticas", err)
		return
	}

	lc.SetCacheHeaders(c, 300) // Cache por 5 minutos
	lc.ResponseSuccess(c, estadisticas)
}

