package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/khaopod/backend/internal/core/domain"
	"github.com/khaopod/backend/internal/core/port/input"
	"github.com/khaopod/backend/internal/core/port/output"
)

// AdminCatalogService = use case จัดการสินค้าในหลังบ้าน (implements input.AdminCatalogUseCase)
type AdminCatalogService struct {
	catalog  output.CatalogAdminRepository
	products output.ProductRepository
	files    output.FileStorage
	tx       output.TxManager
}

var _ input.AdminCatalogUseCase = (*AdminCatalogService)(nil)

func NewAdminCatalogService(
	catalog output.CatalogAdminRepository,
	products output.ProductRepository,
	files output.FileStorage,
	tx output.TxManager,
) *AdminCatalogService {
	return &AdminCatalogService{catalog: catalog, products: products, files: files, tx: tx}
}

func (s *AdminCatalogService) ListProducts(ctx context.Context, q input.ProductQuery) ([]domain.Product, int, error) {
	page, perPage := q.NormalizedPaging()
	col, desc := parseSort(q.Sort)
	return s.catalog.ListAllProducts(ctx, output.ProductFilter{
		CategorySlug: strings.TrimSpace(q.CategorySlug),
		Search:       strings.TrimSpace(q.Search),
		SortColumn:   col,
		SortDesc:     desc,
		Limit:        perPage,
		Offset:       (page - 1) * perPage,
	})
}

func (s *AdminCatalogService) GetProduct(ctx context.Context, id uint) (*domain.Product, error) {
	return s.catalog.FindProductByID(ctx, id)
}

func (s *AdminCatalogService) CreateProduct(ctx context.Context, cmd input.ProductCommand) (*domain.Product, error) {
	p, err := productFromCommand(&domain.Product{}, cmd)
	if err != nil {
		return nil, err
	}
	if err := s.catalog.CreateProduct(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *AdminCatalogService) UpdateProduct(ctx context.Context, id uint, cmd input.ProductCommand) (*domain.Product, error) {
	p, err := s.catalog.FindProductByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if _, err := productFromCommand(p, cmd); err != nil {
		return nil, err
	}
	if err := s.catalog.UpdateProduct(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

// DeleteProduct ลบถาวร (variants/รูป cascade ใน DB) แล้วเก็บกวาดไฟล์รูปแบบ best-effort
func (s *AdminCatalogService) DeleteProduct(ctx context.Context, id uint) error {
	p, err := s.catalog.FindProductByID(ctx, id)
	if err != nil {
		return err
	}
	if err := s.catalog.DeleteProduct(ctx, id); err != nil {
		return err
	}
	for _, img := range p.Images {
		_ = s.files.Remove(ctx, img.URL) // ไฟล์ seed อยู่นอก /uploads — storage เมินให้เอง
	}
	return nil
}

func (s *AdminCatalogService) CreateVariant(ctx context.Context, productID uint, cmd input.VariantCommand) error {
	if _, err := s.catalog.FindProductByID(ctx, productID); err != nil {
		return err
	}
	v, err := variantFromCommand(&domain.ProductVariant{ProductID: productID}, cmd)
	if err != nil {
		return err
	}
	return s.catalog.CreateVariant(ctx, v)
}

func (s *AdminCatalogService) UpdateVariant(ctx context.Context, variantID uint, cmd input.VariantCommand) error {
	v, err := s.products.GetVariantForUpdate(ctx, variantID) // นอก tx ก็แค่ SELECT ธรรมดา
	if err != nil {
		return err
	}
	if _, err := variantFromCommand(v, cmd); err != nil {
		return err
	}
	return s.catalog.UpdateVariant(ctx, v)
}

func (s *AdminCatalogService) DeleteVariant(ctx context.Context, variantID uint) error {
	return s.catalog.DeleteVariant(ctx, variantID)
}

func (s *AdminCatalogService) AddImage(ctx context.Context, productID uint, content []byte, ext string) error {
	p, err := s.catalog.FindProductByID(ctx, productID)
	if err != nil {
		return err
	}
	// ผูกลำดับรูปในชื่อไฟล์ด้วย — timestamp อย่างเดียวชนกันได้เมื่ออัปโหลดติด ๆ กัน
	// (นาฬิกา Windows หยาบระดับมิลลิวินาที)
	name := fmt.Sprintf("products/%d-%d-%d%s", productID, time.Now().UnixNano(), len(p.Images), ext)
	url, err := s.files.Save(ctx, name, content)
	if err != nil {
		return err
	}
	img := &domain.ProductImage{
		ProductID: productID,
		URL:       url,
		IsPrimary: len(p.Images) == 0, // รูปแรกของสินค้า = รูปปกอัตโนมัติ
		SortOrder: len(p.Images),
	}
	if err := s.catalog.AddImage(ctx, img); err != nil {
		_ = s.files.Remove(ctx, url) // ไม่ทิ้งไฟล์กำพร้า
		return err
	}
	return nil
}

func (s *AdminCatalogService) DeleteImage(ctx context.Context, imageID uint) error {
	url, err := s.catalog.DeleteImage(ctx, imageID)
	if err != nil {
		return err
	}
	_ = s.files.Remove(ctx, url)
	return nil
}

func (s *AdminCatalogService) SetPrimaryImage(ctx context.Context, productID, imageID uint) error {
	// ปลดปกเดิม + ตั้งปกใหม่ต้องอยู่ tx เดียว — DB มี unique partial index คุมปกได้ 1 รูป/สินค้า
	return s.tx.WithinTx(ctx, func(ctx context.Context) error {
		return s.catalog.SetPrimaryImage(ctx, productID, imageID)
	})
}

var slugAllowed = regexp.MustCompile(`[^a-z0-9ก-๙-]+`)

// normalizeSlug ทำ slug ให้อยู่ในรูปแบบ url-safe (ตัวเล็ก, เว้นวรรค→ขีด, ตัดอักขระแปลก)
func normalizeSlug(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, " ", "-")
	s = slugAllowed.ReplaceAllString(s, "")
	return strings.Trim(s, "-")
}

func productFromCommand(p *domain.Product, cmd input.ProductCommand) (*domain.Product, error) {
	name := strings.TrimSpace(cmd.Name)
	slug := normalizeSlug(cmd.Slug)
	if name == "" || slug == "" || cmd.BasePrice.IsNegative() {
		return nil, domain.ErrInvalidInput
	}
	p.Name = name
	p.Slug = slug
	p.Description = strings.TrimSpace(cmd.Description)
	p.BasePrice = cmd.BasePrice
	p.CategoryID = cmd.CategoryID
	p.IsActive = cmd.IsActive
	p.IsFeatured = cmd.IsFeatured
	return p, nil
}

func variantFromCommand(v *domain.ProductVariant, cmd input.VariantCommand) (*domain.ProductVariant, error) {
	name := strings.TrimSpace(cmd.Name)
	if name == "" || cmd.Price.IsNegative() || cmd.Stock < 0 {
		return nil, domain.ErrInvalidInput
	}
	v.Name = name
	v.Color = strings.TrimSpace(cmd.Color)
	v.SKU = strings.TrimSpace(cmd.SKU)
	v.Price = cmd.Price
	v.StockQuantity = cmd.Stock
	v.IsActive = cmd.IsActive
	return v, nil
}
