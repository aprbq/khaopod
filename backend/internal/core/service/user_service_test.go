package service

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/khaopod/backend/internal/core/domain"
	"github.com/khaopod/backend/internal/core/port/input"
)

func TestGetProfile_NotFound(t *testing.T) {
	repo := newFakeUserRepo()
	svc := NewUserService(repo, &fakeFileStorage{})
	_, err := svc.GetProfile(context.Background(), 999)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestUpdateProfile_UpdatesOnlyProvidedFields(t *testing.T) {
	repo := newFakeUserRepo()
	u := repo.seed(&domain.User{Email: "a@b.com", DisplayName: "เดิม", Phone: "0810000000"})
	svc := NewUserService(repo, &fakeFileStorage{})

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

func TestUpdateAvatar_SavesFileAndRemovesOld(t *testing.T) {
	repo := newFakeUserRepo()
	u := repo.seed(&domain.User{Email: "a@b.com", AvatarURL: "https://lh3.googleusercontent.com/photo.jpg"})
	files := &fakeFileStorage{}
	svc := NewUserService(repo, files)

	content := []byte("png-bytes")
	got, err := svc.UpdateAvatar(context.Background(), u.ID, input.UpdateAvatarCommand{Content: content, Ext: ".png"})
	if err != nil {
		t.Fatalf("update avatar: %v", err)
	}
	if !strings.HasPrefix(got.AvatarURL, "/uploads/avatars/") || !strings.HasSuffix(got.AvatarURL, ".png") {
		t.Fatalf("avatar url should point to uploaded file, got %q", got.AvatarURL)
	}
	if !bytes.Equal(files.saved[got.AvatarURL], content) {
		t.Fatalf("file content not saved for %q", got.AvatarURL)
	}
	// รูปเดิม (จาก Google) ต้องถูกสั่งลบแบบ best-effort — storage เป็นคนเมิน URL ภายนอกเอง
	if len(files.removed) != 1 || files.removed[0] != "https://lh3.googleusercontent.com/photo.jpg" {
		t.Fatalf("old avatar should be removed, got %v", files.removed)
	}
}

func TestUpdateAvatar_UserNotFound(t *testing.T) {
	svc := NewUserService(newFakeUserRepo(), &fakeFileStorage{})
	_, err := svc.UpdateAvatar(context.Background(), 999, input.UpdateAvatarCommand{Content: []byte("x"), Ext: ".png"})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestUpdateAvatar_DBErrorCleansUpNewFile(t *testing.T) {
	repo := newFakeUserRepo()
	u := repo.seed(&domain.User{Email: "a@b.com"})
	repo.updateErr = errors.New("db down")
	files := &fakeFileStorage{}
	svc := NewUserService(repo, files)

	_, err := svc.UpdateAvatar(context.Background(), u.ID, input.UpdateAvatarCommand{Content: []byte("x"), Ext: ".jpg"})
	if err == nil {
		t.Fatal("want error when repo update fails")
	}
	// ไฟล์ที่เพิ่งเซฟต้องถูกลบทิ้ง ไม่เหลือไฟล์กำพร้า
	if len(files.saved) != 0 {
		t.Fatalf("new file should be cleaned up, still have %v", files.saved)
	}
}
