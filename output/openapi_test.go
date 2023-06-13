package output

import (
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gofoji/foji/cfg"
	"github.com/gofoji/foji/embed"
	"github.com/gofoji/foji/input/openapi"
	"github.com/gofoji/foji/stringlist"
)

func TestGetType(t *testing.T) {
	doc, err := openapi3.NewLoader().LoadFromFile("testdata/openapi.yaml")
	require.NoError(t, err)

	ctx := getContext(doc)

	for _, testcase := range []struct {
		name, pkg, want string
	}{
		{"does.not.exist", "", ""},
		{"Nest", "foreign", "local.Nest"},
		{"Main", "local", "Main"},
		{"Main.total", "foreign", "int32"},
		{"Nest.id", "foreign", "int64"},
		{"Nest.label", "foreign", "*string"},
		{"Main.list", "foreign", "[]local.Nest"},
		{"Main.override_name", "foreign", "typeOverride"},
		{"EmptyObject", "local", "DefaultObject"},
		{"Nada", "foreign", "any"},
		{"EmptyAlias", "local", "myOverride"},
		{"Foo.bad", "local", "INVALID x-go-type: map[foo:bar]"},
		{"Foo", "local", "Foo"},
		{"Complex.status", "local", "ComplexStatusEnum"},
		{"Complex.main", "local", "Main"},
		{"Complex.nests", "local", "[]Nest"},
		{"Complex", "local", "Complex"},
		{"MapTest", "local", "*uuid.UUID"},
		{"PtrToString", "local", "*foo.Bar"},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			got := ctx.GetType(testcase.pkg, testcase.name, evalPath(doc, testcase.name))
			assert.Equal(t, testcase.want, got)
		})
	}
}

func getContext(doc *openapi3.T) OpenAPIFileContext {
	defaults, err := cfg.LoadYaml(embed.FojiDotYaml)
	if err != nil {
		panic(err)
	}

	config := cfg.Config{Processes: cfg.Processes{"openAPI": cfg.Process{Maps: cfg.Maps{
		Type: stringlist.StringMap{"Main.override_name": "typeOverride", "EmptyAlias": "myOverride", "object": "DefaultObject"},
	}, Params: cfg.ParamMap{"Package": "local"}}}}
	config = config.Merge(defaults)

	ctx := OpenAPIFileContext{
		Context: Context{
			Process: config.Processes["openAPI"],
		},
		File: openapi.File{
			API: doc,
		},
	}

	return ctx
}

func getSchema(doc *openapi3.T, name string) *openapi3.SchemaRef {
	for key, value := range doc.Components.Schemas {
		if key == name {
			return value
		}
	}

	return nil
}

func getProperty(schema *openapi3.SchemaRef, name string) *openapi3.SchemaRef {
	if schema == nil {
		return nil
	}

	for key, value := range schema.Value.Properties {
		if key == name {
			return value
		}
	}

	for _, s := range schema.Value.AllOf {
		value := getProperty(s, name)
		if value != nil {
			return value
		}
	}

	return nil
}

func evalPath(doc *openapi3.T, path string) *openapi3.SchemaRef {
	pathStep := strings.Split(path, ".")

	i := 1
	s := getSchema(doc, pathStep[0])
	for i < len(pathStep) {
		s = getProperty(s, pathStep[i])
		i++
	}

	return s
}
