package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// SincronizacionSucursal representa el estado de sincronización de una sucursal
type SincronizacionSucursal struct {
	BaseModel
	SucursalID                  uuid.UUID                   `json:"sucursal_id" db:"sucursal_id" binding:"required"`
	UltimaSync                  time.Time                   `json:"ultima_sync" db:"ultima_sync" default:"NOW()"`
	EstadoGeneral               EstadoSincronizacionDetalle `json:"estado_general" db:"estado_general" default:"pendiente"`
	TotalRegistrosPendientes    int                         `json:"total_registros_pendientes" db:"total_registros_pendientes" default:"0"`
	TotalRegistrosSincronizados int                         `json:"total_registros_sincronizados" db:"total_registros_sincronizados" default:"0"`
	TotalErrores                int                         `json:"total_errores" db:"total_errores" default:"0"`
	ProximaSync                 *time.Time                  `json:"proxima_sync,omitempty" db:"proxima_sync"`
	ConfiguracionSync           ConfiguracionSincronizacion `json:"configuracion_sync" db:"configuracion_sync"`
	EstadisticasSync            *EstadisticasSincronizacion `json:"estadisticas_sync,omitempty" db:"estadisticas_sync"`
	UltimoError                 *string                     `json:"ultimo_error,omitempty" db:"ultimo_error"`
	FechaUltimoError            *time.Time                  `json:"fecha_ultimo_error,omitempty" db:"fecha_ultimo_error"`
	VersionDatos                int                         `json:"version_datos" db:"version_datos" default:"1"`
	HashIntegridad              *string                     `json:"hash_integridad,omitempty" db:"hash_integridad"`
	MetadatosSync               *MetadatosSincronizacion    `json:"metadatos_sync,omitempty" db:"metadatos_sync"`

	// Relaciones
	Sucursal *Sucursal `json:"sucursal,omitempty" gorm:"foreignKey:SucursalID"`
}

// LogSincronizacion representa un log de sincronización
type LogSincronizacion struct {
	BaseModel
	SucursalID         uuid.UUID            `json:"sucursal_id" db:"sucursal_id" binding:"required"`
	TipoOperacion      TipoOperacionSync    `json:"tipo_operacion" db:"tipo_operacion" binding:"required"`
	EntidadAfectada    string               `json:"entidad_afectada" db:"entidad_afectada" binding:"required"`
	RegistroID         *uuid.UUID           `json:"registro_id,omitempty" db:"registro_id"`
	Accion             AccionSincronizacion `json:"accion" db:"accion" binding:"required"`
	Estado             EstadoOperacionSync  `json:"estado" db:"estado" default:"pendiente"`
	FechaOperacion     time.Time            `json:"fecha_operacion" db:"fecha_operacion" default:"NOW()"`
	FechaProcesamiento *time.Time           `json:"fecha_procesamiento,omitempty" db:"fecha_procesamiento"`
	DatosAntes         *json.RawMessage     `json:"datos_antes,omitempty" db:"datos_antes"`
	DatosDespues       *json.RawMessage     `json:"datos_despues,omitempty" db:"datos_despues"`
	ErrorDetalle       *string              `json:"error_detalle,omitempty" db:"error_detalle"`
	CodigoError        *string              `json:"codigo_error,omitempty" db:"codigo_error"`
	Reintentos         int                  `json:"reintentos" db:"reintentos" default:"0"`
	MaxReintentos      int                  `json:"max_reintentos" db:"max_reintentos" default:"3"`
	ProximoReintento   *time.Time           `json:"proximo_reintento,omitempty" db:"proximo_reintento"`
	Prioridad          PrioridadProceso     `json:"prioridad" db:"prioridad" default:"normal"`
	DatosAdicionales   *DatosAdicionalesLog `json:"datos_adicionales,omitempty" db:"datos_adicionales"`
	HashOperacion      *string              `json:"hash_operacion,omitempty" db:"hash_operacion"`
	TiempoEjecucion    *int                 `json:"tiempo_ejecucion_ms,omitempty" db:"tiempo_ejecucion_ms"`

	// Relaciones
	Sucursal *Sucursal `json:"sucursal,omitempty" gorm:"foreignKey:SucursalID"`
}

