package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Usuario representa un usuario del sistema
type Usuario struct {
	BaseModel
	Rut                   string                 `json:"rut" db:"rut" binding:"required" validate:"required,rut"`
	Nombre                string                 `json:"nombre" db:"nombre" binding:"required" validate:"required,max=255"`
	Apellido              *string                `json:"apellido,omitempty" db:"apellido"`
	Email                 *string                `json:"email,omitempty" db:"email" validate:"omitempty,email"`
	Telefono              *string                `json:"telefono,omitempty" db:"telefono"`
	Rol                   RolUsuario             `json:"rol" db:"rol" binding:"required"`
	SucursalID            *uuid.UUID             `json:"sucursal_id,omitempty" db:"sucursal_id"`
	PasswordHash          string                 `json:"-" db:"password_hash"`
	Salt                  string                 `json:"-" db:"salt"`
	Activo                bool                   `json:"activo" db:"activo" default:"true"`
	UltimoAcceso          *time.Time             `json:"ultimo_acceso,omitempty" db:"ultimo_acceso"`
	IntentosFallidos      int                    `json:"intentos_fallidos" db:"intentos_fallidos" default:"0"`
	BloqueadoHasta        *time.Time             `json:"bloqueado_hasta,omitempty" db:"bloqueado_hasta"`
	ConfiguracionPersonal *ConfiguracionPersonal `json:"configuracion_personal,omitempty" db:"configuracion_personal"`
	PermisosEspeciales    *PermisosEspeciales    `json:"permisos_especiales,omitempty" db:"permisos_especiales"`
	CachePermisos         *CachePermisos         `json:"cache_permisos,omitempty" db:"cache_permisos"`
	HashSesionActiva      *string                `json:"-" db:"hash_sesion_activa"`
	UltimoTerminalID      *uuid.UUID             `json:"ultimo_terminal_id,omitempty" db:"ultimo_terminal_id"`
	PreferenciasUI        *PreferenciasUI        `json:"preferencias_ui,omitempty" db:"preferencias_ui"`

	// Relaciones
	Sucursal *Sucursal `json:"sucursal,omitempty" gorm:"foreignKey:SucursalID"`
}

// ConfiguracionPersonal contiene configuraciones específicas del usuario
type ConfiguracionPersonal struct {
	IdiomaPreferido        string            `json:"idioma_preferido"`
	ZonaHoraria            string            `json:"zona_horaria"`
	FormatoFecha           string            `json:"formato_fecha"`
	FormatoHora            string            `json:"formato_hora"`
	NotificacionesEmail    bool              `json:"notificaciones_email"`
	NotificacionesPush     bool              `json:"notificaciones_push"`
	TemaInterfaz           string            `json:"tema_interfaz"` // "claro", "oscuro", "auto"
	ConfiguracionesPersonalizadas map[string]interface{} `json:"configuraciones_personalizadas,omitempty"`
}

// PermisosEspeciales contiene permisos adicionales específicos del usuario
type PermisosEspeciales struct {
	PuedeAnularVentas         bool     `json:"puede_anular_ventas"`
	PuedeAplicarDescuentos    bool     `json:"puede_aplicar_descuentos"`
	PuedeModificarPrecios     bool     `json:"puede_modificar_precios"`
	PuedeAccederReportes      bool     `json:"puede_acceder_reportes"`
	PuedeGestionarUsuarios    bool     `json:"puede_gestionar_usuarios"`
	PuedeConfigurarsistema    bool     `json:"puede_configurar_sistema"`
	SucursalesPermitidas      []string `json:"sucursales_permitidas,omitempty"`
	ModulosPermitidos         []string `json:"modulos_permitidos,omitempty"`
	LimiteDescuentoPorcentaje *float64 `json:"limite_descuento_porcentaje,omitempty"`
	LimiteMontoVenta          *float64 `json:"limite_monto_venta,omitempty"`
}

// CachePermisos contiene permisos en cache para acceso rápido
type CachePermisos struct {
	PermisosCalculados map[string]bool `json:"permisos_calculados"`
	FechaCalculado     time.Time       `json:"fecha_calculado"`
	ValidoHasta        time.Time       `json:"valido_hasta"`
	Version            int             `json:"version"`
}

// PreferenciasUI contiene preferencias de interfaz de usuario
type PreferenciasUI struct {
	TamañoFuente           string            `json:"tamaño_fuente"`
	ColorTema              string            `json:"color_tema"`
	MostrarAyudas          bool              `json:"mostrar_ayudas"`
	SonidosHabilitados     bool              `json:"sonidos_habilitados"`
	AnimacionesHabilitadas bool              `json:"animaciones_habilitadas"`
	LayoutPersonalizado    map[string]interface{} `json:"layout_personalizado,omitempty"`
	AtajosPersonalizados   map[string]string `json:"atajos_personalizados,omitempty"`
}

// Implementar driver.Valuer para tipos JSON personalizados
func (c ConfiguracionPersonal) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c *ConfiguracionPersonal) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, c)
}

