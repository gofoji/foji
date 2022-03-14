package cfg

import (
	"fmt"
	"sort"

	"github.com/gofoji/foji/stringlist"
)

type (
	Error        string
	FileHandler  func(string) error     // Used for post generation processing
	ParamMap     map[string]interface{} // Generic bucket of params passed to templates
	Processes    map[string]Process
	FileInputMap map[string]FileInput
	Output       map[string]stringlist.StringMap

	DBInput struct {
		// DB Connection string.
		// e.g. `host=localhost dbname=MyProject sslmode=disable`
		Connection string `yaml:",omitempty"`
		// Regex for filtering the properties.
		// e.g. ".*\\.schema_migrations"
		Filter stringlist.Strings `yaml:",omitempty"`
	}

	FileInput struct {
		// Files to process, supports glob syntax https://golang.org/pkg/path/filepath/#Match
		Files stringlist.Strings
		// Optional regex for filtering the files.  The filename must not match any of the filter
		// expressions to be considered valid.
		Filter  stringlist.Strings   `yaml:",omitempty"`
		Rewrite stringlist.StringMap `yaml:",omitempty"` // Optional rules for rewriting file names
	}

	// Maps are a set of lookups for mapping various attributes (type, name, case) for welds.
	Maps struct {
		Type     stringlist.StringMap `yaml:",omitempty"` // Type maps (varchar -> string)
		Nullable stringlist.StringMap `yaml:",omitempty"` // Type maps for nullable types (varchar -> sql.NullString)
		Name     stringlist.StringMap `yaml:",omitempty"` // Name maps (addr_l1 -> address_line_1)
		// special purpose case settings. Can be used in templates with "caseType" command.
		Case stringlist.StringMap `yaml:",omitempty"`
	}

	// Process encapsulates all data for executing a weld.
	Process struct {
		// Used to make a bundle of processes, if populated all other attributes are ignored
		Processes stringlist.Strings `yaml:",omitempty,flow"`
		// ID of the process (used for bundle processes), populated by Processes.Merge
		ID string `yaml:"-"`
		// Output format, used to get defaults for naming, mapping, post processor
		Format string `yaml:",omitempty"`
		// Default case function (e.g snake, pascal, camel, kebab)
		Case string `yaml:",omitempty"`
		// Used for mapping data from Input to Output
		Maps Maps `yaml:",omitempty"`
		// List of post-processing commands for each file generated (commonly used to invoke formatters like goimports)
		Post stringlist.Strings2D `yaml:",omitempty,flow"`
		// Custom parameters that can be passed into each template
		Params ParamMap `yaml:",omitempty"`
		// List of files to use as input
		Files FileInput `yaml:",omitempty"`
		// Root directory for outputs
		RootDir string `yaml:",omitempty"`
		// ID of shared resources used for file input
		Resources stringlist.Strings `yaml:",omitempty,flow"`
		Output    `yaml:",inline"`
	}

	Config struct {
		// Database connection
		DB DBInput `yaml:",omitempty"`
		// List of file sets
		Files FileInputMap `yaml:",omitempty"`
		// Format Defaults (e.g. go, openapi, swift).
		Formats Processes `yaml:",omitempty"`
		// Map of all processes
		Processes Processes `yaml:",omitempty"`
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
