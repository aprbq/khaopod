package rest

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/khaopod/backend/internal/adapter/inbound/rest/response"
)

// readImageUpload อ่านไฟล์รูปจาก multipart field: ตรวจขนาด + ชนิดจาก magic bytes จริง
// ถ้าไม่ผ่านจะตอบ error response ให้เองแล้วคืน ok=false — ผู้เรียกแค่ return
func readImageUpload(c *gin.Context, field string, maxBytes int) (content []byte, ext string, ok bool) {
	fh, err := c.FormFile(field)
	if err != nil {
		response.Error(c, 400, response.CodeBadRequest, "กรุณาแนบไฟล์รูปในฟิลด์ "+field)
		return nil, "", false
	}
	tooBig := fmt.Sprintf("ไฟล์รูปต้องไม่เกิน %dMB", maxBytes>>20)
	if fh.Size > int64(maxBytes) {
		response.Error(c, 422, response.CodeValidation, tooBig)
		return nil, "", false
	}
	f, err := fh.Open()
	if err != nil {
		mapError(c, err)
		return nil, "", false
	}
	defer f.Close()
	// LimitReader เผื่อ Size ใน multipart header โกหก
	content, err = io.ReadAll(io.LimitReader(f, int64(maxBytes)+1))
	if err != nil {
		mapError(c, err)
		return nil, "", false
	}
	if len(content) > maxBytes {
		response.Error(c, 422, response.CodeValidation, tooBig)
		return nil, "", false
	}
	ext, valid := avatarExt(http.DetectContentType(content))
	if !valid {
		response.Error(c, 422, response.CodeValidation, "รองรับเฉพาะไฟล์ JPG, PNG หรือ WebP")
		return nil, "", false
	}
	return content, ext, true
}
