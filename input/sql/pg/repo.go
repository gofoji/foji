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
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
	Prepare(ctx context.Context, name, sql string) (sd *pgconn.StatementDescription, err error)
}

type Repo struct {
	db        DB
	typeCache map[uint32]string
}

func New(db DB) *Repo {
	return &Repo{db: db, typeCache: map[uint32]string{}}
}

func (r *Repo) GetTypeNameByID(ctx context.Context, id uint32) (string, error) {
	const qry = `select typname from pg_type where oid = $1`

	if result, ok := r.typeCache[id]; ok {
		return result, nil
	}

	result := ""

	err := r.db.QueryRow(ctx, qry, id).Scan(&result)
	if err != nil {
		return "", fmt.Errorf("scan: %w", err)
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

		q.Params[i].Type = t
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

	name := string(f.Name)
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
		Type:          t,
		TypeID:        f.DataTypeOID,
		Query:         q,
	}

	return p, nil
}
