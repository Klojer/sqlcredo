package sqlcredo_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/Klojer/sqlcredo"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/mattn/go-sqlite3"
)

type TestEntity struct {
	ID   string `db:"id"`
	Name string `db:"name"`
}

func TestNewSQLCredo(t *testing.T) {
	_, _ = newTestCase(t)
}

func TestSQLCredo_InitSchema(t *testing.T) {
	tests := []struct {
		name    string
		schema  string
		wantErr bool
	}{
		{
			name: "Valid schema creation",
			schema: `CREATE TABLE test_table (
				id TEXT PRIMARY KEY,
				name TEXT NOT NULL
			)`,
			wantErr: false,
		},
		{
			name:    "Invalid SQL syntax",
			schema:  "INVALID SQL QUERY",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, ctx := newTestCase(t)

			result, err := c.UnderTest.InitSchema(ctx, tt.schema)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestSQLCredo_DebugFunc(t *testing.T) {
	c, ctx := newTestCase(t)

	debugCalled := false
	debugFunc := func(query string, args ...any) {
		debugCalled = true
	}
	assert.NotEqual(t, fmt.Sprintf("%p", debugFunc),
		fmt.Sprintf("%p", c.UnderTest.GetDebugFunc()))

	credoWithDebug := c.UnderTest.WithDebugFunc(debugFunc)
	assert.Equal(t, fmt.Sprintf("%p", debugFunc),
		fmt.Sprintf("%p", c.UnderTest.GetDebugFunc()))

	schema := `CREATE TABLE test_table (id TEXT PRIMARY KEY, name TEXT NOT NULL)`
	_, err := credoWithDebug.InitSchema(ctx, schema)
	assert.NoError(t, err)
	assert.True(t, debugCalled)
}

func TestSQLCredo_BeginTx(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() (context.Context, func())
		wantErr bool
	}{
		{
			name: "successful transaction",
			setup: func() (context.Context, func()) {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				return ctx, cancel
			},
			wantErr: false,
		},
		{
			name: "cancelled context",
			setup: func() (context.Context, func()) {
				ctx, cancel := context.WithCancel(context.Background())
				cancel() // cancel immediately
				return ctx, func() {}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := newTestCase(t)
			ctx, cancel := tt.setup()
			defer cancel()

			tx, err := c.UnderTest.BeginTx(ctx, nil)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, tx)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, tx)
				// Clean up transaction
				err = tx.Rollback()
				assert.NoError(t, err)
			}
		})
	}
}

func TestSQLCredo_WithTx(t *testing.T) {
	c, _ := newTestCase(t)

	// Create a mock executor
	mockExec := &mockSQLExecutor{}

	// Call WithTx
	credoWithTx := c.UnderTest.WithTx(mockExec)

	// Verify it's a different instance
	assert.NotEqual(t, fmt.Sprintf("%p", c.UnderTest), fmt.Sprintf("%p", credoWithTx))

	// Verify driver is preserved
	assert.Equal(t, c.UnderTest.GetDriver(), credoWithTx.GetDriver())
}

// mockSQLExecutor implements SQLExecutor for testing
type mockSQLExecutor struct{}

func (m *mockSQLExecutor) SelectOne(ctx context.Context, dest any, query string, args ...any) error {
	return nil
}

func (m *mockSQLExecutor) SelectMany(ctx context.Context, dest any, query string, args ...any) error {
	return nil
}

func (m *mockSQLExecutor) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return nil, nil
}

func TestSQLCredo_GetDriver(t *testing.T) {
	c, _ := newTestCase(t)

	assert.Equal(t, "sqlite3", c.UnderTest.GetDriver())
}

type testCaseData struct {
	ctx       context.Context
	ctxCancel func()

	db        *sql.DB
	UnderTest sqlcredo.SQLCredo[TestEntity, string]
}

func newTestCase(t *testing.T) (*testCaseData, context.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	require.NotNil(t, db)

	c := &testCaseData{
		ctx:       ctx,
		ctxCancel: cancel,

		db:        db,
		UnderTest: sqlcredo.NewSQLCredo[TestEntity, string](db, "sqlite3", "test_table", "id"),
	}

	t.Cleanup(func() {
		c.TearDown(t)
	})

	return c, ctx
}

func (c *testCaseData) TearDown(t *testing.T) {
	c.ctxCancel()
	assert.NoError(t, c.db.Close())
}
