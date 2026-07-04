-- ============================================================================
--  ระบบขายสินค้า "เพจกองบัญชาการข่าวปด" (KBC News Shop)
--  Database schema สำหรับ PostgreSQL 14+
--
--  Flow ล็อกอิน: ล็อกอินผ่าน Google/Gmail แบบไม่ต้องใส่รหัสผ่าน
--               ระบบจะส่ง OTP ไปที่อีเมลเพื่อยืนยันตัวตนแทน
--
--  วิธีใช้:  psql -U <user> -d <dbname> -f kbc_news_shop_schema.sql
-- ============================================================================


-- ---------------------------------------------------------------------------
-- 0) EXTENSIONS
-- ---------------------------------------------------------------------------
CREATE EXTENSION IF NOT EXISTS pgcrypto;   -- gen_random_uuid(), crypt(), digest()
CREATE EXTENSION IF NOT EXISTS citext;     -- อีเมลแบบ case-insensitive


-- ---------------------------------------------------------------------------
-- 1) ENUM TYPES  (สถานะต่าง ๆ ของระบบ)
-- ---------------------------------------------------------------------------
CREATE TYPE otp_purpose      AS ENUM ('login', 'verify_email', 'change_email');

CREATE TYPE order_status      AS ENUM (
    'pending',      -- รอยืนยัน/รอชำระเงิน
    'paid',         -- ชำระเงินแล้ว (รอตรวจสอบ/รอแพ็ก)
    'preparing',    -- กำลังจัดเตรียมสินค้า
    'shipped',      -- จัดส่งแล้ว
    'delivered',    -- ถึงมือลูกค้าแล้ว
    'completed',    -- ปิดออเดอร์สมบูรณ์
    'cancelled',    -- ยกเลิก
    'refunded'      -- คืนเงินแล้ว
);

CREATE TYPE payment_status    AS ENUM (
    'unpaid',          -- ยังไม่จ่าย
    'pending_review',  -- แจ้งโอน/อัปสลิปแล้ว รอแอดมินตรวจ
    'paid',            -- ยืนยันการชำระเงินแล้ว
    'failed',          -- ล้มเหลว
    'refunded'         -- คืนเงิน
);

CREATE TYPE payment_method    AS ENUM ('promptpay', 'bank_transfer', 'cod', 'credit_card');

CREATE TYPE cart_status       AS ENUM ('active', 'converted', 'abandoned');


-- ---------------------------------------------------------------------------
-- 2) TRIGGER FUNCTION สำหรับอัปเดต updated_at อัตโนมัติ
-- ---------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;


-- ===========================================================================
--  ส่วนที่ 1: ผู้ใช้ & การยืนยันตัวตน (AUTH)
-- ===========================================================================

-- ---------------------------------------------------------------------------
-- 3) users  — ผู้ใช้งาน
--    ไม่มีคอลัมน์ password เพราะเป็นระบบ passwordless (Google + OTP)
-- ---------------------------------------------------------------------------
CREATE TABLE users (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    public_id       UUID          NOT NULL DEFAULT gen_random_uuid(),  -- id สำหรับโชว์ภายนอก
    email           CITEXT        NOT NULL,
    email_verified  BOOLEAN       NOT NULL DEFAULT FALSE,   -- true เมื่อยืนยัน OTP แล้ว
    display_name    TEXT,
    avatar_url      TEXT,
    phone           TEXT,                                   -- เบอร์โทร (optional)
    role            TEXT          NOT NULL DEFAULT 'customer'  -- 'customer' | 'admin'
                    CHECK (role IN ('customer', 'admin')),
    is_active       BOOLEAN       NOT NULL DEFAULT TRUE,
    last_login_at   TIMESTAMPTZ,
    created_at      TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ   NOT NULL DEFAULT now(),

    CONSTRAINT uq_users_email     UNIQUE (email),
    CONSTRAINT uq_users_public_id UNIQUE (public_id)
);

