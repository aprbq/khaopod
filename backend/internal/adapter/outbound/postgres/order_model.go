package postgres

import (
	"time"

	"github.com/shopspring/decimal"

	"github.com/khaopod/backend/internal/core/domain"
)

// persistence model ของ addresses / orders / order_items / order_status_history / payments
// gorm tag อยู่ที่นี่เท่านั้น — domain entity เป็น struct ล้วน

type addressRow struct {
	ID            uint      `gorm:"column:id;primaryKey"`
	UserID        uint      `gorm:"column:user_id"`
	RecipientName string    `gorm:"column:recipient_name"`
	Phone         string    `gorm:"column:phone"`
	AddressLine   string    `gorm:"column:address_line"`
	Subdistrict   string    `gorm:"column:subdistrict"`
	District      string    `gorm:"column:district"`
	Province      string    `gorm:"column:province"`
	PostalCode    string    `gorm:"column:postal_code"`
	Country       string    `gorm:"column:country"`
	Note          *string   `gorm:"column:note"`
	IsDefault     bool      `gorm:"column:is_default"`
	CreatedAt     time.Time `gorm:"column:created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at"`
}

func (addressRow) TableName() string { return "addresses" }

func toAddressDomain(r addressRow) domain.Address {
	return domain.Address{
		ID:            r.ID,
		UserID:        r.UserID,
		RecipientName: r.RecipientName,
		Phone:         r.Phone,
		AddressLine:   r.AddressLine,
		Subdistrict:   r.Subdistrict,
		District:      r.District,
		Province:      r.Province,
		PostalCode:    r.PostalCode,
		Country:       r.Country,
		Note:          deref(r.Note),
		IsDefault:     r.IsDefault,
		CreatedAt:     r.CreatedAt,
		UpdatedAt:     r.UpdatedAt,
	}
}

func toAddressRow(a *domain.Address) addressRow {
	return addressRow{
		ID:            a.ID,
		UserID:        a.UserID,
		RecipientName: a.RecipientName,
		Phone:         a.Phone,
		AddressLine:   a.AddressLine,
		Subdistrict:   a.Subdistrict,
		District:      a.District,
		Province:      a.Province,
		PostalCode:    a.PostalCode,
		Country:       a.Country,
		Note:          nilIfEmpty(a.Note),
		IsDefault:     a.IsDefault,
	}
}

type orderRow struct {
	ID              uint            `gorm:"column:id;primaryKey"`
	OrderNumber     string          `gorm:"column:order_number;->"` // DB gen ผ่าน DEFAULT — ห้ามเขียนทับ
	UserID          uint            `gorm:"column:user_id"`
	Subtotal        decimal.Decimal `gorm:"column:subtotal"`
	ShippingFee     decimal.Decimal `gorm:"column:shipping_fee"`
	DiscountAmount  decimal.Decimal `gorm:"column:discount_amount"`
	TotalAmount     decimal.Decimal `gorm:"column:total_amount"`
	Status          string          `gorm:"column:status"`
	PaymentStatus   string          `gorm:"column:payment_status"`
	PaymentMethod   *string         `gorm:"column:payment_method"`
	ShipRecipient   string          `gorm:"column:ship_recipient"`
	ShipPhone       string          `gorm:"column:ship_phone"`
	ShipAddress     string          `gorm:"column:ship_address"`
	ShipSubdistrict string          `gorm:"column:ship_subdistrict"`
	ShipDistrict    string          `gorm:"column:ship_district"`
	ShipProvince    string          `gorm:"column:ship_province"`
	ShipPostalCode  string          `gorm:"column:ship_postal_code"`
	ShipCountry     string          `gorm:"column:ship_country"`
	CustomerNote    *string         `gorm:"column:customer_note"`
	PlacedAt        time.Time       `gorm:"column:placed_at;->"`
	CreatedAt       time.Time       `gorm:"column:created_at;->"`
	UpdatedAt       time.Time       `gorm:"column:updated_at;->"`
}

func (orderRow) TableName() string { return "orders" }

