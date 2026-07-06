package service

import (
	"context"

	"github.com/khaopod/backend/internal/core/domain"
	"github.com/khaopod/backend/internal/core/port/input"
	"github.com/khaopod/backend/internal/core/port/output"
)

// CartService = use case ของตะกร้า — เช็คสต็อก (soft) + ความเป็นเจ้าของ ก่อนแก้ตะกร้า
// (การกันขายเกินจริง ๆ จะทำตอนสร้างออเดอร์ด้วย row lock ไม่ใช่ที่นี่)
type CartService struct {
	carts    output.CartRepository
	products output.ProductRepository
}

var _ input.CartUseCase = (*CartService)(nil)

func NewCartService(carts output.CartRepository, products output.ProductRepository) *CartService {
	return &CartService{carts: carts, products: products}
}

func (s *CartService) Get(ctx context.Context, userID uint) (*domain.Cart, error) {
	return s.carts.Get(ctx, userID)
}

func (s *CartService) AddItem(ctx context.Context, userID uint, in input.AddItemCommand) (*domain.Cart, error) {
	if in.Quantity <= 0 {
		return nil, domain.ErrInvalidQuantity
	}
	// ต้องเป็น variant ที่มีจริงและยังขายอยู่
	variant, err := s.products.FindVariantByID(ctx, in.VariantID)
	if err != nil {
		return nil, err
	}

	cart, err := s.carts.Get(ctx, userID)
	if err != nil {
		return nil, err
	}
	// จำนวนใหม่ = ของเดิมในตะกร้า + ที่เพิ่ง add ต้องไม่เกินสต็อก
	existing := 0
	if it, ok := cart.ItemByVariant(in.VariantID); ok {
		existing = it.Quantity
	}
	if existing+in.Quantity > variant.StockQuantity {
		return nil, domain.ErrOutOfStock
	}

	if err := s.carts.UpsertItem(ctx, userID, in.VariantID, in.Quantity); err != nil {
		return nil, err
	}
	return s.carts.Get(ctx, userID)
}

func (s *CartService) UpdateItem(ctx context.Context, userID, itemID uint, quantity int) (*domain.Cart, error) {
	if quantity <= 0 {
		return nil, domain.ErrInvalidQuantity
	}
	cart, err := s.carts.Get(ctx, userID)
	if err != nil {
		return nil, err
	}
	// item ต้องอยู่ในตะกร้าของ user เอง (กัน IDOR) — ไม่เจอ = ไม่ใช่ของเรา
	var item *domain.CartItem
	for i := range cart.Items {
		if cart.Items[i].ID == itemID {
			item = &cart.Items[i]
			break
		}
	}
	if item == nil {
		return nil, domain.ErrNotFound
	}
	if quantity > item.Stock {
		return nil, domain.ErrOutOfStock
	}

	if err := s.carts.SetItemQty(ctx, userID, itemID, quantity); err != nil {
		return nil, err
	}
	return s.carts.Get(ctx, userID)
}

func (s *CartService) RemoveItem(ctx context.Context, userID, itemID uint) (*domain.Cart, error) {
	if err := s.carts.RemoveItem(ctx, userID, itemID); err != nil {
		return nil, err
	}
	return s.carts.Get(ctx, userID)
}

func (s *CartService) Clear(ctx context.Context, userID uint) error {
	return s.carts.Clear(ctx, userID)
}
