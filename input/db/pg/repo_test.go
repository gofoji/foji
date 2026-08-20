package pg_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gofoji/foji/input/db"
	"github.com/gofoji/foji/input/db/pg"
)

/*
These tests cover schema introspection without a live Postgres by stubbing the
pg.DB interface. Repo.Get and the toDB conversions it depends on are otherwise
only reachable through `make test_generate` against a real database.
*/

// queryName reads the `--# <name>` marker the generated queries open with, so a
// fixture can be selected without matching on the whole SQL body.
func queryName(sql string) string {
	line, _, _ := strings.Cut(sql, "\n")

	fields := strings.Fields(line)
	if len(fields) < 2 || fields[0] != "--#" {
		return ""
	}

	return fields[1]
}

// fakeDB answers each generated introspection query from canned rows.
type fakeDB struct {
	rows map[string][][]any
	errs map[string]error
}

func (f fakeDB) Query(_ context.Context, sql string, _ ...any) (pgx.Rows, error) {
	name := queryName(sql)

	err := f.errs[name]
	if err != nil {
		return nil, err
	}

	return &fakeRows{rows: f.rows[name]}, nil
}

func (f fakeDB) Exec(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}

func (f fakeDB) QueryRow(_ context.Context, _ string, _ ...any) pgx.Row {
	return nil
}

// fakeRows replays a slice of rows. Scan assigns by position, so a fixture row
// has to match the column order of the generated query.
type fakeRows struct {
	rows [][]any
	pos  int
}

func (r *fakeRows) Next() bool {
	r.pos++

	return r.pos <= len(r.rows)
}

func (r *fakeRows) Scan(dest ...any) error {
	row := r.rows[r.pos-1]
	if len(dest) != len(row) {
		return fmt.Errorf("scan: fixture has %d columns, query wants %d", len(row), len(dest))
	}

	for i, d := range dest {
		target := reflect.ValueOf(d).Elem()
		value := reflect.ValueOf(row[i])

		if !value.Type().AssignableTo(target.Type()) {
			return fmt.Errorf("scan column %d: cannot assign %s to %s", i, value.Type(), target.Type())
		}

		target.Set(value)
	}

	return nil
}

func (r *fakeRows) Close()                                       {}
func (r *fakeRows) Err() error                                   { return nil }
func (r *fakeRows) CommandTag() pgconn.CommandTag                { return pgconn.CommandTag{} }
func (r *fakeRows) FieldDescriptions() []pgconn.FieldDescription { return nil }
func (r *fakeRows) Values() ([]any, error)                       { return nil, nil }
func (r *fakeRows) RawValues() [][]byte                          { return nil }
func (r *fakeRows) Conn() *pgx.Conn                              { return nil }

// Column orders below mirror the q.Scan calls in queries_gen.go:
//
//	GetTables:      id, schema, name, type, canUpdate, canInsert, canDelete, comment
//	GetColumns:     schema, table, name, ordinal, nullable, hasDefault, type, comment
//	GetIndexes:     id, schema, name, table, isUnique, isPrimary, columns, comment
//	GetForeignKeys: id, schema, name, table, columns, foreignSchema, foreignTable, foreignColumns, comment
//	GetEnums:       id, name, schema, values, comment
func fixture() map[string][][]any {
	return map[string][][]any{
		"GetTables": {
			{uint32(1), "public", "users", "table", true, true, true, "people"},
			{uint32(2), "public", "orders", "table", true, true, true, ""},
			{uint32(3), "public", "user_summary", "view", false, false, false, ""},
			{uint32(4), "app", "settings", "table", true, true, true, ""},
		},
		"GetColumns": {
			{"public", "users", "id", int16(1), false, true, "uuid", "pk"},
			{"public", "users", "email", int16(2), false, false, "varchar", ""},
			{"public", "orders", "id", int16(1), false, true, "uuid", ""},
			{"public", "orders", "user_id", int16(2), true, false, "uuid", ""},
			{"app", "settings", "key", int16(1), false, false, "varchar", ""},
		},
		"GetIndexes": {
			{uint32(10), "public", "users_pkey", "users", true, true, []string{"id"}, ""},
			{uint32(11), "public", "users_email_key", "users", true, false, []string{"email"}, ""},
			{uint32(12), "public", "orders_pkey", "orders", true, true, []string{"id"}, ""},
		},
		"GetForeignKeys": {
			{uint32(20), "public", "orders_user_id_fkey", "orders", []string{"user_id"}, "public", "users", []string{"id"}, "owner"},
		},
		"GetEnums": {
			{uint32(30), "status", "public", []string{"active", "inactive"}, "order status"},
		},
	}
}

