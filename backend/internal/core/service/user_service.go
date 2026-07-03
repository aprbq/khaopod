package service

import (
	"context"

	"github.com/khaopod/backend/internal/core/domain"
	"github.com/khaopod/backend/internal/core/port/input"
	"github.com/khaopod/backend/internal/core/port/output"
)

// UserService = use case ของโปรไฟล์ผู้ใช้ (implements input.UserUseCase)
type UserService struct {
	users output.UserRepository
}

var _ input.UserUseCase = (*UserService)(nil)

func NewUserService(users output.UserRepository) *UserService {
	return &UserService{users: users}
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
