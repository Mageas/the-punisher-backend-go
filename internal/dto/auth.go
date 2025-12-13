package dto

type LoginRequestDto struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginResponseDto struct {
	AccessToken string `json:"access_token"`
}
