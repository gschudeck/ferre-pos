package mocks

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"time"

	"github.com/stretchr/testify/mock"
	"ferre_pos_apis/internal/models"
)

// MockDatabase mock de la base de datos
type MockDatabase struct {
	mock.Mock
}

// NewMockDatabase crea un nuevo mock de base de datos
func NewMockDatabase() *MockDatabase {
	return &MockDatabase{}
}

// Query mock del método Query
func (m *MockDatabase) Query(query string, args ...interface{}) (*sql.Rows, error) {
	arguments := m.Called(query, args)
	return arguments.Get(0).(*sql.Rows), arguments.Error(1)
}

// QueryContext mock del método QueryContext
func (m *MockDatabase) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	arguments := m.Called(ctx, query, args)
	return arguments.Get(0).(*sql.Rows), arguments.Error(1)
}

// QueryRow mock del método QueryRow
func (m *MockDatabase) QueryRow(query string, args ...interface{}) *sql.Row {
	arguments := m.Called(query, args)
	return arguments.Get(0).(*sql.Row)
}

// QueryRowContext mock del método QueryRowContext
func (m *MockDatabase) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	arguments := m.Called(ctx, query, args)
	return arguments.Get(0).(*sql.Row)
}

// Exec mock del método Exec
func (m *MockDatabase) Exec(query string, args ...interface{}) (sql.Result, error) {
	arguments := m.Called(query, args)
	return arguments.Get(0).(sql.Result), arguments.Error(1)
}

// ExecContext mock del método ExecContext
func (m *MockDatabase) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	arguments := m.Called(ctx, query, args)
	return arguments.Get(0).(sql.Result), arguments.Error(1)
}

// BeginTx mock del método BeginTx
func (m *MockDatabase) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	arguments := m.Called(ctx, opts)
	return arguments.Get(0).(*sql.Tx), arguments.Error(1)
}

// Ping mock del método Ping
func (m *MockDatabase) Ping() error {
	arguments := m.Called()
	return arguments.Error(0)
}

// Close mock del método Close
func (m *MockDatabase) Close() error {
	arguments := m.Called()
	return arguments.Error(0)
}

// GetStats mock del método GetStats
func (m *MockDatabase) GetStats() sql.DBStats {
	arguments := m.Called()
	return arguments.Get(0).(sql.DBStats)
}

// MockResult mock de sql.Result
type MockResult struct {
	mock.Mock
}

func (m *MockResult) LastInsertId() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockResult) RowsAffected() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

// MockRows mock de sql.Rows
type MockRows struct {
	mock.Mock
	columns []string
	data    [][]driver.Value
	index   int
}

func NewMockRows(columns []string, data [][]driver.Value) *MockRows {
	return &MockRows{
		columns: columns,
		data:    data,
		index:   -1,
	}
}

func (m *MockRows) Columns() ([]string, error) {
	return m.columns, nil
}

func (m *MockRows) Close() error {
	return nil
}

func (m *MockRows) Next() bool {
	m.index++
	return m.index < len(m.data)
}

func (m *MockRows) Scan(dest ...interface{}) error {
	if m.index >= len(m.data) || m.index < 0 {
		return errors.New("no rows")
	}
	
	row := m.data[m.index]
	for i, value := range row {
		if i < len(dest) {
			switch d := dest[i].(type) {
			case *string:
				if value != nil {
					*d = value.(string)
				}
			case *int:
				if value != nil {
					*d = value.(int)
				}
			case *int64:
				if value != nil {
					*d = value.(int64)
				}
			case *float64:
				if value != nil {
					*d = value.(float64)
				}
			case *bool:
				if value != nil {
					*d = value.(bool)
				}
			case *time.Time:
				if value != nil {
					*d = value.(time.Time)
				}
			}
		}
	}
	return nil
}

// MockUserRepository mock del repositorio de usuarios
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*models.Usuario, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.Usuario), args.Error(1)
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*models.Usuario, error) {
	args := m.Called(ctx, username)
	return args.Get(0).(*models.Usuario), args.Error(1)
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.Usuario) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Update(ctx context.Context, user *models.Usuario) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) List(ctx context.Context, limit, offset int) ([]*models.Usuario, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*models.Usuario), args.Error(1)
}

// MockProductRepository mock del repositorio de productos
type MockProductRepository struct {
	mock.Mock
}

func (m *MockProductRepository) GetByID(ctx context.Context, id string) (*models.Producto, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.Producto), args.Error(1)
}

func (m *MockProductRepository) GetByCodigo(ctx context.Context, codigo string) (*models.Producto, error) {
	args := m.Called(ctx, codigo)
	return args.Get(0).(*models.Producto), args.Error(1)
}

func (m *MockProductRepository) GetByCodigoBarras(ctx context.Context, codigoBarras string) (*models.Producto, error) {
	args := m.Called(ctx, codigoBarras)
	return args.Get(0).(*models.Producto), args.Error(1)
}

