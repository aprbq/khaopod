package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORS อนุญาตให้ frontend ที่อยู่คนละ origin (เช่น Vite dev ที่ :5173) ยิงตรงมา backend ได้
// browser เป็นคนบังคับ CORS — ถ้า backend ไม่ส่ง header อนุญาต request ข้าม origin จะถูกบล็อก
// echo เฉพาะ origin ที่อยู่ใน allowlist (ห้ามใช้ "*" คู่กับ credentials เพราะ browser จะปฏิเสธ)
func CORS(allowedOrigins []string) gin.HandlerFunc {
	allowed := make(map[string]bool, len(allowedOrigins))
	for _, o := range allowedOrigins {
		allowed[o] = true
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin != "" && allowed[origin] {
			h := c.Writer.Header()
			h.Set("Access-Control-Allow-Origin", origin)
			h.Add("Vary", "Origin") // response ต่าง origin แคชแยกกัน
			h.Set("Access-Control-Allow-Credentials", "true")
			h.Set("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT, DELETE, OPTIONS")
			h.Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			h.Set("Access-Control-Max-Age", "600") // แคช preflight 10 นาที ลดจำนวน OPTIONS
		}

		// preflight (OPTIONS) จบตรงนี้ ไม่ต้องวิ่งเข้า handler จริง
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
