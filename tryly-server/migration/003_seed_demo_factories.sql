-- ============================================================
-- SEED: Demo Factories + Showcases
-- Platform: Tryly (Pet Food OEM B2B)
-- อ้างอิงสคีมา: migration/001_rebuild_schema.sql + 002_seed_lbi_master.sql
-- หมายเหตุ: สำหรับ Development/Testing — ให้มี sample data พร้อมใช้
-- ============================================================

BEGIN;

-- Platform config (หากยังไม่มี)
INSERT INTO platform_config (label, default_commission_rate, vat_rate, effective_from)
SELECT 'Default platform', 5.00, 7.00, NOW()
WHERE NOT EXISTS (SELECT 1 FROM platform_config LIMIT 1);

-- ---------------------------------------------------------------------------
-- Factory users (role = FT)
-- ---------------------------------------------------------------------------

INSERT INTO users (role, email, phone, password_hash, is_active, created_at, updated_at)
VALUES
    ('FT', 'factory.petwear@wemake.co.th',    '0810000002', '$2a$10$CwTycUXWue0Thq9StjUM0uJ8aZ5tV3SuY.9BxZ/XfYHhiFpLVPO0e', TRUE, '2026-04-22 14:01:10', '2026-04-22 14:01:10'),
    ('FT', 'factory.pawout@wemake.co.th',       '0810000003', '$2a$10$CwTycUXWue0Thq9StjUM0uJ8aZ5tV3SuY.9BxZ/XfYHhiFpLVPO0e', TRUE, '2026-04-22 14:01:10', '2026-04-22 14:01:10'),
    ('FT', 'factory.printpack@wemake.co.th',  '0810000004', '$2a$10$CwTycUXWue0Thq9StjUM0uJ8aZ5tV3SuY.9BxZ/XfYHhiFpLVPO0e', TRUE, '2026-04-22 14:01:10', '2026-04-22 14:01:10'),
    ('FT', 'factory.nutripet@wemake.co.th',   '0810000005', '$2a$10$CwTycUXWue0Thq9StjUM0uJ8aZ5tV3SuY.9BxZ/XfYHhiFpLVPO0e', TRUE, '2026-04-22 14:01:10', '2026-04-22 14:01:10'),
    ('FT', 'factory.pettech@wemake.co.th',    '0810000006', '$2a$10$CwTycUXWue0Thq9StjUM0uJ8aZ5tV3SuY.9BxZ/XfYHhiFpLVPO0e', TRUE, '2026-04-22 14:01:10', '2026-04-22 14:01:10'),
    ('FT', 'factory.petfoods@wemake.co.th',   '0810000007', '$2a$10$hVcz6kwc206zqTlZfi75jeeSadd9m1SAHZKE8Yqd4nRu.ar5/rNoS', TRUE, '2026-04-22 14:02:42', '2026-05-10 15:20:55'),
    ('FT', 'factory.rank1@wemake.co.th',      '0810000008', '$2a$10$CwTycUXWue0Thq9StjUM0uJ8aZ5tV3SuY.9BxZ/XfYHhiFpLVPO0e', TRUE, '2026-04-22 14:01:10', '2026-04-22 14:01:10')
ON CONFLICT (email) DO NOTHING;

-- Wallets (หนึ่งกระเป๋าต่อโรงงาน)
INSERT INTO wallets (user_id, good_fund, pending_fund)
SELECT u.user_id, 0, 0
FROM users u
WHERE u.email LIKE 'factory.%@wemake.co.th'
  AND NOT EXISTS (SELECT 1 FROM wallets w WHERE w.user_id = u.user_id);

-- ---------------------------------------------------------------------------
-- factory_profiles
-- ---------------------------------------------------------------------------

INSERT INTO factory_profiles (
    user_id, approval_status, factory_type_id, config_id,
    factory_name, description, tax_id, min_order, lead_time_desc,
    province_id, image_url, background_image_url,
    rating, review_count, completed_orders,
    submitted_at, verified_at
)
SELECT
    u.user_id, 'AP', ft.factory_type_id,
    (SELECT config_id FROM platform_config ORDER BY config_id LIMIT 1),
    v.factory_name, v.description, v.tax_id, v.min_order, v.lead_time_desc,
    1, v.image_url, v.background_image_url,
    v.rating, v.review_count, v.completed_orders,
    v.submitted_at, v.verified_at
