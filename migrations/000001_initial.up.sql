-- Users
CREATE TABLE users (
    id TEXT PRIMARY KEY,
    login TEXT NOT NULL UNIQUE,
    hashed_password TEXT NOT NULL,
    balance NUMERIC(12,2) NOT NULL DEFAULT 0,
    withdraw NUMERIC(12,2) NOT NULL DEFAULT 0
);

-- Индексы для users
CREATE INDEX idx_users_id ON users(id);
CREATE INDEX idx_users_login ON users(login);

-- Orders
CREATE TABLE orders (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status TEXT NOT NULL,
    created TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    amount NUMERIC(12,2) NOT NULL
);

-- Индексы для orders
CREATE INDEX idx_orders_id ON orders(id);
CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_user_created ON orders(user_id, created DESC);

-- Withdraws
CREATE TABLE withdraws (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    amount NUMERIC(12,2) NOT NULL
);

-- Индексы для withdraws
CREATE INDEX idx_withdraws_id ON withdraws(id);
CREATE INDEX idx_withdraws_user_id ON withdraws(user_id);
CREATE INDEX idx_withdraws_user_created ON withdraws(user_id, created DESC);
