DROP INDEX IF EXISTS idx_seller_documents_seller_id;
DROP TABLE IF EXISTS seller_documents;

ALTER TABLE sellers
    DROP COLUMN IF EXISTS commercial_register,
    DROP COLUMN IF EXISTS tax_number,
    DROP COLUMN IF EXISTS legal_name;
