# Backend — Khaopod News Shop

Go + Gin + GORM + PostgreSQL แบบ **Hexagonal Architecture (Ports & Adapters)**
สโคปปัจจุบัน: **Auth (passwordless OTP + Google) และ User profile** เท่านั้น

## โครงสร้าง

```
internal/
├── core/                    ← ในหกเหลี่ยม (ห้าม import gin/gorm/jwt)
│   ├── domain/              ← entity + domain logic (user, otp/session, errors)
│   ├── port/input/          ← use-case interface (AuthUseCase, UserUseCase)
│   ├── port/output/         ← repo/mailer/tokenizer/google/txmanager interface
│   └── service/             ← use-case impl + unit test (fake output port)
└── adapter/
    ├── inbound/rest/        ← Gin handler, dto, middleware, rate limit, response envelope
    └── outbound/
        ├── postgres/        ← GORM: persistence model + repo + txmanager
        ├── mailer/          ← ส่ง OTP ผ่าน SMTP
        ├── auth/            ← JWT tokenizer (access token)
        └── google/          ← verify Google id_token (JWKS)
```

## เริ่มใช้งาน

```bash
cp .env.example .env         # แล้วแก้ค่า secret ให้เรียบร้อย

# รัน migration (ต้องมี golang-migrate)
migrate -path migrations -database "$DATABASE_URL" up

go run ./cmd/server          # dev server (default :8080)
go test ./...                # unit test (core + handler)
gofmt -l . && go vet ./...   # format / lint
```

> `go test -race` ต้องใช้ gcc/MinGW เวอร์ชันใหม่ — เครื่อง dev นี้ติดตั้ง MinGW 8.1.0 (เก่า)
> ทำให้ race binary โหลดไม่ขึ้น (`0xc0000139`); ปกติ `go test ./...` ผ่านครบ

## Endpoint (ตรงกับ docs/rest_api.md)

| Method | Path | สิทธิ์ |
|--------|------|--------|
| POST | /v1/auth/otp/request | 🔓 (rate limit ต่อ IP + อีเมล) |
| POST | /v1/auth/google | 🔓 |
| POST | /v1/auth/otp/verify | 🔓 (rate limit ต่อ IP) |
| POST | /v1/auth/refresh | 🔓 (rotate refresh token) |
| POST | /v1/auth/logout | 🔒 |
| GET  | /v1/auth/me | 🔒 |
| GET  | /v1/me | 🔒 |
| PATCH | /v1/me | 🔒 |

## หมายเหตุความปลอดภัยที่ทำไว้แล้ว

- OTP เก็บเฉพาะ **HMAC-SHA256** (ไม่เก็บเลขจริง), อายุสั้น, ใช้ครั้งเดียว, ล็อกเมื่อกรอกผิดเกิน `OTP_MAX_ATTEMPTS`
- refresh token เก็บเป็น **SHA256 hash**, ตั้งเป็น httpOnly cookie, rotate ทุกครั้งที่ refresh
- rate limit endpoint OTP ทั้งต่อ IP และต่ออีเมล กัน brute-force
- logout มี ownership check (revoke ได้เฉพาะ session ของตัวเอง) กัน IDOR
- secret ทุกตัวอ่านจาก env, ไม่ log OTP/token
