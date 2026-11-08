BEGIN;

-- PostgreSQL cannot remove an enum value; the type change is kept.
ALTER TABLE products
    DROP COLUMN IF EXISTS download_url,
    DROP COLUMN IF EXISTS delivery_instructions;

COMMIT;
