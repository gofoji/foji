{{- define "propertyDeclaration"}}
    {{- $key := .RuntimeParams.key }}
    {{- $schema := .RuntimeParams.schema }}
    {{- $typeName := .RuntimeParams.typeName }}
    {{- $isRequired := .RuntimeParams.isRequired }}
    {{- goDoc $schema.Value.Description }}
    {{- $type := $.GetType .PackageName (print $typeName " " $key) $schema }}
    {{- if $isRequired }}
        {{ pascal $key }} {{ $type }} `json:"{{$key}}"`
    {{- else }}
        {{ pascal $key }} {{ if $schema.Value.Nullable }}*{{ end }}{{ $type }} `json:"{{$key}},omitempty{{- if $.SchemaIsObject $schema -}},omitzero{{ end }}"`
    {{- end }}
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

func (e {{ $enumType }}) Value() (driver.Value, error) {
	return json.Marshal(e.String())
}

func (e *{{ $enumType }}) Scan(src interface{}) error {
	s, ok := src.(string)
	if !ok {
		return fmt.Errorf("{{ $enumType }}.scan: scanned a %T, not []byte", src) //nolint
	}

	*e = New{{ $enumType }}(s)

	return nil
}

{{ end -}}
{{- end -}}

{{- define "validateField"}}
    {{- $fieldName := .RuntimeParams.fieldName }}
    {{- $schema := .RuntimeParams.schema }}
    {{- $typeName := .RuntimeParams.typeName }}
    {{- $isPointer := .RuntimeParams.isPointer }}

    {{- $fieldType := $.GetType $.PackageName "" $schema }}
    {{- $pascalField := "" -}}
    {{- $fieldDot := "p" -}}
    {{- if notEmpty $fieldName -}}
        {{- $fieldType = $.GetType $.PackageName $fieldName $schema }}
        {{- $pascalField = pascal $fieldName -}}
        {{- $fieldDot = print "p." $pascalField -}}
    {{- end -}}
    {{- $fieldDeref := $fieldDot -}}
    {{- if $isPointer -}}
        {{- $fieldDeref = print "*" $fieldDot -}}
    {{- end -}}

    {{- if or (notEmpty $schema.Ref)  ($schema.Value.Type.Is "object") }}
        {{ if $.HasValidation $schema }}
            {{ if $isPointer }}
                if {{$fieldDot}} != nil {
                    if subErr := {{ $fieldDot }}.Validate(); subErr != nil {
                        _ = err.Add("{{ $fieldName }}", subErr)
                    }
                }
            {{ else }}
                if subErr := {{ $fieldDot }}.Validate(); subErr != nil {
                    _ = err.Add("{{ $fieldName }}", subErr)
                }
            {{ end }}
        {{- end -}}
    {{- else if $schema.Value.Type.Is "array" }}
        {{- if gt $schema.Value.MinItems 0 }}

            if len({{ $fieldDeref }}) < {{ $schema.Value.MinItems }} {
            _ = err.Add("{{$fieldName}}", "length must be >= {{ $schema.Value.MinItems }}")
            }
        {{- end }}
        {{- if isNotNil $schema.Value.MaxItems }}

            if len({{ $fieldDeref }}) > {{ $schema.Value.MaxItems }} {
            _ = err.Add("{{$fieldName}}", "length must be <= {{ $schema.Value.MaxItems }}")
            }
        {{- end }}
    {{- else if or ($schema.Value.Type.Is "number") ($schema.Value.Type.Is "integer") }}
        {{- if isNotNil $schema.Value.Min }}

    if  {{ $fieldDeref }} <{{ if $schema.Value.ExclusiveMin }}={{end}} {{ $schema.Value.Min }} {
        _ = err.Add("{{$fieldName}}", "must be >{{ if not $schema.Value.ExclusiveMin }}={{end}} {{ $schema.Value.Min }}")
    }
        {{- end }}
        {{- if isNotNil $schema.Value.Max }}

    if {{ $fieldDeref }} >{{ if $schema.Value.ExclusiveMax }}={{end}} {{ $schema.Value.Max }} {
        _ = err.Add("{{$fieldName}}", "must be <{{ if not $schema.Value.ExclusiveMax }}={{end}} {{ $schema.Value.Max }}")
    }
        {{- end }}
        {{- if isNotNil $schema.Value.MultipleOf }}
            {{- if $schema.Value.Type.Is "integer" }}

    if {{ $fieldDeref }} % {{ $schema.Value.MultipleOf }} != 0 {
        _ = err.Add("{{$fieldName}}", "must be multiple of {{ $schema.Value.MultipleOf }}")
    }
            {{- else }}

    if math.Mod({{ if not (eq $fieldType "float64") }}float64({{ end }}{{ $fieldDeref }}{{ if not (eq $fieldType "float64") }}){{end}}, {{ $schema.Value.MultipleOf }}) != 0 {
        _ = err.Add("{{$fieldName}}", "must be multiple of {{ $schema.Value.MultipleOf }}")
    }
            {{- end }}
        {{- end }}
    {{- else if $schema.Value.Type.Is "string" }}
        {{- if gt $schema.Value.MinLength 0 }}

    if len({{ $fieldDeref }}) < {{ $schema.Value.MinLength }} {
        _ = err.Add("{{$fieldName}}", "length must be >= {{ $schema.Value.MinLength }}")
    }
        {{- end }}
        {{- if isNotNil $schema.Value.MaxLength }}

    if len({{ $fieldDeref }}) > {{ $schema.Value.MaxLength }} {
        _ = err.Add("{{$fieldName}}", "length must be <= {{ $schema.Value.MaxLength }}")
    }
        {{- end }}
        {{- if notEmpty $schema.Value.Pattern }}

    if {{ $fieldDeref }} != "" && !{{ camel $typeName }}{{ $pascalField }}Pattern.MatchString(string({{ $fieldDeref }}))  {
        _ = err.Add("{{$fieldName}}", `must match "{{ $schema.Value.Pattern }}"`)
    }
        {{- end }}
    {{- end }}
{{- end -}}


