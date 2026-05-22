package handler

import (
	"net/http"
	"strings"
	"time"

	"Web_lab/internal/httpapi"
	"Web_lab/internal/users/dto"
	"Web_lab/internal/users/middleware"
	userModels "Web_lab/internal/users/models"
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
	stateStore  map[string]string
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

func (h *AuthHandler) CreateAuthMiddleware() gin.HandlerFunc {
	return middleware.AuthMiddleware(h.service.ValidateAccessToken)
}

// Register регистрирует пользователя по email и паролю.
// @Summary      Регистрация пользователя
// @Description  Создает нового пользователя и возвращает его публичный профиль
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body dto.RegisterDTO true "Данные для регистрации"
// @Success      201 {object} httpapi.RegisterResponse "Пользователь успешно зарегистрирован"
// @Failure      400 {object} httpapi.ErrorResponse "Неверный формат запроса"
// @Failure      409 {object} httpapi.ErrorResponse "Email уже существует"
// @Failure      500 {object} httpapi.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, httpapi.ErrorResponse{Error: err.Error()})
		return
	}

	user, err := h.service.Register(req)
	if err != nil {
		if err.Error() == "email already exists" {
			c.JSON(http.StatusConflict, httpapi.ErrorResponse{Error: err.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, httpapi.ErrorResponse{Error: "failed to register"})
		return
	}

	var responseUser userModels.UserResponse = user.ToResponse()
	c.JSON(http.StatusCreated, httpapi.RegisterResponse{
		Message: "user registered",
		User:    responseUser,
	})
}

// Login выполняет вход и устанавливает access/refresh токены в HttpOnly cookies.
// @Summary      Вход пользователя
// @Description  Аутентифицирует пользователя и устанавливает access_token и refresh_token в HttpOnly cookies
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body dto.LoginDTO true "Учетные данные"
// @Success      200 {object} httpapi.MessageResponse "Успешный вход"
// @Failure      400 {object} httpapi.ErrorResponse "Неверный формат запроса"
// @Failure      401 {object} httpapi.ErrorResponse "Неверный email или пароль"
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, httpapi.ErrorResponse{Error: err.Error()})
		return
	}

	accessToken, refreshToken, err := h.service.Login(req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, httpapi.ErrorResponse{Error: err.Error()})
		return
	}

	h.setCookies(c, accessToken, refreshToken)
	c.JSON(http.StatusOK, httpapi.MessageResponse{Message: "logged in"})
}

// Whoami возвращает данные текущего пользователя.
// @Summary      Информация о себе
// @Description  Возвращает публичный профиль авторизованного пользователя
// @Tags         auth
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} userModels.UserResponse "Профиль пользователя"
// @Failure      401 {object} httpapi.ErrorResponse "Не авторизован"
// @Failure      404 {object} httpapi.ErrorResponse "Пользователь не найден"
// @Router       /auth/whoami [get]
func (h *AuthHandler) Whoami(c *gin.Context) {
	userIDStr, ok := middleware.GetUserID(c.Request.Context())
	if !ok {
		c.JSON(http.StatusUnauthorized, httpapi.ErrorResponse{Error: "unauthorized"})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, httpapi.ErrorResponse{Error: "invalid user id"})
		return
	}

	response, err := h.service.Whoami(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, httpapi.ErrorResponse{Error: "user not found"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// Logout завершает текущую сессию пользователя.
// @Summary      Выход
// @Description  Отзывает текущие access/refresh токены и очищает cookies
// @Tags         auth
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} httpapi.MessageResponse "Успешный выход"
// @Failure      401 {object} httpapi.ErrorResponse "Не авторизован"
// @Failure      500 {object} httpapi.ErrorResponse "Ошибка выхода"
// @Router       /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	userIDStr, ok := middleware.GetUserID(c.Request.Context())
	if !ok {
		c.JSON(http.StatusUnauthorized, httpapi.ErrorResponse{Error: "unauthorized"})
		return
	}

	userID, _ := uuid.Parse(userIDStr)
	accessToken, _ := c.Cookie("access_token")
	refreshToken, _ := c.Cookie("refresh_token")

	if err := h.service.Logout(userID, accessToken, refreshToken); err != nil {
		c.JSON(http.StatusInternalServerError, httpapi.ErrorResponse{Error: "failed to logout"})
		return
	}

	h.clearCookies(c)
	c.JSON(http.StatusOK, httpapi.MessageResponse{Message: "logged out"})
}

