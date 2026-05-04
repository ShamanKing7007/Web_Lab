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

const (
	accessTokenType  = "access"
	refreshTokenType = "refresh"
)

type AuthService struct {
	userRepo         *repository.UserRepository
	tokenRepo        *repository.TokenRepository
	jwtAccessSecret  string
	jwtRefreshSecret string
	jwtAccessTTL     time.Duration
	jwtRefreshTTL    time.Duration
}

func NewAuthService(
	userRepo *repository.UserRepository,
	tokenRepo *repository.TokenRepository,
	jwtAccessSecret string,
	jwtRefreshSecret string,
	jwtAccessTTL time.Duration,
	jwtRefreshTTL time.Duration,
) *AuthService {
	return &AuthService{
		userRepo:         userRepo,
		tokenRepo:        tokenRepo,
		jwtAccessSecret:  jwtAccessSecret,
		jwtRefreshSecret: jwtRefreshSecret,
		jwtAccessTTL:     jwtAccessTTL,
		jwtRefreshTTL:    jwtRefreshTTL,
	}
}

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

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) Login(dto dto.LoginDTO) (string, string, error) {
	user, err := s.userRepo.FindByEmail(dto.Email)
	if err != nil {
		return "", "", errors.New("invalid credentials")
	}

	if !crypto.CheckPassword(dto.Password, user.PasswordHash, user.Salt) {
		return "", "", errors.New("invalid credentials")
	}

	accessToken, refreshToken, err := s.CreateTokens(user.ID)
	if err != nil {
		return "", "", err
	}

	if err := s.SaveTokenPair(user.ID, accessToken, refreshToken); err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *AuthService) Whoami(userID uuid.UUID) (*models.User, error) {
	return s.userRepo.FindByID(userID)
}

func (s *AuthService) AccessTokenTTL() time.Duration {
	return s.jwtAccessTTL
}

func (s *AuthService) RefreshTokenTTL() time.Duration {
	return s.jwtRefreshTTL
}

func (s *AuthService) ValidateAccessToken(accessToken string) (string, error) {
	claims, err := crypto.ParseToken(accessToken, s.jwtAccessSecret)
	if err != nil {
		return "", err
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return "", err
	}

	if _, err := s.findMatchingToken(userID, accessToken, accessTokenType); err != nil {
		return "", err
	}

	return userID.String(), nil
}

func (s *AuthService) Logout(userID uuid.UUID, accessToken, refreshToken string) error {
	if accessToken != "" {
		if tokenRecord, err := s.findMatchingToken(userID, accessToken, accessTokenType); err == nil {
			if err := s.tokenRepo.RevokeByID(tokenRecord.ID); err != nil {
				return err
			}
		}
	}

	if refreshToken != "" {
		if tokenRecord, err := s.findMatchingToken(userID, refreshToken, refreshTokenType); err == nil {
			if err := s.tokenRepo.RevokeByID(tokenRecord.ID); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *AuthService) LogoutAll(userID uuid.UUID) error {
	return s.tokenRepo.RevokeAll(userID)
}

func (s *AuthService) CreateTokens(userID uuid.UUID) (string, string, error) {
	accessToken, err := crypto.GenerateAccessToken(userID, s.jwtAccessSecret, s.jwtAccessTTL)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := crypto.GenerateRefreshToken(userID, s.jwtRefreshSecret, s.jwtRefreshTTL)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *AuthService) SaveTokenPair(userID uuid.UUID, accessToken, refreshToken string) error {
	if err := s.saveToken(userID, accessToken, accessTokenType, s.jwtAccessTTL); err != nil {
		return err
	}

	if err := s.saveToken(userID, refreshToken, refreshTokenType, s.jwtRefreshTTL); err != nil {
		return err
	}

	return nil
}

func (s *AuthService) RefreshTokens(refreshToken string) (string, string, error) {
	claims, err := crypto.ParseToken(refreshToken, s.jwtRefreshSecret)
	if err != nil {
		return "", "", errors.New("invalid refresh token")
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return "", "", errors.New("invalid user id in token")
	}

	tokenRecord, err := s.findMatchingToken(userID, refreshToken, refreshTokenType)
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

	if err := s.SaveTokenPair(userID, newAccessToken, newRefreshToken); err != nil {
		return "", "", err
	}

	return newAccessToken, newRefreshToken, nil
}

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

func (s *AuthService) saveToken(userID uuid.UUID, rawToken, tokenType string, ttl time.Duration) error {
	tokenHash, err := crypto.HashOpaqueToken(rawToken)
	if err != nil {
		return err
	}

	tokenRecord := &models.Token{
		ID:        uuid.New(),
		UserID:    userID,
		Type:      tokenType,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(ttl),
		Revoked:   false,
	}

	return s.tokenRepo.Create(tokenRecord)
}

func (s *AuthService) findMatchingToken(userID uuid.UUID, rawToken, tokenType string) (*models.Token, error) {
	tokens, err := s.tokenRepo.FindActiveByUser(userID, tokenType)
	if err != nil {
		return nil, err
	}

	for i := range tokens {
		if crypto.CheckOpaqueToken(rawToken, tokens[i].TokenHash) {
			return &tokens[i], nil
		}
	}

	return nil, errors.New("token not found")
}
