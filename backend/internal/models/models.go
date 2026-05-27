package models

import (
	"time"
)

// Car represents a vehicle available for rental.
type Car struct {
	ID           uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Make         string    `json:"make" gorm:"size:50;not null;index"`
	Model        string    `json:"model" gorm:"size:50;not null;index"`
	Year         int       `json:"year" gorm:"not null"`
	Category     string    `json:"category" gorm:"size:30;not null;index"` // economy, standard, luxury, suv, van
	DailyRate    float64   `json:"daily_rate" gorm:"type:decimal(10,2);not null"`
	Available    bool      `json:"available" gorm:"default:true;index"`
	Location     string    `json:"location" gorm:"size:100;not null;index"`
	Seats        int       `json:"seats" gorm:"default:5"`
	Transmission string    `json:"transmission" gorm:"size:20;default:'automatic'"` // automatic, manual
	FuelType     string    `json:"fuel_type" gorm:"size:20;default:'petrol'"`        // petrol, diesel, electric, hybrid
	Features     string    `json:"features" gorm:"type:nvarchar(max)"`               // JSON array stored as string
	ImageURL     string    `json:"image_url" gorm:"size:500"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Customer represents a registered user.
type Customer struct {
	ID            uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	FirstName     string    `json:"first_name" gorm:"size:50;not null"`
	LastName      string    `json:"last_name" gorm:"size:50;not null"`
	Email         string    `json:"email" gorm:"size:100;not null;uniqueIndex"`
	Phone         string    `json:"phone" gorm:"size:20"`
	LicenseNumber string    `json:"license_number" gorm:"size:50"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	Bookings      []Booking `json:"bookings,omitempty" gorm:"foreignKey:CustomerID"`
}

// Booking represents a car rental reservation.
type Booking struct {
	ID           uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	CustomerID   uint      `json:"customer_id" gorm:"not null;index"`
	CarID        uint      `json:"car_id" gorm:"not null;index"`
	PickupDate   time.Time `json:"pickup_date" gorm:"not null;index"`
	ReturnDate   time.Time `json:"return_date" gorm:"not null"`
	PickupLoc    string    `json:"pickup_location" gorm:"size:100;not null"`
	ReturnLoc    string    `json:"return_location" gorm:"size:100;not null"`
	TotalDays    int       `json:"total_days" gorm:"not null"`
	TotalAmount  float64   `json:"total_amount" gorm:"type:decimal(10,2);not null"`
	Status       string    `json:"status" gorm:"size:20;default:'pending';index"` // pending, confirmed, active, completed, cancelled
	Notes        string    `json:"notes" gorm:"type:nvarchar(max)"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Customer     *Customer `json:"customer,omitempty" gorm:"foreignKey:CustomerID"`
	Car          *Car      `json:"car,omitempty" gorm:"foreignKey:CarID"`
}

// AIRecommendation stores AI-generated car suggestions for analytics.
type AIRecommendation struct {
	ID           uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	CustomerID   uint      `json:"customer_id" gorm:"index"`
	Query        string    `json:"query" gorm:"type:nvarchar(max)"`
	Response     string    `json:"response" gorm:"type:nvarchar(max)"`
	CarIDs       string    `json:"car_ids" gorm:"type:nvarchar(500)"` // JSON array of recommended car IDs
	ModelUsed    string    `json:"model_used" gorm:"size:50"`
	TokensUsed   int       `json:"tokens_used"`
	LatencyMs    int       `json:"latency_ms"`
	CreatedAt    time.Time `json:"created_at"`
}

// ETLLog tracks ETL pipeline runs for observability.
type ETLLog struct {
	ID           uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	PipelineName string    `json:"pipeline_name" gorm:"size:100;not null;index"`
	Status       string    `json:"status" gorm:"size:20;not null"` // running, success, failed
	RecordsIn    int       `json:"records_in"`
	RecordsOut   int       `json:"records_out"`
	ErrorMsg     string    `json:"error_msg" gorm:"type:nvarchar(max)"`
	StartedAt    time.Time `json:"started_at"`
	FinishedAt   *time.Time `json:"finished_at"`
}

// SearchRequest is the incoming search payload from the frontend.
type SearchRequest struct {
	Location   string    `json:"location" binding:"required"`
	PickupDate time.Time `json:"pickup_date" binding:"required"`
	ReturnDate time.Time `json:"return_date" binding:"required"`
	Category   string    `json:"category"`
	MaxRate    float64   `json:"max_daily_rate"`
	Seats      int       `json:"min_seats"`
}

// AIQueryRequest is the payload for the AI recommendation endpoint.
type AIQueryRequest struct {
	CustomerID uint   `json:"customer_id"`
	Query      string `json:"query" binding:"required"`
	Context    string `json:"context"` // optional extra context (e.g. trip type)
}

// CustomerStats is returned on the customer home page dashboard.
type CustomerStats struct {
	ActiveBookings    int64   `json:"active_bookings"`
	CompletedBookings int64   `json:"completed_bookings"`
	CancelledBookings int64   `json:"cancelled_bookings"`
	TotalSpent        float64 `json:"total_spent"`
}

// CustomerDashboard is the full payload returned for the home page.
type CustomerDashboard struct {
	Customer       *Customer       `json:"customer"`
	Stats          *CustomerStats  `json:"stats"`
	CurrentBookings []Booking      `json:"current_bookings"`
	PastBookings    []Booking      `json:"past_bookings"`
}

// BookingRequest is the payload to create a new booking.
type BookingRequest struct {
	CustomerID  uint      `json:"customer_id" binding:"required"`
	CarID       uint      `json:"car_id" binding:"required"`
	PickupDate  time.Time `json:"pickup_date" binding:"required"`
	ReturnDate  time.Time `json:"return_date" binding:"required"`
	PickupLoc   string    `json:"pickup_location" binding:"required"`
	ReturnLoc   string    `json:"return_location" binding:"required"`
	Notes       string    `json:"notes"`
}