// LogoutAll завершает все активные сессии пользователя.
// @Summary      Выход со всех устройств
// @Description  Отзывает все сохраненные токены пользователя и очищает cookies
// @Tags         auth
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} httpapi.MessageResponse "Выход со всех устройств выполнен"
// @Failure      401 {object} httpapi.ErrorResponse "Не авторизован"
// @Failure      500 {object} httpapi.ErrorResponse "Ошибка выхода"
// @Router       /auth/logout-all [post]
func (h *AuthHandler) LogoutAll(c *gin.Context) {
	userIDStr, ok := middleware.GetUserID(c.Request.Context())
	if !ok {
		c.JSON(http.StatusUnauthorized, httpapi.ErrorResponse{Error: "unauthorized"})
		return
	}

	userID, _ := uuid.Parse(userIDStr)
	if err := h.service.LogoutAll(userID); err != nil {
		c.JSON(http.StatusInternalServerError, httpapi.ErrorResponse{Error: "failed to logout"})
		return
	}

	h.clearCookies(c)
	c.JSON(http.StatusOK, httpapi.MessageResponse{Message: "logged out from all devices"})
}

// OAuthInit начинает OAuth авторизацию через провайдера.
// @Summary      OAuth авторизация
// @Description  При обычном открытии в браузере делает redirect на страницу авторизации OAuth-провайдера. Для Swagger UI и других fetch-клиентов возвращает URL авторизации в JSON.
// @Tags         oauth
// @Produce      json
// @Param        provider path string true "OAuth-провайдер" Enums(yandex)
// @Success      200 {object} httpapi.OAuthInitResponse "URL авторизации для Swagger UI и fetch-клиентов"
// @Success      302 {string} string "Редирект на страницу провайдера"
// @Failure      400 {object} httpapi.ErrorResponse "Неподдерживаемый провайдер"
// @Failure      500 {object} httpapi.ErrorResponse "Ошибка генерации state"
// @Router       /auth/oauth/{provider} [get]
func (h *AuthHandler) OAuthInit(c *gin.Context) {
	provider := c.Param("provider")

	if h.oauthConfig == nil {
		c.JSON(http.StatusInternalServerError, httpapi.ErrorResponse{Error: "OAuth not configured"})
		return
	}

	if provider != "yandex" {
		c.JSON(http.StatusBadRequest, httpapi.ErrorResponse{Error: "unsupported provider"})
		return
	}

	state, err := oauth.GenerateState()
	if err != nil {
		c.JSON(http.StatusInternalServerError, httpapi.ErrorResponse{Error: "failed to generate state"})
		return
	}

	h.stateStore[state] = time.Now().Format(time.RFC3339)
	h.cleanupState()

	authURL := oauth.GetYandexAuthURL(h.oauthConfig, state)

	if shouldReturnOAuthURLJSON(c) {
		c.JSON(http.StatusOK, httpapi.OAuthInitResponse{
			Message:          "open authorization_url in browser to continue OAuth flow",
			AuthorizationURL: authURL,
		})
		return
	}

	c.Redirect(http.StatusFound, authURL)
}

// OAuthCallback завершает OAuth flow и создает локальную сессию.
// @Summary      OAuth callback
// @Description  Проверяет state, получает данные пользователя у OAuth-провайдера и выдает локальные cookies
// @Tags         oauth
// @Produce      json
// @Param        provider path string true "OAuth-провайдер" Enums(yandex)
// @Param        code query string true "Код авторизации"
// @Param        state query string true "Параметр защиты от CSRF"
// @Success      302 {string} string "Редирект после успешной авторизации"
// @Failure      400 {object} httpapi.ErrorResponse "Отсутствует code или state"
// @Failure      403 {object} httpapi.ErrorResponse "Неверный state"
// @Failure      502 {object} httpapi.ErrorResponse "Ошибка получения данных от провайдера"
// @Router       /auth/oauth/{provider}/callback [get]
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
		c.JSON(http.StatusBadRequest, httpapi.ErrorResponse{Error: "missing code or state"})
		return
	}

	if _, exists := h.stateStore[state]; !exists {
		c.JSON(http.StatusForbidden, httpapi.ErrorResponse{Error: "invalid state"})
		return
	}
	delete(h.stateStore, state)

	if provider != "yandex" {
		c.JSON(http.StatusBadRequest, httpapi.ErrorResponse{Error: "unsupported provider"})
		return
	}

	yandexToken, err := oauth.ExchangeCodeForToken(h.oauthConfig, code)
	if err != nil {
		c.JSON(http.StatusBadGateway, httpapi.ErrorResponse{Error: "failed to exchange token"})
		return
	}

	userInfo, err := oauth.GetYandexUserInfo(yandexToken)
	if err != nil {
		c.JSON(http.StatusBadGateway, httpapi.ErrorResponse{Error: "failed to get user info"})
		return
	}

	user, err := oauth.FindOrCreateUser(h.userRepo, userInfo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, httpapi.ErrorResponse{Error: "failed to create user"})
		return
	}

	accessToken, refreshToken, err := h.service.CreateTokens(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, httpapi.ErrorResponse{Error: "failed to create tokens"})
		return
	}

	if err := h.service.SaveTokenPair(user.ID, accessToken, refreshToken); err != nil {
		c.JSON(http.StatusInternalServerError, httpapi.ErrorResponse{Error: "failed to save session"})
		return
	}

	h.setCookies(c, accessToken, refreshToken)
	c.Redirect(http.StatusFound, "/")
}