// ConflictoSincronizacion representa un conflicto durante la sincronización
type ConflictoSincronizacion struct {
	BaseModel
	SucursalID         uuid.UUID            `json:"sucursal_id" db:"sucursal_id" binding:"required"`
	EntidadAfectada    string               `json:"entidad_afectada" db:"entidad_afectada" binding:"required"`
	RegistroID         uuid.UUID            `json:"registro_id" db:"registro_id" binding:"required"`
	TipoConflicto      TipoConflicto        `json:"tipo_conflicto" db:"tipo_conflicto" binding:"required"`
	EstadoConflicto    EstadoConflicto      `json:"estado_conflicto" db:"estado_conflicto" default:"pendiente"`
	FechaDeteccion     time.Time            `json:"fecha_deteccion" db:"fecha_deteccion" default:"NOW()"`
	FechaResolucion    *time.Time           `json:"fecha_resolucion,omitempty" db:"fecha_resolucion"`
	DatosCentral       json.RawMessage      `json:"datos_central" db:"datos_central" binding:"required"`
	DatosSucursal      json.RawMessage      `json:"datos_sucursal" db:"datos_sucursal" binding:"required"`
	ResolucionAplicada *ResolucionConflicto `json:"resolucion_aplicada,omitempty" db:"resolucion_aplicada"`
	UsuarioResolucion  *uuid.UUID           `json:"usuario_resolucion,omitempty" db:"usuario_resolucion"`
	Observaciones      *string              `json:"observaciones,omitempty" db:"observaciones"`
	MetadatosConflicto *MetadatosConflicto  `json:"metadatos_conflicto,omitempty" db:"metadatos_conflicto"`
	Severidad          SeveridadConflicto   `json:"severidad" db:"severidad" default:"media"`
	ImpactoEstimado    *ImpactoConflicto    `json:"impacto_estimado,omitempty" db:"impacto_estimado"`

	// Relaciones
	Sucursal       *Sucursal `json:"sucursal,omitempty" gorm:"foreignKey:SucursalID"`
	UsuarioResolve *Usuario  `json:"usuario_resolve,omitempty" gorm:"foreignKey:UsuarioResolucion"`
}

// ConfiguracionSincronizacionGlobal representa la configuración global de sincronización
type ConfiguracionSincronizacionGlobal struct {
	BaseModel
	Nombre                      string                      `json:"nombre" db:"nombre" binding:"required"`
	Descripcion                 *string                     `json:"descripcion,omitempty" db:"descripcion"`
	Activa                      bool                        `json:"activa" db:"activa" default:"true"`
	IntervaloSyncMinutos        int                         `json:"intervalo_sync_minutos" db:"intervalo_sync_minutos" default:"15"`
	HorarioSyncInicio           string                      `json:"horario_sync_inicio" db:"horario_sync_inicio" default:"06:00"`
	HorarioSyncFin              string                      `json:"horario_sync_fin" db:"horario_sync_fin" default:"23:00"`
	DiasHabilitados             []int                       `json:"dias_habilitados" db:"dias_habilitados"` // 0=Domingo, 1=Lunes, etc.
	ConfiguracionEntidades      ConfiguracionEntidadesSync  `json:"configuracion_entidades" db:"configuracion_entidades"`
	ConfiguracionReintentos     ConfiguracionReintentos     `json:"configuracion_reintentos" db:"configuracion_reintentos"`
	ConfiguracionConflictos     ConfiguracionConflictos     `json:"configuracion_conflictos" db:"configuracion_conflictos"`
	ConfiguracionNotificaciones ConfiguracionNotificaciones `json:"configuracion_notificaciones" db:"configuracion_notificaciones"`
	LimitesRendimiento          LimitesRendimiento          `json:"limites_rendimiento" db:"limites_rendimiento"`
	ConfiguracionSeguridad      ConfiguracionSeguridadSync  `json:"configuracion_seguridad" db:"configuracion_seguridad"`
	MetricasObjetivo            *MetricasObjetivo           `json:"metricas_objetivo,omitempty" db:"metricas_objetivo"`
	VersionConfiguracion        int                         `json:"version_configuracion" db:"version_configuracion" default:"1"`
}

// Enums para sincronización

type EstadoSincronizacionDetalle string

const (
	SyncPendiente     EstadoSincronizacionDetalle = "pendiente"
	SyncEnProceso     EstadoSincronizacionDetalle = "en_proceso"
	SyncCompletada    EstadoSincronizacionDetalle = "completada"
	SyncError         EstadoSincronizacionDetalle = "error"
	SyncPausada       EstadoSincronizacionDetalle = "pausada"
	SyncDeshabilitada EstadoSincronizacionDetalle = "deshabilitada"
)

type TipoOperacionSync string

const (
	OperacionCrear       TipoOperacionSync = "crear"
	OperacionActualizar  TipoOperacionSync = "actualizar"
	OperacionEliminar    TipoOperacionSync = "eliminar"
	OperacionSincronizar TipoOperacionSync = "sincronizar"
	OperacionValidar     TipoOperacionSync = "validar"
)

type AccionSincronizacion string

const (
	AccionEnviar   AccionSincronizacion = "enviar"
	AccionRecibir  AccionSincronizacion = "recibir"
	AccionValidar  AccionSincronizacion = "validar"
	AccionResolver AccionSincronizacion = "resolver"
)

type EstadoOperacionSync string

const (
	OperacionPendiente  EstadoOperacionSync = "pendiente"
	OperacionProcesando EstadoOperacionSync = "procesando"
	OperacionCompletada EstadoOperacionSync = "completada"
	OperacionError      EstadoOperacionSync = "error"
	OperacionCancelada  EstadoOperacionSync = "cancelada"
)

