package rest

import (
	"errors"

	"github.com/gin-gonic/gin"

	"github.com/khaopod/backend/internal/adapter/inbound/rest/response"
	"github.com/khaopod/backend/internal/core/domain"
)

// mapError = จุดเดียวที่รู้จักทั้ง domain error และ HTTP status/code
// core คืน sentinel error เชิงความหมาย ที่นี่แปลงเป็น response envelope
func mapError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		response.Error(c, 404, response.CodeNotFound, "ไม่พบข้อมูล")
	case errors.Is(err, domain.ErrForbidden):
		response.Error(c, 403, response.CodeForbidden, "ไม่มีสิทธิ์เข้าถึง")
	case errors.Is(err, domain.ErrUnauthorized):
		response.Error(c, 401, response.CodeUnauthorized, "กรุณาเข้าสู่ระบบ")
	case errors.Is(err, domain.ErrInactiveUser):
		response.Error(c, 403, response.CodeAccountInactive, "บัญชีถูกระงับการใช้งาน")
	case errors.Is(err, domain.ErrInvalidInput):
		response.Error(c, 422, response.CodeValidation, "ข้อมูลไม่ถูกต้อง")
	case errors.Is(err, domain.ErrInvalidOTP):
		response.Error(c, 400, response.CodeInvalidOTP, "รหัส OTP ไม่ถูกต้องหรือหมดอายุแล้ว")
	case errors.Is(err, domain.ErrOTPExpired):
		response.Error(c, 400, response.CodeOTPExpired, "รหัส OTP หมดอายุแล้ว กรุณาขอรหัสใหม่")
	case errors.Is(err, domain.ErrTooManyAttempts):
		response.Error(c, 429, response.CodeTooManyAttempts, "กรอกรหัสผิดหลายครั้งเกินไป กรุณาขอรหัสใหม่")
	case errors.Is(err, domain.ErrInvalidToken):
		response.Error(c, 401, response.CodeInvalidToken, "เซสชันไม่ถูกต้องหรือหมดอายุ")
	case errors.Is(err, domain.ErrGoogleToken):
		response.Error(c, 401, response.CodeGoogleAuth, "ยืนยันตัวตนด้วย Google ไม่สำเร็จ")
	default:
		response.Error(c, 500, response.CodeInternal, "เกิดข้อผิดพลาดภายในระบบ")
	}
}
