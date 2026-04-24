package middleware

import "context"

type contextKey string

const UserIDKey contextKey = "user_id"

// SetUserID кладёт userID в контекст
func SetUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

// GetUserID извлекает userID из контекста
func GetUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}