type TipoConflicto string

const (
	ConflictoModificacionConcurrente TipoConflicto = "modificacion_concurrente"
	ConflictoEliminacionConcurrente  TipoConflicto = "eliminacion_concurrente"
	ConflictoValidacionDatos         TipoConflicto = "validacion_datos"
	ConflictoIntegridadReferencial   TipoConflicto = "integridad_referencial"
	ConflictoReglaNegocios           TipoConflicto = "regla_negocios"
)

type EstadoConflicto string

const (
	ConflictoPendiente EstadoConflicto = "pendiente"
	ConflictoResuelto  EstadoConflicto = "resuelto"
	ConflictoIgnorado  EstadoConflicto = "ignorado"
	ConflictoEscalado  EstadoConflicto = "escalado"
)

type SeveridadConflicto string

const (
	SeveridadBaja    SeveridadConflicto = "baja"
	SeveridadMedia   SeveridadConflicto = "media"
	SeveridadAlta    SeveridadConflicto = "alta"
	SeveridadCritica SeveridadConflicto = "critica"
)

// Estructuras JSON

type ConfiguracionSincronizacion struct {
	HabilitarSync              bool                   `json:"habilitar_sync"`
	IntervaloMinutos           int                    `json:"intervalo_minutos"`
	SincronizarProductos       bool                   `json:"sincronizar_productos"`
	SincronizarStock           bool                   `json:"sincronizar_stock"`
	SincronizarVentas          bool                   `json:"sincronizar_ventas"`
	SincronizarClientes        bool                   `json:"sincronizar_clientes"`
	SincronizarUsuarios        bool                   `json:"sincronizar_usuarios"`
	MaxRegistrosPorLote        int                    `json:"max_registros_por_lote"`
	TimeoutSegundos            int                    `json:"timeout_segundos"`
	ReintentoAutomatico        bool                   `json:"reintento_automatico"`
	NotificarErrores           bool                   `json:"notificar_errores"`
	ConfiguracionPersonalizada map[string]interface{} `json:"configuracion_personalizada,omitempty"`
}

type EstadisticasSincronizacion struct {
	TotalSincronizaciones    int        `json:"total_sincronizaciones"`
	SincronizacionesExitosas int        `json:"sincronizaciones_exitosas"`
	SincronizacionesError    int        `json:"sincronizaciones_error"`
	TiempoPromedioSync       float64    `json:"tiempo_promedio_sync_ms"`
	UltimaSyncExitosa        *time.Time `json:"ultima_sync_exitosa,omitempty"`
	RegistrosPorMinuto       float64    `json:"registros_por_minuto"`
	TasaExito                float64    `json:"tasa_exito_porcentaje"`
	FechaCalculado           time.Time  `json:"fecha_calculado"`
}

type MetadatosSincronizacion struct {
	VersionCliente          string                 `json:"version_cliente"`
	VersionServidor         string                 `json:"version_servidor"`
	TipoConexion            string                 `json:"tipo_conexion"`
	CalidadConexion         string                 `json:"calidad_conexion"`
	LatenciaPromedio        float64                `json:"latencia_promedio_ms"`
	AnchoBandaDisponible    float64                `json:"ancho_banda_mbps"`
	CompressionUsada        bool                   `json:"compression_usada"`
	EncriptacionUsada       bool                   `json:"encriptacion_usada"`
	MetadatosPersonalizados map[string]interface{} `json:"metadatos_personalizados,omitempty"`
}

type DatosAdicionalesLog struct {
	UsuarioOperacion   *uuid.UUID             `json:"usuario_operacion,omitempty"`
	TerminalOperacion  *uuid.UUID             `json:"terminal_operacion,omitempty"`
	DireccionIP        *string                `json:"direccion_ip,omitempty"`
	UserAgent          *string                `json:"user_agent,omitempty"`
	TamañoDatos        int                    `json:"tamaño_datos_bytes"`
	ChecksumDatos      *string                `json:"checksum_datos,omitempty"`
	MetadatosOperacion map[string]interface{} `json:"metadatos_operacion,omitempty"`
}

type ResolucionConflicto struct {
	TipoResolucion          string                 `json:"tipo_resolucion"` // "usar_central", "usar_sucursal", "fusionar", "manual"
	DatosResolucion         json.RawMessage        `json:"datos_resolucion"`
	JustificacionResolucion string                 `json:"justificacion_resolucion"`
	FechaResolucion         time.Time              `json:"fecha_resolucion"`
	UsuarioResolucion       uuid.UUID              `json:"usuario_resolucion"`
	MetadatosResolucion     map[string]interface{} `json:"metadatos_resolucion,omitempty"`
}

