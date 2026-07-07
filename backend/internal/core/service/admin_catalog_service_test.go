package service

import (
	"context"
	"errors"
	"testing"

	"github.com/shopspring/decimal"

	"github.com/khaopod/backend/internal/core/domain"
	"github.com/khaopod/backend/internal/core/port/input"
)

func newCatalogSvc() (*AdminCatalogService, *fakeCatalogRepo, *fakeFileStorage) {
	repo := newFakeCatalogRepo()
	files := &fakeFileStorage{}
	products := &fakeProductRepo{variants: map[uint]domain.ProductVariant{}}
	return NewAdminCatalogService(repo, products, files, fakeTx{}), repo, files
}

func productCmd(name, slug string) input.ProductCommand {
	return input.ProductCommand{
		Name: name, Slug: slug,
		BasePrice: decimal.NewFromInt(290),
		IsActive:  true,
	}
}

func TestAdminCreateProduct_NormalizesSlug(t *testing.T) {
	svc, _, _ := newCatalogSvc()

	p, err := svc.CreateProduct(context.Background(), productCmd("เสื้อยืดข่าวปด", "  Khaopod Tee 2026  "))
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if p.Slug != "khaopod-tee-2026" {
		t.Fatalf("slug should be normalized, got %q", p.Slug)
	}
	if p.ID == 0 {
		t.Fatal("id should be set after insert")
	}
}

func TestAdminCreateProduct_RejectsInvalid(t *testing.T) {
	svc, _, _ := newCatalogSvc()

	cases := []input.ProductCommand{
		productCmd("", "slug"),   // ไม่มีชื่อ
		productCmd("ชื่อ", "   "), // slug ว่าง
		{Name: "ชื่อ", Slug: "s", BasePrice: decimal.NewFromInt(-1)}, // ราคาติดลบ
	}
	for i, cmd := range cases {
		if _, err := svc.CreateProduct(context.Background(), cmd); !errors.Is(err, domain.ErrInvalidInput) {
			t.Fatalf("case %d: want ErrInvalidInput, got %v", i, err)
		}
	}
}

func TestAdminCreateVariant_Validates(t *testing.T) {
	svc, _, _ := newCatalogSvc()
	p, _ := svc.CreateProduct(context.Background(), productCmd("เสื้อ", "tee"))

	// ดี
	err := svc.CreateVariant(context.Background(), p.ID, input.VariantCommand{
		Name: "ไซซ์ M", Color: "ดำ", Price: decimal.NewFromInt(290), Stock: 10, IsActive: true,
	})
	if err != nil {
		t.Fatalf("create variant: %v", err)
	}
	// สต็อกติดลบ
	err = svc.CreateVariant(context.Background(), p.ID, input.VariantCommand{
		Name: "ไซซ์ L", Price: decimal.NewFromInt(290), Stock: -1,
	})
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("want ErrInvalidInput, got %v", err)
	}
	// สินค้าไม่มีจริง
	err = svc.CreateVariant(context.Background(), 999, input.VariantCommand{
		Name: "ไซซ์ M", Price: decimal.NewFromInt(290),
	})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestAdminAddImage_FirstImageIsPrimary(t *testing.T) {
	svc, repo, files := newCatalogSvc()
	p, _ := svc.CreateProduct(context.Background(), productCmd("เสื้อ", "tee"))

	if err := svc.AddImage(context.Background(), p.ID, []byte("img1"), ".jpg"); err != nil {
		t.Fatalf("add image: %v", err)
	}
	if err := svc.AddImage(context.Background(), p.ID, []byte("img2"), ".jpg"); err != nil {
		t.Fatalf("add image 2: %v", err)
	}
	if !repo.images[1].IsPrimary || repo.images[2].IsPrimary {
		t.Fatalf("first image should be primary: img1=%v img2=%v", repo.images[1].IsPrimary, repo.images[2].IsPrimary)
	}
	if len(files.saved) != 2 {
		t.Fatalf("files saved = %d", len(files.saved))
	}
}

func TestAdminDeleteImage_RemovesFile(t *testing.T) {
	svc, _, files := newCatalogSvc()
	p, _ := svc.CreateProduct(context.Background(), productCmd("เสื้อ", "tee"))
	_ = svc.AddImage(context.Background(), p.ID, []byte("img"), ".jpg")

	if err := svc.DeleteImage(context.Background(), 1); err != nil {
		t.Fatalf("delete image: %v", err)
	}
	if len(files.removed) != 1 {
		t.Fatalf("file should be removed, got %v", files.removed)
	}
}

func TestAdminDeleteProduct_PropagatesConflict(t *testing.T) {
	svc, repo, _ := newCatalogSvc()
	p, _ := svc.CreateProduct(context.Background(), productCmd("เสื้อ", "tee"))
	repo.deleteErr = domain.ErrConflict // จำลอง FK ขวาง (มีของอยู่ในตะกร้าลูกค้า)

	if err := svc.DeleteProduct(context.Background(), p.ID); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("want ErrConflict, got %v", err)
	}
}

func TestAdminListUsers_ClampsPaging(t *testing.T) {
	repo := newFakeUserRepo()
	repo.seed(&domain.User{Email: "a@b.com"})
	repo.seed(&domain.User{Email: "c@d.com"})
	svc := NewAdminUserService(repo)

	users, total, err := svc.ListUsers(context.Background(), -5, -1)
	if err != nil {
		t.Fatalf("list users: %v", err)
	}
	if total != 2 || len(users) != 2 {
		t.Fatalf("want 2 users, got %d/%d", len(users), total)
	}
}