CREATE TRIGGER trg_users_updated
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();


-- ---------------------------------------------------------------------------
-- 4) user_oauth_accounts — บัญชี OAuth ที่ผูกกับผู้ใช้ (Google)
--    แยกตารางเพื่อรองรับการผูกหลาย provider ในอนาคต
-- ---------------------------------------------------------------------------
CREATE TABLE user_oauth_accounts (
    id               BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id          BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider         TEXT        NOT NULL DEFAULT 'google',   -- 'google'
    provider_user_id TEXT        NOT NULL,                    -- Google "sub" (subject id)
    provider_email   CITEXT,
    raw_profile      JSONB,                                   -- เก็บ profile ดิบจาก Google
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),

    -- ห้ามมีบัญชี provider เดียวกัน + sub เดียวกันซ้ำ
    CONSTRAINT uq_oauth_provider_uid UNIQUE (provider, provider_user_id)
);

CREATE INDEX idx_oauth_user_id ON user_oauth_accounts(user_id);


-- ---------------------------------------------------------------------------
-- 5) otp_codes — รหัส OTP ที่ส่งไปทางอีเมล
--    เก็บเฉพาะ "แฮชของ OTP" ไม่เก็บเลขจริง (ปลอดภัยกว่า)
-- ---------------------------------------------------------------------------
CREATE TABLE otp_codes (
    id           BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id      BIGINT       REFERENCES users(id) ON DELETE CASCADE,  -- อาจ null ถ้ายังไม่มี user
    email        CITEXT       NOT NULL,          -- อีเมลปลายทางที่ส่ง OTP ไป
    purpose      otp_purpose  NOT NULL DEFAULT 'login',
    code_hash    TEXT         NOT NULL,          -- เช่น digest(code || secret, 'sha256')
    expires_at   TIMESTAMPTZ  NOT NULL,          -- ปกติ 5–10 นาที
    consumed_at  TIMESTAMPTZ,                    -- เวลาที่ถูกใช้สำเร็จ (null = ยังไม่ใช้)
    attempts     SMALLINT     NOT NULL DEFAULT 0,   -- จำนวนครั้งที่กรอกผิด
    max_attempts SMALLINT     NOT NULL DEFAULT 5,
    request_ip   INET,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),

    CONSTRAINT chk_otp_attempts CHECK (attempts >= 0)
);

CREATE INDEX idx_otp_email_purpose ON otp_codes(email, purpose);
CREATE INDEX idx_otp_expires        ON otp_codes(expires_at);
-- เร่งการหา OTP ที่ยัง "ใช้งานได้" (ยังไม่ถูกใช้)
CREATE INDEX idx_otp_active         ON otp_codes(email) WHERE consumed_at IS NULL;


-- ---------------------------------------------------------------------------
-- 6) auth_sessions — เซสชัน/refresh token หลังล็อกอินสำเร็จ
-- ---------------------------------------------------------------------------
CREATE TABLE auth_sessions (
    id                 BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id            BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token_hash TEXT        NOT NULL,        -- เก็บแฮชของ refresh token
    user_agent         TEXT,
    ip_address         INET,
    expires_at         TIMESTAMPTZ NOT NULL,
    revoked_at         TIMESTAMPTZ,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_session_token UNIQUE (refresh_token_hash)
);

CREATE INDEX idx_sessions_user ON auth_sessions(user_id);


-- ===========================================================================
--  ส่วนที่ 2: แคตตาล็อกสินค้า (PRODUCTS)
-- ===========================================================================

-- ---------------------------------------------------------------------------
-- 7) categories — หมวดหมู่สินค้า (รองรับหมวดย่อยแบบ tree)
-- ---------------------------------------------------------------------------
CREATE TABLE categories (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    parent_id   BIGINT      REFERENCES categories(id) ON DELETE SET NULL,
    name        TEXT        NOT NULL,
    slug        TEXT        NOT NULL,
    description TEXT,
    sort_order  INT         NOT NULL DEFAULT 0,
    is_active   BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_category_slug UNIQUE (slug)
);

