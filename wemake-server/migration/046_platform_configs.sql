BEGIN;

CREATE TABLE IF NOT EXISTS platform_configs (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

INSERT INTO platform_configs (key, value)
VALUES ('shipping_days', '7')
ON CONFLICT (key) DO NOTHING;

COMMIT;