type MetadatosConflicto struct {
	CamposConflictivos          []string               `json:"campos_conflictivos"`
	ValoresCentral              map[string]interface{} `json:"valores_central"`
	ValoresSucursal             map[string]interface{} `json:"valores_sucursal"`
	FechaModificacionCentral    *time.Time             `json:"fecha_modificacion_central,omitempty"`
	FechaModificacionSucursal   *time.Time             `json:"fecha_modificacion_sucursal,omitempty"`
	UsuarioModificacionCentral  *uuid.UUID             `json:"usuario_modificacion_central,omitempty"`
	UsuarioModificacionSucursal *uuid.UUID             `json:"usuario_modificacion_sucursal,omitempty"`
	ContextoAdicional           map[string]interface{} `json:"contexto_adicional,omitempty"`
}

type ImpactoConflicto struct {
	ImpactoOperacional       string                 `json:"impacto_operacional"` // "bajo", "medio", "alto", "critico"
	ImpactoFinanciero        string                 `json:"impacto_financiero"`
	UsuariosAfectados        int                    `json:"usuarios_afectados"`
	ProcesosAfectados        []string               `json:"procesos_afectados"`
	TiempoEstimadoResolucion int                    `json:"tiempo_estimado_resolucion_minutos"`
	CostoEstimado            *float64               `json:"costo_estimado,omitempty"`
	MetricasImpacto          map[string]interface{} `json:"metricas_impacto,omitempty"`
}

type ConfiguracionEntidadesSync struct {
	Productos               ConfiguracionEntidad            `json:"productos"`
	Stock                   ConfiguracionEntidad            `json:"stock"`
	Ventas                  ConfiguracionEntidad            `json:"ventas"`
	Clientes                ConfiguracionEntidad            `json:"clientes"`
	Usuarios                ConfiguracionEntidad            `json:"usuarios"`
	EntidadesPersonalizadas map[string]ConfiguracionEntidad `json:"entidades_personalizadas,omitempty"`
}

type ConfiguracionEntidad struct {
	Habilitada                 bool                   `json:"habilitada"`
	Prioridad                  int                    `json:"prioridad"`
	IntervaloMinutos           int                    `json:"intervalo_minutos"`
	MaxRegistrosPorLote        int                    `json:"max_registros_por_lote"`
	CamposExcluidos            []string               `json:"campos_excluidos,omitempty"`
	FiltrosSync                map[string]interface{} `json:"filtros_sync,omitempty"`
	ValidacionesPersonalizadas []string               `json:"validaciones_personalizadas,omitempty"`
}

type ConfiguracionReintentos struct {
	MaxReintentos         int      `json:"max_reintentos"`
	IntervaloBaseSegundos int      `json:"intervalo_base_segundos"`
	FactorBackoff         float64  `json:"factor_backoff"`
	MaxIntervaloSegundos  int      `json:"max_intervalo_segundos"`
	ReintentarErrores     []string `json:"reintentar_errores"`
	NoReintentarErrores   []string `json:"no_reintentar_errores"`
}

type ConfiguracionConflictos struct {
	ResolucionAutomatica  bool              `json:"resolucion_automatica"`
	PreferenciaPorDefecto string            `json:"preferencia_por_defecto"` // "central", "sucursal"
	EscalarConflictos     bool              `json:"escalar_conflictos"`
	TiempoEscaladoMinutos int               `json:"tiempo_escalado_minutos"`
	NotificarConflictos   bool              `json:"notificar_conflictos"`
	ReglasResolucion      map[string]string `json:"reglas_resolucion,omitempty"`
}

type ConfiguracionNotificaciones struct {
	HabilitarNotificaciones  bool              `json:"habilitar_notificaciones"`
	NotificarErrores         bool              `json:"notificar_errores"`
	NotificarConflictos      bool              `json:"notificar_conflictos"`
	NotificarCompletado      bool              `json:"notificar_completado"`
	EmailsNotificacion       []string          `json:"emails_notificacion"`
	WebhooksNotificacion     []string          `json:"webhooks_notificacion,omitempty"`
	PlantillasPersonalizadas map[string]string `json:"plantillas_personalizadas,omitempty"`
}

type LimitesRendimiento struct {
	MaxConcurrencia        int     `json:"max_concurrencia"`
	MaxMemoriaMB           int     `json:"max_memoria_mb"`
	MaxTiempoEjecucionMin  int     `json:"max_tiempo_ejecucion_min"`
	MaxRegistrosPorSegundo int     `json:"max_registros_por_segundo"`
	LimiteAnchoBandaMbps   float64 `json:"limite_ancho_banda_mbps"`
	MonitoreoRendimiento   bool    `json:"monitoreo_rendimiento"`
}

type ConfiguracionSeguridadSync struct {
	RequiereAutenticacion bool     `json:"requiere_autenticacion"`
	RequiereEncriptacion  bool     `json:"requiere_encriptacion"`
	ValidarIntegridad     bool     `json:"validar_integridad"`
	LogearOperaciones     bool     `json:"logear_operaciones"`
	RetenerLogsHoras      int      `json:"retener_logs_horas"`
	IpsPermitidas         []string `json:"ips_permitidas,omitempty"`
	TokensAPI             []string `json:"tokens_api,omitempty"`
}

