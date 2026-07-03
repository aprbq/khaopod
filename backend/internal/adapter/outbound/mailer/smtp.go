package mailer

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/khaopod/backend/internal/core/port/output"
)

// SMTPMailer ส่ง OTP ทางอีเมลผ่าน SMTP (implements output.Mailer)
// dev แนะนำใช้ Mailpit/MailHog (localhost:1025) — จะได้ไม่ต้องส่งออกจริงและไม่ leak OTP
type SMTPMailer struct {
	host string
	port int
	user string
	pass string
	from string
}

var _ output.Mailer = (*SMTPMailer)(nil)

func NewSMTPMailer(host string, port int, user, pass, from string) *SMTPMailer {
	return &SMTPMailer{host: host, port: port, user: user, pass: pass, from: from}
}

func (m *SMTPMailer) SendOTP(_ context.Context, email, code string, ttlSeconds int) error {
	minutes := ttlSeconds / 60
	subject := "รหัส OTP สำหรับเข้าสู่ระบบ Khaopod News Shop"
	body := fmt.Sprintf(
		"รหัส OTP ของคุณคือ %s\n\nรหัสนี้จะหมดอายุใน %d นาที และใช้ได้ครั้งเดียว\nหากคุณไม่ได้ร้องขอ กรุณาเพิกเฉยอีเมลนี้",
		code, minutes,
	)

	msg := strings.Join([]string{
		"From: " + m.from,
		"To: " + email,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=\"UTF-8\"",
		"",
		body,
	}, "\r\n")

	addr := fmt.Sprintf("%s:%d", m.host, m.port)
	var auth smtp.Auth
	if m.user != "" {
		auth = smtp.PlainAuth("", m.user, m.pass, m.host)
	}
	// auth = nil ได้สำหรับ SMTP dev (Mailpit) ที่ไม่ต้อง login
	return smtp.SendMail(addr, auth, m.from, []string{email}, []byte(msg))
}
