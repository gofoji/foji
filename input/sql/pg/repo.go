package pg

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/gofoji/foji/input/sql"
)

// DB is the common interface for database operations.
type DB interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Prepare(ctx context.Context, name, sql string) (sd *pgconn.StatementDescription, err error)
}

type Repo struct {
	db        DB
	typeCache map[uint32]DataType
}

type DataType struct {
	ID       uint32
	Name     string
	Nullable bool
}

func New(db DB) *Repo {
	return &Repo{db: db, typeCache: map[uint32]DataType{}}
}

func (r *Repo) GetTypeNameByID(ctx context.Context, id uint32) (DataType, error) {
	if result, ok := r.typeCache[id]; ok {
		return result, nil
	}

	const qry = `SELECT typname, NOT typnotnull AS nullable
FROM pg_type
WHERE oid = $1`

	result := DataType{ID: id}

	err := r.db.QueryRow(ctx, qry, id).Scan(&result.Name, &result.Nullable)
	if err != nil {
		return DataType{}, fmt.Errorf("GetTypeNameByID: %w", err)
	}

	r.typeCache[id] = result

	return result, nil
}

func (r *Repo) GetTableNameByID(ctx context.Context, id uint32) (string, string, error) {
	const qry = `SELECT nspname as schema, relname as name
FROM pg_class c
         JOIN pg_namespace n ON c.relnamespace = n.oid
WHERE c.oid = $1`

	var schema, name string

	return schema, name, r.db.QueryRow(ctx, qry, id).Scan(&schema, &name)
}

func (r *Repo) DescribeQuery(ctx context.Context, q *sql.Query) error {
	sd, err := r.db.Prepare(ctx, q.Name, q.SQL)
	if err != nil {
		return fmt.Errorf("prepare: %w", err)
	}

	for i, oid := range sd.ParamOIDs {
		t, err := r.GetTypeNameByID(ctx, oid)
		if err != nil {
			return fmt.Errorf("unable to locate data type: %d: %w", oid, err)
		}

		q.Params[i].Type = t.Name
	}

	q.Result.IsSingleTable = true

	var (
		fields sql.Params
		table  uint32
	)

	for i, f := range sd.Fields {
		p, err := r.describeParam(ctx, q, f, fields, i)
		if err != nil {
			return err
		}

		fields = append(fields, &p)

		if table != 0 && table != f.TableOID {
			q.Result.IsSingleTable = false
		}

		table = f.TableOID
	}

	q.Result.Params = fields

	if q.Result.IsSingleTable && table != 0 {
		q.Result.Schema, q.Result.Table, err = r.GetTableNameByID(ctx, table)
		if err != nil {
			return fmt.Errorf("unable to locate result table: %w", err)
		}
	}

	return nil
}

func (r *Repo) describeParam(ctx context.Context, q *sql.Query, f pgconn.FieldDescription, fields sql.Params, i int) (sql.Param, error) {
	t, err := r.GetTypeNameByID(ctx, f.DataTypeOID)
	if err != nil {
		return sql.Param{}, fmt.Errorf("unable to locate data type: %d: %w", f.DataTypeOID, err)
	}

	name := f.Name
	if fields != nil && fields.ByName(name) != nil {
		schema, tableName, err := r.GetTableNameByID(ctx, f.TableOID)
		if err != nil {
			return sql.Param{}, fmt.Errorf("unable to locate data type: %d: %w", f.DataTypeOID, err)
		}

		name = tableName + "_" + name
		if fields.ByName(name) != nil {
			name = schema + "_" + name
		}
	}

	p := sql.Param{
		Ordinal:       i,
		QueryPosition: i,
		Name:          name,
		Type:          t.Name,
		Nullable:      t.Nullable,
		TypeID:        f.DataTypeOID,
		Query:         q,
	}

	return p, nil
}
