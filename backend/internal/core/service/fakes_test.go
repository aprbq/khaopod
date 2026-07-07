package service

import (
	"context"

	"github.com/khaopod/backend/internal/core/domain"
	"github.com/khaopod/backend/internal/core/port/output"
)

// fake output-port implementations สำหรับเทส core โดยไม่ต้องมี DB/HTTP จริง

type fakeUserRepo struct {
	byID      map[uint]*domain.User
	byEmail   map[string]*domain.User
	nextID    uint
	createErr error
	updateErr error
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{byID: map[uint]*domain.User{}, byEmail: map[string]*domain.User{}, nextID: 1}
}

func (r *fakeUserRepo) seed(u *domain.User) *domain.User {
	if u.ID == 0 {
		u.ID = r.nextID
		r.nextID++
	}
	if u.PublicID == "" {
		u.PublicID = "pub-" + u.Email
	}
	r.byID[u.ID] = u
	r.byEmail[domain.NormalizeEmail(u.Email)] = u
	return u
}

func (r *fakeUserRepo) FindByID(_ context.Context, id uint) (*domain.User, error) {
	if u, ok := r.byID[id]; ok {
		return u, nil
	}
	return nil, domain.ErrNotFound
}

func (r *fakeUserRepo) FindByPublicID(_ context.Context, pid string) (*domain.User, error) {
	for _, u := range r.byID {
		if u.PublicID == pid {
			return u, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *fakeUserRepo) FindByEmail(_ context.Context, email string) (*domain.User, error) {
	if u, ok := r.byEmail[domain.NormalizeEmail(email)]; ok {
		return u, nil
	}
	return nil, domain.ErrNotFound
}

func (r *fakeUserRepo) Create(_ context.Context, u *domain.User) error {
	if r.createErr != nil {
		return r.createErr
	}
	r.seed(u)
	return nil
}

func (r *fakeUserRepo) Update(_ context.Context, u *domain.User) error {
	if r.updateErr != nil {
		return r.updateErr
	}
	r.byID[u.ID] = u
	r.byEmail[domain.NormalizeEmail(u.Email)] = u
	return nil
}

func (r *fakeUserRepo) ListAll(_ context.Context, limit, offset int) ([]domain.User, int, error) {
	var all []domain.User
	for id := uint(1); id < r.nextID; id++ {
		if u, ok := r.byID[id]; ok {
			all = append(all, *u)
		}
	}
	total := len(all)
	if offset > total {
		offset = total
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return all[offset:end], total, nil
}

// fakeFileStorage เก็บไฟล์ใน memory — จำทั้งที่เซฟและที่สั่งลบ ไว้ตรวจใน assert
type fakeFileStorage struct {
	saved   map[string][]byte
	removed []string
	saveErr error
}

func (f *fakeFileStorage) Save(_ context.Context, relPath string, content []byte) (string, error) {
	if f.saveErr != nil {
		return "", f.saveErr
	}
	if f.saved == nil {
		f.saved = map[string][]byte{}
	}
	url := "/uploads/" + relPath
	f.saved[url] = content
	return url, nil
}

func (f *fakeFileStorage) Remove(_ context.Context, url string) error {
	f.removed = append(f.removed, url)
	delete(f.saved, url)
	return nil
}

type fakeOTPRepo struct {
	items       []*domain.OTPCode
	nextID      uint
	invalidated int
}

func newFakeOTPRepo() *fakeOTPRepo { return &fakeOTPRepo{nextID: 1} }

func (r *fakeOTPRepo) Create(_ context.Context, o *domain.OTPCode) error {
	o.ID = r.nextID
	r.nextID++
	r.items = append(r.items, o)
	return nil
}

func (r *fakeOTPRepo) FindLatestActive(_ context.Context, email string, p domain.OTPPurpose) (*domain.OTPCode, error) {
	for i := len(r.items) - 1; i >= 0; i-- {
		o := r.items[i]
		if domain.NormalizeEmail(o.Email) == domain.NormalizeEmail(email) && o.Purpose == p && !o.IsConsumed() {
			return o, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *fakeOTPRepo) Save(_ context.Context, o *domain.OTPCode) error {
	for i, it := range r.items {
		if it.ID == o.ID {
			r.items[i] = o
			return nil
		}
	}
	return domain.ErrNotFound
}

func (r *fakeOTPRepo) InvalidateActive(_ context.Context, email string, p domain.OTPPurpose) error {
	r.invalidated++
	return nil
}

type fakeSessionRepo struct {
	byHash map[string]*domain.Session
	nextID uint
}

func newFakeSessionRepo() *fakeSessionRepo {
	return &fakeSessionRepo{byHash: map[string]*domain.Session{}, nextID: 1}
}

func (r *fakeSessionRepo) Create(_ context.Context, s *domain.Session) error {
	s.ID = r.nextID
	r.nextID++
	r.byHash[s.RefreshTokenHash] = s
	return nil
}

func (r *fakeSessionRepo) FindByTokenHash(_ context.Context, hash string) (*domain.Session, error) {
	if s, ok := r.byHash[hash]; ok {
		return s, nil
	}
	return nil, domain.ErrNotFound
}

func (r *fakeSessionRepo) Save(_ context.Context, s *domain.Session) error {
	r.byHash[s.RefreshTokenHash] = s
	return nil
}

type fakeOAuthRepo struct{ upserts int }

func (r *fakeOAuthRepo) Upsert(_ context.Context, _ *domain.OAuthAccount) error {
	r.upserts++
	return nil
}

type fakeMailer struct {
	sentTo   string
	sentCode string
	err      error
}

func (m *fakeMailer) SendOTP(_ context.Context, email, code string, _ int) error {
	if m.err != nil {
		return m.err
	}
	m.sentTo = email
	m.sentCode = code
	return nil
}

type fakeTokenizer struct{ issued int }

func (t *fakeTokenizer) IssueAccess(u *domain.User) (string, int, error) {
	t.issued++
	return "access-for-" + u.Email, 900, nil
}

func (t *fakeTokenizer) ParseAccess(string) (*output.AccessClaims, error) {
	return nil, domain.ErrInvalidToken
}

type fakeGoogle struct {
	identity *output.GoogleIdentity
	err      error
}

func (g *fakeGoogle) Verify(context.Context, string) (*output.GoogleIdentity, error) {
	if g.err != nil {
		return nil, g.err
	}
	return g.identity, nil
}

// fakeProductRepo — เก็บสินค้าใน memory + บันทึก filter ที่ service ส่งมา (ไว้ตรวจ normalize)
type fakeProductRepo struct {
	items      []domain.Product
	variants   map[uint]domain.ProductVariant // สำหรับ FindVariantByID
	lastF      output.ProductFilter
	findErr    error
	listErr    error
	variantErr error
}

func (r *fakeProductRepo) List(_ context.Context, f output.ProductFilter) ([]domain.Product, int, error) {
	r.lastF = f
	if r.listErr != nil {
		return nil, 0, r.listErr
	}
	return r.items, len(r.items), nil
}

func (r *fakeProductRepo) FindBySlug(_ context.Context, slug string) (*domain.Product, error) {
	if r.findErr != nil {
		return nil, r.findErr
	}
	for i := range r.items {
		if r.items[i].Slug == slug {
			return &r.items[i], nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *fakeProductRepo) FindVariantByID(_ context.Context, id uint) (*domain.ProductVariant, error) {
	if r.variantErr != nil {
		return nil, r.variantErr
	}
	if v, ok := r.variants[id]; ok {
		return &v, nil
	}
	return nil, domain.ErrNotFound
}

// fake ไม่มี lock จริง — พฤติกรรมเทียบเท่า FindVariantByID (เทส service ไม่ทดสอบ concurrency ของ DB)
func (r *fakeProductRepo) GetVariantForUpdate(_ context.Context, id uint) (*domain.ProductVariant, error) {
	if v, ok := r.variants[id]; ok {
		return &v, nil
	}
	return nil, domain.ErrNotFound
}

func (r *fakeProductRepo) SaveVariantStock(_ context.Context, v *domain.ProductVariant) error {
	r.variants[v.ID] = *v
	return nil
}

func (r *fakeProductRepo) ListCategories(context.Context) ([]domain.Category, error) {
	return nil, nil
}

// fakeCatalogRepo — แคตตาล็อกใน memory สำหรับเทสหลังบ้าน
type fakeCatalogRepo struct {
	byID      map[uint]*domain.Product
	nextID    uint
	deleteErr error
	images    map[uint]*domain.ProductImage
	nextImgID uint
}

func newFakeCatalogRepo() *fakeCatalogRepo {
	return &fakeCatalogRepo{byID: map[uint]*domain.Product{}, nextID: 1, images: map[uint]*domain.ProductImage{}, nextImgID: 1}
}

func (r *fakeCatalogRepo) ListAllProducts(_ context.Context, _ output.ProductFilter) ([]domain.Product, int, error) {
	var out []domain.Product
	for id := uint(1); id < r.nextID; id++ {
		if p, ok := r.byID[id]; ok {
			out = append(out, *p)
		}
	}
	return out, len(out), nil
}

func (r *fakeCatalogRepo) FindProductByID(_ context.Context, id uint) (*domain.Product, error) {
	if p, ok := r.byID[id]; ok {
		return p, nil
	}
	return nil, domain.ErrNotFound
}

func (r *fakeCatalogRepo) CreateProduct(_ context.Context, p *domain.Product) error {
	p.ID = r.nextID
	r.nextID++
	r.byID[p.ID] = p
	return nil
}

func (r *fakeCatalogRepo) UpdateProduct(_ context.Context, p *domain.Product) error {
	r.byID[p.ID] = p
	return nil
}

func (r *fakeCatalogRepo) DeleteProduct(_ context.Context, id uint) error {
	if r.deleteErr != nil {
		return r.deleteErr
	}
	delete(r.byID, id)
	return nil
}

func (r *fakeCatalogRepo) CreateVariant(_ context.Context, v *domain.ProductVariant) error {
	p := r.byID[v.ProductID]
	v.ID = uint(len(p.Variants) + 1)
	p.Variants = append(p.Variants, *v)
	return nil
}

func (r *fakeCatalogRepo) UpdateVariant(_ context.Context, v *domain.ProductVariant) error {
	for _, p := range r.byID {
		for i := range p.Variants {
			if p.Variants[i].ID == v.ID {
				p.Variants[i] = *v
			}
		}
	}
	return nil
}

func (r *fakeCatalogRepo) DeleteVariant(context.Context, uint) error { return nil }

func (r *fakeCatalogRepo) AddImage(_ context.Context, img *domain.ProductImage) error {
	img.ID = r.nextImgID
	r.nextImgID++
	r.images[img.ID] = img
	if p, ok := r.byID[img.ProductID]; ok {
		p.Images = append(p.Images, *img)
	}
	return nil
}

func (r *fakeCatalogRepo) DeleteImage(_ context.Context, id uint) (string, error) {
	img, ok := r.images[id]
	if !ok {
		return "", domain.ErrNotFound
	}
	delete(r.images, id)
	return img.URL, nil
}

func (r *fakeCatalogRepo) SetPrimaryImage(_ context.Context, productID, imageID uint) error {
	for _, img := range r.images {
		if img.ProductID == productID {
			img.IsPrimary = img.ID == imageID
		}
	}
	return nil
}

// fakeTx เรียก fn ตรง ๆ ไม่มี transaction จริง
type fakeTx struct{}

func (fakeTx) WithinTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

// fakeAddressRepo — ที่อยู่ใน memory (scope ด้วย userID เหมือน repo จริง)
type fakeAddressRepo struct {
	byID   map[uint]*domain.Address
	nextID uint
}

func newFakeAddressRepo() *fakeAddressRepo {
	return &fakeAddressRepo{byID: map[uint]*domain.Address{}, nextID: 1}
}

func (r *fakeAddressRepo) seed(a *domain.Address) *domain.Address {
	if a.ID == 0 {
		a.ID = r.nextID
		r.nextID++
	}
	r.byID[a.ID] = a
	return a
}

func (r *fakeAddressRepo) ListByUser(_ context.Context, userID uint) ([]domain.Address, error) {
	var out []domain.Address
	for id := uint(1); id < r.nextID; id++ { // วนตาม id ให้ลำดับคงที่
		if a, ok := r.byID[id]; ok && a.UserID == userID {
			out = append(out, *a)
		}
	}
	return out, nil
}

func (r *fakeAddressRepo) FindByID(_ context.Context, userID, id uint) (*domain.Address, error) {
	if a, ok := r.byID[id]; ok && a.UserID == userID {
		cp := *a
		return &cp, nil
	}
	return nil, domain.ErrNotFound
}

func (r *fakeAddressRepo) Create(_ context.Context, a *domain.Address) error {
	r.seed(a)
	return nil
}

func (r *fakeAddressRepo) Update(_ context.Context, a *domain.Address) error {
	r.byID[a.ID] = a
	return nil
}

func (r *fakeAddressRepo) Delete(_ context.Context, userID, id uint) error {
	if a, ok := r.byID[id]; ok && a.UserID == userID {
		delete(r.byID, id)
	}
	return nil
}

func (r *fakeAddressRepo) ClearDefault(_ context.Context, userID uint) error {
	for _, a := range r.byID {
		if a.UserID == userID {
			a.IsDefault = false
		}
	}
	return nil
}

// fakeOrderRepo — คำสั่งซื้อใน memory + จำสิ่งที่ service สั่ง (status/payment) ไว้ตรวจ
type fakeOrderRepo struct {
	byNumber       map[string]*domain.Order
	nextID         uint
	statusLog      []domain.OrderStatus
	payments       []*domain.Payment
	createErr      error
	paymentErr     error
	lastChangedBy  uint
	lastVerifiedBy uint
}

func newFakeOrderRepo() *fakeOrderRepo {
	return &fakeOrderRepo{byNumber: map[string]*domain.Order{}, nextID: 1}
}

func (r *fakeOrderRepo) Create(_ context.Context, o *domain.Order) error {
	if r.createErr != nil {
		return r.createErr
	}
	o.ID = r.nextID
	r.nextID++
	o.OrderNumber = fakeOrderNumber(o.ID)
	r.byNumber[o.OrderNumber] = o
	return nil
}

func fakeOrderNumber(id uint) string { return "ORD-TEST-" + string(rune('0'+id)) }

func (r *fakeOrderRepo) ListByUser(_ context.Context, userID uint, f output.OrderListFilter) ([]domain.Order, int, error) {
	var out []domain.Order
	for _, o := range r.byNumber {
		if o.UserID == userID && (f.Status == "" || o.Status == f.Status) {
			out = append(out, *o)
		}
	}
	return out, len(out), nil
}

func (r *fakeOrderRepo) FindByNumber(_ context.Context, userID uint, number string) (*domain.Order, error) {
	if o, ok := r.byNumber[number]; ok && o.UserID == userID {
		return o, nil
	}
	return nil, domain.ErrNotFound
}

func (r *fakeOrderRepo) UpdateStatus(_ context.Context, orderID uint, status domain.OrderStatus, ps domain.PaymentStatus, _ string, changedBy uint) error {
	r.statusLog = append(r.statusLog, status)
	r.lastChangedBy = changedBy
	for _, o := range r.byNumber {
		if o.ID == orderID {
			o.Status = status
			o.PaymentStatus = ps
		}
	}
	return nil
}

func (r *fakeOrderRepo) CreatePayment(_ context.Context, p *domain.Payment) error {
	if r.paymentErr != nil {
		return r.paymentErr
	}
	p.ID = uint(len(r.payments) + 1)
	r.payments = append(r.payments, p)
	return nil
}

// ---- ฝั่งแอดมิน ----

func (r *fakeOrderRepo) ListAll(_ context.Context, f output.OrderListFilter) ([]domain.Order, int, error) {
	var out []domain.Order
	for _, o := range r.byNumber {
		if f.Status == "" || o.Status == f.Status {
			out = append(out, *o)
		}
	}
	return out, len(out), nil
}

func (r *fakeOrderRepo) FindByNumberAny(_ context.Context, number string) (*domain.Order, error) {
	if o, ok := r.byNumber[number]; ok {
		return o, nil
	}
	return nil, domain.ErrNotFound
}

func (r *fakeOrderRepo) FindByIDAny(_ context.Context, id uint) (*domain.Order, error) {
	for _, o := range r.byNumber {
		if o.ID == id {
			return o, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *fakeOrderRepo) FindPaymentByID(_ context.Context, id uint) (*domain.Payment, error) {
	for _, p := range r.payments {
		if p.ID == id {
			return p, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *fakeOrderRepo) SavePaymentVerdict(_ context.Context, paymentID uint, status domain.PaymentStatus, verifiedBy uint) error {
	for _, p := range r.payments {
		if p.ID == paymentID {
			p.Status = status
		}
	}
	r.lastVerifiedBy = verifiedBy
	return nil
}

func (r *fakeOrderRepo) Summary(_ context.Context) (*domain.AdminSummary, error) {
	return &domain.AdminSummary{OrdersTotal: len(r.byNumber)}, nil
}

// harness รวม dependency ทั้งหมดไว้ประกอบ service ในเทส
type harness struct {
	users    *fakeUserRepo
	otps     *fakeOTPRepo
	sessions *fakeSessionRepo
	oauth    *fakeOAuthRepo
	mailer   *fakeMailer
	tokens   *fakeTokenizer
	google   *fakeGoogle
	svc      *AuthService
}

func newHarness() *harness {
	h := &harness{
		users:    newFakeUserRepo(),
		otps:     newFakeOTPRepo(),
		sessions: newFakeSessionRepo(),
		oauth:    &fakeOAuthRepo{},
		mailer:   &fakeMailer{},
		tokens:   &fakeTokenizer{},
		google:   &fakeGoogle{},
	}
	h.svc = NewAuthService(h.users, h.otps, h.sessions, h.oauth, h.mailer, h.tokens, h.google, fakeTx{}, testAuthConfig())
	return h
}
