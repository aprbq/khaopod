-- ที่อยู่จัดส่ง + คำสั่งซื้อ + ชำระเงิน (addresses / orders / order_items / order_status_history / payments)

DO $$ BEGIN
    CREATE TYPE order_status AS ENUM (
        'pending', 'paid', 'preparing', 'shipped', 'delivered', 'completed', 'cancelled', 'refunded'
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
    CREATE TYPE payment_status AS ENUM ('unpaid', 'pending_review', 'paid', 'failed', 'refunded');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
    CREATE TYPE payment_method AS ENUM ('promptpay', 'bank_transfer', 'cod', 'credit_card');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

-- addresses — ที่อยู่ของผู้ใช้ (โครงสร้างแบบไทย)
CREATE TABLE IF NOT EXISTS addresses (
    id             BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id        BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    recipient_name TEXT        NOT NULL,
    phone          TEXT        NOT NULL,
    address_line   TEXT        NOT NULL,
    subdistrict    TEXT        NOT NULL,
    district       TEXT        NOT NULL,
    province       TEXT        NOT NULL,
    postal_code    TEXT        NOT NULL,
    country        TEXT        NOT NULL DEFAULT 'TH',
    note           TEXT,
    is_default     BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_postal_code CHECK (postal_code ~ '^[0-9]{5}$')
);

CREATE INDEX IF NOT EXISTS idx_addresses_user ON addresses(user_id);
-- ที่อยู่หลักได้เพียง 1 ที่ต่อผู้ใช้
CREATE UNIQUE INDEX IF NOT EXISTS uq_default_address_per_user
    ON addresses(user_id) WHERE is_default = TRUE;

DROP TRIGGER IF EXISTS trg_addresses_updated ON addresses;
CREATE TRIGGER trg_addresses_updated
    BEFORE UPDATE ON addresses
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ตัวสร้างเลขออเดอร์แบบอ่านง่าย เช่น ORD-20260707-000001
CREATE SEQUENCE IF NOT EXISTS order_number_seq;

CREATE OR REPLACE FUNCTION generate_order_number()
RETURNS TEXT AS $$
BEGIN
    RETURN 'ORD-' || to_char(now(), 'YYYYMMDD') || '-'
           || lpad(nextval('order_number_seq')::text, 6, '0');
END;
$$ LANGUAGE plpgsql;

-- orders — เก็บที่อยู่จัดส่งเป็น snapshot ในตัวออเดอร์ เผื่อผู้ใช้แก้/ลบที่อยู่ทีหลัง
CREATE TABLE IF NOT EXISTS orders (
    id               BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    order_number     TEXT           NOT NULL DEFAULT generate_order_number(),
    user_id          BIGINT         NOT NULL REFERENCES users(id) ON DELETE RESTRICT,

    subtotal         NUMERIC(12,2)  NOT NULL DEFAULT 0 CHECK (subtotal >= 0),
    shipping_fee     NUMERIC(12,2)  NOT NULL DEFAULT 0 CHECK (shipping_fee >= 0),
    discount_amount  NUMERIC(12,2)  NOT NULL DEFAULT 0 CHECK (discount_amount >= 0),
    total_amount     NUMERIC(12,2)  NOT NULL DEFAULT 0 CHECK (total_amount >= 0),

    status           order_status   NOT NULL DEFAULT 'pending',
    payment_status   payment_status NOT NULL DEFAULT 'unpaid',
    payment_method   payment_method,

    ship_recipient   TEXT           NOT NULL,
    ship_phone       TEXT           NOT NULL,
    ship_address     TEXT           NOT NULL,
    ship_subdistrict TEXT           NOT NULL,
    ship_district    TEXT           NOT NULL,
    ship_province    TEXT           NOT NULL,
    ship_postal_code TEXT           NOT NULL,
    ship_country     TEXT           NOT NULL DEFAULT 'TH',

    customer_note    TEXT,
    admin_note       TEXT,
    placed_at        TIMESTAMPTZ    NOT NULL DEFAULT now(),
    created_at       TIMESTAMPTZ    NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ    NOT NULL DEFAULT now(),

    CONSTRAINT uq_order_number UNIQUE (order_number)
);

CREATE INDEX IF NOT EXISTS idx_orders_user   ON orders(user_id);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
CREATE INDEX IF NOT EXISTS idx_orders_placed ON orders(placed_at DESC);

DROP TRIGGER IF EXISTS trg_orders_updated ON orders;
CREATE TRIGGER trg_orders_updated
    BEFORE UPDATE ON orders
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- order_items — snapshot ชื่อ/ราคา ณ ตอนซื้อ เพราะราคาสินค้าอาจเปลี่ยนภายหลัง
CREATE TABLE IF NOT EXISTS order_items (
    id                 BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    order_id           BIGINT        NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_variant_id BIGINT        REFERENCES product_variants(id) ON DELETE SET NULL,
    product_name       TEXT          NOT NULL,
    variant_name       TEXT          NOT NULL,
    unit_price         NUMERIC(12,2) NOT NULL CHECK (unit_price >= 0),
    quantity           INT           NOT NULL CHECK (quantity > 0),
    line_total         NUMERIC(12,2) NOT NULL CHECK (line_total >= 0),
    created_at         TIMESTAMPTZ   NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_order_items_order ON order_items(order_id);

-- order_status_history — audit trail ของการเปลี่ยนสถานะ
CREATE TABLE IF NOT EXISTS order_status_history (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    order_id    BIGINT       NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    status      order_status NOT NULL,
    note        TEXT,
    changed_by  BIGINT       REFERENCES users(id) ON DELETE SET NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_status_history_order ON order_status_history(order_id);

-- payments — การชำระเงิน (โอน/พร้อมเพย์ + อัปสลิป)
CREATE TABLE IF NOT EXISTS payments (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    order_id        BIGINT         NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    method          payment_method NOT NULL,
    amount          NUMERIC(12,2)  NOT NULL CHECK (amount >= 0),
    status          payment_status NOT NULL DEFAULT 'pending_review',
    slip_url        TEXT,
    transaction_ref TEXT,
    paid_at         TIMESTAMPTZ,
    verified_by     BIGINT         REFERENCES users(id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ    NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ    NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_payments_order ON payments(order_id);

DROP TRIGGER IF EXISTS trg_payments_updated ON payments;
CREATE TRIGGER trg_payments_updated
    BEFORE UPDATE ON payments
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
