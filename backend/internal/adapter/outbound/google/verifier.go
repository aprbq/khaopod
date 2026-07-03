package google

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/khaopod/backend/internal/core/domain"
	"github.com/khaopod/backend/internal/core/port/output"
)

const certsURL = "https://www.googleapis.com/oauth2/v3/certs"

var validIssuers = []string{"accounts.google.com", "https://accounts.google.com"}

// Verifier ตรวจ Google id_token ด้วย public key จาก JWKS ของ Google (implements output.GoogleVerifier)
// ไม่พึ่ง SDK หนัก ๆ — ดึง certs เอง + cache ตามอายุ
type Verifier struct {
	clientID string
	client   *http.Client

	mu        sync.Mutex
	keys      map[string]*rsa.PublicKey
	fetchedAt time.Time
	ttl       time.Duration
}

var _ output.GoogleVerifier = (*Verifier)(nil)

func NewVerifier(clientID string) *Verifier {
	return &Verifier{
		clientID: clientID,
		client:   &http.Client{Timeout: 10 * time.Second},
		keys:     map[string]*rsa.PublicKey{},
		ttl:      time.Hour,
	}
}

func (v *Verifier) Verify(ctx context.Context, idToken string) (*output.GoogleIdentity, error) {
	if v.clientID == "" {
		// ไม่ได้ตั้งค่า Google client id → ถือว่าใช้ไม่ได้ (ไม่ verify แบบหลวม ๆ)
		return nil, domain.ErrGoogleToken
	}

	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(idToken, claims, v.keyfunc(ctx),
		jwt.WithAudience(v.clientID),
		jwt.WithIssuer(validIssuers[0]),
		jwt.WithValidMethods([]string{"RS256"}),
	)
	if err != nil {
		// issuer อาจเป็นอีกรูปแบบ (มี https://) ลองใหม่แบบไม่ผูก issuer แล้วเช็คเอง
		if _, err2 := jwt.ParseWithClaims(idToken, claims, v.keyfunc(ctx),
			jwt.WithAudience(v.clientID),
			jwt.WithValidMethods([]string{"RS256"}),
		); err2 != nil {
			return nil, domain.ErrGoogleToken
		}
		if !validIssuer(claims) {
			return nil, domain.ErrGoogleToken
		}
	}

	sub, _ := claims["sub"].(string)
	email, _ := claims["email"].(string)
	if sub == "" || email == "" {
		return nil, domain.ErrGoogleToken
	}
	name, _ := claims["name"].(string)
	picture, _ := claims["picture"].(string)
	verified, _ := claims["email_verified"].(bool)

	return &output.GoogleIdentity{
		Subject:       sub,
		Email:         email,
		EmailVerified: verified,
		Name:          name,
		Picture:       picture,
	}, nil
}

func validIssuer(claims jwt.MapClaims) bool {
	iss, _ := claims["iss"].(string)
	for _, ok := range validIssuers {
		if iss == ok {
			return true
		}
	}
	return false
}

// keyfunc เลือก RSA public key ตาม kid ใน header
func (v *Verifier) keyfunc(ctx context.Context) jwt.Keyfunc {
	return func(token *jwt.Token) (any, error) {
		kid, _ := token.Header["kid"].(string)
		if kid == "" {
			return nil, domain.ErrGoogleToken
		}
		key, err := v.keyByID(ctx, kid)
		if err != nil {
			return nil, err
		}
		return key, nil
	}
}

func (v *Verifier) keyByID(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	v.mu.Lock()
	defer v.mu.Unlock()

	if key, ok := v.keys[kid]; ok && time.Since(v.fetchedAt) < v.ttl {
		return key, nil
	}
	if err := v.refresh(ctx); err != nil {
		return nil, err
	}
	if key, ok := v.keys[kid]; ok {
		return key, nil
	}
	return nil, domain.ErrGoogleToken
}

type jwk struct {
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// refresh ดึง JWKS ล่าสุดจาก Google (ผู้เรียกถือ lock อยู่แล้ว)
func (v *Verifier) refresh(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, certsURL, nil)
	if err != nil {
		return err
	}
	resp, err := v.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var body struct {
		Keys []jwk `json:"keys"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return err
	}

	keys := make(map[string]*rsa.PublicKey, len(body.Keys))
	for _, k := range body.Keys {
		pub, err := k.toRSA()
		if err != nil {
			continue
		}
		keys[k.Kid] = pub
	}
	v.keys = keys
	v.fetchedAt = time.Now()
	return nil
}

func (k jwk) toRSA() (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(k.N)
	if err != nil {
		return nil, err
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(k.E)
	if err != nil {
		return nil, err
	}
	// exponent เป็น big-endian — เติม 0 ข้างหน้าให้ครบ 4 ไบต์ก่อนตีความเป็น uint32
	if len(eBytes) > 4 {
		return nil, domain.ErrGoogleToken
	}
	epadded := make([]byte, 4)
	copy(epadded[4-len(eBytes):], eBytes)
	return &rsa.PublicKey{
		N: new(big.Int).SetBytes(nBytes),
		E: int(binary.BigEndian.Uint32(epadded)),
	}, nil
}
