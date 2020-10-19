openapi: "3.0.0"
info:
  title: {{case .Params.Name}}
  version: 0.1.0
paths:
{{- range .Resources }}
  {{- $resource := . }}
  /{{ camel .Name }}:
    get:
      operationId: list{{ pascal .Name }}
      summary: List {{ pascal .Name }}
      tags:
        - {{ case .Name }}
      parameters:
        - $ref: '#/components/parameters/offset'
        - $ref: '#/components/parameters/limit'
      responses:
        '200':
          description: |-
            200 response
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/{{ case (pluralUniqueName .Name) }}'
        default:
          description: unexpected error
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Error"
{{- if not .Table.ReadOnly }}
    post:
      operationId: create{{ pascal .Name }}
      summary: Create {{ pascal .Name }}
      tags:
        - {{ case .Name }}
      responses:
        '201':
          description: |-
            200 response
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/{{ pascal .Name }}'
{{end -}}
{{if .PK -}}
{{with .PK }}
  /{{ camel $resource.Name }}/{ {{- camel .Name -}} }:
    parameters:
      - name: {{ camel .Name }}
        in: path
        description: {{ camel .Name }} of {{ pascal $resource.Name }}
        required: true
        schema:
          type: {{ .Type }}
{{- with .Format }}
          format: {{ . }}
{{end -}}
{{end }}
    get:
      description: Get {{ pascal .Name }}
      operationId: get{{ pascal .Name }}
      summary: Get {{ pascal .Name }}
      tags:
        - {{ case .Name }}
      responses:
        '200':
          description: successful response
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/{{ pascal .Name }}'
        default:
          description: unexpected error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
{{- if not .Table.ReadOnly }}
    post:
      description: updates a {{ pascal .Name }}
      operationId: update{{ pascal .Name }}
      summary: Update {{ pascal .Name }}
      tags:
        - {{ case .Name }}
      responses:
        '200':
          description: {{ pascal .Name }} updated
        default:
          description: unexpected error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
    delete:
      description: deletes a {{ pascal .Name }}
      operationId: delete{{ pascal .Name }}
      summary: Delete {{ pascal .Name }}
      tags:
        - {{ case .Name }}
      responses:
        '204':
          description: {{ pascal .Name }} deleted
        default:
          description: unexpected error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
{{end -}}
{{end -}}
{{end -}}
components:
  responses:
    BadRequest:
      description: The specified resource was not found
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'
    NotFound:
      description: The specified resource was not found
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'
    Unauthorized:
      description: Unauthorized
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'
  parameters:
    limit:
      name: limit
      in: query
      description: How many items to return for each page
      required: false
      schema:
        type: integer
        format: int32
        default: 100
        maximum: 255
        minimum: 10
    offset:
      name: offset
      in: query
      description: Offset of page to return
      required: false
      schema:
        type: integer
        format: int32
        default: 0
        minimum: 0
  schemas:
{{- range .Resources }}
    {{pascal .Name}}:
      type: object
      properties:
{{- range .Properties }}
        {{ camel .Name }}:
          type: {{ .Type }}
{{- with .Format }}
          format: {{ . }}
{{- end -}}
{{end}}
    {{case (pluralUniqueName .Name) }}:
      type: object
      properties:
        page:
          type: integer
        pageSize:
          type: integer
        totalRecordCount:
          type: integer
        list:
            type: array
            items:
              $ref: '#/components/schemas/{{ pascal .Name }}'
{{- end}}
    Error:
      type: object
      required:
        - code
        - message
      properties:
        code:
          type: integer
          format: int32
        message:
          type: string