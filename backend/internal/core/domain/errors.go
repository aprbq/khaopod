package domain

import "errors"

// Sentinel error เชิงความหมาย — core คืนค่าพวกนี้ออกไป
// inbound adapter (HTTP) เป็นคนแปลงเป็น status/JSON เอง core ไม่รู้จัก HTTP
var (
	ErrNotFound     = errors.New("not found")
	ErrForbidden    = errors.New("forbidden")
	ErrConflict     = errors.New("resource in use or duplicated")
	ErrUnauthorized = errors.New("unauthorized")
	ErrInactiveUser = errors.New("user is inactive")
	ErrInvalidInput = errors.New("invalid input")

	// Cart / stock
	ErrOutOfStock      = errors.New("out of stock")
	ErrInvalidQuantity = errors.New("invalid quantity")

	// Order / payment
	ErrCartEmpty           = errors.New("cart is empty")
	ErrOrderNotCancellable = errors.New("order cannot be cancelled")
	ErrPaymentNotAllowed   = errors.New("payment not allowed for this order")
	ErrAmountMismatch      = errors.New("payment amount does not match order total")

	// Auth / OTP
	ErrInvalidOTP      = errors.New("invalid otp")
	ErrOTPExpired      = errors.New("otp expired")
	ErrTooManyAttempts = errors.New("too many otp attempts")
	ErrInvalidToken    = errors.New("invalid or expired token")
	ErrGoogleToken     = errors.New("invalid google id token")
)
