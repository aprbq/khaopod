package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/khaopod/backend/internal/core/domain"
	"github.com/khaopod/backend/internal/core/port/output"
)

type ProductRepo struct{ db *gorm.DB }

var _ output.ProductRepository = (*ProductRepo)(nil)

func NewProductRepo(db *gorm.DB) *ProductRepo { return &ProductRepo{db: db} }

func (r *ProductRepo) List(ctx context.Context, f output.ProductFilter) ([]domain.Product, int, error) {
	// base query — หน้าร้านเห็นเฉพาะสินค้าที่ active
	base := dbFromContext(ctx, r.db).Model(&productRow{}).Where("is_active = ?", true)
	if f.CategorySlug != "" {
		base = base.Where("category_id IN (SELECT id FROM categories WHERE slug = ?)", f.CategorySlug)
	}
	if f.Search != "" {
		base = base.Where("name ILIKE ?", "%"+f.Search+"%")
	}

	// นับทั้งหมดก่อน limit/offset (ไว้คำนวณ total_pages)
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// SortColumn มาจาก whitelist ใน service แล้ว จึงประกอบเป็น ORDER BY ได้อย่างปลอดภัย
	order := f.SortColumn
	if f.SortDesc {
		order += " DESC"
	} else {
		order += " ASC"
	}

	var rows []productRow
	err := base.
		Preload("Category").
		Preload("Variants", "is_active = ?", true).
		Preload("Images", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order, id") }).
		Order(order).
		Limit(f.Limit).
		Offset(f.Offset).
		Find(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	products := make([]domain.Product, 0, len(rows))
	for _, row := range rows {
		products = append(products, toProductDomain(row))
	}
	return products, int(total), nil
}

func (r *ProductRepo) FindBySlug(ctx context.Context, slug string) (*domain.Product, error) {
	var row productRow
	err := dbFromContext(ctx, r.db).
		Preload("Category").
		Preload("Variants", "is_active = ?", true).
		Preload("Images", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order, id") }).
		Where("slug = ? AND is_active = ?", slug, true).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	p := toProductDomain(row)
	return &p, nil
}

func (r *ProductRepo) FindVariantByID(ctx context.Context, id uint) (*domain.ProductVariant, error) {
	var row variantRow
	err := dbFromContext(ctx, r.db).
		Where("id = ? AND is_active = ?", id, true).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	v := domain.ProductVariant{
		ID:            row.ID,
		ProductID:     row.ProductID,
		SKU:           deref(row.SKU),
		Name:          row.VariantName,
		Color:         deref(row.Color),
		Price:         row.Price,
		StockQuantity: row.StockQuantity,
		IsActive:      row.IsActive,
	}
	return &v, nil
}