type MetricasObjetivo struct {
	TasaExitoMinima        float64            `json:"tasa_exito_minima_porcentaje"`
	TiempoMaximoSyncMin    int                `json:"tiempo_maximo_sync_min"`
	LatenciaMaximaMs       float64            `json:"latencia_maxima_ms"`
	DisponibilidadMinima   float64            `json:"disponibilidad_minima_porcentaje"`
	MetricasPersonalizadas map[string]float64 `json:"metricas_personalizadas,omitempty"`
}

// Implementar driver.Valuer para tipos JSON personalizados
func (c ConfiguracionSincronizacion) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c *ConfiguracionSincronizacion) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, c)
}

func (e EstadisticasSincronizacion) Value() (driver.Value, error) {
	return json.Marshal(e)
}

func (e *EstadisticasSincronizacion) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, e)
}

func (m MetadatosSincronizacion) Value() (driver.Value, error) {
	return json.Marshal(m)
}

func (m *MetadatosSincronizacion) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, m)
}

func (d DatosAdicionalesLog) Value() (driver.Value, error) {
	return json.Marshal(d)
}

func (d *DatosAdicionalesLog) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, d)
}

func (r ResolucionConflicto) Value() (driver.Value, error) {
	return json.Marshal(r)
}

func (r *ResolucionConflicto) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, r)
}

func (m MetadatosConflicto) Value() (driver.Value, error) {
	return json.Marshal(m)
}

func (m *MetadatosConflicto) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, m)
}

func (i ImpactoConflicto) Value() (driver.Value, error) {
	return json.Marshal(i)
}

func (i *ImpactoConflicto) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, i)
}

func (c ConfiguracionEntidadesSync) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c *ConfiguracionEntidadesSync) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, c)
}

func (c ConfiguracionReintentos) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c *ConfiguracionReintentos) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, c)
}

func (c ConfiguracionConflictos) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c *ConfiguracionConflictos) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, c)
}

func (c ConfiguracionNotificaciones) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c *ConfiguracionNotificaciones) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, c)
}

func (l LimitesRendimiento) Value() (driver.Value, error) {
	return json.Marshal(l)
}

func (l *LimitesRendimiento) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, l)
}

func (c ConfiguracionSeguridadSync) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c *ConfiguracionSeguridadSync) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, c)
}

func (m MetricasObjetivo) Value() (driver.Value, error) {
	return json.Marshal(m)
}

func (m *MetricasObjetivo) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, m)
}

// DTOs para SincronizacionSucursal

type SincronizacionSucursalResponseDTO struct {
	ID                          uuid.UUID                   `json:"id"`
	SucursalID                  uuid.UUID                   `json:"sucursal_id"`
	UltimaSync                  time.Time                   `json:"ultima_sync"`
	EstadoGeneral               EstadoSincronizacionDetalle `json:"estado_general"`
	TotalRegistrosPendientes    int                         `json:"total_registros_pendientes"`
	TotalRegistrosSincronizados int                         `json:"total_registros_sincronizados"`
	TotalErrores                int                         `json:"total_errores"`
	ProximaSync                 *time.Time                  `json:"proxima_sync,omitempty"`
	ConfiguracionSync           ConfiguracionSincronizacion `json:"configuracion_sync"`
	EstadisticasSync            *EstadisticasSincronizacion `json:"estadisticas_sync,omitempty"`
	UltimoError                 *string                     `json:"ultimo_error,omitempty"`
	FechaUltimoError            *time.Time                  `json:"fecha_ultimo_error,omitempty"`
	VersionDatos                int                         `json:"version_datos"`
	FechaCreacion               time.Time                   `json:"fecha_creacion"`
	FechaModificacion           time.Time                   `json:"fecha_modificacion"`
	SucursalNombre              *string                     `json:"sucursal_nombre,omitempty"`
}

type SincronizacionSucursalUpdateDTO struct {
	ConfiguracionSync *ConfiguracionSincronizacion `json:"configuracion_sync,omitempty"`
	ProximaSync       *time.Time                   `json:"proxima_sync,omitempty"`
}

// DTOs para LogSincronizacion

