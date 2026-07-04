# REST API Spec — เพจกองบัญชาการข่าวปด (KBC News Shop)

เอกสารสเปค REST API ทั้งระบบ อิงตาม database schema ที่ออกแบบไว้

---

## 1. ข้อตกลงทั่วไป (Conventions)

**Base URL**
```
https://api.kbcnews.shop/api/v1
```
(dev: `http://localhost:8080/api/v1`)

**Authentication**
ใช้ JWT ผ่าน header ทุก endpoint ที่ต้องล็อกอิน:
```
Authorization: Bearer <access_token>
```
- `access_token` อายุสั้น (เช่น 15 นาที) — ใช้เรียก API
- `refresh_token` อายุยาว (เช่น 30 วัน) — ใช้ขอ access token ใหม่ (เก็บใน httpOnly cookie จะปลอดภัยสุด)

**สิทธิ์การเข้าถึง (ในเอกสารนี้)**
- 🔓 = เรียกได้โดยไม่ต้องล็อกอิน
- 🔒 = ต้องล็อกอิน (customer)
- 🛡️ = ต้องเป็น admin เท่านั้น

**รูปแบบ response สำเร็จ** — ห่อด้วย envelope เดียวกันทั้งระบบ
```json
{
  "success": true,
  "data": { }
}
```

**รูปแบบ response ที่เป็น list** — มี pagination
```json
{
  "success": true,
  "data": [ ],
  "meta": {
    "page": 1,
    "per_page": 20,
    "total": 137,
    "total_pages": 7
  }
}
```

**รูปแบบ error**
```json
{
  "success": false,
  "error": {
    "code": "INVALID_OTP",
    "message": "รหัส OTP ไม่ถูกต้องหรือหมดอายุแล้ว"
  }
}
```

**HTTP status codes ที่ใช้**
| Code | ความหมาย |
|------|----------|
| 200  | สำเร็จ |
| 201  | สร้างข้อมูลใหม่สำเร็จ |
| 400  | ข้อมูลที่ส่งมาไม่ถูกต้อง |
| 401  | ยังไม่ล็อกอิน / token หมดอายุ |
| 403  | ไม่มีสิทธิ์ (เช่นไม่ใช่ admin) |
| 404  | ไม่พบข้อมูล |
| 409  | ขัดแย้ง (เช่นสินค้าหมดสต็อก, ตะกร้าซ้ำ) |
| 422  | validation ไม่ผ่าน |
| 429  | เรียกถี่เกินไป (rate limit เช่นขอ OTP รัว ๆ) |
| 500  | เซิร์ฟเวอร์ผิดพลาด |

**Pagination / Sorting / Filtering** — ส่งผ่าน query string
```
?page=1&per_page=20&sort=-created_at&search=เสื้อ&category=tshirt
```

---

## 2. Auth — ล็อกอินแบบ Passwordless (Google + OTP)

Flow: ผู้ใช้ระบุอีเมล (พิมพ์เอง หรือดึงจาก Google) → ระบบส่ง OTP ไปที่อีเมล → ยืนยัน OTP → ได้ token

### 2.1 ขอ OTP ด้วยอีเมลโดยตรง 🔓
```
POST /auth/otp/request
```
Request:
```json
{ "email": "somchai@gmail.com" }
```
Response `200`:
```json
{
  "success": true,
  "data": {
    "email": "somchai@gmail.com",
    "expires_in": 300,
    "message": "ส่งรหัส OTP ไปที่อีเมลแล้ว"
  }
}
```
> มี rate limit ป้องกันการ spam (เช่น ขอได้ 1 ครั้ง/60 วินาที, 5 ครั้ง/ชม.) → เกินจะได้ `429`

### 2.2 เข้าสู่ระบบด้วย Google (ดึงอีเมล + โปรไฟล์ แล้วส่ง OTP ต่อ) 🔓
```
POST /auth/google
```
Frontend ทำ Google Sign-In แล้วส่ง `id_token` มา:
```json
{ "id_token": "eyJhbGciOi..." }
```
Backend จะ verify token กับ Google, ดึงอีเมล/ชื่อ/รูป, สร้าง user ถ้ายังไม่มี แล้วส่ง OTP ไปที่อีเมลนั้น
Response `200`:
```json
{
  "success": true,
  "data": {
    "email": "somchai@gmail.com",
    "display_name": "สมชาย ใจดี",
    "expires_in": 300,
    "message": "ยืนยันตัวตนด้วย Google แล้ว ส่งรหัส OTP ไปที่อีเมล"
  }
}
```

