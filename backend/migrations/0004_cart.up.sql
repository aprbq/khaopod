-- ตะกร้าสินค้า (carts) + รายการในตะกร้า (cart_items)

DO $$ BEGIN
    CREATE TYPE cart_status AS ENUM ('active', 'converted', 'abandoned');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

-- carts — 1 ผู้ใช้มีตะกร้า active ได้ตัวเดียว (บังคับด้วย partial unique index)
CREATE TABLE IF NOT EXISTS carts (
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id    BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status     cart_status NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_active_cart_per_user
    ON carts(user_id) WHERE status = 'active';

DROP TRIGGER IF EXISTS trg_carts_updated ON carts;
CREATE TRIGGER trg_carts_updated
    BEFORE UPDATE ON carts
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- cart_items — สินค้าเดิมในตะกร้าเดียวกันรวมจำนวน (unique cart_id+variant)
CREATE TABLE IF NOT EXISTS cart_items (
    id                 BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    cart_id            BIGINT      NOT NULL REFERENCES carts(id) ON DELETE CASCADE,
    product_variant_id BIGINT      NOT NULL REFERENCES product_variants(id) ON DELETE RESTRICT,
    quantity           INT         NOT NULL CHECK (quantity > 0),
    added_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_cart_variant UNIQUE (cart_id, product_variant_id)
);

CREATE INDEX IF NOT EXISTS idx_cart_items_cart ON cart_items(cart_id);