{{- define "typeDeclaration"}}
{{ $mediaType := .RuntimeParams.mediaType }}
{{ $schema := .RuntimeParams.schema }}
{{- $key := .RuntimeParams.key }}
{{- $label := .RuntimeParams.label }}

{{- if not ($.HasExtension $schema "x-go-type" )}}
{{- $typeName := $.GetType $.PackageName $key $schema }}
// {{ $typeName}}
{{- goDoc $schema.Value.Description }}
//
// OpenAPI {{$label}}: {{ $key }}

    {{- if $.IsDefaultEnum $key $schema }}
        {{- $label := .RuntimeParams.label }}
        {{- template "enum" ($.WithParams "name" $key "schema" $schema "description" (print $label " : " $key ))}}

    {{- else if and ($schema.Value.Type.Permits "object") (gt (len ($.SchemaProperties $schema true)) 0) }}
type {{ pascal $key }} struct {
    {{- range $field, $schemaProp := $.SchemaProperties $schema false}}
        {{- $isRequired := $.IsRequiredProperty $field $schema -}}
        {{- template "propertyDeclaration" ($.WithParams "key" $field "schema" $schemaProp "typeName" $typeName "isRequired" $isRequired)}}
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

{{- $hasValidation := $.HasValidation $schema -}}

{{- /* Nested Types */}}
    {{- range $key, $schemaProp := $.SchemaProperties $schema false }}
        {{- if not (empty $schemaProp.Value.Properties )}}
            {{- if empty $schemaProp.Ref -}}
                {{- template "typeDeclaration" ($.WithParams "mediaType" "application/json" "key" (pascal (print $typeName " " $key)) "schema" $schemaProp "label" (print  $typeName " inline " $key))}}
            {{- end -}}
        {{- else if $schemaProp.Value.Type.Is "array"}}
            {{- if empty $schemaProp.Value.Items.Ref -}}
                {{- $isEnumItem := $.IsDefaultEnum $key $schemaProp.Value.Items }}
                {{- $hasProperties := not (empty ($.SchemaProperties $schemaProp.Value.Items false )) }}
                {{- if or $isEnumItem $hasProperties}}
                    {{- template "typeDeclaration" ($.WithParams "mediaType" "application/json" "key" (pascal (print $typeName " " $key)) "schema" $schemaProp.Value.Items "label" (print  $typeName " inline item " $key))}}
                {{- end }}
            {{- end }}
        {{- end }}
    {{- end }}
        {{- /* Nested Arrays */}}
    {{- if $schema.Value.Type.Is "array"}}
        {{- if empty $schema.Value.Items.Ref -}}
            {{- $isEnumItem := $.IsDefaultEnum $key $schema.Value.Items }}
            {{- $hasProperties := not (empty ($.SchemaProperties $schema.Value.Items false )) }}
            {{- if or $isEnumItem $hasProperties}}
                {{- template "typeDeclaration" ($.WithParams "mediaType" "application/json" "key" (pascal (print $typeName " Item" )) "schema" $schema.Value.Items "label" (print  $typeName " inline item " $key))}}
            {{- end }}
        {{- end }}
    {{- end }}

{{- /*    Regex Validation Patterns */ -}}
    {{- range $key, $schemaProp := $.SchemaProperties $schema false}}
        {{- if notEmpty $schemaProp.Value.Pattern }}
var {{ camel $typeName }}{{ pascal $key }}Pattern = regexp.MustCompile(`{{ $schemaProp.Value.Pattern }}`)
        {{- end}}
    {{- end }}
    {{- if and $hasValidation (notEmpty $schema.Value.Pattern) }}
var {{ camel $key }}Pattern = regexp.MustCompile(`{{ $schema.Value.Pattern }}`)
    {{ end }}

{{- /*    Enums */}}
    {{- range $key, $schemaEnum := $.SchemaEnums $schema }}
        {{- template "enum" ($.WithParams "name" (print $typeName " " $key) "schema" $schemaEnum "description" (print $label " : " $key ))}}
    {{- end -}}

{{if eq $mediaType "application/json" }}
    {{- if or $hasValidation $schema.Value.Required}}

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

        {{ range $field, $schemaProp := $.SchemaProperties $schema false}}
            {{ $typeName := (print $key " " $field) -}}
            {{- if notEmpty $schemaProp.Ref -}}
                {{- $typeName = trimPrefix "#/components/parameters/" $schemaProp.Ref -}}
            {{- end -}}
            {{- $goType := $.GetType $.PackageName $typeName $schemaProp }}
            {{- $isEnum := $.SchemaIsEnum $schemaProp }}
            {{- $hasDefault := isNotNil $schemaProp.Value.Default }}

            {{- if $hasDefault }}
    if _, ok := requiredCheck["{{ $field }}"]; !ok {
                {{- if $isEnum -}}
        v.{{ pascal $field }} = {{- $goType}}{{ pascal (goToken (printf "%#v" $schemaProp.Value.Default)) }}
                {{else -}}
                    {{- if $schemaProp.Value.Nullable }}
        defaultVal := {{ printf "%#v" $schemaProp.Value.Default }}
        v.{{ pascal $field }} = &defaultVal
                    {{ else -}}
        v.{{ pascal $field }} = {{ printf "%#v" $schemaProp.Value.Default }}
                    {{- end }}
                {{- end -}}
    }
            {{- end -}}
        {{end}}

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
{{end}}

{{if or (eq $mediaType "multipart/form-data") (eq $mediaType "application/x-www-form-urlencoded")}}

func ParseForm{{ pascal $key }}(r *http.Request) ({{ pascal $key }}, error) {
	var (
	  parseErrors validation.Errors
	  err error
	{{- if .SchemaPropertiesHaveDefaults $schema }}
	  ok bool
	{{- end }}
      v {{ pascal $key }}
    )

    {{ range $field, $schemaProp := $.SchemaProperties $schema false}}
        {{- $isRequired := $.IsRequiredProperty $field $schema }}
        {{ $typeName := (print $key " " $field) -}}
        {{- if notEmpty $schemaProp.Ref -}}
            {{- $typeName = trimPrefix "#/components/parameters/" $schemaProp.Ref -}}
        {{- end -}}
        {{- $goType := $.GetType $.PackageName $typeName $schemaProp }}
	    {{- $isEnum := $.SchemaIsEnum $schemaProp }}
        {{- $isArray := $schemaProp.Value.Type.Is "array" }}
        {{- $isArrayEnum := $.SchemaIsEnumArray $schemaProp }}
        {{- $enumNew := $.EnumNew $goType }}
        {{- $hasDefault := isNotNil $schemaProp.Value.Default }}

        {{- if $isArray -}}
            {{- if eq $goType "[]int32" -}}
    v.{{ pascal $field }}, {{- if $hasDefault }}ok{{ else }}_{{ end -}}, err = forms.GetInt32Array(r.FormValue, "{{ $field }}", {{ $isRequired }})
            {{- else if $isArrayEnum }}
    v.{{ pascal $field }}, {{- if $hasDefault }}ok{{ else }}_{{ end -}}, err = forms.GetEnumArray(r.FormValue, "{{ $field }}", {{ $isRequired }}, {{ $enumNew }})
            {{- else }}
    v.{{ pascal $field }}, {{- if $hasDefault }}ok{{ else }}_{{ end -}}, err = forms.GetStringArray(r.FormValue, "{{ $field }}", {{ $isRequired }})
            {{- end }}
    if err != nil {
        parseErrors.Add("{{ $field }}", err)
    }
        {{- else -}}
            {{- if eq $goType "bool" -}}
    v.{{ pascal $field }}, {{- if $hasDefault }}ok{{ else }}_{{ end -}}, err = forms.GetBool(r.FormValue, "{{ $field }}", {{ $isRequired }})
            {{- else if eq $goType "int32" }}
    v.{{ pascal $field }}, {{- if $hasDefault }}ok{{ else }}_{{ end -}}, err = forms.GetInt32(r.FormValue, "{{ $field }}", {{ $isRequired }})
            {{- else if eq $goType "int64" }}
    v.{{ pascal $field }}, {{- if $hasDefault }}ok{{ else }}_{{ end -}}, err = forms.GetInt64(r.FormValue, "{{ $field }}", {{ $isRequired }})
            {{- else if eq $goType "time.Time" }}
    v.{{ pascal $field }}, {{- if $hasDefault }}ok{{ else }}_{{ end -}}, err = forms.GetTime(r.FormValue, "{{ $field }}", {{ $isRequired }})
            {{- else if eq $goType "uuid.UUID" }}
    v.{{ pascal $field }}, {{- if $hasDefault }}ok{{ else }}_{{ end -}}, err = forms.GetUUID(r.FormValue, "{{ $field }}", {{ $isRequired }})
            {{- else if $isEnum }}
    v.{{ pascal $field }}, {{- if $hasDefault }}ok{{ else }}_{{ end -}}, err = forms.GetEnum(r.FormValue, "{{ $field }}", {{ $isRequired }}, {{ $enumNew }})
            {{- else if eq $goType "forms.File" }}
    v.{{ pascal $field }}, {{- if $hasDefault }}ok{{ else }}_{{ end -}}, err = forms.GetFile(r, "{{ $field }}", {{ $isRequired }})
            {{- else }}
    v.{{ pascal $field }}, {{- if $hasDefault }}ok{{ else }}_{{ end -}}, err = forms.GetString(r.FormValue, "{{ $field }}", {{ $isRequired }})
            {{- end }}
    if err != nil {
        parseErrors.Add("{{ $field }}", err)
            {{- if $hasDefault }}
    } else if !ok {
        v.{{ pascal $field }} = {{ if $isEnum -}}
                {{- $goType}}{{ pascal (goToken (printf "%#v" $schemaProp.Value.Default)) }}
                {{else -}}
                    {{- if and (eq $goType "time.Time") (eq $schemaProp.Value.Default "") -}}
                        time.Time{}
                    {{else -}}
                        {{ printf "%#v" $schemaProp.Value.Default }}
                    {{- end -}}
                {{- end -}}
    }
            {{else}}
    }
            {{end}}
        {{- end }}

    {{ end }}

	if parseErrors != nil {
		return {{ pascal $key }}{}, parseErrors.GetErr()
	}

    {{ if $hasValidation}}
    if err := v.Validate(); err != nil {
        return {{ pascal $key }}{}, err
    }
    {{ end }}

    return v, nil
}
{{end}}


