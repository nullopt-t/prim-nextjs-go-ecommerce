-- =====================================================================
-- 000004 Merchant lifecycle hardening
-- Status review metadata, slug policy, payout idempotency, and frozen
-- commission components on order lines.
-- =====================================================================

BEGIN;

ALTER TABLE vendors
    ADD COLUMN status_reason text NULL,
    ADD COLUMN reviewed_by uuid NULL REFERENCES users (id) ON DELETE SET NULL,
    ADD COLUMN reviewed_at timestamptz NULL;

-- Slugs are public identifiers: enforce the API's 3-63 character policy at
-- the database boundary as well.
ALTER TABLE vendors DROP CONSTRAINT IF EXISTS vendors_slug_check;
ALTER TABLE vendors ADD CONSTRAINT vendors_slug_check
    CHECK (
        length(slug) BETWEEN 3 AND 63
        AND slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$'
    );

-- Payout retries must be idempotent per vendor. NULL keys preserve the
-- existing manual/admin callers that do not supply one.
ALTER TABLE vendor_payouts ADD COLUMN idempotency_key text NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_vendor_payouts_idempotency
    ON vendor_payouts (vendor_id, idempotency_key)
    WHERE idempotency_key IS NOT NULL;

-- Freeze both sides of the commission split for historical audit. The
-- existing vendor_earning stays the authoritative vendor share.
ALTER TABLE order_items
    ADD COLUMN commission_rate_bps int NOT NULL DEFAULT 0
        CHECK (commission_rate_bps BETWEEN 0 AND 10000),
    ADD COLUMN commission_amount bigint NOT NULL DEFAULT 0
        CHECK (commission_amount >= 0);

COMMIT;
