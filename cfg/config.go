package cfg

import (
	"fmt"
	"sort"

	"github.com/gofoji/foji/stringlist"
)

type (
	Error        string
	FileHandler  func(string) error // Used for post generation processing
	ParamMap     map[string]any     // Generic bucket of params passed to templates
	Processes    map[string]Process
	FileInputMap map[string]FileInput
	Output       map[string]stringlist.StringMap

	DBInput struct {
		// DB Connection string.
		// e.g. `host=localhost dbname=MyProject sslmode=disable`
		Connection string `yaml:"connection,omitempty"`
		// Regex for filtering the properties.
		// e.g. ".*\\.schema_migrations"
		Filter stringlist.Strings `yaml:"filter,omitempty"`
	}

	FileInput struct {
		// Files to process, supports glob syntax https://golang.org/pkg/path/filepath/#Match
		Files stringlist.Strings
		// Optional regex for filtering the files.  The filename must not match any of the filter
		// expressions to be considered valid.
		Filter  stringlist.Strings   `yaml:"filter,omitempty"`
		Rewrite stringlist.StringMap `yaml:"rewrite,omitempty"` // Optional rules for rewriting file names
	}

	// Maps are a set of lookups for mapping various attributes (type, name, case) for welds.
	Maps struct {
		Type     stringlist.StringMap `yaml:"type,omitempty"`     // Type maps (varchar -> string)
		Nullable stringlist.StringMap `yaml:"nullable,omitempty"` // Type maps for nullable types (varchar -> sql.NullString)
		Name     stringlist.StringMap `yaml:"name,omitempty"`     // Name maps (addr_l1 -> address_line_1)
		// special purpose case settings. Can be used in templates with "caseType" command.
		Case stringlist.StringMap `yaml:"case,omitempty"`
	}

	// Process encapsulates all data for executing a weld.
	Process struct {
		Output `yaml:",inline"`

		// Used to make a bundle of processes, if populated all other attributes are ignored
		Processes stringlist.Strings `yaml:"processes,omitempty,flow"`
		// ID of the process (used for bundle processes), populated by Processes.Merge
		ID string `yaml:"-"`
		// Output format, used to get defaults for naming, mapping, post processor
		Format string `yaml:"format,omitempty"`
		// Default case function (e.g snake, pascal, camel, kebab)
		Case string `yaml:"case,omitempty"`
		// Used for mapping data from Input to Output
		Maps Maps `yaml:"maps,omitempty"`
		// List of post-processing commands for each file generated (commonly used to invoke formatters like goimports)
		Post stringlist.Strings2D `yaml:"post,omitempty,flow"`
		// Custom parameters that can be passed into each template
		Params ParamMap `yaml:"params,omitempty"`
		// List of files to use as input
		Files FileInput `yaml:"files,omitempty"`
		// Root directory for outputs
		RootDir string `yaml:"rootDir,omitempty"`
		// ID of shared resources used for file input
		Resources stringlist.Strings `yaml:"resources,omitempty,flow"`
	}

	Config struct {
		// Database connection
		DB DBInput `yaml:"db,omitempty"`
		// List of file sets
		Files FileInputMap `yaml:"files,omitempty"`
		// Format Defaults (e.g. go, openapi, swift).
		Formats Processes `yaml:"formats,omitempty"`
		// Map of all processes
		Processes Processes `yaml:"processes,omitempty"`
	}
)

// Keys is a simple helper that returns the list of keys in the processes map.
func (pp Processes) Keys() stringlist.Strings {
	ss := make(stringlist.Strings, len(pp))
	i := 0

	for x := range pp {
		ss[i] = x
		i++
	}

	sort.Strings(ss)

	return ss
}

func (pp Processes) String() string {
	return pp.Keys().Sort().Join(",")
}

const (
	errProcess          = Error("Process")
	missingBundleFormat = "%w '%s' referenced by bundle `%s` not found.  Possible options: %s"
	missingFormat       = "%w '%s' not found.  Possible options: %s"
)

// Target Converts a list of process names into final list of processes.  Including de-referencing bundles.
func (pp Processes) Target(targets []string) ([]Process, error) {
	var result []Process

	for _, id := range targets {
		p, ok := pp[id]
		if !ok {
			return nil, fmt.Errorf(missingFormat, errProcess, id, pp)
		}

		if len(p.Processes) > 0 {
			for _, subID := range p.Processes {
				sub, ok := pp[subID]
				if !ok {
					return nil, fmt.Errorf(missingBundleFormat, errProcess, subID, id, pp)
				}

				result = append(result, sub)
			}
		} else {
			result = append(result, p)
		}
	}

	return result, nil
}

// HasString returns a string param identified by `name`, otherwise the second value (ok) is false.
func (pp ParamMap) HasString(name string) (string, bool) {
	p, ok := pp[name]
	if !ok {
		return "", false
	}

	s, ok := p.(string)

	return s, ok
}

// GetWithDefault returns a string param identified by `name`, otherwise returns the default.
func (pp ParamMap) GetWithDefault(name, def string) string {
	p, ok := pp[name]
	if !ok {
		return def
	}

	s, ok := p.(string)
	if !ok {
		return def
	}

	return s
}

// All merges all the mapped TypeMaps into a single StringMap, useful for getting the list of all templates.
func (o Output) All() stringlist.StringMap {
	result := stringlist.StringMap{}

	for _, m := range o {
		for k, v := range m {
			result[k] = v
		}
	}

	return result
}

// IsEmpty checks Files globs.
func (f FileInput) IsEmpty() bool {
	return len(f.Files) == 0
}

func (e Error) Error() string {
	return string(e)
}
