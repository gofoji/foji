package openapi

import (
	"context"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gofoji/foji/input"
	"github.com/sirupsen/logrus"
)

type FileGroups []FileGroup

type FileGroup []File

type File struct {
	Input input.File
	API   *openapi3.T
}

func Parse(ctx context.Context, logger logrus.FieldLogger, inGroups []input.FileGroup) (FileGroups, error) {
	result := make(FileGroups, len(inGroups))

	for i, ff := range inGroups {
		var group FileGroup

		for _, f := range ff.Files {
			logger.Infof("Parsing swagger from: %s", f.Source)

			swagger, err := openapi3.NewLoader().LoadFromData(f.Content)
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
