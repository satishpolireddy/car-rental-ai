package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/satishpolireddy/car-rental-ai/auth-service/internal/models"
	"gorm.io/gorm"
)

type UserRepository struct{ db *gorm.DB }

func NewUserRepository(db *gorm.DB) *UserRepository { return &UserRepository{db: db} }

// GetOrCreate looks up a user by email; creates one if they don't exist.
// This is the registration + login merge pattern (OTP-only auth has no separate register step).
func (r *UserRepository) GetOrCreate(ctx context.Context, email, firstName, lastName string) (*models.User, bool, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err == nil {
		return &user, false, nil // existing user
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, fmt.Errorf("lookup user: %w", err)
	}
	// New user
	user = models.User{Email: email, FirstName: firstName, LastName: lastName, Verified: false}
	if err := r.db.WithContext(ctx).Create(&user).Error; err != nil {
		return nil, false, fmt.Errorf("create user: %w", err)
	}
	return &user, true, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uint) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).First(&user, id).Error; err != nil {
		return nil, fmt.Errorf("get user %d: %w", id, err)
	}
	return &user, nil
}

func (r *UserRepository) MarkVerified(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", id).Update("verified", true).Error
}
