-- ============================================================
-- SEED: Master Data (LBI) Tables
-- Platform: Tryly (Pet Food OEM B2B)
-- อ้างอิงสคีมา: migration/001_rebuild_schema.sql
-- หมายเหตุ: รันเพียงครั้งเดียว หลังจาก migration 001
-- ============================================================

BEGIN;

-- ─── 1. lbi_factory_types ─────────────────────────────────────
INSERT INTO lbi_factory_types (type_name, status) VALUES
  ('โรงงานอาหารสัตว์เลี้ยง',         '1'),
  ('โรงงานขนมสัตว์เลี้ยง',           '1'),
  ('โรงงานอาหารเสริมสัตว์เลี้ยง',    '1'),
  ('โรงงานบรรจุภัณฑ์อาหารสัตว์',     '1'),
  ('โรงงานผลิตของเล่นสัตว์เลี้ยง',   '1'),
  ('โรงงานผลิตเสื้อผ้าสัตว์เลี้ยง',  '1'),
  ('โรงงานผลิตวัคซีน/ยาสัตว์',       '1'),
  ('โรงงานผลิตอุปกรณ์สัตว์เลี้ยง',   '1'),
  ('โรงงานแปรรูปวัตถุดิบ',           '1'),
  ('อื่นๆ',                           '1');

-- ─── 2. lbi_certificates ──────────────────────────────────────
INSERT INTO lbi_certificates (cert_name, description, status) VALUES
  ('GMP',        'Good Manufacturing Practice — มาตรฐานการผลิตที่ดี',                                        '1'),
  ('HACCP',      'Hazard Analysis Critical Control Points — ระบบวิเคราะห์อันตรายและจุดวิกฤต',               '1'),
  ('ISO 9001',   'ระบบบริหารคุณภาพ',                                                                         '1'),
  ('ISO 22000',  'ระบบการจัดการความปลอดภัยของอาหาร',                                                         '1'),
  ('Halal',      'ใบรับรองฮาลาล — มาตรฐานอาหารสำหรับผู้นับถือศาสนาอิสลาม',                                  '1'),
  ('อย.',        'สำนักงานคณะกรรมการอาหารและยา (FDA Thailand)',                                              '1'),
  ('มกษ.',       'มาตรฐานสินค้าเกษตร — กรมวิชาการเกษตร',                                                     '1'),
  ('FAMI-QS',    'Feed Additives & Premixtures Quality System',                                              '1'),
  ('BRC',        'British Retail Consortium Global Standard for Food Safety',                                '1'),
  ('GHP',        'Good Hygiene Practice — หลักปฏิบัติด้านสุขลักษณะที่ดี',                                    '1');

-- ─── 3. lbi_shipping_methods ──────────────────────────────────
INSERT INTO lbi_shipping_methods (method_name, status) VALUES
  ('ลูกค้ารับเองที่โรงงาน',  '1'),
  ('จัดส่งเดลิเวอร์รี่',     '1');

-- ─── 4. lbi_production (ขั้นตอนการผลิต) ──────────────────────
INSERT INTO lbi_production (step_name, step_name_th, description, sort_order, step_code) VALUES
  ('Material Preparation',    'จัดเตรียมวัตถุดิบ',    'ตรวจรับและเตรียมวัตถุดิบก่อนเริ่มการผลิต',                          1,  'MAT_PREP'),
  ('Processing / Production', 'ขั้นตอนการผลิต',       'กระบวนการผลิตหลัก เช่น อัดเม็ด อบ ฟรีซดราย ฯลฯ',                  2,  'PROCESSING'),
  ('Quality Control (QC)',    'ตรวจสอบคุณภาพ',         'ตรวจสอบคุณภาพสินค้าก่อนบรรจุ ทดสอบค่าโภชนาการและความปลอดภัย',    3,  'QC'),
  ('READY_TO_SHIP',           'เตรียมจัดส่ง',         'จัดเรียงสินค้า แพ็คลัง เตรียมเอกสารจัดส่ง',                        4,  'READY_TO_SHIP'),
  ('Shipped',                 'จัดส่งแล้ว',           'จัดส่งสินค้าไปยังลูกค้า',                                            5,  'SHIPPED');

