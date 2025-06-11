package users

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	sc "github.com/Klojer/sqlcredo"
	"github.com/Klojer/sqlcredo/pkg/api"
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

func NewRepo(db *sql.DB, driver string, debugFunc api.DebugFunc) *Repo {
	return &Repo{
		SQLCredo: sc.NewSQLCredo[Object, Identity](db, driver, TableName, IDColumn).
			WithDebugFunc(debugFunc),
	}
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
