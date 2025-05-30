// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: queries.sql

package storage

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createOrder = `-- name: CreateOrder :exec
INSERT INTO orders (user_id, number, status, accrual, uploaded_at)
VALUES ($1, $2, $3, $4, $5)
`

type CreateOrderParams struct {
	UserID     pgtype.Int8        `json:"user_id"`
	Number     string             `json:"number"`
	Status     string             `json:"status"`
	Accrual    pgtype.Float8      `json:"accrual"`
	UploadedAt pgtype.Timestamptz `json:"uploaded_at"`
}

func (q *Queries) CreateOrder(ctx context.Context, arg CreateOrderParams) error {
	_, err := q.db.Exec(ctx, createOrder,
		arg.UserID,
		arg.Number,
		arg.Status,
		arg.Accrual,
		arg.UploadedAt,
	)
	return err
}

const createUser = `-- name: CreateUser :one
INSERT INTO users (login, password)
VALUES ($1, $2)
RETURNING id
`

type CreateUserParams struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (int64, error) {
	row := q.db.QueryRow(ctx, createUser, arg.Login, arg.Password)
	var id int64
	err := row.Scan(&id)
	return id, err
}

const createWithdrawal = `-- name: CreateWithdrawal :exec
INSERT INTO withdrawals (user_id, order_number, sum, processed_at)
VALUES ($1, $2, $3, $4)
`

type CreateWithdrawalParams struct {
	UserID      pgtype.Int8        `json:"user_id"`
	OrderNumber string             `json:"order_number"`
	Sum         float64            `json:"sum"`
	ProcessedAt pgtype.Timestamptz `json:"processed_at"`
}

func (q *Queries) CreateWithdrawal(ctx context.Context, arg CreateWithdrawalParams) error {
	_, err := q.db.Exec(ctx, createWithdrawal,
		arg.UserID,
		arg.OrderNumber,
		arg.Sum,
		arg.ProcessedAt,
	)
	return err
}

const getAllOrders = `-- name: GetAllOrders :many
SELECT id, user_id, number, status, accrual, uploaded_at
FROM orders
ORDER BY uploaded_at DESC
`

func (q *Queries) GetAllOrders(ctx context.Context) ([]Order, error) {
	rows, err := q.db.Query(ctx, getAllOrders)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Order
	for rows.Next() {
		var i Order
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.Number,
			&i.Status,
			&i.Accrual,
			&i.UploadedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getOrderByNumber = `-- name: GetOrderByNumber :one
SELECT id, user_id, number, status, accrual, uploaded_at
FROM orders
WHERE number = $1
`

func (q *Queries) GetOrderByNumber(ctx context.Context, number string) (Order, error) {
	row := q.db.QueryRow(ctx, getOrderByNumber, number)
	var i Order
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Number,
		&i.Status,
		&i.Accrual,
		&i.UploadedAt,
	)
	return i, err
}

const getOrdersByUser = `-- name: GetOrdersByUser :many
SELECT number, status, accrual, uploaded_at
FROM orders
WHERE user_id = $1
ORDER BY uploaded_at DESC
`

type GetOrdersByUserRow struct {
	Number     string             `json:"number"`
	Status     string             `json:"status"`
	Accrual    pgtype.Float8      `json:"accrual"`
	UploadedAt pgtype.Timestamptz `json:"uploaded_at"`
}

func (q *Queries) GetOrdersByUser(ctx context.Context, userID pgtype.Int8) ([]GetOrdersByUserRow, error) {
	rows, err := q.db.Query(ctx, getOrdersByUser, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetOrdersByUserRow
	for rows.Next() {
		var i GetOrdersByUserRow
		if err := rows.Scan(
			&i.Number,
			&i.Status,
			&i.Accrual,
			&i.UploadedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getUserBalance = `-- name: GetUserBalance :one
SELECT balance, withdrawn
FROM users
WHERE id = $1
FOR UPDATE
`

type GetUserBalanceRow struct {
	Balance   pgtype.Float8 `json:"balance"`
	Withdrawn pgtype.Float8 `json:"withdrawn"`
}

func (q *Queries) GetUserBalance(ctx context.Context, id int64) (GetUserBalanceRow, error) {
	row := q.db.QueryRow(ctx, getUserBalance, id)
	var i GetUserBalanceRow
	err := row.Scan(&i.Balance, &i.Withdrawn)
	return i, err
}

const getUserByLogin = `-- name: GetUserByLogin :one
SELECT id, login, password, balance, withdrawn
FROM users
WHERE login = $1
`

func (q *Queries) GetUserByLogin(ctx context.Context, login string) (User, error) {
	row := q.db.QueryRow(ctx, getUserByLogin, login)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Login,
		&i.Password,
		&i.Balance,
		&i.Withdrawn,
	)
	return i, err
}

const getWithdrawalsByUser = `-- name: GetWithdrawalsByUser :many
SELECT order_number, sum, processed_at
FROM withdrawals
WHERE user_id = $1
ORDER BY processed_at DESC
`

type GetWithdrawalsByUserRow struct {
	OrderNumber string             `json:"order_number"`
	Sum         float64            `json:"sum"`
	ProcessedAt pgtype.Timestamptz `json:"processed_at"`
}

func (q *Queries) GetWithdrawalsByUser(ctx context.Context, userID pgtype.Int8) ([]GetWithdrawalsByUserRow, error) {
	rows, err := q.db.Query(ctx, getWithdrawalsByUser, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetWithdrawalsByUserRow
	for rows.Next() {
		var i GetWithdrawalsByUserRow
		if err := rows.Scan(&i.OrderNumber, &i.Sum, &i.ProcessedAt); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateBalance = `-- name: UpdateBalance :exec
UPDATE users
SET balance = $2
WHERE id = $1
`

type UpdateBalanceParams struct {
	ID      int64         `json:"id"`
	Balance pgtype.Float8 `json:"balance"`
}

func (q *Queries) UpdateBalance(ctx context.Context, arg UpdateBalanceParams) error {
	_, err := q.db.Exec(ctx, updateBalance, arg.ID, arg.Balance)
	return err
}

const updateOrder = `-- name: UpdateOrder :exec
UPDATE orders
SET status = $1, accrual = $2, uploaded_at = $3
WHERE number = $4
`

type UpdateOrderParams struct {
	Status     string             `json:"status"`
	Accrual    pgtype.Float8      `json:"accrual"`
	UploadedAt pgtype.Timestamptz `json:"uploaded_at"`
	Number     string             `json:"number"`
}

func (q *Queries) UpdateOrder(ctx context.Context, arg UpdateOrderParams) error {
	_, err := q.db.Exec(ctx, updateOrder,
		arg.Status,
		arg.Accrual,
		arg.UploadedAt,
		arg.Number,
	)
	return err
}

const updateWithdrawn = `-- name: UpdateWithdrawn :exec
UPDATE users
SET withdrawn = $2
WHERE id = $1
`

type UpdateWithdrawnParams struct {
	ID        int64         `json:"id"`
	Withdrawn pgtype.Float8 `json:"withdrawn"`
}

func (q *Queries) UpdateWithdrawn(ctx context.Context, arg UpdateWithdrawnParams) error {
	_, err := q.db.Exec(ctx, updateWithdrawn, arg.ID, arg.Withdrawn)
	return err
}
