package middleware

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type validateResponse struct {
	UserID    uint   `json:"user_id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
}

// JWTAuth validates Bearer tokens against the auth service's /auth/validate endpoint.
// On success it injects X-User-ID, X-User-Email, X-User-FirstName headers into
// the proxied request so downstream services can trust them without re-validating.
func JWTAuth(authServiceURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid Authorization header"})
			return
		}

		req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodPost,
			authServiceURL+"/auth/validate", nil)
		if err != nil {
			log.WithError(err).Error("could not create validate request")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "auth service unavailable"})
			return
		}
		req.Header.Set("Authorization", header)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.WithError(err).Error("auth service request failed")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "auth service unavailable"})
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		body, _ := io.ReadAll(resp.Body)
		var claims validateResponse
		if err := json.Unmarshal(body, &claims); err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
			return
		}

		// Inject verified identity headers for downstream services
		c.Request.Header.Set("X-User-ID", fmt.Sprintf("%d", claims.UserID))
		c.Request.Header.Set("X-User-Email", claims.Email)
		c.Request.Header.Set("X-User-FirstName", claims.FirstName)
		c.Set("user_id", claims.UserID)
		c.Next()
	}
}
