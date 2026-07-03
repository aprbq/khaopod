package postgres

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/khaopod/backend/internal/core/domain"
	"github.com/khaopod/backend/internal/core/port/output"
)

type OTPRepo struct{ db *gorm.DB }

var _ output.OTPRepository = (*OTPRepo)(nil)

func NewOTPRepo(db *gorm.DB) *OTPRepo { return &OTPRepo{db: db} }

func (r *OTPRepo) Create(ctx context.Context, o *domain.OTPCode) error {
	row := otpRow{
		UserID:      o.UserID,
		Email:       o.Email,
		Purpose:     string(o.Purpose),
		CodeHash:    o.CodeHash,
		ExpiresAt:   o.ExpiresAt,
		Attempts:    o.Attempts,
		MaxAttempts: o.MaxAttempts,
		RequestIP:   nilIfEmpty(o.RequestIP),
	}
	if err := dbFromContext(ctx, r.db).Create(&row).Error; err != nil {
		return err
	}
	o.ID = row.ID
	o.CreatedAt = row.CreatedAt
	return nil
}

func (r *OTPRepo) FindLatestActive(ctx context.Context, email string, purpose domain.OTPPurpose) (*domain.OTPCode, error) {
	var row otpRow
	err := dbFromContext(ctx, r.db).
		Where("email = ? AND purpose = ? AND consumed_at IS NULL", email, string(purpose)).
		Order("created_at DESC").
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	o := toOTPDomain(row)
	return &o, nil
}

func (r *OTPRepo) Save(ctx context.Context, o *domain.OTPCode) error {
	return dbFromContext(ctx, r.db).Model(&otpRow{ID: o.ID}).Updates(map[string]any{
		"attempts":    o.Attempts,
		"consumed_at": o.ConsumedAt,
	}).Error
}

func (r *OTPRepo) InvalidateActive(ctx context.Context, email string, purpose domain.OTPPurpose) error {
	now := time.Now()
	return dbFromContext(ctx, r.db).Model(&otpRow{}).
		Where("email = ? AND purpose = ? AND consumed_at IS NULL", email, string(purpose)).
		Update("consumed_at", now).Error
}
