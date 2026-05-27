package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/satishpolireddy/car-rental-ai/auth-service/internal/models"
	"github.com/satishpolireddy/car-rental-ai/auth-service/internal/repository"
	"github.com/satishpolireddy/car-rental-ai/auth-service/internal/services"
	log "github.com/sirupsen/logrus"
)

type AuthHandler struct {
	otpService   *services.OTPService
	tokenService *services.TokenService
	userRepo     *repository.UserRepository
}

func NewAuthHandler(otp *services.OTPService, token *services.TokenService, users *repository.UserRepository) *AuthHandler {
	return &AuthHandler{otpService: otp, tokenService: token, userRepo: users}
}

// SendOTP godoc
// POST /auth/send-otp
// Sends a 6-digit OTP to the provided email. Creates user account if first time.
func (h *AuthHandler) SendOTP(c *gin.Context) {
	var req models.SendOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get or create user (no separate registration flow)
	_, isNew, err := h.userRepo.GetOrCreate(c.Request.Context(), req.Email, req.FirstName, req.LastName)
	if err != nil {
		log.WithError(err).Error("GetOrCreate user failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process request"})
		return
	}

	if _, err := h.otpService.Send(c.Request.Context(), req.Email); err != nil {
		log.WithError(err).Error("OTP send failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send OTP"})
		return
	}

	msg := "OTP sent to your email"
	if isNew {
		msg = "Account created — OTP sent to your email"
	}
	// NEVER include the OTP in the response
	c.JSON(http.StatusOK, gin.H{"message": msg, "email": req.Email})
}

// VerifyOTP godoc
// POST /auth/verify-otp
// Verifies OTP and returns JWT access + refresh token pair.
func (h *AuthHandler) VerifyOTP(c *gin.Context) {
	var req models.VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.otpService.Verify(c.Request.Context(), req.Email, req.OTP); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Get the user record
	user, _, err := h.userRepo.GetOrCreate(c.Request.Context(), req.Email, "", "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load user"})
		return
	}

	// Mark verified on first successful OTP
	if !user.Verified {
		_ = h.userRepo.MarkVerified(c.Request.Context(), user.ID)
		user.Verified = true
	}

	// Issue token pair
	pair, err := h.tokenService.IssueTokenPair(c.Request.Context(), user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to issue tokens"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tokens": pair,
		"user": models.UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Phone:     user.Phone,
		},
	})
}

// Refresh godoc
// POST /auth/refresh
// Exchanges a valid refresh token for a new access + refresh token pair.
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req models.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, pair, err := h.tokenService.RefreshTokens(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Optionally enrich user object
	fullUser, _ := h.userRepo.GetByID(c.Request.Context(), user.ID)

	c.JSON(http.StatusOK, gin.H{"tokens": pair, "user": fullUser})
}

// Logout godoc
// POST /auth/logout
// Revokes the provided refresh token.
func (h *AuthHandler) Logout(c *gin.Context) {
	var req models.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_ = h.tokenService.RevokeRefreshToken(c.Request.Context(), req.RefreshToken)
	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

// Me godoc
// GET /auth/me — returns profile of authenticated user (JWT required)
func (h *AuthHandler) Me(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}
	user, err := h.userRepo.GetByID(c.Request.Context(), userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, models.UserResponse{
		ID: user.ID, Email: user.Email,
		FirstName: user.FirstName, LastName: user.LastName, Phone: user.Phone,
	})
}

// ValidateToken godoc
// POST /auth/validate — internal endpoint called by API Gateway to validate JWTs
func (h *AuthHandler) ValidateToken(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if len(authHeader) < 8 {
		c.JSON(http.StatusUnauthorized, gin.H{"valid": false})
		return
	}
	tokenStr := authHeader[7:] // strip "Bearer "
	claims, err := h.tokenService.ValidateAccessToken(tokenStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"valid": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"valid":      true,
		"user_id":    claims.UserID,
		"email":      claims.Email,
		"first_name": claims.FirstName,
	})
}
