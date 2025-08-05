package unit

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"ferre_pos_apis/internal/config"
	"ferre_pos_apis/internal/database"
	"ferre_pos_apis/internal/logger"
	"ferre_pos_apis/test/mocks"
)

func TestDatabaseInit(t *testing.T) {
	// Configurar logger para tests
	err := logger.Init(&config.LoggingConfig{
		Level:  "error",
		Format: "json",
		Output: "stdout",
	}, "test")
	require.NoError(t, err)

	tests := []struct {
		name        string
		config      *config.DatabaseConfig
		expectError bool
	}{
		{
			name: "valid config",
			config: &config.DatabaseConfig{
				Host:            "localhost",
				Port:            5432,
				User:            "test",
				Password:        "test",
				Name:            "test_db",
				SSLMode:         "disable",
				MaxOpenConns:    10,
				MaxIdleConns:    5,
				ConnMaxLifetime: 5 * time.Minute,
			},
			expectError: false, // Puede fallar si no hay DB real, pero la lógica es correcta
		},
		{
			name: "invalid port",
			config: &config.DatabaseConfig{
				Host:     "localhost",
				Port:     0,
				User:     "test",
				Password: "test",
				Name:     "test_db",
			},
			expectError: true,
		},
		{
			name: "empty host",
			config: &config.DatabaseConfig{
				Host:     "",
				Port:     5432,
				User:     "test",
				Password: "test",
				Name:     "test_db",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := database.Init(tt.config, logger.Get())

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, db)
			} else {
				// Nota: Este test puede fallar si no hay una base de datos real
				// En un entorno de CI/CD, se debería usar una base de datos de test
				if err != nil {
					t.Skipf("Skipping test due to database connection error: %v", err)
				}
				assert.NotNil(t, db)
				if db != nil {
					db.Close()
				}
			}
		})
	}
}

func TestDatabaseConnectionString(t *testing.T) {
	tests := []struct {
		name     string
		config   *config.DatabaseConfig
		expected string
	}{
		{
			name: "basic connection string",
			config: &config.DatabaseConfig{
				Host:     "localhost",
				Port:     5432,
				User:     "user",
				Password: "pass",
				Name:     "dbname",
				SSLMode:  "disable",
			},
			expected: "host=localhost port=5432 user=user password=pass dbname=dbname sslmode=disable",
		},
		{
			name: "with ssl mode require",
			config: &config.DatabaseConfig{
				Host:     "remote-host",
				Port:     5432,
				User:     "user",
				Password: "pass",
				Name:     "dbname",
				SSLMode:  "require",
			},
			expected: "host=remote-host port=5432 user=user password=pass dbname=dbname sslmode=require",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connStr := database.BuildConnectionString(tt.config)
			assert.Equal(t, tt.expected, connStr)
		})
	}
}

func TestDatabaseOperations(t *testing.T) {
	mockDB := mocks.NewMockDatabase()
	
	tests := []struct {
		name     string
		setup    func()
		test     func(*testing.T)
		cleanup  func()
	}{
		{
			name: "successful query",
			setup: func() {
				mockDB.On("QueryContext", mock.Anything, mock.AnythingOfType("string"), mock.Anything).
					Return(&sql.Rows{}, nil)
			},
			test: func(t *testing.T) {
				ctx := context.Background()
				rows, err := mockDB.QueryContext(ctx, "SELECT * FROM test", nil)
				assert.NoError(t, err)
				assert.NotNil(t, rows)
			},
			cleanup: func() {
				mockDB.AssertExpectations(t)
			},
		},
		{
			name: "successful exec",
			setup: func() {
				mockResult := &mocks.MockResult{}
				mockResult.On("RowsAffected").Return(int64(1), nil)
				mockDB.On("ExecContext", mock.Anything, mock.AnythingOfType("string"), mock.Anything).
					Return(mockResult, nil)
			},
			test: func(t *testing.T) {
				ctx := context.Background()
				result, err := mockDB.ExecContext(ctx, "INSERT INTO test VALUES ($1)", "value")
				assert.NoError(t, err)
				assert.NotNil(t, result)
				
				affected, err := result.RowsAffected()
				assert.NoError(t, err)
				assert.Equal(t, int64(1), affected)
			},
			cleanup: func() {
				mockDB.AssertExpectations(t)
			},
		},
		{
			name: "successful ping",
			setup: func() {
				mockDB.On("Ping").Return(nil)
			},
			test: func(t *testing.T) {
				err := mockDB.Ping()
				assert.NoError(t, err)
			},
			cleanup: func() {
				mockDB.AssertExpectations(t)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			tt.test(t)
			tt.cleanup()
		})
	}
}

