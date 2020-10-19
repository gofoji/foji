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
- [x] Generate FastHTTPRouter
- [x] Format Encode/Decode
- [x] Input Validation 
- [x] Authentication
- [x] Publish Swagger Spec
- [ ] Authorization 
- [ ] Optionally Merge large param sets into struct 
- [ ] Filter private attributes from Spec for export
- [ ] Support multiple encodings
- [ ] Support AsyncAPI


Standard HTTP

https://github.com/julienschmidt/httprouter

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

[goimports](https://godoc.org/golang.org/x/tools/cmd/goimports) is recommended to execute after generation for final cleanup.  Specifically, the imports support maps based on all possible type conversions and running goimports will cleanup extraneous import statements. 

The Custom SQL evaluator is unable to determine if a result field is nullable in some cases (especially calculated fields).  So if the output template requires a distinction, either map the overrides.Type directly to the type that supports nulls (`varchar -> *string`), or use the qualified param name for the mapping (`GetTodoByCategoryLabel.label_nullable->*string`) 




Generator: Query

TODO: Convert post run to template syntax
TODO: Type func