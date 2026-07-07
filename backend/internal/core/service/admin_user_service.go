package service

import (
	"context"

	"github.com/khaopod/backend/internal/core/domain"
	"github.com/khaopod/backend/internal/core/port/input"
	"github.com/khaopod/backend/internal/core/port/output"
)

// AdminUserService = use case ดูรายชื่อผู้ใช้ในหลังบ้าน (implements input.AdminUserUseCase)
type AdminUserService struct {
	users output.UserRepository
}

var _ input.AdminUserUseCase = (*AdminUserService)(nil)

func NewAdminUserService(users output.UserRepository) *AdminUserService {
	return &AdminUserService{users: users}
}

func (s *AdminUserService) ListUsers(ctx context.Context, limit, offset int) ([]domain.User, int, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return s.users.ListAll(ctx, limit, offset)
}
