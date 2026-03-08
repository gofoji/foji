package output

import (
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gofoji/foji/cfg"
	"github.com/gofoji/foji/errs"
	"github.com/gofoji/foji/foji"
	"github.com/gofoji/foji/input/openapi"
	"github.com/gofoji/foji/stringlist"
)

func loadTestDoc(t *testing.T) *openapi3.T {
	t.Helper()

	doc, err := openapi3.NewLoader().LoadFromFile("testdata/openapi.yaml")
	require.NoError(t, err)

	return doc
}

func TestGetOpHappyResponse(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)
	got := ctx.GetOpHappyResponse("", ctx.API.Paths.Find("/examples/HappyPath").Get)
	assert.Equal(t, "200", got.Key)

	got = ctx.GetOpHappyResponse("", ctx.API.Paths.Find("/examples/HappyPathCreate").Post)
	assert.Equal(t, "201", got.Key)
}

func TestGetOpHappyResponse_JsonContent(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)
	got := ctx.GetOpHappyResponse("foreign", ctx.API.Paths.Find("/examples/JsonResponse").Get)
	assert.Equal(t, "200", got.Key)
	assert.Equal(t, MimeType(ApplicationJSON), got.MimeType)
	assert.True(t, got.MimeType.IsJson())
	assert.Contains(t, got.GoType, "Main")
	assert.True(t, strings.HasPrefix(got.GoType, "*"))
}

func TestGetOpHappyResponse_TextContent(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)
	got := ctx.GetOpHappyResponse("", ctx.API.Paths.Find("/examples/TextResponse").Get)
	assert.Equal(t, "200", got.Key)
	assert.Equal(t, MimeType(TextPlain), got.MimeType)
	assert.True(t, got.MimeType.IsText())
}

func TestGetOpHappyResponse_CsvContent(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)
	got := ctx.GetOpHappyResponse("", ctx.API.Paths.Find("/examples/CsvResponse").Get)
	assert.Equal(t, "200", got.Key)
	assert.Equal(t, MimeType(TextCSV), got.MimeType)
	assert.True(t, got.MimeType.IsCSV())
}

func TestGetOpHappyResponse_HtmlContent(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)
	got := ctx.GetOpHappyResponse("", ctx.API.Paths.Find("/examples/HtmlResponse").Get)
	assert.Equal(t, "200", got.Key)
	assert.Equal(t, MimeType(TextHTML), got.MimeType)
	assert.True(t, got.MimeType.IsHTML())
}

func TestGetOpHappyResponse_ArrayContent(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)
	got := ctx.GetOpHappyResponse("foreign", ctx.API.Paths.Find("/examples/ArrayResponse").Get)
	assert.Equal(t, "200", got.Key)
	assert.True(t, strings.HasPrefix(got.GoType, "[]"))
}

func TestGetOpHappyResponse_MapContent(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)
	got := ctx.GetOpHappyResponse("", ctx.API.Paths.Find("/examples/MapResponse").Get)
	assert.Equal(t, "200", got.Key)
	assert.Equal(t, "map[string]string", got.GoType)
}

func TestGetOpHappyResponse_StreamContent(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)
	got := ctx.GetOpHappyResponse("", ctx.API.Paths.Find("/examples/StreamResponse").Get)
	assert.Equal(t, "200", got.Key)
	assert.Equal(t, "io.Reader", got.GoType)
}

func TestGetOpHappyResponse_WithHeaders(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)
	got := ctx.GetOpHappyResponse("foreign", ctx.API.Paths.Find("/examples/HeaderResponse").Get)
	assert.Equal(t, "200", got.Key)
	assert.Len(t, got.Headers, 2)
	assert.Contains(t, got.Headers, "X-Request-Id")
	assert.Contains(t, got.Headers, "X-Total-Count")
}

func TestGetOpHappyResponse_HeaderOnly(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)
	got := ctx.GetOpHappyResponse("", ctx.API.Paths.Find("/examples/HeaderOnlyResponse").Get)
	assert.Equal(t, "204", got.Key)
	assert.Empty(t, got.GoType)
	assert.Len(t, got.Headers, 1)
}

func TestGetOpHappyResponse_NoContent(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)
	got := ctx.GetOpHappyResponse("", ctx.API.Paths.Find("/examples/NoContentResponse").Get)
	assert.Equal(t, "204", got.Key)
	assert.Empty(t, got.GoType)
}

func TestGetOpHappyResponse_JsonlContent(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)
	got := ctx.GetOpHappyResponse("foreign", ctx.API.Paths.Find("/examples/JsonlResponse").Get)
	assert.Equal(t, "200", got.Key)
	assert.Equal(t, MimeType(ApplicationJSONL), got.MimeType)
	assert.True(t, got.MimeType.IsLongPollingOperation())
}

