package service

import (
	"errors"
	"time"

	"Web_lab/internal/users/crypto"
	"Web_lab/internal/users/dto"
	"Web_lab/internal/users/models"
	"Web_lab/internal/users/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuthService struct {
	userRepo         *repository.UserRepository
	tokenRepo        *repository.TokenRepository
	jwtAccessSecret  string
	jwtRefreshSecret string
}

func NewAuthService(
	userRepo *repository.UserRepository,
	tokenRepo *repository.TokenRepository,
	jwtAccessSecret string,
	jwtRefreshSecret string,
) *AuthService {
	return &AuthService{
		userRepo:         userRepo,
		tokenRepo:        tokenRepo,
		jwtAccessSecret:  jwtAccessSecret,
		jwtRefreshSecret: jwtRefreshSecret,
	}
}

// Register — регистрация нового пользователя
func (s *AuthService) Register(dto dto.RegisterDTO) (*models.User, error) {
	_, err := s.userRepo.FindByEmail(dto.Email)
	if err == nil {
		return nil, errors.New("email already exists")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	hash, salt, err := crypto.HashPassword(dto.Password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		ID:           uuid.New(),
		Email:        dto.Email,
		PasswordHash: hash,
		Salt:         salt,
	}

	err = s.userRepo.Create(user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// Login — вход, возврат access и refresh токенов
func (s *AuthService) Login(dto dto.LoginDTO) (string, string, error) {
	user, err := s.userRepo.FindByEmail(dto.Email)
	if err != nil {
		return "", "", errors.New("invalid credentials")
	}

	if !crypto.CheckPassword(dto.Password, user.PasswordHash, user.Salt) {
		return "", "", errors.New("invalid credentials")
	}

	accessToken, err := crypto.GenerateAccessToken(user.ID, s.jwtAccessSecret, 15*time.Minute)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := crypto.GenerateRefreshToken(user.ID, s.jwtRefreshSecret, 7*24*time.Hour)
	if err != nil {
		return "", "", err
	}

	err = s.SaveRefreshToken(user.ID, refreshToken)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// Whoami — получить пользователя по ID
func (s *AuthService) Whoami(userID uuid.UUID) (*models.User, error) {
	return s.userRepo.FindByID(userID)
}

// Logout — отозвать refresh-токен (по ID пользователя, через middleware)
func (s *AuthService) Logout(userID uuid.UUID, refreshToken string) error {
	tokenRecord, err := s.findMatchingRefreshToken(userID, refreshToken)
	if err != nil {
		return err
	}

	return s.tokenRepo.RevokeByID(tokenRecord.ID)
}

// LogoutAll — отозвать все refresh-токены пользователя
func (s *AuthService) LogoutAll(userID uuid.UUID) error {
	return s.tokenRepo.RevokeAll(userID)
}

// CreateTokens генерирует access и refresh токены (для OAuth)
func (s *AuthService) CreateTokens(userID uuid.UUID) (string, string, error) {
	accessToken, err := crypto.GenerateAccessToken(userID, s.jwtAccessSecret, 15*time.Minute)
	if err != nil {
		return "", "", err
	}
	refreshToken, err := crypto.GenerateRefreshToken(userID, s.jwtRefreshSecret, 7*24*time.Hour)
	if err != nil {
		return "", "", err
	}
	return accessToken, refreshToken, nil
}

// SaveRefreshToken сохраняет хеш refresh-токена в БД
func (s *AuthService) SaveRefreshToken(userID uuid.UUID, refreshToken string) error {
	refreshHash, err := crypto.HashOpaqueToken(refreshToken)
	if err != nil {
		return err
	}

	tokenRecord := &models.Token{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: refreshHash,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		Revoked:   false,
	}
	return s.tokenRepo.Create(tokenRecord)
}

// RefreshTokens — обновление пары токенов
func (s *AuthService) RefreshTokens(refreshToken string) (string, string, error) {
	claims, err := crypto.ParseToken(refreshToken, s.jwtRefreshSecret)
	if err != nil {
		return "", "", errors.New("invalid refresh token")
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return "", "", errors.New("invalid user id in token")
	}

	tokenRecord, err := s.findMatchingRefreshToken(userID, refreshToken)
	if err != nil {
		return "", "", errors.New("invalid refresh token")
	}

	if err := s.tokenRepo.RevokeByID(tokenRecord.ID); err != nil {
		return "", "", err
	}

	newAccessToken, newRefreshToken, err := s.CreateTokens(userID)
	if err != nil {
		return "", "", err
	}

	if err := s.SaveRefreshToken(userID, newRefreshToken); err != nil {
		return "", "", err
	}

	return newAccessToken, newRefreshToken, nil
}

// ForgotPassword создаёт одноразовый токен сброса пароля.
func (s *AuthService) ForgotPassword(email string) (string, error) {
	resetToken, err := crypto.GenerateOpaqueToken()
	if err != nil {
		return "", err
	}

	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return resetToken, nil
	}

	resetHash, err := crypto.HashOpaqueToken(resetToken)
	if err != nil {
		return "", err
	}

	expiresAt := time.Now().Add(30 * time.Minute)
	user.ResetTokenHash = &resetHash
	user.ResetTokenExpiresAt = &expiresAt

	if err := s.userRepo.Update(user); err != nil {
		return "", err
	}

	return resetToken, nil
}

// ResetPassword меняет пароль по валидному reset-токену.
func (s *AuthService) ResetPassword(resetToken, newPassword string) error {
	users, err := s.userRepo.FindUsersWithActiveResetToken()
	if err != nil {
		return err
	}

	var user *models.User
	for i := range users {
		if users[i].ResetTokenHash != nil && crypto.CheckOpaqueToken(resetToken, *users[i].ResetTokenHash) {
			user = &users[i]
			break
		}
	}

	if user == nil {
		return errors.New("invalid reset token")
	}

	hash, salt, err := crypto.HashPassword(newPassword)
	if err != nil {
		return err
	}

	user.PasswordHash = hash
	user.Salt = salt
	user.ResetTokenHash = nil
	user.ResetTokenExpiresAt = nil

	if err := s.userRepo.Update(user); err != nil {
		return err
	}

	return s.tokenRepo.RevokeAll(user.ID)
}

func (s *AuthService) findMatchingRefreshToken(userID uuid.UUID, refreshToken string) (*models.Token, error) {
	tokens, err := s.tokenRepo.FindActiveByUser(userID)
	if err != nil {
		return nil, err
	}

	for i := range tokens {
		if crypto.CheckOpaqueToken(refreshToken, tokens[i].TokenHash) {
			return &tokens[i], nil
		}
	}

	return nil, errors.New("refresh token not found")
}
