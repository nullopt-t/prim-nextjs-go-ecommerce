

DROP TRIGGER IF EXISTS set_updated_at_orders ON orders;
DROP TRIGGER IF EXISTS set_updated_at_order_items ON order_items;

DROP TABLE IF EXISTS order_items;

DROP TABLE IF EXISTS orders;
