package output

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/codemodus/kace"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/rs/zerolog"

	"github.com/gofoji/foji/cfg"
	"github.com/gofoji/foji/input/openapi"
)

const (
	OpenAPIFile = "OpenAPIFile"
	OpenAPITag  = "OpenAPITag"
)

func HasOpenAPIOutput(o cfg.Output) bool {
	return hasAnyOutput(o, OpenAPIFile, OpenAPITag)
}

func OpenAPI(p cfg.Process, fn cfg.FileHandler, l zerolog.Logger, groups openapi.FileGroups, simulate bool) error {
	runner := NewProcessRunner(p.RootDir, fn, l, simulate)

	for _, ff := range groups {
		for _, f := range ff {
			ctx := OpenAPIFileContext{
				Context: Context{Process: p, Logger: l},
				File:    f,
			}

			err := runner.process(p.Output[OpenAPIFile], &ctx)
			if err != nil {
				if len(ctx.RuntimeParams) > 0 {
					return fmt.Errorf("%w:%v", err, ctx.RuntimeParams)
				}

				return err
			}
		}
	}

	return nil
}

type OpenAPIFileContext struct {
	Context
	Imports
	openapi.File
}

func (o *OpenAPIFileContext) GoImports() []string {
	var out []string //nolint:prealloc

	for _, i := range o.Imports {
		if i == o.PackageName() {
			continue
		}

		out = append(out, i)
	}

	return out
}

func (o *OpenAPIFileContext) WithParams(values ...any) (*OpenAPIFileContext, error) {
	ctx, err := o.Context.WithParams(values...)
	if err != nil {
		return nil, err
	}

	out := *o
	out.Context = *ctx

	return &out, nil
}

func (o *OpenAPIFileContext) RefToName(ref string) string {
	modelPackage := o.PackageName()
	parts := strings.Split(ref, "/")

	return modelPackage + "." + o.ToCase(parts[len(parts)-1])
}

func (o *OpenAPIFileContext) GetTypeName(pkg string, s *openapi3.SchemaRef) string {
	ref := o.RefToName(s.Ref)

	if t, ok := o.Maps.Type[ref]; ok {
		return o.CheckPackage(t, pkg)
	}

	return o.CheckPackage(ref, pkg)
}

func (o *OpenAPIFileContext) TypeOnly(name string) string {
	tt := strings.Split(name, ".")

	return tt[len(tt)-1]
}

// getXGoType maps x-go-type declarations to an actual type definition.
// Supports formats:
//
//	x-go-type: full/path/to.type
//	x-go-type: int
func (o *OpenAPIFileContext) getXGoType(currentPackage string, goType any) string {
	if s, ok := goType.(string); ok {
		return o.CheckPackage(s, currentPackage)
	}

	return fmt.Sprintf("INVALID x-go-type: %v", goType)
}

// HasExtensionValue checks if an extension exists and has a truthy value.
// For boolean extensions, it returns the boolean value.
// For other extensions, it returns true if they exist.
func HasExtensionValue(extensions map[string]interface{}, ext string) bool {
	v, ok := extensions[ext]
	if !ok {
		return false
	}

	if b, isBool := v.(bool); isBool {
		return b
	}

	return true
}

func (o *OpenAPIFileContext) OpHasExtension(op *openapi3.Operation, ext string) bool {
	return HasExtensionValue(op.Extensions, ext)
}

func (o *OpenAPIFileContext) SecurityHasExtension(scheme *openapi3.SecuritySchemeRef, ext string) bool {
	return HasExtensionValue(scheme.Value.Extensions, ext)
}

func (o *OpenAPIFileContext) HasExtension(s *openapi3.SchemaRef, ext string) bool {
	_, ok := s.Value.Extensions[ext]

	return ok
}

