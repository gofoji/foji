{{- define "methodSignature"}}
	{{- if not (empty ($.OpSecurity .RuntimeParams.op)) -}}
		user *{{ $.CheckPackage $.Params.Auth "http" -}},
	{{- end }}
	{{- range $param := .RuntimeParams.op.Parameters -}}
		{{ camel $param.Value.Name }}  {{ if and (not $param.Value.Required) (not (eq $param.Value.Schema.Value.Type "array")) }}*{{ end }}{{ $.GetType "" $param.Value.Name $param.Value.Schema }},
	{{- end -}}
	{{- if isNotNil .RuntimeParams.op.RequestBody}}
		{{- $type := $.GetType "" "" (index  .RuntimeParams.op.RequestBody.Value.Content "application/json").Schema}}
		{{- camel $type}} {{ $type -}}
	{{- end -}}
	) (
	{{- $response := $.GetOpHappyResponseType "http" .RuntimeParams.op}}
	{{- if notEmpty $response}} {{$response}},
	{{- end -}}
	 error)
{{- end -}}

{{- define "paramExtractionFunc" -}}
	{{- $source := .RuntimeParams.source }}
	{{- $type := .RuntimeParams.type }}
	{{- pascal $source -}}
	{{- if eq $type "[]string" -}}Strings
	{{- else }}{{pascal $type}}{{end -}}
{{- end -}}

{{- define "paramExtraction" -}}
	{{- $param := .RuntimeParams.param }}
	{{- $goType := ($.GetType "" $param.Value.Name $param.Value.Schema) }}
	{{- $required := $param.Value.Required }}
		{{- if $required }}

	{{ camel $param.Value.Name }}, err := fastutil.{{template "paramExtractionFunc" ($.WithParams "source" $param.Value.In "type" $goType )}}(ctx, "{{ $param.Value.Name }}")
	if err != nil {
		h.errorHandler(ctx, validation.New("{{ $param.Value.Name }}", err.Error()))
		return
	}
		{{- else }}
	{{ camel $param.Value.Name }} := fastutil.{{template "paramExtractionFunc" ($.WithParams "source" $param.Value.In "type" $goType )}}Optional(ctx, "{{ $param.Value.Name }}")
		{{- end }}
{{- end -}}

// Code generated by foji {{ version }}, template: {{ templateFile }}; DO NOT EDIT.

package http

import (
	"context"
	"encoding/json"

	"github.com/bir/iken/fastctx"
	"github.com/bir/iken/fastutil"
	"github.com/bir/iken/validation"
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
	"{{ .Params.Package }}"
)

