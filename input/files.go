package input

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"

	"github.com/gofoji/foji/cfg"
	"github.com/gofoji/foji/files"
	"github.com/gofoji/foji/stringlist"
	"github.com/rs/zerolog"
)

type FileGroup struct {
	cfg.FileInput
	Files []File
}

type File struct {
	Source  string // Original filename
	Name    string // Name after any rewrite conversions
	Content []byte // Contents of file
}

func rewrite(rules stringlist.StringMap, name string) string {
	for match, replace := range rules {
		re := regexp.MustCompile(match)
		if re.MatchString(name) {
			return re.ReplaceAllString(name, replace)
		}
	}

	return name
}

func Parse(_ context.Context, logger zerolog.Logger, input cfg.FileInput) (FileGroup, error) {
	result := FileGroup{FileInput: input}

	loadedFiles := stringlist.Strings{}

	for _, glob := range input.Files {
		logger.Debug().Str("source", glob).Msg("Searching Glob")

		matches, err := files.Glob(glob)
		if err != nil {
			return result, fmt.Errorf("error processing glob: %s: %w", glob, err)
		}

		if len(matches) == 0 {
			logger.Warn().Str("glob", glob).Msg("No matches found")
		}

		for _, filename := range matches {
			// Guard redundant glob patterns
			if loadedFiles.Contains(filename) {
				continue
			}

			if input.Filter.AnyMatches(filename) {
				logger.Debug().Str("file", filename).Msg("Filtering File")

				continue
			}

			fileInfo, err := os.Stat(filename)
			if err != nil {
				return result, fmt.Errorf("error reading file: %s: %w", filename, err)
			}

			if fileInfo.IsDir() {
				continue
			}

			logger.Debug().Str("source", filename).Msg("Reading File")

			b, err := ioutil.ReadFile(filename)
			if err != nil {
				return result, fmt.Errorf("error reading file: %s: %w", filename, err)
			}

			file := File{
				Source:  filename,
				Name:    rewrite(input.Rewrite, filename),
				Content: b,
			}
			logger.Debug().Str("name", file.Name).Msg("File Loaded")
			result.Files = append(result.Files, file)
			loadedFiles = append(loadedFiles, filename)
		}
	}

	return result, nil
}
