package domain

import "time"

// Address = ที่อยู่จัดส่งของผู้ใช้ — struct ล้วน ไม่มี tag ของ gorm
type Address struct {
	ID            uint
	UserID        uint
	RecipientName string
	Phone         string
	AddressLine   string
	Subdistrict   string
	District      string
	Province      string
	PostalCode    string
	Country       string
	Note          string
	IsDefault     bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
