package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type AccessTokenValidator func(token string) (string, error)

// AuthMiddleware проверяет access token из cookie и загружает user_id в контекст.
func AuthMiddleware(validate AccessTokenValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Cookie("access_token")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}

		userID, err := validate(cookie)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		ctx := SetUserID(c.Request.Context(), userID)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