-- ─── 5. lbi_categories ────────────────────────────────────────
-- scope: PD = Product, MT = Material
INSERT INTO lbi_categories (name, scope) VALUES
  ('อาหารเม็ดสัตว์เลี้ยง',          'PD'),
  ('อาหารเปียก / Wet Food',         'PD'),
  ('ขนมสัตว์เลี้ยง',                'PD'),
  ('อาหารฟรีซดราย',                 'PD'),
  ('ท้อปเปอร์ / โรยหน้า',           'PD'),
  ('อาหารเสริมและวิตามิน',          'PD'),
  ('บรรจุภัณฑ์สำเร็จรูป',          'PD'),
  ('ของเล่นสัตว์เลี้ยง',           'PD'),
  ('เสื้อผ้าและเครื่องแต่งกาย',    'PD'),
  ('อุปกรณ์และของใช้',              'PD'),
  ('วัตถุดิบโปรตีนสัตว์',           'MT'),
  ('วัตถุดิบโปรตีนพืช',             'MT'),
  ('วิตามินและแร่ธาตุ',             'MT'),
  ('วัตถุเจือปนอาหารและสารกันเสีย', 'MT'),
  ('วัตถุดิบเส้นใย / ธัญพืช',       'MT'),
  ('วัตถุดิบบรรจุภัณฑ์',            'MT');

-- ─── 6. lbi_provinces ──────────────────────────────────────────
INSERT INTO lbi_provinces (name_th, name_en, status) VALUES
  ('กรุงเทพมหานคร', 'Bangkok',           '1'),
  ('กระบี่',        'Krabi',             '1'),
  ('กาญจนบุรี',    'Kanchanaburi',      '1'),
  ('กาฬสินธุ์',    'Kalasin',           '1'),
  ('กำแพงเพชร',    'Kamphaeng Phet',    '1'),
  ('ขอนแก่น',      'Khon Kaen',         '1'),
  ('จันทบุรี',     'Chanthaburi',       '1'),
  ('ฉะเชิงเทรา',  'Chachoengsao',      '1'),
  ('ชลบุรี',       'Chon Buri',         '1'),
  ('ชัยนาท',       'Chai Nat',          '1');

-- ─── 7. lbi_districts (กรุงเทพ — สำหรับ seed ที่อยู่) ───────────
INSERT INTO lbi_districts (row_id, province_id, name_th, name_en, status) VALUES
  (1, 1, 'เขตพระนคร', 'Phra Nakhon', '1'),
  (2, 1, 'เขตบางรัก', 'Bang Rak',    '1'),
  (3, 1, 'เขตสาทร',   'Sathon',      '1');

SELECT setval(pg_get_serial_sequence('lbi_districts', 'row_id'), GREATEST((SELECT MAX(row_id) FROM lbi_districts), 3));
SELECT setval(pg_get_serial_sequence('lbi_provinces', 'row_id'), GREATEST((SELECT MAX(row_id) FROM lbi_provinces), 10));
SELECT setval(pg_get_serial_sequence('lbi_categories', 'category_id'), GREATEST((SELECT MAX(category_id) FROM lbi_categories), 16));
SELECT setval(pg_get_serial_sequence('lbi_factory_types', 'factory_type_id'), GREATEST((SELECT MAX(factory_type_id) FROM lbi_factory_types), 10));

