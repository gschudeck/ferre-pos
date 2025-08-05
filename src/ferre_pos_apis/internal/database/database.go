package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
	
	"ferre_pos_apis/internal/config"
	"ferre_pos_apis/internal/logger"
)

// DB interfaz de base de datos
type DB interface {
	Close() error
	Ping() error
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	Stats() sql.DBStats
}

// Database estructura principal de base de datos
type Database struct {
	db     *sql.DB
	logger logger.Logger
	config *config.DatabaseConfig
}

var globalDB *Database

// Init inicializa la conexión a la base de datos
func Init(cfg *config.DatabaseConfig, log logger.Logger) (*Database, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("error abriendo conexión a base de datos: %w", err)
	}

	// Configurar pool de conexiones
	db.SetMaxOpenConns(cfg.MaxOpenConnections)
	db.SetMaxIdleConns(cfg.MaxIdleConnections)
	db.SetConnMaxLifetime(cfg.ConnectionMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnectionMaxIdleTime)

	// Verificar conexión
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("error conectando a base de datos: %w", err)
	}

	database := &Database{
		db:     db,
		logger: log,
		config: cfg,
	}

	globalDB = database
	
	log.Info("Conexión a base de datos establecida exitosamente")
	return database, nil
}

// Get obtiene la instancia global de base de datos
func Get() *Database {
	if globalDB == nil {
		panic("base de datos no inicializada. Llamar Init() primero")
	}
	return globalDB
}

// Close cierra la conexión a la base de datos
func (d *Database) Close() error {
	if d.db != nil {
		return d.db.Close()
	}
	return nil
}

// Ping verifica la conexión a la base de datos
func (d *Database) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return d.db.PingContext(ctx)
}
// BeginTx inicia una nueva transacción
func (d *Database) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return d.db.BeginTx(ctx, opts)
	}
	
	// ExecContext ejecuta una query sin retornar filas
func (d *Database) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	result, err := d.db.ExecContext(ctx, query, args...)
	duration := time.Since(start)

	d.logQuery("EXEC", query, args, duration, err)
	return result, err
}

// QueryContext ejecuta una query que retorna filas
func (d *Database) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	start := time.Now()
	rows, err := d.db.QueryContext(ctx, query, args...)
	duration := time.Since(start)

	d.logQuery("QUERY", query, args, duration, err)
	return rows, err
}

// QueryRowContext ejecuta una query que retorna una sola fila
func (d *Database) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	start := time.Now()
	row := d.db.QueryRowContext(ctx, query, args...)
	duration := time.Since(start)

	d.logQuery("QUERY_ROW", query, args, duration, nil)
	return row
}

// PrepareContext prepara una statement
func (d *Database) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return d.db.PrepareContext(ctx, query)
}

// Stats retorna estadísticas de la base de datos
func (d *Database) Stats() sql.DBStats {
	return d.db.Stats()
}

// GetDB retorna la instancia de *sql.DB para casos especiales
func (d *Database) GetDB() *sql.DB {
	return d.db
}

// logQuery registra información de queries ejecutadas
func (d *Database) logQuery(operation, query string, args []interface{}, duration time.Duration, err error) {
	fields := map[string]interface{}{
		"operation":    operation,
		"duration_ms":  duration.Milliseconds(),
		"query":        query,
		"args_count":   len(args),
	}

	if err != nil {
		// Log de error con más detalles
		if pqErr, ok := err.(*pq.Error); ok {
			fields["pg_error_code"] = pqErr.Code
			fields["pg_error_severity"] = pqErr.Severity
			fields["pg_error_detail"] = pqErr.Detail
		}
		d.logger.WithFields(fields).WithError(err).Error("Database query failed")
	} else {
		// Log normal solo para queries lentas o en modo debug
		if duration > 100*time.Millisecond {
			d.logger.WithFields(fields).Warn("Slow database query")
		} else if d.logger != nil {
			d.logger.WithFields(fields).Debug("Database query executed")
		}
	}
}

// Transaction helper para ejecutar operaciones en transacción
func (d *Database) Transaction(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := d.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error iniciando transacción: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				d.logger.WithError(rbErr).Error("Error haciendo rollback de transacción")
			}
		} else {
			if commitErr := tx.Commit(); commitErr != nil {
				err = fmt.Errorf("error haciendo commit de transacción: %w", commitErr)
			}
		}
	}()

	err = fn(tx)
	return err
}

// HealthCheck verifica el estado de la base de datos
func (d *Database) HealthCheck(ctx context.Context) error {
	// Verificar conexión básica
	if err := d.db.PingContext(ctx); err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}

	// Verificar estadísticas del pool
	stats := d.db.Stats()
	if stats.OpenConnections == 0 {
		return fmt.Errorf("no hay conexiones abiertas")
	}

	// Query simple para verificar funcionalidad
	var result int
	err := d.db.QueryRowContext(ctx, "SELECT 1").Scan(&result)
	if err != nil {
		return fmt.Errorf("query de prueba falló: %w", err)
	}

	if result != 1 {
		return fmt.Errorf("query de prueba retornó resultado inesperado: %d", result)
	}

	return nil
}

