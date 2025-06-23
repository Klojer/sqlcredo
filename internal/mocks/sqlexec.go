package mocks

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Klojer/sqlcredo/pkg/api"
)

type SQLExecutor struct {
	Queries []string
	Args    [][]any
}

var _ api.SQLExecutor = &SQLExecutor{}

func NewSQLExecutor() *SQLExecutor {
	return &SQLExecutor{
		Queries: make([]string, 0),
		Args:    make([][]any, 0),
	}
}

func (m *SQLExecutor) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	m.Queries = append(m.Queries, query)
	m.Args = append(m.Args, args)
	return nil, nil
}

func (m *SQLExecutor) SelectOne(ctx context.Context, dest any, query string, args ...any) error {
	m.Queries = append(m.Queries, query)
	m.Args = append(m.Args, args)
	return nil
}

func (m *SQLExecutor) SelectMany(ctx context.Context, dest any, query string, args ...any) error {
	m.Queries = append(m.Queries, query)
	m.Args = append(m.Args, args)
	return nil
}

func (m *SQLExecutor) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return nil, errors.New("not implemented")
}