func TestGetOpHappyResponseKey(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)
	key := ctx.GetOpHappyResponseKey(ctx.API.Paths.Find("/examples/JsonResponse").Get)
	assert.Equal(t, "200", key)
}

func TestGetOpHappyResponseMimeType(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)
	mt := ctx.GetOpHappyResponseMimeType(ctx.API.Paths.Find("/examples/JsonResponse").Get)
	assert.Equal(t, ApplicationJSON, mt)
}

func TestGetOpHappyResponseType(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)
	goType := ctx.GetOpHappyResponseType("foreign", ctx.API.Paths.Find("/examples/JsonResponse").Get)
	assert.Contains(t, goType, "Main")
}

func TestGetOpHappyResponseHeaders(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)
	headers := ctx.GetOpHappyResponseHeaders("foreign", ctx.API.Paths.Find("/examples/HeaderResponse").Get)
	assert.Len(t, headers, 2)
}

func TestGetType(t *testing.T) {
	doc := loadTestDoc(t)
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
		{"Complex.status", "local", "ComplexStatus"},
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

func TestGetType_BinaryField(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)
	got := ctx.GetType("", "BinaryField.upload", evalPath(doc, "BinaryField.upload"))
	assert.Equal(t, "forms.File", got)
}

func TestGetType_EnumSchema(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)
	got := ctx.GetType("local", "EnumSchema", evalPath(doc, "EnumSchema"))
	assert.Equal(t, "EnumSchema", got)
}

func TestGetType_Nil(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)
	got := ctx.GetType("", "", nil)
	assert.Equal(t, "", got)
}

func TestGetTypeName(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)

	ref := doc.Components.Schemas["Nest"]
	// Manually create a schema ref with a ref path
	s := &openapi3.SchemaRef{Ref: "#/components/schemas/Nest", Value: ref.Value}
	got := ctx.GetTypeName("foreign", s)
	assert.Equal(t, "local.Nest", got)

	got = ctx.GetTypeName("local", s)
	assert.Equal(t, "Nest", got)
}

func TestTypeOnly(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)

	assert.Equal(t, "Type", ctx.TypeOnly("pkg.Type"))
	assert.Equal(t, "Type", ctx.TypeOnly("github.com/foo/bar.Type"))
	assert.Equal(t, "simple", ctx.TypeOnly("simple"))
}

func TestRefToName(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)

	assert.Equal(t, "local.Nest", ctx.RefToName("#/components/schemas/Nest"))
	assert.Equal(t, "local.Main", ctx.RefToName("#/components/schemas/Main"))
}

func TestEnumName(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)
	assert.Equal(t, "local.MyEnum", ctx.EnumName("my_enum"))
}

func TestEnumNew(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)
	assert.Equal(t, "local.NewMyType", ctx.EnumNew("local.MyType"))
	assert.Equal(t, "local.NewMyType", ctx.EnumNew("[]local.MyType"))
}

func TestStripArray(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)
	assert.Equal(t, "string", ctx.StripArray("[]string"))
	assert.Equal(t, "string", ctx.StripArray("string"))
}

func TestInit(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)
	ctx.AbortError = errs.ErrNotNeeded
	ctx.Imports = Imports{"foo"}

	err := ctx.Init()
	require.NoError(t, err)
	assert.Nil(t, ctx.AbortError)
	assert.Nil(t, ctx.Imports)
}

func TestComponentSchemas(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)

	schemas := ctx.ComponentSchemas()
	assert.NotNil(t, schemas)
	assert.Contains(t, schemas, "Main")
	assert.Contains(t, schemas, "Nest")
}

func TestComponentSchemas_NilComponents(t *testing.T) {
	ctx := getContext(&openapi3.T{})
	assert.Nil(t, ctx.ComponentSchemas())
}

func TestComponentParameters(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)

	params := ctx.ComponentParameters()
	assert.NotNil(t, params)
	assert.Contains(t, params, "PageParam")
}

func TestComponentParameters_NilComponents(t *testing.T) {
	ctx := getContext(&openapi3.T{})
	assert.Nil(t, ctx.ComponentParameters())
}

func TestCheckAllTypes(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)
	result := ctx.CheckAllTypes("foreign", "time.Time")
	assert.Equal(t, "", result)
	// Should have added imports
	assert.True(t, len(ctx.Imports) > 0)
}

