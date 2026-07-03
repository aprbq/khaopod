package rest

import (
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

func isNumeric(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return len(s) > 0
}
