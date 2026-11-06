-- =====================================================================
-- 000005 User browsing history
-- Per-user recently viewed products, deduplicated and capped by the
-- service's rolling window.
-- =====================================================================

BEGIN;

CREATE TABLE IF NOT EXISTS user_product_views (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    product_id uuid NOT NULL REFERENCES products (id) ON DELETE CASCADE,
    viewed_at  timestamptz NOT NULL DEFAULT now(),
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT uq_user_product_views UNIQUE (user_id, product_id)
);

-- Primary read path: the newest N views for one account.
CREATE INDEX IF NOT EXISTS idx_user_product_views_lookup
    ON user_product_views (user_id, viewed_at DESC);

-- Product deletion cascades benefit from an indexed child lookup.
CREATE INDEX IF NOT EXISTS idx_user_product_views_product_id
    ON user_product_views (product_id);

COMMIT;
