package mocks

import (
	"context"
	"database/sql"

	"github.com/Klojer/sqlcredo/pkg/api"

	"github.com/stretchr/testify/mock"
)

type (
	OnExecCb   func(ctx context.Context, query string, args ...any)
	OnSelectCb func(ctx context.Context, dest any, query string, args ...any)
)

type SQLExecutor struct {
	mock.Mock
	onExecCb       OnExecCb
	onSelectOneCb  OnSelectCb
	onSelectManyCb OnSelectCb
}

var _ api.SQLExecutor = &SQLExecutor{}

func NewSQLExecutor() *SQLExecutor {
	return &SQLExecutor{}
}

func (m *SQLExecutor) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	mockArgs := m.Called(ctx, query, args)
	if m.onExecCb != nil {
		m.onExecCb(ctx, query, args...)
	}
	return mockArgs.Get(0).(sql.Result), mockArgs.Error(1)
}

func (m *SQLExecutor) SelectOne(ctx context.Context, dest any, query string, args ...any) error {
	mockArgs := m.Called(ctx, dest, query, args)
	if m.onSelectOneCb != nil {
		m.onSelectOneCb(ctx, dest, query, args...)
	}
	return mockArgs.Error(0)
}

func (m *SQLExecutor) SelectMany(ctx context.Context, dest any, query string, args ...any) error {
	mockArgs := m.Called(ctx, dest, query, args)
	if m.onSelectManyCb != nil {
		m.onSelectManyCb(ctx, dest, query, args...)
	}
	return mockArgs.Error(0)
}

func (m *SQLExecutor) SetOnExecCb(cb OnExecCb) {
	m.onExecCb = cb
}

func (m *SQLExecutor) SetOnSelectOneCb(cb OnSelectCb) {
	m.onSelectOneCb = cb
}

func (m *SQLExecutor) SetOnSelectManyCb(cb OnSelectCb) {
	m.onSelectManyCb = cb
}

type SQLResult struct {
	lastInsertIdValue int64
	rowsAffectedValue int64
}

var _ sql.Result = &SQLResult{}

func NewSQLResult(lastInsertId int64, rowsAffected int64) *SQLResult {
	return &SQLResult{
		lastInsertIdValue: lastInsertId,
		rowsAffectedValue: rowsAffected,
	}
}

func (r *SQLResult) LastInsertId() (int64, error) {
	return r.lastInsertIdValue, nil
}

func (r *SQLResult) RowsAffected() (int64, error) {
	return r.rowsAffectedValue, nil
}
