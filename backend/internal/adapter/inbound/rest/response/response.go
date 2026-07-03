package response

import "github.com/gin-gonic/gin"

// Envelope helper — ทุก response ในระบบห่อด้วยรูปแบบเดียวกัน (ดู CLAUDE.md / rest_api.md)
// อย่าเขียน format เฉพาะกิจในแต่ละ handler

// Meta ใช้กับ response แบบ list (pagination)
type Meta struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

func OK(c *gin.Context, data any) {
	c.JSON(200, gin.H{"success": true, "data": data})
}

func Created(c *gin.Context, data any) {
	c.JSON(201, gin.H{"success": true, "data": data})
}

func NoContent(c *gin.Context) {
	c.Status(204)
}

func List(c *gin.Context, data any, meta Meta) {
	c.JSON(200, gin.H{"success": true, "data": data, "meta": meta})
}

// Error คืน envelope error — code เป็น machine-readable, message เป็นภาษาไทยสำหรับผู้ใช้
// ใช้ AbortWithStatusJSON เพื่อหยุด middleware chain ที่เหลือ
func Error(c *gin.Context, status int, code, message string) {
	c.AbortWithStatusJSON(status, gin.H{
		"success": false,
		"error":   gin.H{"code": code, "message": message},
	})
}
