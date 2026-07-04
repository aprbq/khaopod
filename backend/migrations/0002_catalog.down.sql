-- rollback migration 0002 — ลบตารางแคตตาล็อก (ตามลำดับ dependency)
DROP TABLE IF EXISTS product_images;
DROP TABLE IF EXISTS product_variants;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS categories;
