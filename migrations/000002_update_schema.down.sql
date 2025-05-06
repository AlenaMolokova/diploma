CREATE TABLE balance (
    user_id BIGINT UNIQUE REFERENCES users(id),
    current DOUBLE PRECISION DEFAULT 0,
    withdrawn DOUBLE PRECISION DEFAULT 0
);

INSERT INTO balance (user_id, current, withdrawn)
SELECT id, balance, withdrawn FROM users;

ALTER TABLE users
DROP COLUMN balance,
DROP COLUMN withdrawn;

ALTER TABLE orders
ALTER COLUMN uploaded_at TYPE TIMESTAMP;

ALTER TABLE withdrawals
ALTER COLUMN processed_at TYPE TIMESTAMP;