package services

import (
	"time"
	
	"github.com/google/uuid"
	"ferre-pos-servidor-central/internal/models"
)

// AuthService define la interfaz para servicios de autenticación
type AuthService interface {
	Login(email, password, terminalID, sucursalID string) (*models.AuthResponse, error)
	Logout(token string) error
	RefreshToken(refreshToken string) (*models.AuthResponse, error)
	ValidateToken(token string) (*models.Usuario, error)
	GenerateToken(user *models.Usuario) (*models.AuthResponse, error)
	ChangePassword(userID uuid.UUID, oldPassword, newPassword string) error
	ResetPassword(email string) error
	ConfirmPasswordReset(token, newPassword string) error
}

// ProductoService define la interfaz para servicios de productos
type ProductoService interface {
	GetProductosPOS(filter models.ProductoFilter) ([]models.ProductoPOSResponseDTO, models.PaginationResponse, error)
	GetProductoPOSByID(id uuid.UUID, sucursalID *uuid.UUID) (*models.ProductoPOSResponseDTO, error)
	GetProductoPOSByCodigoBarra(codigoBarra string, sucursalID *uuid.UUID) (*models.ProductoPOSResponseDTO, error)
	BuscarProductosPOS(query string, sucursalID *uuid.UUID, limit int) ([]models.ProductoPOSResponseDTO, error)
	GetProductosParaEtiquetas(filter models.ProductoFilter) ([]models.ProductoEtiquetaDTO, models.PaginationResponse, error)
	BuscarProductosParaEtiquetas(query string, sucursalID *uuid.UUID, limit int) ([]models.ProductoEtiquetaDTO, error)
	CrearProducto(dto models.ProductoCreateDTO, userID uuid.UUID) (*models.ProductoResponseDTO, error)
	ActualizarProducto(id uuid.UUID, dto models.ProductoUpdateDTO, userID uuid.UUID) (*models.ProductoResponseDTO, error)
	EliminarProducto(id uuid.UUID, userID uuid.UUID) error
}

// StockService define la interfaz para servicios de stock
type StockService interface {
	GetStockProducto(productoID uuid.UUID, sucursalID *uuid.UUID) (*models.StockResponseDTO, error)
	ReservarStock(dto models.ReservaStockDTO, userID uuid.UUID, sucursalID *uuid.UUID) (*models.ReservaStockResponseDTO, error)
	LiberarReserva(reservaID uuid.UUID, userID uuid.UUID) error
	ActualizarStock(dto models.ActualizarStockDTO, userID uuid.UUID) error
	GetMovimientosStock(filter models.MovimientoStockFilter) ([]models.MovimientoStockResponseDTO, models.PaginationResponse, error)
	CrearMovimientoStock(dto models.MovimientoStockCreateDTO, userID uuid.UUID) (*models.MovimientoStockResponseDTO, error)
}

// VentaService define la interfaz para servicios de ventas
type VentaService interface {
	CrearVenta(dto models.VentaCreateDTO, userID uuid.UUID, sucursalID *uuid.UUID, terminalID string) (*models.VentaResponseDTO, error)
	GetVenta(id uuid.UUID, userID uuid.UUID, sucursalID *uuid.UUID) (*models.VentaResponseDTO, error)
	GetVentas(filter models.VentaFilter) ([]models.VentaListDTO, models.PaginationResponse, error)
	AnularVenta(id uuid.UUID, dto models.AnulacionVentaDTO, userID uuid.UUID) error
	GetEstadisticasDia(fecha string, userID uuid.UUID, sucursalID *uuid.UUID) (*models.EstadisticasVentaDTO, error)
	SincronizarDatos(dto models.SincronizacionDTO, userID uuid.UUID, sucursalID *uuid.UUID) (*models.SincronizacionResultDTO, error)
}

// FidelizacionService define la interfaz para servicios de fidelización
type FidelizacionService interface {
	GetCliente(id uuid.UUID, sucursalID *uuid.UUID) (*models.ClienteResponseDTO, error)
	BuscarClientes(query string, sucursalID *uuid.UUID, limit int) ([]models.ClienteListDTO, error)
	CrearCliente(dto models.ClienteCreateDTO, userID uuid.UUID, sucursalID *uuid.UUID) (*models.ClienteResponseDTO, error)
	ActualizarCliente(id uuid.UUID, dto models.ClienteUpdateDTO, userID uuid.UUID) (*models.ClienteResponseDTO, error)
	GetPuntosCliente(clienteID uuid.UUID, sucursalID *uuid.UUID) (*models.PuntosClienteResponseDTO, error)
	CanjearPuntos(dto models.CanjePuntosDTO, userID uuid.UUID, sucursalID *uuid.UUID) (*models.CanjeResponseDTO, error)
	AcumularPuntos(dto models.AcumularPuntosDTO, userID uuid.UUID) (*models.AcumulacionResponseDTO, error)
}

