sqlRepo:
	go run main.go weld sqlRepo

fmt:
	golangci-lint fmt

lint:
	golangci-lint run

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

tools:
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest

.PHONY:	sqlRepo testSchemaList testStub testDumpConfig lint test test_gen cover tidy update updateAll install tools