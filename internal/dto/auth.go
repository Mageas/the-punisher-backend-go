package dto

type LoginRequestDto struct {
	Email      string `json:"email" validate:"required,email"`
	Password   string `json:"password" validate:"required,min=8"`
	RemoteAddr string `json:"-"`
}

type LoginResponseDto struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"-"`
}

type RefreshResponseDto struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"-"`
}

type ChangePasswordRequestDto struct {
	CurrentPassword string `json:"current_password" validate:"required,min=8"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" validate:"required"`
}

type ChangePasswordResponseDto struct {
	Status string `json:"status"`
}

type ForgotPasswordRequestDto struct {
	Email string `json:"email" validate:"required,email"`
}

type ForgotPasswordResponseDto struct {
	Status string `json:"status"`
}

type ResetPasswordRequestDto struct {
	Token           string `json:"token" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" validate:"required"`
}

type ResetPasswordResponseDto struct {
	Status string `json:"status"`
}

type RegisterStatusResponseDto struct {
	RegisterAllowed bool `json:"register_allowed"`
}

type ConfirmEmailResponseDto struct {
	Status string `json:"status"`
}

type ResendConfirmEmailRequestDto struct {
	Email string `json:"email" validate:"required,email"`
}

type ResendConfirmEmailResponseDto struct {
	Status string `json:"status"`
}
