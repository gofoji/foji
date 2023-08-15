{{- define "propertyDeclaration"}}
    {{- $key := .RuntimeParams.key }}
    {{- $schema := .RuntimeParams.schema }}
    {{- $typeName := .RuntimeParams.typeName }}
    {{- goDoc $schema.Value.Description }}
    {{ pascal $key }} {{ $.GetType .PackageName (print $typeName " " $key) $schema }} `json:"{{$key}},omitempty"`
{{- end -}}

{{- define "enum"}}
{{- $schema := .RuntimeParams.schema }}
{{- if and (empty $schema.Ref) (not (empty $schema.Value.Enum)) }}
{{- $enumType := $.GetType $.PackageName .RuntimeParams.name $schema }}

// {{$enumType}}
{{- goDoc .RuntimeParams.description }}
type {{ $enumType }} int8

const (
    Unknown{{ $enumType }} {{ $enumType }} = iota
    {{- range $i, $value := $schema.Value.Enum }}
    {{ $enumType }}{{ pascal (goToken (printf "%v" $value)) }}
    {{- end }}
)

func New{{ $enumType }}(name string) {{ $enumType }} {
    switch name {
    {{- range $schema.Value.Enum  }}
    case "{{ . }}":
        return {{ $enumType }}{{ pascal (goToken (printf "%v" .)) }}
    {{- end }}
    }

    return {{ $enumType }}(0)
}

var  {{ $enumType }}String = map[{{ $enumType }}]string{
    {{- range $schema.Value.Enum }}
        {{ $enumType }}{{ pascal (goToken (printf "%v" .)) }}: "{{ (printf "%v" .) }}",
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

func (e {{ $enumType }}) MarshalJSON() ([]byte, error) {
    return json.Marshal(e.String())
}

{{ end -}}
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
    {{- range $key, $schema := $.SchemaProperties $schema  false}}
        {{- template "propertyDeclaration" ($.WithParams "key" $key "schema" $schema "typeName" $typeName)}}
    {{- end }}
    {{- range $schema.Value.AllOf }}
        {{- if notEmpty .Ref }}

    // OpenAPI Ref: {{ .Ref }}
    {{ $.GetType $.PackageName "" . }}
        {{- end }}
    {{- end }}
}
    {{- else }}
type {{ pascal $key }} {{ $.GetType $.PackageName (pascal (print $typeName " Item" )) $schema }}
    {{- end }}

{{- /* Nested Types */}}
    {{- range $key, $schema := $.SchemaProperties $schema false }}
        {{- if not (empty $schema.Value.Properties )}}
            {{- if empty $schema.Ref -}}
                {{- template "typeDeclaration" ($.WithParams "key" (pascal (print $typeName " " $key)) "schema" $schema "label" (print  $typeName " inline " $key))}}
            {{- end -}}
        {{- else if eq $schema.Value.Type "array"}}
            {{- if empty $schema.Value.Items.Ref -}}
                {{- if not (empty ($.SchemaProperties $schema.Value.Items false ))}}
                    {{- template "typeDeclaration" ($.WithParams "key" (pascal (print $typeName " " $key)) "schema" $schema.Value.Items "label" (print  $typeName " inline item " $key))}}
                {{- end }}
            {{- end }}
        {{- end }}
    {{- end }}
        {{- /* Nested Arrays */}}
    {{- if eq $schema.Value.Type "array"}}
        {{- if empty $schema.Value.Items.Ref -}}
            {{- if not (empty ($.SchemaProperties $schema.Value.Items false ))}}
                {{- template "typeDeclaration" ($.WithParams "key" (pascal (print $typeName " Item" )) "schema" $schema.Value.Items "label" (print  $typeName " inline item " $key))}}
            {{- end }}
        {{- end }}
    {{- end }}

{{- /*    Regex Validation Patterns */ -}}
    {{- range $key, $schema := $.SchemaProperties $schema false}}
        {{- if notEmpty $schema.Value.Pattern }}
var {{ camel $typeName }}{{ pascal $key }}Pattern = regexp.MustCompile(`{{ $schema.Value.Pattern }}`)
        {{- end}}
    {{- end }}

{{- /*    Enums */}}
    {{- range $key, $schema := $.SchemaEnums $schema }}
        {{- template "enum" ($.WithParams "name" (print $typeName " " $key) "schema" $schema "description" (print $label " : " $key ))}}
    {{- end -}}

{{- $hasValidation := $.HasValidation $schema -}}
{{- if or $hasValidation  $schema.Value.Required }}

