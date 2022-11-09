package output

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gofoji/foji/cfg"
	"github.com/gofoji/foji/input/openapi"
	"github.com/rs/zerolog"
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
	return o.CheckPackage(o.RefToName(s.Ref), pkg)
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

func (o *OpenAPIFileContext) GetType(pkg, name string, s *openapi3.SchemaRef) string {
	xPkg, ok := s.Value.Extensions["x-package"]
	if ok {
		customPkg := getExtAsString(xPkg)
		if customPkg == "" {
			return fmt.Sprint("INVALID x-pacakge: ", xPkg.(string))
		}

		pkg = customPkg
	}

	override, ok := s.Value.Extensions["x-go-type"]
	if ok {
		typeName := getExtAsString(override)
		if typeName != "" {
			return o.CheckPackage(typeName, pkg)
		}

		return fmt.Sprint("INVALID x-go-type: ", override.(string))
	}

	if s.Value.Type == "array" {
		r := s.Value.Items.Ref
		if r != "" {
			return "[]" + o.CheckPackage(o.RefToName(r), pkg)
		}

		return "[]" + o.GetType(pkg, name, s.Value.Items)
	}

	if s.Value.Type == "object" {
		// TODO: Nested anonymous structs
		if s.Ref != "" {
			return o.GetTypeName(pkg, s)
		}
	}

	t, ok := o.Maps.Type[name]
	if ok {
		return o.CheckPackage(t, pkg)
	}

	if s.Value.Format != "" {
		t, ok = o.Maps.Type[s.Value.Type+","+s.Value.Format]
		if ok {
			return o.CheckPackage(t, pkg)
		}
	}

	if s.Value.Type == "" {
		if s.Ref != "" {
			return o.CheckPackage(o.RefToName(s.Ref), pkg)
		}

		return "interface{}"
	}

	t, ok = o.Maps.Type[s.Value.Type]
	if ok {
		return o.CheckPackage(t, pkg)
	}

	return fmt.Sprintf("UNKNOWN:name(%s):ref(%s):type(%s)", name, s.Ref, s.Value.Type)
}

func (o *OpenAPIFileContext) Init() error {
	o.AbortError = nil
	o.CheckAllTypes()

	return nil
}

func (o *OpenAPIFileContext) CheckAllTypes() {
	pkg, _ := o.Params.HasString("PackageName")

	for _, s := range o.API.Components.Schemas {
		for key, schema := range s.Value.Properties {
			o.GetType(pkg, key, schema)
		}
	}
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

func (o *OpenAPIFileContext) GetOpHappyResponse(pkg string, op *openapi3.Operation) OpResponse {
	// TODO: Figure out if the mime type can be extracted from the openapi3.MediaType type
	supportedResponseContentTypes := [3]string{"application/json", "application/jsonl", ""}

	for key, r := range op.Responses {
		if len(key) == 3 && key[0] == '2' {
			for _, mimeType := range supportedResponseContentTypes {
				if mimeType == "" {
					return OpResponse{Key: key, MimeType: "", MediaType: nil, GoType: ""}
				}
				mediaType := r.Value.Content.Get(mimeType)
				if mediaType != nil {
					t := o.GetType(pkg, op.OperationID+"."+key, mediaType.Schema)
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

	return OpResponse{Key: "", MimeType: "", MediaType: nil, GoType: ""}
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
