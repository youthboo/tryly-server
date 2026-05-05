-- Migration 044: per-factory RFQ dismissals

CREATE TABLE IF NOT EXISTS factory_rfq_dismissals (
    factory_id   BIGINT      NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    rfq_id       BIGINT      NOT NULL REFERENCES rfqs(rfq_id) ON DELETE CASCADE,
    dismissed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (factory_id, rfq_id)
);

CREATE INDEX IF NOT EXISTS idx_factory_rfq_dismissals_factory
    ON factory_rfq_dismissals(factory_id);
