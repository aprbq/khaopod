package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/khaopod/backend/internal/core/domain"
	"github.com/khaopod/backend/internal/core/port/output"
)

type UserRepo struct{ db *gorm.DB }

var _ output.UserRepository = (*UserRepo)(nil)

func NewUserRepo(db *gorm.DB) *UserRepo { return &UserRepo{db: db} }

func (r *UserRepo) FindByID(ctx context.Context, id uint) (*domain.User, error) {
	return r.findBy(ctx, "id = ?", id)
}

func (r *UserRepo) FindByPublicID(ctx context.Context, publicID string) (*domain.User, error) {
	return r.findBy(ctx, "public_id = ?", publicID)
}

func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	return r.findBy(ctx, "email = ?", email)
}

func (r *UserRepo) findBy(ctx context.Context, query string, arg any) (*domain.User, error) {
	var row userRow
	err := dbFromContext(ctx, r.db).Where(query, arg).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	u := toUserDomain(row)
	return &u, nil
}

func (r *UserRepo) Create(ctx context.Context, u *domain.User) error {
	if u.PublicID == "" {
		u.PublicID = uuid.NewString()
	}
	if u.Role == "" {
		u.Role = domain.RoleCustomer
	}
	row := toUserRow(u)
	if err := dbFromContext(ctx, r.db).Create(&row).Error; err != nil {
		return err
	}
	// เขียนค่าที่ DB สร้างกลับเข้า entity
	u.ID = row.ID
	u.PublicID = row.PublicID
	u.CreatedAt = row.CreatedAt
	u.UpdatedAt = row.UpdatedAt
	return nil
}

func (r *UserRepo) Update(ctx context.Context, u *domain.User) error {
	row := toUserRow(u)
	// อัปเดตเฉพาะฟิลด์ที่แก้ได้ (ไม่แตะ id/public_id/created_at)
	return dbFromContext(ctx, r.db).Model(&userRow{ID: u.ID}).Updates(map[string]any{
		"email":          row.Email,
		"email_verified": row.EmailVerified,
		"display_name":   row.DisplayName,
		"avatar_url":     row.AvatarURL,
		"phone":          row.Phone,
		"role":           row.Role,
		"is_active":      row.IsActive,
		"last_login_at":  row.LastLoginAt,
	}).Error
}