//nolint:cyclop
func (o *OpenAPIFileContext) GetType(currentPackage, name string, s *openapi3.SchemaRef) string {
	if s == nil {
		return ""
	}

	if override, ok := s.Value.Extensions["x-go-type"]; ok {
		return o.getXGoType(currentPackage, override)
	}

	if t, ok := o.Maps.Type[name]; ok {
		return o.CheckPackage(t, currentPackage)
	}

	schemaType := ""
	if s.Value.Type != nil && len(*s.Value.Type) == 1 {
		schemaType = (*s.Value.Type)[0]
	}

	if s.Value.Format != "" {
		if t, ok := o.Maps.Type[schemaType+","+s.Value.Format]; ok {
			return o.CheckPackage(t, currentPackage)
		}
	}

	if s.Ref != "" {
		return o.GetTypeName(currentPackage, s)
	}

	if s.Value.Type.Is("array") {
		return "[]" + o.GetType(currentPackage, name, s.Value.Items)
	}

	if s.Value.Type.Is("string") && s.Value.Format == "binary" {
		return "forms.File"
	}

	if s.Value.Type.Is("object") || s.Value.Type.Is("") || s.Value.Type == nil {
		if len(o.SchemaProperties(s)) == 0 {
			if t, ok := o.Maps.Type[schemaType]; ok {
				return o.CheckPackage(t, currentPackage)
			}

			return "any"
		}

		name = o.PackageName() + "." + kace.Pascal(name)

		return o.CheckPackage(name, currentPackage)
	}

	if o.IsDefaultEnum(name, s) {
		return o.CheckPackage(o.EnumName(name), currentPackage)
	}

	if t, ok := o.Maps.Type[schemaType]; ok {
		return o.CheckPackage(t, currentPackage)
	}

	return fmt.Sprintf("unknown type: name(%s): type(%s)", name, s.Value.Type)
}

func (o *OpenAPIFileContext) EnumName(name string) string {
	// TODO: Support override via template
	return o.PackageName() + "." + kace.Pascal(name)
}

func (o *OpenAPIFileContext) EnumNew(name string) string {
	name = strings.TrimPrefix(name, "[]")

	pos := strings.Index(name, ".") + 1

	return name[:pos] + "New" + name[pos:]
}

func (o *OpenAPIFileContext) StripArray(name string) string {
	if strings.HasPrefix(name, "[]") {
		return name[2:]
	}

	return name
}

func (o *OpenAPIFileContext) Init() error {
	o.AbortError = nil
	o.Imports = nil

	return nil
}

func (o *OpenAPIFileContext) ComponentSchemas() openapi3.Schemas {
	if o.API.Components == nil {
		return nil
	}

	return o.API.Components.Schemas
}

func (o *OpenAPIFileContext) ComponentParameters() openapi3.ParametersMap {
	if o.API.Components == nil {
		return nil
	}

	return o.API.Components.Parameters
}

// CheckAllTypes is a helper to iterate all property references for import requirements.
// This is expected to inject imports for unnecessary packages depending on the template
// generated, the post-processing should remove unused imports.
func (o *OpenAPIFileContext) CheckAllTypes(pkg string, types ...string) string {
	for _, s := range o.ComponentSchemas() {
		for key, schema := range s.Value.Properties {
			o.GetType(pkg, key, schema)
		}

		for _, nested := range s.Value.AllOf {
			for key, schema := range nested.Value.Properties {
				o.GetType(pkg, key, schema)
			}
		}
	}

	for _, s := range types {
		o.CheckPackage(s, pkg)
	}

	return ""
}

func hasValidation(s *openapi3.Schema) bool {
	return s.Min != nil || s.Max != nil || s.MultipleOf != nil || // Number
		s.MinLength > 0 || s.MaxLength != nil || len(s.Pattern) > 0 || // String
		s.MinItems > 0 || s.MaxItems != nil // Array
}

func (o *OpenAPIFileContext) HasValidation(s *openapi3.SchemaRef) bool {
	if hasValidation(s.Value) {
		return true
	}

	for _, p := range s.Value.Properties {
		if o.HasValidation(p) {
			return true
		}
	}

	for _, p := range s.Value.AllOf {
		if o.HasValidation(p) {
			return true
		}
	}

	return false
}

// IsDefaultEnum helper that checks if an enumerated type is overridden (specified externally).
func (o *OpenAPIFileContext) IsDefaultEnum(name string, s *openapi3.SchemaRef) bool {
	if len(s.Value.Enum) == 0 {
		return false
	}

	_, ok := o.Maps.Type[name]

	return !ok
}

func (o *OpenAPIFileContext) HasRequiredProperties(s *openapi3.SchemaRef) bool {
	return len(o.RequiredProperties(s)) > 0
}

