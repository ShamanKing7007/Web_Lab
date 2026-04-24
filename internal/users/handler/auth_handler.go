package handler

import (
	"net/http"
	"time"

	"Web_lab/internal/users/dto"
	"Web_lab/internal/users/middleware"
	"Web_lab/internal/users/oauth"
	"Web_lab/internal/users/repository"
	"Web_lab/internal/users/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthHandler struct {
	service     *service.AuthService
	userRepo    *repository.UserRepository
	oauthConfig *oauth.OAuthConfig
	stateStore  map[string]string // state → timestamp (CSRF защита)
}

func NewAuthHandler(
	svc *service.AuthService,
	userRepo *repository.UserRepository,
	oauthConfig *oauth.OAuthConfig,
) *AuthHandler {
	return &AuthHandler{
		service:     svc,
		userRepo:    userRepo,
		oauthConfig: oauthConfig,
		stateStore:  make(map[string]string),
	}
}

// CreateAuthMiddleware создаёт middleware для проверки JWT из cookie
func (h *AuthHandler) CreateAuthMiddleware(jwtAccessSecret string) gin.HandlerFunc {
	return middleware.AuthMiddleware(jwtAccessSecret)
}

// Register POST /auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.service.Register(req)
	if err != nil {
		if err.Error() == "email already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "user registered",
		"user":    user,
	})
}

// Login POST /auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	accessToken, refreshToken, err := h.service.Login(req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	h.setCookies(c, accessToken, refreshToken)
	c.JSON(http.StatusOK, gin.H{"message": "logged in"})
}

// Whoami GET /auth/whoami
func (h *AuthHandler) Whoami(c *gin.Context) {
	userIDStr, ok := middleware.GetUserID(c.Request.Context())
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	user, err := h.service.Whoami(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// Logout POST /auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	userIDStr, ok := middleware.GetUserID(c.Request.Context())
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID, _ := uuid.Parse(userIDStr)

	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing refresh token"})
		return
	}

	err = h.service.Logout(userID, refreshToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to logout"})
		return
	}

	h.clearCookies(c)
	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

// LogoutAll POST /auth/logout-all
func (h *AuthHandler) LogoutAll(c *gin.Context) {
	userIDStr, ok := middleware.GetUserID(c.Request.Context())
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID, _ := uuid.Parse(userIDStr)

	err := h.service.LogoutAll(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to logout"})
		return
	}

	h.clearCookies(c)
	c.JSON(http.StatusOK, gin.H{"message": "logged out from all devices"})
}

// OAuthInit GET /auth/oauth/:provider — редирект на провайдера
func (h *AuthHandler) OAuthInit(c *gin.Context) {
	provider := c.Param("provider")

	if h.oauthConfig == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "OAuth not configured"})
		return
	}

	if provider != "yandex" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported provider"})
		return
	}

	state, err := oauth.GenerateState()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate state"})
		return
	}

	// Сохраняем state для проверки (с таймаутом 10 мин)
	h.stateStore[state] = time.Now().Format(time.RFC3339)

	// Чистим старые state (старше 15 мин)
	h.cleanupState()

	authURL := oauth.GetYandexAuthURL(h.oauthConfig, state)
	c.Redirect(http.StatusFound, authURL)
}

// OAuthCallback GET /auth/oauth/:provider/callback — обработка ответа
func (h *AuthHandler) OAuthCallback(c *gin.Context) {
	provider := c.Param("provider")
	code := c.Query("code")
	state := c.Query("state")
	errorParam := c.Query("error")

	if errorParam != "" {
		errorDesc := c.Query("error_description")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             errorParam,
			"error_description": errorDesc,
		})
		return
	}

	if code == "" || state == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing code or state"})
		return
	}

	// Проверяем state (CSRF защита)
	if _, exists := h.stateStore[state]; !exists {
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid state"})
		return
	}
	delete(h.stateStore, state)

	if provider != "yandex" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported provider"})
		return
	}

	// Обмениваем code на access token
	yandexToken, err := oauth.ExchangeCodeForToken(h.oauthConfig, code)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to exchange token"})
		return
	}

	// Получаем данные пользователя
	userInfo, err := oauth.GetYandexUserInfo(yandexToken)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to get user info"})
		return
	}

	// Ищем или создаём пользователя
	user, err := oauth.FindOrCreateUser(h.userRepo, userInfo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	// Генерируем локальные JWT
	accessToken, refreshToken, err := h.service.CreateTokens(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create tokens"})
		return
	}

	// Сохраняем refresh token в БД
	if err := h.service.SaveRefreshToken(user.ID, refreshToken); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save session"})
		return
	}

	// Устанавливаем cookies и редиректим
	h.setCookies(c, accessToken, refreshToken)
	c.Redirect(http.StatusFound, "/")
}

// setCookies устанавливает secure cookies
func (h *AuthHandler) setCookies(c *gin.Context, accessToken, refreshToken string) {
	c.SetCookie("access_token", accessToken, 900, "/", "", false, true)
	c.SetCookie("refresh_token", refreshToken, 604800, "/", "", false, true)
}

// clearCookies удаляет cookies
func (h *AuthHandler) clearCookies(c *gin.Context) {
	c.SetCookie("access_token", "", -1, "/", "", false, true)
	c.SetCookie("refresh_token", "", -1, "/", "", false, true)
}

// cleanupState удаляет просроченные state
func (h *AuthHandler) cleanupState() {
	expired := time.Now().Add(-15 * time.Minute)
	for state, ts := range h.stateStore {
		parsed, err := time.Parse(time.RFC3339, ts)
		if err != nil || parsed.Before(expired) {
			delete(h.stateStore, state)
		}
	}
}

// Refresh POST /auth/refresh
func (h *AuthHandler) Refresh(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing refresh token"})
		return
	}

	newAccessToken, newRefreshToken, err := h.service.RefreshTokens(refreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
		return
	}

	h.setCookies(c, newAccessToken, newRefreshToken)
	c.JSON(http.StatusOK, gin.H{"message": "tokens refreshed"})
}

// ForgotPassword POST /auth/forgot-password
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resetToken, err := h.service.ForgotPassword(req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create reset token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "if email exists, reset link sent",
		"reset_token": resetToken,
	})
}

// ResetPassword POST /auth/reset-password
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req struct {
		Token    string `json:"token" binding:"required"`
		Password string `json:"password" binding:"required,min=4"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.ResetPassword(req.Token, req.Password); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or expired reset token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "password reset"})
}
