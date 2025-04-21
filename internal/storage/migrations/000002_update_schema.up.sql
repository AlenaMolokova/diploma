ALTER TABLE users
ADD COLUMN balance DOUBLE PRECISION DEFAULT 0.0,
ADD COLUMN withdrawn DOUBLE PRECISION DEFAULT 0.0;

UPDATE users
SET balance = COALESCE((
    SELECT current
    FROM balance
    WHERE balance.user_id = users.id
), 0.0),
    withdrawn = COALESCE((
    SELECT withdrawn
    FROM balance
    WHERE balance.user_id = users.id
), 0.0);

ALTER TABLE orders
ALTER COLUMN uploaded_at TYPE TIMESTAMP WITH TIME ZONE;

ALTER TABLE withdrawals
ALTER COLUMN processed_at TYPE TIMESTAMP WITH TIME ZONE;

DROP TABLE balance;