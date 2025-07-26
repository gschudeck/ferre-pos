package repositories

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"ferre-pos-servidor-central/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BaseRepository contiene funcionalidades comunes para todos los repositorios
type BaseRepository struct {
	db *gorm.DB
}

// NewBaseRepository crea una nueva instancia del repositorio base
func NewBaseRepository(db *gorm.DB) *BaseRepository {
	return &BaseRepository{db: db}
}

// Create crea un nuevo registro
func (br *BaseRepository) Create(entity interface{}) error {
	return br.db.Create(entity).Error
}

// GetByID obtiene un registro por ID
func (br *BaseRepository) GetByID(id uuid.UUID, entity interface{}) error {
	return br.db.First(entity, "id = ?", id).Error
}

// Update actualiza un registro
func (br *BaseRepository) Update(entity interface{}) error {
	return br.db.Save(entity).Error
}

// Delete elimina un registro (soft delete)
func (br *BaseRepository) Delete(id uuid.UUID, entity interface{}) error {
	return br.db.Delete(entity, "id = ?", id).Error
}

// HardDelete elimina permanentemente un registro
func (br *BaseRepository) HardDelete(id uuid.UUID, entity interface{}) error {
	return br.db.Unscoped().Delete(entity, "id = ?", id).Error
}

// Exists verifica si un registro existe
func (br *BaseRepository) Exists(id uuid.UUID, model interface{}) (bool, error) {
	var count int64
	err := br.db.Model(model).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

// ExistsByField verifica si un registro existe por un campo específico
func (br *BaseRepository) ExistsByField(field string, value interface{}, model interface{}) (bool, error) {
	var count int64
	err := br.db.Model(model).Where(fmt.Sprintf("%s = ?", field), value).Count(&count).Error
	return count > 0, err
}

// GetAll obtiene todos los registros con paginación
func (br *BaseRepository) GetAll(entities interface{}, filter models.PaginationFilter) error {
	query := br.db.Model(entities)

	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}

	if filter.Page > 1 && filter.Limit > 0 {
		offset := (filter.Page - 1) * filter.Limit
		query = query.Offset(offset)
	}

	return query.Find(entities).Error
}

// Count cuenta el total de registros
func (br *BaseRepository) Count(model interface{}) (int64, error) {
	var count int64
	err := br.db.Model(model).Count(&count).Error
	return count, err
}

// CountWithCondition cuenta registros con condición
func (br *BaseRepository) CountWithCondition(model interface{}, condition string, args ...interface{}) (int64, error) {
	var count int64
	err := br.db.Model(model).Where(condition, args...).Count(&count).Error
	return count, err
}

// FindWithCondition busca registros con condición
func (br *BaseRepository) FindWithCondition(entities interface{}, condition string, args ...interface{}) error {
	return br.db.Where(condition, args...).Find(entities).Error
}

// FindOneWithCondition busca un registro con condición
func (br *BaseRepository) FindOneWithCondition(entity interface{}, condition string, args ...interface{}) error {
	return br.db.Where(condition, args...).First(entity).Error
}

// ApplyFilters aplica filtros dinámicos a una consulta
func (br *BaseRepository) ApplyFilters(query *gorm.DB, filters interface{}) *gorm.DB {
	v := reflect.ValueOf(filters)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// Saltar campos nil o vacíos
		if !field.IsValid() || field.IsZero() {
			continue
		}

		// Obtener tag de base de datos
		dbTag := fieldType.Tag.Get("db")
		if dbTag == "" {
			dbTag = fieldType.Tag.Get("json")
		}
		if dbTag == "" {
			continue
		}

		// Aplicar filtro según el tipo
		switch field.Kind() {
		case reflect.Ptr:
			if !field.IsNil() {
				query = br.applyFieldFilter(query, dbTag, field.Elem().Interface())
			}
		default:
			query = br.applyFieldFilter(query, dbTag, field.Interface())
		}
	}

	return query
}

// applyFieldFilter aplica un filtro específico
func (br *BaseRepository) applyFieldFilter(query *gorm.DB, field string, value interface{}) *gorm.DB {
	switch v := value.(type) {
	case string:
		if strings.Contains(field, "nombre") || strings.Contains(field, "descripcion") {
			// Búsqueda parcial para campos de texto
			return query.Where(fmt.Sprintf("%s ILIKE ?", field), "%"+v+"%")
		}
		return query.Where(fmt.Sprintf("%s = ?", field), v)
	case uuid.UUID:
		return query.Where(fmt.Sprintf("%s = ?", field), v)
	case bool:
		return query.Where(fmt.Sprintf("%s = ?", field), v)
	case int, int32, int64:
		return query.Where(fmt.Sprintf("%s = ?", field), v)
	case float32, float64:
		return query.Where(fmt.Sprintf("%s = ?", field), v)
	default:
		return query.Where(fmt.Sprintf("%s = ?", field), v)
	}
}

