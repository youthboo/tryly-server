BEGIN;

-- Material-type (MT) showcases: align DB CHECK with API and RFQ material-sample flows.
ALTER TABLE factory_showcases
    DROP CONSTRAINT IF EXISTS factory_showcases_content_type_check;
ALTER TABLE factory_showcases
    ADD CONSTRAINT factory_showcases_content_type_check
        CHECK (content_type IN ('PD', 'PM', 'ID', 'MT'));

COMMIT;
