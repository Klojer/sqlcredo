package sqlcredo_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sc "gitlab.com/onrooh/sqlcredo"
	"gitlab.com/onrooh/sqlcredo/pkg/model"
)

type TestCaseDesc struct {
	name string
	run  func(*testing.T, TestCaseParams)
}

type TestCaseParams struct {
	Schema string
	Driver string
	DB     *sql.DB
}

type testCaseData struct {
	ctx       context.Context
	ctxCancel func()

	TestUsers    []User
	TestUserPtrs []*User

	db        *sql.DB
	UnderTest UserRepo
}

func newTestCase(t *testing.T, params TestCaseParams) (*testCaseData, context.Context) {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 30*time.Second)
	testUserValues, testUserPtrs := createTestUsers()

	repo := UserRepo{
		SQLCredo: sc.NewSQLCredo[User, Identity](params.DB, params.Driver, tableName, idColumn).
			WithDebugFunc(createDebugFunc(t)),
	}

	_, err := repo.InitSchema(ctx, params.Schema)
	require.NoError(t, err)

	for _, u := range testUserPtrs {
		_, err := repo.Create(ctx, u)
		require.NoError(t, err)
	}

	cnt, err := repo.Count(ctx)
	require.NoError(t, err)
	require.Equal(t, len(testUserPtrs), int(cnt))

	c := &testCaseData{
		ctx:          ctx,
		ctxCancel:    ctxCancel,
		TestUsers:    testUserValues,
		TestUserPtrs: testUserPtrs,
		db:           params.DB,
		UnderTest:    repo,
	}

	t.Cleanup(func() {
		c.TearDown(t)
	})

	return c, ctx
}

func (c *testCaseData) TearDown(t *testing.T) {
	_, err := c.UnderTest.DeleteAll(c.ctx)
	require.NoError(t, err)
	c.ctxCancel()
}

func CaseCreateUser(t *testing.T, params TestCaseParams) {
	c, ctx := newTestCase(t, params)

	expected := &User{"u99", "Gordon", ptr("Gibs"), newTime("1931-09-03")}
	_, err := c.UnderTest.Create(ctx, expected)
	assert.NoError(t, err)

	got, err := c.UnderTest.GetByID(ctx, expected.ID)
	assert.NoError(t, err)
	assert.Equal(t, *expected, got)
}

func CaseGetAllUsers(t *testing.T, params TestCaseParams) {
	c, ctx := newTestCase(t, params)

	got, err := c.UnderTest.GetAll(ctx)
	assert.NoError(t, err)
	assert.Equal(t, c.TestUsers, got)
}

func CaseGetUserByID(t *testing.T, params TestCaseParams) {
	c, ctx := newTestCase(t, params)

	got, err := c.UnderTest.GetByID(ctx, c.TestUsers[2].ID)
	assert.NoError(t, err)
	assert.Equal(t, c.TestUsers[2], got)
}

func CaseGetUsersByIDs(t *testing.T, params TestCaseParams) {
	c, ctx := newTestCase(t, params)

	ids := []Identity{c.TestUserPtrs[1].ID, c.TestUserPtrs[2].ID}
	got, err := c.UnderTest.GetByIDs(ctx, ids)
	assert.NoError(t, err)
	assert.Equal(t, []User{c.TestUsers[1], c.TestUsers[2]}, got)
}

func CaseDeleteUser(t *testing.T, params TestCaseParams) {
	c, ctx := newTestCase(t, params)

	_, err := c.UnderTest.Delete(ctx, c.TestUsers[1].ID)
	assert.NoError(t, err)

	_, err = c.UnderTest.GetByID(ctx, c.TestUsers[1].ID)
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

func CaseUpdateUser(t *testing.T, params TestCaseParams) {
	c, ctx := newTestCase(t, params)

	updated := c.TestUserPtrs[1]
	updated.FirstName = updated.FirstName + "_updated"

	_, err := c.UnderTest.Update(ctx, updated.ID, updated)
	assert.NoError(t, err)

	got, err := c.UnderTest.GetByID(ctx, c.TestUsers[1].ID)
	assert.NoError(t, err)
	assert.Equal(t, *updated, got)
}

func CaseValidatePageRequest(t *testing.T, params TestCaseParams) {
	c, ctx := newTestCase(t, params)

	_, err := c.UnderTest.GetPage(ctx, model.WithPageSize(0))
	assert.ErrorIs(t, err, model.ErrInvalidPageSize)
}

func CaseGetPage(t *testing.T, params TestCaseParams) {
	c, ctx := newTestCase(t, params)

	gotPage1, err := c.UnderTest.GetPage(ctx,
		model.WithPageNumber(0), model.WithPageSize(2), model.WithSort("id"))
	assert.NoError(t, err)
	assert.Equal(t, model.Page[User]{
		Number:     0,
		Size:       2,
		Total:      5,
		TotalPages: 3,
		Content:    c.TestUsers[0:2],
	}, gotPage1)

	gotPage2, err := c.UnderTest.GetPage(ctx,
		model.WithPageNumber(1), model.WithPageSize(2), model.WithSort("id"))
	assert.NoError(t, err)
	assert.Equal(t, model.Page[User]{
		Number:     1,
		Size:       2,
		Total:      5,
		TotalPages: 3,
		Content:    c.TestUsers[2:4],
	}, gotPage2)

	gotPage3, err := c.UnderTest.GetPage(ctx,
		model.WithPageNumber(2), model.WithPageSize(2), model.WithSort("id"))
	assert.NoError(t, err)
	assert.Equal(t, model.Page[User]{
		Number:     2,
		Size:       1,
		Total:      5,
		TotalPages: 3,
		Content:    c.TestUsers[4:],
	}, gotPage3)
}

func CaseCountUsers(t *testing.T, params TestCaseParams) {
	c, ctx := newTestCase(t, params)

	got, err := c.UnderTest.Count(ctx)
	assert.NoError(t, err)
	assert.Equal(t, len(c.TestUserPtrs), int(got))
}

func CaseCountByLastNameExists(t *testing.T, params TestCaseParams) {
	c, ctx := newTestCase(t, params)

	got, err := c.UnderTest.CountByLastNameExists(ctx)
	assert.NoError(t, err)
	assert.Equal(t, map[string]int{
		"with last_name":    3,
		"without last_name": 2,
	}, got)
}

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
	tableName = "users"
	idColumn  = "id"
)

type UserRepo struct {
	sc.SQLCredo[User, Identity]
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

func createTestUsers() ([]User, []*User) {
	values := []User{
		{"u0", "John", ptr("Smith"), newTime("1989-03-05")},
		{"u1", "Carl", nil, newTime("1973-01-09")},
		{"u2", "Ann", ptr("Stone"), newTime("1985-08-01")},
		{"u3", "Ann", ptr("Brick"), newTime("1987-03-02")},
		{"u4", "Antony", nil, newTime("1987-03-02")},
	}
	ptrs := wrapWithPtrs(values)
	return values, ptrs
}

func (r *UserRepo) CountByLastNameExists(ctx context.Context) (map[string]int, error) {
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

func createDebugFunc(t *testing.T) model.DebugFunc {
	return func(query string, args ...any) {
		t.Logf("query: [%s]; args: %+v\n", query, args)
	}
}

func newTime(input string) time.Time {
	result, err := time.Parse("2006-01-02", input)
	if err != nil {
		panic(err)
	}

	return result
}

func wrapWithPtrs[T comparable](input []T) []*T {
	result := make([]*T, 0, len(input))
	for _, i := range input {
		result = append(result, ptr(i))
	}
	return result
}

func ptr[T comparable](input T) *T {
	return &input
}
