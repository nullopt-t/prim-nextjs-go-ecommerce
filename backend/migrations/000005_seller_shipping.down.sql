ALTER TABLE sellers
    DROP COLUMN IF EXISTS shipping_lead_time_days,
    DROP COLUMN IF EXISTS free_shipping_threshold,
    DROP COLUMN IF EXISTS shipping_fee;