type LogSincronizacionResponseDTO struct {
	ID                 uuid.UUID            `json:"id"`
	SucursalID         uuid.UUID            `json:"sucursal_id"`
	TipoOperacion      TipoOperacionSync    `json:"tipo_operacion"`
	EntidadAfectada    string               `json:"entidad_afectada"`
	RegistroID         *uuid.UUID           `json:"registro_id,omitempty"`
	Accion             AccionSincronizacion `json:"accion"`
	Estado             EstadoOperacionSync  `json:"estado"`
	FechaOperacion     time.Time            `json:"fecha_operacion"`
	FechaProcesamiento *time.Time           `json:"fecha_procesamiento,omitempty"`
	ErrorDetalle       *string              `json:"error_detalle,omitempty"`
	CodigoError        *string              `json:"codigo_error,omitempty"`
	Reintentos         int                  `json:"reintentos"`
	MaxReintentos      int                  `json:"max_reintentos"`
	ProximoReintento   *time.Time           `json:"proximo_reintento,omitempty"`
	Prioridad          PrioridadProceso     `json:"prioridad"`
	TiempoEjecucion    *int                 `json:"tiempo_ejecucion_ms,omitempty"`
	FechaCreacion      time.Time            `json:"fecha_creacion"`
	SucursalNombre     *string              `json:"sucursal_nombre,omitempty"`
}

type LogSincronizacionListDTO struct {
	ID                 uuid.UUID            `json:"id"`
	SucursalID         uuid.UUID            `json:"sucursal_id"`
	TipoOperacion      TipoOperacionSync    `json:"tipo_operacion"`
	EntidadAfectada    string               `json:"entidad_afectada"`
	Accion             AccionSincronizacion `json:"accion"`
	Estado             EstadoOperacionSync  `json:"estado"`
	FechaOperacion     time.Time            `json:"fecha_operacion"`
	FechaProcesamiento *time.Time           `json:"fecha_procesamiento,omitempty"`
	Reintentos         int                  `json:"reintentos"`
	TiempoEjecucion    *int                 `json:"tiempo_ejecucion_ms,omitempty"`
	SucursalNombre     *string              `json:"sucursal_nombre,omitempty"`
}

// DTOs para ConflictoSincronizacion

type ConflictoSincronizacionResponseDTO struct {
	ID                 uuid.UUID            `json:"id"`
	SucursalID         uuid.UUID            `json:"sucursal_id"`
	EntidadAfectada    string               `json:"entidad_afectada"`
	RegistroID         uuid.UUID            `json:"registro_id"`
	TipoConflicto      TipoConflicto        `json:"tipo_conflicto"`
	EstadoConflicto    EstadoConflicto      `json:"estado_conflicto"`
	FechaDeteccion     time.Time            `json:"fecha_deteccion"`
	FechaResolucion    *time.Time           `json:"fecha_resolucion,omitempty"`
	DatosCentral       json.RawMessage      `json:"datos_central"`
	DatosSucursal      json.RawMessage      `json:"datos_sucursal"`
	ResolucionAplicada *ResolucionConflicto `json:"resolucion_aplicada,omitempty"`
	UsuarioResolucion  *uuid.UUID           `json:"usuario_resolucion,omitempty"`
	Observaciones      *string              `json:"observaciones,omitempty"`
	MetadatosConflicto *MetadatosConflicto  `json:"metadatos_conflicto,omitempty"`
	Severidad          SeveridadConflicto   `json:"severidad"`
	ImpactoEstimado    *ImpactoConflicto    `json:"impacto_estimado,omitempty"`
	FechaCreacion      time.Time            `json:"fecha_creacion"`
	SucursalNombre     *string              `json:"sucursal_nombre,omitempty"`
	UsuarioNombre      *string              `json:"usuario_nombre,omitempty"`
}

type ConflictoSincronizacionResolverDTO struct {
	TipoResolucion          string          `json:"tipo_resolucion" binding:"required"` // "usar_central", "usar_sucursal", "fusionar", "manual"
	DatosResolucion         json.RawMessage `json:"datos_resolucion,omitempty"`
	JustificacionResolucion string          `json:"justificacion_resolucion" binding:"required"`
	Observaciones           *string         `json:"observaciones,omitempty"`
}

// Filtros para búsqueda

type SincronizacionSucursalFilter struct {
	PaginationFilter
	SortFilter
	SucursalID     *uuid.UUID                   `json:"sucursal_id,omitempty" form:"sucursal_id"`
	Estado         *EstadoSincronizacionDetalle `json:"estado,omitempty" form:"estado"`
	ConErrores     *bool                        `json:"con_errores,omitempty" form:"con_errores"`
	SinSincronizar *bool                        `json:"sin_sincronizar,omitempty" form:"sin_sincronizar"`
}

type LogSincronizacionFilter struct {
	PaginationFilter
	SortFilter
	DateRangeFilter
	SucursalID      *uuid.UUID            `json:"sucursal_id,omitempty" form:"sucursal_id"`
	TipoOperacion   *TipoOperacionSync    `json:"tipo_operacion,omitempty" form:"tipo_operacion"`
	EntidadAfectada *string               `json:"entidad_afectada,omitempty" form:"entidad_afectada"`
	Accion          *AccionSincronizacion `json:"accion,omitempty" form:"accion"`
	Estado          *EstadoOperacionSync  `json:"estado,omitempty" form:"estado"`
	RegistroID      *uuid.UUID            `json:"registro_id,omitempty" form:"registro_id"`
	ConErrores      *bool                 `json:"con_errores,omitempty" form:"con_errores"`
	Prioridad       *PrioridadProceso     `json:"prioridad,omitempty" form:"prioridad"`
}

