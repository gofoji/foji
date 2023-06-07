{{- define "propertyDeclaration"}}
    {{- $key := .RuntimeParams.key }}
    {{- $schema := .RuntimeParams.schema }}
    {{- $typeName := .RuntimeParams.typeName }}
    {{- goDoc $schema.Value.Description }}
    {{- if $.IsDefaultEnum  $key $schema }}
    {{ pascal $key }}  {{ $typeName }}{{ pascal $key }}Enum
    {{- else }}
    {{ pascal $key }} {{ $.GetType .PackageName (print $typeName " " $key) $schema }}
    {{- end }}  `json:"{{$key}},omitempty"`
{{- end -}}

{{- define "enum"}}
{{- $enumType := .RuntimeParams.Type }}
{{- $enums := .RuntimeParams.Values }}

type {{ $enumType }} int8

const (
    Unknown{{ $enumType }} {{ $enumType }} = iota
    {{- range $i, $value := $enums }}
    {{ $enumType }}{{ pascal (goToken $value) }}
    {{- end }}
)

func New{{ $enumType }}(name string) {{ $enumType }} {
    switch name {
    {{- range $enums  }}
    case "{{ . }}":
        return {{ $enumType }}{{ pascal (goToken .) }}
    {{- end }}
    }

    return {{ $enumType }}(0)
}

var  {{ $enumType }}String = map[{{ $enumType }}]string{
    {{- range $enums }}
        {{ $enumType }}{{ pascal (goToken .) }}: "{{ . }}",
    {{- end }}
}

func (e {{ $enumType }}) String() string {
    return {{ $enumType }}String[e]
}

func (e *{{ $enumType }}) UnmarshalJSON(input []byte) (err error) {
	var i int8

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

func (e *{{ $enumType }}) MarshalJSON() ([]byte, error) {
    return json.Marshal(e.String())
}
{{- end -}}

{{- define "typeDeclaration"}}
{{ $schema := .RuntimeParams.schema }}
{{- $key := .RuntimeParams.key }}
{{- $label := .RuntimeParams.label }}

{{- if not ($.HasExtension $schema "x-go-type" )}}
{{- $typeName := $.GetType $.PackageName $key $schema }}
// {{ $typeName}}
{{- goDoc $schema.Value.Description }}
//
// OpenAPI {{$label}}: {{ $key }}

    {{- if in $schema.Value.Type "object" "" }}
type {{ pascal $key }} struct {
    {{- range $key, $schema := $schema.Value.Properties }}
        {{- template "propertyDeclaration" ($.WithParams "key" $key "schema" $schema "typeName" $typeName)}}
    {{- end }}
    {{- range $schema.Value.AllOf }}
        {{- if notEmpty .Ref }}

    // OpenAPI Ref: {{ .Ref }}
    {{ $.GetType $.PackageName "" . }}
        {{- else }}
            {{- range $key, $schema := .Value.Properties }}
                {{- template "propertyDeclaration" ($.WithParams "key" $key "schema" $schema "typeName" $typeName)}}
            {{- end }}
        {{- end }}
    {{- end }}
}
    {{- else }}
type {{ pascal $key }} {{ $.GetType $.PackageName (pascal (print $typeName " Item" )) $schema }}
    {{- end }}

{{- /* Nested Types */}}
    {{- range $key, $schema := $schema.Value.Properties }}
        {{- if empty $schema.Ref -}}
        {{- if not (empty $schema.Value.Properties )}}
            {{- template "typeDeclaration" ($.WithParams "key" (pascal (print $typeName " " $key)) "schema" $schema "label" (print  $typeName " inline " $key))}}
            {{- else if eq $schema.Value.Type "array"}}
                {{- if empty $schema.Value.Items.Ref -}}
                    {{- if not (empty $schema.Value.Items.Value.Properties )}}
                        {{- template "typeDeclaration" ($.WithParams "key" (pascal (print $typeName " " $key)) "schema" $schema.Value.Items "label" (print  $typeName " inline item " $key))}}
                    {{- end }}
                {{- end }}
            {{- end }}
        {{- end }}
    {{- end }}
    {{- /* Nested Arrays */}}
    {{- if empty $schema.Ref -}}
        {{- if eq $schema.Value.Type "array"}}
            {{- if empty $schema.Value.Items.Ref -}}
                {{- if not (empty $schema.Value.Items.Value.Properties )}}
                    {{- template "typeDeclaration" ($.WithParams "key" (pascal (print $typeName " Item" )) "schema" $schema.Value.Items "label" (print  $typeName " inline item " $key))}}
                {{- end }}
            {{- end }}
        {{- end }}
    {{- end }}

{{- /*    Regex Validation Patterns */ -}}
    {{- range $key, $schema := $schema.Value.Properties }}
        {{- if notEmpty $schema.Value.Pattern }}

var {{ camel $typeName }}{{ pascal $key }}Pattern = regexp.MustCompile(`{{ $schema.Value.Pattern }}`)
        {{- end}}
    {{- end }}

{{- /*    Enums */}}
    {{- range $key, $schema := $schema.Value.Properties }}
        {{- if not (empty $schema.Value.Enum) }}
            {{- $enumType := print $typeName (pascal $key) "Enum" }}
            {{- template "enum" ($.WithParams "Type" $enumType "Values" $schema.Value.Enum)}}
        {{- end -}}
    {{- end -}}

    {{- range $schema.Value.AllOf }}
        {{- range $key, $schema := $schema.Value.Properties }}
            {{- if not (empty $schema.Value.Enum) }}
                {{- $enumType := print $typeName (pascal $key) "Enum" }}
                {{- template "enum" ($.WithParams "Type" $enumType "Values" $schema.Value.Enum)}}
            {{- end -}}
        {{- end -}}
    {{- end }}

    {{- if $.HasValidation $schema }}
func (p {{ pascal $key }}) Validate() error {
    var err validation.Errors
        {{- range $key, $schema := $schema.Value.Properties }}
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
        _ = err.Add("{{$key}}", `must match "{{ $schema.Value.Pattern }}"`)
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
}
    {{- end -}}
{{- end -}}
{{- end -}}

// Code generated by foji {{ version }}, template: {{ templateFile }}; DO NOT EDIT.

package {{ .PackageName }}

import (
    "regexp"

{{- .CheckAllTypes .PackageName -}}
{{ range .GoImports }}
    "{{ . }}"
{{- end }}

    "github.com/bir/iken/validation"
)

{{/* Components */}}
{{- range $key, $schema := .AllComponentSchemas }}
    {{- template "typeDeclaration" ($.WithParams "key" $key "schema" $schema "label" "Component")}}
{{- end }}

{{- /* Local Request/Reponse Types */ -}}
{{- range $name, $path := .API.Paths }}
    {{- range $verb, $op := $path.Operations }}
        {{- $bodySchema := $.GetRequestBodyLocal $op}}
        {{- if isNotNil $bodySchema}}
            {{- template "typeDeclaration" ($.WithParams "key" (print $op.OperationID "Request") "schema" $bodySchema "label" (print $op.OperationID " Body") )}}
        {{- end }}
        {{- $opResponse := $.GetOpHappyResponse $.PackageName $op }}
        {{- if isNotNil $opResponse.MediaType }}
            {{- if empty $opResponse.MediaType.Schema.Ref -}}
                {{- template "typeDeclaration" ($.WithParams "key" (print $op.OperationID " Response") "schema" $opResponse.MediaType.Schema "label" (print $op.OperationID " Response") )}}
            {{- end }}
        {{- end }}
    {{- end }}
{{- end }}
