package service

import (
	"context"
	"errors"
	"testing"

	"github.com/shopspring/decimal"

	"github.com/khaopod/backend/internal/core/domain"
	"github.com/khaopod/backend/internal/core/port/input"
)

// fakeCartRepo — ตะกร้าใน memory; enrich item จาก variants ที่แชร์กับ fakeProductRepo
type fakeCartRepo struct {
	items      map[uint]map[uint]int  // userID -> variantID -> quantity
	itemIDs    map[uint]map[uint]uint // userID -> variantID -> itemID (stable)
	variants   map[uint]domain.ProductVariant
	nextItemID uint
}

func newFakeCartRepo(variants map[uint]domain.ProductVariant) *fakeCartRepo {
	return &fakeCartRepo{
		items:      map[uint]map[uint]int{},
		itemIDs:    map[uint]map[uint]uint{},
		variants:   variants,
		nextItemID: 1,
	}
}

func (r *fakeCartRepo) Get(_ context.Context, userID uint) (*domain.Cart, error) {
	c := &domain.Cart{ID: 1, UserID: userID}
	for vid, qty := range r.items[userID] {
		v := r.variants[vid]
		c.Items = append(c.Items, domain.CartItem{
			ID:          r.itemIDs[userID][vid],
			VariantID:   vid,
			ProductName: "สินค้า",
			VariantName: v.Name,
			Color:       v.Color,
			UnitPrice:   v.Price,
			Stock:       v.StockQuantity,
			Quantity:    qty,
		})
	}
	return c, nil
}

func (r *fakeCartRepo) UpsertItem(_ context.Context, userID, variantID uint, qty int) error {
	if r.items[userID] == nil {
		r.items[userID] = map[uint]int{}
		r.itemIDs[userID] = map[uint]uint{}
	}
	if _, ok := r.items[userID][variantID]; !ok {
		r.itemIDs[userID][variantID] = r.nextItemID
		r.nextItemID++
	}
	r.items[userID][variantID] += qty
	return nil
}

func (r *fakeCartRepo) SetItemQty(_ context.Context, userID, itemID uint, qty int) error {
	for vid, id := range r.itemIDs[userID] {
		if id == itemID {
			r.items[userID][vid] = qty
		}
	}
	return nil
}

func (r *fakeCartRepo) RemoveItem(_ context.Context, userID, itemID uint) error {
	for vid, id := range r.itemIDs[userID] {
		if id == itemID {
			delete(r.items[userID], vid)
			delete(r.itemIDs[userID], vid)
		}
	}
	return nil
}

func (r *fakeCartRepo) Clear(_ context.Context, userID uint) error {
	delete(r.items, userID)
	delete(r.itemIDs, userID)
	return nil
}

const cartUID = uint(1)

func cartSvcWith(stock int) (*CartService, map[uint]domain.ProductVariant) {
	variants := map[uint]domain.ProductVariant{
		100: {ID: 100, Name: "ไซซ์ M", Color: "ดำ", Price: decimal.NewFromInt(390), StockQuantity: stock, IsActive: true},
	}
	return NewCartService(newFakeCartRepo(variants), &fakeProductRepo{variants: variants}), variants
}

func TestCartService_AddItem(t *testing.T) {
	ctx := context.Background()

	t.Run("adds item", func(t *testing.T) {
		svc, _ := cartSvcWith(3)
		c, err := svc.AddItem(ctx, cartUID, input.AddItemCommand{VariantID: 100, Quantity: 2})
		if err != nil {
			t.Fatalf("unexpected: %v", err)
		}
		if c.ItemCount() != 2 || len(c.Items) != 1 {
			t.Fatalf("want 2 items in 1 row, got count=%d rows=%d", c.ItemCount(), len(c.Items))
		}
	})

	t.Run("merges quantity on duplicate variant", func(t *testing.T) {
		svc, _ := cartSvcWith(5)
		_, _ = svc.AddItem(ctx, cartUID, input.AddItemCommand{VariantID: 100, Quantity: 2})
		c, err := svc.AddItem(ctx, cartUID, input.AddItemCommand{VariantID: 100, Quantity: 1})
		if err != nil {
			t.Fatalf("unexpected: %v", err)
		}
		if c.ItemCount() != 3 || len(c.Items) != 1 {
			t.Fatalf("want merged 3 in 1 row, got count=%d rows=%d", c.ItemCount(), len(c.Items))
		}
	})

	t.Run("exceeding stock (existing + new) returns ErrOutOfStock", func(t *testing.T) {
		svc, _ := cartSvcWith(3)
		_, _ = svc.AddItem(ctx, cartUID, input.AddItemCommand{VariantID: 100, Quantity: 2})
		_, err := svc.AddItem(ctx, cartUID, input.AddItemCommand{VariantID: 100, Quantity: 2}) // 2+2 > 3
		if !errors.Is(err, domain.ErrOutOfStock) {
			t.Fatalf("want ErrOutOfStock, got %v", err)
		}
	})

	t.Run("unknown variant returns ErrNotFound", func(t *testing.T) {
		svc, _ := cartSvcWith(3)
		_, err := svc.AddItem(ctx, cartUID, input.AddItemCommand{VariantID: 999, Quantity: 1})
		if !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("want ErrNotFound, got %v", err)
		}
	})

	t.Run("non-positive quantity returns ErrInvalidQuantity", func(t *testing.T) {
		svc, _ := cartSvcWith(3)
		_, err := svc.AddItem(ctx, cartUID, input.AddItemCommand{VariantID: 100, Quantity: 0})
		if !errors.Is(err, domain.ErrInvalidQuantity) {
			t.Fatalf("want ErrInvalidQuantity, got %v", err)
		}
	})
}

func TestCartService_UpdateItem(t *testing.T) {
	ctx := context.Background()

	setup := func() (*CartService, uint) {
		svc, _ := cartSvcWith(5)
		c, _ := svc.AddItem(ctx, cartUID, input.AddItemCommand{VariantID: 100, Quantity: 1})
		return svc, c.Items[0].ID
	}

	t.Run("updates quantity", func(t *testing.T) {
		svc, itemID := setup()
		c, err := svc.UpdateItem(ctx, cartUID, itemID, 3)
		if err != nil {
			t.Fatalf("unexpected: %v", err)
		}
		if c.Items[0].Quantity != 3 {
			t.Fatalf("want 3, got %d", c.Items[0].Quantity)
		}
	})

	t.Run("over stock returns ErrOutOfStock", func(t *testing.T) {
		svc, itemID := setup()
		_, err := svc.UpdateItem(ctx, cartUID, itemID, 6) // > 5
		if !errors.Is(err, domain.ErrOutOfStock) {
			t.Fatalf("want ErrOutOfStock, got %v", err)
		}
	})

	t.Run("other user's item returns ErrNotFound (IDOR guard)", func(t *testing.T) {
		svc, itemID := setup()
		_, err := svc.UpdateItem(ctx, cartUID+99, itemID, 2)
		if !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("want ErrNotFound, got %v", err)
		}
	})

	t.Run("non-positive quantity returns ErrInvalidQuantity", func(t *testing.T) {
		svc, itemID := setup()
		_, err := svc.UpdateItem(ctx, cartUID, itemID, 0)
		if !errors.Is(err, domain.ErrInvalidQuantity) {
			t.Fatalf("want ErrInvalidQuantity, got %v", err)
		}
	})
}
