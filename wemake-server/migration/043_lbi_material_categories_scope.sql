-- Migration 043: LBI material categories via categories.scope

ALTER TABLE categories
    ADD COLUMN IF NOT EXISTS scope CHAR(2) NOT NULL DEFAULT 'PD';

ALTER TABLE categories
    DROP CONSTRAINT IF EXISTS categories_scope_check;
ALTER TABLE categories
    ADD CONSTRAINT categories_scope_check
    CHECK (scope IN ('PD', 'MT'));

CREATE INDEX IF NOT EXISTS idx_categories_scope
    ON categories(scope);

INSERT INTO categories (name, scope)
SELECT name, scope
FROM (VALUES
    ('วัตถุประกอบอาหาร', 'MT'),
    ('บรรจุภัณฑ์', 'MT'),
    ('สารเคมีและสารเติมแต่ง', 'MT'),
    ('วัตถุดิบธรรมชาติ', 'MT'),
    ('วัตถุดิบเส้นใยและวัสดุ', 'MT'),
    ('อุปกรณ์และชิ้นส่วนการผลิต', 'MT')
) AS seed(name, scope)
WHERE NOT EXISTS (
    SELECT 1
    FROM categories c
    WHERE c.name = seed.name
);