{{- if $hasValidation }}
    {{- if not (empty $schema.Value.Properties )}}
func (p {{ pascal $key }}) Validate() error {
    var err validation.Errors
        {{ range $fieldName, $schemaProp := $.SchemaProperties $schema true }}
            {{- if $.HasValidation $schemaProp }}
    p.Validate{{ pascal $fieldName }}(&err)
            {{- end }}
        {{- end }}

    return err.GetErr()
}
        {{- range $fieldName, $schemaProp := $.SchemaProperties $schema true }}
            {{- if $.HasValidation $schemaProp }}
                {{- $isRequired := $.IsRequiredProperty $fieldName $schema -}}
                {{- $isPointer := and (not $isRequired) ($schemaProp.Value.Nullable) }}

func (p {{ pascal $key }}) Validate{{ pascal $fieldName }}(err *validation.Errors) {
                {{- if $isPointer }}
	if p.{{ pascal $fieldName }} == nil {
		return
	}

                {{ end -}}
                {{- template "validateField" ($.WithParams "fieldName" $fieldName "schema" $schemaProp "typeName" $key "isPointer" $isPointer ) -}}
}
            {{- end }}
        {{- end }}
    {{ else -}}
func (p {{ pascal $key }}) Validate() error {
    var err validation.Errors
        {{- template "validateField" ($.WithParams "fieldName" "" "schema" $schema "typeName" $key "isPointer" false )}}

    return err.GetErr()
}
    {{- end -}}
{{- end -}}
{{- end -}}
{{- end -}}

