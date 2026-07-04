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
	r.byID[u.ID] = u
	r.byEmail[domain.NormalizeEmail(u.Email)] = u
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
	items   []domain.Product
	lastF   output.ProductFilter
	findErr error
	listErr error
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

// fakeTx เรียก fn ตรง ๆ ไม่มี transaction จริง
type fakeTx struct{}

func (fakeTx) WithinTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
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