type Service interface {
{{- range $name, $path := .API.Paths }}
	{{- range $verb, $op := $path.Operations }}
	{{ pascal $op.OperationID}}(ctx context.Context,
	{{- template "methodSignature" ($.WithParams "op" $op) }}
	{{- end }}
{{- end }}
}

type OpenAPIHandlers struct {
	service      Service
	errorHandler fastutil.ErrorHandlerFunc
	{{- range $security, $value := .API.Components.SecuritySchemes }}
	{{ camel $security }}Auth HttpAuthFunc
	{{- end }}
	{{- if .HasAuthorization }}
	authorize Authorizer
	{{- end}}
{{- range $name, $path := .API.Paths }}
	{{- range $verb, $op := $path.Operations }}
		{{- if not ($.IsSimpleAuth $op) }}
	{{ camel $op.OperationID}}AuthGroups SecurityGroups
		{{- end}}
	{{- end}}
{{- end}}
}

func RegisterHTTP(svc Service, r *router.Router, e fastutil.ErrorHandlerFunc
{{- $hasSecurity := false -}}
{{- range $security, $value := .API.Components.SecuritySchemes -}}
	, {{ camel $security }}Auth
{{- $hasSecurity = true -}}
{{- end -}}
{{- if $hasSecurity }} HttpAuthFunc {{- end -}}
{{- if .HasAuthorization -}}
	, authorize Authorizer
{{- end -}}
) *OpenAPIHandlers {
    s := OpenAPIHandlers{service: svc, errorHandler: e
{{- range $security, $value := .API.Components.SecuritySchemes -}}
	,{{ camel $security }}Auth: {{ camel $security }}Auth
{{- end -}}
	{{- if .HasAuthorization }}, authorize: authorize{{- end -}}
}

{{- range $name, $path := .API.Paths }}

	{{ range $verb, $op := $path.Operations }}
		r.{{upper $verb}}("{{$name}}", s.{{ pascal $op.OperationID}})
	{{- end }}
{{- end }}

{{- range $name, $path := .API.Paths }}
	{{ range $verb, $op := $path.Operations }}
		{{- if not ($.IsSimpleAuth $op) }}
			s.{{ camel $op.OperationID}}AuthGroups = NewSecurityGroups(
			{{- $securityList := $.OpSecurity $op }}
			{{- range $securityGroup := $securityList -}}
				SecurityGroup{}
				{{- range $security, $scopes := $securityGroup -}}
					.Add("{{camel $security}}",{{camel $security}}Auth
					{{- if not (empty $scopes) -}}
						{{- range $scopes -}}
						, "{{.}}"
						{{- end -}}
					{{- end -}}
					)
				{{- end -}}
			{{- end -}})
		{{- end}}
	{{- end}}
{{- end}}

	return &s
}

func (h *OpenAPIHandlers) doJSONWrite(ctx *fasthttp.RequestCtx, code int, obj interface{}) {
	if err := fastutil.JSONWrite(ctx, code, obj); err != nil {
		h.errorHandler(ctx, err)
	}
}

{{- range $name, $path := .API.Paths }}
	{{- range $verb, $op := $path.Operations }}

func (h *OpenAPIHandlers) {{ pascal $op.OperationID}}(ctx *fasthttp.RequestCtx) {
{{- $securityList := $.OpSecurity $op }}
	var (
{{- if not (empty $securityList)}}
		authUser *{{ $.CheckPackage $.Params.Auth "http" }}
{{- end}}
		err error
	)

	fastctx.SetOp(ctx, "{{$op.OperationID}}")

{{- if $.IsSimpleAuth $op }}

{{- $lastAuth := "" }}
	{{- range $securityGroup := $securityList }}
		{{- range $security, $scopes := $securityGroup }}
			{{- if eq $lastAuth $security }}
			{{- else }}

	authUser, err = h.{{ camel $security }}Auth(ctx)
	{{- $lastAuth = $security }}
			{{- end}}
		{{- end}}
	{{- end}}

{{- $authCt := 0 }}
	{{- range $securityGroup := $securityList }}
		{{- range $security, $scopes := $securityGroup }}
			{{- if not (empty $scopes) }}
				{{- if eq $authCt 0 }}
	if err == nil {
				{{- else }}
	if err != nil {
				{{- end }}
		err = h.authorize(authUser{{range $scopes}}, "{{.}}"{{end}})
	{{- $authCt = inc $authCt }}
			{{- end}}
		{{- end}}
	{{- end}}
{{ repeat $authCt "}\n" }}
{{- else }}
	authUser, err = doAuthorize(ctx, {{- if $.HasAuthorization }}h.authorize,{{end}} h.{{ camel $op.OperationID}}AuthGroups...)
{{- end}}
{{- if $.HasAnyAuth $op }}

	if err != nil {
		h.errorHandler(ctx, err)
		return
	}

	{{- if $.RequiresAuthUser $op }}
	if authUser == nil {
		h.errorHandler(ctx, fastutil.ErrUnauthorized)
		return
	}
	{{- end}}
{{- end}}

		{{- range $param := $op.Parameters }}
			{{- template "paramExtraction" ($.WithParams "param" $param) }}
		{{- end }}

	{{- $hasBody := isNotNil $op.RequestBody}}
	{{- if $hasBody }}

	body := {{ $.GetType "" "" (index $op.RequestBody.Value.Content "application/json").Schema }}{}
{{/*	// TODO: Set Defaults*/}}

	err = json.Unmarshal(ctx.PostBody(), &body)
	if err != nil {
		h.errorHandler(ctx, err)
		return
	}

	err = body.Validate()
	if err != nil {
		h.errorHandler(ctx, err)
		return
	}
	{{- end }}

	{{- $response := $.GetOpHappyResponseType "http" $op}}
	{{- if notEmpty $response}}

	response, err := h.service.{{ pascal $op.OperationID}}(ctx,
	{{- else}}

	err = h.service.{{ pascal $op.OperationID}}(ctx,
	{{- end}}
	{{- if not (empty $securityList) -}}
		authUser,
	{{- end -}}
	{{- range $param := $op.Parameters -}}
		{{ camel $param.Value.Name }},
	{{- end -}}
	{{- if $hasBody -}}
		body
	{{- end -}}

	)
	if err != nil {
		h.errorHandler(ctx, err)
		return
	}

{{- /*	// TODO: ? ctx.Response.Header.Set("Access-Control-Allow-Origin", "*")*/}}
{{- /*	// TODO: Code/Encoding based on response*/}}
	{{- $key := $.GetOpHappyResponseKey $op }}
	{{- if notEmpty $response }}

	h.doJSONWrite(ctx, {{$key}}, response)
	{{- else }}

	ctx.Response.SetStatusCode({{$key}})
	{{- end }}
}
	{{- end }}
{{- end }}