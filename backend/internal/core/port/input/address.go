package input

import (
	"context"

	"github.com/khaopod/backend/internal/core/domain"
)

// AddressUseCase = จัดการที่อยู่จัดส่งของผู้ใช้ที่ล็อกอินอยู่ (driving port)
// ทุกเมธอดรับ userID เพื่อ scope ความเป็นเจ้าของ — กัน IDOR ที่ชั้น use case
type AddressUseCase interface {
	List(ctx context.Context, userID uint) ([]domain.Address, error)
	Get(ctx context.Context, userID, id uint) (*domain.Address, error)
	Create(ctx context.Context, userID uint, cmd AddressCommand) (*domain.Address, error)
	Update(ctx context.Context, userID, id uint, cmd AddressCommand) (*domain.Address, error)
	Delete(ctx context.Context, userID, id uint) error
	SetDefault(ctx context.Context, userID, id uint) (*domain.Address, error)
}

// AddressCommand — ข้อมูลที่อยู่ที่ inbound adapter validate รูปแบบแล้ว (เช่น postal_code 5 หลัก)
type AddressCommand struct {
	RecipientName string
	Phone         string
	AddressLine   string
	Subdistrict   string
	District      string
	Province      string
	PostalCode    string
	Note          string
	IsDefault     bool
}
