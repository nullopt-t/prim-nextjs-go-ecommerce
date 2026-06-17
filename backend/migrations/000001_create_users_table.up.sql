CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email CITEXT NOT NULL UNIQUE CHECK(
        email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$'
    ),
    phone TEXT NULL UNIQUE CHECK(
        phone ~* '^\+?[0-9]{10,15}$'
    ),
    email_verified_at TIMESTAMPTZ NULL,
    phone_verified_at TIMESTAMPTZ NULL,
    role TEXT NOT NULL DEFAULT 'customer' CHECK(
        role IN ('customer', 'admin', 'vendor')
    ),
    status TEXT NOT NULL DEFAULT 'active' CHECK(
        status IN ('active', 'inactive', 'suspended', 'locked', 'deleted')
    ),
    suspended_until TIMESTAMPTZ NULL,
    locked_until TIMESTAMPTZ NULL,
    password_hash TEXT NOT NULL,
    password_changed_at TIMESTAMPTZ NULL,
    last_login_at TIMESTAMPTZ NULL,
    last_login_ip TEXT NULL,
    deleted_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX users_email_idx ON users(email)
WHERE deleted_at IS NULL;

CREATE INDEX users_phone_idx ON users(phone)
WHERE deleted_at IS NULL;

CREATE INDEX users_email_verified_idx ON users(id)
WHERE email_verified_at IS NOT NULL
    AND deleted_at IS NULL;

CREATE INDEX users_phone_verified_idx ON users(id)
WHERE phone_verified_at IS NOT NULL
    AND deleted_at IS NULL;

CREATE INDEX users_active_idx ON users(id)
WHERE status = 'active'
    AND deleted_at IS NULL;

CREATE INDEX users_inactive_idx ON users(id)
WHERE status = 'inactive'
    AND deleted_at IS NULL;

CREATE INDEX users_suspended_idx ON users(id)
WHERE status = 'suspended'
    AND deleted_at IS NULL;

CREATE INDEX users_locked_idx ON users(id)
WHERE status = 'locked'
    AND deleted_at IS NULL;

CREATE OR REPLACE FUNCTION set_updated_at() RETURNS TRIGGER AS $$ BEGIN NEW.updated_at = NOW();
RETURN NEW;
END;
$$ LANGUAGE plpgsql;
CREATE TRIGGER users_updated_at BEFORE
UPDATE ON users FOR EACH ROW EXECUTE FUNCTION set_updated_at();
