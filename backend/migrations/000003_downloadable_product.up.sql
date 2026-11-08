BEGIN;

-- Existing databases already ran 000001_init, so the downloadable additions
-- need their own version to reach them.
ALTER TYPE product_type ADD VALUE IF NOT EXISTS 'downloadable';

ALTER TABLE products
    ADD COLUMN IF NOT EXISTS download_url text NULL,
    ADD COLUMN IF NOT EXISTS delivery_instructions text NULL;

COMMIT;