FROM (VALUES
    ('factory.petwear@wemake.co.th',   'โรงงานผลิตเสื้อผ้าสัตว์เลี้ยง',  'แพ็ทแวร์ โรงงานตัดเย็บเสื้อผ้าสัตว์', 'รับผลิตเสื้อผ้า/แพทเทิร์นสำหรับสัตว์เลี้ยง พร้อมปักโลโก้แบรนด์', '0105561234567', 100, '7-10', 'https://images.unsplash.com/photo-1514888286974-6c03e2ca1dba?w=400&q=80', NULL, 4.72, 98,  205, '2026-04-22 14:01:10'::timestamp, '2026-04-22 14:01:10'::timestamp),
    ('factory.pawout@wemake.co.th',      'โรงงานผลิตของเล่นสัตว์เลี้ยง',    'พาวเอ้าท์ โรงงานผลิตของเล่นสัตว์', 'ของเล่นสัตว์เลี้ยงจากวัสดุธรรมชาติ ปลอดภัย รองรับ MOQ เริ่ม 200 ชิ้น', '0105562345678', 200, '12-14', 'https://images.unsplash.com/photo-1587300003388-59208cc962cb?w=400&q=80', NULL, 4.68, 210, 178, '2026-04-22 14:01:10'::timestamp, '2026-04-22 14:01:10'::timestamp),
    ('factory.printpack@wemake.co.th',   'โรงงานบรรจุภัณฑ์อาหารสัตว์',     'พริ้นท์แพ็ค โรงงานบรรจุภัณฑ์สัตว์เลี้ยง', 'บรรจุภัณฑ์สำหรับอาหารและขนมสัตว์เลี้ยง พิมพ์ 4 สี คุณภาพสูง', '0105563456789', 300, '10-12', 'https://images.unsplash.com/photo-1607082348824-0a96f2a4b9da?w=400&q=80', NULL, 4.80, 145, 267, '2026-04-22 14:01:10'::timestamp, '2026-04-22 14:01:10'::timestamp),
    ('factory.nutripet@wemake.co.th',    'โรงงานอาหารเสริมสัตว์เลี้ยง',    'นิวทริเพ็ท โรงงานอาหารเสริมสัตว์', 'ผลิตวิตามินเม็ดเคี้ยว-ผงโปรตีนสำหรับสุนัข/แมว พร้อม อย.', '0105564567890', 400, '12-14', 'https://images.unsplash.com/photo-1471864190281-a93a3070b6de?w=400&q=80', NULL, 4.90, 132, 189, '2026-04-22 14:01:10'::timestamp, '2026-04-22 14:01:10'::timestamp),
    ('factory.pettech@wemake.co.th',     'โรงงานผลิตอุปกรณ์สัตว์เลี้ยง',   'เพ็ทเทค โรงงานอุปกรณ์สัตว์เลี้ยง', 'อุปกรณ์ Pet Tech ระดับพรีเมียม รองรับ Bluetooth และ White-label App', '0105565678901', 80,  '18-21', 'https://images.unsplash.com/photo-1517423440428-a5a00ad493e8?w=400&q=80', NULL, 4.65, 115, 94,  '2026-04-22 14:01:10'::timestamp, '2026-04-22 14:01:10'::timestamp),
    ('factory.petfoods@wemake.co.th',    'โรงงานอาหารสัตว์เลี้ยง',         'เพ็ทฟู้ดส์ โรงงานผลิตอาหารสัตว์', 'ผู้เชี่ยวชาญด้านอาหารสัตว์เลี้ยงพรีเมียม รองรับ OEM/ODM ปรับสูตรให้ตรงตลาด', '0105560123456', 500, '10-14', 'https://images.unsplash.com/photo-1589924691995-400dc9ecc119?w=400&q=80', NULL, 4.85, 124, 312, '2026-04-22 14:02:42'::timestamp, '2026-04-22 14:02:42'::timestamp),
    ('factory.rank1@wemake.co.th',       'โรงงานอาหารสัตว์เลี้ยง',         'โรงงานอันดับ 1', 'โรงงานผลิตอาหารแมวคุณภาพ รับผลิตจำนวนน้อย', '1010876543272', 50, '30', 'https://res.cloudinary.com/ddyc15z1v/image/upload/v1777315256/wemake/c9800610-c74b-4dd6-b29d-670955c9437e.jpg', 'https://res.cloudinary.com/ddyc15z1v/image/upload/v1777386335/wemake/9a7ed188-11c3-4d42-85c6-4adbf1251881.webp', 5.00, 2, 0, '2026-04-29 01:05:26'::timestamp, '2026-04-29 01:05:26'::timestamp)
) AS v(email, factory_type_name, factory_name, description, tax_id, min_order, lead_time_desc, image_url, background_image_url, rating, review_count, completed_orders, submitted_at, verified_at)
JOIN users u ON u.email = v.email
JOIN lbi_factory_types ft ON ft.type_name = v.factory_type_name
ON CONFLICT (user_id) DO NOTHING;

