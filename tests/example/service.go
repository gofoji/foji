package example

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Service struct{}

func (s *Service) GetExamples(ctx context.Context) (*Examples, error) {
	return nil, nil
}

func (s *Service) GetExampleParams(ctx context.Context, k1 string, k2 uuid.UUID, k3 time.Time, k4 int32, k5 int64) (*Example, error) {
	return nil, nil
}

func (s *Service) GetExampleOptional(ctx context.Context, k1 *string, k2 *uuid.UUID, k3 *time.Time, k4 *int32, k5 *int64, k5Default int64) (*Example, error) {
	return nil, nil
}

func (s *Service) GetExampleQuery(ctx context.Context, k1 string, k2 uuid.UUID, k3 time.Time, k4 int32, k5 int64) (*Example, error) {
	return nil, nil
}

func (s *Service) GetTest(ctx context.Context, vehicle GetTestVehicle, vehicleDefault GetTestVehicleDefault, playerID uuid.UUID, color ColorQuery, colorDefault ColorQueryDefault, season Season) (*Example, error) {
	return nil, nil
}
