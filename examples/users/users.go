// Package users demonstrates how to extend sqlcredo with custom SQL queries.
// It provides a complete example of a user repository that includes standard CRUD operations
// and a custom query method for counting users based on whether they have a last name.
package users

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	sc "github.com/Klojer/sqlcredo"
)

const (
	TableName = "users"
	IDColumn  = "id"
)

type Identity string

type Object struct {
	ID        Identity  `db:"id"`
	FirstName string    `db:"first_name"`
	LastName  *string   `db:"last_name"`
	BirthDate time.Time `db:"birth_date"`
}

func (u *Object) String() string {
	return fmt.Sprintf("%v", *u)
}

type Repo struct {
	sc.SQLCredo[Object, Identity]
}

func NewRepo(db *sql.DB, driver string, debugFunc sc.DebugFunc) *Repo {
	return &Repo{
		SQLCredo: sc.NewSQLCredo[Object, Identity](db, driver, TableName, IDColumn).
			WithDebugFunc(debugFunc),
	}
}

func (r *Repo) WithTxx(txExec sc.SQLExecutor) *Repo {
	return &Repo{SQLCredo: r.SQLCredo.WithTx(txExec)}
}

const CountByLastNameExistsQuery = `
SELECT 'with last_name' as category, COUNT(*) as cnt FROM users WHERE last_name IS NOT NULL
UNION
SELECT 'without last_name' as category, COUNT(*) as cnt FROM users WHERE last_name IS NULL;
`

type CountByLastNameExistsCategory struct {
	Name  string `db:"category"`
	Count int    `db:"cnt"`
}

func (r *Repo) CountByLastNameExists(ctx context.Context) (map[string]int, error) {
	var counters []CountByLastNameExistsCategory
	if err := r.SelectMany(ctx, &counters, CountByLastNameExistsQuery); err != nil {
		return nil, fmt.Errorf("unable to select records: %w", err)
	}

	res := map[string]int{}
	for _, c := range counters {
		res[c.Name] = c.Count
	}

	return res, nil
}
