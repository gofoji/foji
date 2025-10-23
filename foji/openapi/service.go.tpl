{{- define "methodSignature"}}
    {{- $path := .RuntimeParams.path -}}
    {{- $op := .RuntimeParams.op -}}
    {{- $body := .GetRequestBody $op -}}
    {{- $package := .RuntimeParams.package -}}
    {{- if not (empty ($.OpSecurity $op)) }} user *{{ $.CheckPackage $.Params.Auth $package -}},{{- end }}
    {{- range $param := $.OpParams $path $op -}}
        {{- $name := (print $op.OperationID " " $param.Value.Name) -}}
        {{ goToken (camel $param.Value.Name) -}}
        {{- if notEmpty $param.Ref -}}
            {{- $name = $param.Value.Name -}}
        {{- end -}}
        {{- if $.ParamIsOptionalType $param }} *{{ end }} {{ $.GetType $package $name $param.Value.Schema }},
    {{- end -}}
    {{- if isNotNil $body}}
        {{- $type := $.GetType $package (print $op.OperationID "Request") $body.Schema }} body {{ $type  -}}
    {{- end -}}
	) (
    {{- $response := $.GetOpHappyResponseType $package .RuntimeParams.op}}
    {{- if notEmpty $response}}{{ $.CheckPackage $response $package}}, {{ end }}
	{{- if gt (len ($.GetOpHappyResponseHeaders $package .RuntimeParams.op)) 0 }}http.Header, {{ end -}}
	error)
{{- end -}}

package {{ .PackageName }}

import (
	"context"
	"errors"

{{- .CheckAllTypes .PackageName ($.Params.GetWithDefault "Auth" "") -}}
{{ range .GoImports }}
	"{{ . }}"
{{- end }}
)

// GenService holds all Unsupported mock endpoints.  You can embed the mock service in real code.
// Simple example:
//
//	type Service struct {
//		GenService
//	}
//
// This allows you to mock the service for a quick start and eventually delete this stub.
type GenService struct {
}

{{- range $name, $path := .File.API.Paths.Map }}
	{{- range $verb, $op := $path.Operations }}

{{ goDoc (pascal $op.OperationID) }}
{{- goDoc $op.Summary }}
{{- goDoc $op.Description }}
func (s *GenService) {{ pascal $op.OperationID}}(ctx context.Context,
{{- template "methodSignature" ($.WithParams "op" $op "path" $path "package" $.PackageName) }}{
{{- $response := $.GetOpHappyResponseType $.PackageName $op}}
{{- if notEmpty $response }}
	return nil, errors.ErrUnsupported
{{- else }}
	return errors.ErrUnsupported
{{- end }}}
	{{- end }}
{{- end }}