type ConflictoSincronizacionFilter struct {
	PaginationFilter
	SortFilter
	DateRangeFilter
	SucursalID        *uuid.UUID          `json:"sucursal_id,omitempty" form:"sucursal_id"`
	EntidadAfectada   *string             `json:"entidad_afectada,omitempty" form:"entidad_afectada"`
	TipoConflicto     *TipoConflicto      `json:"tipo_conflicto,omitempty" form:"tipo_conflicto"`
	EstadoConflicto   *EstadoConflicto    `json:"estado_conflicto,omitempty" form:"estado_conflicto"`
	Severidad         *SeveridadConflicto `json:"severidad,omitempty" form:"severidad"`
	UsuarioResolucion *uuid.UUID          `json:"usuario_resolucion,omitempty" form:"usuario_resolucion"`
	SinResolver       *bool               `json:"sin_resolver,omitempty" form:"sin_resolver"`
}

// Métodos helper

func (s *SincronizacionSucursal) ActualizarEstadisticas() {
	if s.EstadisticasSync == nil {
		s.EstadisticasSync = &EstadisticasSincronizacion{}
	}

	s.EstadisticasSync.TotalSincronizaciones++

	if s.EstadoGeneral == SyncCompletada {
		s.EstadisticasSync.SincronizacionesExitosas++
		now := time.Now()
		s.EstadisticasSync.UltimaSyncExitosa = &now
	} else if s.EstadoGeneral == SyncError {
		s.EstadisticasSync.SincronizacionesError++
	}

	// Calcular tasa de éxito
	if s.EstadisticasSync.TotalSincronizaciones > 0 {
		s.EstadisticasSync.TasaExito = float64(s.EstadisticasSync.SincronizacionesExitosas) / float64(s.EstadisticasSync.TotalSincronizaciones) * 100
	}

	s.EstadisticasSync.FechaCalculado = time.Now()
}

func (s *SincronizacionSucursal) RequiereSincronizacion() bool {
	return s.TotalRegistrosPendientes > 0 || s.EstadoGeneral == SyncError
}

func (s *SincronizacionSucursal) PuedeEjecutarSync() bool {
	return s.EstadoGeneral != SyncEnProceso && s.EstadoGeneral != SyncDeshabilitada
}

func (l *LogSincronizacion) PuedeReintentar() bool {
	return l.Estado == OperacionError && l.Reintentos < l.MaxReintentos
}

func (l *LogSincronizacion) ProgramarReintento() {
	if l.PuedeReintentar() {
		l.Reintentos++
		// Backoff exponencial
		intervalo := time.Duration(l.Reintentos*l.Reintentos) * time.Minute
		proximoReintento := time.Now().Add(intervalo)
		l.ProximoReintento = &proximoReintento
	}
}

func (c *ConflictoSincronizacion) PuedeResolver() bool {
	return c.EstadoConflicto == ConflictoPendiente
}

func (c *ConflictoSincronizacion) Resolver(resolucion ResolucionConflicto, usuarioID uuid.UUID) error {
	if !c.PuedeResolver() {
		return fmt.Errorf("el conflicto no puede ser resuelto en su estado actual: %s", c.EstadoConflicto)
	}

	c.ResolucionAplicada = &resolucion
	c.UsuarioResolucion = &usuarioID
	c.EstadoConflicto = ConflictoResuelto
	now := time.Now()
	c.FechaResolucion = &now

	return nil
}

func (s *SincronizacionSucursal) ToResponseDTO() SincronizacionSucursalResponseDTO {
	dto := SincronizacionSucursalResponseDTO{
		ID:                          s.ID,
		SucursalID:                  s.SucursalID,
		UltimaSync:                  s.UltimaSync,
		EstadoGeneral:               s.EstadoGeneral,
		TotalRegistrosPendientes:    s.TotalRegistrosPendientes,
		TotalRegistrosSincronizados: s.TotalRegistrosSincronizados,
		TotalErrores:                s.TotalErrores,
		ProximaSync:                 s.ProximaSync,
		ConfiguracionSync:           s.ConfiguracionSync,
		EstadisticasSync:            s.EstadisticasSync,
		UltimoError:                 s.UltimoError,
		FechaUltimoError:            s.FechaUltimoError,
		VersionDatos:                s.VersionDatos,
		FechaCreacion:               s.FechaCreacion,
		FechaModificacion:           s.FechaModificacion,
	}

	if s.Sucursal != nil {
		dto.SucursalNombre = &s.Sucursal.Nombre
	}

	return dto
}

