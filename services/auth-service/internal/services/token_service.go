package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/satishpolireddy/car-rental-ai/auth-service/config"
	"github.com/satishpolireddy/car-rental-ai/auth-service/internal/models"
	"github.com/satishpolireddy/car-rental-ai/auth-service/internal/repository"
)

const (
	accessTokenTTL  = 15 * time.Minute
	refreshTokenTTL = 7 * 24 * time.Hour
)

// JWTClaims embeds standard claims and adds user-specific fields.
type JWTClaims struct {
	UserID    uint   `json:"user_id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	jwt.RegisteredClaims
}

// TokenService issues and validates JWT access tokens and opaque refresh tokens.
type TokenService struct {
	cfg         config.JWTConfig
	refreshRepo *repository.RefreshTokenRepository
}

func NewTokenService(cfg config.JWTConfig, repo *repository.RefreshTokenRepository) *TokenService {
	return &TokenService{cfg: cfg, refreshRepo: repo}
}

// IssueTokenPair creates a short-lived JWT access token + long-lived opaque refresh token.
func (s *TokenService) IssueTokenPair(ctx context.Context, user *models.User) (*models.TokenPair, error) {
	// Access token (JWT)
	claims := JWTClaims{
		UserID:    user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "car-rental-auth",
			Subject:   fmt.Sprintf("%d", user.ID),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString([]byte(s.cfg.Secret))
	if err != nil {
		return nil, fmt.Errorf("sign access token: %w", err)
	}

	// Refresh token (opaque random bytes, stored as hash)
	rawRefresh, err := generateSecureToken(32)
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}
	tokenHash := sha256Hex(rawRefresh)
	expiresAt := time.Now().Add(refreshTokenTTL)

	if err := s.refreshRepo.Store(ctx, &models.RefreshToken{
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	}); err != nil {
		return nil, fmt.Errorf("store refresh token: %w", err)
	}

	return &models.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
		ExpiresIn:    int(accessTokenTTL.Seconds()),
		TokenType:    "Bearer",
	}, nil
}

// ValidateAccessToken parses and validates a JWT, returning the claims.
func (s *TokenService) ValidateAccessToken(tokenStr string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(s.cfg.Secret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}
	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}
	return claims, nil
}

// RefreshTokens validates a refresh token and issues a new pair (rotation).
func (s *TokenService) RefreshTokens(ctx context.Context, rawRefresh string) (*models.User, *models.TokenPair, error) {
	hash := sha256Hex(rawRefresh)

	record, err := s.refreshRepo.GetByHash(ctx, hash)
	if err != nil {
		return nil, nil, fmt.Errorf("refresh token not found or expired")
	}
	if record.Revoked || time.Now().After(record.ExpiresAt) {
		return nil, nil, fmt.Errorf("refresh token is expired or revoked")
	}

	// Revoke old token (rotation — one-use refresh tokens)
	if err := s.refreshRepo.Revoke(ctx, record.ID); err != nil {
		return nil, nil, fmt.Errorf("revoke old refresh token: %w", err)
	}

	user := &models.User{ID: record.UserID}
	pair, err := s.IssueTokenPair(ctx, user)
	return user, pair, err
}

// RevokeRefreshToken invalidates a specific refresh token (logout).
func (s *TokenService) RevokeRefreshToken(ctx context.Context, rawRefresh string) error {
	hash := sha256Hex(rawRefresh)
	return s.refreshRepo.RevokeByHash(ctx, hash)
}

func generateSecureToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func sha256Hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}
