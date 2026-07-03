package service

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"math/big"
	"strings"
)

// hashOTP ทำ HMAC-SHA256 ของโค้ดด้วย secret ก่อนเก็บ/เทียบ (เก็บเฉพาะแฮช ไม่เก็บเลขจริง)
func hashOTP(secret, code string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(strings.TrimSpace(code)))
	return hex.EncodeToString(mac.Sum(nil))
}

// hashToken ทำ SHA256 ของ refresh token ก่อนเก็บลง DB (เก็บเฉพาะแฮช)
func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// generateNumericOTP สร้างรหัส OTP ตัวเลขความยาว n แบบสุ่มปลอดภัย (crypto/rand)
func generateNumericOTP(n int) string {
	const digits = "0123456789"
	b := make([]byte, n)
	for i := range b {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			// crypto/rand พังถือเป็นเหตุร้ายแรง — panic ดีกว่าปล่อย OTP ที่เดาได้
			panic("service: secure random failed: " + err.Error())
		}
		b[i] = digits[idx.Int64()]
	}
	return string(b)
}

// generateRefreshToken สร้าง refresh token แบบสุ่ม 32 ไบต์ (hex)
func generateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
