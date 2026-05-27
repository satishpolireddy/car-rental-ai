package repository

import (
	"context"
	"fmt"

	"github.com/satishpolireddy/car-rental-ai/internal/models"
	"gorm.io/gorm"
)

type CustomerRepository struct {
	db *gorm.DB
}

func NewCustomerRepository(db *gorm.DB) *CustomerRepository {
	return &CustomerRepository{db: db}
}

func (r *CustomerRepository) GetByID(ctx context.Context, id uint) (*models.Customer, error) {
	var customer models.Customer
	if err := r.db.WithContext(ctx).First(&customer, id).Error; err != nil {
		return nil, fmt.Errorf("get customer %d: %w", id, err)
	}
	return &customer, nil
}

func (r *CustomerRepository) Create(ctx context.Context, c *models.Customer) error {
	if err := r.db.WithContext(ctx).Create(c).Error; err != nil {
		return fmt.Errorf("create customer: %w", err)
	}
	return nil
}

func (r *CustomerRepository) GetByEmail(ctx context.Context, email string) (*models.Customer, error) {
	var customer models.Customer
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&customer).Error; err != nil {
		return nil, fmt.Errorf("get customer by email: %w", err)
	}
	return &customer, nil
}
