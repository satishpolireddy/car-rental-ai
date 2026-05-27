package services

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"net/smtp"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/satishpolireddy/car-rental-ai/auth-service/config"
	"github.com/satishpolireddy/car-rental-ai/auth-service/internal/models"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

const (
	otpTTL        = 10 * time.Minute
	otpKeyPrefix  = "otp:"
	maxAttempts   = 5
	otpLength     = 6
)

// OTPService handles generation, storage, delivery and verification of OTPs.
// OTPs are stored hashed in Redis (not plain text) — bcrypt cost 10.
type OTPService struct {
	redis  *redis.Client
	email  config.EmailConfig
}

func NewOTPService(rdb *redis.Client, emailCfg config.EmailConfig) *OTPService {
	return &OTPService{redis: rdb, email: emailCfg}
}

// Send generates a 6-digit OTP, hashes it, stores in Redis, and emails it.
func (s *OTPService) Send(ctx context.Context, email string) (string, error) {
	otp, err := generateOTP(otpLength)
	if err != nil {
		return "", fmt.Errorf("generate otp: %w", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(otp), 10)
	if err != nil {
		return "", fmt.Errorf("hash otp: %w", err)
	}

	record := models.OTPRecord{
		HashedOTP: string(hash),
		Email:     email,
		ExpiresAt: time.Now().Add(otpTTL),
		Attempts:  0,
	}

	data, _ := json.Marshal(record)
	if err := s.redis.Set(ctx, otpKey(email), data, otpTTL).Err(); err != nil {
		return "", fmt.Errorf("store otp in redis: %w", err)
	}

	if err := s.sendEmail(email, otp); err != nil {
		// Don't fail — log and continue so as not to leak whether email is registered
		log.WithError(err).WithField("email", email).Error("failed to send OTP email")
	}

	log.WithField("email", email).Info("OTP sent")
	return otp, nil // returned only for testing; never expose in API response
}

// Verify checks the OTP, enforces attempt limits, and cleans up on success.
func (s *OTPService) Verify(ctx context.Context, email, otp string) error {
	data, err := s.redis.Get(ctx, otpKey(email)).Bytes()
	if err == redis.Nil {
		return fmt.Errorf("OTP expired or not found — please request a new code")
	}
	if err != nil {
		return fmt.Errorf("redis get otp: %w", err)
	}

	var record models.OTPRecord
	if err := json.Unmarshal(data, &record); err != nil {
		return fmt.Errorf("parse otp record: %w", err)
	}

	if time.Now().After(record.ExpiresAt) {
		s.redis.Del(ctx, otpKey(email))
		return fmt.Errorf("OTP has expired — please request a new one")
	}

	if record.Attempts >= maxAttempts {
		s.redis.Del(ctx, otpKey(email))
		return fmt.Errorf("too many failed attempts — please request a new OTP")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(record.HashedOTP), []byte(otp)); err != nil {
		// Increment attempt counter
		record.Attempts++
		updated, _ := json.Marshal(record)
		ttl := time.Until(record.ExpiresAt)
		s.redis.Set(ctx, otpKey(email), updated, ttl)
		remaining := maxAttempts - record.Attempts
		return fmt.Errorf("invalid OTP — %d attempt(s) remaining", remaining)
	}

	// Verified — delete OTP so it can't be reused
	s.redis.Del(ctx, otpKey(email))
	return nil
}

func (s *OTPService) sendEmail(to, otp string) error {
	subject := "Your DriveAI Login Code"
	body := fmt.Sprintf(`
Hi there,

Your DriveAI one-time login code is:

    %s

This code expires in 10 minutes. Do not share it with anyone.

If you didn't request this, you can safely ignore this email.

— The DriveAI Team
`, otp)

	msg := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		s.email.From, to, subject, body,
	)

	auth := smtp.PlainAuth("", s.email.User, s.email.Password, s.email.Host)
	addr := fmt.Sprintf("%s:%s", s.email.Host, s.email.Port)
	return smtp.SendMail(addr, auth, s.email.From, []string{to}, []byte(msg))
}

func generateOTP(length int) (string, error) {
	digits := "0123456789"
	otp := make([]byte, length)
	for i := range otp {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", err
		}
		otp[i] = digits[n.Int64()]
	}
	return string(otp), nil
}

func otpKey(email string) string {
	return otpKeyPrefix + email
}
