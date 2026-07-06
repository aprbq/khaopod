package rest

import (
	"github.com/shopspring/decimal"

	"github.com/khaopod/backend/internal/core/domain"
	"github.com/khaopod/backend/internal/core/port/input"
)

// ---- Request DTO (แยกจาก domain entity เสมอ กัน mass-assignment) ----

type requestOTPRequest struct {
	Email string `json:"email" valid:"required,email"`
}

type googleLoginRequest struct {
	IDToken string `json:"id_token" valid:"required"`
}

type verifyOTPRequest struct {
	Email string `json:"email" valid:"required,email"`
	Code  string `json:"code" valid:"required,numeric"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type updateProfileRequest struct {
	DisplayName *string `json:"display_name"`
	Phone       *string `json:"phone"`
}

func (r updateProfileRequest) toCommand() input.UpdateProfileCommand {
	return input.UpdateProfileCommand{DisplayName: r.DisplayName, Phone: r.Phone}
}

// ---- Response DTO ----

type userResponse struct {
	PublicID    string `json:"public_id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url"`
	Phone       string `json:"phone,omitempty"`
	Role        string `json:"role"`
}

func toUserResponse(u *domain.User) userResponse {
	return userResponse{
		PublicID:    u.PublicID,
		Email:       u.Email,
		DisplayName: u.DisplayName,
		AvatarURL:   u.AvatarURL,
		Phone:       u.Phone,
		Role:        string(u.Role),
	}
}

type authResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	TokenType    string       `json:"token_type"`
	ExpiresIn    int          `json:"expires_in"`
	User         userResponse `json:"user"`
}

func toAuthResponse(r *input.AuthResult) authResponse {
	return authResponse{
		AccessToken:  r.AccessToken,
		RefreshToken: r.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    r.ExpiresIn,
		User:         toUserResponse(r.User),
	}
}

// ---- Product response DTO (ตรงกับ docs/rest_api.md §5) ----

type priceRange struct {
	Min decimal.Decimal `json:"min"`
	Max decimal.Decimal `json:"max"`
}

type categoryResponse struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

func toCategoryResponse(c *domain.Category) *categoryResponse {
	if c == nil {
		return nil
	}
	return &categoryResponse{ID: c.ID, Name: c.Name, Slug: c.Slug}
}

// รายการสินค้า (§5.1) — สรุปข้อมูลย่อ + ช่วงราคา/สถานะสต็อกที่คำนวณจาก variants
type productListItem struct {
	ID           uint              `json:"id"`
	Name         string            `json:"name"`
	Slug         string            `json:"slug"`
	BasePrice    decimal.Decimal   `json:"base_price"`
	IsFeatured   bool              `json:"is_featured"`
	PrimaryImage *string           `json:"primary_image"`
	PriceRange   priceRange        `json:"price_range"`
	InStock      bool              `json:"in_stock"`
	Category     *categoryResponse `json:"category"`
}

func toProductListItem(p domain.Product) productListItem {
	min, max := p.PriceRange()
	return productListItem{
		ID:           p.ID,
		Name:         p.Name,
		Slug:         p.Slug,
		BasePrice:    p.BasePrice,
		IsFeatured:   p.IsFeatured,
		PrimaryImage: nilIfEmptyStr(p.PrimaryImage()),
		PriceRange:   priceRange{Min: min, Max: max},
		InStock:      p.InStock(),
		Category:     toCategoryResponse(p.Category),
	}
}

type productImageResponse struct {
	ID        uint   `json:"id"`
	URL       string `json:"url"`
	IsPrimary bool   `json:"is_primary"`
	SortOrder int    `json:"sort_order"`
}

type productVariantResponse struct {
	ID            uint            `json:"id"`
	VariantName   string          `json:"variant_name"`
	Color         string          `json:"color,omitempty"`
	Price         decimal.Decimal `json:"price"`
	StockQuantity int             `json:"stock_quantity"`
	SKU           string          `json:"sku,omitempty"`
}

// รายละเอียดสินค้า (§5.2) = ข้อมูลย่อชุดเดียวกับ list + description + variants + รูปทั้งหมด
// embed productListItem เพื่อให้ฟิลด์สรุป (price_range, in_stock, category ฯลฯ) ตรงกันทั้งสอง endpoint
type productDetailResponse struct {
	productListItem
	Description string                   `json:"description"`
	Images      []productImageResponse   `json:"images"`
	Variants    []productVariantResponse `json:"variants"`
}

func toProductDetailResponse(p *domain.Product) productDetailResponse {
	resp := productDetailResponse{
		productListItem: toProductListItem(*p),
		Description:     p.Description,
		Images:          make([]productImageResponse, 0, len(p.Images)),
		Variants:        make([]productVariantResponse, 0, len(p.Variants)),
	}
	for _, img := range p.Images {
		resp.Images = append(resp.Images, productImageResponse{
			ID:        img.ID,
			URL:       img.URL,
			IsPrimary: img.IsPrimary,
			SortOrder: img.SortOrder,
		})
	}
	for _, v := range p.Variants {
		resp.Variants = append(resp.Variants, productVariantResponse{
			ID:            v.ID,
			VariantName:   v.Name,
			Color:         v.Color,
			Price:         v.Price,
			StockQuantity: v.StockQuantity,
			SKU:           v.SKU,
		})
	}
	return resp
}

func nilIfEmptyStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// ---- Cart (§6) ----

type addCartItemRequest struct {
	VariantID uint `json:"variant_id" valid:"required"`
	Quantity  int  `json:"quantity" valid:"required"`
}

func (r addCartItemRequest) toCommand() input.AddItemCommand {
	return input.AddItemCommand{VariantID: r.VariantID, Quantity: r.Quantity}
}

type updateCartItemRequest struct {
	Quantity int `json:"quantity" valid:"required"`
}

type cartItemResponse struct {
	ID          uint            `json:"id"`
	VariantID   uint            `json:"variant_id"`
	ProductName string          `json:"product_name"`
	VariantName string          `json:"variant_name"`
	Color       string          `json:"color,omitempty"`
	UnitPrice   decimal.Decimal `json:"unit_price"`
	Quantity    int             `json:"quantity"`
	LineTotal   decimal.Decimal `json:"line_total"`
	Image       string          `json:"image,omitempty"`
	InStock     bool            `json:"in_stock"`
}

type cartResponse struct {
	ID        uint               `json:"id"`
	Items     []cartItemResponse `json:"items"`
	Subtotal  decimal.Decimal    `json:"subtotal"`
	ItemCount int                `json:"item_count"`
}

func toCartResponse(c *domain.Cart) cartResponse {
	items := make([]cartItemResponse, 0, len(c.Items))
	for _, it := range c.Items {
		items = append(items, cartItemResponse{
			ID:          it.ID,
			VariantID:   it.VariantID,
			ProductName: it.ProductName,
			VariantName: it.VariantName,
			Color:       it.Color,
			UnitPrice:   it.UnitPrice,
			Quantity:    it.Quantity,
			LineTotal:   it.LineTotal(),
			Image:       it.Image,
			InStock:     it.InStock(),
		})
	}
	return cartResponse{ID: c.ID, Items: items, Subtotal: c.Subtotal(), ItemCount: c.ItemCount()}
}
