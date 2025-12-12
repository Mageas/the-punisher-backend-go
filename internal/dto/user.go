package dto

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type RequestUserDto struct {
	Email     string `json:"email" validate:"required,email"`
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name" validate:"required"`
	Password  string `json:"password" validate:"required,min=8"`
}

func (r *RequestUserDto) Validate() error {
	validate := validator.New()
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