type orderItemRow struct {
	ID               uint            `gorm:"column:id;primaryKey"`
	OrderID          uint            `gorm:"column:order_id"`
	ProductVariantID *uint           `gorm:"column:product_variant_id"`
	ProductName      string          `gorm:"column:product_name"`
	VariantName      string          `gorm:"column:variant_name"`
	UnitPrice        decimal.Decimal `gorm:"column:unit_price"`
	Quantity         int             `gorm:"column:quantity"`
	LineTotal        decimal.Decimal `gorm:"column:line_total"`
}

func (orderItemRow) TableName() string { return "order_items" }

type statusHistoryRow struct {
	ID        uint      `gorm:"column:id;primaryKey"`
	OrderID   uint      `gorm:"column:order_id"`
	Status    string    `gorm:"column:status"`
	Note      *string   `gorm:"column:note"`
	ChangedBy *uint     `gorm:"column:changed_by"` // แอดมินที่แก้ (NULL = ระบบ/ลูกค้า)
	CreatedAt time.Time `gorm:"column:created_at;->"`
}

func (statusHistoryRow) TableName() string { return "order_status_history" }

type paymentRow struct {
	ID             uint            `gorm:"column:id;primaryKey"`
	OrderID        uint            `gorm:"column:order_id"`
	Method         string          `gorm:"column:method"`
	Amount         decimal.Decimal `gorm:"column:amount"`
	Status         string          `gorm:"column:status"`
	SlipURL        *string         `gorm:"column:slip_url"`
	TransactionRef *string         `gorm:"column:transaction_ref"`
	PaidAt         *time.Time      `gorm:"column:paid_at"`
	CreatedAt      time.Time       `gorm:"column:created_at;->"`
}

func (paymentRow) TableName() string { return "payments" }

func toOrderDomain(r orderRow) domain.Order {
	method := domain.PaymentMethod("")
	if r.PaymentMethod != nil {
		method = domain.PaymentMethod(*r.PaymentMethod)
	}
	return domain.Order{
		ID:             r.ID,
		OrderNumber:    r.OrderNumber,
		UserID:         r.UserID,
		Subtotal:       r.Subtotal,
		ShippingFee:    r.ShippingFee,
		DiscountAmount: r.DiscountAmount,
		TotalAmount:    r.TotalAmount,
		Status:         domain.OrderStatus(r.Status),
		PaymentStatus:  domain.PaymentStatus(r.PaymentStatus),
		PaymentMethod:  method,
		Shipping: domain.ShippingAddress{
			Recipient:   r.ShipRecipient,
			Phone:       r.ShipPhone,
			Address:     r.ShipAddress,
			Subdistrict: r.ShipSubdistrict,
			District:    r.ShipDistrict,
			Province:    r.ShipProvince,
			PostalCode:  r.ShipPostalCode,
			Country:     r.ShipCountry,
		},
		CustomerNote: deref(r.CustomerNote),
		PlacedAt:     r.PlacedAt,
		CreatedAt:    r.CreatedAt,
		UpdatedAt:    r.UpdatedAt,
	}
}

func toOrderItemDomain(r orderItemRow) domain.OrderItem {
	variantID := uint(0)
	if r.ProductVariantID != nil {
		variantID = *r.ProductVariantID
	}
	return domain.OrderItem{
		ID:          r.ID,
		VariantID:   variantID,
		ProductName: r.ProductName,
		VariantName: r.VariantName,
		UnitPrice:   r.UnitPrice,
		Quantity:    r.Quantity,
		LineTotal:   r.LineTotal,
	}
}

func toPaymentDomain(r paymentRow) domain.Payment {
	return domain.Payment{
		ID:             r.ID,
		OrderID:        r.OrderID,
		Method:         domain.PaymentMethod(r.Method),
		Amount:         r.Amount,
		Status:         domain.PaymentStatus(r.Status),
		SlipURL:        deref(r.SlipURL),
		TransactionRef: deref(r.TransactionRef),
		PaidAt:         r.PaidAt,
		CreatedAt:      r.CreatedAt,
	}
}
