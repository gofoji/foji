package sql

import (
	"sort"

	"github.com/gofoji/foji/stringlist"
)

type Params []*Param

type Param struct {
	Ordinal       int    // User Defined Order (generally used to order the params in the func call
	QueryPosition int    // Order in the query
	Name          string // Database field name
	Type          string // DB type
	TypeID        uint32 // DB type ID
	Nullable      bool   // True means the param is nullable
	Generated     bool   // Indicates type should be generated (locally defined)
	Query         *Query // The owning query
}

func (p Param) Path() string {
	if p.Query != nil {
		return p.Query.Name + "." + p.Name
	}

	return p.Name
}

func (pp Params) Names() stringlist.Strings {
	ss := make(stringlist.Strings, len(pp))

	for x := range pp {
		ss[x] = pp[x].Name
	}

	return ss
}

func (pp Params) ByOrdinal() Params {
	r := pp.copy()
	sort.Sort(byOrdinal(r))

	return r
}

type byOrdinal Params

func (cc byOrdinal) Len() int {
	return len(cc)
}

func (cc byOrdinal) Less(i, j int) bool {
	return cc[i].Ordinal < cc[j].Ordinal
}

func (cc byOrdinal) Swap(i, j int) {
	cc[i], cc[j] = cc[j], cc[i]
}

func (pp Params) copy() Params {
	r := make(Params, len(pp))

	for i, p := range pp {
		r[i] = p
	}

	return r
}

func (pp Params) ByQuery() Params {
	r := pp.copy()
	sort.Sort(byQuery(r))

	return r
}

type byQuery Params

func (cc byQuery) Len() int {
	return len(cc)
}

func (cc byQuery) Less(i, j int) bool {
	return cc[i].QueryPosition < cc[j].QueryPosition
}

func (cc byQuery) Swap(i, j int) {
	cc[i], cc[j] = cc[j], cc[i]
}

func (pp Params) ByName(name string) *Param {
	for _, p := range pp {
		if p.Name == name {
			return p
		}
	}

	return nil
}
