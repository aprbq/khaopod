package postgres

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"

	"github.com/khaopod/backend/internal/core/domain"
	"github.com/khaopod/backend/internal/core/port/output"
)

// ProductRepo implement CatalogAdminRepository ด้วย (เขียนแคตตาล็อกจากหลังบ้าน)
var _ output.CatalogAdminRepository = (*ProductRepo)(nil)

func (r *ProductRepo) ListCategories(ctx context.Context) ([]domain.Category, error) {
	var rows []categoryRow
	err := dbFromContext(ctx, r.db).
		Where("is_active = ?", true).
		Order("sort_order, id").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]domain.Category, 0, len(rows))
	for _, row := range rows {
		out = append(out, domain.Category{ID: row.ID, ParentID: row.ParentID, Name: row.Name, Slug: row.Slug})
	}
	return out, nil
}

// ListAllProducts เหมือน List ฝั่งหน้าร้าน แต่ไม่กรอง is_active (หลังบ้านต้องเห็นสินค้าที่ปิดขาย)
func (r *ProductRepo) ListAllProducts(ctx context.Context, f output.ProductFilter) ([]domain.Product, int, error) {
	base := dbFromContext(ctx, r.db).Model(&productRow{})
	if f.CategorySlug != "" {
		base = base.Where("category_id IN (SELECT id FROM categories WHERE slug = ?)", f.CategorySlug)
	}
	if f.Search != "" {
		base = base.Where("name ILIKE ?", "%"+f.Search+"%")
	}

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	order := f.SortColumn
	if f.SortDesc {
		order += " DESC"
	}
	var rows []productRow
	err := base.Preload("Category").Preload("Variants").Preload("Images").
		Order(order).Limit(f.Limit).Offset(f.Offset).
		Find(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	out := make([]domain.Product, 0, len(rows))
	for _, row := range rows {
		out = append(out, toProductDomain(row))
	}
	return out, int(total), nil
}

func (r *ProductRepo) FindProductByID(ctx context.Context, id uint) (*domain.Product, error) {
	var row productRow
	err := dbFromContext(ctx, r.db).
		Preload("Category").Preload("Variants").Preload("Images").
		First(&row, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	p := toProductDomain(row)
	return &p, nil
}

func (r *ProductRepo) CreateProduct(ctx context.Context, p *domain.Product) error {
	row := productWriteRow(p)
	if err := dbFromContext(ctx, r.db).Create(&row).Error; err != nil {
		return translateCatalogErr(err)
	}
	p.ID = row.ID
	p.CreatedAt = row.CreatedAt
	p.UpdatedAt = row.UpdatedAt
	return nil
}

func (r *ProductRepo) UpdateProduct(ctx context.Context, p *domain.Product) error {
	row := productWriteRow(p)
	// Select ระบุคอลัมน์เพื่อให้ false/ค่าว่างถูกเขียนด้วย และไม่แตะ relation
	err := dbFromContext(ctx, r.db).Model(&productRow{}).
		Where("id = ?", p.ID).
		Select("name", "slug", "description", "base_price", "category_id", "is_active", "is_featured").
		Updates(map[string]any{
			"name": row.Name, "slug": row.Slug, "description": row.Description,
			"base_price": row.BasePrice, "category_id": row.CategoryID,
			"is_active": row.IsActive, "is_featured": row.IsFeatured,
		}).Error
	return translateCatalogErr(err)
}

func (r *ProductRepo) DeleteProduct(ctx context.Context, id uint) error {
	// variants/รูป cascade ใน DB; ถ้า variant ค้างอยู่ในตะกร้าใคร FK (RESTRICT) จะขวาง → ErrConflict
	return translateCatalogErr(dbFromContext(ctx, r.db).Delete(&productRow{}, id).Error)
}

func (r *ProductRepo) CreateVariant(ctx context.Context, v *domain.ProductVariant) error {
	row := variantWriteRow(v)
	if err := dbFromContext(ctx, r.db).Create(&row).Error; err != nil {
		return translateCatalogErr(err)
	}
	v.ID = row.ID
	return nil
}

func (r *ProductRepo) UpdateVariant(ctx context.Context, v *domain.ProductVariant) error {
	row := variantWriteRow(v)
	err := dbFromContext(ctx, r.db).Model(&variantRow{}).
		Where("id = ?", v.ID).
		Updates(map[string]any{
			"variant_name": row.VariantName, "color": row.Color, "sku": row.SKU,
			"price": row.Price, "stock_quantity": row.StockQuantity, "is_active": row.IsActive,
		}).Error
	return translateCatalogErr(err)
}

func (r *ProductRepo) DeleteVariant(ctx context.Context, id uint) error {
	return translateCatalogErr(dbFromContext(ctx, r.db).Delete(&variantRow{}, id).Error)
}

func (r *ProductRepo) AddImage(ctx context.Context, img *domain.ProductImage) error {
	row := imageRow{
		ProductID: img.ProductID,
		URL:       img.URL,
		AltText:   nilIfEmpty(img.AltText),
		IsPrimary: img.IsPrimary,
		SortOrder: img.SortOrder,
	}
	if err := dbFromContext(ctx, r.db).Create(&row).Error; err != nil {
		return translateCatalogErr(err)
	}
	img.ID = row.ID
	return nil
}

func (r *ProductRepo) DeleteImage(ctx context.Context, id uint) (string, error) {
	db := dbFromContext(ctx, r.db)
	var row imageRow
	if err := db.First(&row, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", domain.ErrNotFound
		}
		return "", err
	}
	if err := db.Delete(&imageRow{}, id).Error; err != nil {
		return "", err
	}
	return row.URL, nil
}

func (r *ProductRepo) SetPrimaryImage(ctx context.Context, productID, imageID uint) error {
	db := dbFromContext(ctx, r.db)
	// ปลดปกเดิมก่อน (unique partial index คุมปกได้ 1 รูป/สินค้า) — caller ครอบ tx ให้แล้ว
	if err := db.Model(&imageRow{}).Where("product_id = ?", productID).Update("is_primary", false).Error; err != nil {
		return err
	}
	res := db.Model(&imageRow{}).
		Where("id = ? AND product_id = ?", imageID, productID).
		Update("is_primary", true)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func productWriteRow(p *domain.Product) productRow {
	return productRow{
		ID:          p.ID,
		CategoryID:  p.CategoryID,
		Name:        p.Name,
		Slug:        p.Slug,
		Description: nilIfEmpty(p.Description),
		BasePrice:   p.BasePrice,
		IsActive:    p.IsActive,
		IsFeatured:  p.IsFeatured,
	}
}

func variantWriteRow(v *domain.ProductVariant) variantRow {
	return variantRow{
		ID:            v.ID,
		ProductID:     v.ProductID,
		SKU:           nilIfEmpty(v.SKU),
		VariantName:   v.Name,
		Color:         nilIfEmpty(v.Color),
		Price:         v.Price,
		StockQuantity: v.StockQuantity,
		IsActive:      v.IsActive,
	}
}

// translateCatalogErr แปลง error ระดับ DB เป็น domain error
// 23503 = foreign key (เช่น variant อยู่ในตะกร้าลูกค้า), 23505 = unique (slug ซ้ำ)
func translateCatalogErr(err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	if strings.Contains(msg, "23503") || strings.Contains(msg, "23505") {
		return domain.ErrConflict
	}
	return err
}
