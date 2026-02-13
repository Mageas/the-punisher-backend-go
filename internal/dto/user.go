package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

type RequestUserDto struct {
	Email     string `json:"email" validate:"required,email"`
	FirstName string `json:"first_name" validate:"required,min=2,max=70"`
	LastName  string `json:"last_name" validate:"required,min=2,max=70"`
	Password  string `json:"password" validate:"required,min=8"`
}

type ReturnUserDto struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func UserFromRepository(u *repository.CreateUserRow) *ReturnUserDto {
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

func UserListFromRepository(users []*repository.CreateUserRow) []*ReturnUserDto {
	dtos := make([]*ReturnUserDto, 0, len(users))

	for _, u := range users {
		if dto := UserFromRepository(u); dto != nil {
			dtos = append(dtos, dto)
		}
	}

	return dtos
}
