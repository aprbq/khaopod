package service

import (
	"context"
	"fmt"
	"time"

	"github.com/khaopod/backend/internal/core/domain"
	"github.com/khaopod/backend/internal/core/port/input"
	"github.com/khaopod/backend/internal/core/port/output"
)

// UserService = use case ของโปรไฟล์ผู้ใช้ (implements input.UserUseCase)
type UserService struct {
	users output.UserRepository
	files output.FileStorage
}

var _ input.UserUseCase = (*UserService)(nil)

func NewUserService(users output.UserRepository, files output.FileStorage) *UserService {
	return &UserService{users: users, files: files}
}

func (s *UserService) GetProfile(ctx context.Context, userID uint) (*domain.User, error) {
	return s.users.FindByID(ctx, userID)
}

func (s *UserService) UpdateProfile(ctx context.Context, userID uint, cmd input.UpdateProfileCommand) (*domain.User, error) {
	u, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	u.ApplyProfile(cmd.DisplayName, cmd.Phone)
	if err := s.users.Update(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

// UpdateAvatar เก็บรูปใหม่ผ่าน FileStorage แล้วสลับ URL ในโปรไฟล์
// ตั้งชื่อไฟล์ด้วย timestamp เพื่อให้ URL เปลี่ยนทุกครั้ง — browser จะไม่เสิร์ฟรูปเก่าจาก cache
func (s *UserService) UpdateAvatar(ctx context.Context, userID uint, cmd input.UpdateAvatarCommand) (*domain.User, error) {
	u, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	name := fmt.Sprintf("avatars/%s-%d%s", u.PublicID, time.Now().UnixMilli(), cmd.Ext)
	url, err := s.files.Save(ctx, name, cmd.Content)
	if err != nil {
		return nil, err
	}

	old := u.AvatarURL
	u.AvatarURL = url
	if err := s.users.Update(ctx, u); err != nil {
		_ = s.files.Remove(ctx, url) // อย่าทิ้งไฟล์กำพร้าถ้าบันทึก DB ไม่ผ่าน
		return nil, err
	}
	// ลบรูปเก่าแบบ best-effort (storage เมิน URL ภายนอกเช่นรูปจาก Google เอง)
	if old != "" && old != url {
		_ = s.files.Remove(ctx, old)
	}
	return u, nil
}
