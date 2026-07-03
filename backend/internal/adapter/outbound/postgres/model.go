package postgres

import (
	"time"

	"github.com/khaopod/backend/internal/core/domain"
)

// persistence model — gorm tag อยู่ที่นี่เท่านั้น (ไม่รั่วเข้า domain)
// nullable column ใช้ pointer/‌sql null แล้ว map ไปเป็นค่า zero ของ domain

type userRow struct {
	ID            uint       `gorm:"column:id;primaryKey"`
	PublicID      string     `gorm:"column:public_id"`
	Email         string     `gorm:"column:email"`
	EmailVerified bool       `gorm:"column:email_verified"`
	DisplayName   *string    `gorm:"column:display_name"`
	AvatarURL     *string    `gorm:"column:avatar_url"`
	Phone         *string    `gorm:"column:phone"`
	Role          string     `gorm:"column:role"`
	IsActive      bool       `gorm:"column:is_active"`
	LastLoginAt   *time.Time `gorm:"column:last_login_at"`
	CreatedAt     time.Time  `gorm:"column:created_at"`
	UpdatedAt     time.Time  `gorm:"column:updated_at"`
}

func (userRow) TableName() string { return "users" }

func toUserDomain(r userRow) domain.User {
	return domain.User{
		ID:            r.ID,
		PublicID:      r.PublicID,
		Email:         r.Email,
		EmailVerified: r.EmailVerified,
		DisplayName:   deref(r.DisplayName),
		AvatarURL:     deref(r.AvatarURL),
		Phone:         deref(r.Phone),
		Role:          domain.Role(r.Role),
		IsActive:      r.IsActive,
		LastLoginAt:   r.LastLoginAt,
		CreatedAt:     r.CreatedAt,
		UpdatedAt:     r.UpdatedAt,
	}
}

func toUserRow(u *domain.User) userRow {
	return userRow{
		ID:            u.ID,
		PublicID:      u.PublicID,
		Email:         u.Email,
		EmailVerified: u.EmailVerified,
		DisplayName:   nilIfEmpty(u.DisplayName),
		AvatarURL:     nilIfEmpty(u.AvatarURL),
		Phone:         nilIfEmpty(u.Phone),
		Role:          string(u.Role),
		IsActive:      u.IsActive,
		LastLoginAt:   u.LastLoginAt,
	}
}

type otpRow struct {
	ID          uint       `gorm:"column:id;primaryKey"`
	UserID      *uint      `gorm:"column:user_id"`
	Email       string     `gorm:"column:email"`
	Purpose     string     `gorm:"column:purpose"`
	CodeHash    string     `gorm:"column:code_hash"`
	ExpiresAt   time.Time  `gorm:"column:expires_at"`
	ConsumedAt  *time.Time `gorm:"column:consumed_at"`
	Attempts    int        `gorm:"column:attempts"`
	MaxAttempts int        `gorm:"column:max_attempts"`
	RequestIP   *string    `gorm:"column:request_ip"`
	CreatedAt   time.Time  `gorm:"column:created_at"`
}

func (otpRow) TableName() string { return "otp_codes" }

func toOTPDomain(r otpRow) domain.OTPCode {
	return domain.OTPCode{
		ID:          r.ID,
		UserID:      r.UserID,
		Email:       r.Email,
		Purpose:     domain.OTPPurpose(r.Purpose),
		CodeHash:    r.CodeHash,
		ExpiresAt:   r.ExpiresAt,
		ConsumedAt:  r.ConsumedAt,
		Attempts:    r.Attempts,
		MaxAttempts: r.MaxAttempts,
		RequestIP:   deref(r.RequestIP),
		CreatedAt:   r.CreatedAt,
	}
}

type sessionRow struct {
	ID               uint       `gorm:"column:id;primaryKey"`
	UserID           uint       `gorm:"column:user_id"`
	RefreshTokenHash string     `gorm:"column:refresh_token_hash"`
	UserAgent        *string    `gorm:"column:user_agent"`
	IPAddress        *string    `gorm:"column:ip_address"`
	ExpiresAt        time.Time  `gorm:"column:expires_at"`
	RevokedAt        *time.Time `gorm:"column:revoked_at"`
	CreatedAt        time.Time  `gorm:"column:created_at"`
}

func (sessionRow) TableName() string { return "auth_sessions" }

func toSessionDomain(r sessionRow) domain.Session {
	return domain.Session{
		ID:               r.ID,
		UserID:           r.UserID,
		RefreshTokenHash: r.RefreshTokenHash,
		UserAgent:        deref(r.UserAgent),
		IPAddress:        deref(r.IPAddress),
		ExpiresAt:        r.ExpiresAt,
		RevokedAt:        r.RevokedAt,
		CreatedAt:        r.CreatedAt,
	}
}

type oauthRow struct {
	ID             uint      `gorm:"column:id;primaryKey"`
	UserID         uint      `gorm:"column:user_id"`
	Provider       string    `gorm:"column:provider"`
	ProviderUserID string    `gorm:"column:provider_user_id"`
	ProviderEmail  *string   `gorm:"column:provider_email"`
	CreatedAt      time.Time `gorm:"column:created_at"`
}

func (oauthRow) TableName() string { return "user_oauth_accounts" }

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
