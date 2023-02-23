package db

import (
	"errors"
	"fmt"
	"sort"

	"github.com/gofoji/foji/stringlist"
)

type DB map[string]*Schema

// Schema holds all the type definitions for a single schema.
type Schema struct {
	Name   string
	Tables Tables
	Enums  Enums
}

type Tables []*Table

// Table contains the definition of a table.
type Table struct {
	ID          uint32
	Name        string      // the original name of the table in the DB
	Type        string      // the table type (e.g. VIEW or BASE TABLE)
	Comment     string      // the comment attached to the table
	Schema      *Schema     // reference to schema `json:"-"`
	Columns     Columns     // ordered list of columns in this table
	ReadOnly    bool        // table allows insert/update/delete
	Indexes     Indexes     // list of indexes in this table
	ForeignKeys ForeignKeys // list of Foreign Keys from the table
	References  ForeignKeys // list of Foreign Keys to the table

	PrimaryKeys Columns // list of columns that are flagged as Primary Keys
}

type (
	Indexes     []*Index
	ForeignKeys []*ForeignKey
	Enums       []*Enum
	Columns     []*Column
)

// Index contains the definition of an index.
type Index struct {
	Name      string  // name of the index in the database
	IsUnique  bool    // true if the index is unique
	IsPrimary bool    // true if the index is for the PK
	Comment   string  // the comment attached to the index
	Columns   Columns `json:"-"` // list of columns in this index
}

// ForeignKey contains the definition of a foreign key.
type ForeignKey struct {
	Name           string  // the original name of the foreign key constraint in the db
	Comment        string  // the comment attached to the ForeignKey
	Columns        Columns `json:"-"` // Schema and Table can be retrieved from the source Column
	ForeignColumns Columns `json:"-"` //
}

// Column contains data about a column in a table.
type Column struct {
	Name         string      // the original name of the column in the DB
	Type         string      // the original type of the column in the DB
	Nullable     bool        // true if the column is not NON-NULL
	HasDefault   bool        // true if the column has a default
	IsPrimaryKey bool        // true if the column is a primary key
	Ordinal      int16       // the column's ordinal position
	Comment      string      // the comment attached to the column
	Table        *Table      `json:"-"`
	ForeignKey   *ForeignKey // foreign key database definition
}

// Enum represents a type that has a set of allowed values.
type Enum struct {
	ID      uint32
	Name    string             // the original name of the enum in the DB
	Values  stringlist.Strings // the list of possible values for this enum
	Comment string
	Schema  *Schema `json:"-"`
}

type Param struct {
	Ordinal       int
	Name          string
	Type          string
	SQLType       string
	QueryPosition int
}

type Params []*Param

func (ss DB) GetTable(schema, table string) (*Table, bool) {
	s, ok := ss[schema]
	if !ok {
		return nil, false
	}

	t, ok := s.GetTable(table)
	if !ok {
		return nil, false
	}

	return t, true
}

func (s Schema) GetTable(name string) (*Table, bool) {
	for _, t := range s.Tables {
		if t.Name == name {
			return t, true
		}
	}

	return nil, false
}

var ErrMissingColumn = errors.New("")

func (t Table) GetColumnsByName(names []string) ([]*Column, error) {
	result := make([]*Column, len(names))

	for i, name := range names {
		c := t.GetColumnByName(name)
		if c == nil {
			return nil, fmt.Errorf("%w%s", ErrMissingColumn, name)
		}

		result[i] = c
	}

	return result, nil
}

func (t Table) GetColumnByName(name string) *Column {
	for _, c := range t.Columns {
		if c.Name == name {
			return c
		}
	}

	return nil
}

func (t Table) GetPK() *Column {
	for _, c := range t.Columns {
		if c.IsPrimaryKey {
			return c
		}
	}

	return nil
}

func (t Table) HasPrimaryKey() bool {
	return len(t.PrimaryKeys) > 0
}

func (t Table) Path() string {
	if t.Schema != nil {
		return t.Schema.Name + "." + t.Name
	}

	return t.Name
}

func (cc Columns) Names() stringlist.Strings {
	ss := make(stringlist.Strings, len(cc))
	for x := range cc {
		ss[x] = cc[x].Name
	}

	return ss
}

func (cc Columns) Types() stringlist.Strings {
	ss := make(stringlist.Strings, len(cc))
	for x := range cc {
		ss[x] = cc[x].Type
	}

	return ss
}

func (cc Columns) ByOrdinal() Columns {
	r := cc.copy()
	sort.Sort(byOrdinal(cc))

	return r
}

type byOrdinal Columns

func (cc byOrdinal) Len() int {
	return len(cc)
}

func (cc byOrdinal) Less(i, j int) bool {
	return cc[i].Ordinal < cc[j].Ordinal
}

func (cc byOrdinal) Swap(i, j int) {
	cc[i], cc[j] = cc[j], cc[i]
}

func (cc Columns) copy() Columns {
	r := make(Columns, len(cc))
	copy(r, cc)

	return r
}

func (cc Columns) Paths() stringlist.Strings {
	ss := make(stringlist.Strings, len(cc))
	for x := range cc {
		ss[x] = cc[x].Path()
	}

	return ss
}

func (c Column) Path() string {
	if c.Table != nil {
		return c.Table.Path() + "." + c.Name
	}

	return c.Name
}

func (pp Params) ByName(name string) *Param {
	for _, p := range pp {
		if p.Name == name {
			return p
		}
	}

	return nil
}

func (i Index) Table() *Table {
	return i.Columns[0].Table
}

func (f ForeignKey) Table() *Table {
	return f.Columns[0].Table
}

func (f ForeignKey) ForeignTable() *Table {
	return f.ForeignColumns[0].Table
}

func (f ForeignKey) Path() string {
	if f.Table() != nil {
		return f.Table().Path() + "." + f.Name
	}

	return f.Name
}

func (e Enum) Path() string {
	if e.Schema != nil {
		return e.Schema.Name + "." + e.Name
	}

	return e.Name
}

func (i Index) Path() string {
	if i.Table() != nil {
		return i.Table().Path() + "." + i.Name
	}

	return i.Name
}