func (p PermisosEspeciales) Value() (driver.Value, error) {
	return json.Marshal(p)
}

func (p *PermisosEspeciales) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, p)
}

func (c CachePermisos) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c *CachePermisos) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, c)
}

func (p PreferenciasUI) Value() (driver.Value, error) {
	return json.Marshal(p)
}

func (p *PreferenciasUI) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, p)
}

// DTOs para Usuario

type UsuarioCreateDTO struct {
	Rut                   string                 `json:"rut" binding:"required" validate:"required,rut"`
	Nombre                string                 `json:"nombre" binding:"required" validate:"required,max=255"`
	Apellido              *string                `json:"apellido,omitempty"`
	Email                 *string                `json:"email,omitempty" validate:"omitempty,email"`
	Telefono              *string                `json:"telefono,omitempty"`
	Rol                   RolUsuario             `json:"rol" binding:"required"`
	SucursalID            *uuid.UUID             `json:"sucursal_id,omitempty"`
	Password              string                 `json:"password" binding:"required,min=8"`
	ConfiguracionPersonal *ConfiguracionPersonal `json:"configuracion_personal,omitempty"`
	PermisosEspeciales    *PermisosEspeciales    `json:"permisos_especiales,omitempty"`
}

type UsuarioUpdateDTO struct {
	Nombre                *string                `json:"nombre,omitempty" validate:"omitempty,max=255"`
	Apellido              *string                `json:"apellido,omitempty"`
	Email                 *string                `json:"email,omitempty" validate:"omitempty,email"`
	Telefono              *string                `json:"telefono,omitempty"`
	Rol                   *RolUsuario            `json:"rol,omitempty"`
	SucursalID            *uuid.UUID             `json:"sucursal_id,omitempty"`
	Activo                *bool                  `json:"activo,omitempty"`
	ConfiguracionPersonal *ConfiguracionPersonal `json:"configuracion_personal,omitempty"`
	PermisosEspeciales    *PermisosEspeciales    `json:"permisos_especiales,omitempty"`
}

type UsuarioResponseDTO struct {
	ID                    uuid.UUID              `json:"id"`
	Rut                   string                 `json:"rut"`
	Nombre                string                 `json:"nombre"`
	Apellido              *string                `json:"apellido,omitempty"`
	Email                 *string                `json:"email,omitempty"`
	Telefono              *string                `json:"telefono,omitempty"`
	Rol                   RolUsuario             `json:"rol"`
	SucursalID            *uuid.UUID             `json:"sucursal_id,omitempty"`
	Activo                bool                   `json:"activo"`
	UltimoAcceso          *time.Time             `json:"ultimo_acceso,omitempty"`
	BloqueadoHasta        *time.Time             `json:"bloqueado_hasta,omitempty"`
	ConfiguracionPersonal *ConfiguracionPersonal `json:"configuracion_personal,omitempty"`
	PermisosEspeciales    *PermisosEspeciales    `json:"permisos_especiales,omitempty"`
	FechaCreacion         time.Time              `json:"fecha_creacion"`
	FechaModificacion     time.Time              `json:"fecha_modificacion"`
	Sucursal              *SucursalListDTO       `json:"sucursal,omitempty"`
}

type UsuarioListDTO struct {
	ID                uuid.UUID    `json:"id"`
	Rut               string       `json:"rut"`
	Nombre            string       `json:"nombre"`
	Apellido          *string      `json:"apellido,omitempty"`
	Email             *string      `json:"email,omitempty"`
	Rol               RolUsuario   `json:"rol"`
	SucursalID        *uuid.UUID   `json:"sucursal_id,omitempty"`
	Activo            bool         `json:"activo"`
	UltimoAcceso      *time.Time   `json:"ultimo_acceso,omitempty"`
	FechaCreacion     time.Time    `json:"fecha_creacion"`
	FechaModificacion time.Time    `json:"fecha_modificacion"`
	SucursalNombre    *string      `json:"sucursal_nombre,omitempty"`
}

