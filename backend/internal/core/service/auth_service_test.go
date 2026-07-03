package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/khaopod/backend/internal/core/domain"
	"github.com/khaopod/backend/internal/core/port/input"
	"github.com/khaopod/backend/internal/core/port/output"
)

const testOTPSecret = "test-secret"

func testAuthConfig() AuthConfig {
	return AuthConfig{
		OTPSecret:      testOTPSecret,
		OTPTTL:         5 * time.Minute,
		OTPLength:      6,
		OTPMaxAttempts: 3,
		RefreshTTL:     24 * time.Hour,
	}
}

// fixedClock ทำให้เวลา deterministic ในเทส
func (h *harness) fixClock(t time.Time) { h.svc.now = func() time.Time { return t } }

// fixOTP บังคับให้ OTP ที่สร้างเป็นค่าที่รู้ล่วงหน้า
func (h *harness) fixOTP(code string) { h.svc.newOTP = func() string { return code } }

func TestRequestOTP_SendsAndStoresHashOnly(t *testing.T) {
	h := newHarness()
	h.fixOTP("482913")

	res, err := h.svc.RequestOTP(context.Background(), input.RequestOTPCommand{Email: "Somchai@Gmail.com ", IP: "1.1.1.1"})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if res.Email != "somchai@gmail.com" {
		t.Fatalf("email should be normalized, got %q", res.Email)
	}
	if res.ExpiresIn != 300 {
		t.Fatalf("want ExpiresIn 300, got %d", res.ExpiresIn)
	}
	if h.mailer.sentCode != "482913" {
		t.Fatalf("mailer should receive the raw code, got %q", h.mailer.sentCode)
	}
	if len(h.otps.items) != 1 {
		t.Fatalf("want 1 otp stored, got %d", len(h.otps.items))
	}
	// ต้องเก็บ "แฮช" ไม่ใช่เลขจริง
	if got := h.otps.items[0].CodeHash; got == "482913" || got == "" {
		t.Fatalf("otp must be stored as hash, got %q", got)
	}
	if h.otps.invalidated != 1 {
		t.Fatalf("should invalidate previous active otps")
	}
}

func TestVerifyOTP_Success(t *testing.T) {
	h := newHarness()
	now := time.Date(2026, 7, 3, 10, 0, 0, 0, time.UTC)
	h.fixClock(now)
	h.fixOTP("111111")

	if _, err := h.svc.RequestOTP(context.Background(), input.RequestOTPCommand{Email: "a@b.com"}); err != nil {
		t.Fatalf("request: %v", err)
	}

	res, err := h.svc.VerifyOTP(context.Background(), input.VerifyOTPCommand{Email: "a@b.com", Code: "111111"})
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if res.AccessToken == "" || res.RefreshToken == "" {
		t.Fatalf("want tokens issued")
	}
	if res.User == nil || !res.User.EmailVerified {
		t.Fatalf("user should be created and email verified")
	}
	if res.User.LastLoginAt == nil {
		t.Fatalf("last_login_at should be set")
	}
	if h.tokens.issued != 1 {
		t.Fatalf("access token should be issued once")
	}
	// refresh token ต้องถูกเก็บแบบแฮช
	if _, ok := h.sessions.byHash[hashToken(res.RefreshToken)]; !ok {
		t.Fatalf("session should store hashed refresh token")
	}
}

func TestVerifyOTP_WrongCodeIncrementsAttempts(t *testing.T) {
	h := newHarness()
	h.fixOTP("111111")
	if _, err := h.svc.RequestOTP(context.Background(), input.RequestOTPCommand{Email: "a@b.com"}); err != nil {
		t.Fatalf("request: %v", err)
	}

	_, err := h.svc.VerifyOTP(context.Background(), input.VerifyOTPCommand{Email: "a@b.com", Code: "000000"})
	if !errors.Is(err, domain.ErrInvalidOTP) {
		t.Fatalf("want ErrInvalidOTP, got %v", err)
	}
	if h.otps.items[0].Attempts != 1 {
		t.Fatalf("attempts should be persisted as 1, got %d", h.otps.items[0].Attempts)
	}
}

func TestVerifyOTP_Expired(t *testing.T) {
	h := newHarness()
	now := time.Date(2026, 7, 3, 10, 0, 0, 0, time.UTC)
	h.fixClock(now)
	h.fixOTP("111111")
	if _, err := h.svc.RequestOTP(context.Background(), input.RequestOTPCommand{Email: "a@b.com"}); err != nil {
		t.Fatalf("request: %v", err)
	}

	// ขยับเวลาไปเกิน TTL
	h.fixClock(now.Add(6 * time.Minute))
	_, err := h.svc.VerifyOTP(context.Background(), input.VerifyOTPCommand{Email: "a@b.com", Code: "111111"})
	if !errors.Is(err, domain.ErrOTPExpired) {
		t.Fatalf("want ErrOTPExpired, got %v", err)
	}
}

func TestVerifyOTP_TooManyAttempts(t *testing.T) {
	h := newHarness()
	h.fixOTP("111111")
	if _, err := h.svc.RequestOTP(context.Background(), input.RequestOTPCommand{Email: "a@b.com"}); err != nil {
		t.Fatalf("request: %v", err)
	}

	// MaxAttempts = 3 → กรอกผิดจนล็อก
	for i := 0; i < 3; i++ {
		_, _ = h.svc.VerifyOTP(context.Background(), input.VerifyOTPCommand{Email: "a@b.com", Code: "000000"})
	}
	// แม้กรอกถูกก็ต้องโดนล็อกแล้ว
	_, err := h.svc.VerifyOTP(context.Background(), input.VerifyOTPCommand{Email: "a@b.com", Code: "111111"})
	if !errors.Is(err, domain.ErrTooManyAttempts) {
		t.Fatalf("want ErrTooManyAttempts, got %v", err)
	}
}

