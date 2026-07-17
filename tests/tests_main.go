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

func csvClientToken(ctx context.Context, user *csvresponse.ExampleAuth) (string, error) {
	return "", nil
}

func clientToken(ctx context.Context, user *example.ExampleAuth) (string, error) {
	return "", nil
}

func clientBasic(ctx context.Context, user *example.ExampleAuth) (string, string, error) {
	return "", "", nil
}

func clientCookie(ctx context.Context, user *example.ExampleAuth) (*http.Cookie, error) {
	return nil, nil
}

func clientWrap(ctx context.Context, inner *http.Client, user *example.ExampleAuth) (*http.Client, error) {
	return inner, nil
}

func clientRaw(r *http.Request, inner *http.Client, user *example.ExampleAuth) (*http.Client, error) {
	return inner, nil
}

func isA[T any](_ T) {}

func main() {
	// Requires the generated Operations to match the Service layer
	var _ csvresponse.Operations = &csvresponse.Service{}

	var ops example.Operations = &example.Service{}
	example.RegisterHTTP(ops, http.NewServeMux(), tokenAuth, tokenAuth, tokenAuth, tokenAuth, rawAuth, authorize)

	var authOps auth.Operations = &auth.Service{}
	auth.RegisterHTTP(authOps, http.NewServeMux(), tokenAuth, tokenAuth, tokenAuth, basicAuth, tokenAuth, rawAuth, rawAuth, rawAuth, rawAuth, authorize)

	// Requires the generated clients to construct with authenticators matching each security scheme, and meet the interface.
	isA[csvresponse.Methods](csvresponse.NewClient("http://localhost", http.DefaultClient, csvClientToken))
	isA[example.ExampleMethods](example.NewExampleClient("http://localhost", http.DefaultClient, clientToken, clientToken, clientToken, clientToken, clientToken))
	isA[auth.Methods](auth.NewClient("http://localhost", http.DefaultClient, clientCookie, clientToken, clientToken, clientBasic, clientToken, clientWrap, clientWrap, clientWrap, clientRaw))

	os.Exit(0)
}