-- ---------------------------------------------------------------------------
-- map_factory_categories
-- ---------------------------------------------------------------------------

INSERT INTO map_factory_categories (factory_id, category_id)
SELECT u.user_id, m.category_id
FROM (VALUES
    ('factory.petwear@wemake.co.th',   9),
    ('factory.pawout@wemake.co.th',      8),
    ('factory.printpack@wemake.co.th',   7),
    ('factory.printpack@wemake.co.th',  16),
    ('factory.nutripet@wemake.co.th',    6),
    ('factory.nutripet@wemake.co.th',   13),
    ('factory.pettech@wemake.co.th',    10),
    ('factory.pettech@wemake.co.th',     9),
    ('factory.petfoods@wemake.co.th',    1),
    ('factory.petfoods@wemake.co.th',    4),
    ('factory.rank1@wemake.co.th',       1),
    ('factory.rank1@wemake.co.th',       2),
    ('factory.rank1@wemake.co.th',       3),
    ('factory.rank1@wemake.co.th',       8)
) AS m(email, category_id)
JOIN users u ON u.email = m.email
WHERE NOT EXISTS (
    SELECT 1 FROM map_factory_categories mfc
    WHERE mfc.factory_id = u.user_id AND mfc.category_id = m.category_id
);

-- ---------------------------------------------------------------------------
-- map_factory_sub_categories
-- ---------------------------------------------------------------------------

INSERT INTO map_factory_sub_categories (factory_id, sub_category_id)
SELECT u.user_id, sc.sub_category_id
FROM (VALUES
    ('factory.petwear@wemake.co.th',   'เสื้อผ้าสุนัข'),
    ('factory.pawout@wemake.co.th',    'ของเล่นสุนัข'),
    ('factory.printpack@wemake.co.th', 'ถุง Stand-up Pouch'),
    ('factory.nutripet@wemake.co.th',  'วิตามินรวม (Multivitamin)'),
    ('factory.pettech@wemake.co.th',   'สายจูง / ปลอกคอ'),
    ('factory.petfoods@wemake.co.th',  'สูตร Grain-Free'),
    ('factory.rank1@wemake.co.th',     'อาหารเม็ดแมว')
) AS m(email, sub_category_name)
JOIN users u ON u.email = m.email
JOIN lbi_sub_categories sc ON sc.name = m.sub_category_name
WHERE NOT EXISTS (
    SELECT 1 FROM map_factory_sub_categories mfs
    WHERE mfs.factory_id = u.user_id AND mfs.sub_category_id = sc.sub_category_id
);

-- ---------------------------------------------------------------------------
-- map_factory_certificates
-- ---------------------------------------------------------------------------

INSERT INTO map_factory_certificates (factory_id, cert_id, document_url, expire_date, cert_number)
SELECT u.user_id, c.cert_id, v.document_url, v.expire_date::date, v.cert_number
FROM (VALUES
    ('factory.petfoods@wemake.co.th', 'GMP',      'https://example.com/certs/petfoods-gmp.pdf',      '2027-12-31', 'GMP-TH-2024-001'),
    ('factory.petfoods@wemake.co.th', 'HACCP',    'https://example.com/certs/petfoods-haccp.pdf',    '2027-06-30', 'HACCP-TH-2024-002'),
    ('factory.nutripet@wemake.co.th',  'อย.',      'https://example.com/certs/nutripet-fda.pdf',      '2028-03-31', 'FDA-2567-089'),
    ('factory.nutripet@wemake.co.th',  'ISO 22000','https://example.com/certs/nutripet-iso22000.pdf','2027-09-30', 'ISO22000-2023-11'),
    ('factory.rank1@wemake.co.th',     'GMP',      'https://example.com/certs/rank1-gmp.pdf',         '2027-01-15', 'GMP-R1-2025')
) AS v(email, cert_name, document_url, expire_date, cert_number)
JOIN users u ON u.email = v.email
JOIN lbi_certificates c ON c.cert_name = v.cert_name
WHERE NOT EXISTS (
    SELECT 1 FROM map_factory_certificates mfc
    WHERE mfc.factory_id = u.user_id AND mfc.cert_id = c.cert_id
);

