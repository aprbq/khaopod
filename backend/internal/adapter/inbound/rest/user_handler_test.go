package rest

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/khaopod/backend/internal/core/domain"
	"github.com/khaopod/backend/internal/core/port/input"
)

// fakeAvatarUC จำ command ที่ handler ส่งเข้ามา ไว้ตรวจว่า sniff ชนิดไฟล์ถูก
type fakeAvatarUC struct {
	fakeUserUC
	gotExt string
}

func (f *fakeAvatarUC) UpdateAvatar(_ context.Context, _ uint, cmd input.UpdateAvatarCommand) (*domain.User, error) {
	f.gotExt = cmd.Ext
	return &domain.User{PublicID: "pub-1", Email: "a@b.com", AvatarURL: "/uploads/avatars/x.png"}, nil
}

func setupUserHandler(uc input.UserUseCase) *gin.Engine {
	gin.SetMode(gin.TestMode)
	e := gin.New()
	e.POST("/me/avatar", NewUserHandler(uc).UploadAvatar)
	return e
}

func doUpload(e *gin.Engine, field, filename string, content []byte) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, _ := w.CreateFormFile(field, filename)
	_, _ = fw.Write(content)
	_ = w.Close()

	req := httptest.NewRequest(http.MethodPost, "/me/avatar", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

// magic bytes ของ PNG — DetectContentType ดูแค่ signature ต้นไฟล์
var pngBytes = []byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a, 0, 0, 0, 0}

func TestUploadAvatar_AcceptsPNG(t *testing.T) {
	uc := &fakeAvatarUC{}
	rec := doUpload(setupUserHandler(uc), "avatar", "me.png", pngBytes)
	if rec.Code != 200 {
		t.Fatalf("want 200, got %d (%s)", rec.Code, rec.Body.String())
	}
	if uc.gotExt != ".png" {
		t.Fatalf("want ext .png from sniffed type, got %q", uc.gotExt)
	}
}

func TestUploadAvatar_RejectsNonImage(t *testing.T) {
	// นามสกุล .png แต่เนื้อไฟล์เป็น text → ต้องดูเนื้อไฟล์จริง ไม่เชื่อชื่อไฟล์
	rec := doUpload(setupUserHandler(&fakeAvatarUC{}), "avatar", "fake.png", []byte("hello not an image"))
	if rec.Code != 422 {
		t.Fatalf("want 422, got %d (%s)", rec.Code, rec.Body.String())
	}
}

func TestUploadAvatar_RejectsMissingFile(t *testing.T) {
	rec := doUpload(setupUserHandler(&fakeAvatarUC{}), "wrong_field", "me.png", pngBytes)
	if rec.Code != 400 {
		t.Fatalf("want 400, got %d (%s)", rec.Code, rec.Body.String())
	}
}

func TestUploadAvatar_RejectsOversize(t *testing.T) {
	big := append(append([]byte{}, pngBytes...), bytes.Repeat([]byte{0}, maxAvatarBytes)...)
	rec := doUpload(setupUserHandler(&fakeAvatarUC{}), "avatar", "big.png", big)
	if rec.Code != 422 {
		t.Fatalf("want 422, got %d (%s)", rec.Code, rec.Body.String())
	}
}
