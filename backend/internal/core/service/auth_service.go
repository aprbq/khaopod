package service

import (
	"context"
	"errors"
	"time"

	"github.com/khaopod/backend/internal/core/domain"
	"github.com/khaopod/backend/internal/core/port/input"
	"github.com/khaopod/backend/internal/core/port/output"
)

// AuthConfig — พารามิเตอร์ด้านความปลอดภัยของ auth (มาจาก config, inject ตอน wire)
type AuthConfig struct {
	OTPSecret      string
	OTPTTL         time.Duration
	OTPLength      int
	OTPMaxAttempts int
	RefreshTTL     time.Duration
}

// AuthService = use case ของ auth (implements input.AuthUseCase)
// พึ่งพาเฉพาะ output port — ทดสอบได้ด้วย fake ไม่ต้องมี DB/HTTP จริง
type AuthService struct {
	users    output.UserRepository
	otps     output.OTPRepository
	sessions output.SessionRepository
	oauth    output.OAuthRepository
	mailer   output.Mailer
	tokens   output.Tokenizer
	google   output.GoogleVerifier
	tx       output.TxManager
	cfg      AuthConfig

	// hook สำหรับ test (deterministic) — ปกติใช้ค่า default
	now        func() time.Time
	newOTP     func() string
	newRefresh func() (string, error)
}

var _ input.AuthUseCase = (*AuthService)(nil)

func NewAuthService(
	users output.UserRepository,
	otps output.OTPRepository,
	sessions output.SessionRepository,
	oauth output.OAuthRepository,
	mailer output.Mailer,
	tokens output.Tokenizer,
	google output.GoogleVerifier,
	tx output.TxManager,
	cfg AuthConfig,
) *AuthService {
	if cfg.OTPLength <= 0 {
		cfg.OTPLength = 6
	}
	if cfg.OTPMaxAttempts <= 0 {
		cfg.OTPMaxAttempts = 5
	}
	return &AuthService{
		users:      users,
		otps:       otps,
		sessions:   sessions,
		oauth:      oauth,
		mailer:     mailer,
		tokens:     tokens,
		google:     google,
		tx:         tx,
		cfg:        cfg,
		now:        time.Now,
		newOTP:     func() string { return generateNumericOTP(cfg.OTPLength) },
		newRefresh: generateRefreshToken,
	}
}

// RequestOTP — ขอ OTP ด้วยอีเมลโดยตรง
func (s *AuthService) RequestOTP(ctx context.Context, cmd input.RequestOTPCommand) (input.OTPChallenge, error) {
	email := domain.NormalizeEmail(cmd.Email)
	if email == "" {
		return input.OTPChallenge{}, domain.ErrInvalidInput
	}

	// ผูก user_id ถ้ามี user อยู่แล้ว (ไม่มีก็ปล่อย nil — จะสร้างตอน verify)
	var userID *uint
	var displayName string
	if u, err := s.users.FindByEmail(ctx, email); err == nil {
		userID = &u.ID
		displayName = u.DisplayName
	} else if !errors.Is(err, domain.ErrNotFound) {
		return input.OTPChallenge{}, err
	}

	if err := s.issueOTP(ctx, email, userID, cmd.IP); err != nil {
		return input.OTPChallenge{}, err
	}

	return input.OTPChallenge{
		Email:       email,
		DisplayName: displayName,
		ExpiresIn:   int(s.cfg.OTPTTL.Seconds()),
	}, nil
}