func TestHasValidation(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)

	validated := doc.Components.Schemas["Validated"]
	assert.True(t, ctx.HasValidation(validated))

	// Main has no validation constraints
	main := doc.Components.Schemas["Main"]
	assert.False(t, ctx.HasValidation(main))

	// Complex has allOf - check recursion
	complex := doc.Components.Schemas["Complex"]
	assert.False(t, ctx.HasValidation(complex))
}

func TestIsDefaultEnum(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)

	enumSchema := doc.Components.Schemas["EnumSchema"]
	assert.True(t, ctx.IsDefaultEnum("EnumSchema", enumSchema))

	// When overridden in type map, should return false
	assert.False(t, ctx.IsDefaultEnum("EmptyAlias", doc.Components.Schemas["EmptyAlias"]))
}

func TestIsRequiredProperty(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)

	main := doc.Components.Schemas["Main"]
	assert.True(t, ctx.IsRequiredProperty("page", main))
	assert.False(t, ctx.IsRequiredProperty("pageSize", main))

	// allOf required
	complex := doc.Components.Schemas["Complex"]
	assert.True(t, ctx.IsRequiredProperty("id", complex))
	assert.False(t, ctx.IsRequiredProperty("status", complex))
}

func TestIsRequiredProperty_AnyOf(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)

	withAnyOf := doc.Components.Schemas["WithAnyOf"]
	// name is required in all anyOf schemas
	assert.True(t, ctx.IsRequiredProperty("name", withAnyOf))
	// age is not required in all
	assert.False(t, ctx.IsRequiredProperty("age", withAnyOf))
}

func TestIsRequiredProperty_OneOf(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)

	withOneOf := doc.Components.Schemas["WithOneOf"]
	// id is required in all oneOf schemas
	assert.True(t, ctx.IsRequiredProperty("id", withOneOf))
	// label is not required in all
	assert.False(t, ctx.IsRequiredProperty("label", withOneOf))
}

func TestHasRequiredProperties(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)

	assert.True(t, ctx.HasRequiredProperties(doc.Components.Schemas["Main"]))
	assert.False(t, ctx.HasRequiredProperties(doc.Components.Schemas["Nest"]))
}

func TestRequiredProperties(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)

	props := ctx.RequiredProperties(doc.Components.Schemas["Main"])
	assert.Contains(t, props, "page")
	assert.NotContains(t, props, "pageSize")
}

func TestSchemaProperties(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)

	props := ctx.SchemaProperties(doc.Components.Schemas["Main"])
	assert.Contains(t, props, "page")
	assert.Contains(t, props, "list")

	// Complex has allOf - should merge properties
	complexProps := ctx.SchemaProperties(doc.Components.Schemas["Complex"])
	assert.Contains(t, complexProps, "id")
	assert.Contains(t, complexProps, "page") // from merged Main
}

func TestSchemaPropertiesHaveDefaults(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)

	assert.True(t, ctx.SchemaPropertiesHaveDefaults(doc.Components.Schemas["WithDefaults"]))
	assert.False(t, ctx.SchemaPropertiesHaveDefaults(doc.Components.Schemas["Nest"]))
}

func TestSchemaEnums(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)

	enums := ctx.SchemaEnums(doc.Components.Schemas["Complex"])
	assert.Contains(t, enums, "status")
	assert.NotContains(t, enums, "id")
}

func TestSchemaIsEnum(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)

	assert.True(t, ctx.SchemaIsEnum(doc.Components.Schemas["EnumSchema"]))
	assert.False(t, ctx.SchemaIsEnum(doc.Components.Schemas["Nest"]))
}

func TestSchemaIsEnumArray(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)

	statuses := doc.Components.Schemas["EnumArray"].Value.Properties["statuses"]
	assert.True(t, ctx.SchemaIsEnumArray(statuses))

	// Non-enum array
	mainList := doc.Components.Schemas["Main"].Value.Properties["list"]
	assert.False(t, ctx.SchemaIsEnumArray(mainList))
}

func TestSchemaContainsAllOf(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)

	assert.True(t, ctx.SchemaContainsAllOf(doc.Components.Schemas["Complex"]))
	assert.False(t, ctx.SchemaContainsAllOf(doc.Components.Schemas["Nest"]))
	assert.False(t, ctx.SchemaContainsAllOf(nil))
}

func TestSchemaIsComplex(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)

	// nil
	assert.False(t, ctx.SchemaIsComplex(nil))

	// has ref -> not complex (it's a reference)
	refSchema := &openapi3.SchemaRef{Ref: "#/components/schemas/Nest", Value: &openapi3.Schema{}}
	assert.False(t, ctx.SchemaIsComplex(refSchema))

	// object type -> complex
	assert.True(t, ctx.SchemaIsComplex(doc.Components.Schemas["Nest"]))

	// allOf -> complex
	assert.True(t, ctx.SchemaIsComplex(doc.Components.Schemas["Complex"]))

	// array of objects -> complex
	assert.True(t, ctx.SchemaIsComplex(doc.Components.Schemas["ArrayOfObjects"]))

	// array of refs -> not complex
	assert.False(t, ctx.SchemaIsComplex(doc.Components.Schemas["ArrayOfRefs"]))

	// simple type -> not complex
	assert.False(t, ctx.SchemaIsComplex(doc.Components.Schemas["EnumSchema"]))
}

