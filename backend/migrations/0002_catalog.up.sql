-- ============================================================================
--  Migration 0002 — Catalog (ยกมาจาก docs/schema.sql เฉพาะส่วน PRODUCTS)
--  categories / products / product_variants / product_images
-- ============================================================================

-- categories — หมวดหมู่สินค้า (รองรับหมวดย่อยแบบ tree)
CREATE TABLE IF NOT EXISTS categories (
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

DROP TRIGGER IF EXISTS trg_categories_updated ON categories;
CREATE TRIGGER trg_categories_updated
    BEFORE UPDATE ON categories
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- products — สินค้าหลัก (ราคา/สต็อกจริงอยู่ที่ product_variants)
CREATE TABLE IF NOT EXISTS products (
    id            BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    category_id   BIGINT        REFERENCES categories(id) ON DELETE SET NULL,
    name          TEXT          NOT NULL,
    slug          TEXT          NOT NULL,
    description   TEXT,
    base_price    NUMERIC(12,2) NOT NULL DEFAULT 0 CHECK (base_price >= 0),
    is_active     BOOLEAN       NOT NULL DEFAULT TRUE,
    is_featured   BOOLEAN       NOT NULL DEFAULT FALSE,
    created_at    TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ   NOT NULL DEFAULT now(),

    CONSTRAINT uq_product_slug UNIQUE (slug)
);

CREATE INDEX IF NOT EXISTS idx_products_category ON products(category_id);
CREATE INDEX IF NOT EXISTS idx_products_active   ON products(is_active) WHERE is_active = TRUE;

DROP TRIGGER IF EXISTS trg_products_updated ON products;
CREATE TRIGGER trg_products_updated
    BEFORE UPDATE ON products
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- product_variants — ตัวเลือกสินค้า (ไซซ์/สี) + ราคา + สต็อก
CREATE TABLE IF NOT EXISTS product_variants (
    id             BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    product_id     BIGINT        NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    sku            TEXT,
    variant_name   TEXT          NOT NULL DEFAULT 'ค่าเริ่มต้น',
    price          NUMERIC(12,2) NOT NULL CHECK (price >= 0),
    stock_quantity INT           NOT NULL DEFAULT 0 CHECK (stock_quantity >= 0),
    weight_grams   INT           CHECK (weight_grams >= 0),
    is_active      BOOLEAN       NOT NULL DEFAULT TRUE,
    created_at     TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ   NOT NULL DEFAULT now(),

    CONSTRAINT uq_variant_sku UNIQUE (sku)
);

CREATE INDEX IF NOT EXISTS idx_variants_product ON product_variants(product_id);

DROP TRIGGER IF EXISTS trg_variants_updated ON product_variants;
CREATE TRIGGER trg_variants_updated
    BEFORE UPDATE ON product_variants
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- product_images — รูปสินค้า (รูปปกได้ 1 รูปต่อสินค้า)
CREATE TABLE IF NOT EXISTS product_images (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    product_id  BIGINT      NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    url         TEXT        NOT NULL,
    alt_text    TEXT,
    is_primary  BOOLEAN     NOT NULL DEFAULT FALSE,
    sort_order  INT         NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_images_product ON product_images(product_id);
CREATE UNIQUE INDEX IF NOT EXISTS uq_primary_image_per_product
    ON product_images(product_id) WHERE is_primary = TRUE;
