sqlRepo:
	go run main.go weld sqlRepo

testSchemaList:
	go run main.go schemaList --config testlist.yaml

testStub:
	go run main.go stubAPI --config testTodo.yaml

testDumpConfig:
	go run main.go dumpConfig --config embed.yaml

lint:
	golangci-lint run --sort-results

test:
	go test ./...

test_gen:
	go run main.go weld openAPI -c tests/csvresponse/foji.yaml
	cd tests; go run tests_main.go

cover:
	go test	-coverprofile cp.out ./...
	go tool cover -html=cp.out

tidy:
	go mod tidy -compat=1.19

update:
	go get -u all

install:
	go run main.go weld embed
	go install

.PHONY:	sqlRepo testSchemaList testStub testDumpConfig lint test cover tidy update install test_gen