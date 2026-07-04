package postgres

import (
	"time"

	"github.com/shopspring/decimal"

	"github.com/khaopod/backend/internal/core/domain"
)

// persistence model ของแคตตาล็อก — gorm tag อยู่ที่นี่เท่านั้น (ไม่รั่วเข้า domain)

type categoryRow struct {
	ID       uint   `gorm:"column:id;primaryKey"`
	ParentID *uint  `gorm:"column:parent_id"`
	Name     string `gorm:"column:name"`
	Slug     string `gorm:"column:slug"`
	IsActive bool   `gorm:"column:is_active"`
}

func (categoryRow) TableName() string { return "categories" }

type productRow struct {
	ID          uint            `gorm:"column:id;primaryKey"`
	CategoryID  *uint           `gorm:"column:category_id"`
	Name        string          `gorm:"column:name"`
	Slug        string          `gorm:"column:slug"`
	Description *string         `gorm:"column:description"`
	BasePrice   decimal.Decimal `gorm:"column:base_price"`
	IsActive    bool            `gorm:"column:is_active"`
	IsFeatured  bool            `gorm:"column:is_featured"`
	CreatedAt   time.Time       `gorm:"column:created_at"`
	UpdatedAt   time.Time       `gorm:"column:updated_at"`

	// relations (preload)
	Category *categoryRow `gorm:"foreignKey:CategoryID;references:ID"`
	Variants []variantRow `gorm:"foreignKey:ProductID;references:ID"`
	Images   []imageRow   `gorm:"foreignKey:ProductID;references:ID"`
}

func (productRow) TableName() string { return "products" }

type variantRow struct {
	ID            uint            `gorm:"column:id;primaryKey"`
	ProductID     uint            `gorm:"column:product_id"`
	SKU           *string         `gorm:"column:sku"`
	VariantName   string          `gorm:"column:variant_name"`
	Price         decimal.Decimal `gorm:"column:price"`
	StockQuantity int             `gorm:"column:stock_quantity"`
	IsActive      bool            `gorm:"column:is_active"`
}

func (variantRow) TableName() string { return "product_variants" }

type imageRow struct {
	ID        uint    `gorm:"column:id;primaryKey"`
	ProductID uint    `gorm:"column:product_id"`
	URL       string  `gorm:"column:url"`
	AltText   *string `gorm:"column:alt_text"`
	IsPrimary bool    `gorm:"column:is_primary"`
	SortOrder int     `gorm:"column:sort_order"`
}

func (imageRow) TableName() string { return "product_images" }

func toProductDomain(r productRow) domain.Product {
	p := domain.Product{
		ID:          r.ID,
		CategoryID:  r.CategoryID,
		Name:        r.Name,
		Slug:        r.Slug,
		Description: deref(r.Description),
		BasePrice:   r.BasePrice,
		IsActive:    r.IsActive,
		IsFeatured:  r.IsFeatured,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
	if r.Category != nil {
		p.Category = &domain.Category{
			ID:       r.Category.ID,
			ParentID: r.Category.ParentID,
			Name:     r.Category.Name,
			Slug:     r.Category.Slug,
		}
	}
	for _, v := range r.Variants {
		p.Variants = append(p.Variants, domain.ProductVariant{
			ID:            v.ID,
			ProductID:     v.ProductID,
			SKU:           deref(v.SKU),
			Name:          v.VariantName,
			Price:         v.Price,
			StockQuantity: v.StockQuantity,
			IsActive:      v.IsActive,
		})
	}
	for _, img := range r.Images {
		p.Images = append(p.Images, domain.ProductImage{
			ID:        img.ID,
			ProductID: img.ProductID,
			URL:       img.URL,
			AltText:   deref(img.AltText),
			IsPrimary: img.IsPrimary,
			SortOrder: img.SortOrder,
		})
	}
	return p
}