### 2.3 ยืนยัน OTP → รับ token 🔓
```
POST /auth/otp/verify
```
Request:
```json
{ "email": "somchai@gmail.com", "code": "482913" }
```
Response `200`:
```json
{
  "success": true,
  "data": {
    "access_token": "eyJhbGciOi...",
    "refresh_token": "def502...",
    "token_type": "Bearer",
    "expires_in": 900,
    "user": {
      "public_id": "550e8400-e29b-41d4-a716-446655440000",
      "email": "somchai@gmail.com",
      "display_name": "สมชาย ใจดี",
      "avatar_url": "https://...",
      "role": "customer"
    }
  }
}
```
Error `400` — `code: INVALID_OTP` / `OTP_EXPIRED` / `TOO_MANY_ATTEMPTS`

### 2.4 ต่ออายุ access token 🔓
```
POST /auth/refresh
```
```json
{ "refresh_token": "def502..." }
```
Response `200` → คืน `access_token` ใหม่ (และ rotate `refresh_token`)

### 2.5 ออกจากระบบ 🔒
```
POST /auth/logout
```
เพิกถอน refresh token/session ปัจจุบัน → `204 No Content`

### 2.6 ดูข้อมูลผู้ใช้ที่ล็อกอินอยู่ 🔒
```
GET /auth/me
```
Response `200` → object `user` เหมือนใน 2.3

---

## 3. Profile — โปรไฟล์ผู้ใช้

### 3.1 ดูโปรไฟล์ 🔒
```
GET /me
```

### 3.2 แก้ไขโปรไฟล์ 🔒
```
PATCH /me
```
```json
{ "display_name": "สมชาย", "phone": "0812345678" }
```

---

## 4. Categories — หมวดหมู่สินค้า

### 4.1 ดูหมวดหมู่ทั้งหมด 🔓
```
GET /categories
```
Response `200`:
```json
{
  "success": true,
  "data": [
    { "id": 1, "name": "เสื้อยืด", "slug": "tshirt", "parent_id": null },
    { "id": 2, "name": "สติกเกอร์", "slug": "sticker", "parent_id": null }
  ]
}
```

### 4.2 จัดการหมวดหมู่ 🛡️
```
POST   /admin/categories
PATCH  /admin/categories/{id}
DELETE /admin/categories/{id}
```

---

## 5. Products — สินค้า

### 5.1 ดูรายการสินค้า (มี filter/search/paginate) 🔓
```
GET /products?category=tshirt&search=ข่าวปด&sort=-created_at&page=1&per_page=20
```
Response `200`:
```json
{
  "success": true,
  "data": [
    {
      "id": 10,
      "name": "เสื้อยืดกองบัญชาการข่าวปด",
      "slug": "kbc-tshirt-01",
      "base_price": 299.00,
      "is_featured": true,
      "primary_image": "https://.../cover.jpg",
      "price_range": { "min": 299.00, "max": 349.00 },
      "in_stock": true,
      "category": { "id": 1, "name": "เสื้อยืด", "slug": "tshirt" }
    }
  ],
  "meta": { "page": 1, "per_page": 20, "total": 42, "total_pages": 3 }
}
```
หมายเหตุ:
- ค่าเงินทุกฟิลด์ (`base_price`, `price_range.min/max`, `variants[].price`) เป็น **string** (serialize จาก decimal เพื่อคงความแม่นยำ ไม่ใช้ float) — ฝั่ง client แปลงเป็นตัวเลขเองตอนคำนวณ
- `category` = filter ด้วย **slug** ของหมวดหมู่ (เช่น `?category=tshirt`); ในผลลัพธ์เป็น object `{id,name,slug}` หรือ `null` ถ้าสินค้าไม่ถูกจัดหมวด
- `sort` รับเฉพาะ `created_at` / `base_price` / `name` (นำหน้าด้วย `-` = จากมากไปน้อย) ค่าอื่นถูกมองข้ามและใช้ค่าเริ่มต้น `-created_at`
- `price_range` / `in_stock` คำนวณจาก variant ที่ยัง active; แสดงเฉพาะสินค้าที่ `is_active = true`

### 5.2 ดูรายละเอียดสินค้า (พร้อม variants + รูปทั้งหมด) 🔓
```
GET /products/{slug}
```
> `variant_name` = ไซซ์ (เช่น "ไซซ์ M"), `color` = สี (เช่น "ขาว"/"ดำ") แยกฟิลด์กัน — `color` เป็น optional (ไม่ส่งมาถ้าสินค้าไม่มีตัวเลือกสี) ฝั่ง UI ใช้เลือกไซซ์กับสีเป็นคนละ selector

