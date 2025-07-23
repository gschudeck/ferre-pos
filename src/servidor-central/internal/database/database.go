package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	_ "github.com/lib/pq"
)

// DatabaseConfig contiene la configuración de la base de datos
type DatabaseConfig struct {
	Host            string `yaml:"host" json:"host"`
	Port            int    `yaml:"port" json:"port"`
	User            string `yaml:"user" json:"user"`
	Password        string `yaml:"password" json:"password"`
	Database        string `yaml:"database" json:"database"`
	SSLMode         string `yaml:"ssl_mode" json:"ssl_mode"`
	MaxOpenConns    int    `yaml:"max_open_conns" json:"max_open_conns"`
	MaxIdleConns    int    `yaml:"max_idle_conns" json:"max_idle_conns"`
	ConnMaxLifetime int    `yaml:"conn_max_lifetime" json:"conn_max_lifetime"` // en minutos
	ConnMaxIdleTime int    `yaml:"conn_max_idle_time" json:"conn_max_idle_time"` // en minutos
	LogLevel        string `yaml:"log_level" json:"log_level"`
	SlowThreshold   int    `yaml:"slow_threshold" json:"slow_threshold"` // en milisegundos
}

// Database contiene las conexiones a la base de datos
type Database struct {
	GORM *gorm.DB
	SQL  *sql.DB
}

// NewDatabase crea una nueva instancia de base de datos
func NewDatabase(config DatabaseConfig) (*Database, error) {
	// Construir DSN
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host,
		config.Port,
		config.User,
		config.Password,
		config.Database,
		config.SSLMode,
	)

	// Configurar logger de GORM
	logLevel := logger.Silent
	switch config.LogLevel {
	case "error":
		logLevel = logger.Error
	case "warn":
		logLevel = logger.Warn
	case "info":
		logLevel = logger.Info
	}

	gormLogger := logger.New(
		log.New(log.Writer(), "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Duration(config.SlowThreshold) * time.Millisecond,
			LogLevel:                  logLevel,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	// Conectar con GORM
	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		PrepareStmt: true,
	})
	if err != nil {
		return nil, fmt.Errorf("error conectando con GORM: %w", err)
	}

	// Obtener conexión SQL subyacente
	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, fmt.Errorf("error obteniendo conexión SQL: %w", err)
	}

	// Configurar pool de conexiones
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(config.ConnMaxLifetime) * time.Minute)
	sqlDB.SetConnMaxIdleTime(time.Duration(config.ConnMaxIdleTime) * time.Minute)

	// Verificar conexión
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("error verificando conexión: %w", err)
	}

	return &Database{
		GORM: gormDB,
		SQL:  sqlDB,
	}, nil
}

// Close cierra las conexiones de base de datos
func (db *Database) Close() error {
	if db.SQL != nil {
		return db.SQL.Close()
	}
	return nil
}

// GetStats obtiene estadísticas de la base de datos
func (db *Database) GetStats() sql.DBStats {
	if db.SQL != nil {
		return db.SQL.Stats()
	}
	return sql.DBStats{}
}

