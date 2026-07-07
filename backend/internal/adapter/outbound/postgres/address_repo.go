package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/khaopod/backend/internal/core/domain"
	"github.com/khaopod/backend/internal/core/port/output"
)

type AddressRepo struct{ db *gorm.DB }

var _ output.AddressRepository = (*AddressRepo)(nil)

func NewAddressRepo(db *gorm.DB) *AddressRepo { return &AddressRepo{db: db} }

func (r *AddressRepo) ListByUser(ctx context.Context, userID uint) ([]domain.Address, error) {
	var rows []addressRow
	err := dbFromContext(ctx, r.db).
		Where("user_id = ?", userID).
		Order("is_default DESC, id"). // ที่อยู่หลักขึ้นก่อนเสมอ
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]domain.Address, 0, len(rows))
	for _, row := range rows {
		out = append(out, toAddressDomain(row))
	}
	return out, nil
}

func (r *AddressRepo) FindByID(ctx context.Context, userID, id uint) (*domain.Address, error) {
	var row addressRow
	// scope ด้วย user_id เสมอ — ของคนอื่นต้องมองไม่เห็น (กัน IDOR)
	err := dbFromContext(ctx, r.db).Where("id = ? AND user_id = ?", id, userID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	a := toAddressDomain(row)
	return &a, nil
}

func (r *AddressRepo) Create(ctx context.Context, a *domain.Address) error {
	row := toAddressRow(a)
	if err := dbFromContext(ctx, r.db).Create(&row).Error; err != nil {
		return err
	}
	a.ID = row.ID
	a.CreatedAt = row.CreatedAt
	a.UpdatedAt = row.UpdatedAt
	return nil
}

func (r *AddressRepo) Update(ctx context.Context, a *domain.Address) error {
	row := toAddressRow(a)
	// Select ระบุคอลัมน์ชัด ๆ เพื่อให้ค่า false/ค่าว่างถูกเขียนด้วย (gorm ข้าม zero value ถ้าไม่ระบุ)
	return dbFromContext(ctx, r.db).Model(&addressRow{}).
		Where("id = ? AND user_id = ?", a.ID, a.UserID).
		Select("recipient_name", "phone", "address_line", "subdistrict", "district",
			"province", "postal_code", "note", "is_default").
		Updates(&row).Error
}

func (r *AddressRepo) Delete(ctx context.Context, userID, id uint) error {
	return dbFromContext(ctx, r.db).
		Where("id = ? AND user_id = ?", id, userID).
		Delete(&addressRow{}).Error
}

func (r *AddressRepo) ClearDefault(ctx context.Context, userID uint) error {
	return dbFromContext(ctx, r.db).Model(&addressRow{}).
		Where("user_id = ? AND is_default = ?", userID, true).
		Update("is_default", false).Error
}
