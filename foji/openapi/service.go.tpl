{{- define "methodSignature"}}
	{{- if not (empty ($.OpSecurity .RuntimeParams.op)) -}}
		user *{{ $.CheckPackage $.Params.Auth $.Params.Package -}},
	{{- end }}
	{{- range $param := .RuntimeParams.op.Parameters -}}
		{{ camel $param.Value.Name }}  {{ if and (not $param.Value.Required) (not (eq $param.Value.Schema.Value.Type "array")) }}*{{ end }}{{ $.GetType "" $param.Value.Name $param.Value.Schema }},
	{{- end -}}
	{{- if isNotNil .RuntimeParams.op.RequestBody}}
		{{- $type := $.GetType .PackageName "" (index  .RuntimeParams.op.RequestBody.Value.Content "application/json").Schema}}
		{{- camel $type}} {{ $type -}}
	{{- end -}}
	) (
	{{- $response := $.GetOpHappyResponseType .PackageName .RuntimeParams.op}}
	{{- if notEmpty $response}} {{$response}},
	{{- end -}}
	error)
{{- end -}}

package {{ .PackageName }}

import (
	"time"

	"github.com/bir/iken/errs"
	"github.com/bir/iken/validation"
	"github.com/valyala/fasthttp"
)

type ServiceError string

func (e ServiceError) Error() string {
	return string(e)
}

const ErrNotImplemented = ServiceError("not implemented")

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