-- ─── 8. lbi_sub_categories ────────────────────────────────────
-- category_id 1–10 = PD, 11–16 = MT (ตามลำดับ insert lbi_categories)
INSERT INTO lbi_sub_categories (category_id, name, status, sort_order) VALUES
  (1, 'อาหารเม็ดสุนัข',                 '1', 1),
  (1, 'อาหารเม็ดแมว',                   '1', 2),
  (1, 'อาหารเม็ดกระต่าย',              '1', 3),
  (1, 'อาหารเม็ดสัตว์เล็ก (หนู/แฮม)', '1', 4),
  (1, 'สูตร Grain-Free',               '1', 5),
  (1, 'สูตร Holistic / Organic',       '1', 6),
  (1, 'สูตรควบคุมน้ำหนัก',             '1', 7),
  (1, 'สูตรลูกสุนัข/ลูกแมว',          '1', 8),
  (1, 'อื่นๆ',                         '1', 99),
  (2, 'Wet Food สุนัข',                '1', 1),
  (2, 'Wet Food แมว',                  '1', 2),
  (2, 'อาหารแมวเลีย (Lickable Treat)', '1', 3),
  (2, 'ซุปและน้ำซุปสัตว์เลี้ยง',      '1', 4),
  (2, 'อาหารกระป๋อง',                  '1', 5),
  (2, 'อื่นๆ',                         '1', 99),
  (3, 'ขนมสุนัข',                      '1', 1),
  (3, 'ขนมแมว',                        '1', 2),
  (3, 'Dental Chew / ขัดฟัน',         '1', 3),
  (3, 'Jerky / เนื้อแผ่น',            '1', 4),
  (3, 'ขนมกระดูก (Bone Treat)',        '1', 5),
  (3, 'Biscuit / คุกกี้สัตว์เลี้ยง',  '1', 6),
  (3, 'อื่นๆ',                         '1', 99),
  (4, 'Freeze-Dried สุนัข',           '1', 1),
  (4, 'Freeze-Dried แมว',             '1', 2),
  (4, 'Freeze-Dried เนื้อสัตว์',      '1', 3),
  (4, 'Freeze-Dried ผัก/ผลไม้',       '1', 4),
  (4, 'อื่นๆ',                         '1', 99),
  (5, 'ท้อปเปอร์สุนัข',               '1', 1),
  (5, 'ท้อปเปอร์แมว',                 '1', 2),
  (5, 'ผงโรยหน้า (Seasoning)',         '1', 3),
  (5, 'อื่นๆ',                         '1', 99),
  (6, 'Probiotic / Prebiotic',        '1', 1),
  (6, 'วิตามินรวม (Multivitamin)',     '1', 2),
  (6, 'Omega-3 / น้ำมันปลา',         '1', 3),
  (6, 'Collagen / Joint Support',     '1', 4),
  (6, 'อาหารเสริมบำรุงขน',            '1', 5),
  (6, 'อาหารเสริมระบบย่อย',           '1', 6),
  (6, 'อื่นๆ',                         '1', 99),
  (7, 'ถุง Stand-up Pouch',           '1', 1),
  (7, 'กระป๋องอลูมิเนียม',            '1', 2),
  (7, 'ซองฟอยล์ / Sachet',            '1', 3),
  (7, 'กล่องกระดาษ / Box',            '1', 4),
  (7, 'ขวดพลาสติก / PET',             '1', 5),
  (7, 'อื่นๆ',                         '1', 99),
  (8, 'ของเล่นสุนัข',                 '1', 1),
  (8, 'ของเล่นแมว',                   '1', 2),
  (8, 'ของเล่นนก/สัตว์เล็ก',           '1', 3),
  (8, 'อื่นๆ',                         '1', 99),
  (9, 'เสื้อผ้าสุนัข',               '1', 1),
  (9, 'เสื้อผ้าแมว',                 '1', 2),
  (9, 'สายจูง / ปลอกคอ',            '1', 3),
  (9, 'รองเท้าสัตว์เลี้ยง',          '1', 4),
  (9, 'อื่นๆ',                         '1', 99),
  (10, 'ชามอาหาร / น้ำ',              '1', 1),
  (10, 'กรง / บ้านสัตว์เลี้ยง',      '1', 2),
  (10, 'อุปกรณ์อาบน้ำและดูแลขน',     '1', 3),
  (10, 'กระบะทราย / ห้องน้ำแมว',     '1', 4),
  (10, 'ที่ข่วนเล็บ',                '1', 5),
  (10, 'อื่นๆ',                         '1', 99),
  (11, 'เนื้อไก่ (Chicken Meal)',     '1', 1),
  (11, 'ปลาและปลาป่น (Fish Meal)',   '1', 2),
  (11, 'เนื้อวัว (Beef)',              '1', 3),
  (11, 'เนื้อหมู (Pork)',             '1', 4),
  (11, 'ไข่และผลิตภัณฑ์ไข่',        '1', 5),
  (11, 'กุ้ง / ซีฟู้ด',             '1', 6),
  (11, 'เนื้อเป็ด (Duck)',            '1', 7),
  (11, 'แมลง (Insect Protein)',       '1', 8),
  (11, 'อื่นๆ',                         '1', 99),
  (12, 'Soy Protein / ถั่วเหลือง',   '1', 1),
  (12, 'ข้าวโพด',                    '1', 2),
  (12, 'ข้าวสาลี',                    '1', 3),
  (12, 'มันสำปะหลัง',                 '1', 4),
  (12, 'ถั่วลันเตา (Pea Protein)',    '1', 5),
  (12, 'อื่นๆ',                         '1', 99),
  (13, 'วิตามิน A / D / E / K',      '1', 1),
  (13, 'วิตามิน B Complex',           '1', 2),
  (13, 'แคลเซียมและฟอสฟอรัส',        '1', 3),
  (13, 'Taurine',                     '1', 4),
  (13, 'L-Carnitine',                 '1', 5),
  (13, 'Zinc / Iron / Selenium',      '1', 6),
  (13, 'อื่นๆ',                         '1', 99),
  (14, 'สารกันเสีย (Preservatives)',  '1', 1),
  (14, 'สารแต่งสี (Color Additives)', '1', 2),
  (14, 'สารให้กลิ่น (Flavoring)',    '1', 3),
  (14, 'สารเพิ่มความข้น (Thickener)', '1', 4),
  (14, 'สารต้านอนุมูลอิสระ',          '1', 5),
  (14, 'อื่นๆ',                         '1', 99),
  (15, 'ข้าวกล้อง / White Rice',     '1', 1),
  (15, 'โอ๊ต (Oat)',                  '1', 2),
  (15, 'Beet Pulp (เส้นใยบีทรูท)',   '1', 3),
  (15, 'ผักและผลไม้อบแห้ง',           '1', 4),
  (15, 'อื่นๆ',                         '1', 99),
  (16, 'ฟิล์มพลาสติก (Film)',        '1', 1),
  (16, 'ถุง Kraft Paper',             '1', 2),
  (16, 'กล่องกระดาษลูกฟูก',          '1', 3),
  (16, 'ฉลากและสติกเกอร์',           '1', 4),
  (16, 'ซิปล็อค / Zip Lock',         '1', 5),
  (16, 'อื่นๆ',                         '1', 99);

