# Fōji
SQL + Swagger = Service boilerplate

Load SQL (table, enum, index)
Parse SQL (queries.sql)

genQuery

genRepo - postgres
stubSwagger - Resources, LCRUD, flag - security - Dev time process

Load Swagger
genService

	- integer types smallint, integer, and bigint are returned as int64
	- floating-point types real and double precision are returned as float64
	- character types char, varchar, and text are returned as string
	- temporal types date, time, timetz, timestamp, and timestamptz are returned as time.Time
	- the boolean type is returned as bool
	- the bytea type is returned as []byte
	

Config:	
Database
Schema

Type Mapping
File Mapping



Sources
    SQL Queries (and database connection)
    Database (schema, tables, views, enums)
    

    

# Config
DB
Package
    Name
    Queries
    Overrides*
NameMapperTemplate
QueryNameMapperTemplate
Overrides
    Type
    Null
    Field
    Name
    
    
 Feature Checklist

- [x] Dump templates
- [X] Parse Custom SQL 
- [X] Read Postgres Schema 
- [x] Generate Repo (pgx) 
- [x] Stub Swagger from Schema 
- [x] Read Swagger Spec 
- [x] Generate chi v5 Router
- [x] Format Encode/Decode
- [x] Input Validation 
- [x] Authentication
- [x] Publish Swagger Spec (RapiDoc)
- [x] Authorization 
- [x] Support [OR auth](https://swagger.io/docs/specification/authentication/)
- [ ] Wrap Open API structs with helpers
- [ ] Primary Tag OpenAPI generation 
- [ ] Optionally Merge large param sets into struct 
- [ ] Filter private attributes from Spec for export
- [ ] Support multiple encodings based on request at runtime
- [ ] Support pluggable encodings
- [ ] Support AsyncAPI

Standard HTTP

Validation

Auth


Supported input models:
- Database schema (inspected directly from a DB connection)
- SQL Queries (read from file and validated against DB connection)
- OpenAPI/Swagger v3 (read from file)
- AsyncAPI (read from file)

Each input model supports various output models:
- SQL Schema
  - List (console debug)
  - OpenAPI Stub (jumpstart for creating a service spec based on db schema)
  - Repo package 
- SQL Queries
  - Repo package
- OpenAPI
  - HTTPRouter
  - Model Validation
  - Model Marshal/Unmarshall
  - Service Definition
- AsyncAPI
  - NATSRouter
  - Model Validation
  - Model Marshal/Unmarshall
  - Service Definition


Output Config
    Go
        Maps
        Case Converter
        
    OpenAPI
        Maps
        Case Converter
        




Customizing output models is streamlined by:
`foji dumpTemplates`

By default, this will write all templates to `./foji/*`.  After updating the templates as desired, update the config to use the new templates.

## Packages

Packages define the set of input and output for each generation.  Some output models are designed to only be used during development (OpenAPI stub) and some are designed for using with a CI tool.  A typical use case is:

1. Define initial data model using tool of choice (the example TODO app uses dbmate for management of schema).
1. Run `foji dbList` to test config _optional_
1. Run `foji dbRepo` to generate repo packages
1. Create custom sql. _optional_
1. Run `foji sqlRepo` to generate custom sql functions on repo _optional_
1. Run `foji openApiStub` to create simple OpenAPIv3 spec
1. Modify the generated OpenAPIv3 spec as appropriate for your needs.
1. Run `foji apiList` to test OpenAPIv3 spec _optional_
1. Run `foji api` to create transport (router), marshaller, validation, security 
1. Implement any business logic
1. Run you service!


## Type Mapping

Each package
  
## Limitations

[goimports](https://godoc.org/golang.org/x/tools/cmd/goimports) is recommended to execute after generation for final 
cleanup.  Specifically, the imports support maps based on all possible type conversions and running goimports will cleanup extraneous import statements. 

The Custom SQL evaluator is unable to determine if a result field is nullable in some cases (especially 
calculated fields).  So if the output template requires a distinction, either map the overrides.Type directly to the 
type that supports nulls (e.g. `varchar -> *string`), or use the qualified param name for the mapping 
(e.g. `GetTodoByCategoryLabel.label_nullable->*string`) 


# Config

foji uses inheritance with simple merging rules for the maps and lists.  Any list is replaced by descendants 
completely.  Maps are merged, where each value for each key is replaced with the inherited value.  You can use 
`foji dumpConfig -d` to see all available processes fully merged.  The flow is  
`default Format -> Format -> default Process -> Process`.

For example the default format for go runs `goimports` in the post processor.  To disable this simply set the post to 
an empty array like:
```yaml
    post:
      -
```

To remove a value from an inherited map key simple set the value to `-`.  For example:

```yaml

```

# Template Conventions

## Naming

Generated code files should be named `*_gen.*`.

Generated go files should have "Code generated" comments as the first line, with a blank line following.  This is 
a standard in the community, with editors giving warnings not to modify.  An example 

```gotemplate
// Code generated by foji {{ version }}, template: {{ templateFile }}; DO NOT EDIT.

```

Example output:
```go
// Code generated by foji 0.3, template: foji/openapi/model.go.tpl; DO NOT EDIT.
```

`version` (string version of the foji tool) and `templateFile` (the currently executing template file) are part of the 
standard template func set (available to all templates).

An exception to the above naming and comment rules are when you generate a file as a "starting point".  For example 
openAPI generates a project `main.go` as an example service.  Once this file is generated, it should never be replaced  
Prefix the file name with `!` to disable generation on existing files.

```yaml
      '!test/cmd/serve/main.go': foji/openapi/main.go.tpl
```

Example log:
```
WARN[0000] skipped, output file exists                   target="!test/cmd/serve/main.go" template=foji/openapi/main.go.tpl
```

## Optional Generation
You can use the `Context` method `NotNeededIf` to abort the current file generation  if the file is not needed.  
Example:

```gotemplate
{{ .NotNeededIf (empty .API.Components.SecuritySchemes) "no security schemes" -}}
```

Example log:
```
INFO[0000] skipped, not needed: no security schemes      target=test/http/auth_gen.go template=foji/openapi/auth.go.tpl
```

# Known Issues

## DB/Schema Parsing
Expression based indexes (like on jsonb fields) are not extracted