CREATE TRIGGER trg_categories_updated
    BEFORE UPDATE ON categories
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();


-- ---------------------------------------------------------------------------
-- 8) products — สินค้าหลัก
--    ราค/สต็อกที่ขายจริงจะอยู่ที่ product_variants (ตัวเลือกย่อย)
-- ---------------------------------------------------------------------------
CREATE TABLE products (
    id            BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    category_id   BIGINT        REFERENCES categories(id) ON DELETE SET NULL,
    name          TEXT          NOT NULL,
    slug          TEXT          NOT NULL,
    description   TEXT,
    base_price    NUMERIC(12,2) NOT NULL DEFAULT 0    -- ราคาตั้งต้น/ราคาโชว์
                  CHECK (base_price >= 0),
    is_active     BOOLEAN       NOT NULL DEFAULT TRUE,  -- โชว์หน้าร้านหรือไม่
    is_featured   BOOLEAN       NOT NULL DEFAULT FALSE, -- สินค้าแนะนำ
    created_at    TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ   NOT NULL DEFAULT now(),

    CONSTRAINT uq_product_slug UNIQUE (slug)
);

CREATE INDEX idx_products_category ON products(category_id);
CREATE INDEX idx_products_active   ON products(is_active) WHERE is_active = TRUE;

CREATE TRIGGER trg_products_updated
    BEFORE UPDATE ON products
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();


-- ---------------------------------------------------------------------------
-- 9) product_variants — ตัวเลือกสินค้า (เช่น ไซซ์ / สี) + ราคา + สต็อก
--    ถ้าสินค้าไม่มีตัวเลือก ให้สร้าง 1 variant เป็นค่าเริ่มต้น
-- ---------------------------------------------------------------------------
CREATE TABLE product_variants (
    id             BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    product_id     BIGINT        NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    sku            TEXT,                              -- รหัสสินค้า (optional)
    variant_name   TEXT          NOT NULL DEFAULT 'ค่าเริ่มต้น',  -- ไซซ์ เช่น "ไซซ์ M"
    color          TEXT,                              -- สี เช่น "ขาว"/"ดำ"; NULL = ไม่มีตัวเลือกสี
    price          NUMERIC(12,2) NOT NULL CHECK (price >= 0),
    stock_quantity INT           NOT NULL DEFAULT 0 CHECK (stock_quantity >= 0),
    weight_grams   INT           CHECK (weight_grams >= 0),  -- ไว้คำนวณค่าส่ง
    is_active      BOOLEAN       NOT NULL DEFAULT TRUE,
    created_at     TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ   NOT NULL DEFAULT now(),

    CONSTRAINT uq_variant_sku UNIQUE (sku)   -- อนุญาต null ได้หลายตัว
);

CREATE INDEX idx_variants_product ON product_variants(product_id);

CREATE TRIGGER trg_variants_updated
    BEFORE UPDATE ON product_variants
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();


