package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/satishpolireddy/car-rental-ai/auth-service/internal/services"
)

type JWTMiddleware struct {
	tokenService *services.TokenService
}

func NewJWTMiddleware(ts *services.TokenService) *JWTMiddleware {
	return &JWTMiddleware{tokenService: ts}
}

func (m *JWTMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid Authorization header"})
			return
		}
		claims, err := m.tokenService.ValidateAccessToken(header[7:])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Next()
	}
}
