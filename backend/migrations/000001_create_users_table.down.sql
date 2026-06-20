DROP TRIGGER IF EXISTS users_updated_at ON users;

DROP INDEX IF EXISTS users_active_idx;

DROP TABLE IF EXISTS users;

DROP EXTENSION IF EXISTS citext;
