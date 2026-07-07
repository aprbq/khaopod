package rest

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/khaopod/backend/internal/adapter/inbound/rest/response"
	"github.com/khaopod/backend/internal/core/port/input"
)

// UserHandler แปลง HTTP ↔ UserUseCase (โปรไฟล์ผู้ใช้ที่ล็อกอินอยู่)
type UserHandler struct {
	users input.UserUseCase
}

func NewUserHandler(users input.UserUseCase) *UserHandler {
	return &UserHandler{users: users}
}

// GET /me (🔒)
func (h *UserHandler) GetMe(c *gin.Context) {
	u, err := h.users.GetProfile(c.Request.Context(), c.GetUint(ctxUserID))
	if err != nil {
		mapError(c, err)
		return
	}
	response.OK(c, toUserResponse(u))
}

// PATCH /me (🔒)
func (h *UserHandler) UpdateMe(c *gin.Context) {
	var in updateProfileRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "รูปแบบข้อมูลไม่ถูกต้อง")
		return
	}
	// phone เป็น optional — validate เฉพาะเมื่อส่งมาและไม่ว่าง
	if in.Phone != nil && *in.Phone != "" && !isNumeric(*in.Phone) {
		response.Error(c, 422, response.CodeValidation, "เบอร์โทรต้องเป็นตัวเลขเท่านั้น")
		return
	}

	u, err := h.users.UpdateProfile(c.Request.Context(), c.GetUint(ctxUserID), in.toCommand())
	if err != nil {
		mapError(c, err)
		return
	}
	response.OK(c, toUserResponse(u))
}

// maxAvatarBytes = เพดานขนาดรูปโปรไฟล์ (ต้องตรงกับที่ระบุใน docs/rest_api.md §3.3)
const maxAvatarBytes = 2 << 20 // 2MB

// POST /me/avatar (🔒) — multipart/form-data ฟิลด์ "avatar"
func (h *UserHandler) UploadAvatar(c *gin.Context) {
	fh, err := c.FormFile("avatar")
	if err != nil {
		response.Error(c, 400, response.CodeBadRequest, "กรุณาแนบไฟล์รูปในฟิลด์ avatar")
		return
	}
	if fh.Size > maxAvatarBytes {
		response.Error(c, 422, response.CodeValidation, "ไฟล์รูปต้องไม่เกิน 2MB")
		return
	}
	f, err := fh.Open()
	if err != nil {
		mapError(c, err)
		return
	}
	defer f.Close()
	// LimitReader เผื่อ Size ใน multipart header โกหก — อ่านเกินเพดานเมื่อไหร่ถือว่าไฟล์ใหญ่เกิน
	content, err := io.ReadAll(io.LimitReader(f, maxAvatarBytes+1))
	if err != nil {
		mapError(c, err)
		return
	}
	if len(content) > maxAvatarBytes {
		response.Error(c, 422, response.CodeValidation, "ไฟล์รูปต้องไม่เกิน 2MB")
		return
	}
	// ตรวจชนิดจาก magic bytes ของเนื้อไฟล์จริง — ไม่เชื่อ Content-Type/นามสกุลที่ client ส่งมา
	ext, ok := avatarExt(http.DetectContentType(content))
	if !ok {
		response.Error(c, 422, response.CodeValidation, "รองรับเฉพาะไฟล์ JPG, PNG หรือ WebP")
		return
	}

	u, err := h.users.UpdateAvatar(c.Request.Context(), c.GetUint(ctxUserID),
		input.UpdateAvatarCommand{Content: content, Ext: ext})
	if err != nil {
		mapError(c, err)
		return
	}
	response.OK(c, toUserResponse(u))
}

// avatarExt แปลง MIME ที่ตรวจได้ → นามสกุลไฟล์ (คืน false = ชนิดที่ไม่รองรับ)
func avatarExt(mime string) (string, bool) {
	switch mime {
	case "image/jpeg":
		return ".jpg", true
	case "image/png":
		return ".png", true
	case "image/webp":
		return ".webp", true
	}
	return "", false
}

func isNumeric(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return len(s) > 0
}