type LoginDTO struct {
	Rut      string `json:"rut" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponseDTO struct {
	Token        string             `json:"token"`
	RefreshToken string             `json:"refresh_token"`
	Usuario      UsuarioResponseDTO `json:"usuario"`
	ExpiresAt    time.Time          `json:"expires_at"`
}

type ChangePasswordDTO struct {
	PasswordActual string `json:"password_actual" binding:"required"`
	PasswordNuevo  string `json:"password_nuevo" binding:"required,min=8"`
}

// Filtros para búsqueda de usuarios

type UsuarioFilter struct {
	PaginationFilter
	SortFilter
	Rut        *string     `json:"rut,omitempty" form:"rut"`
	Nombre     *string     `json:"nombre,omitempty" form:"nombre"`
	Email      *string     `json:"email,omitempty" form:"email"`
	Rol        *RolUsuario `json:"rol,omitempty" form:"rol"`
	SucursalID *uuid.UUID  `json:"sucursal_id,omitempty" form:"sucursal_id"`
	Activo     *bool       `json:"activo,omitempty" form:"activo"`
}

// Métodos helper

func (u *Usuario) SetPassword(password string) error {
	// Generar salt
	salt := uuid.New().String()
	
	// Generar hash con bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password+salt), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	
	u.Salt = salt
	u.PasswordHash = string(hashedPassword)
	return nil
}

func (u *Usuario) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password+u.Salt))
	return err == nil
}

func (u *Usuario) IsBlocked() bool {
	return u.BloqueadoHasta != nil && u.BloqueadoHasta.After(time.Now())
}

func (u *Usuario) IncrementFailedAttempts() {
	u.IntentosFallidos++
	if u.IntentosFallidos >= 5 {
		// Bloquear por 30 minutos
		bloqueadoHasta := time.Now().Add(30 * time.Minute)
		u.BloqueadoHasta = &bloqueadoHasta
	}
}

func (u *Usuario) ResetFailedAttempts() {
	u.IntentosFallidos = 0
	u.BloqueadoHasta = nil
}

func (u *Usuario) UpdateLastAccess() {
	now := time.Now()
	u.UltimoAcceso = &now
}

func (u *Usuario) HasPermission(permission string) bool {
	if u.CachePermisos != nil && u.CachePermisos.ValidoHasta.After(time.Now()) {
		if perm, exists := u.CachePermisos.PermisosCalculados[permission]; exists {
			return perm
		}
	}
	
	// Lógica de permisos basada en rol
	switch u.Rol {
	case RolAdmin:
		return true
	case RolSupervisor:
		return permission != "gestionar_usuarios" && permission != "configurar_sistema"
	case RolCajero:
		return permission == "procesar_ventas" || permission == "consultar_productos"
	case RolVendedor:
		return permission == "crear_notas_venta" || permission == "consultar_productos" || permission == "consultar_stock"
	case RolDespacho:
		return permission == "gestionar_despachos" || permission == "consultar_productos"
	case RolOperadorEtiquetas:
		return permission == "gestionar_etiquetas" || permission == "consultar_productos"
	default:
		return false
	}
}

func (u *Usuario) ToResponseDTO() UsuarioResponseDTO {
	dto := UsuarioResponseDTO{
		ID:                    u.ID,
		Rut:                   u.Rut,
		Nombre:                u.Nombre,
		Apellido:              u.Apellido,
		Email:                 u.Email,
		Telefono:              u.Telefono,
		Rol:                   u.Rol,
		SucursalID:            u.SucursalID,
		Activo:                u.Activo,
		UltimoAcceso:          u.UltimoAcceso,
		BloqueadoHasta:        u.BloqueadoHasta,
		ConfiguracionPersonal: u.ConfiguracionPersonal,
		PermisosEspeciales:    u.PermisosEspeciales,
		FechaCreacion:         u.FechaCreacion,
		FechaModificacion:     u.FechaModificacion,
	}
	
	if u.Sucursal != nil {
		sucursalDTO := u.Sucursal.ToListDTO()
		dto.Sucursal = &sucursalDTO
	}
	
	return dto
}

func (u *Usuario) ToListDTO() UsuarioListDTO {
	dto := UsuarioListDTO{
		ID:                u.ID,
		Rut:               u.Rut,
		Nombre:            u.Nombre,
		Apellido:          u.Apellido,
		Email:             u.Email,
		Rol:               u.Rol,
		SucursalID:        u.SucursalID,
		Activo:            u.Activo,
		UltimoAcceso:      u.UltimoAcceso,
		FechaCreacion:     u.FechaCreacion,
		FechaModificacion: u.FechaModificacion,
	}
	
	if u.Sucursal != nil {
		dto.SucursalNombre = &u.Sucursal.Nombre
	}
	
	return dto
}

func (dto *UsuarioCreateDTO) ToModel() (*Usuario, error) {
	usuario := &Usuario{
		Rut:                   dto.Rut,
		Nombre:                dto.Nombre,
		Apellido:              dto.Apellido,
		Email:                 dto.Email,
		Telefono:              dto.Telefono,
		Rol:                   dto.Rol,
		SucursalID:            dto.SucursalID,
		Activo:                true,
		ConfiguracionPersonal: dto.ConfiguracionPersonal,
		PermisosEspeciales:    dto.PermisosEspeciales,
	}
	
	// Establecer contraseña
	if err := usuario.SetPassword(dto.Password); err != nil {
		return nil, fmt.Errorf("error al establecer contraseña: %w", err)
	}
	
	return usuario, nil
}

// Validaciones personalizadas

func (u *Usuario) Validate() error {
	// Validar formato RUT (implementación básica)
	if len(u.Rut) < 9 || len(u.Rut) > 12 {
		return fmt.Errorf("formato de RUT inválido")
	}
	
	// Validar email si está presente
	if u.Email != nil && *u.Email != "" {
		// Validación básica de email (en producción usar una librería)
		if len(*u.Email) < 5 || !contains(*u.Email, "@") {
			return fmt.Errorf("formato de email inválido")
		}
	}
	
	return nil
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || (len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

