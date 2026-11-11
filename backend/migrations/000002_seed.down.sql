-- Remove every seeded row. The seed owns whole tables, so the deletes are
-- table-wide and ordered so that no foreign key is violated on the way down.

-- Reviews depend on order items; review votes depend on reviews.
DELETE FROM review_votes;
DELETE FROM reviews;

-- Order dependents before their parents.
DELETE FROM order_status_history;
DELETE FROM coupon_redemptions;
DELETE FROM payments;
DELETE FROM order_return_items;
DELETE FROM order_returns;
DELETE FROM order_items;
DELETE FROM orders;

-- Cart and inventory.
DELETE FROM cart_items;
DELETE FROM carts;
DELETE FROM inventory_reservations;
DELETE FROM inventory_ledgers;

-- Wishlist and browsing history reference products, variants, and users.
DELETE FROM wishlist_items;
DELETE FROM user_product_views;

-- Catalog children before parents.
DELETE FROM variant_media;
DELETE FROM product_tag_assignments;
DELETE FROM bundle_items;
DELETE FROM product_bundles;
DELETE FROM product_variants;
DELETE FROM products;
DELETE FROM category_attributes;
DELETE FROM seller_attributes;
DELETE FROM product_categories;
DELETE FROM product_tags;
DELETE FROM product_brands;

-- Promotions reference sellers.
DELETE FROM coupons;

-- Marketplace money trail: payouts restrict seller deletion.
DELETE FROM seller_payouts;
DELETE FROM sellers;

-- Customer data.
DELETE FROM addresses;
DELETE FROM users;

-- Media.
DELETE FROM storage_objects;
