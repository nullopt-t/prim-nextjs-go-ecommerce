-- -- Delete all order items referencing seeded products first (child table)
-- -- This also cleans up any test orders that use the seeded products
-- DO $$
-- BEGIN
--     IF EXISTS (SELECT 1 FROM pg_tables WHERE schemaname = 'public' AND tablename = 'order_items') THEN
--         DELETE FROM order_items;
--     END IF;
-- END $$;


-- -- Check if the order items table exists before deleting
-- DO $$
-- BEGIN
--     IF EXISTS (SELECT 1 FROM pg_tables WHERE schemaname = 'public' AND tablename = 'order_items') THEN
--         DELETE FROM order_items;
--     END IF;
-- END $$;

-- -- Check if the orders table exists before deleting
-- DO $$
-- BEGIN
--     IF EXISTS (SELECT 1 FROM pg_tables WHERE schemaname = 'public' AND tablename = 'orders') THEN
--         DELETE FROM orders;
--     END IF;
-- END $$;

-- -- Check if the products table exists before deleting
-- DO $$
-- BEGIN
--     IF EXISTS (SELECT 1 FROM pg_tables WHERE schemaname = 'public' AND tablename = 'products') THEN
--         DELETE FROM products;
--     END IF;
-- END $$;

-- DO $$
-- BEGIN
--     IF EXISTS (SELECT 1 FROM pg_tables WHERE schemaname = 'public' AND tablename = 'cart_items') THEN
--         DELETE FROM cart_items;
--     END IF;
-- END $$;

-- DO $$
-- BEGIN
--     IF EXISTS (SELECT 1 FROM pg_tables WHERE schemaname = 'public' AND tablename = 'carts') THEN
--         DELETE FROM carts;
--     END IF;
-- END $$;

-- DO $$
-- BEGIN
--     IF EXISTS (SELECT 1 FROM pg_tables WHERE schemaname = 'public' AND tablename = 'users') THEN
--         DELETE FROM users;
--     END IF;
-- END $$;