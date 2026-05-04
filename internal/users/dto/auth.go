package dto

type RegisterDTO struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required,min=4" example:"mypassword"`
}

type LoginDTO struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required" example:"mypassword"`
}

// LoginResponse ответ при успешном входе.
type LoginResponse struct {
	AccessToken  string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." description:"Access token"`
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." description:"Refresh token"`
	TokenType    string `json:"token_type" example:"Bearer" description:"Тип токена"`
	ExpiresIn    int64  `json:"expires_in" example:"3600" description:"Время жизни access token в секундах"`
}

// TokenPair пара токенов.
type TokenPair struct {
	AccessToken  string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// RefreshRequest запрос на обновление токенов.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// ForgotPasswordRequest запрос на сброс пароля.
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email" example:"user@example.com"`
}

// ForgotPasswordResponse ответ с токеном сброса.
type ForgotPasswordResponse struct {
	ResetToken string `json:"reset_token" example:"abc123def456"`
	Message    string `json:"message" example:"if email exists, reset link sent"`
}

// ResetPasswordRequest запрос на сброс пароля.
type ResetPasswordRequest struct {
	Token    string `json:"token" binding:"required" example:"abc123def456"`
	Password string `json:"password" binding:"required,min=4" example:"newpassword123"`
}
