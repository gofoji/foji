package output

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gofoji/foji/cfg"
	"github.com/gofoji/foji/input/openapi"
	"github.com/sirupsen/logrus"
)

const (
	OpenAPIFile = "OpenAPIFile"
	OpenAPITag  = "OpenAPITag"
)

func HasOpenAPIOutput(o cfg.Output) bool {
	return hasAnyOutput(o, OpenAPIFile, OpenAPITag)
}

func OpenAPI(p cfg.Process, fn cfg.FileHandler, logger logrus.FieldLogger, groups openapi.FileGroups, simulate bool) error {
	for _, ff := range groups {
		for _, f := range ff {
			ctx := OpenAPIFileContext{
				Context: Context{Process: p, Logger: logger},
				File:    f,
			}
			ctx.init()

			err := invokeProcess(p.Output[OpenAPIFile], p.RootDir, fn, logger, &ctx, simulate)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

type OpenAPIFileContext struct {
	Context
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
	// TODO should we do some form of lookup?
	return o.CheckPackage(o.RefToName(s.Ref), pkg)
}

func getExtAsString(in interface{}) string {
	bb, ok := in.(json.RawMessage)
	if !ok {
		return ""
	}

	var s string
	err := json.Unmarshal(bb, &s)
	if err != nil {
		return ""
	}
	return s
}

func (o *OpenAPIFileContext) GetType(pkg, name string, s *openapi3.SchemaRef) string {
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
		if s.Ref != "" {
			return o.GetTypeName(pkg, s)
		}
		// TODO: Nested anonymous structs
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

	if s.Value.Type == "" { // TODO: Should this be handled with mapping?
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

func (o *OpenAPIFileContext) init() {
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

// IsDefaultEnum helper that checks if an enumerated type is overridden (specified externally)
func (o *OpenAPIFileContext) IsDefaultEnum(name string, s *openapi3.SchemaRef) bool {
	if len(s.Value.Enum) == 0 {
		return false
	}

	_, ok := o.Maps.Type[name]
	if ok {
		return false
	}

	return true
}

func (o *OpenAPIFileContext) GetOpHappyResponseType(pkg string, op *openapi3.Operation) string {
	key, content := o.GetOpHappyResponse(op)
	if content == nil {
		return ""
	}

	t := o.GetType(pkg, op.OperationID+"."+key, content.Schema)
	if strings.HasPrefix(t, "[]") {
		return t
	}

	return "*" + t
}

func (o *OpenAPIFileContext) GetOpHappyResponse(op *openapi3.Operation) (string, *openapi3.MediaType) {
	// TODO: Support response types besides JSON
	for key, r := range op.Responses {
		if len(key) == 3 && key[0] == '2' {
			return key, r.Value.Content.Get("application/json")
		}
	}

	return "", nil
}

func (o *OpenAPIFileContext) GetOpHappyResponseKey(op *openapi3.Operation) string {
	key, _ := o.GetOpHappyResponse(op)
	return key
}

func securityNames(ss openapi3.SecurityRequirements) []string {
	var out []string
	for _, security := range ss {
		for key := range security {
			out = append(out, key)
		}
	}

	return out
}

func (o *OpenAPIFileContext) OpSecurity(op *openapi3.Operation) []string {
	if op.Security != nil && len(*op.Security) > 0 {
		return securityNames(*op.Security)
	}

	if len(o.File.API.Security) > 0 {
		return securityNames(o.File.API.Security)
	}

	return nil
}