func getSchemas(t *testing.T, rows map[string][][]any, errs map[string]error) (db.DB, error) {
	t.Helper()

	repo := pg.New(fakeDB{rows: rows, errs: errs}, zerolog.Nop())

	return repo.Get(context.Background())
}

func TestRepoGet_GroupsTablesBySchema(t *testing.T) {
	ss, err := getSchemas(t, fixture(), nil)
	require.NoError(t, err)

	require.Len(t, ss, 2)
	require.Contains(t, ss, "public")
	require.Contains(t, ss, "app")

	assert.Len(t, ss["public"].Tables, 3)
	assert.Len(t, ss["app"].Tables, 1)

	users, ok := ss["public"].GetTable("users")
	require.True(t, ok)
	assert.Equal(t, uint32(1), users.ID)
	assert.Equal(t, "people", users.Comment)
	assert.Equal(t, "public", users.Schema.Name)
}

func TestRepoGet_DerivesReadOnlyFromPrivileges(t *testing.T) {
	ss, err := getSchemas(t, fixture(), nil)
	require.NoError(t, err)

	users, ok := ss["public"].GetTable("users")
	require.True(t, ok)
	assert.False(t, users.ReadOnly)

	view, ok := ss["public"].GetTable("user_summary")
	require.True(t, ok)
	assert.True(t, view.ReadOnly, "a table with no insert/update/delete is read only")
}

func TestRepoGet_AttachesColumnsToTables(t *testing.T) {
	ss, err := getSchemas(t, fixture(), nil)
	require.NoError(t, err)

	users, ok := ss["public"].GetTable("users")
	require.True(t, ok)
	require.Len(t, users.Columns, 2)

	assert.Equal(t, []string{"id", "email"}, []string(users.Columns.Names()))

	email := users.GetColumnByName("email")
	require.NotNil(t, email)
	assert.Equal(t, "varchar", email.Type)
	assert.Equal(t, int16(2), email.Ordinal)
	assert.False(t, email.Nullable)
	assert.False(t, email.HasDefault)
	assert.Equal(t, "public.users.email", email.Path())
	assert.Equal(t, users, email.Table)

	id := users.GetColumnByName("id")
	require.NotNil(t, id)
	assert.True(t, id.HasDefault)
}

func TestRepoGet_MarksPrimaryKeys(t *testing.T) {
	ss, err := getSchemas(t, fixture(), nil)
	require.NoError(t, err)

	users, ok := ss["public"].GetTable("users")
	require.True(t, ok)

	require.True(t, users.HasPrimaryKey())
	require.Len(t, users.PrimaryKeys, 1)
	assert.Equal(t, "id", users.PrimaryKeys[0].Name)
	assert.True(t, users.GetColumnByName("id").IsPrimaryKey)
	assert.False(t, users.GetColumnByName("email").IsPrimaryKey)

	pk := users.GetPK()
	require.NotNil(t, pk)
	assert.Equal(t, "id", pk.Name)

	// A unique but non-primary index must not mark its columns.
	require.Len(t, users.Indexes, 2)
	assert.Equal(t, "users_email_key", users.Indexes[1].Name)
	assert.True(t, users.Indexes[1].IsUnique)
	assert.False(t, users.Indexes[1].IsPrimary)
}

