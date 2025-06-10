package model

import (
	"context"
	"database/sql"
)

// CRUD defines a generic interface for basic database operations.
// Type parameters:
//   - T: the entity type being managed
//   - I: the type of the entity's ID field (must be comparable)
type CRUD[T any, I comparable] interface {
	// GetAll retrieves all entities of type T from the database.
	GetAll(ctx context.Context) ([]T, error)

	// GetByID retrieves a single entity by its ID.
	// Returns the zero value of T and an error if the entity is not found.
	GetByID(ctx context.Context, id I) (T, error)

	// GetByIDs retrieves multiple entities by their IDs.
	// The returned slice maintains the same order as the input IDs.
	GetByIDs(ctx context.Context, ids []I) ([]T, error)

	// Create inserts a new entity into the database.
	// The entity pointer must not be nil.
	Create(ctx context.Context, e *T) (sql.Result, error)

	// DeleteAll removes all entities of type T from the database.
	DeleteAll(ctx context.Context) (sql.Result, error)

	// Delete removes a single entity by its ID.
	Delete(ctx context.Context, id I) (sql.Result, error)

	// Update modifies an existing entity identified by its ID.
	// The entity pointer must not be nil.
	Update(ctx context.Context, id I, e *T) (sql.Result, error)
}
