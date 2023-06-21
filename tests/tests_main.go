package main

import (
	"os"

	"tests/csvresponse"
)

func main() {
	// Requires the generated Operations to match the Service layer
	var _ csvresponse.Operations = &csvresponse.Service{}

	os.Exit(0)
}
