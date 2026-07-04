package rest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func newCORSEngine() *gin.Engine {
	gin.SetMode(gin.TestMode)
	e := gin.New()
	e.Use(CORS([]string{"http://localhost:5173"}))
	e.GET("/ping", func(c *gin.Context) { c.String(http.StatusOK, "pong") })
	return e
}

func TestCORS(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		origin         string
		wantStatus     int
		wantAllowOrig  string // "" = ต้องไม่มี header นี้
		wantCredential bool
	}{
		{
			name:           "allowed origin ได้ header ครบ",
			method:         http.MethodGet,
			origin:         "http://localhost:5173",
			wantStatus:     http.StatusOK,
			wantAllowOrig:  "http://localhost:5173",
			wantCredential: true,
		},
		{
			name:          "origin นอก allowlist ไม่ได้ header (browser จะบล็อกเอง)",
			method:        http.MethodGet,
			origin:        "http://evil.example.com",
			wantStatus:    http.StatusOK,
			wantAllowOrig: "",
		},
		{
			name:          "ไม่มี Origin (เช่น curl/same-origin) ผ่านปกติ",
			method:        http.MethodGet,
			origin:        "",
			wantStatus:    http.StatusOK,
			wantAllowOrig: "",
		},
		{
			name:           "preflight OPTIONS จาก origin ที่อนุญาต ตอบ 204 + header",
			method:         http.MethodOptions,
			origin:         "http://localhost:5173",
			wantStatus:     http.StatusNoContent,
			wantAllowOrig:  "http://localhost:5173",
			wantCredential: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(tt.method, "/ping", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			newCORSEngine().ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status: want %d, got %d", tt.wantStatus, rec.Code)
			}
			if got := rec.Header().Get("Access-Control-Allow-Origin"); got != tt.wantAllowOrig {
				t.Errorf("Allow-Origin: want %q, got %q", tt.wantAllowOrig, got)
			}
			gotCred := rec.Header().Get("Access-Control-Allow-Credentials") == "true"
			if gotCred != tt.wantCredential {
				t.Errorf("Allow-Credentials: want %v, got %v", tt.wantCredential, gotCred)
			}
		})
	}
}
