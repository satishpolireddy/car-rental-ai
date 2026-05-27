package services

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/satishpolireddy/car-rental-ai/internal/models"
	"github.com/satishpolireddy/car-rental-ai/internal/repository"
)

// BookingService contains all business logic for booking operations.
// Kept separate from the repository layer to ensure clean separation of concerns.
type BookingService struct {
	bookingRepo *repository.BookingRepository
	carRepo     *repository.CarRepository
}

func NewBookingService(bookingRepo *repository.BookingRepository, carRepo *repository.CarRepository) *BookingService {
	return &BookingService{bookingRepo: bookingRepo, carRepo: carRepo}
}

// CreateBooking validates, calculates pricing, and persists a new booking atomically.
func (s *BookingService) CreateBooking(ctx context.Context, req models.BookingRequest) (*models.Booking, error) {
	// Validate dates
	if req.ReturnDate.Before(req.PickupDate) || req.ReturnDate.Equal(req.PickupDate) {
		return nil, fmt.Errorf("return date must be after pickup date")
	}
	if req.PickupDate.Before(time.Now()) {
		return nil, fmt.Errorf("pickup date cannot be in the past")
	}

	// Check car exists
	car, err := s.carRepo.GetByID(ctx, req.CarID)
	if err != nil {
		return nil, fmt.Errorf("car not found: %w", err)
	}

	// Check availability with date-range query
	available, err := s.bookingRepo.IsCarAvailable(ctx, req.CarID, req)
	if err != nil {
		return nil, fmt.Errorf("availability check: %w", err)
	}
	if !available {
		return nil, fmt.Errorf("car %d is not available for the selected dates", req.CarID)
	}

	// Calculate pricing
	days := int(math.Ceil(req.ReturnDate.Sub(req.PickupDate).Hours() / 24))
	total := car.DailyRate * float64(days)

	booking := &models.Booking{
		CustomerID:  req.CustomerID,
		CarID:       req.CarID,
		PickupDate:  req.PickupDate,
		ReturnDate:  req.ReturnDate,
		PickupLoc:   req.PickupLoc,
		ReturnLoc:   req.ReturnLoc,
		TotalDays:   days,
		TotalAmount: total,
		Status:      "confirmed",
		Notes:       req.Notes,
	}

	if err := s.bookingRepo.Create(ctx, booking); err != nil {
		return nil, err
	}

	// Mark car unavailable in real-time (also handled by availability query, belt-and-suspenders)
	_ = s.carRepo.SetAvailability(ctx, car.ID, false)

	return booking, nil
}

func (s *BookingService) GetBooking(ctx context.Context, id uint) (*models.Booking, error) {
	return s.bookingRepo.GetByID(ctx, id)
}

func (s *BookingService) CancelBooking(ctx context.Context, id uint) error {
	booking, err := s.bookingRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if booking.Status == "completed" || booking.Status == "cancelled" {
		return fmt.Errorf("cannot cancel a %s booking", booking.Status)
	}
	if err := s.bookingRepo.UpdateStatus(ctx, id, "cancelled"); err != nil {
		return err
	}
	// Free the car back up
	return s.carRepo.SetAvailability(ctx, booking.CarID, true)
}

func (s *BookingService) GetCustomerBookings(ctx context.Context, customerID uint) ([]models.Booking, error) {
	return s.bookingRepo.GetByCustomer(ctx, customerID)
}
