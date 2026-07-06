# Khaopod Shop

![Go](https://img.shields.io/badge/Go-1.24-00ADD8?logo=go&logoColor=white)
![React](https://img.shields.io/badge/React-19.0-61DAFB?logo=react&logoColor=white)
![TypeScript](https://img.shields.io/badge/TypeScript-5.7-3178C6?logo=typescript&logoColor=white)
![TailwindCSS](https://img.shields.io/badge/TailwindCSS-4.1-06B6D4?logo=tailwindcss&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16+-4169E1?logo=postgresql&logoColor=white)

ร้านค้าออนไลน์ของเพจ **กองบัญชาการข่าวปด** — ขายเสื้อผ้า ของสะสม และสติกเกอร์แนวสตรีทแวร์
เป็นโปรเจกต์ **monorepo** ที่มีทั้ง backend (Go) และ frontend (React)

---

## ภาพรวม

ระบบร้านค้าออนไลน์ครบวงจร ดีไซน์แนวมินิมอล/สตรีทแวร์ responsive รองรับมือถือ พร้อมหลังบ้านสำหรับแอดมิน

**ฟีเจอร์**
- ล็อกอินแบบยืนยันตัวตนด้วย Google และ OTP ทางอีเมล (ไม่มีรหัสผ่าน)
- จัดการโปรไฟล์และที่อยู่จัดส่ง
- ดูสินค้า หมวดหมู่ และรายละเอียดสินค้า (พร้อมตัวเลือก/ไซซ์)
- ตะกร้าสินค้า และสั่งซื้อ (checkout)
- แจ้งชำระเงิน — อัปโหลดสลิป / PromptPay
- ติดตามสถานะคำสั่งซื้อและเลขพัสดุ
- หลังบ้านแอดมิน — จัดการสินค้า ยืนยันการชำระเงิน อัปเดตสถานะ + เลขพัสดุ

---

## Tech Stack

| ส่วน | เทคโนโลยี |
|------|-----------|
| **Backend** | Go · Gin · GORM · PostgreSQL · JWT · govalidator (Hexagonal Architecture) |
| **Frontend** | React · TypeScript · Vite · Tailwind CSS · TanStack Query · react-hook-form + zod · React Router |
| **Dev tools** | Mailpit (ดักอีเมล OTP ตอน dev) · Docker |

---

## โครงสร้าง Repo

```
├── CLAUDE.md              ← context/convention หลักของโปรเจกต์
├── docs/
│   ├── schema.sql         ← โครงสร้างฐานข้อมูล (source of truth)
│   └── rest_api.md        ← สเปค REST API ทั้งระบบ
├── backend/               ← Go + Gin + GORM (Hexagonal) — ดู backend/README.md
│   ├── cmd/server/        ← composition root
│   ├── internal/
│   │   ├── core/          ← domain + ports + services (ไม่พึ่ง framework)
│   │   └── adapter/       ← inbound (HTTP) + outbound (postgres/mailer/jwt/google)
│   └── migrations/
└── frontend/              ← React + Vite
    └── src/
        ├── components/    ← UI ที่ใช้ซ้ำ (Navbar, ProductCard, ui/...)
        ├── features/      ← จัดกลุ่มตาม domain (auth, products)
        ├── pages/         ← หน้าเพจ (Home, Shop, ProductDetail, Login, Account...)
        └── lib/           ← apiClient, utils
```

---

## วิธีรัน (Local Development)

**ต้องมีก่อน:** Go 1.24+ · Node 20+ · PostgreSQL · Docker (สำหรับ Mailpit)

### 1. ฐานข้อมูล
สร้าง database `khaopod_news_shop` แล้วรัน SQL ใน `backend/migrations/`

### 2. อีเมล (dev) — Mailpit ดัก OTP ไว้ดูในเครื่อง
```bash
docker run -d --name kbc-mailpit -p 1025:1025 -p 8025:8025 axllent/mailpit
```

### 3. Backend
```bash
cd backend
cp .env.example .env      # แล้วแก้ DATABASE_URL / secret ให้ถูก
go run ./cmd/server       # ขึ้นที่ :8080
```

### 4. Frontend
```bash
cd frontend
npm install
npm run dev               # ขึ้นที่ :5173 (proxy /v1 ไป backend ให้อัตโนมัติ)
```

เปิด **http://localhost:5173** แล้วลองล็อกอิน → อ่าน OTP ที่ **http://localhost:8025**

### พอร์ตที่ใช้
| บริการ | URL |
|--------|-----|
| Frontend | http://localhost:5173 |
| Backend API | http://localhost:8080 |
| Mailpit (อ่านอีเมล OTP) | http://localhost:8025 |
| PostgreSQL | localhost:5432 |

---

## เอกสารอ้างอิง
- [`CLAUDE.md`](CLAUDE.md) — ภาพรวม + convention ของโปรเจกต์
- [`docs/schema.sql`](docs/schema.sql) — โครงสร้างฐานข้อมูล
- [`docs/rest_api.md`](docs/rest_api.md) — สเปค API ทุก endpoint
- [`backend/README.md`](backend/README.md) — รายละเอียดฝั่ง backend

---

## ความปลอดภัย
OTP เก็บเป็นแฮช ใช้ครั้งเดียว และมี rate limit กัน brute-force · refresh token เป็น httpOnly cookie พร้อม rotation ทุกครั้ง · ตัดสต็อกอยู่ใน transaction เดียวกับการสร้างออเดอร์ · secret ทั้งหมดอ่านจาก environment variable
