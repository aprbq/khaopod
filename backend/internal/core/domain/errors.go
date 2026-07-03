package domain

import "errors"

// Sentinel error เชิงความหมาย — core คืนค่าพวกนี้ออกไป
// inbound adapter (HTTP) เป็นคนแปลงเป็น status/JSON เอง core ไม่รู้จัก HTTP
var (
	ErrNotFound     = errors.New("not found")
	ErrForbidden    = errors.New("forbidden")
	ErrUnauthorized = errors.New("unauthorized")
	ErrInactiveUser = errors.New("user is inactive")
	ErrInvalidInput = errors.New("invalid input")

	// Auth / OTP
	ErrInvalidOTP      = errors.New("invalid otp")
	ErrOTPExpired      = errors.New("otp expired")
	ErrTooManyAttempts = errors.New("too many otp attempts")
	ErrInvalidToken    = errors.New("invalid or expired token")
	ErrGoogleToken     = errors.New("invalid google id token")
)
