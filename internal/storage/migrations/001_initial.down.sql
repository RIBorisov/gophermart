BEGIN TRANSACTION;

-- 4. withdrawals
DROP TABLE IF EXISTS withdrawals;

-- 3. balance
DROP TABLE IF EXISTS balance;

-- 2. orders
DROP TABLE IF EXISTS orders;

-- 1. users
DROP INDEX IF EXISTS idx_login_is_unique;
DROP TABLE IF EXISTS users;

COMMIT;