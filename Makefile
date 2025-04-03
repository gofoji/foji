MODULE := $(shell go list -m -f {{.Path}})

sqlRepo:
	go run main.go weld sqlRepo

fmt:
	gofumpt -l -w .
	gci write . -s standard -s default -s "prefix($(MODULE))"

lint:
	golangci-lint run --sort-results

test:
	go test ./...

test_gen:
	go run main.go weld openAPI -c tests/csvresponse/foji.yaml
	go run main.go weld openAPI -c tests/example/foji.yaml
	cd tests; go run tests_main.go
	cd tests; go test ./...

cover:
	go test	-coverprofile cp.out ./...
	go tool cover -html=cp.out

tidy:
	go mod tidy

update: updateAll tidy

updateAll:
	go get -u ./...

install:
	go install

.PHONY:	sqlRepo testSchemaList testStub testDumpConfig lint test cover tidy update updateAll install test_gen