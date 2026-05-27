package models

import "time"

// User is the identity record — minimal, no password (OTP-only auth).
type User struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Email     string    `json:"email" gorm:"size:100;not null;uniqueIndex"`
	FirstName string    `json:"first_name" gorm:"size:50;not null"`
	LastName  string    `json:"last_name" gorm:"size:50;not null"`
	Phone     string    `json:"phone" gorm:"size:20"`
	Verified  bool      `json:"verified" gorm:"default:false"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// OTPRecord stores the hashed OTP and expiry in Redis (not SQL) for speed.
// This struct is for marshalling to/from Redis only.
type OTPRecord struct {
	HashedOTP string    `json:"hashed_otp"`
	Email     string    `json:"email"`
	ExpiresAt time.Time `json:"expires_at"`
	Attempts  int       `json:"attempts"` // lockout after 5 failed attempts
}

// RefreshToken stored in DB for revocation support.
type RefreshToken struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID    uint      `json:"user_id" gorm:"not null;index"`
	TokenHash string    `json:"token_hash" gorm:"size:64;not null;uniqueIndex"`
	ExpiresAt time.Time `json:"expires_at" gorm:"not null;index"`
	Revoked   bool      `json:"revoked" gorm:"default:false"`
	CreatedAt time.Time `json:"created_at"`
}

// --- Request / Response types ---

type SendOTPRequest struct {
	Email     string `json:"email" binding:"required,email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type VerifyOTPRequest struct {
	Email string `json:"email" binding:"required,email"`
	OTP   string `json:"otp" binding:"required,len=6"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"` // seconds
	TokenType    string `json:"token_type"`
}

type UserResponse struct {
	ID        uint   `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
}
