CREATE EXTENSION IF NOT EXISTS citext;

DROP TYPE IF EXISTS user_status CASCADE;

CREATE TYPE user_status AS ENUM(
  'active',
  'inactive',
  'suspended',
  'locked',
  'deleted'
);

CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  identifier TEXT NOT NULL,
  status user_status NOT NULL DEFAULT 'active',
  suspended_until TIMESTAMPTZ NULL,
  locked_until TIMESTAMPTZ NULL,
  last_login_at TIMESTAMPTZ NULL,
  last_login_ip TEXT NULL,
  deleted_at TIMESTAMPTZ NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX users_active_idx ON users (id)
WHERE
  status = 'active'
  AND deleted_at IS NULL;

CREATE INDEX users_inactive_idx ON users (id)
WHERE
  status = 'inactive'
  AND deleted_at IS NULL;

CREATE INDEX users_suspended_idx ON users (id)
WHERE
  status = 'suspended'
  AND deleted_at IS NULL;

CREATE INDEX users_locked_idx ON users (id)
WHERE
  status = 'locked'
  AND deleted_at IS NULL;

CREATE TABLE roles (
  id INT PRIMARY KEY,
  code TEXT NOT NULL UNIQUE,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

INSERT INTO
  roles (id, code)
VALUES
  (1, 'OWNER'),
  (2, 'ADMIN'),
  (3, 'VENDOR');

DROP TYPE IF EXISTS vendor_status CASCADE;

CREATE TYPE vendor_status As ENUM('pending', 'approved', 'rejected', 'suspended');

CREATE TABLE vendors (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL UNIQUE,
  store_name TEXT NOT NULL UNIQUE,
  description TEXT NOT NULL DEFAULT '',
  status vendor_status NOT NULL DEFAULT 'pending',
  deleted_at TIMESTAMPTZ NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

CREATE INDEX vendors_store_name_idx ON vendors (store_name);

CREATE INDEX vendors_pending_status_idx ON vendors (status)
WHERE
  status = 'pending';

CREATE INDEX vendors_suspended_status_idx ON vendors (status)
WHERE
  status = 'suspended';

CREATE INDEX vendors_approved_status_idx ON vendors (status)
WHERE
  status = 'approved';

CREATE INDEX vendors_rejected_status_idx ON vendors (status)
WHERE
  status = 'rejected';

CREATE OR REPLACE FUNCTION set_updated_at () RETURNS TRIGGER AS $$ BEGIN NEW.updated_at = NOW();

RETURN NEW;

END;

$$ LANGUAGE plpgsql;

CREATE TRIGGER users_updated_at
BEFORE UPDATE ON users FOR EACH ROW
EXECUTE FUNCTION set_updated_at ();