Response `200`:
```json
{
  "success": true,
  "data": {
    "id": 10,
    "name": "เสื้อยืดกองบัญชาการข่าวปด",
    "slug": "kbc-tshirt-01",
    "base_price": 299.00,
    "is_featured": true,
    "primary_image": "https://.../cover.jpg",
    "price_range": { "min": 299.00, "max": 349.00 },
    "in_stock": true,
    "category": { "id": 1, "name": "เสื้อยืด", "slug": "tshirt" },
    "description": "เสื้อผ้าฝ้าย 100%...",
    "images": [
      { "id": 1, "url": "https://.../1.jpg", "is_primary": true, "sort_order": 0 },
      { "id": 2, "url": "https://.../2.jpg", "is_primary": false, "sort_order": 1 }
    ],
    "variants": [
      { "id": 100, "variant_name": "ไซซ์ M", "color": "ดำ", "price": 299.00, "stock_quantity": 12, "sku": "TS01-M-BK" },
      { "id": 101, "variant_name": "ไซซ์ L", "color": "ดำ", "price": 349.00, "stock_quantity": 0,  "sku": "TS01-L-BK" }
    ]
  }
}
```

### 5.3 จัดการสินค้า / ตัวเลือก / รูป 🛡️
```
POST   /admin/products
PATCH  /admin/products/{id}
DELETE /admin/products/{id}

POST   /admin/products/{id}/variants
PATCH  /admin/variants/{id}
DELETE /admin/variants/{id}

POST   /admin/products/{id}/images        (multipart/form-data อัปรูป)
DELETE /admin/images/{id}
```

---

## 6. Cart — ตะกร้าสินค้า

ทุก endpoint ต้องล็อกอิน 🔒 (ระบบอ้างอิงตะกร้า `active` ของผู้ใช้อัตโนมัติ)

### 6.1 ดูตะกร้าปัจจุบัน
```
GET /cart
```
Response `200`:
```json
{
  "success": true,
  "data": {
    "id": 5,
    "items": [
      {
        "id": 50,
        "variant_id": 100,
        "product_name": "เสื้อยืดกองบัญชาการข่าวปด",
        "variant_name": "ไซซ์ M / สีดำ",
        "unit_price": 299.00,
        "quantity": 2,
        "line_total": 598.00,
        "image": "https://.../1.jpg",
        "in_stock": true
      }
    ],
    "subtotal": 598.00,
    "item_count": 2
  }
}
```

### 6.2 เพิ่มสินค้าลงตะกร้า
```
POST /cart/items
```
```json
{ "variant_id": 100, "quantity": 1 }
```
> ถ้ามี variant นี้ในตะกร้าอยู่แล้ว จะบวกจำนวนเพิ่ม (ไม่สร้างแถวใหม่)
> ถ้าสต็อกไม่พอ → `409 OUT_OF_STOCK`

### 6.3 แก้จำนวนสินค้าในตะกร้า
```
PATCH /cart/items/{itemId}
```
```json
{ "quantity": 3 }
```

### 6.4 ลบสินค้าออกจากตะกร้า
```
DELETE /cart/items/{itemId}
```

### 6.5 ล้างตะกร้าทั้งหมด
```
DELETE /cart
```

---

## 7. Addresses — ที่อยู่จัดส่ง

ทุก endpoint ต้องล็อกอิน 🔒

### 7.1 ดูที่อยู่ทั้งหมด
```
GET /addresses
```

### 7.2 เพิ่มที่อยู่
```
POST /addresses
```
```json
{
  "recipient_name": "สมชาย ใจดี",
  "phone": "0812345678",
  "address_line": "99/1 หมู่ 4 ถ.มิตรภาพ",
  "subdistrict": "ในเมือง",
  "district": "เมืองขอนแก่น",
  "province": "ขอนแก่น",
  "postal_code": "40000",
  "note": "ฝากไว้ที่ร้านหน้าปากซอย",
  "is_default": true
}
```
> `postal_code` ต้องเป็นตัวเลข 5 หลัก ไม่งั้น `422`

### 7.3 ดู / แก้ไข / ลบ ที่อยู่
```
GET    /addresses/{id}
PATCH  /addresses/{id}
DELETE /addresses/{id}
```

### 7.4 ตั้งเป็นที่อยู่หลัก
```
POST /addresses/{id}/default
```
> ระบบจะยกเลิก default ของที่อยู่อื่นให้อัตโนมัติ

---

## 8. Coupons — คูปองส่วนลด (ทางเลือก)