// LoginWithGoogle — verify id_token, สร้าง user ถ้ายังไม่มี, แล้วส่ง OTP ต่อ (ตาม rest_api 2.2)
func (s *AuthService) LoginWithGoogle(ctx context.Context, cmd input.GoogleLoginCommand) (input.OTPChallenge, error) {
	identity, err := s.google.Verify(ctx, cmd.IDToken)
	if err != nil {
		return input.OTPChallenge{}, err
	}
	email := domain.NormalizeEmail(identity.Email)
	if email == "" || !identity.EmailVerified {
		return input.OTPChallenge{}, domain.ErrGoogleToken
	}

	var user *domain.User
	err = s.tx.WithinTx(ctx, func(ctx context.Context) error {
		u, err := s.users.FindByEmail(ctx, email)
		if errors.Is(err, domain.ErrNotFound) {
			u = &domain.User{
				Email:         email,
				EmailVerified: true, // Google ยืนยันอีเมลให้แล้ว
				DisplayName:   identity.Name,
				AvatarURL:     identity.Picture,
				Role:          domain.RoleCustomer,
				IsActive:      true,
			}
			if err := s.users.Create(ctx, u); err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
		user = u
		return s.oauth.Upsert(ctx, &domain.OAuthAccount{
			UserID:         u.ID,
			Provider:       "google",
			ProviderUserID: identity.Subject,
			ProviderEmail:  email,
		})
	})
	if err != nil {
		return input.OTPChallenge{}, err
	}

	if err := s.issueOTP(ctx, email, &user.ID, cmd.IP); err != nil {
		return input.OTPChallenge{}, err
	}

	return input.OTPChallenge{
		Email:       email,
		DisplayName: user.DisplayName,
		ExpiresIn:   int(s.cfg.OTPTTL.Seconds()),
	}, nil
}

// issueOTP — สร้าง OTP ใหม่ (ยกเลิกอันเก่า), เก็บแฮช, แล้วส่งอีเมล
func (s *AuthService) issueOTP(ctx context.Context, email string, userID *uint, ip string) error {
	if err := s.otps.InvalidateActive(ctx, email, domain.OTPPurposeLogin); err != nil {
		return err
	}

	code := s.newOTP()
	now := s.now()
	otp := &domain.OTPCode{
		UserID:      userID,
		Email:       email,
		Purpose:     domain.OTPPurposeLogin,
		CodeHash:    hashOTP(s.cfg.OTPSecret, code),
		ExpiresAt:   now.Add(s.cfg.OTPTTL),
		Attempts:    0,
		MaxAttempts: s.cfg.OTPMaxAttempts,
		RequestIP:   ip,
		CreatedAt:   now,
	}
	if err := s.otps.Create(ctx, otp); err != nil {
		return err
	}

	// ส่งอีเมลหลังบันทึกแฮชแล้ว — โค้ดจริง (code) อยู่ในตัวแปรนี้เท่านั้น ห้าม log
	return s.mailer.SendOTP(ctx, email, code, int(s.cfg.OTPTTL.Seconds()))
}

// VerifyOTP — ยืนยัน OTP → ออก token
func (s *AuthService) VerifyOTP(ctx context.Context, cmd input.VerifyOTPCommand) (*input.AuthResult, error) {
	email := domain.NormalizeEmail(cmd.Email)
	if email == "" || cmd.Code == "" {
		return nil, domain.ErrInvalidOTP
	}

	now := s.now()
	providedHash := hashOTP(s.cfg.OTPSecret, cmd.Code)

	refreshToken, err := s.newRefresh()
	if err != nil {
		return nil, err
	}

	var user *domain.User
	err = s.tx.WithinTx(ctx, func(ctx context.Context) error {
		otp, err := s.otps.FindLatestActive(ctx, email, domain.OTPPurposeLogin)
		if errors.Is(err, domain.ErrNotFound) {
			return domain.ErrInvalidOTP
		} else if err != nil {
			return err
		}

		if verifyErr := otp.Verify(providedHash, now); verifyErr != nil {
			// persist attempts ที่เพิ่มขึ้นเสมอ แม้ verify จะ fail
			if saveErr := s.otps.Save(ctx, otp); saveErr != nil {
				return saveErr
			}
			return verifyErr
		}
		if err := s.otps.Save(ctx, otp); err != nil {
			return err
		}

		// หา user หรือสร้างใหม่ (กรณีขอ OTP ด้วยอีเมลตรงครั้งแรก)
		u, err := s.users.FindByEmail(ctx, email)
		if errors.Is(err, domain.ErrNotFound) {
			u = &domain.User{
				Email:         email,
				EmailVerified: true,
				Role:          domain.RoleCustomer,
				IsActive:      true,
			}
			if err := s.users.Create(ctx, u); err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
		if !u.IsActive {
			return domain.ErrInactiveUser
		}

		u.EmailVerified = true
		u.LastLoginAt = &now
		if err := s.users.Update(ctx, u); err != nil {
			return err
		}
		user = u

		return s.sessions.Create(ctx, &domain.Session{
			UserID:           u.ID,
			RefreshTokenHash: hashToken(refreshToken),
			UserAgent:        cmd.UserAgent,
			IPAddress:        cmd.IP,
			ExpiresAt:        now.Add(s.cfg.RefreshTTL),
			CreatedAt:        now,
		})
	})
	if err != nil {
		return nil, err
	}

	access, expiresIn, err := s.tokens.IssueAccess(user)
	if err != nil {
		return nil, err
	}
	return &input.AuthResult{
		AccessToken:  access,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
		User:         user,
	}, nil
}

// Refresh — ต่ออายุ access token + rotate refresh token
func (s *AuthService) Refresh(ctx context.Context, cmd input.RefreshCommand) (*input.AuthResult, error) {
	if cmd.RefreshToken == "" {
		return nil, domain.ErrInvalidToken
	}
	now := s.now()
	oldHash := hashToken(cmd.RefreshToken)

	newRefresh, err := s.newRefresh()
	if err != nil {
		return nil, err
	}

	var user *domain.User
	err = s.tx.WithinTx(ctx, func(ctx context.Context) error {
		sess, err := s.sessions.FindByTokenHash(ctx, oldHash)
		if errors.Is(err, domain.ErrNotFound) {
			return domain.ErrInvalidToken
		} else if err != nil {
			return err
		}
		if !sess.IsValid(now) {
			return domain.ErrInvalidToken
		}

		// rotation: เพิกถอนอันเก่า สร้างอันใหม่
		sess.Revoke(now)
		if err := s.sessions.Save(ctx, sess); err != nil {
			return err
		}

		u, err := s.users.FindByID(ctx, sess.UserID)
		if err != nil {
			return err
		}
		if !u.IsActive {
			return domain.ErrInactiveUser
		}
		user = u

		return s.sessions.Create(ctx, &domain.Session{
			UserID:           u.ID,
			RefreshTokenHash: hashToken(newRefresh),
			UserAgent:        cmd.UserAgent,
			IPAddress:        cmd.IP,
			ExpiresAt:        now.Add(s.cfg.RefreshTTL),
			CreatedAt:        now,
		})
	})
	if err != nil {
		return nil, err
	}

	access, expiresIn, err := s.tokens.IssueAccess(user)
	if err != nil {
		return nil, err
	}
	return &input.AuthResult{
		AccessToken:  access,
		RefreshToken: newRefresh,
		ExpiresIn:    expiresIn,
		User:         user,
	}, nil
}

// Logout — เพิกถอน session ของ refresh token ที่ให้มา (ต้องเป็นของ user ที่ล็อกอินอยู่)
func (s *AuthService) Logout(ctx context.Context, cmd input.LogoutCommand) error {
	if cmd.RefreshToken == "" {
		return nil // ไม่มี token ให้เพิกถอน ถือว่า logout สำเร็จเงียบ ๆ
	}
	now := s.now()
	hash := hashToken(cmd.RefreshToken)

	sess, err := s.sessions.FindByTokenHash(ctx, hash)
	if errors.Is(err, domain.ErrNotFound) {
		return nil
	} else if err != nil {
		return err
	}
	// ownership: session ต้องเป็นของ user ที่ล็อกอินอยู่ กัน revoke ข้ามคน
	if cmd.UserID != 0 && sess.UserID != cmd.UserID {
		return domain.ErrForbidden
	}
	sess.Revoke(now)
	return s.sessions.Save(ctx, sess)
}
