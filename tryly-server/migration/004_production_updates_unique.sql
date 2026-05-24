-- ============================================================
-- Fix: production_updates upsert ต้องการ UNIQUE (order_id, step_id)
--
-- โค้ด UpsertTx ใช้ ON CONFLICT (order_id, step_id) DO UPDATE
-- แต่ตารางเดิมมี PK เฉพาะ update_id → PostgreSQL คืน error:
--   "there is no unique or exclusion constraint matching the ON CONFLICT specification"
--
-- ก่อนเพิ่ม constraint ต้องล้างคู่ซ้ำ (order_id, step_id) เก็บ row ล่าสุดไว้
-- ============================================================

BEGIN;

DELETE FROM production_updates pu
USING production_updates pu2
WHERE pu.order_id = pu2.order_id
  AND pu.step_id  = pu2.step_id
  AND pu.update_id < pu2.update_id;

-- Add constraint only if it doesn't exist
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'uq_production_updates_order_step'
    ) THEN
        ALTER TABLE production_updates
            ADD CONSTRAINT uq_production_updates_order_step
            UNIQUE (order_id, step_id);
    END IF;
END $$;

COMMIT;
