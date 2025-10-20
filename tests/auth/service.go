package auth

import (
	"context"

	"tests/example"
)

type Service struct {
	GenService
}

func (s *Service) ListAdminUsers(ctx context.Context, user *example.ExampleAuth) ([]User, error) {
	return nil, nil
}

func (s *Service) QueryDataWithApiKey(ctx context.Context, user *example.ExampleAuth, query *string) error {
	return nil
}

func (s *Service) ListDocuments(ctx context.Context, user *example.ExampleAuth) error {
	return nil
}

func (s *Service) CreateDocument(ctx context.Context, user *example.ExampleAuth) error {
	return nil
}

func (s *Service) GetDetailedProfile(ctx context.Context, user *example.ExampleAuth) error {
	return nil
}

func (s *Service) GetProtectedResource(ctx context.Context, user *example.ExampleAuth) error {
	return nil
}

func (s *Service) GetPublicStatus(ctx context.Context) (*GetPublicStatusResponse, error) {
	return nil, nil
}

func (s *Service) GetCurrentUser(ctx context.Context, user *example.ExampleAuth) (*User, error) {
	return nil, nil
}

func (s *Service) GetUserProfile(ctx context.Context, user *example.ExampleAuth) (*User, error) {
	return nil, nil
}
