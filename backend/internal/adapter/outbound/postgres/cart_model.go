package postgres

// persistence model ของตะกร้า — gorm tag อยู่ที่นี่เท่านั้น
// (item ดึงแบบ join flat ใน cart_repo.go จึงไม่ต้องมี row struct ของ cart_item)

type cartRow struct {
	ID     uint   `gorm:"column:id;primaryKey"`
	UserID uint   `gorm:"column:user_id"`
	Status string `gorm:"column:status"`
}

func (cartRow) TableName() string { return "carts" }
