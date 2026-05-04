package httpapi

import "Web_lab/internal/users/models"

// ErrorResponse стандартная ошибка API.
type ErrorResponse struct {
	Error string `json:"error" example:"invalid request"`
}

// MessageResponse простой ответ с сообщением.
type MessageResponse struct {
	Message string `json:"message" example:"operation completed"`
}

// RegisterResponse ответ после регистрации.
type RegisterResponse struct {
	Message string              `json:"message" example:"user registered"`
	User    models.UserResponse `json:"user"`
}

// ForgotPasswordResponse ответ на запрос сброса пароля.
type ForgotPasswordResponse struct {
	Message    string `json:"message" example:"if email exists, reset link sent"`
	ResetToken string `json:"reset_token,omitempty" example:"9f6f82d3d2d64f8d9ff44f3d0f4fb7fd"`
}

// OAuthInitResponse ответ с URL авторизации для клиентов без browser redirect.
type OAuthInitResponse struct {
	Message          string `json:"message" example:"open authorization_url in browser to continue OAuth flow"`
	AuthorizationURL string `json:"authorization_url" example:"https://oauth.yandex.ru/authorize?response_type=code&client_id=your_client_id"`
}

// HealthResponse ответ healthcheck.
type HealthResponse struct {
	Status string `json:"status" example:"ok"`
}