-- ---------------------------------------------------------------------------
-- factory_showcases (ไม่มี showcase_id / ไม่มี view_count, excerpt, tags)
-- content_type: PD=product, PM=promotion, ID=idea, MT=material
-- ---------------------------------------------------------------------------

-- MT — วัตถุดิบ (แพ็ทแวร์)
INSERT INTO factory_showcases (
    factory_id, category_id, content_type, title, content,
    linked_showcases, moq, lead_time_days, status, likes_count, published_at
)
SELECT u.user_id, 15, 'MT', 'ผ้า Polyester กันน้ำ 150D สำหรับเสื้อสัตว์เลี้ยง',
    'ผ้าไมโครไฟเบอร์น้ำหนักเบา ระบายอากาศดี เหมาะสำหรับผลิตเสื้อกันฝนและเสื้อแจ็กเก็ตสุนัข',
    '["https://images.unsplash.com/photo-1558769132-cb1aea458c5e?w=500&h=300&fit=crop"]'::jsonb,
    100, 5, 'AC', 87, '2026-04-01'::timestamp
FROM users u WHERE u.email = 'factory.petwear@wemake.co.th'
ON CONFLICT DO NOTHING;

-- PM — โปรโมชัน (แพ็ทแวร์)
INSERT INTO factory_showcases (
    factory_id, category_id, content_type, title, content,
    linked_showcases, moq, status, likes_count, published_at
)
SELECT u.user_id, 9, 'PM', 'โปรฯ เปิดตัวคอลเลกชัน Summer Pet Wear ลด 12%',
    '> รับผลิตครบแพตเทิร์น เสื้อทีมสัตว์เลี้ยงพร้อมปักโลโก้ หมดเขต 31 มี.ค.

โปรโมชันส่วนลดค่าผลิต 12% สำหรับลูกค้าใหม่ที่สั่งเสื้อผ้าสัตว์เลี้ยงภายใน 31 มีนาคม ขั้นต่ำ 100 ตัว ฟรีปักโลโก้ 1 ตำแหน่ง',
    '["https://images.unsplash.com/photo-1535268647677-300dbf3d78d1?w=400&q=80", "https://images.unsplash.com/photo-1513360371669-4adf3dd7dff8?w=400&q=80"]'::jsonb,
    100, 'AC', 98, '2026-02-27'::timestamp
FROM users u WHERE u.email = 'factory.petwear@wemake.co.th'
ON CONFLICT DO NOTHING;

-- PD — สินค้า (แพ็ทแวร์)
INSERT INTO factory_showcases (
    factory_id, category_id, content_type, title, content,
    linked_showcases, moq, status, likes_count, published_at
)
SELECT u.user_id, 9, 'PD', 'รับผลิตเสื้อกันฝนสุนัขแบบสะท้อนแสง',
    '> วัสดุกันน้ำพร้อมแถบสะท้อนแสง เหมาะกับแบรนด์สาย Outdoor

เสื้อกันฝนสำหรับสุนัขผ้า Ripstop กันน้ำ 5000mm พร้อมแถบสะท้อนแสงรอบตัว',
    '["https://images.unsplash.com/photo-1585110396000-c9ffd4e4b308?w=400&q=80", "https://images.unsplash.com/photo-1526336024174-e58f5cdd8e13?w=400&q=80"]'::jsonb,
    150, 'AC', 88, '2026-02-20'::timestamp
FROM users u WHERE u.email = 'factory.petwear@wemake.co.th'
ON CONFLICT DO NOTHING;