// SucursalService define la interfaz para servicios de sucursal
type SucursalService interface {
	GetConfiguracionPOS(sucursalID uuid.UUID) (*models.ConfiguracionPOSDTO, error)
	ValidarTerminal(terminalID, sucursalID string) error
	GetSucursal(id uuid.UUID) (*models.SucursalResponseDTO, error)
	GetSucursales(filter models.SucursalFilter) ([]models.SucursalListDTO, models.PaginationResponse, error)
	CrearSucursal(dto models.SucursalCreateDTO, userID uuid.UUID) (*models.SucursalResponseDTO, error)
	ActualizarSucursal(id uuid.UUID, dto models.SucursalUpdateDTO, userID uuid.UUID) (*models.SucursalResponseDTO, error)
}

// EtiquetaService define la interfaz para servicios de etiquetas
type EtiquetaService interface {
	GetPlantillas(filter models.PlantillaEtiquetaFilter) ([]models.PlantillaEtiquetaListDTO, models.PaginationResponse, error)
	GetPlantilla(id uuid.UUID, userID uuid.UUID, sucursalID *uuid.UUID) (*models.PlantillaEtiquetaResponseDTO, error)
	CrearPlantilla(dto models.PlantillaEtiquetaCreateDTO, userID uuid.UUID, sucursalID *uuid.UUID) (*models.PlantillaEtiquetaResponseDTO, error)
	ActualizarPlantilla(id uuid.UUID, dto models.PlantillaEtiquetaUpdateDTO, userID uuid.UUID) (*models.PlantillaEtiquetaResponseDTO, error)
	EliminarPlantilla(id uuid.UUID, userID uuid.UUID) error
	DuplicarPlantilla(id uuid.UUID, dto models.DuplicarPlantillaDTO, userID uuid.UUID) (*models.PlantillaEtiquetaResponseDTO, error)
	ValidarPlantilla(dto models.PlantillaEtiquetaCreateDTO) ([]models.ValidationError, error)
	GenerarPreviewPlantilla(id uuid.UUID, dto models.PreviewPlantillaDTO, userID uuid.UUID) ([]byte, string, error)
	
	GenerarEtiqueta(dto models.GenerarEtiquetaDTO, userID uuid.UUID, sucursalID *uuid.UUID) (*models.EtiquetaGeneradaResponseDTO, error)
	GenerarLoteEtiquetas(dto models.GenerarLoteEtiquetasDTO, userID uuid.UUID, sucursalID *uuid.UUID) (*models.LoteEtiquetasResponseDTO, error)
	GetEtiquetasGeneradas(filter models.EtiquetaGeneradaFilter) ([]models.EtiquetaGeneradaListDTO, models.PaginationResponse, error)
	GetEtiqueta(id uuid.UUID, userID uuid.UUID) (*models.EtiquetaGeneradaResponseDTO, error)
	DescargarEtiqueta(id uuid.UUID, userID uuid.UUID) ([]byte, string, error)
	
	GetLotesEtiquetas(filter models.LoteEtiquetasFilter) ([]models.LoteEtiquetasListDTO, models.PaginationResponse, error)
	GetLoteEtiquetas(id uuid.UUID, userID uuid.UUID) (*models.LoteEtiquetasResponseDTO, error)
	DescargarLoteEtiquetas(id uuid.UUID, formato string, userID uuid.UUID) ([]byte, string, string, error)
	CancelarLote(id uuid.UUID, userID uuid.UUID) error
	
	GetEstadisticasEtiquetas(userID uuid.UUID, sucursalID *uuid.UUID, periodo string) (*models.EstadisticasEtiquetasDTO, error)
}

