package rest

import (
	"github.com/khaopod/backend/internal/core/domain"
	"github.com/khaopod/backend/internal/core/port/input"
)

// ---- Request DTO (แยกจาก domain entity เสมอ กัน mass-assignment) ----

type requestOTPRequest struct {
	Email string `json:"email" valid:"required,email"`
}

type googleLoginRequest struct {
	IDToken string `json:"id_token" valid:"required"`
}

type verifyOTPRequest struct {
	Email string `json:"email" valid:"required,email"`
	Code  string `json:"code" valid:"required,numeric"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type updateProfileRequest struct {
	DisplayName *string `json:"display_name"`
	Phone       *string `json:"phone"`
}

func (r updateProfileRequest) toCommand() input.UpdateProfileCommand {
	return input.UpdateProfileCommand{DisplayName: r.DisplayName, Phone: r.Phone}
}

// ---- Response DTO ----

type userResponse struct {
	PublicID    string `json:"public_id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url"`
	Phone       string `json:"phone,omitempty"`
	Role        string `json:"role"`
}

func toUserResponse(u *domain.User) userResponse {
	return userResponse{
		PublicID:    u.PublicID,
		Email:       u.Email,
		DisplayName: u.DisplayName,
		AvatarURL:   u.AvatarURL,
		Phone:       u.Phone,
		Role:        string(u.Role),
	}
}

type authResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	TokenType    string       `json:"token_type"`
	ExpiresIn    int          `json:"expires_in"`
	User         userResponse `json:"user"`
}

func toAuthResponse(r *input.AuthResult) authResponse {
	return authResponse{
		AccessToken:  r.AccessToken,
		RefreshToken: r.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    r.ExpiresIn,
		User:         toUserResponse(r.User),
	}
}
