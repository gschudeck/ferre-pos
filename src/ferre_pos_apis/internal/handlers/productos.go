package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	
	"ferre_pos_apis/internal/database"
	"ferre_pos_apis/internal/logger"
	"ferre_pos_apis/internal/metrics"
	"ferre_pos_apis/internal/models"
	"ferre_pos_apis/pkg/validator"
)

// ProductosHandler handler para operaciones de productos
type ProductosHandler struct {
	db        *database.Database
	logger    logger.Logger
	validator validator.Validator
	metrics   *metrics.Metrics
}

// NewProductosHandler crea un nuevo handler de productos
func NewProductosHandler(db *database.Database, log logger.Logger, val validator.Validator, met *metrics.Metrics) *ProductosHandler {
	return &ProductosHandler{
		db:        db,
		logger:    log,
		validator: val,
		metrics:   met,
	}
}

// List lista productos con filtros y paginación
func (h *ProductosHandler) List(c *gin.Context) {
	start := time.Now()
	defer func() {
		if h.metrics != nil {
			h.metrics.RecordProductoConsulta("pos", getSucursalID(c), "list")
		}
	}()

	// Parámetros de consulta
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	categoriaID := c.Query("categoria_id")
	activos := c.DefaultQuery("activos", "true") == "true"
	conStock := c.Query("con_stock") == "true"
	sucursalID := c.Query("sucursal_id")

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Construir query
	baseQuery := `
		SELECT p.id, p.codigo_interno, p.codigo_barra, p.descripcion, p.descripcion_corta,
		       p.categoria_id, p.marca, p.modelo, p.precio_unitario, p.precio_costo,
		       p.unidad_medida, p.peso, p.dimensiones, p.especificaciones_tecnicas,
		       p.activo, p.requiere_serie, p.permite_fraccionamiento, p.stock_minimo,
		       p.stock_maximo, p.imagen_principal_url, p.imagenes_adicionales,
		       p.fecha_creacion, p.fecha_modificacion, p.usuario_creacion,
		       p.usuario_modificacion, p.popularidad_score, p.cache_codigo_barras_generado,
		       p.configuracion_etiqueta, p.fecha_ultima_etiqueta, p.total_etiquetas_generadas,
		       c.nombre as categoria_nombre,
		       COALESCE(s.cantidad_disponible, 0) as stock_disponible
		FROM productos p
		LEFT JOIN categorias_productos c ON p.categoria_id = c.id
		LEFT JOIN stock_central s ON p.id = s.producto_id AND s.sucursal_id = $1`

	countQuery := `
		SELECT COUNT(*)
		FROM productos p
		LEFT JOIN stock_central s ON p.id = s.producto_id AND s.sucursal_id = $1`

	var conditions []string
	var args []interface{}
	argIndex := 2

	// Filtro por sucursal (siempre requerido)
	if sucursalID == "" {
		sucursalID = getUserSucursalID(c)
	}
	args = append([]interface{}{sucursalID}, args...)

	// Filtros adicionales
	if activos {
		conditions = append(conditions, fmt.Sprintf("p.activo = $%d", argIndex))
		args = append(args, true)
		argIndex++
	}

	if categoriaID != "" {
		conditions = append(conditions, fmt.Sprintf("p.categoria_id = $%d", argIndex))
		args = append(args, categoriaID)
		argIndex++
	}

	if conStock {
		conditions = append(conditions, fmt.Sprintf("COALESCE(s.cantidad_disponible, 0) > 0"))
	}

	// Agregar condiciones WHERE
	if len(conditions) > 0 {
		whereClause := " WHERE " + strings.Join(conditions, " AND ")
		baseQuery += whereClause
		countQuery += whereClause
	}

	// Obtener total de registros
	var total int
	err := h.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		h.logger.WithError(err).Error("Error obteniendo total de productos")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "Error consultando productos",
			},
			RequestID: getRequestID(c),
			Timestamp: time.Now(),
		})
		return
	}

	// Agregar ordenamiento y paginación
	offset := (page - 1) * perPage
	baseQuery += fmt.Sprintf(" ORDER BY p.popularidad_score DESC, p.descripcion ASC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, perPage, offset)

	// Ejecutar query principal
	rows, err := h.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		h.logger.WithError(err).Error("Error consultando productos")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "Error consultando productos",
			},
			RequestID: getRequestID(c),
			Timestamp: time.Now(),
		})
		return
	}
	defer rows.Close()

	var productos []map[string]interface{}
	for rows.Next() {
		var p models.Producto
		var categoriaNombre sql.NullString
		var stockDisponible int

		err := rows.Scan(
			&p.ID, &p.CodigoInterno, &p.CodigoBarra, &p.Descripcion, &p.DescripcionCorta,
			&p.CategoriaID, &p.Marca, &p.Modelo, &p.PrecioUnitario, &p.PrecioCosto,
			&p.UnidadMedida, &p.Peso, &p.Dimensiones, &p.EspecificacionesTecnicas,
			&p.Activo, &p.RequiereSerie, &p.PermiteFraccionamiento, &p.StockMinimo,
			&p.StockMaximo, &p.ImagenPrincipalURL, &p.ImagenesAdicionales,
			&p.FechaCreacion, &p.FechaModificacion, &p.UsuarioCreacion,
			&p.UsuarioModificacion, &p.PopularidadScore, &p.CacheCodigoBarrasGenerado,
			&p.ConfiguracionEtiqueta, &p.FechaUltimaEtiqueta, &p.TotalEtiquetasGeneradas,
			&categoriaNombre, &stockDisponible,
		)
		if err != nil {
			h.logger.WithError(err).Error("Error escaneando producto")
			continue
		}

		producto := map[string]interface{}{
			"id":                         p.ID,
			"codigo_interno":             p.CodigoInterno,
			"codigo_barra":               p.CodigoBarra,
			"descripcion":                p.Descripcion,
			"descripcion_corta":          p.DescripcionCorta,
			"categoria_id":               p.CategoriaID,
			"categoria_nombre":           categoriaNombre.String,
			"marca":                      p.Marca,
			"modelo":                     p.Modelo,
			"precio_unitario":            p.PrecioUnitario,
			"precio_costo":               p.PrecioCosto,
			"unidad_medida":              p.UnidadMedida,
			"peso":                       p.Peso,
			"dimensiones":                p.Dimensiones,
			"especificaciones_tecnicas":  p.EspecificacionesTecnicas,
			"activo":                     p.Activo,
			"requiere_serie":             p.RequiereSerie,
			"permite_fraccionamiento":    p.PermiteFraccionamiento,
			"stock_minimo":               p.StockMinimo,
			"stock_maximo":               p.StockMaximo,
			"stock_disponible":           stockDisponible,
			"imagen_principal_url":       p.ImagenPrincipalURL,
			"imagenes_adicionales":       p.ImagenesAdicionales,
			"popularidad_score":          p.PopularidadScore,
			"configuracion_etiqueta":     p.ConfiguracionEtiqueta,
			"fecha_creacion":             p.FechaCreacion,
			"fecha_modificacion":         p.FechaModificacion,
		}

		productos = append(productos, producto)
	}

	if err = rows.Err(); err != nil {
		h.logger.WithError(err).Error("Error iterando productos")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "Error procesando productos",
			},
			RequestID: getRequestID(c),
			Timestamp: time.Now(),
		})
		return
	}

	// Calcular metadatos de paginación
	totalPages := (total + perPage - 1) / perPage

	h.logger.WithField("duration_ms", time.Since(start).Milliseconds()).
		WithField("total", total).
		WithField("page", page).
		Debug("Productos listados exitosamente")

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    productos,
		Meta: &models.APIMeta{
			Page:       page,
			PerPage:    perPage,
			Total:      total,
			TotalPages: totalPages,
		},
		RequestID: getRequestID(c),
		Timestamp: time.Now(),
	})
}

