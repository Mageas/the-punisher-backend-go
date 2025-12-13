package mock

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/mageas/the-punisher-backend/internal/dto"
)

type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) CreateUser(ctx context.Context, req dto.RequestUserDto) (*dto.ReturnUserDto, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ReturnUserDto), args.Error(1)
}
