-- Add expired_date column to rfqs table
-- expired_date is set on creation as created_at + tconfig('rfq_expired') days

BEGIN;

ALTER TABLE rfqs
ADD COLUMN expired_date DATE;

-- Backfill: set expired_date to created_at + 30 days (default from tconfig)
UPDATE rfqs
SET expired_date = (created_at::date + INTERVAL '30 days')
WHERE expired_date IS NULL;

COMMIT;
