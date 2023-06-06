package output

import (
	"encoding/json"
	"fmt"
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
	var out []string

	for _, i := range o.Imports {
		if i == o.PackageName() {
			continue
		}
		out = append(out, i)
	}

	return out
}

func (o *OpenAPIFileContext) WithParams(values ...interface{}) (*OpenAPIFileContext, error) {
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

func getExtAsString(in interface{}) string {
	bb, ok := in.(json.RawMessage)
	if !ok {
		return ""
	}

	var s string
	if err := json.Unmarshal(bb, &s); err != nil {
		return ""
	}

	return s
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

func (o *OpenAPIFileContext) GetType(currentPackage, name string, s *openapi3.SchemaRef) string {
	xPkg, ok := s.Value.Extensions["x-package"]
	if ok {
		customPkg := getExtAsString(xPkg)
		if customPkg == "" {
			return fmt.Sprint("INVALID x-package: ", xPkg.(string))
		}

		currentPackage = customPkg
	}

	if override, ok := s.Value.Extensions["x-go-type"]; ok {
		return o.getXGoType(currentPackage, override)
	}

	if s.Value.Type == "array" {
		return "[]" + o.GetType(currentPackage, name, s.Value.Items)
	}

	if s.Ref != "" {
		return o.GetTypeName(currentPackage, s)
	}

	if s.Value.Type == "object" {
		if len(s.Value.Properties) == 0 {
			return "any"
		}

		// Anonymous Struct
		name = o.PackageName() + "." + kace.Pascal(name)
		return o.CheckPackage(name, currentPackage)
	}

	t, ok := o.Maps.Type[name]
	if ok {
		return o.CheckPackage(t, currentPackage)
	}

	if s.Value.Format != "" {
		t, ok = o.Maps.Type[s.Value.Type+","+s.Value.Format]
		if ok {
			return o.CheckPackage(t, currentPackage)
		}
	}

	if s.Value.Type == "" {
		if s.Ref != "" {
			return o.CheckPackage(o.RefToName(s.Ref), currentPackage)
		}

		return "any"
	}

	t, ok = o.Maps.Type[s.Value.Type]
	if ok {
		return o.CheckPackage(t, currentPackage)
	}

	return fmt.Sprintf("UNKNOWN:name(%s):ref(%s):type(%s)", name, s.Ref, s.Value.Type)
}

func (o *OpenAPIFileContext) Init() error {
	o.AbortError = nil
	o.Imports = nil

	return nil
}

func (o *OpenAPIFileContext) AllComponentSchemas() openapi3.Schemas {
	if o.API.Components == nil {
		return nil
	}

	return o.API.Components.Schemas
}

// CheckAllTypes is a helper to iterate all property references for import requirements.
// This is expected to inject imports for unnecessary packages depending on the template
// generated, the post-processing should remove unused imports.
func (o *OpenAPIFileContext) CheckAllTypes(pkg string, types ...string) string {
	for _, s := range o.AllComponentSchemas() {
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
	for _, p := range s.Value.Properties {
		if hasValidation(p.Value) {
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
	}

	return nil
}

func (o *OpenAPIFileContext) GetRequestBodyLocal(op *openapi3.Operation) *openapi3.SchemaRef {
	if op.RequestBody != nil && op.RequestBody.Ref == "" && op.RequestBody.Value != nil {
		mediaType := op.RequestBody.Value.Content.Get(ApplicationJSON)
		if mediaType != nil && mediaType.Schema.Ref == "" {
			return mediaType.Schema
		}
	}

	return nil
}

func (o *OpenAPIFileContext) GetOpHappyResponse(pkg string, op *openapi3.Operation) OpResponse {
	supportedResponseContentTypes := []string{ApplicationJSON, ApplicationJSONL, TextPlain, TextHTML}

	happyKey := "200"
	for key, r := range op.Responses {
		if len(key) == 3 && key[0] == '2' {
			happyKey = key

			for _, mimeType := range supportedResponseContentTypes {
				mediaType := r.Value.Content.Get(mimeType)
				if mediaType != nil {
					t := o.GetType(pkg, kace.Pascal(op.OperationID)+" Response", mediaType.Schema)

					var goType string
					if strings.HasPrefix(t, "[]") {
						goType = t
					} else {
						goType = "*" + t
					}

					return OpResponse{Key: key, MimeType: mimeType, MediaType: mediaType, GoType: goType}
				}
			}
		}
	}

	return OpResponse{Key: happyKey, MimeType: "", MediaType: nil, GoType: ""}
}

func (o *OpenAPIFileContext) GetOpHappyResponseKey(op *openapi3.Operation) string {
	// passing "" as pkg because here we only need the Key part for which pkg is not needed
	opResponse := o.GetOpHappyResponse("", op)
	return opResponse.Key
}

func (o *OpenAPIFileContext) GetOpHappyResponseMimeType(op *openapi3.Operation) string {
	// passing "" as pkg because here we only need the MimeType part for which pkg is not needed
	opResponse := o.GetOpHappyResponse("", op)
	return opResponse.MimeType
}

func (o *OpenAPIFileContext) GetOpHappyResponseType(pkg string, op *openapi3.Operation) string {
	opResponse := o.GetOpHappyResponse(pkg, op)
	return opResponse.GoType
}

/* Auth Focused Helpers */

func (o *OpenAPIFileContext) OpSecurity(op *openapi3.Operation) openapi3.SecurityRequirements {
	if op.Security != nil {
		return *op.Security
	}

	return o.File.API.Security
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

	for _, p := range o.API.Paths {
		for _, op := range p.Operations() {
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

	authName := ""

	for _, group := range s {
		for key := range group {
			if authName == "" {
				authName = key
			} else if authName != key {
				return false
			}
		}
	}

	return true
}

func (o *OpenAPIFileContext) HasComplexAuth() bool {
	for _, p := range o.API.Paths {
		for _, op := range p.Operations() {
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
		for key := range group {
			if key != "" {
				return true
			}
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