// SincronizacionService define la interfaz para servicios de sincronización
type SincronizacionService interface {
	GetEstadoSincronizacion(sucursalID uuid.UUID) (*models.SincronizacionSucursalResponseDTO, error)
	GetEstadosSincronizacion(filter models.SincronizacionSucursalFilter) ([]models.SincronizacionSucursalResponseDTO, models.PaginationResponse, error)
	ActualizarConfiguracion(sucursalID uuid.UUID, dto models.SincronizacionSucursalUpdateDTO, userID uuid.UUID) (*models.SincronizacionSucursalResponseDTO, error)
	
	IniciarSincronizacion(sucursalID uuid.UUID, dto models.IniciarSincronizacionDTO, userID uuid.UUID) (*models.ResultadoSincronizacionDTO, error)
	DetenerSincronizacion(sucursalID uuid.UUID, userID uuid.UUID) error
	ReiniciarSincronizacion(sucursalID uuid.UUID, userID uuid.UUID) (*models.ResultadoSincronizacionDTO, error)
	
	GetLogsSincronizacion(filter models.LogSincronizacionFilter) ([]models.LogSincronizacionListDTO, models.PaginationResponse, error)
	GetLogSincronizacion(id uuid.UUID) (*models.LogSincronizacionResponseDTO, error)
	ReintentarOperacion(logID uuid.UUID, userID uuid.UUID) (*models.ResultadoOperacionDTO, error)
	
	GetConflictos(filter models.ConflictoSincronizacionFilter) ([]models.ConflictoSincronizacionResponseDTO, models.PaginationResponse, error)
	GetConflicto(id uuid.UUID) (*models.ConflictoSincronizacionResponseDTO, error)
	ResolverConflicto(id uuid.UUID, dto models.ConflictoSincronizacionResolverDTO, userID uuid.UUID) (*models.ConflictoSincronizacionResponseDTO, error)
	IgnorarConflicto(id uuid.UUID, dto models.IgnorarConflictoDTO, userID uuid.UUID) error
	
	GetConfiguracionGlobal() (*models.ConfiguracionSincronizacionGlobalResponseDTO, error)
	ActualizarConfiguracionGlobal(dto models.ConfiguracionSincronizacionGlobalUpdateDTO, userID uuid.UUID) (*models.ConfiguracionSincronizacionGlobalResponseDTO, error)
	
	GetEstadisticasSincronizacion(periodo string, sucursalID *uuid.UUID) (*models.EstadisticasSincronizacionDTO, error)
	GetMetricasRendimiento(periodo string, limite int) (*models.MetricasRendimientoDTO, error)
	ValidarConectividad(sucursalID uuid.UUID) (*models.ConectividadResultDTO, error)
	LimpiarLogsAntiguos(dto models.LimpiarLogsDTO, userID uuid.UUID) (*models.LimpiezaResultDTO, error)
	ExportarLogs(dto models.ExportarLogsDTO, userID uuid.UUID) ([]byte, string, string, error)
	GetResumenSincronizacion() (*models.ResumenSincronizacionDTO, error)
}

// ReporteService define la interfaz para servicios de reportes
type ReporteService interface {
	GetPlantillasReportes(filter models.PlantillaReporteFilter) ([]models.PlantillaReporteListDTO, models.PaginationResponse, error)
	GetPlantillaReporte(id uuid.UUID, rolUsuario models.RolUsuario, sucursalID *uuid.UUID) (*models.PlantillaReporteResponseDTO, error)
	CrearPlantillaReporte(dto models.PlantillaReporteCreateDTO, userID uuid.UUID) (*models.PlantillaReporteResponseDTO, error)
	ActualizarPlantillaReporte(id uuid.UUID, dto models.PlantillaReporteUpdateDTO, userID uuid.UUID) (*models.PlantillaReporteResponseDTO, error)
	EliminarPlantillaReporte(id uuid.UUID, userID uuid.UUID) error
	ValidarPlantillaReporte(dto models.PlantillaReporteCreateDTO) ([]models.ValidationError, error)
	
	GenerarReporte(dto models.ReporteGeneradoCreateDTO, userID uuid.UUID, sucursalID *uuid.UUID) (*models.ReporteGeneradoResponseDTO, error)
	GetReportesGenerados(filter models.ReporteGeneradoFilter) ([]models.ReporteGeneradoListDTO, models.PaginationResponse, error)
	GetReporteGenerado(id uuid.UUID, userID uuid.UUID) (*models.ReporteGeneradoResponseDTO, error)
	DescargarReporte(id uuid.UUID, userID uuid.UUID) ([]byte, string, string, error)
	CancelarReporte(id uuid.UUID, userID uuid.UUID) error
	CompartirReporte(id uuid.UUID, dto models.CompartirReporteDTO, userID uuid.UUID) error
	
	GetReportesProgramados(filter models.ReporteProgramadoFilter) ([]models.ReporteProgramadoListDTO, models.PaginationResponse, error)
	CrearReporteProgramado(dto models.ReporteProgramadoCreateDTO, userID uuid.UUID, sucursalID *uuid.UUID) (*models.ReporteProgramadoResponseDTO, error)
	ActualizarReporteProgramado(id uuid.UUID, dto models.ReporteProgramadoUpdateDTO, userID uuid.UUID) (*models.ReporteProgramadoResponseDTO, error)
	EliminarReporteProgramado(id uuid.UUID, userID uuid.UUID) error
	
	GetDashboards(filter models.DashboardFilter) ([]models.DashboardListDTO, models.PaginationResponse, error)
	GetDashboard(id uuid.UUID, userID uuid.UUID) (*models.DashboardResponseDTO, error)
	CrearDashboard(dto models.DashboardCreateDTO, userID uuid.UUID, sucursalID *uuid.UUID) (*models.DashboardResponseDTO, error)
	ActualizarDashboard(id uuid.UUID, dto models.DashboardUpdateDTO, userID uuid.UUID) (*models.DashboardResponseDTO, error)
	EliminarDashboard(id uuid.UUID, userID uuid.UUID) error
	ActualizarVisualizacionDashboard(id uuid.UUID) error
	
	GetEstadisticasReportes(userID uuid.UUID, sucursalID *uuid.UUID, periodo string) (*models.EstadisticasReportesDTO, error)
	GenerarPreviewReporte(plantillaID uuid.UUID, dto models.PreviewReporteDTO, userID uuid.UUID, sucursalID *uuid.UUID) ([]byte, string, error)
	ValidarParametrosReporte(plantillaID uuid.UUID, parametros models.ParametrosReporte) ([]models.ValidationError, error)
	ExportarDashboard(id uuid.UUID, formato string, userID uuid.UUID) ([]byte, string, string, error)
	GetDatosWidget(dashboardID uuid.UUID, widgetID string, filtros map[string]interface{}, userID uuid.UUID) (interface{}, error)
}

