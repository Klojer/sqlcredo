package model

import (
	"context"
	"database/sql"
)

type CRUD[T any, I comparable] interface {
	GetAll(ctx context.Context) ([]T, error)
	GetByID(ctx context.Context, id I) (T, error)
	GetByIDs(ctx context.Context, ids []I) ([]T, error)
	Create(ctx context.Context, e *T) (sql.Result, error)
	DeleteAll(ctx context.Context) (sql.Result, error)
	Delete(ctx context.Context, id I) (sql.Result, error)
	Update(ctx context.Context, id I, e *T) (sql.Result, error)
}
