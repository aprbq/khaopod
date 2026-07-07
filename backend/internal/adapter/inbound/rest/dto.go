package rest

import (
	"time"

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
	PublicID    string    `json:"public_id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	AvatarURL   string    `json:"avatar_url"`
	Phone       string    `json:"phone,omitempty"`
	Role        string    `json:"role"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func toUserResponse(u *domain.User) userResponse {
	return userResponse{
		PublicID:    u.PublicID,
		Email:       u.Email,
		DisplayName: u.DisplayName,
		AvatarURL:   u.AvatarURL,
		Phone:       u.Phone,
		Role:        string(u.Role),
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
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
	IsActive      bool            `json:"is_active"`
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
			IsActive:      v.IsActive,
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

// ---- Addresses (§7) ----

type addressRequest struct {
	RecipientName string `json:"recipient_name" valid:"required"`
	Phone         string `json:"phone" valid:"required,numeric"`
	AddressLine   string `json:"address_line" valid:"required"`
	Subdistrict   string `json:"subdistrict" valid:"required"`
	District      string `json:"district" valid:"required"`
	Province      string `json:"province" valid:"required"`
	PostalCode    string `json:"postal_code" valid:"required,matches(^[0-9]{5}$)"`
	Note          string `json:"note"`
	IsDefault     bool   `json:"is_default"`
}

func (r addressRequest) toCommand() input.AddressCommand {
	return input.AddressCommand{
		RecipientName: r.RecipientName,
		Phone:         r.Phone,
		AddressLine:   r.AddressLine,
		Subdistrict:   r.Subdistrict,
		District:      r.District,
		Province:      r.Province,
		PostalCode:    r.PostalCode,
		Note:          r.Note,
		IsDefault:     r.IsDefault,
	}
}

type addressResponse struct {
	ID            uint   `json:"id"`
	RecipientName string `json:"recipient_name"`
	Phone         string `json:"phone"`
	AddressLine   string `json:"address_line"`
	Subdistrict   string `json:"subdistrict"`
	District      string `json:"district"`
	Province      string `json:"province"`
	PostalCode    string `json:"postal_code"`
	Note          string `json:"note,omitempty"`
	IsDefault     bool   `json:"is_default"`
}

func toAddressResponse(a *domain.Address) addressResponse {
	return addressResponse{
		ID:            a.ID,
		RecipientName: a.RecipientName,
		Phone:         a.Phone,
		AddressLine:   a.AddressLine,
		Subdistrict:   a.Subdistrict,
		District:      a.District,
		Province:      a.Province,
		PostalCode:    a.PostalCode,
		Note:          a.Note,
		IsDefault:     a.IsDefault,
	}
}

// ---- Orders (§9) + Payments (§10) ----

type placeOrderRequest struct {
	AddressID     uint   `json:"address_id" valid:"required"`
	PaymentMethod string `json:"payment_method" valid:"required"`
	CustomerNote  string `json:"customer_note"`
}

func (r placeOrderRequest) toCommand() input.PlaceOrderCommand {
	return input.PlaceOrderCommand{
		AddressID:     r.AddressID,
		PaymentMethod: domain.PaymentMethod(r.PaymentMethod),
		CustomerNote:  r.CustomerNote,
	}
}

type cancelOrderRequest struct {
	Reason string `json:"reason"`
}

type orderItemResponse struct {
	ProductName string          `json:"product_name"`
	VariantName string          `json:"variant_name"`
	UnitPrice   decimal.Decimal `json:"unit_price"`
	Quantity    int             `json:"quantity"`
	LineTotal   decimal.Decimal `json:"line_total"`
}

type shippingAddressResponse struct {
	Recipient   string `json:"recipient"`
	Phone       string `json:"phone"`
	Address     string `json:"address"`
	Subdistrict string `json:"subdistrict"`
	District    string `json:"district"`
	Province    string `json:"province"`
	PostalCode  string `json:"postal_code"`
}

type paymentResponse struct {
	ID             uint            `json:"id"` // แอดมินใช้อ้างตอน PATCH /admin/payments/{id}/verify
	Method         string          `json:"method"`
	Amount         decimal.Decimal `json:"amount"`
	Status         string          `json:"status"`
	SlipURL        string          `json:"slip_url,omitempty"`
	TransactionRef string          `json:"transaction_ref,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
}

func toPaymentResponse(p *domain.Payment) *paymentResponse {
	if p == nil {
		return nil
	}
	return &paymentResponse{
		ID:             p.ID,
		Method:         string(p.Method),
		Amount:         p.Amount,
		Status:         string(p.Status),
		SlipURL:        p.SlipURL,
		TransactionRef: p.TransactionRef,
		CreatedAt:      p.CreatedAt,
	}
}

type orderResponse struct {
	OrderNumber     string                  `json:"order_number"`
	Status          string                  `json:"status"`
	PaymentStatus   string                  `json:"payment_status"`
	PaymentMethod   string                  `json:"payment_method,omitempty"`
	Subtotal        decimal.Decimal         `json:"subtotal"`
	DiscountAmount  decimal.Decimal         `json:"discount_amount"`
	ShippingFee     decimal.Decimal         `json:"shipping_fee"`
	TotalAmount     decimal.Decimal         `json:"total_amount"`
	Items           []orderItemResponse     `json:"items"`
	ShippingAddress shippingAddressResponse `json:"shipping_address"`
	CustomerNote    string                  `json:"customer_note,omitempty"`
	Payment         *paymentResponse        `json:"payment,omitempty"`
	PlacedAt        time.Time               `json:"placed_at"`
}

func toOrderResponse(o *domain.Order) orderResponse {
	items := make([]orderItemResponse, 0, len(o.Items))
	for _, it := range o.Items {
		items = append(items, orderItemResponse{
			ProductName: it.ProductName,
			VariantName: it.VariantName,
			UnitPrice:   it.UnitPrice,
			Quantity:    it.Quantity,
			LineTotal:   it.LineTotal,
		})
	}
	return orderResponse{
		OrderNumber:    o.OrderNumber,
		Status:         string(o.Status),
		PaymentStatus:  string(o.PaymentStatus),
		PaymentMethod:  string(o.PaymentMethod),
		Subtotal:       o.Subtotal,
		DiscountAmount: o.DiscountAmount,
		ShippingFee:    o.ShippingFee,
		TotalAmount:    o.TotalAmount,
		Items:          items,
		ShippingAddress: shippingAddressResponse{
			Recipient:   o.Shipping.Recipient,
			Phone:       o.Shipping.Phone,
			Address:     o.Shipping.Address,
			Subdistrict: o.Shipping.Subdistrict,
			District:    o.Shipping.District,
			Province:    o.Shipping.Province,
			PostalCode:  o.Shipping.PostalCode,
		},
		CustomerNote: o.CustomerNote,
		Payment:      toPaymentResponse(o.Payment),
		PlacedAt:     o.PlacedAt,
	}
}

// ---- Admin (§11) ----

type adminSummaryResponse struct {
	OrdersTotal           int             `json:"orders_total"`
	OrdersPending         int             `json:"orders_pending"`
	PaymentsPendingReview int             `json:"payments_pending_review"`
	RevenuePaid           decimal.Decimal `json:"revenue_paid"`
}

// adminOrderResponse = orderResponse + ข้อมูลเจ้าของออเดอร์ (โชว์เฉพาะหลังบ้าน)
type adminOrderResponse struct {
	orderResponse
	UserEmail string `json:"user_email"`
}

func toAdminOrderResponse(o *domain.Order) adminOrderResponse {
	return adminOrderResponse{orderResponse: toOrderResponse(o), UserEmail: o.UserEmail}
}

type updateOrderStatusRequest struct {
	Status string `json:"status" valid:"required"`
	Note   string `json:"note"`
}

type verifyPaymentRequest struct {
	Status string `json:"status" valid:"required"` // "paid" | "failed"
}

// ---- Admin: catalog (§5.3) ----

type productRequest struct {
	Name        string          `json:"name" valid:"required"`
	Slug        string          `json:"slug" valid:"required"`
	Description string          `json:"description"`
	BasePrice   decimal.Decimal `json:"base_price"`
	CategoryID  *uint           `json:"category_id"`
	IsActive    bool            `json:"is_active"`
	IsFeatured  bool            `json:"is_featured"`
}

func (r productRequest) toCommand() input.ProductCommand {
	return input.ProductCommand{
		Name:        r.Name,
		Slug:        r.Slug,
		Description: r.Description,
		BasePrice:   r.BasePrice,
		CategoryID:  r.CategoryID,
		IsActive:    r.IsActive,
		IsFeatured:  r.IsFeatured,
	}
}

type variantRequest struct {
	VariantName string          `json:"variant_name" valid:"required"`
	Color       string          `json:"color"`
	SKU         string          `json:"sku"`
	Price       decimal.Decimal `json:"price"`
	Stock       int             `json:"stock_quantity"`
	IsActive    bool            `json:"is_active"`
}

func (r variantRequest) toCommand() input.VariantCommand {
	return input.VariantCommand{
		Name:     r.VariantName,
		Color:    r.Color,
		SKU:      r.SKU,
		Price:    r.Price,
		Stock:    r.Stock,
		IsActive: r.IsActive,
	}
}

// adminProductResponse = รายละเอียดสินค้า + สถานะที่หน้าร้านไม่โชว์ (is_active)
type adminProductResponse struct {
	productDetailResponse
	IsActive bool `json:"is_active"`
}

func toAdminProductResponse(p *domain.Product) adminProductResponse {
	return adminProductResponse{productDetailResponse: toProductDetailResponse(p), IsActive: p.IsActive}
}

// ---- Admin: users ----

type adminUserResponse struct {
	PublicID    string     `json:"public_id"`
	Email       string     `json:"email"`
	DisplayName string     `json:"display_name"`
	AvatarURL   string     `json:"avatar_url"`
	Phone       string     `json:"phone,omitempty"`
	Role        string     `json:"role"`
	IsActive    bool       `json:"is_active"`
	LastLoginAt *time.Time `json:"last_login_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

func toAdminUserResponse(u *domain.User) adminUserResponse {
	return adminUserResponse{
		PublicID:    u.PublicID,
		Email:       u.Email,
		DisplayName: u.DisplayName,
		AvatarURL:   u.AvatarURL,
		Phone:       u.Phone,
		Role:        string(u.Role),
		IsActive:    u.IsActive,
		LastLoginAt: u.LastLoginAt,
		CreatedAt:   u.CreatedAt,
	}
}
