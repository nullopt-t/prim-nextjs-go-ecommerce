-- =====================================================================
-- CARTS: one active cart per user at most
-- =====================================================================

CREATE UNIQUE INDEX IF NOT EXISTS idx_carts_active_user
    ON carts (user_id)
    WHERE user_id IS NOT NULL AND deleted_at IS NULL;
