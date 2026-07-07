package service

import (
	"context"
	"errors"
	"testing"

	"github.com/khaopod/backend/internal/core/domain"
	"github.com/khaopod/backend/internal/core/port/input"
)

func addrCmd(name string, isDefault bool) input.AddressCommand {
	return input.AddressCommand{
		RecipientName: name,
		Phone:         "0812345678",
		AddressLine:   "99/1 หมู่ 4",
		Subdistrict:   "ในเมือง",
		District:      "เมืองขอนแก่น",
		Province:      "ขอนแก่น",
		PostalCode:    "40000",
		IsDefault:     isDefault,
	}
}

func TestAddressCreate_FirstAddressBecomesDefault(t *testing.T) {
	repo := newFakeAddressRepo()
	svc := NewAddressService(repo, fakeTx{})

	a, err := svc.Create(context.Background(), 1, addrCmd("สมชาย", false))
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if !a.IsDefault {
		t.Fatal("first address should be forced default")
	}
}

func TestAddressCreate_NewDefaultClearsOld(t *testing.T) {
	repo := newFakeAddressRepo()
	old := repo.seed(&domain.Address{UserID: 1, RecipientName: "เดิม", IsDefault: true})
	svc := NewAddressService(repo, fakeTx{})

	a, err := svc.Create(context.Background(), 1, addrCmd("ใหม่", true))
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if !a.IsDefault {
		t.Fatal("new address should be default")
	}
	if repo.byID[old.ID].IsDefault {
		t.Fatal("old default should be cleared")
	}
}

func TestAddressUpdate_CannotUnsetDefault(t *testing.T) {
	repo := newFakeAddressRepo()
	a := repo.seed(&domain.Address{UserID: 1, RecipientName: "หลัก", IsDefault: true})
	svc := NewAddressService(repo, fakeTx{})

	got, err := svc.Update(context.Background(), 1, a.ID, addrCmd("หลัก", false))
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	// ปลด default ผ่าน PATCH ไม่ได้ — ต้องตั้งที่อยู่อื่นเป็นหลักแทน
	if !got.IsDefault {
		t.Fatal("default flag must not be unset via update")
	}
}

func TestAddressSetDefault_SwitchesDefault(t *testing.T) {
	repo := newFakeAddressRepo()
	a := repo.seed(&domain.Address{UserID: 1, RecipientName: "A", IsDefault: true})
	b := repo.seed(&domain.Address{UserID: 1, RecipientName: "B"})
	svc := NewAddressService(repo, fakeTx{})

	got, err := svc.SetDefault(context.Background(), 1, b.ID)
	if err != nil {
		t.Fatalf("set default: %v", err)
	}
	if !got.IsDefault {
		t.Fatal("B should be default now")
	}
	if repo.byID[a.ID].IsDefault {
		t.Fatal("A should no longer be default")
	}
}

func TestAddress_OwnershipEnforced(t *testing.T) {
	repo := newFakeAddressRepo()
	other := repo.seed(&domain.Address{UserID: 2, RecipientName: "คนอื่น"})
	svc := NewAddressService(repo, fakeTx{})

	// ทุก endpoint ที่จับ {id} ต้องมองไม่เห็นของคนอื่น (กัน IDOR)
	if _, err := svc.Get(context.Background(), 1, other.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("get: want ErrNotFound, got %v", err)
	}
	if _, err := svc.Update(context.Background(), 1, other.ID, addrCmd("x", false)); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("update: want ErrNotFound, got %v", err)
	}
	if err := svc.Delete(context.Background(), 1, other.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("delete: want ErrNotFound, got %v", err)
	}
	if _, err := svc.SetDefault(context.Background(), 1, other.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("set default: want ErrNotFound, got %v", err)
	}
}