// GetByID obtiene un producto por ID
func (h *ProductosHandler) GetByID(c *gin.Context) {
	start := time.Now()
	defer func() {
		if h.metrics != nil {
			h.metrics.RecordProductoConsulta("pos", getSucursalID(c), "get_by_id")
		}
	}()

	productID := c.Param("id")
	if productID == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "MISSING_PRODUCT_ID",
				Message: "ID de producto requerido",
			},
			RequestID: getRequestID(c),
			Timestamp: time.Now(),
		})
		return
	}

	// Validar UUID
	if _, err := uuid.Parse(productID); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_PRODUCT_ID",
				Message: "ID de producto inválido",
			},
			RequestID: getRequestID(c),
			Timestamp: time.Now(),
		})
		return
	}

	sucursalID := c.Query("sucursal_id")
	if sucursalID == "" {
		sucursalID = getUserSucursalID(c)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	producto, err := h.getProductoByID(ctx, productID, sucursalID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "PRODUCT_NOT_FOUND",
					Message: "Producto no encontrado",
				},
				RequestID: getRequestID(c),
				Timestamp: time.Now(),
			})
			return
		}

		h.logger.WithError(err).Error("Error consultando producto")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "Error consultando producto",
			},
			RequestID: getRequestID(c),
			Timestamp: time.Now(),
		})
		return
	}

	h.logger.WithField("duration_ms", time.Since(start).Milliseconds()).
		WithField("product_id", productID).
		Debug("Producto consultado exitosamente")

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      producto,
		RequestID: getRequestID(c),
		Timestamp: time.Now(),
	})
}

