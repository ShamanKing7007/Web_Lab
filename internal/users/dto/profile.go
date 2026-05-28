package dto

// UpdateProfileDTO describes profile fields that can be changed by the current user.
// @Description Данные для обновления профиля пользователя.
type UpdateProfileDTO struct {
	DisplayName  *string `json:"display_name,omitempty" binding:"omitempty,max=100" example:"Ivan Petrov"`
	Bio          *string `json:"bio,omitempty" binding:"omitempty,max=500" example:"Backend developer"`
	AvatarFileID *string `json:"avatar_file_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
}
