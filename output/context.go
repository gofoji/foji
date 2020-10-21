package output

import (
	"errors"
	"strings"
	"text/template"

	"github.com/gofoji/foji/cfg"
	"github.com/gofoji/foji/runtime"
	"github.com/gofoji/foji/stringlist"
	"github.com/sirupsen/logrus"
)

type RuntimeParams map[string]interface{}

type Context struct {
	RuntimeParams
	cfg.Process
	Logger     logrus.FieldLogger
	Imports    Imports
	AbortError error
}

func (c *Context) Funcs() template.FuncMap {
	return runtime.CaseFuncs(c.Case)
}

func (c *Context) Aborted() error {
	return c.AbortError
}

func (c *Context) WithParams(values ...interface{}) (*Context, error) {
	out := *c

	if len(values)%2 != 0 {
		return nil, errors.New("invalid dict call")
	}

	out.RuntimeParams = make(map[string]interface{}, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok {
			return nil, errors.New("dict keys must be strings")
		}

		out.RuntimeParams[key] = values[i+1]
	}

	return &out, nil
}

func (c *Context) ToCase(name string) string {
	fn := runtime.Case(c.Case)
	mapper, ok := fn.(stringlist.StringMapper)
	if ok {
		return mapper(name)
	}
	return name
}

func (c *Context) PackageName() string {
	pkg, _ := c.Params.HasString("Package")
	pp := strings.Split(pkg, "/")
	return pp[len(pp)-1]
}

// CheckPackage is used for go type mapping.
// converts "github.com/domain/repo/package/subpackage.Type" to "subpackage.Type"
func (c *Context) CheckPackage(t, pkg string) string {
	tt := strings.Split(t, ".")
	// Base Type
	if len(tt) == 1 {
		return t
	}

	typePkg := strings.Join(tt[0:len(tt)-1], ".")
	// Type defined in same package
	if typePkg == pkg {
		return tt[len(tt)-1]
	}

	// Type defined in external package
	c.Imports.Add(typePkg)
	pp := strings.Split(t, "/")
	return pp[len(pp)-1]
}

type Imports stringlist.Strings

func (ii *Imports) Add(s string) {
	for _, i := range *ii {
		if i == s {
			return
		}
	}
	*ii = append(*ii, s)
}