-- ---------------------------------------------------------------------------
-- 10) product_images — รูปสินค้า
-- ---------------------------------------------------------------------------
CREATE TABLE product_images (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    product_id  BIGINT      NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    url         TEXT        NOT NULL,
    alt_text    TEXT,
    is_primary  BOOLEAN     NOT NULL DEFAULT FALSE,   -- รูปปก
    sort_order  INT         NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_images_product ON product_images(product_id);
-- บังคับให้มีรูปปกได้เพียง 1 รูปต่อสินค้า
CREATE UNIQUE INDEX uq_primary_image_per_product
    ON product_images(product_id) WHERE is_primary = TRUE;


-- ===========================================================================
--  ส่วนที่ 3: ที่อยู่จัดส่ง (ADDRESSES)
-- ===========================================================================

-- ---------------------------------------------------------------------------
-- 11) addresses — ที่อยู่ของผู้ใช้ (โครงสร้างแบบไทย)
-- ---------------------------------------------------------------------------
CREATE TABLE addresses (
    id             BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id        BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    recipient_name TEXT        NOT NULL,               -- ชื่อผู้รับ
    phone          TEXT        NOT NULL,               -- เบอร์ผู้รับ
    address_line   TEXT        NOT NULL,               -- บ้านเลขที่ / หมู่ / ถนน / ซอย
    subdistrict    TEXT        NOT NULL,               -- ตำบล / แขวง
    district       TEXT        NOT NULL,               -- อำเภอ / เขต
    province       TEXT        NOT NULL,               -- จังหวัด
    postal_code    TEXT        NOT NULL,               -- รหัสไปรษณีย์
    country        TEXT        NOT NULL DEFAULT 'TH',
    note           TEXT,                               -- หมายเหตุการจัดส่ง
    is_default     BOOLEAN     NOT NULL DEFAULT FALSE, -- ที่อยู่หลัก
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_postal_code CHECK (postal_code ~ '^[0-9]{5}$')
);

CREATE INDEX idx_addresses_user ON addresses(user_id);
-- ที่อยู่หลักได้เพียง 1 ที่ต่อผู้ใช้
CREATE UNIQUE INDEX uq_default_address_per_user
    ON addresses(user_id) WHERE is_default = TRUE;

CREATE TRIGGER trg_addresses_updated
    BEFORE UPDATE ON addresses
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();


-- ===========================================================================
--  ส่วนที่ 4: ตะกร้าสินค้า (CART)
-- ===========================================================================

-- ---------------------------------------------------------------------------
-- 12) carts — ตะกร้า (1 ผู้ใช้ควรมีตะกร้า active ได้ตัวเดียว)
-- ---------------------------------------------------------------------------
CREATE TABLE carts (
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id    BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status     cart_status NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- บังคับ 1 ตะกร้า active ต่อผู้ใช้
CREATE UNIQUE INDEX uq_active_cart_per_user
    ON carts(user_id) WHERE status = 'active';

CREATE TRIGGER trg_carts_updated
    BEFORE UPDATE ON carts
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();


-- ---------------------------------------------------------------------------
-- 13) cart_items — รายการสินค้าในตะกร้า
-- ---------------------------------------------------------------------------
CREATE TABLE cart_items (
    id                 BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    cart_id            BIGINT      NOT NULL REFERENCES carts(id) ON DELETE CASCADE,
    product_variant_id BIGINT      NOT NULL REFERENCES product_variants(id) ON DELETE RESTRICT,
    quantity           INT         NOT NULL CHECK (quantity > 0),
    added_at           TIMESTAMPTZ NOT NULL DEFAULT now(),

    -- สินค้าเดิมในตะกร้าเดียวกัน = รวมจำนวน (ไม่สร้างแถวซ้ำ)
    CONSTRAINT uq_cart_variant UNIQUE (cart_id, product_variant_id)
);

CREATE INDEX idx_cart_items_cart ON cart_items(cart_id);


-- ===========================================================================
--  ส่วนที่ 5: คำสั่งซื้อ + ชำระเงิน + จัดส่ง (ORDERS)
-- ===========================================================================

-- ---------------------------------------------------------------------------
-- ตัวสร้างเลขออเดอร์แบบอ่านง่าย เช่น  ORD-20260702-000001
-- ---------------------------------------------------------------------------
CREATE SEQUENCE order_number_seq;

CREATE OR REPLACE FUNCTION generate_order_number()
RETURNS TEXT AS $$
BEGIN
    RETURN 'ORD-' || to_char(now(), 'YYYYMMDD') || '-'
           || lpad(nextval('order_number_seq')::text, 6, '0');
END;
$$ LANGUAGE plpgsql;


