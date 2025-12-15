package dto

type LoginRequestDto struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type LoginResponseDto struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"-"`
}

type RefreshResponseDto struct {
	AccessToken string `json:"access_token"`
}
