package output

import (
	"errors"
	"fmt"
	"strings"

	"github.com/codemodus/kace"
	"github.com/gofoji/plates"
	"github.com/rs/zerolog"

	"github.com/gofoji/foji/cfg"
	"github.com/gofoji/foji/input/sql"
	"github.com/gofoji/foji/stringlist"
)

const (
	SQLAll   = "SQLAll"
	SQLFiles = "SQLFiles"
	SQLFile  = "SQLFile"
	SQLQuery = "SQLQuery"
)

func HasSQLOutput(o cfg.Output) bool {
	return hasAnyOutput(o, SQLAll, SQLFiles, SQLFile, SQLQuery)
}

func SQL(p cfg.Process, fn cfg.FileHandler, l zerolog.Logger, fileGroups sql.FileGroups, simulate bool) error {
	base := SQLContext{
		Context:    NewContext(p, l),
		FileGroups: fileGroups,
	}

	runner := NewProcessRunner(p.RootDir, fn, l, simulate)

	err := runner.process(p.Output[SQLAll], &base)
	if err != nil {
		return err
	}

	for _, ff := range fileGroups {
		ctx := SQLFileGroupContext{
			SQLContext: base,
			Files:      ff,
		}

		err := runner.process(p.Output[SQLFiles], &ctx)
		if err != nil {
			return err
		}

		for _, f := range ff {
			ctx := SQLFileContext{
				SQLContext: base,
				File:       f,
			}

			err := runner.process(p.Output[SQLFile], &ctx)
			if err != nil {
				return err
			}

			for _, q := range f.Queries {
				ctx := SQLQueryContext{
					SQLContext: base,
					Query:      q,
				}

				err := runner.process(p.Output[SQLQuery], &ctx)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

//nolint:recvcheck
type SQLContext struct {
	Context
	sql.FileGroups
	Imports
}

type SQLFileGroupContext struct {
	SQLContext

	Files []sql.File
}

type SQLFileContext struct {
	SQLContext
	sql.File
}

type SQLQueryContext struct {
	SQLContext
	sql.Query
}

func (q SQLContext) Parameterize(cc sql.Params, format, pkg string) string {
	ss := make(stringlist.Strings, len(cc))

	for x := range cc {
		ss[x] = fmt.Sprintf(format, kace.Camel(cc[x].Name), q.GetType(cc[x], pkg))
	}

	return strings.Join(ss, ", ")
}

func (q SQLContext) GetType(c *sql.Param, pkg string) string {
	if c.Generated {
		return c.Type
	}

	return ResolveType(q.Maps, func(t string) string { return q.CheckPackage(t, pkg) }, c.Type, c.Nullable, c.Path())
}

var errMissingParam = errors.New("missing Param.Package")

func (q *SQLContext) Init() error {
	name, ok := q.Params.HasString("Package")
	if !ok {
		return errMissingParam
	}

	for _, set := range q.FileGroups {
		for _, ff := range set {
			for _, qry := range ff.Queries {
				q.CheckPackage(qry.Result.Type, name)

				for _, p := range qry.Params {
					q.CheckPackage(p.Type, name)
				}
			}
		}
	}

	return nil
}

func (q *SQLFileGroupContext) Init() error {
	name, ok := q.Params.HasString("Package")
	if !ok {
		return errMissingParam
	}

	for _, ff := range q.Files {
		for _, qry := range ff.Queries {
			q.CheckPackage(qry.Result.Type, name)

			for _, p := range qry.Params {
				q.CheckPackage(p.Type, name)
			}
		}
	}

	return nil
}

func (q *SQLFileContext) Init(p *plates.Factory) error {
	name, ok := q.Params.HasString("Package")
	if !ok {
		return errMissingParam
	}

	if strings.Contains(name, "{{") {
		var err error

		name, err = p.From(name).To(q)
		if err != nil {
			return fmt.Errorf("mapping Package name:%w", err)
		}

		q.Params["Package"] = name
	}

	for _, qry := range q.Queries {
		q.CheckPackage(qry.Result.Type, name)

		for _, p := range qry.Params {
			q.CheckPackage(p.Type, name)
		}
	}

	return nil
}

func (q *SQLQueryContext) Init() error {
	name, ok := q.Process.Params.HasString("Package")
	if !ok {
		return errMissingParam
	}

	for _, p := range q.Params {
		q.CheckPackage(p.Type, name)
	}

	return nil
}
