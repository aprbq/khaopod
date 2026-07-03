package rest

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/khaopod/backend/internal/adapter/inbound/rest/response"
	"github.com/khaopod/backend/internal/core/domain"
	"github.com/khaopod/backend/internal/core/port/output"
)

const (
	ctxUserID   = "user_id"
	ctxRole     = "role"
	ctxPublicID = "public_id"
)

// RequireAuth ตรวจ access token จาก header Authorization: Bearer <token>
// สำเร็จ → เก็บ user_id/role/public_id ลง context ให้ handler ใช้ต่อ
func RequireAuth(tokens output.Tokenizer) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		token, ok := bearerToken(header)
		if !ok {
			response.Error(c, 401, response.CodeUnauthorized, "กรุณาเข้าสู่ระบบ")
			return
		}
		claims, err := tokens.ParseAccess(token)
		if err != nil {
			response.Error(c, 401, response.CodeInvalidToken, "เซสชันไม่ถูกต้องหรือหมดอายุ")
			return
		}
		c.Set(ctxUserID, claims.UserID)
		c.Set(ctxRole, string(claims.Role))
		c.Set(ctxPublicID, claims.PublicID)
		c.Next()
	}
}

// RequireAdmin ต้องผ่าน RequireAuth มาก่อน (มี role ใน context)
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetString(ctxRole) != string(domain.RoleAdmin) {
			response.Error(c, 403, response.CodeForbidden, "ต้องเป็นผู้ดูแลระบบเท่านั้น")
			return
		}
		c.Next()
	}
}

func bearerToken(header string) (string, bool) {
	const prefix = "Bearer "
	if len(header) <= len(prefix) || !strings.EqualFold(header[:len(prefix)], prefix) {
		return "", false
	}
	token := strings.TrimSpace(header[len(prefix):])
	return token, token != ""
}
