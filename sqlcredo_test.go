package sqlcredo_test

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Klojer/sqlcredo/examples/users"
	"github.com/Klojer/sqlcredo/pkg/api"
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

func createTestUsers() ([]users.Object, []*users.Object) {
	values := []users.Object{
		{ID: "u0", FirstName: "John", LastName: ptr("Smith"), BirthDate: newTime("1989-03-05")},
		{ID: "u1", FirstName: "Carl", LastName: nil, BirthDate: newTime("1973-01-09")},
		{ID: "u2", FirstName: "Ann", LastName: ptr("Stone"), BirthDate: newTime("1987-03-01")},
		{ID: "u3", FirstName: "Ann", LastName: ptr("Brick"), BirthDate: newTime("1985-08-02")},
		{ID: "u4", FirstName: "Antony", LastName: nil, BirthDate: newTime("1987-03-02")},
	}
	ptrs := wrapWithPtrs(values)
	return values, ptrs
}

type testCaseData struct {
	ctx       context.Context
	ctxCancel func()

	TestUsers    []users.Object
	TestUserPtrs []*users.Object

	db        *sql.DB
	UnderTest *users.Repo
}

func newTestCase(t *testing.T, params TestCaseParams) (*testCaseData, context.Context) {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 30*time.Second)
	testUserValues, testUserPtrs := createTestUsers()

	repo := users.NewRepo(params.DB, params.Driver, createDebugFunc(t))

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

	expected := &users.Object{
		ID: "u99", FirstName: "Gordon",
		LastName: ptr("Gibs"), BirthDate: newTime("1931-09-03"),
	}
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

	ids := []users.Identity{c.TestUserPtrs[1].ID, c.TestUserPtrs[2].ID}
	got, err := c.UnderTest.GetByIDs(ctx, ids)
	assert.NoError(t, err)
	assert.Equal(t, []users.Object{c.TestUsers[1], c.TestUsers[2]}, got)
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

	_, err := c.UnderTest.GetPage(ctx, api.WithPageSize(0))
	assert.ErrorIs(t, err, api.ErrInvalidPageSize)
}

func CaseGetPage(t *testing.T, params TestCaseParams) {
	c, ctx := newTestCase(t, params)

	gotPage1, err := c.UnderTest.GetPage(ctx,
		api.WithPageNumber(0), api.WithPageSize(2))
	assert.NoError(t, err)
	assert.Equal(t, api.Page[users.Object]{
		Number:     0,
		Size:       2,
		Total:      5,
		TotalPages: 3,
		Content:    c.TestUsers[0:2],
	}, gotPage1)

	gotPage2, err := c.UnderTest.GetPage(ctx,
		api.WithPageNumber(1), api.WithPageSize(2))
	assert.NoError(t, err)
	assert.Equal(t, api.Page[users.Object]{
		Number:     1,
		Size:       2,
		Total:      5,
		TotalPages: 3,
		Content:    c.TestUsers[2:4],
	}, gotPage2)

	gotPage3, err := c.UnderTest.GetPage(ctx,
		api.WithPageNumber(2), api.WithPageSize(2))
	assert.NoError(t, err)
	assert.Equal(t, api.Page[users.Object]{
		Number:     2,
		Size:       1,
		Total:      5,
		TotalPages: 3,
		Content:    c.TestUsers[4:],
	}, gotPage3)
}

func CaseGetPageCustomOrder(t *testing.T, params TestCaseParams) {
	c, ctx := newTestCase(t, params)

	gotPage, err := c.UnderTest.GetPage(ctx,
		api.WithPageNumber(0),
		api.WithPageSize(uint(len(c.TestUsers))),
		api.WithSortBy("first_name"),
		api.WithSortBy("last_name"),
	)
	assert.NoError(t, err)
	assert.Equal(t, strings.TrimSpace(`
Ann Brick
Ann Stone
Antony
Carl
John Smith
`),
		usersToString(gotPage.Content...))
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

func createDebugFunc(t *testing.T) api.DebugFunc {
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

func usersToString(objects ...users.Object) string {
	res := bytes.Buffer{}

	for _, o := range objects {
		if o.LastName != nil {
			res.WriteString(fmt.Sprintf("%s %s", o.FirstName, *o.LastName))
		} else {
			res.WriteString(o.FirstName)
		}
		res.WriteString("\n")
	}

	return strings.TrimSpace(res.String())
}