func (m *MockProductRepository) Create(ctx context.Context, product *models.Producto) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockProductRepository) Update(ctx context.Context, product *models.Producto) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockProductRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockProductRepository) List(ctx context.Context, limit, offset int) ([]*models.Producto, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*models.Producto), args.Error(1)
}

func (m *MockProductRepository) Search(ctx context.Context, query string, limit, offset int) ([]*models.Producto, error) {
	args := m.Called(ctx, query, limit, offset)
	return args.Get(0).([]*models.Producto), args.Error(1)
}

func (m *MockProductRepository) UpdateStock(ctx context.Context, productID string, newStock int) error {
	args := m.Called(ctx, productID, newStock)
	return args.Error(0)
}

// MockVentaRepository mock del repositorio de ventas
type MockVentaRepository struct {
	mock.Mock
}

func (m *MockVentaRepository) GetByID(ctx context.Context, id string) (*models.Venta, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.Venta), args.Error(1)
}

func (m *MockVentaRepository) Create(ctx context.Context, venta *models.Venta) error {
	args := m.Called(ctx, venta)
	return args.Error(0)
}

func (m *MockVentaRepository) Update(ctx context.Context, venta *models.Venta) error {
	args := m.Called(ctx, venta)
	return args.Error(0)
}

func (m *MockVentaRepository) List(ctx context.Context, limit, offset int) ([]*models.Venta, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*models.Venta), args.Error(1)
}

func (m *MockVentaRepository) GetByDateRange(ctx context.Context, start, end time.Time) ([]*models.Venta, error) {
	args := m.Called(ctx, start, end)
	return args.Get(0).([]*models.Venta), args.Error(1)
}

func (m *MockVentaRepository) GetByUser(ctx context.Context, userID string, limit, offset int) ([]*models.Venta, error) {
	args := m.Called(ctx, userID, limit, offset)
	return args.Get(0).([]*models.Venta), args.Error(1)
}

// MockLogger mock del logger
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Debug(args ...interface{}) {
	m.Called(args...)
}

func (m *MockLogger) Info(args ...interface{}) {
	m.Called(args...)
}

func (m *MockLogger) Warn(args ...interface{}) {
	m.Called(args...)
}

func (m *MockLogger) Error(args ...interface{}) {
	m.Called(args...)
}

func (m *MockLogger) Fatal(args ...interface{}) {
	m.Called(args...)
}

func (m *MockLogger) WithField(key string, value interface{}) *MockLogger {
	return m
}

func (m *MockLogger) WithFields(fields map[string]interface{}) *MockLogger {
	return m
}

func (m *MockLogger) WithError(err error) *MockLogger {
	return m
}

// MockMetrics mock de métricas
type MockMetrics struct {
	mock.Mock
}

func (m *MockMetrics) RecordDatabaseQuery(apiName, operation, table string, duration time.Duration, err error) {
	m.Called(apiName, operation, table, duration, err)
}

func (m *MockMetrics) RecordProductoConsulta(apiName, sucursalID, operationType string) {
	m.Called(apiName, sucursalID, operationType)
}

func (m *MockMetrics) RecordVentaProcesada(apiName, sucursalID, tipoDocumento string, monto float64) {
	m.Called(apiName, sucursalID, tipoDocumento, monto)
}

func (m *MockMetrics) RecordStockMovimiento(apiName, sucursalID, tipoMovimiento string, cantidad int) {
	m.Called(apiName, sucursalID, tipoMovimiento, cantidad)
}

func (m *MockMetrics) RecordSyncOperation(apiName, operationType string, recordsProcessed int, duration time.Duration, success bool) {
	m.Called(apiName, operationType, recordsProcessed, duration, success)
}

func (m *MockMetrics) RecordLabelGenerated(apiName, templateType, format string) {
	m.Called(apiName, templateType, format)
}

func (m *MockMetrics) RecordReportGenerated(apiName, reportType, format string, duration time.Duration) {
	m.Called(apiName, reportType, format, duration)
}

// MockValidator mock del validador
type MockValidator struct {
	mock.Mock
}

func (m *MockValidator) ValidateStruct(s interface{}) error {
	args := m.Called(s)
	return args.Error(0)
}

func (m *MockValidator) ValidateEmail(email string) bool {
	args := m.Called(email)
	return args.Bool(0)
}

func (m *MockValidator) ValidatePassword(password string) error {
	args := m.Called(password)
	return args.Error(0)
}

func (m *MockValidator) ValidateProductCode(code string) bool {
	args := m.Called(code)
	return args.Bool(0)
}

func (m *MockValidator) ValidateBarcode(barcode string) bool {
	args := m.Called(barcode)
	return args.Bool(0)
}

func (m *MockValidator) SanitizeInput(input string) string {
	args := m.Called(input)
	return args.String(0)
}

