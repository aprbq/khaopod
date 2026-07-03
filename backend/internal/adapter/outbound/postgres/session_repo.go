package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/khaopod/backend/internal/core/domain"
	"github.com/khaopod/backend/internal/core/port/output"
)

type SessionRepo struct{ db *gorm.DB }

var _ output.SessionRepository = (*SessionRepo)(nil)

func NewSessionRepo(db *gorm.DB) *SessionRepo { return &SessionRepo{db: db} }

func (r *SessionRepo) Create(ctx context.Context, s *domain.Session) error {
	row := sessionRow{
		UserID:           s.UserID,
		RefreshTokenHash: s.RefreshTokenHash,
		UserAgent:        nilIfEmpty(s.UserAgent),
		IPAddress:        nilIfEmpty(s.IPAddress),
		ExpiresAt:        s.ExpiresAt,
		RevokedAt:        s.RevokedAt,
	}
	if err := dbFromContext(ctx, r.db).Create(&row).Error; err != nil {
		return err
	}
	s.ID = row.ID
	s.CreatedAt = row.CreatedAt
	return nil
}

func (r *SessionRepo) FindByTokenHash(ctx context.Context, hash string) (*domain.Session, error) {
	var row sessionRow
	err := dbFromContext(ctx, r.db).Where("refresh_token_hash = ?", hash).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	s := toSessionDomain(row)
	return &s, nil
}

func (r *SessionRepo) Save(ctx context.Context, s *domain.Session) error {
	return dbFromContext(ctx, r.db).Model(&sessionRow{ID: s.ID}).
		Update("revoked_at", s.RevokedAt).Error
}
