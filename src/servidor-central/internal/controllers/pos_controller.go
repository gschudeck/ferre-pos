package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"ferre-pos-servidor-central/internal/models"
	"ferre-pos-servidor-central/internal/services"
)

// POSController maneja las operaciones del API POS
type POSController struct {
	BaseController
	authService         services.AuthService
	productoService     services.ProductoService
	stockService        services.StockService
	ventaService        services.VentaService
	fidelizacionService services.FidelizacionService
	sucursalService     services.SucursalService
}

// NewPOSController crea una nueva instancia del controlador POS
func NewPOSController(
	authService services.AuthService,
	productoService services.ProductoService,
	stockService services.StockService,
	ventaService services.VentaService,
	fidelizacionService services.FidelizacionService,
	sucursalService services.SucursalService,
) *POSController {
	return &POSController{
		authService:         authService,
		productoService:     productoService,
		stockService:        stockService,
		ventaService:        ventaService,
		fidelizacionService: fidelizacionService,
		sucursalService:     sucursalService,
	}
}

// ===== AUTENTICACIÓN =====

// Login autentica un usuario en el sistema POS
func (pc *POSController) Login(c *gin.Context) {
	var loginDTO models.LoginDTO
	if err := pc.ValidateJSON(c, &loginDTO); err != nil {
		pc.ResponseError(c, http.StatusBadRequest, "Datos de login inválidos", err)
		return
	}

	// Obtener información del terminal/sucursal
	terminalID := c.GetHeader("X-Terminal-ID")
	sucursalID := c.GetHeader("X-Sucursal-ID")

	// Autenticar usuario
	authResponse, err := pc.authService.Login(loginDTO.Email, loginDTO.Password, terminalID, sucursalID)
	if err != nil {
		pc.ResponseError(c, http.StatusUnauthorized, "Credenciales inválidas", err)
		return
	}

	pc.LogActivity(c, "login", gin.H{
		"terminal_id": terminalID,
		"sucursal_id": sucursalID,
	})

	pc.ResponseSuccess(c, authResponse)
}

// Logout cierra sesión del usuario
func (pc *POSController) Logout(c *gin.Context) {
	user, err := pc.GetUserFromContext(c)
	if err != nil {
		pc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	token := c.GetHeader("Authorization")
	if err := pc.authService.Logout(token); err != nil {
		pc.ResponseError(c, http.StatusInternalServerError, "Error al cerrar sesión", err)
		return
	}

	pc.LogActivity(c, "logout", gin.H{
		"user_id": user.ID,
	})

	pc.ResponseSuccess(c, gin.H{"message": "Sesión cerrada exitosamente"})
}

// RefreshToken renueva el token de autenticación
func (pc *POSController) RefreshToken(c *gin.Context) {
	var refreshDTO models.RefreshTokenDTO
	if err := pc.ValidateJSON(c, &refreshDTO); err != nil {
		pc.ResponseError(c, http.StatusBadRequest, "Token de refresh inválido", err)
		return
	}

	authResponse, err := pc.authService.RefreshToken(refreshDTO.RefreshToken)
	if err != nil {
		pc.ResponseError(c, http.StatusUnauthorized, "Token de refresh inválido", err)
		return
	}

	pc.ResponseSuccess(c, authResponse)
}

// ===== PRODUCTOS =====

// GetProductos obtiene lista de productos para POS
func (pc *POSController) GetProductos(c *gin.Context) {
	sucursalID, err := pc.GetSucursalFromContext(c)
	if err != nil {
		pc.ResponseError(c, http.StatusBadRequest, "Sucursal requerida", err)
		return
	}

	// Parsear filtros
	filter := models.ProductoFilter{
		PaginationFilter: pc.ParsePagination(c),
		SortFilter:       pc.ParseSort(c),
		SucursalID:       sucursalID,
		Activo:           &[]bool{true}[0], // Solo productos activos
	}

	// Filtros específicos
	if categoria := c.Query("categoria"); categoria != "" {
		filter.Categoria = &categoria
	}
	if codigoBarra := c.Query("codigo_barra"); codigoBarra != "" {
		filter.CodigoBarra = &codigoBarra
	}
	if nombre := c.Query("nombre"); nombre != "" {
		filter.Nombre = &nombre
	}
	if conStock := c.Query("con_stock"); conStock == "true" {
		filter.ConStock = &[]bool{true}[0]
	}

	productos, pagination, err := pc.productoService.GetProductosPOS(filter)
	if err != nil {
		pc.ResponseError(c, http.StatusInternalServerError, "Error al obtener productos", err)
		return
	}

	pc.SetCacheHeaders(c, 300) // Cache por 5 minutos
	pc.ResponsePaginated(c, productos, pagination)
}

// GetProducto obtiene un producto específico por ID o código de barra
func (pc *POSController) GetProducto(c *gin.Context) {
	sucursalID, err := pc.GetSucursalFromContext(c)
	if err != nil {
		pc.ResponseError(c, http.StatusBadRequest, "Sucursal requerida", err)
		return
	}

	// Puede ser ID o código de barra
	identifier := c.Param("id")
	
	var producto *models.ProductoPOSResponseDTO
	
	// Intentar parsear como UUID primero
	if id, err := uuid.Parse(identifier); err == nil {
		producto, err = pc.productoService.GetProductoPOSByID(id, sucursalID)
	} else {
		// Si no es UUID, buscar por código de barra
		producto, err = pc.productoService.GetProductoPOSByCodigoBarra(identifier, sucursalID)
	}

	if err != nil {
		pc.ResponseError(c, http.StatusNotFound, "Producto no encontrado", err)
		return
	}

	pc.SetCacheHeaders(c, 300) // Cache por 5 minutos
	pc.ResponseSuccess(c, producto)
}

// BuscarProductos busca productos por texto
func (pc *POSController) BuscarProductos(c *gin.Context) {
	sucursalID, err := pc.GetSucursalFromContext(c)
	if err != nil {
		pc.ResponseError(c, http.StatusBadRequest, "Sucursal requerida", err)
		return
	}

	query := c.Query("q")
	if query == "" {
		pc.ResponseError(c, http.StatusBadRequest, "Parámetro de búsqueda requerido", nil)
		return
	}

	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 50 {
			limit = l
		}
	}

	productos, err := pc.productoService.BuscarProductosPOS(query, sucursalID, limit)
	if err != nil {
		pc.ResponseError(c, http.StatusInternalServerError, "Error en búsqueda", err)
		return
	}

	pc.SetCacheHeaders(c, 60) // Cache por 1 minuto
	pc.ResponseSuccess(c, productos)
}