// GetByBarcode obtiene un producto por código de barras
func (h *ProductosHandler) GetByBarcode(c *gin.Context) {
	start := time.Now()
	defer func() {
		if h.metrics != nil {
			h.metrics.RecordProductoConsulta("pos", getSucursalID(c), "get_by_barcode")
		}
	}()

	codigo := c.Param("codigo")
	if codigo == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "MISSING_BARCODE",
				Message: "Código de barras requerido",
			},
			RequestID: getRequestID(c),
			Timestamp: time.Now(),
		})
		return
	}

	sucursalID := c.Query("sucursal_id")
	if sucursalID == "" {
		sucursalID = getUserSucursalID(c)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	producto, err := h.getProductoByBarcode(ctx, codigo, sucursalID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "PRODUCT_NOT_FOUND",
					Message: "Producto no encontrado con ese código de barras",
				},
				RequestID: getRequestID(c),
				Timestamp: time.Now(),
			})
			return
		}

		h.logger.WithError(err).Error("Error consultando producto por código de barras")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "Error consultando producto",
			},
			RequestID: getRequestID(c),
			Timestamp: time.Now(),
		})
		return
	}

	h.logger.WithField("duration_ms", time.Since(start).Milliseconds()).
		WithField("barcode", codigo).
		Debug("Producto consultado por código de barras exitosamente")

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      producto,
		RequestID: getRequestID(c),
		Timestamp: time.Now(),
	})
}