-- ---------------------------------------------------------------------------
-- 14) orders — คำสั่งซื้อ
--     เก็บที่อยู่จัดส่งแบบ "snapshot" ลงในออเดอร์เลย เผื่อผู้ใช้แก้/ลบที่อยู่ทีหลัง
-- ---------------------------------------------------------------------------
CREATE TABLE orders (
    id               BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    order_number     TEXT           NOT NULL DEFAULT generate_order_number(),
    user_id          BIGINT         NOT NULL REFERENCES users(id) ON DELETE RESTRICT,

    -- ยอดเงิน
    subtotal         NUMERIC(12,2)  NOT NULL DEFAULT 0 CHECK (subtotal >= 0),
    shipping_fee     NUMERIC(12,2)  NOT NULL DEFAULT 0 CHECK (shipping_fee >= 0),
    discount_amount  NUMERIC(12,2)  NOT NULL DEFAULT 0 CHECK (discount_amount >= 0),
    total_amount     NUMERIC(12,2)  NOT NULL DEFAULT 0 CHECK (total_amount >= 0),

    -- สถานะ
    status           order_status   NOT NULL DEFAULT 'pending',
    payment_status   payment_status NOT NULL DEFAULT 'unpaid',
    payment_method   payment_method,

    -- snapshot ที่อยู่จัดส่ง ณ ตอนสั่งซื้อ
    ship_recipient   TEXT           NOT NULL,
    ship_phone       TEXT           NOT NULL,
    ship_address     TEXT           NOT NULL,
    ship_subdistrict TEXT           NOT NULL,
    ship_district    TEXT           NOT NULL,
    ship_province    TEXT           NOT NULL,
    ship_postal_code TEXT           NOT NULL,
    ship_country     TEXT           NOT NULL DEFAULT 'TH',

    customer_note    TEXT,          -- โน้ตจากลูกค้า
    admin_note       TEXT,          -- โน้ตภายในของแอดมิน
    placed_at        TIMESTAMPTZ    NOT NULL DEFAULT now(),
    created_at       TIMESTAMPTZ    NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ    NOT NULL DEFAULT now(),

    CONSTRAINT uq_order_number UNIQUE (order_number)
);

CREATE INDEX idx_orders_user   ON orders(user_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_placed ON orders(placed_at DESC);

CREATE TRIGGER trg_orders_updated
    BEFORE UPDATE ON orders
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();


-- ---------------------------------------------------------------------------
-- 15) order_items — รายการสินค้าในคำสั่งซื้อ
--     เก็บ snapshot ชื่อ/ราคา ณ ตอนซื้อ เพราะราคาสินค้าอาจเปลี่ยนภายหลัง
-- ---------------------------------------------------------------------------
CREATE TABLE order_items (
    id                 BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    order_id           BIGINT        NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_variant_id BIGINT        REFERENCES product_variants(id) ON DELETE SET NULL,
    product_name       TEXT          NOT NULL,        -- snapshot ชื่อสินค้า
    variant_name       TEXT          NOT NULL,        -- snapshot ชื่อตัวเลือก
    unit_price         NUMERIC(12,2) NOT NULL CHECK (unit_price >= 0),  -- ราคา/ชิ้น ณ ตอนซื้อ
    quantity           INT           NOT NULL CHECK (quantity > 0),
    line_total         NUMERIC(12,2) NOT NULL CHECK (line_total >= 0),  -- unit_price * quantity
    created_at         TIMESTAMPTZ   NOT NULL DEFAULT now()
);

CREATE INDEX idx_order_items_order ON order_items(order_id);


