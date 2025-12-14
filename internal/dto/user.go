package dto

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

var validate = validator.New()

type RequestUserDto struct {
	Email     string `json:"email" validate:"required,email"`
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name" validate:"required"`
	Password  string `json:"password" validate:"required,min=8"`
}

func (r *RequestUserDto) Validate() error {
	return validate.Struct(r)
}

type ReturnUserDto struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func FromRepository(u *repository.CreateUserRow) *ReturnUserDto {
	if u == nil {
		return nil
	}

	return &ReturnUserDto{
		ID:        u.ID,
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

func FromRepositoryList(users []*repository.CreateUserRow) []*ReturnUserDto {
	dtos := make([]*ReturnUserDto, 0, len(users))

	for _, u := range users {
		if dto := FromRepository(u); dto != nil {
			dtos = append(dtos, dto)
		}
	}

	return dtos
}