// IsRequiredProperty helper that checks if a property is required.
func (o *OpenAPIFileContext) IsRequiredProperty(name string, s *openapi3.SchemaRef) bool {
	// property is required if it is listed in the schema's required properties
	if slices.Contains(s.Value.Required, name) {
		return true
	}

	// property is required if any of the allOf schemas require it
	for _, subSchema := range s.Value.AllOf {
		if slices.Contains(subSchema.Value.Required, name) {
			return true
		}
	}

	// property is required if there is at least one anyOf schema and they all require the field
	anyOfWithoutProp := false

	for _, subSchema := range s.Value.AnyOf {
		if !slices.Contains(subSchema.Value.Required, name) {
			anyOfWithoutProp = true
		}
	}

	if !anyOfWithoutProp && len(s.Value.AnyOf) > 0 {
		return true
	}

	// property is required if there is at least one oneOf schema and they all require the field
	oneOfWithoutProp := false

	for _, subSchema := range s.Value.OneOf {
		if !slices.Contains(subSchema.Value.Required, name) {
			oneOfWithoutProp = true
		}
	}

	if !oneOfWithoutProp && len(s.Value.OneOf) > 0 {
		return true
	}

	return false
}

func (o *OpenAPIFileContext) RequiredProperties(schema *openapi3.SchemaRef) openapi3.Schemas {
	out := openapi3.Schemas{}

	for name, ref := range o.SchemaProperties(schema) {
		if o.IsRequiredProperty(name, schema) {
			out[name] = ref
		}
	}

	return out
}

func (o *OpenAPIFileContext) SchemaPropertiesHaveDefaults(schema *openapi3.SchemaRef) bool {
	for _, v := range o.SchemaProperties(schema) {
		if v.Value.Default != nil {
			return true
		}
	}

	return false
}

func (o *OpenAPIFileContext) SchemaProperties(schema *openapi3.SchemaRef) openapi3.Schemas {
	out := openapi3.Schemas{}

	return schemaPropertiesWithEmbedded(schema, out)
}

func schemaPropertiesWithEmbedded(schema *openapi3.SchemaRef, out openapi3.Schemas) openapi3.Schemas {
	for k, v := range schema.Value.Properties {
		out[k] = v
	}

	for _, subSchema := range schema.Value.AllOf {
		schemaPropertiesWithEmbedded(subSchema, out)
	}

	return out
}

func (o *OpenAPIFileContext) SchemaEnums(schema *openapi3.SchemaRef) openapi3.Schemas {
	out := openapi3.Schemas{}

	for k, v := range o.SchemaProperties(schema) {
		if len(v.Value.Enum) > 0 {
			out[k] = v
		}
	}

	return out
}

func (o *OpenAPIFileContext) GetRequestBody(op *openapi3.Operation) *OpBody {
	if op.RequestBody != nil && op.RequestBody.Value != nil {
		mediaType := op.RequestBody.Value.Content.Get(ApplicationJSON)
		if mediaType != nil {
			return &OpBody{MimeType: ApplicationJSON, Schema: mediaType.Schema}
		}

		mediaType = op.RequestBody.Value.Content.Get(TextPlain)
		if mediaType != nil {
			return &OpBody{MimeType: TextPlain, Schema: mediaType.Schema}
		}

		mediaType = op.RequestBody.Value.Content.Get(ApplicationForm)
		if mediaType != nil {
			return &OpBody{MimeType: ApplicationForm, Schema: mediaType.Schema}
		}

		mediaType = op.RequestBody.Value.Content.Get(MultipartForm)
		if mediaType != nil {
			return &OpBody{MimeType: MultipartForm, Schema: mediaType.Schema}
		}
	}

	return nil
}

func (o *OpenAPIFileContext) GetRequestBodySchemas(op *openapi3.Operation) []OpBody {
	if op == nil || op.RequestBody == nil || op.RequestBody.Value == nil {
		return nil
	}

	var out []OpBody

	for k, v := range op.RequestBody.Value.Content {
		if v.Schema == nil {
			continue
		}

		if k == ApplicationJSON || k == ApplicationForm || k == MultipartForm {
			out = append(out, OpBody{MimeType: MimeType(k), Schema: v.Schema})
		}
	}

	return out
}

var knownInterfaces = []string{"string", "io.Reader"}

func happyStatusCode(key string) bool {
	if len(key) != 3 { //nolint:mnd
		return false
	}

	return key[0] == '2' || key[0] == '3'
}

