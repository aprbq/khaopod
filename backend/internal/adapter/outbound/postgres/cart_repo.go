package postgres

import (
	"context"
	"errors"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"github.com/khaopod/backend/internal/core/domain"
	"github.com/khaopod/backend/internal/core/port/output"
)

type CartRepo struct{ db *gorm.DB }

var _ output.CartRepository = (*CartRepo)(nil)

func NewCartRepo(db *gorm.DB) *CartRepo { return &CartRepo{db: db} }

func (r *CartRepo) Get(ctx context.Context, userID uint) (*domain.Cart, error) {
	db := dbFromContext(ctx, r.db)

	var cart cartRow
	err := db.Where("user_id = ? AND status = ?", userID, "active").First(&cart).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &domain.Cart{UserID: userID}, nil // ยังไม่มีตะกร้า = ตะกร้าว่าง
	}
	if err != nil {
		return nil, err
	}

	// join item → variant → product (+ รูปปก) เพื่อ enrich ข้อมูลแสดงผล
	type itemJoin struct {
		ItemID      uint
		VariantID   uint
		Quantity    int
		ProductName string
		VariantName string
		Color       *string
		UnitPrice   decimal.Decimal
		Stock       int
		Image       *string
	}
	var rows []itemJoin
	err = db.Table("cart_items AS ci").
		Select(`ci.id AS item_id, ci.product_variant_id AS variant_id, ci.quantity,
			p.name AS product_name, v.variant_name, v.color, v.price AS unit_price, v.stock_quantity AS stock,
			(SELECT url FROM product_images img WHERE img.product_id = p.id
			 ORDER BY img.is_primary DESC, img.sort_order, img.id LIMIT 1) AS image`).
		Joins("JOIN product_variants v ON v.id = ci.product_variant_id").
		Joins("JOIN products p ON p.id = v.product_id").
		Where("ci.cart_id = ?", cart.ID).
		Order("ci.added_at, ci.id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	dc := &domain.Cart{ID: cart.ID, UserID: userID}
	for _, row := range rows {
		dc.Items = append(dc.Items, domain.CartItem{
			ID:          row.ItemID,
			VariantID:   row.VariantID,
			ProductName: row.ProductName,
			VariantName: row.VariantName,
			Color:       deref(row.Color),
			UnitPrice:   row.UnitPrice,
			Image:       deref(row.Image),
			Stock:       row.Stock,
			Quantity:    row.Quantity,
		})
	}
	return dc, nil
}

func (r *CartRepo) UpsertItem(ctx context.Context, userID, variantID uint, qty int) error {
	db := dbFromContext(ctx, r.db)
	cartID, err := r.ensureActiveCart(db, userID)
	if err != nil {
		return err
	}
	// merge: มี variant นี้อยู่แล้วก็บวกจำนวน (unique cart_id+variant)
	return db.Exec(`
		INSERT INTO cart_items (cart_id, product_variant_id, quantity)
		VALUES (?, ?, ?)
		ON CONFLICT (cart_id, product_variant_id)
		DO UPDATE SET quantity = cart_items.quantity + EXCLUDED.quantity`,
		cartID, variantID, qty).Error
}

func (r *CartRepo) SetItemQty(ctx context.Context, userID, itemID uint, qty int) error {
	// scope ด้วย user_id กัน IDOR — แก้ได้เฉพาะ item ในตะกร้า active ของตัวเอง
	return dbFromContext(ctx, r.db).Exec(`
		UPDATE cart_items SET quantity = ?
		WHERE id = ? AND cart_id IN (SELECT id FROM carts WHERE user_id = ? AND status = 'active')`,
		qty, itemID, userID).Error
}

func (r *CartRepo) RemoveItem(ctx context.Context, userID, itemID uint) error {
	return dbFromContext(ctx, r.db).Exec(`
		DELETE FROM cart_items
		WHERE id = ? AND cart_id IN (SELECT id FROM carts WHERE user_id = ? AND status = 'active')`,
		itemID, userID).Error
}

func (r *CartRepo) Clear(ctx context.Context, userID uint) error {
	return dbFromContext(ctx, r.db).Exec(`
		DELETE FROM cart_items
		WHERE cart_id IN (SELECT id FROM carts WHERE user_id = ? AND status = 'active')`,
		userID).Error
}

// ensureActiveCart คืน id ของตะกร้า active — สร้างใหม่ถ้ายังไม่มี
func (r *CartRepo) ensureActiveCart(db *gorm.DB, userID uint) (uint, error) {
	var cart cartRow
	err := db.Where("user_id = ? AND status = ?", userID, "active").First(&cart).Error
	if err == nil {
		return cart.ID, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, err
	}
	cart = cartRow{UserID: userID, Status: "active"}
	if err := db.Create(&cart).Error; err != nil {
		return 0, err
	}
	return cart.ID, nil
}
