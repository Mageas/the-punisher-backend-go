package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/handler"
	"github.com/mageas/the-punisher-backend/internal/service"
	servicemock "github.com/mageas/the-punisher-backend/internal/service/mock"
)

func TestCreateUserFromHandler(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockService := new(servicemock.MockUserService)
		h := handler.NewUserHandler(mockService)

		reqBody := dto.RequestUserDto{
			Email:     "test@example.com",
			FirstName: "Test",
			LastName:  "User",
			Password:  "password123",
		}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		expectedUser := &dto.ReturnUserDto{
			ID:        uuid.New(),
			Email:     reqBody.Email,
			FirstName: reqBody.FirstName,
			LastName:  reqBody.LastName,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mockService.On("CreateUser", mock.Anything, reqBody).Return(expectedUser, nil)

		// Setup Chi router to handle context properly if needed, but direct call is easier
		// However, handler expects *http.Request
		h.CreateUser(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		mockService := new(servicemock.MockUserService)
		h := handler.NewUserHandler(mockService)

		req, _ := http.NewRequest("POST", "/users", bytes.NewBufferString("invalid-json"))
		w := httptest.NewRecorder()

		h.CreateUser(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("EmailAlreadyExists", func(t *testing.T) {
		mockService := new(servicemock.MockUserService)
		h := handler.NewUserHandler(mockService)

		reqBody := dto.RequestUserDto{
			Email:    "existing@example.com",
			Password: "password",
		}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		mockService.On("CreateUser", mock.Anything, reqBody).Return(nil, service.ErrEmailAlreadyExists)

		h.CreateUser(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("ValidationError", func(t *testing.T) {
		mockService := new(servicemock.MockUserService)
		h := handler.NewUserHandler(mockService)

		reqBody := dto.RequestUserDto{} // Missing fields
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		// Mocking a validation error
		validate := validator.New()
		err := validate.Struct(reqBody)

		mockService.On("CreateUser", mock.Anything, reqBody).Return(nil, err)

		h.CreateUser(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("InternalServerError", func(t *testing.T) {
		mockService := new(servicemock.MockUserService)
		h := handler.NewUserHandler(mockService)

		reqBody := dto.RequestUserDto{
			Email: "error@example.com",
		}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		mockService.On("CreateUser", mock.Anything, reqBody).Return(nil, errors.New("unexpected error"))

		h.CreateUser(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)
	})
}
