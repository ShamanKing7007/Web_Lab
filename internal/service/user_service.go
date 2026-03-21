package service

import (
	"Web_lab/internal/errors"
	"Web_lab/internal/models"
	"Web_lab/internal/repository"
	"Web_lab/internal/validator"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

// DTO для создания пользователя
type CreateUserDTO struct {
	Name     string `json:"name" binding:"required,min=2,max=100"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=4"`
}

// DTO для обновления пользователя
type UpdateUserDTO struct {
	Name     string `json:"name" binding:"omitempty,min=2,max=100"`
	Email    string `json:"email" binding:"omitempty,email"`
	Password string `json:"password" binding:"omitempty,min=4"`
}

// Мета для пагинации
type PaginationMeta struct {
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalPages int   `json:"total_pages"`
}

// Ответ со списком
type UsersResponse struct {
	Data []models.User  `json:"data"`
	Meta PaginationMeta `json:"meta"`
}

// Создание пользователя
func (s *UserService) Create(dto CreateUserDTO) (*models.User, error) {
	// Валидация данных
	if !validator.IsValidName(dto.Name) {
		return nil, errors.ErrValidation
	}
	if !validator.IsValidEmail(dto.Email) {
		return nil, errors.ErrValidation
	}
	if !validator.IsValidPassword(dto.Password) {
		return nil, errors.ErrValidation
	}

	// Хеширование пароля
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(dto.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		ID:       uuid.New(),
		Name:     dto.Name,
		Email:    dto.Email,
		Password: string(hashedPassword),
	}

	err = s.repo.Create(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// Поиск по ID
func (s *UserService) GetByID(id uuid.UUID) (*models.User, error) {
	user, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.ErrNotFound
	}
	return user, nil
}

// Получить всех с пагинацией
func (s *UserService) GetAll(page, limit int) (*UsersResponse, error) {
	// Валидация пагинации
	page, limit = validator.ValidatePagination(page, limit)

	offset := (page - 1) * limit

	users, total, err := s.repo.FindAll(offset, limit)
	if err != nil {
		return nil, err
	}

	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}

	return &UsersResponse{
		Data: users,
		Meta: PaginationMeta{
			Total:      total,
			Page:       page,
			Limit:      limit,
			TotalPages: totalPages,
		},
	}, nil
}

// Обновление (полное)
func (s *UserService) Update(id uuid.UUID, dto UpdateUserDTO) (*models.User, error) {
	// Валидация данных (если предоставлены)
	if dto.Name != "" && !validator.IsValidName(dto.Name) {
		return nil, errors.ErrValidation
	}
	if dto.Email != "" && !validator.IsValidEmail(dto.Email) {
		return nil, errors.ErrValidation
	}
	if dto.Password != "" && !validator.IsValidPassword(dto.Password) {
		return nil, errors.ErrValidation
	}

	user, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.ErrNotFound
	}

	user.Name = dto.Name
	user.Email = dto.Email

	// Обновление пароля, если предоставлен
	if dto.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(dto.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		user.Password = string(hashedPassword)
	}

	err = s.repo.Update(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// Частичное обновление
func (s *UserService) Patch(id uuid.UUID, dto UpdateUserDTO) (*models.User, error) {
	// Валидация данных (если предоставлены)
	if dto.Name != "" && !validator.IsValidName(dto.Name) {
		return nil, errors.ErrValidation
	}
	if dto.Email != "" && !validator.IsValidEmail(dto.Email) {
		return nil, errors.ErrValidation
	}
	if dto.Password != "" && !validator.IsValidPassword(dto.Password) {
		return nil, errors.ErrValidation
	}

	user, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.ErrNotFound
	}

	if dto.Name != "" {
		user.Name = dto.Name
	}
	if dto.Email != "" {
		user.Email = dto.Email
	}
	// Обновление пароля, если предоставлен
	if dto.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(dto.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		user.Password = string(hashedPassword)
	}

	err = s.repo.Update(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// Soft delete
func (s *UserService) Delete(id uuid.UUID) error {
	_, err := s.repo.FindByID(id)
	if err != nil {
		return errors.ErrNotFound
	}

	return s.repo.Delete(id)
}