func TestVerifyOTP_NoActiveCode(t *testing.T) {
	h := newHarness()
	_, err := h.svc.VerifyOTP(context.Background(), input.VerifyOTPCommand{Email: "nobody@b.com", Code: "111111"})
	if !errors.Is(err, domain.ErrInvalidOTP) {
		t.Fatalf("want ErrInvalidOTP for no active code, got %v", err)
	}
}

func TestVerifyOTP_ConsumedCannotReuse(t *testing.T) {
	h := newHarness()
	h.fixOTP("111111")
	if _, err := h.svc.RequestOTP(context.Background(), input.RequestOTPCommand{Email: "a@b.com"}); err != nil {
		t.Fatalf("request: %v", err)
	}
	if _, err := h.svc.VerifyOTP(context.Background(), input.VerifyOTPCommand{Email: "a@b.com", Code: "111111"}); err != nil {
		t.Fatalf("first verify: %v", err)
	}
	// ใช้ซ้ำต้องไม่ได้ (consumed แล้ว → ไม่มี active → invalid)
	_, err := h.svc.VerifyOTP(context.Background(), input.VerifyOTPCommand{Email: "a@b.com", Code: "111111"})
	if !errors.Is(err, domain.ErrInvalidOTP) {
		t.Fatalf("want ErrInvalidOTP on reuse, got %v", err)
	}
}

func TestRefresh_RotatesToken(t *testing.T) {
	h := newHarness()
	h.fixOTP("111111")
	_, _ = h.svc.RequestOTP(context.Background(), input.RequestOTPCommand{Email: "a@b.com"})
	auth, err := h.svc.VerifyOTP(context.Background(), input.VerifyOTPCommand{Email: "a@b.com", Code: "111111"})
	if err != nil {
		t.Fatalf("verify: %v", err)
	}

	refreshed, err := h.svc.Refresh(context.Background(), input.RefreshCommand{RefreshToken: auth.RefreshToken})
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if refreshed.RefreshToken == auth.RefreshToken {
		t.Fatalf("refresh token should be rotated")
	}
	// token เก่าต้องใช้ไม่ได้อีก (revoked)
	if _, err := h.svc.Refresh(context.Background(), input.RefreshCommand{RefreshToken: auth.RefreshToken}); !errors.Is(err, domain.ErrInvalidToken) {
		t.Fatalf("old refresh token should be invalid, got %v", err)
	}
}

func TestLogout_RevokesSession(t *testing.T) {
	h := newHarness()
	h.fixOTP("111111")
	_, _ = h.svc.RequestOTP(context.Background(), input.RequestOTPCommand{Email: "a@b.com"})
	auth, _ := h.svc.VerifyOTP(context.Background(), input.VerifyOTPCommand{Email: "a@b.com", Code: "111111"})

	if err := h.svc.Logout(context.Background(), input.LogoutCommand{UserID: auth.User.ID, RefreshToken: auth.RefreshToken}); err != nil {
		t.Fatalf("logout: %v", err)
	}
	// หลัง logout refresh ต้องใช้ไม่ได้
	if _, err := h.svc.Refresh(context.Background(), input.RefreshCommand{RefreshToken: auth.RefreshToken}); !errors.Is(err, domain.ErrInvalidToken) {
		t.Fatalf("want ErrInvalidToken after logout, got %v", err)
	}
}

func TestLogout_ForbiddenForOtherUsersSession(t *testing.T) {
	h := newHarness()
	h.fixOTP("111111")
	_, _ = h.svc.RequestOTP(context.Background(), input.RequestOTPCommand{Email: "a@b.com"})
	auth, _ := h.svc.VerifyOTP(context.Background(), input.VerifyOTPCommand{Email: "a@b.com", Code: "111111"})

	// ผู้ใช้คนอื่น (id อื่น) พยายาม revoke session นี้ → ต้องโดน ErrForbidden (กัน IDOR)
	err := h.svc.Logout(context.Background(), input.LogoutCommand{UserID: auth.User.ID + 999, RefreshToken: auth.RefreshToken})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("want ErrForbidden, got %v", err)
	}
}

func TestLoginWithGoogle_CreatesUserAndSendsOTP(t *testing.T) {
	h := newHarness()
	h.google.identity = &output.GoogleIdentity{
		Subject:       "google-sub-1",
		Email:         "newuser@gmail.com",
		EmailVerified: true,
		Name:          "New User",
		Picture:       "https://avatar",
	}
	h.fixOTP("222222")

	res, err := h.svc.LoginWithGoogle(context.Background(), input.GoogleLoginCommand{IDToken: "tok"})
	if err != nil {
		t.Fatalf("google login: %v", err)
	}
	if res.DisplayName != "New User" {
		t.Fatalf("want display name from google, got %q", res.DisplayName)
	}
	if _, err := h.users.FindByEmail(context.Background(), "newuser@gmail.com"); err != nil {
		t.Fatalf("user should have been created")
	}
	if h.oauth.upserts != 1 {
		t.Fatalf("oauth link should be upserted")
	}
	if h.mailer.sentCode != "222222" {
		t.Fatalf("otp should still be emailed after google verify")
	}
}

func TestLoginWithGoogle_RejectsUnverifiedEmail(t *testing.T) {
	h := newHarness()
	h.google.identity = &output.GoogleIdentity{Subject: "s", Email: "x@gmail.com", EmailVerified: false}
	_, err := h.svc.LoginWithGoogle(context.Background(), input.GoogleLoginCommand{IDToken: "tok"})
	if !errors.Is(err, domain.ErrGoogleToken) {
		t.Fatalf("want ErrGoogleToken for unverified email, got %v", err)
	}
}
