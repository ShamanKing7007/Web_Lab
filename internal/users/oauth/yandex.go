package oauth

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"Web_lab/internal/users/crypto"
	"Web_lab/internal/users/models"
	"Web_lab/internal/users/repository"

	"github.com/google/uuid"
)

const (
	yandexAuthURL  = "https://oauth.yandex.ru/authorize"
	yandexTokenURL = "https://oauth.yandex.ru/token"
	yandexInfoURL  = "https://login.yandex.ru/info"
)

// YandexUserInfo — данные пользователя от Yandex
type YandexUserInfo struct {
	ID     string `json:"id"`
	Email  string `json:"default_email"`
	Login  string `json:"login"`
	Name   string `json:"real_name"`
	Avatar string `json:"default_avatar_id"`
}

// GenerateState генерирует случайный state для CSRF защиты
func GenerateState() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// GetYandexAuthURL формирует URL для редиректа на Yandex
func GetYandexAuthURL(cfg *OAuthConfig, state string) string {
	params := url.Values{}
	params.Set("response_type", "code")
	params.Set("client_id", cfg.ClientID)
	params.Set("redirect_uri", cfg.CallbackURL)
	params.Set("state", state)
	return fmt.Sprintf("%s?%s", yandexAuthURL, params.Encode())
}

// ExchangeCodeForToken обменивает code на access token Yandex
func ExchangeCodeForToken(cfg *OAuthConfig, code string) (string, error) {
	resp, err := http.PostForm(yandexTokenURL, url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"client_id":     {cfg.ClientID},
		"client_secret": {cfg.ClientSecret},
	})
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("yandex token error: %d - %s", resp.StatusCode, string(body))
	}

	var result struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode token response: %w", err)
	}

	if result.AccessToken == "" {
		return "", fmt.Errorf("empty access token from Yandex")
	}

	return result.AccessToken, nil
}

// GetYandexUserInfo получает данные пользователя от Yandex
func GetYandexUserInfo(accessToken string) (*YandexUserInfo, error) {
	req, err := http.NewRequest("GET", yandexInfoURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "OAuth "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("yandex userinfo error: %d - %s", resp.StatusCode, string(body))
	}

	var info YandexUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("failed to decode userinfo: %w", err)
	}

	if info.ID == "" {
		return nil, fmt.Errorf("empty user ID from Yandex")
	}

	return &info, nil
}

// FindOrCreateUser ищет или создаёт пользователя по данным из Yandex
func FindOrCreateUser(userRepo *repository.UserRepository, yandexInfo *YandexUserInfo) (*models.User, error) {
	// 1. Ищем пользователя по Yandex ID
	user, err := userRepo.FindByYandexID(yandexInfo.ID)
	if err == nil {
		return user, nil
	}

	// 2. Если не найден — ищем по email
	user, err = userRepo.FindByEmail(yandexInfo.Email)
	if err == nil {
		// Пользователь найден по email — привязываем Yandex ID
		if user.YandexID == nil {
			yandexID := yandexInfo.ID
			user.YandexID = &yandexID
			if updateErr := userRepo.Update(user); updateErr != nil {
				return nil, fmt.Errorf("failed to link Yandex ID: %w", updateErr)
			}
		}
		return user, nil
	}

	// 3. Создаём нового пользователя
	hash, salt, hashErr := crypto.HashPassword(uuid.New().String())
	if hashErr != nil {
		return nil, fmt.Errorf("failed to hash random password: %w", hashErr)
	}

	yandexID := yandexInfo.ID
	user = &models.User{
		ID:           uuid.New(),
		Email:        yandexInfo.Email,
		PasswordHash: hash,
		Salt:         salt,
		YandexID:     &yandexID,
	}

	if createErr := userRepo.Create(user); createErr != nil {
		return nil, fmt.Errorf("failed to create user: %w", createErr)
	}

	return user, nil
}