-- ---------------------------------------------------------------------------
-- 16) order_status_history — ประวัติการเปลี่ยนสถานะออเดอร์ (audit trail)
-- ---------------------------------------------------------------------------
CREATE TABLE order_status_history (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    order_id    BIGINT       NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    status      order_status NOT NULL,
    note        TEXT,
    changed_by  BIGINT       REFERENCES users(id) ON DELETE SET NULL,  -- แอดมินที่แก้
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX idx_status_history_order ON order_status_history(order_id);


-- ---------------------------------------------------------------------------
-- 17) payments — การชำระเงิน (รองรับโอน/พร้อมเพย์/อัปสลิป)
-- ---------------------------------------------------------------------------
CREATE TABLE payments (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    order_id        BIGINT         NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    method          payment_method NOT NULL,
    amount          NUMERIC(12,2)  NOT NULL CHECK (amount >= 0),
    status          payment_status NOT NULL DEFAULT 'pending_review',
    slip_url        TEXT,                              -- รูปสลิปที่ลูกค้าอัป
    transaction_ref TEXT,                              -- เลขอ้างอิงธนาคาร/gateway
    paid_at         TIMESTAMPTZ,                       -- เวลาที่ยืนยันจ่ายจริง
    verified_by     BIGINT         REFERENCES users(id) ON DELETE SET NULL, -- แอดมินที่ตรวจสลิป
    created_at      TIMESTAMPTZ    NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ    NOT NULL DEFAULT now()
);

CREATE INDEX idx_payments_order ON payments(order_id);

CREATE TRIGGER trg_payments_updated
    BEFORE UPDATE ON payments
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();


-- ---------------------------------------------------------------------------
-- 18) shipments — ข้อมูลการจัดส่ง / เลขพัสดุ
-- ---------------------------------------------------------------------------
CREATE TABLE shipments (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    order_id        BIGINT      NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    courier         TEXT,                    -- เช่น 'Kerry', 'Flash', 'ไปรษณีย์ไทย', 'J&T'
    tracking_number TEXT,                    -- เลขติดตามพัสดุ
    shipped_at      TIMESTAMPTZ,
    delivered_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_shipments_order    ON shipments(order_id);
CREATE INDEX idx_shipments_tracking ON shipments(tracking_number);

CREATE TRIGGER trg_shipments_updated
    BEFORE UPDATE ON shipments
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();


-- ===========================================================================
--  (ทางเลือก) ส่วนที่ 6: คูปองส่วนลด — เผื่อทำโปรโมชัน
-- ===========================================================================

-- ---------------------------------------------------------------------------
-- 19) coupons — คูปอง/โค้ดส่วนลด
-- ---------------------------------------------------------------------------
CREATE TABLE coupons (
    id             BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    code           TEXT          NOT NULL,
    description    TEXT,
    discount_type  TEXT          NOT NULL CHECK (discount_type IN ('percent', 'fixed')),
    discount_value NUMERIC(12,2) NOT NULL CHECK (discount_value >= 0),
    min_order      NUMERIC(12,2) NOT NULL DEFAULT 0,   -- ยอดขั้นต่ำที่ใช้ได้
    max_uses       INT,                                -- จำกัดจำนวนครั้งทั้งหมด (null = ไม่จำกัด)
    used_count     INT           NOT NULL DEFAULT 0,
    starts_at      TIMESTAMPTZ,
    expires_at     TIMESTAMPTZ,
    is_active      BOOLEAN       NOT NULL DEFAULT TRUE,
    created_at     TIMESTAMPTZ   NOT NULL DEFAULT now(),

    CONSTRAINT uq_coupon_code UNIQUE (code)
);

-- ผูกคูปองกับออเดอร์ (บันทึกว่าออเดอร์นี้ใช้คูปองอะไร ลดไปเท่าไหร่)
ALTER TABLE orders
    ADD COLUMN coupon_id BIGINT REFERENCES coupons(id) ON DELETE SET NULL;


-- ===========================================================================
--  ข้อมูลตัวอย่าง (SEED) — ลบทิ้งได้ถ้าไม่ต้องการ
-- ===========================================================================
INSERT INTO categories (name, slug) VALUES
    ('เสื้อยืด',   'tshirt'),
    ('สติกเกอร์',  'sticker'),
    ('ของสะสม',    'collectible');

-- ============================================================================
--  จบ schema
-- ============================================================================