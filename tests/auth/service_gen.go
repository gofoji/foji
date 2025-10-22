package auth

import (
	"context"
	"errors"

	"tests/example"
)

// GenService holds all Unsupported mock endpoints.  You can embed the mock service in real code.
// Simple example:
//
//	type Service struct {
//		GenService
//	}
//
// This allows you to mock the service for a quick start and eventually delete this stub.
type GenService struct{}

// ListAdminUsers
// List all users (admin only)
// Requires both API key AND bearer token with admin scope
func (s *GenService) ListAdminUsers(ctx context.Context, user *example.ExampleAuth) ([]User, error) {
	return nil, errors.ErrUnsupported
}

// QueryDataWithApiKey
// Query data with API key
// Accepts API key in header, query parameter, or cookie
func (s *GenService) QueryDataWithApiKey(ctx context.Context, user *example.ExampleAuth, query *string) error {
	return errors.ErrUnsupported
}

// ListDocuments
// List documents
// Requires basic authentication
func (s *GenService) ListDocuments(ctx context.Context, user *example.ExampleAuth) error {
	return errors.ErrUnsupported
}

// CreateDocument
// Create document
// Requires API key with bearer token
func (s *GenService) CreateDocument(ctx context.Context, user *example.ExampleAuth) error {
	return errors.ErrUnsupported
}

// GetDetailedProfile
// Get detailed profile
// Requires OpenID Connect authentication
func (s *GenService) GetDetailedProfile(ctx context.Context, user *example.ExampleAuth) error {
	return errors.ErrUnsupported
}

// GetProtectedResource
// Access protected resource
// Multiple authentication options:
// 1. OAuth2 with read scope, OR
// 2. API key (header) + Basic auth, OR
// 3. Bearer token + API
// key (cookie)
func (s *GenService) GetProtectedResource(ctx context.Context, user *example.ExampleAuth) error {
	return errors.ErrUnsupported
}

// GetPublicStatus
// Get API status
// Public endpoint that requires no authentication
func (s *GenService) GetPublicStatus(ctx context.Context) (*GetPublicStatusResponse, error) {
	return nil, errors.ErrUnsupported
}

// GetCurrentUser
// Get current user
// Requires either API key in header OR bearer token
func (s *GenService) GetCurrentUser(ctx context.Context, user *example.ExampleAuth) (*User, error) {
	return nil, errors.ErrUnsupported
}

// GetUserProfile
// Get user profile
// Requires bearer token authentication
func (s *GenService) GetUserProfile(ctx context.Context, user *example.ExampleAuth) (*User, error) {
	return nil, errors.ErrUnsupported
}
