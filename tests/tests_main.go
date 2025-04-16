package main

import (
	"context"
	"net/http"
	"os"

	"tests/csvresponse"
	"tests/example"
)

func tokenAuth(ctx context.Context, token string) (*example.ExampleAuth, error) {
	return &example.ExampleAuth{}, nil
}

func rawAuth(r *http.Request) (*example.ExampleAuth, error) {
	return &example.ExampleAuth{}, nil
}

func authorize(ctx context.Context, user *example.ExampleAuth, scopes []string) error {
	return nil
}

func main() {
	// Requires the generated Operations to match the Service layer
	var _ csvresponse.Operations = &csvresponse.Service{}

	var ops example.Operations = &example.Service{}

	example.RegisterHTTP(ops, http.NewServeMux(), tokenAuth, tokenAuth, rawAuth, authorize)

	os.Exit(0)
}
