package repository

import (
	"context"
	"fmt"

	"github.com/satishpolireddy/car-rental-ai/auth-service/internal/models"
	"gorm.io/gorm"
)

type RefreshTokenRepository struct{ db *gorm.DB }

func NewRefreshTokenRepository(db *gorm.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

func (r *RefreshTokenRepository) Store(ctx context.Context, t *models.RefreshToken) error {
	if err := r.db.WithContext(ctx).Create(t).Error; err != nil {
		return fmt.Errorf("store refresh token: %w", err)
	}
	return nil
}

func (r *RefreshTokenRepository) GetByHash(ctx context.Context, hash string) (*models.RefreshToken, error) {
	var t models.RefreshToken
	if err := r.db.WithContext(ctx).Where("token_hash = ? AND revoked = false", hash).First(&t).Error; err != nil {
		return nil, fmt.Errorf("get refresh token: %w", err)
	}
	return &t, nil
}

func (r *RefreshTokenRepository) Revoke(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Model(&models.RefreshToken{}).Where("id = ?", id).Update("revoked", true).Error
}

func (r *RefreshTokenRepository) RevokeByHash(ctx context.Context, hash string) error {
	return r.db.WithContext(ctx).Model(&models.RefreshToken{}).Where("token_hash = ?", hash).Update("revoked", true).Error
}

func (r *RefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID uint) error {
	return r.db.WithContext(ctx).Model(&models.RefreshToken{}).Where("user_id = ?", userID).Update("revoked", true).Error
}
