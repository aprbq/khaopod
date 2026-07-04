package service

import (
	"context"
	"errors"
	"testing"

	"github.com/khaopod/backend/internal/core/domain"
	"github.com/khaopod/backend/internal/core/port/input"
)

func TestProductService_List_NormalizesPaging(t *testing.T) {
	tests := []struct {
		name       string
		page       int
		perPage    int
		wantLimit  int
		wantOffset int
	}{
		{"defaults when zero", 0, 0, input.DefaultPerPage, 0},
		{"page 2 default size", 2, 0, input.DefaultPerPage, input.DefaultPerPage},
		{"per_page clamped to max", 1, 9999, input.MaxPerPage, 0},
		{"negative page treated as first", -5, 10, 10, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeProductRepo{}
			svc := NewProductService(repo)
			if _, _, err := svc.List(context.Background(), input.ProductQuery{Page: tt.page, PerPage: tt.perPage}); err != nil {
				t.Fatalf("List: %v", err)
			}
			if repo.lastF.Limit != tt.wantLimit {
				t.Fatalf("Limit = %d, want %d", repo.lastF.Limit, tt.wantLimit)
			}
			if repo.lastF.Offset != tt.wantOffset {
				t.Fatalf("Offset = %d, want %d", repo.lastF.Offset, tt.wantOffset)
			}
		})
	}
}

func TestProductService_List_SortWhitelist(t *testing.T) {
	tests := []struct {
		name     string
		sort     string
		wantCol  string
		wantDesc bool
	}{
		{"empty defaults to newest first", "", "created_at", true},
		{"unknown column falls back to default", "-price; DROP TABLE products", "created_at", true},
		{"ascending name", "name", "name", false},
		{"descending base_price", "-base_price", "base_price", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeProductRepo{}
			svc := NewProductService(repo)
			if _, _, err := svc.List(context.Background(), input.ProductQuery{Sort: tt.sort}); err != nil {
				t.Fatalf("List: %v", err)
			}
			if repo.lastF.SortColumn != tt.wantCol || repo.lastF.SortDesc != tt.wantDesc {
				t.Fatalf("sort = (%q, desc=%v), want (%q, desc=%v)",
					repo.lastF.SortColumn, repo.lastF.SortDesc, tt.wantCol, tt.wantDesc)
			}
		})
	}
}

func TestProductService_GetBySlug(t *testing.T) {
	repo := &fakeProductRepo{items: []domain.Product{{Slug: "kbc-tee-hq", Name: "เสื้อยืด"}}}
	svc := NewProductService(repo)

	t.Run("found", func(t *testing.T) {
		p, err := svc.GetBySlug(context.Background(), "kbc-tee-hq")
		if err != nil {
			t.Fatalf("GetBySlug: %v", err)
		}
		if p.Name != "เสื้อยืด" {
			t.Fatalf("got %q", p.Name)
		}
	})

	t.Run("blank slug is not found (no repo call)", func(t *testing.T) {
		_, err := svc.GetBySlug(context.Background(), "   ")
		if !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("want ErrNotFound, got %v", err)
		}
	})

	t.Run("missing slug", func(t *testing.T) {
		_, err := svc.GetBySlug(context.Background(), "nope")
		if !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("want ErrNotFound, got %v", err)
		}
	})
}
