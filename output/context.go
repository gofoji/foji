package output

import (
	"fmt"
	"strings"

	"github.com/gofoji/foji/cfg"
	"github.com/gofoji/foji/runtime"
	"github.com/gofoji/foji/stringlist"
	"github.com/gofoji/plates"
	"github.com/rs/zerolog"
)

type RuntimeParams map[string]interface{}

type Context struct {
	// RuntimeParams are used for parameterized sub-templates
	RuntimeParams
	// Process provides the details of the currently executing Process
	cfg.Process
	// Logger provides logging features to the context helpers and templates
	Logger zerolog.Logger
	// AbortError allows cancelling saving of a file.  See NotNeededIf.
	AbortError error
}

// Funcs defaults the default case funcs based on the Process.Case.
func (c *Context) Funcs() plates.FuncMap {
	return runtime.CaseFuncs(c.Case)
}

// Aborted is used to control file generation based on template execution.  See NotNeededIf.
func (c *Context) Aborted() error {
	return c.AbortError
}

// NotNeededIf given bool is true the execution is aborted, and can be used to prevent generation of a file.
func (c *Context) NotNeededIf(t bool, reason string) (string, error) {
	if t {
		c.AbortError = fmt.Errorf("%w: %s", ErrNotNeeded, reason)

		return "", c.AbortError
	}

	return "", nil
}

// ErrorIf if given bool is true the execution is fatally aborted, and stops processing.
func (c *Context) ErrorIf(t bool, reason string) (string, error) {
	if t {
		c.AbortError = fmt.Errorf("%w: %s", ErrMissingRequirement, reason)

		return "", c.AbortError
	}

	return "", nil
}

const (
	ErrInvalidDictParams = Error("invalid dict params in call to WithParams, must be key and value pairs")
	ErrInvalidDictKey    = Error("invalid dict params in call to WithParams, must be key and value pairs")
)

// WithParams Clones the current context and adds runtime params for each pair of key, value provided.
// Used for executing sub templates that still need access to the context.
func (c *Context) WithParams(values ...interface{}) (*Context, error) {
	if len(values)%2 != 0 {
		return nil, ErrInvalidDictParams
	}

	out := *c
	out.RuntimeParams = make(map[string]interface{}, len(values)/2) //nolint:gomnd

	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok {
			return nil, ErrInvalidDictKey
		}

		out.RuntimeParams[key] = values[i+1]
	}

	return &out, nil
}

// ToCase uses the current default case to map the current string.
func (c *Context) ToCase(name string) string {
	fn := runtime.Case(c.Case)

	mapper, ok := fn.(stringlist.StringMapper)
	if ok {
		return mapper(name)
	}

	return name
}

// PackageName is a helper to extract the package name from a fully qualified package.
// It uses the Process.Params.Package as the source.
// Params.Package "github.com/domain/repo/package/subpackage" => "subpackage".
func (c *Context) PackageName() string {
	pkg, _ := c.Params.HasString("Package")
	pp := strings.Split(pkg, "/")

	return pp[len(pp)-1]
}

// Imports tracks dynamic usage of objects.  Because templates are executed in order, using this to populate a list
// at the top of a generated file requires precalculating all of the imports.  See SQLContext.Init as an example.
// Another option would be to create a buffer of generated code at the beginning, then generate the final output.
type Imports stringlist.Strings

// CheckPackage is used for type mapping.  Currently it is designed for go fully qualified package names.
// Examples:
// "github.com/domain/repo/package/subpackage.Type", "" => "subpackage.Type"
// "time.Time", "" => "time.Time"
// "int", "" => "int"
// "github.com/domain/repo/package/subpackage.Type", "github.com/domain/repo/package/subpackage" => "Type"
// If the type is defined in a separate package the package is added to the import list.
func (ii *Imports) CheckPackage(t, pkg string) string {
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
	ii.Add(typePkg)

	pp := strings.Split(t, "/")

	return pp[len(pp)-1]
}

// Add filters duplicates and appends to the import list.
// Add works on uninitialized Imports objects.
func (ii *Imports) Add(s string) {
	for _, i := range *ii {
		if i == s {
			return
		}
	}

	*ii = append(*ii, s)
}
