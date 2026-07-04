package mailer

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"

	"github.com/khaopod/backend/internal/core/port/output"
)

// SMTPMailer ส่ง OTP ทางอีเมลผ่าน SMTP (implements output.Mailer)
// dev แนะนำใช้ Mailpit/MailHog (localhost:1025) — จะได้ไม่ต้องส่งออกจริงและไม่ leak OTP
type SMTPMailer struct {
	host    string
	port    int
	user    string
	pass    string
	from    string
	timeout time.Duration // bound ทั้ง dial + SMTP conversation กัน request ค้าง
}

var _ output.Mailer = (*SMTPMailer)(nil)

func NewSMTPMailer(host string, port int, user, pass, from string, timeout time.Duration) *SMTPMailer {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &SMTPMailer{host: host, port: port, user: user, pass: pass, from: from, timeout: timeout}
}

func (m *SMTPMailer) SendOTP(ctx context.Context, email, code string, ttlSeconds int) error {
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

	return m.send(ctx, email, []byte(msg))
}

// send = ทำ SMTP conversation เองแทน smtp.SendMail เพื่อคุม timeout ได้
// (smtp.SendMail ไม่มี timeout → ถ้าเซิร์ฟเวอร์ไม่ส่ง greeting request จะค้างตลอดไป)
func (m *SMTPMailer) send(ctx context.Context, to string, msg []byte) error {
	addr := fmt.Sprintf("%s:%d", m.host, m.port)

	// dial ด้วย timeout (เคารพ ctx cancel ด้วย)
	d := net.Dialer{Timeout: m.timeout}
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("smtp dial %s: %w", addr, err)
	}
	// deadline ครอบทุกขั้นหลัง dial (อ่าน greeting, MAIL, RCPT, DATA...) กันค้างกลางทาง
	_ = conn.SetDeadline(time.Now().Add(m.timeout))

	c, err := smtp.NewClient(conn, m.host)
	if err != nil {
		conn.Close()
		return fmt.Errorf("smtp greeting: %w", err)
	}
	defer c.Close()

	if ok, _ := c.Extension("STARTTLS"); ok {
		if err := c.StartTLS(&tls.Config{ServerName: m.host}); err != nil {
			return fmt.Errorf("smtp starttls: %w", err)
		}
	}

	if m.user != "" {
		if ok, _ := c.Extension("AUTH"); ok {
			if err := c.Auth(smtp.PlainAuth("", m.user, m.pass, m.host)); err != nil {
				return fmt.Errorf("smtp auth: %w", err)
			}
		}
	}

	if err := c.Mail(m.from); err != nil {
		return fmt.Errorf("smtp mail from: %w", err)
	}
	if err := c.Rcpt(to); err != nil {
		return fmt.Errorf("smtp rcpt to: %w", err)
	}
	w, err := c.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("smtp write: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("smtp close data: %w", err)
	}
	return c.Quit()
}
