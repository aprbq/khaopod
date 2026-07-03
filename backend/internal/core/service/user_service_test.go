package service

import (
	"context"
	"errors"
	"testing"

	"github.com/khaopod/backend/internal/core/domain"
	"github.com/khaopod/backend/internal/core/port/input"
)

func TestGetProfile_NotFound(t *testing.T) {
	repo := newFakeUserRepo()
	svc := NewUserService(repo)
	_, err := svc.GetProfile(context.Background(), 999)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestUpdateProfile_UpdatesOnlyProvidedFields(t *testing.T) {
	repo := newFakeUserRepo()
	u := repo.seed(&domain.User{Email: "a@b.com", DisplayName: "เดิม", Phone: "0810000000"})
	svc := NewUserService(repo)

	newName := "  ชื่อใหม่  "
	got, err := svc.UpdateProfile(context.Background(), u.ID, input.UpdateProfileCommand{DisplayName: &newName})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if got.DisplayName != "ชื่อใหม่" {
		t.Fatalf("display name should be trimmed/updated, got %q", got.DisplayName)
	}
	// phone ไม่ได้ส่งมา (nil) → ต้องไม่ถูกแตะ
	if got.Phone != "0810000000" {
		t.Fatalf("phone should be unchanged, got %q", got.Phone)
	}
}
