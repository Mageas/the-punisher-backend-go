package domain

import "github.com/google/uuid"

type UserCredentials struct {
	ID           uuid.UUID
	PasswordHash string
}
