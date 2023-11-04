package sqlcredo_test

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"gitlab.com/onrooh/sqlcredo"

	_ "github.com/mattn/go-sqlite3"

	"github.com/stretchr/testify/assert"
)

type Identity string

type User struct {
	ID        Identity  `db:"id"`
	FirstName string    `db:"first_name"`
	LastName  *string   `db:"last_name"`
	BirthDate time.Time `db:"birth_date"`
}

func (u *User) String() string {
	return fmt.Sprintf("%v", *u)
}

var (
	dsn       = ":memory:"
	driver    = "sqlite3"
	tableName = "user"
	idColumn  = "id"

	schema = `
CREATE TABLE IF NOT EXISTS user (
    id TEXT NOT NULL PRIMARY KEY,
    first_name TEXT NOT NULL,
    last_name TEXT NULL,
    birth_date DATETIME NOT NULL
);
`

	testUsers = []*User{
		{"u0", "John", ptr("Smith"), newTime("1989-03-05")},
		{"u1", "Carl", nil, newTime("1973-01-09")},
		{"u2", "Ann", ptr("Stone"), newTime("1985-08-01")},
	}

	testUserValues = []User{
		{"u0", "John", ptr("Smith"), newTime("1989-03-05")},
		{"u1", "Carl", nil, newTime("1973-01-09")},
		{"u2", "Ann", ptr("Stone"), newTime("1985-08-01")},
	}
)

type UserRepo struct {
	sqlcredo.SQLCredo[User, Identity]
}

const CountByLastNameExistsQuery = `
SELECT 'with last_name' as category, COUNT(*) as cnt FROM user WHERE last_name IS NOT NULL
UNION
SELECT 'without last_name' as category, COUNT(*) as cnt FROM user WHERE last_name IS NULL;
`

type CountByLastNameExistsCategory struct {
	Name  string `db:"category"`
	Count int    `db:"cnt"`
}

func (r *UserRepo) CountByLastNameExists() (map[string]int, error) {
	var counters []CountByLastNameExistsCategory
	if err := r.SelectMany(&counters, CountByLastNameExistsQuery); err != nil {
		return nil, fmt.Errorf("failed to select entities: %w", err)
	}

	res := map[string]int{}
	for _, c := range counters {
		res[c.Name] = c.Count
	}

	return res, nil
}

func TestInitSchema(t *testing.T) {
	_, teardown := setup(t)
	defer teardown()
}

func TestCreate(t *testing.T) {
	r, teardown := setup(t)
	defer teardown()

	u := testUsers[0]

	err := r.Create(u)
	assert.NoError(t, err)

	got, err := r.GetByID(u.ID)
	assert.NoError(t, err)
	assert.Equal(t, *u, *got)
}

func TestGetAll(t *testing.T) {
	r, teardown := setup(t, testUsers...)
	defer teardown()

	got, err := r.GetAll()

	assert.NoError(t, err)
	assert.Equal(t, testUsers, got)
}

func TestGetAllValues(t *testing.T) {
	r, teardown := setup(t, testUsers...)
	defer teardown()

	got, err := r.GetAllValues()

	assert.NoError(t, err)
	assert.Equal(t, testUserValues, got)
}

func TestGetPage(t *testing.T) {
	r, teardown := setup(t, testUsers...)
	defer teardown()

	gotPage1, err := r.GetAll(
		sqlcredo.WithLimit(2),
		sqlcredo.WithOffset(0),
		sqlcredo.WithOrderColumn("id"),
	)
	assert.NoError(t, err)
	assert.Equal(t, testUsers[0:2], gotPage1)

	gotPage2, err := r.GetAll(
		sqlcredo.WithLimit(2),
		sqlcredo.WithOffset(2),
		sqlcredo.WithOrderColumn("id"),
	)
	assert.NoError(t, err)
	assert.Equal(t, testUsers[2:], gotPage2)
}

func TestGetByID(t *testing.T) {
	r, teardown := setup(t, testUsers...)
	defer teardown()

	u := testUsers[1]

	got, err := r.GetByID(u.ID)

	assert.NoError(t, err)
	assert.Equal(t, *u, *got)
}

func TestGetValueByID(t *testing.T) {
	r, teardown := setup(t, testUsers...)
	defer teardown()

	u := testUserValues[1]

	got, err := r.GetValueByID(u.ID)

	assert.NoError(t, err)
	assert.Equal(t, u, got)
}

func TestGetByIDs(t *testing.T) {
	r, teardown := setup(t, testUsers...)
	defer teardown()

	u1 := testUsers[1]
	u2 := testUsers[2]

	got, err := r.GetByIDs([]Identity{u1.ID, u2.ID})

	assert.NoError(t, err)
	assert.Equal(t, 2, len(got))
	assert.Equal(t, *u1, *got[0])
	assert.Equal(t, *u2, *got[1])
}

func TestGetValuesByIDs(t *testing.T) {
	r, teardown := setup(t, testUsers...)
	defer teardown()

	u1 := testUserValues[1]
	u2 := testUserValues[2]

	got, err := r.GetValuesByIDs([]Identity{u1.ID, u2.ID})

	assert.NoError(t, err)
	assert.Equal(t, 2, len(got))
	assert.Equal(t, u1, got[0])
	assert.Equal(t, u2, got[1])
}

func TestDelete(t *testing.T) {
	r, teardown := setup(t, testUsers...)
	defer teardown()

	u := testUsers[1]

	err := r.Delete(u.ID)
	assert.NoError(t, err)

	got, err := r.GetAll()
	assert.NoError(t, err)
	assert.Equal(t, []*User{testUsers[0], testUsers[2]}, got)
}

func TestUpdate(t *testing.T) {
	r, teardown := setup(t, testUsers...)
	defer teardown()

	updated := *testUsers[0]
	updated.FirstName = "Bob"

	err := r.Update(updated.ID, &updated)
	assert.NoError(t, err)

	got, err := r.GetByID(testUsers[0].ID)
	assert.NoError(t, err)
	assert.Equal(t, updated, *got)
}

func TestCount(t *testing.T) {
	r, teardown := setup(t, testUsers...)
	defer teardown()

	got, err := r.Count()
	assert.NoError(t, err)
	assert.Equal(t, int64(len(testUsers)), got)
}

func TestCountByLastNameExists(t *testing.T) {
	r, teardown := setup(t, testUsers...)
	defer teardown()

	got, err := r.CountByLastNameExists()
	assert.NoError(t, err)
	assert.Equal(t, map[string]int{
		"with last_name":    2,
		"without last_name": 1,
	}, got)
}

func setup(t *testing.T, initData ...*User) (UserRepo, func()) {
	db, err := sql.Open(driver, dsn)
	assert.NoError(t, err)

	r := UserRepo{
		SQLCredo: sqlcredo.NewSQLCredo[User, Identity](db, driver, tableName, idColumn).
			WithDebugFunc(func(query string, args ...any) {
				fmt.Printf("Query: [%s]; Args: %+v\n", query, args)
			}),
	}

	err = r.InitSchema(schema)
	assert.NoError(t, err)

	for _, u := range initData {
		err := r.Create(u)
		assert.NoError(t, err)
	}

	teardown := func() {
		db.Close()
	}

	return r, teardown
}

func newTime(input string) time.Time {
	result, err := time.Parse("2006-01-02", input)
	if err != nil {
		panic(err)
	}

	return result
}

func ptr[T comparable](input T) *T {
	return &input
}
