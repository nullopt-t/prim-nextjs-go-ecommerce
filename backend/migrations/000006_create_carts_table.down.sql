-- DROP INDEX IF EXISTS carts_user_id_idx;
-- DROP INDEX IF EXISTS carts_deleted_at_idx;
DROP TRIGGER IF EXISTS trigger_carts_updated_at ON carts;
DROP TRIGGER IF EXISTS trigger_cart_items_updated_at ON cart_items;
DROP TABLE IF EXISTS cart_items;
DROP TABLE IF EXISTS carts;