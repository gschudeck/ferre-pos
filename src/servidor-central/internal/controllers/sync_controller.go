package controllers

import (
	"net/http"
	"strconv"

	"ferre-pos-servidor-central/internal/models"
	"ferre-pos-servidor-central/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// SyncController maneja las operaciones del API de Sincronización
type SyncController struct {
	BaseController
	sincronizacionService services.SincronizacionService
	sucursalService       services.SucursalService
}

// NewSyncController crea una nueva instancia del controlador de sincronización
func NewSyncController(
	sincronizacionService services.SincronizacionService,
	sucursalService services.SucursalService,
) *SyncController {
	return &SyncController{
		sincronizacionService: sincronizacionService,
		sucursalService:       sucursalService,
	}
}

// ===== ESTADO DE SINCRONIZACIÓN =====

// GetEstadoSincronizacion obtiene el estado de sincronización de una sucursal
func (sc *SyncController) GetEstadoSincronizacion(c *gin.Context) {
	sucursalID, err := sc.ParseUUID(c, "sucursal_id")
	if err != nil {
		sc.ResponseError(c, http.StatusBadRequest, "ID de sucursal inválido", err)
		return
	}

	user, err := sc.GetUserFromContext(c)
	if err != nil {
		sc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	// Verificar permisos para ver estado de sincronización
	if !sc.CheckPermission(user, "ver_sincronizacion") {
		sc.ResponseError(c, http.StatusForbidden, "Sin permisos para ver sincronización", nil)
		return
	}

	estado, err := sc.sincronizacionService.GetEstadoSincronizacion(sucursalID)
	if err != nil {
		sc.ResponseError(c, http.StatusNotFound, "Estado de sincronización no encontrado", err)
		return
	}

	sc.SetCacheHeaders(c, 60) // Cache por 1 minuto
	sc.ResponseSuccess(c, estado)
}

// GetEstadosSincronizacion obtiene el estado de sincronización de todas las sucursales
func (sc *SyncController) GetEstadosSincronizacion(c *gin.Context) {
	user, err := sc.GetUserFromContext(c)
	if err != nil {
		sc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	// Verificar permisos para ver todos los estados
	if !sc.CheckPermission(user, "ver_todas_sincronizaciones") {
		sc.ResponseError(c, http.StatusForbidden, "Sin permisos para ver todas las sincronizaciones", nil)
		return
	}

	// Parsear filtros
	filter := models.SincronizacionSucursalFilter{
		PaginationFilter: sc.ParsePagination(c),
		SortFilter:       sc.ParseSort(c),
	}

	// Filtros específicos
	if estado := c.Query("estado"); estado != "" {
		estadoSync := models.EstadoSincronizacion(estado)
		filter.Estado = &estadoSync
	}
	if conErrores := c.Query("con_errores"); conErrores == "true" {
		filter.ConErrores = &[]bool{true}[0]
	}
	if sinSincronizar := c.Query("sin_sincronizar"); sinSincronizar == "true" {
		filter.SinSincronizar = &[]bool{true}[0]
	}

	estados, pagination, err := sc.sincronizacionService.GetEstadosSincronizacion(filter)
	if err != nil {
		sc.ResponseError(c, http.StatusInternalServerError, "Error al obtener estados", err)
		return
	}

	sc.SetCacheHeaders(c, 60) // Cache por 1 minuto
	sc.ResponsePaginated(c, estados, pagination)
}

// ActualizarConfiguracionSincronizacion actualiza la configuración de sincronización
func (sc *SyncController) ActualizarConfiguracionSincronizacion(c *gin.Context) {
	sucursalID, err := sc.ParseUUID(c, "sucursal_id")
	if err != nil {
		sc.ResponseError(c, http.StatusBadRequest, "ID de sucursal inválido", err)
		return
	}

	var configDTO models.SincronizacionSucursalUpdateDTO
	if err := sc.ValidateJSON(c, &configDTO); err != nil {
		sc.ResponseError(c, http.StatusBadRequest, "Datos de configuración inválidos", err)
		return
	}

	user, err := sc.GetUserFromContext(c)
	if err != nil {
		sc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	// Verificar permisos para configurar sincronización
	if !sc.CheckPermission(user, "configurar_sincronizacion") {
		sc.ResponseError(c, http.StatusForbidden, "Sin permisos para configurar sincronización", nil)
		return
	}

	estado, err := sc.sincronizacionService.ActualizarConfiguracion(sucursalID, configDTO, user.ID)
	if err != nil {
		sc.ResponseError(c, http.StatusBadRequest, "Error al actualizar configuración", err)
		return
	}

	sc.LogActivity(c, "actualizar_config_sync", gin.H{
		"sucursal_id": sucursalID,
		"config":      configDTO,
	})

	sc.ResponseSuccess(c, estado)
}

// ===== OPERACIONES DE SINCRONIZACIÓN =====

// IniciarSincronizacion inicia una sincronización manual
func (sc *SyncController) IniciarSincronizacion(c *gin.Context) {
	sucursalID, err := sc.ParseUUID(c, "sucursal_id")
	if err != nil {
		sc.ResponseError(c, http.StatusBadRequest, "ID de sucursal inválido", err)
		return
	}

	var iniciarDTO models.IniciarSincronizacionDTO
	if err := sc.ValidateJSON(c, &iniciarDTO); err != nil {
		sc.ResponseError(c, http.StatusBadRequest, "Datos de sincronización inválidos", err)
		return
	}

	user, err := sc.GetUserFromContext(c)
	if err != nil {
		sc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	// Verificar permisos para iniciar sincronización
	if !sc.CheckPermission(user, "iniciar_sincronizacion") {
		sc.ResponseError(c, http.StatusForbidden, "Sin permisos para iniciar sincronización", nil)
		return
	}

	resultado, err := sc.sincronizacionService.IniciarSincronizacion(sucursalID, iniciarDTO, user.ID)
	if err != nil {
		sc.ResponseError(c, http.StatusBadRequest, "Error al iniciar sincronización", err)
		return
	}

	sc.LogActivity(c, "iniciar_sincronizacion", gin.H{
		"sucursal_id": sucursalID,
		"tipo":        iniciarDTO.TipoSincronizacion,
	})

	sc.ResponseSuccess(c, resultado)
}

// DetenerSincronizacion detiene una sincronización en curso
func (sc *SyncController) DetenerSincronizacion(c *gin.Context) {
	sucursalID, err := sc.ParseUUID(c, "sucursal_id")
	if err != nil {
		sc.ResponseError(c, http.StatusBadRequest, "ID de sucursal inválido", err)
		return
	}

	user, err := sc.GetUserFromContext(c)
	if err != nil {
		sc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	// Verificar permisos para detener sincronización
	if !sc.CheckPermission(user, "detener_sincronizacion") {
		sc.ResponseError(c, http.StatusForbidden, "Sin permisos para detener sincronización", nil)
		return
	}

	if err := sc.sincronizacionService.DetenerSincronizacion(sucursalID, user.ID); err != nil {
		sc.ResponseError(c, http.StatusBadRequest, "Error al detener sincronización", err)
		return
	}

	sc.LogActivity(c, "detener_sincronizacion", gin.H{"sucursal_id": sucursalID})
	sc.ResponseSuccess(c, gin.H{"message": "Sincronización detenida exitosamente"})
}

// ReiniciarSincronizacion reinicia una sincronización fallida
func (sc *SyncController) ReiniciarSincronizacion(c *gin.Context) {
	sucursalID, err := sc.ParseUUID(c, "sucursal_id")
	if err != nil {
		sc.ResponseError(c, http.StatusBadRequest, "ID de sucursal inválido", err)
		return
	}

	user, err := sc.GetUserFromContext(c)
	if err != nil {
		sc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	// Verificar permisos para reiniciar sincronización
	if !sc.CheckPermission(user, "reiniciar_sincronizacion") {
		sc.ResponseError(c, http.StatusForbidden, "Sin permisos para reiniciar sincronización", nil)
		return
	}

	resultado, err := sc.sincronizacionService.ReiniciarSincronizacion(sucursalID, user.ID)
	if err != nil {
		sc.ResponseError(c, http.StatusBadRequest, "Error al reiniciar sincronización", err)
		return
	}

	sc.LogActivity(c, "reiniciar_sincronizacion", gin.H{"sucursal_id": sucursalID})
	sc.ResponseSuccess(c, resultado)
}

// ===== LOGS DE SINCRONIZACIÓN =====

// GetLogsSincronizacion obtiene los logs de sincronización
func (sc *SyncController) GetLogsSincronizacion(c *gin.Context) {
	user, err := sc.GetUserFromContext(c)
	if err != nil {
		sc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	// Verificar permisos para ver logs
	if !sc.CheckPermission(user, "ver_logs_sincronizacion") {
		sc.ResponseError(c, http.StatusForbidden, "Sin permisos para ver logs", nil)
		return
	}

	// Parsear filtros
	filter := models.LogSincronizacionFilter{
		PaginationFilter: sc.ParsePagination(c),
		SortFilter:       sc.ParseSort(c),
		DateRangeFilter:  sc.ParseDateRange(c),
	}

	// Filtros específicos
	if sucursalID := c.Query("sucursal_id"); sucursalID != "" {
		if id, err := uuid.Parse(sucursalID); err == nil {
			filter.SucursalID = &id
		}
	}
	if tipoOperacion := c.Query("tipo_operacion"); tipoOperacion != "" {
		tipo := models.TipoOperacionSync(tipoOperacion)
		filter.TipoOperacion = &tipo
	}
	if entidad := c.Query("entidad_afectada"); entidad != "" {
		filter.EntidadAfectada = &entidad
	}
	if accion := c.Query("accion"); accion != "" {
		accionSync := models.AccionSincronizacion(accion)
		filter.Accion = &accionSync
	}
	if estado := c.Query("estado"); estado != "" {
		estadoOp := models.EstadoOperacionSync(estado)
		filter.Estado = &estadoOp
	}
	if conErrores := c.Query("con_errores"); conErrores == "true" {
		filter.ConErrores = &[]bool{true}[0]
	}

	logs, pagination, err := sc.sincronizacionService.GetLogsSincronizacion(filter)
	if err != nil {
		sc.ResponseError(c, http.StatusInternalServerError, "Error al obtener logs", err)
		return
	}

	sc.ResponsePaginated(c, logs, pagination)
}

// GetLogSincronizacion obtiene un log específico
func (sc *SyncController) GetLogSincronizacion(c *gin.Context) {
	logID, err := sc.ParseUUID(c, "id")
	if err != nil {
		sc.ResponseError(c, http.StatusBadRequest, "ID de log inválido", err)
		return
	}

	user, err := sc.GetUserFromContext(c)
	if err != nil {
		sc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	// Verificar permisos para ver logs
	if !sc.CheckPermission(user, "ver_logs_sincronizacion") {
		sc.ResponseError(c, http.StatusForbidden, "Sin permisos para ver logs", nil)
		return
	}

	log, err := sc.sincronizacionService.GetLogSincronizacion(logID)
	if err != nil {
		sc.ResponseError(c, http.StatusNotFound, "Log no encontrado", err)
		return
	}

	sc.ResponseSuccess(c, log)
}

// ReintentarOperacion reintenta una operación fallida
func (sc *SyncController) ReintentarOperacion(c *gin.Context) {
	logID, err := sc.ParseUUID(c, "id")
	if err != nil {
		sc.ResponseError(c, http.StatusBadRequest, "ID de log inválido", err)
		return
	}

	user, err := sc.GetUserFromContext(c)
	if err != nil {
		sc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	// Verificar permisos para reintentar operaciones
	if !sc.CheckPermission(user, "reintentar_operaciones") {
		sc.ResponseError(c, http.StatusForbidden, "Sin permisos para reintentar operaciones", nil)
		return
	}

	resultado, err := sc.sincronizacionService.ReintentarOperacion(logID, user.ID)
	if err != nil {
		sc.ResponseError(c, http.StatusBadRequest, "Error al reintentar operación", err)
		return
	}

	sc.LogActivity(c, "reintentar_operacion", gin.H{"log_id": logID})
	sc.ResponseSuccess(c, resultado)
}

// ===== CONFLICTOS DE SINCRONIZACIÓN =====

// GetConflictos obtiene los conflictos de sincronización
func (sc *SyncController) GetConflictos(c *gin.Context) {
	user, err := sc.GetUserFromContext(c)
	if err != nil {
		sc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	// Verificar permisos para ver conflictos
	if !sc.CheckPermission(user, "ver_conflictos") {
		sc.ResponseError(c, http.StatusForbidden, "Sin permisos para ver conflictos", nil)
		return
	}

	// Parsear filtros
	filter := models.ConflictoSincronizacionFilter{
		PaginationFilter: sc.ParsePagination(c),
		SortFilter:       sc.ParseSort(c),
		DateRangeFilter:  sc.ParseDateRange(c),
	}

	// Filtros específicos
	if sucursalID := c.Query("sucursal_id"); sucursalID != "" {
		if id, err := uuid.Parse(sucursalID); err == nil {
			filter.SucursalID = &id
		}
	}
	if entidad := c.Query("entidad_afectada"); entidad != "" {
		filter.EntidadAfectada = &entidad
	}
	if tipoConflicto := c.Query("tipo_conflicto"); tipoConflicto != "" {
		tipo := models.TipoConflicto(tipoConflicto)
		filter.TipoConflicto = &tipo
	}
	if estado := c.Query("estado_conflicto"); estado != "" {
		estadoConf := models.EstadoConflicto(estado)
		filter.EstadoConflicto = &estadoConf
	}
	if severidad := c.Query("severidad"); severidad != "" {
		sev := models.SeveridadConflicto(severidad)
		filter.Severidad = &sev
	}
	if sinResolver := c.Query("sin_resolver"); sinResolver == "true" {
		filter.SinResolver = &[]bool{true}[0]
	}

	conflictos, pagination, err := sc.sincronizacionService.GetConflictos(filter)
	if err != nil {
		sc.ResponseError(c, http.StatusInternalServerError, "Error al obtener conflictos", err)
		return
	}

	sc.ResponsePaginated(c, conflictos, pagination)
}

// GetConflicto obtiene un conflicto específico
func (sc *SyncController) GetConflicto(c *gin.Context) {
	conflictoID, err := sc.ParseUUID(c, "id")
	if err != nil {
		sc.ResponseError(c, http.StatusBadRequest, "ID de conflicto inválido", err)
		return
	}

	user, err := sc.GetUserFromContext(c)
	if err != nil {
		sc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	// Verificar permisos para ver conflictos
	if !sc.CheckPermission(user, "ver_conflictos") {
		sc.ResponseError(c, http.StatusForbidden, "Sin permisos para ver conflictos", nil)
		return
	}

	conflicto, err := sc.sincronizacionService.GetConflicto(conflictoID)
	if err != nil {
		sc.ResponseError(c, http.StatusNotFound, "Conflicto no encontrado", err)
		return
	}

	sc.ResponseSuccess(c, conflicto)
}

// ResolverConflicto resuelve un conflicto de sincronización
func (sc *SyncController) ResolverConflicto(c *gin.Context) {
	conflictoID, err := sc.ParseUUID(c, "id")
	if err != nil {
		sc.ResponseError(c, http.StatusBadRequest, "ID de conflicto inválido", err)
		return
	}

	var resolucionDTO models.ConflictoSincronizacionResolverDTO
	if err := sc.ValidateJSON(c, &resolucionDTO); err != nil {
		sc.ResponseError(c, http.StatusBadRequest, "Datos de resolución inválidos", err)
		return
	}

	user, err := sc.GetUserFromContext(c)
	if err != nil {
		sc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	// Verificar permisos para resolver conflictos
	if !sc.CheckPermission(user, "resolver_conflictos") {
		sc.ResponseError(c, http.StatusForbidden, "Sin permisos para resolver conflictos", nil)
		return
	}

	conflicto, err := sc.sincronizacionService.ResolverConflicto(conflictoID, resolucionDTO, user.ID)
	if err != nil {
		sc.ResponseError(c, http.StatusBadRequest, "Error al resolver conflicto", err)
		return
	}

	sc.LogActivity(c, "resolver_conflicto", gin.H{
		"conflicto_id":    conflictoID,
		"tipo_resolucion": resolucionDTO.TipoResolucion,
	})

	sc.ResponseSuccess(c, conflicto)
}

// IgnorarConflicto ignora un conflicto de sincronización
func (sc *SyncController) IgnorarConflicto(c *gin.Context) {
	conflictoID, err := sc.ParseUUID(c, "id")
	if err != nil {
		sc.ResponseError(c, http.StatusBadRequest, "ID de conflicto inválido", err)
		return
	}

	var ignorarDTO models.IgnorarConflictoDTO
	if err := sc.ValidateJSON(c, &ignorarDTO); err != nil {
		sc.ResponseError(c, http.StatusBadRequest, "Datos de ignorar inválidos", err)
		return
	}

	user, err := sc.GetUserFromContext(c)
	if err != nil {
		sc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	// Verificar permisos para ignorar conflictos
	if !sc.CheckPermission(user, "ignorar_conflictos") {
		sc.ResponseError(c, http.StatusForbidden, "Sin permisos para ignorar conflictos", nil)
		return
	}

	if err := sc.sincronizacionService.IgnorarConflicto(conflictoID, ignorarDTO, user.ID); err != nil {
		sc.ResponseError(c, http.StatusBadRequest, "Error al ignorar conflicto", err)
		return
	}

	sc.LogActivity(c, "ignorar_conflicto", gin.H{
		"conflicto_id": conflictoID,
		"motivo":       ignorarDTO.Motivo,
	})

	sc.ResponseSuccess(c, gin.H{"message": "Conflicto ignorado exitosamente"})
}

// ===== CONFIGURACIÓN GLOBAL =====

// GetConfiguracionGlobal obtiene la configuración global de sincronización
func (sc *SyncController) GetConfiguracionGlobal(c *gin.Context) {
	user, err := sc.GetUserFromContext(c)
	if err != nil {
		sc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	// Verificar permisos para ver configuración global
	if !sc.CheckPermission(user, "ver_config_global") {
		sc.ResponseError(c, http.StatusForbidden, "Sin permisos para ver configuración global", nil)
		return
	}

	configuracion, err := sc.sincronizacionService.GetConfiguracionGlobal()
	if err != nil {
		sc.ResponseError(c, http.StatusNotFound, "Configuración no encontrada", err)
		return
	}

	sc.SetCacheHeaders(c, 300) // Cache por 5 minutos
	sc.ResponseSuccess(c, configuracion)
}

// ActualizarConfiguracionGlobal actualiza la configuración global
func (sc *SyncController) ActualizarConfiguracionGlobal(c *gin.Context) {
	var configDTO models.ConfiguracionSincronizacionGlobalUpdateDTO
	if err := sc.ValidateJSON(c, &configDTO); err != nil {
		sc.ResponseError(c, http.StatusBadRequest, "Datos de configuración inválidos", err)
		return
	}

	user, err := sc.GetUserFromContext(c)
	if err != nil {
		sc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	// Verificar permisos para actualizar configuración global
	if !sc.CheckPermission(user, "actualizar_config_global") {
		sc.ResponseError(c, http.StatusForbidden, "Sin permisos para actualizar configuración global", nil)
		return
	}

	configuracion, err := sc.sincronizacionService.ActualizarConfiguracionGlobal(configDTO, user.ID)
	if err != nil {
		sc.ResponseError(c, http.StatusBadRequest, "Error al actualizar configuración", err)
		return
	}

	sc.LogActivity(c, "actualizar_config_global", configuracion)
	sc.ResponseSuccess(c, configuracion)
}

// ===== ESTADÍSTICAS Y MÉTRICAS =====

// GetEstadisticasSincronizacion obtiene estadísticas de sincronización
func (sc *SyncController) GetEstadisticasSincronizacion(c *gin.Context) {
	user, err := sc.GetUserFromContext(c)
	if err != nil {
		sc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	// Verificar permisos para ver estadísticas
	if !sc.CheckPermission(user, "ver_estadisticas_sync") {
		sc.ResponseError(c, http.StatusForbidden, "Sin permisos para ver estadísticas", nil)
		return
	}

	periodo := c.Query("periodo") // "dia", "semana", "mes", "año"
	if periodo == "" {
		periodo = "dia"
	}

	sucursalID := c.Query("sucursal_id")
	var sucursalUUID *uuid.UUID
	if sucursalID != "" {
		if id, err := uuid.Parse(sucursalID); err == nil {
			sucursalUUID = &id
		}
	}

	estadisticas, err := sc.sincronizacionService.GetEstadisticasSincronizacion(periodo, sucursalUUID)
	if err != nil {
		sc.ResponseError(c, http.StatusInternalServerError, "Error al obtener estadísticas", err)
		return
	}

	sc.SetCacheHeaders(c, 300) // Cache por 5 minutos
	sc.ResponseSuccess(c, estadisticas)
}

// GetMetricasRendimiento obtiene métricas de rendimiento
func (sc *SyncController) GetMetricasRendimiento(c *gin.Context) {
	user, err := sc.GetUserFromContext(c)
	if err != nil {
		sc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	// Verificar permisos para ver métricas
	if !sc.CheckPermission(user, "ver_metricas_sync") {
		sc.ResponseError(c, http.StatusForbidden, "Sin permisos para ver métricas", nil)
		return
	}

	periodo := c.Query("periodo")
	if periodo == "" {
		periodo = "hora"
	}

	limite := 100
	if limiteStr := c.Query("limite"); limiteStr != "" {
		if l, err := strconv.Atoi(limiteStr); err == nil && l > 0 && l <= 1000 {
			limite = l
		}
	}

	metricas, err := sc.sincronizacionService.GetMetricasRendimiento(periodo, limite)
	if err != nil {
		sc.ResponseError(c, http.StatusInternalServerError, "Error al obtener métricas", err)
		return
	}

	sc.SetCacheHeaders(c, 60) // Cache por 1 minuto
	sc.ResponseSuccess(c, metricas)
}

// ===== UTILIDADES =====

// ValidarConectividad valida la conectividad con una sucursal
func (sc *SyncController) ValidarConectividad(c *gin.Context) {
	sucursalID, err := sc.ParseUUID(c, "sucursal_id")
	if err != nil {
		sc.ResponseError(c, http.StatusBadRequest, "ID de sucursal inválido", err)
		return
	}

	user, err := sc.GetUserFromContext(c)
	if err != nil {
		sc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	// Verificar permisos para validar conectividad
	if !sc.CheckPermission(user, "validar_conectividad") {
		sc.ResponseError(c, http.StatusForbidden, "Sin permisos para validar conectividad", nil)
		return
	}

	resultado, err := sc.sincronizacionService.ValidarConectividad(sucursalID)
	if err != nil {
		sc.ResponseError(c, http.StatusBadRequest, "Error al validar conectividad", err)
		return
	}

	sc.ResponseSuccess(c, resultado)
}

// LimpiarLogsAntiguos limpia logs antiguos de sincronización
func (sc *SyncController) LimpiarLogsAntiguos(c *gin.Context) {
	var limpiarDTO models.LimpiarLogsDTO
	if err := sc.ValidateJSON(c, &limpiarDTO); err != nil {
		sc.ResponseError(c, http.StatusBadRequest, "Datos de limpieza inválidos", err)
		return
	}

	user, err := sc.GetUserFromContext(c)
	if err != nil {
		sc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	// Verificar permisos para limpiar logs
	if !sc.CheckPermission(user, "limpiar_logs") {
		sc.ResponseError(c, http.StatusForbidden, "Sin permisos para limpiar logs", nil)
		return
	}

	resultado, err := sc.sincronizacionService.LimpiarLogsAntiguos(limpiarDTO, user.ID)
	if err != nil {
		sc.ResponseError(c, http.StatusBadRequest, "Error al limpiar logs", err)
		return
	}

	sc.LogActivity(c, "limpiar_logs", resultado)
	sc.ResponseSuccess(c, resultado)
}

// ExportarLogs exporta logs de sincronización
func (sc *SyncController) ExportarLogs(c *gin.Context) {
	var exportarDTO models.ExportarLogsDTO
	if err := sc.ValidateJSON(c, &exportarDTO); err != nil {
		sc.ResponseError(c, http.StatusBadRequest, "Datos de exportación inválidos", err)
		return
	}

	user, err := sc.GetUserFromContext(c)
	if err != nil {
		sc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	// Verificar permisos para exportar logs
	if !sc.CheckPermission(user, "exportar_logs") {
		sc.ResponseError(c, http.StatusForbidden, "Sin permisos para exportar logs", nil)
		return
	}

	archivo, contentType, filename, err := sc.sincronizacionService.ExportarLogs(exportarDTO, user.ID)
	if err != nil {
		sc.ResponseError(c, http.StatusBadRequest, "Error al exportar logs", err)
		return
	}

	sc.LogActivity(c, "exportar_logs", gin.H{
		"formato": exportarDTO.Formato,
		"periodo": exportarDTO.Periodo,
	})

	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, contentType, archivo)
}

// GetResumenSincronizacion obtiene un resumen del estado de sincronización
func (sc *SyncController) GetResumenSincronizacion(c *gin.Context) {
	user, err := sc.GetUserFromContext(c)
	if err != nil {
		sc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	// Verificar permisos para ver resumen
	if !sc.CheckPermission(user, "ver_resumen_sync") {
		sc.ResponseError(c, http.StatusForbidden, "Sin permisos para ver resumen", nil)
		return
	}

	resumen, err := sc.sincronizacionService.GetResumenSincronizacion()
	if err != nil {
		sc.ResponseError(c, http.StatusInternalServerError, "Error al obtener resumen", err)
		return
	}

	sc.SetCacheHeaders(c, 120) // Cache por 2 minutos
	sc.ResponseSuccess(c, resumen)
}
