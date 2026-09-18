package auth

import (
	"errors"
	"net/http"
	"strings"

	"github.com/ErenKarakus1/File-Management-Platform/internal/repository"
	"github.com/gin-gonic/gin"
)

const currentUserKey = "currentUser"

func Middleware(service *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}

		tokenString, ok := strings.CutPrefix(header, "Bearer ")
		if !ok || strings.TrimSpace(tokenString) == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header"})
			return
		}

		claims, err := service.ParseToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		user, err := service.UserByID(c.Request.Context(), claims.UserID)
		if err != nil {
			if errors.Is(err, repository.ErrUserNotFound) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
				return
			}
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "could not load user"})
			return
		}

		c.Set(currentUserKey, user)
		c.Next()
	}
}

func CurrentUser(c *gin.Context) (User, bool) {
	value, exists := c.Get(currentUserKey)
	if !exists {
		return User{}, false
	}

	user, ok := value.(User)
	return user, ok
}
