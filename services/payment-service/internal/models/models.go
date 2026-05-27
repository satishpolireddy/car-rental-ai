package models

import "time"

// Payment tracks every payment lifecycle event.
type Payment struct {
	ID                  uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	BookingID           uint      `json:"booking_id" gorm:"not null;index"`
	CustomerID          uint      `json:"customer_id" gorm:"not null;index"`
	AmountCents         int64     `json:"amount_cents" gorm:"not null"` // stored in cents to avoid float precision issues
	Currency            string    `json:"currency" gorm:"size:3;default:'usd'"`
	Status              string    `json:"status" gorm:"size:20;not null;index"` // pending, succeeded, failed, refunded
	StripePaymentIntent string    `json:"stripe_payment_intent" gorm:"size:100;uniqueIndex"`
	StripeClientSecret  string    `json:"stripe_client_secret" gorm:"size:200"` // returned to frontend for card confirmation
	StripeChargeID      string    `json:"stripe_charge_id" gorm:"size:100"`
	RefundID            string    `json:"refund_id" gorm:"size:100"`
	FailureReason       string    `json:"failure_reason" gorm:"size:500"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// --- Request / Response types ---

type CreatePaymentRequest struct {
	BookingID   uint   `json:"booking_id" binding:"required"`
	CustomerID  uint   `json:"customer_id" binding:"required"`
	AmountCents int64  `json:"amount_cents" binding:"required,min=50"` // Stripe minimum is 50 cents
	Currency    string `json:"currency"`
}

type ConfirmPaymentRequest struct {
	PaymentIntentID string `json:"payment_intent_id" binding:"required"`
}

type RefundRequest struct {
	PaymentID uint   `json:"payment_id" binding:"required"`
	Reason    string `json:"reason"` // duplicate, fraudulent, requested_by_customer
}

type PaymentResponse struct {
	ID                 uint    `json:"id"`
	BookingID          uint    `json:"booking_id"`
	AmountCents        int64   `json:"amount_cents"`
	AmountDisplay      string  `json:"amount_display"` // e.g. "$95.00"
	Currency           string  `json:"currency"`
	Status             string  `json:"status"`
	ClientSecret       string  `json:"client_secret,omitempty"` // only on creation
	PaymentIntentID    string  `json:"payment_intent_id"`
}
