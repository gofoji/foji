{{- define "propertyDeclaration"}}
    {{- $key := .RuntimeParams.key }}
    {{- $schema := .RuntimeParams.schema }}
    {{- $typeName := .RuntimeParams.typeName }}
    {{- if not (empty $schema.Value.Description) }}
    // {{ $schema.Value.Description }}
    {{- end }}
    {{- if $.IsDefaultEnum  $key $schema }}
    {{ pascal $key }}  {{print $typeName (pascal $key) "Enum"}} `json:"{{$key}},omitempty"`
    {{- else }}
    {{ pascal $key }} {{ $.GetType .PackageName $key $schema }} `json:"{{$key}},omitempty"`
    {{- end }}
{{- end -}}

// Code generated by foji {{ version }}, template: {{ templateFile }}; DO NOT EDIT.

package {{ .PackageName }}

import (
{{- range .Imports }}
    "{{ . }}"
{{- end }}
    "github.com/bir/iken/validation"
)
{{ range $key, $schema := .File.API.Components.Schemas }}
{{- if isNil (index $schema.Value.Extensions "x-go-type" ) }}
{{- $typeName := pascal $key }}
//  {{ $typeName }}
{{- if not (empty .Value.Description) }}
//  {{ $schema.Value.Description }}
{{- end}}
//
// OpenAPI Component: {{ $key }}
type {{ pascal $key }} struct {
{{- range $key, $schema := .Value.Properties }}
    {{- template "propertyDeclaration" ($.WithParams "key" $key "schema" $schema "typeName" $typeName)}}
{{- end }}
{{- range .Value.AllOf }}
    {{- if notEmpty .Ref }}
        {{ $.GetTypeName $.PackageName  . }}
    {{- else }}
        {{- range $key, $schema := .Value.Properties }}
            {{- template "propertyDeclaration" ($.WithParams "key" $key "schema" $schema "typeName" $typeName)}}
        {{- end }}
    {{- end }}
{{- end }}

}
{{- range $key, $schema := .Value.Properties }}
    {{- if notEmpty $schema.Value.Pattern }}

var {{ camel $typeName }}{{ pascal $key }}Pattern = regexp.MustCompile(`{{ $schema.Value.Pattern }}`)
    {{- end}}
{{- end }}

{{- range $key, $schema := .Value.Properties }}
    {{- if not (empty $schema.Value.Enum) }}
    {{- $enumType := print $typeName (pascal $key) "Enum" }}

type {{ $enumType }} int32

const (
    Unknown{{ $enumType }} {{ $enumType }} = iota
    {{- range $i, $value := $schema.Value.Enum }}
    {{ $enumType }}{{ pascal (goToken $value) }}
    {{- end }}
)

func New{{ $enumType }}(name string) {{ $enumType }} {
    switch name {
    {{- range $schema.Value.Enum  }}
        case "{{ . }}":
        return {{ $enumType }}{{ pascal (goToken .) }}
    {{- end }}
    }

    return {{ $enumType }}(0)
}


var  {{ $enumType }}String = map[{{ $enumType }}]string{
    {{- range $schema.Value.Enum }}
        {{ $enumType }}{{ pascal (goToken .) }}: "{{ . }}",
    {{- end }}
}

func (e {{ $enumType }}) String() string {
    return {{ $enumType }}String[e]
}

func (e *{{ $enumType }}) UnmarshalJSON(input []byte) (err error) {
	var i int32

	err = json.Unmarshal(input, &i)
	if err == nil {
		*e = {{ $enumType }}(i)
		return nil
	}

	var s string

	err = json.Unmarshal(input, &s)
	if err != nil {
		return err
	}

	*e = New{{ $enumType }}(s)

	return nil
}

func (e *{{ $enumType }}) MarshalJSON() (result []byte, err error) {
    return json.Marshal(e.String())
}
    {{- end}}
{{- end }}

func (p {{ pascal $key }}) Validate() error {
    {{- if $.HasValidation . }}
    var err validation.Errors
{{- range $key, $schema := .Value.Properties }}
    {{- if in $schema.Value.Type "number" "integer" }}
        {{- $fieldType := $.GetType $.PackageName $key $schema }}
        {{- if isNotNil $schema.Value.Min }}

    if p.{{ pascal $key }} <{{ if $schema.Value.ExclusiveMin }}={{end}} {{ $schema.Value.Min }} {
        _ = err.Add("{{$key}}", "must be >{{ if not $schema.Value.ExclusiveMin }}={{end}} {{ $schema.Value.Min }}")
    }
        {{- end }}
        {{- if isNotNil $schema.Value.Max }}

    if p.{{ pascal $key }} >{{ if $schema.Value.ExclusiveMax }}={{end}} {{ $schema.Value.Max }} {
        _ = err.Add("{{$key}}", "must be <{{ if not $schema.Value.ExclusiveMax }}={{end}} {{ $schema.Value.Max }}")
    }
        {{- end }}
        {{- if isNotNil $schema.Value.MultipleOf }}
            {{- if eq $schema.Value.Type "integer" }}

    if p.{{ pascal $key }} % {{ $schema.Value.MultipleOf }} != 0 {
        _ = err.Add("{{$key}}", "must be multiple of {{ $schema.Value.MultipleOf }}")
    }
            {{- else }}
    if math.Mod({{ if not (eq $fieldType "float64") }}float64({{ end }}p.{{ pascal $key }}{{ if not (eq $fieldType "float64") }}){{end}}, {{ $schema.Value.MultipleOf }}) != 0 {
        _ = err.Add("{{$key}}", "must be multiple of {{ $schema.Value.MultipleOf }}")
    }
            {{- end }}
        {{- end }}
    {{- else if eq $schema.Value.Type "string" }}
        {{- $fieldType := $.GetType $.PackageName $key $schema }}
        {{- if gt $schema.Value.MinLength 0 }}

    if len(p.{{ pascal $key }}) < {{ $schema.Value.MinLength }} {
        _ = err.Add("{{$key}}", "length must be >= {{ $schema.Value.MinLength }}")
    }
        {{- end }}
        {{- if isNotNil $schema.Value.MaxLength }}

    if len(p.{{ pascal $key }}) > {{ $schema.Value.MaxLength }} {
        _ = err.Add("{{$key}}", "length must be <= {{ $schema.Value.MaxLength }}")
    }
        {{- end }}
        {{- if notEmpty $schema.Value.Pattern }}

    if !{{ camel $typeName }}{{ pascal $key }}Pattern.MatchString( p.{{ pascal $key }})  {
        _ = err.Add("{{$key}}", "must match {{ $schema.Value.Pattern }}")
    }
        {{- end }}
    {{- else if eq $schema.Value.Type "array" }}
        {{- if gt $schema.Value.MinItems 0 }}

    if len(p.{{ pascal $key }}) < {{ $schema.Value.MinItems }} {
        _ = err.Add("{{$key}}", "length must be >= {{ $schema.Value.MinItems }}")
    }
        {{- end }}
        {{- if isNotNil $schema.Value.MaxItems }}

    if len(p.{{ pascal $key }}) > {{ $schema.Value.MaxItems }} {
        _ = err.Add("{{$key}}", "length must be <= {{ $schema.Value.MaxItems }}")
    }
        {{- end }}
    {{- end }}
{{- end }}

    return err.GetErr()
{{- else }}
    return nil
{{- end }}
}
{{ end }}
{{- end}}

