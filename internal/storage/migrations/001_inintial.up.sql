BEGIN TRANSACTION;

-- 1. users
CREATE TABLE IF NOT EXISTS users(
    user_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    login VARCHAR(200),
    password BYTEA
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_login_is_unique ON users (login);


-- 2. orders
CREATE TABLE IF NOT EXISTS orders(
    order_id VARCHAR(200) UNIQUE NOT NULL,
    user_id UUID NOT NULL,
    status VARCHAR(10) NOT NULL CHECK (status IN ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED')) DEFAULT 'NEW',
    bonus DECIMAL(10, 2) NOT NULL DEFAULT 0.0,
    uploaded_at TIMESTAMP DEFAULT NOW(),

    FOREIGN KEY (user_id) REFERENCES users(user_id)
);

-- 3. balance
CREATE TABLE IF NOT EXISTS balance(
    user_id UUID NOT NULL UNIQUE,
    current DECIMAL(10, 2) NOT NULL DEFAULT 0.0 CHECK (current >= 0),
    withdrawn DECIMAL(10, 2) NOT NULL DEFAULT 0.0,
    updated_at TIMESTAMP DEFAULT NOW(),

    FOREIGN KEY (user_id) REFERENCES users(user_id)
);

-- 4. withdrawals
CREATE TABLE IF NOT EXISTS withdrawals(
    user_id UUID NOT NULL,
    order_id VARCHAR(200) UNIQUE NOT NULL,
    amount DECIMAL(10, 2) NOT NULL DEFAULT 0.0,
    processed_at TIMESTAMP DEFAULT NOW(),

    FOREIGN KEY (user_id) REFERENCES users(user_id)
);

COMMIT;