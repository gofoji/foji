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
	var result FileGroups

	for _, ff := range inGroups {
		var group FileGroup
		for _, f := range ff.Files {
			logger.Infof("Parsing swagger from: %s", f.Source)
			swagger, err := openapi3.NewLoader().LoadFromData([]byte(f.Content))
			if err != nil {
				panic(err)
			}

			d := File{Input: f, API: swagger}

			group = append(group, d)
		}
		result = append(result, group)
	}

	return result, nil
}
