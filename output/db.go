package output

import (
	"fmt"
	"strings"

	"github.com/codemodus/kace"
	"github.com/rs/zerolog"

	"github.com/gofoji/foji/cfg"
	"github.com/gofoji/foji/input/db"
	"github.com/gofoji/foji/stringlist"
)

const (
	DBAll    = "DbAll"
	DBSchema = "DbSchema"
	DBTable  = "DbTable"
	DBEnums  = "DbEnums"
	DBEnum   = "DbEnum"
)

func HasDBOutput(o cfg.Output) bool {
	return hasAnyOutput(o, DBAll, DBSchema, DBTable, DBEnum, DBEnums)
}

func DB(p cfg.Process, fn cfg.FileHandler, logger zerolog.Logger, schemas db.DB, simulate bool) error {
	ctx := SchemasContext{
		Context: Context{Process: p, Logger: logger},
		DB:      schemas,
	}

	runner := NewProcessRunner(p.RootDir, fn, logger, simulate)

	err := runner.process(p.Output[DBAll], &ctx)
	if err != nil {
		return err
	}

	for _, s := range schemas {
		schemaCtx := SchemaContext{
			Schema:         *s,
			SchemasContext: ctx,
		}

		err := runner.process(p.Output[DBSchema], &schemaCtx)
		if err != nil {
			return err
		}

		for _, t := range s.Tables {
			tableCtx := TableContext{
				Table:          *t,
				SchemasContext: ctx,
			}

			err := runner.process(p.Output[DBTable], &tableCtx)
			if err != nil {
				return err
			}
		}

		enumsCtx := EnumsContext{
			Enums:          s.Enums,
			SchemasContext: ctx,
		}

		err = runner.process(p.Output[DBEnums], &enumsCtx)
		if err != nil {
			return err
		}

		for _, e := range s.Enums {
			enumCtx := EnumContext{
				Enum:           *e,
				SchemasContext: ctx,
			}

			err := runner.process(p.Output[DBEnum], &enumCtx)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

type SchemasContext struct {
	Context
	Imports
	db.DB
}

type SchemaContext struct {
	db.Schema
	SchemasContext
}

type TableContext struct {
	db.Table
	SchemasContext
}

type EnumsContext struct {
	db.Enums
	SchemasContext
}

type EnumContext struct {
	db.Enum
	SchemasContext
}

type ResourceMap map[string]Resource

type Resource struct {
	Name       string
	Properties Properties
	Table      *db.Table
	PK         *Property
}

type Properties []*Property

type Property struct {
	Name   string
	Type   string
	Format string
}

func (s *SchemasContext) Parameterize(cc db.Columns, format, pkg string) string {
	ss := make(stringlist.Strings, len(cc))
	for x := range cc {
		ss[x] = fmt.Sprintf(format, kace.Camel(cc[x].Name), s.GetType(cc[x], pkg))
	}

	return strings.Join(ss, ", ")
}

func (s SchemasContext) GetType(c *db.Column, pkg string) string {
	pp := strings.Split(c.Path(), ".")
	for i := range pp {
		p := strings.Join(pp[i:], ".")

		t, ok := s.Maps.Type["."+p]
		if ok {
			return s.CheckPackage(t, pkg)
		}
	}

	if c.Nullable {
		t, ok := s.Maps.Nullable[c.Type]
		if ok {
			return s.CheckPackage(t, pkg)
		}
	}

	t, ok := s.Maps.Type[c.Type]
	if ok {
		return s.CheckPackage(t, pkg)
	}

	return fmt.Sprintf("UNKNOWN:path(%s):type(%s)", c.Path(), c.Type)
}

const ValidTypeElems = 2

// Example type declaration:
// string,date-time

func (s SchemasContext) PropertyFromDB(c *db.Column) *Property {
	if c == nil {
		return nil
	}

	format := ""

	t := s.GetType(c, "")
	if strings.Contains(t, ",") {
		tt := strings.Split(t, ",")
		if len(tt) == ValidTypeElems {
			t = tt[0]
			format = tt[1]
		} else {
			s.Logger.Error().Msgf("Schema Column: %s, Invalid Type declaration:%s", c.Path(), t)
		}
	}

	return &Property{
		Name:   c.Name,
		Type:   t,
		Format: format,
	}
}

func (s SchemasContext) PropertiesFromDB(cc db.Columns) Properties {
	result := Properties{}
	for _, c := range cc {
		result = append(result, s.PropertyFromDB(c))
	}

	return result
}

func (s SchemasContext) Resources() ResourceMap {
	result := ResourceMap{}

	for _, schema := range s.DB {
		for _, table := range schema.Tables {
			r := Resource{
				Name:       table.Name,
				Table:      table,
				Properties: s.PropertiesFromDB(table.Columns),
				PK:         s.PropertyFromDB(table.GetPK()),
			}
			result[table.Path()] = r
		}
	}

	return result
}
