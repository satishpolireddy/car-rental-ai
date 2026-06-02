package services

import (
	"context"
	"fmt"

	"github.com/satishpolireddy/car-rental-ai/payment-service/internal/models"
	"github.com/satishpolireddy/car-rental-ai/payment-service/internal/repository"
	log "github.com/sirupsen/logrus"
	"github.com/stripe/stripe-go/v78"
	"github.com/stripe/stripe-go/v78/paymentintent"
	"github.com/stripe/stripe-go/v78/refund"
)

type PaymentService struct {
	repo      *repository.PaymentRepository
	stripeKey string
}

func NewPaymentService(repo *repository.PaymentRepository, stripeKey string) *PaymentService {
	stripe.Key = stripeKey
	return &PaymentService{repo: repo, stripeKey: stripeKey}
}

func (s *PaymentService) CreatePaymentIntent(ctx context.Context, req models.CreatePaymentRequest) (*models.PaymentResponse, error) {
	currency := req.Currency
	if currency == "" {
		currency = "usd"
	}

	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(req.AmountCents),
		Currency: stripe.String(currency),
		Metadata: map[string]string{
			"booking_id":  fmt.Sprintf("%d", req.BookingID),
			"customer_id": fmt.Sprintf("%d", req.CustomerID),
		},
		AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
			Enabled: stripe.Bool(true),
		},
	}

	pi, err := paymentintent.New(params)
	if err != nil {
		return nil, fmt.Errorf("stripe create payment intent: %w", err)
	}

	payment := &models.Payment{
		BookingID:           req.BookingID,
		CustomerID:          req.CustomerID,
		AmountCents:         req.AmountCents,
		Currency:            currency,
		Status:              "pending",
		StripePaymentIntent: pi.ID,
		StripeClientSecret:  pi.ClientSecret,
	}
	if err := s.repo.Create(ctx, payment); err != nil {
		return nil, err
	}

	log.WithFields(log.Fields{
		"payment_id":        payment.ID,
		"payment_intent_id": pi.ID,
		"amount_cents":      req.AmountCents,
	}).Info("payment intent created")

	return toResponse(payment, true), nil
}

func (s *PaymentService) HandleWebhook(ctx context.Context, event stripe.Event) error {
	switch event.Type {
	case "payment_intent.succeeded":
		piID, _ := event.Data.Object["id"].(string)
		chargeID := ""
		if c, ok := event.Data.Object["latest_charge"].(string); ok {
			chargeID = c
		}
		return s.repo.UpdateByIntentID(ctx, piID, "succeeded", chargeID, "")

	case "payment_intent.payment_failed":
		piID, _ := event.Data.Object["id"].(string)
		reason := ""
		if lpe, ok := event.Data.Object["last_payment_error"].(map[string]interface{}); ok {
			reason, _ = lpe["message"].(string)
		}
		return s.repo.UpdateByIntentID(ctx, piID, "failed", "", reason)

	case "charge.refunded":
		chargeID, _ := event.Data.Object["id"].(string)
		refundID := ""
		if refunds, ok := event.Data.Object["refunds"].(map[string]interface{}); ok {
			if data, ok := refunds["data"].([]interface{}); ok && len(data) > 0 {
				if r, ok := data[0].(map[string]interface{}); ok {
					refundID, _ = r["id"].(string)
				}
			}
		}
		return s.repo.UpdateByChargeID(ctx, chargeID, "refunded", refundID)
	}
	return nil
}

func (s *PaymentService) Refund(ctx context.Context, req models.RefundRequest) (*models.Payment, error) {
	payment, err := s.repo.GetByID(ctx, req.PaymentID)
	if err != nil {
		return nil, err
	}
	if payment.Status != "succeeded" {
		return nil, fmt.Errorf("can only refund succeeded payments, current status: %s", payment.Status)
	}

	reason := stripe.RefundReasonRequestedByCustomer
	if req.Reason == "duplicate" {
		reason = stripe.RefundReasonDuplicate
	} else if req.Reason == "fraudulent" {
		reason = stripe.RefundReasonFraudulent
	}

	r, err := refund.New(&stripe.RefundParams{
		PaymentIntent: stripe.String(payment.StripePaymentIntent),
		Reason:        stripe.String(string(reason)),
	})
	if err != nil {
		return nil, fmt.Errorf("stripe refund: %w", err)
	}

	_ = s.repo.UpdateByIntentID(ctx, payment.StripePaymentIntent, "refunded", "", "")
	_ = s.repo.SetRefundID(ctx, payment.ID, r.ID)
	payment.Status = "refunded"
	payment.RefundID = r.ID

	log.WithFields(log.Fields{"payment_id": payment.ID, "refund_id": r.ID}).Info("payment refunded")
	return payment, nil
}

func (s *PaymentService) GetByBooking(ctx context.Context, bookingID uint) (*models.Payment, error) {
	return s.repo.GetByBookingID(ctx, bookingID)
}

func toResponse(p *models.Payment, includeSecret bool) *models.PaymentResponse {
	r := &models.PaymentResponse{
		ID: p.ID, BookingID: p.BookingID,
		AmountCents: p.AmountCents, Currency: p.Currency, Status: p.Status,
		PaymentIntentID: p.StripePaymentIntent,
		AmountDisplay:   fmt.Sprintf("$%.2f", float64(p.AmountCents)/100),
	}
	if includeSecret {
		r.ClientSecret = p.StripeClientSecret
	}
	return r
}
