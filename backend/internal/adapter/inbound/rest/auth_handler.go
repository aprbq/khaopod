package rest

import (
	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"

	"github.com/khaopod/backend/internal/adapter/inbound/rest/response"
	"github.com/khaopod/backend/internal/core/port/input"
)

const refreshCookieName = "refresh_token"

// AuthHandler แปลง HTTP ↔ AuthUseCase — ไม่มี business logic อยู่ที่นี่
type AuthHandler struct {
	auth         input.AuthUseCase
	users        input.UserUseCase
	refreshTTL   int  // วินาที (อายุ cookie)
	cookieSecure bool // true ใน production (ส่ง cookie เฉพาะ HTTPS)
}

func NewAuthHandler(auth input.AuthUseCase, users input.UserUseCase, refreshTTLSeconds int, cookieSecure bool) *AuthHandler {
	return &AuthHandler{auth: auth, users: users, refreshTTL: refreshTTLSeconds, cookieSecure: cookieSecure}
}

// POST /auth/otp/request
func (h *AuthHandler) RequestOTP(c *gin.Context) {
	var in requestOTPRequest
	if !bindAndValidate(c, &in) {
		return
	}
	ch, err := h.auth.RequestOTP(c.Request.Context(), input.RequestOTPCommand{
		Email: in.Email,
		IP:    c.ClientIP(),
	})
	if err != nil {
		mapError(c, err)
		return
	}
	response.OK(c, gin.H{
		"email":      ch.Email,
		"expires_in": ch.ExpiresIn,
		"message":    "ส่งรหัส OTP ไปที่อีเมลแล้ว",
	})
}

// POST /auth/google
func (h *AuthHandler) Google(c *gin.Context) {
	var in googleLoginRequest
	if !bindAndValidate(c, &in) {
		return
	}
	ch, err := h.auth.LoginWithGoogle(c.Request.Context(), input.GoogleLoginCommand{
		IDToken: in.IDToken,
		IP:      c.ClientIP(),
	})
	if err != nil {
		mapError(c, err)
		return
	}
	response.OK(c, gin.H{
		"email":        ch.Email,
		"display_name": ch.DisplayName,
		"expires_in":   ch.ExpiresIn,
		"message":      "ยืนยันตัวตนด้วย Google แล้ว ส่งรหัส OTP ไปที่อีเมล",
	})
}

// POST /auth/otp/verify
func (h *AuthHandler) VerifyOTP(c *gin.Context) {
	var in verifyOTPRequest
	if !bindAndValidate(c, &in) {
		return
	}
	res, err := h.auth.VerifyOTP(c.Request.Context(), input.VerifyOTPCommand{
		Email:     in.Email,
		Code:      in.Code,
		UserAgent: c.Request.UserAgent(),
		IP:        c.ClientIP(),
	})
	if err != nil {
		mapError(c, err)
		return
	}
	h.setRefreshCookie(c, res.RefreshToken)
	response.OK(c, toAuthResponse(res))
}

// POST /auth/refresh
func (h *AuthHandler) Refresh(c *gin.Context) {
	var in refreshRequest
	_ = c.ShouldBindJSON(&in) // body optional — token อาจมาจาก cookie
	token := h.readRefreshToken(c, in.RefreshToken)

	res, err := h.auth.Refresh(c.Request.Context(), input.RefreshCommand{
		RefreshToken: token,
		UserAgent:    c.Request.UserAgent(),
		IP:           c.ClientIP(),
	})
	if err != nil {
		mapError(c, err)
		return
	}
	h.setRefreshCookie(c, res.RefreshToken)
	response.OK(c, toAuthResponse(res))
}

// POST /auth/logout (🔒)
func (h *AuthHandler) Logout(c *gin.Context) {
	var in logoutRequest
	_ = c.ShouldBindJSON(&in)
	token := h.readRefreshToken(c, in.RefreshToken)

	err := h.auth.Logout(c.Request.Context(), input.LogoutCommand{
		UserID:       c.GetUint(ctxUserID),
		RefreshToken: token,
	})
	if err != nil {
		mapError(c, err)
		return
	}
	h.clearRefreshCookie(c)
	response.NoContent(c)
}

// GET /auth/me (🔒)
func (h *AuthHandler) Me(c *gin.Context) {
	u, err := h.users.GetProfile(c.Request.Context(), c.GetUint(ctxUserID))
	if err != nil {
		mapError(c, err)
		return
	}
	response.OK(c, toUserResponse(u))
}

func (h *AuthHandler) readRefreshToken(c *gin.Context, fromBody string) string {
	if fromBody != "" {
		return fromBody
	}
	if cookie, err := c.Cookie(refreshCookieName); err == nil {
		return cookie
	}
	return ""
}

func (h *AuthHandler) setRefreshCookie(c *gin.Context, token string) {
	// httpOnly กัน JS อ่าน token; Secure ใน production
	c.SetCookie(refreshCookieName, token, h.refreshTTL, "/", "", h.cookieSecure, true)
}

func (h *AuthHandler) clearRefreshCookie(c *gin.Context) {
	c.SetCookie(refreshCookieName, "", -1, "/", "", h.cookieSecure, true)
}

// bindAndValidate: bind JSON แล้ว validate ด้วย govalidator — คืน false ถ้าพลาด (ตอบ error ให้แล้ว)
func bindAndValidate(c *gin.Context, in any) bool {
	if err := c.ShouldBindJSON(in); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "รูปแบบข้อมูลไม่ถูกต้อง")
		return false
	}
	if ok, err := govalidator.ValidateStruct(in); !ok {
		response.Error(c, 422, response.CodeValidation, err.Error())
		return false
	}
	return true
}