// Search busca productos por texto
func (h *ProductosHandler) Search(c *gin.Context) {
	start := time.Now()
	defer func() {
		if h.metrics != nil {
			h.metrics.RecordProductoConsulta("pos", getSucursalID(c), "search")
		}
	}()

	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "MISSING_SEARCH_QUERY",
				Message: "Parámetro de búsqueda 'q' requerido",
			},
			RequestID: getRequestID(c),
			Timestamp: time.Now(),
		})
		return
	}

	// Parámetros adicionales
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	sucursalID := c.Query("sucursal_id")
	conStock := c.Query("con_stock") == "true"

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 50 {
		perPage = 20
	}

	if sucursalID == "" {
		sucursalID = getUserSucursalID(c)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	productos, total, err := h.searchProductos(ctx, query, sucursalID, conStock, page, perPage)
	if err != nil {
		h.logger.WithError(err).Error("Error buscando productos")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "SEARCH_ERROR",
				Message: "Error realizando búsqueda",
			},
			RequestID: getRequestID(c),
			Timestamp: time.Now(),
		})
		return
	}

	totalPages := (total + perPage - 1) / perPage

	h.logger.WithField("duration_ms", time.Since(start).Milliseconds()).
		WithField("query", query).
		WithField("results", len(productos)).
		Debug("Búsqueda de productos completada")

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    productos,
		Meta: &models.APIMeta{
			Page:       page,
			PerPage:    perPage,
			Total:      total,
			TotalPages: totalPages,
		},
		RequestID: getRequestID(c),
		Timestamp: time.Now(),
	})
}

// Create crea un nuevo producto (solo admin/supervisor)
func (h *ProductosHandler) Create(c *gin.Context) {
	var req models.Producto
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_JSON",
				Message: "JSON inválido en el cuerpo de la petición",
			},
			RequestID: getRequestID(c),
			Timestamp: time.Now(),
		})
		return
	}

	// Validar datos
	if err := h.validator.ValidateStruct(&req); err != nil {
		validationErrors := h.validator.GetValidationErrors(err)
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "VALIDATION_ERROR",
				Message: "Error de validación en los datos enviados",
				Details: models.JSONB{"validation_errors": validationErrors},
			},
			RequestID: getRequestID(c),
			Timestamp: time.Now(),
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Generar ID y establecer campos de auditoría
	req.ID = uuid.New()
	req.FechaCreacion = time.Now()
	req.FechaModificacion = time.Now()
	
	if userID := getUserID(c); userID != "" {
		if uid, err := uuid.Parse(userID); err == nil {
			req.UsuarioCreacion = &uid
			req.UsuarioModificacion = &uid
		}
	}

	// Crear producto
	err := h.createProducto(ctx, &req)
	if err != nil {
		h.logger.WithError(err).Error("Error creando producto")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "CREATE_ERROR",
				Message: "Error creando producto",
			},
			RequestID: getRequestID(c),
			Timestamp: time.Now(),
		})
		return
	}

	h.logger.WithField("product_id", req.ID).Info("Producto creado exitosamente")

	c.JSON(http.StatusCreated, models.APIResponse{
		Success:   true,
		Data:      req,
		RequestID: getRequestID(c),
		Timestamp: time.Now(),
	})
}

// Update actualiza un producto existente
func (h *ProductosHandler) Update(c *gin.Context) {
	productID := c.Param("id")
	if _, err := uuid.Parse(productID); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_PRODUCT_ID",
				Message: "ID de producto inválido",
			},
			RequestID: getRequestID(c),
			Timestamp: time.Now(),
		})
		return
	}

	var req models.Producto
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_JSON",
				Message: "JSON inválido en el cuerpo de la petición",
			},
			RequestID: getRequestID(c),
			Timestamp: time.Now(),
		})
		return
	}

	// Validar datos
	if err := h.validator.ValidateStruct(&req); err != nil {
		validationErrors := h.validator.GetValidationErrors(err)
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "VALIDATION_ERROR",
				Message: "Error de validación en los datos enviados",
				Details: models.JSONB{"validation_errors": validationErrors},
			},
			RequestID: getRequestID(c),
			Timestamp: time.Now(),
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Establecer ID y campos de auditoría
	req.ID, _ = uuid.Parse(productID)
	req.FechaModificacion = time.Now()
	
	if userID := getUserID(c); userID != "" {
		if uid, err := uuid.Parse(userID); err == nil {
			req.UsuarioModificacion = &uid
		}
	}

	// Actualizar producto
	err := h.updateProducto(ctx, &req)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "PRODUCT_NOT_FOUND",
					Message: "Producto no encontrado",
				},
				RequestID: getRequestID(c),
				Timestamp: time.Now(),
			})
			return
		}

		h.logger.WithError(err).Error("Error actualizando producto")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UPDATE_ERROR",
				Message: "Error actualizando producto",
			},
			RequestID: getRequestID(c),
			Timestamp: time.Now(),
		})
		return
	}

	h.logger.WithField("product_id", req.ID).Info("Producto actualizado exitosamente")

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      req,
		RequestID: getRequestID(c),
		Timestamp: time.Now(),
	})
}

