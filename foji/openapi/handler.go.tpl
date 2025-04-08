{{- define "methodSignature"}}
    {{- $path := .RuntimeParams.path -}}
    {{- $op := .RuntimeParams.op -}}
    {{- $body := .GetRequestBody $op -}}
    {{- $package := .RuntimeParams.package -}}
    {{- if not (empty ($.OpSecurity $op)) }} user *{{ $.CheckPackage $.Params.Auth $package -}},{{- end }}
    {{- range $param := $.OpParams $path $op -}}
		{{ $typeName := (print $op.OperationID " " $param.Value.Name) -}}
		{{- if notEmpty $param.Ref -}}
			{{- $typeName = trimPrefix "#/components/parameters/" $param.Ref -}}
		{{- end -}}
        {{ goToken (camel $param.Value.Name) -}}
        {{- if $.ParamIsOptionalType $param }} *{{ end }} {{ $.GetType $package $typeName $param.Value.Schema }},
    {{- end -}}
    {{- if isNotNil $body}}
        {{- $type := $.GetType $package (print $op.OperationID " Request") $body.Schema }} body {{ $type  -}}
    {{- end -}}
	) (
    {{- $response := $.GetOpHappyResponseType $package .RuntimeParams.op}}
    {{- if notEmpty $response}}{{ $.CheckPackage $response $package}}, {{ end }}
	{{- if gt (len ($.GetOpHappyResponseHeaders $package .RuntimeParams.op)) 0 }}http.Header, {{ end -}}
	error)
{{- end -}}