### 8.1 ตรวจสอบคูปองก่อนใช้ 🔒
```
POST /coupons/validate
```
```json
{ "code": "NEWS10", "subtotal": 598.00 }
```
Response `200`:
```json
{
  "success": true,
  "data": {
    "code": "NEWS10",
    "discount_type": "percent",
    "discount_value": 10,
    "discount_amount": 59.80,
    "valid": true
  }
}
```
Error `422` — `COUPON_EXPIRED` / `COUPON_MIN_ORDER_NOT_MET` / `COUPON_USAGE_LIMIT`

### 8.2 จัดการคูปอง 🛡️
```
GET    /admin/coupons
POST   /admin/coupons
PATCH  /admin/coupons/{id}
DELETE /admin/coupons/{id}
```

---

## 9. Orders — คำสั่งซื้อ

### 9.1 สร้างคำสั่งซื้อ (checkout จากตะกร้า) 🔒
```
POST /orders
```
Request:
```json
{
  "address_id": 3,
  "payment_method": "promptpay",
  "coupon_code": "NEWS10",
  "customer_note": "รบกวนส่งเร็วครับ"
}
```
Backend จะทำใน transaction เดียว: ตรวจสต็อก → ตัดสต็อก → คัดลอกที่อยู่+สินค้าเป็น snapshot → สร้าง order + order_items → เปลี่ยนตะกร้าเป็น `converted`
Response `201`:
```json
{
  "success": true,
  "data": {
    "order_number": "ORD-20260703-000042",
    "status": "pending",
    "payment_status": "unpaid",
    "payment_method": "promptpay",
    "subtotal": 598.00,
    "discount_amount": 59.80,
    "shipping_fee": 40.00,
    "total_amount": 578.20,
    "items": [
      { "product_name": "เสื้อยืดกองบัญชาการข่าวปด", "variant_name": "ไซซ์ M / สีดำ", "unit_price": 299.00, "quantity": 2, "line_total": 598.00 }
    ],
    "shipping_address": {
      "recipient": "สมชาย ใจดี", "phone": "0812345678",
      "address": "99/1 หมู่ 4 ถ.มิตรภาพ", "subdistrict": "ในเมือง",
      "district": "เมืองขอนแก่น", "province": "ขอนแก่น", "postal_code": "40000"
    },
    "placed_at": "2026-07-03T10:15:00+07:00"
  }
}
```
Error `409` — `OUT_OF_STOCK` (แนบ variant ที่ของหมดมาด้วย) / `CART_EMPTY`

### 9.2 ดูรายการคำสั่งซื้อของตัวเอง 🔒
```
GET /orders?status=shipped&page=1
```

### 9.3 ดูรายละเอียดคำสั่งซื้อ 🔒
```
GET /orders/{orderNumber}
```
คืนข้อมูลออเดอร์เต็ม รวม `payment`, `shipment`, และ `status_history`

### 9.4 ยกเลิกคำสั่งซื้อ 🔒
```
POST /orders/{orderNumber}/cancel
```
```json
{ "reason": "สั่งผิดรายการ" }
```
> ยกเลิกได้เฉพาะตอนสถานะยังเป็น `pending` / `paid` (ก่อนจัดส่ง) และระบบจะคืนสต็อกให้

---

## 10. Payments — การชำระเงิน

### 10.1 แจ้งชำระเงิน / อัปโหลดสลิป 🔒
```
POST /orders/{orderNumber}/payment
```
`multipart/form-data`:
```
method: bank_transfer
amount: 578.20
transaction_ref: (optional)
slip: <ไฟล์รูปสลิป>
```
Response `201` → payment สถานะ `pending_review` รอแอดมินตรวจ

### 10.2 ดูสถานะการชำระเงินของออเดอร์ 🔒
```
GET /orders/{orderNumber}/payment
```

### 10.3 (PromptPay) ขอ QR สำหรับจ่าย 🔒
```
GET /orders/{orderNumber}/payment/qr
```
Response → payload/รูป QR PromptPay ตามยอด `total_amount`

---

## 11. Admin — จัดการหลังบ้าน 🛡️

### 11.1 ดูคำสั่งซื้อทั้งหมด
```
GET /admin/orders?status=paid&search=ORD-2026&page=1
```

### 11.2 อัปเดตสถานะคำสั่งซื้อ
```
PATCH /admin/orders/{orderNumber}/status
```
```json
{ "status": "preparing", "note": "เริ่มแพ็กของ" }
```
> บันทึกลง `order_status_history` พร้อม admin ที่แก้ให้อัตโนมัติ

