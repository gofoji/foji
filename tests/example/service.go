package example

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type Service struct{}

func (s *Service) GetExamples(ctx context.Context) (*Examples, error) {
	return nil, nil
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

func (s *Service) GetTest(ctx context.Context, w http.ResponseWriter, r *http.Request, vehicle GetTestVehicle, vehicleDefault GetTestVehicleDefault, playerID uuid.UUID, color ColorQuery, colorDefault ColorQueryDefault, season Season) (*Example, error) {
	return nil, nil
}
