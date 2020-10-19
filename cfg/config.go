package cfg

import (
	"github.com/gofoji/foji/stringlist"
	"github.com/pkg/errors"
)

type FileHandler func(string) error

type ParamMap map[string]interface{} // Generic bucket of params passed to templates
type Processes map[string]Process

type DBInput struct {
	Connection string             `yaml:",omitempty"` // DB Connection string e.g. `host=localhost dbname=MyProject sslmode=disable`
	Filter     stringlist.Strings `yaml:",omitempty"` // Regex for filtering the properties
}

type FileInputMap map[string]FileInput

type FileInput struct {
	Files   stringlist.Strings   // Files to process, supports glob syntax https://golang.org/pkg/path/filepath/#Match
	Filter  stringlist.Strings   `yaml:",omitempty"` // Regex for filtering the files
	Rewrite stringlist.StringMap `yaml:",omitempty"` // Optional rules for rewriting names
}

func (f FileInput) IsEmpty() bool {
	return len(f.Files) == 0
}

type Output map[string]stringlist.StringMap

func MergeTypesMaps(maps ...stringlist.StringMap) stringlist.StringMap {
	result := stringlist.StringMap{}
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}

// All merges all the mapped TypeMaps into a single StringMap, useful for getting the list of all templates
func (o Output) All() stringlist.StringMap {
	result := stringlist.StringMap{}
	for _, m := range o {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}

type Maps struct {
	Type     stringlist.StringMap `yaml:",omitempty"` // Type maps (varchar -> string)
	Nullable stringlist.StringMap `yaml:",omitempty"` // Type maps for nullable types (varchar -> sql.NullString)
	Name     stringlist.StringMap `yaml:",omitempty"` // Name maps (addr_l1 -> address_line_1)
	Case     stringlist.StringMap `yaml:",omitempty"` // special purpose case settings.  Can be used in templates with "caseType" command
}

type Process struct {
	Processes stringlist.Strings   `yaml:",omitempty,flow"` // Used to make a bundle of processes, if populated all other attributes are ignored
	ID        string               `yaml:"-"`               // ID of the process (used for bundle processes), populated by Processes.Merge
	Format    string               `yaml:",omitempty"`      // Output format, used to get defaults for naming, mapping, post processor
	Case      string               `yaml:",omitempty"`      // Default case function (e.g snake, pascal, camel, kebab)
	Maps      Maps                 `yaml:",omitempty"`      // Used for mapping data from Input to Output
	Post      stringlist.Strings2D `yaml:",omitempty,flow"` // List of post processing commands for each file generated (commonly used to invoke formatters like goimports)
	Params    ParamMap             `yaml:",omitempty"`      // Custom parameters that can be passed into each template
	Files     FileInput            `yaml:",omitempty"`      // List of files to use as input
	RootDir   string               `yaml:",omitempty"`      // Root directory for outputs
	Resources stringlist.Strings   `yaml:",omitempty,flow"` // ID of shared resources used for file input
	Output    `yaml:",inline"`
}

type Config struct {
	DB    DBInput      `yaml:",omitempty"`
	Files FileInputMap `yaml:",omitempty"`

	// Format Defaults (e.g. go, openapi, swift)
	Formats Processes `yaml:",omitempty"`

	// Processes
	Processes Processes `yaml:",omitempty"`
}

func (pp Processes) Keys() stringlist.Strings {
	ss := make(stringlist.Strings, len(pp))
	i := 0
	for x := range pp {
		ss[i] = x
		i += 1
	}

	return ss
}

// Target Converts referenced bundles into final list of processes
func (pp Processes) Target(targets []string) (Processes, error) {
	result := Processes{}
	for _, id := range targets {
		p, ok := pp[id]
		if !ok {
			return nil, errors.Errorf("Process '%s' not found.  Possible options: %s", id, pp.Keys().Sort().Join(","))
		}
		if len(p.Processes) > 0 {
			for _, subID := range p.Processes {
				if _, contains := result[subID]; !contains {
					sub, ok := pp[subID]
					if !ok {
						return nil, errors.Errorf("Process '%s' referenced by bundle `%s` not found.  Possible options: %s", subID, id, pp.Keys().Sort().Join(","))
					}
					result[subID] = sub
				}
			}
		} else {
			result[id] = p
		}
	}
	return result, nil
}

func (pp ParamMap) HasString(name string) (string, bool) {
	p, ok := pp[name]
	if !ok {
		return "", false
	}
	s, ok := p.(string)
	return s, ok
}

func (pp ParamMap) HasStrings(name string) (stringlist.Strings, bool) {
	p, ok := pp[name]
	if !ok {
		return nil, false
	}
	ss, ok := p.(stringlist.Strings)
	return ss, ok
}