-- ID — ไอเดีย (พาวเอ้าท์)
INSERT INTO factory_showcases (
    factory_id, category_id, content_type, title, content,
    linked_showcases, moq, status, likes_count, published_at
)
SELECT u.user_id, 8, 'ID', 'ไอเดียชุดของเล่นยางธรรมชาติขายหน้าร้อน',
    'แนวคิดการรวมของเล่น 3 แบบขายดีเป็นเซ็ตเดียว เพิ่ม perceived value และ margin สำหรับร้านออนไลน์',
    '[]'::jsonb,
    200, 'AC', 211, '2026-02-26'::timestamp
FROM users u WHERE u.email = 'factory.pawout@wemake.co.th'
ON CONFLICT DO NOTHING;

-- PM — บรรจุภัณฑ์ (พริ้นท์แพ็ค)
INSERT INTO factory_showcases (
    factory_id, category_id, content_type, title, content,
    linked_showcases, moq, status, likes_count, published_at
)
SELECT u.user_id, 7, 'PM', 'โปรแพ็กเกจจิ้งซองสแตนดี้ เริ่มต้น 300 ชิ้น',
    'โปรโมชันซองสแตนดี้ซิปล็อค MOQ ต่ำ 300 ใบ ฟรีค่าปรับไฟล์พิมพ์ 1 ครั้ง',
    '["https://images.unsplash.com/photo-1589924691995-400dc9ecc119?w=400&q=80"]'::jsonb,
    300, 'AC', 76, '2026-02-24'::timestamp
FROM users u WHERE u.email = 'factory.printpack@wemake.co.th'
ON CONFLICT DO NOTHING;

-- MT — วัตถุดิบ (พริ้นท์แพ็ค)
INSERT INTO factory_showcases (
    factory_id, category_id, content_type, title, content,
    linked_showcases, moq, lead_time_days, status, likes_count, published_at
)
SELECT u.user_id, 16, 'MT', 'กระดาษคราฟต์ 120 แกรม ย่อยสลายได้ตามธรรมชาติ',
    'กระดาษคราฟต์ผิวเรียบ เหมาะสำหรับทำถุง กล่อง และฉลากสินค้าสัตว์เลี้ยงที่เน้น Eco-friendly',
    '["https://images.unsplash.com/photo-1607082350899-7e105aa886ae?w=500&h=300&fit=crop"]'::jsonb,
    500, 4, 'AC', 95, '2026-04-05'::timestamp
FROM users u WHERE u.email = 'factory.printpack@wemake.co.th'
ON CONFLICT DO NOTHING;

-- PD — อาหารเสริม (นิวทริเพ็ท)
INSERT INTO factory_showcases (
    factory_id, category_id, content_type, title, content,
    linked_showcases, moq, status, likes_count, published_at
)
SELECT u.user_id, 6, 'PD', 'วิตามินเคี้ยวสำหรับสุนัขสูตร Skin & Coat',
    'วิตามินเม็ดเคี้ยวสูตรบำรุงผิวหนังและเส้นขน ผสม Omega-3, Biotin, Zinc รองรับ OEM',
    '["https://images.unsplash.com/photo-1425082661705-1834bfd09dca?w=400&q=80", "https://images.unsplash.com/photo-1555169062-013468b47731?w=400&q=80"]'::jsonb,
    400, 'AC', 132, '2026-02-19'::timestamp
FROM users u WHERE u.email = 'factory.nutripet@wemake.co.th'
ON CONFLICT DO NOTHING;

-- MT — วิตามิน E (นิวทริเพ็ท)
INSERT INTO factory_showcases (
    factory_id, category_id, content_type, title, content,
    linked_showcases, moq, lead_time_days, base_price, status, likes_count, published_at
)
SELECT u.user_id, 13, 'MT', 'วิตามิน E เกรดอาหารสัตว์ ชนิดผง (d-Alpha Tocopheryl)',
    'วิตามิน E บริสุทธิ์ 96% เหมาะสำหรับเติมในสูตรอาหารเสริม ขนม และอาหารสัตว์เลี้ยง',
    '["https://images.unsplash.com/photo-1471864190281-a93a3070b6de?w=500&h=300&fit=crop"]'::jsonb,
    10, 7, NULL, 'AC', 108, '2026-04-07'::timestamp
FROM users u WHERE u.email = 'factory.nutripet@wemake.co.th'
ON CONFLICT DO NOTHING;

