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

type Article struct {
	ID        Identity  `db:"id"`
	Title     string    `db:"title"`
	CreatedAt time.Time `db:"created_at"`
	UserID    Identity  `db:"user_id"`
}

type Object struct {
	ID        Identity  `db:"id"`
	FirstName string    `db:"first_name"`
	LastName  *string   `db:"last_name"`
	BirthDate time.Time `db:"birth_date"`
	Articles  []Article `db:"-"`
}

func (u *Object) String() string {
	return fmt.Sprintf("%v", *u)
}

type ArticlesRepo struct {
	sc.SQLCredo[Article, Identity]
}

func NewArticlesRepo(db *sql.DB, driver string, debugFunc sc.DebugFunc) *ArticlesRepo {
	return &ArticlesRepo{
		SQLCredo: sc.NewSQLCredo[Article, Identity](db, driver, "articles", "id").
			WithDebugFunc(debugFunc),
	}
}

func (r *ArticlesRepo) WithTxx(txExec sc.SQLExecutor) *ArticlesRepo {
	return &ArticlesRepo{SQLCredo: r.SQLCredo.WithTx(txExec)}
}

type Repo struct {
	sc.SQLCredo[Object, Identity]
	articlesRepo *ArticlesRepo
}

func NewRepo(db *sql.DB, driver string, debugFunc sc.DebugFunc) *Repo {
	return &Repo{
		SQLCredo: sc.NewSQLCredo[Object, Identity](db, driver, TableName, IDColumn).
			WithDebugFunc(debugFunc),
		articlesRepo: NewArticlesRepo(db, driver, debugFunc),
	}
}

func (r *Repo) WithTxx(txExec sc.SQLExecutor) *Repo {
	return &Repo{
		SQLCredo:     r.SQLCredo.WithTx(txExec),
		articlesRepo: r.articlesRepo.WithTxx(txExec),
	}
}

func (r *Repo) Articles() *ArticlesRepo {
	return r.articlesRepo
}

func (r *Repo) GetWithArticles(ctx context.Context, dest *Object, id Identity) error {
	if err := r.GetByID(ctx, dest, id); err != nil {
		return fmt.Errorf("unable to get user: %w", err)
	}

	var articles []Article
	query := "SELECT * FROM articles WHERE user_id = ?"
	if err := r.articlesRepo.SelectMany(ctx, &articles, query, id); err != nil {
		return fmt.Errorf("unable to get articles: %w", err)
	}
	dest.Articles = articles

	return nil
}

func (r *Repo) DeleteWithArticles(ctx context.Context, id Identity) error {
	tx, err := r.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("unable to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	txRepo := r.WithTxx(tx)

	query := "DELETE FROM articles WHERE user_id = ?"
	if _, err := txRepo.Articles().Exec(ctx, query, id); err != nil {
		return fmt.Errorf("unable to delete articles: %w", err)
	}

	if _, err := txRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("unable to delete user: %w", err)
	}

	return tx.Commit()
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
