package domain

import (
	"strings"
	"time"
)

type Role string

const (
	RoleCustomer Role = "customer"
	RoleAdmin    Role = "admin"
)

// User = ผู้ใช้งานระบบ (passwordless — ไม่มีฟิลด์รหัสผ่าน)
// เป็น struct ล้วน ไม่มี tag ของ gorm (persistence model แยกอยู่ใน postgres adapter)
type User struct {
	ID            uint
	PublicID      string // UUID สำหรับโชว์ภายนอก (ไม่ leak internal id)
	Email         string
	EmailVerified bool
	DisplayName   string
	AvatarURL     string
	Phone         string
	Role          Role
	IsActive      bool
	LastLoginAt   *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (u *User) IsAdmin() bool { return u.Role == RoleAdmin }

// ApplyProfile อัปเดตเฉพาะฟิลด์ที่ถูกส่งมา (nil = ไม่แตะ)
func (u *User) ApplyProfile(displayName, phone *string) {
	if displayName != nil {
		u.DisplayName = strings.TrimSpace(*displayName)
	}
	if phone != nil {
		u.Phone = strings.TrimSpace(*phone)
	}
}

// NormalizeEmail ทำให้อีเมลเทียบแบบ case-insensitive สม่ำเสมอ
// (คอลัมน์ใน DB เป็น CITEXT อยู่แล้ว แต่ normalize ฝั่งแอปช่วยให้ hash/lookup ตรงกัน)
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