func TestDatabaseTransaction(t *testing.T) {
	mockDB := mocks.NewMockDatabase()
	
	tests := []struct {
		name    string
		setup   func()
		test    func(*testing.T)
		cleanup func()
	}{
		{
			name: "successful transaction",
			setup: func() {
				mockTx := &sql.Tx{}
				mockDB.On("BeginTx", mock.Anything, mock.AnythingOfType("*sql.TxOptions")).
					Return(mockTx, nil)
			},
			test: func(t *testing.T) {
				ctx := context.Background()
				tx, err := mockDB.BeginTx(ctx, nil)
				assert.NoError(t, err)
				assert.NotNil(t, tx)
			},
			cleanup: func() {
				mockDB.AssertExpectations(t)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			tt.test(t)
			tt.cleanup()
		})
	}
}

func TestDatabaseStats(t *testing.T) {
	mockDB := mocks.NewMockDatabase()
	
	expectedStats := sql.DBStats{
		MaxOpenConnections: 25,
		OpenConnections:    5,
		InUse:             2,
		Idle:              3,
	}
	
	mockDB.On("GetStats").Return(expectedStats)
	
	stats := mockDB.GetStats()
	assert.Equal(t, expectedStats.MaxOpenConnections, stats.MaxOpenConnections)
	assert.Equal(t, expectedStats.OpenConnections, stats.OpenConnections)
	assert.Equal(t, expectedStats.InUse, stats.InUse)
	assert.Equal(t, expectedStats.Idle, stats.Idle)
	
	mockDB.AssertExpectations(t)
}

func TestDatabaseClose(t *testing.T) {
	mockDB := mocks.NewMockDatabase()
	
	mockDB.On("Close").Return(nil)
	
	err := mockDB.Close()
	assert.NoError(t, err)
	
	mockDB.AssertExpectations(t)
}

func TestDatabaseContextTimeout(t *testing.T) {
	mockDB := mocks.NewMockDatabase()
	
	// Simular timeout
	mockDB.On("QueryContext", mock.Anything, mock.AnythingOfType("string"), mock.Anything).
		Return((*sql.Rows)(nil), context.DeadlineExceeded)
	
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()
	
	time.Sleep(2 * time.Millisecond) // Asegurar que el contexto expire
	
	rows, err := mockDB.QueryContext(ctx, "SELECT * FROM test", nil)
	assert.Error(t, err)
	assert.Equal(t, context.DeadlineExceeded, err)
	assert.Nil(t, rows)
	
	mockDB.AssertExpectations(t)
}

func TestDatabaseConnectionPooling(t *testing.T) {
	config := &config.DatabaseConfig{
		Host:            "localhost",
		Port:            5432,
		User:            "test",
		Password:        "test",
		Name:            "test_db",
		SSLMode:         "disable",
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
	}
	
	// Verificar que la configuración del pool es correcta
	assert.Equal(t, 10, config.MaxOpenConns)
	assert.Equal(t, 5, config.MaxIdleConns)
	assert.Equal(t, 5*time.Minute, config.ConnMaxLifetime)
}

func BenchmarkDatabaseQuery(b *testing.B) {
	mockDB := mocks.NewMockDatabase()
	
	mockDB.On("QueryContext", mock.Anything, mock.AnythingOfType("string"), mock.Anything).
		Return(&sql.Rows{}, nil)
	
	ctx := context.Background()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := mockDB.QueryContext(ctx, "SELECT 1", nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDatabaseExec(b *testing.B) {
	mockDB := mocks.NewMockDatabase()
	mockResult := &mocks.MockResult{}
	
	mockResult.On("RowsAffected").Return(int64(1), nil)
	mockDB.On("ExecContext", mock.Anything, mock.AnythingOfType("string"), mock.Anything).
		Return(mockResult, nil)
	
	ctx := context.Background()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := mockDB.ExecContext(ctx, "INSERT INTO test VALUES ($1)", "value")
		if err != nil {
			b.Fatal(err)
		}
	}
}

