package repository

import (
	"context"
	"fmt"

	"github.com/satishpolireddy/car-rental-ai/internal/models"
	"gorm.io/gorm"
)

type BookingRepository struct {
	db *gorm.DB
}

func NewBookingRepository(db *gorm.DB) *BookingRepository {
	return &BookingRepository{db: db}
}

func (r *BookingRepository) Create(ctx context.Context, booking *models.Booking) error {
	if err := r.db.WithContext(ctx).Create(booking).Error; err != nil {
		return fmt.Errorf("create booking: %w", err)
	}
	return nil
}

func (r *BookingRepository) GetByID(ctx context.Context, id uint) (*models.Booking, error) {
	var booking models.Booking
	if err := r.db.WithContext(ctx).
		Preload("Customer").
		Preload("Car").
		First(&booking, id).Error; err != nil {
		return nil, fmt.Errorf("get booking %d: %w", id, err)
	}
	return &booking, nil
}

func (r *BookingRepository) GetByCustomer(ctx context.Context, customerID uint) ([]models.Booking, error) {
	var bookings []models.Booking
	if err := r.db.WithContext(ctx).
		Preload("Car").
		Where("customer_id = ?", customerID).
		Order("pickup_date desc").
		Find(&bookings).Error; err != nil {
		return nil, fmt.Errorf("get bookings for customer %d: %w", customerID, err)
	}
	return bookings, nil
}

func (r *BookingRepository) UpdateStatus(ctx context.Context, id uint, status string) error {
	return r.db.WithContext(ctx).
		Model(&models.Booking{}).
		Where("id = ?", id).
		Update("status", status).Error
}

// GetCurrentByCustomer returns active/confirmed bookings (pickup_date >= now or currently active).
func (r *BookingRepository) GetCurrentByCustomer(ctx context.Context, customerID uint) ([]models.Booking, error) {
	var bookings []models.Booking
	if err := r.db.WithContext(ctx).
		Preload("Car").
		Where("customer_id = ?", customerID).
		Where("status IN ?", []string{"confirmed", "active"}).
		Order("pickup_date asc").
		Find(&bookings).Error; err != nil {
		return nil, fmt.Errorf("get current bookings for customer %d: %w", customerID, err)
	}
	return bookings, nil
}

// GetPastByCustomer returns completed or cancelled bookings.
func (r *BookingRepository) GetPastByCustomer(ctx context.Context, customerID uint) ([]models.Booking, error) {
	var bookings []models.Booking
	if err := r.db.WithContext(ctx).
		Preload("Car").
		Where("customer_id = ?", customerID).
		Where("status IN ?", []string{"completed", "cancelled"}).
		Order("return_date desc").
		Find(&bookings).Error; err != nil {
		return nil, fmt.Errorf("get past bookings for customer %d: %w", customerID, err)
	}
	return bookings, nil
}

// GetCustomerStats returns summary counts for the customer dashboard.
func (r *BookingRepository) GetCustomerStats(ctx context.Context, customerID uint) (*models.CustomerStats, error) {
	var stats models.CustomerStats
	r.db.WithContext(ctx).Model(&models.Booking{}).
		Where("customer_id = ? AND status IN ?", customerID, []string{"confirmed", "active"}).
		Count(&stats.ActiveBookings)
	r.db.WithContext(ctx).Model(&models.Booking{}).
		Where("customer_id = ? AND status = ?", customerID, "completed").
		Count(&stats.CompletedBookings)
	r.db.WithContext(ctx).Model(&models.Booking{}).
		Where("customer_id = ? AND status = ?", customerID, "cancelled").
		Count(&stats.CancelledBookings)
	r.db.WithContext(ctx).Model(&models.Booking{}).
		Where("customer_id = ? AND status = ?", customerID, "completed").
		Select("COALESCE(SUM(total_amount), 0)").Scan(&stats.TotalSpent)
	return &stats, nil
}

// IsCarAvailable checks for date-range conflicts using a single indexed query.
func (r *BookingRepository) IsCarAvailable(ctx context.Context, carID uint, req models.BookingRequest) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Booking{}).
		Where("car_id = ?", carID).
		Where("status IN ?", []string{"confirmed", "active"}).
		Where("pickup_date < ? AND return_date > ?", req.ReturnDate, req.PickupDate).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count == 0, nil
}
