package openapi

import (
	"context"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gofoji/foji/input"
)

func TestParse_Success(t *testing.T) {
	spec := []byte(`openapi: "3.0.2"
info:
  title: Test
  version: "1.0"
paths:
  /test:
    get:
      operationId: testOp
      responses:
        "200":
          description: OK
components:
  schemas:
    Foo:
      type: object
      properties:
        name:
          type: string
`)

	inGroups := []input.FileGroup{
		{
			Files: []input.File{
				{Source: "test.yaml", Name: "test.yaml", Content: spec},
			},
		},
	}

	result, err := Parse(context.Background(), zerolog.Nop(), inGroups)
	require.NoError(t, err)
	require.Len(t, result, 1)
	require.Len(t, result[0], 1)

	f := result[0][0]
	assert.Equal(t, "test.yaml", f.Input.Source)
	assert.NotNil(t, f.API)
	assert.Equal(t, "Test", f.API.Info.Title)
	assert.NotNil(t, f.API.Paths.Find("/test"))
	assert.Contains(t, f.API.Components.Schemas, "Foo")
}

func TestParse_MultipleFiles(t *testing.T) {
	spec1 := []byte(`openapi: "3.0.2"
info:
  title: Spec1
  version: "1.0"
paths: {}
`)
	spec2 := []byte(`openapi: "3.0.2"
info:
  title: Spec2
  version: "2.0"
paths: {}
`)

	inGroups := []input.FileGroup{
		{
			Files: []input.File{
				{Source: "a.yaml", Name: "a.yaml", Content: spec1},
				{Source: "b.yaml", Name: "b.yaml", Content: spec2},
			},
		},
	}

	result, err := Parse(context.Background(), zerolog.Nop(), inGroups)
	require.NoError(t, err)
	require.Len(t, result, 1)
	require.Len(t, result[0], 2)
	assert.Equal(t, "Spec1", result[0][0].API.Info.Title)
	assert.Equal(t, "Spec2", result[0][1].API.Info.Title)
}

func TestParse_MultipleGroups(t *testing.T) {
	spec := []byte(`openapi: "3.0.2"
info:
  title: Test
  version: "1.0"
paths: {}
`)

	inGroups := []input.FileGroup{
		{Files: []input.File{{Source: "a.yaml", Content: spec}}},
		{Files: []input.File{{Source: "b.yaml", Content: spec}}},
	}

	result, err := Parse(context.Background(), zerolog.Nop(), inGroups)
	require.NoError(t, err)
	require.Len(t, result, 2)
	require.Len(t, result[0], 1)
	require.Len(t, result[1], 1)
}

func TestParse_EmptyGroups(t *testing.T) {
	result, err := Parse(context.Background(), zerolog.Nop(), nil)
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestParse_EmptyFileGroup(t *testing.T) {
	inGroups := []input.FileGroup{{}}

	result, err := Parse(context.Background(), zerolog.Nop(), inGroups)
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Empty(t, result[0])
}

func TestParse_InvalidYAML(t *testing.T) {
	inGroups := []input.FileGroup{
		{
			Files: []input.File{
				{Source: "bad.yaml", Content: []byte("not: valid: openapi: {{{}}")},
			},
		},
	}

	_, err := Parse(context.Background(), zerolog.Nop(), inGroups)
	assert.Error(t, err)
}

func TestParse_InvalidOpenAPIContent(t *testing.T) {
	// Completely broken content that the loader can't parse
	inGroups := []input.FileGroup{
		{
			Files: []input.File{
				{Source: "bad.yaml", Content: []byte(`[[[invalid`)},
			},
		},
	}

	_, err := Parse(context.Background(), zerolog.Nop(), inGroups)
	assert.Error(t, err)
}