{{- define "paramExtraction" -}}
    {{- $op := .RuntimeParams.op }}
    {{- $param := .RuntimeParams.param }}
    {{- $package := .RuntimeParams.package }}
	{{- $isEnum := $.ParamIsEnum $param }}
	{{ $typeName := (print $op.OperationID " " $param.Value.Name) -}}
	{{- if notEmpty $param.Ref -}}
		{{- $typeName = trimPrefix "#/components/parameters/" $param.Ref -}}
	{{- end -}}
    {{- $goType := $.GetType $package $typeName $param.Value.Schema }}
    {{- $enumNew := $.EnumNew $goType }}
    {{- $required := $param.Value.Required }}
    {{- $hasDefault := isNotNil $param.Value.Schema.Value.Default }}
    {{- $isArrayEnum := $.ParamIsEnumArray $param }}
    {{- $getRequiredParamFunction := "" -}}
    {{- if $param.Value.Schema.Value.Type.Is "array" -}}
        {{- if eq $goType "[]int32" -}}
            {{- $getRequiredParamFunction = "GetInt32Array" -}}
        {{- else if $isArrayEnum }}
            {{- $getRequiredParamFunction = "GetEnumArray" -}}
        {{- else }}
            {{- $getRequiredParamFunction = "GetStringArray" -}}
        {{- end -}}
    {{- else -}}
        {{- if eq $goType "bool" -}}
            {{- $getRequiredParamFunction = "GetBool" -}}
        {{- else if eq $goType "int32" -}}
            {{- $getRequiredParamFunction = "GetInt32" -}}
        {{- else if eq $goType "int64" -}}
            {{- $getRequiredParamFunction = "GetInt64" -}}
        {{- else if eq $goType "time.Time" }}
            {{- $getRequiredParamFunction = "GetTime" -}}
		{{- else if eq $goType "uuid.UUID" }}
			{{- $getRequiredParamFunction = "GetUUID" -}}
        {{- else if $isEnum }}
            {{- $getRequiredParamFunction = "GetEnum" -}}
        {{- else }}
            {{- $getRequiredParamFunction = "GetString" -}}
        {{- end -}}
    {{- end -}}
{{- goDoc $param.Value.Description }}
	{{ if $param.Value.Schema.Value.Type.Is "array" }}
	{{ goToken (camel $param.Value.Name) }}, _, err := params.{{ $getRequiredParamFunction }}{{ pascal $param.Value.In }}(r, "{{ $param.Value.Name }}", {{ $required }}
		{{- if $isArrayEnum -}}, {{ $enumNew  }}{{- end -}})
	if err != nil {
		validationErrors.Add("{{ $param.Value.Name }}", err)
	}
		{{ if $hasDefault }}
	if len({{ goToken (camel $param.Value.Name) }}) == 0 {
	    {{ goToken (camel $param.Value.Name) }} = {{$goType}}{
		    {{- range $val := $.DefaultValues  $param.Value.Schema.Value.Default}}
        {{ if $isArrayEnum}}{{$.StripArray $goType}}{{ pascal (goToken $val) }}{{else}}{{ printf "%#v" $param.Value.Schema.Value.Default }}{{end}},
			{{end}}
		}
	}
		{{end}}


	{{- else if $required }}
	{{ goToken (camel $param.Value.Name) }}, _, err := params.{{ $getRequiredParamFunction }}{{ pascal $param.Value.In }}(r, "{{ $param.Value.Name }}", {{ $required }}
		{{- if or $isEnum $isArrayEnum -}}, {{ $enumNew  }}{{- end -}})
	if err != nil {
		validationErrors.Add("{{ $param.Value.Name }}", err)
	}


	{{- else if $hasDefault }}
	{{- if $isEnum -}}
		{{ goToken (camel $param.Value.Name) }}, ok, err := params.{{ $getRequiredParamFunction }}{{ pascal $param.Value.In }}(r, "{{ $param.Value.Name }}", {{ $required }}, {{ $enumNew  }})
	{{else -}}
		{{ goToken (camel $param.Value.Name) }}, ok, err := params.{{ $getRequiredParamFunction }}{{ pascal $param.Value.In }}(r, "{{ $param.Value.Name }}", {{ $required }})
	{{- end }}
	if err != nil {
		validationErrors.Add("{{ $param.Value.Name }}", err)
	} else if !ok {
	    {{ goToken (camel $param.Value.Name) }} = {{ if $isEnum -}}
			{{- $goType}}{{ pascal (goToken (printf "%#v" $param.Value.Schema.Value.Default)) }}
		{{else -}}
			{{- if and (eq $goType "time.Time") (eq $param.Value.Schema.Value.Default "") -}}
                time.Time{}
            {{else -}}
				{{ printf "%#v" $param.Value.Schema.Value.Default }}
			{{- end}}
		{{- end}}
	}


	{{- else }}
	var {{ goToken (camel $param.Value.Name) }} *{{$goType}}

	{{ goToken (camel $param.Value.Name) }}Val, ok, err := params.{{ $getRequiredParamFunction }}{{ pascal $param.Value.In }}(r, "{{ $param.Value.Name }}", {{ $required }}
    {{- if $isEnum -}}, {{ $enumNew  }}{{- end -}})
	if err != nil {
		validationErrors.Add("{{ $param.Value.Name }}", err)
	}

	if ok {
		{{ goToken (camel $param.Value.Name) }} = &{{ goToken (camel $param.Value.Name) }}Val
	}
	{{- end -}}

{{- end -}}
{{- $package := $.PackageName }}

// Code generated by foji {{ version }}, template: {{ templateFile }}; DO NOT EDIT.

package {{ $package }}

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/bir/iken/httputil"
	"github.com/bir/iken/logctx"
	"github.com/bir/iken/params"
	"github.com/bir/iken/validation"
	"{{ $.Params.Package}}"
{{- .CheckAllTypes $package ($.Params.GetWithDefault "Auth" "") -}}
{{- range .GoImports }}
	"{{ . }}"
{{- end }}
)

