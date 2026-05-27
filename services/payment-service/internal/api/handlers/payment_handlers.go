package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/satishpolireddy/car-rental-ai/payment-service/internal/models"
	"github.com/satishpolireddy/car-rental-ai/payment-service/internal/services"
	log "github.com/sirupsen/logrus"
	"github.com/stripe/stripe-go/v78"
	"github.com/stripe/stripe-go/v78/webhook"
)

type PaymentHandler struct {
	svc           *services.PaymentService
	webhookSecret string
}

func NewPaymentHandler(svc *services.PaymentService, webhookSecret string) *PaymentHandler {
	return &PaymentHandler{svc: svc, webhookSecret: webhookSecret}
}

// POST /payments
func (h *PaymentHandler) CreatePayment(c *gin.Context) {
	var req models.CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.svc.CreatePaymentIntent(c.Request.Context(), req)
	if err != nil {
		log.WithError(err).Error("create payment intent failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create payment"})
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// POST /payments/refund
func (h *PaymentHandler) Refund(c *gin.Context) {
	var req models.RefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	payment, err := h.svc.Refund(c.Request.Context(), req)
	if err != nil {
		log.WithError(err).Error("refund failed")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, payment)
}

// GET /payments/booking/:booking_id
func (h *PaymentHandler) GetByBooking(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("booking_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid booking_id"})
		return
	}
	payment, err := h.svc.GetByBooking(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
		return
	}
	c.JSON(http.StatusOK, payment)
}

// POST /payments/webhook  — called by Stripe, NOT behind auth middleware
// Stripe signs every webhook; we verify the signature before processing.
func (h *PaymentHandler) Webhook(c *gin.Context) {
	const maxBodyBytes = int64(65536)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBodyBytes)

	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "read body failed"})
		return
	}

	sigHeader := c.GetHeader("Stripe-Signature")
	event, err := webhook.ConstructEvent(payload, sigHeader, h.webhookSecret)
	if err != nil {
		log.WithError(err).Warn("stripe webhook signature verification failed")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid signature"})
		return
	}

	// Log raw event for debugging
	raw, _ := json.Marshal(event)
	log.WithFields(log.Fields{"event_type": event.Type, "event_id": event.ID}).
		Debugf("stripe webhook: %s", string(raw))

	if err := h.svc.HandleWebhook(c.Request.Context(), event); err != nil {
		log.WithError(err).Error("handle webhook failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "webhook processing failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"received": true})
}

// GET /health
func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "payment-service",
		"stripe":  stripe.Key != "",
	})
}
