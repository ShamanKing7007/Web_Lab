package service

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"Web_lab/internal/apperrors"
	"Web_lab/internal/cache"
	storageService "Web_lab/internal/storage/service"
	"Web_lab/internal/users/dto"
	"Web_lab/internal/users/models"
	"Web_lab/internal/users/repository"

	"github.com/google/uuid"
)

type ProfileService interface {
	Get(userID uuid.UUID) (*models.UserResponse, error)
	Update(userID uuid.UUID, req dto.UpdateProfileDTO) (*models.UserResponse, error)
}

type ProfileServiceImpl struct {
	userRepo    *repository.UserRepository
	fileService storageService.FileService
	cache       *cache.Service
	cacheTTL    time.Duration
}

func NewProfileService(
	userRepo *repository.UserRepository,
	fileService storageService.FileService,
	cacheService *cache.Service,
	cacheTTL time.Duration,
) ProfileService {
	return &ProfileServiceImpl{
		userRepo:    userRepo,
		fileService: fileService,
		cache:       cacheService,
		cacheTTL:    cacheTTL,
	}
}

func (s *ProfileServiceImpl) Get(userID uuid.UUID) (*models.UserResponse, error) {
	ctx := context.Background()
	key := userProfileKey(userID)

	var cached models.UserResponse
	if ok, err := s.cache.Get(ctx, key, &cached); err == nil && ok {
		return &cached, nil
	} else if err != nil {
		log.Printf("failed to read user profile cache %s: %v", key, err)
	}

	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, apperrors.ErrUserNotFound
	}

	response := user.ToResponse()
	if err := s.cache.Set(ctx, key, response, s.cacheTTL); err != nil {
		log.Printf("failed to write user profile cache %s: %v", key, err)
	}

	return &response, nil
}

func (s *ProfileServiceImpl) Update(userID uuid.UUID, req dto.UpdateProfileDTO) (*models.UserResponse, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, apperrors.ErrUserNotFound
	}

	oldAvatarID := user.AvatarFileID

	if req.DisplayName != nil {
		value := strings.TrimSpace(*req.DisplayName)
		user.DisplayName = &value
	}
	if req.Bio != nil {
		value := strings.TrimSpace(*req.Bio)
		user.Bio = &value
	}
	if req.AvatarFileID != nil {
		avatarID, err := uuid.Parse(strings.TrimSpace(*req.AvatarFileID))
		if err != nil {
			return nil, apperrors.ErrValidation
		}
		metadata, err := s.fileService.GetMetadata(avatarID, userID)
		if err != nil {
			return nil, err
		}
		if !isAvatarMime(metadata.MimeType) {
			return nil, apperrors.ErrValidation
		}
		user.AvatarFileID = &avatarID
	}

	if err := s.userRepo.UpdateProfile(user); err != nil {
		return nil, err
	}

	s.invalidateProfileCache(userID)
	if oldAvatarID != nil && user.AvatarFileID != nil && *oldAvatarID != *user.AvatarFileID {
		if err := s.fileService.Delete(*oldAvatarID, userID); err != nil && !errors.Is(err, apperrors.ErrFileNotFound) {
			log.Printf("failed to delete previous avatar %s: %v", oldAvatarID, err)
		}
	}

	response := user.ToResponse()
	if err := s.cache.Set(context.Background(), userProfileKey(userID), response, s.cacheTTL); err != nil {
		log.Printf("failed to write user profile cache %s: %v", userProfileKey(userID), err)
	}

	return &response, nil
}

func (s *ProfileServiceImpl) invalidateProfileCache(userID uuid.UUID) {
	key := userProfileKey(userID)
	if err := s.cache.Del(context.Background(), key); err != nil {
		log.Printf("failed to invalidate user profile cache %s: %v", key, err)
	}
}

func isAvatarMime(mimeType string) bool {
	return mimeType == "image/png" || mimeType == "image/jpeg" || mimeType == "image/jpg"
}
