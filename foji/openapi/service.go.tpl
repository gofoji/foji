{{/*{{- define "methodSignature"}}*/}}
{{/*	{{- $op := .RuntimeParams.op -}}*/}}
{{/*	{{- if not (empty ($.OpSecurity $op)) }} user *{{ $.CheckPackage $.Params.Auth .PackageName -}},{{- end }}*/}}
{{/*	{{- range $param := $op.Parameters -}}*/}}
{{/*		{{ camel $param.Value.Name }} {{ if and (not $param.Value.Required) (not (eq $param.Value.Schema.Value.Type "array")) }}*{{ end }}{{ $.GetType "" $param.Value.Name $param.Value.Schema }},*/}}
{{/*	{{- end -}}*/}}
{{/*	{{- if isNotNil $op.RequestBody}}*/}}
{{/*		{{- $index := index .RuntimeParams.op.RequestBody.Value.Content "application/json" -}}*/}}
{{/*		{{- if empty $index -}} MISSING*/}}
{{/*		{{ else }}*/}}
{{/*		{{- $type := $.GetType .PackageName (print $op.OperationID "Request") $index.Schema}}*/}}
{{/*		{{- camel $type}} {{ pascal $type -}}*/}}
{{/*		{{ end }}*/}}
{{/*	{{- end -}}*/}}
{{/*	) (*/}}
{{/*	{{- $response := $.GetOpHappyResponseType .PackageName .RuntimeParams.op}}*/}}
{{/*	{{- if notEmpty $response}}{{ $.CheckPackage $response .PackageName -}}, {{ end }}error)*/}}
{{/*{{- end -}}*/}}
{{- define "methodSignature"}}
    {{- $op := .RuntimeParams.op -}}
    {{- $body := .GetRequestBody $op -}}
    {{- if not (empty ($.OpSecurity $op)) }} user *{{ $.CheckPackage $.Params.Auth .PackageName -}},{{- end }}
    {{- range $param := $op.Parameters -}}
        {{ goToken (camel $param.Value.Name) }} {{ if and (not $param.Value.Required) (not (eq $param.Value.Schema.Value.Type "array")) }}*{{ end }}{{ $.GetType "" $param.Value.Name $param.Value.Schema }},
    {{- end -}}
    {{- if isNotNil $body}}
        {{- $type := $.GetType .PackageName (print $op.OperationID "Request") $body.Schema }} body {{ $type  -}}
    {{- end -}}
	) (
    {{- $response := $.GetOpHappyResponseType .PackageName .RuntimeParams.op}}
    {{- if notEmpty $response}}{{ $.CheckPackage $response .PackageName}}, {{ end }}error)
{{- end -}}


package {{ .PackageName }}

import (
	"context"

{{- .CheckAllTypes .PackageName $.Params.Auth -}}
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

{{- range $name, $path := .File.API.Paths }}
	{{- range $verb, $op := $path.Operations }}

{{ goDoc (print (pascal $op.OperationID) " " $op.Description) }}.
func (s *Service) {{ pascal $op.OperationID}}(ctx context.Context,
	{{- template "methodSignature" ($.WithParams "op" $op) }}{
	{{- $response := $.GetOpHappyResponseType $.PackageName $op}}
	{{- if notEmpty $response }}
	return nil, ErrNotImplemented
	{{- else }}
	return ErrNotImplemented
	{{- end }}
}
	{{- end }}
{{- end }}