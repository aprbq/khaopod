package mailer

import (
	"context"
	"net"
	"strconv"
	"testing"
	"time"
)

// เซิร์ฟเวอร์ปลอมที่ accept connection แต่ไม่ส่ง SMTP greeting เลย
// (จำลองอาการที่เจอจริง: มีอะไรยึด port 1025 อยู่แต่ไม่พูด SMTP)
func listenSilent(t *testing.T) (addr string, stop func()) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	done := make(chan struct{})
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			// รับ connection ไว้เฉย ๆ ไม่ตอบอะไร จนกว่าจะปิด
			go func(c net.Conn) {
				<-done
				c.Close()
			}(conn)
		}
	}()
	return ln.Addr().String(), func() { close(done); ln.Close() }
}

// SendOTP ต้องไม่ค้างตลอดไปเมื่อ SMTP ไม่ส่ง greeting — ต้อง error ภายใน timeout
func TestSendOTP_TimesOutWhenServerSilent(t *testing.T) {
	addr, stop := listenSilent(t)
	defer stop()

	host, portStr, _ := net.SplitHostPort(addr)
	port, err := strconv.Atoi(portStr)
	if err != nil {
		t.Fatalf("parse port: %v", err)
	}

	timeout := 300 * time.Millisecond
	m := NewSMTPMailer(host, port, "", "", "no-reply@test.local", timeout)

	start := time.Now()
	err = m.SendOTP(context.Background(), "user@test.local", "123456", 300)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("want error when server never sends greeting, got nil")
	}
	// ต้องคืนภายในเวลาที่สมเหตุสมผล (ไม่ค้าง) — เผื่อ margin จาก timeout
	if elapsed > timeout+2*time.Second {
		t.Fatalf("SendOTP hung too long: %v (timeout=%v)", elapsed, timeout)
	}
}