// GetConnectionStats retorna estadísticas de conexiones
func (d *Database) GetConnectionStats() map[string]interface{} {
	stats := d.db.Stats()
	return map[string]interface{}{
		"max_open_connections":     stats.MaxOpenConnections,
		"open_connections":         stats.OpenConnections,
		"in_use":                   stats.InUse,
		"idle":                     stats.Idle,
		"wait_count":               stats.WaitCount,
		"wait_duration_ms":         stats.WaitDuration.Milliseconds(),
		"max_idle_closed":          stats.MaxIdleClosed,
		"max_idle_time_closed":     stats.MaxIdleTimeClosed,
		"max_lifetime_closed":      stats.MaxLifetimeClosed,
	}
}

// Repository interfaz base para repositorios
type Repository interface {
	GetDB() *Database
}

// BaseRepository implementación base de repositorio
type BaseRepository struct {
	db     *Database
	logger logger.Logger
}

// NewBaseRepository crea un nuevo repositorio base
func NewBaseRepository(db *Database, logger logger.Logger) *BaseRepository {
	return &BaseRepository{
		db:     db,
		logger: logger,
	}
}

// GetDB retorna la instancia de base de datos
func (r *BaseRepository) GetDB() *Database {
	return r.db
}

// ExecWithRetry ejecuta una query con reintentos en caso de error temporal
func (r *BaseRepository) ExecWithRetry(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	maxRetries := 3
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		result, err := r.db.ExecContext(ctx, query, args...)
		if err == nil {
			return result, nil
		}

		lastErr = err

		// Verificar si es un error temporal
		if !isTemporaryError(err) {
			break
		}

		// Esperar antes del siguiente intento
		if attempt < maxRetries-1 {
			waitTime := time.Duration(attempt+1) * 100 * time.Millisecond
			r.logger.WithFields(map[string]interface{}{
				"attempt":   attempt + 1,
				"wait_time": waitTime,
			}).WithError(err).Warn("Reintentando query después de error temporal")
			
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(waitTime):
			}
		}
	}

	return nil, fmt.Errorf("query falló después de %d intentos: %w", maxRetries, lastErr)
}

// QueryWithRetry ejecuta una query de consulta con reintentos
func (r *BaseRepository) QueryWithRetry(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	maxRetries := 3
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		rows, err := r.db.QueryContext(ctx, query, args...)
		if err == nil {
			return rows, nil
		}

		lastErr = err

		if !isTemporaryError(err) {
			break
		}

		if attempt < maxRetries-1 {
			waitTime := time.Duration(attempt+1) * 100 * time.Millisecond
			r.logger.WithFields(map[string]interface{}{
				"attempt":   attempt + 1,
				"wait_time": waitTime,
			}).WithError(err).Warn("Reintentando query después de error temporal")
			
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(waitTime):
			}
		}
	}

	return nil, fmt.Errorf("query falló después de %d intentos: %w", maxRetries, lastErr)
}

// isTemporaryError determina si un error es temporal y se puede reintentar
func isTemporaryError(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok {
		// Códigos de error PostgreSQL que indican errores temporales
		switch pqErr.Code {
		case "40001": // serialization_failure
		case "40P01": // deadlock_detected
		case "53000": // insufficient_resources
		case "53100": // disk_full
		case "53200": // out_of_memory
		case "53300": // too_many_connections
			return true
		}
	}
	return false
}

// BuildWhereClause construye cláusulas WHERE dinámicas
func BuildWhereClause(conditions map[string]interface{}) (string, []interface{}) {
	if len(conditions) == 0 {
		return "", nil
	}

	var whereParts []string
	var args []interface{}
	argIndex := 1

	for column, value := range conditions {
		if value == nil {
			whereParts = append(whereParts, fmt.Sprintf("%s IS NULL", column))
		} else {
			whereParts = append(whereParts, fmt.Sprintf("%s = $%d", column, argIndex))
			args = append(args, value)
			argIndex++
		}
	}

	whereClause := "WHERE " + fmt.Sprintf("(%s)", fmt.Sprintf("%s", whereParts[0]))
	for i := 1; i < len(whereParts); i++ {
		whereClause += " AND " + whereParts[i]
	}

	return whereClause, args
}

// BuildLimitOffset construye cláusulas LIMIT y OFFSET para paginación
func BuildLimitOffset(page, perPage int) (string, []interface{}) {
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 {
		perPage = 20
	}
	if perPage > 1000 {
		perPage = 1000 // Límite máximo
	}

	offset := (page - 1) * perPage
	return "LIMIT $1 OFFSET $2", []interface{}{perPage, offset}
}

// ScanRows helper para escanear múltiples filas
func ScanRows(rows *sql.Rows, scanFunc func() error) error {
	defer rows.Close()

	for rows.Next() {
		if err := scanFunc(); err != nil {
			return err
		}
	}

	return rows.Err()
}