func TestRepoGet_WiresForeignKeysBothDirections(t *testing.T) {
	ss, err := getSchemas(t, fixture(), nil)
	require.NoError(t, err)

	orders, ok := ss["public"].GetTable("orders")
	require.True(t, ok)
	users, ok := ss["public"].GetTable("users")
	require.True(t, ok)

	require.Len(t, orders.ForeignKeys, 1)
	fk := orders.ForeignKeys[0]
	assert.Equal(t, "orders_user_id_fkey", fk.Name)
	assert.Equal(t, "owner", fk.Comment)
	require.Len(t, fk.Columns, 1)
	assert.Equal(t, "user_id", fk.Columns[0].Name)
	require.Len(t, fk.ForeignColumns, 1)
	assert.Equal(t, "id", fk.ForeignColumns[0].Name)

	// The referenced table records the inbound reference.
	require.Len(t, users.References, 1)
	assert.Equal(t, fk, users.References[0])

	// The source column points back at the key.
	assert.Equal(t, fk, orders.GetColumnByName("user_id").ForeignKey)
	assert.Nil(t, orders.GetColumnByName("id").ForeignKey)
}

func TestRepoGet_AttachesEnums(t *testing.T) {
	ss, err := getSchemas(t, fixture(), nil)
	require.NoError(t, err)

	require.Len(t, ss["public"].Enums, 1)

	e := ss["public"].Enums[0]
	assert.Equal(t, uint32(30), e.ID)
	assert.Equal(t, "status", e.Name)
	assert.Equal(t, "order status", e.Comment)
	assert.Equal(t, []string{"active", "inactive"}, []string(e.Values))
	assert.Equal(t, "public.status", e.Path())
	assert.Equal(t, "public", e.Schema.Name)
}

func TestRepoGet_CreatesSchemaForEnumOnlySchema(t *testing.T) {
	rows := fixture()
	rows["GetTables"] = nil
	rows["GetColumns"] = nil
	rows["GetIndexes"] = nil
	rows["GetForeignKeys"] = nil

	ss, err := getSchemas(t, rows, nil)
	require.NoError(t, err)

	require.Contains(t, ss, "public")
	assert.Empty(t, ss["public"].Tables)
	assert.Len(t, ss["public"].Enums, 1)
}

func TestRepoGet_SkipsOrphans(t *testing.T) {
	rows := fixture()
	rows["GetColumns"] = append(rows["GetColumns"],
		[]any{"public", "missing_table", "x", int16(1), false, false, "text", ""})
	rows["GetIndexes"] = append(rows["GetIndexes"],
		[]any{uint32(13), "public", "missing_idx", "missing_table", true, true, []string{"x"}, ""},
		[]any{uint32(14), "public", "bad_col_idx", "users", true, false, []string{"nope"}, ""})
	rows["GetForeignKeys"] = append(rows["GetForeignKeys"],
		[]any{uint32(21), "public", "fk_missing_table", "missing_table", []string{"x"}, "public", "users", []string{"id"}, ""},
		[]any{uint32(22), "public", "fk_missing_target", "orders", []string{"user_id"}, "public", "gone", []string{"id"}, ""},
		[]any{uint32(23), "public", "fk_missing_col", "orders", []string{"nope"}, "public", "users", []string{"id"}, ""})

	ss, err := getSchemas(t, rows, nil)
	require.NoError(t, err, "orphan rows are skipped, not fatal")

	users, ok := ss["public"].GetTable("users")
	require.True(t, ok)
	assert.Len(t, users.Indexes, 2, "index naming an unknown column is skipped")

	orders, ok := ss["public"].GetTable("orders")
	require.True(t, ok)
	assert.Len(t, orders.ForeignKeys, 1, "keys naming an unknown table or column are skipped")

	_, ok = ss["public"].GetTable("missing_table")
	assert.False(t, ok)
}

func TestRepoGet_PropagatesQueryErrors(t *testing.T) {
	errBoom := errors.New("boom")

	for _, query := range []string{"GetTables", "GetColumns", "GetIndexes", "GetForeignKeys", "GetEnums"} {
		t.Run(query, func(t *testing.T) {
			_, err := getSchemas(t, fixture(), map[string]error{query: errBoom})
			require.Error(t, err)
			assert.ErrorIs(t, err, errBoom)
			assert.Contains(t, err.Error(), query)
		})
	}
}
