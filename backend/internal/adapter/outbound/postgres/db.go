package postgres

import (
	"context"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Open เปิดการเชื่อมต่อ PostgreSQL ผ่าน GORM
func Open(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.New(postgres.Config{DSN: dsn}), &gorm.Config{
		// ปิด default transaction ต่อ statement (เราคุม tx เองผ่าน TxManager)
		SkipDefaultTransaction: true,
		Logger:                 logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	return db, sqlDB.Ping()
}

// txKey ใช้ยัด *gorm.DB (tx ปัจจุบัน) ลง context ระหว่าง WithinTx
type txKey struct{}

func withTx(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

// dbFromContext คืน tx ปัจจุบันถ้าอยู่ใน WithinTx, ไม่งั้นคืน db ปกติ
// ทุก repo ต้องเรียกอันนี้เพื่อให้ query ร่วม transaction เดียวกับที่ TxManager เปิด
func dbFromContext(ctx context.Context, def *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx.WithContext(ctx)
	}
	return def.WithContext(ctx)
}
