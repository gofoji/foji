package pg

import (
	"context"

	"github.com/gofoji/foji/input/sql"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

// DB is the common interface for database operations.
type DB interface {
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
	QueryRow(context.Context, string, ...interface{}) pgx.Row
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
		return "", err
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
		return err
	}

	for i, oid := range sd.ParamOIDs {
		t, err := r.GetTypeNameByID(ctx, oid)
		if err != nil {
			return errors.Wrapf(err, "unable to locate data type: %d", oid)
		}

		q.Params[i].Type = t
	}

	q.Result.IsSingleTable = true

	var (
		fields sql.Params
		table  uint32
	)

	for i, f := range sd.Fields {
		t, err := r.GetTypeNameByID(ctx, f.DataTypeOID)
		if err != nil {
			return errors.Wrapf(err, "unable to locate data type: %d", f.DataTypeOID)
		}

		name := string(f.Name)
		if fields != nil && fields.ByName(name) != nil {
			schema, tableName, err := r.GetTableNameByID(ctx, f.TableOID)
			if err != nil {
				return errors.Wrapf(err, "unable to locate data type: %d", f.DataTypeOID)
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
			return errors.Wrap(err, "unable to locate result table")
		}
	}

	return nil
}