-- PD — Pet Tech (เพ็ทเทค)
INSERT INTO factory_showcases (
    factory_id, category_id, content_type, title, content,
    linked_showcases, moq, status, likes_count, published_at
)
SELECT u.user_id, 10, 'PD', 'ปลอกคอ GPS + Activity Tracker สำหรับแบรนด์ Pet Tech',
    'ปลอกคออัจฉริยะ GPS + Accelerometer รองรับแอป White-label iOS/Android แบตเตอรี่ใช้งาน 7 วัน',
    '["https://images.unsplash.com/photo-1518791841217-8f162f1e1131?w=400&q=80"]'::jsonb,
    80, 'AC', 115, '2026-02-16'::timestamp
FROM users u WHERE u.email = 'factory.pettech@wemake.co.th'
ON CONFLICT DO NOTHING;

-- ID — Pet Tech subscription (เพ็ทเทค)
INSERT INTO factory_showcases (
    factory_id, category_id, content_type, title, content,
    linked_showcases, moq, status, likes_count, published_at
)
SELECT u.user_id, 10, 'ID', 'ไอเดียขายชุดปลอกคออัจฉริยะ + แอปสมาชิก',
    'โมเดลธุรกิจ Hardware + Subscription สำหรับแบรนด์ Pet Tech',
    '[]'::jsonb,
    100, 'AC', 173, '2026-02-15'::timestamp
FROM users u WHERE u.email = 'factory.pettech@wemake.co.th'
ON CONFLICT DO NOTHING;

-- PD — อาหารสัตว์ (เพ็ทฟู้ดส์)
INSERT INTO factory_showcases (
    factory_id, category_id, content_type, title, content,
    linked_showcases, moq, lead_time_days, status, likes_count, published_at
)
SELECT u.user_id, 1, 'PD', 'อาหารเม็ด Grain-Free สำหรับแบรนด์ใหม่',
    'รับผลิตอาหารเม็ดสูตร Grain-Free MOQ เริ่ม 500 กก. ปรับสูตรตามกลุ่มเป้าหมาย',
    '["https://images.unsplash.com/photo-1589924691995-400dc9ecc119?w=500&h=300&fit=crop"]'::jsonb,
    500, 14, 'AC', 120, '2026-02-18'::timestamp
FROM users u WHERE u.email = 'factory.petfoods@wemake.co.th'
ON CONFLICT DO NOTHING;

-- PD — โรงงานอันดับ 1 (ตัวอย่าง content ยาวจากสเปก)
INSERT INTO factory_showcases (
    factory_id, category_id, sub_category_id, content_type, title, content,
    linked_showcases, moq, base_price, status, likes_count, published_at
)
SELECT u.user_id, 1, sc.sub_category_id, 'PD', 'อาหารเม็ดสูตร Grain-Free สำหรับแบรนด์ใหม่',
    '# รับผลิตอาหารสัตว์เลี้ยงเกรดพรีเมียม

## หมวดหมู่สินค้าที่รับผลิต
1. ขนมสุนัขแบบฟรีซดราย
2. อาหารเม็ด (Kibble) สูตรโฮลิสติก
3. อาหารเปียก/ขนมแมวเลีย

## เงื่อนไขการผลิต
| รายการ | รายละเอียด |
| --- | --- |
| MOQ | 500 กก. |
| Lead time | 30-45 วัน |',
    '["https://res.cloudinary.com/ddyc15z1v/image/upload/v1778321436/wemake/283600e0-480a-4f6a-aed7-b25e4ad9a2fe.png"]'::jsonb,
    10, 100.00, 'AC', 0, '2026-05-07'::timestamp
FROM users u
JOIN lbi_sub_categories sc ON sc.category_id = 1 AND sc.name = 'อาหารเม็ดแมว'
WHERE u.email = 'factory.rank1@wemake.co.th'
ON CONFLICT DO NOTHING;

COMMIT;

-- ตรวจหลังรัน:
-- SELECT COUNT(*) as factory_users FROM users WHERE role = 'FT';
-- SELECT COUNT(*) as factory_profiles FROM factory_profiles;
-- SELECT COUNT(*) as factory_showcases FROM factory_showcases;
-- SELECT u.email, fp.factory_name, COUNT(fs.showcase_id) as showcase_count
-- FROM users u
-- JOIN factory_profiles fp ON fp.user_id = u.user_id
-- LEFT JOIN factory_showcases fs ON fs.factory_id = u.user_id
-- WHERE u.role = 'FT'
-- GROUP BY u.email, fp.factory_name
-- ORDER BY u.email;