type Operations interface {
{{- range $name, $path := .API.Paths.Map }}
    {{- range $verb, $op := $path.Operations }}
        {{- $opResponse := $.GetOpHappyResponse $package $op }}
	{{ pascal $op.OperationID}}(
		{{- if ($.OpHasExtension $op "x-raw-request" )}}r *http.Request, {{ else }}ctx context.Context, {{ end }}
		{{- if ($.OpHasExtension $op "x-raw-response" )}} w http.ResponseWriter, {{ end }}
        {{- template "methodSignature" ($.WithParams "op" $op "package" $package "path" $path) }}
    {{- end }}
{{- end }}
}

type OpenAPIHandlers struct {
	ops      Operations
{{- if .HasAuthentication }}
    {{- range $security, $value := .API.Components.SecuritySchemes }}
	{{ camel $security }}Auth AuthenticateFunc
    {{- end }}
{{- if .HasAuthorization }}
    authorize AuthorizeFunc
{{- end}}
{{- end}}
{{- range $name, $path := .API.Paths.Map }}
    {{- range $verb, $op := $path.Operations }}
        {{- if not ($.IsSimpleAuth $op) }}
	{{ camel $op.OperationID}}Security SecurityGroups
        {{- end}}
    {{- end}}
{{- end}}
}

type Mux interface {
    Handle(pattern string, handler http.Handler)
}

func RegisterHTTP(ops Operations, r Mux
{{- if .HasAuthentication }}
	{{- range $security, $value := .API.Components.SecuritySchemes -}}
    , {{ camel $security }}Auth
	{{- end }} AuthenticateFunc
	{{- if .HasAuthorization -}}
    , authorize AuthorizeFunc
	{{- end -}}
{{- end -}}
) *OpenAPIHandlers {
	s := OpenAPIHandlers{ops: ops
{{- if .HasAuthentication }}
	{{- range $security, $value := .API.Components.SecuritySchemes -}}
    , {{ camel $security }}Auth: {{ camel $security }}Auth
	{{- end -}}
	{{- if .HasAuthorization -}}
	, authorize: authorize{{- end -}}
{{- end -}}
}

{{- range $name, $path := .API.Paths.Map }}
    {{- range $verb, $op := $path.Operations }}
        {{- if not ($.IsSimpleAuth $op) }}

	s.{{ camel $op.OperationID}}Security = SecurityGroups{
            {{- $securityList := $.OpSecurity $op }}
            {{- range $securityGroup := $securityList }}
		SecurityGroup{
                {{- range $security, $scopes := $securityGroup -}}
					httputil.NewAuthCheck({{camel $security}}Auth
                    {{- if not (empty $scopes) -}}
						, authorize
                        {{- range $scopes -}}
							, "{{.}}"
                        {{- end -}}
                    {{- else -}}
						, nil
                    {{- end -}}
					),
                {{- end -}}
				},
            {{- end -}}
	}
        {{- end}}
    {{- end}}
{{- end}}

{{ range $name, $path := .API.Paths.Map }}
    {{- range $verb, $op := $path.Operations }}
	r.Handle("{{$verb}} {{$name}}", http.HandlerFunc(s.{{ pascal $op.OperationID}}))
    {{- end }}
{{- end }}

	return &s
}

{{- range $name, $path := .API.Paths.Map }}
    {{- range $verb, $op := $path.Operations }}
        {{- $opResponse := $.GetOpHappyResponse $package $op }}
        {{- $opBody := $.GetRequestBody $op }}

{{ goDoc (pascal $op.OperationID) }}
{{- goDoc $op.Summary }}
{{- goDoc $op.Description }}
func (h OpenAPIHandlers) {{ pascal $op.OperationID}}(w http.ResponseWriter, r *http.Request) {
	var err error
        {{- $securityList := $.OpSecurity $op }}

	logctx.AddStrToContext(r.Context(), "op", "{{$op.OperationID}}")

        {{- if $.IsSimpleAuth $op }}
            {{- $lastAuth := "" }}
            {{- range $securityGroup := $securityList }}
                {{- range $security, $scopes := $securityGroup }}
                    {{- if eq $lastAuth $security }}
                    {{- else }}

	user, err := h.{{ camel $security }}Auth(r)
                        {{- $lastAuth = $security -}}
                    {{- end -}}
                {{- end -}}
            {{- end }}
            {{- $authCt := 0 }}
            {{- range $securityGroup := $securityList }}
                {{- range $security, $scopes := $securityGroup }}
                    {{- if not (empty $scopes) }}
                        {{- if eq $authCt 0 }}
	if err == nil {
                        {{- else }}
	if err != nil {
                        {{- end }}
		err = h.authorize(r.Context(), user, []string{ {{range $scopes}}"{{.}}", {{end}} })
                        {{- $authCt = inc $authCt }}
                    {{- end}}
                {{- end}}
            {{- end}}
            {{- repeat $authCt "}\n" }}
        {{- else }}

	user, err := h.{{ camel $op.OperationID}}Security.Auth(r)
        {{- end -}}
        {{- if $.HasAnyAuth $op }}
	if err != nil {
		httputil.ErrorHandler(w, r, err)

		return
	}
        {{- end}}
        {{- if not (empty ($.OpParams $path $op)) }}

	var validationErrors validation.Errors
        {{- range $param := $.OpParams $path $op }}

	{{ template "paramExtraction" ($.WithParams "param" $param "package" $package "op" $op) }}
        {{- end }}

	if validationErrors != nil {
		httputil.ErrorHandler(w, r, validationErrors.GetErr())

		return
	}
		{{- end}}

        {{- $hasBody := not (empty $opBody)}}
		{{- if $hasBody }}
			{{- $bodyType := $.GetType $package (print $op.OperationID " Request") $opBody.Schema}}
			{{- if $opBody.IsJson }}

	var body {{ $bodyType }}
	if err = httputil.GetJSONBody(r.Body, &body); err != nil {
		httputil.ErrorHandler(w, r, err)

		return
	}
        	{{- end }}
			{{- if $opBody.IsText }}

	b, err := io.ReadAll(r.Body)
	if err != nil {
		httputil.ErrorHandler(w, r, validation.Error{Message:"unable to read body", Source:err})

		return
	}

	body := string(b)
        	{{- end -}}
			{{- if or $opBody.IsForm $opBody.IsMultipartForm }}

	body, err := ParseForm{{ $bodyType }}(r)
    if err != nil {
        httputil.ErrorHandler(w, r, validation.Error{Message:"unable to parse form", Source:err})

        return
    }
        	{{- end -}}
        {{- end }}

        {{ if ($.OpHasExtension $op "x-raw-response" )}}
    ww := httputil.WrapWriter(w)
    w = ww
        {{ end }}

        {{- $responseGoType := $opResponse.GoType}}

        {{ if notEmpty $responseGoType }}response, {{ end -}}
        {{- if gt (len $opResponse.Headers) 0 -}}headers, {{ end -}}
	err {{  if or (notEmpty $responseGoType) (gt (len $opResponse.Headers) 0 ) }}:{{ end }}= h.ops.{{ pascal $op.OperationID}}(
        {{- if ($.OpHasExtension $op "x-raw-request" )}}r, {{ else }}r.Context(), {{ end }}
        {{- if ($.OpHasExtension $op "x-raw-response" )}} w, {{ end }}
        {{- if not (empty $securityList) }} user,{{- end -}}
        {{- range $param := $.OpParams $path $op}} {{ goToken (camel $param.Value.Name) }},{{- end -}}
        {{- if $hasBody }} body{{- end -}}

        )
	if err != nil {
		httputil.ErrorHandler(w, r, err)

		return
	}

	    {{ range $header := $opResponse.Headers }}
	    {{ camel $header }} := headers.Values("{{ $header }}")
	for _, v := range {{ camel $header }} {
		if v != "" {
			w.Header().Add("{{ $header }}", v)
		}
	}
	    {{ end }}

        {{ if ($.OpHasExtension $op "x-raw-response" )}}
	if ww.Status() > 0 {
		return
	}
        {{ end }}

        {{- $key := $.GetOpHappyResponseKey $op }}
        {{- if notEmpty $responseGoType }}
			{{- if $opResponse.IsJson }}

	httputil.JSONWrite(w, r, {{$key}}, response)
			{{- else if $opResponse.IsHTML }}

	httputil.HTMLWrite(w, r, {{$key}}, response)
            {{- else if eq $responseGoType "[]byte" }}

	httputil.Write(w, r, "{{$opResponse.MimeType}}",  {{$key}}, response)
            {{- else if eq $responseGoType "io.Reader" }}

	httputil.ReaderWrite(w, r, "{{$opResponse.MimeType}}",  {{$key}}, response)
			{{- else }}

	httputil.Write(w, r, "{{$opResponse.MimeType}}",  {{$key}}, []byte(response))
			{{- end -}}
        {{- else }}

	w.WriteHeader({{$key}})
        {{- end }}
}

    {{- end }}
{{- end }}
