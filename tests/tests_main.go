package main

import (
	"context"
	"net/http"
	"os"

	"tests/auth"
	"tests/csvresponse"
	"tests/example"
)

func tokenAuth(ctx context.Context, token string) (*example.ExampleAuth, error) {
	return &example.ExampleAuth{}, nil
}

func rawAuth(r *http.Request) (*example.ExampleAuth, error) {
	return &example.ExampleAuth{}, nil
}

func basicAuth(ctx context.Context, user, pass string) (*example.ExampleAuth, error) {
	return &example.ExampleAuth{}, nil
}

func authorize(ctx context.Context, user *example.ExampleAuth, scopes []string) error {
	return nil
}

func main() {
	// Requires the generated Operations to match the Service layer
	var _ csvresponse.Operations = &csvresponse.Service{}

	var ops example.Operations = &example.Service{}
	example.RegisterHTTP(ops, http.NewServeMux(), tokenAuth, tokenAuth, tokenAuth, tokenAuth, rawAuth, authorize)

	var authOps auth.Operations = &auth.Service{}
	auth.RegisterHTTP(authOps, http.NewServeMux(), tokenAuth, tokenAuth, tokenAuth, basicAuth, tokenAuth, rawAuth, rawAuth, rawAuth, authorize)

	os.Exit(0)
}