// Refresh обновляет access и refresh токены по refresh cookie.
// @Summary      Обновление токенов
// @Description  Обновляет пару access/refresh токенов по refresh_token из cookie
// @Tags         auth
// @Produce      json
// @Success      200 {object} httpapi.MessageResponse "Токены обновлены"
// @Failure      401 {object} httpapi.ErrorResponse "Отсутствует или неверный refresh токен"
// @Router       /auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, httpapi.ErrorResponse{Error: "missing refresh token"})
		return
	}

	newAccessToken, newRefreshToken, err := h.service.RefreshTokens(refreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, httpapi.ErrorResponse{Error: "invalid or expired token"})
		return
	}

	h.setCookies(c, newAccessToken, newRefreshToken)
	c.JSON(http.StatusOK, httpapi.MessageResponse{Message: "tokens refreshed"})
}

// ForgotPassword создает одноразовый токен сброса пароля.
// @Summary      Запрос сброса пароля
// @Description  Создает токен сброса пароля для указанного email
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body dto.ForgotPasswordRequest true "Email пользователя"
// @Success      200 {object} httpapi.ForgotPasswordResponse "Токен сброса создан"
// @Failure      400 {object} httpapi.ErrorResponse "Неверный формат запроса"
// @Failure      500 {object} httpapi.ErrorResponse "Ошибка создания токена"
// @Router       /auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req dto.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, httpapi.ErrorResponse{Error: err.Error()})
		return
	}

	resetToken, err := h.service.ForgotPassword(req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, httpapi.ErrorResponse{Error: "failed to create reset token"})
		return
	}

	c.JSON(http.StatusOK, httpapi.ForgotPasswordResponse{
		Message:    "if email exists, reset link sent",
		ResetToken: resetToken,
	})
}

// ResetPassword устанавливает новый пароль по reset token.
// @Summary      Сброс пароля
// @Description  Устанавливает новый пароль по валидному токену сброса
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body dto.ResetPasswordRequest true "Токен сброса и новый пароль"
// @Success      200 {object} httpapi.MessageResponse "Пароль успешно сброшен"
// @Failure      400 {object} httpapi.ErrorResponse "Неверный формат запроса или токен"
// @Router       /auth/reset-password [post]
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req dto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, httpapi.ErrorResponse{Error: err.Error()})
		return
	}

	if err := h.service.ResetPassword(req.Token, req.Password); err != nil {
		c.JSON(http.StatusBadRequest, httpapi.ErrorResponse{Error: "invalid or expired reset token"})
		return
	}

	c.JSON(http.StatusOK, httpapi.MessageResponse{Message: "password reset"})
}

func (h *AuthHandler) setCookies(c *gin.Context, accessToken, refreshToken string) {
	accessTTL := h.service.AccessTokenTTL()
	refreshTTL := h.service.RefreshTokenTTL()

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Path:     "/",
		MaxAge:   maxAgeFromDuration(accessTTL),
		Expires:  time.Now().Add(accessTTL),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		MaxAge:   maxAgeFromDuration(refreshTTL),
		Expires:  time.Now().Add(refreshTTL),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func (h *AuthHandler) clearCookies(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func (h *AuthHandler) cleanupState() {
	expired := time.Now().Add(-15 * time.Minute)
	for state, ts := range h.stateStore {
		parsed, err := time.Parse(time.RFC3339, ts)
		if err != nil || parsed.Before(expired) {
			delete(h.stateStore, state)
		}
	}
}

func maxAgeFromDuration(ttl time.Duration) int {
	seconds := int(ttl / time.Second)
	if seconds <= 0 {
		return 1
	}

	return seconds
}

func shouldReturnOAuthURLJSON(c *gin.Context) bool {
	if strings.Contains(c.GetHeader("Accept"), "application/json") {
		return true
	}

	return strings.EqualFold(c.GetHeader("Sec-Fetch-Mode"), "cors")
}
