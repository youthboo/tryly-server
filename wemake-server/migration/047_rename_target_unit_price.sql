BEGIN;

-- Rename target_unit_price → target_price
-- "งบประมาณรวม" (total budget), not per-piece price.
ALTER TABLE rfqs RENAME COLUMN target_unit_price TO target_price;

COMMIT;
