package output

import (
	"fmt"
	"strings"

	"github.com/codemodus/kace"
	"github.com/gofoji/foji/cfg"
	"github.com/gofoji/foji/input/sql"
	"github.com/gofoji/foji/stringlist"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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

func SQL(p cfg.Process, fn cfg.FileHandler, logger logrus.FieldLogger, fileGroups sql.FileGroups, simulate bool) error {
	base := SQLContext{
		Context:    Context{Process: p, Logger: logger},
		FileGroups: fileGroups,
	}

	err := invokeProcess(p.Output[SQLAll], p.RootDir, fn, logger, &base, simulate)
	if err != nil {
		return err
	}
	for _, ff := range fileGroups {
		ctx := SQLFileGroupContext{
			SQLContext: base,
			Files:      ff,
		}

		err := invokeProcess(p.Output[SQLFiles], p.RootDir, fn, logger, &ctx, simulate)
		if err != nil {
			return err
		}

		for _, f := range ff {
			ctx := SQLFileContext{
				SQLContext: base,
				File:       f,
			}
			err := invokeProcess(p.Output[SQLFile], p.RootDir, fn, logger, &ctx, simulate)
			if err != nil {
				return err
			}

			for _, q := range f.Queries {
				ctx := SQLQueryContext{
					SQLContext: base,
					Query:      q,
				}
				err := invokeProcess(p.Output[SQLQuery], p.RootDir, fn, logger, &ctx, simulate)
				if err != nil {
					return err
				}
			}

		}
	}

	return nil
}

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

	pp := strings.Split(c.Path(), ".")
	for i := range pp {
		p := strings.Join(pp[i:], ".")
		t, ok := q.Maps.Type["."+p]
		if ok {
			return stripPackage(t, pkg)
		}
	}

	if c.Nullable {
		t, ok := q.Maps.Nullable[c.Type]
		if ok {
			return stripPackage(t, pkg)
		}
	}

	t, ok := q.Maps.Type[c.Type]
	if ok {
		return stripPackage(t, pkg)
	}

	if strings.ContainsAny(c.Type, "./") {
		// Qualified Name
		return q.CheckPackage(c.Type, pkg)
	}

	return fmt.Sprintf("UNKNOWN:path(%s):type(%s)", c.Path(), c.Type)
}

func (q *SQLContext) Init() error {
	name, ok := q.Params.HasString("Package")
	if !ok {
		return errors.New("missing Param.Package")
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
		return errors.New("missing Param.Package")
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

func (q *SQLFileContext) Init() error {
	name, ok := q.Params.HasString("Package")
	if !ok {
		return errors.New("missing Param.Package")
	}

	for _, qry := range q.File.Queries {
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
		return errors.New("missing Param.Package")
	}

	for _, p := range q.Query.Params {
		q.CheckPackage(p.Type, name)
	}

	return nil
}
