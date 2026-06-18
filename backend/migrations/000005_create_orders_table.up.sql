CREATE OR REPLACE FUNCTION set_updated_at() RETURNS TRIGGER AS $$ BEGIN NEW.updated_at = NOW();
RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    user_id UUID NOT NULL REFERENCES users (id),
    status TEXT NOT NULL DEFAULT 'pending' CHECK(
        status IN (
            'pending',
            'confirmed',
            'paid',
            'processing',
            'shipped',
            'delivered',
            'cancelled',
            'refunded'
        )
    ),
    total_price INT NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now (),
    updated_at timestamptz NOT NULL DEFAULT now (),
    deleted_at timestamptz NULL
);

-- CREATE INDEX IF NOT EXISTS orders_created_at_idx ON orders (created_at DESC);
-- CREATE INDEX IF NOT EXISTS orders_total_price_idx ON orders (total_price DESC);
-- CREATE INDEX IF NOT EXISTS orders_status_idx ON orders (status);
-- CREATE INDEX IF NOT EXISTS orders_user_id_idx ON orders (user_id);
-- CREATE INDEX IF NOT EXISTS orders_deleted_at_idx ON orders (deleted_at);
-- CREATE INDEX IF NOT EXISTS orders_updated_at_idx ON orders (updated_at);

CREATE TRIGGER set_updated_at_orders
BEFORE UPDATE ON orders
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TABLE order_items (
    order_id UUID REFERENCES orders (id) ON DELETE CASCADE,
    product_id UUID REFERENCES products (id),
    quantity INT NOT NULL DEFAULT 1 CHECK(
        quantity > 0
        AND quantity < 1000
    ),
    purchase_price INT NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now (),
    updated_at timestamptz NOT NULL DEFAULT now (),
    PRIMARY KEY (order_id, product_id)
);

CREATE TRIGGER set_updated_at_order_items
BEFORE UPDATE ON order_items
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();