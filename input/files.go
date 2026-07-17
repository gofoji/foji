package input

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"

	"github.com/rs/zerolog"

	"github.com/gofoji/foji/cfg"
	"github.com/gofoji/foji/files"
	"github.com/gofoji/foji/stringlist"
)

var (
	ErrNoGithubToken    = errors.New("missing github authentication token")
	ErrGithubStatusCode = errors.New("github returned non-successful status")
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

func Parse(ctx context.Context, logger zerolog.Logger, input cfg.FileInput) (FileGroup, error) {
	result := FileGroup{FileInput: input}

	fileResults, err := ParseFiles(ctx, logger, input)
	if err != nil {
		return result, err
	}

	urlResults, err := ParseUrls(ctx, logger, input)
	if err != nil {
		return result, err
	}

	githubResults, err := ParseGithubFiles(ctx, logger, input)
	if err != nil {
		return result, err
	}

	result.Files = append(result.Files, fileResults...)
	result.Files = append(result.Files, urlResults...)
	result.Files = append(result.Files, githubResults...)

	return result, nil
}

func ParseFiles(_ context.Context, logger zerolog.Logger, input cfg.FileInput) ([]File, error) {
	result := []File{}

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

			b, err := os.ReadFile(filename)
			if err != nil {
				return result, fmt.Errorf("error reading file: %s: %w", filename, err)
			}

			file := File{
				Source:  filename,
				Name:    rewrite(input.Rewrite, filename),
				Content: b,
			}
			logger.Debug().Str("name", file.Name).Msg("File Loaded")
			result = append(result, file)
			loadedFiles = append(loadedFiles, filename)
		}
	}

	return result, nil
}

func ParseGithubFiles(ctx context.Context, logger zerolog.Logger, input cfg.FileInput) ([]File, error) {
	result := []File{}

	githubToken, ok := os.LookupEnv("GH_TOKEN")

	if !ok && len(input.GithubFiles) > 0 {
		return result, ErrNoGithubToken
	}

	for _, u := range input.GithubFiles {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
		if err != nil {
			return result, fmt.Errorf("error creating request to read url: %s: %w", u, err)
		}

		req.Header.Add("Accept", "application/vnd.github.raw+json")
		req.Header.Add("Authorization", "Bearer "+githubToken)

		logger.Debug().Str("source", u).Msg("Fetching URL")

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return result, fmt.Errorf("error fetching url: %s: %w", u, err)
		}

		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			return result, fmt.Errorf("%w %d for url %s", ErrGithubStatusCode, res.StatusCode, u)
		}

		logger.Debug().Str("source", u).Msg("Reading URL")

		b, err := io.ReadAll(res.Body)
		if err != nil {
			return result, fmt.Errorf("error reading body: %s: %w", u, err)
		}

		file := File{
			Source:  u,
			Name:    rewrite(input.Rewrite, u),
			Content: b,
		}
		logger.Debug().Str("name", file.Name).Msg("File Loaded")
		result = append(result, file)
	}

	return result, nil
}

func ParseUrls(ctx context.Context, logger zerolog.Logger, input cfg.FileInput) ([]File, error) {
	result := []File{}

	for _, u := range input.Urls {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
		if err != nil {
			return result, fmt.Errorf("error creating request to read url: %s: %w", u, err)
		}

		logger.Debug().Str("source", u).Msg("Fetching URL")

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return result, fmt.Errorf("error fetching url: %s: %w", u, err)
		}

		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			return result, fmt.Errorf("%w %d for url %s", ErrGithubStatusCode, res.StatusCode, u)
		}

		logger.Debug().Str("source", u).Msg("Reading URL")

		b, err := io.ReadAll(res.Body)
		if err != nil {
			return result, fmt.Errorf("error reading body: %s: %w", u, err)
		}

		file := File{
			Source:  u,
			Name:    rewrite(input.Rewrite, u),
			Content: b,
		}
		logger.Debug().Str("name", file.Name).Msg("File Loaded")
		result = append(result, file)
	}

	return result, nil
}
