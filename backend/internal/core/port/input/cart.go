package input

import (
	"context"

	"github.com/khaopod/backend/internal/core/domain"
)

// CartUseCase — use case ของตะกร้า (inbound adapter เรียกผ่านอันนี้)
// ทุก op ที่แก้ตะกร้าคืน cart ล่าสุดกลับไปเลย (frontend ไม่ต้อง GET ซ้ำ)
type CartUseCase interface {
	Get(ctx context.Context, userID uint) (*domain.Cart, error)
	AddItem(ctx context.Context, userID uint, in AddItemCommand) (*domain.Cart, error)
	UpdateItem(ctx context.Context, userID, itemID uint, quantity int) (*domain.Cart, error)
	RemoveItem(ctx context.Context, userID, itemID uint) (*domain.Cart, error)
	Clear(ctx context.Context, userID uint) error
}

type AddItemCommand struct {
	VariantID uint
	Quantity  int
}
