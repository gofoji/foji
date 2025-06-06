package example

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type ExampleAuth struct{}

type Service struct{}

type TestInt int32

func (s *Service) GetExamples(ctx context.Context) (*Examples, error) {
	return nil, nil
}

func (s *Service) GetAuthComplex(ctx context.Context, user *ExampleAuth) error {
	return nil
}

func (s *Service) GetAuthSimple(ctx context.Context, user *ExampleAuth) error {
	return nil
}

func (s *Service) GetAuthSimpleMaybe(ctx context.Context, user *ExampleAuth) error {
	return nil
}

func (s *Service) GetAuthSimple2(ctx context.Context, user *ExampleAuth) error {
	return nil
}

func (s *Service) GetAuthSimple2Maybe(ctx context.Context, user *ExampleAuth) error {
	return nil
}

func (s *Service) GetAuthComplexMaybe(ctx context.Context, user *ExampleAuth) error {
	return nil
}

func (s *Service) AddForm(ctx context.Context, body AddFormRequest) (*FooBar, error) {
	return nil, nil
}

func (s *Service) AddMultipartForm(ctx context.Context, body AddMultipartFormRequest) (*FooBar, error) {
	return nil, nil
}

func (s *Service) AddInlinedAllOf(ctx context.Context, body AddInlinedAllOfRequest) (*FooBar, error) {
	return nil, nil
}

func (s *Service) AddInlinedBody(ctx context.Context, body AddInlinedBodyRequest) (*FooBar, error) {
	return nil, nil
}

func (s *Service) GetExampleParams(ctx context.Context, k1 string, k2 uuid.UUID, k3 time.Time, k4 int32, k5 int64, enumTest GetExampleParamsEnumTest) (*Example, error) {
	return nil, nil
}

func (s *Service) GetExampleOptional(ctx context.Context, k1 *string, k2 *uuid.UUID, k3 *time.Time, k4 *int32, k5 *int64, k5Default int64) (*Example, error) {
	return nil, nil
}

func (s *Service) GetExampleQuery(ctx context.Context, k1 string, k2 uuid.UUID, k3 time.Time, k4 int32, k5 int64) (*Example, error) {
	return nil, nil
}

func (s *Service) GetRawRequest(r *http.Request, vehicle GetRawRequestVehicle) (*Example, error) {
	return nil, nil
}

func (s *Service) GetRawResponse(ctx context.Context, w http.ResponseWriter, vehicle GetRawResponseVehicle) (*Example, error) {
	return nil, nil
}

func (s *Service) GetRawRequestResponse(r *http.Request, w http.ResponseWriter, vehicle GetRawRequestResponseVehicle) (*Example, error) {
	return nil, nil
}

func (s *Service) GetRawRequestResponseAndHeaders(r *http.Request, w http.ResponseWriter, vehicle GetRawRequestResponseAndHeadersVehicle) (*Example, http.Header, error) {
	return nil, nil, nil
}

func (s *Service) GetTest(ctx context.Context, vehicle GetTestVehicle, vehicleDefault GetTestVehicleDefault, playerID uuid.UUID, color ColorQuery, colorDefault ColorQueryDefault, season Season) (*Example, error) {
	return nil, nil
}

func (s *Service) NoResponse(ctx context.Context, body Foo) error {
	return nil
}

func (s *Service) HeaderResponse(ctx context.Context) (http.Header, error) {
	return nil, nil
}

func (s *Service) GetComplexSecurity(ctx context.Context, user *ExampleAuth) ([]TestInt, error) {
	return nil, nil
}
