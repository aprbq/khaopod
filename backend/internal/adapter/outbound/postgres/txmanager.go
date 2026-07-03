package postgres

import (
	"context"

	"gorm.io/gorm"

	"github.com/khaopod/backend/internal/core/port/output"
)

// TxManager = Unit of Work ฝั่ง postgres (implements output.TxManager)
// เปิด transaction แล้วยัด *gorm.DB ลง context ให้ repo หยิบไปใช้ร่วมกัน
type TxManager struct {
	db *gorm.DB
}

var _ output.TxManager = (*TxManager)(nil)

func NewTxManager(db *gorm.DB) *TxManager { return &TxManager{db: db} }

func (m *TxManager) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	// ถ้าอยู่ใน tx อยู่แล้ว ใช้อันเดิม (nested-safe)
	if _, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return fn(ctx)
	}
	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(withTx(ctx, tx))
	})
}
