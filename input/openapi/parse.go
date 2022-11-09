package openapi

import (
	"context"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gofoji/foji/input"
	"github.com/rs/zerolog"
)

type FileGroups []FileGroup

type FileGroup []File

type File struct {
	Input input.File
	API   *openapi3.T
}

func Parse(_ context.Context, logger zerolog.Logger, inGroups []input.FileGroup) (FileGroups, error) {
	result := make(FileGroups, len(inGroups))

	for i, ff := range inGroups {
		var group FileGroup

		for _, f := range ff.Files {
			logger.Info().Msgf("Parsing swagger from: %s", f.Source)

			loader := openapi3.NewLoader()
			loader.IsExternalRefsAllowed = true

			swagger, err := loader.LoadFromData(f.Content)
			if err != nil {
				panic(err)
			}

			d := File{Input: f, API: swagger}

			group = append(group, d)
		}

		result[i] = group
	}

	return result, nil
}
