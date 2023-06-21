package main

import (
	"os"

	"tests/csvresponse"
)

func main() {
	var _ csvresponse.Operations = &csvresponse.Service{}

	os.Exit(0)
}
