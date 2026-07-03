package input

import (
	"context"

	"github.com/khaopod/backend/internal/core/domain"
)

// UserUseCase = จัดการโปรไฟล์ผู้ใช้ที่ล็อกอินอยู่ (driving port)
type UserUseCase interface {
	GetProfile(ctx context.Context, userID uint) (*domain.User, error)
	UpdateProfile(ctx context.Context, userID uint, cmd UpdateProfileCommand) (*domain.User, error)
}

// UpdateProfileCommand — ฟิลด์เป็น pointer เพื่อแยก "ไม่ส่งมา" (nil) ออกจาก "ส่งค่าว่าง"
type UpdateProfileCommand struct {
	DisplayName *string
	Phone       *string
}