// Delete elimina un producto (solo admin)
func (h *ProductosHandler) Delete(c *gin.Context) {
	productID := c.Param("id")
	if _, err := uuid.Parse(productID); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_PRODUCT_ID",
				Message: "ID de producto inválido",
			},
			RequestID: getRequestID(c),
			Timestamp: time.Now(),
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// En lugar de eliminar físicamente, marcar como inactivo
	err := h.deactivateProducto(ctx, productID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "PRODUCT_NOT_FOUND",
					Message: "Producto no encontrado",
				},
				RequestID: getRequestID(c),
				Timestamp: time.Now(),
			})
			return
		}

		h.logger.WithError(err).Error("Error desactivando producto")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DELETE_ERROR",
				Message: "Error eliminando producto",
			},
			RequestID: getRequestID(c),
			Timestamp: time.Now(),
		})
		return
	}

	h.logger.WithField("product_id", productID).Info("Producto desactivado exitosamente")

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      gin.H{"message": "Producto eliminado exitosamente"},
		RequestID: getRequestID(c),
		Timestamp: time.Now(),
	})
}

// Métodos auxiliares

// getProductoByID obtiene un producto por ID con información de stock
func (h *ProductosHandler) getProductoByID(ctx context.Context, productID, sucursalID string) (map[string]interface{}, error) {
	query := `
		SELECT p.id, p.codigo_interno, p.codigo_barra, p.descripcion, p.descripcion_corta,
		       p.categoria_id, p.marca, p.modelo, p.precio_unitario, p.precio_costo,
		       p.unidad_medida, p.peso, p.dimensiones, p.especificaciones_tecnicas,
		       p.activo, p.requiere_serie, p.permite_fraccionamiento, p.stock_minimo,
		       p.stock_maximo, p.imagen_principal_url, p.imagenes_adicionales,
		       p.fecha_creacion, p.fecha_modificacion, p.popularidad_score,
		       p.configuracion_etiqueta, p.fecha_ultima_etiqueta, p.total_etiquetas_generadas,
		       c.nombre as categoria_nombre,
		       COALESCE(s.cantidad_disponible, 0) as stock_disponible,
		       COALESCE(s.cantidad, 0) as stock_total,
		       COALESCE(s.cantidad_reservada, 0) as stock_reservado
		FROM productos p
		LEFT JOIN categorias_productos c ON p.categoria_id = c.id
		LEFT JOIN stock_central s ON p.id = s.producto_id AND s.sucursal_id = $2
		WHERE p.id = $1`

	var p models.Producto
	var categoriaNombre sql.NullString
	var stockDisponible, stockTotal, stockReservado int

	err := h.db.QueryRowContext(ctx, query, productID, sucursalID).Scan(
		&p.ID, &p.CodigoInterno, &p.CodigoBarra, &p.Descripcion, &p.DescripcionCorta,
		&p.CategoriaID, &p.Marca, &p.Modelo, &p.PrecioUnitario, &p.PrecioCosto,
		&p.UnidadMedida, &p.Peso, &p.Dimensiones, &p.EspecificacionesTecnicas,
		&p.Activo, &p.RequiereSerie, &p.PermiteFraccionamiento, &p.StockMinimo,
		&p.StockMaximo, &p.ImagenPrincipalURL, &p.ImagenesAdicionales,
		&p.FechaCreacion, &p.FechaModificacion, &p.PopularidadScore,
		&p.ConfiguracionEtiqueta, &p.FechaUltimaEtiqueta, &p.TotalEtiquetasGeneradas,
		&categoriaNombre, &stockDisponible, &stockTotal, &stockReservado,
	)

	if err != nil {
		return nil, err
	}

	producto := map[string]interface{}{
		"id":                         p.ID,
		"codigo_interno":             p.CodigoInterno,
		"codigo_barra":               p.CodigoBarra,
		"descripcion":                p.Descripcion,
		"descripcion_corta":          p.DescripcionCorta,
		"categoria_id":               p.CategoriaID,
		"categoria_nombre":           categoriaNombre.String,
		"marca":                      p.Marca,
		"modelo":                     p.Modelo,
		"precio_unitario":            p.PrecioUnitario,
		"precio_costo":               p.PrecioCosto,
		"unidad_medida":              p.UnidadMedida,
		"peso":                       p.Peso,
		"dimensiones":                p.Dimensiones,
		"especificaciones_tecnicas":  p.EspecificacionesTecnicas,
		"activo":                     p.Activo,
		"requiere_serie":             p.RequiereSerie,
		"permite_fraccionamiento":    p.PermiteFraccionamiento,
		"stock_minimo":               p.StockMinimo,
		"stock_maximo":               p.StockMaximo,
		"stock_disponible":           stockDisponible,
		"stock_total":                stockTotal,
		"stock_reservado":            stockReservado,
		"imagen_principal_url":       p.ImagenPrincipalURL,
		"imagenes_adicionales":       p.ImagenesAdicionales,
		"popularidad_score":          p.PopularidadScore,
		"configuracion_etiqueta":     p.ConfiguracionEtiqueta,
		"fecha_ultima_etiqueta":      p.FechaUltimaEtiqueta,
		"total_etiquetas_generadas":  p.TotalEtiquetasGeneradas,
		"fecha_creacion":             p.FechaCreacion,
		"fecha_modificacion":         p.FechaModificacion,
	}

	return producto, nil
}

