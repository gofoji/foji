package db

import (
	"github.com/gofoji/foji/stringlist"
)

// Filter applies the regexes specified and filter out all matching schema objects.
func (ss DB) Filter(filters stringlist.Strings) DB {
	if len(filters) == 0 {
		return ss
	}

	result := DB{}

	for key, s := range ss {
		if filters.AnyMatches(s.Name) {
			continue
		}

		filtered := s.Filter(filters)
		result[key] = &filtered
	}

	return result
}

func (s Schema) Filter(filters stringlist.Strings) Schema {
	return Schema{
		Name:   s.Name,
		Tables: s.Tables.Filter(filters),
		Enums:  s.Enums.Filter(filters),
	}
}

// Filter applies the regexes specified and filter out all matching schema objects.
func (tt Tables) Filter(filters stringlist.Strings) Tables {
	if len(filters) == 0 {
		return tt
	}

	result := Tables{}

	for _, t := range tt {
		if !filters.AnyMatches(t.Path()) {
			filtered := t.Filter(filters)
			result = append(result, &filtered)
		}
	}

	return result
}

func (t Table) Filter(filters stringlist.Strings) Table {
	return Table{
		ID:          t.ID,
		Name:        t.Name,
		Type:        t.Type,
		Comment:     t.Comment,
		Schema:      t.Schema,
		Columns:     t.Columns.Filter(filters),
		ReadOnly:    t.ReadOnly,
		Indexes:     t.Indexes.Filter(filters),
		ForeignKeys: t.ForeignKeys.Filter(filters),
		References:  t.References.Filter(filters),
		PrimaryKeys: t.PrimaryKeys.Filter(filters),
	}
}

// Filter applies the regexes specified and filter out all matching schema objects.
func (ee Enums) Filter(filters stringlist.Strings) Enums {
	if len(filters) == 0 {
		return ee
	}

	result := Enums{}

	for _, e := range ee {
		if !filters.AnyMatches(e.Path()) {
			filtered := e.Filter(filters)
			result = append(result, &filtered)
		}
	}

	return result
}

// Filter applies the regexes specified and filter out all matching schema objects.
func (cc Columns) Filter(filters stringlist.Strings) Columns {
	if len(filters) == 0 {
		return cc
	}

	result := Columns{}

	for _, c := range cc {
		if !filters.AnyMatches(c.Path()) {
			result = append(result, c)
		}
	}

	return result
}

// Filter applies the regexes specified and filter out all matching schema objects.
func (ii Indexes) Filter(filters stringlist.Strings) Indexes {
	if len(filters) == 0 {
		return ii
	}

	result := Indexes{}

	for _, i := range ii {
		if !filters.AnyMatches(i.Path()) {
			result = append(result, i)
		}
	}

	return result
}

// Filter applies the regexes specified and filter out all matching schema objects.
func (ff ForeignKeys) Filter(filters stringlist.Strings) ForeignKeys {
	if len(filters) == 0 {
		return ff
	}

	result := ForeignKeys{}

	for _, f := range ff {
		if !filters.AnyMatches(f.Path()) {
			result = append(result, f)
		}
	}

	return result
}

func (e Enum) Filter(filters stringlist.Strings) Enum {
	return Enum{
		ID:      e.ID,
		Name:    e.Name,
		Values:  e.Values.Filters(filters),
		Comment: e.Comment,
		Schema:  e.Schema,
	}
}
