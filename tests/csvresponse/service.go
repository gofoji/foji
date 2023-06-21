package csvresponse

import (
	"context"
	"io"
)

type Service struct{}

func (s *Service) GetByteCsv(ctx context.Context) ([]byte, error) {
	return nil, nil
}

func (s *Service) GetReaderCsv(ctx context.Context) (io.Reader, error) {
	return nil, nil
}

func (s *Service) GetStringCsv(ctx context.Context) (*string, error) {
	return nil, nil
}