// Code generated by foji {{ version }}, template: {{ templateFile }}; DO NOT EDIT.

package {{ .PackageName }}

import (
    "errors"
    "regexp"

{{- .CheckAllTypes .PackageName -}}
{{ range .GoImports }}
    "{{ . }}"
{{- end }}

    "github.com/bir/iken/forms"
    "github.com/bir/iken/validation"
)

var ErrMissingRequiredField = errors.New("missing required field")

// Component Schemas

{{ range $key, $schema := .ComponentSchemas }}
    {{- template "typeDeclaration" ($.WithParams "mediaType" "application/json" "key" $key "schema" $schema "label" "Component Schema")}}
{{- end }}

// Component Parameters

{{ range $key, $param := .ComponentParameters }}
        {{- template "paramDeclaration" ($.WithParams "param" $param "name" $key "label" "Component Parameter: ")}}
{{- end }}

{{- define "paramDeclaration"}}
    {{- $param := .RuntimeParams.param }}
    {{- $name := .RuntimeParams.name }}
    {{- $label := .RuntimeParams.label }}
    {{- if empty $param.Ref -}}
        {{- template "enum" ($.WithParams "name" $name "schema" $param.Value.Schema "description" (print $param.Value.Description "\n" $label $param.Value.Name ))}}
        {{- if $param.Value.Schema.Value.Type.Is "array"}}
            {{- template "enum" ($.WithParams "name" $name "schema" $param.Value.Schema.Value.Items "description" (print $label $param.Value.Name " Item"))}}
        {{- end }}
    {{- end -}}
{{- end }}

