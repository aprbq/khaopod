package postgres

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/khaopod/backend/internal/core/domain"
	"github.com/khaopod/backend/internal/core/port/output"
)

type OAuthRepo struct{ db *gorm.DB }

var _ output.OAuthRepository = (*OAuthRepo)(nil)

func NewOAuthRepo(db *gorm.DB) *OAuthRepo { return &OAuthRepo{db: db} }

func (r *OAuthRepo) Upsert(ctx context.Context, a *domain.OAuthAccount) error {
	row := oauthRow{
		UserID:         a.UserID,
		Provider:       a.Provider,
		ProviderUserID: a.ProviderUserID,
		ProviderEmail:  nilIfEmpty(a.ProviderEmail),
	}
	// ชนกันที่ (provider, provider_user_id) → อัปเดต email/user ให้ล่าสุด
	return dbFromContext(ctx, r.db).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "provider"}, {Name: "provider_user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"user_id", "provider_email"}),
	}).Create(&row).Error
}