func TestSchemaIsObject(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)

	assert.True(t, ctx.SchemaIsObject(doc.Components.Schemas["Nest"]))
	// string is also "object" in this helper (for timestamps, uuids)
	assert.True(t, ctx.SchemaIsObject(doc.Components.Schemas["EnumSchema"]))
}

func TestGetRequestBody(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)

	// JSON body
	body := ctx.GetRequestBody(ctx.API.Paths.Find("/examples/JsonBody").Post)
	require.NotNil(t, body)
	assert.Equal(t, MimeType(ApplicationJSON), body.MimeType)

	// Form body
	body = ctx.GetRequestBody(ctx.API.Paths.Find("/examples/FormBody").Post)
	require.NotNil(t, body)
	assert.Equal(t, MimeType(ApplicationForm), body.MimeType)

	// Multipart body
	body = ctx.GetRequestBody(ctx.API.Paths.Find("/examples/MultipartBody").Post)
	require.NotNil(t, body)
	assert.Equal(t, MimeType(MultipartForm), body.MimeType)

	// Text body
	body = ctx.GetRequestBody(ctx.API.Paths.Find("/examples/TextBody").Post)
	require.NotNil(t, body)
	assert.Equal(t, MimeType(TextPlain), body.MimeType)

	// No body (GET)
	body = ctx.GetRequestBody(ctx.API.Paths.Find("/examples/JsonResponse").Get)
	assert.Nil(t, body)
}

func TestGetRequestBodySchemas(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)

	// Multi body
	schemas := ctx.GetRequestBodySchemas(ctx.API.Paths.Find("/examples/MultiBody").Post)
	assert.Len(t, schemas, 3)

	// Nil op
	schemas = ctx.GetRequestBodySchemas(nil)
	assert.Nil(t, schemas)

	// No body
	schemas = ctx.GetRequestBodySchemas(ctx.API.Paths.Find("/examples/JsonResponse").Get)
	assert.Nil(t, schemas)
}

func TestOpParams(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)

	path := ctx.API.Paths.Find("/examples/WithParams/{id}")
	params := ctx.OpParams(path, path.Get)
	// path-level params (shared) + op-level params (id, filter, status, tags, page, mapParam)
	assert.Len(t, params, 7)
}

func TestDefaultValues(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)

	// Empty
	assert.Nil(t, ctx.DefaultValues(""))

	// Simple value
	vals := ctx.DefaultValues("hello")
	assert.Equal(t, []string{"hello"}, vals)

	// CSV array
	vals = ctx.DefaultValues("[a,b,c]")
	assert.Equal(t, []string{"a", "b", "c"}, vals)

	// Single element array
	vals = ctx.DefaultValues("[x]")
	assert.Equal(t, []string{"x"}, vals)

	// Quoted CSV
	vals = ctx.DefaultValues(`["hello world","foo"]`)
	assert.Equal(t, []string{"hello world", "foo"}, vals)
}

func TestParamIsOptionalType(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)

	path := ctx.API.Paths.Find("/examples/WithParams/{id}")

	for _, p := range path.Get.Parameters {
		switch p.Value.Name {
		case "id":
			// required
			assert.False(t, ctx.ParamIsOptionalType(p))
		case "filter":
			// optional string, no default
			assert.True(t, ctx.ParamIsOptionalType(p))
		case "tags":
			// array type - not optional type
			assert.False(t, ctx.ParamIsOptionalType(p))
		case "page":
			// has default - not optional type
			assert.False(t, ctx.ParamIsOptionalType(p))
		case "mapParam":
			// map type - not optional type
			assert.False(t, ctx.ParamIsOptionalType(p))
		}
	}
}

func TestParamIsEnum(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)

	path := ctx.API.Paths.Find("/examples/WithParams/{id}")
	for _, p := range path.Get.Parameters {
		switch p.Value.Name {
		case "status":
			assert.True(t, ctx.ParamIsEnum(p))
		case "filter":
			assert.False(t, ctx.ParamIsEnum(p))
		}
	}
}

