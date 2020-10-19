package output

import (
	"fmt"
	"strings"

	"github.com/codemodus/kace"
	"github.com/gofoji/foji/cfg"
	"github.com/gofoji/foji/input/db"
	"github.com/gofoji/foji/stringlist"
	"github.com/sirupsen/logrus"
)

const (
	DbAll    = "DbAll"
	DbSchema = "DbSchema"
	DbTable  = "DbTable"
	DbEnums  = "DbEnums"
	DbEnum   = "DbEnum"
)

func HasDBOutput(o cfg.Output) bool {
	return hasAnyOutput(o, DbAll, DbSchema, DbTable, DbEnum, DbEnums)
}

func DB(p cfg.Process, fn cfg.FileHandler, logger logrus.FieldLogger, schemas db.DB, simulate bool) error {
	ctx := SchemasContext{
		Context: Context{Process:p, Logger: logger},
		DB:      schemas,
	}

	err := invokeProcess(p.Output[DbAll], p.RootDir, fn, logger, &ctx, simulate)
	if err != nil {
		return err
	}

	for _, s := range schemas {
		schemaCtx := SchemaContext{
			Schema:         *s,
			SchemasContext: ctx,
		}
		err := invokeProcess(p.Output[DbSchema], p.RootDir, fn, logger, &schemaCtx, simulate)
		if err != nil {
			return err
		}
		for _, t := range s.Tables {

			tableCtx := TableContext{
				Table:          *t,
				SchemasContext: ctx,
			}
			err := invokeProcess(p.Output[DbTable], p.RootDir, fn, logger, &tableCtx, simulate)
			if err != nil {
				return err
			}
		}

		enumsCtx := EnumsContext{
			Enums:          s.Enums,
			SchemasContext: ctx,
		}
		err = invokeProcess(p.Output[DbEnums], p.RootDir, fn, logger, &enumsCtx, simulate)
		if err != nil {
			return err
		}

		for _, e := range s.Enums {
			enumCtx := EnumContext{
				Enum:           *e,
				SchemasContext: ctx,
			}
			err := invokeProcess(p.Output[DbEnum], p.RootDir, fn, logger, &enumCtx, simulate)
			if err != nil {
				return err
			}
		}

	}

	return nil
}

type SchemasContext struct {
	Context
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

func (s *SchemasContext) Parameterize(cc db.Columns, format, pkg string) string {
	ss := make(stringlist.Strings, len(cc))
	for x := range cc {
		ss[x] = fmt.Sprintf(format, kace.Camel(cc[x].Name), s.GetType(cc[x], pkg))
	}

	return strings.Join(ss, ", ")
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

func stripPackage(typ, pkg string) string {
	if pkg != "" && strings.HasPrefix(typ, pkg) {
		return typ[len(pkg)+1:]
	}
	return typ
}
func (s SchemasContext) GetType(c *db.Column, pkg string) string {
	pp := strings.Split(c.Path(), ".")
	for i := range pp {
		p := strings.Join(pp[i:], ".")
		t, ok := s.Maps.Type["."+p]
		if ok {
			return stripPackage(t, pkg)
		}
	}

	if c.Nullable {
		t, ok := s.Maps.Nullable[c.Type]
		if ok {
			return stripPackage(t, pkg)
		}
	}
	t, ok := s.Maps.Type[c.Type]
	if ok {
		return stripPackage(t, pkg)
	}

	return fmt.Sprintf("UNKNOWN:path(%s):type(%s)", c.Path(), c.Type)
}

func (s SchemasContext) PropertyFromDB(c *db.Column) *Property {
	if c == nil {
		return nil
	}

	t := s.GetType(c, "")
	format := ""
	if strings.Contains(t, ",") {
		tt := strings.Split(t, ",")
		if len(tt) == 2 {
			t = tt[0]
			format = tt[1]
		} else {
			s.Logger.Errorf("Schema Column: %s, Invalid Type declaration:%s", c.Path(), t)
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