func (p *{{ pascal $key }}) UnmarshalJSON(b []byte) error {
    {{- if $schema.Value.Required }}
    var requiredCheck map[string]any

    if err := json.Unmarshal(b, &requiredCheck); err != nil {
        return validation.Error{err.Error(), fmt.Errorf("{{ pascal $key }}.UnmarshalJSON Required: `%v`: %w", string(b), err)}
    }

    var validationErrors validation.Errors
    {{ range $field := $schema.Value.Required }}
    if _, ok := requiredCheck["{{ $field }}"]; !ok {
        validationErrors.Add("{{ $field }}", ErrMissingRequiredField)
    }
    {{ end }}

    if validationErrors != nil {
        return validationErrors.GetErr()
    }
    {{ end }}
    type  {{ pascal $key }}JSON {{ pascal $key }}
    var parseObject {{ pascal $key }}JSON

    if err := json.Unmarshal(b, &parseObject); err != nil {
        return validation.Error{err.Error(), fmt.Errorf("{{ pascal $key }}.UnmarshalJSON: `%v`: %w", string(b), err)}
    }

    v := {{ pascal $key }}(parseObject)

{{ if $hasValidation}}
    if err := v.Validate(); err != nil {
        return err
    }
{{ end }}

    *p = v

    return nil
}

    {{ if $hasValidation}}
func (p {{ pascal $key }}) MarshalJSON() ([]byte, error) {
    if err := p.Validate(); err != nil {
        return nil, err
    }

    type unvalidated {{ pascal $key }} // Skips the validation check
    b, err := json.Marshal(unvalidated(p))
    if err != nil {
        return nil, fmt.Errorf("{{ pascal $key }}.Marshal: `%+v`: %w", p, err)
    }

    return b, nil
}
    {{ end }}
{{end}}

    {{- if $.HasValidation $schema }}
func (p {{ pascal $key }}) Validate() error {
    var err validation.Errors
        {{- range $key, $schema := $.SchemaProperties $schema true }}
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

    if p.{{ pascal $key }} != "" && !{{ camel $typeName }}{{ pascal $key }}Pattern.MatchString( p.{{ pascal $key }})  {
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
            {{- else if notEmpty $schema.Ref }}
                {{- if $.HasValidation $schema }}

    if subErr := p.{{ pascal $key }}.Validate(); subErr != nil {
        _ = err.Add("{{ $key }}", subErr)
    }
                {{- end -}}
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

var ErrMissingRequiredField = errors.New("missing required field")

// Component Schemas

{{ range $key, $schema := .ComponentSchemas }}
    {{- template "typeDeclaration" ($.WithParams "key" $key "schema" $schema "label" "Component Schema")}}
{{- end }}

// Component Parameters

{{ range $key, $param := .ComponentParameters }}
        {{- template "paramDeclaration" ($.WithParams "param" $param "name" "" "label" "Component Parameter: ")}}
{{- end }}

{{- define "paramDeclaration"}}
    {{- $param := .RuntimeParams.param }}
    {{- $name := .RuntimeParams.name }}
    {{- $label := .RuntimeParams.label }}
    {{- if empty $param.Ref -}}
        {{- template "enum" ($.WithParams "name" (print $name " " $param.Value.Name) "schema" $param.Value.Schema "description" (print $param.Value.Description "\n" $label $param.Value.Name ))}}
        {{- if eq $param.Value.Schema.Value.Type "array"}}
            {{- template "enum" ($.WithParams "name" (print $name " " $param.Value.Name) "schema" $param.Value.Schema.Value.Items "description" (print $label $param.Value.Name " Item"))}}
        {{- end }}
    {{- end -}}
{{- end }}

// Path Operations

{{/* Inline Request/Reponse Types */ -}}
{{ range $name, $path := .API.Paths }}
    {{- range $verb, $op := $path.Operations }}
        {{- /* Inline Request */ -}}
        {{- $bodySchema := $.GetRequestBodyLocal $op}}
        {{- if $.SchemaIsComplex $bodySchema -}}
            {{- template "typeDeclaration" ($.WithParams "key" (print $op.OperationID "Request") "schema" $bodySchema "label" (print $op.OperationID " Body") )}}
        {{- end }}

        {{- /* Inline Response */ -}}
        {{- $opResponse := $.GetOpHappyResponse $.PackageName $op }}
        {{- if isNotNil $opResponse.MediaType }}
            {{- if  $.SchemaIsComplex $opResponse.MediaType.Schema -}}
                {{- template "typeDeclaration" ($.WithParams "key" (print $op.OperationID " Response") "schema" $opResponse.MediaType.Schema "label" (print $op.OperationID " Response") )}}
            {{- end }}
        {{- end }}

        {{- /* Inline Params */ -}}
        {{- range $param := $.OpParams $path $op }}
            {{- template "paramDeclaration" ($.WithParams "param" $param "name" $op.OperationID "label" (print "Op: " $op.OperationID " Param: "))}}
        {{- end }}
    {{- end }}
{{- end }}
