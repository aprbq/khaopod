package domain

import (
	"testing"

	"github.com/shopspring/decimal"
)

func dec(v string) decimal.Decimal { return decimal.RequireFromString(v) }

func TestProduct_InStock(t *testing.T) {
	tests := []struct {
		name     string
		variants []ProductVariant
		want     bool
	}{
		{"no variants", nil, false},
		{"active with stock", []ProductVariant{{IsActive: true, StockQuantity: 3}}, true},
		{"active but zero stock", []ProductVariant{{IsActive: true, StockQuantity: 0}}, false},
		{"has stock but inactive", []ProductVariant{{IsActive: false, StockQuantity: 5}}, false},
		{"mixed — one active in stock", []ProductVariant{
			{IsActive: true, StockQuantity: 0},
			{IsActive: false, StockQuantity: 9},
			{IsActive: true, StockQuantity: 2},
		}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Product{Variants: tt.variants}
			if got := p.InStock(); got != tt.want {
				t.Fatalf("InStock() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProduct_PriceRange(t *testing.T) {
	t.Run("min/max from active variants only", func(t *testing.T) {
		p := Product{
			BasePrice: dec("100"),
			Variants: []ProductVariant{
				{IsActive: true, Price: dec("390")},
				{IsActive: true, Price: dec("420")},
				{IsActive: false, Price: dec("999")}, // ต้องถูกข้าม
			},
		}
		min, max := p.PriceRange()
		if !min.Equal(dec("390")) || !max.Equal(dec("420")) {
			t.Fatalf("PriceRange() = (%s, %s), want (390, 420)", min, max)
		}
	})

	t.Run("fallback to base_price when no active variants", func(t *testing.T) {
		p := Product{
			BasePrice: dec("250"),
			Variants:  []ProductVariant{{IsActive: false, Price: dec("999")}},
		}
		min, max := p.PriceRange()
		if !min.Equal(dec("250")) || !max.Equal(dec("250")) {
			t.Fatalf("PriceRange() = (%s, %s), want (250, 250)", min, max)
		}
	})
}

func TestProduct_PrimaryImage(t *testing.T) {
	t.Run("empty when no images", func(t *testing.T) {
		p := Product{}
		if got := p.PrimaryImage(); got != "" {
			t.Fatalf("PrimaryImage() = %q, want empty", got)
		}
	})

	t.Run("prefers is_primary flag", func(t *testing.T) {
		p := Product{Images: []ProductImage{
			{URL: "a.jpg", SortOrder: 0},
			{URL: "b.jpg", SortOrder: 1, IsPrimary: true},
		}}
		if got := p.PrimaryImage(); got != "b.jpg" {
			t.Fatalf("PrimaryImage() = %q, want b.jpg", got)
		}
	})

	t.Run("falls back to lowest sort_order", func(t *testing.T) {
		p := Product{Images: []ProductImage{
			{URL: "a.jpg", SortOrder: 2},
			{URL: "b.jpg", SortOrder: 1},
		}}
		if got := p.PrimaryImage(); got != "b.jpg" {
			t.Fatalf("PrimaryImage() = %q, want b.jpg", got)
		}
	})
}
