package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/khaopod/backend/internal/core/domain"
	"github.com/khaopod/backend/internal/core/port/output"
)

// JWTTokenizer ออก/ตรวจ access token แบบ HS256 (implements output.Tokenizer)
type JWTTokenizer struct {
	secret []byte
	ttl    time.Duration
	now    func() time.Time
}

var _ output.Tokenizer = (*JWTTokenizer)(nil)

func NewJWTTokenizer(secret string, ttl time.Duration) *JWTTokenizer {
	return &JWTTokenizer{secret: []byte(secret), ttl: ttl, now: time.Now}
}

// claims ที่เก็บใน token — sub = public_id (ไม่ leak internal id ออกไป)
type accessClaims struct {
	UID  uint        `json:"uid"`
	Role domain.Role `json:"role"`
	jwt.RegisteredClaims
}

func (t *JWTTokenizer) IssueAccess(u *domain.User) (string, int, error) {
	now := t.now()
	claims := accessClaims{
		UID:  u.ID,
		Role: u.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   u.PublicID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(t.ttl)),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(t.secret)
	if err != nil {
		return "", 0, err
	}
	return token, int(t.ttl.Seconds()), nil
}

func (t *JWTTokenizer) ParseAccess(token string) (*output.AccessClaims, error) {
	var claims accessClaims
	_, err := jwt.ParseWithClaims(token, &claims, func(tok *jwt.Token) (any, error) {
		if _, ok := tok.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, domain.ErrInvalidToken
		}
		return t.secret, nil
	})
	if err != nil {
		// ทุก error ในการ parse/verify แปลงเป็น domain error เดียว (ไม่ leak รายละเอียด)
		return nil, errors.Join(domain.ErrInvalidToken, err)
	}
	return &output.AccessClaims{
		UserID:   claims.UID,
		PublicID: claims.Subject,
		Role:     claims.Role,
	}, nil
}
