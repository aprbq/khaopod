-- ============================================================================
--  Dev seed — ข้อมูลสินค้าตัวอย่าง (ตรงกับ mock เดิมฝั่ง frontend)
--  ไม่ใช่ migration: รันเองตอน dev เท่านั้น อย่ารันบน production
--    psql "$DATABASE_URL" -f migrations/seed_products.sql
--  รันซ้ำได้ (idempotent) — อิง slug/sku เป็น key
-- ============================================================================

INSERT INTO categories (name, slug) VALUES
    ('เสื้อยืด',  'tshirt'),
    ('ฮู้ดดี้',   'hoodie'),
    ('สติกเกอร์', 'sticker'),
    ('หมวก',      'hat'),
    ('กระเป๋า',   'bag'),
    ('ของสะสม',   'collectible')
ON CONFLICT (slug) DO NOTHING;

-- products — ผูกหมวดหมู่ผ่าน subquery ตาม slug
INSERT INTO products (category_id, name, slug, base_price, is_featured, description) VALUES
    ((SELECT id FROM categories WHERE slug = 'tshirt'),      'เสื้อยืดกองบัญชาการข่าวปด',        'kbc-tee-hq',              390, TRUE,  'สินค้าอย่างเป็นทางการจากเพจกองบัญชาการข่าวปด ผ้าฝ้ายเนื้อดี ใส่สบาย ดีไซน์สตรีทแวร์มินิมอล'),
    ((SELECT id FROM categories WHERE slug = 'tshirt'),      'เสื้อโอเวอร์ไซซ์ BREAKING NEWS',    'breaking-news-oversized', 590, TRUE,  'เสื้อโอเวอร์ไซซ์ทรงสตรีท งานพิมพ์คมชัด'),
    ((SELECT id FROM categories WHERE slug = 'hoodie'),      'ฮู้ดดี้ ข่าวปด CLASSIC',            'khaopod-hoodie-classic',  890, FALSE, 'ฮู้ดดี้ผ้าหนานุ่ม ใส่ได้ทุกฤดู'),
    ((SELECT id FROM categories WHERE slug = 'sticker'),     'สติกเกอร์เซ็ต ข่าวเด่นประจำวัน',    'sticker-daily-set',       120, FALSE, 'สติกเกอร์กันน้ำ 1 เซ็ต'),
    ((SELECT id FROM categories WHERE slug = 'hat'),         'หมวกบักเก็ต PRESS',                'press-bucket-hat',        450, TRUE,  'หมวกบักเก็ตปัก PRESS'),
    ((SELECT id FROM categories WHERE slug = 'bag'),         'กระเป๋าผ้า TOTE ข่าวปด',           'khaopod-tote',            290, FALSE, 'กระเป๋าผ้าแคนวาสหนา'),
    ((SELECT id FROM categories WHERE slug = 'collectible'), 'เข็มกลัด ENAMEL PIN นักข่าว',       'reporter-enamel-pin',     150, FALSE, 'เข็มกลัดอีนาเมลงานสะสม'),
    ((SELECT id FROM categories WHERE slug = 'tshirt'),      'เสื้อยืด LIMITED — ฉบับพิเศษ',       'limited-special-tee',     690, TRUE,  'รุ่นพิเศษจำนวนจำกัด')
ON CONFLICT (slug) DO NOTHING;

-- variants — ไซซ์ M/L/XL ต่อสินค้า (XL หมดสต็อก) เพื่อให้ price_range/in_stock มีความหลากหลาย
INSERT INTO product_variants (product_id, sku, variant_name, price, stock_quantity)
SELECT p.id, p.slug || '-M', 'ไซซ์ M', p.base_price,        12 FROM products p
ON CONFLICT (sku) DO NOTHING;
INSERT INTO product_variants (product_id, sku, variant_name, price, stock_quantity)
SELECT p.id, p.slug || '-L', 'ไซซ์ L', p.base_price + 30,    5 FROM products p
ON CONFLICT (sku) DO NOTHING;
INSERT INTO product_variants (product_id, sku, variant_name, price, stock_quantity)
SELECT p.id, p.slug || '-XL', 'ไซซ์ XL', p.base_price + 30,  0 FROM products p
ON CONFLICT (sku) DO NOTHING;

-- product_images — รูปสินค้า (ไฟล์อยู่ใน backend/migrations/image เสิร์ฟผ่าน /images)
-- รูปทั้ง 12 เป็นเสื้อตัวเดียวกันหลายมุม → ผูกทั้งหมดกับ 'kbc-tee-hq' เป็นแกลเลอรีเดียว
-- รูปแรก (yan01) เป็นรูปปก (is_primary), ที่เหลือเรียงตาม sort_order
-- reset ก่อน insert เพื่อให้ seed เป็น deterministic (รันซ้ำได้ผลเหมือนเดิม)
DELETE FROM product_images;
INSERT INTO product_images (product_id, url, is_primary, sort_order)
SELECT p.id, '/images/' || i.fname, (i.ord = 0), i.ord
FROM products p
CROSS JOIN (VALUES
    ('khaopod-yan01.jpg', 0),
    ('khaopod-yan02.jpg', 1),
    ('khaopod-yan03.jpg', 2),
    ('khaopod-yan04.jpg', 3),
    ('khaopod-yan05.jpg', 4),
    ('khaopod-yan06.jpg', 5),
    ('khaopod-yan07.jpg', 6),
    ('khaopod-yan08.jpg', 7),
    ('khaopod-yan09.jpg', 8),
    ('khaopod-yan10.jpg', 9),
    ('khaopod-yan11.jpg', 10),
    ('khaopod-yan12.jpg', 11)
) AS i(fname, ord)
WHERE p.slug = 'kbc-tee-hq';