// ===== STOCK =====

// GetStockProducto obtiene el stock de un producto
func (pc *POSController) GetStockProducto(c *gin.Context) {
	productoID, err := pc.ParseUUID(c, "producto_id")
	if err != nil {
		pc.ResponseError(c, http.StatusBadRequest, "ID de producto inválido", err)
		return
	}

	sucursalID, err := pc.GetSucursalFromContext(c)
	if err != nil {
		pc.ResponseError(c, http.StatusBadRequest, "Sucursal requerida", err)
		return
	}

	stock, err := pc.stockService.GetStockProducto(productoID, sucursalID)
	if err != nil {
		pc.ResponseError(c, http.StatusNotFound, "Stock no encontrado", err)
		return
	}

	pc.SetCacheHeaders(c, 60) // Cache por 1 minuto
	pc.ResponseSuccess(c, stock)
}

// ReservarStock reserva stock para una venta
func (pc *POSController) ReservarStock(c *gin.Context) {
	var reservaDTO models.ReservaStockDTO
	if err := pc.ValidateJSON(c, &reservaDTO); err != nil {
		pc.ResponseError(c, http.StatusBadRequest, "Datos de reserva inválidos", err)
		return
	}

	user, err := pc.GetUserFromContext(c)
	if err != nil {
		pc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	sucursalID, err := pc.GetSucursalFromContext(c)
	if err != nil {
		pc.ResponseError(c, http.StatusBadRequest, "Sucursal requerida", err)
		return
	}

	reserva, err := pc.stockService.ReservarStock(reservaDTO, user.ID, sucursalID)
	if err != nil {
		pc.ResponseError(c, http.StatusBadRequest, "Error al reservar stock", err)
		return
	}

	pc.LogActivity(c, "reservar_stock", reserva)
	pc.ResponseCreated(c, reserva)
}

// LiberarReserva libera una reserva de stock
func (pc *POSController) LiberarReserva(c *gin.Context) {
	reservaID, err := pc.ParseUUID(c, "reserva_id")
	if err != nil {
		pc.ResponseError(c, http.StatusBadRequest, "ID de reserva inválido", err)
		return
	}

	user, err := pc.GetUserFromContext(c)
	if err != nil {
		pc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	if err := pc.stockService.LiberarReserva(reservaID, user.ID); err != nil {
		pc.ResponseError(c, http.StatusBadRequest, "Error al liberar reserva", err)
		return
	}

	pc.LogActivity(c, "liberar_reserva", gin.H{"reserva_id": reservaID})
	pc.ResponseSuccess(c, gin.H{"message": "Reserva liberada exitosamente"})
}

// ===== VENTAS =====

// CrearVenta crea una nueva venta
func (pc *POSController) CrearVenta(c *gin.Context) {
	var ventaDTO models.VentaCreateDTO
	if err := pc.ValidateJSON(c, &ventaDTO); err != nil {
		pc.ResponseError(c, http.StatusBadRequest, "Datos de venta inválidos", err)
		return
	}

	user, err := pc.GetUserFromContext(c)
	if err != nil {
		pc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	sucursalID, err := pc.GetSucursalFromContext(c)
	if err != nil {
		pc.ResponseError(c, http.StatusBadRequest, "Sucursal requerida", err)
		return
	}

	// Obtener información del terminal
	terminalID := c.GetHeader("X-Terminal-ID")

	venta, err := pc.ventaService.CrearVenta(ventaDTO, user.ID, sucursalID, terminalID)
	if err != nil {
		pc.ResponseError(c, http.StatusBadRequest, "Error al crear venta", err)
		return
	}

	pc.LogActivity(c, "crear_venta", venta)
	pc.ResponseCreated(c, venta)
}

// GetVenta obtiene una venta específica
func (pc *POSController) GetVenta(c *gin.Context) {
	ventaID, err := pc.ParseUUID(c, "id")
	if err != nil {
		pc.ResponseError(c, http.StatusBadRequest, "ID de venta inválido", err)
		return
	}

	user, err := pc.GetUserFromContext(c)
	if err != nil {
		pc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	sucursalID, err := pc.GetSucursalFromContext(c)
	if err != nil {
		pc.ResponseError(c, http.StatusBadRequest, "Sucursal requerida", err)
		return
	}

	venta, err := pc.ventaService.GetVenta(ventaID, user.ID, sucursalID)
	if err != nil {
		pc.ResponseError(c, http.StatusNotFound, "Venta no encontrada", err)
		return
	}

	pc.ResponseSuccess(c, venta)
}

// GetVentas obtiene lista de ventas
func (pc *POSController) GetVentas(c *gin.Context) {
	user, err := pc.GetUserFromContext(c)
	if err != nil {
		pc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	sucursalID, err := pc.GetSucursalFromContext(c)
	if err != nil {
		pc.ResponseError(c, http.StatusBadRequest, "Sucursal requerida", err)
		return
	}

	// Parsear filtros
	filter := models.VentaFilter{
		PaginationFilter: pc.ParsePagination(c),
		SortFilter:       pc.ParseSort(c),
		DateRangeFilter:  pc.ParseDateRange(c),
		SucursalID:       sucursalID,
	}

	// Solo mostrar ventas del usuario si no es admin/gerente
	if !pc.CheckPermission(user, "ver_todas_ventas") {
		filter.UsuarioID = &user.ID
	}

	// Filtros adicionales
	if estado := c.Query("estado"); estado != "" {
		estadoVenta := models.EstadoVenta(estado)
		filter.Estado = &estadoVenta
	}

	ventas, pagination, err := pc.ventaService.GetVentas(filter)
	if err != nil {
		pc.ResponseError(c, http.StatusInternalServerError, "Error al obtener ventas", err)
		return
	}

	pc.ResponsePaginated(c, ventas, pagination)
}

// AnularVenta anula una venta
func (pc *POSController) AnularVenta(c *gin.Context) {
	ventaID, err := pc.ParseUUID(c, "id")
	if err != nil {
		pc.ResponseError(c, http.StatusBadRequest, "ID de venta inválido", err)
		return
	}

	var anulacionDTO models.AnulacionVentaDTO
	if err := pc.ValidateJSON(c, &anulacionDTO); err != nil {
		pc.ResponseError(c, http.StatusBadRequest, "Datos de anulación inválidos", err)
		return
	}

	user, err := pc.GetUserFromContext(c)
	if err != nil {
		pc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	// Verificar permisos para anular ventas
	if !pc.CheckPermission(user, "anular_ventas") {
		pc.ResponseError(c, http.StatusForbidden, "Sin permisos para anular ventas", nil)
		return
	}

	if err := pc.ventaService.AnularVenta(ventaID, anulacionDTO, user.ID); err != nil {
		pc.ResponseError(c, http.StatusBadRequest, "Error al anular venta", err)
		return
	}

	pc.LogActivity(c, "anular_venta", gin.H{
		"venta_id": ventaID,
		"motivo":   anulacionDTO.Motivo,
	})

	pc.ResponseSuccess(c, gin.H{"message": "Venta anulada exitosamente"})
}

// ===== FIDELIZACIÓN =====

// GetCliente obtiene información de un cliente
func (pc *POSController) GetCliente(c *gin.Context) {
	clienteID, err := pc.ParseUUID(c, "id")
	if err != nil {
		pc.ResponseError(c, http.StatusBadRequest, "ID de cliente inválido", err)
		return
	}

	sucursalID, err := pc.GetSucursalFromContext(c)
	if err != nil {
		pc.ResponseError(c, http.StatusBadRequest, "Sucursal requerida", err)
		return
	}

	cliente, err := pc.fidelizacionService.GetCliente(clienteID, sucursalID)
	if err != nil {
		pc.ResponseError(c, http.StatusNotFound, "Cliente no encontrado", err)
		return
	}

	pc.ResponseSuccess(c, cliente)
}

// BuscarClientes busca clientes por texto
func (pc *POSController) BuscarClientes(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		pc.ResponseError(c, http.StatusBadRequest, "Parámetro de búsqueda requerido", nil)
		return
	}

	sucursalID, err := pc.GetSucursalFromContext(c)
	if err != nil {
		pc.ResponseError(c, http.StatusBadRequest, "Sucursal requerida", err)
		return
	}

	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 20 {
			limit = l
		}
	}

	clientes, err := pc.fidelizacionService.BuscarClientes(query, sucursalID, limit)
	if err != nil {
		pc.ResponseError(c, http.StatusInternalServerError, "Error en búsqueda", err)
		return
	}

	pc.ResponseSuccess(c, clientes)
}

// CrearCliente crea un nuevo cliente
func (pc *POSController) CrearCliente(c *gin.Context) {
	var clienteDTO models.ClienteCreateDTO
	if err := pc.ValidateJSON(c, &clienteDTO); err != nil {
		pc.ResponseError(c, http.StatusBadRequest, "Datos de cliente inválidos", err)
		return
	}

	user, err := pc.GetUserFromContext(c)
	if err != nil {
		pc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	sucursalID, err := pc.GetSucursalFromContext(c)
	if err != nil {
		pc.ResponseError(c, http.StatusBadRequest, "Sucursal requerida", err)
		return
	}

	cliente, err := pc.fidelizacionService.CrearCliente(clienteDTO, user.ID, sucursalID)
	if err != nil {
		pc.ResponseError(c, http.StatusBadRequest, "Error al crear cliente", err)
		return
	}

	pc.LogActivity(c, "crear_cliente", cliente)
	pc.ResponseCreated(c, cliente)
}

// GetPuntosCliente obtiene los puntos de fidelización de un cliente
func (pc *POSController) GetPuntosCliente(c *gin.Context) {
	clienteID, err := pc.ParseUUID(c, "cliente_id")
	if err != nil {
		pc.ResponseError(c, http.StatusBadRequest, "ID de cliente inválido", err)
		return
	}

	sucursalID, err := pc.GetSucursalFromContext(c)
	if err != nil {
		pc.ResponseError(c, http.StatusBadRequest, "Sucursal requerida", err)
		return
	}

	puntos, err := pc.fidelizacionService.GetPuntosCliente(clienteID, sucursalID)
	if err != nil {
		pc.ResponseError(c, http.StatusNotFound, "Puntos no encontrados", err)
		return
	}

	pc.ResponseSuccess(c, puntos)
}

// CanjearPuntos canjea puntos de fidelización
func (pc *POSController) CanjearPuntos(c *gin.Context) {
	var canjeDTO models.CanjePuntosDTO
	if err := pc.ValidateJSON(c, &canjeDTO); err != nil {
		pc.ResponseError(c, http.StatusBadRequest, "Datos de canje inválidos", err)
		return
	}

	user, err := pc.GetUserFromContext(c)
	if err != nil {
		pc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	sucursalID, err := pc.GetSucursalFromContext(c)
	if err != nil {
		pc.ResponseError(c, http.StatusBadRequest, "Sucursal requerida", err)
		return
	}

	canje, err := pc.fidelizacionService.CanjearPuntos(canjeDTO, user.ID, sucursalID)
	if err != nil {
		pc.ResponseError(c, http.StatusBadRequest, "Error al canjear puntos", err)
		return
	}

	pc.LogActivity(c, "canjear_puntos", canje)
	pc.ResponseSuccess(c, canje)
}

// ===== CONFIGURACIÓN =====

// GetConfiguracionSucursal obtiene la configuración de la sucursal
func (pc *POSController) GetConfiguracionSucursal(c *gin.Context) {
	sucursalID, err := pc.GetSucursalFromContext(c)
	if err != nil {
		pc.ResponseError(c, http.StatusBadRequest, "Sucursal requerida", err)
		return
	}

	configuracion, err := pc.sucursalService.GetConfiguracionPOS(*sucursalID)
	if err != nil {
		pc.ResponseError(c, http.StatusNotFound, "Configuración no encontrada", err)
		return
	}

	pc.SetCacheHeaders(c, 3600) // Cache por 1 hora
	pc.ResponseSuccess(c, configuracion)
}

// GetEstadisticasVentas obtiene estadísticas de ventas del día
func (pc *POSController) GetEstadisticasVentas(c *gin.Context) {
	user, err := pc.GetUserFromContext(c)
	if err != nil {
		pc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	sucursalID, err := pc.GetSucursalFromContext(c)
	if err != nil {
		pc.ResponseError(c, http.StatusBadRequest, "Sucursal requerida", err)
		return
	}

	fecha := time.Now().Format("2006-01-02")
	if fechaParam := c.Query("fecha"); fechaParam != "" {
		fecha = fechaParam
	}

	estadisticas, err := pc.ventaService.GetEstadisticasDia(fecha, user.ID, sucursalID)
	if err != nil {
		pc.ResponseError(c, http.StatusInternalServerError, "Error al obtener estadísticas", err)
		return
	}

	pc.SetCacheHeaders(c, 300) // Cache por 5 minutos
	pc.ResponseSuccess(c, estadisticas)
}

// ===== UTILIDADES =====

// ValidarConexion valida la conexión del terminal
func (pc *POSController) ValidarConexion(c *gin.Context) {
	terminalID := c.GetHeader("X-Terminal-ID")
	sucursalID := c.GetHeader("X-Sucursal-ID")

	if terminalID == "" || sucursalID == "" {
		pc.ResponseError(c, http.StatusBadRequest, "Headers de terminal y sucursal requeridos", nil)
		return
	}

	// Validar que el terminal existe y está activo
	if err := pc.sucursalService.ValidarTerminal(terminalID, sucursalID); err != nil {
		pc.ResponseError(c, http.StatusUnauthorized, "Terminal no autorizado", err)
		return
	}

	pc.ResponseSuccess(c, gin.H{
		"status":      "connected",
		"terminal_id": terminalID,
		"sucursal_id": sucursalID,
		"timestamp":   time.Now(),
	})
}

// SincronizarDatos sincroniza datos del terminal con el servidor
func (pc *POSController) SincronizarDatos(c *gin.Context) {
	var syncDTO models.SincronizacionDTO
	if err := pc.ValidateJSON(c, &syncDTO); err != nil {
		pc.ResponseError(c, http.StatusBadRequest, "Datos de sincronización inválidos", err)
		return
	}

	user, err := pc.GetUserFromContext(c)
	if err != nil {
		pc.ResponseError(c, http.StatusUnauthorized, "Usuario no autenticado", err)
		return
	}

	sucursalID, err := pc.GetSucursalFromContext(c)
	if err != nil {
		pc.ResponseError(c, http.StatusBadRequest, "Sucursal requerida", err)
		return
	}

	resultado, err := pc.ventaService.SincronizarDatos(syncDTO, user.ID, sucursalID)
	if err != nil {
		pc.ResponseError(c, http.StatusInternalServerError, "Error en sincronización", err)
		return
	}

	pc.LogActivity(c, "sincronizar_datos", resultado)
	pc.ResponseSuccess(c, resultado)
}

