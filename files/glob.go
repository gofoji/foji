package files

import (
	"os"
	"path/filepath"
	"strings"
)

// Glob adds double-star support to the core path/filepath Glob function.
// It's useful when your globs might have double-stars, but you're not sure.
func Glob(pattern string) ([]string, error) {
	if !strings.Contains(pattern, "**") {
		// pass-thru to core package if no double-star
		return filepath.Glob(pattern)
	}

	return Expand(strings.Split(pattern, "**"))
}

// Expand finds matches for the provided Globs.
func Expand(globs []string) ([]string, error) {
	if len(globs) == 0 {
		return nil, nil
	}

	matches := []string{""}

	for _, glob := range globs {
		var hits []string

		hitMap := map[string]bool{}

		for _, match := range matches {
			if match == "" {
				match = "*"

				hits = append(hits, ".")
			}

			paths, err := filepath.Glob(match + glob)
			if err != nil {
				return nil, err
			}

			for _, path := range paths {
				err = filepath.Walk(path, func(path string, _ os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					// save de-duped match from current iteration
					if _, ok := hitMap[path]; !ok {
						hits = append(hits, path)
						hitMap[path] = true
					}

					return nil
				})
				if err != nil {
					return nil, err
				}
			}
		}

		matches = hits
	}

	return matches, nil
}
