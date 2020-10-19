{{- range $key, $value := .DB -}}
Schema: {{red}}{{$key}}{{colorReset}}
{{- range $value.Tables }}
  {{ pascal .Type }}: {{yellow}}{{ case .Name }}{{colorReset}}
  {{- if .ReadOnly }}  {{red}}readonly{{colorReset}} {{end}}
  {{- magenta}}  {{ .Comment }}{{colorReset}}
    Columns{{range .Columns }}
      {{green}}{{ pad (case .Name) (cases .Table.Columns.Names).Max }} {{colorReset}}  {{ pad .Type .Table.Columns.Types.Max }}  {{ if not .Nullable }}NOT NULL{{end}}  {{ if .IsPrimaryKey }}PK{{end}} {{magenta}}{{ .Comment }}{{colorReset}}
    {{- end }}
    {{- if not (empty .Indexes)}}
    Indexes{{range .Indexes }}
      {{ case .Name }} {{ (cases .Columns.Names).Join "," }} {{ if not .IsUnique }}UNIQUE{{end}} {{ if .IsPrimary }}{{cyan}}PK{{colorReset}}{{end}} {{magenta}}{{ .Comment }}{{colorReset}}
    {{- end }}{{end}}
    {{- if not (empty .ForeignKeys)}}
    ForeignKeys{{range .ForeignKeys }}
      {{ case .Name }} ({{ (cases .Columns.Names).Join "," }}) -> {{ case .ForeignTable.Name }}({{ (cases .ForeignColumns.Names).Join "," }}) {{magenta}}{{ .Comment }}{{colorReset}}
    {{- end }}{{end}}
    {{- if not (empty .References)}}
    References{{range .References }}
      {{ case .Name }} {{ .Table.Name }}({{ (cases .Columns.Names).Join "," }}) -> ({{ (cases .ForeignColumns.Names).Join "," }}) {{magenta}}{{ .Comment }}{{colorReset}}
    {{- end }}{{end}}
{{ end }}
{{- range $value.Enums }}
  enum: {{yellow}}{{ case .Name }}{{colorReset}} {{magenta}}{{ .Comment }}{{colorReset}}
    {{- range .Values }}
      {{green}}{{ case . }}{{colorReset}}
    {{- end }}
{{- end }}
{{- end }}