// getProductoByBarcode obtiene un producto por código de barras
func (h *ProductosHandler) getProductoByBarcode(ctx context.Context, codigo, sucursalID string) (map[string]interface{}, error) {
	// Buscar en tabla principal y códigos adicionales
	query := `
		SELECT p.id
		FROM productos p
		WHERE p.codigo_barra = $1 AND p.activo = true
		UNION
		SELECT p.id
		FROM productos p
		JOIN codigos_barra_adicionales cba ON p.id = cba.producto_id
		WHERE cba.codigo_barra = $1 AND cba.activo = true AND p.activo = true
		LIMIT 1`

	var productID string
	err := h.db.QueryRowContext(ctx, query, codigo).Scan(&productID)
	if err != nil {
		return nil, err
	}

	return h.getProductoByID(ctx, productID, sucursalID)
}

// searchProductos busca productos por texto
func (h *ProductosHandler) searchProductos(ctx context.Context, query, sucursalID string, conStock bool, page, perPage int) ([]map[string]interface{}, int, error) {
	// Preparar términos de búsqueda
	searchTerms := strings.Fields(strings.ToLower(query))
	if len(searchTerms) == 0 {
		return nil, 0, fmt.Errorf("términos de búsqueda vacíos")
	}

	// Construir query de búsqueda
	baseQuery := `
		SELECT p.id, p.codigo_interno, p.codigo_barra, p.descripcion, p.descripcion_corta,
		       p.categoria_id, p.marca, p.modelo, p.precio_unitario, p.unidad_medida,
		       p.imagen_principal_url, p.popularidad_score,
		       c.nombre as categoria_nombre,
		       COALESCE(s.cantidad_disponible, 0) as stock_disponible
		FROM productos p
		LEFT JOIN categorias_productos c ON p.categoria_id = c.id
		LEFT JOIN stock_central s ON p.id = s.producto_id AND s.sucursal_id = $1
		WHERE p.activo = true AND (`

	countQuery := `
		SELECT COUNT(*)
		FROM productos p
		LEFT JOIN stock_central s ON p.id = s.producto_id AND s.sucursal_id = $1
		WHERE p.activo = true AND (`

	// Construir condiciones de búsqueda
	var searchConditions []string
	var args []interface{}
	args = append(args, sucursalID)
	argIndex := 2

	for _, term := range searchTerms {
		condition := fmt.Sprintf(`(
			LOWER(p.descripcion) ILIKE $%d OR
			LOWER(p.codigo_interno) ILIKE $%d OR
			LOWER(p.codigo_barra) ILIKE $%d OR
			LOWER(p.marca) ILIKE $%d OR
			LOWER(p.modelo) ILIKE $%d
		)`, argIndex, argIndex, argIndex, argIndex, argIndex)
		
		searchConditions = append(searchConditions, condition)
		args = append(args, "%"+term+"%")
		argIndex++
	}

	searchClause := strings.Join(searchConditions, " AND ")
	baseQuery += searchClause + ")"
	countQuery += searchClause + ")"

	// Agregar filtro de stock si es necesario
	if conStock {
		baseQuery += " AND COALESCE(s.cantidad_disponible, 0) > 0"
		countQuery += " AND COALESCE(s.cantidad_disponible, 0) > 0"
	}

	// Obtener total
	var total int
	err := h.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Agregar ordenamiento y paginación
	offset := (page - 1) * perPage
	baseQuery += fmt.Sprintf(" ORDER BY p.popularidad_score DESC, p.descripcion ASC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, perPage, offset)

	// Ejecutar búsqueda
	rows, err := h.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var productos []map[string]interface{}
	for rows.Next() {
		var p struct {
			ID                uuid.UUID
			CodigoInterno     string
			CodigoBarra       string
			Descripcion       string
			DescripcionCorta  *string
			CategoriaID       *uuid.UUID
			Marca             *string
			Modelo            *string
			PrecioUnitario    float64
			UnidadMedida      string
			ImagenPrincipalURL *string
			PopularidadScore  float64
		}
		var categoriaNombre sql.NullString
		var stockDisponible int

		err := rows.Scan(
			&p.ID, &p.CodigoInterno, &p.CodigoBarra, &p.Descripcion, &p.DescripcionCorta,
			&p.CategoriaID, &p.Marca, &p.Modelo, &p.PrecioUnitario, &p.UnidadMedida,
			&p.ImagenPrincipalURL, &p.PopularidadScore,
			&categoriaNombre, &stockDisponible,
		)
		if err != nil {
			continue
		}

		producto := map[string]interface{}{
			"id":                   p.ID,
			"codigo_interno":       p.CodigoInterno,
			"codigo_barra":         p.CodigoBarra,
			"descripcion":          p.Descripcion,
			"descripcion_corta":    p.DescripcionCorta,
			"categoria_id":         p.CategoriaID,
			"categoria_nombre":     categoriaNombre.String,
			"marca":                p.Marca,
			"modelo":               p.Modelo,
			"precio_unitario":      p.PrecioUnitario,
			"unidad_medida":        p.UnidadMedida,
			"stock_disponible":     stockDisponible,
			"imagen_principal_url": p.ImagenPrincipalURL,
			"popularidad_score":    p.PopularidadScore,
		}

		productos = append(productos, producto)
	}

	return productos, total, rows.Err()
}

// createProducto crea un nuevo producto
func (h *ProductosHandler) createProducto(ctx context.Context, producto *models.Producto) error {
	query := `
		INSERT INTO productos (
			id, codigo_interno, codigo_barra, descripcion, descripcion_corta,
			categoria_id, marca, modelo, precio_unitario, precio_costo,
			unidad_medida, peso, dimensiones, especificaciones_tecnicas,
			activo, requiere_serie, permite_fraccionamiento, stock_minimo,
			stock_maximo, imagen_principal_url, imagenes_adicionales,
			fecha_creacion, fecha_modificacion, usuario_creacion, usuario_modificacion,
			popularidad_score, configuracion_etiqueta
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15, $16, $17, $18, $19, $20,
			$21, $22, $23, $24, $25, $26, $27
		)`

	_, err := h.db.ExecContext(ctx, query,
		producto.ID, producto.CodigoInterno, producto.CodigoBarra, producto.Descripcion,
		producto.DescripcionCorta, producto.CategoriaID, producto.Marca, producto.Modelo,
		producto.PrecioUnitario, producto.PrecioCosto, producto.UnidadMedida, producto.Peso,
		producto.Dimensiones, producto.EspecificacionesTecnicas, producto.Activo,
		producto.RequiereSerie, producto.PermiteFraccionamiento, producto.StockMinimo,
		producto.StockMaximo, producto.ImagenPrincipalURL, producto.ImagenesAdicionales,
		producto.FechaCreacion, producto.FechaModificacion, producto.UsuarioCreacion,
		producto.UsuarioModificacion, producto.PopularidadScore, producto.ConfiguracionEtiqueta,
	)

	return err
}

// updateProducto actualiza un producto existente
func (h *ProductosHandler) updateProducto(ctx context.Context, producto *models.Producto) error {
	query := `
		UPDATE productos SET
			codigo_interno = $2, codigo_barra = $3, descripcion = $4, descripcion_corta = $5,
			categoria_id = $6, marca = $7, modelo = $8, precio_unitario = $9, precio_costo = $10,
			unidad_medida = $11, peso = $12, dimensiones = $13, especificaciones_tecnicas = $14,
			activo = $15, requiere_serie = $16, permite_fraccionamiento = $17, stock_minimo = $18,
			stock_maximo = $19, imagen_principal_url = $20, imagenes_adicionales = $21,
			fecha_modificacion = $22, usuario_modificacion = $23, configuracion_etiqueta = $24
		WHERE id = $1`

	result, err := h.db.ExecContext(ctx, query,
		producto.ID, producto.CodigoInterno, producto.CodigoBarra, producto.Descripcion,
		producto.DescripcionCorta, producto.CategoriaID, producto.Marca, producto.Modelo,
		producto.PrecioUnitario, producto.PrecioCosto, producto.UnidadMedida, producto.Peso,
		producto.Dimensiones, producto.EspecificacionesTecnicas, producto.Activo,
		producto.RequiereSerie, producto.PermiteFraccionamiento, producto.StockMinimo,
		producto.StockMaximo, producto.ImagenPrincipalURL, producto.ImagenesAdicionales,
		producto.FechaModificacion, producto.UsuarioModificacion, producto.ConfiguracionEtiqueta,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// deactivateProducto desactiva un producto
func (h *ProductosHandler) deactivateProducto(ctx context.Context, productID string) error {
	query := `UPDATE productos SET activo = false, fecha_modificacion = NOW() WHERE id = $1`
	
	result, err := h.db.ExecContext(ctx, query, productID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// Funciones auxiliares para obtener datos del contexto

func getUserID(c *gin.Context) string {
	if userID, exists := c.Get("user_id"); exists {
		return userID.(string)
	}
	return ""
}

func getUserSucursalID(c *gin.Context) string {
	if sucursalID, exists := c.Get("sucursal_id"); exists {
		return sucursalID.(string)
	}
	return ""
}

func getSucursalID(c *gin.Context) string {
	sucursalID := c.Query("sucursal_id")
	if sucursalID == "" {
		sucursalID = getUserSucursalID(c)
	}
	return sucursalID
}