SELECT setval(pg_get_serial_sequence('lbi_sub_categories', 'sub_category_id'), GREATEST((SELECT MAX(sub_category_id) FROM lbi_sub_categories), 1));

-- ─── 9. lbi_sub_districts (กรุงเทพ — ตัวอย่าง) ─────────────────
INSERT INTO lbi_sub_districts (district_id, name_th, name_en, zip_code, status) VALUES
  (1, 'พระบรมมหาราชวัง', 'Phra Borom Maha Ratchawang', '10200', '1'),
  (1, 'วังบูรพาภิรมย์',  'Wang Burapha Phirom',        '10200', '1'),
  (1, 'วัดราชบพิธ',      'Wat Ratchabophit',           '10200', '1'),
  (1, 'สำราญราษฎร์',     'Samran Rat',                 '10200', '1'),
  (1, 'ศาลเจ้าพ่อเสือ',  'San Chao Pho Suea',         '10200', '1'),
  (1, 'เสาชิงช้า',       'Sao Chingcha',               '10200', '1'),
  (1, 'บวรนิเวศ',        'Bowon Niwet',                '10200', '1'),
  (1, 'ตลาดยอด',         'Talat Yot',                  '10200', '1'),
  (1, 'ชนะสงคราม',       'Chana Songkhram',            '10200', '1'),
  (1, 'บ้านพานถม',       'Ban Phan Thom',              '10200', '1'),
  (2, 'มหาพฤฒาราม',     'Maha Phruettharam',          '10500', '1'),
  (2, 'สีลม',            'Si Lom',                     '10500', '1'),
  (2, 'สุริยวงศ์',       'Suriya Wong',                '10500', '1'),
  (2, 'บางรัก',          'Bang Rak',                   '10500', '1'),
  (2, 'สี่พระยา',        'Si Phraya',                  '10500', '1'),
  (3, 'ทุ่งมหาเมฆ',     'Thung Maha Mek',             '10120', '1'),
  (3, 'ทุ่งวัดดอน',     'Thung Wat Don',              '10120', '1'),
  (3, 'ยานนาวา',         'Yan Nawa',                   '10120', '1');

SELECT setval(pg_get_serial_sequence('lbi_sub_districts', 'row_id'), GREATEST((SELECT MAX(row_id) FROM lbi_sub_districts), 1));

COMMIT;

-- ตรวจหลังรัน:
-- SELECT COUNT(*) as lbi_provinces FROM lbi_provinces;
-- SELECT COUNT(*) as lbi_categories FROM lbi_categories;
-- SELECT COUNT(*) as lbi_sub_categories FROM lbi_sub_categories;
