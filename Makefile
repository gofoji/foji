genPGQueries:
	go run main.go gq --config pgqueries.yaml

testSchemaList:
	go run main.go schemaList --config testlist.yaml

testStub:
	go run main.go stubAPI --config testTodo.yaml

testDumpConfig:
	go run main.go dumpConfig --config embed.yaml

qc:
	golangci-lint run --enable-all -D prealloc -D lll -D gochecknoglobals,gomnd,nolintlint --skip-dirs test,cmd,scratch,gen --tests=false

tidy:
	go mod tidy

update:
	go get -u all

install:
	go install


.PHONY:	qc testStub testList genPGQueries update tidy