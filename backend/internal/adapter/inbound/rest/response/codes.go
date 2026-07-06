package response

// รวม error code (machine-readable) ไว้ที่เดียว กันสะกดหลุด
const (
	CodeBadRequest      = "BAD_REQUEST"
	CodeValidation      = "VALIDATION_ERROR"
	CodeUnauthorized    = "UNAUTHORIZED"
	CodeForbidden       = "FORBIDDEN"
	CodeNotFound        = "NOT_FOUND"
	CodeOutOfStock      = "OUT_OF_STOCK"
	CodeInvalidOTP      = "INVALID_OTP"
	CodeOTPExpired      = "OTP_EXPIRED"
	CodeTooManyAttempts = "TOO_MANY_ATTEMPTS"
	CodeRateLimited     = "RATE_LIMITED"
	CodeInvalidToken    = "INVALID_TOKEN"
	CodeGoogleAuth      = "GOOGLE_AUTH_FAILED"
	CodeAccountInactive = "ACCOUNT_INACTIVE"
	CodeInternal        = "INTERNAL"
)
