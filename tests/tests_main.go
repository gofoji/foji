package main

import (
	"os"

	"tests/csvresponse"
	"tests/example"
)

func main() {
	// Requires the generated Operations to match the Service layer
	var _ csvresponse.Operations = &csvresponse.Service{}

	var _ example.Operations = &example.Service{}

	os.Exit(0)
}
