-- Per-vendor shipping terms. Checkout partitions the cart by seller and
-- charges each vendor's own fee (waived at that vendor's free-shipping
-- threshold); platform items keep the global store shipping rule.

ALTER TABLE sellers
    ADD COLUMN IF NOT EXISTS shipping_fee bigint NOT NULL DEFAULT 0 CHECK (shipping_fee >= 0),
    ADD COLUMN IF NOT EXISTS free_shipping_threshold bigint NULL CHECK (free_shipping_threshold IS NULL OR free_shipping_threshold >= 0),
    ADD COLUMN IF NOT EXISTS shipping_lead_time_days integer NOT NULL DEFAULT 3 CHECK (shipping_lead_time_days BETWEEN 1 AND 30);
