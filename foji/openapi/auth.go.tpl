{{ .NotNeededIf (not .HasAuthentication) "no security schemes" -}}
{{ .ErrorIf (empty $.Params.Auth) "params.Auth" -}}
// Code generated by foji {{ version }}, template: {{ templateFile }}; DO NOT EDIT.
{{ $packageName := $.PackageName }}
package {{$packageName}}

import (
	"context"
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/bir/iken/httputil"
{{- .CheckAllTypes $packageName $.Params.Auth -}}
{{- range .GoImports }}
	"{{ . }}"
{{- end }}
)

type (
	// AuthenticateFunc is the signature of a function used to authenticate an http request.
	// Given a request, it returns the authenticated user.  If unable to authenticate the
	// request it returns an error.
	AuthenticateFunc = httputil.AuthenticateFunc[*{{ $.CheckPackage $.Params.Auth $packageName }}]

{{if .HasComplexAuth }}
	SecurityGroup  = httputil.SecurityGroup[*{{ $.CheckPackage $.Params.Auth $packageName }}]
	SecurityGroups = httputil.SecurityGroups[*{{ $.CheckPackage $.Params.Auth $packageName }}]
{{- end}}

{{- if .HasAuthorization }}
	AuthorizeFunc  = httputil.AuthorizeFunc[*{{ $.CheckPackage $.Params.Auth $packageName }}]
{{- end}}


// Authenticator takes a key (for example a bearer token) and returns the authenticated user.
	Authenticator = func(ctx context.Context, key string) (*{{ $.CheckPackage $.Params.Auth $packageName }}, error)

{{- if .HasBasicAuth }}

// BasicAuthenticator takes a user/pass pair and returns the authenticated user.
type BasicAuthenticator = func(user,pass string) (*{{ $.CheckPackage $.Params.Auth $packageName }}, error)
{{- end}}

)

{{- if or .HasBasicAuth .HasBearerAuth }}
var (
{{- if .HasBasicAuth }}
    basicAuthPrefix = "Basic "
{{- end}}
{{- if .HasBearerAuth }}
    bearerAuthPrefix = "Bearer "
{{- end}}
)
{{- end}}

{{- range $security, $value := .File.API.Components.SecuritySchemes }}
// {{ pascal $security }}Auth is responsible for extracting "{{$security}}" credentials from a request and calling the
// supplied Authenticator to authenticate
{{- goDoc $value.Value.Description }}
func {{ pascal $security }}Auth(fn {{if eq $value.Value.Scheme "basic"}}Basic{{end}}Authenticator) AuthenticateFunc {
	return func(r *http.Request) (*{{ $.CheckPackage $.Params.Auth $packageName }}, error) {
    {{- if eq $value.Value.Type "apiKey" }}
        {{- if eq $value.Value.In "query" }}
		key := r.URL.Query().Get("{{$value.Value.Name}}")
		if len(key) == 0 {
			return nil, httputil.ErrUnauthorized
		}
        {{- else if eq $value.Value.In "header" }}
		key := r.Header.Get("{{ $value.Value.Name }}")
		if len(key) == 0 {
			return nil, httputil.ErrUnauthorized
		}
        {{- else if eq $value.Value.In "cookie" }}
		cookie := r.Cookie("{{$value.Value.Name}}")
		if cookie == nil || len(cookie.Value) == 0 {
			return nil, httputil.ErrUnauthorized
		}

		key := cookie.Value
        {{end}}

		return fn(r.Context(), key)
    {{- else if eq $value.Value.Type "http" }}
        {{- if eq $value.Value.Scheme "bearer" }}
		key := r.Header.Get("Authorization")
		if len(key) == 0 {
			return nil, httputil.ErrUnauthorized
		}

		if strings.HasPrefix(key, bearerAuthPrefix){
			key = key[7:]
		}

		return fn(r.Context(), key)
        {{- else if eq $value.Value.Scheme "basic" }}
		key := r.Header.Get("Authorization")
		if len(key) == 0 {
			return nil, httputil.ErrBasicAuthenticate
		}

		payload, err := base64.StdEncoding.DecodeString(key[len(basicAuthPrefix):]))
		if err != nil {
			return nil, httputil.ErrBasicAuthenticate
		}

		pair := strings.SplitN(payload, []byte(":"), 2)
		if len(pair) != 2 {
			return nil, httputil.ErrUnauthorized
		}

		return fn(r.Context(), pair[0],pair[1])
        {{- end }}
    {{- else  }}
        // TODO: Support: {{ toJson $value }}
		return nil, httputil.ErrUnauthorized
    {{- end }}
	}
}
{{- end }}
