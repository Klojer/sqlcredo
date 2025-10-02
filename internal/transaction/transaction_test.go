package transaction_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Klojer/sqlcredo/internal/table"
	"github.com/Klojer/sqlcredo/internal/transaction"
	"github.com/Klojer/sqlcredo/pkg/api"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestTransaction_BeginError(t *testing.T) {
	c, ctx := newTestCase(t, func(mock sqlmock.Sqlmock) {
		mock.ExpectBegin().WillReturnError(errors.New("tx issue"))
	})

	_, err := transaction.NewTx[testObj, string](ctx, c.DBX, c.TableInfo, c.Driver, c.DebugFunc, nil)
	assert.ErrorContains(t, err, "unable to begin transaction:")
}

func TestTransaction_Commit(t *testing.T) {
	c, ctx := newTestCase(t, func(mock sqlmock.Sqlmock) {
		query := "INSERT INTO `test_table` \\(`id`, `name`\\) VALUES \\(\\?, \\?\\)"
		mock.ExpectBegin()
		mock.ExpectExec(query).
			WithArgs("#1", "u1").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
	})

	underTest, err := transaction.NewTx[testObj, string](ctx, c.DBX, c.TableInfo, c.Driver, c.DebugFunc, nil)
	assert.NoError(t, err)

	_, err = underTest.Create(ctx, &testObj{
		ID: "#1", Name: "u1",
	})
	assert.NoError(t, err)

	err = underTest.Commit()
	assert.NoError(t, err)
}

func TestTransaction_Rollback(t *testing.T) {
	c, ctx := newTestCase(t, func(mock sqlmock.Sqlmock) {
		query := "INSERT INTO `test_table` \\(`id`, `name`\\) VALUES \\(\\?, \\?\\)"
		mock.ExpectBegin()
		mock.ExpectExec(query).
			WithArgs("#1", "u1").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectRollback()
	})

	underTest, err := transaction.NewTx[testObj, string](ctx, c.DBX, c.TableInfo, c.Driver, c.DebugFunc, nil)
	assert.NoError(t, err)

	_, err = underTest.Create(ctx, &testObj{
		ID: "#1", Name: "u1",
	})
	assert.NoError(t, err)

	err = underTest.Rollback()
	assert.NoError(t, err)
}

type testCaseData struct {
	ctx       context.Context
	ctxCancel func()

	Driver    string
	DBX       *sqlx.DB
	TableInfo table.Info
	DebugFunc api.DebugFunc
}

func newTestCase(t *testing.T, mockCfg func(sqlmock.Sqlmock)) (*testCaseData, context.Context) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	driver := "sqlite3"

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	mockCfg(mock)

	dbx := sqlx.NewDb(db, driver)

	tableInfo := table.Info{Name: "test_table", IDColumn: "id"}

	debugFunc := func(sql string, args ...any) {
		t.Helper()
		t.Logf("Query: '%s'; args: '%v'", sql, args)
	}

	c := &testCaseData{
		ctx:       ctx,
		ctxCancel: cancel,

		Driver:    driver,
		DBX:       dbx,
		TableInfo: tableInfo,
		DebugFunc: debugFunc,
	}

	t.Cleanup(func() {
		t.Helper()

		_ = db.Close()
		cancel()
	})

	return c, ctx
}

type testObj struct {
	ID   string `db:"id"`
	Name string `db:"name"`
}