func (o *OpenAPIFileContext) GetOpHappyResponse(pkg string, op *openapi3.Operation) OpResponse {
	supportedResponseContentTypes := []string{ApplicationJSON, ApplicationJSONL, TextPlain, TextHTML, TextCSV}

	// kin-openapi does not preserve response ordering, so we order here by "happy key"
	// to make sure we choose a happy response deterministically

	happyKeys := []string{}

	for key := range op.Responses.Map() {
		if happyStatusCode(key) {
			happyKeys = append(happyKeys, key)
		}
	}

	slices.Sort(happyKeys)

	for _, key := range happyKeys {
		r := op.Responses.Map()[key]
		for _, mimeType := range supportedResponseContentTypes {
			mediaType := r.Value.Content.Get(mimeType)
			if mediaType != nil {
				mime := MimeType(mimeType)
				t := o.GetType(pkg, kace.Pascal(op.OperationID)+" Response", mediaType.Schema)

				if t == "" {
					// Unknown type, use []byte by default
					t = "[]byte"
				}

				var goType string

				if strings.HasPrefix(t, "[]") || strings.HasPrefix(t, "map[") || slices.Contains(knownInterfaces, t) {
					goType = t
				} else {
					goType = "*" + t
				}

				if r.Value.Headers != nil {
					return OpResponse{Key: key, MimeType: mime, MediaType: mediaType, GoType: goType, Headers: mapKeysSorted(r.Value.Headers)}
				}

				return OpResponse{Key: key, MimeType: mime, MediaType: mediaType, GoType: goType, Headers: []string{}}
			}
		}
	}

	// No response with a supported content type found, but maybe there's a response with headers only
	for _, key := range happyKeys {
		r := op.Responses.Map()[key]
		if len(r.Value.Headers) > 0 {
			return OpResponse{Key: key, MimeType: "", MediaType: nil, GoType: "", Headers: mapKeysSorted(r.Value.Headers)}
		}
	}

	happyKey := "200"
	if len(happyKeys) > 0 {
		// If none of the responses have a supported content type, use the first "happy" response
		happyKey = happyKeys[0]
	}

	return OpResponse{Key: happyKey, MimeType: "", MediaType: nil, GoType: ""}
}

func mapKeysSorted[T any](in map[string]T) []string {
	out := slices.Collect(maps.Keys(in))
	slices.Sort(out)

	return out
}

func (o *OpenAPIFileContext) OpParams(path *openapi3.PathItem, op *openapi3.Operation) openapi3.Parameters {
	var out openapi3.Parameters

	out = append(out, path.Parameters...)
	out = append(out, op.Parameters...)

	return out
}

func (o *OpenAPIFileContext) DefaultValues(val string) []string {
	if val == "" {
		return nil
	}

	if len(val) < 2 || val[0] != '[' || val[len(val)-1] != ']' {
		return []string{val}
	}

	csvReader := csv.NewReader(bytes.NewReader([]byte(val[1 : len(val)-1])))

	records, err := csvReader.ReadAll()
	if err != nil {
		o.AbortError = fmt.Errorf("error reading csv for default: %w: %q", err, val)
		return nil
	}

	if len(records) > 0 {
		return records[0]
	}

	return nil
}

func (o *OpenAPIFileContext) ParamIsOptionalType(param *openapi3.ParameterRef) bool {
	if param.Value.Required {
		return false
	}

	if param.Value.Schema.Value.Type.Is("array") {
		return false
	}

	t := o.GetType("", "", param.Value.Schema)
	if strings.HasPrefix(t, "map[") {
		return false
	}

	return param.Value.Schema.Value.Default == nil
}

func (o *OpenAPIFileContext) ParamIsEnum(param *openapi3.ParameterRef) bool {
	return len(param.Value.Schema.Value.Enum) > 0
}

func (o *OpenAPIFileContext) ParamIsEnumArray(param *openapi3.ParameterRef) bool {
	return param.Value.Schema.Value.Items != nil && len(param.Value.Schema.Value.Items.Value.Enum) > 0
}

func (o *OpenAPIFileContext) SchemaIsEnum(schema *openapi3.SchemaRef) bool {
	return len(schema.Value.Enum) > 0
}

func (o *OpenAPIFileContext) SchemaIsEnumArray(schema *openapi3.SchemaRef) bool {
	return schema.Value.Items != nil && len(schema.Value.Items.Value.Enum) > 0
}

func (o *OpenAPIFileContext) SchemaContainsAllOf(schema *openapi3.SchemaRef) bool {
	return schema != nil && len(schema.Value.AllOf) > 0
}

func (o *OpenAPIFileContext) SchemaIsComplex(schema *openapi3.SchemaRef) bool {
	if schema == nil || schema.Ref != "" {
		return false
	}

	if schema.Value.Type.Is("object") {
		return true
	}

	if len(schema.Value.AllOf) > 0 {
		return true
	}

	if !schema.Value.Type.Is("array") {
		return false
	}

	return schema.Value.Items.Ref == "" && schema.Value.Items.Value.Type.Is("object")
}

