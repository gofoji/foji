package input

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	"github.com/gofoji/foji/cfg"
	"github.com/gofoji/foji/stringlist"
	"github.com/sirupsen/logrus"
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

func Parse(_ context.Context, logger logrus.FieldLogger, input cfg.FileInput) (FileGroup, error) {
	result := FileGroup{FileInput: input}

	loadedFiles := stringlist.Strings{}

	for _, glob := range input.Files {
		logger.WithField("source", glob).Debug("Searching Glob")

		files, err := filepath.Glob(glob)
		if err != nil {
			return result, fmt.Errorf("error processing glob: %s: %w", glob, err)
		}

		for _, filename := range files {
			// Guard redundant glob patterns
			if loadedFiles.Contains(filename) {
				continue
			}

			if input.Filter.AnyMatches(filename) {
				logger.WithField("file", filename).Debug("Filtering File")

				continue
			}

			fileInfo, err := os.Stat(filename)
			if err != nil {
				return result, fmt.Errorf("error reading file: %s: %w", filename, err)
			}

			if fileInfo.IsDir() {
				continue
			}

			logger.WithField("source", filename).Debug("Reading File")

			b, err := ioutil.ReadFile(filename)
			if err != nil {
				return result, fmt.Errorf("error reading file: %s: %w", filename, err)
			}

			file := File{
				Source:  filename,
				Name:    rewrite(input.Rewrite, filename),
				Content: b,
			}
			logger.WithField("name", file.Name).Debug("File Loaded")
			result.Files = append(result.Files, file)
			loadedFiles = append(loadedFiles, filename)
		}
	}

	return result, nil
}