func TestParamIsEnumArray(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)

	path := ctx.API.Paths.Find("/examples/WithParams/{id}")
	for _, p := range path.Get.Parameters {
		switch p.Value.Name {
		case "tags":
			assert.True(t, ctx.ParamIsEnumArray(p))
		case "filter":
			assert.False(t, ctx.ParamIsEnumArray(p))
		}
	}
}

func TestHasExtensionValue(t *testing.T) {
	// No extension
	assert.False(t, HasExtensionValue(nil, "x-foo"))
	assert.False(t, HasExtensionValue(map[string]any{}, "x-foo"))

	// Bool true
	assert.True(t, HasExtensionValue(map[string]any{"x-foo": true}, "x-foo"))

	// Bool false
	assert.False(t, HasExtensionValue(map[string]any{"x-foo": false}, "x-foo"))

	// Non-bool (string) - exists so true
	assert.True(t, HasExtensionValue(map[string]any{"x-foo": "bar"}, "x-foo"))
}

func TestOpHasExtension(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)

	rawOp := ctx.API.Paths.Find("/examples/RawOp").Get
	assert.True(t, ctx.OpHasExtension(rawOp, "x-raw-request"))
	assert.True(t, ctx.OpHasExtension(rawOp, "x-raw-response"))
	assert.True(t, ctx.OpHasExtension(rawOp, "x-raw-body"))
	assert.True(t, ctx.OpHasExtension(rawOp, "x-raw-auth"))

	normalOp := ctx.API.Paths.Find("/examples/JsonResponse").Get
	assert.False(t, ctx.OpHasExtension(normalOp, "x-raw-request"))
}

func TestSecurityHasExtension(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)

	scheme := doc.Components.SecuritySchemes["HeaderAuth"]
	assert.False(t, ctx.SecurityHasExtension(scheme, "x-custom"))
}

func TestHasExtension(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)

	mapTest := doc.Components.Schemas["MapTest"]
	assert.True(t, ctx.HasExtension(mapTest, "x-go-type"))

	nest := doc.Components.Schemas["Nest"]
	assert.False(t, ctx.HasExtension(nest, "x-go-type"))
}

/* Auth Tests */

func TestOpSecurity(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)

	// Op with explicit security
	op := ctx.API.Paths.Find("/examples/HappyPath").Get
	sec := ctx.OpSecurity(op)
	assert.NotEmpty(t, sec)

	// Op without explicit security falls back to global
	op = ctx.API.Paths.Find("/examples/JsonResponse").Get
	sec = ctx.OpSecurity(op)
	assert.NotEmpty(t, sec) // global security
}

func TestHasAuthentication(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)
	assert.True(t, ctx.HasAuthentication())

	// No security schemes
	ctx2 := getContext(&openapi3.T{})
	assert.False(t, ctx2.HasAuthentication())
}

func TestHasAuthorization(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)
	// The ComplexAuth operation has scopes
	assert.True(t, ctx.HasAuthorization())
}

func TestHasAuthorization_NoScopes(t *testing.T) {
	// Build a spec with security but no scopes
	spec := &openapi3.T{
		OpenAPI: "3.0.2",
		Info:    &openapi3.Info{Title: "Test", Version: "1.0"},
		Security: openapi3.SecurityRequirements{
			{"HeaderAuth": {}},
		},
		Paths: &openapi3.Paths{},
	}
	spec.Paths.Set("/test", &openapi3.PathItem{
		Get: &openapi3.Operation{
			OperationID: "test",
			Responses:   openapi3.NewResponses(),
		},
	})

	ctx := getContext(spec)
	assert.False(t, ctx.HasAuthorization())
}

func TestIsSimpleAuth(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)

	// Simple auth - single scheme
	op := ctx.API.Paths.Find("/examples/SimpleAuth").Get
	assert.True(t, ctx.IsSimpleAuth(op))

	// No auth - empty security
	op = ctx.API.Paths.Find("/examples/NoAuth").Get
	assert.True(t, ctx.IsSimpleAuth(op))

	// Complex auth - multiple different schemes
	op = ctx.API.Paths.Find("/examples/ComplexAuth").Get
	assert.False(t, ctx.IsSimpleAuth(op))

	// Public with optional auth - {} and HeaderAuth - different auth names
	op = ctx.API.Paths.Find("/examples/PublicWithOptionalAuth").Get
	assert.False(t, ctx.IsSimpleAuth(op))
}

func TestHasComplexAuth(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)
	assert.True(t, ctx.HasComplexAuth())
}