// Path Operations

{{/* Inline Request/Reponse Types */ -}}
{{ range $name, $path := .API.Paths.Map }}
    {{- range $verb, $op := $path.Operations }}
        {{- /* Inline Request */ -}}
        {{ range $opBody := $.GetRequestBodySchemas $op }}
            {{- if $.SchemaIsComplex $opBody.Schema -}}
                {{- template "typeDeclaration" ($.WithParams "mediaType" $opBody.MimeType "key" (print $op.OperationID " Request") "schema" $opBody.Schema "label" (print $op.OperationID " Body") )}}
            {{- end }}
        {{- end }}

        {{- /* Inline Response */ -}}
        {{- $opResponse := $.GetOpHappyResponse $.PackageName $op }}
        {{- if isNotNil $opResponse.MediaType }}
            {{- if  $.SchemaIsComplex $opResponse.MediaType.Schema -}}
                {{- template "typeDeclaration" ($.WithParams "mediaType" "application/json" "key" (print $op.OperationID " Response") "schema" $opResponse.MediaType.Schema "label" (print $op.OperationID " Response") )}}
            {{- end }}
        {{- end }}

        {{- /* Inline Params */ -}}
        {{- range $param := $.OpParams $path $op }}
            {{- template "paramDeclaration" ($.WithParams "param" $param "name" (print $op.OperationID " " $param.Value.Name) "label" (print "Op: " $op.OperationID " Param: "))}}
        {{- end }}
    {{- end }}
{{- end }}

