{{- define "methodSignature"}}
	{{- if not (empty ($.OpSecurity .RuntimeParams.op)) -}}
		user *{{ $.CheckPackage $.Params.Auth $.Params.Package -}},
	{{- end }}
	{{- range $param := .RuntimeParams.op.Parameters -}}
		{{ camel $param.Value.Name }}  {{ if and (not $param.Value.Required) (not (eq $param.Value.Schema.Value.Type "array")) }}*{{ end }}{{ $.GetType "" $param.Value.Name $param.Value.Schema }},
	{{- end -}}
	{{- $body := .RuntimeParams.op.RequestBody}}
	{{- if isNotNil $body}}
        {{- $jsonBody := (index $body.Value.Content "application/json")}}
        {{- if isNotNil $jsonBody}}
            {{- $bodySchema := $jsonBody.Schema}}
            {{- if notEmpty $bodySchema.Ref }}

            {{- $type := $.GetTypeName $.PackageName $bodySchema}}
            {{- camel $type}} *{{ $type -}}
            {{- else }}

            {{- $type := $.GetType $.PackageName "" $jsonBody.Schema}}
            {{- camel $type}} *{{ $type -}}
            {{- end }}
        {{- end -}}
	{{- end -}}
	) (
	{{- $response := $.GetOpHappyResponseType $.PackageName .RuntimeParams.op}}
	{{- if notEmpty $response}} {{$response}},
	{{- end -}}
	error)
{{- end -}}

package {{ .PackageName }}

import "net/http"

// NewServiceImpl New creates a new service instance.
func NewServiceImpl() *ServiceImpl {
	return &ServiceImpl{}
}

// ServiceImpl implements all business logic for {{ .PackageName }}.
type ServiceImpl struct {
}

{{- range $name, $path := .File.API.Paths }}
	{{- range $verb, $op := $path.Operations }}

{{ goDoc (print (pascal $op.OperationID) " " $op.Description) }}.
func (s *ServiceImpl) {{ pascal $op.OperationID}}(r *http.Request,
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