func TestHasComplexAuth_AllSimple(t *testing.T) {
	spec := &openapi3.T{
		OpenAPI: "3.0.2",
		Info:    &openapi3.Info{Title: "Test", Version: "1.0"},
		Paths:   &openapi3.Paths{},
	}
	spec.Paths.Set("/test", &openapi3.PathItem{
		Get: &openapi3.Operation{
			OperationID: "test",
			Security: &openapi3.SecurityRequirements{
				{"api_key": {}},
			},
			Responses: openapi3.NewResponses(),
		},
	})

	ctx := getContext(spec)
	assert.False(t, ctx.HasComplexAuth())
}

func TestHasBasicAuth(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)
	assert.True(t, ctx.HasBasicAuth())
}

func TestHasBasicAuth_NonePresent(t *testing.T) {
	spec := &openapi3.T{
		OpenAPI: "3.0.2",
		Info:    &openapi3.Info{Title: "Test", Version: "1.0"},
		Components: &openapi3.Components{
			SecuritySchemes: openapi3.SecuritySchemes{
				"api_key": &openapi3.SecuritySchemeRef{
					Value: &openapi3.SecurityScheme{Type: "apiKey", In: "header", Name: "X-API-Key"},
				},
			},
		},
	}
	ctx := getContext(spec)
	assert.False(t, ctx.HasBasicAuth())
}

func TestHasBearerAuth(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)
	assert.True(t, ctx.HasBearerAuth())
}

func TestHasBearerAuth_NonePresent(t *testing.T) {
	spec := &openapi3.T{
		OpenAPI: "3.0.2",
		Info:    &openapi3.Info{Title: "Test", Version: "1.0"},
		Components: &openapi3.Components{
			SecuritySchemes: openapi3.SecuritySchemes{
				"basic": &openapi3.SecuritySchemeRef{
					Value: &openapi3.SecurityScheme{Type: "http", Scheme: "basic"},
				},
			},
		},
	}
	ctx := getContext(spec)
	assert.False(t, ctx.HasBearerAuth())
}

func TestHasAnyAuth(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)

	// Op with auth
	op := ctx.API.Paths.Find("/examples/SimpleAuth").Get
	assert.True(t, ctx.HasAnyAuth(op))

	// Op with no auth
	op = ctx.API.Paths.Find("/examples/NoAuth").Get
	assert.False(t, ctx.HasAnyAuth(op))

	// Op with empty group (public option)
	op = ctx.API.Paths.Find("/examples/PublicWithOptionalAuth").Get
	assert.True(t, ctx.HasAnyAuth(op))
}

func TestRequiresAuthUser(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)

	// Required auth
	op := ctx.API.Paths.Find("/examples/SimpleAuth").Get
	assert.True(t, ctx.RequiresAuthUser(op))

	// No auth
	op = ctx.API.Paths.Find("/examples/NoAuth").Get
	assert.False(t, ctx.RequiresAuthUser(op))

	// Public with optional auth - has empty group, so not required
	op = ctx.API.Paths.Find("/examples/PublicWithOptionalAuth").Get
	assert.False(t, ctx.RequiresAuthUser(op))
}

/* Context Tests */

func TestNewContext(t *testing.T) {
	p := cfg.Process{Params: cfg.ParamMap{"Package": "test"}}
	ctx := NewContext(p, zerolog.Nop())
	assert.Equal(t, p, ctx.Process)
	assert.Nil(t, ctx.AbortError)
}

func TestContext_Aborted(t *testing.T) {
	ctx := Context{}
	assert.Nil(t, ctx.Aborted())

	ctx.AbortError = errs.ErrNotNeeded
	assert.Equal(t, errs.ErrNotNeeded, ctx.Aborted())
}

func TestContext_NotNeededIf(t *testing.T) {
	ctx := Context{}

	s, err := ctx.NotNeededIf(false, "test")
	assert.NoError(t, err)
	assert.Equal(t, "", s)
	assert.Nil(t, ctx.AbortError)

	s, err = ctx.NotNeededIf(true, "test reason")
	assert.Error(t, err)
	assert.Equal(t, "", s)
	assert.ErrorIs(t, ctx.AbortError, errs.ErrNotNeeded)
}

func TestContext_ErrorIf(t *testing.T) {
	ctx := Context{}

	s, err := ctx.ErrorIf(false, "test")
	assert.NoError(t, err)
	assert.Equal(t, "", s)

	s, err = ctx.ErrorIf(true, "test reason")
	assert.Error(t, err)
	assert.Equal(t, "", s)
	assert.ErrorIs(t, ctx.AbortError, errs.ErrMissingRequirement)
}

func TestContext_WithParams(t *testing.T) {
	ctx := Context{Process: cfg.Process{Params: cfg.ParamMap{"Package": "test"}}}

	out, err := ctx.WithParams("key1", "val1", "key2", 42)
	require.NoError(t, err)
	assert.Equal(t, "val1", out.RuntimeParams["key1"])
	assert.Equal(t, 42, out.RuntimeParams["key2"])
}

