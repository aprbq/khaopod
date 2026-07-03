package input

import (
	"context"

	"github.com/khaopod/backend/internal/core/domain"
)

// AuthUseCase = สิ่งที่ระบบ auth ทำได้ (driving port) — inbound HTTP adapter เรียกผ่านอันนี้
type AuthUseCase interface {
	// RequestOTP: ขอ OTP ด้วยอีเมลโดยตรง แล้วส่งไปที่อีเมล
	RequestOTP(ctx context.Context, cmd RequestOTPCommand) (OTPChallenge, error)
	// LoginWithGoogle: verify id_token กับ Google, สร้าง user ถ้ายังไม่มี, แล้วส่ง OTP ต่อ
	LoginWithGoogle(ctx context.Context, cmd GoogleLoginCommand) (OTPChallenge, error)
	// VerifyOTP: ยืนยัน OTP สำเร็จ → ออก access/refresh token
	VerifyOTP(ctx context.Context, cmd VerifyOTPCommand) (*AuthResult, error)
	// Refresh: ต่ออายุ access token + rotate refresh token
	Refresh(ctx context.Context, cmd RefreshCommand) (*AuthResult, error)
	// Logout: เพิกถอน session/refresh token ปัจจุบัน
	Logout(ctx context.Context, cmd LogoutCommand) error
}

type RequestOTPCommand struct {
	Email string
	IP    string
}

type GoogleLoginCommand struct {
	IDToken string
	IP      string
}

// OTPChallenge = ผลลัพธ์หลังส่ง OTP (ยังไม่ล็อกอิน) — ผู้ใช้ต้องเอา code มา verify ต่อ
type OTPChallenge struct {
	Email       string
	DisplayName string
	ExpiresIn   int // วินาที
}

type VerifyOTPCommand struct {
	Email     string
	Code      string
	UserAgent string
	IP        string
}

type RefreshCommand struct {
	RefreshToken string
	UserAgent    string
	IP           string
}

type LogoutCommand struct {
	UserID       uint
	RefreshToken string
}

// AuthResult = token set + ข้อมูล user หลังล็อกอินสำเร็จ
type AuthResult struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int // วินาที (อายุ access token)
	User         *domain.User
}
