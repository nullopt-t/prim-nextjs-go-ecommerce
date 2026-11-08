-- Record customer RMA return requests so cumulative returned quantities per
-- line item can be enforced across multiple requests (not just within one).
CREATE TABLE IF NOT EXISTS order_returns (
    id          uuid NOT NULL,
    order_id    uuid NOT NULL,
    customer_id uuid NOT NULL,
    reason      text NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (id),
    FOREIGN KEY (order_id) REFERENCES orders (id) ON DELETE CASCADE,
    FOREIGN KEY (customer_id) REFERENCES users (id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_order_returns_order_id ON order_returns (order_id);

CREATE TABLE IF NOT EXISTS order_return_items (
    id             uuid NOT NULL,
    return_id      uuid NOT NULL,
    order_item_id  uuid NOT NULL,
    quantity       int NOT NULL,
    PRIMARY KEY (id),
    UNIQUE (return_id, order_item_id),
    CHECK (quantity > 0),
    FOREIGN KEY (return_id) REFERENCES order_returns (id) ON DELETE CASCADE,
    FOREIGN KEY (order_item_id) REFERENCES order_items (id)
);

CREATE INDEX IF NOT EXISTS idx_order_return_items_return_id ON order_return_items (return_id);
CREATE INDEX IF NOT EXISTS idx_order_return_items_order_item_id ON order_return_items (order_item_id);