// ApplySorting aplica ordenamiento a una consulta
func (br *BaseRepository) ApplySorting(query *gorm.DB, sort models.SortFilter) *gorm.DB {
	if sort.SortBy != "" {
		order := sort.SortBy
		if sort.SortOrder != "" {
			order += " " + sort.SortOrder
		}
		query = query.Order(order)
	} else {
		// Ordenamiento por defecto
		query = query.Order("created_at DESC")
	}

	return query
}

// ApplyDateRange aplica filtro de rango de fechas
func (br *BaseRepository) ApplyDateRange(query *gorm.DB, dateRange models.DateRangeFilter, dateField string) *gorm.DB {
	if dateRange.FechaInicio != nil {
		query = query.Where(fmt.Sprintf("%s >= ?", dateField), *dateRange.FechaInicio)
	}

	if dateRange.FechaFin != nil {
		query = query.Where(fmt.Sprintf("%s <= ?", dateField), *dateRange.FechaFin)
	}

	return query
}

// GetPaginatedResults obtiene resultados paginados con conteo total
func (br *BaseRepository) GetPaginatedResults(
	model interface{},
	results interface{},
	filter models.PaginationFilter,
	queryBuilder func(*gorm.DB) *gorm.DB,
) (models.PaginationResponse, error) {

	// Construir consulta base
	baseQuery := br.db.Model(model)
	if queryBuilder != nil {
		baseQuery = queryBuilder(baseQuery)
	}

	// Contar total de registros
	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return models.PaginationResponse{}, err
	}

	// Aplicar paginación
	query := baseQuery
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)

		if filter.Page > 1 {
			offset := (filter.Page - 1) * filter.Limit
			query = query.Offset(offset)
		}
	}

	// Obtener resultados
	if err := query.Find(results).Error; err != nil {
		return models.PaginationResponse{}, err
	}

	// Calcular información de paginación
	totalPages := int(total) / filter.Limit
	if int(total)%filter.Limit > 0 {
		totalPages++
	}

	return models.PaginationResponse{
		IntPage:       filter.Page,
		IntPageSize:   filter.Limit,
		IntTotal:      int64(total),
		IntTotalPages: totalPages,
		BoolHasNext:   filter.Page < totalPages,
		BoolHasPrev:   filter.Page > 1,
	}, nil
}

// BatchCreate crea múltiples registros en lote
func (br *BaseRepository) BatchCreate(entities interface{}, batchSize int) error {
	return br.db.CreateInBatches(entities, batchSize).Error
}

// BatchUpdate actualiza múltiples registros
func (br *BaseRepository) BatchUpdate(model interface{}, updates map[string]interface{}, condition string, args ...interface{}) error {
	return br.db.Model(model).Where(condition, args...).Updates(updates).Error
}

// Transaction ejecuta operaciones dentro de una transacción
func (br *BaseRepository) Transaction(fn func(*gorm.DB) error) error {
	return br.db.Transaction(fn)
}

// WithTransaction retorna un nuevo repositorio con una transacción específica
func (br *BaseRepository) WithTransaction(tx *gorm.DB) *BaseRepository {
	return &BaseRepository{db: tx}
}

// GetDB retorna la instancia de base de datos
func (br *BaseRepository) GetDB() *gorm.DB {
	return br.db
}

// RawQuery ejecuta una consulta SQL cruda
func (br *BaseRepository) RawQuery(sql string, dest interface{}, args ...interface{}) error {
	return br.db.Raw(sql, args...).Scan(dest).Error
}

// Exec ejecuta una consulta SQL sin retornar resultados
func (br *BaseRepository) Exec(sql string, args ...interface{}) error {
	return br.db.Exec(sql, args...).Error
}

// GetLastInsertID obtiene el último ID insertado
func (br *BaseRepository) GetLastInsertID(entity interface{}) (uuid.UUID, error) {
	// Usar reflexión para obtener el campo ID
	v := reflect.ValueOf(entity)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	idField := v.FieldByName("ID")
	if !idField.IsValid() {
		return uuid.Nil, fmt.Errorf("campo ID no encontrado")
	}

	if idField.Type() != reflect.TypeOf(uuid.UUID{}) {
		return uuid.Nil, fmt.Errorf("campo ID no es de tipo UUID")
	}

	return idField.Interface().(uuid.UUID), nil
}

