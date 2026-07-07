package service

import (
	"context"
	"strings"

	"github.com/khaopod/backend/internal/core/domain"
	"github.com/khaopod/backend/internal/core/port/input"
	"github.com/khaopod/backend/internal/core/port/output"
)

// AddressService = use case ของที่อยู่จัดส่ง (implements input.AddressUseCase)
// ทุกเมธอด scope ด้วย userID — repo คืน ErrNotFound เมื่อที่อยู่ไม่ใช่ของ user (กัน IDOR)
type AddressService struct {
	addrs output.AddressRepository
	tx    output.TxManager
}

var _ input.AddressUseCase = (*AddressService)(nil)

func NewAddressService(addrs output.AddressRepository, tx output.TxManager) *AddressService {
	return &AddressService{addrs: addrs, tx: tx}
}

func (s *AddressService) List(ctx context.Context, userID uint) ([]domain.Address, error) {
	return s.addrs.ListByUser(ctx, userID)
}

func (s *AddressService) Get(ctx context.Context, userID, id uint) (*domain.Address, error) {
	return s.addrs.FindByID(ctx, userID, id)
}

func (s *AddressService) Create(ctx context.Context, userID uint, cmd input.AddressCommand) (*domain.Address, error) {
	existing, err := s.addrs.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	a := applyCommand(&domain.Address{UserID: userID, Country: "TH"}, cmd)
	// ที่อยู่แรกของผู้ใช้ให้เป็น default เสมอ — checkout จะได้มีที่อยู่หลักให้เลือกทันที
	if len(existing) == 0 {
		a.IsDefault = true
	}

	err = s.tx.WithinTx(ctx, func(ctx context.Context) error {
		if a.IsDefault {
			// DB มี unique partial index คุม default ต่อ user — ต้องปลดตัวเก่าก่อนใน tx เดียวกัน
			if err := s.addrs.ClearDefault(ctx, userID); err != nil {
				return err
			}
		}
		return s.addrs.Create(ctx, a)
	})
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (s *AddressService) Update(ctx context.Context, userID, id uint, cmd input.AddressCommand) (*domain.Address, error) {
	a, err := s.addrs.FindByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	wasDefault := a.IsDefault
	a = applyCommand(a, cmd)
	// ห้ามปลด default ผ่าน PATCH ตรง ๆ (ต้องย้ายไปตั้งที่อยู่อื่นแทน) — กันผู้ใช้ไม่มีที่อยู่หลักเลย
	if wasDefault {
		a.IsDefault = true
	}

	err = s.tx.WithinTx(ctx, func(ctx context.Context) error {
		if a.IsDefault && !wasDefault {
			if err := s.addrs.ClearDefault(ctx, userID); err != nil {
				return err
			}
		}
		return s.addrs.Update(ctx, a)
	})
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (s *AddressService) Delete(ctx context.Context, userID, id uint) error {
	// เช็คก่อนว่ามีจริงและเป็นของ user นี้ เพื่อให้ตอบ 404 ได้ถูก (DELETE เฉย ๆ แยกไม่ออก)
	if _, err := s.addrs.FindByID(ctx, userID, id); err != nil {
		return err
	}
	return s.addrs.Delete(ctx, userID, id)
}

func (s *AddressService) SetDefault(ctx context.Context, userID, id uint) (*domain.Address, error) {
	a, err := s.addrs.FindByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if a.IsDefault {
		return a, nil // เป็นหลักอยู่แล้ว ไม่ต้องทำอะไร
	}
	a.IsDefault = true
	err = s.tx.WithinTx(ctx, func(ctx context.Context) error {
		if err := s.addrs.ClearDefault(ctx, userID); err != nil {
			return err
		}
		return s.addrs.Update(ctx, a)
	})
	if err != nil {
		return nil, err
	}
	return a, nil
}

func applyCommand(a *domain.Address, cmd input.AddressCommand) *domain.Address {
	a.RecipientName = strings.TrimSpace(cmd.RecipientName)
	a.Phone = strings.TrimSpace(cmd.Phone)
	a.AddressLine = strings.TrimSpace(cmd.AddressLine)
	a.Subdistrict = strings.TrimSpace(cmd.Subdistrict)
	a.District = strings.TrimSpace(cmd.District)
	a.Province = strings.TrimSpace(cmd.Province)
	a.PostalCode = strings.TrimSpace(cmd.PostalCode)
	a.Note = strings.TrimSpace(cmd.Note)
	a.IsDefault = cmd.IsDefault
	return a
}
