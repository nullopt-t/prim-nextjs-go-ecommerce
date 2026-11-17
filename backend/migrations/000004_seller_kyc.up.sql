-- Seller KYC: legal business identity and verification documents submitted
-- through the merchant portal and reviewed by administrators.

ALTER TABLE sellers
    ADD COLUMN IF NOT EXISTS legal_name text NULL,
    ADD COLUMN IF NOT EXISTS tax_number text NULL,
    ADD COLUMN IF NOT EXISTS commercial_register text NULL;

CREATE TABLE IF NOT EXISTS seller_documents (
    id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    seller_id           uuid NOT NULL REFERENCES sellers(id) ON DELETE CASCADE,
    document_type       text NOT NULL CHECK (document_type IN ('commercial_register', 'tax_card', 'national_id', 'other')),
    storage_object_id   uuid NOT NULL REFERENCES storage_objects(id) ON DELETE RESTRICT,
    file_name           text NOT NULL,
    file_size           bigint NOT NULL,
    content_type        text NOT NULL,
    status              text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected')),
    rejection_reason    text NULL,
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now(),
    reviewed_at         timestamptz NULL,
    reviewed_by         uuid NULL REFERENCES users(id) ON DELETE SET NULL,
    UNIQUE (seller_id, document_type)
);

CREATE INDEX IF NOT EXISTS idx_seller_documents_seller_id ON seller_documents(seller_id);