### 11.3 ยืนยัน/ปฏิเสธการชำระเงิน
```
PATCH /admin/payments/{id}/verify
```
```json
{ "status": "paid" }        // หรือ "failed"
```

### 11.4 เพิ่มเลขพัสดุ (จัดส่ง)
```
POST /admin/orders/{orderNumber}/shipment
```
```json
{ "courier": "Flash", "tracking_number": "TH123456789", "shipped_at": "2026-07-04T09:00:00+07:00" }
```
> เมื่อเพิ่มเลขพัสดุ ระบบเปลี่ยนสถานะออเดอร์เป็น `shipped` ให้อัตโนมัติ

### 11.5 แดชบอร์ดสรุป (ทางเลือก)
```
GET /admin/dashboard/summary?from=2026-07-01&to=2026-07-31
```
คืนยอดขาย, จำนวนออเดอร์แยกตามสถานะ, สินค้าขายดี ฯลฯ

---

## 12. สรุปตาราง Endpoint ทั้งหมด

| หมวด | Method | Path | สิทธิ์ |
|------|--------|------|--------|
| Auth | POST | /auth/otp/request | 🔓 |
| Auth | POST | /auth/google | 🔓 |
| Auth | POST | /auth/otp/verify | 🔓 |
| Auth | POST | /auth/refresh | 🔓 |
| Auth | POST | /auth/logout | 🔒 |
| Auth | GET | /auth/me | 🔒 |
| Profile | GET | /me | 🔒 |
| Profile | PATCH | /me | 🔒 |
| Category | GET | /categories | 🔓 |
| Category | POST/PATCH/DELETE | /admin/categories | 🛡️ |
| Product | GET | /products | 🔓 |
| Product | GET | /products/{slug} | 🔓 |
| Product | POST/PATCH/DELETE | /admin/products | 🛡️ |
| Variant/Image | POST/PATCH/DELETE | /admin/... | 🛡️ |
| Cart | GET | /cart | 🔒 |
| Cart | POST | /cart/items | 🔒 |
| Cart | PATCH | /cart/items/{id} | 🔒 |
| Cart | DELETE | /cart/items/{id} | 🔒 |
| Cart | DELETE | /cart | 🔒 |
| Address | GET | /addresses | 🔒 |
| Address | POST | /addresses | 🔒 |
| Address | GET/PATCH/DELETE | /addresses/{id} | 🔒 |
| Address | POST | /addresses/{id}/default | 🔒 |
| Coupon | POST | /coupons/validate | 🔒 |
| Coupon | */admin/coupons | 🛡️ |
| Order | POST | /orders | 🔒 |
| Order | GET | /orders | 🔒 |
| Order | GET | /orders/{orderNumber} | 🔒 |
| Order | POST | /orders/{orderNumber}/cancel | 🔒 |
| Payment | POST | /orders/{orderNumber}/payment | 🔒 |
| Payment | GET | /orders/{orderNumber}/payment | 🔒 |
| Payment | GET | /orders/{orderNumber}/payment/qr | 🔒 |
| Admin | GET | /admin/orders | 🛡️ |
| Admin | PATCH | /admin/orders/{orderNumber}/status | 🛡️ |
| Admin | PATCH | /admin/payments/{id}/verify | 🛡️ |
| Admin | POST | /admin/orders/{orderNumber}/shipment | 🛡️ |
| Admin | GET | /admin/dashboard/summary | 🛡️ |

---

## 13. หมายเหตุด้านความปลอดภัยที่ควรทำ

- **Rate limit** endpoint `/auth/otp/request` และ `/auth/otp/verify` อย่างเข้มงวด (ทั้งต่ออีเมลและต่อ IP) เพื่อกันการ brute-force OTP
- **OTP** ตั้งอายุสั้น (5–10 นาที) ใช้ได้ครั้งเดียว และล็อกหลังกรอกผิดเกิน `max_attempts`
- **ตัดสต็อก** ต้องทำใน DB transaction เดียวกับการสร้างออเดอร์ (`SELECT ... FOR UPDATE`) เพื่อกัน oversell ตอนคนแย่งซื้อพร้อมกัน
- **ตรวจสิทธิ์ความเป็นเจ้าของ** ทุก endpoint ที่มี `{id}` (order, address) ต้องเช็คว่าเป็นของ user ที่ล็อกอินอยู่จริง ไม่งั้นเสี่ยง IDOR
- **refresh token** เก็บใน httpOnly + Secure cookie และทำ token rotation ทุกครั้งที่ refresh
- **อัปโหลดสลิป** จำกัดชนิดไฟล์/ขนาด และสแกนก่อนเก็บ