func (l *LogSincronizacion) ToResponseDTO() LogSincronizacionResponseDTO {
	dto := LogSincronizacionResponseDTO{
		ID:                 l.ID,
		SucursalID:         l.SucursalID,
		TipoOperacion:      l.TipoOperacion,
		EntidadAfectada:    l.EntidadAfectada,
		RegistroID:         l.RegistroID,
		Accion:             l.Accion,
		Estado:             l.Estado,
		FechaOperacion:     l.FechaOperacion,
		FechaProcesamiento: l.FechaProcesamiento,
		ErrorDetalle:       l.ErrorDetalle,
		CodigoError:        l.CodigoError,
		Reintentos:         l.Reintentos,
		MaxReintentos:      l.MaxReintentos,
		ProximoReintento:   l.ProximoReintento,
		Prioridad:          l.Prioridad,
		TiempoEjecucion:    l.TiempoEjecucion,
		FechaCreacion:      l.FechaCreacion,
	}

	if l.Sucursal != nil {
		dto.SucursalNombre = &l.Sucursal.Nombre
	}

	return dto
}

func (l *LogSincronizacion) ToListDTO() LogSincronizacionListDTO {
	dto := LogSincronizacionListDTO{
		ID:                 l.ID,
		SucursalID:         l.SucursalID,
		TipoOperacion:      l.TipoOperacion,
		EntidadAfectada:    l.EntidadAfectada,
		Accion:             l.Accion,
		Estado:             l.Estado,
		FechaOperacion:     l.FechaOperacion,
		FechaProcesamiento: l.FechaProcesamiento,
		Reintentos:         l.Reintentos,
		TiempoEjecucion:    l.TiempoEjecucion,
	}

	if l.Sucursal != nil {
		dto.SucursalNombre = &l.Sucursal.Nombre
	}

	return dto
}

func (c *ConflictoSincronizacion) ToResponseDTO() ConflictoSincronizacionResponseDTO {
	dto := ConflictoSincronizacionResponseDTO{
		ID:                 c.ID,
		SucursalID:         c.SucursalID,
		EntidadAfectada:    c.EntidadAfectada,
		RegistroID:         c.RegistroID,
		TipoConflicto:      c.TipoConflicto,
		EstadoConflicto:    c.EstadoConflicto,
		FechaDeteccion:     c.FechaDeteccion,
		FechaResolucion:    c.FechaResolucion,
		DatosCentral:       c.DatosCentral,
		DatosSucursal:      c.DatosSucursal,
		ResolucionAplicada: c.ResolucionAplicada,
		UsuarioResolucion:  c.UsuarioResolucion,
		Observaciones:      c.Observaciones,
		MetadatosConflicto: c.MetadatosConflicto,
		Severidad:          c.Severidad,
		ImpactoEstimado:    c.ImpactoEstimado,
		FechaCreacion:      c.FechaCreacion,
	}

	if c.Sucursal != nil {
		dto.SucursalNombre = &c.Sucursal.Nombre
	}

	if c.UsuarioResolve != nil {
		nombreUsuario := c.UsuarioResolve.Nombre
		if c.UsuarioResolve.Apellido != nil {
			nombreUsuario += " " + *c.UsuarioResolve.Apellido
		}
		dto.UsuarioNombre = &nombreUsuario
	}

	return dto
}

// Validaciones personalizadas

func (s *SincronizacionSucursal) Validate() error {
	if s.TotalRegistrosPendientes < 0 {
		return fmt.Errorf("total de registros pendientes no puede ser negativo")
	}

	if s.TotalRegistrosSincronizados < 0 {
		return fmt.Errorf("total de registros sincronizados no puede ser negativo")
	}

	if s.TotalErrores < 0 {
		return fmt.Errorf("total de errores no puede ser negativo")
	}

	if s.VersionDatos <= 0 {
		return fmt.Errorf("versión de datos debe ser positiva")
	}

	return nil
}

func (l *LogSincronizacion) Validate() error {
	if l.Reintentos < 0 {
		return fmt.Errorf("reintentos no puede ser negativo")
	}

	if l.MaxReintentos < 0 {
		return fmt.Errorf("máximo de reintentos no puede ser negativo")
	}

	if l.Reintentos > l.MaxReintentos {
		return fmt.Errorf("reintentos no puede ser mayor al máximo permitido")
	}

	return nil
}

func (c *ConflictoSincronizacion) Validate() error {
	if c.FechaResolucion != nil && c.FechaResolucion.Before(c.FechaDeteccion) {
		return fmt.Errorf("fecha de resolución no puede ser anterior a fecha de detección")
	}

	return nil
}

// Funciones de utilidad

func GenerarHashOperacion(entidad string, registroID uuid.UUID, accion AccionSincronizacion) string {
	// En implementación real, generar hash SHA256
	return fmt.Sprintf("hash_%s_%s_%s_%d", entidad, registroID.String(), accion, time.Now().Unix())
}

func CalcularProximaSync(intervaloMinutos int) time.Time {
	return time.Now().Add(time.Duration(intervaloMinutos) * time.Minute)
}
