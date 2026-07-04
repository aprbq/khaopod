-- แยก "สี" ออกจาก variant_name ให้เป็นฟิลด์ของตัวเอง
-- variant_name เก็บไซซ์ (เช่น "ไซซ์ M"), color เก็บสี (เช่น "ขาว"/"ดำ")
-- NULL = สินค้าไม่มีตัวเลือกสี
ALTER TABLE product_variants ADD COLUMN IF NOT EXISTS color TEXT;
