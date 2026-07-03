package rest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/khaopod/backend/internal/core/domain"
	"github.com/khaopod/backend/internal/core/port/input"
)

// fake input port — เทส handler โดยไม่แตะ service จริง
type fakeAuthUC struct {
	verifyErr error
	result    *input.AuthResult
}

func (f *fakeAuthUC) RequestOTP(context.Context, input.RequestOTPCommand) (input.OTPChallenge, error) {
	return input.OTPChallenge{Email: "a@b.com", ExpiresIn: 300}, nil
}
func (f *fakeAuthUC) LoginWithGoogle(context.Context, input.GoogleLoginCommand) (input.OTPChallenge, error) {
	return input.OTPChallenge{}, nil
}
func (f *fakeAuthUC) VerifyOTP(context.Context, input.VerifyOTPCommand) (*input.AuthResult, error) {
	return f.result, f.verifyErr
}
func (f *fakeAuthUC) Refresh(context.Context, input.RefreshCommand) (*input.AuthResult, error) {
	return f.result, nil
}
func (f *fakeAuthUC) Logout(context.Context, input.LogoutCommand) error { return nil }

type fakeUserUC struct{}

func (fakeUserUC) GetProfile(context.Context, uint) (*domain.User, error) { return nil, nil }
func (fakeUserUC) UpdateProfile(context.Context, uint, input.UpdateProfileCommand) (*domain.User, error) {
	return nil, nil
}

func setup(uc input.AuthUseCase) *gin.Engine {
	gin.SetMode(gin.TestMode)
	e := gin.New()
	h := NewAuthHandler(uc, fakeUserUC{}, 3600, false)
	e.POST("/v1/auth/otp/verify", h.VerifyOTP)
	return e
}

func doJSON(e *gin.Engine, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func TestVerifyOTP_ValidationError(t *testing.T) {
	e := setup(&fakeAuthUC{})
	// code ไม่ใช่ตัวเลข → 422 VALIDATION_ERROR
	rec := doJSON(e, "/v1/auth/otp/verify", `{"email":"a@b.com","code":"abc"}`)
	if rec.Code != 422 {
		t.Fatalf("want 422, got %d (%s)", rec.Code, rec.Body.String())
	}
}

func TestVerifyOTP_MapsInvalidOTP(t *testing.T) {
	e := setup(&fakeAuthUC{verifyErr: domain.ErrInvalidOTP})
	rec := doJSON(e, "/v1/auth/otp/verify", `{"email":"a@b.com","code":"123456"}`)
	if rec.Code != 400 {
		t.Fatalf("want 400, got %d", rec.Code)
	}
	var body struct {
		Success bool `json:"success"`
		Error   struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body.Success || body.Error.Code != "INVALID_OTP" {
		t.Fatalf("want INVALID_OTP envelope, got %s", rec.Body.String())
	}
}

func TestVerifyOTP_Success(t *testing.T) {
	e := setup(&fakeAuthUC{result: &input.AuthResult{
		AccessToken:  "acc",
		RefreshToken: "ref",
		ExpiresIn:    900,
		User:         &domain.User{PublicID: "pub", Email: "a@b.com", Role: domain.RoleCustomer},
	}})
	rec := doJSON(e, "/v1/auth/otp/verify", `{"email":"a@b.com","code":"123456"}`)
	if rec.Code != 200 {
		t.Fatalf("want 200, got %d (%s)", rec.Code, rec.Body.String())
	}
	// ต้องตั้ง httpOnly refresh cookie ให้
	if !strings.Contains(rec.Header().Get("Set-Cookie"), refreshCookieName) {
		t.Fatalf("want refresh cookie set, got %q", rec.Header().Get("Set-Cookie"))
	}
}