// HealthCheck verifica el estado de la base de datos
func (db *Database) HealthCheck() error {
	if db.SQL == nil {
		return fmt.Errorf("conexión de base de datos no inicializada")
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	return db.SQL.PingContext(ctx)
}

// BeginTransaction inicia una transacción
func (db *Database) BeginTransaction() *gorm.DB {
	return db.GORM.Begin()
}

// WithTransaction ejecuta una función dentro de una transacción
func (db *Database) WithTransaction(fn func(*gorm.DB) error) error {
	tx := db.BeginTransaction()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// DatabaseManager gestiona múltiples conexiones de base de datos
type DatabaseManager struct {
	databases map[string]*Database
	configs   map[string]DatabaseConfig
}

// NewDatabaseManager crea un nuevo gestor de bases de datos
func NewDatabaseManager() *DatabaseManager {
	return &DatabaseManager{
		databases: make(map[string]*Database),
		configs:   make(map[string]DatabaseConfig),
	}
}

// AddDatabase agrega una nueva configuración de base de datos
func (dm *DatabaseManager) AddDatabase(name string, config DatabaseConfig) error {
	dm.configs[name] = config
	
	db, err := NewDatabase(config)
	if err != nil {
		return fmt.Errorf("error creando base de datos %s: %w", name, err)
	}
	
	dm.databases[name] = db
	return nil
}

// GetDatabase obtiene una conexión de base de datos por nombre
func (dm *DatabaseManager) GetDatabase(name string) (*Database, error) {
	db, exists := dm.databases[name]
	if !exists {
		return nil, fmt.Errorf("base de datos %s no encontrada", name)
	}
	return db, nil
}

// CloseAll cierra todas las conexiones
func (dm *DatabaseManager) CloseAll() error {
	var errors []error
	
	for name, db := range dm.databases {
		if err := db.Close(); err != nil {
			errors = append(errors, fmt.Errorf("error cerrando %s: %w", name, err))
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("errores cerrando bases de datos: %v", errors)
	}
	
	return nil
}

// HealthCheckAll verifica el estado de todas las bases de datos
func (dm *DatabaseManager) HealthCheckAll() map[string]error {
	results := make(map[string]error)
	
	for name, db := range dm.databases {
		results[name] = db.HealthCheck()
	}
	
	return results
}

// GetStatsAll obtiene estadísticas de todas las bases de datos
func (dm *DatabaseManager) GetStatsAll() map[string]sql.DBStats {
	results := make(map[string]sql.DBStats)
	
	for name, db := range dm.databases {
		results[name] = db.GetStats()
	}
	
	return results
}

// DefaultDatabaseConfig retorna una configuración por defecto
func DefaultDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		Host:            "localhost",
		Port:            5432,
		User:            "postgres",
		Password:        "password",
		Database:        "ferre_pos",
		SSLMode:         "disable",
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 30, // 30 minutos
		ConnMaxIdleTime: 15, // 15 minutos
		LogLevel:        "warn",
		SlowThreshold:   200, // 200ms
	}
}

// DatabaseConfigForAPI retorna configuración optimizada por API
func DatabaseConfigForAPI(apiName string, baseConfig DatabaseConfig) DatabaseConfig {
	config := baseConfig
	
	switch apiName {
	case "pos":
		// API POS requiere máximo rendimiento
		config.MaxOpenConns = 50
		config.MaxIdleConns = 10
		config.ConnMaxLifetime = 15 // Conexiones más cortas para alta rotación
		config.SlowThreshold = 100  // Umbral más estricto
		
	case "sync":
		// API Sync maneja operaciones de larga duración
		config.MaxOpenConns = 20
		config.MaxIdleConns = 5
		config.ConnMaxLifetime = 60 // Conexiones más largas
		config.SlowThreshold = 500  // Umbral más permisivo
		
	case "labels":
		// API Labels tiene uso intermitente
		config.MaxOpenConns = 15
		config.MaxIdleConns = 3
		config.ConnMaxLifetime = 45
		config.SlowThreshold = 300
		
	case "reports":
		// API Reports maneja consultas complejas
		config.MaxOpenConns = 30
		config.MaxIdleConns = 8
		config.ConnMaxLifetime = 45
		config.SlowThreshold = 1000 // Consultas más lentas permitidas
	}
	
	return config
}

// MigrationConfig contiene configuración para migraciones
type MigrationConfig struct {
	AutoMigrate     bool     `yaml:"auto_migrate" json:"auto_migrate"`
	CreateTables    bool     `yaml:"create_tables" json:"create_tables"`
	DropTables      bool     `yaml:"drop_tables" json:"drop_tables"`
	SeedData        bool     `yaml:"seed_data" json:"seed_data"`
	MigrationPath   string   `yaml:"migration_path" json:"migration_path"`
	SeedPath        string   `yaml:"seed_path" json:"seed_path"`
	IgnoredTables   []string `yaml:"ignored_tables" json:"ignored_tables"`
}

// RunMigrations ejecuta las migraciones de base de datos
func (db *Database) RunMigrations(config MigrationConfig, models ...interface{}) error {
	if !config.AutoMigrate {
		return nil
	}
	
	// Eliminar tablas si está configurado
	if config.DropTables {
		for i := len(models) - 1; i >= 0; i-- {
			if err := db.GORM.Migrator().DropTable(models[i]); err != nil {
				log.Printf("Error eliminando tabla: %v", err)
			}
		}
	}
	
	// Crear/actualizar tablas
	if config.CreateTables {
		for _, model := range models {
			if err := db.GORM.AutoMigrate(model); err != nil {
				return fmt.Errorf("error en migración: %w", err)
			}
		}
	}
	
	return nil
}

// ConnectionPool gestiona un pool de conexiones personalizado
type ConnectionPool struct {
	db       *Database
	maxConns int
	conns    chan *gorm.DB
}

// NewConnectionPool crea un nuevo pool de conexiones
func NewConnectionPool(db *Database, maxConns int) *ConnectionPool {
	pool := &ConnectionPool{
		db:       db,
		maxConns: maxConns,
		conns:    make(chan *gorm.DB, maxConns),
	}
	
	// Inicializar pool con conexiones
	for i := 0; i < maxConns; i++ {
		pool.conns <- db.GORM.Session(&gorm.Session{})
	}
	
	return pool
}

// Get obtiene una conexión del pool
func (cp *ConnectionPool) Get() *gorm.DB {
	select {
	case conn := <-cp.conns:
		return conn
	default:
		// Si no hay conexiones disponibles, crear una nueva sesión
		return cp.db.GORM.Session(&gorm.Session{})
	}
}

// Put devuelve una conexión al pool
func (cp *ConnectionPool) Put(conn *gorm.DB) {
	select {
	case cp.conns <- conn:
		// Conexión devuelta al pool
	default:
		// Pool lleno, descartar conexión
	}
}

// WithConnection ejecuta una función con una conexión del pool
func (cp *ConnectionPool) WithConnection(fn func(*gorm.DB) error) error {
	conn := cp.Get()
	defer cp.Put(conn)
	return fn(conn)
}

// QueryBuilder ayuda a construir consultas dinámicas
type QueryBuilder struct {
	db     *gorm.DB
	query  *gorm.DB
	errors []error
}

// NewQueryBuilder crea un nuevo constructor de consultas
func NewQueryBuilder(db *gorm.DB) *QueryBuilder {
	return &QueryBuilder{
		db:    db,
		query: db,
	}
}

// Where agrega una condición WHERE
func (qb *QueryBuilder) Where(condition string, args ...interface{}) *QueryBuilder {
	qb.query = qb.query.Where(condition, args...)
	return qb
}

// Join agrega un JOIN
func (qb *QueryBuilder) Join(table string, condition string) *QueryBuilder {
	qb.query = qb.query.Joins(fmt.Sprintf("JOIN %s ON %s", table, condition))
	return qb
}

// LeftJoin agrega un LEFT JOIN
func (qb *QueryBuilder) LeftJoin(table string, condition string) *QueryBuilder {
	qb.query = qb.query.Joins(fmt.Sprintf("LEFT JOIN %s ON %s", table, condition))
	return qb
}

// Order agrega ordenamiento
func (qb *QueryBuilder) Order(order string) *QueryBuilder {
	qb.query = qb.query.Order(order)
	return qb
}

// Limit agrega límite
func (qb *QueryBuilder) Limit(limit int) *QueryBuilder {
	qb.query = qb.query.Limit(limit)
	return qb
}

// Offset agrega offset
func (qb *QueryBuilder) Offset(offset int) *QueryBuilder {
	qb.query = qb.query.Offset(offset)
	return qb
}

// Group agrega GROUP BY
func (qb *QueryBuilder) Group(group string) *QueryBuilder {
	qb.query = qb.query.Group(group)
	return qb
}

// Having agrega HAVING
func (qb *QueryBuilder) Having(condition string, args ...interface{}) *QueryBuilder {
	qb.query = qb.query.Having(condition, args...)
	return qb
}

// Build retorna la consulta construida
func (qb *QueryBuilder) Build() *gorm.DB {
	return qb.query
}

// Execute ejecuta la consulta y retorna resultados
func (qb *QueryBuilder) Execute(dest interface{}) error {
	return qb.query.Find(dest).Error
}

// Count ejecuta un COUNT
func (qb *QueryBuilder) Count() (int64, error) {
	var count int64
	err := qb.query.Count(&count).Error
	return count, err
}

// First obtiene el primer resultado
func (qb *QueryBuilder) First(dest interface{}) error {
	return qb.query.First(dest).Error
}

// DatabaseMetrics contiene métricas de la base de datos
type DatabaseMetrics struct {
	OpenConnections     int           `json:"open_connections"`
	InUseConnections    int           `json:"in_use_connections"`
	IdleConnections     int           `json:"idle_connections"`
	WaitCount           int64         `json:"wait_count"`
	WaitDuration        time.Duration `json:"wait_duration"`
	MaxIdleClosed       int64         `json:"max_idle_closed"`
	MaxIdleTimeClosed   int64         `json:"max_idle_time_closed"`
	MaxLifetimeClosed   int64         `json:"max_lifetime_closed"`
}

// GetMetrics obtiene métricas de la base de datos
func (db *Database) GetMetrics() DatabaseMetrics {
	stats := db.GetStats()
	
	return DatabaseMetrics{
		OpenConnections:     stats.OpenConnections,
		InUseConnections:    stats.InUse,
		IdleConnections:     stats.Idle,
		WaitCount:           stats.WaitCount,
		WaitDuration:        stats.WaitDuration,
		MaxIdleClosed:       stats.MaxIdleClosed,
		MaxIdleTimeClosed:   stats.MaxIdleTimeClosed,
		MaxLifetimeClosed:   stats.MaxLifetimeClosed,
	}
}

