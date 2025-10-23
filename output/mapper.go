package output

import (
	"fmt"
	"strings"

	"github.com/gofoji/foji/cfg"
)

// ResolveType resolves a type to a Go type using the configured type maps.
// It checks in the following order:
// 1. Path-based type map (e.g., ".table.column" -> "string")
// 2. Nullable type map (if column/param is nullable)
// 3. Generic type map (e.g., "varchar" -> "string")
// Returns an UNKNOWN placeholder if no mapping is found.
func ResolveType(maps cfg.Maps, checkFunc func(string) string, columnType string, nullable bool, path string) string {
	// Check path-based mappings first
	pp := strings.Split(path, ".")
	for i := range pp {
		p := strings.Join(pp[i:], ".")
		t, ok := maps.Type["."+p]
		if ok {
			return checkFunc(t)
		}
	}

	// Check nullable type mapping
	if nullable {
		t, ok := maps.Nullable[columnType]
		if ok {
			return checkFunc(t)
		}
	}

	// Check standard type mapping
	t, ok := maps.Type[columnType]
	if ok {
		return checkFunc(t)
	}

	// Check for qualified names (containing . or /)
	if strings.ContainsAny(columnType, "./") {
		return checkFunc(columnType)
	}

	return fmt.Sprintf("UNKNOWN:path(%s):type(%s)", path, columnType)
}