func TestContext_WithParams_OddArgs(t *testing.T) {
	ctx := Context{}
	_, err := ctx.WithParams("key1")
	assert.ErrorIs(t, err, errs.ErrInvalidDictParams)
}

func TestContext_WithParams_NonStringKey(t *testing.T) {
	ctx := Context{}
	_, err := ctx.WithParams(42, "val")
	assert.ErrorIs(t, err, errs.ErrInvalidDictKey)
}

func TestContext_Funcs(t *testing.T) {
	ctx := Context{Process: cfg.Process{Case: "pascal"}}
	funcs := ctx.Funcs()
	assert.NotNil(t, funcs)
}

func TestContext_ToCase(t *testing.T) {
	ctx := Context{Process: cfg.Process{Case: "pascal"}}
	assert.Equal(t, "MyName", ctx.ToCase("my_name"))

	// Unknown case - returns name as-is
	ctx2 := Context{Process: cfg.Process{Case: ""}}
	assert.Equal(t, "my_name", ctx2.ToCase("my_name"))
}

func TestContext_PackageName(t *testing.T) {
	ctx := Context{Process: cfg.Process{Params: cfg.ParamMap{"Package": "github.com/foo/bar/pkg"}}}
	assert.Equal(t, "pkg", ctx.PackageName())

	ctx2 := Context{Process: cfg.Process{Params: cfg.ParamMap{"Package": "simple"}}}
	assert.Equal(t, "simple", ctx2.PackageName())

	ctx3 := Context{Process: cfg.Process{Params: cfg.ParamMap{}}}
	assert.Equal(t, "", ctx3.PackageName())
}

/* GoImports Tests */

func TestGoImports(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)
	ctx.Imports = Imports{"time", "local", "fmt"}

	imports := ctx.GoImports()
	assert.Contains(t, imports, "time")
	assert.Contains(t, imports, "fmt")
	assert.NotContains(t, imports, "local") // same package filtered out
}

func TestGoImports_Empty(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)
	imports := ctx.GoImports()
	assert.Empty(t, imports)
}

/* OpenAPIFileContext WithParams */

func TestOpenAPIFileContext_WithParams(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)

	out, err := ctx.WithParams("key", "value")
	require.NoError(t, err)
	assert.Equal(t, "value", out.RuntimeParams["key"])
	assert.NotNil(t, out.API)
}

func TestOpenAPIFileContext_WithParams_Error(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)

	_, err := ctx.WithParams("odd")
	assert.Error(t, err)
}

/* MimeType Tests */

func TestMimeType_Methods(t *testing.T) {
	tests := []struct {
		mime   MimeType
		isJson bool
		isText bool
		isHTML bool
		isCSV  bool
		isForm bool
		isMP   bool
		isLP   bool
	}{
		{MimeType(ApplicationJSON), true, false, false, false, false, false, false},
		{MimeType(TextPlain), false, true, false, false, false, false, false},
		{MimeType(TextHTML), false, false, true, false, false, false, false},
		{MimeType(TextCSV), false, false, false, true, false, false, false},
		{MimeType(ApplicationForm), false, false, false, false, true, false, false},
		{MimeType(MultipartForm), false, false, false, false, false, true, false},
		{MimeType(ApplicationJSONL), false, false, false, false, false, false, true},
		{MimeType("unknown"), false, false, false, false, false, false, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.mime), func(t *testing.T) {
			assert.Equal(t, tt.isJson, tt.mime.IsJson())
			assert.Equal(t, tt.isText, tt.mime.IsText())
			assert.Equal(t, tt.isHTML, tt.mime.IsHTML())
			assert.Equal(t, tt.isCSV, tt.mime.IsCSV())
			assert.Equal(t, tt.isForm, tt.mime.IsForm())
			assert.Equal(t, tt.isMP, tt.mime.IsMultipartForm())
			assert.Equal(t, tt.isLP, tt.mime.IsLongPollingOperation())
		})
	}
}

func TestMimeType_String(t *testing.T) {
	m := MimeType(ApplicationJSON)
	assert.Equal(t, ApplicationJSON, m.String())
}

/* CheckPackage / Imports Tests */