// UsuarioService define la interfaz para servicios de usuarios
type UsuarioService interface {
	GetUsuario(id uuid.UUID) (*models.UsuarioResponseDTO, error)
	GetUsuarios(filter models.UsuarioFilter) ([]models.UsuarioListDTO, models.PaginationResponse, error)
	CrearUsuario(dto models.UsuarioCreateDTO, userID uuid.UUID) (*models.UsuarioResponseDTO, error)
	ActualizarUsuario(id uuid.UUID, dto models.UsuarioUpdateDTO, userID uuid.UUID) (*models.UsuarioResponseDTO, error)
	EliminarUsuario(id uuid.UUID, userID uuid.UUID) error
	CambiarEstadoUsuario(id uuid.UUID, activo bool, userID uuid.UUID) error
	ActualizarPerfil(id uuid.UUID, dto models.ActualizarPerfilDTO) (*models.UsuarioResponseDTO, error)
	CambiarPassword(id uuid.UUID, dto models.CambiarPasswordDTO) error
	GetPermisos(id uuid.UUID) ([]string, error)
	ActualizarPermisos(id uuid.UUID, permisos []string, userID uuid.UUID) error
}

// NotificacionService define la interfaz para servicios de notificaciones
type NotificacionService interface {
	EnviarNotificacion(dto models.NotificacionDTO) error
	EnviarEmail(to []string, subject, body string, attachments []string) error
	EnviarSMS(to string, message string) error
	EnviarWebhook(url string, payload interface{}) error
	GetPlantillasNotificacion() ([]models.PlantillaNotificacionDTO, error)
	CrearPlantillaNotificacion(dto models.PlantillaNotificacionCreateDTO) (*models.PlantillaNotificacionDTO, error)
}

// CacheService define la interfaz para servicios de cache
type CacheService interface {
	Get(key string) (interface{}, error)
	Set(key string, value interface{}, ttl int) error
	Delete(key string) error
	Clear() error
	Exists(key string) (bool, error)
	GetKeys(pattern string) ([]string, error)
	Increment(key string) (int64, error)
	Decrement(key string) (int64, error)
	SetWithExpiration(key string, value interface{}, expiration time.Time) error
}

// LogService define la interfaz para servicios de logging
type LogService interface {
	LogActivity(userID uuid.UUID, action string, details interface{}) error
	LogError(userID *uuid.UUID, error string, context interface{}) error
	LogAccess(userID uuid.UUID, resource string, action string) error
	GetLogs(filter models.LogFilter) ([]models.LogDTO, models.PaginationResponse, error)
	GetLogsByUser(userID uuid.UUID, filter models.LogFilter) ([]models.LogDTO, models.PaginationResponse, error)
	ExportLogs(filter models.LogFilter, formato string) ([]byte, string, error)
	CleanupOldLogs(daysOld int) error
}

// MetricsService define la interfaz para servicios de métricas
type MetricsService interface {
	RecordMetric(name string, value float64, tags map[string]string) error
	IncrementCounter(name string, tags map[string]string) error
	RecordDuration(name string, duration time.Duration, tags map[string]string) error
	GetMetrics(filter models.MetricsFilter) (*models.MetricsResponseDTO, error)
	GetSystemMetrics() (*models.SystemMetricsDTO, error)
	GetDatabaseMetrics() (*models.DatabaseMetricsDTO, error)
	GetAPIMetrics() (*models.APIMetricsDTO, error)
}

// ConfigService define la interfaz para servicios de configuración
type ConfigService interface {
	GetConfig(key string) (interface{}, error)
	SetConfig(key string, value interface{}) error
	GetAllConfigs() (map[string]interface{}, error)
	ReloadConfig() error
	ValidateConfig() error
	GetConfigHistory(key string) ([]models.ConfigHistoryDTO, error)
	BackupConfig() (string, error)
	RestoreConfig(backupPath string) error
}

