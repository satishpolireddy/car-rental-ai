package repository

import (
	"context"
	"fmt"

	"github.com/satishpolireddy/car-rental-ai/internal/models"
	"gorm.io/gorm"
)

// CarRepository handles all database operations for cars.
// Uses GORM with SQL Server; all queries are parameterised to prevent SQL injection.
type CarRepository struct {
	db *gorm.DB
}

func NewCarRepository(db *gorm.DB) *CarRepository {
	return &CarRepository{db: db}
}

// Search returns available cars matching the criteria.
// Designed to be index-friendly: filters on (available, location, category).
func (r *CarRepository) Search(ctx context.Context, req models.SearchRequest) ([]models.Car, error) {
	var cars []models.Car

	q := r.db.WithContext(ctx).Where("available = ?", true)

	if req.Location != "" {
		q = q.Where("location = ?", req.Location)
	}
	if req.Category != "" {
		q = q.Where("category = ?", req.Category)
	}
	if req.MaxRate > 0 {
		q = q.Where("daily_rate <= ?", req.MaxRate)
	}
	if req.Seats > 0 {
		q = q.Where("seats >= ?", req.Seats)
	}

	// Exclude cars already booked for the requested dates
	subQuery := r.db.Model(&models.Booking{}).
		Select("car_id").
		Where("status IN ?", []string{"confirmed", "active"}).
		Where("pickup_date < ? AND return_date > ?", req.ReturnDate, req.PickupDate)

	q = q.Where("id NOT IN (?)", subQuery).Order("daily_rate asc")

	if err := q.Find(&cars).Error; err != nil {
		return nil, fmt.Errorf("car search query: %w", err)
	}
	return cars, nil
}

func (r *CarRepository) GetByID(ctx context.Context, id uint) (*models.Car, error) {
	var car models.Car
	if err := r.db.WithContext(ctx).First(&car, id).Error; err != nil {
		return nil, fmt.Errorf("get car %d: %w", id, err)
	}
	return &car, nil
}

func (r *CarRepository) GetByIDs(ctx context.Context, ids []uint) ([]models.Car, error) {
	var cars []models.Car
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&cars).Error; err != nil {
		return nil, fmt.Errorf("get cars by ids: %w", err)
	}
	return cars, nil
}

func (r *CarRepository) SetAvailability(ctx context.Context, carID uint, available bool) error {
	return r.db.WithContext(ctx).
		Model(&models.Car{}).
		Where("id = ?", carID).
		Update("available", available).Error
}

func (r *CarRepository) BulkUpsert(ctx context.Context, cars []models.Car) error {
	return r.db.WithContext(ctx).
		Save(&cars).Error
}

func (r *CarRepository) ListLocations(ctx context.Context) ([]string, error) {
	var locations []string
	if err := r.db.WithContext(ctx).
		Model(&models.Car{}).
		Distinct("location").
		Pluck("location", &locations).Error; err != nil {
		return nil, err
	}
	return locations, nil
}
