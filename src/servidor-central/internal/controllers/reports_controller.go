package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"ferre-pos-servidor-central/internal/models"
	"ferre-pos-servidor-central/internal/services"
)

// ReportsController maneja las operaciones del API de Reportes
type ReportsController struct {
	BaseController
	reporteService services.ReporteService
}

// NewReportsController crea una nueva instancia del controlador de reportes
func NewReportsController(reporteService services.ReporteService) *ReportsController {
	return &ReportsController{
		reporteService: reporteService,
	}
}

// ===== PLANTILLAS DE REPORTES =====

// GetPlantillasReportes obtiene lista de plantillas de reportes
func (rc *ReportsController) GetPlantillasReportes(c *gin.Context) {
	user, err := rc.GetUserFromContext(c)
	if err != nil {
		rc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	sucursalID, _ := rc.GetSucursalFromContext(c)

	// Parsear filtros
	filter := models.PlantillaReporteFilter{
		PaginationFilter: rc.ParsePagination(c),
		SortFilter:       rc.ParseSort(c),
		SucursalID:       sucursalID,
	}

	// Filtros específicos
	if nombre := c.Query("nombre"); nombre != "" {
		filter.Nombre = &nombre
	}
	if tipoReporte := c.Query("tipo_reporte"); tipoReporte != "" {
		tipo := models.TipoReporte(tipoReporte)
		filter.TipoReporte = &tipo
	}
	if categoria := c.Query("categoria_reporte"); categoria != "" {
		cat := models.CategoriaReporte(categoria)
		filter.CategoriaReporte = &cat
	}
	if activa := c.Query("activa"); activa != "" {
		activaBool := activa == "true"
		filter.Activa = &activaBool
	}
	if esPublica := c.Query("es_publica"); esPublica != "" {
		publicaBool := esPublica == "true"
		filter.EsPublica = &publicaBool
	}

	// Filtrar por rol del usuario
	filter.RolUsuario = &user.Rol

	plantillas, pagination, err := rc.reporteService.GetPlantillasReportes(filter)
	if err != nil {
		rc.ResponseError(c, http.StatusInternalServerError, "Error al obtener plantillas", err)
		return
	}

	rc.SetCacheHeaders(c, 300) // Cache por 5 minutos
	rc.ResponsePaginated(c, plantillas, pagination)
}

// GetPlantillaReporte obtiene una plantilla específica
func (rc *ReportsController) GetPlantillaReporte(c *gin.Context) {
	plantillaID, err := rc.ParseUUID(c, "id")
	if err != nil {
		rc.ResponseError(c, http.StatusBadRequest, "ID de plantilla inválido", err)
		return
	}

	user, err := rc.GetUserFromContext(c)
	if err != nil {
		rc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	sucursalID, _ := rc.GetSucursalFromContext(c)

	plantilla, err := rc.reporteService.GetPlantillaReporte(plantillaID, user.Rol, sucursalID)
	if err != nil {
		rc.ResponseError(c, http.StatusNotFound, "Plantilla no encontrada", err)
		return
	}

	rc.SetCacheHeaders(c, 600) // Cache por 10 minutos
	rc.ResponseSuccess(c, plantilla)
}

// CrearPlantillaReporte crea una nueva plantilla de reporte
func (rc *ReportsController) CrearPlantillaReporte(c *gin.Context) {
	var plantillaDTO models.PlantillaReporteCreateDTO
	if err := rc.ValidateJSON(c, &plantillaDTO); err != nil {
		rc.ResponseError(c, http.StatusBadRequest, "Datos de plantilla inválidos", err)
		return
	}

	user, err := rc.GetUserFromContext(c)
	if err != nil {
		rc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	// Verificar permisos para crear plantillas
	if !rc.CheckPermission(user, "crear_plantillas_reportes") {
		rc.ResponseError(c, http.StatusForbidden, "Sin permisos para crear plantillas", nil)
		return
	}

	plantilla, err := rc.reporteService.CrearPlantillaReporte(plantillaDTO, user.ID)
	if err != nil {
		rc.ResponseError(c, http.StatusBadRequest, "Error al crear plantilla", err)
		return
	}

	rc.LogActivity(c, "crear_plantilla_reporte", plantilla)
	rc.ResponseCreated(c, plantilla)
}

// ActualizarPlantillaReporte actualiza una plantilla existente
func (rc *ReportsController) ActualizarPlantillaReporte(c *gin.Context) {
	plantillaID, err := rc.ParseUUID(c, "id")
	if err != nil {
		rc.ResponseError(c, http.StatusBadRequest, "ID de plantilla inválido", err)
		return
	}

	var plantillaDTO models.PlantillaReporteUpdateDTO
	if err := rc.ValidateJSON(c, &plantillaDTO); err != nil {
		rc.ResponseError(c, http.StatusBadRequest, "Datos de plantilla inválidos", err)
		return
	}

	user, err := rc.GetUserFromContext(c)
	if err != nil {
		rc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	// Verificar permisos para editar plantillas
	if !rc.CheckPermission(user, "editar_plantillas_reportes") {
		rc.ResponseError(c, http.StatusForbidden, "Sin permisos para editar plantillas", nil)
		return
	}

	plantilla, err := rc.reporteService.ActualizarPlantillaReporte(plantillaID, plantillaDTO, user.ID)
	if err != nil {
		rc.ResponseError(c, http.StatusBadRequest, "Error al actualizar plantilla", err)
		return
	}

	rc.LogActivity(c, "actualizar_plantilla_reporte", plantilla)
	rc.ResponseSuccess(c, plantilla)
}

// EliminarPlantillaReporte elimina una plantilla
func (rc *ReportsController) EliminarPlantillaReporte(c *gin.Context) {
	plantillaID, err := rc.ParseUUID(c, "id")
	if err != nil {
		rc.ResponseError(c, http.StatusBadRequest, "ID de plantilla inválido", err)
		return
	}

	user, err := rc.GetUserFromContext(c)
	if err != nil {
		rc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	// Verificar permisos para eliminar plantillas
	if !rc.CheckPermission(user, "eliminar_plantillas_reportes") {
		rc.ResponseError(c, http.StatusForbidden, "Sin permisos para eliminar plantillas", nil)
		return
	}

	if err := rc.reporteService.EliminarPlantillaReporte(plantillaID, user.ID); err != nil {
		rc.ResponseError(c, http.StatusBadRequest, "Error al eliminar plantilla", err)
		return
	}

	rc.LogActivity(c, "eliminar_plantilla_reporte", gin.H{"plantilla_id": plantillaID})
	rc.ResponseSuccess(c, gin.H{"message": "Plantilla eliminada exitosamente"})
}

// ValidarPlantillaReporte valida una plantilla de reporte
func (rc *ReportsController) ValidarPlantillaReporte(c *gin.Context) {
	var plantillaDTO models.PlantillaReporteCreateDTO
	if err := rc.ValidateJSON(c, &plantillaDTO); err != nil {
		rc.ResponseError(c, http.StatusBadRequest, "Datos de plantilla inválidos", err)
		return
	}

	errores, err := rc.reporteService.ValidarPlantillaReporte(plantillaDTO)
	if err != nil {
		rc.ResponseError(c, http.StatusInternalServerError, "Error al validar plantilla", err)
		return
	}

	if len(errores) > 0 {
		rc.ResponseValidationError(c, errores)
		return
	}

	rc.ResponseSuccess(c, gin.H{
		"valida": true,
		"message": "Plantilla válida",
	})
}

// ===== GENERACIÓN DE REPORTES =====

// GenerarReporte genera un nuevo reporte
func (rc *ReportsController) GenerarReporte(c *gin.Context) {
	var reporteDTO models.ReporteGeneradoCreateDTO
	if err := rc.ValidateJSON(c, &reporteDTO); err != nil {
		rc.ResponseError(c, http.StatusBadRequest, "Datos de reporte inválidos", err)
		return
	}

	user, err := rc.GetUserFromContext(c)
	if err != nil {
		rc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	sucursalID, _ := rc.GetSucursalFromContext(c)

	reporte, err := rc.reporteService.GenerarReporte(reporteDTO, user.ID, sucursalID)
	if err != nil {
		rc.ResponseError(c, http.StatusBadRequest, "Error al generar reporte", err)
		return
	}

	rc.LogActivity(c, "generar_reporte", reporte)
	rc.ResponseCreated(c, reporte)
}

// GetReportesGenerados obtiene lista de reportes generados
func (rc *ReportsController) GetReportesGenerados(c *gin.Context) {
	user, err := rc.GetUserFromContext(c)
	if err != nil {
		rc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	sucursalID, _ := rc.GetSucursalFromContext(c)

	// Parsear filtros
	filter := models.ReporteGeneradoFilter{
		PaginationFilter: rc.ParsePagination(c),
		SortFilter:       rc.ParseSort(c),
		DateRangeFilter:  rc.ParseDateRange(c),
		SucursalID:       sucursalID,
	}

	// Filtros específicos
	if plantillaID := c.Query("plantilla_id"); plantillaID != "" {
		if id, err := uuid.Parse(plantillaID); err == nil {
			filter.PlantillaID = &id
		}
	}
	if estado := c.Query("estado"); estado != "" {
		estadoReporte := models.EstadoReporte(estado)
		filter.Estado = &estadoReporte
	}
	if formato := c.Query("formato_salida"); formato != "" {
		formatoReporte := models.FormatoReporte(formato)
		filter.FormatoSalida = &formatoReporte
	}
	if esPublico := c.Query("es_publico"); esPublico != "" {
		publicoBool := esPublico == "true"
		filter.EsPublico = &publicoBool
	}
	if nombre := c.Query("nombre"); nombre != "" {
		filter.Nombre = &nombre
	}

	// Solo mostrar reportes del usuario si no es admin/gerente
	if !rc.CheckPermission(user, "ver_todos_reportes") {
		filter.UsuarioID = &user.ID
	}

	reportes, pagination, err := rc.reporteService.GetReportesGenerados(filter)
	if err != nil {
		rc.ResponseError(c, http.StatusInternalServerError, "Error al obtener reportes", err)
		return
	}

	rc.ResponsePaginated(c, reportes, pagination)
}

// GetReporteGenerado obtiene un reporte específico
func (rc *ReportsController) GetReporteGenerado(c *gin.Context) {
	reporteID, err := rc.ParseUUID(c, "id")
	if err != nil {
		rc.ResponseError(c, http.StatusBadRequest, "ID de reporte inválido", err)
		return
	}

	user, err := rc.GetUserFromContext(c)
	if err != nil {
		rc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	reporte, err := rc.reporteService.GetReporteGenerado(reporteID, user.ID)
	if err != nil {
		rc.ResponseError(c, http.StatusNotFound, "Reporte no encontrado", err)
		return
	}

	rc.ResponseSuccess(c, reporte)
}

// DescargarReporte descarga el archivo de un reporte
func (rc *ReportsController) DescargarReporte(c *gin.Context) {
	reporteID, err := rc.ParseUUID(c, "id")
	if err != nil {
		rc.ResponseError(c, http.StatusBadRequest, "ID de reporte inválido", err)
		return
	}

	user, err := rc.GetUserFromContext(c)
	if err != nil {
		rc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	archivo, contentType, filename, err := rc.reporteService.DescargarReporte(reporteID, user.ID)
	if err != nil {
		rc.ResponseError(c, http.StatusNotFound, "Archivo no encontrado", err)
		return
	}

	rc.LogActivity(c, "descargar_reporte", gin.H{"reporte_id": reporteID})

	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, contentType, archivo)
}

// CancelarReporte cancela un reporte en proceso
func (rc *ReportsController) CancelarReporte(c *gin.Context) {
	reporteID, err := rc.ParseUUID(c, "id")
	if err != nil {
		rc.ResponseError(c, http.StatusBadRequest, "ID de reporte inválido", err)
		return
	}

	user, err := rc.GetUserFromContext(c)
	if err != nil {
		rc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	if err := rc.reporteService.CancelarReporte(reporteID, user.ID); err != nil {
		rc.ResponseError(c, http.StatusBadRequest, "Error al cancelar reporte", err)
		return
	}

	rc.LogActivity(c, "cancelar_reporte", gin.H{"reporte_id": reporteID})
	rc.ResponseSuccess(c, gin.H{"message": "Reporte cancelado exitosamente"})
}

// CompartirReporte comparte un reporte con otros usuarios
func (rc *ReportsController) CompartirReporte(c *gin.Context) {
	reporteID, err := rc.ParseUUID(c, "id")
	if err != nil {
		rc.ResponseError(c, http.StatusBadRequest, "ID de reporte inválido", err)
		return
	}

	var compartirDTO models.CompartirReporteDTO
	if err := rc.ValidateJSON(c, &compartirDTO); err != nil {
		rc.ResponseError(c, http.StatusBadRequest, "Datos de compartir inválidos", err)
		return
	}

	user, err := rc.GetUserFromContext(c)
	if err != nil {
		rc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	if err := rc.reporteService.CompartirReporte(reporteID, compartirDTO, user.ID); err != nil {
		rc.ResponseError(c, http.StatusBadRequest, "Error al compartir reporte", err)
		return
	}

	rc.LogActivity(c, "compartir_reporte", gin.H{
		"reporte_id": reporteID,
		"usuarios":   compartirDTO.UsuariosIDs,
	})

	rc.ResponseSuccess(c, gin.H{"message": "Reporte compartido exitosamente"})
}

// ===== REPORTES PROGRAMADOS =====

// GetReportesProgramados obtiene lista de reportes programados
func (rc *ReportsController) GetReportesProgramados(c *gin.Context) {
	user, err := rc.GetUserFromContext(c)
	if err != nil {
		rc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	sucursalID, _ := rc.GetSucursalFromContext(c)

	// Parsear filtros
	filter := models.ReporteProgramadoFilter{
		PaginationFilter: rc.ParsePagination(c),
		SortFilter:       rc.ParseSort(c),
		SucursalID:       sucursalID,
	}

	// Filtros específicos
	if activo := c.Query("activo"); activo != "" {
		activoBool := activo == "true"
		filter.Activo = &activoBool
	}

	// Solo mostrar reportes programados del usuario si no es admin/gerente
	if !rc.CheckPermission(user, "ver_todos_reportes_programados") {
		filter.UsuarioID = &user.ID
	}

	reportes, pagination, err := rc.reporteService.GetReportesProgramados(filter)
	if err != nil {
		rc.ResponseError(c, http.StatusInternalServerError, "Error al obtener reportes programados", err)
		return
	}

	rc.ResponsePaginated(c, reportes, pagination)
}

// CrearReporteProgramado crea un nuevo reporte programado
func (rc *ReportsController) CrearReporteProgramado(c *gin.Context) {
	var reporteDTO models.ReporteProgramadoCreateDTO
	if err := rc.ValidateJSON(c, &reporteDTO); err != nil {
		rc.ResponseError(c, http.StatusBadRequest, "Datos de reporte programado inválidos", err)
		return
	}

	user, err := rc.GetUserFromContext(c)
	if err != nil {
		rc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	// Verificar permisos para crear reportes programados
	if !rc.CheckPermission(user, "crear_reportes_programados") {
		rc.ResponseError(c, http.StatusForbidden, "Sin permisos para crear reportes programados", nil)
		return
	}

	sucursalID, _ := rc.GetSucursalFromContext(c)

	reporte, err := rc.reporteService.CrearReporteProgramado(reporteDTO, user.ID, sucursalID)
	if err != nil {
		rc.ResponseError(c, http.StatusBadRequest, "Error al crear reporte programado", err)
		return
	}

	rc.LogActivity(c, "crear_reporte_programado", reporte)
	rc.ResponseCreated(c, reporte)
}

// ActualizarReporteProgramado actualiza un reporte programado
func (rc *ReportsController) ActualizarReporteProgramado(c *gin.Context) {
	reporteID, err := rc.ParseUUID(c, "id")
	if err != nil {
		rc.ResponseError(c, http.StatusBadRequest, "ID de reporte inválido", err)
		return
	}

	var reporteDTO models.ReporteProgramadoUpdateDTO
	if err := rc.ValidateJSON(c, &reporteDTO); err != nil {
		rc.ResponseError(c, http.StatusBadRequest, "Datos de reporte programado inválidos", err)
		return
	}

	user, err := rc.GetUserFromContext(c)
	if err != nil {
		rc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	// Verificar permisos para editar reportes programados
	if !rc.CheckPermission(user, "editar_reportes_programados") {
		rc.ResponseError(c, http.StatusForbidden, "Sin permisos para editar reportes programados", nil)
		return
	}

	reporte, err := rc.reporteService.ActualizarReporteProgramado(reporteID, reporteDTO, user.ID)
	if err != nil {
		rc.ResponseError(c, http.StatusBadRequest, "Error al actualizar reporte programado", err)
		return
	}

	rc.LogActivity(c, "actualizar_reporte_programado", reporte)
	rc.ResponseSuccess(c, reporte)
}

// EliminarReporteProgramado elimina un reporte programado
func (rc *ReportsController) EliminarReporteProgramado(c *gin.Context) {
	reporteID, err := rc.ParseUUID(c, "id")
	if err != nil {
		rc.ResponseError(c, http.StatusBadRequest, "ID de reporte inválido", err)
		return
	}

	user, err := rc.GetUserFromContext(c)
	if err != nil {
		rc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	// Verificar permisos para eliminar reportes programados
	if !rc.CheckPermission(user, "eliminar_reportes_programados") {
		rc.ResponseError(c, http.StatusForbidden, "Sin permisos para eliminar reportes programados", nil)
		return
	}

	if err := rc.reporteService.EliminarReporteProgramado(reporteID, user.ID); err != nil {
		rc.ResponseError(c, http.StatusBadRequest, "Error al eliminar reporte programado", err)
		return
	}

	rc.LogActivity(c, "eliminar_reporte_programado", gin.H{"reporte_id": reporteID})
	rc.ResponseSuccess(c, gin.H{"message": "Reporte programado eliminado exitosamente"})
}

// ===== DASHBOARDS =====

// GetDashboards obtiene lista de dashboards
func (rc *ReportsController) GetDashboards(c *gin.Context) {
	user, err := rc.GetUserFromContext(c)
	if err != nil {
		rc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	sucursalID, _ := rc.GetSucursalFromContext(c)

	// Parsear filtros
	filter := models.DashboardFilter{
		PaginationFilter: rc.ParsePagination(c),
		SortFilter:       rc.ParseSort(c),
		SucursalID:       sucursalID,
	}

	// Filtros específicos
	if esPublico := c.Query("es_publico"); esPublico != "" {
		publicoBool := esPublico == "true"
		filter.EsPublico = &publicoBool
	}

	// Solo mostrar dashboards del usuario si no es admin/gerente
	if !rc.CheckPermission(user, "ver_todos_dashboards") {
		filter.UsuarioID = &user.ID
	}

	dashboards, pagination, err := rc.reporteService.GetDashboards(filter)
	if err != nil {
		rc.ResponseError(c, http.StatusInternalServerError, "Error al obtener dashboards", err)
		return
	}

	rc.ResponsePaginated(c, dashboards, pagination)
}

// GetDashboard obtiene un dashboard específico
func (rc *ReportsController) GetDashboard(c *gin.Context) {
	dashboardID, err := rc.ParseUUID(c, "id")
	if err != nil {
		rc.ResponseError(c, http.StatusBadRequest, "ID de dashboard inválido", err)
		return
	}

	user, err := rc.GetUserFromContext(c)
	if err != nil {
		rc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	dashboard, err := rc.reporteService.GetDashboard(dashboardID, user.ID)
	if err != nil {
		rc.ResponseError(c, http.StatusNotFound, "Dashboard no encontrado", err)
		return
	}

	// Actualizar contador de visualizaciones
	go rc.reporteService.ActualizarVisualizacionDashboard(dashboardID)

	rc.ResponseSuccess(c, dashboard)
}

// CrearDashboard crea un nuevo dashboard
func (rc *ReportsController) CrearDashboard(c *gin.Context) {
	var dashboardDTO models.DashboardCreateDTO
	if err := rc.ValidateJSON(c, &dashboardDTO); err != nil {
		rc.ResponseError(c, http.StatusBadRequest, "Datos de dashboard inválidos", err)
		return
	}

	user, err := rc.GetUserFromContext(c)
	if err != nil {
		rc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	// Verificar permisos para crear dashboards
	if !rc.CheckPermission(user, "crear_dashboards") {
		rc.ResponseError(c, http.StatusForbidden, "Sin permisos para crear dashboards", nil)
		return
	}

	sucursalID, _ := rc.GetSucursalFromContext(c)

	dashboard, err := rc.reporteService.CrearDashboard(dashboardDTO, user.ID, sucursalID)
	if err != nil {
		rc.ResponseError(c, http.StatusBadRequest, "Error al crear dashboard", err)
		return
	}

	rc.LogActivity(c, "crear_dashboard", dashboard)
	rc.ResponseCreated(c, dashboard)
}

// ActualizarDashboard actualiza un dashboard existente
func (rc *ReportsController) ActualizarDashboard(c *gin.Context) {
	dashboardID, err := rc.ParseUUID(c, "id")
	if err != nil {
		rc.ResponseError(c, http.StatusBadRequest, "ID de dashboard inválido", err)
		return
	}

	var dashboardDTO models.DashboardUpdateDTO
	if err := rc.ValidateJSON(c, &dashboardDTO); err != nil {
		rc.ResponseError(c, http.StatusBadRequest, "Datos de dashboard inválidos", err)
		return
	}

	user, err := rc.GetUserFromContext(c)
	if err != nil {
		rc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	// Verificar permisos para editar dashboards
	if !rc.CheckPermission(user, "editar_dashboards") {
		rc.ResponseError(c, http.StatusForbidden, "Sin permisos para editar dashboards", nil)
		return
	}

	dashboard, err := rc.reporteService.ActualizarDashboard(dashboardID, dashboardDTO, user.ID)
	if err != nil {
		rc.ResponseError(c, http.StatusBadRequest, "Error al actualizar dashboard", err)
		return
	}

	rc.LogActivity(c, "actualizar_dashboard", dashboard)
	rc.ResponseSuccess(c, dashboard)
}

// EliminarDashboard elimina un dashboard
func (rc *ReportsController) EliminarDashboard(c *gin.Context) {
	dashboardID, err := rc.ParseUUID(c, "id")
	if err != nil {
		rc.ResponseError(c, http.StatusBadRequest, "ID de dashboard inválido", err)
		return
	}

	user, err := rc.GetUserFromContext(c)
	if err != nil {
		rc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	// Verificar permisos para eliminar dashboards
	if !rc.CheckPermission(user, "eliminar_dashboards") {
		rc.ResponseError(c, http.StatusForbidden, "Sin permisos para eliminar dashboards", nil)
		return
	}

	if err := rc.reporteService.EliminarDashboard(dashboardID, user.ID); err != nil {
		rc.ResponseError(c, http.StatusBadRequest, "Error al eliminar dashboard", err)
		return
	}

	rc.LogActivity(c, "eliminar_dashboard", gin.H{"dashboard_id": dashboardID})
	rc.ResponseSuccess(c, gin.H{"message": "Dashboard eliminado exitosamente"})
}

// ===== CONFIGURACIÓN Y UTILIDADES =====

// GetTiposReportes obtiene los tipos de reportes disponibles
func (rc *ReportsController) GetTiposReportes(c *gin.Context) {
	tipos := []gin.H{
		{"value": "ventas", "label": "Ventas", "descripcion": "Reportes de ventas y facturación"},
		{"value": "stock", "label": "Stock", "descripcion": "Reportes de inventario y stock"},
		{"value": "productos", "label": "Productos", "descripcion": "Reportes de productos y categorías"},
		{"value": "clientes", "label": "Clientes", "descripcion": "Reportes de clientes y fidelización"},
		{"value": "fidelizacion", "label": "Fidelización", "descripcion": "Reportes de programas de fidelización"},
		{"value": "financiero", "label": "Financiero", "descripcion": "Reportes financieros y contables"},
		{"value": "operacional", "label": "Operacional", "descripcion": "Reportes operacionales y de gestión"},
		{"value": "personalizado", "label": "Personalizado", "descripcion": "Reportes personalizados"},
	}

	rc.SetCacheHeaders(c, 3600) // Cache por 1 hora
	rc.ResponseSuccess(c, tipos)
}

// GetCategoriasReportes obtiene las categorías de reportes disponibles
func (rc *ReportsController) GetCategoriasReportes(c *gin.Context) {
	categorias := []gin.H{
		{"value": "comercial", "label": "Comercial", "descripcion": "Reportes comerciales y de ventas"},
		{"value": "inventario", "label": "Inventario", "descripcion": "Reportes de inventario y stock"},
		{"value": "finanzas", "label": "Finanzas", "descripcion": "Reportes financieros y contables"},
		{"value": "operaciones", "label": "Operaciones", "descripcion": "Reportes operacionales"},
		{"value": "marketing", "label": "Marketing", "descripcion": "Reportes de marketing y promociones"},
		{"value": "rrhh", "label": "RRHH", "descripcion": "Reportes de recursos humanos"},
		{"value": "auditoria", "label": "Auditoría", "descripcion": "Reportes de auditoría y control"},
	}

	rc.SetCacheHeaders(c, 3600) // Cache por 1 hora
	rc.ResponseSuccess(c, categorias)
}

// GetFormatosReportes obtiene los formatos de reportes disponibles
func (rc *ReportsController) GetFormatosReportes(c *gin.Context) {
	formatos := []gin.H{
		{"value": "pdf", "label": "PDF", "descripcion": "Formato PDF para impresión"},
		{"value": "excel", "label": "Excel", "descripcion": "Formato Excel para análisis"},
		{"value": "csv", "label": "CSV", "descripcion": "Formato CSV para datos"},
		{"value": "json", "label": "JSON", "descripcion": "Formato JSON para APIs"},
		{"value": "html", "label": "HTML", "descripcion": "Formato HTML para web"},
	}

	rc.SetCacheHeaders(c, 3600) // Cache por 1 hora
	rc.ResponseSuccess(c, formatos)
}

// GetEstadisticasReportes obtiene estadísticas de reportes
func (rc *ReportsController) GetEstadisticasReportes(c *gin.Context) {
	user, err := rc.GetUserFromContext(c)
	if err != nil {
		rc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	sucursalID, _ := rc.GetSucursalFromContext(c)

	periodo := c.Query("periodo") // "dia", "semana", "mes", "año"
	if periodo == "" {
		periodo = "mes"
	}

	estadisticas, err := rc.reporteService.GetEstadisticasReportes(user.ID, sucursalID, periodo)
	if err != nil {
		rc.ResponseError(c, http.StatusInternalServerError, "Error al obtener estadísticas", err)
		return
	}

	rc.SetCacheHeaders(c, 300) // Cache por 5 minutos
	rc.ResponseSuccess(c, estadisticas)
}

// PreviewReporte genera una vista previa de un reporte
func (rc *ReportsController) PreviewReporte(c *gin.Context) {
	plantillaID, err := rc.ParseUUID(c, "plantilla_id")
	if err != nil {
		rc.ResponseError(c, http.StatusBadRequest, "ID de plantilla inválido", err)
		return
	}

	var previewDTO models.PreviewReporteDTO
	if err := rc.ValidateJSON(c, &previewDTO); err != nil {
		rc.ResponseError(c, http.StatusBadRequest, "Datos de preview inválidos", err)
		return
	}

	user, err := rc.GetUserFromContext(c)
	if err != nil {
		rc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	sucursalID, _ := rc.GetSucursalFromContext(c)

	preview, contentType, err := rc.reporteService.GenerarPreviewReporte(plantillaID, previewDTO, user.ID, sucursalID)
	if err != nil {
		rc.ResponseError(c, http.StatusBadRequest, "Error al generar preview", err)
		return
	}

	c.Header("Content-Type", contentType)
	c.Data(http.StatusOK, contentType, preview)
}

// ValidarParametrosReporte valida los parámetros de un reporte
func (rc *ReportsController) ValidarParametrosReporte(c *gin.Context) {
	plantillaID, err := rc.ParseUUID(c, "plantilla_id")
	if err != nil {
		rc.ResponseError(c, http.StatusBadRequest, "ID de plantilla inválido", err)
		return
	}

	var parametrosDTO models.ParametrosReporte
	if err := rc.ValidateJSON(c, &parametrosDTO); err != nil {
		rc.ResponseError(c, http.StatusBadRequest, "Parámetros inválidos", err)
		return
	}

	errores, err := rc.reporteService.ValidarParametrosReporte(plantillaID, parametrosDTO)
	if err != nil {
		rc.ResponseError(c, http.StatusInternalServerError, "Error al validar parámetros", err)
		return
	}

	if len(errores) > 0 {
		rc.ResponseValidationError(c, errores)
		return
	}

	rc.ResponseSuccess(c, gin.H{
		"validos": true,
		"message": "Parámetros válidos",
	})
}

// ExportarDashboard exporta un dashboard
func (rc *ReportsController) ExportarDashboard(c *gin.Context) {
	dashboardID, err := rc.ParseUUID(c, "id")
	if err != nil {
		rc.ResponseError(c, http.StatusBadRequest, "ID de dashboard inválido", err)
		return
	}

	formato := c.Query("formato") // "pdf", "png", "json"
	if formato == "" {
		formato = "pdf"
	}

	user, err := rc.GetUserFromContext(c)
	if err != nil {
		rc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	archivo, contentType, filename, err := rc.reporteService.ExportarDashboard(dashboardID, formato, user.ID)
	if err != nil {
		rc.ResponseError(c, http.StatusBadRequest, "Error al exportar dashboard", err)
		return
	}

	rc.LogActivity(c, "exportar_dashboard", gin.H{
		"dashboard_id": dashboardID,
		"formato":      formato,
	})

	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, contentType, archivo)
}

// GetDatosWidget obtiene datos para un widget específico
func (rc *ReportsController) GetDatosWidget(c *gin.Context) {
	dashboardID, err := rc.ParseUUID(c, "dashboard_id")
	if err != nil {
		rc.ResponseError(c, http.StatusBadRequest, "ID de dashboard inválido", err)
		return
	}

	widgetID := c.Param("widget_id")
	if widgetID == "" {
		rc.ResponseError(c, http.StatusBadRequest, "ID de widget requerido", nil)
		return
	}

	user, err := rc.GetUserFromContext(c)
	if err != nil {
		rc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	// Parámetros opcionales para filtros
	filtros := make(map[string]interface{})
	for key, values := range c.Request.URL.Query() {
		if len(values) > 0 {
			filtros[key] = values[0]
		}
	}

	datos, err := rc.reporteService.GetDatosWidget(dashboardID, widgetID, filtros, user.ID)
	if err != nil {
		rc.ResponseError(c, http.StatusBadRequest, "Error al obtener datos del widget", err)
		return
	}

	rc.SetCacheHeaders(c, 60) // Cache por 1 minuto
	rc.ResponseSuccess(c, datos)
}

