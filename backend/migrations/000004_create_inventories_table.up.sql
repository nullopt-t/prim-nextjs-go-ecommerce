CREATE OR REPLACE FUNCTION set_updated_at() RETURNS TRIGGER AS $$ BEGIN NEW.updated_at = NOW();
RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE IF NOT EXISTS inventories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products (id) ON DELETE CASCADE,
    quantity INTEGER NOT NULL DEFAULT 0 CHECK (quantity >= 0),
    reserved_quantity INTEGER NOT NULL DEFAULT 0 CHECK (reserved_quantity >= 0),
    last_restocked_at TIMESTAMPTZ NULL,
    last_sold_at TIMESTAMPTZ NULL,
    deleted_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW ()
);
-- CREATE INDEX IF NOT EXISTS inventories_product_id_idx ON inventories (product_id);
-- CREATE INDEX IF NOT EXISTS inventories_quantity_idx ON inventories (quantity);
-- CREATE INDEX IF NOT EXISTS inventories_reserved_quantity_idx ON inventories (reserved_quantity);
-- CREATE INDEX IF NOT EXISTS inventories_deleted_at_idx ON inventories (deleted_at);
CREATE TRIGGER update_inventories_updated_at BEFORE
UPDATE ON inventories FOR EACH ROW EXECUTE FUNCTION set_updated_at();