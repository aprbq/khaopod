package postgres

import (
	"context"
	"errors"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"github.com/khaopod/backend/internal/core/domain"
	"github.com/khaopod/backend/internal/core/port/output"
)

type OrderRepo struct{ db *gorm.DB }

var _ output.OrderRepository = (*OrderRepo)(nil)

func NewOrderRepo(db *gorm.DB) *OrderRepo { return &OrderRepo{db: db} }

// Create insert order + items + history แถวแรก — เรียกใน WithinTx ของ service เสมอ
func (r *OrderRepo) Create(ctx context.Context, o *domain.Order) error {
	db := dbFromContext(ctx, r.db)

	method := string(o.PaymentMethod)
	row := orderRow{
		UserID:          o.UserID,
		Subtotal:        o.Subtotal,
		ShippingFee:     o.ShippingFee,
		DiscountAmount:  o.DiscountAmount,
		TotalAmount:     o.TotalAmount,
		Status:          string(o.Status),
		PaymentStatus:   string(o.PaymentStatus),
		PaymentMethod:   nilIfEmpty(method),
		ShipRecipient:   o.Shipping.Recipient,
		ShipPhone:       o.Shipping.Phone,
		ShipAddress:     o.Shipping.Address,
		ShipSubdistrict: o.Shipping.Subdistrict,
		ShipDistrict:    o.Shipping.District,
		ShipProvince:    o.Shipping.Province,
		ShipPostalCode:  o.Shipping.PostalCode,
		ShipCountry:     o.Shipping.Country,
		CustomerNote:    nilIfEmpty(o.CustomerNote),
	}
	if row.ShipCountry == "" {
		row.ShipCountry = "TH"
	}
	if err := db.Create(&row).Error; err != nil {
		return err
	}
	// อ่านค่าที่ DB gen กลับมา (order_number/placed_at มาจาก DEFAULT)
	if err := db.First(&row, row.ID).Error; err != nil {
		return err
	}
	o.ID = row.ID
	o.OrderNumber = row.OrderNumber
	o.PlacedAt = row.PlacedAt
	o.CreatedAt = row.CreatedAt
	o.UpdatedAt = row.UpdatedAt

	for i := range o.Items {
		it := &o.Items[i]
		variantID := it.VariantID
		itemRow := orderItemRow{
			OrderID:          o.ID,
			ProductVariantID: &variantID,
			ProductName:      it.ProductName,
			VariantName:      it.VariantName,
			UnitPrice:        it.UnitPrice,
			Quantity:         it.Quantity,
			LineTotal:        it.LineTotal,
		}
		if err := db.Create(&itemRow).Error; err != nil {
			return err
		}
		it.ID = itemRow.ID
	}

	return db.Create(&statusHistoryRow{OrderID: o.ID, Status: string(o.Status)}).Error
}

