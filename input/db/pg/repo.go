package pg

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/rs/zerolog"

	"github.com/gofoji/foji/input/db"
)

// DB is the common interface for database operations.
type DB interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type Repo struct {
	db        DB
	typeCache map[int64]string
	logger    zerolog.Logger
}

func New(db DB, logger zerolog.Logger) Repo {
	return Repo{db: db, typeCache: map[int64]string{}, logger: logger}
}

func (r Repo) GetTypeNameByID(ctx context.Context, id int64) (string, error) {
	const query = `select typname from pg_type where oid = $1`

	result := ""

	return result, r.db.QueryRow(ctx, query, id).Scan(&result)
}

func (r Repo) Get(ctx context.Context) (db.DB, error) {
	ss := make(map[string]*db.Schema)

	err := r.processTables(ctx, ss)
	if err != nil {
		return nil, err
	}

	err = r.processColumns(ctx, ss)
	if err != nil {
		return nil, err
	}

	err = r.processIndexes(ctx, ss)
	if err != nil {
		return nil, err
	}

	err = r.processForeignKeys(ctx, ss)
	if err != nil {
		return nil, err
	}

	err = r.processEnums(ctx, ss)
	if err != nil {
		return nil, err
	}

	return ss, nil
}

func (r Repo) processTables(ctx context.Context, ss db.DB) error {
	r.logger.Debug().Msg("Loading Tables")

	tt, err := r.GetTables(ctx)
	if err != nil {
		return fmt.Errorf("GetTables:%w", err)
	}

	for _, t := range tt {
		val, ok := ss[t.Schema]
		if !ok {
			val = &db.Schema{
				Name: t.Schema,
			}
			ss[val.Name] = val
		}

		table := t.toDB(val)
		val.Tables = append(val.Tables, &table)
	}

	return nil
}

func (r Repo) processEnums(ctx context.Context, ss db.DB) error {
	r.logger.Debug().Msg("Loading Enums")

	ee, err := r.GetEnums(ctx)
	if err != nil {
		return fmt.Errorf("GetEnums:%w", err)
	}

	for _, e := range ee {
		val, ok := ss[e.Schema]
		if !ok {
			val = &db.Schema{
				Name: e.Schema,
			}
			ss[val.Name] = val
		}

		enum := e.toDB(val)
		val.Enums = append(val.Enums, &enum)
	}

	return nil
}

func (r Repo) processColumns(ctx context.Context, ss db.DB) error {
	r.logger.Debug().Msg("Loading Columns")

	cc, err := r.GetColumns(ctx)
	if err != nil {
		return fmt.Errorf("GetColumns:%w", err)
	}

	for _, c := range cc {
		table, ok := ss.GetTable(c.Schema, c.Table)
		if !ok {
			r.logger.Debug().Msgf("Table (%s.%s) not found for Column (%s), skipping", c.Schema, c.Table, c.Name)

			continue
		}

		column := c.toDB(table)
		table.Columns = append(table.Columns, &column)
	}

	return nil
}

func (r Repo) processIndexes(ctx context.Context, ss db.DB) error {
	r.logger.Debug().Msg("Loading Indexes")

	ii, err := r.GetIndexes(ctx)
	if err != nil {
		return fmt.Errorf("GetIndexes:%w", err)
	}

	for _, i := range ii {
		table, ok := ss.GetTable(i.Schema, i.Table)
		if !ok {
			r.logger.Warn().Msgf("Table (%s.%s) not found for Index (%s), skipping", i.Schema, i.Table, i.Name)

			continue
		}

		cols, err := table.GetColumnsByName(i.Columns)
		if err != nil {
			if err.Error() == "expr" {
				r.logger.Warn().Msgf("Unsupported expression Index (%s), skipping", i.Name)
			} else {
				r.logger.Warn().Msgf("Column (%s) not found for Index (%s), skipping", err, i.Name)
			}

			continue
		}

		index := i.toDB(cols)
		if index.IsPrimary {
			for _, c := range cols {
				c.IsPrimaryKey = true
				table.PrimaryKeys = append(table.PrimaryKeys, c)
			}
		}

		table.Indexes = append(table.Indexes, &index)
	}

	return nil
}

func (r Repo) processForeignKeys(ctx context.Context, ss db.DB) error {
	r.logger.Debug().Msg("Loading Foreign Keys")

	ff, err := r.GetForeignKeys(ctx)
	if err != nil {
		return fmt.Errorf("GetForeignKeys:%w", err)
	}

	for _, f := range ff {
		l := r.logger.With().Str("ForeignKey", f.Name).Logger()

		table, ok := ss.GetTable(f.Schema, f.Table)
		if !ok {
			l.Debug().Msgf("Table (%s.%s) not found, skipping", f.Schema, f.Table)

			continue
		}

		cols, err := table.GetColumnsByName(f.Columns)
		if err != nil {
			l.Debug().Msgf("Column (%s) not found, skipping", err)

			continue
		}

		fTable, ok := ss.GetTable(f.ForeignSchema, f.ForeignTable)
		if !ok {
			l.Debug().Msgf("Table (%s.%s) not found, skipping", f.ForeignSchema, f.ForeignTable)

			continue
		}

		fCols, err := fTable.GetColumnsByName(f.ForeignColumns)
		if err != nil {
			l.Debug().Msgf("Column (%s) not found, skipping", err)

			continue
		}

		fk := f.toDB(cols, fCols)
		table.ForeignKeys = append(table.ForeignKeys, &fk)
		fTable.References = append(fTable.References, &fk)

		for _, c := range cols {
			c.ForeignKey = &fk
		}
	}

	return nil
}

func (t Table) toDB(schema *db.Schema) db.Table {
	return db.Table{
		ID:       t.ID,
		Schema:   schema,
		Name:     t.Name,
		Type:     t.Type,
		Comment:  t.Comment,
		Columns:  nil,
		Indexes:  nil,
		ReadOnly: !t.CanInsert && !t.CanUpdate && !t.CanDelete,
	}
}

func (i Index) toDB(cols []*db.Column) db.Index {
	return db.Index{
		Name:      i.Name,
		IsUnique:  i.IsUnique,
		IsPrimary: i.IsPrimary,
		Columns:   cols,
		Comment:   i.Comment,
	}
}

func (f ForeignKey) toDB(cols, fCols []*db.Column) db.ForeignKey {
	return db.ForeignKey{
		Name:           f.Name,
		Columns:        cols,
		ForeignColumns: fCols,
		Comment:        f.Comment,
	}
}

func (c Column) toDB(table *db.Table) db.Column {
	return db.Column{
		Table:      table,
		Name:       c.Name,
		Type:       c.Type,
		Nullable:   c.Nullable,
		HasDefault: c.HasDefault,
		Comment:    c.Comment,
		Ordinal:    c.Ordinal,
	}
}

func (e Enum) toDB(schema *db.Schema) db.Enum {
	return db.Enum{
		ID:      e.ID,
		Name:    e.Name,
		Values:  e.Values,
		Comment: e.Comment,
		Schema:  schema,
	}
}
