-- name: CreateUser :one
INSERT INTO users (login, password)
VALUES ($1, $2)
RETURNING id;

-- name: CreateOrder :exec
INSERT INTO orders (user_id, number, status, uploaded_at)
VALUES ($1, $2, $3, $4);

-- name: GetOrderByNumber :one
SELECT id, user_id, number, status, accrual, uploaded_at
FROM orders
WHERE number = $1;

-- name: GetOrdersByUser :many
SELECT number, status, accrual, uploaded_at
FROM orders
WHERE user_id = $1
ORDER BY uploaded_at DESC;

-- name: GetAllOrders :many
SELECT id, user_id, number, status, accrual, uploaded_at
FROM orders
ORDER BY uploaded_at DESC;

-- name: CreateWithdrawal :exec
INSERT INTO withdrawals (user_id, order_number, sum, processed_at)
VALUES ($1, $2, $3, $4);

-- name: GetWithdrawalsByUser :many
SELECT order_number, sum, processed_at
FROM withdrawals
WHERE user_id = $1
ORDER BY processed_at DESC;

-- name: GetUserByLogin :one
SELECT id, login, password, balance, withdrawn
FROM users
WHERE login = $1;

-- name: GetUserBalance :one
SELECT balance, withdrawn
FROM users
WHERE id = $1
FOR UPDATE;

-- name: UpdateBalance :exec
UPDATE users
SET balance = balance + $2, withdrawn = withdrawn + $3
WHERE id = $1;

-- name: UpdateOrder :exec
UPDATE orders
SET status = $1, accrual = $2, uploaded_at = $3
WHERE number = $4;

