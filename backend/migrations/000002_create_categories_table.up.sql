CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    name CITEXT UNIQUE NOT NULL,
    deleted_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE OR REPLACE FUNCTION set_updated_at() RETURNS TRIGGER AS $$ BEGIN NEW.updated_at = NOW();
RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- CREATE INDEX IF NOT EXISTS categories_name_lower_idx ON categories (LOWER(name))
-- WHERE deleted_at IS NULL;
-- CREATE INDEX IF NOT EXISTS categories_name_idx ON categories (name)
-- WHERE deleted_at IS NULL;
CREATE TRIGGER categories_updated_at BEFORE
UPDATE ON categories FOR EACH ROW EXECUTE FUNCTION set_updated_at();