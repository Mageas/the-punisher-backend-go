package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mageas/the-punisher-backend/internal/domain"
	"github.com/mageas/the-punisher-backend/internal/dto"
	repositorymock "github.com/mageas/the-punisher-backend/internal/repository/mock"
	"github.com/mageas/the-punisher-backend/internal/service"
)

func TestCreateUserFromService(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(repositorymock.MockUserRepository)
		userService := service.NewUserService(mockRepo)

		ctx := context.Background()
		req := dto.RequestUserDto{
			Email:     "test@example.com",
			FirstName: "Test",
			LastName:  "User",
			Password:  "password123",
		}

		mockRepo.On("EmailExists", ctx, req.Email).Return(false, nil)
		mockRepo.On("Create", ctx, mock.AnythingOfType("*domain.User")).Return(&domain.User{
			ID:        uuid.New(),
			Email:     req.Email,
			FirstName: req.FirstName,
			LastName:  req.LastName,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}, nil)

		user, err := userService.CreateUser(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, req.Email, user.Email)
		mockRepo.AssertExpectations(t)
	})

	t.Run("EmailAlreadyExists", func(t *testing.T) {
		mockRepo := new(repositorymock.MockUserRepository)
		userService := service.NewUserService(mockRepo)

		ctx := context.Background()
		req := dto.RequestUserDto{
			Email: "existing@example.com",
		}

		mockRepo.On("EmailExists", ctx, req.Email).Return(true, nil)

		user, err := userService.CreateUser(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Equal(t, service.ErrEmailAlreadyExists, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("RepositoryError", func(t *testing.T) {
		mockRepo := new(repositorymock.MockUserRepository)
		userService := service.NewUserService(mockRepo)

		ctx := context.Background()
		req := dto.RequestUserDto{
			Email:    "error@example.com",
			Password: "password",
		}

		expectedErr := errors.New("db error")

		mockRepo.On("EmailExists", ctx, req.Email).Return(false, nil)
		mockRepo.On("Create", ctx, mock.AnythingOfType("*domain.User")).Return(nil, expectedErr)

		user, err := userService.CreateUser(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Equal(t, expectedErr, err)
		mockRepo.AssertExpectations(t)
	})
}
