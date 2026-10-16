-- =====================================================================
-- ADMIN SETTINGS: key-value store for store configuration
-- =====================================================================

CREATE TABLE IF NOT EXISTS admin_settings (
    id          uuid NOT NULL DEFAULT gen_random_uuid(),
    key         varchar(100) NOT NULL UNIQUE,
    value       jsonb NOT NULL DEFAULT '{}'::jsonb,
    updated_at  timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (id)
);

COMMIT;
