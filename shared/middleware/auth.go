package middleware

import (
	"ecommerce-app/shared/security"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const ctxUserID = "user_id"

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if h == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}

		parts := strings.SplitN(h, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header"})
			return
		}

		token := parts[1]
		claims, err := security.ParseToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token: " + err.Error()})
			return
		}

		c.Set(ctxUserID, claims.UserID)
		c.Next()
	}
}

func GetUserID(c *gin.Context) (string, bool) {
	v, ok := c.Get(ctxUserID)
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}