func TestCheckPackage(t *testing.T) {
	var ii Imports

	// Base type (no package)
	assert.Equal(t, "int", ii.CheckPackage("int", ""))

	// Same package
	assert.Equal(t, "Type", ii.CheckPackage("pkg.Type", "pkg"))

	// Different package - short
	assert.Equal(t, "time.Time", ii.CheckPackage("time.Time", ""))
	assert.Contains(t, Imports(ii), "time")

	// Different package - full path
	ii = nil
	assert.Equal(t, "uuid.UUID", ii.CheckPackage("github.com/google/uuid.UUID", ""))
	assert.Contains(t, Imports(ii), "github.com/google/uuid")

	// Array prefix
	ii = nil
	assert.Equal(t, "[]uuid.UUID", ii.CheckPackage("[]github.com/google/uuid.UUID", ""))

	// Pointer prefix
	ii = nil
	assert.Equal(t, "*foo.Bar", ii.CheckPackage("*foo.Bar", ""))
}

func TestImports_Add_Dedup(t *testing.T) {
	var ii Imports
	ii.Add("time")
	ii.Add("time")
	ii.Add("fmt")
	assert.Len(t, ii, 2)
}

/* HasOpenAPIOutput */

func TestHasOpenAPIOutput(t *testing.T) {
	o := cfg.Output{
		OpenAPIFile: stringlist.StringMap{"out.go": "tpl.go"},
	}
	assert.True(t, HasOpenAPIOutput(o))

	o2 := cfg.Output{}
	assert.False(t, HasOpenAPIOutput(o2))
}

/* happyStatusCode */

func TestHappyStatusCode(t *testing.T) {
	assert.True(t, happyStatusCode("200"))
	assert.True(t, happyStatusCode("201"))
	assert.True(t, happyStatusCode("301"))
	assert.False(t, happyStatusCode("404"))
	assert.False(t, happyStatusCode("500"))
	assert.False(t, happyStatusCode("20"))   // too short
	assert.False(t, happyStatusCode("2000")) // too long
}

/* Edge case tests for remaining coverage gaps */

func TestGetTypeName_WithOverride(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)

	// Add a type map override for a ref name
	ctx.Maps.Type["local.Nest"] = "CustomNest"
	s := &openapi3.SchemaRef{Ref: "#/components/schemas/Nest", Value: doc.Components.Schemas["Nest"].Value}
	got := ctx.GetTypeName("foreign", s)
	assert.Equal(t, "CustomNest", got)
}

func TestDefaultValues_CSVError(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)
	// Malformed CSV (unclosed quote) - must start with [ and end with ]
	result := ctx.DefaultValues(`["unclosed]`)
	assert.Nil(t, result)
	assert.Error(t, ctx.AbortError)
}

func TestDefaultValues_EmptyBrackets(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)
	result := ctx.DefaultValues("[]")
	assert.Nil(t, result)
}

func TestGetRequestBodySchemas_TextBody(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)
	// Text body has text/plain which is not JSON/Form/Multipart, so should be excluded
	schemas := ctx.GetRequestBodySchemas(ctx.API.Paths.Find("/examples/TextBody").Post)
	assert.Empty(t, schemas)
}

func TestHasAnyAuth_EmptyGroups(t *testing.T) {
	doc := loadTestDoc(t)
	ctx := getContext(doc)

	// HappyPath has: HeaderAuth with scopes + Jwt with no scopes
	op := ctx.API.Paths.Find("/examples/HappyPath").Get
	assert.True(t, ctx.HasAnyAuth(op))
}

func TestHasValidation_WithAllOf(t *testing.T) {
	// Create a schema where validation is in an allOf sub-schema
	schema := &openapi3.SchemaRef{
		Value: &openapi3.Schema{
			AllOf: openapi3.SchemaRefs{
				{
					Value: &openapi3.Schema{
						MinLength: 5,
					},
				},
			},
		},
	}
	doc := loadTestDoc(t)
	ctx := getContext(doc)
	assert.True(t, ctx.HasValidation(schema))
}

func TestHasValidation_WithNestedProperty(t *testing.T) {
	// Validation in a nested property
	schema := &openapi3.SchemaRef{
		Value: &openapi3.Schema{
			Properties: openapi3.Schemas{
				"nested": &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						MinLength: 1,
					},
				},
			},
		},
	}
	doc := loadTestDoc(t)
	ctx := getContext(doc)
	assert.True(t, ctx.HasValidation(schema))
}

/* Helpers */

func getContext(doc *openapi3.T) OpenAPIFileContext {
	defaultConfig, err := cfg.LoadYaml(foji.DefaultConfig)
	if err != nil {
		panic(err)
	}

	config := cfg.Config{Processes: cfg.Processes{"openAPI": cfg.Process{Maps: cfg.Maps{
		Type: stringlist.StringMap{"Main.override_name": "typeOverride", "EmptyAlias": "myOverride", "object": "DefaultObject"},
	}, Params: cfg.ParamMap{"Package": "local"}}}}
	config = config.Merge(defaultConfig)

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
