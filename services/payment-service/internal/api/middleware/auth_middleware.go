package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates Bearer tokens by forwarding to the auth service's /auth/validate endpoint.
// The API Gateway typically handles this, but the payment service also validates directly for defense-in-depth.
func AuthMiddleware(authServiceURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid Authorization header"})
			return
		}

		// Forward validation to auth service
		req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodPost, authServiceURL+"/auth/validate", nil)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "auth check failed"})
			return
		}
		req.Header.Set("Authorization", header)

		resp, err := http.DefaultClient.Do(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		defer resp.Body.Close()

		// Propagate user_id from auth service response header (set by API gateway)
		if userID := resp.Header.Get("X-User-ID"); userID != "" {
			c.Set("user_id", userID)
		}
		c.Next()
	}
}
