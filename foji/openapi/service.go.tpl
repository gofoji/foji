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
	{{- range ($.GetOpHappyResponseHeaders $package .RuntimeParams.op) }}string, {{ end -}}
	error)
{{- end -}}

package {{ .PackageName }}

import (
	"context"

{{- .CheckAllTypes .PackageName ($.Params.GetWithDefault "Auth" "") -}}
{{ range .GoImports }}
	"{{ . }}"
{{- end }}
)

type Error string

func (e Error) Error() string {
	return string(e)
}

const ErrNotImplemented = Error("not implemented")

// New creates a new service instance.
func New() *Service {
	return &Service{}
}

// Service implements all business logic for {{ .PackageName }}.
type Service struct {
}

{{- range $name, $path := .File.API.Paths.Map }}
	{{- range $verb, $op := $path.Operations }}

{{ goDoc (pascal $op.OperationID) }}
{{- goDoc $op.Summary }}
{{- goDoc $op.Description }}
func (s *Service) {{ pascal $op.OperationID}}(ctx context.Context,
	{{- template "methodSignature" ($.WithParams "op" $op "path" $path "package" $.PackageName) }}{
	{{- $response := $.GetOpHappyResponseType $.PackageName $op}}
    return 
	{{- if notEmpty $response -}}nil, {{ end -}}
    {{- range ($.GetOpHappyResponseHeaders $.PackageName $op) -}}"", {{ end -}} 
	ErrNotImplemented
}
	{{- end }}
{{- end }}