func (o *OpenAPIFileContext) SchemaIsObject(schema *openapi3.SchemaRef) bool {
	return schema.Value.Type.Is("object") || schema.Value.Type.Is("string") // to catch timestamps and uuids
}

func (o *OpenAPIFileContext) GetOpHappyResponseKey(op *openapi3.Operation) string {
	// passing "" as pkg because here we only need the Key part for which pkg is not needed
	opResponse := o.GetOpHappyResponse("", op)
	return opResponse.Key
}

func (o *OpenAPIFileContext) GetOpHappyResponseMimeType(op *openapi3.Operation) string {
	// passing "" as pkg because here we only need the MimeType part for which pkg is not needed
	opResponse := o.GetOpHappyResponse("", op)
	return opResponse.String()
}

func (o *OpenAPIFileContext) GetOpHappyResponseType(pkg string, op *openapi3.Operation) string {
	opResponse := o.GetOpHappyResponse(pkg, op)
	return opResponse.GoType
}

func (o *OpenAPIFileContext) GetOpHappyResponseHeaders(pkg string, op *openapi3.Operation) []string {
	opResponse := o.GetOpHappyResponse(pkg, op)
	return opResponse.Headers
}

/* Auth Focused Helpers */

func (o *OpenAPIFileContext) OpSecurity(op *openapi3.Operation) openapi3.SecurityRequirements {
	if op.Security != nil {
		return *op.Security
	}

	return o.API.Security
}

func hasAuthorization(security openapi3.SecurityRequirements) bool {
	for _, ss := range security {
		for _, scopes := range ss {
			if len(scopes) > 0 {
				return true
			}
		}
	}

	return false
}

func (o *OpenAPIFileContext) HasAuthentication() bool {
	return o.API.Components != nil && o.API.Components.SecuritySchemes != nil && len(o.API.Components.SecuritySchemes) > 0
}

func (o *OpenAPIFileContext) HasAuthorization() bool {
	if hasAuthorization(o.API.Security) {
		return true
	}

	for _, p := range o.API.Paths.InMatchingOrder() {
		path := o.API.Paths.Value(p)
		for _, op := range path.Operations() {
			if op.Security != nil && hasAuthorization(*op.Security) {
				return true
			}
		}
	}

	return false
}

func (o *OpenAPIFileContext) IsSimpleAuth(op *openapi3.Operation) bool {
	s := o.OpSecurity(op)
	if len(s) == 0 {
		return true
	}

	var authName *string

	isDifferentAuth := func(key string) bool {
		if authName == nil {
			authName = &key
			return false
		}

		return *authName != key
	}

	for _, group := range s {
		if len(group) == 0 {
			if isDifferentAuth("") {
				return false
			}
		}

		for key := range group {
			if isDifferentAuth(key) {
				return false
			}
		}
	}

	return true
}

func (o *OpenAPIFileContext) HasComplexAuth() bool {
	for _, p := range o.API.Paths.InMatchingOrder() {
		path := o.API.Paths.Value(p)
		for _, op := range path.Operations() {
			if !o.IsSimpleAuth(op) {
				return true
			}
		}
	}

	return false
}

func (o *OpenAPIFileContext) HasBasicAuth() bool {
	for _, ss := range o.API.Components.SecuritySchemes {
		if ss != nil && ss.Value != nil && ss.Value.Scheme == "basic" {
			return true
		}
	}

	return false
}

func (o *OpenAPIFileContext) HasBearerAuth() bool {
	for _, ss := range o.API.Components.SecuritySchemes {
		if ss != nil && ss.Value != nil && ss.Value.Scheme == "bearer" {
			return true
		}
	}

	return false
}

func (o *OpenAPIFileContext) HasAnyAuth(op *openapi3.Operation) bool {
	s := o.OpSecurity(op)
	if len(s) == 0 {
		return false
	}

	for _, group := range s {
		if len(group) > 0 {
			return true
		}
	}

	return false
}

func (o *OpenAPIFileContext) RequiresAuthUser(op *openapi3.Operation) bool {
	s := o.OpSecurity(op)
	if len(s) == 0 {
		return false
	}

	for _, group := range s {
		if len(group) == 0 {
			return false
		}
	}

	return true
}
