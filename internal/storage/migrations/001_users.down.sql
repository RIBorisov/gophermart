BEGIN TRANSACTION;

DROP INDEX IF EXISTS idx_login_is_unique;

DROP TABLE IF EXISTS users;

COMMIT;