func (r *OrderRepo) ListByUser(ctx context.Context, userID uint, f output.OrderListFilter) ([]domain.Order, int, error) {
	db := dbFromContext(ctx, r.db)

	base := db.Model(&orderRow{}).Where("user_id = ?", userID)
	if f.Status != "" {
		base = base.Where("status = ?", string(f.Status))
	}

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []orderRow
	if err := base.Order("placed_at DESC, id DESC").Limit(f.Limit).Offset(f.Offset).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	if len(rows) == 0 {
		return []domain.Order{}, int(total), nil
	}

	// โหลด items ของทุกออเดอร์ในชุดเดียว แล้วจัดกลุ่มกลับ (เลี่ยง N+1)
	ids := make([]uint, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	var itemRows []orderItemRow
	if err := db.Where("order_id IN ?", ids).Order("id").Find(&itemRows).Error; err != nil {
		return nil, 0, err
	}
	itemsByOrder := map[uint][]domain.OrderItem{}
	for _, ir := range itemRows {
		itemsByOrder[ir.OrderID] = append(itemsByOrder[ir.OrderID], toOrderItemDomain(ir))
	}

	out := make([]domain.Order, 0, len(rows))
	for _, row := range rows {
		o := toOrderDomain(row)
		o.Items = itemsByOrder[row.ID]
		out = append(out, o)
	}
	return out, int(total), nil
}

func (r *OrderRepo) FindByNumber(ctx context.Context, userID uint, orderNumber string) (*domain.Order, error) {
	db := dbFromContext(ctx, r.db)

	var row orderRow
	// scope ด้วย user_id — ออเดอร์คนอื่นต้องมองไม่เห็น (กัน IDOR)
	err := db.Where("order_number = ? AND user_id = ?", orderNumber, userID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	o := toOrderDomain(row)

	var itemRows []orderItemRow
	if err := db.Where("order_id = ?", o.ID).Order("id").Find(&itemRows).Error; err != nil {
		return nil, err
	}
	for _, ir := range itemRows {
		o.Items = append(o.Items, toOrderItemDomain(ir))
	}

	// การแจ้งชำระเงินล่าสุด (ถ้ามี)
	var pay paymentRow
	err = db.Where("order_id = ?", o.ID).Order("id DESC").First(&pay).Error
	if err == nil {
		p := toPaymentDomain(pay)
		o.Payment = &p
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	return &o, nil
}

func (r *OrderRepo) UpdateStatus(ctx context.Context, orderID uint, status domain.OrderStatus, paymentStatus domain.PaymentStatus, note string, changedBy uint) error {
	db := dbFromContext(ctx, r.db)
	err := db.Model(&orderRow{}).Where("id = ?", orderID).
		Updates(map[string]any{
			"status":         string(status),
			"payment_status": string(paymentStatus),
		}).Error
	if err != nil {
		return err
	}
	hist := statusHistoryRow{OrderID: orderID, Status: string(status), Note: nilIfEmpty(note)}
	if changedBy != 0 {
		hist.ChangedBy = &changedBy
	}
	return db.Create(&hist).Error
}

func (r *OrderRepo) CreatePayment(ctx context.Context, p *domain.Payment) error {
	row := paymentRow{
		OrderID:        p.OrderID,
		Method:         string(p.Method),
		Amount:         p.Amount,
		Status:         string(p.Status),
		SlipURL:        nilIfEmpty(p.SlipURL),
		TransactionRef: nilIfEmpty(p.TransactionRef),
	}
	if err := dbFromContext(ctx, r.db).Create(&row).Error; err != nil {
		return err
	}
	p.ID = row.ID
	p.CreatedAt = row.CreatedAt
	return nil
}

// ---- ฝั่งแอดมิน (ไม่ scope ด้วย user — สิทธิ์ถูกครอบที่ middleware แล้ว) ----

func (r *OrderRepo) ListAll(ctx context.Context, f output.OrderListFilter) ([]domain.Order, int, error) {
	db := dbFromContext(ctx, r.db)

	base := db.Model(&orderRow{})
	if f.Status != "" {
		base = base.Where("status = ?", string(f.Status))
	}
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []orderRow
	if err := base.Order("placed_at DESC, id DESC").Limit(f.Limit).Offset(f.Offset).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	if len(rows) == 0 {
		return []domain.Order{}, int(total), nil
	}

	ids := make([]uint, 0, len(rows))
	userIDs := make([]uint, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
		userIDs = append(userIDs, row.UserID)
	}

	// email เจ้าของออเดอร์ (โชว์ในหลังบ้าน) — ดึงชุดเดียวแล้ว map กลับ
	var users []struct {
		ID    uint
		Email string
	}
	if err := db.Table("users").Select("id, email").Where("id IN ?", userIDs).Scan(&users).Error; err != nil {
		return nil, 0, err
	}
	emailByUser := make(map[uint]string, len(users))
	for _, u := range users {
		emailByUser[u.ID] = u.Email
	}

	var itemRows []orderItemRow
	if err := db.Where("order_id IN ?", ids).Order("id").Find(&itemRows).Error; err != nil {
		return nil, 0, err
	}
	itemsByOrder := map[uint][]domain.OrderItem{}
	for _, ir := range itemRows {
		itemsByOrder[ir.OrderID] = append(itemsByOrder[ir.OrderID], toOrderItemDomain(ir))
	}

	out := make([]domain.Order, 0, len(rows))
	for _, row := range rows {
		o := toOrderDomain(row)
		o.UserEmail = emailByUser[row.UserID]
		o.Items = itemsByOrder[row.ID]
		out = append(out, o)
	}
	return out, int(total), nil
}

func (r *OrderRepo) FindByNumberAny(ctx context.Context, orderNumber string) (*domain.Order, error) {
	return r.findOneAny(ctx, "order_number = ?", orderNumber)
}

func (r *OrderRepo) FindByIDAny(ctx context.Context, id uint) (*domain.Order, error) {
	return r.findOneAny(ctx, "id = ?", id)
}

func (r *OrderRepo) findOneAny(ctx context.Context, where string, arg any) (*domain.Order, error) {
	db := dbFromContext(ctx, r.db)

	var row orderRow
	err := db.Where(where, arg).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	o := toOrderDomain(row)

	// email เจ้าของออเดอร์ (โชว์ในหลังบ้าน)
	var email string
	if err := db.Table("users").Select("email").Where("id = ?", o.UserID).Scan(&email).Error; err == nil {
		o.UserEmail = email
	}

	var itemRows []orderItemRow
	if err := db.Where("order_id = ?", o.ID).Order("id").Find(&itemRows).Error; err != nil {
		return nil, err
	}
	for _, ir := range itemRows {
		o.Items = append(o.Items, toOrderItemDomain(ir))
	}

	var pay paymentRow
	err = db.Where("order_id = ?", o.ID).Order("id DESC").First(&pay).Error
	if err == nil {
		p := toPaymentDomain(pay)
		o.Payment = &p
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	return &o, nil
}

func (r *OrderRepo) FindPaymentByID(ctx context.Context, id uint) (*domain.Payment, error) {
	var row paymentRow
	err := dbFromContext(ctx, r.db).First(&row, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	p := toPaymentDomain(row)
	return &p, nil
}

func (r *OrderRepo) SavePaymentVerdict(ctx context.Context, paymentID uint, status domain.PaymentStatus, verifiedBy uint) error {
	updates := map[string]any{
		"status":      string(status),
		"verified_by": verifiedBy,
	}
	if status == domain.PaymentPaid {
		updates["paid_at"] = gorm.Expr("now()")
	}
	return dbFromContext(ctx, r.db).Model(&paymentRow{}).Where("id = ?", paymentID).Updates(updates).Error
}

func (r *OrderRepo) Summary(ctx context.Context) (*domain.AdminSummary, error) {
	db := dbFromContext(ctx, r.db)
	s := &domain.AdminSummary{RevenuePaid: decimal.Zero}

	row := struct {
		OrdersTotal   int
		OrdersPending int
		Revenue       decimal.Decimal
	}{}
	err := db.Raw(`
		SELECT COUNT(*) AS orders_total,
		       COUNT(*) FILTER (WHERE status = 'pending') AS orders_pending,
		       COALESCE(SUM(total_amount) FILTER (WHERE payment_status = 'paid'), 0) AS revenue
		FROM orders`).Scan(&row).Error
	if err != nil {
		return nil, err
	}
	s.OrdersTotal = row.OrdersTotal
	s.OrdersPending = row.OrdersPending
	s.RevenuePaid = row.Revenue

	var pendingReview int64
	if err := db.Model(&paymentRow{}).Where("status = ?", "pending_review").Count(&pendingReview).Error; err != nil {
		return nil, err
	}
	s.PaymentsPendingReview = int(pendingReview)
	return s, nil
}
