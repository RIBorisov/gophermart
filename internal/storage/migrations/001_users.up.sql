BEGIN TRANSACTION;

CREATE TABLE IF NOT EXISTS users(
    user_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    login VARCHAR(200),
    password BYTEA
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_login_is_unique ON users (login);

COMMIT;