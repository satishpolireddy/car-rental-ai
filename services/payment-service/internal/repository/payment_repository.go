package repository

import (
	"context"
	"fmt"

	"github.com/satishpolireddy/car-rental-ai/payment-service/internal/models"
	"gorm.io/gorm"
)

type PaymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

func (r *PaymentRepository) Create(ctx context.Context, p *models.Payment) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *PaymentRepository) GetByID(ctx context.Context, id uint) (*models.Payment, error) {
	var p models.Payment
	if err := r.db.WithContext(ctx).First(&p, id).Error; err != nil {
		return nil, fmt.Errorf("payment %d not found: %w", id, err)
	}
	return &p, nil
}

func (r *PaymentRepository) GetByBookingID(ctx context.Context, bookingID uint) (*models.Payment, error) {
	var p models.Payment
	if err := r.db.WithContext(ctx).Where("booking_id = ?", bookingID).First(&p).Error; err != nil {
		return nil, fmt.Errorf("payment for booking %d not found: %w", bookingID, err)
	}
	return &p, nil
}

// UpdateByIntentID is called by webhook handlers to authoritatively update status.
func (r *PaymentRepository) UpdateByIntentID(ctx context.Context, intentID, status, chargeID, failureReason string) error {
	updates := map[string]interface{}{"status": status}
	if chargeID != "" {
		updates["stripe_charge_id"] = chargeID
	}
	if failureReason != "" {
		updates["failure_reason"] = failureReason
	}
	return r.db.WithContext(ctx).
		Model(&models.Payment{}).
		Where("stripe_payment_intent = ?", intentID).
		Updates(updates).Error
}

// UpdateByChargeID is used for charge.refunded webhook events.
func (r *PaymentRepository) UpdateByChargeID(ctx context.Context, chargeID, status, refundID string) error {
	updates := map[string]interface{}{"status": status}
	if refundID != "" {
		updates["refund_id"] = refundID
	}
	return r.db.WithContext(ctx).
		Model(&models.Payment{}).
		Where("stripe_charge_id = ?", chargeID).
		Updates(updates).Error
}

// SetRefundID stores the Stripe refund ID after a manual refund.
func (r *PaymentRepository) SetRefundID(ctx context.Context, paymentID uint, refundID string) error {
	return r.db.WithContext(ctx).
		Model(&models.Payment{}).
		Where("id = ?", paymentID).
		Update("refund_id", refundID).Error
}
