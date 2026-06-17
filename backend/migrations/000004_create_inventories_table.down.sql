-- DROP INDEX IF EXISTS inventories_product_id_idx;
-- DROP INDEX IF EXISTS inventories_quantity_idx;
-- DROP INDEX IF EXISTS inventories_reserved_quantity_idx;
-- DROP INDEX IF EXISTS inventories_deleted_at_idx;
DROP TRIGGER IF EXISTS update_inventories_updated_at ON inventories;
DROP TABLE IF EXISTS inventories;