// BulkInsert inserta múltiples registros de forma optimizada
func (br *BaseRepository) BulkInsert(tableName string, columns []string, values [][]interface{}) error {
	if len(values) == 0 {
		return nil
	}

	// Construir consulta SQL
	placeholders := make([]string, len(values))
	args := make([]interface{}, 0, len(values)*len(columns))

	for i, row := range values {
		rowPlaceholders := make([]string, len(columns))
		for j := range columns {
			rowPlaceholders[j] = "?"
			args = append(args, row[j])
		}
		placeholders[i] = "(" + strings.Join(rowPlaceholders, ",") + ")"
	}

	sql := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES %s",
		tableName,
		strings.Join(columns, ","),
		strings.Join(placeholders, ","),
	)

	return br.db.Exec(sql, args...).Error
}

// Search realiza búsqueda de texto completo
func (br *BaseRepository) Search(model interface{}, results interface{}, searchTerm string, searchFields []string, limit int) error {
	if searchTerm == "" {
		return br.db.Model(model).Limit(limit).Find(results).Error
	}

	query := br.db.Model(model)

	// Construir condiciones de búsqueda
	conditions := make([]string, len(searchFields))
	args := make([]interface{}, len(searchFields))

	for i, field := range searchFields {
		conditions[i] = fmt.Sprintf("%s ILIKE ?", field)
		args[i] = "%" + searchTerm + "%"
	}

	whereClause := strings.Join(conditions, " OR ")
	query = query.Where(whereClause, args...)

	if limit > 0 {
		query = query.Limit(limit)
	}

	return query.Find(results).Error
}

// GetDistinct obtiene valores únicos de un campo
func (br *BaseRepository) GetDistinct(model interface{}, field string, results interface{}) error {
	return br.db.Model(model).Distinct(field).Pluck(field, results).Error
}

// GetRandomRecords obtiene registros aleatorios
func (br *BaseRepository) GetRandomRecords(model interface{}, results interface{}, limit int) error {
	return br.db.Model(model).Order("RANDOM()").Limit(limit).Find(results).Error
}

// Aggregate ejecuta funciones de agregación
func (br *BaseRepository) Aggregate(model interface{}, aggregateFunc string, field string) (float64, error) {
	var result float64
	sql := fmt.Sprintf("SELECT %s(%s) FROM %s", aggregateFunc, field, br.getTableName(model))
	err := br.db.Raw(sql).Scan(&result).Error
	return result, err
}

// getTableName obtiene el nombre de la tabla de un modelo
func (br *BaseRepository) getTableName(model interface{}) string {
	stmt := &gorm.Statement{DB: br.db}
	stmt.Parse(model)
	return stmt.Schema.Table
}

// WithPreload agrega preload a la consulta
func (br *BaseRepository) WithPreload(query *gorm.DB, preloads []string) *gorm.DB {
	for _, preload := range preloads {
		query = query.Preload(preload)
	}
	return query
}

// WithJoins agrega joins a la consulta
func (br *BaseRepository) WithJoins(query *gorm.DB, joins []string) *gorm.DB {
	for _, join := range joins {
		query = query.Joins(join)
	}
	return query
}

// GetWithRelations obtiene un registro con sus relaciones
func (br *BaseRepository) GetWithRelations(id uuid.UUID, entity interface{}, relations []string) error {
	query := br.db
	for _, relation := range relations {
		query = query.Preload(relation)
	}
	return query.First(entity, "id = ?", id).Error
}

// FindWithRelations busca registros con sus relaciones
func (br *BaseRepository) FindWithRelations(entities interface{}, relations []string, condition string, args ...interface{}) error {
	query := br.db
	for _, relation := range relations {
		query = query.Preload(relation)
	}
	return query.Where(condition, args...).Find(entities).Error
}

// SoftDelete realiza eliminación suave
func (br *BaseRepository) SoftDelete(id uuid.UUID, model interface{}) error {
	return br.db.Delete(model, "id = ?", id).Error
}

// Restore restaura un registro eliminado suavemente
func (br *BaseRepository) Restore(id uuid.UUID, model interface{}) error {
	return br.db.Unscoped().Model(model).Where("id = ?", id).Update("deleted_at", nil).Error
}

// GetDeleted obtiene registros eliminados suavemente
func (br *BaseRepository) GetDeleted(entities interface{}, model interface{}) error {
	return br.db.Unscoped().Where("deleted_at IS NOT NULL").Find(entities).Error
}

// CleanupDeleted limpia registros eliminados suavemente después de cierto tiempo
func (br *BaseRepository) CleanupDeleted(model interface{}, daysOld int) error {
	cutoffDate := time.Now().AddDate(0, 0, -daysOld)
	return br.db.Unscoped().Where("deleted_at < ?", cutoffDate).Delete(model).Error
}
